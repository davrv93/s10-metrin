// Package busqueda es la recuperación léxica e híbrida de Metrín:
//
//	consulta → reescritura (original + alias) ─┬─ BM25F (este paquete) ─┐
//	                                           └─ vector (almacen)     ─┴→ RRF → top N → reranker → top k
//
// Indexa documentos genéricos (Doc: id, campos de texto con peso, metadatos)
// para que sirva igual para los fragmentos de kb/fragmentos*.jsonl que para
// los procedimientos de kb/procedimientos/**/*.yml. No depende de rag ni del
// servidor: rag solo tendría que llamar a Buscador.Buscar en vez de
// almacen.Buscar (ver eval/BUSQUEDA.md, «Integración»).
package busqueda

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

// Doc es un documento indexable. Campos lleva el texto por campo (titulo,
// seccion, preguntas, texto, ocr…); el peso de cada campo lo fija la
// configuración. Meta viaja intacto hasta el Resultado.
type Doc struct {
	ID     string
	Campos map[string]string
	Meta   map[string]string
}

// ConfigBM25 son los parámetros de Okapi BM25F.
type ConfigBM25 struct {
	K1 float64 // saturación de frecuencia (defecto 1.2)
	B  float64 // normalización por longitud, 0…1 (defecto 0.75)
	// Pesos por campo. Un campo sin peso usa PesoDefecto; peso 0 lo excluye.
	Pesos       map[string]float64
	PesoDefecto float64
	// PesoExacto: cuánto suma la forma exacta (sin raíz) de cada palabra de
	// la consulta, relativo a la raíz. 0 desactiva el término exacto.
	PesoExacto float64
	// Difuso: tolerancia a faltas. Una raíz de la consulta rara en el índice
	// (df < MinDFDifuso) se completa con hasta 2 raíces del vocabulario a
	// distancia de edición 1 (2 si tiene 8+ letras) y al menos FactorDifuso
	// veces más frecuentes, con peso PesoDifuso. La raíz original se queda.
	// El OCR de las capturas mete faltas en el vocabulario («prespuesto»
	// tiene df 4), pero subir el umbral a df < 20 no mejoró las consultas con
	// faltas del oro (eval/BUSQUEDA.md): se queda en 3 × 10.
	Difuso       bool
	MinDFDifuso  int
	FactorDifuso int
	PesoDifuso   float64
	// Texto que se guarda por documento para el reranker (bytes) y orden de
	// los campos con que se arma.
	MaxTextoGuardado int
	CamposTexto      []string
}

// ConfigPorDefecto es la configuración recomendada por la evaluación
// (eval/BUSQUEDA.md).
func ConfigPorDefecto() ConfigBM25 {
	return ConfigBM25{
		K1: 1.2,
		B:  0.75,
		Pesos: map[string]float64{
			"titulo":    3,
			"seccion":   3,
			"preguntas": 3,
			"entidades": 2,
			"texto":     1,
			"ocr":       0.5,
			"manual":    0.5,
		},
		PesoDefecto: 1,
		// PesoExacto 0: el bono de forma exacta no mejoró en el oro (MRR@10
		// 0,765 sin bono frente a 0,739 con 0,5). Los atajos y códigos no lo
		// necesitan: son un único término cuya raíz es la forma exacta.
		PesoExacto:       0,
		Difuso:           true,
		MinDFDifuso:      3,
		FactorDifuso:     10,
		PesoDifuso:       0.7,
		MaxTextoGuardado: 1800,
		CamposTexto:      []string{"titulo", "seccion", "preguntas", "texto", "ocr"},
	}
}

func (c ConfigBM25) peso(campo string) float64 {
	if w, ok := c.Pesos[campo]; ok {
		return w
	}
	return c.PesoDefecto
}

// Posting: aparición de un término en un campo de un documento.
type Posting struct {
	Doc   int32
	Campo uint8
	TF    uint16
}

// Indice es un índice invertido BM25F en memoria. Es seguro para lecturas
// concurrentes; Agregar toma el cerrojo de escritura.
type Indice struct {
	mu  sync.RWMutex
	cfg ConfigBM25

	campos   []string
	idxCampo map[string]uint8

	ids   []string
	porID map[string]int32
	meta  []map[string]string
	texto []string

	vocab map[string]int32
	terms []string
	post  [][]Posting
	df    []int32

	largos   [][]uint32 // doc → campo → nº de términos (raíces)
	suma     []float64  // campo → suma de largos
	conCampo []int      // campo → documentos con ese campo no vacío

	porLargo map[int][]int32 // largo en runas → ids de raíces (para lo difuso)
}

// NuevoIndice crea un índice vacío.
func NuevoIndice(cfg ConfigBM25) *Indice {
	if cfg.K1 <= 0 {
		cfg.K1 = 1.2
	}
	if cfg.B < 0 || cfg.B > 1 {
		cfg.B = 0.75
	}
	if cfg.PesoDefecto == 0 && cfg.Pesos == nil {
		cfg.PesoDefecto = 1
	}
	if cfg.MaxTextoGuardado <= 0 {
		cfg.MaxTextoGuardado = 1800
	}
	if len(cfg.CamposTexto) == 0 {
		cfg.CamposTexto = []string{"titulo", "seccion", "preguntas", "texto", "ocr"}
	}
	if cfg.MinDFDifuso <= 0 {
		cfg.MinDFDifuso = 3
	}
	if cfg.FactorDifuso <= 0 {
		cfg.FactorDifuso = 10
	}
	if cfg.PesoDifuso <= 0 {
		cfg.PesoDifuso = 0.7
	}
	return &Indice{
		cfg:      cfg,
		idxCampo: map[string]uint8{},
		porID:    map[string]int32{},
		vocab:    map[string]int32{},
		porLargo: map[int][]int32{},
	}
}

// Config devuelve la configuración vigente.
func (ix *Indice) Config() ConfigBM25 {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	return ix.cfg
}

// AjustarParametros cambia k1, b y los pesos sin reindexar (los postings
// guardan la frecuencia cruda por campo). Útil para el barrido de la
// evaluación.
func (ix *Indice) AjustarParametros(k1, b float64, pesos map[string]float64, pesoExacto float64) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	if k1 > 0 {
		ix.cfg.K1 = k1
	}
	if b >= 0 && b <= 1 {
		ix.cfg.B = b
	}
	if pesos != nil {
		ix.cfg.Pesos = pesos
	}
	if pesoExacto >= 0 {
		ix.cfg.PesoExacto = pesoExacto
	}
}

// ErrDuplicado: ya hay un documento con ese ID.
var ErrDuplicado = errors.New("busqueda: id duplicado")

// Agregar indexa un documento.
func (ix *Indice) Agregar(d Doc) error {
	if strings.TrimSpace(d.ID) == "" {
		return errors.New("busqueda: documento sin id")
	}
	ix.mu.Lock()
	defer ix.mu.Unlock()
	if _, ok := ix.porID[d.ID]; ok {
		return fmt.Errorf("%w: %s", ErrDuplicado, d.ID)
	}
	n := int32(len(ix.ids))
	ix.ids = append(ix.ids, d.ID)
	ix.porID[d.ID] = n
	ix.meta = append(ix.meta, d.Meta)
	ix.texto = append(ix.texto, armarTexto(d.Campos, ix.cfg.CamposTexto, ix.cfg.MaxTextoGuardado))
	largos := make([]uint32, len(ix.campos))
	vistos := map[int32]bool{}
	// Orden fijo de campos para que el índice sea determinista.
	nombres := make([]string, 0, len(d.Campos))
	for c := range d.Campos {
		nombres = append(nombres, c)
	}
	sort.Strings(nombres)
	for _, c := range nombres {
		raices, exactos := Terminos(d.Campos[c])
		if len(raices) == 0 {
			continue
		}
		ci, err := ix.campo(c)
		if err != nil {
			return err
		}
		for int(ci) >= len(largos) {
			largos = append(largos, 0)
		}
		largos[ci] = uint32(len(raices))
		ix.suma[ci] += float64(len(raices))
		ix.conCampo[ci]++
		tf := map[string]int{}
		for i := range raices {
			tf[raices[i]]++
			if exactos[i][1:] != raices[i] {
				tf[exactos[i]]++
			}
		}
		for t, f := range tf {
			id := ix.termino(t)
			ix.post[id] = append(ix.post[id], Posting{Doc: n, Campo: ci, TF: uint16(min(f, math.MaxUint16))})
			if !vistos[id] {
				vistos[id] = true
				ix.df[id]++
			}
		}
	}
	ix.largos = append(ix.largos, largos)
	return nil
}

func (ix *Indice) campo(c string) (uint8, error) {
	if i, ok := ix.idxCampo[c]; ok {
		return i, nil
	}
	if len(ix.campos) >= 255 {
		return 0, errors.New("busqueda: demasiados campos distintos")
	}
	i := uint8(len(ix.campos))
	ix.campos = append(ix.campos, c)
	ix.idxCampo[c] = i
	ix.suma = append(ix.suma, 0)
	ix.conCampo = append(ix.conCampo, 0)
	return i, nil
}

func (ix *Indice) termino(t string) int32 {
	if id, ok := ix.vocab[t]; ok {
		return id
	}
	id := int32(len(ix.terms))
	ix.vocab[t] = id
	ix.terms = append(ix.terms, t)
	ix.post = append(ix.post, nil)
	ix.df = append(ix.df, 0)
	if !strings.HasPrefix(t, prefijoExacto) {
		l := utf8.RuneCountInString(t)
		ix.porLargo[l] = append(ix.porLargo[l], id)
	}
	return id
}

func armarTexto(campos map[string]string, orden []string, max int) string {
	var b strings.Builder
	for _, c := range orden {
		v := strings.TrimSpace(campos[c])
		if v == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(v)
		if b.Len() >= max {
			break
		}
	}
	return recortarBytes(b.String(), max)
}

func recortarBytes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	s = s[:max]
	for !utf8.ValidString(s) && len(s) > 0 {
		s = s[:len(s)-1]
	}
	return s
}

// Contar devuelve el número de documentos.
func (ix *Indice) Contar() int {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	return len(ix.ids)
}

// Vocabulario devuelve el número de términos distintos (raíces + exactos).
func (ix *Indice) Vocabulario() int {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	return len(ix.terms)
}

// Texto devuelve el texto guardado de un documento (para el reranker).
func (ix *Indice) Texto(id string) (string, bool) {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	if n, ok := ix.porID[id]; ok {
		return ix.texto[n], true
	}
	return "", false
}

// Meta devuelve los metadatos de un documento (no modificar el mapa).
func (ix *Indice) Meta(id string) (map[string]string, bool) {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	if n, ok := ix.porID[id]; ok {
		return ix.meta[n], true
	}
	return nil, false
}

// TerminoConsulta es un término ya analizado con su peso en la consulta.
type TerminoConsulta struct {
	Termino string
	Peso    float64
}

// Puntuado es un resultado léxico.
type Puntuado struct {
	ID    string
	Score float64
}

// TerminosDe analiza un texto de consulta: cada raíz con peso `peso` y cada
// forma exacta con peso·PesoExacto.
func (ix *Indice) TerminosDe(texto string, peso float64) []TerminoConsulta {
	raices, exactos := Terminos(texto)
	pe := ix.Config().PesoExacto
	out := make([]TerminoConsulta, 0, 2*len(raices))
	for i := range raices {
		if exactos[i][1:] == raices[i] {
			// La raíz ya es la forma exacta (kardex, f7, ctrl+f): el índice no
			// guarda un término «=» aparte, así que la raíz lleva los dos pesos
			// y la palabra no pierde frente a las que sí cambian al raizarse.
			out = append(out, TerminoConsulta{raices[i], peso * (1 + pe)})
			continue
		}
		out = append(out, TerminoConsulta{raices[i], peso})
		if pe > 0 {
			out = append(out, TerminoConsulta{exactos[i], peso * pe})
		}
	}
	return out
}

// BuscarLexico analiza q y devuelve los k documentos con mayor BM25F.
func (ix *Indice) BuscarLexico(q string, k int) []Puntuado {
	return ix.BuscarTerminos(ix.TerminosDe(q, 1), k)
}

// BuscarTerminos puntúa una lista de términos ponderados. Los términos
// repetidos se quedan con el mayor peso (una palabra repetida no cuenta doble).
func (ix *Indice) BuscarTerminos(terms []TerminoConsulta, k int) []Puntuado {
	return ix.BuscarTerminosFiltro(terms, k, nil)
}

// cumple dice si los metadatos tienen todos los pares del filtro (igualdad
// exacta, como el filtro de chromem en almacen.Buscar).
func cumple(meta, filtro map[string]string) bool {
	for k, v := range filtro {
		if meta[k] != v {
			return false
		}
	}
	return true
}

// BuscarTerminosFiltro es BuscarTerminos restringido a los documentos cuyos
// metadatos cumplen el filtro (nil = todos). El IDF sigue siendo el de toda
// la colección, como en un filtro posterior.
func (ix *Indice) BuscarTerminosFiltro(terms []TerminoConsulta, k int, filtro map[string]string) []Puntuado {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	N := len(ix.ids)
	if N == 0 || k <= 0 {
		return nil
	}
	pesos := map[string]float64{}
	orden := []string{}
	agregar := func(t string, w float64) {
		if prev, ok := pesos[t]; !ok {
			orden = append(orden, t)
			pesos[t] = w
		} else if w > prev {
			pesos[t] = w
		}
	}
	for _, t := range terms {
		if t.Peso <= 0 || t.Termino == "" {
			continue
		}
		agregar(t.Termino, t.Peso)
		if ix.cfg.Difuso && !strings.HasPrefix(t.Termino, prefijoExacto) {
			for _, v := range ix.difusos(t.Termino) {
				agregar(v, t.Peso*ix.cfg.PesoDifuso)
			}
		}
	}
	avg := make([]float64, len(ix.campos))
	wc := make([]float64, len(ix.campos))
	for c := range ix.campos {
		if ix.conCampo[c] > 0 {
			avg[c] = ix.suma[c] / float64(ix.conCampo[c])
		}
		wc[c] = ix.cfg.peso(ix.campos[c])
	}
	k1, b := ix.cfg.K1, ix.cfg.B
	scores := make([]float64, N)
	pseudo := make([]float64, N)
	var tocados, delTermino []int32
	for _, t := range orden {
		id, ok := ix.vocab[t]
		if !ok {
			continue
		}
		df := float64(ix.df[id])
		idf := math.Log(1 + (float64(N)-df+0.5)/(df+0.5))
		delTermino = delTermino[:0]
		for _, p := range ix.post[id] {
			w := wc[p.Campo]
			if w == 0 || avg[p.Campo] == 0 {
				continue
			}
			var largo float64
			if l := ix.largos[p.Doc]; int(p.Campo) < len(l) {
				largo = float64(l[p.Campo])
			}
			norm := 1 - b + b*largo/avg[p.Campo]
			if pseudo[p.Doc] == 0 {
				delTermino = append(delTermino, p.Doc)
			}
			pseudo[p.Doc] += w * float64(p.TF) / norm
		}
		qw := pesos[t]
		for _, d := range delTermino {
			s := pseudo[d]
			pseudo[d] = 0
			if scores[d] == 0 {
				tocados = append(tocados, d)
			}
			scores[d] += qw * idf * s * (k1 + 1) / (k1 + s)
		}
	}
	out := make([]Puntuado, 0, len(tocados))
	for _, d := range tocados {
		if len(filtro) > 0 && !cumple(ix.meta[d], filtro) {
			continue
		}
		out = append(out, Puntuado{ID: ix.ids[d], Score: scores[d]})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].ID < out[j].ID
	})
	if len(out) > k {
		out = out[:k]
	}
	return out
}

// difusos devuelve raíces del vocabulario cercanas a t cuando t falta o es
// rarísima (probable falta de ortografía). Requiere el cerrojo de lectura.
func (ix *Indice) difusos(t string) []string {
	if strings.HasPrefix(t, prefijoExacto) {
		return nil
	}
	l := utf8.RuneCountInString(t)
	if l < 4 || !soloLetras(t) {
		return nil
	}
	var df int32
	if id, ok := ix.vocab[t]; ok {
		df = ix.df[id]
	}
	if int(df) >= ix.cfg.MinDFDifuso {
		return nil
	}
	maxD := 1
	if l >= 8 {
		maxD = 2
	}
	minDF := int32(ix.cfg.FactorDifuso) * max(df, 1)
	type cand struct {
		t  string
		d  int
		df int32
	}
	var cs []cand
	rt := []rune(t)
	for ll := l - maxD; ll <= l+maxD; ll++ {
		for _, id := range ix.porLargo[ll] {
			if ix.df[id] < minDF {
				continue
			}
			v := ix.terms[id]
			if v == t {
				continue
			}
			if d := damerau([]rune(v), rt, maxD); d <= maxD {
				cs = append(cs, cand{v, d, ix.df[id]})
			}
		}
	}
	sort.Slice(cs, func(i, j int) bool {
		if cs[i].d != cs[j].d {
			return cs[i].d < cs[j].d
		}
		if cs[i].df != cs[j].df {
			return cs[i].df > cs[j].df
		}
		return cs[i].t < cs[j].t
	})
	var out []string
	for i := 0; i < len(cs) && i < 2; i++ {
		out = append(out, cs[i].t)
	}
	return out
}

func soloLetras(s string) bool {
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

// damerau es la distancia de Damerau-Levenshtein (transposición adyacente)
// con corte: devuelve max+1 en cuanto sabe que la supera.
func damerau(a, b []rune, max int) int {
	if d := len(a) - len(b); d > max || -d > max {
		return max + 1
	}
	prev2 := make([]int, len(b)+1)
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		menor := cur[0]
		for j := 1; j <= len(b); j++ {
			coste := 1
			if a[i-1] == b[j-1] {
				coste = 0
			}
			v := min(prev[j]+1, cur[j-1]+1, prev[j-1]+coste)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				v = min(v, prev2[j-2]+1)
			}
			cur[j] = v
			menor = min(menor, v)
		}
		if menor > max {
			return max + 1
		}
		prev2, prev, cur = prev, cur, prev2
	}
	return prev[len(b)]
}
