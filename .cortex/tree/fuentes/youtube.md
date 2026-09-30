---
title: Videos del canal oficial S10
summary: Los 38 videos públicos del canal oficial «Marketing S10» se
  transcribieron en local con faster-whisper (modelo small). 15 no tienen
  narración y su transcripción está vacía. Las transcripciones automáticas
  tienen errores de nombres.
tags:
  - youtube
  - videos
  - transcripcion
  - whisper
  - oficial-s10
id: 01M3RJS6VKVK0P58GTCNKBKEAR
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:17:07.571Z
---

## Origen

Los videos vienen del canal **«Marketing S10»** (ID UCL1xwXCVt9ZVpSrgyemP50A), el canal oficial de tutoriales que S10 publica para sus clientes. `youtube.py` recorre el canal y 12 listas de reproducción: Portal de Asistencia, Tareo en Línea, Soporte S10, Pedidos Móviles, Aprende S10, Aprobaciones Móviles, Cursos S10, Nuestra Marca (dos), Portal Colaborador, Calidad Móvil y Portal de Proveedores.

## Cómo se transcribieron

1. `descubrir`: arma el inventario en `data/yt/videos.txt` con el formato id|segundos|título. Salen 38 videos.
2. `audio`: baja **solo el audio** a `data/yt/audio/{id}.m4a`.
3. `transcribir`: usa **faster-whisper en local**, con el modelo *small* en CPU, y guarda `{id}.json` (segmentos con tiempos) y `{id}.txt`.
4. `indexar`: trocea el texto con minutos para `kb/fragmentos_yt.jsonl`.

## Qué cubren

- **Soporte:** generar ticket, reabrir ticket, instalar AnyDesk, acceso remoto al ERP y acceso a los manuales.
- **Cursos:** reservar un curso (dos videos), Google Meet, por qué llevar los cursos, cursos de capacitación.
- **Portales y apps:** Portal de Proveedores (nuevo y configuración), Portal del Empleado, Portal del Colaborador, Portal de Asistencia, Aprobaciones y Pedidos Móviles, Tareo en Línea, Calidad Móvil.
- **Nóminas:** UIT, RMV, jornal de construcción civil, AFP, feriados, CTS, gratificación y reintegros.
- **Otros:** instalar S10 Presupuestos, Lean Construction, control de costos, presentación del ERP y «¿Qué es Proptech?».

## Limitaciones

- **15 de los 38 videos no tienen narración** (solo música o texto en pantalla) y su transcripción está vacía: Nuevo Portal de Proveedores, Portal de Asistencia, Aprobaciones Móviles, Pedidos Móviles, los dos de Tareo en Línea, los dos de Calidad Móvil, Cómo llevar los cursos, Por qué llevar los cursos, Cursos de Capacitación, Lean Construction, Control de costos, la presentación del ERP y Proptech.
- Whisper confunde nombres: escribe «ese 10» o «sds» por **S10**, «rp» por **ERP** y «ndesk» por **AnyDesk**.
- Los videos de Nóminas llevan **cifras legales de un año concreto** (UIT 2022, 2024, 2025…). No valen para otro año.

**Confiabilidad:** oficial S10

**Fuentes**
- youtube.py (cabecera y constantes CANAL y PLAYLISTS)
- data/yt/videos.txt, data/yt/transcripciones/
