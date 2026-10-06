// Comando evalbusqueda mide la recuperación de Metrín sobre un conjunto dorado
// (eval/busqueda_oro.jsonl): Recall@5, Recall@10, MRR@10, nDCG@10 y latencia
// p50/p95 de vector solo (como hoy), BM25, híbrida RRF y híbrida + reranker.
//
//	cd metrin
//	go run ./cmd/evalbusqueda \
//	  -kb ../kb -oro eval/busqueda_oro.jsonl -alias eval/alias.yml \
//	  -rerank-url http://127.0.0.1:8091 -rerank-pid $(cat llama.pid)
//
// Otros modos: -barrido (k1, b, peso exacto, k de RRF, peso de alias, N del
// reranker), -pool archivo (candidatos para juzgar a mano), -casos.
// El índice vectorial se construye en memoria con el mismo indexador de
// producción (indexar.KBJSONL sobre fragmentos*.jsonl concatenados).
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"rag-go/internal/almacen"
	"rag-go/internal/busqueda"
	"rag-go/internal/embed"
	"rag-go/internal/indexar"
	"rag-go/internal/rerank"
)

type relevante struct {
	ID     string `json:"id"`
	Manual string `json:"manual"`
	Grado  int    `json:"grado"`
	Nota   string `json:"nota,omitempty"`
}

type casoOro struct {
	ID         string      `json:"id"`
	Consulta   string      `json:"consulta"`
	Tipo       string      `json:"tipo"`
	Sintetico  bool        `json:"sintetico"`
	Origen     string      `json:"origen"`
	Relevantes []relevante `json:"relevantes"`
	rel        map[string]int
}

type metodo struct {
	nombre string
	buscar func(ctx context.Context, q string, k int) ([]string, error)
}

type medida struct {
	r5, r10, mrr, ndcg float64
	lat                []time.Duration
	porTipo            map[string][]float64 // tipo → [r5, r10, mrr, ndcg, n]
}

func main() {
	var (
		dirKB       = flag.String("kb", "../kb", "directorio con fragmentos*.jsonl")
		rutaOro     = flag.String("oro", "eval/busqueda_oro.jsonl", "conjunto dorado")
		rutaAlias   = flag.String("alias", "eval/alias.yml", "alias de consulta (vacío = sin alias)")
		dirProcs    = flag.String("procedimientos", "", "kb/procedimientos: alias de sus preguntas (vacío = <kb>/procedimientos)")
		indexarProc = flag.Bool("indexar-procedimientos", false, "añadir los procedimientos al índice léxico (el oro solo tiene fragmentos)")
		aliasProc   = flag.Bool("alias-procedimientos", true, "usar aliases/preguntas/entidades de kb/procedimientos como alias")
		modelo      = flag.String("modelo", "modelos/potion-es-int8.pjge", "embebedor estático")
		urlRerank   = flag.String("rerank-url", "", "llama-server --reranking (vacío = sin reranker)")
		pidRerank   = flag.Int("rerank-pid", 0, "pid de llama-server para medir su RSS")
		k1          = flag.Float64("k1", 1.2, "BM25 k1")
		b           = flag.Float64("b", 0.75, "BM25 b")
		pesoExacto  = flag.Float64("peso-exacto", 0, "peso del término exacto")
		difMinDF    = flag.Int("difuso-min-df", 0, "df por debajo del cual se corrige una raíz (0 = defecto)")
		difFactor   = flag.Int("difuso-factor", 0, "cuántas veces más frecuente debe ser la corrección (0 = defecto)")
		rrfK        = flag.Float64("rrfk", 60, "k de RRF")
		pesoVector  = flag.Float64("peso-vector", 1, "peso de la lista vectorial en la RRF (la léxica pesa 1)")
		pesoVecSR   = flag.Float64("peso-vector-sin-rerank", -1, "peso del vector en el orden final sin reranker (-1 = igual que -peso-vector)")
		repesRR     = flag.Int("repeticiones-rerank", 1, "repeticiones de los métodos con reranker")
		soloMetodos = flag.String("metodos", "", "nombres exactos de los métodos a evaluar, separados por | (vacío = todos)")
		pesoAlias   = flag.Float64("peso-alias", 0.3, "peso de los alias en BM25")
		topRerank   = flag.Int("top-rerank", 30, "candidatos que pasan por el reranker")
		maxRunas    = flag.Int("rerank-runas", 1200, "runas por documento enviadas al reranker")
		barrido     = flag.Bool("barrido", false, "barrer k1, b, peso exacto, k de RRF, peso de alias y N del reranker")
		rutaPool    = flag.String("pool", "", "escribe aquí los candidatos de cada consulta para juzgar")
		profPool    = flag.Int("pool-k", 15, "profundidad del pool por método")
		casos       = flag.String("casos", "¿qué es un metrado?|¿Cómo registro un nuevo presupuesto en S10?", "consultas a detallar (separadas por |)")
		rutaJSON    = flag.String("json", "", "escribe aquí el detalle por consulta")
		repes       = flag.Int("repeticiones", 3, "veces que se repite cada consulta para la latencia")
		variantesRR = flag.String("rerank-variantes", "", "variantes del reranker N:runas separadas por coma (p. ej. 30:1200,20:800); método «rerank N=… runas=…»")
		cascada     = flag.String("cascada", "", "umbrales del margen relativo (s1−s2)/s1 del orden RRF sin reranker, separados por coma: solo se reordena si el margen es menor; método «cascada margen<…»")
		cacheRR     = flag.Bool("rerank-cache", false, "reutiliza el puntaje de cada par (consulta, documento recortado) entre métodos y variantes: el cross-encoder puntúa cada par por separado, así que la calidad no cambia; la latencia de lo que sale de la caché NO vale")
		cascadaTip  = flag.String("cascada-tipos", "", "tipos del oro que se reordenan siempre en la cascada (p. ej. como|que_es); con -cascada vacío, método «cascada tipos»")
	)
	flag.Parse()
	ctx := context.Background()
	logf := func(f string, a ...any) { fmt.Fprintf(os.Stderr, f+"\n", a...) }

	// 1. Documentos.
	rutas, err := busqueda.RutasFragmentos(*dirKB)
	fallar(err)
	if len(rutas) == 0 {
		fallar(fmt.Errorf("no hay fragmentos*.jsonl en %s", *dirKB))
	}
	docs, err := busqueda.CargarFragmentos(rutas...)
	fallar(err)
	{
		porOriginal, porPar := map[string]int{}, map[string]int{}
		for _, d := range docs {
			porOriginal[d.Meta["id_original"]]++
			porPar[d.Meta["clave"]]++
		}
		rep, repPar := 0, 0
		for _, n := range porOriginal {
			if n > 1 {
				rep++
			}
		}
		for _, n := range porPar {
			if n > 1 {
				repPar++
			}
		}
		logf("fragmentos: %d de %d archivos; ids originales repetidos entre manuales: %d; pares (id, manual) repetidos: %d",
			len(docs), len(rutas), rep, repPar)
	}
	if *dirProcs == "" {
		*dirProcs = filepath.Join(*dirKB, "procedimientos")
	}
	if *indexarProc {
		ps, err := busqueda.CargarProcedimientos(*dirProcs)
		fallar(err)
		docs = append(docs, ps...)
		logf("procedimientos indexados: %d", len(ps))
	}

	// 2. Índice BM25 y su RAM (montículo vivo antes y después, tras GC).
	cfg := busqueda.ConfigPorDefecto()
	cfg.K1, cfg.B, cfg.PesoExacto = *k1, *b, *pesoExacto
	if *difMinDF > 0 {
		cfg.MinDFDifuso = *difMinDF
	}
	if *difFactor > 0 {
		cfg.FactorDifuso = *difFactor
	}
	heapAntes := heapVivo()
	t0 := time.Now()
	lex, err := busqueda.IndexarDocs(cfg, docs)
	fallar(err)
	tIndice := time.Since(t0)
	heapIndice := heapVivo() - heapAntes
	tmpGob := filepath.Join(os.TempDir(), "evalbusqueda-bm25.gob")
	fallar(lex.GuardarArchivo(tmpGob))
	fi, _ := os.Stat(tmpGob)
	t0 = time.Now()
	if _, err := busqueda.CargarIndiceArchivo(tmpGob); err != nil {
		fallar(err)
	}
	tCargaGob := time.Since(t0)
	os.Remove(tmpGob)
	logf("BM25: %d docs, %d términos, %.1f MiB de montículo, indexado en %v, gob %.1f MiB (carga %v)",
		lex.Contar(), lex.Vocabulario(), mib(heapIndice), tIndice.Round(time.Millisecond), mib(uint64(fi.Size())), tCargaGob.Round(time.Millisecond))

	// 3. Índice vectorial (en memoria, mismo indexador que producción).
	emb, err := embed.NuevoEstatico(*modelo)
	fallar(err)
	alm, err := almacen.Abrir("", emb)
	fallar(err)
	cat, err := concatenar(rutas)
	fallar(err)
	t0 = time.Now()
	inf, err := indexar.Ejecutar(ctx, indexar.KBJSONL{Ruta: cat}, alm, nil)
	os.Remove(cat)
	fallar(err)
	logf("vector: %d trozos (%s) en %v", alm.Contar(), inf, time.Since(t0).Round(time.Millisecond))
	vec := busqueda.VectorAlmacen{A: alm}

	// 4. Alias.
	var al *busqueda.Alias
	if *rutaAlias != "" {
		al = busqueda.NuevoAlias()
		fallar(al.CargarAlias(*rutaAlias))
		if *aliasProc {
			fallar(al.AgregarProcedimientos(*dirProcs))
		}
		fallar(al.AgregarConceptos(filepath.Join(*dirKB, "conceptos")))
		logf("alias: %v", al.Origenes)
	}

	// 5. Oro.
	oro, err := cargarOro(*rutaOro, docs, *rutaPool != "")
	fallar(err)
	logf("oro: %d consultas", len(oro))

	// 6. Reranker.
	var rr busqueda.Reordenador
	var cachePares *cachePuntajes
	if *cacheRR {
		cachePares = &cachePuntajes{m: map[string]float64{}}
	}
	if *urlRerank != "" {
		c := rerank.NuevoLlamaServer(*urlRerank)
		c.MaxRunas = *maxRunas
		rr = cachePares.envolver(c, *maxRunas)
	}

	opc := busqueda.OpcionesPorDefecto()
	opc.RRFK, opc.PesoAlias, opc.TopRerank, opc.TamCache = *rrfK, *pesoAlias, *topRerank, 0
	opc.PesoVector, opc.PesoLexico = *pesoVector, 1
	opc.PesoVectorSinRerank = *pesoVecSR
	if *pesoVecSR < 0 {
		opc.PesoVectorSinRerank = *pesoVector
	}
	opc.TimeoutRerank = 0 // en la evaluación se mide el reranker entero
	nuevo := func(modo busqueda.Modo, expandir bool, conRerank bool) *busqueda.Buscador {
		o := opc
		o.Modo, o.Expandir = modo, expandir
		var r busqueda.Reordenador
		if conRerank {
			r = rr
		}
		return busqueda.NuevoBuscador(lex, vec, r, al, o)
	}
	desde := func(bu *busqueda.Buscador) func(context.Context, string, int) ([]string, error) {
		return func(ctx context.Context, q string, k int) ([]string, error) {
			inf, err := bu.BuscarInforme(ctx, q, k)
			if err == nil && len(inf.Degradado) > 0 {
				err = fmt.Errorf("degradado %v: %s %s", inf.Degradado, inf.FalloVector, inf.FalloRerank)
			}
			return idsDe(inf.Resultados), err
		}
	}
	metodos := []metodo{
		{"vector (hoy)", func(ctx context.Context, q string, k int) ([]string, error) {
			// Lo que hace rag.go hoy: almacen.Buscar con la pregunta tal cual;
			// los trozos se agrupan por fragmento para medir a nivel de fragmento.
			cs, err := vec.BuscarVector(ctx, q, k)
			out := make([]string, len(cs))
			for i, c := range cs {
				out[i] = c.ID
			}
			return out, err
		}},
		{"BM25", desde(nuevo(busqueda.ModoLexico, false, false))},
		{"BM25 + alias", desde(nuevo(busqueda.ModoLexico, true, false))},
		{"híbrida RRF", desde(nuevo(busqueda.ModoHibrido, false, false))},
		{"híbrida RRF + alias", desde(nuevo(busqueda.ModoHibrido, true, false))},
	}
	if rr != nil {
		metodos = append(metodos,
			metodo{fmt.Sprintf("híbrida + rerank (top %d)", *topRerank), desde(nuevo(busqueda.ModoHibrido, false, true))},
			metodo{fmt.Sprintf("híbrida + alias + rerank (top %d)", *topRerank), desde(nuevo(busqueda.ModoHibrido, true, true))},
			metodo{fmt.Sprintf("BM25 + rerank (top %d)", *topRerank), desde(nuevo(busqueda.ModoLexico, false, true))},
		)
	}

	// Variantes del reranker (N candidatos × recorte por documento) y cascada.
	var cascadas []*cascadaRR
	if rr != nil {
		for _, v := range listaVariantes(*variantesRR) {
			c := rerank.NuevoLlamaServer(*urlRerank)
			c.MaxRunas = v[1]
			o := opc
			o.Modo, o.Expandir, o.TopRerank = busqueda.ModoHibrido, false, v[0]
			metodos = append(metodos, metodo{fmt.Sprintf("rerank N=%d runas=%d", v[0], v[1]), desde(busqueda.NuevoBuscador(lex, vec, cachePares.envolver(c, v[1]), al, o))})
		}
		tipoDe := map[string]string{}
		for _, c := range oro {
			tipoDe[c.Consulta] = c.Tipo
		}
		tipos := map[string]bool{}
		for _, t := range strings.Split(*cascadaTip, "|") {
			if t = strings.TrimSpace(t); t != "" {
				tipos[t] = true
			}
		}
		// Variantes: por margen (cada umbral), por tipo (si hay -cascada-tipos) y margen o tipo.
		type varCascada struct {
			nombre string
			umbral float64
			tipos  map[string]bool
		}
		var vs []varCascada
		for _, u := range listaFloats(*cascada) {
			vs = append(vs, varCascada{fmt.Sprintf("cascada margen<%.3f (top %d)", u, *topRerank), u, nil})
		}
		if len(tipos) > 0 {
			vs = append(vs, varCascada{fmt.Sprintf("cascada tipos (top %d)", *topRerank), -1, tipos})
			for _, u := range listaFloats(*cascada) {
				vs = append(vs, varCascada{fmt.Sprintf("cascada margen<%.3f o tipos (top %d)", u, *topRerank), u, tipos})
			}
		}
		sinRR, conRR := nuevo(busqueda.ModoHibrido, false, false), nuevo(busqueda.ModoHibrido, false, true)
		for _, v := range vs {
			cc := &cascadaRR{nombre: v.nombre, umbral: v.umbral, tipos: v.tipos, tipoDe: tipoDe, sin: sinRR, con: conRR, decision: map[string]bool{}}
			cascadas = append(cascadas, cc)
			metodos = append(metodos, metodo{v.nombre, cc.buscar})
		}
	}

	if *soloMetodos != "" {
		var elegidos []metodo
		for _, m := range metodos {
			for _, n := range strings.Split(*soloMetodos, "|") {
				if strings.TrimSpace(n) == m.nombre {
					elegidos = append(elegidos, m)
				}
			}
		}
		metodos = elegidos
	}

	if *rutaPool != "" {
		fallar(escribirPool(ctx, *rutaPool, oro, metodos, lex, *profPool))
		logf("pool escrito en %s", *rutaPool)
		return
	}

	// RSS del reranker: muestreo en segundo plano.
	var rss *muestreoRSS
	if *pidRerank > 0 {
		rss = muestrearRSS(*pidRerank)
	}

	// 7. Evaluación.
	detalle := map[string]map[string][]string{}
	fmt.Printf("## Resultados (%d consultas, k1=%.2f b=%.2f exacto=%.2f rrfk=%.0f peso_vector=%.2f alias=%.2f top_rerank=%d)\n\n",
		len(oro), *k1, *b, *pesoExacto, *rrfK, *pesoVector, *pesoAlias, *topRerank)
	fmt.Println("| Método | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | p50 ms | p95 ms |")
	fmt.Println("|---|---|---|---|---|---|---|")
	medidas := map[string]medida{}
	for _, m := range metodos {
		rep := *repes
		if strings.Contains(m.nombre, "rerank") {
			rep = *repesRR
		}
		md, det, err := evaluar(ctx, m, oro, rep)
		fallar(err)
		medidas[m.nombre] = md
		detalle[m.nombre] = det
		fmt.Printf("| %s | %.3f | %.3f | %.3f | %.3f | %.1f | %.1f |\n", m.nombre, md.r5, md.r10, md.mrr, md.ndcg,
			ms(percentil(md.lat, 50)), ms(percentil(md.lat, 95)))
	}
	fmt.Println()
	tipos := []string{}
	for _, c := range oro {
		if !contiene(tipos, c.Tipo) {
			tipos = append(tipos, c.Tipo)
		}
	}
	sort.Strings(tipos)
	fmt.Println("### Por tipo de consulta (Recall@10 / MRR@10)")
	fmt.Println()
	fmt.Printf("| Método |")
	for _, t := range tipos {
		fmt.Printf(" %s (n=%d) |", t, contarTipo(oro, t))
	}
	fmt.Println()
	fmt.Println("|---|" + strings.Repeat("---|", len(tipos)))
	for _, m := range metodos {
		fmt.Printf("| %s |", m.nombre)
		for _, t := range tipos {
			v := medidas[m.nombre].porTipo[t]
			fmt.Printf(" %.3f / %.3f |", v[1]/v[4], v[2]/v[4])
		}
		fmt.Println()
	}
	fmt.Println()
	for _, cc := range cascadas {
		if len(cc.decision) == 0 {
			continue
		}
		n := 0
		for _, v := range cc.decision {
			if v {
				n++
			}
		}
		fmt.Printf("%s: reordenó %d de %d consultas (%.0f %%); margen relativo p10/p25/p50/p75 = %.3f / %.3f / %.3f / %.3f\n\n",
			cc.nombre, n, len(cc.decision), 100*float64(n)/float64(len(cc.decision)),
			cuantil(cc.margenes, 0.10), cuantil(cc.margenes, 0.25), cuantil(cc.margenes, 0.50), cuantil(cc.margenes, 0.75))
	}
	if rss != nil {
		pico, ultimo := rss.parar()
		fmt.Printf("RSS de llama-server (pid %d): pico %.0f MiB, al terminar %.0f MiB\n\n", *pidRerank, float64(pico)/1024, float64(ultimo)/1024)
	}
	fmt.Printf("Índice BM25: %d docs, %d términos, %.1f MiB de montículo Go, construido en %v; gob %.1f MiB, carga %v\n\n",
		lex.Contar(), lex.Vocabulario(), mib(heapIndice), tIndice.Round(time.Millisecond), mib(uint64(fi.Size())), tCargaGob.Round(time.Millisecond))

	// 8. Casos detallados.
	if *casos != "" {
		for _, q := range strings.Split(*casos, "|") {
			fmt.Printf("### Caso: «%s»\n\n", q)
			var rel map[string]int
			for _, c := range oro {
				if c.Consulta == q {
					rel = c.rel
				}
			}
			for _, m := range metodos {
				ids, err := m.buscar(ctx, q, 5)
				fallar(err)
				fmt.Printf("**%s**\n\n", m.nombre)
				for i, id := range ids {
					marca := ""
					if g := rel[id]; g > 0 {
						marca = fmt.Sprintf(" ✔ (grado %d)", g)
					}
					fmt.Printf("%d. `%s` — %s%s\n", i+1, recortar(id, 40), describir(lex, id), marca)
				}
				fmt.Println()
			}
		}
	}

	// 9. Barrido.
	if *barrido {
		barrer(ctx, lex, vec, al, rr, oro, opc, cfg)
	}

	if *rutaJSON != "" {
		b, _ := json.MarshalIndent(detalle, "", " ")
		fallar(os.WriteFile(*rutaJSON, b, 0o644))
	}
}

func evaluar(ctx context.Context, m metodo, oro []casoOro, repes int) (medida, map[string][]string, error) {
	md := medida{porTipo: map[string][]float64{}}
	det := map[string][]string{}
	// Calentamiento.
	for _, c := range oro[:min(3, len(oro))] {
		if _, err := m.buscar(ctx, c.Consulta, 10); err != nil {
			return md, nil, fmt.Errorf("%s: %w", m.nombre, err)
		}
	}
	for _, c := range oro {
		var ids []string
		for r := 0; r < max(repes, 1); r++ {
			t := time.Now()
			var err error
			ids, err = m.buscar(ctx, c.Consulta, 10)
			md.lat = append(md.lat, time.Since(t))
			if err != nil {
				return md, nil, fmt.Errorf("%s «%s»: %w", m.nombre, c.Consulta, err)
			}
		}
		det[c.ID] = ids
		r5, r10, mrr, ndcg := metricas(ids, c.rel)
		md.r5 += r5
		md.r10 += r10
		md.mrr += mrr
		md.ndcg += ndcg
		v := md.porTipo[c.Tipo]
		if v == nil {
			v = make([]float64, 5)
		}
		v[0] += r5
		v[1] += r10
		v[2] += mrr
		v[3] += ndcg
		v[4]++
		md.porTipo[c.Tipo] = v
	}
	n := float64(len(oro))
	md.r5, md.r10, md.mrr, md.ndcg = md.r5/n, md.r10/n, md.mrr/n, md.ndcg/n
	return md, det, nil
}

// metricas: Recall@5 y @10 acotados (aciertos / min(|relevantes|, k): una
// consulta con 12 duplicados válidos puede llegar a 1), MRR@10 del primer
// relevante y nDCG@10 con ganancia 2^grado − 1.
func metricas(ids []string, rel map[string]int) (r5, r10, mrr, ndcg float64) {
	if len(rel) == 0 {
		return
	}
	var h5, h10 int
	var dcg float64
	for i, id := range ids {
		if i >= 10 {
			break
		}
		g := rel[id]
		if g <= 0 {
			continue
		}
		if i < 5 {
			h5++
		}
		h10++
		if mrr == 0 {
			mrr = 1 / float64(i+1)
		}
		dcg += (math.Pow(2, float64(g)) - 1) / math.Log2(float64(i+2))
	}
	grados := make([]int, 0, len(rel))
	for _, g := range rel {
		grados = append(grados, g)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(grados)))
	var idcg float64
	for i, g := range grados {
		if i >= 10 {
			break
		}
		idcg += (math.Pow(2, float64(g)) - 1) / math.Log2(float64(i+2))
	}
	n := float64(len(rel))
	return float64(h5) / math.Min(n, 5), float64(h10) / math.Min(n, 10), mrr, dcg / idcg
}

func barrer(ctx context.Context, lex *busqueda.Indice, vec busqueda.Vectorial, al *busqueda.Alias, rr busqueda.Reordenador,
	oro []casoOro, base busqueda.Opciones, cfg busqueda.ConfigBM25) {
	puntuar := func(bu *busqueda.Buscador) (float64, float64, float64) {
		var r10, mrr, nd float64
		for _, c := range oro {
			rs, err := bu.Buscar(ctx, c.Consulta, 10)
			fallar(err)
			_, a, b, d := metricas(idsDe(rs), c.rel)
			r10 += a
			mrr += b
			nd += d
		}
		n := float64(len(oro))
		return r10 / n, mrr / n, nd / n
	}
	fmt.Println("## Barrido")
	fmt.Println()
	fmt.Println("### BM25 solo: k1 × b (Recall@10 / MRR@10 / nDCG@10)")
	fmt.Println()
	bs := []float64{0.3, 0.5, 0.75, 0.9}
	fmt.Print("| k1 \\ b |")
	for _, b := range bs {
		fmt.Printf(" %.2f |", b)
	}
	fmt.Println()
	fmt.Println("|---|" + strings.Repeat("---|", len(bs)))
	for _, k1 := range []float64{0.6, 0.9, 1.2, 1.5, 2.0} {
		fmt.Printf("| %.1f |", k1)
		for _, b := range bs {
			lex.AjustarParametros(k1, b, nil, cfg.PesoExacto)
			o := base
			o.Modo = busqueda.ModoLexico
			a, m, n := puntuar(busqueda.NuevoBuscador(lex, vec, nil, al, o))
			fmt.Printf(" %.3f / %.3f / %.3f |", a, m, n)
		}
		fmt.Println()
	}
	lex.AjustarParametros(cfg.K1, cfg.B, nil, cfg.PesoExacto)
	fmt.Println()

	fmt.Println("### BM25 solo: peso del término exacto y pesos de campo")
	fmt.Println()
	fmt.Println("| Variante | Recall@10 | MRR@10 | nDCG@10 |")
	fmt.Println("|---|---|---|---|")
	variantes := []struct {
		nombre string
		pesos  map[string]float64
		exacto float64
	}{
		{"defecto (título/sección/preguntas 3, texto 1, ocr 0,5, manual 0,5; exacto 0,5)", cfg.Pesos, 0.5},
		{"sin término exacto", cfg.Pesos, 0},
		{"exacto 1,0", cfg.Pesos, 1},
		{"todos los campos 1 (BM25 plano)", map[string]float64{"titulo": 1, "seccion": 1, "preguntas": 1, "texto": 1, "ocr": 1, "manual": 1}, 0.5},
		{"título/sección 2", map[string]float64{"titulo": 2, "seccion": 2, "preguntas": 3, "texto": 1, "ocr": 0.5, "manual": 0.5}, 0.5},
		{"título/sección 5", map[string]float64{"titulo": 5, "seccion": 5, "preguntas": 3, "texto": 1, "ocr": 0.5, "manual": 0.5}, 0.5},
		{"ocr 0 (sin capturas)", map[string]float64{"titulo": 3, "seccion": 3, "preguntas": 3, "texto": 1, "ocr": 0, "manual": 0.5}, 0.5},
		{"ocr 1", map[string]float64{"titulo": 3, "seccion": 3, "preguntas": 3, "texto": 1, "ocr": 1, "manual": 0.5}, 0.5},
	}
	for _, v := range variantes {
		lex.AjustarParametros(cfg.K1, cfg.B, v.pesos, v.exacto)
		o := base
		o.Modo = busqueda.ModoLexico
		a, m, n := puntuar(busqueda.NuevoBuscador(lex, vec, nil, al, o))
		fmt.Printf("| %s | %.3f | %.3f | %.3f |\n", v.nombre, a, m, n)
	}
	lex.AjustarParametros(cfg.K1, cfg.B, cfg.Pesos, cfg.PesoExacto)
	fmt.Println()

	fmt.Println("### Híbrida: peso de la lista vectorial en la RRF (k de la configuración; Recall@10 / MRR@10 / nDCG@10)")
	fmt.Println()
	fmt.Println("| peso vector | sin alias | con alias |")
	fmt.Println("|---|---|---|")
	for _, pv := range []float64{1, 0.7, 0.5, 0.3, 0.2, 0.1} {
		fmt.Printf("| %.1f |", pv)
		for _, ex := range []bool{false, true} {
			o := base
			o.Modo, o.PesoVector, o.PesoLexico, o.Expandir = busqueda.ModoHibrido, pv, 1, ex
			o.PesoVectorSinRerank = pv
			a, m, n := puntuar(busqueda.NuevoBuscador(lex, vec, nil, al, o))
			fmt.Printf(" %.3f / %.3f / %.3f |", a, m, n)
		}
		fmt.Println()
	}
	fmt.Println()

	fmt.Println("### Híbrida: k de RRF y peso de los alias (Recall@10 / MRR@10 / nDCG@10)")
	fmt.Println()
	pas := []float64{0, 0.2, 0.3, 0.5}
	fmt.Print("| k RRF \\ peso alias |")
	for _, p := range pas {
		fmt.Printf(" %.1f |", p)
	}
	fmt.Println()
	fmt.Println("|---|" + strings.Repeat("---|", len(pas)))
	for _, k := range []float64{10, 30, 60, 100} {
		fmt.Printf("| %.0f |", k)
		for _, p := range pas {
			o := base
			o.Modo, o.RRFK, o.Expandir, o.PesoAlias = busqueda.ModoHibrido, k, p > 0, p
			a, m, n := puntuar(busqueda.NuevoBuscador(lex, vec, nil, al, o))
			fmt.Printf(" %.3f / %.3f / %.3f |", a, m, n)
		}
		fmt.Println()
	}
	fmt.Println()

	if rr == nil {
		return
	}
	fmt.Println("### Reranker: N candidatos (híbrida sin alias, k RRF de la configuración)")
	fmt.Println()
	fmt.Println("| N | Recall@5 | Recall@10 | MRR@10 | nDCG@10 | p50 ms | p95 ms |")
	fmt.Println("|---|---|---|---|---|---|---|")
	for _, n := range []int{10, 20, 30, 50} {
		o := base
		o.Modo, o.TopRerank, o.Expandir = busqueda.ModoHibrido, n, false
		bu := busqueda.NuevoBuscador(lex, vec, rr, al, o)
		md, _, err := evaluar(ctx, metodo{fmt.Sprint(n), func(ctx context.Context, q string, k int) ([]string, error) {
			rs, err := bu.Buscar(ctx, q, k)
			return idsDe(rs), err
		}}, oro, 1)
		fallar(err)
		fmt.Printf("| %d | %.3f | %.3f | %.3f | %.3f | %.1f | %.1f |\n", n, md.r5, md.r10, md.mrr, md.ndcg, ms(percentil(md.lat, 50)), ms(percentil(md.lat, 95)))
	}
	fmt.Println()
}

func escribirPool(ctx context.Context, ruta string, oro []casoOro, metodos []metodo, lex *busqueda.Indice, k int) error {
	f, err := os.Create(ruta)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	for _, c := range oro {
		fmt.Fprintf(w, "\n=== %s «%s» [%s]\n", c.ID, c.Consulta, c.Tipo)
		vistos := map[string][]string{}
		var orden []string
		for _, m := range metodos {
			ids, err := m.buscar(ctx, c.Consulta, k)
			if err != nil {
				return err
			}
			for i, id := range ids {
				if _, ok := vistos[id]; !ok {
					orden = append(orden, id)
				}
				vistos[id] = append(vistos[id], fmt.Sprintf("%s#%d", abreviar(m.nombre), i+1))
			}
		}
		for _, id := range orden {
			marca := "  "
			if g := c.rel[id]; g > 0 {
				marca = fmt.Sprintf("R%d", g)
			}
			meta, _ := lex.Meta(id)
			txt, _ := lex.Texto(id)
			txt = strings.Join(strings.Fields(txt), " ")
			fmt.Fprintf(w, "%s %s | %s | %s\n     %s\n", marca, id, meta["manual"], strings.Join(vistos[id], ","), recortar(txt, 260))
		}
	}
	return nil
}

func abreviar(n string) string {
	r := strings.NewReplacer("vector (hoy)", "V", "BM25 + alias", "Ba", "BM25", "B", "híbrida RRF + alias", "Ha", "híbrida RRF", "H")
	if strings.Contains(n, "rerank") {
		if strings.Contains(n, "alias") {
			return "RRa"
		}
		return "RR"
	}
	return r.Replace(n)
}

func cargarOro(ruta string, docs []busqueda.Doc, permitirVacios bool) ([]casoOro, error) {
	// El oro se escribe con el par (id, manual): el id solo se repite entre
	// manuales (200 ids, 1.061 fragmentos). Cada par debe dar un documento.
	porClave := map[string][]string{}
	for _, d := range docs {
		k := d.Meta["id_original"] + "\x00" + d.Meta["manual"]
		porClave[k] = append(porClave[k], d.ID)
	}
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []casoOro
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<22)
	n := 0
	var errs []string
	for sc.Scan() {
		n++
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "//") {
			continue
		}
		var c casoOro
		if err := json.Unmarshal([]byte(l), &c); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", ruta, n, err)
		}
		c.rel = map[string]int{}
		for _, r := range c.Relevantes {
			ids, ok := porClave[r.ID+"\x00"+r.Manual]
			if !ok {
				errs = append(errs, fmt.Sprintf("%s: (%s, %s) no existe en la KB", c.ID, r.ID, r.Manual))
				continue
			}
			if len(ids) > 1 {
				errs = append(errs, fmt.Sprintf("%s: (%s, %s) es ambiguo: %v", c.ID, r.ID, r.Manual, ids))
				continue
			}
			g := r.Grado
			if g <= 0 {
				g = 1
			}
			c.rel[ids[0]] = g
		}
		if len(c.rel) == 0 && !permitirVacios {
			errs = append(errs, c.ID+": sin relevantes")
		}
		out = append(out, c)
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("oro inválido:\n  %s", strings.Join(errs, "\n  "))
	}
	return out, sc.Err()
}

func concatenar(rutas []string) (string, error) {
	f, err := os.CreateTemp("", "evalbusqueda-kb-*.jsonl")
	if err != nil {
		return "", err
	}
	defer f.Close()
	for _, r := range rutas {
		b, err := os.ReadFile(r)
		if err != nil {
			return "", err
		}
		if _, err := f.Write(b); err != nil {
			return "", err
		}
	}
	return f.Name(), nil
}

type muestreoRSS struct {
	mu     sync.Mutex
	pico   int64
	ultimo int64
	fin    chan struct{}
	hecho  chan struct{}
}

func muestrearRSS(pid int) *muestreoRSS {
	m := &muestreoRSS{fin: make(chan struct{}), hecho: make(chan struct{})}
	leer := func() {
		out, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(pid)).Output()
		if err != nil {
			return
		}
		v, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
		if err != nil {
			return
		}
		m.mu.Lock()
		m.ultimo = v
		m.pico = max(m.pico, v)
		m.mu.Unlock()
	}
	go func() {
		defer close(m.hecho)
		t := time.NewTicker(200 * time.Millisecond)
		defer t.Stop()
		leer()
		for {
			select {
			case <-m.fin:
				leer()
				return
			case <-t.C:
				leer()
			}
		}
	}()
	return m
}

func (m *muestreoRSS) parar() (pico, ultimo int64) {
	close(m.fin)
	<-m.hecho
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pico, m.ultimo
}

func heapVivo() uint64 {
	runtime.GC()
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.HeapAlloc
}

func describir(lex *busqueda.Indice, id string) string {
	m, ok := lex.Meta(id)
	if !ok {
		return "(fuera del índice léxico)"
	}
	t := m["seccion"]
	if t == "" {
		t = m["titulo"]
	}
	if m["pagina"] != "" && m["pagina"] != "0" && m["seccion"] == "" {
		t += ", p. " + m["pagina"]
	}
	tipo := m["tipo"]
	if tipo == "" {
		tipo = "texto"
	}
	return fmt.Sprintf("%s · %s · %s", recortar(m["manual"], 40), recortar(t, 70), tipo)
}

func idsDe(rs []busqueda.Resultado) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.ID
	}
	return out
}

func percentil(ds []time.Duration, p float64) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	s := append([]time.Duration(nil), ds...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	i := int(math.Ceil(p/100*float64(len(s)))) - 1
	return s[max(0, min(i, len(s)-1))]
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }
func mib(b uint64) float64       { return float64(b) / (1 << 20) }

func recortar(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func contiene(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

func contarTipo(oro []casoOro, t string) int {
	n := 0
	for _, c := range oro {
		if c.Tipo == t {
			n++
		}
	}
	return n
}

func fallar(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "evalbusqueda:", err)
		os.Exit(1)
	}
}

// cascadaRR: el reranker solo se llama cuando el orden sin reranker duda
// (margen relativo entre el 1.º y el 2.º menor que el umbral) o cuando el
// tipo de la consulta está en tipos. Mide el coste de una cascada sin tocar
// internal/busqueda: si se reordena, la búsqueda híbrida se repite con
// reranker (~12 ms de más frente a reutilizar las listas).
type cascadaRR struct {
	nombre   string
	umbral   float64
	tipos    map[string]bool
	tipoDe   map[string]string
	sin, con *busqueda.Buscador
	mu       sync.Mutex
	decision map[string]bool // consulta → se reordenó
	margenes []float64       // uno por consulta distinta
}

func (c *cascadaRR) buscar(ctx context.Context, q string, k int) ([]string, error) {
	inf, err := c.sin.BuscarInforme(ctx, q, k)
	if err != nil {
		return nil, err
	}
	m := margenRelativo(inf.Resultados)
	usar := m < c.umbral || c.tipos[c.tipoDe[q]]
	c.mu.Lock()
	if _, visto := c.decision[q]; !visto {
		c.margenes = append(c.margenes, m)
	}
	c.decision[q] = usar
	c.mu.Unlock()
	if !usar {
		return idsDe(inf.Resultados), nil
	}
	inf, err = c.con.BuscarInforme(ctx, q, k)
	if err == nil && len(inf.Degradado) > 0 {
		err = fmt.Errorf("degradado %v: %s %s", inf.Degradado, inf.FalloVector, inf.FalloRerank)
	}
	return idsDe(inf.Resultados), err
}

// margenRelativo: (s1 − s2) / s1 del puntaje final; 1 si hay menos de dos.
func margenRelativo(rs []busqueda.Resultado) float64 {
	if len(rs) < 2 || rs[0].Score <= 0 {
		return 1
	}
	return (rs[0].Score - rs[1].Score) / rs[0].Score
}

func cuantil(xs []float64, q float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	c := append([]float64(nil), xs...)
	sort.Float64s(c)
	return c[min(len(c)-1, int(q*float64(len(c))))]
}

// listaVariantes lee «N:runas,N:runas».
func listaVariantes(s string) [][2]int {
	var out [][2]int
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}
		a, b, ok := strings.Cut(p, ":")
		n, err1 := strconv.Atoi(strings.TrimSpace(a))
		r, err2 := strconv.Atoi(strings.TrimSpace(b))
		if !ok || err1 != nil || err2 != nil || n <= 0 || r <= 0 {
			fallar(fmt.Errorf("-rerank-variantes: %q no es N:runas", p))
		}
		out = append(out, [2]int{n, r})
	}
	return out
}

func listaFloats(s string) []float64 {
	var out []float64
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			fallar(fmt.Errorf("-cascada: %q no es un número", p))
		}
		out = append(out, f)
	}
	return out
}

// cachePuntajes guarda el puntaje del reranker por (runas, consulta, documento).
// Con -rerank-cache las variantes de N comparten los pares ya puntuados.
type cachePuntajes struct {
	mu       sync.Mutex
	m        map[string]float64
	aciertos int
	llamadas int
}

type rerankConCache struct {
	c     *cachePuntajes
	r     busqueda.Reordenador
	runas int
}

// envolver devuelve r tal cual si la caché está apagada (c == nil).
func (c *cachePuntajes) envolver(r busqueda.Reordenador, runas int) busqueda.Reordenador {
	if c == nil {
		return r
	}
	return rerankConCache{c: c, r: r, runas: runas}
}

func (w rerankConCache) Rerank(ctx context.Context, q string, docs []string) ([]float64, error) {
	clave := func(d string) string { return strconv.Itoa(w.runas) + "\x00" + q + "\x00" + recortar(d, w.runas) }
	out := make([]float64, len(docs))
	var faltan []int
	w.c.mu.Lock()
	for i, d := range docs {
		if v, ok := w.c.m[clave(d)]; ok {
			out[i] = v
			w.c.aciertos++
		} else {
			faltan = append(faltan, i)
		}
	}
	w.c.mu.Unlock()
	if len(faltan) == 0 {
		return out, nil
	}
	sub := make([]string, len(faltan))
	for j, i := range faltan {
		sub[j] = docs[i]
	}
	ps, err := w.r.Rerank(ctx, q, sub)
	if err != nil {
		return nil, err
	}
	if len(ps) != len(sub) {
		return nil, fmt.Errorf("rerank: %d puntajes para %d documentos", len(ps), len(sub))
	}
	w.c.mu.Lock()
	w.c.llamadas++
	for j, i := range faltan {
		out[i] = ps[j]
		w.c.m[clave(docs[i])] = ps[j]
	}
	w.c.mu.Unlock()
	return out, nil
}
