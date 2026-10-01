// Package eval es el harness de evaluación de Metrín: una suite
// fija de preguntas con veredicto esperado que pasa por el flujo
// completo (RAG + JEV), graba las decisiones reales en un JSONL
// propio (origen «metrin-eval», visible en el visor de trazas) y
// compara la decisión de JEV —noul, la probabilidad de que la
// conducta tomada sea correcta— y su confianza con lo esperado.
//
// El harness no vuelve a preguntar a JEV: configura el cliente
// JEV del RAG con su propio JSONL, así cada decisión real que toma
// el flujo queda grabada donde el visor la puede leer
// (TRAZAS_REPO=archivo, o cargada en Postgres con trazas/cargar.py).
// Después de cada caso lee las trazas nuevas del archivo para
// obtener el noul, la confianza y la latencia de la decisión real.
package eval

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"rag-go/internal/rag"
)

// Caso es una pregunta fija con el veredicto esperado del flujo:
// «responder» (debe hallar contexto útil y responder con él) o
// «abstener» (debe declarar que no hay contexto suficiente).
//
// Los casos de abstención usan preguntas largas (>8 palabras) a
// propósito: un mensaje corto sin contexto es charla conversacional
// por diseño del flujo, otra conducta válida que no cuenta como
// abstención.
type Caso struct {
	Pregunta string `json:"pregunta"`
	Esperado string `json:"esperado"` // "responder" | "abstener"
	Nota     string `json:"nota,omitempty"`
}

// Opciones del harness.
type Opciones struct {
	K      int     // trozos a recuperar por caso
	Umbral float64 // noul mínimo para que JEV avale la conducta
}

// TrazaDecision es lo que interesa de una traza JEV leída del
// JSONL del harness: la conducta evaluada (el «modo» del estado)
// y el veredicto del servidor (noul, confianza, latencia).
type TrazaDecision struct {
	Modo          string
	Instrucciones string
	Noul          float64
	Confianza     float64
	MS            int64
}

// Resultado es el veredicto de un caso.
type Resultado struct {
	Caso   Caso           `json:"caso"`
	Flujo  string         `json:"flujo"` // responde | abstiene | error
	FlujoOK bool          `json:"flujo_ok"`
	Traza  *TrazaDecision `json:"traza,omitempty"` // nil: no hubo decisión JEV
	JevOK  bool           `json:"jev_ok"`          // noul ≥ umbral
	SinTraza bool         `json:"sin_traza"`
	Acierto bool          `json:"acierto"`
	Error  string         `json:"error,omitempty"`
}

// Informe agrega los resultados de una corrida.
type Informe struct {
	Total           int         `json:"total"`
	Aciertos        int         `json:"aciertos"`
	FlujoOK         int         `json:"flujo_ok"`
	JevOK           int         `json:"jev_ok"`
	SinTraza        int         `json:"sin_traza"`
	ConfianzaMedia  float64     `json:"confianza_media"`
	LatenciaMediaMS int64       `json:"latencia_media_ms"`
	LatenciaP95MS   int64       `json:"latencia_p95_ms"`
	Resultados      []Resultado `json:"resultados"`
}

// Ejecutar pasa cada caso por Preguntar (RAG + JEV) y compara el
// resultado con el veredicto esperado. rutaTrazas es el JSONL que
// el cliente JEV del RAG ya tiene configurado (JEV_TRAZAS): se
// lee por posición, así cada caso recoge solo sus propias trazas.
func Ejecutar(ctx context.Context, r *rag.RAG, rutaTrazas string, casos []Caso, o Opciones) Informe {
	inf := Informe{Resultados: make([]Resultado, 0, len(casos))}
	for _, c := range casos {
		antes := tamañoArchivo(rutaTrazas)
		res, err := r.Preguntar(ctx, c.Pregunta, rag.Opciones{K: o.K})
		trazas, _ := leerTrazasNuevas(rutaTrazas, antes) // un fallo de lectura = sin traza, no = fallo del caso
		resul := veredicto(c, &res, err, ultimaDecision(trazas), o)
		inf.Resultados = append(inf.Resultados, resul)
	}
	return resumen(inf)
}

// veredicto compara lo que hizo el flujo y lo que dijo JEV con lo
// que esperaba el caso. El acierto exige ambas cosas: el flujo hizo
// la conducta esperada y JEV la avaló (noul ≥ umbral).
func veredicto(c Caso, res *rag.Respuesta, err error, t *TrazaDecision, o Opciones) Resultado {
	r := Resultado{Caso: c}
	if err != nil {
		r.Flujo = "error"
		r.Error = err.Error()
		return r
	}
	if res.SinContexto {
		r.Flujo = "abstiene"
	} else {
		r.Flujo = "responde"
	}
	r.FlujoOK = (c.Esperado == "abstener") == (r.Flujo == "abstiene")
	if t == nil {
		r.SinTraza = true
		return r
	}
	r.Traza = t
	r.JevOK = t.Noul >= o.Umbral
	r.Acierto = r.FlujoOK && r.JevOK
	return r
}

// resumen rellena los agregados del informe. Confianza y latencia
// se promedian sobre los casos con traza: sin decisión JEV no hay
// nada que medir.
func resumen(inf Informe) Informe {
	var confianza float64
	var latencias []float64
	for _, r := range inf.Resultados {
		if r.FlujoOK {
			inf.FlujoOK++
		}
		if r.JevOK {
			inf.JevOK++
		}
		if r.SinTraza {
			inf.SinTraza++
		}
		if r.Acierto {
			inf.Aciertos++
		}
		if r.Traza != nil {
			confianza += r.Traza.Confianza
			latencias = append(latencias, float64(r.Traza.MS))
		}
	}
	inf.Total = len(inf.Resultados)
	if n := len(latencias); n > 0 {
		inf.ConfianzaMedia = confianza / float64(n)
		sort.Float64s(latencias)
		inf.LatenciaMediaMS = int64(math.Round(media(latencias)))
		inf.LatenciaP95MS = int64(math.Round(latencias[min(int(math.Ceil(0.95*float64(n))), n)-1]))
	}
	return inf
}

func media(xs []float64) float64 {
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

// trazaVisor es una línea del JSONL de trazas en el formato del
// visor (claves en español; confianza puede ser null).
type trazaVisor struct {
	ID        string          `json:"id"`
	TS        string          `json:"ts"`
	Modelo    string          `json:"modelo"`
	Origen    string          `json:"origen"`
	MS        int64           `json:"ms"`
	Estado    map[string]any  `json:"estado"`
	Preguntas []preguntaVisor `json:"preguntas"`
}

type preguntaVisor struct {
	ID            string          `json:"id"`
	Tipo          string          `json:"tipo"`
	Instrucciones any             `json:"instrucciones"`
	Respuesta     *respuestaVisor `json:"respuesta"`
}

type respuestaVisor struct {
	Tipo      string   `json:"tipo"`
	Noul      float64  `json:"noul"`
	Confianza *float64 `json:"confianza"`
}

// leerTrazasNuevas devuelve las trazas escritas en ruta a partir
// del byte desde. Si el archivo no existe (aún no se escribió
// nada) o se truncó, devuelve lo que haya. Una línea a medio
// escribir se ignora, no falla la corrida.
func leerTrazasNuevas(ruta string, desde int64) ([]trazaVisor, error) {
	st, err := os.Stat(ruta)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	desde = min(desde, st.Size())
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := f.Seek(desde, io.SeekStart); err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024) // trazas con estado grande
	var out []trazaVisor
	for sc.Scan() {
		linea := strings.TrimSpace(sc.Text())
		if linea == "" {
			continue
		}
		var t trazaVisor
		if err := json.Unmarshal([]byte(linea), &t); err != nil {
			continue
		}
		out = append(out, t)
	}
	return out, sc.Err()
}

// ultimaDecision extrae la decisión de la última traza con
// respuesta. El flujo emite a lo sumo una traza por caso, pero se
// recorre hacia atrás por si el archivo tiene una línea vieja.
func ultimaDecision(trazas []trazaVisor) *TrazaDecision {
	for i := len(trazas) - 1; i >= 0; i-- {
		p := trazas[i]
		if len(p.Preguntas) == 0 || p.Preguntas[0].Respuesta == nil {
			continue
		}
		d := &TrazaDecision{
			MS:            p.MS,
			Instrucciones: aTexto(p.Preguntas[0].Instrucciones),
			Noul:          p.Preguntas[0].Respuesta.Noul,
		}
		if modo, ok := p.Estado["modo"].(string); ok {
			d.Modo = modo
		}
		if c := p.Preguntas[0].Respuesta.Confianza; c != nil {
			d.Confianza = *c
		}
		return d
	}
	return nil
}

// aTexto vuelve legible un campo «any» del JSON: las instrucciones
// del flujo son texto, pero no se garantiza por el formato.
func aTexto(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func tamañoArchivo(ruta string) int64 {
	st, err := os.Stat(ruta)
	if err != nil {
		return 0
	}
	return st.Size()
}

// CargarCasos lee una suite de casos desde un JSONL
// ({"pregunta": …, "esperado": "responder"|"abstener", "nota": …}).
func CargarCasos(ruta string) ([]Caso, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var casos []Caso
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		linea := strings.TrimSpace(sc.Text())
		if linea == "" || strings.HasPrefix(linea, "#") {
			continue
		}
		var c Caso
		if err := json.Unmarshal([]byte(linea), &c); err != nil {
			return nil, fmt.Errorf("%s: %w", ruta, err)
		}
		if c.Pregunta == "" {
			return nil, fmt.Errorf("%s: caso sin pregunta", ruta)
		}
		if c.Esperado != "responder" && c.Esperado != "abstener" {
			return nil, fmt.Errorf("%s: veredicto %q (usa «responder» o «abstener»)", ruta, c.Esperado)
		}
		casos = append(casos, c)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(casos) == 0 {
		return nil, fmt.Errorf("%s: la suite está vacía", ruta)
	}
	return casos, nil
}

// Formatear renderiza el informe para consola: una línea por caso
// con su veredicto y al final los agregados.
func Formatear(inf Informe, rutaTrazas, origen string) string {
	var b strings.Builder
	for i, r := range inf.Resultados {
		marca := "✗"
		if r.Acierto {
			marca = "✓"
		}
		fmt.Fprintf(&b, "[%2d/%d] %s %-8s %s\n", i+1, inf.Total, marca, r.Caso.Esperado, recortar(r.Caso.Pregunta, 72))
		if r.Error != "" {
			fmt.Fprintf(&b, "        flujo: error — %s\n", recortar(r.Error, 70))
			continue
		}
		flujo := "✓"
		if !r.FlujoOK {
			flujo = fmt.Sprintf("✗ (esperado %s)", r.Caso.Esperado)
		}
		fmt.Fprintf(&b, "        flujo: %s %s", r.Flujo, flujo)
		if r.SinTraza {
			fmt.Fprintln(&b, " · sin traza JEV (¿servidor caído o camino sin emisión?)")
			continue
		}
		jev := "✓"
		if !r.JevOK {
			jev = "✗"
		}
		fmt.Fprintf(&b, " · JEV: noul=%.3f %s (confianza %.2f, %d ms)\n",
			r.Traza.Noul, jev, r.Traza.Confianza, r.Traza.MS)
		if r.Traza.Instrucciones != "" {
			fmt.Fprintf(&b, "        decisión: «%s»\n", recortar(r.Traza.Instrucciones, 90))
		}
	}
	pct := 0.0
	if inf.Total > 0 {
		pct = 100 * float64(inf.Aciertos) / float64(inf.Total)
	}
	fmt.Fprintf(&b, "── resumen ─────────────────────────────────────────\n")
	fmt.Fprintf(&b, "casos %d · aciertos %d (%.1f%%) · flujo correcto %d/%d · JEV avala %d/%d · sin traza %d\n",
		inf.Total, inf.Aciertos, pct, inf.FlujoOK, inf.Total, inf.JevOK, inf.Total, inf.SinTraza)
	fmt.Fprintf(&b, "confianza media %.3f · latencia media %d ms · p95 %d ms\n",
		inf.ConfianzaMedia, inf.LatenciaMediaMS, inf.LatenciaP95MS)
	fmt.Fprintf(&b, "trazas: %s (origen %s)\n", rutaTrazas, origen)
	return b.String()
}

func recortar(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}
