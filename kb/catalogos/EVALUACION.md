# Evaluación de los clasificadores de Metrín (catálogos YAML)

Generado por `herramientas/evaluar_clasificador_metrin.py` el 2026-10-06. Las tablas se regeneran al volver a correrlo; el bloque «Análisis» está escrito a mano y se conserva.

## Resumen

| | Intenciones (candidato) | Tipo de respuesta (candidato) |
|---|---:|---:|
| Clases / ejemplos de entrenamiento | 17 / 547 | 7 / 358 |
| Frases de prueba aparte | 155 | 338 |
| Modo de puntuación (elegido por leave-one-out del entrenamiento) | centroide | centroide |
| Exactitud top-1 sin umbral | 59.4 % | 54.4 % |
| Top-3 | 80.6 % | 80.8 % |
| Macro-F1 sin umbral | 59.3 % | 54.0 % |
| Umbral elegido con los datos | 0.485 | 0.590 |
| Cobertura / precisión con ese umbral (en muestra) | 30.3 % / 78.7 % | 2.4 % / 75.0 % |
| **Calibración cruzada** cobertura / precisión | **39.8 % / 72.4 %** | **16.8 % / 57.9 %** |
| Macro-F1 con umbral | 33.5 % | 6.3 % |
| Exactitud / macro-F1 con MiniLM (mismo kNN) | 70.3 % / 69.3 % | 67.8 % / 68.4 % |

ACTUAL (`defecto.json`, 4 clases) frente a CANDIDATO, en las 155 frases de prueba de intenciones llevadas a las 4 clases del actual: **71.0 % → 69.0 %**; en la ruta efectiva de hoy (conversación / fuera de alcance / RAG): 82.6 % → 87.1 %.

Casos medidos:

| Frase | Esperado | ACTUAL | CANDIDATO (estático) | MiniLM |
|---|---|---|---|---|
| «¿qué es un metrado?» | consulta_concepto | social (0.569) | trabajo (0.437) | consulta_concepto (0.636) |
| «¿Cómo registro un nuevo presupuesto en S10?» | soporte_como_hacer | trabajo (0.322) | soporte_como_hacer (0.502) | consulta_comercial (0.620) |

<!-- analisis:inicio -->
## Análisis

**1. El techo lo pone el embebedor estático, no la cantidad de ejemplos.** `potion-es-int8` promedia vectores de
subpalabras: el TEMA («plan de cuentas», «metrado», «orden de compra») pesa mucho más que el ACTO («qué es», «cómo»,
«dónde», «no puedo»), que es justo lo que distinguen estas intenciones. Por eso el leave-one-out del propio
entrenamiento ronda el 53–55 % y el kNN k=1 es el peor modo: el vecino más cercano es el que comparte el sustantivo, no la
intención («¿qué es el plan de cuentas?» → reporte, porque los reportes hablan de cuentas). Los centroides promedian
los temas de cada clase y dejan la señal del acto: +6,5 puntos en intenciones y +18 en tipo de respuesta. Ese modo ya
existe en `clasificar.go` (modelo sin `ejemplos`) y se eligió por leave-one-out, sin mirar la prueba. No se amplió el
entrenamiento para «subir la cifra»: con LOO del 53 % el problema no es de cobertura de ejemplos.

**2. Intenciones (candidato).** Top-1 59,4 %, top-3 80,6 %. Bien: agradecimiento, despedida, saludo, humano, comparación
y aprendizaje (F1 71–100 %). Mal: justo las técnicas que más importan a un asistente instruccional —soporte_como_hacer
26 %, ambiguo 29 %, pedir_aclaracion 36 %, reporte 44 %— que comparten vocabulario de S10 y solo se separan por el acto.
La clase `ninguno` no deja pasar ningún falso positivo con el umbral, pero tampoco se reconoce a sí misma (recall 33 %
sin umbral): el ruido acaba en abstención, que es el resultado seguro.

**3. El umbral: la similitud apenas separa aciertos de errores.** La curva cobertura/precisión es casi plana (63–79 % de
precisión entre 0,30 y 0,50). «Precisión ≥ 90 %» solo se alcanza con umbral 0,652 y 1,9 % de cobertura, así que se usó
el criterio de coste (aciertos − 2 × errores): umbral **0,485** en intenciones (30 % de cobertura y 79 % de precisión en
muestra; **40 % / 72 % en calibración cruzada**, con ± 0,07 de dispersión del umbral). En la práctica, bajo el umbral
Metrín debe preguntar con las 3 primeras candidatas (ACLARAR_TAREA): el top-3 contiene la correcta en el 81 % de las
frases (90 % con MiniLM). Para tipo de respuesta el mismo criterio dice que, con el estático, casi nunca conviene decidir
solo (umbral 0,590, 2,4 % de cobertura): su precisión está en 55–60 % en todo el rango, por debajo del 67 % que
equilibra un error contra una aclaración. Sin umbral su macro-F1 es 54,0 % (las reglas actuales, 31,6 %).

**4. ACTUAL frente a CANDIDATO.** En las 4 clases del actual no hay mejora global (71,0 % → 69,0 %), pero la
calidad de los aciertos cambia: 60 de los 73 aciertos del actual en `trabajo` llegan por abstención (similitud < 0,40),
y el actual manda 4 preguntas técnicas a `social` (sacándolas del RAG, el error caro). El candidato no saca ninguna
pregunta técnica de la ruta segura (precisión 100 % en social y ayuda) y en la ruta efectiva de hoy pasa de 82,6 % a
87,1 %. Su debilidad es la contraria: se abstiene mucho (ayuda → trabajo ×28).
- «¿qué es un metrado?»: el actual dice `social` (0,569) —el error caro, desvía del RAG—; el candidato se abstiene
  (pedir_aclaracion 0,437 frente a consulta_concepto 0,419, bajo el umbral): ruta segura, pero no lo reconoce como
  concepto. MiniLM sí: consulta_concepto (0,636).
- «¿Cómo registro un nuevo presupuesto en S10?»: el actual llega a `trabajo` solo por abstención (0,322); el
  candidato la reconoce como soporte_como_hacer (0,502, sobre el umbral). MiniLM la manda a consulta_comercial
  (0,620): «nuevo … S10» se parece a «comprar S10». El embebedor mejor no es mejor en todo.

**5. Tipo de respuesta.** El candidato estático (54,4 %) supera a las reglas actuales (37,9 %; macro-F1 31,6 %), que no
tienen NAVIGATION ni CONFIGURATION y mandan «dónde…» y «cómo configuro…» a procedimiento. Con MiniLM: 67,8 % / macro-F1
68,4 %, y su criterio de coste ya acepta decidir siempre (calibración cruzada 65,5 % de cobertura con 71,1 % de
precisión). Literales: el estático acierta 5 de 7 en top-1 pero todos quedan bajo su umbral; las reglas aciertan 5 de 7
(fallan «donde veo metrados» → procedimiento y «no puedo guardar metrado» → sin tipo); MiniLM, 3 de 7 (manda «como hago
un metrado», «como registro metrado» y «que es metrado» a UNKNOWN): las consultas de 2–3 palabras son difíciles para los
dos embebedores.

**6. Exploración (medida, no integrada).** Usar primero las reglas inequívocas por definición (concepto: «qué es»;
comparación: «diferencia»; problema: «error», «no funciona») y el clasificador para el resto sube el macro-F1 de tipo de
respuesta de 54,0 % a 59,0 % (estático) y de 68,4 % a 70,7 % (MiniLM). Usar también la regla de procedimiento empeora
(49,5 %), porque `esHowTo` incluye «dónde» y «encontrar», que aquí son NAVIGATION. Es la misma idea que la capa léxica
social que ya tiene `Clasificar`.

**Recomendación.** (a) No reemplazar `defecto.json`: el candidato no mejora la decisión de 4 clases y el orquestador
no entiende sus etiquetas. (b) Integrarlo como segunda etapa, dentro de lo que hoy es `trabajo`, para elegir el acto
(procedimiento, concepto, error, reporte…), con su umbral y ACLARAR_TAREA con el top-3 cuando se abstenga. (c) Para el
modo de la V2, capa léxica de actos inequívocos más clasificador; con el estático, el clasificador solo no alcanza.
(d) Evaluar un embebedor contextual en el presupuesto de RAM del servidor: MiniLM-L12 da +11 puntos en intenciones y
+14 en tipo de respuesta con el mismo kNN, pero pesa unos 470 MB en float32 (96 de sus 118 M de parámetros son la
tabla de vocabulario), frente a los 8,9 MB del estático. (e) Recoger consultas reales anonimizadas para la prueba: estas
cifras salen de frases sintéticas escritas por el mismo equipo.
<!-- analisis:fin -->

## Método

- **Mecanismo**: el de Metrín, sin cambios. `rag-go/internal/clasificar.Entrenar` y `Modelo.Clasificar`/`Puntajes` (kNN k=1: para cada intención, el mejor coseno entre la frase y sus ejemplos; texto en minúsculas, sin tildes ni puntuación externa) sobre el embebedor estático `metrin/modelos/potion-es-int8.pjge` (model2vec, 256 dim, int8). Se llama desde `herramientas/clasificador_metrin/` (Go, módulo aparte enlazado por `go.work`; no toca `metrin/`). `rag entrenar-clasificador` no sirve: su tabla `intencionDe` solo acepta 8 categorías y las pliega en 4 clases.
- **Modo de puntuación**: `Puntajes` tiene dos modos sin tocar Go: kNN k=1 si el modelo trae `ejemplos`, centroides (coseno con la media normalizada de cada intención) si no los trae. Se elige el de mayor macro-F1 en leave-one-out del ENTRENAMIENTO (sin mirar la prueba) y el candidato se guarda en ese modo (sin `ejemplos` si gana centroides).
- **Paridad**: los puntajes del modo elegido reimplementados en Python con los vectores de Go reproducen los de Go con una diferencia máxima de 3.1e-08; por eso la comparación con MiniLM usa el mismo algoritmo.
- **Datos**: entrenamiento = `intenciones.yml` + `ruido.yml` (clase `ninguno`) y, aparte, `tipo_respuesta.yml`. Prueba = `intenciones_prueba.yml` y `tipo_respuesta_prueba.yml`, que nunca entran al entrenamiento (el validador rechaza frases iguales tras normalizar y casi iguales ≥ 0,90). Todo es sintético: ver «Límites».
- **Umbral de abstención, elegido con los datos**: se recorre cada puntaje observado de la prueba como umbral y se toma el que **maximiza aciertos − 2 × errores** entre lo que se etiqueta; abstenerse vale 0 (cuesta una pregunta de aclaración, ACLARAR_TAREA) y un error vale el doble (respuesta en el modo equivocado que el usuario tiene que corregir). Equivale a etiquetar mientras la precisión marginal supere 67%. El criterio inicial, «máxima cobertura con precisión ≥ 90%», se descartó porque con este embebedor degenera (solo lo cumplen umbrales que etiquetan un puñado de frases); su punto se sigue informando. Bajo el umbral, intenciones se abstiene y tipo de respuesta cae en UNKNOWN. El umbral se elige sobre la misma prueba (cifra «en muestra»); la estimación honesta es la **calibración cruzada**: 10 particiones estratificadas en dos mitades, se elige el umbral con una y se mide en la otra.
- **Cobertura** = fracción de frases etiquetadas (no abstenidas; en tipo de respuesta, no UNKNOWN). **Precisión** = aciertos entre las etiquetadas.
- **ACTUAL vs CANDIDATO**: el actual solo tiene `social`, `ayuda`, `limite` y `trabajo`. Para comparar sobre la misma prueba se lleva cada etiqueta a esas 4: saludo/despedida/agradecimiento → social; avance_paso, pedir_aclaracion, ambiguo, feedback_respuesta → ayuda (las categorías «seguimiento», «explicaciones», «ambiguas» y «correctivo» del entrenamiento actual); ninguno → limite; el resto (concepto, cómo hacer, error, reporte, funcionalidad, comparación, comercial, humano, aprendizaje) → trabajo. Las salidas se comparan tal como las devuelve `Clasificar` (umbral y capa léxica social incluidos): bajo el umbral, ambos dicen `trabajo`.
- **Reglas actuales de tipo de consulta**: port literal a Python de `tipoConsulta`/`esHowTo` (`internal/rag/orquestador.go` y `rag.go`), sin hilo (no hay «seguimiento»). problema → TROUBLESHOOTING, comparacion → COMPARISON, procedimiento → PROCEDURE, concepto → CONCEPT, aclaracion e informacion_directa → UNKNOWN. Las reglas no tienen NAVIGATION ni CONFIGURATION ("dónde" y "encontrar" las mandan a procedimiento).
- **Embebedor alternativo**: `paraphrase-multilingual-MiniLM-L12-v2` (BERT de 12 capas, 384 dim, multilingüe; uno de los candidatos de §3 de la especificación), leído de la caché local de Hugging Face e inferido en numpy (sin torch ni red), media de tokens normalizada; texto crudo. Mismo kNN y mismo criterio de umbral. Tardó 21 s para 1398 frases.

Reproducir: `.venv/bin/python3 herramientas/validar_catalogos_metrin.py && .venv/bin/python3 herramientas/evaluar_clasificador_metrin.py`. Modelos: `metrin/modelos/clasificador-candidato.json` y `metrin/modelos/clasificador-tipo-respuesta-candidato.json` (formato de `clasificar.Modelo`, se cargan con `rag clasificar --modelo <ruta>`), con el umbral elegido en `umbral`. `defecto.json` NO se ha tocado.

### Modos de puntuación (✓ = elegido por leave-one-out del entrenamiento)

| Catálogo · embebedor | Modo | Macro-F1 LOO (entren.) | Exactitud prueba | Macro-F1 prueba | Umbral | Cruzada cob. / prec. |
|---|---|---:|---:|---:|---:|---:|
| Intenciones · estático | knn1 | 36.6 % | 52.9 % | 53.6 % | 0.765 | 9.1 % / 71.7 % |
| Intenciones · estático | centroide ✓ | 53.2 % | 59.4 % | 59.3 % | 0.485 | 39.8 % / 72.4 % |
| Tipo de respuesta · estático | knn1 | 19.5 % | 36.1 % | 36.0 % | 0.987 | 0.7 % / 1.2 % |
| Tipo de respuesta · estático | centroide ✓ | 54.9 % | 54.4 % | 54.0 % | 0.590 | 16.8 % / 57.9 % |
| Intenciones · MiniLM | knn1 | 43.2 % | 63.2 % | 62.8 % | 0.805 | 35.7 % / 82.6 % |
| Intenciones · MiniLM | centroide ✓ | 61.9 % | 70.3 % | 69.3 % | 0.523 | 71.7 % / 76.8 % |
| Tipo de respuesta · MiniLM | knn1 | 29.6 % | 45.9 % | 46.5 % | 0.866 | 3.0 % / 45.9 % |
| Tipo de respuesta · MiniLM | centroide ✓ | 68.1 % | 67.8 % | 68.4 % | 0.000 | 65.5 % / 71.1 % |

El modo híbrido (½ kNN top-3 + ½ centroide, el que ganó en kddesign) no se puede usar sin cambiar `clasificar.go`; no se mide aquí para no elegir nada que el servicio no ejecute.

## 1. Intenciones

Entrenamiento: modelo estatico-10c8dbd2f167: 17 intenciones, 547 ejemplos, umbral 0.485 → /Users/david.roncal/Downloads/PjgFactSalud_completo/s10-conocimiento/metrin/modelos/clasificador-candidato.json (guardado en modo centroides: sin `ejemplos`)

### 1.1 Por intención (candidato, estático)

| Clase | n | Precisión | Recall | F1 | Recall con umbral | Abstenciones |
|---|---:|---:|---:|---:|---:|---:|
| soporte_como_hacer | 10 | 23.1 % | 30.0 % | 26.1 % | 10.0 % | 8 |
| ambiguo | 9 | 40.0 % | 22.2 % | 28.6 % | 22.2 % | 5 |
| pedir_aclaracion | 9 | 30.8 % | 44.4 % | 36.4 % | 11.1 % | 8 |
| solicitud_reporte | 9 | 44.4 % | 44.4 % | 44.4 % | 11.1 % | 7 |
| avance_paso | 9 | 75.0 % | 33.3 % | 46.2 % | 11.1 % | 8 |
| ninguno | 12 | 80.0 % | 33.3 % | 47.1 % | 0.0 % | 12 |
| consulta_funcionalidad | 9 | 45.5 % | 55.6 % | 50.0 % | 11.1 % | 7 |
| consulta_concepto | 10 | 55.6 % | 50.0 % | 52.6 % | 0.0 % | 10 |
| feedback_respuesta | 9 | 62.5 % | 55.6 % | 58.8 % | 22.2 % | 6 |
| soporte_error | 9 | 62.5 % | 55.6 % | 58.8 % | 11.1 % | 6 |
| consulta_comercial | 8 | 62.5 % | 62.5 % | 62.5 % | 12.5 % | 6 |
| solicitud_aprendizaje | 8 | 66.7 % | 75.0 % | 70.6 % | 0.0 % | 8 |
| solicitud_comparacion | 9 | 61.5 % | 88.9 % | 72.7 % | 55.6 % | 3 |
| solicitud_humano | 8 | 70.0 % | 87.5 % | 77.8 % | 62.5 % | 3 |
| saludo | 9 | 69.2 % | 100.0 % | 81.8 % | 44.4 % | 5 |
| despedida | 9 | 100.0 % | 88.9 % | 94.1 % | 66.7 % | 3 |
| agradecimiento | 9 | 100.0 % | 100.0 % | 100.0 % | 66.7 % | 3 |

Precisión, recall y F1 sin umbral (top-1); las dos últimas columnas, con el umbral elegido. Ordenado de peor a mejor F1.

### 1.2 Confusiones (matriz resumida)

Pares real → predicha más frecuentes (top-1, sin umbral): solicitud_reporte → soporte_como_hacer ×4; ambiguo → pedir_aclaracion ×3; soporte_como_hacer → consulta_concepto ×2; soporte_como_hacer → solicitud_aprendizaje ×2; avance_paso → pedir_aclaracion ×2; pedir_aclaracion → feedback_respuesta ×2; soporte_error → soporte_como_hacer ×2; soporte_error → consulta_funcionalidad ×2; feedback_respuesta → pedir_aclaracion ×2; ninguno → saludo ×2; ninguno → solicitud_humano ×2; despedida → saludo ×1; consulta_concepto → pedir_aclaracion ×1; consulta_concepto → consulta_comercial ×1; consulta_concepto → solicitud_reporte ×1.

Clase `ninguno` (fuera de dominio y basura): 12 frases de prueba; con el umbral, 0 se etiquetan como una intención real (falsos positivos: ninguno). Frases del dominio mandadas a `ninguno`: 0 (—).

Fallos de la prueba (top-1, sin umbral):

- «hasta el lunes»: esperado despedida, salió saludo (0.265, bajo el umbral → abstención)
- «¿qué es un metrado?»: esperado consulta_concepto, salió pedir_aclaracion (0.437, bajo el umbral → abstención)
- «a qué le llaman presupuesto de venta»: esperado consulta_concepto, salió consulta_comercial (0.421, bajo el umbral → abstención)
- «qué es un anticipo a proveedores»: esperado consulta_concepto, salió solicitud_reporte (0.389, bajo el umbral → abstención)
- «q es la conformidad de servicio»: esperado consulta_concepto, salió soporte_como_hacer (0.391, bajo el umbral → abstención)
- «qué vendría a ser una restricción en Lean»: esperado consulta_concepto, salió solicitud_comparacion (0.234, bajo el umbral → abstención)
- «cuenta de gasto vs cuenta de costo en contabilidad»: esperado solicitud_comparacion, salió soporte_como_hacer (0.652)
- «¿qué pasos sigo para publicar un proyecto desde Lookahead?»: esperado soporte_como_hacer, salió avance_paso (0.357, bajo el umbral → abstención)
- «cómo vinculo una partida de control con las partidas del presupuesto»: esperado soporte_como_hacer, salió consulta_concepto (0.516)
- «enséñame a hacer la rendición de una caja chica»: esperado soporte_como_hacer, salió solicitud_aprendizaje (0.372, bajo el umbral → abstención)
- «cómo actualizo S10 con la última revisión»: esperado soporte_como_hacer, salió consulta_funcionalidad (0.341, bajo el umbral → abstención)
- «como le pongo los índices unificados a los insumos»: esperado soporte_como_hacer, salió consulta_concepto (0.291, bajo el umbral → abstención)
- «dónde se registra el parte diario de equipos»: esperado soporte_como_hacer, salió solicitud_reporte (0.430, bajo el umbral → abstención)
- «quiero aprender a manejar el módulo de almacenes»: esperado soporte_como_hacer, salió solicitud_aprendizaje (0.479, bajo el umbral → abstención)
- «ya quedó, ¿ahora?»: esperado avance_paso, salió saludo (0.379, bajo el umbral → abstención)
- «ya apareció el mensaje de conforme»: esperado avance_paso, salió feedback_respuesta (0.400, bajo el umbral → abstención)
- «listo, ya seleccioné el proyecto»: esperado avance_paso, salió solicitud_reporte (0.392, bajo el umbral → abstención)
- «sí, me salió igual que dijiste»: esperado avance_paso, salió pedir_aclaracion (0.354, bajo el umbral → abstención)
- «regresemos a lo que estábamos haciendo»: esperado avance_paso, salió pedir_aclaracion (0.485, bajo el umbral → abstención)
- «ya está guardado, sigamos»: esperado avance_paso, salió soporte_como_hacer (0.309, bajo el umbral → abstención)
- «uy no entendí nada»: esperado pedir_aclaracion, salió feedback_respuesta (0.366, bajo el umbral → abstención)
- «más sencillo porfa»: esperado pedir_aclaracion, salió feedback_respuesta (0.193, bajo el umbral → abstención)
- «¿podrías poner un ejemplo con números?»: esperado pedir_aclaracion, salió consulta_funcionalidad (0.438, bajo el umbral → abstención)
- «dímelo con otras palabras»: esperado pedir_aclaracion, salió solicitud_humano (0.261, bajo el umbral → abstención)
- «¿a qué menú te refieres?»: esperado pedir_aclaracion, salió consulta_concepto (0.425, bajo el umbral → abstención)
- «no puedo entrar al módulo de contabilidad, dice acceso denegado»: esperado soporte_error, salió soporte_como_hacer (0.477, bajo el umbral → abstención)
- «la detracción no se calcula sola»: esperado soporte_error, salió consulta_funcionalidad (0.499)
- «S10 se cierra solito cuando abro el presupuesto»: esperado soporte_error, salió consulta_funcionalidad (0.554)
- «el portal del proveedor no carga las facturas»: esperado soporte_error, salió soporte_como_hacer (0.405, bajo el umbral → abstención)
- «necesito sacar el registro de compras para el contador»: esperado solicitud_reporte, salió soporte_como_hacer (0.577)
- «cómo imprimo las boletas de pago de la quincena»: esperado solicitud_reporte, salió soporte_como_hacer (0.472, bajo el umbral → abstención)
- «listado de proveedores con su RUC»: esperado solicitud_reporte, salió consulta_comercial (0.399, bajo el umbral → abstención)
- «dónde veo los movimientos de un insumo en el almacén de obra»: esperado solicitud_reporte, salió soporte_como_hacer (0.411, bajo el umbral → abstención)
- «quiero el mayor de una cuenta contable»: esperado solicitud_reporte, salió soporte_como_hacer (0.485, bajo el umbral → abstención)
- «¿el sistema permite firmar las órdenes de compra digitalmente?»: esperado consulta_funcionalidad, salió consulta_comercial (0.442, bajo el umbral → abstención)
- «¿cuántos usuarios pueden entrar a la vez?»: esperado consulta_funcionalidad, salió ambiguo (0.240, bajo el umbral → abstención)
- «el portal del colaborador muestra las boletas?»: esperado consulta_funcionalidad, salió solicitud_reporte (0.344, bajo el umbral → abstención)
- «¿se puede tener un mismo insumo con dos precios en proyectos distintos?»: esperado consulta_funcionalidad, salió solicitud_comparacion (0.559)
- «buen dato, eso era»: esperado feedback_respuesta, salió pedir_aclaracion (0.509)
- «la explicación tiene errores»: esperado feedback_respuesta, salió pedir_aclaracion (0.419, bajo el umbral → abstención)
- «no corresponde a mi pregunta»: esperado feedback_respuesta, salió ambiguo (0.458, bajo el umbral → abstención)
- «la fuente que citas es de otro módulo»: esperado feedback_respuesta, salió solicitud_comparacion (0.377, bajo el umbral → abstención)
- «te comparto un PDF para que lo agregues»: esperado solicitud_aprendizaje, salió ninguno (0.331, bajo el umbral → abstención)
- «incorpora este procedimiento interno»: esperado solicitud_aprendizaje, salió soporte_error (0.285, bajo el umbral → abstención)
- «¿qué precio tiene el ERP completo?»: esperado consulta_comercial, salió consulta_concepto (0.622)
- «me gustaría una demostración del módulo de gerencia de proyectos»: esperado consulta_comercial, salió consulta_funcionalidad (0.416, bajo el umbral → abstención)
- «cuánto me costaría agregar 3 usuarios más»: esperado consulta_comercial, salió soporte_error (0.327, bajo el umbral → abstención)
- «escala mi problema a un especialista»: esperado solicitud_humano, salió ambiguo (0.451, bajo el umbral → abstención)
- «oye, una cosa»: esperado ambiguo, salió pedir_aclaracion (0.475, bajo el umbral → abstención)
- «kardex»: esperado ambiguo, salió solicitud_reporte (0.240, bajo el umbral → abstención)
- «lo que te comenté antes»: esperado ambiguo, salió pedir_aclaracion (0.535)
- «necesito información»: esperado ambiguo, salió solicitud_aprendizaje (0.514)
- «valorizaciones»: esperado ambiguo, salió solicitud_comparacion (0.446, bajo el umbral → abstención)
- «algo sobre el módulo de nóminas»: esperado ambiguo, salió consulta_funcionalidad (0.335, bajo el umbral → abstención)
- «¿puedes con eso?»: esperado ambiguo, salió pedir_aclaracion (0.358, bajo el umbral → abstención)
- «¿cuál es la montaña más alta del mundo?»: esperado ninguno, salió saludo (0.199, bajo el umbral → abstención)
- «cómo se prepara un pisco sour»: esperado ninguno, salió soporte_como_hacer (0.360, bajo el umbral → abstención)
- «recomiéndame un perfume»: esperado ninguno, salió solicitud_humano (0.280, bajo el umbral → abstención)
- «probando 1 2 3»: esperado ninguno, salió pedir_aclaracion (0.186, bajo el umbral → abstención)
- «aaaaa bbbb»: esperado ninguno, salió saludo (0.197, bajo el umbral → abstención)
- «cuánto pesa un elefante»: esperado ninguno, salió solicitud_comparacion (0.297, bajo el umbral → abstención)
- «cómo saco mi pasaporte»: esperado ninguno, salió soporte_error (0.259, bajo el umbral → abstención)
- «¿me ayudas a escribir un mensaje de cumpleaños?»: esperado ninguno, salió solicitud_humano (0.382, bajo el umbral → abstención)

### 1.3 Cobertura y precisión según el umbral

| Umbral | Cobertura | Precisión |
|---:|---:|---:|
| 0.300 | 86.5 % | 63.4 % |
| 0.350 | 79.4 % | 65.0 % |
| 0.400 | 61.3 % | 69.5 % |
| 0.450 | 41.9 % | 72.3 % |
| 0.485 ← elegido | 30.3 % | 78.7 % |
| 0.500 | 25.2 % | 76.9 % |
| 0.550 | 10.3 % | 68.8 % |
| 0.600 | 5.2 % | 75.0 % |
| 0.650 | 2.6 % | 75.0 % |
| 0.700 | 1.3 % | 100.0 % |
| 0.750 | 0.0 % | — |
| 0.800 | 0.0 % | — |
| 0.850 | 0.0 % | — |
| 0.900 | 0.0 % | — |
| 0.950 | 0.0 % | — |

Criterio: máx. aciertos − 2×errores. Referencia: la máxima cobertura con precisión ≥ 90% se logra con umbral 0.652 y cubre solo 1.9 %. Calibración cruzada (n=20): umbral 0.469 ± 0.074, cobertura 39.8 %, precisión 72.4 % ± 11.5 %.

### 1.4 ACTUAL (`defecto.json`) frente a CANDIDATO, en las 4 clases del actual

| Clase del actual | n | Recall ACTUAL | Precisión ACTUAL | Recall CANDIDATO | Precisión CANDIDATO |
|---|---:|---:|---:|---:|---:|
| social | 27 | 74.1 % | 66.7 % | 70.4 % | 100.0 % |
| ayuda | 36 | 41.7 % | 75.0 % | 22.2 % | 100.0 % |
| limite | 12 | 16.7 % | 66.7 % | 0.0 % | 0.0 % |
| trabajo | 80 | 91.2 % | 71.6 % | 100.0 % | 62.5 % |
| **global (exactitud)** | 155 | 71.0 % | | 69.0 % | |

- De los aciertos del ACTUAL en `trabajo`, 60 llegan por abstención (similitud < 0,40 → `trabajo` por defecto), no porque el modelo reconozca la frase.
- Ruta efectiva de hoy (social y limite desvían; ayuda y trabajo van al RAG): ACTUAL 82.6 %, CANDIDATO 87.1 %.
- Exactitud FINA del candidato tal como responde `Clasificar` (con su umbral; abstención = fallo): 23.9 %. El actual no tiene etiquetas finas: no se puede medir ahí.

Confusiones del ACTUAL en 4 clases: ayuda → trabajo ×15; limite → trabajo ×9; social → trabajo ×5; ayuda → social ×5; trabajo → social ×4; trabajo → ayuda ×3; social → ayuda ×2; ayuda → limite ×1.

Confusiones del CANDIDATO en 4 clases: ayuda → trabajo ×28; limite → trabajo ×12; social → trabajo ×8.

### 1.5 Casos medidos

- «¿qué es un metrado?» (esperado **consulta_concepto**): ACTUAL social (0.569; top-3 [('social', 0.569), ('trabajo', 0.471), ('ayuda', 0.45)]); CANDIDATO trabajo (0.437; top-3 [('pedir_aclaracion', 0.437), ('consulta_concepto', 0.419), ('avance_paso', 0.315)]); MiniLM consulta_concepto (0.636, con su umbral: consulta_concepto).
- «¿Cómo registro un nuevo presupuesto en S10?» (esperado **soporte_como_hacer**): ACTUAL trabajo (0.322; top-3 [('trabajo', 0.322), ('limite', 0.289), ('ayuda', 0.197)]); CANDIDATO soporte_como_hacer (0.502; top-3 [('soporte_como_hacer', 0.502), ('consulta_comercial', 0.459), ('consulta_funcionalidad', 0.457)]); MiniLM consulta_comercial (0.620, con su umbral: consulta_comercial).

## 2. Tipo de respuesta (modo de la V2)

### 2.1 Precisión, recall y F1 por clase

Con el umbral elegido (0.590; bajo él, UNKNOWN). Entre paréntesis, sin umbral.

| Clase | n | Precisión | Recall | F1 | F1 sin umbral | F1 reglas actuales | F1 MiniLM (con su umbral) |
|---|---:|---:|---:|---:|---:|---:|---:|
| COMPARISON | 34 | 100.0 % | 5.9 % | 11.1 % | 50.0 % | 33.3 % | 73.2 % |
| CONCEPT | 54 | 0.0 % | 0.0 % | 0.0 % | 53.2 % | 79.1 % | 70.9 % |
| CONFIGURATION | 54 | 100.0 % | 5.6 % | 10.5 % | 48.8 % | 0.0 % | 69.5 % |
| NAVIGATION | 54 | 0.0 % | 0.0 % | 0.0 % | 63.9 % | 0.0 % | 73.7 % |
| PROCEDURE | 54 | 0.0 % | 0.0 % | 0.0 % | 46.5 % | 47.3 % | 49.6 % |
| TROUBLESHOOTING | 54 | 100.0 % | 1.9 % | 3.6 % | 61.5 % | 25.4 % | 83.2 % |
| UNKNOWN | 34 | 10.3 % | 100.0 % | 18.7 % | 54.0 % | 35.8 % | 58.7 % |
| **macro** | 338 | | | **6.3 %** | 54.0 % | 31.6 % | **68.4 %** |
| exactitud | | | | 11.8 % | 54.4 % | 37.9 % | 67.8 % |

### 2.2 Matriz de confusión (estático, con umbral)

| real \ predicha | COMPA | CONCE | CONFI | NAVIG | PROCE | TROUB | UNKNO |
|---|---:|---:|---:|---:|---:|---:|---:|
| COMPARISON | **2** | · | · | · | 1 | · | 31 |
| CONCEPT | · | **0** | · | · | 1 | · | 53 |
| CONFIGURATION | · | · | **3** | · | · | · | 51 |
| NAVIGATION | · | · | · | **0** | · | · | 54 |
| PROCEDURE | · | · | · | · | **0** | · | 54 |
| TROUBLESHOOTING | · | · | · | · | · | **1** | 53 |
| UNKNOWN | · | · | · | · | · | · | **34** |

Reglas actuales (`tipoConsulta`):

| real \ predicha | COMPA | CONCE | CONFI | NAVIG | PROCE | TROUB | UNKNO |
|---|---:|---:|---:|---:|---:|---:|---:|
| COMPARISON | **7** | · | · | · | 1 | · | 26 |
| CONCEPT | · | **36** | · | · | · | · | 18 |
| CONFIGURATION | · | · | **0** | · | 49 | · | 5 |
| NAVIGATION | 1 | · | · | **0** | 32 | · | 21 |
| PROCEDURE | · | · | · | · | **43** | 1 | 10 |
| TROUBLESHOOTING | · | 1 | · | · | 3 | **8** | 42 |
| UNKNOWN | · | · | · | · | · | · | **34** |

MiniLM (mismo kNN, con su umbral):

| real \ predicha | COMPA | CONCE | CONFI | NAVIG | PROCE | TROUB | UNKNO |
|---|---:|---:|---:|---:|---:|---:|---:|
| COMPARISON | **26** | 2 | 1 | · | 5 | · | · |
| CONCEPT | 4 | **39** | · | 2 | 4 | · | 5 |
| CONFIGURATION | 1 | 2 | **33** | · | 17 | 1 | · |
| NAVIGATION | 2 | 8 | · | **35** | 6 | · | 3 |
| PROCEDURE | · | 2 | 5 | 2 | **32** | 4 | 9 |
| TROUBLESHOOTING | 3 | 2 | 1 | 1 | 3 | **42** | 2 |
| UNKNOWN | 1 | 1 | 1 | 1 | 8 | · | **22** |

### 2.3 Cobertura y precisión según el umbral (estático)

| Umbral | Cobertura (no UNKNOWN) | Precisión |
|---:|---:|---:|
| 0.300 | 77.8 % | 56.7 % |
| 0.350 | 66.9 % | 57.5 % |
| 0.400 | 50.3 % | 61.2 % |
| 0.450 | 31.7 % | 58.9 % |
| 0.500 | 16.9 % | 54.4 % |
| 0.550 | 7.4 % | 64.0 % |
| 0.590 ← elegido | 2.4 % | 75.0 % |
| 0.600 | 1.8 % | 66.7 % |
| 0.650 | 0.9 % | 66.7 % |
| 0.700 | 0.3 % | 100.0 % |
| 0.750 | 0.0 % | — |
| 0.800 | 0.0 % | — |
| 0.850 | 0.0 % | — |
| 0.900 | 0.0 % | — |
| 0.950 | 0.0 % | — |

Criterio: máx. aciertos − 2×errores. Referencia: la máxima cobertura con precisión ≥ 90% se logra con umbral 0.666 y cubre solo 0.3 %. Calibración cruzada: umbral 0.548 ± 0.105, cobertura 16.8 %, precisión 57.9 % ± 13.1 %.
MiniLM: umbral 0.000 (máx. aciertos − 2×errores), cobertura/precisión en muestra 87.9 % / 69.7 %; cruzada 65.5 % / 71.1 %.

### 2.4 Casos literales pedidos

| Frase | Esperado | Estático (top-1, sim → con umbral) | Reglas actuales | MiniLM |
|---|---|---|---|---|
| «como hago un metrado» | PROCEDURE | PROCEDURE (0.407) → UNKNOWN | PROCEDURE | UNKNOWN (0.624) → UNKNOWN |
| «como registro metrado» | PROCEDURE | PROCEDURE (0.408) → UNKNOWN | PROCEDURE | UNKNOWN (0.598) → UNKNOWN |
| «cómo modifico una partida» | PROCEDURE | COMPARISON (0.372) → UNKNOWN | PROCEDURE | CONFIGURATION (0.399) → CONFIGURATION |
| «que es metrado» | CONCEPT | NAVIGATION (0.395) → UNKNOWN | CONCEPT | UNKNOWN (0.785) → UNKNOWN |
| «no puedo guardar metrado» | TROUBLESHOOTING | TROUBLESHOOTING (0.393) → UNKNOWN | UNKNOWN | TROUBLESHOOTING (0.644) → TROUBLESHOOTING |
| «donde veo metrados» | NAVIGATION | NAVIGATION (0.524) → UNKNOWN | PROCEDURE | NAVIGATION (0.596) → NAVIGATION |
| «metrado s10» | UNKNOWN | UNKNOWN (0.322) → UNKNOWN | UNKNOWN | UNKNOWN (0.540) → UNKNOWN |

Fallos del estático (con umbral): PROCEDURE → UNKNOWN ×54; NAVIGATION → UNKNOWN ×54; CONCEPT → UNKNOWN ×53; TROUBLESHOOTING → UNKNOWN ×53; CONFIGURATION → UNKNOWN ×51; COMPARISON → UNKNOWN ×31; CONCEPT → PROCEDURE ×1; COMPARISON → PROCEDURE ×1.

## 3. Estático frente a MiniLM (mismo kNN, mismo criterio de umbral)

| | Intenciones estático | Intenciones MiniLM | Tipo estático | Tipo MiniLM |
|---|---:|---:|---:|---:|
| Exactitud top-1 | 59.4 % | 70.3 % | 54.4 % | 67.8 % |
| Top-3 | 80.6 % | 90.3 % | 80.8 % | 89.3 % |
| Macro-F1 sin umbral | 59.3 % | 69.3 % | 54.0 % | 68.4 % |
| Modo (LOO) | centroide | centroide | centroide | centroide |
| Umbral elegido | 0.485 | 0.523 | 0.590 | 0.000 |
| Cruzada cobertura / precisión | 39.8 % / 72.4 % | 71.7 % / 76.8 % | 16.8 % / 57.9 % | 65.5 % / 71.1 % |
| Macro-F1 con umbral | 33.5 % | 63.2 % | 6.3 % | 68.4 % |

Peores intenciones con MiniLM (F1 sin umbral): pedir_aclaracion 28.6 %, ambiguo 30.8 %, soporte_como_hacer 52.6 %, consulta_funcionalidad 55.6 %, solicitud_reporte 60.0 %.

## Límites

- Entrenamiento y prueba son **sintéticos** y los escribió el mismo equipo con el mismo estilo: las cifras son optimistas respecto de mensajes reales. Cuando haya consultas reales anonimizadas, deben ser la prueba y se recalibra.
- El umbral elegido «en muestra» usa la misma prueba que mide; la cifra para decidir es la calibración cruzada.
- La prueba de intenciones tiene 8–12 frases por clase: un acierto mueve el recall de una clase 8–12 puntos.
- El candidato no está integrado: el orquestador actual (`internal/rag/orquestador.go`) solo entiende `social`, `limite`, `ayuda` y `trabajo`, y `Clasificar` devuelve `trabajo` bajo el umbral. Usarlo exige mapear las nuevas etiquetas a rutas (campo `ruta` de `intenciones.yml`) y tratar `trabajo` como abstención; no se ha cambiado Go.

