# Metrín: modo traza en el chat

Estado: diseño propuesto (06-10-2026). No implementado.
Ámbito: servicio `metrin` (Go, `s10-conocimiento/metrin/`), página del chat (`internal/servidor/pagina.html`).
Público de este documento: quien implemente el modo traza y quien decida qué mejorar en Metrín.

---

## 1. Problema

Metrín responde preguntas sobre S10 con un pipeline de varias etapas: clasificación, reglas, reescritura del hilo,
búsqueda vectorial, selección de fragmentos, generación y verificación de citas. Cuando una respuesta sale mal, hoy
no se ve en qué etapa falló: la respuesta llega sin el camino que tomó. Eso impide decidir qué mejorar (el clasificador,
el umbral, la búsqueda, el prompt o el modelo).

Se pide un **modo traza**: en el chat, un panel lateral que dibuje el camino de cada mensaje por las etapas, con sus
valores técnicos, para ver de un vistazo dónde está el problema.

## 2. Usuarios y casos

| Quién | Caso |
|---|---|
| Desarrollador de Metrín | Ver por qué una respuesta citó un manual equivocado: ¿la intención fue bien?, ¿cuántos fragmentos pasaron el umbral?, ¿cuál fue su distancia? |
| Responsable de contenido S10 | Ver si una pregunta cayó en «sin contexto suficiente» porque falta un manual o porque la búsqueda no lo encontró. |
| Revisor de calidad | Comparar dos mensajes seguidos del mismo hilo y ver qué cambió la reescritura. |

No es para clientes finales. El modo traza expone puntajes internos y debe estar cerrado por defecto (ver §8).

## 3. Requisitos

### 3.1 Funcionales

- **FR-1.** Un interruptor «Traza» en la cabecera del chat (o en el menú ⋮, igual que los controles de demo de
  kddesign). Activo: se abre el panel lateral.
- **FR-2.** El panel muestra el flujo del último mensaje como diagrama vertical: una tarjeta por etapa, unida a la
  siguiente por una línea. La etapa tomada y la descartada se distinguen a simple vista.
- **FR-3.** Cada etapa muestra sus valores técnicos: entradas, salidas, decisión, umbral, tiempo en ms y el modelo o
  regla que decidió. Al tocar una tarjeta se despliega su detalle (tabla clave-valor).
- **FR-4.** Cada mensaje queda en el hilo con su traza. Se puede volver a un mensaje anterior y ver su traza sin
  repetir la consulta.
- **FR-5.** La traza sale del servicio, no se reconstruye en el navegador. El navegador solo la pinta.
- **FR-6.** Si una etapa no se ejecutó (regla que cortó el flujo, o el LLM no se llamó), aparece como «omitida» con la
  razón. Nunca se oculta.
- **FR-7.** Si una etapa falló, aparece como «error» con el mensaje de error recortado, y la respuesta refleja el
  respaldo que se usó.

### 3.2 No funcionales

- **NFR-1.** Con el modo traza apagado, la respuesta es idéntica byte a byte a la actual. La instrumentación solo
  mide; no cambia decisiones.
- **NFR-2.** Sobrecarga de la traza menor al 5 % de la latencia p50 del `/ask` (medido en el host de producción).
- **NFR-3.** La traza no contiene texto de documentos completo ni datos de identificación: solo identificadores,
  puntajes y recortes de 120 caracteres.
- **NFR-4.** Accesible: el diagrama se puede recorrer con teclado (Tab entre etapas, Enter para el detalle), tiene
  texto alternativo por etapa y respeta `prefers-reduced-motion`.
- **NFR-5.** Responsive: a 390 px el panel ocupa la pantalla completa, con las etapas en columna.

## 4. Pipeline actual (verificado en el código)

Las cifras son las del código fuente. Los valores que dependen del despliegue se indican como «configurable».

| # | Etapa | Dónde | Qué decide | Valores técnicos |
|---|---|---|---|---|
| 1 | Entrada | `rag.go` `Preguntar` | Texto y hilo | `hilo` = últimos turnos (`Opciones.Hilo`) |
| 2 | Regla de desempate sobre aciertos | `rag.go` `esPreguntaSobreAciertos` | Responde «no llevo contador» sin buscar | Corte: `modo=conversacional`, `ruta=regla` |
| 3 | Embebedor | `cmd/rag/main.go` (línea 150) | Vector de la pregunta | `EMBED_PROVIDER` = `estatico` (defecto) → `potion-es-int8.pjge`; o `ollama` → `OLLAMA_EMBED_MODEL` (defecto `nomic-embed-text`) |
| 4 | Clasificador de intención | `orquestador.go` `planificar`; `clasificar.go` `Clasificar` | Intención y similitud coseno | kNN local; `UmbralDefecto = 0.40`, calibrado en valid 22/28 y test 23/28; `clasificador` = `kNN-local:<huella>` o `reglas_seguras` si no hay modelo |
| 5 | Regla de seguridad | `orquestador.go` (líneas 55–62) | Desvía a conversación social o límite si la pregunta es corta | Solo si `!esPreguntaLarga` y la intención es `social` o `limite` |
| 6 | Ruta | `orquestador.go` | `rag` (trabajo) o `conversacion` | `ruta` = `rag` / `conversacion` / `regla` |
| 7 | Reescritura del hilo | `rag.go` `reescribir` | Pregunta efectiva con referencias resueltas | Solo si hay hilo; guarda `PreguntaReescrita` |
| 8 | Búsqueda | `almacen.go` `Buscar` | Top-k fragmentos | `K` por defecto 8 (`Preguntar`); `KCortex` = 8; filtro por `source`, `type`, `ext`; **solo vectorial**, sin BM25/FTS; `Distancia = 1 − similitud` |
| 9 | Selección | `rag.go` `seleccionarContexto` | Fragmentos que van al LLM | `ventanaRelevancia = 0.18`; `maxContextos = 5`; penalización de confianza por tipo de fuente: `tercero-sin-verificar` 0,04, `estado-2013` 0,04, `academico` 0,02 |
| 10 | Evidencia | `rag.go` (`MarcaSinContexto`) | Responde «No tengo contexto suficiente» si no hay fragmentos | Marca fija en el texto de respuesta |
| 11 | Generación | `llm/llm.go` `Chat`; `rag.go` `sistema` | Texto de respuesta | Proveedor `LLM_PROVIDER` = `ollama` (`OLLAMA_MODEL`, defecto `qwen2.5-coder:7b`) o `mlx` (`mlx_lm.server`, Qwen2.5-3B-Instruct-4bit con LoRA); conversación: `sistemaCharla`, ≤ 60 palabras |
| 12 | Verificación | `rag.go` `DiceSinContexto`, `contieneCita`, `citasDe` | Que la respuesta cite lo recuperado | Citas por documento, página y sección |
| 13 | Fotos de pasos | `servidor/fotos_pasos.go`, `servidor.go` `vincularFotos` | Imágenes de los pasos del manual | `fotosPorFuente = 8` |
| 14 | Juicio JEV (opcional) | `jev/jev.go` `Decidir`, `Choice`, `Score`, `Noul` | Decisiones evaluativas (p. ej. si un reporte es seguro de ejecutar) | Activo solo con `JEV_URL`; trazas en el visor `:4761` |
| 15 | Registro de fallos | `rag.go` `registrarFallo` | Guarda en un JSONL cada respuesta sin contexto: `fecha`, `pregunta` completa, `motivo`, `distancia_min`, `fuentes` | Activo solo si `RutaFallos` está definido |

Puntos que el modo traza debe mostrar porque hoy no se ven desde fuera: la intención con su similitud y si pasó el
umbral; la ruta elegida y por qué; cuántos fragmentos se recuperaron, cuántos sobrevivieron a la ventana y con qué
distancia; qué fuentes llevaron penalización; si hubo reescritura del hilo; qué modelo generó y cuánto tardó cada
etapa.

## 5. Diseño

### 5.1 Contrato de traza

La traza es un objeto que viaja en la respuesta de `/ask` bajo la clave `traza`, solo si se pide (`"traza": true` en la
petición o la variable `METRIN_TRAZA=1` en el servicio).

```json
{
  "traza": {
    "version": 1,
    "total_ms": "<entero>",
    "etapas": [
      {
        "id": "clasificar",
        "nombre": "Clasificador de intención",
        "estado": "ok | omitida | error | respaldo",
        "ms": "<entero>",
        "modelo": "kNN-local:<huella> | reglas_seguras",
        "entrada": { "pregunta": "<recorte 120 car.>" },
        "salida": { "intencion": "<texto>", "similitud": "<real>", "umbral": 0.40, "paso_umbral": "<bool>" },
        "razon": "<texto corto>"
      }
    ]
  }
}
```

Reglas del contrato:

- `id` es estable y sirve para que la UI dibuje la misma tarjeta siempre. Las etapas de §4 tienen su `id` fijo.
- `estado = omitida` lleva `razon` (p. ej. «regla de aciertos», «sin hilo, no hay reescritura»).
- `estado = respaldo` indica que se usó el camino de contingencia (p. ej. `error_fallback_rag`).
- Los valores numéricos son crudos: sin redondear en el servicio. La UI redondea.
- Nunca se incluye el texto completo de un fragmento: solo `documento`, `página`, `sección` y `distancia`.
- Versionado: un cambio que rompa el contrato sube `version`. La UI muestra un aviso si no reconoce la versión.

### 5.2 Instrumentación

Un recorder pasado por contexto (`traza.Recorder`) con `Inicio(id)`, `Fin(id, estado, datos)`, `Omitir(id, razon)`.

- Cuando el modo está apagado, el recorder es un no-op que no reserva memoria: NFR-1.
- Cada etapa de §4 llama al recorder en el mismo punto donde toma su decisión. No se duplica lógica: la traza
  registra lo que ya se decidió.
- Se prueba que, con el recorder encendido, la respuesta (`respuesta`, `modo`, `fuentes`, `plan`) es igual a la
  respuesta con el recorder apagado, para el mismo input.

### 5.3 Interfaz

Estructura del panel (escritorio, 360 px a la derecha; móvil, pantalla completa desde abajo):

```
┌─ Traza · última respuesta ─────────────── [×] ┐
│ ● Entrada             ms 0                    │
│   │                                           │
│ ● Regla aciertos      omitida · no aplica     │
│   │                                           │
│ ● Embebedor           potion-es-int8 · ms 2   │
│   │                                           │
│ ● Clasificador        «consulta_metrados»     │
│   │                   sim 0,61 ≥ umbral 0,40  │
│ ● Seguridad           omitida                 │
│   │                                           │
│ ● Ruta                rag                     │
│   │                                           │
│ ● Reescritura        omitida · sin hilo      │
│   │                                           │
│ ● Búsqueda            K 8 · ms 14             │
│   │                   3 de 8 dentro de 0,18   │
│ ● Selección           3 fragmentos · penaliz. │
│   │                                           │
│ ● Generación          qwen2.5-coder:7b · ms … │
│   │                                           │
│ ● Verificación        2 citas OK              │
└───────────────────────────────────────────────┘
   (valores de ejemplo de la forma; no medidos)
```

Visual:

- Tarjeta por etapa, con un punto de estado: verde = ok (`--verde-tinta` de Metrín, `#54E5E8` sobre oscuro), gris =
  omitida, ámbar `#FFC928` = respaldo, rojo = error. El color nunca es el único canal: cada estado tiene texto.
- La línea entre tarjetas sigue el camino tomado; las ramas no tomadas se dibujan punteadas.
- Paleta de Metrín (`metrin/METRIN.md`): `#FFC928` amarillo industrial, `#F47721` naranja, `#343A3D` grafito,
  `#111820` negro LED, `#54E5E8` cian. El panel no usa los tokens de PjgFactSalud: Metrín es otra interfaz.
- Tipografía del chat actual; números en tabla con cifras alineadas a la derecha.
- Sin animaciones decorativas. Una transición de 200 ms al abrir y cerrar.

Interacción:

- Tocar una tarjeta abre su detalle debajo (acordeón). Solo una abierta a la vez.
- Botón «Copiar traza» (JSON) para adjuntarla a un reporte.
- Selector «Mensaje» arriba del panel para cambiar entre los mensajes del hilo.
- Atajo `t` para abrir/cerrar cuando el foco no está en el campo de texto.

### 5.4 Diagrama de flujo de datos

```
usuario ──► /ask (traza=true) ──► Preguntar
                                    │
          ┌─────────────────────────┴───────────────────────────┐
          │ recorder.Inicio/Fin por etapa                       │
          ▼                                                     │
   [1 Entrada]→[2 Regla]→[3 Embebedor]→[4 Clasificador]→[5 Seg.]→[6 Ruta]
                                                                  │
                    ┌─────────────────────────────────────────────┤
                    ▼ conversacion                                ▼ rag
           [conversar: LLM ≤60 palabras]          [7 Reescritura]→[8 Búsqueda]
                    │                                             →[9 Selección]
                    │                                             →[10 Evidencia]
                    │                                             →[11 Generación]
                    │                                             →[12 Verificación]
                    │                                             →[13 Fotos]
                    └──────────────┬──────────────────────────────┘
                                   ▼
                   respuesta + traza JSON ──► panel lateral (pinta, no recalcula)
```

## 6. Seguridad y privacidad

- **Permisos.** El modo traza está apagado por defecto. Para encenderlo en producción se requiere `METRIN_TRAZA=1`
  y que la petición venga de una sesión con rol de desarrollo. Sin rol, la petición con `traza=true` se ignora en
  silencio.
- **Datos.** El registro de fallos (`registrarFallo`) ya guarda la pregunta completa en un JSONL cuando `RutaFallos`
  está definido. La traza recorta la pregunta a 120 caracteres y no agrega datos nuevos sobre la persona. Antes de
  implementar, decidir si el registro de fallos también debe recortarse (ver §9, pregunta 2).
- **Multi-organización.** Los fragmentos de otra organización no deben aparecer en la traza. Se prueba con un
  fragmento de prueba de otro tenant que nunca debe listarse.
- **Cabeceras.** La traza no se cachea: `Cache-Control: no-store`.

## 7. Rendimiento

- Medir antes y después en el host de producción: p50 y p95 de `/ask` con traza apagada y encendida, con las mismas
  300 preguntas del set de evaluación. Umbral de aceptación: sobrecarga < 5 % en p50.
- La traza se arma en memoria y se serializa solo si se pidió. Sin escritura a disco por defecto.
- Si el tamaño de la traza pasa de 16 KB, se truncan los fragmentos a los 5 primeros y se marca `truncada: true`.

## 8. Plan de implementación

| Fase | Entregable | Criterio de salida |
|---|---|---|
| F1 | `traza.Recorder` con no-op, pruebas de contrato y de igualdad de respuesta | Pruebas Go en verde; respuesta con traza apagada idéntica |
| F2 | Instrumentar las etapas 1–13 de §4 | Cada `id` aparece en la traza de un caso de cada ruta (`rag`, `conversacion`, con y sin hilo) |
| F3 | Panel lateral con diagrama, detalle y selector de mensaje | Revisión a 390 px y escritorio; navegación por teclado |
| F4 | Permisos, `Cache-Control`, prueba multi-organización | Fragmento de otro tenant no aparece |
| F5 | Medición de rendimiento y ajuste | Sobrecarga < 5 % en p50 |

Pruebas:

- Go: una prueba por ruta que verifica que cada etapa esperada está en `etapas` con su `estado`.
- Igualdad: la respuesta con y sin traza coincide para las mismas entradas.
- UI: el panel pinta una traza de ejemplo fija y cada estado tiene su texto.

## 9. Preguntas abiertas

1. ¿Quién tiene rol de desarrollo en Metrín hoy: solo el equipo técnico o también el responsable de contenido?
2. ¿La traza se guarda con el mensaje (para revisar después) o solo se ve en la sesión? Guardarla amplía lo que se
   almacena; requiere decisión.
3. ¿Hay un set de evaluación con respuestas esperadas? Sin él, la traza muestra lo que pasó, no si estuvo bien.

## 10. Lo que el modo traza permitirá decidir (hallazgos del código)

Estos puntos salen de leer el código, no de medir. El modo traza sirve para confirmarlos o descartarlos con casos
reales.

1. **Búsqueda solo vectorial.** `almacen.go` `Buscar` embebe la pregunta y consulta por similitud. No hay BM25 ni FTS.
   La especificación (`ESPECIFICACION_IA_METRIN.md` §3) pide búsqueda híbrida para términos exactos como códigos de
   partida o «metrado». La traza mostrará cuándo un código exacto no sale en el top-k.
2. **Umbrales fijos.** `UmbralDefecto = 0.40` del clasificador se calibró con 28 casos (valid 22/28, test 23/28). Y
   `ventanaRelevancia = 0.18` no tiene calibración documentada en este código. La traza permitirá ver las distancias
   reales de los fragmentos que se descartan.
3. **Modelo conversacional por defecto.** `OLLAMA_MODEL` apunta a `qwen2.5-coder:7b`, un modelo orientado a código, para
   una tarea de lenguaje en español. El camino `mlx` con Qwen2.5-3B-Instruct y LoRA existe y no es el defecto. Conviene
   medir ambos con la misma traza.
4. **Embebedor de producción.** Por defecto es el estático `potion-es-int8`. Su calidad frente a `nomic-embed-text` o
   `bge-m3` no está medida en este repositorio (la especificación lo pide en §3).
5. **Reescritura del hilo.** `reescribir` cambia la pregunta antes de buscar. Un error ahí pasa desapercibido porque la
   respuesta suena bien. La traza muestra `PreguntaReescrita` frente a la original.

## 11. Fuera de alcance

- Edición de la base de conocimiento desde el panel.
- Retroalimentación del usuario (👍/👎) y ciclo DPO: lo cubre la especificación §5, no esta fase.
- Cambiar el pipeline. Este documento solo mide y muestra.
