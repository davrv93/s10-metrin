package busqueda

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// Reordenador es la interfaz de rerank.Reranker (se repite aquí para no
// importar el paquete rerank): puntúa cada documento frente a la consulta.
type Reordenador interface {
	Rerank(ctx context.Context, q string, docs []string) ([]float64, error)
}

// Modo de búsqueda.
type Modo string

const (
	ModoHibrido Modo = "hibrido" // vector ∪ BM25 → RRF (→ reranker si hay)
	ModoVector  Modo = "vector"  // solo vectorial
	ModoLexico  Modo = "lexico"  // solo BM25
)

// Opciones del Buscador.
type Opciones struct {
	Modo Modo
	// RRF: puntaje = Σ peso / (RRFK + rango). 60 es el valor de Cormack et
	// al. (2009).
	RRFK       float64
	PesoVector float64
	PesoLexico float64
	// PesoVectorSinRerank: peso del vector para el orden final cuando no
	// hay reranker o este falla. El pool que se le pasa al reranker rinde
	// más con pesos iguales (más diverso), pero como orden final el vector
	// estático resta: medido en eval/BUSQUEDA.md (0,1 ≈ BM25 en Recall y
	// mejor MRR). 0 = usar PesoVector también sin reranker.
	PesoVectorSinRerank float64
	// Candidatos que se piden a cada lista antes de fusionar.
	Candidatos int
	// TopRerank: cuántos de la lista fusionada pasan por el reranker.
	TopRerank int
	// TimeoutRerank: si el reranker no contesta a tiempo se queda el orden
	// RRF con PesoVectorSinRerank (Informe.Degradado = "sin_rerank").
	// 0 = solo el plazo del contexto.
	TimeoutRerank time.Duration
	// Expandir: añade los alias de la consulta a BM25 con PesoAlias.
	Expandir  bool
	PesoAlias float64
	// TamCache: entradas de la LRU de resultados (0 = sin caché).
	TamCache int
}

// OpcionesPorDefecto: la configuración recomendada en eval/BUSQUEDA.md.
func OpcionesPorDefecto() Opciones {
	return Opciones{
		Modo:                ModoHibrido,
		RRFK:                60,
		PesoVector:          1,
		PesoLexico:          1,
		PesoVectorSinRerank: 0.1,
		Candidatos:          50,
		TopRerank:           30,
		// TimeoutRerank: ver eval/BUSQUEDA.md (p95 medido del reranker).
		TimeoutRerank: 3 * time.Second,
		Expandir:      false,
		PesoAlias:     0.3,
		TamCache:      256,
	}
}

// Resultado de una búsqueda.
type Resultado struct {
	ID          string
	Score       float64 // clave del orden final: rerank, si lo hubo; si no, RRF (o BM25 / coseno en los modos simples)
	RRF         float64
	Vector      float64 // similitud coseno; 0 si no salió por vector
	BM25        float64 // puntaje BM25F; 0 si no salió por léxico
	Rerank      float64 // puntaje del reranker (válido si Reordenado)
	Reordenado  bool    // pasó por el reranker
	RangoVector int     // 1…n; 0 = no salió por vector
	RangoBM25   int     // 1…n; 0 = no salió por BM25
	Fuentes     []string
	Meta        map[string]string // del índice léxico o del almacén; no modificar
}

// Informe es la búsqueda con su traza.
type Informe struct {
	Consulta    Consulta
	Resultados  []Resultado
	Degradado   []string // "sin_vector", "sin_rerank"
	FalloVector string
	FalloRerank string
	DeCache     bool
	Tiempos     map[string]time.Duration // vector, lexico, fusion, rerank, total
}

// Buscador combina el índice léxico, el vectorial y el reranker.
type Buscador struct {
	Lexico   *Indice
	Vector   Vectorial   // nil → solo léxico
	Reranker Reordenador // nil → sin rerank
	Alias    *Alias      // nil → sin expansión (sigue normalizando)
	Opc      Opciones
	cache    *LRU[string, Informe]
	// trasLexico: gancho de pruebas (se llama al acabar la búsqueda léxica).
	trasLexico func()
}

// NuevoBuscador completa las opciones vacías con las de defecto.
func NuevoBuscador(lex *Indice, vec Vectorial, rr Reordenador, al *Alias, o Opciones) *Buscador {
	d := OpcionesPorDefecto()
	if o.Modo == "" {
		o.Modo = d.Modo
	}
	if o.RRFK <= 0 {
		o.RRFK = d.RRFK
	}
	if o.PesoVector == 0 && o.PesoLexico == 0 {
		o.PesoVector, o.PesoLexico = d.PesoVector, d.PesoLexico
	}
	if o.Candidatos <= 0 {
		o.Candidatos = d.Candidatos
	}
	if o.TopRerank <= 0 {
		o.TopRerank = d.TopRerank
	}
	if o.PesoAlias < 0 {
		o.PesoAlias = 0
	}
	return &Buscador{Lexico: lex, Vector: vec, Reranker: rr, Alias: al, Opc: o, cache: NuevoLRU[string, Informe](o.TamCache)}
}

// Cache expone la LRU de resultados (para métricas).
func (b *Buscador) Cache() *LRU[string, Informe] { return b.cache }

// Buscar devuelve los k mejores resultados.
func (b *Buscador) Buscar(ctx context.Context, q string, k int) ([]Resultado, error) {
	inf, err := b.BuscarInforme(ctx, q, k)
	return inf.Resultados, err
}

// BuscarFiltrado es Buscar restringido a los documentos cuyos metadatos
// cumplen filtro (igualdad exacta: source, confianza, manual…), como el
// filtro de almacen.Buscar.
func (b *Buscador) BuscarFiltrado(ctx context.Context, q string, k int, filtro map[string]string) ([]Resultado, error) {
	inf, err := b.BuscarInformeFiltrado(ctx, q, k, filtro)
	return inf.Resultados, err
}

// ErrSinFuentes: ni el índice léxico ni el vectorial están disponibles.
var ErrSinFuentes = errors.New("busqueda: sin índice léxico ni vectorial")

// BuscarInforme es Buscar con la traza: consulta reescrita, degradaciones y
// tiempos. Degradación: si falla el vector, sigue solo con BM25; si falla el
// reranker, se queda el orden RRF. Solo es error que fallen las dos listas
// (o que se cancele el contexto).
func (b *Buscador) BuscarInforme(ctx context.Context, q string, k int) (Informe, error) {
	return b.BuscarInformeFiltrado(ctx, q, k, nil)
}

// claveFiltro serializa el filtro en orden para la clave de caché.
func claveFiltro(f map[string]string) string {
	if len(f) == 0 {
		return ""
	}
	ks := make([]string, 0, len(f))
	for k := range f {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	var sb strings.Builder
	for _, k := range ks {
		sb.WriteString(k + "=" + f[k] + ";")
	}
	return sb.String()
}

// BuscarInformeFiltrado es BuscarInforme con filtro de metadatos.
func (b *Buscador) BuscarInformeFiltrado(ctx context.Context, q string, k int, filtro map[string]string) (Informe, error) {
	t0 := time.Now()
	c := b.Alias.Reescribir(q)
	clave := fmt.Sprintf("%s|%d|%s|%s", b.Opc.Modo, k, claveFiltro(filtro), c.Normalizada)
	if inf, ok := b.cache.Obtener(clave); ok {
		inf = copiarInforme(inf)
		inf.Consulta.Original = q
		inf.DeCache = true
		inf.Tiempos = map[string]time.Duration{"total": time.Since(t0)}
		return inf, nil
	}
	inf := Informe{Consulta: c, Tiempos: map[string]time.Duration{}}
	n := max(b.Opc.Candidatos, k)
	modo := b.Opc.Modo

	var (
		wg     sync.WaitGroup
		vec    []Candidato
		errVec error
		lex    []Puntuado
		tVec   time.Duration
		tLex   time.Duration
	)
	usaVector := modo != ModoLexico && b.Vector != nil
	usaLexico := modo != ModoVector && b.Lexico != nil
	if !usaVector && !usaLexico {
		return inf, ErrSinFuentes
	}
	if usaVector {
		wg.Add(1)
		go func() {
			defer wg.Done()
			t := time.Now()
			defer func() {
				if r := recover(); r != nil {
					errVec = fmt.Errorf("pánico en la búsqueda vectorial: %v", r)
				}
				tVec = time.Since(t)
			}()
			if vf, ok := b.Vector.(VectorialFiltrado); ok && len(filtro) > 0 {
				vec, errVec = vf.BuscarVectorFiltro(ctx, c.Normalizada, n, filtro)
				return
			}
			vec, errVec = b.Vector.BuscarVector(ctx, c.Normalizada, n)
			if errVec == nil && len(filtro) > 0 {
				ok := vec[:0]
				for _, x := range vec {
					m := x.Meta
					if b.Lexico != nil {
						if mm, esta := b.Lexico.Meta(x.ID); esta {
							m = mm
						}
					}
					if cumple(m, filtro) {
						ok = append(ok, x)
					}
				}
				vec = ok
			}
		}()
	}
	if usaLexico {
		wg.Add(1)
		go func() {
			defer wg.Done()
			t := time.Now()
			pa := 0.0
			if b.Opc.Expandir {
				pa = b.Opc.PesoAlias
			}
			lex = b.Lexico.BuscarTerminosFiltro(b.Lexico.TerminosConsulta(c, pa), n, filtro)
			tLex = time.Since(t)
			if b.trasLexico != nil {
				b.trasLexico()
			}
		}()
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return inf, err
	}
	if usaVector {
		inf.Tiempos["vector"] = tVec
	}
	if usaLexico {
		inf.Tiempos["lexico"] = tLex
	}
	if errVec != nil {
		inf.FalloVector = errVec.Error()
		inf.Degradado = append(inf.Degradado, "sin_vector")
		vec = nil
		if !usaLexico {
			return inf, fmt.Errorf("busqueda vectorial: %w", errVec)
		}
	}

	t := time.Now()
	res := b.fusionar(vec, lex, modo, b.Opc.PesoVector)
	inf.Tiempos["fusion"] = time.Since(t)
	reordenado := false

	if b.Reranker != nil && len(res) > 0 {
		t := time.Now()
		rctx, cancelar := ctx, context.CancelFunc(func() {})
		if b.Opc.TimeoutRerank > 0 {
			rctx, cancelar = context.WithTimeout(ctx, b.Opc.TimeoutRerank)
		}
		err := b.reordenar(rctx, c.Normalizada, res, textosDe(vec))
		cancelar()
		if err != nil {
			inf.FalloRerank = err.Error()
			inf.Degradado = append(inf.Degradado, "sin_rerank")
		} else {
			reordenado = true
		}
		inf.Tiempos["rerank"] = time.Since(t)
	}
	if !reordenado && modo == ModoHibrido && b.Opc.PesoVectorSinRerank > 0 && b.Opc.PesoVectorSinRerank != b.Opc.PesoVector {
		res = b.fusionar(vec, lex, modo, b.Opc.PesoVectorSinRerank)
	}
	if len(res) > k {
		res = res[:k]
	}
	inf.Resultados = res
	inf.Tiempos["total"] = time.Since(t0)
	// Solo se guarda lo que salió completo: un resultado degradado se vuelve
	// a intentar en la próxima consulta.
	if len(inf.Degradado) == 0 {
		b.cache.Poner(clave, copiarInforme(inf))
	}
	return inf, nil
}

// fusionar combina las listas por RRF. En los modos simples el Score es el
// de la única lista.
func (b *Buscador) fusionar(vec []Candidato, lex []Puntuado, modo Modo, pesoVector float64) []Resultado {
	por := map[string]*Resultado{}
	var orden []string
	obtener := func(id string) *Resultado {
		if r, ok := por[id]; ok {
			return r
		}
		r := &Resultado{ID: id}
		if b.Lexico != nil {
			r.Meta, _ = b.Lexico.Meta(id)
		}
		por[id] = r
		orden = append(orden, id)
		return r
	}
	for i, c := range vec {
		r := obtener(c.ID)
		r.Vector = c.Similitud
		r.RangoVector = i + 1
		r.RRF += pesoVector / (b.Opc.RRFK + float64(i+1))
		r.Fuentes = append(r.Fuentes, "vector")
		if r.Meta == nil {
			r.Meta = c.Meta
		}
	}
	for i, p := range lex {
		r := obtener(p.ID)
		r.BM25 = p.Score
		r.RangoBM25 = i + 1
		r.RRF += b.Opc.PesoLexico / (b.Opc.RRFK + float64(i+1))
		r.Fuentes = append(r.Fuentes, "bm25")
	}
	out := make([]Resultado, 0, len(orden))
	for _, id := range orden {
		r := por[id]
		switch modo {
		case ModoVector:
			r.Score = r.Vector
		case ModoLexico:
			r.Score = r.BM25
		default:
			r.Score = r.RRF
		}
		out = append(out, *r)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		// Empate de RRF: primero el que tiene mejor rango en cualquiera de
		// las listas, y luego por id para ser deterministas.
		mi, mj := mejorRango(out[i]), mejorRango(out[j])
		if mi != mj {
			return mi < mj
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// textosDe: textos del almacén, para documentos que el índice léxico no
// conoce.
func textosDe(vec []Candidato) map[string]string {
	textos := make(map[string]string, len(vec))
	for _, c := range vec {
		textos[c.ID] = c.Texto
	}
	return textos
}

func mejorRango(r Resultado) int {
	m := math.MaxInt
	if r.RangoVector > 0 {
		m = r.RangoVector
	}
	if r.RangoBM25 > 0 && r.RangoBM25 < m {
		m = r.RangoBM25
	}
	return m
}

// reordenar pasa los TopRerank primeros por el reranker y los ordena por su
// puntaje; el resto queda detrás en orden RRF. Si el reranker falla, res no
// cambia.
func (b *Buscador) reordenar(ctx context.Context, q string, res []Resultado, textosVector map[string]string) error {
	top := min(b.Opc.TopRerank, len(res))
	docs := make([]string, top)
	for i := 0; i < top; i++ {
		docs[i] = textosVector[res[i].ID]
		if b.Lexico != nil {
			if t, ok := b.Lexico.Texto(res[i].ID); ok {
				docs[i] = t
			}
		}
	}
	scores, err := b.Reranker.Rerank(ctx, q, docs)
	if err != nil {
		return err
	}
	if len(scores) != top {
		return fmt.Errorf("el reranker devolvió %d puntajes para %d documentos", len(scores), top)
	}
	for i := 0; i < top; i++ {
		if math.IsNaN(scores[i]) || math.IsInf(scores[i], 0) {
			return errors.New("el reranker devolvió un puntaje no finito")
		}
	}
	for i := 0; i < top; i++ {
		res[i].Rerank = scores[i]
		res[i].Reordenado = true
		res[i].Score = scores[i]
		res[i].Fuentes = append(res[i].Fuentes, "rerank")
	}
	sort.SliceStable(res[:top], func(i, j int) bool { return res[i].Rerank > res[j].Rerank })
	return nil
}

func copiarInforme(inf Informe) Informe {
	out := inf
	out.Resultados = make([]Resultado, len(inf.Resultados))
	for i, r := range inf.Resultados {
		r.Fuentes = append([]string(nil), r.Fuentes...)
		out.Resultados[i] = r
	}
	out.Degradado = append([]string(nil), inf.Degradado...)
	out.Consulta.Entidades = append([]string(nil), inf.Consulta.Entidades...)
	out.Consulta.Aliases = append([]string(nil), inf.Consulta.Aliases...)
	out.Tiempos = map[string]time.Duration{}
	for k, v := range inf.Tiempos {
		out.Tiempos[k] = v
	}
	return out
}
