package conocimiento

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"rag-go/internal/v2/tipos"
)

// Procedimiento es un procedimiento cargado: la definición del contrato más su origen y la tabla con la
// que se resuelven sus citas.
type Procedimiento struct {
	tipos.ProcedimientoDef

	Archivo  string `json:"archivo"`
	Linea    int    `json:"linea"`
	Revisado bool   `json:"revisado_por_humano"`

	tabla    map[string]tipos.Fuente // id → (id, manual): la única forma de resolver una cita
	ambiguos map[string]bool         // ids que la tabla da con dos manuales: no se resuelven
	lineas   map[string]int          // id de paso → línea del YAML
}

// Par devuelve la entrada de la tabla «fuentes» (el par id, manual) para un id citado.
func (p *Procedimiento) Par(id string) (tipos.Fuente, bool) {
	if p.ambiguos[id] {
		return tipos.Fuente{}, false
	}
	f, ok := p.tabla[id]
	return f, ok
}

// PasoPorID busca un paso (o subpaso) por su id «<procedimiento>#<n>[.<m>]».
func (p *Procedimiento) PasoPorID(id string) (*tipos.Paso, bool) {
	var buscar func([]tipos.Paso) *tipos.Paso
	buscar = func(ps []tipos.Paso) *tipos.Paso {
		for i := range ps {
			if ps[i].ID == id {
				return &ps[i]
			}
			if s := buscar(ps[i].Sub); s != nil {
				return s
			}
		}
		return nil
	}
	s := buscar(p.Pasos)
	return s, s != nil
}

// EsDeConfiguracion: procedimiento cuyo fin es configurar (título, id, objetivo o aliases).
func (p *Procedimiento) EsDeConfiguracion() bool {
	t := plegar(p.ID + " " + p.Titulo + " " + p.Objetivo + " " + strings.Join(p.Aliases, " "))
	for _, r := range []string{"configur", "parametr", "personaliz"} {
		if strings.Contains(t, r) {
			return true
		}
	}
	return false
}

// --- lectura del YAML -------------------------------------------------------------------------------

type procYAML struct {
	ID                string                 `yaml:"id"`
	Version           int                    `yaml:"version"`
	Modulo            string                 `yaml:"modulo"`
	Titulo            string                 `yaml:"titulo"`
	Objetivo          string                 `yaml:"objetivo"`
	Nivel             string                 `yaml:"nivel"`
	Preguntas         []string               `yaml:"preguntas"`
	Aliases           []string               `yaml:"aliases"`
	Entidades         entidadesYAML          `yaml:"entidades"`
	Prerrequisitos    []tipos.ConFuente      `yaml:"prerrequisitos"`
	Pasos             []pasoYAML             `yaml:"pasos"`
	Verificacion      []tipos.ConFuente      `yaml:"verificacion"`
	ErroresFrecuentes []tipos.ErrorFrecuente `yaml:"errores_frecuentes"`
	Relacionados      []string               `yaml:"relacionados"`
	Fuentes           []fuenteYAML           `yaml:"fuentes"`
	Revision          map[string]any         `yaml:"revision"`
}

// entidadesYAML acepta el mapa {modulo, pantallas, objetos} del esquema o una lista plana.
type entidadesYAML struct {
	Modulo    string
	Pantallas []string
	Objetos   []string
	Otras     []string
}

func (e *entidadesYAML) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, v := n.Content[i].Value, n.Content[i+1]
			var lista []string
			switch v.Kind {
			case yaml.ScalarNode:
				if v.Tag != "!!null" && v.Value != "" {
					lista = []string{v.Value}
				}
			case yaml.SequenceNode:
				if err := v.Decode(&lista); err != nil {
					return err
				}
			default:
				return fmt.Errorf("línea %d: entidades.%s: se esperaba texto o lista", v.Line, k)
			}
			switch k {
			case "modulo":
				if len(lista) > 0 {
					e.Modulo = lista[0]
				}
			case "pantallas":
				e.Pantallas = lista
			case "objetos":
				e.Objetos = lista
			default:
				e.Otras = append(e.Otras, lista...)
			}
		}
	case yaml.SequenceNode:
		return n.Decode(&e.Otras)
	case yaml.ScalarNode:
		if n.Tag != "!!null" && n.Value != "" {
			e.Otras = []string{n.Value}
		}
	}
	return nil
}

// def: el contrato (tipos.EntidadesDef). Lo que venga fuera de {modulo, pantallas, objetos} cuenta como objeto.
func (e entidadesYAML) def() tipos.EntidadesDef {
	return tipos.EntidadesDef{Modulo: e.Modulo, Pantallas: unicos(e.Pantallas), Objetos: unicos(append(append([]string{}, e.Objetos...), e.Otras...))}
}

// entidadesPlanas: módulo + pantallas + objetos, para indexar y detectar.
func entidadesPlanas(e tipos.EntidadesDef) []string {
	var out []string
	if e.Modulo != "" {
		out = append(out, e.Modulo)
	}
	out = append(out, e.Pantallas...)
	out = append(out, e.Objetos...)
	return unicos(out)
}

type pasoYAML struct {
	linea     int
	ID        string     `yaml:"id"`
	N         int        `yaml:"n"`
	Accion    string     `yaml:"accion"`
	Donde     string     `yaml:"donde"`
	Resultado string     `yaml:"resultado"`
	Pagina    int        `yaml:"pagina"`
	Fuente    []string   `yaml:"fuente"`
	Captura   string     `yaml:"captura"`
	Fotos     []fotoYAML `yaml:"fotos"`
	Sub       []pasoYAML `yaml:"sub"`
}

func (p *pasoYAML) UnmarshalYAML(n *yaml.Node) error {
	type plano pasoYAML
	var x plano
	if err := n.Decode(&x); err != nil {
		return err
	}
	*p = pasoYAML(x)
	p.linea = n.Line
	return nil
}

type fotoYAML struct {
	linea int
	tipos.Foto
}

func (f *fotoYAML) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode { // forma corta: solo la ruta
		f.Ruta, f.linea = n.Value, n.Line
		return nil
	}
	if err := n.Decode(&f.Foto); err != nil {
		return err
	}
	f.linea = n.Line
	return nil
}

type fuenteYAML struct {
	linea int
	tipos.Fuente
}

func (f *fuenteYAML) UnmarshalYAML(n *yaml.Node) error {
	if err := n.Decode(&f.Fuente); err != nil {
		return err
	}
	f.linea = n.Line
	return nil
}

var lineaErrRe = regexp.MustCompile(`line (\d+)`)

// errorYAML convierte un error de yaml.v3 en ErrorCarga con su línea.
func errorYAML(ruta string, err error) ErrorCarga {
	msg := strings.TrimPrefix(err.Error(), "yaml: ")
	linea := 0
	if m := lineaErrRe.FindStringSubmatch(msg); m != nil {
		linea, _ = strconv.Atoi(m[1])
	}
	msg = strings.ReplaceAll(msg, "unmarshal errors:\n  ", "")
	msg = strings.ReplaceAll(msg, "\n  ", "; ")
	return ErrorCarga{Archivo: ruta, Linea: linea, Mensaje: "YAML inválido: " + msg, Grave: true}
}

// leerNodoYAML: parsea y exige un único documento cuyo nodo raíz sea un mapa (o una lista si se admite).
func leerNodoYAML(ruta string, datos []byte) (*yaml.Node, *ErrorCarga) {
	var doc yaml.Node
	if err := yaml.Unmarshal(datos, &doc); err != nil {
		e := errorYAML(ruta, err)
		return nil, &e
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, &ErrorCarga{Archivo: ruta, Linea: 1, Mensaje: "archivo vacío", Grave: true}
	}
	return doc.Content[0], nil
}

// LeerProcedimiento lee un procedimiento YAML. Si devuelve nil, el archivo se salta (errores graves con
// archivo y línea). Las citas se comprueban después, en Base, contra el índice de fragmentos.
func LeerProcedimiento(ruta string, datos []byte) (*Procedimiento, []ErrorCarga) {
	nodo, e := leerNodoYAML(ruta, datos)
	if e != nil {
		return nil, []ErrorCarga{*e}
	}
	if nodo.Kind != yaml.MappingNode {
		return nil, []ErrorCarga{{Archivo: ruta, Linea: nodo.Line, Mensaje: "se esperaba un mapa (id, titulo, pasos…)", Grave: true}}
	}
	var y procYAML
	if err := nodo.Decode(&y); err != nil {
		return nil, []ErrorCarga{errorYAML(ruta, err)}
	}
	grave := func(linea int, f string, a ...any) []ErrorCarga {
		return []ErrorCarga{{Archivo: ruta, Linea: linea, Mensaje: fmt.Sprintf(f, a...), Grave: true}}
	}
	if strings.TrimSpace(y.ID) == "" {
		return nil, grave(nodo.Line, "falta «id»")
	}
	if strings.TrimSpace(y.Titulo) == "" {
		return nil, grave(nodo.Line, "%s: falta «titulo»", y.ID)
	}
	if len(y.Pasos) == 0 {
		return nil, grave(lineaClave(nodo, "pasos"), "%s: sin «pasos»", y.ID)
	}

	p := &Procedimiento{
		Archivo:  ruta,
		Linea:    nodo.Line,
		tabla:    map[string]tipos.Fuente{},
		ambiguos: map[string]bool{},
		lineas:   map[string]int{},
	}
	if v, ok := y.Revision["revisado_por_humano"].(bool); ok {
		p.Revisado = v
	}
	var avisos []ErrorCarga
	aviso := func(linea int, f string, a ...any) {
		avisos = append(avisos, ErrorCarga{Archivo: ruta, Linea: linea, Mensaje: fmt.Sprintf(f, a...)})
	}

	fotosVistas := map[string]string{} // ruta → id del paso que ya la lleva
	var convertir func(ps []pasoYAML, prefijo string) ([]tipos.Paso, *ErrorCarga)
	convertir = func(ps []pasoYAML, prefijo string) ([]tipos.Paso, *ErrorCarga) {
		out := make([]tipos.Paso, 0, len(ps))
		for i, py := range ps {
			esperado := i + 1
			if py.N != esperado {
				return nil, &ErrorCarga{Archivo: ruta, Linea: py.linea, Grave: true,
					Mensaje: fmt.Sprintf("%s: paso con n=%d donde tocaba n=%d (los n deben ser 1, 2, 3…)", y.ID, py.N, esperado)}
			}
			idEsperado := prefijo + strconv.Itoa(py.N)
			if strings.TrimSpace(py.Accion) == "" {
				return nil, &ErrorCarga{Archivo: ruta, Linea: py.linea, Grave: true, Mensaje: fmt.Sprintf("%s: paso %s sin «accion»", y.ID, idEsperado)}
			}
			id := py.ID
			if id == "" {
				id = idEsperado
				aviso(py.linea, "paso sin «id»: se usa %s", id)
			} else if id != idEsperado {
				aviso(py.linea, "id de paso %q distinto del esperado %q", id, idEsperado)
			}
			if _, dup := p.lineas[id]; dup {
				return nil, &ErrorCarga{Archivo: ruta, Linea: py.linea, Grave: true, Mensaje: fmt.Sprintf("%s: id de paso repetido %q", y.ID, id)}
			}
			p.lineas[id] = py.linea
			paso := tipos.Paso{
				ID: id, N: py.N, Accion: strings.TrimSpace(py.Accion), Donde: py.Donde, Resultado: py.Resultado,
				Pagina: py.Pagina, Fuente: unicos(py.Fuente), Captura: py.Captura,
			}
			if len(paso.Fuente) == 0 {
				aviso(py.linea, "paso %s sin «fuente»", id)
			}
			// Fotos: forma nueva (fotos) y vieja (captura). Una ruta solo en un paso.
			fotos := py.Fotos
			if py.Captura != "" {
				fotos = append(fotos, fotoYAML{linea: py.linea, Foto: tipos.Foto{Ruta: py.Captura, Pagina: py.Pagina}})
			}
			for _, f := range fotos {
				foto := f.Foto
				foto.Ruta = strings.TrimSpace(foto.Ruta)
				if !rutaFotoValida(foto.Ruta) {
					aviso(f.linea, "foto con ruta no válida %q en el paso %s: se descarta", foto.Ruta, id)
					continue
				}
				calc := idFoto(foto.Ruta)
				if foto.ID == "" {
					foto.ID = calc
				} else if foto.ID != calc {
					aviso(f.linea, "foto %s del paso %s: id distinto de sha1(ruta)[:12]=%s", foto.ID, id, calc)
				}
				if otro, ok := fotosVistas[foto.Ruta]; ok {
					aviso(f.linea, "la foto %s ya está en el paso %s: no se repite en %s", foto.ID, otro, id)
					continue
				}
				fotosVistas[foto.Ruta] = id
				paso.Fotos = append(paso.Fotos, foto)
			}
			if len(py.Sub) > 0 {
				sub, e := convertir(py.Sub, id+".")
				if e != nil {
					return nil, e
				}
				// Los subpasos se numeran «#<n>.<m>»: el prefijo es el id del padre sin «#».
				paso.Sub = sub
			}
			out = append(out, paso)
		}
		return out, nil
	}
	pasos, eg := convertir(y.Pasos, y.ID+"#")
	if eg != nil {
		return nil, append(avisos, *eg)
	}

	// Tabla de fuentes: (id, manual). Un id con dos manuales queda sin resolver.
	for _, f := range y.Fuentes {
		if f.ID == "" || f.Manual == "" {
			aviso(f.linea, "entrada de «fuentes» sin id o sin manual: %q / %q", f.ID, f.Manual)
			continue
		}
		if prev, ok := p.tabla[f.ID]; ok {
			if prev.Manual != f.Manual {
				p.ambiguos[f.ID] = true
				aviso(f.linea, "el id %s figura con dos manuales («%s» y «%s»): sus citas no se resuelven", f.ID, prev.Manual, f.Manual)
			}
			continue
		}
		p.tabla[f.ID] = f.Fuente
	}
	if y.Modulo != "" {
		if dir := filepath.Base(filepath.Dir(ruta)); dir != y.Modulo && dir != "procedimientos" {
			aviso(nodo.Line, "modulo %q no coincide con la carpeta %q", y.Modulo, dir)
		}
	}
	if base := strings.TrimSuffix(filepath.Base(ruta), filepath.Ext(ruta)); base != y.ID {
		aviso(nodo.Line, "el id %q no coincide con el nombre del archivo %q", y.ID, base)
	}

	p.ProcedimientoDef = tipos.ProcedimientoDef{
		ID: y.ID, Version: y.Version, Modulo: y.Modulo, Titulo: strings.TrimSpace(y.Titulo),
		Objetivo: y.Objetivo, Nivel: y.Nivel, Preguntas: y.Preguntas, Aliases: y.Aliases,
		Entidades: y.Entidades.def(), Prerrequisitos: y.Prerrequisitos, Pasos: pasos,
		Verificacion: y.Verificacion, ErroresFrecuentes: y.ErroresFrecuentes, Relacionados: y.Relacionados,
	}
	for _, f := range y.Fuentes {
		if f.ID != "" && f.Manual != "" {
			p.Fuentes = append(p.Fuentes, f.Fuente)
		}
	}
	return p, avisos
}

// lineaClave: línea de la clave k de un mapa (o la del mapa si no está).
func lineaClave(m *yaml.Node, k string) int {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == k {
			return m.Content[i].Line
		}
	}
	return m.Line
}

// citasDe recorre todas las citas de un procedimiento: (dónde, id, línea).
func (p *Procedimiento) citas(fn func(donde, id string, linea int)) {
	for i, x := range p.Prerrequisitos {
		for _, id := range x.Fuente {
			fn(fmt.Sprintf("prerrequisito %d", i+1), id, p.Linea)
		}
	}
	var pasos func([]tipos.Paso)
	pasos = func(ps []tipos.Paso) {
		for _, s := range ps {
			for _, id := range s.Fuente {
				fn("paso "+s.ID, id, p.lineas[s.ID])
			}
			pasos(s.Sub)
		}
	}
	pasos(p.Pasos)
	for i, x := range p.Verificacion {
		for _, id := range x.Fuente {
			fn(fmt.Sprintf("verificación %d", i+1), id, p.Linea)
		}
	}
	for i, x := range p.ErroresFrecuentes {
		for _, id := range x.Fuente {
			fn(fmt.Sprintf("error frecuente %d", i+1), id, p.Linea)
		}
	}
}

// CargarProcedimientos lee kb/procedimientos/**/*.yml (también .yaml), ordenados por ruta. Las carpetas
// que empiezan por «_» (p. ej. _reserva/) no se cargan: son borradores fuera de servicio.
func CargarProcedimientos(dir string) ([]*Procedimiento, []ErrorCarga) {
	return CargarProcedimientosCon(dir, false)
}

// CargarProcedimientosCon permite incluir las carpetas «_…» (reserva).
func CargarProcedimientosCon(dir string, incluirReserva bool) ([]*Procedimiento, []ErrorCarga) {
	var rutas []string
	_ = filepath.WalkDir(dir, func(r string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && r != dir && strings.HasPrefix(d.Name(), "_") && !incluirReserva {
			return filepath.SkipDir
		}
		if !d.IsDir() && (strings.HasSuffix(r, ".yml") || strings.HasSuffix(r, ".yaml")) {
			rutas = append(rutas, r)
		}
		return nil
	})
	sort.Strings(rutas)
	var out []*Procedimiento
	var errs []ErrorCarga
	vistos := map[string]string{}
	for _, r := range rutas {
		datos, err := os.ReadFile(r)
		if err != nil {
			errs = append(errs, ErrorCarga{Archivo: r, Mensaje: err.Error(), Grave: true})
			continue
		}
		p, es := LeerProcedimiento(r, datos)
		errs = append(errs, es...)
		if p == nil {
			continue
		}
		if prev, ok := vistos[p.ID]; ok {
			errs = append(errs, ErrorCarga{Archivo: r, Linea: p.Linea, Grave: true,
				Mensaje: fmt.Sprintf("id %q repetido (ya cargado de %s): se salta", p.ID, prev)})
			continue
		}
		vistos[p.ID] = r
		out = append(out, p)
	}
	return out, errs
}
