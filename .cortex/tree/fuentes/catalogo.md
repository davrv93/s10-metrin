---
title: Catálogo de fuentes y confiabilidad
summary: "Inventario: portal de ayuda público, 22 PDF de s10peru.com, videos del
  canal oficial, dos copias de terceros (Scribd, pdfcoffee), tesis UPN e informe
  CONCYTEC. Faltan los manuales tras el login de miembros. Incluye tabla de
  niveles."
tags:
  - catalogo
  - confiabilidad
  - fuentes-pendientes
  - interno-del-proyecto
id: 01M3RJS6VGTD8MS2RZ64MHHC35
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:17:07.568Z
---

## Qué hay

| Origen | Contenido | Cómo entró |
|---|---|---|
| documentacion.s10peru.com | Páginas públicas del portal de ayuda: detracciones, anticipos, Lookahead, Calidad Móvil, prerrequisitos y menús | `s10kb.py rastrear --sin-login` |
| www.s10peru.com | 22 PDF: sílabos por módulo, manual del Sistema de Ticket, guía de asistencia remota, requisitos de cursos, políticas SGSST y SGSI | cola manual (`data/enlaces.jsonl`) |
| Canal YouTube «Marketing S10» | 38 videos (ver `fuentes/youtube`) | `youtube.py` |
| repositorio.upn.edu.pe | Tesis sobre la implementación de S10 en la logística de una constructora (2025) | acceso abierto |
| transparencia.concytec.gob.pe | Informe técnico de evaluación de S10 (2013) | documento público |
| Scribd y pdfcoffee | Guía de usuario de S10 Presupuestos (192 pp., OCR) y Manual de S10 Costos y Presupuestos (126 pp.) | importación manual, 29 y 30-09-2026 |

## Niveles de confiabilidad

| Nivel | Qué incluye | Cómo usarlo |
|---|---|---|
| **oficial S10** | PDF de s10peru.com, portal de ayuda, videos del canal oficial | Base de la respuesta |
| **copia de tercero sin verificar** | Guía de usuario de S10 Presupuestos (Scribd) y Manual de costos y presupuestos (pdfcoffee) | Útiles, pero pueden ser versiones antiguas; avisar al usuario |
| **caso académico** | Tesis UPN (2025) | Contexto y experiencia de uso, no procedimientos oficiales |
| **documento del Estado** | Informe CONCYTEC Nº009 (2013) | Histórico; los requisitos técnicos y precios están desfasados |
| **interno del proyecto** | METRIN.md, FUENTES.md, README.md | Reglas del proyecto, no datos de S10 |

## Qué falta

1. **Los manuales oficiales del portal de ayuda.** FUENTES.md cuenta 78 páginas tras el muro de miembros. En `data/paginas/`, 64 de los 90 archivos muestran «Debes acceder para ver este contenido». Afecta a los manuales por módulo, portales, apps móviles, instalación, SQL Server y preguntas frecuentes. Se espera la cuenta de miembro de Arturo o una copia que entregue S10.
2. **La tesis UPC (Inocencio):** el enlace directo devuelve HTML; hay que bajarla a mano.
3. **Otras copias de terceros** en Scribd, Academia.edu y Slideshare: piden cuenta o captcha y no se han bajado.

**Confiabilidad:** interno del proyecto

**Fuentes**
- FUENTES.md
- data/enlaces.jsonl, data/pdf_publicos_s10peru.json
- data/paginas/ (conteo de páginas con muro)
