#!/usr/bin/env python3
"""Convierte los tickets de Zendesk exportados (optimiza360_scraping/) en candidatas para el router, anonimizadas.

Uso (desde entrenamiento-router/):
    .venv/bin/python extraer_tickets.py [--tickets ../../optimiza360_scraping/tickets_optimiza360_20261007.json]

Escribe reales/tickets/candidatas.jsonl (fuera de git: son datos de clientes). Una candidata por ticket: el asunto y
el primer mensaje de quien lo abre (la descripción), sin firmas, avisos de correo, adjuntos incrustados, enlaces ni
hilos citados, y con el mismo anonimizado que extraer_reales.py más los nombres de solicitante y asignado. El módulo
S10 y el subtipo que puso la mesa de ayuda van aparte como pista para quien etiqueta; el router nunca los ve.
"""
from __future__ import annotations

import argparse
import json
import re
from pathlib import Path

from extraer_reales import anonimizar

AQUI = Path(__file__).resolve().parent
SALIDA = AQUI / "reales" / "tickets"
DEFECTO = AQUI.parent.parent / "optimiza360_scraping" / "tickets_optimiza360_20261007.json"

CORTES = re.compile(
    r"^\s*(saludos|atentamente|cordialmente|gracias|muchas gracias|slds|un abrazo|best regards|regards)\b"
    r"|^\s*(de|from|enviado|sent|para|to|asunto|subject)\s*:"
    r"|^\s*-{2,}|^\s*_{3,}|^\s*el .{3,80} escribi[oó]:",
    re.I)


def limpiar(texto: str) -> str:
    out = []
    for linea in (texto or "").splitlines():
        if CORTES.search(linea):
            break
        if linea.strip().startswith(">"):
            continue
        out.append(linea)
    t = "\n".join(out)
    t = re.sub(r"\[cid:[^\]]*\]", " ", t)
    t = re.sub(r"\[(https?://|mailto:)[^\]]*\]", " ", t)
    t = re.sub(r"<(https?://|mailto:)[^>]*>", " ", t)
    t = re.sub(r"(?i)this (e-?mail|message) .*$", " ", t, flags=re.S)
    t = re.sub(r"(?i)(este (correo|mensaje)|el contenido de este).{0,40}(confidencial|destinatario).*$", " ", t,
               flags=re.S)
    return t


def nombres_de(f: dict) -> list[str]:
    ns = []
    for k in ("requester", "assignee"):
        v = (f.get(k) or "").strip()
        if v and "@" not in v:
            ns.append(v)
    return ns


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--tickets", type=Path, default=DEFECTO)
    a = ap.parse_args()
    filas = json.load(open(a.tickets, encoding="utf-8"))["filas"]
    SALIDA.mkdir(parents=True, exist_ok=True)
    n = 0
    with open(SALIDA / "candidatas.jsonl", "w", encoding="utf-8") as f:
        for x in sorted(filas, key=lambda r: r["created_at"]):
            nombres = nombres_de(x)
            asunto = anonimizar(x.get("subject") or "", nombres)
            cuerpo = anonimizar(limpiar(x.get("description") or ""), nombres)
            texto = (asunto + ". " + cuerpo).strip(" .")
            palabras = texto.split()
            if len(palabras) < 3:
                continue
            texto = " ".join(palabras[:120])
            f.write(json.dumps({
                "id": f"zendesk-{x['id']}", "grupo": "zendesk", "t": x["created_at"].rstrip("Z"), "texto": texto,
                "contexto": [], "pista": {"modulo_s10": x.get("cf_modulo_s10"), "subtipo": x.get("cf_sub_tipo_ticket"),
                                          "categoria": x.get("cf_categorizacion"), "division": x.get("cf_division_ti")},
            }, ensure_ascii=False) + "\n")
            n += 1
    print(json.dumps({"tickets": len(filas), "candidatas": n, "salida": str(SALIDA / "candidatas.jsonl")},
                     ensure_ascii=False))


if __name__ == "__main__":
    main()
