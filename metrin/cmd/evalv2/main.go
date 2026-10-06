// Command evalv2 es el benchmark V1 frente a V2 de Metrín (docs/V2-RAG-PROCEDURAL.md §12).
//
// Lanza el MISMO dataset (metrin/eval/v2_oro.jsonl) contra una URL de Metrín dos veces —una con
// "version":"v1" y otra con "version":"v2", las dos con "traza":true—, reproduce cada conversación como la
// página (hilo de los últimos 6 turnos y, en V2, la memoria estructurada) y califica el último turno de
// cada caso con las mismas métricas para las dos versiones. La métrica principal es PROCEDURAL ANSWER
// SUCCESS (siete condiciones). Si la V2 no está en el servidor (no devuelve `plan`), se marca «no
// disponible» y se sigue con V1.
//
// Salida: tabla en consola, metrin/eval/V1_VS_V2.md y el JSON crudo metrin/eval/resultados/<fecha>.json.
//
// Uso (desde metrin/):
//
//	go run ./cmd/evalv2 --url http://127.0.0.1:4762 --contenedor metrin-traza-prueba
//	go run ./cmd/evalv2 --muestra 20            # 20 casos alternando categorías
//	go run ./cmd/evalv2 --validar               # solo comprueba el dataset contra kb/ (sin HTTP)
//	go run ./cmd/evalv2 --versiones v1          # solo V1
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type opciones struct {
	URL        string
	Dataset    string
	Raiz       string
	Versiones  []string
	Muestra    int
	Contenedor string
	Timeout    time.Duration
	SalidaMD   string
	SalidaJSON string
	Validar    bool
	Source     string
	K          int
	Estricto   bool
}

// Resultado de una versión.
type ResultadoVersion struct {
	Version    string                  `json:"version"`
	Disponible bool                    `json:"disponible"`
	Motivo     string                  `json:"motivo,omitempty"`
	TurnosPlan int                     `json:"turnos_con_plan"`
	Global     Agregado                `json:"global"`
	PorCat     map[string]Agregado     `json:"por_categoria"`
	Casos      []EvalCaso              `json:"casos"`
	Recursos   []MuestraRecursos       `json:"recursos,omitempty"`
	Crudos     map[string][]TurnoCrudo `json:"crudos"`
	Inicio     time.Time               `json:"inicio"`
	Duracion   float64                 `json:"duracion_s"`
}

// Corrida completa (lo que va al JSON crudo).
type Corrida struct {
	Fecha       time.Time           `json:"fecha"`
	URL         string              `json:"url"`
	Contenedor  string              `json:"contenedor,omitempty"`
	Dataset     string              `json:"dataset"`
	Casos       int                 `json:"casos"`
	Composicion map[string]int      `json:"composicion"`
	Excluidos   []string            `json:"excluidos,omitempty"`
	AvisosKB    []string            `json:"avisos_kb,omitempty"`
	AvisosData  []string            `json:"avisos_dataset,omitempty"`
	KB          map[string]int      `json:"kb"`
	Health      json.RawMessage     `json:"health,omitempty"`
	Versiones   []*ResultadoVersion `json:"versiones"`
}

func main() {
	var op opciones
	var versiones string
	flag.StringVar(&op.URL, "url", "http://127.0.0.1:4762", "URL base de Metrín")
	flag.StringVar(&op.Dataset, "dataset", "", "dataset JSONL (por defecto <metrin>/eval/v2_oro.jsonl)")
	flag.StringVar(&op.Raiz, "raiz", "", "raíz del repo s10-conocimiento (la que tiene kb/); por defecto se busca hacia arriba")
	flag.StringVar(&versiones, "versiones", "v1,v2", "versiones a evaluar, separadas por coma")
	flag.IntVar(&op.Muestra, "muestra", 0, "evaluar solo N casos (alternando categorías, determinista)")
	flag.StringVar(&op.Contenedor, "contenedor", "", "contenedor Docker del que leer RAM y CPU (docker stats --no-stream, solo lectura)")
	flag.DurationVar(&op.Timeout, "timeout", 120*time.Second, "tiempo máximo por petición")
	flag.StringVar(&op.SalidaMD, "md", "", "informe Markdown (por defecto <metrin>/eval/V1_VS_V2.md)")
	flag.StringVar(&op.SalidaJSON, "json", "", "carpeta del JSON crudo (por defecto <metrin>/eval/resultados)")
	flag.BoolVar(&op.Validar, "validar", false, "solo validar el dataset contra kb/ y salir")
	flag.StringVar(&op.Source, "source", "s10-kb", "campo source de la petición (el que manda la página)")
	flag.IntVar(&op.K, "k", 8, "campo k de la petición (el que manda la página)")
	flag.BoolVar(&op.Estricto, "estricto", false, "fallar si algún caso no valida (por defecto se excluye y se avisa)")
	flag.Parse()
	for _, v := range strings.Split(versiones, ",") {
		if v = strings.TrimSpace(strings.ToLower(v)); v != "" {
			op.Versiones = append(op.Versiones, v)
		}
	}
	if err := ejecutar(op); err != nil {
		fmt.Fprintln(os.Stderr, "evalv2:", err)
		os.Exit(1)
	}
}

// buscarRaiz sube desde el directorio actual hasta encontrar kb/procedimientos.
func buscarRaiz() (string, error) {
	dir, _ := os.Getwd()
	for {
		if st, err := os.Stat(filepath.Join(dir, "kb", "procedimientos")); err == nil && st.IsDir() {
			return dir, nil
		}
		padre := filepath.Dir(dir)
		if padre == dir {
			return "", fmt.Errorf("no encontré kb/procedimientos hacia arriba de %s: use --raiz", dir)
		}
		dir = padre
	}
}

func ejecutar(op opciones) error {
	if op.Raiz == "" {
		r, err := buscarRaiz()
		if err != nil {
			return err
		}
		op.Raiz = r
	}
	metrin := filepath.Join(op.Raiz, "metrin")
	if op.Dataset == "" {
		op.Dataset = filepath.Join(metrin, "eval", "v2_oro.jsonl")
	}
	if op.SalidaMD == "" {
		op.SalidaMD = filepath.Join(metrin, "eval", "V1_VS_V2.md")
	}
	if op.SalidaJSON == "" {
		op.SalidaJSON = filepath.Join(metrin, "eval", "resultados")
	}

	base, err := cargarBase(op.Raiz)
	if err != nil {
		return err
	}
	casos, err := cargarDataset(op.Dataset)
	if err != nil {
		return err
	}
	errores, avisos := validarDataset(casos, base)
	fmt.Printf("KB: %d fragmentos, %d procedimientos, %d conceptos · dataset: %d casos (%s)\n",
		len(base.porClave), len(base.Procedimientos), len(base.Conceptos), len(casos), op.Dataset)
	for _, a := range base.Avisos {
		fmt.Println("  aviso KB:", a)
	}
	malos := map[string]bool{}
	for _, e := range errores {
		fmt.Println("  ERROR dataset:", e)
		id, _, _ := strings.Cut(e, ":")
		malos[id] = true
	}
	if len(avisos) > 0 {
		fmt.Printf("  %d avisos del dataset (detalle en el JSON crudo y con --validar)\n", len(avisos))
	}
	if op.Validar {
		for _, a := range avisos {
			fmt.Println("  aviso:", a)
		}
		fmt.Print(composicionTexto(casos))
		if len(errores) > 0 {
			return fmt.Errorf("%d errores en el dataset", len(errores))
		}
		fmt.Println("dataset válido")
		return nil
	}
	if len(errores) > 0 && op.Estricto {
		return fmt.Errorf("%d errores en el dataset (--estricto)", len(errores))
	}
	var validos []Caso
	var excluidos []string
	for _, c := range casos {
		if malos[c.ID] {
			excluidos = append(excluidos, c.ID)
			continue
		}
		validos = append(validos, c)
	}
	validos = muestra(validos, op.Muestra)
	porID := map[string]Caso{}
	comp := map[string]int{}
	for _, c := range validos {
		porID[c.ID] = c
		comp[c.Categoria]++
	}

	corrida := &Corrida{
		Fecha: time.Now(), URL: op.URL, Contenedor: op.Contenedor, Dataset: op.Dataset, Casos: len(validos),
		Composicion: comp, Excluidos: excluidos, AvisosKB: base.Avisos, AvisosData: avisos,
		KB: map[string]int{"fragmentos": len(base.porClave), "procedimientos": len(base.Procedimientos), "conceptos": len(base.Conceptos)},
	}
	corrida.Health = leerHealth(op.URL)

	for _, v := range op.Versiones {
		rv := correrVersion(op, base, v, validos, porID)
		corrida.Versiones = append(corrida.Versiones, rv)
	}

	fmt.Println()
	fmt.Print(tablaConsola(corrida))

	if err := os.MkdirAll(op.SalidaJSON, 0o755); err != nil {
		return err
	}
	rutaJSON := filepath.Join(op.SalidaJSON, corrida.Fecha.Format("2006-01-02T150405")+".json")
	b, _ := json.Marshal(corrida)
	if err := os.WriteFile(rutaJSON, b, 0o644); err != nil {
		return err
	}
	md := informeMarkdown(corrida, op, base, porID, rutaJSON)
	if err := os.MkdirAll(filepath.Dir(op.SalidaMD), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(op.SalidaMD, []byte(md), 0o644); err != nil {
		return err
	}
	fmt.Printf("\nInforme: %s\nJSON crudo: %s\n", op.SalidaMD, rutaJSON)
	return nil
}

func leerHealth(url string) json.RawMessage {
	c := nuevoCliente(url, 5*time.Second, "", 0)
	res, err := c.http.Get(c.url + "/health")
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	var m json.RawMessage
	if json.NewDecoder(res.Body).Decode(&m) != nil {
		return nil
	}
	return m
}

// Preguntas de sondeo para saber si la V2 está en el servidor: si ninguna trae `plan`, no lo está.
var sondeoV2 = []string{"¿Cómo registro un nuevo presupuesto en S10?", "como hago un metrado", "¿qué es un metrado?"}

func correrVersion(op opciones, base *Base, version string, casos []Caso, porID map[string]Caso) *ResultadoVersion {
	rv := &ResultadoVersion{Version: version, Disponible: true, Crudos: map[string][]TurnoCrudo{}, PorCat: map[string]Agregado{}, Inicio: time.Now()}
	ctx := context.Background()
	cli := nuevoCliente(op.URL, op.Timeout, op.Source, op.K)
	if version == "v2" {
		con := false
		var motivo []string
		for _, q := range sondeoV2 {
			cli.nuevaConversacion()
			t := cli.preguntar(ctx, "v2", q)
			switch {
			case t.ErrorRed != "":
				motivo = append(motivo, t.ErrorRed)
			case t.Respuesta != nil && t.Respuesta.Plan != nil:
				con = true
			}
		}
		if !con {
			rv.Disponible = false
			rv.Motivo = "ninguna de las " + fmt.Sprint(len(sondeoV2)) + " preguntas de sondeo con \"version\":\"v2\" devolvió `plan`"
			if len(motivo) > 0 {
				rv.Motivo += " (" + recortarTexto(strings.Join(unicas(motivo), "; "), 200) + ")"
			}
			fmt.Printf("V2: no disponible — %s\n", rv.Motivo)
			return rv
		}
	}
	m := iniciarMuestreo(op.Contenedor, 3*time.Second)
	fmt.Printf("%s: %d casos ", strings.ToUpper(version), len(casos))
	for i, c := range casos {
		cli.nuevaConversacion()
		var obs []Observacion
		for _, q := range c.Turnos {
			t := cli.preguntar(ctx, version, q)
			rv.Crudos[c.ID] = append(rv.Crudos[c.ID], t)
			o := observar(base, version, t)
			if o.PlanPresente {
				rv.TurnosPlan++
			}
			obs = append(obs, o)
		}
		rv.Casos = append(rv.Casos, evaluarCaso(base, c, obs))
		if (i+1)%10 == 0 {
			fmt.Print(".")
		}
	}
	rv.Recursos = m.detener()
	rv.Duracion = time.Since(rv.Inicio).Seconds()
	fmt.Printf(" %.1f s\n", rv.Duracion)
	rv.Global = agregar(rv.Casos, porID)
	porCat := map[string][]EvalCaso{}
	for _, e := range rv.Casos {
		porCat[e.Categoria] = append(porCat[e.Categoria], e)
	}
	for cat, es := range porCat {
		rv.PorCat[cat] = agregar(es, porID)
	}
	return rv
}

func composicionTexto(casos []Caso) string {
	comp := map[string]int{}
	conv, sinEv, conProc := 0, 0, 0
	procs := map[string]int{}
	for _, c := range casos {
		comp[c.Categoria]++
		if len(c.Turnos) > 1 {
			conv++
		}
		if c.SinEvidenciaEsperada {
			sinEv++
		}
		if c.Proc() != "" {
			conProc++
			procs[c.Proc()]++
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "composición: %d casos · ", len(casos))
	for _, cat := range categorias {
		fmt.Fprintf(&b, "%s %d · ", cat, comp[cat])
	}
	fmt.Fprintf(&b, "conversaciones %d · sin evidencia esperada %d · con procedimiento %d (%d procedimientos distintos)\n",
		conv, sinEv, conProc, len(procs))
	ids := make([]string, 0, len(procs))
	for id := range procs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Fprintf(&b, "  %3d  %s\n", procs[id], id)
	}
	return b.String()
}
