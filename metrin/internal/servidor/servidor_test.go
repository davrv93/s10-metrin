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

	"rag-go/internal/almacen"
	"rag-go/internal/aprender"
	"rag-go/internal/llm"
	"rag-go/internal/rag"
)

type unEje struct{}

func (unEje) Nombre() string { return "uno" }
func (unEje) Embeber(context.Context, string) ([]float32, error) {
	return []float32{1, 0}, nil
}

type llmNo struct{}

func (llmNo) Chat(context.Context, []llm.Mensaje) (string, error) {
	return "No tengo contexto suficiente para responder a esa pregunta.", nil
}

func post(t *testing.T, url, token, cuerpo string, v any) int {
	req, _ := http.NewRequest("POST", url, strings.NewReader(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if v != nil {
		json.NewDecoder(res.Body).Decode(v)
	}
	return res.StatusCode
}

func TestFeedbackAdminYEventos(t *testing.T) {
	a, _ := almacen.Abrir("", unEje{})
	d, _ := aprender.Abrir(t.TempDir())
	r := &rag.RAG{Almacen: a, LLM: llmNo{}, MaxDistancia: 0.8}
	srv := httptest.NewServer(Nuevo(r, Opciones{Timeout: 5 * time.Second, Diario: d, TokenAdmin: "secreto"}))
	defer srv.Close()

	// SSE primero, para ver llegar la interacción.
	req, _ := http.NewRequest("GET", srv.URL+"/admin/api/eventos?token=secreto", nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil || resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatalf("sse: %v %v", err, resp)
	}
	lineas := bufio.NewScanner(resp.Body)
	esperar := func(evento string) {
		t.Helper()
		for lineas.Scan() {
			if lineas.Text() == "event: "+evento {
				return
			}
		}
		t.Fatalf("no llegó el evento %s", evento)
	}
	esperar("metricas") // estado inicial al conectar

	var res rag.Respuesta
	if c := post(t, srv.URL+"/ask", "", `{"pregunta":"¿cómo anulo la guía?"}`, &res); c != 200 || res.ID == "" || !res.SinContexto {
		t.Fatalf("/ask %d %+v", c, res)
	}
	esperar("interaccion")

	if c := post(t, srv.URL+"/feedback", "", `{"id":"`+res.ID+`","voto":"negativo","comentario":"no sirve"}`, nil); c != 200 {
		t.Fatalf("/feedback %d", c)
	}
	esperar("feedback")
	if c := post(t, srv.URL+"/feedback", "", `{"id":"x","voto":1}`, nil); c != 404 {
		t.Fatalf("feedback a id desconocido: %d", c)
	}

	// Admin: sin token no; con token sí.
	if c := post(t, srv.URL+"/admin/api/aprender", "", `{}`, nil); c != 401 {
		t.Fatalf("sin token: %d", c)
	}
	var p aprender.PendienteItem
	c := post(t, srv.URL+"/admin/api/aprender", "secreto",
		`{"clave":"como anulo la guia","respuesta":"Almacén → clic derecho sobre la guía → Anular."}`, &p)
	if c != 200 || p.Estado != aprender.Aprendido {
		t.Fatalf("aprender %d %+v", c, p)
	}
	esperar("aprendido")

	req, _ = http.NewRequest("GET", srv.URL+"/admin/api/metricas?dias=7", nil)
	req.Header.Set("X-Admin-Token", "secreto")
	mr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var m aprender.Metricas
	json.NewDecoder(mr.Body).Decode(&m)
	mr.Body.Close()
	if m.Total != 1 || m.Negativos != 1 || m.Cola.Aprendidas != 1 {
		t.Fatalf("métricas %+v", m)
	}
}

func TestAdminDesactivadaSinToken(t *testing.T) {
	a, _ := almacen.Abrir("", unEje{})
	d, _ := aprender.Abrir(t.TempDir())
	srv := httptest.NewServer(Nuevo(&rag.RAG{Almacen: a, LLM: llmNo{}}, Opciones{Diario: d}))
	defer srv.Close()
	res, _ := http.Get(srv.URL + "/admin/api/metricas")
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("sin RAG_ADMIN_TOKEN la API admin debe estar apagada: %d", res.StatusCode)
	}
}
