package busqueda

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func indicePrueba(t *testing.T, cfg ConfigBM25, docs ...Doc) *Indice {
	t.Helper()
	ix := NuevoIndice(cfg)
	for _, d := range docs {
		if err := ix.Agregar(d); err != nil {
			t.Fatal(err)
		}
	}
	return ix
}

func soloTexto(id, texto string) Doc {
	return Doc{ID: id, Campos: map[string]string{"texto": texto}}
}

func ids(ps []Puntuado) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.ID
	}
	return out
}

// Corpus mínimo con puntajes calculados a mano (un campo, sin exactos ni
// difuso): N=3, k1=1.2, b=0.75.
//
//	a: «metrado partida»          (2 términos)
//	b: «partida partida obra»     (3 términos)
//	c: «obra»                     (1 término)
//
// avgdl = 2; idf(t) = ln(1 + (N − df + 0.5)/(df + 0.5)).
func TestBM25CorpusMinimoConocido(t *testing.T) {
	cfg := ConfigBM25{K1: 1.2, B: 0.75, PesoDefecto: 1, Pesos: map[string]float64{"texto": 1}}
	ix := indicePrueba(t, cfg,
		soloTexto("a", "metrado partida"),
		soloTexto("b", "partida partida obra"),
		soloTexto("c", "obra"),
	)
	idf := func(df float64) float64 { return math.Log(1 + (3-df+0.5)/(df+0.5)) }
	bm := func(tf, dl float64) float64 {
		norm := 1 - 0.75 + 0.75*dl/2
		s := tf / norm
		return s * 2.2 / (1.2 + s)
	}
	quiere := map[string]float64{
		"a": idf(2) * bm(1, 2),
		"b": idf(2) * bm(2, 3),
	}
	got := ix.BuscarTerminos([]TerminoConsulta{{"partid", 1}}, 10)
	if len(got) != 2 {
		t.Fatalf("resultados = %v", got)
	}
	for _, p := range got {
		if math.Abs(p.Score-quiere[p.ID]) > 1e-9 {
			t.Errorf("%s: score %.6f, quiere %.6f", p.ID, p.Score, quiere[p.ID])
		}
	}
	// b tiene tf=2 pero es más largo: con estos números gana igualmente.
	if !reflect.DeepEqual(ids(got), []string{"b", "a"}) {
		t.Errorf("orden = %v", ids(got))
	}
	// Dos términos: suma de las contribuciones.
	got = ix.BuscarTerminos([]TerminoConsulta{{"metrad", 1}, {"obra", 1}}, 10)
	q := map[string]float64{
		"a": idf(1) * bm(1, 2),
		"b": idf(2) * bm(1, 3),
		"c": idf(2) * bm(1, 1),
	}
	for _, p := range got {
		if math.Abs(p.Score-q[p.ID]) > 1e-9 {
			t.Errorf("%s: score %.6f, quiere %.6f", p.ID, p.Score, q[p.ID])
		}
	}
	if !reflect.DeepEqual(ids(got), []string{"a", "c", "b"}) {
		t.Errorf("orden = %v", ids(got))
	}
}

func TestBM25PesoDeCampo(t *testing.T) {
	ix := indicePrueba(t, ConfigPorDefecto(),
		Doc{ID: "en-texto", Campos: map[string]string{"titulo": "Datos generales", "texto": "El kardex muestra entradas y salidas del almacén de obra."}},
		Doc{ID: "en-titulo", Campos: map[string]string{"titulo": "Kardex de almacén", "texto": "Muestra entradas y salidas de los materiales de la obra."}},
	)
	got := ix.BuscarLexico("kardex", 5)
	if len(got) != 2 || got[0].ID != "en-titulo" {
		t.Errorf("el título debía pesar más que el texto: %v", got)
	}
}

func TestBM25TerminoExactoYAtajos(t *testing.T) {
	ix := indicePrueba(t, ConfigPorDefecto(),
		soloTexto("plural", "Los metrados de las partidas se procesan juntos."),
		soloTexto("exacto", "El metrado de la partida se procesa aparte."),
		soloTexto("f7", "Pulse F7 para abrir el catálogo de recursos."),
		soloTexto("f1", "Pulse F1 para abrir la ayuda."),
		soloTexto("ctrl", "Con Ctrl + F se busca un insumo en la hoja."),
		soloTexto("codigo", "La partida 01.01.01 está siendo utilizada en Gerencia de Proyectos."),
	)
	got := ix.BuscarLexico("metrado", 2)
	if len(got) != 2 || got[0].ID != "exacto" {
		t.Errorf("la forma exacta debía ganar a la plural: %v", got)
	}
	if got := ix.BuscarLexico("¿qué hace F7?", 5); len(got) != 1 || got[0].ID != "f7" {
		t.Errorf("F7: %v", got)
	}
	if got := ix.BuscarLexico("para qué sirve ctrl+f", 5); len(got) == 0 || got[0].ID != "ctrl" {
		t.Errorf("Ctrl+F: %v", got)
	}
	if got := ix.BuscarLexico("error partida 01.01.01", 5); len(got) == 0 || got[0].ID != "codigo" {
		t.Errorf("código 01.01.01: %v", got)
	}
}

func TestBM25DifusoCorrigeFaltas(t *testing.T) {
	var docs []Doc
	// La raíz correcta tiene que ser frecuente (≥ FactorDifuso veces la del
	// error) para que la corrección se dispare.
	for i := 0; i < 25; i++ {
		docs = append(docs, soloTexto(fmt.Sprintf("d%02d", i), "registro del presupuesto de obra"))
	}
	docs = append(docs, soloTexto("otro", "kardex de almacén"))
	ix := indicePrueba(t, ConfigPorDefecto(), docs...)
	if got := ix.BuscarLexico("prespuesto", 3); len(got) == 0 {
		t.Error("«prespuesto» no encontró «presupuesto»")
	}
	cfg := ConfigPorDefecto()
	cfg.Difuso = false
	ix2 := indicePrueba(t, cfg, docs...)
	if got := ix2.BuscarLexico("prespuesto", 3); len(got) != 0 {
		t.Errorf("sin difuso no debía encontrar nada: %v", got)
	}
}

func TestBM25IDDuplicado(t *testing.T) {
	ix := NuevoIndice(ConfigPorDefecto())
	if err := ix.Agregar(soloTexto("x", "uno")); err != nil {
		t.Fatal(err)
	}
	if err := ix.Agregar(soloTexto("x", "dos")); err == nil {
		t.Error("se esperaba error por id duplicado")
	}
}

func TestPersistenciaGob(t *testing.T) {
	ix := indicePrueba(t, ConfigPorDefecto(),
		Doc{ID: "a", Campos: map[string]string{"titulo": "Registro del presupuesto", "texto": "Datos generales"}, Meta: map[string]string{"m": "1"}},
		soloTexto("b", "kardex de almacén con F7"),
	)
	var buf bytes.Buffer
	if err := ix.Guardar(&buf); err != nil {
		t.Fatal(err)
	}
	ix2, err := CargarIndice(&buf)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"presupuesto", "F7", "kardx almacen"} {
		if a, b := ix.BuscarLexico(q, 5), ix2.BuscarLexico(q, 5); !reflect.DeepEqual(a, b) {
			t.Errorf("%q: %v ≠ %v", q, a, b)
		}
	}
	if m, _ := ix2.Meta("a"); m["m"] != "1" {
		t.Errorf("meta perdida: %v", m)
	}
}
