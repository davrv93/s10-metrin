#!/usr/bin/env python3
"""
s10kb — recolecta la documentación de https://documentacion.s10peru.com y arma
una base de conocimiento (texto troceado en JSONL) a partir de sus manuales PDF.

Principio: el script ve exactamente lo que ve un miembro con sesión iniciada.
Entra con la cuenta que S10 le dio a Arturo (variables S10_USUARIO / S10_CLAVE),
recorre las páginas del sitemap y baja los PDF enlazados desde esas páginas.
NO descarga archivos por URL directa de wp-content/uploads ni desde el sitemap
de adjuntos: eso saltaría el muro de miembros del sitio.

Pasos (cada uno se puede repetir; retoma donde quedó):

  descubrir   sitemap de páginas y posts        -> data/urls.json
  rastrear    login + cada página -> texto y PDF -> data/paginas/*.md, data/enlaces.jsonl
  descargar   PDF encontrados                   -> data/pdf/*.pdf, data/manifiesto.jsonl
  indexar     pdftotext + troceo                 -> kb/fragmentos.jsonl, kb/documentos.json
  todo        los cuatro en orden
  oficial     todo con la cuenta de miembro + rama manuales-oficiales de Cortex
              + fragmentos de Cortex, y avisa a Metrín (kb/.actualizado)

Uso:
  cp .env.example .env   # y completar usuario/clave
  .venv/bin/python s10kb.py todo
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
import time
from dataclasses import dataclass
from pathlib import Path
from urllib.parse import parse_qs, unquote, urljoin, urlparse

import requests
from bs4 import BeautifulSoup

BASE = "https://documentacion.s10peru.com"
RAIZ = Path(__file__).resolve().parent
DATA = RAIZ / "data"
KB = RAIZ / "kb"
UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X) S10-KB/0.1 (base de conocimiento interna; contacto: consultora de Arturo)"
PAUSA = float(os.environ.get("S10_PAUSA", "1.0"))  # segundos entre peticiones: el sitio es de un cliente, no se le carga
HOSTS_PDF = tuple(h.strip() for h in os.environ.get("S10_HOSTS_PDF", "s10peru.com").split(",") if h.strip())
MURO = "swpm-post-not-logged-in-msg"  # marca que pinta Simple Membership cuando falta sesión


# ─────────────────────────────── utilidades ────────────────────────────────
def cargar_env() -> None:
    f = RAIZ / ".env"
    if not f.exists():
        return
    for linea in f.read_text(encoding="utf-8").splitlines():
        linea = linea.strip()
        if not linea or linea.startswith("#") or "=" not in linea:
            continue
        k, v = linea.split("=", 1)
        os.environ.setdefault(k.strip(), v.strip().strip('"').strip("'"))


def log(*a) -> None:
    print(time.strftime("%H:%M:%S"), *a, flush=True)


def slug(texto: str, largo: int = 80) -> str:
    s = re.sub(r"[^a-zA-Z0-9]+", "-", unquote(texto)).strip("-").lower()
    return s[:largo] or "sin-nombre"


def leer_jsonl(p: Path) -> list[dict]:
    if not p.exists():
        return []
    return [json.loads(l) for l in p.read_text(encoding="utf-8").splitlines() if l.strip()]


def agregar_jsonl(p: Path, fila: dict) -> None:
    p.parent.mkdir(parents=True, exist_ok=True)
    with p.open("a", encoding="utf-8") as f:
        f.write(json.dumps(fila, ensure_ascii=False) + "\n")


def normalizar(url: str) -> str:
    """Mismo recurso con http/https o con/sin barra final cuenta una sola vez."""
    u = urlparse(url)
    ruta = u.path if u.path.endswith("/") or "." in u.path.rsplit("/", 1)[-1] else u.path + "/"
    q = f"?{u.query}" if u.query else ""
    return f"https://{u.netloc.lower()}{ruta}{q}"


# ─────────────────────────────── sesión HTTP ───────────────────────────────
def formulario_login(html: str, url: str, usuario: str, clave: str) -> tuple[str, dict]:
    """
    (action, datos) del formulario de acceso tal como lo pinta el sitio: copia los
    campos ocultos (nonce, redirect…) y pone usuario y clave en los campos que el
    formulario tenga. Simple Membership usa swpm_user_name/swpm_password, otros
    temas custom_username/custom_password; si no se reconoce, los de siempre.
    """
    soup = BeautifulSoup(html, "html.parser")
    for form in soup.find_all("form"):
        clave_campo = form.find("input", attrs={"type": "password"})
        if not clave_campo or not clave_campo.get("name"):
            continue
        datos: dict[str, str] = {}
        usuario_campo = None
        for inp in form.find_all(["input", "button"]):
            nombre, tipo = inp.get("name"), (inp.get("type") or "text").lower()
            if not nombre or tipo in ("checkbox", "radio", "reset", "button", "file", "image"):
                continue
            if tipo in ("text", "email") and usuario_campo is None:
                usuario_campo = nombre
            datos[nombre] = inp.get("value", "")
        if usuario_campo is None:
            continue
        datos[usuario_campo] = usuario
        datos[clave_campo["name"]] = clave
        return urljoin(url, form.get("action") or url), datos
    return url, {"custom_username": usuario, "custom_password": clave, "custom_api_login": "Iniciar sesión"}


class Cliente:
    def __init__(self) -> None:
        self.s = requests.Session()
        self.s.headers["User-Agent"] = UA
        self._ultima = 0.0
        self.con_sesion = False

    def _esperar(self) -> None:
        falta = PAUSA - (time.monotonic() - self._ultima)
        if falta > 0:
            time.sleep(falta)
        self._ultima = time.monotonic()

    def get(self, url: str, **kw) -> requests.Response:
        for intento in range(4):
            self._esperar()
            try:
                r = self.s.get(url, timeout=60, **kw)
                if r.status_code in (429, 502, 503, 504):
                    raise requests.HTTPError(f"{r.status_code}")
                return r
            except (requests.ConnectionError, requests.Timeout, requests.HTTPError) as e:
                espera = 5 * (intento + 1)
                log(f"  reintento {intento + 1} en {espera}s ({e}) {url}")
                time.sleep(espera)
        raise RuntimeError(f"no se pudo obtener {url}")

    def cargar_cookies(self, ruta: str) -> None:
        """Usa la sesión iniciada en el navegador: exporta las cookies con una
        extensión (formato Netscape/curl) y pásala con --cookies. Verifica con
        la misma prueba del login por formulario."""
        from http.cookiejar import MozillaCookieJar
        jar = MozillaCookieJar(ruta)
        try:
            jar.load(ignore_discard=True, ignore_expires=True)
        except Exception as e:
            raise SystemExit(f"no pude leer {ruta} como cookies Netscape: {e}")
        for galleta in jar:
            self.s.cookies.set(galleta.name, galleta.value, domain=galleta.domain or ".s10peru.com")
        prueba = self.get(f"{BASE}/manual-de-presupuestos/")
        if MURO in prueba.text:
            raise SystemExit("las cookies no abrieron los manuales (caducadas o de otra cuenta)")
        self.con_sesion = True
        log("sesión cargada desde cookies del navegador")

    def login(self) -> None:
        usuario, clave = os.environ.get("S10_USUARIO"), os.environ.get("S10_CLAVE")
        if not usuario or not clave:
            raise SystemExit(
                "Faltan S10_USUARIO y S10_CLAVE en .env.\n"
                "Los manuales están tras el login de miembros; sin la cuenta de Arturo solo\n"
                "se puede rastrear lo público (añade --sin-login)."
            )
        url = f"{BASE}/membership-login/"
        pagina = self.get(url)  # cookies iniciales
        destino, datos = formulario_login(pagina.text, url, usuario, clave)
        self._esperar()
        r = self.s.post(destino, data=datos, headers={"Referer": url}, timeout=60)
        # Prueba real: una página de manual deja de mostrar el aviso de "debes acceder".
        prueba = self.get(f"{BASE}/manual-de-presupuestos/")
        if MURO in prueba.text:
            raise SystemExit(
                f"El login no abrió los manuales (HTTP {r.status_code}). Revisa usuario/clave "
                "o entra una vez a mano en el navegador por si la cuenta pide confirmar algo."
            )
        self.con_sesion = True
        log("sesión de miembro iniciada")


# ─────────────────────────────── 1. descubrir ──────────────────────────────
def descubrir(c: Cliente) -> list[str]:
    urls: list[str] = []
    for mapa in ("page-sitemap.xml", "post-sitemap.xml"):
        r = c.get(f"{BASE}/{mapa}")
        encontrados = re.findall(r"<loc>([^<]+)</loc>", r.text)
        log(f"{mapa}: {len(encontrados)} URL")
        urls += encontrados
    unicas = sorted({normalizar(u) for u in urls if urlparse(u).netloc.endswith("s10peru.com")})
    DATA.mkdir(exist_ok=True)
    (DATA / "urls.json").write_text(json.dumps(unicas, indent=1, ensure_ascii=False), encoding="utf-8")
    log(f"{len(unicas)} páginas únicas -> data/urls.json")
    return unicas


# ─────────────────────────────── 2. rastrear ───────────────────────────────
def enlaces_pdf(soup: BeautifulSoup, url_pagina: str) -> list[tuple[str, str]]:
    """PDF enlazados o incrustados en la página: (url, texto del enlace)."""
    halla: dict[str, str] = {}

    def anotar(bruto: str | None, texto: str = "") -> None:
        if not bruto:
            return
        u = urljoin(url_pagina, bruto.strip())
        # visores tipo pdf.js / pdf-embedder: ...viewer.html?file=<pdf>
        q = parse_qs(urlparse(u).query)
        for clave in ("file", "url", "src", "pdf"):
            if clave in q and ".pdf" in q[clave][0].lower():
                u = urljoin(url_pagina, unquote(q[clave][0]))
                break
        if ".pdf" in unquote(urlparse(u).path).lower():
            halla.setdefault(u, texto.strip())

    for a in soup.find_all("a", href=True):
        anotar(a["href"], a.get_text(" ", strip=True))
    for tag, attr in (("iframe", "src"), ("embed", "src"), ("object", "data"), ("source", "src")):
        for t in soup.find_all(tag):
            anotar(t.get(attr))
    for t in soup.find_all(attrs={"data-pdf-url": True}):
        anotar(t["data-pdf-url"])
    for t in soup.find_all(attrs={"data-src": True}):
        anotar(t["data-src"])
    return list(halla.items())


def texto_pagina(soup: BeautifulSoup) -> tuple[str, str]:
    titulo = (soup.find("h1") or soup.find("title"))
    titulo = titulo.get_text(" ", strip=True) if titulo else ""
    cuerpo = soup.find("main") or soup.find(attrs={"data-elementor-type": "wp-page"}) or soup.body or soup
    for basura in cuerpo.find_all(["script", "style", "noscript", "nav", "header", "footer", "form"]):
        basura.decompose()
    lineas = [l.strip() for l in cuerpo.get_text("\n").splitlines()]
    return titulo, "\n".join(l for l in lineas if l)


def rastrear(c: Cliente, urls: list[str], refrescar: bool = False) -> dict:
    """
    Guarda el texto de cada página y anota sus PDF. Con `refrescar` vuelve a bajar
    también las ya vistas y solo reescribe las que cambiaron (huella sha256 del texto).
    Devuelve {'nuevas', 'cambiadas', 'iguales', 'pdf_nuevos', 'errores'}.
    """
    previas: dict[str, dict] = {}
    for f in leer_jsonl(DATA / "paginas.jsonl"):
        previas[f["pagina"]] = f          # la última fila de cada página manda
    # una página vista tras el muro (sin sesión) no cuenta como hecha: se repite con login
    hechos = {u for u, f in previas.items() if not f.get("muro_de_miembros") and f.get("estado") == 200}
    (DATA / "paginas").mkdir(parents=True, exist_ok=True)
    vistos_pdf = {e["url"] for e in leer_jsonl(DATA / "enlaces.jsonl")}
    stats = {"nuevas": 0, "cambiadas": 0, "iguales": 0, "pdf_nuevos": 0, "errores": 0}
    bloqueadas = 0
    for i, url in enumerate(urls, 1):
        if url in hechos and not refrescar:
            continue
        r = c.get(url)
        if r.status_code != 200:
            log(f"[{i}/{len(urls)}] {r.status_code} {url}")
            stats["errores"] += 1
            if url not in previas:
                previas[url] = {"pagina": url, "estado": r.status_code}
            continue
        soup = BeautifulSoup(r.text, "html.parser")
        muro = MURO in r.text
        bloqueadas += muro
        pdfs = enlaces_pdf(soup, url)
        titulo, texto = texto_pagina(soup)
        antes = previas.get(url, {})
        if antes.get("titulo_fijo"):
            titulo = antes["titulo"]      # títulos puestos a mano no se pisan al refrescar
        if muro and not c.con_sesion and antes.get("estado") == 200 and not antes.get("muro_de_miembros", True):
            # ya se bajó con la cuenta de miembro: un rastreo público no la pisa con el aviso del muro
            stats["iguales"] += 1
            continue
        huella = hashlib.sha256(texto.encode()).hexdigest()[:16]
        u = urlparse(url)
        # el portal conserva sus nombres de siempre; otros sitios llevan el dominio delante
        base = u.path if u.netloc.endswith("documentacion.s10peru.com") else f"{u.netloc.replace('www.', '')}{u.path}"
        nombre = slug(base) + ".md"
        if antes.get("huella") == huella and antes.get("estado") == 200:
            stats["iguales"] += 1
            continue
        stats["cambiadas" if antes.get("estado") == 200 else "nuevas"] += 1
        (DATA / "paginas" / nombre).write_text(f"# {titulo}\n\nFuente: {url}\n\n{texto}\n", encoding="utf-8")
        nuevos = 0
        for pdf, etiqueta in pdfs:
            if pdf in vistos_pdf:
                continue
            vistos_pdf.add(pdf)
            nuevos += 1
            agregar_jsonl(DATA / "enlaces.jsonl", {"url": pdf, "etiqueta": etiqueta, "pagina": url, "titulo_pagina": titulo})
        stats["pdf_nuevos"] += nuevos
        fila = {"pagina": url, "estado": 200, "titulo": titulo, "archivo": f"paginas/{nombre}",
                "muro_de_miembros": muro, "con_sesion": c.con_sesion, "pdf_nuevos": nuevos, "huella": huella,
                "visto": time.strftime("%Y-%m-%dT%H:%M:%S")}
        if antes.get("titulo_fijo"):
            fila["titulo_fijo"] = True
        previas[url] = fila
        marca = " [MURO: sin sesión]" if muro else ""
        log(f"[{i}/{len(urls)}] {titulo[:60]!r} pdf+{nuevos}{marca}")
    # se reescribe el archivo con una fila por página (sin duplicados)
    (DATA / "paginas.jsonl").write_text("".join(json.dumps(f, ensure_ascii=False) + "\n" for f in previas.values()), encoding="utf-8")
    if bloqueadas and not c.con_sesion:
        log(f"{bloqueadas} páginas piden sesión de miembro. Completa .env y vuelve a correr sin --sin-login.")
    log(f"páginas: {stats}")
    return stats


# PDF que S10 enlaza desde páginas abiertas de su web (sílabos, manuales públicos…)
def pdfs_publicos(c: Cliente, sitio: str = "https://www.s10peru.com") -> int:
    idx = c.get(f"{sitio}/sitemap_index.xml").text
    mapas = [m for m in re.findall(r"<loc>([^<]+)</loc>", idx) if re.search(r"(page|post)-sitemap", m)]
    paginas = [u for m in mapas for u in re.findall(r"<loc>([^<]+)</loc>", c.get(m).text)]
    vistos = {e["url"] for e in leer_jsonl(DATA / "enlaces.jsonl")}
    nuevos = 0
    for u in paginas:
        r = c.get(u)
        if r.status_code != 200:
            continue
        for pdf in dict.fromkeys(re.findall(r'href="([^"]+\.pdf[^"]*)"', r.text, re.I)):
            if pdf in vistos:
                continue
            vistos.add(pdf)
            nuevos += 1
            nombre = unquote(urlparse(pdf).path.rsplit("/", 1)[-1])[:-4]
            agregar_jsonl(DATA / "enlaces.jsonl", {"url": pdf, "etiqueta": re.sub(r"[-_]+", " ", nombre).strip(),
                                                   "pagina": u, "titulo_pagina": "s10peru.com (público)", "origen": "s10peru-publico"})
            log(f"PDF nuevo: {pdf}")
    log(f"{len(paginas)} páginas revisadas en {sitio}; {nuevos} PDF nuevos en cola")
    return nuevos


# Un sitio WordPress completo por su sitemap (p. ej. optimiza360.pe), sin plantillas por defecto
PLANTILLAS_WP = ("sample-page", "hello-world", "home-2")


def sitio(c: Cliente, dominio: str) -> dict:
    base = f"https://{dominio}"
    urls: list[str] = []
    for mapa in ("wp-sitemap.xml", "sitemap_index.xml", "sitemap.xml"):
        r = c.get(f"{base}/{mapa}")
        if r.status_code != 200 or "<loc>" not in r.text:
            continue
        for loc in re.findall(r"<loc>([^<]+)</loc>", r.text):
            if loc.endswith(".xml"):
                if re.search(r"(post|page)", loc):
                    urls += re.findall(r"<loc>([^<]+)</loc>", c.get(loc).text)
            else:
                urls.append(loc)
        break
    urls = [normalizar(u) for u in dict.fromkeys(urls) if not any(p in u for p in PLANTILLAS_WP)]
    log(f"{dominio}: {len(urls)} páginas en el sitemap")
    return rastrear(c, urls, refrescar=True)


# ─────────────────────────────── 3. descargar ──────────────────────────────
EXT_IMG = (".png", ".jpg", ".jpeg", ".webp", ".gif")


def medios(c: Cliente, limite: int = 0) -> dict:
    """Baja el HTML crudo y las imágenes de cada página vista (data/html + data/imagenes).
    Idempotente: salta lo ya bajado. Las imágenes van a data/imagenes/<slug>/NN.ext
    y se anotan en data/imagenes.jsonl {pagina, archivo, url} para Cortex y el RAG.
    """
    from bs4 import BeautifulSoup
    filas = [f for f in leer_jsonl(DATA / "paginas.jsonl")
             if f.get("estado") == 200 and f.get("archivo") and not f.get("muro_de_miembros")]
    if limite:
        filas = filas[:limite]
    (DATA / "html").mkdir(parents=True, exist_ok=True)
    (DATA / "imagenes").mkdir(parents=True, exist_ok=True)
    prev = {}
    for f in leer_jsonl(DATA / "imagenes.jsonl"):
        prev[f["url"]] = f["archivo"]
    stats = {"html": 0, "imagenes": 0, "saltadas": 0, "errores": 0}
    for i, f in enumerate(filas, 1):
        base = Path(f["archivo"]).stem
        hpath = DATA / "html" / f"{base}.html"
        if not hpath.exists() or MURO in hpath.read_text(encoding="utf-8"):
            try:
                r = c.get(f["pagina"])
                if r.status_code != 200:
                    stats["errores"] += 1
                    continue
                if MURO in r.text:
                    stats["errores"] += 1
                    continue
                hpath.write_text(r.text, encoding="utf-8")
                stats["html"] += 1
            except Exception as e:  # noqa: BLE001
                log(f"[{i}/{len(filas)}] html falló {f['pagina']}: {e}")
                stats["errores"] += 1
                continue
        try:
            soup = BeautifulSoup(hpath.read_text(encoding="utf-8"), "html.parser")
        except Exception:  # noqa: BLE001
            stats["errores"] += 1
            continue
        vistas = 0
        for img in soup.find_all("img", src=True):
            u = urljoin(f["pagina"], img["src"].strip())
            pu = urlparse(unquote(u))
            if not pu.netloc.endswith("s10peru.com"):
                continue
            ext = os.path.splitext(pu.path)[1].lower()
            if ext not in EXT_IMG:
                continue
            if u in prev:
                continue
            try:
                r = c.get(u)
                if r.status_code != 200 or not r.content:
                    continue
                nombre = hashlib.sha256(u.encode()).hexdigest()[:12] + ext
                rel = f"imagenes/{base}/{nombre}"
                (DATA / "imagenes" / base).mkdir(parents=True, exist_ok=True)
                (DATA / rel).write_bytes(r.content)
                prev[u] = rel
                agregar_jsonl(DATA / "imagenes.jsonl",
                              {"pagina": f["pagina"], "archivo": rel, "url": u,
                               "alt": (img.get("alt") or "").strip()[:200]})
                vistas += 1
            except Exception as e:  # noqa: BLE001
                log(f"  imagen falló {u}: {e}")
        stats["imagenes"] += vistas
        if vistas == 0:
            stats["saltadas"] += 1
        if i % 20 == 0:
            log(f"[{i}/{len(filas)}] html+{stats['html']} img+{stats['imagenes']}")
    log(f"medios: {stats}")
    return stats


def ocr_imagenes() -> None:
    """Extrae texto de imágenes del portal y lo conserva junto a su procedencia."""
    import shutil

    motor = shutil.which("tesseract")
    if not motor:
        log("OCR de imágenes omitido: instala tesseract con `brew install tesseract tesseract-lang`")
        return
    disponibles = subprocess.run([motor, "--list-langs"], capture_output=True, text=True).stdout
    idioma = "spa+eng" if "spa" in disponibles and "eng" in disponibles else (
        "spa" if "spa" in disponibles else "eng")
    entrada = leer_jsonl(DATA / "imagenes.jsonl")
    salida = DATA / "imagenes_ocr.jsonl"
    previas = {f.get("archivo"): f for f in leer_jsonl(salida)}
    filas = []
    for i, fila in enumerate(entrada, 1):
        archivo = DATA / fila["archivo"]
        if not archivo.exists():
            continue
        anterior = previas.get(fila["archivo"])
        if anterior and anterior.get("mtime_ns") == archivo.stat().st_mtime_ns:
            filas.append(anterior)
            continue
        r = subprocess.run([motor, str(archivo), "stdout", "-l", idioma, "--psm", "6"],
                           capture_output=True, text=True)
        filas.append({**fila, "texto": r.stdout.strip(), "idioma_ocr": idioma,
                      "mtime_ns": archivo.stat().st_mtime_ns})
        if i % 25 == 0:
            log(f"OCR imágenes {i}/{len(entrada)}")
    salida.write_text("".join(json.dumps(f, ensure_ascii=False) + "\n" for f in filas), encoding="utf-8")
    log(f"OCR imágenes: {len(filas)} documentos -> {salida.relative_to(RAIZ)}")


def ampliar() -> list[str]:
    """Descubre páginas que el sitemap no lista (está desactualizado): extrae
    los enlaces internos del HTML ya bajado y los suma a urls.json. Sin red."""
    from bs4 import BeautifulSoup
    base = {normalizar(u) for u in json.loads((DATA / "urls.json").read_text(encoding="utf-8"))} \
        if (DATA / "urls.json").exists() else set()
    nuevas = set()
    for h in sorted((DATA / "html").glob("*.html")):
        soup = BeautifulSoup(h.read_text(encoding="utf-8"), "html.parser")
        for a in soup.find_all("a", href=True):
            u = urljoin(BASE + "/", a["href"].strip())
            p = urlparse(u)
            if not p.netloc.endswith("s10peru.com"):
                continue
            if re.search(r"\.(pdf|png|jpe?g|webp|gif|css|js|zip|rar)(\?|$)", p.path, re.I):
                continue
            if any(s in p.path for s in ("/wp-admin", "/wp-login", "/feed", "/comments")):
                continue
            u = normalizar(u)
            if u not in base:
                nuevas.add(u)
    todas = sorted(base | nuevas)
    (DATA / "urls.json").write_text(json.dumps(todas, indent=1, ensure_ascii=False), encoding="utf-8")
    log(f"ampliar: {len(nuevas)} nuevas, {len(todas)} totales -> data/urls.json")
    return sorted(nuevas)


def descargar(c: Cliente, lote: int = 200) -> None:
    """Baja en colas: como mucho `lote` PDF por corrida; la siguiente sigue donde quedó."""
    enlaces = leer_jsonl(DATA / "enlaces.jsonl")
    hechos = {m["url"] for m in leer_jsonl(DATA / "manifiesto.jsonl")}
    pendientes = [e for e in enlaces if e["url"] not in hechos]
    log(f"{len(pendientes)} PDF pendientes; esta corrida baja hasta {lote}")
    (DATA / "pdf").mkdir(parents=True, exist_ok=True)
    for i, e in enumerate(pendientes[:lote], 1):
        url = e["url"]
        host = urlparse(url).netloc.lower()
        if not any(host == h or host.endswith("." + h) for h in HOSTS_PDF):
            log(f"[{i}/{lote}] externo, se anota y no se baja: {url}")
            agregar_jsonl(DATA / "manifiesto.jsonl", {**e, "estado": "externo"})
            continue
        r = c.get(url, stream=True)
        tipo = r.headers.get("content-type", "")
        if r.status_code != 200 or ("pdf" not in tipo and "octet-stream" not in tipo):
            log(f"[{i}/{lote}] {r.status_code} {tipo} (no es PDF) {url}")
            agregar_jsonl(DATA / "manifiesto.jsonl", {**e, "estado": f"http-{r.status_code}", "tipo": tipo})
            continue
        base = Path(urlparse(url).path).stem
        nombre = slug(e.get("etiqueta") if base in ("content", "") else base) + ".pdf"
        destino = DATA / "pdf" / nombre
        h = hashlib.sha256()
        with destino.open("wb") as f:
            for bloque in r.iter_content(1 << 16):
                f.write(bloque)
                h.update(bloque)
        tam = destino.stat().st_size
        agregar_jsonl(DATA / "manifiesto.jsonl", {
            **e, "estado": "ok", "archivo": f"pdf/{nombre}", "bytes": tam, "sha256": h.hexdigest(),
            "descargado": time.strftime("%Y-%m-%dT%H:%M:%S"),
        })
        log(f"[{i}/{lote}] {nombre} {tam // 1024} KiB")


# ─────────────────────────────── importar ──────────────────────────────────
def importar() -> None:
    """Registra en el manifiesto los PDF puestos a mano en entrada/ (bajados con tu cuenta)."""
    entrada = RAIZ / "entrada"
    entrada.mkdir(exist_ok=True)
    (DATA / "pdf").mkdir(parents=True, exist_ok=True)
    ya = {m.get("sha256") for m in leer_jsonl(DATA / "manifiesto.jsonl")}
    n = 0
    for f in sorted(entrada.glob("*.pdf")):
        datos = f.read_bytes()
        h = hashlib.sha256(datos).hexdigest()
        if h in ya:
            continue
        nombre = slug(f.stem) + ".pdf"
        (DATA / "pdf" / nombre).write_bytes(datos)
        agregar_jsonl(DATA / "manifiesto.jsonl", {
            "url": f"manual://{f.name}", "etiqueta": f.stem.replace("-", " "), "pagina": "importado a mano",
            "titulo_pagina": "Importado a mano", "origen": "manual", "estado": "ok",
            "archivo": f"pdf/{nombre}", "bytes": len(datos), "sha256": h,
            "descargado": time.strftime("%Y-%m-%dT%H:%M:%S"),
        })
        ya.add(h)
        n += 1
        log(f"importado {f.name}")
    log(f"{n} PDF nuevos desde entrada/. Corre `indexar` para sumarlos a la base.")


# ─────────────────────────────── ocr ───────────────────────────────────────
def texto_pdf(pdf: Path) -> str:
    res = subprocess.run(["pdftotext", "-layout", "-enc", "UTF-8", str(pdf), "-"], capture_output=True, text=True)
    return res.stdout if res.returncode == 0 else ""


def _norm(t: str) -> str:
    return re.sub(r"[^a-z0-9áéíóúñü]+", "", t.lower())


def _ocr_paginas(pdf: Path, motor: Path) -> list[str]:
    import tempfile
    with tempfile.TemporaryDirectory() as tmp:
        log(f"OCR {pdf.name}: rasterizando…")
        subprocess.run(["pdftoppm", "-r", "170", "-png", str(pdf), f"{tmp}/p"], check=True)
        imgs = sorted(Path(tmp).glob("p-*.png"))
        salida: list[str] = []
        for i in range(0, len(imgs), 20):  # de 20 en 20 para ver avance
            if sys.platform == "darwin":
                r = subprocess.run([str(motor), *map(str, imgs[i:i + 20])], capture_output=True, text=True)
                lote = r.stdout.split("\f")
            else:  # Docker/Linux: tesseract en español
                lote = [subprocess.run(["tesseract", str(im), "stdout", "-l", "spa"], capture_output=True, text=True).stdout
                        for im in imgs[i:i + 20]]
            lote += [""] * (len(imgs[i:i + 20]) - len(lote))
            salida += lote[: len(imgs[i:i + 20])]
            log(f"  {min(i + 20, len(imgs))}/{len(imgs)} páginas")
    return salida


def ocr(completo: bool = False) -> None:
    """
    OCR con el motor Vision de macOS -> data/ocr/<nombre>.txt (páginas separadas por \f).
      - por defecto: solo PDF escaneados (sin capa de texto).
      - completo: todos los PDF con imágenes. Por página, si el texto nativo es pobre
        (<40 palabras) se usa el OCR; si no, se conserva el nativo y se le suman las
        líneas de las capturas de pantalla que no estén ya en él (menús, campos, botones).
    """
    motor = RAIZ / "herramientas" / "ocr"
    if sys.platform == "darwin" and not motor.exists():
        subprocess.run(["swiftc", "-O", str(RAIZ / "herramientas" / "ocr.swift"), "-o", str(motor)], check=True)
    (DATA / "ocr").mkdir(parents=True, exist_ok=True)
    for m in leer_jsonl(DATA / "manifiesto.jsonl"):
        if m.get("estado") != "ok":
            continue
        pdf = DATA / m["archivo"]
        if not pdf.exists():
            continue  # bajado en otra máquina: su OCR (si lo hubo) viaja en data/ocr
        destino = DATA / "ocr" / (pdf.stem + ".txt")
        nativo = texto_pdf(pdf)
        escaneado = len(nativo.split()) <= 30
        if destino.exists() and (escaneado or destino.stat().st_mtime > pdf.stat().st_mtime and not completo):
            continue
        if not escaneado and not completo:
            continue
        if not escaneado:
            lista = subprocess.run(["pdfimages", "-list", str(pdf)], capture_output=True, text=True).stdout
            if len(lista.splitlines()) <= 2:
                continue  # sin imágenes: el texto nativo ya es todo
            marca = destino.with_suffix(".completo")
            if marca.exists():
                continue
        paginas_ocr = _ocr_paginas(pdf, motor)
        if escaneado:
            destino.write_text("\f".join(paginas_ocr), encoding="utf-8")
        else:
            paginas_nat = nativo.split("\f")
            unidas, sumadas = [], 0
            for k, nat in enumerate(paginas_nat):
                o = paginas_ocr[k] if k < len(paginas_ocr) else ""
                if len(nat.split()) < 40:
                    unidas.append(o if len(o.split()) > len(nat.split()) else nat)
                    continue
                vistos = _norm(nat)
                extra = [l for l in o.splitlines() if len(_norm(l)) > 3 and _norm(l) not in vistos]
                sumadas += len(extra)
                unidas.append(nat + ("\n\n[Texto en capturas de pantalla]\n" + "\n".join(extra) if extra else ""))
            destino.write_text("\f".join(unidas), encoding="utf-8")
            destino.with_suffix(".completo").write_text("", encoding="utf-8")
            log(f"  +{sumadas} líneas de capturas")
        log(f"OCR listo -> data/ocr/{destino.name}")


# ─────────────────────────────── 4. indexar ────────────────────────────────
@dataclass
class Trozo:
    texto: str
    pagina_pdf: int


def trocear(paginas: list[str], tam: int = 1500, solape: int = 200) -> list[Trozo]:
    """Trozos de ~tam caracteres que respetan párrafos y guardan la página del PDF."""
    trozos: list[Trozo] = []
    buf, pag_ini = "", 1
    for n, pagina in enumerate(paginas, 1):
        for parrafo in re.split(r"\n\s*\n", pagina):
            p = re.sub(r"[ \t]+", " ", parrafo).strip()
            if not p:
                continue
            if not buf:
                pag_ini = n
            if len(buf) + len(p) + 1 > tam and buf:
                trozos.append(Trozo(buf.strip(), pag_ini))
                buf, pag_ini = buf[-solape:] + "\n", n
            buf += p + "\n"
    if buf.strip():
        trozos.append(Trozo(buf.strip(), pag_ini))
    return trozos


CONFIANZA = {
    "s10peru-publico": "oficial",        # PDF que S10 publica en su web
    "academico-abierto": "academico",    # tesis e informes: casos concretos, no el uso estándar
    "manual": "tercero-sin-verificar",   # copias bajadas a mano de Scribd, pdfcoffee…: pueden ser versiones viejas
}


def confianza(m: dict) -> str:
    # lo que decidan los gestores desde el panel manda sobre la regla por origen
    manual = DATA / "confianza_manual.json"
    if manual.exists():
        elegido = json.loads(manual.read_text(encoding="utf-8")).get(m.get("url", ""))
        if elegido:
            return elegido
    if "concytec" in m.get("url", ""):
        return "estado-2013"
    return CONFIANZA.get(m.get("origen", ""), "oficial")  # sin origen = portal de miembros de S10


def indexar() -> None:
    KB.mkdir(exist_ok=True)
    salida = KB / "fragmentos.jsonl"
    # PDF que no están en esta máquina (data/pdf no se versiona): se conservan sus fragmentos previos
    previos: dict[str, list[dict]] = {}
    for f in leer_jsonl(salida):
        previos.setdefault(f["documento"], []).append(f)
    docs_previos = {d["id"]: d for d in json.loads((KB / "documentos.json").read_text(encoding="utf-8"))} \
        if (KB / "documentos.json").exists() else {}
    salida.write_text("", encoding="utf-8")
    documentos = []
    total = 0
    for m in leer_jsonl(DATA / "manifiesto.jsonl"):
        if m.get("estado") != "ok":
            continue
        pdf = DATA / m["archivo"]
        via_ocr = DATA / "ocr" / (pdf.stem + ".txt")
        doc_id = m["sha256"][:12]
        if not pdf.exists() and not via_ocr.exists():
            viejos = previos.get(doc_id, [])
            for f in viejos:
                agregar_jsonl(salida, f)
            if doc_id in docs_previos:
                documentos.append(docs_previos[doc_id])
            total += len(viejos)
            log(f"{pdf.name}: no está en esta máquina, se conservan {len(viejos)} fragmentos")
            continue
        texto = texto_pdf(pdf) if pdf.exists() else ""
        if via_ocr.exists():  # escaneado, o nativo enriquecido con las capturas
            texto = via_ocr.read_text(encoding="utf-8")
        paginas = texto.split("\f")
        trozos = trocear(paginas)
        documentos.append({
            "id": doc_id, "titulo": m.get("etiqueta") or pdf.stem, "manual": m.get("titulo_pagina"),
            "archivo": m["archivo"], "fuente": m["url"], "pagina_web": m["pagina"],
            "paginas": len(paginas), "fragmentos": len(trozos), "sin_texto": not texto.strip(),
            "confianza": confianza(m),
            "ocr": via_ocr.exists() and texto == via_ocr.read_text(encoding="utf-8"),
        })
        for k, t in enumerate(trozos):
            agregar_jsonl(salida, {
                "id": f"{doc_id}-{k:04d}", "documento": doc_id, "manual": m.get("titulo_pagina"),
                "titulo": m.get("etiqueta") or pdf.stem, "pagina": t.pagina_pdf,
                "fuente": m["url"], "confianza": confianza(m), "texto": t.texto,
            })
        total += len(trozos)
        aviso = "  (sin texto: corre el paso `ocr`)" if not texto.strip() else ""
        log(f"{pdf.name}: {len(paginas)} pág, {len(trozos)} fragmentos{aviso}")
    # las páginas web también entran: suelen explicar a qué módulo pertenece cada PDF
    # Con HTML guardado (paso `medios`) la página entra por secciones: cada paso
    # lleva sus capturas y Metrín las pone junto al paso que ilustran.
    from secciones_html import secciones, secciones_texto, texto_seccion
    fotos_por_url = {f["url"]: f["archivo"] for f in leer_jsonl(DATA / "imagenes.jsonl")}
    for p in leer_jsonl(DATA / "paginas.jsonl"):
        if p.get("estado") != 200 or p.get("muro_de_miembros"):
            continue
        confianza_web = "propio" if "optimiza360.pe" in p["pagina"] else "oficial"
        html = DATA / "html" / (Path(p["archivo"]).stem + ".html")
        texto_web = (DATA / p["archivo"]).read_text(encoding="utf-8")
        secs = secciones(html.read_text(encoding="utf-8"), p["pagina"], p.get("titulo") or "", fotos_por_url) \
            if html.exists() else []
        if not secs:
            secs = secciones_texto(texto_web)
        for k, sec in enumerate(secs):
            titulo = p.get("titulo") or ""
            if sec["titulo"] and sec["titulo"] != titulo:
                titulo = f"{titulo} › {sec['titulo']}" if titulo else sec["titulo"]
            agregar_jsonl(salida, {
                "id": f"web-{slug(p['pagina'], 40)}-s{k:03d}", "documento": "web", "manual": p.get("titulo"),
                "titulo": titulo, "seccion": sec["titulo"], "pagina": None, "fuente": p["pagina"],
                "confianza": confianza_web, "texto": texto_seccion(sec),
                "pasos": sec["pasos"] if any(ps["fotos"] for ps in sec["pasos"]) else [],
            })
            total += 1
        if secs:
            continue
        for k, t in enumerate(trocear([texto_web])):
            agregar_jsonl(salida, {
                "id": f"web-{slug(p['pagina'], 40)}-{k:03d}", "documento": "web", "manual": p.get("titulo"),
                "titulo": p.get("titulo"), "pagina": None, "fuente": p["pagina"],
                "confianza": confianza_web, "texto": t.texto,
            })
            total += 1
    # Las imágenes son documentos secundarios de su página/manual de origen.
    titulos_pagina = {p.get("pagina"): p.get("titulo", "Manual S10")
                      for p in leer_jsonl(DATA / "paginas.jsonl")}
    for im in leer_jsonl(DATA / "imagenes_ocr.jsonl"):
        texto = im.get("texto", "").strip()
        if not texto:
            continue
        doc_id = "img-" + hashlib.sha256(im["url"].encode()).hexdigest()[:12]
        titulo = im.get("alt") or Path(im["archivo"]).name
        contenido = f"Imagen del manual S10: {titulo}. Texto reconocido por OCR: {texto}"
        agregar_jsonl(salida, {
            "id": f"{doc_id}-0000", "documento": doc_id, "tipo": "imagen",
            "manual": titulos_pagina.get(im.get("pagina"), "Manual S10"),
            "titulo": titulo, "pagina": None, "fuente": im.get("pagina"),
            "imagen": im["archivo"], "url_imagen": im["url"], "confianza": "oficial",
            "texto": contenido,
        })
        documentos.append({"id": doc_id, "titulo": titulo, "manual": "Imagen del portal S10",
                           "archivo": im["archivo"], "fuente": im["url"],
                           "pagina_web": im.get("pagina"), "paginas": 1, "fragmentos": 1,
                           "sin_texto": False, "confianza": "oficial", "ocr": True})
        total += 1
    (KB / "documentos.json").write_text(json.dumps(documentos, indent=1, ensure_ascii=False), encoding="utf-8")
    log(f"{len(documentos)} documentos indexados, {total} fragmentos -> kb/fragmentos.jsonl")


# ─────────────────────────────── CLI ───────────────────────────────────────
def main() -> None:
    cargar_env()
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("paso", choices=["descubrir", "rastrear", "descargar", "importar", "ocr", "indexar", "todo", "oficial",
                                     "pdfs-publicos", "sitio", "medios", "ampliar"])
    ap.add_argument("dominio", nargs="?", help="para `sitio`: dominio a rastrear, p. ej. optimiza360.pe")
    ap.add_argument("--refrescar", action="store_true", help="rastrear: vuelve a bajar también las páginas ya vistas")
    ap.add_argument("--sin-login", action="store_true", help="solo lo público (útil para probar el rastreo)")
    ap.add_argument("--cookies", default="", help="fichero de cookies Netscape (curl -c) de una sesión iniciada en el navegador: evita el login por formulario")
    ap.add_argument("--limite", type=int, default=0, help="procesa solo las N primeras páginas")
    ap.add_argument("--lote", type=int, default=200, help="máximo de PDF por corrida de `descargar` (cola)")
    ap.add_argument("--completo", action="store_true", help="ocr: también las capturas de pantalla de PDF con texto")
    a = ap.parse_args()

    c = Cliente()
    if a.paso == "oficial":
        if a.sin_login:
            raise SystemExit("`oficial` es justamente lo que pide sesión de miembro: quita --sin-login")
        a.paso, a.completo, a.lote = "todo", True, max(a.lote, 1000)
        oficial = True
    else:
        oficial = False
    necesita_red = a.paso in ("rastrear", "descargar", "todo", "medios")
    if a.paso == "pdfs-publicos":
        pdfs_publicos(c)
        return
    if a.paso == "ampliar":
        ampliar()
        return
    if a.paso == "sitio":
        if not a.dominio:
            raise SystemExit("uso: s10kb.py sitio <dominio>")
        sitio(c, a.dominio)
        return
    if necesita_red and not a.sin_login:
        if a.cookies:
            c.cargar_cookies(a.cookies)
        else:
            c.login()

    if a.paso == "medios":
        medios(c, a.limite)
        return

    urls = json.loads((DATA / "urls.json").read_text()) if (DATA / "urls.json").exists() else []
    if a.paso in ("descubrir", "todo") or (a.paso == "rastrear" and not urls):
        urls = descubrir(c)
    if a.limite:
        urls = urls[: a.limite]
    if a.paso in ("rastrear", "todo"):
        rastrear(c, urls, refrescar=a.refrescar)
    if oficial and a.paso == "todo":
        medios(c)
        ocr_imagenes()
    if a.paso in ("descargar", "todo"):
        descargar(c, a.lote)
    if a.paso == "importar":
        importar()
    if a.paso in ("ocr", "todo"):
        ocr(a.completo)
    if a.paso in ("indexar", "todo"):
        indexar()
    if oficial:
        cerrar_oficial()


def cerrar_oficial() -> None:
    """Tras bajar los manuales: cuántos siguen tras el muro, rama de Cortex, fragmentos y aviso a Metrín."""
    filas = [p for p in leer_jsonl(DATA / "paginas.jsonl") if "documentacion.s10peru.com" in p["pagina"]]
    muro = [p["pagina"] for p in filas if p.get("muro_de_miembros")]
    log(f"portal: {len(filas) - len(muro)} páginas con contenido, {len(muro)} siguen tras el muro")
    for u in muro[:10]:
        log(f"  sin acceso con esta cuenta: {u}")
    for paso in (["oficial_a_cortex.py", "--cargar"], ["cortex_a_kb.py"]):
        r = subprocess.run([sys.executable, str(RAIZ / paso[0]), *paso[1:]], cwd=RAIZ)
        if r.returncode != 0:
            log(f"{paso[0]} terminó con código {r.returncode}; lo demás ya quedó indexado")
    (KB / ".actualizado").write_text(time.strftime("%Y-%m-%dT%H:%M:%S"), encoding="utf-8")
    log("kb/.actualizado tocado: Metrín recarga su índice")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit("\ninterrumpido; se retoma donde quedó al volver a correr")
