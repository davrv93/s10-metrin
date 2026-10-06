package busqueda

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"rag-go/internal/indexar"
)

func TestLeerYAMLSubconjunto(t *testing.T) {
	src := `# comentario
id: presupuestos.registrar   # comentario al final
version: 1
titulo: "Registrar un presupuesto \"nuevo\""
preguntas:
  - "cómo creo un presupuesto"
  - como registro un prespuesto
vacia: []
pasos:
  - n: 1
    accion: "Ingrese a **Datos Generales**: escenario"
    fuente: [web-a-s032, "7c40a9548710-0018"]
    sub:
      - n: 1
        accion: "Marque **Fórmula polinómica**."
  - n: 2
    accion: 'Pulse # Adicionar'
fuentes:
  - {id: web-a-s032, manual: "Manual de Presupuestos", pagina: null}
lista_al_mismo_nivel:
- a
- b
revision: {generado: 2026-10-06, por: claude, revisado_por_humano: false}
`
	v, err := LeerYAML([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	m := yamlMapa(v)
	if m["id"] != "presupuestos.registrar" || m["version"] != "1" {
		t.Errorf("escalares: %v %v", m["id"], m["version"])
	}
	if m["titulo"] != `Registrar un presupuesto "nuevo"` {
		t.Errorf("titulo = %q", m["titulo"])
	}
	if !reflect.DeepEqual(yamlTextos(m["preguntas"]), []string{"cómo creo un presupuesto", "como registro un prespuesto"}) {
		t.Errorf("preguntas = %v", m["preguntas"])
	}
	if l, ok := m["vacia"].([]any); !ok || len(l) != 0 {
		t.Errorf("vacia = %#v", m["vacia"])
	}
	pasos := yamlLista(m["pasos"])
	if len(pasos) != 2 {
		t.Fatalf("pasos = %#v", pasos)
	}
	p1 := yamlMapa(pasos[0])
	if p1["accion"] != "Ingrese a **Datos Generales**: escenario" {
		t.Errorf("accion = %q", p1["accion"])
	}
	if !reflect.DeepEqual(yamlTextos(p1["fuente"]), []string{"web-a-s032", "7c40a9548710-0018"}) {
		t.Errorf("fuente = %#v", p1["fuente"])
	}
	if sub := yamlLista(p1["sub"]); len(sub) != 1 || yamlMapa(sub[0])["accion"] != "Marque **Fórmula polinómica**." {
		t.Errorf("sub = %#v", p1["sub"])
	}
	if yamlMapa(pasos[1])["accion"] != "Pulse # Adicionar" {
		t.Errorf("comillas simples con #: %#v", pasos[1])
	}
	f := yamlMapa(yamlLista(m["fuentes"])[0])
	if f["manual"] != "Manual de Presupuestos" || f["pagina"] != nil {
		t.Errorf("mapa en línea = %#v", f)
	}
	if !reflect.DeepEqual(yamlTextos(m["lista_al_mismo_nivel"]), []string{"a", "b"}) {
		t.Errorf("lista al mismo nivel = %#v", m["lista_al_mismo_nivel"])
	}
	if yamlMapa(m["revision"])["revisado_por_humano"] != "false" {
		t.Errorf("revision = %#v", m["revision"])
	}
	for _, malo := range []string{"a: |\n  texto\n", "a: [1, 2\n", "a: &ancla x\n"} {
		if _, err := LeerYAML([]byte(malo)); err == nil {
			t.Errorf("se esperaba error con %q", malo)
		}
	}
}

// Los ids de CargarFragmentos tienen que ser los mismos que pone
// indexar.KBJSONL (el índice vectorial), o la fusión RRF no casa documentos.
func TestCargarFragmentosMismosIDsQueIndexar(t *testing.T) {
	dir := t.TempDir()
	a := `{"id":"web-x-s001","manual":"Manual de Presupuestos","titulo":"Manual de Presupuestos › 3.3 Registro","seccion":"3.3 Registro","texto":"Ingrese a Datos generales.","confianza":"oficial"}
{"id":"web-x-s001","manual":"Manual de Almacenes","titulo":"Manual de Almacenes › 2 Kardex","seccion":"2 Kardex","texto":"El kardex.","confianza":"oficial"}
{"id":"web-x-s001","manual":"Manual de Presupuestos","titulo":"Manual de Presupuestos › 3.3 Registro","seccion":"3.3 Registro","texto":"Ingrese a Datos generales.","confianza":"oficial"}

{"id":"img-1-0000","tipo":"imagen","manual":"Manual de Presupuestos","titulo":"1.png","texto":"Imagen del manual S10: 1.png. Texto reconocido por OCR: Metrado","confianza":"oficial"}
`
	b := `{"id":"7c40-0018","documento":"7c40","manual":"Importado a mano","titulo":"Guia de Usuario de S10 Presupuestos","pagina":11,"texto":"Iniciando el registro del nuevo presupuesto","confianza":"tercero-sin-verificar"}
{"id":"web-x-s001","manual":"Manual de Compras","seccion":"1 Orden","texto":"Orden de compra.","confianza":"oficial"}
`
	pa := filepath.Join(dir, "fragmentos.jsonl")
	pb := filepath.Join(dir, "fragmentos_z.jsonl")
	os.WriteFile(pa, []byte(a), 0o644)
	os.WriteFile(pb, []byte(b), 0o644)
	rutas, _ := RutasFragmentos(dir)
	docs, err := CargarFragmentos(rutas...)
	if err != nil {
		t.Fatal(err)
	}
	var nuestros []string
	for _, d := range docs {
		nuestros = append(nuestros, d.ID)
	}
	// indexar lee un único archivo: el cat de docker-entrada.sh.
	cat := filepath.Join(dir, "kb.jsonl")
	os.WriteFile(cat, []byte(a+b), 0o644)
	ds, err := indexar.KBJSONL{Ruta: cat}.Listar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var suyos []string
	for _, d := range ds {
		suyos = append(suyos, strings.TrimSuffix(d.Ruta, ".txt"))
	}
	if !reflect.DeepEqual(nuestros, suyos) {
		t.Errorf("ids distintos:\n nuestros %v\n indexar  %v", nuestros, suyos)
	}
	// indexar solo salta una línea idéntica si es la última versión vista de
	// ese id; aquí la repetida llega después de otra versión y entra como
	// documento aparte (con sufijo). Se replica tal cual.
	if len(docs) != len(ds) || len(docs) != 6 {
		t.Errorf("docs = %d, indexar = %d, quiere 6", len(docs), len(ds))
	}
	// Campos: la imagen va a ocr, la página del PDF manda su título a manual.
	por := map[string]Doc{}
	for _, d := range docs {
		por[d.Meta["id_original"]+"|"+d.Meta["manual"]] = d
	}
	img := por["img-1-0000|Manual de Presupuestos"]
	if img.Campos["ocr"] == "" || img.Campos["texto"] != "" || img.Campos["titulo"] != "" {
		t.Errorf("imagen: %#v", img.Campos)
	}
	pdf := por["7c40-0018|Importado a mano"]
	if pdf.Campos["titulo"] != "" || !strings.Contains(pdf.Campos["manual"], "Guia de Usuario") || pdf.Meta["pagina"] != "11" {
		t.Errorf("pdf: %#v %#v", pdf.Campos, pdf.Meta)
	}
	sec := por["web-x-s001|Manual de Presupuestos"]
	if sec.Campos["seccion"] != "3.3 Registro" || sec.Campos["titulo"] != "" {
		t.Errorf("sección (el título repetido no se cuenta dos veces): %#v", sec.Campos)
	}
}

func TestCargarProcedimientos(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "presupuestos"), 0o755)
	os.WriteFile(filepath.Join(dir, "ESQUEMA.md"), []byte("# no es yaml"), 0o644)
	os.MkdirAll(filepath.Join(dir, "_reserva", "compras"), 0o755)
	os.WriteFile(filepath.Join(dir, "_reserva", "compras", "x.yml"), []byte("id: compras.apartado\ntitulo: \"No cargar\"\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "presupuestos", "presupuestos.registrar.yml"), []byte(`id: presupuestos.registrar
modulo: presupuestos
titulo: "Registrar un presupuesto nuevo"
objetivo: "Registrar los datos generales."
preguntas:
  - "cómo creo un presupuesto nuevo"
aliases: ["alta de presupuesto"]
entidades:
  modulo: "Presupuestos"
  pantallas: ["Datos Generales", "Catálogo de Presupuestos"]
  objetos: ["presupuesto", "subpresupuesto"]
pasos:
  - n: 1
    accion: "Ingrese al escenario de **Datos Generales**."
    fuente: [web-a-s032]
    sub:
      - n: 1
        accion: "Pulse **Adicionar**."
        fuente: [img-1-0000]
`), 0o644)
	docs, err := CargarProcedimientos(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("docs = %d (la carpeta _reserva no se carga)", len(docs))
	}
	d := docs[0]
	if d.ID != "proc:presupuestos.registrar" || d.Meta["tipo"] != "procedimiento" || d.Meta["fuentes"] != "web-a-s032,img-1-0000" {
		t.Errorf("doc = %#v", d)
	}
	if !strings.Contains(d.Campos["preguntas"], "alta de presupuesto") || !strings.Contains(d.Campos["texto"], "Pulse Adicionar.") ||
		!strings.Contains(d.Campos["entidades"], "Catálogo de Presupuestos") || d.Meta["pantallas"] != "Datos Generales|Catálogo de Presupuestos" {
		t.Errorf("campos = %#v", d.Campos)
	}
	if ds, err := CargarProcedimientos(filepath.Join(dir, "no-existe")); err != nil || ds != nil {
		t.Errorf("directorio ausente: %v %v", ds, err)
	}
	// Un índice mixto (fragmentos + procedimientos) los encuentra por sus
	// preguntas.
	ix := indicePrueba(t, ConfigPorDefecto(), d, soloTexto("frag", "Kardex de almacén"))
	if got := ix.BuscarLexico("como doy de alta un presupuesto", 1); len(got) != 1 || got[0].ID != d.ID {
		t.Errorf("búsqueda de procedimiento: %v", got)
	}
}

func TestReescribirConservaOriginalYExpande(t *testing.T) {
	dir := t.TempDir()
	alias := filepath.Join(dir, "alias.yml")
	os.WriteFile(alias, []byte(`acciones:
  - [registrar, ingresar, cargar, "dar de alta"]
terminos:
  - {termino: "análisis de precios unitarios", sinonimos: [apu, "costo unitario"]}
  - {termino: metrado, sinonimos: ["cantidad de obra"]}
`), 0o644)
	procs := filepath.Join(dir, "procedimientos", "presupuestos")
	os.MkdirAll(procs, 0o755)
	os.WriteFile(filepath.Join(procs, "p.yml"), []byte(`id: presupuestos.metrados
titulo: "Registrar los metrados de las partidas"
entidades:
  pantallas: ["Hoja del Presupuesto"]
  objetos: ["partida", "metrado"]
preguntas:
  - "como cargo los metrados de una partida"
`), 0o644)
	a := NuevoAlias()
	if err := a.CargarAlias(alias); err != nil {
		t.Fatal(err)
	}
	if err := a.AgregarProcedimientos(filepath.Join(dir, "procedimientos")); err != nil {
		t.Fatal(err)
	}
	if err := a.AgregarConceptos(filepath.Join(dir, "conceptos")); err != nil {
		t.Fatalf("conceptos ausente no es error: %v", err)
	}
	c := a.Reescribir("¿Cómo hago para cargar un metrado en la partida?")
	if c.Original != "¿Cómo hago para cargar un metrado en la partida?" || c.Normalizada != "cómo hago para cargar un metrado en la partida" {
		t.Errorf("original/normalizada: %+v", c)
	}
	// «metrado» viene de alias.yml y «partida» de las entidades del
	// procedimiento (sin duplicar «metrado»).
	if !reflect.DeepEqual(c.Entidades, []string{"metrado", "partida"}) {
		t.Errorf("entidades = %v", c.Entidades)
	}
	quiere := []string{"Registrar los metrados de las partidas", "cantidad de obra", "dar de alta", "ingresar", "registrar"}
	sort.Strings(quiere)
	if !reflect.DeepEqual(c.Aliases, quiere) {
		t.Errorf("aliases = %q\n quiere %q", c.Aliases, quiere)
	}
	// Los términos de la original siguen con peso 1 en BM25.
	ix := indicePrueba(t, ConfigPorDefecto(), soloTexto("x", "metrado"))
	terms := ix.TerminosConsulta(c, 0.3)
	pesos := map[string]float64{}
	for _, tc := range terms {
		pesos[tc.Termino] = tc.Peso
	}
	if pesos["metrad"] != 1 || pesos["carg"] != 1 || pesos["ingres"] != 0.3 {
		t.Errorf("pesos = %v", pesos)
	}
	// Atajos y comillas son entidades aunque no haya diccionario.
	var nada *Alias
	c = nada.Reescribir(`¿qué hace F7 en «Hoja del Presupuesto»? y Ctrl+F`)
	if !reflect.DeepEqual(c.Entidades, []string{"f7", "ctrl+f", "Hoja del Presupuesto"}) || len(c.Aliases) != 0 {
		t.Errorf("sin diccionario: %+v", c)
	}
	// APU como sinónimo del término largo.
	c = a.Reescribir("como edito el apu")
	if !reflect.DeepEqual(c.Entidades, []string{"análisis de precios unitarios"}) {
		t.Errorf("apu: %+v", c)
	}
}
