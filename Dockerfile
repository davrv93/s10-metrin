# Imagen de trabajo de s10-conocimiento: ingesta (s10kb.py, youtube.py), tutoriales
# (tutor.py), panel (admin.py) y Cortex (cortexboard). La misma imagen sirve a los
# servicios `admin` y `cortex` del compose; el código y los datos llegan por volumen.
FROM node:22-slim AS node

FROM python:3.13-slim

# HTTPS para apt: en redes con proxy que intercepta HTTP, los .deb llegan alterados (Hash Sum mismatch).
RUN sed -i 's#http://deb.debian.org#https://deb.debian.org#g' /etc/apt/sources.list.d/debian.sources

RUN apt-get -o Acquire::Retries=5 update && apt-get -o Acquire::Retries=5 install -y --no-install-recommends --fix-missing \
        poppler-utils tesseract-ocr tesseract-ocr-spa ffmpeg socat curl ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Node desde su imagen oficial (evita ~300 paquetes node-* de Debian)
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
    && ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx

# Cortex fijado a la versión que creó .cortex/
RUN npm install -g cortexboard@0.4.0 && npm cache clean --force

WORKDIR /app
COPY requirements.txt requirements-servicios.txt ./
RUN pip install --no-cache-dir -r requirements.txt -r requirements-servicios.txt

# Los scripts buscan .venv/bin/python y .venv/bin/yt-dlp: en la imagen apuntan al Python del sistema.
RUN mkdir -p /opt/venv-compat/bin \
    && ln -s "$(which python)" /opt/venv-compat/bin/python \
    && ln -s "$(which yt-dlp)" /opt/venv-compat/bin/yt-dlp

ENV EN_DOCKER=1 PYTHONUNBUFFERED=1 TZ=America/Lima
COPY docker/entrada.sh /usr/local/bin/entrada.sh
RUN chmod +x /usr/local/bin/entrada.sh
ENTRYPOINT ["entrada.sh"]
