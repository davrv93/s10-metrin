# Piloto LoRA de estilo Metrín — informe y decisión

Fecha: 2026-09-30. Mac arm64, 24 GB. MLX 0.32.3, mlx-lm 0.31.3.
Base: `mlx-community/Qwen2.5-3B-Instruct-4bit` (sin gate en HF, licencia Qwen).
Se eligió Qwen sobre Llama 3.2 por licencia abierta; consecuencia: `mlx_lm.fuse`
solo soporta llama/mistral/mixtral, así que no hay GGUF. Despliegue = adapter
separado + `mlx_lm.server` (reversible, tal como el plan pide como alternativa).

## Secuencia de modelos (español, edge, ligero)

| Situación | Pieza | Dónde corre | Estado |
|---|---|---|---|
| Embeddings ES | potion-es-int8 256d, 8.9 MB (ya vendored) | Go, en binario | en producción |
| Clasificación intenciones | centroides 4 clases, 38 KB | Go, embarcado (`internal/clasificar`) | entrenado y medido |
| Conversacional estilo | Qwen2.5-3B 4bit + LoRA 16 capas, adapter 25.4 MB | GPU Mac vía `mlx_lm.server` | entrenado y medido |
| Hechos S10 | RAG sin cambios (fuentes mandan) | Go + índice | sin cambios |

## Dataset (solo tono, sin hechos S10)

`generar_dataset.py` determinista (semilla 20260930): 300 ejemplos únicos,
split 244/30/30 por categoría (saludos, explicaciones, ambiguas, sin evidencia,
límites, gracias, correctivo, seguimiento). Hashes en
`artifacts/dataset-hashes.json`, copias inmutables en `artifacts/`.

## Línea base y entrenamiento

24 prompts inéditos en `eval/linea_base.jsonl` (temp 0.0, max 150).
LoRA: batch 2, 300 iters, lr 1e-5, 16 capas, seq 1024, mask-prompt.
Train 0.140, val 0.159, pico 3.16 GB. Adapter sha256 `3e18bf1e4abf5f91…` (25.4 MB),
checkpoints cada 100 pasos conservados.

## Evaluación adapter vs base

`eval/comparacion.md` trae los 24 pares para rúbrica humana 1-5 del plan
(neutro, claridad, personalidad, honestidad, no regresión) — **pendiente tu
revisión**. Automático: largo medio 249 → 69 caracteres (sin verbosidad nueva);
tono Metrín consistente ("Hola. ¿Qué necesitas resolver?"). Lunar menor: ante
"Buenas noches" responde "Hola." sin espejar el momento del día.

## Clasificador (Go puro)

7 clases finas se solapaban (ambigua/aclaración/correctivo); se fusionó a 4:
social, ayuda, limite, trabajo (68 textos, umbral 0.40 calibrado en valid).
Valid 22/28, test 23/28; los 11 fallos restantes caen a `trabajo` (ruta segura
RAG), cero desvíos inseguros. Comandos: `rag entrenar-clasificador`,
`rag clasificar "texto" [--json --todos]`. Aún NO intercala el serve (sin
cambio de conducta); es el siguiente paso tras tu visto bueno.

## Integración compose

`metrin` acepta `LLM_PROVIDER=ollama|mlx` + `MLX_URL` (defecto
`host.docker.internal:8080`; MLX no corre en contenedor Linux: va en el Mac).
Imagen reconstruida y sana en docker (`/health` OK, `clasificar` embarcado OK).
Probar estilo en chat: arranca el server del Mac y pon `LLM_PROVIDER=mlx` en
`.env`, luego `docker compose up -d metrin`.

## Decisión

**Opción B promovida a demo local (2026-09-30):** Qwen2.5-3B + LoRA vía
`mlx_lm.server` (`metrin/servir-mlx.sh`, :8080), `LLM_PROVIDER=mlx` en `.env`,
`RAG_TEMPERATURA=0`. Opción A (Llama+GGUF) descartada por demora de licencia.

Regresión encontrada y corregida: Qwen2.5-3B filtra portugués aun con temp 0 y
orden explícita (fuente 100% española). Guarda en `rag.go`: `parecePortugues`
(à/ã/õ/ç + lista cerrada, cero falsos positivos en ES medidos), un reintento
con refuerzo y fallback a fragmentos. 3 tests nuevos, 19 paquetes en verde.
Verificado en vivo por `/ask` en docker: español limpio, modo respuesta.

Pendiente: tu rúbrica 1-5 en `eval/comparacion.md` y cablear el clasificador al
serve. Producción t3.small ni se evalúa (exige GPU Apple).

## 2026-09-30 tarde: ruta dinámica + rúbrica automática

**Ruta dinámica (Go, `rag.go`):** cada pregunta hace dos pases — general top-K
y curado `manual=Cortex` top-4 (exportador normalizado a `manual=Cortex`,
147 fragmentos). El prompt etiqueta secciones (proyecto vs manuales) y ordena
elaborar en dos partes diciendo de qué lado viene cada cosa. Verificado en
vivo: `[CONOCIMIENTO DEL PROYECTO] … [MANUALES Y DOCUMENTOS] …` con citas.
Test `TestDoblePaseCortexEtiquetaSecciones` en verde.

**Discernimiento NO va en LoRA:** enseñar citas/hechos en el adapter violaría
el plan (hechos mandan del RAG). El discernimiento vive en retrieval+prompt:
dinámico, inspeccionable y reversible.

**Rúbrica automática (`juez.py`, juez = base sin adapter):** 16/24 pares
puntuados; el juez 3B satura todo en 5/4 sin discriminar base vs adaptado.
Conclusión honesta: el juez pequeño no sirve como gate; la rúbrica humana
sigue mandando. `eval/rubrica_auto.jsonl` guarda el intento.

## 2026-09-30 noche: ruta conversacional cableada al serve

Saludos ("hola", "gracias", "jaja") caían al RAG → "sin contexto": el
clasificador existía pero nadie lo llamaba. Ahora `Preguntar` clasifica
primero; social/ayuda/limite responden directo con estilo Metrín
(`Modo: conversacional`, sin retrieval ni fallo registrado), trabajo sigue al
RAG. Tres causas de fragilidad encontradas y medidas:
1) caso/puntuación ("hola" vs "Hola"), 2) tildes ("como estas" vs "cómo estás"),
3) centroides que diluyen cortesías → kNN k=1 sobre 83 ejemplos (38 KB
embarcados). Normalización: minúsculas + pliegue áéíóúü + trim signos (ñ se
conserva). Umbral 0.40, cero desvíos inseguros. 28 tests en verde.
Verificado en docker: hola/jaja/gracias/me das risa/como estas → conversacional.
