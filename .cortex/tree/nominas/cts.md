---
title: Cálculo de la CTS
summary: Los conceptos de CTS se activan solos en mayo y noviembre. Se procesa
  el Resumen de concepto por periodo del mes de pago, se calcula con el proceso
  estándar (con estado Procesar si no hay tareo), se valida en Vistas, se emiten
  las boletas y se genera un documento de pago por banco.
tags:
  - nominas
  - cts
  - beneficios-sociales
  - procedimiento
  - oficial
id: 01M3RJP5Y8HK9FK9ERP2RMHEGD
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.328Z
---

El sistema **activa los conceptos de CTS automáticamente en mayo y noviembre**. En el ejemplo del video el depósito vence el **15 de noviembre** y se acumula de mayo a octubre, así que no hace falta el tareo de noviembre.

**Procedimiento**
1. Ejecutar el escenario **Resumen de concepto por periodo** con el mes de pago.
2. Calcular la nómina con el **proceso estándar**.
3. Como aún no hay tareo del mes, marcar a todos los trabajadores y, con clic derecho, cambiar el estado a **Procesar**.
4. Validar en el escenario **Vistas**, con el reporte de CTS. Muestra las variables que propone el sistema y los **últimos cuatro sueldos**, que se informan al banco en el depósito. Se puede exportar a Excel.
5. Preparar la **boleta de liquidación de CTS**, que detalla conceptos y variables.
6. Verificar e imprimir las boletas. Necesitan el **check de impresión**, y la fecha de impresión en el ejemplo es como máximo el 15 de noviembre. Entregarlas a cada trabajador.
7. Enviar el monto por **planillas de pago**. El sistema emite **un documento de pago por banco**.

**Puntos de control**
- Cada trabajador necesita una **cuenta bancaria de CTS**.
- El pago es el acumulado de las **provisiones** mensuales del semestre. Revisarlas mes a mes en los reportes.

**Temario de la capacitación oficial de CTS:** aspectos generales, conceptos remunerativos y no remunerativos, tiempo computable, trabajadores con derecho, oportunidad de pago, libre disponibilidad, provisión y pago, cálculo en S10, boletas de liquidación, documentos por banco y declaración en la PLAME. La carta de retiro de CTS se arma como plantilla de documento en Nóminas.

**Confiabilidad:** oficial S10 (video del canal oficial y temario). Las fechas legales salen del video y no indican año: verificar la norma vigente.

**Fuentes:** video XObV2BOrsIc «Cómo calcular la CTS con el S10 ERP» (0:08-6:54); data/texto/temario-capacitacion-de-cts.txt (pág. 1); data/texto/silabus-nominas-avanzado.txt (días 4-5).
