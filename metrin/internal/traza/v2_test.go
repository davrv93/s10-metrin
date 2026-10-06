package traza

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTurnoV1SinRastroDeV2(t *testing.T) {
	r := Nuevo("hola")
	r.Fin(EtapaEntrada, EstadoOK, Datos{"pregunta": "hola", "hilo_turnos": 0})
	tr := r.Cerrar()
	b, _ := json.Marshal(tr)
	if strings.Contains(string(b), `"agente"`) || strings.Contains(string(b), `"etapas_v2"`) || tr.EtapasV2 != nil {
		t.Fatalf("un turno V1 no lleva nada de la V2: %s", b)
	}
	if len(tr.Etapas) != 18 {
		t.Fatalf("las 18 etapas de V1 no cambian: %d", len(tr.Etapas))
	}
}

func TestEtapasV2EnOrdenYConMinimos(t *testing.T) {
	ids := IDsV2()
	quiero := []string{"tipo_respuesta", "decision", "plan_consulta", "busqueda_lexica", "busqueda_vectorial", "fusion",
		"rerank", "procedimiento", "pasos", "fotos_paso", "plan", "quality_gate", "renderizado"}
	if strings.Join(ids, ",") != strings.Join(quiero, ",") {
		t.Fatalf("catálogo V2: %v", ids)
	}
	for _, id := range ids {
		if _, choca := indice[id]; choca {
			t.Fatalf("la etapa V2 %q choca con una de V1", id)
		}
	}
	r := Nuevo("¿y luego?")
	v := r.V2()
	if v != r.V2() {
		t.Fatal("V2() devuelve siempre el mismo sub-registro")
	}
	v.Inicio(EtapaV2Plan)
	time.Sleep(time.Millisecond)
	v.Fin(EtapaV2Plan, EstadoOK, Datos{"tipo": "PROCEDURE", "plantilla": "PASO", "sin_evidencia": false, "texto": strings.Repeat("x", 300)})
	v.NoTomada(EtapaV2Fusion, "memoria: sin búsqueda")
	v.Inicio(EtapaV2Renderizado) // empezó y no terminó
	tr := r.Cerrar()
	if tr.Agente != AgenteV2 || len(tr.EtapasV2) != len(ids) {
		t.Fatalf("traza V2: %+v", tr)
	}
	for i, e := range tr.EtapasV2 {
		if e.ID != ids[i] || !EstadoValido(e.Estado) {
			t.Fatalf("etapa %d: %+v", i, e)
		}
		for _, k := range MinimosV2(e.ID) {
			if _, ok := e.Datos[k]; !ok {
				t.Errorf("%s sin %q", e.ID, k)
			}
		}
		if e.Estado == EstadoOmitida && e.Razon == "" {
			t.Errorf("%s omitida sin razón", e.ID)
		}
	}
	plan := tr.EtapasV2[indiceV2[EtapaV2Plan]]
	if plan.Ms == nil || plan.Estado != EstadoOK || len([]rune(plan.Datos["texto"].(string))) > MaxCaracteres {
		t.Fatalf("plan: %+v", plan)
	}
	if e := tr.EtapasV2[indiceV2[EtapaV2Fusion]]; e.Estado != EstadoNoTomada || e.Ms != nil {
		t.Fatalf("fusion: %+v", e)
	}
	if e := tr.EtapasV2[indiceV2[EtapaV2Renderizado]]; e.Estado != EstadoError {
		t.Fatalf("una etapa iniciada y sin cerrar es error: %+v", e)
	}
	if e := tr.EtapasV2[indiceV2[EtapaV2Decision]]; e.Estado != EstadoOmitida {
		t.Fatalf("lo pendiente queda omitido: %+v", e)
	}
}

func TestOportunidadesV2TrasLasDeV1(t *testing.T) {
	r := Nuevo("x")
	r.Fin(EtapaRuta, EstadoOK, Datos{"ruta": "rag"})
	r.Fin(EtapaGeneracion, EstadoError, nil) // oportunidad V1
	v := r.V2()
	v.Fin(EtapaV2TipoRespuesta, EstadoOK, Datos{"dudoso": true, "confianza": 0.3})
	v.Fin(EtapaV2Decision, EstadoRespaldo, Datos{"respaldo": true})
	v.Fin(EtapaV2Rerank, EstadoRespaldo, nil)
	v.Fin(EtapaV2QualityGate, EstadoError, Datos{"passed": false})
	v.Fin(EtapaV2Plan, EstadoOK, Datos{"sin_evidencia": true})
	tr := r.Cerrar()
	if len(tr.Oportunidades) != 6 || tr.Oportunidades[0].Etapa != EtapaGeneracion {
		t.Fatalf("oportunidades: %+v", tr.Oportunidades)
	}
	for i, o := range tr.Oportunidades {
		if o.N != i+1 {
			t.Fatalf("numeración continua: %+v", tr.Oportunidades)
		}
	}
	if e := tr.EtapasV2[indiceV2[EtapaV2TipoRespuesta]]; e.Estado != EstadoAlerta {
		t.Fatalf("tipo dudoso sube a alerta: %+v", e)
	}
	if e := tr.EtapasV2[indiceV2[EtapaV2Rerank]]; e.Estado != EstadoRespaldo {
		t.Fatalf("respaldo se queda en respaldo: %+v", e)
	}
}

func TestRecorderV2NilEsNoOp(t *testing.T) {
	var r *Recorder
	v := r.V2()
	if v != nil {
		t.Fatal("sin recorder no hay sub-registro")
	}
	v.Inicio(EtapaV2Plan)
	v.Fin(EtapaV2Plan, EstadoOK, Datos{"x": 1})
	v.FinDuracion(EtapaV2Plan, EstadoOK, time.Second, nil)
	v.Omitir(EtapaV2Plan, "x")
	v.NoTomada(EtapaV2Plan, "x")
	v.Dato(EtapaV2Plan, "x", 1)
	v.Razon(EtapaV2Plan, "x")
	v.Modelo(EtapaV2Plan, "x")
	v.CambiarEstado(EtapaV2Plan, EstadoOK, "x")
	if v.Pendiente(EtapaV2Plan) || v.Estado(EtapaV2Plan) != "" {
		t.Fatal("nil no tiene estado")
	}
	if _, ok := v.DatoDe(EtapaV2Plan, "x"); ok {
		t.Fatal("nil no tiene datos")
	}
	// Un id desconocido tampoco hace nada.
	Nuevo("x").V2().Fin("no_existe", EstadoOK, nil)
}
