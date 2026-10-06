package conocimiento

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"rag-go/internal/v2/tipos"
)

// Concepto es una entrada del glosario (kb/conceptos/*.yml) con su origen y sus citas.
type Concepto struct {
	tipos.ConceptoDef

	Archivo string `json:"archivo"`
	Linea   int    `json:"linea"`
	Modulo  string `json:"modulo,omitempty"`
	Notas   string `json:"notas,omitempty"`

	citas    []citaYAML              // como vienen en el YAML
	tabla    map[string]tipos.Fuente // id → (id, manual), armada por Base al resolver
	ambiguos map[string]bool
}

// Par devuelve el par (id, manual) con el que se resuelve una cita del concepto.
func (c *Concepto) Par(id string) (tipos.Fuente, bool) {
	if c.ambiguos[id] {
		return tipos.Fuente{}, false
	}
	f, ok := c.tabla[id]
	return f, ok
}

// Nombres: el término y sus sinónimos.
func (c *Concepto) Nombres() []string { return unicos(append([]string{c.Termino}, c.Sinonimos...)) }

// citaYAML: una cita del concepto. Hoy los YAML traen solo el id, con el manual en el comentario de
// la línea («- web-…-s045  # Manual de Presupuestos › 3.5.2.5 Ingreso de metrados…»); también se
// admite {id, manual} en línea o una tabla «fuentes» como la de los procedimientos.
type citaYAML struct {
	ID         string
	Manual     string // explícito ({id, manual}) o tomado del comentario
	Seccion    string
	DeComent   bool // el manual sale del comentario
	Linea      int
	Comentario string
}

type citasYAML []citaYAML

func (cs *citasYAML) UnmarshalYAML(n *yaml.Node) error {
	items := []*yaml.Node{n}
	if n.Kind == yaml.SequenceNode {
		items = n.Content
	}
	for _, it := range items {
		switch it.Kind {
		case yaml.ScalarNode:
			if it.Tag == "!!null" || strings.TrimSpace(it.Value) == "" {
				continue
			}
			c := citaYAML{ID: strings.TrimSpace(it.Value), Linea: it.Line, Comentario: it.LineComment}
			if m, s := manualDeComentario(it.LineComment); m != "" {
				c.Manual, c.Seccion, c.DeComent = m, s, true
			}
			*cs = append(*cs, c)
		case yaml.MappingNode:
			var f tipos.Fuente
			if err := it.Decode(&f); err != nil {
				return err
			}
			*cs = append(*cs, citaYAML{ID: f.ID, Manual: f.Manual, Seccion: f.Seccion, Linea: it.Line})
		default:
			return fmt.Errorf("línea %d: fuente: se esperaba un id o {id, manual}", it.Line)
		}
	}
	return nil
}

// manualDeComentario: «# Manual de Presupuestos › 3.5.2.5 Ingreso…» → («Manual de Presupuestos», «3.5.2.5 …»).
func manualDeComentario(c string) (manual, seccion string) {
	c = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(c), "#"))
	if c == "" {
		return "", ""
	}
	if i := strings.Index(c, "›"); i >= 0 {
		return strings.TrimSpace(c[:i]), strings.TrimSpace(c[i+len("›"):])
	}
	return "", ""
}

type conceptoYAML struct {
	ID                         string        `yaml:"id"`
	Termino                    string        `yaml:"termino"`
	Sinonimos                  []string      `yaml:"sinonimos"`
	Aliases                    []string      `yaml:"aliases"`
	Definicion                 string        `yaml:"definicion"`
	EnS10                      string        `yaml:"en_s10"`
	Donde                      string        `yaml:"donde"`
	Ejemplo                    string        `yaml:"ejemplo"`
	ProcedimientosRelacionados []string      `yaml:"procedimientos_relacionados"`
	Relacionados               []string      `yaml:"relacionados"`
	Fuente                     citasYAML     `yaml:"fuente"`
	Fuentes                    []fuenteYAML  `yaml:"fuentes"`
	Modulo                     string        `yaml:"modulo"`
	Entidades                  entidadesYAML `yaml:"entidades"`
	Notas                      string        `yaml:"notas"`
}

// LeerConceptos lee un archivo del glosario: un concepto (mapa), una lista de conceptos o un mapa con
// la clave «conceptos». Un concepto inválido se salta y se informa; los demás del archivo se cargan.
func LeerConceptos(ruta string, datos []byte) ([]*Concepto, []ErrorCarga) {
	nodo, e := leerNodoYAML(ruta, datos)
	if e != nil {
		return nil, []ErrorCarga{*e}
	}
	var items []*yaml.Node
	switch nodo.Kind {
	case yaml.SequenceNode:
		items = nodo.Content
	case yaml.MappingNode:
		items = []*yaml.Node{nodo}
		for i := 0; i+1 < len(nodo.Content); i += 2 {
			if nodo.Content[i].Value == "conceptos" && nodo.Content[i+1].Kind == yaml.SequenceNode {
				items = nodo.Content[i+1].Content
			}
		}
	default:
		return nil, []ErrorCarga{{Archivo: ruta, Linea: nodo.Line, Mensaje: "se esperaba un mapa o una lista de conceptos", Grave: true}}
	}
	base := strings.TrimSuffix(filepath.Base(ruta), filepath.Ext(ruta))
	var out []*Concepto
	var errs []ErrorCarga
	for _, it := range items {
		var y conceptoYAML
		if err := it.Decode(&y); err != nil {
			ec := errorYAML(ruta, err)
			if ec.Linea == 0 {
				ec.Linea = it.Line
			}
			errs = append(errs, ec)
			continue
		}
		if strings.TrimSpace(y.Termino) == "" || strings.TrimSpace(y.Definicion) == "" {
			errs = append(errs, ErrorCarga{Archivo: ruta, Linea: it.Line, Grave: true, Mensaje: "concepto sin «termino» o sin «definicion»"})
			continue
		}
		id := y.ID
		if id == "" {
			if len(items) == 1 {
				id = base
			} else {
				id = slug(y.Termino)
			}
		}
		enS10 := y.EnS10
		if enS10 == "" {
			enS10 = y.Donde
		}
		c := &Concepto{
			ConceptoDef: tipos.ConceptoDef{
				ID: id, Termino: strings.TrimSpace(y.Termino), Sinonimos: unicos(append(y.Sinonimos, y.Aliases...)),
				Definicion: strings.TrimSpace(y.Definicion), EnS10: strings.TrimSpace(enS10), Ejemplo: strings.TrimSpace(y.Ejemplo),
				ProcedimientosRelacionados: unicos(append(y.ProcedimientosRelacionados, y.Relacionados...)),
			},
			Archivo: ruta, Linea: it.Line, Modulo: firstNonEmpty(y.Modulo, y.Entidades.Modulo), Notas: y.Notas,
			citas: y.Fuente,
		}
		// Una tabla «fuentes» (como la de los procedimientos) manda sobre el comentario.
		tabla := map[string]citaYAML{}
		for _, f := range y.Fuentes {
			if f.ID != "" && f.Manual != "" {
				tabla[f.ID] = citaYAML{ID: f.ID, Manual: f.Manual, Seccion: f.Seccion, Linea: f.linea}
			}
		}
		for i, cy := range c.citas {
			if t, ok := tabla[cy.ID]; ok {
				c.citas[i].Manual, c.citas[i].Seccion, c.citas[i].DeComent = t.Manual, t.Seccion, false
			}
		}
		for _, cy := range c.citas {
			c.Fuente = append(c.Fuente, cy.ID)
		}
		c.Fuente = unicos(c.Fuente)
		if len(c.Fuente) == 0 {
			errs = append(errs, ErrorCarga{Archivo: ruta, Linea: it.Line, Mensaje: fmt.Sprintf("concepto %s sin «fuente»", id)})
		}
		out = append(out, c)
	}
	return out, errs
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			return x
		}
	}
	return ""
}

func slug(s string) string {
	var b strings.Builder
	guion := false
	for _, r := range plegar(s) {
		if esLetra(r) {
			b.WriteRune(r)
			guion = false
		} else if !guion && b.Len() > 0 {
			b.WriteByte('_')
			guion = true
		}
	}
	return strings.Trim(b.String(), "_")
}

// CargarConceptos lee kb/conceptos/*.yml (y .yaml). Ids repetidos: se queda el primero.
func CargarConceptos(dir string) ([]*Concepto, []ErrorCarga) {
	var rutas []string
	for _, pat := range []string{"*.yml", "*.yaml"} {
		m, _ := filepath.Glob(filepath.Join(dir, pat))
		rutas = append(rutas, m...)
	}
	sort.Strings(rutas)
	var out []*Concepto
	var errs []ErrorCarga
	vistos := map[string]string{}
	for _, r := range rutas {
		datos, err := os.ReadFile(r)
		if err != nil {
			errs = append(errs, ErrorCarga{Archivo: r, Mensaje: err.Error(), Grave: true})
			continue
		}
		cs, es := LeerConceptos(r, datos)
		errs = append(errs, es...)
		for _, c := range cs {
			if prev, ok := vistos[c.ID]; ok {
				errs = append(errs, ErrorCarga{Archivo: r, Linea: c.Linea, Grave: true,
					Mensaje: fmt.Sprintf("concepto %q repetido (ya cargado de %s): se salta", c.ID, prev)})
				continue
			}
			vistos[c.ID] = r
			out = append(out, c)
		}
	}
	return out, errs
}
