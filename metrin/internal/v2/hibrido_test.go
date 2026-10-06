package v2

import (
	"context"
	"strings"
	"testing"

	"rag-go/internal/rag"
	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

// recHibrido: el recuperador de prueba que, como internal/v2/conocimiento con la búsqueda híbrida, deja sus
// cifras en «hibrido» de las cuatro etapas de búsqueda sin cerrarlas.
type recHibrido struct {
	*recuperadorFalso
	datos map[string]traza.Datos
}

func (r recHibrido) Buscar(ctx context.Context, c tipos.Consulta, clase string, k int) ([]tipos.Candidato, []string, error) {
	v := traza.De(ctx).V2()
	for id, d := range r.datos {
		v.Dato(id, claveHibrido, d)
	}
	return r.recuperadorFalso.Buscar(ctx, c, clase, k)
}

func datosHibrido(vecFallo string, reordenado bool) map[string]traza.Datos {
	dv := traza.Datos{"activo": true, "resultados": 12, "ms": 9.5, "fallo": nil}
	if vecFallo != "" {
		dv["fallo"] = vecFallo
	}
	return map[string]traza.Datos{
		traza.EtapaV2BusquedaLexica:    {"resultados": 50, "ms": 0.6},
		traza.EtapaV2BusquedaVectorial: dv,
		traza.EtapaV2Fusion:            {"resultados": 50, "llamadas": 1, "candidatos": []map[string]any{{"id": "f1", "clase": "fragmento"}}},
		traza.EtapaV2Rerank:            {"activo": reordenado, "reordenado": reordenado, "fallo": nil},
	}
}

// Los procedimientos se buscan solo por léxico («sin_vector»), pero si la búsqueda híbrida de fragmentos corrió con
// vector, la etapa vectorial no es «solo léxica»; sin reranker configurado, el reranker queda en respaldo y lo dice.
func TestTrazaIntegraLaBusquedaHibrida(t *testing.T) {
	b := armarAgente()
	b.rec.responder = func(n int, c tipos.Consulta, clase string) ([]tipos.Candidato, []string, error) {
		r, _, err := porTema(n, c, clase)
		return r, []string{"sin_vector: búsqueda léxica", "sin_reranker: puntaje híbrido sin reordenar"}, err
	}
	b.ag.Recuperador = recHibrido{b.rec, datosHibrido("", false)}
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	comprobarV2(t, res)
	vec := etapaV2(tr, traza.EtapaV2BusquedaVectorial)
	if vec.Estado != traza.EstadoOK || oportunidadEn(tr, traza.EtapaV2BusquedaVectorial) || !strings.Contains(vec.Razon, "fragmentos") {
		t.Fatalf("vectorial: %+v", vec)
	}
	if h, _ := vec.Datos[claveHibrido].(map[string]any); h == nil || h["resultados"] != 12 {
		t.Fatalf("las cifras de la híbrida van en la etapa: %+v", vec.Datos)
	}
	rr := etapaV2(tr, traza.EtapaV2Rerank)
	if rr.Estado != traza.EstadoRespaldo || !strings.Contains(rr.Razon, "RERANK_URL") {
		t.Fatalf("rerank: %+v", rr)
	}
	if fu := etapaV2(tr, traza.EtapaV2Fusion); fu.Datos[claveHibrido] == nil || fu.Datos["candidatos"] == nil {
		t.Fatalf("la fusión conserva los candidatos del núcleo y suma los de la híbrida: %+v", fu.Datos)
	}

	// Vector de fragmentos caído: respaldo, con su causa; reranker que reordenó: activo.
	b = armarAgente()
	b.ag.Recuperador = recHibrido{b.rec, datosHibrido("chromem: colección vacía", true)}
	_, tr = preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	if vec := etapaV2(tr, traza.EtapaV2BusquedaVectorial); vec.Estado != traza.EstadoRespaldo || !strings.Contains(vec.Razon, "chromem") {
		t.Fatalf("vector caído: %+v", vec)
	}
	if rr := etapaV2(tr, traza.EtapaV2Rerank); rr.Estado != traza.EstadoOK || rr.Datos["activo"] != true {
		t.Fatalf("reranker activo: %+v", rr)
	}
}
