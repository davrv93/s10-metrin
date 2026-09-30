#!/usr/bin/env python3
"""
Arma kb/catalogo_metrin.json: módulos, temas y preguntas sugeridas para la UI de Metrín.

La agrupación sale de Cortex: cada rama del árbol es un módulo (título y resumen del nodo de rama),
cada nodo hijo es un tema, y las preguntas vienen del banco (data/preguntas/preguntas.jsonl),
que ya sabe a qué nodo pertenece cada una. Solo se usan preguntas bien escritas (nivel 0).

  .venv/bin/python herramientas/catalogo_metrin.py

Metrín lo sirve en GET /catalogo (lo lee de /kb en cada petición: no hace falta reconstruir).
"""
from __future__ import annotations

import json
import re
import time
from pathlib import Path

RAIZ = Path(__file__).resolve().parent.parent
ARBOL = RAIZ / ".cortex" / "tree"
BANCO = RAIZ / "data" / "preguntas" / "preguntas.jsonl"
SALIDA = RAIZ / "kb" / "catalogo_metrin.json"

# Cómo se agrupan las ramas de Cortex en la UI. Una rama nueva que no esté aquí cae en «Más temas».
GRUPOS = [
    ("oficiales", "Manuales oficiales S10", ["manuales-oficiales"]),
    ("empresa", "Optimiza 360", ["optimiza360"]),
]
MODULOS_OFICIALES = [  # subgrupos de manuales-oficiales por etiqueta del nodo
    ("presupuestos", "Presupuestos"), ("gerencia-proyectos", "Gerencia de proyectos"),
    ("compras", "Compras"), ("almacenes", "Almacenes"), ("nominas", "Nóminas"),
    ("contabilidad", "Contabilidad"), ("facturacion", "Facturación"),
    ("administrativo", "Administrativo"), ("portales", "Portales"), ("soporte", "Soporte"),
]
NOMBRES = {  # nombre corto para el menú (el título del nodo de Cortex va como descripción)
    "manuales-oficiales": "Manuales oficiales", "optimiza360": "Optimiza 360",
}
FUERA = {"fuentes", "ui-pantallas"}               # internas del proyecto, no temas para el usuario
PLEGAR = {"ui-pantallas/17w0yki8iew": "nominas"}  # pantallas validadas: se muestran dentro de su módulo
POR_TEMA = 6


def leer_nodo(f: Path) -> dict | None:
    m = re.match(r"---\n(.*?)\n---\n?(.*)", f.read_text(encoding="utf-8"), re.S)
    if not m:
        return None
    fm = m.group(1)
    titulo = (re.search(r"^title:\s*(.+)$", fm, re.M) or [None, ""])[1].strip().strip('"')
    resumen = re.search(r"^summary:\s*(.+?)(?=\n[a-z_]+:|\Z)", fm, re.M | re.S)
    resumen = re.sub(r"\s+", " ", resumen.group(1)).strip().strip('"') if resumen else ""
    tags = re.search(r"^tags:\s*\n((?:\s+-\s*.+\n?)+)", fm, re.M)
    tags = re.findall(r"-\s*(\S+)", tags.group(1)) if tags else []
    return {"titulo": titulo, "resumen": resumen, "tags": tags, "cuerpo": m.group(2)}


DEPENDE = re.compile(r"\b(ac[aá]|aqu[ií]|este m[oó]dulo|esta secci[oó]n|el caso|ese|esa|eso|esto|dicho|mencionad[oa]|arriba|sqlglot|sql|tabla|columna|join|nodo|cortex|upn|fragmento|curso|s[ií]labo|comando|select|validaci[oó]n|temas trae|manual de|portal de ayuda|S\/ ?[0-9]|ahorros)\b", re.I)
ORDEN_TIPO = {"procedimiento": 0, "problema": 1, "concepto": 2, "dato": 3, "comparacion": 4, "contexto": 5}
# Los dos tipos de aprendizaje que muestra la UI: concepto (qué es) y cómo se hace.
TIPO_CONCEPTO = {"concepto"}
TIPO_COMO = {"procedimiento"}
# Título de apartado que describe una acción → pregunta «cómo se hace»; el resto, concepto.
INFINITIVO = re.compile(r"^(crear|configurar|registr(?:ar|e)|ingresar|generar|elaborar|aplicar|realizar|procesar|"
                        r"imprimir|anular|modificar|actualizar|importar|exportar|asignar|definir|habilitar|validar|"
                        r"verificar|calcular|distribuir|conciliar|aprobar|rechazar|revisar|solicitar|ejecutar|"
                        r"parametrizar|instalar|activar|desactivar|cargar|enviar|descargar)\b", re.I)
VERBO = re.compile(r"^(creaci[oó]n|crear|configuraci[oó]n|configurar|registr[ae]|ingresar|ingreso|generar|generaci[oó]n|"
                   r"elaboraci[oó]n|elaborar|aplicaci[oó]n|aplicar|realizaci[oó]n|realizar|procesar|procesamiento|"
                   r"impresi[oó]n|imprimir|anulaci[oó]n|anular|modificaci[oó]n|modificar|actualizaci[oó]n|actualizar|"
                   r"importaci[oó]n|importar|exportaci[oó]n|exportar|asignaci[oó]n|asignar|definir|definici[oó]n|"
                   r"habilitar|validar|validaci[oó]n|verificar|calcular|c[aá]lculo|distribuir|conciliar|aprobar|"
                   r"rechazar|revisar|solicitar|ejecutar|parametrizar|parametrizaci[oó]n|instalaci[oó]n|instalar|"
                   r"activaci[oó]n|activar|desactivar|carga|cargar|env[ií]o|enviar|descarga|descargar)\b", re.I)


def pregunta_apartado(titulo: str) -> str:
    """Apartado del índice → pregunta de aprendizaje: «cómo hacer» si describe acción, concepto si no."""
    if INFINITIVO.match(titulo):
        return f"¿Cómo {titulo[0].lower() + titulo[1:]} en S10?"
    if VERBO.match(titulo):
        return f"¿Cómo hacer «{titulo}» paso a paso?"
    return f"¿Qué es «{titulo}»?"


def valida(q: dict) -> bool:
    p = q["pregunta"]
    return 35 <= len(p) <= 100 and p.startswith("¿") and p.endswith("?") and not DEPENDE.search(p)


def reparto_tema(pool: list[dict], tipos: set[str], cursor: int, n: int = 3) -> tuple[list[str], int]:
    """Toma n preguntas del tipo pedido, rotando el cursor para que temas hermanos no repitan."""
    cand = [q for q in pool if q["tipo"] in tipos and valida(q)]
    if not cand:
        return [], cursor
    out, i, vueltas = [], cursor, 0
    while len(out) < min(n, len(cand)) and vueltas <= len(cand):
        q = cand[i % len(cand)]
        i += 1
        vueltas += 1
        if q["pregunta"] not in out:
            out.append(q["pregunta"])
    return out, i % len(cand) if cand else cursor


def buenas(preguntas: list[dict], tipos: set[str] | None = None,
           vistas: set[str] | None = None) -> list[str]:
    """Hasta POR_TEMA preguntas autónomas (se entienden sin contexto), claras y de tipos variados."""
    cand = [q for q in preguntas
            if 35 <= len(q["pregunta"]) <= 100 and q["pregunta"].startswith("¿") and q["pregunta"].endswith("?")
            and not DEPENDE.search(q["pregunta"])
            and (tipos is None or q["tipo"] in tipos)
            and not (vistas and q["pregunta"] in vistas)]
    cand.sort(key=lambda q: (ORDEN_TIPO.get(q["tipo"], 9), abs(len(q["pregunta"]) - 62), q["id"]))
    elegidas, tipos_vistos, inicios = [], {}, set()
    for ronda in (1, 2, 9):                          # rondas: variedad de tipo y de arranque
        for q in cand:
            if len(elegidas) >= POR_TEMA:
                break
            inicio = " ".join(q["pregunta"].lower().split()[:2])
            if q in elegidas or tipos_vistos.get(q["tipo"], 0) >= ronda or (ronda < 9 and inicio in inicios):
                continue
            elegidas.append(q)
            tipos_vistos[q["tipo"]] = tipos_vistos.get(q["tipo"], 0) + 1
            inicios.add(inicio)
    return [q["pregunta"] for q in elegidas[:POR_TEMA]]


def ejemplos_de_reporte(f: Path) -> list[str]:
    """Los nodos de reportes traen «Ejemplos: a; b» o «Responde preguntas como: a; b»: son preguntas de usuario reales."""
    t = f.read_text(encoding="utf-8")
    m = re.search(r"(?:Ejemplos|Responde preguntas como):\s*(.+?)(?:\.\s|\.\"|\n\n|$)", t, re.S)
    if not m:
        return []
    out = []
    for e in re.split(r";", re.sub(r"\s+", " ", m.group(1))):
        e = e.strip().strip('".')
        if len(e) > 3:
            out.append("¿" + e[0].upper() + e[1:] + "?" if not e.startswith("¿") else e)
    return out


def detalle_manual(n: dict) -> tuple[str, list[dict]]:
    """Enlace e índice secuencial del manual, con una consulta ligada a cada apartado."""
    cuerpo = n["cuerpo"]
    fuente = re.search(r"^\s*-\s+\*\*P[aá]gina:\*\*\s+(https?://\S+)", cuerpo, re.M)
    indice = re.search(r"^## Contenido\s*\n(.*?)(?=^## |\Z)", cuerpo, re.M | re.S)
    lineas = re.findall(r"^\s*-\s+(.+?)\s*$", indice.group(1), re.M) if indice else []
    secciones = []
    for linea in lineas:
        m = re.match(r"^(\d+(?:\.\d+)*)\.?\s+(.+?)\s*$", linea)
        numero, titulo = (m.group(1), m.group(2)) if m else ("", linea.strip())
        texto = f"{numero} {titulo}".strip()
        secciones.append({"numero": numero, "titulo": titulo, "nivel": numero.count(".") + 1 if numero else 1,
                          "preguntas": [pregunta_apartado(titulo)]})
    return (fuente.group(1).rstrip(").,;") if fuente else "", secciones)


def submodulos_oficiales(banco_rama: dict[str, list]) -> dict:
    """Manuales oficiales agrupados por módulo (etiqueta del nodo). Cada tema
    enseña dos listas de aprendizaje: conceptos (qué es) y procedimientos (cómo
    hacerlo), tomadas del banco del módulo y rotadas entre temas hermanos."""
    base = ARBOL / "manuales-oficiales"
    if not base.is_dir():
        return {}
    por_mod = {}
    for f in sorted(base.glob("*.md")):
        if f.stem == "_node":
            continue
        n = leer_nodo(f)
        if not n or not n["titulo"]:
            continue
        mod = next((m for m, _ in MODULOS_OFICIALES if m in n["tags"]), "varios")
        fuente, secciones = detalle_manual(n)
        por_mod.setdefault(mod, []).append((f.stem, n, fuente, secciones))
    etiquetas = dict(MODULOS_OFICIALES)
    etiquetas["varios"] = "Varios"
    out = {}
    for mod, hijos in por_mod.items():
        pool, cursor_c, cursor_p = banco_rama.get(mod, []), 0, 0
        temas = []
        for stem, n, fuente, secciones in hijos:
            conceptos, cursor_c = reparto_tema(pool, TIPO_CONCEPTO, cursor_c)
            como, cursor_p = reparto_tema(pool, TIPO_COMO, cursor_p)
            if not conceptos:
                conceptos = [f"¿Qué es «{n['titulo']}» y qué incluye?"]
            temas.append({"id": f"manuales-oficiales/{stem}", "titulo": n["titulo"][:100],
                          "resumen": n["resumen"][:360], "fuente": fuente,
                          "secciones": secciones, "conceptos": conceptos, "procedimientos": como})
        out[f"oficial-{mod}"] = {"id": f"oficial-{mod}", "nombre": etiquetas[mod],
                                 "titulo": f"Manuales oficiales: {etiquetas[mod]}", "tipo": "manuales",
                                 "resumen": "", "temas": temas}
    return out


def main() -> None:
    banco: dict[str, list] = {}
    banco_rama: dict[str, list] = {}
    if BANCO.exists():
        for l in BANCO.read_text(encoding="utf-8").splitlines():
            q = json.loads(l)
            banco.setdefault(q["nodo"], []).append(q)
            banco_rama.setdefault(q["rama"], []).append(q)

    def aprender(nodo_id: str) -> tuple[list[str], list[str]]:
        """Del banco del nodo: qué es (concepto) y cómo hacerlo (procedimiento)."""
        pool = banco.get(nodo_id, [])
        return buenas(pool, TIPO_CONCEPTO), buenas(pool, TIPO_COMO)

    modulos: dict[str, dict] = {}
    # una rama es `<rama>.md` (sin hijos) o `<rama>/_node.md` (con hijos)
    ramas = {f.stem: f for f in ARBOL.glob("*.md") if f.stem != "_node"}
    ramas.update({d.name: d / "_node.md" for d in ARBOL.iterdir() if d.is_dir() and (d / "_node.md").exists()})
    for rid, f in sorted(ramas.items()):
        if rid in FUERA:
            continue
        n = leer_nodo(f)
        modulos[rid] = {"id": rid, "nombre": NOMBRES.get(rid, n["titulo"]), "titulo": n["titulo"],
                        "resumen": n["resumen"], "temas": []}
    for f in sorted(ARBOL.rglob("*/*.md")):
        if f.stem == "_node":
            continue
        path = f.relative_to(ARBOL).with_suffix("").as_posix()
        rama = PLEGAR.get(path, path.split("/")[0])
        if rama not in modulos or (path.split("/")[0] in FUERA and path not in PLEGAR):
            continue
        n = leer_nodo(f)
        if not n or n["titulo"].startswith("(Reemplazado)"):
            continue
        conceptos, como = aprender(path)
        if conceptos or como:
            modulos[rama]["temas"].append({"id": path, "titulo": n["titulo"].replace(" · validada", ""),
                                           "conceptos": conceptos, "procedimientos": como})
    for m in modulos.values():                       # preguntas del nodo de rama, como tema «General»
        if m["id"] == "reportes-sql":
            continue
        conceptos, como = aprender(m["id"])
        if conceptos or como:
            m["temas"].insert(0, {"id": m["id"], "titulo": "General",
                                  "conceptos": conceptos, "procedimientos": como})

    grupos, usados = [], set()
    if "manuales-oficiales" in modulos:
        oficial = submodulos_oficiales(banco_rama)
        modulos.update(oficial)
        modulos.pop("manuales-oficiales", None)
        GRUPOS[0] = ("oficiales", "Manuales oficiales S10", sorted(oficial))
    for gid, gtitulo, ramas in GRUPOS:
        ms = [modulos[r] for r in ramas if r in modulos and modulos[r]["temas"]]
        usados.update(ramas)
        if ms:
            grupos.append({"id": gid, "titulo": gtitulo, "modulos": ms})
    otros = [m for k, m in modulos.items() if k not in usados and m["temas"]]
    if otros:
        grupos.append({"id": "otros", "titulo": "Más temas", "modulos": otros})

    # destacadas: una pregunta por módulo, para la bienvenida
    destacadas = []
    for g in grupos:
        for m in g["modulos"]:
            tema = next((t for t in m["temas"] if t.get("conceptos") or t.get("procedimientos")), None)
            if tema:
                pregunta = (tema.get("procedimientos") or tema.get("conceptos"))[0]
                destacadas.append({"modulo": m["id"], "pregunta": pregunta})
    total = sum(len(t.get("conceptos", [])) + len(t.get("procedimientos", []))
                for g in grupos for m in g["modulos"] for t in m["temas"])
    manuales = [t for g in grupos for m in g["modulos"] if m.get("tipo") == "manuales" for t in m["temas"]]
    total_manuales = len(manuales)
    total_secciones = sum(len(t["secciones"]) for t in manuales)
    total_preguntas_indice = sum(len(s["preguntas"]) for t in manuales for s in t["secciones"])
    datos = {"generado": time.strftime("%Y-%m-%dT%H:%M:%S"), "total_preguntas": total + total_preguntas_indice,
             "total_preguntas_banco": total,
             "total_manuales": total_manuales, "total_secciones": total_secciones,
             "total_preguntas_indice": total_preguntas_indice,
             "grupos": grupos, "destacadas": destacadas}
    SALIDA.parent.mkdir(exist_ok=True)
    SALIDA.write_text(json.dumps(datos, ensure_ascii=False, indent=1), encoding="utf-8")
    print(f"{len(grupos)} grupos · {sum(len(g['modulos']) for g in grupos)} módulos · "
          f"{sum(len(m['temas']) for g in grupos for m in g['modulos'])} temas · {total_manuales} manuales · "
          f"{total_secciones} apartados secuenciales -> {SALIDA.relative_to(RAIZ)}")


if __name__ == "__main__":
    main()
