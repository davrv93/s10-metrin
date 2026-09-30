#!/usr/bin/env python3
"""Exporta el conocimiento Cortex (.cortex/tree + items con sustancia) al
formato fragmentos JSONL que indexa `rag index-kb`.

Salida: kb/fragmentos_cortex.jsonl (docker-entrada.sh lo concatena solo).
IDs estables por ruta: reejecutar no duplica; index-kb es incremental.
Se omiten drafts/, rules/ y activity/ (ruido operativo, no conocimiento).
"""
import hashlib
import json
import re
from pathlib import Path

RAIZ = Path(__file__).resolve().parent
CORTEX = RAIZ / ".cortex"
SALIDA = RAIZ / "kb" / "fragmentos_cortex.jsonl"

OMITIR = {"drafts", "rules", "activity"}


def partir(path):
    """(frontmatter dict simple, cuerpo). Solo claves planas que nos importan."""
    texto = path.read_text(encoding="utf-8")
    fm = {}
    cuerpo = texto
    if texto.startswith("---"):
        fin = texto.find("\n---", 3)
        if fin > 0:
            for linea in texto[3:fin].strip().splitlines():
                if ":" in linea and not linea.startswith((" ", "-", "\t")):
                    k, v = linea.split(":", 1)
                    fm[k.strip()] = v.strip().strip("'\"")
            cuerpo = texto[fin + 4:].strip()
    return fm, cuerpo


def titulo_de(fm, cuerpo, path):
    if fm.get("title"):
        return fm["title"]
    m = re.search(r"^#{1,2}\s+(.+)$", cuerpo, re.M)
    if m:
        return m.group(1).strip()
    return path.stem.replace("-", " ")


def main():
    filas = []
    for md in sorted((CORTEX / "tree").rglob("*.md")):
        rel = md.relative_to(CORTEX).as_posix()
        if rel.split("/")[1] in OMITIR:
            continue
        fm, cuerpo = partir(md)
        if fm.get("status") in {"deprecated", "archived"}:
            continue
        if len(cuerpo) < 80:
            continue
        rama = rel.split("/")[1] if "/" in rel else ""
        extra = []
        if fm.get("status"):
            extra.append(f"Estado: {fm['status']}.")
        if fm.get("tags"):
            extra.append(f"Etiquetas: {fm['tags']}.")
        if fm.get("type"):
            extra.append(f"Tipo: {fm['type']}.")
        texto = ((" ".join(extra) + "\n" if extra else "") + cuerpo).strip()
        fid = "cortex-" + hashlib.sha1(rel.encode()).hexdigest()[:16]
        filas.append({
            "id": fid,
            "documento": "cortex:" + rel,
            "manual": "Cortex",
            "titulo": titulo_de(fm, cuerpo, md),
            "pagina": 0,
            "fuente": "",
            "texto": texto,
        })
    for item_dir in sorted((CORTEX / "items").iterdir()):
        md = item_dir / "item.md"
        if not md.exists():
            continue
        fm, cuerpo = partir(md)
        if fm.get("status") in {"todo", "cancelled"} or len(cuerpo) < 80:
            continue
        tipo = fm.get("type", "item")
        texto = f"Tipo: {tipo}. Estado: {fm.get('status', '')}.\n{cuerpo}".strip()
        fid = "cortex-" + hashlib.sha1(fm.get("id", item_dir.name).encode()).hexdigest()[:16]
        filas.append({
            "id": fid,
            "documento": f"cortex:items/{item_dir.name}",
            "manual": "Cortex",
            "titulo": fm.get("title", item_dir.name),
            "pagina": 0,
            "fuente": "",
            "texto": texto,
        })
    SALIDA.parent.mkdir(parents=True, exist_ok=True)
    with open(SALIDA, "w", encoding="utf-8") as fh:
        for f in filas:
            fh.write(json.dumps(f, ensure_ascii=False) + "\n")
    h = hashlib.sha256(SALIDA.read_bytes()).hexdigest()[:16]
    print(f"cortex: {len(filas)} fragmentos en {SALIDA.name} sha256={h}")


if __name__ == "__main__":
    main()
