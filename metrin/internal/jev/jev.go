// Package jev habla con la API de decisión JEV de jeva.cpp
// (llama-server, POST /v1/systemone). Responde preguntas de
// decisión —choice, score y noul— directamente desde los logits
// del modelo, sin generar tokens: el servidor arma la respuesta
// durante el prefill. El chat conversacional sigue en internal/llm.
package jev

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

// Pregunta es una decisión a evaluar sobre el estado compartido.
// Criterios depende del tipo: mapa nombre→descripción (choice;
// la descripción puede ser nil), lista de 2–10 niveles (score) o
// mapa "true"/"false" (noul; opcional).
type Pregunta struct {
	Tipo          string `json:"type"`
	Instrucciones any    `json:"instructions,omitempty"`
	Criterios     any    `json:"criteria,omitempty"`
}

// Solicitud es el cuerpo de /v1/systemone. Estado acepta string,
// objeto o array; cada pregunta recibe el mismo estado y solo sus
// propias instrucciones y criterios.
type Solicitud struct {
	Modelo    string              `json:"model,omitempty"`
	Estado    any                 `json:"state"`
	Preguntas map[string]Pregunta `json:"questions"`
}

// Respuesta es la respuesta a una pregunta. Los campos que no
// corresponden al tipo quedan en cero. Probabilidades usa los
// nombres de opción (choice) o índices "0"…"n-1" (score); Leyenda
// conserva las descripciones originales del score; Noul es la
// probabilidad de true. Confianza es 1−H(p)/log(N): concentración
// de la distribución, no corrección calibrada.
type Respuesta struct {
	Tipo           string             `json:"type"`
	Choice         string             `json:"choice,omitempty"`
	Probabilidades map[string]float64 `json:"probabilities,omitempty"`
	Confianza      float64            `json:"confidence,omitempty"`
	Score          float64            `json:"score,omitempty"`
	Leyenda        map[string]any     `json:"legend,omitempty"`
	Noul           float64            `json:"noul,omitempty"`
}

type uso struct {
	TokensEntrada int `json:"input_tokens"`
	TokensSalida  int `json:"output_tokens"`
}

type respuestaSolicitud struct {
	Modelo     string              `json:"model"`
	Respuestas map[string]Respuesta `json:"answers"`
	Uso        uso                  `json:"usage"`
}

// Cliente habla con un llama-server de jeva.cpp. En modo un solo
// modelo, Modelo puede ir vacío; con varios (router) selecciona.
// Trazas, si no es nil, graba cada decisión en el JSONL del
// visor de trazas (ver trazas.go).
type Cliente struct {
	URL    string
	Modelo string
	APIKey string
	HTTP   *http.Client
	Trazas *RegistroTrazas
}

// Nuevo crea un cliente para un llama-server con la API JEV.
func Nuevo(url, modelo string, timeout time.Duration) *Cliente {
	return &Cliente{
		URL:    strings.TrimRight(url, "/"),
		Modelo: modelo,
		HTTP:   &http.Client{Timeout: timeout},
	}
}

// Decidir envía todas las preguntas en una sola llamada y devuelve
// las respuestas por ID. La API procesa las preguntas en las
// ranuras del servidor; una que falla falla la petición entera.
func (c *Cliente) Decidir(ctx context.Context, estado any, preguntas map[string]Pregunta) (map[string]Respuesta, error) {
	if len(preguntas) == 0 {
		return nil, fmt.Errorf("jev: sin preguntas")
	}
	cuerpo, err := json.Marshal(Solicitud{Modelo: c.Modelo, Estado: estado, Preguntas: preguntas})
	if err != nil {
		return nil, fmt.Errorf("jev: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+"/v1/systemone", bytes.NewReader(cuerpo))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	t0 := time.Now()
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jev decisión (%s): %w", c.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("jev decisión: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	var r respuestaSolicitud
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("jev decisión: %w", err)
	}
	if c.Trazas != nil {
		modelo := r.Modelo
		if modelo == "" {
			modelo = c.Modelo
		}
		// La decisión ya está tomada: si falla el registro, la
		// traza se pierde pero la petición sigue siendo válida.
		_ = c.Trazas.guardar(nuevaTraza(modelo, c.Trazas.Origen, estado, preguntas, r.Respuestas,
			time.Since(t0).Milliseconds(), r.Uso.TokensEntrada))
	}
	return r.Respuestas, nil
}

const idUna = "q"

// Choice evalúa una decisión entre opciones nombradas con su
// descripción (puede ser nil). Devuelve la opción de mayor
// probabilidad; los empates se resuelven por orden de la petición.
func (c *Cliente) Choice(ctx context.Context, estado any, instrucciones any, criterios map[string]any) (Respuesta, error) {
	rs, err := c.Decidir(ctx, estado, map[string]Pregunta{
		idUna: {Tipo: "choice", Instrucciones: instrucciones, Criterios: criterios},
	})
	if err != nil {
		return Respuesta{}, err
	}
	return rs[idUna], nil
}

// Score evalúa el estado en una escala de 2–10 niveles y devuelve
// el índice esperado (suma i·p[i]) con su distribución y leyenda.
func (c *Cliente) Score(ctx context.Context, estado any, instrucciones any, criterios []any) (Respuesta, error) {
	rs, err := c.Decidir(ctx, estado, map[string]Pregunta{
		idUna: {Tipo: "score", Instrucciones: instrucciones, Criterios: criterios},
	})
	if err != nil {
		return Respuesta{}, err
	}
	return rs[idUna], nil
}

// Noul evalúa una proposición binaria y devuelve la probabilidad
// de true. descripciones puede ser nil o llevar solo "true"/"false".
func (c *Cliente) Noul(ctx context.Context, estado any, instrucciones any, descripciones map[string]any) (Respuesta, error) {
	rs, err := c.Decidir(ctx, estado, map[string]Pregunta{
		idUna: {Tipo: "noul", Instrucciones: instrucciones, Criterios: descripciones},
	})
	if err != nil {
		return Respuesta{}, err
	}
	return rs[idUna], nil
}
