#!/usr/bin/env python3
"""Benchmark de modelos LOCALES para la V2 de Metrín: motor de DECISIÓN y GENERADOR que solo reformula.

Mide, para cada GGUF, con llama-server en Metal y en CPU de 4 hilos (estimación del EC2 sin GPU):

  decisión   tipo de respuesta (7 clases) y suficiencia de evidencia (si/no), salida JSON con json_schema
             (gramática) y sin ella; exactitud, F1 por clase, % de JSON válido, latencia p50/p95, RSS.
  generación reformular un plan ya validado (pasos con **negritas**, ids de foto) devolviendo el mismo JSON;
             % de salidas válidas (mismos pasos, orden, negritas y fotos, sin términos nuevos), latencia,
             tokens/s de prompt y de generación, RSS.
  Jev        el servidor de decisión estilo Jev (API POST /v1/systemone), en MLX (:8765, el que ya corre) o
             su GGUF con jeva.cpp (Metal y CPU).

Conjuntos: metrin/eval/modelos/decision.jsonl y generacion.jsonl. Resultados: metrin/eval/modelos/resultados/.
Informe: metrin/eval/MODELOS.md. Solo biblioteca estándar.

Uso (desde s10-conocimiento/):
  .venv/bin/python3 herramientas/benchmark_modelos.py modelos
  .venv/bin/python3 herramientas/benchmark_modelos.py correr  --modelo qwen3.5-2b --modo metal
  .venv/bin/python3 herramientas/benchmark_modelos.py correr  --modelo qwen3.5-2b --modo cpu4      (subconjunto)
  .venv/bin/python3 herramientas/benchmark_modelos.py jev     --url http://127.0.0.1:8765 --nombre jev-0.8b-mlx
  .venv/bin/python3 herramientas/benchmark_modelos.py jev     --gguf <ruta> --modo cpu4 --nombre jev-0.8b-gguf
  .venv/bin/python3 herramientas/benchmark_modelos.py ciego   (exporta la muestra a ciegas para puntuar)
  .venv/bin/python3 herramientas/benchmark_modelos.py resumen (tablas en Markdown)

Cada `correr` levanta UN llama-server en 127.0.0.1:18431 (o --puerto), mide y lo apaga siempre (try/finally).
"""
from __future__ import annotations

import argparse
import collections
import glob
import json
import os
import random
import re
import statistics
import subprocess
import sys
import threading
import time
import unicodedata
import urllib.error
import urllib.request

RAIZ = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DIR_EVAL = os.path.join(RAIZ, 'metrin', 'eval', 'modelos')
DIR_RES = os.path.join(DIR_EVAL, 'resultados')
DIR_CIEGO = os.path.join(DIR_EVAL, 'ciego')
CACHE = os.path.expanduser('~/.cache/metrin-modelos')
OLLAMA = os.path.expanduser('~/.ollama/models/blobs')
LLAMA_SERVER = '/opt/homebrew/bin/llama-server'
JEVA_SERVER = os.path.join(RAIZ, 'jeva.cpp', 'build', 'bin', 'llama-server')

# ---------------------------------------------------------------------------------------------------------------
# Modelos candidatos. `pensar`: plantilla híbrida con razonamiento (se apaga con enable_thinking=False).
MODELOS = {
    'qwen3.5-2b': dict(gguf=f'{CACHE}/Qwen3.5-2B-Q4_K_M.gguf', hf='unsloth/Qwen3.5-2B-GGUF', licencia='Apache-2.0',
                       pensar=True),
    'qwen3.5-4b': dict(gguf=f'{CACHE}/Qwen3.5-4B-Q4_K_M.gguf', hf='unsloth/Qwen3.5-4B-GGUF', licencia='Apache-2.0',
                       pensar=True),
    'qwen3-1.7b': dict(gguf=f'{OLLAMA}/sha256-3d0b790534fe4b79525fc3692950408dca41171676ed7e21db57af5c65ef6ab6',
                       hf='ollama qwen3:1.7b (Q4_K_M)', licencia='Apache-2.0', pensar=True),
    'qwen3-4b-2507': dict(gguf=f'{CACHE}/Qwen3-4B-Instruct-2507-Q4_K_M.gguf', hf='unsloth/Qwen3-4B-Instruct-2507-GGUF',
                          licencia='Apache-2.0', pensar=False),
    'granite-4.0-micro': dict(gguf=f'{CACHE}/granite-4.0-micro-Q4_K_M.gguf', hf='ibm-granite/granite-4.0-micro-GGUF',
                              licencia='Apache-2.0', pensar=False),
    'qwen2.5-coder-7b': dict(gguf=f'{OLLAMA}/sha256-60e05f2100071479f596b964f89f510f057ce397ea22f2833a0cfe029bfc2463',
                             hf='ollama qwen2.5-coder:7b (Q4_K_M), defecto actual de Metrín', licencia='Apache-2.0',
                             pensar=False),
}

TIPOS = ['CONCEPT', 'PROCEDURE', 'NAVIGATION', 'TROUBLESHOOTING', 'CONFIGURATION', 'COMPARISON', 'UNKNOWN']
DEF_TIPOS = {
    'CONCEPT': 'pregunta qué es, qué significa o para qué sirve un término o elemento de S10',
    'PROCEDURE': 'pide cómo hacer una tarea (registrar, crear, modificar, eliminar, generar, imprimir, importar, '
                 'copiar, anular…), es decir, los pasos',
    'NAVIGATION': 'pregunta dónde está o cómo llegar a una pantalla, menú, opción, botón o reporte',
    'TROUBLESHOOTING': 'reporta un error, un mensaje, o algo que no funciona, no aparece, no guarda o no cuadra',
    'CONFIGURATION': 'pide cómo configurar, parametrizar, activar o definir opciones, valores por defecto, permisos '
                     'o parámetros del sistema',
    'COMPARISON': 'pide la diferencia entre dos o más cosas, o cuál conviene usar',
    'UNKNOWN': 'el mensaje es demasiado corto o vago: nombra un tema pero no dice qué quiere saber o hacer',
}

SISTEMA_TIPO = (
    'Clasificas mensajes de usuarios del ERP S10 (software para empresas de construcción) según el TIPO DE '
    'RESPUESTA que necesitan. Opciones:\n'
    + '\n'.join(f'- {k}: {v}.' for k, v in DEF_TIPOS.items())
    + '\nResponde solo con JSON, sin texto adicional: {"tipo": "<OPCIÓN>"}')

SISTEMA_EVID = (
    'Decides si la EVIDENCIA (fragmentos de los manuales oficiales de S10) basta para responder la PREGUNTA. '
    'Responde "si" solo si algún fragmento contiene directamente lo que la pregunta pide (la definición, los pasos, '
    'el lugar, la causa o el dato). Responde "no" si los fragmentos solo tratan el mismo tema, responden otra '
    'pregunta parecida o les falta el dato pedido. No uses conocimiento propio.\n'
    'Responde solo con JSON, sin texto adicional: {"suficiente": "si"} o {"suficiente": "no"}')

SISTEMA_GEN = (
    'Eres el redactor de Metrín, asistente del ERP S10. Recibes un PLAN de respuesta ya validado, en JSON. Tu única '
    'tarea es redactar mejor el texto de cada paso en español neutro, claro y breve (trato de usted, modo '
    'imperativo) y escribir una introducción de una sola oración.\n'
    'Reglas obligatorias:\n'
    '1. Devuelve los mismos pasos, en el mismo orden, con los mismos "n", "id" y "fotos", copiados tal cual.\n'
    '2. Copia literalmente cada término entre ** ** (con sus asteriscos) en su mismo paso. No quites ni añadas '
    'negritas.\n'
    '3. No inventes menús, botones, pantallas, atajos de teclado, pasos ni datos que no estén en el plan.\n'
    '4. No cambies el significado de ningún paso. Si un paso ya es claro, déjalo casi igual.\n'
    'Responde solo con JSON: {"intro": "...", "pasos": [{"n": 1, "id": "...", "texto": "...", "fotos": ["..."]}]}')

ESQ_TIPO = {'type': 'object', 'properties': {'tipo': {'type': 'string', 'enum': TIPOS}}, 'required': ['tipo'],
            'additionalProperties': False}
ESQ_EVID = {'type': 'object', 'properties': {'suficiente': {'type': 'string', 'enum': ['si', 'no']}},
            'required': ['suficiente'], 'additionalProperties': False}
ESQ_GEN = {'type': 'object', 'properties': {
    'intro': {'type': 'string'},
    'pasos': {'type': 'array', 'items': {'type': 'object', 'properties': {
        'n': {'type': 'integer'}, 'id': {'type': 'string'}, 'texto': {'type': 'string'},
        'fotos': {'type': 'array', 'items': {'type': 'string'}}},
        'required': ['n', 'id', 'texto', 'fotos'], 'additionalProperties': False}}},
    'required': ['intro', 'pasos'], 'additionalProperties': False}

# Subconjunto para CPU (la calidad sale de Metal; en CPU se mide sobre todo latencia y RSS).
SUB_CPU_EVID = [f'evid-{i:03d}' for i in list(range(1, 11)) + list(range(21, 31))]
SUB_CPU_GEN = ['gen-001', 'gen-006', 'gen-011', 'gen-016', 'gen-021', 'gen-026', 'gen-031']


# ---------------------------------------------------------------------------------------------------------------
def leer_jsonl(ruta):
    with open(ruta) as f:
        return [json.loads(l) for l in f if l.strip()]


def casos():
    d = leer_jsonl(os.path.join(DIR_EVAL, 'decision.jsonl'))
    g = leer_jsonl(os.path.join(DIR_EVAL, 'generacion.jsonl'))
    return [x for x in d if x['tarea'] == 'tipo_respuesta'], [x for x in d if x['tarea'] == 'evidencia'], g


def post(url, cuerpo, timeout=900):
    req = urllib.request.Request(url, data=json.dumps(cuerpo).encode(), headers={'Content-Type': 'application/json'})
    t0 = time.perf_counter()
    with urllib.request.urlopen(req, timeout=timeout) as r:
        d = json.loads(r.read())
    return d, (time.perf_counter() - t0) * 1000


def pct(xs, p):
    if not xs:
        return None
    xs = sorted(xs)
    k = (len(xs) - 1) * p / 100
    i = int(k)
    return xs[i] if i + 1 >= len(xs) else xs[i] + (xs[i + 1] - xs[i]) * (k - i)


def rss_mb(pid):
    try:
        out = subprocess.run(['ps', '-o', 'rss=', '-p', str(pid)], capture_output=True, text=True).stdout.strip()
        return int(out) / 1024 if out else None
    except Exception:
        return None


def huella_mb(pid):
    """phys_footprint de macOS: a diferencia del RSS, incluye la memoria de Metal (pesos y caché KV en GPU)."""
    try:
        out = subprocess.run(['footprint', '-p', str(pid)], capture_output=True, text=True, timeout=10).stdout
        m = re.search(r'Footprint:\s*([\d.]+)\s*(KB|MB|GB)', out)
        if not m:
            return None
        return float(m.group(1)) * {'KB': 1 / 1024, 'MB': 1, 'GB': 1024}[m.group(2)]
    except Exception:
        return None


class Muestreador(threading.Thread):
    """Muestrea RSS y huella física (footprint) de un proceso cada 0,5 s; guarda los picos."""

    def __init__(self, pid):
        super().__init__(daemon=True)
        self.pid, self.pico, self.pico_huella, self.parar = pid, 0.0, 0.0, threading.Event()

    def run(self):
        while not self.parar.is_set():
            r = rss_mb(self.pid)
            if r:
                self.pico = max(self.pico, r)
            h = huella_mb(self.pid)
            if h:
                self.pico_huella = max(self.pico_huella, h)
            self.parar.wait(0.5)


class Servidor:
    """Un llama-server (o el de jeva.cpp) en un puerto propio; se apaga siempre al salir."""

    def __init__(self, binario, gguf, modo, puerto, log, ctx=4096, extra=()):
        self.url = f'http://127.0.0.1:{puerto}'
        self.args = [binario, '-m', gguf, '--host', '127.0.0.1', '--port', str(puerto), '-c', str(ctx), '-np', '1',
                     '--cache-ram', '0', '--no-webui']
        if modo == 'metal':
            self.args += ['-ngl', '99']
        elif modo == 'cpu4':
            self.args += ['-ngl', '0', '-dev', 'none', '-t', '4', '-tb', '4']
        else:
            raise ValueError(modo)
        self.args += list(extra)
        self.log = log

    def __enter__(self):
        self.flog = open(self.log, 'w')
        t0 = time.time()
        self.p = subprocess.Popen(self.args, stdout=self.flog, stderr=subprocess.STDOUT)
        while True:
            if self.p.poll() is not None:
                raise RuntimeError(f'llama-server terminó al arrancar; ver {self.log}')
            try:
                with urllib.request.urlopen(self.url + '/health', timeout=2) as r:
                    if r.status == 200:
                        break
            except Exception:
                pass
            if time.time() - t0 > 300:
                self.__exit__(None, None, None)
                raise RuntimeError('llama-server no respondió en 300 s')
            time.sleep(0.5)
        self.carga_s = time.time() - t0
        self.rss_carga = rss_mb(self.p.pid)
        self.huella_carga = huella_mb(self.p.pid)
        self.m = Muestreador(self.p.pid)
        self.m.start()
        return self

    def __exit__(self, *a):
        try:
            if hasattr(self, 'm'):
                self.m.parar.set()
            self.p.terminate()
            try:
                self.p.wait(20)
            except subprocess.TimeoutExpired:
                self.p.kill()
                self.p.wait(10)
        finally:
            self.flog.close()


# ---------------------------------------------------------------------------------------------------------------
# Decisión con chat + json_schema (o libre)

def extraer_json(txt):
    txt = (txt or '').strip()
    txt = re.sub(r'^<think>.*?</think>\s*', '', txt, flags=re.S)
    try:
        return json.loads(txt), True
    except Exception:
        pass
    m = re.search(r'\{.*\}', txt, flags=re.S)
    if m:
        try:
            return json.loads(m.group(0)), False
        except Exception:
            pass
    return None, False


def chat(url, sistema, usuario, esquema, max_tokens, pensar):
    cuerpo = {'messages': [{'role': 'system', 'content': sistema}, {'role': 'user', 'content': usuario}],
              'temperature': 0, 'top_k': 1, 'max_tokens': max_tokens, 'cache_prompt': True, 'seed': 7}
    if esquema is not None:
        cuerpo['response_format'] = {'type': 'json_schema', 'json_schema': {'name': 'salida', 'schema': esquema}}
    if pensar:
        cuerpo['chat_template_kwargs'] = {'enable_thinking': False}
    d, ms = post(url + '/v1/chat/completions', cuerpo)
    msg = d['choices'][0]['message']
    return (msg.get('content') or ''), ms, d.get('timings', {}), d['choices'][0].get('finish_reason')


def texto_evidencia(c):
    partes = [f'PREGUNTA: {c["pregunta"]}', '', 'EVIDENCIA:']
    for i, f in enumerate(c['fragmentos'], 1):
        partes += [f'[{i}] {f["manual"]} · {f["seccion"]}', f['texto'], '']
    return '\n'.join(partes).strip()


def correr_decision(url, tipos, evids, esquema_on, pensar):
    reg = []
    for c in tipos:
        out, ms, tm, fin = chat(url, SISTEMA_TIPO, f'Mensaje: {c["pregunta"]}', ESQ_TIPO if esquema_on else None,
                                24 if esquema_on else 64, pensar)
        js, estricto = extraer_json(out)
        pred = js.get('tipo') if isinstance(js, dict) else None
        valido = estricto and pred in TIPOS
        reg.append(dict(id=c['id'], tarea='tipo', esperado=c['esperado'], aceptables=c['aceptables'],
                        pred=pred if pred in TIPOS else None, valido=valido, ms=ms, salida=out[:200],
                        prompt_n=tm.get('prompt_n'), gen_n=tm.get('predicted_n')))
    for c in evids:
        out, ms, tm, fin = chat(url, SISTEMA_EVID, texto_evidencia(c), ESQ_EVID if esquema_on else None,
                                16 if esquema_on else 64, pensar)
        js, estricto = extraer_json(out)
        pred = js.get('suficiente') if isinstance(js, dict) else None
        if isinstance(pred, str):
            pred = pred.lower().replace('í', 'i')
        valido = estricto and pred in ('si', 'no')
        reg.append(dict(id=c['id'], tarea='evidencia', esperado=c['esperado'], aceptables=[c['esperado']],
                        pred=pred if pred in ('si', 'no') else None, valido=valido, ms=ms, salida=out[:200],
                        prompt_n=tm.get('prompt_n'), gen_n=tm.get('predicted_n')))
    return reg


# ---------------------------------------------------------------------------------------------------------------
# Generación: validación por código (lo mismo que haría el quality gate de §6)

NEG = re.compile(r'\*\*(.+?)\*\*')
CITA = re.compile(r'[«“"]([^»”"]{2,60})[»”"]')
ATAJO = re.compile(r'\b(F\d{1,2}|(?:Ctrl|Alt|Shift|Mayús)\s*\+\s*\w+)\b')
MAYUS = re.compile(r'(?<![\w*])([A-ZÁÉÍÓÚÑ][\wáéíóúñü]*(?:\d+)?)')
INOFENSIVOS = {'s10', 'erp', 'metrin'}  # vocabulario del propio prompt, no son menús


def norm(s):
    s = unicodedata.normalize('NFD', (s or '').lower())
    return ' '.join(''.join(ch for ch in s if unicodedata.category(ch) != 'Mn').split())


def terminos_nuevos(texto, ref_norm, ref_palabras):
    """Términos con pinta de interfaz (citas, atajos, palabras con mayúscula a media oración) ausentes del plan."""
    nuevos = []
    plano = texto.replace('**', '')
    for q in CITA.findall(plano):
        if norm(q) not in ref_norm:
            nuevos.append(q)
    for a in ATAJO.findall(plano):
        if norm(a) not in ref_norm:
            nuevos.append(a)
    for m in MAYUS.finditer(plano):
        w, i = m.group(1), m.start()
        prev = plano[:i].rstrip()
        if not prev or prev[-1] in '.!?:;¿¡(«"“' or re.search(r'(^|\s)\d+[.)]$', prev):
            continue  # comienzo de oración o de enumeración
        nw = norm(w)
        if nw in INOFENSIVOS or nw in ref_palabras:
            continue
        nuevos.append(w)
    return nuevos


def validar_generacion(caso, salida_txt):
    plan, r = caso['plan'], caso['restricciones']
    ref = ' '.join([plan['titulo'], plan.get('objetivo', '')] + [s['texto'] for s in plan['pasos']])
    ref_norm = norm(ref)
    ref_palabras = set(re.findall(r'\w+', ref_norm))
    js, _ = extraer_json(salida_txt)
    v = dict(json=False, n_pasos=False, orden=False, fotos=False, negritas=False, negritas_laxas=False,
             sin_nuevos=False, nuevos=[], copia=0.0, ratio=None)
    if not isinstance(js, dict) or not isinstance(js.get('pasos'), list) or not isinstance(js.get('intro', ''), str):
        return v, js
    pasos = js['pasos']
    if not all(isinstance(p, dict) and isinstance(p.get('texto'), str) for p in pasos):
        return v, js
    v['json'] = True
    v['n_pasos'] = len(pasos) == r['n_pasos']
    v['orden'] = v['n_pasos'] and all(p.get('id') == i and p.get('n') == s['n']
                                      for p, i, s in zip(pasos, r['ids'], plan['pasos']))
    v['fotos'] = v['n_pasos'] and all(list(p.get('fotos') or []) == f for p, f in zip(pasos, r['fotos_por_paso']))
    if v['n_pasos']:
        v['negritas'] = all(collections.Counter(' '.join(x.split()) for x in NEG.findall(p['texto'])) ==
                            collections.Counter(' '.join(x.split()) for x in neg)
                            for p, neg in zip(pasos, r['negritas_por_paso']))
        v['negritas_laxas'] = all(set(norm(x) for x in NEG.findall(p['texto'])) == set(norm(x) for x in neg)
                                  for p, neg in zip(pasos, r['negritas_por_paso']))
        v['copia'] = sum(norm(p['texto']) == norm(s['texto']) for p, s in zip(pasos, plan['pasos'])) / len(pasos)
    intro = js.get('intro', '') or ''
    nuevos = terminos_nuevos(intro, ref_norm, ref_palabras)
    if NEG.findall(intro):  # negritas en la intro: solo las del plan
        neg_plan = {norm(x) for n in r['negritas_por_paso'] for x in n}
        nuevos += [x for x in NEG.findall(intro) if norm(x) not in neg_plan]
    for p in pasos:
        nuevos += terminos_nuevos(p['texto'], ref_norm, ref_palabras)
    v['nuevos'] = nuevos
    v['sin_nuevos'] = not nuevos
    largo_in = sum(len(s['texto']) for s in plan['pasos'])
    v['ratio'] = round(sum(len(p['texto']) for p in pasos) / max(1, largo_in), 3)
    v['valida'] = all(v[k] for k in ('json', 'n_pasos', 'orden', 'fotos', 'negritas', 'sin_nuevos'))
    return v, js


SISTEMA_GEN_TEXTOS = SISTEMA_GEN.replace(
    'Responde solo con JSON: {"intro": "...", "pasos": [{"n": 1, "id": "...", "texto": "...", "fotos": ["..."]}]}',
    'Responde solo con JSON: {"intro": "...", "textos": ["texto del paso 1", "texto del paso 2", ...]}, un texto '
    'por paso y en el mismo orden. Los "id" y las "fotos" los pone el sistema: no los escribas.')


SISTEMA_GEN_REESCRIBIR = (
    'Eres el redactor de Metrín, asistente del ERP S10. Recibes un PLAN de respuesta ya validado, en JSON. Reescribe '
    'cada paso como UNA instrucción directa en español neutro (trato de usted, modo imperativo: «Elija…», '
    '«Registre…»), breve y fácil de seguir, y escribe una introducción de una sola oración.\n'
    'Reglas obligatorias:\n'
    '1. Un texto por paso, en el mismo orden. No juntes, no partas, no quites ni añadas pasos.\n'
    '2. Quita rótulos del manual como «Llamada a:» o «Llamada 2-b:», pero conserva la instrucción que traen.\n'
    '3. Copia literalmente cada término entre ** ** (con sus asteriscos) en su mismo paso. No quites ni añadas '
    'negritas.\n'
    '4. No inventes menús, botones, pantallas, atajos de teclado ni datos que no estén en el plan.\n'
    'Responde solo con JSON: {"intro": "...", "textos": ["texto del paso 1", "texto del paso 2", ...]}. Los "id" y '
    'las "fotos" los pone el sistema: no los escribas.')


def esquema_textos(n):
    return {'type': 'object', 'properties': {
        'intro': {'type': 'string'},
        'textos': {'type': 'array', 'items': {'type': 'string'}, 'minItems': n, 'maxItems': n}},
        'required': ['intro', 'textos'], 'additionalProperties': False}


def textos_a_pasos(salida_txt, plan):
    """Variante «solo textos»: el código pone n, id y fotos; el modelo solo redacta."""
    js, _ = extraer_json(salida_txt)
    if not isinstance(js, dict) or not isinstance(js.get('textos'), list):
        return salida_txt
    pasos = [{'n': s['n'], 'id': s['id'], 'texto': t if isinstance(t, str) else '', 'fotos': s['fotos']}
             for s, t in zip(plan['pasos'], js['textos'])]
    if len(js['textos']) != len(plan['pasos']):
        pasos = pasos + [{'n': None, 'id': None, 'texto': '', 'fotos': []}] * max(0, len(js['textos']) - len(pasos))
    return json.dumps({'intro': js.get('intro', ''), 'pasos': pasos}, ensure_ascii=False)


def plan_para_llm(plan):
    return {'titulo': plan['titulo'], 'objetivo': plan.get('objetivo', ''), 'parte': plan.get('parte'),
            'pasos': [{'n': s['n'], 'id': s['id'], 'texto': s['texto'], 'fotos': s['fotos']} for s in plan['pasos']]}


def correr_generacion(url, gens, pensar, variante='completa'):
    reg = []
    for c in gens:
        usuario = 'PLAN:\n' + json.dumps(plan_para_llm(c['plan']), ensure_ascii=False)
        if variante in ('textos', 'reescribir'):
            sis = SISTEMA_GEN_TEXTOS if variante == 'textos' else SISTEMA_GEN_REESCRIBIR
            out, ms, tm, fin = chat(url, sis, usuario, esquema_textos(len(c['plan']['pasos'])), 1800,
                                    pensar)
            crudo = out
            out = textos_a_pasos(out, c['plan'])
        else:
            out, ms, tm, fin = chat(url, SISTEMA_GEN, usuario, ESQ_GEN, 1800, pensar)
        v, _ = validar_generacion(c, out)
        v.setdefault('valida', False)
        pms, gms = tm.get('prompt_ms') or 0, tm.get('predicted_ms') or 0
        reg.append(dict(id=c['id'], tarea='generacion', ms=ms, salida=out, fin=fin, **v,
                        prompt_n=tm.get('prompt_n'), gen_n=tm.get('predicted_n'),
                        prompt_tps=tm.get('prompt_per_second'), gen_tps=tm.get('predicted_per_second'),
                        prompt_ms=pms, gen_ms=gms))
    return reg


# ---------------------------------------------------------------------------------------------------------------
# Métricas

def f1s(regs, clases):
    out = {}
    for k in clases:
        tp = sum(1 for r in regs if r['pred'] == k and r['esperado'] == k)
        fp = sum(1 for r in regs if r['pred'] == k and r['esperado'] != k)
        fn = sum(1 for r in regs if r['pred'] != k and r['esperado'] == k)
        p = tp / (tp + fp) if tp + fp else 0.0
        rc = tp / (tp + fn) if tp + fn else 0.0
        out[k] = round(2 * p * rc / (p + rc), 3) if p + rc else 0.0
    return out


def metricas_decision(regs):
    m = {}
    for tarea, clases in (('tipo', TIPOS), ('evidencia', ['si', 'no'])):
        rs = [r for r in regs if r['tarea'] == tarea]
        if not rs:
            continue
        ms = [r['ms'] for r in rs]
        f = f1s(rs, clases)
        m[tarea] = dict(n=len(rs), exactitud=round(sum(r['pred'] == r['esperado'] for r in rs) / len(rs), 3),
                        exactitud_laxa=round(sum(r['pred'] in r['aceptables'] for r in rs) / len(rs), 3),
                        json_valido=round(sum(r['valido'] for r in rs) / len(rs), 3),
                        f1=f, f1_macro=round(sum(f.values()) / len(f), 3),
                        p50_ms=round(pct(ms, 50)), p95_ms=round(pct(ms, 95)),
                        errores=[(r['id'], r['esperado'], r['pred']) for r in rs if r['pred'] != r['esperado']])
    return m


def metricas_generacion(regs):
    if not regs:
        return {}
    n = len(regs)
    q = lambda k: round(sum(bool(r.get(k)) for r in regs) / n, 3)
    ms = [r['ms'] for r in regs]
    ptps = [r['prompt_tps'] for r in regs if r.get('prompt_tps')]
    gtps = [r['gen_tps'] for r in regs if r.get('gen_tps')]
    return dict(n=n, valida=q('valida'), json=q('json'), n_pasos=q('n_pasos'), orden=q('orden'), fotos=q('fotos'),
                negritas=q('negritas'), negritas_laxas=q('negritas_laxas'), sin_nuevos=q('sin_nuevos'),
                copia_media=round(statistics.mean(r.get('copia') or 0 for r in regs), 3),
                ratio_medio=round(statistics.mean(r['ratio'] for r in regs if r.get('ratio')), 3)
                if any(r.get('ratio') for r in regs) else None,
                p50_ms=round(pct(ms, 50)), p95_ms=round(pct(ms, 95)), max_ms=round(max(ms)),
                prompt_tps=round(statistics.median(ptps), 1) if ptps else None,
                gen_tps=round(statistics.median(gtps), 1) if gtps else None,
                tokens_salida_p50=round(pct([r['gen_n'] or 0 for r in regs], 50)),
                tokens_entrada_p50=round(pct([r['prompt_n'] or 0 for r in regs], 50)),
                truncadas=sum(r.get('fin') == 'length' for r in regs),
                nuevos_ejemplos=sorted({t for r in regs for t in r.get('nuevos', [])})[:40])


def guardar(nombre, datos):
    os.makedirs(DIR_RES, exist_ok=True)
    ruta = os.path.join(DIR_RES, f'{nombre}.json')
    with open(ruta, 'w') as f:
        json.dump(datos, f, ensure_ascii=False, indent=1)
    print('→', os.path.relpath(ruta, RAIZ))


# ---------------------------------------------------------------------------------------------------------------
def uso_gpu():
    """«Device Utilization %» del acelerador (ioreg, sin sudo). Sirve para no medir latencias con la GPU ocupada."""
    out = subprocess.run(['ioreg', '-r', '-d', '1', '-c', 'IOAccelerator'], capture_output=True, text=True).stdout
    m = re.search(r'"Device Utilization %"=(\d+)', out)
    return int(m.group(1)) if m else None


def esperar_gpu_libre(umbral=15, seguidas=3, max_s=900):
    t0, ok, hist = time.time(), 0, []
    while time.time() - t0 < max_s:
        u = uso_gpu()
        hist.append(u)
        ok = ok + 1 if (u is not None and u < umbral) else 0
        if ok >= seguidas:
            return True, hist[-seguidas:]
        time.sleep(5)
    return False, hist[-5:]


def cmd_correr(a):
    cfg = MODELOS[a.modelo]
    tipos, evids, gens = casos()
    if a.modo == 'cpu4':
        evids = [c for c in evids if c['id'] in SUB_CPU_EVID]
        gens = [c for c in gens if c['id'] in SUB_CPU_GEN]
    if a.solo_decision:
        gens = []
    if a.solo_generacion:
        tipos, evids = [], []
    if a.latencia:  # pasada corta solo para latencia, con la GPU libre (las de calidad se miden en la completa)
        tipos = tipos[::3]
        evids = [c for c in evids if c['id'] in SUB_CPU_EVID[:5] + SUB_CPU_EVID[10:15]]
        gens = [c for c in gens if c['id'] in SUB_CPU_GEN]
    if a.limite:  # prueba de humo
        tipos, evids, gens = tipos[:a.limite], evids[:a.limite], gens[:max(1, a.limite // 3)]
    extra = ['--reasoning', 'off'] if cfg['pensar'] else []
    os.makedirs(DIR_RES, exist_ok=True)
    log = os.path.join(DIR_RES, f'{a.modelo}__{a.modo}.log')
    res = dict(modelo=a.modelo, modo=a.modo, gguf=os.path.basename(cfg['gguf']), origen=cfg['hf'],
               licencia=cfg['licencia'], tam_mb=round(os.path.getsize(cfg['gguf']) / 2 ** 20),
               fecha=time.strftime('%Y-%m-%d %H:%M'), args=None)
    if a.latencia and a.modo == 'metal':
        libre, hist = esperar_gpu_libre()
        res['gpu_antes'] = dict(libre=libre, uso=hist)
        print('  GPU libre:', libre, hist)
    with Servidor(LLAMA_SERVER, cfg['gguf'], a.modo, a.puerto, log, extra=extra) as s:
        res['args'] = ' '.join(s.args[1:]).replace(os.path.expanduser('~'), '~')
        res['carga_s'] = round(s.carga_s, 1)
        res['rss_carga_mb'] = round(s.rss_carga or 0)
        print(f'[{a.modelo} {a.modo}] cargado en {s.carga_s:.1f} s, RSS {s.rss_carga:.0f} MB, '
              f'huella {s.huella_carga or 0:.0f} MB')
        chat(s.url, SISTEMA_TIPO, 'Mensaje: hola', ESQ_TIPO, 8, cfg['pensar'])  # calentamiento
        t0 = time.time()
        res['variante_generacion'] = a.variante
        dec = correr_decision(s.url, tipos, evids, True, cfg['pensar'])
        res['decision'] = metricas_decision(dec)
        print(f'  decisión (json_schema) {time.time() - t0:.0f} s:',
              {k: (v['exactitud'], v['json_valido']) for k, v in res['decision'].items()})
        res['rss_decision_mb'] = round(s.m.pico)
        res['huella_decision_mb'] = round(s.m.pico_huella)
        if a.modo == 'metal' and tipos and not a.latencia:
            t0 = time.time()
            lib = correr_decision(s.url, tipos, evids, False, cfg['pensar'])
            res['decision_libre'] = metricas_decision(lib)
            print(f'  decisión (libre) {time.time() - t0:.0f} s:',
                  {k: (v['exactitud'], v['json_valido']) for k, v in res['decision_libre'].items()})
        else:
            lib = []
        gen = []
        if gens:
            t0 = time.time()
            gen = correr_generacion(s.url, gens, cfg['pensar'], a.variante)
            res['generacion'] = metricas_generacion(gen)
            g = res['generacion']
            print(f'  generación {time.time() - t0:.0f} s: válida {g["valida"]}, p50 {g["p50_ms"]} ms, '
                  f'{g["prompt_tps"]} / {g["gen_tps"]} tok/s')
        res['rss_pico_mb'] = round(s.m.pico)
        res['huella_carga_mb'] = round(s.huella_carga or 0)
        res['huella_pico_mb'] = round(s.m.pico_huella)
    res['registros'] = dict(decision=dec, decision_libre=lib, generacion=gen)
    guardar(f'{a.modelo}__{a.modo}{a.sufijo}', res)


# ---------------------------------------------------------------------------------------------------------------
# Jev: API /v1/systemone (choice / noul desde los logits, sin generar)

def jev_decidir(url, tipos, evids, modelo):
    reg = []
    crit = {k: v for k, v in DEF_TIPOS.items()}
    for c in tipos:
        cuerpo = {'state': {'mensaje': c['pregunta'], 'contexto': 'usuario del ERP S10 (construcción)'},
                  'questions': {'q': {'type': 'choice',
                                      'instructions': '¿Qué tipo de respuesta necesita este mensaje?',
                                      'criteria': crit}}}
        if modelo:
            cuerpo['model'] = modelo
        d, ms = post(url + '/v1/systemone', cuerpo)
        ans = d['answers']['q']
        reg.append(dict(id=c['id'], tarea='tipo', esperado=c['esperado'], aceptables=c['aceptables'],
                        pred=ans.get('choice'), valido=ans.get('choice') in TIPOS, ms=ms,
                        conf=ans.get('confidence'), prompt_n=d.get('usage', {}).get('input_tokens')))
    for c in evids:
        cuerpo = {'state': {'pregunta': c['pregunta'],
                            'evidencia': [f'{f["manual"]} · {f["seccion"]}\n{f["texto"]}' for f in c['fragmentos']]},
                  'questions': {'q': {'type': 'noul',
                                      'instructions': '¿Algún fragmento de la evidencia contiene directamente lo que '
                                                      'la pregunta pide (definición, pasos, lugar, causa o dato)? '
                                                      'Que solo trate el mismo tema no basta.'}}}
        if modelo:
            cuerpo['model'] = modelo
        d, ms = post(url + '/v1/systemone', cuerpo)
        p = d['answers']['q'].get('noul')
        reg.append(dict(id=c['id'], tarea='evidencia', esperado=c['esperado'], aceptables=[c['esperado']],
                        pred=('si' if p >= 0.5 else 'no') if p is not None else None, valido=p is not None,
                        ms=ms, noul=p, prompt_n=d.get('usage', {}).get('input_tokens')))
    return reg


def cmd_jev(a):
    tipos, evids, _ = casos()
    if a.modo == 'cpu4':
        evids = [c for c in evids if c['id'] in SUB_CPU_EVID]
    res = dict(modelo=a.nombre, modo=a.modo, fecha=time.strftime('%Y-%m-%d %H:%M'))
    if a.url:  # servidor ya en marcha (MLX): no se toca, solo se mide su RSS
        pid = subprocess.run(['pgrep', '-f', 'jev-style serve'], capture_output=True, text=True).stdout.split()
        m = Muestreador(int(pid[0])) if pid else None
        if m:
            m.start()
        res['rss_carga_mb'] = round(rss_mb(int(pid[0])) or 0) if pid else None
        post(a.url + '/v1/systemone', {'state': 'hola', 'questions': {'q': {'type': 'noul', 'instructions': '¿Es un saludo?'}}})
        reg = jev_decidir(a.url, tipos, evids, a.jev_modelo)
        if m:
            m.parar.set()
            res['rss_pico_mb'] = round(m.pico)
            res['huella_pico_mb'] = round(m.pico_huella)
    else:
        log = os.path.join(DIR_RES, f'{a.nombre}__{a.modo}.log')
        os.makedirs(DIR_RES, exist_ok=True)
        extra = a.extra.split() if a.extra else []
        with Servidor(JEVA_SERVER, a.gguf, a.modo, a.puerto, log, extra=extra) as s:
            res['args'] = ' '.join(s.args[1:]).replace(os.path.expanduser('~'), '~')
            res['carga_s'] = round(s.carga_s, 1)
            res['rss_carga_mb'] = round(s.rss_carga or 0)
            post(s.url + '/v1/systemone', {'state': 'hola', 'questions': {'q': {'type': 'noul', 'instructions': '¿Es un saludo?'}}})
            reg = jev_decidir(s.url, tipos, evids, None)
            res['rss_pico_mb'] = round(s.m.pico)
            res['huella_pico_mb'] = round(s.m.pico_huella)
    res['decision'] = metricas_decision(reg)
    print({k: (v['exactitud'], v['p50_ms'], v['p95_ms']) for k, v in res['decision'].items()})
    res['registros'] = dict(decision=reg)
    guardar(f'{a.nombre}__{a.modo}', res)


# ---------------------------------------------------------------------------------------------------------------
# Puntuación de calidad a ciegas: se exporta una muestra sin nombres de modelo; la clave va aparte.

RUBRICA = """Rúbrica (1–5), se puntúa la salida entera (intro + pasos) frente al plan de entrada:
5 español neutro y natural, gramática correcta, imperativo de usted consistente, breve y fiel al plan.
4 fiel y correcto, con un detalle menor (una frase algo torpe, una redundancia, un cambio tú/usted aislado).
3 se entiende pero suena forzado o repetitivo, o un paso pierde un matiz sin llegar a cambiar la instrucción.
2 errores gramaticales, palabras en inglés o regionalismos, o un paso cambia de sentido.
1 inservible: otro idioma, texto roto, pasos fundidos o perdidos, contenido inventado."""


def cmd_ciego(a):
    _, _, gens = casos()
    rnd = random.Random(20261006)
    # Muestra estratificada: mitad planes de procedimiento (texto ya pulido), mitad de respaldo desde fragmentos
    # (texto crudo del manual, donde reformular sí cambia algo).
    proc = [g['id'] for g in gens if g.get('origen') != 'fragmento']
    frag = [g['id'] for g in gens if g.get('origen') == 'fragmento']
    ids = sorted(rnd.sample(proc, a.n // 2) + rnd.sample(frag, a.n - a.n // 2))
    salidas = collections.defaultdict(dict)
    rutas = sorted(glob.glob(os.path.join(DIR_RES, '*__metal.json')) +
                   glob.glob(os.path.join(DIR_RES, '*__metal__reescribir.json')))
    for ruta in rutas:
        d = json.load(open(ruta))
        etiqueta = os.path.basename(ruta)[:-5].replace('__metal', '')  # modelo o modelo__reescribir
        for r in d.get('registros', {}).get('generacion', []):
            if r['id'] in ids:
                salidas[r['id']][etiqueta] = r['salida']
    os.makedirs(DIR_CIEGO, exist_ok=True)
    clave, lineas = {}, ['# Muestra a ciegas para puntuar la calidad del español', '', RUBRICA, '']
    porid = {g['id']: g for g in gens}
    for gid in ids:
        mods = sorted(salidas[gid])
        rnd.shuffle(mods)
        clave[gid] = {}
        lineas += [f'## {gid}', '', 'PLAN (entrada):', '']
        lineas += [f'{s["n"]}. {s["texto"]}' for s in porid[gid]['plan']['pasos']] + ['']
        for i, m in enumerate(mods):
            et = 'ABCDEFGHIJKLMNOP'[i]
            clave[gid][et] = m
            js, _ = extraer_json(salidas[gid][m])
            if isinstance(js, dict) and isinstance(js.get('pasos'), list):
                txt = [f'intro: {js.get("intro", "")}'] + [f'{p.get("n")}. {p.get("texto")}' for p in js['pasos']
                                                           if isinstance(p, dict)]
            else:
                txt = ['(no es JSON) ' + salidas[gid][m][:600]]
            lineas += [f'### {gid} · {et}', ''] + txt + ['']
    open(os.path.join(DIR_CIEGO, 'muestra.md'), 'w').write('\n'.join(lineas))
    json.dump(clave, open(os.path.join(DIR_CIEGO, 'clave.json'), 'w'), indent=1)
    print('→ muestra.md y clave.json en', os.path.relpath(DIR_CIEGO, RAIZ), '(puntajes en puntajes.json)')


# ---------------------------------------------------------------------------------------------------------------
def cmd_resumen(a):
    filas = []
    for ruta in sorted(glob.glob(os.path.join(DIR_RES, '*.json'))):
        d = json.load(open(ruta))
        d['_archivo'] = os.path.basename(ruta)[:-5]
        filas.append(d)
    calidad = {}
    pj, cl = os.path.join(DIR_CIEGO, 'puntajes.json'), os.path.join(DIR_CIEGO, 'clave.json')
    if os.path.exists(pj) and os.path.exists(cl):
        p, c = json.load(open(pj)), json.load(open(cl))
        acc = collections.defaultdict(list)
        for gid, et in p.items():
            for e, nota in et.items():
                acc[c[gid][e]].append(nota)
        calidad = {m: (round(statistics.mean(v), 2), len(v)) for m, v in acc.items()}
    print('| modelo | modo | tipo exact. (laxa) | tipo F1 macro | evid. exact. | JSON válido (esquema / libre) '
          '| decisión p50/p95 ms (tipo) | evid. p50/p95 ms | RSS / huella pico MB |')
    print('|---|---|---|---|---|---|---|---|---|')
    for d in filas:
        de, li = d.get('decision', {}), d.get('decision_libre', {})
        t, e = de.get('tipo', {}), de.get('evidencia', {})
        jl = '/'.join(str(li[k]['json_valido']) for k in ('tipo', 'evidencia') if k in li) or '–'
        if not t and not e:
            continue
        print(f"| {d['_archivo']} | {d['modo']} | {t.get('exactitud')} ({t.get('exactitud_laxa')}) | "
              f"{t.get('f1_macro')} | {e.get('exactitud')} | {t.get('json_valido')} / {jl} | "
              f"{t.get('p50_ms')}/{t.get('p95_ms')} | {e.get('p50_ms')}/{e.get('p95_ms')} | "
              f"{d.get('rss_pico_mb')} / {d.get('huella_pico_mb')} |")
    print()
    print('| modelo | modo | n | válidas | orden | fotos | negritas | sin nuevos | copia | p50/p95 ms '
          '| prompt tok/s | gen tok/s | calidad (n) |')
    print('|---|---|---|---|---|---|---|---|---|---|---|---|---|')
    for d in filas:
        g = d.get('generacion')
        if not g:
            continue
        q = calidad.get(d['_archivo'].replace('__metal', ''))
        print(f"| {d['_archivo']} | {d['modo']} | {g['n']} | {g['valida']} | {g['orden']} | {g['fotos']} | "
              f"{g['negritas']} | {g['sin_nuevos']} | {g['copia_media']} | {g['p50_ms']}/{g['p95_ms']} | "
              f"{g['prompt_tps']} | {g['gen_tps']} | {q[0] if q else '–'} ({q[1] if q else ''}) |")
    print()
    for d in filas:
        t = d.get('decision', {}).get('tipo')
        if t:
            print(d['_archivo'], 'F1 por clase:', t['f1'])


def cmd_velocidad(a):
    """Throughput puro con llama-bench (pp = prompt, tg = generación, tokens/s). Con los tokens medios de cada tarea
    sale una latencia estimada que no depende de la carga del momento en que corrieron las tareas."""
    cfg = MODELOS[a.modelo] if a.modelo in MODELOS else dict(gguf=a.gguf)
    args = ['/opt/homebrew/bin/llama-bench', '-m', cfg['gguf'], '-p', str(a.pp), '-n', str(a.tg), '-r', str(a.r),
            '-o', 'json']
    args += ['-ngl', '0', '-dev', 'none', '-t', '4'] if a.modo == 'cpu4' else ['-ngl', '99']
    carga = os.getloadavg()[0]
    out = subprocess.run(args, capture_output=True, text=True, timeout=3600).stdout
    filas = json.loads(out)
    res = dict(modelo=a.modelo, modo=a.modo, carga_1min=round(carga, 1), fecha=time.strftime('%Y-%m-%d %H:%M'),
               args=' '.join(args[1:]).replace(os.path.expanduser('~'), '~'))
    for f in filas:
        clave = 'pp_tps' if f.get('n_prompt') else 'tg_tps'
        res[clave] = round(f['avg_ts'], 1)
        res[clave + '_sd'] = round(f.get('stddev_ts', 0), 1)
    print(res)
    os.makedirs(DIR_RES, exist_ok=True)
    with open(os.path.join(DIR_RES, 'velocidad.jsonl'), 'a') as f:
        f.write(json.dumps(res, ensure_ascii=False) + '\n')


def cmd_modelos(a):
    for k, v in MODELOS.items():
        ok = os.path.exists(v['gguf'])
        print(f"{k:20s} {'OK ' if ok else 'FALTA'} {round(os.path.getsize(v['gguf']) / 2 ** 20) if ok else '-':>6} MB  "
              f"{v['licencia']:10s} {v['hf']}")


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sp = ap.add_subparsers(dest='cmd', required=True)
    c = sp.add_parser('correr')
    c.add_argument('--modelo', required=True, choices=sorted(MODELOS))
    c.add_argument('--modo', choices=['metal', 'cpu4'], default='metal')
    c.add_argument('--puerto', type=int, default=18431)
    c.add_argument('--solo-decision', action='store_true')
    c.add_argument('--limite', type=int, default=0, help='prueba de humo: solo los N primeros casos')
    c.add_argument('--sufijo', default='', help='sufijo del archivo de resultados')
    c.add_argument('--solo-generacion', action='store_true')
    c.add_argument('--latencia', action='store_true',
                   help='pasada corta de latencia: espera a que la GPU esté libre y mide un subconjunto')
    c.add_argument('--variante', choices=['completa', 'textos', 'reescribir'], default='completa',
                   help='generación: JSON completo (n, id, texto, fotos) o solo los textos (el código pone el resto)')
    j = sp.add_parser('jev')
    j.add_argument('--url', help='servidor Jev ya en marcha (p. ej. http://127.0.0.1:8765)')
    j.add_argument('--jev-modelo', default=None)
    j.add_argument('--gguf', help='GGUF para levantar con jeva.cpp')
    j.add_argument('--modo', choices=['metal', 'cpu4', 'mlx'], default='mlx')
    j.add_argument('--nombre', required=True)
    j.add_argument('--puerto', type=int, default=18432)
    j.add_argument('--extra', default='', help='argumentos extra para llama-server de jeva.cpp')
    ci = sp.add_parser('ciego')
    ci.add_argument('--n', type=int, default=8)
    sp.add_parser('resumen')
    sp.add_parser('modelos')
    v = sp.add_parser('velocidad')
    v.add_argument('--modelo', required=True)
    v.add_argument('--gguf', default=None, help='GGUF fuera del registro (p. ej. Jev)')
    v.add_argument('--modo', choices=['metal', 'cpu4'], default='cpu4')
    v.add_argument('--pp', type=int, default=512)
    v.add_argument('--tg', type=int, default=128)
    v.add_argument('-r', type=int, default=2)
    a = ap.parse_args()
    dict(correr=cmd_correr, jev=cmd_jev, ciego=cmd_ciego, resumen=cmd_resumen, modelos=cmd_modelos,
         velocidad=cmd_velocidad)[a.cmd](a)


if __name__ == '__main__':
    main()
