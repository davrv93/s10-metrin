package v2

// Interfaces del núcleo V2 que implementa el paquete de conocimiento (internal/v2/conocimiento) además de las de
// internal/v2/tipos (Recuperador, GenerationEngine, QualityGate, DecisionEngine). Contrato estable: cambiarlas es
// cambiar el contrato entre agentes y se avisa. Especificación: docs/V2-RAG-PROCEDURAL.md §3–§6.
//
// Regla: los modelos proponen, el motor de decisión decide, el código controla, la evidencia demuestra, el LLM
// redacta y el renderizador presenta. Ninguna implementación de estas interfaces inventa texto: todo sale de kb/.

import (
	"context"

	"rag-go/internal/v2/tipos"
)

// Aliaser detecta en una pregunta los términos del ERP (kb/catalogos, kb/conceptos, aliases de los procedimientos).
//
// Analizar NO cambia el texto: devuelve
//   - entidades: los términos exactos del ERP que aparecen en el texto, escritos como en S10 («Metrado», «APU», «F7»);
//   - aliases: sinónimos y formas alternativas para expandir la búsqueda (nunca sustituyen a la original);
//   - modulo: el módulo de S10 al que apuntan, o "" si no se sabe.
//
// SinAlias es la implementación vacía (por defecto, mientras no haya catálogo cargado).
type Aliaser interface {
	Analizar(texto string) (entidades, aliases []string, modulo string)
}

// Constructor arma el plan de respuesta (tipos.Plan, §5) a partir de la evidencia recuperada:
//   - PROCEDURE / NAVIGATION / CONFIGURATION / TROUBLESHOOTING: procedimiento con pasos, las fotos de CADA paso
//     (solo las que la fuente asocia a ese paso), prerrequisitos, verificación y fuentes;
//   - CONCEPT / COMPARISON: definición(es) del glosario con su fuente;
//   - sin evidencia que respalde nada: Plan{SinEvidencia: true}.
//
// e.Tipo es el tipo de respuesta decidido y e.Memoria el procedimiento en curso. Los candidatos vienen ordenados
// de mejor a peor. Un error se trata como «sin plan»: el núcleo responde con la respuesta segura y lo anota.
type Constructor interface {
	Construir(ctx context.Context, e tipos.Estado, cands []tipos.Candidato) (tipos.Plan, error)
}

// Catalogo da acceso directo, por id, a lo que ya está cargado de kb/. La memoria del procedimiento lo usa para
// «¿y luego?», «listo» o «no me sale» SIN buscar otra vez (el paso siguiente sale del mismo procedimiento).
type Catalogo interface {
	Procedimiento(id string) (tipos.ProcedimientoDef, bool)
	Concepto(id string) (tipos.ConceptoDef, bool)
}

// Continuador (OPCIONAL; añadido el 06-10-2026, aditivo): si el Constructor también lo implementa, el núcleo le pide
// la parte siguiente del procedimiento en curso («¿y luego?», «listo», retomar) en lugar de armarla él con el
// Catalogo. mem.PasoActual es el último paso YA mostrado; el plan devuelto trae TODOS los pasos en Pasos y la parte
// a mostrar en PasosMostrados (desde mem.PasoActual+1). Pasado el último paso: plantilla PROCEDIMIENTO_FIN.
type Continuador interface {
	Continuar(mem tipos.Memoria) (tipos.Plan, error)
}

// SinAlias es el Aliaser vacío: no detecta nada y la consulta queda solo con sus términos exactos.
type SinAlias struct{}

func (SinAlias) Analizar(string) ([]string, []string, string) { return nil, nil, "" }
