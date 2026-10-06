package rag

// Instrumentación del modo traza (docs/CONTRATO-traza-metrin.md). Todo pasa por
// un *traza.Recorder que viaja en el contexto; con el modo apagado no hay
// recorder y estas funciones ni se llaman (o regresan al instante): la
// respuesta y las decisiones del RAG no cambian. Aquí solo se anota lo que el
// flujo ya decidió, con los valores que el flujo ya calculó.

import (
	"context"
	"math"
	"strconv"
	"time"

	"rag-go/internal/almacen"
	"rag-go/internal/clasificar"
	"rag-go/internal/embed"
	"rag-go/internal/jev"
	"rag-go/internal/llm"
	"rag-go/internal/traza"
)

// maxDescartados: candidatos que no entraron al contexto y se muestran con su
// distancia (sirven para ver dónde cae el corte).
const maxDescartados = 8

// embMedido envuelve el embebedor del clasificador para medir cuánto tarda en
// embeber la pregunta. Solo se usa con el modo traza encendido; delega todo.
type embMedido struct {
	embed.Embebedor
	dur      time.Duration
	err      error
	llamadas int
}

func (m *embMedido) Embeber(ctx context.Context, texto string) ([]float32, error) {
	t0 := time.Now()
	v, err := m.Embebedor.Embeber(ctx, texto)
	m.dur += time.Since(t0)
	m.llamadas++
	if err != nil {
		m.err = err
	}
	return v, err
}

// describirEmbebedor: proveedor y modelo del embebedor ("" si no se conoce).
func describirEmbebedor(e embed.Embebedor) (proveedor, modelo string) {
	switch x := e.(type) {
	case nil:
		return "", ""
	case *embMedido:
		return describirEmbebedor(x.Embebedor)
	case *embed.Estatico:
		return "estatico", x.Nombre()
	case *embed.Ollama:
		return "ollama", x.Modelo
	}
	return "", e.Nombre()
}

// describirLLM: proveedor y modelo del LLM ("" si es un doble de prueba u
// otro Chateador cuyo modelo no se expone).
func describirLLM(c Chateador) (proveedor, modelo string) {
	switch x := c.(type) {
	case *llm.Ollama:
		return "ollama", x.Modelo
	case *llm.MLX:
		return "mlx", x.Modelo
	}
	return "", ""
}

// nulo: "" viaja como null (dato no expuesto o que no aplica).
func nulo(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func filtroTraza(f map[string]string) any {
	if len(f) == 0 {
		return nil
	}
	out := make(map[string]string, len(f))
	for k, v := range f {
		out[k] = v
	}
	return out
}

func errTexto(err error) string { return traza.ResumirError(err) }

func trazarEntrada(rec *traza.Recorder, pregunta string, o Opciones) {
	rec.Fin(traza.EtapaEntrada, traza.EstadoOK, traza.Datos{
		"pregunta":    pregunta,
		"hilo_turnos": len(o.Hilo),
		"k":           o.K,
		"filtro":      filtroTraza(o.Filtro),
	})
}

func trazarAciertos(rec *traza.Recorder, aplica bool) {
	razon := "no pregunta por los aciertos de hoy: sigue el flujo"
	if aplica {
		razon = "pregunta por los aciertos de hoy: respuesta fija, sin buscar ni llamar al modelo"
	}
	rec.Fin(traza.EtapaReglaAciertos, traza.EstadoOK, traza.Datos{"aplica": aplica})
	rec.Razon(traza.EtapaReglaAciertos, razon)
}

func trazarRuta(rec *traza.Recorder, plan Orquestacion) {
	razon := "trabajo: búsqueda con fuentes"
	switch plan.Ruta {
	case rutaConversacion:
		razon = "charla directa, sin búsqueda"
	case rutaRegla:
		razon = "regla «aciertos»: respuesta fija"
	}
	rec.Fin(traza.EtapaRuta, traza.EstadoOK, traza.Datos{
		"ruta":          plan.Ruta,
		"tipo_consulta": plan.TipoConsulta,
		"intencion":     plan.Intencion,
	})
	rec.Razon(traza.EtapaRuta, razon)
}

// trazarSinClasificador: no hay modelo (o no hay embebedor) y planificar
// manda todo a la ruta segura sin clasificar.
func trazarSinClasificador(rec *traza.Recorder) {
	rec.OmitirCon(traza.EtapaClasificador, traza.EstadoOmitida,
		"sin clasificador cargado: todo va a la ruta segura (rag)",
		traza.Datos{"clasificador": "reglas_seguras", "intencion": clasificar.Trabajo})
	rec.Modelo(traza.EtapaClasificador, "reglas_seguras")
	rec.Omitir(traza.EtapaSeguridad, "sin clasificador no hay etiqueta conversacional que desviar")
}

// trazarClasificacion anota el embebedor (solo el embebido de la pregunta
// para el clasificador) y el clasificador (el tiempo restante: kNN).
func (r *RAG) trazarClasificacion(rec *traza.Recorder, pregunta, nombre string, medido *embMedido, total time.Duration, intencion string, similitud float64, err error) {
	prov, mod := describirEmbebedor(r.Emb)
	estado := traza.EstadoOK
	razon := "ms = embeber la pregunta para el clasificador; cada pase de la búsqueda la vuelve a embeber y eso va en «busqueda»"
	if medido.err != nil {
		estado = traza.EstadoError
		razon = "no pudo embeber la pregunta: " + errTexto(medido.err)
	}
	rec.FinDuracion(traza.EtapaEmbebedor, estado, medido.dur, traza.Datos{
		"proveedor": nulo(prov),
		"modelo":    nulo(mod),
		"llamadas":  medido.llamadas,
	})
	rec.Razon(traza.EtapaEmbebedor, razon)
	rec.Modelo(traza.EtapaEmbebedor, mod)

	knn := max(total-medido.dur, 0)
	rec.Modelo(traza.EtapaClasificador, nombre)
	if err != nil {
		rec.FinDuracion(traza.EtapaClasificador, traza.EstadoRespaldo, knn, traza.Datos{
			"clasificador": "error_fallback_rag",
			"intencion":    clasificar.Trabajo,
			"umbral":       r.Clasificador.Umbral,
		})
		rec.Razon(traza.EtapaClasificador, "falló, siguió por la ruta segura (rag): "+errTexto(err))
		rec.Omitir(traza.EtapaSeguridad, "el clasificador falló: no hay etiqueta que desviar")
		return
	}
	d := traza.Datos{
		"clasificador":  nombre,
		"intencion":     intencion,
		"umbral":        r.Clasificador.Umbral,
		"similitud":     nil,
		"paso_umbral":   nil,
		"charla_lexica": clasificar.PuntajeSocial(pregunta) == 1,
	}
	razon = "intención «" + intencion + "» sin similitud (ningún ejemplo comparable)"
	if !math.IsInf(similitud, 0) && !math.IsNaN(similitud) {
		d["similitud"] = traza.Redondear(similitud)
		d["paso_umbral"] = similitud >= r.Clasificador.Umbral
		comparacion := " ≥ "
		if similitud < r.Clasificador.Umbral {
			comparacion = " < "
		}
		razon = "intención «" + intencion + "»: similitud " + traza.Decimal(similitud, 3) + comparacion +
			"umbral " + traza.Decimal(r.Clasificador.Umbral, 2)
	}
	rec.FinDuracion(traza.EtapaClasificador, traza.EstadoOK, knn, d)
	rec.Razon(traza.EtapaClasificador, razon)
}

func trazarSeguridad(rec *traza.Recorder, intencion string, larga, desvia bool) {
	razon := "intención «" + intencion + "»: no hay nada que desviar"
	switch {
	case desvia:
		razon = "mensaje corto con etiqueta «" + intencion + "»: va a charla directa"
	case intencion == clasificar.Social || intencion == "limite":
		razon = "pregunta larga con «?»: se queda en rag pese a la etiqueta «" + intencion + "»"
	}
	rec.Fin(traza.EtapaSeguridad, traza.EstadoOK, traza.Datos{
		"pregunta_larga":  larga,
		"desvia_a_charla": desvia,
		"intencion":       intencion,
	})
	rec.Razon(traza.EtapaSeguridad, razon)
}

func trazarReescritura(rec *traza.Recorder, c Chateador, original, reescrita string, err error) {
	prov, mod := describirLLM(c)
	estado, razon := traza.EstadoOK, "la pregunta efectiva resuelve referencias del hilo"
	switch {
	case err != nil:
		estado, razon = traza.EstadoError, "se usó la pregunta original · "+errTexto(err)
	case reescrita == "":
		estado, razon = traza.EstadoRespaldo, "reescritura vacía: se usa la pregunta original"
	}
	if prov == "" {
		razon += " · modelo no expuesto"
	}
	rec.Fin(traza.EtapaReescritura, estado, traza.Datos{
		"original":  original,
		"reescrita": nulo(reescrita),
		"proveedor": nulo(prov),
		"modelo":    nulo(mod),
	})
	rec.Razon(traza.EtapaReescritura, razon)
	rec.Modelo(traza.EtapaReescritura, mod)
}

// trazarBusqueda anota los pases de búsqueda. Los pases opcionales dejaron su
// marca con Dato (k_oficial, consulta_compacta, k_cortex) al ejecutarse.
func trazarBusqueda(rec *traza.Recorder, o Opciones, consulta string, refuerzo bool, trozos, oficiales, cortex []almacen.Resultado, msGeneral int64, errCortex error) {
	consultas := 1
	if refuerzo {
		consultas += 3 // páginas 11, 12 y 17 de la guía de registro
	}
	for _, marca := range []string{"k_oficial", "consulta_compacta", "k_cortex"} {
		if _, ok := rec.DatoDe(traza.EtapaBusqueda, marca); ok {
			consultas++
		}
	}
	d := traza.Datos{
		"metodo":            "vectorial",
		"k":                 o.K,
		"filtro":            filtroTraza(o.Filtro),
		"resultados":        len(trozos) + len(oficiales) + len(cortex),
		"general":           len(trozos),
		"oficiales":         len(oficiales),
		"cortex":            len(cortex),
		"consultas":         consultas,
		"consulta":          consulta,
		"refuerzo_registro": refuerzo,
		"ms_general":        msGeneral,
	}
	razon := "solo vectorial (sin BM25); ms incluye embeber la consulta en cada pase"
	if errCortex != nil {
		d["error_cortex"] = errTexto(errCortex)
		razon = "el pase Cortex falló y se ignoró: " + errTexto(errCortex)
	}
	rec.Fin(traza.EtapaBusqueda, traza.EstadoOK, d)
	rec.Razon(traza.EtapaBusqueda, razon)
}

func fragmentoDe(t almacen.Resultado) traza.Fragmento {
	f := traza.Fragmento{
		Documento:    t.Metadata["title"],
		Seccion:      t.Metadata["section"],
		Cita:         t.Metadata["cita"],
		Confianza:    t.Metadata["confianza"],
		Distancia:    traza.Redondear(t.Distancia),
		Penalizacion: penalizacionConfianza[t.Metadata["confianza"]],
	}
	if f.Documento == "" {
		f.Documento = t.Metadata["document_id"]
	}
	if p, err := strconv.Atoi(t.Metadata["page"]); err == nil && p > 0 {
		f.Pagina = &p
	}
	return f
}

// trazarSeleccion: candidatos ya vienen ordenados y con la penalización sumada
// por seleccionarContexto; corte es el límite que esa función calculó.
func trazarSeleccion(rec *traza.Recorder, candidatos, seleccionados []almacen.Resultado, corte, maxDistancia float64) {
	fragmentos := make([]traza.Fragmento, 0, len(seleccionados))
	elegidos := map[string]bool{}
	for _, t := range seleccionados {
		fragmentos = append(fragmentos, fragmentoDe(t))
		elegidos[t.ID] = true
	}
	descartados := []traza.Fragmento{}
	vistos := map[string]bool{}
	for _, c := range candidatos {
		if len(descartados) == maxDescartados {
			break
		}
		if elegidos[c.ID] || vistos[c.ID] {
			continue
		}
		vistos[c.ID] = true
		f := fragmentoDe(c)
		dentro := c.Distancia <= corte
		f.DentroDelCorte = &dentro
		descartados = append(descartados, f)
	}
	penalizaciones := make(map[string]float64, len(penalizacionConfianza))
	for k, v := range penalizacionConfianza {
		penalizaciones[k] = v
	}
	d := traza.Datos{
		"ventana":        ventanaRelevancia,
		"max_contextos":  maxContextos,
		"max_distancia":  maxDistancia,
		"corte":          nil,
		"mejor":          nil,
		"candidatos":     len(candidatos),
		"seleccionados":  len(seleccionados),
		"fragmentos":     fragmentos,
		"descartados":    descartados,
		"penalizaciones": penalizaciones,
	}
	razon := "sin candidatos: no hay nada que seleccionar"
	if len(candidatos) > 0 {
		d["corte"] = traza.Redondear(corte)
		d["mejor"] = traza.Redondear(candidatos[0].Distancia)
		razon = "entran los que quedan dentro de mejor + ventana (y bajo max_distancia), sin citas repetidas, hasta " + strconv.Itoa(maxContextos)
	}
	if regla, ok := rec.DatoDe(traza.EtapaSeleccion, "regla"); ok {
		nombre, _ := regla.(string)
		razon = "la regla «" + nombre + "» reemplazó la selección vectorial"
	}
	rec.Fin(traza.EtapaSeleccion, traza.EstadoOK, d)
	rec.Razon(traza.EtapaSeleccion, razon)
}

func trazarEvidencia(rec *traza.Recorder, res Respuesta, maxDistancia float64, razon string) {
	if razon == "" {
		if res.SinContexto {
			razon = res.Motivo
		} else {
			razon = "distancia mínima " + traza.Decimal(res.DistanciaMin, 3) + " ≤ umbral " + traza.Decimal(maxDistancia, 2)
		}
	}
	rec.Fin(traza.EtapaEvidencia, traza.EstadoOK, traza.Datos{
		"sin_contexto":  res.SinContexto,
		"distancia_min": res.DistanciaMin,
		"max_distancia": maxDistancia,
		"motivo":        nulo(res.Motivo),
	})
	rec.Razon(traza.EtapaEvidencia, razon)
}

// trazarRegistroFijo: la intención de registrar un presupuesto responde con
// los pasos verificados de la guía, sin LLM.
func trazarRegistroFijo(rec *traza.Recorder, res Respuesta, maxDistancia float64) {
	trazarEvidencia(rec, res, maxDistancia, "hay páginas de la guía de registro: respuesta fija con pasos verificados")
	rec.Omitir(traza.EtapaGeneracion, "respuesta fija de registro de presupuesto: pasos verificados, sin LLM")
	rec.Omitir(traza.EtapaVerificacion, "respuesta fija: no hay texto del modelo que verificar")
}

func trazarGeneracion(rec *traza.Recorder, c Chateador, res Respuesta, contextos int, err error, respaldo bool) {
	prov, mod := describirLLM(c)
	estado, razon := traza.EstadoOK, "respuesta con "+strconv.Itoa(contextos)+" fragmento(s) de contexto"
	switch {
	case err != nil:
		estado, razon = traza.EstadoError, "respaldo: "+errTexto(err)
	case respaldo:
		estado, razon = traza.EstadoError, "respaldo: el modelo no respondió en español"
	}
	if prov == "" {
		razon += " · modelo no expuesto"
	}
	rec.Fin(traza.EtapaGeneracion, estado, traza.Datos{
		"proveedor":     nulo(prov),
		"modelo":        nulo(mod),
		"modo":          res.Modo,
		"ms_llm":        res.MsLLM,
		"contextos":     contextos,
		"tipo_consulta": res.Plan.TipoConsulta,
	})
	rec.Razon(traza.EtapaGeneracion, razon)
	rec.Modelo(traza.EtapaGeneracion, mod)
}

func trazarVerificacion(rec *traza.Recorder, fuentes []Fuente, diceSinContexto bool) {
	citas := make([]string, 0, len(fuentes))
	for _, f := range fuentes {
		if f.Cita != "" {
			citas = append(citas, f.Cita)
		}
	}
	estado := traza.EstadoOK
	razon := "solo se comprueba la frase de abstención; la comprobación cita a cita no existe en el código"
	if diceSinContexto {
		estado = traza.EstadoAlerta
		razon = "el modelo dijo que el contexto no alcanza: se retiran las fuentes y se registra el fallo"
	}
	rec.Fin(traza.EtapaVerificacion, estado, traza.Datos{
		"citas":             citas,
		"n_citas":           len(citas),
		"dice_sin_contexto": diceSinContexto,
	})
	rec.Razon(traza.EtapaVerificacion, razon)
}

func trazarCharla(rec *traza.Recorder, c Chateador, intencion string, res Respuesta, err error) {
	prov, mod := describirLLM(c)
	estado, razon := traza.EstadoOK, "charla breve sin búsqueda (≤ 60 palabras)"
	switch {
	case err != nil:
		estado, razon = traza.EstadoError, "saludo de respaldo: "+errTexto(err)
	case res.Motivo != "":
		estado, razon = traza.EstadoError, res.Motivo
	}
	if prov == "" {
		razon += " · modelo no expuesto"
	}
	rec.Fin(traza.EtapaCharla, estado, traza.Datos{
		"modelo":    nulo(mod),
		"proveedor": nulo(prov),
		"intencion": intencion,
		"ms_llm":    res.MsLLM,
	})
	rec.Razon(traza.EtapaCharla, razon)
	rec.Modelo(traza.EtapaCharla, mod)
}

func trazarCharlaFija(rec *traza.Recorder, intencion string) {
	rec.Fin(traza.EtapaCharla, traza.EstadoOK, traza.Datos{"modelo": nil, "intencion": intencion})
	rec.Razon(traza.EtapaCharla, "intención «"+intencion+"»: respuesta fija, sin llamar al modelo")
}

func trazarJEV(rec *traza.Recorder, cli *jev.Cliente, resp jev.Respuesta, err error) {
	d := traza.Datos{"activo": true, "modelo": nulo(cli.Modelo)}
	estado, razon := traza.EstadoOK, "noul = p(true) de que la decisión del turno sea correcta"
	if err != nil {
		estado, razon = traza.EstadoError, errTexto(err)
	} else {
		d["noul"] = traza.Redondear(resp.Noul)
		d["confianza"] = traza.Redondear(resp.Confianza)
	}
	rec.Fin(traza.EtapaJEV, estado, d)
	rec.Razon(traza.EtapaJEV, razon)
	rec.Modelo(traza.EtapaJEV, cli.Modelo)
}

// anotarFallo registra el fallo como siempre y, con el modo traza, anota si
// quedó escrito.
func (r *RAG) anotarFallo(ctx context.Context, res Respuesta) error {
	rec := traza.De(ctx)
	rec.Inicio(traza.EtapaRegistroFallos)
	err := r.registrarFallo(res)
	if rec == nil {
		return err
	}
	switch {
	case r.RutaFallos == "":
		rec.OmitirCon(traza.EtapaRegistroFallos, traza.EstadoOmitida, "RutaFallos vacío: registro apagado", traza.Datos{"registrado": false})
	case err != nil:
		rec.Fin(traza.EtapaRegistroFallos, traza.EstadoError, traza.Datos{"registrado": false})
		rec.Razon(traza.EtapaRegistroFallos, "no se pudo escribir el fallo: "+errTexto(err))
	default:
		rec.Fin(traza.EtapaRegistroFallos, traza.EstadoOK, traza.Datos{"registrado": true})
		rec.Razon(traza.EtapaRegistroFallos, "pregunta, motivo, distancia mínima y fuentes van al JSONL de fallos")
	}
	return err
}

// completarTraza cierra lo que el flujo no tocó, con la razón de su ruta, y
// anota la etapa «fin». La llama Preguntar solo con el modo traza encendido.
func (r *RAG) completarTraza(rec *traza.Recorder, res Respuesta, err error) {
	if err != nil {
		rec.FallarIniciadas(errTexto(err))
	}
	omitir := func(id, razon string) {
		if rec.Pendiente(id) {
			rec.Omitir(id, razon)
		}
	}
	noTomada := func(id, razon string) {
		if rec.Pendiente(id) {
			rec.NoTomada(id, razon)
		}
	}
	rama := []string{traza.EtapaReescritura, traza.EtapaBusqueda, traza.EtapaSeleccion,
		traza.EtapaEvidencia, traza.EtapaGeneracion, traza.EtapaVerificacion}

	ruta, _ := rec.DatoDe(traza.EtapaRuta, "ruta")
	switch ruta {
	case rutaRegla:
		const razon = "la regla «aciertos» respondió sin buscar ni llamar al modelo"
		for _, id := range rama {
			omitir(id, razon)
		}
		omitir(traza.EtapaCharla, razon)
	case rutaConversacion:
		for _, id := range rama {
			noTomada(id, "ruta conversacion: charla directa, sin búsqueda")
		}
	default:
		noTomada(traza.EtapaCharla, "ruta rag: respuesta con fuentes")
		omitir(traza.EtapaReescritura, "sin hilo, no hay reescritura")
		if err != nil {
			for _, id := range rama {
				omitir(id, "el flujo se cortó por un error antes de esta etapa")
			}
		}
		if sin, _ := rec.DatoDe(traza.EtapaEvidencia, "sin_contexto"); sin == true {
			omitir(traza.EtapaGeneracion, "sin contexto suficiente: no se llama al LLM de respuesta")
		}
		omitir(traza.EtapaVerificacion, "no hubo texto del modelo que verificar")
	}

	if rec.Pendiente(traza.EtapaEmbebedor) {
		prov, mod := describirEmbebedor(r.Emb)
		d := traza.Datos{"proveedor": nulo(prov), "modelo": nulo(mod)}
		razon := "sin clasificador: la pregunta solo se embebe dentro de la búsqueda (su tiempo va en «busqueda»)"
		if r.Emb == nil {
			razon += " · embebedor del almacén no expuesto"
		}
		if e := rec.Estado(traza.EtapaBusqueda); e == traza.EstadoOK || e == traza.EstadoError {
			rec.Fin(traza.EtapaEmbebedor, traza.EstadoOK, d)
			rec.Razon(traza.EtapaEmbebedor, razon)
			rec.Modelo(traza.EtapaEmbebedor, mod)
		} else {
			rec.OmitirCon(traza.EtapaEmbebedor, traza.EstadoOmitida, "no se embebió la pregunta en este turno", d)
		}
	}
	if rec.Pendiente(traza.EtapaJEV) {
		if r.JEV == nil {
			rec.OmitirCon(traza.EtapaJEV, traza.EstadoOmitida, "JEV apagado (sin JEV_URL)", traza.Datos{"activo": false})
		} else {
			rec.OmitirCon(traza.EtapaJEV, traza.EstadoOmitida, "este camino no pide juicio a JEV",
				traza.Datos{"activo": true, "modelo": nulo(r.JEV.Modelo)})
		}
	}
	if rec.Pendiente(traza.EtapaRegistroFallos) {
		razon := "no hubo fallo que registrar"
		if r.RutaFallos == "" {
			razon = "RutaFallos vacío: registro apagado"
		}
		rec.OmitirCon(traza.EtapaRegistroFallos, traza.EstadoOmitida, razon, traza.Datos{"registrado": false})
	}

	estado, razon := traza.EstadoOK, ""
	if err != nil {
		estado, razon = traza.EstadoError, "error: "+errTexto(err)
	}
	rec.Fin(traza.EtapaFin, estado, traza.Datos{
		"modo":          res.Modo,
		"ruta":          res.Plan.Ruta,
		"tipo_consulta": res.Plan.TipoConsulta,
		"sin_contexto":  res.SinContexto,
		"fuentes":       len(res.Fuentes),
		"motivo":        nulo(res.Motivo),
	})
	rec.Razon(traza.EtapaFin, razon)
}
