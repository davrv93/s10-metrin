# Portal de Facturación Electrónica - Credenciales API REST SUNAT - Portal de Ayuda S10

Fuente: https://documentacion.s10peru.com/portal-de-facturacion-electronica-credenciales-api-rest-sunat/

MANUAL DE USO
Facturación Electrónica S10 - Credenciales API REST SUNAT
Contenido
Credenciales API REST SUNAT
Generando las credenciales
Registrando las credenciales API REST SUNAT en el Portal de Facturación Electrónica S10
Dar de baja guías de remisión electrónica (GRE)
Consulta de guías de remisión eletrónica (GRE)
DE QUÉ TRATA
1. Credenciales API REST SUNAT
La generación de los nuevos documentos electrónicos (Guías de remisión electrónica – GRE) deben realizarse utilizando los servicios propios de la SUNAT. Para habilitar esta funcionalidad y enlazarla con S10 ERP, es necesario crear sus credenciales API REST desde la web de «SUNAT Operaciones en línea».
Estas credenciales son un conjunto de passwords que validará la SUNAT para confirmar la titularidad de los documentos (GRE) que envíe desde S10 ERP.
CREDENCIALES
2. Generando las credenciales
Para generar las credenciales debemos ingresar a la pagina web de la SUNAT:
https://www.sunat.gob.pe/sol.html  y después ingresamos a la opción de «Mis Trámites y Consultas».
El portal de la SUNAT le solicitará autenticarse con su credenciales de clave SOL.
Después de ingresar al portal deberá ubicarse en la opción de «Personas» >  «Gestión Credenciales de API SUNAT»
Una vez en la opción deberemos registrar un formulario con unos datos de referencia:
Nombre de su aplicación:
S10 Facturación
URL de su aplicación:
https://www.s10peru.com/
En la opciones elegiremos a :
GRE Emision de Comprobantes /v1/contribuyente/gem
En el tipo de alcance seleccionaremos:
Desktop
Con estos datos elegidos procederemos a confirmar con el botón «Guardar». Posteriormente de procesada la información se nos mostrará el mismo formulario pero con 2 datos adicionales:
ID
y
CLAVE.
Estas son las credenciales de la API REST SUNAT que debemos conservar para realizar los envíos de documentos desde S10 ERP. Por favor copie los valores de ID y Clave para utilizarlos en el siguiente paso.
PORTAL DE FACTURACION ELECTRONICA S10
3. Registrando las credenciales API REST SUNAT en el Portal de Facturación Electrónica S10.
Para poder realizar los envíos de los documentos electrónicos (GRE) el Portal de Facturación Electrónica S10 debe conocerlos para utilizarlos cuando haga la conexión a los servicios de SUNAT, por ello debemos registrarlos en los datos de su empresa(s).
Accedemos al Portal de Facturación Electrónica con el
usuario administrador
de su empresa : https://s10net.com:18095/
Una vez dentro del portal, ingrese a la opción del menú:
Gestión de Empresa> Buscar Empresa
. Después seleccione la empresa la cual emitirá los documentos electrónicos y haciendo click derecho accedemos a la opción
Credenciales API SUNAT
.
https://s10net.com:18095/
En el formulario procederemos a registrar los datos y credenciales obtenidas previamente en la web de la SUNAT.
Usuario SOL: Usuario de la
cuenta principal
con el cual ingresamos a la web de SUNAT Operaciones en Línea .
Clave SOL: Clave SOL de la
cuenta principal
con el cual ingresamos a la web de SUNAT Operaciones en Línea.
Cred. API ID: ID de las credenciales API REST generadas.
Cred. API Clave: CLAVE de las credenciales API REST generadas.
Activo: Dejar seleccionada esta opción.
Una vez registrados los valores finalizamos la operación dando click en el botón
Guardar Credenciales
. Con esto S10 ERP ya esta listo y configurado para emitir los nuevos documentos electrónicos (GRE).
Validar
Credenciales SUNAT :
Con las credenciales ya guardadas en el portal, debemos verificar su validez conectándonos a los servicios API REST de SUNAT. Para este motivo hacemos click en el botón de validación. Si todos los datos ingresados son correctos obtendremos un mensaje de confirmación.
IMPORTANTE :
Antes de ejecutar la opción de «Validar Credenciales SUNAT», asegúrese de haber ejecutado exitosamente el botón de «Guardar Credenciales».
Si alguna de las credenciales son incorrectas obtendremos un mensaje de error de autenticación.
IMPORTANTE:
Es obligatorio que las credenciales que registremos estén validadas. Si el error de autenticación persiste
NO PODREMOS
ENVIAR
ningún documento electrónico relacionado a este servicio. Por ello es importante que se asegure que los valores registrados son los correctos y los datos de su clave SOL corresponden a la de su cuenta principal.
IMPORTANTE:
Si por algún motivo cambia los valores de Usuario SOL /Clave SOL / API Id / API Clave, deberá actualizar los nuevos valores en el Portal de Facturación Electrónica S10, ya que desde el momento que se modifiquen, SUNAT rechazará sus envíos hasta hacerlos con las credenciales actualizadas.
GUIA DE REMISIÓN ELECTRÓNICA
4. Dar de baja guías de remisión electrónica (GRE)
Si fuera necesario dar de baja alguna de las guías de remisión electrónica que emitamos, deberemos acceder a la pagina web  de la SUNAT:  https://www.sunat.gob.pe/sol.html
en la opción de «Mis Trámites y Consultas».
Después de ingresar al portal deberá ubicarse en la opción de «Empresas» >  «Baja de GRE»
Ingresamos en el formulario los datos de la guía de remisión que deseamos dar de baja.
Tipo de GRE: GRE – Remitente
Serie de la GRE: Serie del documento
Número de la GRE: Número del documento
Tipo de Baja: Seleccionar la opción correcta
Hacemos click en
Siguiente,
después de esto es posible que nos aparezca una alerta del portal donde se nos indique que la SUNAT puede fiscalizar este proceso. Damos click en
Entendido.
El portal nos mostrará un preliminar del documento a dar de baja y de estar todo conforme finalizamos haciendo click en la opción de
Dar de baja
De estar todo conforme el portal nos mostrará la aceptación del proceso. Desde aquí sera posible descargar una constancia de la baja en PDF o enviarla a un correo electrónico.
GUIA DE REMISIÓN ELECTRÓNICA
4. Consulta de guías de remisión electrónica (GRE)
Puede consultar en el portal web de SUNAT las guías de remisión electrónica que haya emitido. Para ello dirijase al menú de opciones especificado en el punto 4 de este documento y hacer click en «Empresas» >  «Consulta de GRE».
Elegimos el tipo de
GRE emitidas
y la búsqueda
Individual
Procedemos a ingresar los criterios de búsqueda
Tipo de GRE: GRE – Remitente
Serie: Serie del documento
Número: Número del documento
Hacemos click en
Siguiente
Los resultados de la búsqueda seran mostrados con la información relevante del documento, desde aquí será posible descargar el XML del documento o visualizar un PDF de la guía en el formato de SUNAT.
