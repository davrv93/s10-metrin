#!/usr/bin/env python3
"""Arma la rama `manuales-oficiales` de Cortex con lo bajado del portal de miembros.

Un nodo por página del portal (documentacion.s10peru.com) que se abrió con la
cuenta de miembro: título, módulo, PDF que enlaza, índice de secciones sacado del
texto y enlace a la fuente. El texto completo NO va al nodo: ya está troceado en
kb/fragmentos.jsonl para el RAG. Cortex guarda el mapa, no la copia.

Salida: data/cortex-borradores/manuales-oficiales.json (mismo formato que los
demás borradores). Con --cargar lo sube al Cortex local (herramientas/cargar_cortex.py).

Uso: python3 oficial_a_cortex.py [--cargar]
"""
import json
import re
import subprocess
import sys
from pathlib import Path

RAIZ = Path(__file__).resolve().parent
DATA = RAIZ / "data"
KB = RAIZ / "kb"
SALIDA = DATA / "cortex-borradores" / "manuales-oficiales.json"
RAMA = "manuales-oficiales"
PORTAL = "documentacion.s10peru.com"

# palabra en el título -> rama de Cortex del módulo (la más específica primero)
MODULOS = [
    ("portal", "portales"), ("presupuest", "presupuestos"), ("almac", "almacenes"),
    ("compra", "compras"), ("pedido", "compras"), ("proveedor", "compras"),
    ("contab", "contabilidad"), ("factur", "facturacion"), ("venta", "facturacion"),
    ("nómina", "nominas"), ("nomina", "nominas"), ("planilla", "nominas"), ("asistencia", "nominas"),
    ("tareo", "gerencia-proyectos"), ("lookahead", "gerencia-proyectos"), ("calidad", "gerencia-proyectos"),
    ("proyecto", "gerencia-proyectos"), ("administra", "administrativo"), ("tesorer", "administrativo"),
    ("caja", "administrativo"), ("detracci", "administrativo"), ("instala", "soporte"), ("actualiza", "soporte"),
]


def leer_jsonl(p):
    if not p.exists():
        return []
    return [json.loads(l) for l in p.read_text(encoding="utf-8").splitlines() if l.strip()]


def slug(t, largo=60):
    t = t.lower()
    for a, b in zip("áéíóúüñ", "aeiouun"):
        t = t.replace(a, b)
    return re.sub(r"[^a-z0-9]+", "-", t).strip("-")[:largo].strip("-") or "sin-nombre"


def modulo(titulo):
    t = titulo.lower()
    return next((m for clave, m in MODULOS if clave in t), None)


def texto_pdf(pdf):
    ocr = DATA / "ocr" / (pdf.stem + ".txt")
    if ocr.exists():
        return ocr.read_text(encoding="utf-8")
    r = subprocess.run(["pdftotext", "-layout", "-enc", "UTF-8", str(pdf), "-"], capture_output=True, text=True)
    return r.stdout if r.returncode == 0 else ""


TITULAR = re.compile(r"^\s*((\d+(\.\d+){0,2})[.)]?\s+[A-ZÁÉÍÓÚÑ][^\n]{3,70})\s*$")


def secciones(texto, maximo=25):
    """Encabezados numerados («1.2 Registro de pedidos»), sin repetir los del índice del PDF."""
    vistos, salida = set(), []
    for linea in texto.splitlines():
        m = TITULAR.match(linea)
        if not m:
            continue
        t = re.sub(r"[ .]{3,}\s*\d+$", "", m.group(1)).strip()   # quita «....... 12» del índice
        clave = t.lower()
        if clave in vistos or len(t) < 6:
            continue
        vistos.add(clave)
        salida.append(t)
        if len(salida) == maximo:
            break
    return salida


def resumen(texto, titulo):
    lineas = [l.strip() for l in texto.splitlines()
              if len(l.strip()) > 40 and not l.startswith(("#", "Fuente:")) and titulo not in l]
    r = " ".join(lineas)
    if len(r) > 280:
        r = r[:280].rsplit(" ", 1)[0] + "…"
    return r or f"Manual oficial de S10 publicado en el portal de miembros: {titulo}."


def main():
    paginas = [p for p in leer_jsonl(DATA / "paginas.jsonl")
               if PORTAL in p["pagina"] and p.get("estado") == 200 and not p.get("muro_de_miembros")
               and p.get("con_sesion")]
    pdfs = {}
    for m in leer_jsonl(DATA / "manifiesto.jsonl"):
        if m.get("estado") == "ok" and PORTAL in m.get("pagina", ""):
            pdfs.setdefault(m["pagina"], []).append(m)
    docs = {d["fuente"]: d for d in json.loads((KB / "documentos.json").read_text(encoding="utf-8"))} \
        if (KB / "documentos.json").exists() else {}

    nodos, filas = [], []
    usados = set()
    for p in sorted(paginas, key=lambda p: p.get("titulo") or ""):
        titulo = (p.get("titulo") or "").replace("Acceso a ", "").replace("Acceso ", "").strip() or p["pagina"]
        base = slug(re.sub(r"^manual( de)? ", "", titulo, flags=re.I))
        ruta, n = base, 2
        while ruta in usados:
            ruta, n = f"{base}-{n}", n + 1
        usados.add(ruta)
        texto_web = (DATA / p["archivo"]).read_text(encoding="utf-8")
        mod = modulo(titulo)
        cuerpo = [f"Manual oficial de S10, bajado del portal de miembros con la cuenta de la empresa.",
                  "", f"- **Página:** {p['pagina']}"]
        if mod:
            cuerpo.append(f"- **Módulo:** `{mod}` (ver esa rama para el resumen redactado)")
        texto_total = texto_web
        adjuntos = pdfs.get(p["pagina"], [])
        if adjuntos:
            cuerpo += ["", "## PDF", "", "| Documento | Páginas | Fragmentos en la KB |", "|---|---|---|"]
            for m in adjuntos:
                d = docs.get(m["url"], {})
                cuerpo.append(f"| [{m.get('etiqueta') or Path(m['archivo']).stem}]({m['url']}) "
                              f"| {d.get('paginas', '?')} | {d.get('fragmentos', '?')} |")
                pdf = DATA / m["archivo"]
                if pdf.exists():
                    texto_total += "\n" + texto_pdf(pdf)
        indice = secciones(texto_total)
        if indice:
            cuerpo += ["", "## Contenido", ""] + [f"- {s}" for s in indice]
        cuerpo += ["", "El texto completo está en la base de fragmentos de Metrín (`kb/fragmentos.jsonl`) "
                   "y se cita con el nombre del manual y la página.", "", "**Confiabilidad:** oficial S10"]
        nodos.append({
            "path": f"{RAMA}/{ruta}", "title": titulo[:120], "summary": resumen(texto_web, titulo)[:300],
            "body": "\n".join(cuerpo), "tags": ["oficial-s10", "manual", "portal-miembros"] + ([mod] if mod else []),
        })
        filas.append(f"| [{titulo}]({p['pagina']}) | {mod or '—'} | {len(adjuntos)} | `{RAMA}/{ruta}` |")

    if not nodos:
        sys.exit("No hay páginas del portal bajadas con sesión de miembro. Corre antes: s10kb.py oficial")
    raiz = {
        "path": RAMA, "title": "Manuales oficiales S10 (portal de miembros)",
        "summary": (f"{len(nodos)} manuales oficiales de documentacion.s10peru.com bajados con la cuenta de "
                    "miembro. Es la fuente de mayor confianza: si otra fuente la contradice, manda esta."),
        "body": "\n".join([
            "Base oficial de conocimiento de S10. Cada hijo es una página del portal de miembros con sus PDF, "
            "su índice y el enlace a la fuente. Se regenera con `s10kb.py oficial`.", "",
            "## Prioridad", "",
            "1. **Manuales oficiales** (esta rama) y PDF públicos de s10peru.com.",
            "2. Videos del canal oficial.",
            "3. Casos académicos y copias de terceros: solo si lo oficial no cubre el tema, y avisando.", "",
            "## Manuales", "", "| Manual | Módulo | PDF | Nodo |", "|---|---|---|---|", *filas, "",
            "**Confiabilidad:** oficial S10"]),
        "tags": ["oficial-s10", "manuales", "fuentes"],
    }
    SALIDA.parent.mkdir(parents=True, exist_ok=True)
    SALIDA.write_text(json.dumps([raiz, *nodos], indent=1, ensure_ascii=False) + "\n", encoding="utf-8")
    print(f"{len(nodos)} manuales -> {SALIDA.relative_to(RAIZ)}")
    if "--cargar" in sys.argv:
        if not (RAIZ / ".cortex" / ".secrets.yaml").exists():
            print("Cortex sin token local (.cortex/.secrets.yaml): cárgalo luego con "
                  f"herramientas/cargar_cortex.py {SALIDA.relative_to(RAIZ)}")
            return
        r = subprocess.run([sys.executable, str(RAIZ / "herramientas" / "cargar_cortex.py"), str(SALIDA)], cwd=RAIZ)
        sys.exit(r.returncode)


if __name__ == "__main__":
    main()
