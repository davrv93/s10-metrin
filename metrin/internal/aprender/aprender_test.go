package aprender

import (
	"context"
	"hash/fnv"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"rag-go/internal/almacen"
	"rag-go/internal/indexar"
	"rag-go/internal/llm"
	"rag-go/internal/rag"
)

// bolsa: embebedor determinista (bolsa de palabras hasheada).
type bolsa struct{}

func (bolsa) Nombre() string { return "bolsa" }
func (bolsa) Embeber(_ context.Context, t string) ([]float32, error) {
	v := make([]float32, 64)
	for _, w := range strings.Fields(strings.ToLower(t)) {
		h := fnv.New32a()
		h.Write([]byte(strings.Trim(w, ".,¿?¡!:;")))
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

type llmEco struct{}

func (llmEco) Chat(_ context.Context, m []llm.Mensaje) (string, error) {
	return "Según la fuente, se hace desde el menú Almacén.", nil
}

func preparar(t *testing.T) (*Diario, *rag.RAG, string) {
	dir := t.TempDir()
	d, err := Abrir(dir)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := almacen.Abrir("", bolsa{})
	return d, &rag.RAG{Almacen: a, LLM: llmEco{}, MaxDistancia: 0.3}, dir
}

func preguntar(t *testing.T, d *Diario, r *rag.RAG, q string) (rag.Respuesta, string) {
	res, err := r.Preguntar(context.Background(), q, rag.Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	id, err := d.Registrar(res)
	if err != nil {
		t.Fatal(err)
	}
	return res, id
}

func TestCicloCompleto(t *testing.T) {
	d, r, dir := preparar(t)
	ctx := context.Background()
	eventos, soltar := d.Bus.Suscribir()
	defer soltar()

	// 1. Sin contexto: se encola (y dos veces la misma suma «veces»).
	res, _ := preguntar(t, d, r, "¿Cómo anulo una guía de remisión?")
	if !res.SinContexto {
		t.Fatalf("índice vacío: esperaba sin contexto %+v", res)
	}
	preguntar(t, d, r, "como anulo una guia de remision")
	pend := d.Pendientes(Pendiente)
	if len(pend) != 1 || pend[0].Veces != 2 {
		t.Fatalf("pendientes=%+v", pend)
	}
	if ev := <-eventos; ev.Tipo != "interaccion" {
		t.Fatalf("primer evento %q", ev.Tipo)
	}

	// 2. El admin enseña: se indexa en vivo y la siguiente vez responde.
	if _, err := d.Aprender(ctx, r.Almacen, Leccion{Clave: pend[0].Clave,
		Respuesta: "En Almacén, clic derecho sobre la guía y elegir Anular."}); err != nil {
		t.Fatal(err)
	}
	res, id := preguntar(t, d, r, "¿Cómo anulo una guía de remisión?")
	if res.SinContexto || res.Fuentes[0].Cita != TituloAprendido {
		t.Fatalf("tras aprender debía responder con la lección: %+v", res)
	}
	if got := d.Pendientes(Aprendido); len(got) != 1 {
		t.Fatalf("aprendidos=%+v", got)
	}
	b, _ := os.ReadFile(d.RutaAprendidos())
	doc, err := indexar.DocumentoKB([]byte(strings.TrimSpace(string(b))))
	if err != nil {
		t.Fatal(err)
	}
	// El index-kb del próximo arranque debe verlo igual (no reindexa).
	if e, _ := r.Almacen.Estado(doc.Clave); e.Version != doc.Version {
		t.Fatal("la versión indexada no coincide con la del aprendidos.jsonl")
	}

	// 3. 👎 a una respuesta de trabajo la reabre en la cola con el comentario.
	if err := d.Votar(id, -1, "faltó el paso de SUNAT"); err != nil {
		t.Fatal(err)
	}
	d.Votar(id, -1, "faltó el paso de SUNAT y el asiento") // el comentario llega aparte
	for _, p := range d.Pendientes(Pendiente) {
		if p.Clave == ClaveDe("¿Cómo anulo una guía de remisión?") {
			t.Fatalf("un 👎 sobre lo aprendido no reabre la lección: %+v", p)
		}
	}
	res, id2 := preguntar(t, d, r, "¿Y la nota de crédito?")
	d.Votar(id2, -1, "")
	d.Votar(id2, -1, "no sirvió")
	for _, p := range d.Pendientes(Pendiente) {
		if p.Clave == ClaveDe("¿Y la nota de crédito?") && (p.Veces != 1 || p.Respuesta != "" || p.Comentario != "no sirvió" || p.Origen != OrigenSinContexto) {
			t.Fatalf("votos repetidos no suman veces ni guardan la negativa como respuesta: %+v", p)
		}
	}
	_ = res
	if err := d.Votar("no-existe", 1, ""); err != ErrNoEncontrado {
		t.Fatalf("voto a id desconocido: %v", err)
	}

	// 4. Repaso: una pendiente que ya tiene fuente queda resuelta.
	preguntar(t, d, r, "¿Qué es el kardex valorizado?")
	r.Almacen.Reemplazar(ctx, "s10kb:x", indexar.FuenteS10KB, "1", []almacen.Trozo{{
		ID: almacen.IDTrozo("s10kb:x", 0), Texto: "¿Qué es el kardex valorizado? Es el reporte de movimientos con costo.",
		Metadata: map[string]string{"source": indexar.FuenteS10KB, "cita": "Manual Almacén"},
	}})
	rep, err := d.Repasar(ctx, r, "admin")
	if err != nil || rep.Resueltas != 1 {
		t.Fatalf("repaso %+v %v", rep, err)
	}

	m := d.Metricas(0)
	if m.Total != 5 || m.SinContexto != 4 || m.Respondidas != 1 || m.Negativos != 2 || m.Cola.Resueltas != 1 || m.Cola.Aprendidas != 1 {
		t.Fatalf("métricas %+v", m)
	}
	if m.TasaRespuesta != 0.2 || m.Satisfaccion != 0 {
		t.Fatalf("tasas %+v", m)
	}

	// 5. Todo sobrevive a un reinicio.
	d2, err := Abrir(dir)
	if err != nil {
		t.Fatal(err)
	}
	m2 := d2.Metricas(0)
	if m2.Total != m.Total || m2.Negativos != 2 || m2.Cola != m.Cola || m2.Repaso == nil {
		t.Fatalf("tras reabrir %+v", m2)
	}
}

func TestProxima(t *testing.T) {
	base := time.Date(2026, 9, 30, 1, 0, 0, 0, time.UTC)
	if p := proxima(base, 2, 0); !p.Equal(base.Add(time.Hour)) {
		t.Fatalf("hoy a las 02:00: %v", p)
	}
	if p := proxima(base, 0, 30); p.Day() != 1 || p.Month() != 10 {
		t.Fatalf("mañana a las 00:30: %v", p)
	}
}
