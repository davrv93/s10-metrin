---
title: Asignar código de producto Sunat
summary: Debido a la inclusión de nuevas normativas obligatorias por parte de
  SUNAT para el año 2020, S10 está buscando alinearse al incluir un nuevo
  catálogo para “Código de Productos SUNAT según el estándar UNSPSC” y un
  escenario para catalogar los recursos existentes en el sistema…
tags:
  - oficial-s10
  - manual
  - portal-miembros
id: 01M3SP286ZESARNMKP62MB0AXS
status: active
updated_by: ai-agent
updated_at: 2026-09-30T18:09:53.562Z
---

Manual oficial de S10, bajado del portal de miembros con la cuenta de la empresa.

- **Página:** https://documentacion.s10peru.com/asignar-codigo-de-producto-sunat/

## Contenido

- 1. Configuración:
- 13.1 Registro de Inventario Permanente Valorizado
- 2. Asignación de Códigos de Producto SUNAT
- 3. Uso de Códigos de Producto SUNAT en Libros Electrónicos
- 13.01 Registro del Inventario Permanente Valorizado

## Texto de la página

# Asignar código de producto Sunat

Fuente: https://documentacion.s10peru.com/asignar-codigo-de-producto-sunat/

Debido a la inclusión de nuevas normativas obligatorias por parte de SUNAT para el año 2020, S10 está buscando alinearse al incluir un nuevo catálogo para “Código de Productos SUNAT según el estándar UNSPSC” y un escenario para catalogar los recursos existentes en el sistema según estos códigos internacionales.
1. Configuración:
Table of Contents
Toggle
Para iniciar con el uso de estos códigos de productos SUNAT, se debe activar la configuración “Código de Producto SUNAT Obligatorio en Libros Electrónicos”
Esta configuración hará que se valide que todos los recursos usados en los libros electrónicos (que usen los códigos de productos SUNAT), tengan un código asignado, de no tenerlo se mostrará como una observación, los libros afectados serían:
13.1 Registro de Inventario Permanente Valorizado
Libros de Inventarios y Balances
2. Asignación de Códigos de Producto SUNAT
El escenario para catalogar los recursos de S10 con los códigos de Producto SUNAT, se encuentra en el módulo de Almacenes, en el grupo Configuración.
En este escenario podremos ver la lista de todos los recursos existentes que estén Activos en el S10, y podremos asignarles el código de producto SUNAT que le corresponda.
Asignar y Quitar Código de Producto SUNAT a Recurso S10
a) Ubicarse en el escenario y pulsar clic derecho sobre el recurso al que desee asignar el código SUNAT. (Se puede seleccionar varios a la vez y asignar el código SUNAT)
(Se puede seleccionar varios a la vez y asignar el código SUNAT, solo debe sombrearse varios recursos a la vez y pulsar clic derecho en esta selección)
b) Seleccionar el código de Producto SUNAT del catálogo correspondiente con doble clic. (La información del catálogo fue incluida por medio de una actualización al S10)
Importar Código de Producto SUNAT a Recurso S10
Al igual que en otros escenarios, también se cuenta con una opción para importar y así asignar de forma masiva el Código de Producto SUNAT a los recursos S10.
a) Para ello debe usar la opción “Importar de Microsoft Excel”
b) Considere que para importar, requerirá preparar previamente un Excel con la información necesaria, los únicos campos que debe preparar son:
Código Recurso: se toma del “Catálogo de recursos”
Código producto Sunat: se toma del “Catálogo de Productos SUNAT” (se sugiere exportar la información de este catálogo al usar la opción “Asignar”.
c) Con el Excel preparado ya podrá usarlo en la opción de “Importar de Microsoft Excel”
3. Uso de Códigos de Producto SUNAT en Libros Electrónicos
13.01 Registro del Inventario Permanente Valorizado
a) El campo correspondiente al código es el campo “8.Código Producto SUNAT”
b) Si se tiene la configuración (1. Configuración) y aún no se ha asignado ningún código de producto SUNAT, se mostrará como observación el mensaje “El Campo ‘Código Producto SUNAT’ no puede ser Nulo, verifique la configuración.”
Si este es el caso, debe asignar el código en el escenario de Almacenes (2. Asignación de código de Producto SUNAT)
03.07 Libro de Inventarios y Balance – Detalle del saldo de la cuenta 20 – Mercaderis y la cuenta 21
a) El campo correspondiente al código es el campo “Cod. CUBSO” y la descripción del código es “CUBSO”

El texto completo también está troceado en `kb/fragmentos.jsonl` y se cita con el nombre del manual y la página.

**Confiabilidad:** oficial S10
