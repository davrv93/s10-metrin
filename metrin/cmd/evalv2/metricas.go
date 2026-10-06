package main

// Métricas. Cada definición está escrita aquí y se copia al informe (definiciones en informe.go). Todas se
// calculan con el mismo código para V1 y V2 a partir de la Observacion; la única diferencia es cómo se
// sabe a qué paso del procedimiento corresponde un paso entregado (ver mapearPasos).

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------------------------
// Correspondencia paso entregado ↔ paso del procedimiento

// Palabras vacías adicionales SOLO para comparar pasos entregados con el YAML: las formas de «tú» de los
// verbos de interfaz que el validador ya ignora en «usted» (V1 tutea: «pulsa», «elige»; el YAML no).
var vaciasTu = map[string]bool{"pulsa": true, "presiona": true, "elige": true, "selecciona": true, "usa": true,
	"utiliza": true, "ingresa": true, "ingrese": true, "haz": true, "dale": true}

// Negritas demasiado comunes para identificar un paso por sí solas.
var negritasGenericas = map[string]bool{"nuevo": true, "aceptar": true, "adicionar": true, "grabar": true,
	"guardar": true, "cancelar": true, "modificar": true, "eliminar": true, "buscar": true, "imprimir": true,
	"procesar": true, "cerrar": true, "salir": true, "si": true, "no": true, "ok": true, "editar": true,
	"agregar": true, "insertar": true, "copiar": true, "pegar": true}

func raicesPaso(s string) map[string]bool { return raices(s, vaciasTu) }

// cubre: ¿el texto entregado corresponde a esa unidad (paso o subpaso) del YAML? Devuelve la similitud.
//
//	regla A (raíces): comparte ≥ min(2, |raíces del paso|) raíces y ≥ 50 % de las raíces del paso.
//	regla B (menú):   contiene, como frase completa, una negrita del paso que la identifica: ≥ 4 letras, no
//	                  genérica («Aceptar», «Nuevo»…) y exclusiva de ese paso dentro del procedimiento
//	                  (`exclusivas`; «Datos Generales» en los pasos 1 y 13 no identifica a ninguno).
func cubre(texto string, u UnidadPaso, exclusivas map[string]bool) (bool, float64) {
	rs, rl := raicesPaso(u.Accion), raicesPaso(texto)
	comunes := 0
	for x := range rs {
		if rl[x] {
			comunes++
		}
	}
	sim := 0.0
	if len(rs) > 0 {
		sim = float64(comunes) / float64(len(rs))
	}
	if len(rs) > 0 && comunes >= min(2, len(rs)) && sim >= 0.5 {
		return true, sim
	}
	for _, n := range negritas(u.Accion) {
		if len([]rune(n)) >= 4 && !negritasGenericas[n] && (exclusivas == nil || exclusivas[n]) && contieneFrase(texto, n) {
			return true, math.Max(sim, 0.5)
		}
	}
	return false, sim
}

// negritasExclusivas: negritas del procedimiento que aparecen (en negrita o como frase) en una sola unidad.
func negritasExclusivas(us []UnidadPaso) map[string]bool {
	out := map[string]bool{}
	for _, u := range us {
		for _, n := range negritas(u.Accion) {
			veces := 0
			for _, v := range us {
				if contieneFrase(v.Accion, n) {
					veces++
				}
			}
			if veces == 1 {
				out[n] = true
			}
		}
	}
	return out
}

// Mapeo: para cada paso entregado, los n de primer nivel del procedimiento esperado que cubre.
type Mapeo struct {
	Cubre     [][]int     // por paso entregado
	PosDeN    map[int]int // n → posición del paso entregado que lo cubre en la alineación
	Entregado []int       // n de primer nivel entregados (ordenados)
	Desorden  bool        // no existe una alineación monótona (o, con ids, hay un paso repetido o fuera de orden)
}

// mapearPasos.
//
// V2 (paso con id): por id — «<procedimiento esperado>#n[.m]» → n; un id de otro procedimiento no cubre nada.
// Orden: los n deben crecer estrictamente en el orden en que vienen (sin repetidos).
//
// V1 (sin id): por texto con `cubre`; una línea puede cubrir varios pasos (V1 suele fundirlos). Orden: se
// busca una alineación monótona — recorriendo los pasos cubiertos de menor a mayor n, cada uno toma la primera
// línea que lo cubre en o después de la línea del paso anterior (voraz; es óptimo para esta pregunta). Si para
// algún paso no hay tal línea, el orden no se conserva.
func mapearPasos(pasos []PasoObs, proc *Procedimiento) Mapeo {
	m := Mapeo{Cubre: make([][]int, len(pasos)), PosDeN: map[int]int{}}
	if proc == nil {
		return m
	}
	us := unidades(proc)
	exclusivas := negritasExclusivas(us)
	validos := map[int]bool{}
	for _, n := range pasosNivel1(proc) {
		validos[n] = true
	}
	candidatas := map[int][]int{} // n → líneas que lo cubren, en orden
	ultimo := 0
	for i, p := range pasos {
		if p.ID != "" || p.N > 0 {
			n := 0
			if strings.HasPrefix(p.ID, proc.ID+"#") {
				n = nivel1DeID(p.ID)
			} else if p.ID == "" {
				n = p.N
			}
			if validos[n] {
				m.Cubre[i] = []int{n}
				if n <= ultimo {
					m.Desorden = true
				}
				ultimo = n
				if _, ya := m.PosDeN[n]; !ya {
					m.PosDeN[n] = i
				}
			}
			continue
		}
		vistos := map[int]bool{}
		for _, u := range us {
			if ok, _ := cubre(p.Texto, u, exclusivas); ok && !vistos[u.Nivel1] {
				vistos[u.Nivel1] = true
				m.Cubre[i] = append(m.Cubre[i], u.Nivel1)
				candidatas[u.Nivel1] = append(candidatas[u.Nivel1], i)
			}
		}
		sort.Ints(m.Cubre[i])
	}
	if len(candidatas) > 0 {
		var ns []int
		for n := range candidatas {
			ns = append(ns, n)
		}
		sort.Ints(ns)
		cur := 0
		for _, n := range ns {
			pos := -1
			for _, i := range candidatas[n] {
				if i >= cur {
					pos = i
					break
				}
			}
			if pos < 0 {
				m.Desorden = true
				pos = candidatas[n][0]
			} else {
				cur = pos
			}
			m.PosDeN[n] = pos
		}
	}
	for n := range m.PosDeN {
		m.Entregado = append(m.Entregado, n)
	}
	sort.Ints(m.Entregado)
	return m
}

// ordenConservado (condición 3): ver mapearPasos.
func ordenConservado(m Mapeo) bool { return !m.Desorden }

// ---------------------------------------------------------------------------------------------
// Recuperación: Recall@k, MRR@10, nDCG@10

// metricasRanking. Unidad relevante = cada entrada de fragmentos_relevantes (un id puede resolver a varias
// claves de la KB; basta una). Un ítem del ranking «acierta» una entrada si alguna de sus claves está entre
// las de la entrada.
//
//	Recall@k = entradas acertadas por los k primeros ítems / entradas
//	MRR@10   = 1 / posición del primer ítem que acierta alguna entrada (0 si ninguno en los 10 primeros)
//	nDCG@10  = Σ g_i / log2(i+1) ÷ Σ_{i ≤ min(|R|,10)} 1 / log2(i+1), con g_i = 1 si el ítem i acierta
//	           una entrada que ningún ítem anterior había acertado (relevancia binaria, sin dobles cuentas)
func metricasRanking(ranking []ItemRanking, relevantes [][]string) (r5, r10, mrr, ndcg float64) {
	if len(relevantes) == 0 {
		return 0, 0, 0, 0
	}
	acierta := func(it ItemRanking, entrada []string) bool {
		for _, a := range it.Claves {
			for _, b := range entrada {
				if a == b {
					return true
				}
			}
		}
		return false
	}
	cubiertas := map[int]bool{}
	dcg := 0.0
	for i := 0; i < len(ranking) && i < 10; i++ {
		nueva := false
		for j, e := range relevantes {
			if acierta(ranking[i], e) {
				if mrr == 0 {
					mrr = 1 / float64(i+1)
				}
				if !cubiertas[j] {
					cubiertas[j] = true
					nueva = true
				}
			}
		}
		if nueva {
			dcg += 1 / math.Log2(float64(i+2))
		}
		if i == 4 {
			r5 = float64(len(cubiertas)) / float64(len(relevantes))
		}
	}
	if len(ranking) < 5 {
		r5 = float64(len(cubiertas)) / float64(len(relevantes))
	}
	r10 = float64(len(cubiertas)) / float64(len(relevantes))
	idcg := 0.0
	for i := 0; i < min(len(relevantes), 10); i++ {
		idcg += 1 / math.Log2(float64(i+2))
	}
	return r5, r10, mrr, dcg / idcg
}

// ---------------------------------------------------------------------------------------------
// Evaluación de un caso

// PAS: las siete condiciones de PROCEDURAL ANSWER SUCCESS.
type PAS struct {
	C1Procedimiento bool `json:"c1_procedimiento"`
	C2Pasos         bool `json:"c2_pasos"`
	C3Orden         bool `json:"c3_orden"`
	C4SinInventar   bool `json:"c4_sin_inventar"`
	C5Fotos         bool `json:"c5_fotos"`
	C6Intencion     bool `json:"c6_intencion"`
	C7Fuentes       bool `json:"c7_fuentes"`
	Exito           bool `json:"exito"`
}

func (p PAS) Condiciones() []bool {
	return []bool{p.C1Procedimiento, p.C2Pasos, p.C3Orden, p.C4SinInventar, p.C5Fotos, p.C6Intencion, p.C7Fuentes}
}

// EvalCaso: todas las métricas de un caso (último turno) para una versión.
type EvalCaso struct {
	Caso          string        `json:"caso"`
	Categoria     string        `json:"categoria"`
	Version       string        `json:"version"`
	Observaciones []Observacion `json:"observaciones"` // una por turno

	TipoOK *bool `json:"tipo_ok,omitempty"`

	Recall5  *float64 `json:"recall5,omitempty"`
	Recall10 *float64 `json:"recall10,omitempty"`
	MRR      *float64 `json:"mrr,omitempty"`
	NDCG10   *float64 `json:"ndcg10,omitempty"`

	ProcOK     *bool `json:"procedimiento_ok,omitempty"`
	ConceptoOK *bool `json:"concepto_ok,omitempty"`

	PasosEsperadosTurno []int    `json:"pasos_esperados_turno,omitempty"`
	PasosEntregados     []int    `json:"pasos_entregados,omitempty"`
	RecallPasos         *float64 `json:"recall_pasos,omitempty"`
	OrdenOK             *bool    `json:"orden_ok,omitempty"`

	PasosTotales       int      `json:"pasos_totales"`
	PasosInventados    int      `json:"pasos_inventados"`
	PasosAjenos        int      `json:"pasos_ajenos"` // no son del procedimiento esperado, pero tienen respaldo
	NegritasInventadas []string `json:"negritas_inventadas,omitempty"`
	FotosSinEvidencia  []string `json:"fotos_sin_evidencia,omitempty"`
	Invencion          bool     `json:"invencion"`
	ConContenido       bool     `json:"con_contenido"`

	FotosEsperadas   int `json:"fotos_esperadas"`
	FotosPresentes   int `json:"fotos_presentes"`    // esperadas que aparecen en la respuesta (donde sea)
	FotosEnSuPaso    int `json:"fotos_en_su_paso"`   // esperadas colocadas en su paso correcto
	FotosAsignadas   int `json:"fotos_asignadas"`    // fotos colocadas en un paso del procedimiento esperado
	FotosAsignadasOK int `json:"fotos_asignadas_ok"` // …y que el YAML asocia a ese paso
	FotosPasoAjeno   int `json:"fotos_paso_ajeno"`   // colocadas bajo un paso que no es del procedimiento

	PAS           *PAS   `json:"pas,omitempty"`
	PASMotivo     string `json:"pas_motivo,omitempty"` // por qué no aplica
	ExitoTarea    *bool  `json:"exito_tarea,omitempty"`
	AbstencionOK  *bool  `json:"abstencion_ok,omitempty"`
	CitaRelevante bool   `json:"cita_relevante"`

	Respaldo     bool `json:"respaldo"`
	SinEvidencia bool `json:"sin_evidencia"`
	GeneracionOK bool `json:"generacion_ok"`
	Error        bool `json:"error"`
}

func pb(b bool) *bool       { return &b }
func pf(f float64) *float64 { return &f }

// evaluarCaso califica el último turno; los anteriores solo aportan contexto (y, en una continuación, el
// último paso que ya se entregó).
func evaluarCaso(b *Base, c Caso, obs []Observacion) EvalCaso {
	e := EvalCaso{Caso: c.ID, Categoria: c.Categoria, Observaciones: obs}
	if len(obs) == 0 {
		return e
	}
	o := obs[len(obs)-1]
	e.Version = o.Version
	e.Error = o.Error != ""
	e.Respaldo = o.Respaldo
	e.SinEvidencia = o.SinEvidencia
	e.GeneracionOK = o.Error == "" && strings.TrimSpace(o.Respuesta) != "" && !o.Respaldo
	e.ConContenido = o.Error == "" && !o.Respaldo && !o.SinEvidencia

	proc := b.Procedimientos[c.Proc()]
	conc := b.BuscarConcepto(c.Concepto())

	// --- clasificación
	if c.TipoEsperado != "" {
		ok := o.TipoPredicho == c.TipoEsperado || (o.TipoPredicho != "" && contiene(c.TiposAceptables, o.TipoPredicho))
		e.TipoOK = pb(ok)
	}

	// --- recuperación
	var relevantes [][]string
	clavesRelevantes := map[string]bool{}
	for _, r := range c.FragmentosRelevantes {
		var ks []string
		for _, fr := range b.Resolver(r.Clave()) {
			ks = append(ks, fr.Clave())
			clavesRelevantes[fr.Clave()] = true
		}
		if len(ks) > 0 {
			relevantes = append(relevantes, ks)
		}
	}
	if len(relevantes) > 0 {
		r5, r10, mrr, ndcg := metricasRanking(o.Ranking, relevantes)
		e.Recall5, e.Recall10, e.MRR, e.NDCG10 = pf(r5), pf(r10), pf(mrr), pf(ndcg)
	}

	// --- evidencia disponible para juzgar invenciones
	fuentesProc := map[string]bool{}
	if proc != nil {
		fuentesProc = b.FuentesProcedimiento(proc)
	}
	fuentesConc := map[string]bool{}
	if conc != nil {
		fuentesConc = b.FuentesConcepto(conc)
	}
	citadas := map[string]bool{}
	refRaices := map[string]bool{}
	var refTexto strings.Builder
	fotosEvidencia := map[string]bool{}
	for _, k := range o.Citadas {
		citadas[k] = true
		if fr := b.porClave[k]; fr != nil {
			refTexto.WriteString(fr.Texto + "\n")
			for _, f := range fr.Fotos {
				fotosEvidencia[claveFoto(f, "")] = true
			}
		}
		if fuentesProc[k] || fuentesConc[k] || clavesRelevantes[k] {
			e.CitaRelevante = true
		}
	}
	// El plan de V2 trae pasos de un YAML validado: su texto también es evidencia de sus negritas.
	if pv := b.Procedimientos[o.ProcID]; pv != nil {
		for _, u := range unidades(pv) {
			refTexto.WriteString(u.Accion + "\n")
		}
		for _, x := range pv.Errores {
			refTexto.WriteString(x.Sintoma + "\n" + x.Solucion + "\n")
		}
	}
	for k := range raices(refTexto.String()) {
		refRaices[k] = true
	}
	for _, f := range o.FotosAvaladas {
		fotosEvidencia[f] = true
	}
	for _, p := range b.Procedimientos {
		for _, fs := range fotosPorPaso(p) {
			for _, f := range fs {
				fotosEvidencia[f] = true
			}
		}
	}

	// --- pasos: correspondencia, orden, invenciones
	m := mapearPasos(o.Pasos, proc)
	e.PasosTotales = len(o.Pasos)
	e.PasosEntregados = m.Entregado
	refNorm := normalizar(refTexto.String())
	for i, p := range o.Pasos {
		if len(m.Cubre[i]) > 0 {
			continue
		}
		if soporteLexico(p.Texto, refRaices) {
			e.PasosAjenos++
		} else {
			e.PasosInventados++
		}
	}
	for _, n := range o.Negritas {
		if !strings.Contains(refNorm, n) {
			e.NegritasInventadas = append(e.NegritasInventadas, n)
		}
	}
	for _, f := range o.Fotos {
		if !fotosEvidencia[f] {
			e.FotosSinEvidencia = append(e.FotosSinEvidencia, f)
		}
	}
	e.Invencion = e.PasosInventados > 0 || len(e.NegritasInventadas) > 0 || len(e.FotosSinEvidencia) > 0

	// --- procedimiento y concepto identificados
	if proc != nil {
		cubreAlguno := len(m.Entregado) > 0
		citaProc := false
		for k := range citadas {
			if fuentesProc[k] {
				citaProc = true
				break
			}
		}
		// V1: cita la tabla `fuentes` del YAML y, si se pidió un procedimiento, su texto cubre ≥ 1 paso.
		// En NAVIGATION / TROUBLESHOOTING / CONFIGURATION basta la cita (no se esperan pasos).
		if o.PlanPresente {
			e.ProcOK = pb(o.ProcID == proc.ID)
		} else {
			e.ProcOK = pb(o.Error == "" && citaProc && (c.TipoEsperado != "PROCEDURE" || cubreAlguno))
		}
	}
	if c.Concepto() != "" {
		switch {
		case o.PlanPresente && conc != nil:
			e.ConceptoOK = pb(o.ConceptoID == conc.ID)
		case o.PlanPresente && o.ConceptoID != "":
			e.ConceptoOK = pb(normalizar(o.ConceptoID) == normalizar(c.Concepto()) ||
				strings.HasSuffix(normalizar(o.ConceptoID), "."+normalizar(c.Concepto())))
		default:
			citaConc := false
			for k := range citadas {
				if fuentesConc[k] || clavesRelevantes[k] {
					citaConc = true
				}
			}
			e.ConceptoOK = pb(e.ConContenido && citaConc)
		}
	}

	// --- pasos esperados en este turno (continuación: los que siguen al último entregado antes)
	// Solo una respuesta PROCEDURE se califica por pasos y fotos; en los demás tipos el procedimiento es el
	// ancla de la evidencia (citas, errores frecuentes), no una lista de pasos que entregar.
	var esperados []int
	if proc != nil && c.TipoEsperado == "PROCEDURE" {
		esperados = c.PasosEsperados
		if esperados == nil {
			esperados = pasosNivel1(proc)
		}
		esperados = append([]int(nil), esperados...)
		sort.Ints(esperados)
		if c.Continuacion && len(obs) > 1 {
			prev := mapearPasos(obs[len(obs)-2].Pasos, proc)
			maxPrev := 0
			for _, n := range prev.Entregado {
				maxPrev = max(maxPrev, n)
			}
			var resto []int
			for _, n := range esperados {
				if n > maxPrev {
					resto = append(resto, n)
				}
			}
			esperados = resto
		}
		// Entrega por partes declarada por el plan (V2): la parte mostrada debe ser el comienzo de lo
		// esperado y tener al menos min(|esperados|, ParteMinima) pasos.
		if o.EntregaParcial && len(esperados) > 0 {
			k := min(len(esperados), max(len(o.Pasos), ParteMinima))
			esperados = esperados[:k]
		}
		e.PasosEsperadosTurno = esperados
		if len(esperados) > 0 {
			hit := 0
			for _, n := range esperados {
				if _, ok := m.PosDeN[n]; ok {
					hit++
				}
			}
			e.RecallPasos = pf(float64(hit) / float64(len(esperados)))
			e.OrdenOK = pb(ordenConservado(m))
		}
	}

	// --- fotos
	fotosRespuesta := map[string]bool{}
	for _, f := range o.Fotos {
		fotosRespuesta[f] = true
	}
	var fpp map[int][]string
	if proc != nil && c.TipoEsperado == "PROCEDURE" {
		fpp = fotosPorPaso(proc)
		setEsperados := map[int]bool{}
		for _, n := range esperados {
			setEsperados[n] = true
		}
		for ns, fs := range c.FotosEsperadas {
			n, _ := strconv.Atoi(ns)
			if !setEsperados[n] {
				continue
			}
			for _, f := range fs {
				e.FotosEsperadas++
				if fotosRespuesta[f] {
					e.FotosPresentes++
				}
				if pos, ok := m.PosDeN[n]; ok && contiene(o.Pasos[pos].Fotos, f) {
					e.FotosEnSuPaso++
				} else {
					// la foto pudo quedar en otra línea que también cubre n
					for i, p := range o.Pasos {
						if contiene(m.Cubre[i], n) && contiene(p.Fotos, f) {
							e.FotosEnSuPaso++
							break
						}
					}
				}
			}
		}
		for i, p := range o.Pasos {
			for _, f := range p.Fotos {
				if len(m.Cubre[i]) == 0 {
					e.FotosPasoAjeno++
					continue
				}
				e.FotosAsignadas++
				for _, n := range m.Cubre[i] {
					if contiene(fpp[n], f) {
						e.FotosAsignadasOK++
						break
					}
				}
			}
		}
	}

	// --- PROCEDURAL ANSWER SUCCESS
	aplicaPAS := c.TipoEsperado == "PROCEDURE" && proc != nil && !c.SinEvidenciaEsperada
	switch {
	case !aplicaPAS && c.TipoEsperado == "PROCEDURE" && proc == nil && !c.SinEvidenciaEsperada:
		e.PASMotivo = "procedimiento implícito (sin YAML)"
	case aplicaPAS && len(esperados) == 0:
		e.PASMotivo = "continuación sin pasos pendientes"
		aplicaPAS = false
	}
	if aplicaPAS {
		var p PAS
		p.C1Procedimiento = e.ProcOK != nil && *e.ProcOK
		p.C2Pasos = e.RecallPasos != nil && *e.RecallPasos == 1
		p.C3Orden = e.OrdenOK != nil && *e.OrdenOK && len(m.Entregado) > 0
		p.C4SinInventar = e.PasosInventados == 0 && len(e.NegritasInventadas) == 0
		// C5: ninguna foto mal asignada y, si los pasos entregados tienen fotos esperadas, al menos la
		// mitad está en su paso.
		esperadasEntregadas := 0
		for ns, fs := range c.FotosEsperadas {
			n, _ := strconv.Atoi(ns)
			if _, ok := m.PosDeN[n]; ok && contiene(esperados, n) {
				esperadasEntregadas += len(fs)
			}
		}
		// C3, C5 y C7 piden ≥ 1 paso del procedimiento entregado: sin pasos no hay fotos que asignar ni pasos que
		// respaldar, y darlas por cumplidas «en vacío» inflaría su tasa (C4, «no inventa», sí se cumple sin pasos).
		hayPasos := len(m.Entregado) > 0
		p.C5Fotos = hayPasos && e.FotosAsignadas == e.FotosAsignadasOK && len(e.FotosSinEvidencia) == 0 &&
			(esperadasEntregadas == 0 || float64(e.FotosEnSuPaso) >= 0.5*float64(esperadasEntregadas))
		p.C6Intencion = o.Error == "" && !o.Respaldo && !o.SinEvidencia && len(o.Pasos) > 0 &&
			(o.TipoPredicho == "PROCEDURE" || contiene(c.TiposAceptables, o.TipoPredicho))
		p.C7Fuentes = hayPasos && pasosRespaldados(b, proc, o, m, fuentesProc, refRaices)
		p.Exito = p.C1Procedimiento && p.C2Pasos && p.C3Orden && p.C4SinInventar && p.C5Fotos && p.C6Intencion && p.C7Fuentes
		e.PAS = &p
	}

	// --- éxito de la tarea por categoría y abstención
	terminos := terminosPresentes(c.TerminosEsperados, o.Respuesta+"\n"+o.TextoPlan)
	limpio := e.ConContenido && !e.Invencion
	aclaraBien := o.Aclaracion && !e.Invencion && o.Error == "" && contiene(c.TiposAceptables, "UNKNOWN")
	switch {
	case c.SinEvidenciaEsperada:
		ok := o.Error == "" && (o.SinEvidencia || o.Aclaracion) && len(o.Pasos) == 0 && !e.Invencion
		e.AbstencionOK = pb(ok)
		e.ExitoTarea = pb(ok)
	case e.PAS != nil:
		e.ExitoTarea = pb(e.PAS.Exito || aclaraBien)
	case c.TipoEsperado == "UNKNOWN":
		ok := (o.Aclaracion && !e.Invencion && o.Error == "") ||
			(limpio && o.TipoPredicho != "UNKNOWN" && contiene(c.TiposAceptables, o.TipoPredicho) && e.CitaRelevante)
		e.ExitoTarea = pb(ok)
	case c.TipoEsperado == "SOCIAL":
		e.ExitoTarea = pb(o.Error == "" && o.TipoPredicho == "SOCIAL" && len(o.Pasos) == 0)
	case c.TipoEsperado == "CONCEPT":
		e.ExitoTarea = pb(aclaraBien || (limpio && *e.TipoOK && e.ConceptoOK != nil && *e.ConceptoOK && terminos))
	case c.TipoEsperado == "PROCEDURE": // implícito: sin YAML
		e.ExitoTarea = pb(aclaraBien || (limpio && *e.TipoOK && len(o.Pasos) > 0 && e.CitaRelevante && terminos))
	default: // NAVIGATION, TROUBLESHOOTING, CONFIGURATION, COMPARISON
		e.ExitoTarea = pb(aclaraBien || (limpio && *e.TipoOK && e.CitaRelevante && terminos))
	}
	return e
}

// ParteMinima: tamaño mínimo de la primera parte cuando el plan entrega por partes (§3: «por partes si
// hay más de 6 pasos»). Una parte de 1–2 pasos no cuenta como «trae los pasos correctos».
const ParteMinima = 3

// pasosRespaldados (condición 7): lo citado incluye ≥ 1 fragmento de la tabla `fuentes` del procedimiento y
// cada paso entregado que corresponde a un paso esperado tiene respaldo: V2 → su `fuente` incluye una del
// paso del YAML, o su texto tiene soporte léxico en lo citado; V1 → su texto tiene soporte léxico en el
// texto de los fragmentos citados (V1 cita a nivel de respuesta, no de paso).
func pasosRespaldados(b *Base, proc *Procedimiento, o Observacion, m Mapeo, fuentesProc, refRaices map[string]bool) bool {
	citaProc := false
	for _, k := range o.Citadas {
		if fuentesProc[k] {
			citaProc = true
			break
		}
	}
	if !citaProc {
		return false
	}
	fuentesDeN := map[int]map[string]bool{}
	for _, u := range unidades(proc) {
		if fuentesDeN[u.Nivel1] == nil {
			fuentesDeN[u.Nivel1] = map[string]bool{}
		}
		for _, id := range u.Fuente {
			for _, fr := range b.ResolverEn(id, proc.Fuentes) {
				fuentesDeN[u.Nivel1][fr.Clave()] = true
			}
		}
	}
	for i, p := range o.Pasos {
		if len(m.Cubre[i]) == 0 {
			continue
		}
		ok := false
		if len(p.Fuentes) > 0 {
			for _, id := range p.Fuentes {
				for _, fr := range b.ResolverEn(id, proc.Fuentes) {
					for _, n := range m.Cubre[i] {
						if fuentesDeN[n][fr.Clave()] {
							ok = true
						}
					}
				}
			}
		}
		if !ok && soporteLexico(p.Texto, refRaices) {
			ok = true
		}
		if !ok {
			return false
		}
	}
	return true
}

// terminosPresentes: cada término (o una de sus alternativas «a|b|c») aparece como frase completa,
// sin distinguir mayúsculas ni tildes.
func terminosPresentes(terminos []string, texto string) bool {
	for _, t := range terminos {
		ok := false
		for _, alt := range strings.Split(t, "|") {
			if contieneFrase(texto, alt) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func ordenarCadenas(xs []string) { sort.Strings(xs) }

// ---------------------------------------------------------------------------------------------
// Agregados

// Agregado: métricas de un conjunto de casos (global o una categoría) para una versión.
type Agregado struct {
	Casos  int `json:"casos"`
	Turnos int `json:"turnos"`

	ClasificacionAcc *Proporcion    `json:"clasificacion"`
	ClasifComparable *Proporcion    `json:"clasificacion_comparable"`
	Recall5          *Media         `json:"recall5"`
	Recall10         *Media         `json:"recall10"`
	MRR              *Media         `json:"mrr"`
	NDCG10           *Media         `json:"ndcg10"`
	Procedimiento    *Proporcion    `json:"acierto_procedimiento"`
	Concepto         *Proporcion    `json:"acierto_concepto"`
	RecallPasos      *Media         `json:"recall_pasos"`
	RecallImagenes   *Proporcion    `json:"recall_imagenes"`
	PasoFoto         *Proporcion    `json:"exactitud_paso_foto"`
	Invencion        *Proporcion    `json:"tasa_invencion"`
	InvencionPasos   *Proporcion    `json:"tasa_invencion_pasos"`
	Respaldo         *Proporcion    `json:"tasa_respaldo"`
	SinEvidencia     *Proporcion    `json:"tasa_sin_evidencia"`
	Abstencion       *Proporcion    `json:"abstencion_correcta"`
	FalsaAbstencion  *Proporcion    `json:"falsa_abstencion"`
	Generacion       *Proporcion    `json:"exito_generacion"`
	ExitoTarea       *Proporcion    `json:"exito_tarea"`
	PAS              *Proporcion    `json:"pas"`
	PASCondiciones   [7]*Proporcion `json:"pas_condiciones"`
	LatP50, LatP95   float64        `json:"-"`
	Latencia         Latencias      `json:"latencia"`
	Errores          int            `json:"errores"`
}

type Latencias struct {
	ClienteP50 float64 `json:"cliente_p50_ms"`
	ClienteP95 float64 `json:"cliente_p95_ms"`
	TrazaP50   float64 `json:"traza_p50_ms"`
	TrazaP95   float64 `json:"traza_p95_ms"`
}

type Proporcion struct {
	Si    int     `json:"si"`
	Total int     `json:"total"`
	Valor float64 `json:"valor"`
}

type Media struct {
	Suma  float64 `json:"-"`
	N     int     `json:"n"`
	Valor float64 `json:"valor"`
}

func (p *Proporcion) add(si bool) {
	p.Total++
	if si {
		p.Si++
	}
	p.Valor = float64(p.Si) / float64(p.Total)
}

func (p *Proporcion) addN(si, total int) {
	p.Si += si
	p.Total += total
	if p.Total > 0 {
		p.Valor = float64(p.Si) / float64(p.Total)
	}
}

func (m *Media) add(v float64) {
	m.Suma += v
	m.N++
	m.Valor = m.Suma / float64(m.N)
}

// Tipos que V1 sabe expresar: la «clasificación comparable» solo cuenta casos con estos tipos esperados.
var tiposComparables = map[string]bool{"CONCEPT": true, "PROCEDURE": true, "TROUBLESHOOTING": true, "COMPARISON": true, "SOCIAL": true}

func agregar(evals []EvalCaso, casos map[string]Caso) Agregado {
	a := Agregado{}
	nuevaP := func() *Proporcion { return &Proporcion{} }
	nuevaM := func() *Media { return &Media{} }
	a.ClasificacionAcc, a.ClasifComparable, a.Procedimiento, a.Concepto = nuevaP(), nuevaP(), nuevaP(), nuevaP()
	a.Recall5, a.Recall10, a.MRR, a.NDCG10, a.RecallPasos = nuevaM(), nuevaM(), nuevaM(), nuevaM(), nuevaM()
	a.RecallImagenes, a.PasoFoto, a.Invencion, a.InvencionPasos = nuevaP(), nuevaP(), nuevaP(), nuevaP()
	a.Respaldo, a.SinEvidencia, a.Abstencion, a.FalsaAbstencion = nuevaP(), nuevaP(), nuevaP(), nuevaP()
	a.Generacion, a.ExitoTarea, a.PAS = nuevaP(), nuevaP(), nuevaP()
	for i := range a.PASCondiciones {
		a.PASCondiciones[i] = nuevaP()
	}
	var latC, latT []float64
	for _, e := range evals {
		c := casos[e.Caso]
		a.Casos++
		for _, o := range e.Observaciones {
			a.Turnos++
			if o.Error != "" {
				a.Errores++
			}
			latC = append(latC, o.MsCliente)
			if o.MsTraza > 0 {
				latT = append(latT, o.MsTraza)
			}
			a.Respaldo.add(o.Respaldo)
			a.SinEvidencia.add(o.SinEvidencia)
			a.Generacion.add(o.Error == "" && strings.TrimSpace(o.Respuesta) != "" && !o.Respaldo)
		}
		if e.TipoOK != nil {
			a.ClasificacionAcc.add(*e.TipoOK)
			if tiposComparables[c.TipoEsperado] {
				a.ClasifComparable.add(*e.TipoOK)
			}
		}
		if e.Recall5 != nil {
			a.Recall5.add(*e.Recall5)
			a.Recall10.add(*e.Recall10)
			a.MRR.add(*e.MRR)
			a.NDCG10.add(*e.NDCG10)
		}
		if e.ProcOK != nil {
			a.Procedimiento.add(*e.ProcOK)
		}
		if e.ConceptoOK != nil {
			a.Concepto.add(*e.ConceptoOK)
		}
		if e.RecallPasos != nil {
			a.RecallPasos.add(*e.RecallPasos)
		}
		if e.FotosEsperadas > 0 {
			a.RecallImagenes.addN(e.FotosPresentes, e.FotosEsperadas)
		}
		if e.FotosAsignadas > 0 {
			a.PasoFoto.addN(e.FotosAsignadasOK, e.FotosAsignadas)
		}
		if e.ConContenido {
			a.Invencion.add(e.Invencion)
		}
		if e.PasosTotales > 0 {
			a.InvencionPasos.addN(e.PasosInventados, e.PasosTotales)
		}
		if e.AbstencionOK != nil {
			a.Abstencion.add(*e.AbstencionOK)
		} else if !e.Error && c.TipoEsperado != "UNKNOWN" && c.TipoEsperado != "SOCIAL" {
			a.FalsaAbstencion.add(e.SinEvidencia)
		}
		if e.ExitoTarea != nil {
			a.ExitoTarea.add(*e.ExitoTarea)
		}
		if e.PAS != nil {
			a.PAS.add(e.PAS.Exito)
			for i, v := range e.PAS.Condiciones() {
				a.PASCondiciones[i].add(v)
			}
		}
	}
	a.Latencia = Latencias{
		ClienteP50: percentil(latC, 50), ClienteP95: percentil(latC, 95),
		TrazaP50: percentil(latT, 50), TrazaP95: percentil(latT, 95),
	}
	return a
}

// percentil por el método del rango más cercano (nearest-rank): el valor en la posición ⌈p/100 · n⌉.
func percentil(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	k := int(math.Ceil(p / 100 * float64(len(s))))
	k = max(1, min(k, len(s)))
	return s[k-1]
}
