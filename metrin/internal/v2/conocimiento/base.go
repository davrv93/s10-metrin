package conocimiento

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"rag-go/internal/v2/tipos"
)

// Opciones de carga. Lo vacío se busca solo (ver RutaKB, RutaPlantillas, RutaDatos).
type Opciones struct {
	DirKB              string   // kb/ (procedimientos/, conceptos/, fragmentos*.jsonl)
	ArchivoPlantillas  string   // metrin/plantillas/respuestas.yml
	DirDatos           string   // data/ (las fotos «imagenes/…» se sirven desde aquí); opcional
	ArchivosFragmentos []string // si se dan, sustituyen a <DirKB>/fragmentos*.jsonl
	SinFragmentos      bool     // no cargar el índice de fragmentos (las citas no se comprueban)
	IncluirReserva     bool     // cargar también kb/procedimientos/_reserva/ (y toda carpeta «_…»)
}

// Base es el conocimiento cargado. Es de solo lectura después de Cargar: se comparte entre peticiones.
type Base struct {
	Procedimientos []*Procedimiento
	Conceptos      []*Concepto
	Plantillas     *Plantillas
	Fragmentos     *Fragmentos // nil con SinFragmentos
	Errores        []ErrorCarga

	DirKB, DirDatos, ArchivoPlantillas string

	porProc     map[string]*Procedimiento
	porConcepto map[string]*Concepto
	alias       []Alias
}

// RutaKB: $RAG_KB, /kb, ../kb, kb, ../../kb… (el primero que tenga procedimientos/ o fragmentos.jsonl).
func RutaKB() string {
	if r := os.Getenv("RAG_KB"); r != "" {
		return r
	}
	for _, c := range []string{"/kb", "../kb", "kb", "../../kb", "../../../kb", "../../../../kb"} {
		if existe(filepath.Join(c, "procedimientos")) || existe(filepath.Join(c, "fragmentos.jsonl")) {
			return c
		}
	}
	return "/kb"
}

// RutaPlantillas: $RAG_PLANTILLAS o el primero que exista de las rutas habituales («» si ninguno).
func RutaPlantillas() string {
	if r := os.Getenv("RAG_PLANTILLAS"); r != "" {
		return r
	}
	for _, c := range []string{"/srv/plantillas/respuestas.yml", "/plantillas/respuestas.yml", "plantillas/respuestas.yml",
		"metrin/plantillas/respuestas.yml", "../plantillas/respuestas.yml", "../../../plantillas/respuestas.yml",
		"../../../../metrin/plantillas/respuestas.yml"} {
		if existe(c) {
			return c
		}
	}
	return ""
}

// RutaDatos: $RAG_MEDIA_DATOS o data/ (la misma búsqueda que servidor.rutaDatos).
func RutaDatos() string {
	if r := os.Getenv("RAG_MEDIA_DATOS"); r != "" {
		return r
	}
	for _, c := range []string{"/datos", "../data", "data", "../../../../data"} {
		if existe(filepath.Join(c, "imagenes")) {
			return c
		}
	}
	return ""
}

func existe(r string) bool { _, err := os.Stat(r); return err == nil }

// Cargar lee todo. Nunca falla: lo que no se pudo leer queda en Errores (graves: se saltó).
func Cargar(o Opciones) *Base {
	if o.DirKB == "" {
		o.DirKB = RutaKB()
	}
	if o.ArchivoPlantillas == "" {
		o.ArchivoPlantillas = RutaPlantillas()
	}
	b := &Base{DirKB: o.DirKB, DirDatos: o.DirDatos, ArchivoPlantillas: o.ArchivoPlantillas}
	if !o.SinFragmentos {
		arch := o.ArchivosFragmentos
		if len(arch) == 0 {
			arch = archivosFragmentos(o.DirKB)
		}
		if len(arch) > 0 {
			fs, errs := CargarFragmentos(arch...)
			b.Fragmentos = fs
			b.Errores = append(b.Errores, errs...)
		} else {
			b.Errores = append(b.Errores, ErrorCarga{Archivo: filepath.Join(o.DirKB, "fragmentos*.jsonl"),
				Mensaje: "no hay índice de fragmentos: las citas no se comprueban"})
		}
	}
	ps, errs := CargarProcedimientosCon(filepath.Join(o.DirKB, "procedimientos"), o.IncluirReserva)
	b.Errores = append(b.Errores, errs...)
	cs, errs := CargarConceptos(filepath.Join(o.DirKB, "conceptos"))
	b.Errores = append(b.Errores, errs...)
	pl, errs := CargarPlantillas(o.ArchivoPlantillas)
	b.Errores = append(b.Errores, errs...)
	b.Plantillas = pl
	b.armar(ps, cs)
	return b
}

// NuevaBase arma una Base con datos ya leídos (para pruebas y para quien cargue por su cuenta).
func NuevaBase(ps []*Procedimiento, cs []*Concepto, pl *Plantillas, fs *Fragmentos) *Base {
	if pl == nil {
		pl = PlantillasEmbebidas()
	}
	b := &Base{Plantillas: pl, Fragmentos: fs}
	b.armar(ps, cs)
	return b
}

func (b *Base) armar(ps []*Procedimiento, cs []*Concepto) {
	b.porProc = map[string]*Procedimiento{}
	b.porConcepto = map[string]*Concepto{}
	for _, p := range ps {
		if _, ok := b.porProc[p.ID]; ok {
			continue
		}
		b.porProc[p.ID] = p
		b.Procedimientos = append(b.Procedimientos, p)
	}
	for _, c := range cs {
		if _, ok := b.porConcepto[c.ID]; ok {
			continue
		}
		b.porConcepto[c.ID] = c
		b.Conceptos = append(b.Conceptos, c)
	}
	b.validarProcedimientos()
	b.resolverConceptos()
	b.armarAliases()
}

// Proc devuelve el procedimiento cargado (o el implícito «implicito:<id>@<manual>», armado desde los
// pasos de esa sección del manual).
func (b *Base) Proc(id string) (*Procedimiento, bool) {
	if p, ok := b.porProc[id]; ok {
		return p, true
	}
	if strings.HasPrefix(id, PrefijoImplicito) {
		return b.procedimientoImplicito(id)
	}
	return nil, false
}

// Conc devuelve el concepto cargado.
func (b *Base) Conc(id string) (*Concepto, bool) { c, ok := b.porConcepto[id]; return c, ok }

// Procedimiento cumple v2.Catalogo: la definición por id (también la de un procedimiento implícito).
func (b *Base) Procedimiento(id string) (tipos.ProcedimientoDef, bool) {
	p, ok := b.Proc(id)
	if !ok {
		return tipos.ProcedimientoDef{}, false
	}
	return p.ProcedimientoDef, true
}

// Concepto cumple v2.Catalogo.
func (b *Base) Concepto(id string) (tipos.ConceptoDef, bool) {
	c, ok := b.Conc(id)
	if !ok {
		return tipos.ConceptoDef{}, false
	}
	return c.ConceptoDef, true
}

// Analizar cumple v2.Aliaser: términos del ERP nombrados en el texto (como se escriben en S10), aliases
// para expandir la búsqueda y el módulo, si todo lo detectado es de uno solo. No cambia el texto.
func (b *Base) Analizar(texto string) (entidades, aliases []string, modulo string) {
	d := b.Detectar(texto)
	return d.Entidades, d.Aliases, d.Modulo
}

// Graves: los errores que hicieron saltar algo.
func (b *Base) Graves() []ErrorCarga {
	var out []ErrorCarga
	for _, e := range b.Errores {
		if e.Grave {
			out = append(out, e)
		}
	}
	return out
}

// Resumen para el registro de arranque.
func (b *Base) Resumen() string {
	graves := len(b.Graves())
	pl := "embebidas"
	if b.Plantillas != nil && b.Plantillas.Archivo != "" {
		pl = b.Plantillas.Archivo
	}
	return fmt.Sprintf("conocimiento V2: %d procedimientos, %d conceptos, %d fragmentos, plantillas %s; %d errores, %d avisos",
		len(b.Procedimientos), len(b.Conceptos), b.Fragmentos.Total(), pl, graves, len(b.Errores)-graves)
}

// --- citas -------------------------------------------------------------------------------------------

// Cita resuelta: el par (id, manual) y, si hay índice, el fragmento.
type Cita struct {
	ID        string     `json:"id"`
	Manual    string     `json:"manual"`
	Seccion   string     `json:"seccion,omitempty"`
	Pagina    int        `json:"pagina,omitempty"`
	Fragmento *Fragmento `json:"-"`
}

// hayIndice: con índice de fragmentos, una cita solo resuelve si el par existe.
func (b *Base) hayIndice() bool { return b.Fragmentos != nil && b.Fragmentos.Total() > 0 }

func (b *Base) citaDePar(f tipos.Fuente) (Cita, error) {
	c := Cita{ID: f.ID, Manual: f.Manual, Seccion: f.Seccion, Pagina: f.Pagina}
	if !b.hayIndice() {
		return c, nil
	}
	fr, ok := b.Fragmentos.Buscar(f.ID, f.Manual)
	if !ok {
		if n := len(b.Fragmentos.PorID(f.ID)); n > 0 {
			return c, fmt.Errorf("(%s, %s) no existe en kb/fragmentos*.jsonl (el id existe en %d manual(es) distinto(s))", f.ID, f.Manual, n)
		}
		return c, fmt.Errorf("(%s, %s) no existe en kb/fragmentos*.jsonl", f.ID, f.Manual)
	}
	c.Fragmento = fr
	if c.Seccion == "" {
		c.Seccion = fr.Seccion
	}
	if c.Pagina == 0 {
		c.Pagina = fr.Pagina
	}
	return c, nil
}

// CitaProcedimiento resuelve una cita de un procedimiento con SU tabla «fuentes» (nunca solo por id).
func (b *Base) CitaProcedimiento(p *Procedimiento, id string) (Cita, error) {
	if p.ambiguos[id] {
		return Cita{ID: id}, fmt.Errorf("el id %s figura con dos manuales en la tabla de %s", id, p.ID)
	}
	f, ok := p.tabla[id]
	if !ok {
		return Cita{ID: id}, fmt.Errorf("el id %s no está en la tabla «fuentes» de %s", id, p.ID)
	}
	return b.citaDePar(f)
}

// CitaConcepto resuelve una cita de un concepto (tabla, {id, manual}, comentario o id de un solo manual).
func (b *Base) CitaConcepto(c *Concepto, id string) (Cita, error) {
	if c.ambiguos[id] {
		return Cita{ID: id}, fmt.Errorf("el id %s es ambiguo (varios manuales) y el concepto %s no dice cuál", id, c.ID)
	}
	f, ok := c.tabla[id]
	if !ok {
		return Cita{ID: id}, fmt.Errorf("el concepto %s no dice de qué manual es %s", c.ID, id)
	}
	return b.citaDePar(f)
}

// TerminoRespaldado: el término aparece literal (sin tildes ni mayúsculas) en algún fragmento que el
// concepto cita. Es la condición para escribirlo en negrita.
func (b *Base) TerminoRespaldado(conceptoID, termino string) bool {
	c, ok := b.Conc(conceptoID)
	t := plegar(termino)
	if !ok || t == "" {
		return false
	}
	for _, id := range c.Fuente {
		if ci, err := b.CitaConcepto(c, id); err == nil && ci.Fragmento != nil && strings.Contains(ci.Fragmento.TextoPlegado(), t) {
			return true
		}
	}
	return false
}

// CitaPar resuelve un par (id, manual) directamente (afirmaciones, procedimientos implícitos).
func (b *Base) CitaPar(id, manual string) (Cita, error) {
	return b.citaDePar(tipos.Fuente{ID: id, Manual: manual})
}

// validarProcedimientos: avisos (no graves) para citas que no resuelven, tabla sin citar, relacionados
// inexistentes y fotos que la fuente no asocia a su paso.
func (b *Base) validarProcedimientos() {
	for _, p := range b.Procedimientos {
		citados := map[string]bool{}
		p.citas(func(donde, id string, linea int) {
			citados[id] = true
			if _, err := b.CitaProcedimiento(p, id); err != nil {
				b.Errores = append(b.Errores, ErrorCarga{Archivo: p.Archivo, Linea: linea, Mensaje: donde + ": cita sin resolver: " + err.Error()})
			}
		})
		for _, f := range p.Fuentes {
			if !citados[f.ID] {
				b.Errores = append(b.Errores, ErrorCarga{Archivo: p.Archivo, Linea: p.Linea, Mensaje: "la fuente " + f.ID + " de la tabla no la cita nadie"})
			}
			if EsMarketing(f.Manual, "") {
				b.Errores = append(b.Errores, ErrorCarga{Archivo: p.Archivo, Linea: p.Linea, Mensaje: "fuente de marketing en la tabla: " + f.Manual})
			}
		}
		for _, r := range p.Relacionados {
			if _, ok := b.porProc[r]; !ok {
				b.Errores = append(b.Errores, ErrorCarga{Archivo: p.Archivo, Linea: p.Linea, Mensaje: "relacionado inexistente: " + r})
			}
		}
		if b.hayIndice() {
			var fotos func([]tipos.Paso)
			fotos = func(ps []tipos.Paso) {
				for _, s := range ps {
					for _, f := range s.Fotos {
						if !b.fuenteAsociaFoto(p, &s, f.Ruta) {
							b.Errores = append(b.Errores, ErrorCarga{Archivo: p.Archivo, Linea: p.lineas[s.ID],
								Mensaje: fmt.Sprintf("foto %s del paso %s: ninguna fuente del paso la asocia", f.ID, s.ID)})
						}
					}
					fotos(s.Sub)
				}
			}
			fotos(p.Pasos)
		}
		if b.DirDatos != "" {
			p.recorrerFotos(func(paso *tipos.Paso, f tipos.Foto) {
				if !strings.HasPrefix(f.Ruta, "http") && !existe(filepath.Join(b.DirDatos, f.Ruta)) {
					b.Errores = append(b.Errores, ErrorCarga{Archivo: p.Archivo, Linea: p.lineas[paso.ID],
						Mensaje: fmt.Sprintf("foto %s del paso %s: no existe %s", f.ID, paso.ID, filepath.Join(b.DirDatos, f.Ruta))})
				}
			})
		}
	}
}

func (p *Procedimiento) recorrerFotos(fn func(*tipos.Paso, tipos.Foto)) {
	var rec func([]tipos.Paso)
	rec = func(ps []tipos.Paso) {
		for i := range ps {
			for _, f := range ps[i].Fotos {
				fn(&ps[i], f)
			}
			rec(ps[i].Sub)
		}
	}
	rec(p.Pasos)
}

// fuenteAsociaFoto: la foto está en pasos[k].fotos de una sección que el paso cita, o es la imagen de un
// fragmento imagen que el paso cita (ESQUEMA.md, «Regla de asociación»).
func (b *Base) fuenteAsociaFoto(p *Procedimiento, s *tipos.Paso, ruta string) bool {
	for _, id := range s.Fuente {
		c, err := b.CitaProcedimiento(p, id)
		if err != nil || c.Fragmento == nil {
			continue
		}
		if c.Fragmento.Imagen == ruta {
			return true
		}
		for _, pf := range c.Fragmento.Pasos {
			for _, f := range pf.Fotos {
				if f == ruta {
					return true
				}
			}
		}
	}
	return false
}

// resolverConceptos arma la tabla (id → manual) de cada concepto: tabla/{id, manual} del YAML, el
// comentario «# Manual › sección», o el id si existe en un único manual. Si no, la cita no resuelve.
func (b *Base) resolverConceptos() {
	for _, c := range b.Conceptos {
		c.tabla = map[string]tipos.Fuente{}
		c.ambiguos = map[string]bool{}
		for _, cy := range c.citas {
			f := tipos.Fuente{ID: cy.ID, Manual: cy.Manual, Seccion: cy.Seccion}
			if f.Manual == "" && b.hayIndice() {
				if frs := b.Fragmentos.PorID(cy.ID); len(frs) == 1 {
					f.Manual, f.Seccion = frs[0].Manual, frs[0].Seccion
				}
			}
			if f.Manual == "" {
				c.ambiguos[cy.ID] = true
				n := 0
				if b.Fragmentos != nil {
					n = len(b.Fragmentos.PorID(cy.ID))
				}
				b.Errores = append(b.Errores, ErrorCarga{Archivo: c.Archivo, Linea: cy.Linea,
					Mensaje: fmt.Sprintf("concepto %s: la cita %s no dice el manual (el id está en %d manuales): no se resuelve", c.ID, cy.ID, n)})
				continue
			}
			if prev, ok := c.tabla[cy.ID]; ok && prev.Manual != f.Manual {
				c.ambiguos[cy.ID] = true
				b.Errores = append(b.Errores, ErrorCarga{Archivo: c.Archivo, Linea: cy.Linea,
					Mensaje: fmt.Sprintf("concepto %s: la cita %s aparece con dos manuales", c.ID, cy.ID)})
				continue
			}
			c.tabla[cy.ID] = f
			ci, err := b.CitaConcepto(c, cy.ID)
			if err != nil {
				b.Errores = append(b.Errores, ErrorCarga{Archivo: c.Archivo, Linea: cy.Linea,
					Mensaje: "concepto " + c.ID + ": cita sin resolver: " + err.Error()})
			} else if ci.Fragmento != nil && (EsMarketing(ci.Manual, ci.Fragmento.Confianza) || EsTangencial(ci.Manual, ci.Fragmento.Confianza)) {
				b.Errores = append(b.Errores, ErrorCarga{Archivo: c.Archivo, Linea: cy.Linea,
					Mensaje: "concepto " + c.ID + ": la cita " + cy.ID + " es de marketing o tangencial (" + ci.Manual + "): el gate rechazará la respuesta"})
			}
		}
		for _, r := range c.ProcedimientosRelacionados {
			if _, ok := b.porProc[r]; !ok {
				b.Errores = append(b.Errores, ErrorCarga{Archivo: c.Archivo, Linea: c.Linea,
					Mensaje: "concepto " + c.ID + ": procedimiento relacionado inexistente: " + r})
			}
		}
	}
}

// --- fuentes del plan --------------------------------------------------------------------------------

// marketingRe: páginas que no se citan (ESQUEMA.md: marketing, menús web, «Mes: …», portadas).
var marketingRe = regexp.MustCompile(`^(Optimiza 360|s10peru\.com \(público\)|Mes: |S10 ERP, El Software|CURSOS S10|` +
	`Soporte Técnico \||Servicio de Implantación \||Sidebar|Home |panel \d|Portada |CONTACTO|Join Us|unete|topbar|` +
	`varios widgets|Página de ejemplo|Acceso de miembro|Perfil$|Registro$|Formulario Feedback|Temporal$|Sin categoría$)`)

// EsMarketing: fuente de marketing o de navegación web, que no respalda un paso.
func EsMarketing(manual, confianza string) bool {
	return confianza == "propio" || marketingRe.MatchString(strings.TrimSpace(manual))
}

// EsTangencial: fuente que no es documentación de uso de S10 (tesis, transparencia, 2013).
func EsTangencial(manual, confianza string) bool {
	switch confianza {
	case "academico", "estado-2013":
		return true
	}
	return strings.HasPrefix(manual, "Tesis ") || strings.HasPrefix(manual, "CONCYTEC")
}

// NombreManual: el nombre que se muestra. Los PDF importados a mano guardan el título en la sección.
func NombreManual(c Cita) string {
	if c.Manual == "Importado a mano" {
		if c.Fragmento != nil && c.Fragmento.Titulo != "" {
			return c.Fragmento.Titulo
		}
		if c.Seccion != "" {
			return c.Seccion
		}
	}
	return c.Manual
}

var seccionImagenRe = regexp.MustCompile(`(?i)\.(png|jpe?g|gif|webp)$`)

// ResumirFuentes agrupa citas por manual (en orden de aparición), con sus páginas y secciones.
//
// Una entrada por SECCIÓN («Manual de Presupuestos › 3.3 Registro del presupuesto (1)», que es como V1 cita
// ese fragmento); los documentos paginados (PDF importados) van en una entrada por documento con sus
// páginas. Las capturas (sección «x.png») no abren entrada propia: cuentan en la de su manual si este no
// tiene otra. Cada entrada lleva en Citas los pares (id, manual) que resume.
func ResumirFuentes(citas []Cita) []tipos.FuentePlan {
	type acum struct {
		fp      tipos.FuentePlan
		paginas map[int]bool
	}
	var orden []string
	m := map[string]*acum{}
	conSeccion := map[string]bool{}
	for _, c := range citas {
		nombre := NombreManual(c)
		if nombre == "" {
			continue
		}
		sec := c.Seccion
		if sec == nombre || seccionImagenRe.MatchString(sec) {
			sec = ""
		}
		if sec != "" {
			conSeccion[nombre] = true
		}
		k := nombre + "\x00" + sec
		a, ok := m[k]
		if !ok {
			a = &acum{fp: tipos.FuentePlan{Manual: nombre, Seccion: sec}, paginas: map[int]bool{}}
			m[k] = a
			orden = append(orden, k)
		}
		if c.Pagina > 0 {
			a.paginas[c.Pagina] = true
		}
		a.fp.Citas = append(a.fp.Citas, tipos.Fuente{ID: c.ID, Manual: c.Manual, Seccion: c.Seccion, Pagina: c.Pagina})
	}
	out := []tipos.FuentePlan{}
	var sueltas []tipos.Fuente // capturas de un manual que ya tiene secciones
	for _, k := range orden {
		a := m[k]
		if a.fp.Seccion == "" && len(a.paginas) == 0 && conSeccion[a.fp.Manual] {
			sueltas = append(sueltas, a.fp.Citas...)
			continue
		}
		for p := range a.paginas {
			a.fp.Paginas = append(a.fp.Paginas, p)
		}
		sort.Ints(a.fp.Paginas)
		out = append(out, a.fp)
	}
	for _, c := range sueltas { // a la primera entrada de su manual
		for i := range out {
			if out[i].Manual == NombreManual(Cita{ID: c.ID, Manual: c.Manual, Seccion: c.Seccion}) || out[i].Manual == c.Manual {
				out[i].Citas = append(out[i].Citas, c)
				break
			}
		}
	}
	return out
}
