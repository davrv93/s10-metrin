package main

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"rag-go/internal/v2/tipos"
)

// ---------------------------------------------------------------------------------------------
// Base sintética: tres fragmentos con texto conocido y un procedimiento de tres pasos.

const (
	rutaF1 = "imagenes/manual-de-presupuestos/aaa.png"
	rutaF2 = "imagenes/manual-de-presupuestos/bbb.png"
	rutaF3 = "imagenes/manual-de-presupuestos/ccc.png"
)

func baseSintetica(t *testing.T) *Base {
	t.Helper()
	b := nuevaBase("")
	b.agregarFragmento(&Fragmento{ID: "web-s045", Manual: "Manual de Presupuestos", Titulo: "Manual de Presupuestos › 3.5 Hoja",
		Cita: "Manual de Presupuestos › 3.5 Hoja",
		Texto: "Ubique el cursor en la partida dentro de la hoja del presupuesto. Ingrese el metrado en la celda Metrado. " +
			"Use el botón Grabar para guardar los metrados de la partida.",
		Fotos: []string{rutaF1, rutaF2}})
	b.agregarFragmento(&Fragmento{ID: "web-s045", Manual: "Manual de Compras", Titulo: "Manual de Compras › 3.5 Pedidos",
		Cita: "Manual de Compras › 3.5 Pedidos", Texto: "Registre el pedido de compra con sus insumos."})
	b.agregarFragmento(&Fragmento{ID: "img-ccc-0000", Manual: "Manual de Presupuestos", Titulo: "ccc.png", Cita: "ccc.png",
		Texto: "Imagen del manual S10: Hoja del presupuesto, celda de metrados, procesar presupuesto", Imagen: rutaF3,
		Fotos: []string{rutaF3}})
	b.agregarFragmento(&Fragmento{ID: "web-s099", Manual: "Manual de Presupuestos", Titulo: "Manual de Presupuestos › 5.1 Procesar",
		Cita: "Manual de Presupuestos › 5.1 Procesar", Texto: "Use la opción Procesar para recalcular el presupuesto con los metrados."})
	p := &Procedimiento{
		ID: "presupuestos.ingresar-metrados", Modulo: "presupuestos", Titulo: "Ingresar metrados",
		Pasos: []PasoYAML{
			{N: 1, ID: "presupuestos.ingresar-metrados#1", Accion: "Ubique el cursor en la partida de la **hoja del presupuesto**.",
				Fuente: []string{"web-s045"}, Fotos: []FotoYAML{{ID: claveFoto(rutaF1, ""), Ruta: rutaF1}}},
			{N: 2, ID: "presupuestos.ingresar-metrados#2", Accion: "Ingrese la cantidad en la celda **Metrado** de la partida.",
				Fuente: []string{"web-s045"}, Fotos: []FotoYAML{{ID: claveFoto(rutaF2, ""), Ruta: rutaF2}}},
			{N: 3, ID: "presupuestos.ingresar-metrados#3", Accion: "Elija **Procesar** para recalcular el presupuesto con los metrados.",
				Fuente: []string{"web-s099", "img-ccc-0000"}, Fotos: []FotoYAML{{ID: claveFoto(rutaF3, ""), Ruta: rutaF3}}},
		},
		Errores: []ErrorYAML{{Sintoma: "No se guarda el metrado.", Solucion: "Use el botón Grabar.", Fuente: []string{"web-s045"}}},
		Fuentes: []FuenteYAML{
			{ID: "web-s045", Manual: "Manual de Presupuestos"},
			{ID: "web-s099", Manual: "Manual de Presupuestos"},
			{ID: "img-ccc-0000", Manual: "Manual de Presupuestos"},
		},
	}
	b.Procedimientos[p.ID] = p
	return b
}

func casoSintetico() Caso {
	proc := "presupuestos.ingresar-metrados"
	return Caso{
		ID: "t-proc", Categoria: "PROCEDURE", Turnos: []string{"como registro metrado"}, TipoEsperado: "PROCEDURE",
		ProcedimientoEsperado: &proc,
		FotosEsperadas: map[string][]string{
			"1": {claveFoto(rutaF1, "")}, "2": {claveFoto(rutaF2, "")}, "3": {claveFoto(rutaF3, "")},
		},
		FragmentosRelevantes: []RefFragmento{{ID: "web-s045", Manual: "Manual de Presupuestos"}, {ID: "web-s099", Manual: "Manual de Presupuestos"}},
		Sintetico:            true,
	}
}

// planPerfecto: una respuesta V2 que cumple las siete condiciones.
func planPerfecto() *tipos.Plan {
	id := "presupuestos.ingresar-metrados"
	return &tipos.Plan{
		Version: 2, Tipo: tipos.Procedimiento,
		Procedimiento: &tipos.Ref{ID: id, Titulo: "Ingresar metrados", Confianza: 0.9},
		Pasos: []tipos.PasoPlan{
			{N: 1, ID: id + "#1", Texto: "Ubique el cursor en la partida de la **hoja del presupuesto**.",
				Fotos: []tipos.Foto{{ID: claveFoto(rutaF1, ""), Ruta: rutaF1}}, Fuente: []string{"web-s045"}},
			{N: 2, ID: id + "#2", Texto: "Ingrese la cantidad en la celda **Metrado** de la partida.",
				Fotos: []tipos.Foto{{ID: claveFoto(rutaF2, ""), Ruta: rutaF2}}, Fuente: []string{"web-s045"}},
			{N: 3, ID: id + "#3", Texto: "Elija **Procesar** para recalcular el presupuesto con los metrados.",
				Fotos: []tipos.Foto{{ID: claveFoto(rutaF3, ""), Ruta: rutaF3}}, Fuente: []string{"web-s099", "img-ccc-0000"}},
		},
		Fuentes: []tipos.FuentePlan{{Manual: "Manual de Presupuestos"}},
	}
}

func turnoV2(p *tipos.Plan, respuesta string) TurnoCrudo {
	r := &respuestaAsk{Respuesta: respuesta, Modo: "v2", Plan: p}
	return TurnoCrudo{Status: 200, Respuesta: r, MsCliente: 10}
}

func turnoV1(respuesta, tipoConsulta string, fuentes ...fuenteAsk) TurnoCrudo {
	r := &respuestaAsk{Respuesta: respuesta, Modo: "tutorial", Fuentes: fuentes}
	r.Orquestacion.TipoConsulta = tipoConsulta
	r.Orquestacion.Ruta = "rag"
	return TurnoCrudo{Status: 200, Respuesta: r, MsCliente: 10}
}

func evaluarV2(t *testing.T, b *Base, c Caso, p *tipos.Plan) EvalCaso {
	t.Helper()
	o := observar(b, "v2", turnoV2(p, "texto renderizado"))
	return evaluarCaso(b, c, []Observacion{o})
}

func condiciones(e EvalCaso) [7]bool {
	var out [7]bool
	copy(out[:], e.PAS.Condiciones())
	return out
}

// ---------------------------------------------------------------------------------------------
// PROCEDURAL ANSWER SUCCESS: el caso perfecto y cada condición fallando por separado

func TestPASPerfecto(t *testing.T) {
	b := baseSintetica(t)
	e := evaluarV2(t, b, casoSintetico(), planPerfecto())
	if e.PAS == nil {
		t.Fatal("PAS debería aplicar")
	}
	if got := condiciones(e); got != [7]bool{true, true, true, true, true, true, true} || !e.PAS.Exito {
		t.Fatalf("condiciones = %v, éxito = %v; quería las 7 en true", got, e.PAS.Exito)
	}
	if *e.RecallPasos != 1 || !*e.OrdenOK || e.Invencion {
		t.Fatalf("recall pasos %v, orden %v, invención %v", *e.RecallPasos, *e.OrdenOK, e.Invencion)
	}
	if e.FotosEsperadas != 3 || e.FotosPresentes != 3 || e.FotosAsignadas != 3 || e.FotosAsignadasOK != 3 {
		t.Fatalf("fotos esperadas/presentes/asignadas/ok = %d/%d/%d/%d", e.FotosEsperadas, e.FotosPresentes, e.FotosAsignadas, e.FotosAsignadasOK)
	}
	// Recuperación sin traza: lo citado en orden (s045 y luego s099/img) → R@5 = 1, MRR = 1.
	if *e.Recall5 != 1 || *e.MRR != 1 || math.Abs(*e.NDCG10-1) > 1e-9 {
		t.Fatalf("R@5 %v MRR %v nDCG %v", *e.Recall5, *e.MRR, *e.NDCG10)
	}
	if !*e.TipoOK || !*e.ProcOK || !*e.ExitoTarea {
		t.Fatal("tipo, procedimiento y éxito de la tarea deberían ser correctos")
	}
}

func TestPASCadaCondicionFallaSola(t *testing.T) {
	b := baseSintetica(t)
	casos := []struct {
		nombre  string
		cond    int // índice 0..6 de la condición que debe fallar
		cambiar func(p *tipos.Plan)
	}{
		{"C1 procedimiento equivocado", 0, func(p *tipos.Plan) {
			p.Procedimiento = &tipos.Ref{ID: "presupuestos.procesar-presupuesto", Titulo: "Procesar"}
		}},
		{"C2 falta un paso", 1, func(p *tipos.Plan) { p.Pasos = p.Pasos[:2] }},
		{"C3 orden alterado", 2, func(p *tipos.Plan) { p.Pasos[1], p.Pasos[2] = p.Pasos[2], p.Pasos[1] }},
		{"C4 paso inventado", 3, func(p *tipos.Plan) {
			p.Pasos = append(p.Pasos, tipos.PasoPlan{N: 4, ID: "x#4", Texto: "Reinicie Windows y llame a Microsoft para reinstalar el antivirus."})
		}},
		{"C5 foto en el paso equivocado", 4, func(p *tipos.Plan) {
			p.Pasos[0].Fotos, p.Pasos[1].Fotos = p.Pasos[1].Fotos, p.Pasos[0].Fotos
		}},
		{"C6 intención: responde como concepto", 5, func(p *tipos.Plan) { p.Tipo = tipos.Concepto }},
		{"C7 pasos sin fuentes", 6, func(p *tipos.Plan) {
			for i := range p.Pasos {
				p.Pasos[i].Fuente = nil
			}
		}},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			p := planPerfecto()
			tc.cambiar(p)
			e := evaluarV2(t, b, casoSintetico(), p)
			if e.PAS == nil {
				t.Fatal("PAS debería aplicar")
			}
			got := condiciones(e)
			want := [7]bool{true, true, true, true, true, true, true}
			want[tc.cond] = false
			if got != want {
				t.Fatalf("condiciones = %v, quería %v", got, want)
			}
			if e.PAS.Exito {
				t.Fatal("con una condición en falso el PAS no puede tener éxito")
			}
		})
	}
}

func TestPASNoAplica(t *testing.T) {
	b := baseSintetica(t)
	c := casoSintetico()
	c.ProcedimientoEsperado = nil
	c.FotosEsperadas = nil
	e := evaluarV2(t, b, c, planPerfecto())
	if e.PAS != nil || e.PASMotivo == "" {
		t.Fatalf("sin YAML esperado el PAS no aplica (motivo %q)", e.PASMotivo)
	}
	c = casoSintetico()
	c.TipoEsperado = "CONCEPT"
	if e := evaluarV2(t, b, c, planPerfecto()); e.PAS != nil {
		t.Fatal("PAS solo aplica a tipo_esperado PROCEDURE")
	}
}

// ---------------------------------------------------------------------------------------------
// Entrega por partes y continuación («¿y luego?»)

func TestEntregaPorPartesYContinuacion(t *testing.T) {
	b := baseSintetica(t)
	// Primer turno: el plan trae los 3 pasos pero muestra solo el 1 y el 2 (ParteMinima = 3 → no alcanza).
	p := planPerfecto()
	p.PasosMostrados = []int{1, 2}
	o1 := observar(b, "v2", turnoV2(p, ""))
	if !o1.EntregaParcial || len(o1.Pasos) != 2 {
		t.Fatalf("entrega parcial %v con %d pasos", o1.EntregaParcial, len(o1.Pasos))
	}
	e := evaluarCaso(b, casoSintetico(), []Observacion{o1})
	if !reflect.DeepEqual(e.PasosEsperadosTurno, []int{1, 2, 3}) || e.PAS.C2Pasos {
		t.Fatalf("una parte de 2 pasos (< ParteMinima) no cumple C2: esperados %v", e.PasosEsperadosTurno)
	}
	// Continuación: el turno anterior entregó 1–2; «¿y luego?» espera solo el 3.
	c := casoSintetico()
	c.Turnos = []string{"como registro metrado", "¿y luego?"}
	c.Continuacion = true
	p2 := planPerfecto()
	p2.Pasos = p2.Pasos[2:]
	o2 := observar(b, "v2", turnoV2(p2, ""))
	e = evaluarCaso(b, c, []Observacion{o1, o2})
	if !reflect.DeepEqual(e.PasosEsperadosTurno, []int{3}) || !e.PAS.Exito {
		t.Fatalf("continuación: esperados %v, PAS %+v", e.PasosEsperadosTurno, e.PAS)
	}
	// Si el turno anterior ya entregó todo, la continuación no tiene pasos pendientes: PAS no aplica.
	o1b := observar(b, "v2", turnoV2(planPerfecto(), ""))
	e = evaluarCaso(b, c, []Observacion{o1b, o2})
	if e.PAS != nil || e.PASMotivo != "continuación sin pasos pendientes" {
		t.Fatalf("PAS %+v motivo %q", e.PAS, e.PASMotivo)
	}
}

// ---------------------------------------------------------------------------------------------
// Heurística de V1: pasos por líneas numeradas, fotos markdown, fuentes por cita

func TestV1Heuristica(t *testing.T) {
	b := baseSintetica(t)
	resp := "Para ingresar metrados:\n" +
		"1. Ubica el cursor en la partida de la hoja del presupuesto.\n" +
		"![Captura del manual](fotos/" + rutaF1 + ")\n" +
		"2. Escribe la cantidad en la celda Metrado de la partida.\n" +
		"![Captura del manual](fotos/" + rutaF2 + ")\n" +
		"3. Pulsa Procesar para recalcular el presupuesto con los metrados."
	tu := turnoV1(resp, "procedimiento",
		fuenteAsk{Cita: "Manual de Presupuestos › 3.5 Hoja"},
		fuenteAsk{Cita: "Manual de Presupuestos › 5.1 Procesar", Fotos: []string{rutaF3}},
		fuenteAsk{Cita: "Una cita que no está en la KB"})
	o := observar(b, "v1", tu)
	if o.PlanPresente || o.TipoPredicho != "PROCEDURE" || len(o.Pasos) != 3 {
		t.Fatalf("plan %v tipo %q pasos %d", o.PlanPresente, o.TipoPredicho, len(o.Pasos))
	}
	if len(o.CitasSinResolver) != 1 || len(o.Citadas) != 2 {
		t.Fatalf("citadas %v sin resolver %v", o.Citadas, o.CitasSinResolver)
	}
	e := evaluarCaso(b, casoSintetico(), []Observacion{o})
	if got := condiciones(e); got != [7]bool{true, true, true, true, true, true, true} {
		t.Fatalf("V1 bien formada: condiciones %v (entregados %v, inventados %d)", got, e.PasosEntregados, e.PasosInventados)
	}
	// Foto 3 está en la galería (fuentes[].fotos): cuenta para recall, no para la exactitud paso ↔ foto.
	if e.FotosPresentes != 3 || e.FotosAsignadas != 2 || e.FotosAsignadasOK != 2 || e.FotosEnSuPaso != 2 {
		t.Fatalf("presentes %d asignadas %d ok %d en su paso %d", e.FotosPresentes, e.FotosAsignadas, e.FotosAsignadasOK, e.FotosEnSuPaso)
	}

	// Respuesta de respaldo (sin modelo): ni pasos ni intención.
	r := turnoV1("Encontré fuentes relacionadas, pero no pude preparar un resumen ahora.", "procedimiento",
		fuenteAsk{Cita: "Manual de Presupuestos › 3.5 Hoja"})
	r.Respuesta.Modo = "modelo_no_disponible"
	o = observar(b, "v1", r)
	e = evaluarCaso(b, casoSintetico(), []Observacion{o})
	if !o.Respaldo || e.PAS.C6Intencion || e.PAS.C2Pasos || e.PAS.C1Procedimiento || e.GeneracionOK || e.ConContenido {
		t.Fatalf("respaldo mal calificado: %+v", e.PAS)
	}
	// Sin pasos: C3, C5 y C7 no se cumplen «en vacío»; C4 (no inventa) sí.
	if e.PAS.C3Orden || e.PAS.C5Fotos || e.PAS.C7Fuentes || !e.PAS.C4SinInventar {
		t.Fatalf("condiciones sin pasos: %+v", e.PAS)
	}

	// Una línea inventada (sin soporte en lo citado) y otra ajena pero respaldada.
	resp2 := resp + "\n4. Reinstala Windows desde el disco de recuperación del fabricante.\n5. Registre el pedido de compra con sus insumos."
	tu2 := turnoV1(resp2, "procedimiento", fuenteAsk{Cita: "Manual de Presupuestos › 3.5 Hoja"},
		fuenteAsk{Cita: "Manual de Compras › 3.5 Pedidos"}, fuenteAsk{Cita: "Manual de Presupuestos › 5.1 Procesar"})
	e = evaluarCaso(b, casoSintetico(), []Observacion{observar(b, "v1", tu2)})
	if e.PasosInventados != 1 || e.PasosAjenos != 1 || e.PAS.C4SinInventar || !e.Invencion {
		t.Fatalf("inventados %d ajenos %d C4 %v", e.PasosInventados, e.PasosAjenos, e.PAS.C4SinInventar)
	}
}

// TestAlineacionMonotona: V1 funde pasos y repite el nombre de un escenario; el orden se juzga con la
// alineación monótona, y una negrita repetida en dos pasos no identifica a ninguno.
func TestAlineacionMonotona(t *testing.T) {
	p := &Procedimiento{ID: "x.p", Pasos: []PasoYAML{
		{N: 1, ID: "x.p#1", Accion: "Ingrese al escenario **Datos Generales**."},
		{N: 2, ID: "x.p#2", Accion: "Elija **Nuevo SubItem** en el catálogo de presupuestos."},
		{N: 3, ID: "x.p#3", Accion: "Haga doble clic para trasladarlo al árbol de **Datos Generales**."},
		{N: 4, ID: "x.p#4", Accion: "Registre el **plazo** de ejecución en días calendario."},
	}}
	if ex := negritasExclusivas(unidades(p)); ex["datos generales"] || !ex["plazo"] || !ex["nuevo subitem"] {
		t.Fatalf("exclusivas %v", ex)
	}
	lineas := func(ts ...string) []PasoObs {
		var out []PasoObs
		for i, x := range ts {
			out = append(out, PasoObs{Pos: i, Texto: x})
		}
		return out
	}
	bien := lineas("Ingresa al escenario Datos Generales.", "Elige Nuevo SubItem en el catálogo de presupuestos.",
		"Completa descripción, cliente y plazo; haz doble clic para trasladarlo al árbol de Datos Generales.")
	m := mapearPasos(bien, p)
	if !ordenConservado(m) || !reflect.DeepEqual(m.Entregado, []int{1, 2, 3, 4}) {
		t.Fatalf("orden %v entregados %v", ordenConservado(m), m.Entregado)
	}
	// «plazo» (negrita exclusiva del paso 4) está en la línea 2, la misma del paso 3: varios pasos por línea valen.
	if m.PosDeN[4] != 2 || m.PosDeN[3] != 2 {
		t.Fatalf("posiciones %v", m.PosDeN)
	}
	mal := lineas("Elige Nuevo SubItem en el catálogo de presupuestos.", "Ingresa al escenario Datos Generales.")
	if ordenConservado(mapearPasos(mal, p)) {
		t.Fatal("el paso 2 antes que el 1 rompe el orden")
	}
	// Con ids (V2): repetir un paso o volver atrás rompe el orden.
	ids := []PasoObs{{ID: "x.p#1"}, {ID: "x.p#2"}, {ID: "x.p#2"}}
	if ordenConservado(mapearPasos(ids, p)) {
		t.Fatal("paso repetido")
	}
}

func TestPasosDeTexto(t *testing.T) {
	txt := "Intro\n**Paso 1:** Abra el escenario.\n![x](fotos/a/b.png)\n2) Elija Nuevo.\n   sigue el paso 2\n![y](/fotos/c/d.png)\nNota final\n![z](fotos/e.png)"
	pasos, sueltas := pasosDeTexto(txt)
	if len(pasos) != 2 || pasos[0].Texto != "Abra el escenario." || pasos[1].Texto != "Elija Nuevo." {
		t.Fatalf("pasos %+v", pasos)
	}
	if !reflect.DeepEqual(pasos[0].Fotos, []string{claveFoto("a/b.png", "")}) || !reflect.DeepEqual(pasos[1].Fotos, []string{claveFoto("c/d.png", "")}) {
		t.Fatalf("fotos por paso %v %v", pasos[0].Fotos, pasos[1].Fotos)
	}
	if !reflect.DeepEqual(sueltas, []string{claveFoto("e.png", "")}) {
		t.Fatalf("sueltas %v", sueltas)
	}
	if p, _ := pasosDeTexto("- uno\n- dos"); len(p) != 2 {
		t.Fatalf("viñetas: %d pasos", len(p))
	}
	if p, _ := pasosDeTexto("Solo prosa, sin pasos."); len(p) != 0 {
		t.Fatal("la prosa no entrega pasos")
	}
}

// ---------------------------------------------------------------------------------------------
// Recuperación

func TestMetricasRanking(t *testing.T) {
	rel := [][]string{{"A"}, {"B"}, {"C"}}
	ranking := []ItemRanking{{Claves: []string{"X"}}, {Claves: []string{"A"}}, {Claves: []string{"A"}}, {Claves: []string{"B", "Z"}}}
	r5, r10, mrr, ndcg := metricasRanking(ranking, rel)
	if math.Abs(r5-2.0/3) > 1e-9 || math.Abs(r10-2.0/3) > 1e-9 {
		t.Fatalf("recall %v %v", r5, r10)
	}
	if mrr != 0.5 {
		t.Fatalf("MRR %v", mrr)
	}
	// g = [0,1,0,1] (el segundo A no suma) → DCG = 1/log2(3) + 1/log2(5); IDCG = 1 + 1/log2(3) + 1/log2(4)
	want := (1/math.Log2(3) + 1/math.Log2(5)) / (1 + 1/math.Log2(3) + 1/math.Log2(4))
	if math.Abs(ndcg-want) > 1e-9 {
		t.Fatalf("nDCG %v, quería %v", ndcg, want)
	}
	// Recall@5 frente a @10: el acierto en la posición 7 solo cuenta para @10.
	largo := make([]ItemRanking, 7)
	largo[6] = ItemRanking{Claves: []string{"C"}}
	r5, r10, mrr, _ = metricasRanking(largo, rel)
	if r5 != 0 || math.Abs(r10-1.0/3) > 1e-9 || math.Abs(mrr-1.0/7) > 1e-9 {
		t.Fatalf("r5 %v r10 %v mrr %v", r5, r10, mrr)
	}
	if r5, r10, mrr, ndcg := metricasRanking(nil, rel); r5+r10+mrr+ndcg != 0 {
		t.Fatal("ranking vacío = 0")
	}
}

// ---------------------------------------------------------------------------------------------
// Abstención, clasificación, agregados

func TestAbstencionYClasificacion(t *testing.T) {
	b := baseSintetica(t)
	c := Caso{ID: "t-sin", Categoria: "PROCEDURE", Turnos: []string{"como exporto a SAP"}, TipoEsperado: "PROCEDURE",
		SinEvidenciaEsperada: true, Sintetico: true}
	r := turnoV1("No tengo contexto suficiente para responder a esa pregunta.", "procedimiento")
	r.Respuesta.SinContexto = true
	r.Respuesta.Modo = "sin_contexto"
	e := evaluarCaso(b, c, []Observacion{observar(b, "v1", r)})
	if e.AbstencionOK == nil || !*e.AbstencionOK || !*e.ExitoTarea {
		t.Fatal("decir SIN_EVIDENCIA cuando no hay evidencia es una abstención correcta")
	}
	// V1 no tiene NAVIGATION: «informacion_directa» no equivale a ningún tipo de V2.
	c2 := Caso{ID: "t-nav", Categoria: "NAVIGATION", Turnos: []string{"donde veo metrados"}, TipoEsperado: "NAVIGATION",
		FragmentosRelevantes: []RefFragmento{{ID: "web-s045", Manual: "Manual de Presupuestos"}}, Sintetico: true}
	e = evaluarCaso(b, c2, []Observacion{observar(b, "v1", turnoV1("En la hoja del presupuesto.", "informacion_directa",
		fuenteAsk{Cita: "Manual de Presupuestos › 3.5 Hoja"}))})
	if *e.TipoOK || *e.ExitoTarea || !e.CitaRelevante {
		t.Fatalf("tipo %v éxito %v cita relevante %v", *e.TipoOK, *e.ExitoTarea, e.CitaRelevante)
	}
	// Tipos aceptables en un caso ambiguo.
	c3 := Caso{ID: "t-amb", Categoria: "AMBIGUOUS", Turnos: []string{"metrado s10"}, TipoEsperado: "UNKNOWN",
		TiposAceptables: []string{"UNKNOWN", "CONCEPT"}, FragmentosRelevantes: []RefFragmento{{ID: "web-s045", Manual: "Manual de Presupuestos"}}, Sintetico: true}
	e = evaluarCaso(b, c3, []Observacion{observar(b, "v1", turnoV1("Un metrado es la cantidad de la partida en la hoja del presupuesto.",
		"concepto", fuenteAsk{Cita: "Manual de Presupuestos › 3.5 Hoja"}))})
	if !*e.TipoOK || !*e.ExitoTarea {
		t.Fatal("CONCEPT está entre los tipos aceptables del caso ambiguo")
	}
}

func TestAgregadosYPercentil(t *testing.T) {
	if p := percentil([]float64{5, 1, 4, 2, 3}, 50); p != 3 {
		t.Fatalf("p50 %v", p)
	}
	if p := percentil([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 95); p != 10 {
		t.Fatalf("p95 %v", p)
	}
	b := baseSintetica(t)
	c := casoSintetico()
	bien := evaluarV2(t, b, c, planPerfecto())
	mal := planPerfecto()
	mal.Pasos = mal.Pasos[:1]
	e2 := evaluarV2(t, b, c, mal)
	e2.Caso = "t-proc-2"
	a := agregar([]EvalCaso{bien, e2}, map[string]Caso{"t-proc": c, "t-proc-2": c})
	if a.PAS.Total != 2 || a.PAS.Si != 1 || a.PAS.Valor != 0.5 {
		t.Fatalf("PAS agregado %+v", a.PAS)
	}
	if a.PASCondiciones[1].Si != 1 || a.PASCondiciones[0].Si != 2 {
		t.Fatalf("condiciones agregadas C1 %+v C2 %+v", a.PASCondiciones[0], a.PASCondiciones[1])
	}
	if math.Abs(a.RecallPasos.Valor-(1+1.0/3)/2) > 1e-9 {
		t.Fatalf("recall de pasos medio %v", a.RecallPasos.Valor)
	}
}

func TestRefFragmentoFormas(t *testing.T) {
	var c Caso
	l := `{"id":"x","categoria":"CONCEPT","turnos":["a",{"texto":"b"}],"tipo_esperado":"CONCEPT",` +
		`"fragmentos_relevantes":[{"id":"web-s045","manual":"Manual de Presupuestos"},"web-s099@Manual de Presupuestos","img-ccc-0000"]}`
	if err := json.Unmarshal([]byte(l), &c); err != nil {
		t.Fatal(err)
	}
	want := []RefFragmento{{"web-s045", "Manual de Presupuestos"}, {"web-s099", "Manual de Presupuestos"}, {"img-ccc-0000", ""}}
	if !reflect.DeepEqual(c.FragmentosRelevantes, want) || !reflect.DeepEqual(c.Turnos, []string{"a", "b"}) {
		t.Fatalf("%+v %v", c.FragmentosRelevantes, c.Turnos)
	}
	b := baseSintetica(t)
	// «web-s045» a secas es ambiguo (dos manuales): el validador lo rechaza.
	c2 := Caso{ID: "amb", Categoria: "CONCEPT", Turnos: []string{"q"}, TipoEsperado: "CONCEPT", Sintetico: true,
		FragmentosRelevantes: []RefFragmento{{ID: "web-s045"}}}
	if errs, _ := validarDataset([]Caso{c2}, b); len(errs) != 1 || !strings.Contains(errs[0], "ambiguo") {
		t.Fatalf("errores %v", errs)
	}
}

// ---------------------------------------------------------------------------------------------
// Lector YAML

func TestYAMLMinimo(t *testing.T) {
	src := `id: presupuestos.x  # comentario
version: 1
titulo: "Un \"título\" # no es comentario"
lista: [a, "b, c", 3]
mapa: {id: web-1, manual: "Manual de Presupuestos", pagina: null}
pasos:
  - n: 1
    fuente: [x, "7c40-0018"]
    fotos:
      - id: "0b6b"
        ruta_o_url: "imagenes/a.png"
    sub:
      - n: 1
        accion: "sub"
  - n: 2
vacio:
preguntas:
  - "uno"
  - dos
`
	m, err := yamlMinimo(src)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(m)
	want := `{"id":"presupuestos.x","lista":["a","b, c",3],"mapa":{"id":"web-1","manual":"Manual de Presupuestos","pagina":null},` +
		`"pasos":[{"fotos":[{"id":"0b6b","ruta_o_url":"imagenes/a.png"}],"fuente":["x","7c40-0018"],"n":1,"sub":[{"accion":"sub","n":1}]},{"n":2}],` +
		`"preguntas":["uno","dos"],"titulo":"Un \"título\" # no es comentario","vacio":null,"version":1}`
	if string(b) != want {
		t.Fatalf("\n got %s\nwant %s", b, want)
	}
	if _, err := yamlMinimo("a: |\n  bloque"); err == nil {
		t.Fatal("un bloque literal está fuera del subconjunto")
	}
}

// TestYAMLIgualQuePyYAML compara el lector mínimo con PyYAML en todos los YAML de kb/procedimientos y
// kb/conceptos. Se salta si no hay repo o no hay PyYAML (no es una dependencia de las pruebas).
func TestYAMLIgualQuePyYAML(t *testing.T) {
	raiz, err := filepath.Abs("../../..")
	if err != nil {
		t.Skip(err)
	}
	py := filepath.Join(raiz, ".venv", "bin", "python3")
	if _, err := os.Stat(py); err != nil {
		t.Skip("sin .venv con PyYAML")
	}
	var archivos []string
	for _, patron := range []string{"kb/procedimientos/*/*.yml", "kb/conceptos/*.yml"} {
		m, _ := filepath.Glob(filepath.Join(raiz, patron))
		archivos = append(archivos, m...)
	}
	if len(archivos) == 0 {
		t.Skip("sin YAML en kb/")
	}
	for _, a := range archivos {
		src, _ := os.ReadFile(a)
		mio, err := yamlMinimo(string(src))
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(a), err)
			continue
		}
		out, err := exec.Command(py, "-c", "import sys,json,yaml\nprint(json.dumps(yaml.safe_load(open(sys.argv[1])),ensure_ascii=False,default=str,sort_keys=True))", a).Output()
		if err != nil {
			t.Skipf("PyYAML no disponible: %v", err)
		}
		var suyo any
		json.Unmarshal(out, &suyo)
		var mioGen any
		bm, _ := json.Marshal(mio)
		json.Unmarshal(bm, &mioGen)
		if !reflect.DeepEqual(mioGen, suyo) {
			t.Errorf("%s: el lector mínimo difiere de PyYAML", strings.TrimPrefix(a, raiz+"/"))
		}
	}
}

// ---------------------------------------------------------------------------------------------
// Flujo HTTP: sondeo de V2, hilo como la página y memoria reenviada

func TestCorrerVersionConServidorSimulado(t *testing.T) {
	b := baseSintetica(t)
	var recibidas []peticionAsk
	conV2 := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p peticionAsk
		json.NewDecoder(r.Body).Decode(&p)
		recibidas = append(recibidas, p)
		res := map[string]any{"respuesta": "Encontré fuentes relacionadas, pero no pude preparar un resumen ahora.",
			"modo": "modelo_no_disponible", "orquestacion": map[string]any{"tipo_consulta": "procedimiento", "ruta": "rag"},
			"traza": map[string]any{"version": 1, "total_ms": 3.5, "etapas": []any{}}}
		if p.Version == "v2" && conV2 {
			res = map[string]any{"respuesta": "1. …", "modo": "v2", "plan": planPerfecto(),
				"memoria": map[string]any{"procedimiento_id": "presupuestos.ingresar-metrados", "paso_actual": 3}}
		}
		json.NewEncoder(w).Encode(res)
	}))
	defer srv.Close()

	c := casoSintetico()
	c.Turnos = []string{"como registro metrado", "¿y luego?"}
	op := opciones{URL: srv.URL, Timeout: 5 * time.Second, Source: "s10-kb", K: 8}
	rv := correrVersion(op, b, "v2", []Caso{c}, map[string]Caso{c.ID: c})
	if !rv.Disponible || rv.TurnosPlan != 2 || rv.Global.Casos != 1 {
		t.Fatalf("disponible %v turnos con plan %d casos %d", rv.Disponible, rv.TurnosPlan, rv.Global.Casos)
	}
	// Las 3 de sondeo + 2 del caso. El segundo turno lleva la memoria del primero y el hilo de la página.
	if len(recibidas) != 5 {
		t.Fatalf("%d peticiones", len(recibidas))
	}
	t1, t2 := recibidas[3], recibidas[4]
	if len(t1.Memoria) != 0 || !strings.Contains(string(t2.Memoria), `"paso_actual":3`) {
		t.Fatalf("memoria t1 %s t2 %s", t1.Memoria, t2.Memoria)
	}
	// hilo = turnos ANTERIORES (como la página): vacío en el primero, [usuario, asistente] en el segundo.
	if !t1.Traza || t1.Version != "v2" || t1.Hilo == nil || len(t1.Hilo) != 0 || len(t2.Hilo) != 2 ||
		t2.Hilo[0].Texto != "como registro metrado" || t2.Hilo[1].Rol != "asistente" {
		t.Fatalf("hilo t1 %+v t2 %+v", t1.Hilo, t2.Hilo)
	}

	// Sin V2 en el servidor: no disponible, sin fallar; V1 corre igual.
	conV2 = false
	rv = correrVersion(op, b, "v2", []Caso{c}, map[string]Caso{c.ID: c})
	if rv.Disponible || rv.Motivo == "" || len(rv.Casos) != 0 {
		t.Fatalf("V2 sin plan debería quedar no disponible: %+v", rv.Motivo)
	}
	rv = correrVersion(op, b, "v1", []Caso{c}, map[string]Caso{c.ID: c})
	if !rv.Disponible || rv.Global.Casos != 1 || rv.Global.Respaldo.Valor != 1 {
		t.Fatalf("V1: %+v", rv.Global)
	}
}
