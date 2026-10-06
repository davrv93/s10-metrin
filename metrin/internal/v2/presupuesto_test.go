package v2

import (
	"fmt"
	"testing"

	"rag-go/internal/v2/tipos"
)

func TestPresupuestoCortaEnCadaTope(t *testing.T) {
	p := NuevoPresupuesto(LimitesDefecto())
	casos := []struct {
		gastar func() error
		max    int
		nombre string
	}{
		{p.Paso, 5, "MAX_AGENT_STEPS"},
		{p.Decision, 4, "MAX_DECISION_CALLS"},
		{p.Busqueda, 2, "MAX_SEARCH_ITERATIONS"},
		{p.Herramienta, 5, "MAX_TOOL_CALLS"},
		{p.Generacion, 1, "MAX_GENERATIONS"},
		{p.Regeneracion, 1, "MAX_REGENERATIONS"},
	}
	for _, c := range casos {
		for i := 0; i < c.max; i++ {
			if err := c.gastar(); err != nil {
				t.Fatalf("%s: la vuelta %d debía caber: %v", c.nombre, i+1, err)
			}
		}
		// Pasado el tope, falla siempre (y no cuenta más): no hay forma de seguir gastando.
		for i := 0; i < 3; i++ {
			err := c.gastar()
			l, ok := EsLimite(err)
			if !ok || l.Limite != c.nombre || l.Max != c.max {
				t.Fatalf("%s: esperaba ErrLimite, %v", c.nombre, err)
			}
		}
	}
	if p.Pasos != 5 || p.Busquedas != 2 || p.Regeneraciones != 1 {
		t.Fatalf("contadores: %+v", p)
	}
	if u := p.Uso(); u["pasos"] != "5/5" || u["busquedas"] != "2/2" {
		t.Fatalf("uso: %v", u)
	}
	if _, ok := EsLimite(fmt.Errorf("envuelto: %w", &ErrLimite{Limite: "X"})); !ok {
		t.Fatal("EsLimite debe ver un límite envuelto")
	}
}

func TestPresupuestoCeroNoPermiteNada(t *testing.T) {
	p := NuevoPresupuesto(tipos.Limites{})
	if p.Paso() == nil || p.Busqueda() == nil || p.Decision() == nil {
		t.Fatal("con topes en 0 no se gasta nada")
	}
}
