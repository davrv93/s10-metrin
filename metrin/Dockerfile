# Metrín: RAG en Go (chat en :4760). Al arrancar indexa, de forma incremental,
# la base de s10-conocimiento (kb/fragmentos*.jsonl) montada en /kb.
FROM golang:alpine AS build
ENV GOTOOLCHAIN=auto
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /rag ./cmd/rag

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata poppler-utils
WORKDIR /srv
COPY --from=build /rag /usr/local/bin/rag
COPY modelos ./modelos
COPY docker-entrada.sh /usr/local/bin/entrada.sh
RUN chmod +x /usr/local/bin/entrada.sh
ENV RAG_DATOS=/srv/datos EMBED_PROVIDER=estatico EMBED_MODELO_ESTATICO=/srv/modelos/potion-es-int8.pjge TZ=America/Lima
EXPOSE 4760
ENTRYPOINT ["entrada.sh"]
