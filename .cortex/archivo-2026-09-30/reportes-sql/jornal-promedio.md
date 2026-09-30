---
title: REP-03 — Promedio de horas tareadas por categoría
summary: "Reporte aprobado. Ejemplos: promedio de horas por categoría; quién
  trabaja más horas operarios o oficiales."
tags:
  - reporte
  - sql
id: 01M3RSX1EYVECRGDESWHZ7WZ8M
status: active
updated_by: ai-agent
updated_at: 2026-09-30T09:21:33.150Z
---

Reporte SQL aprobado del ERP S10: REP-03 «Promedio de horas tareadas por categoría».
Responde preguntas como: promedio de horas por categoría; quién trabaja más horas operarios o oficiales.
Parámetros: desde (fecha, p. ej. 2024-01-01); hasta (fecha, p. ej. 2024-12-31); categoria (texto, p. ej. vacío = sin filtro).
SQL pre-aprobado (la IA solo lo rellena, nunca lo inventa): SELECT w.categoria, ROUND(AVG(t.horas), 1) AS horas_promedio, COUNT(*) AS registros FROM tareas t JOIN trabajadores w ON w.id = t.trabajador_id WHERE (:categoria IS NULL OR w.categoria = :categoria) AND t.fecha BETWEEN :desde AND :hasta GROUP BY w.categoria ORDER BY w.categoria
Se ejecuta con el catálogo de reportes.py: validación sqlglot (solo SELECT), sesión de base de datos de solo lectura y máximo 500 filas.

**Uso:** `reportes.py preguntar "…"` — la plantilla solo se rellena, no se edita. Cambios de SQL = nueva versión del catálogo revisada a mano.
