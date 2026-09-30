---
title: Análisis de precios unitarios y catálogos
summary: "El APU de cada partida se arma con recursos (mano de obra, materiales,
  equipos, subcontratos) y subpartidas del catálogo, más el rendimiento diario.
  Las partidas y recursos nuevos se crean en Catálogos. Búsqueda: tecleo directo
  o F8 con hasta 3 cadenas."
tags:
  - presupuestos
  - apu
  - catalogos
  - rendimiento
  - tercero
id: 01M3RJP5VMQVGNSE7Z4JV0PJYJ
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.244Z
---

## Catálogos principales (menú Catálogos)
- **Recursos**: mano de obra, materiales, equipos y subcontratos. Todo recurso de un presupuesto debe estar aquí.
- **Unidades**.
- **Partidas**, **Títulos** y **Mis Favoritos**.
- **Plantillas**: grupos de recursos para insertar en un APU.
- **Identificadores**: clientes, proveedores, obreros, AFP…
- **Lugares**: departamento > provincia > distrito.
- **Índices unificados**: no se registran nuevos salvo que lo haga la institución que los rige; agrupan recursos para la fórmula polinómica.

Un registro solo se puede borrar si no está en uso.

## Buscar
- **Búsqueda global**: con el cursor en la lista, teclear el inicio del nombre (p. ej. «CON»).
- **Búsqueda local**: tras la global, **F8** abre tres casillas para hasta tres cadenas que el nombre contenga, en cualquier orden.

## Crear una partida en el catálogo
Por *Catálogos > Partidas* o con clic derecho desde la hoja. Los grupos se crean como sub-ítems del árbol. En la ventana **Partidas**:
- Código autogenerado.
- Descripción: hasta 250 caracteres.
- Descripción alterna, por ejemplo en inglés.
- Unidad y Especialidad.
- **Grupo**: *Partida* o *Estimada*.
- Estado: *Sin análisis* al crearla.
- Jornada: 8 h por defecto.
- Peso: para electromecánicas.

También sirve *Duplicar* (solo la partida, o la partida con su APU).

## Rendimiento
Es el rendimiento diario de mano de obra y de equipo (iguales por defecto). *Cuadrillas* modifica la duración y el rendimiento.

**Rendimiento para transporte** equilibra el cargador frontal y los volquetes. Se ingresan distancia, velocidades, tiempos de carga y descarga, eficiencia (1-100), esponjamiento, capacidad del volquete y número de volquetes, y se recalcula hasta aceptar.

## Registrar recursos en el APU
Con el cursor en el APU, clic derecho:
- **Adicionar recurso**: en mano de obra y equipo se ingresa la *cuadrilla*; en materiales, la cantidad por unidad; en herramientas manuales, un %.
- *Adicionar recurso del presupuesto*.
- *Adicionar subpartida*.
- *Adicionar plantilla*.
- *Copiar análisis de otra partida*: solo si el APU está vacío.

Un recurso nuevo se crea en *Catálogos > Recursos* (ejemplo: «PIEDRA DE HUAMANGA», m3). *Reasignar recurso* cambia un recurso por otro. El APU se puede *Exportar a Microsoft Excel*.

**Estados del APU**: sin análisis, sin revisión, análisis copia, revisado y protegido.

**Confiabilidad:** copia de tercero sin verificar; puede no corresponder a la versión actual. El sílabo oficial confirma los temas de catálogo, APU y rendimiento para transporte.

## Fuentes
- guia-de-usuario-de-s10-presupuestos.pdf, págs. 7, 53-63
- manual-de-s10-costos-y-presupuestos.pdf, págs. 7, 62-75, 107-116
- silabus-presupuestos.pdf, pág. 2
