package v2

import (
	"fmt"
	"testing"

	"rag-go/internal/rag"
	"rag-go/internal/v2/tipos"
)

func entorno(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestConfigPorDefectoApagada(t *testing.T) {
	c := LeerConfigDe(entorno(nil))
	if c.Version != tipos.V1 || c.Habilitada || c.Porcentaje != 0 || c.Activa() {
		t.Fatalf("la V2 debe estar apagada por defecto: %+v", c)
	}
	if c.Limites != LimitesDefecto() || c.MotorDecision != MotorReglas || c.TipoUmbral != UmbralTipoDefecto {
		t.Fatalf("valores por defecto: %+v", c)
	}
	want := tipos.Limites{MaxPasos: 5, MaxDecisiones: 4, MaxBusquedas: 2, MaxHerramientas: 5, MaxGeneraciones: 1,
		MaxRegeneraciones: 1, TimeoutDecisionMs: 1500, TimeoutGeneracionMs: 5000}
	if c.Limites != want {
		t.Fatalf("límites del megaprompt: %+v", c.Limites)
	}
	if got := c.Elegir("v2", "x"); got != tipos.V1 {
		t.Fatalf("con la V2 apagada, «version» de la petición no cuenta: %s", got)
	}
}

func TestConfigLeeVariables(t *testing.T) {
	c := LeerConfigDe(entorno(map[string]string{
		"AGENT_VERSION": " V2 ", "AGENT_V2_ENABLED": "true", "AGENT_V2_PERCENTAGE": "150",
		"MAX_AGENT_STEPS": "7", "MAX_SEARCH_ITERATIONS": "3", "DECISION_TIMEOUT_MS": "200",
		"DECISION_ENGINE": "Local", "DECISION_MODEL": "jev-style-0.8b-decision-v3", "DECISION_URL": "http://x:1",
		"V2_TIPO_MODELO": "/kb/catalogos/tipo_respuesta.yml", "V2_TIPO_UMBRAL": "0.62", "V2_EVIDENCIA_MIN": "0.1",
	}))
	if c.Version != tipos.V2 || !c.Habilitada || c.Porcentaje != 100 || !c.Activa() {
		t.Fatalf("interruptor: %+v", c)
	}
	if c.Limites.MaxPasos != 7 || c.Limites.MaxBusquedas != 3 || c.Limites.TimeoutDecisionMs != 200 || c.Limites.MaxDecisiones != 4 {
		t.Fatalf("límites: %+v", c.Limites)
	}
	if c.MotorDecision != MotorLocal || c.ModeloDecision != "jev-style-0.8b-decision-v3" || c.URLDecision != "http://x:1" {
		t.Fatalf("motor: %+v", c)
	}
	if c.TipoModelo == "" || c.TipoUmbral != 0.62 || c.EvidenciaMin != 0.1 {
		t.Fatalf("tipo/evidencia: %+v", c)
	}
}

// Búsqueda de fragmentos: híbrida por defecto, reranker apagado (RERANK_URL vacío) y valores inválidos ignorados.
func TestConfigBusquedaYReranker(t *testing.T) {
	c := LeerConfigDe(entorno(nil))
	if c.Busqueda != BusquedaHibrida || c.RerankURL != "" || c.RerankTimeoutMs != RerankTimeoutDefectoMs {
		t.Fatalf("defecto: busqueda %q, rerank %q, timeout %d", c.Busqueda, c.RerankURL, c.RerankTimeoutMs)
	}
	c = LeerConfigDe(entorno(map[string]string{"V2_BUSQUEDA": " Lexica ", "RERANK_URL": " http://127.0.0.1:8091 ", "RERANK_TIMEOUT_MS": "8000"}))
	if c.Busqueda != BusquedaLexica || c.RerankURL != "http://127.0.0.1:8091" || c.RerankTimeoutMs != 8000 {
		t.Fatalf("variables: busqueda %q, rerank %q, timeout %d", c.Busqueda, c.RerankURL, c.RerankTimeoutMs)
	}
	c = LeerConfigDe(entorno(map[string]string{"V2_BUSQUEDA": "vectorial", "RERANK_TIMEOUT_MS": "-1"}))
	if c.Busqueda != BusquedaHibrida || c.RerankTimeoutMs != RerankTimeoutDefectoMs {
		t.Fatalf("inválidos: busqueda %q, timeout %d", c.Busqueda, c.RerankTimeoutMs)
	}
}

func TestConfigValoresInvalidosNoEncienden(t *testing.T) {
	c := LeerConfigDe(entorno(map[string]string{
		"AGENT_VERSION": "v3", "AGENT_V2_ENABLED": "quizas", "AGENT_V2_PERCENTAGE": "-5",
		"MAX_AGENT_STEPS": "muchos", "MAX_TOOL_CALLS": "-1", "DECISION_ENGINE": "openrouter", "V2_TIPO_UMBRAL": "x",
	}))
	if c.Activa() || c.Porcentaje != 0 || c.Limites != LimitesDefecto() || c.MotorDecision != MotorReglas || c.TipoUmbral != UmbralTipoDefecto {
		t.Fatalf("un valor inválido deja el de por defecto: %+v", c)
	}
}

func TestElegirVersion(t *testing.T) {
	hab := ConfigDefecto()
	hab.Habilitada = true
	global := ConfigDefecto()
	global.Version = tipos.V2
	todo := hab
	todo.Porcentaje = 100
	casos := []struct {
		nombre string
		c      Config
		pedida string
		quiero tipos.Version
	}{
		{"habilitada pide v2", hab, "v2", tipos.V2},
		{"habilitada pide V2 con espacios", hab, " V2 ", tipos.V2},
		{"habilitada sin pedir y 0 %", hab, "", tipos.V1},
		{"habilitada pide algo raro", hab, "v9", tipos.V1},
		{"AGENT_VERSION=v2", global, "", tipos.V2},
		{"AGENT_VERSION=v2 sin habilitar ignora la petición v1", global, "v1", tipos.V2},
		{"habilitada + v2 global: la petición v1 manda", func() Config { c := global; c.Habilitada = true; return c }(), "v1", tipos.V1},
		{"100 %", todo, "", tipos.V2},
		{"100 % pero pide v1", todo, "v1", tipos.V1},
		{"porcentaje sin habilitar no cuenta", func() Config { c := ConfigDefecto(); c.Porcentaje = 100; return c }(), "", tipos.V1},
	}
	for _, c := range casos {
		if got := c.c.Elegir(c.pedida, "conv-1"); got != c.quiero {
			t.Errorf("%s: %s, quería %s", c.nombre, got, c.quiero)
		}
	}
}

func TestCubetaEstableYRepartida(t *testing.T) {
	// FNV-1a: el valor no depende del proceso ni de la versión de Go.
	if a, b := Cubeta("¿Cómo registro un metrado?"), Cubeta("  ¿cómo   registro un METRADO? "); a != b {
		t.Fatalf("misma clave normalizada, cubetas distintas: %d %d", a, b)
	}
	if got := Cubeta("conv:abc"); got != Cubeta("conv:abc") || got < 0 || got > 99 {
		t.Fatalf("cubeta %d", got)
	}
	c := ConfigDefecto()
	c.Habilitada, c.Porcentaje = true, 30
	n := 0
	for i := 0; i < 2000; i++ {
		if c.Elegir("", fmt.Sprintf("conversacion-%d", i)) == tipos.V2 {
			n++
		}
	}
	if n < 480 || n > 720 {
		t.Fatalf("30 %% de 2000 debería rondar 600, salió %d", n)
	}
}

func TestClaveTurnoEstableEnLaConversacion(t *testing.T) {
	primera := "¿Cómo registro un metrado?"
	k1 := ClaveTurno(primera, rag.Opciones{})
	k2 := ClaveTurno("¿y luego?", rag.Opciones{Hilo: []rag.Turno{{Rol: "usuario", Texto: primera}, {Rol: "asistente", Texto: "Paso 1"}}})
	if k1 != k2 {
		t.Fatalf("la clave debe ser la misma en toda la conversación: %q %q", k1, k2)
	}
	if k := ClaveTurno("x", rag.Opciones{Conversacion: "abc"}); k != "conv:abc" {
		t.Fatalf("el id de conversación manda: %q", k)
	}
}

// RERANK_TOP_N y RERANK_MAX_RUNAS: 20 × 800 por defecto (BUSQUEDA.md, «Reranker encendido»); 0 vale (= el defecto
// de internal/busqueda y de internal/rerank) y un negativo o un texto se ignoran.
func TestConfigRerankTopNYRunas(t *testing.T) {
	c := LeerConfigDe(entorno(nil))
	if RerankTopNDefecto != 20 || RerankMaxRunasDefecto != 800 || c.RerankTopN != RerankTopNDefecto || c.RerankMaxRunas != RerankMaxRunasDefecto {
		t.Fatalf("defecto: top %d, runas %d", c.RerankTopN, c.RerankMaxRunas)
	}
	c = LeerConfigDe(entorno(map[string]string{"RERANK_TOP_N": " 30 ", "RERANK_MAX_RUNAS": "1200"}))
	if c.RerankTopN != 30 || c.RerankMaxRunas != 1200 {
		t.Fatalf("variables: top %d, runas %d", c.RerankTopN, c.RerankMaxRunas)
	}
	c = LeerConfigDe(entorno(map[string]string{"RERANK_TOP_N": "0", "RERANK_MAX_RUNAS": "0"}))
	if c.RerankTopN != 0 || c.RerankMaxRunas != 0 {
		t.Fatalf("cero: top %d, runas %d", c.RerankTopN, c.RerankMaxRunas)
	}
	c = LeerConfigDe(entorno(map[string]string{"RERANK_TOP_N": "-3", "RERANK_MAX_RUNAS": "mucho"}))
	if c.RerankTopN != RerankTopNDefecto || c.RerankMaxRunas != RerankMaxRunasDefecto {
		t.Fatalf("inválidos: top %d, runas %d", c.RerankTopN, c.RerankMaxRunas)
	}
}
