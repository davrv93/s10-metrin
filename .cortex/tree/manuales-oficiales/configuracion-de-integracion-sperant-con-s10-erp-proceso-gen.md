---
title: Configuración de Integración Sperant con S10 ERP – Proceso General
summary: En el siguiente Gráfico se detalla el proceso de relación Sperant y S10
  La integración de ambas aplicaciones le permitirá como empresa gestionar sus
  proyectos inmobiliarios de manera integral y eficiente. La configuración se
  realiza en el módulo de Facturación, doble clic en el…
tags:
  - oficial-s10
  - manual
  - portal-miembros
id: 01M3SP2880ADY3ASS5E9PP42X6
status: active
updated_by: ai-agent
updated_at: 2026-09-30T18:09:53.683Z
---

Manual oficial de S10, bajado del portal de miembros con la cuenta de la empresa.

- **Página:** https://documentacion.s10peru.com/configuracion-integracion-sperant-s10/

## Contenido

- 1. Generalidades
- 1.1. Acceso al módulo
- 2. Configuración
- 3. Sincronización de Proyectos
- 4. Sincronización de Unidades
- 5. Órdenes de Venta
- 6. Registro de Cuentas Bancarias:
- 7. Registro de Libro Caja
- 8. Conciliación Bancaria
- 9. Boletaje

## Texto de la página

# Configuración de Integración Sperant con S10 ERP – Proceso General

Fuente: https://documentacion.s10peru.com/configuracion-integracion-sperant-s10/

Flujo de Integración
En el siguiente Gráfico se detalla el proceso de relación Sperant y S10
1. Generalidades
Table of Contents
Toggle
La integración de ambas aplicaciones le permitirá como empresa gestionar sus proyectos inmobiliarios de manera integral y eficiente.
1.1. Acceso al módulo
La configuración se realiza en el módulo de Facturación, doble clic en el botón
que aparece el tapiz de su PC.
Ingrese usuario y contraseña correspondiente
2. Configuración
Mediante la barra estándar ubique el siguiente icono en forma de tuerca, dar doble clic
Esta acción invoca a la siguiente ventana donde se da inicio al proceso de integración mediante la configuración de ciertos parámetros como documentos y recursos que permitirán relacionar el SPERANT con S10.
Luego debe ubicarse en la pestaña SPERANT (1),
(1) Se registra mediante el catalogo tipo de documento “Boletas de Venta por Cobrar – Electrónica”, comprobante que será entregado cuando se trate de una persona Natural.
(2) Se registra mediante el catalogo tipo de documento “Factura por Cobrar – Electrónica”, comprobante que será entregado cuando se trate de una persona Jurídica.
(3) Se extrae del catálogo la Forma de pago con la que se realizara la configuración.
(4) Esta información es facilitada por Sperant para la respectiva configuración.
(5) Según la norma el % mínimo de boletaje es del 3% pero queda a libre decisión del cliente.
(6) (8) (10) (11) (12) Se registra los recursos que serán utilizados para la integración, ubicándolos en la siguiente ruta catalogo Recursos/Subcontratos y Servicios/Administrativo/Recursos SPERANT
(7) Se configura el documento que será entregado al cliente por el anticipo
(9) Se registra mediante el catalogo tipo de documento “Comisión”, que se relacionara con la comisión del Vendedor.
3. Sincronización de Proyectos
El siguiente proceso se realiza desde la interface del SPERANT
Previamente debe ingresar al módulo de gerencia
Desde escenario Datos generales ubique en el árbol el proyecto que se requiere relacionar con el SPERANT
Luego, en el escenario de datos generales, seleccione Datos Adicionales
Configure activando la opción remarcada, luego dar clic en aceptar
Seguidamente mediante la interface de Sperant se ubica el proyecto a relacionar
Como forma de verificar la Sincronización podemos ubicarlo en el catálogo de proyectos.
Al dar clic derecho modificar, nos mostrara en el campo Proyecto Sperant el proyecto que ha sido Sincronizado
4. Sincronización de Unidades
De igual forma la relación de unidades con S10 se visualiza en el catálogo de recursos, luego de sincronizar en el Sperant se generan recursos en el S10 dentro del Grupo de nombre Sperant, dentro se divide por Tipo de Obra, luego por Proyecto, los recursos genéricos son por tipo de Inmueble y los recursos específicos son los inmuebles.
5. Órdenes de Venta
Los documentos (Ordenes de Venta) generados desde Sperant se registraran de forma automática en S10 y se pueden visualizar en el escenario Documentos de Venta y se verifica en la columna REGISTRADO POR indica que provienen de SperantS10
Si ingresamos a uno de los Documentos de Venta podemos ver el indicador de la relación del documento con Sperant
– Minuta Firmada: cuando se muestra esta opción activada indica que La minuta ya Fue Firmada en este caso se procede a generar de forma directa el Boletaje según indica la norma. (Se actualiza desde el Sperant)
-Fecha de Minuta nos muestra la fecha cuando fue firmado el documento minuta. (Se actualiza desde el Sperant)
-Boletaje histórico se utiliza cuando ya se tiene documentos anteriores, se actualiza desde el S10
6. Registro de Cuentas Bancarias:
Para registrar los abonos de separación o cuotas iniciales es necesario crear las cuentas bancarias desde el módulo Administrativo
Entre los tipos de Cuentas también se incluyó POS para poder calificar una cuenta de acumulación de todo lo cobrado por POS
7. Registro de Libro Caja
Luego de generar las cuentas, es necesario crear los Libros de Caja como se muestra a continuación, se recomienda crear un libro de caja anual sin elegir la cuenta bancaria para que acumule los abonos de las diferentes cuentas. El diario de caja se registrara automáticamente al momento de realizar el abono desde Sperant.
Diario caja veo los abonos que provienen del Sperant.
8. Conciliación Bancaria
Luego de todo el proceso de cobranza y movimiento bancario el proceso termina con la Conciliación bancaria el proceso se realiza desde el escenario Estado Bancario, en este proceso se compara los movimientos enviados desde el banco ya sea en un TXT como en banco BCP o ingresado de forma manual con los movimientos registrados en el sistema.
9. Boletaje
Para el proceso inmobiliario se implementó un nuevo Grupo del mismo nombre el cual contiene los escenarios Boletaje, Anulación de Venta, Cambio de Titularidad y entrega de Inmuebles para procesos específicos.
a.- Escenario de boletaje en el cual se emiten las boletas en 2 situaciones principales cuando ya se tiene firmado la Minuta o cuando en monto minino se encuentra superado adjunto imagen
También se muestra el tipo de boletaje ABONOS POSTERIORES donde las órdenes de venta que ya han realizado su primer boletaje y que ahora es necesario boletear los demás abonos, en este caso la boleta es por cada abono
b.- El escenario de Anulación de Venta contiene las órdenes de venta enviadas por sperant para ser anuladas, el paso a seguir es generar la nota de crédito y cerrar la orden de venta
C.-Cambio de titularidad y/o inmueble contiene las órdenes de venta que pasaran de un cliente a otro o que están cambiando de un inmueble a otro. El nuevo cliente y/o los nuevos inmuebles estarán en la nueva orden de venta generada automáticamente por Sperant. Los pasos a seguir son generar la nota de crédito en la orden de venta original y generar la boleta en la nueva orden de venta para realizar el pase
d.- Entrega de Inmuebles contiene la información de las órdenes de venta que ya han terminado todo el proceso de venta y el Sperant informa que debe ser entregado. Luego de ello en el S10 se puede definir el costo de venta y con esa información generar los asientos de venta y costo del inmueble entregado de manera automática.

El texto completo también está troceado en `kb/fragmentos.jsonl` y se cita con el nombre del manual y la página.

**Confiabilidad:** oficial S10
