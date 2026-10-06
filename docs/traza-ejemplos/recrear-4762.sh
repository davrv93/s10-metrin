#!/usr/bin/env bash
# Reconstruye y recrea el contenedor de PRUEBA de Metrín (127.0.0.1:4762) con la traza y la V2 activables por petición.
# No toca el contenedor de siempre (4760). Ejecutar desde s10-conocimiento/.
# Reranker de la V2 (búsqueda híbrida de fragmentos): llama-server --reranking en la Mac, :8091 (comando en
# metrin/eval/BUSQUEDA.md, «Reranker encendido»). Encendido por defecto; para apagarlo: RERANK_URL= ./recrear-4762.sh
# Clases que reordena (V2_RERANK_CLASES; BUSQUEDA.md §12): vacío = el defecto del código. Variantes medidas:
#   V2_RERANK_CLASES=fragmento ./recrear-4762.sh · V2_RERANK_CLASES=fragmento,procedimiento,concepto ./recrear-4762.sh
set -euo pipefail
RERANK_URL="${RERANK_URL-http://host.docker.internal:8091}"
RERANK_TOP_N="${RERANK_TOP_N:-20}"
RERANK_MAX_RUNAS="${RERANK_MAX_RUNAS:-800}"
RERANK_TIMEOUT_MS="${RERANK_TIMEOUT_MS:-3000}"
V2_RERANK_CLASES="${V2_RERANK_CLASES-}"
cd "$(dirname "$0")/../.."
docker build -q -t s10-metrin:traza metrin
docker rm -f metrin-traza-prueba >/dev/null 2>&1 || true
docker run -d --name metrin-traza-prueba -p 127.0.0.1:4762:4760 \
  -e METRIN_TRAZA=1 -e RAG_MAX_DISTANCIA=0.80 -e TZ=America/Lima \
  -e AGENT_VERSION=v1 -e AGENT_V2_ENABLED=true -e AGENT_V2_PERCENTAGE=0 -e DECISION_ENGINE=reglas \
  -e V2_TIPO_MODELO=/srv/modelos/clasificador-tipo-respuesta-candidato.json \
  -e V2_BUSQUEDA=hibrida -e RERANK_URL="$RERANK_URL" -e RERANK_TOP_N="$RERANK_TOP_N" \
  -e RERANK_MAX_RUNAS="$RERANK_MAX_RUNAS" -e RERANK_TIMEOUT_MS="$RERANK_TIMEOUT_MS" -e V2_RERANK_CLASES="$V2_RERANK_CLASES" \
  -e LLM_PROVIDER=mlx -e MLX_URL=http://host.docker.internal:8080 \
  -e OLLAMA_URL=http://host.docker.internal:11434 -e OLLAMA_MODEL=mlx-community/Qwen2.5-3B-Instruct-4bit \
  -e JEV_URL=http://host.docker.internal:8765 -e JEV_MODEL=jev-style-0.8b-decision-v3 \
  -e EMBED_PROVIDER=estatico -e EMBED_MODELO_ESTATICO=/srv/modelos/potion-es-int8.pjge -e RAG_DATOS=/srv/datos \
  --add-host host.docker.internal:host-gateway \
  -v metrin-traza-datos:/srv/datos -v "$PWD/kb:/kb:ro" -v "$PWD/data:/datos:ro" -v "$PWD/metrin/plantillas:/srv/plantillas:ro" \
  --entrypoint rag s10-metrin:traza serve --addr 0.0.0.0:4760 >/dev/null
for i in $(seq 1 30); do curl -sf http://127.0.0.1:4762/health >/dev/null && break; sleep 1; done
curl -s http://127.0.0.1:4762/health; echo
