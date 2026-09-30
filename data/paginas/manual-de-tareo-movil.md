# Manual de Tareo Móvil

Fuente: https://documentacion.s10peru.com/manual-de-tareo-movil/

Table of Contents
Toggle
1. Manual Tareo Móvil para el ERP
1.1 Generalidades
Tareo móvil es un servicio en la nube que S10 brinda a los usuarios del S10 ERP edición Especial.
Esta aplicación móvil te permite registrar el tareo en un dispositivo móvil, Celular o Tablet, en las obras, para luego enviar la información a la oficina principal. Permite el registro de la información en forma conectada o desconectada.
1.2. Instalación y Configuración
Los datos de los usuarios que accederán a la aplicación móvil, deben ser proporcionados al área de ventas de S10 al momento de contratar el servicio. Debe indicar el correo electrónico del usuario y el RUC de la empresa asociada. También es necesario proporcionar el IP público de la base de datos así como el nombre de ésta.
Esta aplicación está disponible para Android y IOS. Busque la aplicación en Play Store o App Store según el dispositivo que tenga.
Valida tus Datos
Luego de descargar la aplicación es necesario validar sus datos para la activación, en donde colocará su RUC y dirección de correo electrónico.
Genera tu código de activación
Después de ingresar sus datos y hacer clic en el botón
Guardar
, entonces se habilitará un nuevo botón para generar el código de activación.
Recibe el código en tu bandeja
Se le enviará un mensaje a su correo electrónico con el código de activación.
Reingresa una nueva contraseña
Esté código debe ser registrado en el móvil, y luego se pedirá ingresar una nueva contraseña.
1.3 Sincronización de información
1.3.1 Publicación de Tareo
También es posible indicar si el proyecto multiproyecto será
tareable
o no, si está activado, el multiproyecto no se podrá tarear desde el móvil, en ese caso solo se podrá tarear en los proyectos destino.
La columna
Tareo con Alterno Trabajador
, si está activado, en el móvil se mostrará el nombre del trabajador antecedido de su código alterno.
Entonces si tiene activada la columna Tareo Móvil, cuando ingrese a la lista de partes de tareo, se habilitará la opción de publicación.
Se publica cada parte en forma independiente.
Antes de la publicación, el parte ya debe tener registrado la lista de trabajadores y un tareador asignado.
Existen dos formas de publicación: Nube o local. En la nube, es el tipo de publicación por defecto, significa que el S10 accederá a un servicio web de la nube para publicar los partes, este servicio se encuentra en un servidor de Amazon y es proporcionado por S10. De forma local es necesario que se instale un servicio web en el servidor donde se encuentra instalado el S10.
Al ingresar a la ventana de publicación por primera vez, en la pestaña
“
Por Publicar”
se muestra al tareador asignado al parte, para que tenga el acceso al tareo desde un dispositivo móvil. Haga clic en el campo
¿publicar?
y luego clic en el botón
Publicar
.
También permite realizar la publicación para invitados, sean o no usuarios del S10. Use la opción
Adicionar Invitado
, y este será habilitado como usuario en la aplicación móvil, pero solo para ver la información del tareo ya registrado.
En la pestaña Publicados, se muestran a todos aquellos usuarios a los que se ha otorgado el acceso al tareo en cuestión, desde un dispositivo móvil.
Todas las publicaciones en todos los proyectos pueden ser visualizadas en el escenario: Publicación Tareo Móvil.
Desde aquí podrá identificar los partes que fueron enviados al móvil por cada usuario y el estado en que se encuentran.
Aquí tenemos las siguientes opciones:
1.3.2 Edición de Tareo
La edición del tareo se realiza en el dispositivo móvil y luego mediante la opción Sincronizar    se actualiza esta información siempre y cuando se tenga conexión a Internet.
Para el tipo de publicación local esta información se actualiza directamente en la base de datos del S10 ERP.
Para el tipo de publicación en la nube es necesario ejecutar la opción
Actualizar desde el Móvil
seleccionando el parte.
1.3.3 Cierre del Tareo
Una vez que se haya terminado de registrar la información desde el dispositivo móvil, es necesario que se dé por terminado la edición del parte, para ello debe cambiar el estado en cada línea, a
Concluído
.
2. Manual Tareo Móvil para la App
2.1 Sincronizar el Tareo
Esta opción permite:
Obtener la información de tareo desde la base de datos del S10 ERP.
Enviar información registrada en el dispositivo móvil hacia la base de datos del S10 ERP.
La primera vez que sincroniza, se obtiene toda la información de los partes de tareo desde la base de datos del S10 ERP, y es guardada en el dispositivo móvil, esto comprende:
Trabajadores
Conceptos de horas
Proyectos
Actividades o partidas de control
Frentes
Destinos Específicos
Una vez que la información se encuentre registrada en el dispositivo móvil, y se haya ingresado datos de tareo, cuando se utilice esta opción, se verifica si se hizo alguna modificación en los datos, de ser así la información será actualizada en el S10 ERP.
2.2 Selección del Proyecto
Luego de sincronizar la información, se muestra en un listado todos los partes publicados para el usuario.
Seleccione el parte a editar seleccionando el icono
2.4 Resumen
Permite visualizar el consolidado del tareo de la semana.
Se muestra la lista de trabajadores y el resumen de horas de tareo por cada día de la semana.
Este icono permite expandir o contraer la información por trabajador.
Cuando este expandido
muestra una línea por cada trabajador y proyecto destino con el total de horas de todas las rotaciones. Cuando está contraído
se muestra una sola línea por cada trabajador.
Utilizar este icono para ordenar la información.
Etiquetas:
Australia Brasil
