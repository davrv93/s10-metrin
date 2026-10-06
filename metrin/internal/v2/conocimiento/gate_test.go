package conocimiento

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"rag-go/internal/v2/tipos"
)

func copiar(t *testing.T, p tipos.Plan) tipos.Plan {
	t.Helper()
	b, _ := json.Marshal(p)
	var q tipos.Plan
	if err := json.Unmarshal(b, &q); err != nil {
		t.Fatal(err)
	}
	return q
}

// evaluarRegla: el caso debe fallar por la regla dada (al menos) y descontar su peso, o pasar si regla == "".
// Puede arrastrar otras reglas cuando es lo correcto (un paso sin fuente tampoco respalda sus negritas).
func evaluarRegla(t *testing.T, g *Gate, nombre, regla string, p tipos.Plan, texto string, e tipos.Estado) {
	t.Helper()
	res := g.Evaluar(p, texto, e)
	if regla == "" {
		if !res.Paso || res.Puntaje != 1 || len(res.Problemas) > 0 {
			t.Errorf("%s: debía pasar: %v", nombre, res.Problemas)
		}
		return
	}
	if res.Paso || res.Puntaje >= 1 {
		t.Errorf("%s: debía fallar por «%s» y pasó (%.2f)", nombre, regla, res.Puntaje)
		return
	}
	hay := false
	for _, pr := range res.Problemas {
		hay = hay || strings.HasPrefix(pr, regla+": ")
	}
	if !hay {
		t.Errorf("%s: no falló por «%s»: %v", nombre, regla, res.Problemas)
	}
	if tope := 1 - pesosGate[regla]/1.15; res.Puntaje > tope+0.001 {
		t.Errorf("%s: puntaje %.3f, debía ser ≤ %.3f", nombre, res.Puntaje, tope)
	}
}

func TestGate_Reglas(t *testing.T) {
	en := nuevoEntorno(t)
	ctx := context.Background()
	e := estado(en.b, "como registro metrado", tipos.Procedimiento)
	base := planDe(t, en, "como registro metrado", tipos.Procedimiento)
	texto, _ := en.gen.Redactar(ctx, base)
	en.gen.FotosEnTexto = true
	textoFotos, _ := en.gen.Redactar(ctx, base)
	en.gen.FotosEnTexto = false

	evaluarRegla(t, en.g, "plan bueno", "", base, texto, e)
	evaluarRegla(t, en.g, "plan bueno con fotos en el texto", "", base, textoFotos, e)

	type caso struct {
		nombre, regla string
		mod           func(p *tipos.Plan) // cambia el plan; el texto se vuelve a redactar
		texto         func(s string) string
		e             *tipos.Estado
	}
	conceptoE := estado(en.b, "como registro metrado", tipos.Concepto)
	casos := []caso{
		// intención
		{nombre: "pidió CONCEPT y respondió PROCEDURE", regla: "intencion", e: &conceptoE},
		{nombre: "SIN_EVIDENCIA con pasos", regla: "intencion", mod: func(p *tipos.Plan) { p.SinEvidencia = true }},
		// evidencia
		{nombre: "cita fuera de la tabla", regla: "evidencia", mod: func(p *tipos.Plan) { p.Pasos[0].Fuente = []string{"web-no-existe"} }},
		{nombre: "cita de otro paso", regla: "evidencia", mod: func(p *tipos.Plan) { p.Pasos[0].Fuente = append(p.Pasos[0].Fuente, "img-bbb222bbb222-0000") }},
		{nombre: "paso sin fuente", regla: "evidencia", mod: func(p *tipos.Plan) { p.Pasos[2].Fuente = nil }},
		{nombre: "prerrequisito inventado", regla: "evidencia", mod: func(p *tipos.Plan) { p.Prerrequisitos[0].Texto = "Debe tener licencia Premium." }},
		// negritas
		{nombre: "menú que no está en la fuente", regla: "negritas", mod: func(p *tipos.Plan) {
			p.Pasos[0].Texto = "Ubíquese en la partida del escenario **Hoja del presupuesto** y abra **Menú Archivo**."
		}},
		// orden
		{nombre: "pasos invertidos", regla: "orden", mod: func(p *tipos.Plan) { p.Pasos[0], p.Pasos[1] = p.Pasos[1], p.Pasos[0] }},
		{nombre: "falta un paso", regla: "orden", mod: func(p *tipos.Plan) { p.Pasos = p.Pasos[:2]; p.PasosMostrados = []int{1, 2} }},
		{nombre: "paso inventado", regla: "orden", mod: func(p *tipos.Plan) {
			x := p.Pasos[2]
			x.N, x.ID = 4, "presupuestos.registrar-metrado#4"
			p.Pasos = append(p.Pasos, x)
			p.PasosMostrados = []int{1, 2, 3, 4}
		}},
		{nombre: "pasos mostrados salteados", regla: "orden", mod: func(p *tipos.Plan) { p.PasosMostrados = []int{1, 3} }},
		// fotos
		{nombre: "foto de otro paso", regla: "fotos", mod: func(p *tipos.Plan) {
			p.Pasos[0].Fotos = append(p.Pasos[0].Fotos, p.Pasos[1].Fotos...)
			p.Pasos[1].Fotos = nil
		}},
		{nombre: "foto del mismo id de sección en otro manual", regla: "fotos", mod: func(p *tipos.Plan) {
			r := "imagenes/manual-de-almacenes/ccc333ccc333.png"
			p.Pasos[2].Fotos = []tipos.Foto{{ID: idFoto(r), Ruta: r}}
		}},
		{nombre: "foto repetida en dos pasos", regla: "fotos", mod: func(p *tipos.Plan) { p.Pasos[1].Fotos = append(p.Pasos[1].Fotos, p.Pasos[0].Fotos...) }},
		// fuentes
		{nombre: "fuente de marketing", regla: "fuentes", mod: func(p *tipos.Plan) {
			p.Fuentes = append(p.Fuentes, tipos.FuentePlan{Manual: "Optimiza 360 · Servicios: ERP S10"})
		}},
		{nombre: "fuente tangencial (nadie la cita)", regla: "fuentes", mod: func(p *tipos.Plan) {
			p.Fuentes = append(p.Fuentes, tipos.FuentePlan{Manual: "Manual de Compras"})
		}},
		{nombre: "sin fuentes", regla: "fuentes", mod: func(p *tipos.Plan) { p.Fuentes = []tipos.FuentePlan{} },
			texto: func(string) string { return texto }},
		// texto
		{nombre: "negrita fuera del plan", regla: "texto", texto: func(s string) string { return s + "\nLuego abra **Archivo**." }},
		{nombre: "atajo fuera del plan", regla: "texto", texto: func(s string) string { return s + "\nTambién puede usar Ctrl+G." }},
		{nombre: "ruta de menú fuera del plan", regla: "texto", texto: func(s string) string { return s + "\nEntre por Herramientas › Opciones › Avanzado." }},
		{nombre: "marca de cita", regla: "texto", texto: func(s string) string { return s + " [F1]" }},
		{nombre: "página que no está en las fuentes", regla: "texto", texto: func(s string) string { return s + "\nVea la página 99." }},
		{nombre: "manual que no está en las fuentes", regla: "texto", texto: func(s string) string { return s + "\nSegún el Manual de Nóminas, así es." }},
		{nombre: "paso que no se muestra", regla: "texto", texto: func(s string) string { return s + "\n7. Cierre todo." }},
		{nombre: "foto bajo otro paso", regla: "texto", texto: func(string) string {
			ls := strings.Split(textoFotos, "\n")
			for i, l := range ls { // sube la foto del paso 2 debajo del paso 1
				if strings.Contains(l, "bbb222bbb222") {
					ls = append(ls[:i], ls[i+1:]...)
					break
				}
			}
			for i, l := range ls {
				if strings.HasPrefix(l, "1. ") {
					ls = append(ls[:i+1], append([]string{"![x](fotos/imagenes/manual-de-presupuestos/bbb222bbb222.png)"}, ls[i+1:]...)...)
					break
				}
			}
			return strings.Join(ls, "\n")
		}},
		{nombre: "foto que el plan no trae", regla: "texto", texto: func(s string) string { return s + "\n![x](fotos/imagenes/otra/zzz.png)" }},
		{nombre: "texto vacío", regla: "texto", texto: func(string) string { return " " }},
		// afirmaciones
		{nombre: "afirmación con un par (id, manual) inexistente", regla: "afirmaciones", mod: func(p *tipos.Plan) { p.Afirmaciones[1].Manual = "Manual de Nóminas" }},
		{nombre: "afirmación sin manual", regla: "afirmaciones", mod: func(p *tipos.Plan) { p.Afirmaciones[0].Manual = "" }},
		{nombre: "afirmación de un paso que no está", regla: "afirmaciones", mod: func(p *tipos.Plan) { p.Afirmaciones[1].Paso = 9 }},
	}
	for _, c := range casos {
		p := copiar(t, base)
		txt := texto
		if c.mod != nil {
			c.mod(&p)
			if c.texto == nil {
				var err error
				if txt, err = en.gen.Redactar(ctx, p); err != nil {
					txt = texto
				}
			}
		}
		if c.texto != nil {
			txt = c.texto(txt)
		}
		ee := e
		if c.e != nil {
			ee = *c.e
		}
		evaluarRegla(t, en.g, c.nombre, c.regla, p, txt, ee)
	}
}

func TestGate_OtrosTipos(t *testing.T) {
	en := nuevoEntorno(t)
	ctx := context.Background()
	// Concepto: pasa; con la definición cambiada, evidencia; con pasos, intención.
	e := estado(en.b, "que es metrado", tipos.Concepto)
	p := planDe(t, en, "que es metrado", tipos.Concepto)
	txt, _ := en.gen.Redactar(ctx, p)
	evaluarRegla(t, en.g, "concepto bueno", "", p, txt, e)
	q := copiar(t, p)
	q.Concepto.Definicion = "Cuantificación de las cantidades de obra de cada partida, según la norma técnica 2025."
	txt2, _ := en.gen.Redactar(ctx, q)
	evaluarRegla(t, en.g, "definición ampliada", "evidencia", q, txt2, e)

	// Aclaración para una pregunta procedimental: válida (no inventa).
	a := planDe(t, en, "partida", tipos.Desconocido)
	ta, _ := en.gen.Redactar(ctx, a)
	evaluarRegla(t, en.g, "aclaración", "", a, ta, estado(en.b, "partida", tipos.Procedimiento))

	// SIN_EVIDENCIA: pasa, y no admite negritas en el texto.
	s := planDe(t, en, "receta de ceviche", tipos.Procedimiento)
	ts, _ := en.gen.Redactar(ctx, s)
	evaluarRegla(t, en.g, "sin evidencia", "", s, ts, estado(en.b, "receta de ceviche", tipos.Procedimiento))
	evaluarRegla(t, en.g, "sin evidencia con menú", "texto", s, ts+" Use **Archivo**.", estado(en.b, "receta de ceviche", tipos.Procedimiento))

	// Errores: uno que no está en el procedimiento.
	pe := planDe(t, en, "no puedo guardar metrado", tipos.Problema)
	pe.Errores[0].Solucion = "Reinstale S10."
	te, _ := en.gen.Redactar(ctx, pe)
	evaluarRegla(t, en.g, "solución inventada", "evidencia", pe, te, estado(en.b, "no puedo guardar metrado", tipos.Problema))

	// Una fuente con el nombre crudo del manual de la tabla («Importado a mano», como la arma el núcleo
	// en sus planes de memoria) es la misma que su título: no es tangencial.
	pm := planDe(t, en, "cómo modifico una partida", tipos.Procedimiento)
	pm.Fuentes = append(pm.Fuentes, tipos.FuentePlan{Manual: "Importado a mano", Seccion: "Guia de Usuario de S10 Presupuestos"})
	tm, _ := en.gen.Redactar(ctx, pm)
	evaluarRegla(t, en.g, "manual crudo de la tabla", "", pm, tm, estado(en.b, "cómo modifico una partida", tipos.Procedimiento))

	// Procedimiento inexistente.
	pi := copiar(t, planDe(t, en, "como registro metrado", tipos.Procedimiento))
	pi.Procedimiento.ID = "presupuestos.no-existe"
	ti, _ := en.gen.Redactar(ctx, pi)
	if res := en.g.Evaluar(pi, ti, estado(en.b, "x", tipos.Procedimiento)); res.Paso || !strings.HasPrefix(res.Problemas[0], "evidencia: el procedimiento") {
		t.Errorf("procedimiento inexistente: %v", res.Problemas)
	}
}
