#!/usr/bin/env python3
"""Pasa la prueba REAL (reales/prueba_real.jsonl, mensajes de clientes, fuera de git) por la V2 completa de los dos
contenedores de medir-v2.sh (4763 sin router, 4764 con router) y cuenta qué contesta Metrín.

    .venv/bin/python medir_reales.py

Solo imprime agregados; el detalle por mensaje queda en reales/medicion_v2.jsonl (fuera de git).
"""
import json
import urllib.request
from collections import Counter
from pathlib import Path

AQUI = Path(__file__).resolve().parent
URLS = {"sin_router": "http://127.0.0.1:4763/ask", "con_router": "http://127.0.0.1:4764/ask"}


def preguntar(url: str, texto: str) -> dict:
    cuerpo = json.dumps({"pregunta": texto, "version": "v2", "k": 8, "source": "s10-kb"}).encode()
    req = urllib.request.Request(url, cuerpo, {"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=60) as r:
        d = json.load(r)
    p = d.get("plan") or {}
    return {"modo": d.get("modo"), "tipo": p.get("tipo"), "proc": (p.get("procedimiento") or {}).get("id"),
            "opciones": [o.get("id") for o in p.get("opciones") or []]}


def main():
    filas = [json.loads(l) for l in open(AQUI / "reales/prueba_real.jsonl", encoding="utf-8")]
    res = {k: Counter() for k in URLS}
    with open(AQUI / "reales/medicion_v2.jsonl", "w", encoding="utf-8") as f:
        for x in filas:
            fila = {"id": x["id"], "esperado": x["clase"]}
            for nombre, url in URLS.items():
                r = preguntar(url, x["texto"])
                fila[nombre] = r
                c = res[nombre]
                if x["clase"] == "_ninguno":
                    c["sin_caso"] += 1
                    c["sin_caso_responde_procedimiento"] += bool(r["proc"])
                    c["sin_caso_aclara"] += r["modo"] == "aclaracion"
                    c["sin_caso_otro"] += not r["proc"] and r["modo"] != "aclaracion"
                else:
                    c["con_caso"] += 1
                    c["con_caso_acierta"] += r["proc"] == x["clase"]
                    c["con_caso_opcion_correcta"] += x["clase"] in r["opciones"]
                    c["con_caso_responde_otro"] += bool(r["proc"]) and r["proc"] != x["clase"]
            f.write(json.dumps(fila, ensure_ascii=False) + "\n")
    for nombre, c in res.items():
        s, n = max(c["sin_caso"], 1), max(c["con_caso"], 1)
        print(f"{nombre}: sin caso {c['sin_caso']}: responde un procedimiento {c['sin_caso_responde_procedimiento'] / s:.1%}, "
              f"aclara {c['sin_caso_aclara'] / s:.1%}, otra cosa (social, sin evidencia, concepto) {c['sin_caso_otro'] / s:.1%} | "
              f"con caso {c['con_caso']}: acierta {c['con_caso_acierta']}/{n}, ofrece el correcto {c['con_caso_opcion_correcta']}, "
              f"responde otro {c['con_caso_responde_otro']}")


if __name__ == "__main__":
    main()
