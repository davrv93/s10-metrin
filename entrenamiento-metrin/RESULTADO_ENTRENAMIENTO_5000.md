# Lote conversacional Metrín: 5.000 ejemplos

## Dataset

Generado de forma determinista por `generar_dataset_5000.py` (semilla `20260930`). Son ejemplos sintéticos basados en plantillas revisadas, no 5.000 respuestas redactadas individualmente ni datos técnicos del S10.

| Intención | Conversaciones |
|---|---:|
| Aclaraciones | 800 |
| Texto aportado: explicación y seguimiento | 800 |
| Correcciones | 600 |
| Reformulaciones | 600 |
| Seguimiento | 600 |
| Saludos | 400 |
| Límites de alcance | 400 |
| Cierres | 400 |
| Organización de pasos a partir de notas del usuario | 400 |

`data/metrin-5000/conversaciones.jsonl` conserva los 5.000 diálogos completos. Los archivos que consume MLX tienen 4.500/250/250 ejemplos y un intercambio por fila, porque el endpoint actual entrega al modelo la pregunta presente, no el historial del chat. Datos y hashes están en `data/metrin-5000/manifest.json`.

## Entrenamientos

Base: `mlx-community/Qwen2.5-3B-Instruct-4bit`; LoRA con 16 capas, batch 2, tasa `1e-5`, 2.250 iteraciones (una pasada aproximada por las 4.500 filas de entrenamiento), secuencia máxima 1.024 y checkpoints cada 500 pasos.

La primera corrida (`artifacts/metrin-lora-5000`) usó ejemplos de varios turnos sin historial en inferencia. Los prompts de control mostraron varias respuestas descontextualizadas; se conserva para auditoría, pero no se recomienda.

La segunda corrida (`artifacts/metrin-lora-5000-v2`) usó un intercambio por ejemplo. Pérdida final: entrenamiento `0,108`, validación `0,112` (25 batches de validación); memoria máxima `3,683 GB`. Log: `artifacts/train-5000-v2.log`.

## Evaluación y estado

Los mismos 24 prompts reservados se generaron en `artifacts/eval-5000-v2/comparacion.md`. El adaptador v2 tiende a respuestas breves, mantiene un tono profesional y rechaza varias solicitudes fuera del alcance de S10. Aún hay respuestas genéricas o poco ajustadas a la intención, por ejemplo ante algunos agradecimientos y seguimientos vagos. La pérdida baja no sustituye la evaluación humana de calidad.

**No está activado en Metrín ni reemplaza `artifacts/metrin-lora`.** La comparación humana anterior no se modificó. Completar una rúbrica sobre `artifacts/eval-5000-v2/comparacion.md` antes de cambiar `servir-mlx.sh` o el `.env`.
