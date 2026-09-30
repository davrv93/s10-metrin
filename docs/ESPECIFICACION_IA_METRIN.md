# Especificación de IA para Metrín

Este documento amplía el [plan de implementación](PLAN_IMPLEMENTACION_CHAT_METRIN.md) con los requisitos de modelos, recuperación de conocimiento, aprendizaje, memoria y experiencia conversacional. Es una especificación de trabajo: los tamaños, latencias y métricas objetivo se validarán con pruebas en el hardware elegido.

## Decisiones que prevalecen

- El nombre del asistente es **Metrín**. La especificación de personaje vive en [METRIN.md](s10-conocimiento/METRIN.md).
- Metrín se expresa en español neutro, de forma cercana y profesional, sin gentilicios ni jerga regional.
- Ningún modelo aprende automáticamente de conversaciones o reacciones. El feedback genera datos para revisión; la incorporación exige curación, evaluación y aprobación.
- Toda respuesta basada en conocimiento debe citar fuentes recuperadas. La confianza no se acepta solo porque el LLM la declare: se contrasta con recuperación, clasificación calibrada y reglas del dominio.
- La inferencia online y los procesos offline de ajuste son servicios y ciclos operativos separados.

## 1. Servicio conversacional

### Objetivo

Responder consultas sobre ERP S10, construcción, presupuestos, metrados, planillas y módulos contables usando documentación autorizada y el contexto de la organización.

### Candidatos iniciales

Comparar en una prueba reproducible:

- Llama 3.2 3B instruct cuantizado GGUF.
- Phi-3 Mini instruct cuantizado.
- Qwen 2.5 3B instruct cuantizado.
- Mistral 7B instruct en GGUF Q4_K_M.

No tratar la lista como recomendación definitiva. Verificar licencias, idioma, calidad técnica, memoria total del proceso y licencia de redistribución del modelo exacto antes de adoptar uno.

### Hardware y runtime

- Perfil de referencia solicitado: EC2 t3.large, 8 GB RAM, 2 vCPU, sin GPU. Reservar memoria para SO, Docker, API, workers y base de datos; el modelo no dispone de los 8 GB completos.
- Probar `llama.cpp` como primera opción de inferencia CPU con GGUF. Evaluar vLLM únicamente si un hardware posterior con GPU lo justifica.
- Mantener inferencia en un contenedor separado. Limitar concurrencia, longitud de contexto y generación para evitar agotar RAM.
- Parámetros iniciales para evaluación, no valores finales: temperature 0.2–0.4 para soporte técnico, top_p 0.85–0.95, repeat_penalty 1.05–1.15 y límite de salida según tipo de consulta. Registrar parámetros por versión.
- Objetivo de TTFT menor a 3 s: medir por modelo, cuantización, prompt, contexto y concurrencia. No prometerlo en CPU sin benchmark; transmitir tokens al cliente conforme se generan.

### Streaming

El cliente crea un mensaje con `client_message_id`. La API responde con `request_id` y emite eventos SSE tipados (`started`, `token`, `source`, `completed`, `error`). Al reconectar, el cliente retoma por `request_id` o consulta el estado persistido. No guardar cada token por separado: persistir mensaje final y estado de ejecución.

### Entregables de esta fase

- Tabla de modelos con RAM pico del proceso completo, tokens/s, TTFT y evaluación humana de español técnico.
- Dockerfile/comando del runtime elegido y configuración reproducible.
- Servicio mínimo HTTP de inferencia y script local de evaluación. El prototipo Python/FastAPI puede servir para medir el modelo; la integración final debe respetar la arquitectura Go/Rust ya definida.
- Informe de carga con 1, 2 y concurrencia objetivo; p50/p95 de TTFT, tokens/s, RAM y tasa de errores.

## 2. Clasificación de intención y tópico

### Taxonomía inicial

1. `saludo`
2. `despedida`
3. `agradecimiento`
4. `consulta_presupuestos_s10`
5. `consulta_metrados`
6. `consulta_planillas`
7. `consulta_contabilidad`
8. `consulta_compras`
9. `consulta_costos`
10. `consulta_configuracion_s10`
11. `solicitud_reporte`
12. `solicitud_comparacion`
13. `soporte_error`
14. `soporte_como_hacer`
15. `cotizacion_servicio`
16. `consulta_comercial`
17. `solicitud_humano`
18. `feedback_respuesta`
19. `solicitud_aprendizaje`
20. `fuera_de_dominio`
21. `ambiguo`

Las entidades incluyen módulo, proyecto, obra, periodo, versión de presupuesto, código de partida y documento, solo cuando aparecen o se pueden resolver con permisos. La taxonomía y los alias se versionan.

### Diseño

- Comparar BETO, MiniLM multilingüe ajustado y DistilBERT multilingüe en CPU y memoria disponible.
- Generar ejemplos sintéticos con un modelo grande solo como borradores; deduplicar, revisar y mantener un set de evaluación humano separado.
- Salida versionada: `{ "intent": "...", "confidence": 0.0, "entities": {}, "model_version": "..." }`.
- Calibrar el umbral con matriz de confusión y costo de errores. Umbral bajo o clase ambigua deriva a una pregunta de aclaración o a un humano; no ejecuta acciones.
- Inferencia batch menor a 50 ms es una meta que debe medirse en el host de producción, con tamaño de lote y CPU indicados.

### Entregables

- Esquema JSONL de entrenamiento/evaluación con procedencia, licencia y split.
- Pipeline de generación asistida, limpieza y fine-tuning con HuggingFace Trainer.
- API de inferencia batch y reporte de latencia/precisión por clase.

## 3. Recuperación de conocimiento (RAG)

### Fuentes

- Manuales S10 con acceso autorizado y referencia a página/sección.
- Normativa de construcción: solo documentos oficiales, versión y fecha vigentes verificadas.
- Documentación interna de Optimiza 360, con permisos y nivel de confidencialidad.
- Conversaciones resueltas, anonimizadas, revisadas y aprobadas para reutilización.

### Pipeline

```text
Fuentes permitidas
  -> extracción de texto/OCR
  -> normalización y clasificación de sensibilidad
  -> fragmentos con documento, página, sección, versión y permisos
  -> embeddings + índice léxico
  -> búsqueda híbrida (semántica + BM25/FTS)
  -> fusión de resultados
  -> reranker opcional
  -> filtro de permisos y umbral de relevancia
  -> respuesta con citas verificables
```

### Decisiones a probar

- Comparar `paraphrase-multilingual-MiniLM-L12-v2`, `BAAI/bge-m3` y `intfloat/multilingual-e5` con un set etiquetado de consultas de construcción/S10. Respetar prefijos requeridos por el modelo elegido.
- Para la primera versión, preferir PostgreSQL con `pgvector` y búsqueda de texto completo si el volumen lo permite; reduce servicios y aplica permisos en una fuente común. Comparar con Qdrant cuando el volumen, latencia o operación lo justifiquen. FAISS/Chroma quedan para prototipos o índices locales, no para ser fuente de verdad.
- Chunking inicial: separar por títulos, tablas y secciones; medir fragmentos de 400–800 tokens con solape de 60–120 tokens. Nunca separar una tabla de su encabezado o leyenda.
- Guardar metadatos: `document_id`, nombre, versión/fecha, página, sección, tipo de fuente, organización, sensibilidad, URL autorizada y hash del contenido.
- Recuperación híbrida: semántica para paráfrasis y BM25/FTS para términos exactos como S10, códigos de partida, “metrado” o “valorización”. Fusión configurable y evaluación Recall@k/MRR.
- Reranker cross-encoder pequeño opcional; medir mejora de relevancia frente a costo de latencia y RAM.
- La respuesta cita documento, página y sección. Si no hay fragmentos suficientes o los permisos no permiten mostrar la fuente, decir que no se encontró evidencia suficiente.

### Entregables

- Diagrama y scripts de indexación/consulta.
- Set de evaluación con consultas reales anonimizadas y relevancia juzgada.
- Pruebas que impiden recuperar documentos de otra organización.
- Medidas de Recall@k, MRR, latencia, RAM y exactitud de citas.

## 4. Ajuste de estilo mediante LoRA/QLoRA

El ajuste sirve para formato, vocabulario y estilo estable; no sustituye RAG para memorizar manuales o información cambiante.

Para el piloto local en Mac con Apple Silicon, seguir la [guía de MLX-LM para Metrín](PLAN_LORA_METRIN_MLX.md). Es una alternativa experimental a la ruta QLoRA/GPU externa descrita abajo; exige validar compatibilidad del modelo, exportación y capacidad del entorno de inferencia antes de producción.

- Dataset candidato: hasta 2.000 ejemplos sintéticos y reales anonimizados, con licencia/procedencia, eliminación de datos personales y revisión humana.
- Formato Alpaca o ShareGPT versionado, con conjunto de evaluación que no entre al entrenamiento.
- Entrenar QLoRA en GPU externa; no entrenar en el EC2 CPU de 8 GB.
- Probar `r` 8/16, `alpha` 16/32, dropout 0.05, módulos target propios de la arquitectura del modelo; documentar la configuración seleccionada y sus razones.
- Comparar adaptador separado con merge/exportación GGUF. Verificar calidad, licencia, compatibilidad con llama.cpp, tamaño y regresiones antes de publicar.
- Mitigar olvido con ejemplos base, mezcla de datos, evaluación de capacidades previas y rollback por versión.
- Métricas: pérdida/perplexity como señales de entrenamiento, exactitud/citas en tareas reales y rúbrica humana de claridad, tono, seguridad y utilidad. BLEU no es suficiente para evaluar respuestas abiertas.
- Versionar `adapter-vN`, dataset, commit, modelo base, parámetros, evaluación y artefacto; no hacer hot-swap sin chequeo de compatibilidad y rollback.

### Entregables

- Configuración/script PEFT + TRL para entrenamiento en entorno GPU.
- Scripts documentados de merge y conversión GGUF según la versión del modelo.
- Matriz de hiperparámetros, evaluación comparativa y procedimiento de rollback.

## 5. Feedback y ciclo DPO

### Recolección

Capturar 👍/👎, corrección propuesta, rating opcional 1–5 y comentario libre con consentimiento y retención definida. Asociar `message_id`, versión de modelo, fuentes recuperadas, intención y tenant; separar identidad de contenido cuando sea posible.

### Ciclo

```text
Feedback
  -> cola de revisión y anonimización
  -> etiquetado de causa (incorrecta, incompleta, tono, fuente, formato)
  -> pares chosen/rejected aprobados
  -> evaluación offline
  -> entrenamiento DPO/ORPO en GPU aislada
  -> validación de regresión y revisión humana
  -> artefacto versionado
  -> despliegue gradual + rollback
```

- Empezar evaluando DPO frente a ORPO; PPO añade complejidad y requiere una señal de recompensa confiable. Seleccionar por calidad medida, no por popularidad.
- Un voto positivo no prueba que la respuesta sea correcta; nunca generar automáticamente un par de entrenamiento solo por 👍/👎.
- Cadencia mensual como punto de partida; ejecutar entrenamiento solo cuando exista volumen mínimo de ejemplos curados y una mejora estadísticamente útil en evaluación.
- Rúbrica anti-reward-hacking: verificar citas, exactitud contra fuente, permisos, cumplimiento de formato y negativa correcta cuando falta evidencia. Un LLM juez solo puede complementar evaluación humana y tests deterministas.

### Entregables

- Esquema de feedback con consentimiento, retención y anonimización.
- Herramienta de curación y generación revisable de pares.
- Pipeline DPO/ORPO en GPU, evaluación y aprobación antes de desplegar.

## 6. Cortex: orquestación y memoria

Implementar el router en Go o en un servicio Rust con contrato claro, conservando un único responsable de orquestación. No mantener dos grafos de agentes paralelos.

- **Memoria de corto plazo:** últimos turnos pertinentes, recortados por presupuesto de tokens y guardados en PostgreSQL según retención.
- **Memoria de largo plazo:** resúmenes aprobados y hechos persistentes con fuente/consentimiento; embedding en `pgvector` al inicio.
- **Memoria episódica:** errores, correcciones y resoluciones, anonimizadas y sujetas a revisión; no convertir una corrección aislada en verdad global.
- **Router:** clasificador y reglas deterministas seleccionan RAG, acción ERP, soporte o humano. Modelo generativo no ejecuta SQL libre ni elige tenant.
- **Planificador:** divide consultas complejas en subtareas con permiso por cada herramienta, límites de costo y confirmación para acciones con efectos.
- **Política de olvido:** TTL para datos temporales, límites por tenant y resúmenes con trazabilidad al historial fuente; borrar derivados al cumplir solicitud de eliminación.
- **“No sé”:** baja confianza calibrada, ausencia de evidencia recuperada, conflicto de fuentes o salida fuera de política. Registrar pendiente; no afirmar que ya aprendió.
- Cada respuesta conserva `request_id`, versión de modelos, fuentes y herramientas utilizadas, con logs que no expongan contenido sensible por defecto.

LangGraph, CrewAI, AutoGen y LlamaIndex son candidatos solo si resuelven una necesidad concreta. Primero implementar el flujo determinista y observable; añadir un framework de agentes tras comparar dependencias, estado, persistencia, latencia y operación en 8 GB.

### Entregables

- Diagrama de estados y reglas de enrutamiento.
- Esquema de memoria, TTL, anonimización y borrado.
- Trazas de decisiones y herramientas con referencias de fuente.

## 7. Aprendizaje controlado

### Estados

```text
RECIBIDA -> SIN_EVIDENCIA -> PENDIENTE_REVISION -> INVESTIGANDO
  -> ESPERA_HUMANA (si crítica) -> VALIDADA -> INDEXANDO -> DISPONIBLE
  -> EVALUACION_MODELO -> ENTRENAMIENTO (solo por lote aprobado)
```

Estados alternativos: `RECHAZADA`, `FUENTE_NO_CONFIABLE`, `DUPLICADA`, `FALLIDA`. Toda transición deja actor, fecha, motivo y referencias.

### Controles

- Fuentes externas en lista permitida; respetar términos de uso, robots, rate limit y atribución. No usar scraping de login/captcha ni eludir controles de acceso.
- Clasificar autoridad y fecha; normativas exigen fuente oficial vigente. Fuentes débiles no se indexan como instrucciones aprobadas.
- Una persona revisa material crítico, datos normativos, seguridad, contabilidad o acciones financieras antes de publicarlo.
- Primero añadir conocimiento validado al índice RAG. Reentrenar LoRA/DPO solo por lote, umbral de ejemplos aprobado y mejora demostrada.
- Metrín puede decir que registró la consulta. Solo puede confirmar que “ya está disponible” después de validación e indexación exitosas.
- La celebración visual se activa por el estado real `DISPONIBLE`, no por un 👍 ni por crear un pendiente.

### Métricas

Cobertura del dominio evaluado, tiempo hasta respuesta validada, porcentaje de pendientes resueltos, tasa de corrección, citas válidas, reversiones y errores introducidos por actualizaciones.

### Entregables

- Worker de investigación/indexación con fuentes autorizadas y rate limits.
- Bandeja Resuma para revisión, rechazo, edición y aprobación.
- Reindexación idempotente, registro de versiones y auditoría de cambios.

## 8. UI/UX del chat y panel

### Chat

- Layout central de conversación con panel contextual de fuentes y acciones en escritorio; en móvil, fuentes en panel desplegable accesible.
- Widget Metrín anclado a la esquina inferior, que puede asomarse al abrir el chat sin tapar mensajes ni controles.
- Mensajes con Markdown sanitizado: encabezados, listas, tablas, código, enlaces y citas. Nunca renderizar HTML arbitrario.
- Gráficos con Chart.js o Recharts solo cuando la respuesta contenga datos estructurados y el gráfico añada valor.
- Adjuntos con URL autenticada, permisos y vencimiento; no exponer rutas internas de archivos.
- Feedback 👍/👎, “Corregir respuesta”, “Reformular pregunta” y “Hablar con una persona”. Mostrar derivación cuando falte evidencia, haya baja confianza calibrada o el usuario la solicite.
- Estados LED/pose: feliz, pensando, confundido, sorprendido, aprendiendo y celebrando. No usar “aprendiendo” hasta que exista un trabajo activo; “celebrando” cuando la revisión/indexación confirme disponibilidad.
- Motion controla transiciones generales; `apple-motion-lite` queda encapsulado para gestos de la mascota. Cargar bajo demanda, pausar fuera de pantalla y respetar `prefers-reduced-motion`.
- Accesibilidad WCAG AA, contraste medido, teclado, foco visible, etiquetas ARIA y alternativa textual a expresiones animadas.
- Responsive para móvil, tablet y escritorio; área táctil suficiente y mensajes legibles sin scroll horizontal.

### Panel Resuma

- Bandeja de consultas sin evidencia y pendientes de aprendizaje.
- Revisión de fuentes, citas, sensibilidad, duplicados y conflicto de versiones.
- Curación de feedback y pares de preferencia; estado de datasets, adaptadores e índices.
- Control de acceso por tenant/rol y bitácora de decisiones.
- Componentes React/shadcn o Radix son referencia visual/funcional; integrarlos solo como isla React aislada si se justifica, no mezclarlos dentro de componentes Resuma.

### Entregables

- Wireframe descriptivo y estados vacíos/carga/error.
- Tokens compatibles con marca Optimiza 360, definidos una vez y usados por web, chat y panel.
- Prompt visual para Figma AI/v0 con Metrín, sus seis estados y límites de accesibilidad.
- Pruebas manuales con teclado, lector de pantalla y viewports objetivo.

## 9. Contrato de respuesta

La API devuelve JSON validado; el cliente representa `answer_md` tras sanitizar Markdown y enlaces.

```json
{
  "answer_md": "Respuesta con estructura clara y citas visibles.",
  "mood": "happy|thinking|confused|surprised|learning|celebrating",
  "sources": [
    {"doc": "Manual S10", "page": 12, "section": "Presupuestos", "url": null}
  ],
  "attachments": [
    {"type": "pdf|excel|link", "url": "https://...", "expires_at": "..."}
  ],
  "confidence": 0.0,
  "confidence_basis": ["retrieval_score", "classifier", "policy"],
  "follow_up_suggestions": ["..."],
  "learning_status": "none|queued|under_review|indexed",
  "learned": false
}
```

`confidence` se deriva de señales evaluadas y calibradas; no confiar en un número generado por el modelo. `learned` solo es `true` cuando el conocimiento fue validado, indexado y quedó disponible para consultas autorizadas. Omitir adjuntos no aplicables. Fuentes y campos internos no visibles pueden guardarse en auditoría, no en la respuesta pública.

## 10. Prompt de sistema base

```text
Eres Metrín, el asistente de Optimiza 360 para ERP S10 y procesos de construcción.
Respondes en español neutro, de forma cercana, profesional y clara. No usas
gentilicios ni jerga regional.

Usa únicamente contexto autorizado y fuentes recuperadas para afirmar datos
técnicos. Cita documento, página y sección. Si las fuentes no bastan, dilo con
claridad, pide la precisión que falte o deriva a una persona. No inventes citas,
resultados, archivos, permisos ni acciones realizadas.

No afirmes que aprendiste algo por recibir feedback o registrar una pendiente.
Solo confirma que quedó disponible cuando el sistema indique que fue revisado,
validado e indexado.

Organiza las respuestas en pasos cuando expliquen procedimientos y usa tablas
cuando comparen datos. Solicita confirmación antes de acciones con efectos.
Ofrece seguimiento, corrección de pregunta o atención humana cuando corresponda.
```

## 11. Secuencia de entrega

1. Fijar el modelo conversacional y clasificador mediante benchmarks con consultas de evaluación aprobadas.
2. Implementar RAG con citas y aislamiento por tenant antes de habilitar respuestas técnicas productivas.
3. Entregar Cortex determinista, SSE, observabilidad y panel Resuma.
4. Conectar WhatsApp, feedback y aprendizaje humano en modo controlado.
5. Ejecutar LoRA/DPO offline solo después de contar con datos curados, set de regresión y rollback.

No habilitar aprendizaje autónomo, entrenamiento en línea ni ejecución de operaciones del ERP desde el LLM en el MVP.
