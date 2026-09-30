# rag-go — RAG local en Go

Indexa **dos orígenes** —un repo en disco (la landing de ejemplo) y un bucket
S3 de **Garage**— en **un mismo índice vectorial** y responde preguntas en
español con un **LLM local** (Ollama). Todo en Go: no hay Python en ninguna
parte.

## Arquitectura

```
                 index-repo <ruta>                     index-s3 [--prefix docs/]
                        │                                       │
          ┌─────────────▼─────────────┐           ┌─────────────▼──────────────┐
          │ Repo (WalkDir)            │           │ S3 (aws-sdk-go-v2)         │
          │ salta node_modules/.git/  │           │ endpoint propio, path-style│
          │ dist · versión = sha256   │           │ región garage · versión=ETag│
          └─────────────┬─────────────┘           └─────────────┬──────────────┘
                        └──────────────┬────────────────────────┘
                                       ▼
                    indexar.Ejecutar  (incremental)
                    ├─ versión igual  → se salta (ni se descarga)
                    ├─ distinta       → borra trozos viejos, trocea, embebe
                    └─ ya no existe   → borra sus trozos
                                       │
            trocear: HTML sin script/style/etiquetas → 800 runas, solape 150
                                       │
          embed: «estatico» (model2vec potion int8, Go puro) | «ollama»
                                       │
                                       ▼
          almacen: chromem-go persistido en ./datos/chromem  (+ estado-*.json)
          id = sha1("repo:<ruta>:<n>") | sha1("s3:<key>:<n>")
          metadatos: source · type · ext · path · chunk · cita
                                       │
      ask "<pregunta>" [--source …]    │           serve → POST /ask, GET /
                        └──────────────▼──────────────┘
                 rag.Preguntar: top-k con filtro de metadatos
                 ├─ distancia mínima > RAG_MAX_DISTANCIA → «sin contexto»
                 ├─ prompt en español con etiquetas [repo:…#n] [s3:bucket/key#n]
                 ├─ Ollama /api/chat (temperatura 0.2, timeout 300 s)
                 └─ «No tengo contexto suficiente…» → datos/sin_respuesta.jsonl
```

Paquetes (`internal/`):

| paquete    | qué hace |
|------------|----------|
| `trocear`  | troceo por runas con solape, limpieza de HTML, extensiones válidas |
| `embed`    | `Estatico` (formato PJGE del Analista, copiado de `pjgfarma-analista/analista/interno/preguntas/modelo.go`) y `Ollama` (`/api/embed`) |
| `almacen`  | chromem-go + estado por documento; búsqueda con filtro |
| `indexar`  | orígenes `Repo` y `S3`, bucle incremental |
| `llm`      | cliente de `/api/chat` de Ollama |
| `rag`      | recuperación + prompt + registro de fallos |
| `servidor` | HTTP `:4760` (`POST /ask`, `GET /health`, página en `/`) |

### Por qué estas piezas

- **Embeddings estáticos** (por defecto): `potion-multilingual-128M` destilado
  a 32 420 piezas en español, int8, 256 dimensiones, 8,9 MB. Se carga en ~20 ms
  y embebe en microsegundos, sin GPU ni red. Es el mismo modelo del Analista;
  el código se **copió** (no se importa su paquete `internal`).
- **chromem-go**: base vectorial embebida, búsqueda exhaustiva por coseno,
  filtro `where` por igualdad de metadatos, persistencia en un `.gob` por
  trozo. Para miles de trozos sobra; para millones habría que cambiar.
- Una **colección por embebedor** (`rag-estatico-<huella>`,
  `rag-ollama-nomic-embed-text`): cambiar de `EMBED_PROVIDER` no mezcla
  espacios vectoriales; cada uno reindexa lo suyo.

## Instalación

Requisitos: Go 1.27, Ollama con `qwen2.5-coder:7b`, un Garage accesible.

```bash
cd rag-go
cp ../pjgfarma-analista/analista/modelos/potion-es-int8.pjge modelos/   # no se versiona (8,9 MB)
cp .env.example .env && chmod 600 .env                                   # y rellena las claves S3
go build -o rag ./cmd/rag
go test ./... && go vet ./...
```

Ollama:

```bash
ollama serve &                      # o: open -a Ollama
curl -s localhost:11434/api/tags    # debe responder
# solo si EMBED_PROVIDER=ollama:
ollama pull nomic-embed-text
```

> En este Mac el puerto 11434 lo tenía tomado una extensión de VS Code (no
> respondía), así que Ollama se levantó con `OLLAMA_HOST=127.0.0.1:11435
> ollama serve` y `.env` lleva `OLLAMA_URL=http://localhost:11435`.

## Comandos

```bash
./rag index-repo ejemplo-landing
./rag subir-s3 ejemplo-s3/docs --prefix docs/     # siembra el bucket de ejemplo
./rag index-s3 --prefix docs/
./rag index-kb ../s10-conocimiento/kb/fragmentos.jsonl

./rag ask "¿cómo está configurado el formulario de contacto?"
./rag ask "¿qué permite configurar el módulo de presupuestos?" --source s10-kb --solo-buscar
./rag ask "¿cuánto cuesta el plan Pro?" --source garage-s3
./rag ask "¿qué valida el formulario?" --source landing-repo --type codigo --k 4
./rag ask "precio anual" --ext .csv --json
./rag ask "precio del plan Pro" --solo-buscar     # solo recuperación, sin LLM

./rag serve                                        # http://127.0.0.1:4760
curl -s -X POST localhost:4760/ask -d '{"pregunta":"¿cuánto cuesta el Pro?","source":"garage-s3","k":8}'
```

Filtros de `ask`: `--source landing-repo|garage-s3|s10-kb`, `--type
doc|html|codigo|estilo|datos`, `--ext .md` (se combinan con Y).

La fuente `s10-kb` importa el JSONL producido por `s10-conocimiento`, conserva
manual, título, página y URL en los metadatos de recuperación y permite responder
con citas como `Guia de Usuario de S10 Presupuestos, p. 8`. El indexado es
incremental por ID de fragmento; si un fragmento desaparece del JSONL, se retira
del índice de esta fuente.

Reindexar es incremental: volver a lanzar `index-repo`/`index-s3` solo toca
lo que cambió y borra los trozos de documentos que ya no existen. En S3 el
borrado se limita al `--prefix` listado.

## Configuración (`.env` o entorno; el entorno manda)

| variable | defecto | |
|---|---|---|
| `EMBED_PROVIDER` | `estatico` | `estatico` \| `ollama` |
| `EMBED_MODELO_ESTATICO` | `./modelos/potion-es-int8.pjge` | |
| `OLLAMA_URL` | `http://localhost:11434` | LLM y embeddings |
| `OLLAMA_MODEL` | `qwen2.5-coder:7b` | también vale `qwen3-coder` |
| `OLLAMA_EMBED_MODEL` | `nomic-embed-text` | |
| `RAG_TEMPERATURA` / `RAG_TIMEOUT` | `0.2` / `300` | |
| `RAG_MAX_DISTANCIA` | `0.80` | distancia coseno (1 − similitud) del mejor trozo; por encima no se llama al LLM |
| `RAG_DATOS` | `./datos` | índice, estado y `sin_respuesta.jsonl` |
| `S3_ENDPOINT` `S3_REGION` `S3_BUCKET` | `http://localhost:4790` `garage` `rag-demo` | |
| `S3_ACCESS_KEY` `S3_SECRET_KEY` | — | solo en `.env` (gitignored, 600) |

El umbral 0.80 es para los embeddings estáticos: con ellos, las preguntas
sobre los documentos quedan en 0.55–0.75 y las ajenas (capital de Mongolia,
un mundial de fútbol) en 0.88–0.97. Con `nomic-embed-text` hay que
recalibrarlo con `ask --solo-buscar`.

## Garage

- API S3 en `http://localhost:4790` (contenedor `edisys_s3`, puerto 3900
  dentro). **Path-style obligatorio** (`UsePathStyle: true`): Garage solo
  sirve virtual-host si se le configura `root_domain`.
- **Región `garage`**: Garage comprueba la región de la firma SigV4 contra su
  `s3_region` (`garage` por defecto en este contenedor).
- El SDK v2 reciente añade checksums CRC32 por defecto; aquí se piden solo
  cuando la operación los exige (`RequestChecksumCalculationWhenRequired`)
  para no depender de la versión de Garage.
- Clave y bucket propios del RAG, sin tocar los de EDISYS:

```bash
docker exec edisys_s3 /garage key create rag-demo
docker exec edisys_s3 /garage bucket create rag-demo
docker exec edisys_s3 /garage bucket allow --read --write --owner rag-demo --key rag-demo
# la secret se vuelca a .env sin imprimirla:
docker exec edisys_s3 /garage key info --show-secret rag-demo | sed -n 's/^Secret key: *//p'
```

## Datos de ejemplo

- `ejemplo-landing/` — landing de **TurnoClaro** (SaaS ficticio de turnos):
  `index.html` (hero, funciones, planes **sin precios**, FAQ, formulario),
  `src/contacto.js` (validación y envío), `styles.css`, `README.md`,
  `package.json`.
- `ejemplo-s3/docs/` — lo que se sube a `rag-demo/docs/`: hoja de tarifas
  (el **precio del Pro** vive solo aquí), FAQ de soporte, política de
  privacidad, manual de instalación, notas de versión, SLA y un CSV de
  precios.

## Registro de fallos

Cada pregunta sin contexto suficiente —por distancia o porque el modelo lo
dice— se añade a `datos/sin_respuesta.jsonl` con fecha, motivo, distancia
mínima y fuentes consultadas: es la lista de huecos de la documentación.
