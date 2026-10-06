// Package servidor expone el RAG por HTTP: POST /ask, POST /ask/stream y una
// página mínima.
package servidor

import (
	"bytes"
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
	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

//go:embed pagina.html
var pagina []byte

type peticion struct {
	Pregunta string      `json:"pregunta"`
	Source   string      `json:"source"`
	K        int         `json:"k"`
	Hilo     []rag.Turno `json:"hilo"`
	// Traza: «true» pide el modo traza. Se guarda en crudo para que cualquier
	// otro valor se ignore igual que antes de existir el campo.
	Traza json.RawMessage `json:"traza"`
	// V2 (docs/V2-RAG-PROCEDURAL.md): versión pedida («v1»|«v2», solo cuenta con
	// AGENT_V2_ENABLED=true), id de la conversación y memoria del procedimiento.
	// En crudo por lo mismo que «traza»: un valor raro se ignora, nunca da 400.
	Version      json.RawMessage `json:"version"`
	Conversacion json.RawMessage `json:"conversacion"`
	Memoria      json.RawMessage `json:"memoria"`
}

// Opciones del servidor.
type Opciones struct {
	// Traza habilita el modo traza (METRIN_TRAZA=1|true). Apagado, la
	// respuesta de /ask es la de siempre aunque la petición traiga «traza».
	Traza bool
	// V2: estado del interruptor de la V2 para /health (nil = sin V2: /health no
	// cambia).
	V2 any
}

// texto devuelve el valor si el JSON crudo es una cadena (o un número, para el
// id de conversación); cualquier otra cosa es "".
func texto(crudo json.RawMessage) string {
	var s string
	if json.Unmarshal(crudo, &s) == nil {
		return s
	}
	var n json.Number
	if json.Unmarshal(crudo, &n) == nil {
		return n.String()
	}
	return ""
}

// opcionesV2 pasa a rag.Opciones los campos de la V2 de la petición. Una
// memoria que no se puede leer se ignora (la V2 la reconstruye del hilo).
func (p peticion) opcionesV2(o *rag.Opciones) {
	o.Version = texto(p.Version)
	o.Conversacion = texto(p.Conversacion)
	if len(bytes.TrimSpace(p.Memoria)) > 0 {
		var m tipos.Memoria
		if json.Unmarshal(p.Memoria, &m) == nil {
			o.Memoria = &m
		}
	}
}

// pideTraza: solo el literal JSON true pide la traza.
func (p peticion) pideTraza() bool {
	return string(bytes.TrimSpace(p.Traza)) == "true"
}

// conTraza es la respuesta de /ask más la clave «traza».
type conTraza struct {
	rag.Respuesta
	Traza *traza.Traza `json:"traza"`
}

// Nuevo devuelve el manejador HTTP.
func Nuevo(r *rag.RAG, timeout time.Duration, opciones ...Opciones) http.Handler {
	var op Opciones
	if len(opciones) > 0 {
		op = opciones[0]
	}
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
		h := map[string]any{"ok": true, "trozos": r.Almacen.Contar(), "traza": op.Traza}
		if op.V2 != nil {
			h["v2"] = op.V2
		}
		escribirJSON(w, http.StatusOK, h)
	})
	// Fotos de las páginas (data/imagenes). FileServer limpia la ruta pedida
	// y http.Dir no abre nada fuera del directorio de datos.
	mux.Handle("GET /fotos/", http.StripPrefix("/fotos/", http.FileServer(http.Dir(rutaDatos()))))
	mux.HandleFunc("GET /fotos/manual/{slug}/{page}", fotoPaginaManual)
	mux.HandleFunc("POST /ask", func(w http.ResponseWriter, req *http.Request) {
		p, o, rec, ok := leerPeticion(w, req, op)
		if !ok {
			return
		}
		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()
		ctx = traza.ConContexto(ctx, rec)
		res, err := r.Preguntar(ctx, p.Pregunta, o)
		if err != nil {
			if rec != nil {
				w.Header().Set("Cache-Control", "no-store")
				escribirJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error(), "traza": rec.Cerrar()})
				return
			}
			escribirJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if rec != nil {
			w.Header().Set("Cache-Control", "no-store")
		}
		escribirJSON(w, http.StatusOK, terminar(res, rec))
	})
	// Igual que /ask, en NDJSON: {"tipo":"delta","texto":…} según el modelo
	// escribe y, al final, {"tipo":"fin","respuesta":{…}} con la respuesta
	// definitiva (la que manda, el mismo JSON que /ask: con «plan» en un turno
	// V2 y con «traza» si se pidió) o {"tipo":"error","error":…[,"traza":…]}.
	// Un turno V2 no emite deltas: su respuesta (plantillas) sale entera en «fin».
	mux.HandleFunc("POST /ask/stream", func(w http.ResponseWriter, req *http.Request) {
		p, o, rec, ok := leerPeticion(w, req, op)
		if !ok {
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
		if rec != nil {
			w.Header().Set("Cache-Control", "no-store")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		w.Header().Set("X-Accel-Buffering", "no") // que nginx no junte los trozos
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		rc := http.NewResponseController(w)
		evento := func(v any) {
			enc.Encode(v)
			rc.Flush()
		}
		o.Emitir = func(t string) { evento(map[string]string{"tipo": "delta", "texto": t}) }
		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()
		ctx = traza.ConContexto(ctx, rec)
		res, err := r.Preguntar(ctx, p.Pregunta, o)
		if err != nil {
			e := map[string]any{"tipo": "error", "error": err.Error()}
			if rec != nil {
				e["traza"] = rec.Cerrar()
			}
			evento(e)
			return
		}
		evento(map[string]any{"tipo": "fin", "respuesta": terminar(res, rec)})
	})
	return mux
}

// leerPeticion decodifica el cuerpo de /ask y /ask/stream y arma las opciones
// (V1, V2 y modo traza); si no sirve, responde 400 y devuelve ok=false.
// rec solo existe con el servicio habilitado (METRIN_TRAZA) y "traza": true
// en la petición; si no, es nil y nada de lo que sigue cambia.
func leerPeticion(w http.ResponseWriter, req *http.Request, op Opciones) (peticion, rag.Opciones, *traza.Recorder, bool) {
	var p peticion
	if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&p); err != nil {
		escribirJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return p, rag.Opciones{}, nil, false
	}
	p.Pregunta = strings.TrimSpace(p.Pregunta)
	if p.Pregunta == "" {
		escribirJSON(w, http.StatusBadRequest, map[string]string{"error": "falta «pregunta»"})
		return p, rag.Opciones{}, nil, false
	}
	o := rag.Opciones{K: p.K, Hilo: p.Hilo}
	if p.Source != "" {
		o.Filtro = map[string]string{"source": p.Source}
	}
	p.opcionesV2(&o)
	var rec *traza.Recorder
	if op.Traza && p.pideTraza() {
		rec = traza.Nuevo(p.Pregunta)
	}
	return p, o, rec, true
}

// terminar vincula las fotos de las fuentes y las coloca en la respuesta (V1;
// en un turno V2 las fotos van por paso en «plan» y las pinta la página). Con
// el modo traza anota la etapa «fotos» y devuelve la respuesta con «traza».
func terminar(res rag.Respuesta, rec *traza.Recorder) any {
	rec.Inicio(traza.EtapaFotos)
	antes := res.Respuesta
	if res.Version != tipos.V2 {
		vincularFotos(res.Fuentes)
		if !res.SinContexto {
			res.Respuesta = colocarFotos(res.Respuesta, res.Fuentes)
		}
	}
	if rec == nil {
		return res
	}
	if res.Version == tipos.V2 {
		rec.OmitirCon(traza.EtapaFotos, traza.EstadoOmitida, "turno V2: las fotos van por paso en «plan» (etapa fotos_paso)",
			traza.Datos{"fotos": 0, "fuentes_con_fotos": 0, "colocadas": false})
	} else {
		trazarFotos(rec, res, antes)
	}
	return conTraza{Respuesta: res, Traza: rec.Cerrar()}
}

// trazarFotos anota la etapa de fotos de pasos (modo traza).
func trazarFotos(rec *traza.Recorder, res rag.Respuesta, antes string) {
	fotos, conFotos := 0, 0
	for _, f := range res.Fuentes {
		fotos += len(f.Fotos)
		if len(f.Fotos) > 0 {
			conFotos++
		}
	}
	if len(res.Fuentes) == 0 {
		rec.OmitirCon(traza.EtapaFotos, traza.EstadoOmitida, "sin fuentes: no hay fotos que vincular",
			traza.Datos{"fotos": 0, "fuentes_con_fotos": 0, "colocadas": false})
		return
	}
	razon := "fotos de la página o de los pasos de cada fuente (tope " + strconv.Itoa(fotosPorFuente) + " por fuente)"
	if res.SinContexto {
		razon = "sin contexto: las fotos no se colocan en la respuesta"
	}
	rec.Fin(traza.EtapaFotos, traza.EstadoOK, traza.Datos{
		"fotos":             fotos,
		"fuentes_con_fotos": conFotos,
		"colocadas":         res.Respuesta != antes,
		"tope_por_fuente":   fotosPorFuente,
	})
	rec.Razon(traza.EtapaFotos, razon)
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
