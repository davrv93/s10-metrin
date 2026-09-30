---
title: Ingresos por Identificar – Procesos de IPI
summary: Los IPIS son aquellos abonos que han sido depositados en la cuenta
  recaudadora pero que aun no han sido identificados. Mediante el modulo de
  facturación se debe ubicar la siguiente opción en forma de tuerca Debe ubicar
  la pestaña Sperant (3) es aquí donde se realizará la…
tags:
  - oficial-s10
  - manual
  - portal-miembros
id: 01M3SP28BSCPKYMRZAYNZG749E
status: active
updated_by: ai-agent
updated_at: 2026-09-30T17:47:59.694Z
---

Manual oficial de S10, bajado del portal de miembros con la cuenta de la empresa.

- **Página:** https://documentacion.s10peru.com/ingresos-identificar-proceso-ipi/

## Texto de la página

# Ingresos por Identificar – Procesos de IPI

Fuente: https://documentacion.s10peru.com/ingresos-identificar-proceso-ipi/

Los IPIS son aquellos abonos que han sido depositados en la cuenta recaudadora pero que aun no han sido identificados.
S10: Configuración
Table of Contents
Toggle
Mediante el modulo de facturación se debe ubicar la siguiente opción en forma de tuerca
Debe ubicar la pestaña Sperant (3) es aquí donde se realizará la configuración General,
Previamente se debe registrar en el catálogo tipo de Documento y Recurso Especifico la información a utilizar.
S10: Recursos de Venta
Acceda al siguiente escenario
El usuario debe registrar los recursos de ventas, para este caso solo debe ser adicionado un solo recurso siendo IPI.
Esta acción se realiza de forma global y No por EMPRESA.
A continuación mostrara la venta Vectores de Venta, siendo aquí donde debe adicionar el recurso que previamente fue registrado por el usuario en el Catalogo de Recursos.
Finalmente dar clic en la opción Aceptar para que complete el registro
S10: Lista de Precios
Acceda al siguiente escenario
Es el escenario donde se registra la Lista de Precio, cabe indicar que debe realizarse por cada EMPRESA
Seguidamente debe adicionar a la Lista de Precio el recurso de venta considerando como dato básico del recurso un precio y cantidad 1.
Teniendo como resultado la lista de precio que será utilizada en la OV global.
S10: Registro de socio de Negocio
Se adiciona el registro de un cliente global
S10: Orden de Venta
Se realiza el registro de una orden de venta global por cada EMPRESA,
Se registra la OV con los datos previamente configurados como el Cliente, la Lista de Precio y Proyecto.
Se recomienda hacer uso del Proyecto Principal de cada EMPRESA
Adicionalmente mediante la pestaña Sperant, se debe marcar el check de Ingresos por Identificar es así como se define que la orden de venta será de tipo IPI.
SPERANT : Envío de IPIs
Desde Sperant debe ubicar el modulo Control de Pagos seguidamente de Pagos dando clic en la opción Pagos no Identificados
Seguidamente mediante la opción Nuevo IPI, permitirá realizar el registro del abono
A continuación debe ubicarse el registro del IPI para que este seaenviado al S10
Puede haber demora en que la orden de venta sea reconocida, espere un momento antes de sincronizar o puede enviarse duplicados.
Puede verificar la sincronización verificando el campo ID EXTERNO
S10: Verificación de recepción de IPIs
Para ello se debe ingresar al escenario Por Cobrar y ubicar el Tipo de documento que se registro en la configuración global.

El texto completo también está troceado en `kb/fragmentos.jsonl` y se cita con el nombre del manual y la página.

**Confiabilidad:** oficial S10
