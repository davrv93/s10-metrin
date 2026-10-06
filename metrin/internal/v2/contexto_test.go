package v2

import (
	"encoding/json"
	"strings"
	"testing"

	"rag-go/internal/rag"
	"rag-go/internal/v2/tipos"
)

func TestReescrituraNoPierdeTerminosExactos(t *testing.T) {
	pregunta := "¿Con F7 se abre el metrado del APU 01.02.03? ¿Y con Ctrl + F?"
	c := Reescribir(pregunta, nil)
	if c.Original != pregunta {
		t.Fatalf("la original no se toca: %q", c.Original)
	}
	for _, w := range []string{"f7", "metrado", "apu", "01.02.03", "ctrl", "con"} {
		if !strings.Contains(" "+c.Normalizada+" ", " "+w) {
			t.Errorf("la normalizada perdió %q: %q", w, c.Normalizada)
		}
	}
	for _, w := range []string{"F7", "APU", "01.02.03", "Ctrl + F"} {
		found := false
		for _, e := range c.Entidades {
			found = found || e == w
		}
		if !found {
			t.Errorf("falta la entidad exacta %q en %v", w, c.Entidades)
		}
	}
	// La compacta de la expansión tampoco pierde los términos.
	x := Expandir(c, aliasFalso{})
	for _, w := range []string{"f7", "metrado", "apu"} {
		if !strings.Contains(x.Normalizada, w) {
			t.Errorf("la compacta perdió %q: %q", w, x.Normalizada)
		}
	}
	if x.Original != pregunta || strings.Contains(" "+x.Normalizada+" ", " con ") {
		t.Errorf("expandir: %+v", x)
	}
}

func TestNormalizar(t *testing.T) {
	if got := Normalizar("¿Cómo   AÑADO el Metrado, «año» 3.5?"); got != "como añado el metrado año 3.5" {
		t.Fatalf("%q", got)
	}
}

func TestEstadoSeparaHechosDeInferencias(t *testing.T) {
	// «metrado» lo escribe el usuario (hecho); el módulo Presupuestos lo deduce el Aliaser (inferencia).
	e := ConstruirEstado("¿Cómo registro un metrado?", nil, tipos.Memoria{}, origenMemoria, aliasFalso{})
	if d, ok := e.Hechos[prefijoEntidad+"Metrado"]; !ok || d.Origen != origenUsuario || d.Confianza != 1 {
		t.Errorf("la entidad escrita es un hecho: %+v", e.Hechos)
	}
	if _, ok := e.Hechos[claveModulo]; ok {
		t.Errorf("el módulo deducido por alias NO es un hecho: %+v", e.Hechos)
	}
	if d, ok := e.Inferencias[claveModulo]; !ok || d.Valor != "Presupuestos" || d.Confianza >= 1 {
		t.Errorf("el módulo deducido va a inferencias: %+v", e.Inferencias)
	}
	// Si el usuario nombra el módulo, pasa a hecho.
	e2 := ConstruirEstado("¿Cómo registro un metrado en presupuestos?", nil, tipos.Memoria{}, origenMemoria, aliasFalso{})
	if d, ok := e2.Hechos[claveModulo]; !ok || d.Origen != origenUsuario {
		t.Errorf("módulo escrito por el usuario = hecho: %+v", e2.Hechos)
	}
	// Memoria: la que manda el cliente es un hecho; la reconstruida del hilo, una inferencia.
	mem := tipos.Memoria{ProcedimientoID: "metrados.registrar", PasoActual: 2}
	cli := ConstruirEstado("listo", nil, mem, origenMemoria, nil)
	if _, ok := cli.Hechos[claveProcCurso]; !ok {
		t.Errorf("memoria del cliente = hecho: %+v", cli.Hechos)
	}
	hilo := ConstruirEstado("listo", nil, mem, origenHilo, nil)
	if _, ok := hilo.Hechos[claveProcCurso]; ok {
		t.Errorf("memoria reconstruida NO es un hecho: %+v", hilo.Hechos)
	}
	if d := hilo.Inferencias[claveProcCurso]; d.Origen != origenHilo || d.Confianza >= 1 {
		t.Errorf("memoria reconstruida = inferencia del hilo: %+v", hilo.Inferencias)
	}
	for _, est := range []tipos.Estado{e, e2, cli, hilo} {
		for k, d := range est.Inferencias {
			if _, dup := est.Hechos[k]; dup {
				t.Errorf("%q está a la vez en hechos e inferencias", k)
			}
			if d.Origen == origenUsuario {
				t.Errorf("una inferencia no puede venir del usuario: %s %+v", k, d)
			}
		}
	}
	if len(e.Desconocidos) == 0 || e.Desconocidos[0] != "procedimiento" {
		t.Errorf("desconocidos: %v", e.Desconocidos)
	}
}

func TestEstadoHiloRecortadoYReferencia(t *testing.T) {
	var hilo []rag.Turno
	for i := 0; i < 6; i++ {
		hilo = append(hilo, rag.Turno{Rol: "usuario", Texto: strings.Repeat("metrado ", 60)}, rag.Turno{Rol: "asistente", Texto: "ok"})
	}
	hilo = append(hilo, rag.Turno{Rol: "sistema", Texto: "x"})
	e := ConstruirEstado("¿y eso dónde está?", hilo, tipos.Memoria{}, origenMemoria, aliasFalso{})
	if len(e.Hilo) != maxTurnosHilo {
		t.Fatalf("hilo recortado a %d turnos: %d", maxTurnosHilo, len(e.Hilo))
	}
	for _, x := range e.Hilo {
		if len([]rune(x.Texto)) > maxRunasTurno || x.Rol == "sistema" {
			t.Fatalf("turno sin recortar: %+v", x)
		}
	}
	if d, ok := e.Inferencias[claveReferencia]; !ok || d.Origen != origenHilo {
		t.Fatalf("«eso» se infiere del hilo: %+v", e.Inferencias)
	}
	if len(e.Consulta.Aliases) == 0 || e.Consulta.Original != "¿y eso dónde está?" {
		t.Fatalf("la referencia amplía con alias sin tocar la original: %+v", e.Consulta)
	}
	// Viaja al motor de decisión: se serializa sin problema y es compacto.
	b, err := json.Marshal(e)
	if err != nil || len(b) > 4096 {
		t.Fatalf("estado no compacto: %d bytes, %v", len(b), err)
	}
}
