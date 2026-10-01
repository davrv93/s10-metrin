#!/usr/bin/env python3
"""Sync cortex_semillas from SQLite to Cortex file-based tree (.cortex/tree/)."""

from __future__ import annotations

import json
import sqlite3
from pathlib import Path
from datetime import datetime, timezone


DB_PATH = Path(__file__).resolve().parent.parent.parent / "data" / "preguntas.db"
CORTEX_TREE = Path(__file__).resolve().parent.parent.parent / ".cortex" / "tree"
SEMILLAS_ROOT = CORTEX_TREE / "semillas-generadas"


def slug(texto: str, largo: int = 80) -> str:
    import re
    s = re.sub(r"[^a-zA-Z0-9]+", "-", texto).strip("-").lower()
    return s[:largo] or "sin-nombre"


def generar_id() -> str:
    import random, string
    return "".join(random.choices(string.ascii_uppercase + string.digits, k=26))


def yaml_str(value: str) -> str:
    escaped = value.replace("\\", "\\\\").replace('"', '\\"').replace("\n", "\\n")
    return f'"{escaped}"'


def sync() -> None:
    if not DB_PATH.exists():
        raise SystemExit(f"DB no encontrada: {DB_PATH}")

    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    rows = conn.execute("""
        WITH primera_candidata AS (
            SELECT nodo_id, modulo, seccion, nodo
            FROM metrin_preguntas_candidatas
            GROUP BY nodo_id
        )
        SELECT s.nodo_id, s.pregunta, s.respuesta_esperada, s.tipo,
               s.jev_claridad, s.jev_dificultad, s.jev_tipo_conf, s.rubricas, s.vb_at,
               c.modulo, c.seccion, c.nodo
        FROM cortex_semillas s
        LEFT JOIN primera_candidata c ON c.nodo_id = s.nodo_id
        WHERE s.vb = 1
        ORDER BY s.nodo_id, s.vb_at
    """).fetchall()

    if not rows:
        raise SystemExit("No hay semillas con VB para sincronizar")

    agrupadas: dict[str, list[sqlite3.Row]] = {}
    for r in rows:
        agrupadas.setdefault(r["nodo_id"], []).append(r)

    print(f"Semillas a sincronizar: {len(rows)}")
    print(f"Nodos únicos: {len(agrupadas)}")
    print(f"Destino: {SEMILLAS_ROOT}")

    creados = 0
    actualizados = 0

    for nodo_id, semillas in agrupadas.items():
        partes = nodo_id.split("/")
        if len(partes) >= 3:
            modulo = partes[0]
            seccion = partes[1]
            nombre = partes[2]
        else:
            modulo = partes[0] if partes else "sin-modulo"
            seccion = partes[1] if len(partes) > 1 else modulo
            nombre = partes[2] if len(partes) > 2 else slug(nodo_id)

        rel_path = SEMILLAS_ROOT / modulo / seccion / f"{nombre}.md"
        rel_path.parent.mkdir(parents=True, exist_ok=True)

        primera = semillas[0]
        title = f"Semillas — {primera['modulo'] or modulo} › {primera['seccion'] or seccion} › {primera['nodo'] or nombre}"
        summary = f"{len(semillas)} pregunta(s) generadas y validadas con VB para {nodo_id}"

        frontmatter = [
            f"title: {yaml_str(title)}",
            f"summary: {yaml_str(summary)}",
            "tags:",
            '  - "oficial-s10"',
            '  - "semilla-preguntas"',
            '  - "generacion-automatica"',
            '  - "vb"',
        ]
        if primera["modulo"]:
            frontmatter.append(f'  - "{slug(primera["modulo"])}"')
        frontmatter += [
            f"nodo_id: {yaml_str(nodo_id)}",
            f"modulo: {yaml_str(primera['modulo'] or modulo)}",
            f"seccion: {yaml_str(primera['seccion'] or seccion)}",
            f"nombre: {yaml_str(primera['nodo'] or nombre)}",
            f"vb_at: {yaml_str(primera['vb_at'])}",
        ]

        body_lines = [
            f"# {title}",
            "",
            f"- **Nodo:** `{nodo_id}`",
            f"- **Módulo:** {primera['modulo'] or modulo}",
            f"- **Sección:** {primera['seccion'] or seccion}",
            f"- **Preguntas generadas:** {len(semillas)}",
            f"- **VB otorgado:** {primera['vb_at']}",
            "",
            "---",
            "",
            "## Preguntas",
            "",
        ]

        for i, s in enumerate(semillas, 1):
            rubricas = {}
            if s["rubricas"]:
                try:
                    rubricas = json.loads(s["rubricas"])
                except json.JSONDecodeError:
                    rubricas = {}
            body_lines += [
                f"### {i}. {s['pregunta']}",
                "",
                f"- **tipo:** {s['tipo']}",
                f"- **respuesta_esperada:** {s['respuesta_esperada']}",
                f"- **dificultad:** {s['jev_dificultad']}",
                f"- **claridad:** {s['jev_claridad']}",
                f"- **tipo_conf:** {s['jev_tipo_conf']}",
            ]
            if rubricas:
                body_lines.append("- **rubricas:**")
                for k, v in rubricas.items():
                    if isinstance(v, dict):
                        prob = v.get("prob", v)
                        body_lines.append(f"  - {k}: {prob}")
                    else:
                        body_lines.append(f"  - {k}: {v}")
            body_lines += ["", "---", ""]

        content = "\n".join(["---"] + frontmatter + ["---"] + body_lines)

        existing_id = None
        if rel_path.exists():
            for line in rel_path.read_text(encoding="utf-8").splitlines():
                if line.startswith("id: ") and len(line.split(":", 1)[1].strip()) == 26:
                    existing_id = line.split(":", 1)[1].strip()
                    break

        if existing_id:
            actualizados += 1
        else:
            creados += 1

        rel_path.write_text(content, encoding="utf-8")

    print(f"Creados: {creados}")
    print(f"Actualizados: {actualizados}")
    print(f"Total archivos: {creados + actualizados}")


if __name__ == "__main__":
    sync()
