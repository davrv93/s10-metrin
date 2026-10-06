package v2

// Continuaciones con referencia: «¿y cómo la hago?», «¿y eso dónde está?», «ya, y cómo registro uno?». La pregunta
// no nombra su tema; apunta a lo último que se habló. Se resuelve ANTES de buscar, sin modelo, con la memoria y el
// hilo. (Lo que la memoria resuelve sin pregunta nueva —«¿y luego?», «listo», «sí, enséñame», «no me sale»— va por
// InterpretarMemoria, en memoria.go.)
//
// Reglas, todas deterministas:
//  1. Elíptica: como mucho maxPalabrasElipsis palabras, no nombra ningún término del ERP (Aliaser ni términos
//     exactos) y lleva un marcador de referencia: un demostrativo («eso», «ahí»), un pronombre átono pegado a un
//     verbo («la hago», «lo imprimo», «hacerlo»), «uno»/«una» al final, o empieza por «y …» / «ya, y …» sin
//     contenido propio.
//  2. Lo que pide sale de la pregunta: «dónde» o «en qué pantalla» → NAVIGATION; «qué es», «qué significa», «para
//     qué sirve» → CONCEPT; «cómo», «pasos» o un verbo de acción → PROCEDURE. Solo manda si el clasificador no sabe
//     (UNKNOWN o SOCIAL): una pregunta elíptica casi no le da con qué.
//  3. El referente: de la memoria, el procedimiento OFRECIDO y el CONCEPTO explicado en la respuesta anterior (ids:
//     hecho); si no, el procedimiento en curso o el suspendido (su título, como texto); si la memoria no dice nada,
//     el último turno del usuario en el hilo que tenga contenido (inferencia, confianza 0,6).
//  4. Sin palabras propias y con un id de la memoria, se va directo a ese procedimiento o concepto, sin buscar. Con
//     palabras propias («¿y cómo lo imprimo para la SUNAT?») o con un referente que es solo texto, la consulta
//     normalizada suma los términos del referente (la original no se toca) y se busca como siempre.
//  5. Elíptica, sin nada propio y sin referente: se pide aclaración (como «¿y luego?» sin procedimiento en curso).

import (
	"strings"

	"rag-go/internal/rag"
	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

const (
	maxPalabrasElipsis = 8
	confianzaHilo      = 0.6
	confianzaMemHilo   = 0.9 // memoria reconstruida del último plan del hilo
	maxTurnosReferente = 2   // turnos de usuario hacia atrás en los que se busca el referente
)

// Referencia: a qué apunta una pregunta elíptica y qué pide.
type Referencia struct {
	Pide          tipos.TipoRespuesta // por la palabra interrogativa; "" = no lo dice
	Procedimiento string              // id del procedimiento ofrecido (memoria)
	Concepto      string              // id del concepto explicado (memoria)
	Texto         string              // términos del referente que se suman a la consulta normalizada
	Origen        string              // memoria | hilo | "" (sin referente)
	Confianza     float64
	Propias       []string // palabras de contenido de la propia pregunta
}

// Hay: se encontró un referente.
func (r Referencia) Hay() bool { return r.Origen != "" }

// Directa: sin palabras propias y con un id de la memoria (la que mandó el cliente): se va a él sin buscar.
func (r Referencia) Directa() bool {
	return len(r.Propias) == 0 && r.Origen == origenMemoria && (r.Procedimiento != "" || r.Concepto != "")
}

// Describir: el referente en una línea (traza y estado).
func (r Referencia) Describir() string {
	var partes []string
	if r.Procedimiento != "" {
		partes = append(partes, "procedimiento "+r.Procedimiento)
	}
	if r.Concepto != "" {
		partes = append(partes, "concepto "+r.Concepto)
	}
	if len(partes) == 0 && r.Texto != "" {
		partes = append(partes, "«"+r.Texto+"»")
	}
	if len(partes) == 0 {
		return "sin referente"
	}
	return strings.Join(partes, " + ")
}

var (
	demostrativos = conjuntoFrases(`eso|esto|ese|esa|este|esta|esos|esas|estos|estas|aquello|aquel|aquella|ahi|alli|alla|aqui`)
	indefinidos   = conjuntoFrases(`uno|una|unos|unas`)
	atonos        = conjuntoFrases(`lo|la|los|las|le|les`)
	encliticos    = []string{"selo", "sela", "selos", "selas", "los", "las", "les", "lo", "la", "le"}

	// relleno: palabras de pregunta, conectores y muletillas (además de «parada», en agente.go).
	relleno = conjuntoFrases(`y|ya|ok|okey|oki|bueno|listo|entonces|pues|oye|porfa|favor|ahora|luego|despues|tambien|
igual|asi|bien|mas|cual|cuales|cuando|quien|como|cmo|donde|dnd|que|q|se|te|nos|mis|tu|tus|no|si|otra|otro|vez|
ademas|aparte|eh|ah|mmm|gracias|primero|alguno|alguna|mismo|misma|parte|aca`)

	// verbosGenericos: acciones que no dicen sobre QUÉ («la hago», «lo registro»). Lo que no está aquí cuenta como
	// contenido propio: ante la duda, se busca (no se va directo).
	verbosGenericos = conjuntoFrases(`hago|hacer|hace|hacemos|haga|hice|hizo|registro|registrar|registra|registramos|
creo|crear|crea|creamos|configuro|configurar|configura|imprimo|imprimir|imprime|veo|ver|vea|ves|uso|usar|usa|usan|
saco|sacar|saca|encuentro|encontrar|encuentra|pongo|poner|pone|cambio|cambiar|cambia|busco|buscar|busca|genero|
generar|genera|lleno|llenar|agrego|agregar|anado|anadir|añado|añadir|borro|borrar|elimino|eliminar|modifico|modificar|edito|editar|
abro|abrir|entro|entrar|ingreso|ingresar|cargo|cargar|proceso|procesar|guardo|guardar|grabo|grabar|subo|subir|mando|
mandar|envio|enviar|activo|activar|asigno|asignar|consulto|consultar|reviso|revisar|ubico|ubicar|dejo|dejar|
queda|quedo|doy|dar|da|armo|armar|completo|completar|empiezo|empezar|comienzo|comenzar|sigo|seguir|continuo|
continuar|tengo|tener|puedo|poder|debo|deber|necesito|quiero|hay|es|son|esta|estan|estar|ser|viene|vienen|significa|
refiere|sirve|sirven|llama|llaman|aparece|sale|trata|funciona|pasos|paso`)

	// verbosNoAccion: genéricos que no piden hacer nada («es», «hay», «sirve»…).
	verbosNoAccion = conjuntoFrases(`es|son|esta|estan|estar|ser|hay|viene|vienen|significa|refiere|sirve|sirven|llama|
llaman|aparece|sale|trata|funciona|tengo|tener|puedo|poder|debo|deber|necesito|quiero|queda|quedo|paso`)
)

// sinEnclitico: «hacerlo» → «hacer», «registrarla» → «registrar» (solo si queda un infinitivo o un gerundio).
func sinEnclitico(w string) (string, bool) {
	for _, e := range encliticos {
		if base, ok := strings.CutSuffix(w, e); ok && len(base) >= 3 {
			if strings.HasSuffix(base, "ar") || strings.HasSuffix(base, "er") || strings.HasSuffix(base, "ir") ||
				strings.HasSuffix(base, "ando") || strings.HasSuffix(base, "endo") {
				return base, true
			}
		}
	}
	return w, false
}

func esVerboGenerico(w string) bool {
	if verbosGenericos[w] {
		return true
	}
	base, ok := sinEnclitico(w)
	return ok && verbosGenericos[base]
}

// palabrasPropias: el contenido de la pregunta (lo que no es relleno, referencia ni acción genérica).
func palabrasPropias(norm string) []string {
	var out []string
	for _, w := range strings.Fields(norm) {
		if len(w) < 2 || parada[w] || relleno[w] || demostrativos[w] || indefinidos[w] || atonos[w] || esVerboGenerico(w) {
			continue
		}
		out = append(out, w)
	}
	return out
}

// marcadorReferencia: la pregunta apunta a algo dicho antes.
func marcadorReferencia(ws []string) bool {
	for i, w := range ws {
		switch {
		case demostrativos[w]:
			return true
		case indefinidos[w] && i == len(ws)-1 && i > 0:
			return true // «registro uno»
		case atonos[w] && i+1 < len(ws) && (esVerboGenerico(ws[i+1]) || (len(ws[i+1]) >= 3 && strings.HasSuffix(ws[i+1], "o"))):
			return true // «la hago», «lo imprimo», «se lo asigno»
		}
		if _, ok := sinEnclitico(w); ok {
			return true // «hacerlo», «configurarla»
		}
	}
	return false
}

// EsElipsis: la pregunta es corta, no nombra nada del ERP y apunta a algo anterior (regla 1).
func EsElipsis(pregunta string, al Aliaser) bool {
	n := Normalizar(pregunta)
	ws := strings.Fields(n)
	if len(ws) < 2 || len(ws) > maxPalabrasElipsis {
		return false
	}
	if al == nil {
		al = SinAlias{}
	}
	if ents, _, _ := al.Analizar(pregunta); len(ents) > 0 || len(TerminosExactos(pregunta)) > 0 {
		return false
	}
	if marcadorReferencia(ws) {
		return true
	}
	empieza := ws[0] == "y" || ((ws[0] == "ya" || ws[0] == "entonces" || ws[0] == "ok" || ws[0] == "bueno") && ws[1] == "y")
	return empieza && len(palabrasPropias(n)) == 0
}

// PidePregunta: lo que pide la palabra interrogativa (regla 2); "" si no lo dice.
func PidePregunta(pregunta string) tipos.TipoRespuesta {
	t := " " + Normalizar(pregunta) + " "
	for _, f := range []string{" donde ", " dnd ", " en que pantalla ", " en que escenario ", " en que menu ", " en que ventana ",
		" en que parte ", " en que modulo ", " en que opcion "} {
		if strings.Contains(t, f) {
			return tipos.Navegacion
		}
	}
	for _, f := range []string{" que es ", " que son ", " que significa ", " que viene a ser ", " a que se refiere ",
		" para que sirve ", " de que se trata ", " que quiere decir "} {
		if strings.Contains(t, f) {
			return tipos.Concepto
		}
	}
	if strings.Contains(t, " como ") || strings.Contains(t, " cmo ") || strings.Contains(t, " pasos ") {
		return tipos.Procedimiento
	}
	for _, w := range strings.Fields(t) {
		if esVerboGenerico(w) && !verbosNoAccion[w] {
			if base, ok := sinEnclitico(w); !ok || !verbosNoAccion[base] {
				return tipos.Procedimiento
			}
		}
	}
	return ""
}

// ResolverReferencia aplica las reglas 1–3. ok = la pregunta es elíptica (aunque no haya referente: ver Hay).
// origenMem dice si la memoria la mandó el cliente (hecho) o se reconstruyó del hilo (inferencia).
func ResolverReferencia(pregunta string, mem tipos.Memoria, origenMem string, hilo []rag.Turno, al Aliaser, cat Catalogo) (Referencia, bool) {
	if !EsElipsis(pregunta, al) {
		return Referencia{}, false
	}
	r := Referencia{Pide: PidePregunta(pregunta), Propias: palabrasPropias(Normalizar(pregunta))}
	titulo := func(id string) string {
		if cat != nil {
			if p, ok := cat.Procedimiento(id); ok && p.Titulo != "" {
				return p.Titulo
			}
		}
		return ""
	}
	var textos []string
	if mem.Concepto != "" {
		r.Concepto = mem.Concepto
		if cat != nil {
			if c, ok := cat.Concepto(mem.Concepto); ok && c.Termino != "" {
				textos = append(textos, c.Termino)
			}
		}
	}
	if mem.Ofrecido != "" {
		r.Procedimiento = mem.Ofrecido
		textos = append(textos, titulo(mem.Ofrecido))
	}
	if r.Procedimiento == "" && r.Concepto == "" {
		switch {
		case mem.ProcedimientoID != "":
			textos = append(textos, titulo(mem.ProcedimientoID))
		case mem.Suspendido != nil && mem.Suspendido.ProcedimientoID != "":
			textos = append(textos, titulo(mem.Suspendido.ProcedimientoID))
		}
	}
	r.Texto = strings.Join(strings.Fields(strings.Join(textos, " ")), " ")
	if r.Texto != "" || r.Procedimiento != "" || r.Concepto != "" {
		r.Origen, r.Confianza = origenMemoria, 1
		if origenMem == origenHilo {
			r.Origen, r.Confianza = origenHilo, confianzaMemHilo
		}
		return r, true
	}
	// Último recurso: el último turno del usuario con contenido.
	vistos := 0
	for i := len(hilo) - 1; i >= 0 && vistos < maxTurnosReferente; i-- {
		if hilo[i].Rol != "usuario" || strings.TrimSpace(hilo[i].Texto) == "" {
			continue
		}
		vistos++
		if p := palabrasPropias(Normalizar(hilo[i].Texto)); len(p) > 0 {
			r.Texto, r.Origen, r.Confianza = strings.Join(p, " "), origenHilo, confianzaHilo
			return r, true
		}
	}
	return r, true
}

// ConsultaConReferencia: la consulta del turno con los términos del referente en la normalizada (la original no se
// toca). Así los cuenta el recuperador como términos del usuario.
func ConsultaConReferencia(c tipos.Consulta, r Referencia) tipos.Consulta {
	if r.Texto == "" {
		return c
	}
	nc := c
	nc.Entidades = append([]string(nil), c.Entidades...)
	nc.Aliases = append([]string(nil), c.Aliases...)
	nc.Normalizada = strings.TrimSpace(c.Normalizada + " " + Normalizar(r.Texto))
	return nc
}

// --- aceptar una oferta («sí, enséñame») -------------------------------------------------------------------

// restoAceptacion: lo que puede seguir a un «sí» al aceptar una oferta («sí, enséñame», «claro, muéstrame los pasos»).
var (
	inicioAceptacion = conjuntoFrases(`si|claro|dale|ok|okey|va|bueno|perfecto|ya|porfa|ensename|ensenamelo|enseñame|enseñamelo|
muestrame|muestramelo|explicame|quiero`)
	restoAceptacion = conjuntoFrases(`si|claro|dale|ok|va|bueno|perfecto|ya|por|favor|porfa|ensename|ensenamelo|ensenamelos|enseñame|
enseñamelo|enseñamelos|
muestrame|muestramelo|explicame|explicamelo|dime|como|se|hace|hago|los|las|el|la|pasos|paso|adelante|quiero|me|
gustaria|aprender|vamos|sigamos|empecemos|empezar|eso|esto|de|una|vez`)
)

// pideEnsenar: una sola palabra que acepta la oferta («enséñame»).
var pideEnsenar = conjuntoFrases(`ensename|ensenamelo|enseñame|enseñamelo|muestrame|muestramelo|explicame|porfa`)

// AceptaOferta: el mensaje acepta lo ofrecido («sí», «sí, enséñame», «dale, muéstrame los pasos», «¿y luego?»).
func AceptaOferta(n string) bool {
	if afirmaciones[n] || esAvanceExplicito(n) {
		return true
	}
	ws := strings.Fields(n)
	if len(ws) == 0 || len(ws) > 7 || !inicioAceptacion[ws[0]] {
		return false
	}
	for _, w := range ws[1:] {
		if !restoAceptacion[w] {
			return false
		}
	}
	return len(ws) > 1 || pideEnsenar[ws[0]]
}

// --- en el orquestador -------------------------------------------------------------------------------------

// TextoSinReferente: pregunta elíptica sin nada a qué referirse (regla 5).
const TextoSinReferente = "¿A qué se refiere? Dígame la tarea o el término de S10 y le guío con el manual."

// claveReferenciaResuelta: la referencia resuelta en el estado (inferencia).
const claveReferenciaResuelta = "referencia"

// aplicarReferencia: el tipo (regla 2, solo si el clasificador no sabe) y la consulta con el referente (regla 4).
func (t *turno) aplicarReferencia(ref Referencia, tipo tipos.TipoRespuesta) tipos.TipoRespuesta {
	t.ref = &ref
	if !ref.Hay() {
		return tipo
	}
	t.estado.Inferencias[claveReferenciaResuelta] = tipos.Dato{Valor: ref.Describir(), Origen: ref.Origen, Confianza: ref.Confianza}
	t.estado.Consulta = ConsultaConReferencia(t.estado.Consulta, ref)
	if (tipo != tipos.Desconocido && tipo != tipos.Social) || ref.Pide == "" {
		return tipo
	}
	anterior := tipo
	tipo = ref.Pide
	t.metodo = "referencia"
	t.estado.Tipo, t.estado.TipoConfianza = tipo, ref.Confianza
	t.estado.Inferencias["tipo"] = tipos.Dato{Valor: string(tipo), Origen: ref.Origen, Confianza: ref.Confianza}
	var desc []string
	for _, d := range t.estado.Desconocidos {
		if d != "tarea" {
			desc = append(desc, d)
		}
	}
	t.estado.Desconocidos = desc
	if t.v != nil {
		t.v.Dato(traza.EtapaV2TipoRespuesta, "tipo", string(tipo))
		t.v.Dato(traza.EtapaV2TipoRespuesta, "metodo", "referencia")
		t.v.Dato(traza.EtapaV2TipoRespuesta, "clasificador", string(anterior))
		t.v.Dato(traza.EtapaV2TipoRespuesta, "referencia", ref.Describir())
		t.v.Razon(traza.EtapaV2TipoRespuesta, "pregunta elíptica: «"+string(tipo)+"» por la pregunta; el clasificador dio «"+
			string(anterior)+"»; referente: "+ref.Describir())
		t.v.Modelo(traza.EtapaV2TipoRespuesta, "referencia (reglas)")
	}
	return tipo
}

// candidatosReferencia: los ids de la memoria como candidatos (regla 4), según lo que pide el tipo. Vacío = buscar.
func (t *turno) candidatosReferencia(tipo tipos.TipoRespuesta) []tipos.Candidato {
	r := t.ref
	mk := func(id, clase string) tipos.Candidato {
		return tipos.Candidato{ID: id, Clase: clase, Puntaje: r.Confianza, Lexico: 1,
			Meta: map[string]string{"cobertura": "1.000", "referencia": r.Origen}}
	}
	proc := func() []tipos.Candidato {
		if _, ok := t.procedimiento(r.Procedimiento); ok {
			return []tipos.Candidato{mk(r.Procedimiento, "procedimiento")}
		}
		return nil
	}
	conc := func() []tipos.Candidato {
		if t.a.Catalogo != nil && r.Concepto != "" {
			if _, ok := t.a.Catalogo.Concepto(r.Concepto); ok {
				return []tipos.Candidato{mk(r.Concepto, "concepto")}
			}
		}
		return nil
	}
	switch tipo {
	case tipos.Procedimiento, tipos.Configuracion:
		return proc()
	case tipos.Navegacion:
		return append(proc(), conc()...)
	case tipos.Concepto:
		return conc()
	}
	return nil
}

// trazarReferenciaDirecta: el plan de consulta es la memoria; no hubo búsqueda.
func (t *turno) trazarReferenciaDirecta(cands []tipos.Candidato) {
	t.altProcedimiento = nil
	if t.v == nil {
		return
	}
	c := t.estado.Consulta
	t.v.Fin(traza.EtapaV2PlanConsulta, traza.EstadoOK, traza.Datos{
		"clases": []string{}, "consulta": c.Normalizada, "original": c.Original, "iteraciones": 0, "expandida": false,
		"referencia": t.ref.Describir(), "origen_referencia": t.ref.Origen, "fuentes": len(cands),
		"confianza": traza.Redondear(t.ref.Confianza),
	})
	t.v.Razon(traza.EtapaV2PlanConsulta, "referencia resuelta con la memoria ("+t.ref.Describir()+"): sin búsqueda")
	for _, id := range idsBusqueda {
		if t.v.Pendiente(id) {
			t.v.OmitirCon(id, traza.EstadoNoTomada, "referencia resuelta con la memoria: sin búsqueda", nil)
		}
	}
}

// trazarReferenciaBusqueda: la búsqueda llevó los términos del referente (regla 4).
func (t *turno) trazarReferenciaBusqueda() {
	if t.v == nil || t.ref == nil || !t.ref.Hay() {
		return
	}
	t.v.Dato(traza.EtapaV2PlanConsulta, "referencia", t.ref.Describir())
	t.v.Dato(traza.EtapaV2PlanConsulta, "origen_referencia", t.ref.Origen)
	t.v.Dato(traza.EtapaV2PlanConsulta, "texto_referencia", t.ref.Texto)
}
