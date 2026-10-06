package servidor

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rag-go/internal/almacen"
	"rag-go/internal/llm"
	"rag-go/internal/rag"
)

type unEje struct{}

func (unEje) Nombre() string                                     { return "un-eje" }
func (unEje) Embeber(context.Context, string) ([]float32, error) { return []float32{1}, nil }

type modeloEnTrozos []string

func (m modeloEnTrozos) Chat(context.Context, []llm.Mensaje) (string, error) {
	return strings.Join(m, ""), nil
}

func (m modeloEnTrozos) ChatStream(_ context.Context, _ []llm.Mensaje, emitir func(string)) (string, error) {
	for _, t := range m {
		emitir(t)
	}
	return strings.Join(m, ""), nil
}

func TestAskStreamEmiteDeltasYTerminaConLaRespuesta(t *testing.T) {
	t.Setenv("RAG_IMAGENES", "/no/existe.jsonl")
	ctx := context.Background()
	a, _ := almacen.Abrir("", unEje{})
	a.Reemplazar(ctx, "s3:docs/factura.md", almacen.FuenteS3, "v1", []almacen.Trozo{{
		ID: almacen.IDTrozo("s3:docs/factura.md", 0), Texto: "Para emitir una factura abre Ventas.",
		Metadata: map[string]string{"source": almacen.FuenteS3, "cita": "s3:docs/factura.md#0"},
	}})
	r := &rag.RAG{Almacen: a, LLM: modeloEnTrozos{"Abre ", "Ventas."}, MaxDistancia: 0.5}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/ask/stream", strings.NewReader(`{"pregunta":"¿cómo emito una factura?"}`))
	Nuevo(r, 5*time.Second).ServeHTTP(rr, req)

	if rr.Code != 200 || !strings.HasPrefix(rr.Header().Get("Content-Type"), "application/x-ndjson") {
		t.Fatalf("código %d, tipo %q", rr.Code, rr.Header().Get("Content-Type"))
	}
	var deltas []string
	var fin *rag.Respuesta
	sc := bufio.NewScanner(rr.Body)
	for sc.Scan() {
		var e struct {
			Tipo      string         `json:"tipo"`
			Texto     string         `json:"texto"`
			Respuesta *rag.Respuesta `json:"respuesta"`
			Error     string         `json:"error"`
		}
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			t.Fatalf("línea no es JSON: %q", sc.Text())
		}
		switch e.Tipo {
		case "delta":
			deltas = append(deltas, e.Texto)
		case "fin":
			fin = e.Respuesta
		default:
			t.Fatalf("evento inesperado: %s", sc.Text())
		}
	}
	if strings.Join(deltas, "|") != "Abre |Ventas." {
		t.Errorf("deltas = %q", deltas)
	}
	if fin == nil || fin.Respuesta != "Abre Ventas." || len(fin.Fuentes) != 1 {
		t.Fatalf("fin = %+v", fin)
	}
}

func TestAskStreamSinPreguntaEs400(t *testing.T) {
	rr := httptest.NewRecorder()
	Nuevo(nil, time.Second).ServeHTTP(rr, httptest.NewRequest("POST", "/ask/stream", strings.NewReader(`{}`)))
	if rr.Code != 400 {
		t.Fatalf("código %d, quería 400", rr.Code)
	}
}
