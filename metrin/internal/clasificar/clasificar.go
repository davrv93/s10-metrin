// Package clasificar asigna la intención de un mensaje corto de chat.
//
// Es un clasificador por centroides sobre los embeddings estáticos en español
// que ya usa el RAG (sin red ni GPU): cada intención promedia sus ejemplos y
// el mensaje nuevo hereda la intención del centroide más cercano por coseno.
// Si ninguna supera el umbral, responde "trabajo" (ruta segura: el RAG con
// fuentes decide, nunca el clasificador).
package clasificar

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"rag-go/internal/embed"
)

// Intenciones que el clasificador puede devolver. "trabajo" es la ruta por
// defecto: pregunta de obra o S10 que debe resolver el RAG con fuentes.
const (
	Trabajo     = "trabajo"
	// UmbralDefecto calibrado en valid (22/28) y test (23/28): por debajo
	// todo va a trabajo (ruta segura RAG). Cero desvíos inseguros medidos.
	UmbralDefecto = 0.40
)

//go:embed defecto.json
var defecto []byte

// Intencion es un centroide (resumen) con su nombre.
type Intencion struct {
	Nombre     string    `json:"nombre"`
	Centroide  []float32 `json:"centroide"`
	Dim        int       `json:"dim"`
	Ejemplos   int       `json:"ejemplos"`
}

// Ejemplo es un texto de entrenamiento con su vector: el clasificador decide
// por vecino más cercano (kNN k=1). Memoriza cortesías exactas ("jaja") y
// generaliza por cercanía de embeddings; el umbral manda lo dudoso a trabajo.
type Ejemplo struct {
	Texto     string    `json:"texto"`
	Intencion string    `json:"intencion"`
	Vec       []float32 `json:"vec"`
}

// Modelo es el conjunto ordenado de centroides más los ejemplos (kNN).
type Modelo struct {
	HuellaEmb   string      `json:"huella_emb"`
	Umbral      float64     `json:"umbral"`
	Intenciones []Intencion `json:"intenciones"`
	Ejemplos    []Ejemplo   `json:"ejemplos"`
}

// CargarDefecto devuelve el modelo embarcado en el binario (funciona en Docker
// sin volúmenes ni reentrenamiento).
func CargarDefecto() (*Modelo, error) {
	return decodificar(defecto)
}

// Cargar lee un modelo guardado con Guardar.
func Cargar(ruta string) (*Modelo, error) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	return decodificar(b)
}

func decodificar(b []byte) (*Modelo, error) {
	var m Modelo
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if len(m.Intenciones) == 0 {
		return nil, fmt.Errorf("clasificar: modelo sin intenciones")
	}
	return &m, nil
}

// Guardar persiste el modelo en JSON.
func (m *Modelo) Guardar(ruta string) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ruta, append(b, '\n'), 0o644)
}

// normalizarTexto baja a minúsculas, pliega tildes y quita puntuación externa:
// los embeddings estáticos distinguen "Hola" de "hola" y "cómo" de "como".
// La ñ se conserva (año != ano). Se aplica al entrenar y al clasificar.
var plegar = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u",
)

func normalizarTexto(s string) string {
	s = plegar.Replace(strings.ToLower(strings.TrimSpace(s)))
	return strings.Trim(s, " \t?!.,;:")
}

// Entrenar promedia los embeddings normalizados de cada intención.
// ejemplos mapea intención -> textos; ignora intenciones vacías.
func Entrenar(ctx context.Context, e embed.Embebedor, huella string, umbral float64, ejemplos map[string][]string) (*Modelo, error) {
	nombres := make([]string, 0, len(ejemplos))
	for n, ts := range ejemplos {
		if len(ts) > 0 {
			nombres = append(nombres, n)
		}
	}
	sort.Strings(nombres)
	m := &Modelo{HuellaEmb: huella, Umbral: umbral}
	for _, n := range nombres {
		var suma []float32
		nv := 0
		for _, t := range ejemplos[n] {
			v, err := e.Embeber(ctx, normalizarTexto(t))
			if err != nil {
				return nil, fmt.Errorf("clasificar %s: %w", n, err)
			}
			v = embed.Normalizar(v)
			if suma == nil {
				suma = make([]float32, len(v))
			}
			for i, x := range v {
				suma[i] += x
			}
			nv++
		}
		c := embed.Normalizar(suma)
		m.Intenciones = append(m.Intenciones, Intencion{
			Nombre: n, Centroide: c, Dim: len(c), Ejemplos: nv,
		})
	}
	// Guarda cada ejemplo normalizado para kNN.
	for _, n := range nombres {
		for _, t := range ejemplos[n] {
			v, err := e.Embeber(ctx, normalizarTexto(t))
			if err != nil {
				return nil, fmt.Errorf("clasificar %s: %w", n, err)
			}
			m.Ejemplos = append(m.Ejemplos, Ejemplo{Texto: t, Intencion: n, Vec: embed.Normalizar(v)})
		}
	}
	if len(m.Intenciones) == 0 {
		return nil, fmt.Errorf("clasificar: sin ejemplos")
	}
	return m, nil
}

// Puntajes devuelve la similitud coseno con cada centroide, ordenada de mayor
// a menor. Sirve para calibrar el umbral fuera de línea.
func (m *Modelo) Puntajes(ctx context.Context, e embed.Embebedor, texto string) ([]Intencion, []float64, error) {
	v, err := e.Embeber(ctx, normalizarTexto(texto))
	if err != nil {
		return nil, nil, err
	}
	v = embed.Normalizar(v)
	// Con ejemplos: mejor similitud por intención (kNN k=1). Sin ellos
	// (modelos viejos): centroides.
	mejorPor := map[string]float64{}
	if len(m.Ejemplos) > 0 {
		for _, ej := range m.Ejemplos {
			if len(ej.Vec) != len(v) {
				continue
			}
			if s := cos(v, ej.Vec); s > mejorPor[ej.Intencion] {
				mejorPor[ej.Intencion] = s
			}
		}
	} else {
		for _, in := range m.Intenciones {
			if len(in.Centroide) == len(v) {
				mejorPor[in.Nombre] = cos(v, in.Centroide)
			}
		}
	}
	nombres := make([]string, 0, len(mejorPor))
	for n := range mejorPor {
		nombres = append(nombres, n)
	}
	sort.Strings(nombres)
	type par struct {
		in  Intencion
		sim float64
	}
	pares := make([]par, 0, len(nombres))
	for _, n := range nombres {
		pares = append(pares, par{Intencion{Nombre: n}, mejorPor[n]})
	}
	sort.Slice(pares, func(a, b int) bool { return pares[a].sim > pares[b].sim })
	ins := make([]Intencion, 0, len(pares))
	sims := make([]float64, 0, len(pares))
	for _, p := range pares {
		ins = append(ins, p.in)
		sims = append(sims, p.sim)
	}
	return ins, sims, nil
}

func cos(a, b []float32) float64 {
	var s float64
	for i, x := range a {
		s += float64(x) * float64(b[i])
	}
	return s
}

// Clasificar devuelve la intención ganadora y su similitud coseno.
// Por debajo del umbral (o empate con trabajo) responde trabajo.
func (m *Modelo) Clasificar(ctx context.Context, e embed.Embebedor, texto string) (string, float64, error) {
	ins, sims, err := m.Puntajes(ctx, e, texto)
	// Charla pura («bien y tu», «ok», «jajaja»): si el embedding la iba a
	// mandar a trabajo (puntaje bajo en mensajes cortos) o no la entiende,
	// gana la capa léxica. Lo que el embedding ya reconoce (ayuda, límite)
	// se respeta.
	charla := PuntajeSocial(texto) == 1
	if err != nil {
		if charla {
			return Social, 1, nil
		}
		return "", 0, err
	}
	if len(ins) == 0 {
		if charla {
			return Social, 1, nil
		}
		return Trabajo, math.Inf(-1), nil
	}
	mejor, mejorSim := ins[0].Nombre, sims[0]
	if mejorSim < m.Umbral || mejor == Trabajo {
		if charla {
			return Social, 1, nil
		}
		return Trabajo, mejorSim, nil
	}
	return mejor, mejorSim, nil
}
