package conocimiento

import (
	"math"
	"sort"
)

// BM25 mínimo propio. internal/busqueda (BM25 + vector + RRF, de otro agente) todavía no tiene una API
// estable; cuando la tenga, se enchufa detrás de tipos.Recuperador sin cambiar a quien llama.
//
// Detalles:
//   - Campos con peso: el texto de cada campo suma tf × peso (título y aliases pesan más que el objetivo).
//   - IDF global: se calcula sobre TODO el corpus de fragmentos (el idioma de los manuales), no sobre los
//     pocos procedimientos. Así un término del dominio que ningún procedimiento cubre («metrado» sin
//     procedimiento de metrados) baja la cobertura, y una palabra desconocida (una falta, un relleno) pesa
//     poco. Sin corpus, el IDF es el del propio índice.
//   - Cobertura: además del BM25 crudo (para ordenar), cada resultado lleva la fracción del peso IDF de la
//     consulta que el documento cubre, en [0, 1], en todo el documento y en sus campos PRIMARIOS (título,
//     aliases, preguntas; término y sinónimos; síntoma; sección). El «Puntaje» léxico es la media de las
//     dos: un término que solo sale de pasada en una definición larga no basta para elegir ese documento.

const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

type campo struct {
	texto    string
	peso     float64
	primario bool
}

func prim(t string, w float64) campo { return campo{t, w, true} }
func sec(t string, w float64) campo  { return campo{t, w, false} }

type docBM25 struct {
	id    string
	clase string
	tf    map[string]float64
	prim  map[string]bool // raíces de los campos primarios
	largo float64
	texto string // lo que ve el reranker/vector
	meta  map[string]string
}

type indiceBM25 struct {
	docs    []docBM25
	df      map[string]int
	largoMd float64
	idf     func(t string) float64
}

// dfGlobal: frecuencia de documento de cada raíz en un corpus grande (los fragmentos).
type dfGlobal struct {
	n   int
	df  map[string]int
	vis map[string]bool // se reutiliza entre documentos
}

func nuevoDFGlobal() *dfGlobal { return &dfGlobal{df: map[string]int{}, vis: map[string]bool{}} }

// agregar suma un documento (sus palabras ya plegadas, como las da palabras()).
func (g *dfGlobal) agregar(ws []string) {
	g.n++
	clear(g.vis)
	for _, w := range ws {
		if len(w) < 2 || vacias[w] {
			continue
		}
		t := raiz(w)
		if !g.vis[t] {
			g.vis[t] = true
			g.df[t]++
		}
	}
}

func idfDe(n, df int) float64 {
	return math.Log(1 + (float64(n)-float64(df)+0.5)/(float64(df)+0.5))
}

func nuevoIndice(docs []docBM25, g *dfGlobal) *indiceBM25 {
	ix := &indiceBM25{docs: docs, df: map[string]int{}}
	total := 0.0
	for _, d := range docs {
		total += d.largo
		for t := range d.tf {
			ix.df[t]++
		}
	}
	if len(docs) > 0 {
		ix.largoMd = total / float64(len(docs))
	}
	n := len(docs)
	if g != nil && g.n > 0 {
		ix.idf = func(t string) float64 {
			df := g.df[t]
			if df == 0 {
				df = ix.df[t] // término de los YAML que no está en los manuales
			}
			return idfDe(g.n, df)
		}
	} else {
		ix.idf = func(t string) float64 { return idfDe(n, ix.df[t]) }
	}
	return ix
}

func nuevoDoc(id, clase string, campos []campo, meta map[string]string) docBM25 {
	d := docBM25{id: id, clase: clase, tf: map[string]float64{}, prim: map[string]bool{}, meta: meta}
	var txt []byte
	for _, c := range campos {
		ts := terminos(c.texto)
		for _, t := range ts {
			d.tf[t] += c.peso
			if c.primario {
				d.prim[t] = true
			}
		}
		d.largo += float64(len(ts)) * c.peso
		if c.texto != "" && c.peso >= 1 {
			txt = append(txt, c.texto...)
			txt = append(txt, '\n')
		}
	}
	d.texto = string(txt)
	return d
}

// terminoConsulta: raíz con su peso (1 la del usuario; menos los sinónimos y aliases añadidos). origen:
// la raíz del usuario que un sinónimo expande («graba» ← «guard»): si el documento trae el sinónimo,
// esa raíz cuenta como cubierta (al 80 %).
type terminoConsulta struct {
	t      string
	peso   float64
	origen string
}

const coberturaSinonimo = 0.8

type resultadoBM25 struct {
	doc       int
	bm25      float64
	cobertura float64 // (todo el documento + campos primarios) / 2
}

func (ix *indiceBM25) buscar(q []terminoConsulta) []resultadoBM25 {
	if len(ix.docs) == 0 || len(q) == 0 {
		return nil
	}
	// Peso de cada raíz (el mayor) y sinónimos de cada raíz del usuario.
	pesos := map[string]float64{}
	sinon := map[string][]string{}
	for _, x := range q {
		if x.peso > pesos[x.t] {
			pesos[x.t] = x.peso
		}
		if x.origen != "" {
			sinon[x.origen] = append(sinon[x.origen], x.t)
		}
	}
	// La cobertura se mide sobre las raíces del usuario (peso 1).
	totalUsuario := 0.0
	for t, p := range pesos {
		if p >= 1 {
			totalUsuario += ix.idf(t)
		}
	}
	var out []resultadoBM25
	for i, d := range ix.docs {
		s, cub, cubP := 0.0, 0.0, 0.0
		fuerza := func(t string) (float64, float64) { // en todo el documento, en sus campos primarios
			tf := d.tf[t]
			if tf == 0 {
				return 0, 0
			}
			norm := tf * (bm25K1 + 1) / (tf + bm25K1*(1-bm25B+bm25B*d.largo/max(ix.largoMd, 1)))
			prim := 0.0
			if d.prim[t] {
				prim = 1
			}
			return norm, prim
		}
		for t, p := range pesos {
			norm, _ := fuerza(t)
			s += p * ix.idf(t) * norm
		}
		if s <= 0 {
			continue
		}
		for t, p := range pesos {
			if p < 1 {
				continue
			}
			idf := ix.idf(t)
			f, fp := fuerza(t)
			f = math.Min(1, f) // una aparición en un documento de largo medio = 1
			for _, sn := range sinon[t] {
				g, gp := fuerza(sn)
				f = math.Max(f, coberturaSinonimo*math.Min(1, g))
				fp = math.Max(fp, coberturaSinonimo*gp)
			}
			cub += idf * f
			cubP += idf * fp
		}
		c := 0.0
		if totalUsuario > 0 {
			c = (math.Min(1, cub/totalUsuario) + math.Min(1, cubP/totalUsuario)) / 2
		}
		out = append(out, resultadoBM25{doc: i, bm25: s, cobertura: c})
	}
	sort.SliceStable(out, func(a, b int) bool {
		if out[a].bm25 != out[b].bm25 {
			return out[a].bm25 > out[b].bm25
		}
		return ix.docs[out[a].doc].id < ix.docs[out[b].doc].id
	})
	return out
}
