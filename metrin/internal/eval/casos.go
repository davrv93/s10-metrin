package eval

// Suite fija por defecto del harness. Los casos «responder» usan
// preguntas con respaldo real en la base indexada (kb/fragmentos*.
// jsonl: FAQ, vídeos, reportes, pantallas y el PDF de políticas);
// los «abstener» son fuera de dominio (o trampa: mezclan términos
// del corpus con un tema imposible) y son largos a propósito,
// porque un mensaje corto sin contexto es charla conversacional
// por diseño, no abstención.
//
// Para añadir o cambiar casos sin recompilar: rag eval --casos
// mis-casos.jsonl (mismo formato: pregunta, esperado, nota).

func CasosPorDefecto() []Caso {
	return []Caso{
		{
			Pregunta: "¿Qué es exactamente S10 ERP y para qué tipo de empresas está pensado?",
			Esperado: "responder",
			Nota:     "FAQ de la base: definición y sector objetivo (construcción e inmobiliario).",
		},
		{
			Pregunta: "¿Quién fabrica el S10? Me refiero a la empresa dueña del software.",
			Esperado: "responder",
			Nota:     "FAQ de la base: SISTEMA10 SAC, la empresa detrás del ERP.",
		},
		{
			Pregunta: "¿Cómo se registra el valor de la UIT y de la remuneración mínima vital en el módulo de Nóminas?",
			Esperado: "responder",
			Nota:     "Vídeo oficial de Nóminas: cambio de UIT y RMV.",
		},
		{
			Pregunta: "¿Cómo se registran de forma manual los porcentajes de las AFP en el ERP S10?",
			Esperado: "responder",
			Nota:     "Vídeo oficial de Nóminas: registro manual de porcentajes de AFP.",
		},
		{
			Pregunta: "¿Qué reporte muestra las horas y los trabajadores tareados por día, y con qué campos se pide?",
			Esperado: "responder",
			Nota:     "Catálogo de reportes: REP-02 (desde, hasta, categoría).",
		},
		{
			Pregunta: "¿Dónde está la pantalla de Constantes por Fecha General en el módulo de Nóminas y qué campos de formulario tiene?",
			Esperado: "responder",
			Nota:     "Base de pantallas: «por Fecha General» (Nóminas).",
		},
		{
			Pregunta: "¿Qué compromiso de prevención en salud y seguridad tiene SISTEMA10 con sus colaboradores según su política SGSST?",
			Esperado: "responder",
			Nota:     "PDF oficial: Política del Sistema de Gestión de Seguridad y Salud en el Trabajo.",
		},
		{
			Pregunta: "¿Cómo reservo un vuelo internacional desde Lima a Madrid con la aerolínea oficial del proyecto?",
			Esperado: "abstener",
			Nota:     "Fuera de dominio: viajes no están en la base.",
		},
		{
			Pregunta: "¿Qué ingredientes necesito para preparar un ceviche de camote con mango para veinte invitados?",
			Esperado: "abstener",
			Nota:     "Fuera de dominio: recetas de cocina.",
		},
		{
			Pregunta: "¿Cuál es el precio del barril de petróleo Brent hoy en el mercado de futuros de Londres?",
			Esperado: "abstener",
			Nota:     "Fuera de dominio y cambiante: finanzas, no ERP.",
		},
		{
			Pregunta: "¿Cómo configuro la velocidad del ventilador de la tarjeta gráfica desde el módulo de Nóminas?",
			Esperado: "abstener",
			Nota:     "Trampa: término del corpus (Nóminas) con acción imposible; no debe inventar pasos.",
		},
		{
			Pregunta: "¿Qué días feriados hay en Japón durante el año dos mil veintiséis?",
			Esperado: "abstener",
			Nota:     "Fuera de dominio: calendario japonés.",
		},
		{
			Pregunta: "¿Cómo se poda un árbol de mango en la oficina de Gerencia de Proyectos del ERP?",
			Esperado: "abstener",
			Nota:     "Trampa: módulo del corpus (Gerencia de Proyectos) con tema ajeno.",
		},
	}
}
