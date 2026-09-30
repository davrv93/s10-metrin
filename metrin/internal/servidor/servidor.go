// Package servidor expone el RAG por HTTP: POST /ask y una página mínima.
package servidor

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"rag-go/internal/rag"
)

//go:embed pagina.html
var pagina []byte

type peticion struct {
	Pregunta string      `json:"pregunta"`
	Source   string      `json:"source"`
	K        int         `json:"k"`
	Hilo     []rag.Turno `json:"hilo"`
}

// Nuevo devuelve el manejador HTTP.
func Nuevo(r *rag.RAG, timeout time.Duration) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(pagina)
	})
	// Catálogo de módulos y preguntas sugeridas (lo arma s10-conocimiento/herramientas/catalogo_metrin.py
	// desde Cortex). Se lee del disco en cada petición: se actualiza sin reconstruir la imagen.
	mux.HandleFunc("GET /catalogo", func(w http.ResponseWriter, _ *http.Request) {
		for _, ruta := range rutasCatalogo() {
			if b, err := os.ReadFile(ruta); err == nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set("Cache-Control", "no-cache")
				w.Write(b)
				return
			}
		}
		escribirJSON(w, http.StatusNotFound, map[string]string{"error": "catálogo no disponible"})
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]any{"ok": true, "trozos": r.Almacen.Contar()})
	})
	mux.HandleFunc("POST /ask", func(w http.ResponseWriter, req *http.Request) {
		var p peticion
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&p); err != nil {
			escribirJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
			return
		}
		p.Pregunta = strings.TrimSpace(p.Pregunta)
		if p.Pregunta == "" {
			escribirJSON(w, http.StatusBadRequest, map[string]string{"error": "falta «pregunta»"})
			return
		}
		o := rag.Opciones{K: p.K, Hilo: p.Hilo}
		if p.Source != "" {
			o.Filtro = map[string]string{"source": p.Source}
		}
		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()
		res, err := r.Preguntar(ctx, p.Pregunta, o)
		if err != nil {
			escribirJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		escribirJSON(w, http.StatusOK, res)
	})
	return mux
}

func rutasCatalogo() []string {
	if r := os.Getenv("RAG_CATALOGO"); r != "" {
		return []string{r}
	}
	return []string{"/kb/catalogo_metrin.json", "../kb/catalogo_metrin.json", "kb/catalogo_metrin.json"}
}

func escribirJSON(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	json.NewEncoder(w).Encode(v)
}
