---
title: "Construcción civil: aporte al Fondo de Capacitación"
summary: Para el aporte del Fondo de Capacitación se crean un recurso y dos
  conceptos (tarifa y cálculo REDONDEAR(tarifa × días asistidos)). Se suman a
  Total aportaciones en las plantillas semanal y de reintegros, se muestran en
  boletas y se genera un documento de pago a la Cámara Peruana de la
  Construcción.
tags:
  - nominas
  - construccion-civil
  - aportes
  - conceptos
  - oficial
id: 01M3RJP5Y5M19N0XZDX5N894PZ
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.325Z
---

Configuración de un **aporte nuevo** en la planilla de construcción civil. Sirve de modelo para crear cualquier concepto nuevo.

1. En el **catálogo de recursos**, crear el recurso «Aportes obreros fondo capacitación».
2. Crear dos conceptos:
   - **Tarifa del fondo:** tipo de dato *constante matricial* y tipo de valor *moneda*. No se muestra en la boleta.
   - **Fondo capacitación:** tipo de dato *general* y tipo de valor *moneda*. Se muestra en la boleta y **contabiliza**.
3. Agregar los dos conceptos a la **plantilla de fórmulas** de la planilla semanal, antes de *Total aportaciones*.
4. Formular *Fondo capacitación* con la función **REDONDEAR**: tarifa × variable del concepto de **días asistidos**.
5. Sumar *Fondo capacitación* en la fórmula de *Total aportaciones*. Si hay dos conceptos de total de aportes, incluirlo en los dos. Grabar la plantilla y salir.
6. Repetir los pasos 3 a 5 en la plantilla de **reintegros retroactivos**, con la misma fórmula.
7. En **Reportes**, agregar el concepto al reporte de la planilla semanal. Marcar el check **Aplica** en la *boleta de pago semanal* y en la *boleta de pago de reintegros*.
8. En la **Matriz de Constantes**, registrar la tarifa. El video dice que el convenio fija **0.20** por día asistido.
9. Agregar un **documento de pago** para el aporte, que se paga a la **Cámara Peruana de la Construcción**.
10. Calcular la planilla. Verificar que el concepto salga en reportes y boletas, y que al enviar a Facturación se genere la planilla de pago del aporte.

**Confiabilidad:** oficial S10 (video del canal oficial). La tarifa de 0.20 es la del convenio vigente cuando se publicó el video, que no indica el año: verificar la cifra actual.

**Fuentes:** video NYUDXu-XmyA «S10 Nóminas - Nuevo Aporte Fondo de Capacitación en Construcción Civil» (0:10 recurso; 0:31-2:02 conceptos; 2:19-4:46 plantillas; 5:05-5:48 reportes y boletas; 6:19-7:35 tarifa y documento de pago; 7:55-8:36 verificación).
