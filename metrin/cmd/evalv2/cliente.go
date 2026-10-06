package main

// Cliente de /ask que reproduce a la página (internal/servidor/pagina.html, función `preguntar`):
//
//   memoria.push({rol:'usuario', texto:q})                      ← ANTES de enviar
//   cuerpo = {pregunta:q, source:'s10-kb', k:8, hilo: memoria.slice(0,-1).slice(-6), traza:true}
//            (hilo = solo los turnos ANTERIORES: desde el 06-10-2026 la pregunta actual ya no va en el hilo)
//   memoria.push({rol:'asistente', texto: respuesta.slice(0,300)}); memoria = memoria.slice(-6)
//   si falla: se quita el turno del usuario
//
// Además manda `version` ("v1" | "v2") y, en V2, la `memoria` estructurada que devolvió el turno anterior
// (tipos.Memoria: «viaja en el hilo, la guarda la página y la devuelve»). Supuesto documentado: la V2 la
// devuelve en la clave `memoria` de la respuesta y la recibe en la clave `memoria` de la petición.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"rag-go/internal/v2/tipos"
)

type turnoHilo struct {
	Rol   string `json:"rol"`
	Texto string `json:"texto"`
}

type peticionAsk struct {
	Pregunta string          `json:"pregunta"`
	Source   string          `json:"source,omitempty"`
	K        int             `json:"k,omitempty"`
	Hilo     []turnoHilo     `json:"hilo"`
	Traza    bool            `json:"traza"`
	Version  string          `json:"version"`
	Memoria  json.RawMessage `json:"memoria,omitempty"`
}

type fuenteAsk struct {
	ID        string   `json:"id,omitempty"`
	Manual    string   `json:"manual,omitempty"`
	Cita      string   `json:"cita"`
	Documento string   `json:"documento"`
	Pagina    int      `json:"pagina"`
	URL       string   `json:"url"`
	Fotos     []string `json:"fotos"`
	Distancia float64  `json:"distancia"`
}

type etapaAsk struct {
	ID     string         `json:"id"`
	Tipo   string         `json:"tipo"`
	Estado string         `json:"estado"`
	Ms     *float64       `json:"ms"`
	Razon  string         `json:"razon"`
	Datos  map[string]any `json:"datos"`
}

type trazaAsk struct {
	Version  int        `json:"version"`
	TotalMs  float64    `json:"total_ms"`
	Truncada bool       `json:"truncada"`
	Etapas   []etapaAsk `json:"etapas"`
}

func (t *trazaAsk) Etapa(id string) *etapaAsk {
	if t == nil {
		return nil
	}
	for i := range t.Etapas {
		if t.Etapas[i].ID == id {
			return &t.Etapas[i]
		}
	}
	return nil
}

type respuestaAsk struct {
	Pregunta     string `json:"pregunta"`
	Respuesta    string `json:"respuesta"`
	Modo         string `json:"modo"`
	Orquestacion struct {
		Intencion    string `json:"intencion"`
		TipoConsulta string `json:"tipo_consulta"`
		Ruta         string `json:"ruta"`
	} `json:"orquestacion"`
	Fuentes     []fuenteAsk     `json:"fuentes"`
	SinContexto bool            `json:"sin_contexto"`
	Motivo      string          `json:"motivo"`
	Reescrita   string          `json:"pregunta_reescrita"`
	Plan        *tipos.Plan     `json:"plan"`
	Memoria     json.RawMessage `json:"memoria"`
	Traza       *trazaAsk       `json:"traza"`
	Error       string          `json:"error"`
}

// TurnoCrudo es lo que se guarda en el JSON de resultados por cada turno enviado.
type TurnoCrudo struct {
	Pregunta  string          `json:"pregunta"`
	Peticion  peticionAsk     `json:"peticion"`
	Status    int             `json:"status"`
	ErrorRed  string          `json:"error_red,omitempty"`
	MsCliente float64         `json:"ms_cliente"`
	Respuesta *respuestaAsk   `json:"-"`
	Crudo     json.RawMessage `json:"respuesta"`
}

type cliente struct {
	url     string
	http    *http.Client
	source  string
	k       int
	traza   bool
	memoria []turnoHilo
	memV2   json.RawMessage
}

func nuevoCliente(url string, timeout time.Duration, source string, k int) *cliente {
	return &cliente{url: strings.TrimRight(url, "/"), http: &http.Client{Timeout: timeout}, source: source, k: k, traza: true}
}

// nuevaConversacion: lo que hace el botón «Nueva conversación» (memoria = []).
func (c *cliente) nuevaConversacion() { c.memoria = nil; c.memV2 = nil }

func (c *cliente) preguntar(ctx context.Context, version, q string) TurnoCrudo {
	hilo := c.memoria // turnos anteriores; la pregunta actual va solo en «pregunta» (como la página)
	if len(hilo) > 6 {
		hilo = hilo[len(hilo)-6:]
	}
	c.memoria = append(c.memoria, turnoHilo{Rol: "usuario", Texto: q})
	p := peticionAsk{Pregunta: q, Source: c.source, K: c.k, Hilo: append([]turnoHilo{}, hilo...), Traza: c.traza, Version: version}
	if version == "v2" && len(c.memV2) > 0 && string(c.memV2) != "null" {
		p.Memoria = c.memV2
	}
	t := TurnoCrudo{Pregunta: q, Peticion: p}
	cuerpo, _ := json.Marshal(p)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/ask", bytes.NewReader(cuerpo))
	if err != nil {
		t.ErrorRed = err.Error()
		c.memoria = c.memoria[:len(c.memoria)-1]
		return t
	}
	req.Header.Set("Content-Type", "application/json")
	t0 := time.Now()
	res, err := c.http.Do(req)
	if err != nil {
		t.MsCliente = msDesde(t0)
		t.ErrorRed = err.Error()
		c.memoria = c.memoria[:len(c.memoria)-1]
		return t
	}
	b, _ := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	res.Body.Close()
	t.MsCliente = msDesde(t0)
	t.Status = res.StatusCode
	t.Crudo = json.RawMessage(b)
	if !json.Valid(b) {
		t.Crudo, _ = json.Marshal(string(b))
	}
	var r respuestaAsk
	if err := json.Unmarshal(b, &r); err != nil {
		t.ErrorRed = "respuesta no es JSON: " + err.Error()
	}
	t.Respuesta = &r
	if res.StatusCode != http.StatusOK {
		if t.ErrorRed == "" {
			t.ErrorRed = fmt.Sprintf("HTTP %d: %s", res.StatusCode, recortarTexto(r.Error, 120))
		}
		c.memoria = c.memoria[:len(c.memoria)-1]
		return t
	}
	resp := r.Respuesta
	if resp == "" {
		resp = "No recibí una respuesta del servicio."
	}
	c.memoria = append(c.memoria, turnoHilo{Rol: "asistente", Texto: recortarRunas(resp, 300)})
	if len(c.memoria) > 6 {
		c.memoria = c.memoria[len(c.memoria)-6:]
	}
	if len(r.Memoria) > 0 {
		c.memV2 = r.Memoria
	}
	return t
}

func msDesde(t time.Time) float64 { return float64(time.Since(t).Microseconds()) / 1000 }

func recortarRunas(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// ---------------------------------------------------------------------------------------------
// RAM y CPU: `docker stats --no-stream` del contenedor objetivo (solo lectura)

type MuestraRecursos struct {
	En     time.Time `json:"en"`
	CPU    float64   `json:"cpu_pct"`
	MemMiB float64   `json:"mem_mib"`
}

type muestreador struct {
	contenedor string
	mu         sync.Mutex
	muestras   []MuestraRecursos
	errores    int
	parar      chan struct{}
	listo      chan struct{}
}

func iniciarMuestreo(contenedor string, cada time.Duration) *muestreador {
	m := &muestreador{contenedor: contenedor, parar: make(chan struct{}), listo: make(chan struct{})}
	if contenedor == "" {
		close(m.listo)
		return m
	}
	go func() {
		defer close(m.listo)
		for {
			if s, err := leerDockerStats(contenedor); err == nil {
				m.mu.Lock()
				m.muestras = append(m.muestras, s)
				m.mu.Unlock()
			} else {
				m.mu.Lock()
				m.errores++
				m.mu.Unlock()
			}
			select {
			case <-m.parar:
				return
			case <-time.After(cada):
			}
		}
	}()
	return m
}

func (m *muestreador) detener() []MuestraRecursos {
	if m.contenedor != "" {
		close(m.parar)
	}
	<-m.listo
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MuestraRecursos(nil), m.muestras...)
}

func leerDockerStats(contenedor string) (MuestraRecursos, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", "{{json .}}", contenedor).Output()
	if err != nil {
		return MuestraRecursos{}, err
	}
	var d struct {
		CPUPerc  string `json:"CPUPerc"`
		MemUsage string `json:"MemUsage"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out), &d); err != nil {
		return MuestraRecursos{}, err
	}
	cpu, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(d.CPUPerc), "%"), 64)
	usado, _, _ := strings.Cut(d.MemUsage, "/")
	return MuestraRecursos{En: time.Now(), CPU: cpu, MemMiB: aMiB(strings.TrimSpace(usado))}, nil
}

func aMiB(s string) float64 {
	unidades := []struct {
		suf string
		f   float64
	}{{"GiB", 1024}, {"MiB", 1}, {"KiB", 1.0 / 1024}, {"GB", 1e9 / (1 << 20)}, {"MB", 1e6 / (1 << 20)}, {"kB", 1e3 / (1 << 20)}, {"B", 1.0 / (1 << 20)}}
	for _, u := range unidades {
		if strings.HasSuffix(s, u.suf) {
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(s, u.suf)), 64)
			if err != nil {
				return 0
			}
			return v * u.f
		}
	}
	return 0
}
