package v2

import (
	"context"
	"os"
	"strings"
	"testing"

	"rag-go/internal/rag"
	"rag-go/internal/traza"
	"rag-go/internal/v2/conocimiento"
	"rag-go/internal/v2/tipos"
)

// --- reglas (unitarias) ---------------------------------------------------------------------------------------

func TestEsElipsis(t *testing.T) {
	si := []string{"¿y cómo la hago?", "¿y eso dónde está?", "ya, y como registro uno?", "¿y cómo lo imprimo para la sunat?",
		"como hago eso", "¿y cómo se hace?", "¿y dónde está?", "¿hacerlo cómo?", "entonces y como se configura"}
	no := []string{
		"¿Cómo registro un metrado?",     // nombra un término del ERP (aliasFalso: Metrado)
		"como hago la formula",           // «la» es artículo: no hay referencia
		"lo del otro dia no me salio",    // «lo del»: no es un pronombre pegado a un verbo
		"¿y cómo configuro la tecla F7?", // término exacto
		"eso",                            // una sola palabra: nada que resolver
		"¿y como lo hago? necesito registrar la valorizacion mensual del subcontrato de la obra", // larga
	}
	for _, q := range si {
		if !EsElipsis(q, aliasFalso{}) {
			t.Errorf("%q debería ser elíptica", q)
		}
	}
	for _, q := range no {
		if EsElipsis(q, aliasFalso{}) {
			t.Errorf("%q no es elíptica", q)
		}
	}
}

func TestPidePregunta(t *testing.T) {
	casos := map[string]tipos.TipoRespuesta{
		"¿y eso dónde está?": tipos.Navegacion, "¿en qué pantalla lo veo?": tipos.Navegacion,
		"¿y eso qué es?": tipos.Concepto, "¿para qué sirve eso?": tipos.Concepto,
		"¿y cómo la hago?": tipos.Procedimiento, "ya, y registro uno": tipos.Procedimiento, "¿hacerlo?": tipos.Procedimiento,
		"¿y eso?": "",
	}
	for q, quiero := range casos {
		if got := PidePregunta(q); got != quiero {
			t.Errorf("%q: %q, quiero %q", q, got, quiero)
		}
	}
}

func TestResolverReferencia(t *testing.T) {
	cat := nuevoCatalogo()
	q := "¿y cómo la hago?"

	// Ofrecido y concepto de la memoria del cliente: ids, directa.
	r, ok := ResolverReferencia(q, tipos.Memoria{Concepto: "metrado", Ofrecido: "metrados.registrar"}, origenMemoria, nil, aliasFalso{}, cat)
	if !ok || r.Procedimiento != "metrados.registrar" || r.Concepto != "metrado" || !r.Directa() || r.Pide != tipos.Procedimiento ||
		!strings.Contains(r.Texto, "Metrado") || !strings.Contains(r.Texto, "Registrar un metrado") || r.Confianza != 1 {
		t.Fatalf("memoria: %+v %v", r, ok)
	}
	// La misma memoria reconstruida del hilo: es una inferencia, no se va directo (se busca con su texto).
	if r, _ := ResolverReferencia(q, tipos.Memoria{Ofrecido: "metrados.registrar"}, origenHilo, nil, aliasFalso{}, cat); r.Directa() || r.Origen != origenHilo || r.Texto == "" {
		t.Errorf("memoria del hilo: %+v", r)
	}
	// Procedimiento en curso: su título como texto, sin ir directo (no se reinicia el procedimiento).
	if r, _ := ResolverReferencia(q, tipos.Memoria{ProcedimientoID: "config.unidades", PasoActual: 1}, origenMemoria, nil, aliasFalso{}, cat); r.Directa() || r.Texto != "Configurar unidades" {
		t.Errorf("en curso: %+v", r)
	}
	// Con palabras propias no se va directo aunque haya ids.
	if r, _ := ResolverReferencia("¿y cómo lo imprimo para la sunat?", tipos.Memoria{Ofrecido: "metrados.registrar"}, origenMemoria, nil, aliasFalso{}, cat); r.Directa() || len(r.Propias) != 1 {
		t.Errorf("propias: %+v", r)
	}
	// Sin memoria: el último turno del usuario con contenido (se salta el que no tiene).
	hilo := []rag.Turno{{Rol: "usuario", Texto: "que es la formula polinomica"}, {Rol: "asistente", Texto: "Es …"},
		{Rol: "usuario", Texto: "gracias"}, {Rol: "asistente", Texto: "De nada"}}
	if r, ok := ResolverReferencia(q, tipos.Memoria{}, origenMemoria, hilo, aliasFalso{}, cat); !ok || r.Origen != origenHilo ||
		r.Texto != "formula polinomica" || r.Confianza != confianzaHilo || r.Directa() {
		t.Errorf("hilo: %+v", r)
	}
	// Nada a qué referirse.
	if r, ok := ResolverReferencia("como hago eso", tipos.Memoria{}, origenMemoria, nil, aliasFalso{}, cat); !ok || r.Hay() {
		t.Errorf("sin referente: %+v %v", r, ok)
	}
	// No elíptica: no se toca.
	if _, ok := ResolverReferencia("¿Cómo registro un metrado?", tipos.Memoria{Ofrecido: "config.unidades"}, origenMemoria, nil, aliasFalso{}, cat); ok {
		t.Error("una pregunta con su propio tema no es una referencia")
	}
}

func TestAceptaOferta(t *testing.T) {
	for _, s := range []string{"si", "sí, enséñame", "dale, muéstrame los pasos", "¿y luego?", "enséñame", "claro, por favor", "ok"} {
		if !AceptaOferta(Normalizar(s)) {
			t.Errorf("%q acepta la oferta", s)
		}
	}
	for _, s := range []string{"ya", "ok gracias", "perfecto", "no", "sí pero cómo configuro las unidades", "gracias"} {
		if AceptaOferta(Normalizar(s)) {
			t.Errorf("%q no acepta la oferta", s)
		}
	}
}

func TestMemoriaGuardaConceptoYOfrecido(t *testing.T) {
	c := conceptoMetrado()
	p := tipos.Plan{Tipo: tipos.Concepto, Concepto: &c, Opciones: []tipos.Ref{{ID: "metrados.registrar", Titulo: "Registrar un metrado"}}}
	m := MemoriaDePlan(p, tipos.Memoria{})
	if m.Concepto != "metrado" || m.Ofrecido != "metrados.registrar" || m.ProcedimientoID != "" {
		t.Fatalf("concepto con oferta: %+v", m)
	}
	// Dos opciones (aclaración): ambiguo, no se ofrece nada. Comparación: dos conceptos, ninguno.
	p2 := tipos.Plan{Tipo: tipos.Desconocido, Opciones: []tipos.Ref{{ID: "a"}, {ID: "b"}}}
	p3 := tipos.Plan{Tipo: tipos.Comparacion, Concepto: &c, Conceptos: []tipos.ConceptoDef{c, c}}
	if m := MemoriaDePlan(p2, tipos.Memoria{}); m.Ofrecido != "" {
		t.Errorf("dos opciones: %+v", m)
	}
	if m := MemoriaDePlan(p3, tipos.Memoria{}); m.Concepto != "" {
		t.Errorf("comparación: %+v", m)
	}
	// SIN_EVIDENCIA no deja nada.
	if m := MemoriaDePlan(tipos.Plan{SinEvidencia: true, Concepto: &c}, tipos.Memoria{}); m.Concepto != "" {
		t.Errorf("sin evidencia: %+v", m)
	}
	// Con lo ofrecido: «sí» lo empieza; con un procedimiento en curso manda el procedimiento.
	if a := InterpretarMemoria("sí, enséñame", m); a != MemOfrecido {
		t.Errorf("acepta: %s", a)
	}
	if a := InterpretarMemoria("¿y luego?", tipos.Memoria{Ofrecido: "x"}); a != MemOfrecido {
		t.Errorf("¿y luego? tras ofrecer: %s", a)
	}
	if a := InterpretarMemoria("sí", tipos.Memoria{ProcedimientoID: "y", PasoActual: 1, Ofrecido: "x"}); a == MemOfrecido {
		t.Errorf("en curso: %s", a)
	}
	if s := Suspender(tipos.Memoria{ProcedimientoID: "y", PasoActual: 2, Ofrecido: "x", Concepto: "c"}); s.Suspendido.Ofrecido != "" || s.Suspendido.Concepto != "" {
		t.Errorf("lo suspendido es el procedimiento: %+v", s.Suspendido)
	}
}

// --- integración con los dobles del núcleo -------------------------------------------------------------------

// constructorConOferta: el doble del constructor que, como el real, ofrece el procedimiento relacionado del concepto.
type constructorConOferta struct{ *constructorFalso }

func (c constructorConOferta) Construir(ctx context.Context, e tipos.Estado, cands []tipos.Candidato) (tipos.Plan, error) {
	p, err := c.constructorFalso.Construir(ctx, e, cands)
	if err == nil && p.Concepto != nil && p.Concepto.ID == "metrado" {
		p.Opciones = []tipos.Ref{{ID: "metrados.registrar", Titulo: "Registrar un metrado", Confianza: 1}}
		p.Siguiente = &tipos.Siguiente{Plantilla: "OFRECER_PROCEDIMIENTO", Texto: "¿Quiere que le enseñe a registrar un metrado?"}
	}
	return p, err
}

func agenteConOferta() *banco {
	b := armarAgente()
	b.ag.Constructor = constructorConOferta{b.cons}
	return b
}

// Concepto → «¿y cómo la hago?»: directo al procedimiento ofrecido, sin buscar.
func TestReferencia_ConceptoYComoLaHago(t *testing.T) {
	b := agenteConOferta()
	res, _ := preguntarConTraza(t, b.ag, "¿Qué es un metrado?", rag.Opciones{})
	if res.Memoria == nil || res.Memoria.Concepto != "metrado" || res.Memoria.Ofrecido != "metrados.registrar" {
		t.Fatalf("el concepto deja su referencia en la memoria: %+v", res.Memoria)
	}
	antes := b.rec.total()
	res, tr := preguntarConTraza(t, b.ag, "¿y cómo la hago?", rag.Opciones{Memoria: res.Memoria,
		Hilo: []rag.Turno{{Rol: "usuario", Texto: "¿Qué es un metrado?"}, {Rol: "asistente", Texto: "Cuantificación…"}}})
	comprobarV2(t, res)
	if b.rec.total() != antes {
		t.Errorf("no debía buscar: %d llamadas", b.rec.total()-antes)
	}
	if res.PlanV2.Tipo != tipos.Procedimiento || res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "metrados.registrar" ||
		res.Memoria.ProcedimientoID != "metrados.registrar" {
		t.Fatalf("plan: %+v / memoria %+v", res.PlanV2, res.Memoria)
	}
	pc := etapaV2(tr, traza.EtapaV2PlanConsulta)
	if pc.Estado != traza.EstadoOK || pc.Datos["referencia"] == nil || etapaV2(tr, traza.EtapaV2Fusion).Estado != traza.EstadoNoTomada {
		t.Fatalf("traza: %+v", pc)
	}
}

// Concepto → «sí, enséñame» y «¿y luego?»: el procedimiento ofrecido desde el paso 1 (acción de memoria).
func TestReferencia_AceptarOferta(t *testing.T) {
	for _, msg := range []string{"sí, enséñame", "¿y luego?"} {
		b := agenteConOferta()
		mem := &tipos.Memoria{Concepto: "metrado", Ofrecido: "metrados.registrar"}
		res, tr := preguntarConTraza(t, b.ag, msg, rag.Opciones{Memoria: mem})
		if b.rec.total() != 0 || res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "metrados.registrar" ||
			res.Memoria.ProcedimientoID != "metrados.registrar" || res.Memoria.Ofrecido != "" {
			t.Fatalf("%q: %+v / %+v", msg, res.PlanV2, res.Memoria)
		}
		if d := etapaV2(tr, traza.EtapaV2TipoRespuesta).Datos; d["accion_memoria"] != string(MemOfrecido) {
			t.Errorf("%q: traza %+v", msg, d)
		}
	}
	// Sin oferta, «¿y luego?» sigue pidiendo aclaración.
	b := agenteConOferta()
	if res, _ := preguntarConTraza(t, b.ag, "¿y luego?", rag.Opciones{Memoria: &tipos.Memoria{Concepto: "metrado"}}); res.Respuesta != TextoSinProcedimiento {
		t.Errorf("sin oferta: %q", res.Respuesta)
	}
}

// «¿y eso dónde está?» tras un concepto: NAVIGATION con el procedimiento ofrecido y el concepto como candidatos.
func TestReferencia_DondeEsta(t *testing.T) {
	b := agenteConOferta()
	res, _ := preguntarConTraza(t, b.ag, "¿y eso dónde está?", rag.Opciones{Memoria: &tipos.Memoria{Concepto: "metrado", Ofrecido: "metrados.registrar"}})
	if b.rec.total() != 0 || res.PlanV2.Tipo != tipos.Navegacion {
		t.Fatalf("navegación: %d búsquedas, %+v", b.rec.total(), res.PlanV2)
	}
	e := b.cons.estados[len(b.cons.estados)-1]
	if e.Inferencias[claveReferenciaResuelta].Origen != origenMemoria {
		t.Errorf("la referencia va al estado como inferencia: %+v", e.Inferencias)
	}
}

// Sin memoria: el referente sale del hilo y entra en la consulta normalizada (la original no cambia).
func TestReferencia_DelHilo(t *testing.T) {
	b := armarAgente()
	hilo := []rag.Turno{{Rol: "usuario", Texto: "metrado, que viene a ser"}, {Rol: "asistente", Texto: "¿Me cuenta qué tarea…?"}}
	res, tr := preguntarConTraza(t, b.ag, "¿y cómo lo hago?", rag.Opciones{Hilo: hilo})
	if b.rec.total() == 0 {
		t.Fatal("con un referente de texto se busca")
	}
	c := b.rec.llamadas[0].consulta
	if c.Original != "¿y cómo lo hago?" || !strings.Contains(c.Normalizada, "metrado") {
		t.Fatalf("consulta: %+v", c)
	}
	if res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "metrados.registrar" {
		t.Fatalf("plan: %+v", res.PlanV2)
	}
	if d := etapaV2(tr, traza.EtapaV2PlanConsulta).Datos; d["origen_referencia"] != origenHilo {
		t.Errorf("traza: %+v", d)
	}
}

// Elíptica y sin nada a qué referirse: aclaración, sin buscar. Una pregunta con su propio tema no cambia.
func TestReferencia_SinReferenteYNoEliptica(t *testing.T) {
	b := armarAgente()
	res, _ := preguntarConTraza(t, b.ag, "como hago eso", rag.Opciones{})
	if res.Modo != "aclaracion" || res.Respuesta != TextoSinReferente || b.rec.total() != 0 {
		t.Fatalf("sin referente: %+v (%d búsquedas)", res, b.rec.total())
	}
	b = agenteConOferta()
	res, _ = preguntarConTraza(t, b.ag, "¿Cómo configuro las unidades?", rag.Opciones{Memoria: &tipos.Memoria{Concepto: "metrado", Ofrecido: "metrados.registrar"}})
	if res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "config.unidades" || b.rec.total() == 0 {
		t.Fatalf("no elíptica: %+v", res.PlanV2)
	}
}

// Si la oferta que se ve es retomar un procedimiento suspendido, «ofrecido» no queda en la memoria.
func TestReferencia_RetomaTapaLaOferta(t *testing.T) {
	b := agenteConOferta()
	mem := &tipos.Memoria{ProcedimientoID: "config.unidades", PasoActual: 1}
	res, _ := preguntarConTraza(t, b.ag, "¿Qué es un metrado?", rag.Opciones{Memoria: mem})
	if res.PlanV2.Siguiente == nil || res.PlanV2.Siguiente.Plantilla != "RETOMA" || res.Memoria.Ofrecido != "" || res.Memoria.Suspendido == nil {
		t.Fatalf("retoma: %+v / %+v", res.PlanV2.Siguiente, res.Memoria)
	}
}

// --- integración con kb/ real ---------------------------------------------------------------------------------

// agenteReal: el núcleo con el conocimiento real (como cmd/rag/v2.go, sin búsqueda híbrida ni clasificador kNN).
func agenteReal(t *testing.T) *Agente {
	t.Helper()
	const kb = "../../../kb"
	if _, err := os.Stat(kb + "/fragmentos.jsonl"); err != nil {
		t.Skip("sin kb/ real")
	}
	b := conocimiento.Cargar(conocimiento.Opciones{DirKB: kb, ArchivoPlantillas: "../../plantillas/respuestas.yml"})
	rec := conocimiento.NuevoRecuperador(b)
	cfg := ConfigDefecto()
	cfg.Version = tipos.V2
	return &Agente{Config: cfg, Aliaser: b, Catalogo: b, Recuperador: rec, Constructor: conocimiento.NuevoConstructor(b, rec),
		Plantillas: conocimiento.NuevoMotorPlantillas(b), Gate: conocimiento.NuevoGate(b)}
}

func TestReferencia_KBReal(t *testing.T) {
	ag := agenteReal(t)
	ctx := context.Background()
	turno := func(q string, mem *tipos.Memoria, hilo []rag.Turno) rag.Respuesta {
		t.Helper()
		res, err := ag.Preguntar(ctx, q, rag.Opciones{Memoria: mem, Hilo: hilo})
		if err != nil || res.PlanV2 == nil || res.Memoria == nil {
			t.Fatalf("%q: %+v %v", q, res, err)
		}
		return res
	}
	c := turno("qué es el kardex", nil, nil)
	if c.PlanV2.Concepto == nil || c.Memoria.Concepto != c.PlanV2.Concepto.ID {
		t.Fatalf("concepto: %+v / %+v", c.PlanV2, c.Memoria)
	}
	if c.Memoria.Ofrecido == "" {
		t.Skipf("el glosario ya no ofrece un procedimiento para %s", c.PlanV2.Concepto.ID)
	}
	hilo := []rag.Turno{{Rol: "usuario", Texto: "qué es el kardex"}, {Rol: "asistente", Texto: c.Respuesta}}
	for _, q := range []string{"sí, enséñame", "¿y cómo se hace eso?"} {
		r := turno(q, c.Memoria, hilo)
		if r.PlanV2.Procedimiento == nil || r.PlanV2.Procedimiento.ID != c.Memoria.Ofrecido || len(r.PlanV2.Pasos) == 0 || r.Modo != "respuesta" {
			t.Errorf("%q: quiero %s; plan %+v, modo %s, motivo %s", q, c.Memoria.Ofrecido, r.PlanV2.Procedimiento, r.Modo, r.Motivo)
		}
	}
}
