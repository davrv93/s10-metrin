package conocimiento

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"rag-go/internal/v2/tipos"
)

// Clases de documento que indexa el Recuperador.
const (
	ClaseProcedimiento = "procedimiento"
	ClaseConcepto      = "concepto"
	ClaseError         = "error"     // errores_frecuentes de los procedimientos (TROUBLESHOOTING)
	ClaseFragmento     = "fragmento" // fragmentos oficiales con pasos (procedimiento implícito)
)

// Degradaciones informadas en «degradado» (tipos.Recuperador).
const (
	DegSinVector     = "sin_vector: búsqueda léxica"
	DegSinReranker   = "sin_reranker: puntaje híbrido sin reordenar"
	DegVectorError   = "vector_error"
	DegRerankerError = "reranker_error"
	DegFragmentosLex = "fragmentos: solo léxica"
	DegSinFragmentos = "sin_indice_fragmentos"
)

// Vectorizador embebe un texto. internal/embed.Embebedor lo cumple tal cual.
type Vectorizador interface {
	Embeber(ctx context.Context, texto string) ([]float32, error)
}

// Reordenador (reranker local): un puntaje por documento frente a la consulta; mayor = más relevante.
type Reordenador interface {
	Reordenar(ctx context.Context, consulta string, documentos []string) ([]float64, error)
}

// Recuperador implementa tipos.Recuperador sobre el conocimiento procedural.
type Recuperador struct {
	base    *Base
	indices map[string]*indiceBM25

	Vector    Vectorizador // nil → solo léxica
	TopFusion int          // candidatos que salen de la fusión (30)

	// Reranker de las clases del BM25 propio (rerank_clases.go): reordena los TopRerank primeros candidatos de las
	// clases de ClasesRerank con un texto representativo de cada uno (procedimiento: título, módulo, objetivo,
	// aliases y preguntas; concepto: término, sinónimos y definición). nil → el orden léxico de siempre. Si falla o se
	// pasa del tiempo, también el orden léxico, y queda en «degradado» y en la traza (etapa rerank, «clases»).
	Reranker     Reordenador
	ClasesRerank map[string]bool // nil → ClasesRerankDefecto (procedimiento y concepto)
	TopRerank    int             // candidatos que pasan por el reranker (TopRerankDefecto)
	RunasRerank  int             // recorte del texto representativo (RunasRerankDefecto)

	// Hibrido: búsqueda híbrida (internal/busqueda) para la clase «fragmento» (hibrido.go). nil → la clase usa
	// solo el BM25 propio, como antes. Procedimientos, conceptos y errores no la usan.
	Hibrido           Hibrido
	CandidatosHibrido int // resultados que se piden a la híbrida (CandidatosHibridoDefecto)

	mu      sync.Mutex
	vecDocs map[string][][]float32

	voc     *vocabulario  // palabras de los manuales y de los YAML, para corregir faltas de la consulta
	fragDoc map[Clave]int // (id, manual) → documento del índice de fragmentos con pasos
}

var _ tipos.Recuperador = (*Recuperador)(nil)

// NuevoRecuperador indexa procedimientos, conceptos, errores frecuentes y fragmentos con pasos.
func NuevoRecuperador(b *Base) *Recuperador {
	var g *dfGlobal
	voc := nuevoVocabulario()
	if b.Fragmentos.Total() > 0 {
		g = nuevoDFGlobal()
		for _, fr := range b.Fragmentos.Todos() {
			ws := palabras(fr.Titulo + "\n" + fr.Texto)
			g.agregar(ws)
			voc.agregarPalabras(ws)
		}
	}
	r := &Recuperador{base: b, indices: map[string]*indiceBM25{}, TopFusion: 30, vecDocs: map[string][][]float32{}, voc: voc}

	var procs, errs []docBM25
	for _, p := range b.Procedimientos {
		campos := []campo{prim(p.Titulo, 3), sec(p.Objetivo, 1), sec(strings.Join(entidadesPlanas(p.Entidades), " · "), 1.5)}
		for _, a := range p.Aliases {
			campos = append(campos, prim(a, 3))
		}
		for _, q := range p.Preguntas {
			campos = append(campos, prim(q, 1))
		}
		campos = append(campos, sec(strings.Join(terminosNegrita(p), " · "), 1.5))
		procs = append(procs, nuevoDoc(p.ID, ClaseProcedimiento, campos, map[string]string{
			"titulo": p.Titulo, "modulo": p.Modulo, "configuracion": strconv.FormatBool(p.EsDeConfiguracion())}))
		for i, e := range p.ErroresFrecuentes {
			cs := []campo{prim(e.Sintoma, 2), sec(e.Solucion, 1), sec(p.Titulo, 1)}
			for _, a := range p.Aliases {
				cs = append(cs, sec(a, 1))
			}
			errs = append(errs, nuevoDoc(fmt.Sprintf("%s#error%d", p.ID, i+1), ClaseError, cs,
				map[string]string{"procedimiento": p.ID, "indice": strconv.Itoa(i), "titulo": p.Titulo}))
		}
	}
	var conceptos []docBM25
	for _, c := range b.Conceptos {
		cs := []campo{prim(c.Termino, 4), sec(c.Definicion, 1), sec(c.EnS10, 0.5)}
		for _, s := range c.Sinonimos {
			cs = append(cs, prim(s, 3))
		}
		conceptos = append(conceptos, nuevoDoc(c.ID, ClaseConcepto, cs, map[string]string{"termino": c.Termino}))
	}
	var frs []docBM25
	for _, f := range b.Fragmentos.Todos() {
		if !fragmentoConPasos(f) {
			continue
		}
		frs = append(frs, nuevoDoc(f.ID, ClaseFragmento, []campo{prim(f.Seccion, 3), sec(f.Titulo, 1), sec(f.Texto, 1)},
			map[string]string{"manual": f.Manual, "seccion": f.Seccion, "titulo": f.Titulo}))
	}
	for _, ds := range [][]docBM25{procs, errs, conceptos} {
		for _, d := range ds {
			voc.agregar(d.texto)
		}
	}
	voc.cerrar()
	r.indices[ClaseProcedimiento] = nuevoIndice(procs, g)
	r.indices[ClaseError] = nuevoIndice(errs, g)
	r.indices[ClaseConcepto] = nuevoIndice(conceptos, g)
	r.indices[ClaseFragmento] = nuevoIndice(frs, g)
	r.fragDoc = make(map[Clave]int, len(frs))
	for i, d := range r.indices[ClaseFragmento].docs {
		r.fragDoc[Clave{d.id, d.meta["manual"]}] = i
	}
	return r
}

// fragmentoConPasos: sección oficial (no marketing ni tangencial) con al menos dos pasos con texto.
func fragmentoConPasos(f *Fragmento) bool {
	if f.EsImagen() || f.Confianza != "oficial" || EsMarketing(f.Manual, f.Confianza) || EsTangencial(f.Manual, f.Confianza) {
		return false
	}
	n := 0
	for _, p := range f.Pasos {
		if strings.TrimSpace(p.Texto) != "" {
			n++
		}
	}
	return n >= 2
}

// terminosNegrita: los ** ** de todos los pasos (nombres de menús, botones y campos).
func terminosNegrita(p *Procedimiento) []string {
	var out []string
	var rec func([]tipos.Paso)
	rec = func(ps []tipos.Paso) {
		for _, s := range ps {
			out = append(out, negritas(s.Accion)...)
			rec(s.Sub)
		}
	}
	rec(p.Pasos)
	return unicos(out)
}

// consultaTerminos: raíces del usuario (peso 1, con las faltas corregidas si el vocabulario de los
// manuales lo permite) y expansiones (aliases, entidades, sinónimos de verbo).
func (r *Recuperador) consultaTerminos(c tipos.Consulta) []terminoConsulta {
	var q []terminoConsulta
	vis := map[string]float64{}
	add := func(t string, p float64, origen string) {
		if vis[t] >= p && origen == "" {
			return
		}
		if p > vis[t] {
			vis[t] = p
		}
		q = append(q, terminoConsulta{t, p, origen})
	}
	usuario := r.voc.terminosCorregidos(c.Original + " " + c.Normalizada)
	for _, t := range usuario {
		add(t, 1, "")
	}
	for _, t := range usuario {
		for _, s := range sinonimosVerbo[t] {
			add(s, 0.5, t)
		}
	}
	for _, a := range append(append([]string{}, c.Aliases...), c.Entidades...) {
		for _, t := range terminos(a) {
			add(t, 0.5, "")
		}
	}
	return q
}

func textoConsulta(c tipos.Consulta) string {
	if c.Normalizada != "" && c.Normalizada != c.Original {
		return c.Original + " | " + c.Normalizada
	}
	return c.Original
}

// Buscar devuelve hasta k candidatos de la clase, con lo que se degradó.
func (r *Recuperador) Buscar(ctx context.Context, c tipos.Consulta, clase string, k int) ([]tipos.Candidato, []string, error) {
	return r.buscar(ctx, c, clase, k, true)
}

// buscar: Buscar con el reranker de clase opcional. conRerank=false lo usan las búsquedas internas del constructor
// que no responden a la pregunta (p. ej. el procedimiento que se ofrece tras un concepto, buscado por el término).
func (r *Recuperador) buscar(ctx context.Context, c tipos.Consulta, clase string, k int, conRerank bool) ([]tipos.Candidato, []string, error) {
	ix, ok := r.indices[clase]
	if !ok {
		return nil, nil, fmt.Errorf("clase desconocida %q (procedimiento | concepto | error | fragmento)", clase)
	}
	if k <= 0 {
		k = 10
	}
	if clase == ClaseFragmento && r.base.Fragmentos.Total() == 0 {
		return nil, []string{DegSinFragmentos}, nil
	}
	if clase == ClaseFragmento && r.Hibrido != nil {
		return r.buscarFragmentosHibrido(ctx, c, ix, k)
	}
	cands, deg := r.buscarLexico(ctx, c, clase, ix, k, conRerank)
	return cands, deg, nil
}

// candLex: un candidato del BM25 propio (con su vector y su reranker, si los hubo).
type candLex struct {
	doc            int
	bm25, cub, vec float64
	rrf, puntaje   float64
	rerank         *float64
	rangoLex       int // posición antes del reranker (1…)
}

// buscarLexico: el BM25 propio de la clase (con vector y reranker de clase si los hay).
func (r *Recuperador) buscarLexico(ctx context.Context, c tipos.Consulta, clase string, ix *indiceBM25, k int, conRerank bool) ([]tipos.Candidato, []string) {
	var degradado []string
	lex := ix.buscar(r.consultaTerminos(c))
	top := r.TopFusion
	if top <= 0 {
		top = 30
	}

	// Vector (opcional): coseno contra los documentos de la clase; fusión por RRF.
	var cos map[int]float64
	switch {
	case clase == ClaseFragmento:
		degradado = append(degradado, DegFragmentosLex)
	case r.Vector == nil:
		degradado = append(degradado, DegSinVector)
	default:
		var err error
		cos, err = r.similitudes(ctx, clase, ix, textoConsulta(c))
		if err != nil {
			degradado = append(degradado, DegVectorError+": "+err.Error()+"; búsqueda léxica")
			cos = nil
		}
	}

	porDoc := map[int]*candLex{}
	for i, x := range lex {
		if i >= top {
			break
		}
		porDoc[x.doc] = &candLex{doc: x.doc, bm25: x.bm25, cub: x.cobertura, rrf: 1 / (60 + float64(i+1))}
	}
	if cos != nil {
		type dv struct {
			doc int
			v   float64
		}
		var vs []dv
		for d, v := range cos {
			vs = append(vs, dv{d, v})
		}
		sort.Slice(vs, func(a, b int) bool {
			if vs[a].v != vs[b].v {
				return vs[a].v > vs[b].v
			}
			return vs[a].doc < vs[b].doc
		})
		for i, x := range vs {
			if i >= top {
				break
			}
			cd := porDoc[x.doc]
			if cd == nil {
				cd = &candLex{doc: x.doc}
				porDoc[x.doc] = cd
			}
			cd.rrf += 1 / (60 + float64(i+1))
		}
		for d, cd := range porDoc {
			cd.vec = cos[d]
		}
	}
	cands := make([]*candLex, 0, len(porDoc))
	for _, cd := range porDoc {
		if cos != nil {
			cd.puntaje = 0.5*cd.cub + 0.5*math.Max(0, cd.vec)
		} else {
			cd.puntaje = cd.cub
		}
		cands = append(cands, cd)
	}
	sort.Slice(cands, func(a, b int) bool {
		if cos != nil && cands[a].rrf != cands[b].rrf {
			return cands[a].rrf > cands[b].rrf
		}
		if cands[a].bm25 != cands[b].bm25 {
			return cands[a].bm25 > cands[b].bm25
		}
		return ix.docs[cands[a].doc].id < ix.docs[cands[b].doc].id
	})
	if len(cands) > top {
		cands = cands[:top]
	}
	for i, cd := range cands {
		cd.rangoLex = i + 1
	}

	// Reranker de clase (opcional): reordena los primeros (rerank_clases.go).
	switch {
	case r.Reranker == nil:
		degradado = append(degradado, DegSinReranker)
	case conRerank && r.rerankActivo(clase):
		degradado = append(degradado, r.reordenarClase(ctx, c, clase, ix, cands, k)...)
	}

	if len(cands) > k {
		cands = cands[:k]
	}
	out := make([]tipos.Candidato, 0, len(cands))
	for _, cd := range cands {
		d := ix.docs[cd.doc]
		meta := map[string]string{"cobertura": strconv.FormatFloat(cd.cub, 'f', 3, 64)}
		for k, v := range d.meta {
			meta[k] = v
		}
		if cd.rerank != nil {
			meta["rango_lexico"] = strconv.Itoa(cd.rangoLex) // dónde lo tenía el orden léxico antes del reranker
		}
		out = append(out, tipos.Candidato{ID: d.id, Clase: clase, Puntaje: redondear(cd.puntaje), Lexico: redondear(cd.bm25),
			Vector: redondear(cd.vec), Rerank: cd.rerank, Meta: meta})
	}
	return out, degradado
}

// Puntuar: la cobertura léxica (en [0, 1]) de un documento concreto frente a la consulta. Sirve para
// poner en nuestra escala un candidato que llegó de otro buscador (p. ej. internal/busqueda).
func (r *Recuperador) Puntuar(c tipos.Consulta, clase, id, manual string) float64 {
	ix, ok := r.indices[clase]
	if !ok {
		return 0
	}
	for _, x := range ix.buscar(r.consultaTerminos(c)) {
		d := ix.docs[x.doc]
		if d.id == id && (manual == "" || d.meta["manual"] == "" || d.meta["manual"] == manual) {
			return x.cobertura
		}
	}
	return 0
}

// similitudes: coseno de la consulta con cada documento de la clase (vectores calculados una vez).
func (r *Recuperador) similitudes(ctx context.Context, clase string, ix *indiceBM25, consulta string) (map[int]float64, error) {
	r.mu.Lock()
	docs, ok := r.vecDocs[clase]
	if !ok {
		docs = make([][]float32, len(ix.docs))
		for i, d := range ix.docs {
			v, err := r.Vector.Embeber(ctx, d.texto)
			if err != nil {
				r.mu.Unlock()
				return nil, err
			}
			docs[i] = v
		}
		r.vecDocs[clase] = docs
	}
	r.mu.Unlock()
	q, err := r.Vector.Embeber(ctx, consulta)
	if err != nil {
		return nil, err
	}
	out := map[int]float64{}
	for i, v := range docs {
		out[i] = coseno(q, v)
	}
	return out, nil
}

func coseno(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var p, na, nb float64
	for i := range a {
		p += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return p / (math.Sqrt(na) * math.Sqrt(nb))
}

// aUnidad: un puntaje de reranker a [0, 1] (los logits pasan por una sigmoide).
func aUnidad(v float64) float64 {
	if v >= 0 && v <= 1 {
		return v
	}
	return 1 / (1 + math.Exp(-v))
}

func redondear(v float64) float64 { return math.Round(v*1000) / 1000 }
