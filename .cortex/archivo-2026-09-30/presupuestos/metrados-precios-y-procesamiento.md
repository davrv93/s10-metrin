---
title: Metrados, precios y procesamiento
summary: Ingresar metrados en la hoja (o vincularlos a Excel) → procesar con F9
  → cargar los precios faltantes (sin IGV) en Recursos y precios → volver a
  procesar. El tipo 3 hace coincidir el costo directo con el total de recursos.
tags:
  - presupuestos
  - metrados
  - procesamiento
  - precios
  - tercero
id: 01M3RJP5VQNJBZKS1RZ665NJ9P
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.247Z
---

## Metrados
Se escriben directamente en la columna **Metrado** de cada partida. Los decimales se definen en *Datos adicionales*.

Formas alternativas:
- **Pegar de otras aplicaciones**: actualiza metrados con valores copiados de Excel.
- **Procesos especiales > Vincular metrado a celda Excel**, y después *Tomar metrado de celda Excel* o *Desvincular*.
- **Pegar presupuesto** con multiplicador, para obras modulares.

*Agrupar partidas iguales* suma los metrados de una partida usada en varios ítems.

## Configuración del procesamiento
Botón de configuración o **Shift+F1** > pestaña *Procesamiento de Presupuesto*. Solo sa o quien tenga atributos de supervisor la ve, y rige para toda la empresa.
- **Tipo 1**: multiplica cantidad × precio y redondea en cada recurso, así que el monto de la hoja puede no coincidir con el total de recursos.
- **Tipo 3**: corrige las desviaciones para que coincidan y calcula consolidados por partida y por título.

Verificaciones que se pueden activar: metrados, APU, precios de recursos, cantidades en APU y porcentaje al 100 % de las estimadas. Mientras el presupuesto está en elaboración se recomienda desactivarlas.

## Procesar
1. Botón *Procesar* o **F9** > **Continuar**. El sistema calcula los precios unitarios, los costos por tipo de recurso, el costo directo e indirecto y el consolidado de recursos.
2. Si faltan datos, pregunta, por ejemplo, «¿Ingresará los 28 precios faltantes?» > **Sí**.
3. En la ventana **Recursos y precios**, las flechitas marcan los recursos sin precio. Ingresar los precios **sin IGV** y cerrar.
4. Volver a procesar.

El precio de un recurso se guarda **por lugar, fecha y presupuesto**.

## Resultado del procesamiento
La ventana muestra estadísticas (ítems faltantes y verificados, metrados, APU, precios, recursos y subpartidas con cantidad cero) y costos: directo, indirecto, total, mano de obra, materiales, equipos y subcontratos.

Si aparece en rojo **«El Costo Directo difiere del total de recursos!»**, hay que subir los decimales de precios o de incidencias y procesar hasta que desaparezca.

## Otras herramientas de la hoja
Procesar partida, **Calcular incidencias**, Sumar títulos, filtros (propias, ajenas, sin análisis, con vínculo Excel…), Reasignar recurso, Eliminar precios diferentes en subpresupuestos y **Factores de rendimiento** (temporales, sobre partidas propias).

**Confiabilidad:** copia de tercero sin verificar; puede no corresponder a la versión actual. El sílabo oficial confirma «Registro de metrados», «Procesamiento del presupuesto» e «Ingreso de precios».

## Fuentes
- guia-de-usuario-de-s10-presupuestos.pdf, págs. 30-37
- manual-de-s10-costos-y-presupuestos.pdf, págs. 34-41, 85-91, 99-101 y 105-106
- syllabus-presupuestos-s10erp.pdf, pág. 1
