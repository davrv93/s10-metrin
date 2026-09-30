package aprender

import (
	"sort"
	"time"
)

// Dia agrega las interacciones de una fecha.
type Dia struct {
	Fecha       string `json:"fecha"`
	Total       int    `json:"total"`
	Respondidas int    `json:"respondidas"`
	SinContexto int    `json:"sin_contexto"`
	Positivos   int    `json:"positivos"`
	Negativos   int    `json:"negativos"`
}

// Metricas dice qué tan bien le va al agente.
type Metricas struct {
	Desde          string `json:"desde"`
	Total          int    `json:"total"`
	Consultas      int    `json:"consultas"` // de trabajo (sin charla)
	Respondidas    int    `json:"respondidas"`
	SinContexto    int    `json:"sin_contexto"`
	Conversacional int    `json:"conversacional"`
	NoDisponible   int    `json:"modelo_no_disponible"`
	// TasaRespuesta: respondidas con fuentes / consultas de trabajo.
	TasaRespuesta float64 `json:"tasa_respuesta"`
	Positivos     int     `json:"positivos"`
	Negativos     int     `json:"negativos"`
	// Satisfaccion: 👍 / (👍 + 👎). TasaVoto: interacciones con voto / total.
	Satisfaccion float64 `json:"satisfaccion"`
	TasaVoto     float64 `json:"tasa_voto"`
	// ConSugerencias: sin contexto a las que se les ofreció reformular.
	ConSugerencias int     `json:"con_sugerencias"`
	LatenciaMedia  int64   `json:"latencia_media_ms"`
	LatenciaP90    int64   `json:"latencia_p90_ms"`
	DistanciaMedia float64 `json:"distancia_media"`
	Cola           struct {
		Pendientes  int `json:"pendientes"`
		Resueltas   int `json:"resueltas"`
		Aprendidas  int `json:"aprendidas"`
		Descartadas int `json:"descartadas"`
	} `json:"cola"`
	PorDia  []Dia   `json:"por_dia"`
	Repaso  *Repaso `json:"ultimo_repaso,omitempty"`
	EnCurso *Repaso `json:"repaso_en_curso,omitempty"`
}

// Metricas calcula los indicadores de los últimos `dias` días (0 = todo).
func (d *Diario) Metricas(dias int) Metricas {
	var m Metricas
	var desde time.Time
	if dias > 0 {
		desde = time.Now().AddDate(0, 0, -dias+1)
		desde = time.Date(desde.Year(), desde.Month(), desde.Day(), 0, 0, 0, 0, desde.Location())
		m.Desde = desde.Format("2006-01-02")
	}
	d.mu.Lock()
	porDia := map[string]*Dia{}
	var lat []int64
	var sumaDist float64
	nDist := 0
	for _, id := range d.orden {
		in := d.inter[id]
		t, err := time.Parse(time.RFC3339, in.Fecha)
		if err != nil || (dias > 0 && t.Before(desde)) {
			continue
		}
		dia := t.Local().Format("2006-01-02")
		dd := porDia[dia]
		if dd == nil {
			dd = &Dia{Fecha: dia}
			porDia[dia] = dd
		}
		m.Total++
		dd.Total++
		switch {
		case in.Modo == "conversacional":
			m.Conversacional++
		case in.SinContexto:
			m.SinContexto++
			dd.SinContexto++
			if in.Sugerencias > 0 {
				m.ConSugerencias++
			}
		case in.Modo == "modelo_no_disponible":
			m.NoDisponible++
		default:
			m.Respondidas++
			dd.Respondidas++
		}
		if in.Modo != "conversacional" {
			lat = append(lat, in.MsBusqueda+in.MsLLM)
			if in.DistanciaMin > 0 && in.DistanciaMin < 1 {
				sumaDist += in.DistanciaMin
				nDist++
			}
		}
		switch in.Voto {
		case 1:
			m.Positivos++
			dd.Positivos++
		case -1:
			m.Negativos++
			dd.Negativos++
		}
	}
	for _, p := range d.pend {
		switch p.Estado {
		case Pendiente:
			m.Cola.Pendientes++
		case Resuelto:
			m.Cola.Resueltas++
		case Aprendido:
			m.Cola.Aprendidas++
		case Descartado:
			m.Cola.Descartadas++
		}
	}
	if d.ultimo != nil {
		u := *d.ultimo
		m.Repaso = &u
	}
	if d.enCurso != nil {
		c := *d.enCurso
		m.EnCurso = &c
	}
	d.mu.Unlock()

	m.Consultas = m.Total - m.Conversacional
	m.TasaRespuesta = razon(m.Respondidas, m.Consultas)
	m.Satisfaccion = razon(m.Positivos, m.Positivos+m.Negativos)
	m.TasaVoto = razon(m.Positivos+m.Negativos, m.Total)
	if nDist > 0 {
		m.DistanciaMedia = float64(int(sumaDist/float64(nDist)*1000+0.5)) / 1000
	}
	if len(lat) > 0 {
		sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
		var s int64
		for _, x := range lat {
			s += x
		}
		m.LatenciaMedia = s / int64(len(lat))
		m.LatenciaP90 = lat[(len(lat)*9)/10]
	}
	m.PorDia = make([]Dia, 0, len(porDia))
	for _, dd := range porDia {
		m.PorDia = append(m.PorDia, *dd)
	}
	sort.Slice(m.PorDia, func(i, j int) bool { return m.PorDia[i].Fecha < m.PorDia[j].Fecha })
	return m
}

func razon(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(int(float64(a)/float64(b)*1000+0.5)) / 1000
}
