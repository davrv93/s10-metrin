package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChatStreamEmiteTrozosYDevuelveElTexto(t *testing.T) {
	var pedido map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&pedido)
		for _, trozo := range []string{"Hola", ", ", "mundo"} {
			fmt.Fprintf(w, `{"message":{"role":"assistant","content":%q},"done":false}`+"\n", trozo)
		}
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":""},"done":true}`)
	}))
	defer srv.Close()

	o := NuevoOllama(srv.URL, "qwen2.5:0.5b", 0.2, time.Second)
	o.Hilos, o.MaxTokens = 2, 300
	var trozos []string
	texto, err := o.ChatStream(context.Background(), []Mensaje{{Role: "user", Content: "hola"}}, func(s string) { trozos = append(trozos, s) })
	if err != nil {
		t.Fatal(err)
	}
	if texto != "Hola, mundo" || strings.Join(trozos, "|") != "Hola|, |mundo" {
		t.Fatalf("texto=%q trozos=%q", texto, trozos)
	}
	if pedido["stream"] != true {
		t.Errorf("stream = %v, quería true", pedido["stream"])
	}
	op, _ := pedido["options"].(map[string]any)
	if op["num_thread"] != float64(2) || op["num_predict"] != float64(300) {
		t.Errorf("options = %v, quería num_thread 2 y num_predict 300", op)
	}
}

func TestChatStreamPropagaElErrorDeOllama(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, `{"error":"model not found"}`)
	}))
	defer srv.Close()
	_, err := NuevoOllama(srv.URL, "x", 0, time.Second).ChatStream(context.Background(), nil, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "model not found") {
		t.Fatalf("err = %v", err)
	}
}

func TestChatSinHilosNiTopeNoMandaEsasOpciones(t *testing.T) {
	var pedido map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&pedido)
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":"ok"}}`)
	}))
	defer srv.Close()
	if _, err := NuevoOllama(srv.URL, "x", 0.2, time.Second).Chat(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	op, _ := pedido["options"].(map[string]any)
	if _, hay := op["num_thread"]; hay {
		t.Errorf("options = %v: num_thread no debía ir", op)
	}
	if _, hay := op["num_predict"]; hay {
		t.Errorf("options = %v: num_predict no debía ir", op)
	}
}
