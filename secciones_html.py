#!/usr/bin/env python3
"""Parte el HTML de un manual del portal en secciones con sus pasos y capturas.

El texto plano de data/paginas/*.md pierde dónde iba cada imagen. Aquí se recorre
data/html/<slug>.html en el orden del documento: cada encabezado (h1–h4) abre una
sección; los párrafos, ítems de lista y celdas son pasos; cada <img> se pega al
paso que la precede (la captura ilustra lo que se acaba de explicar) o, si abre
la sección, al primer paso que venga.

Resultado por sección: {"titulo", "pasos": [{"texto", "fotos": [ruta, ...]}]}
con rutas relativas a data/ (las de data/imagenes.jsonl), que Metrín sirve en
/fotos/<ruta> y coloca debajo del paso de la respuesta que corresponde.

Uso suelto, para revisar una página: python3 secciones_html.py data/html/<slug>.html
"""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path
from urllib.parse import unquote, urljoin

from bs4 import BeautifulSoup, NavigableString, Tag

MURO = "swpm-post-not-logged-in-msg"
ENCABEZADOS = {"h1", "h2", "h3", "h4"}
BLOQUES = {"p", "li", "td", "th", "dt", "dd", "pre", "blockquote", "h5", "h6"}
FUERA = {"script", "style", "noscript", "nav", "header", "footer", "form", "aside", "button", "svg", "iframe",
         "figcaption"}  # el pie de una captura no es un paso
RAIZ_CONTENIDO = (".entry-content", "[data-elementor-type='wp-page']", "[data-elementor-type='wp-post']",
                  "article", "main", "#content")
MAX_SECCION = 1800   # caracteres de pasos por fragmento; más largo se parte entre pasos


def _limpio(t: str) -> str:
    return re.sub(r"\s+", " ", t).strip()


def _fuentes_img(img: Tag, pagina: str) -> list[str]:
    urls = []
    for attr in ("src", "data-src", "data-lazy-src", "data-orig-file"):
        v = (img.get(attr) or "").strip()
        if v and not v.startswith("data:"):
            urls.append(urljoin(pagina, v))
    return urls


def _raiz(soup: BeautifulSoup) -> Tag:
    for sel in RAIZ_CONTENIDO:
        nodo = soup.select_one(sel)
        if nodo and len(nodo.get_text(strip=True)) > 40:
            return nodo
    return soup.body or soup


def eventos(html: str, pagina: str, fotos_por_url: dict[str, str]):
    """Secuencia ('titulo', texto) | ('texto', texto) | ('foto', ruta) en orden de lectura."""
    soup = BeautifulSoup(html, "html.parser")
    for basura in soup.find_all(FUERA):
        basura.decompose()
    salida: list[tuple[str, str]] = []
    linea: list[str] = []   # texto suelto (spans, strong, <br>) entre bloques

    def soltar():
        t = _limpio(" ".join(linea))
        linea.clear()
        if t:
            salida.append(("texto", t))

    def foto(img: Tag):
        for u in _fuentes_img(img, pagina):
            ruta = fotos_por_url.get(u) or fotos_por_url.get(unquote(u))
            if ruta:
                salida.append(("foto", ruta))
                return

    def recorrer(nodo: Tag):
        for hijo in nodo.children:
            if isinstance(hijo, NavigableString):
                if hijo.strip() and type(hijo) is NavigableString:
                    linea.append(str(hijo))
                continue
            if not isinstance(hijo, Tag):
                continue
            nombre = hijo.name
            if nombre == "br":
                soltar()
            elif nombre == "img":
                soltar()
                foto(hijo)
            elif nombre in ENCABEZADOS:
                soltar()
                t = _limpio(hijo.get_text(" "))
                if t:
                    salida.append(("titulo", t))
                for img in hijo.find_all("img"):
                    foto(img)
            elif nombre in BLOQUES and not hijo.find(list(BLOQUES | ENCABEZADOS | {"ul", "ol", "table", "div"})):
                soltar()
                t = _limpio(hijo.get_text(" "))
                if t:
                    salida.append(("texto", t))
                for img in hijo.find_all("img"):
                    foto(img)
            elif nombre in ("span", "strong", "b", "em", "i", "a", "u", "small", "code", "mark", "sup", "sub"):
                if hijo.find("img"):
                    recorrer(hijo)
                else:
                    linea.append(hijo.get_text(" "))
            else:
                soltar()
                recorrer(hijo)
                soltar()
    recorrer(_raiz(soup))
    soltar()
    return salida


def secciones(html: str, pagina: str, titulo_pagina: str, fotos_por_url: dict[str, str]) -> list[dict]:
    if MURO in html:
        return []
    salida: list[dict] = []
    actual = {"titulo": titulo_pagina, "pasos": []}
    pendientes: list[str] = []   # fotos antes del primer paso de la sección

    def cerrar():
        if pendientes and actual["pasos"]:
            actual["pasos"][-1]["fotos"] += pendientes
        if actual["pasos"]:
            salida.append(dict(actual))

    for tipo, valor in eventos(html, pagina, fotos_por_url):
        if tipo == "titulo":
            if valor == titulo_pagina and not actual["pasos"]:
                continue
            cerrar()
            actual = {"titulo": valor, "pasos": []}
            pendientes = []
        elif tipo == "texto":
            if actual["pasos"] and actual["pasos"][-1]["texto"] == valor:
                continue
            actual["pasos"].append({"texto": valor, "fotos": pendientes})
            pendientes = []
        elif tipo == "foto":
            if actual["pasos"]:
                if valor not in actual["pasos"][-1]["fotos"]:
                    actual["pasos"][-1]["fotos"].append(valor)
            elif valor not in pendientes:
                pendientes.append(valor)
    cerrar()
    return [p for s in salida for p in _partir(s)]


def secciones_texto(texto: str) -> list[dict]:
    """Fallback para páginas cuyo HTML guardado solo contiene el armazón del sitio.

    `data/paginas/*.md` conserva a veces el cuerpo completo cuando el HTML de
    `medios` no lo trae. Se parte por encabezados numerados, manteniendo su
    jerarquía en el título; cada párrafo queda como paso y las capturas se
    asocian a nivel de fuente si no hay posición HTML verificable.
    """
    numerado = re.compile(r"^\s*(\d+(?:\.\d+)+)\.?\s+(.+?)\s*$")
    principal = re.compile(r"^\s*(\d+)\.?\s+(.+?)\s*$")
    imperativos = (
        "haga ", "haga clic", "ingrese ", "registre ", "seleccione ", "elija ",
        "ubique ", "use ", "abra ", "pulse ", "presione ", "marque ",
        "primero ", "segundo ", "luego ", "despues ", "después ", "se debe ",
        "para ", "si ", "sí ", "no ",
    )

    def encabezado(linea: str) -> tuple[str, int] | None:
        m = numerado.match(linea)
        if m:
            numero, titulo = m.group(1), m.group(2).strip()
            if len(titulo) <= 140 and not titulo.endswith((".", ";", "?", "!")):
                return f"{numero} {titulo}", numero.count(".") + 1
            return None
        m = principal.match(linea)
        if not m:
            return None
        numero, titulo = m.group(1), m.group(2).strip()
        if not titulo or len(titulo) > 120 or titulo.endswith((".", ";", "?", "!")):
            return None
        bajo = titulo.lower()
        if titulo.isupper() or (titulo[0].isupper() and not bajo.startswith(imperativos)):
            return f"{numero} {titulo}", 1
        return None

    salida: list[dict] = []
    jerarquia: list[tuple[int, str]] = []
    pasos: list[dict] = []
    iniciado = False

    def cerrar() -> None:
        if pasos and jerarquia:
            salida.append({"titulo": " › ".join(t for _, t in jerarquia), "pasos": list(pasos)})
        pasos.clear()

    for bruto in texto.splitlines():
        linea = _limpio(bruto)
        if not linea or linea.lower() in {"table of contents", "toggle"}:
            continue
        h = encabezado(linea)
        if h:
            cerrar()
            titulo, nivel = h
            jerarquia[:] = [(n, t) for n, t in jerarquia if n < nivel]
            jerarquia.append((nivel, titulo))
            iniciado = True
            continue
        if not iniciado:
            continue
        if pasos and pasos[-1]["texto"] == linea:
            continue
        pasos.append({"texto": linea, "fotos": []})
    cerrar()
    return [p for s in salida for p in _partir(s)]


def _partir(seccion: dict) -> list[dict]:
    """Secciones largas en trozos que no cortan un paso (ni separan su captura)."""
    trozos, buf, largo = [], [], 0
    for paso in seccion["pasos"]:
        if buf and largo + len(paso["texto"]) > MAX_SECCION:
            trozos.append(buf)
            buf, largo = [], 0
        buf.append(paso)
        largo += len(paso["texto"]) + 1
    if buf:
        trozos.append(buf)
    if len(trozos) == 1:
        return [seccion]
    return [{"titulo": f"{seccion['titulo']} ({k})", "pasos": t} for k, t in enumerate(trozos, 1)]


def texto_seccion(s: dict) -> str:
    return f"## {s['titulo']}\n" + "\n".join(p["texto"] for p in s["pasos"])


if __name__ == "__main__":
    ruta = Path(sys.argv[1])
    datos = ruta.resolve().parent.parent
    fotos = {}
    for l in (datos / "imagenes.jsonl").read_text(encoding="utf-8").splitlines() if (datos / "imagenes.jsonl").exists() else []:
        f = json.loads(l)
        fotos[f["url"]] = f["archivo"]
    pagina = sys.argv[2] if len(sys.argv) > 2 else "https://documentacion.s10peru.com/"
    for s in secciones(ruta.read_text(encoding="utf-8"), pagina, "", fotos):
        print(f"\n## {s['titulo']}")
        for p in s["pasos"]:
            print(f"  - {p['texto'][:110]}")
            for f in p["fotos"]:
                print(f"      [foto] {f}")
