package conocimiento_test

// Contrato con el núcleo (internal/v2/interfaces.go y internal/v2/tipos) y una vuelta completa por el
// orquestador del núcleo con las piezas de este paquete. Paquete externo para no crear un ciclo de imports.

import (
	"context"
	"strings"
	"testing"

	"rag-go/internal/rag"
	v2 "rag-go/internal/v2"
	"rag-go/internal/v2/conocimiento"
	"rag-go/internal/v2/tipos"
)

var (
	_ v2.Aliaser             = (*conocimiento.Base)(nil)
	_ v2.Catalogo            = (*conocimiento.Base)(nil)
	_ v2.Constructor         = (*conocimiento.Constructor)(nil)
	_ v2.Continuador         = (*conocimiento.Constructor)(nil)
	_ tipos.Recuperador      = (*conocimiento.Recuperador)(nil)
	_ tipos.GenerationEngine = (*conocimiento.MotorPlantillas)(nil)
	_ tipos.GenerationEngine = (*conocimiento.MotorLocal)(nil)
	_ tipos.QualityGate      = (*conocimiento.Gate)(nil)
)

func TestIntegracion_AgenteDelNucleo(t *testing.T) {
	b := conocimiento.Cargar(conocimiento.Opciones{DirKB: "testdata/kb", ArchivoPlantillas: "testdata/plantillas/respuestas.yml"})
	r := conocimiento.NuevoRecuperador(b)
	c := conocimiento.NuevoConstructor(b, r)
	a := v2.Nuevo(v2.Config{Version: tipos.V2, Limites: v2.LimitesDefecto()})
	a.Aliaser, a.Recuperador, a.Constructor, a.Catalogo = b, r, c, b
	a.Plantillas, a.Gate = conocimiento.NuevoMotorPlantillas(b), conocimiento.NuevoGate(b)
	ctx := context.Background()

	res, err := a.Preguntar(ctx, "¿Cómo registro un metrado en S10?", rag.Opciones{})
	if err != nil || res.PlanV2 == nil {
		t.Fatalf("sin plan: %v %+v", err, res)
	}
	t.Logf("procedimental → %s %v; %q", res.PlanV2.Tipo, res.PlanV2.Procedimiento, res.Respuesta)
	if res.PlanV2.Tipo == tipos.Procedimiento {
		if res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "presupuestos.registrar-metrado" || len(res.PlanV2.Pasos) != 3 {
			t.Errorf("plan procedimental: %+v", res.PlanV2)
		}
		if !strings.Contains(res.Respuesta, "Hoja del presupuesto") {
			t.Errorf("respuesta: %q", res.Respuesta)
		}
	}

	res, _ = a.Preguntar(ctx, "¿qué es un metrado?", rag.Opciones{})
	t.Logf("conceptual → %s; %q", res.PlanV2.Tipo, res.Respuesta)
	if res.PlanV2.Tipo == tipos.Concepto && (res.PlanV2.Concepto == nil || res.PlanV2.Concepto.ID != "metrado") {
		t.Errorf("plan conceptual: %+v", res.PlanV2)
	}

	// «¿y luego?» con el procedimiento en curso: el núcleo pide la parte siguiente al Continuador.
	mem := tipos.Memoria{ProcedimientoID: "presupuestos.modificar-partida", PasoActual: 6}
	res, _ = a.Preguntar(ctx, "¿y luego?", rag.Opciones{Memoria: &mem})
	t.Logf("continuación → %v; %q", res.PlanV2.PasosMostrados, res.Respuesta)
	if res.PlanV2 == nil || len(res.PlanV2.PasosMostrados) == 0 || res.PlanV2.PasosMostrados[0] != 7 {
		t.Errorf("continuación: %+v", res.PlanV2)
	}
}

// El núcleo (elegirProcedimiento) y el constructor (ambiguo) miden el empate de dos candidatos del reranker con el
// mismo margen de logits.
func TestContrato_MargenEmpateRerank(t *testing.T) {
	if v2.MargenProcedimientoRerank != conocimiento.MargenRerankEmpate {
		t.Fatalf("núcleo %v ≠ constructor %v", v2.MargenProcedimientoRerank, conocimiento.MargenRerankEmpate)
	}
}
