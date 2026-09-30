#!/usr/bin/env python3
"""
Genera un banco de preguntas sobre los nodos de Cortex, con 3 variantes mal escritas cada una.

  .venv/bin/python herramientas/generar_preguntas.py --total 14000          # genera (reanudable)
  .venv/bin/python herramientas/generar_preguntas.py --total 14000 --armar  # solo arma los archivos finales

Uso previsto: evaluar y ajustar la búsqueda de Metrín (cada pregunta sabe qué nodo debe encontrar).
No es dataset de tono para LoRA (ver docs/PLAN_LORA_METRIN_MLX.md): no lleva respuestas.

Salida en data/preguntas/:
  preguntas.jsonl   una fila por pregunta base, con sus 3 variantes
  plano.jsonl       una fila por texto (base + variantes): nivel 0 = bien escrita, 1-3 = con errores
  resumen.json      reparto por nodo y por tipo
  crudo/            tandas devueltas por el modelo (permite reanudar)

Preguntas base: Claude (SDK si hay ANTHROPIC_API_KEY; si no, `claude -p`), con el cuerpo del nodo
como única fuente. Variantes: reglas deterministas (semilla por id), sin modelo:
  1 · escritura rápida: sin tildes, sin ¿, abreviaturas (q, xq, d, tb, pa)
  2 · ortografía: b/v, s/c/z, ll/y, h muda, g/j, letras cambiadas u omitidas
  3 · gramática: concordancia rota, artículos que faltan o sobran, muletillas, orden alterado
"""
from __future__ import annotations

import argparse
import hashlib
import json
import math
import random
import re
import subprocess
import sys
import time
import unicodedata
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

RAIZ = Path(__file__).resolve().parent.parent
ARBOL = RAIZ / ".cortex" / "tree"
SALIDA = RAIZ / "data" / "preguntas"
CRUDO = SALIDA / "crudo"
EXCLUIR_RAMAS = ()
INCLUIR_UI = ("ui-pantallas/17w0yki8iew",)            # de ui-pantallas solo entran las pantallas validadas
EXCLUIR_NODOS = ("tutoriales", "tutoriales/crear-un-presupuesto-de-obra-en-s10")
POR_LLAMADA = 70
MIN_PALABRAS = 40   # nodos con menos cuerpo se ignoran (ajustable con --min-palabras)
ENFOQUES = [
    ("concepto", "qué es, para qué sirve, qué incluye, qué diferencia hay entre conceptos del nodo"),
    ("procedimiento", "cómo se hace algo paso a paso, en qué menú o ventana, qué dato va en qué campo, en qué orden"),
    ("problema", "situaciones reales: algo no sale, un error, un caso raro, qué hago si..., qué pasa cuando..."),
    ("usuario", "preguntas cortas y directas de un usuario apurado en obra u oficina (residente, almacenero, contador, asistente de RRHH, gerente)"),
]


ANGULOS = ["pregunta de alguien nuevo en la empresa", "pregunta de alguien con años usando S10", "pregunta con un caso numérico o de fecha concreta",
           "pregunta que empieza contando el contexto de la obra", "pregunta muy corta, casi telegráfica", "pregunta de un jefe que supervisa a otro",
           "pregunta comparando dos opciones", "pregunta sobre un error o algo que no cuadra", "pregunta de quien viene de Excel",
           "pregunta de cierre de mes o de año"]


def log(*a):
    print(time.strftime("%H:%M:%S"), *a, flush=True)


# ─────────────────────────────── nodos ─────────────────────────────────────
def nodos() -> list[dict]:
    out = []
    for f in sorted(ARBOL.rglob("*.md")):
        t = f.read_text(encoding="utf-8")
        m = re.match(r"---\n(.*?)\n---\n?(.*)", t, re.S)
        if not m:
            continue
        fm, cuerpo = m.groups()
        rel = f.relative_to(ARBOL).with_suffix("").as_posix()
        path = "" if rel == "_node" else rel.replace("/_node", "")
        if path.split("/")[0] in EXCLUIR_RAMAS or path in EXCLUIR_NODOS:
            continue
        if path.startswith("ui-pantallas") and path not in INCLUIR_UI:
            continue
        titulo = (re.search(r"^title:\s*(.+)$", fm, re.M) or [None, ""])[1].strip().strip('"')
        resumen = re.search(r"^summary:\s*(.+?)(?=\n[a-z_]+:|\Z)", fm, re.M | re.S)
        resumen = re.sub(r"\s+", " ", resumen.group(1)).strip().strip('"') if resumen else ""
        if len(cuerpo.split()) < MIN_PALABRAS:
            continue
        out.append({"path": path, "titulo": titulo, "resumen": resumen, "cuerpo": cuerpo.strip(),
                    "rama": path.split("/")[0] if path else "(raíz)", "palabras": len(cuerpo.split())})
    return out


def reparto(ns: list[dict], total: int, minimo: int = 60) -> dict[str, int]:
    pesos = {n["path"]: math.sqrt(n["palabras"]) for n in ns}
    suma = sum(pesos.values())
    cuota = {p: max(minimo, int(total * w / suma)) for p, w in pesos.items()}
    # ajuste fino para que sumen exactamente `total`
    dif = total - sum(cuota.values())
    orden = sorted(cuota, key=lambda p: -pesos[p])
    i = 0
    while dif:
        p = orden[i % len(orden)]
        paso = 1 if dif > 0 else -1
        if cuota[p] + paso >= minimo:
            cuota[p] += paso
            dif -= paso
        i += 1
    return cuota


# ─────────────────────────────── modelo ────────────────────────────────────
SISTEMA = """Generas preguntas de usuarios reales del ERP peruano S10 (constructoras e inmobiliarias) y de Optimiza 360.
Cada pregunta debe poder responderse SOLO con el contenido del nodo dado. Español de Perú, natural, variado:
cambia persona, longitud, registro (formal/coloquial) y forma de empezar. Nada de numerarlas ni repetir plantillas.
Devuelve SOLO JSON: {"preguntas": [{"pregunta": "...", "tipo": "concepto|procedimiento|problema|dato|comparacion|contexto"}]}"""


def _extraer_json(t: str):
    t = re.sub(r"```(?:json)?", "", t)
    i = t.find("{")
    return json.loads(t[i:t.rfind("}") + 1])


def preguntar_modelo(nodo: dict, enfoque: tuple, n: int, evitar: list[str]) -> list[dict]:
    usuario = (f"Nodo `{nodo['path'] or '(raíz)'}` — {nodo['titulo']}\nResumen: {nodo['resumen']}\n\n"
               f"Contenido del nodo:\n{nodo['cuerpo'][:6000]}\n\n"
               f"Enfoque de esta tanda: {enfoque[0]} — {enfoque[1]}.\n"
               f"Ángulo propio de esta tanda (para no coincidir con otras tandas del mismo nodo): {ANGULOS[hash(nodo['path'] + enfoque[0] + str(n)) % len(ANGULOS)]}.\n"
               f"Genera exactamente {n} preguntas distintas entre sí.\n")
    if evitar:
        usuario += "No repitas ni parafrasees de cerca estas ya hechas:\n" + "\n".join(f"- {q}" for q in evitar[-80:])
    try:
        import anthropic, os
        if os.environ.get("ANTHROPIC_API_KEY"):
            c = anthropic.Anthropic()
            with c.messages.stream(model="claude-opus-5", max_tokens=16000, system=SISTEMA,
                                   messages=[{"role": "user", "content": usuario}],
                                   output_config={"effort": "low"}) as s:
                r = s.get_final_message()
            return _extraer_json("".join(b.text for b in r.content if b.type == "text"))["preguntas"]
    except ImportError:
        pass
    r = subprocess.run(["claude", "-p", "--output-format", "json", "--append-system-prompt", SISTEMA,
                        "--disallowedTools", "Bash,Edit,Write,Read,Glob,Grep,WebFetch,WebSearch,Task,NotebookEdit"],
                       input=usuario, capture_output=True, text=True, cwd="/tmp", timeout=900)
    d = json.loads(r.stdout)
    if d.get("is_error"):
        raise RuntimeError(d.get("result", "")[:200])
    return _extraer_json(d["result"])["preguntas"]


# ─────────────────────────────── variantes ─────────────────────────────────
def sin_tildes(t: str) -> str:
    t = t.replace("ñ", "\0").replace("Ñ", "\1")                      # la ñ no es una tilde: se conserva
    t = "".join(c for c in unicodedata.normalize("NFKD", t) if not unicodedata.combining(c))
    return t.replace("\0", "ñ").replace("\1", "Ñ")


def norm(t: str) -> str:
    return re.sub(r"[^a-z0-9ñ ]+", "", sin_tildes(t.lower())).strip()


ABREV = [(r"\bpor qu[eé]\b", "xq"), (r"\bporque\b", "xq"), (r"\bque\b", "q"), (r"\bqu[eé]\b", "q"), (r"\btambi[eé]n\b", "tb"),
         (r"\bde\b", "d"), (r"\bpara\b", "pa"), (r"\bpor favor\b", "porfa"), (r"\bcomo\b", "cmo"), (r"\bestoy\b", "toy"),
         (r"\bmuchas gracias\b", "grax"), (r"\bpresupuesto\b", "ppto"), (r"\bn[uú]mero\b", "nro")]


def nivel1(t: str, rnd: random.Random) -> str:
    s = sin_tildes(t).lower().replace("¿", "").replace("¡", "")
    for pat, rep in ABREV:
        if rnd.random() < 0.55:
            s = re.sub(pat, rep, s)
    if rnd.random() < 0.6:
        s = s.rstrip("?").rstrip()
    if rnd.random() < 0.25:
        s = s.replace(",", "")
    return re.sub(r"\s+", " ", s).strip()


CAMBIOS_ORTO = [
    (r"v", "b"), (r"b", "v"), (r"ll", "y"), (r"y(?=[aeiou])", "ll"), (r"z", "s"), (r"c(?=[ei])", "s"),
    (r"s(?=[ei])", "c"), (r"(?<![a-z])h", ""), (r"g(?=[ei])", "j"), (r"j(?=[ei])", "g"), (r"qu(?=[ei])", "k"),
    (r"x", "cs"), (r"rr", "r"), (r"ción\b", "sion"), (r"á", "a"), (r"é", "e"), (r"í", "i"), (r"ó", "o"), (r"ú", "u"),
]


def _tocar_palabra(p: str, rnd: random.Random) -> str:
    if len(p) < 4:
        return p
    op = rnd.choice(["trans", "omite", "dobla"])
    i = rnd.randrange(1, len(p) - 1)
    if op == "trans":
        return p[:i - 1] + p[i] + p[i - 1] + p[i + 1:]
    if op == "omite":
        return p[:i] + p[i + 1:]
    return p[:i] + p[i] + p[i:]


def nivel2(t: str, rnd: random.Random) -> str:
    palabras = t.split()
    cand = [i for i, w in enumerate(palabras) if len(w) > 3]
    rnd.shuffle(cand)
    hechos = 0
    for i in cand:
        if hechos >= rnd.randint(2, 4):
            break
        w = palabras[i]
        pats = [(p, r) for p, r in CAMBIOS_ORTO if re.search(p, w)]
        if pats and rnd.random() < 0.75:
            p, r = rnd.choice(pats)
            nuevo = re.sub(p, r, w, count=1)
        else:
            nuevo = _tocar_palabra(w, rnd)
        if nuevo != w:
            palabras[i] = nuevo
            hechos += 1
    s = " ".join(palabras)
    if rnd.random() < 0.5:
        s = s.replace("¿", "")
    return s


MULETILLAS = ["oye", "una consulta", "disculpa", "hola", "buenas", "consulta rapida", "oe", "ayuda", "urgente", "a ver"]
GENERO = {"el": "la", "la": "el", "un": "una", "una": "un", "los": "las", "las": "los", "del": "de la", "al": "a la"}
ARTICULOS = {"el", "la", "los", "las", "un", "una", "del", "al"}


def nivel3(t: str, rnd: random.Random) -> str:
    s = sin_tildes(t) if rnd.random() < 0.7 else t
    s = s.replace("¿", "").rstrip("?")
    w = s.split()
    ops = rnd.sample(["genero", "quita_art", "plural", "orden", "duplica", "verbo"], k=rnd.randint(2, 3))
    for op in ops:
        idx = [i for i, x in enumerate(w) if x.lower() in ARTICULOS]
        if op == "genero" and idx:
            i = rnd.choice(idx)
            w[i] = GENERO.get(w[i].lower(), w[i])
        elif op == "quita_art" and idx and len(w) > 4:
            del w[rnd.choice(idx)]
        elif op == "plural" and idx:
            i = rnd.choice(idx)
            if i + 1 < len(w) and len(w[i + 1]) > 3:
                w[i + 1] = w[i + 1][:-1] if w[i + 1].endswith("s") else w[i + 1] + "s"
        elif op == "orden" and len(w) > 5:
            i = rnd.randrange(1, len(w) - 2)
            w[i], w[i + 1] = w[i + 1], w[i]
        elif op == "duplica" and len(w) > 3:
            i = rnd.randrange(len(w))
            w.insert(i, w[i])
        elif op == "verbo":
            for i, x in enumerate(w):
                y = re.sub(r"(puedo|debo|tengo|hago|necesito)$", lambda m: {"puedo": "puede", "debo": "debe", "tengo": "tiene", "hago": "hace", "necesito": "necesita"}[m.group(1)], x)
                if y == x:
                    y = re.sub(r"(puede|debe|tiene|hace)$", lambda m: {"puede": "pueden", "debe": "deben", "tiene": "tienen", "hace": "hacen"}[m.group(1)], x)
                if y != x:
                    w[i] = y
                    break
    s = " ".join(w)
    if rnd.random() < 0.55:
        s = f"{rnd.choice(MULETILLAS)} {s[:1].lower() + s[1:]}"
    if rnd.random() < 0.5:
        s = nivel2(s, rnd)
    return re.sub(r"\s+", " ", s).strip()


def variantes(texto: str, semilla: str) -> list[dict]:
    rnd = random.Random(int(hashlib.sha256(semilla.encode()).hexdigest()[:12], 16))
    out, vistos = [], {texto}
    for nivel, f in ((1, nivel1), (2, nivel2), (3, nivel3)):
        v = f(texto, rnd)
        for _ in range(6):                     # garantiza que cada variante sea distinta
            if v not in vistos:
                break
            v = f(texto, rnd)
        if v in vistos:
            v = nivel2(v, rnd) if nivel != 2 else nivel3(v, rnd)
        vistos.add(v)
        out.append({"nivel": nivel, "texto": v})
    return out


# ─────────────────────────────── generar y armar ───────────────────────────
def tandas_de(nodo: dict, cuota: int) -> list[tuple]:
    objetivo = int(cuota * 1.2) + 5              # margen para descartar repetidas
    n = max(math.ceil(objetivo / POR_LLAMADA), len(ENFOQUES))   # cada enfoque cubierto al menos una vez
    return [(nodo, ENFOQUES[k % len(ENFOQUES)], math.ceil(objetivo / n), k) for k in range(n)]


def archivo_tanda(nodo: dict, k: int) -> Path:
    return CRUDO / f"{(nodo['path'] or 'raiz').replace('/', '__')}__{k}.json"


def correr_tanda(nodo, enfoque, n, k) -> int:
    dest = archivo_tanda(nodo, k)
    if dest.exists():
        return len(json.loads(dest.read_text())["preguntas"])
    previas = []
    for j in range(k):
        p = archivo_tanda(nodo, j)
        if p.exists():
            previas += [q["pregunta"] for q in json.loads(p.read_text())["preguntas"]]
    for intento in range(3):
        try:
            qs = preguntar_modelo(nodo, enfoque, n, previas)
            qs = [q for q in qs if isinstance(q, dict) and len(str(q.get("pregunta", ""))) > 8]
            dest.write_text(json.dumps({"nodo": nodo["path"], "enfoque": enfoque[0], "preguntas": qs}, ensure_ascii=False))
            return len(qs)
        except Exception as e:
            log(f"  reintento {intento + 1} {nodo['path']}#{k}: {e}")
            time.sleep(10 * (intento + 1))
    return 0


META = re.compile(r"\b(nodo|nodos|cortex|rama del|este documento|la base de conocimiento|fragmento|caveat|confiabilidad)\b", re.I)


def armar(ns: list[dict], cuota: dict[str, int], sufijo: str = "") -> None:
    filas, plano, resumen = [], [], {"por_nodo": {}, "por_tipo": {}, "por_rama": {}}
    vistas_global: set[str] = set()
    for nodo in ns:
        vistas, propias = set(), []
        k = 0
        while archivo_tanda(nodo, k).exists():
            d = json.loads(archivo_tanda(nodo, k).read_text())
            for q in d["preguntas"]:
                clave = norm(q["pregunta"])
                if not clave or clave in vistas or clave in vistas_global or META.search(q["pregunta"]):
                    continue    # repetidas (también entre nodos) y preguntas sobre la base en sí, no sobre el ERP
                vistas.add(clave)
                propias.append((q, d["enfoque"]))
            k += 1
        propias = propias[:cuota[nodo["path"]]]
        vistas_global.update(norm(q["pregunta"]) for q, _ in propias)
        for i, (q, enf) in enumerate(propias):
            texto = re.sub(r"\s+", " ", q["pregunta"]).strip()
            pid = f"{(nodo['path'] or 'raiz').replace('/', '.')}-{i:04d}"
            vs = variantes(texto, pid)
            fila = {"id": pid, "nodo": nodo["path"], "titulo_nodo": nodo["titulo"], "rama": nodo["rama"],
                    "tipo": q.get("tipo", "concepto"), "enfoque": enf, "pregunta": texto, "variantes": vs}
            filas.append(fila)
            plano.append({"id": pid, "id_base": pid, "nodo": nodo["path"], "nivel": 0, "texto": texto})
            plano += [{"id": f"{pid}-v{v['nivel']}", "id_base": pid, "nodo": nodo["path"], "nivel": v["nivel"], "texto": v["texto"]} for v in vs]
            resumen["por_tipo"][fila["tipo"]] = resumen["por_tipo"].get(fila["tipo"], 0) + 1
            resumen["por_rama"][nodo["rama"]] = resumen["por_rama"].get(nodo["rama"], 0) + 1
        resumen["por_nodo"][nodo["path"] or "(raíz)"] = {"cuota": cuota[nodo["path"]], "obtenidas": len(propias)}
    SALIDA.mkdir(parents=True, exist_ok=True)
    (SALIDA / f"preguntas{sufijo}.jsonl").write_text("".join(json.dumps(f, ensure_ascii=False) + "\n" for f in filas), encoding="utf-8")
    (SALIDA / f"plano{sufijo}.jsonl").write_text("".join(json.dumps(f, ensure_ascii=False) + "\n" for f in plano), encoding="utf-8")
    resumen["total_base"], resumen["total_textos"] = len(filas), len(plano)
    resumen["faltan"] = {p: v for p, v in resumen["por_nodo"].items() if v["obtenidas"] < v["cuota"]}
    (SALIDA / f"resumen{sufijo}.json").write_text(json.dumps(resumen, indent=1, ensure_ascii=False), encoding="utf-8")
    log(f"armado: {len(filas)} preguntas base, {len(plano)} textos; nodos con faltante: {len(resumen['faltan'])}")


def main():
    global ENFOQUES, MIN_PALABRAS
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--total", type=int, default=14000)
    ap.add_argument("--hilos", type=int, default=4)
    ap.add_argument("--armar", action="store_true", help="solo arma los archivos finales con lo ya generado")
    ap.add_argument("--solo", help="limita a nodos cuyo path empiece así (pruebas)")
    ap.add_argument("--enfoques", help="subconjunto de ENFOQUES por coma (ej: concepto,procedimiento)")
    ap.add_argument("--min-palabras", type=int, default=40, help="ignora nodos con menos palabras")
    ap.add_argument("--min-cuota", type=int, default=60, help="preguntas mínimas por nodo")
    ap.add_argument("--sufijo", default="", help="sufijo de salida (ej: _aprendizaje) para no tocar el banco de evaluación")
    a = ap.parse_args()
    MIN_PALABRAS = a.min_palabras
    if a.enfoques:
        elegidos = {e.strip() for e in a.enfoques.split(",")}
        filtrados = [e for e in ENFOQUES if e[0] in elegidos]
        if not filtrados:
            ap.error(f"enfoques desconocidos: {sorted(elegidos)}")
        ENFOQUES = filtrados
    ns = nodos()
    if a.solo:
        ns = [n for n in ns if n["path"].startswith(a.solo)]
    if not ns:
        ap.error("sin nodos para ese filtro")
    cuota = reparto(ns, a.total, a.min_cuota)
    CRUDO.mkdir(parents=True, exist_ok=True)
    if not a.armar:
        tandas = [t for n in ns for t in tandas_de(n, cuota[n["path"]])]
        # cada tanda es una llamada independiente: todas en paralelo (las repetidas se filtran al armar)
        pendientes = [t for t in tandas if not archivo_tanda(t[0], t[3]).exists()]
        log(f"{len(ns)} nodos, {len(tandas)} tandas ({len(pendientes)} pendientes), {a.hilos} llamadas a la vez")
        hechos = 0
        with ThreadPoolExecutor(a.hilos) as ex:
            futuros = {ex.submit(correr_tanda, *t): f"{t[0]['path'] or '(raíz)'}#{t[3]}" for t in pendientes}
            for f in as_completed(futuros):
                hechos += 1
                log(f"[{hechos}/{len(futuros)}] {futuros[f]}: {f.result()} preguntas crudas")
    armar(ns, cuota, a.sufijo)


if __name__ == "__main__":
    main()
