#!/usr/bin/env python3
"""Evaluate the 5k adapter without touching the existing human-scored eval."""
import json
import sys
from pathlib import Path

from mlx_lm import generate, load
from mlx_lm.sample_utils import make_sampler

ROOT = Path(__file__).resolve().parent
MODEL = "mlx-community/Qwen2.5-3B-Instruct-4bit"
ADAPTER = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / "artifacts" / "metrin-lora-5000"
OUT = ROOT / "artifacts" / "eval-5000"
SYSTEM = (
    "Eres Metrín, asistente de Optimiza 360. Respondes en español neutro, "
    "con calidez, claridad y profesionalismo. No inventas datos ni afirmas "
    "acciones que no realizaste."
)


def main():
    base = [json.loads(line) for line in (ROOT / "eval" / "linea_base.jsonl").open(encoding="utf-8")]
    model, tokenizer = load(MODEL, adapter_path=str(ADAPTER))
    OUT.mkdir(parents=True, exist_ok=True)
    results = []
    for item in base:
        prompt = (
            f"<|im_start|>system\n{SYSTEM}<|im_end|>\n"
            f"<|im_start|>user\n{item['prompt']}<|im_end|>\n"
            "<|im_start|>assistant\n"
        )
        answer = generate(
            model, tokenizer, prompt=prompt, max_tokens=150,
            sampler=make_sampler(temp=0.0), verbose=False,
        ).strip()
        results.append({"id": item["id"], "prompt": item["prompt"], "respuesta": answer})
        print(f"[{len(results)}/{len(base)}] {item['prompt'][:50]}", flush=True)

    with (OUT / "adaptado.jsonl").open("w", encoding="utf-8") as f:
        for row in results:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")
    with (OUT / "comparacion.md").open("w", encoding="utf-8") as f:
        f.write("# Base vs LoRA 5k\n\nEvaluación cualitativa pendiente; conjunto de prompts reservado.\n\n")
        for before, after in zip(base, results):
            f.write(f"## {before['id'] + 1}. {before['prompt']}\n\n")
            f.write(f"**Base:** {before['respuesta']}\n\n**LoRA 5k:** {after['respuesta']}\n\n")
    mean = sum(len(row["respuesta"]) for row in results) / max(1, len(results))
    print(f"OK: {OUT}; longitud media LoRA={mean:.0f} caracteres")


if __name__ == "__main__":
    main()
