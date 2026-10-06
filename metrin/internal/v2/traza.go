package v2

// Instrumentación de la V2 en el modo traza (docs/CONTRATO-traza-metrin.md, «Etapas V2»). Todo pasa por el
// sub-registro traza.Recorder.V2(); sin modo traza es nil y estas funciones no hacen nada (o regresan al instante).
// Solo anotan lo que el orquestador ya decidió: no cambian la respuesta.

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"rag-go/internal/rag"
	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

const razonTurnoV2 = "turno V2: el camino está en «etapas_v2»"

// maxLista: tope de elementos de las listas que van en «datos».
const maxLista = 10

// idsBusqueda: etapas que puede anotar el propio recuperador (internal/v2/conocimiento, internal/busqueda); si no
// lo hace, las deriva el núcleo de los candidatos.
var idsBusqueda = []string{traza.EtapaV2BusquedaLexica, traza.EtapaV2BusquedaVectorial, traza.EtapaV2Fusion, traza.EtapaV2Rerank}

func itoa(n int) string { return strconv.Itoa(n) }

func recortarLista(xs []string) []string {
	if len(xs) > maxLista {
		return append(append([]string(nil), xs[:maxLista]...), "…")
	}
	return xs
}

func (t *turno) trazarTipo(r ResultadoTipo, final tipos.TipoRespuesta, metodo string, d time.Duration) {
	if t.v == nil {
		return
	}
	datos := traza.Datos{
		"tipo":       string(final),
		"propuesta":  string(r.Tipo),
		"confianza":  traza.Redondear(r.Confianza),
		"metodo":     metodo,
		"dudoso":     r.Dudoso(),
		"umbral":     t.a.Config.TipoUmbral,
		"margen":     t.a.Config.TipoMargen,
		"candidatos": tiposATexto(r.Candidatos),
		"fuentes":    nil,
	}
	if r.TipoKNN != "" {
		datos["knn"] = string(r.TipoKNN)
		datos["similitud_knn"] = traza.Redondear(r.SimKNN)
		datos["margen_knn"] = traza.Redondear(r.MargenKNN)
	}
	if r.IntencionV1 != "" {
		datos["intencion_v1"] = r.IntencionV1
		datos["similitud_v1"] = traza.Redondear(r.SimilitudV1)
	}
	modelo := "reglas"
	switch {
	case strings.HasPrefix(metodo, "decision:"):
		modelo = "motor de decisión: " + strings.TrimPrefix(metodo, "decision:")
	case metodo == "knn" && t.a.Clasificador != nil && t.a.Clasificador.Tipo != nil:
		modelo = "kNN-tipo:" + t.a.Clasificador.Tipo.HuellaEmb
	case metodo == "intencion_v1":
		modelo = "kNN-local (intención V1)"
	}
	razon := "«" + string(final) + "» por " + metodo + ", confianza " + traza.Decimal(r.Confianza, 2)
	if r.Dudoso() {
		razon = "ni regla clara ni kNN con margen: decidió el motor de decisión («" + string(final) + "»)"
	}
	t.v.FinDuracion(traza.EtapaV2TipoRespuesta, traza.EstadoOK, d, datos)
	t.v.Razon(traza.EtapaV2TipoRespuesta, razon)
	t.v.Modelo(traza.EtapaV2TipoRespuesta, modelo)
}

// trazarTipoMemoria: el tipo lo fija la memoria del procedimiento (sin clasificar ni buscar).
func (t *turno) trazarTipoMemoria(tipo tipos.TipoRespuesta, razon string) {
	t.metodo = "memoria"
	t.estado.Tipo, t.estado.TipoConfianza = tipo, 1
	if t.v == nil {
		return
	}
	t.v.Fin(traza.EtapaV2TipoRespuesta, traza.EstadoOK, traza.Datos{
		"tipo": string(tipo), "confianza": 1.0, "metodo": "memoria", "umbral": t.a.Config.TipoUmbral,
		"accion_memoria": string(t.accion), "fuentes": nil,
	})
	t.v.Razon(traza.EtapaV2TipoRespuesta, razon)
	t.v.Modelo(traza.EtapaV2TipoRespuesta, "memoria")
	t.trazarSinBusqueda(traza.EstadoNoTomada, "memoria del procedimiento: sin búsqueda nueva")
}

func (t *turno) anotarDecision(entrada map[string]any, d tipos.Decision, dur time.Duration, respaldo string) {
	t.msDecision += dur
	entrada["choice"] = d.Eleccion
	entrada["fuente"] = d.Fuente
	entrada["confianza"] = traza.Redondear(d.Confianza)
	entrada["ms"] = float64(dur.Microseconds()) / 1000
	if respaldo != "" {
		entrada["respaldo"] = traza.Recortar(respaldo)
		if t.respaldoDecision == "" {
			t.respaldoDecision = respaldo
		}
	}
	t.decisiones = append(t.decisiones, entrada)
}

func (t *turno) trazarSinBusqueda(estado, razon string) {
	if t.v == nil {
		return
	}
	for _, id := range append([]string{traza.EtapaV2PlanConsulta}, idsBusqueda...) {
		if t.v.Pendiente(id) {
			t.v.OmitirCon(id, estado, razon, nil)
		}
	}
}

// trazarBusquedaIteracion deriva las cuatro etapas de búsqueda de los candidatos (las que el recuperador no anotó).
func (t *turno) trazarBusquedaIteracion(antes map[string]bool, cands []tipos.Candidato, degradado, errores []string, it int, dur time.Duration) {
	if t.v == nil {
		return
	}
	derivar := func(id string) bool {
		if t.derivoBusqueda {
			return true
		}
		return antes[id] && t.v.Pendiente(id)
	}
	if !derivar(traza.EtapaV2Fusion) {
		return // el recuperador anota sus propias etapas
	}
	t.derivoBusqueda = true
	lex, vec, rr := 0, 0, 0
	for _, c := range cands {
		if c.Lexico > 0 {
			lex++
		}
		if c.Vector > 0 {
			vec++
		}
		if c.Rerank != nil {
			rr++
		}
	}
	tiene := func(subs ...string) string {
		for _, d := range degradado {
			l := strings.ToLower(d)
			for _, s := range subs {
				if strings.Contains(l, s) {
					return d
				}
			}
		}
		return ""
	}
	comunes := traza.Datos{"iteracion": it, "fuentes": nil, "confianza": nil}
	con := func(extra traza.Datos) traza.Datos {
		d := traza.Datos{}
		for k, v := range comunes {
			d[k] = v
		}
		for k, v := range extra {
			d[k] = v
		}
		return d
	}
	estado := func(deg string) string {
		if deg != "" {
			return traza.EstadoRespaldo
		}
		return traza.EstadoOK
	}
	dl := tiene("lexic", "bm25")
	t.v.Fin(traza.EtapaV2BusquedaLexica, estado(dl), con(traza.Datos{"resultados": lex}))
	t.v.Razon(traza.EtapaV2BusquedaLexica, siVacio(dl, "candidatos con puntaje léxico (derivado por el núcleo)"))
	dv := tiene("vector")
	t.v.Fin(traza.EtapaV2BusquedaVectorial, estado(dv), con(traza.Datos{"resultados": vec}))
	t.v.Razon(traza.EtapaV2BusquedaVectorial, siVacio(dv, "candidatos con puntaje vectorial (derivado por el núcleo)"))
	fusion := traza.Datos{"resultados": len(cands), "mejor": nil, "fuentes": len(cands), "degradado": recortarLista(degradado)}
	if len(cands) > 0 {
		fusion["mejor"] = traza.Redondear(cands[0].Puntaje)
		fusion["candidatos"] = candidatosTraza(cands)
	}
	if len(errores) > 0 {
		fusion["errores"] = recortarLista(errores)
	}
	estFusion, razonFusion := traza.EstadoOK, "candidatos ordenados por puntaje híbrido; ms = llamadas al recuperador"
	if len(errores) > 0 {
		estFusion, razonFusion = traza.EstadoError, "el recuperador falló: "+errores[0]
	}
	t.v.FinDuracion(traza.EtapaV2Fusion, estFusion, dur, con(fusion))
	t.v.Razon(traza.EtapaV2Fusion, razonFusion)
	drr := tiene("rerank")
	datosRR := traza.Datos{"activo": rr > 0, "resultados": rr}
	if rr > 0 {
		datosRR["candidatos"] = candidatosTraza(cands)
	}
	t.v.Fin(traza.EtapaV2Rerank, estado(drr), con(datosRR))
	t.v.Razon(traza.EtapaV2Rerank, siVacio(drr, "candidatos con puntaje del reranker (derivado por el núcleo)"))
}

// candidatosTraza: los mejores candidatos (hasta maxLista) con lo que necesita el benchmark (cmd/evalv2) para
// Recall@k, MRR y nDCG con el par (id, manual): id, clase, manual y puntaje.
func candidatosTraza(cands []tipos.Candidato) []map[string]any {
	out := make([]map[string]any, 0, min(len(cands), maxLista))
	for i, c := range cands {
		if i == maxLista {
			break
		}
		m := map[string]any{"id": traza.Recortar(c.ID), "clase": c.Clase, "manual": nulo(traza.Recortar(c.Meta["manual"])),
			"puntaje": traza.Redondear(c.Puntaje)}
		if c.Rerank != nil {
			m["rerank"] = traza.Redondear(*c.Rerank)
		}
		out = append(out, m)
	}
	return out
}

func siVacio(s, def string) string {
	if s != "" {
		return "degradado: " + s
	}
	return def
}

func (t *turno) trazarPlanConsulta(c tipos.Consulta, clases [][]string, iteraciones, resultados int, motivo string) {
	if t.v == nil {
		return
	}
	estado, razon := traza.EstadoOK, "evidencia suficiente en la iteración "+itoa(iteraciones)
	if motivo != "" {
		estado, razon = traza.EstadoAlerta, motivo
	}
	t.v.Fin(traza.EtapaV2PlanConsulta, estado, traza.Datos{
		"clases":      clases,
		"consulta":    c.Normalizada,
		"original":    c.Original,
		"entidades":   recortarLista(c.Entidades),
		"aliases":     recortarLista(c.Aliases),
		"modulo":      nulo(c.Modulo),
		"iteraciones": iteraciones,
		"expandida":   iteraciones > 1,
		"fuentes":     resultados,
		"confianza":   nil,
	})
	t.v.Razon(traza.EtapaV2PlanConsulta, razon)
}

func nulo(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// trazarPlan anota procedimiento, pasos, fotos por paso y plan.
func (t *turno) trazarPlan(p tipos.Plan) {
	if t.v == nil {
		return
	}
	if p.Procedimiento != nil {
		t.v.Fin(traza.EtapaV2Procedimiento, traza.EstadoOK, traza.Datos{
			"id": p.Procedimiento.ID, "titulo": p.Procedimiento.Titulo,
			"confianza":    traza.Redondear(p.Procedimiento.Confianza),
			"alternativas": recortarLista(t.altProcedimiento),
			"decidido_por": nulo(t.decididoProc),
			"fuentes":      len(p.Fuentes),
		})
		t.v.Razon(traza.EtapaV2Procedimiento, "«"+p.Procedimiento.Titulo+"»")
		t.v.Modelo(traza.EtapaV2Procedimiento, siVacioTexto(t.decididoProc, t.metodo))
	} else if t.v.Pendiente(traza.EtapaV2Procedimiento) {
		t.v.Omitir(traza.EtapaV2Procedimiento, "el plan no es de un procedimiento ("+string(p.Tipo)+")")
	}
	if len(p.Pasos) > 0 {
		mostrados := p.PasosMostrados
		if len(mostrados) == 0 {
			for _, x := range p.Pasos {
				mostrados = append(mostrados, x.N)
			}
		}
		ids := make([]string, 0, len(p.Pasos))
		fotos, conFoto := 0, 0
		var idsFotos []string
		for _, x := range p.Pasos {
			ids = append(ids, x.ID)
		}
		for _, x := range PasosVisibles(p) {
			fotos += len(x.Fotos)
			if len(x.Fotos) > 0 {
				conFoto++
			}
			for _, f := range x.Fotos {
				idsFotos = append(idsFotos, itoa(x.N)+":"+f.ID)
			}
		}
		t.v.Fin(traza.EtapaV2Pasos, traza.EstadoOK, traza.Datos{
			"pasos": len(p.Pasos), "mostrados": mostrados, "ids": recortarLista(ids), "confianza": nil, "fuentes": len(p.Fuentes),
		})
		t.v.Razon(traza.EtapaV2Pasos, "pasos del procedimiento, en orden")
		estFotos, razonFotos := traza.EstadoOK, "solo las fotos que la fuente asocia a cada paso"
		if fotos == 0 {
			razonFotos = "ningún paso trae foto asociada en la fuente (mejor sin foto que una ajena)"
		}
		t.v.Fin(traza.EtapaV2FotosPaso, estFotos, traza.Datos{
			"fotos": fotos, "pasos_con_foto": conFoto, "ids": recortarLista(idsFotos), "confianza": nil, "fuentes": nil,
		})
		t.v.Razon(traza.EtapaV2FotosPaso, razonFotos)
	} else {
		for _, id := range []string{traza.EtapaV2Pasos, traza.EtapaV2FotosPaso} {
			t.v.OmitirCon(id, traza.EstadoOmitida, "el plan no trae pasos ("+string(p.Tipo)+")", nil)
		}
	}
	datos := traza.Datos{
		"tipo": string(p.Tipo), "plantilla": nulo(p.Plantilla), "sin_evidencia": p.SinEvidencia,
		"fuentes": len(p.Fuentes), "confianza": nil, "afirmaciones": len(p.Afirmaciones),
	}
	if p.Procedimiento != nil {
		datos["confianza"] = traza.Redondear(p.Procedimiento.Confianza)
	}
	if p.Siguiente != nil {
		datos["siguiente"] = p.Siguiente.Plantilla
	}
	razon := "plan " + string(p.Tipo)
	if p.SinEvidencia {
		razon = "SIN_EVIDENCIA: nada en kb/ respalda una respuesta"
	}
	t.v.Fin(traza.EtapaV2Plan, traza.EstadoOK, datos)
	t.v.Razon(traza.EtapaV2Plan, razon)
}

func siVacioTexto(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

func (t *turno) trazarGate(nombre string, final, primero tipos.ResultadoCalidad, d time.Duration) {
	if t.v == nil {
		return
	}
	estado, razon := traza.EstadoOK, "plan y texto respaldados por la evidencia"
	switch {
	case !final.Paso:
		estado, razon = traza.EstadoError, strings.Join(final.Problemas, "; ")
	case t.gateIntentos > 1:
		estado, razon = traza.EstadoAlerta, "el primer texto falló ("+strings.Join(primero.Problemas, "; ")+"); pasó al regenerar sin LLM"
	}
	t.v.FinDuracion(traza.EtapaV2QualityGate, estado, d, traza.Datos{
		"passed": final.Paso, "score": traza.Redondear(final.Puntaje), "issues": recortarLista(final.Problemas),
		"intentos": t.gateIntentos, "gate": nombre, "confianza": traza.Redondear(final.Puntaje), "fuentes": nil,
	})
	t.v.Razon(traza.EtapaV2QualityGate, siVacioTexto(razon, "rechazado"))
	t.v.Modelo(traza.EtapaV2QualityGate, "gate "+nombre)
}

// claveHibrido: el dato con las cifras de la búsqueda híbrida de fragmentos (conocimiento.ClaveTrazaHibrido).
const claveHibrido = "hibrido"

// trazarHibrido integra la búsqueda híbrida de fragmentos en las etapas de búsqueda. El recuperador deja sus cifras
// reales (busqueda.Informe) en «hibrido» de las cuatro etapas sin cerrarlas, porque puede correr en una iteración
// del núcleo o desde el constructor, cuando las etapas ya se derivaron de los procedimientos. Aquí decide el estado:
// si el vector corrió sobre los fragmentos, la búsqueda vectorial no fue «solo léxica» (aunque procedimientos y
// conceptos se busquen solo por BM25); si el reranker reordenó, el reranker está activo; si fallaron, respaldo.
func (t *turno) trazarHibrido() {
	if t.v == nil {
		return
	}
	dato := func(id string) traza.Datos {
		x, _ := t.v.DatoDe(id, claveHibrido)
		m, _ := x.(traza.Datos)
		return m
	}
	fu := dato(traza.EtapaV2Fusion)
	if fu == nil {
		return
	}
	cerrar := func(id, estado, razon string) {
		if t.v.Pendiente(id) {
			t.v.Fin(id, estado, traza.Datos{"iteracion": nil, "fuentes": nil, "confianza": nil})
		}
		t.v.CambiarEstado(id, estado, razon)
	}
	if e, _ := fu["error"].(string); e != "" {
		cerrar(traza.EtapaV2Fusion, traza.EstadoRespaldo, "fragmentos: la búsqueda híbrida falló ("+e+"): solo léxica propia")
		return
	}
	if t.v.Pendiente(traza.EtapaV2Fusion) {
		cerrar(traza.EtapaV2Fusion, traza.EstadoOK, "fragmentos: RRF de BM25F y vector ("+itoa(aEntero(fu["resultados"]))+" resultados)")
	}
	if t.v.Pendiente(traza.EtapaV2BusquedaLexica) {
		cerrar(traza.EtapaV2BusquedaLexica, traza.EstadoOK, "fragmentos: BM25F sobre kb/fragmentos*.jsonl")
	}
	if vec := dato(traza.EtapaV2BusquedaVectorial); vec != nil {
		activo, _ := vec["activo"].(bool)
		switch fallo, _ := vec["fallo"].(string); {
		case fallo != "":
			cerrar(traza.EtapaV2BusquedaVectorial, traza.EstadoRespaldo, "degradado: el vector de fragmentos falló ("+fallo+"): solo BM25F")
		case activo:
			cerrar(traza.EtapaV2BusquedaVectorial, traza.EstadoOK, "vector sobre los fragmentos: "+itoa(aEntero(vec["resultados"]))+
				" resultados; procedimientos y conceptos: solo léxica")
		}
	}
	if rr := dato(traza.EtapaV2Rerank); rr != nil {
		activo, _ := rr["activo"].(bool)
		reordenado, _ := rr["reordenado"].(bool)
		switch fallo, _ := rr["fallo"].(string); {
		case fallo != "":
			cerrar(traza.EtapaV2Rerank, traza.EstadoRespaldo, "degradado: el reranker falló ("+fallo+"): orden RRF")
		case reordenado:
			cerrar(traza.EtapaV2Rerank, traza.EstadoOK, "reranker sobre los primeros fragmentos de la fusión")
			t.v.Dato(traza.EtapaV2Rerank, "activo", true)
		case !activo:
			cerrar(traza.EtapaV2Rerank, traza.EstadoRespaldo, "sin reranker en los fragmentos (RERANK_URL vacío o «fragmento» fuera de V2_RERANK_CLASES): orden por puntaje y RRF")
		}
	}
}

// claveRerankClases: el dato que el recuperador deja en la etapa rerank con el reranker de clase (procedimiento,
// concepto; conocimiento.ClaveTrazaRerankClases): una entrada por clase con reordenado, fallo, orden léxico y orden
// del reranker.
const claveRerankClases = "clases"

// trazarRerankClases ajusta la etapa rerank con el reranker de clase. Va después de trazarHibrido: si una clase se
// reordenó, la etapa no es «sin reranker» aunque los fragmentos no lo usen; si falló en alguna, es respaldo (el
// recuperador se quedó con el orden léxico de esa clase).
func (t *turno) trazarRerankClases() {
	if t.v == nil {
		return
	}
	x, _ := t.v.DatoDe(traza.EtapaV2Rerank, claveRerankClases)
	m, _ := x.(map[string]any)
	if len(m) == 0 {
		return
	}
	clases := make([]string, 0, len(m))
	for k := range m {
		clases = append(clases, k)
	}
	sort.Strings(clases)
	var hechas, fallidas []string
	for _, k := range clases {
		d, _ := m[k].(map[string]any)
		if r, _ := d["reordenado"].(bool); r {
			hechas = append(hechas, k)
		} else if f, _ := d["fallo"].(string); f != "" {
			fallidas = append(fallidas, k+" ("+f+")")
		}
	}
	if t.v.Pendiente(traza.EtapaV2Rerank) {
		t.v.Fin(traza.EtapaV2Rerank, traza.EstadoOK, traza.Datos{"activo": len(hechas) > 0, "resultados": nil,
			"iteracion": nil, "fuentes": nil, "confianza": nil})
	}
	hibridoFallo := false
	if h, ok := t.v.DatoDe(traza.EtapaV2Rerank, claveHibrido); ok {
		if hm, ok := h.(map[string]any); ok {
			f, _ := hm["fallo"].(string)
			hibridoFallo = f != ""
		}
	}
	switch {
	case len(fallidas) > 0:
		t.v.CambiarEstado(traza.EtapaV2Rerank, traza.EstadoRespaldo, "degradado: el reranker falló en "+strings.Join(fallidas, ", ")+
			": orden léxico")
	case len(hechas) > 0 && !hibridoFallo:
		t.v.CambiarEstado(traza.EtapaV2Rerank, traza.EstadoOK, "reranker para elegir: "+strings.Join(hechas, ", "))
		t.v.Dato(traza.EtapaV2Rerank, "activo", true)
	}
}

func aEntero(x any) int {
	switch n := x.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}

// cerrarTraza: decisión, redacción y las etapas de V1 (que en un turno V2 no corren, salvo la charla).
func (t *turno) cerrarTraza(res rag.Respuesta) {
	if t.v == nil {
		return
	}
	t.trazarHibrido()
	t.trazarRerankClases()
	motor := t.a.motor().Nombre()
	if len(t.decisiones) == 0 {
		t.v.OmitirCon(traza.EtapaV2Decision, traza.EstadoOmitida, "ninguna decisión dudosa: decidieron las reglas del código",
			traza.Datos{"motor": motor, "llamadas": 0, "decisiones": []map[string]any{}, "respaldo": false})
	} else {
		estado, razon := traza.EstadoOK, itoa(len(t.decisiones))+" decisión(es) con el motor «"+motor+"»"
		if t.respaldoDecision != "" {
			estado, razon = traza.EstadoRespaldo, t.respaldoDecision
		}
		ultima := t.decisiones[len(t.decisiones)-1]
		t.v.FinDuracion(traza.EtapaV2Decision, estado, t.msDecision, traza.Datos{
			"motor": motor, "llamadas": len(t.decisiones), "decisiones": t.decisiones,
			"respaldo": t.respaldoDecision != "", "confianza": ultima["confianza"], "fuentes": nil,
		})
		t.v.Razon(traza.EtapaV2Decision, razon)
		t.v.Modelo(traza.EtapaV2Decision, motor)
	}
	estado, razon := traza.EstadoOK, "texto desde el plan ("+t.modoRedaccion+")"
	if len(t.degradado) > 0 {
		for _, d := range t.degradado {
			if strings.Contains(d, "LLM") || strings.Contains(d, "plantillas") {
				estado = traza.EstadoRespaldo
			}
		}
		razon = strings.Join(t.degradado, "; ")
	}
	if t.modoRedaccion == "segura" {
		estado, razon = traza.EstadoRespaldo, siVacioTexto(res.Motivo, "respuesta segura")
	}
	t.v.FinDuracion(traza.EtapaV2Renderizado, estado, t.msLLM, traza.Datos{
		"generador": nulo(t.generador), "modo": nulo(t.modoRedaccion), "caracteres": len([]rune(res.Respuesta)),
		"presupuesto": t.pres.Uso(), "degradado": recortarLista(t.degradado),
		"fuentes": len(res.Fuentes), "confianza": nil,
	})
	t.v.Razon(traza.EtapaV2Renderizado, razon)
	t.v.Modelo(traza.EtapaV2Renderizado, siVacioTexto(t.generador, "—"))

	// Etapas de V1: entrada, ruta y fin con los datos del turno; el resto, omitidas (salvo la charla, que la
	// anota rag.Conversar en los mensajes SOCIAL, y «fotos», que la anota el servidor).
	rec := t.rec
	if rec.Pendiente(traza.EtapaEntrada) {
		rec.Fin(traza.EtapaEntrada, traza.EstadoOK, traza.Datos{
			"pregunta": t.pregunta, "hilo_turnos": len(t.o.Hilo), "k": t.o.K, "version": string(tipos.V2),
		})
	}
	if rec.Pendiente(traza.EtapaRuta) {
		rec.Fin(traza.EtapaRuta, traza.EstadoOK, traza.Datos{
			"ruta": "v2", "tipo_consulta": res.Plan.TipoConsulta, "intencion": res.Plan.Intencion,
		})
		rec.Razon(traza.EtapaRuta, "agente V2: tipo de respuesta «"+string(t.estado.Tipo)+"»")
	}
	for _, id := range traza.IDs() {
		if id == traza.EtapaFin || id == traza.EtapaFotos || !rec.Pendiente(id) {
			continue
		}
		datos := traza.Datos(nil)
		if id == traza.EtapaJEV {
			datos = traza.Datos{"activo": false}
		}
		if id == traza.EtapaRegistroFallos {
			datos = traza.Datos{"registrado": false}
		}
		rec.OmitirCon(id, traza.EstadoOmitida, razonTurnoV2, datos)
	}
	rec.Fin(traza.EtapaFin, traza.EstadoOK, traza.Datos{
		"modo": res.Modo, "ruta": "v2", "tipo_consulta": res.Plan.TipoConsulta, "sin_contexto": res.SinContexto,
		"fuentes": len(res.Fuentes), "motivo": nulo(res.Motivo), "version": string(tipos.V2),
	})
}
