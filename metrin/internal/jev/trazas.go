package jev

// Emisor de trazas: cada decisión JEV real se graba en un JSONL
// (una traza por línea) con el mismo formato que lee el visor
// de trazas (trazas/trazas.py). El visor consume el archivo por
// la abstracción RepositorioArchivo; en Fase 3/5 las filas van
// a Postgres (trazas/esquema.sql) y el emisor no cambia.
//
// El registro es opt-in: solo cuando se adjunta un RegistroTrazas
// al Cliente. Grabar falla en silencio: la decisión ya está
// tomada; perder la traza no puede romper la petición.

import (
	"crypto/rand"
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

// RegistroTrazas añade trazas a un JSONL. Es seguro para
// concurrencia: sirve un servidor que decide en paralelo.
type RegistroTrazas struct {
	Ruta   string // archivo JSONL (se crea si no existe)
	Origen string // de dónde viene la decisión (p. ej. "reportes")

	mu sync.Mutex
}

// trazaRespuesta es la respuesta en formato del visor: claves
// en español (probabilidades, confianza, leyenda), igual que
// los mocks de trazas/mock/trazas.json.
type trazaRespuesta struct {
	Tipo           string              `json:"tipo"`
	Choice         string              `json:"choice,omitempty"`
	Probabilidades map[string]float64  `json:"probabilidades,omitempty"`
	Confianza      *float64            `json:"confianza,omitempty"`
	Score          float64             `json:"score,omitempty"`
	Leyenda        map[string]any      `json:"leyenda,omitempty"`
	Noul           float64             `json:"noul,omitempty"`
}

// trazaPregunta es una pregunta con su respuesta, en orden
// estable (el mapa de la petición no garantiza orden).
type trazaPregunta struct {
	ID            string          `json:"id"`
	Tipo          string          `json:"tipo"`
	Instrucciones any             `json:"instrucciones,omitempty"`
	Criterios     any             `json:"criterios,omitempty"`
	Respuesta     *trazaRespuesta `json:"respuesta,omitempty"`
}

// traza es la traza completa: la petición a /v1/systemone y
// lo que el servidor respondió. Correcta es el veredicto de la
// muestra etiquetada: las trazas reales aún no se evalúan
// (null = sin evaluar).
type traza struct {
	ID            string          `json:"id"`
	TS            string          `json:"ts"`
	Modelo        string          `json:"modelo,omitempty"`
	Origen        string          `json:"origen"`
	Plantilla     string          `json:"plantilla,omitempty"`
	MS            int64           `json:"ms"`
	TokensEntrada int             `json:"tokens_entrada,omitempty"`
	Estado        any             `json:"estado"`
	Preguntas     []trazaPregunta `json:"preguntas"`
	Correcta      *bool           `json:"correcta"`
}

// nuevaTraza arma la traza de una decisión ya respondida.
// respuestas viene del servidor, indexada por id de pregunta.
func nuevaTraza(modelo, origen string, estado any, preguntas map[string]Pregunta, respuestas map[string]Respuesta, ms int64, tokens int) traza {
	ids := make([]string, 0, len(preguntas))
	for id := range preguntas {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	ps := make([]trazaPregunta, 0, len(ids))
	for _, id := range ids {
		p := preguntas[id]
		tp := trazaPregunta{ID: id, Tipo: p.Tipo, Instrucciones: p.Instrucciones, Criterios: p.Criterios}
		if r, ok := respuestas[id]; ok {
			tp.Respuesta = respuestaATraza(r)
		}
		ps = append(ps, tp)
	}
	return traza{
		ID:            nuevoULID(),
		TS:            time.Now().Format(time.RFC3339),
		Modelo:        modelo,
		Origen:        origen,
		MS:            ms,
		TokensEntrada: tokens,
		Estado:        estado,
		Preguntas:     ps,
	}
}

// respuestaATraza traduce la respuesta de la API (claves en
// inglés) al formato del visor. Confianza es *float64: un 0.0
// de la API (distribución uniforme) se guarda como ausente,
// igual que el visor muestra «sin confianza».
func respuestaATraza(r Respuesta) *trazaRespuesta {
	t := &trazaRespuesta{
		Tipo:           r.Tipo,
		Choice:         r.Choice,
		Probabilidades: r.Probabilidades,
		Score:          r.Score,
		Leyenda:        r.Leyenda,
		Noul:           r.Noul,
	}
	if r.Confianza > 0 {
		c := r.Confianza
		t.Confianza = &c
	}
	return t
}

// guardar añade una traza al JSONL.
func (r *RegistroTrazas) guardar(t traza) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, err := os.OpenFile(r.Ruta, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

// alfabetoULID es Crockford base32, sin vocales: ids de 26
// caracteres, lexicográficamente ordenables por tiempo.
const alfabetoULID = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// nuevoULID genera un ULID: 48 bits de milisegundos + 80 bits
// aleatorios. El visor no exige ULID estricto, pero así los ids
// son únicos y ordenables sin mirar el ts.
func nuevoULID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand no falla en la práctica; último recurso
		// determinista por tiempo para no bloquear la decisión.
		for i := range b {
			b[i] = byte(time.Now().UnixNano() >> (uint(i%8) * 8))
		}
	}
	ms := uint64(time.Now().UnixMilli())
	out := make([]byte, 26)
	for i := 0; i < 10; i++ { // 48 bits de tiempo → 10 caracteres
		out[i] = alfabetoULID[(ms>>uint(45-i*5))&0x1f]
	}
	for i := 0; i < 16; i++ { // 80 bits aleatorios → 16 caracteres
		bit := uint(i * 5)
		v := uint(b[bit/8]) >> (bit % 8)
		if bit%8 > 3 {
			v |= uint(b[bit/8+1]) << (8 - bit%8)
		}
		out[10+i] = alfabetoULID[v&0x1f]
	}
	return string(out)
}
