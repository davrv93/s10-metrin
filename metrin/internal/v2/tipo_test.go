package v2

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rag-go/internal/clasificar"
	"rag-go/internal/embed"
	"rag-go/internal/v2/tipos"
)

func TestTipoDeRespuestaEjemplosDelMegaprompt(t *testing.T) {
	var c *Clasificador // sin modelos: solo reglas
	for pregunta, quiero := range map[string]tipos.TipoRespuesta{
		"¿Qué es un metrado?":                      tipos.Concepto,
		"¿Cómo registro un metrado?":               tipos.Procedimiento,
		"¿Dónde encuentro los metrados?":           tipos.Navegacion,
		"No puedo guardar el metrado":              tipos.Problema,
		"¿Cómo configuro las unidades?":            tipos.Configuracion,
		"¿Qué diferencia hay entre APU y partida?": tipos.Comparacion,
		"q es un insumo":                           tipos.Concepto,
		"me sale error al configurar la moneda":    tipos.Problema,
		"hola":                                     tipos.Social,
		"¿cómo estás?":                             tipos.Social,
		"metrados":                                 tipos.Desconocido,
	} {
		r := c.Clasificar(context.Background(), pregunta)
		if r.Tipo != quiero {
			t.Errorf("%q → %s (%s, %.2f), quería %s", pregunta, r.Tipo, r.Metodo, r.Confianza, quiero)
		}
	}
}

func TestTipoConflictoVaAlMotorDeDecision(t *testing.T) {
	var c *Clasificador
	r := c.Clasificar(context.Background(), "¿Qué es un metrado y cómo lo registro?")
	if !r.Dudoso() || len(r.Candidatos) != 2 || r.Tipo != tipos.Concepto {
		t.Fatalf("conflicto concepto/procedimiento: %+v", r)
	}
	r = c.Clasificar(context.Background(), "metrados")
	if r.Tipo != tipos.Desconocido || !r.Dudoso() {
		t.Fatalf("sin señal: %+v", r)
	}
	r = c.Clasificar(context.Background(), "¿Cómo registro un metrado?")
	if r.Dudoso() || r.Metodo != "reglas" {
		t.Fatalf("regla clara: %+v", r)
	}
}

// emb3 es un embebedor de tres ejes (concepto, procedimiento, navegación) para probar el kNN de tipo.
type emb3 struct{}

func (emb3) Nombre() string { return "tres-ejes" }
func (emb3) Embeber(_ context.Context, s string) ([]float32, error) {
	switch {
	case strings.Contains(s, "significa") || strings.Contains(s, "definicion"):
		return []float32{1, 0, 0}, nil
	case strings.Contains(s, "pasos"):
		return []float32{0, 1, 0}, nil
	}
	return []float32{0, 0, 1}, nil
}

func TestTipoKNNDecideSinRegla(t *testing.T) {
	m, err := clasificar.Entrenar(context.Background(), emb3{}, "tres-ejes", 0.5, map[string][]string{
		"CONCEPT":   {"definicion de partida"},
		"PROCEDURE": {"pasos para emitir"},
	})
	if err != nil {
		t.Fatal(err)
	}
	c := &Clasificador{Emb: emb3{}, Tipo: m, Umbral: 0.5, Margen: 0.05}
	r := c.Clasificar(context.Background(), "kardex significa algo")
	if r.Tipo != tipos.Concepto || r.Metodo != "knn" || r.Confianza < 0.99 {
		t.Fatalf("el kNN decide cuando no hay regla: %+v", r)
	}
	// Una regla clara no la pisa el kNN.
	r = c.Clasificar(context.Background(), "¿Cómo registro un metrado? significa")
	if r.Tipo != tipos.Procedimiento || r.Metodo != "reglas" {
		t.Fatalf("regla clara: %+v", r)
	}
}

func TestEjemplosDelCatalogoDeTipo(t *testing.T) {
	ruta := filepath.Join("..", "..", "..", "kb", "catalogos", "tipo_respuesta.yml")
	if _, err := os.Stat(ruta); err != nil {
		t.Skip("sin catálogo kb/catalogos/tipo_respuesta.yml")
	}
	ej, err := EjemplosCatalogoTipo(ruta)
	if err != nil {
		t.Fatal(err)
	}
	for _, clase := range []string{"CONCEPT", "PROCEDURE", "NAVIGATION", "TROUBLESHOOTING", "CONFIGURATION", "COMPARISON", "UNKNOWN"} {
		if len(ej[clase]) < 20 {
			t.Errorf("%s: %d ejemplos", clase, len(ej[clase]))
		}
		for _, x := range ej[clase] {
			if strings.HasPrefix(x, "\"") || strings.HasPrefix(x, "- ") || x == "" {
				t.Fatalf("%s: ejemplo mal leído %q", clase, x)
			}
		}
	}
	if _, ok := ej["clases"]; ok {
		t.Fatal("solo clases del contrato")
	}
}

// Medición (informativa) de reglas + kNN con margen sobre el conjunto de PRUEBA del catálogo (nunca se entrena con
// él), con el modelo candidato y el embebedor estático reales si están en la máquina. No falla por la cifra: deja
// el dato para calibrar V2_TIPO_UMBRAL y V2_TIPO_MARGEN.
func TestMedicionTipoConCatalogoReal(t *testing.T) {
	modelo := filepath.Join("..", "..", "modelos", "potion-es-int8.pjge")
	tipoJSON := filepath.Join("..", "..", "modelos", "clasificador-tipo-respuesta-candidato.json")
	prueba := filepath.Join("..", "..", "..", "kb", "catalogos", "tipo_respuesta_prueba.yml")
	for _, r := range []string{modelo, prueba, tipoJSON} {
		if _, err := os.Stat(r); err != nil {
			t.Skip("falta " + r)
		}
	}
	emb, err := embed.NuevoEstatico(modelo)
	if err != nil {
		t.Skip(err)
	}
	m, err := CargarModeloTipo(context.Background(), tipoJSON, emb, UmbralTipoDefecto)
	if err != nil {
		t.Fatal(err)
	}
	casos, err := EjemplosCatalogoTipo(prueba)
	if err != nil {
		t.Skip(err)
	}
	medir := func(umbral, margen float64, conKNN bool) {
		c := &Clasificador{Emb: emb, Umbral: umbral, Margen: margen}
		if conKNN {
			c.Tipo = m
		}
		total, decididos, aciertosDec, reglas, knn := 0, 0, 0, 0, 0
		for clase, xs := range casos {
			for _, x := range xs {
				r := c.Clasificar(context.Background(), x)
				total++
				if r.Dudoso() {
					continue
				}
				decididos++
				if r.Metodo == "knn" {
					knn++
				} else {
					reglas++
				}
				if string(r.Tipo) == clase {
					aciertosDec++
				}
			}
		}
		t.Logf("umbral %.2f margen %.2f kNN=%v: decide %d/%d (%.1f %%; reglas %d, kNN %d) con precisión %.1f %%; el resto va al motor",
			umbral, margen, conKNN, decididos, total, 100*float64(decididos)/float64(max(total, 1)), reglas, knn,
			100*float64(aciertosDec)/float64(max(decididos, 1)))
	}
	medir(UmbralTipoDefecto, MargenTipoDefecto, false)
	for _, u := range []float64{0.40, 0.45, 0.50, 0.59} {
		for _, mg := range []float64{0.02, 0.05, 0.08} {
			medir(u, mg, true)
		}
	}
}
