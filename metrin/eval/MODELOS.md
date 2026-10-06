# Modelos locales para la V2 de Metrín: decisión y generación

Medición del 06-10-2026 (02:20–06:15, hora de Lima) en la Mac de desarrollo (M4 Pro, 12 núcleos, 24 GB), con
llama-server de Homebrew (`version 0.5.0, build 11146`). Responde a la fase 9 de `docs/V2-RAG-PROCEDURAL.md`: no
cambiar el modelo sin benchmark y no usar un modelo grande para decisiones simples. **No se tocó el código del
servicio.** Script: `herramientas/benchmark_modelos.py`. Datos y resultados: `metrin/eval/modelos/`.

## 1. Conclusión

| Papel | Mac de desarrollo (Metal) | EC2 de 8 GB sin GPU (4 vCPU) | EC2 de 16 GB |
|---|---|---|---|
| `DECISION_MODEL`, tipo de respuesta | **`qwen3.5-4b`** Q4_K_M: 98,9 % (89/90), p50 445 ms con la GPU compartida | **`qwen3.5-2b`** Q4_K_M, solo cuando las reglas duden: 87,8–88,9 %, **422 ms p50 medido en 4 hilos de la M4** (en el EC2, 1,5–3× más: §6.4), 2,8 GB de RSS | igual que con 8 GB: el límite es la CPU, no la RAM (un 4B tarda ~1 s ya en la Mac) |
| `DECISION_MODEL`, suficiencia de evidencia | `qwen3.5-4b` (82,5 %; sus 7 fallos son «no», el error seguro) o el Jev MLX que ya corre en `:8765` (82,5 %, 190 ms) | **sin LLM**: puntaje del reranker + reglas. Ningún candidato entra en `DECISION_TIMEOUT_MS = 1500` en CPU (el más rápido, 2,4 s, acierta el 55 %) | ídem |
| `GENERATION_MODEL` | **ninguno: plantillas.** Para experimentar con planes de respaldo: `qwen3-4b-2507` con el contrato «solo textos» (§5.2) | **ninguno: plantillas** | ídem |

**La V2 debe generar con plantillas por defecto** y usar un LLM pequeño solo para el tipo de respuesta cuando las reglas
duden. Las cifras:

1. **El LLM casi no aporta texto.** Con la consigna de no cambiar el sentido, devuelve idénticos entre el 74 % y el 98 %
   de los pasos; lo único nuevo es una oración de introducción, que una plantilla escribe en 0 ms. Con una consigna de
   reescribir de verdad (`qwen3-4b-2507`, §5.2) sí limpia el texto crudo del manual, pero la nota a ciegas en esos planes
   queda igual que copiando (3,50 de 5) porque a la vez mete erratas («Ingle al módulo») o cambia quién hace la acción.
2. **No entra en el presupuesto.** `GENERATION_TIMEOUT_MS = 5000`. Un plan de 5–6 pasos con la salida JSON completa
   tarda en Metal, sin contención, de 4,1 s (`qwen3-1.7b`) a 12 s (7B); en 4 hilos de CPU, 16,6 s (`qwen3.5-2b`,
   medido) a ~80 s (7B, estimado). Devolviendo solo los textos baja a ~10,7 s en CPU con `qwen3.5-2b`: solo leer el
   plan ya cuesta 5,2 s.
3. **Cuando reformula, rompe el plan.** `qwen3.5-4b`, el que más reescribe (copia 74 %), pierde o cambia negritas en el
   13,9 % de los planes; `qwen3.5-2b` alteró un `id` de paso («emision-electronica» → «emision-electica») y, al pedirle
   reescribir, quitó las `**` en el 86 % de los planes. El quality gate lo atrapa, pero entonces el LLM solo suma
   latencia.
4. **Para decidir sí sirve un modelo pequeño, y el de 7B no compra nada**: `qwen3.5-2b` acierta el tipo en el 88 % con
   ~23 tokens nuevos de prompt y 8 de salida (~0,4 s en CPU); el 7B acierta el 92 %, falla toda la evidencia y tarda
   2–4 veces más.

El modelo actual por defecto, **`qwen2.5-coder:7b`, no conviene para ningún papel**: es un modelo de código, ocupa
4,7 GB, en tipo de respuesta empata con `qwen3-4b-2507` (92 %), en evidencia contesta «no» a los 40 casos (50 %), y
tarda 12 s por plan en Metal sin contención y ~80 s estimados en CPU (§4.1, §5.1, §6.2).

## 2. Candidatos

Licencia y GGUF comprobados en Hugging Face el 06-10-2026 (API de modelos: `cardData.license` y lista de archivos).

| Modelo (clave en el script) | GGUF usado | Tamaño | Licencia | Notas |
|---|---|---|---|---|
| `qwen3.5-2b` | `unsloth/Qwen3.5-2B-GGUF` Q4_K_M | 1,28 GB | Apache-2.0 | Híbrido (Gated DeltaNet). `enable_thinking: false` y `--reasoning off` |
| `qwen3.5-4b` | `unsloth/Qwen3.5-4B-GGUF` Q4_K_M | 2,74 GB | Apache-2.0 | Ídem |
| `qwen3-1.7b` | blob de Ollama `qwen3:1.7b` (Q4_K_M, ya estaba en la Mac) | 1,36 GB | Apache-2.0 | Sin descarga. `enable_thinking: false` |
| `qwen3-4b-2507` | `unsloth/Qwen3-4B-Instruct-2507-GGUF` Q4_K_M | 2,50 GB | Apache-2.0 | Solo instrucción, sin razonamiento |
| `granite-4.0-micro` | `ibm-granite/granite-4.0-micro-GGUF` Q4_K_M (oficial de IBM) | 2,10 GB | Apache-2.0 | Español entre sus idiomas oficiales |
| `qwen2.5-coder-7b` | blob de Ollama `qwen2.5-coder:7b` (Q4_K_M) | 4,68 GB | Apache-2.0 | **Defecto actual de Metrín**; hizo de techo |
| Jev 0.8B v3 | servicio MLX de `:8765` (`jev-style serve --release 0.8b-v3 --precision 8bit`) y `jeva.cpp/models/Jev-Style-0.8B-Decision-v3-Q4_K_M.gguf` | 0,53 GB (GGUF) | ver ficha de `chaoliangUNSW` | API `POST /v1/systemone` (decisión por logits, sin generar) |

Existen y quedaron sin medir por tiempo: Granite 4.0 1B (`ibm-granite/granite-4.0-1b-GGUF`), Ministral 3 3B
(`mistralai/Ministral-3-3B-Instruct-2512-GGUF`, Apache-2.0), SmolLM3-3B (`ggml-org/SmolLM3-3B-GGUF`, Apache-2.0),
Phi-4-mini (MIT) y Llama 3.2 3B (licencia Llama 3.2). Descartados por licencia, como decía la investigación previa:
Qwen2.5-3B (licencia de investigación) y Hunyuan. Ninguna descarga falló.

## 3. Método

### 3.1 Conjuntos (`metrin/eval/modelos/`)

| Archivo | Casos | Qué mide | Cómo se armó |
|---|---|---|---|
| `decision.jsonl`, `tarea: tipo_respuesta` | 90 | 7 clases: CONCEPT 14, PROCEDURE 19, NAVIGATION 12, TROUBLESHOOTING 13, CONFIGURATION 11, COMPARISON 9, UNKNOWN 12 | Escritas a mano, con faltas, sin tildes, cortas y largas; incluye las 9 del encargo. 5 casos ambiguos llevan `aceptables` (exactitud «laxa»). Trampas a propósito: «Explícame paso a paso cómo…» (PROCEDURE), «no encuentro dónde está…» (NAVIGATION) |
| `decision.jsonl`, `tarea: evidencia` | 40 (20 sí / 20 no) | ¿Basta la evidencia para responder? | Pregunta + 2 fragmentos reales de `kb/fragmentos.jsonl` (manuales oficiales, recortados a 800 caracteres: el modelo ve justo ese texto). Etiquetadas a mano sobre el texto recortado; 11 de los «no» son difíciles (mismo tema, otra pregunta). Cada caso dice por qué en `nota` |
| `generacion.jsonl` | 36 | Reformular sin romper el plan | 28 planes de `kb/procedimientos/**/*.yml` (instantánea de las 02:20, cuando había 43 procedimientos: uno por procedimiento, repartidos por módulo, primeros ≤ 6 pasos, como la «entrega por partes») + 8 planes de **respaldo** construidos con fragmentos oficiales que traen `pasos` (texto crudo del manual; negritas marcadas a mano solo sobre nombres de interfaz literales). Fotos con `id = sha1(ruta)[:12]`, como en el esquema |

Todo es sintético en las preguntas; los fragmentos y los pasos son reales.

### 3.2 Cómo se pregunta

- **Decisión** con `POST /v1/chat/completions`, `temperature 0`, `response_format: json_schema` con un `enum`
  (`{"tipo": …}` o `{"suficiente": "si"|"no"}`). Se repite en Metal sin gramática para medir el JSON válido «libre».
  El prompt de sistema (las 7 definiciones) se reutiliza de la caché: solo se procesan ~20 tokens nuevos por
  pregunta de tipo y ~265 por pregunta de evidencia.
- **Jev** con su propia API `POST /v1/systemone`: `choice` con las mismas 7 definiciones como criterios, y `noul`
  (umbral 0,5) para la evidencia.
- **Generación**: el plan va como JSON compacto (`titulo`, `objetivo`, `pasos[{n, id, texto, fotos}]`) y se pide el
  mismo JSON con `json_schema`. Variantes: `textos` (el modelo solo devuelve `{"intro", "textos": [...]}` y el código
  pone `n`, `id` y `fotos`) y `reescribir` (además se le pide imperativo de usted y quitar los rótulos «Llamada a:»).
  Los prompts están en el script (`SISTEMA_TIPO`, `SISTEMA_EVID`, `SISTEMA_GEN`, `SISTEMA_GEN_TEXTOS`,
  `SISTEMA_GEN_REESCRIBIR`).

### 3.3 Generación válida (lo comprueba el código, como el quality gate de §6)

JSON con la forma pedida; mismos pasos con el mismo `n` e `id` en el mismo orden; mismas fotos en cada paso; mismas
negritas en cada paso (exactas tras normalizar espacios); y **sin términos nuevos**: ninguna cita «…», atajo de
teclado ni palabra con mayúscula a media oración que no esté en el plan (se excluyen `S10`, `ERP` y `Metrín`, que
vienen del prompt). Se informan además la tasa de copia y la razón de longitud salida/entrada.

### 3.4 Calidad del español, a ciegas

Muestra estratificada fija (semilla 20261006): 4 planes de procedimiento y 4 de respaldo. Las salidas de cada plan se
barajan con letras A–E sin nombre de modelo (`ciego/muestra_lote1.md`); se puntuaron antes de abrir la clave
(`ciego/clave_lote1.json`) y se dejaron los motivos en `ciego/notas_lote1.md`. Las salidas de la variante
«reescribir» (§5.2) se puntuaron igual en un segundo lote (`muestra_lote2.md`, `clave_lote2.json`,
`notas_lote2.md`). Rúbrica 1–5:

| Nota | Criterio |
|---|---|
| 5 | Español neutro y natural, gramática correcta, imperativo de usted consistente, breve y fiel al plan |
| 4 | Fiel y correcto, con un detalle menor (frase algo torpe, redundancia, un tú aislado) |
| 3 | Se entiende pero suena forzado o repetitivo, o un paso pierde un matiz sin cambiar la instrucción. **Tope para una copia literal de un paso crudo del manual** («Llamada a:», infinitivos, «la stock»): el generador no aportó legibilidad |
| 2 | Errores gramaticales, palabras en inglés o regionalismos, o un paso cambia de sentido |
| 1 | Inservible: otro idioma, texto roto, pasos fundidos o perdidos, contenido inventado |

### 3.5 Condiciones de la medición (importante para leer las latencias)

La Mac no estaba dedicada. Durante casi toda la medición:

- otro agente evaluaba un reranker en `llama-server :8091` (después también `:8093`), con la GPU al 97–100 %
  (`ioreg`, «Device Utilization %») y, desde las 05:15, en CPU con ~600 % de uso;
- corrían además MLX en `:8080`, Jev en `:8765`, Ollama y la VM de Docker; la carga media subió de 3,5 a **60**, con
  11 GB de swap ocupados;
- la Mac estuvo con batería (llegó al 15 %) y se durmió 14 minutos a las 02:36.

Por eso: **la exactitud y la validez son fiables** (decodificación voraz, `temperature 0`, `top_k 1`), pero **las
latencias de Metal desde las 02:52 y casi todas las de CPU están infladas** por contención. Para la CPU se midió además
el rendimiento bruto con `llama-bench` (`pp512` y `tg64`, 4 hilos, sin GPU) anotando la carga, y la latencia de cada
tarea se estima como `tokens_nuevos / pp + tokens_salida / tg` (§6). Esa estimación coincide con lo medido cuando la
máquina estaba tranquila: `qwen3.5-2b` en CPU, estimado 0,48 s por decisión de tipo frente a 0,42 s medido, y 19,6 s
por plan frente a 16,6 s medido.

## 4. Resultados de decisión

### 4.1 Tipo de respuesta (90 casos) y suficiencia de evidencia (40)

Metal, con `json_schema`. «Laxa» acepta la segunda etiqueta de los 5 casos ambiguos.

| Modelo | Tipo: exactitud (laxa) | F1 macro | Evidencia: exactitud | F1 «sí» / «no» | p50 / p95 ms tipo | p50 / p95 ms evidencia |
|---|---|---|---|---|---|---|
| **qwen3.5-4b** | **0,989** (1,000) | **0,989** | 0,825 | 0,788 / 0,85 | 445 / 594 | 1.136 / 1.583 |
| qwen3-4b-2507 | 0,911 (0,933) | 0,894 | **0,900** | 0,905 / 0,895 | 191 / 257 | 694 / 1.060 |
| granite-4.0-micro | 0,900 (0,933) | 0,892 | 0,800 | 0,818 / 0,778 | 552 / 806 | 1.156 / 1.677 |
| qwen3.5-2b | 0,878 (0,900) | 0,872 | 0,550 | 0,182 / 0,690 | 1.096* / 2.302 | 751 / 917 |
| qwen3-1.7b | 0,833 (0,856) | 0,812 | 0,700 | 0,571 / 0,769 | 242 / 381 | 559 / 746 |
| Jev 0.8B v3, MLX `:8765` | 0,811 (0,833) | 0,799 | 0,825 | 0,844 / 0,800 | 185 / 248 | 190 / 301 |
| Jev 0.8B v3, GGUF en `jeva.cpp` | 0,156 | 0,038 | 0,500 | — | 125 / 184 | 185 / 284 |
| qwen2.5-coder-7b (defecto actual) | 0,922 (0,956) | 0,911 | 0,500 | 0,000 / 0,667 | 795 / 1.377 | 2.322 / 5.291 |

\* `qwen3.5-2b` sin gramática: mismo acierto y p50 398 ms; en CPU con gramática, 422 ms. Los 1.096 ms parecen
sobrecoste de la gramática con su vocabulario de 248 k tokens en Metal, o contención; no se aisló.

F1 por clase (tipo), Metal:

| Modelo | CONCEPT | PROCEDURE | NAVIGATION | TROUBLESHOOTING | CONFIGURATION | COMPARISON | UNKNOWN |
|---|---|---|---|---|---|---|---|
| qwen3.5-4b | 0,97 | 1,00 | 1,00 | 1,00 | 1,00 | 1,00 | 0,96 |
| qwen3-4b-2507 | 0,80 | 1,00 | 0,96 | 1,00 | 1,00 | 1,00 | 0,50 |
| granite-4.0-micro | 0,85 | 0,97 | 0,89 | 0,92 | 1,00 | 0,94 | 0,67 |
| qwen3.5-2b | 0,81 | 0,92 | 1,00 | 0,92 | 0,87 | 0,88 | 0,70 |
| qwen3-1.7b | 0,80 | 0,95 | 0,80 | 0,92 | 0,84 | 0,88 | 0,50 |
| Jev MLX | 0,74 | 0,97 | 0,91 | 0,83 | 0,76 | 0,88 | 0,50 |

Lectura:

- **UNKNOWN es la clase difícil para todos** salvo `qwen3.5-4b`: los mensajes de solo tema («metrado s10», «orden de
  compra», «quiero saber sobre costos») salen como CONCEPT. Es barato resolverlo por reglas antes del modelo (mensaje
  corto sin verbo ni interrogativo → `ACLARAR_TAREA`), pero esa regla hay que medirla en un conjunto aparte: con este
  conjunto ya visto, sería ajustarla a la prueba.
- Los nueve casos del encargo: `qwen3.5-4b`, `qwen3-4b-2507` y `qwen3.5-2b` aciertan los nueve; `granite` y
  `qwen3-1.7b` fallan «metrado s10» (CONCEPT y NAVIGATION) y Jev MLX además «¿Dónde encuentro los metrados?»
  (UNKNOWN). «¿Qué es un metrado?» sale CONCEPT en todos (V1 lo mandaba a social).
- **El 7B no compra nada en decisión**: en tipo empata con `qwen3-4b-2507` (92 % frente a 91 %, un caso) y en
  evidencia contesta «no» a los 40 casos (50 %). Ocupa casi el doble y es 2–4 veces más lento.
- **Evidencia**: `qwen3.5-2b` dice «no» a casi todo (18 de los 20 «sí» los marcó «no») y no sirve para esta decisión.
  Los fallos de `qwen3.5-4b` son todos «no» ante evidencia suficiente: un error **seguro** (lleva a buscar otra vez o a
  `SIN_EVIDENCIA`). `qwen3-4b-2507` falla menos, pero 3 de sus 4 fallos son «sí» ante evidencia insuficiente: el error
  **peligroso**, porque deja responder sin respaldo. Jev MLX tiene 6 «sí» indebidos de 7 fallos.
- **JSON**: 100 % válido con `json_schema` en todos. Sin gramática también 100 %, salvo `qwen3-1.7b` (89 de 90 en
  tipo). La gramática no cambió ningún acierto: conviene dejarla porque garantiza el `enum`.

### 4.2 Jev

- **El servicio MLX de `:8765`** es el más rápido (185–190 ms p50, decide por logits sin generar) y empata con
  `qwen3.5-4b` en evidencia (0,825), pero en tipo queda en 0,811. Su huella física es de **3,6 GB** (`footprint`; el
  RSS de 91 MB engaña porque MLX vive en memoria de Metal): para 0,8B parámetros es mucha memoria.
- **El mismo modelo en GGUF con `jeva.cpp`**, que es lo único que correría en el EC2 (allí no hay MLX), **no decide**:
  con la plantilla por defecto de `jeva.cpp` la distribución sale casi uniforme (confianza 0,04, siempre la primera
  opción; `noul` ≈ 0,7 para todo). Apagando el razonamiento de la plantilla
  (`--chat-template-kwargs {"enable_thinking":false}`) sale igual: 15,6 % y 50 %. La causa probable es que la
  *release* de Jev trae su propio lector de decisiones (`readout_config.json` y el binario `jev-score` del paquete
  `jev-style`, con backends `mlx`, `torch` y `gguf`), y `jeva.cpp` con su plantilla genérica no lo reproduce. **En el
  EC2, Jev tendría que ir con `jev-style serve` y su backend `gguf` o `torch`; eso no se midió.**

### 4.3 CPU de 4 hilos (`-ngl 0 -dev none -t 4 -tb 4`)

En los casos comunes, la CPU contesta lo mismo que Metal salvo 0–2 casos por modelo (1 de 110 en `qwen3.5-2b`, 2 de 110
en `qwen3-4b-2507`; el redondeo numérico es distinto). Las latencias:

| Modelo | Tipo p50 / p95 ms | Evidencia p50 / p95 ms | Casos (tipo / evid.) | Carga media durante la corrida | ¿Fiable? |
|---|---|---|---|---|---|
| qwen3.5-2b | **422 / 535** | 2.449 / 4.045 | 90 / 20 | baja (02:31–02:35) | **sí** |
| qwen3-1.7b | 1.028 / 2.190 | 9.517 / 17.375 | 30 / 10 | 15–25 | no, contención |
| granite-4.0-micro | 1.182 / 2.421 | 15.116 / 25.318 | 30 / 10 | 25–38 | no, contención |
| qwen3-4b-2507 | 3.742 / 8.554 | 21.194 / 37.406 | 90 / 20 | 26–40 y swap | no, contención |
| qwen3.5-4b | 7.945 / 39.752 | 74.859 / 197.217 | 30 / 10 | 22–60 | no, contención extrema (acierto: 29/30 y 10/10) |
| Jev GGUF (`jeva.cpp`) | 1.840 / 2.786 | 729 / 1.129 | 90 / 20 | alta | no decide (§4.2) |

## 5. Resultados de generación

### 5.1 Variante «completa» (el modelo devuelve el JSON entero), Metal, 36 planes

| Modelo | Válidas | Orden e ids | Fotos | Negritas | Sin términos nuevos | Copia | Calidad a ciegas (procedimiento / respaldo) | p50 / p95 s | Tokens de salida p50 |
|---|---|---|---|---|---|---|---|---|---|
| **qwen3-4b-2507** | **100 %** | 100 % | 100 % | 100 % | 100 % | 92 % | **4,12** (4,75 / 3,50) | 19,8 / 34,3* | 498 |
| qwen3.5-2b | 97,2 % | 97,2 % (cambió un `id`) | 100 % | 100 % | 100 % | 98 % | 3,50 (3,75 / 3,25) | **6,3** / 15,5 | 510 |
| granite-4.0-micro | 94,4 % | 100 % | 97,2 % | 97,2 % | 100 % | 89 % | 3,62 (3,75 / 3,50) | 15,5 / 19,9* | 456 |
| qwen3-1.7b | 91,7 % | 100 % | 100 % | 94,4 % | 97,2 % («PLAN») | 97 % | 3,50 (4,00 / 3,00) | 8,4 / 12,2* | 510 |
| qwen3.5-4b | 86,1 % | 100 % | 100 % | 86,1 % | 100 % | 74 % | **4,12** (5,00 / 3,25) | 20,4 / 32,1* | 401 |
| qwen2.5-coder-7b | solo humo: 1/1 válida, 12,0 s en Metal sin contención | | | | | | | | 452 |

\* Con la GPU compartida con el reranker de `:8091`. Sin contención (prueba de humo de las 02:19, un plan):
`qwen3.5-2b` 5,0 s, `qwen3-1.7b` 4,1 s, `qwen3-4b-2507` 6,9 s, `granite-4.0-micro` 6,7 s, `qwen3.5-4b` 8,1 s,
`qwen2.5-coder-7b` 12,0 s. **Ninguno baja de 4 s ni en Metal libre**, y el presupuesto es 5 s.

Lectura:

- **La salida es casi toda estructura copiada.** ~500 tokens de salida para 5–6 pasos: `id` largos, fotos, comillas y
  sangría que los modelos añaden aunque el plan vaya compacto. Por eso tarda tanto aun copiando.
- **Las notas a ciegas las decide la intro**, porque los pasos vuelven copiados. En los planes de procedimiento (texto
  ya pulido) las intros son buenas (3,75–5). En los de respaldo (texto crudo del manual), casi todos copian «Llamada a:»,
  infinitivos y erratas («la stock», «cabera»): 3,0–3,5. Limpiaron el texto crudo, en algún plan, `qwen3-4b-2507`
  (corrige gramática y «cabecera»), `granite` (quita los rótulos) y `qwen3.5-4b` (pasa a imperativo de usted); este
  último, en el plan de importación de Excel, además perdió un requisito («en una carpeta del servidor y cerrado»),
  cambió el sentido de un paso y movió negritas (nota 2). `granite` reescribió un plan entero en tú.
- Errores de forma vistos: un `id` alterado (`qwen3.5-2b`), negritas perdidas o añadidas (`qwen3.5-4b` 5 planes,
  `qwen3-1.7b` 2, `granite` 1; en la muestra, `qwen3.5-4b` puso «**Llamada a**» en negrita y `qwen3-1.7b` quitó las
  de un plan entero), una foto cambiada (`granite`), «publicados» → «publicadas» (error de concordancia introducido por
  `qwen3.5-2b` y `granite`), intros en tú pese a pedir usted.

### 5.2 Variantes «solo textos» y «reescribir»

Contrato «solo textos»: el modelo devuelve `{"intro", "textos": [...]}` con exactamente un texto por paso (lo fuerza la
gramática con `minItems = maxItems`) y el código pone `n`, `id` y `fotos`. Así un `id` o una foto **no se pueden**
romper, y la salida baja de ~500 a ~190 tokens.

| Variante | Modelo | Modo | Válidas | Negritas | Sin nuevos | Copia (procedimiento / respaldo) | Calidad a ciegas | Tokens de salida p50 | p50 medido | p50 estimado sin contención |
|---|---|---|---|---|---|---|---|---|---|---|
| `textos` (misma consigna) | qwen3.5-2b | CPU 4 hilos, 7 planes | 100 % | 100 % | 100 % | 97 % | — | **191** (frente a 522) | 26,1 s (carga 16–30) | **10,7 s** (frente a 19,6 s) |
| `reescribir` (imperativo, sin «Llamada a:») | **qwen3-4b-2507** | Metal, 36 | **100 %** | 100 % | 100 % | 94 % / **11 %** | 3,75 (4,00 / 3,50) | 193 | 15,8 s* | ~3,6 s |
| `reescribir` | qwen3.5-2b | Metal, 36 | **11 %** | 14 % | 97 % | 29 % / 20 % | 2,50 (2,25 / 2,75) | 170 | 9,5 s* | ~1,9 s |

\* GPU compartida. La estimación usa el pp/tg que cada modelo dio en la prueba de humo sin contención.

Lectura:

- **Con el contrato «solo textos» y una consigna de reescritura, `qwen3-4b-2507` sí reformula** los planes de respaldo
  (solo copia el 11 % de sus pasos) sin perder una negrita ni meter un término nuevo, y sigue respetando los
  procedimientos ya pulidos (copia el 94 %). Es el único resultado que justificaría un LLM generador, y solo en la Mac.
- Pero **la calidad no sube**: 3,50 en los planes de respaldo, igual que copiando. Al reescribir corrigió «la stock» y
  ordenó el plan de importación de Excel, y a la vez escribió «Ingle al módulo» (una errata que ninguna regla del gate
  detecta) y convirtió «genera un pedido» (lo hace el sistema) en «genere un pedido» (lo haría la persona).
- `qwen3.5-2b` no aguanta la consigna de reescribir: **quita las `**` en el 86 % de los planes**, filtra la consigna a
  la intro («Este plan de respuesta validado instruye…») y en un plan inventó un paso en tú.
- En CPU, aun con 63 % menos tokens de salida, `qwen3.5-2b` necesita ~10,7 s por plan: solo leer el plan (~480 tokens a
  ~92 tok/s) ya son 5,2 s, más que todo el presupuesto.

## 6. CPU y EC2

### 6.1 Rendimiento bruto en 4 hilos de la M4 Pro (`llama-bench -ngl 0 -dev none -t 4 -p 512 -n 64 -r 1`)

| Modelo | pp (tok/s) | tg (tok/s) | Carga al medir | Nota |
|---|---|---|---|---|
| Jev 0.8B v3 (GGUF) | 152,4 | 57,4 | 17,8 | |
| qwen3.5-2b | 91,8 | 35,1 | 3,7 | limpio (antes, a las 02:20: 137 / 52) |
| qwen3-1.7b | 85,7 | 43,3 | 3,8 | limpio |
| granite-4.0-micro | 35,6 | 22,7 | 4,3 | limpio |
| qwen3-4b-2507 | 16,8 | 3,8 | 5,5 | **no fiable**: tg cuatro veces peor que lo que da su tamaño; swap. Por tamaño: ~30 / ~19 |
| qwen3.5-4b | 10,3 | 3,7 | 21,5 | **no fiable**. Por tamaño: ~40 / ~16 |
| qwen2.5-coder-7b | 9,6 | 7,4 | 15,8 | **no fiable**. Por tamaño: ~16 / ~10 |

«Por tamaño»: la generación en CPU está limitada por el ancho de banda de memoria (tg ∝ 1/GB del GGUF) y el prompt
por cómputo (pp ∝ 1/parámetros); se escaló desde `granite-4.0-micro` (denso) y `qwen3.5-2b` (híbrido), medidos con
carga baja. Es una estimación, no una medida.

### 6.2 Latencia estimada en 4 hilos de la Mac

`tokens_nuevos / pp + tokens_salida / tg`, con los tokens medianos de cada tarea (tipo: ~20 nuevos + 8 de salida;
evidencia: ~265 + 9; generación: ~480 + ~500):

| Modelo | Tipo | Evidencia | Generación (plan de 6 pasos) | Huella física en CPU (RSS) |
|---|---|---|---|---|
| qwen3.5-2b | **0,5 s** (medido 0,42) | 3,2 s (medido 2,4) | **19,6 s** (medido 16,6) | **1,6 GB** (RSS 2,8) |
| qwen3-1.7b | 0,4 s | 3,3 s | 17,5 s | 1,6 GB (RSS 2,5) |
| granite-4.0-micro | 0,8 s | 7,8 s | 33 s | 2,5 GB (RSS 3,9) |
| qwen3-4b-2507 | ~1,0 s | ~9 s | ~42 s | 3,1 GB (RSS 4,5) |
| qwen3.5-4b | ~1,0 s | ~7 s | ~36 s | 3,1 GB (RSS 4,9) |
| qwen2.5-coder-7b | ~1,3 s | ~17 s | **~80 s** | ~5 GB |

La huella física (`footprint`) es la memoria que no se puede devolver al sistema (KV, búferes, pesos reempaquetados);
el RSS suma además páginas del GGUF mapeado que el sistema puede soltar. En el EC2, planificar con el RSS.

### 6.3 Qué entra en el EC2 de 8 GB

- **Decisión de tipo con `qwen3.5-2b`**: ~0,5 s en la Mac, 2,8 GB de RSS en el peor caso, contexto de 4.096. Entra
  junto al servicio y al reranker (`bge-reranker-v2-m3` Q4_K_M, 0,44 GB de pesos). Con 16 GB no cambia la elección:
  un 4B entra en RAM, pero su ~1 s en la Mac pasa del presupuesto de 1,5 s en el EC2 (§6.4).
- **Suficiencia de evidencia con LLM: no.** Procesar ~265 tokens de fragmentos cuesta 2,4–9 s en CPU. Usar el puntaje
  del reranker con umbral calibrado y las reglas de §6 de la especificación; o Jev si se arregla su GGUF (§4.2).
- **Generación con LLM: no.** 17–80 s por plan, frente a 5 s de presupuesto.

### 6.4 De la Mac al EC2

No se midió en el EC2 (la tarea era solo en la Mac). Los 4 hilos de la M4 Pro son núcleos de rendimiento con mucho
ancho de banda de memoria; 4 vCPU de un EC2 x86 son 2 núcleos con hyper-threading y menos ancho de banda. Supuesto
razonable, **sin medir**: entre 1,5× y 3× más lento que estas cifras. Si la instancia es de la familia `t3`/`t3a`
(CPU con créditos), la inferencia sostenida además agota los créditos. **Antes de encender un modelo en el EC2:**
correr ahí `benchmark_modelos.py velocidad --modelo qwen3.5-2b --modo cpu4` y `correr --modelo qwen3.5-2b --modo cpu4
--solo-decision`.

## 7. Reproducir

Desde `s10-conocimiento/` (usa `.venv/bin/python3`; solo biblioteca estándar):

```bash
B=herramientas/benchmark_modelos.py
.venv/bin/python3 $B modelos                                   # qué GGUF están y dónde
.venv/bin/python3 $B correr --modelo qwen3.5-2b --modo metal   # decisión (con y sin gramática) + 36 planes
.venv/bin/python3 $B correr --modelo qwen3.5-2b --modo cpu4    # 90 tipo + 20 evidencia + 7 planes
.venv/bin/python3 $B correr --modelo qwen3.5-4b --modo cpu4 --latencia --solo-decision --sufijo __decision
.venv/bin/python3 $B correr --modelo qwen3-4b-2507 --modo metal --solo-generacion --variante reescribir --sufijo __reescribir
.venv/bin/python3 $B jev --url http://127.0.0.1:8765 --nombre jev-0.8b-v3-mlx --modo mlx
.venv/bin/python3 $B jev --gguf jeva.cpp/models/Jev-Style-0.8B-Decision-v3-Q4_K_M.gguf --modo cpu4 --nombre jev-0.8b-v3-gguf
.venv/bin/python3 $B velocidad --modelo qwen3.5-2b --modo cpu4 --tg 64 -r 1
.venv/bin/python3 $B ciego && .venv/bin/python3 $B resumen
```

Cada `correr` levanta **un** llama-server en `127.0.0.1:18431` (Jev en `:18432`) y lo apaga al terminar. Las líneas
exactas que usó:

```bash
# Metal
llama-server -m ~/.cache/metrin-modelos/Qwen3.5-2B-Q4_K_M.gguf --host 127.0.0.1 --port 18431 -c 4096 -np 1 \
  --cache-ram 0 --no-webui -ngl 99 --reasoning off
# CPU, 4 hilos, sin GPU
llama-server -m ~/.cache/metrin-modelos/Qwen3.5-2B-Q4_K_M.gguf --host 127.0.0.1 --port 18431 -c 4096 -np 1 \
  --cache-ram 0 --no-webui -ngl 0 -dev none -t 4 -tb 4 --reasoning off
# Jev GGUF (API /v1/systemone)
jeva.cpp/build/bin/llama-server -m jeva.cpp/models/Jev-Style-0.8B-Decision-v3-Q4_K_M.gguf --host 127.0.0.1 \
  --port 18432 -c 4096 -np 1 --cache-ram 0 --no-webui -ngl 0 -dev none -t 4 -tb 4
```

`--reasoning off` solo va con los Qwen3/Qwen3.5 híbridos (más `chat_template_kwargs: {"enable_thinking": false}` en
cada petición). `--cache-ram 0` evita que la caché de prompts en RAM (8 GB por defecto) infle el RSS; la caché de la
ranura sigue reutilizando el prompt de sistema.

Archivos:

- `metrin/eval/modelos/decision.jsonl`, `generacion.jsonl`: conjuntos (congelados).
- `metrin/eval/modelos/resultados/<modelo>__<modo>[__variante].json`: métricas y cada salida; `.log`: log del servidor;
  `velocidad.jsonl`: `llama-bench`.
- `metrin/eval/modelos/ciego/`: muestra a ciegas, clave, puntajes y notas.
- Modelos en `~/.cache/metrin-modelos/` (fuera del repo): 4 GGUF, **8,6 GB** (2 más se leyeron de los blobs de
  Ollama, sin copia).

## 8. Límites de esta medición

- Latencias contaminadas por contención (§3.5); se recomienda repetir `velocidad` y las corridas `cpu4` de los 4B con
  la máquina tranquila y, sobre todo, en el EC2.
- 90 + 40 + 36 casos: un acierto de diferencia en tipo son 1,1 puntos; en evidencia, 2,5. La diferencia entre
  `qwen3.5-4b` (98,9 %) y `qwen3-4b-2507` (91,1 %) en tipo sí es clara (7 casos); entre los 2B y el 1,7B, no tanto.
- Preguntas sintéticas escritas por quien las etiquetó; la etiqueta de los mensajes de solo tema (UNKNOWN) es una
  decisión de diseño discutible: con la laxa, 3 de esos 12 cuentan también como CONCEPT.
- La rúbrica de calidad la aplicó una sola persona (este agente), sobre 8 planes.
- El validador no ve erratas dentro de una palabra común («Ingle» por «Ingrese») ni cambios de sujeto («genera» →
  «genere»). Si algún día se enciende un LLM generador, el quality gate necesita además una comprobación por diferencias:
  que cada palabra nueva del paso exista en el plan o en un diccionario, y que los verbos no cambien de persona.
- `qwen2.5-coder-7b`: la corrida completa en Metal se cortó por una desconexión a las 04:51; se repitió solo la
  decisión (completa, 90 + 40); de generación hay prueba de humo (1 plan) y estimación de CPU.
