// Arnés del clasificador de Metrín: entrena un modelo CANDIDATO y puntúa frases
// con el MISMO código que usa el servicio (rag-go/internal/clasificar y el
// embebedor estático rag-go/internal/embed). No modifica nada de metrin/.
//
// Por qué existe: `rag entrenar-clasificador` (cmd/rag/cmd_clasificar.go) solo
// acepta las 8 categorías de su tabla `intencionDe` y las pliega en 4 clases
// (social, ayuda, trabajo, limite). Para medir una taxonomía nueva hace falta
// llamar a clasificar.Entrenar con las etiquetas tal cual; eso hace este arnés.
//
// Uso (desde esta carpeta; el go.work enlaza ../../metrin):
//
//	go run . entrenar --ejemplos ejemplos.json --salida candidato.json [--umbral 0.40]
//	go run . puntuar  --frases frases.json [--clasificador candidato.json] > puntajes.jsonl
//	go run . embeber  --frases frases.json > vectores.jsonl
//
// ejemplos.json = {"intencion": ["texto", ...], ...}
// frases.json   = ["texto", ...]
// Sin --clasificador, puntuar usa el modelo embarcado (defecto.json, el ACTUAL).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rag-go/internal/clasificar"
	"rag-go/internal/embed"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: clasificadormetrin entrenar|puntuar [opciones]")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "entrenar":
		err = entrenar(os.Args[2:])
	case "puntuar":
		err = puntuar(os.Args[2:])
	case "embeber":
		err = embeber(os.Args[2:])
	default:
		err = fmt.Errorf("subcomando desconocido: %s", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func modeloEmbDefecto() string {
	return filepath.Join("..", "..", "metrin", "modelos", "potion-es-int8.pjge")
}

func entrenar(args []string) error {
	fs := flag.NewFlagSet("entrenar", flag.ContinueOnError)
	ejemplos := fs.String("ejemplos", "", "JSON {intención: [textos]}")
	salida := fs.String("salida", "", "ruta del modelo candidato")
	umbral := fs.Float64("umbral", clasificar.UmbralDefecto, "similitud mínima")
	modeloEmb := fs.String("emb", modeloEmbDefecto(), "modelo estático .pjge")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *ejemplos == "" || *salida == "" {
		return fmt.Errorf("faltan --ejemplos o --salida")
	}
	b, err := os.ReadFile(*ejemplos)
	if err != nil {
		return err
	}
	var por map[string][]string
	if err := json.Unmarshal(b, &por); err != nil {
		return err
	}
	e, err := embed.NuevoEstatico(*modeloEmb)
	if err != nil {
		return err
	}
	m, err := clasificar.Entrenar(context.Background(), e, e.Nombre(), *umbral, por)
	if err != nil {
		return err
	}
	if err := m.Guardar(*salida); err != nil {
		return err
	}
	total := 0
	for _, in := range m.Intenciones {
		total += in.Ejemplos
	}
	fmt.Fprintf(os.Stderr, "modelo %s: %d intenciones, %d ejemplos, umbral %.3f → %s\n",
		m.HuellaEmb, len(m.Intenciones), total, m.Umbral, *salida)
	return nil
}

type candidata struct {
	Intencion string  `json:"intencion"`
	Similitud float64 `json:"similitud"`
}

type salidaFrase struct {
	Texto      string      `json:"texto"`
	Intencion  string      `json:"intencion"` // lo que devuelve Clasificar (umbral + capa social incluidos)
	Similitud  float64     `json:"similitud"`
	Social     float64     `json:"puntaje_social"`
	Candidatas []candidata `json:"candidatas"` // Puntajes: mejor similitud por intención (kNN k=1)
	Error      string      `json:"error,omitempty"`
}

func puntuar(args []string) error {
	fs := flag.NewFlagSet("puntuar", flag.ContinueOnError)
	frases := fs.String("frases", "", "JSON [\"texto\", ...]")
	ruta := fs.String("clasificador", "", "modelo (vacío = embarcado, defecto.json)")
	modeloEmb := fs.String("emb", modeloEmbDefecto(), "modelo estático .pjge")
	if err := fs.Parse(args); err != nil {
		return err
	}
	b, err := os.ReadFile(*frases)
	if err != nil {
		return err
	}
	var textos []string
	if err := json.Unmarshal(b, &textos); err != nil {
		return err
	}
	var m *clasificar.Modelo
	if *ruta == "" {
		m, err = clasificar.CargarDefecto()
	} else {
		m, err = clasificar.Cargar(*ruta)
	}
	if err != nil {
		return err
	}
	e, err := embed.NuevoEstatico(*modeloEmb)
	if err != nil {
		return err
	}
	if m.HuellaEmb != e.Nombre() {
		return fmt.Errorf("el clasificador se entrenó con %s y el embebedor es %s", m.HuellaEmb, e.Nombre())
	}
	ctx := context.Background()
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	for _, t := range textos {
		s := salidaFrase{Texto: t, Social: clasificar.PuntajeSocial(t)}
		in, sim, err := m.Clasificar(ctx, e, t)
		if err != nil {
			s.Error = err.Error()
		} else {
			s.Intencion, s.Similitud = in, sim
		}
		ins, sims, err := m.Puntajes(ctx, e, t)
		if err == nil {
			for i, ci := range ins {
				s.Candidatas = append(s.Candidatas, candidata{ci.Nombre, sims[i]})
			}
		}
		if err := enc.Encode(s); err != nil {
			return err
		}
	}
	return nil
}

// normalizarTexto replica clasificar.normalizarTexto (no exportada): minúsculas,
// tildes plegadas y puntuación externa fuera. Es lo que ve el embebedor al
// entrenar y al clasificar.
var plegar = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u")

func normalizarTexto(s string) string {
	s = plegar.Replace(strings.ToLower(strings.TrimSpace(s)))
	return strings.Trim(s, " \t?!.,;:")
}

// embeber escribe el vector normalizado de cada frase (tras normalizarTexto),
// para medir el mismo kNN fuera de Go con otros embebedores.
func embeber(args []string) error {
	fs := flag.NewFlagSet("embeber", flag.ContinueOnError)
	frases := fs.String("frases", "", "JSON [\"texto\", ...]")
	modeloEmb := fs.String("emb", modeloEmbDefecto(), "modelo estático .pjge")
	if err := fs.Parse(args); err != nil {
		return err
	}
	b, err := os.ReadFile(*frases)
	if err != nil {
		return err
	}
	var textos []string
	if err := json.Unmarshal(b, &textos); err != nil {
		return err
	}
	e, err := embed.NuevoEstatico(*modeloEmb)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	for _, t := range textos {
		fila := map[string]any{"texto": t}
		v, err := e.Embeber(context.Background(), normalizarTexto(t))
		if err != nil {
			fila["error"] = err.Error()
		} else {
			fila["vec"] = embed.Normalizar(v)
		}
		if err := enc.Encode(fila); err != nil {
			return err
		}
	}
	return nil
}
