package v2

import (
	"context"
	"errors"
	"strings"
	"sync"

	"rag-go/internal/rag"
	"rag-go/internal/v2/tipos"
)

// Dobles de las piezas de conocimiento (las reales viven en internal/v2/conocimiento).

// procMetrado: procedimiento de prueba con tres pasos; el 1 y el 3 traen foto, el 2 no.
func procMetrado() tipos.ProcedimientoDef {
	return tipos.ProcedimientoDef{
		ID: "metrados.registrar", Titulo: "Registrar un metrado", Modulo: "Presupuestos",
		Prerrequisitos: []tipos.ConFuente{{Texto: "Tener un presupuesto abierto.", Fuente: []string{"f1"}}},
		Pasos: []tipos.Paso{
			{ID: "metrados.registrar#1", N: 1, Accion: "Abra la hoja del presupuesto.", Fuente: []string{"f1"},
				Fotos: []tipos.Foto{{ID: "img_1", Ruta: "imagenes/p/1.png", Pagina: 11}}},
			{ID: "metrados.registrar#2", N: 2, Accion: "Elija la partida y pulse Metrado.", Fuente: []string{"f1"}},
			{ID: "metrados.registrar#3", N: 3, Accion: "Escriba las cantidades y pulse Aceptar.", Fuente: []string{"f2"},
				Fotos: []tipos.Foto{{ID: "img_3", Ruta: "imagenes/p/3.png", Pagina: 12}}},
		},
		Verificacion:      []tipos.ConFuente{{Texto: "El total de la partida cambia.", Fuente: []string{"f2"}}},
		ErroresFrecuentes: []tipos.ErrorFrecuente{{Sintoma: "No aparece Metrado", Solucion: "Elija primero una partida.", Fuente: []string{"f2"}}},
		Fuentes: []tipos.Fuente{
			{ID: "f1", Manual: "Manual de Presupuestos", Pagina: 11},
			{ID: "f2", Manual: "Manual de Presupuestos", Pagina: 12},
		},
	}
}

func procUnidades() tipos.ProcedimientoDef {
	return tipos.ProcedimientoDef{
		ID: "config.unidades", Titulo: "Configurar unidades",
		Pasos: []tipos.Paso{
			{ID: "config.unidades#1", N: 1, Accion: "Abra Tablas › Unidades.", Fuente: []string{"u1"}},
			{ID: "config.unidades#2", N: 2, Accion: "Pulse Nuevo y guarde.", Fuente: []string{"u1"}},
		},
		Fuentes: []tipos.Fuente{{ID: "u1", Manual: "Manual de Presupuestos", Pagina: 5}},
	}
}

func conceptoMetrado() tipos.ConceptoDef {
	return tipos.ConceptoDef{ID: "metrado", Termino: "Metrado", Definicion: "Cuantificación de las partidas de una obra.",
		EnS10: "Se registra en la hoja del presupuesto.", Fuente: []string{"c1"}}
}

type catalogoFalso struct {
	procs     map[string]tipos.ProcedimientoDef
	conceptos map[string]tipos.ConceptoDef
}

func nuevoCatalogo() *catalogoFalso {
	c := &catalogoFalso{procs: map[string]tipos.ProcedimientoDef{}, conceptos: map[string]tipos.ConceptoDef{}}
	for _, p := range []tipos.ProcedimientoDef{procMetrado(), procUnidades()} {
		c.procs[p.ID] = p
	}
	c.conceptos["metrado"] = conceptoMetrado()
	return c
}

func (c *catalogoFalso) Procedimiento(id string) (tipos.ProcedimientoDef, bool) {
	p, ok := c.procs[id]
	return p, ok
}

func (c *catalogoFalso) Concepto(id string) (tipos.ConceptoDef, bool) {
	x, ok := c.conceptos[id]
	return x, ok
}

// llamada al recuperador.
type llamada struct {
	consulta tipos.Consulta
	clase    string
	k        int
}

// recuperadorFalso responde según la clase y el número de llamada; cuenta todo.
type recuperadorFalso struct {
	mu        sync.Mutex
	llamadas  []llamada
	responder func(n int, c tipos.Consulta, clase string) ([]tipos.Candidato, []string, error)
}

func (r *recuperadorFalso) Buscar(_ context.Context, c tipos.Consulta, clase string, k int) ([]tipos.Candidato, []string, error) {
	r.mu.Lock()
	r.llamadas = append(r.llamadas, llamada{c, clase, k})
	n := len(r.llamadas)
	r.mu.Unlock()
	if r.responder == nil {
		return nil, nil, nil
	}
	return r.responder(n, c, clase)
}

func (r *recuperadorFalso) total() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.llamadas)
}

// porTema: el recuperador «real» de las pruebas de integración: procedimientos y conceptos por palabras.
func porTema(_ int, c tipos.Consulta, clase string) ([]tipos.Candidato, []string, error) {
	q := c.Normalizada
	switch clase {
	case "procedimiento":
		switch {
		case strings.Contains(q, "unidades"):
			return []tipos.Candidato{{ID: "config.unidades", Clase: clase, Puntaje: 0.8, Lexico: 3}}, nil, nil
		case strings.Contains(q, "metrado"):
			return []tipos.Candidato{{ID: "metrados.registrar", Clase: clase, Puntaje: 0.9, Lexico: 4, Vector: 0.7}}, nil, nil
		}
	case "concepto":
		if strings.Contains(q, "metrado") {
			return []tipos.Candidato{{ID: "metrado", Clase: clase, Puntaje: 0.95, Lexico: 5}}, nil, nil
		}
	}
	return nil, nil, nil
}

// constructorFalso arma el plan desde el catálogo con el primer candidato (lo que hará conocimiento).
type constructorFalso struct {
	cat     *catalogoFalso
	err     error
	estados []tipos.Estado
}

func (c *constructorFalso) Construir(_ context.Context, e tipos.Estado, cands []tipos.Candidato) (tipos.Plan, error) {
	c.estados = append(c.estados, e)
	if c.err != nil {
		return tipos.Plan{}, c.err
	}
	if len(cands) == 0 {
		return PlanSinEvidencia(e.Tipo), nil
	}
	top := cands[0]
	switch top.Clase {
	case "concepto":
		def, ok := c.cat.Concepto(top.ID)
		if !ok {
			return PlanSinEvidencia(e.Tipo), nil
		}
		return tipos.Plan{Version: 2, Tipo: e.Tipo, Concepto: &def, Plantilla: "CONCEPTO",
			Fuentes: []tipos.FuentePlan{{Manual: "Glosario", Paginas: []int{3}}}}, nil
	case "procedimiento":
		def, ok := c.cat.Procedimiento(top.ID)
		if !ok {
			return PlanSinEvidencia(e.Tipo), nil
		}
		p := tipos.Plan{Version: 2, Tipo: e.Tipo, Procedimiento: &tipos.Ref{ID: def.ID, Titulo: def.Titulo, Confianza: top.Puntaje},
			Prerrequisitos: def.Prerrequisitos, Verificacion: def.Verificacion, Fuentes: fuentesPlan(def, nil),
			Plantilla: "PROCEDIMIENTO_INTRO"}
		for _, x := range def.Pasos {
			p.Pasos = append(p.Pasos, pasoPlan(x, x.N))
		}
		return p, nil
	}
	return tipos.Plan{Version: 2, Tipo: e.Tipo, Intro: "Según el manual: " + top.ID,
		Pasos:   []tipos.PasoPlan{{N: 1, ID: top.ID + "#1", Texto: "Paso del fragmento.", Fuente: []string{top.ID}}},
		Fuentes: []tipos.FuentePlan{{Manual: "Manual X", Paginas: []int{1}}}}, nil
}

// generadorFalso: el «LLM» que reformula (o cae).
type generadorFalso struct {
	nombre   string
	err      error
	llamadas int
}

func (g *generadorFalso) Nombre() string { return g.nombre }
func (g *generadorFalso) Redactar(_ context.Context, p tipos.Plan) (string, error) {
	g.llamadas++
	if g.err != nil {
		return "", g.err
	}
	return "LLM: " + RenderBasico(p), nil
}

// gateFalso pasa según una secuencia (true/false por intento); sin secuencia, pasa siempre.
type gateFalso struct {
	pasa   []bool
	textos []string
}

func (g *gateFalso) Evaluar(_ tipos.Plan, texto string, _ tipos.Estado) tipos.ResultadoCalidad {
	i := len(g.textos)
	g.textos = append(g.textos, texto)
	ok := true
	if i < len(g.pasa) {
		ok = g.pasa[i]
	}
	if ok {
		return tipos.ResultadoCalidad{Paso: true, Puntaje: 1}
	}
	return tipos.ResultadoCalidad{Paso: false, Puntaje: 0.4, Problemas: []string{"menú fuera del plan"}}
}

// charlaFalsa es la charla de V1.
type charlaFalsa struct {
	intenciones []string
}

func (c *charlaFalsa) Conversar(_ context.Context, intencion, pregunta string, _ []rag.Turno) (rag.Respuesta, error) {
	c.intenciones = append(c.intenciones, intencion)
	return rag.Respuesta{Pregunta: pregunta, Respuesta: "¡Hola! ¿En qué te ayudo?", Modo: "conversacional"}, nil
}

// aliasFalso: «metrado» → módulo Presupuestos y alias «cómputo».
type aliasFalso struct{}

func (aliasFalso) Analizar(texto string) ([]string, []string, string) {
	n := Normalizar(texto)
	if strings.Contains(n, "metrado") {
		return []string{"Metrado"}, []string{"computo de cantidades"}, "Presupuestos"
	}
	return nil, nil, ""
}

// armarAgente: un agente con todos los dobles y V2 forzada.
type banco struct {
	ag    *Agente
	rec   *recuperadorFalso
	cons  *constructorFalso
	cat   *catalogoFalso
	gen   *generadorFalso
	gate  *gateFalso
	charl *charlaFalsa
}

func armarAgente() *banco {
	cfg := ConfigDefecto()
	cfg.Version = tipos.V2
	cat := nuevoCatalogo()
	b := &banco{
		rec:   &recuperadorFalso{responder: porTema},
		cons:  &constructorFalso{cat: cat},
		cat:   cat,
		gate:  &gateFalso{},
		charl: &charlaFalsa{},
	}
	b.ag = &Agente{Config: cfg, Recuperador: b.rec, Constructor: b.cons, Catalogo: cat, Gate: b.gate,
		Charla: b.charl, Aliaser: aliasFalso{}}
	return b
}

var errCaido = errors.New("connection refused")
