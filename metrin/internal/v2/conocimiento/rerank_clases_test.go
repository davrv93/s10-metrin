package conocimiento

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

// rerankSimulado: un logit por documento según una regla, con registro de lo que recibió.
type rerankSimulado struct {
	mu       sync.Mutex
	logit    func(doc string) float64
	err      error
	cuantos  int // devuelve este número de puntajes si > 0 (para probar respuestas incompletas)
	esperar  bool
	llamadas int
	docs     [][]string
	consulta []string
}

func (r *rerankSimulado) Reordenar(ctx context.Context, q string, docs []string) ([]float64, error) {
	r.mu.Lock()
	r.llamadas++
	r.docs = append(r.docs, append([]string(nil), docs...))
	r.consulta = append(r.consulta, q)
	r.mu.Unlock()
	if r.esperar {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if r.err != nil {
		return nil, r.err
	}
	n := len(docs)
	if r.cuantos > 0 {
		n = r.cuantos
	}
	out := make([]float64, n)
	for i := range out {
		out[i] = r.logit(docs[i%len(docs)])
	}
	return out, nil
}

// aFavor: logit 6 a los documentos que contienen s, −3 al resto.
func aFavor(s string) func(string) float64 {
	return func(d string) float64 {
		if strings.Contains(d, s) {
			return 6
		}
		return -3
	}
}

// El reranker de clase elige: lo que prefiere pasa al frente, con su logit en Rerank, el puntaje en la escala del
// constructor (sigmoide) y el rango léxico en la meta. Recibe el texto representativo, no el del índice.
func TestRerankClases_ReordenaProcedimientos(t *testing.T) {
	b := cargarFixture(t)
	q := consulta(b, "como registro el metrado de la partida")
	sin, degSin, _ := NuevoRecuperador(b).Buscar(context.Background(), q, ClaseProcedimiento, 3)
	if len(sin) < 2 || sin[0].ID != "presupuestos.registrar-metrado" || !tiene(degSin, DegSinReranker) {
		t.Fatalf("línea base: %v %+v", degSin, sin)
	}

	rr := &rerankSimulado{logit: aFavor("Modificar una partida")}
	r := NuevoRecuperador(b)
	r.Reranker = rr
	cs, deg, err := r.Buscar(context.Background(), q, ClaseProcedimiento, 3)
	if err != nil || tiene(deg, DegSinReranker) || tiene(deg, DegRerankerError) {
		t.Fatalf("err %v, degradado %v", err, deg)
	}
	if cs[0].ID != "presupuestos.modificar-partida" || cs[0].Rerank == nil || *cs[0].Rerank != 6 {
		t.Fatalf("el reranker no eligió: %+v", cs)
	}
	if want := redondear(1 / (1 + math.Exp(-6))); cs[0].Puntaje != want {
		t.Errorf("puntaje %v, quiero la sigmoide del logit %v", cs[0].Puntaje, want)
	}
	if cs[0].Meta["rango_lexico"] == "" || cs[0].Meta["rango_lexico"] == "1" || cs[0].Meta["cobertura"] == "" {
		t.Errorf("meta: %v", cs[0].Meta)
	}
	if rr.llamadas != 1 {
		t.Errorf("%d llamadas al reranker, quiero 1", rr.llamadas)
	}
	// Texto representativo: título (módulo), objetivo, aliases y preguntas; nunca vacío.
	var vio bool
	for _, d := range rr.docs[0] {
		if strings.HasPrefix(d, "Modificar una partida del presupuesto (Presupuestos)\nCambiar la descripción") &&
			strings.Contains(d, "modificar partida; editar partida") && strings.Contains(d, "cómo modifico una partida") {
			vio = true
		}
		if strings.TrimSpace(d) == "" {
			t.Errorf("documento vacío enviado al reranker")
		}
	}
	if !vio {
		t.Errorf("texto representativo inesperado: %q", rr.docs[0])
	}
	// La consulta del reranker es la pregunta, sin los alias de la expansión.
	if rr.consulta[0] != "como registro el metrado de la partida" {
		t.Errorf("consulta al reranker: %q", rr.consulta[0])
	}
}

// Un fallo del reranker (error, tiempo agotado, respuesta incompleta o no finita) no cambia NADA del resultado:
// mismos candidatos, mismo orden, mismos puntajes que sin reranker. Solo queda anotado.
func TestRerankClases_FalloNoCambiaNada(t *testing.T) {
	b := cargarFixture(t)
	ctx := context.Background()
	for _, q := range []string{"como registro metrado", "partida", "como configuro los datos adicionales"} {
		for _, clase := range []string{ClaseProcedimiento, ClaseConcepto} {
			c := consulta(b, q)
			base, _, _ := NuevoRecuperador(b).Buscar(ctx, c, clase, 5)
			fallos := map[string]*rerankSimulado{
				"error":      {err: errors.New("connection refused")},
				"incompleto": {logit: aFavor("x"), cuantos: 1},
				"no finito":  {logit: func(string) float64 { return math.NaN() }},
				"sin tiempo": {esperar: true},
			}
			for nombre, rr := range fallos {
				r := NuevoRecuperador(b)
				r.Reranker = rr
				cctx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
				cs, deg, err := r.Buscar(cctx, c, clase, 5)
				cancel()
				if err != nil {
					t.Fatalf("%s/%s/%s: error %v (debe degradar, no fallar)", q, clase, nombre, err)
				}
				if len(base) == 1 && nombre == "incompleto" {
					continue // un solo documento: un puntaje es la respuesta completa
				}
				if len(base) > 0 && !tiene(deg, DegRerankerError) {
					t.Errorf("%s/%s/%s: degradado %v", q, clase, nombre, deg)
				}
				if !reflect.DeepEqual(cs, base) {
					t.Errorf("%s/%s/%s: el fallo cambió el resultado:\n %+v\n %+v", q, clase, nombre, cs, base)
				}
			}
		}
	}
}

// Solo las clases pedidas pasan por el reranker; los errores frecuentes y la búsqueda interna del constructor
// (conRerank=false) nunca.
func TestRerankClases_SoloLasClasesPedidas(t *testing.T) {
	b := cargarFixture(t)
	ctx := context.Background()
	rr := &rerankSimulado{logit: aFavor("Modificar")}
	r := NuevoRecuperador(b)
	r.Reranker = rr
	r.ClasesRerank = map[string]bool{ClaseConcepto: true}
	cs, deg, _ := r.Buscar(ctx, consulta(b, "como registro metrado"), ClaseProcedimiento, 3)
	if rr.llamadas != 0 || cs[0].Rerank != nil || tiene(deg, DegSinReranker) || tiene(deg, DegRerankerError) {
		t.Errorf("procedimiento fuera de ClasesRerank: %d llamadas, %v, %+v", rr.llamadas, deg, cs[0])
	}
	if _, _, _ = r.Buscar(ctx, consulta(b, "que es metrado"), ClaseConcepto, 3); rr.llamadas != 1 {
		t.Errorf("concepto en ClasesRerank: %d llamadas", rr.llamadas)
	}
	r.ClasesRerank = nil // defecto: procedimiento y concepto
	for _, clase := range []string{ClaseError, ClaseFragmento} {
		antes := rr.llamadas
		if _, _, _ = r.Buscar(ctx, consulta(b, "no puedo guardar metrado"), clase, 3); rr.llamadas != antes {
			t.Errorf("%s no debe pasar por el reranker de clase", clase)
		}
	}
	antes := rr.llamadas
	if _, _, _ = r.buscar(ctx, consulta(b, "metrado"), ClaseProcedimiento, 1, false); rr.llamadas != antes {
		t.Errorf("la búsqueda interna sin reranker lo llamó")
	}
}

// TopRerank: solo los N primeros del orden léxico se reordenan; el resto queda detrás, en su orden.
func TestRerankClases_TopN(t *testing.T) {
	b := cargarFixture(t)
	rr := &rerankSimulado{logit: aFavor("Configurar")}
	r := NuevoRecuperador(b)
	r.Reranker = rr
	r.TopRerank = 1
	cs, _, _ := r.Buscar(context.Background(), consulta(b, "partida presupuesto metrado datos"), ClaseProcedimiento, 1)
	if len(rr.docs) != 1 || len(rr.docs[0]) != 1 {
		t.Fatalf("con TopRerank=1 y k=1 debe ir 1 documento: %v", rr.docs)
	}
	if cs[0].Rerank == nil {
		t.Errorf("el primero debe llevar el puntaje del reranker: %+v", cs[0])
	}
}

func TestTextoRerank(t *testing.T) {
	b := cargarFixture(t)
	c := b.TextoRerank(ClaseConcepto, "partida", 0)
	if c != "Partida (partidas): Unidad de trabajo del presupuesto." {
		t.Errorf("concepto: %q", c)
	}
	p := b.TextoRerank(ClaseProcedimiento, "presupuestos.modificar-partida", 0)
	if !strings.HasPrefix(p, "Modificar una partida del presupuesto (Presupuestos)\n") {
		t.Errorf("procedimiento: %q", p)
	}
	corto := b.TextoRerank(ClaseProcedimiento, "presupuestos.modificar-partida", 60)
	if n := utf8.RuneCountInString(corto); n > 60 || n < 40 || !strings.HasPrefix(p, corto) {
		t.Errorf("recorte a 60 runas: %d %q", n, corto)
	}
	if b.TextoRerank(ClaseProcedimiento, "no-existe", 0) != "" || b.TextoRerank(ClaseError, "x", 0) != "" {
		t.Error("id o clase desconocidos deben dar texto vacío")
	}
}

// La traza deja, por clase, si reordenó, el orden léxico, el del reranker y el fallo.
func TestRerankClases_Traza(t *testing.T) {
	b := cargarFixture(t)
	rec := traza.Nuevo("como registro metrado")
	ctx := traza.ConContexto(context.Background(), rec)
	r := NuevoRecuperador(b)
	r.Reranker = &rerankSimulado{logit: aFavor("Modificar")}
	_, _, _ = r.Buscar(ctx, consulta(b, "como registro el metrado de la partida"), ClaseProcedimiento, 3)
	r.Reranker = &rerankSimulado{err: errors.New("timeout")}
	_, _, _ = r.Buscar(ctx, consulta(b, "que es metrado"), ClaseConcepto, 3)
	x, ok := rec.V2().DatoDe(traza.EtapaV2Rerank, ClaveTrazaRerankClases)
	m, _ := x.(map[string]any)
	if !ok || len(m) != 2 {
		t.Fatalf("dato «clases»: %v", x)
	}
	p := m[ClaseProcedimiento].(traza.Datos)
	if p["reordenado"] != true || p["cambio_primero"] != true || p["fallo"] != nil || len(p["orden"].([]map[string]any)) == 0 {
		t.Errorf("procedimiento: %v", p)
	}
	c := m[ClaseConcepto].(traza.Datos)
	if c["reordenado"] != false || c["fallo"] == nil || len(c["orden_lexico"].([]string)) == 0 {
		t.Errorf("concepto: %v", c)
	}
}

// Constructor: con candidatos del reranker, elige el primero del reranker y el empate se mide en logits.
func TestConstructor_EmpateConReranker(t *testing.T) {
	b := cargarFixture(t)
	c := NuevoConstructor(b, nil)
	q := tipos.Consulta{Original: "como cambio algo de la partida"}
	mk := func(id string, logit, bm25 float64) candEf {
		v := logit
		return candEf{tipos.Candidato{ID: id, Clase: ClaseProcedimiento, Lexico: bm25, Rerank: &v}, aUnidad(v)}
	}
	// Sigmoides casi iguales (0,998 y 0,993) pero el reranker prefiere con claridad al primero: no es empate,
	// aunque el BM25 del segundo sea mayor (antes eso daba aclaración).
	cs := []candEf{mk("presupuestos.modificar-partida", 6.2, 3), mk("presupuestos.registrar-metrado", 5, 9)}
	if c.ambiguo(q, cs) {
		t.Error("logits separados por 1,2: no es empate")
	}
	cs = []candEf{mk("presupuestos.modificar-partida", 5.2, 3), mk("presupuestos.registrar-metrado", 5, 9)}
	if !c.ambiguo(q, cs) {
		t.Error("logits separados por 0,2 (< logit(0,6)): empate")
	}
	// Sin reranker en alguno de los dos, la regla de siempre.
	cs[1].Rerank = nil
	cs[1].ef = 0.9
	if c.ambiguo(q, cs) {
		t.Error("sin reranker en el segundo: decide la regla léxica (BM25 lejos por arriba no empata con ef lejos)")
	}
	if math.Abs(MargenRerankEmpate-math.Log(1.5)) > 1e-12 {
		t.Errorf("MargenRerankEmpate = %v", MargenRerankEmpate)
	}
}

// De punta a punta: con el reranker de clase el plan usa el procedimiento que elige el reranker; si falla, el de
// siempre.
func TestConstructor_PlanConRerankerYSinEl(t *testing.T) {
	b := cargarFixture(t)
	q := consulta(b, "como registro el metrado de la partida")
	e := tipos.Estado{Pregunta: q.Original, Consulta: q, Tipo: tipos.Procedimiento}
	plan := func(rr Reordenador) tipos.Plan {
		r := NuevoRecuperador(b)
		r.Reranker = rr
		cs, _, _ := r.Buscar(context.Background(), q, ClaseProcedimiento, 5)
		p, err := NuevoConstructor(b, r).Construir(context.Background(), e, cs)
		if err != nil || p.Procedimiento == nil {
			t.Fatalf("plan: %v %+v", err, p)
		}
		return p
	}
	if p := plan(&rerankSimulado{logit: aFavor("Modificar")}); p.Procedimiento.ID != "presupuestos.modificar-partida" {
		t.Errorf("con reranker: %s", p.Procedimiento.ID)
	}
	if p := plan(&rerankSimulado{err: errors.New("caído")}); p.Procedimiento.ID != "presupuestos.registrar-metrado" {
		t.Errorf("reranker caído: %s", p.Procedimiento.ID)
	}
	if p := plan(nil); p.Procedimiento.ID != "presupuestos.registrar-metrado" {
		t.Errorf("sin reranker: %s", p.Procedimiento.ID)
	}
	// El reranker que no ve relación con nada (logits negativos) NO rechaza lo que la cobertura léxica ya
	// aceptaba: «registrar metrado» cubre la pregunta entera y se acepta igual.
	r := NuevoRecuperador(b)
	r.Reranker = &rerankSimulado{logit: func(string) float64 { return -4 }}
	cs, _, _ := r.Buscar(context.Background(), q, ClaseProcedimiento, 5)
	p, _ := NuevoConstructor(b, r).Construir(context.Background(), e, cs)
	if p.Procedimiento == nil || p.Procedimiento.ID != "presupuestos.registrar-metrado" {
		t.Errorf("logits negativos y cobertura completa: %+v", p.Procedimiento)
	}
}

// Aceptación con el reranker de clase: el puntaje efectivo es el mayor entre la sigmoide del logit y la cobertura
// léxica del mismo candidato. El reranker suma (acepta lo que la cobertura no alcanzaba) y no quita.
func TestConstructor_AceptacionConReranker(t *testing.T) {
	b := cargarFixture(t)
	c := NuevoConstructor(b, nil)
	mk := func(logit float64, cub string) tipos.Candidato {
		v := logit
		return tipos.Candidato{ID: "presupuestos.modificar-partida", Clase: ClaseProcedimiento, Puntaje: redondear(aUnidad(v)),
			Lexico: 2, Rerank: &v, Meta: map[string]string{"cobertura": cub}}
	}
	casos := []struct {
		nombre string
		x      tipos.Candidato
		ef     float64
		acepta bool
	}{
		{"el reranker suma: cobertura baja, logit alto", mk(3, "0.200"), redondear(aUnidad(3)), true},
		{"el reranker no quita: logit negativo, cobertura alta", mk(-2, "0.900"), 0.9, true},
		{"ninguno de los dos", mk(-4, "0.200"), 0.2, false},
	}
	for _, x := range casos {
		var inf Informe
		cs := c.candidatos(context.Background(), tipos.Consulta{Original: "q"}, []tipos.Candidato{x.x}, ClaseProcedimiento, 5, &inf)
		if len(cs) != 1 || cs[0].ef != x.ef {
			t.Errorf("%s: ef %+v, quiero %v", x.nombre, cs, x.ef)
			continue
		}
		if got := c.acepta(cs, 0, c.UmbralProcedimiento, porID); got != x.acepta {
			t.Errorf("%s: acepta = %v", x.nombre, got)
		}
	}
	// Un fragmento reordenado por la híbrida sigue con su escala de siempre (la sigmoide), sin el piso léxico.
	v := -2.0
	fr := tipos.Candidato{ID: "f", Clase: ClaseFragmento, Puntaje: redondear(aUnidad(v)), Rerank: &v, Meta: map[string]string{"cobertura": "0.900"}}
	var inf Informe
	if cs := c.candidatos(context.Background(), tipos.Consulta{Original: "q"}, []tipos.Candidato{fr}, ClaseFragmento, 5, &inf); cs[0].ef != fr.Puntaje {
		t.Errorf("fragmento: ef %v, quiero %v", cs[0].ef, fr.Puntaje)
	}
}
