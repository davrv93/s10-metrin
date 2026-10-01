# Plan de Reinserción de Imágenes y Actualización KB - s10-conocimiento

**Fecha:** 2026-10-01  
**Objetivo:** Descargar imágenes faltantes de páginas ya rastreadas, reindexar KB, actualizar Cortex y Metrín.  
**Idempotente:** Cada paso verifica qué ya está hecho y solo procesa lo pendiente.

---

## Estado Actual (2026-10-01)

| Métrica | Valor |
|---------|-------|
| URLs descubiertas (`data/urls.json`) | 90 |
| Páginas rastreadas (`data/paginas.jsonl`, estado 200) | 137 |
| Páginas con HTML guardado (`data/html/`) | 137 |
| Páginas con imágenes descargadas (`data/imagenes.jsonl`) | 59 |
| Total imágenes descargadas | 3,672 |
| Páginas SIN imágenes | 78 |
| **OCR completado** (`data/imagenes_ocr.jsonl`) | **3,672 / 3,672** ✅ |
| **Fragmentos KB web/PDF** (`kb/fragmentos.jsonl`) | **5,865** |
| **Fragmentos Cortex** (`kb/fragmentos_cortex.jsonl`) | **1,355** |
| **Fragmentos FAQ** (`kb/fragmentos_faq.jsonl`) | **808** |
| **Total fragmentos KB** | **8,028** |
| **Última actualización Metrín** (`kb/.actualizado`) | **2026-10-01T10:39:44** |

### Manuales PRINCIPALES con imágenes YA descargadas ✓
- `manual-de-gerencia-proyectos` — 136 imágenes
- `manual-de-presupuestos` — 422 imágenes
- `manual-de-almacenes`, `manual-de-aprobaciones`, `manual-de-calidad-movil`
- `manual-de-compras`, `manual-de-contabilidad`, `manual-de-facturacion*`
- `manual-de-lean`, `manual-de-nominas`, `manual-instalacion-s10-erp`
- `manual-portal-proveedor`, `nuevo-manual-portal-proveedor`, `manual-de-tareo-movil`

### Páginas CON contenido PERO sin imágenes (prioridad alta)
Estas páginas tienen HTML con imágenes `s10peru.com` pero no se bajaron:

| Página | Título | Imágenes en HTML |
|--------|--------|------------------|
| `manual-aprobaciones` | Acceso a Manual de Aprobaciones Móviles | 1 (logo) |
| `manual-del-colaborador` | Acceso a Manual del Colaborador | 1 (logo) |
| `manual-integracion-crm-sperant` | Manual de Integración con CRM Sperant | 1 (logo) |
| `manual-muros` | Acceso a Manual de Muros | 1 (logo) |
| `manual-pedidos-moviles` | Acceso a Manual de Pedidos Móviles | 1 (logo) |
| `manual-tablero` | Acceso a Manual de Tablero Web | 1 (logo) |
| `manual-tareo-linea` | Manual de Tareo en Línea | 1 (logo) |
| `video-calidad-movil` | Videos de Calidad Móvil | 1 (logo) |
| `video-empleado` | Videos de Portal del Empleado | 1 (logo) |
| `video-facturacion-electronica` | Videos de Portal de Facturación Electrónica | 1 (logo) |
| `video-instala-presupuesto` | Video Instalación Presupuesto | 1 (logo) |
| `video-lean` | Videos Tutoriales de Lean | 1 (logo) |
| `video-nominas` | Videos Tutoriales de Nóminas | 1 (logo) |
| `video-presupuestos` | Videos Tutoriales de Presupuestos | 1 (logo) |
| `video-proveedor` | Videos de Portal del Proveedor | 1 (logo) |
| `video-tareo-movil` | Videos de Tareo Móvil | 1 (logo) |

> **Nota:** Muchas son páginas de "Acceso a Manual" / "Videos de..." — son índices/landing con solo el logo. El contenido real está en los manuales principales (ya tienen imágenes).

### Páginas SIN imágenes (prioridad baja - índices, archivos, categorías)
- Páginas de archivo mensual (2019/03, 2019/05, etc.)
- Categorías y tags
- `optimiza360.pe` y `www.s10peru.com` (páginas públicas)
- Páginas de login/registro/membership

---

## Requisitos Previos

### Credenciales de miembro S10 (necesarias para `medios` en páginas con muro)
El `.env` actual tiene credenciales demo (`usuario_demo` / `usuariodemo`) que **no funcionan**.

**Opciones:**
1. **Credenciales reales** de Arturo/S10 en `.env` → `S10_USUARIO` / `S10_CLAVE`
2. **Cookies del navegador** → exportar con extensión (formato Netscape) y usar `--cookies cookies.txt`

> Sin credenciales válidas, `medios` falla al intentar re-validar páginas con muro. Las páginas que YA tienen HTML sin muro se procesarían igual, pero el CLI exige login para el paso `medios`.

---

## Plan de Pasos (Ejecutar en orden)

### Paso 1: Descargar imágenes faltantes (`medios`)
```bash
cd /Users/david.roncal/Downloads/PjgFactSalud_completo/s10-conocimiento
.venv/bin/python s10kb.py medios --limite 0
```
- **Idempotente:** Salta imágenes ya en `data/imagenes.jsonl` (por URL)
- **Duración estimada:** ~2-5 min (3,672 ya hechas, ~100-200 nuevas)
- **Requisito:** Login válido O cookies de navegador
- **Verificación:** `wc -l data/imagenes.jsonl` debe aumentar

> Si no hay credenciales: usar `--cookies cookies-s10.txt` (ya existe en `data/`) o exportar nuevas.

### Paso 2: OCR de imágenes nuevas (`ocr_imagenes`)
```bash
.venv/bin/python -c "from s10kb import ocr_imagenes; ocr_imagenes()"
```
- **Idempotente:** Solo OCR imágenes sin entrada en `data/imagenes_ocr.jsonl` o con mtime cambiado
- **Requisito:** `tesseract` instalado (`brew install tesseract tesseract-lang`)
- **Verificación:** `wc -l data/imagenes_ocr.jsonl` aumenta

### Paso 3: Reindexar KB completa (`indexar`)
```bash
.venv/bin/python s10kb.py indexar
```
- **Idempotente:** Relee todo `data/html/`, `data/pdf/`, `data/imagenes_ocr.jsonl` y regenera `kb/fragmentos.jsonl` + `kb/documentos.json`
- **Incluye:** Fragmentos por sección (HTML), pasos con capturas, OCR de imágenes, PDFs
- **Verificación:** `wc -l kb/fragmentos.jsonl` y revisar `kb/documentos.json`

### Paso 4: Subir a Cortex (rama `manuales-oficiales`)
```bash
.venv/bin/python oficial_a_cortex.py --cargar
```
- **Idempotente:** Crea/actualiza nodos en rama `manuales-oficiales`; sin tablero → borrador en `data/cortex-borradores/`
- **Verificación:** Revisar `data/cortex-borradores/manuales-oficiales.json` o tablero Cortex

### Paso 5: Actualizar KB de Metrín desde Cortex
```bash
.venv/bin/python cortex_a_kb.py
```
- **Idempotente:** Regenera `kb/fragmentos_cortex.jsonl` y `kb/fragmentos_faq.jsonl`; toca `kb/.actualizado`
- **Verificación:** `cat kb/.actualizado` muestra timestamp nuevo; Metrín recarga solo

### Paso 6: Validar en Metrín (chat)
```bash
# Probar que el chatbot responde con capturas
curl -X POST http://localhost:4760/chat \
  -H "Content-Type: application/json" \
  -d '{"pregunta": "¿Cómo creo un presupuesto en S10?", "contexto": "manual-de-presupuestos"}'
```
- Verificar que la respuesta incluye `![Captura del manual](fotos/...)`

---

## Checklist de Progreso (marcar al completar)

- [x] **Paso 1** - `medios` completado — **0 imágenes nuevas** (todas 3,672 ya descargadas; cookies `data/cookies-s10.txt` funcionan)
- [x] **Paso 2** - OCR imágenes completado — **3,672/3,672** (`data/imagenes_ocr.jsonl`)
- [x] **Paso 3** - `indexar` completado — **5,865 fragmentos** en `kb/fragmentos.jsonl` (incluye OCR)
- [x] **Paso 4** - `oficial_a_cortex.py --cargar` completado — borrador en `data/cortex-borradores/manuales-oficiales-actualizado.json` (1,347 manuales)
- [x] **Paso 5** - `cortex_a_kb.py` completado — 1,355 frag. Cortex + 808 FAQ; `kb/.actualizado` tocado (2026-10-01T10:39:44)
- [ ] **Paso 6** - Validación en Metrín pendiente

---

## Comandos de Verificación Rápida

```bash
# Imágenes totales
wc -l data/imagenes.jsonl

# Páginas con imágenes vs total
.venv/bin/python3 -c "
import json
pags = set()
with open('data/paginas.jsonl') as f:
    for l in f:
        d=json.loads(l)
        if d.get('estado')==200 and d.get('archivo') and not d.get('muro_de_miembros'):
            pags.add(d['pagina'])
imgs = set()
with open('data/imagenes.jsonl') as f:
    for l in f:
        d=json.loads(l)
        imgs.add(d['pagina'])
print(f'Con imágenes: {len(imgs)} / {len(pags)} total')
print(f'Faltantes: {len(pags)-len(imgs)}')
"

# Fragmentos KB
wc -l kb/fragmentos.jsonl
wc -l kb/fragmentos_cortex.jsonl
wc -l kb/fragmentos_faq.jsonl

# Timestamp última actualización Metrín
cat kb/.actualizado
```

---

## Log de Ejecución (2026-10-01)

| Paso | Comando | Resultado | Tiempo |
|------|---------|-----------|--------|
| 1 | `s10kb.py medios --cookies data/cookies-s10.txt` | 0 nuevas (3,672 existentes), 137 saltadas | ~1 min |
| 2 | `ocr_imagenes()` (background) | 3,672/3,672 completadas | ~25 min |
| 3 | `s10kb.py indexar` (1ª) | 237 docs, 2,532 frags | ~2 min |
| 4 | `oficial_a_cortex.py --cargar` | 1,347 manuales → borrador Cortex | ~30 seg |
| 5 | `cortex_a_kb.py` (1ª) | 1,355 Cortex + 808 FAQ; .actualizado tocado | ~10 seg |
| 3b | `s10kb.py indexar` (2ª, post-OCR) | **3,570 docs, 5,865 frags** | ~3 min |
| 5b | `cortex_a_kb.py` (2ª) | 1,355 Cortex + 808 FAQ; .actualizado tocado | ~10 seg |
| 6 | Validación Metrín | Pendiente | — |

> **Nota:** El OCR (paso 2) corrió en background (`bgp_0f6fe870d001unosGbD8hK7qxw`). Tras terminar, se re-ejecutó `indexar` (paso 3b) para incluir el OCR en la KB, y luego `cortex_a_kb.py` + tocar `.actualizado` (paso 5b).

---

## Si se Acaban Créditos / Sesión

1. El plan se retoma en el **último paso no marcado** ✓
2. Todos los pasos son **idempotentes**: volver a correr no duplica ni rompe
3. Los datos intermedios (`data/html/`, `data/imagenes/`, `data/imagenes.jsonl`, `kb/`) persisten en disco
4. Solo necesitas credenciales válidas para `medios` (paso 1)

---

## Notas Adicionales

- **`manual-de-gerencia-proyectos`:** El README menciona que "aún cae al troceado genérico porque el HTML guardado no contiene su cuerpo". Verificar si el HTML actual (892K chars, 823 imgs, 262K texto principal) ya está bien. Si no, re-rastrear con sesión válida.
- **Credenciales:** Pedir a Arturo usuario/clave reales de `documentacion.s10peru.com` (cuenta de miembro Simple Membership).
- **Cookies alternativas:** `data/cookies-s10.txt` y `data/cookies.txt` existen; probar con `--cookies data/cookies-s10.txt`.