package trocear

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTrocearSolapeYRunas(t *testing.T) {
	// 2000 runas multibyte (ñ = 2 bytes): si se troceara por bytes, los
	// tamaños y el solape saldrían a la mitad y habría UTF-8 roto.
	texto := strings.Repeat("ñandú", 400) // 5 runas × 400 = 2000
	trozos := Trocear(texto, 800, 150)
	// pasos de 650: 0-800, 650-1450, 1300-2000
	if len(trozos) != 3 {
		t.Fatalf("esperaba 3 trozos, hay %d", len(trozos))
	}
	for i, tr := range trozos {
		if !utf8.ValidString(tr) {
			t.Fatalf("trozo %d con UTF-8 roto", i)
		}
	}
	if n := utf8.RuneCountInString(trozos[0]); n != 800 {
		t.Errorf("trozo 0: %d runas, esperaba 800", n)
	}
	if n := utf8.RuneCountInString(trozos[2]); n != 700 {
		t.Errorf("trozo 2: %d runas, esperaba 700", n)
	}
	// Solape: las últimas 150 runas del trozo 0 son las primeras del trozo 1.
	r0, r1 := []rune(trozos[0]), []rune(trozos[1])
	if string(r0[650:]) != string(r1[:150]) {
		t.Error("el solape de 150 runas no coincide")
	}
}

func TestTrocearCorto(t *testing.T) {
	if got := Trocear("  hola  ", 800, 150); len(got) != 1 || got[0] != "hola" {
		t.Fatalf("got %q", got)
	}
	if got := Trocear("   \n ", 800, 150); got != nil {
		t.Fatalf("texto vacío debería dar nil, got %q", got)
	}
	// justo 800: un solo trozo, sin uno extra de solape
	if got := Trocear(strings.Repeat("a", 800), 800, 150); len(got) != 1 {
		t.Fatalf("800 runas → 1 trozo, hay %d", len(got))
	}
}

func TestLimpiarHTML(t *testing.T) {
	html := `<!doctype html><html><head><title>Hola &amp; adiós</title>
<style>.x{color:red}</style><script>alert("no")</script></head>
<body><!-- comentario secreto --><h1>Plan <b>Pro</b></h1><p>Precio&nbsp;en la hoja</p>
<script type="module">
  const k = "<p>dentro</p>";
</script><svg><text>icono</text></svg></body></html>`
	got := LimpiarHTML(html)
	for _, no := range []string{"color:red", "alert", "comentario", "dentro", "icono", "<", ">"} {
		if strings.Contains(got, no) {
			t.Errorf("sobra %q en %q", no, got)
		}
	}
	for _, si := range []string{"Hola & adiós", "Plan Pro", "Precio en la hoja"} {
		if !strings.Contains(got, si) {
			t.Errorf("falta %q en %q", si, got)
		}
	}
}

func TestPrepararSoloLimpiaHTML(t *testing.T) {
	md := "# Título\n\n<b>negrita literal</b> en markdown"
	if got := Preparar("a.md", md); !strings.Contains(got[0], "<b>") {
		t.Error("un .md no debe pasar por LimpiarHTML")
	}
	if got := Preparar("a.HTML", "<p>x</p>"); got[0] != "x" {
		t.Errorf("got %q", got)
	}
}

func TestEsTexto(t *testing.T) {
	for _, r := range []string{"a.md", "b/c.TSX", "x.astro", "y.yml", "d.csv"} {
		if !EsTexto(r) {
			t.Errorf("%s debería ser texto", r)
		}
	}
	for _, r := range []string{"a.png", "b.go", "Makefile", "c.pjge"} {
		if EsTexto(r) {
			t.Errorf("%s no debería ser texto", r)
		}
	}
}
