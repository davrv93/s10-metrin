package v2

// Respuesta segura y piezas de respaldo del núcleo, sin modelo:
//
//   - Basico (tipos.GenerationEngine): texto determinista desde el plan. Es el respaldo cuando no hay plantillas
//     cargadas y es SIEMPRE el que redacta la respuesta segura (nunca un LLM).
//   - GateMinimo (tipos.QualityGate): comprobaciones mínimas por código cuando el gate de conocimiento no está.
//   - PlanSeguro: lo que sí está respaldado del plan (pasos con fuente, sin fotos ni afirmaciones), o SIN_EVIDENCIA.

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"rag-go/internal/v2/tipos"
)

// TextoSinEvidencia: respuesta cuando nada en kb/ respalda una respuesta.
const TextoSinEvidencia = "No encontré en los manuales de S10 un procedimiento o una definición que respalde una respuesta a esa pregunta, y prefiero no inventarla. ¿Me dice en qué módulo está y qué intenta hacer?"

// TextoAclarar: pregunta de aclaración por defecto (UNKNOWN).
const TextoAclarar = "¿Me cuenta qué tarea quiere hacer en S10 y en qué módulo? Así le guío con los pasos del manual."

// TextoSinProcedimiento: «¿y luego?» sin procedimiento en curso.
const TextoSinProcedimiento = "No tengo un procedimiento en curso en esta conversación. ¿Qué tarea quiere hacer en S10?"

// Basico redacta el plan sin modelo y sin plantillas: estructura fija, todo sale del plan.
type Basico struct{}

func (Basico) Nombre() string { return "basico" }

func (Basico) Redactar(_ context.Context, p tipos.Plan) (string, error) { return RenderBasico(p), nil }

// RenderBasico escribe el plan en texto. Las fotos no van en el texto: la página las pinta desde plan.pasos.
func RenderBasico(p tipos.Plan) string {
	if p.SinEvidencia {
		return TextoSinEvidencia
	}
	var b strings.Builder
	linea := func(s string) {
		if s = strings.TrimSpace(s); s != "" {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(s)
		}
	}
	if p.Intro != "" {
		linea(p.Intro)
	}
	if c := p.Concepto; c != nil {
		linea(fmt.Sprintf("%s: %s", c.Termino, c.Definicion))
		if c.EnS10 != "" {
			linea("En S10: " + c.EnS10)
		}
	}
	if p.Procedimiento != nil && p.Intro == "" && len(PasosVisibles(p)) > 0 {
		linea(p.Procedimiento.Titulo + ":")
	}
	if len(p.Prerrequisitos) > 0 {
		linea("Antes de empezar:")
		for _, x := range p.Prerrequisitos {
			linea("- " + x.Texto)
		}
	}
	for _, x := range PasosVisibles(p) {
		linea(strconv.Itoa(x.N) + ". " + x.Texto)
	}
	for _, e := range p.Errores {
		linea("Si ve «" + e.Sintoma + "»: " + e.Solucion)
	}
	if len(p.Verificacion) > 0 {
		linea("Para comprobar que quedó bien:")
		for _, x := range p.Verificacion {
			linea("- " + x.Texto)
		}
	}
	if p.Siguiente != nil {
		linea(p.Siguiente.Texto)
	}
	if b.Len() == 0 {
		return TextoAclarar
	}
	return b.String()
}

// PasosVisibles: los pasos que se enseñan en este turno (PasosMostrados; sin él, todos).
func PasosVisibles(p tipos.Plan) []tipos.PasoPlan {
	if len(p.PasosMostrados) == 0 {
		return p.Pasos
	}
	quiero := map[int]bool{}
	for _, n := range p.PasosMostrados {
		quiero[n] = true
	}
	var out []tipos.PasoPlan
	for _, x := range p.Pasos {
		if quiero[x.N] {
			out = append(out, x)
		}
	}
	return out
}

// GateMinimo comprueba por código lo mínimo que el núcleo puede ver sin la evidencia de kb/: hay texto; un plan
// con contenido trae fuentes; los pasos van en orden, sin repetirse y con fuente; cada foto tiene id y ruta.
type GateMinimo struct{}

func (GateMinimo) Evaluar(p tipos.Plan, texto string, _ tipos.Estado) tipos.ResultadoCalidad {
	var problemas []string
	if strings.TrimSpace(texto) == "" {
		problemas = append(problemas, "texto vacío")
	}
	conContenido := len(p.Pasos) > 0 || p.Concepto != nil || len(p.Errores) > 0 || len(p.Verificacion) > 0
	if !p.SinEvidencia && conContenido && len(p.Fuentes) == 0 {
		problemas = append(problemas, "plan sin fuentes")
	}
	prev := 0
	vistos := map[int]bool{}
	for _, x := range p.Pasos {
		if vistos[x.N] {
			problemas = append(problemas, fmt.Sprintf("paso %d repetido", x.N))
		}
		if x.N <= prev {
			problemas = append(problemas, fmt.Sprintf("paso %d fuera de orden", x.N))
		}
		vistos[x.N], prev = true, x.N
		if len(x.Fuente) == 0 {
			problemas = append(problemas, fmt.Sprintf("paso %d sin fuente", x.N))
		}
		for _, f := range x.Fotos {
			if f.ID == "" || f.Ruta == "" {
				problemas = append(problemas, fmt.Sprintf("foto sin id o ruta en el paso %d", x.N))
			}
		}
	}
	puntaje := 1 - 0.25*float64(len(problemas))
	if puntaje < 0 {
		puntaje = 0
	}
	return tipos.ResultadoCalidad{Paso: len(problemas) == 0, Puntaje: puntaje, Problemas: problemas}
}

// PlanSeguro deja solo lo respaldado: pasos con fuente (sin fotos: «mejor un paso sin foto»), el concepto si
// trae fuente, verificación y prerrequisitos con fuente; nada de afirmaciones ni errores sin fuente. Si no queda
// nada, el plan SIN_EVIDENCIA.
func PlanSeguro(p tipos.Plan) tipos.Plan {
	s := tipos.Plan{Version: 2, Tipo: p.Tipo, Procedimiento: p.Procedimiento, Fuentes: p.Fuentes, Plantilla: "RESPUESTA_SEGURA"}
	visibles := map[int]bool{}
	for _, x := range PasosVisibles(p) {
		visibles[x.N] = true
	}
	for _, x := range p.Pasos {
		if len(x.Fuente) > 0 && strings.TrimSpace(x.Texto) != "" {
			x.Fotos = nil
			s.Pasos = append(s.Pasos, x)
			if visibles[x.N] {
				s.PasosMostrados = append(s.PasosMostrados, x.N)
			}
		}
	}
	if len(s.Pasos) > 0 && len(s.PasosMostrados) == 0 {
		s.Pasos = nil // lo que había que enseñar no tenía fuente
	}
	if p.Concepto != nil && len(p.Concepto.Fuente) > 0 {
		c := *p.Concepto
		s.Concepto = &c
	}
	s.Prerrequisitos = conFuente(p.Prerrequisitos)
	s.Verificacion = conFuente(p.Verificacion)
	for _, e := range p.Errores {
		if len(e.Fuente) > 0 {
			s.Errores = append(s.Errores, e)
		}
	}
	if (len(s.Pasos) == 0 && s.Concepto == nil && len(s.Errores) == 0) || len(s.Fuentes) == 0 {
		return PlanSinEvidencia(p.Tipo)
	}
	return s
}

func conFuente(xs []tipos.ConFuente) []tipos.ConFuente {
	var out []tipos.ConFuente
	for _, x := range xs {
		if len(x.Fuente) > 0 {
			out = append(out, x)
		}
	}
	return out
}

// PlanSinEvidencia: el plan de SIN_EVIDENCIA para un tipo.
func PlanSinEvidencia(t tipos.TipoRespuesta) tipos.Plan {
	return tipos.Plan{Version: 2, Tipo: t, SinEvidencia: true, Fuentes: []tipos.FuentePlan{}, Plantilla: "SIN_EVIDENCIA"}
}
