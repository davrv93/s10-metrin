package v2

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"rag-go/internal/rag"
	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

// preguntarConTraza responde un turno V2 con el modo traza encendido y comprueba el contrato de las etapas.
func preguntarConTraza(t *testing.T, ag *Agente, pregunta string, o rag.Opciones) (rag.Respuesta, *traza.Traza) {
	t.Helper()
	rec := traza.Nuevo(pregunta)
	res, err := ag.Preguntar(traza.ConContexto(context.Background(), rec), pregunta, o)
	if err != nil {
		t.Fatal(err)
	}
	tr := rec.Cerrar()
	if tr.Agente != "v2" || len(tr.EtapasV2) != len(traza.IDsV2()) || len(tr.Etapas) != len(traza.IDs()) {
		t.Fatalf("traza V2: agente %q, %d etapas V2, %d etapas V1", tr.Agente, len(tr.EtapasV2), len(tr.Etapas))
	}
	for i, id := range traza.IDsV2() {
		e := tr.EtapasV2[i]
		if e.ID != id || !traza.EstadoValido(e.Estado) {
			t.Fatalf("etapa V2 %d: %+v", i, e)
		}
		if (e.Estado == traza.EstadoOmitida || e.Estado == traza.EstadoNoTomada) && e.Razon == "" {
			t.Errorf("%s: %s sin razón", id, e.Estado)
		}
		for _, k := range traza.MinimosV2(id) {
			if _, ok := e.Datos[k]; !ok {
				t.Errorf("%s: falta el dato mínimo %q", id, k)
			}
		}
	}
	for _, e := range tr.Etapas {
		if e.Estado == "" || ((e.Estado == traza.EstadoOmitida || e.Estado == traza.EstadoNoTomada) && e.Razon == "") {
			t.Errorf("V1 %s: %+v", e.ID, e)
		}
	}
	if b, err := json.Marshal(tr); err != nil || len(b) > traza.MaxBytes {
		t.Errorf("traza V2 de %d bytes (%v)", len(b), err)
	}
	return res, tr
}

func etapaV2(tr *traza.Traza, id string) traza.Etapa {
	for _, e := range tr.EtapasV2 {
		if e.ID == id {
			return e
		}
	}
	return traza.Etapa{}
}

func oportunidadEn(tr *traza.Traza, etapa string) bool {
	for _, o := range tr.Oportunidades {
		if o.Etapa == etapa {
			return true
		}
	}
	return false
}

func comprobarV2(t *testing.T, res rag.Respuesta) {
	t.Helper()
	if res.Version != tipos.V2 || res.PlanV2 == nil || res.Memoria == nil || res.PlanV2.Version != 2 || res.Fuentes == nil {
		t.Fatalf("respuesta V2 incompleta: %+v", res)
	}
	if res.Plan.Ruta != "v2" && res.Plan.Ruta != "" {
		t.Fatalf("ruta: %+v", res.Plan)
	}
}

// --- Recursividad controlada -------------------------------------------------------------------

func TestEvidenciaInsuficienteBuscaOtraVezYBasta(t *testing.T) {
	b := armarAgente()
	b.rec.responder = func(n int, c tipos.Consulta, clase string) ([]tipos.Candidato, []string, error) {
		if n == 1 {
			return nil, nil, nil // primera búsqueda: nada
		}
		return porTema(n, c, clase)
	}
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	comprobarV2(t, res)
	if res.Modo != "respuesta" || res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "metrados.registrar" {
		t.Fatalf("debió responder con el procedimiento tras expandir: %+v", res)
	}
	// 1 llamada en la 1.ª iteración (procedimientos) + 2 en la 2.ª (procedimientos y fragmentos).
	if b.rec.total() != 3 || b.rec.llamadas[1].consulta.Original != "¿Cómo registro un metrado?" {
		t.Fatalf("llamadas: %+v", b.rec.llamadas)
	}
	if x := b.rec.llamadas[1].consulta; strings.Contains(" "+x.Normalizada+" ", " como ") || len(x.Aliases) == 0 {
		t.Fatalf("la 2.ª búsqueda va expandida (compacta + alias): %+v", x)
	}
	if d := etapaV2(tr, "plan_consulta").Datos; d["iteraciones"] != 2 || d["expandida"] != true {
		t.Fatalf("plan_consulta: %+v", d)
	}
	if m := res.Memoria; m.ProcedimientoID != "metrados.registrar" || m.PasoActual != 3 {
		t.Fatalf("memoria tras mostrar los 3 pasos: %+v", m)
	}
}

func TestInsuficienteDosVecesEsSinEvidencia(t *testing.T) {
	b := armarAgente()
	b.rec.responder = nil // nunca encuentra nada
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	comprobarV2(t, res)
	if !res.SinContexto || !res.PlanV2.SinEvidencia || res.Modo != "sin_contexto" || res.Respuesta != TextoSinEvidencia {
		t.Fatalf("SIN_EVIDENCIA: %+v", res)
	}
	if b.rec.total() != 3 || len(b.cons.estados) != 0 {
		t.Fatalf("2 iteraciones (3 llamadas) y sin construir plan: %d llamadas", b.rec.total())
	}
	if !strings.Contains(res.Motivo, "MAX_SEARCH_ITERATIONS") || len(res.Fuentes) != 0 {
		t.Fatalf("motivo: %q", res.Motivo)
	}
	if etapaV2(tr, "plan").Datos["sin_evidencia"] != true || !oportunidadEn(tr, "plan") || etapaV2(tr, "quality_gate").Estado != traza.EstadoOmitida {
		t.Fatalf("traza: %+v", tr.EtapasV2)
	}
}

func TestCaidaDelLLMUsaPlantillas(t *testing.T) {
	b := armarAgente()
	b.gen = &generadorFalso{nombre: "llm-local", err: errCaido}
	b.ag.Generador = b.gen
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	if res.Modo != "respuesta" || strings.HasPrefix(res.Respuesta, "LLM:") || b.gen.llamadas != 1 {
		t.Fatalf("sin LLM → plantillas: %+v", res)
	}
	if !strings.Contains(res.Motivo, "sin LLM → plantillas") || !strings.Contains(res.Respuesta, "Abra la hoja del presupuesto.") {
		t.Fatalf("degradación anotada: %q / %q", res.Motivo, res.Respuesta)
	}
	if e := etapaV2(tr, "renderizado"); e.Estado != traza.EstadoRespaldo || !oportunidadEn(tr, "renderizado") {
		t.Fatalf("renderizado: %+v", e)
	}
	// Con el LLM vivo, redacta él.
	b2 := armarAgente()
	b2.ag.Generador = &generadorFalso{nombre: "llm-local"}
	if res, _ := b2.ag.Preguntar(context.Background(), "¿Cómo registro un metrado?", rag.Opciones{}); !strings.HasPrefix(res.Respuesta, "LLM:") {
		t.Fatalf("con LLM: %q", res.Respuesta)
	}
}

func TestCaidaDelRerankerQuedaAnotada(t *testing.T) {
	b := armarAgente()
	b.rec.responder = func(n int, c tipos.Consulta, clase string) ([]tipos.Candidato, []string, error) {
		r, _, err := porTema(n, c, clase)
		return r, []string{"sin_reranker: connection refused"}, err
	}
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	if res.Modo != "respuesta" || !strings.Contains(res.Motivo, "sin_reranker") {
		t.Fatalf("responde con el puntaje híbrido y lo anota: %+v", res)
	}
	if e := etapaV2(tr, "rerank"); e.Estado != traza.EstadoRespaldo || !oportunidadEn(tr, "rerank") {
		t.Fatalf("rerank: %+v / %+v", e, tr.Oportunidades)
	}
	if e := etapaV2(tr, "busqueda_vectorial"); e.Estado != traza.EstadoOK {
		t.Fatalf("el vector sí estaba: %+v", e)
	}
}

func TestGateFallaDosVecesRespuestaSegura(t *testing.T) {
	b := armarAgente()
	b.gate.pasa = []bool{false, false}
	b.ag.Generador = &generadorFalso{nombre: "llm-local"}
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	if res.Modo != "respuesta_segura" || !strings.Contains(res.Motivo, "quality gate") {
		t.Fatalf("respuesta segura: %+v", res)
	}
	if len(b.gate.textos) != 2 || !strings.HasPrefix(b.gate.textos[0], "LLM:") || strings.HasPrefix(b.gate.textos[1], "LLM:") {
		t.Fatalf("un reintento y SIN LLM: %q", b.gate.textos)
	}
	for _, p := range res.PlanV2.Pasos {
		if len(p.Fotos) != 0 || len(p.Fuente) == 0 {
			t.Fatalf("la segura solo lleva pasos con fuente y sin fotos: %+v", p)
		}
	}
	if res.Respuesta != RenderBasico(*res.PlanV2) {
		t.Fatalf("la segura la redacta el renderizador básico: %q", res.Respuesta)
	}
	if e := etapaV2(tr, "quality_gate"); e.Estado != traza.EstadoError || e.Datos["intentos"] != 2 || !oportunidadEn(tr, "quality_gate") {
		t.Fatalf("quality_gate: %+v", e)
	}
}

func TestGateFallaUnaVezRegeneraSinLLM(t *testing.T) {
	b := armarAgente()
	b.gate.pasa = []bool{false, true}
	b.ag.Generador = &generadorFalso{nombre: "llm-local"}
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	if res.Modo != "respuesta" || strings.HasPrefix(res.Respuesta, "LLM:") {
		t.Fatalf("pasó al regenerar sin LLM: %+v", res)
	}
	if e := etapaV2(tr, "quality_gate"); e.Estado != traza.EstadoAlerta || e.Datos["passed"] != true {
		t.Fatalf("quality_gate: %+v", e)
	}
	if d := etapaV2(tr, "renderizado").Datos; d["modo"] != "regenerado_sin_llm" {
		t.Fatalf("renderizado: %+v", d)
	}
}

func TestLimiteDePasosDaRespuestaSegura(t *testing.T) {
	b := armarAgente()
	b.ag.Config.Limites.MaxPasos = 2 // buscar + construir: no queda para redactar
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	if res.Modo != "respuesta_segura" || !strings.Contains(res.Motivo, "MAX_AGENT_STEPS") {
		t.Fatalf("límite → respuesta segura: %+v", res)
	}
	if d := etapaV2(tr, "renderizado").Datos["presupuesto"].(map[string]any); d["pasos"] != "2/2" {
		t.Fatalf("presupuesto: %+v", d)
	}
	b.ag.Config.Limites.MaxPasos = 0
	if res, _ := b.ag.Preguntar(context.Background(), "¿Cómo registro un metrado?", rag.Opciones{}); !res.SinContexto {
		t.Fatalf("sin pasos no hay búsqueda: %+v", res)
	}
}

func TestErroresDeConocimientoNoFallanEnSilencio(t *testing.T) {
	b := armarAgente()
	b.cons.err = errCaido
	res, _ := b.ag.Preguntar(context.Background(), "¿Cómo registro un metrado?", rag.Opciones{})
	if !res.SinContexto || !strings.Contains(res.Motivo, "constructor") {
		t.Fatalf("constructor caído: %+v", res)
	}
	b2 := armarAgente()
	b2.rec.responder = func(int, tipos.Consulta, string) ([]tipos.Candidato, []string, error) { return nil, nil, errCaido }
	res, tr := preguntarConTraza(t, b2.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	if !res.SinContexto || !strings.Contains(res.Motivo, "recuperador") || etapaV2(tr, "fusion").Estado != traza.EstadoError {
		t.Fatalf("recuperador caído: %+v", res)
	}
	vacio := &Agente{Config: b.ag.Config}
	if res, err := vacio.Preguntar(context.Background(), "¿Cómo registro un metrado?", rag.Opciones{}); err != nil || !res.SinContexto || !strings.Contains(res.Motivo, "sin recuperador") {
		t.Fatalf("sin piezas: %+v %v", res, err)
	}
}

func TestTipoDudosoConMotorFueraDeOpciones(t *testing.T) {
	b := armarAgente()
	b.ag.Decision = &Simulada{Elecciones: map[string]string{DecTipoRespuesta: "POEMA"}}
	res, tr := preguntarConTraza(t, b.ag, "metrados", rag.Opciones{})
	if res.PlanV2.Tipo != tipos.Desconocido || res.Modo != "aclaracion" || b.rec.total() != 0 {
		t.Fatalf("fuera de opciones → reglas → UNKNOWN (aclarar, sin buscar): %+v", res)
	}
	if e := etapaV2(tr, "decision"); e.Estado != traza.EstadoRespaldo || e.Datos["respaldo"] != true || !oportunidadEn(tr, "decision") {
		t.Fatalf("decision: %+v", e)
	}
	if e := etapaV2(tr, "tipo_respuesta"); e.Datos["dudoso"] != true || !oportunidadEn(tr, "tipo_respuesta") {
		t.Fatalf("tipo_respuesta: %+v", e)
	}
	// Un motor que sí decide dentro de las opciones manda.
	b2 := armarAgente()
	b2.ag.Decision = &Simulada{Elecciones: map[string]string{DecTipoRespuesta: "PROCEDURE"}}
	if res, _ := b2.ag.Preguntar(context.Background(), "metrados", rag.Opciones{}); res.PlanV2.Tipo != tipos.Procedimiento {
		t.Fatalf("el motor decide: %+v", res.PlanV2)
	}
}

func TestZonaGrisDeEvidenciaLaDecideElMotor(t *testing.T) {
	b := armarAgente()
	b.ag.Config.EvidenciaMin = 0.95 // el procedimiento (0,9) queda en la zona gris
	sim := &Simulada{Elecciones: map[string]string{DecSuficiencia: Suficiente}}
	b.ag.Decision = sim
	res, _ := b.ag.Preguntar(context.Background(), "¿Cómo registro un metrado?", rag.Opciones{})
	if res.Modo != "respuesta" || sim.Llamadas() != 1 || b.rec.total() != 1 {
		t.Fatalf("el motor dio la evidencia por suficiente: %+v (llamadas %d, búsquedas %d)", res.Modo, sim.Llamadas(), b.rec.total())
	}
}

func TestEmpateDeProcedimientosLoDecideElMotor(t *testing.T) {
	b := armarAgente()
	b.rec.responder = func(_ int, _ tipos.Consulta, clase string) ([]tipos.Candidato, []string, error) {
		if clase != "procedimiento" {
			return nil, nil, nil
		}
		return []tipos.Candidato{
			{ID: "metrados.registrar", Clase: clase, Puntaje: 0.80},
			{ID: "config.unidades", Clase: clase, Puntaje: 0.79},
		}, nil, nil
	}
	b.ag.Decision = &Simulada{Elecciones: map[string]string{DecProcedimiento: "config.unidades"}}
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro las unidades del metrado?", rag.Opciones{})
	if res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "config.unidades" {
		t.Fatalf("el motor eligió el segundo: %+v", res.PlanV2.Procedimiento)
	}
	if d := etapaV2(tr, "procedimiento").Datos; d["decidido_por"] != "decision:simulada" {
		t.Fatalf("procedimiento: %+v", d)
	}
}

// Con el reranker de clase, el empate se mide en logits: dos sigmoides casi iguales (0,998 y 0,993) con logits
// separados no van al motor; dos logits a menos de logit(0,6) sí.
func TestEmpateConRerankerSeMideEnLogits(t *testing.T) {
	for _, x := range []struct {
		l0, l1   float64
		decision bool
	}{{6.2, 5.0, false}, {5.2, 5.0, true}} {
		b := armarAgente()
		l0, l1 := x.l0, x.l1
		b.rec.responder = func(_ int, _ tipos.Consulta, clase string) ([]tipos.Candidato, []string, error) {
			if clase != "procedimiento" {
				return nil, nil, nil
			}
			return []tipos.Candidato{
				{ID: "metrados.registrar", Clase: clase, Puntaje: 0.998, Rerank: &l0},
				{ID: "config.unidades", Clase: clase, Puntaje: 0.993, Rerank: &l1},
			}, nil, nil
		}
		sim := &Simulada{Elecciones: map[string]string{DecProcedimiento: "config.unidades"}}
		b.ag.Decision = sim
		res, _ := preguntarConTraza(t, b.ag, "¿Cómo registro las unidades del metrado?", rag.Opciones{})
		if got := res.PlanV2.Procedimiento != nil && res.PlanV2.Procedimiento.ID == "config.unidades"; got != x.decision {
			t.Errorf("logits %.1f/%.1f: decidió el motor = %v, quiero %v (%+v)", l0, l1, got, x.decision, res.PlanV2.Procedimiento)
		}
	}
}

// La etapa rerank refleja el reranker de clase: reordenó → ok y activo aunque los fragmentos no lo usen; falló →
// respaldo con la clase y el motivo.
func TestTrazaRerankDeClase(t *testing.T) {
	for _, x := range []struct {
		datos  map[string]any
		estado string
		razon  string
	}{
		{map[string]any{"procedimiento": traza.Datos{"reordenado": true, "fallo": nil}}, traza.EstadoOK, "reranker para elegir: procedimiento"},
		{map[string]any{"concepto": traza.Datos{"reordenado": false, "fallo": "timeout"}}, traza.EstadoRespaldo, "falló en concepto (timeout)"},
	} {
		b := armarAgente()
		datos := x.datos
		b.rec.responder = func(n int, c tipos.Consulta, clase string) ([]tipos.Candidato, []string, error) {
			return porTema(n, c, clase)
		}
		ag := b.ag
		ag.Recuperador = recuperadorConTraza{b.rec, datos}
		_, tr := preguntarConTraza(t, ag, "¿Cómo registro un metrado?", rag.Opciones{})
		e := etapaV2(tr, "rerank")
		if e.Estado != x.estado || !strings.Contains(e.Razon, x.razon) {
			t.Errorf("etapa rerank: %s %q, quiero %s con %q", e.Estado, e.Razon, x.estado, x.razon)
		}
		if x.estado == traza.EstadoOK && e.Datos["activo"] != true {
			t.Errorf("activo: %+v", e.Datos)
		}
	}
}

// recuperadorConTraza deja el dato «clases» en la etapa rerank, como el recuperador real.
type recuperadorConTraza struct {
	r     *recuperadorFalso
	datos map[string]any
}

func (x recuperadorConTraza) Buscar(ctx context.Context, c tipos.Consulta, clase string, k int) ([]tipos.Candidato, []string, error) {
	traza.De(ctx).V2().Dato(traza.EtapaV2Rerank, claveRerankClases, x.datos)
	return x.r.Buscar(ctx, c, clase, k)
}

// --- Integración con dobles ----------------------------------------------------------------------

func TestIntegracionPreguntaConceptual(t *testing.T) {
	b := armarAgente()
	res, tr := preguntarConTraza(t, b.ag, "¿Qué es un metrado?", rag.Opciones{})
	comprobarV2(t, res)
	if res.PlanV2.Tipo != tipos.Concepto || res.PlanV2.Concepto == nil || len(res.PlanV2.Pasos) != 0 {
		t.Fatalf("concepto sin pasos: %+v", res.PlanV2)
	}
	if b.rec.llamadas[0].clase != "concepto" || !strings.Contains(res.Respuesta, "Cuantificación") {
		t.Fatalf("busca en conceptos y define: %+v / %q", b.rec.llamadas, res.Respuesta)
	}
	if res.Plan.TipoConsulta != "concepto" || res.Memoria.ProcedimientoID != "" || etapaV2(tr, "pasos").Estado != traza.EstadoOmitida {
		t.Fatalf("orquestación/memoria/traza: %+v %+v", res.Plan, res.Memoria)
	}
}

func TestIntegracionPreguntaProcedimentalConFotos(t *testing.T) {
	b := armarAgente()
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	comprobarV2(t, res)
	p := res.PlanV2
	if p.Tipo != tipos.Procedimiento || len(p.Pasos) != 3 {
		t.Fatalf("plan: %+v", p)
	}
	// Cada foto en SU paso: 1 → img_1, 2 → ninguna, 3 → img_3.
	quiero := map[int]string{1: "img_1", 2: "", 3: "img_3"}
	for _, x := range p.Pasos {
		got := ""
		if len(x.Fotos) > 0 {
			got = x.Fotos[0].ID
		}
		if got != quiero[x.N] || len(x.Fotos) > 1 {
			t.Errorf("paso %d: fotos %+v", x.N, x.Fotos)
		}
	}
	if d := etapaV2(tr, "fotos_paso").Datos; d["fotos"] != 2 || d["pasos_con_foto"] != 2 {
		t.Fatalf("fotos_paso: %+v", d)
	}
	if len(res.Fuentes) != 1 || res.Fuentes[0].Cita != "Manual de Presupuestos · p. 11, 12" {
		t.Fatalf("fuentes en el formato de V1: %+v", res.Fuentes)
	}
	if e := etapaV2(tr, "tipo_respuesta"); e.Datos["metodo"] != "reglas" || e.Datos["tipo"] != "PROCEDURE" {
		t.Fatalf("tipo: %+v", e)
	}
	// El estado que recibe el constructor separa hechos e inferencias.
	if e := b.cons.estados[0]; e.Tipo != tipos.Procedimiento || e.Hechos["evidencia"].Origen != origenBusqueda {
		t.Fatalf("estado del constructor: %+v", e)
	}
}

func TestIntegracionPreguntaAmbigua(t *testing.T) {
	b := armarAgente()
	res, tr := preguntarConTraza(t, b.ag, "metrados", rag.Opciones{})
	if res.Modo != "aclaracion" || res.PlanV2.Tipo != tipos.Desconocido || res.PlanV2.Plantilla != "ACLARAR_TAREA" || b.rec.total() != 0 {
		t.Fatalf("ambigua → aclarar sin buscar: %+v", res)
	}
	if res.Respuesta != TextoAclarar || etapaV2(tr, "plan_consulta").Estado == traza.EstadoOK {
		t.Fatalf("aclaración: %q", res.Respuesta)
	}
}

func TestIntegracionSocialUsaLaCharlaDeV1(t *testing.T) {
	b := armarAgente()
	mem := &tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 2}
	res, tr := preguntarConTraza(t, b.ag, "gracias", rag.Opciones{Memoria: mem})
	if res.Modo != "conversacional" || len(b.charl.intenciones) != 1 || b.rec.total() != 0 {
		t.Fatalf("social: %+v", res)
	}
	if res.Memoria.ProcedimientoID != "metrados.registrar" || res.Memoria.PasoActual != 2 {
		t.Fatalf("una cortesía no toca el procedimiento en curso: %+v", res.Memoria)
	}
	if e := etapaV2(tr, "busqueda_lexica"); e.Estado != traza.EstadoNoTomada {
		t.Fatalf("social sin búsqueda: %+v", e)
	}
}

func TestIntegracionContinuacion(t *testing.T) {
	b := armarAgente()
	mem := &tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 1}
	res, tr := preguntarConTraza(t, b.ag, "¿y luego?", rag.Opciones{Memoria: mem})
	comprobarV2(t, res)
	if b.rec.total() != 0 {
		t.Fatalf("«¿y luego?» NO busca: %d llamadas", b.rec.total())
	}
	if v := PasosVisibles(*res.PlanV2); len(v) != 1 || v[0].N != 2 || res.Memoria.PasoActual != 2 || len(res.PlanV2.Pasos) != 3 {
		t.Fatalf("paso siguiente: %+v / %+v", res.PlanV2, res.Memoria)
	}
	if e := etapaV2(tr, "tipo_respuesta"); e.Datos["metodo"] != "memoria" || etapaV2(tr, "plan_consulta").Estado != traza.EstadoNoTomada {
		t.Fatalf("traza de la memoria: %+v", e)
	}
	// «listo» → paso 3 (último, con verificación); «listo» otra vez → fin y memoria vacía.
	res, _ = b.ag.Preguntar(context.Background(), "listo", rag.Opciones{Memoria: res.Memoria})
	if v := PasosVisibles(*res.PlanV2); v[0].N != 3 || len(res.PlanV2.Verificacion) != 1 || v[0].Fotos[0].ID != "img_3" {
		t.Fatalf("paso 3: %+v", res.PlanV2)
	}
	res, _ = b.ag.Preguntar(context.Background(), "ya", rag.Opciones{Memoria: res.Memoria})
	if res.PlanV2.Plantilla != "PROCEDIMIENTO_FIN" || res.Memoria.ProcedimientoID != "" || b.rec.total() != 0 {
		t.Fatalf("fin: %+v / %+v", res.PlanV2, res.Memoria)
	}
	// «no me sale» → errores frecuentes del paso en curso, sin buscar.
	res, _ = b.ag.Preguntar(context.Background(), "no me sale", rag.Opciones{Memoria: &tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 2}})
	if res.PlanV2.Tipo != tipos.Problema || len(res.PlanV2.Errores) != 1 || res.Memoria.PasoActual != 2 || b.rec.total() != 0 {
		t.Fatalf("no me sale: %+v", res.PlanV2)
	}
}

func TestIntegracionContinuacionDesdeElHilo(t *testing.T) {
	b := armarAgente()
	primera, _ := b.ag.Preguntar(context.Background(), "¿Cómo registro un metrado?", rag.Opciones{})
	paso1, _ := PlanPaso(procMetrado(), 1)
	plan, _ := json.Marshal(paso1)
	_ = primera
	hilo := []rag.Turno{{Rol: "usuario", Texto: "¿Cómo registro un metrado?"}, {Rol: "asistente", Texto: "Paso 1…", Plan: plan}}
	antes := b.rec.total()
	res, _ := b.ag.Preguntar(context.Background(), "siguiente", rag.Opciones{Hilo: hilo})
	if PasosVisibles(*res.PlanV2)[0].N != 2 || b.rec.total() != antes {
		t.Fatalf("memoria reconstruida del último plan del hilo: %+v", res.PlanV2)
	}
	// «¿y luego?» sin nada en curso: aclarar, nunca buscar a ciegas.
	res, _ = b.ag.Preguntar(context.Background(), "¿y luego?", rag.Opciones{})
	if res.Modo != "aclaracion" || res.Respuesta != TextoSinProcedimiento || b.rec.total() != antes {
		t.Fatalf("sin procedimiento: %+v", res)
	}
}

func TestIntegracionCambioDeTemaRetomaUnaVez(t *testing.T) {
	b := armarAgente()
	en := &tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 2}
	// Cambio de tema: se contesta y se ofrece retomar.
	res, _ := b.ag.Preguntar(context.Background(), "¿Qué es un metrado?", rag.Opciones{Memoria: en})
	if res.PlanV2.Tipo != tipos.Concepto || res.Memoria.Suspendido == nil || res.Memoria.ProcedimientoID != "" {
		t.Fatalf("suspendido: %+v / %+v", res.PlanV2, res.Memoria)
	}
	if s := res.PlanV2.Siguiente; s == nil || s.Plantilla != "RETOMA" || !strings.Contains(s.Texto, "Registrar un metrado") || s.Paso != 2 {
		t.Fatalf("oferta RETOMA: %+v", s)
	}
	busquedas := b.rec.total()
	// «sí» → se retoma en el paso en curso, sin buscar.
	ret, _ := b.ag.Preguntar(context.Background(), "sí", rag.Opciones{Memoria: res.Memoria})
	if ret.Memoria.ProcedimientoID != "metrados.registrar" || !ret.Memoria.Retomado || PasosVisibles(*ret.PlanV2)[0].N != 2 || b.rec.total() != busquedas {
		t.Fatalf("retomado: %+v / %+v", ret.Memoria, ret.PlanV2)
	}
	// Otro cambio de tema con el procedimiento ya retomado: se abandona (se retoma UNA vez).
	otra, _ := b.ag.Preguntar(context.Background(), "¿Qué es un metrado?", rag.Opciones{Memoria: ret.Memoria})
	if otra.Memoria.Suspendido != nil || otra.Memoria.ProcedimientoID != "" || (otra.PlanV2.Siguiente != nil && otra.PlanV2.Siguiente.Plantilla == "RETOMA") {
		t.Fatalf("abandonado: %+v", otra.Memoria)
	}
	// Si tras la oferta insiste en el tema nuevo, también se abandona.
	insiste, _ := b.ag.Preguntar(context.Background(), "¿Cómo configuro las unidades?", rag.Opciones{Memoria: res.Memoria})
	if insiste.Memoria.Suspendido != nil || insiste.Memoria.ProcedimientoID != "config.unidades" {
		t.Fatalf("insiste: %+v", insiste.Memoria)
	}
}

func TestIntegracionSinEvidenciaYProcedimientoDesconocido(t *testing.T) {
	b := armarAgente()
	res, _ := b.ag.Preguntar(context.Background(), "¿Cómo emito una guía de remisión?", rag.Opciones{})
	if !res.SinContexto || res.PlanV2.Plantilla != "SIN_EVIDENCIA" || len(res.Fuentes) != 0 {
		t.Fatalf("sin evidencia: %+v", res)
	}
	// Memoria que apunta a un procedimiento que ya no está cargado: aclarar y vaciar.
	res, _ = b.ag.Preguntar(context.Background(), "¿y luego?", rag.Opciones{Memoria: &tipos.Memoria{ProcedimientoID: "borrado", PasoActual: 1}})
	if res.Modo != "aclaracion" || res.Memoria.ProcedimientoID != "" || !strings.Contains(res.Motivo, "no está cargado") {
		t.Fatalf("procedimiento desconocido: %+v", res)
	}
}

func TestVersionDelAgente(t *testing.T) {
	a := &Agente{Config: ConfigDefecto()}
	if a.Version("x", rag.Opciones{Version: "v2"}) != tipos.V1 {
		t.Fatal("apagada: siempre v1")
	}
	a.Config.Habilitada = true
	if a.Version("x", rag.Opciones{Version: "v2"}) != tipos.V2 || a.Version("x", rag.Opciones{}) != tipos.V1 {
		t.Fatal("habilitada: la petición manda")
	}
}

// continuadorFalso: un Constructor que además sabe continuar (como el de conocimiento).
type continuadorFalso struct {
	*constructorFalso
	vistas []tipos.Memoria
}

func (c *continuadorFalso) Continuar(mem tipos.Memoria) (tipos.Plan, error) {
	c.vistas = append(c.vistas, mem)
	def, _ := c.cat.Procedimiento(mem.ProcedimientoID)
	if mem.PasoActual >= len(def.Pasos) {
		return tipos.Plan{Version: 2, Tipo: tipos.Procedimiento, Procedimiento: refDe(def), Plantilla: "PROCEDIMIENTO_FIN",
			Pasos: pasosPlan(def), Fuentes: fuentesPlan(def, nil)}, nil
	}
	p, _ := PlanPaso(def, mem.PasoActual+1)
	p.Plantilla = "PASOS_BLOQUE"
	return p, nil
}

func TestContinuacionDelegaEnElConstructor(t *testing.T) {
	b := armarAgente()
	cont := &continuadorFalso{constructorFalso: b.cons}
	b.ag.Constructor = cont
	res, _ := b.ag.Preguntar(context.Background(), "¿y luego?", rag.Opciones{Memoria: &tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 1}})
	if len(cont.vistas) != 1 || cont.vistas[0].PasoActual != 1 || res.PlanV2.Plantilla != "PASOS_BLOQUE" || res.Memoria.PasoActual != 2 || b.rec.total() != 0 {
		t.Fatalf("continuar con el constructor: %+v / %+v", cont.vistas, res.Memoria)
	}
	// Retomar: Continuar recibe el paso anterior al suspendido (enseña desde el suspendido).
	sus := &tipos.Memoria{Suspendido: &tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 2}}
	res, _ = b.ag.Preguntar(context.Background(), "retomemos", rag.Opciones{Memoria: sus})
	if cont.vistas[1].PasoActual != 1 || PasosVisibles(*res.PlanV2)[0].N != 2 || !res.Memoria.Retomado {
		t.Fatalf("retomar con el constructor: %+v / %+v", cont.vistas, res.Memoria)
	}
	// Fin: la memoria se vacía.
	res, _ = b.ag.Preguntar(context.Background(), "listo", rag.Opciones{Memoria: &tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 3}})
	if res.PlanV2.Plantilla != "PROCEDIMIENTO_FIN" || res.Memoria.ProcedimientoID != "" {
		t.Fatalf("fin: %+v", res.Memoria)
	}
}

func TestPlanConsultaPorTipo(t *testing.T) {
	for tipo, quiero := range map[tipos.TipoRespuesta][2]string{
		tipos.Procedimiento: {"procedimiento", "procedimiento,fragmento"},
		tipos.Navegacion:    {"procedimiento", "procedimiento,fragmento"},
		tipos.Configuracion: {"procedimiento", "procedimiento,fragmento"},
		tipos.Concepto:      {"concepto", "concepto,fragmento"},
		tipos.Comparacion:   {"concepto", "concepto,fragmento"},
		tipos.Problema:      {"error", "error,procedimiento,fragmento"},
	} {
		for i, q := range quiero {
			got := strings.Join(nombresClases(PlanConsulta(tipo, i+1)), ",")
			if got != q {
				t.Errorf("%s iteración %d: %s, quería %s", tipo, i+1, got, q)
			}
		}
		if n := len(PlanConsulta(tipo, 1)) + len(PlanConsulta(tipo, 2)); n > LimitesDefecto().MaxHerramientas {
			t.Errorf("%s: %d herramientas > MAX_TOOL_CALLS", tipo, n)
		}
	}
}

func TestHiloSinEcoDeLaPregunta(t *testing.T) {
	h := []rag.Turno{{Rol: "usuario", Texto: "¿qué es el kardex?"}, {Rol: "asistente", Texto: "x"}, {Rol: "usuario", Texto: " ¿Cómo  registro un metrado? "}}
	if got := HiloSinEco("¿Cómo registro un metrado?", h); len(got) != 2 {
		t.Fatalf("el eco final se quita: %+v", got)
	}
	if got := HiloSinEco("otra", h); len(got) != 3 {
		t.Fatalf("sin eco no se toca: %+v", got)
	}
	// Primer turno de la página: el hilo es solo la pregunta → hilo vacío, sin «referencia_hilo».
	b := armarAgente()
	res, _ := b.ag.Preguntar(context.Background(), "¿y eso dónde está?", rag.Opciones{Hilo: []rag.Turno{{Rol: "usuario", Texto: "¿y eso dónde está?"}}})
	if len(b.rec.llamadas) > 0 && len(b.rec.llamadas[0].consulta.Aliases) > 0 {
		t.Fatalf("la pregunta no se toma por seguimiento de sí misma: %+v", b.rec.llamadas[0].consulta)
	}
	_ = res
}

func TestTrazaLlevaLosCandidatosParaElBenchmark(t *testing.T) {
	b := armarAgente()
	b.rec.responder = func(n int, c tipos.Consulta, clase string) ([]tipos.Candidato, []string, error) {
		r, d, err := porTema(n, c, clase)
		for i := range r {
			r[i].Meta = map[string]string{"manual": "Manual de Presupuestos"}
			x := 0.7
			r[i].Rerank = &x
		}
		return r, d, err
	}
	_, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	for _, id := range []string{"fusion", "rerank"} {
		cs, ok := etapaV2(tr, id).Datos["candidatos"].([]map[string]any)
		if !ok || len(cs) == 0 || cs[0]["id"] != "metrados.registrar" || cs[0]["clase"] != "procedimiento" ||
			cs[0]["manual"] != "Manual de Presupuestos" || cs[0]["puntaje"] != 0.9 {
			t.Fatalf("%s: candidatos %+v", id, etapaV2(tr, id).Datos["candidatos"])
		}
	}
	// Y sobreviven a la serialización con la forma que lee cmd/evalv2.
	b2, _ := json.Marshal(tr)
	var m struct {
		EtapasV2 []struct {
			ID    string `json:"id"`
			Datos struct {
				Candidatos []struct {
					ID, Clase, Manual string
					Puntaje           float64
				} `json:"candidatos"`
			} `json:"datos"`
		} `json:"etapas_v2"`
	}
	json.Unmarshal(b2, &m)
	for _, e := range m.EtapasV2 {
		if e.ID == "fusion" && (len(e.Datos.Candidatos) == 0 || e.Datos.Candidatos[0].Manual != "Manual de Presupuestos") {
			t.Fatalf("fusion serializada: %+v", e)
		}
	}
}
