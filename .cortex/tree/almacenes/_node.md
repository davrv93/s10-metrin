---
title: Almacenes (S10 ERP)
summary: "Almacenes en S10: configurar almacenes por proyecto (partidas de
  control, frentes, destinos, ubicación), registrar ingresos (11 tipos: I OC, I
  TI…) y egresos (13 tipos: E PC, E TE…), y consultar stock, kardex, recursos
  con problemas y equipos disponibles."
tags:
  - almacenes
  - inventario
  - s10
  - logistica
  - oficial
id: 01M3RJ9FKPFZXSME3CGDB9YQMY
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.205Z
---

El módulo de Almacenes registra cada movimiento de materiales del proyecto con un tipo de documento codificado. El curso oficial dura **4 días de 4 horas**.

**Estructura del curso oficial:**
1. **Configuración:** usuarios y permisos, proyecto y usuarios de proyecto y almacén, nuevos almacenes, documentos de movimiento, partidas de control, destinos específicos, frentes, roles por unidad operativa y ubicación física.
2. **Ingresos:** 11 tipos, desde la compra y la orden de compra hasta transferencias, préstamos, fabricación e inventario.
3. **Egresos:** 13 tipos, entre ellos salidas a partida de control (con descarga directa o con pedido), transferencias, préstamos, bajas y mermas y devoluciones.
4. **Resultados:** almacén de activos, fondos rotatorios, stock, **kardex**, recursos con problemas y, a nivel general, ingresos y egresos, stock general y equipos disponibles.

**Relación con otros módulos:**
- Los ingresos por OC cierran el ciclo de [compras].
- Los egresos a partida de control imputan el consumo a la estructura de control de [gerencia-proyectos], con la que se agrupan los resultados operativos.
- La factura de un anticipo al proveedor se enlaza con la OC y el ingreso mediante una regularización compuesta.

**Hijos:** configuración de almacenes; movimientos de ingreso; egresos y transferencias; anticipos y regularización compuesta; caso UPN (ingresos, salidas e inventario).

Limitación: el Manual de Almacenes del portal exige cuenta, así que solo se conocen los temas de los sílabos y el caso de anticipo, que es público.

**Confiabilidad:** oficial S10 (sílabos y portal de ayuda).

**Fuentes**
- data/texto/syllabus-almacenes-s10erp.txt (Curso de Almacenes, 4 días)
- data/texto/silabus-almacenes.txt
- data/paginas/caso-anticipo.md; data/paginas/manual-de-almacenes.md (restringido)
