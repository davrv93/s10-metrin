---
title: Anticipos a proveedor y regularización compuesta
summary: "Si el proveedor factura un % de la OC como anticipo: registrar la
  factura 1 relacionada con la OC (recurso ANTICIPO); al llegar el material,
  quitar la regularización del ingreso, registrar la factura 2 contra el
  ingreso, amortizar con correctivo negativo y cerrar con regularización
  compuesta."
tags:
  - almacenes
  - anticipos
  - facturacion
  - orden-de-compra
  - oficial
id: 01M3RJP5XKCYKP73KJ0MN5V607
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.307Z
---

Caso público del portal de ayuda ("Manual de factura de anticipos de Almacén con compuesta"). Se aplica cuando el proveedor factura un porcentaje de la OC como adelanto antes de despachar, y en una segunda factura entrega todo el material descontando ese anticipo.

**Ejemplo del portal:** OC 585 por una caja de tóner, con anticipo del 30 % en la FT 001-000020 y entrega del material con la FT 001-000025.

**Pasos:**
1. **Factura del anticipo.** Clic derecho → *Adicionar* → llenar los datos de la factura → en *Relacionado con* elegir **"ORDEN DE COMPRA O SERVICIO"** → seleccionar la OC. Por configuración previa, el sistema asigna el recurso **ANTICIPO**, que se puede reasignar según la moneda y la contabilidad. Cantidad 1 y P.U. sin IGV igual al monto facturado. El monto del adelanto es editable.
2. **Ingreso al almacén.** El sistema enlaza el ingreso con la primera factura por defecto. Hay que hacer clic derecho → **quitar regularización**, porque el ingreso corresponde a la segunda factura.
3. **Segunda factura**, relacionada con el ingreso de almacén.
4. **Amortización.** Sobre el recurso, clic derecho → **adicionar correctivo** → recurso ANTICIPO con **monto negativo** igual al 30 % de la OC.
5. **Regularización compuesta**, para que todo quede enlazado. En la pestaña *Documentos* se agrega la primera factura, se hace doble clic, luego pestaña *Asignación de precio* → *Aceptar*.

**Error típico que evita:** dejar el ingreso ligado a la factura del anticipo (paso 2), con lo que el material queda regularizado contra el documento equivocado.

Nota: el ejemplo del portal tiene fechas incoherentes (anticipo el 30/04/2017 y entrega el 05/05/2016). No afecta a los pasos.

**Confiabilidad:** oficial S10 (portal de ayuda).

**Fuentes**
- data/paginas/caso-anticipo.md (documentacion.s10peru.com/caso-anticipo/)
