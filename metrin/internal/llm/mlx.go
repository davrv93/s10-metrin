// Cliente para mlx_lm.server (API compatible con OpenAI, /v1/chat/completions).
// El adaptador LoRA vive en el servidor, así que el modelo es el que el
// servidor cargó al arrancar; Modelo solo se informa en la petición.
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

// MLX habla con un mlx_lm.server local (GPU del Mac) con el adaptador Metrín.
type MLX struct {
	URL         string
	Modelo      string
	Temperatura float64
	HTTP        *http.Client
}

func NuevoMLX(url, modelo string, temperatura float64, timeout time.Duration) *MLX {
	return &MLX{URL: strings.TrimRight(url, "/"), Modelo: modelo, Temperatura: temperatura, HTTP: &http.Client{Timeout: timeout}}
}

// Chat devuelve el contenido del mensaje del asistente.
func (m *MLX) Chat(ctx context.Context, msgs []Mensaje) (string, error) {
	cuerpo, _ := json.Marshal(map[string]any{
		"model":       m.Modelo,
		"messages":    msgs,
		"temperature": m.Temperatura,
		"max_tokens":  240,
		"stream":      false,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.URL+"/v1/chat/completions", bytes.NewReader(cuerpo))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("mlx chat (%s): %w", m.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("mlx chat: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	var r struct {
		Choices []struct {
			Message Mensaje `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("mlx chat: %w", err)
	}
	if len(r.Choices) == 0 {
		return "", fmt.Errorf("mlx chat: sin choices")
	}
	return strings.TrimSpace(r.Choices[0].Message.Content), nil
}
