// Package reportes: preguntas en lenguaje natural -> reportes SQL, sin SQL libre.
// La IA (o el usuario) elige y rellena una plantilla del catálogo aprobado
// (reportes/plantillas.json, la misma fuente que consume el prototipo Python).
package reportes

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// Límites de protección. La garantía real de solo-lectura es el rol de la BD.
const (
	MaxFilas        = 500
	TimeoutConsulta = 15 // segundos
)

// Plantilla y Parametro reflejan reportes/plantillas.json.
type Plantilla struct {
	ID          string               `json:"id"`
	Nombre      string               `json:"nombre"`
	Titulo      string               `json:"titulo"`
	Descripcion string               `json:"descripcion"`
	Ejemplos    []string             `json:"ejemplos"`
	SQL         string               `json:"sql"`
	Parametros  map[string]Parametro `json:"parametros"`
}

// Parametro es un parámetro declarado de una plantilla.
type Parametro struct {
	Tipo    string  `json:"tipo"`
	Defecto *string `json:"defecto"`
}

// Catalogo es el JSON completo del catálogo.
type Catalogo struct {
	Version    string      `json:"version"`
	Nota       string      `json:"nota"`
	Plantillas []Plantilla `json:"plantillas"`
}

// CargarCatalogo lee el archivo del catálogo.
func CargarCatalogo(ruta string) (Catalogo, error) {
	var c Catalogo
	raw, err := os.ReadFile(ruta)
	if err != nil {
		return c, fmt.Errorf("catálogo %s: %w", ruta, err)
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, fmt.Errorf("catálogo: %w", err)
	}
	if len(c.Plantillas) == 0 {
		return c, fmt.Errorf("catálogo sin plantillas: %s", ruta)
	}
	return c, nil
}

// ── normalización y recuperación ───────────────────────────────────────────

var palabrasVacias = map[string]bool{
	"de": true, "la": true, "el": true, "los": true, "las": true, "del": true,
	"y": true, "o": true, "a": true, "en": true, "que": true, "cuanto": true,
	"cuantos": true, "cuantas": true, "cual": true, "cuales": true, "muestrame": true,
	"muestra": true, "dame": true, "por": true, "para": true, "con": true, "un": true,
	"una": true, "es": true, "son": true, "al": true, "sobre": true, "reporte": true,
	"reportes": true, "quiero": true, "necesito": true, "hay": true, "se": true,
	"su": true, "sus": true, "select": true, "the": true,
}

// normaliza: minúsculas, sin acentos, sin palabras vacías.
func normaliza(t string) []string {
	t = strings.ToLower(t)
	t = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) { // diacríticos: á->a
			return -1
		}
		return r
	}, t)
	var out []string
	for _, p := range strings.FieldsFunc(t, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if !palabrasVacias[p] {
			out = append(out, p)
		}
	}
	return out
}

// Recuperar devuelve la plantilla con más solapamiento léxico con la pregunta.
func (c Catalogo) Recuperar(pregunta string) (*Plantilla, float64, error) {
	q := map[string]bool{}
	for _, p := range normaliza(pregunta) {
		q[p] = true
	}
	if len(q) == 0 {
		return nil, 0, fmt.Errorf("pregunta vacía")
	}
	var mejor *Plantilla
	puntaje := 0.0
	for i := range c.Plantillas {
		p := &c.Plantillas[i]
		corpus := normaliza(p.Descripcion + " " + strings.ReplaceAll(p.Nombre, "-", " ") +
			" " + p.Titulo + " " + strings.Join(p.Ejemplos, " ") + " " + strings.ToLower(p.SQL))
		set := map[string]bool{}
		for _, t := range corpus {
			set[t] = true
		}
		inter := 0
		for t := range q {
			if set[t] {
				inter++
			}
		}
		if s := float64(inter) / float64(len(q)); s > puntaje {
			mejor, puntaje = p, s
		}
	}
	if mejor == nil || puntaje == 0 {
		return nil, 0, fmt.Errorf("ninguna plantilla corresponde a la pregunta")
	}
	return mejor, puntaje, nil
}

// ── resolución de parámetros ───────────────────────────────────────────────

var (
	reRango   = regexp.MustCompile(`entre (20\d{2})-(\d{2}) y (20\d{2})-(\d{2})`)
	rePeriodo = regexp.MustCompile(`\b(20\d{2})-(\d{2})\b`)
	reMesAño  = regexp.MustCompile(`(enero|febrero|marzo|abril|mayo|junio|julio|agosto|setiembre|septiembre|octubre|noviembre|diciembre)\s+(?:de\s+)?(20\d{2})`)
	reAño     = regexp.MustCompile(`\b(20\d{2})\b`)
	reCateg   = regexp.MustCompile(`\b(operari[oa]s?|oficial(es)?|peon(es)?)\b`)

	meses = map[string]int{"enero": 1, "febrero": 2, "marzo": 3, "abril": 4, "mayo": 5,
		"junio": 6, "julio": 7, "agosto": 8, "setiembre": 9, "septiembre": 9,
		"octubre": 10, "noviembre": 11, "diciembre": 12}
)

// Resolver extrae parámetros de la pregunta con reglas explícitas (sin IA:
// no hay alucinación de parámetros).
func Resolver(pregunta string) map[string]any {
	t := strings.ToLower(pregunta)
	t = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, t)
	p := map[string]any{}

	if m := reRango.FindStringSubmatch(t); m != nil {
		p["desde"] = fmt.Sprintf("%s-%s-01", m[1], m[2])
		p["hasta"] = fmt.Sprintf("%s-%s-28", m[3], m[4])
	}
	if m := rePeriodo.FindStringSubmatch(t); m != nil {
		p["periodo"] = fmt.Sprintf("%s-%s", m[1], m[2])
	}
	if m := reMesAño.FindStringSubmatch(t); m != nil {
		mm := meses[m[1]]
		p["periodo"] = fmt.Sprintf("%s-%02d", m[2], mm)
		if _, ok := p["desde"]; !ok {
			p["desde"] = fmt.Sprintf("%s-%02d-01", m[2], mm)
			p["hasta"] = fmt.Sprintf("%s-%02d-31", m[2], mm)
		}
	}
	if m := reAño.FindStringSubmatch(t); m != nil {
		if _, ok := p["desde"]; !ok {
			p["desde"] = m[1] + "-01-01"
			p["hasta"] = m[1] + "-12-31"
		}
	}
	if m := reCateg.FindStringSubmatch(t); m != nil {
		switch {
		case strings.HasPrefix(m[1], "operari"):
			p["categoria"] = "operario"
		case strings.HasPrefix(m[1], "oficial"):
			p["categoria"] = "oficial"
		default:
			p["categoria"] = "peón"
		}
	}
	return p
}

// Completar rellena con los defectos declarados y recorta a lo declarado.
func (t *Plantilla) Completar(params map[string]any) map[string]any {
	out := map[string]any{}
	for nombre, spec := range t.Parametros {
		if v, ok := params[nombre]; ok {
			out[nombre] = v
		} else if spec.Defecto != nil {
			out[nombre] = *spec.Defecto
		} else {
			out[nombre] = nil
		}
	}
	return out
}
