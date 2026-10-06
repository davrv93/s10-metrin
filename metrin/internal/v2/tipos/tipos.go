// Package tipos es el contrato compartido de la V2 de Metrín (RAG procedural multimodal).
//
// Lo usan el núcleo (internal/v2), el conocimiento procedural (internal/v2/conocimiento) y la
// evaluación (cmd/evalv2). Especificación: docs/V2-RAG-PROCEDURAL.md. Regla: los modelos proponen,
// el motor de decisión decide, el código controla, la evidencia demuestra, el LLM redacta y el
// renderizador presenta. Ningún tipo de aquí lleva texto inventado por un modelo sin validar.
//
// Cambiar este archivo es cambiar el contrato: se avisa a los demás paquetes y se ajustan sus pruebas.
package tipos

import "context"

// ---------------------------------------------------------------------------------------------
// Configuración

// Version del agente para un turno.
type Version string

const (
	V1 Version = "v1"
	V2 Version = "v2"
)

// Limites del turno (variables de entorno; valores por defecto del megaprompt).
type Limites struct {
	MaxPasos            int // MAX_AGENT_STEPS, 5
	MaxDecisiones       int // MAX_DECISION_CALLS, 4
	MaxBusquedas        int // MAX_SEARCH_ITERATIONS, 2
	MaxHerramientas     int // MAX_TOOL_CALLS, 5
	MaxGeneraciones     int // MAX_GENERATIONS, 1
	MaxRegeneraciones   int // MAX_REGENERATIONS, 1
	TimeoutDecisionMs   int // DECISION_TIMEOUT_MS, 1500
	TimeoutGeneracionMs int // GENERATION_TIMEOUT_MS, 5000
}

// ---------------------------------------------------------------------------------------------
// Tipo de respuesta (catálogo kb/catalogos/tipo_respuesta.yml)

type TipoRespuesta string

const (
	Concepto      TipoRespuesta = "CONCEPT"
	Procedimiento TipoRespuesta = "PROCEDURE"
	Navegacion    TipoRespuesta = "NAVIGATION"
	Problema      TipoRespuesta = "TROUBLESHOOTING"
	Configuracion TipoRespuesta = "CONFIGURATION"
	Comparacion   TipoRespuesta = "COMPARISON"
	Desconocido   TipoRespuesta = "UNKNOWN"
	// Fuera del manual: los resuelve la ruta conversacional de V1 (saludo, despedida, fuera de dominio).
	Social TipoRespuesta = "SOCIAL"
)

// ---------------------------------------------------------------------------------------------
// Estado estructurado (context builder)

// Dato es un valor con su origen. Un dato inferido NUNCA pasa a Hechos.
type Dato struct {
	Valor     string  `json:"valor"`
	Origen    string  `json:"origen"`    // usuario | hilo | memoria | clasificador | decision
	Confianza float64 `json:"confianza"` // 1 para lo que dijo el usuario
}

// Memoria del procedimiento en curso: viaja en el hilo (lo guarda la página y lo devuelve). Vacía («{}»)
// significa «sin procedimiento en curso» y también borra uno anterior (al abandonarlo): por eso /ask la manda
// siempre en V2, aunque esté vacía, y la página y el arnés la reenvían tal cual.
type Memoria struct {
	ProcedimientoID  string `json:"procedimiento_id,omitempty"`
	PasoActual       int    `json:"paso_actual,omitempty"`
	PasosCompletados []int  `json:"pasos_completados,omitempty"`
	// Tema suspendido por un cambio de intención: se retoma una sola vez.
	Suspendido *Memoria `json:"suspendido,omitempty"`
	Retomado   bool     `json:"retomado,omitempty"`

	// Aditivo (núcleo, 06-10-2026): a qué puede referirse el turno siguiente aunque no haya un procedimiento en
	// curso. Concepto: id del glosario que la respuesta explicó (uno solo; no en COMPARISON). Ofrecido: id del
	// procedimiento que la respuesta ofreció enseñar (plan.opciones con UNA opción: el procedimiento relacionado de
	// un concepto, o la única opción de una aclaración); se borra si la oferta que se ve es retomar un suspendido.
	// Los escribe v2.MemoriaDePlan; cada respuesta los reescribe (una que no los trae los deja vacíos). Los usan,
	// sin LLM, v2.InterpretarMemoria («sí», «sí, enséñame», «¿y luego?» → lo ofrecido desde el paso 1) y
	// v2.ResolverReferencia («¿y cómo la hago?», «¿y eso dónde está?» → directo a ese procedimiento o concepto).
	Concepto string `json:"concepto,omitempty"`
	Ofrecido string `json:"ofrecido,omitempty"`
}

// Consulta normalizada (reescritura): la original no se pierde nunca.
type Consulta struct {
	Original    string   `json:"original"`
	Normalizada string   `json:"normalizada"`
	Entidades   []string `json:"entidades,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	Modulo      string   `json:"modulo,omitempty"`
}

// Estado es lo que reciben el motor de decisión y el planificador. Compacto: nada de historial entero.
type Estado struct {
	Pregunta      string          `json:"pregunta"`
	Consulta      Consulta        `json:"consulta"`
	Hilo          []Turno         `json:"hilo,omitempty"` // los últimos turnos, recortados
	Tipo          TipoRespuesta   `json:"tipo"`
	TipoConfianza float64         `json:"tipo_confianza"`
	Hechos        map[string]Dato `json:"hechos"`
	Inferencias   map[string]Dato `json:"inferencias"`
	Desconocidos  []string        `json:"desconocidos"`
	Memoria       Memoria         `json:"memoria"`
}

type Turno struct {
	Rol   string `json:"rol"`
	Texto string `json:"texto"`
}

// ---------------------------------------------------------------------------------------------
// Decisiones (motor tipo Jev: opciones cerradas, nunca párrafos)

type Decision struct {
	Nombre    string  `json:"decision"` // response_type | procedure | evidence_sufficiency | next_action | needs_more_search | response_quality | missing_information | retrieval_strategy
	Eleccion  string  `json:"choice"`   // una de Opciones
	Confianza float64 `json:"confidence"`
	Fuente    string  `json:"fuente"` // reglas | local | remota | simulada
}

// Pregunta al motor: qué decidir y entre qué opciones.
type PreguntaDecision struct {
	Nombre   string   `json:"decision"`
	Opciones []string `json:"opciones"`
}

// DecisionEngine decide entre opciones cerradas. Una salida fuera de las opciones es un error.
type DecisionEngine interface {
	Nombre() string
	Decidir(ctx context.Context, e Estado, preguntas []PreguntaDecision) ([]Decision, error)
}

// ---------------------------------------------------------------------------------------------
// Conocimiento procedural (kb/procedimientos/**/*.yml, kb/conceptos/*.yml)

type Fuente struct {
	ID      string `json:"id" yaml:"id"`
	Manual  string `json:"manual" yaml:"manual"`
	Seccion string `json:"seccion,omitempty" yaml:"seccion"`
	Pagina  int    `json:"pagina,omitempty" yaml:"pagina"`
}

type Foto struct {
	ID      string `json:"id" yaml:"id"`
	Ruta    string `json:"ruta" yaml:"ruta_o_url"`
	Pagina  int    `json:"pagina,omitempty" yaml:"pagina"`
	Caption string `json:"caption,omitempty" yaml:"caption"`
	OCR     string `json:"ocr,omitempty" yaml:"ocr"`
}

type ConFuente struct {
	Texto  string   `json:"texto" yaml:"texto"`
	Fuente []string `json:"fuente" yaml:"fuente"`
}

type ErrorFrecuente struct {
	Sintoma  string   `json:"sintoma" yaml:"sintoma"`
	Solucion string   `json:"solucion" yaml:"solucion"`
	Fuente   []string `json:"fuente" yaml:"fuente"`
}

type Paso struct {
	ID        string   `json:"id" yaml:"id"` // <procedimiento>#<n>
	N         int      `json:"n" yaml:"n"`
	Accion    string   `json:"accion" yaml:"accion"`
	Donde     string   `json:"donde,omitempty" yaml:"donde"`
	Resultado string   `json:"resultado,omitempty" yaml:"resultado"`
	Pagina    int      `json:"pagina,omitempty" yaml:"pagina"`
	Fuente    []string `json:"fuente" yaml:"fuente"`
	Captura   string   `json:"captura,omitempty" yaml:"captura"` // compatibilidad con la v1 del esquema
	Fotos     []Foto   `json:"fotos,omitempty" yaml:"fotos"`
	Sub       []Paso   `json:"sub,omitempty" yaml:"sub"`
}

// EntidadesDef: en los YAML reales `entidades` es un mapa {modulo, pantallas, objetos} (06-10-2026).
type EntidadesDef struct {
	Modulo    string   `json:"modulo,omitempty" yaml:"modulo"`
	Pantallas []string `json:"pantallas,omitempty" yaml:"pantallas"`
	Objetos   []string `json:"objetos,omitempty" yaml:"objetos"`
}

type ProcedimientoDef struct {
	ID                string           `json:"id" yaml:"id"`
	Version           int              `json:"version" yaml:"version"`
	Modulo            string           `json:"modulo" yaml:"modulo"`
	Titulo            string           `json:"titulo" yaml:"titulo"`
	Objetivo          string           `json:"objetivo" yaml:"objetivo"`
	Nivel             string           `json:"nivel" yaml:"nivel"`
	Preguntas         []string         `json:"preguntas" yaml:"preguntas"`
	Aliases           []string         `json:"aliases,omitempty" yaml:"aliases"`
	Entidades         EntidadesDef     `json:"entidades,omitempty" yaml:"entidades"`
	Prerrequisitos    []ConFuente      `json:"prerrequisitos" yaml:"prerrequisitos"`
	Pasos             []Paso           `json:"pasos" yaml:"pasos"`
	Verificacion      []ConFuente      `json:"verificacion" yaml:"verificacion"`
	ErroresFrecuentes []ErrorFrecuente `json:"errores_frecuentes" yaml:"errores_frecuentes"`
	Relacionados      []string         `json:"relacionados" yaml:"relacionados"`
	Fuentes           []Fuente         `json:"fuentes" yaml:"fuentes"`
}

type ConceptoDef struct {
	ID                         string   `json:"id" yaml:"id"`
	Termino                    string   `json:"termino" yaml:"termino"`
	Sinonimos                  []string `json:"sinonimos,omitempty" yaml:"sinonimos"`
	Definicion                 string   `json:"definicion" yaml:"definicion"`
	EnS10                      string   `json:"en_s10,omitempty" yaml:"en_s10"`
	Ejemplo                    string   `json:"ejemplo,omitempty" yaml:"ejemplo"`
	ProcedimientosRelacionados []string `json:"procedimientos_relacionados,omitempty" yaml:"procedimientos_relacionados"`
	Fuente                     []string `json:"fuente" yaml:"fuente"`

	// Aditivo (conocimiento, 06-10-2026): Notas = `notas` del YAML tal cual (nota editorial: «definido por uso»,
	// «no confundir con…»); la página no la pinta. Aviso = la línea que SÍ se muestra a la persona cuando la
	// definición está armada por uso, sin definición formal en las fuentes (kb/conceptos/SIN_FUENTE.md, decisión del
	// usuario «mantener con aviso»); sale de la parte «aviso» de la plantilla CONCEPTO. Vacío en los demás términos.
	Notas string `json:"notas,omitempty" yaml:"notas"`
	Aviso string `json:"aviso,omitempty" yaml:"-"`
}

// ---------------------------------------------------------------------------------------------
// Recuperación (búsqueda híbrida + reranker; implementación en internal/busqueda e internal/rerank)

type Candidato struct {
	ID      string            `json:"id"`
	Clase   string            `json:"clase"` // procedimiento | concepto | fragmento
	Puntaje float64           `json:"puntaje"`
	Vector  float64           `json:"vector,omitempty"`
	Lexico  float64           `json:"lexico,omitempty"`
	Rerank  *float64          `json:"rerank,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// Recuperador busca en un índice (procedimientos, conceptos o fragmentos). Debe degradar solo:
// sin vector → léxica; sin reranker → puntaje híbrido. Nunca falla en silencio: devuelve Degradado.
type Recuperador interface {
	Buscar(ctx context.Context, c Consulta, clase string, k int) (res []Candidato, degradado []string, err error)
}

// ---------------------------------------------------------------------------------------------
// Plan de respuesta (contrato interno; lo pinta la página, no el LLM)

type FuentePlan struct {
	Manual  string `json:"manual"`
	Seccion string `json:"seccion,omitempty"` // UNA sección (una entrada por sección del manual)
	Paginas []int  `json:"paginas,omitempty"`
	// Aditivo (conocimiento, 06-10-2026): los pares (id, manual) citados que resume esta entrada, para
	// resolverla sin interpretar el texto («Manual › sección»).
	Citas []Fuente `json:"citas,omitempty"`
}

type PasoPlan struct {
	N      int      `json:"n"`
	ID     string   `json:"id"`
	Texto  string   `json:"texto"`
	Fotos  []Foto   `json:"fotos,omitempty"` // solo las que la fuente asocia a ESTE paso
	Fuente []string `json:"fuente"`
	Pagina int      `json:"pagina,omitempty"`
}

type Afirmacion struct {
	Texto     string `json:"texto"`
	Fragmento string `json:"fragmento"`
	Manual    string `json:"manual,omitempty"`
	Pagina    int    `json:"pagina,omitempty"`
	Paso      int    `json:"paso,omitempty"`
}

type Plan struct {
	Version        int              `json:"version"` // 2
	Tipo           TipoRespuesta    `json:"tipo"`
	Procedimiento  *Ref             `json:"procedimiento,omitempty"`
	Concepto       *ConceptoDef     `json:"concepto,omitempty"`
	Intro          string           `json:"intro,omitempty"`
	Prerrequisitos []ConFuente      `json:"prerrequisitos,omitempty"`
	Pasos          []PasoPlan       `json:"pasos,omitempty"`
	PasosMostrados []int            `json:"pasos_mostrados,omitempty"` // entrega por partes
	Verificacion   []ConFuente      `json:"verificacion,omitempty"`
	Errores        []ErrorFrecuente `json:"errores,omitempty"`
	Siguiente      *Siguiente       `json:"siguiente,omitempty"`
	Fuentes        []FuentePlan     `json:"fuentes"`
	Afirmaciones   []Afirmacion     `json:"afirmaciones,omitempty"`
	SinEvidencia   bool             `json:"sin_evidencia,omitempty"`
	Plantilla      string           `json:"plantilla,omitempty"` // plantilla de metrin/plantillas/respuestas.yml usada

	// Aditivo (conocimiento, 06-10-2026): COMPARISON trae los dos conceptos (Concepto = el primero) y UNKNOWN
	// los 2–3 procedimientos candidatos de la aclaración (ACLARAR_TAREA), para que la página los ofrezca.
	Conceptos []ConceptoDef `json:"conceptos,omitempty"`
	Opciones  []Ref         `json:"opciones,omitempty"`
}

type Ref struct {
	ID        string  `json:"id"`
	Titulo    string  `json:"titulo"`
	Confianza float64 `json:"confianza"`
}

type Siguiente struct {
	Plantilla string `json:"plantilla"` // PREGUNTA_AVANCE | RETOMA | ACLARAR_TAREA …
	Paso      int    `json:"paso,omitempty"`
	Texto     string `json:"texto"`
}

// ---------------------------------------------------------------------------------------------
// Generación y control de calidad

// GenerationEngine convierte el plan en texto. Puede reformular, nunca cambiar pasos, fotos,
// negritas ni citas. Las plantillas (sin modelo) son la implementación por defecto.
type GenerationEngine interface {
	Nombre() string
	Redactar(ctx context.Context, p Plan) (texto string, err error)
}

type ResultadoCalidad struct {
	Paso      bool     `json:"passed"`
	Puntaje   float64  `json:"score"`
	Problemas []string `json:"issues"`
}

// QualityGate valida plan y texto por código, contra la evidencia. Sin modelo.
type QualityGate interface {
	Evaluar(p Plan, texto string, e Estado) ResultadoCalidad
}
