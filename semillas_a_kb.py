#!/usr/bin/env python3
"""Preguntas validadas por JEV → tarjetas del RAG que apuntan a su sección con capturas.

El pipeline (metrin/pipeline/preguntas.py) genera preguntas por nodo del HTML
oficial, JEV las puntúa y las que pasan el umbral quedan con VB en
`cortex_semillas` (data/preguntas.db) y en Cortex (.cortex/tree/semillas-generadas).
Cada nodo es una sección del manual: la misma que `indexar` guarda en
kb/fragmentos.jsonl con sus pasos y fotos.

Salidas:
  kb/fragmentos_semillas.jsonl  una tarjeta por pregunta: se busca por la pregunta
                                (`busqueda`) y devuelve el texto y los `pasos` con
                                fotos de su sección; `semilla` es su id estable.
  kb/semillas_metrin.json       las preguntas por página del manual, para el menú.

Fuente de las semillas: data/preguntas.db si existe; si no, los .md de Cortex.
Una semilla cuyo nodo no se encuentra en la KB se informa y no se publica:
sin sección no hay respuesta ni fotos que mostrar.

Uso: python3 semillas_a_kb.py
"""
from __future__ import annotations

import hashlib
import json
import re
import sqlite3
import sys
from pathlib import Path
from urllib.parse import unquote

RAIZ = Path(__file__).resolve().parent
DB = RAIZ / "data" / "preguntas.db"
SEMILLAS_MD = RAIZ / ".cortex" / "tree" / "semillas-generadas"
FRAGMENTOS = RAIZ / "kb" / "fragmentos.jsonl"
SALIDA = RAIZ / "kb" / "fragmentos_semillas.jsonl"
MENU = RAIZ / "kb" / "semillas_metrin.json"
FUERA = {"deprecated", "archived", "rechazada"}   # estados de Cortex que no se publican


def slug_pipeline(texto: str, largo: int = 80) -> str:
    """El slug de metrin/pipeline/preguntas.py (nodo_id): sin quitar tildes."""
    s = re.sub(r"[^a-zA-Z0-9]+", "-", texto).strip("-").lower()
    return s[:largo] or "sin-nombre"


def slug_kb(texto: str, largo: int = 80) -> str:
    """El slug de s10kb.py (ids de fragmento)."""
    s = re.sub(r"[^a-zA-Z0-9]+", "-", unquote(texto)).strip("-").lower()
    return s[:largo] or "sin-nombre"


def indice_de(nodo_id: str) -> int | None:
    m = re.search(r"-(\d{3})$", nodo_id)
    return int(m.group(1)) if m else None


def semillas_db() -> list[dict]:
    conn = sqlite3.connect(DB)
    conn.row_factory = sqlite3.Row
    filas = conn.execute("""
        WITH c AS (SELECT nodo_id, MIN(url) AS url, MIN(modulo) AS modulo, MIN(seccion) AS seccion, MIN(nodo) AS nodo
                   FROM metrin_preguntas_candidatas GROUP BY nodo_id)
        SELECT s.id, s.nodo_id, s.pregunta, s.tipo, s.vb_at, s.jev_claridad, s.rubricas,
               c.url, c.modulo, c.seccion, c.nodo
        FROM cortex_semillas s LEFT JOIN c ON c.nodo_id = s.nodo_id
        WHERE s.vb = 1 ORDER BY s.nodo_id, s.id
    """).fetchall()
    return [dict(f) for f in filas]


def _frontmatter(texto: str) -> tuple[dict, str]:
    fm, cuerpo = {}, texto
    if texto.startswith("---"):
        fin = texto.find("\n---", 3)
        if fin > 0:
            for linea in texto[3:fin].splitlines():
                if ":" in linea and not linea.startswith((" ", "-")):
                    k, v = linea.split(":", 1)
                    v = v.strip()
                    if v.startswith('"') and v.endswith('"'):
                        v = json.loads(v)
                    fm[k.strip()] = v
            cuerpo = texto[fin + 4:]
    return fm, cuerpo


def semillas_md() -> list[dict]:
    out = []
    for md in sorted(SEMILLAS_MD.rglob("*.md")):
        fm, cuerpo = _frontmatter(md.read_text(encoding="utf-8"))
        if not fm.get("nodo_id") or str(fm.get("status", "")).lower() in FUERA:
            continue
        bloques = re.split(r"^### \d+\. ", cuerpo, flags=re.M)[1:]
        for b in bloques:
            pregunta = b.splitlines()[0].strip()
            tipo = re.search(r"^- \*\*tipo:\*\* (.+)$", b, re.M)
            out.append({"nodo_id": fm["nodo_id"], "pregunta": pregunta, "tipo": tipo.group(1).strip() if tipo else None,
                        "url": fm.get("url"), "modulo": fm.get("modulo"), "seccion": fm.get("seccion"),
                        "nodo": fm.get("nombre"), "vb_at": fm.get("vb_at")})
    return out


def cargar_secciones() -> tuple[dict, dict]:
    """Fragmentos de sección del HTML oficial: por id y por (manual, sección) en slug del pipeline."""
    por_id, por_titulo = {}, {}
    for linea in FRAGMENTOS.read_text(encoding="utf-8").splitlines():
        f = json.loads(linea)
        if f.get("documento") != "web" or "seccion" not in f:
            continue
        por_id[f["id"]] = f
        clave = (slug_pipeline(f.get("manual") or ""), slug_pipeline(f.get("seccion") or ""))
        por_titulo.setdefault(clave, []).append(f)
    return por_id, por_titulo


def buscar_seccion(s: dict, por_id: dict, por_titulo: dict) -> dict | None:
    k = indice_de(s["nodo_id"])
    if s.get("url") and k is not None:
        f = por_id.get(f"web-{slug_kb(s['url'], 40)}-s{k:03d}")
        if f:
            return f
    partes = s["nodo_id"].split("/")
    nombre = re.sub(r"-\d{3}$", "", partes[-1])
    modulo = slug_pipeline(s["modulo"]) if s.get("modulo") else partes[0]
    candidatos = por_titulo.get((modulo, nombre), [])
    if len(candidatos) > 1 and k is not None:
        exactos = [f for f in candidatos if f["id"].endswith(f"-s{k:03d}")]
        candidatos = exactos or candidatos
    return candidatos[0] if candidatos else None


def main() -> None:
    if DB.exists():
        semillas, origen = semillas_db(), DB.relative_to(RAIZ)
    elif SEMILLAS_MD.exists():
        semillas, origen = semillas_md(), SEMILLAS_MD.relative_to(RAIZ)
    else:
        sys.exit("No hay semillas: falta data/preguntas.db y .cortex/tree/semillas-generadas/. "
                 "Corre antes el pipeline (preguntas.py juzgar → vb → sync_cortex_semillas.py).")
    por_id, por_titulo = cargar_secciones()
    filas, menu, huerfanas, vistas = [], {}, [], set()
    for s in semillas:
        sec = buscar_seccion(s, por_id, por_titulo)
        if not sec:
            huerfanas.append(s["nodo_id"])
            continue
        clave = (sec["id"], s["pregunta"].strip().lower())
        if clave in vistas:
            continue
        vistas.add(clave)
        sid = "semilla-" + hashlib.sha1(f"{s['nodo_id']}|{s['pregunta']}".encode()).hexdigest()[:16]
        fotos = sum(len(p.get("fotos", [])) for p in sec.get("pasos") or [])
        filas.append({
            "id": sid, "documento": "semilla", "semilla": sid, "tipo": s.get("tipo") or "",
            "manual": sec.get("manual"), "titulo": sec.get("titulo"), "seccion": sec.get("seccion"),
            "pagina": None, "fuente": sec.get("fuente"), "confianza": "oficial",
            "busqueda": s["pregunta"], "texto": sec["texto"], "pasos": sec.get("pasos") or [],
            "nodo_id": s["nodo_id"],
        })
        menu.setdefault(sec.get("fuente"), []).append({"id": sid, "pregunta": s["pregunta"], "tipo": s.get("tipo") or "",
                                                     "seccion": sec.get("seccion"), "fotos": fotos})
    SALIDA.write_text("".join(json.dumps(f, ensure_ascii=False) + "\n" for f in filas), encoding="utf-8")
    MENU.write_text(json.dumps(menu, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    con_fotos = sum(1 for f in filas if any(p.get("fotos") for p in f["pasos"]))
    print(f"{len(semillas)} semillas con VB desde {origen}: {len(filas)} publicadas "
          f"({con_fotos} con capturas) en {len(menu)} páginas; {len(huerfanas)} sin sección en la KB")
    for n in sorted(set(huerfanas))[:10]:
        print(f"  sin sección: {n}")


if __name__ == "__main__":
    main()
