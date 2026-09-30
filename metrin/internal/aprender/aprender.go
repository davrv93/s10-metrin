// Package aprender cierra el ciclo de Metrín: registra cada respuesta y el
// feedback (👍/👎), junta lo que no supo responder en una cola de pendientes,
// lo repasa cada noche contra el índice actualizado y deja que el admin le
// enseñe una respuesta (se indexa en vivo). También calcula las métricas del
// agente y emite los eventos que el panel sigue por SSE.
//
// Archivos en el directorio de datos (volumen de Metrín):
//
//	interacciones.jsonl  una línea por respuesta (append)
//	feedback.jsonl       una línea por voto (append; el último voto gana)
//	pendientes.json      cola de preguntas por aprender (estado mutable)
//	aprendidos.jsonl     fragmentos enseñados por el admin, formato kb JSONL
//	                     (docker-entrada.sh los suma al index-kb del arranque)
package aprender

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"rag-go/internal/almacen"
	"rag-go/internal/indexar"
	"rag-go/internal/rag"
)

// Estados de un pendiente.
const (
	Pendiente  = "pendiente"  // nadie lo resolvió todavía
	Resuelto   = "resuelto"   // el repaso nocturno ya lo responde con fuentes
	Aprendido  = "aprendido"  // el admin enseñó la respuesta
	Descartado = "descartado" // fuera de alcance o ruido
)

// Orígenes de un pendiente.
const (
	OrigenSinContexto = "sin_contexto"
	OrigenNegativo    = "feedback_negativo"
)

// Interaccion es una respuesta de Metrín tal como se registró.
type Interaccion struct {
	ID           string  `json:"id"`
	Fecha        string  `json:"fecha"`
	Pregunta     string  `json:"pregunta"`
	Respuesta    string  `json:"respuesta"`
	Modo         string  `json:"modo"`
	SinContexto  bool    `json:"sin_contexto"`
	DistanciaMin float64 `json:"distancia_min"`
	Fuentes      int     `json:"fuentes"`
	Sugerencias  int     `json:"sugerencias"`
	MsBusqueda   int64   `json:"ms_busqueda"`
	MsLLM        int64   `json:"ms_llm"`
	Voto         int     `json:"voto,omitempty"` // +1 / -1; se une desde feedback.jsonl
	Comentario   string  `json:"comentario,omitempty"`
}

// Feedback es un voto sobre una interacción.
type Feedback struct {
	ID         string `json:"id"`
	Fecha      string `json:"fecha"`
	Voto       int    `json:"voto"`
	Comentario string `json:"comentario,omitempty"`
}

// PendienteItem es una pregunta que Metrín debe aprender.
type PendienteItem struct {
	Clave       string `json:"clave"` // pregunta normalizada (deduplica)
	Pregunta    string `json:"pregunta"`
	Origen      string `json:"origen"`
	Estado      string `json:"estado"`
	Veces       int    `json:"veces"`
	Primera     string `json:"primera"`
	Ultima      string `json:"ultima"`
	Interaccion string `json:"interaccion,omitempty"` // última interacción que la originó
	Respuesta   string `json:"respuesta,omitempty"`   // mala (feedback) o la que la resolvió
	Comentario  string `json:"comentario,omitempty"`
	Intentos    int    `json:"intentos"` // repasos nocturnos que no la resolvieron
	Resuelto    string `json:"resuelto,omitempty"`
}

// Repaso resume la última corrida de aprendizaje.
type Repaso struct {
	Inicio    string `json:"inicio"`
	Fin       string `json:"fin,omitempty"`
	Revisadas int    `json:"revisadas"`
	Resueltas int    `json:"resueltas"`
	Origen    string `json:"origen"` // noche | admin
}

// Diario guarda todo en disco y mantiene índices en memoria.
type Diario struct {
	dir   string
	Bus   *Bus
	mu    sync.Mutex
	inter map[string]*Interaccion
	orden []string // ids en orden de llegada
	pend  map[string]*PendienteItem
	// repaso en curso (nil si no hay) y el último terminado
	enCurso *Repaso
	ultimo  *Repaso
}

// ErrNoEncontrado: id o clave desconocidos.
var ErrNoEncontrado = errors.New("no encontrado")

// ErrRepasoEnCurso: ya hay un repaso corriendo.
var ErrRepasoEnCurso = errors.New("ya hay un repaso en curso")

// Abrir carga (o crea) el diario en dir.
func Abrir(dir string) (*Diario, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	d := &Diario{dir: dir, Bus: NuevoBus(), inter: map[string]*Interaccion{}, pend: map[string]*PendienteItem{}}
	if err := leerJSONL(d.ruta("interacciones.jsonl"), func(b []byte) {
		var in Interaccion
		if json.Unmarshal(b, &in) == nil && in.ID != "" {
			if _, ya := d.inter[in.ID]; !ya {
				d.orden = append(d.orden, in.ID)
			}
			d.inter[in.ID] = &in
		}
	}); err != nil {
		return nil, err
	}
	if err := leerJSONL(d.ruta("feedback.jsonl"), func(b []byte) {
		var f Feedback
		if json.Unmarshal(b, &f) == nil {
			if in, ok := d.inter[f.ID]; ok {
				in.Voto, in.Comentario = f.Voto, f.Comentario
			}
		}
	}); err != nil {
		return nil, err
	}
	if b, err := os.ReadFile(d.ruta("pendientes.json")); err == nil {
		var lista []*PendienteItem
		if err := json.Unmarshal(b, &lista); err != nil {
			return nil, fmt.Errorf("pendientes.json dañado: %w", err)
		}
		for _, p := range lista {
			d.pend[p.Clave] = p
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if b, err := os.ReadFile(d.ruta("repaso.json")); err == nil {
		var r Repaso
		if json.Unmarshal(b, &r) == nil {
			d.ultimo = &r
		}
	}
	return d, nil
}

func (d *Diario) ruta(n string) string { return filepath.Join(d.dir, n) }

// RutaAprendidos es el JSONL de fragmentos enseñados.
func (d *Diario) RutaAprendidos() string { return d.ruta("aprendidos.jsonl") }

func leerJSONL(ruta string, f func([]byte)) error {
	fh, err := os.Open(ruta)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer fh.Close()
	sc := bufio.NewScanner(fh)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		if b := sc.Bytes(); len(strings.TrimSpace(string(b))) > 0 {
			f(b)
		}
	}
	return sc.Err()
}

func anexar(ruta string, v any) error {
	f, err := os.OpenFile(ruta, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

// guardarPendientes reescribe la cola de forma atómica. Llamar con mu tomado.
func (d *Diario) guardarPendientes() error {
	lista := make([]*PendienteItem, 0, len(d.pend))
	for _, p := range d.pend {
		lista = append(lista, p)
	}
	sort.Slice(lista, func(i, j int) bool { return lista[i].Primera < lista[j].Primera })
	b, err := json.MarshalIndent(lista, "", " ")
	if err != nil {
		return err
	}
	tmp := d.ruta("pendientes.json.tmp")
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, d.ruta("pendientes.json"))
}

func ahora() string { return time.Now().Format(time.RFC3339) }

func nuevoID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x%s", time.Now().UnixMilli(), hex.EncodeToString(b))
}

// ClaveDe normaliza una pregunta para deduplicar pendientes.
func ClaveDe(p string) string {
	p = strings.ToLower(strings.Trim(strings.TrimSpace(p), "¿?¡!.,;: "))
	p = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u").Replace(p)
	return strings.Join(strings.Fields(p), " ")
}

func recortar(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "…"
}

// Registrar anota una respuesta, le asigna id y, si no tuvo contexto, la
// encola para aprender. Devuelve el id.
func (d *Diario) Registrar(res rag.Respuesta) (string, error) {
	in := &Interaccion{
		ID: nuevoID(), Fecha: ahora(), Pregunta: res.Pregunta, Respuesta: recortar(res.Respuesta, 800),
		Modo: res.Modo, SinContexto: res.SinContexto, DistanciaMin: res.DistanciaMin,
		Fuentes: len(res.Fuentes), Sugerencias: len(res.Sugerencias),
		MsBusqueda: res.MsBusqueda, MsLLM: res.MsLLM,
	}
	d.mu.Lock()
	err := anexar(d.ruta("interacciones.jsonl"), in)
	if err == nil {
		d.inter[in.ID] = in
		d.orden = append(d.orden, in.ID)
		if res.SinContexto {
			d.encolar(in, OrigenSinContexto, "")
			err = d.guardarPendientes()
		}
	}
	d.mu.Unlock()
	if err != nil {
		return "", err
	}
	d.Bus.Publicar("interaccion", in)
	return in.ID, nil
}

// encolar suma la pregunta a la cola (o la reabre si ya estaba resuelta).
// Llamar con mu tomado.
func (d *Diario) encolar(in *Interaccion, origen, comentario string) *PendienteItem {
	clave := ClaveDe(in.Pregunta)
	p, ok := d.pend[clave]
	if !ok {
		p = &PendienteItem{Clave: clave, Pregunta: in.Pregunta, Origen: origen, Estado: Pendiente, Primera: in.Fecha}
		d.pend[clave] = p
	}
	// «veces» cuenta preguntas, no votos: un 👎 (o su comentario) sobre la
	// misma respuesta no la suma otra vez.
	if !ok || origen != OrigenNegativo || p.Interaccion != in.ID {
		p.Veces++
	}
	p.Ultima = ahora()
	p.Interaccion = in.ID
	if origen == OrigenNegativo {
		if comentario != "" {
			p.Comentario = comentario
		}
		if !in.SinContexto {
			// Respondió, pero no convenció: queda la respuesta para revisarla.
			p.Origen, p.Respuesta = origen, in.Respuesta
		}
		if p.Estado != Aprendido {
			p.Estado = Pendiente // un 👎 reabre lo «resuelto» o descartado
		}
	}
	return p
}

// Votar registra 👍 (+1) o 👎 (-1) sobre una interacción. Un 👎 en una
// respuesta de trabajo la manda a la cola de aprendizaje.
func (d *Diario) Votar(id string, voto int, comentario string) error {
	if voto != 1 && voto != -1 {
		return fmt.Errorf("voto debe ser 1 o -1")
	}
	comentario = recortar(comentario, 500)
	d.mu.Lock()
	in, ok := d.inter[id]
	if !ok {
		d.mu.Unlock()
		return ErrNoEncontrado
	}
	f := Feedback{ID: id, Fecha: ahora(), Voto: voto, Comentario: comentario}
	err := anexar(d.ruta("feedback.jsonl"), f)
	var p *PendienteItem
	if err == nil {
		in.Voto, in.Comentario = voto, comentario
		if voto < 0 && in.Modo != "conversacional" {
			p = d.encolar(in, OrigenNegativo, comentario)
			err = d.guardarPendientes()
		}
	}
	d.mu.Unlock()
	if err != nil {
		return err
	}
	d.Bus.Publicar("feedback", map[string]any{"id": id, "voto": voto, "comentario": comentario, "pregunta": in.Pregunta})
	if p != nil {
		d.Bus.Publicar("pendiente", p)
	}
	return nil
}

// Pendientes lista la cola filtrada por estado ("" = todos), lo más pedido
// primero.
func (d *Diario) Pendientes(estado string) []PendienteItem {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []PendienteItem
	for _, p := range d.pend {
		if estado == "" || p.Estado == estado {
			out = append(out, *p)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Veces != out[j].Veces {
			return out[i].Veces > out[j].Veces
		}
		return out[i].Ultima > out[j].Ultima
	})
	return out
}

// Recientes devuelve las últimas n interacciones (la más nueva primero).
func (d *Diario) Recientes(n int) []Interaccion {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []Interaccion
	for i := len(d.orden) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, *d.inter[d.orden[i]])
	}
	return out
}

// Descartar saca una pregunta de la cola.
func (d *Diario) Descartar(clave string) error {
	d.mu.Lock()
	p, ok := d.pend[clave]
	if !ok {
		d.mu.Unlock()
		return ErrNoEncontrado
	}
	p.Estado, p.Resuelto = Descartado, ahora()
	err := d.guardarPendientes()
	cp := *p
	d.mu.Unlock()
	if err == nil {
		d.Bus.Publicar("pendiente", cp)
	}
	return err
}

// Leccion es lo que el admin enseña.
type Leccion struct {
	Clave     string   `json:"clave"`     // pendiente que resuelve (opcional)
	Pregunta  string   `json:"pregunta"`  // obligatoria si no hay clave
	Variantes []string `json:"variantes"` // otras formas de preguntarlo
	Respuesta string   `json:"respuesta"` // texto aprobado
	Titulo    string   `json:"titulo"`    // cómo se cita (defecto: «Respuesta aprobada por Optimiza 360»)
}

// TituloAprendido es la cita por defecto de lo enseñado.
const TituloAprendido = "Respuesta aprobada por Optimiza 360"

// Aprender guarda la lección como fragmento kb (con `busqueda` = pregunta y
// variantes), lo indexa en vivo y marca el pendiente como aprendido.
func (d *Diario) Aprender(ctx context.Context, a *almacen.Almacen, l Leccion) (PendienteItem, error) {
	l.Respuesta = strings.TrimSpace(l.Respuesta)
	l.Pregunta = strings.TrimSpace(l.Pregunta)
	d.mu.Lock()
	var p *PendienteItem
	if l.Clave != "" {
		p = d.pend[l.Clave]
		if p == nil {
			d.mu.Unlock()
			return PendienteItem{}, ErrNoEncontrado
		}
		if l.Pregunta == "" {
			l.Pregunta = p.Pregunta
		}
	}
	d.mu.Unlock()
	if l.Pregunta == "" || len([]rune(l.Respuesta)) < 20 {
		return PendienteItem{}, fmt.Errorf("hace falta la pregunta y una respuesta de al menos 20 caracteres")
	}
	if l.Titulo == "" {
		l.Titulo = TituloAprendido
	}
	busqueda := []string{l.Pregunta}
	for _, v := range l.Variantes {
		if v = strings.TrimSpace(v); v != "" {
			busqueda = append(busqueda, v)
		}
	}
	h := sha1.Sum([]byte(ClaveDe(l.Pregunta)))
	frag := map[string]any{
		"id":        "aprendido-" + hex.EncodeToString(h[:8]),
		"documento": "metrin:aprendidos",
		"manual":    rag.ManualCortex, // entra en el pase de conocimiento curado
		"titulo":    l.Titulo,
		"pagina":    0,
		"fuente":    "",
		"busqueda":  strings.Join(busqueda, "\n"),
		"texto":     "Pregunta: " + l.Pregunta + "\nRespuesta aprobada: " + l.Respuesta,
	}
	raw, err := json.Marshal(frag)
	if err != nil {
		return PendienteItem{}, err
	}
	doc, err := indexar.DocumentoKB(raw)
	if err != nil {
		return PendienteItem{}, err
	}
	if err := indexar.IndexarUno(ctx, indexar.KBJSONL{}, doc, a); err != nil {
		return PendienteItem{}, err
	}
	d.mu.Lock()
	err = anexar(d.RutaAprendidos(), json.RawMessage(raw))
	if err == nil {
		clave := l.Clave
		if clave == "" {
			clave = ClaveDe(l.Pregunta)
		}
		p = d.pend[clave]
		if p == nil {
			p = &PendienteItem{Clave: clave, Pregunta: l.Pregunta, Origen: "admin", Primera: ahora(), Ultima: ahora()}
			d.pend[clave] = p
		}
		p.Estado, p.Respuesta, p.Resuelto = Aprendido, l.Respuesta, ahora()
		err = d.guardarPendientes()
	}
	var cp PendienteItem
	if p != nil {
		cp = *p
	}
	d.mu.Unlock()
	if err != nil {
		return PendienteItem{}, err
	}
	d.Bus.Publicar("aprendido", cp)
	return cp, nil
}

// Repasar vuelve a preguntar cada pendiente contra el índice actual (que el
// programador pudo actualizar con fuentes nuevas). Las que ya se responden
// con fuentes quedan «resuelto». Emite el progreso por el bus.
func (d *Diario) Repasar(ctx context.Context, r *rag.RAG, origen string) (Repaso, error) {
	d.mu.Lock()
	if d.enCurso != nil {
		d.mu.Unlock()
		return Repaso{}, ErrRepasoEnCurso
	}
	rep := &Repaso{Inicio: ahora(), Origen: origen}
	d.enCurso = rep
	var claves []string
	for k, p := range d.pend {
		if p.Estado == Pendiente {
			claves = append(claves, k)
		}
	}
	d.mu.Unlock()
	sort.Strings(claves)
	d.Bus.Publicar("repaso_inicio", map[string]any{"total": len(claves), "origen": origen})

	for i, k := range claves {
		if ctx.Err() != nil {
			break
		}
		d.mu.Lock()
		p := d.pend[k]
		pregunta := ""
		if p != nil && p.Estado == Pendiente {
			pregunta = p.Pregunta
		}
		d.mu.Unlock()
		if pregunta == "" {
			continue
		}
		res, err := r.Preguntar(ctx, pregunta, rag.Opciones{K: 8, SinRegistro: true})
		resuelta := err == nil && !res.SinContexto && res.Modo == "respuesta"
		d.mu.Lock()
		rep.Revisadas++
		if p.Estado == Pendiente { // el admin pudo resolverla mientras tanto
			if resuelta {
				p.Estado, p.Respuesta, p.Resuelto = Resuelto, recortar(res.Respuesta, 800), ahora()
				rep.Resueltas++
			} else {
				p.Intentos++
			}
		}
		cp := *p
		_ = d.guardarPendientes()
		d.mu.Unlock()
		d.Bus.Publicar("repaso_avance", map[string]any{
			"i": i + 1, "total": len(claves), "pregunta": pregunta, "resuelta": resuelta, "pendiente": cp,
		})
	}
	d.mu.Lock()
	rep.Fin = ahora()
	d.ultimo, d.enCurso = rep, nil
	b, _ := json.Marshal(rep)
	_ = os.WriteFile(d.ruta("repaso.json"), b, 0o644)
	out := *rep
	d.mu.Unlock()
	d.Bus.Publicar("repaso_fin", out)
	return out, ctx.Err()
}

// ProgramarNoche corre Repasar cada día a la hora «HH:MM» (hora local del
// contenedor, TZ=America/Lima) hasta que ctx termine.
func (d *Diario) ProgramarNoche(ctx context.Context, r *rag.RAG, hora string, log func(string, ...any)) error {
	h, err := time.Parse("15:04", hora)
	if err != nil {
		return fmt.Errorf("RAG_HORA_APRENDER=%q: usa HH:MM", hora)
	}
	go func() {
		for {
			t := proxima(time.Now(), h.Hour(), h.Minute())
			log("aprendizaje: próximo repaso %s", t.Format("2006-01-02 15:04"))
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(t)):
			}
			rep, err := d.Repasar(ctx, r, "noche")
			if err != nil && !errors.Is(err, context.Canceled) {
				log("aprendizaje: %v", err)
			}
			log("aprendizaje: %d revisadas, %d resueltas", rep.Revisadas, rep.Resueltas)
		}
	}()
	return nil
}

func proxima(desde time.Time, hh, mm int) time.Time {
	t := time.Date(desde.Year(), desde.Month(), desde.Day(), hh, mm, 0, 0, desde.Location())
	if !t.After(desde) {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

// EstadoRepaso devuelve el repaso en curso (si hay) y el último terminado.
func (d *Diario) EstadoRepaso() (enCurso, ultimo *Repaso) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.enCurso != nil {
		c := *d.enCurso
		enCurso = &c
	}
	if d.ultimo != nil {
		u := *d.ultimo
		ultimo = &u
	}
	return
}
