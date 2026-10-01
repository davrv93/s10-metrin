# Orquestación de Metrín

Estado: primera capa explícita implementada en `metrin/internal/rag`.

## Flujo de una consulta

```mermaid
flowchart TD
  A[Pregunta + hilo] --> B[Clasificador local en español]
  B -->|social / límite claro| C[Respuesta conversacional]
  B -->|trabajo / ayuda / duda / error| D[Ruta segura RAG]
  B -->|fallo o modelo incompatible| D
  D --> E[Identificar estilo: seguimiento, problema, comparación, procedimiento, concepto o dato]
  E --> F[Reescribir seguimiento con el hilo, si corresponde]
  F --> G[Buscar manuales oficiales S10]
  G --> H[Combinar búsqueda general + Cortex]
  H --> I[Seleccionar evidencia y abstenerse si no alcanza]
  I --> J[Prompt según tipo de consulta]
  J --> K[Respuesta + citas + traza de orquestación]
```

La similitud del clasificador es coseno, **no es una probabilidad calibrada**. El clasificador solo decide si es seguro evitar la búsqueda documental para charla social o un límite claro. No debe decidir hechos, escoger una respuesta técnica sin evidencia ni activar acciones del ERP. Si falla, si no está cargado o si cambia el fingerprint del embedding, se usa RAG.

## Capas y categorías

| Capa                     | Valores                                                                                                    | Responsable                                                 |
| ------------------------ | ---------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| Intención conversacional | `social`, `ayuda`, `limite`, `trabajo`                                                                     | kNN local sobre `potion-es-int8` + reglas léxicas de charla |
| Tipo de consulta técnica | `procedimiento`, `concepto`, `comparacion`, `problema`, `seguimiento`, `aclaracion`, `informacion_directa` | Reglas explícitas en español; no invoca otro modelo         |
| Ruta                     | `conversacion`, `rag`, `regla`                                                                             | Orquestador Go                                              |
| Forma final              | `conversacional`, `tutorial`, `respuesta`, `sin_contexto`, etc.                                            | RAG según la evidencia y el tipo de consulta                |

La API incluye `orquestacion` con intención, tipo de consulta, ruta, clasificador y similitud observada para poder revisar decisiones. El campo `modo` conserva su significado previo para no romper clientes.

## Política de fuentes

1. Cada consulta elegible busca manuales oficiales S10 (`confianza=oficial`) como conjunto propio; así una sección no desaparece solo por quedar fuera del top-K general.
2. Esos pasajes se combinan con resultados generales y Cortex. La autoridad desempata, pero una fuente no oficial mucho más relevante puede ganar; no se oculta el origen ni se inventa una respuesta.
3. Solo una pregunta procedimental con evidencia de tutorial o de una sección HTML oficial recibe el prompt extenso de pasos. Los conceptos no se convierten automáticamente en tutoriales.
4. Sin evidencia suficiente, el sistema se abstiene y conserva la consulta para revisión.

## Estado del clasificador y del entrenamiento

El artefacto actual no es un LLM pequeño afinado: es kNN sobre embeddings españoles locales, empaquetado en Go. Contiene 115 ejemplos en cuatro clases. Su evaluación documentada (validación 22/28, prueba 23/28) tiene apenas 28 textos por split; los fallos medidos cayeron a RAG, pero la muestra es demasiado pequeña para afirmar alta confiabilidad.

`data/preguntas/preguntas.jsonl` contiene 14.000 preguntas generadas desde Cortex con tipos `concepto`, `procedimiento`, `problema`, `dato`, `comparacion` y `contexto`; es útil para un experimento offline de tipo de consulta, pero **no son 14.000 mensajes reales de usuarios**. Las 56.000 variantes ortográficas sirven al retrieval, no deben contarse como conversaciones independientes. No se debe entrenar el LoRA conversacional con manuales ni usar pérdida de entrenamiento como prueba de exactitud.

**Decisión actual:** no hacer fine-tuning todavía. Primero ampliar la evaluación, separar por manual/nodo para evitar fuga de paráfrasis, probar falsos positivos que sacarían consultas técnicas del RAG, y recoger una muestra de preguntas reales anonimizadas y revisadas. Si esa evaluación demuestra confusiones sistemáticas en los tipos técnicos, entonces se entrena un segundo clasificador local sobre las etiquetas curadas del banco y se compara contra las reglas actuales. La decisión solo se promueve si mejora métricas por clase sin desviar preguntas técnicas.

## Calidad de la base oficial HTML observada

Reindexé el material descargado sin red y reconstruí el Metrín local. La KB
quedó con 1.314 filas JSONL: 466 fragmentos HTML por sección; 172 secciones
oficiales de `documentacion.s10peru.com`; 51 secciones oficiales incluyen
pasos con captura y contienen 153 referencias de imagen. Las 268 referencias
de captura revisadas apuntan a archivos existentes. El servidor local cargó
6.093 trozos.

Hay 60 HTML huérfanos con la marca del muro de miembros, pero al cruzarlos con
`data/paginas.jsonl` ninguno corresponde a una fila indexada. El indexador
omite páginas marcadas con `muro_de_miembros`; esa protección se conserva.
Sigue pendiente un caso importante: el HTML guardado de `Manual de Gerencia de
Proyectos` contiene el armazón del sitio y no el cuerpo del manual; su `.md` sí
contiene el texto extenso, pero todavía se parte por tamaño y no por sección.
No contar ese manual como cubierto por el nuevo filtro por secciones.

## Decisiones JEV y trazas

Además de la orquestación anterior, Metrín pide **decisiones evaluativas** a
jeva.cpp (`POST /v1/systemone`, llama-server con la API JEV): choice, score y
noul se resuelven desde los logits durante el prefill, sin generar tokens.
Las decisiones no cambian la respuesta que ya se dio: son para revisar
*a posteriori* si la conducta fue correcta. Sin `JEV_URL` el flujo de siempre
no cambia (todo es opt-in).

| Dónde | Config | Pregunta (noul) | Origen de la traza |
| ----- | ------ | --------------- | ------------------ |
| Reportes (`internal/reportes`) | `JEV_ORIGEN` (defecto `metrin`) | «¿Es seguro ejecutar este reporte contra la base de producción?» | `reportes` |
| Chat ask/serve (`internal/rag`) | `JEV_ORIGEN_CHAT` (defecto `metrin-chat`) | respaldada por el contexto / abstención correcta / charla correcta | `metrin-chat` |

En el chat se emiten tres trazas por turno según la ruta: (1) respuesta
normal — «¿está respaldada únicamente por el contexto recuperado y la
pregunta, sin inventar datos, acciones ni fuentes?»; (2) abstención por falta
de contexto; (3) charla conversacional — «¿responder con una charla breve sin
buscar en los manuales fue la decisión correcta?». El estado de la traza lleva
pregunta, modo, tipo de consulta, ruta, fuentes (citas recortadas) y la
respuesta recortada a 500 caracteres. Si la petición a JEV o el registro
fallan, la traza se pierde pero el chat sigue intacto.

Cada decisión exitosa se graba en un JSONL (`JEV_TRAZAS`, una traza por
línea: id ULID, ts, modelo, origen, plantilla, ms, tokens, estado, preguntas
con sus respuestas y `correcta` = veredicto de muestra etiquetada, `null` si
no se evaluó). El emisor es `metrin/internal/jev/trazas.go`.

### Visor de trazas (`trazas/trazas.py`, puerto 4761)

La capa de datos es una abstracción (`RepositorioTrazas`); el API no cambia
según el origen. `TRAZAS_REPO` elige:

- `mock` (defecto): `trazas/mock/trazas.json`.
- `archivo[:ruta]`: el JSONL que emite Metrín (revalidación por mtime).
- `postgres`: tablas `traces` + `jev` de `trazas/esquema.sql` (requiere
  `TRAZAS_PG_DSN`; psycopg se importa al conectar).

### Postgres local de trazas

El compose levanta `trazas-pg` (postgres:16-alpine, host `127.0.0.1:5433`,
usuario/contraseña `metrin`/`metrin`, base `trazas`). `esquema.sql` se aplica
solo en el primer arranque (volumen `trazas-pg-data`). Para probar el visor
contra Postgres:

```bash
cd s10-conocimiento
# .env (local, ignorado): TRAZAS_REPO=postgres y
# TRAZAS_PG_DSN=postgresql://metrin:metrin@trazas-pg:5432/trazas
docker compose up -d trazas-pg trazas
# sembrar con los mocks (o un JSONL de metrin): upsert por id
docker compose exec trazas python trazas/cargar.py
docker compose exec trazas-pg psql -U metrin -d trazas -c 'SELECT tipo, count(*) FROM jev GROUP BY tipo;'
curl -s http://127.0.0.1:4761/api/salud      # {"ok":true,"origen":"postgres"}
curl -s http://127.0.0.1:4761/api/estadisticas
```

`trazas/cargar.py` carga `mock/trazas.json` o un JSONL de Metrín en las dos
tablas (upsert por `id`; recargar no duplica). El visor también arranca sin
psycopg: solo `TRAZAS_REPO=postgres` lo necesita (añadido a
`requirements-servicios.txt`).

## Harness de evaluación (`rag eval`)

Suite fija de casos contra el flujo completo —clasificador, búsqueda, LLM y
JEV— para medir si Metrín responde o se abstiene donde debe. No es un test
unitario: cada caso corre el pipeline real y el acierto exige **dos cosas**:
la forma final del turno coincide con lo esperado (respondió/abstuvo) **y**
la decisión de JEV avala la conducta (`noul ≥ umbral`, defecto 0.5).

```bash
cd s10-conocimiento/metrin
rag eval                            # suite fija (13 casos)
rag eval --casos casos.jsonl --k 8 --umbral 0.5
rag eval --json informe.json        # informe JSON para el gate de CI
```

- **Casos** (`internal/eval/casos.go`): 13 curados sobre el corpus real —
  7 «responder» (qué es S10 y para quién, quién lo fabrica, UIT/RMV y
  porcentajes de AFP en Nóminas, reporte REP-02, pantalla Constantes por
  Fecha General, política SGSST) y 6 «abstener» (vuelos, ceviche, Brent,
  ventilador de GPU *con la palabra «Nóminas»* como distractor, feriados en
  Japón, poda de árbol en Gerencia de Proyectos). Un JSONL propio lleva una
  línea por caso: `{"pregunta":…, "esperado":"responder"|"abstener", "nota":…}`
  (`#` abre comentarios).
- **Trazas del harness**: se graban en su propio JSONL (`--trazas`, defecto
  `datos/eval-trazas.jsonl`) con origen `metrin-eval` (`--origen`),
  separadas de las del chat. Son trazas JEV normales: el visor las lee con
  `TRAZAS_REPO=archivo` y `trazas/cargar.py` las carga en Postgres.
- **Informe**: por caso ✓/✗ con noul, confianza y latencia; al final,
  totales, aciertos, confianza media y latencia media y p95. Un caso sin
  traza nueva cuenta como `SinTraza`, nunca como acierto. Los fallos de
  veredicto son datos del informe, no errores del comando: el gate de CI
  lee `--json`.
- **Gate**: `JEV_URL` vacío o JEV caído fallan pronto (el `pingJEV` previo
  comprueba el servidor con una decisión trivial, sin grabar traza).

La abstención de los «abstener» **no** viene del umbral de distancia
(recuperan pasajes a d≈0.46–0.74 < 0.80): viene del camino 2 —el LLM marca
«No tengo contexto suficiente» ante contexto irrelevante (`MarcaSinContexto`).
Por eso el harness necesita LLM y JEV vivos.

Verificado en vivo (2026-10-01) con el LLM mlx del Mac (`LLM_PROVIDER=mlx`,
`OLLAMA_MODEL=mlx-community/Qwen2.5-3B-Instruct-4bit`): «¿Qué es S10?»
responde; «¿Qué aerolíneas vuelan desde Lima a Bogotá?» y el caso trampa del
ventilador de GPU abstienen vía `MarcaSinContexto`. Los tests sin servicios
vivos cubren veredictos, parseo de trazas y resumen: `go test ./internal/eval/...`.

## Verificación

Los cambios de orquestación pasan `go test ./...` en `metrin/` y
`.venv/bin/python -m unittest trazas.test_trazas` en la raíz. Para comprobar
la conducta en una imagen local, reconstruir el servicio Metrín después de que
se regenere la KB y confirmar `/health`, `orquestacion`, citas oficiales y
respuestas con pasos junto a una captura conocida. La sesión de consultas es
read-only por conexión (`_query_only=1` en SQLite,
`default_transaction_read_only` en PostgreSQL): ni un INSERT accidental entra.
Nota: `modernc.org/sqlite` ignora `mode=ro`; su interruptor de solo-lectura es
`_query_only`.
