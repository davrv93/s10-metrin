# Cómo realizar el pago de una factura afecta a detracción y subir las constancias al sistema 10

Fuente: https://documentacion.s10peru.com/realizar-pago-factura-afecta-detraccion-subir-constancias/

Al tener una factura afecta a detracción, se debe tener en consideración que la detracción se debe mostrar independiente en el registro. Ver imagen
Validar en el campo “CRONOGRAMA DE PAGOS”, que la detracción se encuentre independiente, para poder de esta manera en el modulo administrativo procesar el pago exclusivamente de la detracción. Ver imagen
Para el pago de las detracciones, nos ubicamos en el modulo administrativo, escenario pagos, y nos ubicamos en el campo ordenes de pago. Clic derecho, adicionar, y vamos a elegir la opción tributos. Ver imagen
Se tiene que indicar dentro de la orden de pago, que dicha orden será como pago electrónico, luego seleccionar si el pago que se va realizar será una detracción o autodetracción, para el presente caso estamos indicando que el pago que realizaremos será detracción. Ver imagen
Al realizar el procedimiento antes mencionado hacer clic derecho en la parte inferior espacio en blanco, y daremos clic en la única opción “Documento con tributo”. Y se procede a seleccionar el documento que se quiere cancelar.
Ver imagen.
Luego de generar el pago, se tiene que enviar a aprobación, clic derecho “Enviar a aprobación”. Ver imagen
Luego de enviar la aprobación, clic derecho aprobar y confirmar la aprobación. Ver imagen
Luego para visualizar el documento aprobado, ir al campo “Pendientes de pago” y seleccionar tributos, ya que el documento que se está pagando es una detracción.
Luego hacer clic derecho, y seleccionar la opción pago electrónico, la finalidad de seleccionar la opción es que el sistema emite automáticamente un txt con el detalle de todos los documentos para subir a la página de la SUNAT. Ver imagen.
Dentro del pago propiamente ya generado también el sistema nos alerta para generar el archivo del pago electrónico. Ver imagen
Luego de generar, el archivo de texto lo podemos ubicar en la siguiente ruta: Disco C, carpeta S102000, carpeta Pagos. Donde encontraremos el archivo de texto y para confirmar se procede a abrir el archivo. Dicho archivo de texto se va utilizar para realizar el pago de las detracciones desde la web de SUNAT, utilizando la clave sol. Inmediatamente luego de realizar el pago mediante la Web de la SUNAT, se obtendrá de la misma web un archivo de texto con todas las constancias y sus respectivos detalles.
La manera de importar el archivo de texto con los números de constancia de cada documento pagado, es ubicarnos en el escenario pagos, ir al pago que se ha generado, clic derecho y seleccionar la opción de “importar constancias de detracción”
Finalmente, al culminar la importación de las constancias de detracción, para poder comprobar que efectivamente se llegó a importar las constancias de detracción, ingresamos a la factura y podremos apreciar que se encuentra ya registrado el número de detracción, requisito indispensable para el correcto envío del libro electrónico de compras. Ver imagen
