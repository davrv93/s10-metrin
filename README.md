# s10-conocimiento

Base de conocimiento a partir de los manuales de https://documentacion.s10peru.com.

Los manuales están tras el login de miembros del sitio (Simple Membership). El
script entra con una cuenta de miembro y baja solo lo que esa sesión ve desde
las páginas del sitio. No descarga archivos por URL directa ni desde el sitemap
de adjuntos.

## Metrín

El plan del piloto de ajuste de estilo LoRA en Apple Silicon está en
[`docs/PLAN_LORA_METRIN_MLX.md`](docs/PLAN_LORA_METRIN_MLX.md). El ajuste no
sustituye la recuperación RAG ni autoriza aprendizaje automático de conversaciones.

## Uso

```bash
python3 -m venv .venv && .venv/bin/pip install requests beautifulsoup4
cp .env.example .env          # completar S10_USUARIO y S10_CLAVE (cuenta de Arturo)
.venv/bin/python s10kb.py todo            # descubre, rastrea, baja hasta 200 PDF e indexa
.venv/bin/python s10kb.py descargar       # siguiente cola de 200
.venv/bin/python s10kb.py importar        # suma los PDF dejados a mano en entrada/
.venv/bin/python s10kb.py ocr             # PDF escaneados -> texto (Vision de macOS, sin instalar nada)
.venv/bin/python s10kb.py indexar         # rehace kb/ con todo lo bajado
```

Requiere `pdftotext` (poppler): `brew install poppler`.

Para reproducir todo en otra máquina: [`REPRODUCCION.md`](REPRODUCCION.md)
(dependencias exactas en [`requirements.txt`](requirements.txt), ya verificadas
en venv limpio).

| Paso | Salida |
|---|---|
| `descubrir` | `data/urls.json` (páginas y posts del sitemap) |
| `rastrear` | `data/paginas/*.md`, `data/paginas.jsonl`, `data/enlaces.jsonl` |
| `descargar --lote N` | `data/pdf/*.pdf`, `data/manifiesto.jsonl` (sha256, fuente, página) |
| `indexar` | `kb/fragmentos.jsonl` (trozos de ~1500 caracteres con manual, página y fuente), `kb/documentos.json` |

Todo es reanudable: cada paso salta lo ya hecho. Una petición por segundo por
defecto (`S10_PAUSA`). `--sin-login` rastrea solo lo público, para probar.

## YouTube: tutoriales del canal oficial

El canal **Marketing S10** (`UCL1xwXCVt9ZVpSrgyemP50A`) publica los tutoriales
por módulo que S10 hace para sus clientes. `youtube.py` baja SOLO el audio, lo
transcribe en local (faster-whisper, sin enviar nada a terceros) y arma
fragmentos con minuto de inicio y enlace con timestamp:

```bash
.venv/bin/pip install yt-dlp faster-whisper "av==18.1.0"   # av>=19 rompe faster-whisper
.venv/bin/python youtube.py todo               # descubrir -> audio -> transcribir -> indexar
.venv/bin/python youtube.py transcribir --modelo small
```

| Paso | Salida |
|---|---|
| `descubrir` | `data/yt/videos.txt` (canal + 12 playlists, deduplicado) |
| `audio` | `data/yt/audio/{id}.m4a` (`bajados.txt` evita repetir) |
| `transcribir` | `data/yt/transcripciones/{id}.json` (segmentos con tiempos) y `.txt` |
| `indexar` | `kb/fragmentos_yt.jsonl`, `kb/videos.json` |

Estado (2026-09-30): 38 videos, 23 con contenido hablado → 35 fragmentos. Los 15
restantes son promos con música (sin habla), quedan marcados con 0 segmentos.
La KB final para búsqueda une `kb/fragmentos.jsonl` (web/PDF) y
`kb/fragmentos_yt.jsonl` (videos).

## Cortex y preguntas frecuentes (`cortex_a_kb.py`)

```bash
python3 cortex_a_kb.py   # tras editar .cortex/ o el banco data/preguntas/
```

| Salida | Qué es |
|---|---|
| `kb/fragmentos_cortex.jsonl` | Un fragmento por nodo/ítem de `.cortex/` (Metrín lo trocea en 800 caracteres) |
| `kb/fragmentos_faq.jsonl` | Una tarjeta por pregunta del banco: se **busca** por la pregunta y sus escrituras con errores (campo `busqueda`, lo único que se embebe) y **devuelve** el pasaje del nodo de Cortex que la responde (`texto`, ≤ 1200 caracteres, sin trocear) |

Las tarjetas existen porque el embebedor estático no encontraba el trozo correcto
cuando la respuesta es una línea dentro de un nodo largo (p. ej. «¿Cómo llego a
adicionar correctivo…?»). Llevan la misma cita que su nodo, así que la fuente no
se repite en la respuesta.

## Pantallas del ERP (`ui.py`)

Del video a un esquema de formulario. Los tutoriales de S10 muestran las
pantallas del ERP, así que sirven de fuente sin depender del login del portal:

```bash
.venv/bin/python ui.py capturas 17W0yKI8iew   # frames del video (ffmpeg, deduplicados)
.venv/bin/python ui.py ocr                    # tesseract spa: palabras con caja y bloque/línea
.venv/bin/python ui.py vision                 # gemma3:4b (Ollama): qué campo es y de qué tipo
.venv/bin/python ui.py esquema                # fusiona OCR + VLM -> esquema revisable
.venv/bin/python ui.py pantallas              # agrupa capturas en pantallas distintas
.venv/bin/python ui.py render --esquema data/ui/esquema/<x>.json   # HTML interactivo
```

| Salida | Qué es |
|---|---|
| `data/ui/capturas/` | frames por video (uno por segundo, sin repetidos) |
| `data/ui/ocr/` | palabras con caja, confianza y jerarquía de tesseract |
| `data/ui/vision/` | campos que ve el VLM (JSON forzado con `format`, no texto libre) |
| `data/ui/esquema/` | fusión: etiqueta, tipo, valor y caja por campo |
| `data/ui/pantallas.json` | cuántas pantallas distintas hay y qué capturas las forman |
| `render/*.html` | formulario sobre la captura: arrastrar, corregir y exportar |

**Por qué dos motores.** El OCR da la cadena exacta y las coordenadas; el VLM
entiende qué es un campo y de qué tipo. Ninguno alcanza solo: la geometría pura
confunde paneles y celdas de grilla con campos, y el VLM confunde letras
(`UIT`→`UT`, `Fijos`→`Fluidos`) y lee el calendario del date-picker como campos.
La fusión usa el OCR para las cadenas y las cajas, y el VLM para la semántica.

**El esquema es un borrador.** Se cierra abriendo el HTML, arrastrando cada campo
a su sitio, corrigiendo el tipo y exportando el JSON: quien ve la pantalla
corrige en segundos lo que las heurísticas no aciertan. Medido en un tutorial de
132 s: 60 capturas → **2 pantallas dominantes** (el resto, frames transitorios).

## Reportes SQL dinámicos (`reportes.py`)

Preguntas en lenguaje natural → reportes, **sin que la IA escriba SQL**: elige y
rellena una plantilla del catálogo aprobado (`reportes/plantillas.json`), valida
con sqlglot (solo SELECT) y ejecuta con sesión read-only.

```bash
.venv/bin/python reportes.py sembrar     # BD demo con los datos citados del video (UIT 4300/4400/5150)
.venv/bin/python reportes.py catalogo
.venv/bin/python reportes.py preguntar "¿Qué valor tiene la UIT 2024?"
.venv/bin/python reportes.py preguntar "costo de planilla de junio 2024" --bd postgresql://lectura:***@host/db
```

Fases: recuperar plantilla (léxico) → resolver parámetros (fechas/periodo/categoría) →
validar (sqlglot: SELECT único; DDL/DML rechazado) → ejecutar (500 filas máx,
timeout 15 s en PostgreSQL) → tabla con el SQL ejecutado. Contra la BD real de S10
solo falta lo que debe dar Arturo: cadena de conexión y **usuario de solo lectura**.

## Tutoriales automáticos (`tutor.py`)

Genera guías paso a paso a partir de la base de conocimiento y las publica en
`tutoriales/*.md` y en Cortex, bajo la rama `tutoriales/`.

```bash
.venv/bin/python tutor.py preguntar "¿Cómo creo un presupuesto de obra?"
.venv/bin/python tutor.py auto --max 8     # propone tareas desde el árbol de Cortex y genera las que falten
.venv/bin/python tutor.py listar
```

Por cada tutorial: **planificar** (pasos y consultas) → **recuperar** (BM25 sobre
`kb/fragmentos*.jsonl`, incluidos los de YouTube, más búsqueda en Cortex) →
**evaluar** (si la evidencia de un paso no alcanza, una ronda extra de búsqueda) →
**generar** (Markdown con citas `[F#]`/`[N#]`, pasos sin evidencia marcados como
"No documentado" y aviso en los que dependen de copias de terceros).

`auto` es idempotente: rehace los tutoriales cuya base cambió (entraron fuentes
nuevas) y agrega tareas nuevas. Programarlo, por ejemplo cada noche:

```cron
0 3 * * * cd /ruta/s10-conocimiento && .venv/bin/python tutor.py auto --max 5 >> tutoriales/auto.log 2>&1
```

Motor: SDK de Anthropic (`claude-opus-5`) si hay `ANTHROPIC_API_KEY`; si no, el
CLI de Claude Code (`claude -p`). Requiere el tablero de Cortex encendido
(`npx cortexboard start --port 4748`) para buscar nodos y publicar; sin él,
genera igual el Markdown.

## Panel de administración (`admin.py`)

```bash
.venv/bin/python admin.py        # http://127.0.0.1:4750  (usuario y clave: ADMIN_USUARIO / ADMIN_CLAVE en .env)
```

- **Resumen**: cifras de la base y accesos.
- **Cortex**: estado del tablero y botón para entrar con sesión iniciada (genera un enlace de `cortexboard login` de 10 minutos por clic).
- **Fuentes**: PDF (con su texto extraído, OCR y nivel de confianza editable), páginas del portal, videos de YouTube con su transcripción, y documentos internos (Metrín, catálogo).
- **Agregar fuentes**: subir PDF, añadir un video de YouTube o una página web pública, o rastrear el portal de miembros con la cuenta de `.env`. Todo corre en segundo plano y reindexa al terminar.
- **Tutoriales**: pedir uno o correr la generación automática, y leerlos.
- **Herramientas** y **Almacenamiento**: qué se usa y cuánto pesa cada carpeta.

## Todo junto con Docker

```bash
docker compose up -d --build         # cortex :4748 · panel :4750 · Metrín :4760 (solo 127.0.0.1)
docker compose --profile llm up -d   # además Ollama dentro del compose
docker compose --profile tareas run --rm herramientas s10kb.py indexar   # cualquier script
```

| Servicio | Qué es | Imagen |
|---|---|---|
| `cortex` | Árbol de conocimiento (cortexboard), publicado con socat porque solo escucha en 127.0.0.1 | `Dockerfile` |
| `admin` | Panel de administración | `Dockerfile` |
| `metrin` | Chat de Metrín (RAG en Go). Al arrancar indexa `kb/fragmentos*.jsonl` de forma incremental | `metrin/Dockerfile` |
| `ollama` | Modelo local para las respuestas de Metrín (perfil `llm`) | `ollama/ollama` |

- El proyecto se monta en `/app`: lo que se agrega desde el panel queda en esta carpeta.
- Metrín responde con Ollama. Por defecto usa el de la Mac (`host.docker.internal:11434`); con el perfil `llm`, pon `OLLAMA_URL=http://ollama:11434` en `.env` y baja el modelo: `docker compose exec ollama ollama pull qwen2.5-coder:7b`.
- En Docker el OCR usa tesseract (español); en la Mac, Vision. Los tutoriales en Docker necesitan `ANTHROPIC_API_KEY` en `.env`.
- Cortex rechaza peticiones con `Host` distinto de localhost: los servicios mandan `Host: localhost:4748` (`S10_CORTEX_HOST`).
- Puertos del host configurables: `S10_PUERTO_CORTEX`, `S10_PUERTO_ADMIN`, `S10_PUERTO_METRIN`.

## Actualización programada (`programador.py`)

Servicio `programador` del compose (o `.venv/bin/python programador.py` fuera de Docker).

| Fuente | Qué corre | Frecuencia por defecto |
|---|---|---|
| Portal de ayuda S10 | `s10kb.py rastrear --sin-login --refrescar` | cada día |
| PDF públicos de s10peru.com | `s10kb.py pdfs-publicos` → `descargar --sin-login` → `ocr --completo` | cada semana |
| optimiza360.pe | `s10kb.py sitio optimiza360.pe` | cada semana |
| YouTube (canal oficial) | `youtube.py todo` | cada semana |

- Frecuencia, pausa y "ejecutar ahora" se manejan desde el panel, sección **Programación** (`data/programacion.json`).
- Cada corrida termina con `s10kb.py indexar`. Si hubo cambios, toca `kb/.actualizado` y Metrín recarga su índice solo.
- Cada corrida queda en Cortex: una entrada en el historial de actividad y el nodo `fuentes/programacion` con la tabla de últimas corridas.
- `--refrescar` detecta cambios por huella del texto: solo reescribe las páginas que cambiaron. Los títulos puestos a mano (`titulo_fijo`) se conservan.
- Correr una fuente a mano: `.venv/bin/python programador.py --una optimiza360`.
