// Package conocimiento es el conocimiento procedural de la V2 de Metrín (docs/V2-RAG-PROCEDURAL.md).
//
// Carga al arrancar los procedimientos (kb/procedimientos/**/*.yml), el glosario (kb/conceptos/*.yml),
// las plantillas (metrin/plantillas/respuestas.yml) y el índice de fragmentos (kb/fragmentos*.jsonl), y
// ofrece al núcleo (internal/v2) cinco piezas que cumplen el contrato de internal/v2/tipos:
//
//   - Base: lo cargado, con los errores de carga (archivo y línea). Un YAML inválido se salta y se
//     informa; nunca tumba el servicio.
//   - Recuperador (tipos.Recuperador): BM25 propio sobre procedimientos, conceptos, errores frecuentes y
//     fragmentos con pasos; vector y reranker opcionales (Vectorizador, Reordenador). Sin vector →
//     léxica; sin reranker → puntaje híbrido; siempre informado en «degradado».
//   - Constructor: evidencia y plan (tipos.Plan) por tipo de respuesta.
//   - Plantillas y Local (tipos.GenerationEngine): texto desde el plan sin modelo, o reformulado por un
//     LLM local que no puede tocar pasos, fotos, negritas ni citas.
//   - Gate (tipos.QualityGate): control por código, sin modelo.
//
// Citas. Un id de kb/fragmentos*.jsonl NO es único (las secciones web truncan el slug del manual:
// ESQUEMA.md, «ids ambiguos»). Toda cita se resuelve con el par (id, manual): en los procedimientos, con
// su tabla «fuentes»; en los conceptos, con su tabla «fuentes» si la traen, con {id, manual} en línea o,
// como hoy, con el manual escrito en el comentario de la línea («# Manual de X › sección»). Nunca solo
// con el id, salvo que el id exista en un único manual (y entonces el par queda determinado).
//
// Fotos. Una foto entra en un paso del plan solo si el YAML se la asigna a ese paso (o, en un
// procedimiento implícito, si la sección fuente la pone en ese pasos[k]); un paso sin foto en la fuente
// queda sin foto. Las rutas son las de la fuente («imagenes/<manual>/<hash>.png») y se sirven en
// GET /fotos/<ruta>, como en V1 (URLFoto). Por defecto el texto no las lleva: la página las pinta desde
// plan.pasos[].fotos; MotorPlantillas.FotosEnTexto las pone como ![…](fotos/…) debajo de su paso.
//
// Contrato con el núcleo (internal/v2/interfaces.go): Base cumple Aliaser y Catalogo; Constructor cumple
// Constructor y Continuador (contrato_test.go lo comprueba al compilar).
//
// YAML: gopkg.in/yaml.v3 (errores con línea). kb/procedimientos/_reserva/ (toda carpeta «_…») no se
// carga salvo Opciones.IncluirReserva.
//
// Búsqueda. internal/busqueda (BM25 + vector + RRF) e internal/rerank los construyen otros agentes y aún
// no tienen una API estable; este paquete usa un BM25 mínimo propio detrás de tipos.Recuperador y acepta
// el vector y el reranker por interfaz, para enchufarlos sin cambiar a quien lo llama.
package conocimiento
