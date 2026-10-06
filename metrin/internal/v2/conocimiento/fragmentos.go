package conocimiento

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// PasoFuente es un paso de una sección del manual con las capturas que la propia sección le asocia.
type PasoFuente struct {
	Texto string   `json:"texto"`
	Fotos []string `json:"fotos"`
}

// Fragmento de kb/fragmentos*.jsonl. Su clave es (ID, Manual): el ID solo no es único.
type Fragmento struct {
	ID        string       `json:"id"`
	Documento string       `json:"documento,omitempty"`
	Manual    string       `json:"manual"`
	Titulo    string       `json:"titulo,omitempty"`
	Seccion   string       `json:"seccion,omitempty"`
	Pagina    int          `json:"pagina,omitempty"`
	Fuente    string       `json:"fuente,omitempty"`
	Confianza string       `json:"confianza,omitempty"`
	Texto     string       `json:"texto"`
	Tipo      string       `json:"tipo,omitempty"`
	Imagen    string       `json:"imagen,omitempty"`
	URLImagen string       `json:"url_imagen,omitempty"`
	Pasos     []PasoFuente `json:"pasos,omitempty"`
	Archivo   string       `json:"-"`
	Linea     int          `json:"-"`
}

// Clave de un fragmento: el par que sí es único.
type Clave struct {
	ID     string
	Manual string
}

// TextoPlegado: título, sección, texto (o OCR) y pasos del fragmento, plegados (sin tildes ni
// mayúsculas). No se guarda: el índice es de solo lectura y se comparte entre peticiones.
func (f *Fragmento) TextoPlegado() string {
	var b strings.Builder
	b.WriteString(f.Titulo)
	b.WriteByte('\n')
	b.WriteString(f.Seccion)
	b.WriteByte('\n')
	b.WriteString(f.Texto)
	for _, p := range f.Pasos {
		b.WriteByte('\n')
		b.WriteString(p.Texto)
	}
	return plegar(b.String())
}

// EsImagen: fragmento de una captura (su texto es el OCR).
func (f *Fragmento) EsImagen() bool { return f.Tipo == "imagen" || f.Imagen != "" }

// Fragmentos es el índice de kb/fragmentos*.jsonl por (id, manual), por id y por ruta de imagen.
type Fragmentos struct {
	porClave  map[Clave]*Fragmento
	porID     map[string][]*Fragmento
	porImagen map[string]*Fragmento
	todos     []*Fragmento

	Archivos   []string
	Duplicados int // (id, manual) repetidos: se queda el primero
}

// Total de fragmentos indexados.
func (fs *Fragmentos) Total() int {
	if fs == nil {
		return 0
	}
	return len(fs.todos)
}

// Buscar resuelve el par (id, manual).
func (fs *Fragmentos) Buscar(id, manual string) (*Fragmento, bool) {
	if fs == nil {
		return nil, false
	}
	f, ok := fs.porClave[Clave{id, manual}]
	return f, ok
}

// PorID devuelve TODOS los fragmentos con ese id (uno por manual). No sirve para citar: ver Buscar.
func (fs *Fragmentos) PorID(id string) []*Fragmento {
	if fs == nil {
		return nil
	}
	return fs.porID[id]
}

// Imagen devuelve el fragmento imagen (con su OCR) de una ruta «imagenes/<manual>/<hash>.png».
func (fs *Fragmentos) Imagen(ruta string) (*Fragmento, bool) {
	if fs == nil {
		return nil, false
	}
	f, ok := fs.porImagen[ruta]
	return f, ok
}

// Todos los fragmentos, en el orden de carga.
func (fs *Fragmentos) Todos() []*Fragmento {
	if fs == nil {
		return nil
	}
	return fs.todos
}

// fragmentoJSON: «pagina» viene como número, null o (en algunas fuentes) texto.
type fragmentoJSON struct {
	Fragmento
	PaginaRaw any `json:"pagina"`
}

// CargarFragmentos lee los JSONL. Una línea ilegible se salta y se informa; nunca aborta.
func CargarFragmentos(archivos ...string) (*Fragmentos, []ErrorCarga) {
	fs := &Fragmentos{
		porClave:  map[Clave]*Fragmento{},
		porID:     map[string][]*Fragmento{},
		porImagen: map[string]*Fragmento{},
	}
	var errs []ErrorCarga
	for _, ruta := range archivos {
		f, err := os.Open(ruta)
		if err != nil {
			errs = append(errs, ErrorCarga{Archivo: ruta, Mensaje: err.Error(), Grave: true})
			continue
		}
		fs.Archivos = append(fs.Archivos, ruta)
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 1<<20), 16<<20)
		linea := 0
		for sc.Scan() {
			linea++
			b := sc.Bytes()
			if len(strings.TrimSpace(string(b))) == 0 {
				continue
			}
			var j fragmentoJSON
			if err := json.Unmarshal(b, &j); err != nil {
				errs = append(errs, ErrorCarga{Archivo: ruta, Linea: linea, Mensaje: "JSON inválido: " + err.Error(), Grave: true})
				continue
			}
			fr := j.Fragmento
			fr.Pagina = paginaDe(j.PaginaRaw)
			fr.Archivo, fr.Linea = filepath.Base(ruta), linea
			if fr.ID == "" {
				errs = append(errs, ErrorCarga{Archivo: ruta, Linea: linea, Mensaje: "fragmento sin id", Grave: true})
				continue
			}
			c := Clave{fr.ID, fr.Manual}
			if _, ok := fs.porClave[c]; ok {
				fs.Duplicados++
				continue
			}
			p := &fr
			fs.porClave[c] = p
			fs.porID[fr.ID] = append(fs.porID[fr.ID], p)
			if p.Imagen != "" {
				if _, ok := fs.porImagen[p.Imagen]; !ok {
					fs.porImagen[p.Imagen] = p
				}
			}
			fs.todos = append(fs.todos, p)
		}
		if err := sc.Err(); err != nil {
			errs = append(errs, ErrorCarga{Archivo: ruta, Linea: linea, Mensaje: "lectura: " + err.Error(), Grave: true})
		}
		f.Close()
	}
	return fs, errs
}

func paginaDe(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(x))
		return n
	}
	return 0
}

// archivosFragmentos: kb/fragmentos*.jsonl, ordenados (fragmentos.jsonl primero).
func archivosFragmentos(dirKB string) []string {
	m, _ := filepath.Glob(filepath.Join(dirKB, "fragmentos*.jsonl"))
	sort.Slice(m, func(i, j int) bool {
		bi, bj := filepath.Base(m[i]), filepath.Base(m[j])
		if (bi == "fragmentos.jsonl") != (bj == "fragmentos.jsonl") {
			return bi == "fragmentos.jsonl"
		}
		return bi < bj
	})
	return m
}

// ErrorCarga es un problema al cargar un archivo. Grave: el elemento (archivo, procedimiento, concepto,
// línea) se saltó. Si no, es un aviso: se cargó, pero algo no cuadra (p. ej. una cita sin resolver).
type ErrorCarga struct {
	Archivo string `json:"archivo"`
	Linea   int    `json:"linea,omitempty"`
	Mensaje string `json:"mensaje"`
	Grave   bool   `json:"grave"`
}

func (e ErrorCarga) Error() string {
	nivel := "aviso"
	if e.Grave {
		nivel = "error"
	}
	if e.Linea > 0 {
		return fmt.Sprintf("%s:%d: %s: %s", e.Archivo, e.Linea, nivel, e.Mensaje)
	}
	return fmt.Sprintf("%s: %s: %s", e.Archivo, nivel, e.Mensaje)
}
