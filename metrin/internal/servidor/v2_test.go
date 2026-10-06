package servidor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rag-go/internal/rag"
	"rag-go/internal/v2"
	"rag-go/internal/v2/tipos"
)

// conAgente: el mismo RAG de las pruebas con un orquestador V2 enchufado con la configuración dada.
func conAgente(t *testing.T, c v2.Config) *rag.RAG {
	t.Helper()
	r := armarRAG(t)
	ag := v2.Nuevo(c)
	ag.Charla = r
	r.V2 = ag
	return r
}

// Igualdad de V1: con la V2 apagada (por defecto), /ask responde byte a byte lo mismo que sin V2, traiga o no la
// petición los campos nuevos (version, memoria, conversacion, plan en el hilo), incluso con valores raros. Con la
// V2 habilitada, pedir «v1» también da V1 exacta.
func TestAskV1IdenticoConV2Apagada(t *testing.T) {
	aislarMedios(t)
	sinV2 := Nuevo(armarRAG(t), 5*time.Second)
	apagada := Nuevo(conAgente(t, v2.ConfigDefecto()), 5*time.Second)
	apagadaConTraza := Nuevo(conAgente(t, v2.ConfigDefecto()), 5*time.Second, Opciones{Traza: true})
	habilitada := v2.ConfigDefecto()
	habilitada.Habilitada = true
	hab := Nuevo(conAgente(t, habilitada), 5*time.Second)

	extras := map[string]string{
		"version v2":               `,"version":"v2"`,
		"version V2 mayúsculas":    `,"version":"V2"`,
		"version numérica":         `,"version":2`,
		"version objeto":           `,"version":{"x":1}`,
		"memoria válida":           `,"memoria":{"procedimiento_id":"metrados.registrar","paso_actual":2}`,
		"memoria basura":           `,"memoria":"basura"`,
		"memoria null":             `,"memoria":null`,
		"conversacion numérica":    `,"conversacion":123`,
		"todo junto":               `,"version":"v2","conversacion":"abc","memoria":{"paso_actual":1}`,
		"campo desconocido":        `,"otro":true`,
		"traza:true con V2 pedida": `,"traza":true,"version":"v2"`,
	}
	for _, pregunta := range []string{
		"¿qué precio tiene el Pro?",                                  // rag con fuente
		"¿cómo es el formulario de contacto que tiene validaciones?", // sin contexto
		"¿Cuántas preguntas acertaste hoy?",                          // regla
		"¿y luego?",                                                  // continuación (en V2 sería memoria)
	} {
		base := `{"pregunta":` + jsonTexto(pregunta)
		rrRef, ref := pedir(t, sinV2, base+`}`)
		if rrRef.Code != 200 {
			t.Fatalf("%q: %d %s", pregunta, rrRef.Code, ref)
		}
		if strings.Contains(string(ref), `"plan"`) || strings.Contains(string(ref), `"memoria"`) || strings.Contains(string(ref), `"version"`) {
			t.Fatalf("%q: una respuesta V1 no lleva plan/memoria/version: %s", pregunta, ref)
		}
		for nombre, extra := range extras {
			for hNombre, h := range map[string]http.Handler{"sin V2": sinV2, "V2 apagada": apagada} {
				rr, got := pedir(t, h, base+extra+`}`)
				if rr.Code != rrRef.Code || sinTiempos(got) != sinTiempos(ref) {
					t.Errorf("%q / %s / %s: la respuesta cambió\nref: %s\ngot: %s", pregunta, hNombre, nombre, ref, got)
				}
			}
		}
		// V2 habilitada pero la petición pide v1 (o no pide nada, con 0 %): V1 exacta.
		for _, extra := range []string{`,"version":"v1"`, ``, `,"version":"v3"`} {
			if rr, got := pedir(t, hab, base+extra+`}`); rr.Code != 200 || sinTiempos(got) != sinTiempos(ref) {
				t.Errorf("%q / habilitada %s: cambió\nref: %s\ngot: %s", pregunta, extra, ref, got)
			}
		}
		// Con traza: igual a V1 salvo la clave traza, y la traza no trae nada de la V2.
		_, conT := pedir(t, apagadaConTraza, base+`,"traza":true,"version":"v2"}`)
		var m map[string]any
		if err := json.Unmarshal(conT, &m); err != nil {
			t.Fatal(err)
		}
		tr := m["traza"].(map[string]any)
		if _, ok := tr["etapas_v2"]; ok || tr["agente"] != nil || len(tr["etapas"].([]any)) != 18 {
			t.Errorf("%q: traza V1 con rastro de V2: %v", pregunta, tr["agente"])
		}
		delete(m, "traza")
		var mr map[string]any
		json.Unmarshal(ref, &mr)
		a, _ := json.Marshal(m)
		b, _ := json.Marshal(mr)
		if sinTiempos(a) != sinTiempos(b) {
			t.Errorf("%q: con traza cambió algo más\nref: %s\ncon: %s", pregunta, b, a)
		}
	}
}

func TestAskHiloConPlanRaroNoRompeV1(t *testing.T) {
	aislarMedios(t)
	sinV2 := Nuevo(armarRAG(t), 5*time.Second)
	apagada := Nuevo(conAgente(t, v2.ConfigDefecto()), 5*time.Second)
	hilo := `[{"rol":"usuario","texto":"¿qué es el Pro?"},{"rol":"asistente","texto":"Un plan."%s}]`
	ref := `{"pregunta":"¿qué precio tiene?","hilo":` + strings.Replace(hilo, "%s", "", 1) + `}`
	rrRef, cuerpoRef := pedir(t, sinV2, ref)
	if rrRef.Code != 200 {
		t.Fatalf("%d %s", rrRef.Code, cuerpoRef)
	}
	for _, plan := range []string{`,"plan":"basura"`, `,"plan":{"version":2,"pasos":"x"}`, `,"plan":null`, `,"plan":[1,2]`} {
		cuerpo := `{"pregunta":"¿qué precio tiene?","hilo":` + strings.Replace(hilo, "%s", plan, 1) + `}`
		for nombre, h := range map[string]http.Handler{"sin V2": sinV2, "V2 apagada": apagada} {
			rr, got := pedir(t, h, cuerpo)
			if rr.Code != 200 || sinTiempos(got) != sinTiempos(cuerpoRef) {
				t.Errorf("%s %s: cambió\nref: %s\ngot: %s", nombre, plan, cuerpoRef, got)
			}
		}
	}
}

// Con la V2 habilitada y «version»: "v2", la respuesta conserva los campos de V1 y añade plan, memoria y version.
func TestAskV2PedidaDevuelvePlanYMemoria(t *testing.T) {
	aislarMedios(t)
	c := v2.ConfigDefecto()
	c.Habilitada = true
	h := Nuevo(conAgente(t, c), 5*time.Second, Opciones{Traza: true})
	rr, cuerpo := pedir(t, h, `{"pregunta":"¿Cómo registro un metrado?","version":"v2","traza":true}`)
	if rr.Code != 200 {
		t.Fatalf("%d %s", rr.Code, cuerpo)
	}
	var res struct {
		rag.Respuesta
		Traza struct {
			Agente   string           `json:"agente"`
			Etapas   []map[string]any `json:"etapas"`
			EtapasV2 []map[string]any `json:"etapas_v2"`
		} `json:"traza"`
	}
	if err := json.Unmarshal(cuerpo, &res); err != nil {
		t.Fatal(err)
	}
	if res.Version != tipos.V2 || res.PlanV2 == nil || res.Memoria == nil || res.Plan.Ruta != "v2" || res.Respuesta.Respuesta == "" {
		t.Fatalf("V2: %s", cuerpo)
	}
	// Sin piezas de conocimiento enchufadas: SIN_EVIDENCIA, sin fallar.
	if !res.SinContexto || !res.PlanV2.SinEvidencia || res.Fuentes == nil {
		t.Fatalf("sin conocimiento cargado: %s", cuerpo)
	}
	if res.Traza.Agente != "v2" || len(res.Traza.EtapasV2) != 13 || len(res.Traza.Etapas) != 18 {
		t.Fatalf("traza V2: agente %q, %d/%d etapas", res.Traza.Agente, len(res.Traza.EtapasV2), len(res.Traza.Etapas))
	}
	for _, e := range res.Traza.Etapas {
		if e["id"] == "fotos" && e["estado"] != "omitida" {
			t.Errorf("en V2 las fotos van por paso en el plan: %v", e)
		}
	}
	for _, clave := range []string{`"plan":`, `"memoria":`, `"version":"v2"`, `"orquestacion":`, `"fuentes":`, `"modo":`} {
		if !strings.Contains(string(cuerpo), clave) {
			t.Errorf("falta %s", clave)
		}
	}
}

func TestHealthConYSinV2(t *testing.T) {
	r := armarRAG(t)
	for _, caso := range []struct {
		v2    any
		tiene bool
	}{{nil, false}, {map[string]any{"habilitada": true}, true}} {
		rr := httptest.NewRecorder()
		Nuevo(r, time.Second, Opciones{V2: caso.v2}).ServeHTTP(rr, httptest.NewRequest("GET", "/health", nil))
		var h map[string]any
		json.Unmarshal(rr.Body.Bytes(), &h)
		if _, ok := h["v2"]; ok != caso.tiene || h["ok"] != true {
			t.Errorf("health = %v", h)
		}
	}
}
