package conocimiento

import (
	"strings"
	"testing"

	"rag-go/internal/v2/tipos"
)

const (
	s050 = "web-https-documentacion-s10peru-com-manual-d-s050"
	s010 = "web-https-documentacion-s10peru-com-manual-d-s010"
	a020 = "web-https-documentacion-s10peru-com-manual-a-s020"
)

// cargarFixture: testdata/kb con las plantillas de prueba (formato del archivo real).
func cargarFixture(t *testing.T) *Base {
	t.Helper()
	b := Cargar(Opciones{DirKB: "testdata/kb", ArchivoPlantillas: "testdata/plantillas/respuestas.yml"})
	return b
}

// cargarFixtureEmbebidas: testdata/kb sin archivo de plantillas (solo las embebidas).
func cargarFixtureEmbebidas(t *testing.T) *Base {
	t.Helper()
	return Cargar(Opciones{DirKB: "testdata/kb", ArchivoPlantillas: "testdata/no-existe.yml"})
}

func TestCargar_FixturesValidos(t *testing.T) {
	b := cargarFixture(t)
	for _, e := range b.Errores {
		// La única queja esperada es la plantilla ROTA del archivo de prueba.
		if !strings.Contains(e.Mensaje, "ROTA") {
			t.Errorf("error inesperado: %s", e.Error())
		}
	}
	if len(b.Procedimientos) != 3 {
		t.Fatalf("procedimientos: %d, quiero 3 (la carpeta _reserva no se carga)", len(b.Procedimientos))
	}
	if _, ok := b.Proc("presupuestos.borrador"); ok {
		t.Error("cargó un procedimiento de _reserva")
	}
	if len(b.Conceptos) != 4 {
		t.Fatalf("conceptos: %d, quiero 4 (metrado, partida, presupuesto_meta, presupuesto_venta)", len(b.Conceptos))
	}
	if b.Fragmentos.Total() != 10 {
		t.Errorf("fragmentos: %d, quiero 10", b.Fragmentos.Total())
	}
	p, ok := b.Proc("presupuestos.registrar-metrado")
	if !ok {
		t.Fatal("falta presupuestos.registrar-metrado")
	}
	if p.Entidades.Modulo != "Presupuestos" || len(p.Entidades.Pantallas) != 2 || len(p.Entidades.Objetos) != 2 {
		t.Errorf("entidades mal leídas: %+v", p.Entidades)
	}
	if len(p.Pasos) != 3 || p.Pasos[0].Donde == "" || len(p.Pasos[0].Fotos) != 1 || len(p.Pasos[2].Fotos) != 0 {
		t.Errorf("pasos mal leídos: %+v", p.Pasos)
	}
	// Catálogo del núcleo: definición por id.
	if d, ok := b.Procedimiento("presupuestos.registrar-metrado"); !ok || d.Titulo != p.Titulo {
		t.Error("Base no cumple Catalogo.Procedimiento")
	}
	if c, ok := b.Concepto("metrado"); !ok || c.Termino != "Metrado" {
		t.Error("Base no cumple Catalogo.Concepto")
	}
}

func TestCargar_Reserva(t *testing.T) {
	b := Cargar(Opciones{DirKB: "testdata/kb", ArchivoPlantillas: "testdata/no-existe.yml", IncluirReserva: true})
	if _, ok := b.Proc("presupuestos.borrador"); !ok {
		t.Error("con IncluirReserva debe cargar _reserva/")
	}
}

func TestCargar_Invalidos(t *testing.T) {
	b := Cargar(Opciones{DirKB: "testdata/invalidos", ArchivosFragmentos: []string{"testdata/kb/fragmentos.jsonl"}, ArchivoPlantillas: "testdata/no-existe.yml"})
	// Lo válido carga; lo inválido se salta con archivo y línea.
	ids := map[string]bool{}
	for _, p := range b.Procedimientos {
		ids[p.ID] = true
	}
	if !ids["presupuestos.repetido"] || !ids["presupuestos.ambiguo"] || len(ids) != 2 {
		t.Errorf("procedimientos cargados: %v; quiero repetido (el original) y ambiguo", ids)
	}
	if p, _ := b.Proc("presupuestos.repetido"); p == nil || p.Titulo != "Original" {
		t.Error("con id repetido se queda el primero (a-original.yml)")
	}
	buscar := func(archivo, texto string, grave bool, conLinea bool) {
		t.Helper()
		for _, e := range b.Errores {
			if strings.HasSuffix(e.Archivo, archivo) && strings.Contains(e.Mensaje, texto) && e.Grave == grave {
				if conLinea && e.Linea == 0 {
					t.Errorf("%s: falta la línea en %q", archivo, e.Error())
				}
				return
			}
		}
		t.Errorf("no se informó %q en %s (grave=%v). Errores:\n%v", texto, archivo, grave, b.Errores)
	}
	buscar("roto.yml", "YAML inválido", true, true)
	buscar("sin-pasos.yml", "sin «pasos»", true, true)
	buscar("n-saltado.yml", "n=3 donde tocaba n=2", true, true)
	buscar("tipo-malo.yml", "YAML inválido", true, true)
	buscar("b-duplicado.yml", "repetido", true, true)
	buscar("presupuestos.ambiguo.yml", "dos manuales", false, true)
	buscar("presupuestos.ambiguo.yml", "no está en la tabla", false, true)
	buscar("presupuestos.ambiguo.yml", "ruta no válida", false, true)
	buscar("presupuestos.ambiguo.yml", "distinto de sha1", false, true)
	buscar("presupuestos.ambiguo.yml", "ninguna fuente del paso la asocia", false, true)
	// Conceptos.
	buscar("conceptos/roto.yml", "YAML inválido", true, true)
	buscar("sin-definicion.yml", "sin «termino» o sin «definicion»", true, true)
	buscar("sin-manual.yml", "no dice el manual", false, true)
	if _, ok := b.Conc("unico"); !ok {
		t.Error("el concepto con un id de un solo manual debe cargar")
	} else if c, _ := b.Conc("unico"); len(c.Fuente) != 1 {
		t.Error("unico: falta la fuente")
	} else if ci, err := b.CitaConcepto(c, a020); err != nil || ci.Manual != "Manual de Almacenes" {
		t.Errorf("un id de un solo manual se resuelve: %+v %v", ci, err)
	}
	// Ninguno tumba nada: el error es texto con archivo:línea.
	for _, e := range b.Graves() {
		if !strings.Contains(e.Error(), ":") {
			t.Errorf("error sin archivo: %v", e)
		}
	}
}

func TestIdsAmbiguos_ParIdManual(t *testing.T) {
	b := cargarFixture(t)
	if n := len(b.Fragmentos.PorID(s050)); n != 2 {
		t.Fatalf("el id %s debe estar en 2 manuales; está en %d", s050, n)
	}
	p, _ := b.Proc("presupuestos.registrar-metrado")
	c, err := b.CitaProcedimiento(p, s050)
	if err != nil || c.Manual != "Manual de Presupuestos" || c.Fragmento == nil || !strings.Contains(c.Fragmento.Texto, "metrado") {
		t.Fatalf("la cita debe resolverse con el manual de la tabla: %+v %v", c, err)
	}
	alm, err := b.CitaPar(s050, "Manual de Almacenes")
	if err != nil || alm.Fragmento == c.Fragmento {
		t.Error("el mismo id en otro manual es otro fragmento")
	}
	if _, err := b.CitaPar(s050, "Manual de Nóminas"); err == nil {
		t.Error("un par (id, manual) inexistente no se resuelve aunque el id exista")
	}
	if _, err := b.CitaProcedimiento(p, "no-esta-en-la-tabla"); err == nil {
		t.Error("una cita fuera de la tabla no se resuelve")
	}
	// Concepto: el manual sale del comentario de la línea (s010 está en Presupuestos y en Compras).
	m, _ := b.Conc("metrado")
	ci, err := b.CitaConcepto(m, s010)
	if err != nil || ci.Manual != "Manual de Presupuestos" || !strings.Contains(ci.Fragmento.Texto, "cuantificación") {
		t.Errorf("cita del concepto por comentario: %+v %v", ci, err)
	}
	// Tabla «fuentes» en el concepto y {id, manual} en línea.
	for _, id := range []string{"partida", "presupuesto_meta", "presupuesto_venta"} {
		cc, _ := b.Conc(id)
		if ci, err := b.CitaConcepto(cc, s010); err != nil || ci.Manual != "Manual de Presupuestos" {
			t.Errorf("%s: %+v %v", id, ci, err)
		}
	}
}

func TestCargar_CapturaViejaYFotos(t *testing.T) {
	b := cargarFixture(t)
	p, _ := b.Proc("presupuestos.modificar-partida")
	s := p.Pasos[6]
	if len(s.Fotos) != 1 || s.Fotos[0].Ruta != "imagenes/manual-de-presupuestos/ddd007ddd007.png" || s.Fotos[0].ID != idFoto(s.Fotos[0].Ruta) {
		t.Errorf("captura (forma vieja) → foto con id sha1: %+v", s.Fotos)
	}
	c, _ := b.Proc("presupuestos.configurar-datos-adicionales")
	if len(c.Pasos[1].Sub) != 1 || c.Pasos[1].Sub[0].ID != "presupuestos.configurar-datos-adicionales#2.1" || len(c.Pasos[1].Sub[0].Fotos) != 1 {
		t.Errorf("subpasos mal leídos: %+v", c.Pasos[1])
	}
}

func TestPlantillas_FormaPartesRespaldo(t *testing.T) {
	ps, errs := CargarPlantillas("testdata/plantillas/respuestas.yml")
	if ps.MaxPasosBloque != 3 {
		t.Errorf("config.max_pasos_por_bloque: %d", ps.MaxPasosBloque)
	}
	rota := false
	for _, e := range errs {
		if strings.Contains(e.Mensaje, "ROTA") && e.Linea > 0 {
			rota = true
		}
	}
	if !rota {
		t.Errorf("la plantilla ROTA (parte sin variantes) debe informarse con su línea: %v", errs)
	}
	// Variante elegible: la primera con todos sus slots.
	s := ps.R("PROCEDIMIENTO_INTRO", map[string]string{"TAREA": "registrar el metrado", "MODULO": "Presupuestos", "TOTAL_PASOS": "3"},
		OpcRelleno{Omitir: []string{"arranque", "cita", "prerrequisitos"}})
	// La plantilla pone «**{{MODULO}}**», pero sin respaldo (Negrita nil) el módulo va sin negrita.
	if s != "Vamos a registrar el metrado en el módulo Presupuestos. Son 3 pasos." {
		t.Errorf("intro: %q", s)
	}
	s = ps.R("PROCEDIMIENTO_INTRO", map[string]string{"TAREA": "registrar el metrado", "MODULO": "Presupuestos", "TOTAL_PASOS": "3"},
		OpcRelleno{Omitir: []string{"arranque", "cita", "prerrequisitos"}, Negrita: func(slot, v string) bool { return slot == "MODULO" }})
	if !strings.Contains(s, "**Presupuestos**") {
		t.Errorf("con respaldo, la negrita se conserva: %q", s)
	}
	// Sin MODULO, la primera variante no es elegible: se usa la segunda.
	s = ps.R("PROCEDIMIENTO_INTRO", map[string]string{"TAREA": "registrar el metrado", "TOTAL_PASOS": "3"}, OpcRelleno{Omitir: []string{"arranque", "cita"}})
	if !strings.HasPrefix(s, "Le enseño a registrar el metrado.") || strings.Contains(s, "{{") {
		t.Errorf("variante elegible: %q", s)
	}
	// Respaldo con slot: CONCEPTO sin TAREA_SUGERIDA usa el respaldo de «oferta».
	s = ps.R("CONCEPTO", map[string]string{"TERMINO": "Metrado", "DEFINICION": "x.", "CITA": "Manual de Presupuestos"})
	if !strings.Contains(s, "¿Le queda alguna duda sobre Metrado?") || strings.Contains(s, "**") || strings.Contains(s, "En S10") {
		t.Errorf("respaldo: %q", s)
	}
	// Sin CITA (parte obligatoria del archivo) → la embebida, cuyas líneas «?» son opcionales.
	if s = ps.R("CONCEPTO", map[string]string{"TERMINO": "Metrado", "DEFINICION": "x."}); s != "Metrado: x." {
		t.Errorf("embebida con líneas opcionales: %q", s)
	}
	// Semilla: elige entre variantes elegibles de forma determinista.
	a := ps.R("PREGUNTA_AVANCE", map[string]string{"N_SIGUIENTE": "4"})
	b := ps.R("PREGUNTA_AVANCE", map[string]string{"N_SIGUIENTE": "4"}, OpcRelleno{Semilla: 1})
	if a == b || !strings.Contains(b, "paso 4") {
		t.Errorf("semilla: %q / %q", a, b)
	}
	// Plantilla que el archivo no trae: la embebida.
	if s := ps.R("COMPROBACION", map[string]string{"VERIFICACION_FINAL": "- ok"}); !strings.HasPrefix(s, "Para comprobar") {
		t.Errorf("embebida: %q", s)
	}
	// Una parte obligatoria sin dato → no se puede rellenar (ni la embebida) → false.
	if _, ok := ps.Rellenar("PASOS_BLOQUE", map[string]string{}); ok {
		t.Error("PASOS_BLOQUE sin PASOS no se puede rellenar")
	}
	// Sin archivo: solo embebidas, sin error.
	pe, errs := CargarPlantillas("testdata/no-existe.yml")
	if len(errs) != 0 || pe.Archivo != "" || pe.R("SIN_EVIDENCIA", nil) == "" {
		t.Errorf("sin archivo: %v", errs)
	}
}

func TestPlantillas_ArchivoReal(t *testing.T) {
	ps, errs := CargarPlantillas("../../../plantillas/respuestas.yml")
	if ps.Archivo == "" {
		t.Skip("sin metrin/plantillas/respuestas.yml")
	}
	for _, e := range errs {
		t.Errorf("plantillas reales: %s", e.Error())
	}
	for _, n := range []string{"PROCEDIMIENTO_INTRO", "PASOS_BLOQUE", "PREGUNTA_AVANCE", "CONCEPTO", "COMPARACION", "ERROR_FRECUENTE", "NAVEGACION", "ACLARAR_TAREA", "SIN_EVIDENCIA", "PROCEDIMIENTO_FIN"} {
		if p, ok := ps.Get(n); !ok || p.Origen != "archivo" {
			t.Errorf("falta %s en el archivo real", n)
		}
	}
}

func TestAliaser(t *testing.T) {
	b := cargarFixture(t)
	ents, alias, modulo := b.Analizar("como registro metrado en la hoja del presupuesto")
	if !contieneStr(ents, "Metrado") || !contieneStr(ents, "Hoja del presupuesto") {
		t.Errorf("entidades: %v", ents)
	}
	if !contieneStr(alias, "metrados") {
		t.Errorf("aliases (sinónimos del glosario): %v", alias)
	}
	if modulo != "presupuestos" {
		t.Errorf("módulo: %q", modulo)
	}
	_, alias, _ = b.Analizar("quiero registrar metrado")
	if !contieneStr(alias, "Registrar el metrado de una partida") {
		t.Errorf("alias de procedimiento → título: %v", alias)
	}
	c := b.Expandir(tipos.Consulta{Original: "que es un presupuesto meta"})
	if c.Original != "que es un presupuesto meta" || !contieneStr(c.Entidades, "Presupuesto meta") {
		t.Errorf("Expandir no debe tocar la original y debe sumar entidades: %+v", c)
	}
	if contieneStr(c.Entidades, "Metrado") {
		t.Error("detectó un término que no está")
	}
}

func contieneStr(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}
