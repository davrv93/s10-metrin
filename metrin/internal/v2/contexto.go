package v2

// Context Builder: arma el estado estructurado del turno (tipos.Estado) y la consulta normalizada (tipos.Consulta).
//
// Reglas:
//   - Hechos solo con lo que consta: lo que escribió el usuario (origen «usuario») y la memoria que devolvió el
//     cliente (origen «memoria»). Todo lo deducido —módulo por alias, memoria reconstruida del hilo, referencia a
//     un turno anterior— va a Inferencias con su confianza. Nunca una inferencia como hecho.
//   - La pregunta original no se toca. La normalizada solo pliega mayúsculas y tildes y quita signos sueltos: no
//     quita palabras, así que no se pierden términos exactos («metrado», «APU», «F7»). Los términos del ERP que se
//     reconocen por su forma (siglas, teclas de función, atajos, códigos con puntos, términos entre comillas) y los
//     del Aliaser van además, tal cual, en Entidades.
//   - El hilo se recorta: últimos turnos y textos cortos. El estado viaja al motor de decisión.

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"rag-go/internal/rag"
	"rag-go/internal/v2/tipos"
)

const (
	maxTurnosHilo   = 4
	maxRunasTurno   = 200
	origenUsuario   = "usuario"
	origenHilo      = "hilo"
	origenMemoria   = "memoria"
	origenClasif    = "clasificador"
	origenDecision  = "decision"
	origenBusqueda  = "busqueda" // hecho medido por el código sobre lo recuperado
	origenReglas    = "reglas"   // propuesta de las reglas para una decisión (inferencia)
	claveModulo     = "modulo"
	claveProcCurso  = "procedimiento_en_curso"
	clavePasoActual = "paso_actual"
	claveReferencia = "referencia_hilo"
	prefijoEntidad  = "entidad:"
)

// plegarTildes: minúsculas sin tildes; la ñ se conserva (año ≠ ano), igual que internal/clasificar.
var plegarTildes = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "à", "a", "è", "e", "ì", "i", "ò", "o", "ù", "u")

// Normalizar deja el texto en minúsculas, sin tildes y sin signos sueltos, con los espacios colapsados. Conserva
// las letras, los dígitos, la ñ, el «+» de los atajos («ctrl+f») y el punto entre dígitos («01.02.03»).
func Normalizar(s string) string {
	s = plegarTildes.Replace(strings.ToLower(s))
	r := []rune(s)
	var b strings.Builder
	b.Grow(len(s))
	for i, c := range r {
		switch {
		case unicode.IsLetter(c) || unicode.IsDigit(c) || c == '+':
			b.WriteRune(c)
		case c == '.' && i > 0 && i+1 < len(r) && unicode.IsDigit(r[i-1]) && unicode.IsDigit(r[i+1]):
			b.WriteRune(c)
		default:
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

var (
	reAtajo   = regexp.MustCompile(`(?i)\b(ctrl|control|shift|may[uú]s|alt)(\s*\+\s*[a-z0-9]+)+\b`)
	reTeclaF  = regexp.MustCompile(`(?i)\bF(1[0-2]|[1-9])\b`)
	reSigla   = regexp.MustCompile(`\b[A-ZÁÉÍÓÚÑ][A-ZÁÉÍÓÚÑ0-9]{1,6}\b`)
	reCodigo  = regexp.MustCompile(`\b\d+(\.\d+)+\b`)
	reComilla = regexp.MustCompile(`[«"“']([^«»"“”']{2,60})[»"”']`)
)

// TerminosExactos devuelve, tal como están escritos, los términos que se reconocen por su forma: atajos
// («Ctrl+F»), teclas de función («F7»), siglas («APU», «S10»), códigos con puntos y términos entre comillas.
func TerminosExactos(texto string) []string {
	var out []string
	for _, re := range []*regexp.Regexp{reAtajo, reTeclaF, reSigla, reCodigo} {
		out = append(out, re.FindAllString(texto, -1)...)
	}
	for _, m := range reComilla.FindAllStringSubmatch(texto, -1) {
		out = append(out, strings.TrimSpace(m[1]))
	}
	return unicosOrden(out)
}

// unicosOrden quita repetidos (sin distinguir mayúsculas) conservando el primero que aparece.
func unicosOrden(xs []string) []string {
	vistos := map[string]bool{}
	var out []string
	for _, x := range xs {
		k := strings.ToLower(strings.TrimSpace(x))
		if k == "" || vistos[k] {
			continue
		}
		vistos[k] = true
		out = append(out, strings.TrimSpace(x))
	}
	return out
}

// Reescribir arma la consulta del turno sin modelo: original intacta, normalizada, entidades (términos exactos +
// los del Aliaser) y alias. modulo es el que detectó el Aliaser ("" si nada).
func Reescribir(pregunta string, al Aliaser) tipos.Consulta {
	if al == nil {
		al = SinAlias{}
	}
	ents, alias, modulo := al.Analizar(pregunta)
	return tipos.Consulta{
		Original:    pregunta,
		Normalizada: Normalizar(pregunta),
		Entidades:   unicosOrden(append(TerminosExactos(pregunta), ents...)),
		Aliases:     unicosOrden(alias),
		Modulo:      modulo,
	}
}

// HiloRecortado pasa el hilo de V1 al estado: los últimos turnos de usuario y asistente, con textos cortos.
func HiloRecortado(hilo []rag.Turno) []tipos.Turno {
	var out []tipos.Turno
	for _, t := range hilo {
		if t.Rol != "usuario" && t.Rol != "asistente" {
			continue
		}
		out = append(out, tipos.Turno{Rol: t.Rol, Texto: recortarRunas(t.Texto, maxRunasTurno)})
	}
	if len(out) > maxTurnosHilo {
		out = out[len(out)-maxTurnosHilo:]
	}
	return out
}

func recortarRunas(s string, n int) string {
	r := []rune(strings.Join(strings.Fields(s), " "))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n])
}

// esReferencia: la pregunta apunta a algo del turno anterior («eso», «ahí», «su», «lo mismo»).
func esReferencia(norm string) bool {
	t := " " + norm + " "
	for _, w := range []string{" eso ", " esto ", " ese ", " esa ", " ahi ", " alli ", " aqui ", " su ", " sus ", " lo mismo ", " lo anterior ", " el otro ", " la otra ", " lo de antes "} {
		if strings.Contains(t, w) {
			return true
		}
	}
	return false
}

// ConstruirEstado arma el estado del turno. mem es la memoria del procedimiento en curso y origenMem dice de dónde
// salió: «memoria» (la devolvió el cliente: hecho) u «hilo» (reconstruida del último plan: inferencia).
func ConstruirEstado(pregunta string, hilo []rag.Turno, mem tipos.Memoria, origenMem string, al Aliaser) tipos.Estado {
	if al == nil {
		al = SinAlias{}
	}
	c := Reescribir(pregunta, al)
	e := tipos.Estado{
		Pregunta:    pregunta,
		Consulta:    c,
		Hilo:        HiloRecortado(hilo),
		Tipo:        tipos.Desconocido,
		Hechos:      map[string]tipos.Dato{},
		Inferencias: map[string]tipos.Dato{},
		Memoria:     mem,
	}
	e.Hechos["pregunta"] = tipos.Dato{Valor: pregunta, Origen: origenUsuario, Confianza: 1}
	for _, ent := range c.Entidades {
		if contieneTermino(c.Normalizada, ent) {
			e.Hechos[prefijoEntidad+ent] = tipos.Dato{Valor: ent, Origen: origenUsuario, Confianza: 1}
		} else {
			// El Aliaser la dedujo de un sinónimo: el usuario no la escribió.
			e.Inferencias[prefijoEntidad+ent] = tipos.Dato{Valor: ent, Origen: origenClasif, Confianza: 0.8}
		}
	}
	if c.Modulo != "" {
		if contieneTermino(c.Normalizada, c.Modulo) {
			e.Hechos[claveModulo] = tipos.Dato{Valor: c.Modulo, Origen: origenUsuario, Confianza: 1}
		} else {
			e.Inferencias[claveModulo] = tipos.Dato{Valor: c.Modulo, Origen: origenClasif, Confianza: 0.8}
		}
	}
	if mem.ProcedimientoID != "" {
		paso := tipos.Dato{Valor: strconv.Itoa(mem.PasoActual), Origen: origenMem}
		proc := tipos.Dato{Valor: mem.ProcedimientoID, Origen: origenMem}
		if origenMem == origenMemoria {
			proc.Confianza, paso.Confianza = 1, 1
			e.Hechos[claveProcCurso], e.Hechos[clavePasoActual] = proc, paso
		} else {
			proc.Origen, paso.Origen = origenHilo, origenHilo
			proc.Confianza, paso.Confianza = 0.9, 0.9
			e.Inferencias[claveProcCurso], e.Inferencias[clavePasoActual] = proc, paso
		}
	}
	// Referencia a un turno anterior: se infiere (no se afirma) a qué apunta, y sus términos solo amplían la
	// búsqueda como alias; la consulta del usuario no cambia.
	if esReferencia(c.Normalizada) {
		for i := len(hilo) - 1; i >= 0; i-- {
			if hilo[i].Rol != "usuario" || strings.TrimSpace(hilo[i].Texto) == "" {
				continue
			}
			e.Inferencias[claveReferencia] = tipos.Dato{Valor: recortarRunas(hilo[i].Texto, 120), Origen: origenHilo, Confianza: 0.6}
			ents, alias, _ := al.Analizar(hilo[i].Texto)
			e.Consulta.Aliases = unicosOrden(append(append(e.Consulta.Aliases, TerminosExactos(hilo[i].Texto)...), append(ents, alias...)...))
			break
		}
	}
	e.Desconocidos = desconocidos(e)
	return e
}

// desconocidos: lo que el turno no dice y la respuesta podría necesitar.
func desconocidos(e tipos.Estado) []string {
	var out []string
	if _, h := e.Hechos[claveModulo]; !h {
		if _, i := e.Inferencias[claveModulo]; !i {
			out = append(out, claveModulo)
		}
	}
	if e.Memoria.ProcedimientoID == "" {
		out = append(out, "procedimiento")
	}
	return out
}

// contieneTermino: el término aparece en el texto normalizado como palabra(s) completa(s).
func contieneTermino(norm, termino string) bool {
	t := Normalizar(termino)
	if t == "" {
		return false
	}
	return strings.Contains(" "+norm+" ", " "+t+" ")
}
