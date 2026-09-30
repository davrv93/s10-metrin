#!/usr/bin/env python3
"""
Asocia a cada pregunta del banco (data/preguntas/preguntas.jsonl) una respuesta de Metrín basada
en su nodo de Cortex, y arma un dataset LoRA en formato MLX-LM (`messages`).

  .venv/bin/python herramientas/generar_respuestas.py --hilos 20          # genera (reanudable)
  .venv/bin/python herramientas/generar_respuestas.py --armar             # solo arma train/valid/test

Diseño (RAG consciente): el mensaje del usuario lleva el PASAJE de Cortex que respalda la respuesta,
igual que el RAG de Metrín en producción. El adaptador aprende a responder desde el contexto dado
(con tono Metrín y citando la fuente), no a memorizar cifras de S10. Un 5 % de ejemplos lleva un
pasaje que NO contiene la respuesta y enseña a decirlo sin inventar.

Por cada pregunta base el modelo devuelve: respuesta + 1-3 citas textuales del nodo. Las citas se
verifican contra el nodo (deben aparecer tal cual, salvo espacios); si no aparecen, se descarta.

Salida en data/lora/:
  crudo/                     respuestas del modelo por tanda (reanudable)
  respuestas.jsonl           pregunta base → respuesta, pasaje, nodo
  train.jsonl valid.jsonl test.jsonl   formato MLX-LM {"messages":[system,user,assistant]}
                             split por pregunta base (sus 4 escrituras caen en el mismo split)
  resumen.json               conteos, descartes y sha256 de cada archivo
"""
from __future__ import annotations

import argparse
import hashlib
import json
import random
import re
import subprocess
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import generar_preguntas as gp  # noqa: E402  mismos nodos y normalización que el banco

RAIZ = gp.RAIZ
PREGUNTAS = RAIZ / "data" / "preguntas" / "preguntas.jsonl"
SALIDA = RAIZ / "data" / "lora"
CRUDO = SALIDA / "crudo"
POR_LLAMADA = 25
SISTEMA_ENTRENO = ("Eres Metrín, asistente de Optimiza 360. Respondes en español neutro, con calidez, claridad y "
                   "profesionalismo. Respondes solo con el contexto que recibes y citas su fuente; si el contexto no "
                   "trae la respuesta, lo dices y sugieres a quién acudir. No inventas datos ni afirmas acciones que no realizaste.")
SEMILLA = 20260930


def log(*a):
    print(time.strftime("%H:%M:%S"), *a, flush=True)


SISTEMA_GEN = """Eres Metrín, el asistente de obra de Optimiza 360 para el ERP S10: cercano, claro, profesional, en español neutro.
Respondes preguntas de usuarios USANDO SOLO el contenido del nodo de conocimiento que se te da. Reglas:
- Nada que no esté en el nodo. Si el nodo no alcanza para responder bien, dilo en una frase y sugiere soporte S10 u Optimiza 360.
- Respuestas de 40 a 160 palabras. Pasos numerados cuando es un procedimiento. Sin saludos largos ni relleno.
- Si el nodo trae cifras legales o de un año concreto (UIT, RMV, jornal), menciona el año y pide verificar la cifra vigente.
- Menciona la fuente de forma natural al final («Según la guía de S10 Presupuestos…», «En el video oficial de S10…»), sin inventar páginas.
- No hables del «nodo», de «Cortex» ni de «la base de conocimiento».
Por cada pregunta devuelve también 1 a 3 CITAS copiadas TEXTUALMENTE del contenido del nodo (frases o líneas completas, sin cambiar
ni una palabra) que respaldan tu respuesta.
Devuelve SOLO JSON: {"respuestas":[{"id":"...","respuesta":"...","citas":["...","..."],"alcanza":true}]}"""


def _json(t: str):
    t = re.sub(r"```(?:json)?", "", t)
    return json.loads(t[t.find("{"):t.rfind("}") + 1])


def llamar(usuario: str) -> dict:
    import os
    if os.environ.get("ANTHROPIC_API_KEY"):
        import anthropic
        c = anthropic.Anthropic()
        with c.messages.stream(model="claude-opus-5", max_tokens=32000, system=SISTEMA_GEN,
                               messages=[{"role": "user", "content": usuario}], output_config={"effort": "low"}) as s:
            r = s.get_final_message()
        return _json("".join(b.text for b in r.content if b.type == "text"))
    r = subprocess.run(["claude", "-p", "--output-format", "json", "--append-system-prompt", SISTEMA_GEN,
                        "--disallowedTools", "Bash,Edit,Write,Read,Glob,Grep,WebFetch,WebSearch,Task,NotebookEdit"],
                       input=usuario, capture_output=True, text=True, cwd="/tmp", timeout=1200)
    d = json.loads(r.stdout)
    if d.get("is_error"):
        raise RuntimeError(d.get("result", "")[:200])
    return _json(d["result"])


# ─────────────────────────────── generar ───────────────────────────────────
def tandas() -> list[tuple]:
    nodos = {n["path"]: n for n in gp.nodos()}
    por_nodo: dict[str, list] = {}
    for l in PREGUNTAS.read_text(encoding="utf-8").splitlines():
        q = json.loads(l)
        por_nodo.setdefault(q["nodo"], []).append(q)
    out = []
    for p, qs in por_nodo.items():
        for k in range(0, len(qs), POR_LLAMADA):
            out.append((nodos[p], qs[k:k + POR_LLAMADA], k // POR_LLAMADA))
    return out


def archivo(nodo: dict, k: int) -> Path:
    return CRUDO / f"{(nodo['path'] or 'raiz').replace('/', '__')}__{k}.json"


def correr(nodo: dict, qs: list[dict], k: int) -> int:
    dest = archivo(nodo, k)
    if dest.exists():
        return len(json.loads(dest.read_text())["respuestas"])
    usuario = (f"Nodo: {nodo['titulo']}\n\nContenido del nodo:\n<<<\n{nodo['cuerpo']}\n>>>\n\n"
               "Preguntas (responde todas, con su id):\n" + "\n".join(f"- id={q['id']}: {q['pregunta']}" for q in qs))
    for intento in range(3):
        try:
            d = llamar(usuario)
            rs = [r for r in d.get("respuestas", []) if isinstance(r, dict) and r.get("id") and r.get("respuesta")]
            dest.write_text(json.dumps({"nodo": nodo["path"], "respuestas": rs}, ensure_ascii=False))
            return len(rs)
        except Exception as e:
            log(f"  reintento {intento + 1} {nodo['path']}#{k}: {e}")
            time.sleep(15 * (intento + 1))
    return 0


# ─────────────────────────────── armar ─────────────────────────────────────
def _plano(t: str) -> str:
    return re.sub(r"\s+", " ", t).strip().lower()


def verificar_citas(citas: list, cuerpo: str) -> list[str]:
    base = _plano(cuerpo)
    ok = []
    for c in citas or []:
        c = str(c).strip().strip("«»\"'")
        if len(c) >= 12 and _plano(c) in base:
            ok.append(c)
    return ok


NO_ALCANZA = [
    "Revisé lo que tengo a mano y no encuentro eso en las fuentes de este tema, así que prefiero no adivinar. "
    "Si me das el nombre del módulo o de la pantalla, lo busco de nuevo; y si es urgente, el soporte de S10 u Optimiza 360 te pueden orientar.",
    "Con la información que recibí no puedo responder eso con seguridad: habla de otro tema. "
    "¿Me cuentas en qué módulo de S10 estás trabajando? Así busco en el lugar correcto.",
    "Esta vez no tengo evidencia para responderte: el material que encontré no trata ese punto. "
    "Prueba reformulando con el nombre del proceso, o consúltalo con la mesa de ayuda de Optimiza 360.",
    "No quiero darte un dato inventado: lo que tengo a la vista no cubre tu pregunta. "
    "Si me indicas el módulo (Presupuestos, Nóminas, Almacenes…) vuelvo a buscar.",
]


def contexto(citas: list[str], titulo: str) -> str:
    return f"Contexto ({titulo}):\n" + "\n".join(f"- {c}" for c in citas)


def ejemplo(pregunta: str, ctx: str, respuesta: str) -> dict:
    return {"messages": [{"role": "system", "content": SISTEMA_ENTRENO},
                         {"role": "user", "content": f"{ctx}\n\nPregunta: {pregunta}"},
                         {"role": "assistant", "content": respuesta}]}


def armar() -> None:
    nodos = {n["path"]: n for n in gp.nodos()}
    preguntas = {json.loads(l)["id"]: json.loads(l) for l in PREGUNTAS.read_text(encoding="utf-8").splitlines()}
    respuestas, desc = {}, {"sin_respuesta": 0, "citas_no_verificadas": 0, "meta": 0}
    for f in sorted(CRUDO.glob("*.json")):
        d = json.loads(f.read_text())
        cuerpo = nodos[d["nodo"]]["cuerpo"] if d["nodo"] in nodos else ""
        for r in d["respuestas"]:
            if r["id"] not in preguntas:
                continue
            if gp.META.search(r["respuesta"]) or re.search(r"\bCortex\b", r["respuesta"]):
                desc["meta"] += 1
                continue
            citas = verificar_citas(r.get("citas"), cuerpo)
            if not citas:
                desc["citas_no_verificadas"] += 1
                continue
            respuestas[r["id"]] = {"id": r["id"], "nodo": d["nodo"], "titulo": nodos[d["nodo"]]["titulo"],
                                   "respuesta": r["respuesta"].strip(), "citas": citas, "alcanza": r.get("alcanza", True)}
    desc["sin_respuesta"] = len(preguntas) - len(respuestas) - desc["meta"] - desc["citas_no_verificadas"]
    SALIDA.mkdir(parents=True, exist_ok=True)
    (SALIDA / "respuestas.jsonl").write_text("".join(json.dumps({**v, "pregunta": preguntas[k]["pregunta"]}, ensure_ascii=False) + "\n"
                                                     for k, v in respuestas.items()), encoding="utf-8")

    rnd = random.Random(SEMILLA)
    ids = sorted(respuestas)
    rnd.shuffle(ids)
    n = len(ids)
    corte = {"test": set(ids[:int(n * .05)]), "valid": set(ids[int(n * .05):int(n * .10)])}
    split = lambda i: "test" if i in corte["test"] else "valid" if i in corte["valid"] else "train"  # noqa: E731
    salida = {"train": [], "valid": [], "test": []}
    por_nodo = {}
    for i in ids:
        por_nodo.setdefault(respuestas[i]["nodo"], []).append(i)
    nodos_lista = sorted(por_nodo)
    abst = 0
    for i in ids:
        r, q, s = respuestas[i], preguntas[i], split(i)
        ctx = contexto(r["citas"], r["titulo"])
        for texto in [q["pregunta"]] + [v["texto"] for v in q["variantes"]]:
            salida[s].append(ejemplo(texto, ctx, r["respuesta"]))
        # 5 %: el mismo pedido con el pasaje de OTRO tema → no inventa
        if rnd.random() < 0.05:
            otro = rnd.choice([p for p in nodos_lista if p.split("/")[0] != r["nodo"].split("/")[0]] or nodos_lista)
            o = respuestas[rnd.choice(por_nodo[otro])]
            salida[s].append(ejemplo(rnd.choice([q["pregunta"]] + [v["texto"] for v in q["variantes"]]),
                                     contexto(o["citas"], o["titulo"]), rnd.choice(NO_ALCANZA)))
            abst += 1
    resumen = {"preguntas_con_respuesta": len(respuestas), "descartes": desc, "ejemplos_sin_evidencia": abst,
               "sistema": SISTEMA_ENTRENO, "semilla": SEMILLA, "archivos": {}}
    for s, filas in salida.items():
        rnd.shuffle(filas)
        p = SALIDA / f"{s}.jsonl"
        p.write_text("".join(json.dumps(f, ensure_ascii=False) + "\n" for f in filas), encoding="utf-8")
        resumen["archivos"][s] = {"ejemplos": len(filas), "sha256": hashlib.sha256(p.read_bytes()).hexdigest()}
    (SALIDA / "resumen.json").write_text(json.dumps(resumen, indent=1, ensure_ascii=False), encoding="utf-8")
    log(f"armado: {len(respuestas)} respuestas · " + " · ".join(f"{s} {v['ejemplos']}" for s, v in resumen["archivos"].items())
        + f" · descartes {desc}")


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--hilos", type=int, default=20)
    ap.add_argument("--armar", action="store_true")
    ap.add_argument("--solo", help="limita a nodos cuyo path empiece así (pruebas)")
    a = ap.parse_args()
    CRUDO.mkdir(parents=True, exist_ok=True)
    if not a.armar:
        ts = [t for t in tandas() if not a.solo or t[0]["path"].startswith(a.solo)]
        pend = [t for t in ts if not archivo(t[0], t[2]).exists()]
        log(f"{len(ts)} tandas ({len(pend)} pendientes), {a.hilos} llamadas a la vez")
        hechos = 0
        with ThreadPoolExecutor(a.hilos) as ex:
            fut = {ex.submit(correr, *t): f"{t[0]['path'] or '(raíz)'}#{t[2]}" for t in pend}
            for f in as_completed(fut):
                hechos += 1
                log(f"[{hechos}/{len(fut)}] {fut[f]}: {f.result()} respuestas")
    armar()


if __name__ == "__main__":
    main()
