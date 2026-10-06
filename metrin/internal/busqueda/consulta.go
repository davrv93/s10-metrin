package busqueda

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Consulta es la pregunta reescrita para buscar. La original no se pierde
// nunca: Aliases solo SUMA términos (con peso menor en BM25) y los términos
// técnicos detectados quedan en Entidades.
type Consulta struct {
	Original    string   `json:"original"`
	Normalizada string   `json:"normalizada"` // clave de caché y texto que va al embebedor
	Entidades   []string `json:"entidades"`   // atajos, códigos, frases entre comillas y términos del glosario
	Aliases     []string `json:"aliases"`     // expansiones añadidas
}

// Alias es el diccionario de expansión. Sale de archivos de datos, no del
// código: eval/alias.yml (acciones y términos), y si existen,
// kb/procedimientos/**/*.yml (aliases y preguntas) y kb/conceptos/*.yml
// (sinónimos).
type Alias struct {
	// grupos de verbos de acción equivalentes («registrar», «ingresar»,
	// «cargar»…); cada miembro puede ser de varias palabras.
	acciones [][]string
	accionDe map[string][]int // raíz de una palabra → grupos
	// términos técnicos: forma canónica + sinónimos.
	terminos []terminoAlias
	// procedimientos: título + aliases, disparados por parecido con sus
	// preguntas.
	procs []procAlias
	// UmbralPregunta: parecido mínimo (proporción de raíces de la pregunta
	// del procedimiento presentes en la consulta, y viceversa, la media) para
	// añadir el título del procedimiento como alias.
	UmbralPregunta float64
	// Origenes: de dónde salió cada parte (para la traza).
	Origenes []string
	vistos   map[string]bool // términos ya registrados como entidad
}

type terminoAlias struct {
	canonico string
	formas   [][]string // cada forma como secuencia de raíces
	textos   []string   // formas en texto (para devolverlas como alias)
}

type procAlias struct {
	titulo    string
	aliases   []string
	preguntas []map[string]bool // raíces de cada pregunta
}

// NuevoAlias crea un diccionario vacío.
func NuevoAlias() *Alias {
	return &Alias{accionDe: map[string][]int{}, UmbralPregunta: 0.75}
}

// CargarAlias lee eval/alias.yml (o cualquier archivo con el mismo formato):
//
//	acciones:
//	  - [registrar, ingresar, cargar, "dar de alta"]
//	terminos:
//	  - {termino: metrado, sinonimos: [metrados, "cantidad de obra"]}
func (a *Alias) CargarAlias(ruta string) error {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return err
	}
	v, err := LeerYAML(b)
	if err != nil {
		return fmt.Errorf("%s: %w", ruta, err)
	}
	m := yamlMapa(v)
	for _, g := range yamlLista(m["acciones"]) {
		a.AgregarAcciones(yamlTextos(g)...)
	}
	for _, t := range yamlLista(m["terminos"]) {
		tm := yamlMapa(t)
		a.AgregarTermino(yamlTexto(tm["termino"]), yamlTextos(tm["sinonimos"])...)
	}
	a.Origenes = append(a.Origenes, ruta)
	return nil
}

// AgregarAcciones añade un grupo de verbos equivalentes.
func (a *Alias) AgregarAcciones(miembros ...string) {
	if len(miembros) < 2 {
		return
	}
	g := len(a.acciones)
	a.acciones = append(a.acciones, miembros)
	for _, mm := range miembros {
		raices, _ := Terminos(mm)
		if len(raices) == 1 { // los de varias palabras solo se añaden, no disparan
			a.accionDe[raices[0]] = append(a.accionDe[raices[0]], g)
		}
	}
}

// agregarEntidad registra un término sin sinónimos una sola vez.
func (a *Alias) agregarEntidad(t string) {
	k := Plegar(strings.TrimSpace(t))
	if k == "" {
		return
	}
	if a.vistos == nil {
		a.vistos = map[string]bool{}
	}
	if a.vistos[k] {
		return
	}
	a.vistos[k] = true
	a.AgregarTermino(t)
}

// AgregarTermino añade un término técnico con sus sinónimos.
func (a *Alias) AgregarTermino(canonico string, sinonimos ...string) {
	canonico = strings.TrimSpace(canonico)
	if canonico == "" {
		return
	}
	if a.vistos == nil {
		a.vistos = map[string]bool{}
	}
	a.vistos[Plegar(canonico)] = true
	t := terminoAlias{canonico: canonico}
	for _, f := range append([]string{canonico}, sinonimos...) {
		raices, _ := Terminos(f)
		if len(raices) == 0 {
			continue
		}
		t.formas = append(t.formas, raices)
		t.textos = append(t.textos, f)
	}
	if len(t.formas) > 0 {
		a.terminos = append(a.terminos, t)
	}
}

// AgregarProcedimientos lee aliases y preguntas de kb/procedimientos/**/*.yml.
// Un directorio inexistente no es error.
func (a *Alias) AgregarProcedimientos(dir string) error {
	docs, err := CargarProcedimientos(dir)
	if err != nil || len(docs) == 0 {
		return err
	}
	for _, d := range docs {
		p := procAlias{titulo: d.Meta["titulo"]}
		// CargarProcedimientos junta preguntas y aliases en un campo; se
		// vuelven a leer por separado para distinguirlos.
		b, err := os.ReadFile(d.Meta["ruta"])
		if err != nil {
			return err
		}
		v, _ := LeerYAML(b)
		m := yamlMapa(v)
		p.aliases = yamlTextos(m["aliases"])
		// Pantallas y objetos de `entidades`: términos del ERP que, si
		// aparecen en la consulta, quedan como entidades (sin alias).
		ent := yamlMapa(m["entidades"])
		for _, t := range append(yamlTextos(ent["pantallas"]), yamlTextos(ent["objetos"])...) {
			a.agregarEntidad(t)
		}
		for _, q := range yamlTextos(m["preguntas"]) {
			if s := conjuntoRaices(q); len(s) > 0 {
				p.preguntas = append(p.preguntas, s)
			}
		}
		if p.titulo != "" && len(p.preguntas) > 0 {
			a.procs = append(a.procs, p)
		}
	}
	a.Origenes = append(a.Origenes, dir)
	return nil
}

// AgregarConceptos lee kb/conceptos/*.yml: cada archivo es un concepto con
// su nombre (termino, nombre, concepto o titulo) y sinonimos/aliases. Un
// directorio inexistente no es error.
func (a *Alias) AgregarConceptos(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	var n int
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !(strings.HasSuffix(p, ".yml") || strings.HasSuffix(p, ".yaml")) {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		v, err := LeerYAML(b)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		for _, e := range yamlLista(v) {
			m := yamlMapa(e)
			if m == nil {
				continue
			}
			nombre := ""
			for _, k := range []string{"termino", "nombre", "concepto", "titulo"} {
				if s := yamlTexto(m[k]); s != "" {
					nombre = s
					break
				}
			}
			sin := append(yamlTextos(m["sinonimos"]), yamlTextos(m["aliases"])...)
			if nombre != "" {
				a.AgregarTermino(nombre, sin...)
				n++
			}
		}
		return nil
	})
	if err == nil && n > 0 {
		a.Origenes = append(a.Origenes, dir)
	}
	return err
}

func conjuntoRaices(s string) map[string]bool {
	raices, _ := Terminos(s)
	m := map[string]bool{}
	for _, r := range raices {
		m[r] = true
	}
	return m
}

var reComillas = regexp.MustCompile(`["«“]([^"»”]{2,80})["»”]`)

// Reescribir arma la Consulta. Con a == nil solo normaliza y detecta
// entidades sintácticas (atajos, códigos, comillas).
func (a *Alias) Reescribir(q string) Consulta {
	c := Consulta{Original: q, Normalizada: Normalizar(q)}
	toks := Tokenizar(c.Normalizada)
	var raices []string
	presentes := map[string]bool{}
	for _, t := range toks {
		if t.Especial {
			c.Entidades = append(c.Entidades, t.Texto)
		}
		if !t.Especial && vacias[t.Texto] {
			continue
		}
		r := t.Texto
		if !t.Especial {
			r = Raiz(t.Texto)
		}
		raices = append(raices, r)
		presentes[r] = true
	}
	for _, m := range reComillas.FindAllStringSubmatch(q, -1) {
		c.Entidades = append(c.Entidades, strings.TrimSpace(m[1]))
	}
	if a == nil {
		c.Entidades = unicos(c.Entidades)
		return c
	}
	var aliases []string
	// Términos técnicos: si aparece cualquiera de sus formas, el canónico es
	// entidad y las demás formas, alias.
	for _, t := range a.terminos {
		hallada := -1
		for i, f := range t.formas {
			if contieneSecuencia(raices, f) {
				hallada = i
				break
			}
		}
		if hallada < 0 {
			continue
		}
		c.Entidades = append(c.Entidades, t.canonico)
		for i, txt := range t.textos {
			if i != hallada {
				aliases = append(aliases, txt)
			}
		}
	}
	// Verbos de acción: se añaden los equivalentes del grupo.
	for _, r := range raices {
		for _, g := range a.accionDe[r] {
			aliases = append(aliases, a.acciones[g]...)
		}
	}
	// Procedimientos: si la consulta se parece a una de sus preguntas, su
	// título (y sus aliases) entran como alias.
	if len(presentes) > 0 {
		for _, p := range a.procs {
			if mejorParecido(presentes, p.preguntas) >= a.UmbralPregunta {
				aliases = append(aliases, p.titulo)
				aliases = append(aliases, p.aliases...)
			}
		}
	}
	// Fuera los alias que no añaden ninguna raíz nueva.
	var finales []string
	vistos := map[string]bool{}
	for _, al := range aliases {
		k := Plegar(strings.TrimSpace(al))
		if k == "" || vistos[k] {
			continue
		}
		vistos[k] = true
		rs, _ := Terminos(al)
		nueva := false
		for _, r := range rs {
			if !presentes[r] {
				nueva = true
				break
			}
		}
		if nueva {
			finales = append(finales, al)
		}
	}
	sort.Strings(finales)
	c.Aliases = finales
	c.Entidades = unicos(c.Entidades)
	return c
}

func contieneSecuencia(h, aguja []string) bool {
	if len(aguja) == 0 || len(aguja) > len(h) {
		return false
	}
	for i := 0; i+len(aguja) <= len(h); i++ {
		ok := true
		for j := range aguja {
			if h[i+j] != aguja[j] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// mejorParecido: el máximo, entre las preguntas, de la media de las dos
// coberturas (|A∩B|/|A| y |A∩B|/|B|). Con una sola cobertura, una consulta de
// una palabra disparaba cualquier procedimiento que la contuviera.
func mejorParecido(q map[string]bool, preguntas []map[string]bool) float64 {
	mejor := 0.0
	for _, p := range preguntas {
		if len(p) == 0 {
			continue
		}
		comun := 0
		for r := range p {
			if q[r] {
				comun++
			}
		}
		s := (float64(comun)/float64(len(p)) + float64(comun)/float64(len(q))) / 2
		mejor = max(mejor, s)
	}
	return mejor
}

// TerminosConsulta convierte la Consulta en términos ponderados: los de la
// original con peso 1 y los de los alias con pesoAlias (sin repetir raíces
// de la original).
func (ix *Indice) TerminosConsulta(c Consulta, pesoAlias float64) []TerminoConsulta {
	terms := ix.TerminosDe(c.Normalizada, 1)
	if pesoAlias <= 0 || len(c.Aliases) == 0 {
		return terms
	}
	ya := map[string]bool{}
	for _, t := range terms {
		ya[t.Termino] = true
	}
	for _, al := range c.Aliases {
		for _, t := range ix.TerminosDe(al, pesoAlias) {
			if !ya[t.Termino] {
				ya[t.Termino] = true
				terms = append(terms, t)
			}
		}
	}
	return terms
}
