// Package embed produce los vectores del índice. Dos proveedores:
//
//   - estatico: model2vec (potion-multilingual recortado a español, int8) en
//     Go puro, sin GPU ni red. Es el formato PJGE del Analista (modelo.go).
//   - ollama: POST /api/embed contra un Ollama local (nomic-embed-text).
//
// Los dos devuelven vectores de norma 1, que es lo que espera chromem-go.
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// Embebedor convierte texto en un vector de norma 1.
type Embebedor interface {
	Embeber(ctx context.Context, texto string) ([]float32, error)
	// Nombre identifica proveedor+modelo; va en el nombre de la colección
	// para no mezclar vectores de dimensiones o espacios distintos.
	Nombre() string
}

// Estatico envuelve un Modelo PJGE.
type Estatico struct{ M *Modelo }

// NuevoEstatico carga el fichero .pjge.
func NuevoEstatico(ruta string) (*Estatico, error) {
	m, err := CargarArchivo(ruta)
	if err != nil {
		return nil, fmt.Errorf("modelo estático %s: %w", ruta, err)
	}
	return &Estatico{M: m}, nil
}

func (e *Estatico) Nombre() string { return "estatico-" + e.M.Huella()[:12] }

func (e *Estatico) Embeber(_ context.Context, texto string) ([]float32, error) {
	v := e.M.Embeber(texto)
	if esCero(v) {
		return nil, ErrSinPiezas
	}
	return v, nil
}

// ErrSinPiezas: el texto no tiene ninguna pieza conocida por el modelo
// (vector nulo, que no se puede normalizar).
var ErrSinPiezas = errors.New("el texto no tiene piezas conocidas por el modelo")

func esCero(v []float32) bool {
	for _, x := range v {
		if x != 0 {
			return false
		}
	}
	return true
}

// Ollama llama a /api/embed.
type Ollama struct {
	URL    string // p. ej. http://localhost:11434
	Modelo string // p. ej. nomic-embed-text
	HTTP   *http.Client
}

func NuevoOllama(url, modelo string) *Ollama {
	return &Ollama{URL: strings.TrimRight(url, "/"), Modelo: modelo, HTTP: &http.Client{Timeout: 120 * time.Second}}
}

func (o *Ollama) Nombre() string {
	return "ollama-" + strings.NewReplacer(":", "-", "/", "-").Replace(o.Modelo)
}

func (o *Ollama) Embeber(ctx context.Context, texto string) ([]float32, error) {
	cuerpo, _ := json.Marshal(map[string]any{"model": o.Modelo, "input": texto})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.URL+"/api/embed", bytes.NewReader(cuerpo))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama embed: HTTP %d", resp.StatusCode)
	}
	var r struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}
	if len(r.Embeddings) == 0 {
		return nil, errors.New("ollama embed: respuesta sin vectores")
	}
	return Normalizar(r.Embeddings[0]), nil
}

// Normalizar devuelve v con norma 1 (o v si es nulo).
func Normalizar(v []float32) []float32 {
	var s float64
	for _, x := range v {
		s += float64(x) * float64(x)
	}
	if s == 0 {
		return v
	}
	n := math.Sqrt(s)
	out := make([]float32, len(v))
	for i, x := range v {
		out[i] = float32(float64(x) / n)
	}
	return out
}
