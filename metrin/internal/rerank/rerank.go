// Package rerank reordena candidatos con un cross-encoder local.
//
// Implementaciones:
//   - LlamaServer: cliente de `llama-server --reranking` (llama.cpp), endpoint
//     /v1/rerank (Jina/OpenAI-compatible; también acepta la respuesta de TEI).
//     Modelo evaluado: bge-reranker-v2-m3 (Apache-2.0) en GGUF Q4_K_M.
//   - Noop: deja el orden de entrada.
//   - ConRespaldo: envuelve otro Reranker con un tiempo máximo; si falla o se
//     pasa, responde con Noop (el orden de entrada) y avisa por AlFallar.
//
// No llama a ninguna API externa.
package rerank

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// Reranker puntúa cada documento frente a la consulta: un puntaje por
// documento, en el mismo orden; mayor es más relevante.
type Reranker interface {
	Rerank(ctx context.Context, q string, docs []string) ([]float64, error)
}

// Noop conserva el orden: el primero recibe el puntaje más alto.
type Noop struct{}

func (Noop) Rerank(_ context.Context, _ string, docs []string) ([]float64, error) {
	out := make([]float64, len(docs))
	for i := range docs {
		out[i] = float64(len(docs) - i)
	}
	return out, nil
}

// LlamaServer es el cliente de llama-server con --reranking.
type LlamaServer struct {
	URL    string // p. ej. http://127.0.0.1:8091
	Modelo string // informativo; llama-server usa el modelo cargado
	// Ruta del endpoint; defecto /v1/rerank. Si responde 404 se prueba
	// /rerank (versiones antiguas de llama.cpp y TEI).
	Ruta string
	// MaxRunas recorta cada documento antes de enviarlo (defecto 1200):
	// el coste del cross-encoder crece con la longitud y los primeros
	// párrafos de un fragmento son los que deciden.
	MaxRunas int
	HTTP     *http.Client
}

// NuevoLlamaServer crea el cliente con un http.Client propio.
func NuevoLlamaServer(url string) *LlamaServer {
	return &LlamaServer{URL: strings.TrimRight(url, "/"), HTTP: &http.Client{Timeout: 60 * time.Second}}
}

type peticion struct {
	Model     string   `json:"model,omitempty"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n"`
}

type resultado struct {
	Index          int      `json:"index"`
	RelevanceScore *float64 `json:"relevance_score"`
	Score          *float64 `json:"score"`
}

// ErrRespuesta: el servidor respondió algo que no se puede interpretar.
var ErrRespuesta = errors.New("rerank: respuesta inválida")

func (l *LlamaServer) Rerank(ctx context.Context, q string, docs []string) ([]float64, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	max := l.MaxRunas
	if max <= 0 {
		max = 1200
	}
	recortados := make([]string, len(docs))
	for i, d := range docs {
		recortados[i] = recortarRunas(d, max)
		if strings.TrimSpace(recortados[i]) == "" {
			recortados[i] = "-" // llama-server rechaza documentos vacíos
		}
	}
	cuerpo, _ := json.Marshal(peticion{Model: l.Modelo, Query: q, Documents: recortados, TopN: len(docs)})
	ruta := l.Ruta
	if ruta == "" {
		ruta = "/v1/rerank"
	}
	b, codigo, err := l.post(ctx, ruta, cuerpo)
	if err == nil && codigo == http.StatusNotFound && l.Ruta == "" {
		b, codigo, err = l.post(ctx, "/rerank", cuerpo)
	}
	if err != nil {
		return nil, err
	}
	if codigo != http.StatusOK {
		return nil, fmt.Errorf("rerank: HTTP %d: %s", codigo, recortarRunas(string(b), 200))
	}
	return interpretar(b, len(docs))
}

func (l *LlamaServer) post(ctx context.Context, ruta string, cuerpo []byte) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(l.URL, "/")+ruta, bytes.NewReader(cuerpo))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	c := l.HTTP
	if c == nil {
		c = http.DefaultClient
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("rerank: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	return b, resp.StatusCode, err
}

// interpretar acepta {"results":[{"index","relevance_score"}]} (Jina,
// OpenAI-compatible, llama.cpp) y [{"index","score"}] (TEI).
func interpretar(b []byte, n int) ([]float64, error) {
	var rs []resultado
	var env struct {
		Results []resultado `json:"results"`
	}
	if err := json.Unmarshal(b, &env); err == nil && env.Results != nil {
		rs = env.Results
	} else if err := json.Unmarshal(b, &rs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRespuesta, err)
	}
	out := make([]float64, n)
	visto := make([]bool, n)
	for _, r := range rs {
		if r.Index < 0 || r.Index >= n {
			return nil, fmt.Errorf("%w: índice %d fuera de rango", ErrRespuesta, r.Index)
		}
		s := r.RelevanceScore
		if s == nil {
			s = r.Score
		}
		if s == nil {
			return nil, fmt.Errorf("%w: resultado sin puntaje", ErrRespuesta)
		}
		out[r.Index] = *s
		visto[r.Index] = true
	}
	for i, v := range visto {
		if !v {
			return nil, fmt.Errorf("%w: falta el documento %d", ErrRespuesta, i)
		}
	}
	return out, nil
}

func recortarRunas(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

// ConRespaldo aplica Timeout al reranker principal y, si falla, devuelve el
// orden de entrada (Noop) sin error. AlFallar, si está, recibe el motivo
// (para la traza o el registro).
type ConRespaldo struct {
	Principal Reranker
	Timeout   time.Duration // 0 = sin límite propio (solo el del contexto)
	AlFallar  func(error)
}

func (c ConRespaldo) Rerank(ctx context.Context, q string, docs []string) ([]float64, error) {
	if c.Principal == nil {
		return Noop{}.Rerank(ctx, q, docs)
	}
	cctx := ctx
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		cctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}
	type res struct {
		s   []float64
		err error
	}
	ch := make(chan res, 1)
	go func() {
		s, err := c.Principal.Rerank(cctx, q, docs)
		ch <- res{s, err}
	}()
	var r res
	select {
	case r = <-ch:
	case <-cctx.Done():
		r = res{nil, fmt.Errorf("rerank: %w", cctx.Err())}
	}
	if r.err == nil && len(r.s) != len(docs) {
		r.err = fmt.Errorf("%w: %d puntajes para %d documentos", ErrRespuesta, len(r.s), len(docs))
	}
	if r.err != nil {
		if c.AlFallar != nil {
			c.AlFallar(r.err)
		}
		return Noop{}.Rerank(ctx, q, docs)
	}
	return r.s, nil
}
