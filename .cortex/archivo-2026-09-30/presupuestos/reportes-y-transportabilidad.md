---
title: Reportes, exportación y traslado de la base de datos
summary: Reportes desde Datos Generales (resúmenes) y desde la Hoja (estándar
  interno/cliente, desagregado, APU, recursos). Vista preliminar exporta a
  Word/Excel. Transportabilidad exporta presupuestos a otra base; Utilitarios
  hace copias de seguridad para llevarlas a otra PC.
tags:
  - presupuestos
  - reportes
  - exportacion
  - copia-de-seguridad
  - tercero
  - oficial
id: 01M3RJP5W0BK598QHXMKYN8TCE
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.256Z
---

## Imprimir
Hay dos botones: **Imprimir** (directo a la impresora de Windows) y **Vista preliminar** (en pantalla; desde ahí se exporta a Word, Excel, etc.).

**Desde Datos Generales:**
- *Datos generales*.
- *Resumen del presupuesto estándar*: montos globales por subpresupuesto.
- *Resumen desagregado*: con mano de obra, materiales, equipo y subcontratos.

**Desde la Hoja del Presupuesto** (según el cursor esté en el presupuesto o en un subpresupuesto):
- **Estándar interno**: muestra el código de catálogo de la partida.
- **Estándar cliente**: el que se entrega al dueño de la obra.
- **Desagregado precios unitarios**: formato horizontal por tipo de recurso.
- **Análisis de precios unitarios**: estándar, *Formato entidades*, *Resumen de partidas iguales* (descripción propia o del catálogo) y consolidado.
- **Recursos y precios** del subpresupuesto. Para las subpartidas se elige *detallado*. Todos los recursos de un presupuesto con varios subpresupuestos se imprimen desde el escenario **Precios**.

Otros:
- *Diseño de cabeceras para reportes*: hasta 10 líneas.
- Membrete y logotipo.
- Impresión con **ítem alterno**.
- **Presupuesto descompuesto**, exportable a Excel.
- *Tiempos para programación*: duración = metrado / rendimiento / cuadrillas.
- **Exportar a** MS Project o Excel: la información no vuelve a S10.

## Exportar presupuestos (Transportabilidad)
El asistente *Exportación de presupuestos* pide los presupuestos, los recursos, partidas y títulos, y el **nombre de la base de datos destino**, que crea en el servidor en uso. Para entrar a esa base: *Archivo > Iniciar sesión como usuario distinto*.

## Llevar la base a otra PC
1. En *Utilitarios > Mantenimiento de base de datos*, **Crear copia de seguridad** (también existe *Compactar y eliminar*).
2. Copiar el archivo de la carpeta `S102000\backup` a un medio externo.
3. En la PC destino, que debe tener la misma versión de S10, pegarlo en la misma carpeta y restaurarlo.

Para la copia compacta debe existir `b2k.exe` en `S102000\backupS10`.

## Migrar desde DOS
**Importar datos DOS 7.x 8.x** pasa las DBF a SQL Server; requiere el archivo `modeldos`. Access se importa de forma similar.

**Confiabilidad:** mixta. El sílabo oficial menciona reportes y copias de seguridad y restauración. Los menús, rutas y asistentes solo aparecen en la guía de tercero y **pueden no corresponder a la versión actual**; las rutas `S102000` son de una versión antigua.

## Fuentes
- guia-de-usuario-de-s10-presupuestos.pdf, págs. 51-52, 66, 110-112 y 117-124
- manual-de-s10-costos-y-presupuestos.pdf, págs. 60-62, 88 y 117-121
- silabus-presupuestos.pdf, pág. 1
