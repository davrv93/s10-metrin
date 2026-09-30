#!/bin/sh
# Indexa la base y levanta el chat. index-kb trata el archivo como la base COMPLETA
# (borra lo que no aparece), así que los fragmentos de PDF/web y los de YouTube se
# juntan en uno antes. Es incremental: lo que no cambió se salta.
#
# Cuando el programador actualiza fuentes toca /kb/.actualizado: se reindexa y el
# chat se reinicia (corte de ~1 s) para cargar el índice nuevo.
marca() { stat -c %Y /kb/.actualizado 2>/dev/null || echo 0; }
trap 'kill $PID 2>/dev/null; exit 0' TERM INT
while true; do
  # Lo que el admin le enseñó (aprendidos.jsonl, en el volumen de datos) entra
  # también: si no, index-kb lo borraría del índice al reiniciar.
  cat /kb/fragmentos*.jsonl "${RAG_DATOS:-/srv/datos}/aprendidos.jsonl" > /tmp/kb.jsonl 2>/dev/null
  [ -s /tmp/kb.jsonl ] && rag index-kb /tmp/kb.jsonl || true
  VISTA=$(marca)
  rag serve --addr 0.0.0.0:${RAG_PUERTO:-4760} &
  PID=$!
  while kill -0 $PID 2>/dev/null; do
    sleep 20
    if [ "$(marca)" != "$VISTA" ]; then
      echo "base actualizada: recargando índice"
      kill $PID; wait $PID 2>/dev/null
      break
    fi
  done
done
