# Manual de factura de anticipos de Almacén con compuesta

Fuente: https://documentacion.s10peru.com/caso-anticipo/

Caso Anticipo
1.- Anticipo facturado (en factura 1 y amortización en factura 2):
Este caso aplica cuando el proveedor nos factura un % del valor total de la OC como anticipo, es decir aun no nos despacha el material pero nos solicita el adelanto para asegurar la venta o por otros fines; y en la segunda factura que nos emita se considerara el ingreso total del material que en la mayoría es por el valor total de la OC menos el anticipo otorgado en la primera factura.
Ejemplo: La empresa DIMERC PERU S.A.C, nos venderá una caja de cartucho toner con OC 585 proyecto ADM-008-POS-2016 y nos pide un adelanto del 30% el cual nos factura con FT 001-000020 con fecha 30/04/2017 y el 05/05/2016 nos entregan mercadería con factura FT 001-000025.
PASOS PARA EL REGISTRO:
1.1- Registro de factura por el 30% de adelanto de la OC 585, al relacionarlo directamente desde la OC el sistema lo considera como adelanto, teniendo la facultad de editar el monto por el cual será el adelanto.
* Click derecho adicionar>>> llenamos los datos de la factura >>> y en el campo Relacionado con: elegimos “ORDEN DE COMPRA O SERVICIO”.
*Seleccionamos la OC
*Por defecto de acuerdo a una previa configuración el sistema le asigna a la factura el recurso “ANTICIPO” el cual es reasignable (para fines de la moneda y la contabilidad deseada), en cantidad ponemos (1) y en el P.U sin IGV colocamos el monto que nos factura el proveedor.
2.- El proveedor envía la mercadería y segunda factura donde nos indica que nos envía toda la mercadería y la amortización del anticipo del 30%, por lo tanto hay que realizar el ingreso al almacén y el registro de la factura.
Ingreso al almacén de mercadería:
Por defecto el sistema enlaza el ingreso con el registro de la primera factura por lo tanto hay que desvincular con click derecho quitar regularización, por la sencilla razón de que este ingreso debe estar relacionado con la segunda factura.
Ahora registraremos la segunda factura esta vez relacionada con el ingreso de almacén:
Ahora tenemos la 2da factura registrada pero falta amortizar el 30% anticipo
La amortización se realiza ubicándonos sobre el recurso con click derecho adicionar correctivo y elegimos el recurso ANTICIPO pero esta vez con el monto en negativo por el valor del 30% de la OC
Ahora para culminar y si deseamos que todo este enlazado debemos realizar una regularización compuesta, con la finalidad de que todo guarde relación.
Nos ubicamos en la pestaña documentos y agregamos la primera factura creada
Luego de seleccionar con doble click nos figura la siguiente ventana donde nos ubicaremos luego en la pestaña asignación de precio y le damos aceptar y habremos culminado este caso.
