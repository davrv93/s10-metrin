package main

// Extracción: convierte la respuesta de /ask (V1 o V2) en una Observacion común, que es lo único que ven
// las métricas. Así las dos versiones se califican con el mismo código; lo que cambia es de dónde sale
// cada dato:
//
//   dato                 V2 (con `plan`)                              V1 (sin `plan`: heurística)
//   tipo de respuesta    plan.tipo                                    orquestacion.tipo_consulta (tabla tipoV1)
//   procedimiento        plan.procedimiento.id                        se infiere en las métricas (fuentes + pasos)
//   pasos entregados     plan.pasos (o plan.pasos_mostrados)          líneas numeradas de `respuesta` (si no hay, viñetas)
//   foto de cada paso    plan.pasos[].fotos                           `![…](fotos/<ruta>)` debajo de la línea del paso
//   fuentes citadas      plan.pasos[].fuente + `fuentes`              `fuentes[].cita` → fragmentos de la KB con esa cita
//   ranking (Recall@k)   traza: rerank → fusion → vectorial/lexica    `fuentes` en orden + traza seleccion.descartados

import (
	"regexp"
	"strings"

	"rag-go/internal/v2/tipos"
)

// PasoObs es un paso entregado en la respuesta.
type PasoObs struct {
	Pos     int      `json:"pos"`               // orden en la respuesta (0, 1, …)
	N       int      `json:"n,omitempty"`       // V2: n de primer nivel según el plan
	ID      string   `json:"id,omitempty"`      // V2: id del paso en el plan
	Texto   string   `json:"texto"`             // texto del paso
	Fotos   []string `json:"fotos,omitempty"`   // claves de foto colocadas en ESTE paso
	Fuentes []string `json:"fuentes,omitempty"` // V2: ids que cita el paso
}

// ItemRanking es un resultado recuperado, en orden, con los fragmentos de la KB que representa.
type ItemRanking struct {
	Etiqueta string   `json:"etiqueta"`
	Claves   []string `json:"claves,omitempty"`
}

// Observacion: lo que hizo una versión en un turno.
type Observacion struct {
	Version          string        `json:"version"`
	PlanPresente     bool          `json:"plan_presente"`
	Error            string        `json:"error,omitempty"`
	MsCliente        float64       `json:"ms_cliente"`
	MsTraza          float64       `json:"ms_traza,omitempty"`
	TipoPredicho     string        `json:"tipo_predicho"`
	TipoCrudo        string        `json:"tipo_crudo"`
	Modo             string        `json:"modo"`
	Ruta             string        `json:"ruta,omitempty"`
	SinEvidencia     bool          `json:"sin_evidencia"`
	Respaldo         bool          `json:"respaldo"`
	Aclaracion       bool          `json:"aclaracion"`
	Respuesta        string        `json:"-"`
	ProcID           string        `json:"procedimiento,omitempty"`
	ConceptoID       string        `json:"concepto,omitempty"`
	Pasos            []PasoObs     `json:"pasos,omitempty"`
	EntregaParcial   bool          `json:"entrega_parcial,omitempty"`
	Citadas          []string      `json:"citadas,omitempty"` // claves (id@manual) de lo citado
	CitasSinResolver []string      `json:"citas_sin_resolver,omitempty"`
	Ranking          []ItemRanking `json:"ranking,omitempty"`
	Fotos            []string      `json:"fotos,omitempty"`    // claves de todas las fotos presentes
	FotosAvaladas    []string      `json:"-"`                  // fotos que el servidor ata a una fuente
	Negritas         []string      `json:"negritas,omitempty"` // términos entre ** ** de la respuesta
	TextoPlan        string        `json:"-"`                  // textos del plan (V2) para buscar términos
}

// Tabla V1 → tipo de respuesta de V2. Sin equivalente: informacion_directa, aclaracion, seguimiento (V1
// no distingue NAVIGATION, CONFIGURATION ni UNKNOWN): su tipo predicho queda vacío y cuenta como fallo.
var tipoV1 = map[string]string{
	"procedimiento":    "PROCEDURE",
	"concepto":         "CONCEPT",
	"problema":         "TROUBLESHOOTING",
	"comparacion":      "COMPARISON",
	"social":           "SOCIAL",
	"fuera_de_alcance": "SOCIAL",
	"metricas":         "SOCIAL",
}

var (
	reNumerada = regexp.MustCompile(`(?i)^\s*(?:[*_]{1,2})?(?:paso\s+)?(\d{1,2})\s*(?:[*_]{1,2})?\s*[.):\-–]\s*(?:[*_]{1,2})?\s*(.+)$`)
	reVineta   = regexp.MustCompile(`^\s*[-•]\s+(.+)$|^\s*\*\s+(.+)$`)
	reFoto     = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)\)`)
)

// rutaFotoMarkdown: «fotos/imagenes/x.png» o «/fotos/…» → «imagenes/x.png».
func rutaFotoMarkdown(s string) string {
	s = strings.TrimPrefix(s, "/")
	return strings.TrimPrefix(s, "fotos/")
}

// pasosDeTexto parte la respuesta en pasos (líneas numeradas; si no hay, viñetas) y ata cada foto
// markdown al paso cuya línea la precede. Devuelve además las fotos que no quedaron bajo ningún paso.
func pasosDeTexto(texto string) (pasos []PasoObs, sueltas []string) {
	lineas := strings.Split(texto, "\n")
	for _, modo := range []string{"numerada", "vineta"} {
		pasos, sueltas = nil, nil
		actual := -1
		for _, l := range lineas {
			if fs := reFoto.FindAllStringSubmatch(l, -1); len(fs) > 0 && strings.TrimSpace(reFoto.ReplaceAllString(l, "")) == "" {
				for _, f := range fs {
					k := claveFoto(rutaFotoMarkdown(f[1]), "")
					if actual >= 0 {
						pasos[actual].Fotos = append(pasos[actual].Fotos, k)
					} else {
						sueltas = append(sueltas, k)
					}
				}
				continue
			}
			var cuerpo string
			if modo == "numerada" {
				if m := reNumerada.FindStringSubmatch(l); m != nil {
					cuerpo = m[2]
				}
			} else if m := reVineta.FindStringSubmatch(l); m != nil {
				cuerpo = m[1] + m[2]
			}
			if cuerpo != "" {
				pasos = append(pasos, PasoObs{Pos: len(pasos), Texto: strings.TrimSpace(cuerpo)})
				actual = len(pasos) - 1
				continue
			}
			if strings.TrimSpace(l) == "" {
				continue
			}
			// Una línea que no es paso corta la asociación de fotos con el paso anterior… salvo que sea la
			// continuación del mismo paso (sangrada).
			if !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
				actual = -1
			}
		}
		if len(pasos) > 0 {
			return pasos, sueltas
		}
	}
	// Sin pasos: todas las fotos del texto quedan sueltas.
	for _, f := range reFoto.FindAllStringSubmatch(texto, -1) {
		sueltas = append(sueltas, claveFoto(rutaFotoMarkdown(f[1]), ""))
	}
	return nil, sueltas
}

// esRespaldoV1: el texto fijo que sale cuando no hay modelo, o una etapa marcada como respaldo.
func esRespaldo(r *respuestaAsk) bool {
	m := strings.ToLower(r.Modo + " " + r.Motivo)
	if strings.Contains(m, "no_disponible") || strings.Contains(m, "no disponible") || strings.Contains(m, "respaldo") ||
		strings.Contains(m, "segura") {
		return true
	}
	if r.Traza != nil {
		for _, e := range r.Traza.Etapas {
			if e.Estado == "respaldo" {
				return true
			}
		}
	}
	return false
}

// pideAclaracion (V1, heurística): sin pasos, sin respaldo, y la respuesta termina en pregunta.
func pideAclaracion(texto string, pasos int) bool {
	t := strings.TrimSpace(texto)
	if pasos > 0 || t == "" {
		return false
	}
	r := []rune(t)
	cola := string(r[max(0, len(r)-200):])
	return strings.Contains(cola, "?")
}

// observar arma la Observacion de un turno.
func observar(b *Base, version string, t TurnoCrudo) Observacion {
	o := Observacion{Version: version, MsCliente: t.MsCliente}
	if t.ErrorRed != "" || t.Respuesta == nil || t.Status != 200 {
		o.Error = t.ErrorRed
		if o.Error == "" {
			o.Error = "sin respuesta"
		}
		return o
	}
	r := t.Respuesta
	o.Respuesta = r.Respuesta
	o.Modo = r.Modo
	o.Ruta = r.Orquestacion.Ruta
	o.SinEvidencia = r.SinContexto
	o.Respaldo = esRespaldo(r)
	if r.Traza != nil {
		o.MsTraza = r.Traza.TotalMs
	}
	o.Negritas = negritas(r.Respuesta)

	// Fuentes de la respuesta (las dos versiones conservan `fuentes`).
	citadas := map[string]bool{}
	for _, f := range r.Fuentes {
		var frs []*Fragmento
		if f.ID != "" {
			if f.Manual != "" {
				frs = b.Resolver(claveFragmento(f.ID, f.Manual))
			}
			if len(frs) == 0 {
				frs = b.Resolver(f.ID)
			}
		}
		if len(frs) == 0 {
			frs = b.PorCita(f.Cita)
		}
		if len(frs) == 0 {
			o.CitasSinResolver = append(o.CitasSinResolver, f.Cita)
		}
		for _, fr := range frs {
			citadas[fr.Clave()] = true
		}
		for _, foto := range f.Fotos {
			k := claveFoto(foto, "")
			o.Fotos = append(o.Fotos, k)
			o.FotosAvaladas = append(o.FotosAvaladas, k)
		}
	}

	if r.Plan != nil {
		o.PlanPresente = true
		observarPlan(b, r, &o, citadas)
	} else {
		observarV1(b, r, &o, citadas)
	}
	for k := range citadas {
		o.Citadas = append(o.Citadas, k)
	}
	ordenarCadenas(o.Citadas)
	o.Fotos = unicas(o.Fotos)
	return o
}

func observarV1(b *Base, r *respuestaAsk, o *Observacion, citadas map[string]bool) {
	o.TipoCrudo = r.Orquestacion.TipoConsulta
	o.TipoPredicho = tipoV1[r.Orquestacion.TipoConsulta]
	if r.Orquestacion.Ruta == "conversacion" || r.Modo == "conversacional" {
		o.TipoPredicho = "SOCIAL"
	}
	if !o.Respaldo && !o.SinEvidencia {
		pasos, sueltas := pasosDeTexto(r.Respuesta)
		o.Pasos = pasos
		for _, p := range pasos {
			o.Fotos = append(o.Fotos, p.Fotos...)
		}
		o.Fotos = append(o.Fotos, sueltas...)
		o.FotosAvaladas = append(o.FotosAvaladas, sueltas...) // colocarFotos solo pone fotos de las fuentes
		for _, p := range pasos {
			o.FotosAvaladas = append(o.FotosAvaladas, p.Fotos...)
		}
	}
	o.Aclaracion = !o.Respaldo && !o.SinEvidencia && o.TipoPredicho != "SOCIAL" && pideAclaracion(r.Respuesta, len(o.Pasos))
	// Ranking: lo entregado (fuentes, en su orden) y después los mejores descartados de la traza.
	vistos := map[string]bool{}
	agregar := func(cita string) {
		k := normalizarCita(cita)
		if k == "" || vistos[k] {
			return
		}
		vistos[k] = true
		it := ItemRanking{Etiqueta: cita}
		for _, fr := range b.PorCita(cita) {
			it.Claves = append(it.Claves, fr.Clave())
		}
		o.Ranking = append(o.Ranking, it)
	}
	for _, f := range r.Fuentes {
		agregar(f.Cita)
	}
	if e := r.Traza.Etapa("seleccion"); e != nil {
		if l, ok := e.Datos["descartados"].([]any); ok {
			for _, x := range l {
				if m, ok := x.(map[string]any); ok {
					c, _ := m["cita"].(string)
					if c == "" {
						c, _ = m["documento"].(string)
					}
					agregar(c)
				}
			}
		}
	}
}

func observarPlan(b *Base, r *respuestaAsk, o *Observacion, citadas map[string]bool) {
	p := r.Plan
	o.TipoCrudo = string(p.Tipo)
	o.TipoPredicho = string(p.Tipo)
	if p.Procedimiento != nil {
		o.ProcID = p.Procedimiento.ID
	}
	if p.Concepto != nil {
		o.ConceptoID = p.Concepto.ID
	}
	if p.SinEvidencia {
		o.SinEvidencia = true
	}
	o.Aclaracion = p.Tipo == tipos.Desconocido || p.Plantilla == "ACLARAR_TAREA" ||
		(p.Siguiente != nil && p.Siguiente.Plantilla == "ACLARAR_TAREA")
	manuales := map[string]bool{}
	for _, f := range p.Fuentes {
		manuales[f.Manual] = true
	}
	proc := b.Procedimientos[o.ProcID]
	resolver := func(id string) []*Fragmento {
		if proc != nil {
			if frs := b.ResolverEn(id, proc.Fuentes); len(frs) == 1 {
				return frs
			}
		}
		frs := b.Resolver(id)
		if len(frs) > 1 && len(manuales) > 0 {
			var filtrados []*Fragmento
			for _, fr := range frs {
				if manuales[fr.Manual] {
					filtrados = append(filtrados, fr)
				}
			}
			if len(filtrados) > 0 {
				frs = filtrados
			}
		}
		return frs
	}
	mostrados := map[int]bool{}
	for _, n := range p.PasosMostrados {
		mostrados[n] = true
	}
	var textoPlan strings.Builder
	textoPlan.WriteString(p.Intro + "\n")
	for _, ps := range p.Pasos {
		n := nivel1DeID(ps.ID)
		if n == 0 {
			n = ps.N
		}
		textoPlan.WriteString(ps.Texto + "\n")
		if len(mostrados) > 0 && !mostrados[n] && !mostrados[ps.N] {
			o.EntregaParcial = true
			continue
		}
		po := PasoObs{Pos: len(o.Pasos), N: n, ID: ps.ID, Texto: ps.Texto, Fuentes: ps.Fuente}
		for _, f := range ps.Fotos {
			k := claveFoto(f.Ruta, f.ID)
			po.Fotos = append(po.Fotos, k)
			o.Fotos = append(o.Fotos, k)
		}
		for _, id := range ps.Fuente {
			for _, fr := range resolver(id) {
				citadas[fr.Clave()] = true
			}
		}
		o.Pasos = append(o.Pasos, po)
	}
	for _, x := range append(append([]tipos.ConFuente{}, p.Prerrequisitos...), p.Verificacion...) {
		textoPlan.WriteString(x.Texto + "\n")
		for _, id := range x.Fuente {
			for _, fr := range resolver(id) {
				citadas[fr.Clave()] = true
			}
		}
	}
	for _, e := range p.Errores {
		textoPlan.WriteString(e.Sintoma + "\n" + e.Solucion + "\n")
		for _, id := range e.Fuente {
			for _, fr := range resolver(id) {
				citadas[fr.Clave()] = true
			}
		}
	}
	if p.Concepto != nil {
		textoPlan.WriteString(p.Concepto.Definicion + "\n" + p.Concepto.EnS10 + "\n")
		for _, id := range p.Concepto.Fuente {
			for _, fr := range resolver(id) {
				citadas[fr.Clave()] = true
			}
		}
	}
	for _, a := range p.Afirmaciones {
		for _, fr := range resolver(a.Fragmento) {
			citadas[fr.Clave()] = true
		}
	}
	o.TextoPlan = textoPlan.String()
	for _, t := range negritas(o.TextoPlan) {
		o.Negritas = append(o.Negritas, t)
	}
	o.Negritas = unicas(o.Negritas)

	// Ranking: candidatos de la traza V2 (rerank → fusion → vectorial/lexica). Si no hay, lo citado en orden.
	for _, etapa := range []string{"rerank", "fusion", "vectorial", "lexica", "procedimiento"} {
		e := r.Traza.Etapa(etapa)
		if e == nil {
			continue
		}
		for _, clave := range []string{"candidatos", "top", "resultados", "seleccionados"} {
			l, ok := e.Datos[clave].([]any)
			if !ok || len(l) == 0 {
				continue
			}
			for _, x := range l {
				m, ok := x.(map[string]any)
				if !ok {
					continue
				}
				o.Ranking = append(o.Ranking, itemDeCandidato(b, m))
			}
			if len(o.Ranking) > 0 {
				return
			}
		}
	}
	vistos := map[string]bool{}
	for _, ps := range o.Pasos {
		for _, id := range ps.Fuentes {
			if vistos[id] {
				continue
			}
			vistos[id] = true
			it := ItemRanking{Etiqueta: id}
			for _, fr := range resolver(id) {
				it.Claves = append(it.Claves, fr.Clave())
			}
			o.Ranking = append(o.Ranking, it)
		}
	}
	for _, f := range r.Fuentes {
		if vistos[f.Cita] {
			continue
		}
		vistos[f.Cita] = true
		it := ItemRanking{Etiqueta: f.Cita}
		for _, fr := range b.PorCita(f.Cita) {
			it.Claves = append(it.Claves, fr.Clave())
		}
		o.Ranking = append(o.Ranking, it)
	}
}

// itemDeCandidato traduce un tipos.Candidato de la traza V2 a fragmentos de la KB: un procedimiento o un
// concepto representa a los fragmentos de su tabla de fuentes.
func itemDeCandidato(b *Base, m map[string]any) ItemRanking {
	id, _ := m["id"].(string)
	clase, _ := m["clase"].(string)
	manual, _ := m["manual"].(string)
	if meta, ok := m["meta"].(map[string]any); ok && manual == "" {
		manual, _ = meta["manual"].(string)
	}
	it := ItemRanking{Etiqueta: clase + ":" + id}
	switch {
	case clase == "procedimiento" || b.Procedimientos[id] != nil:
		if p := b.Procedimientos[id]; p != nil {
			for k := range b.FuentesProcedimiento(p) {
				it.Claves = append(it.Claves, k)
			}
		}
	case clase == "concepto":
		if c := b.BuscarConcepto(id); c != nil {
			for k := range b.FuentesConcepto(c) {
				it.Claves = append(it.Claves, k)
			}
		}
	default:
		var frs []*Fragmento
		if manual != "" {
			frs = b.Resolver(claveFragmento(id, manual))
		}
		if len(frs) == 0 {
			frs = b.Resolver(id)
		}
		for _, fr := range frs {
			it.Claves = append(it.Claves, fr.Clave())
		}
	}
	ordenarCadenas(it.Claves)
	return it
}

func unicas(xs []string) []string {
	vistos := map[string]bool{}
	var out []string
	for _, x := range xs {
		if x != "" && !vistos[x] {
			vistos[x] = true
			out = append(out, x)
		}
	}
	return out
}
