package conocimiento

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"rag-go/internal/v2/tipos"
)

// Gate (tipos.QualityGate): valida plan y texto por código, contra la evidencia de kb/. Sin modelo.
//
// Reglas (cada problema empieza por su regla):
//
//	intencion     el tipo del plan responde al tipo pedido y trae lo que ese tipo exige
//	evidencia     cada paso, prerrequisito, verificación y error cita fuentes que existen como (id, manual)
//	              y que el procedimiento le asigna; la definición es la del glosario
//	negritas      cada término entre ** ** de un paso aparece en sus fuentes (texto, pasos u OCR)
//	orden         pasos en orden, sin duplicados, sin inventados y sin faltar frente al procedimiento
//	fotos         cada foto es de su paso (y la fuente la asocia), sin repetirse
//	fuentes       hay fuentes, ninguna de marketing ni tangencial, y cada una la respalda una cita del plan
//	texto         el texto no trae negritas, atajos, rutas, citas, fotos ni números de paso fuera del plan
//	afirmaciones  cada afirmación apunta a un (fragmento, manual) que existe
type Gate struct {
	Base *Base
}

var _ tipos.QualityGate = (*Gate)(nil)

// NuevoGate sobre una base cargada.
func NuevoGate(b *Base) *Gate { return &Gate{Base: b} }

// Pesos de cada regla en el puntaje (1 − suma de las que fallan / suma total).
var pesosGate = map[string]float64{
	"intencion": 0.25, "evidencia": 0.20, "negritas": 0.15, "orden": 0.15,
	"fotos": 0.10, "fuentes": 0.10, "texto": 0.15, "afirmaciones": 0.05,
}

type evaluacion struct {
	g         *Gate
	p         tipos.Plan
	proc      *Procedimiento
	problemas []string
	fallan    map[string]bool
}

func (ev *evaluacion) mal(regla, f string, a ...any) {
	ev.problemas = append(ev.problemas, regla+": "+fmt.Sprintf(f, a...))
	ev.fallan[regla] = true
}

// Evaluar cumple tipos.QualityGate.
func (g *Gate) Evaluar(p tipos.Plan, texto string, e tipos.Estado) tipos.ResultadoCalidad {
	ev := &evaluacion{g: g, p: p, fallan: map[string]bool{}}
	if p.Procedimiento != nil {
		ev.proc, _ = g.Base.Proc(p.Procedimiento.ID)
	}
	ev.intencion(e)
	ev.evidencia()
	ev.negritas()
	ev.negritasConceptos()
	ev.orden()
	ev.fotos()
	ev.fuentes()
	ev.texto(texto)
	ev.afirmaciones()

	total, resta := 0.0, 0.0
	for r, w := range pesosGate {
		total += w
		if ev.fallan[r] {
			resta += w
		}
	}
	puntaje := math.Round((1-resta/total)*1000) / 1000
	return tipos.ResultadoCalidad{Paso: len(ev.problemas) == 0, Puntaje: puntaje, Problemas: ev.problemas}
}

// --- intención ---------------------------------------------------------------------------------------

func (ev *evaluacion) intencion(e tipos.Estado) {
	p := ev.p
	if e.Tipo != "" && p.Tipo != e.Tipo {
		aclara := p.Tipo == tipos.Desconocido && p.Siguiente != nil
		// Una comparación entre dos sinónimos del mismo concepto se responde con ese concepto.
		if e.Tipo == tipos.Comparacion && p.Tipo == tipos.Concepto && p.Concepto != nil {
			if cc, ok := ev.g.Base.Conc(p.Concepto.ID); ok && len(ev.g.Base.formasNombradas(cc, e.Pregunta)) >= 2 {
				aclara = true
			}
		}
		if !aclara {
			ev.mal("intencion", "se pidió %s y el plan es %s", e.Tipo, p.Tipo)
		}
	}
	if p.SinEvidencia {
		if len(p.Pasos) > 0 || p.Concepto != nil || len(p.Errores) > 0 {
			ev.mal("intencion", "plan SIN_EVIDENCIA con contenido")
		}
		return
	}
	switch p.Tipo {
	case tipos.Procedimiento, tipos.Configuracion:
		if p.Procedimiento == nil || len(p.Pasos) == 0 {
			ev.mal("intencion", "%s sin procedimiento o sin pasos", p.Tipo)
		}
	case tipos.Navegacion:
		if len(p.Pasos) == 0 && (p.Concepto == nil || p.Concepto.EnS10 == "") {
			ev.mal("intencion", "NAVIGATION sin ruta documentada")
		}
	case tipos.Problema:
		if len(p.Errores) == 0 {
			ev.mal("intencion", "TROUBLESHOOTING sin errores documentados")
		}
	case tipos.Concepto:
		if p.Concepto == nil {
			ev.mal("intencion", "CONCEPT sin concepto")
		}
		if len(p.Pasos) > 0 {
			ev.mal("intencion", "CONCEPT con pasos")
		}
	case tipos.Comparacion:
		if len(p.Conceptos) != 2 {
			ev.mal("intencion", "COMPARISON con %d conceptos (se esperan 2)", len(p.Conceptos))
		}
	case tipos.Desconocido:
		if p.Siguiente == nil || strings.TrimSpace(p.Siguiente.Texto) == "" {
			ev.mal("intencion", "UNKNOWN sin pregunta de aclaración")
		}
	}
}

// --- evidencia ---------------------------------------------------------------------------------------

// cita resuelve una cita del plan: con la tabla del procedimiento (o el par del implícito).
func (ev *evaluacion) cita(id string) (Cita, error) {
	if ev.proc == nil {
		return Cita{}, fmt.Errorf("el plan no tiene un procedimiento conocido para resolver %s", id)
	}
	return ev.g.Base.CitaProcedimiento(ev.proc, id)
}

// fuentesDePaso: las de su definición (con subpasos).
func fuentesDePaso(s *tipos.Paso) map[string]bool {
	m := map[string]bool{}
	var rec func(*tipos.Paso)
	rec = func(x *tipos.Paso) {
		for _, id := range x.Fuente {
			m[id] = true
		}
		for i := range x.Sub {
			rec(&x.Sub[i])
		}
	}
	rec(s)
	return m
}

func (ev *evaluacion) evidencia() {
	p := ev.p
	if p.Procedimiento != nil && ev.proc == nil && (len(p.Pasos) > 0 || len(p.Errores) > 0) {
		ev.mal("evidencia", "el procedimiento %s no existe", p.Procedimiento.ID)
		return
	}
	for _, x := range p.Pasos {
		if len(x.Fuente) == 0 {
			ev.mal("evidencia", "paso %d sin fuente", x.N)
			continue
		}
		var def *tipos.Paso
		if ev.proc != nil {
			def, _ = ev.proc.PasoPorID(x.ID)
		}
		asignadas := map[string]bool{}
		if def != nil {
			asignadas = fuentesDePaso(def)
		}
		for _, id := range x.Fuente {
			if _, err := ev.cita(id); err != nil {
				ev.mal("evidencia", "paso %d: %v", x.N, err)
			} else if def != nil && !asignadas[id] {
				ev.mal("evidencia", "paso %d cita %s, que el procedimiento no le asigna", x.N, id)
			}
		}
	}
	conFuente := func(que string, xs []tipos.ConFuente, def []tipos.ConFuente) {
		for i, x := range xs {
			if len(x.Fuente) == 0 {
				ev.mal("evidencia", "%s %d sin fuente", que, i+1)
			}
			for _, id := range x.Fuente {
				if _, err := ev.cita(id); err != nil {
					ev.mal("evidencia", "%s %d: %v", que, i+1, err)
				}
			}
			if ev.proc != nil && !contieneConFuente(def, x.Texto) {
				ev.mal("evidencia", "%s %d no está en el procedimiento: «%s»", que, i+1, recortar(x.Texto, 60))
			}
		}
	}
	var defPre, defVer []tipos.ConFuente
	if ev.proc != nil {
		defPre, defVer = ev.proc.Prerrequisitos, ev.proc.Verificacion
	}
	conFuente("prerrequisito", p.Prerrequisitos, defPre)
	conFuente("verificación", p.Verificacion, defVer)
	for i, x := range p.Errores {
		if len(x.Fuente) == 0 {
			ev.mal("evidencia", "error %d sin fuente", i+1)
		}
		for _, id := range x.Fuente {
			if _, err := ev.cita(id); err != nil {
				ev.mal("evidencia", "error %d: %v", i+1, err)
			}
		}
		if ev.proc != nil {
			ok := false
			for _, d := range ev.proc.ErroresFrecuentes {
				ok = ok || (d.Sintoma == x.Sintoma && d.Solucion == x.Solucion)
			}
			if !ok {
				ev.mal("evidencia", "error %d no está en el procedimiento: «%s»", i+1, recortar(x.Sintoma, 60))
			}
		}
	}
	for _, c := range ev.conceptos() {
		g, ok := ev.g.Base.Conc(c.ID)
		if !ok {
			ev.mal("evidencia", "el concepto %s no está en el glosario", c.ID)
			continue
		}
		if c.Definicion != g.Definicion || c.EnS10 != g.EnS10 {
			ev.mal("evidencia", "la definición de %s no es la del glosario", c.ID)
		}
		if len(c.Fuente) == 0 {
			ev.mal("evidencia", "el concepto %s no tiene fuente", c.ID)
		}
		for _, id := range c.Fuente {
			if _, err := ev.g.Base.CitaConcepto(g, id); err != nil {
				ev.mal("evidencia", "%v", err)
			}
		}
	}
}

func contieneConFuente(xs []tipos.ConFuente, t string) bool {
	for _, x := range xs {
		if x.Texto == t {
			return true
		}
	}
	return false
}

// conceptos del plan (sin repetir).
func (ev *evaluacion) conceptos() []tipos.ConceptoDef {
	var out []tipos.ConceptoDef
	vis := map[string]bool{}
	if ev.p.Concepto != nil {
		out = append(out, *ev.p.Concepto)
		vis[ev.p.Concepto.ID] = true
	}
	for _, c := range ev.p.Conceptos {
		if !vis[c.ID] {
			vis[c.ID] = true
			out = append(out, c)
		}
	}
	return out
}

// --- negritas ----------------------------------------------------------------------------------------

func (ev *evaluacion) negritas() {
	if !ev.g.Base.hayIndice() || ev.proc == nil {
		return // sin índice no se puede comprobar (ya avisado al cargar)
	}
	for _, x := range ev.p.Pasos {
		ts := negritas(x.Texto)
		if len(ts) == 0 {
			continue
		}
		corpus := ev.corpusPaso(x)
		for _, t := range ts {
			if !strings.Contains(corpus, plegar(t)) {
				ev.mal("negritas", "paso %d: «%s» no aparece en sus fuentes", x.N, t)
			}
		}
	}
	for i, x := range append(append([]tipos.ConFuente{}, ev.p.Prerrequisitos...), ev.p.Verificacion...) {
		ts := negritas(x.Texto)
		if len(ts) == 0 {
			continue
		}
		corpus := ev.corpusIDs(x.Fuente)
		for _, t := range ts {
			if !strings.Contains(corpus, plegar(t)) {
				ev.mal("negritas", "elemento %d: «%s» no aparece en sus fuentes", i+1, t)
			}
		}
	}
	for i, x := range ev.p.Errores {
		corpus := ev.corpusIDs(x.Fuente)
		for _, t := range negritas(x.Sintoma + " " + x.Solucion) {
			if !strings.Contains(corpus, plegar(t)) {
				ev.mal("negritas", "error %d: «%s» no aparece en sus fuentes", i+1, t)
			}
		}
	}
}

// negritasConceptos: las negritas de una definición o de «en_s10» deben estar en las fuentes del concepto.
func (ev *evaluacion) negritasConceptos() {
	for _, c := range ev.conceptos() {
		g, ok := ev.g.Base.Conc(c.ID)
		if !ok || !ev.g.Base.hayIndice() {
			continue
		}
		var b strings.Builder
		for _, id := range c.Fuente {
			if ci, err := ev.g.Base.CitaConcepto(g, id); err == nil && ci.Fragmento != nil {
				b.WriteString(ci.Fragmento.TextoPlegado() + "\n")
			}
		}
		for _, t := range negritas(c.Definicion + " " + c.EnS10 + " " + c.Ejemplo) {
			if !strings.Contains(b.String(), plegar(t)) {
				ev.mal("negritas", "concepto %s: «%s» no aparece en sus fuentes", c.ID, t)
			}
		}
	}
}

func (ev *evaluacion) corpusIDs(ids []string) string {
	var b strings.Builder
	for _, id := range ids {
		if c, err := ev.cita(id); err == nil && c.Fragmento != nil {
			b.WriteString(c.Fragmento.TextoPlegado())
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// corpusPaso: sus fragmentos fuente + el OCR de sus fotos (fragmento imagen o el del YAML).
func (ev *evaluacion) corpusPaso(x tipos.PasoPlan) string {
	var b strings.Builder
	b.WriteString(ev.corpusIDs(x.Fuente))
	for _, f := range x.Fotos {
		if fr, ok := ev.g.Base.Fragmentos.Imagen(f.Ruta); ok {
			b.WriteString(fr.TextoPlegado())
			b.WriteByte('\n')
		}
	}
	if def, ok := ev.proc.PasoPorID(x.ID); ok {
		var rec func(*tipos.Paso)
		rec = func(s *tipos.Paso) {
			for _, f := range s.Fotos {
				b.WriteString(plegar(f.Caption + "\n" + f.OCR))
				b.WriteByte('\n')
			}
			for i := range s.Sub {
				rec(&s.Sub[i])
			}
		}
		rec(def)
	}
	return b.String()
}

// --- orden -------------------------------------------------------------------------------------------

func (ev *evaluacion) orden() {
	p := ev.p
	prev := 0
	ids := map[string]bool{}
	ns := map[int]bool{}
	for _, x := range p.Pasos {
		if ns[x.N] || ids[x.ID] {
			ev.mal("orden", "paso %d (%s) repetido", x.N, x.ID)
		}
		if x.N <= prev {
			ev.mal("orden", "paso %d fuera de orden (después del %d)", x.N, prev)
		}
		prev, ns[x.N], ids[x.ID] = x.N, true, true
		if ev.proc != nil {
			if x.N < 1 || x.N > len(ev.proc.Pasos) || ev.proc.Pasos[x.N-1].ID != x.ID {
				ev.mal("orden", "paso %d (%s) no es el paso %d del procedimiento", x.N, x.ID, x.N)
			}
		}
	}
	if ev.proc != nil && (p.Tipo == tipos.Procedimiento || p.Tipo == tipos.Configuracion) && !p.SinEvidencia {
		for _, s := range ev.proc.Pasos {
			if !ns[s.N] {
				ev.mal("orden", "falta el paso %d del procedimiento", s.N)
			}
		}
	}
	prevM := 0
	for i, n := range p.PasosMostrados {
		if !ns[n] {
			ev.mal("orden", "se muestra el paso %d, que no está en el plan", n)
		}
		if n <= prevM || (i > 0 && n != prevM+1 && p.Tipo != tipos.Navegacion) {
			ev.mal("orden", "pasos mostrados fuera de orden o salteados: %v", p.PasosMostrados)
			break
		}
		prevM = n
	}
}

// --- fotos -------------------------------------------------------------------------------------------

func (ev *evaluacion) fotos() {
	vistas := map[string]int{}
	for _, x := range ev.p.Pasos {
		var def *tipos.Paso
		if ev.proc != nil {
			def, _ = ev.proc.PasoPorID(x.ID)
		}
		permitidas := map[string]tipos.Foto{}
		duenas := map[string]*tipos.Paso{}
		if def != nil {
			var rec func(*tipos.Paso)
			rec = func(s *tipos.Paso) {
				for _, f := range s.Fotos {
					permitidas[f.ID] = f
					duenas[f.ID] = s
				}
				for i := range s.Sub {
					rec(&s.Sub[i])
				}
			}
			rec(def)
		}
		for _, f := range x.Fotos {
			if f.ID == "" || f.Ruta == "" {
				ev.mal("fotos", "paso %d: foto sin id o sin ruta", x.N)
				continue
			}
			if otro, ok := vistas[f.ID]; ok {
				ev.mal("fotos", "la foto %s está en los pasos %d y %d", f.ID, otro, x.N)
			}
			vistas[f.ID] = x.N
			d, ok := permitidas[f.ID]
			if !ok || d.Ruta != f.Ruta {
				ev.mal("fotos", "paso %d: la foto %s no es de este paso", x.N, f.ID)
				continue
			}
			if idFoto(f.Ruta) != f.ID {
				ev.mal("fotos", "paso %d: la foto %s no corresponde a su ruta", x.N, f.ID)
			}
			if !rutaFotoValida(f.Ruta) {
				ev.mal("fotos", "paso %d: ruta no válida %q", x.N, f.Ruta)
			}
			if ev.g.Base.hayIndice() && !ev.g.Base.fuenteAsociaFoto(ev.proc, duenas[f.ID], f.Ruta) {
				ev.mal("fotos", "paso %d: ninguna fuente del paso asocia la foto %s", x.N, f.ID)
			}
			if ev.g.Base.DirDatos != "" && !strings.HasPrefix(f.Ruta, "http") && !existe(filepath.Join(ev.g.Base.DirDatos, f.Ruta)) {
				ev.mal("fotos", "paso %d: la foto %s no existe en %s", x.N, f.ID, ev.g.Base.DirDatos)
			}
		}
	}
}

// --- fuentes -----------------------------------------------------------------------------------------

// citasDelPlan: todas las citas resueltas del plan (pasos, prerrequisitos, verificación, errores, conceptos).
func (ev *evaluacion) citasDelPlan() []Cita {
	var ids []string
	for _, x := range ev.p.Pasos {
		ids = append(ids, x.Fuente...)
	}
	for _, x := range append(append([]tipos.ConFuente{}, ev.p.Prerrequisitos...), ev.p.Verificacion...) {
		ids = append(ids, x.Fuente...)
	}
	for _, x := range ev.p.Errores {
		ids = append(ids, x.Fuente...)
	}
	var out []Cita
	for _, id := range unicos(ids) {
		if c, err := ev.cita(id); err == nil {
			out = append(out, c)
		}
	}
	for _, c := range ev.conceptos() {
		if g, ok := ev.g.Base.Conc(c.ID); ok {
			for _, id := range c.Fuente {
				if ci, err := ev.g.Base.CitaConcepto(g, id); err == nil {
					out = append(out, ci)
				}
			}
		}
	}
	return out
}

func (ev *evaluacion) fuentes() {
	p := ev.p
	conContenido := len(p.Pasos) > 0 || p.Concepto != nil || len(p.Errores) > 0 || len(p.Conceptos) > 0
	if p.SinEvidencia || !conContenido {
		return
	}
	if len(p.Fuentes) == 0 {
		ev.mal("fuentes", "plan sin fuentes")
		return
	}
	citas := ev.citasDelPlan()
	respaldadas := map[string]bool{}
	for _, c := range citas {
		respaldadas[NombreManual(c)] = true
		respaldadas[c.Manual] = true // «Importado a mano» y su título son la misma fuente
		if c.Fragmento != nil && (EsMarketing(c.Manual, c.Fragmento.Confianza) || EsTangencial(c.Manual, c.Fragmento.Confianza)) {
			ev.mal("fuentes", "la cita %s es de una fuente de marketing o tangencial (%s)", c.ID, c.Manual)
		}
	}
	for _, f := range p.Fuentes {
		switch {
		case EsMarketing(f.Manual, ""):
			ev.mal("fuentes", "fuente de marketing: %s", f.Manual)
		case EsTangencial(f.Manual, ""):
			ev.mal("fuentes", "fuente tangencial: %s", f.Manual)
		case !respaldadas[f.Manual]:
			ev.mal("fuentes", "«%s» no la respalda ninguna cita del plan (tangencial)", f.Manual)
		}
	}
}

// --- texto -------------------------------------------------------------------------------------------

var (
	pasoNumRe  = regexp.MustCompile(`^(\d+)\.\s`)
	imagenRe   = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)\)`)
	encabezado = regexp.MustCompile(`(?i)^paso \d+ de \d+\.?$`)
	paginaRe   = regexp.MustCompile(`(?i)\b(?:p\.|p[aá]g\.|p[aá]gina)\s*(\d+)`)
	marcaRe    = regexp.MustCompile(`\[F\d+\]`)
	manualRe   = regexp.MustCompile(`Manual (?:de |del )?(?:[A-ZÁÉÍÓÚÑ][\p{L}]+)(?: (?:de |del |y )?[A-ZÁÉÍÓÚÑ][\p{L}]+)*|Gu[ií]a de Usuario de [^,;)\n_]+`)
)

// corpusPlan: todo lo que el plan dice, plegado (para atajos y rutas del texto).
func (ev *evaluacion) corpusPlan() string {
	p := ev.p
	var xs []string
	xs = append(xs, p.Intro)
	for _, x := range p.Pasos {
		xs = append(xs, x.Texto)
		for _, f := range x.Fotos { // el pie de foto también sale en el texto (FotosEnTexto)
			xs = append(xs, f.Caption)
		}
	}
	for _, x := range append(append([]tipos.ConFuente{}, p.Prerrequisitos...), p.Verificacion...) {
		xs = append(xs, x.Texto)
	}
	for _, x := range p.Errores {
		xs = append(xs, x.Sintoma, x.Solucion)
	}
	for _, c := range ev.conceptos() {
		xs = append(xs, c.Termino, c.Definicion, c.EnS10, c.Ejemplo)
	}
	if p.Siguiente != nil {
		xs = append(xs, p.Siguiente.Texto)
	}
	for _, o := range p.Opciones {
		xs = append(xs, o.Titulo)
	}
	for _, f := range p.Fuentes {
		xs = append(xs, f.Manual, f.Seccion)
	}
	if ev.proc != nil {
		xs = append(xs, ev.proc.Titulo, ev.proc.Entidades.Modulo)
		for _, s := range ev.proc.Pasos {
			xs = append(xs, s.Donde)
		}
	}
	return plegar(strings.Join(xs, "\n"))
}

// negritasPermitidas: las de los elementos del plan (pasos, prerrequisitos, verificación, errores,
// definiciones), que la regla «negritas» comprueba contra SUS fuentes, y el término de un concepto solo
// si aparece en un fragmento que ese concepto cita. Nada más: ni el módulo ni un título de sección.
func (ev *evaluacion) negritasPermitidas() map[string]bool {
	p := ev.p
	m := map[string]bool{}
	add := func(s string) {
		for _, t := range negritas(s) {
			m[plegar(t)] = true
		}
	}
	for _, x := range p.Pasos {
		add(x.Texto)
	}
	for _, x := range append(append([]tipos.ConFuente{}, p.Prerrequisitos...), p.Verificacion...) {
		add(x.Texto)
	}
	for _, x := range p.Errores {
		add(x.Sintoma + " " + x.Solucion)
	}
	for _, c := range ev.conceptos() {
		for _, n := range append([]string{c.Termino}, c.Sinonimos...) {
			if ev.g.Base.TerminoRespaldado(c.ID, n) {
				m[plegar(n)] = true
			}
		}
		add(c.Definicion + " " + c.EnS10)
	}
	return m
}

func (ev *evaluacion) texto(texto string) {
	p := ev.p
	if strings.TrimSpace(texto) == "" {
		ev.mal("texto", "texto vacío")
		return
	}
	permitidas := ev.negritasPermitidas()
	revisar := func(donde, s string) {
		for _, t := range negritas(s) {
			if !permitidas[plegar(t)] && !encabezado.MatchString(strings.TrimSpace(t)) {
				ev.mal("texto", "%s: «%s» en negrita no viene del plan", donde, t)
			}
		}
	}
	revisar("intro", p.Intro)
	revisar("texto", texto)

	corpus := ev.corpusPlan()
	for _, a := range atajoRe.FindAllString(texto, -1) {
		if !strings.Contains(corpus, plegar(a)) {
			ev.mal("texto", "atajo «%s» fuera del plan", a)
		}
	}
	for _, m := range marcaRe.FindAllString(texto, -1) {
		ev.mal("texto", "marca de cita «%s» que el plan no usa", m)
	}
	paginas := map[string]bool{}
	for _, f := range p.Fuentes {
		for _, n := range f.Paginas {
			paginas[strconv.Itoa(n)] = true
		}
	}
	for _, x := range p.Pasos {
		if x.Pagina > 0 {
			paginas[strconv.Itoa(x.Pagina)] = true
		}
	}
	for _, m := range paginaRe.FindAllStringSubmatch(texto, -1) {
		if !paginas[m[1]] {
			ev.mal("texto", "página %s que no está en las fuentes del plan", m[1])
		}
	}
	var nombres []string
	for _, f := range p.Fuentes {
		nombres = append(nombres, plegar(f.Manual))
	}
	for _, m := range manualRe.FindAllString(texto, -1) {
		mm := plegar(strings.TrimSpace(m))
		ok := false
		for _, n := range nombres {
			ok = ok || strings.HasPrefix(n, mm) || strings.HasPrefix(mm, n)
		}
		if !ok && !strings.Contains(corpus, mm) {
			ev.mal("texto", "cita «%s» fuera de las fuentes del plan", m)
		}
	}
	ev.rutas(texto, corpus)
	ev.lineasDePasos(texto)
}

// rutas: en «A › B › C» los tramos del medio deben estar en el plan; del primero basta su final y del
// último su comienzo (el resto es la frase que los rodea). Las líneas de cita no cuentan.
func (ev *evaluacion) rutas(texto, corpus string) {
	for _, l := range strings.Split(texto, "\n") {
		t := strings.TrimSpace(l)
		// Las líneas de cita y las de foto (su texto alternativo es el de la fuente) no cuentan.
		if !rutaSepRe.MatchString(t) || strings.HasPrefix(t, "_") || strings.HasPrefix(t, "![") || strings.HasPrefix(plegar(t), "fuente") {
			continue
		}
		segs := rutaSepRe.Split(t, -1)
		for i, s := range segs {
			s = strings.Trim(s, " .,:;*_()«»\"")
			if s == "" {
				continue
			}
			ws := strings.Fields(s)
			ok := false
			switch {
			case i == 0:
				for k := 1; k <= len(ws) && !ok; k++ {
					ok = strings.Contains(corpus, plegar(strings.Join(ws[len(ws)-k:], " ")))
				}
			case i == len(segs)-1:
				for k := len(ws); k >= 1 && !ok; k-- {
					ok = strings.Contains(corpus, plegar(strings.Trim(strings.Join(ws[:k], " "), ".,:;")))
				}
			default:
				ok = strings.Contains(corpus, plegar(s))
			}
			if !ok {
				ev.mal("texto", "ruta de menú fuera del plan: «%s»", recortar(t, 80))
				break
			}
		}
	}
}

// lineasDePasos: los «n.» del texto son pasos mostrados del plan, en orden, y cada foto enlazada va
// debajo de SU paso.
func (ev *evaluacion) lineasDePasos(texto string) {
	p := ev.p
	mostr := map[int]bool{}
	for _, n := range mostrados(p) {
		mostr[n] = true
	}
	fotoDe := map[string]int{}
	for _, x := range p.Pasos {
		for _, f := range x.Fotos {
			fotoDe[f.Ruta] = x.N
			fotoDe[URLFoto(f.Ruta)] = x.N
		}
	}
	actual, prev := 0, 0
	var vistos []int
	for _, l := range strings.Split(texto, "\n") {
		if m := pasoNumRe.FindStringSubmatch(l); m != nil && len(p.Pasos) > 0 {
			n, _ := strconv.Atoi(m[1])
			if !mostr[n] {
				ev.mal("texto", "el texto trae el paso %d, que no se muestra en el plan", n)
			}
			if n <= prev {
				ev.mal("texto", "pasos del texto fuera de orden (%d después de %d)", n, prev)
			}
			actual, prev = n, n
			vistos = append(vistos, n)
		}
		for _, m := range imagenRe.FindAllStringSubmatch(l, -1) {
			dueno, ok := fotoDe[m[1]]
			switch {
			case !ok:
				ev.mal("texto", "foto %s que el plan no trae", m[1])
			case dueno != actual:
				ev.mal("texto", "la foto %s va bajo el paso %d y es del paso %d", m[1], actual, dueno)
			}
		}
	}
	if len(vistos) > 0 && (p.Tipo == tipos.Procedimiento || p.Tipo == tipos.Configuracion) {
		want := mostrados(p)
		sort.Ints(vistos)
		if len(vistos) != len(want) {
			ev.mal("texto", "el texto trae %d pasos y el plan muestra %d", len(vistos), len(want))
		}
	}
}

// --- afirmaciones ------------------------------------------------------------------------------------

func (ev *evaluacion) afirmaciones() {
	ns := map[int]bool{}
	for _, x := range ev.p.Pasos {
		ns[x.N] = true
	}
	for i, a := range ev.p.Afirmaciones {
		if a.Fragmento == "" || a.Manual == "" {
			ev.mal("afirmaciones", "afirmación %d sin par (fragmento, manual)", i+1)
			continue
		}
		if _, err := ev.g.Base.CitaPar(a.Fragmento, a.Manual); err != nil {
			ev.mal("afirmaciones", "afirmación %d: %v", i+1, err)
		}
		if a.Paso > 0 && !ns[a.Paso] {
			ev.mal("afirmaciones", "afirmación %d apunta al paso %d, que no está en el plan", i+1, a.Paso)
		}
	}
}
