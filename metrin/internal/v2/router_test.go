package v2

import (
	"context"
	"testing"

	"rag-go/internal/rag"
	"rag-go/internal/router"
	"rag-go/internal/v2/tipos"
)

// routerFalso devuelve siempre la misma decisión y cuenta las llamadas.
type routerFalso struct {
	d       router.Decision
	llamado int
}

func (r *routerFalso) Decidir(string) router.Decision { r.llamado++; return r.d }

// constructorEspia guarda los candidatos que recibe y delega en el falso de siempre.
type constructorEspia struct {
	*constructorFalso
	cands [][]tipos.Candidato
}

func (c *constructorEspia) Construir(ctx context.Context, e tipos.Estado, cands []tipos.Candidato) (tipos.Plan, error) {
	c.cands = append(c.cands, append([]tipos.Candidato(nil), cands...))
	return c.constructorFalso.Construir(ctx, e, cands)
}

func conRouter(d router.Decision) (*banco, *routerFalso, *constructorEspia) {
	b := armarAgente()
	rf := &routerFalso{d: d}
	esp := &constructorEspia{constructorFalso: b.cons}
	b.ag.Router, b.ag.Constructor = rf, esp
	return b, rf, esp
}

func procsDe(cs []tipos.Candidato) []tipos.Candidato {
	var out []tipos.Candidato
	for _, c := range cs {
		if c.Clase == "procedimiento" {
			out = append(out, c)
		}
	}
	return out
}

func TestRouterEligeYRestringe(t *testing.T) {
	b, rf, esp := conRouter(router.Decision{Accion: router.Elegir,
		Candidatos: []router.Candidato{{ID: "config.unidades", Probabilidad: 0.8}}})
	res, tr := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	comprobarV2(t, res)
	if rf.llamado != 1 {
		t.Fatalf("el router debió consultarse una vez: %d", rf.llamado)
	}
	ps := procsDe(esp.cands[len(esp.cands)-1])
	if len(ps) != 1 || ps[0].ID != "config.unidades" || ps[0].Puntaje < 0.6 || ps[0].Meta["router"] != router.Elegir {
		t.Fatalf("solo debe quedar el elegido, aceptable para el constructor: %+v", ps)
	}
	if res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "config.unidades" {
		t.Fatalf("plan: %+v", res.PlanV2)
	}
	if d := etapaV2(tr, "plan_consulta").Datos; d["router"] == nil {
		t.Fatalf("la decisión del router va en la traza: %+v", d)
	}
}

func TestRouterRespondeAunqueLaBusquedaNoHalle(t *testing.T) {
	b, _, _ := conRouter(router.Decision{Accion: router.Elegir,
		Candidatos: []router.Candidato{{ID: "metrados.registrar", Probabilidad: 0.9}}})
	b.rec.responder = nil // la búsqueda no trae nada
	res, _ := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	if res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "metrados.registrar" {
		t.Fatalf("con el router, la falta de evidencia léxica no es SIN_EVIDENCIA: %+v", res)
	}
}

func TestRouterDelegaNoCambiaNada(t *testing.T) {
	b, _, esp := conRouter(router.Decision{Accion: router.Delegar})
	res, _ := preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	ps := procsDe(esp.cands[len(esp.cands)-1])
	if len(ps) != 1 || ps[0].ID != "metrados.registrar" || ps[0].Meta["router"] != "" {
		t.Fatalf("delegar deja los candidatos de la búsqueda tal cual: %+v", ps)
	}
	if res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "metrados.registrar" {
		t.Fatalf("plan: %+v", res.PlanV2)
	}
}

func TestRouterEmpateDejaOpcionesIguales(t *testing.T) {
	b, _, esp := conRouter(router.Decision{Accion: router.AclararCaso, Candidatos: []router.Candidato{
		{ID: "metrados.registrar", Probabilidad: 0.4}, {ID: "config.unidades", Probabilidad: 0.35}}})
	preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	ps := procsDe(esp.cands[len(esp.cands)-1])
	if len(ps) != 2 || ps[0].Puntaje != ps[1].Puntaje || ps[0].Lexico != ps[1].Lexico || ps[0].Puntaje < 0.6 {
		t.Fatalf("las opciones deben empatar para que el constructor pregunte: %+v", ps)
	}
}

func TestRouterIgnoraIdsFueraDelCatalogo(t *testing.T) {
	b, _, esp := conRouter(router.Decision{Accion: router.Elegir,
		Candidatos: []router.Candidato{{ID: "no.existe", Probabilidad: 0.9}}})
	preguntarConTraza(t, b.ag, "¿Cómo registro un metrado?", rag.Opciones{})
	ps := procsDe(esp.cands[len(esp.cands)-1])
	if len(ps) != 1 || ps[0].ID != "metrados.registrar" {
		t.Fatalf("un id que el catálogo no conoce no debe tocar los candidatos: %+v", ps)
	}
}

func TestRouterNoSeUsaEnConceptos(t *testing.T) {
	b, rf, _ := conRouter(router.Decision{Accion: router.Elegir,
		Candidatos: []router.Candidato{{ID: "config.unidades", Probabilidad: 0.9}}})
	preguntarConTraza(t, b.ag, "¿Qué es un metrado?", rag.Opciones{})
	if rf.llamado != 0 {
		t.Fatalf("una pregunta de concepto no pasa por el router: %d llamadas", rf.llamado)
	}
}

func TestRouterFiltraErroresDelElegido(t *testing.T) {
	b, _, _ := conRouter(router.Decision{Accion: router.Elegir,
		Candidatos: []router.Candidato{{ID: "metrados.registrar", Probabilidad: 0.9}}})
	b.rec.responder = func(_ int, _ tipos.Consulta, clase string) ([]tipos.Candidato, []string, error) {
		if clase != "error" {
			return nil, nil, nil
		}
		return []tipos.Candidato{
			{ID: "config.unidades#error1", Clase: "error", Puntaje: 0.9, Meta: map[string]string{"procedimiento": "config.unidades"}},
			{ID: "metrados.registrar#error1", Clase: "error", Puntaje: 0.7, Meta: map[string]string{"procedimiento": "metrados.registrar"}},
		}, nil, nil
	}
	tr := &turno{a: b.ag, pregunta: "no me deja guardar el metrado"}
	out, usado := tr.rutear(tipos.Problema, func() []tipos.Candidato {
		cs, _, _ := b.rec.responder(1, tipos.Consulta{}, "error")
		return cs
	}())
	if !usado {
		t.Fatal("el router debió intervenir")
	}
	for _, c := range out {
		if c.Clase == "error" && c.Meta["procedimiento"] != "metrados.registrar" {
			t.Fatalf("quedó un error de otro procedimiento: %+v", out)
		}
	}
}

func TestRouterRescataTipoDesconocido(t *testing.T) {
	const q = "lo del metrado en la hoja ese"
	sin, _ := preguntarConTraza(t, armarAgente().ag, q, rag.Opciones{})
	if sin.PlanV2.Tipo != tipos.Desconocido {
		t.Skipf("las reglas ya clasifican %q: la prueba necesita una pregunta UNKNOWN", q)
	}
	b, _, _ := conRouter(router.Decision{Accion: router.Elegir,
		Candidatos: []router.Candidato{{ID: "metrados.registrar", Probabilidad: 0.9}}})
	res, _ := preguntarConTraza(t, b.ag, q, rag.Opciones{})
	if res.PlanV2.Procedimiento == nil || res.PlanV2.Procedimiento.ID != "metrados.registrar" || res.PlanV2.Tipo != tipos.Procedimiento {
		t.Fatalf("tipo UNKNOWN con el router seguro: debió responder el procedimiento como PROCEDURE: %+v", res.PlanV2)
	}
}

func TestRouterNoRescataPalabraSuelta(t *testing.T) {
	b, _, _ := conRouter(router.Decision{Accion: router.Elegir,
		Candidatos: []router.Candidato{{ID: "metrados.registrar", Probabilidad: 0.99}}})
	res, _ := preguntarConTraza(t, b.ag, "kardex", rag.Opciones{})
	if res.PlanV2.Procedimiento != nil {
		t.Fatalf("una palabra suelta no es una tarea: no debe responder un procedimiento: %+v", res.PlanV2)
	}
}

func TestRouterEmpateRescataTipoDesconocidoConOpciones(t *testing.T) {
	const q = "lo del metrado en la hoja ese"
	b, _, esp := conRouter(router.Decision{Accion: router.AclararCaso, Candidatos: []router.Candidato{
		{ID: "metrados.registrar", Probabilidad: 0.4}, {ID: "config.unidades", Probabilidad: 0.35}}})
	res, _ := preguntarConTraza(t, b.ag, q, rag.Opciones{})
	if res.PlanV2.Tipo != tipos.Procedimiento || len(esp.cands) == 0 || len(procsDe(esp.cands[len(esp.cands)-1])) != 2 {
		t.Fatalf("tipo UNKNOWN y empate del router: la V2 debe ver las dos opciones del router como PROCEDURE: %+v", res.PlanV2)
	}
}
