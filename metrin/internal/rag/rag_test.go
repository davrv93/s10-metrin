package rag

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rag-go/internal/almacen"
	"rag-go/internal/clasificar"
	"rag-go/internal/indexar"
	"rag-go/internal/llm"
)

// dosEjes: embebedor de prueba con dos dimensiones, «precio» y «formulario».
type dosEjes struct{}

func (dosEjes) Nombre() string { return "dos-ejes" }
func (dosEjes) Embeber(_ context.Context, t string) ([]float32, error) {
	if strings.Contains(t, "precio") {
		return []float32{1, 0}, nil
	}
	return []float32{0, 1}, nil
}

type llmFijo struct {
	respuesta string
	visto     []llm.Mensaje
}

func (l *llmFijo) Chat(_ context.Context, m []llm.Mensaje) (string, error) {
	l.visto = m
	return l.respuesta, nil
}

type llmCaido struct{}

func (llmCaido) Chat(context.Context, []llm.Mensaje) (string, error) {
	return "", errors.New("modelo no disponible")
}

func preparar(t *testing.T, respuesta string) (*RAG, *llmFijo, string) {
	ctx := context.Background()
	a, _ := almacen.Abrir("", dosEjes{})
	a.Reemplazar(ctx, "s3:docs/tarifas.md", almacen.FuenteS3, "e1", []almacen.Trozo{{
		ID: almacen.IDTrozo("s3:docs/tarifas.md", 0), Texto: "El plan Pro tiene precio de 49 €",
		Metadata: map[string]string{"source": almacen.FuenteS3, "cita": "s3:rag-demo/docs/tarifas.md#0"},
	}})
	l := &llmFijo{respuesta: respuesta}
	fallos := filepath.Join(t.TempDir(), "sin_respuesta.jsonl")
	return &RAG{Almacen: a, LLM: l, MaxDistancia: 0.5, RutaFallos: fallos}, l, fallos
}

func TestPreguntaConContexto(t *testing.T) {
	r, l, fallos := preparar(t, "Cuesta 49 € [s3:rag-demo/docs/tarifas.md#0]")
	res, err := r.Preguntar(context.Background(), "¿qué precio tiene el Pro?", Opciones{})
	if err != nil || res.SinContexto {
		t.Fatalf("%v %+v", err, res)
	}
	if !strings.Contains(l.visto[1].Content, "FUENTE: s3:rag-demo/docs/tarifas.md#0") {
		t.Error("el prompt no lleva la etiqueta de cita del trozo")
	}
	if _, err := os.Stat(fallos); !os.IsNotExist(err) {
		t.Error("no debía registrarse como fallo")
	}
}

func TestSinContextoPorDistancia(t *testing.T) {
	r, l, fallos := preparar(t, "no debería llamarse")
	res, _ := r.Preguntar(context.Background(), "¿cómo es el formulario de contacto que tiene validaciones?", Opciones{})
	if !res.SinContexto || l.visto != nil {
		t.Fatalf("distancia 1 > 0.5: no debe llamar al LLM; %+v", res)
	}
	b, _ := os.ReadFile(fallos)
	if !strings.Contains(string(b), "¿cómo es el formulario") {
		t.Fatalf("no se registró la pregunta: %q", b)
	}
}

func TestSinContextoCortoConversa(t *testing.T) {
	r, _, fallos := preparar(t, "")
	sec := &llmSecuencia{respuestas: []string{"Bien, ¿y tú?"}}
	r.LLM = sec
	res, err := r.Preguntar(context.Background(), "bien y tú", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Modo != "conversacional" || res.Respuesta != "Bien, ¿y tú?" {
		t.Fatalf("corto sin contexto debió conversar: %+v", res)
	}
	b, _ := os.ReadFile(fallos)
	if !strings.Contains(string(b), "bien y tú") {
		t.Fatalf("igual debió registrarse el fallo: %q", b)
	}
}

func TestSinContextoPorElModelo(t *testing.T) {
	r, _, fallos := preparar(t, "No tengo contexto suficiente para responder a esa pregunta.")
	res, _ := r.Preguntar(context.Background(), "¿el precio incluye IVA en Chile?", Opciones{})
	if !res.SinContexto {
		t.Fatal("la negativa del modelo debe marcar sin contexto")
	}
	b, _ := os.ReadFile(fallos)
	if strings.Count(string(b), "\n") != 1 {
		t.Fatalf("esperaba una línea en %s: %q", fallos, b)
	}
}

type llmSecuencia struct {
	respuestas []string
	llamadas   int
}

func (l *llmSecuencia) Chat(_ context.Context, _ []llm.Mensaje) (string, error) {
	r := l.respuestas[l.llamadas]
	l.llamadas++
	return r, nil
}

func TestPortuguesReintentaUnaVez(t *testing.T) {
	r, _, _ := preparar(t, "")
	r.LLM = &llmSecuencia{respuestas: []string{
		"Também permite uma estrutura",
		"Permite una estructura [s3:rag-demo/docs/tarifas.md#0]",
	}}
	res, err := r.Preguntar(context.Background(), "¿qué precio tiene el Pro?", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Modo != "respuesta" || parecePortugues(res.Respuesta) {
		t.Fatalf("reintento debió rescatar español: %+v", res)
	}
	if r.LLM.(*llmSecuencia).llamadas != 2 {
		t.Fatalf("esperaba 2 llamadas, hubo %d", r.LLM.(*llmSecuencia).llamadas)
	}
}

func TestPortuguesPersistenteCaenFragmentos(t *testing.T) {
	r, _, _ := preparar(t, "")
	r.LLM = &llmSecuencia{respuestas: []string{"Também uma coisa", "Não há como"}}
	res, err := r.Preguntar(context.Background(), "¿qué precio tiene el Pro?", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Modo != "modelo_no_disponible" || parecePortugues(res.Respuesta) {
		t.Fatalf("PT persistente debió ocultar la salida del modelo: %+v", res)
	}
	if strings.Contains(res.Respuesta, "También una cosa") {
		t.Fatal("no debe mostrar una salida no fiable")
	}
}

func TestSeleccionaSoloContextoCercanoYEliminaCitasDuplicadas(t *testing.T) {
	trozos := []almacen.Resultado{
		{ID: "primero", Distancia: 0.10, Texto: "Mejor coincidencia", Metadata: map[string]string{"cita": "Guía, p. 10"}},
		{ID: "duplicado", Distancia: 0.12, Texto: "Mismo origen", Metadata: map[string]string{"cita": "Guía, p. 10"}},
		{ID: "cercano", Distancia: 0.24, Texto: "Contexto complementario", Metadata: map[string]string{"cita": "Manual, p. 3"}},
		{ID: "lejano", Distancia: 0.35, Texto: "Tema distinto", Metadata: map[string]string{"cita": "Video"}},
	}
	seleccionados := seleccionarContexto(trozos, 0.8)
	if len(seleccionados) != 2 || seleccionados[0].ID != "primero" || seleccionados[1].ID != "cercano" {
		t.Fatalf("contexto seleccionado inesperado: %+v", seleccionados)
	}
}

func TestFalloLLMNoExponeFragmentosCrudos(t *testing.T) {
	r, _, _ := preparar(t, "")
	r.LLM = llmCaido{}
	res, err := r.Preguntar(context.Background(), "¿qué precio tiene el Pro?", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Modo != "modelo_no_disponible" || len(res.Fuentes) == 0 {
		t.Fatalf("respuesta de error inesperada: %+v", res)
	}
	if strings.Contains(res.Respuesta, "El plan Pro tiene precio") {
		t.Fatalf("se filtró el fragmento en bruto: %q", res.Respuesta)
	}
}

func TestDoblePaseCortexEtiquetaSecciones(t *testing.T) {
	ctx := context.Background()
	a, _ := almacen.Abrir("", dosEjes{})
	a.Reemplazar(ctx, "s10kb:c1", indexar.FuenteS10KB, "e1", []almacen.Trozo{{
		ID: almacen.IDTrozo("s10kb:c1", 0), Texto: "El presupuesto puede ser Venta, Meta o Línea Base",
		Metadata: map[string]string{"source": indexar.FuenteS10KB, "manual": "Cortex", "cita": "Cortex: presupuestos"},
	}})
	l := &llmFijo{respuesta: "Hay tres tipos [Cortex: presupuestos]"}
	r := &RAG{Almacen: a, LLM: l, MaxDistancia: 0.9, RutaFallos: filepath.Join(t.TempDir(), "f.jsonl")}
	res, err := r.Preguntar(ctx, "¿qué tipos de presupuesto hay?", Opciones{})
	if err != nil || res.SinContexto {
		t.Fatalf("%v %+v", err, res)
	}
	prompt := l.visto[1].Content
	if !strings.Contains(prompt, "El presupuesto puede ser Venta, Meta o Línea Base") {
		t.Fatal("prompt sin el contexto recuperado")
	}
	if strings.Contains(prompt, "Elabora en dos partes") {
		t.Fatal("el prompt conserva una instrucción de dos secciones que añade ruido")
	}
}

func TestConsultaTiposPresupuestoPriorizaClasificacionCortex(t *testing.T) {
	curado := almacen.Resultado{
		ID: "tipos", Distancia: 0.72,
		Texto:    "El manual nombra los tipos Venta, Meta y Línea Base.",
		Metadata: map[string]string{"manual": ManualCortex, "cita": "Cortex: tipos de presupuesto"},
	}
	ruido := almacen.Resultado{
		ID: "ruido", Distancia: 0.12,
		Texto:    "Para buscar un presupuesto, ubíquelo en el árbol y haga doble clic.",
		Metadata: map[string]string{"cita": "Guía, p. 67"},
	}
	for _, pregunta := range []string{
		"¿Qué tipos de presupuestos existen?",
		"¿Qué presupuestos existen?",
		"¿Qué presupuestos hay?",
		"tipos de presupuesto",
	} {
		if !esConsultaTiposPresupuesto(pregunta) {
			t.Errorf("no reconoció la consulta de clasificación: %q", pregunta)
		}
	}
	if esConsultaTiposPresupuesto("presupuestos") {
		t.Fatal("no debe asumir una intención de clasificación en una consulta ambigua")
	}
	seleccionados := seleccionarContexto([]almacen.Resultado{ruido, curado}, 0.8)
	if elegido := contextoTiposPresupuesto([]almacen.Resultado{curado}); elegido == nil || elegido.ID != "tipos" {
		t.Fatalf("no priorizó la clasificación curada; contexto vectorial: %+v", seleccionados)
	}
}

func TestHiloReescribeYConservaHistorial(t *testing.T) {
	r, _, _ := preparar(t, "")
	r.MaxDistancia = 2
	sec := &llmSecuencia{respuestas: []string{
		"¿Quién es el CEO de Optimiza 360?",
		"Es Arturo Dupont [s3:rag-demo/docs/tarifas.md#0]",
	}}
	r.LLM = sec
	hilo := []Turno{
		{Rol: "usuario", Texto: "¿qué es optimiza 360?"},
		{Rol: "asistente", Texto: "Es una consultora."},
	}
	res, err := r.Preguntar(context.Background(), "¿quién es su ceo?", Opciones{Hilo: hilo})
	if err != nil {
		t.Fatal(err)
	}
	if sec.llamadas != 2 {
		t.Fatalf("reescritura + respuesta = 2 llamadas, hubo %d", sec.llamadas)
	}
	if res.PreguntaReescrita != "¿Quién es el CEO de Optimiza 360?" {
		t.Fatalf("no reescribió: %q", res.PreguntaReescrita)
	}
}

func TestSinHiloUnaSolaLlamada(t *testing.T) {
	r, _, _ := preparar(t, "")
	sec := &llmSecuencia{respuestas: []string{"Cuesta 49 € [s3:rag-demo/docs/tarifas.md#0]"}}
	r.LLM = sec
	if _, err := r.Preguntar(context.Background(), "¿qué precio tiene el Pro?", Opciones{}); err != nil {
		t.Fatal(err)
	}
	if sec.llamadas != 1 {
		t.Fatalf("sin hilo no se reescribe: %d llamadas", sec.llamadas)
	}
}

func armarConRuta(t *testing.T, respuesta string) (*RAG, *llmSecuencia) {
	ctx := context.Background()
	m, err := clasificar.Entrenar(ctx, dosEjes{}, "dos-ejes", 0.0, map[string][]string{
		"social":  {"hola", "gracias"},
		"trabajo": {"precio del plan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	r, _, _ := preparar(t, "")
	r.Emb = dosEjes{}
	r.Clasificador = m
	sec := &llmSecuencia{respuestas: []string{respuesta}}
	r.LLM = sec
	return r, sec
}

func TestSaludoVaConversacional(t *testing.T) {
	r, sec := armarConRuta(t, "Hola, ¿en qué te ayudo?")
	res, err := r.Preguntar(context.Background(), "hola", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Modo != "conversacional" || len(res.Fuentes) != 0 {
		t.Fatalf("saludo debió ir directo: %+v", res)
	}
	if sec.llamadas != 1 || res.Respuesta != "Hola, ¿en qué te ayudo?" {
		t.Fatalf("llamadas=%d respuesta=%q", sec.llamadas, res.Respuesta)
	}
}

func TestTrabajoSigueAlRAG(t *testing.T) {
	r, sec := armarConRuta(t, "Cuesta 49 € [s3:rag-demo/docs/tarifas.md#0]")
	res, err := r.Preguntar(context.Background(), "¿qué precio tiene el Pro?", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Modo != "respuesta" {
		t.Fatalf("trabajo debió ir al RAG: %+v", res)
	}
	if sec.llamadas != 1 {
		t.Fatalf("sin hilo: 1 llamada, hubo %d", sec.llamadas)
	}
}

func TestParecePortugues(t *testing.T) {
	pt := []string{"Também permite uma estrutura", "não há informação", "dos documentos", "à medida"}
	es := []string{"El plan Pro cuesta 49 €", "Nos ayuda a resolver dudas", "Como se indica arriba", "Para continuar, dime más"}
	for _, s := range pt {
		if !parecePortugues(s) {
			t.Errorf("debió detectar PT: %q", s)
		}
	}
	for _, s := range es {
		if parecePortugues(s) {
			t.Errorf("falso positivo ES: %q", s)
		}
	}
}
