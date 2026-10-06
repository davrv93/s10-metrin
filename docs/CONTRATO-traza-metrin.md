# Contrato de la traza de Metrín (v1)

Diseño: `docs/DISENO-modo-traza-metrin.md` y el lienzo «Metrín · Modo traza».
Este contrato lo comparten el servicio Go (`metrin/internal/traza`, `rag`, `servidor`) y la página (`pagina.html`).

## Activación

- La traza existe solo si la variable `METRIN_TRAZA` vale `1` o `true` en el servicio **y** la petición trae
  `"traza": true`. En cualquier otro caso la respuesta no lleva la clave `traza` y es idéntica a la de hoy.
- `GET /health` añade `"traza": true|false` (si el servicio la tiene habilitada). La página muestra el botón «Traza»
  solo si es `true`.
- Respuesta con traza: cabecera `Cache-Control: no-store`.

## Petición

`POST /ask` acepta, además de lo actual (`pregunta`, `k`, `hilo`, `source`), el campo `"traza": true`.

## Respuesta: clave `traza`

```json
{
  "version": 1,
  "total_ms": 41.3,
  "pregunta": "<recorte de 120 caracteres>",
  "truncada": false,
  "etapas": [
    {
      "id": "clasificador",
      "nombre": "Clasificador kNN",
      "tipo": "modelo",
      "estado": "alerta",
      "ms": 0.4,
      "modelo": "kNN-local:estatico-10c8dbd2f167",
      "razon": "etiqueta «social» en una pregunta que siguió por la ruta rag",
      "datos": { "intencion": "social", "similitud": 0.569, "umbral": 0.40, "paso_umbral": true }
    }
  ],
  "oportunidades": [
    { "n": 1, "etapa": "clasificador", "nivel": "alerta", "texto": "…" }
  ]
}
```

- `etapas` trae SIEMPRE las 18 etapas, en este orden y con estos `id`. Las que no se ejecutaron llevan
  `estado` `omitida` (no aplicaba en la ruta tomada) o `no_tomada` (rama que no siguió el mensaje), con `razon`.
- `tipo` decide la forma en la interfaz: `evento` = círculo, `codigo` = rectángulo, `modelo` = píldora,
  `decision` = rombo.
- `estado` decide el color: `ok` verde, `alerta` naranja, `error` rojo, `respaldo` amarillo,
  `omitida` / `no_tomada` gris punteado.
- `ms` es un número con decimales o `null` si no se midió. `modelo` y `razon` pueden faltar.
- `datos` solo lleva identificadores, puntajes y recortes de 120 caracteres; nunca el texto completo de un fragmento.
- Si la traza serializada supera 16 KB, `seleccion.datos.fragmentos` se recorta a 5 y `truncada` es `true`.

| # | id | nombre | tipo | datos mínimos |
|---|---|---|---|---|
| 1 | `inicio` | Pregunta | evento | — |
| 2 | `entrada` | Entrada | codigo | `pregunta`, `hilo_turnos` |
| 3 | `regla_aciertos` | Regla «aciertos» | decision | `aplica` |
| 4 | `embebedor` | Embebedor | modelo | `proveedor`, `modelo` |
| 5 | `clasificador` | Clasificador kNN | modelo | `clasificador`, `intencion`, `similitud`, `umbral`, `paso_umbral` |
| 6 | `seguridad` | ¿Social y corta? | decision | `pregunta_larga`, `desvia_a_charla` |
| 7 | `ruta` | Ruta | decision | `ruta`, `tipo_consulta` |
| 8 | `reescritura` | Reescritura del hilo | modelo | `original`, `reescrita` |
| 9 | `busqueda` | Búsqueda | codigo | `metodo` («vectorial»), `k`, `filtro`, `resultados` |
| 10 | `seleccion` | Selección | codigo | `ventana`, `max_contextos`, `corte`, `seleccionados`, `fragmentos[{documento, pagina, distancia, penalizacion}]` |
| 11 | `evidencia` | ¿Hay contexto? | decision | `sin_contexto`, `distancia_min`, `motivo` |
| 12 | `generacion` | Generación LLM | modelo | `proveedor`, `modelo`, `modo` |
| 13 | `verificacion` | Verificación de citas | codigo | `citas` |
| 14 | `fotos` | Fotos de pasos | codigo | `fotos` |
| 15 | `charla` | Charla LLM | modelo | `modelo` |
| 16 | `jev` | Juicio JEV | modelo | `activo` |
| 17 | `registro_fallos` | Registro de fallos | codigo | `registrado` |
| 18 | `fin` | Respuesta | evento | `modo` |

## Oportunidades

Las calcula el servicio con reglas fijas sobre hechos del turno (nunca con un modelo). Numeradas desde 1, en el
orden de las etapas. Reglas v1:

- `clasificador`: similitud por debajo del umbral → `alerta` («bajo el umbral, fue a la ruta segura»); intención
  `social` o `limite` y ruta `rag` → `alerta` («etiqueta conversacional en una pregunta de trabajo»).
- `seleccion`: algún fragmento seleccionado con penalización de confianza → `alerta`.
- `evidencia`: sin contexto → `alerta`.
- `generacion` o `charla`: error del modelo o texto de respaldo → `error`.
- `reescritura`, `jev`: error → `error`.

## Etapas V2

Añadido el 06-10-2026 (V2 de Metrín, `docs/V2-RAG-PROCEDURAL.md`). Es **aditivo**: las 18 etapas de arriba no cambian
ni de `id` ni de orden, y en un turno V1 la traza no lleva nada de esta sección (ni `agente` ni `etapas_v2`).
Código: `metrin/internal/traza/v2.go` (catálogo y sub-registro `Recorder.V2()`) y `metrin/internal/v2/traza.go`
(lo que anota el orquestador).

En un turno V2 la traza trae además:

```json
{
  "agente": "v2",
  "etapas_v2": [ { "id": "tipo_respuesta", "nombre": "Tipo de respuesta", "tipo": "modelo", "estado": "ok", "ms": 0.2,
                   "modelo": "reglas", "razon": "«PROCEDURE» por reglas, confianza 0,90",
                   "datos": { "tipo": "PROCEDURE", "confianza": 0.9, "fuentes": null, "metodo": "reglas", "umbral": 0.4 } } ]
}
```

- `etapas_v2` trae SIEMPRE las 13 etapas, en este orden, con la misma forma y los mismos `estado` y `tipo` que las de
  V1 (omitidas o no tomadas con `razon`; `ms` null si no se midió; textos de `datos` recortados a 120 caracteres).
- Toda etapa V2 lleva además los datos `confianza` y `fuentes` (null si no aplica).
- Las 18 etapas de V1 siguen presentes: `entrada`, `ruta` (`ruta: "v2"`) y `fin` (`version: "v2"`) con datos; `charla`
  si el mensaje fue SOCIAL (lo contesta la charla de V1); `fotos` omitida (en V2 las fotos van por paso en `plan`); el
  resto `omitida` con razón «turno V2: el camino está en «etapas_v2»».
- Las cuatro de búsqueda las puede anotar el propio recuperador; si no lo hace, el núcleo las deriva de los candidatos
  (`Lexico`, `Vector`, `Rerank`) y de la lista `degradado` del recuperador (`…vector…` → `busqueda_vectorial` en
  `respaldo`; `…rerank…` → `rerank` en `respaldo`).

| # | id | nombre | tipo | datos mínimos (además de `confianza`, `fuentes`) |
|---|---|---|---|---|
| 1 | `tipo_respuesta` | Tipo de respuesta | modelo | `tipo`, `metodo` (reglas · knn · intencion_v1 · decision:<motor> · memoria), `umbral`; también `dudoso`, `margen`, `candidatos`, `knn`, `similitud_knn`, `margen_knn` |
| 2 | `decision` | Motor de decisión | decision | `motor`, `llamadas`, `decisiones[{decision, choice, fuente, confianza, ms, opciones, propuesta, respaldo?}]`, `respaldo` |
| 3 | `plan_consulta` | Plan de consulta | codigo | `clases` (por iteración), `consulta`, `iteraciones`; también `original`, `entidades`, `aliases`, `modulo`, `expandida` |
| 4 | `busqueda_lexica` | Búsqueda léxica | codigo | `resultados`; `iteracion` |
| 5 | `busqueda_vectorial` | Búsqueda vectorial | codigo | `resultados` |
| 6 | `fusion` | Fusión | codigo | `resultados`, `mejor`; `candidatos` (clase:id), `degradado`, `errores` |
| 7 | `rerank` | Reranker | modelo | `activo`, `resultados` |
| 8 | `procedimiento` | ¿Qué procedimiento? | decision | `id`; `titulo`, `alternativas`, `decidido_por` |
| 9 | `pasos` | Pasos | codigo | `pasos` (total del procedimiento), `mostrados` (entrega por partes); `ids` |
| 10 | `fotos_paso` | Fotos por paso | codigo | `fotos`, `pasos_con_foto` (de los pasos mostrados); `ids` («n:id_foto») |
| 11 | `plan` | Plan de respuesta | codigo | `tipo`, `plantilla`, `sin_evidencia`; `siguiente`, `afirmaciones`, `respuesta_segura` |
| 12 | `quality_gate` | Quality gate | codigo | `passed`, `score`, `issues`, `intentos` (1 o 2); `gate` (conocimiento · minimo) |
| 13 | `renderizado` | Redacción | modelo | `generador`, `modo` (llm · plantillas · regenerado_sin_llm · basico · segura · charla); `caracteres`, `presupuesto`, `degradado` |

Estados propios de la V2:

- `decision`: `respaldo` si el motor falló, se pasó del timeout `DECISION_TIMEOUT_MS`, eligió fuera de las opciones o se
  agotó `MAX_DECISION_CALLS` (decidieron las reglas); `omitida` si ninguna decisión fue dudosa.
- `quality_gate`: `ok` (pasó a la primera), `alerta` (pasó al regenerar sin LLM), `error` (falló dos veces: respuesta
  segura); `omitida` en aclaraciones, SIN_EVIDENCIA y charla (no afirman nada del ERP).
- `renderizado`: `respaldo` si cayó el LLM (plantillas) o salió la respuesta segura.
- `plan_consulta`: `alerta` si se agotaron las búsquedas sin evidencia; `no_tomada` con la memoria («¿y luego?») o la
  charla.

Oportunidades V2 (reglas fijas, numeradas a continuación de las de V1):

- `tipo_respuesta` con `dudoso: true` → `alerta` (ni regla clara ni kNN con margen: decidió el motor).
- `decision` con `respaldo: true` → `alerta`.
- `busqueda_vectorial` o `rerank` en `respaldo` → `alerta` (degradado: solo léxica / puntaje híbrido).
- `plan` con `sin_evidencia: true` → `alerta`.
- `quality_gate` con `passed: false` → `error`.
- `renderizado` en `respaldo` o `error` → `alerta`.

### Petición y respuesta de `/ask` en V2

Petición (todos opcionales; leídos en crudo: un valor de otro tipo se ignora, nunca da 400):

- `"version": "v1" | "v2"`: solo cuenta con `AGENT_V2_ENABLED=true` (chat de prueba). Apagada, se ignora y la respuesta
  es V1 byte a byte.
- `"conversacion": "<id>"`: clave del porcentaje `AGENT_V2_PERCENTAGE` (sin ella: el primer mensaje del usuario en el
  hilo, o la pregunta).
- `"memoria": {procedimiento_id, paso_actual, pasos_completados, suspendido, retomado}`: la que devolvió el turno
  anterior. Sin ella, la V2 la reconstruye del último `plan` de un turno `asistente` del `hilo` (cada turno acepta
  `"plan"` opcional; V1 lo ignora).

Respuesta en un turno V2: los campos de V1 (`respuesta`, `modo`, `orquestacion` con `ruta: "v2"`, `fuentes`,
`sin_contexto`, `motivo`…) más `plan` (`tipos.Plan`: `pasos` trae TODOS los pasos y `pasos_mostrados` la parte que se
enseña; cada foto va en su paso), `memoria` (siempre, aunque sea `{}`) y `"version": "v2"`. En V1 esas tres claves no
aparecen. `GET /health` añade `"v2": {version, habilitada, porcentaje, motor}` solo si la V2 está construida.
