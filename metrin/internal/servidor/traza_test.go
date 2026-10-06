package servidor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"rag-go/internal/almacen"
	"rag-go/internal/llm"
	"rag-go/internal/rag"
)

// Dobles mínimos (los de rag_test.go son internos de ese paquete).
type ejes struct{}

func (ejes) Nombre() string { return "dos-ejes" }
func (ejes) Embeber(_ context.Context, t string) ([]float32, error) {
	if strings.Contains(t, "precio") {
		return []float32{1, 0}, nil
	}
	return []float32{0, 1}, nil
}

type llmEco struct{ texto string }

func (l llmEco) Chat(context.Context, []llm.Mensaje) (string, error) { return l.texto, nil }

func armarRAG(t *testing.T) *rag.RAG {
	t.Helper()
	ctx := context.Background()
	a, err := almacen.Abrir("", ejes{})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Reemplazar(ctx, "s3:docs/tarifas.md", almacen.FuenteS3, "e1", []almacen.Trozo{{
		ID: almacen.IDTrozo("s3:docs/tarifas.md", 0), Texto: "El plan Pro tiene precio de 49 €",
		Metadata: map[string]string{"source": almacen.FuenteS3, "cita": "s3:rag-demo/docs/tarifas.md#0", "title": "Tarifas", "page": "2"},
	}}); err != nil {
		t.Fatal(err)
	}
	return &rag.RAG{Almacen: a, LLM: llmEco{"Cuesta 49 €."}, MaxDistancia: 0.5,
		RutaFallos: filepath.Join(t.TempDir(), "sin_respuesta.jsonl")}
}

// pedir hace POST /ask y devuelve el cuerpo y la respuesta.
func pedir(t *testing.T, h http.Handler, cuerpo string) (*httptest.ResponseRecorder, []byte) {
	t.Helper()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("POST", "/ask", strings.NewReader(cuerpo)))
	return rr, rr.Body.Bytes()
}

var reMs = regexp.MustCompile(`"ms_(busqueda|llm)":\d+`)

// sinTiempos: ms_busqueda y ms_llm son los únicos campos que dependen del reloj.
func sinTiempos(b []byte) string { return reMs.ReplaceAllString(string(b), `"ms_$1":0`) }

func aislarMedios(t *testing.T) {
	t.Setenv("RAG_IMAGENES", filepath.Join(t.TempDir(), "no-hay-imagenes.jsonl"))
	t.Setenv("RAG_MEDIA_DATOS", t.TempDir())
}

func TestAskSinTrazaEsIdentico(t *testing.T) {
	aislarMedios(t)
	r := armarRAG(t)
	apagado := Nuevo(r, 5*time.Second)
	apagadoExplicito := Nuevo(r, 5*time.Second, Opciones{Traza: false})
	encendido := Nuevo(r, 5*time.Second, Opciones{Traza: true})
	for _, pregunta := range []string{
		"¿qué precio tiene el Pro?",                                  // rag con fuente
		"¿cómo es el formulario de contacto que tiene validaciones?", // sin contexto
		"¿Cuántas preguntas acertaste hoy?",                          // regla
	} {
		base := `{"pregunta":` + jsonTexto(pregunta)
		rrRef, ref := pedir(t, apagado, base+`}`)
		if rrRef.Code != 200 {
			t.Fatalf("%q: código %d: %s", pregunta, rrRef.Code, ref)
		}
		for nombre, caso := range map[string]struct {
			h      http.Handler
			cuerpo string
		}{
			"apagado + traza:true":         {apagado, base + `,"traza":true}`},
			"apagado explícito":            {apagadoExplicito, base + `}`},
			"encendido sin campo":          {encendido, base + `}`},
			"encendido + traza:false":      {encendido, base + `,"traza":false}`},
			"encendido + traza:\"si\"":     {encendido, base + `,"traza":"si"}`},
			"encendido + traza:null":       {encendido, base + `,"traza":null}`},
			"encendido + traza:{\"x\":1}":  {encendido, base + `,"traza":{"x":1}}`},
			"apagado + traza:[1,2] (raro)": {apagado, base + `,"traza":[1,2]}`},
		} {
			rr, got := pedir(t, caso.h, caso.cuerpo)
			if rr.Code != rrRef.Code || sinTiempos(got) != sinTiempos(ref) {
				t.Errorf("%q / %s: la respuesta cambió\nref: %s\ngot: %s", pregunta, nombre, ref, got)
			}
			if cc := rr.Header().Get("Cache-Control"); cc != "" {
				t.Errorf("%q / %s: sin traza no debe llevar Cache-Control (%q)", pregunta, nombre, cc)
			}
		}

		// Con traza: la respuesta es la misma más la clave «traza».
		rr, conT := pedir(t, encendido, base+`,"traza":true}`)
		if rr.Code != 200 || rr.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%q: con traza código %d, Cache-Control %q", pregunta, rr.Code, rr.Header().Get("Cache-Control"))
		}
		var conMapa, refMapa map[string]any
		if err := json.Unmarshal(conT, &conMapa); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(ref, &refMapa); err != nil {
			t.Fatal(err)
		}
		tr, ok := conMapa["traza"].(map[string]any)
		if !ok {
			t.Fatalf("%q: falta la clave traza: %s", pregunta, conT)
		}
		if etapas := tr["etapas"].([]any); len(etapas) != 18 {
			t.Errorf("%q: %d etapas", pregunta, len(etapas))
		}
		delete(conMapa, "traza")
		a, _ := json.Marshal(conMapa)
		b, _ := json.Marshal(refMapa)
		if sinTiempos(a) != sinTiempos(b) {
			t.Errorf("%q: con traza cambió algo más que la clave traza\nref: %s\ncon: %s", pregunta, b, a)
		}
	}
}

func TestAskConTrazaAnotaFotosYFin(t *testing.T) {
	aislarMedios(t)
	h := Nuevo(armarRAG(t), 5*time.Second, Opciones{Traza: true})
	_, cuerpo := pedir(t, h, `{"pregunta":"¿qué precio tiene el Pro?","traza":true}`)
	var res struct {
		Traza struct {
			Version  int     `json:"version"`
			TotalMs  float64 `json:"total_ms"`
			Pregunta string  `json:"pregunta"`
			Etapas   []struct {
				ID     string         `json:"id"`
				Estado string         `json:"estado"`
				Datos  map[string]any `json:"datos"`
			} `json:"etapas"`
		} `json:"traza"`
	}
	if err := json.Unmarshal(cuerpo, &res); err != nil {
		t.Fatal(err)
	}
	if res.Traza.Version != 1 || res.Traza.Pregunta != "¿qué precio tiene el Pro?" {
		t.Fatalf("traza: %+v", res.Traza)
	}
	estados := map[string]string{}
	for _, e := range res.Traza.Etapas {
		estados[e.ID] = e.Estado
		if e.ID == "fotos" && e.Datos["fotos"] != float64(0) {
			t.Errorf("fotos: %+v", e.Datos)
		}
	}
	for id, quiero := range map[string]string{"fotos": "ok", "fin": "ok", "busqueda": "ok", "generacion": "ok"} {
		if estados[id] != quiero {
			t.Errorf("%s = %q, quería %q", id, estados[id], quiero)
		}
	}
}

func TestHealthDiceSiHayTraza(t *testing.T) {
	r := armarRAG(t)
	for _, encendida := range []bool{false, true} {
		rr := httptest.NewRecorder()
		Nuevo(r, time.Second, Opciones{Traza: encendida}).ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))
		var h map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &h); err != nil {
			t.Fatal(err)
		}
		if h["traza"] != encendida || h["ok"] != true {
			t.Errorf("health = %v, quería traza=%v", h, encendida)
		}
	}
}

func jsonTexto(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
