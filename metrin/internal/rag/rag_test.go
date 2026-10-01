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

func TestRecuperaManualOficialAunqueNoEsteEnTopGeneral(t *testing.T) {
	r, _, _ := preparar(t, "El manual confirma el dato.")
	ctx := context.Background()
	if err := r.Almacen.Reemplazar(ctx, "s10kb:manual-seccion", indexar.FuenteS10KB, "v1", []almacen.Trozo{{
		ID: almacen.IDTrozo("s10kb:manual-seccion", 0), Texto: "El precio del plan está descrito en esta sección oficial.",
		Metadata: map[string]string{
			"source": indexar.FuenteS10KB, "document_id": "web-manual-presupuestos",
			"title": "Manual de Presupuestos › Registro", "section": "Registro",
			"source_url": "https://documentacion.s10peru.com/manual-de-presupuestos/",
			"confianza":  "oficial", "cita": "Manual de Presupuestos › Registro",
		},
	}}); err != nil {
		t.Fatal(err)
	}
	res, err := r.Preguntar(ctx, "¿qué precio tiene el Pro?", Opciones{K: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Fuentes) == 0 || !strings.Contains(res.Fuentes[0].Cita, "Manual de Presupuestos") {
		t.Fatalf("el pase oficial debe rescatar la sección manual incluso con K=1: %+v", res.Fuentes)
	}
}

func TestSinContextoPorDistancia(t *testing.T) {
	r, l, fallos := preparar(t, "no debería llamarse")
	res, _ := r.Preguntar(context.Background(), "¿cómo es el formulario de contacto que tiene validaciones?", Opciones{})
	if !res.SinContexto || l.visto != nil {
		t.Fatalf("distancia 1 > 0.5: no debe llamar al LLM; %+v", res)
	}
	if len(res.Fuentes) != 0 {
		t.Fatalf("una respuesta sin contexto no debe citar resultados irrelevantes: %+v", res.Fuentes)
	}
	b, _ := os.ReadFile(fallos)
	if !strings.Contains(string(b), "¿cómo es el formulario") {
		t.Fatalf("no se registró la pregunta: %q", b)
	}
}

func TestPreguntaSobreAciertosNoBuscaDocumentosNiInventaConteo(t *testing.T) {
	r, l, _ := preparar(t, "")
	res, err := r.Preguntar(context.Background(), "¿Cuántas preguntas acertaste hoy?", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Modo != "conversacional" || res.SinContexto || len(res.Fuentes) != 0 || l.visto != nil {
		t.Fatalf("la pregunta meta no debe ir al RAG ni al LLM: %+v, llamadas=%v", res, l.visto)
	}
	if !strings.Contains(res.Respuesta, "No llevo un contador fiable") {
		t.Fatalf("debe reconocer que no tiene una métrica verificable: %q", res.Respuesta)
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

func TestSeguimientosDeUbicacionSeReconocenComoGuia(t *testing.T) {
	for _, pregunta := range []string{"muéstrame dónde es", "¿dónde lo encuentro?", "me guías?"} {
		if !esHowTo(pregunta) {
			t.Errorf("%q debería activar el modo de orientación", pregunta)
		}
	}
}

func TestTipoConsultaOrquestaIntencionesEnEspanol(t *testing.T) {
	hilo := []Turno{{Rol: "usuario", Texto: "¿Qué es una partida de control?"}, {Rol: "asistente", Texto: "..."}}
	casos := []struct {
		pregunta string
		hilo     []Turno
		esperado string
	}{
		{"¿Cómo registro una partida?", nil, "procedimiento"},
		{"¿Qué significa partida de control?", nil, "concepto"},
		{"¿Cuál es la diferencia entre presupuesto venta y meta?", nil, "comparacion"},
		{"No aparece el botón y sale un error", nil, "problema"},
		{"¿Cuánto es el plazo indicado?", nil, "informacion_directa"},
		{"¿Y eso dónde lo encuentro?", hilo, "seguimiento"},
		{"No entiendo esa explicación", nil, "aclaracion"},
	}
	for _, caso := range casos {
		if got := tipoConsulta(caso.pregunta, caso.hilo); got != caso.esperado {
			t.Errorf("%q: tipo=%q, esperaba %q", caso.pregunta, got, caso.esperado)
		}
	}
}

func TestConsultaCompactaPreservaTerminosDelManual(t *testing.T) {
	if got := consultaCompacta("¿Qué es la partida de control según el Manual de Gerencia de Proyectos?"); got != "partida control gerencia proyectos" {
		t.Errorf("consulta compacta de Gerencia = %q", got)
	}
	if got := consultaCompacta("¿Cómo se factura un anticipo desde el módulo de Almacenes?"); got != "factura anticipo almacenes" {
		t.Errorf("consulta compacta de Almacenes = %q", got)
	}
}

func TestOrquestacionHaceVisibleLaRutaSegura(t *testing.T) {
	r, _ := armarConRuta(t, "")
	for _, caso := range []struct {
		pregunta  string
		ruta      string
		intencion string
	}{
		{"hola", rutaConversacion, clasificar.Social},
		{"¿qué precio tiene el Pro?", rutaRAG, clasificar.Trabajo},
	} {
		plan := r.planificar(context.Background(), caso.pregunta, nil)
		if plan.Ruta != caso.ruta || plan.Intencion != caso.intencion {
			t.Errorf("%q: plan=%+v; ruta/intención esperadas %q/%q", caso.pregunta, plan, caso.ruta, caso.intencion)
		}
		if plan.Clasificador == "" {
			t.Errorf("%q: el plan no identifica el clasificador", caso.pregunta)
		}
	}
}

func TestSeccionManualOficialEsEvidenciaTutorial(t *testing.T) {
	metadata := map[string]string{
		"confianza": "oficial", "source_url": "https://documentacion.s10peru.com/manual-de-presupuestos/",
		"section": "Registro del presupuesto",
	}
	if !esTutorial([]almacen.Resultado{{Metadata: metadata}}) {
		t.Fatal("un fragmento de sección oficial debe habilitar respuesta procedimental detallada")
	}
	metadata["confianza"] = "tercero-sin-verificar"
	if esTutorial([]almacen.Resultado{{Metadata: metadata}}) {
		t.Fatal("una copia no oficial no debe marcarse como guía oficial")
	}
}

func TestRegistroPresupuestoPriorizaPaginasDeAlta(t *testing.T) {
	if !esRegistroNuevoPresupuesto("¿Cómo registro un presupuesto?") {
		t.Fatal("no reconoció la intención de registrar un presupuesto")
	}
	candidatos := []almacen.Resultado{
		{Metadata: map[string]string{"title": "Guia de Usuario de S10 Presupuestos", "page": "11"}, Distancia: 0.2},
		{Metadata: map[string]string{"title": "Guia de Usuario de S10 Presupuestos", "page": "12"}, Distancia: 0.3},
		{Metadata: map[string]string{"title": "Guia de Usuario de S10 Presupuestos", "page": "17"}, Distancia: 0.4},
		{Metadata: map[string]string{"title": "Guia de Usuario de S10 Presupuestos", "page": "67"}, Distancia: 0.1},
		{Metadata: map[string]string{"title": "Manual de Gerencia de Proyectos", "page": "11"}, Distancia: 0.1},
	}
	got := contextoRegistroNuevoPresupuesto(candidatos)
	if len(got) != 3 {
		t.Fatalf("quería las tres páginas de alta del presupuesto, obtuvo %d", len(got))
	}
}

func TestRegistroPresupuestoRespondeConPasosVerificados(t *testing.T) {
	r, l, _ := preparar(t, "respuesta del modelo que no debe usarse")
	ctx := context.Background()
	var trozos []almacen.Trozo
	for _, pagina := range []string{"11", "12", "17"} {
		trozos = append(trozos, almacen.Trozo{
			ID:    "guia-p" + pagina,
			Texto: "Registro del nuevo presupuesto en Datos Generales, Nuevo SubItem, Adicionar y doble clic.",
			Metadata: map[string]string{"title": "Guia de Usuario de S10 Presupuestos", "page": pagina,
				"source": "s10-kb", "cita": "Guia de Usuario de S10 Presupuestos, p. " + pagina},
		})
	}
	if err := r.Almacen.Reemplazar(ctx, "guia-presupuestos", "kb", "v1", trozos); err != nil {
		t.Fatal(err)
	}
	res, err := r.Preguntar(ctx, "¿Cómo registro un presupuesto?", Opciones{Filtro: map[string]string{"source": "s10-kb"}})
	if err != nil {
		t.Fatal(err)
	}
	if l.visto != nil {
		t.Fatal("la intención de registro debe usar los pasos verificados, no generación libre")
	}
	for _, esperado := range []string{"Datos Generales", "Nuevo SubItem", "Adicionar", "doble clic"} {
		if !strings.Contains(res.Respuesta, esperado) {
			t.Errorf("respuesta sin paso %q: %s", esperado, res.Respuesta)
		}
	}
	if strings.Contains(strings.ToLower(res.Respuesta), "dimensiones") || strings.Contains(res.Respuesta, "Archivo Central") {
		t.Fatalf("respuesta mezcló otro procedimiento: %s", res.Respuesta)
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

func TestPreguntaLargaConInterrogacionVaRAG(t *testing.T) {
	r, _, _ := preparar(t, "")
	r.MaxDistancia = 2
	r.Emb = dosEjes{}
	m, err := clasificar.Entrenar(context.Background(), dosEjes{}, "dos-ejes", 0.0, map[string][]string{
		"social":  {"hola", "bien y tú"},
		"trabajo": {"precio del plan"},
	})
	if err != nil {
		t.Fatal(err)
	}
	r.Clasificador = m
	sec := &llmSecuencia{respuestas: []string{"Es una consultora [s3:rag-demo/docs/tarifas.md#0]"}}
	r.LLM = sec
	res, err := r.Preguntar(context.Background(), "bien, oye, qué es optimiza 360?", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Modo == "conversacional" && res.Motivo == "" {
		t.Fatalf("pregunta larga debió ir al RAG, no directo a charla: %+v", res)
	}
}

func TestTutorialDetallaSinTope(t *testing.T) {
	ctx := context.Background()
	a, _ := almacen.Abrir("", dosEjes{})
	a.Reemplazar(ctx, "s10kb:t1", indexar.FuenteS10KB, "e1", []almacen.Trozo{{
		ID: almacen.IDTrozo("s10kb:t1", 0), Texto: "precio con tutorial incluido",
		Metadata: map[string]string{"source": indexar.FuenteS10KB, "document_id": "cortex:tree/tutoriales/crear", "cita": "Tutorial: crear"},
	}})
	l := &llmFijo{respuesta: "1. Abre el módulo. 2. Registra."}
	r := &RAG{Almacen: a, LLM: l, MaxDistancia: 2, RutaFallos: filepath.Join(t.TempDir(), "f.jsonl")}
	res, err := r.Preguntar(ctx, "cómo ver precio con tutorial", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Modo != "tutorial" {
		t.Fatalf("how-to + contexto tutorial debió dar modo tutorial: %+v", res)
	}
	res2, err := r.Preguntar(ctx, "qué es el precio", Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	if res2.Modo != "respuesta" {
		t.Fatalf("pregunta conceptual no debe disparar tutorial: %+v", res2)
	}
	if esTutorial([]almacen.Resultado{{Metadata: map[string]string{"document_id": "x"}}}) {
		t.Fatal("sin tutorial no debe activar")
	}
}

func TestEsHowTo(t *testing.T) {
	si := []string{"cómo creo un presupuesto?", "pasos para calcular CTS",
		"guía de nóminas", "ayúdame a registrar", "cómo se anula"}
	no := []string{"qué es el jornal", "los precios quedan amarrados?",
		"hola", "cuánto cuesta"}
	for _, s := range si {
		if !esHowTo(s) {
			t.Errorf("debió ser how-to: %q", s)
		}
	}
	for _, s := range no {
		if esHowTo(s) {
			t.Errorf("no debió ser how-to: %q", s)
		}
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

func TestManualOficialGanaACopiaDeTerceroIgualDeCercana(t *testing.T) {
	copia := almacen.Resultado{ID: "copia", Distancia: 0.20, Texto: "Paso de una guía vieja",
		Metadata: map[string]string{"cita": "Guía Scribd, p. 4", "confianza": "tercero-sin-verificar"}}
	oficial := almacen.Resultado{ID: "oficial", Distancia: 0.22, Texto: "Paso del manual oficial",
		Metadata: map[string]string{"cita": "Manual de Compras, p. 9", "confianza": "oficial"}}
	sel := seleccionarContexto([]almacen.Resultado{copia, oficial}, 0.8)
	if len(sel) != 2 || sel[0].ID != "oficial" {
		t.Fatalf("el manual oficial debía ir primero: %+v", sel)
	}
	lejano := almacen.Resultado{ID: "lejano", Distancia: 0.40, Texto: "Otro tema",
		Metadata: map[string]string{"cita": "Manual de Almacenes, p. 1", "confianza": "oficial"}}
	cercana := almacen.Resultado{ID: "cercana", Distancia: 0.10, Texto: "Respuesta exacta",
		Metadata: map[string]string{"cita": "Guía Scribd, p. 7", "confianza": "tercero-sin-verificar"}}
	if sel := seleccionarContexto([]almacen.Resultado{lejano, cercana}, 0.8); sel[0].ID != "cercana" {
		t.Fatalf("una copia mucho más cercana no debe perder: %+v", sel)
	}
}
