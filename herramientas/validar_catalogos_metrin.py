#!/usr/bin/env python3
"""Valida los catálogos, el glosario y las plantillas de Metrín. Sale con código ≠ 0 si algo falla.

    .venv/bin/python3 herramientas/validar_catalogos_metrin.py            # todo
    .venv/bin/python3 herramientas/validar_catalogos_metrin.py --avisos   # muestra también los avisos

Qué comprueba:
  * kb/catalogos/intenciones.yml y tipo_respuesta.yml (entrenamiento): esquema, `sintetico: true`, mínimo de ejemplos
    por clase, sin vacíos ni duplicados (dentro ni entre clases, ni siquiera salvo acentos, signos o mayúsculas), rutas y
    modos válidos, plantillas que existen y que declaran la acción de la intención.
  * kb/catalogos/ruido.yml (clase «ninguno»): esquema y que no repita nada del entrenamiento.
  * *_prueba.yml: mínimos por clase, mismas clases que el entrenamiento, casos obligatorios presentes con su etiqueta, y
    que NO se solape con el entrenamiento: igual tras normalizar = error; casi igual (max(Jaccard, SequenceMatcher)
    ≥ 0,90) = error; 0,85–0,90 = aviso.
  * kb/conceptos/*.yml: esquema, id = nombre de archivo, entre 30 y 50 términos, cada `fuente` existe en
    kb/fragmentos*.jsonl, cada `procedimientos_relacionados` existe en kb/procedimientos/, y que exista SIN_FUENTE.md.
  * metrin/plantillas/respuestas.yml: plantillas obligatorias, campos, cada {parte} de `forma` existe (y cada parte se
    usa), cada {{SLOT}} usado está declarado en `slots` y protegido, cada etiqueta de `prohibido` está en
    `reglas_prohibido`, acciones conocidas, índices de `fijas`/`por_accion` válidos y sin jerga prohibida.
"""
from __future__ import annotations

import argparse
import difflib
import glob
import json
import os
import re
import sys
import unicodedata

try:
    import yaml
except ImportError:  # pragma: no cover
    sys.exit("Falta PyYAML: usa .venv/bin/python3 (requirements.txt la fija)")

RAIZ = os.path.normpath(os.path.join(os.path.dirname(os.path.abspath(__file__)), ".."))
CAT = os.path.join(RAIZ, "kb", "catalogos")
CONCEPTOS = os.path.join(RAIZ, "kb", "conceptos")
PROCEDIMIENTOS = os.path.join(RAIZ, "kb", "procedimientos")
PLANTILLAS = os.path.join(RAIZ, "metrin", "plantillas", "respuestas.yml")

RUTAS = {"rag", "procedimiento", "concepto", "conversacion", "humano"}
MODOS = {"concepto", "procedimiento", "navegacion", "diagnostico", "configuracion", "comparacion", "aclarar"}
CLASES_TIPO = ["CONCEPT", "PROCEDURE", "NAVIGATION", "TROUBLESHOOTING", "CONFIGURATION", "COMPARISON", "UNKNOWN"]
MIN_PRUEBA_TIPO = {"CONCEPT": 50, "PROCEDURE": 50, "NAVIGATION": 50, "TROUBLESHOOTING": 50, "CONFIGURATION": 50,
                   "COMPARISON": 30, "UNKNOWN": 30}
CASOS_INTENCIONES = {"¿qué es un metrado?": "consulta_concepto",
                     "¿Cómo registro un nuevo presupuesto en S10?": "soporte_como_hacer"}
CASOS_TIPO = {"como hago un metrado": "PROCEDURE", "como registro metrado": "PROCEDURE",
              "cómo modifico una partida": "PROCEDURE", "que es metrado": "CONCEPT",
              "no puedo guardar metrado": "TROUBLESHOOTING", "donde veo metrados": "NAVIGATION",
              "metrado s10": "UNKNOWN"}
PLANTILLAS_OBLIGATORIAS = ["PROCEDIMIENTO_INTRO", "PASO", "PASOS_BLOQUE", "PREGUNTA_AVANCE", "VERIFICACION",
                           "ERROR_FRECUENTE", "CONCEPTO", "ACLARAR_MODULO", "ACLARAR_TAREA", "RETOMA", "SIN_EVIDENCIA",
                           "DERIVAR_HUMANO", "SALUDO", "DESPEDIDA"]
TIPOS_PLANTILLA = {"mensaje", "paso", "pregunta", "aviso", "retoma"}
ANIMOS = {"happy", "thinking", "confused", "surprised"}       # learning/celebrating: solo por estado real (§7–§8)
CLAVES_CONCEPTO = {"id", "termino", "sinonimos", "definicion", "procedimientos_relacionados", "fuente", "revision"}
CLAVES_CONCEPTO_OPC = {"en_s10", "ejemplo", "notas"}
MIN_TERMINOS, MAX_TERMINOS = 30, 50
PARECIDO_ERROR, PARECIDO_AVISO = 0.90, 0.85

# Formas de tuteo: las plantillas van de usted (como kb/procedimientos/ESQUEMA.md).
TUTEO = {"te", "tu", "ti", "contigo", "quieres", "quieras", "tienes", "tengas", "puedes", "necesitas", "prefieres",
         "sabes", "cuentame", "dime", "avisame", "hazlo", "hazlos", "lograste", "termines"}
RE_SLOT = re.compile(r"\{\{([A-Z0-9_]+)\}\}")
RE_PARTE = re.compile(r"(?<!\{)\{([a-z0-9_]+)\}(?!\})")


# --- utilidades compartidas (las importa evaluar_clasificador_metrin.py) -------------------------------------------
def normalizar_clave(texto: str) -> str:
    """Clave de duplicados: minúsculas, sin diacríticos, sin signos, espacios colapsados. Si no queda nada (solo signos),
    el texto crudo sin espacios extremos."""
    t = unicodedata.normalize("NFD", (texto or "").lower())
    t = "".join(c for c in t if unicodedata.category(c) != "Mn")
    t = re.sub(r"[^a-z0-9\s]", " ", t)
    t = re.sub(r"\s+", " ", t).strip()
    return t or (texto or "").strip()


def parecido(a: str, b: str) -> float:
    """max(Jaccard de palabras, SequenceMatcher) sobre claves normalizadas (el mismo criterio que kddesign)."""
    sa, sb = set(a.split()), set(b.split())
    jac = len(sa & sb) / len(sa | sb) if sa and sb else 0.0
    return max(jac, difflib.SequenceMatcher(None, a, b).ratio())


def leer_yaml(ruta: str):
    with open(ruta, encoding="utf-8") as fh:
        return yaml.safe_load(fh)


def clases_de(datos: dict) -> dict:
    """{clase: definición} de un catálogo; acepta `intenciones:` o `clases:`."""
    return (datos or {}).get("intenciones") or (datos or {}).get("clases") or {}


def ejemplos_de(datos: dict) -> dict:
    """{clase: [frases]} de un catálogo o de su prueba."""
    out = {}
    for nom, d in clases_de(datos).items():
        ej = d.get("ejemplos") if isinstance(d, dict) else d
        out[nom] = [str(x) for x in (ej or [])]
    return out


def ruido_de(datos: dict) -> list:
    """Todas las frases de ruido.yml (todas sus listas)."""
    return [str(x) for k, v in (datos or {}).items() if isinstance(v, list) for x in v]


class Informe:
    def __init__(self):
        self.errores, self.avisos = [], []

    def error(self, donde, msg):
        self.errores.append(f"[{donde}] {msg}")

    def aviso(self, donde, msg):
        self.avisos.append(f"[{donde}] {msg}")


# --- catálogos de entrenamiento -------------------------------------------------------------------------------------
def validar_entrenamiento(inf: Informe, nombre: str, datos: dict, plantillas: dict, es_tipo: bool) -> dict:
    if not isinstance(datos, dict):
        inf.error(nombre, "no es un mapa YAML")
        return {}
    if datos.get("sintetico") is not True:
        inf.error(nombre, "falta `sintetico: true`")
    minimo = int(datos.get("minimo_ejemplos") or 0)
    if minimo < (40 if es_tipo else 25):
        inf.error(nombre, f"minimo_ejemplos={minimo} por debajo del exigido ({40 if es_tipo else 25})")
    clases = clases_de(datos)
    if not clases:
        inf.error(nombre, "sin clases/intenciones")
        return {}
    if es_tipo:
        faltan = set(CLASES_TIPO) - set(clases)
        sobran = set(clases) - set(CLASES_TIPO)
        if faltan or sobran:
            inf.error(nombre, f"clases esperadas {CLASES_TIPO}; faltan {sorted(faltan)}, sobran {sorted(sobran)}")
    vistos = {}
    for nom, d in clases.items():
        donde = f"{nombre}:{nom}"
        if not isinstance(d, dict):
            inf.error(donde, "la clase no es un mapa")
            continue
        if not str(d.get("descripcion") or "").strip():
            inf.error(donde, "falta descripcion")
        if es_tipo:
            if d.get("modo") not in MODOS:
                inf.error(donde, f"modo {d.get('modo')!r} no está en {sorted(MODOS)}")
            accion = d.get("modo")
        else:
            if d.get("ruta") not in RUTAS:
                inf.error(donde, f"ruta {d.get('ruta')!r} no está en {sorted(RUTAS)}")
            if not str(d.get("accion") or "").strip():
                inf.error(donde, "falta accion")
            accion = d.get("accion")
        for p in d.get("plantillas") or []:
            if p not in plantillas:
                inf.error(donde, f"plantilla {p} no existe en respuestas.yml")
            elif accion not in (plantillas[p].get("acciones") or []):
                inf.error(donde, f"la plantilla {p} no declara la acción/modo «{accion}»")
        ejemplos = d.get("ejemplos") or []
        if len(ejemplos) < minimo:
            inf.error(donde, f"{len(ejemplos)} ejemplos (mínimo {minimo})")
        for t in ejemplos:
            if not isinstance(t, str) or not t.strip():
                inf.error(donde, f"ejemplo vacío o no textual: {t!r}")
                continue
            k = normalizar_clave(t)
            if k in vistos:
                otro = vistos[k]
                inf.error(donde, f"duplicado {t!r} (= {otro[1]!r} en «{otro[0]}»)")
            else:
                vistos[k] = (nom, t)
    return vistos


def validar_ruido(inf: Informe, datos: dict, entrenamiento: dict) -> dict:
    nombre = "ruido.yml"
    if not isinstance(datos, dict):
        inf.error(nombre, "no es un mapa YAML")
        return {}
    if datos.get("sintetico") is not True:
        inf.error(nombre, "falta `sintetico: true`")
    if datos.get("clase") != "ninguno":
        inf.error(nombre, "falta `clase: ninguno`")
    frases = ruido_de(datos)
    if len(frases) < 40:
        inf.error(nombre, f"solo {len(frases)} frases (mínimo 40)")
    vistos = {}
    for t in frases:
        if not t.strip():
            inf.error(nombre, "frase vacía")
            continue
        k = normalizar_clave(t)
        if k in vistos:
            inf.error(nombre, f"duplicado {t!r}")
        vistos[k] = ("ninguno", t)
        if k in entrenamiento:
            inf.error(nombre, f"{t!r} está también en la intención «{entrenamiento[k][0]}»")
    return vistos


def validar_prueba(inf: Informe, nombre: str, datos: dict, entrenamiento: dict, clases_ent: set, minimos: dict,
                   casos: dict, extra_clases: set) -> None:
    if not isinstance(datos, dict):
        inf.error(nombre, "no es un mapa YAML")
        return
    if datos.get("sintetico") is not True:
        inf.error(nombre, "falta `sintetico: true`")
    prueba = ejemplos_de(datos)
    faltan = clases_ent - set(prueba)
    sobran = set(prueba) - clases_ent - extra_clases
    if faltan:
        inf.error(nombre, f"clases sin prueba: {sorted(faltan)}")
    if sobran:
        inf.error(nombre, f"clases que no existen en el entrenamiento: {sorted(sobran)}")
    claves_ent = list(entrenamiento.items())
    vistos = {}
    for nom, frases in prueba.items():
        minimo = minimos.get(nom, minimos.get("*", 0))
        if len(frases) < minimo:
            inf.error(f"{nombre}:{nom}", f"{len(frases)} frases (mínimo {minimo})")
        for f in frases:
            k = normalizar_clave(f)
            if not f.strip():
                inf.error(f"{nombre}:{nom}", "frase vacía")
                continue
            if k in vistos:
                inf.error(f"{nombre}:{nom}", f"duplicado dentro de la prueba: {f!r}")
            vistos[k] = nom
            if k in entrenamiento:
                inf.error(f"{nombre}:{nom}", f"{f!r} es IGUAL a un ejemplo de entrenamiento ({entrenamiento[k][0]})")
                continue
            mejor, kmax = 0.0, None
            for ke, (ie, te) in claves_ent:
                s = parecido(k, ke)
                if s > mejor:
                    mejor, kmax = s, (ie, te)
            if mejor >= PARECIDO_ERROR:
                inf.error(f"{nombre}:{nom}", f"{f!r} casi igual ({mejor:.2f}) a {kmax[1]!r} ({kmax[0]}) del entrenamiento")
            elif mejor >= PARECIDO_AVISO:
                inf.aviso(f"{nombre}:{nom}", f"{f!r} parecida ({mejor:.2f}) a {kmax[1]!r} ({kmax[0]})")
    for texto, etiqueta in casos.items():
        donde = [c for c, fs in prueba.items() if texto in fs]
        if donde != [etiqueta]:
            inf.error(nombre, f"caso obligatorio {texto!r} debe estar en «{etiqueta}» (está en {donde or 'ninguna'})")


# --- glosario -------------------------------------------------------------------------------------------------------
def ids_fragmentos() -> dict:
    """{id: {manuales}}. Ojo: en kb/fragmentos.jsonl 200 ids de secciones web se repiten entre manuales
    (kb/procedimientos/ESQUEMA.md, «ids ambiguos»): el id solo no identifica el fragmento."""
    ids: dict = {}
    for ruta in glob.glob(os.path.join(RAIZ, "kb", "fragmentos*.jsonl")):
        with open(ruta, encoding="utf-8") as fh:
            for linea in fh:
                linea = linea.strip()
                if linea:
                    d = json.loads(linea)
                    ids.setdefault(d["id"], set()).add(d.get("manual") or "")
    return ids


def ids_procedimientos() -> set:
    ids = set()
    for ruta in glob.glob(os.path.join(PROCEDIMIENTOS, "**", "*.y*ml"), recursive=True):
        ids.add(os.path.splitext(os.path.basename(ruta))[0])
        try:
            d = leer_yaml(ruta)
            if isinstance(d, dict) and d.get("id"):
                ids.add(str(d["id"]))
        except Exception:  # un procedimiento roto no es asunto de este validador
            pass
    return ids


def validar_conceptos(inf: Informe) -> int:
    archivos = sorted(glob.glob(os.path.join(CONCEPTOS, "*.yml")))
    if not os.path.exists(os.path.join(CONCEPTOS, "SIN_FUENTE.md")):
        inf.error("conceptos", "falta kb/conceptos/SIN_FUENTE.md")
    if not MIN_TERMINOS <= len(archivos) <= MAX_TERMINOS:
        inf.error("conceptos", f"{len(archivos)} términos (se esperan {MIN_TERMINOS}–{MAX_TERMINOS})")
    frag = ids_fragmentos()
    procs = ids_procedimientos()
    terminos, sinonimos = {}, {}
    for ruta in archivos:
        stem = os.path.splitext(os.path.basename(ruta))[0]
        donde = f"conceptos/{stem}"
        try:
            d = leer_yaml(ruta)
        except yaml.YAMLError as e:
            inf.error(donde, f"YAML inválido: {e}")
            continue
        if not isinstance(d, dict):
            inf.error(donde, "no es un mapa")
            continue
        faltan = CLAVES_CONCEPTO - set(d)
        sobran = set(d) - CLAVES_CONCEPTO - CLAVES_CONCEPTO_OPC
        if faltan:
            inf.error(donde, f"faltan claves {sorted(faltan)}")
        if sobran:
            inf.error(donde, f"claves no permitidas {sorted(sobran)}")
        if d.get("id") != stem:
            inf.error(donde, f"id {d.get('id')!r} ≠ nombre de archivo")
        if not re.fullmatch(r"[a-z0-9_]+", stem):
            inf.error(donde, "el id debe ser minúsculas ASCII y guiones bajos")
        t = str(d.get("termino") or "").strip()
        if not t:
            inf.error(donde, "termino vacío")
        elif normalizar_clave(t) in terminos:
            inf.error(donde, f"término repetido con {terminos[normalizar_clave(t)]}")
        else:
            terminos[normalizar_clave(t)] = stem
        if not str(d.get("definicion") or "").strip():
            inf.error(donde, "definicion vacía")
        for clave in ("en_s10", "ejemplo", "notas"):
            if clave in d and not str(d.get(clave) or "").strip():
                inf.error(donde, f"{clave} presente pero vacío (omítelo)")
        if not isinstance(d.get("sinonimos"), list):
            inf.error(donde, "sinonimos debe ser una lista")
        else:
            for s in d["sinonimos"]:
                k = normalizar_clave(str(s))
                if k in sinonimos and sinonimos[k] != stem:
                    inf.aviso(donde, f"sinónimo {s!r} también en {sinonimos[k]}")
                sinonimos[k] = stem
        fuentes = d.get("fuente")
        if not isinstance(fuentes, list) or not fuentes:
            inf.error(donde, "fuente debe ser una lista no vacía")
        else:
            ambiguas = 0
            for f in fuentes:
                if str(f) not in frag:
                    inf.error(donde, f"fuente {f!r} no existe en kb/fragmentos*.jsonl")
                elif len(frag[str(f)]) > 1:
                    ambiguas += 1
            if ambiguas:
                inf.ambiguas = getattr(inf, "ambiguas", 0) + ambiguas
        rel = d.get("procedimientos_relacionados")
        if not isinstance(rel, list):
            inf.error(donde, "procedimientos_relacionados debe ser una lista")
        else:
            for p in rel:
                if str(p) not in procs:
                    inf.error(donde, f"procedimiento relacionado {p!r} no existe en kb/procedimientos/")
        rev = d.get("revision")
        if not isinstance(rev, dict) or {"generado", "por", "revisado_por_humano"} - set(rev or {}):
            inf.error(donde, "revision debe tener generado, por y revisado_por_humano")
        elif not isinstance(rev.get("revisado_por_humano"), bool):
            inf.error(donde, "revision.revisado_por_humano debe ser booleano")
    if getattr(inf, "ambiguas", 0):
        inf.aviso("conceptos", f"{inf.ambiguas} citas usan un id repetido en varios manuales de kb/fragmentos.jsonl: "
                  "resolver por (id, manual), no solo por id (el manual va en el comentario de cada fuente)")
    return len(archivos)


# --- plantillas -----------------------------------------------------------------------------------------------------
def contar_frases(texto: str) -> int:
    t = RE_SLOT.sub("X", texto).strip()
    if not t:
        return 0
    return max(1, len(re.findall(r"[.!?…](?=\s|$)", t)))


def validar_plantillas(inf: Informe, datos: dict, acciones_conocidas: set) -> None:
    nombre = "respuestas.yml"
    if not isinstance(datos, dict):
        inf.error(nombre, "no es un mapa YAML")
        return
    for clave in ("version", "config", "slots", "reglas_prohibido", "plantillas"):
        if clave not in datos:
            inf.error(nombre, f"falta `{clave}`")
    slots = set((datos.get("slots") or {}).keys())
    reglas = set((datos.get("reglas_prohibido") or {}).keys())
    config = datos.get("config") or {}
    jerga = [normalizar_clave(j) for j in config.get("jerga_prohibida") or []]
    acciones_conocidas = acciones_conocidas | set((datos.get("acciones_extra") or {}).keys())
    plantillas = datos.get("plantillas") or {}
    for p in PLANTILLAS_OBLIGATORIAS:
        if p not in plantillas:
            inf.error(nombre, f"falta la plantilla obligatoria {p}")
    usados_global = set()
    for nom, pl in plantillas.items():
        donde = f"plantilla {nom}"
        if not isinstance(pl, dict):
            inf.error(donde, "no es un mapa")
            continue
        for clave in ("tipo", "acciones", "protegidos", "prohibido", "max_frases", "forma", "partes"):
            if clave not in pl:
                inf.error(donde, f"falta `{clave}`")
        if pl.get("tipo") not in TIPOS_PLANTILLA:
            inf.error(donde, f"tipo {pl.get('tipo')!r} no está en {sorted(TIPOS_PLANTILLA)}")
        if pl.get("animo") is not None and pl.get("animo") not in ANIMOS:
            inf.error(donde, f"animo {pl.get('animo')!r} no está en {sorted(ANIMOS)}")
        acciones = pl.get("acciones") or []
        if not acciones:
            inf.error(donde, "acciones vacía")
        for a in acciones:
            if a not in acciones_conocidas:
                inf.error(donde, f"acción desconocida {a!r} (ni de los catálogos ni de acciones_extra)")
        mf = pl.get("max_frases")
        if not isinstance(mf, int) or mf < 1:
            inf.error(donde, "max_frases debe ser un entero ≥ 1")
        for etiqueta in pl.get("prohibido") or []:
            if etiqueta not in reglas:
                inf.error(donde, f"etiqueta prohibida {etiqueta!r} no descrita en reglas_prohibido")
        if not pl.get("prohibido"):
            inf.error(donde, "prohibido vacío")
        partes = pl.get("partes") or {}
        if not isinstance(partes, dict) or not partes:
            inf.error(donde, "partes vacía")
            continue
        forma = str(pl.get("forma") or "")
        en_forma = RE_PARTE.findall(forma)
        for p in en_forma:
            if p not in partes:
                inf.error(donde, f"la forma usa {{{p}}} pero no hay esa parte")
        for p in partes:
            if p not in en_forma:
                inf.error(donde, f"la parte {p!r} no se usa en la forma")
        textos = [forma]
        for p, variantes in partes.items():
            if not isinstance(variantes, list) or not variantes:
                inf.error(donde, f"la parte {p!r} no tiene variantes")
                continue
            for v in variantes:
                if not isinstance(v, str):
                    inf.error(donde, f"variante no textual en {p!r}: {v!r}")
            textos += [v for v in variantes if isinstance(v, str)]
            fija = p in (pl.get("fijas") or {})
            if len(variantes) < 2 and not fija and not pl.get("sin_modelo"):
                inf.aviso(donde, f"la parte {p!r} tiene una sola variante y no está en `fijas`")
        for p, i in (pl.get("fijas") or {}).items():
            if p not in partes or not isinstance(i, int) or not 0 <= i < len(partes[p]):
                inf.error(donde, f"fijas.{p}={i!r} no apunta a una variante")
        for acc, restr in (pl.get("por_accion") or {}).items():
            if acc not in acciones:
                inf.error(donde, f"por_accion.{acc}: la plantilla no declara esa acción")
            for p, idx in (restr or {}).items():
                if p not in partes or not all(isinstance(i, int) and 0 <= i < len(partes[p]) for i in idx or []):
                    inf.error(donde, f"por_accion.{acc}.{p}={idx!r} no apunta a variantes válidas")
        for p, txt in (pl.get("respaldo") or {}).items():
            if p not in partes:
                inf.error(donde, f"respaldo de una parte inexistente: {p!r}")
            textos.append(str(txt))
        usados = set()
        for t in textos:
            usados |= set(RE_SLOT.findall(t))
        usados_global |= usados
        for s in usados - slots:
            inf.error(donde, f"slot {{{{{s}}}}} usado pero no declarado en `slots`")
        protegidos = set(pl.get("protegidos") or [])
        for s in protegidos - slots:
            inf.error(donde, f"protegido {s!r} no declarado en `slots`")
        for s in usados - protegidos:
            inf.error(donde, f"el slot {{{{{s}}}}} se usa pero no está en `protegidos`")
        for s in protegidos - usados:
            inf.aviso(donde, f"protegido {s!r} no se usa en la plantilla")
        # jerga regional, fuera de los slots
        for t in textos:
            limpio = " " + normalizar_clave(RE_PARTE.sub(" ", RE_SLOT.sub(" ", t))) + " "
            for j in jerga:
                if f" {j} " in limpio:
                    inf.error(donde, f"jerga prohibida «{j}» en {t!r}")
            if datos.get("trato") == "usted":
                for w in sorted(set(limpio.split()) & TUTEO):
                    inf.error(donde, f"tuteo «{w}» en {t!r} (el trato es usted)")
        # tope de frases (aproximado): la variante más larga de cada parte, o su respaldo
        if isinstance(mf, int):
            total = 0
            for p in en_forma:
                cands = [contar_frases(v) for v in partes.get(p) or [] if isinstance(v, str)]
                total += max(cands or [0])
            if total > mf:
                inf.aviso(donde, f"la combinación más larga tiene ~{total} frases (max_frases {mf})")
    for s in slots - usados_global:
        inf.aviso(nombre, f"slot {s!r} declarado y sin uso")


# --- principal ------------------------------------------------------------------------------------------------------
def validar_todo() -> Informe:
    inf = Informe()
    rutas = {n: os.path.join(CAT, n) for n in ("intenciones.yml", "intenciones_prueba.yml", "ruido.yml",
                                                "tipo_respuesta.yml", "tipo_respuesta_prueba.yml")}
    datos = {}
    for n, r in rutas.items():
        if not os.path.exists(r):
            inf.error(n, "no existe")
            continue
        try:
            datos[n] = leer_yaml(r)
        except yaml.YAMLError as e:
            inf.error(n, f"YAML inválido: {e}")
    try:
        plant = leer_yaml(PLANTILLAS) if os.path.exists(PLANTILLAS) else None
    except yaml.YAMLError as e:
        inf.error("respuestas.yml", f"YAML inválido: {e}")
        plant = None
    if plant is None:
        inf.error("respuestas.yml", f"no existe o está vacío: {PLANTILLAS}")
        plant = {}
    plantillas = plant.get("plantillas") or {}

    ent_int, ent_tipo = {}, {}
    if "intenciones.yml" in datos:
        ent_int = validar_entrenamiento(inf, "intenciones.yml", datos["intenciones.yml"], plantillas, es_tipo=False)
    ruido = {}
    if "ruido.yml" in datos:
        ruido = validar_ruido(inf, datos["ruido.yml"], ent_int)
    if "intenciones_prueba.yml" in datos:
        validar_prueba(inf, "intenciones_prueba.yml", datos["intenciones_prueba.yml"], {**ent_int, **ruido},
                       set(clases_de(datos.get("intenciones.yml", {}))),
                       {"*": int(datos["intenciones_prueba.yml"].get("minimo_ejemplos") or 8)},
                       CASOS_INTENCIONES, {"ninguno"})
    if "tipo_respuesta.yml" in datos:
        ent_tipo = validar_entrenamiento(inf, "tipo_respuesta.yml", datos["tipo_respuesta.yml"], plantillas, es_tipo=True)
    if "tipo_respuesta_prueba.yml" in datos:
        validar_prueba(inf, "tipo_respuesta_prueba.yml", datos["tipo_respuesta_prueba.yml"], ent_tipo,
                       set(clases_de(datos.get("tipo_respuesta.yml", {}))), MIN_PRUEBA_TIPO, CASOS_TIPO, set())

    acciones = {str(d.get("accion")) for d in clases_de(datos.get("intenciones.yml", {})).values() if isinstance(d, dict)}
    acciones |= {str(d.get("modo")) for d in clases_de(datos.get("tipo_respuesta.yml", {})).values() if isinstance(d, dict)}
    validar_plantillas(inf, plant, acciones)
    inf.n_conceptos = validar_conceptos(inf)
    inf.resumen = {n: {c: len(v) for c, v in ejemplos_de(d).items()} for n, d in datos.items() if n != "ruido.yml"}
    inf.resumen["ruido.yml"] = len(ruido_de(datos.get("ruido.yml")))
    inf.resumen["plantillas"] = len(plantillas)
    return inf


def main(argv=None) -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--avisos", action="store_true", help="muestra también los avisos")
    a = ap.parse_args(argv)
    inf = validar_todo()
    for n, r in inf.resumen.items():
        if isinstance(r, dict):
            tot = sum(r.values())
            print(f"{n}: {len(r)} clases, {tot} frases (mín {min(r.values()) if r else 0}, máx {max(r.values()) if r else 0})")
        else:
            print(f"{n}: {r}")
    print(f"conceptos: {inf.n_conceptos} términos")
    if a.avisos:
        for x in inf.avisos:
            print("AVISO ", x)
    for x in inf.errores:
        print("ERROR ", x)
    print(f"\n{len(inf.errores)} errores, {len(inf.avisos)} avisos" + ("" if a.avisos or not inf.avisos else " (--avisos para verlos)"))
    print("OK" if not inf.errores else "FALLA")
    return 1 if inf.errores else 0


if __name__ == "__main__":
    sys.exit(main())
