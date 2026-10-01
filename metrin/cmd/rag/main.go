// Comando rag: RAG local en Go sobre un repo y un bucket S3 (Garage).
//
//	rag index-repo <ruta>
//	rag index-s3 [--prefix docs/]
//	rag index-kb <fragmentos.jsonl>
//	rag subir-s3 <dir> [--prefix docs/]      siembra el bucket de ejemplo
//	rag ask "<pregunta>" [--source landing-repo|garage-s3] [--type T] [--ext .md] [--k 8] [--json]
//	rag serve [--addr :4760]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"rag-go/internal/almacen"
	"rag-go/internal/clasificar"
	"rag-go/internal/config"
	"rag-go/internal/embed"
	"rag-go/internal/indexar"
	"rag-go/internal/jev"
	"rag-go/internal/llm"
	"rag-go/internal/rag"
	"rag-go/internal/servidor"
)

const uso = `uso:
  rag index-repo <ruta>
  rag index-s3 [--prefix docs/]
  rag index-kb <fragmentos.jsonl>
  rag subir-s3 <dir> [--prefix docs/]
  rag ask "<pregunta>" [--source landing-repo|garage-s3] [--type doc|html|codigo|estilo|datos] [--ext .md] [--k 8] [--json] [--solo-buscar]
  rag serve [--addr :4760]
  rag entrenar-clasificador [--datos dir] [--salida f] [--umbral 0.40]
  rag clasificar "<texto>" [--modelo f] [--json]

Configuración por entorno o .env (ver .env.example).`

func main() {
	if err := config.CargarEnv(".env"); err != nil {
		fallar(err)
	}
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, uso)
		os.Exit(2)
	}
	cfg := config.Leer()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cmd, args := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "index-repo":
		err = cmdIndexRepo(ctx, cfg, args)
	case "index-s3":
		err = cmdIndexS3(ctx, cfg, args)
	case "index-kb":
		err = cmdIndexKB(ctx, cfg, args)
	case "subir-s3":
		err = cmdSubirS3(ctx, cfg, args)
	case "ask":
		err = cmdAsk(ctx, cfg, args)
	case "reporte":
		err = cmdReporte(ctx, cfg, args)
	case "serve":
		err = cmdServe(ctx, cfg, args)
	case "entrenar-clasificador":
		err = cmdEntrenarClasificador(ctx, cfg, args)
	case "clasificar":
		err = cmdClasificar(ctx, cfg, args)
	case "-h", "--help", "help":
		fmt.Println(uso)
	default:
		fmt.Fprintf(os.Stderr, "subcomando desconocido %q\n%s\n", cmd, uso)
		os.Exit(2)
	}
	if err != nil {
		fallar(err)
	}
}

func fallar(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

// separar deja que los flags vayan antes o después de los posicionales
// (flag de la stdlib se detiene en el primer posicional).
func separar(fs *flag.FlagSet, args []string) ([]string, error) {
	var flags, pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		if strings.HasPrefix(a, "-") && len(a) > 1 {
			flags = append(flags, a)
			nombre := strings.TrimLeft(a, "-")
			if strings.Contains(nombre, "=") {
				continue
			}
			if f := fs.Lookup(nombre); f != nil {
				if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
					continue
				}
				if i+1 < len(args) {
					flags = append(flags, args[i+1])
					i++
				}
			}
			continue
		}
		pos = append(pos, a)
	}
	return pos, fs.Parse(flags)
}

// nuevoClienteJEV devuelve nil sin JEV_URL: las decisiones
// JEV son opt-in. Con JEV_TRAZAS, cada decisión se graba en
// el JSONL que lee el visor de trazas (TRAZAS_REPO=archivo).
func nuevoClienteJEV(cfg config.Config) *jev.Cliente {
	if cfg.JEVURL == "" {
		return nil
	}
	c := jev.Nuevo(cfg.JEVURL, cfg.JEVModelo, time.Duration(cfg.TimeoutSeg)*time.Second)
	if cfg.JEVTrazas != "" {
		c.Trazas = &jev.RegistroTrazas{Ruta: cfg.JEVTrazas, Origen: cfg.JEVOrigen}
	}
	return c
}

func nuevoEmbebedor(cfg config.Config) (embed.Embebedor, error) {
	switch cfg.EmbedProvider {
	case "estatico":
		return embed.NuevoEstatico(cfg.ModeloEstatico)
	case "ollama":
		return embed.NuevoOllama(cfg.OllamaURL, cfg.EmbedModelo), nil
	default:
		return nil, fmt.Errorf("EMBED_PROVIDER=%q: usa estatico u ollama", cfg.EmbedProvider)
	}
}

func abrirAlmacen(cfg config.Config) (*almacen.Almacen, error) {
	e, err := nuevoEmbebedor(cfg)
	if err != nil {
		return nil, err
	}
	return almacen.Abrir(cfg.DirDatos, e)
}

func logf(f string, a ...any) { fmt.Printf(f+"\n", a...) }

func cmdIndexRepo(ctx context.Context, cfg config.Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("uso: rag index-repo <ruta>")
	}
	t0 := time.Now()
	a, err := abrirAlmacen(cfg)
	if err != nil {
		return err
	}
	tCarga := time.Since(t0)
	inf, err := indexar.Ejecutar(ctx, indexar.Repo{Raiz: args[0]}, a, logf)
	if err != nil {
		return err
	}
	logf("repo %s: %s | trozos en índice=%d | carga modelo %s, total %s", args[0], inf, a.Contar(), tCarga.Round(time.Millisecond), time.Since(t0).Round(time.Millisecond))
	return nil
}

func cmdIndexKB(ctx context.Context, cfg config.Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("uso: rag index-kb <fragmentos.jsonl>")
	}
	t0 := time.Now()
	a, err := abrirAlmacen(cfg)
	if err != nil {
		return err
	}
	inf, err := indexar.Ejecutar(ctx, indexar.KBJSONL{Ruta: args[0]}, a, logf)
	if err != nil {
		return err
	}
	logf("S10 KB: %s | trozos en índice=%d | total %s", inf, a.Contar(), time.Since(t0).Round(time.Millisecond))
	return nil
}

func clienteS3(cfg config.Config, prefijo string) (indexar.S3, error) {
	c, err := indexar.NuevoClienteS3(indexar.ConfigS3{
		Endpoint: cfg.S3Endpoint, Region: cfg.S3Region,
		AccessKey: cfg.S3AccessKey, SecretKey: cfg.S3SecretKey, Bucket: cfg.S3Bucket,
	})
	if err != nil {
		return indexar.S3{}, err
	}
	return indexar.S3{Cliente: c, Bucket: cfg.S3Bucket, Prefijo: prefijo}, nil
}

func cmdIndexS3(ctx context.Context, cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("index-s3", flag.ContinueOnError)
	prefijo := fs.String("prefix", "", "prefijo de las keys (p. ej. docs/)")
	if _, err := separar(fs, args); err != nil {
		return err
	}
	t0 := time.Now()
	a, err := abrirAlmacen(cfg)
	if err != nil {
		return err
	}
	o, err := clienteS3(cfg, *prefijo)
	if err != nil {
		return err
	}
	inf, err := indexar.Ejecutar(ctx, o, a, logf)
	if err != nil {
		return err
	}
	logf("s3://%s/%s: %s | trozos en índice=%d | total %s", cfg.S3Bucket, *prefijo, inf, a.Contar(), time.Since(t0).Round(time.Millisecond))
	return nil
}

func cmdSubirS3(ctx context.Context, cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("subir-s3", flag.ContinueOnError)
	prefijo := fs.String("prefix", "docs/", "prefijo destino")
	pos, err := separar(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return fmt.Errorf("uso: rag subir-s3 <dir> [--prefix docs/]")
	}
	o, err := clienteS3(cfg, *prefijo)
	if err != nil {
		return err
	}
	ents, err := os.ReadDir(pos[0])
	if err != nil {
		return err
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(pos[0], e.Name()))
		if err != nil {
			return err
		}
		tipo := mime.TypeByExtension(filepath.Ext(e.Name()))
		if tipo == "" {
			tipo = "text/plain; charset=utf-8"
		}
		if err := o.Subir(ctx, *prefijo+e.Name(), b, tipo); err != nil {
			return fmt.Errorf("subir %s: %w", e.Name(), err)
		}
		logf("  ↑ s3://%s/%s%s (%d B)", cfg.S3Bucket, *prefijo, e.Name(), len(b))
	}
	return nil
}

func nuevoRAG(cfg config.Config) (*rag.RAG, error) {
	a, err := abrirAlmacen(cfg)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(cfg.TimeoutSeg) * time.Second
	var conversacional rag.Chateador
	switch cfg.LLMProvider {
	case "mlx":
		conversacional = llm.NuevoMLX(cfg.MLXURL, cfg.LLMModelo, cfg.Temperatura, timeout)
	case "ollama", "":
		conversacional = llm.NuevoOllama(cfg.OllamaURL, cfg.LLMModelo, cfg.Temperatura, timeout)
	default:
		return nil, fmt.Errorf("LLM_PROVIDER=%q: usa ollama o mlx", cfg.LLMProvider)
	}
	e, err := nuevoEmbebedor(cfg)
	if err != nil {
		return nil, err
	}
	clas, err := clasificar.CargarDefecto()
	if err != nil {
		logf("sin clasificador (%v): todo va al RAG", err)
	} else if clas.HuellaEmb != "" && clas.HuellaEmb != e.Nombre() {
		logf("clasificador incompatible (%s != %s): se desactiva y todo va al RAG", clas.HuellaEmb, e.Nombre())
		clas = nil
	}
	return &rag.RAG{
		Almacen:      a,
		LLM:          conversacional,
		Emb:          e,
		Clasificador: clas,
		MaxDistancia: cfg.MaxDistancia,
		DistanciaSemilla: cfg.DistanciaSemilla,
		RutaFallos:   filepath.Join(cfg.DirDatos, "sin_respuesta.jsonl"),
	}, nil
}

func cmdAsk(ctx context.Context, cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("ask", flag.ContinueOnError)
	source := fs.String("source", "", "landing-repo | garage-s3")
	tipo := fs.String("type", "", "doc | html | codigo | estilo | datos")
	ext := fs.String("ext", "", "extensión, p. ej. .md")
	k := fs.Int("k", 8, "trozos a recuperar")
	comoJSON := fs.Bool("json", false, "salida JSON")
	soloBuscar := fs.Bool("solo-buscar", false, "muestra los trozos recuperados sin llamar al LLM")
	pos, err := separar(fs, args)
	if err != nil {
		return err
	}
	if len(pos) == 0 {
		return fmt.Errorf(`uso: rag ask "<pregunta>" [--source …] [--k 8]`)
	}
	r, err := nuevoRAG(cfg)
	if err != nil {
		return err
	}
	filtro := map[string]string{}
	for clave, v := range map[string]string{"source": *source, "type": *tipo, "ext": *ext} {
		if v != "" {
			filtro[clave] = v
		}
	}
	pregunta := strings.Join(pos, " ")
	if *soloBuscar {
		t0 := time.Now()
		trozos, err := r.Almacen.Buscar(ctx, pregunta, *k, filtro)
		if err != nil {
			return err
		}
		for _, t := range trozos {
			fmt.Printf("d=%.3f  %-40s %s\n", t.Distancia, t.Metadata["cita"], recortar(t.Texto, 70))
		}
		fmt.Printf("(%d trozos en %s)\n", len(trozos), time.Since(t0).Round(time.Microsecond))
		return nil
	}
	res, err := r.Preguntar(ctx, pregunta, rag.Opciones{K: *k, Filtro: filtro})
	if err != nil {
		return err
	}
	if *comoJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	}
	fmt.Println(res.Respuesta)
	fmt.Println()
	fmt.Println("Fuentes recuperadas (distancia coseno):")
	for _, f := range res.Fuentes {
		fmt.Printf("  %-45s d=%.3f\n", f.Cita, f.Distancia)
	}
	if res.SinContexto {
		fmt.Printf("\n[sin contexto: %s → registrada en %s]\n", res.Motivo, r.RutaFallos)
	}
	fmt.Printf("\nTiempos: búsqueda %d ms, LLM %d ms\n", res.MsBusqueda, res.MsLLM)
	return nil
}

func recortar(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

func cmdServe(ctx context.Context, cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:"+cfg.Puerto, "dirección de escucha")
	if _, err := separar(fs, args); err != nil {
		return err
	}
	r, err := nuevoRAG(cfg)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:              *addr,
		Handler:           servidor.Nuevo(r, time.Duration(cfg.TimeoutSeg)*time.Second),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(c)
	}()
	logf("escuchando en http://%s (trozos en índice: %d)", *addr, r.Almacen.Contar())
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
