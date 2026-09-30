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
    return {"titulo": titulo, "resumen": resumen, "tags": tags}


DEPENDE = re.compile(r"\b(ac[aá]|aqu[ií]|este m[oó]dulo|esta secci[oó]n|el caso|ese|esa|eso|esto|dicho|mencionad[oa]|arriba|sqlglot|sql|tabla|columna|join|nodo|cortex|upn|fragmento|curso|s[ií]labo|comando|select|validaci[oó]n|temas trae|manual de|portal de ayuda|S\/ ?[0-9]|ahorros)\b", re.I)
ORDEN_TIPO = {"procedimiento": 0, "problema": 1, "concepto": 2, "dato": 3, "comparacion": 4, "contexto": 5}


def buenas(preguntas: list[dict]) -> list[str]:
    """Hasta POR_TEMA preguntas autónomas (se entienden sin contexto), claras y de tipos variados."""
    cand = [q for q in preguntas
            if 35 <= len(q["pregunta"]) <= 100 and q["pregunta"].startswith("¿") and q["pregunta"].endswith("?")
            and not DEPENDE.search(q["pregunta"])]
    cand.sort(key=lambda q: (ORDEN_TIPO.get(q["tipo"], 9), abs(len(q["pregunta"]) - 62), q["id"]))
    elegidas, tipos, inicios = [], {}, set()
    for ronda in (1, 2, 9):                          # rondas: variedad de tipo y de arranque
        for q in cand:
            if len(elegidas) >= POR_TEMA:
                break
            inicio = " ".join(q["pregunta"].lower().split()[:2])
            if q in elegidas or tipos.get(q["tipo"], 0) >= ronda or (ronda < 9 and inicio in inicios):
                continue
            elegidas.append(q)
            tipos[q["tipo"]] = tipos.get(q["tipo"], 0) + 1
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


def pregunta_de_manual(titulo):
    """Sugerencia sintetizada para un manual oficial (sin banco)."""
    corto = titulo[:60].rstrip()
    p = f"¿Qué explica el manual de {corto}?"
    return [p if p.endswith("?") else p + "?"]


def submodulos_oficiales():
    """Manuales oficiales agrupados por módulo (etiqueta del nodo). Las
    preguntas salen del banco (rama = módulo): procesos reales, no genéricos.
    Sin banco para el módulo, se sintetiza una por manual."""
    base = ARBOL / "manuales-oficiales"
    if not base.is_dir():
        return {}
    banco = {}
    if BANCO.exists():
        for l in BANCO.read_text(encoding="utf-8").splitlines():
            q = json.loads(l)
            banco.setdefault(q.get("rama"), []).append(q)
    por_mod = {}
    for f in sorted(base.glob("*.md")):
        n = leer_nodo(f)
        if not n or not n["titulo"]:
            continue
        mod = next((m for m, _ in MODULOS_OFICIALES if m in n["tags"]), "varios")
        por_mod.setdefault(mod, []).append((f.stem, n))
    etiquetas = dict(MODULOS_OFICIALES)
    etiquetas["varios"] = "Varios"
    out = {}
    for mod, hijos in por_mod.items():
        candidatas = buenas(banco.get(mod, []))
        temas = []
        for i, (stem, n) in enumerate(hijos):
            if candidatas:
                qs = [candidatas[(i + j) % len(candidatas)] for j in range(min(2, len(candidatas)))]
            else:
                qs = pregunta_de_manual(n["titulo"])
            temas.append({"id": f"manuales-oficiales/{stem}",
                          "titulo": n["titulo"][:80], "preguntas": qs})
        out[f"oficial-{mod}"] = {"id": f"oficial-{mod}", "nombre": etiquetas[mod],
                                 "titulo": f"Manuales oficiales: {etiquetas[mod]}",
                                 "resumen": "", "temas": temas}
    return out


def main() -> None:
    banco: dict[str, list] = {}
    if BANCO.exists():
        for l in BANCO.read_text(encoding="utf-8").splitlines():
            q = json.loads(l)
            banco.setdefault(q["nodo"], []).append(q)

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
        qs = ejemplos_de_reporte(f) if rama == "reportes-sql" else buenas(banco.get(path, []))
        if qs:
            modulos[rama]["temas"].append({"id": path, "titulo": n["titulo"].replace(" · validada", ""), "preguntas": qs})
    for m in modulos.values():                       # preguntas del nodo de rama, como tema «General»
        if m["id"] == "reportes-sql":
            continue
        qs = buenas(banco.get(m["id"], []))
        if qs:
            m["temas"].insert(0, {"id": m["id"], "titulo": "General", "preguntas": qs})

    grupos, usados = [], set()
    if "manuales-oficiales" in modulos:
        oficial = submodulos_oficiales()
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
    destacadas = [{"modulo": m["id"], "pregunta": m["temas"][min(1, len(m["temas"]) - 1)]["preguntas"][0]}
                  for g in grupos for m in g["modulos"]]
    total = sum(len(t["preguntas"]) for g in grupos for m in g["modulos"] for t in m["temas"])
    datos = {"generado": time.strftime("%Y-%m-%dT%H:%M:%S"), "total_preguntas": total,
             "grupos": grupos, "destacadas": destacadas}
    SALIDA.parent.mkdir(exist_ok=True)
    SALIDA.write_text(json.dumps(datos, ensure_ascii=False, indent=1), encoding="utf-8")
    print(f"{len(grupos)} grupos · {sum(len(g['modulos']) for g in grupos)} módulos · "
          f"{sum(len(m['temas']) for g in grupos for m in g['modulos'])} temas · {total} preguntas -> {SALIDA.relative_to(RAIZ)}")


if __name__ == "__main__":
    main()
