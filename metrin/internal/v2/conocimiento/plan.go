package conocimiento

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"rag-go/internal/v2/tipos"
)

// PrefijoImplicito marca un procedimiento «implícito»: el que se arma con los pasos[{texto, fotos}] de
// una sección oficial del manual cuando ningún procedimiento YAML cubre la tarea. Su id es
// «implicito:<id del fragmento>@<manual>» (el par que resuelve la cita).
const PrefijoImplicito = "implicito:"

// IDImplicito arma el id de un procedimiento implícito.
func IDImplicito(id, manual string) string { return PrefijoImplicito + id + "@" + manual }

// EsImplicito: el plan (o la referencia) es de un procedimiento implícito.
func EsImplicito(ref *tipos.Ref) bool {
	return ref != nil && strings.HasPrefix(ref.ID, PrefijoImplicito)
}

func parseImplicito(ref string) (id, manual string, ok bool) {
	s, ok := strings.CutPrefix(ref, PrefijoImplicito)
	if !ok {
		return "", "", false
	}
	id, manual, ok = strings.Cut(s, "@")
	return id, manual, ok && id != "" && manual != ""
}

var numeracionRe = regexp.MustCompile(`^\s*[\d.]+\s+|\s*\(\d+\)\s*$`)

// procedimientoImplicito arma el procedimiento desde los pasos de la sección (texto literal, sin
// negritas; cada foto en el paso al que la sección la asocia, sin repetir).
func (b *Base) procedimientoImplicito(ref string) (*Procedimiento, bool) {
	id, manual, ok := parseImplicito(ref)
	if !ok {
		return nil, false
	}
	fr, ok := b.Fragmentos.Buscar(id, manual)
	if !ok || !fragmentoConPasos(fr) {
		return nil, false
	}
	titulo := strings.TrimSpace(numeracionRe.ReplaceAllString(fr.Seccion, ""))
	if titulo == "" {
		titulo = fr.Titulo
	}
	f := tipos.Fuente{ID: id, Manual: manual, Seccion: fr.Seccion, Pagina: fr.Pagina}
	p := &Procedimiento{
		ProcedimientoDef: tipos.ProcedimientoDef{ID: ref, Titulo: titulo, Fuentes: []tipos.Fuente{f}},
		Archivo:          fr.Archivo, Linea: fr.Linea,
		tabla: map[string]tipos.Fuente{id: f}, ambiguos: map[string]bool{}, lineas: map[string]int{},
	}
	vistas := map[string]bool{}
	for _, pf := range pasosConTexto(fr) {
		n := len(p.Pasos) + 1
		paso := tipos.Paso{ID: ref + "#" + strconv.Itoa(n), N: n, Accion: pf.Texto, Pagina: fr.Pagina, Fuente: []string{id}}
		for _, r := range pf.Fotos {
			if !rutaFotoValida(r) || vistas[r] {
				continue
			}
			vistas[r] = true
			paso.Fotos = append(paso.Fotos, tipos.Foto{ID: idFoto(r), Ruta: r, Pagina: fr.Pagina, Caption: pf.Texto})
		}
		p.Pasos = append(p.Pasos, paso)
	}
	return p, true
}

func pasosConTexto(fr *Fragmento) []PasoFuente {
	var out []PasoFuente
	for _, pf := range fr.Pasos {
		if t := strings.TrimSpace(pf.Texto); t != "" {
			out = append(out, PasoFuente{Texto: t, Fotos: pf.Fotos})
		}
	}
	return out
}

// Constructor arma el plan (tipos.Plan) por tipo de respuesta. Cumple v2.Constructor.
type Constructor struct {
	Base *Base
	Rec  *Recuperador // para buscar lo que los candidatos del núcleo no traigan (puede ser nil)

	MaxSinPartes  int // con más pasos que esto, la entrega es por partes (TamanoParte)
	PasosPorParte int // pasos por parte (TamanoParte)

	// Umbrales sobre el puntaje del Recuperador (cobertura léxica en [0, 1], o el híbrido/reranker).
	// PROVISIONALES (06-10-2026): fijados a mano con las consultas de las pruebas; se calibran con
	// metrin/eval/v2_oro.jsonl cuando exista.
	UmbralProcedimiento float64 // 0.6
	UmbralConcepto      float64 // 0.6
	UmbralError         float64 // 0.6
	UmbralFragmento     float64 // 0.7
	UmbralOpcion        float64 // 0.4: mínimo para ofrecer un procedimiento en una aclaración

	// Dominio: un candidato bajo su umbral también se acepta si cubre al menos UmbralMinimo de la consulta
	// y su BM25 supera en RatioDominio al mejor candidato de OTRO elemento (otro procedimiento, otro
	// concepto). Las preguntas largas o con relleno («Buenas tardes, soy asistente de costos…») bajan la
	// cobertura aunque un procedimiento destaque con claridad; sin un rival cerca, ese es la respuesta.
	UmbralMinimo float64 // 0.3
	RatioDominio float64 // 1.4
}

// NuevoConstructor con los valores por defecto.
func NuevoConstructor(b *Base, r *Recuperador) *Constructor {
	n := TamanoParte(b)
	return &Constructor{Base: b, Rec: r, MaxSinPartes: n, PasosPorParte: n,
		UmbralProcedimiento: 0.6, UmbralConcepto: 0.6, UmbralError: 0.6, UmbralFragmento: 0.7, UmbralOpcion: 0.4,
		UmbralMinimo: 0.3, RatioDominio: 1.4}
}

// Informe dice por qué salió ese plan (para la traza).
type Informe struct {
	Ruta       string            `json:"ruta"` // procedimiento | implicito | concepto | comparacion | navegacion | errores | aclaracion | sin_evidencia | continuacion
	Motivo     string            `json:"motivo,omitempty"`
	Degradado  []string          `json:"degradado,omitempty"`
	Candidatos []tipos.Candidato `json:"candidatos,omitempty"`
}

// Construir cumple v2.Constructor.
func (c *Constructor) Construir(ctx context.Context, e tipos.Estado, cands []tipos.Candidato) (tipos.Plan, error) {
	p, _, err := c.ConstruirConInforme(ctx, e, cands)
	return p, err
}

// ConstruirConInforme arma el plan y explica la ruta. cands son los candidatos del núcleo (de mejor a
// peor); si no traen la clase que hace falta, se busca con el Recuperador propio.
func (c *Constructor) ConstruirConInforme(ctx context.Context, e tipos.Estado, cands []tipos.Candidato) (tipos.Plan, Informe, error) {
	q := e.Consulta
	if q.Original == "" {
		q.Original = e.Pregunta
	}
	var inf Informe
	var p tipos.Plan
	switch e.Tipo {
	case tipos.Procedimiento, tipos.Configuracion:
		p = c.procedural(ctx, e.Tipo, q, cands, &inf)
	case tipos.Navegacion:
		p = c.navegacion(ctx, q, cands, &inf)
	case tipos.Problema:
		p = c.problema(ctx, e, q, cands, &inf)
	case tipos.Concepto:
		p = c.concepto(ctx, q, cands, &inf)
	case tipos.Comparacion:
		p = c.comparacion(ctx, q, cands, &inf)
	case tipos.Desconocido, "":
		p = c.aclaracion(ctx, q, cands, "", &inf)
	case tipos.Social:
		return tipos.Plan{}, inf, fmt.Errorf("tipo SOCIAL: lo resuelve la ruta conversacional de V1")
	default:
		return tipos.Plan{}, inf, fmt.Errorf("tipo de respuesta desconocido %q", e.Tipo)
	}
	return p, inf, nil
}

// --- candidatos ---------------------------------------------------------------------------------------

type candEf struct {
	tipos.Candidato
	ef float64 // puntaje efectivo en [0, 1]
}

// candidatos de una clase: los del núcleo si los trae (re-puntuados a nuestra escala si vienen de otro
// buscador), si no los del Recuperador propio.
func (c *Constructor) candidatos(ctx context.Context, q tipos.Consulta, cands []tipos.Candidato, clase string, k int, inf *Informe) []candEf {
	var propios []tipos.Candidato
	for _, x := range cands {
		if x.Clase == clase {
			propios = append(propios, x)
		}
	}
	if len(propios) == 0 && c.Rec != nil {
		res, deg, err := c.Rec.Buscar(ctx, q, clase, k)
		inf.Degradado = unicos(append(inf.Degradado, deg...))
		if err == nil {
			propios = res
		}
	}
	out := make([]candEf, 0, len(propios))
	for _, x := range propios {
		ef := x.Puntaje
		if _, mio := x.Meta["cobertura"]; !mio {
			switch {
			case x.Rerank != nil:
				ef = aUnidad(*x.Rerank)
			case c.Rec != nil:
				ef = c.Rec.Puntuar(q, clase, x.ID, x.Meta["manual"])
			}
		}
		out = append(out, candEf{x, ef})
	}
	inf.Candidatos = append(inf.Candidatos, propios...)
	return out
}

// acepta: el candidato i pasa su umbral o domina (ver UmbralMinimo y RatioDominio). clave dice qué
// candidatos son «el mismo elemento» (los errores de un mismo procedimiento no compiten entre sí). Sin
// rival a la vista, el dominio exige además la mitad del camino entre el mínimo y el umbral.
func (c *Constructor) acepta(cs []candEf, i int, umbral float64, clave func(candEf) string) bool {
	x := cs[i]
	if x.ef >= umbral {
		return true
	}
	if x.ef < c.UmbralMinimo || x.Lexico <= 0 || c.RatioDominio <= 0 {
		return false
	}
	k, rival := clave(x), 0.0
	for j, y := range cs {
		if j != i && clave(y) != k && y.Lexico > rival {
			rival = y.Lexico
		}
	}
	if rival == 0 {
		return x.ef >= (c.UmbralMinimo+umbral)/2
	}
	return x.Lexico >= c.RatioDominio*rival
}

func porID(x candEf) string            { return x.ID }
func porProcedimiento(x candEf) string { return x.Meta["procedimiento"] }

func (c *Constructor) procsValidos(cs []candEf) []candEf {
	var out []candEf
	for _, x := range cs {
		if _, ok := c.Base.porProc[x.ID]; ok {
			out = append(out, x)
		}
	}
	return out
}

// ambiguo: los dos primeros empatan (puntaje y BM25) y la consulta no nombra a uno solo de ellos.
func (c *Constructor) ambiguo(q tipos.Consulta, cs []candEf) bool {
	if len(cs) < 2 || cs[1].ef < c.UmbralProcedimiento || cs[1].ef < cs[0].ef-0.05 {
		return false
	}
	if cs[0].Lexico > 0 && cs[1].Lexico < 0.95*cs[0].Lexico {
		return false
	}
	a, b := c.Base.porProc[cs[0].ID], c.Base.porProc[cs[1].ID]
	if a == nil || b == nil {
		return false
	}
	nombra := func(p *Procedimiento) bool {
		for _, x := range q.Aliases {
			if x == p.Titulo {
				return true
			}
		}
		return false
	}
	return nombra(a) == nombra(b)
}

// --- PROCEDURE / CONFIGURATION ------------------------------------------------------------------------

func (c *Constructor) procedural(ctx context.Context, tipo tipos.TipoRespuesta, q tipos.Consulta, cands []tipos.Candidato, inf *Informe) tipos.Plan {
	cs := c.procsValidos(c.candidatos(ctx, q, cands, ClaseProcedimiento, 10, inf))
	if tipo == tipos.Configuracion {
		// Primero los procedimientos de configuración que pasan el umbral.
		sort.SliceStable(cs, func(i, j int) bool {
			ci := cs[i].ef >= c.UmbralProcedimiento && c.Base.porProc[cs[i].ID].EsDeConfiguracion()
			cj := cs[j].ef >= c.UmbralProcedimiento && c.Base.porProc[cs[j].ID].EsDeConfiguracion()
			return ci && !cj
		})
	}
	if len(cs) > 0 && c.acepta(cs, 0, c.UmbralProcedimiento, porID) {
		if c.ambiguo(q, cs) {
			return c.aclaracion(ctx, q, cands, "dos procedimientos empatan", inf)
		}
		inf.Ruta = "procedimiento"
		return c.PlanProcedimiento(c.Base.porProc[cs[0].ID], tipo, 1, cs[0].ef)
	}
	if p, ok := c.implicito(ctx, tipo, q, cands, inf); ok {
		return p
	}
	if len(c.opciones(cs)) > 0 {
		return c.aclaracion(ctx, q, cands, "ningún procedimiento pasa el umbral", inf)
	}
	return c.sinEvidencia(tipo, "ningún procedimiento ni sección con pasos respalda la tarea", inf)
}

// PlanProcedimiento: los pasos en orden (subpasos dentro de su paso), cada uno con SUS fotos, sus
// fuentes y una afirmación → evidencia; prerrequisitos y verificación; fuentes por manual con páginas.
// desde: primer paso a mostrar (entrega por partes).
func (c *Constructor) PlanProcedimiento(p *Procedimiento, tipo tipos.TipoRespuesta, desde int, confianza float64) tipos.Plan {
	pl := tipos.Plan{Version: 2, Tipo: tipo, Procedimiento: &tipos.Ref{ID: p.ID, Titulo: p.Titulo, Confianza: redondear(confianza)}}
	var citas []Cita
	vis := map[Clave]bool{}
	citar := func(ids []string) {
		for _, id := range ids {
			ci, err := c.Base.CitaProcedimiento(p, id)
			if err != nil || vis[Clave{ci.ID, ci.Manual}] {
				continue
			}
			vis[Clave{ci.ID, ci.Manual}] = true
			citas = append(citas, ci)
		}
	}
	afirmar := func(texto string, ids []string, pagina, paso int) {
		if len(ids) == 0 {
			return
		}
		a := tipos.Afirmacion{Texto: texto, Fragmento: ids[0], Pagina: pagina, Paso: paso}
		if ci, err := c.Base.CitaProcedimiento(p, ids[0]); err == nil {
			a.Manual = ci.Manual
			if a.Pagina == 0 {
				a.Pagina = ci.Pagina
			}
		}
		pl.Afirmaciones = append(pl.Afirmaciones, a)
	}
	// La parte que se entrega: los prerrequisitos van con la primera y la verificación con la que termina.
	total := len(p.Pasos)
	if desde < 1 {
		desde = 1
	}
	hasta := total
	if total > c.maxSinPartes() {
		hasta = min(total, desde+c.pasosPorParte()-1)
	}
	if desde == 1 {
		for _, x := range p.Prerrequisitos {
			pl.Prerrequisitos = append(pl.Prerrequisitos, x)
			citar(x.Fuente)
			afirmar(x.Texto, x.Fuente, 0, 0)
		}
	}
	for _, s := range p.Pasos {
		pp := tipos.PasoPlan{N: s.N, ID: s.ID, Texto: textoPaso(s, ""), Pagina: s.Pagina, Fuente: append([]string(nil), s.Fuente...)}
		pp.Fotos = fotosPlan(s.Fotos)
		citar(s.Fuente)
		afirmar(s.Accion, s.Fuente, s.Pagina, s.N)
		var subs func([]tipos.Paso)
		subs = func(ss []tipos.Paso) {
			for _, sub := range ss {
				pp.Fuente = unicos(append(pp.Fuente, sub.Fuente...))
				pp.Fotos = append(pp.Fotos, fotosPlan(sub.Fotos)...)
				citar(sub.Fuente)
				afirmar(sub.Accion, sub.Fuente, sub.Pagina, s.N)
				subs(sub.Sub)
			}
		}
		subs(s.Sub)
		pl.Pasos = append(pl.Pasos, pp)
	}
	if hasta >= total {
		for _, x := range p.Verificacion {
			pl.Verificacion = append(pl.Verificacion, x)
			citar(x.Fuente)
			afirmar(x.Texto, x.Fuente, 0, 0)
		}
	}
	pl.Fuentes = ResumirFuentes(citas)
	for n := desde; n <= hasta; n++ {
		pl.PasosMostrados = append(pl.PasosMostrados, n)
	}
	slots := slotsProcedimiento(p, &pl)
	if desde == 1 {
		pl.Plantilla = "PROCEDIMIENTO_INTRO"
		pl.Intro = c.Base.Plantillas.R("PROCEDIMIENTO_INTRO", slots, OpcRelleno{Omitir: []string{"prerrequisitos", "arranque", "cita"}})
		if EsImplicito(pl.Procedimiento) {
			pl.Plantilla = "PROCEDIMIENTO_IMPLICITO"
			pl.Intro = c.Base.Plantillas.R("PROCEDIMIENTO_IMPLICITO", slots)
		}
	} else {
		pl.Plantilla = "PASOS_BLOQUE"
	}
	if hasta < total {
		slots["N_SIGUIENTE"] = strconv.Itoa(hasta + 1)
		pl.Siguiente = &tipos.Siguiente{Plantilla: "PREGUNTA_AVANCE", Paso: hasta + 1, Texto: c.Base.Plantillas.R("PREGUNTA_AVANCE", slots)}
	}
	return pl
}

// TamanoParte: pasos por parte de la entrega. Fuente única: config.max_pasos_por_bloque de
// metrin/plantillas/respuestas.yml; sin ese dato, 6 (§3). Con más pasos que esto, la entrega va por
// partes de este tamaño. El núcleo no tiene su propio tamaño: pide la parte siguiente al Continuador.
func TamanoParte(b *Base) int {
	if b != nil && b.Plantillas != nil && b.Plantillas.MaxPasosBloque > 0 {
		return b.Plantillas.MaxPasosBloque
	}
	return 6
}

// TamanoParte del constructor (el que usa en la entrega por partes).
func (c *Constructor) TamanoParte() int { return c.pasosPorParte() }

func (c *Constructor) maxSinPartes() int {
	if c.MaxSinPartes > 0 {
		return c.MaxSinPartes
	}
	return 6
}

func (c *Constructor) pasosPorParte() int {
	if c.PasosPorParte > 0 {
		return c.PasosPorParte
	}
	return 6
}

// textoPaso: la acción y, debajo, los subpasos numerados «n.m.» con sangría.
func textoPaso(s tipos.Paso, prefijo string) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(s.Accion))
	num := prefijo + strconv.Itoa(s.N)
	for _, sub := range s.Sub {
		b.WriteString("\n   " + num + "." + strconv.Itoa(sub.N) + ". ")
		b.WriteString(strings.ReplaceAll(textoPaso(sub, num+"."), "\n", "\n   "))
	}
	return b.String()
}

// fotosPlan: las fotos del paso sin el OCR (es ruidoso: sirve para buscar, no para mostrar).
func fotosPlan(fs []tipos.Foto) []tipos.Foto {
	var out []tipos.Foto
	for _, f := range fs {
		f.OCR = ""
		out = append(out, f)
	}
	return out
}

// slotsProcedimiento: los datos de las plantillas (respuestas.yml, «slots»).
func slotsProcedimiento(p *Procedimiento, pl *tipos.Plan) map[string]string {
	s := map[string]string{
		"TAREA":       minusculaInicial(p.Titulo),
		"MODULO":      p.Entidades.Modulo,
		"OBJETIVO":    p.Objetivo,
		"TOTAL_PASOS": strconv.Itoa(len(p.Pasos)),
		"CITA":        CitaTexto(pl.Fuentes),
	}
	if len(pl.PasosMostrados) > 0 {
		s["N_DESDE"] = strconv.Itoa(pl.PasosMostrados[0])
		s["N_HASTA"] = strconv.Itoa(pl.PasosMostrados[len(pl.PasosMostrados)-1])
	}
	if len(p.Prerrequisitos) > 0 {
		s["PRERREQUISITOS"] = lista(textos(p.Prerrequisitos))
	}
	if len(p.Verificacion) > 0 {
		s["VERIFICACION_FINAL"] = lista(textos(p.Verificacion))
	}
	if len(p.Fuentes) > 0 {
		s["MANUAL"] = p.Fuentes[0].Manual
		s["SECCION"] = p.Fuentes[0].Seccion
	}
	return s
}

func textos(xs []tipos.ConFuente) []string {
	var out []string
	for _, x := range xs {
		out = append(out, x.Texto)
	}
	return out
}

func lista(xs []string) string {
	var b strings.Builder
	for i, x := range xs {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("- " + strings.TrimSpace(x))
	}
	return b.String()
}

// CitaTexto: la cita para leer, agrupada por manual: «Manual de Presupuestos › 3.3 Registro del
// presupuesto (1); 3.4 Registro del subpresupuesto; Guia de Usuario de S10 Presupuestos, p. 11, 12». Con más
// de 3 secciones de un manual, las 3 primeras y «…» (la lista completa va en plan.fuentes).
func CitaTexto(fps []tipos.FuentePlan) string {
	type grupo struct {
		secs    []string
		paginas []int
	}
	var orden []string
	g := map[string]*grupo{}
	for _, f := range fps {
		x, ok := g[f.Manual]
		if !ok {
			x = &grupo{}
			g[f.Manual] = x
			orden = append(orden, f.Manual)
		}
		if f.Seccion != "" {
			x.secs = append(x.secs, f.Seccion)
		}
		for _, p := range f.Paginas {
			if !contieneInt(x.paginas, p) {
				x.paginas = append(x.paginas, p)
			}
		}
	}
	var partes []string
	for _, m := range orden {
		x := g[m]
		s := m
		if secs := unicos(x.secs); len(secs) > 0 {
			if len(secs) > 3 {
				secs = append(secs[:3], "…")
			}
			s += " › " + strings.Join(secs, "; ")
		}
		if len(x.paginas) > 0 {
			sort.Ints(x.paginas)
			ps := make([]string, len(x.paginas))
			for i, n := range x.paginas {
				ps[i] = strconv.Itoa(n)
			}
			s += ", p. " + strings.Join(ps, ", ")
		}
		partes = append(partes, s)
	}
	return strings.Join(partes, "; ")
}

// Continuar: la parte siguiente del procedimiento en curso («¿y luego?», «listo»), sin buscar. Pasado el
// último paso, el plan de cierre (PROCEDIMIENTO_FIN) con la verificación.
func (c *Constructor) Continuar(mem tipos.Memoria) (tipos.Plan, error) {
	p, ok := c.Base.Proc(mem.ProcedimientoID)
	if !ok {
		return tipos.Plan{}, fmt.Errorf("procedimiento en curso desconocido: %q", mem.ProcedimientoID)
	}
	desde := mem.PasoActual + 1
	if desde <= len(p.Pasos) {
		return c.PlanProcedimiento(p, tipos.Procedimiento, desde, 1), nil
	}
	pl := c.PlanProcedimiento(p, tipos.Procedimiento, len(p.Pasos)+1, 1)
	pl.PasosMostrados = nil
	pl.Siguiente = nil
	pl.Plantilla = "PROCEDIMIENTO_FIN"
	pl.Intro = c.Base.Plantillas.R("PROCEDIMIENTO_FIN", slotsProcedimiento(p, &pl), OpcRelleno{Omitir: []string{"oferta"}})
	return pl, nil
}

// --- procedimiento implícito (fragmentos con pasos) -----------------------------------------------------

func (c *Constructor) implicito(ctx context.Context, tipo tipos.TipoRespuesta, q tipos.Consulta, cands []tipos.Candidato, inf *Informe) (tipos.Plan, bool) {
	if c.Base.Fragmentos.Total() == 0 {
		return tipos.Plan{}, false
	}
	for _, x := range c.candidatos(ctx, q, cands, ClaseFragmento, 5, inf) {
		if x.ef < c.UmbralFragmento {
			break
		}
		manual := x.Meta["manual"]
		if manual == "" {
			if frs := c.Base.Fragmentos.PorID(x.ID); len(frs) == 1 {
				manual = frs[0].Manual
			} else {
				continue // id ambiguo sin manual: no se cita solo con el id
			}
		}
		p, ok := c.Base.procedimientoImplicito(IDImplicito(x.ID, manual))
		if !ok {
			continue
		}
		inf.Ruta = "implicito"
		inf.Motivo = "sin procedimiento YAML: pasos de la sección " + p.Fuentes[0].Seccion + " (" + manual + ")"
		return c.PlanProcedimiento(p, tipo, 1, x.ef), true
	}
	return tipos.Plan{}, false
}

// --- NAVIGATION ------------------------------------------------------------------------------------------

func (c *Constructor) navegacion(ctx context.Context, q tipos.Consulta, cands []tipos.Candidato, inf *Informe) tipos.Plan {
	cs := c.procsValidos(c.candidatos(ctx, q, cands, ClaseProcedimiento, 5, inf))
	if len(cs) > 0 && c.acepta(cs, 0, c.UmbralProcedimiento, porID) {
		p := c.Base.porProc[cs[0].ID]
		if pl, ok := c.planRuta(p, q, cs[0].ef); ok {
			inf.Ruta = "navegacion"
			// Si la pregunta nombra un término del glosario con «en_s10», va también (dónde está en S10).
			if n := c.conceptosNombrados(q); len(n) > 0 && n[0].EnS10 != "" {
				pc := c.planConcepto(n[0], tipos.Navegacion)
				pl.Concepto = pc.Concepto
				pl.Fuentes = fusionarFuentes(pl.Fuentes, pc.Fuentes)
				if len(pc.Afirmaciones) > 1 { // la de «en_s10»
					pl.Afirmaciones = append(pl.Afirmaciones, pc.Afirmaciones[1:]...)
				}
			}
			return pl
		}
	}
	// Sin ruta en un procedimiento: el «en_s10» del glosario, si lo nombra.
	if cc := c.mejorConcepto(ctx, q, cands, inf); cc != nil && cc.EnS10 != "" {
		pl := c.planConcepto(cc, tipos.Navegacion)
		pl.Plantilla = "NAVEGACION"
		inf.Ruta = "navegacion"
		inf.Motivo = "ruta del glosario (en_s10)"
		return pl
	}
	return c.sinEvidencia(tipos.Navegacion, "la ruta no está documentada en los procedimientos ni en el glosario", inf)
}

// planRuta: los pasos que dicen dónde ocurren (campo «donde», sin repetir lugar). Si ninguno lo dice,
// el primer paso cuyo menú o pantalla (un término entre ** **) nombra lo que la pregunta busca, o el
// primero de los tres primeros que nombre alguno. Sin eso, la ruta no está documentada.
func (c *Constructor) planRuta(p *Procedimiento, q tipos.Consulta, confianza float64) (tipos.Plan, bool) {
	var ruta []tipos.Paso
	var lugares []string
	visto := map[string]bool{}
	for _, s := range p.Pasos {
		d := strings.TrimSpace(s.Donde)
		if d == "" || visto[plegar(d)] {
			continue
		}
		visto[plegar(d)] = true
		ruta = append(ruta, s)
		lugares = append(lugares, d)
	}
	if len(ruta) == 0 {
		pide := map[string]bool{}
		for _, t := range terminos(q.Original) {
			pide[t] = true
		}
		for _, s := range p.Pasos {
			for _, t := range terminos(strings.Join(negritas(s.Accion), " ")) {
				if pide[t] && len(ruta) == 0 {
					ruta = []tipos.Paso{s}
				}
			}
		}
		for i := 0; len(ruta) == 0 && i < len(p.Pasos) && i < 3; i++ {
			if len(negritas(p.Pasos[i].Accion)) > 0 {
				ruta = []tipos.Paso{p.Pasos[i]}
			}
		}
	}
	if len(ruta) == 0 {
		return tipos.Plan{}, false
	}
	full := c.PlanProcedimiento(p, tipos.Navegacion, 1, confianza)
	pl := tipos.Plan{Version: 2, Tipo: tipos.Navegacion, Procedimiento: full.Procedimiento, Plantilla: "NAVEGACION"}
	var citas []Cita
	vis := map[Clave]bool{}
	for _, s := range ruta {
		pl.Pasos = append(pl.Pasos, tipos.PasoPlan{N: s.N, ID: s.ID, Texto: strings.TrimSpace(s.Accion), Fotos: fotosPlan(s.Fotos),
			Fuente: append([]string(nil), s.Fuente...), Pagina: s.Pagina})
		pl.PasosMostrados = append(pl.PasosMostrados, s.N)
		for _, id := range s.Fuente {
			if ci, err := c.Base.CitaProcedimiento(p, id); err == nil && !vis[Clave{ci.ID, ci.Manual}] {
				vis[Clave{ci.ID, ci.Manual}] = true
				citas = append(citas, ci)
			}
		}
		for _, a := range full.Afirmaciones {
			if a.Paso == s.N && a.Texto == s.Accion {
				pl.Afirmaciones = append(pl.Afirmaciones, a)
			}
		}
	}
	pl.Fuentes = ResumirFuentes(citas)
	if len(lugares) > 0 {
		slots := map[string]string{"RUTA_MENU": strings.Join(lugares, " › "), "CITA": CitaTexto(pl.Fuentes)}
		if d := c.Base.Detectar(q.Original); len(d.Entidades) > 0 {
			slots["ELEMENTO"] = d.Entidades[0]
		}
		pl.Intro = c.Base.Plantillas.R("NAVEGACION", slots, OpcRelleno{Omitir: []string{"cita", "oferta"}})
	}
	return pl, true
}

// --- TROUBLESHOOTING -------------------------------------------------------------------------------------

func (c *Constructor) problema(ctx context.Context, e tipos.Estado, q tipos.Consulta, cands []tipos.Candidato, inf *Informe) tipos.Plan {
	const maxErrores = 3
	var proc *Procedimiento
	var errs []int
	ecs := c.candidatos(ctx, q, cands, ClaseError, 10, inf)
	enCurso := e.Memoria.ProcedimientoID
	if enCurso != "" { // con un procedimiento en curso, solo cuentan sus errores
		var propios []candEf
		for _, x := range ecs {
			if x.Meta["procedimiento"] == enCurso {
				propios = append(propios, x)
			}
		}
		ecs = propios
	}
	if len(ecs) > 0 && c.acepta(ecs, 0, c.UmbralError, porProcedimiento) {
		proc = c.Base.porProc[ecs[0].Meta["procedimiento"]]
		for _, x := range ecs {
			if proc == nil || x.Meta["procedimiento"] != proc.ID || x.ef < c.UmbralMinimo || len(errs) >= maxErrores {
				continue
			}
			i, _ := strconv.Atoi(x.Meta["indice"])
			errs = append(errs, i)
		}
	}
	if proc == nil && enCurso != "" {
		proc, _ = c.Base.Proc(enCurso)
	}
	if proc == nil {
		cs := c.procsValidos(c.candidatos(ctx, q, cands, ClaseProcedimiento, 5, inf))
		if len(cs) > 0 && c.acepta(cs, 0, c.UmbralProcedimiento, porID) {
			proc = c.Base.porProc[cs[0].ID]
		}
	}
	if proc == nil {
		return c.sinEvidencia(tipos.Problema, "ningún error frecuente ni procedimiento coincide", inf)
	}
	if len(errs) == 0 {
		for i := range proc.ErroresFrecuentes {
			if i < maxErrores {
				errs = append(errs, i)
			}
		}
	}
	ref := &tipos.Ref{ID: proc.ID, Titulo: proc.Titulo, Confianza: 1}
	if len(errs) == 0 {
		pl := c.sinEvidencia(tipos.Problema, "el procedimiento «"+proc.Titulo+"» no documenta errores frecuentes", inf)
		pl.Procedimiento = ref
		return pl
	}
	pl := tipos.Plan{Version: 2, Tipo: tipos.Problema, Procedimiento: ref, Plantilla: "ERROR_FRECUENTE"}
	var citas []Cita
	vis := map[Clave]bool{}
	citar := func(ids []string) {
		for _, id := range ids {
			if ci, err := c.Base.CitaProcedimiento(proc, id); err == nil && !vis[Clave{ci.ID, ci.Manual}] {
				vis[Clave{ci.ID, ci.Manual}] = true
				citas = append(citas, ci)
			}
		}
	}
	for _, i := range errs {
		ef := proc.ErroresFrecuentes[i]
		pl.Errores = append(pl.Errores, ef)
		citar(ef.Fuente)
		if len(ef.Fuente) > 0 {
			a := tipos.Afirmacion{Texto: ef.Sintoma + " → " + ef.Solucion, Fragmento: ef.Fuente[0]}
			if ci, err := c.Base.CitaProcedimiento(proc, ef.Fuente[0]); err == nil {
				a.Manual, a.Pagina = ci.Manual, ci.Pagina
			}
			pl.Afirmaciones = append(pl.Afirmaciones, a)
		}
	}
	for _, v := range proc.Verificacion {
		pl.Verificacion = append(pl.Verificacion, v)
		citar(v.Fuente)
	}
	pl.Fuentes = ResumirFuentes(citas)
	inf.Ruta = "errores"
	return pl
}

// --- CONCEPT / COMPARISON ---------------------------------------------------------------------------------

// mejorConcepto: el nombrado en la pregunta (término o sinónimo) o el mejor candidato sobre el umbral.
func (c *Constructor) mejorConcepto(ctx context.Context, q tipos.Consulta, cands []tipos.Candidato, inf *Informe) *Concepto {
	if n := c.conceptosNombrados(q); len(n) > 0 {
		return n[0]
	}
	var cs []candEf
	for _, x := range c.candidatos(ctx, q, cands, ClaseConcepto, 5, inf) {
		if _, ok := c.Base.porConcepto[x.ID]; ok {
			cs = append(cs, x)
		}
	}
	if len(cs) > 0 && c.acepta(cs, 0, c.UmbralConcepto, porID) {
		return c.Base.porConcepto[cs[0].ID]
	}
	return nil
}

// conceptosNombrados: los conceptos cuyo término o sinónimo aparece literal en la pregunta, en orden.
func (c *Constructor) conceptosNombrados(q tipos.Consulta) []*Concepto {
	d := c.Base.Detectar(q.Original)
	type pos struct {
		c *Concepto
		i int
	}
	var ps []pos
	vis := map[string]bool{}
	t := plegar(q.Original)
	for _, a := range d.Remite {
		if a.Clase != ClaseConcepto || vis[a.ID] {
			continue
		}
		if cc, ok := c.Base.porConcepto[a.ID]; ok {
			vis[a.ID] = true
			ps = append(ps, pos{cc, indiceFrase(t, a.plegada)})
		}
	}
	sort.SliceStable(ps, func(i, j int) bool { return ps[i].i < ps[j].i })
	out := make([]*Concepto, len(ps))
	for i, p := range ps {
		out[i] = p.c
	}
	return out
}

func (c *Constructor) concepto(ctx context.Context, q tipos.Consulta, cands []tipos.Candidato, inf *Informe) tipos.Plan {
	cc := c.mejorConcepto(ctx, q, cands, inf)
	if cc == nil {
		return c.sinEvidencia(tipos.Concepto, "el término no está en el glosario", inf)
	}
	inf.Ruta = "concepto"
	pl := c.planConcepto(cc, tipos.Concepto)
	// Procedimiento relacionado para ofrecer: el del glosario o el mejor por el término.
	var ofrecido *Procedimiento
	for _, id := range cc.ProcedimientosRelacionados {
		if p, ok := c.Base.porProc[id]; ok {
			ofrecido = p
			break
		}
	}
	if ofrecido == nil && c.Rec != nil {
		qq := tipos.Consulta{Original: strings.Join(cc.Nombres(), " ")}
		if res, _, err := c.Rec.Buscar(ctx, qq, ClaseProcedimiento, 1); err == nil && len(res) > 0 && res[0].Puntaje >= c.UmbralProcedimiento {
			ofrecido = c.Base.porProc[res[0].ID]
		}
	}
	if ofrecido != nil {
		pl.Opciones = []tipos.Ref{{ID: ofrecido.ID, Titulo: ofrecido.Titulo, Confianza: 1}}
		pl.Siguiente = &tipos.Siguiente{Plantilla: "OFRECER_PROCEDIMIENTO",
			Texto: c.Base.Plantillas.R("OFRECER_PROCEDIMIENTO", map[string]string{"TAREA_SUGERIDA": minusculaInicial(ofrecido.Titulo)})}
	}
	return pl
}

// planConcepto: definición del glosario con su fuente resuelta (id, manual) y la afirmación.
func (c *Constructor) planConcepto(cc *Concepto, tipo tipos.TipoRespuesta) tipos.Plan {
	def := cc.ConceptoDef
	pl := tipos.Plan{Version: 2, Tipo: tipo, Concepto: &def, Plantilla: "CONCEPTO"}
	var citas []Cita
	for _, id := range cc.Fuente {
		if ci, err := c.Base.CitaConcepto(cc, id); err == nil {
			citas = append(citas, ci)
		}
	}
	pl.Fuentes = ResumirFuentes(citas)
	if pl.Fuentes == nil {
		pl.Fuentes = []tipos.FuentePlan{}
	}
	if len(citas) > 0 {
		pl.Afirmaciones = append(pl.Afirmaciones, tipos.Afirmacion{Texto: cc.Definicion, Fragmento: citas[0].ID, Manual: citas[0].Manual, Pagina: citas[0].Pagina})
		if cc.EnS10 != "" {
			pl.Afirmaciones = append(pl.Afirmaciones, tipos.Afirmacion{Texto: cc.EnS10, Fragmento: citas[0].ID, Manual: citas[0].Manual, Pagina: citas[0].Pagina})
		}
	}
	return pl
}

func (c *Constructor) comparacion(ctx context.Context, q tipos.Consulta, cands []tipos.Candidato, inf *Informe) tipos.Plan {
	cs := c.conceptosNombrados(q)
	if len(cs) < 2 {
		vis := map[string]bool{}
		for _, x := range cs {
			vis[x.ID] = true
		}
		for _, x := range c.candidatos(ctx, q, cands, ClaseConcepto, 5, inf) {
			if len(cs) >= 2 || x.ef < c.UmbralConcepto*0.8 {
				break
			}
			if cc, ok := c.Base.porConcepto[x.ID]; ok && !vis[x.ID] {
				vis[x.ID] = true
				cs = append(cs, cc)
			}
		}
	}
	if len(cs) == 1 {
		// «¿es lo mismo X que Y?» con X e Y del mismo concepto: el glosario los da como sinónimos.
		if fs := c.Base.formasNombradas(cs[0], q.Original); len(fs) >= 2 {
			pl := c.planConcepto(cs[0], tipos.Concepto)
			otra := fs[0]
			if plegar(otra) == plegar(cs[0].Termino) {
				otra = fs[1]
			}
			pl.Intro = "En el glosario de S10, «" + otra + "» figura como sinónimo de «" + cs[0].Termino + "»."
			inf.Ruta, inf.Motivo = "concepto", "los términos comparados son sinónimos de un mismo concepto del glosario"
			return pl
		}
	}
	if len(cs) < 2 {
		return c.sinEvidencia(tipos.Comparacion, "no hay dos términos del glosario que comparar", inf)
	}
	a, b := c.planConcepto(cs[0], tipos.Comparacion), c.planConcepto(cs[1], tipos.Comparacion)
	pl := tipos.Plan{Version: 2, Tipo: tipos.Comparacion, Concepto: a.Concepto, Conceptos: []tipos.ConceptoDef{*a.Concepto, *b.Concepto},
		Plantilla: "COMPARACION", Afirmaciones: append(a.Afirmaciones, b.Afirmaciones...)}
	pl.Fuentes = fusionarFuentes(a.Fuentes, b.Fuentes)
	inf.Ruta = "comparacion"
	return pl
}

func fusionarFuentes(a, b []tipos.FuentePlan) []tipos.FuentePlan {
	out := append([]tipos.FuentePlan{}, a...)
	for _, f := range b {
		ya := false
		for i := range out {
			if out[i].Manual == f.Manual && out[i].Seccion == f.Seccion {
				ya = true
				for _, p := range f.Paginas {
					if !contieneInt(out[i].Paginas, p) {
						out[i].Paginas = append(out[i].Paginas, p)
					}
				}
				sort.Ints(out[i].Paginas)
				for _, c := range f.Citas {
					if !contieneFuente(out[i].Citas, c) {
						out[i].Citas = append(out[i].Citas, c)
					}
				}
			}
		}
		if !ya {
			out = append(out, f)
		}
	}
	return out
}

func contieneFuente(xs []tipos.Fuente, x tipos.Fuente) bool {
	for _, y := range xs {
		if y.ID == x.ID && y.Manual == x.Manual {
			return true
		}
	}
	return false
}

func contieneInt(xs []int, x int) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// --- UNKNOWN y SIN_EVIDENCIA ---------------------------------------------------------------------------

// opciones: hasta 3 procedimientos sobre el umbral de opción, sin repetir título.
func (c *Constructor) opciones(cs []candEf) []tipos.Ref {
	var out []tipos.Ref
	vis := map[string]bool{}
	for _, x := range cs {
		if len(out) >= 3 || x.ef < c.UmbralOpcion {
			break
		}
		p := c.Base.porProc[x.ID]
		if p == nil || vis[p.Titulo] {
			continue
		}
		vis[p.Titulo] = true
		out = append(out, tipos.Ref{ID: p.ID, Titulo: p.Titulo, Confianza: redondear(x.ef)})
	}
	return out
}

func (c *Constructor) aclaracion(ctx context.Context, q tipos.Consulta, cands []tipos.Candidato, motivo string, inf *Informe) tipos.Plan {
	cs := c.procsValidos(c.candidatos(ctx, q, cands, ClaseProcedimiento, 10, inf))
	pl := tipos.Plan{Version: 2, Tipo: tipos.Desconocido, Plantilla: "ACLARAR_TAREA", Fuentes: []tipos.FuentePlan{}}
	pl.Opciones = c.opciones(cs)
	slots := map[string]string{}
	if len(pl.Opciones) > 0 {
		var ts []string
		for _, o := range pl.Opciones {
			ts = append(ts, minusculaInicial(o.Titulo))
		}
		slots["OPCIONES_TAREA"] = enumerar(ts)
	}
	pl.Siguiente = &tipos.Siguiente{Plantilla: "ACLARAR_TAREA", Texto: c.Base.Plantillas.R("ACLARAR_TAREA", slots)}
	inf.Ruta = "aclaracion"
	inf.Motivo = motivo
	return pl
}

// enumerar: «a, b o c».
func enumerar(xs []string) string {
	switch len(xs) {
	case 0:
		return ""
	case 1:
		return xs[0]
	}
	return strings.Join(xs[:len(xs)-1], ", ") + " o " + xs[len(xs)-1]
}

func (c *Constructor) sinEvidencia(t tipos.TipoRespuesta, motivo string, inf *Informe) tipos.Plan {
	inf.Ruta = "sin_evidencia"
	inf.Motivo = motivo
	return tipos.Plan{Version: 2, Tipo: t, SinEvidencia: true, Fuentes: []tipos.FuentePlan{}, Plantilla: "SIN_EVIDENCIA"}
}
