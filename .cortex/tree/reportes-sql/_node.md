---
title: Reportes SQL del ERP (catálogo aprobado)
summary: "Preguntas en lenguaje natural a reportes SIN SQL generado por IA:
  elige y rellena plantillas pre-aprobadas, valida con sqlglot (solo SELECT) y
  ejecuta en sesión read-only."
tags:
  - reportes
  - sql
  - metodo
  - seguridad
id: 01M3RSX1A1Z30C8BCM1X8E3RFM
status: active
updated_by: ai-agent
updated_at: 2026-09-30T09:21:32.994Z
---

Método: nada de SQL libre. La IA recupera la plantilla del catálogo aprobado (reportes/plantillas.json), resuelve parámetros con reglas explícitas (fechas, periodo, categoría) y sqlglot valida por AST que la consulta sea un SELECT único (rechaza INSERT/UPDATE/DROP/ATTACH incluso anidados). La ejecución va en sesión read-only de la BD con tope de 500 filas: la garantía real es el rol de la base, no la cortesía del código.

Catálogo: 6 plantillas (constantes UIT/RMV, tareo, jornal, planilla, gratificaciones, CTS). Probado contra BD demo sembrada con los datos citados del video del piloto (UIT 4300/4400/5150, RMV 930).

Pendiente para la BD real de S10 (debe dar Arturo): cadena de conexión y usuario de SOLO LECTURA. Con eso basta `reportes.py preguntar "…" --bd postgresql://…`.
