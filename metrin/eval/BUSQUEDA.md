# Búsqueda híbrida (vector + BM25, RRF) y reranker local para Metrín — evaluación

Fecha de la medición: 2026-10-06. Base: `kb/fragmentos*.jsonl` (8 135 fragmentos, 6 archivos), índice vectorial
construido en memoria con el indexador de producción (`indexar.KBJSONL` sobre los archivos concatenados, como
`docker-entrada.sh`: 15 277 trozos) y el embebedor estático `modelos/potion-es-int8.pjge`. Máquina: Apple M4 Pro
(12 núcleos, 24 GB), **compartida con otros agentes durante la medición**: las latencias varían entre corridas y
se dan por corrida.

Todas las tablas son copia de la salida de `go run ./cmd/evalbusqueda`; los comandos para regenerarlas están en §9.
Nada es estimado.

## 1. Resumen

Configuración recomendada (corrida D, §4.2 bis), 72 consultas:

| | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | p50 | p95 |
|---|---|---|---|---|---|---|
| (1) vector solo, como hoy | 0,201 | 0,222 | 0,332 | 0,212 | 10 ms | 13 ms |
| (2) BM25 solo | 0,637 | 0,645 | 0,765 | 0,630 | 0,5 ms | 1,8 ms |
| (3) híbrida RRF (orden final con el vector a 0,1) | 0,616 | 0,631 | 0,780 | 0,611 | 12 ms | 15 ms |
| (4) híbrida + reranker sobre el top 30 | **0,724** | **0,644** | **0,929** | **0,693** | 2,0 s | 2,3 s |

Como referencia, la RRF «de libro» con pesos iguales y sin reranker da 0,432 / 0,470 / 0,634 / 0,451.

- **El vector solo es el punto débil**: con el embebedor estático encuentra un relevante en el top 10 en menos de la
  mitad de las consultas (MRR 0,33) y en «¿qué es un metrado?» sigue poniendo primero la biografía del CEO.
- **BM25 bien hecho (BM25F, raíz ligera, difuso, atajos) triplica el Recall** del vector (0,645 frente a 0,222) y
  cuesta menos de 2 ms.
- **La fusión RRF con pesos iguales empeora a BM25**: el vector mete en la lista las FAQ de Optimiza 360 (la biografía
  del CEO, los servicios). Con el vector a 0,1 la híbrida queda a la par de BM25 en Recall y mejora el MRR (0,780).
- **El reranker es lo que más mejora la calidad**: MRR@10 de 0,780 (mejor sin reranker) a 0,929 y nDCG@10 de 0,630
  a 0,693. Al reranker le conviene el pool de **pesos iguales** (más diverso): con el vector a 0,2 en el pool el MRR
  cae a 0,890. Por eso el `Buscador` usa un peso para el pool y otro (`PesoVectorSinRerank`) para el orden final
  cuando no hay reranker o falla.
- **La expansión con alias no mejora** en ninguna variante medida (BM25: Recall@10 0,645 → 0,619 con `alias.yml`,
  → 0,592 añadiendo los procedimientos). Queda implementada y desactivada por defecto.
- Coste del reranker: ~2 s por consulta con 30 candidatos (Metal) y **~880 MB de huella física** (668 MB solo CPU).
  En la máquina de producción (Docker, sin Metal) su latencia está por medir; la fila (3) ya es una mejora grande
  por ~12 ms y 32 MiB.

## 2. Qué se construyó

Paquetes nuevos, sin dependencias Go nuevas (el `go.mod` no se tocó):

### `internal/busqueda`

- `Doc{ID, Campos map[string]string, Meta map[string]string}`; `NuevoIndice(ConfigBM25)`, `Indice.Agregar(Doc)`.
  Sirve igual para fragmentos y para procedimientos: los campos son genéricos y su peso lo fija la configuración.
- **BM25F (Okapi)** con `K1`, `B` y peso por campo (`titulo`, `seccion`, `preguntas` = 3; `entidades` = 2;
  `texto` = 1; `ocr`, `manual` = 0,5). Los postings guardan la frecuencia cruda por campo, así que k1, b y pesos
  se cambian sin reindexar (`AjustarParametros`).
- **Normalización española** (`texto.go`): minúsculas, sin diacríticos (ñ → n), palabras vacías del español más
  relleno de pregunta («hago», «quiero», «porfa»), y **raíz ligera propia** en vez de Snowball: Snowball funde
  «metrado» con «metro» y «partida» con «parte»; la raíz ligera solo quita plural, género, «-mente», «-ación»,
  «-amiento», gerundio, infinitivo y participio con longitudes mínimas (registro / registrar / registra /
  registrado comparten raíz; metrado ≠ metro, probado en `TestRaizLigera`). **n-gramas descartados**: multiplican el
  índice y la tolerancia a faltas sale más barata con la **expansión difusa** del vocabulario (Damerau-Levenshtein
  ≤ 1, ≤ 2 con 8+ letras, para raíces con df < 3 y solo hacia raíces ≥ 10 veces más frecuentes). Comprobado en el
  índice real: «presupueto», «elavoro», «obrreros», «tranferir», «cotisaciones», «avanse» y «valoriso» encuentran su
  raíz; «partdas», «presios» y «prespuesto» no, porque el OCR de las capturas ya tiene esas mismas faltas (df 3–5).
  Subir el umbral a df < 20 (× 20) las corrige también, pero **no mejoró** las 13 consultas con faltas del oro
  (BM25: Recall@10 0,564 → 0,556, MRR@10 0,603 → 0,577), así que se queda df < 3 (`MinDFDifuso`, `FactorDifuso`).
- **Términos exactos y atajos**: `F8`, `Shift+F1`, `Ctrl + F`, `Control+C` y códigos como `01.01.01` son un único
  término que no se raíza ni se parte. Cada palabra se indexa además en su forma exacta (prefijo `=`), con un
  bono configurable (`PesoExacto`).
- **Reescritura de la consulta** (`consulta.go`): `Consulta{Original, Normalizada, Entidades, Aliases}`. La original
  no pierde nunca un término; los alias solo suman con `PesoAlias`. Los alias salen de datos: `eval/alias.yml`
  (grupos de verbos de acción y términos técnicos con sinónimos), `aliases`, `preguntas` y `entidades` de
  `kb/procedimientos/**/*.yml` (se salta `_reserva/`) y `sinonimos` de `kb/conceptos/*.yml` si existe.
- **Fusión RRF** (`hibrido.go`): `Buscador.Buscar(ctx, q, k) []Resultado{ID, Score, RRF, Vector, BM25, Rerank,
  Reordenado, RangoVector, RangoBM25, Fuentes, Meta}`; `BuscarInforme` añade la consulta reescrita, las degradaciones
  y los tiempos por etapa (para el modo traza). Búsquedas vectorial y léxica **en goroutines paralelas**.
- **Degradación**: si falla el vector, sigue solo con BM25 (`Degradado = ["sin_vector"]`); si falla o se pasa de
  `TimeoutRerank` el reranker, se queda el orden y el puntaje RRF (`"sin_rerank"`). Solo es error que fallen las dos
  listas. Lo degradado no entra en caché.
- **Cachés LRU** de tamaño configurable: resultados por consulta normalizada (+ modo, k y filtro) en el `Buscador`,
  y `EmbebedorConCache` para el vector de la consulta (envuelve un `embed.Embebedor` conservando `Nombre()`, así que
  abre la misma colección). Las copias que devuelve la caché son independientes.
- **Filtros por metadatos** (`BuscarFiltrado`): igualdad exacta con las mismas claves que `indexar` pone en el
  almacén (`source`, `confianza`, `manual`, `title`, `page`, `cita`…), así los pases de `rag` (oficiales, Cortex) se
  pueden expresar igual.
- **Clave compuesta**: los ids de las secciones web truncan el manual a 8 caracteres y se repiten entre manuales
  (200 ids, 1 061 fragmentos). El cargador replica la desambiguación de `indexar.KBJSONL` (id + sufijo de huella),
  así los ids casan con el índice vectorial (`TestCargarFragmentosMismosIDsQueIndexar`,
  `TestVectorAlmacenCasaConElIndiceLexico`), y guarda `Meta["clave"] = id|manual`. Comprobado sobre la KB real:
  0 pares (id, manual) repetidos; el oro se escribe y se resuelve por el par.
- Carga: `CargarFragmentos(rutas...)`, `RutasFragmentos(dirKB)`, `CargarProcedimientos(dir)`. Persistencia opcional en
  gob (`GuardarArchivo` / `CargarIndiceArchivo`, versión del analizador incluida).
- `VectorAlmacen` adapta `almacen.Almacen.Buscar` tal cual (solo lectura): pide k·3 trozos y se queda con el mejor
  trozo de cada fragmento.
- Lector YAML mínimo (`yaml.go`) del subconjunto de `kb/procedimientos/ESQUEMA.md`, para no añadir dependencias.

### `internal/rerank`

- `Reranker{Rerank(ctx, q string, docs []string) ([]float64, error)}`.
- `LlamaServer`: cliente de `llama-server --reranking` (`POST /v1/rerank`, Jina/OpenAI-compatible; si da 404 prueba
  `/rerank`; acepta también la respuesta de TEI). Recorta cada documento a `MaxRunas` (1 200) y no envía vacíos.
  Verificado contra llama.cpp 0.5.0 (build 11146): las cuatro rutas `/v1/rerank`, `/rerank`, `/v1/reranking`,
  `/reranking` responden igual, con `results[{index, relevance_score}]` ordenados por puntaje (logits sin sigmoide).
- `Noop` (orden de entrada) y `ConRespaldo{Principal, Timeout, AlFallar}` (si falla o se pasa, devuelve el orden de
  entrada sin error y avisa por `AlFallar`). El `Buscador` tiene además su propio `TimeoutRerank`, que conserva el
  puntaje RRF y lo anota en el informe.

### `cmd/evalbusqueda`

Evaluación sobre el oro: métricas, latencia, RAM, barrido de parámetros, casos detallados y modo `-pool` para juzgar.

### Pruebas

`go vet ./internal/busqueda/... ./internal/rerank/... ./cmd/evalbusqueda/...` y
`go test -race ./internal/busqueda/... ./internal/rerank/...` en verde. Cubren: tokenización (atajos, códigos,
plegado), raíz ligera, normalización, BM25 con un corpus mínimo cuyos puntajes se calculan a mano, peso de campo,
término exacto y atajos, difuso, id duplicado, persistencia gob, RRF (puntajes exactos, rangos y fuentes), modos
simples, degradación con el vector caído y con el reranker caído / lento / con respuesta corta, reordenado del top,
caché = sin caché (incluidas consultas que normalizan igual y mutación de lo devuelto), búsquedas en paralelo (el
vector de prueba solo contesta cuando el léxico ya terminó), uso concurrente con `-race`, caché del embebedor,
YAML, ids iguales a `indexar`, procedimientos (con `_reserva/` y `entidades`), reescritura con alias, filtros y una
prueba de integración con `almacen` + `indexar` reales. En `rerank`: cliente contra `httptest` (mapeo por índice,
recorte, ruta alternativa y formato TEI, respuestas inválidas, HTTP 501), `ConRespaldo` con timeout, con HTTP 500 y
con el servidor caído.

## 3. Conjunto dorado (`eval/busqueda_oro.jsonl`)

72 consultas, 561 juicios de relevancia (288 de grado 2, 273 de grado 1; media 7,8 por consulta, de 2 a 19).
Formato por línea: `{id, consulta, tipo, sintetico, origen, relevantes: [{id, manual, grado}]}`.

- **Consultas.** 2 reales (`sintetico: false`): las dos del caso reportado, «¿qué es un metrado?» y «¿Cómo registro
  un nuevo presupuesto en S10?». Las otras 70 son **sintéticas** (`sintetico: true`): las escribí imitando cómo
  pregunta un usuario (coloquiales, sin tildes, 13 con faltas: «presupueto», «elavoro», «obrreros», «tranferir»,
  «cotisaciones», «ago el sierre», «avanse», «valoriso un sub contrato», «kiero»). No hay más preguntas reales
  disponibles: `datos/sin_respuesta.jsonl` y las trazas solo tienen las preguntas fuera de dominio del arnés.
  Tipos: 51 «cómo hago X», 13 «qué es X», 4 de término exacto (F8, F9, Shift+F1, Control+C), 3 «dónde», 1 de
  mensaje de error.
- **Origen de cada consulta** (campo `origen`): los pasos de los cinco tutoriales de `tutoriales/` con sus citas
  [F…] (presupuesto, CTS, ingreso a almacén, orden de compra, factura electrónica) y secciones de manuales oficiales
  con `pasos` (Almacenes, Compras, Nóminas, Contabilidad, Administrativo, Gerencia de Proyectos, Facturación,
  instalación, soporte).
- **Relevancia, juzgada a mano por mí**, con este procedimiento:
  1. Semillas: las citas [F…] de los tutoriales, resueltas a ids de fragmento (título + página, vídeo + minuto,
     manual + URL) y las secciones del manual oficial de la tarea.
  2. **Pooling** (como en TREC): el top 10 de cada uno de los 7 métodos (vector, BM25, BM25 + alias, híbrida,
     híbrida + alias y las dos variantes con reranker) se unió y se juzgó candidato por candidato, viendo sección,
     encabezados de la página del PDF y un extracto.
  3. Búsquedas dirigidas en el texto de la KB para los duplicados del mismo contenido (la sección web, su semilla de
     Cortex, las páginas de los dos PDF importados a mano, el vídeo y su transcripción).
- **Grados**: 2 = responde la consulta (la sección o página que explica el procedimiento o la definición);
  1 = útil pero parcial (un subpaso, una sección vecina, un sílabo que lo nombra, una captura con OCR legible).
- **Criterios de exclusión**: no cuentan como relevantes los nodos de Cortex que contienen un **manual entero**
  (`cortex-061e96…` Manual de Presupuestos, `cortex-d9c6…` Gerencia de Proyectos, etc.: no son una unidad citable),
  las FAQ de Optimiza 360, las páginas de marketing y las fichas de pantalla que solo listan campos por OCR.
- **Casos sin respuesta en la KB**: «¿qué es un metrado?» no tiene una definición en ningún fragmento (lo más cercano
  es «Para elaborar el presupuesto es necesario tener los metrados», §1.4.1, y la introducción del manual);
  «toma de inventario físico» (q39) no tiene procedimiento, solo menciones (todo grado 1). Se dejan a propósito.
- **Sesgos conocidos**: (a) las semillas y las búsquedas dirigidas son léxicas, lo que puede favorecer a BM25; el
  pooling incluye el top 10 del vector para compensarlo. (b) 16 de las 72 consultas se parecen mucho (Jaccard ≥ 0,6)
  a alguna `pregunta` de `kb/procedimientos/`, escritas por otro agente con el mismo propósito; con alias de
  procedimientos activados eso podría inflar la expansión (no la infló: ver §4.3). (c) Con 72 consultas, diferencias
  de ±0,015 en una métrica equivalen a una consulta: los parámetros se eligen por mesetas, no por el máximo.
- **Recall acotado**: Recall@k = aciertos en el top k / min(|relevantes|, k), para que una consulta con 12 duplicados
  válidos pueda llegar a 1. MRR@10 es del primer relevante; nDCG@10 usa ganancia 2^grado − 1.

## 4. Resultados

### 4.1 Configuración inicial (k1 = 1,2; b = 0,75; exacto 0,5; RRF k = 60 con pesos iguales; alias 0,3; reranker top 30)

| Método | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | p50 ms | p95 ms |
|---|---|---|---|---|---|---|
| vector (hoy) | 0,201 | 0,222 | 0,332 | 0,212 | 6,1 | 7,2 |
| BM25 | 0,638 | 0,637 | 0,739 | 0,625 | 0,3 | 1,0 |
| BM25 + alias | 0,602 | 0,605 | 0,738 | 0,595 | 0,4 | 1,1 |
| híbrida RRF | 0,447 | 0,470 | 0,607 | 0,446 | 7,0 | 8,5 |
| híbrida RRF + alias | 0,438 | 0,472 | 0,641 | 0,450 | 6,9 | 8,5 |
| híbrida + rerank (top 30) | 0,733 | 0,656 | 0,926 | 0,702 | 2 165 | 3 034 |
| híbrida + alias + rerank (top 30) | 0,713 | 0,635 | 0,926 | 0,685 | 2 908 | 7 928 ¹ |
| BM25 + rerank (top 30) | 0,700 | 0,640 | 0,904 | 0,685 | 2 634 | 8 074 ¹ |

¹ p95 contaminado: en esos minutos compilé y corrí pruebas en la misma máquina. Las latencias limpias del reranker
están en §4.2 y §4.4.

Por tipo de consulta (Recall@10 / MRR@10):

| Método | cómo (51) | dónde (3) | error (1) | exacto (4) | qué es (13) |
|---|---|---|---|---|---|
| vector (hoy) | 0,245 / 0,369 | 0,125 / 0,167 | 0,000 / 0,000 | 0,097 / 0,167 | 0,209 / 0,304 |
| BM25 | 0,628 / 0,751 | 0,431 / 0,261 | 0,900 / 1,000 | 0,797 / 1,000 | 0,654 / 0,699 |
| BM25 + alias | 0,625 / 0,757 | 0,431 / 0,528 | 0,900 / 1,000 | 0,742 / 1,000 | 0,504 / 0,609 |
| híbrida RRF | 0,483 / 0,617 | 0,347 / 0,548 | 0,400 / 1,000 | 0,456 / 0,625 | 0,460 / 0,543 |
| híbrida RRF + alias | 0,494 / 0,664 | 0,347 / 0,583 | 0,400 / 1,000 | 0,428 / 0,625 | 0,438 / 0,539 |
| híbrida + rerank (top 30) | 0,661 / 0,925 | 0,431 / 1,000 | 0,700 / 1,000 | 0,703 / 1,000 | 0,671 / 0,885 |
| híbrida + alias + rerank | 0,647 / 0,925 | 0,431 / 1,000 | 0,700 / 1,000 | 0,703 / 1,000 | 0,606 / 0,885 |
| BM25 + rerank (top 30) | 0,642 / 0,914 | 0,389 / 0,667 | 0,700 / 1,000 | 0,753 / 1,000 | 0,652 / 0,885 |

### 4.2 Sin bono de término exacto y con el vector a peso 0,2 en la RRF

| Método | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | p50 ms | p95 ms |
|---|---|---|---|---|---|---|
| vector (hoy) | 0,201 | 0,222 | 0,332 | 0,212 | 9,4 | 11,4 |
| BM25 | 0,637 | 0,645 | 0,765 | 0,630 | 0,5 | 1,7 |
| BM25 + alias | 0,587 | 0,592 | 0,703 | 0,573 | 0,6 | 1,9 |
| híbrida RRF (vector 0,2) | 0,564 | 0,609 | 0,764 | 0,581 | 9,7 | 12,7 |
| híbrida RRF + alias | 0,542 | 0,575 | 0,736 | 0,550 | 9,6 | 14,3 |
| híbrida + rerank (top 30) | 0,687 | 0,641 | 0,890 | 0,684 | 2 035 | 2 356 |
| híbrida + alias + rerank (top 30) | 0,685 | 0,642 | 0,891 | 0,678 | 2 348 | 2 650 |
| BM25 + rerank (top 30) | 0,682 | 0,641 | 0,888 | 0,683 | 2 097 | 2 381 |

Separando los dos cambios (corrida C: sin bono exacto y **pesos iguales**):

| Método | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | p50 ms | p95 ms |
|---|---|---|---|---|---|---|
| híbrida RRF (pesos iguales) | 0,432 | 0,470 | 0,634 | 0,451 | 9,9 | 15,9 |
| híbrida + rerank (top 30) | 0,724 | 0,644 | 0,929 | 0,693 | 2 180 | 2 553 |

Lectura: el bono exacto no cambia nada con reranker (0,926 / 0,929 de MRR) y resta sin él; el peso del vector sí
importa, y en sentidos opuestos: 0,1–0,2 es lo mejor como orden final sin reranker, 1 es lo mejor como pool del
reranker. Por eso el `Buscador` tiene dos pesos (`PesoVector` para el pool, `PesoVectorSinRerank` para el orden
final cuando no hay reranker o este falla).

### 4.2 bis Configuración recomendada, medida de punta a punta (corrida D, valores por defecto del código)

k1 = 1,2; b = 0,75; exacto 0; RRF k = 60; pool del reranker con pesos iguales; sin reranker, vector a 0,1;
reranker sobre los 30 primeros; sin alias.

| Método | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | p50 ms | p95 ms |
|---|---|---|---|---|---|---|
| (1) vector (hoy) | 0,201 | 0,222 | 0,332 | 0,212 | 10,0 | 12,7 |
| (2) BM25 | 0,637 | 0,645 | 0,765 | 0,630 | 0,5 | 1,8 |
| (3) híbrida RRF (orden final con vector 0,1) | 0,616 | 0,631 | 0,780 | 0,611 | 12,4 | 15,1 |
| (4) híbrida + rerank (pool con pesos iguales, top 30) | 0,724 | 0,644 | 0,929 | 0,693 | 2 041 | 2 291 |

Por tipo (Recall@10 / MRR@10):

| Método | cómo (51) | dónde (3) | error (1) | exacto (4) | qué es (13) |
|---|---|---|---|---|---|
| vector (hoy) | 0,245 / 0,369 | 0,125 / 0,167 | 0,000 / 0,000 | 0,097 / 0,167 | 0,209 / 0,304 |
| BM25 | 0,621 / 0,752 | 0,431 / 0,250 | 0,900 / 1,000 | 0,825 / 1,000 | 0,714 / 0,846 |
| híbrida RRF | 0,609 / 0,773 | 0,431 / 0,528 | 0,800 / 1,000 | 0,797 / 1,000 | 0,701 / 0,782 |
| híbrida + rerank | 0,640 / 0,928 | 0,431 / 1,000 | 0,700 / 1,000 | 0,731 / 1,000 | 0,675 / 0,887 |

El reranker mejora sobre todo **qué sale primero** (MRR 0,78 → 0,93): el Recall@10 apenas cambia porque solo
reordena los 30 candidatos que ya trajo la fusión.

### 4.3 Expansión con alias

Peso de los alias 0,3 sobre BM25 (exacto 0) y sobre la híbrida con vector 0,1; «con procedimientos» = además los
`aliases`, `preguntas` y `entidades` de los 60 procedimientos de `kb/procedimientos/`.

| | Recall@5 | Recall@10 | MRR@10 | nDCG@10 |
|---|---|---|---|---|
| BM25 sin alias | 0,637 | 0,645 | 0,765 | 0,630 |
| BM25 + `alias.yml` | 0,593 | 0,619 | 0,735 | 0,594 |
| BM25 + `alias.yml` + procedimientos | 0,587 | 0,592 | 0,703 | 0,573 |
| híbrida (vector 0,1) sin alias | — | 0,631 | 0,780 | 0,611 |
| híbrida (vector 0,1) + `alias.yml` | 0,595 | 0,608 | 0,766 | 0,591 |
| híbrida (vector 0,1) + `alias.yml` + procedimientos | — | 0,586 | 0,744 | 0,571 |
| híbrida + rerank sin alias (corrida A) | 0,733 | 0,656 | 0,926 | 0,702 |
| híbrida + rerank + alias (corrida A) | 0,713 | 0,635 | 0,926 | 0,685 |

**La expansión empeora en todas las variantes medidas**, más con los alias de procedimientos, y sobre todo en las
preguntas «qué es» (Recall@10 0,714 → 0,509 en BM25): los verbos de acción («registrar», «crear», «cargar») y los
títulos de procedimientos parecidos arrastran secciones de otra tarea (p. ej. «¿Cómo registro un nuevo presupuesto?»
con alias pone primero «Registro manual de un Socio de Negocio»). La única mejora fue de MRR con la fusión de pesos
iguales (0,607 → 0,641), que no es la configuración recomendada. **Recomendación: `Expandir = false`.** La reescritura
sigue siendo útil para la traza (consulta normalizada, entidades) y la caché, y los alias de procedimientos tienen
más sentido para **elegir el procedimiento** (el `id` del procedimiento como unidad de recuperación) que para
expandir BM25 sobre fragmentos.

### 4.4 Barrido de parámetros

BM25 solo, sin bono exacto (Recall@10 / MRR@10 / nDCG@10):

| k1 \ b | 0,30 | 0,50 | 0,75 | 0,90 |
|---|---|---|---|---|
| 0,6 | 0,605 / 0,731 / 0,589 | 0,617 / 0,726 / 0,598 | 0,629 / 0,740 / 0,608 | 0,619 / 0,740 / 0,602 |
| 0,9 | 0,607 / 0,730 / 0,594 | 0,621 / 0,736 / 0,608 | 0,640 / 0,757 / 0,623 | 0,624 / 0,752 / 0,610 |
| 1,2 | 0,604 / 0,741 / 0,598 | 0,632 / 0,741 / 0,613 | **0,645 / 0,765 / 0,630** | 0,643 / 0,767 / 0,625 |
| 1,5 | 0,613 / 0,761 / 0,609 | 0,629 / 0,768 / 0,620 | 0,637 / 0,772 / 0,629 | 0,633 / 0,780 / 0,625 |
| 2,0 | 0,590 / 0,768 / 0,603 | 0,612 / 0,773 / 0,619 | 0,623 / 0,787 / 0,632 | 0,621 / 0,779 / 0,624 |

Meseta en b = 0,75 con k1 entre 0,9 y 2,0 (diferencias ≤ 0,02, una o dos consultas). b bajo (0,3) pierde: las
páginas largas de PDF y los nodos de Cortex largos necesitan normalización por longitud.

Bono exacto y pesos de campo (BM25, k1 1,2, b 0,75):

| Variante | Recall@10 | MRR@10 | nDCG@10 |
|---|---|---|---|
| exacto 0,5 (título/sección/preguntas 3, texto 1, ocr 0,5, manual 0,5) | 0,637 | 0,739 | 0,625 |
| **sin bono exacto** | **0,645** | **0,765** | **0,630** |
| exacto 1,0 | 0,623 | 0,751 | 0,619 |
| todos los campos con peso 1 (BM25 plano) | 0,609 | 0,746 | 0,606 |
| título/sección 2 | 0,628 | 0,736 | 0,618 |
| título/sección 5 | 0,640 | 0,752 | 0,625 |
| OCR 0 (sin capturas) | 0,627 | 0,739 | 0,621 |
| OCR 1 | 0,625 | 0,738 | 0,616 |

Los pesos de campo sí aportan (+0,036 de Recall@10 frente a BM25 plano); el OCR a 0,5 es mejor que quitarlo o
igualarlo al texto.

Peso del vector en la RRF, sin reranker (k = 60; exacto 0):

| peso vector | sin alias | con alias |
|---|---|---|
| 1,0 | 0,470 / 0,634 / 0,451 | 0,468 / 0,644 / 0,449 |
| 0,7 | 0,512 / 0,672 / 0,490 | 0,510 / 0,666 / 0,482 |
| 0,5 | 0,527 / 0,696 / 0,507 | 0,515 / 0,678 / 0,492 |
| 0,3 | 0,580 / 0,733 / 0,553 | 0,556 / 0,712 / 0,529 |
| 0,2 | 0,609 / 0,764 / 0,581 | 0,575 / 0,736 / 0,550 |
| **0,1** | **0,631 / 0,780 / 0,611** | 0,586 / 0,744 / 0,571 |

k de RRF × peso de alias, sin reranker (Recall@10 / MRR@10 / nDCG@10), con vector 0,2 (corrida B) y con pesos
iguales (corrida C):

| k RRF | vector 0,2, sin alias | vector 0,2, alias 0,3 | pesos iguales, sin alias | pesos iguales, alias 0,3 |
|---|---|---|---|---|
| 10 | 0,641 / 0,775 / 0,622 | 0,593 / 0,728 / 0,569 | 0,516 / 0,675 / 0,497 | 0,490 / 0,677 / 0,479 |
| 30 | 0,627 / 0,784 / 0,611 | 0,586 / 0,749 / 0,573 | 0,475 / 0,638 / 0,457 | 0,471 / 0,650 / 0,455 |
| 60 | 0,609 / 0,764 / 0,581 | 0,575 / 0,736 / 0,550 | 0,470 / 0,634 / 0,451 | 0,468 / 0,644 / 0,449 |
| 100 | 0,576 / 0,751 / 0,556 | 0,554 / 0,722 / 0,532 | 0,471 / 0,634 / 0,451 | 0,472 / 0,643 / 0,449 |

Sin reranker, k pequeño ayuda (k = 10 con vector 0,2 iguala a BM25 solo), porque acentúa el primer puesto de cada
lista; el k de RRF con reranker solo se midió en 60.

N de candidatos del reranker (pool con pesos iguales, exacto 0, corrida C):

| N | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | p50 ms | p95 ms |
|---|---|---|---|---|---|---|
| 10 | 0,558 | 0,470 | 0,847 | 0,556 | 713 | 955 |
| 20 | 0,694 | 0,607 | 0,913 | 0,666 | 1 429 | 1 784 |
| **30** | **0,724** | **0,644** | **0,929** | **0,693** | 2 105 | 2 452 |
| 50 | 0,693 | 0,644 | 0,916 | 0,690 | 3 769 | 4 337 |

Con el pool de vector 0,2 (corrida B) el reranker rinde menos en todos los N (N = 30: MRR 0,890, nDCG 0,684;
N = 50: 0,903 / 0,691).

## 5. Los dos casos, uno por uno (top 5)

Configuración recomendada (corrida D). ✔ = relevante en el oro (grado).

**«¿qué es un metrado?»** (caso real; la KB no tiene una definición: los relevantes son §1.4.1 «Metrados del proyecto
para elaborar el presupuesto», la introducción del manual y §3.5.2.5 «Ingreso de metrados», con sus semillas de
Cortex)

| # | (1) vector, hoy | (2) BM25 | (3) híbrida RRF | (4) híbrida + rerank |
|---|---|---|---|---|
| 1 | FAQ: Arturo Dupont, CEO y fundador de Optimiza 360 | MP §3.5.2.5 Ingreso de metrados ✔1 | Cortex §3.5.2.5 Ingreso de metrados ✔1 | Guía de Usuario p. 95 (vincular metrado a celda Excel) |
| 2 | FAQ: Optimiza 360 | Cortex §3.5.2.5 ✔1 | MP §3.5.2.5 ✔1 | GP §4.6.10 Valorización del subcontrato |
| 3 | FAQ: S10 Conocimiento | MP §1.4.1 Metrados del proyecto ✔2 | MP §1.4.1 ✔2 | Cortex §4.6.10 |
| 4 | FAQ: Arturo Dupont (otra variante) | Manual Costos y Presupuestos p. 101 | Manual Costos y Presupuestos p. 101 | Manual Costos y Presupuestos p. 86 (metrados faltantes) |
| 5 | FAQ: S10 Conocimiento | Cortex §1.4.1 ✔2 | Cortex §1.4.1 ✔2 | Cortex §1.4.1 ✔2 |

El vector reproduce el fallo reportado (la biografía del CEO primero; distancias de 0,57–0,60 en el informe
original). BM25 y la híbrida ponen 4 relevantes en el top 5. **El reranker empeora este caso** (1 relevante): sin una
definición en la base, el cross-encoder prefiere pasajes que usan «metrado» en un procedimiento concreto. Con el bono
exacto a 0,5 (corrida A) BM25 tampoco acertaba (páginas con muchas apariciones literales de «metrado»). Lo que de
verdad resuelve esta pregunta es **contenido**: un concepto «metrado» en `kb/conceptos/` o en Cortex.

**«¿Cómo registro un nuevo presupuesto en S10?»** (caso real)

| # | (1) vector, hoy | (2) BM25 | (3) híbrida RRF | (4) híbrida + rerank |
|---|---|---|---|---|
| 1 | Guía p. 2 (portada e índice) | Cortex §4.5.1 Registro de nuevos títulos en el catálogo | Guía p. 2 | Manual Costos y Presupuestos p. 13 «1.3.2 Iniciando el registro del nuevo presupuesto» ✔2 |
| 2 | Guía p. 127 (accesos de usuarios) | Guía p. 2 | Manual C. y P. p. 80 | Guía p. 11 «1.1.2 Iniciando el registro del nuevo presupuesto» ✔2 |
| 3 | Guía p. 67 (opciones de Datos generales) | Manual C. y P. p. 13 ✔2 | Cortex §4.5.1 | Guía p. 2 |
| 4 | MP §5.1.1 Datos generales (1) | MP §5.3.3 Importar presupuesto | Manual C. y P. p. 13 ✔2 | Guía p. 10 «1.1 Registro del nuevo presupuesto» ✔1 |
| 5 | Guía p. 18 | Manual C. y P. p. 80 | MP §5.3.3 | Manual C. y P. p. 79 «2.4.3.1 Nuevo» ✔1 |

Vector: 0 relevantes en el top 5. BM25 e híbrida: 1. **Con reranker: 4, y las dos páginas de grado 2 en los
puestos 1 y 2.** (La regla fija de `rag.go` para esta pregunta —páginas 11, 12 y 17 de la Guía— deja de ser
necesaria con el reranker.)

## 6. RAM y latencia

| Componente | Medida | Cómo |
|---|---|---|
| Índice BM25 (8 135 docs, 87 831 términos con raíces y formas exactas) | **32,4 MiB** de montículo Go vivo | `runtime.MemStats.HeapAlloc` tras GC, antes y después de indexar; incluye ~13 MiB de texto guardado para el reranker (1 800 bytes por documento) |
| Construcción del índice | 0,66–2,5 s | variable por la carga de la máquina |
| Índice BM25 en gob | 17,9–19,4 MiB en disco, carga en 57–141 ms | `GuardarArchivo` / `CargarIndiceArchivo` |
| Búsqueda BM25 | p50 0,3–0,6 ms, p95 1–2,5 ms | 72 consultas × 3 repeticiones |
| Búsqueda vectorial (chromem, 15 277 trozos) | p50 6–10 ms, p95 7–21 ms | ídem |
| Híbrida (las dos en paralelo + RRF) | p50 7–12 ms, p95 8,5–16 ms | ídem; domina el vector |
| Reranker bge-reranker-v2-m3 Q4_K_M (438 MB en disco), Metal | **huella física 880 MB** (`footprint`); RSS de `ps` 189–203 MiB, que **no** cuenta los búferes de Metal | llama-server `-c 8192 -b 8192 -ub 8192 --parallel 1` |
| Ídem con `-c 2048 -b 2048 -ub 2048` | huella 878 MB (no baja); p50 3,0 s, p95 5,1 s (más lento) | no se truncó ningún documento (270–400 tokens cada uno) |
| Ídem solo CPU (`-ngl 0`) | huella 668 MB, RSS 426 MiB | latencia **no medida de forma fiable**: la máquina estaba a carga media 65 por otros agentes (≈ 17 s por documento en esos minutos) |
| Híbrida + rerank top 30, Metal | p50 2,0–2,2 s, p95 2,3–2,6 s (corridas limpias) | N = 20: p50 1,4 s; N = 10: 0,7 s |

Caché: un acierto de la LRU de resultados devuelve una copia sin volver a buscar (probado igual al resultado sin
caché en `TestCacheDaLoMismoQueSinCache`); la del embebedor evita volver a embeber la misma consulta normalizada.

**Ojo con producción.** Los contenedores de Metrín corren en Docker (Linux, sin Metal): el reranker iría solo CPU,
con ~670 MB de RAM y una latencia que hay que medir en esa máquina antes de activarlo; si no cabe en el presupuesto
de latencia, la configuración «híbrida sin reranker» (fila 3) ya mejora mucho a la de hoy (Recall@10 0,222 → 0,631,
MRR 0,332 → 0,780) por ~12 ms y 32 MiB.

## 7. Configuración recomendada

Valores ya puestos como defecto en `busqueda.ConfigPorDefecto()` y `busqueda.OpcionesPorDefecto()`:

| Parámetro | Valor | Por qué (cifras arriba) |
|---|---|---|
| BM25 k1 | 1,2 | meseta 0,9–2,0 con b 0,75; 1,2 es el máximo de Recall y nDCG y el valor estándar |
| BM25 b | 0,75 | b 0,3 pierde 0,04 de Recall@10 |
| Pesos de campo | título / sección / preguntas 3, entidades 2, texto 1, OCR 0,5, manual 0,5 | +0,036 de Recall@10 sobre BM25 plano; OCR 0 y 1 peores |
| Bono de término exacto | 0 | 0,5 resta 0,026 de MRR; con reranker da igual; los atajos no lo necesitan |
| Difuso (faltas) | activo: df < 3, distancia 1 (2 con 8+ letras), vecino ≥ 10× más frecuente | 7 de 10 faltas comprobadas se corrigen; un umbral más agresivo no mejoró el subconjunto con faltas |
| Fusión RRF | k = 60, pool con pesos iguales | mejor pool para el reranker (MRR 0,929 frente a 0,890 con vector 0,2) |
| Orden final sin reranker | vector 0,1 (`PesoVectorSinRerank`) | Recall@10 0,631, MRR 0,780: el mejor MRR sin reranker |
| Reranker | bge-reranker-v2-m3 Q4_K_M, top **N = 30**, 1 200 runas por documento | N = 30 es el máximo; N = 20 si hace falta bajar a ~1,4 s (−0,016 MRR, −0,027 nDCG) |
| `TimeoutRerank` | 3 s con Metal | p95 medido 2,3–2,6 s; en CPU hay que medirlo |
| Alias | desactivados (`Expandir = false`) | empeoran Recall y MRR en todas las variantes medidas |
| Cachés | 256 resultados; embebedor a criterio | igualdad probada |

## 8. Integración en `rag` (no hecha: `internal/rag/` lo editan otros agentes)

Lo que habría que cambiar, todo en `internal/rag/rag.go` (y la construcción en `cmd/rag/main.go`):

1. **Construcción** (`cmd/rag/main.go`): cargar los fragmentos con `busqueda.CargarFragmentos(busqueda.RutasFragmentos(dirKB)...)`
   (o el gob guardado), `busqueda.IndexarDocs(busqueda.ConfigPorDefecto(), docs)`, envolver el embebedor de
   consulta en `busqueda.NuevoEmbebedorConCache`, y crear
   `busqueda.NuevoBuscador(lex, busqueda.VectorAlmacen{A: almacen}, rerank.NuevoLlamaServer(url), nil, busqueda.OpcionesPorDefecto())`
   (con `Reranker = nil` si no hay `RERANK_URL`). Añadir a `config.go` `RERANK_URL` y, si se quiere, `BUSQUEDA_MODO`.
2. **`RAG.preguntar`, etapa de búsqueda**: las llamadas a `r.Almacen.Buscar(ctx, busqueda, o.K, o.Filtro)`, el pase de
   oficiales (`filtroOficial`), el de la consulta compacta y el de Cortex (`manual = Cortex`) pasan a ser
   `r.Busqueda.BuscarFiltrado(ctx, busqueda, k, filtro)` con los mismos filtros: `Meta` lleva las mismas claves que
   el almacén (`source`, `confianza`, `manual`, `title`, `page`, `cita`). Con el reranker, el pase de oficiales y las
   reglas `esRegistroNuevoPresupuesto` / páginas 11-12-17 probablemente sobran (§5); hay que comprobarlo con la traza.
3. **`seleccionarContextoCorte`** es la función que más cambia: hoy corta por `Distancia` (1 − coseno) con
   `ventanaRelevancia = 0.18` y `MaxDistancia`; con la búsqueda nueva el orden lo da `Resultado.Score` (logit del
   reranker, o RRF si no hay). Propuesta: tomar los `maxContextos` primeros por `Score`, mantener la deduplicación por
   cita y la penalización por `confianza` como desempate, y sustituir el umbral de distancia por un umbral de
   puntaje del reranker (`MarcaSinContexto` si el primero está por debajo), calibrado con el modo traza. Para no
   romper `DistanciaMin` y el registro de fallos, la conversión puede rellenar `almacen.Resultado.Distancia` con
   `1 − Vector` (la similitud coseno sigue en el resultado).
4. **Traza**: `Buscador.BuscarInforme` devuelve consulta normalizada, entidades, alias, degradaciones
   (`sin_vector`, `sin_rerank`) y tiempos por etapa (`vector`, `lexico`, `fusion`, `rerank`), listos para
   `trazarBusqueda`.
5. **Procedimientos**: `busqueda.CargarProcedimientos("kb/procedimientos")` produce `Doc` con id `proc:<id>`; se pueden
   indexar en el mismo índice o en uno aparte para elegir primero el procedimiento (sus `preguntas` pesan 3) y luego
   los fragmentos que cita.

## 9. Cómo reproducir

```bash
# 1. Modelo (una vez; Apache-2.0, ficha https://huggingface.co/gpustack/bge-reranker-v2-m3-GGUF, base BAAI/bge-reranker-v2-m3)
mkdir -p ~/.cache/metrin-modelos
curl -L -o ~/.cache/metrin-modelos/bge-reranker-v2-m3-Q4_K_M.gguf \
  https://huggingface.co/gpustack/bge-reranker-v2-m3-GGUF/resolve/main/bge-reranker-v2-m3-Q4_K_M.gguf

# 2. Reranker SOLO para la evaluación (llama.cpp 0.5.0, build 11146)
/opt/homebrew/bin/llama-server -m ~/.cache/metrin-modelos/bge-reranker-v2-m3-Q4_K_M.gguf \
  --reranking --host 127.0.0.1 --port 8091 -c 8192 -b 8192 -ub 8192 --parallel 1 --no-webui &
echo $! > /tmp/llama-rerank.pid

# 3. Evaluación (desde metrin/)
go run ./cmd/evalbusqueda -kb ../kb -oro eval/busqueda_oro.jsonl -alias eval/alias.yml \
  -rerank-url http://127.0.0.1:8091 -rerank-pid $(cat /tmp/llama-rerank.pid) -repeticiones 3 -barrido
#   corrida A (inicial):     -peso-exacto 0.5 -peso-vector 1
#   corrida B:               -peso-exacto 0   -peso-vector 0.2
#   corrida C:               -peso-exacto 0   -peso-vector 1 -metodos "vector (hoy)|BM25|híbrida RRF|híbrida + rerank (top 30)"
#   corrida D (recomendada): valores por defecto + -peso-vector-sin-rerank 0.1
#   corrida E (alias):       -peso-vector 0.1 -alias-procedimientos=false -metodos "BM25 + alias|híbrida RRF + alias"
#   pool para juzgar:        -pool /tmp/pool.txt -pool-k 10
#   umbral del difuso:       -difuso-min-df 20 -difuso-factor 20 (experimento de §2; defecto 3 × 10)

# 4. Apagar el reranker
kill $(cat /tmp/llama-rerank.pid)

# Pruebas
go vet ./internal/busqueda/... ./internal/rerank/... ./cmd/evalbusqueda/...
go test -race ./internal/busqueda/... ./internal/rerank/...
```

Licencia del modelo verificada en la ficha de Hugging Face: `apache-2.0` tanto en `gpustack/bge-reranker-v2-m3-GGUF`
como en el original `BAAI/bge-reranker-v2-m3`. El modelo vive fuera del repositorio. No se usó ninguna API externa.

## 10. Limitaciones

- 70 de las 72 consultas son sintéticas; las juzgué yo solo, sin segundo juez. Conviene sumar preguntas reales
  según lleguen (el modo traza las registra) y repetir.
- El oro mide fragmentos; los procedimientos no se indexaron en la evaluación (`-indexar-procedimientos` existe).
- La línea base «vector (hoy)» es `almacen.Buscar` con la pregunta tal cual; `rag.go` además hace pases filtrados
  (oficiales, consulta compacta, Cortex) y reglas fijas, que no se reproducen aquí.
- La latencia del reranker solo CPU (la de Docker) queda por medir en una máquina sin otra carga.

## 11. Reranker encendido (06-10-2026, segunda medición)

El usuario quiere el reranker **prendido** para ver si mejora, y más rápido que los ~2 s de §4. Todo lo de esta
sección se midió el 06-10-2026 entre las 04:20 y las 05:40 en la misma Mac (M4 Pro), **compartida**: con el reranker
parado, `ioreg` daba la GPU al 100 % de uso por otros procesos (otros `llama-server`, Ollama, MLX) y la carga media
osciló entre 4 y 90. Por eso las latencias valen **por ventana**: las comparaciones de servidor se hicieron
intercaladas y se repitió la de referencia. Ninguna cifra es estimada.

### 11.1 Qué se cambió

- `cmd/evalbusqueda` (aditivo): `-rerank-variantes N:runas,…` (un método por variante), `-rerank-cache` (el
  cross-encoder puntúa cada par consulta–documento por separado, así que las variantes de N reutilizan los puntajes:
  la calidad es exacta, la latencia de lo que sale de la caché no vale), `-cascada umbrales` y `-cascada-tipos`.
- V2 (aditivo, con prueba `TestConfigRerankTopNYRunas`): `RERANK_TOP_N` (candidatos que pasan por el reranker) y
  `RERANK_MAX_RUNAS` (runas de cada uno) en `internal/v2/config.go` y `cmd/rag/v2.go`; por defecto **20 y 800**. Se
  suman a `RERANK_URL` (vacía = apagado, sigue siendo el defecto) y `RERANK_TIMEOUT_MS` (3000). `/health` los muestra
  en `v2.busqueda`. `internal/busqueda` e `internal/rerank` no se tocaron: `Opciones.TopRerank` y
  `LlamaServer.MaxRunas` ya existían.
- La latencia de servidor se midió reenviando al reranker las **cargas reales** de la evaluación (las 72 consultas
  con sus 30 candidatos en el orden del pool, grabadas con un proxy), una consulta tras otra: es solo el tiempo HTTP
  del reranker; la búsqueda híbrida suma 12–25 ms.

### 11.2 Candidatos (N) × recorte (runas): calidad y latencia

bge-reranker-v2-m3 Q4_K_M, Metal, `--parallel 1`. Calidad: oro de 72 consultas (exacta, no depende de la carga).
Latencia: 40 consultas reales, ventana de 05:04 a 05:10 (carga 4–15). Tokens por documento: medidos con `usage` de
llama-server (incluyen la consulta y los separadores).

| N | runas (tokens/doc) | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | p50 ms | p95 ms |
|---|---|---|---|---|---|---|---|
| sin reranker (híbrida RRF, vector 0,1) | — | 0,616 | 0,631 | 0,780 | 0,611 | 12–26 | 16–90 |
| 30 | 1 200 (295) | **0,724** | **0,644** | **0,929** | **0,693** | 1 593 | 3 148 |
| 30 | 800 (214) | 0,692 | 0,629 | 0,918 | 0,673 | 1 981 ¹ | 2 330 ¹ |
| 30 | 500 (144) | 0,635 | 0,599 | 0,882 | 0,637 | 1 147 | 1 436 |
| 20 | 1 200 (289) | 0,694 | 0,607 | 0,913 | 0,666 | 1 077 | 1 233 |
| **20** | **800 (211)** | 0,676 | 0,599 | **0,910** | 0,654 | **828** | **967** |
| 20 | 500 | 0,618 | 0,566 | 0,880 | 0,616 | — | — |
| 15 | 1 200 | 0,620 | 0,547 | 0,894 | 0,618 | — | — |
| 15 | 800 (209) | 0,609 | 0,542 | 0,890 | 0,613 | 700 | 824 |
| 15 | 500 | 0,570 | 0,519 | 0,861 | 0,587 | — | — |
| 10 | 1 200 (282) | 0,558 | 0,470 | 0,847 | 0,556 | 727 | 1 111 |
| 10 | 800 | 0,547 | 0,470 | 0,854 | 0,553 | — | — |
| 10 | 500 | 0,519 | 0,470 | 0,846 | 0,541 | — | — |

¹ medida cuando la carga subió de 4 a 15 (otro proceso empezó a usar la GPU): no comparable con su vecina 30 × 1 200.

- El coste es casi lineal en N × tokens (≈ 0,16 ms por token más ~10 ms por documento): **solo N y el recorte mueven
  la latencia**.
- Recortar a 800 runas cuesta poco (MRR −0,011 con N = 30, −0,003 con N = 20); a 500 ya cuesta 0,03–0,05.
- Con N < 30, Recall@10 queda **por debajo de no usar reranker** (0,599 frente a 0,631 con N = 20): el reranker
  solo reordena los N primeros del pool de pesos iguales, que es peor que el orden final sin reranker en la cola.
  Lo que mejora es lo de arriba: MRR, Recall@5 y nDCG.

### 11.3 Parámetros del servidor

Misma carga de trabajo (N = 20, 800 runas salvo donde se indica), mismas 40 consultas.

| Variante | p50 ms | p95 ms | Lectura |
|---|---|---|---|
| `--parallel 1 -ub 8192` | 828; repetida 1 010 | 967; repetida 1 179 | referencia (ventana 05:04–05:15) |
| `--parallel 1 -ub 1024` | 1 076 | 1 266 | igual dentro del ruido (±20 % entre ventanas) |
| `--parallel 4 -c 4096 -ub 4096` | 905 | 1 187 | igual |
| `--parallel 8 -c 8192 -ub 8192` | 1 232 | 2 147 | **peor**: junta varias secuencias en un lote y la atención crece con el lote |
| `--parallel 8`, N = 30 × 1 200 | 2 214 | 3 216 | peor que `--parallel 1` (1 593 / 3 148) |
| Q4_0 en vez de Q4_K_M, `-ub 1024` | 971 | 1 293 | sin ganancia |
| `-fa off`, N = 30 × 1 200 (ventana cargada) | 7 531 | 14 903 | peor que `-fa auto` en la misma ventana (5 887 / 10 319) |
| `-ub 512` | — | — | HTTP 500: hay documentos de más de 512 tokens (el lote debe contener la secuencia entera) |
| Solo CPU (`-ngl 0 -t 8`, `-ub 1024`) | ≈ 2,9 s **por documento** | — | **no medible de forma justa**: otro `llama-server` con `-ngl 0` ocupaba la CPU (carga 40–90). No es viable en esta Mac |
| `--parallel 1 -c 2048 -ub 2048` (**elegido**), intercalado con Qwen3 | 1 428–1 562 | 1 886–2 029 | ventana de 05:36–05:39, carga 27–33 |

RAM (`footprint`): con `-c 2048 -b 2048 -ub 2048`, **540 MB** de huella física (pico 617 MB), frente a los 880 MB
de `-ub 8192` de §6. `ps` da 63 MB de RSS porque no cuenta los búferes de Metal.

### 11.4 Cascada

Se reordena solo si el orden sin reranker duda: margen relativo (s1 − s2) / s1 entre el 1.º y el 2.º de la RRF
(vector 0,1) menor que un umbral, o si el tipo de la consulta es «cómo» o «qué es» (PROCEDURE / CONCEPT). N = 30,
1 200 runas.

| Variante | Consultas reordenadas | Recall@5 | Recall@10 | MRR@10 | nDCG@10 |
|---|---|---|---|---|---|
| siempre | 72 de 72 | 0,724 | 0,644 | 0,929 | 0,693 |
| margen < 0,02 | 38 (53 %) | 0,665 | 0,641 | 0,875 | 0,654 |
| margen < 0,05 | 66 (92 %) | 0,724 | 0,646 | 0,926 | 0,693 |
| margen < 0,10 | 72 (100 %) | 0,724 | 0,644 | 0,929 | 0,693 |
| tipo PROCEDURE / CONCEPT | 64 (89 %) | 0,716 | 0,649 | 0,909 | 0,685 |
| margen < 0,02 o tipo | 68 (94 %) | 0,716 | 0,644 | 0,909 | 0,681 |

El margen de la RRF casi nunca es grande (p10 / p25 / p50 / p75 = 0,007–0,009 / 0,014 / 0,018 / 0,035): la fusión
rara vez «está segura». Ahorrar la mitad de las llamadas cuesta 0,054 de MRR, y mientras se reordene más del 5 % de
las consultas el p95 sigue siendo el del reranker. **La cascada no se adopta.**

### 11.5 Otros rerankers

| Modelo (GGUF) | Licencia (ficha de HF) | Disco | N × runas | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | Latencia |
|---|---|---|---|---|---|---|---|---|
| bge-reranker-v2-m3 Q4_K_M (`gpustack/bge-reranker-v2-m3-GGUF`) | Apache-2.0 | 438 MB | 20 × 800 | 0,676 | 0,599 | **0,910** | 0,654 | p50 1 428–1 562 / p95 1 886–2 029 (intercalado) |
| Qwen3-Reranker-0.6B Q8_0 (`ggml-org/Qwen3-Reranker-0.6B-Q8_0-GGUF`) | Apache-2.0 | 639 MB | 20 × 800 | 0,678 | 0,601 | 0,887 | 0,658 | p50 3 086–3 536 / p95 4 464–5 061 (intercalado): **×2,2**. Lleva una plantilla de instrucción: 323 tokens por documento frente a 210 |
| bge-reranker-base Q8_0 (`xinming0111/bge-reranker-base-Q8_0-GGUF`, base `BAAI/bge-reranker-base`) | MIT | 304 MB | 30 × 800 | 0,612 | 0,579 | 0,855 | 0,605 | ventana cargada, no comparable. Contexto de 512 tokens: con 1 200 runas da HTTP 400 |
| ídem | | | 20 × 800 | 0,605 | 0,579 | 0,861 | 0,607 | |
| gte-multilingual-reranker-base Q8_0 (`gpustack/…-GGUF`) | Apache-2.0 | 332 MB | — | — | — | — | — | llama.cpp 0.5.0 no lo carga: «unknown model architecture: 'new'» |

Ninguno más pequeño mejora la relación calidad / latencia: el Qwen3 0.6B no es más pequeño (0,6 B frente a 0,57 B de
XLM-R large), tarda el doble y rinde menos; el bge base (278 M) es el único más pequeño con soporte en llama.cpp, pero
pierde 0,05 de MRR y queda por debajo de no usar reranker en Recall@10 y nDCG. Los GGUF descargados viven en
`~/.cache/metrin-modelos/`, fuera del repositorio.

### 11.6 Configuración elegida

**bge-reranker-v2-m3 Q4_K_M, N = 20, 800 runas, `--parallel 1 -c 2048 -b 2048 -ub 2048`, `RERANK_TIMEOUT_MS=3000`.**

- Frente a 30 × 1 200 (el defecto anterior): MRR@10 0,910 frente a 0,929 (−0,019, una o dos consultas de 72), nDCG
  0,654 frente a 0,693, Recall@5 0,676 frente a 0,724; a cambio, **p50 828 ms frente a 1 593 ms (−48 %) y p95 967 ms
  frente a 3 148 ms** en la misma ventana, y 540 MB de RAM en vez de 880.
- Frente a no usar reranker: MRR +0,130, Recall@5 +0,060, nDCG +0,043; Recall@10 −0,032.
- **El objetivo de p95 < 600 ms no se alcanza en esta Mac** con una calidad razonable: lo más rápido con MRR ≥ 0,89
  (15 × 800) da p95 824 ms, y bajar de 600 ms pide N ≤ 10, que deja Recall@10 en 0,470 (peor que sin reranker). Con la
  GPU libre los números bajan (las corridas limpias de §4 daban 2,0 s para 30 × 1 200, como aquí 1,6 s), pero no un
  factor 3.
- El tiempo máximo de 3 s cubre el p95 de las ventanas cargadas (1,9–2,0 s): en el benchmark de §11.7 no se pasó
  ninguna llamada (máximo 2 256 ms). En un pico de la GPU (mientras corría otra evaluación de rerankers) una consulta
  de prueba manual sí se pasó (3 063 ms) y la V2 siguió con el orden RRF, como debe.

Comando (dejado corriendo; log fuera del repositorio):

```bash
nohup /opt/homebrew/bin/llama-server -m ~/.cache/metrin-modelos/bge-reranker-v2-m3-Q4_K_M.gguf \
  --reranking --host 127.0.0.1 --port 8091 --no-webui -c 2048 -b 2048 -ub 2048 --parallel 1 \
  > ~/.cache/metrin-modelos/llama-rerank.log 2>&1 &
echo $! > ~/.cache/metrin-modelos/llama-rerank.pid
# Parar: kill $(cat ~/.cache/metrin-modelos/llama-rerank.pid)
```

Desde Docker: **`http://host.docker.internal:8091`**. En Docker Desktop para Mac basta escuchar en 127.0.0.1:
`host.docker.internal` (con `--add-host host.docker.internal:host-gateway`) llega al loopback de la Mac. Comprobado:

```bash
docker run --rm --add-host host.docker.internal:host-gateway curlimages/curl -s \
  http://host.docker.internal:8091/v1/rerank -H 'Content-Type: application/json' \
  -d '{"query":"¿qué es un metrado?","documents":["El metrado es la cantidad de cada partida.","Arturo Dupont es CEO de Optimiza 360."],"top_n":2}'
# → results: índice 0 con 5,27; índice 1 con −10,98
```

En Linux (el EC2) eso no vale: allí `host-gateway` no llega a lo que escucha en loopback; habría que escuchar en la
IP del puente de Docker o correr el reranker como servicio del compose.

Contenedor de prueba (`docs/traza-ejemplos/recrear-4762.sh`): `V2_BUSQUEDA=hibrida`,
`RERANK_URL=http://host.docker.internal:8091`, `RERANK_TOP_N=20`, `RERANK_MAX_RUNAS=800`, `RERANK_TIMEOUT_MS=3000`;
`RERANK_URL= ./recrear-4762.sh` lo recrea sin reranker.

Reproducir las tablas:

```bash
cd metrin
# 11.2 y 11.4 (calidad; con -rerank-cache la latencia de las variantes que no son la primera de su recorte no vale)
go run ./cmd/evalbusqueda -peso-vector-sin-rerank 0.1 -rerank-url http://127.0.0.1:8091 -rerank-cache \
  -repeticiones 1 -casos "" -top-rerank 30 -rerank-runas 1200 \
  -rerank-variantes "30:1200,20:1200,15:1200,10:1200,30:800,20:800,15:800,10:800,30:500,20:500,15:500,10:500" \
  -cascada "0.02,0.05,0.1,0.2,0.35" -cascada-tipos "como|que_es"
```

### 11.7 Efecto en el benchmark V1 / V2 (`cmd/evalv2`, 194 casos, 212 turnos)

Contenedor de prueba 4762 reconstruido con la integración terminada. V1 con el LLM local (MLX, Qwen2.5-3B + LoRA en
`:8080`) y con la página que ya no mete la pregunta actual en `hilo`. V2 con la búsqueda híbrida de fragmentos, sin y
con el reranker elegido. Corridas: V1 + V2 con reranker el 06-10-2026 a las 05:40 (`eval/resultados/2026-10-06T054011.json`,
es la que publica `eval/V1_VS_V2.md`); V2 sin reranker a las 06:17 (`eval/resultados/2026-10-06T061703.json`,
`RERANK_URL= ./docs/traza-ejemplos/recrear-4762.sh`).

| Métrica | V1 (MLX, hilo arreglado) | V2 sin reranker | V2 con reranker (20 × 800) |
|---|---|---|---|
| **PROCEDURAL ANSWER SUCCESS** | 0,0 % (0/55) | **60,0 % (33/55)** | **60,0 % (33/55)** |
| Tasa de invención (turnos con contenido) | 17,1 % (30/175) | 0,0 % (0/173) | 0,0 % (0/174) |
| Tasa de invención por paso | 9,8 % (55/563) | 0,0 % (0/192) | 0,0 % (0/204) |
| Falsa abstención | 8,0 % (13/163) | 6,1 % (10/163) | 5,5 % (9/163) |
| MRR@10 | 0,163 | 0,551 | 0,557 |
| Recall@5 | 0,141 | 0,492 | 0,494 |
| Recall@10 | 0,191 | 0,535 | 0,536 |
| nDCG@10 | 0,130 | 0,505 | 0,507 |
| Precisión de clasificación | 37,6 % (73/194) | 72,2 % (140/194) | 73,2 % (142/194) |
| Éxito de la tarea | 5,7 % (11/194) | 64,4 % (125/194) | 63,9 % (124/194) |
| Latencia cliente p50 / p95 (ms) | 7 147 / 28 141 | 2 / 24 | 8 / 1 665 |
| Latencia servidor (traza) p50 / p95 (ms) | 7 139 / 28 101 | 0,8 / 22 | 3,3 / 1 659 |
| RAM del contenedor (media) | 249 MiB | 322 MiB (1 muestra) | 274 MiB, más 540 MB del reranker en la Mac |

**Dónde actúa el reranker.** Hoy solo en la clase `fragmento` de la V2 (procedimientos implícitos y evidencia de
respaldo); procedimientos, conceptos y errores frecuentes siguen con el BM25 propio de la V2 (por eso el `motivo` de
las respuestas sigue diciendo «sin_reranker: puntaje híbrido sin reordenar»: habla de ese recuperador, no del de
fragmentos). En la traza V2, la etapa `rerank` del recuperador híbrido **corrió en 27 de 212 turnos** (estado `ok`,
`reordenado: true`, `llamadas: 1`, nunca `noop`), sin ningún tiempo agotado ni error; su tiempo fue p50 1 537 ms,
p95 2 030 ms, máximo 2 256 ms (frente al límite de 3 000). En los otros 185 turnos la etapa queda `omitida`,
`no_tomada` o `respaldo` porque el plan se resolvió con un procedimiento o concepto.

**Qué cambió, caso por caso** (los otros 190 casos dan exactamente lo mismo):

| Caso | Pregunta | Sin reranker | Con reranker | Efecto |
|---|---|---|---|---|
| proc-046 | «instalacion del s10 en una laptop nueva paso a paso» | SIN_EVIDENCIA | responde con la sección del Manual de Instalación (MRR 0 → 1, 2 fotos) | **mejora**: una falsa abstención menos |
| proc-024 | «ingreso IC con factura del proveedor paso a paso» | pide aclaración | responde un procedimiento implícito (tipo correcto, procedimiento equivocado) | neutro: el PAS sigue fallando |
| proc-051 | «como reemplazo un recurso por otro en todas las partidas del presupuesto» | pide aclaración | procedimiento implícito, tipo correcto | neutro: el éxito sigue fallando |
| proc-050 | «cómo modifico una partida» | pide aclaración (cuenta como éxito) | responde con la sección «2.2 Configuración (9)» del Manual de Presupuestos (un procedimiento implícito que no es el esperado) | **empeora** el éxito de la tarea |

**Lectura.** El reranker **no empeora el PAS (60,0 % en los dos) ni la invención (0 % en los dos)**; quita una falsa
abstención (6,1 % → 5,5 %) y sube MRR en 0,006, pero cuesta un caso de éxito de la tarea (64,4 % → 63,9 %), 1,6 s de
p95 en la V2 y 540 MB de RAM. Con 55 casos PAS y 163 con respuesta, un caso son 1,8 y 0,6 puntos: **el efecto neto es
nulo dentro del ruido**. Recomendación: dejarlo encendido solo en el contenedor de prueba para seguir observándolo en
la traza, y **apagado por defecto** (`RERANK_URL` vacía) donde importe la latencia; el sitio donde sí podría mover el
PAS es el recuperador de procedimientos y conceptos, que hoy no pasa por el reranker.

> Superado el mismo 06-10-2026: con el reranker también para elegir procedimientos y conceptos el PAS sube a 81,8 %
> (§12), y el usuario decidió dejarlo encendido en la configuración de despliegue (docs/V2-RAG-PROCEDURAL.md §13).

## 12. Reranker para ELEGIR procedimientos y conceptos (06-10-2026, tercera medición)

Decisión del usuario: «pruébalo para elegirlo». Hasta aquí el reranker solo reordenaba la clase `fragmento`; el
procedimiento o el concepto que responde lo elegía el BM25 propio de `internal/v2/conocimiento`. Código:
`internal/v2/conocimiento/rerank_clases.go` (con sus pruebas en `rerank_clases_test.go`), `V2_RERANK_CLASES` en
`internal/v2/config.go` y el cableado en `cmd/rag/v2.go`.

### 12.1 Dónde fallaba la elección (medido antes de tocar nada)

Con las 112 preguntas de un solo turno del benchmark que tienen `procedimiento_esperado` (agregado, sin mirar casos):

| Dónde está el procedimiento esperado en el BM25 propio | Preguntas |
|---|---|
| 1.º del orden BM25 | 94 (83,9 %) |
| entre los 5 primeros / 10 / 20 | 108 / 111 / **112 (100 %)** |
| 1.º después de que el núcleo reordena por puntaje (cobertura) | 86 (76,8 %) |
| 1.º y con cobertura ≥ 0,6 (pasa el umbral sin la regla de dominio) | 60 (53,6 %) |

Conceptos: 32 de 33 primeros en el BM25 (33 de 33 entre los 5 primeros). El cuello de botella no es encontrar el
candidato (está entre los 20 primeros en todas), sino **ordenarlo primero y darlo por bueno**. Por eso los candidatos
que se reordenan son los 20 primeros del BM25 propio y **no se suman los de la búsqueda híbrida**: la híbrida indexa
fragmentos, no procedimientos, y no hay nada que recuperar que el BM25 propio no traiga ya.

### 12.2 Qué se hizo

- **Texto representativo** (`Base.TextoRerank`): procedimiento → «título (módulo)», objetivo, aliases y preguntas del
  YAML; concepto → «término (sinónimos): definición». Recortado a `RERANK_MAX_RUNAS` (800) en el último espacio. Media
  medida en 40 llamadas reales: 737 runas por documento.
- **Consulta**: la pregunta original más lo que la normalizada añade (la misma que la de fragmentos; sin alias).
- **Candidatos**: los `RERANK_TOP_N` (20) primeros del orden léxico; los demás quedan detrás, en su orden.
- **Puntaje y orden**: el logit pasa por la sigmoide (la misma conversión que `fragmento`), así que 0,6 ⇔ logit ≥ 0,405.
  El logit crudo va en `Candidato.Rerank` y el rango léxico en `meta.rango_lexico`.
- **Empate** (constructor `ambiguo` y núcleo `elegirProcedimiento`): con los dos candidatos del reranker, se compara la
  diferencia de logits con logit(0,6) ≈ 0,405 (`MargenRerankEmpate`), no la sigmoide, que se satura cerca de 1. Una
  prueba de contrato comprueba que los dos márgenes coinciden.
- **Aceptación** (variante B′, la elegida): el constructor acepta con el MAYOR entre la sigmoide del logit y la
  cobertura léxica del mismo candidato. El reranker ordena y puede aceptar lo que la cobertura no alcanzaba, pero no
  rechaza lo que el léxico ya aceptaba (ver 12.4).
- **Degradación**: sin reranker, con error, con respuesta incompleta o no finita, o con el tiempo agotado
  (`RERANK_TIMEOUT_MS`), el orden y el puntaje léxicos de siempre, idénticos (`TestRerankClases_FalloNoCambiaNada`
  compara candidato por candidato). Queda en `motivo` («reranker_error (procedimiento): …; orden y puntaje léxicos») y
  en la traza: etapa `rerank`, dato `clases` con, por clase, `reordenado`, `fallo`, `ms`, `llamadas`, el orden léxico y
  el del reranker con sus logits; la etapa pasa a `respaldo` si falló.
- **Lo que no pasa por el reranker**: los errores frecuentes (TROUBLESHOOTING, clase `error`: no se midió) y la búsqueda
  interna del constructor que, tras explicar un concepto, busca por el término el procedimiento que ofrece.
- **Configuración**: `V2_RERANK_CLASES=fragmento,procedimiento,concepto` (defecto; `ninguna` = nada). Solo actúa con
  `RERANK_URL`. `/health` dice las clases que de verdad se reordenan (`v2.busqueda.rerank_clases`).

### 12.3 Benchmark completo de la V2 (194 casos, 212 turnos)

`go run ./cmd/evalv2 --url http://127.0.0.1:4762 --contenedor metrin-traza-prueba --versiones v2`, contenedor de prueba
reconstruido con `docs/traza-ejemplos/recrear-4762.sh` y el reranker de la Mac (Metal, `:8091`, 20 × 800). Misma
imagen para A, B y C; B′ es B más la regla de aceptación de 12.2. La Mac estaba compartida (carga media 8–18 durante
las corridas): las latencias valen para comparar entre ellas, no como cifra absoluta.

| Métrica | A · reranker solo en fragmento (como hoy) | B · fragmento + procedimiento + concepto, sigmoide | **B′ · ídem, el reranker no quita (elegida)** | C · sin reranker |
|---|---|---|---|---|
| **PROCEDURAL ANSWER SUCCESS** | 60,0 % (33/55) | 76,4 % (42/55) | **81,8 % (45/55)** | 60,0 % (33/55) |
| Acierto de procedimiento | 63,2 % (79/125) | 64,8 % (81/125) | **76,8 % (96/125)** | 63,2 % (79/125) |
| Acierto de concepto | 75,0 % (27/36) | 80,6 % (29/36) | 80,6 % (29/36) | 75,0 % (27/36) |
| Precisión de clasificación | 73,2 % (142/194) | 77,8 % (151/194) | 77,8 % (151/194) | 72,2 % (140/194) |
| Tasa de invención (turnos con contenido) | 0,0 % (0/174) | 0,0 % (0/166) | 0,0 % (0/180) | 0,0 % (0/173) |
| Falsa abstención | 5,5 % (9/163) | 9,8 % (16/163) | **1,8 % (3/163)** | 6,1 % (10/163) |
| Abstención correcta | 100 % (12/12) | 100 % (12/12) | 91,7 % (11/12) | 100 % (12/12) |
| Tasa de SIN_EVIDENCIA | 9,9 % (21/212) | 13,7 % (29/212) | 7,1 % (15/212) | 10,4 % (22/212) |
| MRR@10 | 0,557 | 0,608 | **0,666** | 0,551 |
| Recall@5 / Recall@10 | 0,494 / 0,536 | 0,510 / 0,559 | 0,566 / 0,626 | 0,492 / 0,535 |
| Recall de pasos | 0,606 | 0,774 | 0,828 | 0,606 |
| Éxito de la tarea | 63,9 % (124/194) | 67,5 % (131/194) | **71,6 % (139/194)** | 64,4 % (125/194) |
| Latencia cliente p50 / p95 (ms) | 4 / 1 666 | 651 / 2 045 | 883 / 2 051 | 4 / 55 |
| Latencia servidor (traza) p50 / p95 (ms) | 2,2 / 1 664 | 642 / 1 970 | 822 / 2 014 | 1,6 / 50 |
| RAM del contenedor de Metrín (media) | 303 MiB | 305 MiB | 302 MiB | 300 MiB |

JSON crudo: A `eval/resultados/2026-10-06T094316.json`, B `2026-10-06T093955.json`, B′ `2026-10-06T094835.json`,
C `2026-10-06T094444.json`. Con 55 casos PAS un caso son 1,8 puntos; con 163 casos con respuesta, 0,6. Repetida B′
con el código final y la configuración por defecto (sin pasar `V2_RERANK_CLASES`; `2026-10-06T101635.json`): las
mismas cifras de calidad, una a una; latencia p50 764 ms y p95 1 751 ms (otra ventana de carga).

Por categoría (B′): PROCEDURE PAS 82 % y clasificación 93 % (A: 60 % y 78 %); CONCEPT éxito 82 % (A: 77 %);
NAVIGATION éxito 67 % y SIN_EVIDENCIA 4 % (igual que A); TROUBLESHOOTING éxito 58 % (A: 55 %); CONFIGURATION éxito
38 % (A: 33 %). La latencia la pagan las preguntas que buscan procedimientos o conceptos: p50 de PROCEDURE 1,3 s, de
CONCEPT 0,5 s (en A, milisegundos), porque cada búsqueda de esas clases es una llamada al reranker.

### 12.4 Por qué B′ y no B (y qué se arriesga)

B mejoró el PAS pero empeoró la falsa abstención (5,5 % → 9,8 %): NAVIGATION pasó de 4 % a 24 % de SIN_EVIDENCIA y
CONFIGURATION de 14 % a 27 %. Mirado en agregado (la traza de cada turno trae los logits): en los turnos que B pasó a
SIN_EVIDENCIA, el reranker dejaba primero **al mismo candidato que el léxico** (rango léxico 1 en 5 de 5 de
NAVIGATION, 3 de 3 de CONFIGURATION y 4 de 5 de PROCEDURE), pero con un logit cercano a 0 (mediana −0,02 en
NAVIGATION), bajo el umbral de la sigmoide. Un «¿dónde está…?» frente al texto de un procedimiento («Registrar…») no
da un logit alto aunque sea el procedimiento correcto: la escala absoluta del logit frente a un texto representativo
(no un pasaje que responde) no está calibrada. B′ cambia una sola regla general, sin constantes nuevas: para aceptar
cuenta el mayor de los dos puntajes. Resultado: NAVIGATION y CONFIGURATION vuelven a lo de A y el PAS sube 3 casos más.

Lo que hay que saber antes de usarlo:

1. **B′ se diseñó después de ver B.** La regla es general (no mira casos), pero se eligió con este mismo benchmark;
   no hay un conjunto aparte con procedimientos etiquetados para confirmarla. Si se suman preguntas reales, conviene
   repetir la comparación.
2. **Una abstención correcta menos (12/12 → 11/12).** Una pregunta sin evidencia en los manuales (un pago en una
   moneda que S10 no documenta) recibe el procedimiento más cercano (el pago con cheque): su cobertura léxica (0,46)
   queda bajo el umbral, pero la regla de dominio del constructor lo acepta ahora que va primero, aunque el reranker
   le daba logit −1,4. No inventa pasos (la invención sigue en 0 %), pero responde algo que no se preguntó. No se
   corrigió para no ajustar la regla a un caso del benchmark.
3. **Latencia.** Con Metal, p50 de 0,9 s y p95 de 2,0 s en la V2 (A: 4 ms y 1,7 s). Sin GPU, ver 12.5.

Regla del encargo: queda por defecto la variante que gana en PAS sin empeorar la invención → **B′**
(`V2_RERANK_CLASES=fragmento,procedimiento,concepto`).

### 12.5 RAM y latencia del reranker en Docker (solo CPU)

El servicio `reranker` del `docker-compose.yml` (imagen `ghcr.io/ggml-org/llama.cpp:server` del 28-09-2026, los mismos
`--reranking --parallel 1 -c 2048 -b 2048 -ub 2048`, healthcheck del compose) se probó con `docker run` y el mismo
montaje (Docker no tenía redes libres para un proyecto aparte; los contenedores del proyecto no se tocaron). Arranca
sano en ~25 s. Carga de trabajo: 40 llamadas reales de la V2 al elegir procedimiento (32 de 20 candidatos, el resto
de 5–19; 737 runas de media por documento), cada una contra el contenedor (solo CPU, la VM de Docker Desktop con 12
hilos arm64) y contra el llama-server nativo con Metal, intercaladas. La Mac estaba cargada por otros procesos
(carga media 15–22), así que la cifra de CPU es un orden de magnitud, no la de un servidor dedicado.

| Dónde | p50 | p95 | Máximo | RAM |
|---|---|---|---|---|
| Docker, solo CPU (servicio del compose) | 16 680 ms | 45 612 ms | 61 278 ms | 370 MiB en reposo; pico 782 MiB (`docker stats`, 300 muestras) |
| Mac nativo, Metal (`:8091`) | 913 ms | 1 178 ms | 1 320 ms | 540 MB de huella (§11.3) |

Solo las llamadas de 20 candidatos: CPU p50 18,7 s (≈ 0,94 s por documento), Metal p50 0,98 s. Las 40 llamadas en
CPU tardaron más de 3 s (la más rápida, 4,0 s con 5 candidatos): con `RERANK_TIMEOUT_MS=3000` la V2 sigue con el orden
léxico (la degradación no cambia el resultado: `TestRerankClases_FalloNoCambiaNada`). **En la CPU sin GPU de un
servidor la latencia no está medida.**
`mem_limit` del servicio: 1 GB.

Reproducir:

```bash
cd s10-conocimiento
# Contenedor de prueba con cada variante (reranker de la Mac en :8091)
V2_RERANK_CLASES=fragmento ./docs/traza-ejemplos/recrear-4762.sh                        # A
V2_RERANK_CLASES=fragmento,procedimiento,concepto ./docs/traza-ejemplos/recrear-4762.sh  # B′ (y el defecto)
RERANK_URL= ./docs/traza-ejemplos/recrear-4762.sh                                        # C
cd metrin && go run ./cmd/evalv2 --url http://127.0.0.1:4762 --contenedor metrin-traza-prueba --versiones v2 \
  --md /tmp/variante.md        # --md fuera de eval/: sin él, sobrescribe eval/V1_VS_V2.md con solo la V2
# Pruebas
go test ./internal/v2/... -run 'Rerank|Aviso|Empate|Aceptacion'
```

(B se midió con el mismo código sin la regla de aceptación de 12.2; para repetirla hay que quitar el caso
`x.Rerank != nil && ClasesRerankDefecto[clase]` de `Constructor.candidatos`.)
