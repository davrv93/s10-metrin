package busqueda

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// vectorFijo devuelve siempre la misma lista (o un error) y cuenta llamadas.
type vectorFijo struct {
	lista   []Candidato
	err     error
	espera  time.Duration
	llamado atomic.Int32
	ultimaQ atomic.Value
}

func (v *vectorFijo) BuscarVector(ctx context.Context, q string, k int) ([]Candidato, error) {
	v.llamado.Add(1)
	v.ultimaQ.Store(q)
	if v.espera > 0 {
		select {
		case <-time.After(v.espera):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if v.err != nil {
		return nil, v.err
	}
	return v.lista[:min(k, len(v.lista))], nil
}

// rerankFijo puntúa por una tabla texto → puntaje.
type rerankFijo struct {
	puntaje func(doc string) float64
	err     error
	espera  time.Duration
}

func (r rerankFijo) Rerank(ctx context.Context, _ string, docs []string) ([]float64, error) {
	if r.espera > 0 {
		select {
		case <-time.After(r.espera):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if r.err != nil {
		return nil, r.err
	}
	out := make([]float64, len(docs))
	for i, d := range docs {
		out[i] = r.puntaje(d)
	}
	return out, nil
}

func corpusHibrido(t *testing.T) *Indice {
	return indicePrueba(t, ConfigPorDefecto(),
		Doc{ID: "metrado", Campos: map[string]string{"titulo": "Metrados del proyecto", "texto": "El metrado es la cantidad de cada partida."}},
		Doc{ID: "partida", Campos: map[string]string{"titulo": "Adicionar partida", "texto": "Clic derecho, Adicionar Partida."}},
		Doc{ID: "ceo", Campos: map[string]string{"titulo": "Fundador", "texto": "Biografía del fundador de la empresa."}},
		Doc{ID: "imagen", Campos: map[string]string{"ocr": "Metrado Unidad Precio Parcial"}},
	)
}

func idsR(rs []Resultado) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.ID
	}
	return out
}

func TestRRFFusionaYMarcaFuentes(t *testing.T) {
	lex := corpusHibrido(t)
	vec := &vectorFijo{lista: []Candidato{{ID: "ceo", Similitud: 0.43}, {ID: "metrado", Similitud: 0.42}, {ID: "imagen", Similitud: 0.40}}}
	b := NuevoBuscador(lex, vec, nil, nil, Opciones{RRFK: 60})
	inf, err := b.BuscarInforme(context.Background(), "¿Qué es un metrado?", 10)
	if err != nil {
		t.Fatal(err)
	}
	lexOrden := ids(lex.BuscarLexico("qué es un metrado", 10))
	rango := func(id string, l []string) int {
		for i, x := range l {
			if x == id {
				return i + 1
			}
		}
		return 0
	}
	vecOrden := []string{"ceo", "metrado", "imagen"}
	for _, r := range inf.Resultados {
		quiere := 0.0
		if rv := rango(r.ID, vecOrden); rv > 0 {
			quiere += 1 / (60 + float64(rv))
		}
		if rl := rango(r.ID, lexOrden); rl > 0 {
			quiere += 1 / (60 + float64(rl))
		}
		if math.Abs(r.RRF-quiere) > 1e-12 || r.Score != r.RRF {
			t.Errorf("%s: RRF %.6f, quiere %.6f", r.ID, r.RRF, quiere)
		}
		if r.RangoVector != rango(r.ID, vecOrden) || r.RangoBM25 != rango(r.ID, lexOrden) {
			t.Errorf("%s: rangos %d/%d", r.ID, r.RangoVector, r.RangoBM25)
		}
	}
	// «metrado» sale en las dos listas: debe ganar al CEO, que solo sale por
	// vector aunque allí sea primero.
	if inf.Resultados[0].ID != "metrado" {
		t.Errorf("primero = %s, quiere metrado (%v)", inf.Resultados[0].ID, idsR(inf.Resultados))
	}
	if !reflect.DeepEqual(inf.Resultados[0].Fuentes, []string{"vector", "bm25"}) {
		t.Errorf("fuentes = %v", inf.Resultados[0].Fuentes)
	}
	if got := vec.ultimaQ.Load(); got != "qué es un metrado" {
		t.Errorf("el vector recibió %q, quiere la consulta normalizada", got)
	}
}

func TestModosSimples(t *testing.T) {
	lex := corpusHibrido(t)
	vec := &vectorFijo{lista: []Candidato{{ID: "ceo", Similitud: 0.43}, {ID: "metrado", Similitud: 0.42}}}
	bv := NuevoBuscador(lex, vec, nil, nil, Opciones{Modo: ModoVector})
	rs, _ := bv.Buscar(context.Background(), "metrado", 5)
	if !reflect.DeepEqual(idsR(rs), []string{"ceo", "metrado"}) || rs[0].Score != 0.43 {
		t.Errorf("modo vector: %v", rs)
	}
	bl := NuevoBuscador(lex, vec, nil, nil, Opciones{Modo: ModoLexico})
	rs, _ = bl.Buscar(context.Background(), "metrado", 5)
	if vec.llamado.Load() != 1 {
		t.Error("el modo léxico no debía consultar el vector")
	}
	if len(rs) == 0 || rs[0].ID != "metrado" || rs[0].Score != rs[0].BM25 {
		t.Errorf("modo léxico: %v", rs)
	}
}

func TestDegradaSiFallaElVector(t *testing.T) {
	lex := corpusHibrido(t)
	vec := &vectorFijo{err: errors.New("chromem caído")}
	b := NuevoBuscador(lex, vec, nil, nil, Opciones{TamCache: 8})
	inf, err := b.BuscarInforme(context.Background(), "metrado de partida", 5)
	if err != nil {
		t.Fatalf("no debía fallar: %v", err)
	}
	if !reflect.DeepEqual(inf.Degradado, []string{"sin_vector"}) || inf.FalloVector == "" {
		t.Errorf("degradado = %v / %q", inf.Degradado, inf.FalloVector)
	}
	if !reflect.DeepEqual(idsR(inf.Resultados), ids(lex.BuscarLexico("metrado de partida", 5))) {
		t.Errorf("con el vector caído el orden debía ser el léxico: %v", idsR(inf.Resultados))
	}
	// Lo degradado no se guarda en caché.
	if b.Cache().Largo() != 0 {
		t.Error("un resultado degradado no debía quedar en caché")
	}
	// Sin índice léxico y con el vector caído sí es error.
	b2 := NuevoBuscador(nil, vec, nil, nil, Opciones{})
	if _, err := b2.Buscar(context.Background(), "metrado", 5); err == nil {
		t.Error("sin ninguna lista debía ser error")
	}
}

func TestDegradaSiFallaElReranker(t *testing.T) {
	lex := corpusHibrido(t)
	vec := &vectorFijo{lista: []Candidato{{ID: "ceo", Similitud: 0.43}, {ID: "metrado", Similitud: 0.42}, {ID: "imagen", Similitud: 0.41}}}
	base := NuevoBuscador(lex, vec, nil, nil, Opciones{})
	quiere, _ := base.Buscar(context.Background(), "metrado", 5)

	for nombre, rr := range map[string]Reordenador{
		"error":   rerankFijo{err: errors.New("llama-server caído")},
		"timeout": rerankFijo{espera: time.Second, puntaje: func(string) float64 { return 1 }},
		"corto":   cortoReranker{},
	} {
		b := NuevoBuscador(lex, vec, rr, nil, Opciones{TimeoutRerank: 50 * time.Millisecond})
		t0 := time.Now()
		inf, err := b.BuscarInforme(context.Background(), "metrado", 5)
		if err != nil {
			t.Fatalf("%s: no debía fallar: %v", nombre, err)
		}
		if nombre == "timeout" && time.Since(t0) > 500*time.Millisecond {
			t.Errorf("timeout: tardó %v", time.Since(t0))
		}
		if !reflect.DeepEqual(inf.Degradado, []string{"sin_rerank"}) {
			t.Errorf("%s: degradado = %v", nombre, inf.Degradado)
		}
		if !reflect.DeepEqual(inf.Resultados, quiere) {
			t.Errorf("%s: debía quedar el puntaje híbrido\n got %v\nwant %v", nombre, inf.Resultados, quiere)
		}
		for _, r := range inf.Resultados {
			if r.Reordenado || r.Score != r.RRF {
				t.Errorf("%s: %s marcado como reordenado o sin puntaje RRF", nombre, r.ID)
			}
		}
	}
}

type cortoReranker struct{}

func (cortoReranker) Rerank(context.Context, string, []string) ([]float64, error) {
	return []float64{1}, nil
}

func TestRerankerReordenaElTop(t *testing.T) {
	lex := corpusHibrido(t)
	vec := &vectorFijo{lista: []Candidato{{ID: "ceo", Similitud: 0.43}, {ID: "metrado", Similitud: 0.42}, {ID: "imagen", Similitud: 0.41}}}
	// El reranker prefiere la imagen (texto con «Unidad»).
	rr := rerankFijo{puntaje: func(d string) float64 {
		switch {
		case strings.Contains(d, "Unidad"):
			return 5
		case strings.Contains(d, "cantidad"):
			return 3
		}
		return -1
	}}
	b := NuevoBuscador(lex, vec, rr, nil, Opciones{TopRerank: 30})
	rs, err := b.Buscar(context.Background(), "metrado", 3)
	if err != nil {
		t.Fatal(err)
	}
	if rs[0].ID != "imagen" || rs[1].ID != "metrado" || !rs[0].Reordenado || rs[0].Score != 5 {
		t.Errorf("orden tras rerank: %v", rs)
	}
}

func TestCacheDaLoMismoQueSinCache(t *testing.T) {
	lex := corpusHibrido(t)
	vec := &vectorFijo{lista: []Candidato{{ID: "ceo", Similitud: 0.43}, {ID: "metrado", Similitud: 0.42}, {ID: "imagen", Similitud: 0.41}}}
	al := NuevoAlias()
	al.AgregarAcciones("registrar", "ingresar", "cargar")
	sin := NuevoBuscador(lex, vec, nil, al, Opciones{Expandir: true, TamCache: 0})
	con := NuevoBuscador(lex, vec, nil, al, Opciones{Expandir: true, TamCache: 4})
	ctx := context.Background()
	for _, q := range []string{"¿Qué es un metrado?", "qué es un metrado", "como cargo una partida", "fundador"} {
		a, err := sin.BuscarInforme(ctx, q, 5)
		if err != nil {
			t.Fatal(err)
		}
		b1, _ := con.BuscarInforme(ctx, q, 5)
		b2, _ := con.BuscarInforme(ctx, q, 5)
		if !b2.DeCache {
			t.Errorf("%q: la segunda consulta debía salir de la caché", q)
		}
		for _, b := range []Informe{b1, b2} {
			if !reflect.DeepEqual(a.Resultados, b.Resultados) {
				t.Errorf("%q: con caché ≠ sin caché\n%v\n%v", q, a.Resultados, b.Resultados)
			}
			if b.Consulta.Original != q || !reflect.DeepEqual(a.Consulta.Aliases, b.Consulta.Aliases) {
				t.Errorf("%q: consulta distinta: %+v", q, b.Consulta)
			}
		}
	}
	// «¿Qué es un metrado?» y «qué es un metrado» normalizan igual: la
	// segunda forma ya estaba en caché.
	if con.Cache().Aciertos < 5 {
		t.Errorf("aciertos = %d", con.Cache().Aciertos)
	}
	// Modificar lo devuelto no altera la caché.
	r1, _ := con.Buscar(ctx, "fundador", 5)
	r1[0].ID = "alterado"
	r1[0].Fuentes[0] = "x"
	r2, _ := con.Buscar(ctx, "fundador", 5)
	if r2[0].ID == "alterado" || r2[0].Fuentes[0] == "x" {
		t.Error("la caché devolvió un resultado compartido")
	}
}

// vectorQueEsperaAlLexico solo contesta cuando la búsqueda léxica ya
// terminó: si el Buscador las corriera una tras otra (vector primero), se
// quedaría esperando hasta el plazo y fallaría.
type vectorQueEsperaAlLexico struct{ listo chan struct{} }

func (v vectorQueEsperaAlLexico) BuscarVector(ctx context.Context, q string, k int) ([]Candidato, error) {
	select {
	case <-v.listo:
		return []Candidato{{ID: "ceo", Similitud: 0.4}}, nil
	case <-time.After(time.Second):
		return nil, errors.New("el léxico no corrió en paralelo")
	}
}

func TestBusquedasEnParalelo(t *testing.T) {
	lex := corpusHibrido(t)
	v := vectorQueEsperaAlLexico{listo: make(chan struct{})}
	b := NuevoBuscador(lex, v, nil, nil, Opciones{})
	var una sync.Once
	b.trasLexico = func() { una.Do(func() { close(v.listo) }) }
	inf, err := b.BuscarInforme(context.Background(), "metrado", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(inf.Degradado) != 0 {
		t.Fatalf("las búsquedas no corrieron en paralelo: %v %s", inf.Degradado, inf.FalloVector)
	}
	// Uso concurrente del mismo Buscador (correr con go test -race).
	vec := &vectorFijo{lista: []Candidato{{ID: "ceo", Similitud: 0.4}, {ID: "metrado", Similitud: 0.3}}}
	bc := NuevoBuscador(lex, vec, rerankFijo{puntaje: func(string) float64 { return 1 }}, NuevoAlias(), Opciones{TamCache: 2})
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q := []string{"metrado", "partida", "fundador"}[i%3]
			if _, err := bc.Buscar(context.Background(), q, 3); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
}

type embebedorContado struct {
	n atomic.Int32
}

func (e *embebedorContado) Nombre() string { return "contado" }
func (e *embebedorContado) Embeber(_ context.Context, t string) ([]float32, error) {
	e.n.Add(1)
	if t == "" {
		return nil, errors.New("vacío")
	}
	return []float32{float32(len(t)), 1}, nil
}

func TestEmbebedorConCache(t *testing.T) {
	base := &embebedorContado{}
	e := NuevoEmbebedorConCache(base, 2)
	ctx := context.Background()
	v1, _ := e.Embeber(ctx, "metrado")
	v1[0] = 999 // no debe contaminar la caché
	v2, _ := e.Embeber(ctx, "metrado")
	if base.n.Load() != 1 || v2[0] != 7 {
		t.Errorf("llamadas = %d, v2 = %v", base.n.Load(), v2)
	}
	e.Embeber(ctx, "a")
	e.Embeber(ctx, "b") // expulsa «metrado»
	e.Embeber(ctx, "metrado")
	if base.n.Load() != 4 {
		t.Errorf("tras expulsar: llamadas = %d, quiere 4", base.n.Load())
	}
	if _, err := e.Embeber(ctx, ""); err == nil {
		t.Error("el error debía propagarse")
	}
	if _, err := e.Embeber(ctx, ""); err == nil || base.n.Load() != 6 {
		t.Error("los errores no se guardan en caché")
	}
	if e.Nombre() != "contado" {
		t.Error("Nombre debe ser el del embebedor envuelto (misma colección)")
	}
}

// Sin reranker (o con el reranker caído) el orden final usa
// PesoVectorSinRerank; con reranker, el pool sale de PesoVector.
func TestPesoVectorSinRerank(t *testing.T) {
	lex := corpusHibrido(t)
	// «ceo» solo sale por vector (primero); «partida» solo por BM25 (primero).
	vec := &vectorFijo{lista: []Candidato{{ID: "ceo", Similitud: 0.43}, {ID: "metrado", Similitud: 0.41}}}
	ctx := context.Background()
	iguales, _ := NuevoBuscador(lex, vec, nil, nil, Opciones{PesoVector: 1, PesoLexico: 1}).Buscar(ctx, "partida", 5)
	bajo, _ := NuevoBuscador(lex, vec, nil, nil, Opciones{PesoVector: 0.1, PesoLexico: 1}).Buscar(ctx, "partida", 5)
	if reflect.DeepEqual(idsR(iguales), idsR(bajo)) {
		t.Fatalf("la prueba necesita órdenes distintos: %v", idsR(iguales))
	}
	if idsR(bajo)[2] != "ceo" || idsR(iguales)[1] != "ceo" {
		t.Fatalf("órdenes inesperados: iguales %v, bajo %v", idsR(iguales), idsR(bajo))
	}
	for nombre, rr := range map[string]Reordenador{"sin reranker": nil, "reranker caído": rerankFijo{err: errors.New("caído")}} {
		b := NuevoBuscador(lex, vec, rr, nil, Opciones{PesoVector: 1, PesoLexico: 1, PesoVectorSinRerank: 0.1})
		got, err := b.Buscar(ctx, "partida", 5)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, bajo) {
			t.Errorf("%s: el orden debía ser el de peso 0,1\n got %v\nwant %v", nombre, idsR(got), idsR(bajo))
		}
	}
	// Con reranker que funciona, el pool es el de pesos iguales: «ceo» entra
	// en el top 2 que ve el reranker.
	var visto []string
	rr := rerankFijo{puntaje: func(d string) float64 { visto = append(visto, d); return 0 }}
	b := NuevoBuscador(lex, vec, rr, nil, Opciones{PesoVector: 1, PesoLexico: 1, PesoVectorSinRerank: 0.1, TopRerank: 2})
	if _, err := b.Buscar(ctx, "partida", 5); err != nil {
		t.Fatal(err)
	}
	if len(visto) != 2 || !strings.Contains(visto[1], "fundador") {
		t.Errorf("el reranker debía ver el pool de pesos iguales: %q", visto)
	}
}
