---
title: Anticipo de proveedor facturado y amortizado
summary: "Anticipo de proveedor sobre una OC: la factura 1 se relaciona con la
  OC con el recurso ANTICIPO. Al llegar la mercadería se quita la regularización
  del ingreso, la factura 2 se relaciona con el ingreso y se amortiza con un
  correctivo ANTICIPO en negativo. Cierra una regularización compuesta."
tags:
  - facturacion
  - anticipos
  - orden-de-compra
  - almacen
  - oficial
id: 01M3RJP5YZVSPV7CB7XHNWVP8N
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.351Z
---

Caso del portal «Manual de factura de anticipos de Almacén con compuesta». Se usa cuando el proveedor factura un porcentaje de la orden de compra antes de despachar y después emite la factura por el total menos el anticipo. Ejemplo del portal: adelanto del 30 % sobre la OC 585.

**1. Factura del anticipo**
1. En el registro de documentos por pagar, clic derecho **Adicionar** y llenar los datos de la factura.
2. En **Relacionado con**, elegir «ORDEN DE COMPRA O SERVICIO» y seleccionar la OC. El sistema la trata como adelanto y deja editar el monto.
3. El sistema asigna el recurso **ANTICIPO**, que se puede cambiar según la moneda y la contabilidad. Cantidad **1** y, en P.U. sin IGV, el monto facturado.

**2. Ingreso a almacén**
4. Registrar el ingreso de la mercadería. El sistema lo enlaza por defecto con la primera factura, así que hay que desvincularlo con clic derecho **Quitar regularización**. El ingreso debe quedar con la segunda factura.

**3. Factura final y amortización**
5. Registrar la segunda factura relacionada con el **ingreso de almacén**.
6. Sobre el recurso, clic derecho **Adicionar correctivo**, elegir **ANTICIPO** y poner el monto **en negativo** (el 30 % de la OC).

**4. Enlace final**
7. Hacer una **regularización compuesta**. En la pestaña **Documentos**, agregar la primera factura.
8. Abrirla con doble clic, ir a la pestaña **Asignación de precio** y pulsar **Aceptar**.

La página muestra capturas en cada paso, pero las imágenes no se descargaron.

**Confiabilidad:** oficial S10 (portal de ayuda público).

**Fuentes:** data/paginas/caso-anticipo.md (caso 1, pasos 1.1 y 2); data/texto/syllabus-administrativo-s10erp.txt (día 1).
