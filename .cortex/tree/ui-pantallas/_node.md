---
title: Pantallas del ERP documentadas por visión
summary: "Pantallas reales del ERP S10 reconstruidas desde los videos oficiales.
  Validada: Constantes por Fecha General (Nóminas). Los nodos por captura son
  duplicados reemplazados."
tags:
  - ui
  - pantalla
id: 01M3RQG0AW4HDM2T81W6XVPJ2G
status: active
updated_by: ai-agent
updated_at: 2026-09-30T09:02:43.803Z
---

Método: los videos del canal oficial muestran las pantallas reales del ERP. Por captura se corre OCR (tesseract, cadena exacta y caja) y un VLM local (gemma3:4b vía Ollama, semántica: qué campo es y de qué tipo), y se fusionan: la etiqueta final es la del OCR solo si coincide palabra por palabra, el valor y el tipo vienen del VLM. El resultado es un borrador revisable en el HTML que genera `ui.py render` (arrastrar y exportar).

**Lecciones del piloto (2026-09-30, video 17W0yKI8iew):** 60 capturas = 1 pantalla en dos estados (grilla y detalle); los títulos de ventana que lee la máquina NO sirven (S10→510/$10, la barra de estado aparece como pantalla), por eso los títulos de los nodos se fijan a mano en CANONICAS y un nodo por video. Limitaciones medidas: el VLM confunde letras (UIT→UT) y lee el calendario del date-picker; el OCR a veces cruza celdas vecinas. Regenerable con: `ui.py kb && ui.py cortex && herramientas/cargar_cortex.py data/cortex-borradores/pantallas-ui.json`.

**Validación 2026-09-30:** `ui-pantallas/17w0yki8iew` revisada a ojo contra sus 60 capturas; corregidos campos falsos del OCR. Los otros 67 nodos de la rama eran duplicados de esa misma pantalla y quedaron marcados como reemplazados.
