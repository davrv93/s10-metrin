package conocimiento

import (
	"crypto/sha1"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

// plegar: minúsculas, sin diacríticos y con los espacios normalizados. Es la forma en que se comparan
// términos entre ** ** con su fuente (ESQUEMA.md: «sin distinguir mayúsculas ni tildes, espacios
// normalizados»).
func plegar(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	espacio := false
	for _, r := range s {
		r = unicode.ToLower(r)
		switch r {
		case 'á', 'à', 'â', 'ä', 'ã':
			r = 'a'
		case 'é', 'è', 'ê', 'ë':
			r = 'e'
		case 'í', 'ì', 'î', 'ï':
			r = 'i'
		case 'ó', 'ò', 'ô', 'ö', 'õ':
			r = 'o'
		case 'ú', 'ù', 'û', 'ü':
			r = 'u'
		case 'ñ':
			r = 'n'
		}
		if unicode.IsSpace(r) {
			espacio = true
			continue
		}
		if espacio && b.Len() > 0 {
			b.WriteByte(' ')
		}
		espacio = false
		b.WriteRune(r)
	}
	return b.String()
}

// palabras parte el texto plegado en palabras (letras y dígitos).
func palabras(s string) []string {
	return strings.FieldsFunc(plegar(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// LargoRaiz: las palabras se comparan por su prefijo de 5 letras (registre / registrar / registro →
// «regis»), como el validador de procedimientos y servidor/fotos_pasos.go.
const LargoRaiz = 5

func raiz(w string) string {
	r := []rune(w)
	if len(r) > LargoRaiz {
		return string(r[:LargoRaiz])
	}
	return w
}

// vacias: palabras sin contenido para buscar. Incluye el relleno típico de una pregunta («cómo hago»,
// «quiero», «dónde veo») y los verbos genéricos de interfaz: no distinguen un procedimiento de otro.
var vacias = conjunto(`a al algo algun alguna algunas alguno algunos ante antes asi aun bajo bien cada como con contra
cual cuales cuando de del desde donde dos e el ella ellas ello ellos en entre era es esa esas ese eso esos esta estas
este esto estos fue ha hace hacia han hasta hay la las le les lo los mas me mi mis mediante mismo muy ni no nos o otra
otras otro otros para pero poco por porque pues que quien se segun ser si sin sino sobre solo son su sus tal tambien
tan te tiene tienen todo todos tu tus un una unas uno unos y ya yo usted ustedes
hago hacer hace haga hacemos puedo puede pueden podria quiero quisiera necesito debo tengo tenemos favor porfa
ayuda ayudame ayudar dime explicame explica ensename ensena saber sabes veo ver vemos mostrar muestra
s10 erp sistema programa modulo paso pasos manera forma proceso procedimiento hola gracias
clic click boton opcion ventana pantalla elija seleccione pulse presione
pongo poner pone ponen puse coloco colocar meto meter hice hizo hacemos dejo dejar deja quede queda
debajo encima abajo arriba dentro adentro fuera afuera luego despues ahora ayer hoy siempre nunca todavia aun
buenas buenos dias tardes noches estimado estimada saludos oye bueno vale
significa significado refiere`)

// fonetica: forma «como suena» de una palabra plegada, para tolerar faltas de ortografía comunes
// («kiero», «aser», «nuebo», «komo»): qu/c(a,o,u)/k → k; c(e,i)/z → s; v/w → b; h muda; ll → y;
// g(e,i) → j; letras dobles una vez.
var reemplazosFoneticos = strings.NewReplacer("ch", "x", "qu", "k", "ll", "y", "ge", "je", "gi", "ji", "h", "", "v", "b", "w", "b", "z", "s")

func fonetica(w string) string {
	w = reemplazosFoneticos.Replace(w)
	var b strings.Builder
	var prev byte
	for i := 0; i < len(w); i++ {
		c := w[i]
		if c == 'c' {
			if i+1 < len(w) && (w[i+1] == 'e' || w[i+1] == 'i') {
				c = 's'
			} else {
				c = 'k'
			}
		}
		if c == prev {
			continue
		}
		b.WriteByte(c)
		prev = c
	}
	return b.String()
}

// distancia1: a y b difieren en a lo sumo una edición (sustitución, inserción o borrado).
func distancia1(a, b string) bool {
	if a == b {
		return true
	}
	la, lb := len(a), len(b)
	if la-lb > 1 || lb-la > 1 {
		return false
	}
	if la < lb {
		a, b, la, lb = b, a, lb, la
	}
	i, j, ediciones := 0, 0, 0
	for i < la && j < lb {
		if a[i] == b[j] {
			i++
			j++
			continue
		}
		ediciones++
		if ediciones > 1 {
			return false
		}
		if la > lb {
			i++
		} else {
			i++
			j++
		}
	}
	return ediciones+(la-i)+(lb-j) <= 1
}

func conjunto(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

// terminos devuelve las raíces de contenido del texto, en orden y con repetición.
func terminos(s string) []string {
	var out []string
	for _, w := range palabras(s) {
		if len(w) < 2 || vacias[w] {
			continue
		}
		out = append(out, raiz(w))
	}
	return out
}

// sinonimosVerbo: formas en que el usuario pide la misma acción con otro verbo («guardar» frente al
// «grabar» de los manuales). Son pocas y de interfaz; los sinónimos de dominio vienen de los YAML
// (aliases de procedimientos, sinónimos del glosario). La consulta los suma con peso reducido.
var sinonimosVerbo = map[string][]string{
	"guard": {"graba"},
	"graba": {"guard"},
	"modif": {"edita", "cambi"},
	"edita": {"modif"},
	"cambi": {"modif"},
	"crear": {"regis", "nuevo"},
	"creo":  {"regis", "nuevo"},
	"ingre": {"regis"},
	"carga": {"regis", "ingre"},
	"agreg": {"regis", "adici"},
	"adici": {"agreg"},
	"borra": {"elimi"},
	"elimi": {"borra"},
	"consu": {"revis"},
	"revis": {"consu"},
	"impri": {"repor"},
}

// negritaRe: términos entre ** ** (menús, botones, campos, ventanas).
var negritaRe = regexp.MustCompile(`\*\*([^*\n]+?)\*\*`)

// negritas devuelve los términos entre ** ** del texto, tal como están.
func negritas(s string) []string {
	var out []string
	for _, m := range negritaRe.FindAllStringSubmatch(s, -1) {
		if t := strings.TrimSpace(m[1]); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// sinNegritas quita las marcas ** ** y deja el término.
func sinNegritas(s string) string { return negritaRe.ReplaceAllString(s, "$1") }

// idFoto: los 12 primeros hex de sha1(ruta_o_url), como en ESQUEMA.md.
func idFoto(ruta string) string {
	h := sha1.Sum([]byte(ruta))
	return hex.EncodeToString(h[:])[:12]
}

var rutaFotoRe = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)

// rutaFotoValida: la misma regla que servidor/fotos_pasos.go (ruta relativa bajo data/, sin «..»), o una
// URL http(s).
func rutaFotoValida(r string) bool {
	if strings.HasPrefix(r, "https://") || strings.HasPrefix(r, "http://") {
		return !strings.ContainsAny(r, " \n\"<>()")
	}
	return rutaFotoRe.MatchString(r) && !strings.Contains(r, "..") && !strings.HasPrefix(r, "/")
}

// URLFoto: cómo la pinta la página. Las rutas relativas se sirven en GET /fotos/<ruta> (servidor.go);
// las URL van tal cual.
func URLFoto(ruta string) string {
	if strings.HasPrefix(ruta, "http://") || strings.HasPrefix(ruta, "https://") {
		return ruta
	}
	return "fotos/" + ruta
}

// contieneFrase dice si el término plegado aparece en el texto plegado como palabra(s) completas.
func contieneFrase(textoPlegado, termino string) bool {
	t := plegar(termino)
	if t == "" {
		return false
	}
	for i := 0; ; {
		j := strings.Index(textoPlegado[i:], t)
		if j < 0 {
			return false
		}
		j += i
		antes := j == 0 || !esLetra(rune(textoPlegado[j-1]))
		fin := j + len(t)
		despues := fin >= len(textoPlegado) || !esLetra(rune(textoPlegado[fin]))
		if antes && despues {
			return true
		}
		i = j + 1
		if i >= len(textoPlegado) {
			return false
		}
	}
}

func esLetra(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

// minusculaInicial: «Registrar un presupuesto» → «registrar un presupuesto», para frases como «¿Quiere
// que le enseñe a …?». Respeta siglas («WBS»).
func minusculaInicial(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	if len(r) > 1 && unicode.IsUpper(r[1]) {
		return s
	}
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func unicos(xs []string) []string {
	visto := map[string]bool{}
	var out []string
	for _, x := range xs {
		if x == "" || visto[x] {
			continue
		}
		visto[x] = true
		out = append(out, x)
	}
	return out
}
