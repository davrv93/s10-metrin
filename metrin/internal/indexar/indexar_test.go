package indexar

import (
	"context"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rag-go/internal/almacen"
)

// bolsa es un embebedor de prueba determinista: bolsa de palabras hasheada.
type bolsa struct{}

func (bolsa) Nombre() string { return "bolsa" }
func (bolsa) Embeber(_ context.Context, t string) ([]float32, error) {
	v := make([]float32, 64)
	for _, w := range strings.Fields(strings.ToLower(t)) {
		h := fnv.New32a()
		h.Write([]byte(strings.Trim(w, ".,¿?¡!:;")))
		v[h.Sum32()%64]++
	}
	var s float64
	for _, x := range v {
		s += float64(x * x)
	}
	if s == 0 {
		v[0] = 1
		return v, nil
	}
	for i := range v {
		v[i] = float32(float64(v[i]) / math.Sqrt(s))
	}
	return v, nil
}

// origenFijo es un origen en memoria: clave → (versión, contenido).
type origenFijo struct {
	fuente string
	docs   map[string][2]string
	cargas *int
}

func (o origenFijo) Fuente() string                 { return o.fuente }
func (o origenFijo) Cita(d Documento, n int) string { return d.Clave + "#" + itoa(n) }
func (o origenFijo) Listar(context.Context) ([]Documento, error) {
	var out []Documento
	for k, vc := range o.docs {
		c := vc[1]
		out = append(out, Documento{Clave: k, Ruta: strings.SplitN(k, ":", 2)[1], Version: vc[0],
			Cargar: func(context.Context) (string, error) { *o.cargas++; return c, nil }})
	}
	return out, nil
}

func escribir(t *testing.T, raiz, rel, contenido string) {
	t.Helper()
	p := filepath.Join(raiz, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestIDsEstables(t *testing.T) {
	a := almacen.IDTrozo("repo:index.html", 0)
	if a != almacen.IDTrozo("repo:index.html", 0) {
		t.Fatal("el id no es estable")
	}
	// sha1("repo:index.html:0")
	if a != "6d02d50e63874899be1cf3c5f2609d4595f2924e" {
		t.Fatalf("id = %s, esperaba sha1(\"repo:index.html:0\")", a)
	}
	distintos := map[string]bool{
		a: true, almacen.IDTrozo("repo:index.html", 1): true,
		almacen.IDTrozo("s3:index.html", 0): true, almacen.IDTrozo("repo:index.htm", 0): true,
	}
	if len(distintos) != 4 {
		t.Fatal("ids repetidos entre claves o trozos distintos")
	}
}

func TestIncrementalRepo(t *testing.T) {
	ctx := context.Background()
	raiz := t.TempDir()
	escribir(t, raiz, "index.html", "<h1>Hola</h1><script>x()</script><p>formulario de contacto</p>")
	escribir(t, raiz, "README.md", strings.Repeat("documentación del proyecto ", 60)) // > 800 runas
	escribir(t, raiz, "src/app.js", "const endpoint = '/api/contacto';")
	escribir(t, raiz, "node_modules/lib/index.js", "no me indexes")
	escribir(t, raiz, "dist/bundle.js", "tampoco")
	escribir(t, raiz, ".git/config", "[core]")
	escribir(t, raiz, "logo.png", "\x89PNG")

	dirDatos := t.TempDir()
	a, err := almacen.Abrir(dirDatos, bolsa{})
	if err != nil {
		t.Fatal(err)
	}
	inf, err := Ejecutar(ctx, Repo{Raiz: raiz}, a, nil)
	if err != nil {
		t.Fatal(err)
	}
	if inf.Nuevos != 3 || inf.Vistos != 3 {
		t.Fatalf("primer indexado: %s (esperaba 3 nuevos, sin node_modules/dist/.git/png)", inf)
	}
	total := a.Contar()
	if total < 4 { // README tiene 2+ trozos
		t.Fatalf("trozos=%d", total)
	}

	// Segunda pasada, reabriendo desde disco: nada cambia, nada se reescribe.
	a, err = almacen.Abrir(dirDatos, bolsa{})
	if err != nil {
		t.Fatal(err)
	}
	if a.Contar() != total {
		t.Fatalf("tras reabrir hay %d trozos, esperaba %d", a.Contar(), total)
	}
	inf, _ = Ejecutar(ctx, Repo{Raiz: raiz}, a, nil)
	if inf.Iguales != 3 || inf.Trozos != 0 || inf.Nuevos+inf.Cambiados+inf.Borrados != 0 {
		t.Fatalf("reindexado sin cambios reescribió algo: %s", inf)
	}

	// Cambia uno (y encoge de 2+ trozos a 1) y borra otro.
	escribir(t, raiz, "README.md", "corto")
	os.Remove(filepath.Join(raiz, "src/app.js"))
	inf, _ = Ejecutar(ctx, Repo{Raiz: raiz}, a, nil)
	if inf.Cambiados != 1 || inf.Borrados != 1 || inf.Iguales != 1 {
		t.Fatalf("esperaba 1 cambiado, 1 borrado, 1 igual: %s", inf)
	}
	// index.html (1) + README (1) = 2: no quedan trozos huérfanos.
	if a.Contar() != 2 {
		t.Fatalf("quedan %d trozos, esperaba 2 (sin huérfanos)", a.Contar())
	}
	rs, _ := a.Buscar(ctx, "endpoint api contacto", 5, nil)
	for _, r := range rs {
		if r.Metadata["path"] == "src/app.js" {
			t.Fatal("el documento borrado sigue saliendo en la búsqueda")
		}
	}
}

func TestIncrementalNoCargaLoIgual(t *testing.T) {
	ctx := context.Background()
	a, _ := almacen.Abrir("", bolsa{})
	cargas := 0
	o := origenFijo{fuente: almacen.FuenteS3, cargas: &cargas, docs: map[string][2]string{
		"s3:docs/a.md": {"etag-1", "precio del plan pro"},
		"s3:docs/b.md": {"etag-2", "política de privacidad"},
	}}
	Ejecutar(ctx, o, a, nil)
	if cargas != 2 {
		t.Fatalf("cargas=%d", cargas)
	}
	Ejecutar(ctx, o, a, nil)
	if cargas != 2 {
		t.Fatalf("con el mismo ETag no debe descargarse nada; cargas=%d", cargas)
	}
	o.docs["s3:docs/a.md"] = [2]string{"etag-3", "precio nuevo"}
	inf, _ := Ejecutar(ctx, o, a, nil)
	if cargas != 3 || inf.Cambiados != 1 {
		t.Fatalf("cargas=%d %s", cargas, inf)
	}
}

func TestKBJSONLUsaTituloYEnlaceDeYouTube(t *testing.T) {
	dir := t.TempDir()
	escribir(t, dir, "videos.jsonl", `{"id":"yt-abcdefghijk-0000","documento":"yt-abcdefghijk","video":"¿Qué es Optimiza 360?","url":"https://youtu.be/abcdefghijk?t=5","desde":"0:05","texto":"Presentación de la empresa."}`)
	docs, err := (KBJSONL{Ruta: filepath.Join(dir, "videos.jsonl")}).Listar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("recuperé %d documentos; quería 1", len(docs))
	}
	if got := docs[0].Metadata["title"]; got != "¿Qué es Optimiza 360?" {
		t.Errorf("title = %q", got)
	}
	if got := docs[0].Metadata["cita"]; got != "¿Qué es Optimiza 360?, 0:05" {
		t.Errorf("cita = %q", got)
	}
	if got := docs[0].Metadata["source_url"]; got != "https://youtu.be/abcdefghijk?t=5" {
		t.Errorf("source_url = %q", got)
	}
}

func TestBorradoNoCruzaFuentes(t *testing.T) {
	ctx := context.Background()
	a, _ := almacen.Abrir("", bolsa{})
	n := 0
	repo := origenFijo{fuente: almacen.FuenteRepo, cargas: &n, docs: map[string][2]string{"repo:a.md": {"1", "hola repo"}}}
	s3 := origenFijo{fuente: almacen.FuenteS3, cargas: &n, docs: map[string][2]string{"s3:a.md": {"1", "hola s3"}}}
	Ejecutar(ctx, repo, a, nil)
	Ejecutar(ctx, s3, a, nil)
	s3.docs = map[string][2]string{}
	inf, _ := Ejecutar(ctx, s3, a, nil)
	if inf.Borrados != 1 || a.Contar() != 1 {
		t.Fatalf("vaciar S3 debe borrar solo lo de S3: %s, quedan %d", inf, a.Contar())
	}
}

func TestBusquedaConFiltro(t *testing.T) {
	ctx := context.Background()
	a, _ := almacen.Abrir("", bolsa{})
	n := 0
	Ejecutar(ctx, origenFijo{fuente: almacen.FuenteRepo, cargas: &n, docs: map[string][2]string{
		"repo:index.html": {"1", "<p>plan pro precio en la hoja de tarifas</p>"},
		"repo:app.js":     {"1", "formulario de contacto validación"},
	}}, a, nil)
	Ejecutar(ctx, origenFijo{fuente: almacen.FuenteS3, cargas: &n, docs: map[string][2]string{
		"s3:docs/tarifas.md": {"1", "plan pro precio 49 euros al mes"},
	}}, a, nil)

	rs, err := a.Buscar(ctx, "precio plan pro", 8, map[string]string{"source": almacen.FuenteS3})
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 1 || rs[0].Metadata["source"] != almacen.FuenteS3 || rs[0].Metadata["cita"] != "s3:docs/tarifas.md#0" {
		t.Fatalf("filtro source=garage-s3: %+v", rs)
	}
	rs, _ = a.Buscar(ctx, "precio plan pro", 8, map[string]string{"source": almacen.FuenteRepo, "ext": ".js"})
	if len(rs) != 1 || rs[0].Metadata["path"] != "app.js" {
		t.Fatalf("filtro source+ext: %+v", rs)
	}
	rs, _ = a.Buscar(ctx, "precio plan pro", 8, map[string]string{"type": "html"})
	if len(rs) != 1 || rs[0].Metadata["path"] != "index.html" {
		t.Fatalf("filtro type=html: %+v", rs)
	}
	// k mayor que el índice no rompe (chromem exige k <= total).
	rs, _ = a.Buscar(ctx, "precio plan pro", 50, nil)
	if len(rs) != 3 {
		t.Fatalf("sin filtro con k=50: %d resultados", len(rs))
	}
	if rs[0].Distancia > rs[len(rs)-1].Distancia {
		t.Fatal("resultados no ordenados por distancia")
	}
}
