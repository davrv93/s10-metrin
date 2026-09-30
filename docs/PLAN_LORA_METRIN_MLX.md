# Plan piloto: LoRA de estilo para Metrín con MLX-LM

Guía para ejecutar un primer ajuste local en un Mac con Apple Silicon. El objetivo es mejorar el tono y el formato de Metrín; **no** enseñarle procedimientos del S10 ni sustituir el RAG. Este documento describe el trabajo, no significa que el entrenamiento ya se haya ejecutado.

## Objetivo y límites

- Identidad: **Metrín**, asistente de Optimiza 360 para ERP S10 y construcción.
- Registro: español neutro, cercano, profesional y claro; sin gentilicios ni jerga regional. Humor sutil y emojis, solo si aportan.
- El adaptador debe mejorar saludos, claridad, preguntas de seguimiento, admisión de incertidumbre, agradecimientos y correcciones.
- Los hechos, pasos de S10 y respuestas con fuentes siguen siendo responsabilidad del RAG. No incluir manuales ni instrucciones técnicas inventadas en el dataset de estilo.
- Metrín no debe afirmar que aprendió, guardó, investigará o notificará algo salvo que un flujo real haya completado esa acción.

## Paso 1: comprobar el Mac y preparar el entorno

Confirmar que el Mac usa Apple Silicon, la versión de macOS y la memoria disponible. MLX está dirigido a Apple Silicon; no asumir que una receta o tiempo de entrenamiento aplica a cualquier Mac.

```bash
mkdir -p entrenamiento-metrin/{data,artifacts,eval}
cd entrenamiento-metrin
python3 -m venv .venv
source .venv/bin/activate
python -m pip install --upgrade pip
python -m pip install "mlx-lm[train]"
python -m pip freeze > artifacts/requirements-mlx.txt
```

Fijar en el registro del experimento la versión de MLX-LM, macOS, modelo exacto y licencia. Revisar el acceso y los términos del modelo base antes de descargarlo o redistribuirlo.

## Paso 2: crear y revisar el dataset

MLX-LM admite ejemplos conversacionales `messages` en JSONL. Preparar un ejemplo por línea y separar los archivos para no evaluar con datos vistos durante el entrenamiento:

```text
data/metrin/train.jsonl
data/metrin/valid.jsonl
data/metrin/test.jsonl
```

Formato de ejemplo, exclusivamente de tono:

```json
{"messages":[{"role":"system","content":"Eres Metrín, asistente de Optimiza 360. Respondes en español neutro, con calidez, claridad y profesionalismo. No inventas datos ni afirmas acciones que no realizaste."},{"role":"user","content":"No entiendo la respuesta."},{"role":"assistant","content":"Vamos paso a paso. ¿Qué parte te gustaría que explique de otra manera?"}]}
```

Como piloto, comenzar con **300 ejemplos curados en total**, por ejemplo 240 para entrenamiento, 30 para validación y 30 para prueba. No es una cantidad que garantice mejora: importa más la consistencia y diversidad que generar cientos de respuestas repetitivas.

Distribuir los ejemplos entre saludos y conversación breve, explicaciones claras, preguntas ambiguas, falta de evidencia, límites de alcance, agradecimientos, logros y feedback correctivo. Mantener las respuestas técnicas como plantillas de estilo sin datos del S10. La respuesta de “no tengo evidencia” debe ser transparente y no prometer investigación, almacenamiento ni avisos futuros.

Antes de entrenar, validar que cada línea sea JSON válido; eliminar duplicados, datos personales, secretos y conversaciones privadas sin autorización; verificar procedencia/licencia y revisar manualmente una muestra. Guardar una copia inmutable del dataset y su hash.

## Paso 3: establecer línea base y rúbrica

Crear entre 20 y 30 prompts que no aparezcan en el entrenamiento. Ejecutarlos primero contra el modelo base y guardar entradas, salidas, configuración y versión. Puntuar de 1 a 5:

| Criterio | Qué revisar |
|---|---|
| Español neutro | No aparecen jergas, gentilicios ni regionalismos. |
| Claridad | Respuesta directa, ordenada y de longitud adecuada. |
| Personalidad | Cálida y profesional, con humor moderado cuando encaja. |
| Honestidad | Reconoce ambigüedad o falta de evidencia; no inventa capacidades. |
| No regresión | No fabrica procedimientos, cuentas, citas ni hechos de S10. |

Incluir casos como “Hola”, “Gracias”, “No entiendo”, “Está mal”, una pregunta ambigua y una pregunta técnica cuya respuesta no esté en el contexto. Comparar base y adaptado con los mismos prompts y revisión humana. No decidir por pérdida/perplexity solamente.

## Paso 4: entrenamiento corto

Elegir y fijar un modelo instruct compatible con MLX-LM y con los requisitos de licencia. **Llama 3.2 3B Instruct** es un candidato para probar; comprueba acceso, licencia, disponibilidad de pesos MLX y compatibilidad antes de comprometer el flujo de exportación. No tratar Qwen u otro modelo como intercambiable: soporte de entrenamiento MLX no garantiza que `mlx_lm.fuse` pueda exportarlo a GGUF.

Comando inicial para un piloto, sujeto a la versión instalada:

```bash
MODEL="mlx-community/Llama-3.2-3B-Instruct-4bit"

mlx_lm.lora \
  --model "$MODEL" \
  --train \
  --data ./data/metrin \
  --batch-size 2 \
  --num-layers 16 \
  --iters 300 \
  --learning-rate 1e-5 \
  --adapter-path ./artifacts/metrin-lora \
  --save-every 100 \
  --steps-per-eval 50 \
  --max-seq-length 1024 \
  --mask-prompt
```

En la CLI actual, el parámetro para capas es `--num-layers`, no `--lora-layers`; `--iters` indica pasos de optimización, no épocas. Con 240 ejemplos, batch 2 y 300 pasos, el número efectivo de pasadas depende del muestreo y de la versión/configuración. Confirmar las opciones con `mlx_lm.lora --help`; bajar longitud o batch si hay presión de memoria. Hacer primero un ensayo breve y registrar duración y memoria. **No asumir 20–40 minutos ni un tamaño fijo del adaptador**: medir en el equipo real.

Los checkpoints cada 100 pasos permiten recuperar progreso. Conservar logs, configuración y errores junto al artefacto.

## Paso 5: probar el adaptador

```bash
mlx_lm.chat \
  --model "$MODEL" \
  --adapter-path ./artifacts/metrin-lora \
  --system-prompt "Eres Metrín, asistente de Optimiza 360. Respondes en español neutro, con calidez, claridad y profesionalismo. No inventas datos ni afirmas acciones que no realizaste."
```

Ejecutar la rúbrica con las mismas entradas de la línea base. Revisar especialmente que el adaptador no vuelva más verboso al bot, no fuerce chistes/emojis, no invente conocimiento técnico y no haga promesas de aprendizaje. Si la calidad no mejora o aparece una regresión, no promover el adaptador; ajustar dataset/parámetros o volver al modelo base.

## Paso 6: fusionar y evaluar exportación

La fusión/exportación depende de la arquitectura y del soporte de la versión de MLX-LM. El camino GGUF de `mlx_lm.fuse` está limitado a arquitecturas compatibles (entre ellas Llama, Mistral y Mixtral); confirmar el modelo exacto y la CLI instalada antes de ejecutar. La exportación GGUF de MLX es F16; para un GGUF cuantizado se requiere una etapa posterior con llama.cpp.

Solo como prueba si el modelo elegido está soportado:

```bash
mlx_lm.fuse \
  --model "$MODEL" \
  --adapter-path ./artifacts/metrin-lora \
  --save-path ./artifacts/metrin-fused \
  --dequantize \
  --export-gguf \
  --gguf-path metrin-style-f16.gguf

llama-quantize \
  ./artifacts/metrin-fused/metrin-style-f16.gguf \
  ./artifacts/metrin-fused/metrin-style-Q4_K_M.gguf \
  Q4_K_M
```

Verificar que el archivo se abra, que la cuantización no degrade la rúbrica y que el runtime objetivo soporte el modelo. Mantener el adaptador separado como alternativa reversible; guardar hashes y versiones tanto del modelo base como de cada artefacto.

## Paso 7: decisión de despliegue

El servidor de producción descrito en `CLAUDE.md` es actualmente una **t3.small con 2 GiB de RAM**, compartida con otros servicios. No desplegar allí un modelo 3B basándose en el tamaño del GGUF: medir RAM pico del runtime, contexto, concurrencia y servicios coexistentes en un entorno aislado. Para la primera demo, preferir inferencia local en el Mac o una máquina separada con capacidad medida. Producción requiere aprobación de infraestructura, licencia, seguridad, monitoreo, rollback y pruebas de carga.

## Entregables y criterio de cierre

- Dataset versionado y validado (`train`, `valid`, `test`), hash y registro de procedencia.
- Línea base, rúbrica y comparación humana antes/después.
- Adaptador, logs, configuración reproducible y modelo base identificado por revisión/hash.
- Informe de exportación GGUF (o motivo de incompatibilidad), tamaño real y prueba en runtime objetivo.
- Decisión documentada: promover, iterar o descartar. Sin mejora clara y sin regresiones críticas, no se despliega.

## Referencias oficiales

- [MLX-LM: LoRA y formatos de datos](https://github.com/ml-explore/mlx-lm/blob/main/mlx_lm/LORA.md)
- [MLX-LM: opciones del chat](https://github.com/ml-explore/mlx-lm/blob/main/mlx_lm/chat.py)
- [MLX-LM: fusión y exportación](https://github.com/ml-explore/mlx-lm/blob/main/mlx_lm/fuse.py)
- [llama.cpp: cuantización GGUF](https://github.com/ggml-org/llama.cpp/blob/master/tools/quantize/README.md)
