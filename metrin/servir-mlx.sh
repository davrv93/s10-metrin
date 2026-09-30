#!/bin/sh
# Sirve Qwen2.5-3B + adaptador LoRA Metrín en http://127.0.0.1:8080 (GPU del Mac).
# Uso: ./servir-mlx.sh [--port 8080]
# Detener: pkill -f mlx_lm.server
set -e
PUERTO=8080
[ "$1" = "--port" ] && PUERTO="$2"
cd "$(dirname "$0")/.."
pkill -f "mlx_lm.server.*$PUERTO" 2>/dev/null || true
exec ./entrenamiento-metrin/.venv/bin/python -m mlx_lm.server \
  --model "mlx-community/Qwen2.5-3B-Instruct-4bit" \
  --adapter-path ./entrenamiento-metrin/artifacts/metrin-lora \
  --host 127.0.0.1 --port "$PUERTO"
