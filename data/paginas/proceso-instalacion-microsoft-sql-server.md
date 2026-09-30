# El proceso de instalación de Microsoft SQL Server

Fuente: https://documentacion.s10peru.com/proceso-instalacion-microsoft-sql-server/

Este documento va dirigido a profesionales de tecnología interesados en entender el proceso de instalación de Microsoft SQL Server, sus componentes, las recomendaciones de seguridad y el significado de todas las opciones disponibles durante dicho proceso.
INTRODUCCIÓN AL DOCUMENTO
Si bien la instalación de Microsoft SQL Server no es complicada, es de gran importancia conocer acerca de lo que se está realizando en cada uno de los pasos del proceso, realizar una instalación a ciegas podría terminar en una instalación de más o menos servicios de los necesarios, en la implementación de malas prácticas de seguridad entre muchas otras.
Antes de Instalar
Es importante que antes de instalar SQL Server se tengan en cuenta algunas recomendaciones importantes:
1. Decidir la edición de SQL Server que se desea instalar
2. Revisar que se cumpla con los requerimientos de hardware y software necesarios para instalar SQL Server, la información al respecto se encuentra en la documentación del producto, al final de este documento se presenta el link hacia dicha documentación
3. Crear cuentas para los servicios de SQL Server, estas cuentas deben ser creadas con privilegios mínimos ya que, durante el proceso de instalación, el asistente les asignará los permisos necesarios para ejecutar los respectivos servicios.
La creación de estas cuentas de servicio NO es obligatoria para poder instalar SQL
Server, pero es una buena práctica de seguridad
Instalando
A continuación, se muestra una guía paso a paso de la instalación de SQL Server, con sus componentes de administración.
Paso 01:
“En SQL Server Installation Center” es posible revisar información detallada acerca de requerimientos para la instalación, recomendaciones de seguridad y adicionalmente realizar un chequeo de la configuración del sistema.
Haga clic en “System Configuration Checker”
Paso 02:
Revise el reporte y haga clic en OK
Paso 03:
Ahora, vaya al tab “Installation”, y allí seleccione la opción “New SQL Server stand alone installation or add features to an existing installation”
Paso 04:
Observe de nuevo el reporte y haga clic en“OK”
Paso 05:
Si está instalando una versión de pruebas (cómo en este ejemplo) de SQL Server, podrá seleccionar la opción correspondiente para la edición que desee; en una instalación diferente, agregue la clave de producto y haga clic en “Next”
Paso 06:
Ahora, lea los términos de licencia y luego, si está de acuerdo seleccione la opción correspondiente y haga clic en “Next”
Paso 07:
A continuación, se instalan componentes de soporte necesarios para la instalación, haga clic en “Install” para instalarlos
Paso 08:
Ahora haga clic en “Next”
Paso 09:
Ahora, deberá seleccionar las características de SQL server que desea instalar; Asegúrese de instalar Database Engine Services, y haga clic en “Next”
Paso 10:
A continuación tendrá que decidir si la instancia que va instalar es una instancia por defecto o nombrada, en el segundo caso tendrá que asignar a esta un nombre con el cual la reconocerá a futuro; si la instancia es creada por defecto, la forma de conectarse a esta desde servidores o equipos clientes remotos, será por medio del nombre de la máquina o de la dirección ip de la misma. Haga clic en “Next”
Paso 11:
En la siguiente ventana, se encuentra un análisis de requerimientos de espacio, cuando se haya comprobado que cuenta con el espacio de almacenamiento suficiente, haga clic en “Next”
Paso 12
:
Ahora, usted deberá configurar las cuentas con las cuales se ejecutará el servicio; la recomendación es utilizar diferentes cuentas, sin embargo, en la imagen de la derecha usted puede observar cómo una cuenta es utilizada para ejecutar más de un servicio, en la parte inferior podría seleccionar la opción para utilizar la misma cuenta para todos los servicios, en cuyo caso solamente tendrá que escribir credenciales una vez, pero no estará cumpliendo con buenas prácticas de seguridad. Después de configurar las cuentas, haga clic en el tab “Collation”
Paso 13:
En Collation, debe seleccionar “SQL_Latin1_General_CP1_CI_AS”, haga click en Next.
Paso 14:
Ahora, tendrá que definir el modelo de autenticación Mixto, deberá escribir una contraseña para el usuario administrador tipo SQL; Recuerde que el modo mixto permite la utilización de inicios de sesión tipo SQL (usuarios que no hacen parte de Windows) y es utilizada para dar acceso a SQL Server desde aplicaciones, entre otras cosas.
Paso 15:
Agregue también como administrador a cualquier usuario que vaya a cumplir con dicha tarea, por ejemplo el usuario que está ejecutando la instalación (Add current User) Haga clic en “Data Directories”
Ahora revise las ubicaciones físicas donde va a quedar instalado SQL Server y cada uno de sus componentes, Haga clic en Next
Paso 16:
Ahora, seleccione las opciones para que se envíen reportes de errores y de uso de características hacia Microsoft y haga clic en “Next”
Paso 17:
Haga clic en “Next”
Paso 18:
Revise el resumen y haga clic en “Install”
Paso 19:
La instalación está siendo realizada
Paso 20:
La instalación ha sido completada
Paso 21:
Si desea ver un resumen de la instalación, aquí encuentra un link hacia dicho registro de resumen; Haga clic en “Close” para salir, la instalación ha sido terminada.
La instalación de todas las ediciones y componentes de SQ L Server es similar al ejemplo que s e mostró anteriormente.
Tips para el proceso de instalación:
La instalación debe ser realizada con el perfil de usuario administrador de Windows.
Se recomienda según Microsoft no instalar en el mismo servidor de dominio.
Instalar SQL Server Management Studio (SSMS)
Ingresar a propiedades/Opciones de la BD, seleccionar el nivel de compatibilidad de acuerdo a su versión de SQL.
No se recomienda utilizar la versión 2016 por problemas de rendimiento
