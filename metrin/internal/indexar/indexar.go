// Package indexar recorre un origen (repo en disco o bucket S3), decide qué
// documentos cambiaron y actualiza el almacén de forma incremental.
package indexar

import (
	"context"
	"fmt"
	"path"
	"strings"

	"rag-go/internal/almacen"
	"rag-go/internal/trocear"
)

// Documento es una entrada del origen. Version es barata de obtener (hash del
// contenido en disco, ETag en S3); Cargar solo se llama si cambió.
type Documento struct {
	Clave    string // "repo:<ruta>" o "s3:<key>"
	Ruta     string // ruta relativa o key, para extensión y metadatos
	Version  string
	Cargar   func(ctx context.Context) (string, error)
	Metadata map[string]string
}

// Origen lista los documentos de una fuente.
type Origen interface {
	Fuente() string
	Listar(ctx context.Context) ([]Documento, error)
	// Cita da el texto de la cita de un trozo n, p. ej. "repo:index.html#2".
	Cita(d Documento, n int) string
}

// Informe de un indexado.
type Informe struct {
	Vistos, Nuevos, Cambiados, Iguales, Borrados, Trozos int
}

func (i Informe) String() string {
	return fmt.Sprintf("documentos=%d nuevos=%d cambiados=%d iguales=%d borrados=%d trozos_escritos=%d",
		i.Vistos, i.Nuevos, i.Cambiados, i.Iguales, i.Borrados, i.Trozos)
}

// Tipo clasifica el documento para el filtro "type".
func Tipo(ruta string) string {
	switch strings.ToLower(path.Ext(ruta)) {
	case ".md", ".mdx", ".txt":
		return "doc"
	case ".html", ".astro":
		return "html"
	case ".css":
		return "estilo"
	case ".js", ".ts", ".tsx":
		return "codigo"
	case ".json", ".yaml", ".yml", ".xml", ".csv":
		return "datos"
	}
	return "otro"
}

// Ejecutar indexa el origen en el almacén: salta lo igual, reemplaza lo
// cambiado y borra lo que ya no existe en el origen.
func Ejecutar(ctx context.Context, o Origen, a *almacen.Almacen, log func(string, ...any)) (Informe, error) {
	if log == nil {
		log = func(string, ...any) {}
	}
	docs, err := o.Listar(ctx)
	if err != nil {
		return Informe{}, err
	}
	var inf Informe
	vistos := make(map[string]bool, len(docs))
	for _, d := range docs {
		inf.Vistos++
		vistos[d.Clave] = true
		prev, existia := a.Estado(d.Clave)
		if existia && prev.Version == d.Version {
			inf.Iguales++
			continue
		}
		contenido, err := d.Cargar(ctx)
		if err != nil {
			return inf, fmt.Errorf("leer %s: %w", d.Clave, err)
		}
		textos := trocear.Preparar(d.Ruta, contenido)
		ext := strings.ToLower(path.Ext(d.Ruta))
		trozos := make([]almacen.Trozo, len(textos))
		for n, t := range textos {
			metadata := map[string]string{
				"source": o.Fuente(),
				"type":   Tipo(d.Ruta),
				"ext":    ext,
				"path":   d.Ruta,
				"chunk":  fmt.Sprint(n),
				"cita":   o.Cita(d, n),
			}
			for k, v := range d.Metadata {
				metadata[k] = v
			}
			trozos[n] = almacen.Trozo{
				ID:       almacen.IDTrozo(d.Clave, n),
				Texto:    t,
				Metadata: metadata,
			}
		}
		if err := a.Reemplazar(ctx, d.Clave, o.Fuente(), d.Version, trozos); err != nil {
			return inf, err
		}
		inf.Trozos += len(trozos)
		if existia {
			inf.Cambiados++
			log("  ~ %s (%d trozos)", d.Clave, len(trozos))
		} else {
			inf.Nuevos++
			log("  + %s (%d trozos)", d.Clave, len(trozos))
		}
	}
	for _, clave := range a.Claves(o.Fuente()) {
		if ab, ok := o.(interface{ Abarca(string) bool }); ok && !ab.Abarca(clave) {
			continue // fuera del prefijo listado: no se toca
		}
		if !vistos[clave] {
			if err := a.Borrar(ctx, clave); err != nil {
				return inf, err
			}
			inf.Borrados++
			log("  - %s", clave)
		}
	}
	return inf, a.Guardar()
}
