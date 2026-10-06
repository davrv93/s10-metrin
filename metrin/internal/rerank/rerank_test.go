package rerank

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNoopConservaElOrden(t *testing.T) {
	s, err := Noop{}.Rerank(context.Background(), "q", []string{"a", "b", "c"})
	if err != nil || !reflect.DeepEqual(s, []float64{3, 2, 1}) {
		t.Errorf("Noop = %v, %v", s, err)
	}
}

// servidor imita a llama-server: devuelve los resultados ordenados por
// puntaje (no por índice), como hace el real.
func servidor(t *testing.T, ruta string, puntajes map[string]float64, retraso time.Duration) (*httptest.Server, *atomic.Value) {
	t.Helper()
	ultima := &atomic.Value{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != ruta {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "método o tipo", http.StatusBadRequest)
			return
		}
		var p peticion
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ultima.Store(p)
		if retraso > 0 {
			select {
			case <-time.After(retraso):
			case <-r.Context().Done():
				return
			}
		}
		type res struct {
			Index int     `json:"index"`
			Score float64 `json:"relevance_score"`
		}
		var rs []res
		for i, d := range p.Documents {
			rs = append(rs, res{i, puntajes[d]})
		}
		// orden por puntaje descendente
		for i := 1; i < len(rs); i++ {
			for j := i; j > 0 && rs[j].Score > rs[j-1].Score; j-- {
				rs[j], rs[j-1] = rs[j-1], rs[j]
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"model": p.Model, "object": "list", "results": rs})
	})), ultima
}

func TestLlamaServerMapeaPorIndice(t *testing.T) {
	srv, ultima := servidor(t, "/v1/rerank", map[string]float64{"metrado": 6.3, "ceo": -11, "imagen": -2.9}, 0)
	defer srv.Close()
	c := NuevoLlamaServer(srv.URL)
	c.Modelo = "bge-reranker-v2-m3"
	s, err := c.Rerank(context.Background(), "¿qué es un metrado?", []string{"ceo", "metrado", "imagen"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, []float64{-11, 6.3, -2.9}) {
		t.Errorf("puntajes = %v", s)
	}
	p := ultima.Load().(peticion)
	if p.Query != "¿qué es un metrado?" || p.TopN != 3 || p.Model != "bge-reranker-v2-m3" {
		t.Errorf("petición = %+v", p)
	}
}

func TestLlamaServerRecortaYNoMandaVacios(t *testing.T) {
	srv, ultima := servidor(t, "/v1/rerank", nil, 0)
	defer srv.Close()
	c := NuevoLlamaServer(srv.URL)
	c.MaxRunas = 5
	if _, err := c.Rerank(context.Background(), "q", []string{"áéíóúxyz", "  "}); err != nil {
		t.Fatal(err)
	}
	p := ultima.Load().(peticion)
	if p.Documents[0] != "áéíóú" || p.Documents[1] != "-" {
		t.Errorf("documentos enviados = %q", p.Documents)
	}
}

func TestLlamaServerRutaAlternativaYFormatoTEI(t *testing.T) {
	// Servidor antiguo: solo /rerank, y responde al estilo TEI.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rerank" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`[{"index":1,"score":0.9},{"index":0,"score":0.1}]`))
	}))
	defer srv.Close()
	s, err := NuevoLlamaServer(srv.URL).Rerank(context.Background(), "q", []string{"a", "b"})
	if err != nil || !reflect.DeepEqual(s, []float64{0.1, 0.9}) {
		t.Errorf("TEI = %v, %v", s, err)
	}
}

func TestLlamaServerRespuestasInvalidas(t *testing.T) {
	for nombre, cuerpo := range map[string]string{
		"falta un doc":   `{"results":[{"index":0,"relevance_score":1}]}`,
		"fuera de rango": `{"results":[{"index":0,"relevance_score":1},{"index":5,"relevance_score":2}]}`,
		"sin puntaje":    `{"results":[{"index":0},{"index":1}]}`,
		"no es json":     `<html>`,
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(cuerpo)) }))
		_, err := NuevoLlamaServer(srv.URL).Rerank(context.Background(), "q", []string{"a", "b"})
		if !errors.Is(err, ErrRespuesta) {
			t.Errorf("%s: err = %v", nombre, err)
		}
		srv.Close()
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "modelo sin --reranking", http.StatusNotImplemented)
	}))
	defer srv.Close()
	if _, err := NuevoLlamaServer(srv.URL).Rerank(context.Background(), "q", []string{"a"}); err == nil || !strings.Contains(err.Error(), "501") {
		t.Errorf("HTTP 501: %v", err)
	}
}

func TestConRespaldoTimeoutDevuelveNoop(t *testing.T) {
	srv, _ := servidor(t, "/v1/rerank", map[string]float64{"a": 1, "b": 9}, 500*time.Millisecond)
	defer srv.Close()
	var motivo error
	c := ConRespaldo{Principal: NuevoLlamaServer(srv.URL), Timeout: 50 * time.Millisecond, AlFallar: func(err error) { motivo = err }}
	t0 := time.Now()
	s, err := c.Rerank(context.Background(), "q", []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Since(t0); d > 300*time.Millisecond {
		t.Errorf("no respetó el timeout: %v", d)
	}
	if !reflect.DeepEqual(s, []float64{2, 1}) {
		t.Errorf("respaldo = %v, quiere el orden original", s)
	}
	if motivo == nil || !errors.Is(motivo, context.DeadlineExceeded) {
		t.Errorf("AlFallar = %v", motivo)
	}
}

func TestConRespaldoErrorYServidorCaido(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	url := srv.URL
	var fallos int
	c := ConRespaldo{Principal: NuevoLlamaServer(url), Timeout: time.Second, AlFallar: func(error) { fallos++ }}
	if s, err := c.Rerank(context.Background(), "q", []string{"a", "b", "c"}); err != nil || !reflect.DeepEqual(s, []float64{3, 2, 1}) {
		t.Errorf("HTTP 500: %v %v", s, err)
	}
	srv.Close() // conexión rechazada
	if s, err := c.Rerank(context.Background(), "q", []string{"a", "b"}); err != nil || !reflect.DeepEqual(s, []float64{2, 1}) {
		t.Errorf("servidor caído: %v %v", s, err)
	}
	if fallos != 2 {
		t.Errorf("fallos = %d", fallos)
	}
}

func TestConRespaldoPasaElPrincipal(t *testing.T) {
	srv, _ := servidor(t, "/v1/rerank", map[string]float64{"a": 1, "b": 9}, 0)
	defer srv.Close()
	c := ConRespaldo{Principal: NuevoLlamaServer(srv.URL), Timeout: time.Second}
	if s, err := c.Rerank(context.Background(), "q", []string{"a", "b"}); err != nil || !reflect.DeepEqual(s, []float64{1, 9}) {
		t.Errorf("principal: %v %v", s, err)
	}
	if s, _ := (ConRespaldo{}).Rerank(context.Background(), "q", []string{"a", "b"}); !reflect.DeepEqual(s, []float64{2, 1}) {
		t.Errorf("sin principal: %v", s)
	}
}
