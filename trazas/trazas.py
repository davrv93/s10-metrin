#!/usr/bin/env python3
"""
trazas.py — visor de trazas de decisiones JEV (jeva.cpp, POST /v1/systemone).

  .venv/bin/python trazas/trazas.py        # http://127.0.0.1:4761

Cada traza guarda una petición de decisión: el estado compartido, las
preguntas (choice, score, noul) con sus criterios, y las respuestas del
servidor (opción elegida, distribución de probabilidades, confianza,
score esperado o probabilidad de true), más latencia, modelo y tokens.

La capa de datos es una abstracción (RepositorioTrazas). Hoy lee mocks
(mock/trazas.json) para desarrollar y evaluar el visor; con el esquema
de Postgres (Fase 3/5) entra RepositorioPostgres y el API no cambia.

Solo escucha en 127.0.0.1. Sin autenticación: es una herramienta de
evaluación local, igual que el chat de Metrín.
"""
from __future__ import annotations

import json
import os
import statistics
from pathlib import Path

from flask import Flask, jsonify, render_template, request

RAIZ = Path(__file__).resolve().parent
PUERTO = int(os.environ.get("TRAZAS_PUERTO", "4761"))
HOST = os.environ.get("TRAZAS_HOST", "127.0.0.1")

TIPOS = ("choice", "score", "noul")


# ─────────────────────────── capa de datos ──────────────────────────
class RepositorioTrazas:
    """Contrato del visor. El Postgres de Fase 3/5 implementa lo mismo."""

    def listar(self) -> list[dict]:
        raise NotImplementedError

    def obtener(self, traza_id: str) -> dict | None:
        raise NotImplementedError

    def origen(self) -> str:
        raise NotImplementedError


class RepositorioMock(RepositorioTrazas):
    """Lee mock/trazas.json (repositorio plano, sin dependencias)."""

    def __init__(self, ruta: Path):
        self.ruta = ruta
        self._trazas: list[dict] | None = None

    def _cargar(self) -> list[dict]:
        if self._trazas is None:
            with open(self.ruta, encoding="utf-8") as f:
                self._trazas = json.load(f)["trazas"]
        return self._trazas

    def listar(self) -> list[dict]:
        return self._cargar()

    def obtener(self, traza_id: str) -> dict | None:
        for t in self._cargar():
            if t["id"] == traza_id:
                return t
        return None

    def origen(self) -> str:
        return "mocks (" + self.ruta.name + ")"


class RepositorioArchivo(RepositorioTrazas):
    """Lee un JSONL (una traza por línea): el archivo que emite
    el cliente Go de Metrín (metrin/internal/jev) cuando decide
    con jeva.cpp. Revalidación por mtime: trazas nuevas aparecen
    sin reiniciar el visor."""

    def __init__(self, ruta: Path):
        self.ruta = ruta
        self._trazas: list[dict] | None = None
        self._mtime: float | None = None

    def _cargar(self) -> list[dict]:
        try:
            m = self.ruta.stat().st_mtime
        except FileNotFoundError:
            return []
        if self._trazas is None or m != self._mtime:
            trazas: list[dict] = []
            with open(self.ruta, encoding="utf-8") as f:
                for linea in f:
                    if linea.strip():
                        trazas.append(json.loads(linea))
            self._trazas = trazas
            self._mtime = m
        return self._trazas

    def listar(self) -> list[dict]:
        return self._cargar()

    def obtener(self, traza_id: str) -> dict | None:
        for t in self._cargar():
            if t["id"] == traza_id:
                return t
        return None

    def origen(self) -> str:
        return "archivo (" + self.ruta.name + ")"


class RepositorioPostgres(RepositorioTrazas):
    """Lee las tablas traces + jev (trazas/esquema.sql, Fase 3/5)
    y reconstruye el JSON anidado del visor con un JOIN.
    psycopg se importa al conectar: el visor arranca sin él."""

    def __init__(self, dsn: str):
        self.dsn = dsn

    def _conectar(self):
        import psycopg  # noqa: PLC0415 — solo si TRAZAS_REPO=postgres

        return psycopg.connect(self.dsn)

    def _trazas(self, donde: str, args: tuple) -> list[dict]:
        with self._conectar() as cn, cn.cursor() as cur:
            cur.execute(
                """
                SELECT t.id, t.ts, t.modelo, t.origen, t.plantilla, t.ms,
                       t.tokens_entrada, t.estado, t.correcta,
                       j.pregunta_id, j.tipo, j.instrucciones, j.criterios, j.respuesta
                FROM traces t
                LEFT JOIN jev j ON j.trace_id = t.id
                """ + donde + "\n                ORDER BY t.ts DESC, t.id DESC, j.pregunta_id",
                args,
            )
            por_id: dict[str, dict] = {}
            orden: list[str] = []
            for (tid, ts, modelo, origen, plantilla, ms, tokens, estado,
                 correcta, pid, tipo, instrucciones, criterios, respuesta) in cur.fetchall():
                if tid not in por_id:
                    por_id[tid] = {
                        "id": tid,
                        "ts": ts.isoformat() if hasattr(ts, "isoformat") else str(ts),
                        "modelo": modelo or "",
                        "origen": origen or "",
                        "plantilla": plantilla,
                        "ms": ms or 0,
                        "tokens_entrada": tokens or 0,
                        "estado": estado,
                        "preguntas": [],
                        "correcta": correcta,
                    }
                    orden.append(tid)
                if pid is not None:
                    por_id[tid]["preguntas"].append({
                        "id": pid,
                        "tipo": tipo,
                        "instrucciones": instrucciones,
                        "criterios": criterios,
                        "respuesta": respuesta,
                    })
            return [por_id[i] for i in orden]

    def listar(self) -> list[dict]:
        return self._trazas("", ())

    def obtener(self, traza_id: str) -> dict | None:
        trazas = self._trazas("WHERE t.id = %s", (traza_id,))
        return trazas[0] if trazas else None

    def origen(self) -> str:
        return "postgres"


# ───────────────────────────── modelo ───────────────────────────────
def confianza_pregunta(p: dict) -> float | None:
    """Confianza de una pregunta; noul no la trae (la da el JEV)."""
    r = p.get("respuesta") or {}
    c = r.get("confianza")
    return c if isinstance(c, (int, float)) else None


def resumen(traza: dict) -> dict:
    """Traza recortada para la lista: lo que necesita el visor."""
    preguntas = traza.get("preguntas") or []
    confianzas = [c for c in (confianza_pregunta(p) for p in preguntas) if c is not None]
    return {
        "id": traza["id"],
        "ts": traza.get("ts", ""),
        "modelo": traza.get("modelo", ""),
        "origen": traza.get("origen", ""),
        "ms": traza.get("ms", 0),
        "tokens_entrada": traza.get("tokens_entrada", 0),
        "tipos": sorted({p["tipo"] for p in preguntas}),
        "n_preguntas": len(preguntas),
        "confianza": round(statistics.mean(confianzas), 3) if confianzas else None,
        "correcta": traza.get("correcta"),
        "titulo": _titulo(preguntas),
    }


def _titulo(preguntas: list[dict]) -> str:
    for p in preguntas:
        ins = p.get("instrucciones")
        if isinstance(ins, str) and ins.strip():
            return ins.strip()
        if isinstance(ins, dict) and ins.get("texto"):
            return str(ins["texto"]).strip()
    return "traza sin instrucciones"


def filtrar(trazas: list[dict], args) -> list[dict]:
    """Filtros de la lista: q, tipo, modelo, fallidas."""
    q = (args.get("q", "") or "").strip().lower()
    tipo = (args.get("tipo", "") or "").strip().lower()
    modelo = (args.get("modelo", "") or "").strip()
    solo_falladas = (args.get("fallidas", "") or "").strip().lower() in ("1", "true", "si", "sí")

    def _texto(t: dict) -> str:
        partes = [json.dumps(t.get("estado") or {}, ensure_ascii=False)]
        for p in t.get("preguntas") or []:
            partes.append(str(p.get("instrucciones") or ""))
            partes.append(json.dumps(p.get("criterios") or {}, ensure_ascii=False))
            partes.append(json.dumps(p.get("respuesta") or {}, ensure_ascii=False))
        return " ".join(partes).lower()

    out = []
    for t in trazas:
        if tipo and tipo not in {p["tipo"] for p in t.get("preguntas") or []}:
            continue
        if modelo and t.get("modelo") != modelo:
            continue
        if solo_falladas and t.get("correcta") is not False:
            continue
        if q and q not in _texto(t):
            continue
        out.append(t)
    return out


def estadisticas(trazas: list[dict]) -> dict:
    """Cifras del visor: se recalculan sobre la marcha (mock; el SQL
    del esquema Postgres dará lo mismo con GROUP BY)."""
    todas_p = [p for t in trazas for p in t.get("preguntas") or []]
    confianzas = [c for c in (confianza_pregunta(p) for p in todas_p) if c is not None]
    latencias = sorted(t.get("ms", 0) for t in trazas)
    evaluadas = [t for t in trazas if t.get("correcta") is not None]
    aciertos = sum(1 for t in evaluadas if t.get("correcta") is True)

    por_modelo: dict[str, int] = {}
    for t in trazas:
        por_modelo[t.get("modelo", "?")] = por_modelo.get(t.get("modelo", "?"), 0) + 1
    por_tipo: dict[str, int] = {}
    for p in todas_p:
        por_tipo[p["tipo"]] = por_tipo.get(p["tipo"], 0) + 1

    def pct(sorted_list: list, q: float) -> int:
        if not sorted_list:
            return 0
        i = min(len(sorted_list) - 1, int(round(q * (len(sorted_list) - 1))))
        return sorted_list[i]

    def cubetas(vals: list[float], bordes: list[float]) -> dict:
        etiquetas = [f"<{bordes[0]}"] + [f"{bordes[i]}–{bordes[i+1]}" for i in range(len(bordes) - 1)] + [f"≥{bordes[-1]}"]
        cuentas = [0] * len(etiquetas)
        for v in vals:
            i = 0
            while i < len(bordes) and v >= bordes[i]:
                i += 1
            cuentas[i] += 1
        return {"etiquetas": etiquetas, "valores": cuentas}

    return {
        "total": len(trazas),
        "por_modelo": por_modelo,
        "por_tipo": por_tipo,
        "confianza_media": round(statistics.mean(confianzas), 3) if confianzas else None,
        "confianza_min": round(min(confianzas), 3) if confianzas else None,
        "confianza_max": round(max(confianzas), 3) if confianzas else None,
        "latencia": {
            "media": round(statistics.mean(latencias)) if latencias else 0,
            "p50": pct(latencias, 0.5),
            "p95": pct(latencias, 0.95),
        },
        "aciertos": {
            "n": aciertos,
            "total": len(evaluadas),
            "pct": round(100 * aciertos / len(evaluadas), 1) if evaluadas else None,
        },
        "latencia_cubetas": cubetas([float(v) for v in latencias], [50, 100, 200, 500]),
        "confianza_cubetas": cubetas(confianzas, [0.4, 0.6, 0.8]),
    }


# ───────────────────────────── servicio ─────────────────────────────
def repo_desde_env() -> RepositorioTrazas:
    """Repositorio según TRAZAS_REPO:
      mock (defecto) · archivo[:ruta] (JSONL de metrin) · postgres
    El Postgres de Fase 3/5 entra por aquí sin cambiar el API."""
    # Solo el modo va en minúsculas: la ruta del archivo es
    # caso-sensible (p. ej. .../T/tmp... en macOS).
    valor = os.environ.get("TRAZAS_REPO", "mock").strip()
    modo, _, arg = valor.partition(":")
    modo = modo.lower()
    if modo == "archivo":
        ruta = arg.strip() or os.environ.get(
            "TRAZAS_ARCHIVO", str(RAIZ / "datos" / "trazas.jsonl"))
        return RepositorioArchivo(Path(ruta))
    if modo == "postgres":
        dsn = os.environ.get("TRAZAS_PG_DSN", "").strip()
        if not dsn:
            raise RuntimeError("TRAZAS_REPO=postgres necesita TRAZAS_PG_DSN")
        return RepositorioPostgres(dsn)
    return RepositorioMock(RAIZ / "mock" / "trazas.json")


def crear_app(repo: RepositorioTrazas | None = None) -> Flask:
    app = Flask(__name__, template_folder="templates")
    repo = repo or repo_desde_env()

    @app.get("/")
    def visor():
        return render_template("visor.html")

    @app.get("/api/salud")
    def salud():
        return jsonify({"ok": True, "origen": repo.origen()})

    @app.get("/api/trazas")
    def api_trazas():
        limite = request.args.get("limite", default=50, type=int)
        offset = request.args.get("offset", default=0, type=int)
        trazas = filtrar(repo.listar(), request.args)
        # Las más nuevas primero; el mock ya viene ordenado, pero el
        # Postgres hará ORDER BY ts DESC.
        trazas = sorted(trazas, key=lambda t: t.get("ts", ""), reverse=True)
        return jsonify({
            "total": len(trazas),
            "limite": max(1, min(limite, 200)),
            "offset": max(0, offset),
            "trazas": [resumen(t) for t in trazas[offset:offset + max(1, min(limite, 200))]],
        })

    @app.get("/api/trazas/<traza_id>")
    def api_traza(traza_id):
        t = repo.obtener(traza_id)
        if t is None:
            return jsonify({"error": "traza no encontrada"}), 404
        return jsonify(t)

    @app.get("/api/estadisticas")
    def api_estadisticas():
        return jsonify(estadisticas(filtrar(repo.listar(), request.args)))

    @app.get("/api/modelos")
    def api_modelos():
        modelos = sorted({t.get("modelo", "?") for t in repo.listar()})
        return jsonify({"modelos": modelos})

    return app


if __name__ == "__main__":
    app = crear_app()
    print(f"visor de trazas en http://{HOST}:{PUERTO}")
    app.run(host=HOST, port=PUERTO, debug=False)
