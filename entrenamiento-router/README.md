# Router de casos de Metrín

Clasificador que recibe la pregunta y elige el procedimiento de `kb/procedimientos/` (el «caso del tutorial»), con el
árbol de decisión de [`docs/TOC-RUTEO-METRIN.md`](../docs/TOC-RUTEO-METRIN.md): módulo primero, abstenerse antes que
equivocarse y, si dos casos empatan, preguntar. En Go vive en `metrin/internal/router` y entra en la V2 con
`V2_ROUTER` (`metrin/internal/v2/router.go`).

## El modelo

Es el fine-tuning del modelo de embeddings estático que Metrín ya usa (`minishlab/potion-multilingual-128M`, revisión
`73908c3`), recortado al vocabulario de S10, más una cabeza lineal:

```
v = media de las filas de los tokens de la pregunta, norma L2      (= embed.Modelo.Embeber en Go)
p = softmax(W·v + b)                                               (60 procedimientos + «_ninguno»)
```

Se entrenan a la vez las filas del modelo estático y la cabeza. Se exporta en el formato PJGE que Go ya lee, en int8,
así que corre en el EC2 sin GPU ni dependencias.

| Archivo (en `metrin/modelos/router/`) | Qué es |
|---|---|
| `router-s10.pjge` | Embeddings afinados, int8 (14,5 MB, va en git por excepción en `.gitignore`) |
| `router-s10.cabeza.json` | Clases, módulo de cada clase, `W`, `b`, umbrales calibrados y sha256 del `.pjge` |
| `router-s10.paridad.jsonl` | 40 frases con embedding y probabilidades de Python, para la prueba de paridad en Go |
| `router-s10.json` | Procedencia y tamaño |

## Datos

`construir_datos.py` arma los conjuntos:

- **Entrenamiento:** título, `preguntas` y `aliases` de cada YAML, más `datos/parafrasis/*.jsonl`: 36 preguntas por
  procedimiento (PROCEDURE, NAVIGATION, CONFIGURATION y TROUBLESHOOTING, con faltas de ortografía y frases cortas) y 420
  preguntas «_ninguno» (conceptos, temas sin procedimiento y casi-vecinos como «pago con criptomonedas»). Las escribieron
  agentes leyendo cada YAML el 07-10-2026. Cada pregunta se duplica en minúsculas y sin tildes.
- **Preguntas reales de clientes** (desde el 07-10-2026): 4 exportaciones de grupos de WhatsApp de soporte S10.
  `extraer_reales.py` saca los mensajes de clientes y los anonimiza (teléfonos, documentos, correos, menciones, nombres
  de participantes); 1 292 candidatas a pregunta, etiquetadas por agentes en procedimiento / «_ninguno» /
  «no_pregunta». Resultado: solo **17** corresponden a uno de los 60 procedimientos; 206 son preguntas de S10 sin
  procedimiento y 1 069 no son preguntas de uso (accesos, «su apoyo», estados). Corte temporal por grupo: 60 % más
  antiguo para entrenar y calibrar, 40 % más reciente como prueba real (478 mensajes, 6 con caso).
  **Todo eso vive en `reales/` y en los `datos/*.jsonl` generados, fuera de git** (repo público). Los textos reales
  tampoco deciden el vocabulario del `.pjge` ni entran en el archivo de paridad.
- **Calibración:** 15 % de las paráfrasis, separado por clase, más una quinta parte de las reales de entrenamiento.
  Aquí se eligen los umbrales del árbol.
- **Prueba:** el primer turno de los 194 casos de `metrin/eval/v2_oro.jsonl`. Nunca se usa para entrenar ni calibrar:
  se descarta toda pregunta igual o con ≥ 80 % de palabras en común con una de la prueba.

## Reproducir

```bash
cd entrenamiento-router
../../.venv/bin/uv venv -p 3.13 .venv
../../.venv/bin/uv pip install -p .venv/bin/python "model2vec[train]==0.9.0" pyyaml scikit-learn
.venv/bin/python -c "from huggingface_hub import snapshot_download; snapshot_download('minishlab/potion-multilingual-128M', \
  revision='73908c3438cf03b6a01bcb9611d62b23d0726f08', local_dir='base/potion-multilingual-128M')"
# opcional, con las exportaciones de WhatsApp en reales/crudo/<grupo>/*.txt y sus etiquetas en reales/etiquetas*.jsonl:
.venv/bin/python extraer_reales.py
.venv/bin/python construir_datos.py
.venv/bin/python entrenar.py            # ~4 min en la Mac; deja todo en salida/
.venv/bin/python evaluar_modelo.py salida reales/prueba_real.jsonl datos/prueba.jsonl
cp salida/router-s10.{pjge,json,cabeza.json,paridad.jsonl} ../metrin/modelos/router/
(cd ../metrin && go test ./internal/router/ ./internal/v2/)
bash medir-v2.sh                        # desde s10-conocimiento/: benchmark de la V2 con y sin router
.venv/bin/python medir_reales.py        # con los contenedores de medir-v2.sh: la prueba real por la V2 completa
```

`destilar.py` es copia sin cambios del de `pjgfarma-analista`; `entrenar.py` importa de él el recorte, la cuantización
y la escritura PJGE.

## Resultados (07-10-2026)

El modelo en `metrin/modelos/router/` es el **v9: 218 procedimientos sobre MiniLM destilado** (sha256 del `.pjge` `d3a3e03c…`, 7,2 MB), con los umbrales de castigo 3 más la regla de ventaja (`ratio` 2,5, `acepta_min` 0,3) y el filtro de uso apagado. Las tablas de abajo cuentan cómo se llegó.

| Prueba (router solo, int8) | v1: solo sintético | **v2: + reales** |
|---|---|---|
| Real, 472 mensajes sin caso: se abstiene | 34,5 % | **77,3 %** |
| Real, 472 sin caso: elige un procedimiento por error | 1,5 % | **0,85 %** |
| Real, 6 con caso: procedimiento correcto 1.º | 5/6 | 6/6 |
| Sintética, 125 con caso: procedimiento correcto 1.º | 89,6 % | 88,0 % |
| Sintética, 125 con caso: elige y se equivoca | 4,0 % | **1,6 %** |

La prueba real con caso son 6 preguntas: sirve de indicio, no de cifra.

**v3 (07-10-2026, 7 grupos de WhatsApp, 1 849 candidatas): no se instaló.** Los 3 grupos nuevos aportaron 557
candidatas y solo 2 con procedimiento. Sobre la misma prueba real ampliada (684 mensajes, 6 con caso) y la sintética:

| | v2 (instalado) | v3 |
|---|---|---|
| Real, 678 sin caso: se abstiene | 77,3 % | 78,8 % |
| Real, 678 sin caso: elige por error | 1,33 % | 1,62 % |
| Real, 6 con caso: correcto 1.º | 6/6 | 6/6 |
| Sintética, 125 con caso: elige y se equivoca | **1,6 %** | 4,0 % |

Más negativos no mejoran: lo que falta son preguntas con caso, y los clientes casi no hacen preguntas que el catálogo
cubra. El v3 queda en `salida/v3-7grupos/` (fuera de git).

Detalle del v1 (125 preguntas con caso y 69 sin caso de `v2_oro.jsonl`):

| | Sin afinar (cabeza sola) | Afinado | Afinado int8 (lo que corre Go) |
|---|---|---|---|
| Procedimiento correcto 1.º | 86,4 % | **89,6 %** | 89,6 % |
| Entre los 3 primeros | 96,8 % | 98,4 % | 98,4 % |
| Módulo correcto | 98,4 % | 97,6 % | 97,6 % |
| Elige y acierta | 62,4 % | 69,6 % | 69,6 % |
| Elige y se equivoca | 2,4 % | 4,0 % | 4,0 % |
| Pregunta, con la opción correcta | 33,6 % | 24,8 % | 24,8 % |

Referencia de la V2 completa: 63,2 % de acierto de procedimiento sin reranker (lo que da el EC2) y 76,8 % con
reranker (solo la Mac). Las preguntas «sin caso» de la prueba son sobre todo de concepto y palabras sueltas: en la V2
no llegan al router, porque el clasificador de tipo las manda antes a conceptos o a aclaración.

Dentro de la V2: ver `docs/TOC-RUTEO-METRIN.md` §8.

### v4 con tickets de Zendesk (07-10-2026, instalado)

`extraer_tickets.py` toma los 1 005 tickets exportados en `optimiza360_scraping/` (mesa de ayuda de Edifica, 10-2024 a
10-2026; asunto + primer mensaje, anonimizados, en `reales/tickets/`, fuera de git). Etiquetados: **3** con
procedimiento, 55 de uso de S10 sin procedimiento, 947 que no son preguntas de uso (el 75 % no es S10: impresoras,
Microsoft, laptops). Mismo corte temporal: los 398 tickets más recientes entran en la prueba real (1 082 mensajes).

| Router solo, int8 | v2 | **v4** | v4, otras 2 semillas |
|---|---|---|---|
| Real, 1 076 sin caso: se abstiene | 79,9 % | 79,9 % | 78,2–80,8 % |
| Real, 1 076 sin caso: elige por error | 0,84 % | **0,56 %** | 0,56–1,02 % |
| Real, 6 con caso: elige bien | 2/6 | **3/6** | 2–3/6 |
| Sintética, 125 con caso: elige y se equivoca | 1,6 % | 4,8 % | 3,2–4,8 % |

El 1,6 % del v2 parece una semilla favorable: v1, v3 y las semillas del v4 dan 3,2–4,8 %. Dentro de la V2 completa
(`medir-v2.sh`, sin reranker) el v4 da acierto de procedimiento 74,4 % (igual que el v2), PAS 67,3 % (v2: 65,5 %) y
falsa abstención 2,5 % (v2: 1,8 %); con los 1 082 mensajes reales por la V2 acierta 3 de 6 con caso (sin router: 1) y
responde un procedimiento al 4,3 % de los mensajes sin caso (sin router: 3,3 %).

### v5 con 125 procedimientos (07-10-2026, instalado)

El catálogo pasó de 60 a 125 (10 de `_reserva/` + 55 redactados desde los manuales oficiales, priorizando las secciones
a las que apuntan preguntas reales sin caso). 36 paráfrasis por procedimiento nuevo; los negativos sintéticos se
revisaron contra el catálogo nuevo (`_ninguno.revisado.jsonl`, solo cambios hacia procedimientos nuevos) y 26 mensajes
reales «sin caso» pasaron a un procedimiento nuevo (`reales/reetiquetas.jsonl`, fuera de git).

| V2 completa (`medir-v2.sh`), sin reranker | v4 · 60 proc. | **v5 · 125 · castigo 3** | v5 · 125 · castigo 10 |
|---|---|---|---|
| Acierto de procedimiento (oro sintético) | 74,4 % | **74,4 %** | 61,6 % |
| PROCEDURAL ANSWER SUCCESS | 67,3 % | 61,8 % | 47,3 % |
| Falsa abstención | 2,5 % | 1,8 % | 2,5 % |
| Sin router, mismo catálogo | 63,2 % | 54,4 % | 54,4 % |

Router solo: en los procedimientos nuevos (calibración) elige bien el 62 % y tiene el correcto entre los 3 primeros el
95 %; con mensajes reales sin caso elige por error el 1,78 % (v4: 0,47 %). Con castigo 10 ese error baja a 0,93 %, pero
la V2 pregunta en vez de responder en 39 de 125 casos. Se eligió castigo 3. Más procedimientos también le cuestan a la
búsqueda sin router (63,2 → 54,4 %): el router es lo que mantiene el acierto.

### Ronda de mejora (08-10-2026): medición nueva, filtro de uso y aclaración

- **Medición.** `datos/prueba_nuevos.jsonl`: 160 preguntas escritas sin ver el entrenamiento (2 por cada uno de los 65
  procedimientos nuevos + 30 sin caso); `construir_datos.py` excluye del entrenamiento lo que se le parezca. Correcciones
  de etiqueta para la prueba del router en `CORRECCIONES_ORO` y `datos/prueba_revision.jsonl` (`v2_oro.jsonl` no se toca).
- **Resultado con el v5 instalado.** Router solo: elige bien el 88 % de los nuevos, 0 % mal, el correcto entre los 3
  primeros el 100 %. Dentro de la V2 completa: 115/130 bien (88 %), 11 ofrecidos en la aclaración, 0 equivocados; en las
  30 sin caso responde un procedimiento en 6.
- **Diagnóstico del 61,8 % de PAS** (v4: 67,3 %): ninguna elección nueva equivocada; 5 respuestas pasaron a aclaración
  porque ahora hay un vecino (copia de seguridad / actualizar base de datos, cálculo de nómina / renta de quinta,
  polinómica / registrar presupuesto). Varias son dudas legítimas.
- **Filtro «¿es pregunta de uso?»** (`entrenar_filtro`, cabeza `filtro`, `Umbrales.Uso` en Go): rechaza el 88 % de la
  coordinación sintética, pero con mensajes reales, para no callar preguntas de verdad, solo el 18–28 %, y no cambia el
  1,8 %. Ese error viene de preguntas reales de uso cuyo tema no tiene procedimiento, no de la coordinación. Queda en el
  código, sin activar (el v5 no lo trae).
- **Elegir por ventaja** (`ratio`, `acepta_min`): la calibración no la elige. Queda disponible.
- **Aclaración por módulo** ofrece ahora los 3 procedimientos más probables (`internal/v2/router.go`). En el benchmark no
  cambia nada medible.

Siguiente palanca para el 1,8 %: más procedimientos para los temas reales sin cobertura, o un embebedor mejor (MiniLM
destilado a PJGE, punto 4 del plan).

### v7 con 167 procedimientos (08-10-2026, instalado)

Tanda 2 del escenario B1: 42 procedimientos más desde secciones oficiales accionables aún sin usar (de 155 candidatas;
quedan ~110). Nuevos módulos con contenido: Calidad Móvil (carpeta nueva). Prueba ciega `datos/prueba_nuevos2.jsonl`
(99: 2 por procedimiento + 15 sin caso) y revisión de `prueba_nuevos.jsonl` (`nuevo-149` → PLAME).

| Router solo, mismos umbrales, sin filtro | v5 · 125 | **v7 · 167** |
|---|---|---|
| Oro antiguo (129 con caso): elige bien / mal | 66,7 % / 3,1 % | 63,6 % / 3,1 % |
| Procedimientos tanda 1 (131): elige bien / mal | 87,0 % / 0,8 % | 85,5 % / 0,0 % |
| Procedimientos tanda 2 (84): elige bien / mal / top-3 | 0 % / 11,9 % / 0 % | **81,0 % / 1,2 % / 97,6 %** |
| Sin caso de la tanda 2 (15, 8 casi-vecinas): elige | 26,7 % | 53,3 % |
| Real (12 con caso / 1 070 sin caso) | 2 bien, 1,8 % error | idéntico |

`entrenar.py` exporta siempre el filtro de uso, pero solo lo activa con `--filtro`: calibrado con datos sintéticos
bloqueaba 9 de las 12 preguntas reales con caso. La V2 completa (`medir-v2.sh`) no se midió en esta tanda: el disco de
la Mac estaba lleno (108 MiB libres) y la imagen Docker no cabía.

### v8: MiniLM destilado en lugar de potion (08-10-2026)

`sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2` destilado a estático con model2vec 0.9.0 (PCA 256; mismo
formato PJGE, mismo código Go). Reproducir:

```bash
.venv/bin/python -c "from model2vec.distill import distill; distill(model_name='sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2', pca_dims=256).save_pretrained('base/minilm-l12-pca256')"
.venv/bin/python entrenar.py --base base/minilm-l12-pca256 --procedencia "paraphrase-multilingual-MiniLM-L12-v2 (model2vec, PCA 256)"
```

**Trampa encontrada:** el tokenizador de MiniLM trae relleno (*padding*) y recorte activos; `encode_batch` metía tokens
de relleno en la media y Go no (la prueba de paridad falló). `entrenar.py` y `evaluar_modelo.py` los desactivan siempre.

| V2 completa (`medir-v2.sh`) | v7 · potion | **v8 · MiniLM** |
|---|---|---|
| Acierto de procedimiento (oro antiguo) | 71,2 % | **72,8 %** |
| PROCEDURAL ANSWER SUCCESS | 60,0 % | **61,8 %** |
| Falsa abstención / abstención correcta | 2,5 % / 83,3 % | **1,8 % / 91,7 %** |
| Reales sin caso: responde un procedimiento | 6,1 % | **5,2 %** |
| Tanda 1 (131): bien + ofrece / equivocadas | 113 + 12 / **0** | 106 + 19 / 2 |
| Tanda 2 (84): bien + ofrece / equivocadas | 70 + 9 / 2 | **74 + 6 / 1** |

Router solo (mismos umbrales): oro 68,2 % frente a 63,6 %; reales sin caso, error 0,8 % frente a 1,8 %; tanda 1 80,2 %
frente a 85,5 %. Diferencias pequeñas, casi todas a favor de MiniLM, con el modelo a la mitad de tamaño.

### v9 con 218 procedimientos y regla de ventaja (08-10-2026, instalado)

La tanda 3 convierte 51 secciones más de los manuales en procedimientos: 29 de gerencia de proyectos, 9 de almacenes, 7
de presupuestos, 3 de facturación y 3 de nóminas (`datos/grupos_nuevos3.json`). Se hizo igual que la tanda 2:
- 36 paráfrasis por procedimiento (`datos/parafrasis/nuevos3-*.jsonl`).
- Una prueba ciega de 117 preguntas (`datos/prueba_nuevos3.jsonl`).
- La revisión de los «sin caso» anteriores:
  - 5 de las pruebas de las tandas 1 y 2 (`prueba_nuevos_revision3.jsonl`).
  - Ninguno del oro (`prueba_revision3.jsonl`, vacío).
  - 13 negativos sintéticos (`revisado: cambiado_t3`; se descartaron 3 dudosos).
  - 2 reales.

`entrenar.py` y `construir_datos.py` leen ahora todas las revisiones de cada tanda (`prueba_nuevos_revision*.jsonl` y
`prueba_revision*.jsonl`).

**Con los umbrales de v8 la V2 empeoraba.** El acierto de procedimiento bajaba del 72,8 % al 65,6 % y el PAS del 61,8 %
al 50,9 %. El router seguía poniendo primero el procedimiento correcto en 10 de los 11 casos del oro que se perdían.
Con 218 clases la probabilidad se reparte más, el primero quedaba por debajo de `acepta` (0,6) y el árbol pedía
aclaración. La regla de ventaja (elegir si el primero saca 2,5 veces al segundo y pasa de 0,3) lo corrige. Se eligió
con un barrido sobre las pruebas del router solo:

| Router solo: elige bien / elige mal | umbrales v8 | **ratio 2,5, mín. 0,3** |
|---|---|---|
| Oro (129 con caso) | 56,6 % / 1,5 % | **76,0 % / 3,9 %** |
| Tandas 1–3 (322 con caso) | 71,7 % / 0,0 % | **82,3 % / 0,9 %** |
| Reales sin caso (1.069): elige alguno | 0,7 % | 1,9 % |

| V2 completa (`medir-v2.sh`) | v8 · 167 | v9 · umbrales v8 | **v9 · ratio** |
|---|---|---|---|
| Acierto de procedimiento | 72,8 % | 65,6 % | **76,0 %** |
| PROCEDURAL ANSWER SUCCESS | 61,8 % | 50,9 % | **65,5 %** |
| Falsa abstención / abstención correcta | 1,8 % / 91,7 % | 2,5 % / 91,7 % | **1,8 % / 91,7 %** |
| Reales sin caso: responde un procedimiento | **5,2 %** | 5,1 % | 5,5 % |
| Tanda 1 (134): bien + ofrece / equivocadas | 106 + 19 / 2 (de 131) | 98 + 29 / 3 | **108 + 19 / 3** |
| Tanda 2 (86): bien + ofrece / equivocadas | 74 + 6 / 1 (de 84) | 68 + 13 / 2 | **74 + 7 / 2** |
| Tanda 3 (102): bien + ofrece / equivocadas | — | 83 + 16 / 1 | **91 + 8 / 1** |

La V2 sin router también baja con el catálogo grande: el PAS pasa del 58,2 % al 52,7 %. Hay más procedimientos
parecidos que compiten en la búsqueda. Errores que quedan en el oro: «programación de avance por periodos» va a
`registrar-avances-en-la-rama-meta` en vez de `planificar-en-cronograma-por-periodos`. `cfg-002` (configurar la fórmula
polinómica) elige `elaborar-formula-polinomica`, y el oro espera `registrar-presupuesto-nuevo`.


Al reentrenar, `entrenar.py` vuelve a calibrar los umbrales y deja `ratio` en 0. La regla de ventaja se pone a mano en
`router-s10.cabeza.json`, que no cambia la huella del `.pjge`.
