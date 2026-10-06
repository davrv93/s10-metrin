package main

// Conocimiento de referencia que usa el arnés para calificar: fragmentos de kb/*.jsonl, procedimientos de
// kb/procedimientos/**/*.yml y conceptos de kb/conceptos/*.yml. Todo se lee del disco en cada corrida: el
// arnés no guarda copia de los pasos, así que el dataset solo nombra ids y se valida contra lo vigente.

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// ---------------------------------------------------------------------------------------------
// Fragmentos

// Fragmento de kb/*.jsonl. La clave única es (id, manual): ver «ids ambiguos» en ESQUEMA.md.
type Fragmento struct {
	ID      string
	Manual  string
	Titulo  string
	Seccion string
	Pagina  int
	Cita    string // la misma que arma el indexador de V1 (titulo + «, p. N» o «, <desde>»)
	Texto   string // texto + pasos[].texto (incluye el OCR de los fragmentos imagen)
	Imagen  string // ruta de la imagen si es un fragmento imagen
	Fotos   []string
}

// Clave única de un fragmento.
func (f *Fragmento) Clave() string { return claveFragmento(f.ID, f.Manual) }

func claveFragmento(id, manual string) string { return id + "@" + manual }

type fragmentoJSON struct {
	ID        string          `json:"id"`
	Documento string          `json:"documento"`
	Manual    string          `json:"manual"`
	Titulo    string          `json:"titulo"`
	Video     string          `json:"video"`
	Desde     string          `json:"desde"`
	Seccion   string          `json:"seccion"`
	Pagina    json.RawMessage `json:"pagina"`
	Texto     string          `json:"texto"`
	Imagen    string          `json:"imagen"`
	Pasos     []struct {
		Texto string   `json:"texto"`
		Fotos []string `json:"fotos"`
	} `json:"pasos"`
}

// ---------------------------------------------------------------------------------------------
// Procedimientos y conceptos (claves del YAML en las etiquetas json: se decodifican desde yamlMinimo)

type FotoYAML struct {
	ID      string `json:"id"`
	Ruta    string `json:"ruta_o_url"`
	Caption string `json:"caption"`
}

type PasoYAML struct {
	N      int        `json:"n"`
	ID     string     `json:"id"`
	Accion string     `json:"accion"`
	Fuente []string   `json:"fuente"`
	Fotos  []FotoYAML `json:"fotos"`
	Sub    []PasoYAML `json:"sub"`
}

type FuenteYAML struct {
	ID      string `json:"id"`
	Manual  string `json:"manual"`
	Seccion string `json:"seccion"`
}

type ErrorYAML struct {
	Sintoma  string   `json:"sintoma"`
	Solucion string   `json:"solucion"`
	Fuente   []string `json:"fuente"`
}

type Procedimiento struct {
	ID        string       `json:"id"`
	Modulo    string       `json:"modulo"`
	Titulo    string       `json:"titulo"`
	Preguntas []string     `json:"preguntas"`
	Aliases   []string     `json:"aliases"`
	Pasos     []PasoYAML   `json:"pasos"`
	Errores   []ErrorYAML  `json:"errores_frecuentes"`
	Fuentes   []FuenteYAML `json:"fuentes"`
	Archivo   string       `json:"-"`
}

type Concepto struct {
	ID         string   `json:"id"`
	Termino    string   `json:"termino"`
	Sinonimos  []string `json:"sinonimos"`
	Definicion string   `json:"definicion"`
	Fuente     []string `json:"fuente"`
	// Algunos glosarios citan con la tabla de fuentes de los procedimientos.
	Fuentes []FuenteYAML `json:"fuentes"`
	Archivo string       `json:"-"`
	// kb/conceptos no trae tabla `fuentes`: cada id lleva un comentario con el título del fragmento
	// («- web-…-s045  # Manual de Presupuestos › 3.5.2.5 …»), que desambigua los ids repetidos.
	TituloFuente map[string]string `json:"-"`
}

// Base reúne todo lo que el arnés necesita para calificar.
type Base struct {
	Raiz           string
	porClave       map[string]*Fragmento
	porID          map[string][]*Fragmento
	porCita        map[string][]*Fragmento
	Procedimientos map[string]*Procedimiento
	Conceptos      map[string]*Concepto
	Avisos         []string
}

func nuevaBase(raiz string) *Base {
	return &Base{
		Raiz:           raiz,
		porClave:       map[string]*Fragmento{},
		porID:          map[string][]*Fragmento{},
		porCita:        map[string][]*Fragmento{},
		Procedimientos: map[string]*Procedimiento{},
		Conceptos:      map[string]*Concepto{},
	}
}

// agregarFragmento indexa un fragmento por (id, manual), por id y por cita. Un duplicado se ignora.
func (b *Base) agregarFragmento(fr *Fragmento) {
	if _, ya := b.porClave[fr.Clave()]; ya {
		return // el indexador también descarta el duplicado idéntico
	}
	b.porClave[fr.Clave()] = fr
	b.porID[fr.ID] = append(b.porID[fr.ID], fr)
	b.porCita[normalizarCita(fr.Cita)] = append(b.porCita[normalizarCita(fr.Cita)], fr)
}

func cargarBase(raiz string) (*Base, error) {
	b := nuevaBase(raiz)
	archivos, _ := filepath.Glob(filepath.Join(raiz, "kb", "fragmentos*.jsonl"))
	if len(archivos) == 0 {
		return nil, fmt.Errorf("no hay kb/fragmentos*.jsonl en %s", raiz)
	}
	sort.Strings(archivos)
	for _, a := range archivos {
		if err := b.cargarJSONL(a); err != nil {
			return nil, err
		}
	}
	b.cargarProcedimientos(filepath.Join(raiz, "kb", "procedimientos"))
	b.cargarConceptos(filepath.Join(raiz, "kb", "conceptos"))
	return b, nil
}

func (b *Base) cargarJSONL(ruta string) error {
	f, err := os.Open(ruta)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for sc.Scan() {
		linea := strings.TrimSpace(sc.Text())
		if linea == "" {
			continue
		}
		var j fragmentoJSON
		if err := json.Unmarshal([]byte(linea), &j); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(ruta), err)
		}
		fr := &Fragmento{ID: strings.TrimSpace(j.ID), Manual: j.Manual, Seccion: j.Seccion, Imagen: j.Imagen}
		fmt.Sscan(string(j.Pagina), &fr.Pagina)
		// Igual que internal/indexar/kb_jsonl.go: titulo, si no video, si no documento.
		fr.Titulo = strings.TrimSpace(j.Titulo)
		if fr.Titulo == "" {
			fr.Titulo = strings.TrimSpace(j.Video)
		}
		if fr.Titulo == "" {
			fr.Titulo = j.Documento
		}
		fr.Cita = fr.Titulo
		if j.Desde != "" {
			fr.Cita += ", " + j.Desde
		} else if fr.Pagina > 0 {
			fr.Cita += fmt.Sprintf(", p. %d", fr.Pagina)
		}
		var t strings.Builder
		t.WriteString(j.Texto)
		for _, p := range j.Pasos {
			t.WriteString("\n" + p.Texto)
			fr.Fotos = append(fr.Fotos, p.Fotos...)
		}
		if fr.Imagen != "" {
			fr.Fotos = append(fr.Fotos, fr.Imagen)
		}
		fr.Texto = t.String()
		b.agregarFragmento(fr)
	}
	return sc.Err()
}

func normalizarCita(c string) string { return strings.Join(strings.Fields(normalizar(c)), " ") }

// cargarProcedimientos lee <dir>/<modulo>/<id>.yml. Las carpetas que empiezan por «_» o «.» (p. ej.
// `_reserva`, plantillas retiradas) no son parte del conjunto vigente y se saltan.
func (b *Base) cargarProcedimientos(dir string) {
	filepath.WalkDir(dir, func(ruta string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() && ruta != dir && (strings.HasPrefix(d.Name(), "_") || strings.HasPrefix(d.Name(), ".")) {
			return filepath.SkipDir
		}
		if err != nil || d.IsDir() || !(strings.HasSuffix(ruta, ".yml") || strings.HasSuffix(ruta, ".yaml")) {
			return nil
		}
		m, err := cargarYAML(ruta, b.Raiz)
		if err != nil {
			b.Avisos = append(b.Avisos, "procedimiento ilegible: "+err.Error())
			return nil
		}
		var p Procedimiento
		if err := decodificar(m, &p); err != nil || p.ID == "" {
			b.Avisos = append(b.Avisos, fmt.Sprintf("procedimiento sin forma esperada: %s (%v)", filepath.Base(ruta), err))
			return nil
		}
		p.Archivo = ruta
		b.Procedimientos[p.ID] = &p
		return nil
	})
}

func (b *Base) cargarConceptos(dir string) {
	filepath.WalkDir(dir, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !(strings.HasSuffix(ruta, ".yml") || strings.HasSuffix(ruta, ".yaml")) {
			return nil
		}
		m, err := cargarYAML(ruta, b.Raiz)
		if err != nil {
			b.Avisos = append(b.Avisos, "concepto ilegible: "+err.Error())
			return nil
		}
		// Un archivo puede traer un concepto o una lista «conceptos: [...]».
		var lista []map[string]any
		if l, ok := m["conceptos"].([]any); ok {
			for _, x := range l {
				if mm, ok := x.(map[string]any); ok {
					lista = append(lista, mm)
				}
			}
		} else {
			lista = append(lista, m)
		}
		titulos := titulosEnComentarios(ruta)
		for _, mm := range lista {
			var c Concepto
			if err := decodificar(mm, &c); err != nil || c.ID == "" {
				continue
			}
			c.Archivo = ruta
			c.TituloFuente = titulos
			b.Conceptos[c.ID] = &c
		}
		return nil
	})
}

// Resolver un id del dataset o de una cita del plan: «id@manual» o «id» a secas (todas las variantes).
func (b *Base) Resolver(ref string) []*Fragmento {
	if f, ok := b.porClave[ref]; ok {
		return []*Fragmento{f}
	}
	if strings.Contains(ref, "@") {
		return nil
	}
	return b.porID[ref]
}

// ResolverEn resuelve un id con la tabla «fuentes» de un procedimiento (par id, manual).
func (b *Base) ResolverEn(id string, tabla []FuenteYAML) []*Fragmento {
	for _, f := range tabla {
		if f.ID == id && f.Manual != "" {
			if fr, ok := b.porClave[claveFragmento(id, f.Manual)]; ok {
				return []*Fragmento{fr}
			}
		}
	}
	return b.Resolver(id)
}

// MaxFragmentosPorCita: una cita que el indexador de V1 da a más fragmentos que esto («S10 Conocimiento» la
// llevan 156 nodos de Cortex) no identifica nada: se trata como no resoluble. Si no, su texto sumado respaldaría
// casi cualquier paso y la detección de invenciones dejaría de medir.
const MaxFragmentosPorCita = 3

// PorCita: los fragmentos de la KB que el indexador de V1 cita así (vacío si la cita es demasiado genérica).
func (b *Base) PorCita(cita string) []*Fragmento {
	frs := b.porCita[normalizarCita(cita)]
	if len(frs) > MaxFragmentosPorCita {
		return nil
	}
	return frs
}

// FuentesProcedimiento: claves de todos los fragmentos de la tabla «fuentes».
func (b *Base) FuentesProcedimiento(p *Procedimiento) map[string]bool {
	out := map[string]bool{}
	for _, f := range p.Fuentes {
		for _, fr := range b.ResolverEn(f.ID, p.Fuentes) {
			out[fr.Clave()] = true
		}
	}
	return out
}

var reComentarioFuente = regexp.MustCompile(`^\s*-\s*"?([^"#\s]+)"?\s+#\s*(.+?)\s*$`)

// titulosEnComentarios lee «- <id>  # <título>» de un YAML de conceptos.
func titulosEnComentarios(ruta string) map[string]string {
	out := map[string]string{}
	b, err := os.ReadFile(ruta)
	if err != nil {
		return out
	}
	for _, l := range strings.Split(string(b), "\n") {
		if m := reComentarioFuente.FindStringSubmatch(l); m != nil {
			out[m[1]] = m[2]
		}
	}
	return out
}

// ResolverConcepto: un id de `fuente` de un concepto, desambiguado por el título de su comentario.
func (b *Base) ResolverConcepto(id string, c *Concepto) []*Fragmento {
	frs := b.ResolverEn(id, c.Fuentes)
	if len(frs) > 1 {
		if t := normalizarCita(c.TituloFuente[id]); t != "" {
			var elegidos []*Fragmento
			for _, fr := range frs {
				if normalizarCita(fr.Titulo) == t || strings.HasPrefix(t, normalizarCita(fr.Manual)) {
					elegidos = append(elegidos, fr)
				}
			}
			// Primero el título exacto; si no, el manual del comentario.
			for _, fr := range elegidos {
				if normalizarCita(fr.Titulo) == t {
					return []*Fragmento{fr}
				}
			}
			if len(elegidos) > 0 {
				return elegidos
			}
		}
	}
	return frs
}

// FuentesConcepto: claves de los fragmentos que cita un concepto.
func (b *Base) FuentesConcepto(c *Concepto) map[string]bool {
	out := map[string]bool{}
	for _, id := range c.Fuente {
		for _, fr := range b.ResolverConcepto(id, c) {
			out[fr.Clave()] = true
		}
	}
	for _, f := range c.Fuentes {
		for _, fr := range b.ResolverEn(f.ID, c.Fuentes) {
			out[fr.Clave()] = true
		}
	}
	return out
}

// BuscarConcepto por id, término o sinónimo (normalizados).
func (b *Base) BuscarConcepto(ref string) *Concepto {
	if ref == "" {
		return nil
	}
	if c, ok := b.Conceptos[ref]; ok {
		return c
	}
	n := normalizar(ref)
	for _, c := range b.Conceptos {
		if normalizar(c.ID) == n || normalizar(c.Termino) == n || strings.HasSuffix(normalizar(c.ID), "."+n) {
			return c
		}
		for _, s := range c.Sinonimos {
			if normalizar(s) == n {
				return c
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------------------------
// Pasos de un procedimiento como unidades comparables

// UnidadPaso es un paso (o subpaso) con el paso de primer nivel al que pertenece.
type UnidadPaso struct {
	Nivel1 int    // n del paso de primer nivel
	ID     string // id del YAML (#12 o #12.3)
	Accion string
	Fuente []string
	Fotos  []string // claves de foto (sha1(ruta)[:12])
}

func unidades(p *Procedimiento) []UnidadPaso {
	var out []UnidadPaso
	var rec func(ps []PasoYAML, nivel1 int)
	rec = func(ps []PasoYAML, nivel1 int) {
		for _, x := range ps {
			n1 := nivel1
			if n1 == 0 {
				n1 = x.N
			}
			u := UnidadPaso{Nivel1: n1, ID: x.ID, Accion: x.Accion, Fuente: x.Fuente}
			for _, f := range x.Fotos {
				u.Fotos = append(u.Fotos, claveFoto(f.Ruta, f.ID))
			}
			out = append(out, u)
			rec(x.Sub, n1)
		}
	}
	rec(p.Pasos, 0)
	return out
}

// pasosNivel1 devuelve los n de primer nivel, en orden.
func pasosNivel1(p *Procedimiento) []int {
	var out []int
	for _, x := range p.Pasos {
		out = append(out, x.N)
	}
	return out
}

// fotosPorPaso: n de primer nivel → claves de foto (incluye las de sus subpasos).
func fotosPorPaso(p *Procedimiento) map[int][]string {
	out := map[int][]string{}
	for _, u := range unidades(p) {
		out[u.Nivel1] = append(out[u.Nivel1], u.Fotos...)
	}
	return out
}

// nivel1DeID: «proc#12.3» → 12; «proc#4» → 4; sin «#» → 0.
func nivel1DeID(id string) int {
	_, s, ok := strings.Cut(id, "#")
	if !ok {
		return 0
	}
	s, _, _ = strings.Cut(s, ".")
	n, _ := strconv.Atoi(s)
	return n
}

// claveFoto: el id de ESQUEMA.md, sha1(ruta)[:12]; si no hay ruta, el id que venga.
func claveFoto(ruta, id string) string {
	ruta = strings.TrimSpace(ruta)
	ruta = strings.TrimPrefix(ruta, "/fotos/")
	ruta = strings.TrimPrefix(ruta, "fotos/")
	if ruta == "" {
		return strings.TrimSpace(id)
	}
	h := sha1.Sum([]byte(ruta))
	return hex.EncodeToString(h[:])[:12]
}

// ---------------------------------------------------------------------------------------------
// Texto: normalización, raíces de contenido y soporte léxico (mismo método que el validador)

// LargoRaiz, UmbralMinRaices, UmbralFraccion y las palabras vacías copian herramientas/validar_procedimientos.py
// (§ «Soporte léxico» de ESQUEMA.md; valores vigentes el 06-10-2026: 2 raíces y 60 %).
const (
	LargoRaiz       = 5
	UmbralMinRaices = 2
	UmbralFraccion  = 0.6
)

var palabrasVacias = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`a al algo algun alguna algunas alguno algunos ante antes asi aun bajo bien cada como con contra cual cuales cuando de del desde donde dos e el ella ellas ello ellos en entre era es esa esas ese eso esos esta estas este esto estos fue ha hace hacia han hasta hay la las le les lo los mas me mediante mismo muy ni no nos o otra otras otro otros para pero poco por porque pues que quien se segun ser si sin sino sobre solo son su sus tal tambien tan tanto te tiene tienen todo todos tras tu un una unas uno unos usted y ya ademas luego despues primero ahora cual cuya cuyo debe deben dentro puede pueden sera seran esta estan ese caso vez haga clic doble derecho izquierdo boton botones use utilice pulse presione elija seleccione ventana opcion opciones menu sistema muestra mostrara aparece pantalla escenario modulo s10`) {
		palabrasVacias[w] = true
	}
}

var sinMarcasNorm = strings.NewReplacer("“", "\"", "”", "\"", "’", "'", "‘", "'")

// normalizar: minúsculas, sin tildes, comillas tipográficas simples, espacios colapsados.
func normalizar(s string) string {
	s = norm.NFKD.String(strings.ToLower(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.Join(strings.Fields(sinMarcasNorm.Replace(b.String())), " ")
}

var reNegrita = regexp.MustCompile(`\*\*([^*\n]{1,80})\*\*`)

func tokens(s string) []string {
	return strings.FieldsFunc(normalizar(s), func(r rune) bool { return !(unicode.IsLetter(r) || unicode.IsDigit(r)) })
}

// raices: palabras de contenido (≥ 4 letras, sin vacías ni números) reducidas a 5 letras. `extra` añade
// palabras vacías (enteras, antes de recortar).
func raices(s string, extra ...map[string]bool) map[string]bool {
	out := map[string]bool{}
	for _, t := range tokens(strings.ReplaceAll(s, "**", " ")) {
		if len([]rune(t)) < 4 || palabrasVacias[t] || esNumero(t) {
			continue
		}
		vacia := false
		for _, e := range extra {
			vacia = vacia || e[t]
		}
		if vacia {
			continue
		}
		r := []rune(t)
		if len(r) > LargoRaiz {
			r = r[:LargoRaiz]
		}
		out[string(r)] = true
	}
	return out
}

func esNumero(t string) bool {
	for _, r := range t {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// soporteLexico: el criterio del validador. Pasa si comparte al menos 2 raíces (o todas, si tiene menos)
// y al menos el 60 % de las suyas con el texto de referencia.
func soporteLexico(texto string, referencia map[string]bool) bool {
	r := raicesPaso(texto)
	if len(r) == 0 {
		return true // nada que respaldar («Pulse Aceptar.» se juzga por sus negritas)
	}
	comunes := 0
	for x := range r {
		if referencia[x] {
			comunes++
		}
	}
	return comunes >= min(UmbralMinRaices, len(r)) && float64(comunes) >= UmbralFraccion*float64(len(r))
}

// negritas devuelve los términos entre ** ** normalizados.
func negritas(s string) []string {
	var out []string
	for _, m := range reNegrita.FindAllStringSubmatch(s, -1) {
		t := normalizar(strings.Trim(m[1], " :.,;"))
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// contieneFrase: la frase normalizada aparece en el texto normalizado como palabras completas.
func contieneFrase(texto, frase string) bool {
	t := " " + strings.Join(tokens(texto), " ") + " "
	f := strings.Join(tokens(frase), " ")
	if f == "" {
		return false
	}
	return strings.Contains(t, " "+f+" ")
}
