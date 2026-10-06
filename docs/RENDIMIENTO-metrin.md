# Metrín: cifras de rendimiento y tiempos que hay que mejorar

Fecha: 06-10-2026. Rama `feat/metrin-v2-traza`. Este documento junta lo medido hoy y fija una meta para cada cifra.
El detalle de cada medición está en la fuente que se indica. La lista general de pendientes está en
[`PENDIENTES.md`](../PENDIENTES.md).

**Cómo leer estas cifras:**

- **Dónde se midió:** todo en la Mac de desarrollo (M4 Pro, 24 GB), casi siempre con otros procesos ocupando la CPU
  y la GPU (carga media de 15 a 60). Las latencias pueden estar infladas; los aciertos no, porque la decodificación
  es determinista.
- **Calidad optimista:** se midió sobre 194 preguntas **sintéticas** (`metrin/eval/v2_oro.jsonl`), y algunas reglas
  se diseñaron mirando ese mismo conjunto.
- **Producción sin medir:** nada se midió en un servidor de producción (sin GPU, 2–4 vCPU, 8 GB).

---

## 1. Tiempos

| Qué | Hoy | Meta | Por qué importa | Fuente |
|---|---|---|---|---|
| **Reranker solo en CPU** (Docker, sin GPU) | **p50 16,7 s · p95 45,6 s** por llamada | p95 < 600 ms | Con `RERANK_TIMEOUT_MS=3000`, un servidor sin GPU espera 3 s y vuelve al orden léxico: la calidad cae de 81,8 % a 60 % y la respuesta tarda ~3 s más. **Es lo más urgente.** | `docs/V2-RAG-PROCEDURAL.md` §13 |
| Reranker en la Mac (Metal, 20 candidatos × 800 runas) | p50 0,83 s · p95 0,97 s | p95 < 600 ms | Con 10 candidatos sí baja de 600 ms, pero el Recall@10 cae a 0,47, peor que sin reranker (0,63) | `metrin/eval/BUSQUEDA.md` §11 |
| **Respuesta V2 completa, con reranker** | p50 0,9 s · **p95 2,1 s** | p95 < 1 s | Sin reranker tarda 4 ms / 55 ms: casi todo el tiempo es del reranker | `docs/V2-RAG-PROCEDURAL.md` §14 |
| **Respuesta V1** (MLX encendido) | **p50 7,1 s · p95 28 s** | — | Además inventa en el 17 % de las respuestas. Conviene pasar el tráfico a la V2 antes que acelerar V1 | `metrin/eval/V1_VS_V2.md` |
| LLM para redactar (si se usara) | 4–12 s por plan con Metal · 16,6 s en CPU (qwen3.5-2b) · ~80 s en CPU (7B, estimado) | < 5 s (`GENERATION_TIMEOUT_MS`) | No cabe. Por eso la V2 redacta con plantillas; hoy no afecta | `metrin/eval/MODELOS.md` |
| LLM para decidir «¿hay evidencia suficiente?» | 2,4–9 s en CPU | < 1,5 s (`DECISION_TIMEOUT_MS`) | No cabe. Hoy se decide con reglas | `metrin/eval/MODELOS.md` |
| LLM para decidir el tipo de pregunta (qwen3.5-2b, CPU) | 0,42 s | < 1,5 s | Cabe. Falta integrarlo en la V2 | `metrin/eval/MODELOS.md` |
| Jev local (`:8765`) | > 20 s en una prueba | < 1,5 s | Hoy la V2 no lo usa (`DECISION_ENGINE=reglas`) | informe del núcleo V2 |
| Búsqueda híbrida (sin reranker) | +22 ms en el p95 | — | Barata, pero sin reranker no cambia ninguna respuesta del benchmark | `docs/V2-RAG-PROCEDURAL.md` §14 |
| Arranque de los índices de la V2 | ~1,1 s | — | Aceptable | informe de conocimiento V2 |

## 2. Memoria

El EC2 de producción tiene 8 GB, compartidos con el resto de servicios.

| Qué | Hoy | Comentario |
|---|---|---|
| Reranker (llama-server, bge-reranker-v2-m3 Q4_K_M) | 370 MiB en reposo · **782 MiB de pico** | El compose le pone `mem_limit: 1g`. Con Metal en la Mac se midieron 540–880 MB |
| Búsqueda híbrida | **+126 MiB** (el contenedor V2 pasa de 200 a 326 MiB) | Sin reranker no aporta calidad |
| Índice BM25 | 32 MiB en memoria · 18–19 MiB en disco | Carga en 57–141 ms |
| Clasificador MiniLM (opcional) | +470 MB | +13 puntos de clasificación |
| LLM de decisión qwen3.5-2b (opcional) | 2,8 GB de RSS | El límite es la CPU, no la RAM |

## 3. Calidad (V2 por defecto: reranker en las tres clases)

| Métrica | Hoy | Meta | Comentario |
|---|---|---|---|
| PROCEDURAL ANSWER SUCCESS | 81,8 % (45/55) | ≥ 90 % **con preguntas reales** | Sin reranker: 60 %. V1: 0 % |
| Acierto de procedimiento | 76,8 % | ≥ 90 % | |
| Acierto de concepto | 80,6 % | ≥ 90 % | |
| **Clasificación del tipo de pregunta** | **77,8 %** | ≥ 90 % | La causa de la mayoría de los fallos que quedan. El embebedor estático reconoce el tema pero no el acto («qué es», «cómo», «dónde») |
| MRR@10 de la V2 | 0,666 | ≥ 0,8 | El MRR de la búsqueda sola sobre su conjunto dorado es 0,910 con reranker |
| Recall de imágenes | ~33–35 % | subir | Hoy solo el **58,9 %** de los pasos de los YAML tiene foto |
| Abstención correcta | 11/12 | 12/12 | El fallo: «criptomonedas» recibe el procedimiento de pago con cheque |
| Invención | 0 % | mantener en 0 % | V1: 17,1 % |
| Falsa abstención | 1,8 % | mantener | |
| Paso ↔ foto | 100 % | mantener | |

## 4. Por dónde empezar

1. **El reranker en el servidor real.** Medirlo en la máquina donde va a correr, o en un contenedor limitado a 2–4
   vCPU sin GPU. Si no llega, comparar tres salidas:
   - GPU;
   - reranker más pequeño o cuantizado más agresivo, con menos candidatos solo para procedimiento y concepto;
   - usarlo solo para desempatar, cuando el BM25 duda.

   Meta: la V2 completa con p95 < 1 s sin perder el 81,8 %.
2. **La clasificación del tipo de pregunta (77,8 %).** Probar MiniLM (+470 MB) o qwen3.5-2b solo cuando las reglas
   duden (0,42 s en CPU). Meta: ≥ 90 %.
3. **Preguntas reales.** Sin un conjunto real ninguna meta de calidad se puede confirmar. Es el requisito para todas
   las demás.
4. **Más fotos en los pasos** (58,9 %). Revisar si los pasos sin foto tienen captura en la fuente.

## 5. Cómo volver a medir

Desde `s10-conocimiento/metrin` (go en `/opt/homebrew/bin/go`):

```bash
# Calidad y latencia de punta a punta, V1 frente a V2 (contenedor de prueba 4762)
../docs/traza-ejemplos/recrear-4762.sh
go run ./cmd/evalv2 --url http://127.0.0.1:4762 --contenedor metrin-traza-prueba

# Búsqueda y reranker (conjunto dorado de 72 consultas)
go run ./cmd/evalbusqueda

# Modelos locales de decisión y generación
../.venv/bin/python3 ../herramientas/benchmark_modelos.py --help

# RAM de un contenedor
docker stats --no-stream metrin-traza-prueba
```

Para que las latencias sean comparables, medir con la máquina en reposo y anotar la carga (`uptime`) junto a cada
cifra.
