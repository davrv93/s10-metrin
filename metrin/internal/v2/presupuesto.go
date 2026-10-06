package v2

// Presupuesto del turno: la recursividad controlada. Cada acción gasta de su contador ANTES de hacerse; al pasar
// el tope devuelve *ErrLimite y el orquestador sale por la respuesta segura (o, para las decisiones, por las
// reglas). Ningún camino vuelve atrás sin gastar, así que un turno no puede entrar en bucle.
//
// Qué cuenta:
//   - paso (MAX_AGENT_STEPS): una vuelta del ciclo del agente: buscar, construir el plan, redactar o regenerar.
//     El camino más largo (buscar, buscar, construir, redactar, regenerar) usa exactamente 5;
//   - decisión (MAX_DECISION_CALLS): cada consulta al motor de decisión;
//   - búsqueda (MAX_SEARCH_ITERATIONS): cada iteración de búsqueda (la primera y cada expansión);
//   - herramienta (MAX_TOOL_CALLS): cada llamada al recuperador (una por índice consultado);
//   - generación (MAX_GENERATIONS) y regeneración (MAX_REGENERATIONS): redacciones del texto.

import (
	"errors"
	"fmt"

	"rag-go/internal/v2/tipos"
)

// ErrLimite: se agotó un contador del presupuesto.
type ErrLimite struct {
	Limite string // nombre de la variable de entorno
	Max    int
}

func (e *ErrLimite) Error() string {
	return fmt.Sprintf("límite %s=%d alcanzado", e.Limite, e.Max)
}

// EsLimite dice si err es (o envuelve) un límite del presupuesto.
func EsLimite(err error) (*ErrLimite, bool) {
	var l *ErrLimite
	ok := errors.As(err, &l)
	return l, ok
}

// Presupuesto lleva la cuenta de un turno. No es seguro para varias goroutines: es de un solo turno.
type Presupuesto struct {
	Lim            tipos.Limites
	Pasos          int
	Decisiones     int
	Busquedas      int
	Herramientas   int
	Generaciones   int
	Regeneraciones int
}

// NuevoPresupuesto crea el presupuesto de un turno.
func NuevoPresupuesto(l tipos.Limites) *Presupuesto { return &Presupuesto{Lim: l} }

func gastar(n *int, max int, nombre string) error {
	if *n >= max {
		return &ErrLimite{Limite: nombre, Max: max}
	}
	*n++
	return nil
}

func (p *Presupuesto) Paso() error { return gastar(&p.Pasos, p.Lim.MaxPasos, "MAX_AGENT_STEPS") }
func (p *Presupuesto) Decision() error {
	return gastar(&p.Decisiones, p.Lim.MaxDecisiones, "MAX_DECISION_CALLS")
}
func (p *Presupuesto) Busqueda() error {
	return gastar(&p.Busquedas, p.Lim.MaxBusquedas, "MAX_SEARCH_ITERATIONS")
}
func (p *Presupuesto) Herramienta() error {
	return gastar(&p.Herramientas, p.Lim.MaxHerramientas, "MAX_TOOL_CALLS")
}
func (p *Presupuesto) Generacion() error {
	return gastar(&p.Generaciones, p.Lim.MaxGeneraciones, "MAX_GENERATIONS")
}
func (p *Presupuesto) Regeneracion() error {
	return gastar(&p.Regeneraciones, p.Lim.MaxRegeneraciones, "MAX_REGENERATIONS")
}

// Uso resume lo gastado y los topes (para la traza).
func (p *Presupuesto) Uso() map[string]any {
	return map[string]any{
		"pasos":          fmt.Sprintf("%d/%d", p.Pasos, p.Lim.MaxPasos),
		"decisiones":     fmt.Sprintf("%d/%d", p.Decisiones, p.Lim.MaxDecisiones),
		"busquedas":      fmt.Sprintf("%d/%d", p.Busquedas, p.Lim.MaxBusquedas),
		"herramientas":   fmt.Sprintf("%d/%d", p.Herramientas, p.Lim.MaxHerramientas),
		"generaciones":   fmt.Sprintf("%d/%d", p.Generaciones, p.Lim.MaxGeneraciones),
		"regeneraciones": fmt.Sprintf("%d/%d", p.Regeneraciones, p.Lim.MaxRegeneraciones),
	}
}
