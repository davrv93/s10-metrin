package conocimiento

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Plantilla instruccional de metrin/plantillas/respuestas.yml. Dos formas:
//
//   - compuesta (la del archivo): «forma» con {parte}; cada parte tiene variantes con huecos {{SLOT}}; una
//     variante solo es elegible si todos sus slots tienen dato; si ninguna lo es, se usa «respaldo» (que
//     puede ser ""); sin respaldo, la parte es obligatoria y la plantilla no se puede rellenar.
//   - simple (las embebidas): «texto» con huecos {{SLOT}}; una línea que empieza por «?» es opcional (se
//     quita si le falta un dato); si le falta un dato a una línea obligatoria, no se puede rellenar.
//
// La variante se elige de forma determinista: la de «fijas» o la primera elegible (con Semilla, por
// hash). «por_accion» no se aplica aquí: lo decide quien llama, si lo necesita.
type Plantilla struct {
	Nombre     string              `json:"nombre"`
	Tipo       string              `json:"tipo,omitempty"`
	Texto      string              `json:"texto,omitempty"`
	Forma      string              `json:"forma,omitempty"`
	Partes     map[string][]string `json:"partes,omitempty"`
	Respaldo   map[string]string   `json:"respaldo,omitempty"`
	Fijas      map[string]int      `json:"fijas,omitempty"`
	SinModelo  bool                `json:"sin_modelo,omitempty"`
	MaxFrases  int                 `json:"max_frases,omitempty"`
	Protegidos []string            `json:"protegidos,omitempty"`
	Origen     string              `json:"origen"` // archivo | embebida
	Linea      int                 `json:"linea,omitempty"`
}

// Plantillas: las del archivo y, debajo, las embebidas mínimas para lo que el archivo no traiga o no se
// pueda rellenar con los datos del plan.
type Plantillas struct {
	Archivo        string
	MaxPasosBloque int // config.max_pasos_por_bloque del archivo (0 = no lo dice)
	m              map[string]Plantilla
	embebidas      map[string]Plantilla
}

// Plantillas mínimas embebidas, con los mismos slots que el archivo. Trato de usted. Sin negritas
// propias: el quality gate exige que todo ** ** venga de la fuente (o del módulo/término del plan).
var plantillasEmbebidas = map[string]string{
	"PROCEDIMIENTO_INTRO":     "Para {{TAREA}}, siga estos pasos.",
	"PROCEDIMIENTO_IMPLICITO": "No tengo una guía revisada para esto; la sección «{{SECCION}}» del {{MANUAL}} lo describe así:",
	"PRERREQUISITOS":          "Antes de empezar:\n{{PRERREQUISITOS}}",
	"PASOS_BLOQUE":            "Estos son los pasos {{N_DESDE}} a {{N_HASTA}} de {{TOTAL_PASOS}}:\n\n{{PASOS}}\n\n?_Fuente: {{CITA}}_",
	"PASO":                    "{{N}}. {{ACCION}}",
	"PREGUNTA_AVANCE":         "Cuando lo tenga, dígame y seguimos con el paso {{N_SIGUIENTE}}.",
	"COMPROBACION":            "Para comprobar que quedó bien:\n{{VERIFICACION_FINAL}}",
	"VERIFICACION":            "Para comprobar el paso {{N}}: {{VERIFICACION}}",
	"PROCEDIMIENTO_FIN":       "Esos eran los {{TOTAL_PASOS}} pasos para {{TAREA}}.",
	"ERROR_FRECUENTE":         "El manual describe ese caso: {{SINTOMA}}\nLo que indica: {{SOLUCION}}\n?_({{CITA}})_",
	"CONCEPTO":                "{{TERMINO}}: {{DEFINICION}}\n?En S10: {{EN_S10}}\n?_({{CITA}})_",
	"OFRECER_PROCEDIMIENTO":   "¿Quiere que le enseñe a {{TAREA_SUGERIDA}}?",
	"COMPARACION":             "Así los define la documentación:\n- {{TERMINO_A}}: {{DEFINICION_A}}\n- {{TERMINO_B}}: {{DEFINICION_B}}\n?_Fuentes: {{CITA}}_",
	"NAVEGACION":              "Lo encuentra en {{RUTA_MENU}}.\n?_({{CITA}})_",
	"ACLARAR_TAREA":           "Para no darle pasos equivocados, necesito un dato más. ¿Qué quiere hacer: {{OPCIONES_TAREA}}?",
	"SIN_EVIDENCIA":           "No encontré eso en los manuales que tengo. ¿Prefiere reformular la pregunta, o que la deje registrada para que la revise el equipo?",
	"FUENTES":                 "_Fuente: {{CITA}}_",
}

func nuevasPlantillas() *Plantillas {
	ps := &Plantillas{m: map[string]Plantilla{}, embebidas: map[string]Plantilla{}}
	for k, v := range plantillasEmbebidas {
		ps.embebidas[k] = Plantilla{Nombre: k, Texto: v, Origen: "embebida"}
	}
	return ps
}

// PlantillasEmbebidas devuelve solo las mínimas embebidas (sin archivo).
func PlantillasEmbebidas() *Plantillas { return nuevasPlantillas() }

// Nombres de las plantillas disponibles (archivo + embebidas), ordenados.
func (ps *Plantillas) Nombres() []string {
	vis := map[string]bool{}
	for k := range ps.m {
		vis[k] = true
	}
	for k := range ps.embebidas {
		vis[k] = true
	}
	var out []string
	for k := range vis {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Get devuelve la plantilla (la del archivo si existe).
func (ps *Plantillas) Get(nombre string) (Plantilla, bool) {
	if p, ok := ps.m[nombre]; ok {
		return p, true
	}
	p, ok := ps.embebidas[nombre]
	return p, ok
}

// OpcRelleno: Omitir quita partes de la forma (p. ej. «arranque» de PROCEDIMIENTO_INTRO cuando los pasos
// van en el mismo mensaje); Semilla elige entre variantes elegibles (0 = la primera, determinista).
//
// Negrita decide si un slot que la plantilla envuelve en negrita («**{{MODULO}}**», «**{{TERMINO}}**»)
// la conserva. Regla: un término va entre ** ** SOLO si aparece literal (normalizado) en un fragmento
// citado por el elemento del que sale; si no, va sin negrita. Sin Negrita (nil), ninguno la conserva. Las
// negritas que ya traen los datos (la acción de un paso) no se tocan: las valida el gate contra su fuente.
type OpcRelleno struct {
	Omitir  []string
	Semilla uint64
	Negrita func(slot, valor string) bool
}

// Rellenar pone los slots en la plantilla. Si la del archivo no se puede rellenar (falta un dato
// obligatorio), usa la embebida; si tampoco, devuelve false.
func (ps *Plantillas) Rellenar(nombre string, slots map[string]string, o ...OpcRelleno) (string, bool) {
	var op OpcRelleno
	if len(o) > 0 {
		op = o[0]
	}
	if p, ok := ps.m[nombre]; ok {
		if s, ok := p.rellenar(slots, op); ok {
			return s, true
		}
	}
	if p, ok := ps.embebidas[nombre]; ok {
		return p.rellenar(slots, op)
	}
	return "", false
}

// R es Rellenar sin el bool (cadena vacía si no se puede).
func (ps *Plantillas) R(nombre string, slots map[string]string, o ...OpcRelleno) string {
	s, _ := ps.Rellenar(nombre, slots, o...)
	return s
}

var (
	slotRe   = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)
	parteRe  = regexp.MustCompile(`\{([a-z_][a-z0-9_]*)\}`)
	blancos  = regexp.MustCompile(`\n{3,}`)
	espacios = regexp.MustCompile(`[ \t]{2,}`)
)

// elegible: todos los slots de la variante tienen dato no vacío.
func elegible(v string, slots map[string]string) bool {
	for _, m := range slotRe.FindAllStringSubmatch(v, -1) {
		if strings.TrimSpace(slots[m[1]]) == "" {
			return false
		}
	}
	return true
}

var slotNegritaRe = regexp.MustCompile(`\*\*\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}\*\*`)

// desnegritar quita las ** ** que la plantilla pone alrededor de un slot sin respaldo.
func desnegritar(v string, slots map[string]string, op OpcRelleno) string {
	return slotNegritaRe.ReplaceAllStringFunc(v, func(h string) string {
		slot := slotNegritaRe.FindStringSubmatch(h)[1]
		if op.Negrita != nil && op.Negrita(slot, slots[slot]) {
			return h
		}
		return "{{" + slot + "}}"
	})
}

func ponerSlots(v string, slots map[string]string) string {
	return slotRe.ReplaceAllStringFunc(v, func(h string) string {
		return slots[slotRe.FindStringSubmatch(h)[1]]
	})
}

func (p Plantilla) rellenar(slots map[string]string, op OpcRelleno) (string, bool) {
	if p.Forma == "" {
		var ls []string
		for _, l := range strings.Split(p.Texto, "\n") {
			opcional := strings.HasPrefix(l, "?")
			l = strings.TrimPrefix(l, "?")
			if !elegible(l, slots) {
				if opcional {
					continue
				}
				return "", false
			}
			ls = append(ls, ponerSlots(desnegritar(l, slots, op), slots))
		}
		return limpiar(strings.Join(ls, "\n")), true
	}
	omitir := map[string]bool{}
	for _, x := range op.Omitir {
		omitir[x] = true
	}
	ok := true
	s := parteRe.ReplaceAllStringFunc(p.Forma, func(h string) string {
		parte := parteRe.FindStringSubmatch(h)[1]
		if omitir[parte] {
			return ""
		}
		vars := p.Partes[parte]
		var eleg []int
		for i, v := range vars {
			if elegible(v, slots) {
				eleg = append(eleg, i)
			}
		}
		if f, hay := p.Fijas[parte]; hay && f < len(vars) && elegible(vars[f], slots) {
			return ponerSlots(desnegritar(vars[f], slots, op), slots)
		}
		if len(eleg) > 0 {
			i := eleg[0]
			if op.Semilla != 0 {
				i = eleg[int(op.Semilla%uint64(len(eleg)))]
			}
			return ponerSlots(desnegritar(vars[i], slots, op), slots)
		}
		if r, hay := p.Respaldo[parte]; hay {
			if elegible(r, slots) {
				return ponerSlots(desnegritar(r, slots, op), slots)
			}
			return ""
		}
		ok = false
		return ""
	})
	if !ok {
		return "", false
	}
	return limpiar(s), true
}

// limpiar: quita espacios al final de cada línea, líneas que quedaron con solo espacios y saltos de más.
func limpiar(s string) string {
	ls := strings.Split(s, "\n")
	for i, l := range ls {
		sangria := ""
		if strings.HasPrefix(l, "   ") { // conserva la sangría de los subpasos
			sangria = "   "
		}
		l = espacios.ReplaceAllString(strings.TrimSpace(l), " ") // una parte vacía deja dos espacios
		ls[i] = sangria + l
		if l == "" {
			ls[i] = ""
		}
	}
	s = strings.Join(ls, "\n")
	s = blancos.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

// plantillaYAML: los campos del archivo (ver la cabecera de respuestas.yml).
type plantillaYAML struct {
	Tipo       string              `yaml:"tipo"`
	Texto      string              `yaml:"texto"`
	Forma      string              `yaml:"forma"`
	Partes     map[string][]string `yaml:"partes"`
	Respaldo   map[string]string   `yaml:"respaldo"`
	Fijas      map[string]int      `yaml:"fijas"`
	SinModelo  bool                `yaml:"sin_modelo"`
	MaxFrases  int                 `yaml:"max_frases"`
	Protegidos []string            `yaml:"protegidos"`
	Variantes  []string            `yaml:"variantes"`
}

// CargarPlantillas lee metrin/plantillas/respuestas.yml. Sin archivo: solo las embebidas (sin error).
// Una plantilla inválida se salta (queda la embebida) y se informa con su línea.
func CargarPlantillas(ruta string) (*Plantillas, []ErrorCarga) {
	ps := nuevasPlantillas()
	if ruta == "" {
		return ps, nil
	}
	datos, err := os.ReadFile(ruta)
	if err != nil {
		if os.IsNotExist(err) {
			return ps, nil
		}
		return ps, []ErrorCarga{{Archivo: ruta, Mensaje: err.Error(), Grave: true}}
	}
	ps.Archivo = ruta
	nodo, e := leerNodoYAML(ruta, datos)
	if e != nil {
		return ps, []ErrorCarga{*e}
	}
	if nodo.Kind != yaml.MappingNode {
		return ps, []ErrorCarga{{Archivo: ruta, Linea: nodo.Line, Mensaje: "se esperaba un mapa con «plantillas»", Grave: true}}
	}
	var errs []ErrorCarga
	var lista *yaml.Node
	for i := 0; i+1 < len(nodo.Content); i += 2 {
		switch nodo.Content[i].Value {
		case "plantillas":
			lista = nodo.Content[i+1]
		case "config":
			var cfg struct {
				Max int `yaml:"max_pasos_por_bloque"`
			}
			if nodo.Content[i+1].Decode(&cfg) == nil {
				ps.MaxPasosBloque = cfg.Max
			}
		}
	}
	if lista == nil || lista.Kind != yaml.MappingNode {
		return ps, []ErrorCarga{{Archivo: ruta, Linea: nodo.Line, Mensaje: "falta el mapa «plantillas»: se usan las embebidas", Grave: true}}
	}
	for i := 0; i+1 < len(lista.Content); i += 2 {
		nombre, v := lista.Content[i].Value, lista.Content[i+1]
		var y plantillaYAML
		if v.Kind == yaml.ScalarNode {
			y.Texto = v.Value
		} else if err := v.Decode(&y); err != nil {
			ec := errorYAML(ruta, err)
			ec.Mensaje = "plantilla " + nombre + ": " + ec.Mensaje
			errs = append(errs, ec)
			continue
		}
		p := Plantilla{Nombre: nombre, Tipo: y.Tipo, Texto: y.Texto, Forma: y.Forma, Partes: y.Partes, Respaldo: y.Respaldo,
			Fijas: y.Fijas, SinModelo: y.SinModelo, MaxFrases: y.MaxFrases, Protegidos: y.Protegidos, Origen: "archivo", Linea: v.Line}
		if p.Texto == "" && p.Forma == "" && len(y.Variantes) > 0 {
			p.Texto = y.Variantes[0]
		}
		if p.Texto == "" && p.Forma == "" {
			errs = append(errs, ErrorCarga{Archivo: ruta, Linea: v.Line, Mensaje: "plantilla " + nombre + " sin «forma» ni «texto»: se salta"})
			continue
		}
		// Toda {parte} de la forma debe tener variantes o respaldo.
		var faltan []string
		for _, m := range parteRe.FindAllStringSubmatch(p.Forma, -1) {
			if _, ok := p.Partes[m[1]]; !ok {
				if _, ok := p.Respaldo[m[1]]; !ok {
					faltan = append(faltan, m[1])
				}
			}
		}
		if len(faltan) > 0 {
			errs = append(errs, ErrorCarga{Archivo: ruta, Linea: v.Line, Grave: true,
				Mensaje: fmt.Sprintf("plantilla %s: la forma usa partes sin variantes: %s (se usa la embebida)", nombre, strings.Join(faltan, ", "))})
			continue
		}
		ps.m[nombre] = p
	}
	return ps, errs
}
