// Package llm habla con un LLM conversacional (Ollama o mlx_lm.server).
// Ambos implementan Chat y el RAG los usa indistintamente; Ollama además
// transmite la respuesta a trozos (ChatStream).
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Mensaje struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Ollama struct {
	URL         string
	Modelo      string
	Temperatura float64
	Hilos       int // num_thread; 0 = lo decide Ollama (en CPU suele usar uno solo)
	MaxTokens   int // num_predict; 0 = sin tope
	HTTP        *http.Client
}

func (o *Ollama) opciones() map[string]any {
	op := map[string]any{"temperature": o.Temperatura}
	if o.Hilos > 0 {
		op["num_thread"] = o.Hilos
	}
	if o.MaxTokens > 0 {
		op["num_predict"] = o.MaxTokens
	}
	return op
}

func NuevoOllama(url, modelo string, temperatura float64, timeout time.Duration) *Ollama {
	return &Ollama{URL: strings.TrimRight(url, "/"), Modelo: modelo, Temperatura: temperatura, HTTP: &http.Client{Timeout: timeout}}
}

// Chat devuelve el contenido del mensaje del asistente.
func (o *Ollama) Chat(ctx context.Context, msgs []Mensaje) (string, error) {
	cuerpo, _ := json.Marshal(map[string]any{
		"model":    o.Modelo,
		"messages": msgs,
		"stream":   false,
		"options":  o.opciones(),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.URL+"/api/chat", bytes.NewReader(cuerpo))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama chat (%s): %w", o.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("ollama chat: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	var r struct {
		Message Mensaje `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("ollama chat: %w", err)
	}
	return strings.TrimSpace(r.Message.Content), nil
}

// ChatStream es Chat con streaming: llama a emitir con cada trozo de texto
// según llega y devuelve el mensaje completo. emitir corre en esta misma
// goroutine, en orden.
func (o *Ollama) ChatStream(ctx context.Context, msgs []Mensaje, emitir func(string)) (string, error) {
	cuerpo, _ := json.Marshal(map[string]any{
		"model":    o.Modelo,
		"messages": msgs,
		"stream":   true,
		"options":  o.opciones(),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.URL+"/api/chat", bytes.NewReader(cuerpo))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama chat (%s): %w", o.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("ollama chat: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	var todo strings.Builder
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		var r struct {
			Message Mensaje `json:"message"`
			Done    bool    `json:"done"`
			Error   string  `json:"error"`
		}
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return "", fmt.Errorf("ollama chat: %w", err)
		}
		if r.Error != "" {
			return "", fmt.Errorf("ollama chat: %s", r.Error)
		}
		if r.Message.Content != "" {
			todo.WriteString(r.Message.Content)
			emitir(r.Message.Content)
		}
		if r.Done {
			return strings.TrimSpace(todo.String()), nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("ollama chat: %w", err)
	}
	return "", fmt.Errorf("ollama chat: el stream terminó sin done")
}
