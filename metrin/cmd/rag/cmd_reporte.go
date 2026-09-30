package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"rag-go/internal/reportes"
)
// cmdReporte: `rag reporte "<pregunta>" [--bd ruta|url] [--plantilla id] [--sql]`.
// Elige plantilla del catálogo aprobado, resuelve parámetros, valida y ejecuta.
func cmdReporte(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("reporte", flag.ContinueOnError)
	preguntaFS := fs.String("pregunta", "", "pregunta en lenguaje natural")
	bd := fs.String("bd", "", "ruta SQLite o URL postgresql:// (read-only)")
	plantillaID := fs.String("plantilla", "", "fuerza una plantilla por id o nombre")
	catalogoRuta := fs.String("catalogo", "../reportes/plantillas.json", "ruta del catálogo aprobado")
	verSQL := fs.Bool("sql", false, "muestra plantilla, parámetros y SQL sin ejecutar")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pregunta := *preguntaFS
	if len(fs.Args()) > 0 && pregunta == "" {
		pregunta = strings.Join(fs.Args(), " ")
	}
	if pregunta == "" {
		return fmt.Errorf(`uso: rag reporte "<pregunta>" [--bd ruta|url] [--plantilla id] [--sql]`)
	}

	cat, err := reportes.CargarCatalogo(*catalogoRuta)
	if err != nil {
		return err
	}

	var p *reportes.Plantilla
	puntaje := 1.0
	if *plantillaID != "" {
		for i := range cat.Plantillas {
			if cat.Plantillas[i].ID == *plantillaID || cat.Plantillas[i].Nombre == *plantillaID {
				p = &cat.Plantillas[i]
				break
			}
		}
		if p == nil {
			return fmt.Errorf("no existe la plantilla %q en el catálogo", *plantillaID)
		}
	} else {
		p, puntaje, err = cat.Recuperar(pregunta)
		if err != nil {
			return err
		}
	}

	params := p.Completar(reportes.Resolver(pregunta))
	if *verSQL {
		fmt.Printf("plantilla: %s %s (puntaje %.2f)\n", p.ID, p.Titulo, puntaje)
		fmt.Printf("parámetros: %v\n", params)
		fmt.Println(p.SQL)
		return nil
	}

	if err := reportes.Validar(p.SQL); err != nil {
		return fmt.Errorf("plantilla %s no pasa la validación: %w", p.ID, err)
	}
	if *bd == "" {
		return fmt.Errorf("falta --bd (ruta SQLite o URL postgresql://); la demo: data/demo_s10.db")
	}

	res, err := reportes.Ejecutar(ctx, p.SQL, params, *bd)
	if err != nil {
		return err
	}

	fmt.Printf("═ %s — %s (plantilla %s, %.0f%% léxico)\n", p.ID, p.Titulo, p.ID, puntaje*100)
	fmt.Printf("parámetros: %v\n", params)
	if len(res.Filas) == 0 {
		fmt.Println("(sin filas)")
		return nil
	}
	anchos := make([]int, len(res.Columnas))
	for i, c := range res.Columnas {
		anchos[i] = len(c)
	}
	celdas := make([][]string, len(res.Filas))
	for r, f := range res.Filas {
		celdas[r] = make([]string, len(f.Valores))
		for i, v := range f.Valores {
			celdas[r][i] = fmt.Sprintf("%v", v)
			if len(celdas[r][i]) > anchos[i] {
				anchos[i] = len(celdas[r][i])
			}
		}
	}
	for i, c := range res.Columnas {
		fmt.Printf("%*s | ", anchos[i], c)
	}
	fmt.Println()
	for _, a := range anchos {
		fmt.Print(strings.Repeat("-", a) + "-+-")
	}
	fmt.Println()
	for _, fila := range celdas {
		for i, c := range fila {
			fmt.Printf("%*s | ", anchos[i], c)
		}
		fmt.Println()
	}
	if res.Truncada {
		fmt.Printf("(truncada en %d filas)\n", reportes.MaxFilas)
	}
	fmt.Printf("(%d filas en %d ms)\n", len(res.Filas), res.MS)
	return nil
}
