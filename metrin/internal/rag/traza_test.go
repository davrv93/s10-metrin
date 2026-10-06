package rag

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"rag-go/internal/clasificar"
	"rag-go/internal/traza"
)

// conTraza pregunta con el modo traza encendido y devuelve la traza cerrada.
func conTraza(t *testing.T, r *RAG, pregunta string, o Opciones) (Respuesta, *traza.Traza) {
	t.Helper()
	rec := traza.Nuevo(pregunta)
	res, err := r.Preguntar(traza.ConContexto(context.Background(), rec), pregunta, o)
	if err != nil {
		t.Fatal(err)
	}
	tr := rec.Cerrar()
	if len(tr.Etapas) != len(traza.IDs()) {
		t.Fatalf("etapas = %d", len(tr.Etapas))
	}
	for i, id := range traza.IDs() {
		e := tr.Etapas[i]
		if e.ID != id {
			t.Fatalf("etapa %d = %q, quería %q", i, e.ID, id)
		}
		if !traza.EstadoValido(e.Estado) {
			t.Errorf("%s: estado %q", id, e.Estado)
		}
		if (e.Estado == traza.EstadoOmitida || e.Estado == traza.EstadoNoTomada) && e.Razon == "" {
			t.Errorf("%s: %s sin razón", id, e.Estado)
		}
	}
	return res, tr
}

func etapa(tr *traza.Traza, id string) traza.Etapa {
	for _, e := range tr.Etapas {
		if e.ID == id {
			return e
		}
	}
	return traza.Etapa{}
}

// esperar comprueba el estado de varias etapas a la vez.
func esperar(t *testing.T, tr *traza.Traza, estados map[string]string) {
	t.Helper()
	for id, quiero := range estados {
		if got := etapa(tr, id); got.Estado != quiero {
			t.Errorf("%s: estado %q, quería %q (razón: %s)", id, got.Estado, quiero, got.Razon)
		}
	}
}

func TestTrazaRutaRAG(t *testing.T) {
	r, _ := armarConRuta(t, "Cuesta 49 € [s3:rag-demo/docs/tarifas.md#0]")
	res, tr := conTraza(t, r, "¿qué precio tiene el Pro?", Opciones{})
	if res.Modo != "respuesta" {
		t.Fatalf("modo %q", res.Modo)
	}
	esperar(t, tr, map[string]string{
		"inicio": "ok", "entrada": "ok", "regla_aciertos": "ok", "embebedor": "ok",
		"clasificador": "ok", "seguridad": "ok", "ruta": "ok", "reescritura": "omitida",
		"busqueda": "ok", "seleccion": "ok", "evidencia": "ok", "generacion": "ok",
		"verificacion": "ok", "charla": "no_tomada", "jev": "omitida",
		"registro_fallos": "omitida", "fin": "ok",
	})
	if d := etapa(tr, "regla_aciertos").Datos; d["aplica"] != false {
		t.Errorf("regla_aciertos.aplica = %v", d["aplica"])
	}
	cl := etapa(tr, "clasificador")
	if cl.Datos["intencion"] != clasificar.Trabajo || cl.Datos["paso_umbral"] != true || cl.Modelo != "kNN-local:dos-ejes" {
		t.Errorf("clasificador: %+v", cl)
	}
	if cl.Ms == nil || etapa(tr, "embebedor").Ms == nil {
		t.Error("embebedor y clasificador deben llevar ms")
	}
	if d := etapa(tr, "embebedor").Datos; d["modelo"] != "dos-ejes" {
		t.Errorf("embebedor: %+v", d)
	}
	if d := etapa(tr, "ruta").Datos; d["ruta"] != "rag" || d["tipo_consulta"] == nil {
		t.Errorf("ruta: %+v", d)
	}
	b := etapa(tr, "busqueda").Datos
	if b["metodo"] != "vectorial" || b["k"] != 8 || b["resultados"].(int) < 1 {
		t.Errorf("busqueda: %+v", b)
	}
	s := etapa(tr, "seleccion").Datos
	fr := s["fragmentos"].([]traza.Fragmento)
	if s["ventana"] != ventanaRelevancia || s["max_contextos"] != maxContextos || s["seleccionados"] != 1 || len(fr) != 1 || s["corte"] == nil {
		t.Errorf("seleccion: %+v", s)
	}
	if fr[0].Cita != "s3:rag-demo/docs/tarifas.md#0" {
		t.Errorf("fragmento: %+v", fr[0])
	}
	if d := etapa(tr, "evidencia").Datos; d["sin_contexto"] != false || d["distancia_min"] != res.DistanciaMin {
		t.Errorf("evidencia: %+v", d)
	}
	// El doble de prueba no expone su modelo: null y razón, no un nombre inventado.
	g := etapa(tr, "generacion")
	if g.Datos["modelo"] != nil || g.Datos["proveedor"] != nil || g.Datos["modo"] != "respuesta" {
		t.Errorf("generacion: %+v", g)
	}
	if d := etapa(tr, "verificacion").Datos; d["n_citas"] != 1 {
		t.Errorf("verificacion: %+v", d)
	}
	if d := etapa(tr, "jev").Datos; d["activo"] != false {
		t.Errorf("jev: %+v", d)
	}
	if d := etapa(tr, "fin").Datos; d["modo"] != "respuesta" || d["ruta"] != "rag" {
		t.Errorf("fin: %+v", d)
	}
	if len(tr.Oportunidades) != 0 {
		t.Errorf("sin hechos raros no hay oportunidades: %+v", tr.Oportunidades)
	}
}

func TestTrazaRutaConversacion(t *testing.T) {
	r, _ := armarConRuta(t, "Hola, ¿en qué te ayudo?")
	res, tr := conTraza(t, r, "hola", Opciones{})
	if res.Modo != "conversacional" {
		t.Fatalf("modo %q", res.Modo)
	}
	esperar(t, tr, map[string]string{
		"regla_aciertos": "ok", "embebedor": "ok", "clasificador": "ok", "seguridad": "ok",
		"ruta": "ok", "charla": "ok",
		"reescritura": "no_tomada", "busqueda": "no_tomada", "seleccion": "no_tomada",
		"evidencia": "no_tomada", "generacion": "no_tomada", "verificacion": "no_tomada",
		"registro_fallos": "omitida", "fin": "ok",
	})
	if d := etapa(tr, "seguridad").Datos; d["desvia_a_charla"] != true || d["pregunta_larga"] != false {
		t.Errorf("seguridad: %+v", d)
	}
	if d := etapa(tr, "ruta").Datos; d["ruta"] != "conversacion" || d["tipo_consulta"] != "social" {
		t.Errorf("ruta: %+v", d)
	}
	if d := etapa(tr, "clasificador").Datos; d["intencion"] != clasificar.Social {
		t.Errorf("clasificador: %+v", d)
	}
}

func TestTrazaReglaAciertos(t *testing.T) {
	r, sec := armarConRuta(t, "")
	res, tr := conTraza(t, r, "¿Cuántas preguntas acertaste hoy?", Opciones{})
	if res.Plan.Ruta != rutaRegla || sec.llamadas != 0 {
		t.Fatalf("la regla debió responder sin LLM: %+v", res)
	}
	esperar(t, tr, map[string]string{
		"regla_aciertos": "ok", "ruta": "ok",
		// El clasificador corre antes que la regla (planificar va primero).
		"embebedor": "ok", "clasificador": "ok",
		"reescritura": "omitida", "busqueda": "omitida", "seleccion": "omitida",
		"evidencia": "omitida", "generacion": "omitida", "verificacion": "omitida",
		"charla": "omitida", "registro_fallos": "omitida", "fin": "ok",
	})
	if d := etapa(tr, "regla_aciertos").Datos; d["aplica"] != true {
		t.Errorf("regla_aciertos: %+v", d)
	}
	if d := etapa(tr, "ruta").Datos; d["ruta"] != "regla" || d["tipo_consulta"] != "metricas" {
		t.Errorf("ruta: %+v", d)
	}
	if d := etapa(tr, "fin").Datos; d["modo"] != "conversacional" {
		t.Errorf("fin: %+v", d)
	}
}

func TestTrazaRAGConHilo(t *testing.T) {
	r, _, _ := preparar(t, "")
	r.MaxDistancia = 2
	r.LLM = &llmSecuencia{respuestas: []string{
		"¿Quién es el CEO de Optimiza 360?",
		"Es Arturo Dupont [s3:rag-demo/docs/tarifas.md#0]",
	}}
	hilo := []Turno{
		{Rol: "usuario", Texto: "¿qué es optimiza 360?"},
		{Rol: "asistente", Texto: "Es una consultora."},
	}
	res, tr := conTraza(t, r, "¿quién es su ceo?", Opciones{Hilo: hilo})
	esperar(t, tr, map[string]string{
		"reescritura": "ok", "busqueda": "ok", "seleccion": "ok", "evidencia": "ok",
		"generacion": "ok", "verificacion": "ok", "charla": "no_tomada",
		// Sin clasificador: reglas_seguras.
		"clasificador": "omitida", "seguridad": "omitida", "embebedor": "ok",
	})
	re := etapa(tr, "reescritura")
	if re.Datos["original"] != "¿quién es su ceo?" || re.Datos["reescrita"] != res.PreguntaReescrita || re.Ms == nil {
		t.Errorf("reescritura: %+v", re)
	}
	if d := etapa(tr, "entrada").Datos; d["hilo_turnos"] != 2 {
		t.Errorf("entrada: %+v", d)
	}
	if d := etapa(tr, "clasificador").Datos; d["clasificador"] != "reglas_seguras" {
		t.Errorf("clasificador: %+v", d)
	}
	if e := etapa(tr, "embebedor"); e.Ms != nil {
		t.Errorf("sin clasificador el embebido va dentro de la búsqueda: ms debe ser null, %+v", e)
	}
}

func TestTrazaModeloCaido(t *testing.T) {
	r, _, _ := preparar(t, "")
	r.LLM = llmCaido{}
	res, tr := conTraza(t, r, "¿qué precio tiene el Pro?", Opciones{})
	if res.Modo != "modelo_no_disponible" {
		t.Fatalf("modo %q", res.Modo)
	}
	esperar(t, tr, map[string]string{"generacion": "error", "verificacion": "omitida", "fin": "ok"})
	if len(tr.Oportunidades) != 1 || tr.Oportunidades[0].Etapa != "generacion" || tr.Oportunidades[0].Nivel != "error" {
		t.Errorf("oportunidades: %+v", tr.Oportunidades)
	}
	if d := etapa(tr, "fin").Datos; d["modo"] != "modelo_no_disponible" {
		t.Errorf("fin: %+v", d)
	}
}

func TestTrazaSinContexto(t *testing.T) {
	r, _, _ := preparar(t, "no debería llamarse")
	_, tr := conTraza(t, r, "¿cómo es el formulario de contacto que tiene validaciones?", Opciones{})
	esperar(t, tr, map[string]string{
		"evidencia": "alerta", "generacion": "omitida", "verificacion": "omitida",
		"registro_fallos": "ok", "charla": "no_tomada",
	})
	if d := etapa(tr, "registro_fallos").Datos; d["registrado"] != true {
		t.Errorf("registro_fallos: %+v", d)
	}
	if d := etapa(tr, "evidencia").Datos; d["sin_contexto"] != true || d["motivo"] == nil {
		t.Errorf("evidencia: %+v", d)
	}
	var hay bool
	for _, o := range tr.Oportunidades {
		hay = hay || (o.Etapa == "evidencia" && o.Nivel == "alerta")
	}
	if !hay {
		t.Errorf("sin contexto debe dar oportunidad en evidencia: %+v", tr.Oportunidades)
	}
}

func TestTrazaCharlaDeRespaldo(t *testing.T) {
	r, _, _ := preparar(t, "")
	r.LLM = &llmSecuencia{respuestas: []string{"Bien, ¿y tú?"}}
	res, tr := conTraza(t, r, "bien y tú", Opciones{})
	if res.Plan.Ruta != "conversacion_fallback" {
		t.Fatalf("ruta %q", res.Plan.Ruta)
	}
	esperar(t, tr, map[string]string{
		"evidencia": "alerta", "charla": "respaldo", "generacion": "omitida", "registro_fallos": "ok",
	})
	if d := etapa(tr, "fin").Datos; d["ruta"] != "conversacion_fallback" {
		t.Errorf("fin: %+v", d)
	}
}

func TestTrazaSocialEnPreguntaDeTrabajo(t *testing.T) {
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
	r.LLM = &llmSecuencia{respuestas: []string{"Es una consultora [s3:rag-demo/docs/tarifas.md#0]"}}
	_, tr := conTraza(t, r, "bien, oye, qué es optimiza 360?", Opciones{})
	esperar(t, tr, map[string]string{"clasificador": "alerta", "seguridad": "ok", "ruta": "ok"})
	if d := etapa(tr, "seguridad").Datos; d["pregunta_larga"] != true || d["desvia_a_charla"] != false {
		t.Errorf("seguridad: %+v", d)
	}
	if len(tr.Oportunidades) == 0 || tr.Oportunidades[0].Etapa != "clasificador" {
		t.Errorf("oportunidades: %+v", tr.Oportunidades)
	}
}

var reTiempos = regexp.MustCompile(`"ms_(busqueda|llm)":\d+`)

// sinTiempos: los únicos campos de /ask que dependen del reloj.
func sinTiempos(b []byte) string {
	return reTiempos.ReplaceAllString(string(b), `"ms_$1":0`)
}

// La traza solo mira: para las mismas entradas, la respuesta es la misma con
// y sin recorder en el contexto.
func TestTrazaNoCambiaLaRespuesta(t *testing.T) {
	casos := []struct {
		nombre   string
		armar    func(t *testing.T) *RAG
		pregunta string
		o        Opciones
	}{
		{"rag", func(t *testing.T) *RAG {
			r, _ := armarConRuta(t, "Cuesta 49 € [s3:rag-demo/docs/tarifas.md#0]")
			return r
		}, "¿qué precio tiene el Pro?", Opciones{}},
		{"conversacion", func(t *testing.T) *RAG {
			r, _ := armarConRuta(t, "Hola, ¿en qué te ayudo?")
			return r
		}, "hola", Opciones{}},
		{"aciertos", func(t *testing.T) *RAG {
			r, _ := armarConRuta(t, "")
			return r
		}, "¿Cuántas preguntas acertaste hoy?", Opciones{}},
		{"hilo", func(t *testing.T) *RAG {
			r, _, _ := preparar(t, "")
			r.MaxDistancia = 2
			r.LLM = &llmSecuencia{respuestas: []string{"¿Quién es el CEO de Optimiza 360?", "Es Arturo Dupont"}}
			return r
		}, "¿quién es su ceo?", Opciones{Hilo: []Turno{{Rol: "usuario", Texto: "¿qué es optimiza 360?"}, {Rol: "asistente", Texto: "Una consultora."}}}},
		{"sin_contexto", func(t *testing.T) *RAG {
			r, _, _ := preparar(t, "x")
			return r
		}, "¿cómo es el formulario de contacto que tiene validaciones?", Opciones{}},
		{"modelo_caido", func(t *testing.T) *RAG {
			r, _, _ := preparar(t, "")
			r.LLM = llmCaido{}
			return r
		}, "¿qué precio tiene el Pro?", Opciones{}},
		{"portugues", func(t *testing.T) *RAG {
			r, _, _ := preparar(t, "")
			r.LLM = &llmSecuencia{respuestas: []string{"Também uma coisa", "Não há como"}}
			return r
		}, "¿qué precio tiene el Pro?", Opciones{}},
	}
	for _, c := range casos {
		sin, errSin := c.armar(t).Preguntar(context.Background(), c.pregunta, c.o)
		rec := traza.Nuevo(c.pregunta)
		con, errCon := c.armar(t).Preguntar(traza.ConContexto(context.Background(), rec), c.pregunta, c.o)
		if (errSin == nil) != (errCon == nil) {
			t.Errorf("%s: errores distintos %v / %v", c.nombre, errSin, errCon)
		}
		a, _ := json.Marshal(sin)
		b, _ := json.Marshal(con)
		if sinTiempos(a) != sinTiempos(b) {
			t.Errorf("%s: la traza cambió la respuesta\nsin: %s\ncon: %s", c.nombre, a, b)
		}
		if rec.Cerrar() == nil {
			t.Errorf("%s: sin traza", c.nombre)
		}
	}
}
