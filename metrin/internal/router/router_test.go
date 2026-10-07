package router

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"testing"
)

// Modelo real: lo exporta entrenamiento-router/entrenar.py y se copia a modelos/router/ (el .pjge va en git por
// excepción en .gitignore, como potion-es-int8 en AgenteS10).
const (
	rutaCabeza  = "../../modelos/router/router-s10.cabeza.json"
	rutaParidad = "../../modelos/router/router-s10.paridad.jsonl"
)

func cargarReal(t *testing.T) *Router {
	t.Helper()
	for _, r := range []string{rutaCabeza, "../../modelos/router/router-s10.pjge"} {
		if _, err := os.Stat(r); err != nil {
			t.Skip("falta el router (entrenamiento-router/README.md: entrenar y copiar a modelos/router/)")
		}
	}
	r, err := Cargar(rutaCabeza)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// Go y Python deben dar el mismo embedding y las mismas probabilidades (int8 decuantizado en los dos).
func TestParidadConPython(t *testing.T) {
	r := cargarReal(t)
	f, err := os.Open(rutaParidad)
	if err != nil {
		t.Skip("falta router-s10.paridad.jsonl")
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<22)
	n := 0
	for sc.Scan() {
		var x struct {
			Texto string    `json:"texto"`
			Emb   []float64 `json:"emb"`
			P     []float64 `json:"p"`
		}
		if err := json.Unmarshal(sc.Bytes(), &x); err != nil {
			t.Fatal(err)
		}
		v := r.emb.Embeber(x.Texto)
		for j := range v {
			if math.Abs(float64(v[j])-x.Emb[j]) > 2e-5 {
				t.Fatalf("%q: embedding[%d] Go %.6f, Python %.6f", x.Texto, j, v[j], x.Emb[j])
			}
		}
		p := r.Probabilidades(x.Texto)
		for j := range p {
			if math.Abs(p[j]-x.P[j]) > 1e-4 {
				t.Fatalf("%q: p[%d] Go %.6f, Python %.6f", x.Texto, j, p[j], x.P[j])
			}
		}
		n++
	}
	if n == 0 {
		t.Fatal("paridad vacía")
	}
}

func TestDecidirCoherente(t *testing.T) {
	r := cargarReal(t)
	for _, q := range []string{"como ingreso los metrados de una partida", "hola buenas tardes", "kardex de un material"} {
		d := r.Decidir(q)
		switch d.Accion {
		case Elegir:
			if len(d.Candidatos) != 1 || d.Candidatos[0].Probabilidad < r.U.Acepta {
				t.Fatalf("%q: elegir sin cumplir el umbral: %+v", q, d)
			}
		case AclararCaso, AclararModulo:
			if len(d.Candidatos) < 2 && d.Accion == AclararCaso {
				t.Fatalf("%q: aclarar con menos de dos opciones: %+v", q, d)
			}
		case Delegar:
		default:
			t.Fatalf("%q: acción desconocida %q", q, d.Accion)
		}
		if len(d.Top) == 0 {
			t.Fatalf("%q: sin top", q)
		}
	}
}

func TestCargarRechazaHuellaAjena(t *testing.T) {
	cargarReal(t)
	dir := t.TempDir()
	raw, _ := os.ReadFile(rutaCabeza)
	var c map[string]any
	_ = json.Unmarshal(raw, &c)
	c["sha256_pjge"] = "0000"
	c["pjge"] = "../../../" + "modelos/router/router-s10.pjge"
	b, _ := json.Marshal(c)
	ruta := dir + "/cabeza.json"
	_ = os.WriteFile(ruta, b, 0o644)
	if _, err := Cargar(ruta); err == nil {
		t.Fatal("cargó una cabeza cuyo .pjge no coincide")
	}
}
