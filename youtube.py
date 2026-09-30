#!/usr/bin/env python3
"""
youtube.py — convierte los tutoriales de YouTube del canal oficial S10 en base
de conocimiento (texto troceado en JSONL).

Fuente: canal "Marketing S10" (UCL1xwXCVt9ZVpSrgyemP50A), videos públicos que
la propia empresa publicó para sus clientes. Se baja SOLO audio, se transcribe
en local con faster-whisper y se indexa por temas con timestamps.

Pasos (cada uno reanudable):

  descubrir    inventario del canal + playlists    -> data/yt/videos.txt
  audio        baja el audio de cada video         -> data/yt/audio/{id}.m4a
  transcribir  faster-whisper en local             -> data/yt/transcripciones/{id}.json y .txt
  indexar      trocea con minutos                  -> kb/fragmentos_yt.jsonl, kb/videos.json
  todo         los cuatro en orden

Uso:
  .venv/bin/python youtube.py todo
  .venv/bin/python youtube.py transcribir --modelo small
"""
from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
import time
from pathlib import Path

# reutiliza las utilidades del proyecto (log, jsonl, env, slug)
sys.path.insert(0, str(Path(__file__).resolve().parent))
from s10kb import RAIZ, DATA, agregar_jsonl, cargar_env, leer_jsonl, log, slug

YT = DATA / "yt"
AUDIO = YT / "audio"
TRANSC = YT / "transcripciones"
KB = RAIZ / "kb"

CANAL = "UCL1xwXCVt9ZVpSrgyemP50A"  # Marketing S10 — canal oficial de tutoriales
PLAYLISTS = [
    "PL-iFIkPfIRvTE0T0sb2TlAmpa9u2jfP83",  # Portal de Asistencia
    "PL-iFIkPfIRvRUvPEGHv1_FtOdjhoytOHq",  # Tareo en Línea
    "PL-iFIkPfIRvT091daBuel7VNEyWMvhxkL",  # Soporte S10
    "PL-iFIkPfIRvRtaRvNa6lIUU1zjGxO1-v3",  # Pedidos Móviles S10
    "PL-iFIkPfIRvQk3A2gBcPGbJnM5wlB7Vd9",  # Aprende S10
    "PL-iFIkPfIRvQHBwdOM8b-DC7YTyiXpQl_",  # Aprobaciones móviles
    "PL-iFIkPfIRvQxBH8h7ZY8LCS_8DnffmjV",  # Cursos S10
    "PL-iFIkPfIRvTjAf_Y_aGg45bk9YJV49u6",  # Nuestra Marca
    "PL-iFIkPfIRvRF9OP2hnv_kLsTvZzjtfiY",  # Nuestra Marca (2)
    "PL-iFIkPfIRvQbM9ArRjx1jM0PtcMmP_np",  # Portal Colaborador
    "PL-iFIkPfIRvS1fBIQmayGl-FNu3j8-k_z",  # Calidad Movil - S10
    "PL-iFIkPfIRvRU9F0Stsquv6dDhaDdUHnP",  # Portal de Proveedores - S10
]


def ytbin() -> str:
    local = RAIZ / ".venv" / "bin" / "yt-dlp"
    return str(local) if local.exists() else "yt-dlp"


def mmss(seg: float) -> str:
    s = int(seg)
    return f"{s // 60}:{s % 60:02d}"


# ─────────────────────────────── 0. descubrir ──────────────────────────────
def descubrir() -> list[dict]:
    YT.mkdir(parents=True, exist_ok=True)
    filas: dict[str, dict] = {}
    destinos = [f"https://www.youtube.com/channel/{CANAL}/videos"] + [
        f"https://www.youtube.com/playlist?list={p}" for p in PLAYLISTS
    ]
    for i, url in enumerate(destinos, 1):
        r = subprocess.run(
            [ytbin(), "--flat-playlist", "--print", "%(id)s|%(duration)s|%(title)s", url],
            capture_output=True, text=True, timeout=180,
        )
        n = 0
        for linea in (r.stdout or "").splitlines():
            partes = linea.split("|", 2)
            if len(partes) != 3 or not re.fullmatch(r"[\w-]{11}", partes[0]):
                continue
            vid, dur, titulo = partes
            dur = int(dur) if dur.isdigit() else None
            filas.setdefault(vid, {"id": vid, "duracion": dur, "titulo": titulo.strip()})
            n += 1
        log(f"[{i}/{len(destinos)}] {n} videos: {url.rsplit('/', 1)[-1][:50]}")
    salida = sorted(filas.values(), key=lambda f: (f["titulo"] or "").lower())
    (YT / "videos.txt").write_text(
        "\n".join(f"{f['id']}|{f['duracion'] or ''}|{f['titulo']}" for f in salida) + "\n",
        encoding="utf-8",
    )
    total = sum(f["duracion"] or 0 for f in salida)
    log(f"{len(salida)} videos únicos, ~{total // 60} min de audio -> data/yt/videos.txt")
    return salida


def inventario() -> list[dict]:
    """Lee data/yt/videos.txt; si no existe, lo descubre."""
    f = YT / "videos.txt"
    if not f.exists():
        return descubrir()
    out = []
    for linea in f.read_text(encoding="utf-8").splitlines():
        partes = (linea.split("|", 2) + ["", ""])[:3]
        if re.fullmatch(r"[\w-]{11}", partes[0]):
            out.append({"id": partes[0], "duracion": int(partes[1]) if partes[1].isdigit() else None,
                        "titulo": partes[2]})
    return out


# ─────────────────────────────── 1. audio ──────────────────────────────────
def audio(inventario: list[dict]) -> None:
    AUDIO.mkdir(parents=True, exist_ok=True)
    lista = YT / "lista_ids.txt"
    lista.write_text("\n".join(f["id"] for f in inventario) + "\n", encoding="utf-8")
    r = subprocess.run(
        [ytbin(), "-a", str(lista),
         "-f", "bestaudio[ext=m4a]/bestaudio",
         "-o", str(AUDIO / "%(id)s.%(ext)s"),
         "--download-archive", str(YT / "bajados.txt"),
         "--sleep-requests", "1", "--no-overwrites", "-q"],
        capture_output=True, text=True, timeout=3600,
    )
    ok = sorted(p.stem for p in AUDIO.glob("*.m4a")) + sorted(p.stem for p in AUDIO.glob("*.webm"))
    faltan = [f["id"] for f in inventario if f["id"] not in ok]
    log(f"audio listo: {len(ok)}/{len(inventario)}" + (f"; sin audio: {faltan}" if faltan else ""))
    if r.returncode not in (0,):
        log(f"yt-dlp avisó: {(r.stderr or '').strip()[-300:]}")


# ─────────────────────────────── 2. transcribir ────────────────────────────
def transcribir(inventario: list[dict], modelo: str = "small") -> None:
    from faster_whisper import WhisperModel

    TRANSC.mkdir(parents=True, exist_ok=True)
    titulos = {f["id"]: f for f in inventario}
    pistas = sorted(p for p in AUDIO.glob("*.m4a")) + sorted(p for p in AUDIO.glob("*.webm"))
    pendientes = [p for p in pistas if not (TRANSC / f"{p.stem}.json").exists()]
    if not pendientes:
        log("nada pendiente de transcribir")
        return
    log(f"{len(pendientes)} audios por transcribir (modelo {modelo}, esto toma su tiempo)")
    m = WhisperModel(modelo, device="cpu", compute_type="int8")
    manifiesto = {e["id"]: e for e in leer_jsonl(YT / "transcripciones.jsonl")}
    for i, p in enumerate(pendientes, 1):
        info_v = titulos.get(p.stem, {})
        t0 = time.monotonic()
        segmentos, info = m.transcribe(str(p), language="es", vad_filter=True,
                                       vad_parameters={"min_silence_duration_ms": 500})
        filas, texto = [], []
        for s in segmentos:
            filas.append({"ini": round(s.start, 1), "fin": round(s.end, 1), "texto": s.text.strip()})
            texto.append(s.text.strip())
        (TRANSC / f"{p.stem}.json").write_text(json.dumps(filas, ensure_ascii=False, indent=1), encoding="utf-8")
        (TRANSC / f"{p.stem}.txt").write_text("\n".join(texto) + "\n", encoding="utf-8")
        manifiesto[p.stem] = {
            "id": p.stem, "titulo": info_v.get("titulo") or p.stem,
            "url": f"https://youtu.be/{p.stem}", "duracion": info_v.get("duracion"),
            "lenguaje": info.language, "segmentos": len(filas),
            "modelo": modelo, "transcrito": time.strftime("%Y-%m-%dT%H:%M:%S"),
        }
        log(f"[{i}/{len(pendientes)}] {p.stem} {len(filas)} segmentos en {time.monotonic() - t0:.0f}s — {(info_v.get('titulo') or '')[:50]}")
        # manifiesto en disco tras CADA video: una corrida cortada no pierde lo hecho
        (YT / "transcripciones.jsonl").write_text(
            "\n".join(json.dumps(manifiesto[k], ensure_ascii=False) for k in sorted(manifiesto)) + "\n",
            encoding="utf-8",
        )
    recomponer_manifiesto(inventario)


def recomponer_manifiesto(inv: list[dict] | None = None) -> None:
    """Reconstruye transcripciones.jsonl desde los .json en disco (idempotente)."""
    inv = inv if inv is not None else inventario()
    titulos = {f["id"]: f for f in inv}
    manifiesto = {e["id"]: e for e in leer_jsonl(YT / "transcripciones.jsonl")}
    for f in sorted(TRANSC.glob("*.json")):
        if f.stem in manifiesto:
            continue
        filas = json.loads(f.read_text(encoding="utf-8"))
        info_v = titulos.get(f.stem, {})
        manifiesto[f.stem] = {
            "id": f.stem, "titulo": info_v.get("titulo") or f.stem,
            "url": f"https://youtu.be/{f.stem}", "duracion": info_v.get("duracion"),
            "lenguaje": "es", "segmentos": len(filas), "modelo": "?",
            "transcrito": time.strftime("%Y-%m-%dT%H:%M:%S", time.localtime(f.stat().st_mtime)),
        }
    (YT / "transcripciones.jsonl").write_text(
        "\n".join(json.dumps(manifiesto[k], ensure_ascii=False) for k in sorted(manifiesto)) + "\n",
        encoding="utf-8",
    )


# ─────────────────────────────── 3. indexar ────────────────────────────────
def indexar(tam: int = 1500, solape: int = 200) -> None:
    KB.mkdir(exist_ok=True)
    salida = KB / "fragmentos_yt.jsonl"
    salida.write_text("", encoding="utf-8")
    videos = []
    total = 0
    for e in leer_jsonl(YT / "transcripciones.jsonl"):
        f = TRANSC / f"{e['id']}.json"
        if not f.exists():
            continue
        segmentos = json.loads(f.read_text(encoding="utf-8"))
        if not segmentos:
            continue
        trozos, buf, ini, fin = [], "", segmentos[0]["ini"], segmentos[0]["ini"]
        for s in segmentos:
            if buf and len(buf) + len(s["texto"]) + 1 > tam:
                trozos.append((buf.strip(), ini, fin))
                buf, ini = buf[-solape:] + " ", s["ini"]
            if not buf:
                ini = s["ini"]
            buf += s["texto"] + " "
            fin = s["fin"]
        if buf.strip():
            trozos.append((buf.strip(), ini, fin))
        doc_id = f"yt-{e['id']}"
        videos.append({
            "id": doc_id, "titulo": e["titulo"], "url": e["url"], "duracion": e.get("duracion"),
            "segmentos": len(segmentos), "fragmentos": len(trozos), "modelo": e.get("modelo"),
        })
        for k, (texto, d, h) in enumerate(trozos):
            agregar_jsonl(salida, {
                "id": f"{doc_id}-{k:04d}", "documento": doc_id, "tipo": "video",
                "canal": "Marketing S10 (canal oficial S10)", "video": e["titulo"],
                "desde": mmss(d), "hasta": mmss(h), "segundos": [round(d), round(h)],
                "url": f"{e['url']}?t={int(d)}", "texto": texto,
            })
        total += len(trozos)
        log(f"{e['titulo'][:60]}: {len(trozos)} fragmentos")
    (KB / "videos.json").write_text(json.dumps(videos, indent=1, ensure_ascii=False), encoding="utf-8")
    log(f"{len(videos)} videos indexados, {total} fragmentos -> kb/fragmentos_yt.jsonl")


# ─────────────────────────────── CLI ───────────────────────────────────────
def main() -> None:
    cargar_env()
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("paso", choices=["descubrir", "audio", "transcribir", "indexar", "todo"])
    ap.add_argument("--modelo", default="small", help="tiny|base|small|medium|large-v3 (faster-whisper)")
    a = ap.parse_args()

    inv = inventario() if a.paso != "descubrir" else []
    if a.paso in ("descubrir", "todo"):
        inv = descubrir()
    if a.paso in ("audio", "todo"):
        audio(inv)
    if a.paso in ("transcribir", "todo"):
        transcribir(inv, a.modelo)
    if a.paso in ("indexar", "todo"):
        indexar()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit("\ninterrumpido; se retoma donde quedó al volver a correr")
