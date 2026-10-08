package v2

// Router de casos (docs/TOC-RUTEO-METRIN.md, internal/router): un clasificador entrenado elige el procedimiento con
// el árbol «módulo primero → abstenerse → empate = pregunta». Actúa sobre los candidatos que ya trajo la búsqueda,
// antes del constructor:
//
//   - elegir: solo queda el procedimiento elegido (se añade si la búsqueda no lo trajo), con puntaje ≥ el umbral del
//     constructor. En TROUBLESHOOTING quedan además solo los errores de ese procedimiento, si hay alguno.
//   - aclarar (caso o módulo): las 2–3 opciones empatan a propósito (mismo puntaje y mismo léxico) para que el
//     constructor haga la aclaración de siempre («dos procedimientos empatan»).
//   - delegar: no cambia nada; la V2 decide como antes (R2: abstenerse antes que equivocarse).
//
// Sin router (Agente.Router nil) la V2 es byte a byte la de antes.

import (
	"strconv"
	"strings"

	"rag-go/internal/router"
	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

// RouterCasos: internal/router.Router (o un doble en las pruebas).
type RouterCasos interface {
	Decidir(texto string) router.Decision
}

// Puntajes que el router da a sus candidatos. PuntajeRouterElegido ≥ UmbralProcedimiento (0,6) del constructor para
// que acepte; PuntajeRouterEmpate también lo pasa, pero igual en todas las opciones, así el constructor ve un empate
// y pregunta en vez de elegir.
const (
	PuntajeRouterElegido = 0.9
	PuntajeRouterEmpate  = 0.65
	origenRouter         = "router"
)

// usaRouter: los tipos cuya respuesta es un procedimiento.
func usaRouter(tipo tipos.TipoRespuesta) bool {
	switch tipo {
	case tipos.Procedimiento, tipos.Navegacion, tipos.Configuracion, tipos.Problema:
		return true
	}
	return false
}

// rutear aplica la decisión del router a los candidatos. usado=false: el router no intervino (o no hay router).
func (t *turno) rutear(tipo tipos.TipoRespuesta, cands []tipos.Candidato) ([]tipos.Candidato, bool) {
	if t.a.Router == nil || !usaRouter(tipo) {
		return cands, false
	}
	d := t.decisionRouter()
	ids := make([]string, 0, len(d.Candidatos))
	for _, c := range d.Candidatos {
		ids = append(ids, c.ID)
	}
	if t.v != nil {
		top := make([]map[string]any, 0, len(d.Top))
		for _, c := range d.Top {
			top = append(top, map[string]any{"id": c.ID, "p": traza.Redondear(c.Probabilidad)})
		}
		t.v.Dato(traza.EtapaV2PlanConsulta, "router", traza.Datos{"accion": d.Accion, "candidatos": ids, "top": top,
			"p_ninguno": traza.Redondear(d.Ninguno), "modulo": d.Modulo, "p_modulo": traza.Redondear(d.PModulo)})
	}
	var out []tipos.Candidato
	switch d.Accion {
	case router.Elegir:
		id := ids[0]
		if _, ok := t.procedimiento(id); !ok {
			t.razonRouter("el router eligió " + id + ", que no está en el catálogo: decide la V2")
			return cands, false
		}
		hayErrores := false
		for _, c := range cands {
			if c.Clase == "error" && c.Meta["procedimiento"] == id {
				hayErrores = true
			}
		}
		for _, c := range cands {
			switch {
			case c.Clase == "procedimiento":
				continue
			case c.Clase == "error" && hayErrores && c.Meta["procedimiento"] != id:
				continue
			}
			out = append(out, c)
		}
		out = append([]tipos.Candidato{candidatoRouter(id, PuntajeRouterElegido, d.Candidatos[0].Probabilidad, "elegir")}, out...)
		t.anotarRouter(id, d.Candidatos[0].Probabilidad)
		t.razonRouter("router: " + id + " (p " + strconv.FormatFloat(d.Candidatos[0].Probabilidad, 'f', 2, 64) + ")")
	case router.AclararCaso, router.AclararModulo:
		if tipo == tipos.Problema {
			// Los errores no tienen aclaración por empate de procedimientos: decide la V2.
			return cands, false
		}
		ops := ids
		if d.Accion == router.AclararModulo {
			// Los 3 procedimientos más probables, sean del módulo que sean (medido el 07-10-2026: ofrecer el mejor de cada
			// módulo casi nunca incluía el correcto).
			ops = nil
			for _, c := range d.Top {
				ops = append(ops, c.ID)
			}
		}
		var validas []string
		for _, id := range ops {
			if _, ok := t.procedimiento(id); ok {
				validas = append(validas, id)
			}
		}
		if len(validas) < 2 {
			return cands, false
		}
		for _, id := range validas {
			out = append(out, candidatoRouter(id, PuntajeRouterEmpate, 0, d.Accion))
		}
		for _, c := range cands {
			if c.Clase != "procedimiento" {
				out = append(out, c)
			}
		}
		t.razonRouter("router duda (" + d.Accion + "): " + strings.Join(validas, ", "))
	default:
		return cands, false
	}
	return out, true
}

// decisionRouter: la decisión del router para la pregunta del turno, calculada una sola vez.
func (t *turno) decisionRouter() router.Decision {
	if t.decRouter == nil {
		d := t.a.Router.Decidir(t.pregunta)
		t.decRouter = &d
	}
	return *t.decRouter
}

// MinPalabrasTipoRouter: una pregunta con menos palabras es un tema suelto («kardex», «sunat»), no una tarea: se
// sigue pidiendo aclaración aunque el router la asocie a un procedimiento (regla 6 de metrin/eval/construir_v2_oro.py).
const MinPalabrasTipoRouter = 3

// tipoPorRouter: si el clasificador de tipo no sabe (UNKNOWN) pero el router ELIGE un procedimiento, o duda entre 2–3
// casos del MISMO módulo (aclarar_caso), para una pregunta de al menos MinPalabrasTipoRouter palabras, la pregunta se
// trata como PROCEDURE: o se responde, o la aclaración ofrece esos casos («¿te refieres a A o a B?») en vez de la
// aclaración genérica sin opciones. Medido el 07-10-2026: 13 de 125 preguntas con caso tenían el procedimiento bien
// elegido por el router y terminaban en aclaración por el tipo; en las preguntas reales, 2 de 6 terminaban en la
// aclaración genérica con el caso correcto primero entre las opciones del router (docs/TOC-RUTEO-METRIN.md §8).
func (t *turno) tipoPorRouter(tipo tipos.TipoRespuesta) tipos.TipoRespuesta {
	if t.a.Router == nil || len(strings.Fields(t.pregunta)) < MinPalabrasTipoRouter {
		return tipo
	}
	d := t.decisionRouter()
	if d.Accion != router.Elegir && d.Accion != router.AclararCaso {
		return tipo
	}
	t.razonRouter("tipo UNKNOWN, pero el router " + d.Accion + " (" + d.Candidatos[0].ID + "…): se trata como PROCEDURE")
	// Como aplicarReferencia (referencia.go): el tipo del estado es el que ve el constructor, y la tarea ya no es un
	// desconocido.
	p := d.Candidatos[0].Probabilidad
	t.metodo = "router"
	t.estado.Tipo, t.estado.TipoConfianza = tipos.Procedimiento, p
	if t.estado.Inferencias == nil {
		t.estado.Inferencias = map[string]tipos.Dato{}
	}
	t.estado.Inferencias["tipo"] = tipos.Dato{Valor: string(tipos.Procedimiento), Origen: origenRouter, Confianza: p}
	var desc []string
	for _, x := range t.estado.Desconocidos {
		if x != "tarea" {
			desc = append(desc, x)
		}
	}
	t.estado.Desconocidos = desc
	if t.v != nil {
		t.v.Dato(traza.EtapaV2TipoRespuesta, "tipo", string(tipos.Procedimiento))
		t.v.Dato(traza.EtapaV2TipoRespuesta, "metodo", "router")
		t.v.Dato(traza.EtapaV2TipoRespuesta, "clasificador", string(tipo))
	}
	return tipos.Procedimiento
}

func candidatoRouter(id string, puntaje, p float64, accion string) tipos.Candidato {
	return tipos.Candidato{ID: id, Clase: "procedimiento", Puntaje: puntaje, Lexico: 1,
		Meta: map[string]string{"cobertura": strconv.FormatFloat(puntaje, 'f', 3, 64), "router": accion,
			"p_router": strconv.FormatFloat(p, 'f', 3, 64)}}
}

func (t *turno) anotarRouter(id string, p float64) {
	if t.estado.Inferencias == nil {
		t.estado.Inferencias = map[string]tipos.Dato{}
	}
	t.estado.Inferencias["procedimiento_router"] = tipos.Dato{Valor: id, Origen: origenRouter, Confianza: p}
}

func (t *turno) razonRouter(s string) {
	if t.v != nil {
		t.v.Razon(traza.EtapaV2PlanConsulta, s)
	}
}
