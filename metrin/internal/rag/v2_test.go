package rag

import (
	"context"
	"encoding/json"
	"testing"

	"rag-go/internal/v2/tipos"
)

// agenteFalso hace de orquestador V2 (el real vive en internal/v2, que importa este paquete).
type agenteFalso struct {
	version   tipos.Version
	vistas    []Opciones
	preguntas int
}

func (a *agenteFalso) Version(_ string, o Opciones) tipos.Version {
	a.vistas = append(a.vistas, o)
	return a.version
}

func (a *agenteFalso) Preguntar(_ context.Context, pregunta string, _ Opciones) (Respuesta, error) {
	a.preguntas++
	return Respuesta{Pregunta: pregunta, Respuesta: "v2", Version: tipos.V2, PlanV2: &tipos.Plan{Version: 2}}, nil
}

// Con un agente V2 que elige v1, Preguntar es V1 exacta; con v2, delega sin tocar nada de V1.
func TestHookV2(t *testing.T) {
	ref, _, _ := preparar(t, "Cuesta 49 € [s3:rag-demo/docs/tarifas.md#0]")
	sin, errSin := ref.Preguntar(context.Background(), "¿qué precio tiene el Pro?", Opciones{})

	r, l, _ := preparar(t, "Cuesta 49 € [s3:rag-demo/docs/tarifas.md#0]")
	ag := &agenteFalso{version: tipos.V1}
	r.V2 = ag
	o := Opciones{Version: "v2", Conversacion: "c1", Memoria: &tipos.Memoria{PasoActual: 1}}
	con, errCon := r.Preguntar(context.Background(), "¿qué precio tiene el Pro?", o)
	a, _ := json.Marshal(sin)
	b, _ := json.Marshal(con)
	if errSin != nil || errCon != nil || sinTiempos(a) != sinTiempos(b) || ag.preguntas != 0 {
		t.Fatalf("V1 con agente que elige v1 debe ser idéntica\nsin: %s\ncon: %s", a, b)
	}
	if len(ag.vistas) != 1 || ag.vistas[0].Version != "v2" || ag.vistas[0].Conversacion != "c1" {
		t.Fatalf("el agente decide con las opciones de la petición: %+v", ag.vistas)
	}

	ag.version = tipos.V2
	l.visto = nil
	res, err := r.Preguntar(context.Background(), "¿qué precio tiene el Pro?", o)
	if err != nil || res.Respuesta != "v2" || ag.preguntas != 1 || l.visto != nil {
		t.Fatalf("v2 delega y no llama al LLM de V1: %+v", res)
	}
	// PreguntarV1 ignora el interruptor.
	if res, _ := r.PreguntarV1(context.Background(), "¿qué precio tiene el Pro?", o); res.Version != "" || res.PlanV2 != nil {
		t.Fatalf("PreguntarV1: %+v", res)
	}
}

func TestRespuestaV1SinCamposV2EnJSON(t *testing.T) {
	r, _, _ := preparar(t, "Cuesta 49 €")
	res, _ := r.Preguntar(context.Background(), "¿qué precio tiene el Pro?", Opciones{})
	var m map[string]any
	b, _ := json.Marshal(res)
	json.Unmarshal(b, &m)
	for _, k := range []string{"plan", "memoria", "version"} {
		if _, ok := m[k]; ok {
			t.Errorf("V1 no lleva %q: %s", k, b)
		}
	}
	// Un turno del hilo con plan crudo no cambia el JSON de V1 ni falla al leerse.
	var turnos []Turno
	if err := json.Unmarshal([]byte(`[{"rol":"asistente","texto":"x","plan":"basura"}]`), &turnos); err != nil || string(turnos[0].Plan) != `"basura"` {
		t.Fatalf("plan crudo: %v %+v", err, turnos)
	}
}
