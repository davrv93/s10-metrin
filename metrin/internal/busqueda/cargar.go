package busqueda

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// versionMetadatosKB debe ser igual a la de indexar/kb_jsonl.go: entra en la
// huella con la que el indexador desambigua ids repetidos, y los ids de este
// paquete tienen que coincidir con los del índice vectorial para fusionar.
const versionMetadatosKB = "kb-jsonl-metadata-v5"

type fragmentoJSON struct {
	ID        string `json:"id"`
	Documento string `json:"documento"`
	Manual    string `json:"manual"`
	Titulo    string `json:"titulo"`
	Seccion   string `json:"seccion"`
	Video     string `json:"video"`
	Pagina    any    `json:"pagina"`
	Fuente    string `json:"fuente"`
	URL       string `json:"url"`
	Desde     string `json:"desde"`
	Texto     string `json:"texto"`
	Confianza string `json:"confianza"`
	Busqueda  string `json:"busqueda"`
	Tipo      string `json:"tipo"`
	Pantalla  string `json:"pantalla"`
	Modulo    string `json:"modulo"`
	Imagen    string `json:"imagen"`
}

// RutasFragmentos devuelve kb/fragmentos*.jsonl en orden de bytes, el mismo
// que `cat /kb/fragmentos*.jsonl` de docker-entrada.sh (locale C).
func RutasFragmentos(dirKB string) ([]string, error) {
	rs, err := filepath.Glob(filepath.Join(dirKB, "fragmentos*.jsonl"))
	if err != nil {
		return nil, err
	}
	sort.Strings(rs)
	return rs, nil
}

// CargarFragmentos lee uno o varios JSONL de fragmentos como si fueran uno
// solo concatenado (así los indexa docker-entrada.sh) y devuelve un Doc por
// fragmento. El ID es el mismo que pone indexar.KBJSONL: el id del fragmento
// o, si se repite con otro contenido, id + "-" + 12 hex de su huella. Meta
// lleva el par (id_original, manual), que es la clave única de
// kb/procedimientos/ESQUEMA.md.
//
// Campos: titulo (encabezado propio del fragmento), seccion, preguntas (las
// variantes de búsqueda de las FAQ), texto, ocr (texto de las capturas) y
// manual (nombre del documento). Los pasos[] no se añaden: su texto ya está
// dentro de texto.
func CargarFragmentos(rutas ...string) ([]Doc, error) {
	var docs []Doc
	ids := map[string]string{}
	claves := map[string]string{}
	for _, ruta := range rutas {
		f, err := os.Open(ruta)
		if err != nil {
			return nil, err
		}
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
		linea := 0
		for sc.Scan() {
			linea++
			raw := sc.Bytes()
			if len(strings.TrimSpace(string(raw))) == 0 {
				continue
			}
			var fr fragmentoJSON
			if err := json.Unmarshal(raw, &fr); err != nil {
				f.Close()
				return nil, fmt.Errorf("%s:%d: %w", ruta, linea, err)
			}
			fr.ID = strings.TrimSpace(fr.ID)
			fr.Texto = strings.TrimSpace(fr.Texto)
			if fr.ID == "" || fr.Texto == "" {
				f.Close()
				return nil, fmt.Errorf("%s:%d: requiere id y texto", ruta, linea)
			}
			h := sha256.New()
			h.Write(raw)
			h.Write([]byte("\x00" + versionMetadatosKB))
			version := hex.EncodeToString(h.Sum(nil))
			original := fr.ID
			id := fr.ID
			if previa, ok := ids[id]; ok {
				if previa == version {
					continue
				}
				id += "-" + version[:12]
			}
			if previa, ok := claves[id]; ok {
				if previa == version {
					continue
				}
				f.Close()
				return nil, fmt.Errorf("%s:%d: colisión de id %q", ruta, linea, id)
			}
			ids[original] = version
			claves[id] = version
			docs = append(docs, docFragmento(id, original, fr))
		}
		err = sc.Err()
		f.Close()
		if err != nil {
			return nil, err
		}
	}
	return docs, nil
}

func docFragmento(id, original string, fr fragmentoJSON) Doc {
	pagina := ""
	switch p := fr.Pagina.(type) {
	case float64:
		pagina = strconv.Itoa(int(p))
	case string:
		pagina = p
	}
	titulo := strings.TrimSpace(fr.Titulo)
	if titulo == "" {
		titulo = strings.TrimSpace(fr.Video)
	}
	seccion := strings.TrimSpace(fr.Seccion)
	campos := map[string]string{}
	manual := strings.TrimSpace(strings.Join([]string{fr.Manual, fr.Modulo}, " "))
	// El título útil es el último tramo de «Manual › Sección»; si repite la
	// sección no se cuenta dos veces.
	encabezado := titulo
	if i := strings.LastIndex(encabezado, "›"); i >= 0 {
		encabezado = strings.TrimSpace(encabezado[i+len("›"):])
	}
	if encabezado == seccion {
		encabezado = ""
	}
	switch {
	case fr.Tipo == "imagen":
		// El título es el nombre del PNG: no aporta.
		campos["ocr"] = fr.Texto
		encabezado = ""
	case seccion == "" && pagina != "" && pagina != "0" && fr.Documento != "" && !strings.HasPrefix(fr.Documento, "cortex:"):
		// Página de un PDF: el «título» es el nombre del documento entero
		// (se repite en cientos de páginas), así que va con el manual.
		manual = strings.TrimSpace(manual + " " + titulo)
		encabezado = ""
		campos["texto"] = fr.Texto
	default:
		campos["texto"] = fr.Texto
	}
	if encabezado != "" {
		campos["titulo"] = encabezado
	}
	if seccion != "" {
		campos["seccion"] = seccion
	}
	if fr.Pantalla != "" && fr.Pantalla != titulo {
		campos["seccion"] = strings.TrimSpace(campos["seccion"] + " " + fr.Pantalla)
	}
	if b := strings.TrimSpace(fr.Busqueda); b != "" {
		campos["preguntas"] = b
	}
	if manual != "" {
		campos["manual"] = manual
	}
	fuente := strings.TrimSpace(fr.Fuente)
	if fuente == "" {
		fuente = strings.TrimSpace(fr.URL)
	}
	meta := map[string]string{
		// clave: el par (id, manual) es lo único que identifica una sección
		// web (el id trunca el manual a 8 caracteres y se repite entre
		// manuales; ver kb/procedimientos/ESQUEMA.md).
		"clave":       original + "|" + fr.Manual,
		"id_original": original,
		"manual":      fr.Manual,
		"documento":   fr.Documento,
		"titulo":      titulo,
		"seccion":     seccion,
		"pagina":      pagina,
		"confianza":   fr.Confianza,
		"tipo":        fr.Tipo,
		"fuente":      fuente,
	}
	if fr.Imagen != "" {
		meta["imagen"] = fr.Imagen
	}
	// Las mismas claves que pone indexar.KBJSONL en el almacén vectorial,
	// para que los filtros de rag (source, confianza, manual, title, page)
	// valgan igual en las dos búsquedas.
	titulo2 := titulo
	if titulo2 == "" {
		titulo2 = fr.Documento
	}
	pag := pagina
	if pag == "" {
		pag = "0"
	}
	cita := titulo2
	if fr.Desde != "" {
		cita += ", " + fr.Desde
	} else if pag != "0" {
		cita += ", p. " + pag
	}
	meta["source"] = "s10-kb"
	meta["title"] = titulo2
	meta["section"] = seccion
	meta["page"] = pag
	meta["cita"] = cita
	meta["document_id"] = fr.Documento
	meta["path"] = id + ".txt"
	return Doc{ID: id, Campos: campos, Meta: meta}
}

// PrefijoProcedimiento distingue los ids de procedimiento de los de
// fragmento.
const PrefijoProcedimiento = "proc:"

// CargarProcedimientos lee kb/procedimientos/**/*.yml (si el directorio no
// existe devuelve nil, nil). Salta las carpetas que empiezan por «_» o «.»
// (_reserva/: plantillas apartadas). Cada procedimiento es un Doc con ID
// "proc:<id>" y campos titulo, preguntas (preguntas + aliases), entidades
// (pantallas y objetos de `entidades`) y texto (objetivo, prerrequisitos,
// pasos con subpasos, verificación y errores).
func CargarProcedimientos(dir string) ([]Doc, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	var rutas []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && p != dir && (strings.HasPrefix(d.Name(), "_") || strings.HasPrefix(d.Name(), ".")) {
			return filepath.SkipDir
		}
		if !d.IsDir() && (strings.HasSuffix(p, ".yml") || strings.HasSuffix(p, ".yaml")) {
			rutas = append(rutas, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(rutas)
	var docs []Doc
	for _, r := range rutas {
		b, err := os.ReadFile(r)
		if err != nil {
			return nil, err
		}
		v, err := LeerYAML(b)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r, err)
		}
		m := yamlMapa(v)
		if m == nil {
			continue
		}
		id := yamlTexto(m["id"])
		if id == "" {
			id = strings.TrimSuffix(filepath.Base(r), filepath.Ext(r))
		}
		var texto []string
		if o := yamlTexto(m["objetivo"]); o != "" {
			texto = append(texto, o)
		}
		for _, campo := range []string{"prerrequisitos", "verificacion"} {
			for _, e := range yamlLista(m[campo]) {
				texto = append(texto, yamlTexto(yamlMapa(e)["texto"]))
			}
		}
		var fuentes []string
		texto = append(texto, textoPasos(yamlLista(m["pasos"]), &fuentes)...)
		for _, e := range yamlLista(m["errores_frecuentes"]) {
			em := yamlMapa(e)
			texto = append(texto, yamlTexto(em["sintoma"]), yamlTexto(em["solucion"]))
		}
		preguntas := append(yamlTextos(m["preguntas"]), yamlTextos(m["aliases"])...)
		ent := yamlMapa(m["entidades"])
		pantallas, objetos := yamlTextos(ent["pantallas"]), yamlTextos(ent["objetos"])
		docs = append(docs, Doc{
			ID: PrefijoProcedimiento + id,
			Campos: map[string]string{
				"titulo":    yamlTexto(m["titulo"]),
				"preguntas": strings.Join(preguntas, "\n"),
				"entidades": strings.Join(append(append([]string(nil), pantallas...), objetos...), "\n"),
				"texto":     strings.Join(noVacios(texto), "\n"),
			},
			Meta: map[string]string{
				"tipo":      "procedimiento",
				"id":        id,
				"modulo":    yamlTexto(m["modulo"]),
				"titulo":    yamlTexto(m["titulo"]),
				"nivel":     yamlTexto(m["nivel"]),
				"version":   yamlTexto(m["version"]),
				"ruta":      r,
				"fuentes":   strings.Join(unicos(fuentes), ","),
				"pantallas": strings.Join(pantallas, "|"),
				"objetos":   strings.Join(objetos, "|"),
			},
		})
	}
	return docs, nil
}

func textoPasos(pasos []any, fuentes *[]string) []string {
	var out []string
	for _, p := range pasos {
		pm := yamlMapa(p)
		if pm == nil {
			continue
		}
		for _, c := range []string{"accion", "donde", "resultado"} {
			if s := yamlTexto(pm[c]); s != "" {
				out = append(out, strings.ReplaceAll(s, "**", ""))
			}
		}
		*fuentes = append(*fuentes, yamlTextos(pm["fuente"])...)
		out = append(out, textoPasos(yamlLista(pm["sub"]), fuentes)...)
	}
	return out
}

func noVacios(ss []string) []string {
	out := ss[:0]
	for _, s := range ss {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func unicos(ss []string) []string {
	vistos := map[string]bool{}
	var out []string
	for _, s := range ss {
		if !vistos[s] {
			vistos[s] = true
			out = append(out, s)
		}
	}
	return out
}

// IndexarDocs agrega docs a un índice nuevo.
func IndexarDocs(cfg ConfigBM25, docs []Doc) (*Indice, error) {
	ix := NuevoIndice(cfg)
	for _, d := range docs {
		if err := ix.Agregar(d); err != nil {
			return nil, err
		}
	}
	return ix, nil
}
