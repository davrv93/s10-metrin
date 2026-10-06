package busqueda

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Normalización del texto para la búsqueda léxica.
//
// Decisiones (medidas en eval/BUSQUEDA.md):
//
//   - Plegado: minúsculas y sin diacríticos (á→a, ü→u, ñ→n). Los usuarios
//     escriben sin tildes («como registro», «nomina») y el OCR las pierde.
//   - Palabras vacías: lista cerrada del español más el relleno típico de una
//     pregunta («hago», «quiero», «porfa»). No se quitan negaciones útiles ni
//     números.
//   - Raíz ligera propia (Raiz), no Snowball: Snowball funde «metrado» con
//     «metro» (ambos → «metr») y «partida» con «parte»; en un ERP de
//     construcción esas diferencias son justo las que importan. La raíz ligera
//     solo quita plural, género, «-mente», «-ación», «-amiento», gerundio,
//     infinitivo y participio con longitudes mínimas, de modo que registro,
//     registrar, registra y registrado comparten raíz y metrado no se toca con
//     metro.
//   - n-gramas descartados: con 8 000 fragmentos largos multiplican el índice
//     por ~5 y la tolerancia a faltas se consigue más barato con la expansión
//     difusa del vocabulario en la consulta (Indice.difusos).
//   - Términos exactos: además de la raíz, cada palabra se indexa tal cual
//     (prefijo «=») y la consulta suma ese término con PesoExacto. Así «metrado»
//     exacto pesa más que «metrados» y un atajo como «F7» o «Ctrl+F» es un único
//     término que no se raíza ni se parte.

// plegados reemplaza las letras con diacrítico más comunes en textos en
// español, portugués y francés (los manuales traen algo de cada uno por OCR).
var plegados = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ä': 'a', 'ã': 'a', 'å': 'a',
	'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'ó': 'o', 'ò': 'o', 'ô': 'o', 'ö': 'o', 'õ': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
	'ñ': 'n', 'ç': 'c', 'ý': 'y', 'ÿ': 'y',
	'º': 'o', 'ª': 'a',
}

// Plegar devuelve s en minúsculas y sin diacríticos.
func Plegar(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		r = unicode.ToLower(r)
		if p, ok := plegados[r]; ok {
			r = p
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Atajos de teclado: «Ctrl+F», «ctrl + shift + s», «Shift+F1», «Alt+F4»,
// «Control+C», «Mayús+F1». Se reconocen sobre el texto ya plegado y se
// convierten en un solo término («ctrl+f»).
var (
	reAtajo = regexp.MustCompile(`\b(ctrl|control|shift|mayus|mayusculas|alt)(\s*\+\s*(ctrl|control|shift|mayus|alt|f1[0-2]|f[1-9]|supr|enter|intro|tab|esc|inicio|fin|[a-z0-9]))+\b`)
	// Códigos con puntos: partidas «01.01.01», secciones «3.3.1», cuentas «60.1».
	reCodigo = regexp.MustCompile(`\b\d+(\.\d+)+\b`)
	// Teclas de función sueltas: F1…F12.
	reTeclaF = regexp.MustCompile(`^f(1[0-2]|[1-9])$`)
)

var nombreTecla = map[string]string{"control": "ctrl", "mayus": "shift", "mayusculas": "shift"}

// normalizarAtajo deja «ctrl + F» como «ctrl+f».
func normalizarAtajo(a string) string {
	partes := strings.Split(a, "+")
	for i, p := range partes {
		p = strings.TrimSpace(p)
		if n, ok := nombreTecla[p]; ok {
			p = n
		}
		partes[i] = p
	}
	return strings.Join(partes, "+")
}

// tramo es un token especial localizado en el texto plegado.
type tramo struct {
	ini, fin int
	texto    string
}

// Token es una unidad del texto normalizado.
type Token struct {
	Texto    string // plegado
	Especial bool   // atajo, tecla de función o código: no se raíza
}

// Tokenizar pliega el texto y lo parte en tokens. No quita palabras vacías
// (eso lo hace Terminos), para que la detección de frases y entidades vea el
// texto completo.
func Tokenizar(s string) []Token {
	t := Plegar(s)
	var out []Token
	// Primero los especiales: se sustituyen por un hueco para que el corte
	// por palabras no los parta.
	var especiales []tramo
	for _, m := range reAtajo.FindAllStringIndex(t, -1) {
		especiales = append(especiales, tramo{m[0], m[1], normalizarAtajo(t[m[0]:m[1]])})
	}
	for _, m := range reCodigo.FindAllStringIndex(t, -1) {
		solapa := false
		for _, e := range especiales {
			if m[0] < e.fin && e.ini < m[1] {
				solapa = true
				break
			}
		}
		if !solapa {
			especiales = append(especiales, tramo{m[0], m[1], t[m[0]:m[1]]})
		}
	}
	// Recorrido en orden del texto.
	pos := 0
	partir := func(seg string) {
		for _, w := range strings.FieldsFunc(seg, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}) {
			if utf8.RuneCountInString(w) < 2 && !esDigito(w) {
				continue
			}
			out = append(out, Token{Texto: w, Especial: reTeclaF.MatchString(w)})
		}
	}
	sort.Slice(especiales, func(i, j int) bool { return especiales[i].ini < especiales[j].ini })
	for _, e := range especiales {
		if e.ini < pos {
			continue
		}
		partir(t[pos:e.ini])
		out = append(out, Token{Texto: e.texto, Especial: true})
		pos = e.fin
	}
	partir(t[pos:])
	return out
}

func esDigito(w string) bool {
	for _, r := range w {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return w != ""
}

// vacias: palabras vacías del español (lista de Snowball, plegada) más el
// relleno de las preguntas de usuario. Se aplican al texto y a la consulta.
var vacias = conjunto(`de la que el en y a los del se las por un para con no una su al lo como mas pero sus le ya o
este si porque esta entre cuando muy sin sobre tambien me hasta hay donde quien desde todo nos durante todos uno les
ni contra otros ese eso ante ellos e esto mi antes algunos que unos yo otro otras otra el tanto esa estos mucho
quienes nada muchos cual poco ella estar estas algunas algo nosotros mi mis tu te ti tu tus ellas nosotras vosotros
vosotras os mio mia mios mias tuyo tuya tuyos tuyas suyo suya suyos suyas nuestro nuestra nuestros nuestras vuestro
vuestra vuestros vuestras esos esas estoy estas esta estamos estais estan este estes estemos esteis esten estare
estaras estara estaremos estareis estaran estaria estarias estariamos estariais estarian estaba estabas estabamos
estabais estaban estuve estuviste estuvo estuvimos estuvisteis estuvieron he has ha hemos habeis han haya hayas
hayamos hayais hayan habre habras habra habremos habreis habran habria habrias habriamos habriais habrian habia
habias habiamos habiais habian hube hubiste hubo hubimos hubisteis hubieron soy eres es somos sois son sea seas
seamos seais sean sere seras sera seremos sereis seran seria serias seriamos seriais serian era eras eramos erais
eran fui fuiste fue fuimos fuisteis fueron tengo tienes tiene tenemos teneis tienen tenga tengas tengamos tengais
tengan tendre tendras tendra tendremos tendreis tendran tendria tendrias tendriamos tendriais tendrian tenia tenias
teniamos teniais tenian tuve tuviste tuvo tuvimos tuvisteis tuvieron
hago hace hacer haces hacemos puedo puede pueden podria quiero quisiera necesito necesita saber sabes alguien
ayuda ayudame ayudar porfa favor gracias hola oye bueno pues osea tipo cosa cosas forma manera paso pasos q k xq
porq pq d x dnd q tb tmb cuales cuanto cuanta cuantos cuantas`)

func conjunto(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

// EsVacia dice si el término plegado es palabra vacía.
func EsVacia(w string) bool { return vacias[w] }

// Raiz reduce una palabra plegada a su raíz ligera. Los tokens especiales
// (atajos, códigos, teclas de función) y los que llevan dígitos no se tocan.
func Raiz(w string) string {
	for _, r := range w {
		if !unicode.IsLetter(r) {
			return w
		}
	}
	n := utf8.RuneCountInString(w)
	if n <= 3 {
		return w
	}
	largo := func(s string) int { return utf8.RuneCountInString(s) }
	quitar := func(s, suf string, min int) (string, bool) {
		if strings.HasSuffix(s, suf) && largo(s)-largo(suf) >= min {
			return s[:len(s)-len(suf)], true
		}
		return s, false
	}
	// 1. adverbios
	if s, ok := quitar(w, "mente", 4); ok {
		w = s
	}
	// 2. plural
	switch {
	case strings.HasSuffix(w, "ces") && largo(w) > 4:
		w = w[:len(w)-3] + "z" // luces→luz, lapices→lapiz
	case strings.HasSuffix(w, "iones") && largo(w) > 6:
		w = w[:len(w)-2] // acciones→accion
	case strings.HasSuffix(w, "es") && largo(w) > 4 && esConsonante(w[len(w)-3]):
		w = w[:len(w)-2] // valores→valor, almacenes→almacen
	case strings.HasSuffix(w, "s") && largo(w) > 3 && !strings.HasSuffix(w, "ss") && !strings.HasSuffix(w, "us"):
		w = w[:len(w)-1]
	}
	// 3. derivación y flexión verbal (solo una; longitudes mínimas para no
	// mutilar palabras cortas ni fundir metrado con metro).
	for _, r := range []struct {
		suf string
		min int
	}{
		{"aciones", 4}, {"acion", 4}, {"amiento", 4}, {"imiento", 4},
		{"ando", 4}, {"iendo", 4},
		{"ado", 5}, {"ada", 5}, {"ido", 5}, {"ida", 5},
		{"ar", 4}, {"er", 4}, {"ir", 4},
	} {
		if s, ok := quitar(w, r.suf, r.min); ok {
			return s
		}
	}
	// 4. género / vocal final
	for _, v := range []string{"a", "o", "e"} {
		if s, ok := quitar(w, v, 4); ok {
			return s
		}
	}
	return w
}

func esConsonante(b byte) bool {
	return b >= 'a' && b <= 'z' && !strings.ContainsRune("aeiou", rune(b))
}

// Termino es un término del índice: la raíz (sin prefijo) o la forma exacta
// (prefijo «=»).
const prefijoExacto = "="

// Terminos devuelve, para cada token que no es palabra vacía, su raíz y su
// forma exacta. Es el analizador del índice y de la consulta.
func Terminos(s string) (raices, exactos []string) {
	for _, t := range Tokenizar(s) {
		if !t.Especial && vacias[t.Texto] {
			continue
		}
		r := t.Texto
		if !t.Especial {
			r = Raiz(t.Texto)
		}
		raices = append(raices, r)
		exactos = append(exactos, prefijoExacto+t.Texto)
	}
	return raices, exactos
}

// Normalizar es la forma canónica de una consulta para la caché y para el
// embebedor: minúsculas, sin signos de apertura/cierre ni puntuación suelta,
// espacios simples. Conserva las tildes (el modelo vectorial las distingue) y
// los «+» y «.» internos de atajos y códigos.
func Normalizar(q string) string {
	q = strings.ToLower(q)
	var b strings.Builder
	b.Grow(len(q))
	rs := []rune(q)
	for i, r := range rs {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case (r == '+' || r == '.' || r == '-' || r == '/') && i > 0 && i < len(rs)-1 &&
			!unicode.IsSpace(rs[i-1]) && !unicode.IsSpace(rs[i+1]) && rs[i+1] != '?' && rs[i+1] != '!':
			b.WriteRune(r)
		case r == '+':
			b.WriteRune(r) // «ctrl + f»
		default:
			b.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
