---
title: Constantes por Fecha General (S10 Nóminas) · validada
summary: "Pantalla de S10 Nóminas, módulo General: grilla de conceptos (UIT,
  RMV) y grilla de valores por fecha. Clic derecho > Adicionar para registrar un
  valor nuevo con fechas, monto y observación. Validada a ojo sobre el video
  oficial."
tags:
  - ui
  - pantalla
  - nominas
  - validado
  - oficial
id: 01M3RRCP5TD9G5NE2AE7WQBQFW
status: active
updated_by: ai-agent
updated_at: 2026-09-30T09:02:43.832Z
---

**Pantalla validada a ojo** sobre las 60 capturas del tutorial oficial [«S10 Nóminas - Cambio de UIT y RMV (2025)»](https://youtu.be/17W0yKI8iew), revisión del 30-09-2026. Es **una sola pantalla**; las capturas muestran sus estados: grilla, menú contextual, calendario y valores cargados.

### Dónde está
- Aplicación **S10 Nóminas** · menú `Archivo · Ver · Catálogos · Herramientas · ?`
- Panel lateral de módulos: Inicio, Nominas, Informes, Organización, Configuración, **General**, Actividades, Pagos, Gestión RRHH, Asistencia, Utilitarios.
- Módulo **General** → escenario **«Constantes por Fecha General»** (título gris de la ventana).
- Encima de la grilla: «Arrastre hacia aquí el encabezado de la columna que desea agrupar» (agrupación por columna).

### Grilla superior: conceptos
Barra de búsqueda: contador de registro (`1/2`), **Campo** (lista, por defecto «Grupo»), modo **«En cualquier posición»** (lista), cuadro de texto, botones limpiar (X) y buscar (lupa).

| Columna | Valores vistos |
|---|---|
| Grupo | 02 CONCEPTOS EMPLEADOS · 04 CONCEPTOS GENERALES |
| Subgrupo | 02010 VARIABLES AUXILIARES EMPLEADOS · 04001 DATOS FIJOS Y PORCENTUALES |
| Código | 02010008 · 04001016 |
| Concepto | UIT · REMUNERACION MINIMA VITAL |
| Abreviatura | UIT · REMUNERACION MINIMA VITAL |

### Grilla inferior: valores por fecha del concepto seleccionado
| Columna | Tipo | Nota |
|---|---|---|
| Fecha inicio | fecha | botón «…» abre un calendario con «Today» |
| Fecha fin | fecha | puede quedar vacía si el valor sigue vigente |
| Valor | número | alineado a la derecha |
| Observaciones | texto | p. ej. «UIT 2025» |

**Clic derecho** sobre la grilla inferior: **Adicionar** (nueva línea) y **Eliminar**.

### Procedimiento que muestra el video
1. En **Constantes por Fecha General**, seleccionar el concepto **UIT** en la grilla superior.
2. Clic derecho en la grilla inferior → **Adicionar**.
3. Elegir **Fecha inicio** y **Fecha fin** con el calendario, escribir el **Valor** y una **Observación**.
4. Seleccionar **REMUNERACION MINIMA VITAL**: primero poner la **Fecha fin** de la línea anterior, luego **Adicionar** la nueva con su fecha de inicio, sin fecha fin, y el valor vigente.
5. Si estos conceptos no aparecen en el escenario, comunicarse con soporte.

### Valores que aparecen (ejemplo del video, enero 2025)
- UIT: 4300 (2020) · 4400 (2021) · 5150 (2024) · **5350 (2025)**
- RMV: 1025 (desde 01/01/2024, cerrado al 31/12/2024) · **1130 (desde 01/01/2025, sin fecha fin)**

Son las cifras que registró quien grabó el tutorial: **vigentes para 2025; verificar las del año en curso**.

### Correcciones al borrador de OCR
- Descartados: «Cempo[Grupo» (era la etiqueta «Campo»), «Tema» (es el botón del panel lateral, no un campo), «Tipo Valor» (no aparece en pantalla) y 18 filas «04/01xx» (lecturas falsas de la grilla).
- Títulos de ventana mal leídos por la máquina («$10 Nóminas», «510 Nóminas», «S10 Nomina», «Datos=NOMINAS_FINAL»): son la barra de título y la barra de estado de esta misma pantalla. La barra de estado muestra servidor y base de datos del equipo de quien grabó; no son datos del producto.

Fuentes: tutorial oficial https://youtu.be/17W0yKI8iew (capturas 0001–0060 en `data/ui/capturas/17W0yKI8iew/`) y su transcripción.

**Confiabilidad:** oficial S10 (pantalla real del video oficial), validada por inspección visual.
