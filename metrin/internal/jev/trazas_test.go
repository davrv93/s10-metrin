package jev

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// leerTrazas lee el JSONL del emisor y devuelve las trazas
// en orden de escritura.
func leerTrazas(t *testing.T, ruta string) []map[string]any {
	t.Helper()
	b, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatalf("leer %s: %v", ruta, err)
	}
	var trazas []map[string]any
	for _, linea := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if linea == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(linea), &m); err != nil {
			t.Fatalf("línea no es JSON: %v (%s)", err, linea)
		}
		trazas = append(trazas, m)
	}
	return trazas
}

func TestDecirGuardaTrazaEnJSONL(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "trazas.jsonl")
	c, _ := nuevoFalso(t, respuestaDos, http.StatusOK)
	c.Trazas = &RegistroTrazas{Ruta: ruta, Origen: "reportes"}

	estado := map[string]any{"mensaje": "roto"}
	if _, err := c.Decidir(context.Background(), estado, map[string]Pregunta{
		"departamento": {Tipo: "choice", Criterios: map[string]any{
			"sales": "New purchases", "support": "Problems with an order"}},
		"severidad": {Tipo: "score", Criterios: []any{"a", "b"}},
		"devolucion": {Tipo: "noul"},
	}); err != nil {
		t.Fatal(err)
	}

	trazas := leerTrazas(t, ruta)
	if len(trazas) != 1 {
		t.Fatalf("trazas = %d, quería 1", len(trazas))
	}
	tr := trazas[0]
	if id, _ := tr["id"].(string); len(id) != 26 {
		t.Fatalf("id = %q, quería ULID de 26 caracteres", id)
	}
	if tr["modelo"] != "local-model" {
		t.Fatalf("modelo = %v", tr["modelo"])
	}
	if tr["origen"] != "reportes" {
		t.Fatalf("origen = %v", tr["origen"])
	}
	if tr["ts"] == "" {
		t.Fatal("ts vacío")
	}
	if ms, _ := tr["ms"].(float64); ms < 0 {
		t.Fatalf("ms = %v", ms)
	}
	if tk, _ := tr["tokens_entrada"].(float64); tk != 120 {
		t.Fatalf("tokens_entrada = %v, quería 120", tk)
	}
	est, _ := tr["estado"].(map[string]any)
	if est["mensaje"] != "roto" {
		t.Fatalf("estado = %v", tr["estado"])
	}

	// Preguntas ordenadas por id: departamento, devolucion, severidad.
	ps, _ := tr["preguntas"].([]any)
	if len(ps) != 3 {
		t.Fatalf("preguntas = %d, quería 3", len(ps))
	}
	ids := make([]string, 3)
	for i, p := range ps {
		ids[i] = p.(map[string]any)["id"].(string)
	}
	if strings.Join(ids, ",") != "departamento,devolucion,severidad" {
		t.Fatalf("preguntas fuera de orden: %v", ids)
	}

	// Respuesta traducida al formato del visor (claves en español).
	dep := ps[0].(map[string]any)["respuesta"].(map[string]any)
	if dep["choice"] != "support" {
		t.Fatalf("choice = %v", dep["choice"])
	}
	prob, _ := dep["probabilidades"].(map[string]any)
	if prob["support"] != 0.8 {
		t.Fatalf("probabilidades = %v", dep["probabilidades"])
	}
	if dep["confianza"] != 0.72 {
		t.Fatalf("confianza = %v", dep["confianza"])
	}
	// Noul: solo trae tipo y noul.
	dev := ps[1].(map[string]any)["respuesta"].(map[string]any)
	if dev["noul"] != 0.91 || dev["tipo"] != "noul" {
		t.Fatalf("noul = %v", dev)
	}
	// correcta es null en trazas reales (aún sin evaluar).
	if v, ok := tr["correcta"]; !ok || v != nil {
		t.Fatalf("correcta = %v, quería null", tr["correcta"])
	}
}

func TestDecirSinTrazasNoEscribeArchivo(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "trazas.jsonl")
	c, _ := nuevoFalso(t, respuestaDos, http.StatusOK) // Trazas = nil
	if _, err := c.Decidir(context.Background(), "x", map[string]Pregunta{"q": {Tipo: "noul"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ruta); !os.IsNotExist(err) {
		t.Fatalf("el archivo no debió crearse: %v", err)
	}
}

func TestDecirGravaFalloDeEscrituraEnSilencio(t *testing.T) {
	// Ruta = un directorio: OpenFile falla siempre. La
	// decisión ya está tomada: no puede fallar la petición.
	c, _ := nuevoFalso(t, respuestaDos, http.StatusOK)
	c.Trazas = &RegistroTrazas{Ruta: t.TempDir(), Origen: "reportes"}
	rs, err := c.Decidir(context.Background(), "x", map[string]Pregunta{"devolucion": {Tipo: "noul"}})
	if err != nil {
		t.Fatalf("un fallo de escritura no puede fallar la decisión: %v", err)
	}
	if rs["devolucion"].Noul != 0.91 {
		t.Fatalf("respuesta perdida: %#v", rs)
	}
}

func TestGuardarAnadeTrazas(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "trazas.jsonl")
	r := &RegistroTrazas{Ruta: ruta, Origen: "o"}
	for i := 0; i < 2; i++ {
		if err := r.guardar(nuevaTraza("m", "o", "e",
			map[string]Pregunta{"q": {Tipo: "noul"}},
			map[string]Respuesta{"q": {Tipo: "noul", Noul: 0.5}},
			10, 5)); err != nil {
			t.Fatal(err)
		}
	}
	trazas := leerTrazas(t, ruta)
	if len(trazas) != 2 {
		t.Fatalf("líneas = %d, quería 2", len(trazas))
	}
	if trazas[0]["id"] == trazas[1]["id"] {
		t.Fatal("ids repetidos")
	}
}

func TestRespuestaATrazaOmiteConfianzaCero(t *testing.T) {
	// Confianza 0 = distribución uniforme: el visor la muestra
	// como «sin confianza», igual que ausente.
	rt := respuestaATraza(Respuesta{Tipo: "choice", Choice: "a"})
	if rt.Confianza != nil {
		t.Fatalf("confianza = %v, quería nil", *rt.Confianza)
	}
	rt = respuestaATraza(Respuesta{Tipo: "choice", Confianza: 0.9})
	if rt.Confianza == nil || *rt.Confianza != 0.9 {
		t.Fatalf("confianza = %v, quería 0.9", rt.Confianza)
	}
}

func TestNuevoULID(t *testing.T) {
	vistos := map[string]bool{}
	for i := 0; i < 100; i++ {
		id := nuevoULID()
		if len(id) != 26 {
			t.Fatalf("id = %q (%d caracteres)", id, len(id))
		}
		for _, r := range id {
			if !strings.ContainsRune(alfabetoULID, r) {
				t.Fatalf("carácter %q fuera del alfabeto ULID", r)
			}
		}
		if vistos[id] {
			t.Fatalf("id repetido: %q", id)
		}
		vistos[id] = true
	}
}
