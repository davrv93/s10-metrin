#!/usr/bin/env python3
"""Generación de preguntas por plantilla fija (sin Step API key).

Usa templates determinísticos basados en el contenido del nodo.
Ideal para probar el pipeline completo sin depender de la API de generación.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
import time
from pathlib import Path
from typing import Any

import requests

STEP_API_KEY = os.environ.get("STEP_API_KEY", "")
STEP_BASE_URL = os.environ.get("STEP_BASE_URL", "https://api.stepfun.com/v1")
STEP_MODEL = os.environ.get("STEP_MODEL", "step-3.7-flash")
JEV_URL = os.environ.get("JEV_URL", "http://localhost:8080")
JEV_MODEL = os.environ.get("JEV_MODEL", "jev-style-0.8b")
PAUSA = float(os.environ.get("S10_PAUSA", "0.5"))

DIR_RAIZ = Path(__file__).resolve().parent.parent
S10_RAIZ = DIR_RAIZ.parent
DATA = S10_RAIZ / "data"
HTML_DIR = DATA / "html"
IMAGENES_JSONL = DATA / "imagenes.jsonl"
PAGINAS_JSONL = DATA / "paginas.jsonl"
DB_PATH = DATA / "preguntas.db"
FRAGMENTOS_JSONL = S10_RAIZ / "kb" / "fragmentos.jsonl"

TIPOS_PERMITIDOS = ["procedimiento", "concepto", "informacion_directa", "comparacion", "problema"]
NIVELES_DIFICULTAD = ["fácil", "media", "difícil"]
RUBRICAS = [
    ("claridad", 0.25, 0.80,
     "¿La pregunta '%s' es clara, sin ambigüedades, sin dobles negaciones y sin pronombres ambiguos?"),
    ("especificidad", 0.20, 0.75,
     "¿La pregunta '%s' menciona términos concretos del ERP S10 (módulo, función, campo, botón) y no es genérica?"),
    ("respondibilidad", 0.25, 0.85,
     "¿La pregunta '%s' se puede responder ÚNICAMENTE con la información contenida en el siguiente texto? Texto: '%s'"),
    ("utilidad", 0.20, 0.75,
     "¿Un usuario nuevo del ERP S10 haría realmente la pregunta '%s'? ¿Responde a una necesidad operativa real?"),
    ("tono", 0.10, 0.85,
     "¿La pregunta '%s' usa español neutro, profesional, sin jergas regionales ni coloquialismos?"),
]


def log(*a: Any) -> None:
    print(time.strftime("%H:%M:%S"), *a, flush=True)


def cargar_secciones_html():
    import importlib.util
    ruta = S10_RAIZ / "secciones_html.py"
    spec = importlib.util.spec_from_file_location("secciones_html", ruta)
    mod = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(mod)  # type: ignore[union-attr]
    return mod


def slug(texto: str, largo: int = 80) -> str:
    s = re.sub(r"[^a-zA-Z0-9]+", "-", texto).strip("-").lower()
    return s[:largo] or "sin-nombre"


@staticmethod
def _plantillas(nombre_nodo: str, texto: str) -> list[dict[str, str]]:
    bajo = nombre_nodo.lower() + " " + texto.lower()
    candidatos: list[dict[str, str]] = []

    def add(pregunta: str, tipo: str, respuesta: str) -> None:
        candidatos.append({"pregunta": pregunta.strip(), "tipo": tipo, "respuesta_esperada": respuesta.strip()[:200]})

    if any(k in bajo for k in ["cómo", "como", "paso", "procedimiento", "ingresar", "registrar", "crear", "generar"]):
        add(f"¿Cuál es el procedimiento para {nombre_nodo.lower()}?", "procedimiento",
            f"Pasos para {nombre_nodo.lower()} según el manual S10.")
    if any(k in bajo for k in ["qué es", "concepto", "definición", "tipo", "clase", "categoría"]):
        add(f"¿Qué es {nombre_nodo.lower()} en el módulo?", "concepto",
            f"{nombre_nodo}: definición y propósito según el manual.")
    if any(k in bajo for k in ["dónde", "donde", "ubicación", "menú", "ruta", "acceso"]):
        add(f"¿Dónde se encuentra la función de {nombre_nodo.lower()} en el ERP?", "informacion_directa",
            f"Ruta de acceso a {nombre_nodo.lower()} en el sistema.")
    if any(k in bajo for k in ["cuándo", "cuando", "fecha", "plazo", "periodo", "vigencia"]):
        add(f"¿Cuándo se debe utilizar {nombre_nodo.lower()}?", "informacion_directa",
            f"Condiciones de uso de {nombre_nodo.lower()}.")
    if any(k in bajo for k in ["requisito", "prerrequisito", "permiso", "rol", "acceso"]):
        add(f"¿Qué requisitos son necesarios para usar {nombre_nodo.lower()}?", "procedimiento",
            f"Requisitos previos para {nombre_nodo.lower()}.")
    if any(k in bajo for k in ["error", "problema", "falla", "incidencia", "solución", "solucion"]):
        add(f"¿Qué hacer si ocurre un error en {nombre_nodo.lower()}?", "problema",
            f"Pasos de solución de problemas para {nombre_nodo.lower()}.")
    if not candidatos:
        add(f"¿Cómo funciona {nombre_nodo.lower()} en el ERP S10?", "informacion_directa",
            f"Funcionamiento de {nombre_nodo.lower()} según documentación oficial.")
    # Limitar a 3 y quitar duplicados
    vistos: set[str] = set()
    unicas: list[dict[str, str]] = []
    for c in candidatos:
        if c["pregunta"] not in vistos:
            vistos.add(c["pregunta"])
            unicas.append(c)
        if len(unicas) >= 3:
            break
    return unicas


class JevClient:
    MOCK = False

    def __init__(self, url: str, model: str) -> None:
        self.url = url.rstrip("/")
        self.model = model
        self.session = requests.Session()
        self._ultima = 0.0
        if not url or url == "mock":
            JevClient.MOCK = True
            log("[jev] Modo mock activado (JEV_URL vacío o 'mock')")

    def _esperar(self) -> None:
        falta = PAUSA - (time.monotonic() - self._ultima)
        if falta > 0:
            time.sleep(falta)
        self._ultima = time.monotonic()

    def decidir(self, estado: Any, preguntas: dict[str, Any]) -> dict[str, Any]:
        if JevClient.MOCK:
            return {qid: _mock_resp(p) for qid, p in preguntas.items()}
        self._esperar()
        body = {"model": self.model, "state": estado, "questions": preguntas}
        r = self.session.post(f"{self.url}/v1/systemone", json=body, timeout=120)
        r.raise_for_status()
        return r.json()["answers"]

    def noul(self, estado: Any, instrucciones: str) -> float:
        resp = self.decidir(estado, {"q": {"type": "noul", "instructions": instrucciones, "criteria": {}}})
        return resp["q"]["noul"]

    def choice(self, estado: Any, instrucciones: str, criterios: dict[str, Any]) -> tuple[str, float]:
        resp = self.decidir(estado, {"q": {"type": "choice", "instructions": instrucciones, "criteria": criterios}})
        return resp["q"]["choice"], resp["q"]["confidence"]

    def score(self, estado: Any, instrucciones: str, criterios: list[Any]) -> tuple[float, float]:
        resp = self.decidir(estado, {"q": {"type": "score", "instructions": instrucciones, "criteria": criterios}})
        return resp["q"]["score"], resp["q"]["confidence"]


def _mock_resp(p: Any) -> dict[str, Any]:
    tipo = p.get("type", "noul")
    if tipo == "noul":
        return {"type": "noul", "noul": 0.85, "confidence": 0.85}
    if tipo == "choice":
        criterios = ((p.get("criteria") or {})) if isinstance(p.get("criteria"), dict) else {}
        opciones = list(criterios.keys())
        elegida = opciones[0] if opciones else "informacion_directa"
        return {"type": "choice", "choice": elegida, "confidence": 0.80, "probabilities": {elegida: 0.80}}
    if tipo == "score":
        criterios = p.get("criteria") or []
        return {"type": "score", "score": 1.0, "confidence": 0.80, "legend": {str(i): c for i, c in enumerate(criterios)}}
    return {"type": tipo}


def juzgar_pregunta(jev: JevClient, pregunta: str, texto_nodo: str) -> dict[str, Any]:
    scores: dict[str, float] = {}
    rubricas_raw: dict[str, Any] = {}
    estado = {"pregunta": pregunta, "texto_nodo": texto_nodo[:2000]}
    for nombre, peso, umbral, plantilla in RUBRICAS:
        try:
            if "%s" in plantilla and "%s" in plantilla.replace("%s", "", 1):
                instrucciones = plantilla % (pregunta, texto_nodo[:500])
            else:
                instrucciones = plantilla % (pregunta,)
            prob = jev.noul(estado, instrucciones)
            scores[nombre] = round(prob, 4)
            rubricas_raw[nombre] = {"prob": round(prob, 4), "umbral": umbral, "peso": peso}
        except Exception as e:
            log(f"  Jev noul falló ({nombre}): {e}")
            scores[nombre] = 0.0
            rubricas_raw[nombre] = {"prob": 0.0, "umbral": umbral, "peso": peso, "error": str(e)}
            continue
    try:
        tipo_elegido, tipo_conf = jev.choice(
            estado,
            "¿Qué tipo es esta pregunta sobre el ERP S10?",
            {t: {"desc": t} for t in TIPOS_PERMITIDOS},
        )
        rubricas_raw["tipo"] = {"elegido": tipo_elegido, "conf": round(tipo_conf, 4)}
    except Exception as e:
        log(f"  Jev choice falló: {e}")
        tipo_elegido, tipo_conf = "informacion_directa", 0.0
        rubricas_raw["tipo"] = {"elegido": tipo_elegido, "conf": 0.0, "error": str(e)}
    try:
        dif_score, dif_conf = jev.score(
            estado,
            "¿Qué dificultad tiene esta pregunta para un usuario nuevo del ERP S10?",
            [{"value": "0", "label": "fácil"}, {"value": "1", "label": "media"}, {"value": "2", "label": "difícil"}],
        )
        dif_idx = int(round(dif_score))
        dificultad = NIVELES_DIFICULTAD[min(max(dif_idx, 0), 2)]
        rubricas_raw["dificultad"] = {"score": round(dif_score, 4), "label": dificultad, "conf": round(dif_conf, 4)}
    except Exception as e:
        log(f"  Jev score falló: {e}")
        dificultad = "media"
        rubricas_raw["dificultad"] = {"score": 1.0, "label": dificultad, "conf": 0.0, "error": str(e)}
    puntaje = sum(scores[nombre] * peso for nombre, peso, _, _ in RUBRICAS)
    umbrales_ok = all(scores[nombre] >= umbral for nombre, _, umbral, _ in RUBRICAS)
    if puntaje >= 0.80 and umbrales_ok:
        veredicto = "aceptada"
    elif puntaje >= 0.65:
        veredicto = "revision"
    else:
        veredicto = "rechazada"
    return {
        "rubricas": scores,
        "rubricas_raw": rubricas_raw,
        "puntaje_global": round(puntaje, 4),
        "veredicto": veredicto,
        "tipo": tipo_elegido,
        "tipo_conf": round(tipo_conf, 4),
        "dificultad": dificultad,
    }


# ---------------------------------------------------------------------------
# Extracción de nodos desde HTML
# ---------------------------------------------------------------------------

class Nodo:
    def __init__(self, id: str, modulo: str, seccion: str, contenido: str, nodo: str, url: str, texto: str):
        self.id = id
        self.modulo = modulo
        self.seccion = seccion
        self.contenido = contenido
        self.nodo = nodo
        self.url = url
        self.texto = texto


def extraer_nodos_desde_html() -> list[Nodo]:
    mod = cargar_secciones_html()
    fotos_por_url: dict[str, str] = {}
    if IMAGENES_JSONL.exists():
        for l in IMAGENES_JSONL.read_text(encoding="utf-8").splitlines():
            if not l.strip():
                continue
            f = json.loads(l)
            fotos_por_url[f["url"]] = f["archivo"]
    paginas: dict[str, dict[str, Any]] = {}
    if PAGINAS_JSONL.exists():
        for l in PAGINAS_JSONL.read_text(encoding="utf-8").splitlines():
            if not l.strip():
                continue
            p = json.loads(l)
            paginas[p["pagina"]] = p
    nodos: list[Nodo] = []
    htmls = sorted(HTML_DIR.glob("*.html"))
    log(f"HTMLs encontrados: {len(htmls)}")
    SKIP_PATTERNS = [
        "/2019/", "/2020/", "/2021/", "/2022/", "/2023/", "/2024/", "/2025/",
        "/author/", "/category/", "/tag/", "/mes-", "/month-",
        "membership-", "/contacto-2019", "/formulario-feedback", "/home2",
        "/inicio/", "/descargas/", "/docs/", "/facturacion/", "/instalaciones/",
        "/prerrequisito-base-datos/", "/problemas-acceso/", "/publicacion-de-proyecto-lookahead/",
        "/realizar-copia-seguridad-base-datos/", "/sidebar-general-dic-19/", "/topbar-prueba/",
        "/unete/", "/uso-de-calidad-movil/", "/varios-widgets-item/", "/panel-02/",
        "/pagina-ejemplo/", "/portada-bienvenida/", "/prueba-manual-mayo/",
        "/acceso-a-manual-", "/acceso-manual-", "/nuevo-manual-", "/video-",
        "/acceso-a-manual-portal-", "/acceso-manual-portal-", "/acceso-portal-",
        "/acceso-a-manual-s10-", "/acceso-manual-s10-",
        "/manual-de-administracion-s10/", "/manual-de-almacenes-s10/",
        "/manual-de-compras-s10/", "/manual-de-gerencia-de-proyectos-s10/",
        "/manual-de-presupuestos-s10/", "/manual-de-proyectos-s10/",
        "/manual-de-s10/", "/manual-de-ventas-s10/",
        "/inicio-de-sesion/", "/registro-de-usuario/", "/restablecer-contrasena/",
        "/perfil-de-usuario/", "/miembros/", "/miembros-area/",
        "/contacto/", "/formulario/", "/formulario-de-contacto/",
        "/home-s10/", "/inicio-s10/", "/bienvenido-s10/",
    ]
    SKIP_TITLE_PREFIXES = [
        "acceso a manual", "acceso manual", "video ", "videos de",
        "membership", "login", "registro", "restablecer", "perfil",
        "acceso a manual portal", "acceso portal", "acceso s10",
        "manual de administracion s10", "manual de almacenes s10",
        "manual de compras s10", "manual de gerencia de proyectos s10",
        "manual de presupuestos s10", "manual de proyectos s10",
        "manual de s10", "manual de ventas s10",
        "inicio de sesion", "registro de usuario", "restablecer contraseña",
        "perfil de usuario", "miembros", "contacto", "formulario",
        "home s10", "inicio s10", "bienvenido s10",
        "mes:", "categoría:", "etiqueta:", "archivo:",
    ]
    for hpath in htmls:
        slug_pagina = hpath.stem
        pagina_url = None
        pagina_titulo = None
        for p in paginas.values():
            if p.get("archivo") and Path(p["archivo"]).stem == slug_pagina:
                pagina_url = p["pagina"]
                pagina_titulo = p.get("titulo") or ""
                break
        if not pagina_url:
            continue
        if any(pagina_url.startswith("https://documentacion.s10peru.com" + pat) for pat in SKIP_PATTERNS):
            continue
        if any(pagina_titulo.lower().startswith(pref) for pref in SKIP_TITLE_PREFIXES):
            continue
        html = hpath.read_text(encoding="utf-8")
        secs = mod.secciones(html, pagina_url, pagina_titulo, fotos_por_url)
        if not secs:
            continue
        modulo = pagina_titulo or slug_pagina
        seccion_actual = modulo
        contenido_actual = modulo
        for k, sec in enumerate(secs):
            titulo_sec = sec["titulo"]
            partes = [p.strip() for p in titulo_sec.split("›")]
            if len(partes) >= 3:
                modulo, seccion_actual, contenido_actual = partes[0], partes[1], partes[2]
                nodo_nombre = partes[-1]
            elif len(partes) == 2:
                seccion_actual, contenido_actual = partes[0], partes[1]
                nodo_nombre = partes[1]
            else:
                nodo_nombre = titulo_sec
            textos_pasos = [p["texto"] for p in sec["pasos"] if p["texto"].strip()]
            if not textos_pasos:
                continue
            texto_nodo = "\n".join(textos_pasos)
            if len(texto_nodo.strip()) < 40:
                continue
            nodo_id = f"{slug(modulo)}/{slug(seccion_actual)}/{slug(nodo_nombre)}-{k:03d}"
            nodos.append(Nodo(nodo_id, modulo, seccion_actual, contenido_actual, nodo_nombre, pagina_url, texto_nodo))
    log(f"Nodos extraídos: {len(nodos)}")
    return nodos


# ---------------------------------------------------------------------------
# Persistencia (SQLite local)
# ---------------------------------------------------------------------------

import sqlite3

def init_db(db_path: Path = DB_PATH) -> sqlite3.Connection:
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    conn.executescript("""
        CREATE TABLE IF NOT EXISTS metrin_preguntas_candidatas (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            nodo_id TEXT NOT NULL,
            modulo TEXT NOT NULL,
            seccion TEXT NOT NULL,
            contenido TEXT NOT NULL,
            nodo TEXT NOT NULL,
            url TEXT,
            pregunta TEXT NOT NULL,
            tipo TEXT,
            respuesta_esperada TEXT,
            modelo_generador TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS metrin_preguntas_juzgadas (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            candidata_id INTEGER NOT NULL REFERENCES metrin_preguntas_candidatas(id) ON DELETE CASCADE,
            nodo_id TEXT NOT NULL,
            rubrica_claridad REAL,
            rubrica_especificidad REAL,
            rubrica_respondibilidad REAL,
            rubrica_utilidad REAL,
            rubrica_tono REAL,
            rubrica_dificultad INTEGER,
            jev_tipo TEXT,
            jev_tipo_conf REAL,
            rubricas_raw TEXT,
            puntaje_global REAL,
            veredicto TEXT NOT NULL DEFAULT 'revision',
            vb INTEGER DEFAULT 0,
            vb_por TEXT,
            vb_at TIMESTAMP,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(candidata_id)
        );
        CREATE TABLE IF NOT EXISTS cortex_semillas (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            nodo_id TEXT NOT NULL,
            pregunta TEXT NOT NULL,
            respuesta_esperada TEXT,
            tipo TEXT,
            origen TEXT DEFAULT 'generacion_automatica',
            vb INTEGER DEFAULT 0,
            vb_por TEXT,
            vb_at TIMESTAMP,
            jev_claridad REAL,
            jev_dificultad INTEGER,
            jev_tipo_conf REAL,
            rubricas TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(nodo_id, pregunta)
        );
        CREATE INDEX IF NOT EXISTS ix_juzgadas_nodo ON metrin_preguntas_juzgadas(nodo_id);
        CREATE INDEX IF NOT EXISTS ix_juzgadas_vb ON metrin_preguntas_juzgadas(vb);
        CREATE INDEX IF NOT EXISTS ix_semillas_vb ON cortex_semillas(vb);
    """)
    return conn


def persistir_candidata(conn: sqlite3.Connection, fila: dict[str, Any]) -> int:
    cur = conn.execute("""
        INSERT INTO metrin_preguntas_candidatas
            (nodo_id, modulo, seccion, contenido, nodo, url, pregunta, tipo, respuesta_esperada, modelo_generador)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    """, (
        fila["nodo_id"], fila["modulo"], fila["seccion"], fila["contenido"], fila["nodo"],
        fila.get("url"), fila["pregunta"], fila.get("tipo"), fila.get("respuesta_esperada"), fila.get("modelo_generador"),
    ))
    conn.commit()
    return cur.lastrowid


def persistir_juzgada(conn: sqlite3.Connection, candidata_id: int, nodo_id: str, res: dict[str, Any]) -> None:
    conn.execute("""
        INSERT OR REPLACE INTO metrin_preguntas_juzgadas
            (candidata_id, nodo_id, rubrica_claridad, rubrica_especificidad, rubrica_respondibilidad,
             rubrica_utilidad, rubrica_tono, rubrica_dificultad, jev_tipo, jev_tipo_conf,
             rubricas_raw, puntaje_global, veredicto)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    """, (
        candidata_id, nodo_id,
        res["rubricas"].get("claridad"), res["rubricas"].get("especificidad"),
        res["rubricas"].get("respondibilidad"), res["rubricas"].get("utilidad"),
        res["rubricas"].get("tono"),
        0 if res["dificultad"] == "fácil" else 1 if res["dificultad"] == "media" else 2,
        res["tipo"], res["tipo_conf"],
        json.dumps(res["rubricas_raw"], ensure_ascii=False),
        res["puntaje_global"], res["veredicto"],
    ))
    conn.commit()


def publicar_semilla(conn: sqlite3.Connection, nodo_id: str, pregunta: str, respuesta_esperada: str,
                     tipo: str, res: dict[str, Any]) -> None:
    conn.execute("""
        INSERT OR REPLACE INTO cortex_semillas
            (nodo_id, pregunta, respuesta_esperada, tipo, origen, vb, vb_por, vb_at,
             jev_claridad, jev_dificultad, jev_tipo_conf, rubricas)
        VALUES (?, ?, ?, ?, ?, 1, 'auto', CURRENT_TIMESTAMP, ?, ?, ?, ?)
    """, (
        nodo_id, pregunta, respuesta_esperada, tipo, "generacion_automatica",
        res["rubricas"].get("claridad"),
        0 if res["dificultad"] == "fácil" else 1 if res["dificultad"] == "media" else 2,
        res["tipo_conf"],
        json.dumps(res["rubricas_raw"], ensure_ascii=False),
    ))
    conn.commit()


def publicar_vb(conn: sqlite3.Connection, min_puntaje: float = 0.90, min_rubrica: float = 0.85, vb_por: str = "auto") -> int:
    publicadas = 0
    cur = conn.execute("""
        SELECT j.id, j.nodo_id, j.puntaje_global, c.pregunta, c.respuesta_esperada, c.tipo, j.rubricas_raw
        FROM metrin_preguntas_juzgadas j
        JOIN metrin_preguntas_candidatas c ON c.id = j.candidata_id
        WHERE j.veredicto = 'aceptada' AND j.vb = 0
          AND j.puntaje_global >= ?
    """, (min_puntaje,))
    for row in cur:
        rubricas = json.loads(row["rubricas_raw"] or "{}")
        umbrales_ok = all(rubricas.get(nombre, {}).get("prob", 0) >= min_rubrica for nombre, _, _, _ in RUBRICAS)
        if not umbrales_ok:
            continue
        publicar_semilla(
            conn, row["nodo_id"], row["pregunta"], row["respuesta_esperada"] or "",
            row["tipo"] or "informacion_directa",
            {
                "rubricas": {n: rubricas.get(n, {}).get("prob", 0) for n, _, _, _ in RUBRICAS},
                "rubricas_raw": rubricas,
                "dificultad": "media",
                "tipo": row["tipo"],
                "tipo_conf": 0.0,
            },
        )
        conn.execute("UPDATE metrin_preguntas_juzgadas SET vb=1, vb_por=?, vb_at=CURRENT_TIMESTAMP WHERE id=?",
                     (vb_por, row["id"]))
        conn.commit()
        publicadas += 1
    log(f"Publicadas en cortex_semillas con VB: {publicadas}")
    return publicadas


def stats(conn: sqlite3.Connection) -> None:
    cur = conn.execute("SELECT veredicto, COUNT(*) FROM metrin_preguntas_juzgadas GROUP BY veredicto")
    log("Distribución de veredictos:")
    for row in cur:
        log(f"  {row[0]}: {row[1]}")
    cur = conn.execute("SELECT COUNT(*) FROM cortex_semillas WHERE vb=1")
    log(f"Semillas publicadas (VB): {cur.fetchone()[0]}")


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main() -> None:
    ap = argparse.ArgumentParser(description="Generación por plantilla + Jev")
    ap.add_argument("paso", choices=["generar", "juzgar", "vb", "stats", "todo"])
    ap.add_argument("--limite", type=int, default=0, help="máx nodos a procesar (0 = todos)")
    ap.add_argument("--min-puntaje", type=float, default=0.90, help="umbral VB automático")
    ap.add_argument("--min-rubrica", type=float, default=0.85, help="umbral mínimo por rúbrica")
    ap.add_argument("--vb-por", default="auto", help="quién da el VB")
    args = ap.parse_args()

    jev = JevClient(JEV_URL, JEV_MODEL)
    conn = init_db()

    if args.paso == "stats":
        stats(conn)
        return

    nodos = extraer_nodos_desde_html()
    if not nodos:
        raise SystemExit("No se extrajeron nodos desde HTML")

    if args.paso == "generar":
        filas: list[dict[str, Any]] = []
        total = len(nodos) if args.limite <= 0 else min(args.limite, len(nodos))
        for i, nodo in enumerate(nodos[:total], 1):
            log(f"[{i}/{total}] Nodo: {nodo.modulo} › {nodo.seccion} › {nodo.nodo}")
            candidatas = _plantillas(nodo.nodo, nodo.texto)
            if not candidatas:
                continue
            for c in candidatas:
                filas.append({
                    "nodo_id": nodo.id, "modulo": nodo.modulo, "seccion": nodo.seccion,
                    "contenido": nodo.contenido, "nodo": nodo.nodo, "url": nodo.url,
                    "pregunta": c["pregunta"], "tipo": c["tipo"], "respuesta_esperada": c["respuesta_esperada"],
                    "modelo_generador": "plantilla-fija-v1", "texto_nodo": nodo.texto,
                })
        for fila in filas:
            persistir_candidata(conn, fila)
        log(f"Persistidas {len(filas)} candidatas en {DB_PATH}")
        return

    if args.paso == "juzgar":
        candidatas = []
        for row in conn.execute("SELECT * FROM metrin_preguntas_candidatas WHERE id NOT IN (SELECT candidata_id FROM metrin_preguntas_juzgadas)"):
            candidatas.append(dict(row))
        if not candidatas:
            log("No hay candidatas pendientes de juzgar")
            return
        log(f"Candidatas a juzgar: {len(candidatas)}")
        for i, c in enumerate(candidatas, 1):
            texto_nodo = c.get("texto_nodo") or ""
            if not texto_nodo:
                # Intentar recuperar desde el nodo
                texto_nodo = ""
            log(f"[{i}/{len(candidatas)}] Juzgando: {c['pregunta'][:80]}...")
            res = juzgar_pregunta(jev, c["pregunta"], texto_nodo)
            candidata_id = c["id"]
            persistir_juzgada(conn, candidata_id, c["nodo_id"], res)
        stats(conn)
        return

    if args.paso == "vb":
        publicadas = publicar_vb(conn, args.min_puntaje, args.min_rubrica, args.vb_por)
        log(f"VB aplicado a {publicadas} preguntas")
        stats(conn)
        return

    if args.paso == "todo":
        filas: list[dict[str, Any]] = []
        total = len(nodos) if args.limite <= 0 else min(args.limite, len(nodos))
        for i, nodo in enumerate(nodos[:total], 1):
            log(f"[{i}/{total}] Nodo: {nodo.modulo} › {nodo.seccion} › {nodo.nodo}")
            candidatas = _plantillas(nodo.nodo, nodo.texto)
            if not candidatas:
                continue
            for c in candidatas:
                filas.append({
                    "nodo_id": nodo.id, "modulo": nodo.modulo, "seccion": nodo.seccion,
                    "contenido": nodo.contenido, "nodo": nodo.nodo, "url": nodo.url,
                    "pregunta": c["pregunta"], "tipo": c["tipo"], "respuesta_esperada": c["respuesta_esperada"],
                    "modelo_generador": "plantilla-fija-v1", "texto_nodo": nodo.texto,
                })
        for fila in filas:
            persistir_candidata(conn, fila)
        log(f"Persistidas {len(filas)} candidatas")
        # Juzgar
        for i, c in enumerate(filas, 1):
            log(f"[{i}/{len(filas)}] Juzgando: {c['pregunta'][:80]}...")
            res = juzgar_pregunta(jev, c["pregunta"], c["texto_nodo"])
            candidata_id = c.get("id")
            if candidata_id is None:
                candidata_id = persistir_candidata(conn, c)
            persistir_juzgada(conn, candidata_id, c["nodo_id"], res)
        publicar_vb(conn, args.min_puntaje, args.min_rubrica, args.vb_por)
        stats(conn)
        return


if __name__ == "__main__":
    main()
