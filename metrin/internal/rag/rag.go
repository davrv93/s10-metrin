// Package rag une búsqueda y LLM: recupera trozos, arma el prompt en español
// con citas, pregunta al modelo y registra lo que no se pudo responder.
package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"rag-go/internal/almacen"
	"rag-go/internal/clasificar"
	"rag-go/internal/embed"
	"rag-go/internal/indexar"
	"rag-go/internal/llm"
)

// Chateador es lo que el RAG necesita del LLM (interfaz para poder probar).
type Chateador interface {
	Chat(ctx context.Context, msgs []llm.Mensaje) (string, error)
}

type RAG struct {
	Almacen      *almacen.Almacen
	LLM          Chateador
	Emb          embed.Embebedor       // nil = sin ruta conversacional
	Clasificador *clasificar.Modelo   // nil = todo va al RAG
	MaxDistancia float64 // si el mejor trozo está más lejos, no hay contexto
	RutaFallos   string  // datos/sin_respuesta.jsonl
	mu           sync.Mutex
}

// Fuente citada en una respuesta.
type Fuente struct {
	Cita      string  `json:"cita"`
	Distancia float64 `json:"distancia"`
	Documento string  `json:"documento,omitempty"`
	Pagina    int     `json:"pagina,omitempty"`
	URL       string  `json:"url,omitempty"`
}

type Respuesta struct {
	Pregunta          string        `json:"pregunta"`
	Respuesta         string        `json:"respuesta"`
	Modo              string        `json:"modo"`
	Fuentes           []Fuente      `json:"fuentes"`
	SinContexto       bool          `json:"sin_contexto"`
	Motivo            string        `json:"motivo,omitempty"`
	TiempoBusq        time.Duration `json:"-"`
	TiempoLLM         time.Duration `json:"-"`
	MsBusqueda        int64         `json:"ms_busqueda"`
	MsLLM             int64         `json:"ms_llm"`
	DistanciaMin      float64       `json:"distancia_min"`
	PreguntaReescrita string        `json:"pregunta_reescrita,omitempty"`
}

// Turno es un mensaje previo de la conversación ("usuario" o "asistente").
type Turno struct {
	Rol   string `json:"rol"`
	Texto string `json:"texto"`
}

// Opciones de una pregunta.
type Opciones struct {
	K      int
	Filtro map[string]string // source, type, ext
	Hilo   []Turno           // últimos turnos; sirve para resolver referencias ("su", "eso")
}

// MarcaSinContexto es la frase que se pide al modelo cuando el contexto no
// alcanza; se usa también para detectarlo.
const MarcaSinContexto = "No tengo contexto suficiente"

const sistema = `Eres Metrín, asistente de Optimiza 360 para ERP S10 y procesos de construcción.
Reglas:
1. Responde SIEMPRE en español neutro, de forma cercana, profesional y clara; no uses regionalismos ni jerga. Nunca escribas en portugués ni en ningún otro idioma.
2. Usa SOLO la información del CONTEXTO. No inventes datos, configuraciones, acciones ni fuentes.
3. Si el contexto no alcanza, responde exactamente: "` + MarcaSinContexto + ` para responder a esa pregunta." y nada más.
4. Contesta primero la pregunta, con palabras propias; no copies párrafos del contexto ni enumeres títulos de fuentes.
5. Sé breve: hasta 120 palabras, salvo que el usuario pida pasos; para procedimientos usa como máximo 5 pasos cortos.
6. No incluyas encabezados como "conocimiento del proyecto" ni cites etiquetas entre corchetes: la interfaz muestra las fuentes por separado.
7. Si la pregunta es sobre Optimiza 360, responde con información sobre la empresa; no mezcles procedimientos de S10 que no ayuden a contestarla.
8. Si el contexto habla de temas distintos, usa solo el que responde a la pregunta. Si falta evidencia, dilo con la frase indicada.
9. No incluyas teléfonos, correos, datos personales ni nombres de personas en una respuesta general, salvo que el usuario pregunte expresamente por un contacto y la fuente sea pertinente.
10. No afirmes que aprendiste o guardaste conocimiento; este servicio solo registra preguntas sin contexto.
11. Si preguntan qué tipos de presupuestos existen, distingue la clasificación del manual (Venta, Meta y Línea Base) de los términos enumerados en el sílabo (base, real, oferta, meta y línea base). No menciones logotipos, asignaciones, copias o permisos como tipos.`

const maxContextos = 5
const ventanaRelevancia = 0.18

// ManualCortex: valor del campo manual para el conocimiento curado del
// proyecto. El segundo pase lo recupera siempre etiquetado.
const ManualCortex = "Cortex"

// KCortex: trozos curados del segundo pase.
const KCortex = 8

// Preguntar responde una pregunta.
func (r *RAG) Preguntar(ctx context.Context, pregunta string, o Opciones) (Respuesta, error) {
	if o.K <= 0 {
		o.K = 8
	}
	res := Respuesta{Pregunta: pregunta, Modo: "respuesta"}
	// Ruta conversacional: social y límite responden directo sin retrieval.
	// Ayuda va al RAG (necesita contexto para ayudar de verdad); si el RAG
	// no halla nada, el fallback de mensaje corto conversa. Si el
	// clasificador falla, se sigue al RAG (ruta segura).
	// Excepción: pregunta larga con "?" ("bien, oye, qué es optimiza?")
	// casi siempre pide datos: va al RAG aunque huela a charla.
	if r.Clasificador != nil && r.Emb != nil {
		if in, _, err := r.Clasificador.Clasificar(ctx, r.Emb, pregunta); err == nil && (in == clasificar.Social || in == "limite") && !esPreguntaLarga(pregunta) {
			return r.conversar(ctx, in, pregunta, o.Hilo)
		}
	}
	// Hilo: la pregunta efectiva resuelve referencias contra la conversación.
	efectiva := pregunta
	if len(o.Hilo) > 0 {
		if re, err := r.reescribir(ctx, o.Hilo, pregunta); err == nil && re != "" {
			efectiva = re
			res.PreguntaReescrita = re
		}
	}
	t0 := time.Now()
	trozos, err := r.Almacen.Buscar(ctx, efectiva, o.K, o.Filtro)
	res.TiempoBusq = time.Since(t0)
	res.MsBusqueda = res.TiempoBusq.Milliseconds()
	if err != nil {
		return res, err
	}
	// Segundo pase dinámico: conocimiento curado Cortex. Solo cuando la
	// pregunta puede tocar s10-kb (sin filtro o con ese source).
	var cortex []almacen.Resultado
	if o.Filtro == nil || o.Filtro["source"] == "" || o.Filtro["source"] == indexar.FuenteS10KB {
		cortex, _ = r.Almacen.Buscar(ctx, efectiva, KCortex, map[string]string{"manual": ManualCortex})
	}
	candidatos := append(append([]almacen.Resultado(nil), trozos...), cortex...)
	seleccionados := seleccionarContexto(candidatos, r.MaxDistancia)
	if esConsultaTiposPresupuesto(pregunta) {
		if contexto := contextoTiposPresupuesto(cortex); contexto != nil {
			seleccionados = []almacen.Resultado{*contexto}
		}
	}
	res.DistanciaMin = 1
	if len(seleccionados) > 0 {
		res.DistanciaMin = redondear(seleccionados[0].Distancia)
	}
	for _, t := range seleccionados {
		cita := fuenteDe(t)
		if !contieneCita(res.Fuentes, cita.Cita) {
			res.Fuentes = append(res.Fuentes, cita)
		}
	}
	if len(seleccionados) == 0 || res.DistanciaMin > r.MaxDistancia {
		res.SinContexto = true
		res.Modo = "sin_contexto"
		res.Motivo = fmt.Sprintf("distancia mínima %.3f > umbral %.3f", res.DistanciaMin, r.MaxDistancia)
		if len(trozos) == 0 {
			res.Motivo = "el índice no devolvió trozos"
		}
		// Mensaje corto sin contexto = casi siempre charla, no pregunta de
		// trabajo ("bien y tú", "ok", "¿precio?"). Se conversa en vez de
		// levantar un muro; igual se registra el fallo para aprender.
		if len(palabras(pregunta)) <= 8 {
			conv, err := r.conversar(ctx, "charla", pregunta, o.Hilo)
			if err == nil {
				conv.Motivo = "sin contexto útil; respuesta conversacional (" + res.Motivo + ")"
				_ = r.registrarFallo(res)
				return conv, nil
			}
		}
		res.Respuesta = MarcaSinContexto + " para responder a esa pregunta: no encontré nada relacionado en los documentos indexados."
		return res, r.registrarFallo(res)
	}

	var ctxTxt strings.Builder
	for _, t := range seleccionados {
		fmt.Fprintf(&ctxTxt, "FUENTE: %s\n%s\n\n", t.Metadata["cita"], t.Texto)
	}
	historial := ""
	if len(o.Hilo) > 0 {
		var h strings.Builder
		h.WriteString("CONVERSACIÓN PREVIA (resuelve pronombres y referencias con ella):\n")
		empieza := max(0, len(o.Hilo)-4)
		for _, t := range o.Hilo[empieza:] {
			if t.Rol != "usuario" && t.Rol != "asistente" {
				continue
			}
			fmt.Fprintf(&h, "%s: %s\n", t.Rol, recortar(t.Texto, 200))
		}
		historial = h.String()
	}
	// Modo tutorial: si el contexto es una guía paso a paso, se responde con
	// todos los pasos detallados, sin el tope de brevedad.
	instruccion := "Responde directamente con los datos pertinentes del contexto. No menciones las etiquetas FUENTE ni describas cómo hiciste la búsqueda."
	if esTutorial(seleccionados) {
		res.Modo = "tutorial"
		instruccion = "Es una guía paso a paso: responde con TODOS los pasos necesarios, numerados, cada uno con la acción concreta (dónde hacer clic, qué llenar, qué validar). Sin límite de palabras; la brevedad no aplica aquí."
	}
	msgs := []llm.Mensaje{
		{Role: "system", Content: sistema},
		{Role: "user", Content: "CONTEXTO:\n" + ctxTxt.String() + historial + "PREGUNTA: " + efectiva + "\n\n" + instruccion},
	}
	t1 := time.Now()
	texto, err := r.LLM.Chat(ctx, msgs)
	res.TiempoLLM = time.Since(t1)
	res.MsLLM = res.TiempoLLM.Milliseconds()
	if err != nil {
		res.Respuesta = "Encontré fuentes relacionadas, pero no pude preparar un resumen ahora. Inténtalo de nuevo en unos minutos."
		res.Modo = "modelo_no_disponible"
		res.Motivo = "el modelo conversacional no está disponible"
		return res, nil
	}
	// Qwen2.5-3B puede filtrar portugués: reintenta una vez y nunca lo muestra.
	if parecePortugues(texto) {
		msgs = append(msgs, llm.Mensaje{Role: "user", Content: "Reescribe tu respuesta anterior ÚNICAMENTE en español neutro, sin una sola palabra en portugués."})
		if t2, err2 := r.LLM.Chat(ctx, msgs); err2 == nil {
			texto = t2
		}
	}
	if parecePortugues(texto) {
		res.Respuesta = "Encontré fuentes relacionadas, pero no pude redactar una respuesta fiable en español. Inténtalo de nuevo en unos minutos."
		res.Modo = "modelo_no_disponible"
		res.Motivo = "el modelo no pudo responder en español"
		return res, nil
	}
	res.Respuesta = texto
	if DiceSinContexto(texto) {
		res.SinContexto = true
		res.Motivo = "el modelo dijo que el contexto no alcanza"
		return res, r.registrarFallo(res)
	}
	return res, nil
}

func esConsultaTiposPresupuesto(pregunta string) bool {
	t := strings.ToLower(pregunta)
	if !strings.Contains(t, "presupuest") {
		return false
	}
	return strings.Contains(t, "tipo") || strings.Contains(t, "clase") ||
		strings.Contains(t, "existen") || strings.Contains(t, "hay") ||
		strings.Contains(t, "cuáles") || strings.Contains(t, "cuales")
}

func contextoTiposPresupuesto(candidatos []almacen.Resultado) *almacen.Resultado {
	for i := range candidatos {
		t := strings.ToLower(candidatos[i].Texto)
		if candidatos[i].Metadata["manual"] == ManualCortex &&
			strings.Contains(t, "venta") && strings.Contains(t, "meta") &&
			strings.Contains(t, "línea base") {
			return &candidatos[i]
		}
	}
	return nil
}

// parecePortugues detecta filtraciones del otro idioma del modelo (à/ã/õ/ç
// no existen en español + lista cerrada de palabras exclusivas del portugués).
func parecePortugues(t string) bool {
	t = strings.ToLower(" " + t + " ")
	if strings.ContainsAny(t, "àãõç") {
		return true
	}
	for _, p := range []string{"também", "tambem", " uma ", " umas ", " não ", " nao ",
		" dos ", " das ", " dum ", " num ", " numa ", " neste ", " nesta ",
		"desse ", " dessa ", " disso ", " você ", " voce ", " obrigado",
		" muito ", " foi ", " têm ", " são ", " sao ", " pois ", " através ",
		"atráves", " estão ", " estão"} {
		if strings.Contains(t, p) {
			return true
		}
	}
	return false
}

func seleccionarContexto(candidatos []almacen.Resultado, maxDistancia float64) []almacen.Resultado {
	if len(candidatos) == 0 {
		return nil
	}
	sort.SliceStable(candidatos, func(i, j int) bool {
		return candidatos[i].Distancia < candidatos[j].Distancia
	})
	mejor := candidatos[0].Distancia
	limite := min(maxDistancia, mejor+ventanaRelevancia)
	seleccionados := make([]almacen.Resultado, 0, maxContextos)
	vistos := map[string]bool{}
	for _, t := range candidatos {
		if t.Distancia > limite || len(seleccionados) == maxContextos {
			break
		}
		clave := claveCita(t.Metadata["cita"])
		if clave == "" {
			clave = strings.ToLower(strings.TrimSpace(t.Metadata["title"] + ":" + t.Metadata["page"]))
			if clave == ":" {
				clave = t.ID
			}
		}
		if vistos[clave] {
			continue
		}
		vistos[clave] = true
		seleccionados = append(seleccionados, t)
	}
	return seleccionados
}

func contieneCita(fuentes []Fuente, cita string) bool {
	clave := claveCita(cita)
	for _, f := range fuentes {
		if claveCita(f.Cita) == clave {
			return true
		}
	}
	return false
}

func claveCita(cita string) string {
	return strings.ToLower(strings.Join(strings.Fields(cita), " "))
}

// reescribir convierte una pregunta con referencias ("su", "eso", "ahí") en
// autónoma usando el hilo. Si falla, devuelve "" y se usa la original.
func (r *RAG) reescribir(ctx context.Context, hilo []Turno, pregunta string) (string, error) {
	var h strings.Builder
	empieza := max(0, len(hilo)-6)
	for _, t := range hilo[empieza:] {
		if t.Rol != "usuario" && t.Rol != "asistente" {
			continue
		}
		fmt.Fprintf(&h, "%s: %s\n", t.Rol, recortar(t.Texto, 300))
	}
	re, err := r.LLM.Chat(ctx, []llm.Mensaje{
		{Role: "system", Content: "Reescribes preguntas de seguimiento como preguntas autónomas en español. Responde SOLO con la pregunta reescrita, sin comillas ni explicaciones."},
		{Role: "user", Content: "CONVERSACIÓN:\n" + h.String() + "PREGUNTA NUEVA: " + pregunta},
	})
	if err != nil {
		return "", err
	}
	re = strings.TrimSpace(re)
	if i := strings.Index(re, "\n"); i >= 0 {
		re = strings.TrimSpace(re[:i])
	}
	return recortar(re, 500), nil
}

func recortar(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n])
}

// sistemaCharla: charla directa sin retrieval (el clasificador ya decidió que
// no es pregunta de trabajo). Corta, en español, sin prometer acciones.
const sistemaCharla = `Eres Metrín, asistente de Optimiza 360. Respondes SIEMPRE en español neutro, con calidez y brevedad (máximo 60 palabras). Nunca escribas en portugués ni en ningún otro idioma. No inventas datos ni afirmas acciones que no realizaste. Si el mensaje menciona temas de obra, S10, presupuestos o pide explicaciones técnicas, NO los expliques: pide con amabilidad que precisen la pregunta. Si piden algo fuera de tu alcance (poemas, tareas escolares), declínalo amable y ofrece ayuda con S10 u obra.`

// conversar responde charla directa con el estilo Metrín, sin retrieval.
// Límite no llama al modelo: respuesta fija (el 3B no obedece el rechazo).
func (r *RAG) conversar(ctx context.Context, intencion, pregunta string, hilo []Turno) (Respuesta, error) {
	res := Respuesta{Pregunta: pregunta, Modo: "conversacional"}
	if intencion == "limite" {
		res.Respuesta = "Eso está fuera de mi alcance, pero te ayudo con S10 y obra. ¿Qué necesitas?"
		return res, nil
	}
	msgs := []llm.Mensaje{{Role: "system", Content: sistemaCharla}}
	if len(hilo) > 0 {
		var h strings.Builder
		h.WriteString("Conversación previa:\n")
		empieza := max(0, len(hilo)-4)
		for _, t := range hilo[empieza:] {
			if t.Rol != "usuario" && t.Rol != "asistente" {
				continue
			}
			fmt.Fprintf(&h, "%s: %s\n", t.Rol, recortar(t.Texto, 200))
		}
		msgs = append(msgs, llm.Mensaje{Role: "user", Content: h.String() + "Mensaje nuevo (" + intencion + "): " + pregunta})
	} else {
		msgs = append(msgs, llm.Mensaje{Role: "user", Content: pregunta})
	}
	t1 := time.Now()
	texto, err := r.LLM.Chat(ctx, msgs)
	res.TiempoLLM = time.Since(t1)
	res.MsLLM = res.TiempoLLM.Milliseconds()
	if err != nil {
		res.Respuesta = "Hola, ¿en qué te ayudo hoy?"
		res.Modo = "conversacional"
		res.Motivo = "modelo no disponible; saludo de respaldo"
		return res, nil
	}
	if parecePortugues(texto) {
		msgs = append(msgs, llm.Mensaje{Role: "user", Content: "Reescribe ÚNICAMENTE en español neutro."})
		if t2, err2 := r.LLM.Chat(ctx, msgs); err2 == nil {
			texto = t2
		}
	}
	if parecePortugues(texto) {
		texto = "Hola, ¿en qué te ayudo hoy?"
		res.Motivo = "el modelo mezcló portugués; saludo de respaldo"
	}
	res.Respuesta = texto
	return res, nil
}

// esPreguntaLarga: "?" con más de 3 palabras ("cómo estás?" queda fuera).
func esPreguntaLarga(s string) bool {
	if !strings.Contains(s, "?") {
		return false
	}
	return len(palabras(s)) > 3
}

// esTutorial: el contexto trae una guía paso a paso (documento de tutorial
// entre los 3 mejores). Ahí la brevedad sobra: hay que detallar.
func esTutorial(trozos []almacen.Resultado) bool {
	for i, t := range trozos {
		if i >= 3 {
			break
		}
		if strings.Contains(strings.ToLower(t.Metadata["document_id"]), "tutorial") {
			return true
		}
	}
	return false
}

func palabras(s string) []string {
	return strings.Fields(s)
}

// fuenteDe arma la cita de un resultado para la respuesta.
func fuenteDe(t almacen.Resultado) Fuente {
	pagina := 0
	fmt.Sscan(t.Metadata["page"], &pagina)
	return Fuente{
		Cita:      t.Metadata["cita"],
		Distancia: redondear(t.Distancia),
		Documento: t.Metadata["title"],
		Pagina:    pagina,
		URL:       t.Metadata["source_url"],
	}
}

// DiceSinContexto detecta la frase de rechazo (o variantes cercanas).
func DiceSinContexto(texto string) bool {
	t := strings.ToLower(texto)
	for _, m := range []string{"no tengo contexto suficiente", "no hay contexto suficiente", "el contexto no alcanza", "no tengo suficiente contexto", "no dispongo de información suficiente"} {
		if strings.Contains(t, m) {
			return true
		}
	}
	return false
}

func (r *RAG) registrarFallo(res Respuesta) error {
	if r.RutaFallos == "" {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(r.RutaFallos), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(r.RutaFallos, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	linea, _ := json.Marshal(map[string]any{
		"fecha":         time.Now().UTC().Format(time.RFC3339),
		"pregunta":      res.Pregunta,
		"motivo":        res.Motivo,
		"distancia_min": res.DistanciaMin,
		"fuentes":       res.Fuentes,
	})
	_, err = f.Write(append(linea, '\n'))
	return err
}

func redondear(x float64) float64 { return float64(int(x*1000+0.5)) / 1000 }
