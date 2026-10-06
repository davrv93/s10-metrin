package conocimiento

import (
	"sort"
	"strings"

	"rag-go/internal/v2/tipos"
)

// Alias es una forma de nombrar algo del ERP y aquello a lo que remite. Sale de los YAML: aliases y
// entidades de los procedimientos, término y sinónimos del glosario, nombre del módulo.
type Alias struct {
	Forma    string `json:"forma"`    // tal como está en el YAML
	Clase    string `json:"clase"`    // procedimiento | concepto | entidad | modulo
	ID       string `json:"id"`       // procedimiento o concepto al que remite ("" en entidades sueltas)
	Canonico string `json:"canonico"` // título del procedimiento, término del concepto o nombre de la entidad
	Modulo   string `json:"modulo,omitempty"`
	plegada  string
}

// Detectado: lo que la reescritura de la consulta encontró en el texto.
type Detectado struct {
	Aliases   []string `json:"aliases,omitempty"`   // canónicos y aliases hermanos de lo detectado
	Entidades []string `json:"entidades,omitempty"` // pantallas, objetos y términos nombrados
	Modulo    string   `json:"modulo,omitempty"`    // si todo lo detectado es de un mismo módulo
	Remite    []Alias  `json:"remite,omitempty"`    // cada coincidencia, con su procedimiento o concepto
}

// Aliases devuelve todos los aliases conocidos (de los más largos a los más cortos).
func (b *Base) Aliases() []Alias {
	if b.alias == nil {
		return nil
	}
	return append([]Alias(nil), b.alias...)
}

func (b *Base) armarAliases() {
	var out []Alias
	add := func(forma, clase, id, canon, modulo string) {
		f := plegar(strings.TrimSpace(forma))
		if len([]rune(f)) < 3 {
			return
		}
		out = append(out, Alias{Forma: forma, Clase: clase, ID: id, Canonico: canon, Modulo: modulo, plegada: f})
	}
	for _, p := range b.Procedimientos {
		add(p.Titulo, ClaseProcedimiento, p.ID, p.Titulo, p.Modulo)
		for _, a := range p.Aliases {
			add(a, ClaseProcedimiento, p.ID, p.Titulo, p.Modulo)
		}
		for _, e := range append(append([]string{}, p.Entidades.Pantallas...), p.Entidades.Objetos...) {
			add(e, "entidad", "", e, p.Modulo)
		}
		if p.Entidades.Modulo != "" {
			add(p.Entidades.Modulo, "modulo", "", p.Entidades.Modulo, p.Modulo)
		}
	}
	for _, c := range b.Conceptos {
		for _, n := range c.Nombres() {
			add(n, ClaseConcepto, c.ID, c.Termino, c.Modulo)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return len(out[i].plegada) > len(out[j].plegada) })
	b.alias = out
}

// Detectar busca aliases y entidades en el texto (frase completa, sin tildes ni mayúsculas). Una forma
// contenida en otra ya detectada («presupuesto» dentro de «presupuesto meta») no se cuenta aparte.
func (b *Base) Detectar(texto string) Detectado {
	t := plegar(texto)
	var d Detectado
	var tramos [][2]int
	modulos := map[string]bool{}
	for _, a := range b.alias {
		i := indiceFrase(t, a.plegada)
		if i < 0 {
			continue
		}
		dentro := false
		for _, tr := range tramos {
			if i >= tr[0] && i+len(a.plegada) <= tr[1] && len(a.plegada) < tr[1]-tr[0] {
				dentro = true
			}
		}
		if dentro {
			continue
		}
		tramos = append(tramos, [2]int{i, i + len(a.plegada)})
		d.Remite = append(d.Remite, a)
		switch a.Clase {
		case ClaseProcedimiento:
			// La tarea: su título y sus otras formas (hasta 4) para expandir la búsqueda.
			d.Aliases = append(d.Aliases, a.Canonico)
			if p, ok := b.porProc[a.ID]; ok {
				for i, x := range p.Aliases {
					if i >= 4 {
						break
					}
					d.Aliases = append(d.Aliases, x)
				}
			}
		case ClaseConcepto:
			// El término como se escribe en S10, y sus sinónimos para expandir.
			d.Entidades = append(d.Entidades, a.Canonico)
			if c, ok := b.porConcepto[a.ID]; ok {
				d.Aliases = append(d.Aliases, c.Sinonimos...)
			}
		default:
			d.Entidades = append(d.Entidades, a.Canonico)
		}
		if a.Modulo != "" {
			modulos[a.Modulo] = true
		}
	}
	d.Aliases = unicos(d.Aliases)
	d.Entidades = unicos(d.Entidades)
	if len(modulos) == 1 {
		for m := range modulos {
			d.Modulo = m
		}
	}
	return d
}

func indiceFrase(t, frase string) int {
	for i := 0; ; {
		j := strings.Index(t[i:], frase)
		if j < 0 {
			return -1
		}
		j += i
		fin := j + len(frase)
		if (j == 0 || !esLetra(rune(t[j-1]))) && (fin >= len(t) || !esLetra(rune(t[fin]))) {
			return j
		}
		i = j + 1
		if i >= len(t) {
			return -1
		}
	}
}

// Expandir completa la consulta con lo detectado (aliases, entidades, módulo). Nunca toca Original ni
// Normalizada: los términos exactos del usuario no se pierden.
func (b *Base) Expandir(c tipos.Consulta) tipos.Consulta {
	texto := c.Original
	if c.Normalizada != "" {
		texto += " \n " + c.Normalizada
	}
	d := b.Detectar(texto)
	c.Aliases = unicos(append(c.Aliases, d.Aliases...))
	c.Entidades = unicos(append(c.Entidades, d.Entidades...))
	if c.Modulo == "" {
		c.Modulo = d.Modulo
	}
	return c
}

// formasNombradas: las formas distintas (término o sinónimos) de un concepto que aparecen en el texto.
// Una aparición dentro de otra forma más larga («meta» dentro de «presupuesto meta») no cuenta aparte.
func (b *Base) formasNombradas(c *Concepto, texto string) []string {
	t := plegar(texto)
	type tramo struct{ i, f int }
	apariciones := map[string][]tramo{}
	for _, n := range c.Nombres() {
		p := plegar(n)
		for desde := 0; desde < len(t); {
			i := indiceFrase(t[desde:], p)
			if i < 0 {
				break
			}
			apariciones[n] = append(apariciones[n], tramo{desde + i, desde + i + len(p)})
			desde += i + len(p)
		}
	}
	var out []string
	for _, n := range c.Nombres() {
		for _, a := range apariciones[n] {
			dentro := false
			for m, ts := range apariciones {
				for _, x := range ts {
					if m != n && x.i <= a.i && a.f <= x.f && x.f-x.i > a.f-a.i {
						dentro = true
					}
				}
			}
			if !dentro {
				out = append(out, n)
				break
			}
		}
	}
	return out
}

// Sinonimos: forma plegada → canónico, para quien prefiera un diccionario (reescritura del núcleo).
func (b *Base) Sinonimos() map[string]string {
	out := map[string]string{}
	for _, a := range b.alias {
		if _, ok := out[a.plegada]; !ok {
			out[a.plegada] = a.Canonico
		}
	}
	return out
}
