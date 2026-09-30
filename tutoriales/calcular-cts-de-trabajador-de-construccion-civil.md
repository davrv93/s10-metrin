# Calcular CTS de trabajador de construcción civil

## Para qué sirve

La CTS es un beneficio social derivado de una obligación legal, cuya finalidad es cubrir las contingencias que origina el cese del trabajador [F17]. En S10 ERP el cálculo se hace desde el módulo Nóminas: el sistema activa los conceptos de CTS de forma automática en los meses de mayo y noviembre, y considera los días acumulados del semestre (por ejemplo, de mayo a octubre para el pago de noviembre) [F19]. El resultado es la boleta de liquidación de CTS y un documento de pago por cada banco [F22].

## Antes de empezar

- Cada trabajador debe contar con una cuenta bancaria de CTS [F22].
- Los jornales por categoría (operario, oficial, peón) deben estar actualizados en el concepto **Jornal diario** de la Matriz de Constantes [F9][F10].
- Debe existir el reporte de CTS en el escenario **Vistas** [F19][F22].
- La UIT y la Remuneración Mínima Vital deben estar registradas en el escenario **Constantes por fechas** del grupo General [F18][N1].
- Si los conceptos de CTS o las constantes no aparecen en el escenario, solicitar la configuración al soporte de S10 [F18][N1].

## Pasos

1. **Abrir el módulo Nóminas y ubicar el periodo de pago.**
   1. Ingresar al módulo Nóminas [F18].
   2. Ejecutar el escenario **Resumen de concepto por periodo** [F19].
   3. Seleccionar el mes en el cual se procederá al pago de la CTS [F19].
   4. La selección de empresa u obra en esta ventana: **No documentado en las fuentes disponibles**. Consultar el Portal de Ayuda de S10 o abrir un ticket de soporte [F6].

2. **Verificar el registro del trabajador en el régimen de construcción civil.**
   1. Confirmar que el trabajador figure en el personal disponible del proyecto multiproyecto y del proyecto destino [F3][N3].
   2. Confirmar la categoría del trabajador (operario, oficial, peón), porque de ella depende el jornal diario aplicado [F9][F10].
   3. La ventana exacta donde se valida la categoría y la fecha de ingreso del trabajador: **No documentado en las fuentes disponibles**. Consultar el manual de Nóminas o el soporte de S10 [F6].

3. **Revisar los jornales y conceptos que alimentan el cálculo.**
   1. Abrir el escenario **Matriz de Constantes**, agrupado por **Categoría** [F9][F10].
   2. Verificar el valor del concepto **Jornal diario** por categoría. Los valores del video son 74.30 para operario, 58.45 para oficial y 52.50 para peón [F9][F10].
   3. **Cifra legal:** estos jornales corresponden al convenio colectivo mostrado en el video, que no indica el año [F9][F10][N7]. Verificar el jornal del convenio vigente antes de calcular.
   4. Para la bonificación por alta especialización, cambiar la agrupación superior de **Categoría** a **Ocupación** y revisar la ocupación **Operario electromecánico**; en el convenio del video el porcentaje pasó del 15 % al 18 % [F9][F10].
   5. La lista de conceptos remunerativos y no remunerativos computables para CTS es tema de la capacitación oficial de CTS [F17], pero su detalle (jornal básico, BUC, asignación familiar, sobretiempo habitual) **no está documentado en las fuentes disponibles**. Consultar el temario o la grabación de esa capacitación [F17].

4. **Validar la fórmula del concepto CTS y el periodo de devengue.**
   1. No configurar la fórmula manualmente: el sistema está preparado para activar los conceptos de CTS de forma automática en mayo y noviembre [F19].
   2. Confirmar que el sistema considere los días acumulados del semestre. En el ejemplo del video se acumula de mayo a octubre [F19].
   3. Los nombres de campos y la sintaxis de la fórmula del concepto CTS: **No documentado en las fuentes disponibles**. Para conceptos nuevos en general, la plantilla de fórmulas se edita como se muestra para el aporte al Fondo de Capacitación [F12][N8]; confirmar con soporte si se requiere ajustar la fórmula de CTS.

5. **Cargar los jornales y días laborados del periodo.**
   1. Para la CTS no es necesario tener el tareo del mes de pago, porque el sistema solo considera los días acumulados del semestre anterior [F19].
   2. Si se requiere el tareo, registrarlo desde el **tareo multiproyecto** y procesar el escenario **Resumen de concepto por día** [F3][N3].
   3. Refrescar el escenario **Resumen de concepto por periodo** para el mes de pago [F19].

6. **Ejecutar el cálculo.**
   1. En el grupo **Nóminas**, calcular la planilla con el **proceso estándar** [F19][N3].
   2. Al no existir tareo del mes, el sistema muestra a los trabajadores sin procesar. Marcar a todos los trabajadores [F19].
   3. Con clic derecho, cambiar el estado a **Procesar** [F19].
   4. Ejecutar el cálculo y esperar a que concluya el proceso [F19].

7. **Revisar el reporte y validar montos.**
   1. Abrir el escenario **Vistas** y seleccionar el reporte de CTS [F19][F22].
   2. Verificar que las variables que propone el sistema sean las correctas [F19].
   3. Revisar el valor de los **últimos cuatro sueldos**, dato que se informa al banco al momento del depósito [F22].
   4. Exportar la información a Excel para un análisis más detallado [F19][F22].
   5. Verificar las provisiones mensuales mediante reportes: el pago de la CTS es el acumulado de las provisiones del semestre completo [F22].

8. **Emitir boletas y generar el pago por banco.**
   1. Preparar la boleta de CTS [F22].
   2. Verificar la boleta. Debe contar con el **check de impresión** [F22].
   3. Emitir y entregar la boleta a cada trabajador. La boleta de CTS detalla los conceptos y variables involucrados [F22].
   4. Enviar el monto a pagar mediante las **planillas de pago** [F22].
   5. El sistema emite un **documento de pago por cada banco** [F22].
   6. Opcionalmente, publicar la CTS en el **Portal del Empleado**, que admite el envío de documentos de CTS al trabajador [F25][N10].
   7. El cierre del periodo como paso separado: **No documentado en las fuentes disponibles**. Consultar el manual de Nóminas o el soporte de S10 [F6].

## Advertencias

- **Fechas legales:** el video indica el 15 de noviembre como fecha máxima de pago y de impresión de la boleta, y no señala el año de la fuente [F19][F22][N11]. Verificar la norma vigente antes de fijar la fecha.
- **Jornales y porcentajes:** los valores 74.30 / 58.45 / 52.50 y el 18 % de alta especialización provienen de un convenio colectivo cuyo año no consta en la fuente [F9][F10][N7]. Verificar el convenio vigente.
- **UIT y RMV:** el procedimiento de registro mostrado corresponde a 2025 [F18][N1]. Los montos cambian cada año; verificar la cifra vigente.
- **No recalcular semanas ya pagadas:** al recalcular se actualiza y sobrescribe la información ya calculada [F9][F10].
- **Sin cuenta bancaria no hay depósito:** es necesario que todo trabajador cuente con una cuenta bancaria de CTS [F22].
- **El envío a Facturación es obligatorio para pagar:** si no se envía la información, no se podrá realizar ningún pago [F13].
- **Fuentes de video:** los procedimientos citados provienen de videos y sílabos oficiales de S10, y los nombres de escenarios salen de transcripciones automáticas [N11][N7]. Confirmar en pantalla el nombre exacto de cada escenario antes de operar.
## Fuentes

- **[F3]** Silabus Nominas Basico, pág. 3 · https://www.s10peru.com/wp-content/uploads/2020/08/Silabus-Nominas-Basico.pdf
- **[F6]** 1. Manual de Cliente de Sistema de Ticket v2, pág. 1 · https://www.s10peru.com/wp-content/uploads/2018/10/1.-Manual-de-Cliente-de-Sistema-de-Ticket-v2.pdf
- **[F9]** Video: Cambio de Jornal - Construcción civil en S10 ERP, min 0:09–2:25 · https://youtu.be/TdbnLyEaN_E?t=9
- **[F10]** Video: S10 Nóminas - Incremento de Jornal (Construcción Civil), min 0:01–2:16 · https://youtu.be/6Kd187zEqOs?t=1
- **[F12]** Video: S10 Nóminas -  Nuevo Aporte Fondo de Capacitación en Construcción Civil, min 0:03–4:19 · https://youtu.be/NYUDXu-XmyA?t=3
- **[F13]** Video: Calculo Reintegro retroactivos - Construccion civil - 2021-2022, min 4:44–6:38 · https://youtu.be/umiFsebieZs?t=284
- **[F17]** Temario capacitacion de CTS, pág. 1 · https://www.s10peru.com/wp-content/uploads/2020/05/Temario-capacitacion-de-CTS.pdf
- **[F18]** Video: S10 Nóminas - Cambio de UIT y RMV (2025), min 0:00–2:11 · https://youtu.be/17W0yKI8iew?t=0
- **[F19]** Video: Cómo calcular la CTS con el S10 ERP, min 0:08–3:32 · https://youtu.be/XObV2BOrsIc?t=8
- **[F22]** Video: Cómo calcular la CTS con el S10 ERP, min 3:32–7:04 · https://youtu.be/XObV2BOrsIc?t=212
- **[F25]** Video: ¿Qué es y cómo funciona el portal del Empleado? - Grupo S10, min 0:00–1:50 · https://youtu.be/QHH4oK0m9cI?t=0
- **[N1]** nodo de Cortex `nominas/parametros-legales-uit-rmv-afp` · Parámetros legales: UIT, RMV y porcentajes AFP
- **[N3]** nodo de Cortex `nominas/ciclo-de-calculo` · Ciclo de cálculo de una planilla
- **[N7]** nodo de Cortex `nominas/construccion-civil-jornales` · Construcción civil: cambio de jornal por convenio
- **[N8]** nodo de Cortex `nominas/aporte-fondo-capacitacion` · Construcción civil: aporte al Fondo de Capacitación
- **[N10]** nodo de Cortex `portales/empleado` · Portal del Empleado y del Colaborador
- **[N11]** nodo de Cortex `nominas/cts` · Cálculo de la CTS

---
_Generado por tutor.py el 2026-09-30 02:35._