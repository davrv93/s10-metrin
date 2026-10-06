# Pendientes — Metrín V2 y modo traza

Fecha: 06-10-2026. Lo hecho está en [`docs/V2-RAG-PROCEDURAL.md`](docs/V2-RAG-PROCEDURAL.md), con resultados en §14;
la foto general del repo, en [`STATUS.md`](STATUS.md); las cifras de rendimiento y sus metas, en
[`docs/RENDIMIENTO-metrin.md`](docs/RENDIMIENTO-metrin.md). Aquí va solo lo que falta. Por cada punto: qué, por qué y dónde.

Estado de partida:

- **Código:** rama `feat/metrin-v2-traza` subida a `origin`, sin fusionar a `main` ni desplegar.
- **Versión por defecto:** V1. La V2 está disponible por petición y desde el selector de la página.
- **Benchmark:** PAS 81,8 % con reranker (60 % sin él) e invención 0 %, sobre 194 casos **sintéticos**.

---

## P1 · Antes de que la V2 la use alguien

1. **Preguntas reales de usuarios.** Todos los casos de `metrin/eval/v2_oro.jsonl` y de los catálogos son
   sintéticos. Además, varias reglas (la regla «no quita» del reranker, los umbrales del recuperador) se diseñaron
   mirando ese mismo benchmark. Hace falta un conjunto aparte con preguntas reales, anonimizadas y etiquetadas, para
   confirmar el 81,8 %.
2. **Medir el reranker donde vaya a correr.** En la Mac, con Metal, el p50 es 0,9 s. En Docker solo con CPU sale
   16,7 s de p50 y 45,6 s de p95 por llamada.
   - Con `RERANK_TIMEOUT_MS=3000`, un servidor sin GPU espera 3 s y vuelve al orden léxico: rinde como «sin
     reranker» (60 %) y más lento.
   - Opciones: GPU, una CPU mucho más rápida, o `RERANK_URL` vacía en ese servidor.
   - Dónde: `docs/V2-RAG-PROCEDURAL.md` §13.
3. **Montar las plantillas en el compose.** Ni `docker-compose.yml` monta `metrin/plantillas` ni el `Dockerfile` la
   copia. Por eso la V2 del compose usa las plantillas mínimas embebidas y pierde:
   - el aviso de los términos definidos por uso;
   - el tamaño de bloque de 3 pasos;
   - el texto en usted.

   El contenedor de prueba (`recrear-4762.sh`) sí las monta en `/srv/plantillas`. Arreglo: un `COPY plantillas` en
   el `Dockerfile` o un volumen en el compose. El código ya busca `/srv/plantillas/respuestas.yml`.
4. **Revisión humana del contenido.**
   - Los 60 procedimientos llevan `revisado_por_humano: false`.
   - Prioridad, porque salen sobre todo de capturas: el asiento manual y el cierre permanente de Contabilidad, y el
     plazo de 7 días para enviar a SUNAT. Ver `kb/procedimientos/COBERTURA.md`.
   - Los 6 términos definidos por uso (metrado, partida, subpresupuesto, título, escenario, CTS) se mantienen con
     aviso, por decisión del usuario del 06-10. Siguen sin revisar por una persona.
5. **Probar el servicio `reranker` en el compose real.** Solo se probó con `docker run` con los mismos argumentos,
   porque Docker no tenía redes libres para levantar otro proyecto. Falta `metrin/descargar-reranker.sh` y
   `docker compose up reranker metrin` en el entorno donde se vaya a usar.
6. **Repositorio público.** `github.com/davrv93/s10-metrin` es público y ahora trae los 60 procedimientos y el
   glosario, que salen de los manuales de S10 (en `main` ya había contenido de manuales). Confirmar que está bien
   publicarlos.

## P2 · Calidad de la V2

7. **Una respuesta indebida nueva con el reranker.** «¿Cómo pago con criptomonedas?» recibe el procedimiento de pago
   con cheque: no inventa pasos, pero contesta otra cosa. La abstención correcta baja de 12/12 a 11/12. Corregirlo
   con una regla general (p. ej. exigir que el término principal de la pregunta aparezca en el candidato), no con un
   caso fijo.
8. **Umbrales provisionales.** `V2_TIPO_UMBRAL` (0,40), `V2_TIPO_MARGEN` (0,08), `V2_EVIDENCIA_MIN` (0) y los del
   recuperador (0,6 / 0,3 / 1,4). Calibrarlos fuera de muestra con el conjunto del punto 1.
9. **Clasificación del tipo de respuesta (77,8 %).** El embebedor estático reconoce el tema pero no el acto («qué
   es», «cómo», «dónde»). MiniLM daría +13 puntos a cambio de 470 MB. Decidir. Ver `kb/catalogos/EVALUACION.md`.
10. **Motor de decisión local en vivo.** La V2 decide con reglas (`DECISION_ENGINE=reglas`). Recomendación de
    `metrin/eval/MODELOS.md`: qwen3.5-4b en la Mac y qwen3.5-2b en el EC2, solo cuando las reglas duden. Falta
    integrarlo y medirlo dentro de la V2. El Jev en GGUF no decide (sale casi uniforme); `jev-style serve` con otro
    backend queda sin probar.
11. **«No me sale».** Devuelve todos los errores frecuentes del procedimiento: no hay relación error ↔ paso en los
    YAML.
12. **La V2 no escribe en `sin_respuesta.jsonl`.** Lo deja en la traza y en `motivo`. Decidir si debe registrar sus
    SIN_EVIDENCIA ahí también (ver el punto 21, privacidad).
13. **`cmd/evalv2` no lee `etapas_v2`.** En la V2, Recall@k, MRR y nDCG salen de los fragmentos citados y no de los
    candidatos de la traza, aunque `V1_VS_V2.md` diga otra cosa. Corregir el arnés o el texto.
14. **Regenerar `metrin/eval/V1_VS_V2.md`** con la configuración por defecto final (reranker en las tres clases). Hoy
    refleja la corrida con el reranker solo en fragmentos. Las cifras finales están en `docs/V2-RAG-PROCEDURAL.md`
    §14 y en los JSON de `metrin/eval/resultados/`.

## P3 · Contenido (`kb/`)

15. **Segunda tanda de procedimientos con fuente** (detalle en `kb/procedimientos/COBERTURA.md`):
    - Gerencia de Proyectos: pedido automático, aprobar pedidos, avances en la rama meta, resultados operativos,
      valorizar subcontrato, MS Project.
    - Compras: orden de servicio y pedido a almacén.
    - Facturación: letras, fondo rotatorio y detracciones.
    - Nóminas: ficha del trabajador y PLAME.
    - Presupuestos: precios por presupuesto y por grupos, cabeceras de reportes, búsqueda en catálogos.
    - Activación Sentinel y Calidad Móvil.
    - Tareo Móvil: activar las 2 de `_reserva/`.
16. **Tareas sin fuente suficiente:**
    - cerrar el proyecto;
    - crear empresas y sucursales;
    - tipo de cambio;
    - cobranza por depósito;
    - importar presupuestos antiguos;
    - ingresos y egresos menores de Almacenes;
    - saldos y stock valorizado;
    - imprimir y enviar la orden de compra.

    Hacen falta fuentes o la cuenta de miembro (`S10_USUARIO`/`S10_CLAVE`, que también falta en `STATUS.md`).
17. **ids de fragmentos repetidos.** En `kb/fragmentos.jsonl`, 200 ids de secciones web se repiten entre manuales
    (1.061 fragmentos afectados): el id corta el nombre del manual a 8 caracteres. La V2 lo esquiva con el par
    (id, manual).
    - Arreglo de raíz: que el indexador use el slug completo, y reindexar.
    - Falta comprobar si el índice vectorial de V1 pierde o mezcla fragmentos por esto.
18. **Glosario:**
    - 9 `procedimientos_relacionados` apuntan a procedimientos que no existen: cotización ×3, recurso ×2, tareo ×2,
      cuadro_comparativo y hoja_del_presupuesto.
    - Los conceptos indican el manual de cada cita en un comentario de línea; conviene pasarlo a una tabla `fuentes`
      como en los procedimientos.
19. **Tutoriales** (`tutoriales/*.md`): no se convirtieron a YAML. Las plantillas de Presupuestos cubren el de «crear
    presupuesto desde cero» por partes.

## P4 · Modo traza y V1

20. **Permiso por rol.** Metrín no tiene sesiones; la única barrera de la traza es `METRIN_TRAZA`. En un entorno
    abierto, cualquiera que llegue a la página puede pedir la traza.
21. **Privacidad del registro de fallos.** `registrarFallo` guarda la pregunta completa en `sin_respuesta.jsonl`.
    Decidir si se recorta, como hace la traza (120 caracteres).
22. **Reglas de oportunidades de la traza.** La regla «bajo el umbral» del clasificador salta en casi todas las
    preguntas de trabajo (similitud 0,31–0,36): mete ruido. Revisarla junto con el umbral.
23. **`RAG_MAX_DISTANCIA=0,80` es laxo en V1.** Una pregunta ajena (el horario de un comedor) encontró «contexto» a
    0,534.
24. **El generador de V1 (`qwen2.5-coder:7b`) no sirve.** Con MLX encendido, V1 inventa en el 17 % de las respuestas
    y tarda de 7 a 28 s. Opciones: el LoRA de MLX, plantillas, o mover el tráfico a la V2.
25. **Página:**
    - **Fallo en el teléfono** (anterior a la traza): con una galería de fotos, el chat se ensancha a 784 px y corta
      la cabecera y la caja de texto. Hoy solo está arreglado con la traza activa. El arreglo global es una línea:
      `.main{grid-template-columns:minmax(0,1fr)}`.
    - Algunos textos de V1 tutean («¿Qué quieres saber?»), mientras la V2 trata de usted.
    - IBM Plex Mono no se carga; se usa `ui-monospace`.
26. **Fuentes «tercero-sin-verificar».** Los 3 fragmentos que usa V1 para registrar un presupuesto vienen de los PDF
    importados a mano. Los procedimientos los usan solo junto a la sección oficial.

## P5 · Repositorio, despliegue y la Mac

27. **Fusionar `feat/metrin-v2-traza` a `main`** (o abrir un PR) cuando se decida. No hay CI.
28. **El trabajo del 01-10 sigue sin commit** en `metrin/`:
    - archivos: `rag.go`, `main.go`, `config.go`, `go.mod`, `go.sum`, `.env.example`, `cmd_reporte.go`,
      `reportes/`, `pipeline/`;
    - la función `evaluarChat`, que mezcla los dos trabajos;
    - de otras sesiones: `trazas/cargar.py`, el servicio landing-cd del compose, `.cortex/` y `data/`.

    Es de otra sesión: que lo commitee quien lo hizo.
29. **Contenedor de siempre (4760).** Sigue con la imagen de V1, sin traza ni V2. Recrearlo con la imagen nueva
    cuando se decida.
30. **Limpieza:**
    - 4 JSON intermedios en `metrin/eval/resultados/` (~6 MB cada uno) sin commit: borrarlos o subirlos.
    - En `~/.cache/metrin-modelos/` hay GGUF de prueba que se pueden borrar (Qwen3-Reranker 639 MB, bge-m3 Q4_0
      422 MB, bge-base 304 MB) y 8,6 GB de modelos del benchmark.
31. **Servicios que quedaron corriendo en la Mac:** MLX en `:8080` (`pkill -f mlx_lm.server`) y el reranker en
    `:8091` (`kill $(cat ~/.cache/metrin-modelos/llama-rerank.pid)`). Apagarlos si no se usan.
