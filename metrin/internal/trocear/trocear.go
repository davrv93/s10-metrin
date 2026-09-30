// Package trocear parte documentos en trozos solapados (por runas) y limpia
// el HTML antes de trocearlo.
package trocear

import (
	"path"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	Tamano = 800
	Solape = 150
)

// Extensiones que se indexan como texto.
var Extensiones = map[string]bool{
	".md": true, ".mdx": true, ".txt": true, ".html": true, ".css": true,
	".js": true, ".ts": true, ".tsx": true, ".astro": true, ".json": true,
	".yaml": true, ".yml": true, ".csv": true, ".xml": true,
}

// Ignorados: directorios que no se recorren.
var Ignorados = map[string]bool{"node_modules": true, ".git": true, "dist": true}

// EsTexto dice si la ruta tiene una extensión indexable.
func EsTexto(ruta string) bool { return Extensiones[strings.ToLower(path.Ext(ruta))] }

// Trocear parte texto en trozos de tamano runas; cada trozo empieza solape
// runas antes de donde acabó el anterior. Recorta espacios en los extremos y
// descarta trozos vacíos.
func Trocear(texto string, tamano, solape int) []string {
	if tamano <= 0 {
		tamano = Tamano
	}
	if solape < 0 || solape >= tamano {
		solape = 0
	}
	texto = strings.TrimSpace(texto)
	if texto == "" {
		return nil
	}
	if !utf8.ValidString(texto) {
		texto = strings.ToValidUTF8(texto, "�")
	}
	r := []rune(texto)
	var out []string
	paso := tamano - solape
	for ini := 0; ini < len(r); ini += paso {
		fin := min(ini+tamano, len(r))
		if t := strings.TrimSpace(string(r[ini:fin])); t != "" {
			out = append(out, t)
		}
		if fin == len(r) {
			break
		}
	}
	return out
}

var (
	reBloque   = regexp.MustCompile(`(?is)<(script|style|noscript|template|svg)\b[^>]*>.*?</(script|style|noscript|template|svg)\s*>`)
	reComent   = regexp.MustCompile(`(?s)<!--.*?-->`)
	reSalto    = regexp.MustCompile(`(?i)</?(p|div|section|article|header|footer|main|nav|li|ul|ol|h[1-6]|br|tr|table|form|label|details|summary|dt|dd)\b[^>]*>`)
	reEtiqueta = regexp.MustCompile(`(?s)<[^>]+>`)
	reEspacios = regexp.MustCompile(`[ \t\r\f\v]+`)
	reLineas   = regexp.MustCompile(`\n\s*\n+`)
	entidades  = strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'", "&aacute;", "á", "&eacute;", "é", "&iacute;", "í", "&oacute;", "ó", "&uacute;", "ú", "&ntilde;", "ñ", "&euro;", "€", "&copy;", "©")
)

// LimpiarHTML quita script/style/comentarios y etiquetas, conserva el texto
// visible con saltos de línea entre bloques y decodifica entidades comunes.
// Los atributos útiles para preguntas (name, type, required, action…) se
// pierden a propósito: la lógica del formulario vive en su JS, que se indexa
// aparte.
func LimpiarHTML(html string) string {
	s := reComent.ReplaceAllString(html, " ")
	s = reBloque.ReplaceAllString(s, " ")
	s = reSalto.ReplaceAllString(s, "\n")
	s = reEtiqueta.ReplaceAllString(s, " ")
	s = entidades.Replace(s)
	s = reEspacios.ReplaceAllString(s, " ")
	lineas := strings.Split(s, "\n")
	for i, l := range lineas {
		lineas[i] = strings.TrimSpace(l)
	}
	s = strings.Join(lineas, "\n")
	s = reLineas.ReplaceAllString(s, "\n")
	return strings.TrimSpace(s)
}

// Preparar limpia (si es HTML) y trocea con los valores por defecto.
func Preparar(ruta, contenido string) []string {
	if e := strings.ToLower(path.Ext(ruta)); e == ".html" || e == ".htm" {
		contenido = LimpiarHTML(contenido)
	}
	return Trocear(contenido, Tamano, Solape)
}
