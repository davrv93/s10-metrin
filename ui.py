#!/usr/bin/env python3
"""
ui.py — del video tutorial al esquema de pantalla: capturas, OCR y extracción de campos.

Fuente de las capturas: los videos del canal oficial de S10 (Marketing S10). De
cada uno se sacan frames con ffmpeg y se deduplican; no hace falta el login del
portal ni una sesión de invitado.

Pasos:
  capturas <id|url>   baja el video y extrae frames  -> data/ui/capturas/{id}/*.png
  ocr [--video ID]    tesseract (spa) con cajas      -> data/ui/ocr/{id}_{n}.jsonl
  campos [--video ID] cuadros + etiquetas -> campos  -> data/ui/campos/{id}_{n}.json

El OCR clásico no "entiende" la pantalla: da texto y coordenadas. Los campos se
deducen con geometría (cuadros con borde) y vecindad (la etiqueta a la izquierda
o encima). Es una primera pasada revisable; para tipos finos se puede sumar
después un modelo de visión (gemma3:4b vía ollama, ya elegido para eso).
"""
from __future__ import annotations

import argparse
import base64
import json
import os
import re
import subprocess
import sys
import time
from dataclasses import dataclass, field
from pathlib import Path

import cv2
import pytesseract
import requests

sys.path.insert(0, str(Path(__file__).resolve().parent))
from s10kb import DATA, KB, RAIZ, agregar_jsonl, leer_jsonl, log, slug

UI = DATA / "ui"
CAPTURAS = UI / "capturas"
OCR = UI / "ocr"
CAMPOS = UI / "campos"
VISION = UI / "vision"
ESQUEMA = UI / "esquema"
VIDEO = DATA / "yt" / "video"
TESSDATA = DATA / "tessdata"

# servidor de Ollama: ojo, en esta máquina corre en 11435, no en el 11434 por defecto
OLLAMA = os.environ.get("S10_VLM_HOST", "http://127.0.0.1:11435")
MODELO_VLM = os.environ.get("S10_VLM", "gemma3:4b")

PROMPT_VLM = """Analiza esta captura de pantalla del ERP S10 y extrae los campos de entrada de datos.

Reglas:
- Incluye SOLO campos de formulario (entradas de datos). NO las columnas de tablas ni grillas, ni botones, ni etiquetas del menú.
- Máximo 20 campos. Si hay más, quédate con los del formulario visible.
- Si un campo está vacío, usa "valor": "".
- Si no hay campos de formulario, devuelve "campos": [].
- No inventes campos que no se vean en la imagen."""

# esquema nativo de Ollama: obliga a una salida JSON válida y corta
ESQUEMA_VLM = {
    "type": "object",
    "properties": {
        "pantalla": {"type": "string"},
        "modulo": {"type": "string"},
        "campos": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "etiqueta": {"type": "string"},
                    "tipo": {"type": "string",
                             "enum": ["texto", "numero", "fecha", "codigo", "lista", "booleano", "texto-largo"]},
                    "valor": {"type": "string"},
                    "seccion": {"type": "string", "enum": ["cabecera", "detalle", "filtros", "otro"]},
                },
                "required": ["etiqueta", "tipo", "valor", "seccion"],
            },
        },
    },
    "required": ["pantalla", "campos"],
}

if TESSDATA.exists():
    os.environ.setdefault("TESSDATA_PREFIX", str(TESSDATA))

# ─────────────────────────────── 1. capturas ───────────────────────────────
def capturas(fuente: str) -> None:
    """Baja el video (si falta) y extrae un frame por segundo, sin repetidos."""
    vid = fuente.rsplit("/", 1)[-1]
    if "watch?v=" in fuente or "youtu.be/" in fuente:
        vid = re.search(r"(?:v=|youtu\.be/)([\w-]{11})", fuente).group(1)
    VIDEO.mkdir(parents=True, exist_ok=True)
    mp4 = VIDEO / f"{vid}.mp4"
    if not mp4.exists():
        log(f"bajando video {vid}…")
        subprocess.run(
            [str(RAIZ / ".venv" / "bin" / "yt-dlp"), "-f",
             "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720]",
             "--merge-output-format", "mp4", "-o", str(VIDEO / "%(id)s.%(ext)s"),
             f"https://youtu.be/{vid}"],
            capture_output=True, text=True, timeout=1200,
        )
    destino = CAPTURAS / vid
    destino.mkdir(parents=True, exist_ok=True)
    if any(destino.glob("*.png")):
        log(f"{vid}: capturas ya extraídas ({len(list(destino.glob('*.png')))})")
        return
    subprocess.run(
        ["ffmpeg", "-v", "error", "-i", str(mp4),
         "-vf", "fps=1,mpdecimate=hi=64*12:lo=64*5:frac=0.33",
         "-fps_mode", "vfr", str(destino / "%04d.png")],
        capture_output=True, text=True, timeout=1200,
    )
    log(f"{vid}: {len(list(destino.glob('*.png')))} capturas únicas -> {destino.relative_to(RAIZ)}")


# ─────────────────────────────── 2. OCR ────────────────────────────────────
@dataclass
class Palabra:
    texto: str
    x: int
    y: int
    w: int
    h: int
    conf: float
    # jerarquía de tesseract: bloque, párrafo y línea. Agrupar por línea evita
    # mezclar columnas distintas que caen en la misma banda horizontal.
    blk: int = 0
    par: int = 0
    lin: int = 0

    @property
    def x2(self) -> int:
        return self.x + self.w

    @property
    def cy(self) -> float:
        return self.y + self.h / 2


def ocr_imagen(png: Path) -> tuple[list[Palabra], tuple[int, int]]:
    img = cv2.imread(str(png))
    if img is None:
        raise RuntimeError(f"no se pudo leer {png}")
    alto, ancho = img.shape[:2]
    d = pytesseract.image_to_data(img, lang="spa", config="--psm 3",
                                  output_type=pytesseract.Output.DICT)
    palabras: list[Palabra] = []
    for i, texto in enumerate(d["text"]):
        t = (texto or "").strip()
        conf = float(d["conf"][i])
        if not t or conf < 40:
            continue
        palabras.append(Palabra(t, int(d["left"][i]), int(d["top"][i]),
                                int(d["width"][i]), int(d["height"][i]), conf,
                                blk=int(d["block_num"][i]), par=int(d["par_num"][i]),
                                lin=int(d["line_num"][i])))
    return palabras, (ancho, alto)


def paso_ocr(video: str | None) -> None:
    OCR.mkdir(parents=True, exist_ok=True)
    for carpeta in sorted(CAPTURAS.iterdir()) if CAPTURAS.exists() else []:
        if not carpeta.is_dir() or (video and carpeta.name != video):
            continue
        for png in sorted(carpeta.glob("*.png")):
            salida = OCR / f"{carpeta.name}_{png.stem}.jsonl"
            if salida.exists():
                continue
            palabras, (ancho, alto) = ocr_imagen(png)
            salida.write_text("\n".join(
                json.dumps({"t": p.texto, "x": p.x, "y": p.y, "w": p.w, "h": p.h,
                            "conf": round(p.conf, 1), "blk": p.blk, "par": p.par, "lin": p.lin},
                           ensure_ascii=False) for p in palabras
            ) + "\n", encoding="utf-8")
            log(f"{png.stem}: {len(palabras)} palabras ({ancho}x{alto})")


# ─────────────────────────────── 3. campos ─────────────────────────────────
@dataclass
class Cuadro:
    x: int
    y: int
    w: int
    h: int
    dentro: list[str] = field(default_factory=list)

    @property
    def x2(self) -> int:
        return self.x + self.w

    @property
    def y2(self) -> int:
        return self.y + self.h


def detectar_cuadros(png: Path) -> tuple[list[Cuadro], tuple[int, int]]:
    """Recuadros de entrada: bordes finos y bajos, como los del ERP (no paneles)."""
    img = cv2.imread(str(png))
    alto, ancho = img.shape[:2]
    gris = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    _, binv = cv2.threshold(gris, 190, 255, cv2.THRESH_BINARY_INV)
    horiz = cv2.morphologyEx(binv, cv2.MORPH_OPEN, cv2.getStructuringElement(cv2.MORPH_RECT, (30, 1)))
    vert = cv2.morphologyEx(binv, cv2.MORPH_OPEN, cv2.getStructuringElement(cv2.MORPH_RECT, (1, 12)))
    lineas = cv2.bitwise_or(horiz, vert)
    contornos, _ = cv2.findContours(lineas, cv2.RETR_LIST, cv2.CHAIN_APPROX_SIMPLE)
    crudos: list[Cuadro] = []
    for c in contornos:
        x, y, w, h = cv2.boundingRect(c)
        if not (35 <= w <= 700 and 14 <= h <= 45):
            continue  # descarta paneles, barras y bloques altos
        if y > alto - 45 or x < 3:
            continue  # descarta barra de estado y el borde del panel lateral
        crudos.append(Cuadro(x, y, w, h))
    crudos.sort(key=lambda c: c.w * c.h, reverse=True)
    limpios: list[Cuadro] = []
    for c in crudos:
        if any(abs(o.x - c.x) <= 3 and abs(o.y - c.y) <= 3 and abs(o.w - c.w) <= 4
               and abs(o.h - c.h) <= 4 for o in limpios):
            continue  # mismo recuadro detectado dos veces por el contorno interno/externo
        limpios.append(c)
    return sorted(limpios, key=lambda c: (c.y, c.x)), (ancho, alto)


def solapan_x(a: Palabra | Cuadro, b: Palabra | Cuadro, minimo: float = 0.55) -> bool:
    izq, der = max(a.x, b.x), min(a.x + a.w, b.x + b.w)
    return der > izq and (der - izq) >= minimo * min(a.w, b.w)


def campos_de_captura(png: Path, palabras: list[Palabra]) -> dict:
    cuadros, _ = detectar_cuadros(png)
    usadas: set[int] = set()
    for c in cuadros:
        for i, p in enumerate(palabras):
            cx, cy = p.x + p.w / 2, p.cy
            if c.x <= cx <= c.x2 and c.y <= cy <= c.y2:
                c.dentro.append(p.texto)
                usadas.add(i)

    # filas de grilla: 3+ recuadros con la misma altura y alineados horizontalmente
    en_grilla: set[int] = set()
    por_linea: dict[int, list[int]] = {}
    for i, c in enumerate(cuadros):
        por_linea.setdefault(round(c.y / 5), []).append(i)
    for indices in por_linea.values():
        if len(indices) >= 3:
            xs = sorted(cuadros[i].x for i in indices)
            if xs[-1] - xs[0] > 300:
                en_grilla.update(indices)

    salida = []
    for i, c in enumerate(cuadros):
        if i in en_grilla:
            continue
        # etiqueta: texto por ENCIMA (misma columna) o a la IZQUIERDA (misma línea)
        arriba, izquierda = [], []
        for j, p in enumerate(palabras):
            if j in usadas:
                continue
            if solapan_x(p, c) and 0 <= c.y - (p.y + p.h) <= 26:
                arriba.append((p.y, p.x, p.texto))
            elif abs(p.cy - (c.y + c.h / 2)) <= 9 and 0 < c.x - p.x2 <= 240:
                izquierda.append((p.y, p.x, p.texto))
        arriba.sort()
        izquierda.sort()
        etiqueta = " ".join(t for _, _, t in (arriba or izquierda)).strip()
        valor = " ".join(c.dentro).strip()
        if not etiqueta and not valor:
            continue
        salida.append({"etiqueta": etiqueta, "valor": valor,
                       "caja": [c.x, c.y, c.w, c.h], "tipo": tipo_probable(etiqueta, valor)})
    return {
        "imagen": png.name, "video": png.parent.name,
        "campos": salida, "cuadros_detectados": len(cuadros),
        "celdas_de_grilla": len(en_grilla),
    }


def tipo_probable(etiqueta: str, valor: str) -> str:
    e = (etiqueta or "").lower()
    if re.fullmatch(r"\s*\d{1,2}[/-]\d{1,2}[/-]\d{2,4}\s*", valor or ""):
        return "fecha"
    if re.search(r"fecha|periodo|desde|hasta", e):
        return "fecha"
    if re.search(r"observacion|descripcion|comentario|glosa", e):
        return "texto-largo"
    if re.search(r"importe|monto|valor|total|saldo|sueldo|porcentaje|%|cantidad|jornal|remuneracion", e):
        return "numero"
    if re.search(r"c[oó]digo|ruc|dni|documento|n[uú]mero|nro", e):
        return "codigo"
    if not valor:
        return "vacio"
    if re.fullmatch(r"[\d.,]+", valor):
        return "numero"
    return "texto"


def paso_campos(video: str | None) -> None:
    CAMPOS.mkdir(parents=True, exist_ok=True)
    for jl in sorted(OCR.glob("*.jsonl")):
        if video and not jl.stem.startswith(video + "_"):
            continue
        vid, marco = jl.stem.split("_", 1)[0], jl.stem.split("_", 1)[1]
        png = CAPTURAS / vid / f"{marco}.png"
        if not png.exists():
            continue
        salida = CAMPOS / f"{jl.stem}.json"
        if salida.exists():
            continue
        filas = leer_jsonl(jl)
        palabras = [Palabra(f["t"], f["x"], f["y"], f["w"], f["h"], f["conf"]) for f in filas]
        info = campos_de_captura(png, palabras)
        salida.write_text(json.dumps(info, ensure_ascii=False, indent=1), encoding="utf-8")
        log(f"{jl.stem}: {len(info['campos'])} campos de {info['cuadros_detectados']} cuadros")


# ─────────────────────── 4. visión (VLM local) ─────────────────────────────
def vision_captura(png: Path, timeout: int = 300) -> dict:
    """Pide al VLM local la lista de campos de formulario de una captura."""
    b64 = base64.b64encode(png.read_bytes()).decode()
    r = requests.post(
        f"{OLLAMA}/api/generate",
        json={"model": MODELO_VLM, "prompt": PROMPT_VLM, "images": [b64],
              "stream": False, "format": ESQUEMA_VLM,
              "options": {"temperature": 0, "num_predict": 3000}},
        timeout=timeout,
    )
    r.raise_for_status()
    texto = r.json().get("response", "")
    try:
        return json.loads(texto)
    except json.JSONDecodeError:
        m = re.search(r"\{.*\}", texto, re.S)
        if not m:
            raise
        return json.loads(m.group(0))


def paso_vision(video: str | None, limite: int) -> None:
    VISION.mkdir(parents=True, exist_ok=True)
    for carpeta in sorted(CAPTURAS.iterdir()) if CAPTURAS.exists() else []:
        if not carpeta.is_dir() or (video and carpeta.name != video):
            continue
        pngs = sorted(carpeta.glob("*.png"))
        if limite:
            pngs = pngs[:limite]
        for png in pngs:
            salida = VISION / f"{carpeta.name}_{png.stem}.json"
            if salida.exists():
                continue
            t0 = time.monotonic()
            try:
                datos = vision_captura(png)
            except Exception as e:
                log(f"{png.stem}: error {e}")
                continue
            datos["imagen"] = png.name
            datos["video"] = carpeta.name
            datos["modelo"] = MODELO_VLM
            salida.write_text(json.dumps(datos, ensure_ascii=False, indent=1), encoding="utf-8")
            log(f"{png.stem}: {len(datos.get('campos', []))} campos en {time.monotonic() - t0:.0f}s — {(datos.get('pantalla') or '')[:40]}")


# ────────────── 4. fusión OCR + VLM → esquema validable ───────────────────
def lineas_ocr(filas: list[dict]) -> list[dict]:
    """Agrupa las palabras del OCR en líneas.

    Si el OCR guardó la jerarquía de tesseract (blk/par/lin) se usa esa: es la
    segmentación real y no junta columnas distintas. Si no, se cae a bandas de Y.
    """
    grupos: dict = {}
    tiene_jerarquia = bool(filas) and "lin" in filas[0]
    for w in filas:
        clave = (w["blk"], w["par"], w["lin"]) if tiene_jerarquia else round(w["y"] / 6)
        grupos.setdefault(clave, []).append(w)
    lineas = []
    for k in sorted(grupos):
        ws = sorted(grupos[k], key=lambda w: w["x"])
        texto = " ".join(w["t"] for w in ws).strip()
        if len(texto) < 2:
            continue
        x = min(w["x"] for w in ws)
        y = min(w["y"] for w in ws)
        lineas.append({
            "texto": texto, "x": x, "y": y,
            "w": max(w["x"] + w["w"] for w in ws) - x,
            "h": max(w["y"] + w["h"] for w in ws) - y,
            "palabras": ws,
        })
    return lineas


def candidatos_ocr(lineas: list[dict], max_n: int = 6) -> list[dict]:
    """Frases candidatas: n-gramas consecutivos de palabras del OCR, con su caja.

    Buscar sobre la línea completa falla (las líneas del ERP juntan varias
    columnas); buscar por palabra suelta da coincidencias espurias. El n-grama
    es el punto medio.
    """
    salida = []
    for l in lineas:
        ws = l["palabras"]
        for n in range(1, min(max_n, len(ws)) + 1):
            for i in range(len(ws) - n + 1):
                grupo = ws[i:i + n]
                texto = " ".join(w["t"] for w in grupo)
                if len(texto.strip()) < 3:
                    continue
                x = min(w["x"] for w in grupo)
                y = min(w["y"] for w in grupo)
                salida.append({
                    "texto": texto, "n": n, "plano": plano(texto),
                    "caja": [x, y, max(w["x"] + w["w"] for w in grupo) - x,
                             max(w["y"] + w["h"] for w in grupo) - y],
                })
    return salida


def plano(texto: str) -> str:
    """minúsculas sin acentos ni signos, para comparar (UIT == uit)."""
    import unicodedata
    t = unicodedata.normalize("NFKD", texto or "").encode("ascii", "ignore").decode()
    return re.sub(r"[^a-z0-9 ]", " ", t.lower()).strip()


def cotejar(etiqueta: str, candidatos: list[dict], umbral: float = 0.7) -> dict | None:
    """Busca la frase de OCR más parecida a lo que dijo el VLM.

    El VLM entiende la pantalla pero confunde letras (UIT→UT, Fijos→Fluidos);
    el OCR da la cadena exacta. Se devuelve el texto del OCR y su caja.

    Reglas para no aceptar basura: la contención solo puntúa con frases largas
    ("ea" dentro de "empleado" no vale) y se premia el n-grama más específico.
    """
    from difflib import SequenceMatcher
    v = plano(etiqueta)
    if len(v) < 3:
        return None
    mejor, puntaje, mejor_n = None, 0.0, 0
    for c in candidatos:
        p = c["plano"]
        if not p:
            continue
        s = SequenceMatcher(None, v, p).ratio()
        if p == v:
            s = 1.0
        # la contención solo vale entre frases de tamaño comparable: aceptar
        # "VITAL" dentro de "remuneracion minima vital" da una etiqueta falsa
        elif len(v) >= 5 and len(p) >= 0.6 * len(v) and (p in v or v in p):
            s = max(s, 0.88)
        # a igualdad de puntaje, la frase más larga (más específica)
        if s > puntaje + 0.02 or (abs(s - puntaje) <= 0.02 and c["n"] > mejor_n):
            mejor, puntaje, mejor_n = c, s, c["n"]
    if mejor and puntaje >= umbral:
        return {"texto": mejor["texto"], "puntaje": round(puntaje, 2), "caja": mejor["caja"]}
    return None


RUIDO_CAMPO = re.compile(r"^(today|hoy|\d{1,2})$", re.I)


def esquema_de_captura(ocr_jsonl: Path, vision_json: Path) -> dict:
    lineas = lineas_ocr(leer_jsonl(ocr_jsonl))
    candidatos = candidatos_ocr(lineas)
    vlm = json.loads(vision_json.read_text(encoding="utf-8"))
    vid, marco = vision_json.stem.split("_", 1)
    campos, vistos = [], set()
    for c in vlm.get("campos", []):
        etiqueta = (c.get("etiqueta") or "").strip()
        valor = (c.get("valor") or "").strip()
        if not etiqueta or RUIDO_CAMPO.match(etiqueta) or RUIDO_CAMPO.match(valor):
            continue  # residuos del date-picker y numeritos de calendario
        clave = plano(etiqueta)
        if clave in vistos:
            continue
        vistos.add(clave)
        corre = cotejar(etiqueta, candidatos)
        etiqueta_ocr = corre["texto"] if corre else None
        # El OCR manda solo si coincide palabra por palabra (subconjunto o
        # superconjunto). Si no, su lectura cruzó celdas y se queda el VLM; la
        # caja del OCR se guarda igual, que es lo que sirve para posicionar.
        usar_ocr = False
        if corre:
            pw = set(plano(etiqueta_ocr).split())
            vw = set(plano(etiqueta).split())
            usar_ocr = corre["puntaje"] >= 0.95 or pw <= vw or vw <= pw
        if etiqueta_ocr:
            etiqueta_ocr = re.sub(r"^[^\w\u00c0-\u024f]+|[^\w\u00c0-\u024f]+$", "", etiqueta_ocr).strip()
        campos.append({
            "etiqueta": etiqueta_ocr if (corre and usar_ocr and etiqueta_ocr) else etiqueta,
            "etiqueta_vlm": etiqueta,
            "etiqueta_ocr": etiqueta_ocr or None,
            "tipo": c.get("tipo", "texto"),
            "valor": valor,
            "seccion": c.get("seccion", "otro"),
            "caja_etiqueta": corre["caja"] if corre else None,
            "coincidencia_ocr": corre["puntaje"] if corre else None,
        })
    titulo = cotejar(vlm.get("pantalla") or "", candidatos, 0.6)
    return {
        "pantalla": titulo["texto"] if titulo else (vlm.get("pantalla") or ""),
        "pantalla_vlm": vlm.get("pantalla") or "",
        "modulo": vlm.get("modulo") or "",
        "video": vid, "captura": f"{marco}.png", "modelo": vlm.get("modelo"),
        "campos": campos,
    }


def paso_esquema(video: str | None, frame: str | None) -> None:
    ESQUEMA.mkdir(parents=True, exist_ok=True)
    for vj in sorted(VISION.glob("*.json")):
        if video and not vj.stem.startswith(video + "_"):
            continue
        if frame and not vj.stem.endswith("_" + frame.zfill(4)):
            continue
        oj = OCR / f"{vj.stem}.jsonl"
        if not oj.exists():
            continue
        salida = ESQUEMA / f"{vj.stem}.json"
        if salida.exists():
            continue
        esq = esquema_de_captura(oj, vj)
        salida.write_text(json.dumps(esq, ensure_ascii=False, indent=1), encoding="utf-8")
        log(f"{vj.stem}: {len(esq['campos'])} campos — pantalla {esq['pantalla'][:40]!r}")


# ───────── 4ter. agrupar capturas en pantallas distintas ───────────────────
def paso_pantallas(video: str | None) -> None:
    """Agrupa las capturas en pantallas: huella visual (dhash) o campos en comـún.

    El título que devuelve el VLM no sirve para agrupar (repite "S10 Nóminas"
    en casi todas), así que el agrupador se apoya en la imagen.
    """
    from PIL import Image
    archivos = sorted(ESQUEMA.glob("*.json"))
    if video:
        archivos = [p for p in archivos if p.stem.startswith(video + "_")]
    if not archivos:
        raise SystemExit("no hay esquemas; corre antes: ui.py esquema")
    datos = [(p, json.loads(p.read_text(encoding="utf-8"))) for p in archivos]

    def dhash(ruta: Path, lado: int = 8) -> list[int]:
        im = Image.open(ruta).convert("L").resize((lado + 1, lado))
        px = im.tobytes()
        return [int(px[r * (lado + 1) + c] < px[r * (lado + 1) + c + 1])
                for r in range(lado) for c in range(lado)]

    grupos: list[dict] = []
    for p, d in datos:
        png = CAPTURAS / d.get("video", "") / d.get("captura", "")
        h = dhash(png) if png.exists() else None
        firma = {plano(c["etiqueta"]) for c in d["campos"] if c.get("etiqueta")}
        for g in grupos:
            visual = h is not None and g["h"] is not None and \
                sum(x != y for x, y in zip(h, g["h"])) <= 8
            textual = len(firma & g["firma"]) / max(1, len(firma | g["firma"])) >= 0.5
            if visual or textual:
                g["capturas"].append(d["captura"])
                g["firma"] |= firma
                if not g["titulo"]:
                    g["titulo"] = d["pantalla"]
                break
        else:
            grupos.append({"h": h, "firma": firma, "capturas": [d["captura"]],
                           "ejemplo": p.stem, "titulo": d["pantalla"]})
    log(f"{len(grupos)} pantallas distintas en {len(datos)} capturas")
    for g in sorted(grupos, key=lambda g: -len(g["capturas"])):
        log(f"  {g['ejemplo']}: {(g['titulo'] or '?')[:34]!r} — {len(g['capturas'])} capturas, "
            f"{len(g['firma'])} etiquetas")
    salida = UI / "pantallas.json"
    salida.write_text(json.dumps(
        [{**{k: v for k, v in g.items() if k not in ("h", "firma")},
          "firma": sorted(g["firma"])} for g in grupos],
        ensure_ascii=False, indent=1), encoding="utf-8")
    log(f"-> {salida.relative_to(RAIZ)}")


# ─────────────── 4quinquies. pantallas → KB y Cortex ───────────────────────
TIPO_DETALLE = {
    "texto": "texto", "texto-largo": "texto largo", "numero": "número",
    "fecha": "fecha", "codigo": "código", "lista": "lista de opciones", "booleano": "casilla",
}


def campos_legibles(esq: dict) -> list[str]:
    """Una línea por campo, para el corpus del RAG y los nodos de Cortex."""
    out = []
    for c in esq.get("campos", []):
        partes = [f"**{c['etiqueta']}** ({TIPO_DETALLE.get(c['tipo'], c['tipo'])})"]
        if c.get("valor"):
            partes.append(f"con valor «{c['valor']}»")
        if c.get("seccion") and c["seccion"] != "otro":
            partes.append(f"[{c['seccion']}]")
        if c.get("coincidencia_ocr"):
            partes.append(f"(confirmado por OCR {c['coincidencia_ocr']})")
        out.append("- " + " ".join(partes))
    return out


def paso_kb() -> None:
    """Convierte los esquemas en fragmentos para Metrín/tutor y un corpus para Cortex.

    Escribe kb/fragmentos_pantallas.jsonl (glob kb/fragmentos*.jsonl, así que
    Metrín y tutor.py lo levantan solos) y data/ui/corpus_pantallas.jsonl con
    un documento por pantalla para los borradores del árbol.
    """
    KB.mkdir(exist_ok=True)
    salida = KB / "fragmentos_pantallas.jsonl"
    salida.write_text("", encoding="utf-8")
    corpus = []
    total = 0
    for p in sorted(ESQUEMA.glob("*.json")):
        esq = json.loads(p.read_text(encoding="utf-8"))
        campos = campos_legibles(esq)
        if not campos:
            continue
        vid = esq.get("video", "")
        marco = esq.get("captura", "").removesuffix(".png").lstrip("0") or "0"
        url = f"https://youtu.be/{vid}?t={int(marco)}" if vid else ""
        titulo = esq.get("pantalla") or f"captura {esq.get('captura')}"
        doc_id = p.stem
        texto = (f"Pantalla del ERP S10: «{titulo}» (módulo Nóminas, ventana de Constantes "
                 f"por Fecha General). Campos de formulario detectados en la captura "
                 f"{esq.get('captura')} del video {vid}:\n" + "\n".join(campos))
        trozos = trocear_texto(texto)
        for k, t in enumerate(trozos):
            agregar_jsonl(salida, {
                "id": f"{doc_id}-{k:03d}", "documento": doc_id, "tipo": "pantalla",
                "titulo": titulo, "pantalla": titulo, "modulo": esq.get("modulo") or "Nóminas",
                "video": vid, "url": url, "captura": esq.get("captura"),
                "campos": len(esq.get("campos", [])), "texto": t,
            })
            total += 1
        corpus.append({"id": doc_id, "titulo": titulo, "texto": texto, "url": url,
                       "campos": len(esq.get("campos", [])),
                       "esquema": esq.get("campos", [])})
    (UI / "corpus_pantallas.jsonl").write_text(
        "\n".join(json.dumps(c, ensure_ascii=False) for c in corpus) + "\n", encoding="utf-8")
    log(f"{len(corpus)} pantallas, {total} fragmentos -> {salida.relative_to(RAIZ)}")


def trocear_texto(texto: str, tam: int = 1400, solape: int = 150) -> list[str]:
    if len(texto) <= tam:
        return [texto]
    trozos, buf = [], ""
    for parrafo in texto.split("\n"):
        if buf and len(buf) + len(parrafo) + 1 > tam:
            trozos.append(buf.strip())
            buf = buf[-solape:] + "\n"
        buf += parrafo + "\n"
    if buf.strip():
        trozos.append(buf.strip())
    return trozos


# Título canónico por video: el OCR/VLM lee mal los títulos de ventana
# ("S10"→"510"/"$10", la barra de estado como pantalla). Lo canónico lo fija
# la revisión humana; el resto de nodos generados son duplicados que el dueño
# borra en el tablero (la API lo reserva a humanos).
CANONICAS = {
    "17W0yKI8iew": "Constantes por Fecha General (S10 Nóminas)",
}


def paso_cortex() -> None:
    """Genera el borrador data/cortex-borradores/pantallas-ui.json (el humano lo aprueba en el tablero).

    Un nodo por video (la pantalla dominante), no por captura: generar por
    captura creaba duplicados ("510 Nóminas", "$10 Nóminas", "S10 Nomina" son
    lecturas distintas de la misma ventana) y basura de la barra de estado.
    """
    corpus_f = UI / "corpus_pantallas.jsonl"
    if not corpus_f.exists():
        raise SystemExit("corre antes: ui.py kb")
    corpus = leer_jsonl(corpus_f)
    por_video: dict[str, list[dict]] = {}
    for c in corpus:
        vid = c.get("url", "").split("/")[-1].split("?")[0] or c["id"].split("_")[0]
        por_video.setdefault(vid, []).append(c)

    nodos = [{
        "path": "ui-pantallas",
        "title": "Pantallas del ERP documentadas por visión",
        "summary": "Esquemas de campos extraídos de los tutoriales oficiales (capturas + OCR tesseract + gemma3:4b, fusionados). Un nodo por video/pantalla dominante; el esquema es borrador hasta que S10 lo valida en el HTML de ui.py render.",
        "body": (
            "Método: los videos del canal oficial muestran las pantallas reales del ERP. "
            "Por captura se corre OCR (tesseract, cadena exacta y caja) y un VLM local "
            "(gemma3:4b vía Ollama, semántica: qué campo es y de qué tipo), y se fusionan: "
            "la etiqueta final es la del OCR solo si coincide palabra por palabra, el valor "
            "y el tipo vienen del VLM. El resultado es un borrador revisable en el HTML "
            "que genera `ui.py render` (arrastrar y exportar).\n\n**Lecciones del piloto "
            "(2026-09-30, video 17W0yKI8iew):** 60 capturas = 1 pantalla en dos estados "
            "(grilla y detalle); los títulos de ventana que lee la máquina NO sirven "
            "(S10→510/$10, la barra de estado aparece como pantalla), por eso los títulos "
            "de los nodos se fijan a mano en CANONICAS y un nodo por video. Limitaciones "
            "medidas: el VLM confunde letras (UIT→UT) y lee el calendario del date-picker; "
            "el OCR a veces cruza celdas vecinas. Regenerable con: "
            "`ui.py kb && ui.py cortex && herramientas/cargar_cortex.py data/cortex-borradores/pantallas-ui.json`."
        ),
        "tags": ["ui", "pantallas", "ocr", "vlm", "metodo"],
    }]
    for vid, docs in sorted(por_video.items()):
        titulo = CANONICAS.get(vid) or f"Pantalla vista en el tutorial {vid}"
        mejor = max(docs, key=lambda c: c["campos"])
        estado = [d for d in docs if d["id"] != mejor["id"]]
        estado_txt = ""
        if estado:
            alt = max(estado, key=lambda c: c["campos"])
            estado_txt = ("\n\n### Estado alternativo (grilla/listado)\n\n"
                          + "\n".join(campos_legibles({"campos": alt["esquema"]})[:12]))
        nodos.append({
            "path": f"ui-pantallas/{slug(vid)}",
            "title": titulo,
            "summary": f"Pantalla del tutorial {vid}: {mejor['campos']} campos del formulario "
                       f"detectados por OCR+visión. Borrador pendiente de validación visual de S10.",
            "body": (
                f"Pantalla «{titulo}», vista en el tutorial {mejor['url']} "
                f"(la muestran {len(docs)} capturas del video).\n\n### Formulario (detalle)\n\n"
                + "\n".join(campos_legibles({"campos": mejor["esquema"]}))
                + estado_txt
                + "\n\n**Estado:** borrador generado por OCR + visión, consolidado a un nodo "
                "por video tras la revisión del 2026-09-30 (los nodos por captura eran "
                "duplicados con títulos mal leídos). La fuente autorizada sigue siendo el "
                "manual del módulo (pendiente de acceso al portal). Validar/corregir en el "
                "HTML de `ui.py render` y volver a exportar."
            ),
            "tags": ["ui", "pantalla", "nominas"],
        })
    destino = DATA / "cortex-borradores" / "pantallas-ui.json"
    destino.parent.mkdir(exist_ok=True)
    destino.write_text(json.dumps(nodos, ensure_ascii=False, indent=1), encoding="utf-8")
    log(f"{len(nodos)} nodos borrador -> {destino.relative_to(RAIZ)} (cargar con herramientas/cargar_cortex.py)")


# ───────────────────── 5. renderizador (esquema → HTML) ────────────────────
PLANTILLA = """<!doctype html>
<html lang="es"><head><meta charset="utf-8">
<title>__TITULO__</title>
<style>
  :root { --borde: #1f6feb; --fondo: rgba(31,111,235,.08); }
  body { margin: 0; font: 13px/1.4 system-ui, sans-serif; background: #12151a; color: #e6e8eb;
         display: flex; gap: 16px; padding: 16px; }
  #lienzo { position: relative; flex: 0 0 auto; }
  #lienzo img { display: block; width: __ANCHO__px; }
  .campo { position: absolute; box-sizing: border-box; border: 1px dashed var(--borde);
           background: var(--fondo); border-radius: 3px; min-width: 60px; min-height: 18px; }
  .campo > .asa { position: absolute; inset: 0; cursor: move; }
  .campo > .eti { position: absolute; left: 0; top: -14px; font-size: 10px; color: #7aa7ff;
                  white-space: nowrap; pointer-events: none; }
  .campo input, .campo select { position: absolute; inset: 1px; border: 0; background: rgba(255,255,255,.85);
                                font-size: 11px; padding: 1px 3px; width: calc(100% - 2px); }
  #panel { flex: 1 1 320px; min-width: 300px; }
  #panel h1 { font-size: 15px; margin: 0 0 8px; }
  #panel .sub { color: #9aa4b2; margin-bottom: 10px; }
  table { border-collapse: collapse; width: 100%; }
  th, td { text-align: left; border-bottom: 1px solid #262b33; padding: 3px 4px; font-size: 12px; }
  select { background: #1b2027; color: #e6e8eb; border: 1px solid #303842; border-radius: 3px; }
  button { background: #1f6feb; color: #fff; border: 0; border-radius: 4px; padding: 6px 10px; cursor: pointer; }
  button.sec { background: #303842; }
  pre { background: #0d1117; border: 1px solid #262b33; border-radius: 4px; padding: 8px;
        max-height: 260px; overflow: auto; font-size: 11px; }
  .nota { color: #d0a24c; font-size: 11px; margin-top: 8px; }
</style></head><body>
<div id="lienzo"><img id="fondo" alt="captura del ERP" src="__IMAGEN__">__CAMPOS__</div>
<div id="panel">
  <h1>__TITULO__</h1>
  <div class="sub">Arrastra cada recuadro hasta su posición real, ajusta el tipo y exporta el esquema corregido.</div>
  <table id="tabla"><thead><tr><th>Campo</th><th>Tipo</th><th>x,y</th></tr></thead><tbody></tbody></table>
  <p><button id="exportar">Exportar esquema</button>
     <button class="sec" id="copiar">Copiar JSON</button>
     <button class="sec" id="reset">Reiniciar posiciones</button></p>
  <div class="nota">Posiciones deducidas del OCR y del VLM: son un borrador. Al mover y exportar,
    el esquema queda corregido a mano por quien ve la pantalla.</div>
  <pre id="vista"></pre>
</div>
<script>
const ESQUEMA = __ESQUEMA__;
const lienzo = document.getElementById('lienzo');
function pintar() {
  document.querySelectorAll('.campo').forEach(e => e.remove());
  const tbody = document.querySelector('#tabla tbody');
  tbody.innerHTML = '';
  ESQUEMA.campos.forEach((c, i) => {
    const d = document.createElement('div');
    d.className = 'campo';
    d.style.left = c.x + 'px'; d.style.top = c.y + 'px';
    d.style.width = c.ancho + 'px'; d.style.height = c.alto + 'px';
    d.innerHTML = '<span class="eti"></span><span class="asa"></span>';
    d.querySelector('.eti').textContent = c.etiqueta;
    const inp = document.createElement('input');
    inp.value = c.valor || '';
    d.appendChild(inp);
    let arr = null;
    const asa = d.querySelector('.asa');
    asa.addEventListener('mousedown', e => {
      arr = { x: e.clientX - c.x, y: e.clientY - c.y };
      e.preventDefault();
    });
    window.addEventListener('mousemove', e => {
      if (!arr) return;
      c.x = Math.max(0, e.clientX - arr.x); c.y = Math.max(0, e.clientY - arr.y);
      d.style.left = c.x + 'px'; d.style.top = c.y + 'px';
      tbody.rows[i].cells[2].textContent = c.x + ',' + c.y;
    });
    window.addEventListener('mouseup', () => { arr = null; });
    lienzo.appendChild(d);
    const tr = document.createElement('tr');
    tr.innerHTML = '<td></td><td></td><td>' + c.x + ',' + c.y + '</td>';
    const eti = document.createElement('input');
    eti.value = c.etiqueta; eti.title = 'VLM: ' + (c.etiqueta_vlm || '—') + ' | OCR: ' + (c.etiqueta_ocr || '—');
    eti.style.width = '100%';
    eti.addEventListener('change', () => {
      c.etiqueta = eti.value;
      document.querySelectorAll('.campo')[i].querySelector('.eti').textContent = eti.value;
      ver();
    });
    tr.cells[0].appendChild(eti);
    const sel = document.createElement('select');
    ['texto','numero','fecha','codigo','lista','booleano','texto-largo'].forEach(t => {
      const o = document.createElement('option'); o.value = t; o.textContent = t;
      if (t === c.tipo) o.selected = true; sel.appendChild(o);
    });
    sel.addEventListener('change', () => { c.tipo = sel.value; ver(); });
    tr.cells[1].appendChild(sel);
    tbody.appendChild(tr);
  });
  ver();
}
function ver() { document.getElementById('vista').textContent = JSON.stringify(ESQUEMA, null, 1); }
document.getElementById('exportar').addEventListener('click', () => {
  const b = new Blob([JSON.stringify(ESQUEMA, null, 1)], { type: 'application/json' });
  const a = document.createElement('a');
  a.href = URL.createObjectURL(b); a.download = '__ARCHIVO__';
  a.click(); URL.revokeObjectURL(a.href);
});
document.getElementById('copiar').addEventListener('click', () => {
  navigator.clipboard.writeText(JSON.stringify(ESQUEMA, null, 1));
});
document.getElementById('reset').addEventListener('click', () => {
  ESQUEMA.campos.forEach((c, i) => { c.x = c.x0; c.y = c.y0; }); pintar();
});
ESQUEMA.campos.forEach(c => {
  if (c.caja_etiqueta) { c.x0 = c.caja_etiqueta[0] + c.caja_etiqueta[2] + 6; c.y0 = c.caja_etiqueta[1]; }
  else { c.x0 = 40; c.y0 = 40; }
  c.x = c.x0; c.y = c.y0;
  c.ancho = c.ancho || 140; c.alto = c.alto || (c.caja_etiqueta ? c.caja_etiqueta[3] + 4 : 20);
});
pintar();
</script>
</body></html>
"""


def paso_render(ruta: Path) -> None:
    esq = json.loads(ruta.read_text(encoding="utf-8"))
    png = CAPTURAS / esq.get("video", "") / esq.get("captura", "")
    if not png.exists():
        raise SystemExit(f"no está la captura {png}")
    from PIL import Image
    ancho = Image.open(png).width
    imagen = "data:image/png;base64," + base64.b64encode(png.read_bytes()).decode()
    campos = []
    for c in esq["campos"]:
        campos.append({
            "etiqueta": c["etiqueta"], "tipo": c["tipo"], "valor": c.get("valor", ""),
            "etiqueta_vlm": c.get("etiqueta_vlm"), "etiqueta_ocr": c.get("etiqueta_ocr"),
            "caja_etiqueta": c.get("caja_etiqueta"), "coincidencia_ocr": c.get("coincidencia_ocr"),
        })
    esquema_web = {"pantalla": esq["pantalla"], "pantalla_vlm": esq.get("pantalla_vlm"),
                   "fuente": {"video": esq.get("video"), "captura": esq.get("captura"),
                              "modelo": esq.get("modelo")},
                   "campos": campos}
    html = (PLANTILLA
            .replace("__TITULO__", esq["pantalla"] or esq["captura"])
            .replace("__ANCHO__", str(ancho))
            .replace("__IMAGEN__", imagen)
            .replace("__CAMPOS__", "")
            .replace("__ESQUEMA__", json.dumps(esquema_web, ensure_ascii=False))
            .replace("__ARCHIVO__", f"{ruta.stem}.corregido.json"))
    salida = RAIZ / "render" / f"{ruta.stem}.html"
    salida.parent.mkdir(exist_ok=True)
    salida.write_text(html, encoding="utf-8")
    log(f"{len(campos)} campos -> {salida.relative_to(RAIZ)}  (abrir en el navegador)")


# ─────────────────────────────── CLI ───────────────────────────────────────
def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="paso", required=True)
    p1 = sub.add_parser("capturas", help="baja el video y extrae frames")
    p1.add_argument("fuente", help="id de YouTube o URL")
    for nombre in ("ocr", "campos", "vision", "esquema", "pantallas", "kb", "cortex"):
        p = sub.add_parser(nombre)
        p.add_argument("--video", help="limita a un id de video")
        if nombre == "vision":
            p.add_argument("--limite", type=int, default=0,
                           help="procesa solo las N primeras capturas (0 = todas)")
        if nombre == "esquema":
            p.add_argument("--frame", help="solo un marco, p. ej. 0011")
    p5 = sub.add_parser("render", help="esquema JSON -> HTML con el formulario sobre la captura")
    p5.add_argument("--esquema", required=True, help="ruta del JSON de data/ui/esquema/")
    a = ap.parse_args()
    if a.paso == "capturas":
        capturas(a.fuente)
    elif a.paso == "ocr":
        paso_ocr(a.video)
    elif a.paso == "campos":
        paso_campos(a.video)
    elif a.paso == "vision":
        paso_vision(a.video, a.limite)
    elif a.paso == "esquema":
        paso_esquema(a.video, a.frame)
    elif a.paso == "pantallas":
        paso_pantallas(a.video)
    elif a.paso == "kb":
        paso_kb()
    elif a.paso == "cortex":
        paso_cortex()
    elif a.paso == "render":
        paso_render(Path(a.esquema))


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit("\ninterrumpido; se retoma donde quedó al volver a correr")
