package traza

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// contrato: las 18 etapas de docs/CONTRATO-traza-metrin.md, en orden.
var contrato = []struct{ id, nombre, tipo string }{
	{"inicio", "Pregunta", "evento"},
	{"entrada", "Entrada", "codigo"},
	{"regla_aciertos", "Regla «aciertos»", "decision"},
	{"embebedor", "Embebedor", "modelo"},
	{"clasificador", "Clasificador kNN", "modelo"},
	{"seguridad", "¿Social y corta?", "decision"},
	{"ruta", "Ruta", "decision"},
	{"reescritura", "Reescritura del hilo", "modelo"},
	{"busqueda", "Búsqueda", "codigo"},
	{"seleccion", "Selección", "codigo"},
	{"evidencia", "¿Hay contexto?", "decision"},
	{"generacion", "Generación LLM", "modelo"},
	{"verificacion", "Verificación de citas", "codigo"},
	{"fotos", "Fotos de pasos", "codigo"},
	{"charla", "Charla LLM", "modelo"},
	{"jev", "Juicio JEV", "modelo"},
	{"registro_fallos", "Registro de fallos", "codigo"},
	{"fin", "Respuesta", "evento"},
}

// porJSON serializa y vuelve a leer la traza como la vería la página.
func porJSON(t *testing.T, tr *Traza) map[string]any {
	t.Helper()
	b, err := json.Marshal(tr)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestLas18EtapasSiempreEnOrden(t *testing.T) {
	// Recorder vacío: nadie anotó nada y aun así salen las 18.
	tr := Nuevo("¿qué es un metrado?").Cerrar()
	if len(tr.Etapas) != len(contrato) {
		t.Fatalf("etapas = %d, el contrato pide %d", len(tr.Etapas), len(contrato))
	}
	for i, c := range contrato {
		e := tr.Etapas[i]
		if e.ID != c.id || e.Nombre != c.nombre || e.Tipo != c.tipo {
			t.Errorf("etapa %d = %s/%s/%s, quería %s/%s/%s", i+1, e.ID, e.Nombre, e.Tipo, c.id, c.nombre, c.tipo)
		}
		if !EstadoValido(e.Estado) {
			t.Errorf("%s: estado %q fuera del contrato", e.ID, e.Estado)
		}
		if (e.Estado == EstadoOmitida || e.Estado == EstadoNoTomada) && e.Razon == "" {
			t.Errorf("%s: %s sin razón", e.ID, e.Estado)
		}
		for _, k := range Minimos(e.ID) {
			if _, ok := e.Datos[k]; !ok {
				t.Errorf("%s: falta el dato mínimo %q", e.ID, k)
			}
		}
	}
	if got := IDs(); strings.Join(got, ",") != strings.Join(func() []string {
		var ids []string
		for _, c := range contrato {
			ids = append(ids, c.id)
		}
		return ids
	}(), ",") {
		t.Fatalf("IDs() = %v", got)
	}
	m := porJSON(t, tr)
	for _, k := range []string{"version", "total_ms", "pregunta", "truncada", "etapas", "oportunidades"} {
		if _, ok := m[k]; !ok {
			t.Errorf("la traza no lleva la clave %q", k)
		}
	}
	if m["version"].(float64) != 1 {
		t.Errorf("version = %v", m["version"])
	}
	if _, ok := m["oportunidades"].([]any); !ok {
		t.Errorf("oportunidades debe ser una lista (aunque vacía), es %T", m["oportunidades"])
	}
	for _, e := range m["etapas"].([]any) {
		etapa := e.(map[string]any)
		if _, ok := etapa["ms"]; !ok {
			t.Errorf("%v: falta «ms» (número o null)", etapa["id"])
		}
		if _, ok := etapa["datos"].(map[string]any); !ok {
			t.Errorf("%v: «datos» debe ser un objeto", etapa["id"])
		}
	}
}

func TestEstadosYTiempos(t *testing.T) {
	r := Nuevo("p")
	r.Inicio(EtapaEntrada)
	r.Fin(EtapaEntrada, EstadoOK, Datos{"pregunta": "p", "hilo_turnos": 0})
	r.FinDuracion(EtapaClasificador, EstadoOK, 1234567*time.Nanosecond, Datos{"intencion": "trabajo"})
	r.Fin(EtapaRuta, "inventado", nil) // estado fuera del contrato
	r.Omitir(EtapaReescritura, "sin hilo, no hay reescritura")
	r.NoTomada(EtapaCharla, "ruta rag")
	r.Inicio(EtapaBusqueda) // empieza y nunca termina
	tr := r.Cerrar()
	e := func(id string) Etapa { return tr.Etapas[indice[id]] }

	if ms := e(EtapaClasificador).Ms; ms == nil || *ms != 1.2 {
		t.Errorf("ms con un decimal: %v", ms)
	}
	if e(EtapaEntrada).Ms == nil {
		t.Error("Inicio+Fin debe medir ms")
	}
	if e(EtapaRuta).Estado != EstadoError {
		t.Errorf("un estado inválido se registra como error, no %q", e(EtapaRuta).Estado)
	}
	if e(EtapaRuta).Ms != nil {
		t.Error("Fin sin Inicio: ms debe ser null")
	}
	if got := e(EtapaReescritura); got.Estado != EstadoOmitida || got.Razon == "" || got.Ms != nil {
		t.Errorf("omitida: %+v", got)
	}
	if got := e(EtapaCharla); got.Estado != EstadoNoTomada || got.Razon != "ruta rag" {
		t.Errorf("no_tomada: %+v", got)
	}
	if got := e(EtapaBusqueda); got.Estado != EstadoError || got.Razon == "" {
		t.Errorf("iniciada sin cierre debe quedar en error: %+v", got)
	}
	if got := e(EtapaGeneracion); got.Estado != EstadoOmitida || got.Razon == "" {
		t.Errorf("pendiente debe quedar omitida con razón: %+v", got)
	}
	if tr.TotalMs < 0 {
		t.Errorf("total_ms = %v", tr.TotalMs)
	}
	for _, et := range tr.Etapas {
		if !EstadoValido(et.Estado) {
			t.Errorf("%s: estado %q", et.ID, et.Estado)
		}
	}
}

func TestRecortesA120(t *testing.T) {
	largo := strings.Repeat("palabra ", 60) // 480 caracteres
	r := Nuevo(largo)
	r.Fin(EtapaEntrada, EstadoOK, Datos{
		"pregunta": largo,
		"lista":    []string{largo},
		"mapa":     map[string]string{"source": largo},
		"anidado":  map[string]any{"x": largo},
	})
	r.Razon(EtapaEntrada, largo)
	r.Modelo(EtapaEntrada, largo)
	r.Fin(EtapaSeleccion, EstadoOK, Datos{"fragmentos": []Fragmento{{Documento: largo, Cita: largo, Seccion: largo}}})
	tr := r.Cerrar()
	b, _ := json.Marshal(tr)
	var m struct {
		Pregunta string `json:"pregunta"`
		Etapas   []struct {
			Razon  string         `json:"razon"`
			Modelo string         `json:"modelo"`
			Datos  map[string]any `json:"datos"`
		} `json:"etapas"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	cabe := func(donde, s string) {
		if n := utf8.RuneCountInString(s); n > MaxCaracteres {
			t.Errorf("%s: %d caracteres (> %d)", donde, n, MaxCaracteres)
		}
	}
	cabe("pregunta", m.Pregunta)
	entrada := m.Etapas[indice[EtapaEntrada]]
	cabe("razon", entrada.Razon)
	cabe("modelo", entrada.Modelo)
	cabe("datos.pregunta", entrada.Datos["pregunta"].(string))
	cabe("datos.lista", entrada.Datos["lista"].([]any)[0].(string))
	cabe("datos.mapa", entrada.Datos["mapa"].(map[string]any)["source"].(string))
	cabe("datos.anidado", entrada.Datos["anidado"].(map[string]any)["x"].(string))
	fr := m.Etapas[indice[EtapaSeleccion]].Datos["fragmentos"].([]any)[0].(map[string]any)
	for _, k := range []string{"documento", "cita", "seccion"} {
		cabe("fragmento."+k, fr[k].(string))
	}
	if got := Recortar("  dos   líneas\nde texto "); got != "dos líneas de texto" {
		t.Errorf("Recortar no deja una sola línea: %q", got)
	}
	if got := Recortar(largo); !strings.HasSuffix(got, "…") {
		t.Errorf("un recorte debe terminar en «…»: %q", got)
	}
}

func TestTruncaA16KB(t *testing.T) {
	r := Nuevo("p")
	var fr []Fragmento
	for i := 0; i < 200; i++ {
		fr = append(fr, Fragmento{Documento: strings.Repeat("d", 110), Cita: strings.Repeat("c", 110), Distancia: 0.2})
	}
	r.Fin(EtapaSeleccion, EstadoOK, Datos{"fragmentos": fr, "seleccionados": len(fr)})
	tr := r.Cerrar()
	if !tr.Truncada {
		t.Fatal("pasó de 16 KB y no se marcó truncada")
	}
	if got := tr.Etapas[indice[EtapaSeleccion]].Datos["fragmentos"].([]Fragmento); len(got) != FragmentosTruncada {
		t.Fatalf("fragmentos tras truncar = %d, quería %d", len(got), FragmentosTruncada)
	}
	if b, _ := json.Marshal(tr); len(b) > MaxBytes {
		t.Fatalf("la traza truncada mide %d B (> %d)", len(b), MaxBytes)
	}

	chica := Nuevo("p")
	chica.Fin(EtapaSeleccion, EstadoOK, Datos{"fragmentos": fr[:3]})
	if tr := chica.Cerrar(); tr.Truncada || len(tr.Etapas[indice[EtapaSeleccion]].Datos["fragmentos"].([]Fragmento)) != 3 {
		t.Fatal("una traza pequeña no se trunca")
	}
}

func TestOportunidadesPorRegla(t *testing.T) {
	casos := []struct {
		nombre string
		anotar func(r *Recorder)
		etapa  string
		nivel  string
		estado string
	}{
		{"bajo el umbral y ruta rag", func(r *Recorder) {
			r.Fin(EtapaRuta, EstadoOK, Datos{"ruta": "rag"})
			r.Fin(EtapaClasificador, EstadoOK, Datos{"intencion": "trabajo", "similitud": 0.31, "umbral": 0.40, "paso_umbral": false})
		}, EtapaClasificador, EstadoAlerta, EstadoAlerta},
		{"social en ruta rag", func(r *Recorder) {
			r.Fin(EtapaRuta, EstadoOK, Datos{"ruta": "rag"})
			r.Fin(EtapaClasificador, EstadoOK, Datos{"intencion": "social", "similitud": 0.569, "umbral": 0.40, "paso_umbral": true})
		}, EtapaClasificador, EstadoAlerta, EstadoAlerta},
		{"limite en ruta rag", func(r *Recorder) {
			r.Fin(EtapaRuta, EstadoOK, Datos{"ruta": "rag"})
			r.Fin(EtapaClasificador, EstadoOK, Datos{"intencion": "limite", "similitud": 0.7, "umbral": 0.40, "paso_umbral": true})
		}, EtapaClasificador, EstadoAlerta, EstadoAlerta},
		{"fragmento penalizado", func(r *Recorder) {
			r.Fin(EtapaSeleccion, EstadoOK, Datos{"fragmentos": []Fragmento{{Documento: "Guía Scribd", Confianza: "tercero-sin-verificar", Penalizacion: 0.04}}})
		}, EtapaSeleccion, EstadoAlerta, EstadoAlerta},
		{"sin contexto", func(r *Recorder) {
			r.Fin(EtapaEvidencia, EstadoOK, Datos{"sin_contexto": true, "distancia_min": 0.91, "max_distancia": 0.8})
		}, EtapaEvidencia, EstadoAlerta, EstadoAlerta},
		{"generación con error", func(r *Recorder) {
			r.Fin(EtapaGeneracion, EstadoError, nil)
			r.Razon(EtapaGeneracion, "el modelo falló")
		}, EtapaGeneracion, EstadoError, EstadoError},
		{"charla con texto de respaldo", func(r *Recorder) {
			r.Fin(EtapaCharla, EstadoError, nil)
		}, EtapaCharla, EstadoError, EstadoError},
		{"reescritura con error", func(r *Recorder) {
			r.Fin(EtapaReescritura, EstadoError, nil)
		}, EtapaReescritura, EstadoError, EstadoError},
		{"jev con error", func(r *Recorder) {
			r.Fin(EtapaJEV, EstadoError, Datos{"activo": true})
		}, EtapaJEV, EstadoError, EstadoError},
	}
	for _, c := range casos {
		r := Nuevo("p")
		c.anotar(r)
		tr := r.Cerrar()
		if len(tr.Oportunidades) != 1 {
			t.Errorf("%s: %d oportunidades, quería 1: %+v", c.nombre, len(tr.Oportunidades), tr.Oportunidades)
			continue
		}
		o := tr.Oportunidades[0]
		if o.N != 1 || o.Etapa != c.etapa || o.Nivel != c.nivel || o.Texto == "" {
			t.Errorf("%s: oportunidad %+v", c.nombre, o)
		}
		if got := tr.Etapas[indice[c.etapa]].Estado; got != c.estado {
			t.Errorf("%s: estado de %s = %q, quería %q", c.nombre, c.etapa, got, c.estado)
		}
	}

	// Sin hechos que la disparen, ninguna regla salta.
	r := Nuevo("p")
	r.Fin(EtapaRuta, EstadoOK, Datos{"ruta": "conversacion"})
	r.Fin(EtapaClasificador, EstadoOK, Datos{"intencion": "social", "similitud": 0.9, "umbral": 0.40, "paso_umbral": true})
	r.Fin(EtapaSeleccion, EstadoOK, Datos{"fragmentos": []Fragmento{{Documento: "Manual", Confianza: "oficial"}}})
	r.Fin(EtapaEvidencia, EstadoOK, Datos{"sin_contexto": false})
	if tr := r.Cerrar(); len(tr.Oportunidades) != 0 {
		t.Errorf("social en ruta conversacion no es oportunidad: %+v", tr.Oportunidades)
	}

	// Varias a la vez: numeradas desde 1 en el orden de las etapas.
	r = Nuevo("p")
	r.Fin(EtapaGeneracion, EstadoError, nil)
	r.Fin(EtapaEvidencia, EstadoOK, Datos{"sin_contexto": true})
	r.Fin(EtapaRuta, EstadoOK, Datos{"ruta": "rag"})
	r.Fin(EtapaClasificador, EstadoOK, Datos{"intencion": "social", "similitud": 0.57, "umbral": 0.40, "paso_umbral": true})
	tr := r.Cerrar()
	var orden []string
	for i, o := range tr.Oportunidades {
		if o.N != i+1 {
			t.Errorf("numeración: %+v", tr.Oportunidades)
		}
		orden = append(orden, o.Etapa)
	}
	if strings.Join(orden, ",") != "clasificador,evidencia,generacion" {
		t.Errorf("orden de oportunidades = %v", orden)
	}
}

func TestRecorderNilEsNoOp(t *testing.T) {
	var r *Recorder
	r.Inicio(EtapaBusqueda)
	r.Fin(EtapaBusqueda, EstadoOK, Datos{"k": 8})
	r.FinDuracion(EtapaBusqueda, EstadoOK, time.Millisecond, nil)
	r.Omitir(EtapaBusqueda, "x")
	r.NoTomada(EtapaBusqueda, "x")
	r.OmitirCon(EtapaBusqueda, EstadoOmitida, "x", nil)
	r.Dato(EtapaBusqueda, "k", 8)
	r.Razon(EtapaBusqueda, "x")
	r.Modelo(EtapaBusqueda, "x")
	r.CambiarEstado(EtapaBusqueda, EstadoAlerta, "x")
	r.FallarIniciadas("x")
	if v, ok := r.DatoDe(EtapaBusqueda, "k"); v != nil || ok {
		t.Error("DatoDe sobre nil")
	}
	if r.Estado(EtapaBusqueda) != "" || r.Pendiente(EtapaBusqueda) {
		t.Error("Estado/Pendiente sobre nil")
	}
	if r.Cerrar() != nil {
		t.Error("Cerrar sobre nil debe devolver nil")
	}
	ctx := context.Background()
	if ConContexto(ctx, nil) != ctx {
		t.Error("ConContexto con nil debe devolver el mismo contexto")
	}
	if De(ctx) != nil {
		t.Error("sin recorder en el contexto, De debe ser nil")
	}
	// Con el modo apagado, instrumentar no reserva memoria.
	if n := testing.AllocsPerRun(100, func() {
		rec := De(ctx)
		rec.Inicio(EtapaBusqueda)
		rec.Fin(EtapaBusqueda, EstadoOK, nil)
		rec.Omitir(EtapaGeneracion, "x")
		rec.Dato(EtapaBusqueda, "k", 8)
	}); n != 0 {
		t.Errorf("el recorder apagado reservó memoria: %v por llamada", n)
	}
	vivo := Nuevo("p")
	if De(ConContexto(ctx, vivo)) != vivo {
		t.Error("De no devuelve el recorder del contexto")
	}
}

func TestResumirErrorConservaLaCausa(t *testing.T) {
	raiz := errors.New("connection refused")
	largo := fmt.Errorf("mlx chat (http://host.docker.internal:8080): %w",
		fmt.Errorf("Post \"http://host.docker.internal:8080/v1/chat/completions\": %w",
			fmt.Errorf("dial tcp 192.168.65.254:8080: connect: %w", raiz)))
	got := ResumirError(largo)
	if got != "mlx chat (http://host.docker.internal:8080): connection refused" {
		t.Errorf("ResumirError = %q", got)
	}
	if utf8.RuneCountInString(got) > MaxCaracteres {
		t.Errorf("más de %d caracteres", MaxCaracteres)
	}
	if got := ResumirError(errors.New("ollama chat: HTTP 500: boom")); got != "ollama chat: HTTP 500: boom" {
		t.Errorf("sin envoltura debe quedar igual: %q", got)
	}
	if ResumirError(nil) != "" {
		t.Error("nil debe dar texto vacío")
	}
}

func TestHabilitada(t *testing.T) {
	for v, quiero := range map[string]bool{"1": true, "true": true, "TRUE": true, " 1 ": true,
		"": false, "0": false, "false": false, "si": false, "yes": false} {
		if Habilitada(v) != quiero {
			t.Errorf("Habilitada(%q) = %v", v, !quiero)
		}
	}
}
