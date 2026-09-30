---
title: Secuencia de uso y registro del presupuesto
summary: "Orden: registrar el presupuesto y sus subpresupuestos en Datos
  Generales → armar partidas → metrados → procesar → precios → pie → gastos
  generales → procesar → fórmula polinómica → planeamiento → reportes. Al
  cambiar lugar o fecha, los precios quedan en cero."
tags:
  - presupuestos
  - datos-generales
  - flujo
  - registro
  - tercero
id: 01M3RJP5VBDWBQY9VXTGC10C94
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.235Z
---

## Secuencia de uso (guía, 1.02)
1. **Datos generales**: registrar el presupuesto y sus subpresupuestos.
2. **Hoja del presupuesto**: registrar partidas, luego metrados, procesar, ingresar el precio de los recursos.
3. Registrar el **diseño del pie de presupuesto**.
4. **Cálculo analítico de gastos generales**.
5. Volver a procesar y elaborar la **fórmula polinómica**.
6. **Planeamiento**: proyecto, calendario, WBS, períodos, exportar e importar de MS Project.
7. **Reportes**.

## Registrar un presupuesto nuevo
1. Entrar al escenario **Datos Generales**, clic derecho en el árbol > **Nuevo**.
2. En el **Catálogo de Presupuestos**, elegir o crear un grupo con *Nuevo SubItem* (p. ej. OBRAS VIALES > PUENTES) y, dentro, *Nuevo SubItem* otra vez para el presupuesto.
3. Completar la ventana **Presupuesto**:
   - Descripción.
   - **Cliente**, desde el catálogo de identificadores; si no existe, se crea con tipo *Cliente*.
   - **Ubicación geográfica**: distrito.
   - **Fecha**: los precios se guardan por fecha y lugar.
   - **Plazo** en días calendario: dato informativo.
   - **Jornada diaria**: horas, influye en el rendimiento.
   - **Doble moneda** y **Moneda base**.
   - *Presupuesto Base*: editable e informativo, para licitaciones.
4. En **Datos adicionales** van los decimales de precios, incidencias (al menos 2 más que los precios) y metrados, la casilla *Fórmula Polinómica*, la casilla *APU tipo 2* (con check calcula para carreteras, sin check para edificaciones) y el factor de cambio para imprimir en moneda alterna.
5. **Adicionar**, y doble clic para llevarlo al árbol.
6. Sobrescribir el nombre del subpresupuesto (p. ej. ESTRUCTURAS). Se añaden más con clic derecho > *Adicionar Subpresupuesto*.

## Cuidado con lugar y fecha
Si se cambia el lugar o la fecha, **los precios quedan en cero**. Para recuperarlos hay que volver al lugar y la fecha que tenían precios y procesar. La opción **Modificar fecha/lugar no precios** conserva los precios y deja rastro en *Histórico*.

## Otras opciones de Datos generales
Copiar/Pegar presupuesto (permite sumar varios o multiplicar los metrados, por ejemplo por 4 baterías de baños), **Duplicar** (sirve para generar el Meta), Replicar en otro grupo, Membrete, Logotipo, Grupo de trabajo y Enviar a.

**Confiabilidad:** copia de tercero sin verificar; puede no corresponder a la versión actual. El sílabo oficial confirma el tema «Registro del presupuesto».

## Fuentes
- guia-de-usuario-de-s10-presupuestos.pdf, págs. 5 y 10-18, 67-71
- manual-de-s10-costos-y-presupuestos.pdf, págs. 12-22 y 79-81
- silabus-presupuestos.pdf, pág. 2
