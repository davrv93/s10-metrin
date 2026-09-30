# Reproducir el pipeline en otra máquina

Guía probada para levantar `s10kb.py` (portal) y `youtube.py` (tutoriales) en
una máquina limpia. Verificada en macOS arm64 (Apple Silicon) con Python 3.13.7;
al final están las notas para Linux/Windows y x86.

## 1. Requisitos previos

| Requisito | Versión probada | Para qué |
|---|---|---|
| Python | 3.13.7 (sirve 3.10+) | correr los scripts |
| ffmpeg | Homebrew 2026-09 | solo `s10kb.py indexar` (pdftotext) y diagnóstico de audio |
| poppler (`pdftotext`) | Homebrew | solo `s10kb.py indexar` |
| yt-dlp | 2026.8.19 | descargar audio (también se instala por pip) |

```bash
# macOS (Homebrew)
brew install python ffmpeg poppler

# Debian/Ubuntu
sudo apt install python3-venv ffmpeg poppler-utils

# Windows: instalar Python, y ffmpeg/poppler con winget o manualmente
```

ffmpeg **no** hace falta para transcribir: faster-whisper decodifica el audio
con PyAV. Para `s10kb.py indexar` lo necesario es `pdftotext` (poppler). ffmpeg
solo se usó aquí para diagnóstico (`ffprobe`/`volumedetect`) y lo pide yt-dlp
en algunos casos como fallback de postproceso.

> **Verificado el 2026-09-30**: venv limpio desde este `requirements.txt` →
> imports OK, `youtube.py todo` completo con idéntico resultado (23 videos,
> 35 fragmentos).

## 2. Instalar dependencias exactas

```bash
cd s10-conocimiento
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
```

`requirements.txt` está congelado con las versiones con las que se verificó el
pipeline. La restricción crítica:

> **`av` debe quedarse en 18.x** (aquí `18.1.0`). La línea 19.x de PyAV quitó
> el parámetro `metadata_errors` de `av.open()` y faster-whisper 1.2.1 explota
> con `TypeError: open() got an unexpected keyword argument 'metadata_errors'`.
> Si algún día se actualiza faster-whisper, revisar esto antes de subir `av`.

También ojo: en Python 3.13 **no hay wheels de `av` 14.x** (falla compilando
desde fuente); 18.1.0 sí tiene wheel.

## 3. Verificar la instalación (2 minutos)

```bash
.venv/bin/python - <<'PY'
import av, faster_whisper, yt_dlp, requests, bs4
print("av", av.__version__)                      # esperado: 18.1.0
from faster_whisper import WhisperModel
m = WhisperModel("tiny", device="cpu", compute_type="int8")
print("faster-whisper OK")
PY
.venv/bin/yt-dlp --version                       # esperado: 2026.8.19
which pdftotext                                  # solo necesario para `indexar` del portal
```

## 4. Correr el pipeline de YouTube

```bash
.venv/bin/python youtube.py todo                 # descubrir -> audio -> transcribir -> indexar
```

Notas:

- **Primer arranque**: Whisper descarga el modelo `small` (~460 MB) a
  `~/.cache/huggingface`. Sin red salvo esa primera vez (y para yt-dlp).
- **Tiempo**: en un portátil M-series ~35 min para ~93 min de audio con
  `small/int8`. En x86 sin GPU puede tardar el doble o triple; usar
  `--modelo base` para una prueba rápida.
- **Reanudable**: `audio` no repite descargas (`data/yt/bajados.txt`);
  `transcribir` salta videos ya transcritos (un `.json` por video en
  `data/yt/transcripciones/`); `indexar` rehace `kb/` completo en segundos.
- Corridas largas: ejecutar dentro de `tmux`/`screen` o dejar la terminal
  abierta; si se corta, volver a correr el mismo paso y sigue donde quedó.
- Copias de seguridad: basta respaldar `data/yt/transcripciones/` y `kb/` para
  no volver a transcribir nunca.

## 4bis. Pantallas del ERP (`ui.py`)

Extrae los campos de formulario de las pantallas del ERP, a partir de los videos:

```bash
brew install tesseract                    # OCR (solo trae eng)
mkdir -p data/tessdata && cp "$(brew --prefix)/Cellar/tesseract/"*/share/tessdata/{eng,osd}.traineddata data/tessdata/
curl -sL -o data/tessdata/spa.traineddata \
  https://github.com/tesseract-ocr/tessdata_fast/raw/main/spa.traineddata

.venv/bin/python ui.py capturas <id-de-video>   # frames -> data/ui/capturas/
.venv/bin/python ui.py ocr                      # tesseract spa con cajas y jerarquía
.venv/bin/python ui.py vision --limite 10       # VLM local (semántica de campos)
.venv/bin/python ui.py esquema                  # fusiona OCR + VLM -> data/ui/esquema/
.venv/bin/python ui.py pantallas                # agrupa capturas en pantallas distintas
.venv/bin/python ui.py render --esquema data/ui/esquema/<x>.json   # HTML para validar
```

El `render` genera un HTML autocontenido (la captura va embebida en base64): se
abre en el navegador, se arrastra cada campo a su posición real, se corrige el
tipo y se exporta el JSON corregido. Esa es la validación humana del esquema.

El paso `vision` usa **Ollama con `gemma3:4b`** (multimodal, ~3.3 GB):

```bash
ollama pull gemma3:4b
```

> **Ojo con el puerto.** En esta máquina el servidor de Ollama corre en el
> **11435**, no en el 11434 por defecto (lo levantó otra herramienta). `ui.py`
> asume 11435; si tu servidor usa el puerto normal, exporta
> `S10_VLM_HOST=http://127.0.0.1:11434`. El modelo se cambia con `S10_VLM`.

Lecciones medidas (2026-09-30, video `17W0yKI8iew`):

- El OCR de tesseract con `spa` lee el texto del ERP muy bien y da coordenadas.
- La geometría sola **no** reconstruye el formulario: sin semántica confunde
  paneles y celdas de grilla con campos.
- El VLM sí extrae campos reales (`Fecha Inicio = 01/01/2020`, `Valor = 4300`),
  pero mete ruido (lee el calendario del date-picker como campos) y confunde
  mayúsculas (`UIT`→`UT`, `Fijos`→`Fluidos`). Lo correcto es **contrastar VLM
  contra OCR**: el OCR da la cadena exacta, el VLM la semántica.
- La salida JSON se fuerza con un esquema nativo de Ollama (`ESQUEMA_VLM` en
  `ui.py`): sin él el modelo devuelve JSON truncado e inválido.
- El título que devuelve el VLM no sirve para agrupar pantallas (repite
  "S10 Nóminas" en casi todas): `pantallas` agrupa por huella visual (dhash) y
  por campos en común, y después confirma el nombre con el OCR.
- Presupuesto por video de 132 s: ~20 s por captura en `vision` (60 capturas ≈
  20 min), OCR ~1 s por captura y fusión instantánea. `vision` es reanudable:
  si se corta, se vuelve a correr y sigue donde quedó.
- Para que Metrín (el modelo RAG) conozca las pantallas: `ui.py kb` genera
  `kb/fragmentos_pantallas.jsonl` (lo levantan solos Metrín en Docker y
  `tutor.py`, por el glob `kb/fragmentos*.jsonl`). Para indexar a mano:
  `metrin/rag index-kb` — **ojo: reemplaza TODA la fuente `s10-kb`**, así que
  hay que pasarle la concatenación de todos los fragmentos, como hace el
  docker-entry: `cat ../kb/fragmentos*.jsonl > /tmp/kb.jsonl && ./rag index-kb /tmp/kb.jsonl`.
- Para el árbol de conocimiento: `ui.py cortex` genera
  `data/cortex-borradores/pantallas-ui.json` y
  `herramientas/cargar_cortex.py` lo carga como **borradores** del actor
  ai-agent: quedan pendientes de aprobación humana en el tablero.

## 5. Correr el pipeline del portal (s10kb.py)

```bash
cp .env.example .env        # completar S10_USUARIO y S10_CLAVE (cuenta de miembro)
.venv/bin/python s10kb.py todo
```

Sin credenciales: `.venv/bin/python s10kb.py rastrear --sin-login` (solo lo
público; los manuales están tras el login de Simple Membership).

## 6. Notas por plataforma

- **macOS arm64 (verificado)**: todo funciona como está; `ctranslate2` usa
  CPU con `int8` (el script fuerza `device="cpu"`).
- **Linux x86_64**: los wheels de `av`, `ctranslate2` y `onnxruntime` existen
  para glibc estándar (Debian/Ubuntu ok). Mismos comandos.
- **Windows**: crear el venv igual (`.venv\Scripts\pip install -r
  requirements.txt`); yt-dlp y PyAV publican wheels para Windows.
- **Apple Silicon con GPU (opcional, no probado)**: ctranslate2 soporta MPS
  solo parcialmente; no hace falta para este volumen (38 videos).
- Si pip intenta compilar `av` desde fuente: es que tu Python no tiene wheel
  para esa versión → usar Python 3.10–3.13, y mantener `av==18.1.0`.

## 7. Checklist rápido

- [ ] `av --version` → 18.1.0 (dentro del venv)
- [ ] transcripción de prueba con `tiny` OK
- [ ] `youtube.py todo` termina con "38 audios" y `kb/fragmentos_yt.jsonl`
- [ ] `s10kb.py todo` con credenciales → `kb/fragmentos.jsonl`
- [ ] `ui.py ocr` y `ui.py vision` funcionan (tesseract + `gemma3:4b` en Ollama)
- [ ] Respaldar `data/yt/transcripciones/`, `data/pdf/`, `kb/`
