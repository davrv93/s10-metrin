package conocimiento

import (
	"context"
	"strings"
	"testing"

	"rag-go/internal/v2/tipos"
)

type entorno struct {
	b   *Base
	r   *Recuperador
	c   *Constructor
	g   *Gate
	gen *MotorPlantillas
}

func nuevoEntorno(t *testing.T) entorno {
	b := cargarFixture(t)
	r := NuevoRecuperador(b)
	return entorno{b: b, r: r, c: NuevoConstructor(b, r), g: NuevoGate(b), gen: NuevoMotorPlantillas(b)}
}

func estado(b *Base, q string, tipo tipos.TipoRespuesta) tipos.Estado {
	return tipos.Estado{Pregunta: q, Tipo: tipo, Consulta: consulta(b, q)}
}

// plan + texto + gate; el gate debe pasar en todos los planes buenos.
func (en entorno) plan(t *testing.T, q string, tipo tipos.TipoRespuesta) (tipos.Plan, Informe, string) {
	t.Helper()
	e := estado(en.b, q, tipo)
	p, inf, err := en.c.ConstruirConInforme(context.Background(), e, nil)
	if err != nil {
		t.Fatalf("%q: %v", q, err)
	}
	txt, err := en.gen.Redactar(context.Background(), p)
	if err != nil {
		t.Fatalf("%q: redactar: %v", q, err)
	}
	if res := en.g.Evaluar(p, txt, e); !res.Paso {
		t.Errorf("%q: el gate rechaza un plan del constructor: %v\n%s", q, res.Problemas, txt)
	}
	return p, inf, txt
}

func fotosDe(p tipos.Plan, n int) []string {
	for _, x := range p.Pasos {
		if x.N == n {
			var ids []string
			for _, f := range x.Fotos {
				ids = append(ids, f.ID)
			}
			return ids
		}
	}
	return nil
}

func TestPlan_Procedimiento(t *testing.T) {
	en := nuevoEntorno(t)
	p, inf, txt := en.plan(t, "como registro metrado", tipos.Procedimiento)
	if inf.Ruta != "procedimiento" || p.Procedimiento == nil || p.Procedimiento.ID != "presupuestos.registrar-metrado" {
		t.Fatalf("procedimiento: %+v %+v", inf, p.Procedimiento)
	}
	if len(p.Pasos) != 3 || p.Pasos[0].N != 1 || p.Pasos[2].N != 3 || p.Pasos[1].ID != "presupuestos.registrar-metrado#2" {
		t.Errorf("pasos en orden: %+v", p.Pasos)
	}
	// Paso ↔ foto: cada paso con SUS fotos; el 3 sin foto; la foto de Almacenes (mismo id de sección) nunca.
	if f := fotosDe(p, 1); len(f) != 1 || f[0] != "88de6597f8de" {
		t.Errorf("fotos del paso 1: %v", f)
	}
	if f := fotosDe(p, 2); len(f) != 1 || f[0] != "23ab8dea6ec1" {
		t.Errorf("fotos del paso 2: %v", f)
	}
	if f := fotosDe(p, 3); len(f) != 0 {
		t.Errorf("el paso 3 no tiene foto en su fuente y queda sin foto: %v", f)
	}
	for _, x := range p.Pasos {
		for _, f := range x.Fotos {
			if f.ID == "b5f675189908" || strings.Contains(f.Ruta, "almacenes") {
				t.Errorf("entró la foto de otro manual: %+v", f)
			}
			if f.OCR != "" {
				t.Error("el OCR no viaja en el plan")
			}
		}
	}
	if len(p.PasosMostrados) != 3 || p.Siguiente != nil {
		t.Errorf("3 pasos: entrega completa: %v %+v", p.PasosMostrados, p.Siguiente)
	}
	if len(p.Prerrequisitos) != 1 || len(p.Verificacion) != 1 || len(p.Errores) != 0 {
		t.Errorf("prerrequisitos/verificación: %+v", p)
	}
	if len(p.Fuentes) != 1 || p.Fuentes[0].Manual != "Manual de Presupuestos" || !strings.Contains(p.Fuentes[0].Seccion, "4.2 Registro de metrados") {
		t.Errorf("fuentes por manual: %+v", p.Fuentes)
	}
	pasosAfirmados := map[int]bool{}
	for _, a := range p.Afirmaciones {
		if a.Manual != "Manual de Presupuestos" || a.Fragmento == "" {
			t.Errorf("afirmación sin par (id, manual): %+v", a)
		}
		pasosAfirmados[a.Paso] = true
	}
	for n := 1; n <= 3; n++ {
		if !pasosAfirmados[n] {
			t.Errorf("el paso %d no tiene afirmación → evidencia", n)
		}
	}
	if strings.Contains(txt, "**Presupuestos**") || strings.Contains(p.Intro, "**") {
		t.Error("el módulo no va en negrita: no sale de una fuente del elemento")
	}
	for _, s := range []string{"Vamos a registrar el metrado de una partida en el módulo Presupuestos.", "Antes de empezar", "1. Ubíquese", "2. Haga clic", "3. Registre", "Para comprobar que quedó bien", "_Fuente: Manual de Presupuestos › 4.2 Registro de metrados_"} {
		if !strings.Contains(txt, s) {
			t.Errorf("falta %q en:\n%s", s, txt)
		}
	}
	if strings.Contains(txt, "![") {
		t.Error("por defecto las fotos no van en el texto: las pinta la página desde el plan")
	}
}

func TestPlan_EntregaPorPartesYContinuar(t *testing.T) {
	en := nuevoEntorno(t)
	// Tamaño de parte: config.max_pasos_por_bloque del archivo de plantillas (3 en el de prueba).
	if en.c.TamanoParte() != 3 || TamanoParte(cargarFixtureEmbebidas(t)) != 6 {
		t.Errorf("tamaño de parte: %d (sin archivo: %d)", en.c.TamanoParte(), TamanoParte(cargarFixtureEmbebidas(t)))
	}
	p, _, txt := en.plan(t, "cómo modifico una partida", tipos.Procedimiento)
	if p.Procedimiento.ID != "presupuestos.modificar-partida" || len(p.Pasos) != 8 {
		t.Fatalf("plan: %+v", p.Procedimiento)
	}
	if got := p.PasosMostrados; len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Errorf("primera parte: %v", got)
	}
	if p.Siguiente == nil || p.Siguiente.Plantilla != "PREGUNTA_AVANCE" || p.Siguiente.Paso != 4 {
		t.Errorf("siguiente: %+v", p.Siguiente)
	}
	if !strings.Contains(txt, "Estos son los pasos 1 a 3 de 8") || strings.Contains(txt, "4. Elija") || !strings.Contains(txt, "paso 4") {
		t.Errorf("texto de la primera parte:\n%s", txt)
	}
	// Paso 7: la captura de la forma vieja es su foto; el 3 y el 5 sin foto.
	if f := fotosDe(p, 7); len(f) != 1 || f[0] != "d1fa4a0b710f" {
		t.Errorf("captura vieja: %v", f)
	}
	if len(fotosDe(p, 3)) != 0 || len(fotosDe(p, 5)) != 0 {
		t.Error("pasos sin foto en la fuente quedan sin foto")
	}
	// El paso 8 cita un PDF importado: la fuente sale con su título y su página.
	var guia *tipos.FuentePlan
	for i := range p.Fuentes {
		if p.Fuentes[i].Manual == "Guia de Usuario de S10 Presupuestos" {
			guia = &p.Fuentes[i]
		}
	}
	if guia == nil || len(guia.Paginas) != 1 || guia.Paginas[0] != 21 {
		t.Errorf("fuente del PDF importado: %+v", p.Fuentes)
	}

	// «¿y luego?»: la parte siguiente, sin buscar.
	m, err := en.c.Continuar(tipos.Memoria{ProcedimientoID: "presupuestos.modificar-partida", PasoActual: 3})
	if err != nil || len(m.PasosMostrados) != 3 || m.PasosMostrados[0] != 4 || m.Siguiente == nil || m.Siguiente.Paso != 7 {
		t.Fatalf("parte del medio: %+v %v", m.PasosMostrados, err)
	}
	q, err := en.c.Continuar(tipos.Memoria{ProcedimientoID: "presupuestos.modificar-partida", PasoActual: 6})
	if err != nil || len(q.PasosMostrados) != 2 || q.PasosMostrados[0] != 7 || q.Siguiente != nil || q.Plantilla != "PASOS_BLOQUE" || q.Intro != "" {
		t.Fatalf("continuar: %+v %v", q, err)
	}
	txt2, _ := en.gen.Redactar(context.Background(), q)
	if !strings.Contains(txt2, "pasos 7 a 8 de 8") || strings.Contains(txt2, "1. En la") {
		t.Errorf("segunda parte:\n%s", txt2)
	}
	if res := en.g.Evaluar(q, txt2, tipos.Estado{Tipo: tipos.Procedimiento}); !res.Paso {
		t.Errorf("gate segunda parte: %v", res.Problemas)
	}
	// Pasado el último: cierre.
	fin, err := en.c.Continuar(tipos.Memoria{ProcedimientoID: "presupuestos.modificar-partida", PasoActual: 8})
	if err != nil || fin.Plantilla != "PROCEDIMIENTO_FIN" || len(fin.PasosMostrados) != 0 {
		t.Errorf("fin: %+v %v", fin, err)
	}
	if _, err := en.c.Continuar(tipos.Memoria{ProcedimientoID: "no.existe"}); err == nil {
		t.Error("continuar un procedimiento inexistente es un error")
	}
}

func TestPlan_Concepto(t *testing.T) {
	en := nuevoEntorno(t)
	p, inf, txt := en.plan(t, "que es metrado", tipos.Concepto)
	if inf.Ruta != "concepto" || p.Concepto == nil || p.Concepto.ID != "metrado" || len(p.Pasos) != 0 {
		t.Fatalf("concepto: %+v", p)
	}
	if len(p.Opciones) != 1 || p.Opciones[0].ID != "presupuestos.registrar-metrado" || p.Siguiente == nil {
		t.Errorf("procedimiento relacionado para ofrecer: %+v %+v", p.Opciones, p.Siguiente)
	}
	if p.Procedimiento != nil {
		t.Error("un concepto no abre un procedimiento en curso")
	}
	// Una entrada por sección, cada una con sus pares (id, manual).
	if len(p.Fuentes) != 2 || p.Fuentes[0].Seccion != "1.2 Conceptos básicos" || p.Fuentes[1].Seccion != "4.2 Registro de metrados" ||
		len(p.Fuentes[0].Citas) != 1 || p.Fuentes[0].Citas[0].ID != s010 || p.Fuentes[0].Citas[0].Manual != "Manual de Presupuestos" {
		t.Errorf("fuentes: %+v", p.Fuentes)
	}
	// «Metrado» está en las fuentes del concepto: la plantilla lo pone en negrita.
	for _, s := range []string{"**Metrado**: Cuantificación", "En S10: Se registra", "¿Quiere que le enseñe a registrar el metrado de una partida?"} {
		if !strings.Contains(txt, s) {
			t.Errorf("falta %q:\n%s", s, txt)
		}
	}
	p, inf, _ = en.plan(t, "que es un algoritmo genetico", tipos.Concepto)
	if !p.SinEvidencia || inf.Ruta != "sin_evidencia" {
		t.Errorf("término fuera del glosario → SIN_EVIDENCIA: %+v", p)
	}
}

func TestPlan_Navegacion(t *testing.T) {
	en := nuevoEntorno(t)
	p, inf, txt := en.plan(t, "donde veo metrados", tipos.Navegacion)
	if inf.Ruta != "navegacion" || p.Procedimiento == nil || p.Procedimiento.ID != "presupuestos.registrar-metrado" {
		t.Fatalf("navegación: %+v %+v", inf, p)
	}
	// Ruta solo con los pasos que documentan dónde ocurren (campo «donde»).
	if len(p.Pasos) != 2 || p.Pasos[0].N != 1 || p.Pasos[1].N != 2 {
		t.Errorf("ruta: %+v", p.Pasos)
	}
	if !strings.Contains(p.Intro, "Escenario Hoja del presupuesto › Ventana Planilla de metrados") {
		t.Errorf("intro: %q", p.Intro)
	}
	if p.Concepto == nil || !strings.Contains(txt, "En S10: Se registra en la Planilla de metrados") {
		t.Errorf("el término nombrado aporta su «en_s10»:\n%s", txt)
	}
	// Procedimiento sin «donde» y cuyo primer paso no nombra pantalla: sin ruta → SIN_EVIDENCIA.
	pr, _ := LeerProcedimiento("x.yml", []byte(`id: presupuestos.x
titulo: "Hacer algo raro"
aliases: ["algo raro"]
pasos:
  - {n: 1, id: "presupuestos.x#1", accion: "Haga algo raro.", fuente: [`+s050+`]}
fuentes:
  - {id: `+s050+`, manual: "Manual de Presupuestos"}
`))
	b := NuevaBase([]*Procedimiento{pr}, nil, nil, en.b.Fragmentos)
	c := NuevoConstructor(b, NuevoRecuperador(b))
	q, _, _ := c.ConstruirConInforme(context.Background(), estado(b, "donde hago algo raro", tipos.Navegacion), nil)
	if !q.SinEvidencia {
		t.Errorf("sin ruta documentada → SIN_EVIDENCIA: %+v", q)
	}
}

func TestPlan_Problema(t *testing.T) {
	en := nuevoEntorno(t)
	p, inf, txt := en.plan(t, "no puedo guardar metrado", tipos.Problema)
	if inf.Ruta != "errores" || len(p.Errores) != 1 || p.Procedimiento.ID != "presupuestos.registrar-metrado" {
		t.Fatalf("errores: %+v %+v", inf, p)
	}
	if !strings.Contains(txt, "No se puede grabar el metrado.") || !strings.Contains(txt, "unidad de medida") {
		t.Errorf("texto:\n%s", txt)
	}
	// Procedimiento en curso sin errores documentados → SIN_EVIDENCIA, con el procedimiento.
	e := estado(en.b, "no me sale", tipos.Problema)
	e.Memoria = tipos.Memoria{ProcedimientoID: "presupuestos.modificar-partida", PasoActual: 2}
	q, _, _ := en.c.ConstruirConInforme(context.Background(), e, nil)
	if !q.SinEvidencia || q.Procedimiento == nil || q.Procedimiento.ID != "presupuestos.modificar-partida" {
		t.Errorf("sin errores_frecuentes → SIN_EVIDENCIA: %+v", q)
	}
	// Con memoria, los errores de OTRO procedimiento no cuentan.
	e = estado(en.b, "no puedo guardar metrado", tipos.Problema)
	e.Memoria = tipos.Memoria{ProcedimientoID: "presupuestos.configurar-datos-adicionales"}
	q, _, _ = en.c.ConstruirConInforme(context.Background(), e, nil)
	if !q.SinEvidencia || q.Procedimiento.ID != "presupuestos.configurar-datos-adicionales" {
		t.Errorf("memoria: %+v", q)
	}
}

func TestPlan_Configuracion(t *testing.T) {
	en := nuevoEntorno(t)
	p, _, txt := en.plan(t, "como configuro la formula polinomica del presupuesto", tipos.Configuracion)
	if p.Tipo != tipos.Configuracion || p.Procedimiento.ID != "presupuestos.configurar-datos-adicionales" {
		t.Fatalf("configuración: %+v", p.Procedimiento)
	}
	// El subpaso va dentro de su paso, sangrado, y su foto con el paso padre.
	if !strings.Contains(p.Pasos[1].Texto, "\n   2.1. Elija la cantidad de **decimales**.") {
		t.Errorf("subpaso: %q", p.Pasos[1].Texto)
	}
	if f := fotosDe(p, 2); len(f) != 1 || f[0] != "66c53a4c3194" {
		t.Errorf("foto del subpaso: %v", f)
	}
	if !strings.Contains(txt, "   2.1. Elija") {
		t.Errorf("texto:\n%s", txt)
	}
}

func TestPlan_Comparacion(t *testing.T) {
	en := nuevoEntorno(t)
	p, _, txt := en.plan(t, "diferencia entre presupuesto venta y presupuesto meta", tipos.Comparacion)
	if len(p.Conceptos) != 2 || p.Conceptos[0].ID != "presupuesto_venta" || p.Conceptos[1].ID != "presupuesto_meta" || p.Concepto.ID != "presupuesto_venta" {
		t.Fatalf("comparación (en el orden de la pregunta): %+v", p.Conceptos)
	}
	if !strings.Contains(txt, "Presupuesto venta: Presupuesto que se entrega") || !strings.Contains(txt, "Presupuesto meta: Presupuesto que sirve") {
		t.Errorf("texto:\n%s", txt)
	}
	q, _, _ := en.plan(t, "diferencia entre metrado y algoritmo", tipos.Comparacion)
	if !q.SinEvidencia {
		t.Errorf("un solo término del glosario → SIN_EVIDENCIA: %+v", q)
	}
}

func TestPlan_AclaracionYSinEvidencia(t *testing.T) {
	en := nuevoEntorno(t)
	p, _, txt := en.plan(t, "partida", tipos.Desconocido)
	if p.Tipo != tipos.Desconocido || len(p.Opciones) < 2 || len(p.Opciones) > 3 || p.Siguiente == nil || p.Siguiente.Plantilla != "ACLARAR_TAREA" {
		t.Fatalf("aclaración: %+v", p)
	}
	if !strings.Contains(txt, "modificar una partida del presupuesto") {
		t.Errorf("las opciones van en la pregunta:\n%s", txt)
	}
	p, inf, txt := en.plan(t, "como preparo un ceviche", tipos.Procedimiento)
	if !p.SinEvidencia || inf.Ruta != "sin_evidencia" || len(p.Fuentes) != 0 || p.Fuentes == nil {
		t.Errorf("sin evidencia: %+v %+v", p, inf)
	}
	if txt != "No encontré eso en los manuales que tengo." {
		t.Errorf("texto: %q", txt)
	}
	// Dos procedimientos que empatan → aclaración en vez de elegir a ciegas.
	mk := func(id, titulo string) *Procedimiento {
		pr, errs := LeerProcedimiento(id+".yml", []byte(`id: `+id+`
titulo: "`+titulo+`"
aliases: ["registrar cliente"]
preguntas: ["como registro un cliente"]
pasos:
  - {n: 1, id: "`+id+`#1", accion: "Pulse **Grabar**.", fuente: [`+s050+`]}
fuentes:
  - {id: `+s050+`, manual: "Manual de Presupuestos"}
`))
		if pr == nil {
			t.Fatal(errs)
		}
		return pr
	}
	b := NuevaBase([]*Procedimiento{mk("compras.registrar-cliente", "Registrar un cliente"), mk("facturacion.registrar-cliente", "Registrar un cliente nuevo")}, nil, nil, en.b.Fragmentos)
	c := NuevoConstructor(b, NuevoRecuperador(b))
	q, inf, _ := c.ConstruirConInforme(context.Background(), estado(b, "como registro un cliente", tipos.Procedimiento), nil)
	if q.Tipo != tipos.Desconocido || len(q.Opciones) != 2 || inf.Motivo != "dos procedimientos empatan" {
		t.Errorf("empate → aclaración: %+v %+v", q, inf)
	}
}

func TestPlan_ImplicitoDesdeFragmentos(t *testing.T) {
	en := nuevoEntorno(t)
	p, inf, txt := en.plan(t, "como registro una guia de remision", tipos.Procedimiento)
	if inf.Ruta != "implicito" || !EsImplicito(p.Procedimiento) || p.Plantilla != "PROCEDIMIENTO_IMPLICITO" {
		t.Fatalf("implícito: %+v %+v", inf, p.Procedimiento)
	}
	if p.Procedimiento.ID != IDImplicito(a020, "Manual de Almacenes") {
		t.Errorf("id del implícito: %s", p.Procedimiento.ID)
	}
	// Tres pasos con texto (el vacío se salta), cada uno con las fotos que la sección le asocia,
	// y una foto repetida en la sección no se repite en el plan.
	if len(p.Pasos) != 3 {
		t.Fatalf("pasos: %+v", p.Pasos)
	}
	if f := fotosDe(p, 1); len(f) != 1 || f[0] != "e03621f2f582" {
		t.Errorf("paso 1: %v", f)
	}
	if len(fotosDe(p, 2)) != 0 {
		t.Error("paso 2 sin foto")
	}
	if f := fotosDe(p, 3); len(f) != 1 || f[0] != "eba3e2d534d7" {
		t.Errorf("paso 3 (sin repetir la del paso 1): %v", f)
	}
	if !strings.Contains(txt, "No tengo una guía revisada") || !strings.Contains(txt, "Manual de Almacenes") {
		t.Errorf("el texto lo marca como implícito:\n%s", txt)
	}
	// La memoria del núcleo puede seguirlo: el catálogo lo resuelve por id.
	if d, ok := en.b.Procedimiento(p.Procedimiento.ID); !ok || len(d.Pasos) != 3 {
		t.Error("el catálogo no resuelve el procedimiento implícito")
	}
	if _, err := en.c.Continuar(tipos.Memoria{ProcedimientoID: p.Procedimiento.ID, PasoActual: 1}); err != nil {
		t.Errorf("continuar el implícito: %v", err)
	}
}

// Candidatos del núcleo: se usan en su orden; los de otro buscador se re-puntúan a esta escala.
func TestPlan_CandidatosDelNucleo(t *testing.T) {
	en := nuevoEntorno(t)
	e := estado(en.b, "como registro metrado", tipos.Procedimiento)
	ajenos := []tipos.Candidato{{ID: "presupuestos.registrar-metrado", Clase: ClaseProcedimiento, Puntaje: 0.016}}
	p, inf, err := en.c.ConstruirConInforme(context.Background(), e, ajenos)
	if err != nil || p.Procedimiento == nil || p.Procedimiento.ID != "presupuestos.registrar-metrado" || p.Procedimiento.Confianza < 0.6 {
		t.Errorf("candidato ajeno re-puntuado: %+v %+v %v", p.Procedimiento, inf, err)
	}
	// Un candidato ajeno que no cubre la consulta no pasa el umbral aunque venga primero.
	malos := []tipos.Candidato{{ID: "presupuestos.configurar-datos-adicionales", Clase: ClaseProcedimiento, Puntaje: 0.9}}
	p, _, _ = en.c.ConstruirConInforme(context.Background(), e, malos)
	if p.Procedimiento != nil && p.Procedimiento.ID == "presupuestos.configurar-datos-adicionales" {
		t.Error("eligió un candidato que no respalda la consulta")
	}
	if _, _, err := en.c.ConstruirConInforme(context.Background(), estado(en.b, "hola", tipos.Social), nil); err == nil {
		t.Error("SOCIAL no es de este constructor")
	}
}

// Asociación paso ↔ foto en TODOS los procedimientos reales: cada foto del plan es de su paso (o de un
// subpaso suyo) en el YAML, ninguna se repite y un paso sin foto en el YAML sale sin foto. Y el plan
// entero, redactado con las plantillas reales, pasa el gate.
func TestReal_PlanesYGate(t *testing.T) {
	b, r := recuperadorReal(t)
	c := NuevoConstructor(b, r)
	g := NuevoGate(b)
	m := NuevoMotorPlantillas(b)
	m.FotosEnTexto = true
	pasos, conFoto, partes := 0, 0, 0
	for _, pr := range b.Procedimientos {
		for desde := 1; desde <= len(pr.Pasos); desde += c.pasosPorParte() {
			pl := c.PlanProcedimiento(pr, tipos.Procedimiento, desde, 1)
			partes++
			vistas := map[string]bool{}
			for i, x := range pl.Pasos {
				def := pr.Pasos[i]
				permitidas := map[string]bool{}
				var rec func(tipos.Paso)
				rec = func(s tipos.Paso) {
					for _, f := range s.Fotos {
						permitidas[f.ID+"|"+f.Ruta] = true
					}
					for _, sub := range s.Sub {
						rec(sub)
					}
				}
				rec(def)
				if len(permitidas) == 0 && len(x.Fotos) > 0 {
					t.Errorf("%s: el paso %d no tiene fotos en el YAML y el plan le pone %d", pr.ID, x.N, len(x.Fotos))
				}
				for _, f := range x.Fotos {
					if !permitidas[f.ID+"|"+f.Ruta] {
						t.Errorf("%s: la foto %s no es del paso %d", pr.ID, f.ID, x.N)
					}
					if vistas[f.ID] {
						t.Errorf("%s: la foto %s se repite", pr.ID, f.ID)
					}
					vistas[f.ID] = true
				}
				if desde == 1 {
					pasos++
					if len(x.Fotos) > 0 {
						conFoto++
					}
				}
			}
			txt, err := m.Redactar(context.Background(), pl)
			if err != nil {
				t.Fatalf("%s: %v", pr.ID, err)
			}
			if res := g.Evaluar(pl, txt, tipos.Estado{Tipo: tipos.Procedimiento}); !res.Paso {
				t.Errorf("%s (desde %d): el gate rechaza el plan real: %v", pr.ID, desde, res.Problemas)
			}
		}
	}
	t.Logf("%d procedimientos, %d partes, %d pasos, %d con foto (%.0f%%)", len(b.Procedimientos), partes, pasos, conFoto, 100*float64(conFoto)/float64(max(pasos, 1)))
}

// Prerrequisitos solo con la primera parte; verificación solo con la que termina el procedimiento.
func TestPlan_PrerrequisitosYVerificacionPorParte(t *testing.T) {
	en := nuevoEntorno(t)
	c := *en.c
	c.MaxSinPartes, c.PasosPorParte = 2, 2
	p, _ := en.b.Proc("presupuestos.registrar-metrado")
	a := c.PlanProcedimiento(p, tipos.Procedimiento, 1, 1)
	if len(a.PasosMostrados) != 2 || len(a.Prerrequisitos) != 1 || len(a.Verificacion) != 0 {
		t.Errorf("primera parte: mostrados %v, prerrequisitos %d, verificación %d", a.PasosMostrados, len(a.Prerrequisitos), len(a.Verificacion))
	}
	b, _ := c.Continuar(tipos.Memoria{ProcedimientoID: p.ID, PasoActual: 2})
	if len(b.PasosMostrados) != 1 || len(b.Prerrequisitos) != 0 || len(b.Verificacion) != 1 {
		t.Errorf("última parte: mostrados %v, prerrequisitos %d, verificación %d", b.PasosMostrados, len(b.Prerrequisitos), len(b.Verificacion))
	}
	for _, x := range []tipos.Plan{a, b} {
		txt, _ := en.gen.Redactar(context.Background(), x)
		if res := en.g.Evaluar(x, txt, tipos.Estado{Tipo: tipos.Procedimiento}); !res.Paso {
			t.Errorf("gate: %v", res.Problemas)
		}
	}
}

// Comparar dos formas del mismo concepto («recurso» / «insumo»): el glosario los da como sinónimos.
func TestPlan_ComparacionDeSinonimos(t *testing.T) {
	en := nuevoEntorno(t)
	q, inf2, txt2 := en.plan(t, "que es un presupuesto meta, es lo mismo que meta?", tipos.Comparacion)
	if q.Tipo != tipos.Concepto || q.Concepto == nil || q.Concepto.ID != "presupuesto_meta" || !strings.Contains(q.Intro, "sinónimo") {
		t.Errorf("sinónimos → concepto: %+v %+v", q, inf2)
	}
	if !strings.Contains(txt2, "figura como sinónimo de «Presupuesto meta»") {
		t.Errorf("texto:\n%s", txt2)
	}
}
