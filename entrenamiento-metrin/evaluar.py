#!/usr/bin/env python3
"""Evalúa el adaptador con los mismos 24 prompts de la línea base.

Uso: evaluar.py [modelo] [adapter-path]
Guarda eval/adaptado.jsonl y eval/comparacion.md (base vs adaptado lado a lado
para rúbrica humana 1-5 del plan).
"""
import json
import sys
from pathlib import Path

from mlx_lm import load, generate
from mlx_lm.sample_utils import make_sampler

RAIZ = Path(__file__).resolve().parent
MODELO = sys.argv[1] if len(sys.argv) > 1 else "mlx-community/Qwen2.5-3B-Instruct-4bit"
ADAPTER = sys.argv[2] if len(sys.argv) > 2 else str(RAIZ / "artifacts" / "metrin-lora")

SYSTEM = (
    "Eres Metrín, asistente de Optimiza 360. Respondes en español neutro, "
    "con calidez, claridad y profesionalismo. No inventas datos ni afirmas "
    "acciones que no realizaste."
)


def main():
    base = [json.loads(l) for l in open(RAIZ / "eval" / "linea_base.jsonl", encoding="utf-8")]
    modelo, tokenizer = load(MODELO, adapter_path=ADAPTER)
    print(f"adapter OK: {ADAPTER}", flush=True)
    adaptadas = []
    for r in base:
        a = generate(
            modelo, tokenizer,
            prompt=f"<|im_start|>system\n{SYSTEM}<|im_end|>\n<|im_start|>user\n{r['prompt']}<|im_end|>\n<|im_start|>assistant\n",
            max_tokens=150, sampler=make_sampler(temp=0.0), verbose=False,
        ).strip()
        adaptadas.append(a)
        print(f"[{r['id'] + 1}/{len(base)}] {r['prompt'][:40]}", flush=True)
    with open(RAIZ / "eval" / "adaptado.jsonl", "w", encoding="utf-8") as fh:
        for r, a in zip(base, adaptadas):
            fh.write(json.dumps({"id": r["id"], "prompt": r["prompt"],
                                 "respuesta": a}, ensure_ascii=False) + "\n")
    with open(RAIZ / "eval" / "comparacion.md", "w", encoding="utf-8") as fh:
        fh.write("# Base vs adaptado (rúbrica humana 1-5: neutro, claridad, "
                 "personalidad, honestidad, no regresión)\n\n")
        for r, a in zip(base, adaptadas):
            fh.write(f"## {r['id']}. {r['prompt']}\n\n**Base:** {r['respuesta']}\n\n"
                     f"**Adaptado:** {a}\n\n")
    lb = sum(len(r["respuesta"]) for r in base) / len(base)
    la = sum(len(a) for a in adaptadas) / len(base)
    print(f"largo medio base={lb:.0f} adaptado={la:.0f} (ojo: más verboso = regresión)")
    print("OK: eval/adaptado.jsonl eval/comparacion.md")


if __name__ == "__main__":
    main()
