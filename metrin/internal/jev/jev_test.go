package jev

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// falso es un servidor /v1/systemone de mentira: devuelve la
// respuesta dada y graba la última petición para inspeccionarla.
type falso struct {
	respuesta string
	codigo    int
	peticion  string
	cabeceras http.Header
}

func (f *falso) servir(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/systemone" {
		http.NotFound(w, r)
		return
	}
	b, _ := io.ReadAll(r.Body)
	f.peticion = string(b)
	f.cabeceras = r.Header.Clone()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(f.codigo)
	io.WriteString(w, f.respuesta)
}

func nuevoFalso(t *testing.T, respuesta string, codigo int) (*Cliente, *falso) {
	t.Helper()
	f := &falso{respuesta: respuesta, codigo: codigo}
	srv := httptest.NewServer(http.HandlerFunc(f.servir))
	t.Cleanup(srv.Close)
	return Nuevo(srv.URL, "local-model", 5*time.Second), f
}

const respuestaDos = `{
  "model": "local-model",
  "answers": {
    "departamento": {"type": "choice", "choice": "support",
      "probabilities": {"sales": 0.2, "support": 0.8}, "confidence": 0.72},
    "severidad": {"type": "score", "score": 1.4,
      "probabilities": {"0": 0.6, "1": 0.4},
      "legend": {"0": "Minor inconvenience", "1": "Product cannot be used"},
      "confidence": 0.67},
    "devolucion": {"type": "noul", "noul": 0.91}
  },
  "usage": {"input_tokens": 120, "output_tokens": 0}
}`

func TestDecirParseaRespuestasPorID(t *testing.T) {
	c, _ := nuevoFalso(t, respuestaDos, http.StatusOK)
	rs, err := c.Decidir(context.Background(),
		"My order arrived broken.",
		map[string]Pregunta{
			"departamento": {Tipo: "choice", Criterios: map[string]any{
				"sales": "New purchases", "support": "Problems with an order"}},
			"severidad": {Tipo: "score", Criterios: []any{
				"Minor inconvenience", "Product cannot be used"}},
			"devolucion": {Tipo: "noul"},
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 3 {
		t.Fatalf("respuestas = %d, quería 3", len(rs))
	}
	if rs["departamento"].Choice != "support" || rs["departamento"].Confianza != 0.72 {
		t.Fatalf("choice inesperado: %#v", rs["departamento"])
	}
	if rs["departamento"].Probabilidades["support"] != 0.8 {
		t.Fatalf("probabilidades inesperadas: %#v", rs["departamento"].Probabilidades)
	}
	if rs["severidad"].Score != 1.4 || rs["severidad"].Leyenda["1"] != "Product cannot be used" {
		t.Fatalf("score inesperado: %#v", rs["severidad"])
	}
	if rs["devolucion"].Noul != 0.91 {
		t.Fatalf("noul inesperado: %#v", rs["devolucion"])
	}
}

func TestDecirEnviaCuerpoYModelo(t *testing.T) {
	c, f := nuevoFalso(t, respuestaDos, http.StatusOK)
	_, err := c.Decidir(context.Background(),
		map[string]any{"mensaje": "roto"},
		map[string]Pregunta{"q1": {Tipo: "noul", Instrucciones: "¿quiere reembolso?"}})
	if err != nil {
		t.Fatal(err)
	}
	var cuerpo struct {
		Modelo    string              `json:"model"`
		Estado    map[string]any      `json:"state"`
		Preguntas map[string]Pregunta `json:"questions"`
	}
	if err := json.Unmarshal([]byte(f.peticion), &cuerpo); err != nil {
		t.Fatalf("cuerpo no es JSON válido: %v (%s)", err, f.peticion)
	}
	if cuerpo.Modelo != "local-model" {
		t.Fatalf("modelo = %q", cuerpo.Modelo)
	}
	if cuerpo.Estado["mensaje"] != "roto" {
		t.Fatalf("estado = %#v", cuerpo.Estado)
	}
	if cuerpo.Preguntas["q1"].Tipo != "noul" || instruccionesJSON(cuerpo.Preguntas["q1"]) != "¿quiere reembolso?" {
		t.Fatalf("pregunta = %#v", cuerpo.Preguntas["q1"])
	}
	if ct := f.cabeceras.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q", ct)
	}
}

func instruccionesJSON(p Pregunta) string {
	b, _ := json.Marshal(p.Instrucciones)
	return strings.Trim(string(b), `"`)
}

func TestDecirSinPreguntasFalla(t *testing.T) {
	c, _ := nuevoFalso(t, respuestaDos, http.StatusOK)
	if _, err := c.Decidir(context.Background(), "x", nil); err == nil {
		t.Fatal("sin preguntas debería fallar")
	}
}

func TestDecirErrorHTTP(t *testing.T) {
	c, _ := nuevoFalso(t, `{"detail": "bad question"}`, http.StatusUnprocessableEntity)
	_, err := c.Decidir(context.Background(), "x", map[string]Pregunta{"q": {Tipo: "noul"}})
	if err == nil || !strings.Contains(err.Error(), "422") {
		t.Fatalf("quería error 422, obtuve %v", err)
	}
}

func TestDecirClaveAPI(t *testing.T) {
	c, f := nuevoFalso(t, respuestaDos, http.StatusOK)
	c.APIKey = "secreta"
	if _, err := c.Decidir(context.Background(), "x", map[string]Pregunta{"q": {Tipo: "noul"}}); err != nil {
		t.Fatal(err)
	}
	if got := f.cabeceras.Get("Authorization"); got != "Bearer secreta" {
		t.Fatalf("authorization = %q", got)
	}
}

func TestNuevoQuitaBarraFinal(t *testing.T) {
	c := Nuevo("http://ejemplo:8080/", "", time.Second)
	if c.URL != "http://ejemplo:8080" {
		t.Fatalf("URL = %q", c.URL)
	}
}

func TestChoiceScoreNoulUsanUnaPregunta(t *testing.T) {
	const resp = `{"model":"m","answers":{"q":{"type":"choice","choice":"B","probabilities":{"A":0.3,"B":0.7},"confidence":0.6}},"usage":{"input_tokens":9,"output_tokens":0}}`
	c, f := nuevoFalso(t, resp, http.StatusOK)

	r, err := c.Choice(context.Background(), "estado", "instrucciones", map[string]any{"A": "a", "B": "b"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Choice != "B" {
		t.Fatalf("choice = %q", r.Choice)
	}
	var cuerpo map[string]any
	if err := json.Unmarshal([]byte(f.peticion), &cuerpo); err != nil {
		t.Fatal(err)
	}
	preguntas := cuerpo["questions"].(map[string]any)
	if len(preguntas) != 1 {
		t.Fatalf("choice debería enviar 1 pregunta, envió %d", len(preguntas))
	}
	q := preguntas["q"].(map[string]any)
	if q["type"] != "choice" {
		t.Fatalf("type = %v", q["type"])
	}

	// Score y noul: mismo id, distinto tipo.
	for _, caso := range []struct {
		nombre string
		llamar func() error
		tipo   string
	}{
		{"score", func() error {
			_, err := c.Score(context.Background(), "e", "i", []any{"bajo", "alto"})
			return err
		}, "score"},
		{"noul", func() error {
			_, err := c.Noul(context.Background(), "e", "i", nil)
			return err
		}, "noul"},
	} {
		if err := caso.llamar(); err != nil {
			t.Fatal(err)
		}
		var c2 map[string]any
		if err := json.Unmarshal([]byte(f.peticion), &c2); err != nil {
			t.Fatal(err)
		}
		q2 := c2["questions"].(map[string]any)["q"].(map[string]any)
		if q2["type"] != caso.tipo {
			t.Fatalf("%s: type = %v, quería %s", caso.nombre, q2["type"], caso.tipo)
		}
	}
}
