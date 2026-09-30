---
title: Movimientos de ingreso al almacén
summary: "S10 codifica 11 tipos de ingreso: I AC (ajuste), I C (compra), I DP
  (devolución de préstamo), I DS (devolución de saldo), I F (fabricación), I I
  (inventario), I OC (orden de compra), I P (préstamo), I R (reintegro), I TE /
  I TI (transferencia externa/interna)."
tags:
  - almacenes
  - ingresos
  - orden-de-compra
  - oficial
id: 01M3RJP5XC910SWF0YCDZEM7M4
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.300Z
---

El segundo día del curso oficial de Almacenes cubre los ingresos. Cada tipo tiene su propio código de documento:

| Código | Movimiento |
|---|---|
| I AC | Ingreso por Ajuste de Cantidad |
| I C | Ingreso por Compra |
| I DP | Ingreso por Devolución de Préstamo |
| I DS | Ingreso por Devolución de Saldo |
| I F | Ingreso por Fabricación |
| I I | Ingreso por Inventario |
| I OC | Ingreso por Orden de Compra |
| I P | Ingreso por Préstamo |
| I R | Ingreso por Reintegro |
| I TE | Ingreso por Transferencia Externa |
| I TI | Ingreso por Transferencia Interna |

Después vienen los **reportes e informes de ingresos**.

**Cómo encaja:**
- **I OC** es el que cierra una orden de compra emitida en Compras.
- Cada ingreso de transferencia o préstamo tiene su contraparte en egresos: E TE/E TI y E P/E DP.
- En el caso de anticipos, el portal oficial indica que el sistema enlaza el ingreso con la primera factura (la del anticipo) y que hay que quitar esa regularización para relacionarlo con la segunda factura (ver almacenes/anticipos-y-regularizacion-compuesta).

**Referencia académica:** en la tesis UPN, el ingreso lo registra el analista logístico en S10 después de la verificación documental y física y de la firma de la guía por el residente. Es la práctica de una empresa concreta.

Limitación: los sílabos no explican la diferencia entre I C e I OC ni cuándo usar I DS o I R.

**Confiabilidad:** oficial S10 (con una mención puntual a la tesis UPN).

**Fuentes**
- data/texto/syllabus-almacenes-s10erp.txt (Segundo día)
- data/texto/silabus-almacenes.txt (Segundo día)
- data/paginas/caso-anticipo.md
- data/texto/an-lisis-de-la-implementaci-n-del-sistema-erp-s10-en-la-log-stica-de-una-constru.txt (tesis UPN 2025): pp. 17-19
