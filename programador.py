#!/usr/bin/env python3
"""
programador.py — actualiza las fuentes cada X horas y deja constancia en Cortex.

  .venv/bin/python programador.py            # bucle (en Docker es el servicio `programador`)
  .venv/bin/python programador.py --una portal-s10    # corre una fuente ya y sale

Configuración (la edita el panel): data/programacion.json
  {"<id>": {"cada_horas": 24, "activo": true}}
Estado (lo escribe este programa):  data/programacion_estado.json
Pedidos «ejecutar ahora» del panel: data/programacion_pedidos/<id>

Tras cada corrida con cambios: reindexa la base, avisa a Metrín (kb/.actualizado)
y registra en Cortex una entrada de actividad + el nodo `fuentes/programacion`.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
import time
from pathlib import Path

import requests

RAIZ = Path(__file__).resolve().parent
DATA, KB = RAIZ / "data", RAIZ / "kb"
CONFIG, ESTADO = DATA / "programacion.json", DATA / "programacion_estado.json"
PEDIDOS, LOGS = DATA / "programacion_pedidos", DATA / "programacion_logs"
PY = sys.executable
CORTEX = os.environ.get("S10_CORTEX_API", "http://localhost:4748/api")

FUENTES = {
    "portal-s10": {"nombre": "Portal de ayuda S10 (páginas públicas)", "cada_horas": 24,
                   "pasos": [["s10kb.py", "rastrear", "--sin-login", "--refrescar"]]},
    "manuales-oficiales": {"nombre": "Manuales oficiales S10 (portal de miembros)", "cada_horas": 168,
                           "pasos": [["s10kb.py", "oficial"]]},
    "pdfs-s10peru": {"nombre": "PDF públicos de s10peru.com", "cada_horas": 168,
                     "pasos": [["s10kb.py", "pdfs-publicos"], ["s10kb.py", "descargar", "--sin-login"],
                               ["s10kb.py", "ocr", "--completo"]]},
    "optimiza360": {"nombre": "Sitio optimiza360.pe", "cada_horas": 168,
                    "pasos": [["s10kb.py", "sitio", "optimiza360.pe"]]},
    "youtube": {"nombre": "Canal oficial S10 en YouTube", "cada_horas": 168,
                "pasos": [["youtube.py", "todo"]]},
}


def log(*a) -> None:
    print(time.strftime("%Y-%m-%d %H:%M:%S"), *a, flush=True)


def leer(p: Path) -> dict:
    try:
        return json.loads(p.read_text(encoding="utf-8")) if p.exists() else {}
    except json.JSONDecodeError:
        return {}


def guardar(p: Path, d: dict) -> None:
    tmp = p.with_suffix(".tmp")
    tmp.write_text(json.dumps(d, indent=1, ensure_ascii=False), encoding="utf-8")
    tmp.replace(p)


def config() -> dict:
    c = leer(CONFIG)
    cambio = False
    for fid, f in FUENTES.items():
        if fid not in c:
            c[fid] = {"cada_horas": f["cada_horas"], "activo": True}
            cambio = True
    if cambio:
        guardar(CONFIG, c)
    return c


def contar() -> dict:
    def lineas(p: Path) -> int:
        return sum(1 for l in p.read_text(encoding="utf-8").splitlines() if l.strip()) if p.exists() else 0
    return {
        "fragmentos": sum(lineas(f) for f in KB.glob("fragmentos*.jsonl")),
        "pdf": sum(1 for l in (DATA / "manifiesto.jsonl").read_text().splitlines() if '"estado": "ok"' in l) if (DATA / "manifiesto.jsonl").exists() else 0,
        "paginas": lineas(DATA / "paginas.jsonl"),
        "videos": sum(1 for f in (DATA / "yt" / "transcripciones").glob("*.txt") if f.stat().st_size > 20) if (DATA / "yt" / "transcripciones").exists() else 0,
    }


# ─────────────────────────────── Cortex ────────────────────────────────────
def _h() -> dict:
    sec = (RAIZ / ".cortex/.secrets.yaml").read_text()
    h = {"Authorization": "Bearer " + re.search(r"ai-agent:\s*(\S+)", sec).group(1), "Content-Type": "application/json"}
    if os.environ.get("S10_CORTEX_HOST"):
        h["Host"] = os.environ["S10_CORTEX_HOST"]
    return h


def avisar_cortex(fid: str, e: dict) -> None:
    try:
        h = _h()
        estado = leer(ESTADO)
        conf = config()
        filas = []
        for k, f in FUENTES.items():
            s, c = estado.get(k, {}), conf.get(k, {})
            frecuencia = f"cada {c.get('cada_horas')} h" if c.get("activo") else "pausada"
            filas.append(f"| {f['nombre']} | {frecuencia} | {s.get('ultima', '—')} | {s.get('estado', '—')} | {s.get('resumen', '—')} |")
        cuerpo = ("Actualización automática de las fuentes (servicio `programador`). "
                  "La frecuencia se cambia en el panel de administración, sección Programación.\n\n"
                  "| Fuente | Frecuencia | Última corrida | Estado | Resultado |\n|---|---|---|---|---|\n" + "\n".join(filas) +
                  "\n\nCada corrida con cambios reindexa la base y Metrín recarga su índice. "
                  "El historial detallado está en la actividad de Cortex.\n\n**Confiabilidad:** interno del proyecto.")
        requests.put(f"{CORTEX}/node/fuentes/programacion", headers=h, timeout=15, json={
            "title": "Programación de fuentes",
            "summary": "Cuándo se actualizó por última vez cada fuente (portal S10, PDF de s10peru.com, optimiza360.pe, YouTube), cada cuánto corre y qué trajo.",
            "body": cuerpo, "tags": ["fuentes", "programacion", "interno"],
            "reason": f"Corrida programada de {fid}",
        })
        r = requests.post(f"{CORTEX}/activity", headers=h, timeout=15, json={
            "action": "other",
            "summary": f"{FUENTES[fid]['nombre']}: {e['resumen']}"[:200],
            "why": f"Actualización programada cada {e.get('cada_horas')} h. Estado: {e['estado']}. Duración: {e.get('duracion_s', 0)} s.",
            "refs": ["fuentes/programacion"],
        })
        if r.status_code >= 300:
            log(f"  Cortex rechazó la actividad: {r.status_code} {r.text[:200]}")
    except Exception as ex:  # Cortex apagado no debe frenar la actualización
        log(f"  Cortex no disponible: {ex}")


# ─────────────────────────────── ejecutar ──────────────────────────────────
def ejecutar(fid: str, motivo: str) -> dict:
    f, conf = FUENTES[fid], config().get(fid, {})
    LOGS.mkdir(parents=True, exist_ok=True)
    antes, t0 = contar(), time.time()
    estado = leer(ESTADO)
    estado[fid] = {**estado.get(fid, {}), "estado": "corriendo", "inicio": time.strftime("%Y-%m-%d %H:%M")}
    guardar(ESTADO, estado)
    log(f"▶ {f['nombre']} ({motivo})")
    salida, ok = [], True
    for paso in f["pasos"] + [["s10kb.py", "indexar"], ["herramientas/catalogo_metrin.py"]]:
        r = subprocess.run([PY, *paso], cwd=RAIZ, capture_output=True, text=True)
        salida.append(f"$ {' '.join(paso)}\n{r.stdout[-6000:]}{r.stderr[-2000:]}")
        if r.returncode != 0:
            ok = False
            break
    despues = contar()
    delta = {k: despues[k] - antes[k] for k in despues}
    partes = [f"{'+' if v > 0 else ''}{v} {k}" for k, v in delta.items() if v]
    # cambios de contenido sin cambio de cantidad (páginas reescritas)
    cambiadas = sum(int(m) for m in re.findall(r"'cambiadas': (\d+)", "\n".join(salida)))
    if cambiadas:
        partes.append(f"{cambiadas} páginas actualizadas")
    resumen = ", ".join(partes) or "sin cambios"
    if not ok:
        resumen = "falló: " + (salida[-1].strip().splitlines() or ["?"])[-1][:150]
    elif partes:
        (KB / ".actualizado").write_text(time.strftime("%Y-%m-%dT%H:%M:%S"))   # Metrín recarga su índice
    registro = LOGS / f"{fid}.log"
    registro.write_text(f"{time.strftime('%Y-%m-%d %H:%M:%S')} · {motivo}\n\n" + "\n\n".join(salida), encoding="utf-8")
    e = {"estado": "ok" if ok else "error", "ultima": time.strftime("%Y-%m-%d %H:%M"), "resumen": resumen,
         "duracion_s": int(time.time() - t0), "motivo": motivo, "cada_horas": conf.get("cada_horas"),
         "proxima": time.time() + conf.get("cada_horas", 24) * 3600, "ultima_ts": time.time()}
    estado = leer(ESTADO)
    estado[fid] = e
    guardar(ESTADO, estado)
    log(f"  {e['estado']}: {resumen} ({e['duracion_s']} s)")
    avisar_cortex(fid, e)
    return e


def toca(fid: str, conf: dict, estado: dict) -> str | None:
    if (PEDIDOS / fid).exists():
        (PEDIDOS / fid).unlink(missing_ok=True)
        return "pedido desde el panel"
    c = conf.get(fid, {})
    if not c.get("activo") or not c.get("cada_horas"):
        return None
    ultima = estado.get(fid, {}).get("ultima_ts", 0)
    return "programada" if time.time() - ultima >= c["cada_horas"] * 3600 else None


def bucle() -> None:
    PEDIDOS.mkdir(parents=True, exist_ok=True)
    log("programador en marcha: " + ", ".join(FUENTES))
    # al arrancar no se dispara todo de golpe: las fuentes sin historial esperan su primer ciclo
    estado = leer(ESTADO)
    for fid in FUENTES:
        estado.setdefault(fid, {"estado": "en espera", "ultima_ts": time.time(), "ultima": "—", "resumen": "aún no corre",
                                "proxima": time.time() + config()[fid]["cada_horas"] * 3600})
    guardar(ESTADO, estado)
    while True:
        conf, estado = config(), leer(ESTADO)
        for fid in FUENTES:
            motivo = toca(fid, conf, estado)
            if motivo:
                try:
                    ejecutar(fid, motivo)
                except Exception as ex:
                    log(f"  error inesperado en {fid}: {ex}")
                break      # una fuente por vuelta: los pedidos del panel no esperan a las demás
        time.sleep(20)


if __name__ == "__main__":
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--una", choices=list(FUENTES), help="corre esa fuente ya y termina")
    a = ap.parse_args()
    if a.una:
        config()
        ejecutar(a.una, "manual")
    else:
        bucle()
