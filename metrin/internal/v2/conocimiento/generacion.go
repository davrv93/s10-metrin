package conocimiento

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"rag-go/internal/v2/tipos"
)

// MotorPlantillas (tipos.GenerationEngine, el de por defecto): arma el texto desde el plan con las
// plantillas de metrin/plantillas/respuestas.yml, sin modelo. Todo dato sale del plan; las plantillas
// solo ponen el marco («Estos son los pasos 1 a 6 de 14:», «¿Seguimos con el paso 7?»).
type MotorPlantillas struct {
	Plantillas *Plantillas
	// Base: para saber si un término del glosario está en sus fuentes (y puede ir en negrita). Sin Base,
	// las plantillas no ponen negritas propias.
	Base *Base
	// FotosEnTexto: además de plan.pasos[].fotos, pone cada foto como ![…](fotos/<ruta>) debajo de SU paso.
	// Por defecto no: la página pinta paso → foto desde el plan (como servidor/pagina.html y el núcleo).
	FotosEnTexto bool
}

var _ tipos.GenerationEngine = (*MotorPlantillas)(nil)

// NuevoMotorPlantillas usa las plantillas de la base (o las embebidas si no hay base).
func NuevoMotorPlantillas(b *Base) *MotorPlantillas {
	if b == nil || b.Plantillas == nil {
		return &MotorPlantillas{Plantillas: PlantillasEmbebidas()}
	}
	return &MotorPlantillas{Plantillas: b.Plantillas, Base: b}
}

// negritaConcepto: un slot del término (TERMINO, TERMINO_A, TERMINO_B) conserva la negrita de la
// plantilla solo si el término aparece en un fragmento que su concepto cita.
func (m *MotorPlantillas) negritaConcepto(porSlot map[string]string) func(slot, valor string) bool {
	return func(slot, valor string) bool {
		id, ok := porSlot[slot]
		return ok && m.Base != nil && m.Base.TerminoRespaldado(id, valor)
	}
}

func (m *MotorPlantillas) Nombre() string { return "plantillas" }

func (m *MotorPlantillas) pl() *Plantillas {
	if m.Plantillas == nil {
		m.Plantillas = PlantillasEmbebidas()
	}
	return m.Plantillas
}

// Redactar arma el texto. Error solo si el plan es incoherente con su tipo (p. ej. CONCEPT sin concepto).
func (m *MotorPlantillas) Redactar(_ context.Context, p tipos.Plan) (string, error) {
	pl := m.pl()
	var partes []string
	add := func(s string) {
		if s = strings.TrimSpace(s); s != "" {
			partes = append(partes, s)
		}
	}
	fin := func() (string, error) { return strings.Join(partes, "\n\n"), nil }
	cita := CitaTexto(p.Fuentes)

	if p.SinEvidencia {
		add(pl.R("SIN_EVIDENCIA", nil))
		return fin()
	}
	switch p.Tipo {
	case tipos.Desconocido:
		if p.Siguiente != nil && p.Siguiente.Texto != "" {
			add(p.Siguiente.Texto)
		} else {
			add(pl.R("ACLARAR_TAREA", nil))
		}
		return fin()

	case tipos.Concepto:
		if p.Concepto == nil {
			return "", fmt.Errorf("plan CONCEPT sin concepto")
		}
		s := slotsConcepto(p.Concepto, cita)
		if len(p.Opciones) > 0 {
			s["TAREA_SUGERIDA"] = minusculaInicial(p.Opciones[0].Titulo)
		}
		add(p.Intro)
		add(pl.R("CONCEPTO", s, OpcRelleno{Negrita: m.negritaConcepto(map[string]string{"TERMINO": p.Concepto.ID})}))
		return fin()

	case tipos.Comparacion:
		if len(p.Conceptos) < 2 {
			return "", fmt.Errorf("plan COMPARISON sin dos conceptos")
		}
		a, b := p.Conceptos[0], p.Conceptos[1]
		add(pl.R("COMPARACION", map[string]string{"TERMINO_A": a.Termino, "DEFINICION_A": a.Definicion,
			"TERMINO_B": b.Termino, "DEFINICION_B": b.Definicion, "CITA": cita},
			OpcRelleno{Negrita: m.negritaConcepto(map[string]string{"TERMINO_A": a.ID, "TERMINO_B": b.ID})}))
		return fin()

	case tipos.Problema:
		if len(p.Errores) == 0 {
			return "", fmt.Errorf("plan TROUBLESHOOTING sin errores")
		}
		for i, e := range p.Errores {
			s := map[string]string{"SINTOMA": e.Sintoma, "SOLUCION": e.Solucion, "CITA": cita}
			op := OpcRelleno{}
			if i < len(p.Errores)-1 {
				op.Omitir = []string{"cita", "cierre"}
			}
			add(pl.R("ERROR_FRECUENTE", s, op))
		}
		return fin()

	case tipos.Navegacion:
		if len(p.Pasos) == 0 {
			if p.Concepto == nil || p.Concepto.EnS10 == "" {
				return "", fmt.Errorf("plan NAVIGATION sin ruta ni concepto")
			}
			add(pl.R("CONCEPTO", slotsConcepto(p.Concepto, cita), OpcRelleno{Omitir: []string{"definicion", "oferta"}}))
			return fin()
		}
		add(p.Intro)
		add(m.pasosTexto(p))
		if p.Concepto != nil && p.Concepto.EnS10 != "" {
			add(pl.R("CONCEPTO", slotsConcepto(p.Concepto, ""), OpcRelleno{Omitir: []string{"definicion", "cita", "oferta"}}))
		}
		// Cita y oferta de NAVEGACION (la ruta ya va en la intro); sin la del archivo, solo la cita.
		if np, ok := pl.Get("NAVEGACION"); ok && np.Forma != "" {
			add(pl.R("NAVEGACION", map[string]string{"CITA": cita, "RUTA_MENU": "-", "ELEMENTO": "-"}, OpcRelleno{Omitir: []string{"ruta"}}))
		} else {
			add(pl.R("FUENTES", map[string]string{"CITA": cita}))
		}
		return fin()
	}

	// PROCEDURE / CONFIGURATION
	if len(p.Pasos) == 0 {
		return "", fmt.Errorf("plan %s sin pasos", p.Tipo)
	}
	if p.Plantilla == "PROCEDIMIENTO_FIN" {
		if p.Intro != "" {
			add(p.Intro)
		} else {
			add(pl.R("PROCEDIMIENTO_FIN", map[string]string{"TOTAL_PASOS": strconv.Itoa(len(p.Pasos))}))
		}
		return fin()
	}
	add(p.Intro)
	mostrados := mostrados(p)
	if len(mostrados) > 0 && mostrados[0] == 1 && len(p.Prerrequisitos) > 0 {
		s := map[string]string{"PRERREQUISITOS": lista(textos(p.Prerrequisitos))}
		if t, ok := pl.Rellenar("PROCEDIMIENTO_INTRO", s, OpcRelleno{Omitir: []string{"inicio", "objetivo", "alcance", "arranque", "cita"}}); ok && t != "" {
			add(t)
		} else {
			add(pl.R("PRERREQUISITOS", s))
		}
	}
	s := map[string]string{"PASOS": m.pasosTexto(p), "TOTAL_PASOS": strconv.Itoa(len(p.Pasos)), "CITA": cita}
	if len(mostrados) > 0 {
		s["N_DESDE"], s["N_HASTA"] = strconv.Itoa(mostrados[0]), strconv.Itoa(mostrados[len(mostrados)-1])
	}
	add(pl.R("PASOS_BLOQUE", s))
	if p.Siguiente != nil && p.Siguiente.Texto != "" {
		add(p.Siguiente.Texto)
	} else if len(p.Verificacion) > 0 {
		add(pl.R("COMPROBACION", map[string]string{"VERIFICACION_FINAL": lista(textos(p.Verificacion))}))
	}
	return fin()
}

func slotsConcepto(c *tipos.ConceptoDef, cita string) map[string]string {
	return map[string]string{"TERMINO": c.Termino, "DEFINICION": c.Definicion, "EN_S10": c.EnS10, "CITA": cita}
}

// mostrados: los n de la parte que se entrega (todos si el plan no lo dice).
func mostrados(p tipos.Plan) []int {
	if len(p.PasosMostrados) > 0 {
		return p.PasosMostrados
	}
	out := make([]int, 0, len(p.Pasos))
	for _, x := range p.Pasos {
		out = append(out, x.N)
	}
	return out
}

// pasosTexto: «n. acción» por paso mostrado (los subpasos ya van sangrados dentro del texto) y, con
// FotosEnTexto, sus fotos debajo, en el orden paso → foto → paso → foto.
func (m *MotorPlantillas) pasosTexto(p tipos.Plan) string {
	ver := map[int]bool{}
	for _, n := range mostrados(p) {
		ver[n] = true
	}
	var b strings.Builder
	for _, x := range p.Pasos {
		if !ver[x.N] {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(strconv.Itoa(x.N) + ". " + strings.TrimSpace(x.Texto))
		if m.FotosEnTexto {
			for _, f := range x.Fotos {
				b.WriteString("\n![" + altFoto(f) + "](" + URLFoto(f.Ruta) + ")")
			}
		}
	}
	return b.String()
}

func altFoto(f tipos.Foto) string {
	t := strings.NewReplacer("[", "", "]", "", "(", "", ")", "", "\n", " ", "*", "").Replace(f.Caption)
	t = strings.TrimSpace(t)
	if r := []rune(t); len(r) > 80 {
		t = string(r[:80]) + "…"
	}
	if t == "" {
		return "Captura del manual"
	}
	return t
}
