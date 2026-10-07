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

El modelo en `metrin/modelos/router/` es el **v2, con preguntas reales** (sha256 del `.pjge` `0c1f9b1d…`).

| Prueba (router solo, int8) | v1: solo sintético | **v2: + reales** |
|---|---|---|
| Real, 472 mensajes sin caso: se abstiene | 34,5 % | **77,3 %** |
| Real, 472 sin caso: elige un procedimiento por error | 1,5 % | **0,85 %** |
| Real, 6 con caso: procedimiento correcto 1.º | 5/6 | 6/6 |
| Sintética, 125 con caso: procedimiento correcto 1.º | 89,6 % | 88,0 % |
| Sintética, 125 con caso: elige y se equivoca | 4,0 % | **1,6 %** |

La prueba real con caso son 6 preguntas: sirve de indicio, no de cifra.

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
