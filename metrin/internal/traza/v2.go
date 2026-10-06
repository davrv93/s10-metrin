package traza

// Etapas V2 (docs/CONTRATO-traza-metrin.md, «Etapas V2»). Son un conjunto APARTE: las 18 etapas de V1 no cambian
// ni de id ni de orden, y en un turno V1 la traza no lleva nada de aquí (ni «agente» ni «etapas_v2»).
//
// Solo el orquestador V2 (internal/v2) pide el sub-registro con Recorder.V2(); al cerrar la traza, si existe, se
// añaden «agente»: "v2" y «etapas_v2» con las 13 etapas siempre en el orden de catalogoV2, más sus oportunidades
// (numeradas a continuación de las de V1).

import (
	"fmt"
	"sync"
	"time"
)

// Ids de las 13 etapas V2, en el orden del contrato.
const (
	EtapaV2TipoRespuesta     = "tipo_respuesta"
	EtapaV2Decision          = "decision"
	EtapaV2PlanConsulta      = "plan_consulta"
	EtapaV2BusquedaLexica    = "busqueda_lexica"
	EtapaV2BusquedaVectorial = "busqueda_vectorial"
	EtapaV2Fusion            = "fusion"
	EtapaV2Rerank            = "rerank"
	EtapaV2Procedimiento     = "procedimiento"
	EtapaV2Pasos             = "pasos"
	EtapaV2FotosPaso         = "fotos_paso"
	EtapaV2Plan              = "plan"
	EtapaV2QualityGate       = "quality_gate"
	EtapaV2Renderizado       = "renderizado"
)

// AgenteV2 es el valor de Traza.Agente en un turno V2.
const AgenteV2 = "v2"

// minimosComunesV2: toda etapa V2 lleva su confianza y su número de fuentes (null si no aplica).
var minimosComunesV2 = []string{"confianza", "fuentes"}

var catalogoV2 = [...]definicion{
	{EtapaV2TipoRespuesta, "Tipo de respuesta", TipoModelo, []string{"tipo", "metodo", "umbral"}},
	{EtapaV2Decision, "Motor de decisión", TipoDecision, []string{"motor", "llamadas", "decisiones", "respaldo"}},
	{EtapaV2PlanConsulta, "Plan de consulta", TipoCodigo, []string{"clases", "consulta", "iteraciones"}},
	{EtapaV2BusquedaLexica, "Búsqueda léxica", TipoCodigo, []string{"resultados"}},
	{EtapaV2BusquedaVectorial, "Búsqueda vectorial", TipoCodigo, []string{"resultados"}},
	{EtapaV2Fusion, "Fusión", TipoCodigo, []string{"resultados", "mejor"}},
	{EtapaV2Rerank, "Reranker", TipoModelo, []string{"activo", "resultados"}},
	{EtapaV2Procedimiento, "¿Qué procedimiento?", TipoDecision, []string{"id"}},
	{EtapaV2Pasos, "Pasos", TipoCodigo, []string{"pasos", "mostrados"}},
	{EtapaV2FotosPaso, "Fotos por paso", TipoCodigo, []string{"fotos", "pasos_con_foto"}},
	{EtapaV2Plan, "Plan de respuesta", TipoCodigo, []string{"tipo", "plantilla", "sin_evidencia"}},
	{EtapaV2QualityGate, "Quality gate", TipoCodigo, []string{"passed", "score", "issues", "intentos"}},
	{EtapaV2Renderizado, "Redacción", TipoModelo, []string{"generador", "modo"}},
}

const numEtapasV2 = len(catalogoV2)

var indiceV2 = func() map[string]int {
	m := make(map[string]int, numEtapasV2)
	for i, d := range catalogoV2 {
		m[d.id] = i
	}
	return m
}()

// IDsV2 devuelve los ids de las etapas V2 en el orden del contrato.
func IDsV2() []string {
	out := make([]string, numEtapasV2)
	for i, d := range catalogoV2 {
		out[i] = d.id
	}
	return out
}

// MinimosV2 devuelve las claves de datos que el contrato exige a una etapa V2.
func MinimosV2(id string) []string {
	i, ok := indiceV2[id]
	if !ok {
		return nil
	}
	return append(append([]string(nil), minimosComunesV2...), catalogoV2[i].minimos...)
}

// RecorderV2 anota las etapas V2 de un turno. Un puntero nil es un registro apagado: sus métodos no hacen nada.
type RecorderV2 struct {
	mu      sync.Mutex
	etapas  [numEtapasV2]Etapa
	fase    [numEtapasV2]uint8
	inicios [numEtapasV2]time.Time
}

// V2 devuelve el sub-registro de etapas V2 (lo crea la primera vez). Con r nil devuelve nil.
func (r *Recorder) V2() *RecorderV2 {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.v2 == nil {
		v := &RecorderV2{}
		for i, d := range catalogoV2 {
			v.etapas[i] = Etapa{ID: d.id, Nombre: d.nombre, Tipo: d.tipo, Datos: Datos{}}
		}
		r.v2 = v
	}
	return r.v2
}

func (v *RecorderV2) indice(id string) (int, bool) {
	if v == nil {
		return 0, false
	}
	i, ok := indiceV2[id]
	return i, ok
}

// Inicio arranca el cronómetro de una etapa V2.
func (v *RecorderV2) Inicio(id string) {
	i, ok := v.indice(id)
	if !ok {
		return
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	v.inicios[i] = time.Now()
	if v.fase[i] == pendiente {
		v.fase[i] = iniciada
	}
}

// Fin cierra una etapa V2 con su estado y datos (se suman a los ya anotados).
func (v *RecorderV2) Fin(id, estado string, datos Datos) { v.cerrar(id, estado, -1, datos) }

// FinDuracion cierra una etapa V2 con un tiempo medido aparte.
func (v *RecorderV2) FinDuracion(id, estado string, d time.Duration, datos Datos) {
	v.cerrar(id, estado, d, datos)
}

func (v *RecorderV2) cerrar(id, estado string, d time.Duration, datos Datos) {
	i, ok := v.indice(id)
	if !ok {
		return
	}
	if !EstadoValido(estado) {
		estado = EstadoError
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	e := &v.etapas[i]
	e.Estado = estado
	switch {
	case d >= 0:
		e.Ms = msDe(d)
	case v.fase[i] == iniciada && !v.inicios[i].IsZero():
		e.Ms = msDe(time.Since(v.inicios[i]))
	}
	for k, x := range datos {
		e.Datos[k] = x
	}
	v.fase[i] = cerrada
}

// Omitir marca una etapa V2 que no aplicaba en la ruta tomada.
func (v *RecorderV2) Omitir(id, razon string) { v.OmitirCon(id, EstadoOmitida, razon, nil) }

// NoTomada marca una etapa V2 de una rama que el turno no siguió.
func (v *RecorderV2) NoTomada(id, razon string) { v.OmitirCon(id, EstadoNoTomada, razon, nil) }

// OmitirCon es Omitir/NoTomada con datos.
func (v *RecorderV2) OmitirCon(id, estado, razon string, datos Datos) {
	i, ok := v.indice(id)
	if !ok {
		return
	}
	if estado != EstadoOmitida && estado != EstadoNoTomada {
		estado = EstadoOmitida
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	e := &v.etapas[i]
	e.Estado = estado
	e.Ms = nil
	e.Razon = razon
	for k, x := range datos {
		e.Datos[k] = x
	}
	v.fase[i] = cerrada
}

// Dato fija un dato de una etapa V2 sin cerrarla.
func (v *RecorderV2) Dato(id, clave string, valor any) {
	i, ok := v.indice(id)
	if !ok {
		return
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	v.etapas[i].Datos[clave] = valor
}

// DatoDe lee un dato ya anotado de una etapa V2.
func (v *RecorderV2) DatoDe(id, clave string) (any, bool) {
	i, ok := v.indice(id)
	if !ok {
		return nil, false
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	x, ok := v.etapas[i].Datos[clave]
	return x, ok
}

// Razon fija la razón de una etapa V2 (se recorta a 120 caracteres al cerrar la traza).
func (v *RecorderV2) Razon(id, razon string) {
	if i, ok := v.indice(id); ok {
		v.mu.Lock()
		v.etapas[i].Razon = razon
		v.mu.Unlock()
	}
}

// Modelo fija el modelo, motor o regla que decidió la etapa V2.
func (v *RecorderV2) Modelo(id, modelo string) {
	if i, ok := v.indice(id); ok {
		v.mu.Lock()
		v.etapas[i].Modelo = modelo
		v.mu.Unlock()
	}
}

// CambiarEstado cambia el estado de una etapa V2 (y su razón, si se da).
func (v *RecorderV2) CambiarEstado(id, estado, razon string) {
	if !EstadoValido(estado) {
		return
	}
	if i, ok := v.indice(id); ok {
		v.mu.Lock()
		v.etapas[i].Estado = estado
		if razon != "" {
			v.etapas[i].Razon = razon
		}
		v.mu.Unlock()
	}
}

// Estado devuelve el estado de una etapa V2 cerrada ("" si sigue pendiente).
func (v *RecorderV2) Estado(id string) string {
	i, ok := v.indice(id)
	if !ok {
		return ""
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.fase[i] != cerrada {
		return ""
	}
	return v.etapas[i].Estado
}

// Pendiente dice si una etapa V2 aún no se cerró.
func (v *RecorderV2) Pendiente(id string) bool {
	i, ok := v.indice(id)
	if !ok {
		return false
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.fase[i] != cerrada
}

// etapasCerradas arma las 13 etapas V2 como las de V1: lo pendiente queda omitido, lo iniciado y sin cerrar en
// error, los datos mínimos siempre presentes y los textos recortados.
func (v *RecorderV2) etapasCerradas() []Etapa {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make([]Etapa, numEtapasV2)
	for i := range v.etapas {
		e := v.etapas[i]
		datos := make(Datos, len(e.Datos)+len(catalogoV2[i].minimos)+len(minimosComunesV2))
		for k, x := range e.Datos {
			datos[k] = sanear(x)
		}
		e.Datos = datos
		switch v.fase[i] {
		case pendiente:
			e.Estado = EstadoOmitida
			e.Ms = nil
			if e.Razon == "" {
				e.Razon = "no se ejecutó en este turno"
			}
		case iniciada:
			e.Estado = EstadoError
			e.Ms = msDe(time.Since(v.inicios[i]))
			if e.Razon == "" {
				e.Razon = "la etapa empezó y no terminó"
			}
		}
		for _, k := range append(append([]string(nil), minimosComunesV2...), catalogoV2[i].minimos...) {
			if _, ok := e.Datos[k]; !ok {
				e.Datos[k] = nil
			}
		}
		e.Razon = Recortar(e.Razon)
		e.Modelo = Recortar(e.Modelo)
		out[i] = e
	}
	return out
}

// cerrarV2 añade a la traza las etapas V2 y sus oportunidades, si el turno fue V2. Lo llama Cerrar.
func cerrarV2(r *Recorder, t *Traza) {
	r.mu.Lock()
	v := r.v2
	r.mu.Unlock()
	if v == nil {
		return
	}
	t.Agente = AgenteV2
	t.EtapasV2 = v.etapasCerradas()
	t.Oportunidades = append(t.Oportunidades, aplicarReglasV2(t.EtapasV2, len(t.Oportunidades))...)
}

// aplicarReglasV2: reglas fijas sobre los hechos de las etapas V2 (nunca un modelo). Misma semántica que las de V1:
// una regla que se cumple sube la etapa de «ok» a su nivel y añade una oportunidad numerada tras las de V1.
func aplicarReglasV2(etapas []Etapa, previas int) []Oportunidad {
	out := []Oportunidad{}
	agregar := func(e *Etapa, nivel, texto string) {
		if e.Estado == EstadoOK || (e.Estado == EstadoAlerta && nivel == EstadoError) {
			e.Estado = nivel
		}
		if e.Razon == "" {
			e.Razon = Recortar(texto)
		}
		out = append(out, Oportunidad{N: previas + len(out) + 1, Etapa: e.ID, Nivel: nivel, Texto: texto})
	}
	for i := range etapas {
		e := &etapas[i]
		if !ejecutada(e) {
			continue
		}
		switch e.ID {
		case EtapaV2TipoRespuesta:
			if d, _ := e.Datos["dudoso"].(bool); d {
				conf, _ := e.Datos["confianza"].(float64)
				agregar(e, EstadoAlerta, fmt.Sprintf(
					"Tipo de respuesta dudoso (ni regla clara ni kNN con margen; confianza %s): lo decidió el motor de decisión. Faltan reglas o ejemplos parecidos en el catálogo de tipos.",
					Decimal(conf, 2)))
			}
		case EtapaV2Decision:
			if r, _ := e.Datos["respaldo"].(bool); r {
				agregar(e, EstadoAlerta, "El motor de decisión falló o respondió fuera de las opciones ("+e.Razon+"): se decidió por reglas.")
			}
		case EtapaV2BusquedaVectorial:
			if e.Estado == EstadoRespaldo {
				agregar(e, EstadoAlerta, "Sin búsqueda vectorial: la recuperación fue solo léxica.")
			}
		case EtapaV2Rerank:
			if e.Estado == EstadoRespaldo {
				agregar(e, EstadoAlerta, "Sin reranker: el orden es el del puntaje híbrido.")
			}
		case EtapaV2Plan:
			if s, _ := e.Datos["sin_evidencia"].(bool); s {
				agregar(e, EstadoAlerta, "Sin evidencia que respalde una respuesta: falta un procedimiento o concepto sobre el tema en kb/.")
			}
		case EtapaV2QualityGate:
			if p, ok := e.Datos["passed"].(bool); ok && !p {
				agregar(e, EstadoError, "El quality gate rechazó la respuesta ("+e.Razon+"): salió la respuesta segura.")
			}
		case EtapaV2Renderizado:
			if e.Estado == EstadoRespaldo || e.Estado == EstadoError {
				agregar(e, EstadoAlerta, "Redacción degradada ("+e.Razon+"): revisar que el generador local esté disponible.")
			}
		}
	}
	return out
}
