package v2

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rag-go/internal/v2/tipos"
)

func TestReglasUsaLaPropuestaDelCodigo(t *testing.T) {
	e := tipos.Estado{Inferencias: map[string]tipos.Dato{PrefijoRegla + DecSuficiencia: {Valor: Insuficiente, Confianza: 0.5}}}
	ds, err := Reglas{}.Decidir(context.Background(), e, []tipos.PreguntaDecision{
		{Nombre: DecSuficiencia, Opciones: []string{Suficiente, Insuficiente}},
		{Nombre: "otra", Opciones: []string{"a", "b"}},
	})
	if err != nil || len(ds) != 2 || ds[0].Eleccion != Insuficiente || ds[1].Eleccion != "a" || ds[0].Fuente != MotorReglas {
		t.Fatalf("%+v %v", ds, err)
	}
}

// servidorJev simula /v1/systemone: responde la elección dada o tarda.
func servidorJev(t *testing.T, eleccion string, demora time.Duration) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/systemone" {
			http.NotFound(w, r)
			return
		}
		var sol struct {
			Questions map[string]struct {
				Type     string         `json:"type"`
				Criteria map[string]any `json:"criteria"`
			} `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&sol); err != nil {
			http.Error(w, err.Error(), 422)
			return
		}
		if demora > 0 {
			select {
			case <-time.After(demora):
			case <-r.Context().Done():
				return
			}
		}
		answers := map[string]any{}
		for id, q := range sol.Questions {
			if q.Type != "choice" || len(q.Criteria) == 0 {
				http.Error(w, "esperaba choice con criterios", 422)
				return
			}
			answers[id] = map[string]any{"type": "choice", "choice": eleccion, "confidence": 0.8}
		}
		json.NewEncoder(w).Encode(map[string]any{"model": "jev-prueba", "answers": answers})
	}))
	t.Cleanup(s.Close)
	return s
}

func TestJevElegidaDentroDeLasOpciones(t *testing.T) {
	s := servidorJev(t, "CONCEPT", 0)
	m := NuevoJev(MotorLocal, s.URL, "", time.Second)
	ds, err := m.Decidir(context.Background(), tipos.Estado{Pregunta: "¿qué es?"},
		[]tipos.PreguntaDecision{{Nombre: DecTipoRespuesta, Opciones: []string{"CONCEPT", "UNKNOWN"}}})
	if err != nil || ds[0].Eleccion != "CONCEPT" || ds[0].Fuente != MotorLocal || ds[0].Confianza != 0.8 {
		t.Fatalf("%+v %v", ds, err)
	}
}

func TestJevFueraDeOpcionesEsError(t *testing.T) {
	s := servidorJev(t, "PIZZA", 0)
	m := NuevoJev(MotorRemota, s.URL, "", time.Second)
	_, err := m.Decidir(context.Background(), tipos.Estado{},
		[]tipos.PreguntaDecision{{Nombre: DecTipoRespuesta, Opciones: []string{"CONCEPT", "UNKNOWN"}}})
	if !errors.Is(err, ErrFueraDeOpciones) {
		t.Fatalf("una elección fuera de las opciones es un error: %v", err)
	}
}

func TestEstadoParaModeloSinPropuestas(t *testing.T) {
	e := tipos.Estado{Inferencias: map[string]tipos.Dato{PrefijoRegla + "x": {Valor: "a"}, "modulo": {Valor: "P"}}}
	m := estadoParaModelo(e)
	inf := m["inferencias"].(map[string]tipos.Dato)
	if _, ok := inf[PrefijoRegla+"x"]; ok || inf["modulo"].Valor != "P" {
		t.Fatalf("el modelo decide sin ver la propuesta de las reglas: %+v", inf)
	}
}

// En el orquestador: fuera de opciones, error y timeout → decide Reglas y queda anotado.
func TestDecidirConRespaldoAReglas(t *testing.T) {
	casos := map[string]tipos.DecisionEngine{
		"fuera de opciones": &Simulada{Elecciones: map[string]string{DecSuficiencia: "quizas"}},
		"error":             &Simulada{Err: errCaido},
		"timeout":           &Simulada{Demora: 2 * time.Second},
		"jev fuera":         NuevoJev(MotorLocal, servidorJev(t, "PIZZA", 0).URL, "", time.Second),
		"jev lento":         NuevoJev(MotorLocal, servidorJev(t, Suficiente, 2*time.Second).URL, "", 5*time.Second),
	}
	for nombre, motor := range casos {
		b := armarAgente()
		b.ag.Decision = motor
		b.ag.Config.Limites.TimeoutDecisionMs = 50
		tr := &turno{a: b.ag, ctx: context.Background(), pres: NuevoPresupuesto(b.ag.Config.Limites),
			estado: tipos.Estado{Hechos: map[string]tipos.Dato{}, Inferencias: map[string]tipos.Dato{}}}
		t0 := time.Now()
		d := tr.decidir(DecSuficiencia, []string{Suficiente, Insuficiente}, Insuficiente, 0.5)
		if d.Eleccion != Insuficiente || d.Fuente != MotorReglas {
			t.Errorf("%s: debió decidir Reglas con la propuesta: %+v", nombre, d)
		}
		if tr.respaldoDecision == "" || len(tr.decisiones) != 1 || tr.decisiones[0]["respaldo"] == nil {
			t.Errorf("%s: el respaldo debe quedar anotado: %+v", nombre, tr.decisiones)
		}
		if time.Since(t0) > time.Second {
			t.Errorf("%s: el timeout DECISION_TIMEOUT_MS no cortó (%s)", nombre, time.Since(t0))
		}
		if _, ok := tr.estado.Inferencias[PrefijoRegla+DecSuficiencia]; ok {
			t.Errorf("%s: la propuesta no debe quedar en el estado", nombre)
		}
	}
}

func TestDecidirTopeDeDecisiones(t *testing.T) {
	b := armarAgente()
	sim := &Simulada{Elecciones: map[string]string{DecSuficiencia: Suficiente}}
	b.ag.Decision = sim
	tr := &turno{a: b.ag, ctx: context.Background(), pres: NuevoPresupuesto(b.ag.Config.Limites),
		estado: tipos.Estado{Hechos: map[string]tipos.Dato{}, Inferencias: map[string]tipos.Dato{}}}
	for i := 0; i < 10; i++ {
		tr.decidir(DecSuficiencia, []string{Suficiente, Insuficiente}, Insuficiente, 0.5)
	}
	if sim.Llamadas() != 4 || len(tr.decisiones) != 10 {
		t.Fatalf("MAX_DECISION_CALLS=4: el motor se llamó %d veces", sim.Llamadas())
	}
}

func TestNuevoMotor(t *testing.T) {
	c := ConfigDefecto()
	for motor, nombre := range map[string]string{"": MotorReglas, MotorReglas: MotorReglas, MotorLocal: MotorLocal, MotorSimulada: MotorSimulada} {
		c.MotorDecision = motor
		m, err := NuevoMotor(c)
		if err != nil || m.Nombre() != nombre {
			t.Errorf("%q: %v %v", motor, m, err)
		}
	}
	c.MotorDecision = MotorLocal
	if m, _ := NuevoMotor(c); m.(*Jev).Cliente.URL != URLDecisionLocal {
		t.Errorf("local sin URL usa %s", URLDecisionLocal)
	}
	c.MotorDecision = MotorRemota
	if _, err := NuevoMotor(c); err == nil {
		t.Error("remota sin DECISION_URL es un error")
	}
}
