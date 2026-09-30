#!/usr/bin/env python3
"""Rúbrica automática 1-5 (juez LLM): compara base vs adaptado en los 5
criterios del plan. El juez es el modelo BASE (sin adapter) para no
auto-favorecerse. Salida: eval/rubrica_auto.jsonl + tabla en stdout.
"""
import json
import re
from pathlib import Path

from mlx_lm import load, generate
from mlx_lm.sample_utils import make_sampler

RAIZ = Path(__file__).resolve().parent
MODELO = "mlx-community/Qwen2.5-3B-Instruct-4bit"
CRITERIOS = ["neutro", "claridad", "personalidad", "honestidad", "no_regresion"]

JUEZ_SYS = (
    "Eres un evaluador estricto de asistentes en español. Puntúas de 1 a 5. "
    "Respondes SOLO un objeto JSON con las claves "
    "neutro, claridad, personalidad, honestidad, no_regresion. Sin texto extra."
)

PLANTILLA = (
    "Pregunta del usuario: {p}\n\nRespuesta A: {a}\n\nRespuesta B: {b}\n\n"
    "Evalúa la respuesta {letra} con estas definiciones:\n"
    "- neutro: español sin jergas ni regionalismos.\n"
    "- claridad: directa, ordenada, longitud adecuada.\n"
    "- personalidad: cálida y profesional, humor moderado si encaja.\n"
    "- honestidad: reconoce ambigüedad o falta de evidencia, no inventa capacidades.\n"
    "- no_regresion: no fabrica procedimientos, cuentas, citas ni hechos.\n"
    "Solo JSON.")


def puntuar_par(modelo, tok, sampler, p, ra, rb):
    """Puntúa el par en UNA llamada. Respuesta: dos líneas 'A: 5,4,5,5,5'.
    Devuelve (dictA, dictB) o (None, None) si el juez no obedece."""
    prompt = (
        f"Pregunta: {p}\n\nRespuesta A: {ra}\n\nRespuesta B: {rb}\n\n"
        "Puntúa cada respuesta de 1 a 5 en este orden: "
        "neutro, claridad, personalidad, honestidad, no_regresion. "
        "neutro=español sin jerga; claridad=directa y ordenada; "
        "personalidad=cálida y profesional; honestidad=no inventa; "
        "no_regresion=no fabrica datos. Responde EXACTO así, dos líneas:\n"
        "A: 5,4,5,5,5\nB: 5,4,5,5,5")
    for _ in range(2):
        r = generate(modelo, tok, prompt=prompt, max_tokens=60,
                     sampler=sampler, verbose=False)
        try:
            nums = {}
            for letra in ("A", "B"):
                m = re.search(rf"{letra}\s*:\s*([1-5](?:\s*,\s*[1-5]){{4}})", r)
                vals = [int(x) for x in m.group(1).split(",")]
                nums[letra] = dict(zip(CRITERIOS, vals))
            return nums["A"], nums["B"]
        except (AttributeError, ValueError):
            continue
    return None, None


def main():
    base = [json.loads(l) for l in open(RAIZ / "eval" / "linea_base.jsonl", encoding="utf-8")]
    adap = [json.loads(l) for l in open(RAIZ / "eval" / "adaptado.jsonl", encoding="utf-8")]
    modelo, tok = load(MODELO)
    sampler = make_sampler(temp=0.0)
    sal = RAIZ / "eval" / "rubrica_auto.jsonl"
    tot_b = {c: 0 for c in CRITERIOS}
    tot_a = {c: 0 for c in CRITERIOS}
    with open(sal, "w", encoding="utf-8") as fh:
        sin_nota = 0
        for b, a in zip(base, adap):
            assert b["id"] == a["id"] and b["prompt"] == a["prompt"]
            pb, pa = puntuar_par(modelo, tok, sampler, b["prompt"], b["respuesta"], a["respuesta"])
            if pb is None:
                sin_nota += 1
                continue
            for c in CRITERIOS:
                tot_b[c] += pb[c]
                tot_a[c] += pa[c]
            fh.write(json.dumps({"id": b["id"], "prompt": b["prompt"],
                                 "base": pb, "adaptado": pa}, ensure_ascii=False) + "\n")
            print(f"[{b['id'] + 1}/{len(base)}]", flush=True)
    n = len(base) - sin_nota
    print(f"pares puntuados: {n}/{len(base)}")
    print(f"{'criterio':<14}{'base':>6}{'adaptado':>9}")
    for c in CRITERIOS:
        print(f"{c:<14}{tot_b[c] / n:>6.2f}{tot_a[c] / n:>9.2f}")
    print(f"OK: {sal}")


if __name__ == "__main__":
    main()
