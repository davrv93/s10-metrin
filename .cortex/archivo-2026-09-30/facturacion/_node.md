---
title: Facturación en S10 ERP
summary: "Facturación registra y administra los documentos por pagar y por
  cobrar: facturas, boletas, notas de crédito y débito, anticipos y documentos
  internos. Controla vencimientos, recibe los documentos de pago de Nóminas y
  fija el plazo de envío a SUNAT."
tags:
  - facturacion
  - documentos-por-pagar
  - documentos-por-cobrar
  - oficial
id: 01M3RJ9FKQXZ51X34XPX5VG31Q
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.218Z
---

El portal lo describe como el módulo que ayuda «con el registro, creación y administración de los documentos por pagar y cobrar» y permite controlar las fechas de vencimiento de pagos y cobros.

**Qué entra en Facturación**
- **Documentos por pagar** de proveedores: sin relación o relacionados con orden de compra, orden de servicio, guía de almacén o subcontrato.
- **Documentos por cobrar**: sin relación, desde un documento de venta o desde el cronograma de facturación de una orden de venta.
- **Notas de crédito y débito** de los dos lados.
- **Documentos internos**: anticipos y entregas a rendir.
- **Documentos de pago** que vienen de Nóminas (planillas, gratificación, CTS por banco, aportes). Nóminas los envía y aparecen aquí con el detalle por colaborador.

**Hijos de esta rama**
- `facturacion/documentos-y-catalogos-tributarios`: catálogos de impuestos y documentos, detracción y percepción, padrones, talonarios.
- `facturacion/plazo-envio-documento-electronico`: ampliar el plazo de envío a SUNAT.
- `facturacion/anticipo-de-proveedor`: anticipo facturado y su amortización.

**Salida:** tras registrar el documento, el pago se tramita en el módulo **Administrativo** con órdenes de pago y aprobaciones. Si la factura está afecta a detracción, el cronograma de pagos debe separar la detracción.

**Contenido restringido:** «Manual de Facturación», «Manual de Facturación Electrónica», «Configuración del catálogo de NO deducibles», «Anulación de Ventas» y «Asignar código de producto Sunat» existen en el portal, pero son solo para miembros.

**Confiabilidad:** oficial S10 (portal de ayuda, sílabos y videos del canal oficial).

**Fuentes:** data/paginas/facturacion.md; data/texto/syllabus-administrativo-s10erp.txt (días 1 y 3); video n9i-5fjeqKE (4:27-4:54); data/paginas/realizar-pago-factura-afecta-detraccion-subir-constancias.md.
