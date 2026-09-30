package aprender

import (
	"sync"
	"time"
)

// Evento es lo que viaja por SSE: `event: <Tipo>` y `data: <Datos en JSON>`.
type Evento struct {
	Tipo  string `json:"tipo"`
	Fecha string `json:"fecha"`
	Datos any    `json:"datos"`
}

// Bus reparte eventos a los suscriptores (una conexión SSE cada uno). Nunca
// bloquea al que publica: si un cliente va lento, pierde eventos.
type Bus struct {
	mu   sync.Mutex
	subs map[chan Evento]struct{}
}

func NuevoBus() *Bus { return &Bus{subs: map[chan Evento]struct{}{}} }

// Suscribir devuelve el canal de eventos y la función para soltarlo.
func (b *Bus) Suscribir() (<-chan Evento, func()) {
	c := make(chan Evento, 64)
	b.mu.Lock()
	b.subs[c] = struct{}{}
	b.mu.Unlock()
	return c, func() {
		b.mu.Lock()
		if _, ok := b.subs[c]; ok {
			delete(b.subs, c)
			close(c)
		}
		b.mu.Unlock()
	}
}

// Publicar envía el evento a todos sin esperar.
func (b *Bus) Publicar(tipo string, datos any) {
	ev := Evento{Tipo: tipo, Fecha: time.Now().Format(time.RFC3339), Datos: datos}
	b.mu.Lock()
	defer b.mu.Unlock()
	for c := range b.subs {
		select {
		case c <- ev:
		default:
		}
	}
}
