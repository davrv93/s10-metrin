package busqueda

import (
	"reflect"
	"testing"
)

func textos(ts []Token) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.Texto
	}
	return out
}

func TestTokenizarPliegaYConservaEspeciales(t *testing.T) {
	casos := []struct {
		entrada string
		quiere  []string
	}{
		{"¿Cómo registro el METRADO?", []string{"como", "registro", "el", "metrado"}},
		{"Use Ctrl + F para buscar", []string{"use", "ctrl+f", "para", "buscar"}},
		{"Pulse F7 o Shift+F1", []string{"pulse", "f7", "shift+f1"}},
		{"Control+C y Mayús+F1", []string{"ctrl+c", "shift+f1"}},
		{"La partida 01.01.01 está en uso", []string{"la", "partida", "01.01.01", "esta", "en", "uso"}},
		{"Año: 2025, nómina", []string{"ano", "2025", "nomina"}},
	}
	for _, c := range casos {
		got := textos(Tokenizar(c.entrada))
		// «y», «o»: una letra, se descartan.
		if !reflect.DeepEqual(got, c.quiere) {
			t.Errorf("Tokenizar(%q) = %q, quiere %q", c.entrada, got, c.quiere)
		}
	}
}

func TestTokensEspeciales(t *testing.T) {
	for _, tk := range Tokenizar("F7 ctrl+f 3.3.1 metrado") {
		esp := tk.Texto != "metrado"
		if tk.Especial != esp {
			t.Errorf("%q: Especial=%v, quiere %v", tk.Texto, tk.Especial, esp)
		}
	}
}

func TestRaizLigera(t *testing.T) {
	iguales := [][]string{
		{"metrado", "metrados"},
		{"registro", "registrar", "registra", "registrado", "registros", "registre"},
		{"presupuesto", "presupuestos"},
		{"partida", "partidas"},
		{"almacen", "almacenes"},
		{"configuracion", "configurar"},
		{"aprobacion", "aprobar", "aprobado"},
		{"factura", "facturar", "facturacion"},
		{"valorizacion", "valorizar", "valorizado"},
		{"nomina", "nominas"},
	}
	for _, g := range iguales {
		r0 := Raiz(Plegar(g[0]))
		for _, w := range g[1:] {
			if r := Raiz(Plegar(w)); r != r0 {
				t.Errorf("Raiz(%q)=%q ≠ Raiz(%q)=%q", w, r, g[0], r0)
			}
		}
	}
	distintos := [][2]string{
		{"metrado", "metro"},
		{"partida", "parte"},
		{"obra", "obrero"},
	}
	for _, d := range distintos {
		if Raiz(d[0]) == Raiz(d[1]) {
			t.Errorf("Raiz funde %q y %q en %q", d[0], d[1], Raiz(d[0]))
		}
	}
	for _, w := range []string{"f7", "ctrl+f", "01.01.01", "s10", "apu", "uit"} {
		if Raiz(w) != w {
			t.Errorf("Raiz(%q) = %q, no debía cambiar", w, Raiz(w))
		}
	}
}

func TestTerminosQuitaVaciasYDaExactos(t *testing.T) {
	raices, exactos := Terminos("¿Qué es un metrado de partidas?")
	if !reflect.DeepEqual(raices, []string{"metrad", "partid"}) {
		t.Errorf("raices = %q", raices)
	}
	if !reflect.DeepEqual(exactos, []string{"=metrado", "=partidas"}) {
		t.Errorf("exactos = %q", exactos)
	}
}

func TestNormalizar(t *testing.T) {
	casos := map[string]string{
		"¿Qué es un METRADO?":              "qué es un metrado",
		"  qué   es un metrado ":           "qué es un metrado",
		"¿Cómo uso Ctrl+F?":                "cómo uso ctrl+f",
		"partida 01.01.01. Luego, guardar": "partida 01.01.01 luego guardar",
		"F7!":                              "f7",
	}
	for in, quiere := range casos {
		if got := Normalizar(in); got != quiere {
			t.Errorf("Normalizar(%q) = %q, quiere %q", in, got, quiere)
		}
	}
}
