package indexar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"

	"rag-go/internal/almacen"
	"rag-go/internal/trocear"
)

// Repo es un directorio en disco (un repo git clonado, por ejemplo).
type Repo struct{ Raiz string }

func (r Repo) Fuente() string { return almacen.FuenteRepo }

func (r Repo) Cita(d Documento, n int) string { return "repo:" + d.Ruta + "#" + itoa(n) }

// Listar recorre la raíz saltando node_modules, .git y dist. La versión es el
// sha256 del contenido: hay que leer el fichero, pero es barato y no depende
// de mtime (que git no conserva).
func (r Repo) Listar(ctx context.Context) ([]Documento, error) {
	var out []Documento
	err := filepath.WalkDir(r.Raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != r.Raiz && trocear.Ignorados[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() || !trocear.EsTexto(p) {
			return nil
		}
		rel, err := filepath.Rel(r.Raiz, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		h := sha256.Sum256(b)
		contenido := string(b)
		out = append(out, Documento{
			Clave:   "repo:" + rel,
			Ruta:    rel,
			Version: hex.EncodeToString(h[:]),
			Cargar:  func(context.Context) (string, error) { return contenido, nil },
		})
		return ctx.Err()
	})
	return out, err
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
