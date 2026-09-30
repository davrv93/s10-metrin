---
title: Pie de presupuesto, gastos generales y fórmula polinómica
summary: El pie calcula lo que va bajo el costo directo con variables (nDirecto,
  C_J/C_M/C_E/C_S, GGP, P_T). GGP en Macro habilita los gastos generales
  analíticos. La fórmula polinómica agrupa índices unificados en hasta 8
  monomios de al menos 0.05, omitiendo el IGV.
tags:
  - presupuestos
  - pie-de-presupuesto
  - gastos-generales
  - formula-polinomica
  - tercero
id: 01M3RJP5VWC2PV17CQAXTBSWM9
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.252Z
---

## Diseño del pie de presupuesto
Es todo lo que va debajo del **COSTO DIRECTO** (gastos generales, utilidad, impuestos). Se hace después de procesar, en el escenario **Diseño Pie de Presupuesto**, eligiendo el presupuesto en el árbol. La opción *Diseño para todos los Subpresupuestos* lo aplica a todos.

Columnas: N° línea, Descripción, **Variable**, **Macro** y *Omitir Polinómica*.

Ejemplo de la guía para Venta:

| Descripción | Variable | Macro |
|---|---|---|
| Gastos generales | GG | nDirecto*0.15 (o nDirecto*GGP) |
| Utilidad | UTI | nDirecto*0.10 |
| Subtotal | ST | nDirecto+GG+UTI |
| IGV | IGV | ST*0.19 (tasa de la época) |
| Total presupuesto | P_T | ST+IGV |

- Variables **obligatorias**: `nDirecto` (costo directo) y `P_T`, que convierte el total a letras.
- Variables **auxiliares**: `C_J` mano de obra, `C_M` materiales, `C_E` equipos, `C_S` subcontratos.
- **GGP** en Macro es obligatorio para usar gastos generales analíticos.
- Para el Meta, la guía muestra un pie solo con costo directo y total.

Después hay que volver a procesar.

## Gastos generales analíticos
Cubren lo que no está en los APU (sueldo del residente, alquiler de oficina, teléfono, créditos). Se recomiendan para el presupuesto de venta.
1. En *Gastos Generales*, clic derecho > **Adicionar rubro**, desde el Catálogo de Rubros (personal profesional, personal técnico, alquiler de equipo menor…).
2. Por cada rubro, **Adicionar concepto** con unidad, personas, % de participación, tiempo y sueldo o jornal.
3. **Calcular porcentaje**: rubros ÷ presupuesto, que el procesamiento toma como GGP.
4. Procesar y pasar a la fórmula polinómica.

## Fórmula polinómica
Calcula el reajuste **K** con los índices unificados del INEI. Reglas que cita la guía:
- Hasta **8 monomios**.
- Coeficiente mínimo **0.05**, al milésimo.
- Los coeficientes suman 1.
- Un monomio compuesto admite hasta 3 elementos (sub-monomios).

Pasos:
1. **Omitir** el IGV en el pie y hacer la fórmula sobre el presupuesto de venta.
2. En el árbol, elegir el subpresupuesto > *¿Desea elaborar la Fórmula Polinómica?* > **Sí**. Antes, asignar índice unificado a los recursos que no lo tengan.
3. **Agrupamiento preliminar**: arrastrar un índice sobre otro para agruparlo; se deshace con clic derecho. Agrupar semejantes hasta llegar al 5 % y, de preferencia, no agrupar el **índice 47** (mano de obra).
4. **Conformación de monomios**: escribir el mismo número en la columna *Monomio* para juntar dos o tres.
5. **Reemplazo de símbolos** y verificar.

**Confiabilidad:** copia de tercero sin verificar (IGV 19 %: versión antigua); puede no corresponder a la versión actual. El sílabo oficial confirma los tres temas.

## Fuentes
- guia-de-usuario-de-s10-presupuestos.pdf, págs. 38-50
- manual-de-s10-costos-y-presupuestos.pdf, págs. 46-59
- syllabus-presupuestos-s10erp.pdf, pág. 1
