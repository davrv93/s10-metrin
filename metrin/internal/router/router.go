// Package router elige el procedimiento (el «caso del tutorial») que responde una pregunta, con el árbol de decisión
// de docs/TOC-RUTEO-METRIN.md §5: módulo primero, abstenerse antes que equivocarse y, si hay empate, preguntar.
//
// El modelo lo entrena entrenamiento-router/entrenar.py y es exactamente esto, sin dependencias:
//
//	v = embed.Modelo.Embeber(pregunta)     // embeddings estáticos afinados (router-s10.pjge), media y norma L2
//	p = softmax(W·v + b)                   // una clase por procedimiento + «_ninguno»
//
// La cabeza (W, b, clases, módulo de cada clase y umbrales calibrados) va en router-s10.cabeza.json, que guarda la
// huella sha256 del .pjge con el que se entrenó: si no coincide, no carga.
package router

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"rag-go/internal/embed"
)

// Acciones del árbol.
const (
	Elegir        = "elegir"         // un procedimiento, con confianza y margen suficientes
	AclararCaso   = "aclarar_caso"   // 2–3 procedimientos del mismo módulo empatan: se pregunta cuál
	AclararModulo = "aclarar_modulo" // ni el módulo está claro: se pregunta el módulo
	Delegar       = "delegar"        // gana «ninguno» o nada llega al mínimo: decide la V2 como siempre
)

// Umbrales del árbol (calibrados por entrenar.py en el conjunto de calibración, nunca en la prueba).
type Umbrales struct {
	Minimo float64 `json:"minimo"` // probabilidad mínima del mejor procedimiento para no delegar
	Modulo float64 `json:"modulo"` // probabilidad mínima del módulo (suma de sus procedimientos)
	Acepta float64 `json:"acepta"` // probabilidad mínima para elegir
	Margen float64 `json:"margen"` // diferencia mínima con el segundo del mismo módulo para elegir
}

type cabeza struct {
	Version    int         `json:"version"`
	Pjge       string      `json:"pjge"`
	Sha256Pjge string      `json:"sha256_pjge"`
	Clases     []string    `json:"clases"`
	Modulo     []string    `json:"modulo"`
	Ninguno    string      `json:"ninguno"`
	Umbrales   Umbrales    `json:"umbrales"`
	W          [][]float32 `json:"W"`
	B          []float32   `json:"b"`
}

// Router es de solo lectura tras Cargar: se puede usar desde varias goroutines.
type Router struct {
	emb     *embed.Modelo
	clases  []string
	modulo  []string
	ninguno int
	w       [][]float32
	b       []float32
	U       Umbrales
	Huella  string // sha256 del .pjge (12 primeros caracteres en la traza)
}

// Candidato: un procedimiento (o módulo) con su probabilidad.
type Candidato struct {
	ID           string  `json:"id"`
	Probabilidad float64 `json:"p"`
}

// Decision: lo que dice el árbol para una pregunta.
type Decision struct {
	Accion     string      `json:"accion"`
	Candidatos []Candidato `json:"candidatos,omitempty"` // procedimientos (elegir, aclarar_caso) o módulos (aclarar_modulo)
	Ninguno    float64     `json:"p_ninguno"`
	Modulo     string      `json:"modulo,omitempty"`
	PModulo    float64     `json:"p_modulo,omitempty"`
	Top        []Candidato `json:"top"` // los 3 procedimientos más probables, para la traza
}

// Cargar lee la cabeza (JSON) y el .pjge que nombra, en la misma carpeta.
func Cargar(rutaCabeza string) (*Router, error) {
	raw, err := os.ReadFile(rutaCabeza)
	if err != nil {
		return nil, err
	}
	var c cabeza
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("router: cabeza %s: %w", rutaCabeza, err)
	}
	if c.Version != 1 {
		return nil, fmt.Errorf("router: versión de cabeza %d no soportada", c.Version)
	}
	rutaPjge := filepath.Join(filepath.Dir(rutaCabeza), c.Pjge)
	datos, err := os.ReadFile(rutaPjge)
	if err != nil {
		return nil, err
	}
	suma := sha256.Sum256(datos)
	huella := hex.EncodeToString(suma[:])
	if huella != c.Sha256Pjge {
		return nil, fmt.Errorf("router: %s no es el .pjge con el que se entrenó la cabeza (sha256 %s, esperado %s)",
			c.Pjge, huella[:12], short(c.Sha256Pjge))
	}
	m, err := embed.CargarArchivo(rutaPjge)
	if err != nil {
		return nil, fmt.Errorf("router: %w", err)
	}
	n := len(c.Clases)
	if n < 2 || len(c.Modulo) != n || len(c.W) != n || len(c.B) != n {
		return nil, fmt.Errorf("router: cabeza incoherente (%d clases, %d módulos, %d filas W, %d b)", n, len(c.Modulo), len(c.W), len(c.B))
	}
	for i, fila := range c.W {
		if len(fila) != m.Dim() {
			return nil, fmt.Errorf("router: fila %d de W con %d columnas; el embedding tiene %d", i, len(fila), m.Dim())
		}
	}
	r := &Router{emb: m, clases: c.Clases, modulo: c.Modulo, ninguno: -1, w: c.W, b: c.B, U: c.Umbrales, Huella: huella}
	for i, cl := range c.Clases {
		if cl == c.Ninguno {
			r.ninguno = i
		}
	}
	if r.ninguno < 0 {
		return nil, fmt.Errorf("router: la clase %q no está entre las clases", c.Ninguno)
	}
	return r, nil
}

func short(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

// Clases devuelve los procedimientos que conoce (sin «ninguno»).
func (r *Router) Clases() []string {
	out := make([]string, 0, len(r.clases)-1)
	for i, c := range r.clases {
		if i != r.ninguno {
			out = append(out, c)
		}
	}
	return out
}

// Probabilidades: softmax(W·v + b), en el orden de las clases de la cabeza.
func (r *Router) Probabilidades(texto string) []float64 {
	v := r.emb.Embeber(texto)
	z := make([]float64, len(r.clases))
	maxz := math.Inf(-1)
	for i, fila := range r.w {
		s := float64(r.b[i])
		for j, x := range fila {
			s += float64(x) * float64(v[j])
		}
		z[i] = s
		if s > maxz {
			maxz = s
		}
	}
	var tot float64
	for i := range z {
		z[i] = math.Exp(z[i] - maxz)
		tot += z[i]
	}
	for i := range z {
		z[i] /= tot
	}
	return z
}

// Decidir aplica el árbol (la misma función que decidir() de entrenar.py).
func (r *Router) Decidir(texto string) Decision {
	p := r.Probabilidades(texto)
	orden := make([]int, 0, len(p)-1)
	for i := range p {
		if i != r.ninguno {
			orden = append(orden, i)
		}
	}
	sort.SliceStable(orden, func(a, b int) bool { return p[orden[a]] > p[orden[b]] })
	d := Decision{Ninguno: p[r.ninguno]}
	for _, i := range orden[:min(3, len(orden))] {
		d.Top = append(d.Top, Candidato{ID: r.clases[i], Probabilidad: p[i]})
	}
	if p[r.ninguno] >= p[orden[0]] || p[orden[0]] < r.U.Minimo {
		d.Accion = Delegar
		return d
	}
	pm := map[string]float64{}
	var mods []string
	for _, i := range orden {
		m := r.modulo[i]
		if _, ok := pm[m]; !ok {
			mods = append(mods, m)
		}
		pm[m] += p[i]
	}
	sort.SliceStable(mods, func(a, b int) bool { return pm[mods[a]] > pm[mods[b]] })
	d.Modulo, d.PModulo = mods[0], pm[mods[0]]
	if pm[mods[0]] < r.U.Modulo {
		d.Accion = AclararModulo
		for _, m := range mods[:min(3, len(mods))] {
			d.Candidatos = append(d.Candidatos, Candidato{ID: m, Probabilidad: pm[m]})
		}
		return d
	}
	var dentro []int
	for _, i := range orden {
		if r.modulo[i] == mods[0] {
			dentro = append(dentro, i)
		}
	}
	p1, p2 := p[dentro[0]], 0.0
	if len(dentro) > 1 {
		p2 = p[dentro[1]]
	}
	if p1 >= r.U.Acepta && p1-p2 >= r.U.Margen {
		d.Accion = Elegir
		d.Candidatos = []Candidato{{ID: r.clases[dentro[0]], Probabilidad: p1}}
		return d
	}
	d.Accion = AclararCaso
	for _, i := range dentro[:min(3, len(dentro))] {
		d.Candidatos = append(d.Candidatos, Candidato{ID: r.clases[i], Probabilidad: p[i]})
	}
	return d
}
