# Ingreso de materiales al almacén de obra desde orden de compra y consulta de kardex

**Para qué sirve.** Registrar en el módulo de Almacenes del S10 ERP la entrada física de materiales que llegan a obra respaldados por una orden de compra (OC), usando el movimiento *Ingreso por Orden de Compra (I OC)* [F10][F13], y después verificar el resultado en las consultas de **stock** y **kardex** del almacén [F12][F15]. Con esto se cierra el ciclo que empieza en Compras con la emisión de la OC [N13][N15] y queda trazabilidad entre la guía de remisión, el ingreso y el saldo del insumo. En el caso académico de una constructora limeña, la integración de compras, pedidos, ingresos y salidas en el ERP mejoró la trazabilidad y la disponibilidad de información confiable [F29].

## Antes de empezar

- Usuario con permisos sobre el módulo y sobre el almacén de destino. El registro de usuarios y permisos, del proyecto, de los usuarios del proyecto y de los almacenes, y de los atributos de uso corresponde a la configuración inicial de Almacenes [F10][F13][N4].
- Almacén de obra ya creado, y definidos los **documentos que usa la empresa en los movimientos** [F10][F13][N4].
- Configuración de partidas de control, destinos específicos, frentes, roles por unidad operativa y ubicación física [F10][F13][N4].
- Orden de compra ya emitida en Compras, a partir de un pedido aprobado o de una cotización [N15].
- Documentación del proveedor: guía de remisión y factura. En la práctica descrita en la tesis UPN (2025), vigilancia verifica la documentación, se inspecciona el material, el residente firma la guía y el analista logístico recién entonces registra el ingreso en el S10 [F16][N2].

## Pasos

1. **Abrir el módulo de Almacenes y ubicarse en el almacén de destino.**
   1. Ingrese al módulo de Almacenes con su usuario [F10][F13].
   2. Seleccione la empresa, el proyecto y el almacén donde entrará el material. No documentado en las fuentes disponibles el nombre exacto de la ventana ni la secuencia de selección; consulte el Manual de Almacenes del portal de documentación, que exige cuenta de acceso [N7].

2. **Verificar la orden de compra que respalda el ingreso.**
   1. Ubique la OC emitida para el proveedor y revise insumos, cantidades y precios pactados [N15].
   2. Compare la OC con la guía de remisión y la factura recibidas [F16][N2].
   3. No documentado en las fuentes disponibles la pantalla de consulta de OC dentro de Almacenes; consulte el manual restringido de Compras o el Portal de Proveedores [N13].

3. **Crear el movimiento de ingreso con el tipo correcto.**
   1. Adicione un nuevo movimiento de ingreso. En las guías públicas del portal, la creación de registros se hace con **clic derecho → Adicionar** [F2][N1].
   2. Elija el tipo **Ingreso por Orden de Compra (I OC)**, que es el movimiento previsto para cerrar una OC. No use *Ingreso por Compra (I C)*, que es un tipo distinto [F10][F13][N14].
   3. Los sílabos no explican la diferencia entre I C e I OC; defina el criterio con soporte S10 antes de estandarizarlo en la obra [N14].
   4. Registre el documento de respaldo según los documentos configurados para movimientos [F10][F13][N4].

4. **Cargar los insumos y ajustar las cantidades recibidas.**
   1. Traiga al detalle del ingreso los insumos de la OC seleccionada.
   2. Ajuste cada cantidad a lo realmente recibido según la guía de remisión, no a lo pedido en la OC [F16][N2].
   3. Si hay faltantes, daños o discrepancias, no fuerce el ingreso: en el flujo documentado, el caso se observa, se rechaza o se reprograma [F16][N2].
   4. No documentado en las fuentes disponibles el nombre del comando para importar líneas de la OC ni el manejo de ingresos parciales; verifíquelo en el sistema con un caso de prueba y consulte el manual de Almacenes [N7][N14].

5. **Completar la valorización del ingreso.**
   1. Registre precio unitario, moneda, tipo de cambio, flete e impuestos. **No documentado en las fuentes disponibles** qué campos de valorización expone la nota de ingreso ni cómo se prorratea el flete; consulte el Manual de Almacenes restringido [N7].
   2. Los catálogos de **moneda, formas de pago y tipo de cambio** y el catálogo de **tipos de impuesto** se definen en el módulo Administrativo Contable, igual que la asignación de detracciones o percepciones a los recursos [F17].
   3. Si el proveedor facturó un anticipo sobre la OC, el sistema enlaza el ingreso con la primera factura (la del anticipo). Use **clic derecho → quitar regularización** y relacione el ingreso con la segunda factura, la de la entrega [F2][N1][N11].

6. **Grabar y aprobar el ingreso.**
   1. Grabe el movimiento.
   2. No documentado en las fuentes disponibles si el ingreso de almacén requiere un paso de aprobación propio ni quién lo autoriza. En el caso UPN, la autorización es previa y presencial: el residente firma la guía y el analista registra [F16][N2]. El circuito *enviar a aprobación → aprobar* está documentado para órdenes de pago del módulo Administrativo, no para ingresos de almacén [F14][N12].
   3. Verifique con soporte S10 el efecto exacto del grabado sobre el stock. Puede emitir un ticket desde los links de soporte del portal [F18].

7. **Consultar stock y kardex del insumo.**
   1. Abra las consultas de **stock** y **kardex** del módulo, disponibles junto con **recursos con problemas**, ingresos y egresos generales, stock general y equipos disponibles [F12][F15][N19].
   2. Confirme que el movimiento de ingreso aparece con su cantidad y documento, y revise el saldo resultante.
   3. Revise **recursos con problemas** para detectar insumos con datos inconsistentes. Los sílabos no definen qué considera el sistema un "recurso con problemas" [N19].
   4. El **método de valorización del kardex** (promedio, FIFO u otro) no está documentado en las fuentes disponibles; confírmelo con soporte S10 antes de usar el costo unitario para decisiones [N19].
   5. Genere los **reportes e informes de ingresos** para el control periódico [F10][F13].

## Advertencias

- Registrar el ingreso sin conformidad física es la causa directa de descuadres. En el caso UPN (2025) se midió un **error de stock de 12.31%** entre el kardex previo en Excel (S/ 24,476) y el conteo físico (S/ 27,489.80); esa cifra mide el control anterior al ERP, no al S10 [F26][F27][N2]. Cifras del año 2025 y de una sola empresa: verifique los valores de su propia obra.
- En la misma obra se reportó una **merma de 17%** en enchapes por transporte y mala manipulación, unos **S/ 4,532**, atribuida a almacenes temporales inadecuados [F27][F29][N2]. Datos de 2025, referenciales.
- Los autores recomiendan **inventarios cíclicos quincenales o mensuales** para reducir las diferencias entre el ERP y el conteo físico [F29][N2]. Es una recomendación académica, no un parámetro del sistema.
- En el caso de anticipos, dejar el ingreso ligado a la factura del adelanto regulariza el material contra el documento equivocado [N1].
- Los nombres de menús, ventanas y atajos del módulo de Almacenes no constan en fuentes públicas: el manual del portal exige cuenta [N7]. No asuma rutas; valídelas en su instalación.
- Este tutorial no cubre la corrección de un ingreso mal grabado. Para devolver material al proveedor existen los movimientos *Egreso por Devolución al Proveedor* y *Egreso por Devolución de Orden de Compra (E DOC)*, y para corregir cantidades *Egreso por Ajuste de Cantidad (E AC)* [F12][F15][N19]; su procedimiento no está documentado en las fuentes disponibles.
## Fuentes

- **[F2]** Manual de factura de anticipos de Almacén con compuesta · https://documentacion.s10peru.com/caso-anticipo/
- **[F10]** Silabus Almacenes, pág. 1 · https://www.s10peru.com/wp-content/uploads/2020/08/Silabus-Almacenes.pdf
- **[F12]** Silabus Almacenes, pág. 3 · https://www.s10peru.com/wp-content/uploads/2020/08/Silabus-Almacenes.pdf
- **[F13]** Syllabus Almacenes S10ERP, pág. 1 · https://www.s10peru.com/wp-content/uploads/2018/08/Syllabus_Almacenes_S10ERP.pdf
- **[F14]** Cómo realizar el pago de una factura afecta a detracción y subir las constancias al sistema 10 · https://documentacion.s10peru.com/realizar-pago-factura-afecta-detraccion-subir-constancias/
- **[F15]** Syllabus Almacenes S10ERP, pág. 1 · https://www.s10peru.com/wp-content/uploads/2018/08/Syllabus_Almacenes_S10ERP.pdf
- **[F16]** Análisis de la implementación del sistema ERP S10 en la logística de una constructora (UPN), pág. 18 · https://repositorio.upn.edu.pe/backend/api/core/bitstreams/9c558024-0307-41d8-b040-63696966fd36/content
- **[F17]** Syllabus Administrativo S10ERP, pág. 1 · https://www.s10peru.com/wp-content/uploads/2018/08/Syllabus_Administrativo_S10ERP.pdf
- **[F18]** 1. Manual de Cliente de Sistema de Ticket v2, pág. 3 · https://www.s10peru.com/wp-content/uploads/2018/10/1.-Manual-de-Cliente-de-Sistema-de-Ticket-v2.pdf
- **[F26]** Análisis de la implementación del sistema ERP S10 en la logística de una constructora (UPN), pág. 33 · https://repositorio.upn.edu.pe/backend/api/core/bitstreams/9c558024-0307-41d8-b040-63696966fd36/content
- **[F27]** Análisis de la implementación del sistema ERP S10 en la logística de una constructora (UPN), pág. 32 · https://repositorio.upn.edu.pe/backend/api/core/bitstreams/9c558024-0307-41d8-b040-63696966fd36/content
- **[F29]** Análisis de la implementación del sistema ERP S10 en la logística de una constructora (UPN), pág. 35 · https://repositorio.upn.edu.pe/backend/api/core/bitstreams/9c558024-0307-41d8-b040-63696966fd36/content
- **[N1]** nodo de Cortex `almacenes/anticipos-y-regularizacion-compuesta` · Anticipos a proveedor y regularización compuesta
- **[N2]** nodo de Cortex `almacenes/caso-upn-ingresos-salidas-e-inventario` · Caso UPN: ingresos, salidas e inventario con S10
- **[N4]** nodo de Cortex `almacenes/configuracion-de-almacenes` · Configuración de almacenes
- **[N7]** nodo de Cortex `almacenes` · Almacenes (S10 ERP)
- **[N11]** nodo de Cortex `facturacion/anticipo-de-proveedor` · Anticipo de proveedor facturado y amortizado
- **[N12]** nodo de Cortex `administrativo/ordenes-de-pago` · Órdenes de pago y aprobación
- **[N13]** nodo de Cortex `compras` · Compras y Pedidos (S10 ERP)
- **[N14]** nodo de Cortex `almacenes/movimientos-de-ingreso` · Movimientos de ingreso al almacén
- **[N15]** nodo de Cortex `compras/pedidos-cotizaciones-y-ordenes` · Pedidos, cotizaciones y órdenes de compra
- **[N19]** nodo de Cortex `almacenes/egresos-y-transferencias` · Egresos, transferencias y consultas de stock y kardex

---
_Generado por tutor.py el 2026-09-30 02:37._