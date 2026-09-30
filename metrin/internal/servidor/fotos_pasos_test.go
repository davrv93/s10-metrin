package servidor

import (
	"strings"
	"testing"

	"rag-go/internal/rag"
)

func fuenteBoletaje() []rag.Fuente {
	return []rag.Fuente{{
		Cita: "Boletaje › Cambio de titularidad y/o inmueble",
		Fotos: []string{"imagenes/boletaje/a1.png", "imagenes/boletaje/b2.png",
			"imagenes/boletaje/c3.png", "imagenes/boletaje/sin-paso.png"},
		Pasos: []rag.Paso{
			{Texto: "Acceda al escenario Cambio de titularidad y/o inmueble dentro del grupo Boletaje.", Fotos: []string{"imagenes/boletaje/a1.png"}},
			{Texto: "Seleccione la orden de venta que debe pasar de un cliente a otro.", Fotos: []string{"imagenes/boletaje/b2.png"}},
			{Texto: "Genere la Nota de Crédito en la orden de venta original.", Fotos: []string{"imagenes/boletaje/c3.png"}},
			{Texto: "Configure el catálogo de conceptos tributarios.", Fotos: []string{"imagenes/boletaje/sin-paso.png"}},
		},
	}}
}

func TestColocarFotosDebajoDeSuPaso(t *testing.T) {
	respuesta := strings.Join([]string{
		"**Titularidad** en Boletaje es el paso de una orden de venta de un titular a otro.",
		"",
		"1. **Acceder al escenario «Cambio de titularidad y/o inmueble»** dentro del grupo Boletaje.",
		"2. **Seleccionar la orden de venta** que debe pasar de un cliente a otro.",
		"3. **Generar la Nota de Crédito** en la orden de venta original.",
	}, "\n")
	fuentes := fuenteBoletaje()
	got := colocarFotos(respuesta, fuentes)
	want := strings.Join([]string{
		"**Titularidad** en Boletaje es el paso de una orden de venta de un titular a otro.",
		"",
		"1. **Acceder al escenario «Cambio de titularidad y/o inmueble»** dentro del grupo Boletaje.",
		"![Captura del manual](fotos/imagenes/boletaje/a1.png)",
		"2. **Seleccionar la orden de venta** que debe pasar de un cliente a otro.",
		"![Captura del manual](fotos/imagenes/boletaje/b2.png)",
		"3. **Generar la Nota de Crédito** en la orden de venta original.",
		"![Captura del manual](fotos/imagenes/boletaje/c3.png)",
	}, "\n")
	if got != want {
		t.Fatalf("respuesta:\n%s\n\nesperada:\n%s", got, want)
	}
	if f := fuentes[0].Fotos; len(f) != 1 || f[0] != "imagenes/boletaje/sin-paso.png" {
		t.Fatalf("en la galería solo debe quedar la captura sin paso: %v", f)
	}
}

func TestColocarFotosSinParecidoNoToca(t *testing.T) {
	respuesta := "Para anular una venta use el escenario Anulación de venta."
	fuentes := fuenteBoletaje()
	if got := colocarFotos(respuesta, fuentes); got != respuesta {
		t.Fatalf("no debía insertar capturas: %q", got)
	}
	if len(fuentes[0].Fotos) != 4 {
		t.Fatalf("la galería debía quedar intacta: %v", fuentes[0].Fotos)
	}
}

func TestColocarFotosRechazaRutasRaras(t *testing.T) {
	fuentes := []rag.Fuente{{Pasos: []rag.Paso{{Texto: "Genere la Nota de Crédito en la orden de venta original.",
		Fotos: []string{"../secreto.png", "x.png)\n<script>"}}}}}
	respuesta := "1. Genere la Nota de Crédito en la orden de venta."
	if got := colocarFotos(respuesta, fuentes); got != respuesta {
		t.Fatalf("no debía insertar rutas inválidas: %q", got)
	}
}
