package v2

// Motores de decisión (tipos.DecisionEngine): eligen entre opciones CERRADAS, nunca escriben párrafos.
//
//   - Reglas (por defecto): toma la propuesta que el código dejó en Estado.Inferencias["regla:<decisión>"]; sin
//     propuesta válida, la primera opción. Nunca falla: es el respaldo de todos los demás.
//   - Local / Remota: un servidor tipo Jev (jeva.cpp o jev-style, POST /v1/systemone) con preguntas «choice».
//     Reusa el cliente de internal/jev. Una elección fuera de las opciones es un error.
//   - Simulada: para pruebas (elecciones fijas, error o demora).
//
// El orquestador llama al motor con el timeout DECISION_TIMEOUT_MS; un error, un timeout o una elección fuera de
// las opciones → decide Reglas, y la traza lo anota («decision», respaldo).

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"rag-go/internal/jev"
	"rag-go/internal/v2/tipos"
)

// Nombres de las decisiones que pide el orquestador.
const (
	DecTipoRespuesta = "response_type"
	DecSuficiencia   = "evidence_sufficiency"
	DecProcedimiento = "procedure"
)

// Opciones de evidence_sufficiency.
const (
	Suficiente   = "suficiente"
	Insuficiente = "insuficiente"
)

// PrefijoRegla: clave de Estado.Inferencias donde el código deja su propuesta para una decisión.
const PrefijoRegla = "regla:"

// ErrFueraDeOpciones: el motor eligió algo que no estaba entre las opciones.
var ErrFueraDeOpciones = errors.New("elección fuera de las opciones")

// ---------------------------------------------------------------------------------------------
// Reglas

// Reglas decide con la propuesta del código. Es determinista y no falla.
type Reglas struct{}

func (Reglas) Nombre() string { return MotorReglas }

func (Reglas) Decidir(_ context.Context, e tipos.Estado, ps []tipos.PreguntaDecision) ([]tipos.Decision, error) {
	out := make([]tipos.Decision, 0, len(ps))
	for _, p := range ps {
		d := tipos.Decision{Nombre: p.Nombre, Fuente: MotorReglas, Confianza: 0.5}
		if prop, ok := e.Inferencias[PrefijoRegla+p.Nombre]; ok && contiene(p.Opciones, prop.Valor) {
			d.Eleccion, d.Confianza = prop.Valor, prop.Confianza
		} else if len(p.Opciones) > 0 {
			d.Eleccion = p.Opciones[0]
		}
		out = append(out, d)
	}
	return out, nil
}

// ---------------------------------------------------------------------------------------------
// Local / Remota (servidor tipo Jev)

// Jev decide con un servidor de la API JEV (/v1/systemone), local o remoto.
type Jev struct {
	Cliente *jev.Cliente
	Fuente  string // local | remota
}

// NuevoJev crea el motor. El timeout del cliente es de respaldo: el orquestador ya corta con DECISION_TIMEOUT_MS.
func NuevoJev(fuente, url, modelo string, timeout time.Duration) *Jev {
	return &Jev{Cliente: jev.Nuevo(url, modelo, timeout), Fuente: fuente}
}

func (j *Jev) Nombre() string { return j.Fuente }

// Decidir envía todas las preguntas en una sola llamada. Cualquier respuesta que falte o elija fuera de las
// opciones invalida la llamada entera (error): el orquestador decide entonces por reglas.
func (j *Jev) Decidir(ctx context.Context, e tipos.Estado, ps []tipos.PreguntaDecision) ([]tipos.Decision, error) {
	if len(ps) == 0 {
		return nil, nil
	}
	preguntas := make(map[string]jev.Pregunta, len(ps))
	ids := make([]string, len(ps))
	for i, p := range ps {
		if len(p.Opciones) == 0 {
			return nil, fmt.Errorf("decisión %s sin opciones", p.Nombre)
		}
		id := p.Nombre
		if _, ya := preguntas[id]; ya {
			id = fmt.Sprintf("%s_%d", p.Nombre, i)
		}
		ids[i] = id
		criterios := make(map[string]any, len(p.Opciones))
		for _, o := range p.Opciones {
			criterios[o] = descripcionOpcion(p.Nombre, o)
		}
		preguntas[id] = jev.Pregunta{Tipo: "choice", Instrucciones: instruccionDecision(p.Nombre), Criterios: criterios}
	}
	rs, err := j.Cliente.Decidir(ctx, estadoParaModelo(e), preguntas)
	if err != nil {
		return nil, err
	}
	out := make([]tipos.Decision, 0, len(ps))
	for i, p := range ps {
		r, ok := rs[ids[i]]
		if !ok {
			return nil, fmt.Errorf("decisión %s: el motor no respondió", p.Nombre)
		}
		if !contiene(p.Opciones, r.Choice) {
			return nil, fmt.Errorf("decisión %s: %q: %w", p.Nombre, r.Choice, ErrFueraDeOpciones)
		}
		out = append(out, tipos.Decision{Nombre: p.Nombre, Eleccion: r.Choice, Confianza: r.Confianza, Fuente: j.Fuente})
	}
	return out, nil
}

// estadoParaModelo: el estado compacto que ve el modelo. Sin las propuestas de las reglas (para que decida por
// sí mismo) y sin la memoria vacía.
func estadoParaModelo(e tipos.Estado) map[string]any {
	inf := map[string]tipos.Dato{}
	for k, v := range e.Inferencias {
		if !strings.HasPrefix(k, PrefijoRegla) {
			inf[k] = v
		}
	}
	m := map[string]any{
		"pregunta":     e.Pregunta,
		"consulta":     e.Consulta,
		"tipo":         e.Tipo,
		"hechos":       e.Hechos,
		"inferencias":  inf,
		"desconocidos": e.Desconocidos,
	}
	if len(e.Hilo) > 0 {
		m["hilo"] = e.Hilo
	}
	if e.Memoria.ProcedimientoID != "" {
		m["memoria"] = e.Memoria
	}
	return m
}

func instruccionDecision(nombre string) string {
	switch nombre {
	case DecTipoRespuesta:
		return "Eres el motor de decisión de un asistente del ERP S10 (construcción). Elige qué FORMA de respuesta pide el usuario con su pregunta. No respondas la pregunta."
	case DecSuficiencia:
		return "¿La evidencia recuperada (hechos de origen «busqueda») basta para responder la pregunta del usuario sin inventar nada?"
	case DecProcedimiento:
		return "¿Cuál de estos procedimientos del manual responde la pregunta del usuario?"
	}
	return "Elige la opción que corresponde al estado."
}

var descripcionesTipo = map[string]string{
	string(tipos.Concepto):      "pregunta qué es, qué significa o para qué sirve un término",
	string(tipos.Procedimiento): "pregunta cómo hacer una operación en varios pasos",
	string(tipos.Navegacion):    "pregunta dónde está algo en la interfaz (menú, pantalla, botón)",
	string(tipos.Problema):      "algo falla o no sale como debería (error, no guarda, no aparece)",
	string(tipos.Configuracion): "pregunta cómo preparar o parametrizar el sistema",
	string(tipos.Comparacion):   "pregunta la diferencia entre dos o más cosas",
	string(tipos.Desconocido):   "no se identifica ninguna tarea concreta: hay que pedir aclaración",
	string(tipos.Social):        "saludo, cortesía o tema ajeno a S10",
}

func descripcionOpcion(decision, opcion string) any {
	switch decision {
	case DecTipoRespuesta:
		if d, ok := descripcionesTipo[opcion]; ok {
			return d
		}
	case DecSuficiencia:
		if opcion == Suficiente {
			return "sí: hay evidencia del tema preguntado"
		}
		return "no: la evidencia es de otro tema o no hay"
	}
	return nil
}

// ---------------------------------------------------------------------------------------------
// Simulada (pruebas)

// Simulada responde con elecciones fijas por decisión (sin elección fija: la primera opción). Con Err falla; con
// Demora espera (respetando el contexto: así se prueba el timeout). Cuenta las llamadas.
type Simulada struct {
	Elecciones map[string]string
	Confianza  float64
	Err        error
	Demora     time.Duration

	mu       sync.Mutex
	llamadas int
}

func (s *Simulada) Nombre() string { return MotorSimulada }

// Llamadas devuelve cuántas veces se llamó a Decidir.
func (s *Simulada) Llamadas() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.llamadas
}

func (s *Simulada) Decidir(ctx context.Context, _ tipos.Estado, ps []tipos.PreguntaDecision) ([]tipos.Decision, error) {
	s.mu.Lock()
	s.llamadas++
	s.mu.Unlock()
	if s.Demora > 0 {
		select {
		case <-time.After(s.Demora):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if s.Err != nil {
		return nil, s.Err
	}
	conf := s.Confianza
	if conf == 0 {
		conf = 0.9
	}
	out := make([]tipos.Decision, 0, len(ps))
	for _, p := range ps {
		e, ok := s.Elecciones[p.Nombre]
		if !ok && len(p.Opciones) > 0 {
			e = p.Opciones[0]
		}
		out = append(out, tipos.Decision{Nombre: p.Nombre, Eleccion: e, Confianza: conf, Fuente: MotorSimulada})
	}
	return out, nil
}

// ---------------------------------------------------------------------------------------------

// NuevoMotor construye el motor de DECISION_ENGINE.
func NuevoMotor(c Config) (tipos.DecisionEngine, error) {
	timeout := time.Duration(max(c.Limites.TimeoutDecisionMs, 1)) * time.Millisecond * 2
	switch c.MotorDecision {
	case "", MotorReglas:
		return Reglas{}, nil
	case MotorLocal:
		url := c.URLDecision
		if url == "" {
			url = URLDecisionLocal
		}
		return NuevoJev(MotorLocal, url, c.ModeloDecision, timeout), nil
	case MotorRemota:
		if c.URLDecision == "" {
			return nil, fmt.Errorf("DECISION_ENGINE=remota necesita DECISION_URL")
		}
		return NuevoJev(MotorRemota, c.URLDecision, c.ModeloDecision, timeout), nil
	case MotorSimulada:
		return &Simulada{}, nil
	}
	return nil, fmt.Errorf("DECISION_ENGINE=%q: usa reglas, local, remota o simulada", c.MotorDecision)
}

func contiene(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
