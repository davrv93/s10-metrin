# Cambio de Titular con Minuta

Fuente: https://documentacion.s10peru.com/cambio-de-titular-con-minuta/

Caso – Abonos Boletados con envió de Minuta
Para poder llevar a cabo este proceso debemos cumplir con el Boletaje de todos los abonos enviados desde el Sperant a S10
SPERANT
Previamente debe estar registrado la Orden de Venta y por ende debe ser Sincronizada en el S10 mediante la siguiente opción
En la parte superior debe ubicarse en la opcion Cronograma/Pagos realizando asi el registro del cronograma respectivo del cliente
En la parte superior debe ubicar la opción Nuevo Pago, la cual desencadena una ficha
A continuación el cronograma debe ser enviado aprobación mediante la siguiente opción.
La aprobación del cronograma conlleva a que se habilite la opción de Realizar abono mediante el icono del $
Se registra los datos del abono y mediante la opcion de Sincronizar este se envia al S10.
De ser exitoso el envió nos muestra los check de color verde caso contrario se mostraran de color rojo
Enviar Minuta
La opción de enviar Minuta al S10 se activara únicamente cuando se haya cumplido con los Flujos de proceso de Separación y Venta subiendo finalmente el Contrato y/o Minuta.
Flujo del Proceso: Separación
Se debe aprobar cada uno de los pasos dentro del flujo, dándole clic a cada línea se habilitara la siguiente ventana donde se marcara el check para Aprobar
Cabe indicar que dentro de cada paso podemos cargar un archivo en cualquier formato
Finalizada la aprobación de cada paso, se habilitara la siguiente opción permitiendo subir algún contrato o compromiso de la separación
Adjuntando archivos solo en formato PDF
A continuación se debe solicitar la aprobación respectiva, enviando una notificación a los aprobadores del proceso.
De igual forma deben aprobarse los siguientes flujos, como primer paso debe iniciar cada uno de los procesos el de Aprobación y Venta
A continuación debe seleccionar el flujo de acuerdo a lo establecido por cada empresa
Seguidamente debe ser aprobado cada paso que se encuentra dentro de cada uno de los flujos remarcados
Finalmente debe subirse el contrato de Venta en formato PDF
La acción de subir dicho Contrato de Venta, habilita la opción de Enviar Minuta al S10
Al hacer uso de la opción Enviar Minuta solicitara registrar la fecha de Contrato
Al darle clic en Enviar esta información viajara al S10
S10
En el escenario documentos de ventas se registran de manera automática todas las órdenes de venta que han sido enviadas desde el Sperant
Mediante la pestaña Cronograma de Facturación que se encuentra dentro de la Orden de Venta podemos ver el detalle de los abonos enviados desde Sperant
Mediante la pestaña Sperant que se encuentra dentro de la Orden de Venta podemos ver el detalle del Envió de Minuta
Como se indicaba inicialmente para realizar el proceso de “Cambio de Titular” si uno de los abonos cuenta con Boleta, los demás abonos también deben estar Boleteados, caso contrario no permitirá hacer el cambio.
Boletaje
En el escenario Boletaje se procede a generar los comprobantes siendo Facturas para Persona Jurídica o Boleta para Persona Natural, de acuerdo a lo configurado.
Tipo de Boletaje
Mediante el siguiente filtro se mostraran los abonos de acuerdo a lo enviado.
Abono hito inicial: Aquí se muestran los abonos a los cuales se le generara un comprobante inicial, siendo Factura o Boleta tendrán la suma del total de los abonos enviados hasta la fecha.
Abonos posteriores: Aquí mostrara los abonos siguientes, emitidos después del primer comprobante.
Sperant
Para realizar el proceso de cambio de titular debe ubicarse dentro de la proforma
Mediante la opción del Lápiz permitirá editar el nombre del Nuevo Titular del Inmueble, desplegando así el listado de los clientes ya registrados
Finalmente dar clic en la opción guardar así el proceso guardará el cambio del nuevo titular
Cambio de Titular
Mediante el siguiente escenario se registran los cambios de titular que pueden presentarse en el proceso de la venta de inmuebles.
Mediante la acción de Cambio de Titular que se realiza en el Sperant, el S10 almacena la siguiente información
Datos del antiguo cliente y su venta, así como el dato del nuevo cliente y su venta
Al hacer clic sobre la grilla mostrara las siguientes opciones
Ver documento –
Origen: Nos muestra el detalle de la Orden de venta Inicial
Ver Documento –
Destino: Nos muestra la nueva orden de venta cuyo documento aún se muestra como
Cotización
Generar facturación:
Esta opción se muestra cuando el estado del cambio de titular está aún
Pendiente
porque no se ha concluido con el flujo.
Al hacer clic en la opción, permitirá seleccionar una fecha que servirá para generar todos los documentos.
Al finalizar la acción muestra el siguiente mensaje
Orden de Venta Original
Cumple el siguiente proceso:
Se registra una Nota de Crédito vinculada a la boleta considerando que la nota debe sumar el monto de la boleta o boletas según sea el caso.
La orden de venta original luego de registrar la nota de crédito debe ser colocada en estado CERRADO POR EL USUARIO
El tipo de documento nota de crédito lo tomas de la configuración de facturación, tal como se muestra en la imagen.
Orden de venta Nueva
Cumple el siguiente proceso:
• La orden de venta nueva, que estaba en tipo cotización pasa a tipo Orden de Venta
• En cronograma facturación se vinculada la Boleta
• La orden de venta nueva como ya tiene boleta, su estado pasa a PARCIALMENTE
FACTURADO
Finalmente, en el módulo de administrativo se registra la compensación y en automático para cancelar las notas de crédito con las boletas. Notas de crédito por pagar, y boletas por cobrar
