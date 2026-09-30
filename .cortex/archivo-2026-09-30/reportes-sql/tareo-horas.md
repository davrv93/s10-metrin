---
title: REP-02 — Horas y trabajadores tareados por día
summary: "Reporte aprobado. Ejemplos: horas tareadas de mayo 2024; asistencia de
  operarios."
tags:
  - reporte
  - sql
id: 01M3RSX1BWNCSVXY7N5MQJS6MA
status: active
updated_by: ai-agent
updated_at: 2026-09-30T09:21:33.052Z
---

Reporte SQL aprobado del ERP S10: REP-02 «Horas y trabajadores tareados por día».
Responde preguntas como: horas tareadas de mayo 2024; asistencia de operarios; tareo por día.
Parámetros: desde (fecha, p. ej. 2024-01-01); hasta (fecha, p. ej. 2024-12-31); categoria (texto, p. ej. vacío = sin filtro).
SQL pre-aprobado (la IA solo lo rellena, nunca lo inventa): SELECT t.fecha, COUNT(*) AS trabajadores, SUM(t.horas) AS horas FROM tareas t JOIN trabajadores w ON w.id = t.trabajador_id WHERE (:categoria IS NULL OR w.categoria = :categoria) AND t.fecha BETWEEN :desde AND :hasta GROUP BY t.fecha ORDER BY t.fecha
Se ejecuta con el catálogo de reportes.py: validación sqlglot (solo SELECT), sesión de base de datos de solo lectura y máximo 500 filas.

**Uso:** `reportes.py preguntar "…"` — la plantilla solo se rellena, no se edita. Cambios de SQL = nueva versión del catálogo revisada a mano.
