---
title: REP-01 — Constantes por fecha (UIT, RMV) vigentes en un rango
summary: "Reporte aprobado. Ejemplos: qué valor tiene la UIT; UIT 2024."
tags:
  - reporte
  - sql
id: 01M3RSX1B0RXJWHRZH9D9B8G2Q
status: active
updated_by: ai-agent
updated_at: 2026-09-30T09:21:33.024Z
---

Reporte SQL aprobado del ERP S10: REP-01 «Constantes por fecha (UIT, RMV) vigentes en un rango».
Responde preguntas como: qué valor tiene la UIT; UIT 2024; remuneración mínima vital 2020; constantes vigentes.
Parámetros: desde (fecha, p. ej. vacío = sin filtro); hasta (fecha, p. ej. vacío = sin filtro).
SQL pre-aprobado (la IA solo lo rellena, nunca lo inventa): SELECT concepto, abreviatura, desde, hasta, valor, observaciones FROM constantes_fecha WHERE (:desde IS NULL OR hasta >= :desde) AND (:hasta IS NULL OR desde <= :hasta) ORDER BY concepto, desde
Se ejecuta con el catálogo de reportes.py: validación sqlglot (solo SELECT), sesión de base de datos de solo lectura y máximo 500 filas.

**Uso:** `reportes.py preguntar "…"` — la plantilla solo se rellena, no se edita. Cambios de SQL = nueva versión del catálogo revisada a mano.
