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

# 6) Preguntas validadas → Metrín (tarjetas con su sección y capturas + menú)
.venv/bin/python semillas_a_kb.py
.venv/bin/python herramientas/catalogo_metrin.py
touch kb/.actualizado    # Metrín reindexa solo
```

### Cómo responde Metrín con las preguntas validadas

- `semillas_a_kb.py` ubica cada semilla con VB en su sección del HTML oficial
  (url + índice del `nodo_id`; si falta la url, por manual y sección) y escribe
  `kb/fragmentos_semillas.jsonl`: se busca por la pregunta y devuelve el texto y
  los `pasos` con fotos de esa sección. Las semillas sin sección se informan y
  no se publican. Un nodo de Cortex con `status: deprecated` o `archived` sale.
- **Elegida en el menú** (`/ask` con `semilla`): no pasa por el clasificador; la
  sección de esa semilla es el único contexto y el tipo de consulta es el que
  le dio el pipeline. `orquestacion.clasificador = semilla_validada`.
- **Escrita**: pasa por el clasificador de intención como siempre. Si la mejor
  evidencia es una tarjeta de semilla a menos de `RAG_DISTANCIA_SEMILLA`
  (0.15, sin calibrar), la respuesta adopta su tipo de consulta y
  `orquestacion.semilla` dice cuál fue.
- En ambos casos las capturas salen de los `pasos` de la sección y el servidor
  las coloca debajo del paso de la respuesta que ilustran.

## Estado actual

- **Fase:** juzgado real en background
- **Pendientes:** 2,553 candidatas
- **Siguiente:** vb → sync → reindex

## Archivos relevantes

- `metrin/pipeline/preguntas.py` — pipeline principal
- `metrin/pipeline/sync_cortex_semillas.py` — sync a `.cortex/tree/`
- `data/preguntas.db` — SQLite local
- `.cortex/tree/s10-conocimiento/semillas-generadas/` — destino Cortex
