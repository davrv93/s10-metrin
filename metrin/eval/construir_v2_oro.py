#!/usr/bin/env python3
"""Construye metrin/eval/v2_oro.jsonl, el dataset de oro del benchmark V1 frente a V2 (docs/V2-RAG-PROCEDURAL.md §12).

Uso (desde la raíz de s10-conocimiento, con el .venv que trae PyYAML):

    .venv/bin/python3 metrin/eval/construir_v2_oro.py            # escribe el JSONL
    .venv/bin/python3 metrin/eval/construir_v2_oro.py --revisar  # solo comprueba, no escribe

Las PREGUNTAS están escritas a mano aquí abajo (sintéticas: no hay mensajes reales de usuarios). Las EXPECTATIVAS no
se escriben a mano: se derivan de kb/ en el momento de construir, con estas reglas de etiquetado:

1. PROCEDURE con procedimiento: `procedimiento_esperado` es un id vigente de kb/procedimientos/<modulo>/ (nunca de
   `_reserva/`). `pasos_esperados` = null (todos los de primer nivel) o la lista que la pregunta pide (p. ej. solo la
   parte del reporte SUNAT del kardex). `fotos_esperadas` = las fotos que el YAML pone en esos pasos (y sus subpasos),
   con su id = sha1(ruta_o_url)[:12]. `fragmentos_relevantes` = la fuente PRINCIPAL (la primera) de cada paso esperado,
   como par {id, manual} resuelto con la tabla `fuentes` del YAML.
2. Continuación («¿y luego?», «listo, ¿qué sigue?»): `continuacion: true` y `pasos_esperados: null`; el arnés espera
   los pasos que siguen al último que la MISMA versión entregó en el turno anterior.
3. NAVIGATION y CONFIGURATION: procedimiento de anclaje + los pasos que nombran el lugar o la opción; relevantes = TODAS
   las fuentes de esos pasos; `terminos_esperados` = nombres de escenario/opción que la respuesta debe decir. Cada
   término (o una de sus alternativas «a|b») se comprueba aquí contra el texto del paso y de sus fuentes.
4. TROUBLESHOOTING: procedimiento de anclaje + un `errores_frecuentes` concreto del YAML; relevantes = las fuentes de
   ese error; términos = palabras de la solución documentada (comprobadas contra el error y sus fuentes).
5. CONCEPT: `concepto_esperado` = id de kb/conceptos; relevantes = su `fuente`, desambiguada por el título del
   comentario de cada línea (los ids `web-…-sNNN` se repiten entre manuales).
6. AMBIGUOUS: tipo UNKNOWN (tema o palabra suelta, sin tarea). Si la palabra tiene entrada en el glosario,
   `tipos_aceptables` admite además CONCEPT y los relevantes son las fuentes del concepto.
7. Sin procedimiento en kb/procedimientos pero con pasos en una sección del manual: `procedimiento_esperado: null`,
   `sin_evidencia_esperada: false` y relevantes a mano («procedimiento implícito»); no cuenta para PAS.
8. Sin evidencia: `sin_evidencia_esperada: true`, sin procedimiento ni relevantes. Solo para temas cuya ausencia se
   comprobó con búsquedas en kb/fragmentos*.jsonl (anotadas en la justificación) o que kb/conceptos/SIN_FUENTE.md
   declara sin definición.
9. Fuga: ninguna pregunta puede coincidir (normalizada) con `preguntas`/`aliases` de un YAML ni con un ejemplo de
   entrenamiento de kb/catalogos (tipo_respuesta.yml, intenciones.yml): V2 recupera y clasifica con esos textos y el
   benchmark mediría memoria, no generalización. Los ejemplos obligatorios del megaprompt se conservan aunque se
   parezcan, y la justificación lo dice.
"""
import argparse
import glob
import hashlib
import json
import os
import re
import sys
import unicodedata

import yaml

RAIZ = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..'))
SALIDA = os.path.join(RAIZ, 'metrin', 'eval', 'v2_oro.jsonl')


def norm(s):
    s = unicodedata.normalize('NFKD', str(s).lower())
    s = ''.join(c for c in s if not unicodedata.combining(c))
    return ' '.join(re.findall(r'[a-z0-9]+', s))


def contiene_frase(texto, frase):
    return (' ' + norm(frase) + ' ') in (' ' + norm(texto) + ' ')


# ------------------------------------------------------------------------------------------------ KB
FRAG = {}        # (id, manual) -> fragmento
FRAG_ID = {}     # id -> [(id, manual)]
for ruta in sorted(glob.glob(os.path.join(RAIZ, 'kb', 'fragmentos*.jsonl'))):
    for linea in open(ruta, encoding='utf-8'):
        if not linea.strip():
            continue
        d = json.loads(linea)
        texto = d.get('texto', '') + '\n' + '\n'.join(p.get('texto', '') for p in d.get('pasos') or [])
        clave = (d['id'], d.get('manual') or '')
        if clave in FRAG:
            continue
        FRAG[clave] = {'titulo': d.get('titulo') or '', 'texto': texto, 'confianza': d.get('confianza') or ''}
        FRAG_ID.setdefault(d['id'], []).append(clave)

PROC = {}
for ruta in sorted(glob.glob(os.path.join(RAIZ, 'kb', 'procedimientos', '*', '*.yml'))):
    d = yaml.safe_load(open(ruta, encoding='utf-8'))
    d['_archivo'] = os.path.relpath(ruta, RAIZ)
    PROC[d['id']] = d

CONC = {}
for ruta in sorted(glob.glob(os.path.join(RAIZ, 'kb', 'conceptos', '*.yml'))):
    d = yaml.safe_load(open(ruta, encoding='utf-8'))
    titulos = {}
    for l in open(ruta, encoding='utf-8'):
        m = re.match(r'^\s*-\s*"?([^"#\s]+)"?\s+#\s*(.+?)\s*$', l)
        if m:
            titulos[m.group(1)] = m.group(2)
    d['_titulos'] = titulos
    d['_archivo'] = os.path.relpath(ruta, RAIZ)
    CONC[d['id']] = d

ENTRENAMIENTO = {}   # texto normalizado -> origen


def _ejemplos(ruta, raiz_clave):
    if not os.path.exists(ruta):
        return
    d = yaml.safe_load(open(ruta, encoding='utf-8')) or {}
    for nombre, clase in (d.get(raiz_clave) or {}).items():
        for e in (clase or {}).get('ejemplos') or []:
            ENTRENAMIENTO.setdefault(norm(e), '%s:%s' % (os.path.basename(ruta), nombre))


_ejemplos(os.path.join(RAIZ, 'kb', 'catalogos', 'tipo_respuesta.yml'), 'clases')
_ejemplos(os.path.join(RAIZ, 'kb', 'catalogos', 'intenciones.yml'), 'intenciones')
for ruta in glob.glob(os.path.join(RAIZ, 'kb', 'procedimientos', '**', '*.yml'), recursive=True):
    d = yaml.safe_load(open(ruta, encoding='utf-8'))
    for t in (d.get('preguntas') or []) + (d.get('aliases') or []):
        ENTRENAMIENTO.setdefault(norm(t), 'preguntas/aliases de ' + d['id'])

ERRORES, AVISOS = [], []


# ------------------------------------------------------------------------------------------------ derivación
def ref_proc(proc, fid):
    for f in proc['fuentes']:
        if f['id'] == fid and (fid, f['manual']) in FRAG:
            return {'id': fid, 'manual': f['manual']}
    claves = FRAG_ID.get(fid, [])
    if len(claves) == 1:
        return {'id': fid, 'manual': claves[0][1]}
    ERRORES.append('%s: fuente %s sin resolver' % (proc['id'], fid))
    return None


def ref_concepto(c, fid):
    claves = FRAG_ID.get(fid, [])
    if len(claves) == 1:
        return {'id': fid, 'manual': claves[0][1]}
    t = norm(c['_titulos'].get(fid, ''))
    for k in claves:
        if norm(FRAG[k]['titulo']) == t:
            return {'id': fid, 'manual': k[1]}
    for k in claves:
        if t.startswith(norm(k[1])):
            return {'id': fid, 'manual': k[1]}
    ERRORES.append('concepto %s: fuente %s ambigua sin título que la resuelva' % (c['id'], fid))
    return None


def pasos_de(proc, ns):
    return [p for p in proc['pasos'] if ns is None or p['n'] in ns]


def fotos_paso(p):
    out = []
    for f in p.get('fotos') or []:
        out.append(hashlib.sha1(f['ruta_o_url'].encode()).hexdigest()[:12])
        if f.get('id') and f['id'] != out[-1]:
            AVISOS.append('foto %s: id %s ≠ sha1(ruta)[:12] %s (se usa el sha1)' % (p['id'], f['id'], out[-1]))
    for s in p.get('sub') or []:
        out += fotos_paso(s)
    return out


def fuentes_paso(p):
    out = list(p.get('fuente') or [])
    for s in p.get('sub') or []:
        out += fuentes_paso(s)
    return out


def texto_paso(p):
    return '\n'.join([p.get('accion', ''), p.get('resultado') or ''] + [texto_paso(s) for s in p.get('sub') or []])


def unicos(refs):
    vistos, out = set(), []
    for r in refs:
        if r and (r['id'], r['manual']) not in vistos:
            vistos.add((r['id'], r['manual']))
            out.append(r)
    return out


def texto_refs(refs):
    return '\n'.join(FRAG[(r['id'], r['manual'])]['texto'] for r in refs if (r['id'], r['manual']) in FRAG)


def comprobar_terminos(cid, terminos, evidencia):
    for t in terminos:
        if not any(contiene_frase(evidencia, alt) for alt in t.split('|')):
            ERRORES.append('%s: el término «%s» no aparece en la evidencia' % (cid, t))


CASOS = []


def base(cid, categoria, turnos, tipo, variantes, **extra):
    c = {'id': cid, 'categoria': categoria, 'turnos': turnos, 'tipo_esperado': tipo}
    c.update(extra)
    c.setdefault('procedimiento_esperado', None)
    c.setdefault('concepto_esperado', None)
    c.setdefault('pasos_esperados', None)
    c.setdefault('fotos_esperadas', {})
    c.setdefault('fragmentos_relevantes', [])
    c.setdefault('sin_evidencia_esperada', False)
    c['sintetico'] = True
    c['variantes'] = list(variantes)
    if len(turnos) > 1 and 'conversacional' not in c['variantes']:
        c['variantes'].append('conversacional')
    CASOS.append(c)
    return c


def P(cid, turnos, proc_id, pasos=None, variantes=(), continuacion=False, categoria='PROCEDURE'):
    proc = PROC.get(proc_id)
    if not proc:
        ERRORES.append('%s: procedimiento %s no existe' % (cid, proc_id))
        return
    ps = pasos_de(proc, pasos)
    fotos = {str(p['n']): fotos_paso(p) for p in ps if fotos_paso(p)}
    rel = unicos([ref_proc(proc, p['fuente'][0]) for p in ps if p.get('fuente')])
    j = '%s v%s (%s): %s; fotos = las que el YAML pone en esos pasos (%d); relevantes = fuente principal de cada paso (%d).' % (
        proc_id, proc.get('version'), proc['_archivo'],
        'pasos %s' % pasos if pasos else 'los %d pasos de primer nivel' % len(proc['pasos']),
        sum(len(v) for v in fotos.values()), len(rel))
    extra = {}
    if continuacion:
        extra['continuacion'] = True
        j = 'Continuación: se esperan los pasos que siguen al último entregado en el turno anterior. ' + j
    base(cid, categoria, turnos, 'PROCEDURE', variantes, procedimiento_esperado=proc_id,
         pasos_esperados=pasos, fotos_esperadas=fotos, fragmentos_relevantes=rel, justificacion=j, **extra)


def ANCLA(cid, categoria, tipo, turnos, proc_id, pasos, terminos, variantes=()):
    """NAVIGATION / CONFIGURATION: los pasos que nombran el lugar o la opción."""
    proc = PROC.get(proc_id)
    if not proc:
        ERRORES.append('%s: procedimiento %s no existe' % (cid, proc_id))
        return
    ps = pasos_de(proc, pasos)
    rel = unicos([ref_proc(proc, f) for p in ps for f in fuentes_paso(p)])
    comprobar_terminos(cid, terminos, '\n'.join(texto_paso(p) for p in ps) + '\n' + texto_refs(rel))
    j = '%s: %s pasos %s de %s; relevantes = todas las fuentes de esos pasos; términos tomados del paso.' % (
        tipo, 'ruta en los' if tipo == 'NAVIGATION' else 'configuración en los', pasos, proc_id)
    base(cid, categoria, turnos, tipo, variantes, procedimiento_esperado=proc_id, fragmentos_relevantes=rel,
         terminos_esperados=terminos, justificacion=j)


def T(cid, turnos, proc_id, err, terminos, variantes=()):
    proc = PROC.get(proc_id)
    if not proc:
        ERRORES.append('%s: procedimiento %s no existe' % (cid, proc_id))
        return
    e = (proc.get('errores_frecuentes') or [])[err]
    rel = unicos([ref_proc(proc, f) for f in e['fuente']])
    comprobar_terminos(cid, terminos, e['sintoma'] + '\n' + e['solucion'] + '\n' + texto_refs(rel))
    j = 'TROUBLESHOOTING: errores_frecuentes[%d] de %s («%s»); relevantes = sus fuentes; términos de la solución.' % (
        err, proc_id, e['sintoma'])
    base(cid, 'TROUBLESHOOTING', turnos, 'TROUBLESHOOTING', variantes, procedimiento_esperado=proc_id,
         fragmentos_relevantes=rel, terminos_esperados=terminos, justificacion=j)


def C(cid, turnos, concepto, variantes=(), terminos=()):
    c = CONC.get(concepto)
    if not c:
        ERRORES.append('%s: concepto %s no existe' % (cid, concepto))
        return
    rel = unicos([ref_concepto(c, f) for f in c.get('fuente') or []])
    if terminos:
        comprobar_terminos(cid, terminos, c['definicion'] + '\n' + (c.get('en_s10') or '') + '\n' + texto_refs(rel))
    j = 'CONCEPT: %s (%s); relevantes = su `fuente`, desambiguada por el título del comentario.' % (concepto, c['_archivo'])
    extra = {'terminos_esperados': list(terminos)} if terminos else {}
    base(cid, 'CONCEPT', turnos, 'CONCEPT', variantes, concepto_esperado=concepto, fragmentos_relevantes=rel,
         justificacion=j, **extra)


def A(cid, turnos, concepto=None, variantes=()):
    extra = {'tipos_aceptables': ['UNKNOWN']}
    j = 'AMBIGUOUS: tema o palabra suelta sin tarea (frontera UNKNOWN de kb/catalogos/tipo_respuesta.yml): se espera ACLARAR_TAREA.'
    if concepto:
        c = CONC[concepto]
        extra['tipos_aceptables'] = ['UNKNOWN', 'CONCEPT']
        extra['fragmentos_relevantes'] = unicos([ref_concepto(c, f) for f in c.get('fuente') or []])
        j += ' El término tiene entrada en el glosario (%s): una definición con oferta de procedimiento también vale.' % concepto
    base(cid, 'AMBIGUOUS', turnos, 'UNKNOWN', variantes, justificacion=j, **extra)


def IMPLICITO(cid, turnos, refs, variantes=(), aceptables=None, nota=''):
    rel = []
    for fid, manual in refs:
        if (fid, manual) not in FRAG:
            ERRORES.append('%s: fragmento %s@%s no existe' % (cid, fid, manual))
        rel.append({'id': fid, 'manual': manual})
    extra = {'tipos_aceptables': aceptables} if aceptables else {}
    base(cid, 'PROCEDURE', turnos, 'PROCEDURE', variantes, fragmentos_relevantes=rel,
         justificacion='Procedimiento implícito: no hay YAML en kb/procedimientos; ' + nota, **extra)


def SE(cid, categoria, tipo, turnos, motivo, variantes=()):
    base(cid, categoria, turnos, tipo, list(variantes) + ['sin_evidencia'], sin_evidencia_esperada=True,
         justificacion='Sin evidencia: ' + motivo)


# ================================================================================================ CASOS
# --- PROCEDURE ----------------------------------------------------------------------------------
P('proc-001', ['como hago un metrado'], 'presupuestos.ingresar-metrados', variantes=['obligatorio', 'corta'])
P('proc-002', ['como registro metrado'], 'presupuestos.ingresar-metrados', variantes=['obligatorio', 'corta'])
P('proc-003', ['quiero que el metrado de cada partida se actualice solito desde mi archivo excel, se puede?'],
  'presupuestos.ingresar-metrados', pasos=[4, 5, 6], variantes=['coloquial'])
P('proc-004', ['kiero aser un presupuesto de una obra nueva en el s10, por donde empiezo'],
  'presupuestos.registrar-presupuesto-nuevo', variantes=['falta_ortografica', 'coloquial'])
P('proc-005', ['Buenas tardes. Soy asistente de costos en una constructora y me encargaron registrar en S10 el presupuesto '
               'de un puente que vamos a licitar. ¿Me puede indicar paso a paso cómo lo doy de alta, con el cliente, '
               'el plazo y la moneda?'], 'presupuestos.registrar-presupuesto-nuevo', variantes=['larga'])
P('proc-006', ['como pongo los titulos y debajo sus partidas en la hoja'], 'presupuestos.registrar-titulos-y-partidas',
  variantes=['coloquial'])
P('proc-007', ['la hoja no me muestra el total del presupuesto, como lo proceso?'], 'presupuestos.procesar-presupuesto',
  variantes=['coloquial'])
P('proc-008', ['pasos para la polinomica del presupuesto'], 'presupuestos.elaborar-formula-polinomica', variantes=['corta'])
P('proc-009', ['como armo los gastos generales detallados por rubros y conceptos'], 'presupuestos.calcular-gastos-generales')
P('proc-010', ['quiero reutilizar partidas de una obra anterior en mi presupuesto actual'],
  'presupuestos.copiar-partidas-de-otro-presupuesto')
P('proc-011', ['tengo el presupuesto armado en una hoja de calculo, se puede subir a s10?'],
  'presupuestos.importar-presupuesto-desde-excel', variantes=['coloquial'])
P('proc-012', ['necesito sacar el reporte estandar cliente de la hoja del presupuesto y mandarlo en excel'],
  'presupuestos.imprimir-reportes', pasos=[3, 4, 7, 8])
P('proc-013', ['pie del presupuesto con gastos generales, utilidad e igv: como lo diseño paso a paso'],
  'presupuestos.disenar-pie-de-presupuesto')
P('proc-014', ['la partida que necesito no existe en el catalogo, como la doy de alta con su analisis de precios'],
  'presupuestos.registrar-partida-en-el-catalogo')
P('proc-015', ['acabamos de ganar la licitacion, como creo la obra en el modulo de gerencia'],
  'gerencia-proyectos.registrar-proyecto', variantes=['coloquial'])
P('proc-016', ['como vinculo mi presupuesto venta y el meta con el proyecto'], 'gerencia-proyectos.asignar-presupuestos-al-proyecto')
P('proc-017', ['como armo las fases del proyecto y les reparto las partidas'], 'gerencia-proyectos.registrar-fases-wbs-y-asignar-partidas')
P('proc-018', ['vamos a tercerizar el tarrajeo, como registro ese subcontrato en s10'], 'gerencia-proyectos.registrar-subcontrato',
  variantes=['coloquial'])
P('proc-019', ['programacion de avance por periodos en gerencia de proyectos, como se hace'],
  'gerencia-proyectos.planificar-en-cronograma-por-periodos')
P('proc-020', ['tengo materiales en el almacen que todavia no estan en el sistema, como los ingreso por inventario'],
  'almacenes.registrar-ingreso-por-inventario')
P('proc-021', ['salida de cemento del almacen hacia una partida de la obra, como la registro'],
  'almacenes.registrar-egreso-a-partida-de-control')
P('proc-022', ['recibi los fierros de la orden de compra, como los registro en almacen'],
  'almacenes.registrar-ingreso-por-orden-de-compra', variantes=['coloquial'])
P('proc-023', ['me llego la factura de unas guias de remision que ya habia ingresado, que hago en almacenes'],
  'almacenes.regularizar-ingreso-por-compra', variantes=['coloquial'])
P('proc-024', ['ingreso IC con factura del proveedor paso a paso'], 'almacenes.registrar-ingreso-por-compra', variantes=['corta'])
P('proc-025', ['como defino un stock minimo para el petroleo del almacen'], 'almacenes.asignar-stock-minimo')
P('proc-026', ['quiero revisar el kardex de los fierros paso a paso'], 'almacenes.consultar-kardex')
P('proc-027', ['desde el modulo de compras y pedidos, como registro un pedido para comprar'], 'compras.registrar-pedido-de-compra')
P('proc-028', ['en gerencia de proyectos necesito pedir acero para la obra, como registro el pedido'],
  'gerencia-proyectos.registrar-pedido-manual-de-compra')
P('proc-029', ['como le doy el visto bueno a un pedido que hizo el almacenero'], 'compras.aprobar-pedido', variantes=['coloquial'])
P('proc-030', ['ya aprobaron el pedido, ahora como saco la orden de compra con proveedor y precios'],
  'compras.generar-orden-de-compra-desde-pedido')
P('proc-031', ['como mando a aprobar la OC y despues la apruebo'], 'compras.aprobar-orden-de-compra', variantes=['corta'])
P('proc-032', ['el cliente me pago en efectivo, como lo registro en caja'], 'administrativo.registrar-cobranza-en-efectivo',
  variantes=['coloquial'])
P('proc-033', ['orden de pago con cheque no negociable, como se hace'], 'administrativo.registrar-pago-con-cheque')
P('proc-034', ['necesito registrar un asiento de ajuste en contabilidad'], 'contabilidad.registrar-asiento-manual')
P('proc-035', ['cierre permanente de enero, como lo hago'], 'contabilidad.cierre-permanente-del-periodo', variantes=['corta'])
P('proc-036', ['tengo que declarar el registro de compras electronico, como genero el archivo'],
  'contabilidad.generar-libros-electronicos')
P('proc-037', ['le cobre de mas a un cliente en una factura, como emito la nota de credito'],
  'facturacion.registrar-nota-de-credito', variantes=['coloquial'])
P('proc-038', ['dar de alta un cliente nuevo con su cuenta de banco en facturacion'], 'facturacion.registrar-cliente')
P('proc-039', ['emiti mal una factura electronica, como la anulo con la comunicacion de baja'],
  'facturacion-electronica.comunicar-baja-de-comprobante')
P('proc-040', ['pasos para sacar la planilla semanal con la renta de quinta incluida'], 'nominas.calcular-la-nomina')
P('proc-041', ['entro un obrero nuevo a la obra, como lo registro en nominas con su dni y su cuenta'],
  'nominas.registrar-trabajador-como-socio-de-negocio', variantes=['coloquial'])
P('proc-042', ['tengo 80 obreros en un excel, como los subo todos de golpe al s10'],
  'nominas.importar-socios-de-negocio-desde-excel', variantes=['coloquial'])
P('proc-043', ['como saco las boletas de pago de la semana para que firmen'], 'nominas.imprimir-boletas-de-pago')
P('proc-044', ['tareo de la semana con la plantilla de excel, como lo registro'], 'nominas.registrar-tareo-estandar')
P('proc-045', ['quiero respaldar la base de datos antes de actualizar el sistema, como hago'],
  'instalacion.sacar-copia-de-seguridad-de-la-base-de-datos')
P('proc-046', ['instalacion del s10 en una laptop nueva paso a paso'], 'instalacion.instalar-s10-erp')
P('proc-047', ['como envio de una vez todas las facturas electronicas pendientes a la sunat'],
  'facturacion-electronica.enviar-comprobantes-a-sunat')
P('proc-048', ['prosupuesto nuebo en s10 komo lo registro'], 'presupuestos.registrar-presupuesto-nuevo',
  variantes=['falta_ortografica', 'corta'])
P('proc-049', ['como ingreso las cantidades de cada partida en la hoja del presupuesto'], 'presupuestos.ingresar-metrados')
IMPLICITO('proc-050', ['cómo modifico una partida'],
          [('img-2e9e7cd3e3d0-0000', 'Manual de Presupuestos'),
           ('web-https-documentacion-s10peru-com-manual-d-s025', 'Manual de Presupuestos')],
          variantes=['obligatorio', 'corta'], aceptables=['PROCEDURE', 'UNKNOWN'],
          nota='la captura del Catálogo de Partidas muestra «Modificar F3» (clic derecho) y la sección 2.2 Configuración (8) '
               'explica que modificar una subpartida de la hoja modifica la del catálogo salvo que sea «propia». No hay '
               'pasos documentados y «partida» puede ser la del catálogo o la de la hoja: pedir aclaración también vale.')
IMPLICITO('proc-051', ['como reemplazo un recurso por otro en todas las partidas del presupuesto'],
          [('img-62718bb4f55e-0000', 'Manual de Presupuestos')],
          nota='la captura del menú de la hoja del presupuesto nombra «reasignar recurso utilizado en el presupuesto» y '
               '«reasignar recurso del catálogo general»; la plantilla presupuestos.reasignar-recurso está en _reserva/ '
               '(no vigente).')
SE('proc-052', 'PROCEDURE', 'PROCEDURE', ['como registro un pago a un proveedor en criptomonedas en administrativo'],
   'búsqueda «criptomoneda|bitcoin» en kb/fragmentos*.jsonl: 0 resultados.')
SE('proc-053', 'PROCEDURE', 'PROCEDURE', ['como calculo la huella de carbono de la obra con s10'],
   'búsqueda «huella» en kb/fragmentos*.jsonl: 0 resultados.')
SE('proc-054', 'PROCEDURE', 'PROCEDURE', ['como mando mis reportes de s10 a power bi para hacer tableros'],
   'búsqueda «power bi» y «tablero(s) de/en excel|power» en kb/fragmentos*.jsonl: 0 resultados.')
# conversaciones que terminan en PROCEDURE
P('proc-055', ['quiero registrar un presupuesto de obra en s10, me guias?', '¿y luego?'], 'presupuestos.registrar-presupuesto-nuevo',
  continuacion=True, variantes=['continuacion'])
P('proc-056', ['como hago un ingreso por inventario al almacen', 'listo, ¿qué sigue?'],
  'almacenes.registrar-ingreso-por-inventario', continuacion=True, variantes=['continuacion'])
P('proc-057', ['la formula polinomica que viene a ser', '¿y cómo la hago?'], 'presupuestos.elaborar-formula-polinomica',
  variantes=['referencia_conversacional'])
P('proc-058', ['donde esta el kardex', '¿y cómo lo imprimo para la sunat?'], 'almacenes.consultar-kardex',
  pasos=[5, 6, 7, 8, 9], variantes=['referencia_conversacional'])
P('proc-059', ['que es un subcontrato', 'ya, y como registro uno?'], 'gerencia-proyectos.registrar-subcontrato',
  variantes=['referencia_conversacional'])
P('proc-060', ['tareo semanal con excel, como es', '¿y luego?', 'ok, ¿y después?'], 'nominas.registrar-tareo-estandar',
  continuacion=True, variantes=['continuacion'])

# --- CONCEPT ------------------------------------------------------------------------------------
C('conc-001', ['que es metrado'], 'metrado', variantes=['obligatorio', 'corta'])
C('conc-002', ['q es un apu'], 'analisis_de_precios_unitarios', variantes=['corta', 'abreviatura'])
C('conc-003', ['a que llaman partida de control en el control de la obra'], 'partida_de_control')
C('conc-004', ['el kardex que informacion me muestra'], 'kardex')
C('conc-005', ['explicame el indice unificado'], 'indice_unificado')
C('conc-006', ['que es eso de las fases o EDT del proyecto'], 'wbs', variantes=['coloquial'])
C('conc-007', ['que es un subcontrato'], 'subcontrato', variantes=['corta'])
C('conc-008', ['define rendimiento de mano de obra'], 'rendimiento')
C('conc-009', ['la compensacion por tiempo de servicios que es'], 'cts')
C('conc-010', ['que es el techo en un proyecto de s10'], 'techo')
C('conc-011', ['presupuesto venta, de que se trata'], 'presupuesto_venta')
C('conc-012', ['para que se usa el presupuesto meta en la obra'], 'presupuesto_meta')
C('conc-013', ['q es una cotizacion en compras'], 'cotizacion', variantes=['abreviatura'])
C('conc-014', ['que es un cuadro comparativo de proveedores'], 'cuadro_comparativo')
C('conc-015', ['que significa egreso de almacen'], 'egreso_de_almacen')
C('conc-016', ['que es una entrega a rendir'], 'entrega_a_rendir')
C('conc-017', ['caja chica que es en s10'], 'fondo_rotatorio', variantes=['sinonimo'])
C('conc-018', ['los gastos generales de una obra que incluyen'], 'gastos_generales')
C('conc-019', ['la guia de remision para que se usa'], 'guia_de_remision')
C('conc-020', ['monomio en la formula polinomica a que se refiere'], 'monomio')
C('conc-021', ['las ordenes de servicio para que son'], 'orden_de_servicio', variantes=['corta'])
C('conc-022', ['que es un parte de equipo'], 'parte_de_equipo')
C('conc-023', ['que es el plan contable'], 'plan_de_cuentas', variantes=['sinonimo'])
C('conc-024', ['planilla electronica que es'], 'planilla_electronica', variantes=['corta'])
C('conc-025', ['que son los resultados operativos de un proyecto'], 'resultados_operativos')
C('conc-026', ['que es un socio de negocio en s10'], 'socio_de_negocio')
C('conc-027', ['para que sirve un subpresupuesto'], 'subpresupuesto')
C('conc-028', ['que significa tarear a los obreros'], 'tareo')
C('conc-029', ['que es el tipo de nomina'], 'tipo_de_nomina')
C('conc-030', ['a que se refiere valorizar el avance en gerencia de proyectos'], 'valorizacion')
C('conc-031', ['que es un anticipo en una orden de compra'], 'anticipo')
C('conc-032', ['que es un recurso en s10, es lo mismo que insumo?'], 'recurso')
C('conc-033', ['cuadrilla que es'], 'cuadrilla', variantes=['corta'])
SE('conc-034', 'CONCEPT', 'CONCEPT', ['detracciones, eso que es'],
   'kb/conceptos/SIN_FUENTE.md: «Detracción: solo el catálogo de porcentajes… No hay definición».')
SE('conc-035', 'CONCEPT', 'CONCEPT', ['lookahead q significa'],
   'kb/conceptos/SIN_FUENTE.md: «Lookahead: solo “publicar un proyecto desde el LOOKAHEAD”… sin explicar qué es».')
SE('conc-036', 'CONCEPT', 'CONCEPT', ['centro de costos a que se refiere en s10'],
   'kb/conceptos/SIN_FUENTE.md: «Centro de costo: solo pasos de configuración…; no hay definición».')
C('conc-037', ['como registro un pedido de compra en el modulo de compras', 'oye, otra cosa: ¿qué es el kardex?'], 'kardex',
  variantes=['cambio_de_tema'])
C('conc-038', ['como proceso el presupuesto', 'gracias. ¿y qué es el pie de presupuesto?'], 'pie_de_presupuesto',
  variantes=['cambio_de_tema'])
C('conc-039', ['pasos para la formula polinomica', 'y eso del indice unificado que es'], 'indice_unificado',
  variantes=['referencia_conversacional'])

# --- NAVIGATION ---------------------------------------------------------------------------------
ANCLA('nav-001', 'NAVIGATION', 'NAVIGATION', ['donde veo metrados'], 'presupuestos.ingresar-metrados', [1],
      ['Hoja del Presupuesto'], variantes=['obligatorio', 'corta'])
ANCLA('nav-002', 'NAVIGATION', 'NAVIGATION', ['en que escenario se registran los asientos manuales'],
      'contabilidad.registrar-asiento-manual', [1], ['Asientos Manuales'])
ANCLA('nav-003', 'NAVIGATION', 'NAVIGATION', ['en que parte de contabilidad estan los estados financieros'],
      'contabilidad.emitir-estados-financieros', [1], ['Estados Financieros'])
ANCLA('nav-004', 'NAVIGATION', 'NAVIGATION', ['dnd estan los libros electronicos pa el ple'],
      'contabilidad.generar-libros-electronicos', [1], ['Libros Electronicos'], variantes=['abreviatura', 'corta'])
ANCLA('nav-005', 'NAVIGATION', 'NAVIGATION', ['en que pantalla se cierran los periodos contables'],
      'contabilidad.cierre-permanente-del-periodo', [1], ['Periodos Contables'])
ANCLA('nav-006', 'NAVIGATION', 'NAVIGATION', ['donde encuentro la opcion de comunicacion de baja'],
      'facturacion-electronica.comunicar-baja-de-comprobante', [2], ['Comunicacion de Baja'])
ANCLA('nav-007', 'NAVIGATION', 'NAVIGATION', ['en que escenario apruebo los pedidos de compra'], 'compras.aprobar-pedido', [2],
      ['Aprobacion'])
ANCLA('nav-008', 'NAVIGATION', 'NAVIGATION', ['donde esta el mantenimiento de usuarios'], 'presupuestos.registrar-usuarios', [1],
      ['Utilitarios'])
ANCLA('nav-009', 'NAVIGATION', 'NAVIGATION', ['en que menu esta lo de la copia de seguridad de la base'],
      'instalacion.sacar-copia-de-seguridad-de-la-base-de-datos', [1], ['Utilitarios'], variantes=['coloquial'])
ANCLA('nav-010', 'NAVIGATION', 'NAVIGATION', ['donde veo las ordenes de compra de un proyecto para aprobarlas'],
      'compras.aprobar-orden-de-compra', [3], ['Orden de Compra'])
ANCLA('nav-011', 'NAVIGATION', 'NAVIGATION', ['en que escenario se ponen los feriados del proyecto'],
      'gerencia-proyectos.configurar-calendario-del-proyecto', [2], ['Calendario'])
ANCLA('nav-012', 'NAVIGATION', 'NAVIGATION', ['donde esta el kardex en almacenes'], 'almacenes.consultar-kardex', [1, 3],
      ['Stock', 'Kardex'])
ANCLA('nav-013', 'NAVIGATION', 'NAVIGATION', ['por donde entro para registrar las salidas del almacen'],
      'almacenes.registrar-egreso-a-partida-de-control', [1], ['Egresos'], variantes=['coloquial'])
ANCLA('nav-014', 'NAVIGATION', 'NAVIGATION', ['donde estan los pedidos para generarles la orden de compra'],
      'compras.generar-orden-de-compra-desde-pedido', [1], ['Pedidos de Compra'])
ANCLA('nav-015', 'NAVIGATION', 'NAVIGATION', ['en que escenario se calcula la nomina'], 'nominas.calcular-la-nomina', [4],
      ['Calculo de Nominas'])
ANCLA('nav-016', 'NAVIGATION', 'NAVIGATION', ['donde preparo las boletas antes de imprimirlas'], 'nominas.imprimir-boletas-de-pago',
      [1], ['Preparacion de Boletas'])
ANCLA('nav-017', 'NAVIGATION', 'NAVIGATION', ['donde esta el escenario de gastos generales'], 'presupuestos.calcular-gastos-generales',
      [1], ['Gastos Generales'])
ANCLA('nav-018', 'NAVIGATION', 'NAVIGATION', ['en que escenario se hace la formula polinomica'],
      'presupuestos.elaborar-formula-polinomica', [1], ['Formula Polinomica'])
ANCLA('nav-019', 'NAVIGATION', 'NAVIGATION', ['donde estan las ordenes de pago en administrativo'],
      'administrativo.registrar-pago-con-cheque', [1], ['Ordenes de Pago'])
ANCLA('nav-020', 'NAVIGATION', 'NAVIGATION', ['donde registro el estado de cuenta que me manda el banco'],
      'administrativo.registrar-estado-bancario-manual', [1], ['Estado Bancario'])
ANCLA('nav-021', 'NAVIGATION', 'NAVIGATION', ['en nominas, en que menu esta el catalogo de socios de negocio'],
      'nominas.registrar-trabajador-como-socio-de-negocio', [1], ['Catalogos'])
ANCLA('nav-022', 'NAVIGATION', 'NAVIGATION', ['donde veo las facturas que todavia no se enviaron a sunat'],
      'facturacion-electronica.enviar-comprobantes-a-sunat', [4], ['Documentos pendientes de envio'])
ANCLA('nav-023', 'NAVIGATION', 'NAVIGATION', ['como registro un proyecto nuevo', 'y el presupuesto meta donde se lo asigno a la obra?'],
      'gerencia-proyectos.asignar-presupuestos-al-proyecto', [1, 2, 3], ['Datos Generales', 'Meta'],
      variantes=['referencia_conversacional'])
SE('nav-024', 'NAVIGATION', 'NAVIGATION', ['donde esta el asistente de inteligencia artificial dentro del s10'],
   'búsqueda «inteligencia artificial|chatgpt» en kb/fragmentos*.jsonl: 0 resultados.')

# --- TROUBLESHOOTING ----------------------------------------------------------------------------
SE('trb-001', 'TROUBLESHOOTING', 'TROUBLESHOOTING', ['no puedo guardar metrado'],
   'búsqueda «(grabar|guardar|graba) … metrado» en kb/fragmentos*.jsonl: 0 resultados; el único error documentado de '
   'presupuestos.ingresar-metrados es otro (Tomar metrado de celda Excel deshabilitada). No hay causa documentada.',
   variantes=['obligatorio', 'corta'])
T('trb-002', ['la opcion tomar metrado de celda excel me sale desactivada'], 'presupuestos.ingresar-metrados', 0,
  ['vinculada|vinculo|vincular'])
T('trb-003', ['no me deja registrar la cuadrilla de la partida nueva'], 'presupuestos.registrar-partida-en-el-catalogo', 0,
  ['rendimiento'])
T('trb-004', ['al mandar la OC a aprobar sale el mensaje no existe grupo de aprobadores'],
  'compras.registrar-aprobadores-de-orden-de-compra', 0, ['Centro de Compra'])
T('trb-005', ['soy aprobador pero el sistema no me deja aprobar la orden de compra'], 'compras.aprobar-orden-de-compra', 0,
  ['atributos|montos|rango'])
T('trb-006', ['no aparece ninguna orden de compra cuando quiero ingresar al almacen'],
  'almacenes.registrar-ingreso-por-orden-de-compra', 0, ['aprobadas'])
T('trb-007', ['el material no me aparece para hacer la salida del almacen'], 'almacenes.registrar-egreso-a-partida-de-control', 0,
  ['stock'])
T('trb-008', ['el boton enviar a sunat esta gris en mi factura'], 'facturacion-electronica.enviar-comprobantes-a-sunat', 0,
  ['talonario'], variantes=['coloquial'])
T('trb-009', ['sunat me rechazo la factura porque dice que la mande fuera de plazo'],
  'facturacion-electronica.enviar-comprobantes-a-sunat', 1, ['7 dias'])
T('trb-010', ['no me calcula la renta de quinta en la planilla'], 'nominas.calcular-la-nomina', 1, ['Renta de Quinta Categoria'])
T('trb-011', ['agregue un dato nuevo y la nomina no lo toma en cuenta'], 'nominas.calcular-la-nomina', 0,
  ['resumen de conceptos'])
T('trb-012', ['intento ver las boletas de pago y no se visualiza nada'], 'nominas.imprimir-boletas-de-pago', 0,
  ['preparadas|preparar|Preparacion de Boletas'])
T('trb-013', ['al ejecutar el pago con cheque me bota error'], 'administrativo.registrar-pago-con-cheque', 0, ['talonario'],
  variantes=['coloquial'])
T('trb-014', ['en el diario de caja solo me deja elegir documentos por cobrar'], 'administrativo.registrar-cobranza-en-efectivo', 0,
  ['dinero|efectivo'])
T('trb-015', ['el cliente no me sale para elegirlo cuando registro el presupuesto'], 'presupuestos.registrar-presupuesto-nuevo', 0,
  ['catalogo'])
T('trb-016', ['al procesar me pide un dato y no termina de procesar'], 'presupuestos.procesar-presupuesto', 0,
  ['metrados|aportes|precios'])
T('trb-017', ['la formula polinomica dice que dos monomios tienen el mismo simbolo'], 'presupuestos.elaborar-formula-polinomica', 1,
  ['Cambie|cambiar'])
T('trb-018', ['un monomio no llega al 0.05, que hago'], 'presupuestos.elaborar-formula-polinomica', 2,
  ['Agrupelo|agrupar|agrupe'])
T('trb-019', ['no puedo eliminar la guia de ingreso que hice ayer'], 'almacenes.registrar-ingreso-por-inventario', 1,
  ['regularizacion|atributos|atributo'])
T('trb-020', ['no puedo modificar ni borrar una guia de egreso'], 'almacenes.registrar-egreso-a-partida-de-control', 3,
  ['protegido|edicion'])
T('trb-021', ['mi compañero no puede ver el presupuesto que registre yo'], 'presupuestos.registrar-usuarios', 0,
  ['invitar'])
T('trb-022', ['en el cronograma por periodos no me deja escribir nada en la grilla'],
  'gerencia-proyectos.planificar-en-cronograma-por-periodos', 0, ['porcentaje|metrado'])
T('trb-023', ['cambie el presupuesto pero el proyecto sigue con los datos viejos'],
  'gerencia-proyectos.asignar-presupuestos-al-proyecto', 2, ['procesamiento tipo 3|procese'])
T('trb-024', ['al instalar el s10 me sale un mensaje de directplay'], 'instalacion.instalar-s10-erp', 0, ['DirectPlay'])
SE('trb-025', 'TROUBLESHOOTING', 'TROUBLESHOOTING', ['me sale licencia vencida al abrir el s10'],
   'búsqueda «licencia (vencida|caducada|expirada)» en kb/fragmentos*.jsonl: 0 resultados.')
SE('trb-026', 'TROUBLESHOOTING', 'TROUBLESHOOTING', ['el s10 se cierra solo cada vez que abro la hoja del presupuesto'],
   'búsqueda «se cierra solo|se cierra el (sistema|programa)» en kb/fragmentos*.jsonl: 0 resultados.')
T('trb-027', ['como hago un metrado', 'no me sale lo de tomar metrado de celda excel, esta en gris'],
  'presupuestos.ingresar-metrados', 0, ['vinculada|vinculo|vincular'], variantes=['no_me_sale'])
T('trb-028', ['tengo que aprobar una OC de cemento, como hago', 'me aparece el aviso: no existe grupo de aprobadores'],
  'compras.aprobar-orden-de-compra', 2, ['Centro de Compra'], variantes=['no_me_sale'])
T('trb-029', ['como calculo la planilla semanal de obreros', 'no me sale la renta de quinta'], 'nominas.calcular-la-nomina', 1,
  ['Renta de Quinta Categoria'], variantes=['no_me_sale'])
T('trb-030', ['quiero crear una partida propia en el catalogo', '¿y luego?', 'no me deja registrar la cuadrilla'],
  'presupuestos.registrar-partida-en-el-catalogo', 0, ['rendimiento'], variantes=['no_me_sale'])
T('trb-031', ['como envio facturas electronicas a sunat', 'no se me activa lo de enviar a sunat'],
  'facturacion-electronica.enviar-comprobantes-a-sunat', 0, ['talonario'], variantes=['no_me_sale'])

# --- CONFIGURATION ------------------------------------------------------------------------------
ANCLA('cfg-001', 'CONFIGURATION', 'CONFIGURATION', ['quiero que al procesar el presupuesto verifique los metrados, como lo configuro'],
      'presupuestos.procesar-presupuesto', [1], ['Verifica Metrados'])
ANCLA('cfg-002', 'CONFIGURATION', 'CONFIGURATION', ['como dejo configurado que el presupuesto use formula polinomica'],
      'presupuestos.registrar-presupuesto-nuevo', [12], ['Datos adicionales'])
ANCLA('cfg-003', 'CONFIGURATION', 'CONFIGURATION', ['cuantos decimales usa la hoja del presupuesto y donde lo cambio'],
      'presupuestos.registrar-presupuesto-nuevo', [12], ['decimales'])
ANCLA('cfg-004', 'CONFIGURATION', 'CONFIGURATION', ['con que letra empieza la serie del talonario de facturas electronicas'],
      'facturacion-electronica.configurar-emision-electronica', [3, 4, 5], ['F'])
ANCLA('cfg-005', 'CONFIGURATION', 'CONFIGURATION', ['que tengo que configurar en la ficha del cliente para mandarle facturas electronicas'],
      'facturacion-electronica.configurar-emision-electronica', [7, 8], ['contacto'])
ANCLA('cfg-006', 'CONFIGURATION', 'CONFIGURATION', ['quiero limitar hasta cuanto puede autorizar cada aprobador de pagos'],
      'administrativo.definir-aprobadores-de-pagos', [6, 7, 8], ['Monto maximo'])
ANCLA('cfg-007', 'CONFIGURATION', 'CONFIGURATION', ['como configuro los aprobadores de ordenes de compra segun el monto'],
      'compras.registrar-aprobadores-de-orden-de-compra', [4, 5, 6], ['Monto minimo'])
ANCLA('cfg-008', 'CONFIGURATION', 'CONFIGURATION', ['necesito crear el centro de compras y asignarle los proyectos'],
      'compras.registrar-centro-de-compra', [3, 4, 9], ['Centro de Compra', 'Proyectos'])
ANCLA('cfg-009', 'CONFIGURATION', 'CONFIGURATION', ['la obra trabaja de 7 a 5, donde ajusto ese horario en el proyecto'],
      'gerencia-proyectos.configurar-calendario-del-proyecto', [5], ['horario'], variantes=['coloquial'])
ANCLA('cfg-010', 'CONFIGURATION', 'CONFIGURATION', ['quiero que mi cronograma sea mensual y no semanal, que configuro'],
      'gerencia-proyectos.definir-periodos', [2], ['Escala de Periodos'])
ANCLA('cfg-011', 'CONFIGURATION', 'CONFIGURATION', ['quiero que un asistente solo entre a la hoja del presupuesto y no toque los catalogos'],
      'presupuestos.registrar-usuarios', [8, 9], ['Acceso a los escenarios', 'Acceso a los catalogos'])
ANCLA('cfg-012', 'CONFIGURATION', 'CONFIGURATION', ['que caracteristicas debe tener el tipo de documento para las entregas a rendir'],
      'facturacion.registrar-entrega-a-rendir', [1, 2, 3], ['Tipo de Documento'])
ANCLA('cfg-013', 'CONFIGURATION', 'CONFIGURATION', ['despues de importar el plan contable, que propiedades le pongo a cada cuenta'],
      'contabilidad.crear-plan-de-cuentas', [10, 11], ['Tipo de Cuenta', 'Naturaleza'])
ANCLA('cfg-014', 'CONFIGURATION', 'CONFIGURATION', ['en el pie de presupuesto quiero que el impuesto no entre en la formula polinomica'],
      'presupuestos.disenar-pie-de-presupuesto', [9], ['impuesto'])
ANCLA('cfg-015', 'CONFIGURATION', 'CONFIGURATION', ['quiero que al pedir materiales sea obligatorio poner la fecha de entrega'],
      'compras.registrar-pedido-de-compra', [2, 3], ['Fecha de entrega obligatoria'])
ANCLA('cfg-016', 'CONFIGURATION', 'CONFIGURATION', ['es la primera vez que voy a tarear, que formato le asigno a la empresa'],
      'nominas.registrar-tareo-estandar', [1, 3], ['Formatos por Empresa|Asignar formato'])
ANCLA('cfg-017', 'CONFIGURATION', 'CONFIGURATION', ['no tengo servidor, como instalo el s10 con su base de datos en mi equipo'],
      'instalacion.instalar-s10-erp', [4, 9], ['SERVIDOR'])
ANCLA('cfg-018', 'CONFIGURATION', 'CONFIGURATION', ['necesito definir los niveles y rubros de mi balance en contabilidad'],
      'contabilidad.configurar-informe-contable', [1, 5], ['Informes Contables'])
ANCLA('cfg-019', 'CONFIGURATION', 'CONFIGURATION', ['qué es el pie de presupuesto', '¿y cómo lo dejo configurado con costo directo y total?'],
      'presupuestos.disenar-pie-de-presupuesto', [7, 8], ['nDirecto|P_T|Macro'], variantes=['referencia_conversacional'])
SE('cfg-020', 'CONFIGURATION', 'CONFIGURATION', ['como pongo el s10 en modo oscuro'],
   'búsqueda «modo oscuro» en kb/fragmentos*.jsonl: 0 resultados.', variantes=['corta'])
SE('cfg-021', 'CONFIGURATION', 'CONFIGURATION', ['como configuro el s10 para que mis empleados marquen teletrabajo'],
   'búsqueda «teletrabajo» en kb/fragmentos*.jsonl: 0 resultados.')

# --- AMBIGUOUS ----------------------------------------------------------------------------------
A('amb-001', ['metrado s10'], concepto='metrado', variantes=['obligatorio', 'corta'])
A('amb-002', ['kardex'], concepto='kardex', variantes=['corta'])
A('amb-003', ['ordenes de compra s10'], concepto='orden_de_compra', variantes=['corta'])
A('amb-004', ['tengo una duda con almacenes'])
A('amb-005', ['lo del otro dia no me salio'], variantes=['coloquial'])
A('amb-006', ['ayuda con presupuestos'], variantes=['corta'])
A('amb-007', ['partidas'], concepto='partida', variantes=['corta'])
A('amb-008', ['nomina'], variantes=['corta'])
A('amb-009', ['sunat'], variantes=['corta'])
A('amb-010', ['pedidos s10'], concepto='pedido', variantes=['corta'])
A('amb-011', ['s10 presupuestos excel'], variantes=['corta'])
A('amb-012', ['subcontratos'], concepto='subcontrato', variantes=['corta'])
A('amb-013', ['reporte'], variantes=['corta'])
A('amb-014', ['valorizacion'], concepto='valorizacion', variantes=['corta'])
A('amb-015', ['necesito ayuda urgente con el sistema'])
A('amb-016', ['como hago eso'], variantes=['corta', 'referencia_sin_contexto'])
A('amb-017', ['¿y luego?'], variantes=['corta', 'referencia_sin_contexto'])
A('amb-018', ['formula'], variantes=['corta'])
A('amb-019', ['cuadrilla obra s10'], concepto='cuadrilla', variantes=['corta'])


# ================================================================================================ comprobaciones
def fuga():
    for c in CASOS:
        for t in c['turnos']:
            n = norm(t)
            if n in ENTRENAMIENTO:
                if 'obligatorio' in c['variantes'] or len(n.split()) <= 2:
                    nota = ' Coincide con «%s» de %s (se conserva: %s).' % (
                        t, ENTRENAMIENTO[n], 'ejemplo obligatorio del megaprompt' if 'obligatorio' in c['variantes']
                        else 'palabra suelta inevitable')
                    if nota not in c['justificacion']:
                        c['justificacion'] += nota
                else:
                    ERRORES.append('%s: «%s» está en el entrenamiento (%s)' % (c['id'], t, ENTRENAMIENTO[n]))
            else:
                tok = set(n.split())
                for e, origen in ENTRENAMIENTO.items():
                    te = set(e.split())
                    jac = len(tok & te) / len(tok | te)
                    if len(tok) >= 4 and jac >= 0.75:
                        AVISOS.append('%s: «%s» se parece a «%s» (%s)' % (c['id'], t, e, origen))
                    elif 'obligatorio' in c['variantes'] and jac >= 0.6:
                        nota = ' Se parece a «%s» de %s (se conserva: ejemplo obligatorio del megaprompt).' % (e, origen)
                        if nota not in c['justificacion']:
                            c['justificacion'] += nota


# Marketing: la misma regla que EsMarketing de metrin/internal/v2/conocimiento/base.go. El quality gate de V2 rechaza
# citar estas fuentes, así que no pueden ser «relevantes» de un caso (lo serían solo por estar en un YAML).
MARKETING = re.compile(r'^(Optimiza 360|s10peru\.com \(público\)|Mes: |S10 ERP, El Software|CURSOS S10|'
                       r'Soporte Técnico \||Servicio de Implantación \||Sidebar|Home |panel \d|Portada |CONTACTO|Join Us|'
                       r'unete|topbar|varios widgets|Página de ejemplo|Acceso de miembro|Perfil$|Registro$|'
                       r'Formulario Feedback|Temporal$|Sin categoría$)')


def es_marketing(manual, confianza):
    return confianza == 'propio' or bool(MARKETING.match((manual or '').strip()))


def sin_marketing():
    for c in CASOS:
        for r in c.get('fragmentos_relevantes') or []:
            f = FRAG.get((r['id'], r['manual']))
            if f is not None and es_marketing(r['manual'], f['confianza']):
                ERRORES.append('%s: el relevante %s@%s es de una fuente de marketing (el quality gate no deja citarla; '
                               'quítela de la `fuente` del YAML del que sale)' % (c['id'], r['id'], r['manual']))


def resumen():
    from collections import Counter
    cat = Counter(c['categoria'] for c in CASOS)
    conv = sum(1 for c in CASOS if len(c['turnos']) > 1)
    print('casos: %d · %s · conversaciones %d · sin evidencia %d · con procedimiento %d (%d distintos) · conceptos %d' % (
        len(CASOS), ' · '.join('%s %d' % kv for kv in sorted(cat.items())), conv,
        sum(c['sin_evidencia_esperada'] for c in CASOS),
        sum(1 for c in CASOS if c['procedimiento_esperado']),
        len({c['procedimiento_esperado'] for c in CASOS if c['procedimiento_esperado']}),
        len({c['concepto_esperado'] for c in CASOS if c['concepto_esperado']})))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--revisar', action='store_true', help='solo comprobar, no escribir')
    a = ap.parse_args()
    ids = [c['id'] for c in CASOS]
    if len(ids) != len(set(ids)):
        ERRORES.append('ids repetidos')
    fuga()
    sin_marketing()
    resumen()
    for w in AVISOS:
        print('aviso:', w)
    for e in ERRORES:
        print('ERROR:', e)
    if ERRORES:
        return 1
    if not a.revisar:
        orden = ['id', 'categoria', 'turnos', 'tipo_esperado', 'tipos_aceptables', 'procedimiento_esperado',
                 'concepto_esperado', 'pasos_esperados', 'continuacion', 'fotos_esperadas', 'fragmentos_relevantes',
                 'terminos_esperados', 'sin_evidencia_esperada', 'sintetico', 'variantes', 'justificacion']
        with open(SALIDA, 'w', encoding='utf-8') as fh:
            for c in CASOS:
                fh.write(json.dumps({k: c[k] for k in orden if k in c}, ensure_ascii=False) + '\n')
        print('escrito:', os.path.relpath(SALIDA, RAIZ))
    return 0


if __name__ == '__main__':
    sys.exit(main())
