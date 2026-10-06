#!/usr/bin/env python3
"""Valida las plantillas de procedimiento de Metrín (kb/procedimientos/<modulo>/<id>.yml).

Qué comprueba (el esquema completo está en kb/procedimientos/ESQUEMA.md):

  1. Esquema: campos obligatorios, tipos, valores cerrados (modulo, nivel), sin campos desconocidos,
     6–12 preguntas, `id` único, con prefijo = módulo e igual a la ruta <modulo>/<id>.yml, `n` consecutivos
     desde 1 (también en `sub`).
  2. Fuentes: cada paso, subpaso, prerrequisito, verificación y error frecuente lleva `fuente` con 1+ ids que
     EXISTEN en kb/fragmentos*.jsonl. Como hay ids repetidos entre manuales (ver ESQUEMA.md, «ids ambiguos»),
     cada id citado debe figurar en la tabla `fuentes` del procedimiento, y es esa tabla —el par (id, manual)—
     la que fija a qué fragmento se refiere. La tabla debe cuadrar con el fragmento real (seccion, pagina) y
     no puede tener entradas que nadie cite. No se admiten fuentes de marketing (Optimiza 360, s10peru.com).
  3. Menús no inventados: todo término entre ** ** aparece literalmente —sin distinguir mayúsculas ni
     acentos, con espacios normalizados— en el texto (o el OCR, o los pasos) de al menos uno de los
     fragmentos que cita ESE elemento.
  4. Soporte léxico: cada `accion` (y cada `resultado`, prerrequisito, verificación y error) comparte un
     mínimo de palabras de contenido con sus fuentes. Por debajo del umbral se marca DUDOSO: no es error
     (salvo con --estricto), es una lista para revisar a mano. Umbral calibrado: ver ESQUEMA.md.
  5. `relacionados` apuntan a ids de procedimientos existentes.
  6. Pasos: `id` = <procedimiento>#<n> (subpasos <n>.<m>), único; `pagina` = la del primer fragmento fuente.
  7. Fotos: cada foto existe en un fragmento fuente de SU paso y la fuente la asocia a ese paso (está en
     `pasos[k].fotos` de una sección citada —y entonces `caption` es ese `pasos[k].texto`— o es un fragmento
     imagen citado por el paso); `id` = sha1(ruta_o_url)[:12]; `ocr` es un prefijo del OCR de esa imagen;
     ninguna foto se repite en dos pasos del mismo procedimiento. Se informa el % de pasos con foto y las
     «fotos dudosas»: las que comparten poco vocabulario con el paso (caption u OCR frente a la acción).
  8. `aliases` (2–10) y `entidades` {modulo, pantallas, objetos}: cada pantalla y cada objeto aparece en
     alguno de los fragmentos de la tabla `fuentes`.

Uso:
  python3 herramientas/validar_procedimientos.py                 # valida y resume
  python3 herramientas/validar_procedimientos.py --indice        # además escribe kb/procedimientos/INDICE.json
  python3 herramientas/validar_procedimientos.py --dudosos       # lista cada elemento dudoso con su puntaje
  python3 herramientas/validar_procedimientos.py --calibrar      # muestra la calibración del umbral léxico
  python3 herramientas/validar_procedimientos.py --estricto      # los dudosos también hacen fallar
  python3 herramientas/validar_procedimientos.py --modulo compras  # solo un módulo (repetible)
  python3 herramientas/validar_procedimientos.py --reserva        # incluye kb/procedimientos/_reserva/ (no van al índice)

Las carpetas que empiezan por «_» (p. ej. `_reserva/`) se ignoran salvo con --reserva: guardan plantillas ya
validadas que no están activas. Para activar una, muévala a <modulo>/ y vuelva a validar.

Sale con 0 si todo está bien, 1 si hay errores (o dudosos con --estricto), 2 si no puede ejecutarse.
Solo biblioteca estándar + PyYAML (está en .venv/ del repo). Sin PyYAML usa un lector mínimo que entiende
exactamente el subconjunto de YAML que usan estos archivos (ver ESQUEMA.md, «Formato»).
"""
import argparse
import glob
import hashlib
import json
import os
import random
import re
import sys
import unicodedata

RAIZ = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

MODULOS = ['presupuestos', 'gerencia-proyectos', 'almacenes', 'compras', 'nominas', 'contabilidad', 'facturacion',
           'administrativo', 'facturacion-electronica', 'portal-proveedor', 'tareo-movil', 'calidad-movil',
           'instalacion']
NIVELES = ['basico', 'intermedio', 'avanzado']
NOMBRE_MODULO = {'presupuestos': 'Presupuestos', 'gerencia-proyectos': 'Gerencia de Proyectos', 'almacenes': 'Almacenes',
                 'compras': 'Compras', 'nominas': 'Nóminas', 'contabilidad': 'Contabilidad', 'facturacion': 'Facturación',
                 'administrativo': 'Administrativo', 'facturacion-electronica': 'Facturación Electrónica',
                 'portal-proveedor': 'Portal del Proveedor', 'tareo-movil': 'Tareo Móvil',
                 'calidad-movil': 'Calidad Móvil', 'instalacion': 'Instalación'}
CAMPOS = ['id', 'version', 'modulo', 'titulo', 'objetivo', 'nivel', 'preguntas', 'aliases', 'entidades',
          'prerrequisitos', 'pasos', 'verificacion', 'errores_frecuentes', 'relacionados', 'fuentes', 'revision']
CAMPOS_PASO_OBL = ['n', 'id', 'accion', 'fuente', 'pagina']
CAMPOS_PASO = CAMPOS_PASO_OBL + ['donde', 'resultado', 'fotos', 'sub']
CAMPOS_FOTO_OBL = ['id', 'ruta_o_url', 'pagina']
CAMPOS_FOTO = CAMPOS_FOTO_OBL + ['caption', 'ocr']
MARKETING = re.compile(r'^(Optimiza 360|s10peru\.com \(público\))')
ID_RE = re.compile(r'^([a-z0-9]+(?:-[a-z0-9]+)*)\.([a-z0-9]+(?:-[a-z0-9]+)*)$')

# Soporte léxico (calibrado con --calibrar; ver ESQUEMA.md)
UMBRAL_MIN = 2       # palabras de contenido compartidas como mínimo...
UMBRAL_RATIO = 0.6   # ...y al menos esta fracción de las palabras de contenido del texto propio
LARGO_RAIZ = 5       # se compara por prefijo (registre/registrar/registro -> «regis»)

STOP = set('''a al algo algun alguna algunas alguno algunos ante antes asi aun bajo bien cada como con contra cual
cuales cuando de del desde donde dos e el ella ellas ello ellos en entre era es esa esas ese eso esos esta estas
este esto estos fue ha hace hacia han hasta hay la las le les lo los mas me mediante mismo muy ni no nos o otra
otras otro otros para pero poco por porque pues que quien se segun ser si sin sino sobre solo son su sus tal tambien
tan tanto te tiene tienen todo todos tras tu un una unas uno unos usted y ya ademas luego despues primero ahora
cual cuya cuyo debe deben dentro puede pueden sera seran esta estan ese caso vez
haga clic doble derecho izquierdo boton botones use utilice pulse presione elija seleccione ventana opcion opciones
menu sistema muestra mostrara aparece pantalla escenario modulo s10'''.split())


def normalizar(s):
    s = unicodedata.normalize('NFKD', str(s).lower())
    s = ''.join(c for c in s if not unicodedata.combining(c))
    s = s.replace('“', '"').replace('”', '"').replace('’', "'").replace('‘', "'")
    return re.sub(r'\s+', ' ', s).strip()


def raices(s):
    out = set()
    for t in re.findall(r'[a-z0-9ñ]+', normalizar(s)):
        if len(t) < 4 or t in STOP or t.isdigit():
            continue
        out.add(t[:LARGO_RAIZ])
    return out


# ---------------------------------------------------------------- lector YAML mínimo (respaldo sin PyYAML)
class ErrorYaml(Exception):
    pass


def _escalar(v):
    v = v.strip()
    if v == '':
        return None
    if v.startswith('"'):
        return json.loads(v)
    if v.startswith('['):
        return [_escalar(x) for x in _partir(v[1:-1])] if v[1:-1].strip() else []
    if v.startswith('{'):
        d = {}
        for par in _partir(v[1:-1]):
            k, _, x = par.partition(':')
            d[k.strip()] = _escalar(x)
        return d
    if v in ('null', '~'):
        return None
    if v in ('true', 'false'):
        return v == 'true'
    if re.fullmatch(r'-?\d+', v):
        return int(v)
    if v.startswith("'") or v.startswith('|') or v.startswith('>') or v.startswith('&') or v.startswith('*'):
        raise ErrorYaml('sintaxis no soportada sin PyYAML: %s' % v[:30])
    return v


def _partir(s):
    out, cur, prof, comillas = [], '', 0, False
    i = 0
    while i < len(s):
        c = s[i]
        if comillas:
            cur += c
            if c == '\\':
                cur += s[i + 1]
                i += 1
            elif c == '"':
                comillas = False
        elif c == '"':
            comillas = True
            cur += c
        elif c in '[{':
            prof += 1
            cur += c
        elif c in ']}':
            prof -= 1
            cur += c
        elif c == ',' and prof == 0:
            out.append(cur)
            cur = ''
        else:
            cur += c
        i += 1
    if cur.strip():
        out.append(cur)
    return out


def _clave_valor(t):
    m = re.match(r'^([A-Za-z_][A-Za-z0-9_]*):(?:\s+(.*))?$', t)
    if not m:
        raise ErrorYaml('línea no reconocida: %s' % t[:60])
    return m.group(1), (m.group(2) or '')


def yaml_minimo(texto):
    lineas = []
    for l in texto.split('\n'):
        if not l.strip() or l.lstrip().startswith('#'):
            continue
        ind = len(l) - len(l.lstrip(' '))
        lineas.append((ind, l.strip()))
    pos = [0]

    def bloque(ind):
        if lineas[pos[0]][1].startswith('- '):
            return lista(ind)
        return mapa(ind)

    def mapa(ind, d=None):
        d = {} if d is None else d
        while pos[0] < len(lineas):
            i, t = lineas[pos[0]]
            if i < ind or t.startswith('- '):
                break
            if i > ind:
                raise ErrorYaml('sangría inesperada: %s' % t[:60])
            k, v = _clave_valor(t)
            pos[0] += 1
            if v == '':
                if pos[0] < len(lineas) and lineas[pos[0]][0] > ind:
                    d[k] = bloque(lineas[pos[0]][0])
                elif pos[0] < len(lineas) and lineas[pos[0]][0] == ind and lineas[pos[0]][1].startswith('- '):
                    d[k] = lista(ind)
                else:
                    d[k] = None
            else:
                d[k] = _escalar(v)
        return d

    def lista(ind):
        out = []
        while pos[0] < len(lineas):
            i, t = lineas[pos[0]]
            if i != ind or not t.startswith('- '):
                break
            resto = t[2:]
            if re.match(r'^[A-Za-z_][A-Za-z0-9_]*:(\s|$)', resto):
                # elemento mapa: la primera clave va en la línea del guion
                lineas[pos[0]] = (ind + 2, resto)
                out.append(mapa(ind + 2))
            else:
                pos[0] += 1
                out.append(_escalar(resto))
        return out

    return mapa(0)


try:
    import yaml as _yaml

    def cargar_yaml(texto):
        return _yaml.safe_load(texto)
    LECTOR = 'PyYAML %s' % _yaml.__version__
except ImportError:  # pragma: no cover
    cargar_yaml = yaml_minimo
    LECTOR = 'lector mínimo (sin PyYAML)'


# ---------------------------------------------------------------- base de fragmentos
class Base:
    def __init__(self, dir_kb):
        self.por_id = {}
        self.img_por_ruta = {}
        self.n = 0
        for fn in sorted(glob.glob(os.path.join(dir_kb, 'fragmentos*.jsonl'))):
            archivo = os.path.basename(fn)
            with open(fn, encoding='utf-8') as fh:
                for l in fh:
                    if not l.strip():
                        continue
                    f = json.loads(l)
                    f['_archivo'] = archivo
                    self.por_id.setdefault(f['id'], []).append(f)
                    self.n += 1
                    if f.get('tipo') == 'imagen':
                        for k in ('imagen', 'url_imagen'):
                            if f.get(k):
                                self.img_por_ruta[f[k]] = f

    @staticmethod
    def manual(f):
        return f.get('manual') or f.get('modulo') or f['_archivo'].replace('.jsonl', '')

    @staticmethod
    def seccion(f):
        return f.get('seccion') or f.get('titulo') or ''

    @staticmethod
    def texto(f):
        if '_norm' not in f:
            partes = [f.get('texto', ''), f.get('seccion') or '', f.get('titulo') or '']
            partes += [p.get('texto', '') for p in f.get('pasos') or []]
            f['_norm'] = normalizar(' '.join(partes))
        return f['_norm']

    @staticmethod
    def raices(f):
        if '_raices' not in f:
            f['_raices'] = raices(Base.texto(f))
        return f['_raices']

    @staticmethod
    def imagenes(f):
        out = set()
        for k in ('imagen', 'url_imagen', 'captura'):
            if f.get(k):
                out.add(f[k])
        for p in f.get('pasos') or []:
            out.update(p.get('fotos') or [])
        return out


# ---------------------------------------------------------------- validación
class Informe:
    def __init__(self):
        self.errores = []
        self.dudosos = []  # (archivo, lugar, puntaje, compartidas, total, texto)
        self.fotos_dudosas = []  # (archivo, lugar, ruta, compartidas, total, texto del paso)

    def error(self, archivo, msg):
        self.errores.append('%s: %s' % (archivo, msg))


def soporte(texto, frags):
    propias = raices(re.sub(r'\*\*', '', texto))
    fuente = set()
    for f in frags:
        fuente |= Base.raices(f)
    comp = propias & fuente
    total = len(propias)
    ratio = (len(comp) / total) if total else 1.0
    ok = total == 0 or (len(comp) >= min(UMBRAL_MIN, total) and ratio >= UMBRAL_RATIO)
    return ok, ratio, len(comp), total


def validar_proc(ruta, d, base, inf, todos_ids, dir_proc):
    rel = os.path.relpath(ruta, dir_proc)
    if not isinstance(d, dict):
        inf.error(rel, 'el archivo no es un mapa YAML')
        return
    for c in CAMPOS:
        if c not in d:
            inf.error(rel, 'falta el campo obligatorio `%s`' % c)
    for c in d:
        if c not in CAMPOS:
            inf.error(rel, 'campo desconocido `%s`' % c)
    pid = d.get('id')
    mod = d.get('modulo')
    if mod not in MODULOS:
        inf.error(rel, 'modulo inválido: %r' % mod)
    m = ID_RE.match(str(pid or ''))
    if not m:
        inf.error(rel, 'id inválido (kebab-case con prefijo de módulo): %r' % pid)
    elif m.group(1) != mod:
        inf.error(rel, 'el prefijo del id (%s) no es el módulo (%s)' % (m.group(1), mod))
    esperado = '%s/%s.yml' % (mod, pid)
    if rel.replace(os.sep, '/') != esperado:
        inf.error(rel, 'la ruta no coincide con el id: se esperaba %s' % esperado)
    if not isinstance(d.get('version'), int) or d.get('version') < 1:
        inf.error(rel, '`version` debe ser un entero >= 1')
    for c in ('titulo', 'objetivo'):
        if not isinstance(d.get(c), str) or not d.get(c).strip():
            inf.error(rel, '`%s` vacío' % c)
    if d.get('nivel') not in NIVELES:
        inf.error(rel, 'nivel inválido: %r' % d.get('nivel'))
    ali = d.get('aliases')
    if not isinstance(ali, list) or not (2 <= len(ali) <= 10) or not all(isinstance(x, str) and x.strip() for x in ali) \
            or len(set(normalizar(x) for x in ali)) != len(ali):
        inf.error(rel, '`aliases` debe ser una lista de 2 a 10 textos distintos')
    pre = d.get('preguntas')
    if not isinstance(pre, list) or not (6 <= len(pre) <= 12):
        inf.error(rel, '`preguntas` debe tener entre 6 y 12 elementos (tiene %s)' % (len(pre) if isinstance(pre, list) else '?'))
    elif len(set(normalizar(p) for p in pre)) != len(pre) or not all(isinstance(p, str) and p.strip() for p in pre):
        inf.error(rel, '`preguntas` con elementos vacíos o repetidos')

    # tabla de fuentes: (id) -> fragmento resuelto
    tabla = {}
    usados = set()
    fuentes = d.get('fuentes')
    if not isinstance(fuentes, list) or not fuentes:
        inf.error(rel, '`fuentes` debe ser una lista no vacía')
        fuentes = []
    for ent in fuentes:
        if not isinstance(ent, dict) or set(ent) != {'id', 'manual', 'seccion', 'pagina'}:
            inf.error(rel, 'entrada de `fuentes` mal formada (id, manual, seccion, pagina): %r' % (ent,))
            continue
        cand = [f for f in base.por_id.get(str(ent['id']), []) if Base.manual(f) == ent['manual']]
        if not cand:
            inf.error(rel, 'fuentes: no existe el fragmento %s en «%s»' % (ent['id'], ent['manual']))
            continue
        if len(cand) > 1:
            inf.error(rel, 'fuentes: (%s, %s) es ambiguo en la base' % (ent['id'], ent['manual']))
            continue
        f = cand[0]
        if ent['id'] in tabla:
            inf.error(rel, 'fuentes: id repetido %s' % ent['id'])
        tabla[ent['id']] = f
        if Base.seccion(f) != ent['seccion']:
            inf.error(rel, 'fuentes: la sección de %s es «%s», no «%s»' % (ent['id'], Base.seccion(f), ent['seccion']))
        if f.get('pagina') != ent['pagina']:
            inf.error(rel, 'fuentes: la página de %s es %r, no %r' % (ent['id'], f.get('pagina'), ent['pagina']))
        if MARKETING.match(Base.manual(f)) or MARKETING.match(str(f.get('titulo', ''))):
            inf.error(rel, 'fuentes: %s es una página de marketing, no un manual' % ent['id'])

    ent = d.get('entidades')
    if not isinstance(ent, dict) or set(ent) != {'modulo', 'pantallas', 'objetos'}:
        inf.error(rel, '`entidades` debe ser {modulo, pantallas, objetos}')
    else:
        if ent['modulo'] != NOMBRE_MODULO.get(mod):
            inf.error(rel, 'entidades.modulo debe ser «%s»' % NOMBRE_MODULO.get(mod))
        for c in ('pantallas', 'objetos'):
            if not isinstance(ent[c], list) or not all(isinstance(x, str) and x.strip() for x in ent[c]):
                inf.error(rel, 'entidades.%s debe ser una lista de textos' % c)
                continue
            for x in ent[c]:
                if not any(normalizar(x) in Base.texto(f) for f in tabla.values()):
                    inf.error(rel, 'entidades.%s: «%s» no aparece en ninguna de sus fuentes' % (c, x))
        if not ent.get('objetos'):
            inf.error(rel, 'entidades.objetos no puede estar vacío')

    def resolver(lista, lugar):
        if not isinstance(lista, list) or not lista:
            inf.error(rel, '%s: falta `fuente` (lista de ids)' % lugar)
            return []
        out = []
        for fid in lista:
            fid = str(fid)
            if fid not in base.por_id:
                inf.error(rel, '%s: el fragmento %s no existe en kb/fragmentos*.jsonl' % (lugar, fid))
                continue
            if fid not in tabla:
                inf.error(rel, '%s: %s no está en la tabla `fuentes` (necesaria para resolverlo)' % (lugar, fid))
                continue
            usados.add(fid)
            out.append(tabla[fid])
        return out

    def negritas(texto, frags, lugar):
        for t in re.findall(r'\*\*(.+?)\*\*', texto or ''):
            nt = normalizar(t)
            if not any(nt in Base.texto(f) for f in frags):
                inf.error(rel, '%s: **%s** no aparece literal en sus fuentes' % (lugar, t))
        if (texto or '').count('**') % 2:
            inf.error(rel, '%s: negritas desparejadas' % lugar)

    def lexico(texto, frags, lugar):
        if not frags or not texto:
            return
        ok, ratio, comp, total = soporte(texto, frags)
        if not ok:
            inf.dudosos.append((rel, lugar, ratio, comp, total, texto))

    def items(clave, campos_txt):
        lst = d.get(clave)
        if not isinstance(lst, list):
            inf.error(rel, '`%s` debe ser una lista (puede estar vacía)' % clave)
            return
        for i, it in enumerate(lst, 1):
            lugar = '%s[%d]' % (clave, i)
            if not isinstance(it, dict):
                inf.error(rel, '%s mal formado' % lugar)
                continue
            esperadas = set(campos_txt) | {'fuente'}
            if set(it) != esperadas:
                inf.error(rel, '%s: campos %s, se esperaban %s' % (lugar, sorted(it), sorted(esperadas)))
            frs = resolver(it.get('fuente'), lugar)
            for c in campos_txt:
                if not isinstance(it.get(c), str) or not it.get(c).strip():
                    inf.error(rel, '%s: `%s` vacío' % (lugar, c))
                    continue
                negritas(it[c], frs, lugar + '.' + c)
                lexico(it[c], frs, lugar + '.' + c)

    items('prerrequisitos', ['texto'])
    items('verificacion', ['texto'])
    items('errores_frecuentes', ['sintoma', 'solucion'])

    cont = {'pasos': 0, 'con_foto': 0, 'fotos': 0}
    ids_paso = set()
    rutas_foto = {}

    def foto(fo, frs, p, lugar):
        if not isinstance(fo, dict):
            inf.error(rel, '%s: foto mal formada' % lugar)
            return
        for c in CAMPOS_FOTO_OBL:
            if c not in fo:
                inf.error(rel, '%s: foto sin `%s`' % (lugar, c))
        for c in fo:
            if c not in CAMPOS_FOTO:
                inf.error(rel, '%s: campo de foto desconocido `%s`' % (lugar, c))
        ruta = str(fo.get('ruta_o_url') or '')
        if fo.get('id') != hashlib.sha1(ruta.encode('utf-8')).hexdigest()[:12]:
            inf.error(rel, '%s: el id de la foto debe ser sha1(ruta_o_url)[:12]' % lugar)
        if ruta in rutas_foto:
            inf.error(rel, '%s: la foto %s ya está en %s' % (lugar, ruta, rutas_foto[ruta]))
        rutas_foto[ruta] = lugar
        if not any(ruta in Base.imagenes(f) for f in frs):
            inf.error(rel, '%s: la foto %s no existe en los fragmentos fuente de este paso' % (lugar, ruta))
            return
        # asociación foto <-> paso según la fuente
        directos = [f for f in frs if ruta in (f.get('imagen'), f.get('url_imagen'), f.get('captura'))]
        textos = [pk.get('texto', '') for f in frs for pk in (f.get('pasos') or []) if ruta in (pk.get('fotos') or [])]
        paginas = {f.get('pagina') for f in directos} | {f.get('pagina') for f in frs
                                                         if any(ruta in (pk.get('fotos') or []) for pk in (f.get('pasos') or []))}
        if not directos and not textos:
            inf.error(rel, '%s: la fuente no asocia la foto %s a este paso' % (lugar, ruta))
        if fo.get('pagina') not in paginas:
            inf.error(rel, '%s: la página de la foto debería ser una de %s' % (lugar, sorted(paginas, key=str)))
        if 'caption' in fo:
            if directos:
                inf.error(rel, '%s: la foto se asocia por la imagen citada (su OCR): no lleva caption' % lugar)
            elif not any(normalizar(fo['caption']) == normalizar(t) for t in textos):
                inf.error(rel, '%s: el caption de la foto no es el texto que la fuente asocia a ella' % lugar)
        if 'ocr' in fo:
            img = base.img_por_ruta.get(ruta)
            ref = normalizar(img['texto'].split('OCR:', 1)[-1]) if img else ''
            if not img or not ref.startswith(normalizar(str(fo['ocr']).rstrip('…'))):
                inf.error(rel, '%s: el ocr de la foto no coincide con el de la imagen en la base' % lugar)
        # foto dudosa: poco vocabulario en común entre el paso y el texto que acompaña a la foto
        # se compara con el OCR completo de la base (el campo `ocr` va recortado a 400 caracteres)
        img_b = base.img_por_ruta.get(ruta)
        ocr_full = img_b['texto'].split('OCR:', 1)[-1] if img_b else (fo.get('ocr') or '')
        asoc = (fo.get('caption') or '') + ' ' + ocr_full
        propias = raices(re.sub(r'\*\*', '', (p.get('accion') or '') + ' ' + (p.get('resultado') or '')))
        comp = propias & raices(asoc)
        if len(comp) < min(2, len(propias)):
            inf.fotos_dudosas.append((rel, lugar, ruta, len(comp), len(propias), p.get('accion') or ''))

    def pasos(lst, prefijo, num_pref):
        if not isinstance(lst, list) or not lst:
            inf.error(rel, '%s: debe ser una lista no vacía' % (prefijo or 'pasos'))
            return
        for i, p in enumerate(lst, 1):
            lugar = '%spaso %s' % (prefijo, i)
            num = '%s%d' % (num_pref, i)
            if not isinstance(p, dict):
                inf.error(rel, '%s mal formado' % lugar)
                continue
            cont['pasos'] += 1
            for c in CAMPOS_PASO_OBL:
                if c not in p:
                    inf.error(rel, '%s: falta `%s`' % (lugar, c))
            for c in p:
                if c not in CAMPOS_PASO:
                    inf.error(rel, '%s: campo desconocido `%s`' % (lugar, c))
            if p.get('n') != i:
                inf.error(rel, '%s: `n` es %r, debería ser %d (consecutivos desde 1)' % (lugar, p.get('n'), i))
            esperado = '%s#%s' % (pid, num)
            if p.get('id') != esperado:
                inf.error(rel, '%s: el id del paso debe ser %s' % (lugar, esperado))
            if p.get('id') in ids_paso:
                inf.error(rel, '%s: id de paso repetido %s' % (lugar, p.get('id')))
            ids_paso.add(p.get('id'))
            frs = resolver(p.get('fuente'), lugar)
            if frs and 'pagina' in p and p['pagina'] != frs[0].get('pagina'):
                inf.error(rel, '%s: `pagina` debe ser la del primer fragmento fuente (%r)' % (lugar, frs[0].get('pagina')))
            if not isinstance(p.get('accion'), str) or not p.get('accion').strip():
                inf.error(rel, '%s: `accion` vacía' % lugar)
            else:
                negritas(p['accion'], frs, lugar + '.accion')
                lexico(p['accion'], frs, lugar + '.accion')
            for c in ('donde', 'resultado'):
                if c in p:
                    if not isinstance(p[c], str) or not p[c].strip():
                        inf.error(rel, '%s: `%s` vacío' % (lugar, c))
                    else:
                        negritas(p[c], frs, lugar + '.' + c)
                        if c == 'resultado':
                            lexico(p[c], frs, lugar + '.' + c)
            if 'fotos' in p:
                if not isinstance(p['fotos'], list) or not p['fotos']:
                    inf.error(rel, '%s: `fotos` debe ser una lista no vacía (u omitirse)' % lugar)
                else:
                    cont['con_foto'] += 1
                    for j, fo in enumerate(p['fotos'], 1):
                        cont['fotos'] += 1
                        foto(fo, frs, p, '%s.foto %d' % (lugar, j))
            if 'sub' in p:
                pasos(p['sub'], lugar + '.', num + '.')

    pasos(d.get('pasos'), '', '')

    relac = d.get('relacionados')
    if not isinstance(relac, list):
        inf.error(rel, '`relacionados` debe ser una lista')
    else:
        for r in relac:
            if r == pid:
                inf.error(rel, 'relacionados: se cita a sí mismo')
            elif r not in todos_ids:
                inf.error(rel, 'relacionados: no existe el procedimiento %s' % r)

    for fid in tabla:
        if fid not in usados:
            inf.error(rel, 'fuentes: %s figura en la tabla pero ningún elemento lo cita' % fid)

    rev = d.get('revision')
    if not isinstance(rev, dict) or set(rev) != {'generado', 'por', 'revisado_por_humano'} \
            or not isinstance(rev.get('revisado_por_humano'), bool):
        inf.error(rel, '`revision` debe ser {generado, por, revisado_por_humano: bool}')
    return cont


def calibrar(base, procs):
    """Compara el soporte de cada acción con su fuente real frente a una fuente al azar del mismo manual."""
    rnd = random.Random(10)
    por_manual = {}
    for lst in base.por_id.values():
        for f in lst:
            if f.get('pasos') or f.get('tipo') == 'imagen':
                por_manual.setdefault(Base.manual(f), []).append(f)
    reales, azar = [], []

    def recorrer(lst, tabla):
        for p in lst or []:
            frs = [tabla[x] for x in p.get('fuente', []) if x in tabla]
            if frs and p.get('accion'):
                reales.append(soporte(p['accion'], frs))
                man = Base.manual(frs[0])
                otros = [f for f in por_manual.get(man, []) if f not in frs]
                if otros:
                    azar.append(soporte(p['accion'], rnd.sample(otros, min(len(frs), len(otros)))))
            recorrer(p.get('sub'), tabla)

    for _, d in procs:
        tabla = {}
        for ent in d.get('fuentes') or []:
            c = [f for f in base.por_id.get(str(ent.get('id')), []) if Base.manual(f) == ent.get('manual')]
            if c:
                tabla[ent['id']] = c[0]
        recorrer(d.get('pasos'), tabla)
    print('Calibración del soporte léxico (umbral: >= %d raíces compartidas y >= %.0f %%)' % (UMBRAL_MIN, UMBRAL_RATIO * 100))
    for nombre, xs in (('fuente real', reales), ('fuente al azar del mismo manual', azar)):
        if not xs:
            continue
        ok = sum(1 for x in xs if x[0])
        rs = sorted(x[1] for x in xs)
        print('  %-34s n=%4d  pasan=%5.1f %%  mediana=%.2f  p10=%.2f  p90=%.2f' % (
            nombre, len(xs), 100.0 * ok / len(xs), rs[len(rs) // 2], rs[len(rs) // 10], rs[(9 * len(rs)) // 10]))
    for umbral in (0.3, 0.4, 0.5, 0.6, 0.7):
        r = sum(1 for x in reales if x[1] >= umbral and x[2] >= min(UMBRAL_MIN, x[3])) / max(1, len(reales))
        a = sum(1 for x in azar if x[1] >= umbral and x[2] >= min(UMBRAL_MIN, x[3])) / max(1, len(azar))
        print('  ratio >= %.1f: pasan %5.1f %% de las reales y %5.1f %% de las al azar' % (umbral, 100 * r, 100 * a))


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n\n')[0])
    ap.add_argument('--dir', default=os.path.join(RAIZ, 'kb', 'procedimientos'))
    ap.add_argument('--kb', default=os.path.join(RAIZ, 'kb'))
    ap.add_argument('--indice', action='store_true', help='escribe INDICE.json si no hay errores')
    ap.add_argument('--dudosos', action='store_true', help='lista los elementos con poco soporte léxico')
    ap.add_argument('--calibrar', action='store_true', help='muestra la calibración del umbral léxico')
    ap.add_argument('--estricto', action='store_true', help='los dudosos también hacen fallar')
    ap.add_argument('--modulo', action='append', help='valida solo este módulo (repetible); el índice exige validar todos')
    ap.add_argument('--reserva', action='store_true', help='valida también las plantillas de _reserva/ (no entran al índice)')
    a = ap.parse_args()

    if not os.path.isdir(a.dir):
        print('No existe %s' % a.dir)
        return 2
    base = Base(a.kb)
    inf = Informe()
    rutas = sorted(glob.glob(os.path.join(a.dir, '**', '*.yml'), recursive=True))
    procs = []
    reserva = set()
    for r in rutas:
        partes = os.path.relpath(r, a.dir).split(os.sep)
        oculta = [x for x in partes[:-1] if x.startswith('_')]
        if oculta and (not a.reserva or partes[0] != '_reserva' or len(oculta) > 1):
            continue
        try:
            with open(r, encoding='utf-8') as fh:
                procs.append((r, cargar_yaml(fh.read())))
            if oculta:
                reserva.add(r)
        except Exception as e:  # noqa: BLE001
            inf.error(os.path.relpath(r, a.dir), 'no se pudo leer el YAML: %s' % e)
    ids = {}
    for r, d in procs:
        if isinstance(d, dict) and d.get('id'):
            if d['id'] in ids:
                inf.error(os.path.relpath(r, a.dir), 'id duplicado %s (también en %s)' % (d['id'], ids[d['id']]))
            ids[d['id']] = os.path.relpath(r, a.dir)
    tot_pasos = tot_cap = tot_fotos = 0
    por_mod = {}
    if a.modulo:
        procs = [(r, d) for r, d in procs if isinstance(d, dict) and d.get('modulo') in a.modulo]
        a.indice = False
    ids_activos = {d.get('id') for r, d in procs if r not in reserva and isinstance(d, dict)}
    for r, d in procs:
        if r in reserva:
            c = validar_proc(r, d, base, inf, set(ids), os.path.join(a.dir, '_reserva'))
        else:
            c = validar_proc(r, d, base, inf, ids_activos, a.dir)
        if c:
            tot_pasos += c['pasos']
            tot_cap += c['con_foto']
            tot_fotos += c['fotos']
        if isinstance(d, dict):
            por_mod[d.get('modulo')] = por_mod.get(d.get('modulo'), 0) + 1

    print('Validador de procedimientos de Metrín  ·  %s  ·  %d fragmentos en la base' % (LECTOR, base.n))
    if reserva:
        print('(incluye %d plantillas de _reserva/, que no entran al índice)' % len(reserva))
    print('Procedimientos: %d  ·  pasos (con subpasos): %d  ·  pasos con foto: %d (%.1f %%)  ·  fotos: %d' % (
        len(procs), tot_pasos, tot_cap, 100.0 * tot_cap / max(1, tot_pasos), tot_fotos))
    for m in MODULOS:
        if por_mod.get(m):
            print('  %-24s %3d' % (m, por_mod[m]))
    print('Dudosos (soporte léxico bajo el umbral): %d' % len(inf.dudosos))
    if a.dudosos or (inf.dudosos and len(inf.dudosos) <= 40):
        for rel, lugar, ratio, comp, total, texto in inf.dudosos:
            print('  ? %s · %s · %d/%d (%.0f %%) · %s' % (rel, lugar, comp, total, 100 * ratio, texto[:110]))
    print('Fotos dudosas (poco vocabulario en común con su paso): %d' % len(inf.fotos_dudosas))
    if a.dudosos or (inf.fotos_dudosas and len(inf.fotos_dudosas) <= 40):
        for rel, lugar, ruta, comp, total, texto in inf.fotos_dudosas:
            print('  ? %s · %s · %s · %d/%d · %s' % (rel, lugar, ruta.split('/')[-1], comp, total, texto[:90]))
    if a.calibrar:
        calibrar(base, procs)
    if inf.errores:
        print('ERRORES: %d' % len(inf.errores))
        for e in inf.errores:
            print('  ✗ ' + e)
    else:
        print('Errores: 0')

    if a.indice:
        if inf.errores:
            print('INDICE.json NO se escribe mientras haya errores.')
        else:
            def contar(lst):
                return sum(1 + contar(p.get('sub')) for p in lst or [])
            indice = {}
            for r, d in sorted([x for x in procs if x[0] not in reserva], key=lambda x: x[1]['id']):
                indice[d['id']] = {'titulo': d['titulo'], 'modulo': d['modulo'], 'nivel': d['nivel'],
                                   'n_pasos': contar(d['pasos']), 'preguntas': d['preguntas'],
                                   'aliases': d['aliases'], 'entidades': d['entidades'],
                                   'archivo': os.path.relpath(r, a.dir).replace(os.sep, '/')}
            with open(os.path.join(a.dir, 'INDICE.json'), 'w', encoding='utf-8') as fh:
                json.dump(indice, fh, ensure_ascii=False, indent=1)
                fh.write('\n')
            print('INDICE.json escrito: %d procedimientos' % len(indice))

    if inf.errores or (a.estricto and inf.dudosos):
        return 1
    return 0


if __name__ == '__main__':
    sys.exit(main())
