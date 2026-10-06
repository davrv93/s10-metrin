package busqueda

import (
	"container/list"
	"context"
	"sync"

	"rag-go/internal/embed"
)

// LRU es una caché en memoria de tamaño fijo, segura para goroutines.
type LRU[K comparable, V any] struct {
	mu    sync.Mutex
	tam   int
	lista *list.List
	items map[K]*list.Element
	// Aciertos y Fallos cuentan las consultas (para la traza y las pruebas).
	Aciertos, Fallos int
}

type entradaLRU[K comparable, V any] struct {
	k K
	v V
}

// NuevoLRU crea una caché de tam entradas; tam <= 0 devuelve nil (sin
// caché: Obtener siempre falla y Poner no hace nada).
func NuevoLRU[K comparable, V any](tam int) *LRU[K, V] {
	if tam <= 0 {
		return nil
	}
	return &LRU[K, V]{tam: tam, lista: list.New(), items: make(map[K]*list.Element, tam)}
}

// Obtener devuelve el valor y lo marca como recién usado.
func (c *LRU[K, V]) Obtener(k K) (V, bool) {
	var cero V
	if c == nil {
		return cero, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.items[k]; ok {
		c.lista.MoveToFront(e)
		c.Aciertos++
		return e.Value.(*entradaLRU[K, V]).v, true
	}
	c.Fallos++
	return cero, false
}

// Poner guarda el valor; si no cabe, expulsa el menos usado.
func (c *LRU[K, V]) Poner(k K, v V) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.items[k]; ok {
		e.Value.(*entradaLRU[K, V]).v = v
		c.lista.MoveToFront(e)
		return
	}
	c.items[k] = c.lista.PushFront(&entradaLRU[K, V]{k, v})
	for c.lista.Len() > c.tam {
		u := c.lista.Back()
		c.lista.Remove(u)
		delete(c.items, u.Value.(*entradaLRU[K, V]).k)
	}
}

// Largo devuelve cuántas entradas hay.
func (c *LRU[K, V]) Largo() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lista.Len()
}

// EmbebedorConCache envuelve un embed.Embebedor con una LRU del vector de
// cada texto. Conserva Nombre(), así que almacen.Abrir abre la misma
// colección. Los errores no se guardan. Conviene envolver solo el embebedor
// de consulta (el de indexar llenaría la caché de textos de documentos).
type EmbebedorConCache struct {
	E     embed.Embebedor
	cache *LRU[string, []float32]
}

// NuevoEmbebedorConCache: tam <= 0 desactiva la caché.
func NuevoEmbebedorConCache(e embed.Embebedor, tam int) *EmbebedorConCache {
	return &EmbebedorConCache{E: e, cache: NuevoLRU[string, []float32](tam)}
}

func (e *EmbebedorConCache) Nombre() string { return e.E.Nombre() }

func (e *EmbebedorConCache) Embeber(ctx context.Context, texto string) ([]float32, error) {
	if v, ok := e.cache.Obtener(texto); ok {
		return append([]float32(nil), v...), nil
	}
	v, err := e.E.Embeber(ctx, texto)
	if err != nil {
		return nil, err
	}
	e.cache.Poner(texto, append([]float32(nil), v...))
	return v, nil
}

// Cache expone la LRU (para métricas).
func (e *EmbebedorConCache) Cache() *LRU[string, []float32] { return e.cache }
