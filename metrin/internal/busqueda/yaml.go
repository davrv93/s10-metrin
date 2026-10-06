package busqueda

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Lector mínimo de YAML para los archivos de datos de Metrín
// (kb/procedimientos/**/*.yml, kb/conceptos/*.yml, eval/alias.yml). Cubre el
// «subconjunto fijo» de kb/procedimientos/ESQUEMA.md y nada más, para no
// añadir una dependencia al módulo:
//
//   - mapas y listas en bloque con sangría de espacios («clave: valor»,
//     «- elemento», «- clave: valor» con las demás claves alineadas);
//   - listas y mapas en línea: [a, "b"], {id: x, manual: "y"}, anidados;
//   - escalares sin comillas, entre comillas dobles (escapes JSON) o simples;
//   - comentarios con «#» al inicio o tras un espacio fuera de comillas.
//
// No admite anclas, etiquetas, bloques «|» / «>» ni cadenas en varias líneas:
// devuelve error si los encuentra. Los escalares se devuelven como string
// (null y ~ como nil); el que lee decide si un valor es número.

type lineaYAML struct {
	ind int
	txt string
	num int
}

// LeerYAML interpreta el subconjunto descrito arriba.
func LeerYAML(datos []byte) (any, error) {
	var ls []lineaYAML
	for i, l := range strings.Split(strings.ReplaceAll(string(datos), "\r\n", "\n"), "\n") {
		if strings.ContainsRune(l, '\t') && strings.TrimLeft(l, " \t") != strings.TrimLeft(l, " ") {
			return nil, fmt.Errorf("yaml línea %d: sangría con tabulador", i+1)
		}
		sin := quitarComentario(l)
		t := strings.TrimRight(sin, " \t")
		if strings.TrimSpace(t) == "" || strings.TrimSpace(t) == "---" {
			continue
		}
		ind := len(t) - len(strings.TrimLeft(t, " "))
		ls = append(ls, lineaYAML{ind: ind, txt: strings.TrimSpace(t), num: i + 1})
	}
	if len(ls) == 0 {
		return nil, nil
	}
	v, i, err := nodoYAML(ls, 0)
	if err != nil {
		return nil, err
	}
	if i < len(ls) {
		return nil, fmt.Errorf("yaml línea %d: sangría inesperada", ls[i].num)
	}
	return v, nil
}

func esGuion(t string) bool { return t == "-" || strings.HasPrefix(t, "- ") }

func nodoYAML(ls []lineaYAML, i int) (any, int, error) {
	if esGuion(ls[i].txt) {
		return secuenciaYAML(ls, i, ls[i].ind)
	}
	if _, _, ok := partirClave(ls[i].txt); !ok {
		v, err := escalarYAML(ls[i].txt, ls[i].num)
		return v, i + 1, err
	}
	return mapaYAML(ls, i, ls[i].ind)
}

func secuenciaYAML(ls []lineaYAML, i, ind int) (any, int, error) {
	out := []any{}
	for i < len(ls) && ls[i].ind == ind && esGuion(ls[i].txt) {
		resto := strings.TrimPrefix(ls[i].txt, "-")
		despl := len(resto) - len(strings.TrimLeft(resto, " "))
		resto = strings.TrimSpace(resto)
		if resto == "" {
			i++
			if i < len(ls) && ls[i].ind > ind {
				v, j, err := nodoYAML(ls, i)
				if err != nil {
					return nil, i, err
				}
				out = append(out, v)
				i = j
			} else {
				out = append(out, nil)
			}
			continue
		}
		if _, _, ok := partirClave(resto); ok && resto[0] != '{' && resto[0] != '[' && resto[0] != '"' && resto[0] != '\'' {
			// «- clave: valor»: el mapa empieza en la columna de la clave.
			col := ind + 1 + despl
			ls[i] = lineaYAML{ind: col, txt: resto, num: ls[i].num}
			v, j, err := mapaYAML(ls, i, col)
			if err != nil {
				return nil, i, err
			}
			out = append(out, v)
			i = j
			continue
		}
		v, err := escalarYAML(resto, ls[i].num)
		if err != nil {
			return nil, i, err
		}
		out = append(out, v)
		i++
	}
	return out, i, nil
}

func mapaYAML(ls []lineaYAML, i, ind int) (any, int, error) {
	m := map[string]any{}
	for i < len(ls) && ls[i].ind == ind && !esGuion(ls[i].txt) {
		k, v, ok := partirClave(ls[i].txt)
		if !ok {
			return nil, i, fmt.Errorf("yaml línea %d: se esperaba «clave: valor»", ls[i].num)
		}
		num := ls[i].num
		i++
		var val any
		switch {
		case v == "|" || v == ">" || strings.HasPrefix(v, "|") || strings.HasPrefix(v, ">"):
			return nil, i, fmt.Errorf("yaml línea %d: bloques «|»/«>» no admitidos", num)
		case strings.HasPrefix(v, "&") || strings.HasPrefix(v, "*") || strings.HasPrefix(v, "!"):
			return nil, i, fmt.Errorf("yaml línea %d: anclas y etiquetas no admitidas", num)
		case v == "":
			if i < len(ls) && (ls[i].ind > ind || (ls[i].ind == ind && esGuion(ls[i].txt))) {
				var err error
				val, i, err = nodoYAML(ls, i)
				if err != nil {
					return nil, i, err
				}
			}
		default:
			var err error
			val, err = escalarYAML(v, num)
			if err != nil {
				return nil, i, err
			}
		}
		m[k] = val
	}
	return m, i, nil
}

// partirClave separa «clave: valor» en el primer «:» fuera de comillas que
// va seguido de espacio o fin de línea.
func partirClave(t string) (string, string, bool) {
	if t == "" || t[0] == '[' || t[0] == '{' {
		return "", "", false
	}
	enD, enS := false, false
	for i := 0; i < len(t); i++ {
		c := t[i]
		switch {
		case enD:
			if c == '\\' {
				i++
			} else if c == '"' {
				enD = false
			}
		case enS:
			if c == '\'' {
				enS = false
			}
		case c == '"':
			enD = true
		case c == '\'':
			enS = true
		case c == ':' && (i == len(t)-1 || t[i+1] == ' '):
			k := strings.TrimSpace(t[:i])
			if len(k) >= 2 && (k[0] == '"' || k[0] == '\'') {
				if s, err := cadenaYAML(k); err == nil {
					k = s
				}
			}
			return k, strings.TrimSpace(t[i+1:]), k != ""
		}
	}
	return "", "", false
}

func quitarComentario(l string) string {
	enD, enS := false, false
	for i := 0; i < len(l); i++ {
		c := l[i]
		switch {
		case enD:
			if c == '\\' {
				i++
			} else if c == '"' {
				enD = false
			}
		case enS:
			if c == '\'' {
				enS = false
			}
		case c == '"':
			enD = true
		case c == '\'':
			enS = true
		case c == '#' && (i == 0 || l[i-1] == ' ' || l[i-1] == '\t'):
			return l[:i]
		}
	}
	return l
}

func escalarYAML(s string, num int) (any, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if s[0] == '[' || s[0] == '{' {
		p := &flujoYAML{s: s, num: num}
		v, err := p.valor()
		if err != nil {
			return nil, err
		}
		p.espacios()
		if p.i != len(p.s) {
			return nil, fmt.Errorf("yaml línea %d: texto sobrante tras %q", num, s[:p.i])
		}
		return v, nil
	}
	if s[0] == '"' || s[0] == '\'' {
		v, err := cadenaYAML(s)
		if err != nil {
			return nil, fmt.Errorf("yaml línea %d: %w", num, err)
		}
		return v, nil
	}
	if s == "null" || s == "~" || s == "Null" || s == "NULL" {
		return nil, nil
	}
	return s, nil
}

func cadenaYAML(s string) (string, error) {
	if s[0] == '"' {
		var v string
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			return "", fmt.Errorf("cadena entre comillas dobles inválida %s", s)
		}
		return v, nil
	}
	if len(s) < 2 || s[len(s)-1] != '\'' {
		return "", fmt.Errorf("cadena entre comillas simples sin cerrar %s", s)
	}
	return strings.ReplaceAll(s[1:len(s)-1], "''", "'"), nil
}

// flujoYAML lee listas y mapas en línea.
type flujoYAML struct {
	s   string
	i   int
	num int
}

func (p *flujoYAML) espacios() {
	for p.i < len(p.s) && (p.s[p.i] == ' ' || p.s[p.i] == '\t') {
		p.i++
	}
}

func (p *flujoYAML) err(msg string) error {
	return fmt.Errorf("yaml línea %d: %s en %q", p.num, msg, p.s)
}

func (p *flujoYAML) valor() (any, error) {
	p.espacios()
	if p.i >= len(p.s) {
		return nil, p.err("valor vacío")
	}
	switch p.s[p.i] {
	case '[':
		p.i++
		out := []any{}
		for {
			p.espacios()
			if p.i < len(p.s) && p.s[p.i] == ']' {
				p.i++
				return out, nil
			}
			v, err := p.valor()
			if err != nil {
				return nil, err
			}
			out = append(out, v)
			p.espacios()
			if p.i < len(p.s) && p.s[p.i] == ',' {
				p.i++
				continue
			}
			if p.i < len(p.s) && p.s[p.i] == ']' {
				p.i++
				return out, nil
			}
			return nil, p.err("lista sin cerrar")
		}
	case '{':
		p.i++
		m := map[string]any{}
		for {
			p.espacios()
			if p.i < len(p.s) && p.s[p.i] == '}' {
				p.i++
				return m, nil
			}
			kv, err := p.plano(":")
			if err != nil {
				return nil, err
			}
			k, _ := kv.(string)
			if p.i >= len(p.s) || p.s[p.i] != ':' {
				return nil, p.err("falta «:» en mapa en línea")
			}
			p.i++
			v, err := p.valor()
			if err != nil {
				return nil, err
			}
			m[k] = v
			p.espacios()
			if p.i < len(p.s) && p.s[p.i] == ',' {
				p.i++
				continue
			}
			if p.i < len(p.s) && p.s[p.i] == '}' {
				p.i++
				return m, nil
			}
			return nil, p.err("mapa sin cerrar")
		}
	}
	return p.plano(",]}")
}

// plano lee una cadena entre comillas o un escalar sin comillas hasta uno de
// los delimitadores.
func (p *flujoYAML) plano(delim string) (any, error) {
	p.espacios()
	if p.i < len(p.s) && (p.s[p.i] == '"' || p.s[p.i] == '\'') {
		q := p.s[p.i]
		j := p.i + 1
		for j < len(p.s) {
			if q == '"' && p.s[j] == '\\' {
				j += 2
				continue
			}
			if p.s[j] == q {
				if q == '\'' && j+1 < len(p.s) && p.s[j+1] == '\'' {
					j += 2
					continue
				}
				break
			}
			j++
		}
		if j >= len(p.s) {
			return nil, p.err("cadena sin cerrar")
		}
		v, err := cadenaYAML(p.s[p.i : j+1])
		if err != nil {
			return nil, err
		}
		p.i = j + 1
		return v, nil
	}
	j := p.i
	for j < len(p.s) && !strings.ContainsRune(delim, rune(p.s[j])) {
		j++
	}
	v := strings.TrimSpace(p.s[p.i:j])
	p.i = j
	if v == "null" || v == "~" || v == "" {
		return nil, nil
	}
	return v, nil
}

// Ayudas para leer el árbol devuelto por LeerYAML.

func yamlTexto(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	}
	return fmt.Sprint(v)
}

func yamlLista(v any) []any {
	if l, ok := v.([]any); ok {
		return l
	}
	if v == nil {
		return nil
	}
	return []any{v}
}

func yamlTextos(v any) []string {
	var out []string
	for _, e := range yamlLista(v) {
		if s := strings.TrimSpace(yamlTexto(e)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func yamlMapa(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}
