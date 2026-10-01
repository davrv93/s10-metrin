package rag

import (
	"context"
	"math"
	"strings"

	"rag-go/internal/clasificar"
)

const (
	rutaConversacion = "conversacion"
	rutaRAG          = "rag"
	rutaRegla        = "regla"
)

// Orquestacion hace visible la decisión del router. SimilitudClasificador es
// la similitud coseno del modelo local, no una probabilidad calibrada.
type Orquestacion struct {
	Intencion             string  `json:"intencion"`
	TipoConsulta          string  `json:"tipo_consulta"`
	Ruta                  string  `json:"ruta"`
	Clasificador          string  `json:"clasificador"`
	SimilitudClasificador float64 `json:"similitud_clasificador,omitempty"`
}

// planificar aplica primero el clasificador español local a la decisión de
// ruta y después reglas explícitas para el tipo de respuesta técnica. Lo técnico
// siempre pasa por retrieval: el clasificador no decide hechos ni fuentes.
func (r *RAG) planificar(ctx context.Context, pregunta string, hilo []Turno) Orquestacion {
	p := Orquestacion{
		Intencion:    clasificar.Trabajo,
		TipoConsulta: tipoConsulta(pregunta, hilo),
		Ruta:         rutaRAG,
		Clasificador: "reglas_seguras",
	}
	if r.Clasificador == nil || r.Emb == nil {
		return p
	}

	p.Clasificador = "kNN-local"
	if r.Clasificador.HuellaEmb != "" {
		p.Clasificador += ":" + r.Clasificador.HuellaEmb
	}
	intencion, similitud, err := r.Clasificador.Clasificar(ctx, r.Emb, pregunta)
	if err != nil {
		p.Clasificador = "error_fallback_rag"
		return p
	}
	p.Intencion = intencion
	if !math.IsInf(similitud, 0) && !math.IsNaN(similitud) {
		p.SimilitudClasificador = similitud
	}

	// Excepción de seguridad: una pregunta con contenido de trabajo no se
	// desvía por una etiqueta conversacional accidental.
	if !esPreguntaLarga(pregunta) && (intencion == clasificar.Social || intencion == "limite") {
		p.Ruta = rutaConversacion
		if intencion == clasificar.Social {
			p.TipoConsulta = "social"
		} else {
			p.TipoConsulta = "fuera_de_alcance"
		}
	}
	return p
}

func tipoConsulta(pregunta string, hilo []Turno) string {
	t := " " + strings.TrimSpace(normalizarConsulta(pregunta)) + " "
	if len(hilo) > 0 && esSeguimiento(t) {
		return "seguimiento"
	}
	if contieneAlguna(t, "error", "falla", "fallo", "problema", "no funciona", "no aparece", "no permite", "se bloquea", "no carga") {
		return "problema"
	}
	if contieneAlguna(t, "diferencia", "comparar", "compara", "comparacion", "versus", " frente a ", "cual conviene", "cual es mejor") {
		return "comparacion"
	}
	if esHowTo(t) {
		return "procedimiento"
	}
	if contieneAlguna(t, "que es", "que significa", "para que sirve", "en que consiste", "define", "definicion", "concepto de") {
		return "concepto"
	}
	if contieneAlguna(t, "no entiendo", "no me queda claro", "explicame de otra forma", "reformula", "mas sencillo", "mas simple") {
		return "aclaracion"
	}
	return "informacion_directa"
}

func esSeguimiento(pregunta string) bool {
	t := " " + strings.TrimSpace(pregunta) + " "
	return contieneAlguna(t, " eso ", " esto ", " aquello ", " ese ", " esa ", " lo anterior", " lo que dijiste", " despues", " siguiente", " ahi", " aqui", " su ", " continua", " continua desde")
}

func normalizarConsulta(s string) string {
	s = strings.ToLower(strings.Join(strings.Fields(s), " "))
	return strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u",
		"¿", " ", "?", " ", "¡", " ", "!", " ", ",", " ", ".", " ", ":", " ", ";", " ",
	).Replace(s)
}

// consultaCompacta quita conectores de la pregunta para una segunda búsqueda
// de manuales. La búsqueda general conserva la frase original; esta variante
// evita que la forma interrogativa diluya términos exactos de S10.
func consultaCompacta(pregunta string) string {
	parada := map[string]bool{}
	for _, palabra := range strings.Fields("a al algo como con cual cuales cuando de del desde donde el ella en es esa ese eso esta este hay la las lo los me mi o para pero por que se si sin su sus un una unos unas y yo manual manuales modulo segun significa") {
		parada[palabra] = true
	}
	var claves []string
	for _, palabra := range strings.Fields(normalizarConsulta(pregunta)) {
		if len([]rune(palabra)) >= 3 && !parada[palabra] {
			claves = append(claves, palabra)
		}
	}
	if len(claves) == 0 {
		return strings.TrimSpace(pregunta)
	}
	return strings.Join(claves, " ")
}

func contieneAlguna(texto string, frases ...string) bool {
	for _, frase := range frases {
		if strings.Contains(texto, frase) {
			return true
		}
	}
	return false
}
