---
title: Manual de Presupuestos
summary: Este es un programa para elaborar presupuestos de todo tipo de obras a
  partir de los metrados. Permite elaborar hasta tres tipos de presupuesto por
  obra, el Venta, Meta y Línea Base, los que son asignados a los proyectos que
  serán utilizados para planificar, ejecutar, controlar…
tags:
  - oficial-s10
  - manual
  - portal-miembros
  - presupuestos
id: 01M3SP28QPSF625TPD9ADJCHFQ
status: active
updated_by: ai-agent
updated_at: 2026-09-30T17:48:00.166Z
---

Manual oficial de S10, bajado del portal de miembros con la cuenta de la empresa.

- **Página:** https://documentacion.s10peru.com/manual-de-presupuestos/
- **Módulo:** `presupuestos` (ver esa rama para el resumen redactado)

## Contenido

- 1 PRESUPUESTOS
- 1.1 Introducción
- 1.2 Tipos de presupuestos
- 1.3 El presupuesto
- 1.3.1 Presupuesto de Edificaciones
- 1.3.2 Presupuesto de un Centro Comercial
- 1.3.3 Presupuesto de Saneamiento
- 1.3.4 Presupuesto de una carretera venta y meta
- 1.3.5 Presupuesto de Instalaciones Eléctricas venta y meta + GG
- 1.3.6 Presupuesto de Edificaciones venta y meta + GG
- 1.3.7 Presupuesto Mecánico venta y meta + GG
- 1.3.8 Presupuesto de un puente venta y meta
- 1.4 Proyecto utilizado para desarrollar el presupuesto
- 1.4.1 Metrados del proyecto para elaborar el presupuesto
- 1.5 Secuencia de uso del módulo de presupuestos
- 1.6 Partes de la hoja del presupuesto
- 2 Generalidades
- 2.1 La carpeta de trabajo
- 2.2 Configuración
- 3 Uso del módulo de presupuestos
- 3.1 Formas de búsqueda en los catálogos
- 3.1.1 Búsqueda global
- 3.1.2 Búsqueda local
- 3.2 Registro de información en el catálogo de presupuestos
- 3.3 Registro del presupuesto

## Texto de la página

# Manual de Presupuestos

Fuente: https://documentacion.s10peru.com/manual-de-presupuestos/

Table of Contents
Toggle
1 PRESUPUESTOS
1.1 Introducción
Este es un programa para elaborar presupuestos de todo tipo de obras a partir de los metrados. Permite elaborar hasta tres tipos de presupuesto por obra, el Venta, Meta y Línea Base, los que son asignados a los proyectos que serán utilizados para planificar, ejecutar, controlar y valorizar labores que se realizan en el módulo de Gerencia de proyectos del S10.
Las facilidades implementadas como la interacción con el Office de Microsoft, permite que la información sea aprovechada e integrada a otros programas de aplicación. Esta guía puede ser utilizada para el uso del módulo de presupuestos de la versión estándar (profesional, empresarial y ERP), sólo la versión ERP dispone del grupo de planeamiento. Es recomendable que tenga instalado el MS Project (con licencia), para la parte de planeamiento.
Para tomar en cuenta:
El S10 al momento de instalar el programa en su PC, también instala el motor de la base de datos del SQL Server Express que es de libre disponibilidad:
La versión 2005, está diseñado para soportar un máximo de 4 GB (tamaño de la base de datos).
La versión 2008 r2, está diseñado para soportar un máximo de 10 GB (tamaño de la base de datos).
Si su base de datos crece más de lo indicado es necesario que disponga de licencias SQL Server.
1.2 Tipos de presupuestos
1.3 El presupuesto
El presupuesto proporciona al proyecto las partidas, metrados, recursos, cantidades y precios; siendo el inicio de la gestión del proyecto. En el caso del presupuesto meta, los recursos son pedidos, cotizados, comprados, almacenados y distribuidos durante la ejecución.
Cuando elabore el presupuesto meta, tenga presente que planificará y controlará, además:
• Las partidas del presupuesto meta obedecen a la realidad constructiva.
• No necesariamente contendrá las mismas partidas del presupuesto venta.
1.3.1 Presupuesto de Edificaciones
Llamada a:
En el subpresupuesto de instalaciones eléctricas y sanitarias hay diferencias marcadas en las partidas, como se muestra a continuación:
Llamada a-1a:
Las partidas del presupuesto venta es por puntos, con la finalidad de estimar el costo.
Llamada a-2a:
Las partidas del presupuesto meta, lo pueden elaborar por niveles, otra alternativa es registrar dentro del análisis de precios unitarios, en una subpartida todos los recursos tipo material o consumible necesario, en otra subpartida la relación de los subcontratos tipo mano de obra que serán valorizados al subcontratista.
Al elaborar el presupuesto tome en cuenta que será programado; en el S10 sólo se programan las partidas o las subpartidas, no los recursos.
Llamada b:
En el presupuesto venta, tiene diseño del pie de presupuesto, donde se considera los gastos generales, utilidad, impuestos.
En el presupuesto meta, no tiene pie de presupuesto, tiene un subpresupuesto detallado de gastos generales.
Llamada c:
En meta elabore un presupuesto de mayores metrados, considere todas las partidas que fueron omitidas u olvidadas y que serán financiadas por el contratista.
Llamada d:
En el presupuesto venta, elabore el presupuesto con las partidas adicionales ordenadas por el propietario del proyecto, las que serán valorizadas.
En el presupuesto meta, elabore el presupuesto con las partidas adicionales ordenadas por el propietario del proyecto.
Análisis de precios unitarios, en el presupuesto venta y meta.
Llamada e:
En el presupuesto venta, en el análisis de precios unitarios la mano de obra por hh, los materiales son considerados simplemente para asumir un costo aproximado, incluye las herramientas menores y manuales.
En el presupuesto meta, en el análisis de precios unitarios la mano de obra es un subcontrato cotizado por nivel, los materiales son los que realmente se utilizarán en el nivel los que serán pedidos y comprados. Las herramientas menores y manuales son consideradas en el subpresupuesto de gastos generales.
Llamada e-1-a:
La unidad del análisis de precios unitario en la venta es por punto.
Llamada e-1-b:
La unidad del análisis de precios unitario en la meta es global.
1.3.2 Presupuesto de un Centro Comercial
En el presupuesto meta: Sectorizar o fasear el proyecto; registre un subpresupuesto parada cada fase.
En el presupuesto meta: Sectorizar o fasear el proyecto; registre un subpresupuesto parada cada fase.
1.3.3 Presupuesto de Saneamiento
Llamada a: Asuma el presupuesto elaborado por institución que licita.
Llamada b: Para el presupuesto meta, Sectorizar o fasear el proyecto y elabore dentro de un subpresupuesto lo que ejecutará cada uno de los subcontratistas, esto facilita la planificación, el control y las valorizaciones de cada uno de los subcontratistas.
1.3.4 Presupuesto de una carretera venta y meta
Llamada a: Asuma el presupuesto venta, elaborado por la institución que licita.
Llamada b: Para el presupuesto meta, sectorizar o fasear el proyecto y elabore dentro de un subpresupuesto lo que ejecutará en cada uno de los tramos. Después de efectuar el movimiento masivo de tierras del tramo considere en una partida un puente, una alcantarilla, un muro de contención, con metrados reales.
1.3.5 Presupuesto de Instalaciones Eléctricas venta y meta + GG
1.3.6 Presupuesto de Edificaciones venta y meta + GG
1.3.7 Presupuesto Mecánico venta y meta + GG
1.3.8 Presupuesto de un puente venta y meta
Llamada a:
Use el presupuesto con el que ganó la licitación.
Llamada a-1:
En el análisis de precios unitarios la mano de obra es genérica, la arena y piedra chancada es comprada, se considera el equipo menor.
Llamada b:
Para el presupuesto meta, considere las partidas bastante descriptivas y para cada elemento.
Llamada b-1:
En el análisis de precios unitarios la mano de obra es específica pensando en el control, se consideran las subpartidas para la elaboración de agregado, no tomar en cuenta el equipo menor.
Llamada c:
Listado del equipo presupuestado.
Llamada c-1:
Muestra el equipo considerado en el presupuesto venta.
Llamada c-2:
En el presupuesto meta solo se considera el equipo mayor del que se efectuará partes diarios de equipo.
1.4 Proyecto utilizado para desarrollar el presupuesto
Con la finalidad de hacer un seguimiento a las opciones del programa y elaborar un presupuesto se desarrolló el croquis de un puente, del que elaboraremos el presupuesto meta. Y para elaborar un presupuesto venta usaremos una edificación LA CASA DE JUAN.
Para el presupuesto meta: El proyecto consta de dos estribos, tres vigas en arco y viga + losa. Desde ahora ya nos preparamos también para la planificación y el control que lo desarrollaremos en planeamiento de Presupuestos y en el módulo de Gerencia de proyectos.
1.4.1 Metrados del proyecto para elaborar el presupuesto
Para elaborar el presupuesto es necesario tener los metrados.
1.5 Secuencia de uso del módulo de presupuestos
1.6 Partes de la hoja del presupuesto
2 Generalidades
2.1 La carpeta de trabajo
Las ventanas que presenta el S10 se asemejan a un escritorio donde tiene todos los documentos necesarios para trabajar en forma ordenada. El entorno de trabajo es el mismo casi en todos los escenarios, a continuación, se describe las partes de un escenario:
Llamada a:
Estos botones son comunes para todos los escenarios;
Imprimir:
Envía el reporte directamente a la impresora previamente definida en Windows.
Vista preliminar:
Muestra el reporte en pantalla, de donde se puede exportar al formato que elija el usuario como Word, Excel, etc.
Retroceder:
Desliza el cursor en el árbol de un presupuesto a otro hacia arriba.
Avanzar:
Desliza el cursor en el árbol de un presupuesto a otro hacia abajo.
Lista:
Muestra o esconde el árbol ocupando su lugar por el escenario actual o activo.
Notas:
Activa el notero, donde permite registrar notas sobre la aplicación, acepta textos, insertar archivos de aplicaciones diferentes como: Word, Excel, AutoCAD. MS Project etc. Disponible para todos los registros, partidas, actividades, calendarios, recursos, etc.
Llamada b:
Son botones exclusivos del escenario donde se ingresa a trabajar.
Llamada c:
Muestra el título de la aplicación, nombre del programa que se viene utilizando, presupuesto elegido por el usuario, así como los botones minimizar , maximizar y cerrar .
Llamada d:
Presenta opciones de menú. Cada menú contiene acciones específicas y estas cambian de acuerdo al escenario que ingresa el usuario, las mismas opciones en forma abreviada se ejecutan desde los botones (iconos).
Llamada d-1:
Opciones del menú Archivo;
Archivo:
Muestra el menú con las siguientes opciones:
Imprimir:
Imprime el contenido del escenario.
Vista preliminar:
Muestra la apariencia en pantalla de un archivo como será impreso.
Iniciar sesión como usuario distinto: Permite reiniciar la aplicación sin salir del módulo de presupuestos, para ingresar como otro usuario. La misma acción se logra sí hace doble clic sobre la carita
mostrada en esta barra de estado. El sistema mostrará el cuadro de diálogo de acceso al sistema, donde puede acceder al sistema como otro usuario también permite el uso de otra base de datos.
Salir:
Sale de la aplicación.
Llamada d-2:
Opciones del menú Ver;
Punteros: Utilice para ubicar información dentro de la hoja del presupuesto, el sistema indicará en una celda con una selección (check), en los registros ubicados.
Presentación: Está disponible para el árbol, sirve para agrupar los presupuestos de acuerdo a:
Árbol:
Muestra o esconde el árbol, equivale al uso del botón .
Llamada d-3:
Opciones del menú Hoja del presupuesto, disponible sólo cuando el cursor está ubicado sobre un subpresupuesto en el árbol del escenario de la hoja del presupuesto;
Llamada d-3-1:
Genera los ítems a las partidas, de acuerdo a la subordinación que existe en la hoja del presupuesto y de acuerdo a la configuración previamente realizada en el menú herramientas – definir ítems.
Llamada d-3-2:
Es un utilitario que permite recodificar los ítems.
Llamada d-3-3:
Después de marcar un bloque de las partidas de un subpresupuesto, este utilitario permite asignar un metrado al bloque o efectuar operaciones con los metrados.
Llamada d-3-4:
Procesa el contenido del subpresupuesto.
Llamada d-3-5:
Es una opción informativa, muestra el resultado del último proceso.
Llamada d-3-6:
Al calcular el presupuesto, toma los datos del diseño del pie de presupuesto para cada partida.
Llamada d-3-7:
En el escenario de datos generales, se encuentra la opción de datos adicionales, donde se registra el factor de cambio, el sistema toma este dato para procesar el presupuesto. Es necesario ingresar a configuración para elegir la opción con factor de cambio, debido a que el resultado solo se aprecia en el reporte.
Llamada d-3-8:
Muestra la ventana del resultado del procesamiento del presupuesto.
Llamada d-3-9:
Muestra los recursos y precios que utiliza el presupuesto.
Llamada d-3-10:
Muestra los análisis de precios unitarios donde falta registrar información.
Llamada d-4:
Permite el acceso del usuario a cualquiera de los catálogos utilizados por el sistema donde puede dar mantenimiento como efectuar nuevos registros, modificar o borrar siempre y cuando el registro no esté relacionado a alguna aplicación; todas estas acciones se realizan sin necesidad de estar elaborando un presupuesto, todos los catálogos son abiertos y puede acceder desde cualquier aplicación.
Recursos: Es el catálogo donde están registrados todos los recursos utilizados por el sistema compuesto por mano de obra, material, equipos y subcontratos.
Unidades: Todas las unidades que utiliza el sistema están registradas en este catálogo de acuerdo a las normas vigentes.
Partidas: En el catálogo de partidas es donde se registran todas las partidas en forma directa sin necesidad de elaborar un presupuesto. Otra forma de ingresar a este catálogo es desde la hoja del presupuesto.
Títulos: El catálogo de títulos tiene registros que serán utilizados en la hoja del presupuesto. Cuando se ingresa desde la hoja del presupuesto es para elegir y registrarlos en la hoja del presupuesto.
Plantillas: Son recursos previamente agrupados por el usuario para luego ser insertados en los análisis de precios unitarios.
Favoritos: Son grupos de partidas previamente preparados por el usuario. Para “armar” favoritos el usuario dispone de los catálogos de partidas y títulos. Lo ideal es tener un conjunto de favoritos para cada tipo de obras. Como favoritos puede tener incluso presupuestos íntegros; los favoritos son “jalados” a la hoja del presupuesto.
Socio de negocio: Es el catálogo donde se registran a todas las personas que tienen que ver con el sistema, como clientes, proveedores, empleados, obreros, subcontratistas, AFP, bancos etc., que luego serán utilizados en alguna aplicación.
Lugares: Son los lugares donde se ejecutan los proyectos y obras están registrados en el sistema; en el caso del Perú, en el nivel 1 los departamentos y para el departamento elegido las provincias en el nivel 2, para la provincia elegida se registraron los distritos en el nivel 3.
Índices unificados: Esta opción, muestra los índices unificados. No efectúe nuevos registros salvo que la institución encargada de regir los haga. Los índices unificados son asignados a los recursos con la finalidad de agruparlos al elaborar la fórmula polinómica.
Llamada d-5:
El menú Herramientas, contiene opciones son comunes para todos los escenarios;
Definir estructura ítem: La estructura de registro de información en los catálogos de definen por esta opción, surte efecto para todos los presupuestos nuevos que registre, de hoy en adelante.
Ítem: Permite configurar el tamaño de letra, color y tipo para los títulos y subtítulos, todos los presupuestos nuevos que se registren tomarán esta configuración. En el escenario de la hoja del presupuesto permite configurar para el presupuesto elegido.
Jerarquía de precios:
Permite definir la estructura de registro del escenario de Precios por grupos.
Jerarquía de Logos:
Permite definir la estructura de registro del catálogo de logotipos.
Definir horario base: Cuando se registra el horario de trabajo por esta opción, en la parte de planeamiento es asignada por defecto al proyecto que se registra.
Correo interno: Activa el correo interno del sistema para enviar o recibir mensajes de un terminal a otro. Está disponible cuando se trabaja en la misma base de datos.
Calculadora: Invoca la calculadora de Windows.
Calendario: Invoca el calendario de Windows.
Limpiar registro de control: Por muchos motivos el sistema puede fallar, estos quedan grabados en el editor de registros; esta opción borra, permitiendo el acceso al escenario donde viene trabajando el usuario.
Llamada e: Es la barra de vistas o de escenarios, está compuesto por grupos y estas a su vez por acceso a los escenarios.
Llamada e-1:
Son grupos, con contiene los accesos a los escenarios.
Llamada e-2:
Es el botón de acceso al escenario.
Llamada f:
Es la estructura de cómo están registrados los presupuestos que se elaboran dentro del S10. Aquí tenemos hasta 5 grupos.
Los botones activos para el árbol en todos los escenarios son:
: Ubica el cursor en el nombre del presupuesto inmediato superior.
: Ubica el cursor en el nombre del presupuesto inmediato inferior.
: Oculta o muestra el árbol.
En el árbol el nombre del presupuesto va acompañado de un libro que cambia de color: (Sólo en el escenario de la hoja del presupuesto):
El color verde del libro en el nombre del presupuesto significa que el presupuesto está en proceso de elaboración.
El color ámbar del libro en el nombre del presupuesto, significa que el presupuesto ha sido procesado, pero le falta elaborar la fórmula polinómica.
El color rojo del libro en el nombre del presupuesto, es cuando está procesado y concluido.
El libro abierto es cuando el presupuesto está en uso.
Muestra el contenido del registro dentro del árbol.
Esconde el contenido del registro dentro del árbol.
Llamada f-1:
Haga que en escritorio sólo estén los presupuestos que viene elaborando o en trabajo, el resto arrastre a las demás ramas.
Llamada f-2:
Esta carpeta contiene los presupuestos de las obras ganadas.
Llamada f-3:
Esta carpeta contiene los presupuestos que están en espera de ganar una licitación o concurso.
Llamada f-4:
Esta carpeta almacena los presupuestos que ya no serán utilizados, solo se puede eliminar los presupuestos que no están asignados a un proyecto.
Llamada f-5:
Esta carpeta contiene los presupuestos que serán eliminados, siempre y cuando no estén asignados a un proyecto.
Llamada g:
Es el escenario.
Llamada g-1:
El sistema trabaja con 10 niveles de títulos, cada nivel tiene un color que lo caracteriza.
Llamada g-2:
Es la partida o actividad que será ejecutada en el proyecto, contiene unidad y metrado.
Llamada g-3:
Muestra el costo total de cada uno de los tipos de recurso que interviene en el análisis de precios unitarios.
Llamada g-4:
Ese botón muestra el detalle de las subpartidas que participan dentro del análisis de precios unitarios.
Llamada g-5:
Cuando elabore presupuestos para instalaciones electromecánicas ingrese el peso, luego dentro del análisis de precios unitarios en la columna cantidad ingrese las hh que son necesarios para el trabajo del peso registrado, además registre los materiales y sus respectivos aportes.
Llamada g-6:
Es el costo unitario calculado.
Llamada g-7:
Es la unidad del análisis de precios unitarios.
Llamada g-8:
Cuando elabore presupuestos para obras civiles, ingrese el rendimiento o la cantidad que quiere lograr en el número de horas de la jornada marcada con la llamada g-12.
Llamada g-9:
Muestra la productividad presupuestada, que es la suma de hh para ejecutar una unidad.
Llamada g-10:
Botones disponibles para el análisis de precios unitarios;
: El color azul indica que el análisis de precios unitarios está en estado sin revisión.
El color amarillo indica que el análisis de precios unitarios está vacío.
El color rojo indica que el análisis de precios unitarios está en estado revisado.
La manito indica que es partida propia de este presupuesto.
: Permite registrar notas al análisis de precios unitarios.
: Invoca a la historia del análisis de precios unitarios, en los presupuestos que participa.
Llamada g-11:
El primer grupo de números, es el código con el que está registrado la partida en el catálogo, el segundo grupo indica el código del presupuesto, y los últimos 2 dígitos es un contador.
Llamada g-12:
Es el número de horas de la jornada interviene en el cálculo del rendimiento.
Llamada g-13:
Son recursos tipo mano de obra.
Llamada g-14:
Son recursos tipo material.
Llamada g-15:
Son recursos tipo equipo.
2.2 Configuración
Muestra la ventana de configuración del sistema, es independiente para cada usuario. Las opciones elegidas son permanentes hasta que el mismo usuario lo vuelva a modificar.
Configuración General:
Contiene opciones generales de configuración del módulo.
Llamada a-1:
Seleccionado – Una vez que ingresa datos en una celda y presione ENTER, el cursor se traslada a la siguiente celda que falta ingresar datos. (Ej.: al ingresar metrados en la hoja del presupuesto, precios, aportes en los análisis de precios unitarios).
No seleccionado – Después de ingresar datos en una celda y presionar ENTER, el cursor permanece en la misma celda.
Llamada a-2:
Seleccionado Activa la opción arrastrar y pegar, disponible en el escenario de la hoja del presupuesto, (ubique el cursor sobre el nombre de un subpresupuesto, luego presionando el botón izquierdo del mouse arrastre a hacia el nombre del subpresupuesto destino). En el caso del escenario del Pie de presupuesto, puede copiar el pie de un presupuesto a otro.
No seleccionado – No se activa la opción de Permitir arrastrar y pegar.
Llamada a-3:
Seleccionado – Cuando se ingresa el índice unificado para una zona geográfica, se replica en las demás zonas.
No seleccionado – El ingreso de datos sólo se efectúa en la celda seleccionada.
Llamada a-4:
Seleccionado – Exporta al MS Project la información de la hoja del presupuesto con los precios de los recursos.
No seleccionado – Exporta al MS Project sin precios.
Llamada a-5:
Seleccionado – Luego de ingresar al catálogo de recursos, solo muestra los recursos genéricos llamados también padres, y se caracterizan porque el código tiene 10 dígitos.
No seleccionado – El catálogo de recursos muestra los recursos genéricos y específicos o llamados también padres e hijos, los que está registrados a 10 y 14 dígitos.
Llamada a-6:
Seleccionado – Cuando un presupuesto tiene grupo de trabajo asignado (se refiere a los usuarios que tienen acceso al módulo), al duplicar el presupuesto hereda el grupo de trabajo.
No Seleccionado – Al duplicar el presupuesto no hereda el grupo de trabajo.
Configuración Procesamiento de Presupuesto:
Disponible para el escenario de la Hoja del presupuesto, permite elegir opciones de procesamiento.
Llamada b-1:
Seleccionado –
Al procesar el presupuesto verifica que en la hoja del presupuesto no falten metrados, caso contrario el sistema solicita el ingreso.
No seleccionado –
Procesa con lo que tiene.
Llamada b-2:
Seleccionado –
Al procesar el presupuesto verifica que en la hoja del presupuesto tenga análisis de precios unitarios de las partidas, caso contrario el sistema solicita el ingreso.
No seleccionado –
Procesa con lo que tiene.
Llamada b-3:
Seleccionado –
Al procesar el presupuesto verifica que todos los recursos tengan precio, caso contrario el sistema solicita el ingreso.
No seleccionado –
Procesa con lo que tiene.
Llamada b-4:
Seleccionado –
Al procesar el presupuesto verifica que dentro de los análisis de precios unitarios todos los recursos tengan cantidades (aportes), caso contrario el sistema solicita el ingreso.
No seleccionado –
Procesa con lo que tiene.
Llamada b-5:
Seleccionado –
Al procesar el presupuesto verifica que dentro de los análisis de precios unitarios todos los recursos estimados tengan los porcentajes que le corresponde y la suma debe llegar al 100%, caso contrario el sistema solicita el ingreso.
No seleccionado –
Procesa con lo que tiene.
Llamada b-6:
Seleccionado –
Al procesar el presupuesto muestra la ventana de opciones de procesamiento.
No seleccionado –
No muestra.
Llamada b-7:
Seleccionado –
Al procesar el presupuesto verifica los recursos considerados en el detalle del recurso.
No seleccionado –
Toma el precio del recurso principal.
Configuración de Análisis de Precios Unitarios:
Llamada c-1:
El botón
muestra la paleta de colores que permite elegir el color para el botón que simboliza los recursos tipo mano de obra, materiales, equipos, subcontratos y subpartidas.
Llamada c-2:
Seleccionado –
Dentro del análisis de precios unitarios, permite definir un rendimiento para mano de obra y otro diferente para el equipo.
No seleccionado –
Dentro del análisis de precios unitarios, toma el mismo rendimiento para mano de obra y para equipo.
Llamada c-3:
Disponible sólo para Ecuador, para el manejo independiente de transporte.
Configuración de Fórmula Polinómica:
Define los parámetros de verificación de la fórmula polinómica.
Llamada d-1:
Define el número máximo de monomios que debe tener la fórmula polinómica.
Llamada d-2: Es el número de submonomios máximo que debe tener un monomio; cuando el índice unificado no llega al 5% (0.05) o cuando se quiere reducir el número monomios, el sistema verifica que se pueden agrupar 3 índices unificados para que formen un monomio.
Llamada d-3:
Define el número de decimales para el factor, generalmente es 3 (hasta el milésimo).
Llamada d-4:
Por norma para que un índice unificado forme un factor es del 0.05 (5%), este valor es fijo.
Configuración de Impresión:
Llamada e-1:
Seleccionado –
Disponible para los reportes de la hoja del presupuesto, análisis de precios unitarios, análisis de precios unitarios de subpartidas, recursos y precios; permite imprimir los reportes con el factor de cambio y la denominación de la moneda alterna.
No seleccionado –
Imprime los reportes en la moneda asignada al presupuesto.
Llamada e-2:
Seleccionado –
Disponible para los reportes de la hoja del presupuesto, análisis de precios unitarios, análisis de precios unitarios de subpartidas, recursos y precios; en los reportes imprime la fecha y hora de impresión, recomendable para los reportes internos.
No seleccionado –
No imprime la fecha y hora de impresión.
Llamada e-3:
Seleccionado –
Al momento de invocar la opción de impresión muestra la ventana para elegir los márgenes.
No seleccionado –
Imprime con el margen definido por defecto.
Llamada e-4:
Seleccionado –
Utilice esta opción cuando el reporte necesite un número de página para compaginar, esta opción permite mostrar la ventana que pide el número de página.
No seleccionado –
No muestra la ventana de ingreso de número de página y por defecto el reporte tendrá el número 1.
Llamada e-5:
Seleccionado –
En los reportes imprime el logotipo asignado al presupuesto.
No seleccionado –
En los reportes no imprime el logotipo.
Llamada e-6:
Seleccionado –
Imprime con el diseño de cabecera elaborado por el usuario en el escenario Cabecera para reportes.
No seleccionado – Imprime el reporte con el diseño elaborado por el S10 por defecto.
Llamada e-7:
Seleccionado –
Disponible para la hoja del presupuesto, imprime con los ítems registrados por el usuario en la columna Ítem Alterno.
No seleccionado –
Imprime la hoja del presupuesto con los ítems generados en forma automática por el S10.
Llamada e-8:
Seleccionado –
Disponible para el reporte de la hoja del presupuesto, imprime con los títulos en los colores seleccionados.
No seleccionado –
Imprime el reporte de la hoja del presupuesto en blanco y negro.
Llamada e-9:
Seleccionado –
El nombre de la partida dentro de la hoja del presupuesto acepta hasta 250 caracteres, pero sí necesita una descripción más larga puede usar la nota e ingresar varias líneas. Esta opción imprime el reporte con las notas como nombre de las partidas.
No seleccionado –
Imprime el reporte de la hoja del presupuesto con los nombres de las partidas que son mostrados.
Llamada e-10:
Seleccionado –
Disponible para el nombre de la partida. Disponible sólo sí en el catálogo de Partidas, también registró el nombre alterno. Esta opción imprime la hoja del presupuesto con los nombres alternos de las partidas.
No seleccionado –
Imprime el reporte de la hoja del presupuesto, con los nombres que muestra.
Llamada e-11: Seleccionado –
Imprime la hoja del presupuesto sólo hasta costo directo.
No seleccionado –
Imprime la hoja del presupuesto con el pie de presupuesto.
Llamada e-12:
Seleccionado –
Imprime en la hoja del presupuesto el % de gastos generales calculado.
No seleccionado – E
n la impresión de la hoja del presupuesto no muestra el % de gastos generales calculado.
Llamada e-13:
Seleccionado –
En la hoja del presupuesto imprime, el nombre de las partidas, la unidad y el metrado, para que el postor registre los precios.
No seleccionado –
Imprime la hoja del presupuesto completo.
Llamada e-14:
Seleccionado –
Imprime la hoja del presupuesto tal como se muestra en el escenario indentado o subordinado con sangría.
No seleccionado –
Imprime la hoja del presupuesto sin indentado.
Llamada e-15:
Seleccionado –
Imprime los análisis unitarios con los recursos que tengan valores mayores o menores a cero.
No seleccionado –
Imprime los análisis unitarios con todos los recursos.
Llamada e-16:
Seleccionado – I
mprime los análisis unitarios con el código de la partida.
No seleccionado –
Imprime los análisis de precios unitarios solo con el nombre de los recursos.
Llamada e-17:
Seleccionado –
Imprime el rendimiento en los análisis de precios unitarios.
No seleccionado –
No imprime el rendimiento en los análisis de precios unitarios.
Llamada e-18:
Seleccionado –
Imprime la cantidad de horas hombre en los análisis de precios unitarios.
No seleccionado –
No imprime la cantidad de horas hombre en los análisis de precios unitarios.
Llamada e-19:
Los factores de cantidad se aplican a las partidas del presupuesto, sí aplicamos desde el nombre del presupuesto afectará a todas las partidas del presupuesto, sí se aplica desde un subpresupuesto afectará al grupo de partidas del subpresupuesto.
Seleccionado –
Imprime los factores de cantidad aplicados a la partida en el reporte de los análisis de precios unitarios.
No seleccionado –
No imprime los factores de cantidad en el reporte de los análisis de precios unitarios.
Llamada e-20:
Seleccionado –
Imprime la cantidad de horas de la jornada en el reporte de los análisis de precios unitarios.
No seleccionado –
No imprime la cantidad de horas de la jornada en el reporte de los análisis de precios unitarios.
Llamada e-21:
Seleccionado –
Imprime el pie de presupuesto en el reporte de los análisis de precios unitarios.
No seleccionado –
No imprime el pie de presupuesto en el reporte de los análisis de precios unitarios.
Llamada e-22:
Seleccionado –
Imprime el nombre de los recursos con la descripción alterna en el reporte de los análisis de precios unitarios.
No seleccionado –
Imprime el nombre de los recursos por defecto en el reporte de los análisis de precios unitarios.
Llamada e-23:
Seleccionado –
Imprime el reporte de los análisis de precios unitario con el factor de conversión (registrado en el escenario de Datos generales – Datos adicionales).
No seleccionado –
Imprime el reporte de los análisis de precios unitarios sólo en la moneda seleccionada para el presupuesto.
Configuración de Planeamiento:
Llamada f-1:
Seleccionado –
En el escenario del Cronograma por periodos, tanto para el modo programado como para la valorización controla no excederse del 100% al ingresar los avances.
No seleccionado –
En el escenario del Cronograma por periodos, tanto para el modo programado como para la valorización permite excederse del 100% al ingresar los avances.
Llamada f-3:
Cuando el presupuesto asignado al proyecto no tiene análisis de precios unitarios (sólo está conformado por partidas formato), es necesario elegir un recurso para que tome el precio de los recursos, caso contrario en sistema enviará un mensaje en el escenario del Actividades.
Llamada f-4:
Seleccionado, permite el uso del MS Project 2007 o superior, licenciado.
No seleccionado, por defecto usa el MS Project 2003.
Configuración Hoja de Presupuesto:
Llamada g-1:
Seleccionado,
muestra el ítem alterno con el código de registro del catálogo.
No seleccionado,
muestra el código alterno de registro, como ítem alterno.
Llamada g-2:
Seleccionado,
cuando seleccione una partida o un título del catálogo, muestra el código alterno de registro, como ítem alterno.
No seleccionado –
muestra el ítem alterno con el código de registro del catálogo.
Llamada g-3
:
Seleccionado –
Solicita el número de dígitos del código para el ítem, recomendado 6
No seleccionado –
Por defecto muestra toda la cantidad de dígitos como código.
Llamada g-4:
Seleccionado –
Es recomendable que esta opción siempre esté seleccionada; cuando se elige una partida del catálogo a la hoja del presupuesto esta se convierte en propia, esto permite que todas las modificaciones tengan efecto sólo en la partida de la hoja del presupuesto y no en el catálogo.
No seleccionado –
Sí elije una partida del catálogo a la hoja del presupuesto y lo modifica, también lo modificará en el catálogo.
Llamada g-5:
Seleccionado –
Es recomendable que esta opción siempre esté seleccionada; cuando invoca una subpartida desde el análisis de precios unitarios ubicado en la hoja del presupuesto, esta se convierte en propia, esto permite que todas las modificaciones tengan efecto sólo en la subpartida de la hoja del presupuesto y no en el catálogo.
No seleccionado –
Sí modifica la subpartida, también lo modificará la partida en el catálogo.
Llamada g-6:
Seleccionado –
Sí elimina o borra una partida de la hoja del presupuesto, y vuelve a seleccionarla del catálogo a la hoja del presupuesto, le dará preferencia a la primera partida propia.
No seleccionado –
Sí elimina o borra una partida de la hoja del presupuesto, y vuelve a seleccionarla del catálogo a la hoja del presupuesto, le dará preferencia a la partida del catálogo.
Llamada g-7:
Seleccionado –
Primero modifique el nombre de la partida en la hoja del presupuesto, luego con clic derecho reasigne una partida del catálogo. La partida reasignada será grabada dentro del catálogo de partidas-partidas propias con el nombre modificado por el usuario.
No seleccionado –
La partida reasignada será grabada dentro del catálogo de partidas-partidas propias con el nombre del catálogo.
Llamada g-8:
Seleccionado –
Al presionar el botón izquierdo del ratón sobre el nombre ubicado en la cabecera de la columna ordena la información en forma ascendente o descendente.
No seleccionado –
No se activa la opción de ordenar por cualquier columna.
Configuración de interfaces:
El S10 ha preparado interfaces para el SAP y el Dynamics de Microsoft, el que se configura mediante la pestaña de “Configuración de Interfaces”.
3 Uso del módulo de presupuestos
Haga doble clic sobre el botón
, que se encuentra en el tapiz de su PC, luego el sistema muestra,
Llamada a:
El usuario “sa” es el administrador del sistema, el encargado del centro de cómputo.
Vaya a utilitarios, escenario de Mantenimiento de usuario, registre a los usuarios que manejarán el módulo de presupuestos, asigne atributos.
3.1 Formas de búsqueda en los catálogos
La forma de ubicar información es igual en todos los catálogos que maneja el sistema, se toma como ejemplo el catálogo de recursos.
3.1.1 Búsqueda global
La búsqueda global consiste en digitar las primeras letras del registro que se quiere ubicar en los catálogos. Ubique el cursor en la columna donde se quiere buscar información.
Llamada a:
Esta celda muestra las letras o números que se va ingresando al momento de efectuar la búsqueda, en los sectores marcados con la llamada a-2 y a-3; no disponible para el árbol.
Llamada a-1:
Para ubicar información en el árbol, debe estar previamente desplegado y el cursor debe estar ubicado es ese segmento y sólo digite letras.
Llamada a-2:
Para ubicar información en una columna que contiene números, el cursor debe estar ubicado es ese segmento y sólo digite números.
Llamada a-3:
Para ubicar información en una columna que contiene letras, el cursor debe estar ubicado es ese segmento y sólo digite letras.
Para deshacer la búsqueda use las teclas .
3.1.2 Búsqueda local
La búsqueda local permite ubicar información en una columna usando la tecla F8 o usando el botón y en la ventana que muestre ingresar hasta tres palabras de la cadena que se quiere ubicar, no importa el orden.
Ubique el cursor en una columna (cualquiera marcada con las llamadas a-2, a-3),
Luego use el botón
Llamada a:
Esta celda es informativa, muestra las palabras que se buscan.
Llamada b:
Ingrese información en cada una de las celdas, no importe el orden.
Llamada b-1:
Efectúa la búsqueda.
Llamada a-4:
Deshace la búsqueda.
3.2 Registro de información en el catálogo de presupuestos
Es importante analizar el tipo de negocio que tiene la empresa con la finalidad de registrar la información en el árbol del catálogo de presupuestos, tome en cuenta la cantidad de presupuestos que elaborará.
Llamada a:
En este nivel permite el registro del 00 al 99.
Llamada b:
En este nivel permite el registro del 00 al 99.
Llamada c:
En este nivel permite el registro del 000 al 999.
Llamada d:
Muestra el código del registro elegido.
Llamada e:
De acuerdo a este registro en el árbol del Escenario de Datos Generales se agruparán los presupuestos.
Ejemplo de registro de información en el catálogo de presupuestos por línea de negocio
Ejemplo de registro de información en el catálogo de presupuestos por años
Llamada a:
Registre un grupo por año.
Llamada b:
Dentro del grupo año, registre los tipos de proyecto que realiza.
En empresas donde se elaboran muchos presupuestos por proyecto.
En proyectos de obras electromecánicas, cuando se tiene varias plantas donde se fabrican partes del proyecto.
• En el módulo de Gerencia de proyecto se usa las dimensiones la juntar la información de Planta A + Planta B + Planta C.
Para ejecutar proyectos a nivel nacional con base de datos centralizada:
3.3 Registro del presupuesto
Registre dos presupuestos, PUENTE EN CARRETERA, que es un presupuesto meta y es desarrollado en la presente guía, el otro es LA CASA DE JUAN, que lo debe registrar el usuario como práctica y seguimiento para algunas opciones.
Llamada 1: El presupuesto PUENTE EN CARRETERA, no tiene diseño del pie de presupuesto, tampoco fórmula polinómica, es un presupuesto meta.
Llamada 2: Para el presupuesto. LA CASA DE JUAN desarrollaremos el pie de presupuesto y la fórmula polinómica, es un presupuesto venta.
Llamada 2-1: Los nombres de los subpresupuestos serán cambiados para efectuar el seguimiento de esta guía.
Acceso al escenario de Datos Generales
En el árbol del catálogo de Presupuestos el sistema sugiere un ordenamiento previo para los presupuestos que registre el usuario.
Para registrar PUENTES, dentro de OBRAS VIALES.
El último nivel de registro en el árbol es la “hojita”. Para registrar el nombre del presupuesto:
Ingreso de datos del presupuesto
Llamada 1:
Invoca a la ventana Datos adicionales; la configuración es para el presupuesto que se registra.
Llamada a:
Elija la pestaña configuración.
Llamada a-1:
Seleccionado – Después de procesar el presupuesto, en el árbol, el botón del subpresupuesto lo pinta de color amarillo indicando que le falta elaborar la fórmula polinómica.
No seleccionado – Después de procesar el presupuesto, en el árbol, el botón del subpresupuesto muestra de color rojo indicando que el presupuesto está concluido.
Llamada a-2:
Seleccionado – Utilice esta opción cuando elabore presupuestos para movimientos de tierras en carreteras, porque los rendimientos son altos (más de 3,500 m3).
No seleccionado – Utilice esta opción cuando elabore presupuestos para edificaciones y obras de arte, porque los rendimientos son bajos (menos de 3,500).
Llamada a-3:
Seleccionado – Imprime el reporte de la hoja del presupuesto con precios unitarios incluidos costos indirectos
No seleccionado – Imprime el reporte de la hoja del presupuesto con precios unitarios sin los costos indirectos
Llamada a-4:
Tipo 1 – Procesar el presupuesto con partidas estimadas.
Tipo 3 – Procesa el presupuesto y calcula los acumulados para la fórmula polinómica y presupuesto descompuesto – RECOMENDADO.
Llamada a-5:
Pie de los análisis de precios unitarios diferentes al Pie de Presupuesto, es necesario que, en el escenario del pie de presupuesto, definan un diseño para los análisis de precios unitarios.
Seleccionado – Imprime el reporte de los análisis de precios unitarios con un diseño propio.
No seleccionado – Imprime el reporte de los análisis de precios unitarios con el diseño que viene por defecto.
Llamada a-6:
Utilice esta opción exclusivamente para imprimir el presupuesto en otra moneda. Ej.: Sí elaboró el presupuesto en Soles, puede imprimir el presupuesto con los valores multiplicados por el factor de cambio de una moneda alterna que puede ser Dólares.
Para imprimir el reporte de la hoja del presupuesto con el factor de cambio, es necesario ingresar a configuración usando el botón de la barra de botones, luego elegir la pestaña Impresión y poner una selección en la celda: “Con factor de cambio”.
En la celda Factor de cambio, ingrese el factor 1/tipo de cambio (se refiere al valor de compra en el banco) y en la celda moneda elija la moneda alterna que puede ser Dólares Americanos.
Llamada a-7:
Antes de elegir el Tipo de Obra, el usuario tiene que registrar previamente el Tipo.
Llamada a-7-1:
Use el botón para elegir el tipo de obra.
Llamada a-7-2:
Ingrese la cantidad, sí es edificación la cantidad de m2 vendibles, si es puente m, sí es carretera km.
Luego use el botón
para grabar los datos y cerrar la ventana CATALOGO DE TIPOS DE OBRA.
Llamada b:
Permite configurar la cantidad de decimales que serán utilizados en la hoja del presupuesto y en los análisis de precios unitarios.
Llamada 2:
Del Catálogo de Socio de negocio elija al cliente; en el caso de no encontrarlo efectúe el registro, para luego elegirlo, en este caso al Ministerio de Transportes y Comunicaciones.
Para que el nombre del cliente sea elegido, necesariamente debe estar registrado en el catálogo.
Retorne a la ventana “PRESUPUESTO”.
Llamada 3:
Al hacer clic en el botón marcado con la llamada 3: de la ventana “Presupuesto”, el sistema muestra el Catálogo de Ubicación Geográfica, para elegir el distrito donde se ejecutará la obra, en este caso Urubamba.
Retorne a la ventana “PRESUPUESTO”
Llamada 4:
Elija la fecha para el presupuesto. Los precios de los recursos son considerados a esta fecha.
IMPORTANTE: Los precios son almacenados de acuerdo a la fecha, lugar y por presupuesto.
Llamada 5:
Registre el plazo de ejecución, en días calendarios, es un dato informativo y generalmente obedece a los requerimientos del cliente.
Llamada 6:
Registre la cantidad de horas de trabajo en un día laborable, influye en el cálculo del rendimiento. EN EL PRESUPUESTO META HAGA QUE LA JORNADA DIARIA SEA LA MISMA EN TODAS LAS PARTIDAS DEL PRESUPUESTO EL MS PROJECT, NO ACEPTA JORNADAS DIFERENTES.
Llamada 7:
El sistema puede elaborar presupuestos en dos tipos de moneda al mismo tiempo, eso significa que permite ingresar los precios de los recursos ya sea en Soles y en Dólares al mismo tiempo, o en Soles para algunos recursos y en Dólares para otros. Cuando esta opción no está seleccionada, el presupuesto es elaborado en una moneda. SI ES PRESUPUESTO META ELABORE EL PRESUPUESTO EN UNA SOLA MONEDA DEBIDO A QUE EL MS PROJECT ACEPTA UNA SOLA MONEDA.
Llamada 8:
Del catálogo elija el tipo de moneda para el presupuesto.
Llamada 9:
Estas celdas son editables y tiene la calidad de informativas, para ingresar los datos del presupuesto base de una licitación.
Llamada 10: Estas celdas son informativas y muestra los resultados del procesamiento del presupuesto.
La ventana “PRESUPUESTO”.
Cuando concluya con el registro de todos los datos, use el botón para grabar y cerrar la ventana PRESUPUESTOS. Otra forma es hacer doble clic sobre el nombre del presupuesto para trasladarlo al árbol del escenario de Datos generales, como se muestra,
3.4 Registro del subpresupuesto
Luego del registro del presupuesto, el sistema crea automáticamente un subpresupuesto con el mismo nombre, el que debe ser editado y cambiado por el usuario. Ingrese Estructuras y luego en la siguiente fila, Gastos generales.
3.5 Registro de información en la hoja del presupuesto
Previamente se registró el presupuesto, en el escenario de Datos Generales.
Acceso al escenario
Es el escenario donde se registra el contenido del presupuesto como son los títulos, las partidas, se ingresa los metrados y se procesa el presupuesto.
Todas las opciones de edición están disponibles para el subpresupuesto.
3.5.1 Configuración previa
Use el botón para ingresar a la ventana de configuración
Es recomendable que todas las partidas que lleguen a la hoja del presupuesto se conviertan automáticamente en propias.
Llamada a:
Hace que todas las partidas que llegan a la hoja del presupuesto se convierten automáticamente en propias, lo que significa que cualquier modificación sólo afectará a la partida que se encuentra en el presupuesto y el origen no será alterado.
Seleccionado: Convierte las partidas en propias.
No seleccionado: No convierte a propias, sí modifica el análisis de pecios unitarios desde el presupuesto también modificará el análisis de precios unitarios origen.
Llamada b:
Seleccionado: Hace propio en forma automática las partidas con sus respectivas subpartidas que se encuentran dentro de un análisis de precios unitarios
No seleccionado: No convierte a propias, sí modifica el análisis de precios unitarios desde el presupuesto también modificará el análisis de precios unitarios origen.
Llamada c:
Seleccionado: Sí elimina una partida propia del presupuesto, cuando “vaya” al catálogo para volverla a elegir, el sistema siempre da preferencia la primera partida propia.
No seleccionado: Sí elimina una partida propia del presupuesto, cuando “vaya” al catálogo “jalará” la partida básica del catálogo.
3.5.2 Formas de registro de información en la hoja del presupuesto
Se dispone de 4 formas:
3.5.2.1 Caso 1) Copiar partidas de presupuestos elaborados
Los registros elegidos se encuentran en el portapapeles.
A continuación, pegue el contenido del portapapeles en el presupuesto destino.
Temas relacionados: Ver: Ítem alterno
3.5.2.2 Caso 2) Uso de favoritos
El usuario previamente debe registrar los favoritos en el catálogo para luego seleccionar desde la hoja del presupuesto. Los favoritos prácticamente son hojas de presupuesto que contiene partidas revisadas y confiables. Para visualizar los favoritos ingrese al menú Catálogos, luego elija Favoritos.
Luego use el botón
para generar los ítems en la hoja del presupuesto.
3.5.2.3 Caso 3) Importar al S10 presupuestos elaborados en hoja Excel
Debemos tener la información en Excel antes de importarla. Las celdas deben contener sólo valores NO FORMULAS.
Grabe y cierre Excel, antes de importarlo al S10.
Para importar hoja Excel;
Ubique la carpeta y luego elija el archivo de Excel con doble clic o abrir.
Elija las filas
Elija el número de fila inicial y final, para fijar el rango de partidas que será importado; en este caso elija la fila 2 hasta la 12.
Identifique las columnas que serán importadas, en este caso:
ITEM, DESCRIPCIÓN, UNIDAD, METRADO, PRECIO 1.
La columna (campo) precio 2, se refiere al precio del recurso en la moneda alterna, esto cuando se elabora presupuestos en doble moneda.
Importar información seleccionada
Requisito: Antes de usar esta opción deben estar definidas las filas y las columnas.
Luego en la hoja del presupuesto, ordene las partidas con relación a los títulos con el botón , y genere los ítems .
Las partidas importadas desde Excel, no tienen análisis de precios unitarios (a esto le llamamos partidas formato).
Este tipo de presupuesto puede ser utilizado para valorizar, ya que no es necesario que tenga análisis de precios unitarios siempre y cuando no solicite adelanto para materiales.
3.5.2.3.1 Asignar análisis de precios unitarios a una partida formato
Visualice en la ventana inferior los “Análisis de precios unitarios”.
Haga clic derecho sobre el nombre de la partida (ventana superior), luego del menú que muestre, elija la opción “Reasignar partida del catálogo”. En el catálogo de partidas ubique una partida cuyo análisis de precios unitarios es la que necesita, elija con doble clic.
No necesariamente la partida tendrá el mismo nombre que el análisis de precios unitarios, al imprimir tomará el nombre de la partida ubicada en la hoja del presupuesto. El análisis de precios unitarios asignado a la partida se convierte en propia y permite ser personalizado.
3.5.2.4 Caso 4) Registro del presupuesto de la forma tradicional
Consiste en registrar el contenido de la hoja del presupuesto, primero los títulos, luego para cada uno de los títulos las partidas que le corresponden.
Modo ver – Para esconder los análisis de precios unitarios, haga que la hoja del presupuesto no muestre la hoja de análisis de precios unitarios en la parte inferior:
3.5.2.4.1 Registro de títulos en la hoja del presupuesto
Datos que serán registrados:
Adicionar títulos a la hoja del presupuesto
En el catálogo de títulos,
El botón
divide horizontalmente el catálogo mostrando la ventana inferior (Registros seleccionados).
El botón
traslada los datos seleccionados a la hoja del presupuesto.
Generar los ítems.
3.5.2.4.2 Registro de partidas en la hoja del presupuesto
Las partidas deben ser registradas previamente en el catálogo de partidas antes de ser elegidas para la hoja del presupuesto.
Registre las partidas:
Selección de partidas del Catálogo de Partidas (el usuario previamente debe registrarlos),
Llamada a:
Cuando el cursor se ubica en TODAS LA PARTIDAS, en el segmento marcado con la llamada b-1 agrupa a todas las partidas del catálogo.
Llamada b:
Cuando el cursor se ubica en OBRAS DE ARTE, en el segmento marcado con la llamada b-1 agrupa sólo a las partidas que pertenece a este grupo.
Llamada b-2:
Muestra el análisis de precios unitarios de la partida elegida en la ventana b-1.
Llamada b-3:
Sólo las partidas elegidas en este segmento pasan a la hoja del presupuesto.
Las partidas son mostradas en la hoja del presupuesto.
Use el botón para generar los ítems
3.5.2.5
Ingreso de metrados en la hoja del presupuesto
Además, están disponibles las opciones:
• Luego de exportar la hoja del presupuesto, se envía por e-mail a la persona que realiza los metrados, quien debe registrarlos en la hoja Excel, luego al recibir de vuelta el e-mail, se abre Excel, se marca un bloque de los metrados se copia al portapapeles y se realiza PEGAR DE OTRAS APLICACIONES, en la celda de metrados en el S10.
• Vincular la celda de metrados del S10, con la celda de metrados de Excel.
3.6 Procesar el presupuesto
3.6.1 Configuración previa al procesamiento
Use el botón para mostrar la ventana de configuración, atributo del sa. Verifique opciones de procesamiento
3.6.2 Procesamiento del presupuesto
Para procesar el presupuesto use el botón ubicado en la barra de botones del escenario de la Hoja del Presupuesto, también puede pulsar la tecla F9 luego el sistema muestra la ventana de diálogo:
Al momento de procesar el presupuesto, el sistema verifica que no le falte metrados en la hoja del presupuesto y en el análisis de precios unitarios todos los recursos tengan aportes (cantidades) y precios. En el caso de que falte algún dato el sistema envía un mensaje con el requerimiento de ingreso del dato. Luego de registrar el dato, cierre la ventana y vuelva a procesar el presupuesto.
3.6.3 Opciones en Recursos y precios
El sistema filtra sólo los recursos que utiliza el presupuesto, luego de procesarlo. Se muestran algunas utilidades con los recursos que pertenecen a las partidas propias de un presupuesto.
3.6.3.1 Partidas que utilizan el recurso
Muestra las partidas donde participa el recurso elegido. Esta opción es informativa.
Luego cierre la ventana.
3.6.3.2 Reasignar recurso
Permite reemplazar un recurso por otro utilizado en el presupuesto o del catálogo, esta opción es útil cuando en el listado de recursos, un recurso aparece dos veces con características diferentes, o se quiere cambiar un recurso por otro.
Llamada 1: Haga clic derecho sobre el recurso que será cambiado.
Llamada 2: De la ventana que muestre, elija el recurso que lo reemplazará.
Luego, use el botón para calcular las cantidades de los recursos.
3.6.3.3 Eliminar recursos
Permite eliminar recursos del listado. Un recurso eliminado no puede ser recuperado, para volverlo a insertar lo tiene que hacer desde el análisis de precios unitarios.
3.6.4 Ingreso de precios de los recursos
Los precios de los recursos son almacenados de acuerdo al lugar y la fecha además por presupuesto.
Luego cierre la ventana de Recursos y precios
Vuelva a procesar el presupuesto usando el botón luego en la ventana de procesamiento elija Continuar.
3.6.5 Resultado del procesamiento
3.6.6 Ratios
Esta opción es informativa que muestra el costo por m de construcción para este tipo de obra, es necesario mostrar la columna donde se encuentran los datos, (toma la información del escenario de datos generales – datos adicionales, donde se ingresó el tipo de obra y la cantidad).
Los datos son mostrados para el subpresupuesto y para cada uno de los títulos y partidas.
Para visualizar las ratios en gráfico,
3.7 Presupuesto venta
Este presupuesto es utilizado para elaborar el pie de presupuesto y la fórmula polinómica.
ç
3.8 Diseño del pie de presupuesto
Esta opción permite particularizar la forma de presentación y cálculo de la parte final del presupuesto a partir del costo directo.
El presupuesto meta no tiene pie de presupuesto, sólo el venta u oferta.
Para el caso del presupuesto venta, el pie de presupuesto es todo aquello que va por debajo del costo directo y se “carga” al índice unificado 39 INDICE GENERAL DE PRECIOS AL CONSUMIDOR (que viene a ser los GASTOS GENERALES Y UTILIDAD). Ingrese a esta opción después de haber concluido el presupuesto y procesado.
3.8.1 Registro del recurso para relacionarlo con el concepto
Cuando la buena pro, es otorgada, el postor solicita un porcentaje del monto total de la oferta, esto comprende el costo directo + gastos generales + utilidad + impuestos.
El diseño del pie de presupuesto debe estar preparado para ser utilizado en CONTRATOS, para el presupuesto que alcance al propietario, llamado presupuesto venta.
Ingrese al módulo de Presupuestos como “sa”
Ingrese al catálogo de Recursos, para registrar los recursos que serán utilizados en el pie como recurso relacionado.
Llamada a:
Ubique la rama Subcontratos y Servicios.
Llamada a-1:
Registre el grupo, Conceptos Pie Presupuesto + Gastos Generales.
Llamada a-1-1:
Registre el subgrupo, Conceptos para Pie Presupuesto y para contratos.
Llamada a-1-1-1:
Registre el recurso, principal y el específico que será utilizado en el pie del presupuesto.
Llamada a-1-1-2:
Registre el recurso, principal y el específico que será utilizado en Contratos, el nombre de cada uno de los subpresupuestos, cada uno de ellos será vinculado con la Fase del WBS.
3.8.2 Registro de Conceptos para el pie de presupuesto
Ingrese por el menú Catálogos a Concepto Pie Presupuesto, para registrar el concepto que será utilizado en el pie, al que se relaciona con un recurso del catálogo.
Datos que serán registrados:
Concepto GASTOS GENERALES
Variable GG
Recurso del catálogo GASTOS GENERALES
Registre otro concepto para utilidad, al registrar el contrato, aparecen ítems para asignar WBS, para cada uno de los subpresupuestos y para los costos indirectos.
Cierre el catálogo de conceptos.
3.8.3 Registro del pie de presupuesto
Ingrese como usuario;
Datos que serán registrados:
Ingrese al escenario del Pie de Presupuesto, luego elija el proyecto.
En la ventana Diseño Pie de Presupuesto,
Llamada a:
Por defecto el pie de presupuesto siempre muestra el Costo directo, en la variable nDirecto el sistema acumula todo el monto presupuestado que corresponde a mano de obra + materiales + equipos + subcontratos.
Llamada b:
Registre manualmente el resto del pie de presupuesto, en la columna Macro ingrese la fórmula.
Llamada b-1:
Las variables pueden ser en minúscula o mayúscula indistintamente.
Llamada b-2:
Ponga una selección para que este ítem no participe en la elaboración de la fórmula polinómica, no es permitido pedir reintegro de impuesto.
Vaya al escenario de la hoja del presupuesto y procese.
Para procesar el presupuesto use el botón
Para visualizar el resultado del procesamiento y pie use el botón
3.9 Gastos generales
Los gastos generales, es la estructura de costos que demanda administrar y ejecutar la obra por parte de la empresa constructora. Considere todo lo que no aparece en los análisis de precios unitarios como son los sueldos de los residentes de obra, alquiler de la oficina, camionetas; gastos de teléfono, costo de los créditos, etc. La condición para elaborar el cálculo analítico de gastos generales es tener registrado la variable ggp o GGP en el Diseño del Pie de Presupuesto en la columna macro. Se recomienda elaborar el cálculo analítico de gastos generales para el presupuesto venta u oferta, para conocer el costo y porcentaje de los gastos generales. Para la meta elabore un presupuesto detallado.
Datos que serán registrados: Para el presupuesto
Acceso al escenario de Gastos Generales
3.9.1 Registro de rubros
3.9.2 Registro de conceptos
Repita el proceso para los demás rubros.
3.9.3 Calcular el porcentaje de gastos generales
Con lo que concluye el cálculo analítico de gastos generales, vaya al escenario de la hoja del presupuesto y procese el presupuesto para que tome el pie de presupuesto.
Reporte,
3.10 Fórmula polinómica
Las fórmulas polinómicas, constituyen un procedimiento convencional de cálculo para obtener el valor de los incrementos de costos que experimentan los presupuestos de obra en el tiempo. El sistema está basado en la incidencia de todos los ELEMENTOS que constituyen una obra y participan en una proporción constante durante todo el tiempo que demanda dicho proceso. Deben cumplir la condición de VERIFICABLES, que fácilmente se puedan revisar los procedimientos de cálculo. Luego de reemplazar los componentes de los índices unificados del INEI, dará como resultado el “K” que permite actualizar el costo de una obra (o parte de ella que es el avance de ejecución) a partir de la fecha del presupuesto base, por el incremento que se genera en cada uno de los elementos componentes, esto en el proceso de valorización.
3.10.1 Requisitos para el Perú
• Una fórmula polinómica está constituida por la incidencia de hasta 8 monomios.
• El coeficiente de incidencia mínimo de un monomio es de 5 centésimos.
• Los valores de los COEFICIENTES DE INCIDENCIA están dados en cifras decimales con aproximación al milésimo.
• La suma de todos los COEFICIENTES DE INCIDENCIA de una fórmula polinómica es de uno (1)
• Los valores del FACTOR DE REAJUSTE (K) están expresados al milésimo.
• Cuando los ELEMENTOS COMPONENTES participan con COEFICIENTES DE INCIDENCIA menores a los 5 centésimos, se pueden agrupar con otros ELEMENTOS COMPONENTES de tal forma que su incidencia asociada supere al mínimo señalado, constituyendo así los monomios compuestos.
• Los MONOMIOS COMPUESTOS pueden estar formados hasta por 3 ELEMENTOS COMPONENTES que se les puede denominar SUB MONOMIOS.
• Un presupuesto para el estado, debe tener máximo 4 fórmulas polinómicas.
Para elaborar la fórmula polinómica, tiene que OMITIR en el pie del presupuesto el ítem relacionado al IMPUESTO.
Acceso al escenario de la Fórmula Polinómica
Esta ventana es mostrada sólo en el caso de que falte índice unificado al recurso. Para asignar:
Llamada a: Esta ventana muestra solo a los recursos que les falta asignar índice unificado. Después de asignar índice unificado al recurso irá desapareciendo.
Después cierre la ventana Catálogo de Índice Unificado.
IMPORTANTE:
• Elija Sí, para elaborar la fórmula polinómica.
• Elija No, cuando ya tiene fórmula polinómica, ingrese sólo para visualizar o imprimir.
• Use el botón sólo en el caso de que el sistema no muestre en forma automática el cuadro de diálogo.
3.10.2 Agrupamiento preliminar
De preferencia no agrupe el índice unificado 47 que corresponde a mano de obra incluido leyes sociales, tampoco el 39 que es el índice general de precios al consumidor y viene a ser los gastos generales y utilidad; luego para los demás índices unificados tome el valor menor y agrupe al mayor, tome el criterio de agrupar semejantes.
En agrupamiento preliminar:
1.- Agrupe los IU que son de la misma familia el que tiene % menor al mayor.
2.- El IU Dólar, tómelo como un comodín, agrupe a lo que menos incomode.
3.- El IU de Herramienta Manual, agrupe a un equipo o mano de obra.
Llamada 1: Esta ventana muestra los valores de los índices unificados. Los valores de las columnas de las llamadas 2 y 3 inicialmente son iguales y la suma vertical llega al 100%.
Para agrupar índices unificados proceda de la siguiente manera:
Marque el índice unificado y mantenga presionado el botón izquierdo del ratón luego arrastre sobre el índice unificado destino donde quiere que se agrupe y luego suelte el botón izquierdo del ratón. Otra forma es usar el menú que muestra al hacer clic derecho.
A continuación, se muestra el agrupamiento preliminar;
Para deshacer un agrupamiento haga clic derecho sobre el índice unificado (a deshacer) y el sistema muestra el menú,
En la ventana que muestre seleccione el índice unificado a deshacer, marque con una selección por más que tenga un solo registro.
3.10.3 Configuración de la fórmula polinómica
Para visualizar los parámetros de la configuración use el botón , que muestra,
3.10.4 Conformación de monomios
Las normas peruanas recomiendan: Que la fórmula polinómica esté integrada hasta por 8 monomios, y estos a su vez hasta por 3 submonomios, puede usar el siguiente orden, pero no es obligatorio.
1 Mano de obra incluido leyes sociales – 2 Materiales – 3 Materiales – 4 Materiales – 5 Materiales – 6 Equipo – 7 Varios – 8 Gastos generales y utilidad
El factor de cada monomio debe tener un valor mínimo de 5% (0.05) para lograr esto se puede agrupar hasta 3 submonomios.
Acceso a conformación de monomios
El botón tome los datos del agrupamiento preliminar.
Llamada a:
Elija la pestaña de conformación de monomios.
Llamada a-1:
Verifique que el factor llegue al 0.05 para que vaya solo formando un monomio.
Llamada a-1-1:
Sí el factor de la mano de obra pasa del 0.05, registre el número 1, (no existe una norma que indique que la mano de obra sea el monomio número uno, se acostumbra).
Llamada a-1-2:
Sí el factor no llega al 0.05 puede agruparlo, registrando el mismo número de monomio, hasta que la suma llegue o pase del 0.05, se agrupa poniendo el mismo número, el reglamento permite agrupar hasta 3 índices unificados para que formen un monomio y pasarán a llamarse submonomios.
Llamada a-1-3:
A los demás índices unificados registre un número correlativo, el reglamente sólo permite tener 8 monomios.
Llamada a-2:
Los porcentajes irán cambiando de acuerdo al agrupamiento.
Llamada a-3:
Cuando se agrupan índices unificados, sólo el que tiene el porcentaje más alto hereda los símbolos.
Sí cambia el agrupamiento preliminar, el botón vuelve a tomar los datos.
Sí la fórmula polinómica está concluida y usa el botón deshace al agrupamiento.
A continuación, se muestra el agrupamiento concluido.
3.10.5 Verificación de la fórmula polinómica
El botón verifica la elaboración de la fórmula polinómica de acuerdo a la configuración previamente registrada.
3.10.6 Reemplazo de símbolos
Cambie los símbolos, no es permitido usar el mismo símbolo en dos monomios.
Use el botón
para visualizar el reporte en pantalla.
Reporte de la fórmula polinómica:
3.11 Impresión de reportes
Para imprimir los reportes el sistema dispone de dos botones:
Use este botón para enviar el reporte directamente a una impresora predeterminada mediante el sistema operativo Windows.
Use este botón para mostrar en pantalla el reporte.
3.11.1 Reportes desde el escenario de Datos Generales
Imprima los resúmenes globales del presupuesto, tiene tres opciones:
Datos generales: Imprime los datos generales del presupuesto.
Resumen del presupuesto estándar: Muestra el presupuesto y sus subpresupuestos con montos globales.
Resumen desagregado: Es el resumen total del presupuesto.
Use el botón del escenario de Datos Generales
3.11.2 Reportes desde el escenario de la Hoja del Presupuesto
Para el subpresupuesto: De la barra de botones use el botón para mostrar en pantalla el reporte.
Estándar interno: Imprime las partidas del subpresupuesto elegido, se caracteriza por que muestra el código con el que está registrada la partida dentro del Catálogo de partidas.
Estándar cliente:
Desagregado precios unitarios:
Imprime las partidas del subpresupuesto en un formato horizontal donde muestra los componentes que corresponde a mano de obra, materiales, equipos y subcontratos de cada una de las partidas del subpresupuesto.
Análisis de precios unitarios:
En el árbol del escenario, ubique el cursor sobre el subpresupuesto, imprime todos los análisis de precios unitarios, están disponibles los siguientes reportes:
Formato entidades: Disponible para las entidades que lo solicitaron.
Resumen de partidas iguales (Descripción propia): Sí una partida interviene en varios ítems de un subpresupuesto, se imprime una sola vez con la descripción propia y con la indicación de todos los ítems donde participa.
Resumen de partidas iguales (Descripción catálogo): Sí una partida interviene en varios ítems de un subpresupuesto, es impreso una sola vez con la descripción del catálogo y con la indicación de todos los ítems.
El S10 no guarda los reportes, sí necesita archivar o enviarlos por e-mail, use exportar informe a partir de una vista previa.
Luego en las opciones de exportación: Esta ventana depende del tipo de Windows disponible en su PC.
Llamada a: Permite enviar el contenido del reporte a un archivo de Excel.
Llamada b: Permite enviar el contenido del reporte a un archivo de Word.
Luego elija la capeta donde se grabará el archivo, e ingrese el nombre.
Use el explorador de Windows, para visualizar los archivos guardados.
Análisis de precios unitarios para el presupuesto:
En el árbol del escenario, ubique el cursor sobre el presupuesto y use el botón
.
Análisis de precios unitarios para el subpresupuesto: En el árbol del escenario, ubique el cursor sobre el subpresupuesto y use el botón .
3.11.2.1 Recursos y precios
Para acceder al listado de recursos y precios que utiliza el subpresupuesto use el botón que aparece en la barra de botones del escenario Hoja del Presupuesto.
a) Impresión de recursos:
Cuando lo haga desde un subpresupuesto sólo imprime los recursos que utiliza el subpresupuesto.
NOTA: El monto presupuestado que aparece en la hoja del presupuesto, no coincide con el del listado de recursos, por lo redondeos. En el grupo PRESUPUESTO DESCOMPUESTO, la vista preliminar o impresión de los recursos tiene implementado las opciones de redondeo, donde puede imprimir los recursos donde sí coinciden al 100% los montos del presupuesto y listado de recursos.
b) Para imprimir las subpartidas que participan sólo en el subpresupuesto, entrando en la ventana de Recursos y precios:
3.11.3 Precios por Presupuesto
Para imprimir todos los recursos y subpartidas del presupuesto (cuando el presupuesto tiene varios subpresupuestos), hágalo desde el escenario de Precios.
Acceso al escenario
Llamada a:
Cuando el cursor se ubica sobre el nombre del presupuesto lista todos los recursos y acumula las cantidades. Con clic derecho muestra el menú, para copiar los precios al portapapeles y luego pegar en otro presupuesto.
Llamada b:
Cuando el cursor se ubica sobre el nombre del subpresupuesto lista sólo los recursos y que participan en este.
3.11.3.1 Listado de recursos y precios
Llamada a:
Este botón despliega el menú para elegir la opción de presentación de recursos.
Llamada b:
Lista todos sólo los recursos que participan en el presupuesto.
Llamada c:
Lista todos sólo las subpartidas que participan en el presupuesto.
Llamada d:
Es un filtro que muestra los pecios de los recursos de acuerdo al mes, año y lugar.
Llamada e:
Es un filtro que contiene todos los índices unificados y de acuerdo a la selección lista los recursos asignados al índice unificado.
Llamada f:
Es un filtro que contiene los factores utilizados en los recursos del presupuesto, de acuerdo a este filtro muestra los recursos afectados por el factor.
4 OPCIONES
4.1 Ejemplos de diseño del pie de presupuesto
Para el presupuesto venta
Caso 1: Usando un porcentaje fijo
Caso 2: Usando la variable GGP
Para el presupuesto meta
4.2 Ejemplo de datos de gastos generales
4.3 Ítem alterno
El ítem alterno es una cadena alfanumérica definida por el usuario, que identifica un título o partida dentro de la hoja del presupuesto. Este ítem también sirve para ordenar los registros dentro de la hoja. Por defecto, la columna ALTERNO del listado, está oculta.
Para mostrar el campo del ítem alterno:
Para hacer visible el campo o columna.
Configuración para imprimir con ítem alterno. De la barra de botones use el botón que muestra:
4.4 Prioridad
A cada una de las partidas del presupuesto, permite asignar prioridad, con la finalidad de esconder o retornarlas a la hoja del presupuesto.
Al esconder una partida el costo no es considerado en el presupuesto.
Al retornar una partida el costo es considerado en el presupuesto.
Registro del catálogo de prioridades
Llamada 1:
Ingrese a catálogos.
Llamada 1-1:
Elija prioridades.
Llamada 1-1-1:
Registre las prioridades hasta en 10 niveles.
Ingrese al escenario de la hoja del presupuesto, para asignar prioridad a una partida.
Llamada 2:
Ubique el cursor sobre un subpresupuesto.
Llamada 2-1:
Elija la partida
Llamada 2-1-1:
Asigne prioridad del catálogo.
Para quitar o retornar partidas con prioridad a la hoja del presupuesto, hágalo desde el nombre del presupuesto o subpresupuesto ubicado en el árbol:
Llamada 3:
En el árbol, elija el nombre del presupuesto, y use clic derecho, del menú elija las opciones marcadas con las llamadas 3-1-1 o 3-1-2.
4.5 Registro de información en los catálogos
4.5.1 Registro de nuevos títulos en el catálogo
Para registrar el título “MUROS Y TABIQUES”
Ingrese al catálogo de títulos ya sea por el menú Catálogos o haciendo clic derecho en la hoja del presupuesto para luego elegir la opción de “Adicionar Título”.
Otra forma de registrar un título es haciendo clic derecho sobre un título existente y del menú que muestre elegir la opción Duplicar, para luego cambiar el título mostrado por el que desea registrar.
4.5.2 Registro de un grupo de partidas en el catálogo
Se recomienda que cada empresa registre las partidas que acostumbre usar, en el catálogo y lo tiene que hacer en un nuevo grupo, efectúe los registros en el árbol.
El botón indica que es un grupo de registros.
El botón indica que es el último nivel de registro en el árbol.
Para registrar “PARTIDAS DE R&R”
Para registrar la partida: Encofrado de elevación caravista.
Llamada 1:
El código es autogenerado por el sistema, pero puede ser ingresado por el usuario. Este código consta de 6 niveles, permitiendo ingresar 2 caracteres en cada nivel.
Llamada 2:
En descripción registre el nombre de la partida. (Hasta un máximo de 250 caracteres).
Llamada 3:
En descripción alterna permite registrar el nombre de la partida en un idioma alterno Ej. Inglés.
Llamada 4:
El botón muestra el catálogo de unidades de donde debe elegir la unidad que corresponda a la partida.
Llamada 5:
En especialidad permite asignar una especialidad a la partida. Ej., Edificaciones, Electromecánicas, etc. Previamente tiene que efectuar el registro en el catálogo que se activa al hacer clic en el botón y luego elegirlo con doble clic sobre el nombre registrado.
Llamada 6:
En grupo el registro puede ser Partida (partida común) o Estimada.
Partida tiene como componentes en la estructura de los análisis de precios unitarios a mano de obra, materiales, equipos y subcontratos.
Estimada tiene como componentes en la estructura de los análisis de precios unitarios a índices globales que dependen del % del precio unitario de la partida.
Llamada 7:
En tipo elija sin asignar. Al elegir por Ej. Mano de obra, toda la partida se comportará como mano de obra.
Llamada 8:
En estado la primera vez el sistema lo califica Sin análisis.
Llamada 9:
En jornada diaria es el número de horas que dura la jornada de trabajo diario, el sistema por defecto asume 8.
Llamada 10:
En peso cuando la partida que se registra es para electromecánicas, se ingresa el peso.
Llamada 11:
El botón adicionar graba el contenido de la ventana y lo cierra.
Registro del rendimiento
Llamada 1:
Metrado Total, cuando invoque al rendimiento desde el escenario de la hoja del presupuesto, sí la partida está siendo utilizada en varios ítems de la misma hoja, suma los metrados. Esta opción es informativa.
Llamada 2:
Permite modificar el número de cuadrillas, el mismo que altera la duración de la partida en días calendarios y por consiguiente el rendimiento es modificado. Al aceptar este rendimiento es trasladado a la hoja del análisis de precios unitarios.
Llamada 3:
Ingrese el rendimiento de la partida que corresponde a la mano de obra.
Llamada 4:
Ingrese el rendimiento que corresponde al equipo. Por defecto el sistema registra el mismo rendimiento para mano de obra y equipo, salvo que por configuración se le indique rendimientos diferentes.
Llamada 5:
Rendimiento para transporte, sí tiene un volumen bastante grande de eliminación de tierra, es necesario buscar un equilibrio entre la capacidad del cargador frontal y el número de volquetes, la idea es tener un equilibrio para no tener horas muertas
Llamada 5a:
El botón , calcula el rendimiento para transporte, use este botón todas las veces que cambie los datos hasta encontrar lo deseado.
Llamada 5b:
Ingrese la distancia de transporte en km.
Llamada 5c:
Se refiere a la velocidad que el vehículo imprimirá al ir a dejar el material de desmonte.
Llamada 5d:
Se refiere a la velocidad que el vehículo vacío imprimirá al regresar a la obra.
Llamada 5e:
Ingrese el tiempo en minutos que demora el cargador frontal en llenar el camión volquete.
Llamada 5f:
Ingrese el tiempo en minutos que demora el camión en descargar.
Llamada 5g:
Ingrese el porcentaje de eficiencia, del 1 al 100.
Llamada 5h:
Ingrese el factor de esponjamiento del material que se traslada.
Llamada 5i:
Ingrese la capacidad del camión volquete en m3.
Llamada 5j:
Ingrese el número de volquetes que se utilizará para eliminar el material, este dato se ingresa reiteradas veces hasta encontrar el número ideal.
Llamada 5k:
Este dato es informativo, se refiere al tiempo que demora el camión volquete en dejar el material y regresar a la obra.
Llamada 5L:
Este dato es informativo, es la cantidad de viajes en un día de trabajo.
Llamada 5m:
Este dato es informativo, es la cantidad total en m3 que traslada el camión volquete en un día.
Llamada 5n:
Rendimiento total diario se refiere a la capacidad del cargador frontal, este dato es el mandante para el cálculo del rendimiento para transporte.
Llamada 5o:
El rendimiento efectivo diario es el resultado aplicado por el factor de eficiencia.
Llamada 5p:
Cuando esté de acuerdo con el resultado del cálculo, use este botón para aceptar el rendimiento para transporte.
Llamada 6:
Opción no disponible en esta versión.
Llamada 7:
Graba y cerrar la ventana.
Registro de recursos en el análisis de precios unitarios
Datos que serán registrados:
Llamada a:
Para registrar la cuadrilla unitaria, es indispensable tener registrado el rendimiento.
Llamada b:
Registre la cantidad de recursos que necesita para efectuar una unidad, mostrada con la llamada b-1. Considera un desperdicio razonable. No ingrese precios.
4.5.3 Registro de recursos en el catálogo
Todos los recursos que utiliza el proyecto deben estar registrados en el catálogo de recursos necesariamente.
Datos para registrar
Ingrese a Catálogos, luego Recursos Genéricos y Específicos.
Luego use el botón
5 ESCENARIOS Y MENUS
5.1 Grupo presupuestos
5.1.1 Datos generales
Botones del escenario
: El sistema graba por defecto el registro del presupuesto, use este botón sólo para grabar los cambios.
Opciones del escenario:
Llamada 1:
Permite el registro de un nuevo presupuesto.
Llamada 2:
Permite ubicar un presupuesto en el árbol de presupuestos.
Llamada 3:
Copia el contenido del presupuesto al portapapeles, use para tener una copia o «juntar» varios presupuestos en uno solo.
Llamada 4:
Previamente el presupuesto destino debe ser creado y luego haber seleccionado el presupuesto origen con la opción COPIAR PRESUPUESTO. En caso de que el presupuesto destino tenga contenido pregunta si desea sumar o acumular al contenido existente. Si responde NO, previamente el contenido del presupuesto destino será borrado, al elija Sí el contenido del presupuesto destino es acumulado.
Esta opción permite armar un presupuesto sumando partes. El contenido del presupuesto destino habrá variado.
Llamada 5:
Duplica el presupuesto en la misma rama es conveniente cambiarle la descripción. Se recomienda convertir las partidas en propias.
Llamada 6:
Esta opción duplica el presupuesto, en otro grupo.
Llamada 7.
Registra una copia igual al presupuesto en el mismo grupo al del presupuesto origen.
Llamada 8:
Al eliminar un presupuesto es trasladado a la PAPELERA DE RECICLAJE, del que posteriormente pueden ser recuperados o eliminados definitivamente de la base de datos, el presupuesto a eliminar no debe estar asignado a un proyecto.
Llamada 9:
Es un utilitario que copia los precios para el nuevo lugar y fecha.
Llamada 10:
Permite modificar el membrete que aparecerá en el encabezado de los reportes.
Llamada 11:
Muestra el submenú con las siguientes opciones,
Llamada 11-1:
Permite definir los ítems que son utilizados en la hoja del presupuesto.
Llamada 11-2:
Alcance, se refiere al acceso del presupuesto por parte de los usuarios;
Normal, el presupuesto es visualizado por el usuario que registra, el “sa” y el jefe de grupo.
Público, todos los usuarios que tienen acceso al módulo de presupuestos visualizan el presupuesto.
Privado, en implementación.
Llamada 11-3:
El usuario que registra el presupuesto, invita a los demás usuarios para que tendrán acceso al presupuesto.
Del catálogo de usuarios, elija a los que tendrán acceso al presupuesto.
Llamada 11-4:
Permite asignar logotipo al presupuesto, que aparecerá como encabezado en los reportes del presupuesto.
Llamada 12:
Esta opción es informativa muestra información sobre el presupuesto, como fecha de registro, el usuario, moneda etc.
Llamada 13:
Traslada el presupuesto a otra rama del árbol.
ESCRITORIO, aquí se ubican los presupuestos que serán editados.
OBRAS GANADAS, cuando va a cerrar un presupuesto.
BANDEJA, cuando se dejará de usar temporalmente el presupuesto.
ARCHIVO CENTRAL, cuando ya no utilizará más el presupuesto.
PAPELERA DE RECICLAJE, aquí se almacenan los presupuestos a eliminar.
5.1.1.1 Hoja del presupuesto
Botones disponibles en la hoja del presupuesto
Llamada 1:
El botón Mover ver, muestra el menú
Llamada 1-1:
Muestra la hoja del presupuesto en toda la pantalla sin división.
Llamada 1-2:
Divide el escenario de la hoja del presupuesto, en la ventana inferior muestra los análisis de precios unitarios.
Llamada 1-3:
En el sector de los análisis de precios unitarios muestra el consolidado o resumen de recursos de la partida elegida en la ventana superior.
Llamada 1-4:
En el sector de los análisis de precios unitarios muestra el consolidado o resumen de recursos con las cantidades e importes afectados por el metrado para la partida elegida en la ventana superior.
Llamada 2:
Desplaza a la izquierda o disminuye la sangría en el registro seleccionado sea este título o partida permitiendo hacer las jerarquías.
Llamada 3:
Desplazar a la derecha, indenta o incrementa la sangría a la partida o título seleccionado o al grupo seleccionado, es decir se desplaza un nivel más.
Llamada 4:
Desplaza hacia abajo una fila el registro seleccionado. Esta opción NO aplica a una selección múltiple, es uno por uno. Luego de reubicar el título o partida, tiene que generar los ítems.
Llamada 5:
Desplaza hacia arriba una fila sea el título o partida seleccionada, Esta opción NO aplica a una selección múltiple.
Llamada 6:
Genera los ítems para los títulos y partidas de la hoja del presupuesto
Cuando exista un error en el indentado de alguno de los títulos o partidas, se enviará un mensaje y no se podrá generar los ítems hasta que corrija el error.
La columna ITEM en el listado no es editable, si requiere que los ítems sean registrados por el usuario, utilice la columna [Ítem Alterno] de la hoja del presupuesto.
Llamada 7:
Procesa, los costos unitarios por partida, en base a los análisis de precios unitarios y los precios de los recursos. Además, se calcula:
• Costos por tipo de recurso para cada una de las partidas.
• Costo directo e indirecto por subpresupuesto y por presupuesto.
• Cantidades e importes por recurso para el presupuesto.
• Consolidado de recursos.
• Importes o parciales de títulos.
El procesamiento tipo 1: Use cuando trabaje con análisis de precios unitarios tipo 2 (cuando los rendimientos son altos – más de 3,500 – ej.: movimiento de tierras) y cuando use partidas estimadas.
El procesamiento tipo 3: Use para presupuestos de edificaciones y obras de arte.
Llamada 8:
Esta opción muestra el listado de recursos y subpartidas utilizados en el subpresupuesto.
Imprimir: Envía el reporte a la impresora previamente determinada desde Windows.
Vista preliminar: Muestra en pantalla el reporte.
Llamada 8-1:
Después de procesar el presupuesto use este botón para mostrar u ocultar los recursos comodín (aquellos que tiene como unidad %).
Llamada 8-2:
Después de aplicar un factor al recurso, calcula las cantidades.
Llamada 8-3: A
ntes de usar este botón, exporte el contenido de la ventana a Excel, luego ingrese el precio de los recursos en la hoja Excel, grabe y cierre Excel, luego use el botón para tomar los datos vinculados y del menú que muestre.
Llamada 8-4:
Calcula el porcentaje de participación del recurso con relación al precio total del subpresupuesto.
Llamada 8-5:
Muestra el listado de recursos agrupado por grupos.
Llamada 8-6:
Graba los recursos y precios, para ser visualizados en el escenario de Precios Generales.
Llamada 8-7:
Esta opción está disponible sólo si previamente se usó la opción de copiar (en el escenario de precios generales).
Llamada 8-8:
Tipo: Permite agrupar los recursos de acuerdo a la selección mostrada en el menú.
Llamada 8-9:
Filtro: Permite agrupar los recursos de acuerdo a:
Llamada 8-10:
En el escenario de Datos Generales, se elige la moneda que utilizará el presupuesto, esa moneda es la principal. Cuando se define que el presupuesto sea en doble moneda, entonces habrá una moneda principal y otra secundaria.
Principal:
Lista los precios con los precios en moneda principal.
Secundaria: Lista los precios con los precios en moneda secundaria.
Llamada 8-11:
Cierra la ventana.
Llamada 8-12:
Cargar: Permite elegir el tipo de dato en la ventana de recursos y precios.
Recursos:
Lista los recursos que participan el subpresupuesto.
Subpartidas: L
ista las subpartidas que participan el subpresupuesto.
Llamada 8-13:
Todos los recursos que participan en el subpresupuesto están agrupados en índices unificados, esta opción filtra los recursos de acuerdo a la selección del filtro.
Llamada 8-14:
En el escenario de Datos generales, la opción de Datos Adicionales, se ingresan los factores.
Sin Factor:
Muestra los precios de acuerdo a lo registrado por el usuario.
Con cantidad subpresupuesto: La cantidad del recurso es multiplicado por el factor de presupuesto.
Llamada 8-15:
Muestra los precios históricos del recurso seleccionado utilizados en otros presupuestos.
Llamada 8-16:
Permite afectar por un factor al precio del recurso seleccionado, esta acción es temporal o definitiva.
Llamada 8-17:
Permite afectar por un factor de cantidad al recurso seleccionado, esta acción es temporal o definitiva
Llamada 8-18:
Permite modificar el nombre del recurso seleccionado.
Llamada 8-19:
Multiplica por un factor el recurso seleccionado, afecta a la cantidad, este proceso es irreversible.
Llamada 8-20:
La ventana de recursos y precios tiene 3 columnas auxiliares, donde permite el registro de cantidades relacionadas con el precio del recurso y esta opción calcula el precio sumando el contenido de las 3 columnas auxiliares.
Llamada 8-21:
Reajusta los precios de los recursos de acuerdo a los índices unificados seleccionados.
Llamada 8-22:
Vincula la celda donde está registrado el precio de un recurso, con el contenido de una celda de Excel.
Llamada 8-23:
Deshace la acción de la llamada 8.22.
Llamada 8-24:
Sí ejecutó la orden de la acción de la llamada 8-22, esta opción toma el precio vinculado.
Llamada 8-25:
Muestra el utilitario para vincular las celdas de los precios, con las celdas de la hoja en Excel, con la finalidad de compatibilizar las filas y las columnas.
Llamada 8-26:
Al elegir un recurso de la lista y hacer doble clic, muestra las partidas donde participa el recurso.
Llamada 8-27:
Muestra la ventana con los campos disponibles para esconder o hacer visible.
Llamada 8-28:
Exporta el contenido de la ventana a una hoja Excel.
Llamada 8-29:
Muestra el menú con las opciones de procesos especiales del recurso elegido.
Llamada 8-29-1:
Elimina un recurso del presupuesto y no existe deshacer.
Llamada 8-29-2:
La cantidad ingresada modifica la cantidad en el análisis de precios unitarios donde participa el recurso.
Llamada 8-29-3:
Cambia el recurso seleccionado por otro del mismo presupuesto, las cantidades son acumuladas luego de procesar o usar el botón marcado con la llamada 8-2.
Llamada 8-29-4:
Cambia el recurso seleccionado por otro del catálogo general de recursos.
Llamada 8-29-5:
Cambia un recurso comodín utilizado en el presupuesto por otro del catálogo.
Llamada 8-29-6:
Cambia el recurso seleccionado por una subpartida.
Llamada 8-29-7:
No disponible en esta versión
Llamada 8-30:
Permite insertar notas al registro.
Llamada 8-31:
Muestra las propiedades del registro.
Llamada 9:
Calcula el porcentaje de incidencia de cada uno de los títulos o partidas con respecto al total del presupuesto (costo directo). Para ver los resultados, haga visible el campo correspondiente a % de inc.
Llamada 10:
Muestra el detalle del costo indirecto para el subpresupuesto, según el diseño del pie.
Llamada 11:
Esta opción es informativa, muestra las partidas con sus respectivos análisis de precios unitarios donde falta datos.
Llamada 12:
Permite filtrar el listado de títulos o partidas por criterios pre-definidos.
Llamada 13:
Permite filtrar el listado de títulos o partidas por niveles.
Se muestran en listado aquellos títulos y partidas hasta un nivel específico.
Con la opción * (asterisco) se muestran todos los niveles.
Llamada 14:
Utilice esta opción cuando tenga la necesidad de MOVER un grupo de títulos y/o partidas a otra posición de fila dentro del mismo listado o inclusive a otro subpresupuesto o presupuesto.
Si el título o partida está siendo utilizada en algún proyecto del módulo de Gerencia de Proyectos no será posible cortarlo, muestra un mensaje indicando el motivo.
Llamada 15:
Copia al portapapeles la selección.
Llamada 16:
Pega el contenido del portapapeles.
Llamada 17:
Suma en forma vertical la información.
5.1.1.1.1 Opciones en el árbol para el presupuesto
Llamada 1: Permite el registro de un nuevo subpresupuesto. El sistema autogenera el código para el nuevo subpresupuesto. Ingrese un texto para la descripción y que no exceda de 250 caracteres.
Llamada 2: Elimina un subpresupuesto siempre y cuando no tenga contenido, caso contrario elimine previamente las partidas.
Llamada 3: Esta opción permite modificar mediante factores las cantidades de los recursos dentro del análisis de precios unitarios, incluye también factores que afecten a las subpartidas.
RECOMENDACIONES: Sí usa factores menores de 1, procure que sea dentro del rango 0.10 a 0.90
También puede usar valores mayores a 1. Después de aplicar los valores procese el presupuesto.
LIMITACIONES: Cuando aplique factores a mano de obra, materiales, equipos, subcontratos NO APLIQUE A SUBPARTIDAS.
PARA DESHACER: Use el botón DESHACER y luego SALIR.
Llamada a: El factor aplicado queda fijo, ya no permite deshacer.
Llamada 4: Permite modificar los rendimientos temporalmente mediante factores para todas las PARTIDAS PROPIAS del subpresupuesto.
RECOMENDACIONES: Sí usa factores menores de 1, procure que sea dentro del rango 0.10 a 0.90. También puede usar valores mayores a 1. No use CERO como factor.
PARA DESHACER: Use el botón DESHACER y luego SALIR.
Cuando los factores han sido aplicados a la partida, el análisis de precios unitarios no puede ser modificado hasta deshacer.
Llamada 5: Procesa el subpresupuesto de acuerdo al submenú.
Llamada 5-1: Procesa todos los subpresupuestos.
Llamada 5-2: Procesa sólo los subpresupuesto inconclusos.
Llamada 6: Muestra el submenú con opciones de exportación.
Llamada 6-1: Envía el contenido del presupuesto al MS Project, por defecto a la versión 2003, y por configuración al 2007 o superior. La información no retorna a la hoja del presupuesto.
Llamada 6-2: Envía el contenido del presupuesto a un archivo de Excel.
Llamada 6-2: Envía los recursos del presupuesto a una nueva base de datos, para ser utilizado por el módulo de valorizaciones del S10 versión 2.0
Llamada 7: Convierte a propias todas las partidas del presupuesto.
Llamada 8: Deshace la opción de la llamada 7.
Llamada 9: Las partidas formato son las que aparecen en la hoja del presupuesto como una línea descriptiva y que no tienen análisis de precios unitarios. Esta opción convierte a partidas formato todo el contenido del presupuesto.
Llamada 10: Genera el ítem alterno según el código con el que está registrada la partida dentro del catálogo, está disponible también para los títulos de un subpresupuesto. Para mostrar esta columna debe acceder a la opción Mostrar Campos.
Indique el número de dígitos que se tomará del código del título o partida.
Llamada 11: Muestra el submenú,
Llamada 11-1: Quita del presupuesto las partidas que tienen asignadas una prioridad, el sistema muestra el catálogo de prioridades para elegir la prioridad a quitar.
Llamada 11-2: Retorna las partidas al presupuesto, del catálogo de prioridades se elige las que retornarán.
Llamada 11-3: Esta opción es informativa.
Llamada 12: Muestra el submenú con opciones que se aplican sobre las partidas propias de todo el presupuesto.
Llamada 12-1: Muestra las partidas del presupuesto. Además, permite filtrar las partidas por los siguientes criterios:
Llamada 12-1-1: Son las que están en el presupuesto, como partidas o subpartidas
Llamada 12-1-2: Lista las partidas propias de este presupuesto que no se utilizan en el presupuesto. Al usar esta opción se activa el botón marcado con la llamada 12-1-6 para eliminarlas.
Llamada 12-1-3: Muestra todas las partidas propias de este presupuesto, la que se están utilizando y las que no se utilizan.
Llamadas 12-1-4: Muestra las partidas con descripciones propias de este presupuesto.
Llamada 12-1-5: Muestra las partidas con unidades propias.
Llamada 12-1-6: Quita o elimina en forma definitiva de la base de datos, es importante eliminar las partidas propias que no se utilizan.
Llamada 12.2: Muestra las partidas del presupuesto que están siendo utilizadas en más de un ítem, sumando sus metrados.
Llamada 12-3. Permite registrar el tiempo unitario para recursos; se refiere al tiempo necesario para ejecutar una unidad en los recursos tipo subcontrato, se usa en el módulo de Gerencia de proyectos.
Llamada 12-4. El precio de un recurso debe ser el mismo para todos los subpresupuestos, caso contrario utilice esta opción para corregir este problema.
Llamada 12-5: Es un utilitario que muestra las partidas utilizadas en el presupuesto cuyo análisis de precios unitarios contiene a la misma partida como subpartida.
Llamada 12-6: Es un utilitario que calcula las incidencias para todos los recursos de tipo mano de obra y equipos en los análisis de precios unitarios de todas las partidas propias del presupuesto. Las incidencias o cantidades se calculan en función a la jornada de la partida, el rendimiento y la cuadrilla
Llamada 12-7: Esta opción calcula las cuadrillas para todos los recursos de tipo mano de obra y equipos en los análisis de precios unitarios de todas las partidas propias del presupuesto.
Llamada 12-8: Esta opción verifica el estado del análisis de precios unitarios de las partidas utilizadas en el presupuesto. Ejemplo, una partida puede estar en estado SIN ANALISIS, a pesar de tener un análisis de precios unitarios. Con esta opción al detectarse este caso automáticamente la partida pasa a un estado de SIN REVISION.
Generalmente estos casos ocurren cuando hemos importado o trasladado información de una versión anterior a esta.
Llamada 12-9: Este utilitario cambia de estado a todas las partidas propias del presupuesto, de acuerdo al siguiente submenú,
Llamada 13: Muestra el submenú,
Llamada 13-1: Permite cambiar la estructura de los ítems sólo para el presupuesto en uso.
Llamada 13-2: Permite adicionar usuario al grupo de trabajo.
Llamada 13-3: Define el modo de acceso al presupuesto, restringiéndolo a un grupo o a todos los usuarios, según como se defina.
NORMAL, solo el usuario que registró el presupuesto y el grupo de trabajo tienen acceso al presupuesto.
PUBLICO, el presupuesto es visualizado por todos los usuarios del sistema.
PRIVADO, en implementación
Por defecto el presupuesto se crea con alcance NORMAL
Llamada 13-4: Permite asignar logotipo al presupuesto.
Llamada 13-5: Muestra la ventana de Datos adicionales, descrito en el escenario de Datos generales.
Llamada 14: Muestra información relacionada del presupuesto.
Llamada 15. Traslada el presupuesto de un grupo a otro en el árbol, de acuerdo al submenú marcado la llamada 15-1.
5.1.1.1.2 Opciones en el árbol para el subpresupuesto
Llamada 1: Permite modificar la descripción (nombre) del subpresupuesto.
Llamada 2: Copia el contenido del subpresupuesto al portapapeles.
Llamada 3: Esta opción está disponible sólo si previamente ejecutó la orden de la llamada 2, Pega el contenido del portapapeles sobre un subpresupuesto destino, muestra el mensaje:
En caso de que el subpresupuesto destino tenga contenido pregunta si desea adicionar al contenido existente. Si responde NO, previamente el contenido del subpresupuesto destino será borrado.
Si desea copiar solo un grupo de títulos y/o partidas, seleccione el rango.
En la ventana que se muestra a continuación se elige el ítem inicio e ítem final.
En este caso el rango de ítems se adiciona al final del subpresupuesto.
Llamada 4: Elimina un subpresupuesto siempre y cuando no tenga contenido Si el subpresupuesto que intenta eliminar tiene contenido le envía un mensaje, que indica “Eliminar previamente las Partidas de este Subpresupuesto”.
Llamada 5: Permite eliminar todas las partidas y títulos de un subpresupuesto.
Llamada 6: Permite eliminar partidas de un subpresupuesto, cuando el metrado es igual al valor ingresado.
Llamada 7: Esta opción funciona de manera similar como la descrita para el presupuesto, pero esta vez afecta sólo a las partidas del subpresupuesto elegido.
Llamada 8: Esta opción funciona de manera similar como la descrita para el presupuesto, pero esta vez afecta sólo a las partidas del subpresupuesto elegido.
Llamada 9: Esta opción permite modificar la jornada diaria de trabajo para todas las partidas propias del subpresupuesto.
Llamada 10: Permite exportar el contenido del subpresupuesto, de acuerdo al submenú,
Llamada 10-1: Exporta el contenido del presupuesto al MS Project. (La información no retorna al S10).
Llamada 10-2: Exporta el contenido del presupuesto a un archivo de Excel.
Llamada 11: Convierte a propias todas las partidas de un subpresupuesto.
Llamada 12: Deshace la acción marcada con la llamada 11.
Llamada 13: Las partidas formato son las que aparecen en la hoja del presupuesto como una línea descriptiva y que no tienen análisis de precios unitarios. Esta opción convierte a partidas formato todos los títulos y partidas de un subpresupuesto.
Llamada 14: Cuando una partida tiene dentro del análisis de precios unitarios varias subpartidas, acumula las cantidades de los subniveles al análisis principal, eliminado las subpartidas.
Llamada 15: Desvincula la celda de metrados del S10, siempre y cuando tengo vínculo.
Llamada 16: Cuando previamente se vinculó una celda de metrado del S10 con una celda de Excel que contiene metrado, toma los metrados.
Llamada 17: Sí las partidas tienen asignado prioridad, permite quitar del subpresupuesto las partidas según la prioridad.
Llamada 18: Está disponible sólo si previamente se usó la opción marcada con la llamada 17; hace que las partidas retornen al presupuesto según la prioridad elegida.
Llamada 19: Muestra la ventana de propiedades, con opciones informativas acerca del subpresupuesto.
5.1.1.1.3 Opciones en la hoja del presupuesto
En la hoja del presupuesto se encuentran los títulos y partidas. A continuación, se describen las opciones del menú.
Llamada 1: Muestra el Catálogo de Títulos para elegir e insertarlos en la hoja del presupuesto.
Llamada 2: Muestra el Catálogo de partidas para elegir e insertarlos en la hoja del presupuesto. Antes de utilizar esta opción es importante que verifique la configuración, específicamente las opciones para hacer propias las partidas. La partida adicionada toma el mismo nivel (indentado) del ítem de la fila actual.
Llamada 3: Muestra el Catálogo de Favoritos, para insertarlo en la hoja del presupuesto.
Favorito es un grupo de títulos y/o partidas pre seleccionados y agrupados por el usuario.
Un favorito puede ser el subpresupuesto íntegro de una obra, no considere los metrados.
El registro de un favorito es similar al de la hoja del presupuesto.
Elija un favorito y el botón SELECCIONAR lo traslada a la hoja del presupuesto.
Llamada 4: Permite reemplazar un título o partida por otro del Catálogo, al hacer clic derecho sobre un registro en la hoja del presupuesto muestra el menú,
Llamada 4-1: Reemplaza la partida del presupuesto por una del catálogo.
Llamada 4-2: Reemplaza la partida elegida por otra del mismo presupuesto.
Llamada 4-3: Reemplaza la partida elegida por otra del catálogo de partidas – partidas propias.
Llamada 4-4: Reemplaza un registro de la hoja del presupuesto por un título del catálogo.
Llamada 4-5: Convierte el registro elegido en formato, (sí es partida elimina sus análisis de precios unitarios).
Llamada 5: Muestra una ventana donde el usuario puede ingresar títulos y partidas como si fuera una hoja Excel.
Luego, la información es mostrada en la hoja del presupuesto.
Para reasignar análisis de precios (del catálogo) a la partida:
Llamada 6: Importa el contenido de la hoja Excel. Ver: Caso 3 – formas de armar la hoja del presupuesto (Cuando se tiene presupuestos elaborados en hoja Excel).
Llamada 7: Permite elegir en nivel de registro, para contraer o expandir.
Llamada 8: Calcula el precio unitario de la partida elegida.
Llamada 9: Al cortar un registro de la hoja del presupuesto, lo coloca en el portapapeles. Utilice esta opción para MOVER un grupo de títulos y/o partidas a otra posición.
Si el título o partida está siendo utilizada en algún proyecto del módulo de Gerencia de Proyectos no será posible cortarlo, muestra un mensaje.
Llamada 10: Copia la selección al portapapeles.
Llamada 11: Pega el contenido del portapapeles.
Llamada 12: Al pegar el contenido del portapapeles muestra la ventana,
Llamada 13: Disponible cuando en el portapapeles se tiene los metrados, elija la celda a partir del cual se pegarán los metrados.
Llamada 14: Elimina el registro previamente elegido de la hoja del presupuesto.
Llamada 15: Al duplica genera una copia idéntica del registro ubicado en la hoja del presupuesto.
Llamada 15-a: Al duplicar genera una copia de una partida de la hoja y luego la convierte en propia.
Llamada 16: Aplica factores de cantidad a los recursos dentro del análisis de precios unitarios.
Son valores que modifican temporalmente o definitivamente las cantidades de los recursos por tipo (mano de obra, materiales, equipos, subcontratos y subpartidas).
Llamada 17: Aplica factor de rendimiento a la partida elegida.
Llamada 18: Permite modificar la jornada diaria de trabajo para la partida elegida, sólo está disponible para las partidas propias del subpresupuesto.
Llamada 19: Esta opción permite acceder a un submenú,
Llamada 19-1: Convierte la partida elegida en propia.
Llamada 19-2: Permite ingresar una descripción particular para el título o partida seleccionada. También puede acceder a esta opción digitando cualquier carácter en la celda que corresponde a la descripción.
Cuando la descripción excede los 250 caracteres, tiene que almacenar esta en las NOTAS, así no tendrá límite en la cantidad de caracteres. La primera nota es tomada como la descripción de la partida. Para que en el reporte muestre estas descripciones largas, tiene que modificar la configuración de impresión,
Llamada 19-3: Permite ingresar una unidad particular para la partida. Esta unidad no se modificará en el catálogo. La modificación de la unidad de la partida debería estar acompañada del cambio de las incidencias del análisis de precios unitarios, estos cambios si son necesarios, deben realizarlos manualmente, porque esta opción está preparada para que ingrese la unidad en forma de texto.
Llamada 19-4: Permite asignar prioridad a la partida.
Llamada 19-5: Permite asignar especialidad a la partida.
Llamada 19-6: Si los metrados de las partidas están en hojas de Excel, esta opción permite vincular o relacionar la celda de metrado de la partida con una celda de Excel donde se encuentra el metrado.
Llamada 19-7: Esta opción está disponible sólo si usó previamente la opción marcada con la llamada 19-6, toma los datos del Excel.
Llamada 19-8: Elimina el vínculo efectuado en la llamada 19-6.
Llamada 19-9: Copia la selección al portapapeles, (títulos y partidas).
Llamada 20: Muestra el submenú,
Llamada 20-1: Muestra la ventana con los campos disponibles, con la finalidad de hacer visible o esconder.
Llamada 21: Esta opción es informativa, con información del registro elegido.
5.1.1.1.4 Opciones en los análisis de precios unitarios
El análisis de precios unitarios de una partida es la estructura de costo y puede estar conformado por sólo por recursos y/o recursos + subpartidas o solo subpartidas, al hacer clic derecho muestra el menú,
Llamada 1: Invoca al catálogo recursos, para adicionar recursos al análisis de precios unitarios. A los recursos tipo mano de obra se ingresa la cuadrilla siempre y cuando tenga rendimiento; para materiales se ingresa cantidad para una unidad, en equipos también se registra la cuadrilla y para herramientas manuales el porcentaje en cantidad.
Llamada 2: Permite adicionar al análisis de precios unitarios recursos que se viene utilizando en el mismo presupuesto.
Llamada 3: Permite adicionar al análisis de precios unitarios una o más partidas seleccionadas desde el catálogo, que toman el nombre de subpartidas.
Llamada 4: Invoca a una plantilla, que es una agrupación de recursos y subpartidas previamente registradas por el usuario; para adicionar al análisis de precios unitarios.
Llamada 5: Esta opción se habilita cuando el análisis de precios unitarios está vacío, muestra el catálogo de partidas para copiar el contenido del análisis de precios unitarios.
Llamada 6: Copia la selección al portapapeles.
Llamada 7: Corta la selección al portapapeles.
Llamada 8: Pega el contenido del portapapeles al análisis de precios unitarios.
Llamada 9: Pega el contenido del portapapeles al análisis de precios unitarios, según las opciones de la ventana,
Llamada 10: Elimina la selección del análisis de precios unitarios.
Llamada 11: Permite modificar el estado de la partida, equivale al uso del botón
La partida puede estar en uno de los siguientes estados:
SIN ANALISIS, cuando se crea una partida en el catálogo de partidas y solo se registra la descripción y unidad, el sistema calificará a esta partida con este estado.
ANALISIS COPIA, cuando la partida fue registrada a partir de otra.
SIN REVISION, cuando se hagan modificaciones a un análisis, el sistema le asignará en forma automática este estado.
REVISADO, es una forma de darle su visto bueno, con el fin de que no sea modificado por error o en forma involuntaria. Este estado puede ser asignado por el administrador del sistema, jefes de grupo y especialistas.
PROTEGIDO, indica que el análisis ha pasado todas las revisiones y podríamos decir que ya está «cerrada». La partida que se encuentre en este estado no podrá ser modificado. Este estado puede ser asignado solo por el administrador del sistema y Jefe de grupo.
La forma del icono representa alguna de las siguientes características de la partida:
Llamada 12: Algunos recursos pueden tener detalle, una especie de análisis de precios unitarios del recurso, compuesto exclusivamente por recursos, al hacer clic derecho sobre el recurso muestra el submenú.
Llamada 12-1: Activa el detalle del recurso, y muestra las opciones del submenú,
Llamada 12-1-1: Muestra el catálogo de recursos para elegir los recursos que formarán parte del detalle.
Llamada 12-1-2: Adiciona al detalle el contenido de la plantilla. (La plantilla debe ser elaborada previamente por el usuario.)
Llamada 12-1-3: Esta opción está disponible sólo cuando el recurso que se elige tiene detalle. Esta opción copia sólo el detalle del recurso origen para pegarlo en el que viene utilizando.
Llamada 12-1-4: Copia los recursos seleccionados al portapapeles.
Llamada 12-1-5: Corta la selección y lo graba en el portapapeles.
Llamada 12-1-6: Pega el contenido del portapapeles.
Llamada 12-1-7: Elimina el recurso del detalle.
Llamada 12-1-8: El costo calculado del detalle del recurso traslada al análisis de precios unitarios.
Llamada 12-1-9: Muestra el submenú,
Llamada 12-1-9-1: Sustituye por otro recurso, según las opciones marcadas con las llamadas 12-1-9-1-1 y 12-1-9-1-2.
Llamada 12-1-9-2: Elimina el recurso en todos los detalles.
Llamada 12-1-10: Importa el contenido de la hoja Excel al detalle.
Llamada 12-1-11: Exporta a un archivo Excel el contenido del detalle del recurso.
Llamada 12-1-12: Muestra información relacionada al recurso.
Llamada 12-1-13: Mostrar campos, utilice esta opción para mostrar u ocultar las columnas del detalle del recurso.
Llamada 13: Al hacer clic derecho sobre un recurso, muestra el submenú.
Llamada 13-1: Esta opción convierte a propio la subpartida seleccionada.
Llamada 13-2: Los análisis de precios unitarios utilizan 4 tipos de recursos; Mano de obra, Materiales, Equipo y Subcontratos y le proporciona subtotales por cada uno de estos grupos, pero muchas veces es necesario que las SUBPARTIDAS sean sumadas a uno de estos grupos o su comportamiento en la partida sea como uno de estos tipos de recursos, por lo que es necesario calificar la subpartida con uno de estos tipos
Llamada 13-3: Permite reemplazar en el análisis de precios unitarios un recurso o subpartida por otro recurso seleccionado del catálogo.ç
Llamada 13-4: Permite reemplazar en el análisis de precios unitarios un recurso o subpartida por otro recurso que se está utilizando en el presupuesto.
Llamada 14-5: Permite reemplazar en el análisis de precios unitarios un recurso o subpartida por otra subpartida seleccionada desde el catálogo.
Llamada 13-6: Reparto total, opción no habilitada.
Llamada 13-7: Esta opción permite asignar factor de cantidad al recurso elegido, la cantidad (aporte) es la que se modifica.
Llamada 13-8: Permite aplicar un factor al rendimiento sólo del análisis de precios unitarios donde se invoca.
Llamada 13-9: Precios históricos, muestra todos los precios que tiene los recursos, para todos los lugares y todos los presupuestos.
Llamada 14: Permite importar los recursos al análisis de precios unitarios desde una hoja Excel. Es importante que los códigos de los recursos coincidan con los del catálogo.
Llamada 15: Envía a un archivo de Excel los recursos del análisis de precios unitarios.
Llamada 16: Muestra en pantalla la apariencia como será impreso el análisis de precios unitarios.
Llamada 17: Envía el reporte a la impresora previamente elegida en Windows.
Llamada 18: Invoca al notero para insertar comentarios y archivos digitales.
Llamada 19: Muestra en una ventana las propiedades del recurso o subpartida seleccionado.
Llamada 20: Permite seleccionar las columnas que se mostrarán en el análisis de precios unitarios.
5.1.2 Diseño del pie de presupuesto
El diseño del pie es para el presupuesto venta.
Botones disponibles:
: La fila elegida cambia de lugar hacia abajo.
: La fila elegida cambia de lugar hacia arriba.
Opciones del escenario:
Llamada 1: Por defecto sólo está disponible para la hoja del presupuesto.
Llamada 2: Son variables obligatorias.
Llamada 2-1: El monto calculado o presupuestado es asignado a esta variable.
Llamada 2-2: La variable P_T convierte los números en letras y hace la suma vertical, use en la columna Variable, en la fila que corresponde al TOTAL DEL PRESUPUESTO.
Llamada 3: Son variables auxiliares, que pueden ser utilizadas en cualquier detalle del diseño del pie de presupuesto.
Llamada 3-1: La variable C_J, contiene el monto calculado de la mano de obra.
Llamada 3-2: La variable C_M, contiene el monto calculado de los materiales.
Llamada 3-3: La variable C_E, contiene el monto calculado de los equipos.
Llamada 3-4: La variable C_S, contiene el monto calculado de los subcontratos.
Llamada 3-5: La variable GGP o ggp, se utiliza en la columna macro en la fila que corresponde a gastos generales cuando se necesita realizar el cálculo analítico de gastos generales.
Llamada 3-6: Disponible sólo para Venezuela.
Llamada 4: Seleccionado, el diseño es utilizado por todos los subpresupuesto.
No seleccionado, para cada subpresupuesto se elabora el diseño del pie de presupuesto.
Llamada 5: Esta celda es informativa, toma el valor del GGP calculado.
Llamada 6: Esta fila viene por defecto y no debe ser eliminada tampoco cambiada de lugar, a la variable nDirecto se le asigna el monto calculado del presupuesto por defecto.
Llamada 7: Son ítems que el usuario debe registrar de acuerdo a sus necesidades, al hacer clic derecho muestra el menú,
Llamada 7-1: Elimina la fila elegida.
Llamada 7-2: Exporta el contenido de la ventana a un archivo Excel.
Llamada 7-3: Imprime en pantalla en contenido de la ventana.
Llamada 7-4: Envía el contenido de la ventana a la impresora previamente definida en Windows.
Llamada 8: Seleccionada, la fila que corresponde al impuesto no participa en la elaboración de la fórmula polinómica.
5.1.3 Gastos generales
Permite calcular el porcentaje y el monto de los gastos generales de acuerdo a los rubros y conceptos, disponible para el presupuesto venta.
Botones disponibles:
: Sí el cursor se ubica en la ventana superior muestra el catálogo de rubros, sí se ubica en la ventana inferior muestra el catálogo de conceptos, con la finalidad de elegirlo.
: Elimina el registro seleccionado.
: Procesa el porcentaje y costo de los gastos generales, es requisito que el cursor se ubique en la ventana superior.
Opciones del escenario:
Llamada 1:
El formato de ingreso de datos, define el comportamiento del tipo de rubro.
Llamada 2:
Este segmento es informativo, muestra las cantidades.
Llamada 2-1:
Este segmento es informativo, muestra el porcentaje de gastos generales calculado.
Llamada 3:
En este segmento se registra los rubros, al hacer clic derecho muestra el menú.
Llamada 3-1:
Muestra el catálogo de rubros para elegirlo.
Llamada 3-2:
Elimina el rubro elegido.
Llamada 4:
Del rubro elegido en la ventana superior, muestra los conceptos, al hacer clic derecho muestra el menú.
Llamada 4-1:
Permite adicionar conceptos, del rubro elegido en la ventana superior.
Llamada 4-2:
Elimina el concepto elegido.
Llamada 4-a:
Se registra el número de personas.
Llamada 4-b:
Se registra el porcentaje de participación de la persona en el proyecto.
Llamada 4-c:
Se registra la cantidad de meses.
Llamada 4-d:
Se registra el sueldo por mes.
5.1.4 Fórmula polinómica
Agrupamiento preliminar: Toma los IU asignados a los recursos que participan en el subpresupuesto con la finalidad de agruparlos entre semejantes, tomando el criterio de tomar el valor mayor al menor. Se agrupan los semejantes más dólar y herramientas manuales.
Botones disponibles: ; Refresca los datos en el escenario.
Opciones del escenario:
Llamada 1: Elija la pestaña de agrupamiento preliminar.
Llamada 1-1: La fórmula polinómica es por subpresupuesto.
Llamada 1-2: Es el número que corresponde al índice unificado.
Llamada 1-3: Es el nombre que corresponde al índice unificado.
Llamada 1-4: Es el porcentaje de participación del IU dentro del presupuesto.
Llamada 1-5: Es el porcentaje luego de agrupar entre IU.
Llamada 1-6: Muestra el agrupamiento efectuado por parte del usuario.
Llamada 1-7: Al hacer clic derecho sobre un IU, muestra el menú.
Llamada 1-7-1: Toma el valor del IU unificado elegido al portapapeles.
Llamada 1-7-2: Acumula al valor del IU unificado destino.
Llamada 1-7-3: Deshace la agrupación del IU elegido, ponga una selección por más que tenga un solo registro.
Llamada 1-7-4: Exporta el contenido de la ventana a un archivo Excel.
En conformación de monomios: se numeran los IU tomando las siguientes recomendaciones del reglamento peruano.
Para que un IU forme un monomio el factor debe tener un valor del 5% o 0.05 y se le asigna un número, caso contrario se agrupa hasta 3 IU y pasarán a formar submonomios, también se agrupa con la finalidad de tener solo 8 monomios.
Sólo permite tener 8 monomios.
Botones disponibles:
: Refresca los datos en el escenario.
: Cuando esté concluido la fórmula polinómica, verifica de acuerdo a la configuración.
Opciones del escenario:
Llamada 2: Elija la pestaña de conformación de monomios.
Llamada 2-1: Registre el número de monomio que le corresponde al IU dentro de la fórmula polinómica.
Llamada 2-2: Estas celdas son informativas, muestra el porcentaje de participación del IU dentro del presupuesto, para que un IU forme un monomio es necesario que tenga el valor del 5%.
Llamada 2-3: Estas celdas son informativas, cuando un IU forma un monomio tiene el 100%, al agrupar dos o más índices muestra el porcentaje de participación de cada uno de los submonomios.
Llamada 2-4: El sistema por defecto toma la primera letra del nombre del IU como símbolo (ver llamada 2-4-1), el usuario debe cambiarlo de acuerdo a su experiencia. Cuando se agrupa dos o tres submonomios solo el IU que tiene mayor porcentaje de participación acumula los símbolos (ver llamada 2-4-2), al hacer clic derecho muestra el menú.
Llamada 2-4-2-1: Exporta el contenido de la ventana a un archivo Excel.
Llamada 2-5: Estas celdas son informativas, muestra el número del IU.
Llamada 2-6: Estas celdas son informativas, muestra el nombre o descripción del IU.
5.1.5 Precios por presupuesto
Este escenario tiene la finalidad de listar todos los precios de los recursos y subpartidas que participan en un presupuesto. Existen dos escenarios que tienen las mismas opciones, el primero se ubica en el grupo de presupuestos que tiene todas las opciones de edición y la segunda se ubica en el grupo de precios que tiene la condición de informativa.
Botones disponibles:
Llamada 1: La barra de botones del escenario.
Llamada 1-1: Muestra o esconde los recursos tipo comodín.
Llamada 1-2: Toma los datos vinculados de Excel.
Llamada 1-3: Calcula las incidencias o porcentaje de participación de recurso dentro del presupuesto.
Llamada 1-4: Muestra los recursos agrupados en un árbol.
Llamada 1-5: Almacenes los precios en el Precios Generales (Grupo precios).
Llamada 1-6: Esta opción está disponible sólo si existe información en el portapapeles.
Llamada 1-7: Es un filtro que agrupa el contenido de la ventana de acuerdo al tipo de recurso.
Llamada 1-8: Es un filtro que agrupa el contenido de la ventana de acuerdo al precio de los recursos.
Llamada 1-9: Es un filtro que agrupa el contenido de la ventana de acuerdo al tipo de moneda.
Opciones del escenario:
Llamada 2: Este segmento contiene los filtros.
Llamada 2-1: Carga recursos o subpartidas que participan en el presupuesto.
Llamada 2-2: Muestra el mes, año y lugar disponible para los cuales hay precio.
Llamada 2-3: Filtra los recursos de acuerdo al IU asignado.
Llamada 2-4: Filtra los recursos de acuerdo al factor asignado.
Llamada 3: Lista los recursos que participan en el presupuesto, al hacer clic derecho sobre un recurso, muestra el menú,
Llamada 3-1: Esta opción es informativa, del recurso seleccionado, muestra los precios históricos con los que participó en otros presupuestos.
Llamada 3-2: Esta opción es informativa, del recurso seleccionado, muestra los precios de compra, disponible sólo sí el recurso se compró utilizando el módulo de Compras del S10.
Llamada 3-3: Al recurso seleccionado permite aplicarle un factor de precio con la finalidad de bajar o subir el monto, también aplica para un grupo de recursos.
Llamada 3-1-1: Ingrese el factor, en caso del ejemplo que se muestra, significa una rebaja del 2%. Luego procese el presupuesto.
Llamada a: Inicialmente muestra el precio original, luego con el factor aplicado.
Llamada 3-4: Al recurso seleccionado permite aplicarle un factor de cantidad con la finalidad de bajar o subir la incidencia dentro del análisis de precios unitarios. Luego procese el presupuesto.
Llamada 3-5: Muestra el submenú.
Llamada 3-5-1: Permite registrar una descripción propia al recurso.
Llamada 3-5-2: Elimina la descripción propia al recurso.
Llamada 3-6: Al recurso seleccionado permite aplicarle un factor al precio, este cambio es permanente.
Llamada 3-7: La grilla de precios contiene 3 columnas auxiliares donde se registra información que influye en el precio del recurso, puede considerar, precio en planta + manipuleo + transporte; este utilitario calcula el precio tomando en cuenta las 3 columnas.
Llamada 3-8: Sí en el escenario de Índices Unificados, se registró información para la fecha del presupuesto y para una fecha posterior, este utilitario reajusta los precios de acuerdo a los índices unificados.
Llamada 3-9: No disponible es esta versión.
Llamada 3-10: Este utilitario se usa cuando se cotizan precios en un lugar alejado del centro de trabajo, que puede ser otra ciudad y necesitamos usar una hoja Excel para cotizar los precios de recursos.
Use para vincular la celda donde se registra el precio del recurso del S10, a una celda de una hoja Excel done se encuentra el precio el recurso;
Procedimiento:
Exporte el contenido del escenario a una hoja Excel y grabe en el servidor.
Registre el precio del recurso en la hoja Excel, grabe el archivo y cierre Excel.
Haga clic derecho en la celda (S10) donde se registra el precio del recurso, luego el sistema muestra la ventana;
Llamada a: Ubique el archivo de Excel que contiene los precios de los recursos (llamada a-1).
Llamada 3-10-1: Muestra la celda que será vinculada.
Llamada 3-10-2: Elija la fila y registre en la celda marcada con la llamada 2-10-2-1.
Llamada 3-10-3: Elija la columna y registre en la celda marcada con la llamada 3-10-3-1.
Llamada 3-10-4: Vincula las celdas. Luego use la opción marcada con la llamada 3-12.
Llamada 3-11: Al desvincular un precio de celda Excel, el usuario puede registrar manualmente el precio.
Llamada 3-12: Reemplaza los precios de las celdas vinculadas a Excel.
Llamada 3-13: Es el mismo utilitario de la llamada 3.10, esta vez para varias celdas.
Use para vincular varias celdas donde se registran los precios de los recursos, a las celdas de una hoja Excel done se encuentra los precios de los recursos.
Procedimiento:
Exporte el contenido del escenario a una hoja Excel y grabe en el servidor.
Registre los precios de los recursos en Excel, grabe el archivo y cierre Excel.
Haga clic derecho donde se encuentran el listado de recursos. Luego,
Llamada 3-13-1: Identifique los recursos y los precios a vincular.
Llamada 3-13-2: Elija la fila inicio y fin, registre en las celdas marcadas con la llamada 3-13-2-1.
Llamada 3-13-3: Elija la columna y registre en la celda marcada con la llamada 3-13-3-1.
Llamada 3-13-4: Use este botón para grabar el contenido de la ventana.
Para tomar los datos de las celdas vinculadas use el botón (llamada 1-2), luego elija la opción TOMAR PRECIOS VINCULADOS EN FORMA GLOBAL
Llamada 3-14: Esta opción es informativa; del recurso elegido muestra las partidas donde participa de acuerdo al submenú marcado con las llamadas 3-14-1 y 3-14-2.
Llamada 3-15: Permite hacer visible o esconder las columnas disponibles de la ventana.
Llamada 3-16: Exporta el contenido de la ventana a un archivo de Excel.
Llamada 3-17: Muestra el submenú,
Llamada 3-17-1: Elimina el recurso elegido de todas las partidas propias del presupuesto.
Llamada 3-17-2: Este utilitario reemplaza el recurso elegido por otro utilizado en el presupuesto.
Llamada 3-17-2: Este utilitario reemplaza el recurso elegido por otro del catálogo.
Llamada 3-18: Muestra datos relacionados con el recurso.
Para copiar los precios de un presupuesto a otro
Para pegar los precios en el presupuesto destino,
Imprimir todos los recursos del presupuesto
Imprimir todas las subpartidas del presupuesto
5.1.6 Diseño de cabeceras para reportes
El S10 para la impresión de los reportes usa el diseño estándar, que viene por defecto. Esta opción permite al usuario registrar cabeceras para reportes, hasta en 10 líneas con el tipo, color y tamaño de letra.
El diseño es para un presupuesto elegido en el árbol. Permite diseñar para los siguientes reportes: Resumen del presupuesto, Presupuesto, Análisis de precios unitarios, Recursos y precios, Fórmula polinómica, Subpartidas.
Botones disponibles:
: Corta la selección al portapapeles.
: Copia la selección al portapapeles.
: Pega el contenido del portapapeles
Opciones del escenario,
Llamada 1: Elija el tipo de reporte.
Llamada 2: Elija el alineamiento.
Llamada 3: Registre el texto, al hacer clic derecho sobre registro muestra el menú.
Llamada 3-1: Elimina la fila elegida.
Llamada 3-2: Elija la fuente para el registro.
Llamada 3-3: Copia el formato del registro, fuente, tamaño de letra, color.
Llamada 3-4: Pega el formato, fuente, tamaño de letra, color.
Llamada 3-5: Envía el contenido de la ventana a un archivo de Excel.
Llamada 4: Este segmento es informativo.
Configuración para impresión con diseño de cabecera
Use el botón
y de la ventana que se active elija la pestaña de Impresión
Llamada a: Seleccionado, imprime con el diseño de cabecera elaborado por el usuario.
No seleccionado, imprime con el diseño de cabecera que viene por defecto.
5.1.7 Tiempos para programación
En este escenario se calcula la duración o plazo de ejecución de las partidas de acuerdo al metrado y rendimiento.
Contiene dos pestañas, para mano de obra y otro para equipos.
Proporciona una idea de la cantidad de grupos de trabajo o “cuadrillas” que se debe utilizar para ejecutar cada una de las partidas y el plazo.
Llamada 1: Elija la pestaña mano de obra.
Llamada 1-1: Esta columna es informativa, muestra el metrado de la partida.
Llamada 1-2: Esta columna es informativa, muestra el rendimiento de la partida.
Llamada 1-3: Esta columna es informativa, muestra el tiempo unitario de la partida, para una cuadrilla.
Llamada 1-4: Esta columna es editable, se registra la cantidad de cuadrillas a utilizar.
Llamada 1-5: Esta columna es informativa, muestra el tiempo unitario de la partida, para una cuadrilla registrada por el usuario, este tiempo es redondeado al inmediato superior.
Llamada 1-6: Al hacer clic en este segmento permite exportar el contenido de la ventana a un archivo Excel.
Esta información no es utilizada en otro escenario.
5.1.8 Presupuesto descompuesto
Un presupuesto contiene partidas y estos a su vez los análisis de precios unitarios que muchas veces tienen subpartidas en diferentes niveles (anidados uno debajo de otro). Este escenario es informativo y muestra el contenido del presupuesto las partidas y subpartidas juntas, para acumular los recursos de acuerdo a la selección.
Use este escenario también para imprimir los recursos del presupuesto para que los montos coincidan con el monto presupuestado, se implementaron opciones de redondeo. En necesario ubicar el cursor en la llamada 1, para luego hacer clic derecho en la llamada 1-1 y elegir vista preliminar.
Botones disponibles:
: Genera las subpartidas.
Opciones del escenario:
Llamada 1: Es un totalizador.
Llamada 1-1: La ventana inferior muestra los recursos que participan de acuerdo a la ubicación del cursor de la ventana superior, al hacer clic derecho muestra el menú.
Llamada 1-1-1: Permite hacer visible o esconder los campos o columnas disponibles en la ventana.
Llamada 1-1-2: Exporta el contenido de la ventana a un archivo Excel.
Llamada 1-1-3: Imprime en pantalla el contenido de la ventana.
Llamada 1-1-4: Envía el contenido de la ventana a la impresora previamente elegida en Windows.
Llamada 2: Al hacer clic derecho en este segmento, muestra el menú,
Llamada 2-1: Contrae el contenido de la ventana al nivel elegido.
Llamada 2-2: Exporta el contenido de la ventana de acuerdo al submenú,
Llamada 2-2-1: Exporta el contenido al MS Project.
Llamada 2-2-2: Exporta el contenido a un archivo Excel.
Llamada 2-2-3: Opción no disponible en esta versión.
5.1.9 Detalle de recursos
Permite efectuar análisis de precios unitarios al recurso, para averiguar el costo y trasladarlo al presupuesto. Estos datos los utilizaremos en el módulo de Gerencia de proyectos en el grupo de Operación de Equipos.
Los análisis de precios unitarios del detalle del recurso no se visualizan desde el presupuesto.
Ejemplos de detalles,
Seleccione los recursos que tendrán detalle,
Para registrar recursos en el detalle,
Para trasladar el precio al presupuesto,
Repita el proceso para cada uno delo equipos y luego vaya a la hoja del presupuesto y procese.
Opciones del escenario,
Llamada 1: Este filtro permite elegir el tipo de recurso que será mostrado en la ventana marcada con la llamada 6.
Llamada 2: Refresca las cantidades y por consiguiente los parciales de los recursos mostrados en la ventana superior.
Llamada 3: Acumulado costo
Llamada 4: Modo, está disponible las vistas de unitario y total
Llamada 5: Seleccionado, activa las opciones de edición del análisis de precios unitarios.
No seleccionado, desactiva las opciones de edición del análisis de precios unitarios.
Llamada 6: Al hacer clic derecho sobre un recurso muestra el menú,
Llamada 6-a: Seleccionado, activa las opciones de edición del análisis de precios unitarios.
No seleccionado, desactiva las opciones de edición del análisis de precios unitarios.
Llamada 6-b: Desactiva las opciones de edición del análisis de precios unitarios.
Llamada 6-c: Copia al detalle del recurso en el catálogo.
Llamada 6-d: Importa el contenido del detalle desde una hoja Excel, es requisito que los códigos de los recursos se encuentren registrados en la base de datos en uso.
Llamada 6-e: Exporta el contenido de la ventana a un archivo de Excel.
Llamada 6-1: Son los recursos del detalle, al hacer clic derecho en esta ventana, muestra el menú,
Llamada 6-1-1: Muestra el catálogo de recursos para adicionar al detalle.
Llamada 6-1-2: Adiciona el contenido de la plantilla al detalle.
Llamada 6-1-3: Adiciona recursos que se están utilizando dentro de algún detalle del presupuesto.
Llamada 6-1-4: Copia la selección al portapapeles.
Llamada 6-1-5: Corta la selección al portapapeles.
Llamada 6-1-6: Pega el contenido del portapapeles al detalle.
Llamada 6-1-7: Elimina el recurso seleccionado.
Llamada 6-1-8: Traslada el precio del detalle al recurso del presupuesto.
Llamada 6-1-9: Muestra el submenú, es un utilitario que permite cambiar un recurso por otro de acuerdo a las opciones marcadas con las llamadas 6-1-9-1-1 y 6-1-9-1-2.
Llamada 6-1-9-2: Elimina el recurso elegido en todos los detalles del presupuesto.
Llamada 6-1-10: Importa el contenido de detalle desde una hoja Excel, los códigos de los recursos a importar deben estar registrados en la base de datos en uso.
Llamada 6-1-11: Exporta el contenido de la ventana a un archivo de Excel.
Llamada 6-1-12: Muestra las propiedades del recurso.
Llamada 6-1-13: Permite hacer visible o esconder las columnas disponibles en la ventana.
5.1.10 Índices unificados
Permite el registro de los índices unificados que serán utilizados para actualizar los precios de los recursos que participan en los presupuestos.
Los índices unificados son números abstractos que permiten relacionar las variaciones de precios que experimentan los elementos componentes desde la fecha del presupuesto base hasta una fecha posterior, los mismos que son publicados mensualmente.
En el Perú, todo el universo de recursos que se utiliza en la construcción está agrupado dentro de 80 índices unificados y va cambiando el número de IU.
Para iniciar el registro de los índices unificados ingrese a configuración usando el botón
Opciones del escenario,
Llamada 1: Este segmento es informativo, muestra las 6 áreas geográficas en las que se dividió el Perú.
Llamada 2: En este segmento se registran los datos relacionados con la resolución que publica los índices unificados, al hacer clic derecho muestra el menú.
Llamada 2-2: Elimina el registro seleccionado y todo el contenido mostrado en la ventana marcada con la llamada 3.
Llamada 2-3: Muestra la ventana configuración de áreas, donde todos los campos son editables, la ventana con la llamada 1 muestra el registro.
Llamada 3: Este segmento es editable, donde se registran los valores de los índices unificados, al hacer clic derecho muestra el menú.
Llamada 3-1: Importa el contenido de la grilla desde una hoja Excel.
Llamada 3-2: Envía el contenido de la grilla a un archivo Excel.
5.2 Grupo Precios
5.2.1 Precios por presupuesto
Muestra los precios que utiliza el presupuesto.
Llamada 1: Elija el botón de acceso al escenario.
Llamada 2: Sí el cursor se ubica sobre el nombre del presupuesto, muestra todos los recursos que participan en los subpresupuestos y las cantidades son acumuladas.
Sí el cursor se ubica sobre un subpresupuesto, solo muestra los recursos que le pertenecen.
Llamada 3: Contiene los filtros disponibles,
Llamada 3-1: Carga en la grilla recursos o subpartidas.
Llamada 3-2: Muestra las fechas y los lugares que contienen los recursos con precios.
Llamada 3-3: Para el índice unificado elegido, muestra en la grilla los recursos que le pertenecen.
Llamada 3-4: Agrupa en la grilla los recursos por factor.
Llamada 4: Lista los recursos del presupuesto, al hacer clic derecho muestra el menú,
Llamada 4-1: Muestra los precios utilizados en otros presupuestos y los precios de acuerdo a los ingresos por compra y órdenes de compra a almacenes.
Llamada 4-2: Este utilitario en informativo, del recurso elegido muestra las partidas donde participa el recurso de acuerdo a las llamadas 4-2-1 y 4-2-2.
5.2.2 Precios por grupos
Este escenario es prácticamente un archivo donde se almacena una relación de recursos que más usa la empresa en sus presupuestos, con la finalidad de registrar los precios para un lugar y fecha determinado. Puede registrar precios para enero, febrero, marzo, abril, etc.
Los precios de estos grupos pueden ser copiados al portapapeles y luego pegados a un presupuesto.
Puede registrar los siguientes grupos:
• Electromecánicas
• Edificaciones
• Puentes
• Carreteras
Llamada 1: Botón de ingreso al escenario.
Llamada 2: Muestra los filtros disponibles en el escenario.
Llamada 2-1: Permite elegir un lugar.
Llamada 2-2: Permite elegir una fecha.
Llamada 2-3: Este filtro que muestra los lugares y fechas que contiene precios, la ventana marcada con la llamada 4-1 lista los recursos y sus precios.
Llamada 3: En este segmento se registran los grupos, previa configuración (Definir estructura).
Llamada 3-1: Para el grupo seleccionado, en el segmento marcado con la llamada 4-1 se registran los recursos y los precios, guarda relación el lugar y la fecha.
Llamada 4: En este segmento se registran los recursos y sus precios de acuerdo a la fecha y el lugar.
Llamada 4-1: Al hacer clic derecho sobre un recurso, muestra el menú,
Llamada 4-1-1: Muestra el catálogo de recursos para elegir y registrarlos en el grupo.
Llamada 4-1-2: Copia la selección al portapapeles.
Llamada 4-1-3: Corta la selección al portapapeles.
Llamada 4-1-4: Pega el contenido del portapapeles en otro grupo.
Llamada 4-1-5: Elimina la selección.
5.2.3 Precios generales
Este escenario es prácticamente un archivo donde se almacena los recursos que utilizan los presupuestos la opción de ALMACENAR EN PRECIOS GENERALES, en la ventana recursos y precios del escenario de la hoja del presupuesto, carga los datos a este escenario.
Cuando elabore un nuevo presupuesto para ese lugar y fecha, tomará los precios almacenados en este escenario, caso contrario pedirá que los ingrese.
Llamada 1: Es el botón de acceso al escenario.
Llamada 2: Contiene los filtros disponibles,
Llamada 2-1: Muestra las fechas y los lugares que contienen los recursos con precios.
Llamada 2-2: Filtra el tipo de recurso a ser mostrado en la grilla.
Llamada 2-3: Agrupa en la grilla los recursos de acuerdo al índice unificado.
Llamada 2-4: Filtra los recursos de acuerdo al precio.
Llamada 3: La grilla muestra los recursos de acuerdo a la selección en los filtros.
Llamada 3-1: Muestra los precios del recurso en la moneda principal utilizada en el presupuesto.
Llamada 3-2: Muestra los precios del recurso en la moneda secundaria (sólo para los presupuestos que se elaboró en doble moneda).
5.2.4 Precios por presupuesto histórico
Este escenario es informativo, muestra en un archivo plano todos los presupuestos y sus recursos elaborado en la base de datos en uso, con la finalidad de explotar la información.
Llamada 1: Botón de acceso al escenario.
Llamada 2: Al presionar este botón muestra los campos disponibles en el escenario con la finalidad de hacer visible, esconder, cambiar de ubicación y agrupar (ver llamada 2-1), de acuerdo a esta selección muestra los datos en la grilla.
Llamada 3: Es el registro agrupado, de acuerdo a la llamada 2-1.
Llamada 3-1: Es el registro agrupado en el segundo nivel, en este caso el presupuesto.
5.3 Grupo Transportabilidad
5.3.1 Exportar presupuestos
Este es un utilitario que permite exportar el presupuesto seleccionado a otra base de datos con la finalidad de independizarlo.
Todas las partidas del presupuesto a exportar deben ser propias.
Para exportar un presupuesto
Del catálogo de presupuesto elija los presupuestos a exportar.
Elija los datos a exportar
Registre el nombre a la nueva base de datos (no use el nombre de un base de datos existente ya que la reemplaza o sobre escribe). Luego use el botón de SIGUIENTE.
Luego use el botón de EXPORTAR y luego ACEPTAR.
El usuario accede a la nueva base de datos desde la barra de menús con la opción Archivo que despliega el menú donde
debe elegir INICIAR LA SESION COMO USUARIO DISTINTO. El mismo efecto tiene cuando se presiona el botón de la barra de estado. Cuando se inicia la sesión con el sistema aparece la ventana:
5.3.2 Proceso para trasladar la base de datos a otra PC
Una vez que ingresó a la base de datos exportada que tiene el nombre de PUENTE,
Saque copia de seguridad de la base de datos, ingrese al escenario de MANTENIMIENTO DE BASE DE DATOS, de la ventana que muestre elija el botón SIGUIENTE,
Ingrese por el explorador a la carpeta S102000\backup y copie en un dispositivo la base de datos.
En la PC destino en la carpeta S102000\backup, pegue el contenido del dispositivo, y utilice la opción de restaurar base de datos.
Cuando saque copia de seguridad compacta, verifique por el explorador que el archivo b2k.exe exista en la carpeta S102000\backupS10, caso contrario comuníquese con las oficinas del S10.
LIMITACIONES:
No use esta opción en versiones no autorizadas.
Sí en su PC tiene SQL de una versión actual, no traslada a versiones anteriores.
5.3.3 Importar presupuesto
Este utilitario está disponible, sí previamente sacó copia de seguridad de la base de datos o exportó el presupuesto a nueva base de datos. Sí registra nuevas partidas, recursos, socio de negocios, es indispensable que en la base de datos origen también los registre, sólo en ese caso se puede importar el presupuesto, resumiendo el concepto, en la base de datos origen y destino deben existir los mismos registros con los mismos códigos, solo así se evitará conflictos.
5.3.4 Importar Datos Windows 1.x, 2.x
Este utilitario permite importar la base de datos de la versión 1 y 2 del S10 (en Access), es requisito que la base de datos a importar esté registrada en una carpeta del servidor. Al importar se crea una nueva base de datos en SQLSERVER.
Para visualizar la base de datos importada, haga doble clic en la cara que se muestra en la barra de estado, en la ventana de acceso al S10, elija la nueva base de datos.
5.3.5 Importar datos DOS 7.x 8.x
Este utilitario permite trasladar la base de datos del módulo de presupuestos de las versiones en DOS que usa DBF a la base de datos SQLSERVER. Es conveniente que previamente el usuario ingrese por el explorador y verifique en la carpeta S10200\backup exista el archivo modeldos, caso contrario comuníquese con las oficinas del S10. Este utilitario permite importar la base de datos en DOS del S10, es requisito que la base de datos a importar esté registrada en una carpeta del servidor. Al importar se crea una nueva base de datos.
Para visualizar la base de datos importada, haga doble clic en la cara que se muestra en la barra de estado, en la ventana de acceso al S10, elija la nueva base de datos.
5.4 Grupo Informes
Los escenarios de este grupo son informativos, opciones disponibles de acuerdo a la configuración adquirida de parte del cliente al S10.
5.4.1 Detallado de partidas
Este escenario es informativo, muestra el contenido de la hoja del presupuesto con la finalidad de explotar la información.
En el árbol, ubique el cursor sobre el nombre del presupuesto o subpresupuesto para visualizar su contenido.
Opciones del escenario,
Llamada 1: Use este botón para hacer visible o esconder los campos o columnas disponibles.
Llamada 1-1: Arrastre a este segmento las cabeceras de las columnas para agrupar la información de la ventana marcada con la llamada 1-2.
5.4.2 Presupuesto detallado
Este escenario es informativo, muestra el contenido de la hoja del presupuesto con las partidas, subpartidas y recursos con la finalidad de explotar la información.
En el árbol, ubique el cursor sobre el nombre del presupuesto para visualizar su contenido.
Opciones del escenario,
Llamada 1: Use este botón para hacer visible o esconder los campos o columnas disponibles.
Llamada 1-1: Arrastre a este segmento las cabeceras de las columnas para agrupar la información de la ventana marcada con la llamada 1-2.
5.4.3 Detallado de partidas – recursos
Este escenario es informativo, muestra el contenido de la hoja del presupuesto con las partidas, subpartidas y recursos, precios unitarios y parciales, con la finalidad de explotar la información y usar el cubo.
En el árbol, ubique el cursor sobre el nombre del presupuesto para visualizar su contenido.
Opciones del escenario,
Llamada 1: Use este botón para hacer visible o esconder los campos o columnas disponibles.
Llamada 1-1: Arrastre a este segmento las cabeceras de las columnas para agrupar la información de la ventana marcada con la llamada 1-2.
Llamada 2: El cubo permite visualizar la información de las columnas de acuerdo al agrupamiento que defina el usuario.
5.5 Grupo Utilitarios
Contiene las opciones de mantenimiento de base de datos y controles de accesos a los escenarios del módulo. El sa es el encargado de los registros.
5.5.1 Registro de usuarios
La finalidad es registrar en este módulo a la persona que lo manejará, se le asigna accesos y atributos.
La elección del personal se efectúa del catálogo de socio de negocio.
Datos que serán registrados
Acceso al escenario
Registre el nombre de la empresa o constructora,
Registro de los Subgrupos: Registre a Personal Técnico y Personal Administrativo
Registro de Usuarios
Llamada a: Uno de los usuarios puede ser elegido como auditor de la base de datos, el botón muestra una lista de visto bueno (VB) de diferentes colores. Elija un color para el auditor, el sistema permite hasta 6 auditores por base de datos
Llamada b: Ponga una selección al usuario encargado de informática.
Acceso al módulo: Sólo se da acceso desde el módulo que el usuario se encuentra trabajando.
Las categorías de usuario: Son etiquetas que se asignan a los usuarios. La categoría se asigna desde el módulo que usara por el usuario. El “sa” asigna a cada uno de los usuarios que tienen categoría, accesos a los escenarios, catálogos.
Administrador de módulo: Tiene acceso total a los módulos del S10, es el encargado de registrar a los demás usuarios y asignarles atributos de manejo del sistema de acuerdo a su jerarquía y al funcionamiento de la organización.
Jefe de grupo: Es el encargado de administrar a un grupo de usuarios que usan el módulo, tiene acceso a todos los escenarios y a la información del grupo al cual pertenece. Asigna atributos de uso que viene por defecto.
Especialista: No debe registrar usuarios, no le permite asignar atributos, tiene acceso a los escenarios que se le encargue.
Usuario: Asigne los escenarios donde puede realizar las sus labores, además permisos y atributos de uso.
Ninguno: Asigne esta categoría para borrar los atributos antes asignados
Acceso a los escenarios
Acceso a los catálogos
Sólo la persona que registre el presupuesto tiene acceso a ese presupuesto, pero puede invitar a otros usuarios para que compartan en la elaboración del presupuesto.
Llamada a: Permite efectuar nuevos registros.
Llamada b: Permite modificar los registros existentes.
Llamada c: Permite eliminar un registro siempre y cuando no esté participando en una aplicación.
Llamada d: Permite registrar notas a los registros existentes.
Llamada e: Permite imprimir los datos de los catálogos.
Llamada f: Permite seleccionar los registros elegidos.
5.5.2 Actualización de base de datos
Es un utilitario que adecúa la base de datos a la nueva estructura, es una labor que debe realizar el “sa” cuando el sistema envíe un aviso.
5.5.3 Mantenimiento de base de datos
Es un utilitario, destinado a dar mantenimiento a las bases de datos, son labores que debe realizar el “sa”.
Llamada a: Es un utilitario que permite sacar copia de seguridad de la base de datos, lo almacena en C:\s10200\backup con el nombre que el usuario registre.
Llamada b: Es un utilitario que permite restaurar una copia de seguridad que se encuentra almacenada en C:\s10200\backup, el usuario debe registrar un nombre diferente a las bases de datos existentes.
Llamada c: Es un utilitario que permite crear una base de datos vacía, el usuario debe registrar un nombre diferente a los existentes.
Llamada d: Es un utilitario que permite eliminar una base de datos existente. La limitación es que no se puede eliminar la base de datos en uso.
Llamada e: La base de datos está compuesta por 3 archivos , , , que se ubican en C:\s10200\data. Es un utilitario recupera la base de datos a partir esos archivos.
Llamada f: Es un utilitario que permite compactar u ordenar la base de datos para que ocupe menos espacio.
Llamada g: Es un utilitario que permite reindexar u ordenar la base de datos, con la finalidad de dar mayor velocidad a los procesos.
Llamada h: Es un utilitario que permite crear tareas para administrar la base de datos en forma automática.
Llamada i: Copia de seguridad y restauración diferencial de base de datos – en implementación.
Llamada j: Ver registro de transacciones, permitirá llevar el historial de cambios de la base de datos, almacenando los valores anteriores antes del cambio, la fecha y hora y el usuario que lo ha cambiado, esta opción es con fines de auditoria.
Llamada k: Historial de actualización de base de datos, almacena los scripts con todos sus datos y si ha sido exitosa la actualización, ayudara a detectar fácilmente que script no han corrido en una base de datos.
Llamada L: Scripts Puntuales, se utiliza para ejecutar y tener el histórico de script que se ha preparado específicamente para un cliente.
6 PLANEAMIENTO
La planificación permite un manejo ordenado de nuestros proyectos y una toma de decisiones acertada en base a información.
Es una visión futura de los que haremos, nos permite predecir los resultados que deseamos lograr, fijando metas a corto y largo plazo. La planificación crea frentes de trabajo.
Una buena planificación asegura que cada actividad cuente con los recursos en forma oportuna, para ser ejecutada.
6.1 ¿Qué esperamos de la planificación?
• Contar con un cronograma gráfico, que nos proporciona la duración bastante aproximada del proyecto y la secuencia constructiva; desde el MS Project.
• Cronograma de avances o de metrados, los que nos comprometemos a ejecutar en cada periodo.
• Cronograma valorizado o de desembolsos, es lo que nos comprometemos a invertir en cada periodo y también lo que debemos lograr valorizar.
• Plan de utilización de recursos, viene a ser la relación de recursos presupuestados distribuidos en el tiempo, permite conocer:
La cantidad y tipo de mano de obra a utilizar en cada periodo
La cantidad y tipo de equipo a utilizar en cada periodo
La cantidad de materiales a utilizar en cada periodo Los subcontratos que emplearemos en el proyecto.
6.2 Que evitamos con la falta de planificación
La improvisación que solo trae desorden y caos. (La del bombero).
S10: Los cambios continuos producen mayor gasto.
6.3 Planificación desde los módulos de Presupuestos y Gerencia de proyectos S10ERP.
El módulo de Gerencia de proyectos del S10ERP utiliza una rama para el presupuesto venta y otra rama para la meta.
El módulo de Presupuestos del S10ERP, sólo dispone de la rama venta (no visible), pero todo lo trabajado es mostrado en la rama venta del módulo de Gerencia de proyectos del S10ERP.
Ambos módulos comparten los escenarios de Datos generales (para Presupuestos no muestra las ramas venta y meta) + Calendario + Actividades + Periodos + Cronograma por Periodos.
Para usar las opciones disponibles en planeamiento se asignará al proyecto un presupuesto meta que tiene subpartidas.
6.4 Registro del Proyecto
Datos del proyecto: Utilice estos datos para registrar el proyecto:
* Este dato varía de acuerdo a la fecha en la que se dictan las clases.
Acceso al escenario de Datos Generales
El proyecto y sus datos complementarios se registran en el escenario de Datos Generales.
Registro del grupo en el Catálogo de Proyectos
Llamada a: En el árbol permite el registro de 99 ítems, para cada uno 999 proyectos. Procure registrar en el árbol los años o las líneas de negocio a las que se dedica su empresa.
Registre: Año 2010, (si ya tiene registrado información es este nivel – omita este paso).
Dentro del grupo Año 2010, registre los datos del proyecto;
Ingreso de datos del proyecto
Llamadas de la ventana “Proyecto”
Llamada 1: Se recomienda usar antes del nombre del proyecto el número de contrato, orden de servicio o número de obra que se ejecuta en el año.
Llamada a: Registre el nombre del propietario del proyecto.
Llamada b: Elija el lugar (distrito), donde se ejecutará el proyecto.
Llamada c: Ingrese la fecha de inicio del proyecto, este dato es importante, ya que a partir de este el sistema genera períodos semanales y mensuales, IMPORTANTE: SI YA TIENE INFORMACIÓN REGISTRADA EN CUALQUIERA DE LOS MÓDULOS – NO – MODIFIQUE ESTA FECHA. La fecha de inicio real del proyecto se modifica en el MS Project.
Llamada d: No ingrese este dato, el sistema lo calcula el sistema automáticamente.
Llamada e: Es el lapso de tiempo en días calendarios que está previsto dure la ejecución del proyecto, es un dato tentativo inicial y sirve para generar los períodos a partir de la fecha de inicio.
Llamada f: Hay muchas empresas que inician la semana para efectos del pago de planillas el día lunes, otros el miércoles. Cuando el sistema genere los períodos semanales, la semana iniciará el día que eligió. Uniformice, se recomienda que empiece el día lunes en todos los proyectos.
Llamada g: No disponible en este módulo.
Llamada h: Elija el mismo tipo de moneda que utilizó en el presupuesto.
Llamada i: Elija el tipo de moneda alterno que utilizará el módulo de almacenes.
Llamada j: Ingrese el nombre de empresa que ejecutará el proyecto.
Llamada k: Sí la empresa tiene sucursales elija la que corresponde. En el catálogo de socio de negocios se registra la empresa y las sucursales.
Llamada L: Ingrese la dirección del proyecto, es utilizado por el módulo de Compras.
Llamada m: Ingrese el número telefónico del local donde funcionan las oficinas del proyecto.
Llamada n: Ingrese el número de fax del local donde funcionan las oficinas del proyecto.
Llamada o: Después de registrar todos los datos use este botón para grabarlos y cerrar la ventana y el proyecto será mostrado en el catálogo de proyectos.
6.4.1 Asignar presupuestos al Proyecto
El presupuesto proporciona al proyecto, las partidas, metrados, recursos, cantidades y precios.
Llamada a: Del catálogo de presupuestos asigne el presupuesto de compromiso, También los presupuestos de adicionales que son ordenados por el cliente.
6.4.2 Definir la escala de periodos para ingreso de avances
La escala elegida, es utilizada en el escenario por Cronograma por Periodos, para ingresar los avances en el modo programado y valorización.
6.5 Calendario
El proyecto usa calendarios para definir los días hábiles y feriados dentro del lapso de ejecución del proyecto. Por defecto, al registrar un proyecto en el escenario de Datos generales el sistema le asigna un calendario base que es utilizado por todas las actividades del proyecto.
Acceso al escenario
6.5.1 Registro de días no laborables
Registre los días no laborables en el lapso de ejecución del proyecto, del lugar donde se ejecutará.
Llamada a: Sí en el catálogo de Feriados ya tiene registrado, al presionar en este botón toma los feriados registrados a nivel nacional.
Luego marque los feriados adicionales del lugar donde se efectúa el proyecto.
6.5.2 Horario de trabajo
En el módulo de presupuestos, no interviene el horario.
6.5.3 Calcular horas y días laborables para los períodos
Efectúe esta opción sólo en el caso de que muestre el botón . El sistema está preparado para efectuar el cálculo en línea.
6.6 Actividades
En el escenario de Actividades, se prepara la información para la planificación física del proyecto, se registra el WBS y se le asignan las partidas del presupuesto.
PMI: El conjunto del proyecto debe ser descompuesto en segmentos más pequeños y manejables con suficiente detalle, organizando el proyecto de acuerdo al proceso constructivo.
S10: Hace participar a todas las partidas de los presupuestos asignados.
A cada una de las fases se le asigna una partida del presupuesto, con los metrados que corresponde.
El conjunto de las fases con sus respectivas actividades pasa al escenario del Cronograma por periodos donde se programa el proyecto.
Actividad es una porción de partida asignada a una fase o WBS.
Cuando se cambia el presupuesto, vaya a S10 Presupuestos y procese (use el procesamiento tipo 3) el presupuesto. Luego use el botón en el escenario de Actividades.
Acceso al escenario de Actividades
Al ubicar el cursor sobre el proyecto en el árbol, toma los datos del presupuesto asignado, mostrando las partidas y subpartidas en una sola hoja con la finalidad de que puedan ser utilizadas en la planificación. Esta labor se realiza sólo la primera vez que se ingresa a este escenario o cuando se asigne más presupuestos al proyecto.
En caso de que el proyecto tenga asignado varios presupuestos, para visualizar las partidas de otros presupuestos asignados.
6.6.1 Asignación automática
Asignación automática, el sistema registra automáticamente las fases o WBS, y también asigna las partidas del presupuesto a cada una de las fases, conservando la estructura del presupuesto, use este tipo de asignación para el proyecto que se le asigne el presupuesto del cliente.
UTILICE ESTA OPCIÓN SOLO CUANDO EL PROYECTO LO USE PARA VALORIZAR.
Use el botón
Llamada a: El sistema registra automáticamente las fases y le asigna las partidas que le corresponde mostrado en la llamada b.
Para asignar un segundo presupuesto, use el botón
para mostrar las partidas del presupuesto, luego el botón
(asignar solo el presupuesto actual).
Luego, vaya al escenario de PERIODOS, y haga visibles los meses después en el escenario del CRONOGRAMA POR PERIODOS, programe y valorice. Para imprimir la valorización elija modo valorización costo, al hacer clic derecho sobre una actividad elija vista preliminar, luego valorización.
6.6.2 Ejemplos de WBS para el registro manual
No te preocupes, tu malestar comprendo que fue natural, siempre con transparencia. Asi no nos creamos mascaras
Llamada a: Es un WBS detallado, favorece a la programación e ingreso de avances.
Llamada b: Es un WBS abreviado, en cada fase se agrupa lo que se tiene que ejecutar.
Llamada a: Sí en el subpresupuesto, de instalaciones eléctricas y sanitarias tiene las partidas por pisos asigne a cada piso lo que le corresponde.
Llamada b: Sí en el subpresupuesto, de instalaciones eléctricas y sanitarias tiene las partidas globales registre una fase.
La programación gráfica en la estructura se realiza por pisos y los acabados por actividades.
Llamada a: Son las obras iniciales.
Llamada b: Son las fases de un tramo.
Llamada c: Registre en cada tramo las fases que le corresponde.
Llamada d: Son las obras generales que tienen duración prolongada y se usan en cada tramo.
Llamada e. Son los gastos generales.
Llamada f: Registre el incremento de obra, a medida que se produzca.
6.6.3 Registro manual del WBS o fases y asignación de partidas
Dependiendo del tipo de proyecto se realiza un análisis de la forma como será construida, en el caso del puente que se muestra en este croquis, se especifican todas las partes del proyecto a construir.
Recomendado para el presupuesto de gestión.
Se prepara la agenda del proyecto en forma realista.
No descomponga la EDT en más partes que la que pueda manejar
Luego en un organigrama se describe las partes del puente en forma detallada, tal como será construida.
Registro de las Fases (WBS)
Definir estructura de registro, debido a que es un catálogo abierto.
Llamada a: En la ventana, estructura de fases permite modificar los niveles y longitud, no debe exceder de 15 dígitos. Luego cierre la ventana.
Ingreso del nombre de la Fases (WBS): Registre los datos de acuerdo a lo mostrado en el organigrama.
Este botón indica que el registro será al mismo nivel.
Este botón indica que el registro será a un nivel inferior.
Estas son las fases registradas:
Llamada a: Haga clic derecho en este segmento para imprimir el reporte.
6.6.4 Preparando las partidas antes de asignar a las Fases
6.6.4.1 Consolidar Partidas
Este utilitario permite independizar subpartidas del presupuesto principal, con la finalidad de programarlos separadamente. Durante el proceso de consolidación los metrados son acumulados.
El presupuesto que se toma como ejemplo para este proyecto, tiene subpartidas para la preparación de agregados en todas las partidas de concreto; las que serán consolidadas.
Esto permitirá programar la preparación de agregados primero, luego la construcción.
Consolide:
Procedimiento:
Repita el proceso para las demás subpartidas. Estas son las subpartidas consolidadas
6.6.4.2 Hacer Asignable
Este utilitario acumula el contenido de las subpartidas a la partida principal, con la finalidad de programar solo la partida principal.
Hacer asignable las siguientes partidas:
Procedimiento:
Repita el proceso para todas las actividades.
A continuación, se muestra las actividades consolidadas:
6.6.5 Asignar partidas a las Fases (WBS)
A cada fase se le asigna las partidas que les corresponden.
Llamada a: Elija la fase.
Llamada b: Asigna la actividad a la fase.
Llamada c: Este segmento muestra las actividades asignadas a la fase.
Llamada c-1: Ingrese el porcentaje del metrado de la actividad que le corresponde a la fase.
Ubicar actividades con saldo: Para verificar que todas las partidas hayan sido asignadas a las fases use el botón y el sistema muestra el cuadro de diálogo,
Use el botón , para ver el reporte en pantalla, luego del menú elija: Fases y actividades.
Para visualizar las actividades asignadas:
Llamada a: Al hacer clic derecho muestra el menú,
Llamada a-1: Permite eliminar la actividad asignada.
Llamada a-2: Envía a un archivo de Excel el contenido de la ventana.
6.7 Períodos
Los periodos son lapsos de tiempo en los cuales se planifica la ejecución de las actividades y el sistema los genera en forma automática de acuerdo a la fecha de inicio del proyecto y el plazo, estos datos son tomados del escenario de Datos Generales. Los períodos definidos son para todo el proyecto.
Labores que se efectúan en este escenario
• Hacer visibles o esconder periodos, para los escenarios donde sean requeridos.
Acceso al escenario de Periodos
Hacer visible los períodos
Por este escenario se hacen visibles o esconder los períodos que se utilizarán en el escenario del Cronograma por Períodos.
Elija solo los periodos mensuales y semanales, como se muestra.
Llamada a: Es el totalizador de los periodos mensuales.
Llamada b: Es el totalizador anual de los periodos mensuales.
Llamada c: Es el totalizador de los periodos semanales.
Llamada d: Es el totalizador anual de los periodos semanales.
Elija solo los periodos mensuales y semanales, como se muestra.
6.8 Planificación Física del proyecto
La planificación permite conocer las metas a largo y mediano plazo así crear frentes de trabajo (qué es lo que debemos avanzar o hacer en cada periodo).
• En el escenario de Actividades se preparó la información que será utilizada en la planificación física del proyecto.
Un manejo adecuado de la planificación nos permite cumplir con el tiempo previsto y calidad.
Para la planificación física del proyecto tenemos dos alternativas:
6.8.1 Planificación manual en el escenario del Cronograma por Períodos
Permite el ingreso de porcentajes o metrados (no costo), en los períodos a programar, sean estos mensuales o semanales, dependiendo de la configuración elegida en el escenario de datos generales.
Acceso al escenario del Cronograma por Periodos
Modo de visualización de datos: Permite elegir el tipo de dato que será visualizado en el escenario. En la barra de botones elija
:
Para ingresar información, sólo tiene que elegir modo porcentaje o metrado (no costo)
Elija la escala de período,
Ingreso manual de datos en el cronograma programado
Llamada a: Cuando se registra el porcentaje en la primera celda a la altura del nombre del proyecto (totalizador), replica el porcentaje en toda la columna.
Llamada b: Cuando se registra el porcentaje en una celda a la altura de un título, replica el porcentaje en las actividades subordinadas a este título.
Llamada c: Cuando se registra el porcentaje en una celda que corresponde a una actividad, el dato registrado solo afecta a la actividad.
Llamada d: Marque un bloque horizontal con varias celdas, luego registre el porcentaje a repartir en el bloque.
Llamada e: Marque un bloque vertical con varias celdas, luego se registre el porcentaje a repetir.
Para visualizar el porcentaje o cantidad de metrado programado; haga clic en la cabecera del periodo semanal para marcar un bloque.
Llamada f: Luego de hacer clic en la cabecera del periodo para marcar un bloque, se utiliza SALDO PROYECTADO, para tomar el saldo del porcentaje o metrado que falta programar.
Calcular recursos:
Para el período elegido y para los metrados o porcentajes ingresados el sistema calcula la cantidad de recursos de acuerdo al análisis de precios unitarios.
Luego elija en la ventana Recalcular, use el botón para calcular los recursos de acuerdo al análisis de precios unitarios.
6.8.2 Planificación gráfica del proyecto usando el MS Project
Las fases o WBS con sus respectivas actividades asignadas en el escenario de Actividades, es utilizada en el escenario del Cronograma por Períodos para efectuar la programación gráfica, de donde se exporta del MS Project.
En el MS Project, se vinculan las barras, luego al actualizar hacia el S10 calcula los recursos.
PARA TOMAR EN CUENTA:
Sí trabaja en LAN, el cliente y servidor – deben tener la misma versión del MS Project.
Si se conecta mediante escritorio remoto – solo utiliza los recursos del servidor, donde tiene instalado el MS Project.
El MS Project no soporta doble moneda, tampoco jornadas diferentes en las partidas, se refiere al número de horas del día.
Configuración de Planeamiento (disponible para el usuario sa),
Llamada f-1:
Seleccionado – En el escenario del Cronograma por periodos, tanto para el modo programado como para la valorización controla no excederse del 100% al ingresar los avances.
No seleccionado – En el escenario del Cronograma por periodos, tanto para el modo programado como para la valorización permite excederse del 100% al ingresar los avances.
Llamada f-3:
Cuando el presupuesto asignado al proyecto no tiene análisis de precios unitarios (sólo está conformado por partidas formato), es necesario elegir un recurso para que tome el precio de los recursos, caso contrario en sistema enviará un mensaje en el escenario del Actividades.
Llamada f-4:
No seleccionado, por defecto el S10 utiliza el MS Project 2003, y este programa tiene que estar instalado en el servidor y el terminal.
Seleccionado, permite el uso del MS Project 2007 o superior, siempre y cuando tenga instalado en el cliente y servidor (USE EL MS PROJECT CON LICENCIA).
.
Acceso al escenario del Cronograma por Periodos
Modo de visualización de datos: En la barra de botones elija :
Elija la Escala de periodo
Duración de las actividades
El metrado/rendimiento, proporciona la duración de la actividad en días.
Para la duración de las actividades en días, es necesario hacer visible los campos de Fase, Descripción, Metrado presupuesto, Rendimiento, Cuadrilla, Duración días.
Haga visible los campos:
Verificando la duración de las actividades
La cantidad de personal o número de cuadrillas considerado para la construcción obedece a la experiencia del programador INICIALMENTE y la necesidad de culminar la actividad en un período determinado.
Considere las siguientes cuadrillas:
A las partidas de gastos generales no se les asigna cuadrillas.
El número de días es el tamaño de la barra en el MS Project.
6.8.2.1 Exportar información del S10 al MS Project
No es necesario exportar los recursos del S10 al MS Project.
IMPORTANTE: Es necesario que la PC del usuario, en panel de control, configuración regional, moneda, como símbolo no tenga (/).
6.8.2.2 Vincule las actividades
Tome en cuenta el orden constructivo: Esta es una programación gruesa.
En el MS Project, use el botón para guardar el archivo en la base de datos del S10.
IMPORTANTE, En el MS Project no use:
Guardar como: Porque le cambia la ruta y corre el riesgo de que el S10 no tome los datos.
No inserte ni quite tareas – sí desea hacerlo ingrese al escenario de Actividades.
6.8.2.3 Actualizar los datos del MS Project al S10
Esta opción, traslada la información del MS Project al S10.
Ahora los datos del MS Project pasaron automáticamente al S10. Luego veremos los reportes de la planificación.
6.9 Resultados de la planificación
Hay dos parámetro que se eligen, la escala de periodo, y el modo que puede ser metrado o costo.
Los reportes se emiten desde el escenario del Cronograma por periodos.
Acceso al escenario
El PMI dice, planifique dentro del monto presupuestado. Para visualizar el monto presupuesto asignado a las fases, y el programado;
Con el botón se elige el tipo de dato que será mostrado en la grilla del escenario del Cronograma por Periodos
También es necesaria la escala de período en el que tiene datos.
Impresión de reportes de la planificación
Elija el botón y del menú que muestre, luego, elija una opción.
6.9.1 Reporte Cronograma por partida
A pesar de tener los datos por fase, en este reporte el sistema muestra la información por partidas para cada uno de los subpresupuestos.
6.9.2 Reporte Valorizado contractual
Utilice esta opción cuando tenga avances registrados en modo valorización.
Para imprimir el reporte, use el botón y visualice los datos en modo valorización costo.
6.9.3 Reporte Cronograma
Puede ser valorizado o metrado, de acuerdo a los datos mostrados en el escenario.
6.9.4 Reporte del Plan de Utilización de Recursos
Es la relación de recursos utilizados en la ejecución del proyecto y cuyas cantidades han sido distribuidos en el tiempo de acuerdo a lo programado. Los datos pueden ser mostrados en cantidad o valorizado de acuerdo a la elección en modo visualización.
Llamada 1: En una hoja muestra los recursos de mano de obra, materiales, equipos y subcontratos.
Llamada 2: En una hoja muestra los recursos de mano de obra.
Llamada 3: En una hoja muestra los recursos tipo material.
Llamada 4: En una hoja muestra los recursos tipo equipo.
Llamada 5: En una hoja muestra los recursos tipo subcontrato.
Exporte el contenido de la ventana superior a un archivo de Excel.
Luego, elija la carpeta y registre el nombre del archivo a guardar.
Vista de Excel con la programación semanal,
Para visualizar las actividades con sus recursos para luego exportar a un archivo de Excel,
Luego, elija la carpeta y registre el nombre del archivo a guardar.
6.10 Valorización en el escenario del Cronograma por periodos
El porcentaje del plan completado, es el término utilizado en gerencia de proyectos y se refiere al registro del trabajo realizado durante la semana. En el S10 se llama valorización.
6.10.1 Registro de Avances
Los avances realmente logrados en obra dentro de un período, se registran en el sistema con la finalidad de valorizar; estos datos son comprados con el programado en el escenacio PROGRAMADO – VALORIZADO.
Estos son los datos que se ingresarán, en metrados.
Acceso al escenario del Cronograma por Periodos
Use el botón para eligir el tipo de dato que será mostrado en la grilla del escenario del Cronograma por Periodos
El ingreso de metrados es manual.
Para los Metrados Registrados, el sistema calcula la cantidad de recursos de acuerdo a los análisis de precios unitarios.
Luego en la ventana que muestre elija RECALCULAR.
Para visualizar los datos en costo
Impresión de la valorización, sólo está disponible para el modo VALORIZACION – COSTO
El reporte de la valorizacion, no tiene pie. En el módulo de Gerencia de proyecto, escenario de Contratos, se registra el resumen de de todas las valorizaciones de los subprsuestos y se diseña el pie.
6.10.2 Programado – Valorizado
Toma los datos del escenario del Cronograma por periodos, del programado y valorizado, con la finalidad de comparlos; a nivel de actividades (partidas del presupuesto), y por grupo de recursos y recursos.
Acceso al escenario,
Llamada a: Borra los campos seleccionados.
Visualice la comparación de lo programado y valorizado,
6.11 Ejemplos de planificación gráfica
6.11.1 Planificación gráfica de edificaciones
6.11.2 Planificación gráfica de saneamiento
6.11.3 Planificación gráfica de carreteras
7 BOTONES DE LOS ESCENARIOS DE PLANEAMIENTO
7.1 Botones del escenario del Actividades
: Expandir todos, muestra todas las partidas del presupuesto activo.
: Contraer todos, muestra el presupuesto sólo al primer nivel.
Ubicar una partida, cuando el puntero del ratón lo ubica sobre una partida, el sistema muestra una ventana donde indica al presupuesto al que pertenece, así como el ítem dentro de la hoja del presupuesto.
Todos, sí este botón está activo muestra todas las actividades asignadas a un grupo de fases, sólo es de consulta.
: Generar actividades, toma los datos del presupuesto. Utilice esta opción todas las veces que cambie datos en el presupuesto.
: Definir nivel de asignación, muestra una ventana para elegir el nivel de asignación de las partidas a las fases.
Al último nivel: Mantiene la estructura del presupuesto con partidas y subpartidas.
Al primer nivel: El presupuesto es mostrado como en la hoja del presupuesto, sin subpartidas.
SI YA TIENE ASIGNADO LAS PARTIDAS A LAS FESES – Y CAMBIA DE NIVEL SE BORRAN LAS ASIGNACIONES.
: Asignación automática, genera automáticamente una estructura en las fases y asigna las partidas del presupuesto.
: Asigna sólo el presupuesto actual, sí el proyecto tiene varios presupuestos asignados, esta opción efectúa una asignación automática del presupuesto activo (que es mostrado en la ventana superior de este escenario).
Ubicar actividad con saldo, verifica que todas las partidas estén asignadas a las fases. Para efectuar esta verificación el contenido de la ventana superior tiene que estar expandida.
7.2 Botones del escenario de períodos
: Elimina los períodos excedentes que se generan en base a la fecha de inicio y plazo, requisito no debe haber, avances programados o valorizados en el escenario del cronograma por periodos.
: Actualiza los datos del proyecto, de acuerdo a la modificación que se realice de los datos de fecha inicio y fin del escenario de datos generales.
: Toma los nuevos datos de inicio y fin del proyecto.
8 SEPARATAS
Ejemplo 1
Este es el subpresupuesto de gastos generales del presupuesto Puente en carretera meta.
El presupuesto meta: No tiene: pie de presupuesto, cálculo analítico de gastos generales, fórmula polinómica.
Ejemplo 2
Este es un presupuesto VENTA, de algunas partidas de edificación, se caracteriza porque dentro de los análisis de precios unitarios se considera herramientas manuales, y equipo menor en las partidas que necesita. En plazo considere: 45 días.
Disponible solo para la versión ERP: Tipo de obra E-1: Acabados costosos importados.
Área de construcción: 240 m2

El texto completo también está troceado en `kb/fragmentos.jsonl` y se cita con el nombre del manual y la página.

**Confiabilidad:** oficial S10
