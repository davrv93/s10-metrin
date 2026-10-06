package main

// Lector del subconjunto fijo de YAML de kb/procedimientos/ESQUEMA.md (el módulo no trae dependencia
// YAML y go.mod no se toca desde aquí). Es el mismo subconjunto que lee `yaml_minimo` de
// herramientas/validar_procedimientos.py: mapas y listas por sangría, textos entre comillas dobles con
// escapes JSON, listas y mapas en línea, null/true/false y números. Si un archivo se sale del
// subconjunto (por ejemplo, un `|` en un glosario), cargarYAML recurre a PyYAML del .venv del repo
// cuando está disponible, y si no, devuelve el error: el arnés avisa y sigue sin ese archivo.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type lineaYAML struct {
	ind   int
	texto string
}

type lectorYAML struct {
	lineas []lineaYAML
	pos    int
}

var reClave = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*):(?:\s+(.*))?$`)
var reEntero = regexp.MustCompile(`^-?\d+$`)
var reDecimal = regexp.MustCompile(`^-?\d+\.\d+$`)

// yamlMinimo convierte el texto en map[string]any / []any / string / float64 / bool / nil.
func yamlMinimo(texto string) (map[string]any, error) {
	l := &lectorYAML{}
	for _, crudo := range strings.Split(strings.ReplaceAll(texto, "\r\n", "\n"), "\n") {
		s := quitarComentario(crudo)
		if strings.TrimSpace(s) == "" {
			continue
		}
		ind := len(s) - len(strings.TrimLeft(s, " "))
		l.lineas = append(l.lineas, lineaYAML{ind, strings.TrimSpace(s)})
	}
	if len(l.lineas) == 0 {
		return map[string]any{}, nil
	}
	m, err := l.mapa(0)
	if err != nil {
		return nil, err
	}
	if l.pos < len(l.lineas) {
		return nil, fmt.Errorf("línea no consumida: %q", recortarTexto(l.lineas[l.pos].texto, 60))
	}
	return m, nil
}

// quitarComentario corta un « #» fuera de comillas (y las líneas que empiezan por #).
func quitarComentario(s string) string {
	t := strings.TrimSpace(s)
	if strings.HasPrefix(t, "#") {
		return ""
	}
	enComillas := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case enComillas && c == '\\':
			i++
		case c == '"':
			enComillas = !enComillas
		case !enComillas && c == '#' && i > 0 && (s[i-1] == ' ' || s[i-1] == '\t'):
			return strings.TrimRight(s[:i], " \t")
		}
	}
	return s
}

func (l *lectorYAML) bloque(ind int) (any, error) {
	if strings.HasPrefix(l.lineas[l.pos].texto, "- ") || l.lineas[l.pos].texto == "-" {
		return l.lista(ind)
	}
	return l.mapa(ind)
}

func (l *lectorYAML) mapa(ind int) (map[string]any, error) {
	d := map[string]any{}
	for l.pos < len(l.lineas) {
		ln := l.lineas[l.pos]
		if ln.ind < ind || strings.HasPrefix(ln.texto, "- ") {
			break
		}
		if ln.ind > ind {
			return nil, fmt.Errorf("sangría inesperada: %q", recortarTexto(ln.texto, 60))
		}
		m := reClave.FindStringSubmatch(ln.texto)
		if m == nil {
			return nil, fmt.Errorf("línea no reconocida: %q", recortarTexto(ln.texto, 60))
		}
		k, v := m[1], strings.TrimSpace(m[2])
		l.pos++
		if v == "" {
			switch {
			case l.pos < len(l.lineas) && l.lineas[l.pos].ind > ind:
				b, err := l.bloque(l.lineas[l.pos].ind)
				if err != nil {
					return nil, err
				}
				d[k] = b
			case l.pos < len(l.lineas) && l.lineas[l.pos].ind == ind && strings.HasPrefix(l.lineas[l.pos].texto, "- "):
				b, err := l.lista(ind)
				if err != nil {
					return nil, err
				}
				d[k] = b
			default:
				d[k] = nil
			}
			continue
		}
		x, err := escalarYAML(v)
		if err != nil {
			return nil, err
		}
		d[k] = x
	}
	return d, nil
}

func (l *lectorYAML) lista(ind int) ([]any, error) {
	out := []any{}
	for l.pos < len(l.lineas) {
		ln := l.lineas[l.pos]
		if ln.ind != ind || !strings.HasPrefix(ln.texto, "- ") {
			break
		}
		resto := strings.TrimSpace(ln.texto[2:])
		if m := reClave.FindStringSubmatch(resto); m != nil && !strings.HasPrefix(resto, "\"") {
			// elemento mapa: la primera clave va en la línea del guion
			l.lineas[l.pos] = lineaYAML{ind + 2, resto}
			m, err := l.mapa(ind + 2)
			if err != nil {
				return nil, err
			}
			out = append(out, m)
			continue
		}
		l.pos++
		x, err := escalarYAML(resto)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, nil
}

func escalarYAML(v string) (any, error) {
	v = strings.TrimSpace(v)
	switch {
	case v == "":
		return nil, nil
	case strings.HasPrefix(v, "\""):
		var s string
		if err := json.Unmarshal([]byte(v), &s); err != nil {
			return nil, fmt.Errorf("texto entre comillas inválido %q: %w", recortarTexto(v, 40), err)
		}
		return s, nil
	case strings.HasPrefix(v, "["):
		if !strings.HasSuffix(v, "]") {
			return nil, fmt.Errorf("lista en línea sin cerrar: %q", recortarTexto(v, 40))
		}
		out := []any{}
		for _, p := range partirEnLinea(v[1 : len(v)-1]) {
			x, err := escalarYAML(p)
			if err != nil {
				return nil, err
			}
			out = append(out, x)
		}
		return out, nil
	case strings.HasPrefix(v, "{"):
		if !strings.HasSuffix(v, "}") {
			return nil, fmt.Errorf("mapa en línea sin cerrar: %q", recortarTexto(v, 40))
		}
		d := map[string]any{}
		for _, par := range partirEnLinea(v[1 : len(v)-1]) {
			k, x, ok := strings.Cut(par, ":")
			if !ok {
				return nil, fmt.Errorf("par sin «:» en mapa en línea: %q", recortarTexto(par, 40))
			}
			val, err := escalarYAML(x)
			if err != nil {
				return nil, err
			}
			d[strings.TrimSpace(k)] = val
		}
		return d, nil
	case v == "null" || v == "~":
		return nil, nil
	case v == "true" || v == "false":
		return v == "true", nil
	case reEntero.MatchString(v) || reDecimal.MatchString(v):
		f, _ := strconv.ParseFloat(v, 64)
		return f, nil
	case strings.HasPrefix(v, "'") || strings.HasPrefix(v, "|") || strings.HasPrefix(v, ">") ||
		strings.HasPrefix(v, "&") || strings.HasPrefix(v, "*"):
		return nil, fmt.Errorf("sintaxis fuera del subconjunto: %q", recortarTexto(v, 30))
	}
	return v, nil
}

// partirEnLinea separa por comas de nivel 0, respetando comillas y anidamiento.
func partirEnLinea(s string) []string {
	var out []string
	var cur strings.Builder
	prof, comillas := 0, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case comillas:
			cur.WriteByte(c)
			if c == '\\' && i+1 < len(s) {
				i++
				cur.WriteByte(s[i])
			} else if c == '"' {
				comillas = false
			}
		case c == '"':
			comillas = true
			cur.WriteByte(c)
		case c == '[' || c == '{':
			prof++
			cur.WriteByte(c)
		case c == ']' || c == '}':
			prof--
			cur.WriteByte(c)
		case c == ',' && prof == 0:
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	if strings.TrimSpace(cur.String()) != "" {
		out = append(out, cur.String())
	}
	return out
}

// cargarYAML lee un archivo: primero con el lector mínimo y, si falla, con PyYAML (solo lectura local).
func cargarYAML(ruta, raizRepo string) (map[string]any, error) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	m, errMin := yamlMinimo(string(b))
	if errMin == nil {
		return m, nil
	}
	for _, py := range []string{filepath.Join(raizRepo, ".venv", "bin", "python3"), "python3"} {
		if _, err := exec.LookPath(py); err != nil && !filepath.IsAbs(py) {
			continue
		}
		if filepath.IsAbs(py) {
			if _, err := os.Stat(py); err != nil {
				continue
			}
		}
		out, err := exec.Command(py, "-c",
			"import sys,json,yaml,datetime\n"+
				"d=yaml.safe_load(open(sys.argv[1],encoding='utf-8'))\n"+
				"print(json.dumps(d,ensure_ascii=False,default=str))", ruta).Output()
		if err != nil {
			continue
		}
		var d map[string]any
		if json.Unmarshal(out, &d) == nil {
			return d, nil
		}
	}
	return nil, fmt.Errorf("%s: %w", filepath.Base(ruta), errMin)
}

// decodificar pasa un valor genérico a una estructura por JSON (las claves YAML van en las etiquetas json).
func decodificar(v any, destino any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, destino)
}

func recortarTexto(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
