package busqueda

import (
	"context"
	"strings"

	"rag-go/internal/almacen"
)

// Candidato es un resultado de la búsqueda vectorial.
type Candidato struct {
	ID        string
	Similitud float64 // coseno, 1 − distancia
	Texto     string  // texto del trozo (por si el índice léxico no tiene el doc)
	Meta      map[string]string
}

// Vectorial es la búsqueda semántica. VectorAlmacen la implementa sobre el
// almacen.Almacen de hoy, sin cambiarlo.
type Vectorial interface {
	BuscarVector(ctx context.Context, q string, k int) ([]Candidato, error)
}

// VectorialFiltrado es un Vectorial que filtra por metadatos en origen. Si el
// Vectorial no lo implementa, el Buscador filtra después.
type VectorialFiltrado interface {
	BuscarVectorFiltro(ctx context.Context, q string, k int, filtro map[string]string) ([]Candidato, error)
}

// VectorAlmacen adapta almacen.Almacen.Buscar (solo lectura). El almacén
// guarda trozos (un fragmento largo puede tener varios): se piden k·Factor y
// se queda el mejor trozo de cada documento.
type VectorAlmacen struct {
	A      *almacen.Almacen
	Filtro map[string]string
	Factor int // defecto 3
	// IDDe da el id de documento de un trozo; defecto IDFragmento.
	IDDe func(almacen.Resultado) string
}

// IDFragmento devuelve el id de fragmento de un trozo indexado por
// indexar.KBJSONL (metadata path = "<id>.txt"); si no lo hay, el id del trozo.
func IDFragmento(r almacen.Resultado) string {
	if p := r.Metadata["path"]; p != "" && r.Metadata["source"] != "" {
		return strings.TrimSuffix(p, ".txt")
	}
	return r.ID
}

func (v VectorAlmacen) BuscarVector(ctx context.Context, q string, k int) ([]Candidato, error) {
	return v.BuscarVectorFiltro(ctx, q, k, nil)
}

// BuscarVectorFiltro suma filtro al Filtro fijo del adaptador.
func (v VectorAlmacen) BuscarVectorFiltro(ctx context.Context, q string, k int, filtro map[string]string) ([]Candidato, error) {
	f := map[string]string{}
	for c, x := range v.Filtro {
		f[c] = x
	}
	for c, x := range filtro {
		f[c] = x
	}
	fac := v.Factor
	if fac <= 0 {
		fac = 3
	}
	idDe := v.IDDe
	if idDe == nil {
		idDe = IDFragmento
	}
	rs, err := v.A.Buscar(ctx, q, k*fac, f)
	if err != nil {
		return nil, err
	}
	out := make([]Candidato, 0, k)
	vistos := map[string]bool{}
	for _, r := range rs {
		id := idDe(r)
		if vistos[id] {
			continue
		}
		vistos[id] = true
		out = append(out, Candidato{ID: id, Similitud: 1 - r.Distancia, Texto: r.Texto, Meta: r.Metadata})
		if len(out) == k {
			break
		}
	}
	return out, nil
}
