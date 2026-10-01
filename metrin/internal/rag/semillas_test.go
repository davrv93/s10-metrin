package rag

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"rag-go/internal/almacen"
	"rag-go/internal/indexar"
)

// tresEjes: «titularidad» y «anular» caen en ejes propios; el resto, en otro.
type tresEjes struct{}

func (tresEjes) Nombre() string { return "tres-ejes" }
func (tresEjes) Embeber(_ context.Context, t string) ([]float32, error) {
	t = strings.ToLower(t)
	switch {
	case strings.Contains(t, "titularidad"):
		return []float32{1, 0, 0}, nil
	case strings.Contains(t, "anular"):
		return []float32{0, 1, 0}, nil
	}
	return []float32{0, 0, 1}, nil
}

const idSemilla = "semilla-0123456789abcdef"

func prepararSemillas(t *testing.T) (*RAG, *llmFijo) {
	ctx := context.Background()
	a, _ := almacen.Abrir("", tresEjes{})
	base := func(cita string) map[string]string {
		return map[string]string{"source": indexar.FuenteS10KB, "confianza": "oficial",
			"source_url": "https://documentacion.s10peru.com/boletaje/", "cita": cita, "title": cita}
	}
	tarjeta := base("Boletaje › Cambio de titularidad")
	tarjeta["section"] = "Cambio de titularidad"
	tarjeta["semilla"] = idSemilla
	tarjeta["semilla_tipo"] = "procedimiento"
	tarjeta["pasos"] = `[{"texto":"Seleccione la orden de venta.","fotos":["imagenes/boletaje/b2.png"]}]`
	a.Reemplazar(ctx, "s10kb:"+idSemilla, indexar.FuenteS10KB, "v1", []almacen.Trozo{{
		ID: almacen.IDTrozo("s10kb:"+idSemilla, 0), Busqueda: "¿Cómo cambio la titularidad de una venta?",
		Texto: "## Cambio de titularidad\nSeleccione la orden de venta.", Metadata: tarjeta,
	}})
	otra := base("Boletaje › Anulación de venta")
	otra["section"] = "Anulación de venta"
	a.Reemplazar(ctx, "s10kb:anular", indexar.FuenteS10KB, "v1", []almacen.Trozo{{
		ID: almacen.IDTrozo("s10kb:anular", 0), Texto: "Para anular una venta use el escenario Anulación.", Metadata: otra,
	}})
	l := &llmFijo{respuesta: "1. Seleccione la orden de venta."}
	return &RAG{Almacen: a, LLM: l, MaxDistancia: 0.8, RutaFallos: filepath.Join(t.TempDir(), "f.jsonl")}, l
}

func TestSemillaElegidaUsaSuSeccionConCapturas(t *testing.T) {
	r, l := prepararSemillas(t)
	res, err := r.Preguntar(context.Background(), "¿Cómo cambio la titularidad de una venta?", Opciones{Semilla: idSemilla})
	if err != nil || res.SinContexto {
		t.Fatalf("%v %+v", err, res)
	}
	if res.Plan.Clasificador != "semilla_validada" || res.Plan.Semilla != idSemilla || res.Plan.TipoConsulta != "procedimiento" {
		t.Fatalf("plan inesperado: %+v", res.Plan)
	}
	if len(res.Fuentes) != 1 || len(res.Fuentes[0].Pasos) != 1 || res.Fuentes[0].Pasos[0].Fotos[0] != "imagenes/boletaje/b2.png" {
		t.Fatalf("la fuente debía traer los pasos con su captura: %+v", res.Fuentes)
	}
	if prompt := l.visto[len(l.visto)-1].Content; strings.Contains(prompt, "anular una venta") {
		t.Fatal("con una semilla elegida el contexto es solo su sección")
	}
}

func TestPreguntaEscritaReconoceLaSemilla(t *testing.T) {
	r, _ := prepararSemillas(t)
	res, err := r.Preguntar(context.Background(), "titularidad de la venta, ¿cómo la cambio?", Opciones{})
	if err != nil || res.SinContexto {
		t.Fatalf("%v %+v", err, res)
	}
	if res.Plan.Semilla != idSemilla || res.Plan.TipoConsulta != "procedimiento" {
		t.Fatalf("debía reconocer la pregunta validada: %+v", res.Plan)
	}
	if res.Plan.Clasificador == "semilla_validada" {
		t.Fatal("una pregunta escrita pasa por el clasificador, no por la ruta de semilla elegida")
	}
}

func TestSemillaDesconocidaSigueElFlujoNormal(t *testing.T) {
	r, _ := prepararSemillas(t)
	res, err := r.Preguntar(context.Background(), "¿Cómo anular una venta?", Opciones{Semilla: "semilla-no-existe"})
	if err != nil || res.SinContexto {
		t.Fatalf("%v %+v", err, res)
	}
	if res.Plan.Clasificador == "semilla_validada" || res.Plan.Semilla != "" {
		t.Fatalf("un id desconocido no debe forzar nada: %+v", res.Plan)
	}
	if !strings.Contains(res.Fuentes[0].Cita, "Anulación") {
		t.Fatalf("debía responder con la sección de anulación: %+v", res.Fuentes)
	}
}
