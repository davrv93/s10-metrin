package rag

import (
	"context"
	"strings"

	"rag-go/internal/almacen"
	"rag-go/internal/clasificar"
)

// Preguntas validadas por JEV (semillas). Cada una es una tarjeta del índice
// que se busca por la pregunta y trae la sección del manual con sus pasos y
// capturas (semillas_a_kb.py).

// distanciaSemilla (RAG_DISTANCIA_SEMILLA): hasta aquí una pregunta escrita se considera la misma que
// una validada y adopta su tipo de consulta. Las tarjetas se embeben con la
// pregunta sola, así que una paráfrasis cercana queda por debajo. Valor
// inicial sin calibrar: medirlo con el embebedor real y el banco de semillas.
const distanciaSemilla = 0.15

// semillaElegida trae la tarjeta de la semilla que el usuario eligió en el
// menú. Un id desconocido no rompe la consulta: sigue el flujo normal.
func (r *RAG) semillaElegida(ctx context.Context, pregunta, id string) *almacen.Resultado {
	id = strings.TrimSpace(id)
	if id == "" || r.Almacen == nil {
		return nil
	}
	rs, err := r.Almacen.Buscar(ctx, pregunta, 1, map[string]string{"semilla": id})
	if err != nil || len(rs) == 0 {
		return nil
	}
	e := rs[0]
	e.Distancia = 0
	return &e
}

func planSemilla(e almacen.Resultado, pregunta string) Orquestacion {
	tipo := e.Metadata["semilla_tipo"]
	if tipo == "" {
		tipo = tipoConsulta(pregunta, nil)
	}
	return Orquestacion{
		Intencion:    clasificar.Trabajo,
		TipoConsulta: tipo,
		Ruta:         rutaRAG,
		Clasificador: "semilla_validada",
		Semilla:      e.Metadata["semilla"],
	}
}

// semillaReconocida: la mejor evidencia es una pregunta validada casi igual a
// la escrita.
func semillaReconocida(seleccionados []almacen.Resultado, umbral float64) *almacen.Resultado {
	if umbral <= 0 {
		umbral = distanciaSemilla
	}
	if len(seleccionados) == 0 {
		return nil
	}
	s := seleccionados[0]
	if s.Metadata["semilla"] == "" || s.Distancia > umbral {
		return nil
	}
	return &s
}
