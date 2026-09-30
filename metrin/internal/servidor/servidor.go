// Package servidor expone el RAG por HTTP: POST /ask y una página mínima.
package servidor

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	// Fotos de las páginas (data/imagenes). FileServer limpia la ruta pedida
	// y http.Dir no abre nada fuera del directorio de datos.
	mux.Handle("GET /fotos/", http.StripPrefix("/fotos/", http.FileServer(http.Dir(rutaDatos()))))
	mux.HandleFunc("GET /fotos/manual/{slug}/{page}", fotoPaginaManual)
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
		vincularFotos(res.Fuentes)
		if !res.SinContexto {
			res.Respuesta = colocarFotos(res.Respuesta, res.Fuentes)
		}
		escribirJSON(w, http.StatusOK, res)
	})
	return mux
}

// fotosPorFuente: tope de fotos que se vinculan a cada fuente.
const fotosPorFuente = 8

// vincularFotos agrega a cada fuente las fotos de su página de origen
// (data/imagenes.jsonl). Se lee en cada pregunta, igual que el catálogo:
// imágenes nuevas aparecen sin reconstruir la imagen.
func vincularFotos(fuentes []rag.Fuente) {
	mapa := mapaImagenes()
	for i := range fuentes {
		if fotos := fotosDePasos(fuentes[i].Pasos); len(fotos) > 0 {
			// sección del manual: solo sus capturas, no las de toda la página
			fuentes[i].Fotos = fotos
			continue
		}
		f := mapa[fuentes[i].URL]
		if len(f) == 0 && fuentes[i].Pagina > 0 && strings.HasPrefix(fuentes[i].URL, "manual://") {
			slug := slugDocumento(fuentes[i].Documento)
			if _, err := pdfManual(slug); err == nil {
				f = []string{"manual/" + slug + "/" + strconv.Itoa(fuentes[i].Pagina)}
			}
		}
		if len(f) == 0 {
			continue
		}
		if len(f) > fotosPorFuente {
			f = f[:fotosPorFuente]
		}
		fuentes[i].Fotos = append([]string(nil), f...)
	}
}

func fotosDePasos(pasos []rag.Paso) []string {
	var out []string
	for _, p := range pasos {
		out = append(out, p.Fotos...)
	}
	return out
}

func slugDocumento(nombre string) string {
	nombre = strings.TrimSuffix(filepath.Base(nombre), filepath.Ext(nombre))
	nombre = strings.ToLower(nombre)
	nombre = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n", " ", "-", "_", "-").Replace(nombre)
	for strings.Contains(nombre, "--") {
		nombre = strings.ReplaceAll(nombre, "--", "-")
	}
	return strings.Trim(nombre, "-.")
}

func pdfManual(slug string) (string, error) {
	if slug == "" {
		return "", os.ErrNotExist
	}
	for _, r := range slug {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return "", os.ErrNotExist
		}
	}
	pdf := filepath.Join(rutaDatos(), "pdf", slug+".pdf")
	st, err := os.Stat(pdf)
	if err != nil || st.IsDir() {
		return "", os.ErrNotExist
	}
	return pdf, nil
}

var renderPDFMu sync.Mutex

func fotoPaginaManual(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	page, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || page < 1 || page > 1000 {
		http.NotFound(w, r)
		return
	}
	pdf, err := pdfManual(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	cache := os.Getenv("RAG_PREVIEWS")
	if cache == "" {
		cache = filepath.Join(os.Getenv("RAG_DATOS"), "previas")
	}
	if cache == "previas" || cache == "" {
		cache = filepath.Join(os.TempDir(), "metrin-previas")
	}
	target := filepath.Join(cache, slug+"-p"+strconv.Itoa(page)+".jpg")
	if _, err := os.Stat(target); err != nil {
		renderPDFMu.Lock()
		defer renderPDFMu.Unlock()
		if _, err := os.Stat(target); err != nil {
			if err := os.MkdirAll(cache, 0o755); err != nil {
				http.Error(w, "No se pudo preparar la vista previa", http.StatusInternalServerError)
				return
			}
			tmp, err := os.MkdirTemp(cache, "pagina-")
			if err != nil {
				http.Error(w, "No se pudo preparar la vista previa", http.StatusInternalServerError)
				return
			}
			defer os.RemoveAll(tmp)
			prefix := filepath.Join(tmp, "pagina")
			ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "pdftoppm", "-f", strconv.Itoa(page), "-l", strconv.Itoa(page), "-scale-to", "1000", "-jpeg", "-jpegopt", "quality=72", "-singlefile", pdf, prefix)
			if err := cmd.Run(); err != nil {
				http.NotFound(w, r)
				return
			}
			if err := os.Rename(prefix+".jpg", target); err != nil {
				http.Error(w, "No se pudo guardar la vista previa", http.StatusInternalServerError)
				return
			}
		}
	}
	f, err := os.Open(target)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = f.WriteTo(w)
}

// mapaImagenes: URL de página -> rutas relativas de sus fotos.
func mapaImagenes() map[string][]string {
	for _, ruta := range rutasImagenes() {
		b, err := os.ReadFile(ruta)
		if err != nil {
			continue
		}
		mapa := map[string][]string{}
		for _, linea := range strings.Split(string(b), "\n") {
			var im struct {
				Pagina  string `json:"pagina"`
				Archivo string `json:"archivo"`
			}
			if json.Unmarshal([]byte(linea), &im) != nil || im.Pagina == "" || im.Archivo == "" {
				continue
			}
			mapa[im.Pagina] = append(mapa[im.Pagina], im.Archivo)
		}
		return mapa
	}
	return nil
}

func rutasImagenes() []string {
	if r := os.Getenv("RAG_IMAGENES"); r != "" {
		return []string{r}
	}
	return []string{"/datos/imagenes.jsonl", "../data/imagenes.jsonl", "data/imagenes.jsonl"}
}

// rutaDatos: directorio raíz que se sirve bajo /fotos/ (data/).
func rutaDatos() string {
	if r := os.Getenv("RAG_MEDIA_DATOS"); r != "" {
		return r
	}
	for _, c := range []string{"/datos", "../data", "data"} {
		if st, err := os.Stat(filepath.Join(c, "imagenes.jsonl")); err == nil && !st.IsDir() {
			return c
		}
	}
	return "/datos"
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
