package conocimiento

import (
	"context"
	"errors"
	"hash/fnv"
	"strings"
	"testing"

	"rag-go/internal/v2/tipos"
)

func consulta(b *Base, q string) tipos.Consulta {
	return b.Expandir(tipos.Consulta{Original: q, Normalizada: plegar(q)})
}

func primero(t *testing.T, r *Recuperador, b *Base, q, clase string) tipos.Candidato {
	t.Helper()
	cs, _, err := r.Buscar(context.Background(), consulta(b, q), clase, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) == 0 {
		t.Fatalf("%q (%s): sin candidatos", q, clase)
	}
	return cs[0]
}

// Consultas del megaprompt contra los fixtures.
func TestRecuperar_ConsultasMegaprompt(t *testing.T) {
	b := cargarFixture(t)
	r := NuevoRecuperador(b)
	casos := []struct{ q, clase, id string }{
		{"como registro metrado", ClaseProcedimiento, "presupuestos.registrar-metrado"},
		{"que es metrado", ClaseConcepto, "metrado"},
		{"donde veo metrados", ClaseProcedimiento, "presupuestos.registrar-metrado"},
		{"no puedo guardar metrado", ClaseError, "presupuestos.registrar-metrado#error1"},
		{"cómo modifico una partida", ClaseProcedimiento, "presupuestos.modificar-partida"},
		{"COMO MODIFICO UNA PARTIDA", ClaseProcedimiento, "presupuestos.modificar-partida"},
		{"como configuro los datos adicionales", ClaseProcedimiento, "presupuestos.configurar-datos-adicionales"},
		{"diferencia entre presupuesto meta y venta", ClaseConcepto, "presupuesto_meta"},
		{"como registro una guia de remision", ClaseFragmento, a020},
	}
	for _, c := range casos {
		got := primero(t, r, b, c.q, c.clase)
		if got.ID != c.id {
			t.Errorf("%q (%s): primero %s (%.2f), quiero %s", c.q, c.clase, got.ID, got.Puntaje, c.id)
		}
		if got.Puntaje <= 0 || got.Puntaje > 1 || got.Lexico <= 0 || got.Meta["cobertura"] == "" {
			t.Errorf("%q: puntajes fuera de escala: %+v", c.q, got)
		}
	}
	// El fragmento de marketing (Optimiza 360) nunca es candidato, aunque hable de guías de remisión.
	cs, _, _ := r.Buscar(context.Background(), consulta(b, "registrar guia de remision"), ClaseFragmento, 10)
	for _, c := range cs {
		if strings.Contains(c.Meta["manual"], "Optimiza") {
			t.Errorf("fragmento de marketing indexado: %+v", c)
		}
	}
	// Los fragmentos llevan su manual: el id solo no basta.
	if cs[0].Meta["manual"] != "Manual de Almacenes" {
		t.Errorf("fragmento sin manual: %+v", cs[0])
	}
	// La pregunta fuera de dominio no supera el umbral de procedimiento.
	if c := primeroOVacio(r, b, "receta de ceviche con limón", ClaseProcedimiento); c.Puntaje >= 0.6 {
		t.Errorf("fuera de dominio con puntaje alto: %+v", c)
	}
	if _, _, err := r.Buscar(context.Background(), consulta(b, "x"), "otra", 3); err == nil {
		t.Error("clase desconocida debe dar error")
	}
}

func primeroOVacio(r *Recuperador, b *Base, q, clase string) tipos.Candidato {
	cs, _, _ := r.Buscar(context.Background(), consulta(b, q), clase, 1)
	if len(cs) == 0 {
		return tipos.Candidato{}
	}
	return cs[0]
}

// vecFalso: bolsa de palabras con hash (determinista), para probar la fusión sin modelo.
type vecFalso struct {
	err    error
	llamas int
}

func (v *vecFalso) Embeber(_ context.Context, s string) ([]float32, error) {
	v.llamas++
	if v.err != nil {
		return nil, v.err
	}
	out := make([]float32, 64)
	for _, t := range terminos(s) {
		h := fnv.New32a()
		h.Write([]byte(t))
		out[h.Sum32()%64]++
	}
	return out, nil
}

type rerankFalso struct {
	err   error
	favor string
}

func (r rerankFalso) Reordenar(_ context.Context, _ string, docs []string) ([]float64, error) {
	if r.err != nil {
		return nil, r.err
	}
	out := make([]float64, len(docs))
	for i, d := range docs {
		if strings.Contains(d, r.favor) {
			out[i] = 5 // logit: pasa por sigmoide
		} else {
			out[i] = -2
		}
	}
	return out, nil
}

func tiene(xs []string, pref string) bool {
	for _, x := range xs {
		if strings.HasPrefix(x, pref) {
			return true
		}
	}
	return false
}

func TestRecuperar_Degradaciones(t *testing.T) {
	b := cargarFixture(t)
	ctx := context.Background()
	q := consulta(b, "como registro metrado")

	// Sin vector ni reranker: léxica, informado.
	r := NuevoRecuperador(b)
	cs, deg, _ := r.Buscar(ctx, q, ClaseProcedimiento, 3)
	if !tiene(deg, DegSinVector) || !tiene(deg, DegSinReranker) || cs[0].ID != "presupuestos.registrar-metrado" {
		t.Errorf("léxica: %v %v", deg, cs)
	}

	// Con vector: híbrido (RRF), sin «sin_vector».
	v := &vecFalso{}
	r = NuevoRecuperador(b)
	r.Vector = v
	cs, deg, _ = r.Buscar(ctx, q, ClaseProcedimiento, 3)
	if tiene(deg, DegSinVector) || !tiene(deg, DegSinReranker) || cs[0].ID != "presupuestos.registrar-metrado" || cs[0].Vector <= 0 {
		t.Errorf("híbrida: %v %+v", deg, cs)
	}
	antes := v.llamas
	_, _, _ = r.Buscar(ctx, q, ClaseProcedimiento, 3)
	if v.llamas-antes != 1 {
		t.Errorf("los vectores de los documentos se calculan una vez: %d llamadas", v.llamas-antes)
	}

	// Vector caído: vuelve a la léxica y lo dice.
	r = NuevoRecuperador(b)
	r.Vector = &vecFalso{err: errors.New("ollama caído")}
	cs, deg, _ = r.Buscar(ctx, q, ClaseProcedimiento, 3)
	if !tiene(deg, DegVectorError) || len(cs) == 0 || cs[0].ID != "presupuestos.registrar-metrado" {
		t.Errorf("vector caído: %v %v", deg, cs)
	}

	// Reranker: reordena (aquí a favor de «Modificar») y su puntaje pasa a [0, 1].
	r = NuevoRecuperador(b)
	r.Reranker = rerankFalso{favor: "Modificar una partida"}
	cs, deg, _ = r.Buscar(ctx, consulta(b, "partida"), ClaseProcedimiento, 3)
	if tiene(deg, DegSinReranker) || cs[0].ID != "presupuestos.modificar-partida" || cs[0].Rerank == nil || cs[0].Puntaje <= 0.9 {
		t.Errorf("reranker: %v %+v", deg, cs)
	}

	// Reranker caído: puntaje híbrido, informado.
	r = NuevoRecuperador(b)
	r.Reranker = rerankFalso{err: errors.New("timeout")}
	cs, deg, _ = r.Buscar(ctx, q, ClaseProcedimiento, 3)
	if !tiene(deg, DegRerankerError) || cs[0].Rerank != nil || cs[0].ID != "presupuestos.registrar-metrado" {
		t.Errorf("reranker caído: %v %+v", deg, cs)
	}

	// Fragmentos: solo léxica (no se embeben miles de fragmentos al vuelo).
	r = NuevoRecuperador(b)
	r.Vector = &vecFalso{}
	_, deg, _ = r.Buscar(ctx, consulta(b, "guia de remision"), ClaseFragmento, 3)
	if !tiene(deg, DegFragmentosLex) {
		t.Errorf("fragmentos: %v", deg)
	}

	// Sin índice de fragmentos: lo dice, no falla.
	bs := Cargar(Opciones{DirKB: "testdata/kb", SinFragmentos: true, ArchivoPlantillas: "testdata/no-existe.yml"})
	_, deg, err := NuevoRecuperador(bs).Buscar(ctx, q, ClaseFragmento, 3)
	if err != nil || !tiene(deg, DegSinFragmentos) {
		t.Errorf("sin índice: %v %v", deg, err)
	}
}

// Los procedimientos reales: sus propias preguntas los recuperan, y las consultas del megaprompt que
// tienen procedimiento real caen en él.
func TestReal_Recuperacion(t *testing.T) {
	b, r := recuperadorReal(t)
	ctx := context.Background()
	total, top1, top3 := 0, 0, 0
	for _, p := range b.Procedimientos {
		for _, q := range p.Preguntas {
			total++
			cs, _, _ := r.Buscar(ctx, consulta(b, q), ClaseProcedimiento, 3)
			for i, c := range cs {
				if c.ID == p.ID {
					if i == 0 {
						top1++
					}
					top3++
					break
				}
			}
		}
	}
	t.Logf("preguntas de los YAML: %d; acierto@1 %.1f%%, acierto@3 %.1f%%", total, 100*float64(top1)/float64(total), 100*float64(top3)/float64(total))
	if float64(top3) < 0.85*float64(total) {
		t.Errorf("acierto@3 bajo: %d de %d", top3, total)
	}
	esperados := []struct{ q, clase, id string }{
		{"como registro metrado", ClaseProcedimiento, "presupuestos.ingresar-metrados"},
		{"que es metrado", ClaseConcepto, "metrado"},
		{"como registro un presupuesto nuevo", ClaseProcedimiento, "presupuestos.registrar-presupuesto-nuevo"},
		{"como elaboro la formula polinomica", ClaseProcedimiento, "presupuestos.elaborar-formula-polinomica"},
	}
	for _, e := range esperados {
		if _, ok := b.Proc(e.id); !ok && e.clase == ClaseProcedimiento {
			continue
		}
		if _, ok := b.Conc(e.id); !ok && e.clase == ClaseConcepto {
			continue
		}
		if got := primero(t, r, b, e.q, e.clase); got.ID != e.id {
			t.Errorf("%q: %s, quiero %s", e.q, got.ID, e.id)
		}
	}
}

// Faltas de ortografía: lo que suena igual o está a una edición de una palabra frecuente se corrige; lo
// desconocido sin corrección se queda (y sigue pesando: «criptomonedas» no debe encontrar nada).
func TestVocabulario_CorrigeFaltas(t *testing.T) {
	v := nuevoVocabulario()
	for i := 0; i < 5; i++ {
		v.agregar("registrar presupuesto nuevo en el catálogo de presupuestos; cuadrilla de la partida")
	}
	v.cerrar()
	casos := map[string]string{"nuebo": "nuevo", "prosupuesto": "presupuesto", "kuadrilla": "cuadrilla", "criptomonedas": "criptomonedas", "partida": "partida"}
	for in, want := range casos {
		if got, vacia := v.corregir(in); got != want || vacia {
			t.Errorf("corregir(%q) = %q (vacía %v), quiero %q", in, got, vacia, want)
		}
	}
	for _, w := range []string{"kiero", "aser", "komo"} {
		if _, vacia := v.corregir(w); !vacia {
			t.Errorf("%q suena a una palabra vacía y debe quitarse", w)
		}
	}
	if got := v.terminosCorregidos("komo registro un prosupuesto nuebo"); strings.Join(got, " ") != "regis presu nuevo" {
		t.Errorf("términos corregidos: %v", got)
	}
}

// Dominio: bajo el umbral, un candidato que cubre lo mínimo y aventaja con claridad al mejor de OTRO
// elemento se acepta; un empate no.
func TestConstructor_Dominio(t *testing.T) {
	c := NuevoConstructor(cargarFixture(t), nil)
	mk := func(id, proc string, ef, bm25 float64) candEf {
		return candEf{tipos.Candidato{ID: id, Lexico: bm25, Meta: map[string]string{"procedimiento": proc}}, ef}
	}
	casos := []struct {
		nombre string
		cs     []candEf
		clave  func(candEf) string
		quiero bool
	}{
		{"sobre el umbral", []candEf{mk("a", "", 0.7, 5), mk("b", "", 0.69, 5)}, porID, true},
		{"domina", []candEf{mk("a", "", 0.35, 20), mk("b", "", 0.3, 10)}, porID, true},
		{"empata", []candEf{mk("a", "", 0.45, 20), mk("b", "", 0.4, 18)}, porID, false},
		{"cubre muy poco", []candEf{mk("a", "", 0.2, 30), mk("b", "", 0.1, 5)}, porID, false},
		{"errores del mismo procedimiento no compiten", []candEf{mk("p#1", "p", 0.35, 20), mk("p#2", "p", 0.35, 19), mk("q#1", "q", 0.2, 8)}, porProcedimiento, true},
		{"sin rival: hace falta medio camino", []candEf{mk("a", "", 0.4, 20)}, porID, false},
		{"sin rival y cubre bastante", []candEf{mk("a", "", 0.5, 20)}, porID, true},
	}
	for _, x := range casos {
		if got := c.acepta(x.cs, 0, 0.6, x.clave); got != x.quiero {
			t.Errorf("%s: acepta = %v", x.nombre, got)
		}
	}
}
