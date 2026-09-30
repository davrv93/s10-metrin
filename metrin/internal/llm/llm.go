// Package llm habla con un LLM conversacional (Ollama o mlx_lm.server),
// sin streaming. Ambos implementan Chat y el RAG los usa indistintamente.
package llm

import (
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
	HTTP        *http.Client
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
		"options":  map[string]any{"temperature": o.Temperatura},
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
