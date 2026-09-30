---
title: REP-06 — Depósito de CTS por trabajador en un periodo
summary: "Reporte aprobado. Ejemplos: CTS de noviembre 2024; depósito de CTS."
tags:
  - reporte
  - sql
id: 01M3RSX1GESCW14PJM4T4A0APK
status: active
updated_by: ai-agent
updated_at: 2026-09-30T09:21:33.198Z
---

Reporte SQL aprobado del ERP S10: REP-06 «Depósito de CTS por trabajador en un periodo».
Responde preguntas como: CTS de noviembre 2024; depósito de CTS.
Parámetros: periodo (periodo, p. ej. 2024-11).
SQL pre-aprobado (la IA solo lo rellena, nunca lo inventa): SELECT w.nombre, w.categoria, ROUND(p.monto, 2) AS cts FROM planilla p JOIN trabajadores w ON w.id = p.trabajador_id WHERE p.concepto LIKE 'CTS%' AND p.periodo = :periodo ORDER BY cts DESC
Se ejecuta con el catálogo de reportes.py: validación sqlglot (solo SELECT), sesión de base de datos de solo lectura y máximo 500 filas.

**Uso:** `reportes.py preguntar "…"` — la plantilla solo se rellena, no se edita. Cambios de SQL = nueva versión del catálogo revisada a mano.
