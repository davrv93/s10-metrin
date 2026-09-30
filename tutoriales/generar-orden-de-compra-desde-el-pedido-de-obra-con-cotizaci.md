# Generar orden de compra desde el pedido de obra con cotizaciones y aprobación móvil

**Para qué sirve.** Cubre el circuito de abastecimiento de obra en S10 ERP: desde el pedido que nace en obra hasta la orden de compra (OC) aprobada y lista para el ingreso al almacén. El módulo se llama "Compras y Pedidos" y su curso oficial de segundo día recorre justamente esta secuencia: generar pedidos de compra y su aprobación, crear cotizaciones, asignar proveedores, recibir cotizaciones, analizar el cuadro comparativo y generar la OC desde la cotización [F2][F3][F14]. El tercer día de la versión de 12 horas añade las apps móviles de Pedidos y de Aprobaciones [F14]. En la práctica de obra, el flujo arranca con el requerimiento del residente, pasa por logística, que cotiza y compara contra el presupuesto, y termina con la OC y el seguimiento hasta la entrega [F1][F16].

## Antes de empezar

- Módulo de Compras configurado: centro de compras con proyectos asignados, atributos de uso y **grupos de aprobación** [F2][F3][F14]. Sin grupos de aprobación no hay quién autorice el pedido ni la OC [N1].
- Pie de las órdenes de compra y de servicio configurado por empresa, más logotipos y firmas digitales [F14].
- Catálogos cargados: recursos, socios de negocio (proveedores) y formas de pago [F2][F3][F14].
- Usuarios y permisos registrados [F2][F3].
- App de Aprobaciones instalada en el teléfono del aprobador, si se usará la vía móvil [F14][F30].
- El detalle de campos, montos y niveles de aprobación está en el Manual de Compras del portal de ayuda, que exige cuenta de cliente [N1][F30].

## Pasos

### 1. Registrar el pedido de obra con los insumos del presupuesto

1.1. Abrir el módulo de Compras y generar un **pedido de compra** [F2][F3][F14].
1.2. Consignar los insumos requeridos. En el flujo de la constructora analizada, el residente envía un proyectado quincenal de requerimientos [F1].
1.3. Asignar el proyecto al que se imputa el pedido, según el centro de compras configurado [F2][F3].
1.4. Contrastar las cantidades y los precios contra el presupuesto de la obra; ese cotejo es el que detecta desviaciones [F1].
1.5. Los campos exactos de la pantalla de pedido: **No documentado en las fuentes disponibles**. Consultar el Manual de Compras del portal de ayuda [F30].
1.6. Alternativa móvil: la APP de Pedidos y Despacho permite generar distintos tipos de pedido y hacerles seguimiento [F14]. Los tipos admitidos y las notificaciones: **No documentado en las fuentes disponibles** [N10].

### 2. Revisar y aprobar el pedido

2.1. Enviar el pedido a aprobación. El sílabo oficial trata "pedidos de compras y su aprobación" como paso obligado antes de la OC [F2][F3][F14].
2.2. El aprobador del grupo configurado revisa el pedido y lo autoriza [F14].
2.3. Los niveles, montos y monedas de aprobación del módulo de Compras: **No documentado en las fuentes disponibles**. Consultar el Manual de Compras [F30]. Como referencia de otro módulo, en Administrativo los aprobadores se definen con moneda y monto límite [F19][F34].
2.4. En el caso documentado de la constructora, las órdenes de compra las aprobaba gerencia [F6]. Eso describe a una empresa concreta, no una regla del sistema [N3].

### 3. Generar la solicitud de cotización y asignar proveedores

3.1. Crear una **cotización** para compra de materiales o servicios [F2][F3][F14].
3.2. **Asignar los proveedores** a la cotización, tomándolos del catálogo de socios de negocio [F2][F3][F14].
3.3. Agrupar los requerimientos por proveedor y por cantidad requerida antes de solicitar precios, como hace logística en el flujo de obra [F4].
3.4. Si la cotización se genera directamente desde el pedido aprobado o se crea aparte: **No documentado en las fuentes disponibles**. Los sílabos enumeran ambos caminos sin describir el enlace [F2][F14]; consultar el Manual de Compras [F30].

### 4. Registrar las cotizaciones recibidas

4.1. Usar la **recepción de cotizaciones**, que en la versión de 12 horas se hace **por proveedor** [F14].
4.2. Registrar el precio ofertado por cada insumo.
4.3. Registrar la forma de pago del proveedor, tomada del catálogo de formas de pago [F2][F3][F14].
4.4. Plazos de entrega y condiciones comerciales: **No documentado en las fuentes disponibles**. Consultar el Manual de Compras [F30].
4.5. Si los precios superan lo esperado, renegociar con el proveedor antes de continuar [F1][F16].

### 5. Evaluar el cuadro comparativo y elegir proveedor

5.1. Abrir el **cuadro comparativo de cotizaciones**; la versión de 12 horas lo llama "análisis del cuadro comparativo" [F2][F3][F14].
5.2. Comparar las ofertas contra el presupuesto de obra, no solo entre sí [F1].
5.3. Seleccionar al proveedor por insumo. Las reglas de adjudicación del comparativo: **No documentado en las fuentes disponibles**. Consultar el Manual de Compras [F30].
5.4. Si hay desviaciones, renegociar y volver al comparativo [F1].

### 6. Generar la orden de compra desde la cotización

6.1. Ejecutar **generar orden de compra desde la cotización** [F2][F3][F14]. Existe también el camino de generar la OC desde el pedido de compra, sin cotización [F2][F3][F14].
6.2. Verificar el proyecto y el centro de compras asociados a la OC [F2][F3].
6.3. Verificar la moneda. El catálogo de monedas y el tipo de cambio se revisan en la configuración del módulo Administrativo Contable [F27][F28].
6.4. Verificar el tratamiento tributario. Los tipos de impuesto, los ítems de detracción y de percepción se definen por catálogo y se asignan a los recursos [F27][F28]. La tasa vigente del IGV no figura en estas fuentes: **No documentado en las fuentes disponibles**; verificar la tasa vigente con SUNAT antes de emitir.
6.5. Cifra legal de referencia: la única tasa impositiva que aparece en el corpus es un IGV de 19 % en un ejemplo de diseño de pie de presupuesto, correspondiente a la época de esa guía [N19]. **No usar ese valor; verificar la tasa de IGV vigente.**
6.6. Verificar el pie de la OC, el logotipo y la firma, configurados previamente por empresa [F14].

### 7. Aprobar la orden de compra desde la app móvil

7.1. Abrir la **APP de Aprobaciones** de S10 en el teléfono [F14][F30][F31].
7.2. Ubicar la orden de compra pendiente y aprobarla con el usuario del nivel correspondiente.
7.3. La secuencia exacta de pantallas, la firma y las notificaciones de la app: **No documentado en las fuentes disponibles**. El Portal de Ayuda S10 lista "Aprobaciones Móviles" con 9 temas, pero el manual exige cuenta de cliente [F30][F33][N20].
7.4. Si la app falla, generar un ticket de soporte con capturas adjuntas; el sistema devuelve un número de ticket para el seguimiento [F5][F17][F20].

### 8. Emitir la OC al proveedor y habilitar el ingreso al almacén

8.1. Imprimir o emitir la OC aprobada con su pie, logotipo y firma [F14].
8.2. Enviarla al proveedor y coordinar el pago según la forma pactada, al crédito o al contado [F16].
8.3. Programar el despacho y hacer el seguimiento logístico hasta la entrega en obra [F16].
8.4. En almacén, recibir el material con el movimiento **Ingreso por Orden de Compra (I OC)** [F26][F29][F38].
8.5. Verificar la documentación, la cantidad y el estado físico antes de registrar el ingreso [F4].
8.6. Si el proveedor factura un anticipo sobre la OC, relacionar esa factura con la OC eligiendo "ORDEN DE COMPRA O SERVICIO" en el campo **Relacionado con**; el sistema la trata como adelanto [F25].

## Advertencias

- Los sílabos oficiales solo enumeran temas. No describen campos, botones ni estados del pedido o de la OC. Todo lo marcado como no documentado debe verificarse en el Manual de Compras del Portal de Ayuda S10, que requiere cuenta de cliente [F30][F33][N1].
- El único porcentaje de IGV presente en las fuentes es 19 %, tomado de un ejemplo de pie de presupuesto de la guía de Presupuestos [N19]. Es un valor de época, no una referencia vigente. Verificar la tasa actual con SUNAT antes de emitir cualquier OC.
- Las cifras del caso de la constructora (16 % de ahorro promedio, S/ 23,445.82 de ahorro, mejora neta de 27 días) provienen de un trabajo académico de la UPN de 2025 sobre una empresa concreta. No son parámetros del sistema [N3].
- Aprobar en la app y aprobar en el escritorio dependen de la misma configuración de grupos de aprobación [F14]. Si nadie está asignado, el pedido o la OC quedan detenidos.
- En el registro de un anticipo, el sistema enlaza el ingreso de almacén con la primera factura por defecto. Hay que quitar esa regularización para vincularlo a la factura final [F25].
- No se documentan aquí atajos de teclado ni nombres de ventanas del módulo de Compras, porque las fuentes disponibles no los consignan.
## Fuentes

- **[F1]** Análisis de la implementación del sistema ERP S10 en la logística de una constructora (UPN), pág. 15 · https://repositorio.upn.edu.pe/backend/api/core/bitstreams/9c558024-0307-41d8-b040-63696966fd36/content
- **[F2]** Syllabus Compras S10ERP, pág. 1 · https://www.s10peru.com/wp-content/uploads/2018/08/Syllabus_Compras_S10ERP.pdf
- **[F3]** Silabus Compras y Pedidos, pág. 1 · https://www.s10peru.com/wp-content/uploads/2020/08/Silabus-Compras-y-Pedidos.pdf
- **[F4]** Análisis de la implementación del sistema ERP S10 en la logística de una constructora (UPN), pág. 17 · https://repositorio.upn.edu.pe/backend/api/core/bitstreams/9c558024-0307-41d8-b040-63696966fd36/content
- **[F5]** Video: Cómo generar un Ticket de Soporte S10, min 2:34–2:39 · https://youtu.be/rrhDrBcQ5pk?t=154
- **[F6]** Análisis de la implementación del sistema ERP S10 en la logística de una constructora (UPN), pág. 27 · https://repositorio.upn.edu.pe/backend/api/core/bitstreams/9c558024-0307-41d8-b040-63696966fd36/content
- **[F14]** Silabus Compras y Pedidos 12 horas, pág. 1 · https://www.s10peru.com/wp-content/uploads/2023/03/Silabus-Compras-y-Pedidos-12-horas.pdf
- **[F16]** Análisis de la implementación del sistema ERP S10 en la logística de una constructora (UPN), pág. 16 · https://repositorio.upn.edu.pe/backend/api/core/bitstreams/9c558024-0307-41d8-b040-63696966fd36/content
- **[F17]** 1. Manual de Cliente de Sistema de Ticket v2, pág. 8 · https://www.s10peru.com/wp-content/uploads/2018/10/1.-Manual-de-Cliente-de-Sistema-de-Ticket-v2.pdf
- **[F19]** Syllabus Administrativo S10ERP, pág. 1 · https://www.s10peru.com/wp-content/uploads/2018/08/Syllabus_Administrativo_S10ERP.pdf
- **[F20]** 1. Manual de Cliente de Sistema de Ticket v2, pág. 5 · https://www.s10peru.com/wp-content/uploads/2018/10/1.-Manual-de-Cliente-de-Sistema-de-Ticket-v2.pdf
- **[F25]** Manual de factura de anticipos de Almacén con compuesta · https://documentacion.s10peru.com/caso-anticipo/
- **[F26]** Silabus Almacenes, pág. 1 · https://www.s10peru.com/wp-content/uploads/2020/08/Silabus-Almacenes.pdf
- **[F27]** Syllabus Administrativo S10ERP, pág. 1 · https://www.s10peru.com/wp-content/uploads/2018/08/Syllabus_Administrativo_S10ERP.pdf
- **[F28]** Silabus Administrativo Contable, pág. 2 · https://www.s10peru.com/wp-content/uploads/2023/05/Silabus-Administrativo-Contable.pdf
- **[F29]** Syllabus Almacenes S10ERP, pág. 1 · https://www.s10peru.com/wp-content/uploads/2018/08/Syllabus_Almacenes_S10ERP.pdf
- **[F30]** Home Portal de Ayuda S10 ex - Portal de Ayuda S10 · https://documentacion.s10peru.com/inicio/
- **[F31]** Sidebar General · https://documentacion.s10peru.com/sidebar-general-dic-19/
- **[F33]** Portal de Ayuda S10 - Portal de Ayuda S10 · https://documentacion.s10peru.com/
- **[F34]** Silabus Administrativo Contable, pág. 3 · https://www.s10peru.com/wp-content/uploads/2023/05/Silabus-Administrativo-Contable.pdf
- **[F38]** Silabus Almacenes, pág. 3 · https://www.s10peru.com/wp-content/uploads/2020/08/Silabus-Almacenes.pdf
- **[N1]** nodo de Cortex `compras/configuracion-y-catalogos` · Configuración de compras y catálogos
- **[N3]** nodo de Cortex `compras/caso-upn-flujo-y-resultados` · Caso UPN: flujo de compras y resultados tras implantar S10
- **[N10]** nodo de Cortex `compras/pedidos-y-aprobaciones-moviles` · Apps de Pedidos y Aprobaciones Móviles
- **[N19]** nodo de Cortex `presupuestos/pie-gastos-generales-y-formula-polinomica` · Pie de presupuesto, gastos generales y fórmula polinómica
- **[N20]** nodo de Cortex `portales/aprobaciones-pedidos-moviles` · Aprobaciones y Pedidos Móviles

---
_Generado por tutor.py el 2026-09-30 02:40._