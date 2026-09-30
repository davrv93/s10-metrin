package servidor

import (
	"regexp"
	"strings"
	"unicode"

	"rag-go/internal/rag"
)

// colocarFotos pone cada captura del manual debajo de la línea de la respuesta
// que explica su paso. El modelo reescribe los pasos con sus palabras, así que
// se compara por palabras de contenido: la línea que comparte más con el paso
// del manual (al menos minComunes y la mitad de las del lado más corto) se
// queda con sus fotos. Las fotos colocadas salen de la galería de su fuente;
// las que no encuentran paso siguen ahí.
func colocarFotos(respuesta string, fuentes []rag.Fuente) string {
	lineas := strings.Split(respuesta, "\n")
	claves := make([]map[string]bool, len(lineas))
	for i, l := range lineas {
		s := strings.TrimSpace(l)
		if s == "" || strings.HasPrefix(s, "![") {
			continue
		}
		claves[i] = palabrasClave(s)
	}
	debajo := map[int][]string{}
	puestas := map[string]bool{}
	for _, f := range fuentes {
		for _, p := range f.Pasos {
			if len(p.Fotos) == 0 {
				continue
			}
			paso := palabrasClave(p.Texto)
			mejor, puntos := -1, 0.0
			for i, c := range claves {
				if c == nil {
					continue
				}
				if v := parecido(paso, c); v > puntos {
					mejor, puntos = i, v
				}
			}
			if mejor < 0 {
				continue
			}
			for _, foto := range p.Fotos {
				if !rutaFotoValida(foto) || puestas[foto] {
					continue
				}
				puestas[foto] = true
				debajo[mejor] = append(debajo[mejor], foto)
			}
		}
	}
	if len(debajo) == 0 {
		return respuesta
	}
	var b strings.Builder
	for i, l := range lineas {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(l)
		for _, foto := range debajo[i] {
			b.WriteString("\n![Captura del manual](fotos/" + foto + ")")
		}
	}
	for i := range fuentes {
		quedan := fuentes[i].Fotos[:0:0]
		for _, foto := range fuentes[i].Fotos {
			if !puestas[foto] {
				quedan = append(quedan, foto)
			}
		}
		fuentes[i].Fotos = quedan
	}
	return b.String()
}

const minComunes = 2

// parecido: 0 si no alcanza el mínimo; si no, la fracción de palabras
// compartidas sobre el lado más corto.
func parecido(a, b map[string]bool) float64 {
	comunes := 0
	for p := range a {
		if b[p] {
			comunes++
		}
	}
	menor := min(len(a), len(b))
	if menor == 0 || comunes < minComunes {
		return 0
	}
	v := float64(comunes) / float64(menor)
	if v < 0.5 {
		return 0
	}
	return v
}

var vacias = map[string]bool{
	"para": true, "como": true, "que": true, "los": true, "las": true, "del": true, "una": true,
	"uno": true, "con": true, "por": true, "sus": true, "este": true, "esta": true, "debe": true,
	"puede": true, "desde": true, "entre": true, "sobre": true, "donde": true, "cual": true,
	"tambien": true, "luego": true, "despues": true, "paso": true, "clic": true, "click": true,
	"opcion": true, "boton": true, "pantalla": true, "ventana": true, "hacer": true, "haga": true,
}

var sinMarcas = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

func palabrasClave(s string) map[string]bool {
	s = sinMarcas.Replace(strings.ToLower(s))
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if len(w) < 3 || vacias[w] {
			continue
		}
		if r := []rune(w); len(r) > 5 { // raíz corta: «genere»/«generar», «seleccione»/«seleccionar»
			w = string(r[:5])
		}
		out[w] = true
	}
	return out
}

var rutaFoto = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)

func rutaFotoValida(r string) bool {
	return rutaFoto.MatchString(r) && !strings.Contains(r, "..") && !strings.HasPrefix(r, "/")
}
