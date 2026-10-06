package conocimiento

// Búsqueda híbrida para la clase «fragmento» (procedimiento implícito y evidencia de respaldo).
//
// internal/busqueda (BM25F con pesos de campo + vector del almacén + RRF, reranker opcional; medido en
// metrin/eval/BUSQUEDA.md) busca sobre TODOS los fragmentos de kb/. De lo que devuelve, solo son candidatos de la
// clase «fragmento» las secciones oficiales con pasos (fragmentoConPasos): las únicas con las que el constructor
// puede armar un procedimiento implícito. El resto queda en la traza.
//
// Escala y orden: el puntaje de un candidato sigue siendo el de este paquete (la cobertura léxica de la consulta, en
// [0, 1]; o el del reranker pasado a [0, 1] si lo hubo), porque los umbrales del constructor están en esa escala. El
// ORDEN es el de la búsqueda híbrida (candidatos sobre todo el corpus, con BM25F y vector); detrás va lo que solo
// encontraba la léxica propia, que sigue siendo candidato: la híbrida suma, no quita.
//
// Procedimientos, conceptos y errores frecuentes siguen con el BM25 propio (acierto@1 medido en
// TestReal_Recuperacion); la híbrida no se usa como señal para ellos (ver metrin/eval/V1_VS_V2.md).

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"rag-go/internal/busqueda"
	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

// Hibrido es la búsqueda híbrida sobre todos los fragmentos. *busqueda.Buscador la cumple tal cual.
type Hibrido interface {
	BuscarInforme(ctx context.Context, q string, k int) (busqueda.Informe, error)
}

// Degradaciones propias de la búsqueda híbrida de fragmentos.
const (
	DegHibridoError       = "fragmentos: la búsqueda híbrida falló"
	DegHibridoSinVector   = "fragmentos: sin vector (solo BM25F)"
	DegHibridoSinRerank   = "fragmentos: sin reranker (orden RRF)"
	DegHibridoRerankFallo = "fragmentos: el reranker falló"
)

// CandidatosHibridoDefecto: resultados que se piden a la búsqueda híbrida. Sobre ~8 000 fragmentos, solo unos
// cientos traen pasos: hacen falta bastantes para que queden candidatos de la clase.
const CandidatosHibridoDefecto = 50

// textoHibrido: la pregunta original y, detrás, las palabras de la normalizada que no estaban en ella (una
// referencia resuelta o la consulta compacta de la expansión). Los alias no van: con BM25F empeoran (BUSQUEDA.md §4.3).
func textoHibrido(c tipos.Consulta) string {
	q := strings.TrimSpace(c.Original)
	vis := map[string]bool{}
	for _, w := range palabras(q) {
		vis[w] = true
	}
	var extra []string
	for _, w := range palabras(c.Normalizada) {
		if !vis[w] {
			vis[w] = true
			extra = append(extra, w)
		}
	}
	if len(extra) > 0 {
		if q != "" {
			q += " "
		}
		q += strings.Join(extra, " ")
	}
	return q
}

// candHibrido: un candidato de la clase con su origen.
type candHibrido struct {
	doc     int
	lex     resultadoBM25 // léxica propia (bm25 = 0 si no salió)
	conLex  bool
	rango   int // posición en la híbrida (1…); 0 = solo de la léxica propia
	res     *busqueda.Resultado
	puntaje float64
}

// buscarFragmentosHibrido: la clase «fragmento» con la búsqueda híbrida. Si la híbrida falla, la léxica propia
// (y lo dice).
func (r *Recuperador) buscarFragmentosHibrido(ctx context.Context, c tipos.Consulta, ix *indiceBM25, k int) ([]tipos.Candidato, []string, error) {
	t0 := time.Now()
	q := textoHibrido(c)
	n := r.CandidatosHibrido
	if n <= 0 {
		n = CandidatosHibridoDefecto
	}
	inf, err := r.Hibrido.BuscarInforme(ctx, q, max(n, k))
	if err != nil {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		trazarHibridoFallo(ctx, q, err, time.Since(t0))
		cands, deg := r.buscarLexico(ctx, c, ClaseFragmento, ix, k, false)
		return cands, append([]string{DegHibridoError + " (" + traza.ResumirError(err) + "): solo léxica propia"}, deg...), nil
	}

	lex := ix.buscar(r.consultaTerminos(c))
	porDoc := make(map[int]resultadoBM25, len(lex))
	for _, x := range lex {
		porDoc[x.doc] = x
	}
	vis := map[int]bool{}
	var cs []*candHibrido
	for i := range inf.Resultados {
		res := &inf.Resultados[i]
		d, ok := r.fragDoc[Clave{res.Meta["id_original"], res.Meta["manual"]}]
		if !ok || vis[d] {
			continue // no es una sección oficial con pasos (o ya está)
		}
		x, conLex := porDoc[d]
		if !conLex && !res.Reordenado {
			continue // sin ningún término de la consulta y sin reranker que lo avale: no es evidencia
		}
		vis[d] = true
		cs = append(cs, &candHibrido{doc: d, lex: x, conLex: conLex, rango: i + 1, res: res})
	}
	top := r.TopFusion
	if top <= 0 {
		top = 30
	}
	for i, x := range lex {
		if i >= top {
			break
		}
		if !vis[x.doc] {
			vis[x.doc] = true
			cs = append(cs, &candHibrido{doc: x.doc, lex: x, conLex: true})
		}
	}
	for _, x := range cs {
		x.puntaje = x.lex.cobertura
		if x.res != nil && x.res.Reordenado {
			x.puntaje = aUnidad(x.res.Rerank)
		}
	}
	// Orden: el de la híbrida (RRF, o el reranker si lo hubo); detrás, lo que solo trajo la léxica propia, por su
	// BM25. El puntaje NO ordena: es el umbral. El constructor arma un procedimiento implícito solo si el MEJOR
	// candidato pasa su umbral (corta en el primero que no), como hacía con el orden léxico propio. Ordenar por
	// puntaje dejaría pasar cualquier sección con mucha cobertura aunque la búsqueda la ponga detrás.
	sort.SliceStable(cs, func(a, b int) bool {
		ca, cb := cs[a], cs[b]
		ra, rb := ca.rango, cb.rango
		if ra == 0 {
			ra = 1 << 30
		}
		if rb == 0 {
			rb = 1 << 30
		}
		if ra != rb {
			return ra < rb
		}
		if ca.lex.bm25 != cb.lex.bm25 {
			return ca.lex.bm25 > cb.lex.bm25
		}
		return ix.docs[ca.doc].id < ix.docs[cb.doc].id
	})
	if len(cs) > k {
		cs = cs[:k]
	}
	out := make([]tipos.Candidato, 0, len(cs))
	for _, x := range cs {
		d := ix.docs[x.doc]
		meta := map[string]string{"cobertura": strconv.FormatFloat(x.lex.cobertura, 'f', 3, 64), "origen": "lexica"}
		for k, v := range d.meta {
			meta[k] = v
		}
		cand := tipos.Candidato{ID: d.id, Clase: ClaseFragmento, Puntaje: redondear(x.puntaje), Lexico: redondear(x.lex.bm25), Meta: meta}
		if x.res != nil {
			meta["origen"] = "hibrida"
			meta["rango_hibrido"] = strconv.Itoa(x.rango)
			meta["rrf"] = strconv.FormatFloat(x.res.RRF, 'f', 5, 64)
			meta["vias"] = strings.Join(x.res.Fuentes, "+")
			cand.Vector = redondear(x.res.Vector)
			if x.res.Reordenado {
				v := x.res.Rerank
				cand.Rerank = &v
			}
		}
		out = append(out, cand)
	}
	deg := degradadoHibrido(r.Hibrido, inf)
	trazarHibrido(ctx, r.Hibrido, q, inf, out, r.fragDoc, time.Since(t0))
	return out, deg, nil
}

// degradadoHibrido: lo que la búsqueda híbrida no pudo usar.
func degradadoHibrido(h Hibrido, inf busqueda.Informe) []string {
	var out []string
	vector, reranker := capacidades(h, inf)
	switch {
	case inf.FalloVector != "":
		out = append(out, DegVectorError+": "+recortarTexto(inf.FalloVector, 80)+"; "+DegHibridoSinVector)
	case !vector:
		out = append(out, DegHibridoSinVector)
	}
	switch {
	case inf.FalloRerank != "":
		out = append(out, DegHibridoRerankFallo+": "+recortarTexto(inf.FalloRerank, 80)+"; orden RRF")
	case !reranker:
		out = append(out, DegHibridoSinRerank)
	}
	return out
}

// capacidades: si el buscador tiene vector y reranker. De un *busqueda.Buscador se sabe por su configuración; de
// otro, por el informe (corrió el vector, reordenó o falló el reranker).
func capacidades(h Hibrido, inf busqueda.Informe) (vector, reranker bool) {
	if b, ok := h.(*busqueda.Buscador); ok && b != nil {
		return b.Vector != nil, b.Reranker != nil
	}
	_, vector = inf.Tiempos["vector"]
	reranker = inf.FalloRerank != "" || (len(inf.Resultados) > 0 && inf.Resultados[0].Reordenado)
	return vector || inf.FalloVector != "", reranker
}

func recortarTexto(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "…"
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

// --- traza -----------------------------------------------------------------------------------------------

// ClaveTrazaHibrido: el dato que el recuperador deja en las cuatro etapas de búsqueda V2 (busqueda_lexica,
// busqueda_vectorial, fusion, rerank) con las cifras reales de busqueda.Informe. No cierra las etapas: la búsqueda
// puede correr dentro de una iteración del núcleo o desde el constructor, ya cerradas; el núcleo lo integra al
// cerrar la traza (internal/v2/traza.go).
const ClaveTrazaHibrido = "hibrido"

const maxCandTraza = 10

func trazarHibrido(ctx context.Context, h Hibrido, q string, inf busqueda.Informe, elegidos []tipos.Candidato, fragDoc map[Clave]int, total time.Duration) {
	v := traza.De(ctx).V2()
	if v == nil {
		return
	}
	llamadas := 1
	if x, ok := v.DatoDe(traza.EtapaV2Fusion, ClaveTrazaHibrido); ok {
		if m, ok := x.(traza.Datos); ok {
			if n, ok := m["llamadas"].(int); ok {
				llamadas = n + 1
			}
		}
	}
	vector, reranker := capacidades(h, inf)
	var opc busqueda.Opciones
	if b, ok := h.(*busqueda.Buscador); ok && b != nil {
		opc = b.Opc
	}
	ids := make([]string, 0, len(elegidos))
	for _, c := range elegidos {
		ids = append(ids, c.ID+"@"+c.Meta["manual"])
	}
	usable := func(r busqueda.Resultado) bool {
		_, ok := fragDoc[Clave{r.Meta["id_original"], r.Meta["manual"]}]
		return ok
	}
	item := func(r busqueda.Resultado, puntaje float64) map[string]any {
		id := r.Meta["id_original"]
		if id == "" {
			id = r.ID
		}
		return map[string]any{"id": traza.Recortar(id), "clase": ClaseFragmento, "manual": nuloTexto(traza.Recortar(r.Meta["manual"])),
			"puntaje": traza.Redondear(puntaje), "con_pasos": usable(r)}
	}
	comunes := func(d traza.Datos) traza.Datos {
		d["llamadas"] = llamadas
		d["consulta"] = traza.Recortar(q)
		d["de_cache"] = inf.DeCache
		return d
	}

	// Léxica (BM25F): los resultados con rango BM25, en ese orden.
	var lexicos []busqueda.Resultado
	var vectoriales []busqueda.Resultado
	for _, r := range inf.Resultados {
		if r.RangoBM25 > 0 {
			lexicos = append(lexicos, r)
		}
		if r.RangoVector > 0 {
			vectoriales = append(vectoriales, r)
		}
	}
	sort.SliceStable(lexicos, func(a, b int) bool { return lexicos[a].RangoBM25 < lexicos[b].RangoBM25 })
	sort.SliceStable(vectoriales, func(a, b int) bool { return vectoriales[a].RangoVector < vectoriales[b].RangoVector })
	topLex := make([]map[string]any, 0, maxCandTraza)
	for i, r := range lexicos {
		if i == maxCandTraza {
			break
		}
		topLex = append(topLex, item(r, r.BM25))
	}
	topVec := make([]map[string]any, 0, maxCandTraza)
	for i, r := range vectoriales {
		if i == maxCandTraza {
			break
		}
		topVec = append(topVec, item(r, r.Vector))
	}
	v.Dato(traza.EtapaV2BusquedaLexica, ClaveTrazaHibrido, comunes(traza.Datos{
		"motor": "BM25F (internal/busqueda)", "ms": msDeInforme(inf, "lexico"), "resultados": len(lexicos), "top": topLex,
	}))
	dv := traza.Datos{"activo": vector, "ms": msDeInforme(inf, "vector"), "resultados": len(vectoriales), "top": topVec, "fallo": nil}
	if inf.FalloVector != "" {
		dv["fallo"] = traza.Recortar(inf.FalloVector)
	}
	v.Dato(traza.EtapaV2BusquedaVectorial, ClaveTrazaHibrido, comunes(dv))

	// Fusión: el orden final de la híbrida (RRF; o el del reranker si lo hubo).
	cands := make([]map[string]any, 0, maxCandTraza)
	usables := 0
	for _, r := range inf.Resultados {
		if usable(r) {
			usables++
		}
	}
	for i, r := range inf.Resultados {
		if i == maxCandTraza {
			break
		}
		m := item(r, r.Score)
		m["rrf"] = traza.Redondear(r.RRF)
		m["rango_vector"] = r.RangoVector
		m["rango_bm25"] = r.RangoBM25
		cands = append(cands, m)
	}
	pesoVector := opc.PesoVectorSinRerank
	reordenado := len(inf.Resultados) > 0 && inf.Resultados[0].Reordenado
	if reordenado || pesoVector == 0 {
		pesoVector = opc.PesoVector
	}
	v.Dato(traza.EtapaV2Fusion, ClaveTrazaHibrido, comunes(traza.Datos{
		"ms": msDeInforme(inf, "fusion"), "total_ms": traza.Redondear(ms(total)), "busqueda_ms": msDeInforme(inf, "total"),
		"resultados": len(inf.Resultados), "con_pasos": usables, "rrf_k": opc.RRFK, "peso_vector": pesoVector,
		"peso_lexico": opc.PesoLexico, "degradado": inf.Degradado, "candidatos": cands, "elegidos": ids,
	}))

	// Reranker.
	dr := traza.Datos{"activo": reranker, "reordenado": reordenado, "ms": msDeInforme(inf, "rerank"), "fallo": nil,
		"top": opc.TopRerank}
	if !reranker {
		dr["motor"] = "noop (RERANK_URL vacío o «fragmento» fuera de V2_RERANK_CLASES)"
	}
	if inf.FalloRerank != "" {
		dr["fallo"] = traza.Recortar(inf.FalloRerank)
	}
	if reordenado {
		rr := make([]map[string]any, 0, maxCandTraza)
		for i, r := range inf.Resultados {
			if i == maxCandTraza || !r.Reordenado {
				break
			}
			rr = append(rr, item(r, r.Rerank))
		}
		dr["candidatos"] = rr
	}
	v.Dato(traza.EtapaV2Rerank, ClaveTrazaHibrido, comunes(dr))
}

// trazarHibridoFallo: la híbrida no devolvió nada (fallaron las dos listas).
func trazarHibridoFallo(ctx context.Context, q string, err error, total time.Duration) {
	v := traza.De(ctx).V2()
	if v == nil {
		return
	}
	v.Dato(traza.EtapaV2Fusion, ClaveTrazaHibrido, traza.Datos{"consulta": traza.Recortar(q), "error": traza.ResumirError(err),
		"total_ms": traza.Redondear(ms(total)), "llamadas": 1})
}

func msDeInforme(inf busqueda.Informe, etapa string) any {
	d, ok := inf.Tiempos[etapa]
	if !ok {
		return nil
	}
	return traza.Redondear(ms(d))
}

func nuloTexto(s string) any {
	if s == "" {
		return nil
	}
	return s
}
