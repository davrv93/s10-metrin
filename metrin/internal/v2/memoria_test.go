package v2

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"rag-go/internal/rag"
	"rag-go/internal/v2/tipos"
)

func TestInterpretarMemoria(t *testing.T) {
	activa := tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 1}
	suspendida := tipos.Memoria{Suspendido: &tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 2}}
	retomada := tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 2, Retomado: true}
	ambas := tipos.Memoria{ProcedimientoID: "config.unidades", PasoActual: 1, Suspendido: suspendida.Suspendido}
	casos := []struct {
		texto  string
		m      tipos.Memoria
		quiero AccionMemoria
	}{
		{"¿y luego?", activa, MemSiguiente},
		{"siguiente", activa, MemSiguiente},
		{"Listo", activa, MemSiguiente},
		{"ya", activa, MemSiguiente},
		{"listo, ¿qué sigue?", activa, MemSiguiente},
		{"ya lo hice, ¿y luego?", activa, MemSiguiente},
		{"no me sale", activa, MemErrorPaso},
		{"me sale un error al grabar", activa, MemErrorPaso},
		{"no encuentro el botón Metrado", activa, MemErrorPaso},
		{"¿Qué es un APU?", activa, MemCambioTema},
		{"¿Qué es un APU?", retomada, MemAbandonar},
		{"sí", suspendida, MemRetomar},
		{"listo", suspendida, MemRetomar},
		{"retomemos", suspendida, MemRetomar},
		{"volvamos a lo de antes", ambas, MemRetomar},
		{"¿Cómo emito una factura?", suspendida, MemAbandonar},
		{"¿y luego?", tipos.Memoria{}, MemSinProcedimiento},
		{"listo", tipos.Memoria{}, MemNinguna}, // acuse sin nada en curso: charla
		{"ok", tipos.Memoria{}, MemNinguna},
		{"¿Cómo registro un metrado?", tipos.Memoria{}, MemNinguna},
	}
	for _, c := range casos {
		if got := InterpretarMemoria(c.texto, c.m); got != c.quiero {
			t.Errorf("%q con %+v → %s, quería %s", c.texto, c.m, got, c.quiero)
		}
	}
}

func TestAvanzarSuspenderRetomar(t *testing.T) {
	m := tipos.Memoria{ProcedimientoID: "p", PasoActual: 1}
	m, fin := Avanzar(m, 3)
	if fin || m.PasoActual != 2 || !reflect.DeepEqual(m.PasosCompletados, []int{1}) {
		t.Fatalf("avanzar 1→2: %+v", m)
	}
	m, _ = Avanzar(m, 3)
	m, fin = Avanzar(m, 3)
	if !fin || m.PasoActual != 3 || !reflect.DeepEqual(m.PasosCompletados, []int{1, 2, 3}) {
		t.Fatalf("terminar: %+v %v", m, fin)
	}
	s := Suspender(tipos.Memoria{ProcedimientoID: "p", PasoActual: 2, Retomado: true})
	if s.ProcedimientoID != "" || s.Suspendido == nil || s.Suspendido.Retomado || s.Suspendido.PasoActual != 2 {
		t.Fatalf("suspender: %+v", s)
	}
	r := Retomar(s)
	if r.ProcedimientoID != "p" || !r.Retomado || r.Suspendido != nil || r.PasoActual != 2 {
		t.Fatalf("retomar: %+v", r)
	}
}

func TestMemoriaDelPlanYDelHilo(t *testing.T) {
	plan, _ := PlanPaso(procMetrado(), 2)
	m := MemoriaDePlan(plan, tipos.Memoria{})
	if m.ProcedimientoID != "metrados.registrar" || m.PasoActual != 2 || !reflect.DeepEqual(m.PasosCompletados, []int{1}) {
		t.Fatalf("memoria de un plan del paso 2: %+v", m)
	}
	if got := MemoriaDePlan(PlanSinEvidencia(tipos.Procedimiento), m); got.ProcedimientoID != "" {
		t.Fatalf("sin evidencia no hay procedimiento en curso: %+v", got)
	}
	b, _ := json.Marshal(plan)
	hilo := []rag.Turno{
		{Rol: "usuario", Texto: "¿Cómo registro un metrado?"},
		{Rol: "asistente", Texto: "Paso 2…", Plan: b},
		{Rol: "usuario", Texto: "¿y luego?"},
	}
	got, ok := MemoriaDeHilo(hilo)
	if !ok || got.ProcedimientoID != "metrados.registrar" || got.PasoActual != 2 {
		t.Fatalf("reconstruida del hilo: %+v %v", got, ok)
	}
	if _, ok := MemoriaDeHilo([]rag.Turno{{Rol: "asistente", Plan: json.RawMessage(`"basura"`)}}); ok {
		t.Fatal("un plan ilegible se ignora")
	}
}

func TestPlanesDeMemoriaNoInventan(t *testing.T) {
	def := procMetrado()
	p, ok := PlanPaso(def, 3)
	v := PasosVisibles(p)
	if !ok || len(p.Pasos) != 3 || len(v) != 1 || v[0].N != 3 || len(v[0].Fotos) != 1 || v[0].Fotos[0].ID != "img_3" {
		t.Fatalf("paso 3 con SU foto: %+v", p)
	}
	if len(p.Verificacion) != 1 || p.Siguiente.Plantilla != "VERIFICACION" {
		t.Fatalf("el último paso trae la verificación: %+v", p)
	}
	if len(p.Fuentes) != 1 || p.Fuentes[0].Manual != "Manual de Presupuestos" || !reflect.DeepEqual(p.Fuentes[0].Paginas, []int{11, 12}) {
		t.Fatalf("fuentes del paso: %+v", p.Fuentes)
	}
	p2, _ := PlanPaso(def, 2)
	if len(PasosVisibles(p2)[0].Fotos) != 0 || p2.Siguiente.Paso != 3 || !reflect.DeepEqual(p2.PasosMostrados, []int{2}) {
		t.Fatalf("el paso 2 no tiene foto en la fuente: %+v", p2)
	}
	e := PlanErrores(def, 2)
	if e.Tipo != tipos.Problema || len(e.Errores) != 1 || PasosVisibles(e)[0].N != 2 {
		t.Fatalf("errores frecuentes con el paso a la vista: %+v", e)
	}
	sin := def
	sin.ErroresFrecuentes = nil
	if e := PlanErrores(sin, 2); e.Plantilla != "VERIFICACION" || len(e.Prerrequisitos) != 1 || len(e.Errores) != 0 {
		t.Fatalf("sin errores documentados: prerrequisitos, nada inventado: %+v", e)
	}
	if m := MemoriaDePlan(PlanFin(def), tipos.Memoria{}); m.ProcedimientoID != "" {
		t.Fatalf("tras el último paso no queda procedimiento en curso: %+v", m)
	}
	if txt := RenderBasico(p2); !strings.Contains(txt, "2. Elija la partida") || strings.Contains(txt, "1. Abra") {
		t.Fatalf("el texto enseña solo lo mostrado: %q", txt)
	}
	if _, ok := PlanPaso(def, 9); ok {
		t.Fatal("un paso que no existe no se inventa")
	}
}
