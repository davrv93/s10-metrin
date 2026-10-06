#!/usr/bin/env bash
# Sube al EC2 (3.135.80.228) las fotos que la V2 de Metrín puede poner en un plan, a
# ~/s10-conocimiento/data/imagenes/, que el compose monta en /datos y Metrín sirve en GET /fotos/.
#
# Las fotos no están en git (data/* va en .gitignore) ni las siembra el CD (data/ está en .deploy-datos),
# así que en el EC2 «las fotos no se ven». Este script sube SOLO las referenciadas:
#   1. kb/procedimientos/**/*.yml   →  pasos[].fotos[].ruta_o_url («imagenes/<manual>/<hash>.png»)
#   2. kb/fragmentos*.jsonl          →  pasos[].fotos de cada fragmento (la V2 arma con ellas los procedimientos
#                                       implícitos de una sección del manual; V1 las usa en sus fuentes)
# Las URL (http…) no se suben: el navegador las pide a su sitio.
#
# Por qué todas las de (2) y no solo las de (1): cuando Metrín arranca y ya existe /datos/imagenes, el quality
# gate de la V2 comprueba que cada foto de un plan exista en disco. Una foto que falte deja de ser un aviso y
# pasa a ser un fallo del plan.
#
# Se ejecuta desde la copia de s10-conocimiento que TIENE data/imagenes (el clon de davrv93/s10-metrin en la Mac;
# AgenteS10 no las tiene). No borra nada en el servidor (sin --delete) y no toca otras carpetas.
#
#   herramientas/subir-imagenes-procedimientos.sh --listar    # solo la lista y el tamaño; no se conecta
#   herramientas/subir-imagenes-procedimientos.sh --simular   # rsync --dry-run contra el servidor
#   herramientas/subir-imagenes-procedimientos.sh             # sube y luego comprueba que no quede nada pendiente
#
# Variables: LLAVE (~/Downloads/OptimizaAgenteS10.pem), DESTINO (ubuntu@3.135.80.228),
#            DIR_REMOTO (s10-conocimiento/data, relativo al home), DATOS (<raíz>/data), KB (<raíz>/kb).
set -euo pipefail

RAIZ="$(cd "$(dirname "$0")/.." && pwd)"
LLAVE="${LLAVE:-$HOME/Downloads/OptimizaAgenteS10.pem}"
DESTINO="${DESTINO:-ubuntu@3.135.80.228}"
DIR_REMOTO="${DIR_REMOTO:-s10-conocimiento/data}"
DATOS="${DATOS:-$RAIZ/data}"
KB="${KB:-$RAIZ/kb}"
MODO="${1:-subir}"
case "$MODO" in
  subir | --listar | --simular) ;;
  *) echo "uso: $0 [--listar|--simular]" >&2; exit 2 ;;
esac

LISTA="$(mktemp)"
trap 'rm -f "$LISTA"' EXIT

# Lista de rutas relativas a data/ («imagenes/…»), sin repetir. Solo la biblioteca estándar de Python.
python3 - "$KB" > "$LISTA" <<'PY'
import glob, json, os, re, sys

kb = sys.argv[1]
valida = re.compile(r'^imagenes/[A-Za-z0-9._/-]+$')
rutas = set()
yml = glob.glob(os.path.join(kb, 'procedimientos', '**', '*.yml'), recursive=True)
if not yml:
    sys.exit(f'no hay procedimientos en {kb}/procedimientos')
for f in yml:
    for linea in open(f, encoding='utf-8'):
        m = re.match(r'\s*(?:-\s*)?ruta_o_url:\s*["\']?([^"\'\s#]+)', linea)
        if m:
            rutas.add(m.group(1))
for f in glob.glob(os.path.join(kb, 'fragmentos*.jsonl')):
    for linea in open(f, encoding='utf-8'):
        if '"pasos"' not in linea:
            continue
        for p in json.loads(linea).get('pasos') or []:
            for r in (p.get('fotos') or []) if isinstance(p, dict) else []:
                if isinstance(r, str):
                    rutas.add(r)
for r in sorted(rutas):
    if r.startswith(('http://', 'https://')):
        continue
    if not valida.match(r) or '..' in r:
        print(f'ruta rara, se salta: {r!r}', file=sys.stderr)
        continue
    print(r)
PY

N=$(wc -l < "$LISTA" | tr -d ' ')
FALTAN=0
BYTES=0
while IFS= read -r r; do
  if [ -f "$DATOS/$r" ]; then
    BYTES=$((BYTES + $(wc -c < "$DATOS/$r")))
  else
    FALTAN=$((FALTAN + 1))
    [ "$FALTAN" -le 5 ] && echo "falta en local: $DATOS/$r" >&2
  fi
done < "$LISTA"
echo "Fotos referenciadas: $N ($((BYTES / 1024 / 1024)) MB, $BYTES bytes) en $DATOS; faltan en local: $FALTAN"
if [ "$FALTAN" -gt 0 ]; then
  echo "Hay fotos que no están en $DATOS: ¿es la copia con data/imagenes? (DATOS=/ruta/a/data)" >&2
  exit 1
fi
if [ "$MODO" = "--listar" ]; then
  cat "$LISTA"
  exit 0
fi

[ -r "$LLAVE" ] || { echo "No puedo leer la llave $LLAVE (LLAVE=/ruta)" >&2; exit 1; }
SSH="ssh -i $LLAVE -o IdentitiesOnly=yes"
# -rltzO sin -p/-o/-g ni --delete, como el CD: los contenedores escriben como root dentro de la carpeta.
# --files-from implica rutas relativas: data/imagenes/<manual>/<hash>.png → ~/$DIR_REMOTO/imagenes/<manual>/<hash>.png
OPC=(-rltzO --files-from="$LISTA" -e "$SSH")
if [ "$MODO" = "--simular" ]; then
  rsync "${OPC[@]}" --dry-run --itemize-changes "$DATOS/" "$DESTINO:$DIR_REMOTO/" | sed -n '1,20p'
  echo "(simulación: no se subió nada)"
  exit 0
fi
rsync "${OPC[@]}" --info=stats1 "$DATOS/" "$DESTINO:$DIR_REMOTO/"
# Comprobación: una segunda pasada en seco no debe encontrar nada que copiar (un permiso de root que impidió
# escribir se vería aquí aunque rsync no hubiese fallado).
PENDIENTES=$(rsync "${OPC[@]}" --dry-run --itemize-changes "$DATOS/" "$DESTINO:$DIR_REMOTO/" | grep -c '^<f' || true)
if [ "$PENDIENTES" != 0 ]; then
  echo "Quedaron $PENDIENTES fotos sin subir (¿carpetas de root en ~/$DIR_REMOTO/imagenes?)" >&2
  exit 1
fi
echo "Listo: $N fotos en $DESTINO:~/$DIR_REMOTO/imagenes/. Metrín las sirve ya en /fotos/, sin reiniciar."
echo "Desde su próximo arranque la V2 además comprueba en disco las fotos de sus planes (RutaDatos se decide al arrancar)."
