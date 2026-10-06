// Package traza registra el camino de una pregunta por las etapas de Metrín
// (modo traza). El contrato que manda es docs/CONTRATO-traza-metrin.md (v1):
// 18 etapas fijas, en orden, con su estado, tiempo y datos técnicos, más las
// oportunidades de mejora que salen de reglas fijas sobre el turno.
//
// El Recorder viaja en el context.Context. Con el modo apagado no hay
// Recorder: De devuelve nil y todos sus métodos son no-op sobre un puntero
// nil, así que la instrumentación no cambia ni la respuesta ni el trabajo.
package traza

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Version del contrato de la traza.
const Version = 1

// MaxCaracteres: tope de los recortes de texto (pregunta, razones, citas).
const MaxCaracteres = 120

// MaxBytes: si la traza serializada pasa de aquí se recortan los fragmentos.
const MaxBytes = 16 << 10

// FragmentosTruncada: fragmentos que se conservan al truncar.
const FragmentosTruncada = 5

// Ids de las 18 etapas, en el orden del contrato.
const (
	EtapaInicio         = "inicio"
	EtapaEntrada        = "entrada"
	EtapaReglaAciertos  = "regla_aciertos"
	EtapaEmbebedor      = "embebedor"
	EtapaClasificador   = "clasificador"
	EtapaSeguridad      = "seguridad"
	EtapaRuta           = "ruta"
	EtapaReescritura    = "reescritura"
	EtapaBusqueda       = "busqueda"
	EtapaSeleccion      = "seleccion"
	EtapaEvidencia      = "evidencia"
	EtapaGeneracion     = "generacion"
	EtapaVerificacion   = "verificacion"
	EtapaFotos          = "fotos"
	EtapaCharla         = "charla"
	EtapaJEV            = "jev"
	EtapaRegistroFallos = "registro_fallos"
	EtapaFin            = "fin"
)

// Estados posibles de una etapa.
const (
	EstadoOK       = "ok"
	EstadoAlerta   = "alerta"
	EstadoError    = "error"
	EstadoRespaldo = "respaldo"
	EstadoOmitida  = "omitida"
	EstadoNoTomada = "no_tomada"
)

// Tipos de etapa (deciden la forma en la interfaz).
const (
	TipoEvento   = "evento"
	TipoCodigo   = "codigo"
	TipoModelo   = "modelo"
	TipoDecision = "decision"
)

// Datos técnicos de una etapa: identificadores, puntajes y recortes.
type Datos = map[string]any

type definicion struct {
	id, nombre, tipo string
	minimos          []string
}

// catalogo: las 18 etapas del contrato, con sus datos mínimos.
var catalogo = [...]definicion{
	{EtapaInicio, "Pregunta", TipoEvento, nil},
	{EtapaEntrada, "Entrada", TipoCodigo, []string{"pregunta", "hilo_turnos"}},
	{EtapaReglaAciertos, "Regla «aciertos»", TipoDecision, []string{"aplica"}},
	{EtapaEmbebedor, "Embebedor", TipoModelo, []string{"proveedor", "modelo"}},
	{EtapaClasificador, "Clasificador kNN", TipoModelo, []string{"clasificador", "intencion", "similitud", "umbral", "paso_umbral"}},
	{EtapaSeguridad, "¿Social y corta?", TipoDecision, []string{"pregunta_larga", "desvia_a_charla"}},
	{EtapaRuta, "Ruta", TipoDecision, []string{"ruta", "tipo_consulta"}},
	{EtapaReescritura, "Reescritura del hilo", TipoModelo, []string{"original", "reescrita"}},
	{EtapaBusqueda, "Búsqueda", TipoCodigo, []string{"metodo", "k", "filtro", "resultados"}},
	{EtapaSeleccion, "Selección", TipoCodigo, []string{"ventana", "max_contextos", "corte", "seleccionados", "fragmentos"}},
	{EtapaEvidencia, "¿Hay contexto?", TipoDecision, []string{"sin_contexto", "distancia_min", "motivo"}},
	{EtapaGeneracion, "Generación LLM", TipoModelo, []string{"proveedor", "modelo", "modo"}},
	{EtapaVerificacion, "Verificación de citas", TipoCodigo, []string{"citas"}},
	{EtapaFotos, "Fotos de pasos", TipoCodigo, []string{"fotos"}},
	{EtapaCharla, "Charla LLM", TipoModelo, []string{"modelo"}},
	{EtapaJEV, "Juicio JEV", TipoModelo, []string{"activo"}},
	{EtapaRegistroFallos, "Registro de fallos", TipoCodigo, []string{"registrado"}},
	{EtapaFin, "Respuesta", TipoEvento, []string{"modo"}},
}

const numEtapas = len(catalogo)

var indice = func() map[string]int {
	m := make(map[string]int, numEtapas)
	for i, d := range catalogo {
		m[d.id] = i
	}
	return m
}()

// IDs devuelve los ids de las etapas en el orden del contrato.
func IDs() []string {
	out := make([]string, numEtapas)
	for i, d := range catalogo {
		out[i] = d.id
	}
	return out
}

// Minimos devuelve las claves de datos que el contrato exige a una etapa.
func Minimos(id string) []string {
	if i, ok := indice[id]; ok {
		return append([]string(nil), catalogo[i].minimos...)
	}
	return nil
}

// EstadoValido dice si s es uno de los estados del contrato.
func EstadoValido(s string) bool {
	switch s {
	case EstadoOK, EstadoAlerta, EstadoError, EstadoRespaldo, EstadoOmitida, EstadoNoTomada:
		return true
	}
	return false
}

// Habilitada interpreta METRIN_TRAZA: solo «1» o «true» la encienden.
func Habilitada(valor string) bool {
	v := strings.TrimSpace(valor)
	return v == "1" || strings.EqualFold(v, "true")
}

// Etapa es una tarjeta del diagrama.
type Etapa struct {
	ID     string   `json:"id"`
	Nombre string   `json:"nombre"`
	Tipo   string   `json:"tipo"`
	Estado string   `json:"estado"`
	Ms     *float64 `json:"ms"`
	Modelo string   `json:"modelo,omitempty"`
	Razon  string   `json:"razon,omitempty"`
	Datos  Datos    `json:"datos"`
}

// Oportunidad de mejora calculada con una regla fija.
type Oportunidad struct {
	N     int    `json:"n"`
	Etapa string `json:"etapa"`
	Nivel string `json:"nivel"`
	Texto string `json:"texto"`
}

// Traza es lo que viaja en la respuesta de /ask bajo la clave «traza».
type Traza struct {
	Version       int           `json:"version"`
	TotalMs       float64       `json:"total_ms"`
	Pregunta      string        `json:"pregunta"`
	Truncada      bool          `json:"truncada"`
	Etapas        []Etapa       `json:"etapas"`
	Oportunidades []Oportunidad `json:"oportunidades"`
	// Solo en turnos V2 (v2.go, contrato «Etapas V2»): en V1 no aparecen.
	Agente   string  `json:"agente,omitempty"`
	EtapasV2 []Etapa `json:"etapas_v2,omitempty"`
}

// Fragmento seleccionado (o descartado) para el contexto del LLM. Nunca
// lleva el texto del fragmento: solo de dónde viene y su distancia.
type Fragmento struct {
	Documento    string  `json:"documento"`
	Pagina       *int    `json:"pagina"`
	Seccion      string  `json:"seccion,omitempty"`
	Cita         string  `json:"cita,omitempty"`
	Confianza    string  `json:"confianza,omitempty"`
	Distancia    float64 `json:"distancia"`
	Penalizacion float64 `json:"penalizacion"`
	// DentroDelCorte (solo en descartados): true = cabía por distancia y quedó
	// fuera por cita repetida o por el tope de contextos.
	DentroDelCorte *bool `json:"dentro_del_corte,omitempty"`
}

const (
	pendiente = iota
	iniciada
	cerrada
)

// Recorder anota las etapas de un turno. Un puntero nil es un recorder
// apagado: todos sus métodos regresan sin hacer nada.
type Recorder struct {
	mu       sync.Mutex
	t0       time.Time
	pregunta string
	etapas   [numEtapas]Etapa
	fase     [numEtapas]uint8
	inicios  [numEtapas]time.Time
	medidos  [numEtapas]bool
	v2       *RecorderV2 // etapas V2 (v2.go); nil en un turno V1
}

// Nuevo crea un recorder para una pregunta y marca la etapa «inicio».
func Nuevo(pregunta string) *Recorder {
	r := &Recorder{t0: time.Now(), pregunta: Recortar(pregunta)}
	for i, d := range catalogo {
		r.etapas[i] = Etapa{ID: d.id, Nombre: d.nombre, Tipo: d.tipo, Datos: Datos{}}
	}
	r.etapas[0].Estado = EstadoOK
	r.fase[0] = cerrada
	return r
}

type claveContexto struct{}

// ConContexto guarda el recorder en el contexto. Con r nil devuelve ctx tal cual.
func ConContexto(ctx context.Context, r *Recorder) context.Context {
	if r == nil {
		return ctx
	}
	return context.WithValue(ctx, claveContexto{}, r)
}

// De devuelve el recorder del contexto, o nil si el modo traza está apagado.
func De(ctx context.Context) *Recorder {
	if ctx == nil {
		return nil
	}
	r, _ := ctx.Value(claveContexto{}).(*Recorder)
	return r
}

func (r *Recorder) etapa(id string) (int, bool) {
	i, ok := indice[id]
	return i, ok
}

// Inicio arranca el cronómetro de una etapa.
func (r *Recorder) Inicio(id string) {
	if r == nil {
		return
	}
	i, ok := r.etapa(id)
	if !ok {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inicios[i] = time.Now()
	if r.fase[i] == pendiente {
		r.fase[i] = iniciada
	}
}

// Fin cierra una etapa con su estado y datos. El tiempo es el transcurrido
// desde Inicio; sin Inicio, ms queda en null. Los datos se suman a los que
// ya hubiera (Dato).
func (r *Recorder) Fin(id, estado string, datos Datos) {
	if r == nil {
		return
	}
	r.cerrar(id, estado, -1, datos)
}

// FinDuracion cierra una etapa con un tiempo medido aparte.
func (r *Recorder) FinDuracion(id, estado string, d time.Duration, datos Datos) {
	if r == nil {
		return
	}
	r.cerrar(id, estado, d, datos)
}

func (r *Recorder) cerrar(id, estado string, d time.Duration, datos Datos) {
	i, ok := r.etapa(id)
	if !ok {
		return
	}
	if !EstadoValido(estado) {
		estado = EstadoError
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	e := &r.etapas[i]
	e.Estado = estado
	switch {
	case d >= 0:
		e.Ms = msDe(d)
		r.medidos[i] = true
	case r.fase[i] == iniciada && !r.inicios[i].IsZero():
		e.Ms = msDe(time.Since(r.inicios[i]))
		r.medidos[i] = true
	}
	for k, v := range datos {
		e.Datos[k] = v
	}
	r.fase[i] = cerrada
}

// Omitir marca una etapa que no aplicaba en la ruta tomada.
func (r *Recorder) Omitir(id, razon string) {
	r.sinEjecutar(id, EstadoOmitida, razon, nil)
}

// NoTomada marca una etapa de una rama que el mensaje no siguió.
func (r *Recorder) NoTomada(id, razon string) {
	r.sinEjecutar(id, EstadoNoTomada, razon, nil)
}

// OmitirCon es Omitir con datos (p. ej. activo=false).
func (r *Recorder) OmitirCon(id, estado, razon string, datos Datos) {
	r.sinEjecutar(id, estado, razon, datos)
}

func (r *Recorder) sinEjecutar(id, estado, razon string, datos Datos) {
	if r == nil {
		return
	}
	i, ok := r.etapa(id)
	if !ok {
		return
	}
	if estado != EstadoOmitida && estado != EstadoNoTomada {
		estado = EstadoOmitida
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	e := &r.etapas[i]
	e.Estado = estado
	e.Ms = nil
	e.Razon = razon
	for k, v := range datos {
		e.Datos[k] = v
	}
	r.fase[i] = cerrada
}

// Dato fija un dato de una etapa sin cerrarla.
func (r *Recorder) Dato(id, clave string, valor any) {
	if r == nil {
		return
	}
	i, ok := r.etapa(id)
	if !ok {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.etapas[i].Datos[clave] = valor
}

// DatoDe lee un dato ya anotado.
func (r *Recorder) DatoDe(id, clave string) (any, bool) {
	if r == nil {
		return nil, false
	}
	i, ok := r.etapa(id)
	if !ok {
		return nil, false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.etapas[i].Datos[clave]
	return v, ok
}

// Razon fija la razón (se recorta a 120 caracteres al cerrar).
func (r *Recorder) Razon(id, razon string) {
	if r == nil {
		return
	}
	if i, ok := r.etapa(id); ok {
		r.mu.Lock()
		r.etapas[i].Razon = razon
		r.mu.Unlock()
	}
}

// Modelo fija el modelo o la regla que decidió la etapa.
func (r *Recorder) Modelo(id, modelo string) {
	if r == nil {
		return
	}
	if i, ok := r.etapa(id); ok {
		r.mu.Lock()
		r.etapas[i].Modelo = modelo
		r.mu.Unlock()
	}
}

// CambiarEstado cambia el estado de una etapa ya cerrada (y su razón, si se da).
func (r *Recorder) CambiarEstado(id, estado, razon string) {
	if r == nil || !EstadoValido(estado) {
		return
	}
	if i, ok := r.etapa(id); ok {
		r.mu.Lock()
		r.etapas[i].Estado = estado
		if razon != "" {
			r.etapas[i].Razon = razon
		}
		r.mu.Unlock()
	}
}

// Estado devuelve el estado anotado de una etapa ("" si sigue pendiente).
func (r *Recorder) Estado(id string) string {
	if r == nil {
		return ""
	}
	i, ok := r.etapa(id)
	if !ok {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.fase[i] != cerrada {
		return ""
	}
	return r.etapas[i].Estado
}

// Pendiente dice si una etapa aún no se cerró (ni Fin ni Omitir).
func (r *Recorder) Pendiente(id string) bool {
	if r == nil {
		return false
	}
	i, ok := r.etapa(id)
	if !ok {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.fase[i] != cerrada
}

// FallarIniciadas cierra con error las etapas que arrancaron y no terminaron
// (el flujo se cortó por un error a mitad de etapa).
func (r *Recorder) FallarIniciadas(razon string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.etapas {
		if r.fase[i] != iniciada {
			continue
		}
		e := &r.etapas[i]
		e.Estado = EstadoError
		e.Ms = msDe(time.Since(r.inicios[i]))
		e.Razon = razon
		r.fase[i] = cerrada
	}
}

// Cerrar arma la traza: rellena lo pendiente, garantiza los datos mínimos,
// aplica las reglas de oportunidades, recorta textos y respeta los 16 KB.
func (r *Recorder) Cerrar() *Traza {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	t := &Traza{
		Version:  Version,
		TotalMs:  *msDe(time.Since(r.t0)),
		Pregunta: r.pregunta,
		Etapas:   make([]Etapa, numEtapas),
	}
	for i := range r.etapas {
		e := r.etapas[i]
		datos := make(Datos, len(e.Datos)+len(catalogo[i].minimos))
		for k, v := range e.Datos {
			datos[k] = sanear(v)
		}
		e.Datos = datos
		switch r.fase[i] {
		case pendiente:
			e.Estado = EstadoOmitida
			e.Ms = nil
			if e.Razon == "" {
				e.Razon = "no se ejecutó en este turno"
			}
		case iniciada:
			e.Estado = EstadoError
			e.Ms = msDe(time.Since(r.inicios[i]))
			if e.Razon == "" {
				e.Razon = "la etapa empezó y no terminó"
			}
		}
		for _, k := range catalogo[i].minimos {
			if _, ok := e.Datos[k]; !ok {
				e.Datos[k] = nil
			}
		}
		e.Razon = Recortar(e.Razon)
		e.Modelo = Recortar(e.Modelo)
		t.Etapas[i] = e
	}
	r.mu.Unlock()

	t.Oportunidades = aplicarReglas(t.Etapas)
	cerrarV2(r, t)
	limitarTamano(t)
	return t
}

// limitarTamano aplica el tope de 16 KB: primero los fragmentos de la
// selección a 5 (lo que pide el contrato); si aun así no cabe, quita los
// descartados y, en último caso, deja solo los datos mínimos.
func limitarTamano(t *Traza) {
	if tamano(t) <= MaxBytes {
		return
	}
	t.Truncada = true
	sel := &t.Etapas[indice[EtapaSeleccion]]
	if fr, ok := sel.Datos["fragmentos"].([]Fragmento); ok && len(fr) > FragmentosTruncada {
		sel.Datos["fragmentos"] = fr[:FragmentosTruncada]
	}
	if tamano(t) <= MaxBytes {
		return
	}
	delete(sel.Datos, "descartados")
	if tamano(t) <= MaxBytes {
		return
	}
	for i := range t.Etapas {
		minimos := Datos{}
		for _, k := range catalogo[i].minimos {
			minimos[k] = t.Etapas[i].Datos[k]
		}
		t.Etapas[i].Datos = minimos
	}
}

func tamano(t *Traza) int {
	b, err := json.Marshal(t)
	if err != nil {
		return 0
	}
	return len(b)
}

// msDe pasa una duración a milisegundos con un decimal.
func msDe(d time.Duration) *float64 {
	ms := math.Round(float64(d)/float64(time.Millisecond)*10) / 10
	return &ms
}

// Recortar deja un texto en una línea y como mucho 120 caracteres (runas);
// si corta, el último es «…».
func Recortar(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) <= MaxCaracteres {
		return s
	}
	return string(r[:MaxCaracteres-1]) + "…"
}

// Redondear deja un puntaje con tres decimales (como distancia_min en /ask).
func Redondear(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}
	return math.Round(x*1000) / 1000
}

// sanear recorta los textos que van en datos: ningún dato lleva más de 120
// caracteres, venga como texto suelto, lista o mapa.
func sanear(v any) any {
	switch x := v.(type) {
	case string:
		return Recortar(x)
	case []string:
		out := make([]string, len(x))
		for i, s := range x {
			out[i] = Recortar(s)
		}
		return out
	case map[string]string:
		out := make(map[string]any, len(x))
		for k, s := range x {
			out[k] = Recortar(s)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, s := range x {
			out[k] = sanear(s)
		}
		return out
	case []Fragmento:
		out := make([]Fragmento, len(x))
		for i, f := range x {
			f.Documento = Recortar(f.Documento)
			f.Seccion = Recortar(f.Seccion)
			f.Cita = Recortar(f.Cita)
			f.Confianza = Recortar(f.Confianza)
			out[i] = f
		}
		return out
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return nil // JSON no admite NaN ni infinito
		}
		return x
	}
	return v
}

// ResumirError deja un error en una línea útil dentro de 120 caracteres: la
// cabecera (p. ej. «mlx chat (http://…:8080)») más la causa más interna
// («connection refused», «context deadline exceeded»), que en el texto
// completo suele quedar al final y se perdería al recortar.
func ResumirError(err error) string {
	if err == nil {
		return ""
	}
	todo := err.Error()
	raiz := err
	for {
		sig := errors.Unwrap(raiz)
		if sig == nil {
			break
		}
		raiz = sig
	}
	causa := raiz.Error()
	cabeza, _, _ := strings.Cut(todo, ": ")
	if causa == todo || cabeza == todo || strings.HasSuffix(cabeza, causa) {
		return Recortar(todo)
	}
	return Recortar(cabeza + ": " + causa)
}

// Decimal escribe un puntaje con coma decimal (texto para personas).
func Decimal(x float64, dec int) string {
	return strings.Replace(strconv.FormatFloat(x, 'f', dec, 64), ".", ",", 1)
}
