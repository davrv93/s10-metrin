---
title: REP-05 — Gratificación por trabajador en un periodo
summary: "Reporte aprobado. Ejemplos: gratificación de julio 2024; quiénes
  cobran gratificación."
tags:
  - reporte
  - sql
id: 01M3RSX1FZV15VGQEHG3W7TRH2
status: active
updated_by: ai-agent
updated_at: 2026-09-30T09:21:33.183Z
---

Reporte SQL aprobado del ERP S10: REP-05 «Gratificación por trabajador en un periodo».
Responde preguntas como: gratificación de julio 2024; quiénes cobran gratificación.
Parámetros: periodo (periodo, p. ej. 2024-07).
SQL pre-aprobado (la IA solo lo rellena, nunca lo inventa): SELECT w.nombre, w.categoria, ROUND(p.monto, 2) AS gratificacion FROM planilla p JOIN trabajadores w ON w.id = p.trabajador_id WHERE p.concepto LIKE 'GRATIFICACION%' AND p.periodo = :periodo ORDER BY gratificacion DESC
Se ejecuta con el catálogo de reportes.py: validación sqlglot (solo SELECT), sesión de base de datos de solo lectura y máximo 500 filas.

**Uso:** `reportes.py preguntar "…"` — la plantilla solo se rellena, no se edita. Cambios de SQL = nueva versión del catálogo revisada a mano.
