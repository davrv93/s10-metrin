# STATUS — s10-conocimiento

## Actualización 2026-10-06: Metrín V2 y modo traza

- **Rama `feat/metrin-v2-traza`** subida a `origin` (sin fusionar a `main`, sin desplegar). Trae el modo traza, la
  V2 procedural (paso → foto), 60 procedimientos YAML, glosario, catálogos, búsqueda híbrida, reranker y benchmarks.
  Detalle y cifras: [`docs/V2-RAG-PROCEDURAL.md`](docs/V2-RAG-PROCEDURAL.md) §14. Pendientes:
  [`PENDIENTES.md`](PENDIENTES.md).
- **Resultado** (194 casos sintéticos): PAS V1 0 % → V2 81,8 % con reranker (60 % sin él), invención V1 17 % → V2 0 %.
- **Servicios locales en la Mac:**

| Servicio | Dónde | Cómo se para |
|---|---|---|
| Conversacional MLX (V1) | `127.0.0.1:8080` | `pkill -f mlx_lm.server` |
| Reranker bge-reranker-v2-m3 (llama-server, Metal) | `127.0.0.1:8091` | `kill $(cat ~/.cache/metrin-modelos/llama-rerank.pid)` |
| JEV | `127.0.0.1:8765` | launchd `com.jevstyle.serve` |
| Contenedor de PRUEBA de Metrín con traza y V2 | `127.0.0.1:4762` | `docker rm -f metrin-traza-prueba`; se recrea con `docs/traza-ejemplos/recrear-4762.sh` |
| Contenedor de siempre (V1) | `127.0.0.1:4760` | sin cambios: imagen anterior, sin V2 |

- **Sin commit a propósito:** el trabajo del 01-10 en `metrin/` (`rag.go`, `main.go`, `config.go`, `go.mod`,
  `cmd_reporte.go`, `reportes`, `pipeline`), que es de otra sesión.

---

Instantánea del **2026-10-01**. Repo: worktree `s10-conocimiento`,
origin `https://github.com/davrv93/s10-metrin.git`, rama `main`.

## 1. Status actual

**Qué es:** base de conocimiento del ERP S10 (manuales oficiales,
portal, YouTube, pantallas, reportes) + **Metrín**, el asistente de
obra (chat RAG en Go) + **JEV**, el juez de decisiones
(`jev-style-0.8b-decision-v3`).

**Servicios verificados vivos hoy:**

| Servicio | Dónde | Estado |
|---|---|---|
| JEV (juez) | `127.0.0.1:8765` (launchd `com.jevstyle.serve`, MLX del Mac) | vivo; `POST /v1/systemone` con `state` OK |
| Conversacional Metrín | `127.0.0.1:8080` (`mlx_lm.server` + LoRA estilo) | vivo (`Qwen2.5-3B-Instruct-4bit`) |
| compose `s10-conocimiento` | cortex :4748 · admin :4750 · metrin :4760 · trazas :4761 | definidos en `docker-compose.yml` |

**`.env` (raíz) relevante:** `JEV_URL=http://host.docker.internal:8765`,
`JEV_MODEL=jev-style-0.8b-decision-v3`, `LLM_PROVIDER=mlx`,
`OLLAMA_MODEL=mlx-community/Qwen2.5-3B-Instruct-4bit`,
`TRAZAS_REPO=postgres` (trazas-pg del compose).

**Falta:** `S10_USUARIO`/`S10_CLAVE` (cuenta de miembro) → `s10kb.py
oficial` no puede correr.

**Pipeline de preguntas:**
- 2,556 candidatas generadas
- **Juzgado real en curso** con Jev (`127.0.0.1:8765`)
- 2,553 pendientes de juzgar; fondo: `data/preguntas.db`
- Siguiente: `vb` → `sync_cortex_semillas.py` → `reindex`

**Base de conocimiento (`kb/`, 16 MB):** 8 135 fragmentos — 5 865
web/PDF + 1 355 Cortex + 808 FAQ + 66 pantallas + 6 reportes + 35
YouTube. `data/` ocupa 437 MB (no versionado).

## 2. Pendientes (por prioridad)

1. **Esperar fin de juzgado real** con Jev y revisar distribución de
    veredictos/umbrales.
2. **Calibrar umbrales Jev** según resultados reales (mock anterior
   usaba 0.85 fijo).
3. **Caso 8/13 del eval responde en vez de abstener** — endurecer
   `RAG_MAX_DISTANCIA` (0.80→0.70) o exigir distancia mínima del mejor
   trozo antes de generar.
4. **JEV califica bajo las abstenciones correctas** (noul 0.228–0.372
   en casos 9–12): recalibrar instrucciones/criterios del prompt o
   umbral por conducta.
5. **Confianza siempre null** en trazas JEV: el modelo/servidor no la
   devuelve; revisar parseo en `metrin/internal/jev/` o aceptar que no
   hay confianza.
6. **`JEV_STATUS.md` actualizado** refleja Jev en :8765, no mock.
7. **Cobertura de fuentes:** el manual de Gerencia de Proyectos cae al
   troceado genérico; 60 HTML huérfanos con muro de miembros sin
   enlazar desde `paginas.jsonl`.
8. **Stack final de Metrín pendiente** (ver
   `docs/IMPLEMENTACION_METRIN_ESTADO.md`): Docker/PostgreSQL/Redis,
   streaming, autenticación, multitenencia; el índice de sesión en
   `/tmp` es temporal.
9. **Git:** main sin commits pendientes locales de `.cortex/` ni
   `data/html/` (regla del dueño).

## 3. AGENTS.md — cómo trabajar en este proyecto

Reglas operativas:

- Empezar cada sesión con `cortex_brief`; **no** leer todo el árbol.
- Buscar en Cortex antes de cambiar código que no se entiende a fondo;
  el conocimiento marcado *stale* puede estar mal: verificar en el
  código y corregir el nodo.
- Escribir todo en **español**; resúmenes de nodo < 300 caracteres.
- Explicar el **porqué** de cada cambio y registrarlo con
  `cortex_log_activity`.
- Revisar `cortex_inbox` antes de empezar trabajo nuevo.
- **Nunca editar `.cortex/rules/`** (son de humanos).
- **No commitear `.cortex/` ni `data/html/`** (regla del dueño).
- El juez es **Jev-Style 0.8B** (`jev-style-0.8b-decision-v3`); las
  decisiones son `POST /v1/systemone` con campo `state`.
- El modelo JEV **no** está en :8080 (ese es el conversacional MLX).

## 4. Orquestación

Ver **`ORCHESTRATION.md`** para el flujo completo del pipeline,
servicios, variables y estado actual.

## 5. SKILL.md — flujos reutilizables

```bash
# Base de conocimiento (venv .venv; requiere pdftotext)
.venv/bin/python s10kb.py oficial        # manuales tras el muro (necesita S10_USUARIO/S10_CLAVE)
.venv/bin/python s10kb.py indexar       # rehace kb/ con todo lo bajado
.venv/bin/python youtube.py todo        # tutoriales del canal: audio -> whisper -> fragmentos
.venv/bin/python ui.py pantallas        # pantallas del ERP desde videos (OCR+VLM)
.venv/bin/python reportes.py preguntar "…"   # NL -> reporte (plantillas aprobadas, sin SQL de IA)
.venv/bin/python tutor.py auto --max 8  # tutoriales automáticos desde el árbol de Cortex
.venv/bin/python admin.py               # panel :4750 (ADMIN_USUARIO/ADMIN_CLAVE)
.venv/bin/python trazas/trazas.py       # visor de trazas JEV :4761

# Pipeline de preguntas
.venv/bin/python metrin/pipeline/preguntas.py generar --limite 0
.venv/bin/python metrin/pipeline/preguntas.py juzgar
.venv/bin/python metrin/pipeline/preguntas.py vb --min-puntaje 0.80 --min-rubrica 0.80
.venv/bin/python metrin/pipeline/sync_cortex_semillas.py
npx -y cortexboard@0.4.0 reindex

# Estado actual del pipeline
# - JEV_URL=http://127.0.0.1:8765 (no mock)
# - JEV_MODEL=jev-style-0.8b-decision-v3
# - Fase actual: juzgado real en background (2,553 candidatas)
# - Siguiente: vb → sync → reindex

# Metrín (Go, metrin/)
./rag index-kb ../kb/fragmentos.jsonl   # índice
./rag ask "¿…?" [--solo-buscar]         # pregunta por CLI
./rag serve --addr 127.0.0.1:4760       # API /ask + /health
./rag eval [--casos f] [--k 8] [--umbral 0.5] [--json f]   # harness de decisión
```

**Receta del eval en vivo:**

```bash
cd metrin && go build -o rag ./cmd/rag
LLM_PROVIDER=mlx MLX_URL=http://127.0.0.1:8080 \
OLLAMA_MODEL=mlx-community/Qwen2.5-3B-Instruct-4bit \
JEV_URL=http://127.0.0.1:8765 JEV_MODEL=jev-style-0.8b-decision-v3 \
RAG_TIMEOUT=120 ./rag eval --trazas /tmp/eval-trazas.jsonl --json /tmp/eval-status.json
```

## 5. CONTEXT.md — qué es y cómo está armado

- **s10-conocimiento** es la fuente de verdad de contenido: rastrea y
  indexa documentación oficial de S10 ERP (documentacion.s10peru.com
  tras login de miembro, PDFs, YouTube, pantallas de video). Confianza
  por fuente: oficial S10 > terceros (penalizados en distancia) >
  casos académicos.
- **Metrín** (`metrin/`, Go): RAG sobre `kb/fragmentos*.jsonl`
  (BM25 + embeddings estáticos `modelos/potion-es-int8.pjge`),
  clasificador de intención, generación con el conversacional MLX
  (:8080) y decisiones evaluativas con JEV (:8765). Abstiene cuando
  la evidencia no alcanza; **no inventa** datos, acciones ni fuentes.
- **JEV** juzga proposiciones (`noul` = p(true)) sobre el estado
  compartido de cada turno; sus decisiones se emiten a un JSONL que
  lee el **visor de trazas** (:4761, `TRAZAS_REPO` mock | archivo |
  postgres). El harness `rag eval` usa el mismo camino (origen
  `metrin-eval`).
- **Cortex** (`.cortex/`, tablero :4748): árbol de conocimiento
  (`tree/` con ramas `manuales-oficiales`, `optimiza360`,
  `semillas-generadas`), reglas, items y actividad. Es la fuente de
  verdad del proyecto; los humanos revisan todo lo que escribe la IA.
- **Programador** (`programador.py`): reindexa portal, manuales, PDFs
  públicos, optimiza360 y YouTube en frecuencias configurables desde
  el panel; cada corrida toca `kb/.actualizado` y Metrín recarga.
- Datos grandes no versionados: `data/` (437 MB: html, pdf, yt, ui,
  cortex-borradores) y `kb/` sí está versionado (16 MB).

> Nota: este documento es una instantánea. Vuelve a correr
> `rag eval` y `git status` antes de decisiones de despliegue.
