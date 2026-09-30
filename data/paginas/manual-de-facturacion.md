# Manual de Facturación

Fuente: https://documentacion.s10peru.com/manual-de-facturacion/

Table of Contents
Toggle
1 Introducción
1.1 La carpeta de trabajo
Las ventanas que presenta el S10 se asemejan a un escritorio donde tiene todos los documentos necesarios para trabajar en forma ordenada. El entorno de trabajo es el mismo casi en todos los escenarios, a continuación se describe las partes de un escenario.
1.2 Acceso al módulo de Facturación
Del tapiz de su PC elija el icono con doble clic.
Esta ventana permite conectarme al entorno de trabajo en este caso al módulo de facturación, donde debe ingresar los siguientes datos que serán proporcionados por el administrador de sistema de su empresa.
2. Catálogos
Para poder visualizar los catálogos nos ubicamos en el Menú > Catálogos
2.1 Moneda
El sistema por defecto nos brinda dos monedas, de Nuevos Soles y Dólares Americanos.
Se considera la moneda Nuevos Soles como moneda Base y Principal.
En caso se quiera adicionar una nueva moneda, elegimos la opción click derecho > Adicionar.
En este caso NO se elige moneda de Control, ya que si elegimos Principal se deshabilita para la moneda Nuevos Soles y la moneda secundaria sería Dólares Americanos.
Si se quiere eliminar se da click derecho > Eliminar
Nos va a salir un mensaje al cual le damos la opción SI.
2.2 Forma de Pago
Se encuentran las modalidades de Pago con las que cuenta la empresa.
Para adicionar una nueva forma de pago se da click derecho > Adicionar
2.3 Tipo de cambio
Éste catálogo se actualiza automáticamente con los datos de la página de la Sunat, en caso esté configurado de esa manera.
2.4 Proyectos
Nos ubicamos en el catálogo Proyectos y adicionamos un proyecto.
2.5 Socio de Negocio
2.5.1 Clientes
Para la importación de Excel se elige el icono como se indica a continuación:
Se ubica la ruta del documento.
Se unen los campos del S10 con la información ingresada en el Excel.
Se ubican las filas que contienen la información, la columna se toma automáticamente de acuerdo a la última columna que se unió con el campo. Finalmente, se da click en Importar.
En caso se quiera agregar manualmente, nos ubicamos en el catálogo de Clientes > Click derecho > Adicionar, sale la siguiente ventana:
En la pestaña TIPO debe agregar, el tipo de socio de negocio al que pertenece.
En la pestaña Cuentas Banco se agrega el número de cuentas del cliente para poder efectuar los pagos en cuenta bancaria.
2.5.2 Proveedores
El proceso para los proveedores es el mismo que el de los clientes.
2.6 Tipo de Impuestos
Se visualizan los Impuestos que maneja la empresa.
En caso se quiera agregar un nuevo impuesto se da click derecho > Adicionar.
Se procede a llenar los datos indicados.
2.7 Tipo de Documentos
En este catálogo contiene la relación de documentos que utilizará la empresa.
• Catálogo de tipo de Documentos de Sunat: Esta información es publicada por la Sunat.
• Tipo documento Base: Son los formatos o formularios de ingreso de datos preparados por el S10 para el uso en los diferentes tipos de documentos para su organización.
2.8 Detracción
Para revisar el catálogo de Detracciones nos en el catálogo Sunat.
2.9 Retención, Percepción y Buenos Contribuyentes
Para agregar los agentes de Retención, Percepción y Buenos Contribuyentes primero se debe bajar de la página de la Sunat los padrones en formato .txt
Se descarga el archivo en formato .zip y se extrae el fichero.
Luego, en el catálogo Sunat se elige la opción Agentes de Retención.
Finalmente, se importan los Agentes de Retención.
3 Accesos
3.1 Creación de las empresas y sucursales
Para agregar una nueva empresa nos ubicamos en el catálogo Socio de Negocio > Empresa del Grupo.
En caso sea sucursal, se deberá agregar un socio de negocio en la pestaña Sucursales.
Finalmente, el árbol quedará de la siguiente forma:
3.2 Definición de usuarios y sus accesos
Creamos un nuevo grupo al cual le vamos a poner el nombre de la empresa.
Ahora nos ubicamos en el Subgrupo creado y agregamos al usuario
En caso el empleado no esté en el catálogo, lo procedemos a crear.
En la pestaña Escenarios nos vamos a dar acceso a los escenarios del módulo.
Ahora, vamos a dar permisos por empresa, para ello nos ubicamos en el grupo Inicio > Datos Generales.
Finalmente, ingresamos al sistema con el usuario creado.
3.3 Definición de Centros de Costos
Para agregar un centro de costos primero debemos definir su estructura.
Tipeamos las descripciones y si queremos cambiamos los colores.
Ahora, importamos el archivo según el formato en Excel del centro de costos.
3.4 Configuración general de registro de documentos
Para poder realizar la configuración se debe ingresar como SA.
Se habilitan y deshabilitan los checks de acuerdo a la realidad de su empresa.
3.5 Periodos contables
3.6 Talonarios
3.6.1 Documentos
Agregamos al usuario que podrá usar el talonario.
Una vez creado el talonario, se le brinda acceso al usuario.
3.6.2 Comprobantes de Retención
Para los comprobantes de Retención primero se debe crear en el catálogo de documentos.
Una vez creado el documento, procedemos a crear el talonario.
Y luego procedemos a agregar el documento.
Asignamos Acceso a Estados al usuario que podrá utilizar el talonario respectivo.
4. Ventas
4.1 Inventario de productos terminados
En el catálogo Recursos vamos a adicionar nuestros productos de la siguiente manera:
4.2 Recursos de Ventas
Primero para poder relacionar el recurso con el proyecto debemos ingresar al Catálogo Configuración.
Ahora, nos ubicamos en el grupo Ventas, escenario Recursos de Ventas.
Asignamos el proyecto a los recursos.
Quedaría de la siguiente manera:
4.3 Lista de precios
Nos ubicamos en el escenario Lista de Precios y le damos adicionar.
Nos aparecerá el siguiente recuadro:
Hacemos lo mismo para cada recurso por proyecto y nos debe quedar de la siguiente manera:
4.4 Cotizaciones
Las cotizaciones son documentos informativos que establecen el precio a los recursos de ventas ingresados en la lista de precios.
4.5 Órdenes de Venta
Una vez aprobada la cotización por parte del cliente se debe generar la Orden de Venta, nos ubicamos en la cotización y le damos click derecho.
4.6 Cronograma de facturación desde Órdenes de Ventas
Cuando haga clic en Aceptar en la Orden de Venta, aparecerá la siguiente ventana:
4.7 Documentos por cobrar desde Órdenes de Ventas
Una vez que se da clic en Generar Documento de Facturación aparecerá la factura.
5 Documentos por cobrar
5.1 Importación de saldos
Para poder importar saldos primero debemos llenar el formato en Excel.
En el módulo de Facturación, creamos un recurso con el nombre RECURSO PARA LA IMPORTACIÓN DE SALDOS.
Ingresamos al sistema como SA, nos ubicamos en el grupo Facturación, escenario Documentos por cobrar y le damos click al icono del lápiz.
Volvemos a ingresar con nuestro usuario y procederemos a la importación de saldos por cobrar.
Aparecerá el siguiente mensaje:
Clic en OK y ubicamos la ruta del archivo.
Relacionamos el campo con las columnas y damos click en Importar.
Nos quedará de la siguiente manera:
Nota: Para que importe los documentos verificar que los socios de negocios existan en el catálogo de clientes, lo mismo que los tipos de documento y la moneda, todos deben estar conforme a los datos ingresados al sistema.
5.2 Agregado de recursos
En el caso de que los documentos por cobrar no cuenten con relación a otro documento, se agregarán recursos para la emisión del mismo.
5.3 Documentos por Cobrar sin relación
Para crear documentos independientes, sin relación a Órdenes de Venta, nos ubicamos en el escenario Documentos por cobrar.
5.4 Registro de documentos con Documento de Venta
Para la creación de documentos con Documentos de Venta es necesario seguir el punto 3.
5.5 Registro de Notas de Crédito o Débito
5.5.1 Nota de Crédito
Para la creación de Notas de Crédito se debe realizar lo siguiente:
Damos click derecho Adicionar y nos aparecerá el siguiente recuadro:
Nos aparecerá las facturas que tenga el socio de negocio.
5.5.2 Nota de Débito
La creación de la Nota de Débito tiene el mismo procedimiento que una Nota de Crédito.
5.6 Registro de documentos por cobrar internos
5.6.1 Anticipos
Primero se debe crear el documento en el catálogo Tipo de Documentos.
Luego nos ubicamos en el escenario Documentos por cobrar y seleccionamos el Anticipo.
5.6.2 Transferencia Bancaria
Agregar el documento en el catálogo Tipo de Documentos.
Desde el escenario Documentos por Cobrar adicionamos el documento.
Al momento de crear la Transferencia Bancaria por cobrar, se creará en el escenario de Documentos por pagar.
5.6.3 Letras Únicas de Cambio
Las letras se generan automáticamente al momento de hacer una Canje con Letras en el módulo Administrativo, pero en el módulo Facturación se configura el documento.
Luego se debe agregar el talonario de las letras.
Agregamos al usuario.
Asignamos los accesos.
Significa que el usuario podrá utilizar el talonario.
NOTA
Se pueden seguir agregando Documentos Internos conforme sea necesario, tales como Nota de Abono Bancario, Intereses por Letras, Préstamos, etc. En los cuales el tipo de documento base será 00 Otros.
5.7 Configuración e impresión
En la Barra Estándar se encuentra un icono de engranaje al cual daremos click.
5.8 Liquidación de Cobranza
Se utiliza cuando un cliente nos da un pago anticipado el cual se va a regularizar con otro documento, para ello utilizaremos el documento Anticipo de Clientes creado en el punto 5.6.1
5.9 Portafolio de Cobranzas
Se utiliza cuando contamos con varias facturas que se asigna a un empleado por cobrar.
6 Documentos por pagar
6.1 Importación de saldos
Este proceso se podrá explicar mejor en la implantación, con el formato en Excel que el implantador le facilitará y que contará con los documentos pendientes según la fecha de corte.
6.2 Configuración General de Documentos
La Configuración General de Documentos, se realiza exclusivamente con el SA.
6.2.1 Impresión
6.2.2 Supervisor 1
6.2.3 Supervisor 2
6.2.4 Supervisor 3
6.2.5 Supervisor 4
Configuración que permitirá realizar el procedimiento de Emisión de Comprobantes electrónicos.
6.3. Documentos por Pagar
6.3.1 Documentos sin Relación
Desde el Grupo Por Pagar, se procede a realizar el registro de operaciones, seleccionando previamente el tipo documento.
Clic Derecho sobre el escenario para continuar con el registro de las características del documento emitido por proveedores o terceros.
Registrar las características del Documento, socio Negocio, Serie y número de documento, Moneda, Forma de Pago, proyectos y Clic en el detalle para agregar los recursos del catálogo y determinar cantidad y precios Unitarios.
6.3.2 Documento relacionado con:
a. Ingreso por Compra: Guía registrada desde el Módulo Almacenes que representa el ingreso de mercadería u Otros a los diferentes Tipos de almacenes creados en S10, la relación de la Guía y el documento representa la correcta regularización de información.
b. Orden de Compra y Servicio: OC o OS registrada desde el módulo de Compras, Es posible relacionarla con el comprobante pero será al 100% de la OC o OS, se recomienda trabajar con los ingresos relacionados a la OC.
c. Subcontrato: el registro de las valorizaciones, pagos a cuenta, adelantos aprobados desde el Módulo de Gerencia de Proyectos determina la estructura que permite relacionarlo con el comprobante de pago para realizar el control correspondiente.
d. Ingreso orden de Compra o Servicio: Son guías registradas desde Almacenes que determinan el ingreso parcial o total por el despacho de la orden de compra, a través de estas operaciones es posible evaluar la correcta entrega del contrato establecido.
6.3.3 Documentos por Pagar Electrónicos
6.3.3.1 Definición de Comprobantes Electrónicos en el Registro.
Operaciones registradas desde el escenario documentos por pagar y se determina la activación del check al confirmar que sean Electrónicos.
6.3.3.2 Configuración Predeterminada de Comprobantes Electrónicos
Configuración predeterminada desde el catálogo del Tipo Documento, se establece que será comprobante electrónico.
Vista del registro del Documento.
6.4 Liquidaciones
6.4.1 Registro del Comprobante Entrega a Rendir.
Iniciaremos con el Registro del Entrega a rendir con la Configuración de un Tipo Documento ENTREGAS A RENDIR que nos permita realizar el pago y posteriormente este pueda ser sustentado.
Clic derecho Adicionar para la creación de este Tipo de Documento.
Indicar las características del Tipo de Documento, descripción, tipo de aplicación, tipo documento base OTROS porque es un documento Interno.
Luego de la Creación del Tipo Documento se procede desde el Módulo de Facturación Grupo por Pagar Escenario Documentos por Pagar la creación del documento ENTREGAS A RENDIR, en el S10 no se realiza ningún desembolso en el módulo Administrativo sino es previamente sustentado con un documento Interno o SUNAT, en este caso se realizará un entrega a rendir a un personal de la empresa con la finalidad que posteriormente me rinda los gastos realizados por dicho desembolso.
Clic derecho dentro del escenario documentos por pagar, para iniciar la creación del Entrega a Rendir.
Luego de utilizar la opción de ADICIONAR para la creación del documento aparecerá una ventana donde ingresaremos los datos de la persona a quien se realizará el desembolso, adicionalmente la fecha, proyecto, glosa, Recurso e importe y luego guardamos el documento con la opción de aceptar.
Vista de documento guardado dentro del escenario Documentos por Pagar.
6.4.2 Pago del Comprobante Entrega Rendir
Luego de tener el documento se procederá al pago del documento ENTREGA A RENDIR, en el módulo Administrativo, grupo Pagos, escenario Orden de Pago.
Clic derecho Adicionar sobre el escenario Orden de Pago y aparecerá una venta donde indicaremos los datos de la cuenta Bancaria con la que realizaremos el desembolso, descripción de la orden de pago, y adjuntaremos el documento entrega a rendir creado anteriormente desde el módulo de facturación.
En el detalle de la orden de pago al utilizar la opción de Agregar Documento aparecerá una ventana CRONOGRAMA DE PAGOS, donde debe aparecer el documento creado Entrega a Rendir.
Para seleccionar el documento Doble clic en la línea del documento y luego utilizar el icono de parrilla.
Al ser agregado el documento a la orden de pago, procederemos a guardar la orden de pago para que se envié a Aprobación.
Una vez grabado la orden de pago debemos enviar a aprobación la Orden de pago, al realizar este paso a los aprobadores del módulo administrativo les llegará un correo electrónico para que ingresen al S10 a realizar la aprobación de la OP.
x
Realizar la Confirmación del envío de la OP.
Al realizar la confirmación de la OP el estado de la orden cambiará a POR APROBAR donde el usuario que sea aprobador deberá de ingresar a aprobar para culminar con el proceso del pago de la Entrega a Rendir.
En este caso el Usuario CONSULTOR1 es aprobador previos permisos, clic derecho opción APROBAR.
Luego de realizar la aprobación de la OP nos dirigimos al escenario PENDIENTES DE PAGO, para efectuar el pago e iniciar con el registro de la Liquidación.
Clic derecho en la línea de la OP, en este caso lo reconocemos por la descripción y el socio negocio a quien se le realizará el pago.
Al utilizar la opción de pagar Documento Actual emergerá una ventana de Pago donde se visualizará el documento ENTREGAS A RENDIR, listo para ser pagado.
Pago – En este caso se realizará con un cheque.
Al Grabar el PAGO utilizando la opción ACEPTAR, el pago realizado se mostrará en el escenario de PAGOS.
6.4.3 Proceso de Liquidación desde Facturación
Luego de culminar con el Pago del documento Entrega a Rendir, recién se apertura una Liquidación que es el escenario que nos ayudará para cancelar los documentos que sustenten por la salida ER.
Para Iniciar con la Creación de las LIQUIDACIONES, en el escenario Clic derecho Adicionar, y aparecerá una ventana para el registro de la información de la Información.
Al utilizar la opción de Adicionar aparecerá una ventana de Liquidaciones iniciaremos registrando el nombre de la persona a quien se le realizó el pago ya que es la persona que nos realizará las rendiciones.
Para el registro de una Liquidación se inicia con agregar el pago realizado a la persona que posteriormente al desembolso rendirá con documentos internos o SUNAT el total del importe total o parcial desembolsado.
Al Utilizar la opción de AGREGAR PAGO aparecerá una ventana y aparecerán exclusivamente los documentos Pagados filtrando los documentos a quien se le está realizando la liquidación.
En esta ventana seleccionaremos con doble clic el documento pagado para agregarlo a la Liquidación.
Luego de seleccionar el documento para agregarlo a la liquidación iniciamos una liquidación pendiente por rendir de esta manera administraremos las salidas de dinero y análisis de sustentos a rendir.
Al iniciar la creación de la Liquidación y agregar el pago la pestaña de Pagos indica que se tiene POR RENDIR S/. 1000.00.
Para el control de los documentos Internos o SUNAT que en este caso CARDENAS ACERO, YESSENIA comience a rendir se trabarán en la pestaña de Documentos.
Para este caso utilizaremos la opción de GENERAR DOCUMENTO crearemos un documento que este rindiendo.
Al seleccionar el tipo de documento planilla de Movilidad, aparecerá un formulario para el registro de la información que este sustentando. Para este ejemplo registraremos un gasto por movilidad Tipo de Documento – Planilla de Movilidad S/. 10.00
Al culminar con el registro del documento interno en este caso por el concepto de Movilidad, asimismo se pueden realizar registros de documentos internos o SUNAT e ir disminuyendo la rendición.
Luego de realizar el registro de este documento procederemos a grabar la liquidación, como aún no está liquidada para continuar con el registro de la información ingresaremos utilizaremos la opción de Modificar.
Para continuar con la Liquidación aún no liquidada utilizaremos la opción de MODIFICAR para continuar.
Para continuar con el registro de la información Crearemos un tipo documento llamado DECLARACIONES SUNAT, doble clic para su creación.
En el tipo documento Declaración SUNAT indicaremos que son intereses por Infracción Tributaria por el monto de S/. 900.00.
Vista de la Liquidación con el control de Saldos.
Vista del Escenario de Liquidaciones.
Para la Devolución del Importe se tiene que realizar un depósito en la cuenta de la Empresa MI CASITA S.A.C. para este caso debemos de registrar un documento Interno para ingresarlo por diario de Caja y luego agregarlo a la liquidación y cerrarla finalmente.
Luego de la Creación del documento Interno se debe realizar el Ingreso del documento interno al diario de caja Módulo Administrativo (Debe ser Ingresado o depositado en la cuenta).
Realizamos la búsqueda del documento para realizar el ingreso del depósito realizado por la entrega a rendir.
Datos Importantes en la Cobranza, en este caso es un deposito indicar Cuenta bancaria y número de Operación para estos casos Guiarse del voucher que le deben de entregar y sustente el depósito para la liquidación y entrega a rendir.
Al grabar el ingreso a la cuenta del depósito realizado por la devolución del Entrega a rendir, se tiene esta vista en el diario de caja.
Luego de realizar el ingreso nos dirigimos al módulo de FACTURACION escenario LIQUIDACIONES y modificamos la que estamos trabajando.
Luego de Abrirse la Liquidación nos ubicamos en la pestaña de DOCUMENTOS y clic derecho opción AGREGAR DOCUMENTO DESDE EL DIARIO DE CAJA.
Al Utilizar la opción aparecerá una ventana CRONOGRAMA DE COBRANZAS donde tendremos que realizar la búsqueda de nuestro depósito, por socio negocio al encontrarlo lo seleccionaremos con doble clic y luego utilizaremos la parrilla.
De esta manera nuestra LIQUIDACION quedará completamente rendida y cuadrada. Clic en Aceptar para grabar.
Al detectar el S10 que la LIQUIDACION está completamente rendida de manera automática convierte de estado registrado a Cerrado.
6.5 Reembolsos
6.5.1 Registro del Reembolso
El registro del reembolso tiene como finalidad registrar información cuyo importe total acumulado dentro de la bolsa es el que se reembolsará al Responsable.
Para realizar la creación del Reembolso sobre el escenario Clic derecho opción Adicional.
Aparecerá la Ventana del reembolso, se tendrán que registrar los datos del Socio Negocio a quien se le reembolsará.
Luego de registrar los datos del reembolso, en el detalle, Pestaña de Documentos se utilizará la opción de GENERAR DOCUMENTO es para la creación de un documento; opción AGREGAR DOCUMENTO es jalar o tomar un documento que se encuentra previamente registrado.
Luego de utilizar la opción de GENERAR DOCUMENTO aparecerá el catálogo de tipo de documento, se debe de seleccionar el tipo de documento a crear, para este caso crearemos una Factura. (Doble clic Factura Terceros).
Al darle doble clic aparecerá el formulario del tipo de documento elegido, en este caso Facturas Terceros, el cual comenzaremos el registro de la información.
Para este caso registraremos una factura por el Concepto de Movilidad por Viaje, indicar el proveedor (Socio negocio), moneda del documento, número del comprobante, fechas del documento, indicar el proyecto que genera el gasto, registrar glosa, agregar el recurso e indicar los importes.
Luego de registrar el detalle del documento, para grabar opción ACEPTAR.
Vista del Reembolso, con el primer documento creado para el reembolso por gastos.
Así como se ha registrado un comprobante dentro del reembolso, es posible registrar más de uno, para este caso solo registraremos un documento que se le reembolsará a ARIAS, JUAN, ya que el proveedor fue pagado al momento del desembolso del gasto.
Luego de grabar el reembolso lo visualizaremos dentro del escenario de esta forma.
Luego de grabar el Reembolso se tiene aún en estado ABIERTO, esto indica que el reembolso se puede continuar trabajando mediante la opción de Modificar.
El área de Finanzas o tesorería se dará cuenta que debe pagar el reembolso cuando el reembolso se encuentre en estado CERRADO.
Vista del Reembolso en estado CERRADO, desde el Módulo Administrativo solo visualizará los que se encuentren cerrados actos para generar el Pago.
6.6 Fondos Rotatorios
6.6.1 Configuración del Fondo Rotatorio.
Configuración Inicial para trabajar los Fondos Rotatorios, El socio negocio o Usuario llamado ADMINISTRADOR DE FONDO ROTATORIO, es el encargado de registrar los fondos rotatorios para todos los proyectos, también tiene atributos para corregir cualquier fondo rotatorio
Verificar que el Usuario sea RESPONSABLE DE FONDO ROTATORIO, es la persona encargada del fondo solo él puede adicionar y su fondo asignado, puede ser responsable de más de un fondo y cada uno tendrá una administración independiente del módulo Facturación o Almacenes en el grupo Inicio Escenario Datos Generales.
6.6.2 Registro del Fondo Rotatorio
Registro del Fondos Rotatorios dentro del Grupo por Pagar escenario Fondos Rotatorios.
Al Utilizar la opción ADICIONAR aparecerá una ventana de fondos rotatorios donde se tendrá que ingresar algunos datos informativos para el inicio de la Caja Chica.
Existen Campos obligatorios por registrar en el Fondo Rotatorio.
Luego de registrar los datos complementarios Iniciaremos con los movimientos de la Caja.
Desde la Ventana de Fondos Rotatorios, existe una pestaña llamada ASIGNACIÓN DE FONDO, al ubicarnos Clic Derecho y aparecerá dos opciones:
a.
Agregar Documento de Pago; Significa Agregar un Documento Previamente registrado en el Módulo de facturación documentos por pagar y que se encuentra previamente PAGADO por el módulo Administrativo.
b.
Generar Documento de Pago; Esta opción nos permitirá crear un documento dentro del fondo rotatorio para que el módulo Administrativo pueda realizar el pago.
Ambas Opciones son válidas, la diferencia una de otra es que (a) permitirá realizar las operaciones o gastos de inmediato es decir crear los reembolsos, (b), creará recién el documento y estará a la espera del pago del documento creado para el inicio de las operaciones de la caja.
En este Caso Trabajaremos con la Primera opción de Agregar un Documento PAGADO ya que la solicitud de Caja Chica se realiza con anticipación previo Correo electrónico o Registro de documento Interno y se solicita su pago.
Desde la pestaña Asignación de Fondo, clic derecho Agregar Documento de Pago.
Al Utilizar la opción mencionada en el paso 6. Aparecerá una ventana que me permitirá realizar la búsqueda del documento Pagado, detectarlo por el nombre del responsable o socio negocio quien llevara el Fondo Rotatorio.
Al Ubicar al Documento Doble Clic.
Luego se tener Seleccionado el Documento Fondo Rotatorio PAGADO utilizar la opción denominada como PARRILLA.
Luego de Agregar el Documento Pagado a la Asignación de Fondo Que indica que estas tomando el dinero que la EMPRESA te está abonando a tu cuenta vía transferencia o por un cheque para el inicio de tus operaciones en la caja Chica, aceptas el Fondo para guardar la aceptación del dinero.
Luego visualizarás en el escenario de Fondos Rotatorios el que estamos trabajando, Clic derecho opción MODIFICAR.
Luego de Abrir el Fondo Rotatorio nos Ubicaremos en la pestaña de MOVIMIENTOS, al visualizar la línea de Asignación de Fondo Indica que Usted ya puede hacer uso del Dinero y podrá crear los Documentos Internos y SUNAT que sustenten sus gastos.
Dentro de la pestaña MOVIMIENTOS podrá Crear los Reembolsos que es el Agrupador de Documentos que tendrá como contenido los documentos que sustenten sus Gastos. Realizaremos en este caso un Ejemplo.
Al darle Clic en generar Reembolso aparecerá una ventana donde registraremos nuestros Gastos.
Iniciaremos con el registro de Documentos dentro del Reembolso, para este caso utilizaremos la opción de Generar Documento que es sinónimo de Crear documentos nuevos la segunda Opción se utilizará siempre y cuando exista documentos creados en documentos por pagar que se puedan agregar.
Al generar Documento aparecerá una ventana de Tipo de documento donde deberás Elegir qué tipo de Documento Crearás.
Seleccionar con Doble clic el Tipo de Documento a Crear.
Al Seleccionar el Tipo de Documento aparecerá el Formulario elegido en este caso Facturas, aquí estamos iniciando con el registro de documentos que sustenten nuestros gastos, en este caso un comprobante SUNAT.
Al Crear un documento SUNAT o Interno debemos de registrar los datos del proveedor, en este caso Realizaremos una compra de Útiles es importante registrar Socio negocio, Número de Documento, Fecha, Glosa, Moneda, el proyecto es propuesto ya que inicialmente se le indicó en el fondo Rotatorio.
Luego de Digitar los datos del comprobante, debemos indicar los datos del detalle del comprobante que es lo que estamos comprando para este caso utilizaremos un catálogo llamado Recursos o Ítems que especifica toda una gama de consumibles entre otros por Utilizar.
La opción de Adicionar me permitirá dentro del documento elegir el recurso que estamos comprando.
Realiza la Búsqueda del recurso para seleccionarlo y agregarlo al Documento.
Seleccionar con Doble clic el recurso encontrado y luego utilizar la Parrilla.
Al Seleccionar el Recurso UTILES DE ESCRITORIO indicaremos la cantidad y precio Total, Cuando se compren en cantidades y precio unitario utilizaremos los campos a definir.
Al seleccionar el recurso, indicaremos los importes de cantidad y precio.
Luego de Digitar la cantidad y el importe del comprobante de pago existe en la parte inferior de documento una opción llamada INGRESO A ALMACÉN, esta opción permite al usuario realizar el ingreso a almacén y realizar la descarga directa eligiendo una partida de control realizándose el consumo en el instante.
Al Guardarse el documento podremos visualizar que el reembolso irá sumando el importe total de cada documento, asimismo dentro del mismo Reembolso podremos crear más documentos SUNAT e Internos. En este Caso guardaremos el reembolso para continuar con el ejercicio.
Guardar el Reembolso.
Vista del Fondo Rotatorio con un Reembolso en estado Registrado, donde el saldo está disminuyendo en base al fondo asignado.
Opción de Ver Reembolso para continuar Registrando los documentos que sustenten los gastos.
Para la Devolución de lo gastado el RESPONSABLE deberá de utilizar la opción de ESTADO (Cerrado) de esta manera el área de FINANZAS podrá saber que debe realizar una devolución del importe que sume el reembolso.
Vista del Fondo Rotatorio guardado, si se le da clic derecho es posible modificarlo para poder continuar con el registro de información.
6.7 Liquidaciones, Reembolsos y Portafolios Detallados
Escenario que permite visualizar de forma detallada cada registro de operación Liquidaciones, reembolsos y Portafolios, es posible realizar la agrupación y visualizar la información origen.
6.8 Cronograma de Pagos
Desde el escenario Cronograma de Pagos, se Observa todos los documentos por Pagar, a través de este escenario es posible realizar reportes para evaluar el estado de los documentos montos programados y saldos pendientes por pagar.
7 Informes
7.1 Compras y Ventas
Visualización de Libros Auxiliares, Compras, Ventas, Honorarios, retenciones, y documentos Internos calificados por registro de gastos no sustentados ante SUNAT.
Impresión de Libros Auxiliares por Tipos.
7.2 Compras y Ventas Detallados
Visualización del registro de información de forma detallada a nivel Recurso, Por centro de costos, Por Proyecto.
7.3 Detracciones
El Escenario Detracciones tiene la Finalidad de emitir informes que permitan al Usuario realizar la explotación de datos como:
• Pendientes por Pagar / Cobrar Detracciones.
• Estado del Cronograma Pagos o Cobranzas del Tipo Detracción para evaluar la declaración Mensual según su condición de Pagado o Cobrado.
• Vista de Comprobantes Pagados / Cobrados que falta el registro de la Constancia Detracción.
Escenario Detracciones permite Visualizar información del Tipo Por Pagar, Por Cobrar, Ambos; Asimismo Permite visualizar todos los periodos o un periodo exclusivo a revisar por el Usuario.
Pendientes por Pagar / Cobrar Detracciones.
Desde el Filtro del Campo ESTADO DETRACCIÓN, es posible visualizar los documentos Pendientes por Pagar o Cobrar, para considerarlos en el periodo a Declarar.
Asimismo se tiene campos que indican Monto Detracción, rubro y % de acuerdo al catálogo de Ítems Detracción cuya configuración se le realiza al Recurso.
Análisis de Información por Periodo, Estado Detracción y campos % Detracción, Rubro Detracción Monto Detracción por documento.
Análisis de Información por Periodo, Estado Detracción, por socio Negocio y por Documento.
Desde el Escenario Detracciones es posible Visualizar el Documento Origen.
Vista de Comprobantes por Periodo, estado PAGADADO, que no tienen registrado el número de Constancia detracción y Fecha de Constancia Detracción.
Desde el Escenario es posible Ver Documento y registrar la Fecha y número de Constancia Detracción.
Vista Todos los Periodos, por Estado Detracción.
Vista Todos los Periodos, Por Estado Detracción, Por Socio Negocio.
Vista de Impresión del Escenario de Declaraciones.
7.4 Autorización de Subcontrato / Contrato
Escenario de Consulta, permite analizar o revisar las autorizaciones de Pago o Cobranza ya sea en el periodo mensual o semanal de valorizaciones, pagos a cuenta, adelantos, solo se visualizan las que se encuentren aprobadas. A través de los campos se Observa la relación con los documentos por Cobrar, Pagar y documentos internos.
7.5 Pedidos (Almacén y Compras), Orden de Compra, Guías de Almacén.
Escenario que permite evaluar las operaciones realizadas en el módulo de Almacenes y Compras, asimismo es posible realizar la trazabilidad de los movimientos.
7.6 Cuentas Bancarias del Socio Negocio
Escenario que permite visualizar la cantidad de Cuentas Bancarias que tiene un socio negocio (Cta. Corriente, Ahorros, Maestra, CTS, Detracción), entre otros; permite visualizar la información a través de la opción Propiedades; a través de la opción Importar desde Excel permite realizar la actualización de Cuentas bancarias para cada Socio Negocio.
7.7 Guías de Ingreso sin Regularizar
Escenario que permite visualizar cuales son los movimientos (Guías de Almacén) que no tienen relación con un comprobante de Pago.
Vista de Movimiento de Almacén sin Relación a un comprobante de Pago.
8 Configuración
8.1 Centros de Costos
8.1.1 Definición de Estructura
Clic derecho desde el escenario, elegir Opción Definir Estructura.
Aparecerá una ventana, el usuario deberá de realizar el registro de la estructura para la creación de los centros de Costos.
8.1.2. Registro de los Centros de Costos
Registro de Centro de Costos, relacionado al Centro de Costo General para la contabilización de operaciones en las que existan cuentas contables que dependan de centro de costos, relacionado al catálogo de proyectos. La Calificación de centros de costos permite establecer reportes de todas las operaciones a través de Líneas de Negocio y destinos de Venta.
Vista Final de los Centros de Costos
8.2 Asignación de Tributos
8.2.1 Configuración del Catálogo de ítems Detracciones / Percepciones
Tener en cuenta que para iniciar con la configuración de recursos con Detracción y Percepción es necesario tener actualizados los Catálogos con los últimos porcentajes y monto Aplicable según los cambios que SUNAT realice en el Tiempo.
a.
Ingresamos al Módulo de Facturación/ Catálogos/ SUNAT/ Detracciones.
Luego de Ubicar el catálogo de Detracciones aparecerá una ventana que permitirá visualizar el catálogo de Ítems Detracciones donde es posible mediante las opciones de Adicionar, Modificar, Duplicar y Eliminar las líneas con la finalidad de Actualizarlas de acuerdo a las publicaciones que SUNAT realice.
Clic Derecho sobre la Línea del Ítems de Detracciones, Aparecen las Siguientes Opciones.
En este Caso Utilizaremos las opción de MODIFICAR, para actualizar los rangos de aplicación de Porcentajes y Montos a Aplicar.
Desde la Ventana del Ítems Detracciones es posible configurar el historial de cambios en % y Montos de Aplicación para que al ser utilizado tome en cuenta según las fechas del registro de los comprobantes según el Ítem configurado en el Recurso.
b.
Ingresamos al Módulo de Facturación/ Catálogos/ SUNAT/ Percepciones.
Luego de Ubicar el catálogo de Percepciones aparecerá una ventana que permitirá visualizar el catálogo de Ítems Percepciones donde es posible mediante las opciones de Adicionar, Modificar, Duplicar y Eliminar las líneas con la finalidad de Actualizarlas de acuerdo a las publicaciones que SUNAT realice.
Clic Derecho sobre la Línea del Ítems de Percepción, Aparecen las Siguientes Opciones.
En este caso se procede a Utilizar la opción de MODIFICAR para actualizar el % y monto aplicación.
Tener en cuenta que para Aplicar la Percepción no solo es configurar el Recurso relacionado a los Ítems de Percepción, En las Propiedades del socio Negocio o Proveedor debe estar calificado como AGENTE DE PERCEPCIÓN de esta manera aplicará el cálculo.
8.2.2 Procedimiento Asignación de Tributos
Existen Tres formas de Asignación de Tributos a los Recursos, son los Siguientes:
8.2.2.1 Asignación desde el Catálogo de Recursos (Propiedades del Recurso).
Ingresar al Catálogo de Recursos, Clic derecho Modificar; Ubicar la pestaña de COTABILIDAD S10, Activar con el Check si es Detracción o Percepción y Clic el botón que mostrará el catálogo de Ítems de Detracción o Percepción según la activación del check.
a. Caso Configuración –
Detracciones Recurso por Recurso.
** El Cálculo de la Detracción aplicará al ser utilizado el recurso configurado y siempre y cuando supere los s/. 700. 00 según la configuración de los Ítems de Detracción en el catálogo y el Importe del Comprobante.
b. Caso Configuración –
Percepción Recurso por Recurso.
** Para aplicar la Percepción el recurso debe estar configurado con los Ítems de Percepción y el Socio Negocio debe estar calificado como Agente de Percepción de esta manera aplicará el porcentaje según el catálogo de Ítems de Percepción y según la Condición del Proveedor.
8.2.2.2 Configuración Masiva de Recursos
Configuración Masiva de Recursos relacionados a los Ítems de Percepción y Detracción, desde el escenario ASIGNACIÓN DE TRIBUTOS desde el Módulo de Facturación.
Desde el Escenario de Asignación de Tributos es posible configuración de manera masiva según la Selección de los recursos y Asignar Percepción o Detracción.
a. Caso Configuración Masiva de Ítems Detracción relacionado a los recursos.
Seleccionar los Recursos a Configurar, Clic Derecho_ Asignar_ Detracción.
Resultado de Asignación Masiva de Recursos con Detracción.
b. Caso Configuración Masiva de Ítems de Percepción relacionado a los recursos.
Seleccionar los Recursos a Configurar, Clic Derecho_ Asignar_ Percepción.
Resultado de Asignación Masiva de Recursos con Percepción.
8.2.2.3 Proceso de Importación Excel.
Ingresar al Módulo de Facturación, Escenario Asignación de Tributos con el Icono de Exportar a Microsoft Excel todos los recursos del Catálogos que se muestran en el escenario Asignación de Tributos.
Al Exportar al Excel aparecerá el listado de Recursos en los que se trabajará los Códigos de Ítems de Percepción y Detracción, relacionado los recursos para que se Importe Posteriormente.
En la Imagen se Observa que el Usuario ha trabajado los Ítems de Detracción relacionado a los Recursos en base a catálogo SUNAT de Ítems de Detracciones, de la misma manera se trabaja los recursos relacionados a Percepción.
Luego de Trabar la Hoja Excel, Guardar para iniciar con la Importación desde el escenario de Asignación de Tributos.
Los Campos Obligatorios son:
a. Código de Recurso: Para Identificar la configuración a realizar.
b. Código de Ítem Detracción: Indicar el Código del ítem a relacionar con el recurso.
c. Código de Ítem Percepción: Indicar el Código de ítem a relacionar con el recurso.
Aparecerá el explorador para realizar la Búsqueda del Archivo a Importar, Seleccionarlo con Doble clic o Clic en el botón Abrir.
Relacionar los Campos S10, con las cabeceras de la Hoja Excel de esta manera se relacionará la información a Importar.
• Seleccionar el Campo S10 Código Recurso y Relacionarlo con la cabecera de la Hoja Excel Columna A.
• Seleccionar el Campo S10 Código Percepción y Relacionarlo con la cabecera de la Hoja Excel Columna G.
• Seleccionar el Campo S10 Código Detracción y Relacionarlo con la cabecera de la Hoja Excel Columna D.
Seleccionar las Filas Del – Al. Posteriormente clic en el botón Importar para la carga de información seleccionada.
Proceso de Importación.
Resultados de la Importación.
