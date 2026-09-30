---
title: REP-04 — Costo de planilla por concepto en un periodo
summary: "Reporte aprobado. Ejemplos: cuánto costó la planilla de junio 2024;
  costo de planilla por concepto."
tags:
  - reporte
  - sql
id: 01M3RSX1FHS4H0VQW4ZKPVRCA4
status: active
updated_by: ai-agent
updated_at: 2026-09-30T09:21:33.169Z
---

Reporte SQL aprobado del ERP S10: REP-04 «Costo de planilla por concepto en un periodo».
Responde preguntas como: cuánto costó la planilla de junio 2024; costo de planilla por concepto.
Parámetros: periodo (periodo, p. ej. 2024-06).
SQL pre-aprobado (la IA solo lo rellena, nunca lo inventa): SELECT p.concepto, COUNT(DISTINCT p.trabajador_id) AS trabajadores, ROUND(SUM(p.monto), 2) AS total FROM planilla p WHERE p.periodo = :periodo GROUP BY p.concepto ORDER BY total DESC
Se ejecuta con el catálogo de reportes.py: validación sqlglot (solo SELECT), sesión de base de datos de solo lectura y máximo 500 filas.

**Uso:** `reportes.py preguntar "…"` — la plantilla solo se rellena, no se edita. Cambios de SQL = nueva versión del catálogo revisada a mano.
