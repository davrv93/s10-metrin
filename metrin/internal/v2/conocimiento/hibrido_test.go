package conocimiento

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"rag-go/internal/busqueda"
	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

// hibridoFalso devuelve un informe fijo (o un error) y recuerda la consulta que recibió.
type hibridoFalso struct {
	inf     busqueda.Informe
	err     error
	consult string
	k       int
}

func (h *hibridoFalso) BuscarInforme(_ context.Context, q string, k int) (busqueda.Informe, error) {
	h.consult, h.k = q, k
	return h.inf, h.err
}

// resultado arma un busqueda.Resultado como los de internal/busqueda (Meta con id_original y manual).
func resultado(id, manual string, rrf, vector float64, rangoV, rangoL int) busqueda.Resultado {
	r := busqueda.Resultado{ID: id, Score: rrf, RRF: rrf, Vector: vector, RangoVector: rangoV, RangoBM25: rangoL,
		Meta: map[string]string{"id_original": id, "manual": manual}}
	if rangoV > 0 {
		r.Fuentes = append(r.Fuentes, "vector")
	}
	if rangoL > 0 {
		r.Fuentes = append(r.Fuentes, "bm25")
		r.BM25 = 3
	}
	return r
}

func informe(rs ...busqueda.Resultado) busqueda.Informe {
	return busqueda.Informe{Resultados: rs, Tiempos: map[string]time.Duration{
		"lexico": time.Millisecond, "vector": 2 * time.Millisecond, "fusion": time.Microsecond, "total": 3 * time.Millisecond}}
}

func ids(cs []tipos.Candidato) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.ID + "@" + c.Meta["manual"]
	}
	return out
}

// La clase «fragmento» con la híbrida: solo secciones oficiales con pasos (el marketing y las capturas quedan
// fuera aunque la híbrida los traiga), el par (id, manual) decide (el mismo id en dos manuales) y el puntaje es la
// cobertura léxica de siempre (la escala de los umbrales del constructor).
func TestHibrido_SoloFragmentosConPasosYMismaEscala(t *testing.T) {
	b := cargarFixture(t)
	q := consulta(b, "como registro una guia de remision")
	lexica, _, _ := NuevoRecuperador(b).Buscar(context.Background(), q, ClaseFragmento, 5)

	h := &hibridoFalso{inf: informe(
		resultado("web-https-s10peru-com-servicios", "Optimiza 360 · Servicios: ERP S10", 0.033, 0.6, 1, 1),
		resultado("img-bbb222bbb222-0000", "Manual de Presupuestos", 0.032, 0.5, 2, 0),
		resultado(a020, "Manual de Almacenes", 0.031, 0.4, 3, 2),
		resultado(s050, "Manual de Almacenes", 0.016, 0.3, 4, 0),
	)}
	r := NuevoRecuperador(b)
	r.Hibrido = h
	cs, deg, err := r.Buscar(context.Background(), q, ClaseFragmento, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) == 0 || cs[0].ID != a020 || cs[0].Meta["manual"] != "Manual de Almacenes" {
		t.Fatalf("primero: %v", ids(cs))
	}
	if cs[0].Puntaje != lexica[0].Puntaje || cs[0].Meta["origen"] != "hibrida" || cs[0].Meta["rango_hibrido"] != "3" || cs[0].Vector != 0.4 {
		t.Errorf("escala o metadatos: híbrida %+v, léxica %+v", cs[0], lexica[0])
	}
	for _, c := range cs {
		if c.Meta["manual"] == "Optimiza 360 · Servicios: ERP S10" || c.ID == "img-bbb222bbb222-0000" {
			t.Errorf("candidato que no es una sección oficial con pasos: %+v", c)
		}
		if c.Clase != ClaseFragmento || c.Meta["cobertura"] == "" {
			t.Errorf("candidato sin clase o sin cobertura: %+v", c)
		}
	}
	if tiene(deg, DegFragmentosLex) || !tiene(deg, DegHibridoSinRerank) {
		t.Errorf("degradado: %v", deg)
	}
	if h.consult != "como registro una guia de remision" || h.k < 5 {
		t.Errorf("consulta a la híbrida: %q k=%d", h.consult, h.k)
	}
}

// Manda el rango híbrido (no el puntaje); lo que la léxica propia encontraba sigue siendo candidato aunque la
// híbrida no lo traiga; y un fragmento sin ningún término de la consulta solo entra si el reranker lo avala.
func TestHibrido_OrdenUnionYReranker(t *testing.T) {
	b := cargarFixture(t)
	ctx := context.Background()
	q := consulta(b, "registro")
	lex, _, _ := NuevoRecuperador(b).Buscar(ctx, q, ClaseFragmento, 5)
	if len(lex) < 2 || lex[0].Puntaje != lex[1].Puntaje {
		t.Skipf("el fixture ya no da un empate: %+v", lex)
	}
	orden := func(rs ...busqueda.Resultado) []string {
		r := NuevoRecuperador(b)
		r.Hibrido = &hibridoFalso{inf: informe(rs...)}
		cs, _, _ := r.Buscar(ctx, q, ClaseFragmento, 5)
		return ids(cs)
	}
	a := orden(resultado(a020, "Manual de Almacenes", 0.03, 0, 0, 1), resultado(s050, "Manual de Presupuestos", 0.02, 0, 0, 2))
	z := orden(resultado(s050, "Manual de Presupuestos", 0.03, 0, 0, 1), resultado(a020, "Manual de Almacenes", 0.02, 0, 0, 2))
	if len(a) < 2 || len(z) < 2 || a[0] != a020+"@Manual de Almacenes" || z[0] != s050+"@Manual de Presupuestos" {
		t.Errorf("a igual puntaje manda el rango híbrido: %v / %v", a, z)
	}
	// Unión: la híbrida no trae nada usable y la léxica propia sí.
	if u := orden(resultado("web-https-s10peru-com-servicios", "Optimiza 360 · Servicios: ERP S10", 0.03, 0.9, 1, 1)); len(u) != len(lex) {
		t.Errorf("la híbrida debe sumar, no quitar: %v frente a %v", u, ids(lex))
	}

	// Sin términos de la consulta: fuera sin reranker; dentro (con el puntaje del reranker en [0, 1]) si reordenó.
	q2 := consulta(b, "criptomonedas")
	r := NuevoRecuperador(b)
	r.Hibrido = &hibridoFalso{inf: informe(resultado(a020, "Manual de Almacenes", 0.03, 0.7, 1, 0))}
	if cs, _, _ := r.Buscar(ctx, q2, ClaseFragmento, 5); len(cs) != 0 {
		t.Errorf("solo por vector y sin reranker no es evidencia: %v", ids(cs))
	}
	rr := resultado(a020, "Manual de Almacenes", 4, 0.7, 1, 0)
	rr.Reordenado, rr.Rerank = true, 4
	r.Hibrido = &hibridoFalso{inf: informe(rr)}
	cs, deg, _ := r.Buscar(ctx, q2, ClaseFragmento, 5)
	if len(cs) != 1 || cs[0].Rerank == nil || cs[0].Puntaje != redondear(aUnidad(4)) || tiene(deg, DegHibridoSinRerank) {
		t.Errorf("con reranker: %+v %v", cs, deg)
	}
}

// Si la híbrida falla, la clase vuelve a la léxica propia y lo dice; un vector caído dentro de la híbrida se informa.
func TestHibrido_Degradaciones(t *testing.T) {
	b := cargarFixture(t)
	ctx := context.Background()
	q := consulta(b, "como registro una guia de remision")
	lex, _, _ := NuevoRecuperador(b).Buscar(ctx, q, ClaseFragmento, 5)

	r := NuevoRecuperador(b)
	r.Hibrido = &hibridoFalso{err: errors.New("busqueda: sin índice léxico ni vectorial")}
	cs, deg, err := r.Buscar(ctx, q, ClaseFragmento, 5)
	if err != nil || len(cs) != len(lex) || cs[0].ID != lex[0].ID || !tiene(deg, DegHibridoError) || !tiene(deg, DegFragmentosLex) {
		t.Errorf("híbrida caída: %v %v %v", ids(cs), deg, err)
	}

	inf := informe(resultado(a020, "Manual de Almacenes", 0.016, 0, 0, 1))
	inf.Degradado, inf.FalloVector = []string{"sin_vector"}, "chromem: colección vacía"
	r.Hibrido = &hibridoFalso{inf: inf}
	_, deg, _ = r.Buscar(ctx, q, ClaseFragmento, 5)
	if !tiene(deg, DegVectorError) {
		t.Errorf("vector caído en la híbrida: %v", deg)
	}

	// Las otras clases no usan la híbrida.
	h := &hibridoFalso{err: errors.New("no debía llamarse")}
	r.Hibrido = h
	if _, deg, _ := r.Buscar(ctx, consulta(b, "como registro metrado"), ClaseProcedimiento, 3); tiene(deg, DegHibridoError) || h.consult != "" {
		t.Errorf("procedimientos con la híbrida: %v", deg)
	}
}

// La traza V2 lleva las cifras reales del informe en las cuatro etapas de búsqueda (sin cerrarlas).
func TestHibrido_Traza(t *testing.T) {
	b := cargarFixture(t)
	rec := traza.Nuevo("q")
	ctx := traza.ConContexto(context.Background(), rec)
	v := rec.V2()
	r := NuevoRecuperador(b)
	r.Hibrido = &hibridoFalso{inf: informe(
		resultado("web-https-s10peru-com-servicios", "Optimiza 360 · Servicios: ERP S10", 0.033, 0.6, 1, 1),
		resultado(a020, "Manual de Almacenes", 0.031, 0.4, 2, 2),
	)}
	for i := 0; i < 2; i++ {
		if _, _, err := r.Buscar(ctx, consulta(b, "como registro una guia de remision"), ClaseFragmento, 5); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{traza.EtapaV2BusquedaLexica, traza.EtapaV2BusquedaVectorial, traza.EtapaV2Fusion, traza.EtapaV2Rerank} {
		x, ok := v.DatoDe(id, ClaveTrazaHibrido)
		d, _ := x.(traza.Datos)
		if !ok || d == nil || d["llamadas"] != 2 || !v.Pendiente(id) {
			t.Fatalf("%s: %v (pendiente %v)", id, x, v.Pendiente(id))
		}
	}
	x, _ := v.DatoDe(traza.EtapaV2Fusion, ClaveTrazaHibrido)
	fu := x.(traza.Datos)
	cands, _ := fu["candidatos"].([]map[string]any)
	if fu["resultados"] != 2 || fu["con_pasos"] != 1 || len(cands) != 2 || cands[1]["id"] != a020 || cands[1]["manual"] != "Manual de Almacenes" ||
		cands[1]["clase"] != ClaseFragmento || cands[1]["con_pasos"] != true || cands[0]["con_pasos"] != false || fu["ms"] == nil {
		t.Errorf("fusión: %+v", fu)
	}
	x, _ = v.DatoDe(traza.EtapaV2BusquedaVectorial, ClaveTrazaHibrido)
	if dv := x.(traza.Datos); dv["resultados"] != 2 || dv["ms"] != 2.0 || dv["activo"] != true {
		t.Errorf("vectorial: %+v", dv)
	}
	x, _ = v.DatoDe(traza.EtapaV2Rerank, ClaveTrazaHibrido)
	if dr := x.(traza.Datos); dr["activo"] != false || dr["reordenado"] != false {
		t.Errorf("rerank: %+v", dr)
	}
}

// El orden es el de la híbrida y el constructor solo arma un procedimiento implícito si el MEJOR candidato pasa su
// umbral: una sección con mucha cobertura que la búsqueda pone detrás no se cuela.
func TestHibrido_ImplicitoSoloConElMejorCandidato(t *testing.T) {
	b := cargarFixture(t)
	ctx := context.Background()
	e := tipos.Estado{Pregunta: "como registro una guia de remision", Tipo: tipos.Procedimiento,
		Consulta: consulta(b, "como registro una guia de remision")}
	plan := func(rs ...busqueda.Resultado) tipos.Plan {
		r := NuevoRecuperador(b)
		r.Hibrido = &hibridoFalso{inf: informe(rs...)}
		p, err := NuevoConstructor(b, r).Construir(ctx, e, nil)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	notaIngreso := resultado(s050, "Manual de Almacenes", 0.033, 0.5, 1, 1) // con pasos, pero no habla de guías
	guia := resultado(a020, "Manual de Almacenes", 0.031, 0.4, 2, 2)
	if p := plan(guia, notaIngreso); p.Procedimiento == nil || p.Procedimiento.ID != IDImplicito(a020, "Manual de Almacenes") {
		t.Fatalf("la guía primero: %+v", p.Procedimiento)
	}
	if p := plan(notaIngreso, guia); p.Procedimiento != nil && strings.HasPrefix(p.Procedimiento.ID, PrefijoImplicito) {
		t.Fatalf("el mejor clasificado no pasa el umbral: no hay implícito (%+v)", p.Procedimiento)
	}
}
