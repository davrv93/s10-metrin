// Comandos del clasificador de intenciones (Go puro, embeddings locales).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"rag-go/internal/clasificar"
	"rag-go/internal/config"
	"rag-go/internal/embed"
)

// categoria del dataset -> intención del clasificador. Las peticiones de ayuda
// (aclarar, ambigua, seguimiento, corrección) comparten centroide "ayuda":
// con 8-17 ejemplos por clase fina los centroides se solapan y la precisión
// cae; 4 clases gruesas sí separan (ver eval en docs).
var intencionDe = map[string]string{
	"saludos":       "social",
	"gracias":       "social",
	"explicaciones": "ayuda",
	"ambiguas":      "ayuda",
	"sin_evidencia": "trabajo",
	"limites":       "limite",
	"correctivo":    "ayuda",
	"seguimiento":   "ayuda",
}

func cmdEntrenarClasificador(ctx context.Context, cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("entrenar-clasificador", flag.ContinueOnError)
	datos := fs.String("datos", "../entrenamiento-metrin/data/metrin", "dir con intenciones.json")
	salida := fs.String("salida", "internal/clasificar/defecto.json", "ruta del modelo")
	umbral := fs.Float64("umbral", clasificar.UmbralDefecto, "similitud mínima")
	if _, err := separar(fs, args); err != nil {
		return err
	}
	b, err := os.ReadFile(filepath.Join(*datos, "intenciones.json"))
	if err != nil {
		return err
	}
	var porCat map[string][]string
	if err := json.Unmarshal(b, &porCat); err != nil {
		return err
	}
	ejemplos := map[string][]string{}
	for cat, textos := range porCat {
		in, ok := intencionDe[cat]
		if !ok {
			return fmt.Errorf("categoría sin intención: %s", cat)
		}
		ejemplos[in] = append(ejemplos[in], textos...)
	}
	e, err := embed.NuevoEstatico(cfg.ModeloEstatico)
	if err != nil {
		return err
	}
	m, err := clasificar.Entrenar(ctx, e, e.Nombre(), *umbral, ejemplos)
	if err != nil {
		return err
	}
	if err := m.Guardar(*salida); err != nil {
		return err
	}
	total := 0
	for _, in := range m.Intenciones {
		logf("intención %-10s ejemplos %3d dim %d", in.Nombre, in.Ejemplos, in.Dim)
		total += in.Ejemplos
	}
	logf("modelo %s (%d ejemplos) en %s", m.HuellaEmb, total, *salida)
	return nil
}

func cmdClasificar(ctx context.Context, cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("clasificar", flag.ContinueOnError)
	modelo := fs.String("modelo", "", "ruta del modelo (vacío = embarcado)")
	enJSON := fs.Bool("json", false, "salida JSON")
	todos := fs.Bool("todos", false, "todos los puntajes (solo con --json)")
	pos, err := separar(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return fmt.Errorf("uso: rag clasificar \"<texto>\" [--modelo f] [--json]")
	}
	var m *clasificar.Modelo
	if *modelo == "" {
		m, err = clasificar.CargarDefecto()
	} else {
		m, err = clasificar.Cargar(*modelo)
	}
	if err != nil {
		return err
	}
	e, err := embed.NuevoEstatico(cfg.ModeloEstatico)
	if err != nil {
		return err
	}
	in, sim, err := m.Clasificar(ctx, e, pos[0])
	if err != nil {
		return err
	}
	if *enJSON {
		sal := map[string]any{"intencion": in, "similitud": sim}
		if *todos {
			ins, sims, err := m.Puntajes(ctx, e, pos[0])
			if err != nil {
				return err
			}
			cand := make([]map[string]any, 0, len(ins))
			for i, ci := range ins {
				cand = append(cand, map[string]any{"intencion": ci.Nombre, "similitud": sims[i]})
			}
			sal["candidatas"] = cand
		}
		b, _ := json.Marshal(sal)
		fmt.Println(string(b))
		return nil
	}
	logf("%s (%.3f)", in, sim)
	return nil
}
