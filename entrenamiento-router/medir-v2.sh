#!/usr/bin/env bash
# Mide el router de casos dentro de la V2 (docs/TOC-RUTEO-METRIN.md §7): dos contenedores de PRUEBA con la misma
# imagen y sin reranker (como corre el EC2), uno sin router (4763) y otro con él (4764), y el benchmark completo
# (cmd/evalv2, 194 casos) contra cada uno. No toca el contenedor de siempre (4760) ni el de traza (4762).
# Ejecutar desde s10-conocimiento/. Informes en $SALIDA (por defecto entrenamiento-router/salida/v2): evalv2 escribe
# por defecto en metrin/eval/V1_VS_V2.md, que está versionado, así que aquí se le da otra ruta.
set -euo pipefail
cd "$(dirname "$0")/.."
SALIDA="${SALIDA:-entrenamiento-router/salida/v2}"
mkdir -p "$SALIDA"
docker build -q -t s10-metrin:router metrin >/dev/null

levantar() { # nombre puerto router
  docker rm -f "$1" >/dev/null 2>&1 || true
  docker run -d --name "$1" -p "127.0.0.1:$2:4760" \
    -e METRIN_TRAZA=1 -e RAG_MAX_DISTANCIA=0.80 -e TZ=America/Lima \
    -e AGENT_VERSION=v1 -e AGENT_V2_ENABLED=true -e AGENT_V2_PERCENTAGE=0 -e DECISION_ENGINE=reglas \
    -e V2_TIPO_MODELO=/srv/modelos/clasificador-tipo-respuesta-candidato.json \
    -e V2_BUSQUEDA=hibrida -e RERANK_URL= -e V2_ROUTER="$3" \
    -e LLM_PROVIDER=ollama -e OLLAMA_URL=http://host.docker.internal:11434 \
    -e EMBED_PROVIDER=estatico -e EMBED_MODELO_ESTATICO=/srv/modelos/potion-es-int8.pjge -e RAG_DATOS=/srv/datos \
    --add-host host.docker.internal:host-gateway \
    -v metrin-traza-datos:/srv/datos -v "$PWD/kb:/kb:ro" -v "$PWD/data:/datos:ro" \
    -v "$PWD/metrin/plantillas:/srv/plantillas:ro" \
    --entrypoint rag s10-metrin:router serve --addr 0.0.0.0:4760 >/dev/null
  for _ in $(seq 1 60); do curl -sf "http://127.0.0.1:$2/health" >/dev/null && return 0; sleep 1; done
  echo "$1 no respondió /health" >&2; docker logs --tail 30 "$1" >&2; return 1
}

levantar metrin-router-base 4763 ""
levantar metrin-router 4764 /srv/modelos/router/router-s10.cabeza.json
docker logs metrin-router 2>&1 | grep -i "router" || true

cd metrin
for x in "metrin-router-base 4763 base" "metrin-router 4764 router"; do
  set -- $x
  go run ./cmd/evalv2 --url "http://127.0.0.1:$2" --contenedor "$1" --versiones v2 \
    --md "../$SALIDA/$3.md" --json "../$SALIDA/$3"
done
