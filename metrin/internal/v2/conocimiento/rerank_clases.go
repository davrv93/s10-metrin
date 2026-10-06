package conocimiento

// Reranker para ELEGIR procedimientos y conceptos (V2_RERANK_CLASES; medido en metrin/eval/BUSQUEDA.md §12).
//
// El BM25 propio de cada clase trae el candidato correcto entre sus primeros con mucha holgura, pero no siempre
// primero ni con cobertura suficiente para pasar el umbral del constructor. Con el reranker, los TopRerank primeros
// del orden léxico se puntúan con un cross-encoder (bge-reranker-v2-m3 por llama-server) frente a la pregunta, sobre
// un texto representativo de cada uno, y ese puntaje decide el orden y pasa a la escala del constructor:
//
//   - Texto: procedimiento → título (y módulo), objetivo, aliases y preguntas; concepto → término, sinónimos y
//     definición. Recortado a RunasRerank runas (el cliente de llama-server recorta otra vez con el mismo tope).
//   - Puntaje: el logit del reranker pasa por la sigmoide (aUnidad, la misma conversión que la clase «fragmento»),
//     así que los umbrales del constructor (0,6 para aceptar, 0,4 para ofrecer) quedan como probabilidad de que el
//     candidato responda a la pregunta: 0,6 ⇔ logit ≥ 0,405. El logit crudo va en Candidato.Rerank. Ese puntaje
//     ORDENA. Para ACEPTAR, el constructor usa el mayor entre él y la cobertura léxica del mismo candidato
//     (Constructor.candidatos): el reranker suma lo que la cobertura no alcanzaba, pero no quita lo que ya aceptaba.
//     Medido: con la sigmoide sola, «¿dónde está…?» y las de configuración daban logit ≈ 0 al MISMO candidato que el
//     léxico ponía primero y caían a SIN_EVIDENCIA (BUSQUEDA.md §12, variante B frente a B′).
//   - Empate: con dos candidatos reordenados, el constructor y el núcleo comparan logits (MargenRerankEmpate), no la
//     sigmoide, que se satura cerca de 1.
//   - Degradación: sin reranker, con error o con el tiempo agotado, el orden y el puntaje léxicos de siempre; el
//     error queda en «degradado» (DegRerankerError) y en la traza (etapa rerank, dato «clases»).
//   - Candidatos de la búsqueda híbrida: no se suman. La híbrida indexa fragmentos, no procedimientos, y el BM25 propio
//     ya trae el esperado entre sus 20 primeros en todas las preguntas de un turno del benchmark (BUSQUEDA.md §12).

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"rag-go/internal/traza"
	"rag-go/internal/v2/tipos"
)

// Valores por defecto del reranker de clase (los de la clase «fragmento»: RERANK_TOP_N y RERANK_MAX_RUNAS).
const (
	TopRerankDefecto   = 20
	RunasRerankDefecto = 800
)

// ClasesRerankDefecto: las clases del BM25 propio que el reranker puede reordenar. «error» (síntomas de
// TROUBLESHOOTING) no está: no se midió.
var ClasesRerankDefecto = map[string]bool{ClaseProcedimiento: true, ClaseConcepto: true}

// MargenRerankEmpate: diferencia de logits bajo la cual dos candidatos reordenados empatan. Es logit(0,6): el
// reranker prefiere al primero sobre el segundo (σ(l₀ − l₁)) con menos confianza que la que el constructor exige
// para aceptar un candidato (UmbralProcedimiento = 0,6). No se calibró con el benchmark.
var MargenRerankEmpate = math.Log(0.6 / 0.4)

// ClaveTrazaRerankClases: el dato de la etapa rerank con lo que hizo el reranker en cada clase.
const ClaveTrazaRerankClases = "clases"

// rerankActivo: hay reranker y la clase está entre las que reordena.
func (r *Recuperador) rerankActivo(clase string) bool {
	if r.Reranker == nil || !ClasesRerankDefecto[clase] {
		return false
	}
	if r.ClasesRerank == nil {
		return true
	}
	return r.ClasesRerank[clase]
}

func (r *Recuperador) topRerank() int {
	if r.TopRerank > 0 {
		return r.TopRerank
	}
	return TopRerankDefecto
}

func (r *Recuperador) runasRerank() int {
	if r.RunasRerank > 0 {
		return r.RunasRerank
	}
	return RunasRerankDefecto
}

// TextoRerank: el texto representativo de un procedimiento o concepto, recortado a n runas (0 = sin recorte).
func (b *Base) TextoRerank(clase, id string, n int) string {
	var ls []string
	switch clase {
	case ClaseProcedimiento:
		p, ok := b.porProc[id]
		if !ok {
			return ""
		}
		t := strings.TrimSpace(p.Titulo)
		if m := strings.TrimSpace(firstNonEmpty(p.Entidades.Modulo, p.Modulo)); m != "" {
			t += " (" + m + ")"
		}
		ls = append(ls, t, strings.TrimSpace(p.Objetivo))
		if len(p.Aliases) > 0 {
			ls = append(ls, strings.Join(p.Aliases, "; "))
		}
		ls = append(ls, p.Preguntas...)
	case ClaseConcepto:
		c, ok := b.porConcepto[id]
		if !ok {
			return ""
		}
		t := strings.TrimSpace(c.Termino)
		if len(c.Sinonimos) > 0 {
			t += " (" + strings.Join(c.Sinonimos, ", ") + ")"
		}
		ls = append(ls, t+": "+strings.TrimSpace(c.Definicion))
	default:
		return ""
	}
	var limpias []string
	for _, l := range ls {
		if l = strings.TrimSpace(l); l != "" {
			limpias = append(limpias, l)
		}
	}
	return recortarRunas(strings.Join(limpias, "\n"), n)
}

// recortarRunas: los primeros n runas, cortando en el último espacio si lo hay cerca (no a media palabra).
func recortarRunas(s string, n int) string {
	if n <= 0 || utf8.RuneCountInString(s) <= n {
		return s
	}
	rs := []rune(s)[:n]
	if i := strings.LastIndexAny(string(rs), " \n"); i > len(string(rs))*3/4 {
		return strings.TrimSpace(string(rs)[:i])
	}
	return string(rs)
}

// reordenarClase reordena en su sitio los primeros topRerank() de cands (en orden léxico) con el reranker y les pone
// el puntaje del reranker en [0, 1]. Lo demás queda detrás, en su orden. Devuelve lo degradado (nada si salió bien).
func (r *Recuperador) reordenarClase(ctx context.Context, c tipos.Consulta, clase string, ix *indiceBM25, cands []*candLex, k int) []string {
	n := min(len(cands), max(r.topRerank(), k))
	if n == 0 {
		return nil
	}
	docs := make([]string, n)
	for i, cd := range cands[:n] {
		docs[i] = r.base.TextoRerank(clase, ix.docs[cd.doc].id, r.runasRerank())
		if docs[i] == "" {
			docs[i] = recortarRunas(ix.docs[cd.doc].texto, r.runasRerank())
		}
	}
	q := textoHibrido(c)
	antes := idsCand(ix, cands[:min(n, maxCandTraza)])
	t0 := time.Now()
	ps, err := r.Reranker.Reordenar(ctx, q, docs)
	dur := time.Since(t0)
	if err == nil && len(ps) != n {
		err = fmt.Errorf("devolvió %d puntajes para %d documentos", len(ps), n)
	}
	if err == nil {
		for _, p := range ps {
			if math.IsNaN(p) || math.IsInf(p, 0) {
				err = fmt.Errorf("puntaje no finito")
				break
			}
		}
	}
	if err != nil {
		trazarRerankClase(ctx, clase, traza.Datos{"activo": true, "reordenado": false, "candidatos": n, "consulta": traza.Recortar(q),
			"ms": traza.Redondear(ms(dur)), "fallo": traza.ResumirError(err), "orden_lexico": antes})
		return []string{DegRerankerError + " (" + clase + "): " + traza.ResumirError(err) + "; orden y puntaje léxicos"}
	}
	for i, cd := range cands[:n] {
		v := ps[i]
		cd.rerank = &v
		cd.puntaje = aUnidad(v)
	}
	cabeza := cands[:n]
	sort.SliceStable(cabeza, func(a, b int) bool { return *cabeza[a].rerank > *cabeza[b].rerank })
	despues := make([]map[string]any, 0, min(n, maxCandTraza))
	for i, cd := range cabeza {
		if i == maxCandTraza {
			break
		}
		despues = append(despues, map[string]any{"id": traza.Recortar(ix.docs[cd.doc].id), "rerank": traza.Redondear(*cd.rerank),
			"puntaje": traza.Redondear(cd.puntaje), "rango_lexico": cd.rangoLex})
	}
	trazarRerankClase(ctx, clase, traza.Datos{"activo": true, "reordenado": true, "candidatos": n, "consulta": traza.Recortar(q),
		"ms": traza.Redondear(ms(dur)), "fallo": nil, "orden_lexico": antes, "orden": despues,
		"cambio_primero": cabeza[0].rangoLex != 1})
	return nil
}

func idsCand(ix *indiceBM25, cs []*candLex) []string {
	out := make([]string, 0, len(cs))
	for _, cd := range cs {
		out = append(out, traza.Recortar(ix.docs[cd.doc].id))
	}
	return out
}

// trazarRerankClase deja lo que hizo el reranker en la clase (una entrada por clase; cada llamada suma «llamadas»).
// No cierra la etapa: la cierra el núcleo (internal/v2/traza.go).
func trazarRerankClase(ctx context.Context, clase string, d traza.Datos) {
	v := traza.De(ctx).V2()
	if v == nil {
		return
	}
	todas := map[string]any{}
	if x, ok := v.DatoDe(traza.EtapaV2Rerank, ClaveTrazaRerankClases); ok {
		if m, ok := x.(map[string]any); ok {
			for k, y := range m {
				todas[k] = y
			}
		}
	}
	llamadas := 1
	if prev, ok := todas[clase].(traza.Datos); ok {
		if n, ok := prev["llamadas"].(int); ok {
			llamadas = n + 1
		}
	}
	d["llamadas"] = llamadas
	todas[clase] = d
	v.Dato(traza.EtapaV2Rerank, ClaveTrazaRerankClases, todas)
}

// EmpateRerank: los dos candidatos vienen del reranker y su diferencia de logits no llega a MargenRerankEmpate.
// ok=false si alguno de los dos no tiene puntaje del reranker (entonces decide la regla léxica de siempre).
func EmpateRerank(a, b tipos.Candidato) (empate, ok bool) {
	if a.Rerank == nil || b.Rerank == nil {
		return false, false
	}
	return *a.Rerank-*b.Rerank < MargenRerankEmpate, true
}
