#!/usr/bin/env python3
"""Exporta el conocimiento Cortex (.cortex/tree + items con sustancia) al
formato fragmentos JSONL que indexa `rag index-kb`.

Salida: kb/fragmentos_cortex.jsonl (docker-entrada.sh lo concatena solo).
IDs estables por ruta: reejecutar no duplica; index-kb es incremental.
Se omiten drafts/, rules/ y activity/ (ruido operativo, no conocimiento).

También arma kb/fragmentos_faq.jsonl: una tarjeta por pregunta del banco
(data/preguntas/preguntas.jsonl). El campo `busqueda` (la pregunta y sus
escrituras con errores) es lo que se embebe; `texto` es el pasaje del nodo de
Cortex que la responde. Así «¿Cómo llego a adicionar correctivo…?» encuentra
el paso exacto del nodo aunque el trozo de 800 caracteres que lo contiene
hable de otras cosas. Misma cita que el nodo: el RAG no duplica la fuente.
"""
import hashlib
import json
import math
import re
from pathlib import Path

RAIZ = Path(__file__).resolve().parent
CORTEX = RAIZ / ".cortex"
SALIDA = RAIZ / "kb" / "fragmentos_cortex.jsonl"
BANCO = RAIZ / "data" / "preguntas" / "preguntas.jsonl"
SALIDA_FAQ = RAIZ / "kb" / "fragmentos_faq.jsonl"
PASAJE = 1200  # runas máximas del pasaje que acompaña a cada pregunta

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
    faq()


VACIAS = set("""a al algo como con cual cuales cuando de del donde el ella en es esa ese eso esta este
hay la las lo los me mi o para pero por que qué se si sin su sus tengo un una uno y ya yo
cómo cuál dónde cuándo puedo debo hago hacer oye hola""".split())


def fichas(texto):
    t = texto.lower()
    for a, b in zip("áéíóúü", "aeiouu"):
        t = t.replace(a, b)
    return {w[:6] for w in re.findall(r"[a-zñ0-9]{3,}", t) if w not in VACIAS}


def nodo_md(nodo):
    arbol = CORTEX / "tree"
    if not nodo:
        return arbol / "_node.md"
    hoja = arbol / (nodo + ".md")
    return hoja if hoja.exists() else arbol / nodo / "_node.md"


def pasaje(bloques, idf, pregunta):
    """Ventana contigua de bloques (≤ PASAJE) alrededor del que más comparte con la pregunta."""
    q = fichas(pregunta)
    puntos = [sum(idf.get(f, 0) for f in q & fichas(b)) for b in bloques]
    i = max(range(len(bloques)), key=lambda k: (puntos[k], -k))
    ini, fin, largo = i, i + 1, len(bloques[i])
    while True:  # crece hacia el vecino con más puntaje mientras quepa
        opciones = [k for k in (ini - 1, fin) if 0 <= k < len(bloques) and largo + len(bloques[k]) <= PASAJE]
        if not opciones:
            break
        k = max(opciones, key=lambda k: (puntos[k], k == ini - 1))
        largo += len(bloques[k])
        ini, fin = min(ini, k), max(fin, k + 1)
    return "\n".join(bloques[ini:fin])[:PASAJE]


def faq():
    if not BANCO.exists():
        return
    nodos = {}
    filas = []
    for linea in open(BANCO, encoding="utf-8"):
        q = json.loads(linea)
        if q["nodo"] not in nodos:
            md = nodo_md(q["nodo"])
            nodo = None
            if md.exists():
                fm, cuerpo = partir(md)
                if fm.get("status") not in {"deprecated", "archived"}:
                    resumen = fm.get("summary", "")
                    bloques = [b.strip() for b in re.split(r"\n\s*\n|\n(?=\s*(?:[-*]|\d+\.|#)\s)", cuerpo) if b.strip()]
                    bloques = [b[:PASAJE] for b in bloques] or [resumen]
                    df = {}
                    for b in bloques:
                        for f in fichas(b):
                            df[f] = df.get(f, 0) + 1
                    idf = {f: math.log(1 + len(bloques) / n) for f, n in df.items()}
                    nodo = (titulo_de(fm, cuerpo, md), bloques, idf, md.relative_to(CORTEX).as_posix())
            nodos[q["nodo"]] = nodo
        nodo = nodos[q["nodo"]]
        if not nodo:
            continue
        titulo, bloques, idf, rel = nodo
        escrituras = [q["pregunta"]] + [v["texto"] for v in q.get("variantes", [])]
        filas.append({
            "id": "faq-" + hashlib.sha1(q["id"].encode()).hexdigest()[:16],
            "documento": "cortex:" + rel,
            "manual": "Cortex",
            "titulo": titulo,
            "pagina": 0,
            "fuente": "",
            "busqueda": "\n".join(escrituras),
            "texto": titulo + "\n" + pasaje(bloques, idf, " ".join(escrituras)),
        })
    with open(SALIDA_FAQ, "w", encoding="utf-8") as fh:
        for f in filas:
            fh.write(json.dumps(f, ensure_ascii=False) + "\n")
    h = hashlib.sha256(SALIDA_FAQ.read_bytes()).hexdigest()[:16]
    print(f"faq: {len(filas)} preguntas en {SALIDA_FAQ.name} sha256={h}")


if __name__ == "__main__":
    main()
