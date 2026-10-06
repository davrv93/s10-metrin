package v2

// Orquestador V2 con recursividad controlada (docs/V2-RAG-PROCEDURAL.md §3):
//
//	memoria → contexto → tipo (→ decisión) → plan de consulta → recuperar → ¿evidencia suficiente?
//	  (no → expandir la consulta y buscar otra vez, máx. MAX_SEARCH_ITERATIONS → SIN_EVIDENCIA)
//	→ ¿qué procedimiento? (→ decisión) → construir evidencia y plan → redactar (LLM o plantillas)
//	→ quality gate → regenerar una vez sin LLM → respuesta segura
//
// Prioridad: CORRECCIÓN > EVIDENCIA > PASOS > IMAGEN > UTILIDAD > CLARIDAD > VELOCIDAD. Cada vuelta gasta del
// presupuesto (presupuesto.go); al agotarse se responde con lo respaldado o SIN_EVIDENCIA, nunca en bucle. Nada
// falla en silencio: cada degradación queda en «motivo» y en la traza (etapas_v2).

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"rag-go/internal/rag"
	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

// Charlador es la charla directa de V1 (rag.RAG.Conversar): la V2 la usa para los mensajes SOCIAL.
type Charlador interface {
	Conversar(ctx context.Context, intencion, pregunta string, hilo []rag.Turno) (rag.Respuesta, error)
}

// Agente es el orquestador V2. Todos los motores son opcionales: lo que falte se degrada (y se anota).
type Agente struct {
	Config       Config
	Clasificador *Clasificador          // nil = solo reglas léxicas
	Aliaser      Aliaser                // nil = SinAlias
	Recuperador  tipos.Recuperador      // nil = sin búsqueda: todo termina en SIN_EVIDENCIA
	Constructor  Constructor            // nil = sin plan: SIN_EVIDENCIA
	Catalogo     Catalogo               // nil = la memoria no puede continuar un procedimiento
	Generador    tipos.GenerationEngine // LLM local que reformula; nil = plantillas
	Plantillas   tipos.GenerationEngine // sin modelo; nil = Basico
	Gate         tipos.QualityGate      // nil = GateMinimo
	Decision     tipos.DecisionEngine   // nil = Reglas
	Charla       Charlador              // nil = saludo fijo
}

// Nuevo crea un agente con la configuración y los respaldos por defecto.
func Nuevo(c Config) *Agente { return &Agente{Config: c} }

// Version decide la versión del turno (implementa rag.AgenteV2).
func (a *Agente) Version(pregunta string, o rag.Opciones) tipos.Version {
	return a.Config.Elegir(o.Version, ClaveTurno(pregunta, o))
}

func (a *Agente) aliaser() Aliaser {
	if a.Aliaser == nil {
		return SinAlias{}
	}
	return a.Aliaser
}

func (a *Agente) plantillas() tipos.GenerationEngine {
	if a.Plantillas == nil {
		return Basico{}
	}
	return a.Plantillas
}

func (a *Agente) gate() (tipos.QualityGate, string) {
	if a.Gate == nil {
		return GateMinimo{}, "minimo"
	}
	return a.Gate, "conocimiento"
}

func (a *Agente) motor() tipos.DecisionEngine {
	if a.Decision == nil {
		return Reglas{}
	}
	return a.Decision
}

// turno: el estado de una pregunta (el Agente se comparte entre peticiones; esto no).
type turno struct {
	a         *Agente
	ctx       context.Context
	pregunta  string
	o         rag.Opciones
	v         *traza.RecorderV2 // nil sin modo traza
	rec       *traza.Recorder
	pres      *Presupuesto
	estado    tipos.Estado
	accion    AccionMemoria
	origenMem string      // memoria (la mandó el cliente) | hilo (reconstruida del último plan)
	ref       *Referencia // pregunta elíptica resuelta (referencia.go); nil si no lo es
	tipoRes   ResultadoTipo
	metodo    string // cómo se decidió el tipo (para «orquestacion.clasificador»)

	decisiones       []map[string]any
	msDecision       time.Duration
	respaldoDecision string // por qué decidió Reglas en lugar del motor ("" = no hubo respaldo)
	msBusqueda       time.Duration
	msLLM            time.Duration
	degradado        []string
	derivoBusqueda   bool
	altProcedimiento []string
	decididoProc     string
	gateIntentos     int
	generador        string
	modoRedaccion    string
}

// Preguntar responde un turno V2 (implementa rag.AgenteV2). No devuelve error: toda falla termina en una
// respuesta segura con su motivo y su traza.
func (a *Agente) Preguntar(ctx context.Context, pregunta string, o rag.Opciones) (rag.Respuesta, error) {
	rec := traza.De(ctx)
	o.Hilo = HiloSinEco(pregunta, o.Hilo)
	t := &turno{a: a, ctx: ctx, pregunta: pregunta, o: o, rec: rec, v: rec.V2(), pres: NuevoPresupuesto(a.Config.Limites)}
	res := t.responder()
	t.cerrarTraza(res)
	return res, nil
}

// HiloSinEco quita del final del hilo el turno de usuario que repite la pregunta actual: la página lo añade desde el
// primer turno y, si no, la pregunta se tomaría por seguimiento de sí misma.
func HiloSinEco(pregunta string, hilo []rag.Turno) []rag.Turno {
	if n := len(hilo); n > 0 && hilo[n-1].Rol == "usuario" &&
		strings.Join(strings.Fields(hilo[n-1].Texto), " ") == strings.Join(strings.Fields(pregunta), " ") {
		return hilo[:n-1]
	}
	return hilo
}

// memoriaEntrada: la de la petición (hecho) o la reconstruida del último plan del hilo (inferencia).
func (t *turno) memoriaEntrada() (tipos.Memoria, string) {
	if t.o.Memoria != nil {
		return *t.o.Memoria, origenMemoria
	}
	if m, ok := MemoriaDeHilo(t.o.Hilo); ok {
		return m, origenHilo
	}
	return tipos.Memoria{}, origenMemoria
}

func (t *turno) responder() rag.Respuesta {
	mem, origen := t.memoriaEntrada()
	t.origenMem = origen
	t.estado = ConstruirEstado(t.pregunta, t.o.Hilo, mem, origen, t.a.aliaser())
	t.accion = InterpretarMemoria(t.pregunta, mem)
	switch t.accion {
	case MemSiguiente:
		return t.continuar(mem)
	case MemOfrecido:
		return t.empezarOfrecido(mem)
	case MemErrorPaso:
		return t.erroresPaso(mem)
	case MemRetomar:
		return t.retomar(mem)
	case MemSinProcedimiento:
		t.trazarTipoMemoria(tipos.Desconocido, "«¿y luego?» sin procedimiento en curso: se pide aclaración")
		return t.aclarar(TextoSinProcedimiento, tipos.Memoria{}, "continuación sin procedimiento en curso")
	}
	return t.preguntaNueva(mem)
}

// preguntaNueva: el camino completo (tipo, búsqueda, plan, redacción, gate). Una pregunta elíptica («¿y cómo la
// hago?») se resuelve antes con la memoria y el hilo (referencia.go).
func (t *turno) preguntaNueva(previa tipos.Memoria) rag.Respuesta {
	ref, eliptica := ResolverReferencia(t.pregunta, previa, t.origenMem, t.o.Hilo, t.a.aliaser(), t.a.Catalogo)
	tipo := t.clasificarTipo()
	if eliptica {
		tipo = t.aplicarReferencia(ref, tipo)
	}
	// Cortesía o duda con un procedimiento en curso: no lo suspende ni lo abandona.
	base := previa
	if tipo != tipos.Social && tipo != tipos.Desconocido {
		switch t.accion {
		case MemCambioTema:
			base = Suspender(previa)
		case MemAbandonar:
			base = tipos.Memoria{}
		}
	}
	t.estado.Memoria = base
	switch {
	case tipo == tipos.Social:
		return t.social(previa)
	case t.ref != nil && !t.ref.Hay() && len(t.ref.Propias) == 0:
		return t.aclarar(TextoSinReferente, base, "pregunta elíptica sin nada a qué referirse: se pide aclaración")
	case tipo == tipos.Desconocido:
		return t.aclarar(TextoAclarar, base, "tipo de respuesta desconocido: se pide aclaración")
	}
	var cands []tipos.Candidato
	if t.ref != nil && t.ref.Directa() {
		cands = t.candidatosReferencia(tipo)
	}
	if len(cands) > 0 {
		t.trazarReferenciaDirecta(cands)
	} else {
		var motivo string
		var ok bool
		cands, motivo, ok = t.buscar(tipo)
		t.trazarReferenciaBusqueda()
		if !ok {
			return t.entregar(PlanSinEvidencia(tipo), base, motivo)
		}
	}
	cands = t.elegirProcedimiento(cands)
	plan, motivo, ok := t.construir(cands)
	if !ok {
		return t.entregar(PlanSinEvidencia(tipo), base, motivo)
	}
	if plan.SinEvidencia {
		motivo = "el constructor no halló evidencia que respalde un plan entre " + itoa(len(cands)) + " candidato(s)"
	}
	return t.entregar(plan, base, motivo)
}

// ---------------------------------------------------------------------------------------------
// Memoria (sin búsqueda)

func (t *turno) procedimiento(id string) (tipos.ProcedimientoDef, bool) {
	if t.a.Catalogo == nil || id == "" {
		return tipos.ProcedimientoDef{}, false
	}
	return t.a.Catalogo.Procedimiento(id)
}

func (t *turno) continuar(mem tipos.Memoria) rag.Respuesta {
	t.trazarTipoMemoria(tipos.Procedimiento, "memoria: «¿y luego?» → paso siguiente del mismo procedimiento, sin búsqueda")
	def, ok := t.procedimiento(mem.ProcedimientoID)
	if !ok {
		return t.aclarar(TextoSinProcedimiento, tipos.Memoria{}, "el procedimiento en curso («"+mem.ProcedimientoID+"») no está cargado")
	}
	if err := t.pres.Paso(); err != nil {
		return t.segura(PlanSinEvidencia(tipos.Procedimiento), mem, err.Error())
	}
	if c, ok := t.a.Constructor.(Continuador); ok {
		if plan, err := c.Continuar(mem); err == nil {
			return t.entregar(plan, mem, "")
		} else {
			t.degradado = appendUnico(t.degradado, "continuar: "+traza.ResumirError(err)+" → plan del catálogo")
		}
	}
	nm, terminado := Avanzar(mem, len(def.Pasos))
	if terminado {
		return t.entregar(PlanFin(def), nm, "")
	}
	plan, ok := PlanPaso(def, nm.PasoActual)
	if !ok {
		return t.entregar(PlanFin(def), nm, "")
	}
	return t.entregar(plan, nm, "")
}

// empezarOfrecido: «sí», «sí, enséñame» o «¿y luego?» tras ofrecer un procedimiento (un concepto con su
// procedimiento relacionado, o una aclaración con una sola opción): ese procedimiento desde el paso 1, sin búsqueda.
func (t *turno) empezarOfrecido(mem tipos.Memoria) rag.Respuesta {
	t.trazarTipoMemoria(tipos.Procedimiento, "memoria: acepta lo ofrecido («"+mem.Ofrecido+"») → desde el paso 1, sin búsqueda")
	base := tipos.Memoria{Suspendido: mem.Suspendido}
	def, ok := t.procedimiento(mem.Ofrecido)
	if !ok {
		return t.aclarar(TextoSinProcedimiento, base, "el procedimiento ofrecido («"+mem.Ofrecido+"») no está cargado")
	}
	if err := t.pres.Paso(); err != nil {
		return t.segura(PlanSinEvidencia(tipos.Procedimiento), base, err.Error())
	}
	if c, ok := t.a.Constructor.(Continuador); ok {
		if plan, err := c.Continuar(tipos.Memoria{ProcedimientoID: def.ID}); err == nil {
			return t.entregar(plan, base, "")
		} else {
			t.degradado = appendUnico(t.degradado, "continuar: "+traza.ResumirError(err)+" → plan del catálogo")
		}
	}
	plan, ok := PlanPaso(def, 1)
	if !ok {
		return t.entregar(PlanFin(def), base, "")
	}
	return t.entregar(plan, base, "")
}

func (t *turno) erroresPaso(mem tipos.Memoria) rag.Respuesta {
	t.trazarTipoMemoria(tipos.Problema, "memoria: «no me sale» → errores frecuentes del procedimiento en curso")
	def, ok := t.procedimiento(mem.ProcedimientoID)
	if !ok {
		return t.aclarar(TextoSinProcedimiento, tipos.Memoria{}, "el procedimiento en curso («"+mem.ProcedimientoID+"») no está cargado")
	}
	if err := t.pres.Paso(); err != nil {
		return t.segura(PlanSinEvidencia(tipos.Problema), mem, err.Error())
	}
	return t.entregar(PlanErrores(def, max(mem.PasoActual, 1)), mem, "")
}

func (t *turno) retomar(mem tipos.Memoria) rag.Respuesta {
	t.trazarTipoMemoria(tipos.Procedimiento, "memoria: se retoma el procedimiento suspendido (una sola vez)")
	r := Retomar(mem)
	def, ok := t.procedimiento(r.ProcedimientoID)
	if !ok {
		return t.aclarar(TextoSinProcedimiento, tipos.Memoria{}, "el procedimiento suspendido («"+r.ProcedimientoID+"») no está cargado")
	}
	if err := t.pres.Paso(); err != nil {
		return t.segura(PlanSinEvidencia(tipos.Procedimiento), r, err.Error())
	}
	desde := max(r.PasoActual, 1)
	if c, ok := t.a.Constructor.(Continuador); ok {
		previo := r
		previo.PasoActual = desde - 1 // Continuar enseña desde el paso siguiente al ya mostrado
		if plan, err := c.Continuar(previo); err == nil {
			return t.entregar(plan, r, "")
		} else {
			t.degradado = appendUnico(t.degradado, "continuar: "+traza.ResumirError(err)+" → plan del catálogo")
		}
	}
	plan, ok := PlanPaso(def, desde)
	if !ok {
		return t.entregar(PlanFin(def), r, "")
	}
	return t.entregar(plan, r, "")
}

func (t *turno) aclarar(texto string, base tipos.Memoria, motivo string) rag.Respuesta {
	return t.entregar(PlanAclarar(texto), base, motivo)
}

// ---------------------------------------------------------------------------------------------
// Tipo de respuesta y decisiones

func (t *turno) clasificarTipo() tipos.TipoRespuesta {
	t.v.Inicio(traza.EtapaV2TipoRespuesta)
	t0 := time.Now()
	r := t.a.Clasificador.Clasificar(t.ctx, t.pregunta)
	t.tipoRes = r
	final, conf, metodo := r.Tipo, r.Confianza, r.Metodo
	if r.Dudoso() {
		// Opciones: las reglas que chocaron, o las mejores del kNN, o todas; siempre con UNKNOWN. La propuesta de
		// las reglas (lo que sale si el motor falla) es UNKNOWN: Metrín no adivina, pregunta.
		opciones := tiposATexto(r.Candidatos)
		if len(opciones) == 0 {
			opciones = tiposATexto(TiposRespuesta)
		}
		if !contiene(opciones, string(tipos.Desconocido)) {
			opciones = append(opciones, string(tipos.Desconocido))
		}
		d := t.decidir(DecTipoRespuesta, opciones, string(tipos.Desconocido), 0.5)
		final, metodo = tipos.TipoRespuesta(d.Eleccion), "decision:"+d.Fuente
		conf = d.Confianza
	}
	t.metodo = metodo
	t.estado.Tipo, t.estado.TipoConfianza = final, conf
	origen := origenClasif
	if strings.HasPrefix(metodo, "decision:") {
		origen = origenDecision
	}
	t.estado.Inferencias["tipo"] = tipos.Dato{Valor: string(final), Origen: origen, Confianza: conf}
	if final == tipos.Desconocido {
		t.estado.Desconocidos = append(t.estado.Desconocidos, "tarea")
	}
	t.trazarTipo(r, final, metodo, time.Since(t0))
	return final
}

func tiposATexto(ts []tipos.TipoRespuesta) []string {
	out := make([]string, 0, len(ts))
	for _, x := range ts {
		out = append(out, string(x))
	}
	return out
}

// decidir consulta al motor de decisión con la propuesta de las reglas. Con el presupuesto agotado, un error, un
// timeout o una elección fuera de las opciones, decide Reglas (la propuesta) y queda anotado.
func (t *turno) decidir(nombre string, opciones []string, propuesta string, confProp float64) tipos.Decision {
	if !contiene(opciones, propuesta) && len(opciones) > 0 {
		propuesta = opciones[0]
	}
	clave := PrefijoRegla + nombre
	t.estado.Inferencias[clave] = tipos.Dato{Valor: propuesta, Origen: origenReglas, Confianza: confProp}
	defer delete(t.estado.Inferencias, clave)
	ps := []tipos.PreguntaDecision{{Nombre: nombre, Opciones: opciones}}
	porReglas := func() tipos.Decision {
		ds, _ := Reglas{}.Decidir(t.ctx, t.estado, ps)
		return ds[0]
	}
	motor := t.a.motor()
	entrada := map[string]any{"decision": nombre, "opciones": len(opciones), "propuesta": propuesta}
	if err := t.pres.Decision(); err != nil {
		d := porReglas()
		t.anotarDecision(entrada, d, 0, "límite: "+err.Error())
		return d
	}
	if _, esReglas := motor.(Reglas); esReglas {
		d := porReglas()
		t.anotarDecision(entrada, d, 0, "")
		return d
	}
	ms := t.a.Config.Limites.TimeoutDecisionMs
	if ms <= 0 {
		ms = LimitesDefecto().TimeoutDecisionMs
	}
	ctx, cancel := context.WithTimeout(t.ctx, time.Duration(ms)*time.Millisecond)
	defer cancel()
	t0 := time.Now()
	ds, err := motor.Decidir(ctx, t.estado, ps)
	dur := time.Since(t0)
	if err == nil && (len(ds) != 1 || !contiene(opciones, ds[0].Eleccion)) {
		elegida := ""
		if len(ds) > 0 {
			elegida = ds[0].Eleccion
		}
		err = fmt.Errorf("%s: %q: %w", nombre, elegida, ErrFueraDeOpciones)
	}
	if err != nil {
		d := porReglas()
		t.anotarDecision(entrada, d, dur, motor.Nombre()+": "+traza.ResumirError(err))
		return d
	}
	d := ds[0]
	if d.Fuente == "" {
		d.Fuente = motor.Nombre()
	}
	t.estado.Inferencias[nombre] = tipos.Dato{Valor: d.Eleccion, Origen: origenDecision, Confianza: d.Confianza}
	t.anotarDecision(entrada, d, dur, "")
	return d
}

// ---------------------------------------------------------------------------------------------
// Búsqueda con expansión

type claseK struct {
	clase string
	k     int
}

// PlanConsulta: qué índices consultar en cada iteración (clases de internal/v2/conocimiento: procedimiento,
// concepto, error, fragmento). La primera va al índice del tipo; la expansión añade fragmentos con pasos (§9: si
// ningún procedimiento cubre la tarea, la V2 cae a los fragmentos). Como mucho 1 + 3 = 4 herramientas por turno.
func PlanConsulta(tipo tipos.TipoRespuesta, iteracion int) []claseK {
	switch tipo {
	case tipos.Concepto, tipos.Comparacion:
		if iteracion <= 1 {
			return []claseK{{"concepto", 5}}
		}
		return []claseK{{"concepto", 5}, {"fragmento", 8}}
	case tipos.Problema:
		if iteracion <= 1 {
			return []claseK{{"error", 5}}
		}
		return []claseK{{"error", 5}, {"procedimiento", 5}, {"fragmento", 8}}
	}
	if iteracion <= 1 {
		return []claseK{{"procedimiento", 5}}
	}
	return []claseK{{"procedimiento", 5}, {"fragmento", 8}}
}

// parada: conectores y palabras de pregunta que diluyen los términos del ERP en la consulta compacta.
var parada = conjuntoFrases(`a|al|algo|como|cmo|con|cual|cuales|cuando|de|del|desde|donde|dnd|el|ella|en|es|esa|ese|
eso|esta|este|hay|la|las|lo|los|me|mi|o|para|pero|por|que|q|se|si|sin|su|sus|un|una|unos|unas|y|yo|puedo|hago|
hacer|quiero|necesito|tengo|debo|porfa|favor|s10|sistema`)

// Compactar quita conectores y palabras de pregunta, y conserva todo lo demás tal cual (los términos exactos no
// se pierden: no son palabras de parada).
func Compactar(norm string) string {
	var out []string
	for _, w := range strings.Fields(norm) {
		if !parada[w] {
			out = append(out, w)
		}
	}
	if len(out) == 0 {
		return norm
	}
	return strings.Join(out, " ")
}

// Expandir arma la consulta de la siguiente iteración: original intacta, normalizada compacta y, como alias, los
// sinónimos de cada entidad.
func Expandir(c tipos.Consulta, al Aliaser) tipos.Consulta {
	if al == nil {
		al = SinAlias{}
	}
	nc := c
	nc.Entidades = append([]string(nil), c.Entidades...)
	alias := append([]string(nil), c.Aliases...)
	for _, ent := range c.Entidades {
		_, a, _ := al.Analizar(ent)
		alias = append(alias, a...)
	}
	nc.Aliases = unicosOrden(alias)
	nc.Normalizada = Compactar(c.Normalizada)
	return nc
}

// buscar recupera con expansión controlada. ok=false → SIN_EVIDENCIA con su motivo.
func (t *turno) buscar(tipo tipos.TipoRespuesta) ([]tipos.Candidato, string, bool) {
	if t.a.Recuperador == nil {
		t.trazarSinBusqueda(traza.EstadoOmitida, "sin recuperador cargado: no hay índice de procedimientos ni conceptos")
		return nil, "sin recuperador cargado", false
	}
	t.v.Inicio(traza.EtapaV2PlanConsulta)
	consulta := t.estado.Consulta
	var clasesUsadas [][]string
	motivo := ""
	for it := 1; ; it++ {
		if err := t.pres.Busqueda(); err != nil {
			motivo = "evidencia insuficiente tras " + itoa(it-1) + " búsqueda(s) (" + err.Error() + ")"
			break
		}
		if err := t.pres.Paso(); err != nil {
			motivo = err.Error()
			break
		}
		clases := PlanConsulta(tipo, it)
		cands, err := t.buscarIteracion(consulta, clases, it)
		clasesUsadas = append(clasesUsadas, nombresClases(clases))
		if err != nil {
			motivo = err.Error()
			break
		}
		if t.suficiente(cands, clases) {
			t.trazarPlanConsulta(consulta, clasesUsadas, it, len(cands), "")
			return cands, "", true
		}
		consulta = Expandir(consulta, t.a.aliaser())
	}
	t.trazarPlanConsulta(consulta, clasesUsadas, len(clasesUsadas), 0, motivo)
	return nil, motivo, false
}

func nombresClases(cs []claseK) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.clase
	}
	return out
}

// buscarIteracion: una llamada al recuperador por índice (cada una es una herramienta). Un índice que falla se
// anota y se sigue con los demás; el límite de herramientas corta la iteración (error).
func (t *turno) buscarIteracion(c tipos.Consulta, clases []claseK, it int) ([]tipos.Candidato, error) {
	antes := map[string]bool{}
	for _, id := range idsBusqueda {
		antes[id] = t.v.Pendiente(id)
	}
	var cands []tipos.Candidato
	var degradado, errores []string
	t0 := time.Now()
	var errLimite error
	for _, cl := range clases {
		if err := t.pres.Herramienta(); err != nil {
			errLimite = err
			break
		}
		res, deg, err := t.a.Recuperador.Buscar(t.ctx, c, cl.clase, cl.k)
		degradado = append(degradado, deg...)
		if err != nil {
			errores = append(errores, cl.clase+": "+traza.ResumirError(err))
			continue
		}
		cands = append(cands, res...)
	}
	dur := time.Since(t0)
	t.msBusqueda += dur
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].Puntaje > cands[j].Puntaje })
	for _, d := range degradado {
		t.degradado = appendUnico(t.degradado, d)
	}
	for _, e := range errores {
		t.degradado = appendUnico(t.degradado, "recuperador "+e)
	}
	t.trazarBusquedaIteracion(antes, cands, degradado, errores, it, dur)
	if errLimite != nil && len(cands) == 0 {
		return nil, errLimite
	}
	return cands, nil
}

func appendUnico(xs []string, x string) []string {
	for _, y := range xs {
		if y == x {
			return xs
		}
	}
	return append(xs, x)
}

// suficiente: hay al menos un candidato de los índices pedidos y el mejor alcanza V2_EVIDENCIA_MIN. Si hay
// candidatos pero bajo el mínimo (zona gris), decide el motor de decisión (evidence_sufficiency).
func (t *turno) suficiente(cands []tipos.Candidato, clases []claseK) bool {
	pedidas := map[string]bool{}
	for _, c := range clases {
		pedidas[c.clase] = true
	}
	var mejor *tipos.Candidato
	n := 0
	for i := range cands {
		if pedidas[cands[i].Clase] {
			n++
			if mejor == nil {
				mejor = &cands[i]
			}
		}
	}
	if mejor == nil {
		t.estado.Hechos["evidencia"] = tipos.Dato{Valor: "sin candidatos", Origen: origenBusqueda, Confianza: 1}
		return false
	}
	t.estado.Hechos["evidencia"] = tipos.Dato{
		Valor:  fmt.Sprintf("%d candidato(s); mejor %s «%s» (puntaje %.3f)", n, mejor.Clase, mejor.ID, mejor.Puntaje),
		Origen: origenBusqueda, Confianza: 1,
	}
	if mejor.Puntaje >= t.a.Config.EvidenciaMin {
		return true
	}
	d := t.decidir(DecSuficiencia, []string{Suficiente, Insuficiente}, Insuficiente, 0.5)
	return d.Eleccion == Suficiente
}

// margenProcedimiento: diferencia relativa de puntaje bajo la cual dos procedimientos «empatan» y decide el motor
// de decisión. PROVISIONAL, sin calibrar.
const margenProcedimiento = 0.05

// elegirProcedimiento: si los dos mejores procedimientos casi empatan, decide el motor (procedure) y el elegido
// pasa al frente. El resto del orden no cambia.
func (t *turno) elegirProcedimiento(cands []tipos.Candidato) []tipos.Candidato {
	var idx []int
	for i, c := range cands {
		if c.Clase == "procedimiento" {
			idx = append(idx, i)
			if len(t.altProcedimiento) < 3 {
				t.altProcedimiento = append(t.altProcedimiento, c.ID)
			}
		}
	}
	t.decididoProc = "puntaje"
	if len(idx) < 2 {
		return cands
	}
	p0, p1 := cands[idx[0]], cands[idx[1]]
	if p0.Puntaje <= 0 || (p0.Puntaje-p1.Puntaje)/p0.Puntaje >= margenProcedimiento {
		return cands
	}
	d := t.decidir(DecProcedimiento, []string{p0.ID, p1.ID}, p0.ID, 0.5)
	t.decididoProc = "decision:" + d.Fuente
	if d.Eleccion != p1.ID {
		return cands
	}
	out := append([]tipos.Candidato{p1}, cands[:idx[1]]...)
	return append(out, cands[idx[1]+1:]...)
}

// construir: evidencia y plan (paquete conocimiento). ok=false → SIN_EVIDENCIA con su motivo.
func (t *turno) construir(cands []tipos.Candidato) (tipos.Plan, string, bool) {
	if t.a.Constructor == nil {
		return tipos.Plan{}, "sin constructor de plan cargado", false
	}
	if err := t.pres.Paso(); err != nil {
		return tipos.Plan{}, err.Error(), false
	}
	t.v.Inicio(traza.EtapaV2Plan)
	plan, err := t.a.Constructor.Construir(t.ctx, t.estado, cands)
	if err != nil {
		return tipos.Plan{}, "constructor: " + traza.ResumirError(err), false
	}
	return plan, "", true
}

// ---------------------------------------------------------------------------------------------
// Redacción, gate y respuesta

// normalizarPlan: lo que el código garantiza de cualquier plan.
func (t *turno) normalizarPlan(p tipos.Plan) tipos.Plan {
	p.Version = 2
	if p.Tipo == "" {
		p.Tipo = t.estado.Tipo
	}
	if p.Fuentes == nil {
		p.Fuentes = []tipos.FuentePlan{}
	}
	return p
}

// sinAfirmaciones: planes que no afirman nada del ERP (aclaración, SIN_EVIDENCIA): no pasan por el gate.
func sinAfirmaciones(p tipos.Plan) bool {
	return p.SinEvidencia || (len(p.Pasos) == 0 && p.Concepto == nil && len(p.Errores) == 0 &&
		len(p.Verificacion) == 0 && len(p.Prerrequisitos) == 0 && len(p.Afirmaciones) == 0)
}

// entregar redacta y valida el plan, y arma la respuesta. base es la memoria antes de aplicar el plan (con el
// procedimiento suspendido, si lo hay).
func (t *turno) entregar(plan tipos.Plan, base tipos.Memoria, motivo string) rag.Respuesta {
	plan = t.normalizarPlan(plan)
	mem := t.memoriaSalida(plan, base)
	if mem.ProcedimientoID == "" && mem.Suspendido != nil && plan.Plantilla != "ACLARAR_TAREA" {
		plan.Siguiente = OfertaRetoma(t.tituloProcedimiento(mem.Suspendido.ProcedimientoID), mem.Suspendido.PasoActual)
		mem.Ofrecido = "" // la oferta que se ve es retomar: un «sí» retoma, no empieza lo ofrecido
	}
	t.trazarPlan(plan)
	if sinAfirmaciones(plan) {
		texto, err := t.a.plantillas().Redactar(t.ctx, plan)
		t.generador, t.modoRedaccion = t.a.plantillas().Nombre(), "plantillas"
		if err != nil || strings.TrimSpace(texto) == "" {
			texto, t.generador, t.modoRedaccion = RenderBasico(plan), "basico", "basico"
		}
		razon := "aclaración: no afirma nada del ERP"
		if plan.SinEvidencia {
			razon = "sin evidencia: no hay afirmaciones que verificar"
		}
		t.v.OmitirCon(traza.EtapaV2QualityGate, traza.EstadoOmitida, razon, nil)
		modo := "aclaracion"
		if plan.SinEvidencia {
			modo = "sin_contexto"
		}
		return t.respuesta(plan, texto, mem, modo, motivo)
	}
	gate, nombreGate := t.a.gate()
	texto, err := t.redactar(plan)
	if err != nil {
		return t.segura(plan, base, err.Error())
	}
	t0 := time.Now()
	cal := gate.Evaluar(plan, texto, t.estado)
	msGate := time.Since(t0)
	t.gateIntentos = 1
	primeros := cal
	if !cal.Paso {
		if err := t.regenerarPermitido(); err != nil {
			t.trazarGate(nombreGate, cal, primeros, msGate)
			return t.segura(plan, base, "quality gate: "+strings.Join(cal.Problemas, "; ")+" ("+err.Error()+")")
		}
		texto2, gen := t.redactarSinLLM(plan)
		t1 := time.Now()
		cal = gate.Evaluar(plan, texto2, t.estado)
		msGate += time.Since(t1)
		t.gateIntentos = 2
		if cal.Paso {
			texto, t.generador, t.modoRedaccion = texto2, gen, "regenerado_sin_llm"
		}
	}
	t.trazarGate(nombreGate, cal, primeros, msGate)
	if !cal.Paso {
		return t.segura(plan, base, "quality gate: "+strings.Join(cal.Problemas, "; "))
	}
	return t.respuesta(plan, texto, mem, "respuesta", motivo)
}

func (t *turno) regenerarPermitido() error {
	if err := t.pres.Regeneracion(); err != nil {
		return err
	}
	return t.pres.Paso()
}

// redactar: el generador (LLM local que reformula) con GENERATION_TIMEOUT_MS; si cae o se agotó el presupuesto de
// generaciones, plantillas; si fallan las plantillas, el renderizador básico.
func (t *turno) redactar(plan tipos.Plan) (string, error) {
	if err := t.pres.Paso(); err != nil {
		return "", err
	}
	if g := t.a.Generador; g != nil {
		if err := t.pres.Generacion(); err != nil {
			t.degradado = appendUnico(t.degradado, "sin LLM → plantillas ("+err.Error()+")")
		} else {
			ms := t.a.Config.Limites.TimeoutGeneracionMs
			if ms <= 0 {
				ms = LimitesDefecto().TimeoutGeneracionMs
			}
			ctx, cancel := context.WithTimeout(t.ctx, time.Duration(ms)*time.Millisecond)
			t0 := time.Now()
			texto, err := g.Redactar(ctx, plan)
			cancel()
			t.msLLM += time.Since(t0)
			if err == nil && strings.TrimSpace(texto) != "" {
				t.generador, t.modoRedaccion = g.Nombre(), "llm"
				return texto, nil
			}
			causa := "texto vacío"
			if err != nil {
				causa = traza.ResumirError(err)
			}
			t.degradado = appendUnico(t.degradado, "sin LLM → plantillas ("+causa+")")
		}
	}
	texto, gen := t.redactarSinLLM(plan)
	t.generador = gen
	if t.modoRedaccion == "" {
		t.modoRedaccion = "plantillas"
	}
	if gen == "basico" {
		t.modoRedaccion = "basico"
	}
	return texto, nil
}

func (t *turno) redactarSinLLM(plan tipos.Plan) (string, string) {
	pl := t.a.plantillas()
	texto, err := pl.Redactar(t.ctx, plan)
	if err != nil || strings.TrimSpace(texto) == "" {
		if err != nil {
			t.degradado = appendUnico(t.degradado, "plantillas → básico ("+traza.ResumirError(err)+")")
		}
		return RenderBasico(plan), "basico"
	}
	return texto, pl.Nombre()
}

// segura: la respuesta segura (§6): lo respaldado del plan, redactado sin modelo, o SIN_EVIDENCIA. No vuelve a
// pasar por el gate (no hay bucle).
func (t *turno) segura(plan tipos.Plan, base tipos.Memoria, motivo string) rag.Respuesta {
	s := t.normalizarPlan(PlanSeguro(plan))
	mem := t.memoriaSalida(s, base)
	t.trazarPlan(s)
	t.v.Dato(traza.EtapaV2Plan, "respuesta_segura", true)
	t.generador, t.modoRedaccion = "basico", "segura"
	modo := "respuesta_segura"
	if s.SinEvidencia {
		modo = "sin_contexto"
	}
	return t.respuesta(s, RenderBasico(s), mem, modo, "respuesta segura: "+motivo)
}

// memoriaSalida: la memoria que deja el plan. Con procedimiento a la vista, su paso actual (conservando lo
// completado, si es el mismo, y el suspendido); sin él, solo el suspendido.
func (t *turno) memoriaSalida(plan tipos.Plan, base tipos.Memoria) tipos.Memoria {
	m := MemoriaDePlan(plan, base)
	if m.ProcedimientoID != "" && m.ProcedimientoID == base.ProcedimientoID {
		m.Retomado = base.Retomado
	}
	if base.Suspendido != nil && (m.ProcedimientoID == "" || m.ProcedimientoID != base.Suspendido.ProcedimientoID) {
		s := *base.Suspendido
		m.Suspendido = &s
	}
	return m
}

func (t *turno) tituloProcedimiento(id string) string {
	if def, ok := t.procedimiento(id); ok && def.Titulo != "" {
		return def.Titulo
	}
	return id
}

// social: la charla de V1 (sin búsqueda). La memoria no cambia.
func (t *turno) social(mem tipos.Memoria) rag.Respuesta {
	t.trazarSinBusqueda(traza.EstadoNoTomada, "social: charla directa de V1, sin búsqueda")
	intencion := t.tipoRes.IntencionV1
	if intencion != "limite" {
		intencion = "social"
	}
	var res rag.Respuesta
	var err error
	if t.a.Charla != nil {
		res, err = t.a.Charla.Conversar(t.ctx, intencion, t.pregunta, t.o.Hilo)
	}
	if t.a.Charla == nil || err != nil {
		res = rag.Respuesta{Pregunta: t.pregunta, Respuesta: "Hola, ¿en qué te ayudo con S10?", Modo: "conversacional"}
		if err != nil {
			res.Motivo = "charla no disponible: " + traza.ResumirError(err)
		}
	}
	plan := tipos.Plan{Version: 2, Tipo: tipos.Social, Fuentes: []tipos.FuentePlan{}}
	t.trazarPlan(plan)
	t.v.OmitirCon(traza.EtapaV2QualityGate, traza.EstadoOmitida, "social: no afirma nada del ERP", nil)
	t.generador, t.modoRedaccion = "charla_v1", "charla"
	res.Plan = rag.Orquestacion{Intencion: intencion, TipoConsulta: "social", Ruta: "v2", Clasificador: t.metodo,
		SimilitudClasificador: redondear(t.estado.TipoConfianza)}
	if res.Fuentes == nil {
		res.Fuentes = []rag.Fuente{}
	}
	m := mem
	res.PlanV2, res.Memoria, res.Version = &plan, &m, tipos.V2
	return res
}

// tipoConsultaV1: el tipo V2 en el vocabulario de «orquestacion.tipo_consulta» de V1.
var tipoConsultaV1 = map[tipos.TipoRespuesta]string{
	tipos.Concepto: "concepto", tipos.Procedimiento: "procedimiento", tipos.Navegacion: "navegacion",
	tipos.Problema: "problema", tipos.Configuracion: "configuracion", tipos.Comparacion: "comparacion",
	tipos.Desconocido: "aclaracion", tipos.Social: "social",
}

// respuesta arma /ask con los campos de V1 más plan, memoria y version.
func (t *turno) respuesta(plan tipos.Plan, texto string, mem tipos.Memoria, modo, motivo string) rag.Respuesta {
	if len(t.degradado) > 0 {
		d := "degradado: " + strings.Join(t.degradado, "; ")
		if motivo == "" {
			motivo = d
		} else {
			motivo += " · " + d
		}
	}
	metodo := t.metodo
	if metodo == "" {
		metodo = "memoria"
	}
	res := rag.Respuesta{
		Pregunta:  t.pregunta,
		Respuesta: texto,
		Modo:      modo,
		Plan: rag.Orquestacion{Intencion: "trabajo", TipoConsulta: tipoConsultaV1[plan.Tipo], Ruta: "v2",
			Clasificador: metodo, SimilitudClasificador: redondear(t.estado.TipoConfianza)},
		Fuentes:     FuentesRespuesta(plan),
		SinContexto: plan.SinEvidencia,
		Motivo:      motivo,
		TiempoBusq:  t.msBusqueda,
		TiempoLLM:   t.msLLM,
		MsBusqueda:  t.msBusqueda.Milliseconds(),
		MsLLM:       t.msLLM.Milliseconds(),
		Version:     tipos.V2,
	}
	p, m := plan, mem
	res.PlanV2, res.Memoria = &p, &m
	return res
}

// FuentesRespuesta pasa las fuentes del plan al formato de /ask de V1 («fuentes»): una por manual y sección.
func FuentesRespuesta(p tipos.Plan) []rag.Fuente {
	out := make([]rag.Fuente, 0, len(p.Fuentes))
	for _, f := range p.Fuentes {
		cita := f.Manual
		if f.Seccion != "" {
			cita += " › " + f.Seccion
		}
		pagina := 0
		if len(f.Paginas) > 0 {
			pagina = f.Paginas[0]
			ps := make([]string, len(f.Paginas))
			for i, x := range f.Paginas {
				ps[i] = itoa(x)
			}
			cita += " · p. " + strings.Join(ps, ", ")
		}
		out = append(out, rag.Fuente{Cita: cita, Documento: f.Manual, Pagina: pagina})
	}
	return out
}

func redondear(x float64) float64 { return traza.Redondear(x) }
