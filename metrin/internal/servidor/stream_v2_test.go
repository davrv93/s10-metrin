package servidor

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rag-go/internal/llm"
	"rag-go/internal/v2"
)

// eventosStream hace POST /ask/stream y separa los eventos: los textos de «delta», la «respuesta» de «fin» (en
// crudo, para compararla con /ask) y el evento «error» si lo hubo.
func eventosStream(t *testing.T, h http.Handler, cuerpo string) (rr *httptest.ResponseRecorder, deltas []string, fin json.RawMessage, fallo map[string]any) {
	t.Helper()
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("POST", "/ask/stream", strings.NewReader(cuerpo)))
	sc := bufio.NewScanner(rr.Body)
	sc.Buffer(make([]byte, 64<<10), 4<<20)
	for sc.Scan() {
		var e struct {
			Tipo      string          `json:"tipo"`
			Texto     string          `json:"texto"`
			Respuesta json.RawMessage `json:"respuesta"`
		}
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			t.Fatalf("línea no es JSON: %q", sc.Text())
		}
		switch e.Tipo {
		case "delta":
			deltas = append(deltas, e.Texto)
		case "fin":
			fin = e.Respuesta
		case "error":
			json.Unmarshal(sc.Bytes(), &fallo)
		default:
			t.Fatalf("evento inesperado: %s", sc.Text())
		}
	}
	return rr, deltas, fin, fallo
}

// sinTraza quita la clave «traza» (sus tiempos cambian en cada llamada) y normaliza ms_*.
func sinTraza(t *testing.T, b []byte) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("no es JSON: %s", b)
	}
	delete(m, "traza")
	out, _ := json.Marshal(m)
	return sinTiempos(out)
}

// El «fin» de /ask/stream es el mismo JSON que /ask, en V1 y en V2, con o sin traza: la página usa siempre
// /ask/stream y pinta la V2 (plan → pasos con fotos) y la traza con la misma respuesta que antes le daba /ask.
func TestAskStreamFinIgualQueAskEnV1YV2(t *testing.T) {
	aislarMedios(t)
	c := v2.ConfigDefecto()
	c.Habilitada = true
	h := Nuevo(conAgente(t, c), 5*time.Second, Opciones{Traza: true})
	for nombre, cuerpo := range map[string]string{
		"V1":              `{"pregunta":"¿qué precio tiene el Pro?"}`,
		"V1 con traza":    `{"pregunta":"¿qué precio tiene el Pro?","traza":true}`,
		"V1 sin contexto": `{"pregunta":"¿cómo es el formulario de contacto que tiene validaciones?"}`,
		"V2":              `{"pregunta":"¿Cómo registro un metrado?","version":"v2"}`,
		"V2 con traza":    `{"pregunta":"¿Cómo registro un metrado?","version":"v2","traza":true}`,
	} {
		rrAsk, ask := pedir(t, h, cuerpo)
		if rrAsk.Code != 200 {
			t.Fatalf("%s: /ask %d %s", nombre, rrAsk.Code, ask)
		}
		rr, deltas, fin, fallo := eventosStream(t, h, cuerpo)
		if rr.Code != 200 || fallo != nil || fin == nil {
			t.Fatalf("%s: /ask/stream %d, error %v, fin %s", nombre, rr.Code, fallo, fin)
		}
		if a, s := sinTraza(t, ask), sinTraza(t, fin); a != s {
			t.Errorf("%s: «fin» distinto de /ask\nask: %s\nfin: %s", nombre, a, s)
		}
		conTraza := strings.Contains(cuerpo, `"traza":true`)
		var m map[string]any
		json.Unmarshal(fin, &m)
		if _, ok := m["traza"]; ok != conTraza {
			t.Errorf("%s: traza en «fin» = %v, quería %v", nombre, ok, conTraza)
		}
		if cc := rr.Header().Get("Cache-Control"); (cc == "no-store") != conTraza {
			t.Errorf("%s: Cache-Control %q", nombre, cc)
		}
		if strings.HasPrefix(nombre, "V2") {
			if m["version"] != "v2" || m["plan"] == nil {
				t.Errorf("%s: el «fin» de un turno V2 trae plan y version: %s", nombre, fin)
			}
			// Un turno V2 no pasa por el modelo conversacional: nada que transmitir.
			if len(deltas) != 0 {
				t.Errorf("%s: un turno V2 no emite deltas: %q", nombre, deltas)
			}
		}
	}
}

// V1 en streaming con la V2 habilitada (pero no pedida) y modo traza: deltas del modelo y «fin» con traza.
func TestAskStreamV1TransmiteConV2HabilitadaYTraza(t *testing.T) {
	aislarMedios(t)
	c := v2.ConfigDefecto()
	c.Habilitada = true
	r := conAgente(t, c)
	r.LLM = modeloEnTrozos{"Cuesta ", "49 €."}
	h := Nuevo(r, 5*time.Second, Opciones{Traza: true})
	_, deltas, fin, fallo := eventosStream(t, h, `{"pregunta":"¿qué precio tiene el Pro?","version":"v1","traza":true}`)
	if fallo != nil || fin == nil {
		t.Fatalf("error %v, fin %s", fallo, fin)
	}
	if strings.Join(deltas, "|") != "Cuesta |49 €." {
		t.Errorf("deltas = %q", deltas)
	}
	var res struct {
		Respuesta string         `json:"respuesta"`
		Version   string         `json:"version"`
		Traza     map[string]any `json:"traza"`
	}
	json.Unmarshal(fin, &res)
	if res.Respuesta != "Cuesta 49 €." || res.Version != "" || res.Traza == nil {
		t.Fatalf("fin = %s", fin)
	}
}

// llmContador: responde siempre lo mismo y cuenta las llamadas (una reescritura del hilo sería una llamada más).
type llmContador struct {
	texto    string
	llamadas *int
}

func (l llmContador) Chat(context.Context, []llm.Mensaje) (string, error) {
	*l.llamadas++
	return l.texto, nil
}

// La página manda el hilo SIN la pregunta actual; la de AgenteS10 (y cualquier cliente viejo) la manda dentro. En
// /ask/stream, en V1 y en V2, las dos formas dan la misma respuesta y el mismo número de llamadas al modelo
// (sinPreguntaActual en V1, HiloSinEco en V2): el eco nunca dispara una reescritura.
func TestAskStreamHiloConOSinPreguntaActual(t *testing.T) {
	aislarMedios(t)
	c := v2.ConfigDefecto()
	c.Habilitada = true
	r := conAgente(t, c)
	llamadas := 0
	r.LLM = llmContador{"Cuesta 49 €.", &llamadas}
	h := Nuevo(r, 5*time.Second)
	for _, caso := range []struct{ pregunta, extra string }{
		{"¿qué precio tiene el Pro?", ``},
		{"¿qué precio tiene el Pro?", `,"version":"v1"`},
		{"¿Cómo registro un metrado?", `,"version":"v2"`},
	} {
		q := jsonTexto(caso.pregunta)
		llamadas = 0
		_, _, finSin, _ := eventosStream(t, h, `{"pregunta":`+q+caso.extra+`,"hilo":[]}`)
		nSin := llamadas
		llamadas = 0
		_, _, finCon, _ := eventosStream(t, h, `{"pregunta":`+q+caso.extra+`,"hilo":[{"rol":"usuario","texto":`+q+`}]}`)
		if llamadas != nSin {
			t.Errorf("%q%s: %d llamadas al modelo con el eco y %d sin él", caso.pregunta, caso.extra, llamadas, nSin)
		}
		if a, b := sinTraza(t, finSin), sinTraza(t, finCon); a != b {
			t.Errorf("%q%s: el eco de la pregunta en el hilo cambió la respuesta\nsin: %s\ncon: %s", caso.pregunta, caso.extra, a, b)
		}
	}
}
