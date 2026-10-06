package main

// El dataset de oro: metrin/eval/v2_oro.jsonl, un caso por línea. Las expectativas se refieren al ÚLTIMO
// turno de `turnos` (los anteriores son contexto de la conversación y se reproducen igual que la página).

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Categorías del dataset (§12 de docs/V2-RAG-PROCEDURAL.md).
var categorias = []string{"CONCEPT", "PROCEDURE", "NAVIGATION", "TROUBLESHOOTING", "CONFIGURATION", "AMBIGUOUS"}

// Caso del dataset.
type Caso struct {
	ID                    string              `json:"id"`
	Categoria             string              `json:"categoria"`
	Turnos                []string            `json:"turnos"`
	TipoEsperado          string              `json:"tipo_esperado"`
	TiposAceptables       []string            `json:"tipos_aceptables,omitempty"`
	ProcedimientoEsperado *string             `json:"procedimiento_esperado"`
	ConceptoEsperado      *string             `json:"concepto_esperado"`
	PasosEsperados        []int               `json:"pasos_esperados"`        // null = todos los de primer nivel
	Continuacion          bool                `json:"continuacion,omitempty"` // «¿y luego?»: los que siguen al último entregado
	FotosEsperadas        map[string][]string `json:"fotos_esperadas"`        // n de primer nivel → claves de foto
	FragmentosRelevantes  []RefFragmento      `json:"fragmentos_relevantes"`  // pares {id, manual}
	TerminosEsperados     []string            `json:"terminos_esperados,omitempty"`
	SinEvidenciaEsperada  bool                `json:"sin_evidencia_esperada"`
	Sintetico             bool                `json:"sintetico"`
	Variantes             []string            `json:"variantes,omitempty"`
	Justificacion         string              `json:"justificacion,omitempty"`
}

// RefFragmento: el par (id, manual), único en kb/fragmentos*.jsonl (el id solo se repite entre manuales:
// ver «ids ambiguos» en kb/procedimientos/ESQUEMA.md). En el JSONL va como {"id": …, "manual": …}; también se
// acepta la cadena "id@manual" o "id" a secas (esta última solo si el id no es ambiguo).
type RefFragmento struct {
	ID     string `json:"id"`
	Manual string `json:"manual,omitempty"`
}

func (r RefFragmento) Clave() string {
	if r.Manual == "" {
		return r.ID
	}
	return claveFragmento(r.ID, r.Manual)
}

func (r *RefFragmento) UnmarshalJSON(b []byte) error {
	var s string
	if json.Unmarshal(b, &s) == nil {
		if i := strings.LastIndex(s, "@"); i > 0 {
			r.ID, r.Manual = s[:i], s[i+1:]
		} else {
			r.ID = s
		}
		return nil
	}
	type alias RefFragmento
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*r = RefFragmento(a)
	return nil
}

func (c *Caso) Proc() string {
	if c.ProcedimientoEsperado == nil {
		return ""
	}
	return *c.ProcedimientoEsperado
}

func (c *Caso) Concepto() string {
	if c.ConceptoEsperado == nil {
		return ""
	}
	return *c.ConceptoEsperado
}

// UnmarshalJSON admite `turnos` como lista de textos o de objetos {texto|pregunta}.
func (c *Caso) UnmarshalJSON(b []byte) error {
	type alias Caso
	var aux struct {
		alias
		Turnos []json.RawMessage `json:"turnos"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	*c = Caso(aux.alias)
	c.Turnos = nil
	for _, t := range aux.Turnos {
		var s string
		if json.Unmarshal(t, &s) == nil {
			c.Turnos = append(c.Turnos, s)
			continue
		}
		var o struct {
			Texto    string `json:"texto"`
			Pregunta string `json:"pregunta"`
		}
		if err := json.Unmarshal(t, &o); err != nil {
			return err
		}
		if o.Texto == "" {
			o.Texto = o.Pregunta
		}
		c.Turnos = append(c.Turnos, o.Texto)
	}
	return nil
}

func (c Caso) MarshalJSON() ([]byte, error) {
	type alias Caso
	return json.Marshal(alias(c))
}

func cargarDataset(ruta string) ([]Caso, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Caso
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	n := 0
	for sc.Scan() {
		n++
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "//") {
			continue
		}
		var c Caso
		if err := json.Unmarshal([]byte(l), &c); err != nil {
			return nil, fmt.Errorf("%s línea %d: %w", ruta, n, err)
		}
		out = append(out, c)
	}
	return out, sc.Err()
}

// validarDataset comprueba cada expectativa contra la KB vigente. Devuelve errores (el caso no se puede
// calificar como está) y avisos (se califica, pero algo cambió en la KB).
func validarDataset(casos []Caso, b *Base) (errores, avisos []string) {
	vistos := map[string]bool{}
	tipos := map[string]bool{"CONCEPT": true, "PROCEDURE": true, "NAVIGATION": true, "TROUBLESHOOTING": true,
		"CONFIGURATION": true, "COMPARISON": true, "UNKNOWN": true, "SOCIAL": true}
	cats := map[string]bool{}
	for _, c := range categorias {
		cats[c] = true
	}
	for _, c := range casos {
		e := func(f string, a ...any) { errores = append(errores, c.ID+": "+fmt.Sprintf(f, a...)) }
		w := func(f string, a ...any) { avisos = append(avisos, c.ID+": "+fmt.Sprintf(f, a...)) }
		if c.ID == "" || vistos[c.ID] {
			e("id vacío o repetido")
		}
		vistos[c.ID] = true
		if !cats[c.Categoria] {
			e("categoría desconocida %q", c.Categoria)
		}
		if !tipos[c.TipoEsperado] {
			e("tipo_esperado desconocido %q", c.TipoEsperado)
		}
		if len(c.Turnos) == 0 {
			e("sin turnos")
		}
		if !c.Sintetico {
			w("sintetico no es true")
		}
		if p := c.Proc(); p != "" {
			proc, ok := b.Procedimientos[p]
			if !ok {
				e("procedimiento_esperado %q no existe en kb/procedimientos", p)
			} else {
				validos := map[int]bool{}
				for _, n := range pasosNivel1(proc) {
					validos[n] = true
				}
				for _, n := range c.PasosEsperados {
					if !validos[n] {
						e("paso %d no existe en %s", n, p)
					}
				}
				fpp := fotosPorPaso(proc)
				for ns, fotos := range c.FotosEsperadas {
					n, _ := strconv.Atoi(ns)
					for _, f := range fotos {
						if !contiene(fpp[n], f) {
							e("la foto %s no pertenece al paso %d de %s", f, n, p)
						}
					}
				}
			}
		} else if len(c.FotosEsperadas) > 0 || c.PasosEsperados != nil {
			e("pasos o fotos esperadas sin procedimiento_esperado")
		}
		if k := c.Concepto(); k != "" && b.BuscarConcepto(k) == nil {
			w("concepto_esperado %q aún no está en kb/conceptos (se califica solo por fragmentos)", k)
		}
		for _, r := range c.FragmentosRelevantes {
			if len(b.Resolver(r.Clave())) == 0 {
				e("fragmento relevante %q no existe en kb/fragmentos*.jsonl", r.Clave())
			} else if r.Manual == "" && len(b.Resolver(r.ID)) > 1 {
				e("fragmento %q es ambiguo (%d manuales): falta «manual»", r.ID, len(b.Resolver(r.ID)))
			}
		}
		if c.SinEvidenciaEsperada && (c.Proc() != "" || len(c.FragmentosRelevantes) > 0) {
			e("sin_evidencia_esperada con procedimiento o fragmentos relevantes")
		}
		if !c.SinEvidenciaEsperada && c.TipoEsperado != "UNKNOWN" && c.TipoEsperado != "SOCIAL" &&
			c.Proc() == "" && len(c.FragmentosRelevantes) == 0 {
			e("caso con respuesta esperada pero sin procedimiento ni fragmentos relevantes")
		}
	}
	sort.Strings(errores)
	sort.Strings(avisos)
	return errores, avisos
}

func contiene[T comparable](xs []T, x T) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// muestra toma N casos alternando categorías en el orden del archivo (determinista, conserva la mezcla).
func muestra(casos []Caso, n int) []Caso {
	if n <= 0 || n >= len(casos) {
		return casos
	}
	porCat := map[string][]Caso{}
	var orden []string
	for _, c := range casos {
		if _, ok := porCat[c.Categoria]; !ok {
			orden = append(orden, c.Categoria)
		}
		porCat[c.Categoria] = append(porCat[c.Categoria], c)
	}
	var out []Caso
	for i := 0; len(out) < n; i++ {
		agrego := false
		for _, cat := range orden {
			if i < len(porCat[cat]) && len(out) < n {
				out = append(out, porCat[cat][i])
				agrego = true
			}
		}
		if !agrego {
			break
		}
	}
	return out
}
