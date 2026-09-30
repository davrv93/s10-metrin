---
title: Conceptos y estructura del presupuesto
summary: Presupuesto > subpresupuestos > títulos > partidas. Cada partida tiene
  un análisis de precios unitarios (APU) con mano de obra, materiales, equipos y
  subcontratos. Las partidas pueden ser del catálogo, propias, formato o
  estimadas.
tags:
  - presupuestos
  - partidas
  - subpresupuesto
  - apu
  - tercero
id: 01M3RJP5V3HTGT569N49GTDK1S
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.227Z
---

## Jerarquía
- **Presupuesto**: la obra (p. ej. «Puente en carretera»). Lleva cliente, lugar, fecha, plazo, jornada y moneda.
- **Subpresupuesto**: una parte del presupuesto (p. ej. «001 Estructuras»). La fórmula polinómica y varios reportes van por subpresupuesto.
- **Títulos**: los agrupadores de la hoja (CONCRETO ARMADO > ENCOFRADO). Salen del **catálogo de títulos** y se jerarquizan con los botones de indentado.
- **Partidas**: las líneas con unidad, metrado, precio y parcial. Salen del **catálogo de partidas**.
- **Ítem**: la numeración (01.01.02). La genera el sistema con el botón *Generar ítems*. No es editable; para una numeración propia se usa la columna **ítem alterno**.

## Partes de la hoja impresa
La **cabecera** (presupuesto, subpresupuesto, cliente, lugar, fecha), el **cuerpo** (ítem, descripción, unidad, metrado, precio, parcial) y el **pie**: COSTO DIRECTO, cargos y TOTAL PRESUPUESTO en letras.

## Análisis de precios unitarios (APU)
Es la estructura de costo de una partida. Se compone de **recursos** y/o **subpartidas**, de cuatro tipos: mano de obra, materiales, equipos y subcontratos. Los recursos «comodín» tienen unidad en % y su precio depende de otros.

## Tipos de partida
- **Del catálogo**: compartida. Un cambio afecta a todos los presupuestos que la usan.
- **Propia**: copia exclusiva del presupuesto. Se recomienda activar en Configuración «Hacer propio en forma automática partidas principales / subpartidas».
- **Formato**: solo una línea descriptiva, sin APU (así llegan las partidas importadas de Excel). Sirve para valorizar.
- **Estimada**: su APU se reparte en porcentajes de grandes grupos, que deben sumar 100 %.

## Tipos de presupuesto
El manual nombra **Venta (oferta), Meta y Línea Base**. La guía recomienda hacer los gastos generales analíticos y la fórmula polinómica sobre el de venta u oferta. El Meta no lleva gastos generales analíticos; «generalmente se efectúa un presupuesto detallado».

## Estado en el árbol (color del librito)
Verde: en elaboración. Ámbar: procesado sin fórmula polinómica. Rojo: procesado y concluido. Libro abierto: en uso.

**Confiabilidad:** copia de tercero sin verificar. Solo los tipos «base, real, oferta, meta y línea base» figuran en el sílabo oficial; el resto puede no corresponder a la versión actual.

## Fuentes
- guia-de-usuario-de-s10-presupuestos.pdf, págs. 4, 9, 13, 20, 36 y 40
- manual-de-s10-costos-y-presupuestos.pdf, págs. 2, 30, 87-88 y 107
- syllabus-presupuestos-s10erp.pdf, pág. 1
