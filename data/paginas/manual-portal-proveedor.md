# Manual del Portal del Proveedor

Fuente: https://documentacion.s10peru.com/manual-portal-proveedor/

Table of Contents
Toggle
1. CALIDAD MOVIL PARA USUARIO PROVEEDOR
1.1 Servicio del S10
S10 Portal de Proveedores es una aplicación basada en web que permite consultar la información de pagos y Órdenes de Compra y Servicio registrados y aprobados desde el S10.
Este servicio está disponible las 24 horas del día y todos los días del año.
Disponibilidad de información en línea.
1.2 Acceso al Portal
El usuario proveedor que desee consultar o ver detalles de los pagos de su empresa deberá ingresar al portal “S10 Portal de Proveedores” en la dirección https://s10its.com/SupplierPortalWeb, para lo cual debe tener la autorización respectiva.
1.2.1 Autenticación del Usuario
El usuario debe ingresar su dirección de correo electrónico y su clave.
Si es que ingresa por primera vez, esta clave se encuentra en el correo electrónico de bienvenida, como el que se muestra a continuación:
Esta clave es temporal por lo que el portal le solicitará que ingrese una nueva clave.
Por cada pago que se le apruebe le llegará un correo como el siguiente:
1.2.2 Recuperación de contraseña
Si olvidó su clave, acceda a este enlace para que el portal le envíe una nueva clave a la dirección electrónica que ingrese.
Al revisar el buzón de correo, habrá llegado un correo enviado por el portal, donde se le asigna una nueva clave temporal.
Con esta clave podrá acceder al portal, y se le solicitará que ingrese una nueva clave.
En el campo Contraseña Anterior, ingrese la clave que se le entregó en el correo electrónico, e ingrese su nueva clave.
1.3. Servicios del Portal
1.3.1 Idioma
Seleccione el idioma de su preferencia. Este se guardará y la próxima vez que inicie sesión se mostrará en el idioma seleccionado. El portal está disponible en los idiomas español e inglés.
1.3.2 Facturación
Esta sección permite la búsqueda y consulta de las facturas.
1.3.2.1 Documentos Consolidado
Aquí se muestran todas las facturas para los cuales se ha programado o realizado algún pago.
Ingrese el rango de fechas a fin de filtrar los documentos. Si se deja en blanco estos campos se mostrarán todos los documentos.
1.3.2.2 Documentos pendientes de pago
Muestra todos aquellos documentos o facturas que se encuentran en estado Emitido o Parcialmente Pagado.
Haciendo clic en el enlace Ver Detalle se abrirá una ventana emergente que mostrará el detalle del documento.
Desde esta ventana es posible descargar el documento en diversos formatos.
1.3.3 Pagos
Son los pagos programados y realizados desde el S10 por las empresas compradoras.
1.3.4 Impresión de Comprobantes
1.3.4.1 Comprobante de Retención
Haciendo clic en el botón Ver Detalle se visualiza el detalle del comprobante. En esta ventana están las opciones para descargar el comprobante en varios formatos.
1.3.4.2 Comprobante de Detracción
Por cada documento pagado, que aplica detracción, se puede visualizar e imprimir la constancia de pago de detracción.
1.3.5 Órdenes de Compra y Servicio
1.3.5.1 Consolidado
Aquí se muestran todas las órdenes de compra y servicio aprobadas por las empresas compradoras.
Haciendo clic en el enlace Ver Detalle puede visualizar el detalle de la orden de compra o servicio e imprimirlo en el formato S10.
1.3.5.2 Pendientes de Atención
Aquí se muestran las órdenes de compra y servicio que se encuentran en estado Parcialmente facturado o Sin Facturación.
1.3.6 Métricas
Muestra de forma gráfica los importes facturados o pagados.
Seleccione el rango de tiempo por el desea visualizar la información.
1.3.7 Gestión de Usuarios
Administre a los usuarios desde está opción. Puede crear más usuarios para que les llegue las notificaciones por correo electrónico y visualizar toda la información de pagos desde el portal.
1.3.8 Configuración
Esta sección permite consultar y actualizar los datos de la empresa proveedora.
1.3.8.1 Mis Datos
Son los datos de la empresa que se registró desde el S10 ERP.
1.3.8.2 Persona de Contacto
Son los datos de persona o contacto principal de la empresa proveedora, el cual fue inscrito cuando se registró al proveedor en el portal.
1.3.8.3 Cuentas Bancarias
Son las cuentas bancarias a las cuales se ha realizado algún pago.
1.3.8.4 Logo de la Empresa
Seleccione el logotipo de la empresa proveedora, el cual se imprimirá en los reportes.
Portal de Proveedores
S10-ERP Especial
2. CALIDAD MOVIL PARA EMPRESA(S10 ERP Especial)
2.1. Configuración
2.1.1 Datos de la empresa
En la ficha de datos de la empresa, debe registrarse el monto mínimo de los pagos (en Soles) para los cuales se notificará a los proveedores.
Si el monto del pago es igual o mayor al monto mínimo, los posteriores pagos para el proveedor con cualquier valor serán notificados.
Si no se desea enviar las notificaciones a ningún proveedor, este dato debe tener valor cero.
Es necesario también registrar a un contacto de la empresa, el cual será inscrito en el Portal del Proveedores como Usuario Administrador.
2.1.2 Registro de Proveedores
Debe registrar el contacto con su respectivo email en la ficha de datos del proveedor, para que se le notifique que se le ha realizado un pago.
En la lista de contactos debe marcarse cuál de ellos será el contacto principal para el Portal de Proveedores.
De no tener ningún contacto registrado, se tomará el correo electrónico de la empresa (ficha General).
Cuando se envíe los documentos por pagar al Portal de Proveedores, se validará que se haya registrado los siguientes datos:
• Dirección electrónica del proveedor (e-mail) del contacto para el portal.
• Dirección del proveedor, siempre y cuando este indique como Domiciliado.
De no tener alguno de estos datos no será posible enviar las notificaciones.
2.1.3 Configuración en el módulo Administrativo
Enviar Ordenes de Pago al Portal de Proveedores:
Si esta activada, se notificará al proveedor cuando se apruebe una orden de pago cuyo importe sea según lo configurado en los datos de la empresa.
Permitir registro de facturas por pagar desde el Portal de Proveedores:
Si está activada, es necesario ingresar el tipo de documento y la ubicación de la base de datos, para que, desde el portal de proveedores, a partir de una orden de compra, el proveedor pueda generar la factura por el importe total.
( IMPORTANTE: Actualmente esta opción no esta implementada)
2.2 Envío de Notificaciones
2.2.1 Configuración inicial
Para los envíos de notificaciones de la empresa hacia sus proveedores, el sistema verifica los datos de la empresa y lo registra en el Portal de Proveedores.
Es necesario que la empresa tenga definido un contacto principal el cual será inscrito en el portal como usuario administrador de la empresa. Es obligatorio que el contacto tenga registrado una dirección de correo electrónico.
Se le enviará al usuario un correo electrónico de notificación, por ser la primera vez será un correo de bienvenida con los datos de acceso al portal.
Este correo es enviado por una cuenta denominada
S10 Alerta.
En este correo encontrará los datos de conexión al portal de proveedores. Utilice el navegador de Internet para acceder al portal.
2.2.2 Órdenes de Pago
Si la configuración tiene activada las notificaciones para las Ordenes de Pago, entonces se enviarán las notificaciones a los proveedores cuando se apruebe la Orden de Pago.
Se enviará al portal si el importe del documento o factura es mayor o igual al monto mínimo configurado en la ficha de datos de la empresa.
2.2.3 Pagos
Por cada proveedor, si el importe total del pago supera el monto mínimo definido en cada empresa, se enviará la notificación al proveedor al generar el pago.
Si el pago es anulado o eliminado, este será quitado también del portal.
2.2.4 Comprobantes
2.2.4.1 Detracción
Cuando se realice el pago del tributo de detracción, estos serán enviados al portal, para que el proveedor pueda consultar e imprimir la constancia.
2.2.4.2 Retención
Si el proveedor ya fue notificado por algún pago, todos los comprobantes de retención serán publicados en el portal.
Si el comprobante de retención es electrónico, se enviará al portal cuando sea aceptado por Sunat.
Los comprobantes que no son electrónicos serán enviados al en el momento que se realice el pago en el S10 ERP.
2.5 Órdenes de Compra / Servicio
Si el proveedor ya fue notificado por algún pago, cuando se apruebe alguna orden de compra o servicio, se enviará la notificación al proveedor.
