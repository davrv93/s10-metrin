package main

// Cableado de la V2 de Metrín (internal/v2). Solo se construye con AGENT_V2_ENABLED=true o AGENT_VERSION=v2;
// apagada, rag.RAG.V2 queda nil y /ask es V1 byte a byte.

import (
	"context"
	"os"
	"time"

	"rag-go/internal/almacen"
	"rag-go/internal/busqueda"
	"rag-go/internal/clasificar"
	"rag-go/internal/config"
	"rag-go/internal/embed"
	"rag-go/internal/indexar"
	"rag-go/internal/rag"
	"rag-go/internal/rerank"
	"rag-go/internal/v2"
	"rag-go/internal/v2/conocimiento"
)

// Las piezas de conocimiento cumplen las interfaces del núcleo (se comprueba al compilar).
var (
	_ v2.Aliaser                 = (*conocimiento.Base)(nil)
	_ v2.Catalogo                = (*conocimiento.Base)(nil)
	_ v2.Constructor             = (*conocimiento.Constructor)(nil)
	_ v2.Continuador             = (*conocimiento.Constructor)(nil)
	_ conocimiento.Hibrido       = (*busqueda.Buscador)(nil)
	_ busqueda.Reordenador       = rerankConRespaldo{}
	_ busqueda.VectorialFiltrado = busqueda.VectorAlmacen{}
)

// nuevoAgenteV2 arma el orquestador V2 con lo que hay: clasificador de tipo (reglas + kNN si V2_TIPO_MODELO
// carga), motor de decisión (DECISION_ENGINE) y la charla de V1 para los mensajes SOCIAL. Las piezas de
// conocimiento (recuperador, constructor, catálogo, plantillas, gate) las pone conocimientoV2.
func nuevoAgenteV2(cfg config.Config, e embed.Embebedor, clas *clasificar.Modelo, r *rag.RAG) (*v2.Agente, error) {
	ag := v2.Nuevo(cfg.V2)
	motor, err := v2.NuevoMotor(cfg.V2)
	if err != nil {
		return nil, err
	}
	ag.Decision = motor
	ag.Charla = r
	cl := &v2.Clasificador{Emb: e, Intencion: clas, Umbral: cfg.V2.TipoUmbral, Margen: cfg.V2.TipoMargen}
	if cfg.V2.TipoModelo != "" {
		m, err := v2.CargarModeloTipo(context.Background(), cfg.V2.TipoModelo, e, cfg.V2.TipoUmbral)
		if err != nil {
			logf("V2: sin modelo de tipo de respuesta (%v): solo reglas léxicas", err)
		} else {
			cl.Tipo = m
			logf("V2: modelo de tipo de respuesta %s (%d clases)", cfg.V2.TipoModelo, len(m.Intenciones))
		}
	}
	ag.Clasificador = cl
	conocimientoV2(ag, cfg.V2, r.Almacen)
	logf("V2 activa: version=%s habilitada=%v porcentaje=%d motor=%s", cfg.V2.Version, cfg.V2.Habilitada,
		cfg.V2.Porcentaje, motor.Nombre())
	return ag, nil
}

// conocimientoV2 enchufa internal/v2/conocimiento: kb/procedimientos, kb/conceptos, plantillas y fragmentos
// (rutas por entorno: RAG_KB, RAG_PLANTILLAS, RAG_MEDIA_DATOS). Nunca falla: lo que no carga se informa y, sin
// procedimientos ni conceptos, la V2 responde SIN_EVIDENCIA (con su motivo en la traza).
func conocimientoV2(ag *v2.Agente, cfg v2.Config, alm *almacen.Almacen) {
	b := conocimiento.Cargar(conocimiento.Opciones{DirDatos: conocimiento.RutaDatos()})
	graves := b.Graves()
	for i, e := range graves {
		if i == 5 {
			logf("V2: … y %d errores graves más de carga", len(graves)-5)
			break
		}
		logf("V2: %v", e)
	}
	rec := conocimiento.NuevoRecuperador(b)
	if cfg.Busqueda == v2.BusquedaHibrida {
		if bu, err := buscadorV2(cfg, b.DirKB, alm); err != nil {
			logf("V2: sin búsqueda híbrida de fragmentos (%v): solo el BM25 propio", err)
			busquedaV2 = map[string]any{"modo": v2.BusquedaLexica, "fallo": err.Error()}
		} else {
			rec.Hibrido = bu
		}
	}
	pl := conocimiento.NuevoMotorPlantillas(b)
	ag.Aliaser, ag.Catalogo, ag.Recuperador = b, b, rec
	ag.Constructor = conocimiento.NuevoConstructor(b, rec)
	ag.Plantillas = pl
	ag.Gate = conocimiento.NuevoGate(b)
	// GENERATION_ENGINE=local + GENERATION_MODEL: un LLM local que solo reformula; si no, plantillas.
	if gen := conocimiento.NuevoMotorGeneracion(b, os.Getenv); gen.Nombre() != pl.Nombre() {
		ag.Generador = gen
	}
	logf("V2: conocimiento de %s: %d procedimientos, %d conceptos, %d avisos de carga (%d graves)",
		b.DirKB, len(b.Procedimientos), len(b.Conceptos), len(b.Errores), len(graves))
}

// busquedaV2: lo que /health dice de la búsqueda de fragmentos de la V2 (lo real, no lo pedido).
var busquedaV2 = map[string]any{"modo": v2.BusquedaLexica}

// buscadorV2 arma la búsqueda híbrida de internal/busqueda para la clase «fragmento» de la V2
// (metrin/eval/BUSQUEDA.md §8): BM25F sobre kb/fragmentos*.jsonl (con los mismos ids que el índice vectorial) y el
// vector del almacén de V1, de solo lectura, filtrado a la fuente s10-kb; RRF con las opciones medidas
// (busqueda.OpcionesPorDefecto). El reranker solo con RERANK_URL (apagado por defecto: ~2 s y ~880 MB con Metal,
// sin medir en la CPU de producción). V1 no usa nada de esto.
func buscadorV2(cfg v2.Config, dirKB string, alm *almacen.Almacen) (*busqueda.Buscador, error) {
	t0 := time.Now()
	rutas, err := busqueda.RutasFragmentos(dirKB)
	if err != nil {
		return nil, err
	}
	docs, err := busqueda.CargarFragmentos(rutas...)
	if err != nil {
		return nil, err
	}
	lex, err := busqueda.IndexarDocs(busqueda.ConfigPorDefecto(), docs)
	if err != nil {
		return nil, err
	}
	o := busqueda.OpcionesPorDefecto()
	var vec busqueda.Vectorial
	if alm != nil && alm.Contar() > 0 {
		vec = busqueda.VectorAlmacen{A: alm, Filtro: map[string]string{"source": indexar.FuenteS10KB}}
	}
	var rr busqueda.Reordenador
	if cfg.RerankURL != "" {
		ms := cfg.RerankTimeoutMs
		if ms <= 0 {
			ms = v2.RerankTimeoutDefectoMs
		}
		ll := rerank.NuevoLlamaServer(cfg.RerankURL)
		ll.MaxRunas = cfg.RerankMaxRunas // RERANK_MAX_RUNAS (0 = el de internal/rerank)
		if cfg.RerankTopN > 0 {
			o.TopRerank = cfg.RerankTopN // RERANK_TOP_N
		}
		rr = rerankConRespaldo{principal: ll, timeout: time.Duration(ms) * time.Millisecond}
		// El Buscador tiene su propio plazo: un poco más largo, para que el fallo que se anote sea el de ConRespaldo.
		o.TimeoutRerank = time.Duration(ms)*time.Millisecond + 250*time.Millisecond
	}
	bu := busqueda.NuevoBuscador(lex, vec, rr, nil, o)
	busquedaV2 = map[string]any{"modo": v2.BusquedaHibrida, "fragmentos": len(docs), "vector": vec != nil, "reranker": rr != nil}
	if rr != nil {
		busquedaV2["rerank_top_n"], busquedaV2["rerank_max_runas"] = o.TopRerank, cfg.RerankMaxRunas
	}
	logf("V2: búsqueda híbrida de fragmentos: %d fragmentos en BM25F, vector %v, reranker %v (%s)",
		len(docs), vec != nil, rr != nil, time.Since(t0).Round(time.Millisecond))
	return bu, nil
}

// rerankConRespaldo es rerank.NuevoLlamaServer(RERANK_URL) envuelto en rerank.ConRespaldo (tiempo máximo; si falla,
// el orden de entrada). Ese orden de entrada (Noop) no es un reordenado: aquí vuelve a ser un error para que el
// Buscador lo anote («sin_rerank», FalloRerank en la traza) y se quede con su orden sin reranker (vector a 0,1), que
// rinde más que el pool de pesos iguales que recibiría el reranker (BUSQUEDA.md §4.2: MRR 0,780 frente a 0,634).
type rerankConRespaldo struct {
	principal rerank.Reranker
	timeout   time.Duration
}

func (r rerankConRespaldo) Rerank(ctx context.Context, q string, docs []string) ([]float64, error) {
	var fallo error
	s, err := rerank.ConRespaldo{Principal: r.principal, Timeout: r.timeout, AlFallar: func(e error) { fallo = e }}.Rerank(ctx, q, docs)
	if fallo != nil {
		return nil, fallo
	}
	return s, err
}

// estadoV2: lo que /health dice de la V2 (nil sin V2: /health no cambia).
func estadoV2(cfg config.Config) any {
	if !cfg.V2.Activa() {
		return nil
	}
	return map[string]any{
		"version":    string(cfg.V2.Version),
		"habilitada": cfg.V2.Habilitada,
		"porcentaje": cfg.V2.Porcentaje,
		"motor":      cfg.V2.MotorDecision,
		"busqueda":   busquedaV2,
	}
}
