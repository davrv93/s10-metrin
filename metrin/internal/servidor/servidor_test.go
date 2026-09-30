package servidor

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rag-go/internal/rag"
)

func TestFotosUsaDirectorioDeMediosSeparadoDelIndice(t *testing.T) {
	mediaDir := t.TempDir()
	indexDir := t.TempDir()
	imagePath := filepath.Join(mediaDir, "imagenes", "manual", "captura.png")
	if err := os.MkdirAll(filepath.Dir(imagePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(imagePath, []byte("image-data"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RAG_MEDIA_DATOS", mediaDir)
	t.Setenv("RAG_DATOS", indexDir)

	h := Nuevo(nil, time.Second)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/fotos/imagenes/manual/captura.png", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("GET foto = %d, quería 200", rr.Code)
	}
	if got := rr.Body.String(); got != "image-data" {
		t.Fatalf("contenido de foto inesperado: %q", got)
	}
}

func TestVinculaPaginaDePDFCitada(t *testing.T) {
	dataDir := t.TempDir()
	pdfDir := filepath.Join(dataDir, "pdf")
	if err := os.MkdirAll(pdfDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pdfDir, "guia-de-usuario-de-s10-presupuestos.pdf"), []byte("pdf"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RAG_MEDIA_DATOS", dataDir)
	fuentes := []rag.Fuente{{
		URL: "manual://Guia de Usuario de S10 Presupuestos.pdf", Documento: "Guia de Usuario de S10 Presupuestos", Pagina: 67,
	}}
	vincularFotos(fuentes)
	if len(fuentes[0].Fotos) != 1 || fuentes[0].Fotos[0] != "manual/guia-de-usuario-de-s10-presupuestos/67" {
		t.Fatalf("vista de página no vinculada: %#v", fuentes[0].Fotos)
	}
}

func TestSirveVistaPreviaDePaginaPDF(t *testing.T) {
	dataDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dataDir, "pdf"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "pdf", "guia.pdf"), []byte("pdf"), 0o644); err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	renderer := filepath.Join(binDir, "pdftoppm")
	if err := os.WriteFile(renderer, []byte("#!/bin/sh\nfor arg do prefix=$arg; done\nprintf preview > \"$prefix.jpg\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RAG_MEDIA_DATOS", dataDir)
	t.Setenv("RAG_PREVIEWS", filepath.Join(t.TempDir(), "cache"))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	h := Nuevo(nil, time.Second)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/fotos/manual/guia/67", nil))
	if rr.Code != 200 || rr.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("respuesta de preview = %d (%s)", rr.Code, rr.Header().Get("Content-Type"))
	}
	body, err := io.ReadAll(rr.Result().Body)
	if err != nil || !strings.EqualFold(string(body), "preview") {
		t.Fatalf("contenido de preview inesperado: %q, %v", body, err)
	}
}
