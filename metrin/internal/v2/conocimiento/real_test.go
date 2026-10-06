package conocimiento

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"rag-go/internal/v2/tipos"
)

// Pruebas contra los YAML REALES de kb/ (los generan otros agentes). Se saltan si kb/ no está.

const dirKBReal = "../../../../kb"

var (
	baseRealOnce sync.Once
	baseReal     *Base
	recRealOnce  sync.Once
	recReal      *Recuperador
)

func recuperadorReal(t *testing.T) (*Base, *Recuperador) {
	b := cargarReal(t)
	recRealOnce.Do(func() { recReal = NuevoRecuperador(b) })
	return b, recReal
}

func cargarReal(t *testing.T) *Base {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dirKBReal, "fragmentos.jsonl")); err != nil {
		t.Skip("sin kb/ real")
	}
	baseRealOnce.Do(func() {
		datos := ""
		if _, err := os.Stat("../../../../data/imagenes"); err == nil {
			datos = "../../../../data"
		}
		baseReal = Cargar(Opciones{DirKB: dirKBReal, ArchivoPlantillas: "../../../plantillas/respuestas.yml", DirDatos: datos})
	})
	return baseReal
}

func TestReal_TodoCargaYCadaCitaResuelve(t *testing.T) {
	b := cargarReal(t)
	t.Log(b.Resumen())
	for _, e := range b.Errores {
		t.Log(e.Error())
	}
	if len(b.Procedimientos) == 0 {
		t.Fatal("no cargó ningún procedimiento real")
	}
	for _, e := range b.Graves() {
		if strings.Contains(e.Archivo, "procedimientos") || strings.Contains(e.Archivo, "conceptos") || strings.Contains(e.Archivo, "plantillas") {
			t.Errorf("YAML real inválido: %s", e.Error())
		}
	}
	// Cada cita de cada procedimiento resuelve con su par (id, manual) contra el índice.
	citas := 0
	for _, p := range b.Procedimientos {
		p.citas(func(donde, id string, linea int) {
			citas++
			c, err := b.CitaProcedimiento(p, id)
			if err != nil {
				t.Errorf("%s:%d %s: %v", p.Archivo, linea, donde, err)
				return
			}
			if c.Fragmento == nil || c.Fragmento.Manual != c.Manual || c.Fragmento.ID != id {
				t.Errorf("%s %s: la cita %s no quedó ligada a su fragmento (%s)", p.ID, donde, id, c.Manual)
			}
		})
		// Cada foto la asocia una fuente de su paso (y existe en data/ si está).
		p.recorrerFotos(func(s *tipos.Paso, f tipos.Foto) {
			if !b.fuenteAsociaFoto(p, s, f.Ruta) {
				t.Errorf("%s: la foto %s no la asocia ninguna fuente del paso %s", p.ID, f.ID, s.ID)
			}
		})
	}
	for _, e := range b.Errores {
		if strings.Contains(e.Mensaje, "no existe imagenes/") || strings.Contains(e.Mensaje, ": no existe ") {
			t.Errorf("foto ausente en data/: %s", e.Error())
		}
	}
	t.Logf("%d procedimientos, %d citas resueltas", len(b.Procedimientos), citas)
	// Conceptos: cada cita resuelve (el manual sale de la tabla, de {id, manual} o del comentario).
	sinResolver := 0
	for _, c := range b.Conceptos {
		for _, id := range c.Fuente {
			if _, err := b.CitaConcepto(c, id); err != nil {
				sinResolver++
				t.Errorf("concepto %s (%s): %v", c.ID, c.Archivo, err)
			}
		}
	}
	t.Logf("%d conceptos, %d citas de conceptos sin resolver", len(b.Conceptos), sinResolver)
}
