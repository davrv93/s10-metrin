# Metrín V2: RAG procedural multimodal

Estado: especificación y plan (06-10-2026). Responde al megaprompt «ERP Assistant V2 · RAG procedural multimodal +
decision engine local» y a sus nueve entregables de la fase 40. V1 no cambia: V2 vive detrás de un interruptor y está
apagada por defecto.

Documentos relacionados:

- `docs/DISENO-modo-traza-metrin.md` y `docs/CONTRATO-traza-metrin.md`: la traza por etapas.
- `kb/procedimientos/ESQUEMA.md`: el YAML de procedimientos, pasos y fotos.
- `kb/catalogos/`, `kb/conceptos/` y `metrin/plantillas/respuestas.yml`: intenciones, tipo de respuesta, glosario y
  plantillas.
- `metrin/eval/BUSQUEDA.md`: búsqueda híbrida y reranker, con cifras.

---

## 1. Arquitectura actual encontrada (V1)

Verificado en el código de `metrin/` (servicio Go `rag-go`):

| Pieza | Dónde | Cómo funciona hoy |
|---|---|---|
| Entrada | `internal/servidor/servidor.go` | `POST /ask` → `rag.Preguntar`. `GET /health`, `/catalogo`, `/fotos/…` |
| Regla previa | `rag.go` `esPreguntaSobreAciertos` | Corta con una respuesta fija |
| Embedding | `internal/embed` | Estático `potion-es-int8` por defecto (`EMBED_PROVIDER`); opción Ollama `nomic-embed-text` |
| Clasificación | `internal/clasificar` | kNN local sobre `defecto.json`, `UmbralDefecto = 0.40` (calibrado con 28 casos); `social.go` |
| Ruta | `rag/orquestador.go` `planificar` | `rag`, `conversacion` o `regla`; excepción por pregunta larga |
| Reescritura | `rag.go` `reescribir` | Solo con hilo |
| Búsqueda | `internal/almacen` (chromem-go) | **Solo vectorial**, `K = 8`, filtros por `source`, `type`, `ext` |
| Selección | `rag.go` `seleccionarContexto` | `ventanaRelevancia = 0.18`, `maxContextos = 5`, penalizaciones por tipo de fuente |
| Modo tutorial | `rag.go` `esTutorial` | Si los fragmentos son de tutorial o de sección oficial, se arma sin LLM |
| Generación | `internal/llm` | Ollama `qwen2.5-coder:7b` por defecto, o `mlx` (Qwen2.5-3B + LoRA) |
| Fotos | `servidor/fotos_pasos.go` | `vincularFotos` y `colocarFotos`: **después** de generar, hasta 8 por fuente |
| Respaldo | `rag.go` | Texto fijo cuando no hay modelo |
| Decisiones | `internal/jev` | Opcional con `JEV_URL`; juez local `jeva.cpp`; trazas en el visor `:4761` |
| Registro de fallos | `rag.go` `registrarFallo` | JSONL con la pregunta completa |
| Base de conocimiento | `kb/fragmentos*.jsonl` | 8.135 fragmentos con manual, sección, página y `confianza`. **~700 traen `pasos[{texto, fotos[]}]`**: la relación paso ↔ foto ya existe en la fuente. 3.543 son de tipo imagen |
| Tutoriales | `tutoriales/*.md` | 7 tutoriales redactados con citas [F1] |
| Pruebas | `*_test.go` | rag, servidor, clasificar, embed, indexar, trocear, eval, jev, reportes |

## 2. Problemas concretos (medidos el 06-10-2026 en local)

1. **Clasificación.** «¿qué es un metrado?» salió como `social` (0,569), y «¿Cómo registro un nuevo presupuesto en S10?»
   quedó bajo el umbral (0,322 < 0,40). El umbral se calibró con 28 casos.
2. **Búsqueda solo vectorial.** Los términos exactos del ERP no pesan. Para «metrado», las 5 fuentes fueron una sección
   de Gerencia de Proyectos, la biografía del CEO, una imagen suelta y dos páginas institucionales.
3. **Selección sin calibrar.** `ventanaRelevancia = 0.18` no tiene medición documentada.
4. **Unidad equivocada.** Se recuperan páginas o fragmentos, no procedimientos. No hay manera de responder «¿y luego?».
5. **Fotos al final.** Se vinculan después del texto, por fuente, y no por paso.
6. **Modelo.** El generador por defecto es un modelo de código; en local no estaba disponible y salió el texto de
   respaldo.
7. **Sin modo concepto.** «¿Qué es X?» y «¿cómo hago X?» siguen el mismo camino.

## 3. Arquitectura V2 propuesta

```
pregunta
 → normalización (original + normalizada + entidades + alias; nunca se pierden los términos exactos)
 → contexto estructurado (hechos / inferencias / desconocidos; procedimiento y paso en curso)
 → tipo de respuesta: CONCEPT | PROCEDURE | NAVIGATION | TROUBLESHOOTING | CONFIGURATION | COMPARISON | UNKNOWN
     (catálogo kb/catalogos/tipo_respuesta.yml; umbral elegido con datos; si duda → motor de decisión)
 → plan de consulta (qué índice, filtros por módulo, alias)
 → búsqueda híbrida: vector ∥ BM25 (en paralelo) → RRF → TOP 30
 → reranker local → TOP N
 → ¿evidencia suficiente?  no → expandir consulta y buscar otra vez (máx. 2 iteraciones) → si sigue sin → SIN_EVIDENCIA
 → constructor de evidencia: procedimiento (pasos + fotos de cada paso + fuentes) o concepto (definición + fuente)
 → plan de respuesta (JSON, §5)
 → generación: plantillas + LLM local que solo reformula (nunca cambia pasos, fotos, menús ni citas)
 → quality gate (§6) → reintento sin LLM → respuesta segura
 → renderizador determinista: paso → foto → paso → foto
```

**Regla:** los modelos proponen, el motor de decisión decide, el código controla, la evidencia demuestra, el LLM redacta
y el renderizador presenta.

### Interruptor

Igual que el megaprompt:

- `AGENT_VERSION=v1|v2`, por defecto `v1`.
- `AGENT_V2_ENABLED`, por defecto `false`.
- `AGENT_V2_PERCENTAGE`, por defecto `0`.
- La petición puede traer `"version"` para el chat de prueba.

La versión se decide al entrar a `/ask` y los estados de V1 y V2 no se mezclan. Con V1, la respuesta es idéntica byte a
byte a la de hoy.

### Límites

Variables de entorno, con estos valores por defecto:

```
MAX_AGENT_STEPS=5  MAX_DECISION_CALLS=4  MAX_SEARCH_ITERATIONS=2  MAX_TOOL_CALLS=5  MAX_GENERATIONS=1  MAX_REGENERATIONS=1
DECISION_TIMEOUT_MS=1500  GENERATION_TIMEOUT_MS=5000
```

Al llegar a un límite se pasa a la respuesta segura, nunca a un bucle.

### Motores intercambiables

Se eligen con una variable de entorno:

| Interfaz | Implementaciones |
|---|---|
| `EmbeddingEngine` | estático, Ollama |
| `DecisionEngine` | reglas (por defecto), local (`jeva.cpp` o llama-server con gramática), remota, simulada |
| `RerankerEngine` | llama-server `--reranking`, noop |
| `GenerationEngine` | plantillas (por defecto), local, remota |
| `VisionEngine` | ninguna por ahora: solo se leen el `caption` y el `ocr` que vengan en la fuente |

### Memoria de conversación

Se guarda `{procedimiento, paso_actual, pasos_completados}` en el hilo:

- «¿y luego?» o «listo» → el paso siguiente, sin búsqueda nueva.
- «No me sale» → los `errores_frecuentes` del paso.
- Un cambio de intención se contesta y se retoma UNA vez (plantilla `RETOMA`). Si la persona insiste en el tema nuevo,
  se abandona el procedimiento anterior.

### Modos de respuesta

| Tipo | Respuesta |
|---|---|
| CONCEPT | Definición del glosario + dónde está en S10 + «¿quieres que te enseñe a…?». Sin pasos |
| PROCEDURE | Paso → foto → paso → foto, con prerrequisitos y verificación. Por partes si hay más de 6 pasos |
| NAVIGATION | La ruta exacta de menús, solo si está documentada |
| TROUBLESHOOTING | Causa documentada → comprobación → solución, con foto si existe |
| CONFIGURATION | Solo lo que respalda la evidencia |
| COMPARISON | Los dos conceptos del glosario, lado a lado |
| UNKNOWN | Una pregunta de aclaración (plantilla `ACLARAR_TAREA`) |

## 4. Modelo de datos

Vive en YAML, en `kb/procedimientos/` y `kb/conceptos/`, y se carga al arrancar:

- **Procedure:** `id`, `titulo`, `objetivo`, `aliases`, `entidades`, `preguntas`, `prerrequisitos`, `pasos`,
  `verificacion`, `errores_frecuentes`, `fuentes`.
- **Step:** `id` (`<procedimiento>#<n>`), `n`, `accion`, `pagina`, `fuente[]`, `fotos[]`.
- **Image:** `id`, `ruta_o_url`, `pagina`, `caption` y `ocr` (si vienen en la fuente), con el paso al que pertenece.
  Una foto va en un paso solo si la fuente la asocia a ese paso.

## 5. Plan de respuesta (contrato interno)

```json
{
  "version": 2,
  "tipo": "PROCEDURE",
  "procedimiento": {"id": "presupuestos.registrar-presupuesto-nuevo", "titulo": "Registrar un presupuesto nuevo", "confianza": 0.91},
  "prerrequisitos": [{"texto": "…", "fuente": ["<id>"]}],
  "pasos": [
    {"n": 1, "id": "presupuestos.registrar-presupuesto-nuevo#1", "texto": "Vaya al escenario **Datos Generales**.",
     "fotos": [{"id": "img_ab12", "ruta": "…", "pagina": 11}], "fuente": ["<id>"]}
  ],
  "verificacion": [{"texto": "…", "fuente": ["<id>"]}],
  "siguiente": {"plantilla": "PREGUNTA_AVANCE", "paso": 4},
  "fuentes": [{"manual": "Guía de Usuario de S10 Presupuestos", "paginas": [11, 12, 17]}],
  "afirmaciones": [{"texto": "…", "evidencia": {"fragmento": "<id>", "pagina": 12, "paso": 2}}]
}
```

`/ask` en V2 conserva los campos de V1 (`respuesta`, `fuentes`, `modo`…) y añade `plan`. El LLM, si reformula, devuelve
el mismo JSON. El backend valida que cada `id` de foto exista y pertenezca al paso que la trae. Un `id` desconocido
invalida la respuesta.

## 6. Quality gate

Se comprueban, por código y sin modelo:

- respondió la intención;
- cada paso tiene evidencia que existe;
- los términos entre `** **` aparecen en su fuente;
- pasos en orden, sin duplicados y sin pasos que falten frente al procedimiento;
- cada foto pertenece a su paso;
- hay fuentes y ninguna es tangencial;
- no aparecen menús, atajos ni citas fuera del plan.

Puntaje 0–1 con la lista de problemas. Si falla: un reintento sin LLM (solo plantillas). Si vuelve a fallar: respuesta
segura con lo que sí está respaldado, o `SIN_EVIDENCIA`. Nunca falla en silencio: queda en la traza.

Degradación progresiva:

- Sin LLM → plantillas.
- Sin reranker → puntaje híbrido.
- Sin vector → léxica.
- Sin motor de decisión → reglas.
- Sin foto → el paso va sin foto.

## 7. Archivos que se modifican

| Archivo | Cambio |
|---|---|
| `metrin/internal/rag/rag.go`, `orquestador.go` | Un punto de entrada: `if version == v2 → v2.Preguntar` al inicio de `Preguntar`. Nada más de V1 cambia |
| `metrin/internal/servidor/servidor.go` | Campo `version` en la petición; adjunta `plan` |
| `metrin/internal/servidor/pagina.html` | Renderizador desde `plan` (paso → foto), botones «Listo / No me sale», topología V2 en la traza |
| `metrin/internal/config/config.go` | Variables de §3 |
| `metrin/internal/traza/` | Etapas V2: `tipo_respuesta`, `plan_consulta`, `lexica`, `vectorial`, `fusion`, `rerank`, `procedimiento`, `pasos`, `fotos_paso`, `plan`, `quality_gate`, `renderizado` |

## 8. Archivos nuevos

| Archivo | Qué es | Estado |
|---|---|---|
| `kb/procedimientos/**.yml` + `ESQUEMA.md`, `INDICE.json`, `COBERTURA.md` | Procedimientos con pasos, fotos y fuentes | Agente en curso |
| `herramientas/validar_procedimientos.py` | Validador de procedimientos | Agente en curso |
| `kb/catalogos/intenciones*.yml`, `tipo_respuesta*.yml`, `ruido.yml`, `EVALUACION.md` | Catálogos con su prueba aparte | Agente en curso |
| `kb/conceptos/*.yml` | Glosario citado | Agente en curso |
| `metrin/plantillas/respuestas.yml` | Plantillas instruccionales | Agente en curso |
| `metrin/internal/busqueda/`, `metrin/internal/rerank/`, `metrin/cmd/evalbusqueda/`, `metrin/eval/` | Búsqueda híbrida, reranker y evaluación | Agente en curso |
| `metrin/internal/v2/` | `contexto.go`, `tipo.go`, `plan_consulta.go`, `evidencia.go`, `planificador.go`, `gate.go`, `memoria.go`, `motores.go`, `presupuesto.go` | Fase 2 |
| `metrin/eval/v2_oro.jsonl` + `cmd/evalv2` | Benchmark V1 frente a V2 | Fase 2 |

## 9. Riesgos

| Riesgo | Mitigación |
|---|---|
| Cambios sin commit de otra sesión (del 01-10) en `rag.go`, `main.go`, `config.go` | Cambios aditivos; nada se revierte; no se hace commit sin que el usuario decida |
| Procedimientos que no cubren una tarea | V2 cae a búsqueda híbrida sobre fragmentos con `pasos`; `COBERTURA.md` dice qué falta |
| Fotos mal asociadas | Solo las que la fuente asocia al paso; el validador lo comprueba; mejor un paso sin foto |
| El reranker o el LLM local no caben en el EC2 de 8 GB | Motores intercambiables; plantillas sin LLM por defecto; medir RAM y latencia antes de desplegar |
| Datos de evaluación sintéticos | Marcados como tales; el usuario revisa una muestra; se suman preguntas reales cuando las haya |
| Dos agentes en el mismo archivo | Integración en serie: arranca cuando terminen los agentes que tocan `rag.go` y `pagina.html` |

## 10. Plan por fases (orden del megaprompt)

| Fase | Contenido | Estado |
|---|---|---|
| 0 | Auditoría (§1–2) | Hecha |
| 1 | Interruptor V2 + prueba de igualdad de V1 | Fase 2 de integración |
| 2 | Estado estructurado | Ídem |
| 3 | Búsqueda híbrida | Agente en curso |
| 4 | Reranker | Agente en curso |
| 5 | Ingestión procedural (procedimientos YAML desde los manuales) | Agente en curso |
| 6 | Asociación paso ↔ foto | Agente en curso (dentro de los YAML) |
| 7 | Recuperación de procedimientos | Integración |
| 8 | Plan de respuesta estructurado | Integración |
| 9 | Motor de generación local | Integración; benchmark de modelos en la Mac antes de elegir |
| 10 | Quality gate | Integración |
| 11 | Motor de decisión | Integración (reglas por defecto; `jeva.cpp` si el clasificador duda) |
| 12 | Recursividad controlada | Integración |
| 13 | Observabilidad (traza V2) | Integración, sobre la traza que ya se está construyendo |
| 14 | Benchmark V1 frente a V2 | Al final |

Ninguna fase avanza si rompe pruebas. Después de cada una:

1. Correr `go test ./...` y los validadores.
2. Comprobar la prueba de igualdad de V1.
3. Documentar.

## 11. Pruebas

- **Unitarias:** tipo de respuesta, reescritura de consulta, BM25, RRF, reranker (servidor simulado, timeout → noop),
  carga y validación de procedimientos, asociación paso ↔ foto, plan JSON, quality gate, degradaciones (sin LLM, sin
  reranker, sin vector), continuación de conversación («¿y luego?»).
- **Integración:** pregunta → recuperación → procedimiento → respuesta, para:
  - pregunta conceptual;
  - pregunta procedimental;
  - pregunta ambigua;
  - pregunta sin evidencia;
  - pregunta con fotos;
  - continuación;
  - caída del LLM.
- **Igualdad:** V1 con y sin el interruptor, byte a byte.

## 12. Benchmark V1 frente a V2

Mismo conjunto para las dos versiones (`metrin/eval/v2_oro.jsonl`; categorías CONCEPT, PROCEDURE, NAVIGATION,
TROUBLESHOOTING, CONFIGURATION y AMBIGUOUS; con faltas, preguntas cortas y largas, y continuaciones).

Métricas:

- precisión de clasificación;
- Recall@5, Recall@10, MRR y nDCG;
- acierto de procedimiento;
- recall de imágenes y exactitud paso ↔ foto;
- tasa de invención (pasos, menús o fotos sin evidencia);
- latencia p50 y p95;
- RAM y CPU;
- tasa de respaldo.

**Métrica principal: PROCEDURAL ANSWER SUCCESS.** Una respuesta acierta si cumple las siete condiciones:

1. identifica el procedimiento;
2. trae los pasos correctos;
3. conserva el orden;
4. no inventa pasos;
5. asigna bien las fotos;
6. responde la intención;
7. las fuentes respaldan los pasos.

V2 se enciende solo si supera a V1 en esa métrica sin empeorar la tasa de invención.
