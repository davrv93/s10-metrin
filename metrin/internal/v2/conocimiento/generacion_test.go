package conocimiento

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rag-go/internal/v2/tipos"
)

func planDe(t *testing.T, en entorno, q string, tipo tipos.TipoRespuesta) tipos.Plan {
	t.Helper()
	p, _, err := en.c.ConstruirConInforme(context.Background(), estado(en.b, q, tipo), nil)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestGeneracion_PlantillasEmbebidas(t *testing.T) {
	b := cargarFixtureEmbebidas(t)
	r := NuevoRecuperador(b)
	en := entorno{b: b, r: r, c: NuevoConstructor(b, r), g: NuevoGate(b), gen: NuevoMotorPlantillas(b)}
	p, _, txt := en.plan(t, "como registro metrado", tipos.Procedimiento)
	for _, s := range []string{"Para registrar el metrado de una partida, siga estos pasos.", "Antes de empezar:", "Estos son los pasos 1 a 3 de 3:",
		"1. Ubíquese en la partida del escenario **Hoja del presupuesto**.", "_Fuente: Manual de Presupuestos › 4.2 Registro de metrados_"} {
		if !strings.Contains(txt, s) {
			t.Errorf("falta %q:\n%s", s, txt)
		}
	}
	if i, j := strings.Index(txt, "1. "), strings.Index(txt, "3. "); i < 0 || j < i {
		t.Error("pasos fuera de orden")
	}
	// Fotos en el texto (opción): paso → foto → paso → foto, cada una debajo de SU paso.
	en.gen.FotosEnTexto = true
	txt, _ = en.gen.Redactar(context.Background(), p)
	i1 := strings.Index(txt, "1. Ubíquese")
	f1 := strings.Index(txt, "(fotos/imagenes/manual-de-presupuestos/aaa111aaa111.png)")
	i2 := strings.Index(txt, "2. Haga clic")
	f2 := strings.Index(txt, "(fotos/imagenes/manual-de-presupuestos/bbb222bbb222.png)")
	i3 := strings.Index(txt, "3. Registre")
	if !(i1 < f1 && f1 < i2 && i2 < f2 && f2 < i3) || strings.Count(txt, "![") != 2 {
		t.Errorf("orden paso → foto:\n%s", txt)
	}
	if res := en.g.Evaluar(p, txt, estado(b, "como registro metrado", tipos.Procedimiento)); !res.Paso {
		t.Errorf("gate con fotos en el texto: %v", res.Problemas)
	}
	// Planes incoherentes: error, no texto inventado.
	if _, err := en.gen.Redactar(context.Background(), tipos.Plan{Tipo: tipos.Concepto}); err == nil {
		t.Error("CONCEPT sin concepto debe dar error")
	}
	if _, err := en.gen.Redactar(context.Background(), tipos.Plan{Tipo: tipos.Procedimiento}); err == nil {
		t.Error("PROCEDURE sin pasos debe dar error")
	}
}

func TestGeneracion_PorTipo(t *testing.T) {
	en := nuevoEntorno(t)
	casos := []struct {
		q    string
		tipo tipos.TipoRespuesta
		hay  []string
	}{
		{"que es metrado", tipos.Concepto, []string{"**Metrado**: Cuantificación", "¿Quiere que le enseñe a registrar el metrado de una partida?"}},
		{"no puedo guardar metrado", tipos.Problema, []string{"No se puede grabar el metrado.", "Verifique que la partida tenga unidad de medida"}},
		{"partida", tipos.Desconocido, []string{"¿Qué quiere hacer:"}},
		{"receta de ceviche", tipos.Procedimiento, []string{"No encontré eso en los manuales"}},
	}
	for _, c := range casos {
		p := planDe(t, en, c.q, c.tipo)
		txt, err := en.gen.Redactar(context.Background(), p)
		if err != nil {
			t.Fatalf("%q: %v", c.q, err)
		}
		for _, s := range c.hay {
			if !strings.Contains(txt, s) {
				t.Errorf("%q: falta %q en:\n%s", c.q, s, txt)
			}
		}
	}
}

func TestGeneracion_ElegirMotor(t *testing.T) {
	b := cargarFixture(t)
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if m := NuevoMotorGeneracion(b, env(nil)); m.Nombre() != "plantillas" {
		t.Errorf("por defecto: %s", m.Nombre())
	}
	if m := NuevoMotorGeneracion(b, env(map[string]string{"GENERATION_ENGINE": "local"})); m.Nombre() != "plantillas" {
		t.Errorf("local sin modelo → plantillas: %s", m.Nombre())
	}
	m := NuevoMotorGeneracion(b, env(map[string]string{"GENERATION_ENGINE": "local", "GENERATION_MODEL": "qwen", "GENERATION_TIMEOUT_MS": "1234"}))
	ml, ok := m.(*MotorLocal)
	if !ok || ml.URL != URLGeneracionLocal || ml.Timeout != 1234*time.Millisecond || m.Nombre() != "local:qwen" {
		t.Errorf("local: %+v", m)
	}
	for in, want := range map[string]string{
		"http://h:8080":                "http://h:8080/v1/chat/completions",
		"http://h:11434/v1/":           "http://h:11434/v1/chat/completions",
		"http://h/v1/chat/completions": "http://h/v1/chat/completions",
	} {
		if got := urlChat(in); got != want {
			t.Errorf("urlChat(%q) = %q", in, got)
		}
	}
}

// servidorLLM: responde con lo que devuelva f a partir del JSON recibido (o un estado HTTP de error).
func servidorLLM(t *testing.T, estado int, f func(r reformulable) any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v1/chat/completions" {
			t.Errorf("ruta: %s", req.URL.Path)
		}
		var cuerpo struct {
			Model    string `json:"model"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		b, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(b, &cuerpo)
		if estado != http.StatusOK {
			w.WriteHeader(estado)
			return
		}
		var r reformulable
		if err := json.Unmarshal([]byte(cuerpo.Messages[len(cuerpo.Messages)-1].Content), &r); err != nil {
			t.Errorf("el motor no mandó el JSON del plan: %v", err)
		}
		var contenido string
		switch x := f(r).(type) {
		case string:
			contenido = x
		default:
			j, _ := json.Marshal(x)
			contenido = string(j)
		}
		json.NewEncoder(w).Encode(map[string]any{"choices": []map[string]any{{"message": map[string]string{"content": contenido}}}})
	}))
}

func motorLocal(b *Base, url string) *MotorLocal {
	return &MotorLocal{URL: url, Modelo: "prueba", Timeout: 2 * time.Second, Plantillas: NuevoMotorPlantillas(b)}
}

func TestGeneracion_LocalAceptaReformulacion(t *testing.T) {
	en := nuevoEntorno(t)
	p := planDe(t, en, "como registro metrado", tipos.Procedimiento)
	srv := servidorLLM(t, http.StatusOK, func(r reformulable) any {
		r.Intro = "Le explico cómo registrar el metrado de una partida en el módulo Presupuestos."
		r.Pasos[0].Texto = "Vaya a la partida, en el escenario **Hoja del presupuesto**."
		return r
	})
	defer srv.Close()
	m := motorLocal(en.b, srv.URL)
	txt, inf, err := m.RedactarConInforme(context.Background(), p)
	if err != nil || !inf.Reformulada || inf.Descartada {
		t.Fatalf("reformulación válida descartada: %+v %v", inf, err)
	}
	if !strings.Contains(txt, "1. Vaya a la partida") || !strings.Contains(txt, "Le explico cómo") || !strings.Contains(txt, "2. Haga clic") {
		t.Errorf("texto reformulado:\n%s", txt)
	}
	// El texto reformulado también pasa el gate (mismas negritas, mismo orden, mismas citas).
	if res := en.g.Evaluar(p, txt, estado(en.b, "como registro metrado", tipos.Procedimiento)); !res.Paso {
		t.Errorf("gate: %v", res.Problemas)
	}
	// Un aviso de SIN_EVIDENCIA o un concepto no van al modelo.
	s := planDe(t, en, "receta de ceviche", tipos.Procedimiento)
	if _, inf, _ := m.RedactarConInforme(context.Background(), s); inf.Reformulada || inf.Motor != "plantillas" {
		t.Errorf("SIN_EVIDENCIA no se reformula: %+v", inf)
	}
}

func TestGeneracion_LocalDescarta(t *testing.T) {
	en := nuevoEntorno(t)
	p := planDe(t, en, "como registro metrado", tipos.Procedimiento)
	base, _ := en.gen.Redactar(context.Background(), p)
	casos := map[string]func(r reformulable) any{
		"cambia una negrita": func(r reformulable) any {
			r.Pasos[1].Texto = "Haga clic en el botón **Metrados**."
			return r
		},
		"inventa una negrita en la intro": func(r reformulable) any {
			r.Intro = "Abra **Archivo** y luego siga los pasos."
			return r
		},
		"cambia una foto": func(r reformulable) any {
			r.Pasos[0].Fotos = []string{"b5f675189908"}
			return r
		},
		"quita la foto de un paso": func(r reformulable) any {
			r.Pasos[1].Fotos = []string{}
			return r
		},
		"cambia el orden": func(r reformulable) any {
			r.Pasos[0], r.Pasos[1] = r.Pasos[1], r.Pasos[0]
			return r
		},
		"quita un paso": func(r reformulable) any {
			r.Pasos = r.Pasos[:2]
			return r
		},
		"añade una cifra": func(r reformulable) any {
			r.Pasos[2].Texto = "Registre 3 cantidades en la columna **Metrado** y pulse **Grabar**."
			return r
		},
		"añade un atajo": func(r reformulable) any {
			r.Pasos[2].Texto = "Registre las cantidades en la columna **Metrado** y pulse **Grabar** (Ctrl+G)."
			return r
		},
		"añade una cita": func(r reformulable) any {
			r.Pasos[0].Texto = "Ubíquese en la partida del escenario **Hoja del presupuesto** (ver el manual, p. 12)."
			return r
		},
		"añade una ruta de menú": func(r reformulable) any {
			r.Intro = r.Intro + " Entre por Herramientas › Opciones."
			return r
		},
		"JSON inválido": func(r reformulable) any { return "esto no es JSON" },
	}
	for nombre, f := range casos {
		srv := servidorLLM(t, http.StatusOK, f)
		txt, err := motorLocal(en.b, srv.URL).Redactar(context.Background(), p)
		srv.Close()
		var d *Descarte
		if !errors.As(err, &d) || len(d.Motivos) == 0 {
			t.Errorf("%s: no se descartó (err=%v)", nombre, err)
		}
		if txt != base {
			t.Errorf("%s: tras descartar, el texto debe ser el de plantillas:\n%s", nombre, txt)
		}
	}
	// Sin LLM (caído, error HTTP, timeout): plantillas.
	srv := servidorLLM(t, http.StatusInternalServerError, nil)
	txt, err := motorLocal(en.b, srv.URL).Redactar(context.Background(), p)
	srv.Close()
	if txt != base || err == nil {
		t.Errorf("HTTP 500 → plantillas con aviso: %v", err)
	}
	fin := make(chan struct{})
	lento := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Second):
		case <-fin:
		}
	}))
	m := motorLocal(en.b, lento.URL)
	m.Timeout = 50 * time.Millisecond
	ini := time.Now()
	txt, inf, _ := m.RedactarConInforme(context.Background(), p)
	tardo := time.Since(ini)
	close(fin)
	lento.Close()
	if txt != base || !inf.Descartada || tardo > time.Second {
		t.Errorf("timeout → plantillas a tiempo (%v, %+v)", tardo, inf)
	}
	if txt, _ := motorLocal(en.b, "http://127.0.0.1:1").Redactar(context.Background(), p); txt != base {
		t.Error("sin servidor → plantillas")
	}
}
