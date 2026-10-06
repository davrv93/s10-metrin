package conocimiento

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"rag-go/internal/v2/tipos"
)

// planYTexto: el plan CONCEPT de «¿qué es <término>?» y su texto con las plantillas, y el gate.
func planYTexto(t *testing.T, b *Base, q string) (tipos.Plan, string, tipos.ResultadoCalidad) {
	t.Helper()
	c := consulta(b, q)
	e := tipos.Estado{Pregunta: q, Consulta: c, Tipo: tipos.Concepto}
	p, err := NuevoConstructor(b, NuevoRecuperador(b)).Construir(context.Background(), e, nil)
	if err != nil || p.Concepto == nil {
		t.Fatalf("%q: plan %v %+v", q, err, p)
	}
	txt, err := NuevoMotorPlantillas(b).Redactar(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	return p, txt, NuevoGate(b).Evaluar(p, txt, e)
}

// Un término definido por uso lleva `notas` y el aviso en el plan, y la línea en el texto (una vez); uno con
// definición formal no lleva aviso ni línea. Con la plantilla del archivo o con las embebidas.
func TestAvisoPorUso_PlanYTexto(t *testing.T) {
	for _, nombre := range []string{"archivo", "embebidas"} {
		b := cargarFixture(t)
		if nombre == "embebidas" {
			b = cargarFixtureEmbebidas(t)
		}
		b.porConcepto["metrado"].Notas = "Las fuentes no traen una definición formal de metrado; se describe solo por su uso en S10."
		p, txt, g := planYTexto(t, b, "que es metrado")
		want := "Los manuales no traen una definición formal de «" + p.Concepto.Termino + "»; se lo describo por su uso en S10."
		if p.Concepto.Aviso != want || p.Concepto.Notas == "" {
			t.Errorf("%s: aviso %q, notas %q", nombre, p.Concepto.Aviso, p.Concepto.Notas)
		}
		if strings.Count(txt, want) != 1 {
			t.Errorf("%s: el texto debe traer el aviso una vez:\n%s", nombre, txt)
		}
		if !g.Paso {
			t.Errorf("%s: el gate rechaza el aviso: %v", nombre, g.Problemas)
		}
		js, _ := json.Marshal(p)
		if !strings.Contains(string(js), `"aviso":`) || !strings.Contains(string(js), `"notas":`) {
			t.Errorf("%s: el JSON del plan no lleva aviso/notas: %s", nombre, js)
		}

		// Con definición formal: ni aviso ni línea, y el JSON no trae los campos (omitempty).
		p, txt, _ = planYTexto(t, b, "que es una partida")
		if p.Concepto.Aviso != "" || strings.Contains(txt, "definición formal") {
			t.Errorf("%s: partida sin notas lleva aviso: %q / %s", nombre, p.Concepto.Aviso, txt)
		}
		if js, _ := json.Marshal(p); strings.Contains(string(js), `"aviso"`) || strings.Contains(string(js), `"notas"`) {
			t.Errorf("%s: omitempty: %s", nombre, js)
		}
	}
}

// Una nota que no es «por uso» (p. ej. «No confundir con…») viaja en notas, pero no hay aviso.
func TestAvisoPorUso_OtrasNotasNoAvisan(t *testing.T) {
	b := cargarFixture(t)
	b.porConcepto["metrado"].Notas = "No confundir con el avance valorizado."
	p, txt, _ := planYTexto(t, b, "que es metrado")
	if p.Concepto.Aviso != "" || p.Concepto.Notas == "" || strings.Contains(txt, "definición formal") {
		t.Errorf("aviso %q, notas %q, texto %q", p.Concepto.Aviso, p.Concepto.Notas, txt)
	}
}

// NAVIGATION no muestra la definición, así que tampoco su aviso.
func TestAvisoPorUso_NavegacionSinAviso(t *testing.T) {
	b := cargarFixture(t)
	def := b.porConcepto["metrado"].ConceptoDef
	def.EnS10 = "Se registra en la hoja del presupuesto."
	def.Aviso = "Los manuales no traen una definición formal de «Metrado»; se lo describo por su uso en S10."
	p := tipos.Plan{Version: 2, Tipo: tipos.Navegacion, Concepto: &def, Fuentes: []tipos.FuentePlan{}}
	txt, err := NuevoMotorPlantillas(b).Redactar(context.Background(), p)
	if err != nil || strings.Contains(txt, "definición formal") || !strings.Contains(txt, "hoja del presupuesto") {
		t.Errorf("navegación: %v %q", err, txt)
	}
}

// Contra kb/ real: los términos marcados son exactamente los seis de kb/conceptos/SIN_FUENTE.md («Decisión del
// usuario: mantener con aviso»), y la plantilla real trae la parte «aviso» (no la embebida).
func TestReal_ConceptosPorUso(t *testing.T) {
	b := cargarReal(t)
	var marcados []string
	for _, c := range b.Conceptos {
		if c.DefinidoPorUso() {
			marcados = append(marcados, c.ID)
		}
	}
	sort.Strings(marcados)
	if want := []string{"cts", "escenario", "metrado", "partida", "subpresupuesto", "titulo"}; strings.Join(marcados, ",") != strings.Join(want, ",") {
		t.Errorf("definidos por uso: %v, quiero %v (si cambió la lista, actualice kb/conceptos/SIN_FUENTE.md)", marcados, want)
	}
	if s := b.Plantillas.Parte("CONCEPTO", "aviso", map[string]string{"TERMINO_POR_USO": "Metrado"}); !strings.Contains(s, "«Metrado»") {
		t.Errorf("metrin/plantillas/respuestas.yml: CONCEPTO sin la parte «aviso»: %q", s)
	}
	p, txt, g := planYTexto(t, b, "¿qué es un metrado?")
	if p.Concepto.ID != "metrado" || p.Concepto.Aviso == "" || strings.Count(txt, p.Concepto.Aviso) != 1 || !g.Paso {
		t.Errorf("metrado real: aviso %q, gate %v, texto:\n%s", p.Concepto.Aviso, g.Problemas, txt)
	}
}
