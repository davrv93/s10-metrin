# Fuentes de la base de conocimiento S10

## Incluidas (descarga libre, publicadas por su dueño)

| Origen | Qué hay | Cómo entró |
|---|---|---|
| documentacion.s10peru.com | 32 páginas públicas del portal (artículos de detracciones, anticipos, Lookahead, Calidad Móvil, prerrequisitos) | `rastrear --sin-login` |
| www.s10peru.com | 22 PDF enlazados desde páginas abiertas: sílabos por módulo, manual del sistema de tickets, guía de asistencia remota, políticas SGSST y SGSI | cola manual en `data/enlaces.jsonl` |
| repositorio.upn.edu.pe | Tesis: implementación de S10 ERP en la logística de una constructora | acceso abierto |
| transparencia.concytec.gob.pe | Informe técnico previo de evaluación del software S10 | documento público del Estado |

## Importadas a mano (29/30-09-2026)

| Documento | Páginas | Texto |
|---|---|---|
| Guía de usuario de S10 Presupuestos (Scribd) | 192 | OCR con Vision de macOS, ~53 500 palabras, 306 fragmentos |
| Manual de S10 Costos y Presupuestos (pdfcoffee) | 126 | texto nativo, 91 fragmentos |

## Pendientes

- **Manuales oficiales del portal** (58 de 90 páginas con muro de miembros): ya hay cuenta de miembro. Se bajan con `s10kb.py oficial` (credenciales en `.env`, nunca en git) y quedan en la rama `manuales-oficiales` de Cortex. Cuando entren, mandan sobre cualquier copia de terceros.
- **Tesis UPC Inocencio** (repositorioacademico.upc.edu.pe/handle/10757/688721): el enlace directo devuelve HTML; bajarla a mano desde el repositorio si interesa.

## Copias de terceros (para bajar a mano con tu cuenta)

Ninguna se descarga sin cuenta, pago o captcha, así que el script no las toca.
Bájalas desde tu navegador, déjalas en `entrada/` y corre
`s10kb.py importar && s10kb.py indexar`.

| Documento | Dónde | Qué pide |
|---|---|---|
| Guía de usuario de S10 Presupuestos (192 pág.) | https://www.scribd.com/document/630199402/Guia-de-Usuario-de-s10-Presupuestos | cuenta Scribd |
| Manual de S10 Presupuestos | https://es.scribd.com/document/688008331/Manual-de-S10-Presupuestos | cuenta Scribd |
| Guía de Presupuestos S10 | https://www.scribd.com/doc/90878564/Guia-de-Presupuestos-S10 | cuenta Scribd |
| Manual de S10 Costos y Presupuestos | https://es.scribd.com/document/233627638/Manual-de-s10-Costos-y-Presupuestos | cuenta Scribd |
| Manual del usuario S10 | https://www.scribd.com/doc/50474640/Manual-Del-Usario-S10 | cuenta Scribd |
| Guía costos y presupuestos con S10 | https://www.academia.edu/44809837/GUIA_COSTOS_Y_PRESUPUESTOS_CON_S10 | cuenta Academia.edu |
| Manual del S10 | https://www.slideshare.net/slideshow/manual-del-s10/68063768 | cuenta Slideshare |
| Manual de S10 | https://www.slideshare.net/slideshow/manual-de-s10-34998515/34998515 | cuenta Slideshare |
| Manual de S10 Costos y Presupuestos | https://pdfcoffee.com/manual-de-s10-costos-y-presupuestos-2-pdf-free.html | captcha |
