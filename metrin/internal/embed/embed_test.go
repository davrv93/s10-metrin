package embed

import (
	"context"
	"math"
	"os"
	"testing"
)

func TestEstaticoReal(t *testing.T) {
	ruta := "../../modelos/potion-es-int8.pjge"
	if _, err := os.Stat(ruta); err != nil {
		t.Skip("falta el modelo (ver README: copiarlo del Analista)")
	}
	e, err := NuevoEstatico(ruta)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a, _ := e.Embeber(ctx, "¿cuánto cuesta el plan Pro?")
	b, _ := e.Embeber(ctx, "Precio del plan Pro: 49 euros al mes")
	c, _ := e.Embeber(ctx, "La capital de Mongolia es Ulán Bator")
	var n float64
	for _, x := range a {
		n += float64(x) * float64(x)
	}
	if math.Abs(n-1) > 1e-4 {
		t.Fatalf("norma %f", n)
	}
	if coseno(a, b) <= coseno(a, c) {
		t.Fatalf("precio~precio %.3f debería superar precio~Mongolia %.3f", coseno(a, b), coseno(a, c))
	}
	if _, err := e.Embeber(ctx, "   "); err != ErrSinPiezas {
		t.Fatalf("texto vacío: %v", err)
	}
}

func coseno(a, b []float32) float64 {
	var s float64
	for i := range a {
		s += float64(a[i]) * float64(b[i])
	}
	return s
}
