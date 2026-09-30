# Metrín: estado de implementación

## Punto alcanzado

Implementada y levantada una primera versión vertical del chatbot sobre el
servicio existente `rag-go`, con la base S10 importada y una interfaz web
operativa en `http://localhost:4760/`. La pantalla se rediseñó para integrar a
Metrín como compañero animado del chat, con estados visuales de espera, respuesta
y falta de evidencia.

## Hecho

- Interfaz adaptable de chat con identidad de Metrín, consultas sugeridas,
  indicador de estado y panel de referencias.
- Personaje CSS original con casco O360, visor LED, chaleco reflectante, botas,
  movimiento idle, reacción al pensar y celebración al responder; respeta
  `prefers-reduced-motion`.
- Logo oficial enlazado desde Optimiza 360 sobre una placa de contraste.
- Animación de scroll: Metrín salta desde el hero, se encoge y acompaña el
  compositor; ciclo de construcción, fórmula de costos sin cifras inventadas,
  hora local y descanso tras inactividad.
- API same-origin `POST /ask` y estado `GET /health`.
- Importador incremental de `s10-conocimiento/kb/fragmentos.jsonl`, incluyendo
  título, página y URL en las fuentes citadas.
- Índice cargado: 537 documentos, 1.197 fragmentos.
- Respuesta basada en referencias y abstención cuando la evidencia no alcanza.
  La consulta de verificación devolvió HTTP 200; búsqueda en 1 ms y 5 fuentes.
- Build Go completado correctamente.
- HTML actualizado servido en `/`; sintaxis del JavaScript embebido comprobada.

## Etapa pendiente

Esto aún no es el stack final Docker/PostgreSQL/Redis ni incluye streaming,
autenticación, multitenencia o integración con Evolution Go. El índice de esta
sesión vive en `/tmp/metrin-rag-data`, así que es temporal y se pierde al
reiniciar el entorno.

## Siguiente paso exacto

1. Afinar el umbral de evidencia y probar preguntas específicas de S10 para
   separar abstenciones correctas de falsos negativos.
2. Instalar/configurar el runtime conversacional y su modelo en el entorno
   destino; comprobar salud y latencia desde `rag-go`.
3. Persistir el índice en un volumen Docker y luego conectar autenticación y
   servicios del stack en iteraciones separadas.

## Ejecución local de esta sesión

El binario está en `/tmp/metrin-rag-build`; el servidor se lanzó con:

```sh
RAG_DATOS=/tmp/metrin-rag-data RAG_TIMEOUT=12 /tmp/metrin-rag-build serve --addr 127.0.0.1:4760
```

Para reconstruir el índice, desde `rag-go/`:

```sh
RAG_DATOS=/tmp/metrin-rag-data /tmp/metrin-rag-build index-kb ../s10-conocimiento/kb/fragmentos.jsonl
```
