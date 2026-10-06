package main

// Informe: tabla en consola y metrin/eval/V1_VS_V2.md. Las definiciones de las métricas están aquí, en
// texto, y son las mismas que implementa metricas.go (si cambia una, se cambian las dos).

import (
	"fmt"
	"sort"
	"strings"
)

const definiciones = `
Todas las métricas de calidad se calculan sobre el **último turno** de cada caso (los anteriores son contexto). Las
operativas (latencia, respaldo, SIN_EVIDENCIA, generación) cuentan **todos** los turnos enviados. El mismo código
califica a V1 y a V2: solo cambia de dónde sale cada dato (ver «Cómo se evalúa V1»).

| Métrica | Definición exacta | Casos que cuentan |
|---|---|---|
| **PROCEDURAL ANSWER SUCCESS (PAS)** | Proporción de casos que cumplen **las siete** condiciones de abajo a la vez | ` + "`tipo_esperado = PROCEDURE`" + `, con ` + "`procedimiento_esperado`" + ` en kb/procedimientos y sin ` + "`sin_evidencia_esperada`" + `. Se excluye una continuación sin pasos pendientes |
| C1 · identifica el procedimiento | V2: ` + "`plan.procedimiento.id`" + ` = esperado. V1: cita ≥ 1 fragmento de la tabla ` + "`fuentes`" + ` del YAML **y** su texto cubre ≥ 1 paso del YAML (en casos no PROCEDURE con procedimiento de anclaje basta la cita) | PAS |
| C2 · pasos correctos | Todos los pasos esperados del turno están entre los entregados (recall de pasos = 1). Pasos esperados = ` + "`pasos_esperados`" + ` (null = todos los de primer nivel); en una continuación, solo los que siguen al último paso que la MISMA versión entregó en el turno anterior; si el plan declara entrega por partes (` + "`pasos_mostrados`" + `), la parte mostrada debe ser el comienzo de lo esperado y tener ≥ min(esperados, 3) pasos | PAS |
| C3 · conserva el orden | ≥ 1 paso entregado y: con ids (V2), los n crecen estrictamente en el orden de la respuesta (sin repetidos); sin ids (V1), existe una alineación monótona paso → línea (recorriendo los pasos cubiertos de menor a mayor n, cada uno cabe en una línea que lo cubre en o después de la del anterior; varios pasos pueden caer en la misma línea) | PAS |
| C4 · no inventa pasos | 0 pasos inventados y 0 negritas inventadas (ver «Tasa de invención»). Es la única condición que se cumple sin pasos: una respuesta de respaldo no inventa nada | PAS |
| C5 · asigna bien las fotos | ≥ 1 paso del procedimiento entregado, ninguna foto colocada en un paso del procedimiento es de otro paso (según el YAML), ninguna foto sin evidencia, y si los pasos entregados tienen ` + "`fotos_esperadas`" + `, ≥ 50 % de ellas está en su paso | PAS |
| C6 · responde la intención | Sin error, sin respaldo, sin SIN_EVIDENCIA, ≥ 1 paso entregado y tipo predicho PROCEDURE (o uno de ` + "`tipos_aceptables`" + `) | PAS |
| C7 · las fuentes respaldan los pasos | ≥ 1 paso del procedimiento entregado, lo citado incluye ≥ 1 fragmento de la tabla ` + "`fuentes`" + ` del YAML y cada paso entregado que corresponde a un paso esperado tiene respaldo: V2, su ` + "`fuente`" + ` incluye una del paso del YAML o su texto tiene soporte léxico en lo citado; V1, soporte léxico de la línea en el texto de los fragmentos citados | PAS |
| Precisión de clasificación | tipo predicho = ` + "`tipo_esperado`" + ` (o está en ` + "`tipos_aceptables`" + `). V2: ` + "`plan.tipo`" + `; V1: ` + "`orquestacion.tipo_consulta`" + ` traducido (procedimiento→PROCEDURE, concepto→CONCEPT, problema→TROUBLESHOOTING, comparacion→COMPARISON, social/fuera_de_alcance/ruta conversación→SOCIAL; informacion_directa, aclaracion y seguimiento no tienen equivalente y cuentan como fallo) | Todos |
| Clasificación comparable | La misma, solo con tipos que V1 sabe expresar (CONCEPT, PROCEDURE, TROUBLESHOOTING, COMPARISON, SOCIAL) | Esos tipos |
| Recall@k (k = 5, 10) | Entradas de ` + "`fragmentos_relevantes`" + ` acertadas por los k primeros ítems del ranking ÷ entradas. Un ítem acierta una entrada si representa alguno de sus fragmentos (par id@manual) | Casos con fragmentos relevantes |
| MRR@10 | 1 / posición del primer ítem que acierta alguna entrada (0 si ninguno en los 10 primeros); media | Ídem |
| nDCG@10 | Σ g_i / log2(i+1) ÷ Σ_{i ≤ min(\|R\|,10)} 1 / log2(i+1); g_i = 1 si el ítem i acierta una entrada aún no acertada (binaria, sin dobles cuentas); media | Ídem |
| Acierto de procedimiento | C1 | Casos con procedimiento_esperado |
| Acierto de concepto | V2: ` + "`plan.concepto.id`" + ` = el de kb/conceptos. V1 (o concepto aún sin YAML): respuesta con contenido que cita ≥ 1 fragmento del concepto o de ` + "`fragmentos_relevantes`" + ` | Casos con concepto_esperado |
| Recall de pasos | Pasos esperados del turno entregados ÷ pasos esperados del turno; media | Casos PROCEDURE con procedimiento |
| Recall de imágenes | Σ fotos esperadas (de los pasos esperados del turno) presentes en la respuesta, en cualquier lugar —incluida la galería de ` + "`fuentes[].fotos`" + `, que la página muestra aunque el texto sea de respaldo— ÷ Σ fotos esperadas (micro) | Casos PROCEDURE con fotos esperadas |
| Exactitud paso ↔ foto | Σ fotos colocadas en un paso del procedimiento esperado que el YAML asocia a ese paso ÷ Σ fotos colocadas en pasos del procedimiento esperado (micro). Las fotos de galería (sin paso) no cuentan | Casos con alguna foto colocada |
| Tasa de invención | Turnos con ≥ 1 invención ÷ turnos con contenido (sin error, sin respaldo, sin SIN_EVIDENCIA). Invención = (a) paso entregado que no corresponde a ningún paso del procedimiento esperado **y** no tiene soporte léxico en los fragmentos citados; (b) término entre ` + "`** **`" + ` que no aparece en lo citado (ni en el YAML del procedimiento que el plan dice usar); (c) foto que no está en ningún fragmento citado, ni la ata el servidor a una fuente, ni está en ningún YAML | Turnos con contenido |
| Tasa de invención por paso | Pasos inventados ÷ pasos entregados (micro) | Casos con pasos |
| Tasa de respaldo | Turnos cuyo modo o motivo dice «no disponible», «respaldo» o «segura», o con una etapa de la traza en estado ` + "`respaldo`" + ` ÷ turnos | Todos los turnos |
| Tasa de SIN_EVIDENCIA | Turnos con ` + "`sin_contexto`" + ` (V1) o ` + "`plan.sin_evidencia`" + ` (V2) ÷ turnos | Todos los turnos |
| Abstención correcta | El sistema dice SIN_EVIDENCIA o pide aclaración, sin pasos y sin invención | Casos con sin_evidencia_esperada |
| Falsa abstención | SIN_EVIDENCIA en un caso que sí tiene respuesta | Casos con respuesta esperada |
| Éxito de generación | HTTP 200, ` + "`respuesta`" + ` no vacía y sin respaldo ÷ turnos | Todos los turnos |
| Éxito de la tarea | Si el caso admite ` + "`UNKNOWN`" + ` en ` + "`tipos_aceptables`" + `, pedir aclaración sin inventar siempre es éxito. PROCEDURE con YAML: PAS. CONCEPT: contenido sin invención, tipo correcto, acierto de concepto y ` + "`terminos_esperados`" + ` presentes. NAVIGATION, TROUBLESHOOTING, CONFIGURATION y PROCEDURE sin YAML («implícito»): contenido sin invención, tipo correcto, cita ≥ 1 fragmento relevante y términos presentes (PROCEDURE además ≥ 1 paso). AMBIGUOUS (UNKNOWN): pide aclaración sin inventar, o responde con un tipo de ` + "`tipos_aceptables`" + ` citando lo relevante. Sin evidencia esperada: abstención correcta | Todos |
| Latencia p50 / p95 | Percentil por rango más cercano del tiempo de pared de cada petición (cliente) y de ` + "`traza.total_ms`" + ` (servidor) | Todos los turnos |
| RAM y CPU | ` + "`docker stats --no-stream`" + ` del contenedor cada ~3 s mientras corre cada versión: media y máximo | Con --contenedor |

**Soporte léxico** (el criterio de herramientas/validar_procedimientos.py): palabras de contenido del texto (minúsculas,
sin tildes, ≥ 4 letras, sin palabras vacías ni verbos genéricos de interfaz) recortadas a 5 letras; pasa si comparte
≥ min(2, n) de ellas y ≥ 60 % con el texto de referencia. Aquí además se ignoran las formas de «tú» de esos verbos
(«pulsa», «elige», «ingresa»…), porque V1 tutea y el YAML no.
`

const metodoV1 = `
V1 no devuelve ` + "`plan`" + `, así que lo procedural se infiere de ` + "`respuesta`" + `, ` + "`fuentes`" + ` y la traza:

- **Pasos entregados**: líneas numeradas de ` + "`respuesta`" + ` («1.», «2)», «Paso 3:»); si no hay, viñetas. Una respuesta en prosa
  entrega 0 pasos.
- **Paso entregado ↔ paso del YAML**: una línea cubre un paso (o subpaso) si comparte ≥ min(2, n) raíces y ≥ 50 % de las
  raíces del paso, o si contiene como frase completa una negrita del paso que lo identifica: no genérica (no «Aceptar»,
  «Nuevo») y exclusiva de ese paso dentro del procedimiento («Datos Generales», que está en los pasos 1 y 13 de
  registrar-presupuesto-nuevo, no identifica a ninguno). Una línea puede cubrir varios pasos (V1 los funde).
- **Orden**: se conserva si existe una alineación monótona paso → línea (ver C3).
- **Procedimiento identificado**: cita ≥ 1 fragmento de la tabla ` + "`fuentes`" + ` del YAML y cubre ≥ 1 paso.
- **Fuentes citadas**: ` + "`fuentes[].cita`" + ` → fragmentos de kb/*.jsonl cuya cita (título + «, p. N») es esa, igual que la arma
  internal/indexar/kb_jsonl.go. Las citas que no salen de kb/*.jsonl (el índice del contenedor tiene más trozos que la
  KB) no se pueden resolver: cuentan como no relevantes y su texto no respalda nada. Tampoco se resuelve una cita
  genérica que la KB da a más de 3 fragmentos (p. ej. «S10 Conocimiento», 156 nodos de Cortex).
- **Fotos**: ` + "`![…](fotos/<ruta>)`" + ` debajo de una línea = foto de ese paso (lo que hace colocarFotos); las de
  ` + "`fuentes[].fotos`" + ` son de galería: cuentan para el recall de imágenes, no para la exactitud paso ↔ foto. Se comparan
  por ` + "`sha1(ruta)[:12]`" + `, el id del ESQUEMA.
- **Ranking para Recall@k/MRR/nDCG**: ` + "`fuentes`" + ` en su orden y, detrás, ` + "`traza.seleccion.descartados`" + ` (los mejores candidatos no
  elegidos, en orden), sin repetir cita. V1 entrega como mucho 5 fuentes, así que Recall@10 depende de los descartados.

**Límites de esta heurística (honestos):**

1. Es léxica: una línea que parafrasea mucho no cubre su paso (falso negativo) y una que comparte el nombre de un menú
   con otro paso puede cubrirlo (falso positivo). Los umbrales no se calibraron contra juicio humano.
2. Es **indulgente con V1** donde no se puede distinguir: un paso que no es del procedimiento pero tiene soporte en lo
   citado no es «inventado» (solo «ajeno»); el procedimiento se da por identificado con 1 sola fuente del YAML; la
   pregunta de aclaración se reconoce por un «?» al final. Si V2 supera a V1 aun así, la ventaja es una cota inferior.
3. V1 cita a nivel de respuesta, no de paso: la condición 7 comprueba que la línea tenga soporte en **algo** de lo
   citado, no en la fuente de ese paso.
4. Las invenciones dentro de un paso bien mapeado (un número, un campo, un orden de clics) solo se detectan si van en
   negrita. V1 casi no usa negritas, así que su tasa de invención queda subestimada.
5. V1 no tiene NAVIGATION, CONFIGURATION ni UNKNOWN: su precisión de clasificación en esas categorías es 0 por
   construcción; por eso se informa también la «clasificación comparable».
6. Si el servicio no tiene modelo de lenguaje (texto de respaldo), V1 no entrega pasos salvo en modo tutorial: la línea
   base mide V1 **tal como está desplegado en el contenedor evaluado**, no su techo con modelo.
`

func pct(p *Proporcion) string {
	if p == nil || p.Total == 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f %% (%d/%d)", 100*p.Valor, p.Si, p.Total)
}

func pctCorto(p *Proporcion) string {
	if p == nil || p.Total == 0 {
		return "—"
	}
	return fmt.Sprintf("%.0f %%", 100*p.Valor)
}

func med(m *Media) string {
	if m == nil || m.N == 0 {
		return "—"
	}
	return fmt.Sprintf("%.3f (n=%d)", m.Valor, m.N)
}

func medCorto(m *Media) string {
	if m == nil || m.N == 0 {
		return "—"
	}
	return fmt.Sprintf("%.2f", m.Valor)
}

type filaMetrica struct {
	nombre string
	valor  func(a Agregado) string
}

var filasGlobales = []filaMetrica{
	{"**PROCEDURAL ANSWER SUCCESS**", func(a Agregado) string { return "**" + pct(a.PAS) + "**" }},
	{"C1 identifica el procedimiento", func(a Agregado) string { return pct(a.PASCondiciones[0]) }},
	{"C2 pasos correctos", func(a Agregado) string { return pct(a.PASCondiciones[1]) }},
	{"C3 conserva el orden", func(a Agregado) string { return pct(a.PASCondiciones[2]) }},
	{"C4 no inventa pasos", func(a Agregado) string { return pct(a.PASCondiciones[3]) }},
	{"C5 asigna bien las fotos", func(a Agregado) string { return pct(a.PASCondiciones[4]) }},
	{"C6 responde la intención", func(a Agregado) string { return pct(a.PASCondiciones[5]) }},
	{"C7 fuentes respaldan los pasos", func(a Agregado) string { return pct(a.PASCondiciones[6]) }},
	{"Éxito de la tarea (todas las categorías)", func(a Agregado) string { return pct(a.ExitoTarea) }},
	{"Precisión de clasificación", func(a Agregado) string { return pct(a.ClasificacionAcc) }},
	{"Clasificación comparable", func(a Agregado) string { return pct(a.ClasifComparable) }},
	{"Recall@5", func(a Agregado) string { return med(a.Recall5) }},
	{"Recall@10", func(a Agregado) string { return med(a.Recall10) }},
	{"MRR@10", func(a Agregado) string { return med(a.MRR) }},
	{"nDCG@10", func(a Agregado) string { return med(a.NDCG10) }},
	{"Acierto de procedimiento", func(a Agregado) string { return pct(a.Procedimiento) }},
	{"Acierto de concepto", func(a Agregado) string { return pct(a.Concepto) }},
	{"Recall de pasos", func(a Agregado) string { return med(a.RecallPasos) }},
	{"Recall de imágenes", func(a Agregado) string { return pct(a.RecallImagenes) }},
	{"Exactitud paso ↔ foto", func(a Agregado) string { return pct(a.PasoFoto) }},
	{"Tasa de invención (turnos con contenido)", func(a Agregado) string { return pct(a.Invencion) }},
	{"Tasa de invención por paso", func(a Agregado) string { return pct(a.InvencionPasos) }},
	{"Tasa de respaldo", func(a Agregado) string { return pct(a.Respaldo) }},
	{"Tasa de SIN_EVIDENCIA", func(a Agregado) string { return pct(a.SinEvidencia) }},
	{"Abstención correcta", func(a Agregado) string { return pct(a.Abstencion) }},
	{"Falsa abstención", func(a Agregado) string { return pct(a.FalsaAbstencion) }},
	{"Éxito de generación", func(a Agregado) string { return pct(a.Generacion) }},
	{"Latencia cliente p50 / p95 (ms)", func(a Agregado) string {
		return fmt.Sprintf("%.0f / %.0f", a.Latencia.ClienteP50, a.Latencia.ClienteP95)
	}},
	{"Latencia servidor (traza) p50 / p95 (ms)", func(a Agregado) string {
		return fmt.Sprintf("%.1f / %.1f", a.Latencia.TrazaP50, a.Latencia.TrazaP95)
	}},
	{"Errores HTTP o de red", func(a Agregado) string { return fmt.Sprint(a.Errores) }},
}

func valorVersion(rv *ResultadoVersion, f func(a Agregado) string) string {
	if rv == nil {
		return "—"
	}
	if !rv.Disponible {
		return "no disponible"
	}
	return f(rv.Global)
}

func buscarVersion(c *Corrida, v string) *ResultadoVersion {
	for _, rv := range c.Versiones {
		if rv.Version == v {
			return rv
		}
	}
	return nil
}

func tablaConsola(c *Corrida) string {
	var b strings.Builder
	v1, v2 := buscarVersion(c, "v1"), buscarVersion(c, "v2")
	fmt.Fprintf(&b, "%-44s  %-26s  %-26s\n", "métrica", "V1", "V2")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("─", 100))
	for _, f := range filasGlobales {
		n := strings.Trim(f.nombre, "*")
		fmt.Fprintf(&b, "%-44s  %-26s  %-26s\n", n, strings.Trim(valorVersion(v1, f.valor), "*"), strings.Trim(valorVersion(v2, f.valor), "*"))
	}
	for _, rv := range c.Versiones {
		if !rv.Disponible {
			continue
		}
		fmt.Fprintf(&b, "\n%s por categoría: %-15s %5s %6s %6s %6s %6s %6s %6s %6s\n", strings.ToUpper(rv.Version), "", "casos", "clasif", "R@5", "R@10", "MRR", "éxito", "PAS", "inv")
		for _, cat := range categorias {
			a, ok := rv.PorCat[cat]
			if !ok {
				continue
			}
			fmt.Fprintf(&b, "  %-30s %5d %6s %6s %6s %6s %6s %6s %6s\n", cat, a.Casos, pctCorto(a.ClasificacionAcc),
				medCorto(a.Recall5), medCorto(a.Recall10), medCorto(a.MRR), pctCorto(a.ExitoTarea), pctCorto(a.PAS), pctCorto(a.Invencion))
		}
		if cpu, mem, ok := resumenRecursos(rv.Recursos); ok {
			fmt.Fprintf(&b, "  recursos: %s · %s\n", cpu, mem)
		}
	}
	return b.String()
}

func resumenRecursos(ms []MuestraRecursos) (cpu, mem string, ok bool) {
	if len(ms) == 0 {
		return "", "", false
	}
	var sc, sm, mc, mm float64
	for _, m := range ms {
		sc += m.CPU
		sm += m.MemMiB
		mc = max(mc, m.CPU)
		mm = max(mm, m.MemMiB)
	}
	n := float64(len(ms))
	return fmt.Sprintf("CPU media %.1f %% · máx %.1f %%", sc/n, mc), fmt.Sprintf("RAM media %.0f MiB · máx %.0f MiB (%d muestras)", sm/n, mm, len(ms)), true
}

func informeMarkdown(c *Corrida, op opciones, base *Base, casos map[string]Caso, rutaJSON string) string {
	var b strings.Builder
	v1, v2 := buscarVersion(c, "v1"), buscarVersion(c, "v2")
	w := func(f string, a ...any) { fmt.Fprintf(&b, f, a...) }

	w("# Metrín · Benchmark V1 frente a V2\n\n")
	w("Generado por `go run ./cmd/evalv2` el %s. **No se edita a mano**: se regenera en cada corrida (el JSON crudo de\n", c.Fecha.Format("02-01-2006 15:04 MST"))
	w("cada corrida queda en `eval/resultados/`). Especificación: `docs/V2-RAG-PROCEDURAL.md` §11–12.\n\n")
	w("| | |\n|---|---|\n")
	w("| Servicio | `%s` %s |\n", c.URL, healthTexto(c))
	if c.Contenedor != "" {
		w("| Contenedor (RAM/CPU) | `%s` |\n", c.Contenedor)
	}
	w("| Dataset | `%s`: %d casos (%s) |\n", relativa(c.Dataset, op.Raiz), c.Casos, composicionCorta(c.Composicion))
	if len(c.Excluidos) > 0 {
		w("| Excluidos por no validar | %d: %s |\n", len(c.Excluidos), strings.Join(c.Excluidos, ", "))
	}
	w("| KB al correr | %d fragmentos · %d procedimientos · %d conceptos |\n", c.KB["fragmentos"], c.KB["procedimientos"], c.KB["conceptos"])
	w("| Versiones | %s |\n", estadoVersiones(c))
	w("| JSON crudo | `%s` |\n\n", relativa(rutaJSON, op.Raiz))

	w("## Resultado\n\n")
	w("Regla de §12: **V2 se enciende solo si supera a V1 en PROCEDURAL ANSWER SUCCESS sin empeorar la tasa de invención.**\n\n")
	w("%s\n\n", veredicto(v1, v2))
	w("| Métrica | V1 | V2 |\n|---|---|---|\n")
	for _, f := range filasGlobales {
		w("| %s | %s | %s |\n", f.nombre, valorVersion(v1, f.valor), valorVersion(v2, f.valor))
	}
	for _, rv := range c.Versiones {
		if cpu, mem, ok := resumenRecursos(rv.Recursos); ok {
			w("| RAM y CPU %s | %s · %s | |\n", strings.ToUpper(rv.Version), cpu, mem)
		}
	}
	w("\n")

	for _, rv := range c.Versiones {
		if !rv.Disponible {
			continue
		}
		w("## %s por categoría\n\n", strings.ToUpper(rv.Version))
		w("| Categoría | Casos | Clasificación | Recall@5 | Recall@10 | MRR | nDCG@10 | Éxito tarea | PAS | Invención | Respaldo | SIN_EVID. | p50 / p95 ms |\n")
		w("|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
		for _, cat := range categorias {
			a, ok := rv.PorCat[cat]
			if !ok {
				continue
			}
			w("| %s | %d | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %.0f / %.0f |\n", cat, a.Casos,
				pctCorto(a.ClasificacionAcc), medCorto(a.Recall5), medCorto(a.Recall10), medCorto(a.MRR), medCorto(a.NDCG10),
				pctCorto(a.ExitoTarea), pctCorto(a.PAS), pctCorto(a.Invencion), pctCorto(a.Respaldo), pctCorto(a.SinEvidencia),
				a.Latencia.ClienteP50, a.Latencia.ClienteP95)
		}
		a := rv.Global
		w("| **Global** | %d | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %.0f / %.0f |\n\n", a.Casos,
			pctCorto(a.ClasificacionAcc), medCorto(a.Recall5), medCorto(a.Recall10), medCorto(a.MRR), medCorto(a.NDCG10),
			pctCorto(a.ExitoTarea), pctCorto(a.PAS), pctCorto(a.Invencion), pctCorto(a.Respaldo), pctCorto(a.SinEvidencia),
			a.Latencia.ClienteP50, a.Latencia.ClienteP95)
		w("%s\n", observacionesCorrida(rv, casos))
	}

	w("## Definiciones de las métricas\n%s\n", definiciones)
	w("## Cómo se evalúa V1 (sin `plan`)\n%s\n", metodoV1)
	w("## Cómo se evalúa V2\n\n")
	w("Con `plan` (contrato `internal/v2/tipos`): tipo = `plan.tipo`; procedimiento = `plan.procedimiento.id`; pasos =\n")
	w("`plan.pasos` (o solo `plan.pasos_mostrados` si entrega por partes), cada uno con su `id` (`<proc>#n`), sus `fotos` y\n")
	w("su `fuente`; ranking = candidatos de la traza (`rerank` → `fusion` → `vectorial`/`lexica`, claves `candidatos`/`top`/\n")
	w("`resultados`), o lo citado en orden si la traza no los trae. Un candidato de clase `procedimiento` representa a los\n")
	w("fragmentos de su tabla `fuentes`. Memoria: el arnés reenvía en `memoria` lo que la respuesta anterior trajo en\n")
	w("`memoria` (supuesto: así la guardará la página). Un turno V2 sin `plan` (p. ej. un saludo por la ruta\n")
	w("conversacional) se califica con la heurística de V1. Si ninguna pregunta de sondeo devuelve `plan`, V2 queda\n")
	w("«no disponible» y no se corre.\n\n")
	w("## Dataset\n\n")
	w("`eval/v2_oro.jsonl`, sintético (`sintetico: true` en todos los casos), generado por `eval/construir_v2_oro.py`: las\n")
	w("preguntas están escritas a mano en ese script y las expectativas se DERIVAN de kb/ al construir (reglas de etiquetado\n")
	w("en su cabecera; cada caso trae su `justificacion`). El script rechaza preguntas iguales a `preguntas`/`aliases` de un\n")
	w("YAML o a un ejemplo de entrenamiento de kb/catalogos, para no medir memoria de V2.\n\n")
	w("%s\n", composicionMarkdown(casos))
	w("Las expectativas se refieren al último turno. `procedimiento_esperado` es un id de `kb/procedimientos` o null; con null\n")
	w("y respuesta esperada, el caso es un «procedimiento implícito» (los pasos están en una sección del manual, citada en\n")
	w("`fragmentos_relevantes`, pero no hay YAML): cuenta para clasificación, recuperación, invención y éxito de la tarea,\n")
	w("no para PAS. El arnés valida cada caso contra la KB vigente antes de correr y excluye (con aviso) los que no validan.\n\n")
	w("## Cómo correrlo\n\n```bash\ncd s10-conocimiento/metrin\n")
	w("go run ./cmd/evalv2 --validar                                   # dataset contra kb/, sin HTTP\n")
	w("go run ./cmd/evalv2 --url http://127.0.0.1:4762 --contenedor metrin-traza-prueba\n")
	w("go run ./cmd/evalv2 --muestra 20 --versiones v1                 # prueba rápida\n")
	w("go test ./cmd/evalv2/...                                        # pruebas del cálculo de métricas\n```\n\n")
	w("El servicio debe tener la traza habilitada (`METRIN_TRAZA=1`): sin ella las métricas siguen saliendo, pero Recall@10\n")
	w("de V1 pierde los descartados y la latencia del servidor queda vacía.\n")
	return b.String()
}

// composicionMarkdown: tabla de la mezcla del dataset evaluado.
func composicionMarkdown(casos map[string]Caso) string {
	var b strings.Builder
	fmt.Fprintf(&b, "| Categoría | Casos | Con procedimiento | Sin evidencia esperada | Conversaciones (2–3 turnos) | Variantes |\n|---|---|---|---|---|---|\n")
	tot := [4]int{}
	for _, cat := range categorias {
		n, proc, sinEv, conv := 0, 0, 0, 0
		vars := map[string]int{}
		for _, c := range casos {
			if c.Categoria != cat {
				continue
			}
			n++
			if c.Proc() != "" {
				proc++
			}
			if c.SinEvidenciaEsperada {
				sinEv++
			}
			if len(c.Turnos) > 1 {
				conv++
			}
			for _, v := range c.Variantes {
				vars[v]++
			}
		}
		if n == 0 {
			continue
		}
		tot[0] += n
		tot[1] += proc
		tot[2] += sinEv
		tot[3] += conv
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %s |\n", cat, n, proc, sinEv, conv, conteoPlano(vars))
	}
	fmt.Fprintf(&b, "| **Total** | %d | %d | %d | %d | |\n", tot[0], tot[1], tot[2], tot[3])
	return b.String()
}

func conteoPlano(m map[string]int) string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	var p []string
	for _, k := range ks {
		p = append(p, fmt.Sprintf("%s %d", k, m[k]))
	}
	return strings.Join(p, ", ")
}

func healthTexto(c *Corrida) string {
	if len(c.Health) == 0 {
		return "(sin /health)"
	}
	return "· `/health` → `" + strings.TrimSpace(string(c.Health)) + "`"
}

func relativa(ruta, raiz string) string {
	if strings.HasPrefix(ruta, raiz+"/") {
		return strings.TrimPrefix(ruta, raiz+"/")
	}
	return ruta
}

func composicionCorta(comp map[string]int) string {
	var p []string
	for _, cat := range categorias {
		if comp[cat] > 0 {
			p = append(p, fmt.Sprintf("%s %d", cat, comp[cat]))
		}
	}
	return strings.Join(p, " · ")
}

func estadoVersiones(c *Corrida) string {
	var p []string
	for _, rv := range c.Versiones {
		if rv.Disponible {
			p = append(p, fmt.Sprintf("%s: corrida en %.0f s (%d turnos con `plan`)", strings.ToUpper(rv.Version), rv.Duracion, rv.TurnosPlan))
		} else {
			p = append(p, fmt.Sprintf("%s: **no disponible** — %s", strings.ToUpper(rv.Version), rv.Motivo))
		}
	}
	return strings.Join(p, " · ")
}

func veredicto(v1, v2 *ResultadoVersion) string {
	switch {
	case v1 == nil || !v1.Disponible:
		return "Sin V1 no hay comparación."
	case v2 == nil:
		return "Solo se corrió V1: esta es la **línea base**."
	case !v2.Disponible:
		return "V2 **no disponible** en el servidor evaluado: esta corrida es la **línea base de V1**. No hay veredicto."
	}
	a, b := v1.Global, v2.Global
	if a.PAS.Total == 0 || b.PAS.Total == 0 {
		return "Sin casos PAS aplicables no hay veredicto."
	}
	mejoraPAS := b.PAS.Valor > a.PAS.Valor
	noPeorInv := b.Invencion.Valor <= a.Invencion.Valor
	switch {
	case mejoraPAS && noPeorInv:
		return fmt.Sprintf("V2 **cumple** la regla: PAS %.1f %% frente a %.1f %%, invención %.1f %% frente a %.1f %%. (Con %d casos PAS, una diferencia de un caso es %.1f puntos: mire el detalle antes de encenderla.)",
			100*b.PAS.Valor, 100*a.PAS.Valor, 100*b.Invencion.Valor, 100*a.Invencion.Valor, b.PAS.Total, 100/float64(b.PAS.Total))
	case mejoraPAS:
		return fmt.Sprintf("V2 **no cumple** la regla: mejora PAS (%.1f %% frente a %.1f %%) pero empeora la invención (%.1f %% frente a %.1f %%).",
			100*b.PAS.Valor, 100*a.PAS.Valor, 100*b.Invencion.Valor, 100*a.Invencion.Valor)
	default:
		return fmt.Sprintf("V2 **no cumple** la regla: PAS %.1f %% frente a %.1f %% de V1.", 100*b.PAS.Valor, 100*a.PAS.Valor)
	}
}

// observacionesCorrida: hechos de la corrida que explican los números (modos, citas sin resolver, fallos de PAS).
func observacionesCorrida(rv *ResultadoVersion, casos map[string]Caso) string {
	var b strings.Builder
	modos := map[string]int{}
	tiposCrudos := map[string]int{}
	citas, sinResolver, turnos, conPasos := 0, 0, 0, 0
	for _, e := range rv.Casos {
		for _, o := range e.Observaciones {
			turnos++
			m := o.Modo
			if o.Error != "" {
				m = "error"
			}
			modos[m]++
			tiposCrudos[o.TipoCrudo]++
			citas += len(o.Citadas) + len(o.CitasSinResolver)
			sinResolver += len(o.CitasSinResolver)
			if len(o.Pasos) > 0 {
				conPasos++
			}
		}
	}
	motivos := map[string]int{}
	seguimientoPrimero, seguimientoConHilo := 0, 0
	for _, ts := range rv.Crudos {
		for i, t := range ts {
			if t.Respuesta == nil {
				continue
			}
			if t.Respuesta.Motivo != "" {
				motivos[t.Respuesta.Motivo]++
			}
			// Hasta el 06-10-2026 la página mandaba la pregunta actual dentro de `hilo` desde el primer turno;
			// ahora `hilo` lleva solo los turnos anteriores (vacío en el primero).
			if i == 0 && t.Respuesta.Orquestacion.TipoConsulta == "seguimiento" {
				seguimientoPrimero++
				if len(t.Peticion.Hilo) > 0 {
					seguimientoConHilo++
				}
			}
		}
	}
	fmt.Fprintf(&b, "Observaciones de %s (automáticas):\n\n", strings.ToUpper(rv.Version))
	fmt.Fprintf(&b, "- Modos de respuesta en %d turnos: %s.\n", turnos, conteo(modos))
	if len(motivos) > 0 {
		fmt.Fprintf(&b, "- Motivos declarados por el servicio: %s.\n", conteo(motivos))
	}
	if seguimientoConHilo > 0 {
		fmt.Fprintf(&b, "- %d preguntas de PRIMER turno salieron con tipo `seguimiento` con la pregunta actual dentro de `hilo` "+
			"(hilo de 1 turno, la página anterior al 06-10-2026) y V1 marca seguimiento si la pregunta contiene «su», «ese», «eso», «después»…; si la reescritura "+
			"falla (sin modelo), el tipo se queda en `seguimiento`, que no equivale a ningún tipo de V2.\n", seguimientoConHilo)
	}
	if n := seguimientoPrimero - seguimientoConHilo; n > 0 {
		fmt.Fprintf(&b, "- %d preguntas de PRIMER turno salieron con tipo `seguimiento` con el `hilo` vacío.\n", n)
	}
	fmt.Fprintf(&b, "- Tipo crudo devuelto: %s.\n", conteo(tiposCrudos))
	fmt.Fprintf(&b, "- Turnos con pasos numerados: %d de %d.\n", conPasos, turnos)
	if citas > 0 {
		fmt.Fprintf(&b, "- Citas que no salen de kb/*.jsonl (no resolubles a un fragmento): %d de %d (%.0f %%).\n", sinResolver, citas, 100*float64(sinResolver)/float64(citas))
	}
	fallos := [7]int{}
	nPAS := 0
	var exitos []string
	for _, e := range rv.Casos {
		if e.PAS == nil {
			continue
		}
		nPAS++
		for i, v := range e.PAS.Condiciones() {
			if !v {
				fallos[i]++
			}
		}
		if e.PAS.Exito {
			exitos = append(exitos, e.Caso)
		}
	}
	if nPAS > 0 {
		nombres := []string{"C1", "C2", "C3", "C4", "C5", "C6", "C7"}
		var p []string
		for i, n := range fallos {
			p = append(p, fmt.Sprintf("%s %d", nombres[i], n))
		}
		fmt.Fprintf(&b, "- Casos PAS que fallan cada condición (de %d): %s.\n", nPAS, strings.Join(p, " · "))
		if len(exitos) > 0 {
			fmt.Fprintf(&b, "- Casos PAS con éxito: %s.\n", strings.Join(exitos, ", "))
		}
	}
	return b.String()
}

func conteo(m map[string]int) string {
	type kv struct {
		k string
		v int
	}
	var l []kv
	for k, v := range m {
		if k == "" {
			k = "(vacío)"
		}
		l = append(l, kv{k, v})
	}
	sort.Slice(l, func(i, j int) bool { return l[i].v > l[j].v || (l[i].v == l[j].v && l[i].k < l[j].k) })
	var p []string
	for _, x := range l {
		p = append(p, fmt.Sprintf("`%s` %d", x.k, x.v))
	}
	return strings.Join(p, ", ")
}
