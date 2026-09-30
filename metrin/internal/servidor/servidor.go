// Package servidor expone el RAG por HTTP: POST /ask, POST /feedback, la
// página del chat y la API de administración (/admin/api/…, con token) para
// métricas, cola de aprendizaje y eventos en vivo (SSE).
package servidor

import (
	"context"
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"rag-go/internal/aprender"
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

// Opciones del servidor más allá del RAG.
type Opciones struct {
	Timeout time.Duration
	Diario  *aprender.Diario // nil = sin registro, feedback ni admin
	// TokenAdmin protege /admin/api/. Vacío = la API admin responde 503.
	TokenAdmin string
}

// Nuevo devuelve el manejador HTTP.
func Nuevo(r *rag.RAG, o Opciones) http.Handler {
	timeout := o.Timeout
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
		op := rag.Opciones{K: p.K, Hilo: p.Hilo}
		if p.Source != "" {
			op.Filtro = map[string]string{"source": p.Source}
		}
		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()
		res, err := r.Preguntar(ctx, p.Pregunta, op)
		if err != nil {
			escribirJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if o.Diario != nil {
			if id, err := o.Diario.Registrar(res); err == nil {
				res.ID = id
			} else {
				log.Printf("registrar interacción: %v", err)
			}
		}
		escribirJSON(w, http.StatusOK, res)
	})
	if o.Diario != nil {
		montarAprendizaje(mux, r, o)
	}
	return mux
}

type votoPeticion struct {
	ID         string `json:"id"`
	Voto       any    `json:"voto"` // 1 / -1 o "positivo" / "negativo"
	Comentario string `json:"comentario"`
}

func leerVoto(v any) int {
	switch x := v.(type) {
	case float64:
		if x > 0 {
			return 1
		} else if x < 0 {
			return -1
		}
	case string:
		switch strings.ToLower(strings.TrimSpace(x)) {
		case "positivo", "+1", "1", "up", "si", "sí":
			return 1
		case "negativo", "-1", "down", "no":
			return -1
		}
	case bool:
		if x {
			return 1
		}
		return -1
	}
	return 0
}

func montarAprendizaje(mux *http.ServeMux, r *rag.RAG, o Opciones) {
	d := o.Diario
	// Feedback del chat: público como /ask (mismo alcance).
	mux.HandleFunc("POST /feedback", func(w http.ResponseWriter, req *http.Request) {
		var p votoPeticion
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&p); err != nil {
			escribirJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
			return
		}
		voto := leerVoto(p.Voto)
		if p.ID == "" || voto == 0 {
			escribirJSON(w, http.StatusBadRequest, map[string]string{"error": "faltan «id» y «voto» (1 o -1)"})
			return
		}
		if err := d.Votar(p.ID, voto, p.Comentario); err != nil {
			estado := http.StatusInternalServerError
			if errors.Is(err, aprender.ErrNoEncontrado) {
				estado = http.StatusNotFound
			}
			escribirJSON(w, estado, map[string]string{"error": err.Error()})
			return
		}
		escribirJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	admin := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			if o.TokenAdmin == "" {
				escribirJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "API admin desactivada: define RAG_ADMIN_TOKEN"})
				return
			}
			if !tokenValido(req, o.TokenAdmin) {
				escribirJSON(w, http.StatusUnauthorized, map[string]string{"error": "token inválido"})
				return
			}
			h(w, req)
		}
	}
	mux.HandleFunc("GET /admin/api/metricas", admin(func(w http.ResponseWriter, req *http.Request) {
		dias, _ := strconv.Atoi(req.URL.Query().Get("dias"))
		escribirJSON(w, http.StatusOK, d.Metricas(dias))
	}))
	mux.HandleFunc("GET /admin/api/pendientes", admin(func(w http.ResponseWriter, req *http.Request) {
		escribirJSON(w, http.StatusOK, d.Pendientes(req.URL.Query().Get("estado")))
	}))
	mux.HandleFunc("GET /admin/api/interacciones", admin(func(w http.ResponseWriter, req *http.Request) {
		n, _ := strconv.Atoi(req.URL.Query().Get("limite"))
		if n <= 0 || n > 500 {
			n = 50
		}
		escribirJSON(w, http.StatusOK, d.Recientes(n))
	}))
	mux.HandleFunc("POST /admin/api/aprender", admin(func(w http.ResponseWriter, req *http.Request) {
		var l aprender.Leccion
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&l); err != nil {
			escribirJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
			return
		}
		p, err := d.Aprender(req.Context(), r.Almacen, l)
		if err != nil {
			estado := http.StatusBadRequest
			if errors.Is(err, aprender.ErrNoEncontrado) {
				estado = http.StatusNotFound
			}
			escribirJSON(w, estado, map[string]string{"error": err.Error()})
			return
		}
		escribirJSON(w, http.StatusOK, p)
	}))
	mux.HandleFunc("POST /admin/api/descartar", admin(func(w http.ResponseWriter, req *http.Request) {
		var p struct {
			Clave string `json:"clave"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&p); err != nil || p.Clave == "" {
			escribirJSON(w, http.StatusBadRequest, map[string]string{"error": "falta «clave»"})
			return
		}
		if err := d.Descartar(p.Clave); err != nil {
			escribirJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		escribirJSON(w, http.StatusOK, map[string]any{"ok": true})
	}))
	// Repaso a demanda: responde al instante y el avance llega por SSE.
	mux.HandleFunc("POST /admin/api/repasar", admin(func(w http.ResponseWriter, _ *http.Request) {
		if c, _ := d.EstadoRepaso(); c != nil {
			escribirJSON(w, http.StatusConflict, map[string]string{"error": aprender.ErrRepasoEnCurso.Error()})
			return
		}
		go func() {
			if _, err := d.Repasar(context.Background(), r, "admin"); err != nil {
				log.Printf("repaso: %v", err)
			}
		}()
		escribirJSON(w, http.StatusAccepted, map[string]any{"ok": true, "eventos": "/admin/api/eventos"})
	}))
	mux.HandleFunc("GET /admin/api/aprendidos.jsonl", admin(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
		http.ServeFile(w, req, d.RutaAprendidos())
	}))
	mux.HandleFunc("GET /admin/api/eventos", admin(func(w http.ResponseWriter, req *http.Request) {
		sse(w, req, d)
	}))
}

// sse emite los eventos del bus: `event: <tipo>` + `data: <json>`. Al
// conectar manda las métricas actuales; cada 20 s un comentario de latido.
func sse(w http.ResponseWriter, req *http.Request, d *aprender.Diario) {
	fl, ok := w.(http.Flusher)
	if !ok {
		escribirJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming no soportado"})
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/event-stream; charset=utf-8")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	eventos, soltar := d.Bus.Suscribir()
	defer soltar()
	enviar := func(tipo string, datos any) bool {
		b, err := json.Marshal(datos)
		if err != nil {
			return true
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", tipo, b); err != nil {
			return false
		}
		fl.Flush()
		return true
	}
	if !enviar("metricas", d.Metricas(14)) {
		return
	}
	latido := time.NewTicker(20 * time.Second)
	defer latido.Stop()
	for {
		select {
		case <-req.Context().Done():
			return
		case <-latido.C:
			if _, err := fmt.Fprint(w, ": latido\n\n"); err != nil {
				return
			}
			fl.Flush()
		case ev, ok := <-eventos:
			if !ok || !enviar(ev.Tipo, ev) {
				return
			}
			// Tras cada cambio, métricas frescas para refrescar los KPI.
			if ev.Tipo != "repaso_avance" && !enviar("metricas", d.Metricas(14)) {
				return
			}
		}
	}
}

func tokenValido(req *http.Request, token string) bool {
	dado := strings.TrimSpace(strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer "))
	if dado == "" {
		dado = req.Header.Get("X-Admin-Token")
	}
	if dado == "" {
		dado = req.URL.Query().Get("token") // EventSource no manda cabeceras
	}
	return dado != "" && subtle.ConstantTimeCompare([]byte(dado), []byte(token)) == 1
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
