package eval

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rag-go/internal/rag"
)

func caso(esperado string) Caso {
	return Caso{Pregunta: "¿prueba?", Esperado: esperado}
}

func trazaSugerida(noul, confianza float64) *TrazaDecision {
	return &TrazaDecision{Modo: "respuesta", Noul: noul, Confianza: confianza, MS: 84}
}

func TestVeredictoResponderConRespaldo(t *testing.T) {
	res := &rag.Respuesta{Modo: "respuesta"}
	r := veredicto(caso("responder"), res, nil, trazaSugerida(0.9, 0.7), Opciones{Umbral: 0.5})
	if !r.FlujoOK || !r.JevOK || !r.Acierto {
		t.Fatalf("esperado acierto, got %+v", r)
	}
}

func TestVeredictoResponderConNoulBajo(t *testing.T) {
	res := &rag.Respuesta{Modo: "respuesta"}
	r := veredicto(caso("responder"), res, nil, trazaSugerida(0.2, 0.6), Opciones{Umbral: 0.5})
	if r.FlujoOK != true || r.JevOK != false || r.Acierto != false {
		t.Fatalf("flujo sí, JEV no, acierto no; got %+v", r)
	}
}

func TestVeredictoAbstenerCorrecto(t *testing.T) {
	res := &rag.Respuesta{SinContexto: true, Modo: "sin_contexto"}
	r := veredicto(caso("abstener"), res, nil, trazaSugerida(0.95, 0.8), Opciones{Umbral: 0.5})
	if !r.FlujoOK || !r.Acierto {
		t.Fatalf("esperado acierto por abstención, got %+v", r)
	}
}

func TestVeredictoAbstenerNoSeAbstiene(t *testing.T) {
	res := &rag.Respuesta{Modo: "respuesta"}
	r := veredicto(caso("abstener"), res, nil, trazaSugerida(0.9, 0.7), Opciones{Umbral: 0.5})
	if r.FlujoOK || r.Acierto {
		t.Fatalf("el flujo respondió donde debía abstenerse, got %+v", r)
	}
}

func TestVeredictoSinTraza(t *testing.T) {
	res := &rag.Respuesta{Modo: "respuesta"}
	r := veredicto(caso("responder"), res, nil, nil, Opciones{Umbral: 0.5})
	if !r.SinTraza || r.Acierto || r.JevOK {
		t.Fatalf("sin traza no puede acertar, got %+v", r)
	}
}

func TestVeredictoError(t *testing.T) {
	r := veredicto(caso("responder"), nil, errPrueba{}, trazaSugerida(0.9, 0.7), Opciones{Umbral: 0.5})
	if r.Flujo != "error" || r.Acierto || r.Error == "" {
		t.Fatalf("esperado error, got %+v", r)
	}
}

type errPrueba struct{}

func (errPrueba) Error() string { return "fallo de red" }

const trazaJSONL = `{"id":"01","ts":"2026-10-01T12:00:00Z","modelo":"m","origen":"metrin-eval","ms":42,"estado":{"modo":"sin_contexto","pregunta":"¿x?"},"preguntas":[{"id":"q","tipo":"noul","instrucciones":"¿Era correcto abstenerse?","criterios":{"true":"sí"},"respuesta":{"tipo":"noul","noul":0.91,"probabilidades":{"true":0.91,"false":0.09}}}]}
{"id":"02","ts":"2026-10-01T12:00:01Z","origen":"metrin-eval","ms":58,"estado":{"modo":"respuesta"},"preguntas":[{"id":"q","tipo":"noul","instrucciones":"¿Respaldo?","respuesta":{"tipo":"noul","noul":0.44,"confianza":null}}]}
`

func TestParseTrazasVisor(t *testing.T) {
	f := filepath.Join(t.TempDir(), "trazas.jsonl")
	if err := os.WriteFile(f, []byte(trazaJSONL), 0o644); err != nil {
		t.Fatal(err)
	}
	trazas, err := leerTrazasNuevas(f, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(trazas) != 2 {
		t.Fatalf("esperadas 2 trazas, got %d", len(trazas))
	}
	d := ultimaDecision(trazas)
	if d == nil || d.Modo != "respuesta" || d.Noul != 0.44 || d.Confianza != 0 || d.MS != 58 {
		t.Fatalf("última decisión mala: %+v", d)
	}
	// La primera traza tiene confianza ausente (null) y modo sin_contexto.
	d0 := ultimaDecision(trazas[:1])
	if d0.Modo != "sin_contexto" || d0.Noul != 0.91 {
		t.Fatalf("primera decisión mala: %+v", d0)
	}
}

func TestLeerTrazasNuevasDesdeOffset(t *testing.T) {
	f := filepath.Join(t.TempDir(), "trazas.jsonl")
	if err := os.WriteFile(f, []byte(trazaJSONL), 0o644); err != nil {
		t.Fatal(err)
	}
	primera := strings.Index(trazaJSONL, "\n") + 1
	trazas, err := leerTrazasNuevas(f, int64(primera))
	if err != nil {
		t.Fatal(err)
	}
	if len(trazas) != 1 || trazas[0].ID != "02" {
		t.Fatalf("esperada solo la traza 02, got %+v", trazas)
	}
}

func TestLeerTrazasInexistente(t *testing.T) {
	trazas, err := leerTrazasNuevas(filepath.Join(t.TempDir(), "no.jsonl"), 0)
	if err != nil || len(trazas) != 0 {
		t.Fatalf("archivo inexistente debe ser vacío y sin error, got %v %v", trazas, err)
	}
}

func TestUltimaDecisionSinRespuesta(t *testing.T) {
	var trazas []trazaVisor
	if err := json.Unmarshal([]byte(`[{"id":"a","preguntas":[{"id":"q","respuesta":null}]}]`), &trazas); err != nil {
		t.Fatal(err)
	}
	if d := ultimaDecision(trazas); d != nil {
		t.Fatalf("traza sin respuesta no produce decisión, got %+v", d)
	}
}

func TestCargarCasos(t *testing.T) {
	f := filepath.Join(t.TempDir(), "casos.jsonl")
	txt := `{"pregunta":"¿a?","esperado":"responder","nota":"x"}
# comentario
{"pregunta":"¿b?","esperado":"abstener"}
`
	if err := os.WriteFile(f, []byte(txt), 0o644); err != nil {
		t.Fatal(err)
	}
	casos, err := CargarCasos(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(casos) != 2 || casos[0].Esperado != "responder" || casos[1].Esperado != "abstener" {
		t.Fatalf("casos mal cargados: %+v", casos)
	}
}

func TestCargarCasosVeredictoInvalido(t *testing.T) {
	f := filepath.Join(t.TempDir(), "casos.jsonl")
	if err := os.WriteFile(f, []byte(`{"pregunta":"¿a?","esperado":"quizás"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := CargarCasos(f); err == nil {
		t.Fatal("esperado error por veredicto inválido")
	}
}

func TestResumenAgregados(t *testing.T) {
	inf := Informe{Resultados: []Resultado{
		{Caso: caso("responder"), Flujo: "responde", FlujoOK: true, Traza: &TrazaDecision{Noul: 0.9, Confianza: 0.6, MS: 100}, JevOK: true, Acierto: true},
		{Caso: caso("responder"), Flujo: "responde", FlujoOK: true, Traza: &TrazaDecision{Noul: 0.8, Confianza: 0.8, MS: 200}, JevOK: true, Acierto: true},
		{Caso: caso("abstener"), Flujo: "abstiene", FlujoOK: true, Traza: &TrazaDecision{Noul: 0.4, Confianza: 0.5, MS: 300}, JevOK: false},
		{Caso: caso("responder"), Flujo: "responde", FlujoOK: true, SinTraza: true},
	}}
	inf = resumen(inf)
	if inf.Total != 4 || inf.Aciertos != 2 || inf.FlujoOK != 4 || inf.JevOK != 2 || inf.SinTraza != 1 {
		t.Fatalf("conteos mal: %+v", inf)
	}
	if math.Abs(inf.ConfianzaMedia-0.6333) > 0.001 {
		t.Fatalf("confianza media %f", inf.ConfianzaMedia)
	}
	if inf.LatenciaMediaMS != 200 || inf.LatenciaP95MS != 300 {
		t.Fatalf("latencias: media %d p95 %d", inf.LatenciaMediaMS, inf.LatenciaP95MS)
	}
}

func TestFormatear(t *testing.T) {
	inf := Informe{Total: 1, Aciertos: 1, FlujoOK: 1, JevOK: 1, ConfianzaMedia: 0.7,
		Resultados: []Resultado{{Caso: caso("responder"), Flujo: "responde", FlujoOK: true,
			Traza: &TrazaDecision{Noul: 0.9, Confianza: 0.7, MS: 84, Instrucciones: "¿Respaldo?"}, JevOK: true, Acierto: true}}}
	s := Formatear(inf, "datos/eval.jsonl", "metrin-eval")
	for _, quiere := range []string{"✓ responder", "noul=0.900", "aciertos 1 (100.0%)", "datos/eval.jsonl", "metrin-eval"} {
		if !strings.Contains(s, quiere) {
			t.Fatalf("el informe no contiene %q:\n%s", quiere, s)
		}
	}
}
