# Plantillas de procedimiento de Metrín — esquema

Metrín es un asistente **instruccional**: enseña a hacer una tarea del ERP S10 paso a paso y cita la fuente.
Estas plantillas fijan, para cada tarea, **qué** se dice (los pasos, verificados contra el manual oficial);
el modelo solo decide **cómo** decirlo (orden de las frases, tono, cuánto resumir). Es el mismo reparto que
funcionó en kddesign (`ddesign-k-v2/agente/app/v2/plantillas.yaml`): el código elige la plantilla y sus datos,
el modelo no puede añadir hechos.

Uso previsto en Metrín:

1. `preguntas` y `aliases` de todas las plantillas forman el **catálogo de intenciones**: la consulta del usuario
   se compara con ellas (léxico o embeddings) y, si hay una plantilla clara, se responde con ella en vez de con
   fragmentos sueltos recuperados por similitud. `entidades` sirve para desambiguar («metrado» en Presupuestos
   frente a «avance» en Gerencia de Proyectos) y para filtrar por módulo o pantalla.
2. Al modelo se le entrega la plantilla ya elegida. Puede cambiar la redacción, nunca los nombres en **negrita**
   (menús, botones, campos), el orden de los pasos ni los datos. Una respuesta que nombre un menú que no está
   en la plantilla se descarta y se muestra la plantilla tal cual.
3. Cada paso lleva sus `fuente`; la respuesta muestra las citas (manual y sección) a partir de la tabla `fuentes`,
   y las `fotos` del paso, que la propia fuente asocia a ese paso.

## Ubicación y formato

- Un archivo por procedimiento: `kb/procedimientos/<modulo>/<id>.yml`. El `id` es igual al nombre del archivo.
- `INDICE.json` lo genera el validador (`--indice`); no se edita a mano. Por cada id trae `titulo`, `modulo`,
  `nivel`, `n_pasos` (con subpasos), `preguntas`, `aliases`, `entidades` y `archivo`.
- `_reserva/<modulo>/<id>.yml`: plantillas ya validadas que no están activas (no entran al índice; el validador
  las ignora salvo con `--reserva`). Para activar una, se mueve a `<modulo>/`. Cualquier carpeta que empiece por
  `_` se trata igual.
- Formato: YAML en un **subconjunto fijo**, para que se pueda leer incluso sin PyYAML:
  textos siempre entre comillas dobles (escapes JSON), listas de ids y de entidades en línea (`[a, b]`), entradas
  de `fuentes` y `revision` como mapas en línea (`{id: …, manual: …}`), `fotos` como lista de mapas en bloque,
  sangría de 2 espacios, sin anclas, sin `|` ni `>`, sin comillas simples. Los ids que empiezan por dígito
  (hashes) van entre comillas.
- Trato: **usted**, en imperativo («Vaya», «Elija», «Registre»), como los manuales oficiales de S10 y como los
  tutoriales de `tutoriales/`. Se usa en todas las plantillas; no se mezcla con «tú».
- Español neutro. Una acción por paso (un paso puede agrupar lo que en pantalla es un solo gesto: «haga clic
  derecho y elija **Nuevo**»).

## Campos

| Campo | Obligatorio | Tipo | Regla |
|---|---|---|---|
| `id` | sí | texto | `<modulo>.<tarea>` en kebab-case; estable (no se renombra: es la clave de la intención). |
| `version` | sí | entero ≥ 1 | Se sube cuando cambian los pasos. |
| `modulo` | sí | cerrado | `presupuestos` · `gerencia-proyectos` · `almacenes` · `compras` · `nominas` · `contabilidad` · `facturacion` · `administrativo` · `facturacion-electronica` · `portal-proveedor` · `tareo-movil` · `calidad-movil` · `instalacion`. Es también la carpeta. |
| `titulo` | sí | texto | Infinitivo: «Registrar un presupuesto nuevo». |
| `objetivo` | sí | texto | Para qué sirve, una frase, con lo que dice el manual. |
| `nivel` | sí | cerrado | `basico` · `intermedio` · `avanzado`. |
| `preguntas` | sí | lista, 6–12 | Formas reales de pedirlo, incluidas coloquiales y con faltas («prespuesto», «como ago»). Sin repetidas. |
| `aliases` | sí | lista, 2–10 | Formas cortas de nombrar la tarea: «registrar metrado», «ingresar metrado», «cargar metrado». |
| `entidades` | sí | `{modulo, pantallas, objetos}` | `modulo`: nombre del módulo («Presupuestos», «Gerencia de Proyectos», «Nóminas»…; fijo por `modulo`). `pantallas`: escenarios, ventanas o catálogos donde ocurre. `objetos`: cosas del ERP que maneja (presupuesto, partida, orden de compra…; ≥ 1). Cada pantalla y objeto **aparece en alguna de sus fuentes**. |
| `prerrequisitos` | sí (puede ser `[]`) | lista de `{texto, fuente}` | Lo que debe existir antes (usuario, catálogo, configuración). |
| `pasos` | sí, ≥ 1 | lista de pasos | Ver abajo. |
| `verificacion` | sí (puede ser `[]`) | lista de `{texto, fuente}` | Cómo sabe el usuario que lo hizo bien, **según la fuente**. Si la fuente no lo dice, lista vacía. |
| `errores_frecuentes` | sí (puede ser `[]`) | lista de `{sintoma, solucion, fuente}` | Solo mensajes, restricciones o advertencias que la fuente menciona. |
| `relacionados` | sí (puede ser `[]`) | lista de ids | Deben existir. |
| `fuentes` | sí, ≥ 1 | lista de `{id, manual, seccion, pagina}` | Tabla de resolución de las citas (ver «ids ambiguos»). Cada entrada se cita al menos una vez. |
| `revision` | sí | `{generado, por, revisado_por_humano}` | `revisado_por_humano: false` hasta que una persona la revise contra el ERP. |

### Paso

| Campo | Obligatorio | Regla |
|---|---|---|
| `n` | sí | 1, 2, 3… consecutivos (también dentro de `sub`). |
| `id` | sí | `<id del procedimiento>#<n>`; en subpasos `#<n>.<m>`. Estable: sirve para enlazar fotos, trazas y comentarios. |
| `accion` | sí | Imperativo, una acción. Los nombres de menús, botones, campos, pestañas y ventanas van entre `** **` **exactamente como en la fuente** (se admite otra capitalización y la falta de tildes). |
| `donde` | no | Menú, escenario o ventana donde ocurre. |
| `resultado` | no | Lo que se ve después, **solo si la fuente lo dice**. |
| `fuente` | sí | 1+ ids de `kb/fragmentos*.jsonl` que respaldan el paso. El primero es el principal. |
| `pagina` | sí | Página del primer fragmento fuente (`null` en los manuales web, que no tienen páginas). |
| `fotos` | no | Lista de fotos (ver abajo). Se omite si la fuente no asocia ninguna foto a este paso. |
| `sub` | no | Subpasos, con la misma forma. |

### Foto

| Campo | Obligatorio | Regla |
|---|---|---|
| `id` | sí | Hash corto estable: los 12 primeros caracteres hex de `sha1(ruta_o_url)`. |
| `ruta_o_url` | sí | Tal como figura en el fragmento: `imagenes/<manual>/<hash>.png` (servida desde `data/imagenes/`) o la URL. |
| `pagina` | sí | Página del fragmento que trae la foto (`null` en web). |
| `caption` | no | Solo si la fuente lo trae: el `pasos[k].texto` al que la sección asocia esa foto, literal. No se pone cuando el paso cita directamente el fragmento imagen (entonces la asociación es por su OCR, y el texto de la sección podría describir otro paso). |
| `ocr` | no | Solo si existe el fragmento imagen de esa ruta: su texto OCR, literal, recortado a 400 caracteres (`…`). Es ruidoso: sirve para buscar, no para citar. |

**Regla de asociación.** Una foto va en un paso **solo si la fuente la asocia a ese paso**: está en
`pasos[k].fotos` de una sección que el paso cita y ese `pasos[k].texto` corresponde a lo que el paso pide, o el
paso cita directamente el fragmento imagen (porque su OCR es lo que nombra el botón). Una misma foto no se repite
en dos pasos del procedimiento. Ante la duda, el paso va sin foto: es preferible un paso sin foto que una foto
equivocada.

## Fuentes

### Qué se puede citar

- **Primero, los manuales oficiales** de `documentacion.s10peru.com` (fragmentos `web-…` con `seccion` y `pasos`,
  `confianza: oficial`) y el **OCR de sus capturas** (fragmentos `img-<hash>-0000`, también oficiales): muchas
  secciones dicen «use el botón» sin nombrarlo, y el nombre solo está en la captura. También valen las demás
  guías oficiales de ese portal (p. ej. «Cómo realizar una copia de seguridad de base de datos»).
- Como complemento, y solo en Presupuestos, los dos PDF importados a mano (`Guia de Usuario de S10 Presupuestos`,
  `Manual de S10 Costos y Presupuestos`; `confianza: tercero-sin-verificar`, ids `7c40a9548710-…` y `c8ac2722518e-…`)
  cuando nombran el botón que la sección oficial calla. Siempre junto a la sección oficial equivalente.
- `fragmentos_pantallas.jsonl`, `fragmentos_faq.jsonl` y `fragmentos_yt.jsonl` son secundarios: se pueden citar,
  pero ninguna plantilla de esta versión los necesitó.
- **Prohibido**: páginas de marketing (`s10peru.com (público)`, `Optimiza 360 · …`, menús web, «Mes: …»).
  El validador rechaza `Optimiza 360` y `s10peru.com (público)`.

### ids ambiguos (importante)

En `kb/fragmentos.jsonl` el id de las secciones web se arma con el slug del manual **truncado a 8 caracteres**:
`web-https-documentacion-s10peru-com-manual-d-s032` es a la vez la sección s032 del manual de Presupuestos,
de Almacenes, de Compras, de Nóminas… **200 ids se repiten y afectan a 1.061 fragmentos.** El par
`(id, manual)` sí es único.

Por eso cada plantilla lleva la tabla `fuentes`: los `fuente` de pasos, prerrequisitos, verificaciones y
errores llevan solo el id, y la tabla dice de qué manual es. Quien consuma la plantilla (Metrín) debe resolver
cada cita con el par `(id, manual)` de la tabla, **nunca solo con el id**. Dentro de una misma plantilla no
puede citarse el mismo id de dos manuales distintos. (Arreglar el id en el indexador —usar el slug completo—
eliminaría el problema; no se ha hecho porque `kb/fragmentos*.jsonl` no se toca desde aquí.)

## Lo que comprueba el validador

`python3 herramientas/validar_procedimientos.py` (usa PyYAML si está —en `.venv/` del repo lo está— y si no,
un lector mínimo del subconjunto de arriba). Sale con código ≠ 0 si algo falla.

1. Esquema: campos obligatorios y desconocidos, valores cerrados, 6–12 preguntas, `id` único, prefijo = módulo,
   ruta = `<modulo>/<id>.yml`, `n` consecutivos.
2. Fuentes: todo elemento con texto lleva `fuente`; cada id existe en `kb/fragmentos*.jsonl` y está en la tabla
   `fuentes`; cada entrada de la tabla existe como `(id, manual)` con esa `seccion` y `pagina`, y alguien la cita;
   ninguna es de marketing.
3. **Menús no inventados**: cada término entre `** **` aparece literal (sin distinguir mayúsculas ni tildes,
   espacios normalizados) en el texto, los pasos o el OCR de al menos uno de los fragmentos que cita ese mismo
   elemento.
4. **Soporte léxico** (ver abajo): por debajo del umbral, el elemento se lista como *dudoso*. No es error salvo
   con `--estricto`.
5. `relacionados` existen.
6. Pasos: `id` = `<procedimiento>#<n>` único, `pagina` = la del primer fragmento fuente.
7. Fotos: existen en un fragmento fuente de su paso **y** la fuente las asocia (están en `pasos[k].fotos` de una
   sección citada o son un fragmento imagen citado); `id` = `sha1(ruta)[:12]`; `caption` es el texto que la fuente
   asocia a la foto; `ocr` es un prefijo del OCR de esa imagen; `pagina` cuadra; no se repiten en dos pasos.
   Informa el **% de pasos con foto** y las **fotos dudosas**: las que comparten menos de 2 raíces de contenido
   con su paso (comparando `accion` + `resultado` con `caption` + el OCR completo de la imagen en la base).
8. `aliases` (2–10, distintos) y `entidades` (`modulo` fijo por módulo; cada pantalla y objeto aparece en
   alguna fuente de la plantilla).

Opciones: `--indice` escribe `INDICE.json` (solo si no hay errores) · `--dudosos` lista todos los dudosos ·
`--calibrar` muestra la calibración · `--modulo X` valida solo un módulo · `--reserva` incluye `_reserva/` ·
`--estricto` (los dudosos también fallan).

### Soporte léxico: método y umbral

Se toman las «palabras de contenido» del texto propio (minúsculas, sin tildes, sin las negritas, ≥ 4 letras,
sin palabras vacías ni verbos genéricos de interfaz como *haga, clic, botón, elija, ventana, opción, sistema*) y
se reducen a su prefijo de 5 letras (*registre / registrar / registro → regis*). Se cuenta cuántas aparecen
también en sus fragmentos fuente.

Un elemento pasa si comparte **al menos 2 raíces** (o todas, si tiene menos de 2) **y al menos el 60 %** de las
suyas.

Calibración (`--calibrar`): se compara cada `accion` con sus fuentes reales y con el mismo número de fragmentos
tomados al azar del mismo manual. Con los 122 primeros pasos de Presupuestos, escritos antes de fijar el umbral:

| umbral de proporción | pasan con su fuente real | pasan con una fuente al azar |
|---|---|---|
| ≥ 50 % | 100 % | 10,4 % |
| **≥ 60 %** | **99,2 %** | **7,8 %** |
| ≥ 70 % | 95,1 % | 4,3 % |

Se eligió 60 %: deja pasar casi todos los pasos bien apoyados y rechaza más de 9 de cada 10 emparejamientos con
una fuente ajena. Por encima, empieza a marcar pasos correctos que solo reordenan la frase del manual.

El mismo método con otro umbral sirve para las **fotos dudosas**: menos de 2 raíces en común entre el paso
(`accion` + `resultado`) y el texto que acompaña a la foto (`caption` + OCR completo). El OCR de las capturas
es ruidoso («Catdlogo», «Andlisis», «Sunitem»), así que una foto dudosa casi siempre es correcta; se listan
para revisarlas a ojo.

Un dudoso no significa que el paso esté mal: suele ser un paso que resume una captura (cuyo OCR es ruidoso) o
que parafrasea mucho. Están listados en `COBERTURA.md` para revisión humana.

## Cómo añadir o cambiar una plantilla

1. Lea las secciones del manual (y el OCR de sus capturas) que describen la tarea. Si no hay pasos suficientes en
   la fuente, **no se crea**: se anota en `COBERTURA.md` con el motivo.
2. Escriba el archivo siguiendo este esquema; copie los nombres de menús y botones tal como están en la fuente.
3. Corra el validador hasta que dé 0 errores; revise sus dudosos; regenere `INDICE.json` con `--indice`.
4. Suba `version` si cambió un paso. Deje `revisado_por_humano: false` hasta que alguien la pruebe en el ERP.
