// Package v2 es el núcleo de la V2 de Metrín (RAG procedural multimodal): interruptor, contexto estructurado,
// tipo de respuesta, motor de decisión, presupuesto, memoria del procedimiento y el orquestador con recursividad
// controlada. Especificación: docs/V2-RAG-PROCEDURAL.md. El conocimiento procedural (recuperador, constructor del
// plan, plantillas y quality gate) vive en internal/v2/conocimiento y entra por las interfaces de interfaces.go y
// de internal/v2/tipos.
//
// V1 no cambia: la V2 está apagada por defecto y, apagada, /ask responde byte a byte igual que antes.
package v2

import (
	"hash/fnv"
	"os"
	"strconv"
	"strings"

	"rag-go/internal/rag"
	"rag-go/internal/v2/tipos"
)

// Motores de decisión (DECISION_ENGINE).
const (
	MotorReglas   = "reglas"
	MotorLocal    = "local"
	MotorRemota   = "remota"
	MotorSimulada = "simulada"
)

// URLDecisionLocal: servidor tipo Jev de la Mac (jev-style serve, POST /v1/systemone). Solo se usa con
// DECISION_ENGINE=local y DECISION_URL vacío.
const URLDecisionLocal = "http://127.0.0.1:8765"

// UmbralTipoDefecto y MargenTipoDefecto: el kNN de tipo de respuesta decide su top-1 sin el motor de decisión solo
// si su similitud llega al umbral y le saca el margen al top-2. Las reglas léxicas claras no pasan por aquí.
//
// PROVISIONALES (06-10-2026), elegidos EN MUESTRA sobre kb/catalogos/tipo_respuesta_prueba.yml (338 frases) con el
// embebedor estático y modelos/clasificador-tipo-respuesta-candidato.json (TestMedicionTipoConCatalogoReal): las
// reglas solas deciden el 66,6 % con 80,4 % de precisión; con umbral 0,40 y margen 0,08 el kNN suma 23 frases
// (6,8 %) con ~78 % de precisión (aciertos − 2 × errores = +8, el mejor de la rejilla 0,40–0,59 × 0,02–0,08). Con
// margen 0,02 el saldo es negativo. kb/catalogos/EVALUACION.md: con este embebedor el kNN es una señal, no un
// decisor; lo dudoso va al motor de decisión y, si falla, a UNKNOWN (aclaración).
const (
	UmbralTipoDefecto = 0.40
	MargenTipoDefecto = 0.08
)

// EvidenciaMinDefecto: puntaje mínimo del mejor candidato para dar la evidencia por suficiente sin consultar al
// motor de decisión. PROVISIONAL: 0 = basta un candidato de la clase pedida, hasta calibrar la escala del
// recuperador híbrido (metrin/eval/BUSQUEDA.md).
const EvidenciaMinDefecto = 0.0

// Búsqueda de fragmentos de la V2 (V2_BUSQUEDA).
const (
	BusquedaHibrida = "hibrida" // internal/busqueda: BM25F + vector del almacén + RRF sobre kb/fragmentos*.jsonl
	BusquedaLexica  = "lexica"  // solo el BM25 propio de internal/v2/conocimiento (lo de antes)
)

// RerankTimeoutDefectoMs: tiempo máximo del reranker. 3 s es el TimeoutRerank de busqueda.OpcionesPorDefecto (p95
// medido con Metal: 2,3–2,6 s, metrin/eval/BUSQUEDA.md §6). En CPU (Docker) está SIN MEDIR.
const RerankTimeoutDefectoMs = 3000

// RerankTopNDefecto y RerankMaxRunasDefecto: candidatos que pasan por el reranker y runas de cada uno (RERANK_TOP_N,
// RERANK_MAX_RUNAS). Medido el 06-10-2026 (metrin/eval/BUSQUEDA.md, «Reranker encendido»): 20 × 800 conserva MRR@10
// 0,910 (0,929 con 30 × 1 200, 0,780 sin reranker) con la mitad de latencia (p50 0,8–1,0 s frente a 1,6 s en la Mac).
// 0 = el valor de internal/busqueda (30) o de internal/rerank (1 200).
const (
	RerankTopNDefecto     = 20
	RerankMaxRunasDefecto = 800
)

// Clases que pasan por el reranker (V2_RERANK_CLASES, separadas por comas). «fragmento»: la búsqueda híbrida de
// fragmentos (internal/busqueda reordena sus primeros resultados); «procedimiento» y «concepto»: el BM25 propio de
// internal/v2/conocimiento, para ELEGIR el procedimiento o el concepto (rerank_clases.go). «ninguna» = ninguna. Solo
// actúan con RERANK_URL. Medido el 06-10-2026 (metrin/eval/BUSQUEDA.md §12).
const (
	RerankFragmento     = "fragmento"
	RerankProcedimiento = "procedimiento"
	RerankConcepto      = "concepto"
	RerankNinguna       = "ninguna"
)

// RerankClasesDefecto: las tres. Medido el 06-10-2026 con el benchmark de la V2 (194 casos; BUSQUEDA.md §12): PAS
// 81,8 % frente a 60,0 % con el reranker solo en fragmentos o sin reranker, invención 0 % en todas; a cambio, p50 de
// ~0,9 s (la Mac, Metal) y una abstención correcta menos (11 de 12).
var RerankClasesDefecto = []string{RerankFragmento, RerankProcedimiento, RerankConcepto}

// Config de la V2. Se lee del entorno (LeerConfig); los valores por defecto dejan la V2 apagada.
type Config struct {
	Version    tipos.Version // AGENT_VERSION: v1 (defecto) | v2 → todos los turnos en V2
	Habilitada bool          // AGENT_V2_ENABLED (false): habilita el porcentaje y el campo «version» de la petición
	Porcentaje int           // AGENT_V2_PERCENTAGE (0–100): turnos en V2 por hash estable de la conversación

	Limites tipos.Limites

	MotorDecision  string // DECISION_ENGINE: reglas (defecto) | local | remota | simulada
	ModeloDecision string // DECISION_MODEL: campo «model» de la petición (vacío = el único del servidor)
	URLDecision    string // DECISION_URL: local → por defecto URLDecisionLocal; remota → obligatoria

	TipoModelo   string  // V2_TIPO_MODELO: modelo kNN de tipo de respuesta (.json de clasificar o catálogo .yml)
	Router       string  // V2_ROUTER: cabeza del router de casos (router-s10.cabeza.json); vacío = sin router (router.go)
	TipoUmbral   float64 // V2_TIPO_UMBRAL: provisional, ver UmbralTipoDefecto
	TipoMargen   float64 // V2_TIPO_MARGEN: provisional, ver MargenTipoDefecto
	EvidenciaMin float64 // V2_EVIDENCIA_MIN: provisional, ver EvidenciaMinDefecto

	Busqueda        string // V2_BUSQUEDA: hibrida (defecto) | lexica — solo la clase «fragmento»
	RerankURL       string // RERANK_URL: llama-server --reranking; vacío (defecto) = sin reranker
	RerankTimeoutMs int    // RERANK_TIMEOUT_MS: ver RerankTimeoutDefectoMs
	RerankTopN      int    // RERANK_TOP_N: candidatos de la fusión que pasan por el reranker (RerankTopNDefecto)
	RerankMaxRunas  int    // RERANK_MAX_RUNAS: runas de cada candidato enviadas al reranker (RerankMaxRunasDefecto)
	// V2_RERANK_CLASES: clases que pasan por el reranker (RerankClasesDefecto). Nombres desconocidos se ignoran; si
	// no queda ninguno válido, el defecto. «ninguna» = el reranker no reordena nada aunque haya RERANK_URL.
	RerankClases []string
}

// Reordena dice si la clase pasa por el reranker: hay RERANK_URL y la clase está en V2_RERANK_CLASES.
func (c Config) Reordena(clase string) bool {
	if c.RerankURL == "" {
		return false
	}
	for _, x := range c.RerankClases {
		if x == clase {
			return true
		}
	}
	return false
}

// LimitesDefecto son los del megaprompt.
func LimitesDefecto() tipos.Limites {
	return tipos.Limites{
		MaxPasos: 5, MaxDecisiones: 4, MaxBusquedas: 2, MaxHerramientas: 5,
		MaxGeneraciones: 1, MaxRegeneraciones: 1,
		TimeoutDecisionMs: 1500, TimeoutGeneracionMs: 5000,
	}
}

// ConfigDefecto: V2 apagada, motor de reglas y los límites del megaprompt.
func ConfigDefecto() Config {
	return Config{
		Version:       tipos.V1,
		Limites:       LimitesDefecto(),
		MotorDecision: MotorReglas,
		TipoUmbral:    UmbralTipoDefecto,
		TipoMargen:    MargenTipoDefecto,
		EvidenciaMin:  EvidenciaMinDefecto,

		Busqueda:        BusquedaHibrida,
		RerankTimeoutMs: RerankTimeoutDefectoMs,
		RerankTopN:      RerankTopNDefecto,
		RerankMaxRunas:  RerankMaxRunasDefecto,
		RerankClases:    append([]string(nil), RerankClasesDefecto...),
	}
}

// leerClasesRerank: «fragmento,procedimiento,concepto» (en cualquier orden, sin repetir). Vacío o sin ningún nombre
// válido → nil (se queda el defecto); «ninguna» → lista vacía.
func leerClasesRerank(v string) []string {
	var out []string
	vis := map[string]bool{}
	for _, x := range strings.Split(v, ",") {
		x = strings.ToLower(strings.TrimSpace(x))
		switch x {
		case RerankNinguna:
			return []string{}
		case RerankFragmento, RerankProcedimiento, RerankConcepto:
			if !vis[x] {
				vis[x] = true
				out = append(out, x)
			}
		}
	}
	return out
}

// LeerConfig lee la configuración de la V2 del entorno.
func LeerConfig() Config { return LeerConfigDe(os.Getenv) }

// LeerConfigDe lee la configuración con una función de entorno (las pruebas pasan un mapa). Un valor inválido
// deja el de por defecto: un error de tecleo nunca enciende la V2.
func LeerConfigDe(env func(string) string) Config {
	c := ConfigDefecto()
	if strings.EqualFold(strings.TrimSpace(env("AGENT_VERSION")), string(tipos.V2)) {
		c.Version = tipos.V2
	}
	c.Habilitada = verdadero(env("AGENT_V2_ENABLED"))
	if p, err := strconv.Atoi(strings.TrimSpace(env("AGENT_V2_PERCENTAGE"))); err == nil {
		c.Porcentaje = min(max(p, 0), 100)
	}
	l := &c.Limites
	entero(env, "MAX_AGENT_STEPS", &l.MaxPasos)
	entero(env, "MAX_DECISION_CALLS", &l.MaxDecisiones)
	entero(env, "MAX_SEARCH_ITERATIONS", &l.MaxBusquedas)
	entero(env, "MAX_TOOL_CALLS", &l.MaxHerramientas)
	entero(env, "MAX_GENERATIONS", &l.MaxGeneraciones)
	entero(env, "MAX_REGENERATIONS", &l.MaxRegeneraciones)
	entero(env, "DECISION_TIMEOUT_MS", &l.TimeoutDecisionMs)
	entero(env, "GENERATION_TIMEOUT_MS", &l.TimeoutGeneracionMs)
	switch m := strings.ToLower(strings.TrimSpace(env("DECISION_ENGINE"))); m {
	case MotorReglas, MotorLocal, MotorRemota, MotorSimulada:
		c.MotorDecision = m
	}
	c.ModeloDecision = strings.TrimSpace(env("DECISION_MODEL"))
	c.URLDecision = strings.TrimSpace(env("DECISION_URL"))
	c.TipoModelo = strings.TrimSpace(env("V2_TIPO_MODELO"))
	c.Router = strings.TrimSpace(env("V2_ROUTER"))
	decimal(env, "V2_TIPO_UMBRAL", &c.TipoUmbral)
	decimal(env, "V2_TIPO_MARGEN", &c.TipoMargen)
	decimal(env, "V2_EVIDENCIA_MIN", &c.EvidenciaMin)
	switch b := strings.ToLower(strings.TrimSpace(env("V2_BUSQUEDA"))); b {
	case BusquedaHibrida, BusquedaLexica:
		c.Busqueda = b
	}
	c.RerankURL = strings.TrimSpace(env("RERANK_URL"))
	entero(env, "RERANK_TIMEOUT_MS", &c.RerankTimeoutMs)
	entero(env, "RERANK_TOP_N", &c.RerankTopN)
	entero(env, "RERANK_MAX_RUNAS", &c.RerankMaxRunas)
	if cs := leerClasesRerank(env("V2_RERANK_CLASES")); cs != nil {
		c.RerankClases = cs
	}
	return c
}

// Activa dice si hay que construir el orquestador V2: AGENT_V2_ENABLED=true o AGENT_VERSION=v2.
func (c Config) Activa() bool { return c.Habilitada || c.Version == tipos.V2 }

// Elegir decide la versión de un turno:
//  1. con AGENT_V2_ENABLED=true, la «version» de la petición manda («v1» o «v2»; el chat de prueba);
//  2. AGENT_VERSION=v2 → v2;
//  3. con AGENT_V2_ENABLED=true, el AGENT_V2_PERCENTAGE % de las conversaciones (hash estable de la clave);
//  4. si no, v1.
func (c Config) Elegir(pedida, clave string) tipos.Version {
	if c.Habilitada {
		switch strings.ToLower(strings.TrimSpace(pedida)) {
		case string(tipos.V1):
			return tipos.V1
		case string(tipos.V2):
			return tipos.V2
		}
	}
	if c.Version == tipos.V2 {
		return tipos.V2
	}
	if c.Habilitada && c.Porcentaje > 0 && Cubeta(clave) < c.Porcentaje {
		return tipos.V2
	}
	return tipos.V1
}

// Cubeta reparte una clave en 0–99 con FNV-1a: la misma clave cae siempre en la misma cubeta (en cualquier
// proceso y versión de Go), así que una conversación no salta de versión entre turnos.
func Cubeta(clave string) int {
	h := fnv.New32a()
	h.Write([]byte(strings.ToLower(strings.Join(strings.Fields(clave), " "))))
	return int(h.Sum32() % 100)
}

// ClaveTurno es la clave del porcentaje: el id de la conversación si la petición lo trae; si no, el primer
// mensaje del usuario en el hilo (estable durante toda la conversación); si no hay hilo, la pregunta, que es
// justo el primer mensaje de los turnos que vendrán.
func ClaveTurno(pregunta string, o rag.Opciones) string {
	if c := strings.TrimSpace(o.Conversacion); c != "" {
		return "conv:" + c
	}
	for _, t := range o.Hilo {
		if t.Rol == "usuario" && strings.TrimSpace(t.Texto) != "" {
			return t.Texto
		}
	}
	return pregunta
}

func verdadero(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "si", "sí", "yes", "on":
		return true
	}
	return false
}

func entero(env func(string) string, k string, dst *int) {
	if n, err := strconv.Atoi(strings.TrimSpace(env(k))); err == nil && n >= 0 {
		*dst = n
	}
}

func decimal(env func(string) string, k string, dst *float64) {
	if x, err := strconv.ParseFloat(strings.TrimSpace(env(k)), 64); err == nil && x >= 0 {
		*dst = x
	}
}
