#!/usr/bin/env python3
"""Arma los conjuntos del router de Metrín (pregunta → procedimiento), docs/TOC-RUTEO-METRIN.md.

Uso (desde entrenamiento-router/):
    .venv/bin/python construir_datos.py

Entradas:
- kb/procedimientos/<módulo>/*.yml (sin _reserva/): título, `preguntas` y `aliases` van siempre a entrenamiento.
- datos/parafrasis/*.jsonl: preguntas escritas por procedimiento y las de «_ninguno» (sin caso).
- metrin/eval/v2_oro.jsonl: SOLO para evaluar. Primer turno de cada caso: con `procedimiento_esperado` se espera ese
  procedimiento; sin él, «_ninguno».

- reales/ (fuera de git; extraer_reales.py + etiquetado): preguntas REALES de clientes de los grupos de WhatsApp,
  etiquetadas con procedimiento, «_ninguno» o «no_pregunta» (las dos últimas son «_ninguno» para el router). Solo las
  que se entienden sin contexto y, si nombran un procedimiento, con seguridad alta o media. Corte temporal por grupo: el
  60 % más antiguo entra en entrenamiento (y una quinta parte de él en calibración); el 40 % más reciente es
  reales/prueba_real.jsonl, que nunca se entrena.

Salidas en datos/: entrenamiento.jsonl, calibracion.jsonl, prueba.jsonl, clases.json y resumen.json; y, si hay
etiquetas reales, reales/prueba_real.jsonl (fuera de git).

Fuga: se descarta toda pregunta de entrenamiento o calibración que, normalizada, sea igual a un turno de la prueba o
comparta con él ≥ 80 % de sus palabras (Jaccard). Es la misma idea que la regla 9 de metrin/eval/construir_v2_oro.py.
"""
from __future__ import annotations

import glob
import json
import random
import re
import unicodedata
from collections import Counter, defaultdict
from pathlib import Path

import yaml

AQUI = Path(__file__).resolve().parent
RAIZ = AQUI.parent
DATOS = AQUI / "datos"
NINGUNO = "_ninguno"
SEMILLA = 20261007
FRACCION_CALIBRACION = 0.15
JACCARD_FUGA = 0.8
REALES = AQUI / "reales"
CORTE_REALES = 0.6          # fracción más antigua de cada grupo que se entrena
CALIBRACION_REALES = 0.2    # de esa parte, cuánto va a calibración


def normalizar(texto: str) -> str:
    t = unicodedata.normalize("NFKD", texto.lower())
    t = "".join(c for c in t if not unicodedata.combining(c))
    t = re.sub(r"[^a-z0-9ñ ]+", " ", t)
    return re.sub(r"\s+", " ", t).strip()


def palabras(texto: str) -> set[str]:
    return set(normalizar(texto).split())


def jaccard(a: set[str], b: set[str]) -> float:
    return len(a & b) / len(a | b) if a and b else 0.0


def cargar_procedimientos() -> dict[str, dict]:
    procs = {}
    for ruta in sorted(glob.glob(str(RAIZ / "kb/procedimientos/*/*.yml"))):
        if "/_reserva/" in ruta:
            continue
        d = yaml.safe_load(open(ruta, encoding="utf-8"))
        pid = Path(ruta).stem
        procs[pid] = {"modulo": pid.split(".", 1)[0], "titulo": d.get("titulo", ""), "yaml": d}
    return procs


def textos_yaml(p: dict) -> list[str]:
    d = p["yaml"]
    out = [p["titulo"]] + list(d.get("preguntas") or []) + list(d.get("aliases") or [])
    return [t for t in out if isinstance(t, str) and t.strip()]


def cargar_prueba(procs: dict) -> list[dict]:
    prueba = []
    for linea in open(RAIZ / "metrin/eval/v2_oro.jsonl", encoding="utf-8"):
        c = json.loads(linea)
        esperado = c.get("procedimiento_esperado") or NINGUNO
        if esperado != NINGUNO and esperado not in procs:
            raise SystemExit(f"{c['id']}: procedimiento {esperado} no existe")
        prueba.append({"texto": c["turnos"][0], "clase": esperado, "id": c["id"], "categoria": c["categoria"]})
    return prueba


def cargar_reales(clases: list[str]) -> tuple[list[dict], list[dict]]:
    """(para entrenar, para probar) con corte temporal por grupo. Vacío si no hay etiquetas."""
    etiquetas = {}
    for ruta in sorted(REALES.glob("etiquetas*.jsonl")) + sorted(REALES.glob("tickets/etiquetas*.jsonl")):
        for linea in open(ruta, encoding="utf-8"):
            if linea.strip():
                x = json.loads(linea)
                etiquetas[x["id"]] = x
    if not etiquetas:
        return [], []
    por_grupo = defaultdict(list)
    lineas = list(open(REALES / "candidatas.jsonl", encoding="utf-8"))
    if (REALES / "tickets/candidatas.jsonl").exists():  # tickets de Zendesk (extraer_tickets.py), grupo «zendesk»
        lineas += list(open(REALES / "tickets/candidatas.jsonl", encoding="utf-8"))
    for linea in lineas:
        c = json.loads(linea)
        e = etiquetas.get(c["id"])
        if not e:
            continue
        clase = e["etiqueta"]
        # «no_pregunta» («su apoyo por favor», pedidos de accesos) entra siempre como «_ninguno»: el router debe
        # abstenerse ante esos mensajes tal como llegan. A las preguntas se les exige entenderse sin el contexto.
        if clase == "no_pregunta":
            clase = NINGUNO
        elif not e.get("pregunta_autonoma"):
            continue
        if clase not in clases:
            continue
        if clase != NINGUNO and e.get("seguridad") not in ("alta", "media"):
            continue
        por_grupo[c["grupo"]].append({"texto": c["texto"], "clase": clase, "origen": "real", "id": c["id"],
                                      "grupo": c["grupo"], "t": c["t"], "intencion": e.get("intencion"),
                                      "categoria": e.get("intencion"), "etiqueta_original": e["etiqueta"]})
    entren, prueba = [], []
    for filas in por_grupo.values():
        filas.sort(key=lambda x: x["t"])
        k = int(len(filas) * CORTE_REALES)
        entren += filas[:k]
        prueba += filas[k:]
    # sin la misma frase a los dos lados del corte («su apoyo por favor» se repite mucho)
    en_prueba = {normalizar(x["texto"]) for x in prueba}
    entren = [x for x in entren if normalizar(x["texto"]) not in en_prueba]
    return entren, prueba


def main():
    procs = cargar_procedimientos()
    clases = sorted(procs) + [NINGUNO]
    prueba = cargar_prueba(procs)
    prueba_norm = {normalizar(x["texto"]) for x in prueba}
    prueba_pal = [palabras(x["texto"]) for x in prueba]

    def fuga(texto: str) -> bool:
        if normalizar(texto) in prueba_norm:
            return True
        p = palabras(texto)
        return any(jaccard(p, q) >= JACCARD_FUGA for q in prueba_pal)

    fijos, generados = [], []
    for pid, p in procs.items():
        for t in textos_yaml(p):
            fijos.append({"texto": t, "clase": pid, "origen": "yaml"})
    errores = Counter()
    for ruta in sorted((DATOS / "parafrasis").glob("*.jsonl")):
        for n, linea in enumerate(open(ruta, encoding="utf-8"), 1):
            linea = linea.strip()
            if not linea:
                continue
            try:
                x = json.loads(linea)
            except json.JSONDecodeError:
                errores[f"{ruta.name}: JSON inválido"] += 1
                continue
            clase = x.get("procedimiento")
            if clase not in clases or not (x.get("texto") or "").strip():
                errores[f"{ruta.name}: clase desconocida {clase!r}"] += 1
                continue
            generados.append({"texto": x["texto"].strip(), "clase": clase, "origen": ruta.stem,
                              "intencion": x.get("intencion")})

    # sin duplicados (normalizados) y sin fuga
    vistos, descartes = set(), Counter()
    def filtrar(filas):
        out = []
        for x in filas:
            k = normalizar(x["texto"])
            if not k or k in vistos:
                descartes["duplicado"] += 1
                continue
            if fuga(x["texto"]):
                descartes["fuga_prueba"] += 1
                continue
            vistos.add(k)
            out.append(x)
        return out
    fijos, generados = filtrar(fijos), filtrar(generados)

    reales_e, reales_p = cargar_reales(clases)
    reales_e = filtrar(reales_e)

    rnd = random.Random(SEMILLA)
    por_clase = defaultdict(list)
    for x in generados:
        por_clase[x["clase"]].append(x)
    entrenamiento, calibracion = list(fijos), []
    for clase in clases:
        filas = por_clase[clase]
        rnd.shuffle(filas)
        k = round(len(filas) * FRACCION_CALIBRACION)
        calibracion += filas[:k]
        entrenamiento += filas[k:]
    rnd.shuffle(reales_e)
    k = round(len(reales_e) * CALIBRACION_REALES)
    calibracion += reales_e[:k]
    entrenamiento += reales_e[k:]
    if reales_p:
        with open(REALES / "prueba_real.jsonl", "w", encoding="utf-8") as f:
            for x in reales_p:
                f.write(json.dumps(x, ensure_ascii=False) + "\n")

    DATOS.mkdir(exist_ok=True)
    for nombre, filas in (("entrenamiento", entrenamiento), ("calibracion", calibracion), ("prueba", prueba)):
        with open(DATOS / f"{nombre}.jsonl", "w", encoding="utf-8") as f:
            for x in filas:
                f.write(json.dumps(x, ensure_ascii=False) + "\n")
    modulos = {pid: p["modulo"] for pid, p in procs.items()}
    (DATOS / "clases.json").write_text(json.dumps(
        {"clases": clases, "modulo": modulos, "titulo": {pid: p["titulo"] for pid, p in procs.items()}},
        ensure_ascii=False, indent=1), encoding="utf-8")

    cuenta = Counter(x["clase"] for x in entrenamiento)
    resumen = {
        "clases": len(clases),
        "entrenamiento": len(entrenamiento),
        "calibracion": len(calibracion),
        "prueba": len(prueba),
        "prueba_con_caso": sum(x["clase"] != NINGUNO for x in prueba),
        "entrenamiento_min_por_clase": min(cuenta[c] for c in clases),
        "entrenamiento_ninguno": cuenta[NINGUNO],
        "descartes": dict(descartes),
        "errores": dict(errores),
        "clases_sin_parafrasis": [c for c in clases if not por_clase[c]],
        "reales_entrenamiento": len(reales_e) - k,
        "reales_calibracion": k,
        "reales_prueba": len(reales_p),
        "reales_prueba_con_caso": sum(x["clase"] != NINGUNO for x in reales_p),
    }
    (DATOS / "resumen.json").write_text(json.dumps(resumen, ensure_ascii=False, indent=1), encoding="utf-8")
    print(json.dumps(resumen, ensure_ascii=False, indent=1))


if __name__ == "__main__":
    main()
