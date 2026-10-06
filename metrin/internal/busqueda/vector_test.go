package busqueda

import (
	"context"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"rag-go/internal/almacen"
	"rag-go/internal/indexar"
)

// embPrueba: bolsa de palabras con hash a 64 dimensiones, normalizada.
// Determinista y sin modelo, suficiente para probar el cableado.
type embPrueba struct{}

func (embPrueba) Nombre() string { return "prueba" }
func (embPrueba) Embeber(_ context.Context, t string) ([]float32, error) {
	v := make([]float32, 64)
	for _, w := range strings.Fields(Plegar(t)) {
		h := fnv.New32a()
		h.Write([]byte(w))
		v[h.Sum32()%64]++
	}
	var s float64
	for _, x := range v {
		s += float64(x * x)
	}
	if s == 0 {
		v[0] = 1
		return v, nil
	}
	for i := range v {
		v[i] = float32(float64(v[i]) / math.Sqrt(s))
	}
	return v, nil
}

// Con el almacén y el indexador de verdad: los ids que devuelve
// VectorAlmacen son los de CargarFragmentos (también para ids repetidos
// entre manuales) y el filtro por metadatos funciona igual en los dos lados.
func TestVectorAlmacenCasaConElIndiceLexico(t *testing.T) {
	dir := t.TempDir()
	kb := `{"id":"web-x-s001","manual":"Manual de Presupuestos","titulo":"Manual de Presupuestos › 3.3 Registro","seccion":"3.3 Registro","texto":"registro del presupuesto en datos generales","confianza":"oficial"}
{"id":"web-x-s001","manual":"Manual de Almacenes","titulo":"Manual de Almacenes › 3.6 Kardex","seccion":"3.6 Kardex","texto":"kardex del almacen historia del recurso","confianza":"oficial"}
{"id":"faq-1","manual":"Cortex","titulo":"Optimiza 360","texto":"consultora fundador presupuesto","busqueda":"quien fundo optimiza"}
`
	ruta := filepath.Join(dir, "fragmentos.jsonl")
	os.WriteFile(ruta, []byte(kb), 0o644)
	docs, err := CargarFragmentos(ruta)
	if err != nil {
		t.Fatal(err)
	}
	lex, _ := IndexarDocs(ConfigPorDefecto(), docs)
	alm, err := almacen.Abrir("", embPrueba{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := indexar.Ejecutar(context.Background(), indexar.KBJSONL{Ruta: ruta}, alm, nil); err != nil {
		t.Fatal(err)
	}
	vec := VectorAlmacen{A: alm}
	cs, err := vec.BuscarVector(context.Background(), "kardex del almacen", 3)
	if err != nil {
		t.Fatal(err)
	}
	var vids []string
	for _, c := range cs {
		if _, ok := lex.Meta(c.ID); !ok {
			t.Errorf("el id vectorial %q no existe en el índice léxico", c.ID)
		}
		vids = append(vids, c.ID)
	}
	var lids []string
	for _, d := range docs {
		lids = append(lids, d.ID)
	}
	sort.Strings(vids)
	sort.Strings(lids)
	if !reflect.DeepEqual(vids, lids) {
		t.Errorf("ids vector %v ≠ léxico %v", vids, lids)
	}
	// Filtro: solo el Manual de Almacenes, en las dos listas.
	b := NuevoBuscador(lex, vec, nil, nil, Opciones{TamCache: 4})
	f := map[string]string{"manual": "Manual de Almacenes"}
	rs, err := b.BuscarFiltrado(context.Background(), "registro kardex presupuesto", 5, f)
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 1 || rs[0].Meta["clave"] != "web-x-s001|Manual de Almacenes" || len(rs[0].Fuentes) != 2 {
		t.Errorf("filtrado: %+v", rs)
	}
	// Sin filtro sale también el de Presupuestos, y la caché distingue filtros.
	rs2, _ := b.Buscar(context.Background(), "registro kardex presupuesto", 5)
	if len(rs2) < 2 {
		t.Errorf("sin filtro: %+v", rs2)
	}
	// Las claves de metadatos son las del almacén (los filtros de rag valen).
	m, _ := lex.Meta(rs[0].ID)
	for _, k := range []string{"source", "confianza", "manual", "title", "page", "cita"} {
		if _, ok := m[k]; !ok {
			t.Errorf("falta la clave %q en Meta", k)
		}
	}
}

func TestFiltroSinVectorialFiltrado(t *testing.T) {
	// vectorFijo no implementa VectorialFiltrado: el Buscador filtra después.
	vec := &vectorFijo{lista: []Candidato{{ID: "ceo", Similitud: 0.5}, {ID: "metrado", Similitud: 0.4}}}
	lex2 := indicePrueba(t, ConfigPorDefecto(),
		Doc{ID: "metrado", Campos: map[string]string{"texto": "metrado"}, Meta: map[string]string{"manual": "MP"}},
		Doc{ID: "ceo", Campos: map[string]string{"texto": "fundador"}, Meta: map[string]string{"manual": "Cortex"}},
	)
	b := NuevoBuscador(lex2, vec, nil, nil, Opciones{})
	rs, err := b.BuscarFiltrado(context.Background(), "metrado fundador", 5, map[string]string{"manual": "MP"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 1 || rs[0].ID != "metrado" {
		t.Errorf("filtro posterior: %+v", rs)
	}
}
