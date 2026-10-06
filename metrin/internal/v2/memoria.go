package v2

// Memoria del procedimiento en curso (tipos.Memoria). Viaja en la petición y en la respuesta («memoria»); si el
// cliente no la manda, se reconstruye del último «plan» del hilo y, si no hay, queda vacía.
//
//   - «¿y luego?», «siguiente», «listo», «ya» → el paso siguiente del MISMO procedimiento, sin búsqueda nueva.
//   - «no me sale», «error», «no me deja» → los errores frecuentes del procedimiento, con el paso actual a la vista.
//   - Otra pregunta con un procedimiento en curso → se contesta, el procedimiento queda suspendido y se ofrece
//     retomarlo UNA vez (plantilla RETOMA). «sí», «listo» o «retomemos» lo retoman; si la persona insiste en el
//     tema nuevo (otra pregunta), se abandona. Un procedimiento ya retomado no se vuelve a suspender: se abandona.
//   - «¿y luego?» sin procedimiento en curso → se pide aclaración; nunca una búsqueda a ciegas.
//   - Tras ofrecer un procedimiento (un concepto con su procedimiento relacionado, o una aclaración con UNA opción),
//     «sí», «sí, enséñame» o «¿y luego?» lo enseñan desde el paso 1, sin búsqueda (memoria.ofrecido).
//   - La memoria guarda también el concepto explicado y lo ofrecido (memoria.concepto, memoria.ofrecido) para
//     resolver una pregunta elíptica («¿y cómo la hago?», «¿y eso dónde está?»): ver referencia.go.

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"rag-go/internal/rag"
	"rag-go/internal/v2/tipos"
)

// AccionMemoria: lo que la memoria hace con el mensaje.
type AccionMemoria string

const (
	MemNinguna          AccionMemoria = "ninguna"           // pregunta nueva, sin procedimiento en curso
	MemSiguiente        AccionMemoria = "siguiente"         // paso siguiente del procedimiento en curso
	MemErrorPaso        AccionMemoria = "error_paso"        // errores frecuentes del paso en curso
	MemRetomar          AccionMemoria = "retomar"           // vuelve al procedimiento suspendido
	MemCambioTema       AccionMemoria = "cambio_tema"       // pregunta nueva: se suspende el procedimiento
	MemAbandonar        AccionMemoria = "abandonar"         // insiste en el tema nuevo: se abandona el anterior
	MemSinProcedimiento AccionMemoria = "sin_procedimiento" // «¿y luego?» sin nada en curso: aclarar
	MemOfrecido         AccionMemoria = "ofrecido"          // «sí» o «¿y luego?» tras ofrecer un procedimiento: desde el paso 1
)

// Frases de avance explícito: piden el paso siguiente aunque no haya nada en curso.
var avanceExplicito = conjuntoFrases(`y luego|luego|y despues|despues|siguiente|el siguiente|siguiente paso|
el siguiente paso|paso siguiente|que sigue|y que sigue|que sigue ahora|y ahora|ahora que|y ahora que|y luego que|
y despues que|que hago ahora|y ahora que hago|que mas|y que mas|continua|continuar|continuemos|sigue|seguimos|
sigamos|el otro paso|otro paso|proximo paso|el proximo`)

// Acuses: con un procedimiento en curso significan «hecho, sigue»; sin él son charla.
var acuses = conjuntoFrases(`listo|lista|listos|ya|ya esta|ya lo hice|ya quedo|ya hice eso|hecho|ok|okey|oki|
ok listo|dale|terminado|termine|perfecto|listo gracias|ok gracias|ya termine|bien|hecho gracias`)

// Afirmaciones que aceptan la oferta de retomar.
var afirmaciones = conjuntoFrases(`si|si por favor|si porfa|claro|de acuerdo|va|ok|dale|bueno|si claro|si dale|
si retomemos|si sigamos|por favor`)

// Prefijos de mensajes cortos que también son avance («listo, ¿y luego?», «ya lo hice, ¿qué sigue?»).
var prefijosAvance = []string{"listo ", "ya ", "y luego", "siguiente", "que sigue", "y despues", "hecho ", "ok listo",
	"ok ya", "ok y luego", "ok siguiente"}

var frasesErrorPaso = []string{"no me sale", "no sale", "me sale error", "me salio error", "me sale un error",
	"me salio un error", "sale error", "error", "no funciona", "no me deja", "no aparece", "no me aparece",
	"no puedo", "no encuentro", "no lo encuentro", "falla", "no carga", "no guarda", "no se guarda", "no me guarda",
	"no se ve", "no veo", "no hay ese", "no existe esa", "no existe ese"}

var frasesRetomar = []string{"retom", "volvamos", "volver al", "volver a lo", "sigamos con", "continuemos con",
	"regresemos", "regresar al", "lo de antes", "lo anterior"}

func conjuntoFrases(s string) map[string]bool {
	m := map[string]bool{}
	for _, f := range strings.Split(strings.ReplaceAll(s, "\n", ""), "|") {
		if f = strings.TrimSpace(f); f != "" {
			m[f] = true
		}
	}
	return m
}

func esAvanceExplicito(n string) bool {
	if avanceExplicito[n] {
		return true
	}
	if len(strings.Fields(n)) > 6 {
		return false
	}
	for _, p := range []string{"y luego", "siguiente", "que sigue", "y despues", "y ahora que"} {
		if strings.HasPrefix(n, p) || strings.HasSuffix(n, p) {
			return true
		}
	}
	return false
}

// esContinuar: avance explícito o acuse corto («listo», «ya», «ok listo», «listo, ¿y luego?»).
func esContinuar(n string) bool {
	if esAvanceExplicito(n) || acuses[n] {
		return true
	}
	if len(strings.Fields(n)) > 6 {
		return false
	}
	for _, p := range prefijosAvance {
		if strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

func esErrorPaso(n string) bool {
	t := " " + n + " "
	for _, f := range frasesErrorPaso {
		if strings.Contains(t, " "+f) {
			return true
		}
	}
	return false
}

func esRetomarExplicito(n string) bool {
	for _, f := range frasesRetomar {
		if strings.Contains(" "+n, " "+f) {
			return true
		}
	}
	return false
}

// InterpretarMemoria decide qué hace la memoria con un mensaje. No busca ni llama a ningún modelo.
func InterpretarMemoria(texto string, m tipos.Memoria) AccionMemoria {
	n := Normalizar(texto)
	activa := m.ProcedimientoID != ""
	suspendido := m.Suspendido != nil && m.Suspendido.ProcedimientoID != ""
	switch {
	case suspendido && esRetomarExplicito(n):
		return MemRetomar
	case activa && esContinuar(n):
		return MemSiguiente
	case activa && esErrorPaso(n):
		return MemErrorPaso
	case !activa && m.Ofrecido != "" && AceptaOferta(n):
		return MemOfrecido
	case !activa && suspendido && (esContinuar(n) || afirmaciones[n]):
		return MemRetomar
	case !activa && esAvanceExplicito(n):
		return MemSinProcedimiento
	case activa && m.Retomado:
		return MemAbandonar // ya se retomó una vez: el tema nuevo manda
	case activa:
		return MemCambioTema
	case suspendido:
		return MemAbandonar
	}
	return MemNinguna
}

// Suspender guarda el procedimiento en curso para ofrecer retomarlo una vez.
func Suspender(m tipos.Memoria) tipos.Memoria {
	s := m
	s.Suspendido, s.Retomado = nil, false
	s.Concepto, s.Ofrecido = "", "" // lo suspendido es el procedimiento, no lo que se ofreció
	s.PasosCompletados = append([]int(nil), m.PasosCompletados...)
	return tipos.Memoria{Suspendido: &s}
}

// Retomar vuelve al procedimiento suspendido (marcado como ya retomado: no se ofrece otra vez).
func Retomar(m tipos.Memoria) tipos.Memoria {
	if m.Suspendido == nil {
		return m
	}
	s := *m.Suspendido
	s.Suspendido, s.Retomado = nil, true
	return s
}

// Avanzar marca como completados los pasos hasta el actual y pasa al siguiente. terminado = no quedan pasos.
func Avanzar(m tipos.Memoria, total int) (tipos.Memoria, bool) {
	nm := m
	nm.PasosCompletados = completarHasta(m.PasosCompletados, m.PasoActual)
	if m.PasoActual >= total {
		nm.PasoActual = total
		return nm, true
	}
	nm.PasoActual = m.PasoActual + 1
	return nm, false
}

func completarHasta(hechos []int, n int) []int {
	set := map[int]bool{}
	for _, x := range hechos {
		set[x] = true
	}
	for i := 1; i <= n; i++ {
		set[i] = true
	}
	out := make([]int, 0, len(set))
	for x := range set {
		if x > 0 {
			out = append(out, x)
		}
	}
	sort.Ints(out)
	return out
}

// MemoriaDePlan: la memoria que deja una respuesta. Con un procedimiento y pasos a la vista, el paso actual es el
// último mostrado; si no, sin procedimiento. previa conserva los completados del mismo procedimiento. Además, el
// concepto explicado (uno solo) y el procedimiento ofrecido (plan.opciones con UNA opción), si los hay.
func MemoriaDePlan(p tipos.Plan, previa tipos.Memoria) tipos.Memoria {
	m := memoriaProcedimiento(p, previa)
	if p.SinEvidencia {
		return m
	}
	if p.Concepto != nil && p.Concepto.ID != "" && len(p.Conceptos) <= 1 {
		m.Concepto = p.Concepto.ID
	}
	if len(p.Opciones) == 1 && p.Opciones[0].ID != "" && p.Opciones[0].ID != m.ProcedimientoID {
		m.Ofrecido = p.Opciones[0].ID
	}
	return m
}

// memoriaProcedimiento: el procedimiento en curso que deja el plan (vacía si no deja ninguno).
func memoriaProcedimiento(p tipos.Plan, previa tipos.Memoria) tipos.Memoria {
	if p.SinEvidencia || p.Procedimiento == nil || p.Procedimiento.ID == "" || p.Plantilla == "PROCEDIMIENTO_FIN" {
		return tipos.Memoria{}
	}
	mostrados := append([]int(nil), p.PasosMostrados...)
	if len(mostrados) == 0 {
		for _, x := range p.Pasos {
			mostrados = append(mostrados, x.N)
		}
	}
	if len(mostrados) == 0 {
		return tipos.Memoria{}
	}
	actual, primero := mostrados[0], mostrados[0]
	for _, n := range mostrados {
		actual, primero = max(actual, n), min(primero, n)
	}
	m := tipos.Memoria{ProcedimientoID: p.Procedimiento.ID, PasoActual: actual}
	if previa.ProcedimientoID == p.Procedimiento.ID {
		m.PasosCompletados = append([]int(nil), previa.PasosCompletados...)
	}
	m.PasosCompletados = completarHasta(m.PasosCompletados, primero-1)
	if len(m.PasosCompletados) == 0 {
		m.PasosCompletados = nil
	}
	return m
}

// MemoriaDeHilo reconstruye la memoria del último plan del asistente en el hilo. ok = había un plan legible.
func MemoriaDeHilo(hilo []rag.Turno) (tipos.Memoria, bool) {
	for i := len(hilo) - 1; i >= 0; i-- {
		if hilo[i].Rol != "asistente" || len(hilo[i].Plan) == 0 || string(hilo[i].Plan) == "null" {
			continue
		}
		var p tipos.Plan
		if json.Unmarshal(hilo[i].Plan, &p) != nil {
			continue
		}
		return MemoriaDePlan(p, tipos.Memoria{}), true
	}
	return tipos.Memoria{}, false
}

// ---------------------------------------------------------------------------------------------
// Planes que salen de la memoria (sin búsqueda): todo viene del procedimiento cargado.

func refDe(def tipos.ProcedimientoDef) *tipos.Ref {
	return &tipos.Ref{ID: def.ID, Titulo: def.Titulo, Confianza: 1}
}

// pasoDe busca el paso n del procedimiento (por N; si no, por posición).
func pasoDe(def tipos.ProcedimientoDef, n int) (tipos.Paso, bool) {
	for _, p := range def.Pasos {
		if p.N == n {
			return p, true
		}
	}
	if n >= 1 && n <= len(def.Pasos) {
		return def.Pasos[n-1], true
	}
	return tipos.Paso{}, false
}

func pasoPlan(p tipos.Paso, n int) tipos.PasoPlan {
	if p.N > 0 {
		n = p.N
	}
	return tipos.PasoPlan{N: n, ID: p.ID, Texto: p.Accion, Fotos: append([]tipos.Foto(nil), p.Fotos...),
		Fuente: append([]string(nil), p.Fuente...), Pagina: p.Pagina}
}

// fuentesPlan agrupa por manual (y sección) las fuentes del procedimiento citadas por ids; sin ids, todas.
func fuentesPlan(def tipos.ProcedimientoDef, ids []string) []tipos.FuentePlan {
	quiero := map[string]bool{}
	for _, id := range ids {
		quiero[id] = true
	}
	var out []tipos.FuentePlan
	pos := map[string]int{}
	for _, f := range def.Fuentes {
		if len(ids) > 0 && !quiero[f.ID] {
			continue
		}
		clave := f.Manual + "\x00" + f.Seccion
		i, ok := pos[clave]
		if !ok {
			i = len(out)
			pos[clave] = i
			out = append(out, tipos.FuentePlan{Manual: f.Manual, Seccion: f.Seccion})
		}
		if f.Pagina > 0 && !contieneInt(out[i].Paginas, f.Pagina) {
			out[i].Paginas = append(out[i].Paginas, f.Pagina)
			sort.Ints(out[i].Paginas)
		}
	}
	if len(out) == 0 && len(ids) > 0 {
		return fuentesPlan(def, nil)
	}
	if out == nil {
		out = []tipos.FuentePlan{}
	}
	return out
}

func contieneInt(xs []int, x int) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// pasosPlan: todos los pasos del procedimiento (el gate compara el plan con el procedimiento completo); la parte
// que se enseña va en PasosMostrados.
func pasosPlan(def tipos.ProcedimientoDef) []tipos.PasoPlan {
	out := make([]tipos.PasoPlan, 0, len(def.Pasos))
	for i, p := range def.Pasos {
		out = append(out, pasoPlan(p, i+1))
	}
	return out
}

// PlanPaso: el procedimiento con el paso n a la vista (continuación «¿y luego?» o retomar) cuando el Constructor
// no sabe continuar. Mismo formato que el Constructor: todos los pasos en Pasos y [n] en PasosMostrados.
func PlanPaso(def tipos.ProcedimientoDef, n int) (tipos.Plan, bool) {
	p, ok := pasoDe(def, n)
	if !ok {
		return tipos.Plan{}, false
	}
	pp := pasoPlan(p, n)
	total := len(def.Pasos)
	plan := tipos.Plan{
		Version: 2, Tipo: tipos.Procedimiento, Procedimiento: refDe(def),
		Pasos: pasosPlan(def), PasosMostrados: []int{pp.N},
		Fuentes: fuentesPlan(def, nil), Plantilla: "PASO",
	}
	if pp.N < total {
		plan.Siguiente = &tipos.Siguiente{Plantilla: "PREGUNTA_AVANCE", Paso: pp.N + 1,
			Texto: "Cuando lo tenga, dígame «listo» y le doy el paso " + strconv.Itoa(pp.N+1) + " de " + strconv.Itoa(total) + "."}
	} else {
		plan.Verificacion = append([]tipos.ConFuente(nil), def.Verificacion...)
		plan.Siguiente = &tipos.Siguiente{Plantilla: "VERIFICACION", Texto: "Es el último paso de «" + def.Titulo + "»."}
	}
	return plan, true
}

// PlanFin: el procedimiento terminó; se muestra la verificación documentada.
func PlanFin(def tipos.ProcedimientoDef) tipos.Plan {
	var ids []string
	for _, v := range def.Verificacion {
		ids = append(ids, v.Fuente...)
	}
	return tipos.Plan{
		Version: 2, Tipo: tipos.Procedimiento, Procedimiento: refDe(def),
		Verificacion: append([]tipos.ConFuente(nil), def.Verificacion...),
		Fuentes:      fuentesPlan(def, ids), Plantilla: "PROCEDIMIENTO_FIN",
		Siguiente: &tipos.Siguiente{Plantilla: "PROCEDIMIENTO_FIN",
			Texto: "Terminaste «" + def.Titulo + "». ¿Te ayudo con otra tarea?"},
	}
}

// PlanErrores: «no me sale» en el paso n. Los errores frecuentes documentados del procedimiento, con el paso a la
// vista; si no hay ninguno documentado, los prerrequisitos (plantilla VERIFICACION). Nada inventado.
func PlanErrores(def tipos.ProcedimientoDef, n int) tipos.Plan {
	plan := tipos.Plan{Version: 2, Tipo: tipos.Problema, Procedimiento: refDe(def), Plantilla: "ERROR_FRECUENTE"}
	var ids []string
	if p, ok := pasoDe(def, n); ok {
		pp := pasoPlan(p, n)
		plan.Pasos, plan.PasosMostrados = pasosPlan(def), []int{pp.N}
		ids = append(ids, pp.Fuente...)
	}
	plan.Errores = append([]tipos.ErrorFrecuente(nil), def.ErroresFrecuentes...)
	for _, e := range plan.Errores {
		ids = append(ids, e.Fuente...)
	}
	if len(plan.Errores) == 0 {
		plan.Plantilla = "VERIFICACION"
		plan.Prerrequisitos = append([]tipos.ConFuente(nil), def.Prerrequisitos...)
		for _, p := range plan.Prerrequisitos {
			ids = append(ids, p.Fuente...)
		}
	}
	plan.Fuentes = fuentesPlan(def, ids)
	plan.Siguiente = &tipos.Siguiente{Plantilla: "PREGUNTA_AVANCE", Paso: n,
		Texto: "Cuando se resuelva, dígame «listo» y seguimos."}
	return plan
}

// PlanAclarar: no hay tarea identificable; se pregunta en vez de adivinar (ACLARAR_TAREA).
func PlanAclarar(texto string) tipos.Plan {
	return tipos.Plan{Version: 2, Tipo: tipos.Desconocido, Fuentes: []tipos.FuentePlan{}, Plantilla: "ACLARAR_TAREA",
		Siguiente: &tipos.Siguiente{Plantilla: "ACLARAR_TAREA", Texto: texto}}
}

// OfertaRetoma: el texto con el que se ofrece volver al procedimiento suspendido.
func OfertaRetoma(titulo string, paso int) *tipos.Siguiente {
	t := "¿Retomamos «" + titulo + "»"
	if paso > 0 {
		t += " en el paso " + strconv.Itoa(paso)
	}
	return &tipos.Siguiente{Plantilla: "RETOMA", Paso: paso, Texto: t + "?"}
}
