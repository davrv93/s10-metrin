package indexar

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const FuenteS10KB = "s10-kb"

const versionMetadatosKB = "kb-jsonl-metadata-v3"

type KBJSONL struct{ Ruta string }

type fragmentoKB struct {
	ID        string `json:"id"`
	Documento string `json:"documento"`
	Manual    string `json:"manual"`
	Titulo    string `json:"titulo"`
	Video     string `json:"video"`
	Pagina    int    `json:"pagina"`
	Fuente    string `json:"fuente"`
	URL       string `json:"url"`
	Desde     string `json:"desde"`
	Texto     string `json:"texto"`
	// Busqueda, si viene, es lo que se embebe en lugar de Texto (p. ej. las
	// preguntas frecuentes que apuntan a un pasaje de Cortex). Texto sigue
	// siendo lo que recibe el LLM, y el fragmento no se trocea.
	Busqueda string `json:"busqueda"`
}

func (KBJSONL) Fuente() string { return FuenteS10KB }

func (KBJSONL) Cita(d Documento, _ int) string { return d.Metadata["cita"] }

func (o KBJSONL) Listar(ctx context.Context) ([]Documento, error) {
	f, err := os.Open(o.Ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var docs []Documento
	ids := map[string]string{}
	claves := map[string]string{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	linea := 0
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		linea++
		raw := append([]byte(nil), scanner.Bytes()...)
		if len(strings.TrimSpace(string(raw))) == 0 {
			continue
		}
		var frag fragmentoKB
		if err := json.Unmarshal(raw, &frag); err != nil {
			return nil, fmt.Errorf("JSONL línea %d: %w", linea, err)
		}
		frag.ID = strings.TrimSpace(frag.ID)
		frag.Texto = strings.TrimSpace(frag.Texto)
		if frag.ID == "" || frag.Texto == "" {
			return nil, fmt.Errorf("JSONL línea %d: requiere id y texto", linea)
		}
		h := sha256.New()
		h.Write(raw)
		h.Write([]byte("\x00" + versionMetadatosKB))
		version := hex.EncodeToString(h.Sum(nil))
		if previa, exists := ids[frag.ID]; exists {
			if previa == version {
				continue
			}
			frag.ID += "-" + version[:12]
		}
		if previa, exists := claves[frag.ID]; exists {
			if previa == version {
				continue
			}
			return nil, fmt.Errorf("JSONL línea %d: colisión de ID %q", linea, frag.ID)
		}
		ids[strings.TrimSuffix(frag.ID, "-"+version[:12])] = version
		claves[frag.ID] = version

		docs = append(docs, documentoKB(frag, version))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return docs, nil
}

// DocumentoKB convierte una línea JSONL de fragmento en un Documento del
// origen s10-kb, con la misma versión que calcularía Listar (así el siguiente
// index-kb lo ve igual y no lo reindexa).
func DocumentoKB(raw []byte) (Documento, error) {
	var frag fragmentoKB
	if err := json.Unmarshal(raw, &frag); err != nil {
		return Documento{}, err
	}
	frag.ID = strings.TrimSpace(frag.ID)
	frag.Texto = strings.TrimSpace(frag.Texto)
	if frag.ID == "" || frag.Texto == "" {
		return Documento{}, fmt.Errorf("fragmento: requiere id y texto")
	}
	h := sha256.New()
	h.Write(raw)
	h.Write([]byte("\x00" + versionMetadatosKB))
	return documentoKB(frag, hex.EncodeToString(h.Sum(nil))), nil
}

func documentoKB(frag fragmentoKB, version string) Documento {
	titulo := strings.TrimSpace(frag.Titulo)
	if titulo == "" {
		titulo = strings.TrimSpace(frag.Video)
	}
	if titulo == "" {
		titulo = frag.Documento
	}
	cita := titulo
	if frag.Desde != "" {
		cita += ", " + frag.Desde
	} else if frag.Pagina > 0 {
		cita += fmt.Sprintf(", p. %d", frag.Pagina)
	}
	fuente := strings.TrimSpace(frag.Fuente)
	if fuente == "" {
		fuente = strings.TrimSpace(frag.URL)
	}
	metadata := map[string]string{
		"source":      FuenteS10KB,
		"document_id": frag.Documento,
		"manual":      frag.Manual,
		"title":       titulo,
		"page":        fmt.Sprint(frag.Pagina),
		"source_url":  fuente,
		"cita":        cita,
	}
	busqueda := strings.TrimSpace(frag.Busqueda)
	if busqueda != "" {
		// Primera línea = la pregunta bien escrita: sirve para sugerir
		// «¿quisiste preguntar…?» cuando no hay respuesta.
		metadata["pregunta"], _, _ = strings.Cut(busqueda, "\n")
	}
	return Documento{
		Clave:    "s10kb:" + frag.ID,
		Ruta:     frag.ID + ".txt",
		Version:  version,
		Metadata: metadata,
		Busqueda: busqueda,
		Cargar:   func(context.Context) (string, error) { return frag.Texto, nil },
	}
}
