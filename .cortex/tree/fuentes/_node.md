---
title: Fuentes de la base de conocimiento
summary: "De dónde sale cada dato de esta base y cuánto fiarse: PDF oficiales de
  s10peru.com, portal de ayuda público, videos del canal oficial, copias de
  terceros, una tesis y un informe del Estado. Faltan los manuales tras el
  login."
tags:
  - fuentes
  - procedencia
  - confiabilidad
  - interno-del-proyecto
id: 01M3RJ9FKQXZ51X34XPX5VG31V
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:17:07.485Z
---

Esta rama documenta la **procedencia** del conocimiento. Sirve para que Metrín sepa cuánto puede afirmar y cuándo debe advertir al usuario.

## Hijos

1. `fuentes/catalogo` — inventario de fuentes, niveles de confiabilidad y lo que falta.
2. `fuentes/youtube` — el canal oficial de S10, sus videos y cómo se transcribieron.
3. `fuentes/metrin` — el asistente que responde con esta base: personaje, tono y reglas de citación.
4. `fuentes/informe-concytec` — la evaluación del Estado sobre S10 Presupuestos (2013).

## Principios

1. **Solo lo que dicen las fuentes.** Si un dato no está, se dice que no está.
2. **Cada nodo cita** su archivo y su página, o el video con su título.
3. **Cada nodo declara su confiabilidad** en una línea antes de «Fuentes»: *oficial S10*, *documento del Estado* o *interno del proyecto*. El catálogo añade dos niveles más para el material de los módulos: *copia de tercero sin verificar* y *caso académico*.
4. **El hueco principal** son los manuales oficiales del portal de ayuda, que están tras el login de miembros. Muchos nodos de portales y soporte dicen «deriva al manual» por esa razón.

## Cómo se obtuvieron

- `s10kb.py` rastrea el portal de ayuda y baja los PDF enlazados.
- `youtube.py` baja el audio de los videos oficiales y lo transcribe en local.
- Los documentos de terceros se importaron a mano en `entrada/`.

**Confiabilidad:** interno del proyecto

**Fuentes**
- FUENTES.md
- README.md
- youtube.py (cabecera)
