package busqueda

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// versionGob cambia si cambia el formato del índice o el analizador (texto.go):
// un índice guardado con otro analizador daría términos que la consulta ya no
// produce.
const versionGob = "busqueda-bm25f-v1"

type instantanea struct {
	Version  string
	Cfg      ConfigBM25
	Campos   []string
	IDs      []string
	Meta     []map[string]string
	Texto    []string
	Terms    []string
	Post     [][]Posting
	DF       []int32
	Largos   [][]uint32
	Suma     []float64
	ConCampo []int
}

// Guardar escribe el índice en gob.
func (ix *Indice) Guardar(w io.Writer) error {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	return gob.NewEncoder(w).Encode(instantanea{
		Version: versionGob, Cfg: ix.cfg, Campos: ix.campos, IDs: ix.ids, Meta: ix.meta,
		Texto: ix.texto, Terms: ix.terms, Post: ix.post, DF: ix.df, Largos: ix.largos,
		Suma: ix.suma, ConCampo: ix.conCampo,
	})
}

// GuardarArchivo escribe el índice de forma atómica (tmp + rename).
func (ix *Indice) GuardarArchivo(ruta string) error {
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return err
	}
	tmp := ruta + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	w := bufio.NewWriterSize(f, 1<<20)
	if err := ix.Guardar(w); err != nil {
		f.Close()
		return err
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, ruta)
}

// CargarIndice lee un índice escrito por Guardar.
func CargarIndice(r io.Reader) (*Indice, error) {
	var s instantanea
	if err := gob.NewDecoder(r).Decode(&s); err != nil {
		return nil, err
	}
	if s.Version != versionGob {
		return nil, fmt.Errorf("índice BM25 de versión %q, se esperaba %q: reindexar", s.Version, versionGob)
	}
	ix := NuevoIndice(s.Cfg)
	ix.campos, ix.ids, ix.meta, ix.texto = s.Campos, s.IDs, s.Meta, s.Texto
	ix.terms, ix.post, ix.df, ix.largos = s.Terms, s.Post, s.DF, s.Largos
	ix.suma, ix.conCampo = s.Suma, s.ConCampo
	for i, c := range ix.campos {
		ix.idxCampo[c] = uint8(i)
	}
	for i, id := range ix.ids {
		ix.porID[id] = int32(i)
	}
	for i, t := range ix.terms {
		ix.vocab[t] = int32(i)
		if len(t) == 0 || t[:1] != prefijoExacto {
			l := utf8.RuneCountInString(t)
			ix.porLargo[l] = append(ix.porLargo[l], int32(i))
		}
	}
	return ix, nil
}

// CargarIndiceArchivo abre y lee un índice guardado.
func CargarIndiceArchivo(ruta string) (*Indice, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return CargarIndice(bufio.NewReaderSize(f, 1<<20))
}
