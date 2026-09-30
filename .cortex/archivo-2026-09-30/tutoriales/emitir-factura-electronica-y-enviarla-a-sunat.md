---
title: Emitir factura electrónica y enviarla a SUNAT
summary: "Tutorial de Facturación Electrónica / Comercial: Emitir factura
  electrónica y enviarla a SUNAT. 8 pasos, 32 citas, 9 pasos sin documentar."
tags:
  - tutorial
  - facturacion-electronica-comerc
id: 01M3RM85D9P1NC9AMG5YSY8HTS
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:42:46.186Z
---

# Emitir factura electrónica y enviarla a SUNAT

**Para qué sirve.** Este tutorial cubre el registro de una factura de venta en el módulo **Facturación** del S10 ERP y su envío a SUNAT como comprobante electrónico [F1][N17]. El módulo Facturación administra los documentos por cobrar y por pagar y controla las fechas de vencimiento [F1]. Los documentos por cobrar se pueden crear sin relación, relacionados con un documento de venta, o desde el cronograma de facturación de una orden de venta [F17][F19][N17]. Atención: el **Portal de Facturación Electrónica** es un portal web distinto del módulo Facturación del ERP [N1]. Sus 25 documentos y videos están detrás del login de miembros, así que los pasos del portal no constan en las fuentes disponibles [N1][N6].

## Antes de empezar

- Tenga configurados los catálogos tributarios: moneda, forma de pago, tipo de cambio, tipos de impuesto, tipos de documento, ítems de detracción e ítems de percepción [F17][F18][N14].
- Cargue y actualice los padrones de agentes de retención, agentes de percepción y buenos contribuyentes [F17][F18][N14].
- Asigne la detracción o la percepción a los recursos. Así el documento reconoce si está afecto [F17][N14].
- Cree las empresas y sucursales, y los usuarios con acceso por sucursal [F17][F18][N14].
- Defina la configuración general de registro de documentos por empresa y las fechas límite de los periodos contables [F17][F18][N14].
- Configure los **talonarios** de los tipos de documento por cobrar y sus accesos, además de la impresión de facturas, boletas y notas de crédito y débito [F17][F18][N14].
- Registre al cliente en el **Catálogo de Clientes** con su RUC [F9][F10][F12].
- Revise el parámetro **«Plazo máximo de envío a SUNAT desde la fecha de emisión»** [F3][N2].

## Pasos

### 1. Verificar la configuración del comprobante electrónico

1. Revise que el talonario del tipo de documento *Factura* esté configurado y que su usuario tenga acceso a él [F17][F18][N14].
2. Confirme el plazo de envío: entre al módulo **Facturación**, abra el escenario **Documentos por cobrar**, pulse el botón de **configuración** del escenario y elija la pestaña **«Del Supervisor (4)»** [F3][N2].
3. Ubique **«Plazo máximo de envío a SUNAT desde la fecha de emisión»** y ponga los días que indique la SUNAT [F3][N2]. Las fuentes no dan ninguna cifra de días; verifique el plazo vigente en SUNAT [N2].
4. Verifique los datos del emisor y el **certificado digital** de firma electrónica. No documentado en las fuentes disponibles: la única activación de certificado que consta es la licencia del software, no el certificado tributario [F8][N3]. Consulte el «Manual de Facturación Electrónica» del portal de ayuda, que es de acceso restringido [N1][N17].

### 2. Validar al cliente en el maestro

1. Abra el **Catálogo de Clientes** [F9][F21].
2. Si el cliente no existe, haga clic derecho sobre el listado y elija **Adicionar** [F9][F12].
3. En la ventana **Identificador**, indique si es **Persona Jurídica**, y llene **Razón Social**, **Abreviatura** y **RUC** [F9][F12].
4. En la pestaña **Tipo Identificador**, elija con doble clic el código **04 – Cliente** en el *Catálogo de Tipo de Identificadores* [F10][F11][F12].
5. Complete la pestaña **Dirección** con la dirección fiscal. Las fuentes muestran la pestaña, pero no detallan sus campos ni qué dirección exige SUNAT [F9][F10]. No documentado en las fuentes disponibles: consulte el manual del módulo.
6. Pulse **Adicionar** para grabar y elija el cliente con doble clic para llevarlo a la ventana que lo solicitó [F10][F11].

### 3. Crear el documento por cobrar

1. En el módulo **Facturación**, abra el escenario **Documentos por cobrar** [F3][N2].
2. Haga clic derecho sobre el listado y elija **Adicionar** [F23][N19].
3. Llene los datos de cabecera del documento [F23][N19].
4. En el campo **Relacionado con**, elija el origen que corresponda. Las fuentes documentan esta lista solo para documentos por pagar, donde una opción es «ORDEN DE COMPRA O SERVICIO» [F23][N19]. Para los documentos por cobrar, los sílabos indican que existen las variantes *sin documento relacionado*, *relacionado con documento de venta* e *internos (anticipos)*, pero no dan los nombres exactos de las opciones en pantalla [F17][F18][N14][N17]. No documentado en las fuentes disponibles.
5. Elija el tipo de comprobante **Factura**, según el talonario configurado [F17][F18][N14].
6. Indique **moneda**, **tipo de cambio** y **forma de pago**, tomados de los catálogos varios [F17][F18][N14].
7. Si la factura nace del cronograma de facturación de una orden de venta, genere el documento por cobrar desde ese cronograma [F19][N17].

### 4. Ingresar el detalle

1. Registre cada recurso o servicio como una línea del detalle [F23][N19].
2. Ingrese la **cantidad** y el **precio unitario sin IGV** de cada línea. La única referencia explícita es el caso del anticipo, donde se usa cantidad 1 y el P.U. sin IGV igual al monto facturado [F23][N19].
3. Asigne el **centro de costo** o la partida de imputación de cada línea. No documentado en las fuentes disponibles: el nombre del campo en la factura no consta. Consulte el «Manual de Facturación» del portal de ayuda, que es restringido [N17].
4. Revise el total del documento antes de grabar. No documentado en las fuentes disponibles.

### 5. Aplicar impuestos, detracción, retención o percepción

1. El sistema reconoce si el documento está afecto a **detracción** o **percepción** por la asignación previa de esos ítems a los recursos [F17][F18][N14].
2. Confirme que el **IGV** se calcule según el catálogo de tipos de impuesto [F17][F18][N14]. Las fuentes no dan la tasa de IGV vigente; el único porcentaje que aparece es «IGV 19%» en un ejemplo del pie de presupuesto de otro módulo [F26]. Ese dato es de un manual sin año indicado y no sirve como tasa actual: verifique la tasa vigente en SUNAT.
3. Si la empresa es agente de retención, o el cliente figura en el padrón de percepción o de buenos contribuyentes, el cálculo depende de los padrones cargados [F17][F18][N14]. El detalle del cálculo no está documentado en las fuentes disponibles.
4. En una factura afecta a detracción, valide el campo **CRONOGRAMA DE PAGOS** y verifique que la detracción figure **independiente** [F2][F28][N14][N21]. Esa separación es lo que permite pagar la detracción por sí sola desde el módulo Administrativo [F2][N21].

### 6. Grabar y aprobar el documento

1. Grabe el documento por cobrar. La numeración correlativa la entrega el **talonario** configurado para el tipo de documento [F17][F18][N14]. El nombre del botón de grabado no consta: no documentado en las fuentes disponibles.
2. Envíe el documento al flujo de aprobación si su empresa lo exige. Las fuentes documentan **clic derecho > Enviar a aprobación** y luego **Aprobar** solo para las órdenes de pago del módulo Administrativo, no para las facturas de venta [F2][N15][N21]. No documentado en las fuentes disponibles para este caso.
3. Imprima la factura con el formato configurado para facturas, boletas y notas de crédito y débito [F17][F18][N14].

### 7. Generar el XML firmado y enviarlo a SUNAT

1. Emita el comprobante dentro del plazo cargado en **«Plazo máximo de envío a SUNAT desde la fecha de emisión»** [F3][N2]. Si el plazo ya venció, el envío se rechaza; el parámetro es el que controla esa validación [F3][N2].
2. Genere el XML firmado y envíelo a SUNAT. No documentado en las fuentes disponibles: los pasos del **Portal de Facturación Electrónica** están detrás del login de miembros [N1][N6]. Consulte el «Manual de Facturación Electrónica» o la sección «Portal de Facturación Electrónica» (25 documentos) del portal de ayuda [F25][N1].
3. Si necesita más días para enviar, amplíe el plazo con el procedimiento del paso 1 de la sección 1 [F3][N2]. La pestaña se llama «Del Supervisor», de modo que probablemente requiera un usuario con ese nivel de permisos; la fuente no lo dice expresamente [N2].

### 8. Consultar la CDR y reprocesar un rechazo

1. Consulte el estado del comprobante y la **constancia de recepción (CDR)** en el Portal de Facturación Electrónica. No documentado en las fuentes disponibles [N1][N6].
2. Descargue el PDF o el XML aceptado. No documentado en las fuentes disponibles: la única descarga que consta en las fuentes es la de documentos laborales del Portal del Empleado, que es otro portal [F6][N4].
3. Si SUNAT rechaza el comprobante, corrija la causa y reenvíelo. Las fuentes no describen el reproceso: no documentado en las fuentes disponibles.
4. Si el problema persiste, genere un **ticket de soporte**: entre a www.s10peru.com/soporte, pulse **Emitir** en «Ticket de Soporte», ingrese con su usuario y clave, elija la categoría, escriba el asunto y la descripción, fije la prioridad, adjunte hasta cuatro archivos y pulse **crear ticket** [F13][F15][F16].

## Advertencias

- El **Portal de Facturación Electrónica** y el módulo **Facturación** del ERP son cosas distintas. La ampliación del plazo de envío pertenece al módulo, no al portal [N1][N2]. Aclare siempre de cuál de los dos habla antes de seguir un procedimiento [N1].
- Los pasos 7 y 8 quedan incompletos por una razón concreta: los 25 documentos y los videos del Portal de Facturación Electrónica exigen cuenta de miembro y no se pudieron leer [N1][N6].
- Toda cifra legal debe verificarse. Las fuentes no traen la tasa de IGV vigente, ni el plazo de días de envío a SUNAT, ni los umbrales de detracción, retención o percepción [F3][N2][N14]. El «IGV 19%» de [F26] es un ejemplo de manual, sin año declarado, y no es la tasa actual. Confirme siempre lo vigente en SUNAT.
- Si la detracción no queda **independiente** en el cronograma de pagos, no se podrá pagar por separado después [F2][N21].
- El número de constancia de detracción es «requisito indispensable para el correcto envío del libro electrónico de compras» [N21]. Aplica a las facturas de proveedor, no a las de venta, pero conviene tenerlo presente al cerrar el periodo [N21][N22].
- No se asuma que el flujo **Enviar a aprobación / Aprobar** de las órdenes de pago aplica igual a las facturas de venta. Las fuentes solo lo documentan para el módulo Administrativo [F2][N15][N21].
- Los sílabos de capacitación enumeran temas, no menús. Los nombres de campos y ventanas que no aparecen citados aquí no están confirmados por ninguna fuente [N14][N5].
## Fuentes

- **[F1]** Facturación · https://documentacion.s10peru.com/facturacion/
- **[F2]** Cómo realizar el pago de una factura afecta a detracción y subir las constancias al sistema 10 · https://documentacion.s10peru.com/realizar-pago-factura-afecta-detraccion-subir-constancias/
- **[F3]** Ampliación de Plazo Máximo de envío de Documento Electrónico · https://documentacion.s10peru.com/ampliacion-plazo-maximo-envio-documento-electronico/
- **[F6]** Video: ¿Qué es y cómo funciona el portal del Empleado? - Grupo S10, min 0:00–1:50 · https://youtu.be/QHH4oK0m9cI?t=0
- **[F8]** Video: Aprende a cómo instalar S10 Presupuestos, min 5:53–8:23 · https://youtu.be/hvqEKVrCXpA?t=353
- **[F9]** Guia de Usuario de S10 Presupuestos, pág. 13 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F10]** Guia de Usuario de S10 Presupuestos, pág. 14 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F11]** Manual de S10 Costos y Presupuestos, pág. 18 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F12]** Manual de S10 Costos y Presupuestos, pág. 17 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F13]** 1. Manual de Cliente de Sistema de Ticket v2, pág. 3 · https://www.s10peru.com/wp-content/uploads/2018/10/1.-Manual-de-Cliente-de-Sistema-de-Ticket-v2.pdf
- **[F15]** 1. Manual de Cliente de Sistema de Ticket v2, pág. 6 · https://www.s10peru.com/wp-content/uploads/2018/10/1.-Manual-de-Cliente-de-Sistema-de-Ticket-v2.pdf
- **[F16]** Video: Cómo generar un Ticket de Soporte S10, min 0:11–2:34 · https://youtu.be/rrhDrBcQ5pk?t=11
- **[F17]** Syllabus Administrativo S10ERP, pág. 1 · https://www.s10peru.com/wp-content/uploads/2018/08/Syllabus_Administrativo_S10ERP.pdf
- **[F18]** Silabus Administrativo Contable, pág. 2 · https://www.s10peru.com/wp-content/uploads/2023/05/Silabus-Administrativo-Contable.pdf
- **[F19]** Silabus Administrativo Contable, pág. 3 · https://www.s10peru.com/wp-content/uploads/2023/05/Silabus-Administrativo-Contable.pdf
- **[F21]** Manual de S10 Costos y Presupuestos, pág. 16 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F23]** Manual de factura de anticipos de Almacén con compuesta · https://documentacion.s10peru.com/caso-anticipo/
- **[F25]** Home Portal de Ayuda S10 ex - Portal de Ayuda S10 · https://documentacion.s10peru.com/inicio/
- **[F26]** Manual de S10 Costos y Presupuestos, pág. 47 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F28]** Cómo realizar el pago de una factura afecta a detracción y subir las constancias al sistema 10 · https://documentacion.s10peru.com/realizar-pago-factura-afecta-detraccion-subir-constancias/
- **[N1]** nodo de Cortex `portales/facturacion-electronica` · Portal de Facturación Electrónica
- **[N2]** nodo de Cortex `facturacion/plazo-envio-documento-electronico` · Ampliar el plazo de envío a SUNAT
- **[N3]** nodo de Cortex `presupuestos/instalacion-y-licencia` · Instalación y activación de S10 Presupuestos
- **[N4]** nodo de Cortex `portales/empleado` · Portal del Empleado y del Colaborador
- **[N5]** nodo de Cortex `compras/configuracion-y-catalogos` · Configuración de compras y catálogos
- **[N6]** nodo de Cortex `portales` · Portales web y apps de gestión
- **[N14]** nodo de Cortex `facturacion/documentos-y-catalogos-tributarios` · Documentos y catálogos tributarios
- **[N15]** nodo de Cortex `administrativo/ordenes-de-pago` · Órdenes de pago y aprobación
- **[N17]** nodo de Cortex `facturacion` · Facturación en S10 ERP
- **[N19]** nodo de Cortex `facturacion/anticipo-de-proveedor` · Anticipo de proveedor facturado y amortizado
- **[N21]** nodo de Cortex `administrativo/pago-de-detracciones` · Pago de detracciones y carga de constancias
- **[N22]** nodo de Cortex `contabilidad/libros-y-estados-financieros` · Libros, registros y estados financieros

---
_Generado por tutor.py el 2026-09-30 02:42._
