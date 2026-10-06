package traza

import "fmt"

// Reglas v1 de oportunidades (contrato, «Oportunidades»). Son fijas y
// deterministas: leen los hechos que dejó cada etapa, nunca preguntan a un
// modelo. Una regla que se cumple sube la etapa de «ok» a su nivel (si ya
// estaba peor, la deja) y añade una oportunidad numerada en orden de etapas.

func aplicarReglas(etapas []Etapa) []Oportunidad {
	out := []Oportunidad{}
	pos := func(id string) *Etapa { return &etapas[indice[id]] }
	// razon: la explicación corta que queda en la tarjeta cuando la regla
	// cambia su estado ("" = conservar la que ya traía la etapa).
	agregar := func(e *Etapa, nivel, razon, texto string) {
		if e.Estado == EstadoOK || (e.Estado == EstadoAlerta && nivel == EstadoError) {
			e.Estado = nivel
			if razon != "" {
				e.Razon = Recortar(razon)
			}
		}
		if e.Razon == "" {
			e.Razon = Recortar(texto)
		}
		out = append(out, Oportunidad{N: len(out) + 1, Etapa: e.ID, Nivel: nivel, Texto: texto})
	}
	ruta, _ := pos(EtapaRuta).Datos["ruta"].(string)

	for i := range etapas {
		e := &etapas[i]
		switch e.ID {
		case EtapaClasificador:
			if !ejecutada(e) || ruta != "rag" {
				continue
			}
			sim, haySim := e.Datos["similitud"].(float64)
			umbral, _ := e.Datos["umbral"].(float64)
			if paso, ok := e.Datos["paso_umbral"].(bool); ok && !paso && haySim {
				agregar(e, EstadoAlerta, "bajo el umbral, fue a la ruta segura", fmt.Sprintf(
					"Similitud %s bajo el umbral %s: fue a la ruta segura (rag). Si el mensaje era charla, al clasificador le faltan ejemplos parecidos.",
					Decimal(sim, 3), Decimal(umbral, 2)))
			}
			intencion, _ := e.Datos["intencion"].(string)
			if intencion == "social" || intencion == "limite" {
				agregar(e, EstadoAlerta, "etiqueta «"+intencion+"» en una pregunta que siguió por la ruta rag", fmt.Sprintf(
					"Etiqueta conversacional «%s» (similitud %s) en una pregunta de trabajo que siguió por rag: revisar los ejemplos de esa intención.",
					intencion, Decimal(sim, 3)))
			}
		case EtapaSeleccion:
			if !ejecutada(e) {
				continue
			}
			fr, _ := e.Datos["fragmentos"].([]Fragmento)
			n, tipo := 0, ""
			for _, f := range fr {
				if f.Penalizacion > 0 {
					n++
					if tipo == "" {
						tipo = f.Confianza
					}
				}
			}
			if n > 0 {
				agregar(e, EstadoAlerta, fmt.Sprintf("%d fragmento(s) con penalización de confianza", n), fmt.Sprintf(
					"%d de %d fragmentos seleccionados llevan penalización de confianza (%s): falta una fuente oficial igual de cercana.",
					n, len(fr), tipo))
			}
		case EtapaEvidencia:
			if sin, ok := e.Datos["sin_contexto"].(bool); ok && sin && ejecutada(e) {
				texto := "Sin contexto suficiente: falta un documento sobre el tema o la búsqueda no lo encontró."
				if dmin, ok := e.Datos["distancia_min"].(float64); ok {
					maxd, _ := e.Datos["max_distancia"].(float64)
					texto = fmt.Sprintf("Sin contexto suficiente (distancia mínima %s, umbral %s): falta un documento sobre el tema o la búsqueda no lo encontró.",
						Decimal(dmin, 3), Decimal(maxd, 2))
				}
				agregar(e, EstadoAlerta, "", texto)
			}
		case EtapaGeneracion, EtapaCharla:
			if e.Estado == EstadoError {
				agregar(e, EstadoError, "", "El modelo falló o se respondió con un texto de respaldo ("+e.Razon+"): revisar que el LLM esté disponible y responda en español.")
			}
		case EtapaReescritura:
			if e.Estado == EstadoError {
				agregar(e, EstadoError, "", "Reescritura del hilo con error ("+e.Razon+"): las referencias del hilo no se resolvieron.")
			}
		case EtapaJEV:
			if e.Estado == EstadoError {
				agregar(e, EstadoError, "", "JEV con error ("+e.Razon+"): el turno quedó sin juicio.")
			}
		}
	}
	return out
}

// ejecutada: la etapa corrió (no está omitida ni en una rama no tomada).
func ejecutada(e *Etapa) bool {
	return e.Estado != EstadoOmitida && e.Estado != EstadoNoTomada
}
