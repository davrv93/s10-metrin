---
title: Manual de Almacenes
summary: El módulo de almacenes es el encargado de administrar los recursos que
  utiliza el proyecto con los movimientos de ingresos y egresos, proporciona
  información al módulo de Gerencia de proyectos para los resultados operativos
  del costo de los recursos tipo material. También…
tags:
  - oficial-s10
  - manual
  - portal-miembros
  - almacenes
id: 01M3SP28CQM6W3FPNGNT78XCQE
status: active
updated_by: ai-agent
updated_at: 2026-09-30T17:47:59.734Z
---

Manual oficial de S10, bajado del portal de miembros con la cuenta de la empresa.

- **Página:** https://documentacion.s10peru.com/manual-de-almacenes/
- **Módulo:** `almacenes` (ver esa rama para el resumen redactado)

## Contenido

- 1. ALMACENES DEL S10 ERP
- 1.1 Entorno de
- 1.1.1 Barra de títulos
- 1.1.2 Barra de menús
- 1.1.3 Barra de herramientas estándar
- 1.1.4 Barra de vistas
- 1.1.5 Árbol
- 1.3 Funcionamiento del módulo de almacenes
- 1.4 Acceso al módulo de Almacenes
- 1.5 Ingrese al módulo de Almacenes como sa
- 1.6 Tipo de documento de SUNAT
- 1.6.1 Tipo de Documento
- 2. Registro del proyecto
- 2.1.1 Definir techo al proyecto
- 2.1.2 Registro del aprobador de los pedidos en el almacén
- 2.1.3 Tipos de recurso que manejará el almacén
- 2.1.4 Usuarios por proyecto
- 2.2 Temas generales
- 2.2.1 Búsqueda global
- 2.2.2 Búsqueda local
- 2.2.3 Registro de recursos en el catálogo
- 2.4 Verifique el registro de los Equipos disponibles
- 2.5 Verifique el registro de las Unidades operativas y los Roles
- 2.6 Verifique el registro de la Ubicación física
- 2.7 Descripción de los escenarios de Ingresos y Egresos

## Texto de la página

# Manual de Almacenes

Fuente: https://documentacion.s10peru.com/manual-de-almacenes/

Table of Contents
Toggle
1. ALMACENES DEL S10 ERP
El módulo de almacenes es el encargado de administrar los recursos que utiliza el proyecto con los movimientos de ingresos y egresos, proporciona información al módulo de Gerencia de proyectos para los resultados operativos del costo de los recursos tipo material.
También controla los almacenes de servicios y activos.
Todos los ingresos se realizan para stock, Los egresos tienen un destino y determina el tipo de almacén.
El módulo de almacenes puede trabajar en forma independiente, sí solo desea controlar los movimientos de los recursos.
Si su empresa dispone de los módulos de Presupuestos, Gerencia de proyectos, Compras y pedidos, Administrativo, Contabilidad, Nóminas, formará parte del S10ERP, que integra la información técnica y financiera, donde el registro de información puede ser registrada desde cualquier módulo y es provechado por todo el sistema.
Un proyecto puede utilizar los siguientes tipos de almacenes:
De obra o de materiales: Usa los recursos tipo material, se caracteriza por los egresos que se realizan hacia una partida de control y en forma adicional se puede asignar también destino específico, frente y equipo; esta información es tomada por el módulo de Gerencia de proyectos para los resultados operativos.
De servicios: Administra solo los servicios, recursos tipo subcontrato o los no estoqueables o almacenables.
De activos: Usa los recursos tipo activo, los movimientos que realiza este almacén generalmente es para prestar las herramientas y equipos a una persona, la que debe devolver. No considera costos, es para controlar la posesión del activo.
Repuestos: Usa los recursos tipo material, los egresos se destinan hacia un activo, tiene la finalidad de controlar todos los consumos que este realiza.
Economato: Usa los recursos tipo material, los egresos se realizan hacia una unidad operativa (una oficina), permite conocer los consumos.
Una empresa puede administrar varios proyectos y cada proyecto puede utilizar varios almacenes, el sistema permite transferir recursos entre los almacenes del mismo proyecto, o transferir recursos de un proyecto a otro.
El registro de facturas relacionadas con los movimientos de ingreso hace que el proveedor no duplique el cobro, y que la información pase a documentos por pagar proporcionando información hacia los módulos administrativos.
1.1 Entorno de
trabajo
del S10
1.1.1 Barra de títulos
Muestra el título de la aplicación, nombre del programa que se viene utilizando, el proyecto elegido por el usuario, así como los botones minimizar , maximizar y cerrar .
1.1.2 Barra de menús
Presenta opciones de menú para la aplicación en uso. Cada menú contiene acciones específicas y estas cambian de acuerdo al escenario que ingresa el usuario, las mismas opciones en forma abreviada se ejecutan desde los botones (iconos).
1.1.2.1 Archivo
Contiene las opciones,
• Grabar: Graba los datos de la ventana activa.
• Imprimir: Imprime el contenido del escenario.
• Vista preliminar: Muestra la apariencia en pantalla de un archivo como será impreso.
• Iniciar sesión como usuario distinto: Permite reiniciar la aplicación sin salir del módulo de Almacenes, para ingresar como otro usuario. La misma acción se logra sí hace doble clic sobre la carita mostrada en esta barra de estado. El sistema mostrará el cuadro de diálogo de acceso al sistema, donde puede acceder al sistema como otro usuario también permite el uso de otra base de datos.
• Salir: Sale de la aplicación.
1.1.2.2 Catálogos
Permite el acceso a cualquiera de los catálogos utilizados por el sistema donde puede dar mantenimiento como efectuar nuevos registros, modificar o borrar siempre y cuando el registro no esté relacionado en alguna aplicación; todas estas acciones se realizan sin necesidad de estar elaborando un movimiento, todos los catálogos son abiertos y puede acceder desde cualquier aplicación.
1.1.2.3 Herramientas
1.1.2.3.1 Definir estructuras
El S10 maneja dos tipos de catálogos.
• De estructura definida, ej.: el catálogo de recursos, donde el usuario no tiene acceso a modificar su estructura.
• De estructura abierta, donde el usuario debe definir los niveles, longitud, y colores de los niveles de registro, ej.: catálogo de partidas de control, destinos específicos.
1.1.2.3.2 Configuración de Almacenes
Muestra la ventana de configuración del sistema, es independiente para cada usuario. Las opciones elegidas son permanentes hasta que el mismo usuario lo vuelva a modificar.
General
Llamada a-1: Seleccionado, Sí realiza el egreso de un recurso y le asigna una partida de control, al realizar un nuevo egreso del mismo recurso, el sistema le propone la misma partida de control.
No seleccionado, el usuario debe asignar manualmente la partida de control al recurso que egresa.
Llamada a-2: Seleccionado, La primera vez que efectúe un ingreso en el formulario siguiere la fecha del día, al efectuar los siguientes movimientos de ingreso volverá a sugerir la última fecha que efectuó el movimiento, esta opción es útil cuando se regulariza información.
No seleccionado, sugiere la fecha del día.
Llamada a-3: Seleccionado, solicita el ingreso del número de guía de remisión.
No seleccionado, no es necesario registrar el número de guía de remisión.
Llamada a-4: Seleccionado, al efectuar un Egreso por Transferencia Interna, envía un correo electrónico al encargado del almacén destino para que se entere.
No seleccionado, no envía correo.
Impresión
Impresión b-1: Seleccionado, el sistema registra la fecha y hora de impresión en los reportes.
No seleccionado, no imprime la fecha y hora en los reportes.
Impresión b-2: Seleccionado, el sistema muestra la ventana que permite modificar los márgenes.
No seleccionado, el sistema imprime el reporte con los márgenes por defecto.
Impresión b-3: Seleccionado, solicita el número de la página.
No seleccionado, por defecto el sistema coloca a la primera hoja el número 1.
Impresión b-4:| Seleccionado, en el reporte de egreso elegido, adiciona el egreso con descarga directa.
No seleccionado, imprime sólo el egreso elegido.
Del Supervisor
Llamada c-1: Recostea o procesa el stock del almacén con la fecha de ingreso de la guía.
Llamada c-2: Recostea o procesa el stock del almacén con la fecha de documento de pago.
Llamada c-3: Recostea o procesa tomando en cuenta la creación de la guía ingrese/egreso.
Llamada c-4: Cuando está seleccionado, en cada movimiento de solicita el tipo de cambio.
Llamada c-5: En los movimientos de ITI e ETI, el sistema utiliza guías que son llenadas automáticamente al registrar el movimiento. Esta opción seleccionada permite el registro manual de recursos al efectuar el movimiento, haciendo que la iniciar el proceso muestre la opción Adicionar en el menú.
Llamada c-6: Seleccionado, obliga que el Socio de Negocio tenga RUC, para el registro de documentos relacionados con el pago.
Llamada c-7: Ingrese el número de dígitos que tiene el RUC en su país.
Llamada c-8: Seleccionado, replica la fecha del Ingreso en el egreso por Transferencia Externa.
Llamada c-9: Seleccionado, muestra las órdenes de compra de todos los proyectos, considerando como consulta. RECOMENDADO, no seleccionado, para visualizar sólo las órdenes de compra que corresponden al proyecto.
1.1.2.3.3 Configuración Pedidos
Configuración General de Pedidos
Llamada 1-a: Seleccionado: Una vez que se ingresa datos en una celda y se presiona ENTER, el cursor se traslada a la siguiente celda que falta ingresar datos.
Llamada 1-b: Seleccionado: Luego de generar un pedido debe asignar obligatoriamente la fecha de entrega, se refiere a que el proveedor conozca las fechas de entrega de los recursos al almacén del proyecto; estas fechas son mostradas en la orden de compra, y también son utilizadas en el escenario del Cronograma de entregas detallado del grupo de Informes, para filtrar por fecha.
No seleccionado: Luego de generar un pedido permite grabar, sin necesidad de asignar fecha de entrega a los recursos.
Llamada 1-c: Seleccionado envía una notificación por correo, a la persona que debe aprobar el pedido.
Llamada 1-d: Seleccionado al generar un pedido sugiere la cantidad mínima de compra. En el catálogo de recursos se registra la cantidad mínima de compra.
Llamada 1-e: Seleccionado al generar un pedido avisa si alguno de los recursos están en otro no atendido.
Llamada 1-f: Al generar un pedido envía un mensaje con la configuración elegida.
Configuración de Impresión en Pedidos
Llamada 2-a: Seleccionado, Pone fecha de impresión al reporte impreso.
Llamada 2-b: Seleccionado, Permite modificar los márgenes al imprimir el reporte.
Llamada 2-c: Seleccionado, Permite registrar el número de página para al imprimir el reporte.
Llamada 2-d: Seleccionado, Permite imprimir el reporte el peso de recurso.
Configuración Del Supervisor en Pedidos
Llamada 3-a: Seleccionado, Solicita fecha de entrega de parte del proveedor al almacén de obra, estas fechas se visualizan en el reporte de la orden de compra y sirve para filtrar información por fechas en el grupo de Informes el escenario de Cronograma de Entregas Por Rangos.
Llamada 3-b: Seleccionado, sí al momento de registrar un pedido A Compra se graba, el sistema solicita ingresar la Partida De Control, este dato solo es informativo.
Llamada 3-c: Es el número de dígitos considerado para PEDIDOS, COTIZACIONES y ORDENES DE COMPRA – se visualiza en número de guía.
Configuración en Facturación
Al registrar en el escenario de Documentos por pagar no solicita centro de costo, glosa.
1.1.2.3.4 Correo interno
Activa el correo interno del sistema para enviar o recibir mensajes de un terminal a otro. Está disponible cuando se trabaja en red y en la misma base de datos.
1.1.2.3.5 Calculadora
Muestra la calculadora de Windows.
1.1.2.3.6 Calendario
Muestra el calendario de Windows.
1.1.2.3.7 Procesos especiales
Limpiar registro de control: Por muchos motivos el sistema puede fallar, estos quedan grabados en el editor de registros; esta opción borra, permitiendo el acceso al escenario donde viene trabajando el usuario.
Asignar propiedad IGV a los recursos: En el catálogo de recursos tenemos una celda, que cuando está seleccionado permite indicarle que ese recurso está afecto del IGV. Esta opción hace que todos los recursos estén afectos del IGV.
Resetear agrupaciones: El escenario donde se efectúan los movimientos, está compuesto con columnas y las “cabeceras” pueden ser arrastradas formando filtros que permiten agrupar por ese concepto. Este utilitario deshace o resetea la selección
Eliminar todos los filtros: Es un utilitario que elimina todos los filtros activos.
Exportar plan contable a archivo texto: El plan contable es enviado a un archivo de texto.
Verificar stock reservado en los pedidos: Al efectuar pedidos del stock con destino a una partida de control, la cantidad es reservada, esto hace que al mostrar el stock no sea real. El utilitario restaura, deshaciendo el pedido.
1.1.2.3.8 Movimientos de almacenes
Este utilitario traslada los movimientos que se efectúan en la obra a la base de datos de la oficina.
Se puede dar el caso de trabajar con una sola base de datos centralizados, que se ubique en la oficina principal, y tener varias obras que no están conectadas.
Requisito de software
• Tener instalado el módulo de almacenes tanto en la oficina principal como en las obras.
El procedimiento para el manejo de información:
• En la oficina principal donde se encuentra la base de datos, preparar toda la información previamente, como es el presupuesto, la planificación.
• Luego sacar una copia de seguridad de la base de datos principal ubicado en la oficina y trasladar a la PC de la obra.
Limitaciones
• En el almacén de obra sólo se efectúan movimientos.
• En caso de que en una obra, necesiten registrar un nuevo recurso, el almacenero debe comunicar previamente a la oficina principal, de su requerimiento. En la oficina principal deben proceder a registrar el recurso y dictar el código a la obra, para que el almacenero registre en el sistema el recurso con el mismo código.
Exportación de movimientos
• Disponible para la oficina principal como para la obra, permite exportar sólo los movimientos y grabarlos en un archivo.
Importación de movimientos
• Disponible para la oficina principal como para la obra, permite importar sólo los movimientos previamente grabados en un archivo.
1.1.3 Barra de herramientas estándar
Presenta iconos que permiten simplificar las operaciones más frecuentes dentro del S10 y las funciones son:
Envía el reporte directamente a la impresora previamente definida en Windows.
Muestra la apariencia de la ventana al imprimirlo, o el reporte de la ventana activa en pantalla de donde se puede exportar al formato que elija el usuario como Word, Excel, etc.
Muestra o esconde el árbol ocupando su lugar por el escenario actual o activo.
1.1.4 Barra de vistas
Esta barra contiene las etiquetas con los nombres de los grupos, estos a su vez contienen los botones de acceso directo a los escenarios. Los grupos que contiene la barra de vistas varían de acuerdo a la personalización, ya sea por empresa o entidad.
1.1.5 Árbol
Escritorio: Contiene sólo los proyectos que se vienen utilizando en alguna aplicación o está en proceso de ejecución. Solo para los proyectos ubicados en esta carpeta están activas todas las opciones de edición. Use clic derecho para ver las opciones del menú.
Proyectos ejecutados: Los proyectos trasladados a esta carpeta no pueden ser modificados por el usuario.
Bandeja: Esta carpeta almacena todos los proyectos que no están en uso y en cualquier momento se pueden trasladar a la carpeta “Escritorio”.
Archivo Central: Esta carpeta almacena los proyectos que ya no serán utilizados.
Papelera de Reciclaje: Todos los proyectos eliminados o borrados son almacenados en la papelera de reciclaje, de donde pueden ser restaurados o borrados en forma definitiva. Existe la opción de vaciar toda la papelera.
1.2 Diagrama de bloques del funcionamiento del almacén con relación al proyecto
Llamada a: Todos los ingresos se relación con la factura y la información pasa a documentos por pagar y por consiguiente a contabilidad.
Los informes varían de acuerdo al tipo de almacén.
En el almacén de materiales utilice solo recursos tipo material.
En el almacén de Servicios, utilice solo recursos tipo subcontrato y equipo
1.3 Funcionamiento del módulo de almacenes
1.4 Acceso al módulo de Almacenes
Del tapiz de su PC elija el botón de acceso directo
1.5 Ingrese al módulo de Almacenes como sa
El (sa) es el administrador del sistema
Sí el módulo de almacenes es utilizado en un lugar diferente al Perú, ingrese a la opción Configuración por País, para registrar la etiquetas que acostumbra usar.
1.6 Tipo de documento de SUNAT
La SUNAT (Superintendencia de Administración Tributaria), en el caso del Perú ha normado el uso de los documentos contables y estos deben ser registrados en el sistema, el encargado de esta labor es el administrador del sistema (sa).
Registro del Tipo de Documento en el Catálogo
Para el caso del Perú, el Catálogo de Tipos de Documentos de SUNAT, ya viene registrado.
1.6.1 Tipo de Documento
Son los documentos que acostumbra usar la empresa, deben ser registrados en el sistema y compatibilizados con los documentos normados por SUNAT. Estos son algunos documentos que usa la empresa.
Para registrar nuevos documentos,
Llamada a: Ingrese el código alterno del tipo de documento registrado
Llamada b: Ingrese el nombre del documento
Llamada c: Ingrese la abreviatura del documento que registra.
Llamada d: Del catálogo elija si es un documento para pagar, cobrar o ambos.
Llamada e: Elija el comportamiento del Tipo de Documento.
Llamada f: Elija el tipo base, para que el sistema lo reconozca en este caso como factura.
Llamada g: Elija el tipo de moneda para el documento.
Llamada h: Cuando tiene selección la celda, genera el asiento contable por recurso, caso contrario por documento.
Llamada i: Registre el tipo de impuesto para el documento.
Para registrar el Impuesto, haga clic derecho en el segmento marcado con la llamada i, luego en el Catálogo de impuestos:
Llamada de la ventana Impuestos:
Llamada i-1: Ingrese el nombre del impuesto.
Llamada i-2: Registre la abreviatura del impuesto.
Llamada i-3: Elija signo positivo afecta la base imponible y cuando es negativo no altera ningún importe.
Llamada i-4: Ingrese la tasa del impuesto en porcentaje.
Llamada i-5: Obligatorio, aplica siempre el puesto al documento.
Llamada i-6: Seleccione cuando el impuesto esté en vigencia.
Llamada i-7: Ingrese una observación relacionada con el impuesto.
Llamada i-8: Este dato es informativo, es la fecha de vigencia del impuesto.
Llamada i-9: Use el botón para grabar los datos y cerrar la ventana.
Ejemplo de registro de documento RECIBO INTERNO, que será utilizado en Fondo Rotatorio.
Cierre el catálogo de Tipo de Documentos
Registre al usuario que manejará el módulo de almacenes y a los aprobadores, asigne atributos de uso y permisos.
Ingrese al módulo como usuario
2. Registro del proyecto
Sí en el módulo de Gerencia de proyectos, ya se efectuó el registro, la información es visualizada en el módulo de Almacenes.
Sí el módulo de almacenes es utilizado en forma independiente, o este módulo es utilizado primero; sí, es necesario registrar el proyecto. Si el registro lo efectúa el usuario (Andrés) el sistema por defecto le proporciona los accesos al proyecto y sus almacenes. Luego registre las Partidas de Control y puede iniciar el registro de los ingresos y egresos.
Utilice estos datos para registrar el proyecto:
Acceso al escenario de Datos Generales: El proyecto y sus datos complementarios se registran en el escenario de Datos Generales.
Registro del grupo en el Catálogo de Proyectos
Registre: Año 2010,
Dentro del grupo Año 2010, registre los datos del proyecto;
Nota: Es recomendable registrar los años en el árbol, para tener en cuenta los proyectos que se ejecutan cada año.
Ingreso de datos del proyecto
Llamada a: Registre el nombre del propietario del proyecto.
Llamada b: Elija el lugar donde se ejecutará el proyecto.
Llamada c: Ingrese la fecha de inicio del proyecto, este dato es importante, ya que a partir de este el sistema genera períodos semanales y mensuales. NO MODIFIQUE ESTA FECHA SI YA TIENE MOVIMIENTOS EN EL ALMACEN.
Llamada d: No ingrese este dato, es informativo, el sistema lo calcula el sistema automáticamente.
Llamada e: Es el lapso de tiempo en días calendario que está previsto dure la ejecución del proyecto, es un dato tentativo inicial y sirve para generar los períodos a partir de la fecha de inicio.
Llamada f: Hay muchas empresas que inician la semana para efectos del pago de planillas el día lunes, otros el miércoles. Cuando el sistema genere los períodos semanales, la semana iniciará el día que eligió. Uniformice, se recomienda que empiece el día lunes en todos los proyectos.
Llamada g: No disponible en este módulo.
Llamada h: Elija el tipo mismo tipo de moneda que utilizó en el presupuesto.
Llamada i: Elija el tipo de moneda alterno que utilizará el módulo de almacenes.
Llamada j: Ingrese el nombre de empresa que ejecutará el proyecto.
Llamada k: Sí la empresa tiene sucursales elija la que corresponde. En el catálogo de identificadores se registra la empresa y las sucursales.
Llamada L: Este dato es informativo, ingrese la dirección del proyecto.
Llamada m: Este dato es informativo, ingrese el número telefónico del local donde funcionan las oficinas del proyecto.
Llamada n: Este dato es informativo, ingrese el número de fax del local donde funcionan las oficinas del proyecto.
Llamada o: Después de registrar todos los datos use este botón para grabarlos y cerrar la ventana y el proyecto será mostrado en el catálogo de proyectos.
Luego,
2.1.1 Definir techo al proyecto
Llamada a: Seleccionado permite elegir el tipo de techo que usará el proyecto en los pedidos.
No seleccionado, al pedir muestra el catálogo de recursos.
Llamada a-1: Seleccionado, al efectuar un pedido solo muestra los recursos de los presupuestos asignados al proyecto, en el escenario de Datos Generales del módulo de Gerencia de proyectos se asigna presupuestos al proyecto.
Llamada a-2: Seleccionado, los pedidos se efectuarán del Techo Libre, que es una especie de bolsa donde previamente se registran los recursos que manejará el proyecto. En el grupo de Pedidos, se encuentra el escenario Techo Libre.
Llamada b: Ingrese el número de serie y el número de recursos para la Guía de Remisión.
Llamada c: Permite configurar los decimales en cantidades, precios y parciales en las guías de ingreso y egreso.
Llamada d: Elija esta pestaña, para registrar a todos los usuarios del almacén.
Llamada d-1: Ponga una selección al o los responsables del manejo del almacén.
2.1.2 Registro del aprobador de los pedidos en el almacén
El usuario que apruebe los pedidos, debe estar registrado previamente en el Catálogo de Socio de Negocios y también debe estar registrado como usuario en el módulo de Almacenes, con sus permisos y accesos respectivos.
Datos para el registro
Acceso al escenario
2.1.3 Tipos de recurso que manejará el almacén
Permite definir o elegir el tipo de recurso que usará el almacén.
Llamada a:
Seleccionado, permite efectuar egresos por venta.
No seleccionado, inhabilita los egresos por venta.
Llamada b:
Muestra los grupos de recursos disponibles, cuando efectúe movimientos de ingreso sólo visualizará los recursos del grupo.
2.1.4 Usuarios por proyecto
Permite adicionar usuarios y el registro de atributos de uso dentro del proyecto, el usuario que tiene la categoría de Jefe de Grupo es el que asigna.
Llamada a:
Permite el registro de usuarios al proyecto y dar atributos de uso.
Llamada a-1:
El usuario que tenga seleccionado, puede asignar techo al proyecto.
Llamada a-2:
El usuario que tenga seleccionado, puede adicionar o eliminar almacenes del proyecto.
Llamada a-3:
Graba el contenido y cierra la ventana.
2.2 Temas generales
2.2.1 Búsqueda global
La búsqueda global consiste en digitar las primeras letras del registro que se quiere ubicar en los catálogos. Ubique el cursor en la columna donde se quiere buscar información.
Llamada a: Esta celda muestra las letras o números que se va ingresando al momento de efectuar la búsqueda.
Llamada a-1: Para ubicar información en el árbol, debe estar previamente desplegado y el cursor debe estar ubicado es ese segmento y sólo digite letras.
Llamada a-2: Para ubicar información en una columna que contiene números, el cursor debe estar ubicado es ese segmento y sólo digite números.
Llamada a-3: Para ubicar información en una columna que contiene letras, el cursor debe estar ubicado es ese segmento y sólo digite letras.
Para deshacer la búsqueda use las teclas .
2.2.2 Búsqueda local
La búsqueda local permite ubicar información en una columna usando la tecla F8 o usando el botón y en la ventana que muestre ingresar hasta tres palabras de la cadena que se quiere ubicar.
Ubique el cursor en una columna (cualquiera marcada con las llamadas a-1, a-2, a-3),
Luego use el botón
Llamada b: Ingrese información en cada una de las celdas, no importe el orden.
Llamada b-1: Efectúa la búsqueda.
Llamada a-4: Deshace la búsqueda.
2.2.3 Registro de recursos en el catálogo
Todos los recursos que utiliza el proyecto deben estar registrados en el catálogo de recursos necesariamente. Antes de registrar un recurso utilice las búsquedas, para no volver a registrar un recurso existente.
Datos para registrar
Ingrese a Catálogos, luego Recursos Genéricos y Específicos.
El registro se efectúa en la rama correspondiente.
El sistema envía un mensaje indicando el registro del recursos específico, elija sí.
Luego use el botón
2.3 Verifique el registro de Partidas de Control, Destinos específicos, Frentes
En el grupo Configuración verifique el registro de los catálogos de: Partidas de Control, Destinos Específicos y Frentes, que son necesarios para asignar como destino a los recursos que egresan del almacén. En el caso de que no existan registros solicite al encargado de Gerencia de proyectos que los registre.
2.4 Verifique el registro de los Equipos disponibles
En el grupo Equipo, escenario de Equipos disponibles, verifique que estén registrados los equipos que viene utilizando el proyecto, que son necesarios para asignar como destino a los recursos que egresan del almacén. En el caso de que no existan registros solicite al encargado de Gerencia de proyectos que los registre.
2.5 Verifique el registro de las Unidades operativas y los Roles
En el grupo Configuración verifique que estén registradas las Unidades Operativas y los roles, que son necesarios para asignar como destino a los recursos que egresan del almacén tipo economato. En el caso de que no existan registros solicite al encargado de Gerencia de proyectos que los registre.
2.6 Verifique el registro de la Ubicación física
En el grupo Configuración verifique que esté registrado la ubicación física que será asignado a los recursos en el escenario de Stock, esto permite que cada recurso tenga un lugar dentro del almacén. En el caso de que no existan registros solicite al encargado de Gerencia de proyectos que los registre.
2.7 Descripción de los escenarios de Ingresos y Egresos
Partes del escenario
Cabecera de los escenarios de Ingresos y Egresos
Llamada a: Muestra los tipos de movimientos disponibles de ingresos, del que debe elegir uno de ellos.
Llamada 1: En estos movimientos muestra una guía del movimiento anterior involucrado, caso contrario permite elegir del catálogo los recursos que forman parte del ingreso, es configurable.
Llamada a-1: Muestra los tipos de movimientos disponibles de egresos, del que debe elegir uno de ellos.
Llamada b: Seleccionado, muestra todos los movimientos, es necesario usar la opción macada con la llamada e.
No seleccionado, permite elegir un rango de fechas (llamadas c, d) para luego usar el botón marcado con la llamada “e”.
Llamada c: Permite elegir la fecha inicio del rango de fechas, desde ese fecha muestra la información en la ventana marcada con la llamada “o”.
Llamada d: Es requisito usar previamente la opción de la llamada c, permite elegir la fecha fin del rango de fechas, hasta ese fecha muestra la información en la ventana marcada con la llamada “o”.
Llamada e: Para el rengo de fechas elegido, refresca o muestra la información, en la ventana marcada con la llamada “o”.
Lista de movimientos
Ingresos
Egresos
Llamada f: Muestra la ventana Filtro Dinámico, para mostrar información en el segmento marcado con la llamada o.
Llamada f-1: Graba la selección efectuada con la llamada f-4 y f-5
Llamada f-2: Permite registrar un nombre al filtro y lo graba.
Llamada f-3: Muestra la lista de Filtros grabado, con la finalidad de abrir la selección.
Llamada f-4: Para filtrar información elija un grupo.
Llamada f-5: Del grupo abierto elija el detalle que será mostrado en el segmento marcado con la llamada “o”, luego es necesario cerrar la ventana de filtro, para deshacer un filtro invoque nuevamente la opción y quite la selección.
Llamada g:
Cuadro agrupar por:
Cuando se agrupa la cabecera de las columnas marcadas con la llamada i-1; los botones marcados con las llamadas m, n graban esa configuración y el botón marcado con la llamada L muestra el agrupamiento grabado.
Llamada h:
Mostrar filtro avanzado: Muestra o esconde las celdas del filtro avanzado. Este utilitario permite ubicar información contenida en el formulario de ingreso, de acuerdo al campo elegido.
Llamada i:
Mostrar campos, muestra la ventana con los campos (columnas) disponibles del escenario con la finalidad de mostrarlas u ocultarlas, demás cambiar de posición.
Llamada j:
Exporta el contenido del formulario a un archivo de Microsoft Excel.
Llamada k:
Cursor automático, cuando está seleccionado el cursor se ubica en la celda que falta ingresar información, caso contrario es necesario usar el puntero del ratón para activar la celda.
Llamada L:
Muestra los grupos de filtro previamente almacenado con los botones de las llamas m, n con la finalidad de abrirlo.
Llamada m:
Graba el agrupamiento de la cabecera de las columnas.
Llamada n:
Graba el agrupamiento de la cabecera de las columnas, el usuario debe asignar un nombre al agrupamiento o aceptar el que siguiere el sistema.
Llamada o-1:
Muestra la lista de ingresos efectuada al almacén. El detalle del formulario de ingreso, tiene las siguientes opciones en el menú emergente, al invocarlo con clic derecho. Este menú cambia de acuerdo al tipo de ingreso elegido
Llamada o-1-1:
Permite el registro de un nuevo movimiento.
Llamada o-1-2:
Muestra el formulario de ingreso para modificar información, si tiene regularización elimine previamente.
Llamada o-1-3:
Elimina el ingreso, siempre y cuando el usuario tenga atributos para hacerlo y no tenga regularización (debe estar registrado en estado registrado). En el escenario de datos Generales se asigna el atributo.
Llamada o-1-4:
Modifica el estado de la guía de ingreso,
Llamada o-4-1:
Este estado permite que la guía sea modificada.
Llamada o-4-2:
Luego de revisar el movimiento, cambie a este estado para que no se modifique accidentalmente la guía de ingreso.
Llamada o-4-2:
La guía cambia a estado protegido cuando cierra el mes.
Llamada o-1-5:
Permite el registro de notas a la guía.
Llamada o-1-6:
Permite efectuar la auditoría a la guía, los resultados de auditorías son visualizadas en el grupo de RESULTADOS, escenario de RESULTADOS DE AUDITORIAS.
Llamada o-1-6-1:
Elija la pestaña Acciones y Resultados que contiene las etiquetas para calificar a la guía.
Llamada o-1-6-2:
Al hacer clic derecho muestra el menú de donde debe elegir ADICIONAR, que graba la fecha y el nombre del usuario que efectuará la auditoría.
Llamada o-1-6-3:
Contiene las etiquetas con las que se califica a las guías. El usuario debe registrarlas previamente.
Llamada o-1-6-4:
Muestra las etiquetas elegidas para guía.
Llamada o-1-6-5:
Muestra la ventana donde permite el registro de notas con relación a la guía calificada. Después de usar la ventana, cierre.
Llamada o-1-7:
Imprime en pantalla la apariencia del reporte de la guía seleccionada.
Llamada o-1-8:
Envía el reporte a la impresora previamente seleccionada en Windows.
Llamada o-1-9:
Muestra la ventana de seguimiento o la trazabilidad de la guía.
Llamada o-2:
Muestra la lista de guías de egresos.
Llamada o-2-1:
Adiciona un nuevo movimiento de egreso.
Llamada o-2-2:
Modifica los datos de un movimiento de egreso previamente registrado.
Llamada o-2-3:
Elimina un movimiento de egresos existente, el estado debe estar en edición.
Llamada o-2-4:
Genera la guía de Remisión de un movimiento de egreso.
Llamada o-2-5:
Modifica el estado de la guía.
Llamada o-2-5-1:
Edición permite modificar los datos de la guía o eliminar.
Llamada o-2-5-2:
Revisado, es un estado que solo es informativo.
Llamada o-2-5-3:
Protegido evita que la guía sea modificada o eliminada.
Llamada o-2-5-4:
Anulado, mantiene la numeración de la guía, y la cantidad de los recursos se ponen en cero, regresa al stock las cantidades egresas.
Llamada o-2-6:
Auditoría, ver Llamada o-1-6.
Llamada o-2-7:
Muestra la ventana donde permite el registro de notas con relación a la guía calificada.
Llamada o-2-8:
Imprime en pantalla la apariencia del reporte de la guía seleccionada.
Llamada o-2-9:
Envía el reporte a la impresora previamente seleccionada en Windows.
Llamada p:
Muestra el contenido o los recursos de la guía.
Al hacer clic derecho permite el recosteo, para los ingresos y egresos es igual.
2.7.1 Tipo de cambio
Disponible solo para los países que usan doble moneda.
Verifique la configuración con el botón
Sí el sistema está configurado para el registro del tipo de cambio, al efectuar un movimiento muestra el siguiente mensaje:
Los movimientos de ingreso y egreso son mostrados y se invocan desde un escenario, para registrar información usan formularios.
2.1 Esquema abreviado para efectuar movimientos en el almacén
2
2.1 Ingresos
Todo recurso debe ser ingresado al almacén y debe estar respaldado por un documento en el que figure la fecha, cantidad, unidad y precio del recurso.
Solo los recursos del stock pueden ser utilizados en los movimientos de egreso.
Todo ingreso va hacia al stock del almacén, salvo aquel recurso que no se almacena como es el caso del concreto premezclado que apenas llega a la obra es utilizado, entonces el sistema genera el formulario de ingreso y luego mediante DESCARGA DIRECTA, genera el formulario de egreso.
2.1.1 Movimientos de ingreso que no requieren de un movimiento previo
2.1.1.1 Procedimiento o pasos a seguir para efectuar un ingreso – Ingreso por Inventario [II]
Los formularios que se utilizan para los ingresos, son los mismos para cualquier ingreso, al hacer ingreso por inventario se explica el procedimiento o pasos a seguir para efectuar un ingreso, se describe la funcionalidad del formulario que se utiliza para realizar el ingreso.
Utilice el ingreso por inventario cuando empiece a controlar la información de un proyecto ya iniciado y todos los recursos que encuentre en el almacén deben ser registrados en el almacén para tener stock y posteriormente hacer los egresos.
Acceso al escenario de ingresos
Llamada a: Solo para emitir los reportes a nivel de proyecto, en el árbol ubique el cursor sobre el nombre del proyecto.
Llamada b: Los movimientos de ingreso solo se efectúan para el almacén elegido.
Datos para registrar.
Elección del tipo de ingreso
Formulario de registro del ingreso
Todos los movimientos de ingresos usan el mismo formulario, cada tipo de ingreso genera un número de guía correlativo y el sistema propone el tipo de documento en algunos tipos de ingreso y en otros casos el usuario debe elegir un documento.
Registro de datos en la cabecera del formulario de ingreso
Llamada 1a:
En esta celda ingrese la fecha del movimiento, el día que el recurso ingresa al almacén.
Llamada 1b:
Elija el tipo de documentos, que respalda el movimiento de ingreso.
Llamada 1c:
Ingrese el número del tipo de documento elegido en la llamada 1b.
Llamada 1d:
Ingrese la fecha del documento elegido en la llamada 1b.
Llamada 1e:
Elija el tipo de moneda que utiliza en el movimiento.
Llamada 1f:
Esta ventana muestra el valor del tipo de cambio. Cuando eligió la fecha de ingreso (llamada 1a) el sistema solicitó el ingreso del tipo de cambio, para editarlo o cambiarlo use el botón
que permite acceder al catálogo de tipos de cambio.
Llamada 1g:
Es el porcentaje de IGV (impuesto general a las ventas), asignado al documento que se viene registrando, este valor es almacenado por documento.
Para cambiar el valor del IGV, ingrese al módulo de Almacenes como (sa) luego use el botón
que aparece en la barra de botones.
Llamada 2a: Barra de botones del cuerpo del documento.
Llamada 2b: Recursos seleccionados que forman parte del ingreso, se eligen del catálogo al invocarlo con clic derecho.
Llamada 2c: Muestra la unidad del recurso.
Llamada 2d: Ingrese la cantidad que ingresa.
Llamada 2e: Ingrese el precio neto sin impuesto.
Llamada 3a: Muestra el nombre completo del recurso.
Llamada 3b: Sí previamente asigno ubicación al recurso dentro del almacén. Esta opción es informativa.
Llamada 3c: Permite el ingreso de observaciones al documento de ingreso. este dato es mostrado en el escenario en una columna; registre descripción corta para agrupar.
Llamada 3d: Esta celda es informativa, muestra el precio parcial del documento.
Llamada 3e: Muestra los montos calculados del documento. Sí no coincide con el documento que registra, devuelva al proveedor para que emita el documento con cálculos correctos.
Llamada 3f: Graba el contenido de la guía.
Llamada 3g: Graba el contenido de la guía y generar el documento de egreso.
Llamada 3h: Envía a la impresora el reporte de la guía de ingreso.
Llamada 3i: Utilice esta opción solo cuando registre recibos de servicios de agua, telefonía, fluido eléctrico, con la finalidad de repartir el costo a varios proyectos
Llamada 3j: Deja sin efecto el registro del ingreso.
2.1.2 Impresión de reportes
Use el botón
para envía el reporte a la impresora previamente elegida en Windows.
Use el botón
para mostrar una vista previa o impresión en pantalla de la forma como será impreso el reporte.
RECOMENDACIÓN: Utilice un ingreso por cada familia de recursos.
2.1.2.1 Ingreso por Compra [IC]
Este movimiento se efectúa cuando los recursos que ingresan al almacén vienen respaldados por una guía de remisión, factura, boleta de venta o ticket. Se supone que no tiene orden de compra.
2.1.2.1.1 Ingreso con Guía de Remisión
Cuando al almacén del proyecto llegan recursos sólo CON GUIAS DE REMISION y en reiteradas veces. Luego se regulariza con documento de pago, la factura.
Este caso se da en entregas parciales de madera, agregados, concreto premezclado y luego nos envían la factura para regularizar el precio. El registro de la factura al regularizar hace que este documento pase a documentos por pagar y a contabilidad.
Para este ejemplo se utiliza CONCRETO PREMEZCLADO, como es un recursos no estoqueable debe USAR DESCARGA DIRECTA para grabar la guía de ingreso y generar la de egreso inmediatamente, donde se asigna Partida de Control (CONCRETO) + Destino específico (CIMENTACION).
Datos para registrar:
Acceso al escenario
Elige el tipo de ingreso:
Elija la opción marcada con la llamada a, en el formulario de Ingreso por Compra registre los datos en la cabecera, luego;
Elija el recurso,
En el formulario de Egreso a partida de Control con Descarga Directa,
Llamada a: Se genera en forma automática.
Llamada b: Elija a la persona encargada del uso del recurso.
Llamada c: Es la fecha del movimiento de egreso, el que no debe ser modificado, la configuración relacionada con este movimiento se encuentra en la pestaña general y se marca con una flecha.
Llamada d:
Esta celda es editable para registrar un número relacionado al movimiento que puede ser de un talonario manejado por el maestro del obra.
Llamada e:
Autorizado por, muestra la lista de las personas que tienen permiso para autorizar egresos.
Es requisito ubicarse en el escenario de Datos Generales, en el árbol ubicar el almacén del proyecto y en la pestaña permisos poner una selección de las personas que pueden aprobar los egresos.
Llamada f:
Indica la procedencia de la partida de control, por defecto muestra el nombre del proyecto
Llamada f-1: Sólo los recursos que tienen asignado partida de control del proyecto pasan a los resultados operativos del módulo de gerencia de proyectos.
Llamada f-2: Al elegir el subcontrato y asignar al recurso como partida de control, es utilizado para descontar de la valorización del subcontratista en el módulo de gerencia de proyectos, esta información no pasa a los resultados operativos.
Llamada h: Muestra los recursos que forman parte del ingreso, al hacer clic derecho sobre un recurso, muestra el menú,
Llamada h-1: Permite asignar partida de control al recurso.
Llamada h-2: Permite asignar destino específico al recurso.
Llamada h-3: Permite asignar frente al recurso.
Llamada h-4: En el caso de que el recurso sea un repuesto, combustible o accesorio, la lista de equipos disponibles para asignar al recurso.
Llamada i: Este segmento es informativo.
Llamada j: Esta celda es editable permite el registro de observaciones a la guía.
Llamada k: Graba el contenido de la guía con lo que termina el movimiento y es mostrado en el escenario de ingresos.
Registre el siguiente documento usando Ingreso por Compra y realice la descarga directa.
2.1.2.1.1.1 Regularizar el Ingreso por Compra con documento de pago
Los movimientos de ingreso por compra, que inicialmente se ingresó con una guía de remisión, se regularizan con la finalidad de vincularlo a un documento de pago, además para verificar que los recursos ingresados solo se paguen una sola vez al proveedor, y la factura pase al módulo de Facturación – documentos por pagar y por consiguiente a Contabilidad.
Regularizar con factura Nro. 001-234, registrar precios para:
Concreto premezclado PREMIX T.I. f´c=100 kg/cm2: S/. 140.
Concreto premezclado PREMIX T.I. f´c=175 kg/cm2: S/. 160.
Llamada a: Permite la regularización de varias guías de remisión registradas de un mismo proveedor, con una factura.
Regularización simple
Luego el sistema muestra las guías de remisión pendientes de regularización
Llamada a1-1: Es la fecha del registro.
Llamada a-1-2: Es la fecha de pago de la factura de parte de la empresa al proveedor.
Llamada a-1-3: Es la fecha de recepción de documento.
Llamada a-1-4: Elija el periodo de aplicación en contabilidad.
Llamada a-1-5: Muestra el periodo elegido para la contabilidad
Llamada a-1-6: Ingrese el número de serie y factura.
Llamada a-1-7: Elija la forma de pago, que es una regla.
Llamada a-1-8: Ingrese el porcentaje de descuento aplicado a todo el documento.
Llamada a-1-9: Ponga una selección si considera revisado el registro.
Llamada a-1-10: Ingrese en esta celda una observación relacionada a la factura.
Llamada a-1-11: Ingrese el precio de los recursos.
Llamada a-1-12: Verifique que estos montos coincida con la factura física.
Llamada a-1-13: Graba el contenido de registro, con lo que concluye.
Llamada b-1: Permite la regularización de varias guías de remisión registradas de un mismo proveedor, con varias facturas.
Llamada b-2: Permite la regularización de varias guías de remisión registradas de diferentes proveedores, con las facturas correspondientes.
2.1.2.1.2 Ingreso por Compra con documento de pago
Es cuando al momento de efectuar el movimiento se registra la factura como documento que respalda el ingreso del recurso al almacén.
Utilice esta opción cuando el ingreso esté respaldado con un documento de pago y se efectúe el asiento para contabilidad. Solo cuando el registro genere el ingreso al almacén será visualizado en este escenario, caso contrario vaya al escenario de Registro de Documentos
Datos que serán utilizados para el ingreso
Acceso al escenario
Elección del tipo de ingreso
Elija la opción de la llamada b.
Luego adicione los recursos
Luego use el botón «Aceptar» para concluir con el registro.
2.1.2.2 Ingreso por Préstamo IP]
Utilice esta opción cuando el proyecto reciba en calidad de préstamo recursos. Los recursos prestados deben ser devueltos, para esto el sistema lleva un control de lo que se ingresa y se devuelve.
Un almacén puede recibir recursos prestados de:
• De otro Proyecto
• De un Identificador (Proveedor, Empleado, Obrero, etc.)
En la barra de vistas elija el acceso al escenario de ingresos
En el árbol elija el proyecto y el almacén
Elija el tipo de ingreso, INGRESO POR PRESTAMO
Efectúe un nuevo ingreso
Registre el tipo cambio
Ingrese los datos de la cabecera en el formulario, elija el documento y registre el número de serie y número.
Luego elija al Socio de Negocio del catálogo.
En el cuerpo del documento adicione los recursos, puede tener como origen:
• Recursos prestados, es cuando en una misma base de datos se maneja la información de varios proyectos y por lo tanto mostrará los recursos que han egresado por préstamo hacia el proyecto en que se está trabajando.
• Desde el catálogo general, es cuando en la base de datos que se trabaja sólo maneja información de este proyecto y llegan recursos desde otro proyecto.
Del catálogo de recursos, elija los recursos prestados, ingrese la cantidad y el precio.
2.1.3 Movimientos que requieren de un movimiento previo
2.1.3.1 Ingreso por Devolución de Préstamo [IDP]
Este movimiento se genera cuando en un proyecto, una persona natural o jurídica (Socio de Negocio) devuelve al almacén un recurso previamente prestado.
Ver Egreso por Préstamo [EP]
Previamente ha tenido que generar movimientos de Egreso por Préstamo hacia el proyecto o Socio de negocio, que ahora devuelve el préstamo.
Un almacén puede recibir en devolución recursos prestados de:
1) De otro proyecto
2) De un Socio de negocio (Proveedor, Empleado, Obrero, etc.)
En la barra de vistas elija el acceso al escenario de ingresos.
En el árbol elija el proyecto y el almacén
Elija el tipo de ingreso, en este caso elija INGRESO POR DEVOLUCION DE PRESTAMO
Efectúe un nuevo ingreso
Registre el tipo cambio
Registre los datos en el detalle del formulario de ingreso
El sistema filtra sólo los recursos prestados al Socio de Negocio; del que debe elegir solo los recursos que devuelve.
Ingrese la cantidad que devuelve.
2.1.3.2 Ingreso por Devolución de Saldo [IDS]
Los saldos de los materiales se generan después de que los recursos egresan hacia una “Partida de Control”, eso significa que el costo íntegro de los recursos es cargado a la partida de control.
Sí después de utilizar los recursos hay sobrantes o saldos, se consideran como utilidad y aprovechamos para darle otros usos.
El saldo se considera como otro recurso más, se ingresa al almacén con la finalidad de controlarlo, entonces incrementa el stock por consiguiente el inventario valorizado y se egresa como cualquier otro recurso.
2.1.3.3 Ingreso por Fabricación [IF]
Existen algunos elementos que necesitan ser fabricados, como paneles para encofrado, escritorios, muebles de cocina, columnas o vigas prefabricas, etc., que luego serán egresados como recursos fabricados, que tiene como destino final la obra o para venta
Para fabricar el recurso PANEL PARA ENCOFRADO DE 4’ x 8’ CON TRIPLAY DE 19 mm, necesitamos:
1) Registrar el recurso PANEL PARA ENCOFRADO DE 4’ x 8’ CON TRIPLAY DE 19 mm; en el Catálogo de Recursos, luego registrar en la pestaña DETALLE DEL RECURSO, ingrese todos los recursos con sus respectivas cantidades que se necesitan para fabricar un PANEL PARA ENCOFRADO DE 4’ x 8’ CON TRIPLAY DE 19 mm.
2) Tener stock suficiente de los recursos que forman parte del detalle del recurso para la cantidad que queremos fabricar.
3) Generar la guía de Egreso para Fabricación de lo que deseamos fabricar; se fabrica y luego,
4) Generar la guía de Ingreso del recurso fabricado.
2.1.3.4 Ingreso por Orden de Compra [IOC]
Las órdenes de compra generadas y aprobadas en el módulo de Compras del S10, pasan al módulo de Almacenes para ingresar todos los recursos comprados. El sistema muestra en calidad de editado todas las órdenes de compra de donde se eligen las que serán ingresas al Almacén.
Esta es la orden de compra generada y aprobada (en el módulo de Compras), que será ingresada,
En la barra de vistas elija el acceso al escenario de ingresos.
En el árbol elija el proyecto y el almacén
Elección del proveedor
Reporte
Solo cuando lleguen los materiales a obra, y en las cantidades que aparecen en la orden de compra ACEPTE, caso contrario registre solo las cantidades que llegaron a obra.
No use esta opción si no tiene el módulo de Compras donde previamente se generó la orden de compra.
2.1.3.5 Ingreso por Reingreso [IR]
Cuando se realiza un egreso de recursos hacia una partida de control, y no se utilizó la totalidad, los recursos sobrantes deben ser devueltos al almacén e ingresados por la modalidad de “Ingreso por Reingreso”, con el fin de descargar los costos de la partida de control a la que fue egresado y cargado.
Previamente debe existir un egreso hacia una partida de control. Disponible para el almacén tipo material.
En la barra de vistas elija el acceso al escenario de ingresos
En el árbol elija el proyecto y el almacén
Elija el tipo de ingreso, INGRESO POR REINGRESO
Efectúe un nuevo ingreso
Registre del tipo cambio
En el cuerpo del documento haga un nuevo ingreso.
El sistema muestra en una ventana todos los recursos egresados hacia las partidas de control, de donde se elige los recursos que serán devueltos.
Reporte
2.1.3.6 Ingreso por Transferencia Externa [ITE]
Permite ingresar al almacén, recursos que son transferidos desde otros proyectos, la cantidad y costo son cargados al proyecto que recibe. La transferencia es permanente, no contempla devolución.
En la barra de vistas elija el acceso al escenario de ingresos
En el árbol elija el proyecto y el almacén
Elija el tipo de ingreso, INGRESO POR TRANSFERENCIA EXTERNA
Efectúe un nuevo ingreso
Registre el tipo cambio
Llamada a: Adicionar (Ingreso por Transferencia Externa manual): permite elegir recursos del catálogo, para luego registrar la cantidad y el precio.
Llamada b: Adicionar desde Guías de Egresos (Ingreso por Transferencia Externa automático): se activa, solo cuando existen guías pendientes de ingreso al almacén.
2.1.3.7 Ingreso por Transferencia Interna [ITI]
Esta opción se utiliza para dar ingreso a recursos transferidos de otro almacén del mismo proyecto, la cantidad y costo serán descargados del almacén que envía los recursos y cargados al almacén que recibe. La transferencia es permanente, no contempla devolución.
En la barra de vistas elija el acceso al escenario de ingresos
En el árbol elija el proyecto y el almacén
Elija el tipo de ingreso, INGRESO POR TRANSFERENCIA INTERNA
Efectúe un nuevo ingreso
Registre del tipo cambio
Llamada a: Adicionar (Ingreso por Transferencia Interna manual): permite elegir recursos del catálogo, para luego registrar la cantidad y el precio.
.
Llamada b: Adicionar desde Guías de Egresos (Ingreso por Transferencia Interna automático): utilice esta opción cuando los almacenes que envía y recibe están conectados, entonces el almacén que envía debe generar previamente un egreso por transferencia interna. El almacén que recibe visualizará la guía para realizar el ingreso.
2.1.4 Movimientos que requieren stock en el almacén
2.1.4.1 Ingreso por Ajuste de Cantidad [IAC]
Previamente se tiene que realizar un inventario físico de los recursos que existen en el almacén y si existen diferencias entre los saldos que muestran el sistema informático y el stock físico de un recurso, se deben realizar las correcciones del caso para igualar las cantidades.
Las diferencias son detectadas al realizar inventarios.
Para resolver el problema de la diferencia en stock:
a) Cuando en el escenario de Inventario, se realiza conteos, permite generar un ajuste automático.
b) Regularización manual, el sistema permite generar guía de ingreso o egreso por ajuste de cantidad según sea el caso.
Desarrollaremos el ítem b).
Al hacer el conteo del recurso ALAMBRE NEGRO RECIDO # 8 se tiene como resultado 22 Kg., pero el sistema indica que debe haber 20 Kg. en stock; en este caso esta diferencia debe ser resuelta ingresando la diferencia de 2 Kg. por ajuste de cantidad.
En la barra de vistas elija el acceso al escenario de ingresos
En el árbol elija el proyecto y el almacén
Elija el tipo de ingreso, INGRESO POR TRANSFERENCIA EXTERNA
Efectúe un nuevo ingreso
Registre del tipo cambio.
En el escenario, haga un nuevo ingreso, registre los datos de la cabecera del documento.
Luego en cuerpo del documento registre los recursos, ingrese las cantidades, el sistema por defecto toma el precios promedio.
2.2 Egresos
Permite realizar los movimientos de salida o egreso de los recursos que se encuentran en el stock del almacén, hacia diferentes destinos.
Para efectuar un movimiento de egreso tome en cuenta lo siguiente:
El usuario que registra los egresos debe tener atributos para hacerlo. Dependiendo del tipo de almacén en el que se encuentre trabajando.
Antes de efectuar un egreso, vaya al escenario de Stock y recostee.
Es necesario tener previamente registrado las partidas de control, destinos específicos, frentes, equipos, unidades operativas.
2.2.1 Procedimiento o pasos a seguir para efectuar un egreso – Egreso a Partida de Control [EPC]
Datos que serán registrados
Acceso al escenario
Elija el tipo de egreso
Registro de datos en la cabecera del formulario Egreso a Partida de Control
Llamada a:
En número de guía se autogenera. Sí ingresa un número, a partir de este se generará automáticamente los siguientes en cada movimiento que realice.
Llamada b:
Permite registrar notas al documento de egreso, acepta insertar archivos digitales.
Llamada c: Este botón muestra el catálogo de identificadores de donde se elige al solicitante.
Llamada d:
Este botón muestra el calendario, de donde se elige la fecha en la que se efectúa el movimiento de egreso (la fecha tiene que estar dentro del periodo de ejecución del proyecto).
Llamada e:
Esta celda es editable y permite el registro del número alterno.
Llamada f:
Este botón permite elegir el origen de la partida de control, por defecto muestra el nombre del proyecto, esto indica que el destino será una partida del control del proyecto.
Sí en el proyecto se usa subcontratos, estos se comportarán como una partida de control. Cuando los egresos tienen como destino un subcontrato, el monto de los recursos egresados puede ser descontado de la valorización del subcontrato en el período involucrado.
2.2.1.1 Detalle del formulario
Llamada a:
El botón Cuadro agrupar por, muestra la celda donde se arrastra la cabera de la columna con la finalidad de agrupar la información.
En el formulario de egreso podemos arrastrar la columna Partida de Control
Luego al desplegar cada una de las partidas de control podemos ver los recursos que forman parte de la PC.
Llamada b:
El botón
Muestra o esconde los Filtros avanzados, que permiten ubicar información.
Llamada c:
Mostrar campos, muestra la ventana con los campos (columnas) disponibles del formulario con la finalidad de mostrarlas u ocultarlas, demás cambiar de posición.
Llamada d:
Use este botón para enviar o exportar el contenido del formulario de egreso hacia una hoja Excel, durante el proceso el usuario debe elegir la carpeta donde se guardará el archivo exportado.
Llamada e:
Cursor automático, cuando está seleccionado el cursor se ubica en la celda que falta ingresar información, caso contrario es necesario usar el puntero del ratón para activar la celda.
Llamada f:
Muestra los grupos de filtro previamente almacenado con los botones de las llamas m, n con la finalidad de abrirlo.
Llamada f-1:
Graba el agrupamiento de la cabecera de las columnas.
Llamada f-2:
Graba el agrupamiento de la cabecera de las columnas, el usuario debe asignar un nombre al agrupamiento o aceptar el que siguiere el sistema.
Llamada g:
Es donde se registran los recursos que forman parte del egreso, cuando el formulario está VACIO muestra el menú.
Llamada g-1:
Muestra el stock de recursos para adicionar recursos al detalle del formulario de egreso.
Llamada g-4:
Elimina los recursos seleccionados del formulario de egreso.
Cuando el formulario TIENE RECURSOS muestra el menú,
Llamada g-a:
Elimina los recursos seleccionados.
Llamada g-b:
Asignar partida de control al recurso o recursos seleccionados.
Llamada g-c:
Asigna partida de control a todos los recursos.
Llamada g-d:
Asignar destino específico al recurso o recursos seleccionados, marque un bloque, luego haga clic derecho sobre el bloque y elija Asignar/Reasignar destino específico, del catálogo elija con doble clic.
Llamada g-f:
Asigna destino específico a todos los recursos.
Llamada g-g:
Asignar frente al recurso o recursos seleccionados, marque un bloque, luego haga clic derecho sobre el bloque y elija Asignar/Reasignar frente, del catálogo elija con doble clic.
Llamada g-h:
Asigna frente a todos los recursos.
Llamada g-i:
En la grilla del formulario existe una columna ACTIVO, con la finalidad de asignar activo al recurso que se egresa. Esta opción permite cambiar de activo al recurso que se egresa.
Llamada g-j:
Cambia de activo a todos los recursos egresados.
Llamada g-k:
Muestra el kardex del recurso.
Llamada g-2:
El sistema toma en cuenta los recursos solicitados por la persona, al invocar esta opción el sistema muestra los recursos que acostumbra pedir el usuario.
Llamada g-3:
Adicionar recursos por regresión, los recursos o materiales que se encuentran en cancha, como es la arena, ladrillos, hormigón, piedra chancada, se controlan de acuerdo a la participación que tienen dentro de una actividad o partida, el sistema calcula la cantidad a egresar de acuerdo a la valorización registrada de acuerdo al avance de obra en el periodo involucrado.
Para desarrollar este tema se toma como ejemplo un proyecto que maneje materiales en cancha.
Llamada g-3-1: Ubíquese en el escenario del Cronograma por periodos del módulo de Gerencia de proyectos, de la rama meta.
Llamada g-3-2: Ubíquese en el escenario del Cronograma por periodos.
Llamada g-3-3: Elija modo Valorización metrado e ingrese lo que avanzó o ejecutó (llamada g-3-5) en el periodo (llamada g-3-4).
Llamada g-3-6: Muestra la cantidad utilizada de arena, el que debe ser egresada por regresión.
Llamada g-3-6-a: El egreso está relacionado con la valorización de la semana mostrada.
Llamada g-3-6-b: Evalúe la cantidad utilizada e ingrese,
Luego asige actividades al recurso a egresar,
2.2.1.2 Impresión de reportes de egresos
Use el botón
para envía el reporte a la impresora previamente elegida en Windows.
Use el botón
para mostrar una vista previa o impresión en pantalla de la forma como será impreso el reporte.
IMPORTANTE: Imprima el reporte del egreso, haga firmar con la persona que retira el recurso del almacén y archive.
2.2.2 Egresos manuales
2.2.2.1 Egreso a Partida de Control [EPC]
Este tipo de egreso es el más usado durante la ejecución, es cuando el residente o maestro de obra envían un pedido al almacén y el almacenero efectúa un movimiento de egreso.
Requisito:
• Tener en el stock del almacén los recursos que serán egresados.
• Tener registrados las Partidas de Control (PC).
• Tener registrado los Destinos Específicos (DE).
• Tener registrado los Frentes (F).
• Tener registrado los Equipos (E).
Las PC se agrupan de acuerdo a su procedencia:
• Directas o propias del proyecto principal.
• Las que se originan a consecuencia de un subcontrato.
Use un egreso para los recursos que usarán partidas de control del proyecto, genere otro movimiento para los recursos que tienen como destino un subcontrato.
Dentro de una misma guía cada recurso puede ser asignado a diferentes PC. Un recurso puede estar más de una vez en el detalle de una misma guía pero con diferentes PC.
Datos para efectuar el egreso:
Use una guía de egreso para cada uno de las cuadrillas (acero y concreto).
En la barra de vistas elija el acceso al escenario de Egresos
En el árbol elija el proyecto y el almacén
Elija el tipo de egreso, EGRESO A PARTIDA DE CONTROL (EPC)
Efectúe un nuevo egreso
Registre del tipo cambio
Registre los datos en la cabecera del documento
En el cuerpo del documento adicione los recursos
Llamada a:
Por defecto, el sistema muestra las Partidas de Control del proyecto, los montos son mostrados en los Resultados Operativos.
El Subcontrato también aparce como si fuera una Partida de Control, los montos cargados no van a los resultados operativos, y son utilizados para deducir o descontarle al subcontratista al momento de efectuar su valorización.
Recursos que forman parte del Egreso
Registre la cantidad y asigne la Partida de Control,
Llamada a:
Si hace clic derecho sobre el recurso, muestra el menú de donde se puede asignar partidas de Control, Destinos Específicos, Frentes, a cada uno de los recursos que forman parte del formulario de egreso.
Reporte
La persona que retira los recursos del almacén debe firmar la guía impresa y esta debe ser archivada.
2.2.2.2 Egreso por Préstamo [EP]
Este tipo de movimiento se utiliza para registrar los recursos que presta el almacén, es requisito tener en stock la cantidad necesaria del recurso a prestar.
El destino puede ser:
1) Otro almacén del mismo proyecto.
2) Un Proyecto (otra obra).
2) Un Socio de negocio (Un proveedor, empleado, obrero, etc.).
Los recursos que se prestan deben ser devueltos, para visualizar los recursos prestados, vaya al grupo de informes, acceso directo Préstamos, y vea Lo que Presté.
Del stock elija el recurso que se presta
2.2.2.3 Egreso por Transferencia Externa [ETE]
Este tipo de movimiento se utiliza para registrar los recursos que se transfieren hacia otro proyecto, es requisito tener en stock la cantidad necesaria del recurso a transferir.
Los recursos que se transfieran tienen el carácter de definitivo y no se espera devolución.
En el almacén del proyecto origen se genera un Egreso por Transferencia Externa y en el almacén destino se registra un Ingreso por Transferencia Externa.
Los proyectos involucrados deben estar registrados en la misma base de datos.
2.2.2.4 Egreso por Transferencia Interna [ETI]
Este tipo de movimiento se utiliza cuando se transfiere recursos entre almacenes del mismo proyecto, es requisito tener en stock la cantidad necesaria del recurso a transferir.
Los recursos que se transfieren tienen el carácter de definitivo y no se espera devolución.
En el almacén origen se genera un “Egreso por Transferencia Interna” y en el almacén destino se genera un “Ingreso por Transferencia Interna”.
2.2.3 Egresos que solo son informativos
2.2.3.1 Egreso a Partida de Control con Descarga Directa [EPCDD]
Este escenario es sólo de consulta, ya que los [E PCDD] son generados desde los diferentes tipos de Ingreso.
2.2.4 Egresos que dependen de un pedido
2.2.4.1 Egreso a Partida de Control con Pedido [EPCP]
Utilice este tipo de movimiento cuando su empresa trabaja con el módulo de Almacenes, Compras, o Gerencia de proyectos, donde el pedido es generado ‘A Almacén’, el pedido reserva toda la cantidad solicitada o lo que hubiera del stock del almacén a través de una Nota de Egreso.
Pasos a seguir para utilizar éste tipo de movimiento:
1.- Desde el módulo de Almacenes, Compras o Gerencia de proyectos, efectúe un pedido “A Almacén”.
Llamada a: Registre el nombre del solicitante.
Llamada b: El número alterno no es obligatorio, si usa un talonario de pedidos puede registrar el que le corresponde.
Llamada c: Si los recursos a pedir tienen corresponden a una orden de producción, registre el número.
Llamada d: Seleccione A Almacén, para que los recursos a pedir sean del stock.
Llamada e: Registre la fecha del pedido.
Llamada f: Elija la procedencia de la partida de control que asignará a los recursos, pueden ser para el proyecto o un subcontrato.
Llamada g: Haga clic derecho en este segmento para elegir los recursos del stock, luego con clic derecho sobre el recurso asigne partida de control, destino específico, frente o equipo.
2.- En el módulo de Almacenes para efectuar el Egreso a Partida de Control con Pedido.
2.2.4.2 Egreso por Transferencia Externa con Pedido [ETEP]
Este movimiento permite transferir recursos de un proyecto a otro, no considera devolución.
Pasos a seguir para utilizar éste tipo de movimiento:
1.- Desde el módulo de Almacenes, Compras o Gerencia de proyectos, efectúe un pedido “A Almacén”.
Llamada a:
Registre el nombre del solicitante.
Llamada b:
Seleccione A Almacén, para que los recursos a pedir sean del stock.
Llamada c:
Seleccione para otro almacén y una de las opciones marcadas con las llamadas c-1, c-2, c-3, c-4.
Llamada c-1:
Permite efectuar egresos para vender recursos del almacén origen a un proyecto y almacén destino marcados con las llamadas c-5 y c-6. Previamente debe estar seleccionada la opción con la llamada c.
Llamada c-2
: Permite efectuar egresos para transferencia del almacén origen a un proyecto y almacén destino marcados con las llamadas c-5 y c-6. Previamente debe estar seleccionada la opción con la llamada c y c-2.
Llamada c-3:
Permite efectuar egresos para transferencia externa del almacén origen a un proyecto y almacén destino marcados con las llamadas c-5 y c-6. Previamente debe estar seleccionada la opción con la llamada c y c-3.
Llamada c-4:
Permite efectuar egresos para transferencia interna del almacén origen a un almacén destino del mismo proyecto, marcados con la llamada c-6. Previamente debe estar seleccionada la opción con la llamada c y c-4.
2.- En el módulo de Almacenes para efectuar el Egreso a Partida de Control con Pedido.
2.2.4.3 Egreso por Transferencia Interna con Pedido [ETEP]
Utilice este movimiento para trasladar los recursos de un almacén a otro dentro del mismo proyecto.
Pasos a seguir para utilizar éste tipo de movimiento:
1.- Desde el módulo de Almacenes, Compras o Gerencia de proyectos, efectúe un pedido “A Almacén”.
Llamada a: Registre el nombre del solicitante.
Llamada b: Seleccione A Almacén, para que los recursos a pedir sean del stock.
Llamada c: Seleccione para otro almacén.
Llamada d: Seleccione Transferencia.
Llamada e: Seleccione Interna.
Llamada f: Por defecto el sistema muestra el nombre del mismo proyecto.
Llamada g: Seleccione el almacén destino.
2.- En el almacén destino efectué un Ingreso por Transferencia Interna con pedido.
2.2.5 Egresos que dependen de información previa en el catalogo de recursos
2.2.5.1 Egreso para Fabricación [EFE]
En el catálogo de recursos debe registrar el recurso a fabricar y también los recursos que serán utilizados en la fabricación.
Llamada a: Registre el nombre del recurso a fabricar.
Llamada b: Registre la unidad.
Llamada c: Elija la pestaña Detalle recurso.
Llamada d: Haga clic derecho en ese segmento y elija los recursos que necesita para fabricar el PANEL PARA ENCOFRADO DE 4” x 8” CON TRIPLAY DE 19 mm.
Es requisito que los recursos que forman parte del Detalle, existan en el stock del almacén y en las cantidades necesarias.
Llamada e: Registre las cantidades para fabricar un panel,
Efectúe un egreso para fabricación, del recurso
Llamada a: Registre el nombre del solicitante.
Llamada b: Registre la fecha del movimiento.
Llamada c: Elija el recurso a fabricar.
Llamada d: Ingrese la cantidad a fabricar.
Llamada e: Es el precio calculado, es un dato informativo.
Llamada f: Este segmento muestra los recursos que serán utilizados en la fabricación del recurso.
Llamada f-1: Para este recurso, la llamada f-2 muestra el stock en el almacén.
Llamada g: Es cantidad unitaria para fabricar un recurso.
Llamada h: Es el precio calculado, es un dato informativo.
Llamada i: Es la cantidad de recursos que se necesita para fabricar 12 paneles, en este caso.
Llamada j: Graba el contenido de la ventana.
Después de fabricar el recurso ingresa al almacén de elementos fabricados.
2.2.6 Egresos que se utilizan para devolver recursos
2.2.6.1 Egreso por Devolución al Proveedor [E DPR]
Este tipo de movimiento se utiliza cuando los recursos adquiridos a un determinado proveedor deben ser devueltos por alguna razón, ej.: no cumplen con las especificaciones técnicas o calidad establecida por la empresa o simplemente se equivocaron en solicitar el recurso.
Para efectuar la devolución, el recurso previamente ha debido ser registrado en el almacén con un de ingreso por compra. El recurso egresado toma el precio promedio del sistema
2.2.6.2 Egreso por Devolución de Préstamo [EDP]
Este tipo de movimiento se utiliza para registrar la devolución de recursos que ingresaron al almacén con un Ingreso por Préstamo.
Sólo se devuelve un recurso prestado.
La devolución se puede realizar a:
1) Un Proyecto (una obra)
2) Un Identificador (Un proveedor, empleado, obrero, etc.)
El sistema almacena a todas las personas naturales o jurídicas que prestaron recursos (con sus respectivas cantidades).
2.2.6.3 Egreso por Devolución por orden de Compra [EDOC]
Este tipo de movimiento se utiliza para registrar la devolución de recursos que ingresaron al almacén con un Ingreso por orden de Compra.
Sólo se devuelve un recurso que ingresó con orden de compra.
Para devolver, en el formulario de Egreso por Devolución por Orden de Compra, invoque al proveedor que devolverán los recursos.
El sistema almacena a todas los proveedores que tiene ingreso por orden de compra, con los recursos y cantidades.
2.2.7 Egresos para vender recursos del proyecto
2.2.7.1 Egreso por Venta [EV]
Este tipo de movimiento se utiliza cuando egresan recursos del stock de almacén con fines de venta. La persona o empresa que compra debe estar registrado como socio de negocio en el sistema. Los recursos egresados toman el precio promedio del sistema.
2.2.7.2 Egreso por Venta por Transferencia [EVT]
Este tipo de movimiento se utiliza cuando egresan recursos del stock de almacén con fines de venta para otro proyecto. Es necesario generar un pedido.
Llamada a: Registre el nombre del solicitante.
Llamada b: Seleccione A Almacén.
Llamada c: Seleccione Para otro almacén.
Llamada d: Seleccione para Venta.
Llamada e: Elija el proyecto destino.
Llamada f: Elija el almacén destino.
Llamada g: Registre la fecha del movimiento.
Llamada h: Haga clic derecho en este segmento y elija del stock al recurso.
Llamada i: Ingrese la cantidad que se vende.
Llamada j: Este segmento es informativo, muestra el stock del recurso en el almacén que vende.
Llamada k: Graba el contenido del pedido y envía a probación.
El pedido debe ser aprobado.
En el proyecto origen haga un Egreso por Venta por Pedido.
El en proyecto destino efectuar el Ingreso por Compra.
2.2.8 Egresos para regular la información del almacén
2.2.8.1 Egreso por Ajuste de Cantidad [EAC]
Sí después de efectuar el inventario, se detectan diferencias entre el stock físico de un recurso y lo que muestra el sistema informático, se deben realizar las correcciones del caso para que las cantidades sean iguales.
Las diferencias pueden ser detectadas al hacer un inventario con el módulo de Almacenes S10.
Sí, luego de realizar el conteo, encontramos que el CEMENTO PÓRTLAND TIPO I (42.5 kg.) se tiene como resultado 3581 bolsas, pero el sistema indica que deben haber 3580 bolsas en stock; la diferencia de 1 bolsa debe ser egresada por Ajuste de Cantidad.
El recurso egresado toma el precio promedio.
2.2.8.2 Egreso por Bajas y Mermas [EBM]
Este tipo de movimiento se genera para controlar aquellos recursos que se han deteriorado o malogrado dentro del almacén o producto del manipuleo.
El costo de estos recursos es cargado al proyecto. El recurso egresado toma el precio promedio.
El recurso que se dará baja se elige del stock.
Al momento de efectuar el egreso por bajas y mermas, en el formulario de egreso, asigne al recurso la partida de control NO CONTRINUTORIO – BAJAS – MERMAS.
3. Stock
Muestra la cantidad de reservas o existencias de un determinado recurso en el almacén, al momento de realizar la consulta.
El resultado de los movimientos de ingresos y egresos, constituye el kardex de un recurso.
Desde este escenario, se pueden realizar las siguientes acciones:
• Se recostea o procesan los recursos para efectuar las sumas y restas y tome el precio promedio.
• Consulta de saldos y stock reservado.
• Se imprime el stock valorizado a una fecha, documento que se acompaña a los resultados operativos que presenta el Residente de obra.
• Se asigna stock mínimo a lo recursos.
• Se asigna ubicación física a los recursos.
• Se visualiza el kardex o historia de los movimientos de los recursos y se ingresa al documento involucrado.
• Desde kardex, se imprime el reporte para la SUNAT.
• Analizar recursos con problemas.
• Se registran los sobrantes del recurso.
Acceso al escenario
3.1 Recosteo
Cuando muestre el botón , junto al nombre de los recursos, significa que deben recostear, para que tome el precio promedio. Elija el botón
Durante el proceso de recosteo utiliza el factor de cambio, en caso de no estar registrado el sistema solicita el ingreso.
3.2 Consulta de saldos y stock reservado
Llamada a: Use las opciones del filtro avanzado para ubicar un recurso.
Llamada b: Muestra el stock del recurso en el almacén.
Llamada c: Cuando se realiza un pedido A Almacén con destino a una partida de control, el sistema reserva la cantidad y resta del stock.
3.3 Impresión del stock valorizado.
Este reporte se adjunta a los resultados operativos que presentará el Residente de Obra, muestra la cantidad de recursos y el costo a la fecha de consulta.
Use el botón
para realizar una vista preliminar.
3.4 Asignar stock mínimo a los recursos.
Dentro del almacén hay recursos que siempre deben tener stock, como es el caso de los combustibles, para registrar el stock mínimo tome en cuenta el tiempo de reposición.
Llamada a: Muestra los recursos que tienen stock mínimo.
Llamada b: Registre la cantidad que será el stock mínimo del recurso, cuando la cantidad es menor la stock se pinta de celeste la fila.
Llamada c: Cuando el stock del recurso está por debajo del mínimo genera un pedido.
3.5 Asignar ubicación física dentro del almacén a un recurso
A cada uno de los recursos que existen en el almacén se les puede asignar una ubicación física.
Requisito previo: En el grupo organización, registre las ubicaciones.
La ubicación física es por almacén
Ingreso a la opción
3.6 Kardex
Es la historia del recurso que muestra todos sus movimientos a través del tiempo.
Para elegir la opción.
Para analizar el recurso CEMENTO PÓRTLAND TIPO I (42.5 Kg.)
Reporte para la SUNAT
Use el botón
Llamada a: Elija SUNAT.
Llamada b: Defina el rango de fechas.
Llamada c: Selecciones todos los recursos.
Llamada d: Imprime el reporte.
3.6.1 Recursos con problemas
Cuando los ingresos se hacen a una fecha y lo egresos a una fecha posterior, no existe ningún problema. El error se genera cuando la fecha del documento de egreso se modifica a una fecha anterior a la del ingreso. El error también se genera cuando se borra la guía de ingreso.
La forma de arreglar es la siguiente; visualice los recursos con problemas, luego haga doble clic sobre el recurso y el sistema mostrará la guía de egreso, modifique la fecha a una posterior.
Para elegir la opción.
3.6.2 Sobrantes del recurso
Es un archivo donde se registran los sobrantes (retazos) del recurso, esta opción sólo es de control, si retira uno de los retazos, modifique la cantidad.
3.7 Guías de Remisión
Este escenario es informativo muestra las guías de remisión generadas en el proyecto.
La guía de remisión se genera a partir de un egreso.
La guía de remisión se utiliza para trasladar bienes o recursos.
Debido a que es un documento pre impreso que requiere de una autorización de SUNAT.
Sólo se generan guías de remisión a partir de los egresos.
3.8 Facturación
El módulo de almacenes es secundario o de ayuda en facturación, este módulo es utilizado para registrar facturas relacionando con un movimiento de ingreso.
Sí registra facturas o documentos de pago desde el escenario de Documentos por pagar debe relacionarlo con un movimiento de ingreso necesariamente.
Los documentos registrados son enviados a documentos por pagar y por consiguiente a contabilidad para generar los asientos en forma automática.
Los documentos contables son facturas, boletas de venta, ticket, etc.
Los tipos de documentos que se utilizarán deben ser registrados previamente en el sistema.
3.8.1 Datos Generales
El usuario encargado encontrará registrada a la empresa y la sucursal. Caso contrario use este escenario para adicionar al usuario y dar permisos.
El usuario encargado de registrar información en Facturación, debe tener categoría asignada y permisos, esta labor se realiza en el grupo de Utilitarios.
Ingrese al módulo de almacenes como sa y acceda al escenario Datos Generales.
Adicione al usuario: ANDRADE ANDRADES, ANDRES,
Del catálogo de socio de negocio elija a: ANDRADE ANDRADES, ANDRES
En el Catálogo de Socio de negocio, registre S10 Ecuador, y califique como SUCURSAL, luego alija.
3.8.2 Periodo Contable
Requisito, los periodos contables deben estar generados:
Ingrese al módulo de almacenes como sa y acceda al escenario
Llamada a: El mes donde se registrará la información tiene que estar en estado Abierto.
Llamada b: Al hacer clic derecho en este segmento, muestra el menú.
Llamada b-1: Permite el registro de un nuevo año.
Llamada b-2: Al modificar un periodo muestra la ventana,
Permite modificar el nombre del perido, estado y hacer visible o esconderlo.
Llamada b-3: Elimina un periodo siempre y cuando no tenga registros relacionados.
3.8.3 Documentos por pagar
En este escenario se registran los documentos por pagar, como son las facturas, boletas de venta, ticket etc., y se relacionan con un movimiento previamente efectuado en ingresos.
Datos que serán utilizados para el ingreso
Elija una sucursal para registrar información
Registro del documento
Llamada a: Permite adicionar un documento de pago y muestra el formulario de ingreso.
Llamada b: Es una plantilla, utilice para el pago de servicios, que cada mes nos envían para pagarlos
Formulario de ingreso (Se describe la opción de la llamada a)
Nota: El botón
indica que el dato que se debe registrar en la celda, lo debemos elegir de un catálogo, en caso de que no exista, previamente registre.
El botón
indica que el dato que se debe registrar en la celda, lo debemos elegir de un listado predefinido por el sistema.
Llamada a: Es la Empresa elegida, para el cual se registrará en documento.
Llamada b: Es un número correlativo, empezará del número que el usuario ingrese, caso contrario del 1.
Llamada c: Use el botón
y del catálogo elija al proveedor.
Llamada d: Registre la fecha de registro del documento.
Llamada e: Registre la fecha de cancelación del documento.
Llamada e-1: Registre la fecha que recibió el documento del proveedor.
Llamada e-2: Registre la fecha del periodo contable para el documento.
Llamada f: Elija el tipo de moneda.
Llamada g: Use el botón registre y elija la forma de pago. Para registrar una nueva forma de pago:
Llamada h: Elija al proveedor.
Llamada i: Ingrese la serie y número del documento.
Llamada j: Esta celda seleccionada toma el tipo de cambio.
Llamada k: Registre el porcentaje de descuento (considerado en la factura).
Llamada L: Relacionado con, muestra las opciones.
Llamada L-1: Ninguno, elija para compras directas como refrigerios.
Llamada L-2: Permite relacionar el documento que se registra con un movimiento de ingreso por compra.
Llamada L-3: Permite relacionar el documento que se registra con una orden de compra o servicio generada. Luego puede hacer un ingreso al almacén.
Llamada L-4: Permite relacionar el documento que se registra con una valorización de subcontrato.
Llamada L-5: Permite relación el documento que se registra con un movimiento ingreso por orden de compra o servicio.
Llamada n: Elija esta opción cuando el gasto no está relacionado a ningún proyecto o centro de costo, esto significa que la empresa sólo usa el módulo de Almacenes del S10 y Contabilidad y solo le interesa registrar el gasto.
Llamada m: Seleccionado, el documento registrado será cargado al proyecto.
Llamada o: Seleccionado, el documento registrado será cargado a un centro de costo personalizado, que es ajeno a la empresa, luego permite exportar la información a Excel.
Llamada p: De acuerdo a la selección marcada con la llamada “m”, elija el proyecto.
Llamada r: Es el lugar donde se registran los recursos del documento, previamente elija la pestaña con la llamada r-a.
Llamada r-1: Invoca al Catálogo de recursos, para elegir todos los recursos que formarán parte del documento.
Llamada r-2: Elimina el recurso seleccionado de la ventana marcada con la llamada r.
Llamada r-3: Cambia por un recurso del catálogo.
Llamada r-4: Invoca al Catálogo de Destinos Específicos Global, para asignar al documento que se registra, esto permite ver los gastos a nivel de destinos.
Llamada r-5: Disponible sólo si previamente se asignó Destinos Específicos Global.
Llamada r-6: Muestra el submenú con las opciones que se detallan a continuación.
Llamada r-6-1: Invoca al Catálogo de Partida de control general, para asignar al documento que se registra, esto permite ver los gastos a nivel de toda la empresa.
Llamada r-6-2: Es un parámetro definido en el módulo de Contabilidad del S10, permite asignar a Estructura de Costo al recurso.
Llamada r-6-3: Permite asignar Centro de Costo al recurso.
Llamada r-7: Asignar IGV
Llamada r-8: Quitar IGV
Llamada r-7: Exporta el contenido de la ventana a una hoja Excel.
Llamada r-b: Muestra las propiedades del recurso.
Llamada r-b-1: Del catálogo elija el motivo de pago.
Llamada r-b-2: Del catálogo elija el método de pago.
Llamada r-c: Genera automáticamente el asiento contable. Existen requisitos previos:
• Ingrese al módulo de Contabilidad del S10.
• En el grupo de Inicio, ingrese al escenario de Datos generales – en el árbol elija la empresa y asigne el Esquema contable.
• En el grupo de Inicio, escenario de Datos generales – en el árbol elija la sucursal registre a los usuarios y asigne permisos.
• En el grupo de Parámetros, escenario de Plan contable – registrar y configure el plan contable.
• En el grupo de Parámetros, escenario de Asignación de cuentas – asignar las cuentas contables a los recursos.
Llamada r-d: Esta pestaña es informativa, muestra el tipo de impuesto aplicado.
Llamada s: Es ventana es informativa, que muestra información después de efectuar el pago.
Llamada t: Muestra los impuestos asignados al recurso.
Llamada u: Permite adicionar recursos al formulario.
Reporte
Opciones implementadas,
Llamadas de la ventana “Documentos por Pagar”
Llamada u-1: Muestra el formulario de registro del documento, para hacer un nuevo registro.
Llamada u-2: Previamente realice un registro normal del documento de pago del proveedor, luego con la opción de la llamada u-15 convierta en plantilla de pago.
Al usar la opción con la llamada u-2, muestra el Catálogo de Plantillas de Cargo para elegir y muestre los datos para efectuar un nuevo registro.
Llamada u-3: Permite modificar el contenido de un registro existente.
Llamada u-4: Elimina registro existente siempre y cuando no esté pagado o tenga un ingreso al almacén.
Llamada u-5: Duplica un registro existente.
Llamada u-6: Envía el contenido de la ventana a un archivo texto para que pueda ser aprovechado por otro sistema.
Llamada u-7: El registro elegido asigna a otro proyecto.
Llamada u-8: Muestra el menú:
Llamada u-8-1: Cambia de centro de costo a los documentos seleccionados.
Llamada u-8-1: Cambia de centro de costo a todos los documentos del escenario.
Llamada u-9: Solo en el caso de haber reasignado documentos, calcula los parciales.
Llamada u-10: Regenera los asientos contables del registro seleccionado.
Llamada u-11: Es un utilitario que verifica el estado del documento registrado.
Llamada u-12: Cambiar tipo de documento.
Llamada u-13: Actualiza documentos relacionados (simple).
Llamada u-14: Actualiza documentos relacionados (compuesta).
Llamada u-15: Una vez registrado un documento, se define como plantilla de cargo
Llamada u-16: Edita la plantilla de cargo registrada.
Llamada u-17: Importa de Excel información.
Llamada u-18: Este utilitario es permite enviar información de acuerdo al siguiente menú:
Llamada u-18-1: Plan de cuentas externo
Llamada u-18-2: Centro de costo personalizado.
Llamada u-18-3: Recursos
Llamada u-18-4: Tipo de documentos.
Llamada u-18-5: Tipo de impuesto.
Llamada u-19: Muestra en pantalla la apariencia del reporte.
Llamada u-20: Envía el reporte a la impresora previamente elegida desde Windows.
3.8.4 Compras y Ventas
Este escenario es informativo,
Acceso al escenario
Llamada a:
Llamada b: Seleccionado muestra todos los periodos que contienen información.
Llamada c: Permite elegir un periodo de consulta, es requisito que no esté seleccionado la opción marcada con la llamada b.
Llamada d: Arrastre a este segmento la cabecera de la columna que desea agrupar.
Llamada e: Muestra el documento registrado, al hacer clic derecho muestra el menú.
Llamada e-1: Procesa los parciales, en caso de haber cambios.
Llamada e-2: En el caso de haber cambiado los documentos relacionados al documento registrado, refresca.
Llamada e-3: Muestra el documento registrado.
3.8.5 Compras y Ventas Detallado
Este escenario muestra el registro de los movimientos Por Cobrar y por pagar, del periodo elegido.
Acceso al escenario
Llamada a: Selección el tipo de registro o libro que desea visualizar en el escenario.
Llamada b: Seleccionado, muestra la información de todos los periodos.
No seleccionado, permite elegir un periodo.
Llamada c: Cuando la opción marcada con la llamada “b” no está seleccionada, permite elegir un periodo.
Llamada d: Este botón contiene los campos o columnas disponibles del escenario, que luego se arrastra al segmento marcado con la llamada d-1, para agrupar la información.
Llamada e: Muestra la información en la grilla, al hacer clic derecho muestra el menú,
Llamada e-1: Muestra el detalle del documento solicitado.
Llamada e-2: Calcula los datos que se muestran en la grilla.
Llamada e-3: Es un utilitario que refresca los documentos relacionados al registro.
3.8.6 Fondo Rotatorio
El fondo rotatorio llamado también caja chica, es una cantidad de dinero que se asigna al encargado de manejar estos fondos en el proyecto, el sistema permite el registro de la asignación de los fondos, y el registro de los documentos que sustentan la rendición al mismo tiempo el ingreso al almacén de los recursos.
Requisitos previos, en el catálogo de recursos registre el recurso: FONDO ROTATORIO.
Luego en el Catálogo Tipo de Documento, esté registrado el documento RECIBO INTERNO.
Verifique la configuración en Documentos por Pagar, use el botón
Ingrese al módulo de Almacenes con sa.
Acceso al escenario: Todos los documentos que registre serán sólo del proyecto.
Luego,
Llamada a: Elija la pestaña Usuarios por Sucursal.
Llamada b: En el segmento marcado con la llamada “b”, haga clic derecho y del catálogo de socio de negocio elija al usuario.
Llamada c: Elija la pestaña Permisos, ubique el nombre del usuario y asigne permisos.
Ingresar como usuario al módulo de Almacenes.
3.8.6.1 Asignar fondos
Acceso al escenario: Todos los documentos que registre serán sólo del proyecto (A).
Acceso al escenario: Todos los documentos que registre es por empresa (B).
Uso de la opción (A), registro de información por Empresa.
Datos que serán registrados
Registro del nombre del Fondo Rotario.
Llamada a: Registre los datos de la cabera del formulario Fondo Rotatorio.
Llamada b: Elija la pestaña Asignación de Fondos.
Llamada c: En este segmento se registra la Asignación de Fondos.
Llamada c-1: Use esta opción cuando previamente tenga generado
Llamada c-2: Use esta opción cuando en el grupo Inicio del módulo de Facturación se generaron talonarios.
En el formulario de Recibo Interno,
En el formulario FONDO ROTATORIO
3.8.6.2 Rembolsos
Para ingresar haga doble clic en la ventana superior sobre el nombre del encargado del fondo rotatorio, el sistema mostrará la ventana FONDO ROTATORIO, elija la ventana MOVIMIENTOS y con clic derecho elija generar rembolso.
3.9 Resultados
Los informes se generar a partir los movimientos de ingresos y egresos, antes de emitir los informes ingrese al escenario de Stock y recostee.
3.10 Resultados por destino – Partidas de control
Permite obtener resultados valorizados en un determinado rango de fechas para diferentes destinos los cuales dependen del tipo de almacén elegido.
Los resultados por destino de los almacenes tipo material son tomados para los resultados operativos del módulo de Gerencia del proyectos.
Acceso al escenario
El botón
procesa los datos del período elegido o el rango de fechas. (Antes de usar este botón elija el periodo o rango de fechas).
Llamada a: Seleccionado: Toma todos los datos de los movimientos de ingresos y egresos desde la fecha que hicieron el primer movimiento hasta el día de la consulta.
No seleccionado: Toma sólo los datos de los movimientos del período elegido con la llamada “d”.
Llamada b: Tiene dos funciones:
• Sí necesita tener un reporte dentro de un rango de fechas, permite elegir la fecha inicio.
• En el caso de que elija un período con el botón marcado con la llamada “d”, esta celda sólo es informativa, que indica el primer día del período.
Llamada c: Tiene dos funciones:
• Sí necesita tener un reporte dentro de un rango de fechas, permite elegir la fecha fin.
• En el caso de que elija un período con el botón marcado con la llamada “d”, esta celda sólo es informativa, que indica el último día del período.
Llamada d: Este botón permite elegir el tipo de período
Llamada e: Lista las partidas de control del proyecto, con la finalidad de elegirla, y en la ventana inferior muestra los recursos utilizados para el periodo que se informa.
Llamada f: Lista los recursos de la partida de control elegida. Al hacer doble clic muestra los documentos de egreso donde está involucrado el recurso
3.11 Resultados por Destinos Específicos
Este escenario permite obtener resultados valorizados en un determinado rango de fechas para diferentes destinos específicos.
Es requisito haber asignado un destino específico al egreso.
Los resultados por destino específico deben ser utilizados en el almacén tipo material.
Informe de los recursos egresados a cada uno de los destinos específicos.
El botón procesa los datos del período elegido o el rango de fechas. (Antes de usar este botón elija el periodo o rango de fechas).
Llamada a: Seleccionado: Toma todos los datos de los movimientos de ingresos y egresos desde la fecha que hicieron el primer movimiento hasta el día de la consulta.
No seleccionado: Toma sólo los datos de los movimientos del período elegido con la llamada “d”.
Llamada b: Tiene dos funciones:
• Sí necesita tener un reporte dentro de un rango de fechas, permite elegir la fecha inicio.
• En el caso de que elija un período con el botón marcado con la llamada “d”, esta celda sólo es informativa, que indica el primer día del período.
Llamada c: Tiene dos funciones:
• Sí necesita tener un reporte dentro de un rango de fechas, permite elegir la fecha fin.
• En el caso de que elija un período con el botón marcado con la llamada “d”, esta celda sólo es informativa, que indica el último día del período.
Llamada d: Este botón permite elegir el tipo de período
Llamada e: Lista los destinos específicos del proyecto, con la finalidad de elegirla, y en la ventana inferior muestra los recursos utilizados para el periodo que se informa.
Llamada f: Lista los recursos del destino específico elegido. Al hacer doble clic muestra los documentos de egreso donde está involucrado el recurso
3.12 Resultados por Frente
Si a los recursos que forman parte del egreso se le asignó Frente, este escenario muestra resultados por frente en un determinado rango de fechas.
3.13 Resultados por Destino Detallado
Este escenario es informativo permite agrupar información de los egresos hacia los destinos específicos.
3.14 Físico valorizado
Este escenario permite obtener resultados del stock físico y valorizado, con los ingresos, egresos, en cantidades y valores, en un determinado rango de fechas
3.15 Resumen por tipo de movimiento
Este escenario es informativo, lista los tipos de ingreso y egreso disponible en el sistema, y los montos que corresponden al movimiento efectuado en un rango de fechas.
3.16 Resultados de Auditorias
Sí en los catálogos se realizaron auditorías, este escenario muestra los resultados de las auditorías, es por almacén.
3.17 Resultados por Recursos de Control
Este escenario es informativo, muestra todos los recursos agrupados a un Recurso de Control, es necesario:
• En el grupo de Organización, escenario de “Recursos de control” registrar el catálogo de los recursos de control.
• Asignar recursos a cada uno de los “Recursos de Control”
3.18 Resultados por clasificación de recursos
Este escenario es informativo, muestra todos los recursos agrupados a un ítem de “Clasificación de Recursos” y cuyos recursos forman parte de un egreso de los siguientes tipos:
Llamada a: En esta ventana se registra el nombre de la clasificación.
Llamada b: Lista todos los recursos presupuestados para el proyecto.
Llamada c: Muestra los recursos asignado a un ítem de elegido de la ventana marcada con la llamada a.
3.19 Pedidos
Permite elaborar pedidos de recursos del proyecto para comprar o para ser utilizados en obra. Antes de pedir verifique en el módulos Compras que esté registrado el Centro de Compra y que el proyecto esté asignado a ese centro de compra, también que exista aprobadores.
Llamada a: En los movimientos de ingreso se registran las facturas que pasan a Documentos por pagar del módulo de Facturación y por consiguiente a contabilidad.
3.19.1 Datos generales
Al proyecto, se asigna a los usuarios con sus respectivos permisos, con la finalidad de que puedan pedir recursos y aprobarlos.
Pide el encargado de manejar el módulo de Gerencia de proyectos y aprueba el Residente de obra.
Acceso
El usuario que manejo el módulo de almacenes, así como a los que aprueban se agrega en la pestaña Usuarios por Almacén.
Luego en la pestaña de permisos, se selecciona los escenarios a los que ingresará.
3.19.2 Escenario de Pedidos
Antes de efectuar los pedidos, se ingresa al escenario de Techo de Materiales, para clasificar los recursos por Techo por Cantidad y Techo por Precio.
Los pedidos se realizan desde el almacén de un proyecto, tome en cuenta el siguiente comentario:
Verifique la configuración de Pedidos
Use el botón
Llamada a: Seleccionado, solicita la elección de la fecha de entrega de los recursos de parte del Proveedor al almacén del proyecto, estas fechas son mostradas en la impresión de la orden de compra y además permite filtrar la información por fechas en el escenario CRONOGRANA DE ENTREGAS DETALLADO.
No seleccionado, acepta grabar el pedido sin fecha de entrega.
Llamada b: Seleccione para enviar correo al aprobador.
Para efectuar un pedido agrupe recursos que distribuye o vende el proveedor, esto facilitará generar orden de compra a partir de un pedido.
Pida combustibles para el proyecto PUENTE
Acceso al escenario de Pedidos
El formulario que se muestra a continuación se usa para efectuar pedidos de compra y pedidos a almacén.
Llamada a: Elija el nombre del solicitante.
Llamada b: Este número se autogenera.
Llamada c: En caso de tener un número alterno, registre, esta celda es editable.
Llamada d: Si tiene orden de producción ingrese el número.
Llamada e: Seleccione en el caso de pedidos a almacén.
Llamada e-1: Elija esta opción cuando el movimiento esté relacionado con pedidos para otro almacén o venta.
Llamada e-1-1: Es necesario que previamente esté seleccionadas las opciones “e” y “e-1”, permite elaborar pedidos de los recursos para venta.
Llamada e-2: Es necesario que previamente estén seleccionadas las opciones “e” y “e-1”, permite elaborar pedidos para transferencia, luego elija la opción “e-3” o “e-4”.
Llamada e-3: Es necesario que previamente estén seleccionadas las opciones “e” y “e-1”, permite elaborar pedidos para transferencia, luego elija la opción e-3 o e-4.
Llamada e-4: Es necesario que previamente estén seleccionadas las opciones “e” y “e-1”, permite elaborar pedidos para transferencia interna, luego elija la opción e-4, para elegir el almacén destino. La trasferencia interna es el traslado de recursos de un almacén a otro dentro del mismo proyecto.
Llamada e-5: Es necesario que previamente estén seleccionadas las opciones “e” y “e-1”, permite elaborar pedidos para transferencia externa, luego elija la opción e-3, para elegir el proyecto y almacén destino use las opciones “e-5” y “e-6”. La trasferencia externa es el traslado de recursos de un almacén a otro proyecto y almacén.
Llamada e-6: Es necesario que previamente estén seleccionadas las opciones “e” y “e-1”, permite elaborar pedidos para transferencia externa, luego elija la opción e-3, para elegir el proyecto y almacén destino use las opciones “e-5” y “e-6”. La trasferencia externa es el traslado de recursos de un almacén a otro proyecto y almacén.
Llamada f: Seleccione en el caso de pedidos a compras.
Llamada g: El estado es informativo.
Llamada h: Registre la fecha del pedido.
Llamada i: Seleccione la prioridad, sirve para agrupar la información por este concepto.
Llamada j: En el caso de haber seleccionado la opción marcada con la llamada “f”, elija el centro de compra que atenderá el pedido.
Llamada k: Esta celda en editable.
Llamada L: A este segmento se adicionan los recursos que forman parte del pedido.
Llamada L-1: Ingrese la cantidad a pedir, luego haga clic derecho e ingrese la fecha de entrega.
Llamada L-2: Estos datos son informativos.
Llamada L-3: Muestra la cantidad pedida anteriormente.
Después de ingresar los datos use el botón «Aceptar» y envíe el pedido a aprobación.
3.19.3 Pedidos automáticos para comprar
Utilice esta opción cuando el proyecto esté programado, y exista información en el escenario del cronograma por periodos dentro del Plan de Utilización de Recursos del escenario del Cronograma por periodos del módulos de Gerencia de proyectos, que proporciona información de los recursos con sus respectivas cantidades, para el proceso de pedido, además incluye los recursos equivalentes registrados en el escenario del Techo.
Todos los recursos tipo material que participan en la ejecución del proyecto, primero deben ser pedidos para comprar (pedido A Compras),
El que pide es el que maneja el módulo de Gerencia de proyectos y autoriza el Residente de obra.
LEAN. Administre recursos presupuestados.
El usuario que maneja el módulo de Gerencia de proyectos – equivale al Último planificador del PMI
El escenario de Pedidos Automáticos, está disponible en el grupo de Pedidos del Módulo de Gerencia de proyectos y en el de Compas y pedidos. Antes de efectuar los pedidos verifique que el proyecto debe estar asignado a un Centro de compra además la persona que pida y apruebe debe estar registrado.
Importante: Sí efectuó cambios en el presupuesto, es necesario que en el escenario de Actividades use el botón para tomar los cambios, luego en el escenario del Cronograma por periodos en modo programado metrado hacer clic derecho sobre una celda que tiene datos y del menú elegir Recalcular por escala; luego ingresar al escenario de Techo para que refresque los recursos y luego ya puede efectuar los pedidos automáticos.
Datos utilizados para efectuar el pedido automático:
Acceso al escenario:
Agrupar los recursos: El sistema acumula la cantidad de los recursos de acuerdo al número de semanas elegido.
Opciones para agrupar las cantidades:
Seleccionado: Cuando se encuentran seleccionadas las opciones a, b, c, y/o d, disminuyen las cantidades de los recursos programados, para pedir.
No seleccionado: Las cantidades de las opciones a, b, c, y/o d, no son considerados en el cálculo del saldo a pedir.
Selección de periodos
El sistema acumula las cantidades de acuerdo a la cantidad de periodos elegidos.
Para acumular las cantidades tenga en consideración la capacidad de almacenamiento.
3.19.3.1 Generar el pedido automático
Al efectuar un pedido tenga en cuenta lo siguiente: Procure agrupar recursos similares, ej: Un pedido para todo lo que compre en una ferretería, otro pedido para lo que compre en una farmacia, otro pedido para el proveedor de agregados, etc., esto permitirá generar orden de compra a partir de un pedido para todos los recursos que contiene el pedido.
Las cantidades que siguiere el sistema en el pedido tienen que ser editado por el usuario con la finalidad de redondearlo.
Al asignar una fecha de entrega por parte del proveedor al almacén o bodega, tome en cuenta lo siguiente:
Capacidad del medio de transporte, en cuantos viajes trasladará.
Capacidad de almacenamiento en obra o proyecto.
El sistema en forma automática genera un correo electrónico avisando al encargado de aprobar el pedido que tiene un documento pendiente de aprobar. Para usar la opción, el IP del servidor de correos debe ser ingresado usando el botón
Después de efectuar el pedido, las cantidades son mostradas en la columna “En pedidos Registrados”.
El encargado de aprobar el pedido, debe ingresar al S10 módulo de Compras y Pedidos, al escenario de Aprobación, para aprobarlo, con lo que concluye el pedido.
El siguiente paso es el proceso de compra, que se realiza en el módulo de Compras y pedidos.
Nota: Los pedidos automáticos son visualizados en el escenario de Pedidos, para eliminar un pedido tiene que estar en estado Registrado.
3.19.4 Aprobación
Todos los pedidos A Compra, deben ser aprobados para que pasen al proceso de compra, el usuario encargado de aprobar el pedido debe estar registrado y tener atributos.
Todos los pedidos deben aprobarse necesariamente para continuar con el trámite.
Acceso al escenario de Aprobación de Pedidos
Llamada a: Permite aprobar el pedido (todos los recursos que forma parte del pedido).
Llamada b: Permite aprobar el pedido de un recurso, total o parcialmente.
3.19.5 Techo presupuestado
Cuando en el presupuesto se consideran recursos genéricos, en los cuales aparecen como tope la cantidad y precio, mediante un utilitario le asignamos recursos específicos o del catálogo para el proceso de procura, tomando en cuenta dos parámetros:
Cantidad: Toma como techo la cantidad del recurso genérico presupuestado y va descontando a medida que van ingresando los recursos específicos o del catálogo.
Precio: Toma como techo el precio del recurso genérico presupuestado y va descontando a medida que van ingresando los recursos del catálogo.
Requisito: En el escenario de Datos Generales, el proyecto debe tener seleccionado TECHO PRESUPUESTADO.
Para el proyecto Puente:
3.19.5.1 Techo precio
Califique al recurso con el tipo de Techo Precio, para que el precio del recurso del catálogo sea descontado del recurso presupuestado. Este escenario está disponible desde los módulos de Gerencia de proyectos además de Compras y pedidos; algunos ejemplos:
Para el caso de nuestro ejemplo, en el presupuesto se consideró el recurso presupuestado, UTILIES DE OFICINA, pero necesitamos comprar cuadernos, lapiceros, que vienen a ser los recursos del catálogo.
Para usar el precio como techo es necesario calificar el recurso.
Acceso al escenario
Calificar el recurso con techo precio y registro de recursos específicos.
Efectúe esta labor antes de pedir los recursos específicos, para cada uno de los recursos genéricos.
Adicione recursos del catálogo,
Edite el techo precio de cada uno de los recursos,
Llamada a: Ingrese la cantidad a pedir.
Llamada b: Ingrese el precio del recurso.
Llamada c: Grabar el contenido de la ventana.
Repita el proceso para los demás recursos equivalentes.
Defina como pedible a los recursos adicionados del catálogo.
Llamada 1: Haga clic derecho sobre el recurso genérico y del menú que muestre elija No pedible.
Llamada 2: Haga clic derecho sobre cada uno de los recursos equivalentes y del menú que muestre elija Es pedible.
Haga que sólo los recursos específicos sean pedibles.
3.19.5.2 Techo por cantidad
Califique al recurso con el tipo de Techo Cantidad, para que el recurso específico sea descontado de la cantidad total presupuestada.
Este escenario está disponible desde los módulos de Gerencia de proyectos y de Compras y pedidos.
Efectúe esta labor antes de pedir los recursos específicos.
Acceso al escenario
Califique al recurso con techo cantidad,
Para adicionar los recursos específicos o equivalentes,
El sistema detecta que tiene recursos específicos o equivalentes y los muestra,
Registre los datos de los recursos equivalentes:
Llamada a:
Es el nombre del recurso genérico presupuestado.
Llamada b:
Son los recursos equivalentes, específicos o del catálogo que serán comprados.
Llamada c:
Es la unidad del recurso genérico.
Llamada d:
Es la unidad del recurso específico.
Llamada e:
Sí en el catálogo de recursos, se registró las equivalencias para el recurso equivalente o específico, el factor es tomado por defecto, en el caso de no tener es necesario registrarlo, ej.: el recurso genérico es por kilogramos y el recurso específico es por varillas, el ACERO CORRUGADO fy=4200 kg/cm2 GRADO 60 VARILLA DE ½” POR 9.00 m, pesa 9.18 kg, este valor se ingresa en la columna factor.
Llamada f:
Ingrese la cantidad de varillas de acero a comprar.
Llamada g:
Este dato es calculado por el sistema tomando el factor y la cantidad.
Llamada h:
Este dato es calculado por el sistema tomando en cuenta el precio presupuestado por kg.
Es necesario que en el catálogo de recursos, el recurso principal tenga registrado a los recursos equivalentes.
Llamada a:
Es el recurso principal.
Llamada a-1:
Es la unidad del recurso principal.
Llamada b:
Son los recursos equivalentes asignados al recurso principal.
Llamada b-1:
Es la unidad del recurso equivalente.
Llamada b-2:
El factor, es relacionado con la unidad del recurso principal, en este caso una varilla de ACERO CORRUGADO fy = 4200 kg/cm2 GRADO 60 VRILLA DE 3/8” equivale a 5.04 kg (en peso).
3.19.6 Techo Libre
Los proyectos calificados o definidos con techo libre, permite el registro manual de recursos que utilizará para efectos de realizar los pedidos de los recursos tipo material.
Al registrar el proyecto en el escenario de Datos generales, defina el proyecto con Techo Libre.
En el escenario de Techo Libre, registre los recursos tipo material y tiene estas alternativas:
• Del escenario de Precios por presupuesto del módulo de Presupuestos exporte a una hoja Excel, para luego importarlos en el escenario del Techo libre que se encuentra disponible en los módulos de Gerencia de proyecto, Compras y pedidos.
• Elegir del catálogo de recursos, para luego ingresar las cantidades y precio.
Utilice esta opción cuando el proyecto no esté planificado y tenga prisa en efectuar el pedido de los recursos.
Cuando efectúe pedidos, los recursos serán mostrados del Techo Libre.
3.19.7 Recursos por Pedido
Este escenario es informativo muestra los recursos y el pedido en el que se encuentra.
Acceso al escenario
Llamada a: En la ventana superior ubique un recurso.
Llamada b: Clic derecho sobre el recurso para mostrar el menú.
Llamada b-1: Muestra el pedido donde se encuentra el recurso.
Llamada b-2: Luego de aprobar el pedido el sistema genera un nuevo documento que puede ser para pedido de compra o pedido a almacén, esta opción muestra el documento generado.
Llamada b-3: Muestra la ventana se seguimiento del recurso, con el camino recorrido por el pedido.
3.19.8 Seguimiento de Recursos
Este escenario es informativo muestra la ventana se seguimiento del recurso, con el camino recorrido por el pedido.
Acceso al escenario
Llamada a: En la ventana superior ubique un recurso.
Llamada b: Muestra el número de pedidos en el que se encuentra el recurso.
Llamada b-1: Muestra el pedido donde se encuentra el recurso.
Llamada b-2: Muestra los documentos generados a partir del pedido.
Llamada b-3: Muestra la ventana se seguimiento del recurso, con el camino recorrido por el pedido.
Llamada c: Muestra el tipo de pedido.
3.19.9 Configuración
3.19.9.1 Partidas de Control
El sistema utiliza Partidas de Control, para el control económico del proyecto.
Las Partidas de Control son utilizadas para asignar a los recursos que egresan del almacén tipo material, también se utiliza en los tareos, partes de equipo y subcontratos.
Los montos de las Partidas de Control son tomados por el módulo de Gerencia de proyectos en los resultados operativos-
Sí en el módulo de Gerencia de proyectos, ya se efectuó el registro, la información es visualizada en el módulo de Almacenes.
S10: Es necesario responde a las preguntas, ¿en que gasta el proyecto?, ¿cómo quieres ver los resultados operativos?
PMI: Cada proyecto debe generar la información económica para el control.
3.19.9.2 Partidas de Control General a nivel de empresa
Es la relación de ítems en lo que generalmente gasta la empresa al ejecutar sus proyectos. A cada una de las Partidas de Control del proyecto se le asigna una Partida de Control General, esto permite controlar en una vista todos los proyectos que ejecuta la empresa.
Ejemplo de partidas de control general
Acceso al catálogo
Verifique la estructura de registro,
Registre de las partidas de Control,
3.19.9.3 Partidas de Control a nivel de proyecto
Las actividades de cada una de las fases o WBS, se asignan a las partidas de Control, con lo que adquieren valor y recursos. Ejemplo todas las partidas de encofrado serán agrupadas en una sola partida que se llamará Encofrado. Es como sí “armara” otra partida con todos los análisis de precios unitarios de encofrado.
En el módulo de Gerencia de proyectos, elija el escenario de Datos Generales, en el árbol elija el proyecto y configure la opción relacionada con las Partidas de Control.
Llamada a: Elija Datos adicionales.
Llamada b: Seleccione Mostrar solo Partidas de Control Activas, para que al momento de efectuar un egreso sólo permite elegir las partidas de control activas.
Ejemplo de partidas de Control en diferentes tipos de proyectos:
Acceso al escenario
Verifique la estructura de registro,
Registro de las Partidas de Control
Registro del segundo nivel,
Registre Costos Directos, luego repita el registro con Costos Indirectos.
Para el registro del último nivel,
Llamada a:
Ingrese el nombre de la Partida de Control
Llamada b:
Sí la Partida de Control tiene unidad ingrese (hay partidas de control que no tienen unidad).
Llamada c:
Del grupo de actividades asignadas a esta partida de
control,
elija la unidad representativa, todos las actividades que tengan la unidad de la partida de control será utilizado para las proyecciones de saldos y ratios.
Llamada d:
Seleccionado permite ser elegido para asignar a los recursos que se egresan.
No seleccionado, cuando la participación de la partida de control ha concluido, ya no aparece para ser elegido y asignado a una partida de control.
Llamada e: Del catálogo elija la Partida de Control General (a nivel de empresa).
3.19.9.4 Destinos Específicos
Registre cada lugar del proyecto, donde se ejecutan las actividades, esta información es utilizada por el módulo de almacenes como un destino para asignar a los egresos hacia una Partida de Control. Este destino permite conocer todos los recursos utilizados en cantidad y costo.
Sí en el módulo de Gerencia de proyectos, ya se efectuó el registro, la información es visualizada en el módulo de Almacenes.
Datos que serán registrados
Llamada a: Considere en forma gruesa cada nivel del edificio.
Llamada b: En acabados considere de cada nivel las zonas comunes y departamentos.
Llamada b: Considere a todo el personal de staff y obreros.
Llamada d: Considere uno a uno los puentes, pontones, alcantarillas, badenes etc.
Llamada e: Considere a todo el personal de staff y obreros.
Acceso al escenario
Verifique la estructura de registro,
3.19.9.5 Frentes
Es un lugar físico del proyecto donde se ejecutarán actividades. Permite organizar el proyecto por frentes. Se utiliza para asignar a los tareos, partes de equipo y egresos desde el módulo de almacenes.
Sí en el módulo de Gerencia de proyectos, ya se efectuó el registro, la información es visualizada en el módulo de Almacenes.
No todos los proyectos requieren el registro de frentes.
Datos que serán registrados,
Acceso al escenario
Verifique la estructura de registro,
Registro del frente
3.19.9.6 Roles por Unidad Operativa
Entiéndase por unidad operativa a cada una de las oficinas, en este caso cada una de las celdas que forman parte del organigrama funcional de la empresa. También unidad operativa es un proyecto específico. En resumen todo lo que se quiera controlar.
Los egresos del almacén tipo economato se realizan hacia una unidad operativa.
A continuación se muestra el organigrama funcional del proyecto.
Acceso al escenario
Definir estructura
Registro de las unidades operativas
Para registrar el último nivel,
Este es registro concluido
3.19.9.7 Ubicación física
Permite el registro de la ubicación física o lugar del almacén en un catálogo, este lugar será asignado posteriormente a un recurso.
El registro de la ubicación es por almacén.
En el escenario de Stock, es donde se asigna ubicación física al recurso.
En el caso de que no existan registros solicite al encargado de Gerencia de proyectos que los registre.
Se muestra como ejemplo un croquis de un almacén, el mismo que debe ser registrado en un catálogo,
Acceso al escenario
Verificar estructura
Registro de la ubicación física
Registre los demás niveles.
3.20 Informes
3.20.1 Órdenes de compra
Este escenario es informativo y permite ver sólo las órdenes de compra que están en calidad de emitidas, esto indica que los recursos comprados no llegaron o no fueron ingresados al almacén, también muestra las órdenes de compra que efectuaron solo ingresos parciales y tiene pendiente la entrega de recursos.
3.20.2 Pedidos pendientes
Este escenario es informativo, lista todos los PEDIDOS A ALMACEN realizados desde el módulo de ¨Pedidos¨, y que faltan ser atendidos.
3.20.3 Préstamos
Este escenario es informativo muestra todos los recursos que me prestaron y también los recursos que mi proyecto ha prestado basándose en los movimientos de ingresos y egresos por préstamos o devoluciones de préstamos. Tiene dos formas de ver este escenario: Lo que «Me prestaron» y «Lo que preste».
Visualizar los préstamos
3.20.4 Transferencias externas
Este escenario muestra todos los recursos transferidos basándose en los movimientos de ingresos y egresos por transferencias externas. Tiene dos formas de ver este escenario: «Lo que transferí» y «Lo que me transfirieron».
Visualizar las transferencias
3.20.5 Cronograma de entrega por rango
Lista los recursos que fueron pedidos desde el módulo de “Pedidos” con las fechas de entrega en un rango de fechas.
Sólo para los recursos cotizados a los cuales se les asignó proveedores, o cuando forma parte de una orden de compra, muestra los datos del proveedor.
3.20.6 Cronograma de entregas detallado
Lista los recursos que fueron pedidos desde el módulo de “Pedidos” con las fechas de entrega de cada uno de los recursos que el proveedor está obligado a entregar al almacén.
Use este escenario para reclamar al proveedor o indicar la fecha de entrega del recurso en el almacén.
3.20.7 Pedidos, OC, guías de almacén
Este escenario es informativo, lista todos los recursos asociados a un proyecto.
3.21 Mano de obra
3.21.1 Personal disponible
Del catálogo de Socio de Negocio que viene a ser una bolsa, se elige a todo el personal que trabajará en el proyecto, directamente por la empresa y a los que trabajarán para el subcontratista. Los costos de la mano de obra incluyen todos los beneficios incluso la liquidación, es lo que realmente nos cuesta un obrero, estos datos nos proporciona el planillero.
A continuación se muestra al personal obrero de casa elegido para la ejecución del proyecto.
Acceso al escenario de Personal Disponible
Llamada a: El costo del tareo que se efectúa en la rama Meta, pasan a los Resultados Operativos.
Llamada b: Solo sirve para el control del tareo.
Ubique la rama Meta en el árbol, y registre el personal obrero de la empresa.
Llamada a: Los tipos de nómina se registran en el módulo de Nóminas.
El sistema muestra el catálogo de personal de donde debe elegir a los obreros,
Luego en el escenario serán mostrados los obreros elegidos.
Ficha personal
Permite el registro de información que será utilizada para el proceso de los tareos del personal obrero, se compatibiliza con el recurso que se utilizó en presupuestos, la categoría que se utilizará para los pagos en planilla y los costos de hora hombre. El registro es para todo el personal disponible.
Para registrar la información,
Luego de menú elija FICHA – VER.
Registre los datos en la ficha.
Llamada 1: Seleccionado, el dato que se registre afecta a todos el personal seleccionado.
Llamada a: Elija la pestaña general.
Llamada a1: Permite elegir un recurso presupuestado.
Llamada a2: Cuando el recurso no está presupuestado, elija un recurso del catálogo.
Llamada a3: Se refiere a la categoría considerada en el presupuesto.
Llamada a4: Se refiere a la categoría considerada en la planilla o nómina de obreros.
Llamada a5: Elija la ocupación del obrero.
Llamada a6: Sí le asigna una partida de control, este será tomado como defecto al momento de elaborar los tareos.
Llamada a7: Esta celda debe estar seleccionada cuando el obrero está trabajando en el proyecto; sí deja de trabajar quite la selección para que ya no sea mostrado en los tareos y órdenes de trabajo.
Llamada a8: Muestra la ventana Socio de negocio, que permite editar o modificar los datos del obrero registrado.
Llamada a9: Muestra la ventana Recursos Humanos, que es informativa, el contenido se registra desde el menú Catálogos – Socios de Negocios donde se elige y registrar información relacionada al obrero.
Llamada a10: Muestra la ventana con el resumen del tareo o asistencia del obrero al proyecto, esta opción es informativa.
Llamada b: Elija esta pestaña para ingresar el tipo de moneda y precios.
Ingrese los precios que se utilizan en el ejemplo, (sólo son referenciales), el precio es el total que le cuesta a la empresa, incluye, salarios, bonificaciones, reintegros, liquidaciones etc.
Llamada b: Elija esta pestaña para elegir el tipo de moneda y precios.
Llamada b1: Elija el tipo de moneda con el que pagará al obrero.
Llamada b2: Es el precio de la hora hombre normal.
Llamada b3: Es el precio de la hora hombre extra al 60%.
Llamada b4: Es el precio de la hora hombre extra al 100%.
Llamada b5: Es el precio del recurso considerado en el presupuesto, es informativo
Llamada c: Luego de ingresar a la pestaña Nóminas, en el detalle (zona en blanco del documento), haga clic derecho y del Catálogo de Procesos, elija lo que corresponde.
Llamada d: En Complementario, ingrese la fecha de ingreso del obrero al proyecto y cuando se retire registre también la fecha.
Repita el proceso para registrar todo el personal obrero de la empresa en la rama Meta, y luego el personal obrero para el subcontrato.
3.21.2 Tareo
Consiste en registrar las horas trabajadas en una partida de control. Cuando no se utiliza las Ordenes de Trabajo en un período (semana), los tareos se pueden realizar manualmente con ayuda de Excel. No deben usar las dos formas en el mismo período.
Registre en el servidor una carpeta donde guardará los archivos en Excel que contienen los tareos.
A la empresa que ejecuta el proyecto se le debe asignar el formato de Excel “Tareos de mano de obra”, esta labor se realiza sólo la primera vez.
Cierre el Catálogo de Formatos por Empresa
Estos son los datos que serán utilizados en la semana 17-2010
De miércoles a viernes ingrese 8.5 horas tipo normal en la partida de control Preparación de agregados.
Para el día sábado 24-04 considerar 5.5 horas tipo normal en la partida de control Preparación de agregados.
Estos son los datos que serán utilizados en la semana 18-2010
De lunes a viernes ingrese el número de horas trabajadas: 8.5 horas tipo normal en la partida de control indicada. El día sábado 01-05-20101 no es laborable, no ingrese datos.
Acceso al escenario del Tareo
Exportar el Tareo del S10 a Excel
Ahora el Tareo será mostrado en una hoja Excel.
Registre el tareo para cada uno de los obreros.
Elija el tipo de hora
Asigne Partida de Control al Tareo
El formato de Excel, toma las Partidas de Control del proyecto para elegirlo; obligatoriamente se debe elegir la Partida de Control, sí un obrero realiza más de una actividad durante el día, utilice una nueva rotación para cada Partida de Control de la misma forma para el sobretiempo.
Después de registrar los tareos use el botón
para grabar el contenido de la ventana y cierre Excel.
Importar el Tareo de Excel al S10
Calcular Resumen periodo, use el botón
,
Llamada a: Utiliza los precios registrados en “Propiedades” del obrero.
Llamada b: Cuando cambia el precio de la mano de obra en el semana, sólo por única vez procese el Tareo seleccionado en “Actuales”, para las siguientes semanas, retorne la selección a “Históricos”.
3.22 Equipo
3.22.1 Equipos disponibles
El equipo propio y de terceros que se utilizará en el proyecto se registra para:
• Al hacer el egreso de un recurso, se asigna como destino un equipo.
• Registro de los partes de equipo, se refiere al registro de las horas trabajadas en una partida de control, información que es aprovechada en el módulo de Gerencia de proyectos.
Sí en el módulo de Gerencia de proyectos, ya se efectuó el registro, la información es visualizada en el módulo de Almacenes.
3.22.1.1 Registro de Equipo, como activo en el catálogo de recursos
Los recursos tipo activo que participan en la ejecución del proyecto, sean propios o de terceros, deben ser registrados en el catálogo de recursos.
Datos que serán registrados:
Acceso al Catálogo de Recursos
En el árbol del Catálogo de Recursos,
Llamada a: En la ventana de registro del recurso específico, registre el nombre completo del recurso.
Llamada b: En la ventana de registro del recurso específico, toma el nombre del recurso genérico para luego incrementarle el nombre completo del recursso especifico.
Repita el proceso para regstrar al otro equipo y luego use el botón
«Modificar»
para grabar la información y cerrar la ventana. El recurso genérico se caracteriza porque tiene 10 dígitos. El recurso específico se caracteriza porque tiene 14 dígitos y este es el que se utiliza como activo.
3.22.1.2 Selección de Equipos Disponibles
Estos son los datos que serán elegidos.
Acceso al escenario Activos Disponibles
El sistema muestra el Catálogo de Recursos, (donde previamente se tienen registrados) todos los equipos propios y de terceros que se usará en la construcción del proyecto,
Los equipos elegidos son mostrados en el escenario “Equipos disponibles”,
Ficha del Equipo
A los equipos que participan en la ejecución del proyecto, se relacionará con el equipo presupuestado, además se le ingresa los precios y datos solicitados.
Para ingresar las propiedades a los equipos:
Use en el orden que se marca en las llamadas.
Llamada a: Muestra el nombre del equipo disponible.
Llamada a-1: Seleccione, para que permita elegir el recurso presupuestado.
Llamada a-2: Elija el recurso presupuestado.
Llamada a-3: Muestra el monto máximo a pagar por el alquiler del equipo, este dato es tomado del presupuesto.
Llamada a-4: Elija el tipo de moneda.
Llamada a-5: De acuerdo a lo sugerido por la llamada “a-3”, ingrese el monto del alquiler.
Llamada a-6: Graba de la ventana.
Llamada e: Muestra la cantidad de hm presupuestadas, de esta se asigna al equipo lo que utilizará, cuando registre el subcontrato para el control de pagos tomará como tope.
Llamada e-1: Registre la cantidad de hm que trabajará el equipo en el proyecto.
Llamada e-2: Es el saldo por asignar de la cantidad de hm presupuestadas.
Llamada f:
Seleccionado: Cuando el costo del equipo registrado no incluye combustibles, del almacén de materiales se egresa el combustible, se asigna partida de control y en destinos específicos se elige el equipo.
No seleccionado: El costo del equipo registrado no toma en cuenta los combustibles.
Llamada g:
Seleccionado: La gestión del equipo es tomado en cuenta para ver los resultado de la operación del equipo.
No seleccionado: La gestión del equipo no es tomado en cuenta para ver los resultado de la operación del equipo.
Llamada h: Muestra la ventana Especificaciones del Equipo;
Llamada h: Muestra la ventana Especificaciones del Equipo;
Llamada h-1: Permite elegir el estado de la Fecha del Equipo
En elaboración
Revisado
Llamada h-2: Es el nombre de equipo registrado en el catálogo de recursos dentro del grupo de activos.
Llamada h-3: Es el nombre alterno del equipo registrado en el catálogo de recursos dentro del grupo de activos.
Llamada h-4: Son los datos relacionados al equipo.
Llamada h-4-1: Elija la unidad, entre pza o und.
Llamada h-4-2: Elija la unidad de medición del trabajo del equipo, de acuerdo a este dato defina las celdas de las llamadas h-5 y h-6.
Llamada h-4-3: Ingrese la marca del equipo.
Llamada h-4-4: Ingrese el modelo del equipo.
Llamada h-4-5: Ingrese la potencia del equipo.
Llamada h-4-6: Ingrese la potencia del equipo.
Llamada h-4-7: Elija la unidad para la potencia.
Llamada h-4-8: Ingrese el número de serie del equipo.
Llamada h-5: Seleccionado: En el módulo de Almacenes al efectuar el egreso del combustible, se elige la partida de control y luego al elegir el equipo, solicita el ingreso del Horómetro. Relacionado con la llamada h-4-2.
No seleccionado: Ignora el ingreso de horas trabajadas del equipo.
Llamada h-5: Seleccionado: En el módulo de Almacenes al efectuar el egreso del combustible, se elige la partida de control y luego el equipo donde se registra el Kilometraje. Relacionado con la llamada h-4-2.
No seleccionado: Ignora el ingreso de Kilometraje.
Llamada h-7: Son los recursos que se registran para una unidad de trabajo.
Llamada h-7-1: Elija del catálogo a los recursos que se tomarán en cuenta en la evaluación de los ratios, se refiere a la cantidad de recursos que es necesario por trabajo sean estas horas o kilómetros.
Llamada h-8: Elija del catálogo a los recursos que se tomarán en cuenta en la evaluación de los ratios, se refiere a la cantidad de recursos que es necesario un una hora de trabajo del equipo.
Llamada h-9: Elija del catálogo a los recursos que se tomarán en cuenta en la evaluación de los ratios, se refiere a la cantidad de recursos que es necesario un kilómetro de trabajo del equipo.
Continuamos con las llamadas de la ventana Ficha Equipo:
Llamada i: Muchas veces el equipo que necesitamos debe producir 1,000 m3 al día, pero en el mercado encontramos solo equipos que producen 500 m3.
Use factor 1 cuando el equipo que encontramos en el mercado rinde igual al previsto en el presupuesto.
Use factor 2 cuando el equipo que encontramos en el mercado rinde la mitad de lo previsto en el presupuesto.
Llamada j: Esta opción es informativa muestra los partes de equipo.
Llamada k: Muestra las valorizaciones donde participa el equipo, esta opción es informativa.
Llamada L: Seleccionado Use esta opción cuando el equipo esté en servicio o trabajando, sólo los equipos disponibles son mostrados para efectuar los partes diarios de equipo.
No seleccionado: Use esta opción cuando el equipo deje de trabajar o lo retiren del proyecto
Llamada m: Invoca la ventana donde se registran las notas.
Llamada n: Es una etiqueta que muestra el estado de la ficha.
Color amarillo: Es cuando la ficha está en proceso de elaboración
Color rojo: Es cuando la ficha del equipo está aprobada y puede ser utilizada en un subcontrato para administrar el trabajo y sus costos.
Llamada o: Muestra los resultados de la operación del equipo, esta opción es informativa.
Llamada p: Muestra el detalle de recurso (es la estructura de costos del recurso o un análisis de precios unitarios), esta opción es informativa.
3.22.2 Partes de Equipo
El trabajo diario del equipo debe ser registrado en el sistema con la finalidad de controlar las horas trabajadas y tener como resultado el costo por la participación del equipo en cada una de las partidas de control. Para el registro de las horas del equipo se utilizará una hoja Excel.
A la empresa que ejecuta el proyecto se le debe asignar el formato de Excel “Parte de Equipo”, Esta labor se realiza una sola vez.
A la pregunta que efectúe el sistema, responda eligiendo la opción Sí. Luego cierre el Catálogo de Formatos por Empresa.
Antes de exportar los partes diarios a Excel, en su PC cree una carpeta “Partes de Equipo PC” donde se almacenará, los archivos.
Datos para el registro:
Acceso al escenario de Partes de Equipo
Exportar a Excel el Parte de Equipo
Elija la carpeta donde se guardarán los Partes de Equipo
Asignar Partida de Control al Parte de Equipo
Ingreso de la cantidad de horas trabajadas
Las horas extras registre en otra rotación. Luego grabe el contenido de la ventana y cierre Excel.
Importar Partes de Equipo de Excel al S10
Para Calcular Resumen Periodo, use el botón . Después del proceso los datos pasan a resultados operativos.
Llamada a: Utilice los precios del histórico para un proceso normal de partes diarios de equipo.
Llamada b: Utilice los precios actuales sólo en la presente semana en el caso de haber algún cambio.
3.22.3 Resumen por Partida de Control
Use el botón
para visualizar el resultado
Llamadas de la ventana Resumen por Partida de Control
Llamada a: Es el monto semanal del trabajo del equipo que será cargado a los Resultados Operativos.
Llamada b: Este botón muestra el menú para visualizar los recursos.
Llamada b-1: En la ventana inferior muestra el nombre de los equipos, el detalle en cantidad y costo de su participación en la semana.
Llamada b-2: En la ventana inferior agrupa por recursos.
Llamada c: Exporta el contenido de la ventana a una hoja Excel.
3.23 General
Este grupo de escenarios es informativo agrupa información de los proyectos que ejecuta la empresa.
3.23.1 Ingresos y Egresos
Muestra información de los movimientos de ingresos y egresos de todos los proyectos que maneja la empresa.
Acceso al escenario Ingresos y Egresos
Luego,
Llamada a: Seleccionado muestra todos los datos desde el inicio de los movimientos hasta la fecha de consulta, luego use el botón de la llamada b para procesar.
No seleccionado, permite usar el rango de fechas marcado con la llamada “c”.
Llamada b: Procesa la información.
Llamada c: En el caso de no usar la opción de la llamada “a”, ingrese el rango de fechas para la consulta de datos, luego use el botón de la llamada “b” para procesar.
Llamada d: Muestra la ventana de filtros.
Llamada e: Muestra o esconde la celda de agrupamiento de la cabecera de las columnas.
Llamada e-1: Muestra los campos agrupados.
Llamada f: Muestra los campos disponibles en el escenario con la finalidad de mostrarlos u ocultarlos además cambiar de ubicación.
3.23.2 Stock General
Muestra información del stock de los almacenes de los que maneja la empresa.
Acceso al escenario Stock General
Luego,
Llamada a: Muestra la ventana de filtros.
Llamada b: Muestra o esconde la celda de agrupamiento de la cabecera de las columnas.
Llamada b-1: Muestra los campos agrupados.
Llamada c: Muestra los campos disponibles en el escenario con la finalidad de mostrarlos u ocultarlos además cambiar de ubicación
Llamada d: Graba el agrupamiento marcado con la llamada b-1.
3.23.3 Equipos Disponibles
Muestra la relación de los equipos utilizados en cada uno de los proyectos que ejecuta la empresa.
3.23.4 Aprobaciones
Este escenario muestra solo los documentos que falta aprobar o parcialmente aprobado, de todos los proyectos que administra la empresa, con la finalidad de aprobarlos. Este escenario está disponible para el usuario que tiene los atributos de aprobación.
Llamada a: Muestra solo los documentos por aprobar.
Llamada a-1: Haga doble clic, para mostrar los pedidos a aprobar.
Llamada a: Use este botón para hacer visible los campos marcados con la llamada a-1.
Llamada b: Esta ventana muestra todos los pedidos pendientes de aprobación.
Llamada b-1: Muestra el tipo de pedido.
Llamada b-2: Muestra el estado de pedido.
Llamada b-3: Del pedido elegido, al hacer doble clic sobre el pedido muestra la ventana con el detalle del pedido, al hacer clic derecho muestra la opción de aprobación.
Llamada c: Muestra el detalle del pedido marcado con la llamada b-3.
3.24 Utilitarios
3.24.1 Mantenimiento de usuarios
Los usuarios que necesitan acceso al módulo, deben estar registrados previamente en el catálogo de Socio de Negocio por Tipo.
Al usuario se le adiciona, en el escenario de mantenimiento de Usuarios, donde se le asigna una categoría y permisos.
Datos que serán registrados
Acceso al escenario
Registre el nombre de la empresa o constructora
Registro de Subgrupos: Registre a Personal Técnico y Personal Administrativo
Registro de usuarios
Llamada a: El sistema permite tener hasta 6 usuarios para auditar la base de datos.
Asigne el color azul: Al usuario que auditará el catálogo de socio de negocio
Asigne el color guinda: Al usuario que auditará el catálogo de recursos
Asigne el color amarillo: Al usuario que auditará el catálogo de unidades
Asigne el color negro: Al usuario que auditará el catálogo de tipos de documentos.
Asigne el color verde: Al usuario que auditará el catálogo de moneda + formas de pago + partidas de control general.
Asigne el color rojo: Al usuario que auditará el catálogo de métodos de pago + motivo de pago + configuración.
Llamada b: Ponga una selección para el encargado de informática de su empresa, permite registrar usuarios asignar categoría y permisos, tiene limitaciones para usar los módulos.
Acceso al módulo
Al usuario que maneja, el módulo así como los que piden y aprueban deben tener asignado una categoría y dar atributos de uso.
En el árbol se muestra a los usuarios:
Asignar categoría al usuario
En el árbol elija un usuario, luego,
Llamada a: Elija la pestaña Módulos.
Llamada b: Use este botón para mostrar las categorías.
Llamada b-1: Administrador de sistemas: Tiene acceso total a los módulos del S10, es el encargado de registrar a los usuarios y asignarles atributos de uso del sistema de acuerdo a su jerarquía y al funcionamiento dentro de la organización.
Llamada b-2: Jefe de grupo: Tiene atribuciones de uso y otorgar permisos al resto de usuario del módulo, hay opciones que son exclusivas del jefe de grupo.
Llamada b-3: Especialista: Tiene atributos de uso del módulo, tiene acceso a los escenarios que se le encargue.
Llamada b-4: Usuario:
Invitado: Es el usuario que sólo puede ver los datos de los escenarios que se le asigne no permitiendo efectuar ninguna otra labor.
Acceso a los escenarios
Para el usuario elegido en el árbol, desde esta ventana se le asigna accesos a los diferentes escenarios, catálogos o funciones que podría cumplir este usuario según a las responsabilidades asignadas.
Llamada a: Elija la pestaña Escenarios.
Llamada b: Muestra la lista de escenarios disponibles.
Llamada c: Clic derecho en este segmento y muestra el menú para dar acceso.
Acceso a los catálogos
Llamada a: Elija la pestaña Catálogos.
Llamada b: Muestra la lista de Catálogos disponibles.
Llamada c: Clic derecho en este segmento y muestra el menú para dar acceso.
Llamada d: Seleccionado, permite efectuar nuevos registros en el catálogo.
Llamada e: Seleccionado, permite modificar registros en el catálogo.
Llamada f: Seleccionado, eliminar registros en el catálogo siempre y cuando no estén participando en una aplicación.
Llamada g: Seleccionado, permite adicionar notas a los registros en el catálogo.
Llamada h: Seleccionado, permite imprimir registro del catálogo.
Llamada i: Seleccione el catálogo que auditará el usuario.
Sólo la persona que registre el proyecto tendrá acceso a este, pero puede invitar a otros usuarios para que compartan la información del proyecto.
Atributos especiales
Llamada a: Elija la pestaña Atributos especiales.
Llamada b: Seleccionado permite aprobar especificaciones de los equipos.
Llamada c: No disponible en esta versión.
3.24.2 Actualización de Base de Datos
Es un utilitario que adecúa la base de datos de acuerdo a la nueva estructura, es una labor que debe realizar el “sa” cuando el sistema envíe un aviso.
3.24.3 Mantenimiento de Base de Datos
Es un utilitario, destinado a dar mantenimiento a las bases de datos, son labores que debe realizar el “sa”.
Llamada a: Es un utilitario que permite sacar copia de seguridad de la base de datos, lo almacena en C:\s10200\backup con el nombre que usuario registre.
Llamada b: Es un utilitario que permite restaurar una copia de seguridad que se encuentra almacenada en C:\s10200\backup con el nombre que usuario registre.
Llamada c: Es un utilitario que permite crear una base de datos vacía, el usuario debe registrar un nombre diferente a los existentes.
Llamada d: Es un utilitario que permite eliminar una base de datos existente, esto ocurre cuando se tiene varia bases de datos. La limitación es que no se puede eliminar la base de datos en uso.
Llamada e: La base se datos está compuesta por 3 archivos , , , que se ubican en C:\s10200\data. Es un utilitario recupera la base de datos a partir esos archivos.
Llamada f: Es un utilitario que permite compactar u ordenar la base de datos para que ocupe menos espacio.
Llamada g: Es un utilitario que permite re indexar u ordenar la base de datos, con la finalidad de dar mayor velocidad a los procesos.
Llamada h: Es un utilitario que permite crear tareas para administrar la base se datLos en forma automática.
Configuración por País
Permite registrar los nombres de las etiquetas, para el país donde utilizarán el S10, tomando como base la información del Perú.
Datos que serán registrados
Registro de la información:
Luego vaya a INICIO – Panel de Control – Configuración Regional – Elija el país y la moneda.
3.25 Descripción de los escenarios y menús
3.25.1 Grupo inicio
3.25.1.1 Datos generales
Permite el registro de atributos de uso dentro del proyecto, el usuario que tiene la categoría de Jefe de Grupo es el que asigna.
Llamada a: Permite el registro de un nuevo proyecto.
Llamada b: Duplica el proyecto y todos sus movimientos.
Llamada c: Envía el proyecto a la papelera de reciclaje.
Llamada d: Permite el registro de un nuevo almacén, el usuario JEFE DE GRUPO debe tener atributos para hacerlo.
Llamada e: Muestra el notero con la finalidad de registrar información relacionada al proyecto.
Llamada f: Cambia la ubicación del proyecto a otra carpeta del árbol.
Llamada g: Refresca o vuelve a tomar los datos del árbol.
Llamada h: Define el alcance del proyecto.
Llamada h-1: Tienen acceso al proyecto, el usuario que lo registró y los que están registrados dentro de Usuarios del proyecto.
Llamada h-2: Tiene acceso al proyecto todos los usuarios que tienen acceso al módulo de Gerencia de proyecto, no importando la categoría.
Llamada h-3: En implementación.
Llamada i: Permite el registro de usuarios al proyecto y dar atributos de uso.
Llamada j: Muestra el submenú de utilitarios.
Llamada j-1: Opciones para las guías de almacén.
Llamada j-1-1: Permite reasignar o cambiar una partida de control por otra en las guías de egreso.
Llamada j-1-1: Permite reasignar o cambiar una partida de control por otra en las guías de egreso de acuerdo al código alterno del recurso.
Llamada j-2: Por defecto el día de inicio de la semana es el lunes, este utilitario permite cambiar el día de inicio a otro día.
Llamada j-3: Es un utilitario que traslada las guías de ingreso y egreso de un almacén a otro, el almacén destino no debe tener movimientos el almacén destino.
Llamada j-4: Cuando los almacenes del proyecto no están conectados a una red, este utilitario exporta las guías de los movimientos del almacén en un rango de fechas, a un archivo. Con la finalidad de trasladar información.
Llamada j-5: Cuando usó la opción marcada con la llamada “j-4”, importa el contenido del archivo a la base de datos.
Llamada k: Es un utilitario que protege la información registrada en el mes, el usuario que cierre el mes debe tener atributos para hacerlo.
Llamada L: Muestra los datos del registro del proyecto.

El texto completo también está troceado en `kb/fragmentos.jsonl` y se cita con el nombre del manual y la página.

**Confiabilidad:** oficial S10
