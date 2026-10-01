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
	reCabecera = regexp.MustCompile(`(?i)^(?:cap[íi]tulo|secci[óo]n|paso|tema|gu[íi]a|p[aá]gina|manual|subsecci[óo]n)\b|^\d+([\.)]|\s+[A-ZÁÉÍÓÚÑ])`)
	reEtiqueta = regexp.MustCompile(`(?s)<[^>]+>`)
	reEspacios = regexp.MustCompile(`[ \t\r\f\v]+`)
	reLineas   = regexp.MustCompile(`\n\s*\n+`)
	entidades  = strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'", "&aacute;", "á", "&eacute;", "é", "&iacute;", "í", "&oacute;", "ó", "&uacute;", "ú", "&ntilde;", "ñ", "&euro;", "€", "&copy;", "©")
)

// fragmentosHTML devuelve bloques de texto en orden, manteniendo títulos y
// subtítulos como límites naturales del contenido. Esto conserva la estructura
// de un manual HTML aunque luego el troceado por tamaño siga aplicándose a cada
// bloque.
func fragmentosHTML(texto string) []string {
	lineas := strings.Split(texto, "\n")
	var bloques []string
	var actual []string
	flush := func() {
		if len(actual) == 0 {
			return
		}
		b := strings.TrimSpace(strings.Join(actual, "\n"))
		if b != "" {
			bloques = append(bloques, b)
		}
		actual = nil
	}
	for _, l := range lineas {
		linea := strings.TrimSpace(l)
		if linea == "" {
			if len(actual) > 0 {
				flush()
			}
			continue
		}
		if len(actual) > 0 && esCabeceraHTML(linea) {
			flush()
		}
		actual = append(actual, linea)
	}
	flush()
	return bloques
}

func esCabeceraHTML(linea string) bool {
	if linea == "" {
		return false
	}
	if utf8.RuneCountInString(linea) > 120 {
		return false
	}
	if reCabecera.MatchString(linea) {
		return true
	}
	if strings.Contains(linea, ":") && strings.Count(linea, " ") <= 8 {
		return true
	}
	if strings.ToUpper(linea) == linea && strings.ContainsAny(linea, "ABCDEFGHIJKLMNOPQRSTUVWXYZÁÉÍÓÚÑ") {
		return true
	}
	return false
}

func esPaginaAccesoHTML(texto string) bool {
	min := strings.ToLower(texto)
	login := strings.Contains(min, "iniciar sesión") || strings.Contains(min, "ingresar") || strings.Contains(min, "login") || strings.Contains(min, "simple membership")
	if !login {
		return false
	}
	if strings.Contains(min, "manual") || strings.Contains(min, "capítulo") || strings.Contains(min, "sección") {
		return false
	}
	palabras := len(strings.Fields(min))
	return palabras <= 80
}

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
		if contenido == "" || esPaginaAccesoHTML(contenido) {
			return nil
		}
		if bloques := fragmentosHTML(contenido); len(bloques) > 1 {
			var trozos []string
			for _, b := range bloques {
				if t := Trocear(b, Tamano, Solape); len(t) > 0 {
					trozos = append(trozos, t...)
				}
			}
			if len(trozos) > 0 {
				return trozos
			}
		}
	}
	return Trocear(contenido, Tamano, Solape)
}
