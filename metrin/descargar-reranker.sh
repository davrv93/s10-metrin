#!/usr/bin/env bash
# Baja el reranker de la V2 de Metrín a metrin/modelos/reranker/, la carpeta que monta el servicio «reranker» del
# docker-compose.yml (docs/V2-RAG-PROCEDURAL.md, «Despliegue»).
#
#   Modelo:   bge-reranker-v2-m3, cuantizado Q4_K_M en GGUF (438 MB)
#   Origen:   https://huggingface.co/gpustack/bge-reranker-v2-m3-GGUF (base BAAI/bge-reranker-v2-m3), revisión fija
#   Licencia: Apache-2.0 (ficha de Hugging Face, las dos)
#
# Comprueba el sha256 (el que publica Hugging Face para ese archivo). Si el archivo ya está y cuadra, no baja nada;
# si una descarga se corta, la retoma; si el sha256 no cuadra, borra lo bajado y sale con error. El modelo no se
# versiona (metrin/.gitignore) ni entra en la imagen de metrin (metrin/.dockerignore).
#
# Uso (desde cualquier carpeta):  sh metrin/descargar-reranker.sh
# Otra carpeta de destino:        RERANKER_DIR=/ruta sh metrin/descargar-reranker.sh
set -eu

REVISION="3093af03b1a635e67b084b1d8c03c5f5e020fd05"
NOMBRE="bge-reranker-v2-m3-Q4_K_M.gguf"
URL="https://huggingface.co/gpustack/bge-reranker-v2-m3-GGUF/resolve/$REVISION/$NOMBRE"
SHA256="e186a244ed455b4ab66ec64339ce7427a6ae13f5c0b5e544de96e50f0f8b3673"
TAMANO=438376864

DIR="${RERANKER_DIR:-$(cd "$(dirname "$0")" && pwd)/modelos/reranker}"
ARCHIVO="$DIR/$NOMBRE"
PARCIAL="$ARCHIVO.parcial"

sha() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | cut -d' ' -f1; else shasum -a 256 "$1" | cut -d' ' -f1; fi
}

mkdir -p "$DIR"
if [ -f "$ARCHIVO" ] && [ "$(sha "$ARCHIVO")" = "$SHA256" ]; then
  echo "Ya está: $ARCHIVO (sha256 correcto)."
  exit 0
fi

echo "Bajando $NOMBRE ($TAMANO bytes) a $DIR ..."
curl -fL --progress-bar --retry 3 --retry-delay 2 -C - -o "$PARCIAL" "$URL"
OBTENIDO="$(sha "$PARCIAL")"
if [ "$OBTENIDO" != "$SHA256" ]; then
  echo "El sha256 no cuadra: $OBTENIDO (se esperaba $SHA256). Se borra lo bajado." >&2
  rm -f "$PARCIAL"
  exit 1
fi
mv "$PARCIAL" "$ARCHIVO"
echo "Listo: $ARCHIVO (sha256 correcto)."
