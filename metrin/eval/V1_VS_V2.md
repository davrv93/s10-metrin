# Metrín · Benchmark V1 frente a V2

Generado por `go run ./cmd/evalv2` el 06-10-2026 05:40 -05. **No se edita a mano**: se regenera en cada corrida (el JSON crudo de
cada corrida queda en `eval/resultados/`). Especificación: `docs/V2-RAG-PROCEDURAL.md` §11–12.

| | |
|---|---|
| Servicio | `http://127.0.0.1:4762` · `/health` → `{"ok":true,"traza":true,"trozos":15277,"v2":{"busqueda":{"fragmentos":8135,"modo":"hibrida","rerank_max_runas":800,"rerank_top_n":20,"reranker":true,"vector":true},"habilitada":true,"motor":"reglas","porcentaje":0,"version":"v1"}}` |
| Contenedor (RAM/CPU) | `metrin-traza-prueba` |
| Dataset | `metrin/eval/v2_oro.jsonl`: 194 casos (CONCEPT 39 · PROCEDURE 60 · NAVIGATION 24 · TROUBLESHOOTING 31 · CONFIGURATION 21 · AMBIGUOUS 19) |
| KB al correr | 8135 fragmentos · 60 procedimientos · 50 conceptos |
| Versiones | V1: corrida en 2114 s (0 turnos con `plan`) · V2: corrida en 46 s (212 turnos con `plan`) |
| JSON crudo | `metrin/eval/resultados/2026-10-06T054011.json` |

## Resultado

Regla de §12: **V2 se enciende solo si supera a V1 en PROCEDURAL ANSWER SUCCESS sin empeorar la tasa de invención.**

V2 **cumple** la regla: PAS 60.0 % frente a 0.0 %, invención 0.0 % frente a 17.1 %. (Con 55 casos PAS, una diferencia de un caso es 1.8 puntos: mire el detalle antes de encenderla.)

| Métrica | V1 | V2 |
|---|---|---|
| **PROCEDURAL ANSWER SUCCESS** | **0.0 % (0/55)** | **60.0 % (33/55)** |
| C1 identifica el procedimiento | 25.5 % (14/55) | 63.6 % (35/55) |
| C2 pasos correctos | 0.0 % (0/55) | 60.0 % (33/55) |
| C3 conserva el orden | 47.3 % (26/55) | 63.6 % (35/55) |
| C4 no inventa pasos | 74.5 % (41/55) | 100.0 % (55/55) |
| C5 asigna bien las fotos | 9.1 % (5/55) | 63.6 % (35/55) |
| C6 responde la intención | 80.0 % (44/55) | 69.1 % (38/55) |
| C7 fuentes respaldan los pasos | 25.5 % (14/55) | 63.6 % (35/55) |
| Éxito de la tarea (todas las categorías) | 5.7 % (11/194) | 63.9 % (124/194) |
| Precisión de clasificación | 37.6 % (73/194) | 73.2 % (142/194) |
| Clasificación comparable | 56.2 % (73/130) | 73.8 % (96/130) |
| Recall@5 | 0.141 (n=171) | 0.494 (n=171) |
| Recall@10 | 0.191 (n=171) | 0.536 (n=171) |
| MRR@10 | 0.163 (n=171) | 0.557 (n=171) |
| nDCG@10 | 0.130 (n=171) | 0.507 (n=171) |
| Acierto de procedimiento | 34.4 % (43/125) | 63.2 % (79/125) |
| Acierto de concepto | 33.3 % (12/36) | 75.0 % (27/36) |
| Recall de pasos | 0.154 (n=55) | 0.606 (n=55) |
| Recall de imágenes | 13.3 % (44/331) | 40.4 % (76/188) |
| Exactitud paso ↔ foto | 25.0 % (8/32) | 100.0 % (81/81) |
| Tasa de invención (turnos con contenido) | 17.1 % (30/175) | 0.0 % (0/174) |
| Tasa de invención por paso | 9.8 % (55/563) | 0.0 % (0/204) |
| Tasa de respaldo | 0.5 % (1/212) | 0.0 % (0/212) |
| Tasa de SIN_EVIDENCIA | 8.5 % (18/212) | 9.9 % (21/212) |
| Abstención correcta | 25.0 % (3/12) | 100.0 % (12/12) |
| Falsa abstención | 8.0 % (13/163) | 5.5 % (9/163) |
| Éxito de generación | 99.5 % (211/212) | 100.0 % (212/212) |
| Latencia cliente p50 / p95 (ms) | 7147 / 28141 | 8 / 1665 |
| Latencia servidor (traza) p50 / p95 (ms) | 7138.9 / 28100.9 | 3.3 / 1659.1 |
| Errores HTTP o de red | 0 | 0 |
| RAM y CPU V1 | CPU media 10.2 % · máx 584.9 % · RAM media 249 MiB · máx 272 MiB (419 muestras) | |
| RAM y CPU V2 | CPU media 10.5 % · máx 43.2 % · RAM media 274 MiB · máx 276 MiB (10 muestras) | |

## V1 por categoría

| Categoría | Casos | Clasificación | Recall@5 | Recall@10 | MRR | nDCG@10 | Éxito tarea | PAS | Invención | Respaldo | SIN_EVID. | p50 / p95 ms |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| CONCEPT | 39 | 56 % | 0.14 | 0.19 | 0.20 | 0.14 | 13 % | — | 3 % | 0 % | 5 % | 17521 / 37688 |
| PROCEDURE | 60 | 82 % | 0.05 | 0.12 | 0.13 | 0.08 | 0 % | 0 % | 29 % | 0 % | 3 % | 6092 / 12255 |
| NAVIGATION | 24 | 0 % | 0.22 | 0.26 | 0.22 | 0.20 | 0 % | — | 21 % | 0 % | 0 % | 9113 / 19727 |
| TROUBLESHOOTING | 31 | 6 % | 0.21 | 0.27 | 0.16 | 0.16 | 6 % | — | 6 % | 3 % | 32 % | 6816 / 27776 |
| CONFIGURATION | 21 | 0 % | 0.21 | 0.25 | 0.13 | 0.14 | 0 % | — | 24 % | 0 % | 0 % | 5306 / 9675 |
| AMBIGUOUS | 19 | 0 % | 0.11 | 0.11 | 0.20 | 0.11 | 21 % | — | 6 % | 0 % | 11 % | 4717 / 9538 |
| **Global** | 194 | 38 % | 0.14 | 0.19 | 0.16 | 0.13 | 6 % | 0 % | 17 % | 0 % | 8 % | 7147 / 28141 |

Observaciones de V1 (automáticas):

- Modos de respuesta en 212 turnos: `respuesta` 127, `tutorial` 80, `conversacional` 4, `modelo_no_disponible` 1.
- Motivos declarados por el servicio: `el modelo dijo que el contexto no alcanza` 18, `el modelo no pudo responder en español` 1.
- Tipo crudo devuelto: `procedimiento` 91, `informacion_directa` 85, `concepto` 29, `social` 4, `problema` 2, `comparacion` 1.
- Turnos con pasos numerados: 94 de 212.
- Citas que no salen de kb/*.jsonl (no resolubles a un fragmento): 79 de 1160 (7 %).
- Casos PAS que fallan cada condición (de 55): C1 41 · C2 55 · C3 29 · C4 14 · C5 50 · C6 11 · C7 41.

## V2 por categoría

| Categoría | Casos | Clasificación | Recall@5 | Recall@10 | MRR | nDCG@10 | Éxito tarea | PAS | Invención | Respaldo | SIN_EVID. | p50 / p95 ms |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| CONCEPT | 39 | 82 % | 0.75 | 0.75 | 0.76 | 0.75 | 77 % | — | 0 % | 0 % | 12 % | 5 / 87 |
| PROCEDURE | 60 | 78 % | 0.40 | 0.49 | 0.61 | 0.47 | 60 % | 60 % | 0 % | 0 % | 10 % | 11 / 1827 |
| NAVIGATION | 24 | 83 % | 0.71 | 0.71 | 0.74 | 0.71 | 67 % | — | 0 % | 0 % | 4 % | 7 / 31 |
| TROUBLESHOOTING | 31 | 55 % | 0.41 | 0.45 | 0.38 | 0.39 | 55 % | — | 0 % | 0 % | 11 % | 12 / 165 |
| CONFIGURATION | 21 | 38 % | 0.36 | 0.40 | 0.30 | 0.31 | 33 % | — | 0 % | 0 % | 14 % | 12 / 1424 |
| AMBIGUOUS | 19 | 95 % | 0.00 | 0.00 | 0.00 | 0.00 | 95 % | — | 0 % | 0 % | 5 % | 7 / 1457 |
| **Global** | 194 | 73 % | 0.49 | 0.54 | 0.56 | 0.51 | 64 % | 60 % | 0 % | 0 % | 10 % | 8 / 1665 |

Observaciones de V2 (automáticas):

- Modos de respuesta en 212 turnos: `respuesta` 137, `aclaracion` 54, `sin_contexto` 21.
- Motivos declarados por el servicio: `degradado: sin_vector: búsqueda léxica; sin_reranker: puntaje híbrido sin reordenar` 140, `tipo de respuesta desconocido: se pide aclaración` 40, `el constructor no halló evidencia que respalde un plan entre 5 candidato(s) · degradado: sin_vector: búsqueda léxica; sin_reranker: puntaje híbrido sin reordenar` 14, `continuación sin procedimiento en curso` 1, `el constructor no halló evidencia que respalde un plan entre 1 candidato(s) · degradado: sin_vector: búsqueda léxica; sin_reranker: puntaje híbrido sin reordenar` 1, `el constructor no halló evidencia que respalde un plan entre 2 candidato(s) · degradado: sin_vector: búsqueda léxica; sin_reranker: puntaje híbrido sin reordenar` 1, `el constructor no halló evidencia que respalde un plan entre 3 candidato(s) · degradado: sin_vector: búsqueda léxica; sin_reranker: puntaje híbrido sin reordenar` 1, `el constructor no halló evidencia que respalde un plan entre 4 candidato(s) · degradado: sin_vector: búsqueda léxica; sin_reranker: puntaje híbrido sin reordenar` 1, `el constructor no halló evidencia que respalde un plan entre 6 candidato(s) · degradado: sin_vector: búsqueda léxica; sin_reranker: puntaje híbrido sin reordenar` 1, `el constructor no halló evidencia que respalde un plan entre 8 candidato(s) · degradado: sin_vector: búsqueda léxica; sin_reranker: puntaje híbrido sin reordenar` 1, `evidencia insuficiente tras 2 búsqueda(s) (límite MAX_SEARCH_ITERATIONS=2 alcanzado) · degradado: sin_vector: búsqueda léxica; sin_reranker: puntaje híbrido sin reordenar` 1, `pregunta elíptica sin nada a qué referirse: se pide aclaración` 1.
- Tipo crudo devuelto: `PROCEDURE` 74, `UNKNOWN` 54, `CONCEPT` 34, `NAVIGATION` 24, `TROUBLESHOOTING` 17, `CONFIGURATION` 9.
- Turnos con pasos numerados: 95 de 212.
- Citas que no salen de kb/*.jsonl (no resolubles a un fragmento): 16 de 763 (2 %).
- Casos PAS que fallan cada condición (de 55): C1 20 · C2 22 · C3 20 · C4 0 · C5 20 · C6 17 · C7 20.
- Casos PAS con éxito: proc-001, proc-002, proc-006, proc-008, proc-010, proc-013, proc-016, proc-019, proc-020, proc-021, proc-023, proc-025, proc-026, proc-028, proc-030, proc-032, proc-033, proc-035, proc-036, proc-037, proc-039, proc-040, proc-041, proc-043, proc-044, proc-045, proc-047, proc-048, proc-049, proc-056, proc-057, proc-059, proc-060.

## Definiciones de las métricas

Todas las métricas de calidad se calculan sobre el **último turno** de cada caso (los anteriores son contexto). Las
operativas (latencia, respaldo, SIN_EVIDENCIA, generación) cuentan **todos** los turnos enviados. El mismo código
califica a V1 y a V2: solo cambia de dónde sale cada dato (ver «Cómo se evalúa V1»).

| Métrica | Definición exacta | Casos que cuentan |
|---|---|---|
| **PROCEDURAL ANSWER SUCCESS (PAS)** | Proporción de casos que cumplen **las siete** condiciones de abajo a la vez | `tipo_esperado = PROCEDURE`, con `procedimiento_esperado` en kb/procedimientos y sin `sin_evidencia_esperada`. Se excluye una continuación sin pasos pendientes |
| C1 · identifica el procedimiento | V2: `plan.procedimiento.id` = esperado. V1: cita ≥ 1 fragmento de la tabla `fuentes` del YAML **y** su texto cubre ≥ 1 paso del YAML (en casos no PROCEDURE con procedimiento de anclaje basta la cita) | PAS |
| C2 · pasos correctos | Todos los pasos esperados del turno están entre los entregados (recall de pasos = 1). Pasos esperados = `pasos_esperados` (null = todos los de primer nivel); en una continuación, solo los que siguen al último paso que la MISMA versión entregó en el turno anterior; si el plan declara entrega por partes (`pasos_mostrados`), la parte mostrada debe ser el comienzo de lo esperado y tener ≥ min(esperados, 3) pasos | PAS |
| C3 · conserva el orden | ≥ 1 paso entregado y: con ids (V2), los n crecen estrictamente en el orden de la respuesta (sin repetidos); sin ids (V1), existe una alineación monótona paso → línea (recorriendo los pasos cubiertos de menor a mayor n, cada uno cabe en una línea que lo cubre en o después de la del anterior; varios pasos pueden caer en la misma línea) | PAS |
| C4 · no inventa pasos | 0 pasos inventados y 0 negritas inventadas (ver «Tasa de invención»). Es la única condición que se cumple sin pasos: una respuesta de respaldo no inventa nada | PAS |
| C5 · asigna bien las fotos | ≥ 1 paso del procedimiento entregado, ninguna foto colocada en un paso del procedimiento es de otro paso (según el YAML), ninguna foto sin evidencia, y si los pasos entregados tienen `fotos_esperadas`, ≥ 50 % de ellas está en su paso | PAS |
| C6 · responde la intención | Sin error, sin respaldo, sin SIN_EVIDENCIA, ≥ 1 paso entregado y tipo predicho PROCEDURE (o uno de `tipos_aceptables`) | PAS |
| C7 · las fuentes respaldan los pasos | ≥ 1 paso del procedimiento entregado, lo citado incluye ≥ 1 fragmento de la tabla `fuentes` del YAML y cada paso entregado que corresponde a un paso esperado tiene respaldo: V2, su `fuente` incluye una del paso del YAML o su texto tiene soporte léxico en lo citado; V1, soporte léxico de la línea en el texto de los fragmentos citados | PAS |
| Precisión de clasificación | tipo predicho = `tipo_esperado` (o está en `tipos_aceptables`). V2: `plan.tipo`; V1: `orquestacion.tipo_consulta` traducido (procedimiento→PROCEDURE, concepto→CONCEPT, problema→TROUBLESHOOTING, comparacion→COMPARISON, social/fuera_de_alcance/ruta conversación→SOCIAL; informacion_directa, aclaracion y seguimiento no tienen equivalente y cuentan como fallo) | Todos |
| Clasificación comparable | La misma, solo con tipos que V1 sabe expresar (CONCEPT, PROCEDURE, TROUBLESHOOTING, COMPARISON, SOCIAL) | Esos tipos |
| Recall@k (k = 5, 10) | Entradas de `fragmentos_relevantes` acertadas por los k primeros ítems del ranking ÷ entradas. Un ítem acierta una entrada si representa alguno de sus fragmentos (par id@manual) | Casos con fragmentos relevantes |
| MRR@10 | 1 / posición del primer ítem que acierta alguna entrada (0 si ninguno en los 10 primeros); media | Ídem |
| nDCG@10 | Σ g_i / log2(i+1) ÷ Σ_{i ≤ min(\|R\|,10)} 1 / log2(i+1); g_i = 1 si el ítem i acierta una entrada aún no acertada (binaria, sin dobles cuentas); media | Ídem |
| Acierto de procedimiento | C1 | Casos con procedimiento_esperado |
| Acierto de concepto | V2: `plan.concepto.id` = el de kb/conceptos. V1 (o concepto aún sin YAML): respuesta con contenido que cita ≥ 1 fragmento del concepto o de `fragmentos_relevantes` | Casos con concepto_esperado |
| Recall de pasos | Pasos esperados del turno entregados ÷ pasos esperados del turno; media | Casos PROCEDURE con procedimiento |
| Recall de imágenes | Σ fotos esperadas (de los pasos esperados del turno) presentes en la respuesta, en cualquier lugar —incluida la galería de `fuentes[].fotos`, que la página muestra aunque el texto sea de respaldo— ÷ Σ fotos esperadas (micro) | Casos PROCEDURE con fotos esperadas |
| Exactitud paso ↔ foto | Σ fotos colocadas en un paso del procedimiento esperado que el YAML asocia a ese paso ÷ Σ fotos colocadas en pasos del procedimiento esperado (micro). Las fotos de galería (sin paso) no cuentan | Casos con alguna foto colocada |
| Tasa de invención | Turnos con ≥ 1 invención ÷ turnos con contenido (sin error, sin respaldo, sin SIN_EVIDENCIA). Invención = (a) paso entregado que no corresponde a ningún paso del procedimiento esperado **y** no tiene soporte léxico en los fragmentos citados; (b) término entre `** **` que no aparece en lo citado (ni en el YAML del procedimiento que el plan dice usar); (c) foto que no está en ningún fragmento citado, ni la ata el servidor a una fuente, ni está en ningún YAML | Turnos con contenido |
| Tasa de invención por paso | Pasos inventados ÷ pasos entregados (micro) | Casos con pasos |
| Tasa de respaldo | Turnos cuyo modo o motivo dice «no disponible», «respaldo» o «segura», o con una etapa de la traza en estado `respaldo` ÷ turnos | Todos los turnos |
| Tasa de SIN_EVIDENCIA | Turnos con `sin_contexto` (V1) o `plan.sin_evidencia` (V2) ÷ turnos | Todos los turnos |
| Abstención correcta | El sistema dice SIN_EVIDENCIA o pide aclaración, sin pasos y sin invención | Casos con sin_evidencia_esperada |
| Falsa abstención | SIN_EVIDENCIA en un caso que sí tiene respuesta | Casos con respuesta esperada |
| Éxito de generación | HTTP 200, `respuesta` no vacía y sin respaldo ÷ turnos | Todos los turnos |
| Éxito de la tarea | Si el caso admite `UNKNOWN` en `tipos_aceptables`, pedir aclaración sin inventar siempre es éxito. PROCEDURE con YAML: PAS. CONCEPT: contenido sin invención, tipo correcto, acierto de concepto y `terminos_esperados` presentes. NAVIGATION, TROUBLESHOOTING, CONFIGURATION y PROCEDURE sin YAML («implícito»): contenido sin invención, tipo correcto, cita ≥ 1 fragmento relevante y términos presentes (PROCEDURE además ≥ 1 paso). AMBIGUOUS (UNKNOWN): pide aclaración sin inventar, o responde con un tipo de `tipos_aceptables` citando lo relevante. Sin evidencia esperada: abstención correcta | Todos |
| Latencia p50 / p95 | Percentil por rango más cercano del tiempo de pared de cada petición (cliente) y de `traza.total_ms` (servidor) | Todos los turnos |
| RAM y CPU | `docker stats --no-stream` del contenedor cada ~3 s mientras corre cada versión: media y máximo | Con --contenedor |

**Soporte léxico** (el criterio de herramientas/validar_procedimientos.py): palabras de contenido del texto (minúsculas,
sin tildes, ≥ 4 letras, sin palabras vacías ni verbos genéricos de interfaz) recortadas a 5 letras; pasa si comparte
≥ min(2, n) de ellas y ≥ 60 % con el texto de referencia. Aquí además se ignoran las formas de «tú» de esos verbos
(«pulsa», «elige», «ingresa»…), porque V1 tutea y el YAML no.

## Cómo se evalúa V1 (sin `plan`)

V1 no devuelve `plan`, así que lo procedural se infiere de `respuesta`, `fuentes` y la traza:

- **Pasos entregados**: líneas numeradas de `respuesta` («1.», «2)», «Paso 3:»); si no hay, viñetas. Una respuesta en prosa
  entrega 0 pasos.
- **Paso entregado ↔ paso del YAML**: una línea cubre un paso (o subpaso) si comparte ≥ min(2, n) raíces y ≥ 50 % de las
  raíces del paso, o si contiene como frase completa una negrita del paso que lo identifica: no genérica (no «Aceptar»,
  «Nuevo») y exclusiva de ese paso dentro del procedimiento («Datos Generales», que está en los pasos 1 y 13 de
  registrar-presupuesto-nuevo, no identifica a ninguno). Una línea puede cubrir varios pasos (V1 los funde).
- **Orden**: se conserva si existe una alineación monótona paso → línea (ver C3).
- **Procedimiento identificado**: cita ≥ 1 fragmento de la tabla `fuentes` del YAML y cubre ≥ 1 paso.
- **Fuentes citadas**: `fuentes[].cita` → fragmentos de kb/*.jsonl cuya cita (título + «, p. N») es esa, igual que la arma
  internal/indexar/kb_jsonl.go. Las citas que no salen de kb/*.jsonl (el índice del contenedor tiene más trozos que la
  KB) no se pueden resolver: cuentan como no relevantes y su texto no respalda nada. Tampoco se resuelve una cita
  genérica que la KB da a más de 3 fragmentos (p. ej. «S10 Conocimiento», 156 nodos de Cortex).
- **Fotos**: `![…](fotos/<ruta>)` debajo de una línea = foto de ese paso (lo que hace colocarFotos); las de
  `fuentes[].fotos` son de galería: cuentan para el recall de imágenes, no para la exactitud paso ↔ foto. Se comparan
  por `sha1(ruta)[:12]`, el id del ESQUEMA.
- **Ranking para Recall@k/MRR/nDCG**: `fuentes` en su orden y, detrás, `traza.seleccion.descartados` (los mejores candidatos no
  elegidos, en orden), sin repetir cita. V1 entrega como mucho 5 fuentes, así que Recall@10 depende de los descartados.

**Límites de esta heurística (honestos):**

1. Es léxica: una línea que parafrasea mucho no cubre su paso (falso negativo) y una que comparte el nombre de un menú
   con otro paso puede cubrirlo (falso positivo). Los umbrales no se calibraron contra juicio humano.
2. Es **indulgente con V1** donde no se puede distinguir: un paso que no es del procedimiento pero tiene soporte en lo
   citado no es «inventado» (solo «ajeno»); el procedimiento se da por identificado con 1 sola fuente del YAML; la
   pregunta de aclaración se reconoce por un «?» al final. Si V2 supera a V1 aun así, la ventaja es una cota inferior.
3. V1 cita a nivel de respuesta, no de paso: la condición 7 comprueba que la línea tenga soporte en **algo** de lo
   citado, no en la fuente de ese paso.
4. Las invenciones dentro de un paso bien mapeado (un número, un campo, un orden de clics) solo se detectan si van en
   negrita. V1 casi no usa negritas, así que su tasa de invención queda subestimada.
5. V1 no tiene NAVIGATION, CONFIGURATION ni UNKNOWN: su precisión de clasificación en esas categorías es 0 por
   construcción; por eso se informa también la «clasificación comparable».
6. Si el servicio no tiene modelo de lenguaje (texto de respaldo), V1 no entrega pasos salvo en modo tutorial: la línea
   base mide V1 **tal como está desplegado en el contenedor evaluado**, no su techo con modelo.

## Cómo se evalúa V2

Con `plan` (contrato `internal/v2/tipos`): tipo = `plan.tipo`; procedimiento = `plan.procedimiento.id`; pasos =
`plan.pasos` (o solo `plan.pasos_mostrados` si entrega por partes), cada uno con su `id` (`<proc>#n`), sus `fotos` y
su `fuente`; ranking = candidatos de la traza (`rerank` → `fusion` → `vectorial`/`lexica`, claves `candidatos`/`top`/
`resultados`), o lo citado en orden si la traza no los trae. Un candidato de clase `procedimiento` representa a los
fragmentos de su tabla `fuentes`. Memoria: el arnés reenvía en `memoria` lo que la respuesta anterior trajo en
`memoria` (supuesto: así la guardará la página). Un turno V2 sin `plan` (p. ej. un saludo por la ruta
conversacional) se califica con la heurística de V1. Si ninguna pregunta de sondeo devuelve `plan`, V2 queda
«no disponible» y no se corre.

## Dataset

`eval/v2_oro.jsonl`, sintético (`sintetico: true` en todos los casos), generado por `eval/construir_v2_oro.py`: las
preguntas están escritas a mano en ese script y las expectativas se DERIVAN de kb/ al construir (reglas de etiquetado
en su cabecera; cada caso trae su `justificacion`). El script rechaza preguntas iguales a `preguntas`/`aliases` de un
YAML o a un ejemplo de entrenamiento de kb/catalogos, para no medir memoria de V2.

| Categoría | Casos | Con procedimiento | Sin evidencia esperada | Conversaciones (2–3 turnos) | Variantes |
|---|---|---|---|---|---|
| CONCEPT | 39 | 0 | 3 | 3 | abreviatura 2, cambio_de_tema 2, coloquial 1, conversacional 3, corta 6, obligatorio 1, referencia_conversacional 1, sin_evidencia 3, sinonimo 2 |
| PROCEDURE | 60 | 55 | 3 | 6 | coloquial 14, continuacion 3, conversacional 6, corta 8, falta_ortografica 2, larga 1, obligatorio 3, referencia_conversacional 3, sin_evidencia 3 |
| NAVIGATION | 24 | 23 | 1 | 1 | abreviatura 1, coloquial 2, conversacional 1, corta 2, obligatorio 1, referencia_conversacional 1, sin_evidencia 1 |
| TROUBLESHOOTING | 31 | 28 | 3 | 5 | coloquial 2, conversacional 5, corta 1, no_me_sale 5, obligatorio 1, sin_evidencia 3 |
| CONFIGURATION | 21 | 19 | 2 | 1 | coloquial 1, conversacional 1, corta 1, referencia_conversacional 1, sin_evidencia 2 |
| AMBIGUOUS | 19 | 0 | 0 | 0 | coloquial 1, corta 16, obligatorio 1, referencia_sin_contexto 2 |
| **Total** | 194 | 125 | 12 | 16 | |

Las expectativas se refieren al último turno. `procedimiento_esperado` es un id de `kb/procedimientos` o null; con null
y respuesta esperada, el caso es un «procedimiento implícito» (los pasos están en una sección del manual, citada en
`fragmentos_relevantes`, pero no hay YAML): cuenta para clasificación, recuperación, invención y éxito de la tarea,
no para PAS. El arnés valida cada caso contra la KB vigente antes de correr y excluye (con aviso) los que no validan.

## Cómo correrlo

```bash
cd s10-conocimiento/metrin
go run ./cmd/evalv2 --validar                                   # dataset contra kb/, sin HTTP
go run ./cmd/evalv2 --url http://127.0.0.1:4762 --contenedor metrin-traza-prueba
go run ./cmd/evalv2 --muestra 20 --versiones v1                 # prueba rápida
go test ./cmd/evalv2/...                                        # pruebas del cálculo de métricas
```

El servicio debe tener la traza habilitada (`METRIN_TRAZA=1`): sin ella las métricas siguen saliendo, pero Recall@10
de V1 pierde los descartados y la latencia del servidor queda vacía.
