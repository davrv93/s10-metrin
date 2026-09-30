---
title: Pago de detracciones y carga de constancias
summary: "Detracción: orden de pago de Tributos marcada como pago electrónico,
  con el documento; aprobarla; en Pendientes de pago > Tributos generar el txt
  (C:\\S102000\\Pagos); pagar en SUNAT con clave SOL e importar las constancias.
  El número queda en la factura para el libro de compras."
tags:
  - administrativo
  - detracciones
  - sunat
  - pagos
  - procedimiento
  - oficial
id: 01M3RJP5Z3TDKZBW65WMADN0DB
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.355Z
---

Guía pública del portal para pagar la detracción de una factura de proveedor.

**Requisito en Facturación:** la detracción debe verse **independiente en el Cronograma de pagos** de la factura. Si no, no se puede pagar por separado.

**Procedimiento**
1. En el módulo **Administrativo**, escenario **Pagos**, ubicarse en **Órdenes de pago**.
2. Clic derecho **Adicionar** y elegir la opción **Tributos**.
3. Indicar que la orden es de **pago electrónico** y elegir **Detracción** (o Autodetracción, según el caso).
4. En la parte inferior, en blanco, clic derecho **Documento con tributo** y seleccionar los documentos que se van a cancelar.
5. Clic derecho **Enviar a aprobación**. Luego clic derecho **Aprobar** y confirmar.
6. Ir a **Pendientes de pago** y seleccionar **Tributos**.
7. Clic derecho **Pago electrónico**. El sistema genera un **archivo .txt** con el detalle de todos los documentos y también avisa desde el pago generado.
8. El archivo queda en **C:\S102000\Pagos**.
9. Pagar las detracciones en la web de **SUNAT** con la clave SOL, subiendo ese txt. SUNAT devuelve otro txt con las **constancias**.
10. En el escenario **Pagos**, sobre el pago generado, clic derecho **Importar constancias de detracción** y cargar el archivo de SUNAT.
11. Abrir la factura y comprobar que tenga el **número de detracción**.

**Por qué importa:** ese número es «requisito indispensable para el correcto envío del **libro electrónico de compras**».

**Contexto del curso:** los ítems de detracción y su asignación a recursos se configuran el día 1, y la orden de pago de tributos y su pago se ven el día 2.

**Confiabilidad:** oficial S10 (portal de ayuda público). La página trae capturas que no se descargaron.

**Fuentes:** data/paginas/realizar-pago-factura-afecta-detraccion-subir-constancias.md (todo el procedimiento); data/texto/syllabus-administrativo-s10erp.txt (días 1-2).
