#!/usr/bin/env python3
"""Línea base: 24 prompts inéditos contra el modelo base (sin adaptador).

Configuración fija y determinista (temp 0.0); la misma se usa con el adaptador
para comparar. Guarda entradas, salidas, config y versión en eval/.
"""
import json
import sys
from pathlib import Path

from mlx_lm import load, generate
from mlx_lm.sample_utils import make_sampler

RAIZ = Path(__file__).resolve().parent
MODELO = sys.argv[1] if len(sys.argv) > 1 else "mlx-community/Qwen2.5-3B-Instruct-4bit"

SYSTEM = (
    "Eres Metrín, asistente de Optimiza 360. Respondes en español neutro, "
    "con calidez, claridad y profesionalismo. No inventas datos ni afirmas "
    "acciones que no realizaste."
)

# Ninguno aparece en el dataset de entrenamiento.
PROMPTS = [
    "Hola, ¿quién eres?",
    "Buenas noches, ¿me atiendes?",
    "Gracias por la ayuda de hoy.",
    "No entendí nada de lo anterior.",
    "¿Me lo explicas con otras palabras?",
    "Eso que dijiste está incorrecto.",
    "No me sirve tu respuesta.",
    "Cuéntame sobre ese tema.",
    "¿Me ayudas con lo que te pedí?",
    "¿Cuánto dinero hay en mi cuenta?",
    "¿Qué dice exactamente el reglamento?",
    "Dame la dirección de la oficina.",
    "Escríbeme un cuento corto.",
    "¿Me ayudas a estudiar matemáticas?",
    "¿Qué sigue ahora?",
    "Dame un ejemplo de lo que dices.",
    "¿Qué más debería tener en cuenta?",
    "Explícame eso en detalle.",
    "¿Por dónde debería empezar?",
    "Hola de nuevo, tengo otra duda.",
    "Perdona, me confundí de tema.",
    "Vale, ¿y cómo continuamos?",
    "No sé cómo formular mi pregunta.",
    "¿Puedes ser más breve?",
]

CONFIG = {"modelo": MODELO, "adapter": None, "temp": 0.0, "max_tokens": 150}


def main():
    modelo, tokenizer = load(MODELO)
    print(f"mlx-lm {__import__('mlx_lm').__version__} modelo={MODELO}", flush=True)
    salida = RAIZ / "eval" / "linea_base.jsonl"
    with open(salida, "w", encoding="utf-8") as fh:
        for i, p in enumerate(PROMPTS):
            r = generate(
                modelo, tokenizer,
                prompt=f"<|im_start|>system\n{SYSTEM}<|im_end|>\n<|im_start|>user\n{p}<|im_end|>\n<|im_start|>assistant\n",
                max_tokens=CONFIG["max_tokens"], sampler=make_sampler(temp=CONFIG["temp"]),
                verbose=False,
            )
            fh.write(json.dumps({"id": i, "prompt": p, "respuesta": r.strip(),
                                 "config": CONFIG}, ensure_ascii=False) + "\n")
            print(f"[{i + 1}/{len(PROMPTS)}] {p[:40]}", flush=True)
    print(f"base OK: {salida}")


if __name__ == "__main__":
    main()
