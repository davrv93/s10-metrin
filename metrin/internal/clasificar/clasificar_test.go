package clasificar

import (
	"context"
	"testing"
)

// falso es un embebedor determinista para pruebas: vectores fijos por texto.
type falso struct{ v map[string][]float32 }

func (falso) Nombre() string { return "falso" }

func (f falso) Embeber(_ context.Context, t string) ([]float32, error) {
	if v, ok := f.v[t]; ok {
		return append([]float32(nil), v...), nil
	}
	return []float32{0, 0, 1}, nil
}

func TestEntrenarYClasificar(t *testing.T) {
	e := falso{v: map[string][]float32{
		"hola": {1, 0, 0}, "buenas": {0.9, 0.1, 0},
		"chau": {0, 0.9, 0.1}, "gracias": {0.1, 0.9, 0},
	}}
	m, err := Entrenar(context.Background(), e, "test", 0.5, map[string][]string{
		"saludo":  {"hola", "buenas"},
		"gracias": {"chau", "gracias"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Intenciones) != 2 {
		t.Fatalf("intenciones = %d", len(m.Intenciones))
	}
	in, sim, err := m.Clasificar(context.Background(), e, "hola")
	if err != nil {
		t.Fatal(err)
	}
	if in != "saludo" || sim < 0.99 {
		t.Fatalf("hola -> %s (%.3f)", in, sim)
	}
	// Lejos de todo: ruta segura trabajo.
	in, _, err = m.Clasificar(context.Background(), e, "texto raro")
	if err != nil {
		t.Fatal(err)
	}
	if in != Trabajo {
		t.Fatalf("desconocido -> %s, esperaba trabajo", in)
	}
}

func TestGuardarCargar(t *testing.T) {
	e := falso{v: map[string][]float32{"hola": {1, 0}}}
	m, err := Entrenar(context.Background(), e, "test", 0.5,
		map[string][]string{"saludo": {"hola"}})
	if err != nil {
		t.Fatal(err)
	}
	ruta := t.TempDir() + "/m.json"
	if err := m.Guardar(ruta); err != nil {
		t.Fatal(err)
	}
	m2, err := Cargar(ruta)
	if err != nil {
		t.Fatal(err)
	}
	in, _, err := m2.Clasificar(context.Background(), e, "hola")
	if err != nil {
		t.Fatal(err)
	}
	if in != "saludo" {
		t.Fatalf("roundtrip -> %s", in)
	}
}
