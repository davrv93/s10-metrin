#!/usr/bin/env python3
"""Genera la biblioteca de PDFs de S10 por módulo y su página índice.

Uso (desde s10-conocimiento/, en la Mac, con Google Chrome instalado):
    python3 herramientas/biblioteca.py              # todo: procedimientos + manuales + index.html
    python3 herramientas/biblioteca.py --solo-indice

Salida en biblioteca/ (fuera de git: los manuales vienen del portal de miembros de S10):

    biblioteca/
      index.html                         buscador y enlaces (rutas relativas: abre igual en la Mac, Docker o el EC2)
      catalogo.json                      lo mismo en datos
      <modulo>/procedimientos/<id>.pdf   los procedimientos curados de kb/procedimientos/ (pasos, fotos y fuentes)
      <modulo>/manuales/<pagina>.pdf     cada página del portal de documentación de S10 (data/html/), por secciones
      <modulo>/documentos/<pdf>          los PDF oficiales bajados (data/pdf/), copiados tal cual

Las secciones y la posición de cada captura salen de secciones_html.py (lo mismo que indexa s10kb.py); las imágenes,
de data/imagenes/. Necesita beautifulsoup4 y pyyaml (están en entrenamiento-router/.venv).
"""
from __future__ import annotations

import argparse
import html
import json
import re
import shutil
import subprocess
import sys
import tempfile
import time
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from pathlib import Path

import yaml

RAIZ = Path(__file__).resolve().parent.parent
SALIDA = RAIZ / "biblioteca"
CONSTRUCCION = SALIDA / ".construccion"
PROCEDIMIENTOS = RAIZ / "kb" / "procedimientos"
IMAGENES = RAIZ / "data" / "imagenes"
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

# Módulo de cada manual por palabras de su carpeta o archivo (la primera regla que coincide manda).
MODULOS = [
    ("sperant", r"sperant|crm"),
    ("portales", r"portal|colaborador|suministros"),
    ("facturacion-electronica", r"facturacion-electronica|sunat|electronic"),
    ("facturacion", r"facturacion|factura|detraccion|venta|boletaje|titular|minuta|confirming|no-deducibles|anticipo"),
    ("moviles", r"movil|aprobaciones|tablero|muros|lean|lookahead|tareo-en-linea"),
    ("nominas", r"nomina|tareo|reparto|utilidades|cts"),
    ("presupuestos", r"presupuesto"),
    ("gerencia-proyectos", r"gerencia"),
    ("almacenes", r"almacen"),
    ("compras", r"compra|pedido"),
    ("contabilidad", r"contab|cierre|apertura|ingresos.*identificar"),
    ("administrativo", r"administra"),
    ("instalacion", r"instalacion|sql|base-de-datos|basedatos|prerrequ|actualizacion|acceso|copia|descargas|admin"),
]
NOMBRES = {
    "presupuestos": "Presupuestos", "gerencia-proyectos": "Gerencia de Proyectos", "almacenes": "Almacenes",
    "compras": "Compras y Pedidos", "administrativo": "Administrativo", "contabilidad": "Contabilidad",
    "facturacion": "Facturación", "facturacion-electronica": "Facturación electrónica", "nominas": "Nóminas",
    "instalacion": "Instalación y administración", "portales": "Portales (proveedor, colaborador, suministros)",
    "sperant": "Integración con Sperant", "moviles": "Aplicaciones móviles", "otros": "Novedades y otros",
}

CSS = """
@page { size: A4; margin: 16mm 14mm 18mm; }
* { box-sizing: border-box; }
body { font-family: -apple-system, "Segoe UI", Helvetica, Arial, sans-serif; color: #1c2430; font-size: 10.5pt;
       line-height: 1.5; margin: 0; }
header { border-bottom: 2px solid #0f4f99; padding-bottom: 8px; margin-bottom: 14px; }
header .modulo { color: #0f4f99; font-size: 9pt; letter-spacing: .08em; text-transform: uppercase; }
h1 { font-size: 18pt; margin: 4px 0 6px; line-height: 1.2; }
h2 { font-size: 13pt; margin: 18px 0 6px; color: #0f4f99; break-after: avoid; }
h3 { font-size: 11pt; margin: 12px 0 4px; break-after: avoid; }
.meta { color: #5a6676; font-size: 9pt; }
.objetivo { background: #eef4fb; padding: 8px 10px; border-radius: 4px; }
ol.pasos > li { margin: 0 0 10px; }
.resultado { color: #1f6b43; }
figure { margin: 6px 0 10px; break-inside: avoid; }
figure img { max-width: 100%; max-height: 120mm; border: 1px solid #d5dce5; }
figcaption { font-size: 8.5pt; color: #5a6676; }
table { border-collapse: collapse; width: 100%; font-size: 9pt; }
td, th { border: 1px solid #d5dce5; padding: 4px 6px; text-align: left; vertical-align: top; }
section.seccion { break-inside: auto; }
.fuente { font-size: 8.5pt; color: #5a6676; word-break: break-all; }
"""


def modulo_de(nombre: str) -> str:
    for mod, patron in MODULOS:
        if re.search(patron, nombre):
            return mod
    return "otros"










def negritas(t: str) -> str:
    return re.sub(r"\*\*(.+?)\*\*", r"<b>\1</b>", html.escape(t))


def parrafos(t: str) -> str:
    partes = [p.strip() for p in re.split(r"\n\s*\n", t) if p.strip()]
    return "".join(f"<p>{negritas(p)}</p>" for p in partes)


def figuras(imgs: list[Path], caption: str = "") -> str:
    return "".join(f'<figure><img src="{p.as_uri()}">' + (f"<figcaption>{html.escape(caption)}</figcaption>" if caption else "")
                   + "</figure>" for p in imgs)


def pagina(titulo: str, modulo: str, cuerpo: str, meta: str = "") -> str:
    return (f'<!doctype html><html lang="es"><head><meta charset="utf-8"><title>{html.escape(titulo)}</title>'
            f"<style>{CSS}</style></head><body><header><div class=\"modulo\">S10 · {html.escape(NOMBRES.get(modulo, modulo))}</div>"
            f"<h1>{html.escape(titulo)}</h1><div class=\"meta\">{meta}</div></header>{cuerpo}</body></html>")


# ---------------------------------------------------------------- procedimientos curados

def html_procedimiento(d: dict) -> str:
    partes = []
    if d.get("objetivo"):
        partes.append(f'<p class="objetivo">{negritas(d["objetivo"])}</p>')
    if d.get("prerrequisitos"):
        partes.append("<h2>Antes de empezar</h2><ul>" + "".join(
            f"<li>{negritas(p.get('texto', '') if isinstance(p, dict) else str(p))}</li>" for p in d["prerrequisitos"]) + "</ul>")

    def pasos(lista) -> str:
        items = []
        for p in lista or []:
            fotos = [IMAGENES.parent / f["ruta_o_url"] for f in p.get("fotos") or [] if f.get("ruta_o_url")]
            fotos = [f for f in fotos if f.exists()]
            cap = next((f.get("caption") for f in p.get("fotos") or [] if f.get("caption")), "")
            items.append("<li>" + negritas(p.get("accion", "")) +
                         (f'<div class="resultado">→ {negritas(p["resultado"])}</div>' if p.get("resultado") else "") +
                         (f'<ol type="a">{pasos(p["subpasos"])}</ol>' if p.get("subpasos") else "") +
                         figuras(fotos, cap) + "</li>")
        return "".join(items)

    partes.append(f'<h2>Pasos</h2><ol class="pasos">{pasos(d.get("pasos"))}</ol>')
    if d.get("verificacion"):
        v = d["verificacion"]
        v = v if isinstance(v, list) else [v]
        partes.append("<h2>Cómo verificar</h2><ul>" + "".join(
            f"<li>{negritas(x.get('texto', '') if isinstance(x, dict) else str(x))}</li>" for x in v) + "</ul>")
    if d.get("errores_frecuentes"):
        filas = "".join(f"<tr><td>{negritas(e.get('sintoma', e.get('error', '')))}</td><td>{negritas(e.get('solucion', ''))}</td></tr>"
                        for e in d["errores_frecuentes"] if isinstance(e, dict))
        partes.append(f"<h2>Errores frecuentes</h2><table><tr><th>Síntoma</th><th>Solución</th></tr>{filas}</table>")
    if d.get("fuentes"):
        fs = d["fuentes"]
        filas = "".join(f"<tr><td>{html.escape(str(v.get('manual', '')))}</td><td>{html.escape(str(v.get('seccion', v.get('titulo', ''))))}</td></tr>"
                        for v in (fs.values() if isinstance(fs, dict) else fs) if isinstance(v, dict))
        if filas:
            partes.append(f"<h2>Fuentes</h2><table><tr><th>Manual</th><th>Sección</th></tr>{filas}</table>")
    return "".join(partes)


def procedimientos() -> list[dict]:
    items = []
    for ruta in sorted(PROCEDIMIENTOS.glob("*/*.yml")):
        if ruta.parent.name.startswith("_"):
            continue
        d = yaml.safe_load(ruta.read_text(encoding="utf-8"))
        mod = d.get("modulo") or ruta.parent.name
        n = len(d.get("pasos") or [])
        items.append({
            "tipo": "procedimiento", "modulo": mod, "titulo": d.get("titulo", ruta.stem), "id": ruta.stem,
            "archivo": f"{mod}/procedimientos/{ruta.stem}.pdf", "detalle": f"{n} pasos · nivel {d.get('nivel', '—')}",
            "buscar": " ".join([d.get("titulo", "")] + list(d.get("aliases") or []) + list(d.get("preguntas") or [])),
            "_html": pagina(d.get("titulo", ruta.stem), mod, html_procedimiento(d),
                            f"Procedimiento verificado contra el manual oficial · {n} pasos · nivel {d.get('nivel', '—')}"),
        })
    return items


# ---------------------------------------------------------------- manuales oficiales

# Textos de la maqueta del sitio que se cuelan como párrafos.
RESTOS_WEB = {"table of contents", "toggle", "tabla de contenidos", "índice"}


def nivel(titulo: str) -> int:
    m = re.match(r"^(\d+(?:\.\d+)*)", titulo.strip())
    return min(4, 1 + m.group(1).count(".")) + 1 if m else 2


def manuales() -> list[dict]:
    """Una página del portal de documentación de S10 = un PDF, con sus secciones en orden y cada captura junto al paso
    que ilustra (secciones_html.py, lo mismo que usa s10kb.py para indexar)."""
    sys.path.insert(0, str(RAIZ))
    from secciones_html import secciones  # noqa: E402

    fotos_por_url = {}
    for linea in open(RAIZ / "data" / "imagenes.jsonl", encoding="utf-8"):
        x = json.loads(linea)
        fotos_por_url[x["url"]] = x["archivo"]
    items, vistos = [], {}
    for linea in open(RAIZ / "data" / "paginas.jsonl", encoding="utf-8"):
        p = json.loads(linea)
        if p.get("estado") != 200 or p.get("muro_de_miembros"):
            continue
        slug = Path(p["archivo"]).stem
        ruta = RAIZ / "data" / "html" / f"{slug}.html"
        if not ruta.exists() or re.search(r"membership|^home|^index|^panel|sin-nombre|^s10peru-com-(soporte|capacitaciones|implantacion)|"
                                          r"^contacto|^docs$|^inicio$|^prueba-|pagina-ejemplo|sidebar|widgets|^optimiza360-pe", slug):
            continue
        titulo = (p.get("titulo") or slug).strip()
        secs = secciones(ruta.read_text(encoding="utf-8"), p["pagina"], titulo, fotos_por_url)
        n_fotos = sum(len(x.get("fotos") or []) for sec in secs for x in sec["pasos"])
        n_texto = sum(len(x["texto"]) for sec in secs for x in sec["pasos"])
        if not secs or (n_texto < 300 and n_fotos == 0):
            continue
        partes = []
        for sec in secs:
            t = sec.get("titulo") or ""
            cuerpo = ""
            for paso in sec["pasos"]:
                if paso["texto"].strip().lower() in RESTOS_WEB:
                    paso = dict(paso, texto="")
                fotos = [RAIZ / "data" / f for f in paso.get("fotos") or []]
                cuerpo += (f"<p>{negritas(paso['texto'])}</p>" if paso["texto"] else "") + figuras([f for f in fotos if f.exists()])
            if not cuerpo:
                continue
            h = nivel(t)
            partes.append(f'<section class="seccion">' + (f"<h{h}>{html.escape(t)}</h{h}>" if t and t != titulo else "")
                          + cuerpo + "</section>")
        mod = modulo_de(slug)
        clave = (mod, titulo.lower())
        if clave in vistos and vistos[clave]["_peso"] >= n_texto:  # la misma página bajada dos veces: la más completa
            continue
        item = {
            "tipo": "manual", "modulo": mod, "titulo": titulo, "id": slug, "archivo": f"{mod}/manuales/{slug}.pdf",
            "detalle": f"{len(secs)} secciones · {n_fotos} capturas" if len(secs) > 1 else f"{n_fotos} capturas",
            "buscar": titulo + " " + " ".join(sec.get("titulo") or "" for sec in secs),
            "fuente": p["pagina"], "_peso": n_texto,
            "_html": pagina(titulo, mod, "".join(partes),
                            f'Manual oficial de S10 · <span class="fuente">{html.escape(p["pagina"])}</span>'),
        }
        if clave in vistos:
            items.remove(vistos[clave])
        vistos[clave] = item
        items.append(item)
    return items


def documentos() -> list[dict]:
    """Los PDF oficiales bajados tal cual (data/pdf/): se copian, no se regeneran."""
    items = []
    for ruta in sorted((RAIZ / "data" / "pdf").glob("*.pdf")):
        mod = modulo_de(ruta.stem)
        titulo = re.sub(r"[-_]+", " ", ruta.stem).strip().capitalize()
        items.append({"tipo": "documento", "modulo": mod, "titulo": titulo, "id": ruta.stem,
                      "archivo": f"{mod}/documentos/{ruta.name}", "detalle": f"PDF original · {ruta.stat().st_size // 1024} KB",
                      "buscar": titulo, "_copiar": ruta})
    return items


# ---------------------------------------------------------------- PDF e índice

def a_pdf(item: dict) -> str:
    destino = SALIDA / item["archivo"]
    destino.parent.mkdir(parents=True, exist_ok=True)
    if item.get("_copiar"):
        shutil.copyfile(item["_copiar"], destino)
        return ""
    fuente = CONSTRUCCION / (item["archivo"].replace("/", "__") + ".html")
    fuente.write_text(item["_html"], encoding="utf-8")
    # Perfil propio por llamada: varios Chrome headless con el mismo perfil se bloquean entre sí. Chrome a veces
    # escribe el PDF y no termina: se da por hecho cuando el archivo existe y su tamaño no cambia en 2 s, y se cierra.
    for _ in (1, 2):
        destino.unlink(missing_ok=True)
        with tempfile.TemporaryDirectory(prefix="biblioteca-chrome-") as perfil:
            proc = subprocess.Popen([CHROME, "--headless=new", "--disable-gpu", "--no-pdf-header-footer",
                                     "--allow-file-access-from-files", f"--user-data-dir={perfil}",
                                     f"--print-to-pdf={destino}", fuente.as_uri()],
                                    stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            inicio, tam, estable = time.time(), -1, 0
            while time.time() - inicio < 600:
                time.sleep(1)
                if destino.exists() and destino.stat().st_size > 0:
                    nuevo_tam = destino.stat().st_size
                    estable = estable + 1 if nuevo_tam == tam else 0
                    tam = nuevo_tam
                    if estable >= 2 or proc.poll() is not None:
                        break
                elif proc.poll() is not None:
                    break
            if proc.poll() is None:
                proc.kill()
            proc.wait()
        if destino.exists() and destino.stat().st_size > 0:
            return ""
    return f"FALLÓ {item['archivo']} (2 intentos)"


INDICE = """<!doctype html>
<html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Biblioteca S10</title>
<style>
:root { --bg:#f5f7fa; --card:#fff; --ink:#1c2430; --muted:#5a6676; --line:#dde3ea; --brand:#0f4f99; --chip:#e8f0fa; }
@media (prefers-color-scheme: dark) { :root { --bg:#0f1620; --card:#17212d; --ink:#e3eaf2; --muted:#9aa9ba; --line:#263446;
  --brand:#7db4f0; --chip:#1d2c3d; } }
* { box-sizing: border-box; }
body { margin:0; background:var(--bg); color:var(--ink); font:15px/1.5 -apple-system,"Segoe UI",Helvetica,Arial,sans-serif;
  padding: 0 16px 48px; }
header { max-width:1100px; margin:0 auto; padding:28px 0 16px; }
h1 { margin:0 0 4px; font-size:26px; } .sub { color:var(--muted); margin:0; }
.barra { max-width:1100px; margin:0 auto 18px; display:flex; flex-wrap:wrap; gap:10px; align-items:center; }
input { flex:1 1 280px; padding:10px 12px; border:1px solid var(--line); border-radius:8px; background:var(--card); color:var(--ink); font-size:15px; }
.filtro { display:flex; gap:6px; flex-wrap:wrap; }
.filtro button { border:1px solid var(--line); background:var(--card); color:var(--ink); border-radius:999px; padding:6px 12px; cursor:pointer; }
.filtro button.activo { background:var(--brand); color:#fff; border-color:var(--brand); }
main { max-width:1100px; margin:0 auto; display:grid; gap:18px; }
section.modulo { background:var(--card); border:1px solid var(--line); border-radius:10px; padding:14px 18px; }
section.modulo h2 { margin:0 0 2px; font-size:18px; } .cuenta { color:var(--muted); font-size:13px; }
h3 { font-size:13px; text-transform:uppercase; letter-spacing:.06em; color:var(--muted); margin:14px 0 6px; }
ul { list-style:none; margin:0; padding:0; display:grid; gap:4px; }
li a { color:var(--brand); text-decoration:none; font-weight:500; } li a:hover { text-decoration:underline; }
li .det { color:var(--muted); font-size:13px; margin-left:6px; }
.vacio { color:var(--muted); padding:20px 0; text-align:center; }
footer { max-width:1100px; margin:24px auto 0; color:var(--muted); font-size:13px; }
</style></head><body>
<header><h1>Biblioteca S10</h1>
<p class="sub">Procedimientos paso a paso y manuales oficiales, por módulo. Generado el __FECHA__ · __TOTAL__ documentos.</p></header>
<div class="barra"><input id="q" type="search" placeholder="Buscar: «aprobar pedido», «portal proveedor», «planilla»…" autofocus>
<div class="filtro" id="filtro"><button data-t="" class="activo">Todo</button><button data-t="procedimiento">Procedimientos</button><button data-t="manual">Manuales</button></div></div>
<main id="lista"></main>
<footer>Los procedimientos están verificados contra el manual oficial; los manuales son la documentación de S10 tal como se descargó. Uso interno.</footer>
<script>
const DATOS = __DATOS__;
const NOMBRES = __NOMBRES__;
let tipo = "";
const norm = s => s.normalize("NFD").replace(/[\\u0300-\\u036f]/g, "").toLowerCase();
function pintar() {
  const q = norm(document.getElementById("q").value.trim());
  const porMod = {};
  for (const d of DATOS) {
    if (tipo && d.tipo !== tipo) continue;
    if (q && !q.split(/\\s+/).every(w => norm(d.buscar).includes(w))) continue;
    (porMod[d.modulo] ||= []).push(d);
  }
  const lista = document.getElementById("lista"); lista.innerHTML = "";
  const mods = Object.keys(NOMBRES).filter(m => porMod[m]);
  if (!mods.length) { lista.innerHTML = '<p class="vacio">Nada coincide con la búsqueda.</p>'; return; }
  for (const m of mods) {
    const s = document.createElement("section"); s.className = "modulo";
    const procs = porMod[m].filter(d => d.tipo === "procedimiento"), mans = porMod[m].filter(d => d.tipo === "manual");
    let h = `<h2>${NOMBRES[m]}</h2><div class="cuenta">${procs.length} procedimientos · ${mans.length} manuales</div>`;
    const ul = xs => "<ul>" + xs.map(d => `<li><a href="${encodeURI(d.archivo)}" target="_blank" rel="noopener">${d.titulo}</a><span class="det">${d.detalle}</span></li>`).join("") + "</ul>";
    if (procs.length) h += "<h3>Procedimientos</h3>" + ul(procs);
    if (mans.length) h += "<h3>Manuales oficiales</h3>" + ul(mans);
    s.innerHTML = h; lista.appendChild(s);
  }
}
document.getElementById("q").addEventListener("input", pintar);
document.getElementById("filtro").addEventListener("click", e => {
  if (!e.target.dataset || e.target.tagName !== "BUTTON") return;
  tipo = e.target.dataset.t;
  document.querySelectorAll("#filtro button").forEach(b => b.classList.toggle("activo", b === e.target));
  pintar();
});
pintar();
</script></body></html>
"""


def escribir_indice(items: list[dict]):
    publicos = [{k: v for k, v in x.items() if not k.startswith("_")} for x in items]
    publicos.sort(key=lambda x: (list(NOMBRES).index(x["modulo"]) if x["modulo"] in NOMBRES else 99, x["tipo"] != "procedimiento",
                                 x["titulo"].lower()))
    (SALIDA / "catalogo.json").write_text(json.dumps(publicos, ensure_ascii=False, indent=1), encoding="utf-8")
    datos = json.dumps([{k: x[k] for k in ("tipo", "modulo", "titulo", "archivo", "detalle", "buscar")} for x in publicos],
                       ensure_ascii=False).replace("</", "<\\/")
    (SALIDA / "index.html").write_text(
        INDICE.replace("__DATOS__", datos).replace("__NOMBRES__", json.dumps(NOMBRES, ensure_ascii=False))
        .replace("__FECHA__", datetime.now().strftime("%d-%m-%Y %H:%M")).replace("__TOTAL__", str(len(publicos))),
        encoding="utf-8")


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--solo-indice", action="store_true")
    ap.add_argument("--hilos", type=int, default=4)
    ap.add_argument("--faltantes", action="store_true", help="solo los PDF que aún no existen")
    a = ap.parse_args()
    items = procedimientos() + manuales() + documentos()
    SALIDA.mkdir(exist_ok=True)
    if not a.solo_indice:
        if not Path(CHROME).exists():
            sys.exit(f"falta Google Chrome en {CHROME}")
        CONSTRUCCION.mkdir(parents=True, exist_ok=True)
        pendientes = [x for x in items if not (a.faltantes and (SALIDA / x["archivo"]).exists()
                                                   and (SALIDA / x["archivo"]).stat().st_size > 0)]
        with ThreadPoolExecutor(a.hilos) as ex:
            fallos = [f for f in ex.map(a_pdf, pendientes) if f]
        shutil.rmtree(CONSTRUCCION, ignore_errors=True)
        for f in fallos:
            print(f, file=sys.stderr)
    escribir_indice(items)
    por_mod = {}
    for x in items:
        por_mod.setdefault(x["modulo"], {"procedimiento": 0, "manual": 0, "documento": 0})[x["tipo"]] += 1
    print(json.dumps({"documentos": len(items), "por_modulo": por_mod},
                     ensure_ascii=False, indent=1))


if __name__ == "__main__":
    main()
