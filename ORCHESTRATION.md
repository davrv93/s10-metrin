# Orquestación — Pipeline de preguntas s10-conocimiento

## Servicios

| Servicio | Puerto | Estado | Notas |
|---|---|---|---|
| JEV (juez) | `127.0.0.1:8765` | vivo | `jev-style-0.8b-decision-v3`, backend MLX |
| Metrín chat | `127.0.0.1:8080` | vivo | `mlx_lm.server` + LoRA estilo |
| Cortex | `127.0.0.1:4748` | vivo | board + API |
| Admin | `127.0.0.1:4750` | vivo | panel |
| Metrín API | `127.0.0.1:4760` | vivo | compose |
| Trazas | `127.0.0.1:4761` | vivo | visor JEV |

## Variables relevantes

```bash
JEV_URL=http://host.docker.internal:8765
JEV_MODEL=jev-style-0.8b-decision-v3
LLM_PROVIDER=mlx
MLX_URL=http://127.0.0.1:8080
OLLAMA_MODEL=mlx-community/Qwen2.5-3B-Instruct-4bit
```

## Pipeline de preguntas

```bash
cd s10-conocimiento

# 1) Generar candidatas desde HTML
.venv/bin/python metrin/pipeline/preguntas.py generar --limite 0

# 2) Juzgar con Jev real
JEV_URL=http://127.0.0.1:8765 JEV_MODEL=jev-style-0.8b-decision-v3 \
  .venv/bin/python metrin/pipeline/preguntas.py juzgar

# 3) Publicar VB
.venv/bin/python metrin/pipeline/preguntas.py vb \
  --min-puntaje 0.80 --min-rubrica 0.80

# 4) Sincronizar a Cortex
.venv/bin/python metrin/pipeline/sync_cortex_semillas.py

# 5) Reindexar Cortex
npx -y cortexboard@0.4.0 reindex
```

## Estado actual

- **Fase:** juzgado real en background
- **Pendientes:** 2,553 candidatas
- **Siguiente:** vb → sync → reindex

## Archivos relevantes

- `metrin/pipeline/preguntas.py` — pipeline principal
- `metrin/pipeline/sync_cortex_semillas.py` — sync a `.cortex/tree/`
- `data/preguntas.db` — SQLite local
- `.cortex/tree/s10-conocimiento/semillas-generadas/` — destino Cortex
