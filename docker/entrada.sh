#!/bin/sh
# /app/.venv es un volumen Linux vacío (tapa el .venv de macOS). Se le enlaza el
# Python del sistema para que admin.py y tutor.py encuentren python y yt-dlp.
set -e
[ -x /app/.venv/bin/python ] || ln -sfn /opt/venv-compat/bin /app/.venv/bin
exec "$@"
