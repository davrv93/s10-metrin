// Package almacen es el índice vectorial: chromem-go persistido en disco
// (un fichero gob por trozo) más un estado.json con la versión de cada
// documento para el reindexado incremental.
package almacen

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	chromem "github.com/philippgille/chromem-go"

	"rag-go/internal/embed"
)

// Fuentes conocidas (metadato "source").
const (
	FuenteRepo = "landing-repo"
	FuenteS3   = "garage-s3"
)

// IDTrozo es el id estable de un trozo: sha1("repo:<ruta>:<n>") o
// sha1("s3:<key>:<n>"). clave es "repo:<ruta>" o "s3:<key>".
func IDTrozo(clave string, n int) string {
	h := sha1.Sum(fmt.Appendf(nil, "%s:%d", clave, n))
	return hex.EncodeToString(h[:])
}

// EstadoDoc es lo que se recuerda de cada documento indexado.
type EstadoDoc struct {
	Fuente  string `json:"source"`
	Version string `json:"version"` // sha256 del contenido (repo) o ETag (S3)
	Trozos  int    `json:"trozos"`
}

// Trozo a indexar.
type Trozo struct {
	ID       string
	Texto    string
	Busqueda string // si no está vacío se embebe esto en vez de Texto
	Metadata map[string]string
}

// Resultado de una búsqueda.
type Resultado struct {
	ID        string
	Texto     string
	Metadata  map[string]string
	Distancia float64 // 1 - similitud coseno
}

type Almacen struct {
	dir    string
	db     *chromem.DB
	col    *chromem.Collection
	emb    embed.Embebedor
	mu     sync.Mutex
	estado map[string]EstadoDoc // clave → estado
}

// Abrir abre (o crea) el índice en dir. Si dir es "" el índice vive solo en
// memoria (para pruebas).
func Abrir(dir string, e embed.Embebedor) (*Almacen, error) {
	a := &Almacen{dir: dir, emb: e, estado: map[string]EstadoDoc{}}
	var err error
	if dir == "" {
		a.db = chromem.NewDB()
	} else {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		a.db, err = chromem.NewPersistentDB(filepath.Join(dir, "chromem"), false)
		if err != nil {
			return nil, fmt.Errorf("abrir chromem: %w", err)
		}
		if b, err := os.ReadFile(a.rutaEstado()); err == nil {
			if err := json.Unmarshal(b, &a.estado); err != nil {
				return nil, fmt.Errorf("estado.json dañado: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	fe := func(ctx context.Context, t string) ([]float32, error) { return e.Embeber(ctx, t) }
	// Una colección por embebedor: cambiar de proveedor no mezcla espacios.
	a.col, err = a.db.GetOrCreateCollection("rag-"+e.Nombre(), nil, fe)
	if err != nil {
		return nil, err
	}
	// Si la colección es nueva pero el estado es de otro embebedor, se
	// olvida el estado para forzar el reindexado.
	if a.col.Count() == 0 && len(a.estado) > 0 {
		a.estado = map[string]EstadoDoc{}
	}
	return a, nil
}

func (a *Almacen) rutaEstado() string {
	return filepath.Join(a.dir, "estado-"+a.emb.Nombre()+".json")
}

// Estado devuelve la versión recordada de un documento.
func (a *Almacen) Estado(clave string) (EstadoDoc, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e, ok := a.estado[clave]
	return e, ok
}

// Claves devuelve las claves de los documentos de una fuente.
func (a *Almacen) Claves(fuente string) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var out []string
	for k, e := range a.estado {
		if e.Fuente == fuente {
			out = append(out, k)
		}
	}
	return out
}

// Reemplazar borra los trozos anteriores del documento y añade los nuevos.
func (a *Almacen) Reemplazar(ctx context.Context, clave, fuente, version string, trozos []Trozo) error {
	if err := a.Borrar(ctx, clave); err != nil {
		return err
	}
	ids := make([]string, 0, len(trozos))
	vecs := make([][]float32, 0, len(trozos))
	metas := make([]map[string]string, 0, len(trozos))
	textos := make([]string, 0, len(trozos))
	for _, t := range trozos {
		clave := t.Texto
		if t.Busqueda != "" {
			clave = t.Busqueda
		}
		v, err := a.emb.Embeber(ctx, clave)
		if errors.Is(err, embed.ErrSinPiezas) {
			continue // trozo sin vocabulario (p. ej. solo símbolos)
		}
		if err != nil {
			return fmt.Errorf("embeber %s: %w", t.ID, err)
		}
		ids = append(ids, t.ID)
		vecs = append(vecs, v)
		metas = append(metas, t.Metadata)
		textos = append(textos, t.Texto)
	}
	if len(ids) > 0 {
		if err := a.col.Add(ctx, ids, vecs, metas, textos); err != nil {
			return err
		}
	}
	a.mu.Lock()
	a.estado[clave] = EstadoDoc{Fuente: fuente, Version: version, Trozos: len(trozos)}
	a.mu.Unlock()
	return nil
}

// Borrar quita todos los trozos del documento y su estado.
func (a *Almacen) Borrar(ctx context.Context, clave string) error {
	a.mu.Lock()
	prev, ok := a.estado[clave]
	a.mu.Unlock()
	if !ok || prev.Trozos == 0 {
		a.mu.Lock()
		delete(a.estado, clave)
		a.mu.Unlock()
		return nil
	}
	ids := make([]string, prev.Trozos)
	for i := range ids {
		ids[i] = IDTrozo(clave, i)
	}
	// chromem ignora los ids que no existen (trozos saltados por vacíos).
	if err := a.col.Delete(ctx, nil, nil, ids...); err != nil {
		return err
	}
	a.mu.Lock()
	delete(a.estado, clave)
	a.mu.Unlock()
	return nil
}

// Guardar persiste el estado (los trozos ya los persiste chromem al añadir).
func (a *Almacen) Guardar() error {
	if a.dir == "" {
		return nil
	}
	a.mu.Lock()
	b, err := json.MarshalIndent(a.estado, "", "  ")
	a.mu.Unlock()
	if err != nil {
		return err
	}
	tmp := a.rutaEstado() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, a.rutaEstado())
}

// Contar devuelve cuántos trozos hay en el índice.
func (a *Almacen) Contar() int { return a.col.Count() }

// Buscar devuelve los k trozos más cercanos que cumplan el filtro de
// metadatos (igualdad exacta: source, type, ext…).
func (a *Almacen) Buscar(ctx context.Context, pregunta string, k int, filtro map[string]string) ([]Resultado, error) {
	n := a.col.Count()
	if n == 0 {
		return nil, nil
	}
	k = min(max(k, 1), n)
	q, err := a.emb.Embeber(ctx, pregunta)
	if errors.Is(err, embed.ErrSinPiezas) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(filtro) == 0 {
		filtro = nil
	}
	rs, err := a.col.QueryEmbedding(ctx, q, k, filtro, nil)
	if err != nil {
		return nil, err
	}
	out := make([]Resultado, len(rs))
	for i, r := range rs {
		out[i] = Resultado{ID: r.ID, Texto: r.Content, Metadata: r.Metadata, Distancia: 1 - float64(r.Similarity)}
	}
	return out, nil
}
