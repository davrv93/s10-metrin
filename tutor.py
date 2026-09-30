#!/usr/bin/env python3
"""
tutor.py — genera tutoriales paso a paso de S10 a partir de la base de conocimiento.

Flujo por tutorial (un grafo simple, sin dependencias de orquestación):

  planificar  la tarea -> título + pasos, cada uno con sus consultas de búsqueda
  recuperar   por paso: fragmentos (BM25 sobre kb/*.jsonl) + nodos de Cortex (API local)
  evaluar     ¿la evidencia alcanza para el paso? si no, una ronda más con consultas nuevas
  generar     Markdown con pasos numerados, citas [F#] y avisos de confiabilidad
  publicar    tutoriales/<slug>.md  +  nodo de Cortex tutoriales/<slug>

Modos:
  .venv/bin/python tutor.py preguntar "¿Cómo creo un presupuesto de obra?"
  .venv/bin/python tutor.py auto --max 8     # propone tareas desde el árbol de Cortex y genera las que falten
  .venv/bin/python tutor.py listar

Motor del modelo (se elige solo):
  - SDK de Anthropic si hay credenciales (ANTHROPIC_API_KEY o perfil de `ant auth login`).
  - si no, el CLI de Claude Code en modo headless (`claude -p`), ya autenticado en esta Mac.
  Forzar con S10_MOTOR=sdk | cli.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import math
import os
import re
import shutil
import subprocess
import sys
import time
import unicodedata
from collections import Counter
from pathlib import Path

import requests

RAIZ = Path(__file__).resolve().parent
KB = RAIZ / "kb"
SALIDA = RAIZ / "tutoriales"
CORTEX = os.environ.get("S10_CORTEX_API", "http://localhost:4748/api")
MODELO = os.environ.get("S10_MODELO", "claude-opus-5")

AVISO_CONFIANZA = {
    "oficial": "oficial S10",
    "tercero-sin-verificar": "copia de tercero, posible versión antigua",
    "academico": "caso académico",
    "estado-2013": "documento del Estado, 2013",
}


def log(*a) -> None:
    print(time.strftime("%H:%M:%S"), *a, flush=True)


# ─────────────────────────────── motor del modelo ──────────────────────────
class Motor:
    """Una llamada de texto al modelo. SDK si hay credenciales; si no, `claude -p`."""

    def __init__(self) -> None:
        elegido = os.environ.get("S10_MOTOR")
        self.cliente = None
        if elegido != "cli":
            try:
                import anthropic

                c = anthropic.Anthropic()
                if elegido == "sdk" or os.environ.get("ANTHROPIC_API_KEY") or os.environ.get("ANTHROPIC_AUTH_TOKEN") \
                        or (Path.home() / ".config/anthropic").exists():
                    self.cliente = c
            except Exception:
                self.cliente = None
        if self.cliente is None and not shutil.which("claude"):
            raise SystemExit("Sin motor: define ANTHROPIC_API_KEY o instala Claude Code (`claude`).")
        self.nombre = f"sdk:{MODELO}" if self.cliente else "cli:claude-code"

    def texto(self, sistema: str, usuario: str, esfuerzo: str = "medium") -> str:
        if self.cliente is not None:
            return self._sdk(sistema, usuario, esfuerzo)
        return self._cli(sistema, usuario)

    def _sdk(self, sistema: str, usuario: str, esfuerzo: str) -> str:
        import anthropic

        try:
            with self.cliente.beta.messages.stream(
                model=MODELO,
                max_tokens=32000,
                system=sistema,
                messages=[{"role": "user", "content": usuario}],
                output_config={"effort": esfuerzo},
                betas=["server-side-fallback-2026-07-01"],
                fallbacks="default",
            ) as s:
                r = s.get_final_message()
        except anthropic.RateLimitError:
            time.sleep(30)
            return self._sdk(sistema, usuario, esfuerzo)
        if r.stop_reason == "refusal":
            raise RuntimeError(f"el modelo declinó la petición ({getattr(r.stop_details, 'category', None)})")
        return "".join(b.text for b in r.content if b.type == "text")

    def _cli(self, sistema: str, usuario: str) -> str:
        cmd = ["claude", "-p", "--output-format", "json", "--append-system-prompt", sistema,
               "--disallowedTools", "Bash,Edit,Write,Read,Glob,Grep,WebFetch,WebSearch,Task,NotebookEdit"]
        r = subprocess.run(cmd, input=usuario, capture_output=True, text=True, cwd="/tmp", timeout=900)
        try:
            d = json.loads(r.stdout)
        except json.JSONDecodeError:
            raise RuntimeError(f"respuesta ilegible del CLI: {r.stdout[:200]} {r.stderr[:200]}")
        if d.get("is_error"):
            raise RuntimeError(f"CLI: {d.get('result', '')[:300]}")
        return d.get("result", "")


def extraer_json(texto: str):
    """Toma el primer objeto/array JSON del texto (el modelo a veces lo envuelve en ```)."""
    t = re.sub(r"```(?:json)?", "", texto)
    ini = min([i for i in (t.find("{"), t.find("[")) if i >= 0], default=-1)
    if ini < 0:
        raise ValueError(f"sin JSON en: {texto[:200]}")
    apertura = t[ini]
    cierre = "}" if apertura == "{" else "]"
    nivel, dentro, esc = 0, False, False
    for k in range(ini, len(t)):
        ch = t[k]
        if dentro:
            esc = (ch == "\\" and not esc)
            if ch == '"' and not esc:
                dentro = False
            continue
        if ch == '"':
            dentro = True
        elif ch == apertura:
            nivel += 1
        elif ch == cierre:
            nivel -= 1
            if nivel == 0:
                return json.loads(t[ini:k + 1])
    raise ValueError("JSON incompleto")


# ─────────────────────────────── recuperación ──────────────────────────────
VACIAS = set("""a al algo como con cual cuando de del desde donde el en entre es esta este esto hay la las le lo los
mas me mi no o para pero por que se si sin sobre su sus te tu un una uno y ya the of and to in is""".split())


def tokens(t: str) -> list[str]:
    t = unicodedata.normalize("NFKD", t.lower())
    t = "".join(c for c in t if not unicodedata.combining(c))
    return [w[:7] for w in re.findall(r"[a-z0-9]{2,}", t) if w not in VACIAS]  # prefijo = stemming barato


class Indice:
    """BM25 sobre todos los kb/fragmentos*.jsonl (PDF, web y YouTube)."""

    def __init__(self) -> None:
        self.docs: list[dict] = []
        for f in sorted(KB.glob("fragmentos*.jsonl")):
            for l in f.read_text(encoding="utf-8").splitlines():
                if l.strip():
                    self.docs.append(self._normalizar(json.loads(l)))
        self.toks = [tokens(f"{d.get('titulo') or ''} {d.get('manual') or ''} {d['texto']}") for d in self.docs]
        self.tf = [Counter(t) for t in self.toks]
        n = len(self.docs)
        df = Counter(w for t in self.toks for w in set(t))
        self.idf = {w: math.log(1 + (n - c + 0.5) / (c + 0.5)) for w, c in df.items()}
        self.prom = sum(map(len, self.toks)) / max(n, 1)
        self.huella = hashlib.sha256("".join(d["id"] for d in self.docs).encode()).hexdigest()[:12]

    @staticmethod
    def _normalizar(d: dict) -> dict:
        """Los fragmentos de YouTube (youtube.py) traen otro esquema: se llevan al común."""
        if d.get("tipo") == "video":
            d.setdefault("titulo", f"Video: {d.get('video')}")
            d.setdefault("pagina", f"min {d.get('desde')}–{d.get('hasta')}")
            d.setdefault("fuente", d.get("url"))
            d.setdefault("confianza", "oficial")  # canal oficial de S10
        d.setdefault("confianza", "oficial")
        return d

    def buscar(self, consulta: str, k: int = 6) -> list[dict]:
        q = tokens(consulta)
        puntos = []
        for i, tf in enumerate(self.tf):
            ln = len(self.toks[i])
            s = sum(self.idf.get(w, 0) * tf[w] * 2.2 / (tf[w] + 1.2 * (0.25 + 0.75 * ln / self.prom)) for w in q if w in tf)
            if s > 0:
                # a igualdad, lo oficial primero
                s *= 1.15 if self.docs[i].get("confianza") == "oficial" else 1.0
                puntos.append((s, i))
        puntos.sort(reverse=True)
        return [self.docs[i] for _, i in puntos[:k]]


class Cortex:
    def __init__(self) -> None:
        sec = (RAIZ / ".cortex/.secrets.yaml").read_text()
        self.h = {"Authorization": "Bearer " + re.search(r"ai-agent:\s*(\S+)", sec).group(1)}
        if os.environ.get("S10_CORTEX_HOST"):  # en Docker: Cortex solo acepta Host: localhost
            self.h["Host"] = os.environ["S10_CORTEX_HOST"]
        try:
            self.vivo = requests.get(f"{CORTEX}/health", headers=self.h, timeout=5).ok
        except requests.RequestException:
            self.vivo = False

    def buscar(self, q: str, k: int = 3) -> list[dict]:
        if not self.vivo:
            return []
        r = requests.get(f"{CORTEX}/search", params={"q": q}, headers=self.h, timeout=15)
        d = r.json()
        res = d.get("results") or d.get("hits") or []
        nodos = []
        for x in res:
            p = x.get("path")
            if p is None or p.startswith("tutoriales"):
                continue
            n = requests.get(f"{CORTEX}/node/{p}", headers=self.h, timeout=15).json()
            n = n.get("node", n)
            nodos.append({"path": p, "title": n.get("title"), "body": n.get("body") or n.get("summary") or ""})
            if len(nodos) >= k:
                break
        return nodos

    def arbol(self) -> list[str]:
        if not self.vivo:
            return []
        titulos = []
        for f in sorted((RAIZ / ".cortex/tree").rglob("*.md")):
            m = re.search(r"^title:\s*(.+)$", f.read_text(encoding="utf-8"), re.M)
            rel = f.relative_to(RAIZ / ".cortex/tree").with_suffix("").as_posix().replace("/_node", "")
            if m and not rel.startswith("tutoriales"):
                titulos.append(f"{rel}: {m.group(1).strip()}")
        return titulos

    def publicar(self, path: str, titulo: str, resumen: str, cuerpo: str, tags: list[str]) -> int:
        if not self.vivo:
            return 0
        if requests.get(f"{CORTEX}/node/tutoriales", headers=self.h, timeout=15).status_code != 200:
            requests.put(f"{CORTEX}/node/tutoriales", headers=self.h, timeout=30, json={
                "title": "Tutoriales", "tags": ["tutorial"],
                "summary": "Guías paso a paso generadas por tutor.py a partir de la base de conocimiento, con citas y avisos de confiabilidad.",
                "body": "Cada nodo hijo es un tutorial generado automáticamente. Ver `tutor.py` y la carpeta `tutoriales/`.",
                "reason": "rama para los tutoriales generados"})
        r = requests.put(f"{CORTEX}/node/{path}", headers=self.h, timeout=30, json={
            "title": titulo[:120], "summary": resumen[:300], "body": cuerpo, "tags": tags[:20],
            "reason": "Tutorial generado por tutor.py (planificar → recuperar → evaluar → generar)"})
        return r.status_code


# ─────────────────────────────── el grafo ──────────────────────────────────
SIS_PLAN = """Eres el planificador de tutoriales del ERP S10 (software peruano para constructoras).
Descompones una tarea de usuario en los pasos que hay que hacer en S10, en orden.
Devuelve SOLO JSON: {"titulo": str, "modulo": str, "pasos": [{"objetivo": str, "consultas": [str, str]}]}
Entre 3 y 8 pasos. Las consultas son frases cortas en español para un buscador por palabras
(usa el vocabulario de S10: partidas, insumos, metrados, hoja del presupuesto, planilla, etc.)."""

SIS_EVAL = """Evalúas si la evidencia recuperada alcanza para explicar un paso de un tutorial de S10.
Devuelve SOLO JSON: {"suficiente": bool, "falta": str, "consultas_extra": [str]}
Suficiente = la evidencia dice concretamente qué hacer (menú, ventana, campo o acción), no solo que el tema existe."""

SIS_GEN = """Redactas tutoriales paso a paso del ERP S10 para usuarios de constructoras, en español neutro.
Reglas estrictas:
- Usa SOLO la evidencia dada. Cita cada afirmación con su marca [F#] o [N#].
- Si un paso no tiene evidencia suficiente, escríbelo igual como paso pero di "No documentado en las fuentes disponibles" y qué habría que consultar.
- Cuando haya cifras legales (UIT, RMV, jornal, porcentajes), di el año de la fuente y pide verificar la vigente.
- No inventes menús, atajos ni nombres de ventanas.
Formato Markdown: "# <título>", un párrafo "Para qué sirve", "## Antes de empezar" (requisitos, si los hay),
"## Pasos" con pasos numerados y sub-pasos, "## Advertencias" y nada más (las fuentes las agrega el sistema).
Sé conciso: entre 600 y 1500 palabras; un sub-paso por acción, sin repetir la misma explicación."""


def slug(t: str) -> str:
    t = unicodedata.normalize("NFKD", t.lower())
    t = "".join(c for c in t if not unicodedata.combining(c))
    return re.sub(r"[^a-z0-9]+", "-", t).strip("-")[:60].rstrip("-")


def generar(tarea: str, motor: Motor, idx: Indice, cx: Cortex) -> Path:
    log(f"▶ {tarea}")
    # 1. planificar
    plan = extraer_json(motor.texto(SIS_PLAN, f"Tarea del usuario: {tarea}", "medium"))
    log(f"  plan: {plan['titulo']} ({len(plan['pasos'])} pasos)")

    # 2-3. recuperar + evaluar, por paso
    evidencia: dict[str, dict] = {}     # id de fragmento -> fragmento
    nodos: dict[str, dict] = {}         # path -> nodo de Cortex
    por_paso = []
    for n, paso in enumerate(plan["pasos"], 1):
        vistos: list[str] = []
        for ronda in range(2):
            consultas = paso["consultas"] if ronda == 0 else paso.get("_extra", [])
            for q in consultas:
                for f in idx.buscar(q, 5):
                    evidencia.setdefault(f["id"], f)
                    if f["id"] not in vistos:
                        vistos.append(f["id"])
                for nd in cx.buscar(q, 2):
                    nodos.setdefault(nd["path"], nd)
            if ronda == 1:
                break
            muestra = "\n\n".join(f"- {evidencia[i]['texto'][:700]}" for i in vistos[:8]) or "(sin evidencia)"
            ev = extraer_json(motor.texto(SIS_EVAL, f"Paso: {paso['objetivo']}\n\nEvidencia:\n{muestra}", "low"))
            if ev.get("suficiente") or not ev.get("consultas_extra"):
                break
            paso["_extra"] = ev["consultas_extra"][:3]
            log(f"  paso {n}: evidencia corta, busco también {paso['_extra']}")
        por_paso.append((paso["objetivo"], vistos[:8]))

    # numerar evidencia en el orden en que la usan los pasos
    orden: list[str] = []
    for _, ids in por_paso:
        orden += [i for i in ids if i not in orden]
    marcas = {fid: f"F{k}" for k, fid in enumerate(orden, 1)}
    marcas_n = {p: f"N{k}" for k, p in enumerate(nodos, 1)}

    bloques = []
    for fid in orden:
        f = evidencia[fid]
        conf = AVISO_CONFIANZA.get(f.get("confianza", "oficial"), f.get("confianza"))
        pag = (f" {f['pagina']}" if str(f.get("pagina", "")).startswith("min") else f" pág. {f['pagina']}") if f.get("pagina") else ""
        bloques.append(f"[{marcas[fid]}] {f.get('titulo')}{pag}\n{f['texto'][:1400]}")
    for p, nd in nodos.items():
        bloques.append(f"[{marcas_n[p]}] (nodo de conocimiento {p}) {nd['title']}\n{nd['body'][:1800]}")
    guia = "\n".join(f"{k}. {obj} -> evidencia sugerida: {', '.join(marcas[i] for i in ids) or 'ninguna'}"
                     for k, (obj, ids) in enumerate(por_paso, 1))

    # 4. generar
    md = motor.texto(
        SIS_GEN,
        f"Tarea: {tarea}\nTítulo propuesto: {plan['titulo']}\n\nPlan:\n{guia}\n\nEvidencia:\n\n" + "\n\n---\n\n".join(bloques),
        "high",
    ).strip()

    # 5. fuentes (las pone el sistema, solo las citadas) y publicar
    citadas = set(re.findall(r"\[(F\d+|N\d+)\]", md))
    fuentes = ["", "## Fuentes", ""]
    for fid in orden:
        if marcas[fid] in citadas:
            f = evidencia[fid]
            pag = ((f", {f['pagina']}" if str(f.get("pagina", "")).startswith("min") else f", pág. {f['pagina']}")
                   if f.get("pagina") else "")
            conf = AVISO_CONFIANZA.get(f.get("confianza", "oficial"), f.get("confianza"))
            fuentes.append(f"- **[{marcas[fid]}]** {f.get('titulo')}{pag} · {f.get('fuente')}")
    for p, nd in nodos.items():
        if marcas_n[p] in citadas:
            fuentes.append(f"- **[{marcas_n[p]}]** nodo de Cortex `{p}` · {nd['title']}")
    confs = Counter(evidencia[fid].get("confianza", "oficial") for fid in orden if marcas[fid] in citadas)
    pie = f"\n\n---\n_Generado por tutor.py el {time.strftime('%Y-%m-%d %H:%M')}._"
    completo = md + "\n".join(fuentes) + pie

    SALIDA.mkdir(exist_ok=True)
    nombre = slug(plan["titulo"])
    destino = SALIDA / f"{nombre}.md"
    destino.write_text(completo, encoding="utf-8")

    no_doc = md.count("No documentado")
    resumen = (f"Tutorial de {plan.get('modulo', 'S10')}: {plan['titulo']}. {len(plan['pasos'])} pasos, "
               f"{len(citadas)} citas" + (f", {no_doc} pasos sin documentar." if no_doc else "."))
    estado = cx.publicar(f"tutoriales/{nombre}", plan["titulo"], resumen, completo,
                         ["tutorial", slug(plan.get("modulo", "s10"))[:30]])
    registrar(tarea, plan["titulo"], destino, idx.huella, len(citadas), no_doc)
    log(f"  ✓ {destino.relative_to(RAIZ)}  ({len(citadas)} citas, {no_doc} sin documentar, Cortex {estado or 'apagado'})")
    return destino


# ─────────────────────────────── modo automático ───────────────────────────
INDICE_TUT = SALIDA / "indice.json"


def cargar_indice() -> dict:
    return json.loads(INDICE_TUT.read_text(encoding="utf-8")) if INDICE_TUT.exists() else {}


def registrar(tarea, titulo, destino, huella, citas, no_doc) -> None:
    d = cargar_indice()
    d[tarea] = {"titulo": titulo, "archivo": destino.name, "huella_kb": huella, "citas": citas,
                "sin_documentar": no_doc, "fecha": time.strftime("%Y-%m-%dT%H:%M:%S")}
    SALIDA.mkdir(exist_ok=True)
    INDICE_TUT.write_text(json.dumps(d, indent=1, ensure_ascii=False), encoding="utf-8")


SIS_TAREAS = """Propones tareas de usuario para tutoriales del ERP S10, a partir del árbol de conocimiento.
Cada tarea es una acción concreta que un usuario de constructora quiere lograr ("Crear un presupuesto de obra desde cero",
"Calcular la CTS de un trabajador de construcción civil"). Solo tareas que el árbol cubra con contenido real.
Devuelve SOLO JSON: {"tareas": [str]}"""


def auto(motor: Motor, idx: Indice, cx: Cortex, maximo: int) -> None:
    hechos = cargar_indice()
    # regenerar lo que se hizo con otra versión de la base (entraron fuentes nuevas)
    viejos = [t for t, v in hechos.items() if v.get("huella_kb") != idx.huella]
    arbol = "\n".join(cx.arbol()) or "(Cortex apagado)"
    ya = "\n".join(f"- {t}" for t in hechos) or "(ninguno)"
    propuestas = extraer_json(motor.texto(
        SIS_TAREAS, f"Árbol de conocimiento:\n{arbol}\n\nTutoriales ya hechos (no repetir):\n{ya}\n\nPropón {maximo} tareas.", "medium"))["tareas"]
    cola = viejos + [t for t in propuestas if t not in hechos]
    log(f"cola: {len(cola)} tutoriales ({len(viejos)} por base actualizada, {len(cola) - len(viejos)} nuevos); se hacen hasta {maximo}")
    for tarea in cola[:maximo]:
        try:
            generar(tarea, motor, idx, cx)
        except Exception as e:  # uno roto no frena al resto
            log(f"  ✗ {tarea}: {e}")


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="modo", required=True)
    p = sub.add_parser("preguntar")
    p.add_argument("tarea")
    a2 = sub.add_parser("auto")
    a2.add_argument("--max", type=int, default=6)
    sub.add_parser("listar")
    a = ap.parse_args()

    if a.modo == "listar":
        for t, v in cargar_indice().items():
            print(f"{v['fecha'][:16]}  {v['citas']:3d} citas  {v['sin_documentar']} s/d  {v['archivo']}  <- {t}")
        return
    motor, idx, cx = Motor(), Indice(), Cortex()
    log(f"motor {motor.nombre} · {len(idx.docs)} fragmentos · Cortex {'en línea' if cx.vivo else 'apagado'}")
    if a.modo == "preguntar":
        generar(a.tarea, motor, idx, cx)
    else:
        auto(motor, idx, cx, a.max)


if __name__ == "__main__":
    main()
