package conocimiento

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"rag-go/internal/v2/tipos"
)

// MotorLocal (tipos.GenerationEngine): un LLM local con API OpenAI-compatible (llama-server u Ollama)
// que SOLO reformula la intro y el texto de los pasos mostrados. Recibe y devuelve el mismo JSON; si
// cambia el número u orden de los pasos, sus ids, sus fotos, un término entre ** **, una cifra, un
// atajo, una ruta de menú o una cita, la reformulación se descarta y el texto sale de las plantillas.
//
// Variables: GENERATION_ENGINE=local, GENERATION_MODEL (vacío = plantillas: el modelo lo elige quien lo
// mide), GENERATION_URL (por defecto http://127.0.0.1:8080/v1, llama-server; Ollama:
// http://127.0.0.1:11434/v1), GENERATION_TIMEOUT_MS (5000).
type MotorLocal struct {
	URL        string
	Modelo     string
	Timeout    time.Duration
	Cliente    *http.Client
	Plantillas *MotorPlantillas // respaldo y renderizador
}

var _ tipos.GenerationEngine = (*MotorLocal)(nil)

// URLGeneracionLocal: llama-server en la Mac (OpenAI-compatible).
const URLGeneracionLocal = "http://127.0.0.1:8080/v1"

// NuevoMotorGeneracion elige el motor por entorno. Sin GENERATION_ENGINE=local o sin GENERATION_MODEL:
// plantillas (por defecto).
func NuevoMotorGeneracion(b *Base, getenv func(string) string) tipos.GenerationEngine {
	if getenv == nil {
		getenv = os.Getenv
	}
	pl := NuevoMotorPlantillas(b)
	if strings.ToLower(strings.TrimSpace(getenv("GENERATION_ENGINE"))) != "local" || strings.TrimSpace(getenv("GENERATION_MODEL")) == "" {
		return pl
	}
	url := strings.TrimSpace(getenv("GENERATION_URL"))
	if url == "" {
		url = URLGeneracionLocal
	}
	ms, err := strconv.Atoi(getenv("GENERATION_TIMEOUT_MS"))
	if err != nil || ms <= 0 {
		ms = 5000
	}
	return &MotorLocal{URL: url, Modelo: strings.TrimSpace(getenv("GENERATION_MODEL")), Timeout: time.Duration(ms) * time.Millisecond, Plantillas: pl}
}

func (m *MotorLocal) Nombre() string { return "local:" + m.Modelo }

// Descarte: la reformulación del modelo no se usó. El texto devuelto junto a él es el de plantillas
// (siempre utilizable); el error solo informa, para la traza.
type Descarte struct{ Motivos []string }

func (d *Descarte) Error() string {
	return "reformulación descartada, se usan plantillas: " + strings.Join(d.Motivos, "; ")
}

// InformeGeneracion: qué pasó en la redacción.
type InformeGeneracion struct {
	Motor       string   `json:"motor"`
	Reformulada bool     `json:"reformulada"`
	Descartada  bool     `json:"descartada"`
	Motivos     []string `json:"motivos,omitempty"`
	Ms          int64    `json:"ms"`
}

// Redactar: texto reformulado si pasa la validación; si no, el de plantillas con un *Descarte.
func (m *MotorLocal) Redactar(ctx context.Context, p tipos.Plan) (string, error) {
	t, inf, err := m.RedactarConInforme(ctx, p)
	if err != nil {
		return t, err
	}
	if inf.Descartada {
		return t, &Descarte{Motivos: inf.Motivos}
	}
	return t, nil
}

// reformulable: lo único que el modelo ve y puede cambiar.
type reformulable struct {
	Intro string      `json:"intro"`
	Pasos []pasoRefor `json:"pasos"`
}

type pasoRefor struct {
	N     int      `json:"n"`
	ID    string   `json:"id"`
	Texto string   `json:"texto"`
	Fotos []string `json:"fotos"`
}

const promptReformular = `Eres el redactor de Metrín, asistente del ERP S10. Recibes un JSON con la introducción y los pasos de un procedimiento.
Devuelve EXACTAMENTE el mismo JSON (mismas claves, mismos pasos, mismo orden, mismos "n", "id" y "fotos") cambiando solo la redacción de "intro" y de cada "texto" para que se lea claro, en usted y en imperativo.
Reglas: no agregues, quites, juntes ni separes pasos ni líneas; copia idénticos los términos entre ** ** y las cifras; no agregues menús, botones, atajos de teclado, rutas, páginas, manuales ni citas. Responde solo con el JSON.`

// RedactarConInforme hace la llamada y la valida. err solo si el plan no se puede redactar ni con plantillas.
func (m *MotorLocal) RedactarConInforme(ctx context.Context, p tipos.Plan) (string, InformeGeneracion, error) {
	inf := InformeGeneracion{Motor: m.Nombre()}
	base, err := m.Plantillas.Redactar(ctx, p)
	if err != nil {
		return "", inf, err
	}
	orig := extraerReformulable(p)
	if p.SinEvidencia || (orig.Intro == "" && len(orig.Pasos) == 0) {
		inf.Motor = m.Plantillas.Nombre()
		return base, inf, nil // nada que reformular: avisos y definiciones van literales
	}
	ini := time.Now()
	nuevo, err := m.llamar(ctx, orig)
	inf.Ms = time.Since(ini).Milliseconds()
	if err != nil {
		inf.Descartada, inf.Motivos = true, []string{"modelo: " + err.Error()}
		return base, inf, nil
	}
	if ms := ValidarReformulacion(orig, nuevo); len(ms) > 0 {
		inf.Descartada, inf.Motivos = true, ms
		return base, inf, nil
	}
	q := aplicarReformulacion(p, nuevo)
	t, err := m.Plantillas.Redactar(ctx, q)
	if err != nil {
		inf.Descartada, inf.Motivos = true, []string{"plantillas: " + err.Error()}
		return base, inf, nil
	}
	inf.Reformulada = true
	return t, inf, nil
}

func extraerReformulable(p tipos.Plan) reformulable {
	r := reformulable{Intro: p.Intro, Pasos: []pasoRefor{}}
	if p.SinEvidencia {
		return r
	}
	switch p.Tipo {
	case tipos.Procedimiento, tipos.Configuracion, tipos.Navegacion:
	default:
		r.Intro = "" // conceptos, errores y aclaraciones: literales
		return r
	}
	ver := map[int]bool{}
	for _, n := range mostrados(p) {
		ver[n] = true
	}
	for _, x := range p.Pasos {
		if !ver[x.N] {
			continue
		}
		pr := pasoRefor{N: x.N, ID: x.ID, Texto: x.Texto, Fotos: []string{}}
		for _, f := range x.Fotos {
			pr.Fotos = append(pr.Fotos, f.ID)
		}
		r.Pasos = append(r.Pasos, pr)
	}
	return r
}

func aplicarReformulacion(p tipos.Plan, r reformulable) tipos.Plan {
	q := p
	q.Intro = r.Intro
	q.Pasos = append([]tipos.PasoPlan(nil), p.Pasos...)
	porID := map[string]string{}
	for _, x := range r.Pasos {
		porID[x.ID] = x.Texto
	}
	for i := range q.Pasos {
		if t, ok := porID[q.Pasos[i].ID]; ok {
			q.Pasos[i].Texto = t
		}
	}
	return q
}

func (m *MotorLocal) llamar(ctx context.Context, r reformulable) (reformulable, error) {
	var vacio reformulable
	if m.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.Timeout)
		defer cancel()
	}
	entrada, _ := json.Marshal(r)
	cuerpo, _ := json.Marshal(map[string]any{
		"model": m.Modelo,
		"messages": []map[string]string{
			{"role": "system", "content": promptReformular},
			{"role": "user", "content": string(entrada)},
		},
		"temperature":     0.2,
		"stream":          false,
		"response_format": map[string]string{"type": "json_object"},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlChat(m.URL), bytes.NewReader(cuerpo))
	if err != nil {
		return vacio, err
	}
	req.Header.Set("Content-Type", "application/json")
	cli := m.Cliente
	if cli == nil {
		cli = http.DefaultClient
	}
	resp, err := cli.Do(req)
	if err != nil {
		return vacio, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return vacio, fmt.Errorf("HTTP %d: %s", resp.StatusCode, recortar(string(b), 200))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &out); err != nil || len(out.Choices) == 0 {
		return vacio, fmt.Errorf("respuesta sin choices: %s", recortar(string(b), 200))
	}
	contenido := strings.TrimSpace(out.Choices[0].Message.Content)
	contenido = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(contenido, "```json"), "```"), "```")
	var r2 reformulable
	if err := json.Unmarshal([]byte(strings.TrimSpace(contenido)), &r2); err != nil {
		return vacio, fmt.Errorf("JSON inválido del modelo: %v", err)
	}
	return r2, nil
}

func urlChat(u string) string {
	u = strings.TrimRight(u, "/")
	switch {
	case strings.HasSuffix(u, "/chat/completions"):
		return u
	case strings.HasSuffix(u, "/v1"):
		return u + "/chat/completions"
	}
	return u + "/v1/chat/completions"
}

func recortar(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

var (
	cifraRe   = regexp.MustCompile(`\d+(?:[.,]\d+)*`)
	atajoRe   = regexp.MustCompile(`(?i)\b(?:ctrl|control|alt|shift|may[uú]s|cmd)\s*\+\s*[a-z0-9]+|\bF(?:1[0-2]|[1-9])\b`)
	citaRe    = regexp.MustCompile(`(?i)\[F?\d+\]|\bp[aá]g(?:ina)?s?\.?\s*\d+|\bp\.\s*\d+|\bfuente\b|\bmanual\b|https?://`)
	rutaSepRe = regexp.MustCompile(`›|→| > `)
)

// ValidarReformulacion compara lo que el modelo devolvió con lo enviado. Lista vacía = se acepta.
func ValidarReformulacion(orig, nuevo reformulable) []string {
	var ms []string
	ms = append(ms, compararTexto("intro", orig.Intro, nuevo.Intro, true)...)
	if len(nuevo.Pasos) != len(orig.Pasos) {
		return append(ms, fmt.Sprintf("cambió el número de pasos: %d → %d", len(orig.Pasos), len(nuevo.Pasos)))
	}
	for i, o := range orig.Pasos {
		n := nuevo.Pasos[i]
		if n.N != o.N || n.ID != o.ID {
			ms = append(ms, fmt.Sprintf("paso %d: cambió el orden o el id (%d %s → %d %s)", i+1, o.N, o.ID, n.N, n.ID))
			continue
		}
		if strings.Join(n.Fotos, ",") != strings.Join(o.Fotos, ",") {
			ms = append(ms, fmt.Sprintf("paso %d: cambiaron las fotos (%v → %v)", o.N, o.Fotos, n.Fotos))
		}
		ms = append(ms, compararTexto("paso "+strconv.Itoa(o.N), o.Texto, n.Texto, false)...)
	}
	return ms
}

// compararTexto: mismas negritas (como multiconjunto, sin tildes ni mayúsculas), mismas líneas, sin
// cifras, atajos, rutas ni citas nuevas, y un largo razonable.
func compararTexto(donde, o, n string, puedeVacio bool) []string {
	var ms []string
	if strings.TrimSpace(n) == "" {
		if puedeVacio && strings.TrimSpace(o) == "" {
			return nil
		}
		return []string{donde + ": texto vacío"}
	}
	if a, b := multiNegritas(o), multiNegritas(n); a != b {
		ms = append(ms, fmt.Sprintf("%s: cambiaron los términos en negrita (%s → %s)", donde, a, b))
	}
	if strings.Count(strings.TrimSpace(o), "\n") != strings.Count(strings.TrimSpace(n), "\n") {
		ms = append(ms, donde+": cambió el número de líneas (subpasos)")
	}
	for _, x := range nuevos(cifraRe, o, n) {
		ms = append(ms, donde+": cifra nueva «"+x+"»")
	}
	for _, x := range nuevos(atajoRe, o, n) {
		ms = append(ms, donde+": atajo nuevo «"+x+"»")
	}
	for _, x := range nuevos(citaRe, o, n) {
		ms = append(ms, donde+": cita o referencia nueva «"+x+"»")
	}
	if len(rutaSepRe.FindAllString(n, -1)) > len(rutaSepRe.FindAllString(o, -1)) {
		ms = append(ms, donde+": ruta de menú nueva")
	}
	if len(n) > 2*len(o)+80 {
		ms = append(ms, donde+": texto demasiado largo")
	}
	return ms
}

func multiNegritas(s string) string {
	var xs []string
	for _, x := range negritas(s) {
		xs = append(xs, plegar(x))
	}
	sort.Strings(xs)
	return strings.Join(xs, "|")
}

// nuevos: coincidencias del patrón en n que no están en o (sin tildes ni mayúsculas).
func nuevos(re *regexp.Regexp, o, n string) []string {
	vis := map[string]bool{}
	for _, x := range re.FindAllString(o, -1) {
		vis[plegar(x)] = true
	}
	var out []string
	for _, x := range re.FindAllString(n, -1) {
		if !vis[plegar(x)] {
			out = append(out, x)
			vis[plegar(x)] = true
		}
	}
	return out
}
