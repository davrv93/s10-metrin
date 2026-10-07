#!/usr/bin/env python3
"""Extrae preguntas de clientes de exportaciones de WhatsApp (grupos de soporte S10) y las anonimiza.

Uso (desde entrenamiento-router/):
    .venv/bin/python extraer_reales.py      # lee reales/crudo/*/*.txt, escribe reales/mensajes.jsonl y reales/candidatas.jsonl

TODO lo de reales/ está fuera de git (.gitignore): son conversaciones reales con datos personales y el repo personal
es público. Ni siquiera anonimizado se versiona: el anonimizado es mecánico (teléfonos, correos, documentos, menciones,
nombres de los participantes) y no garantiza que no quede algo identificable dentro del texto.

- Hablante: «consultor» si su nombre lleva «Optimiza» o está en CONSULTORES; si no, «cliente» (numerado por grupo).
- Se descartan avisos del sistema, multimedia omitida, mensajes eliminados y adjuntos.
- Candidata a pregunta: mensaje de cliente con «?» o con una forma de pedir ayuda (cómo, dónde, no me deja, error…),
  de 3 a 60 palabras. Los mensajes seguidos del mismo cliente en menos de 2 minutos se unen antes de decidir.
"""
from __future__ import annotations

import json
import re
from datetime import datetime
from pathlib import Path

AQUI = Path(__file__).resolve().parent
REALES = AQUI / "reales"

LINEA = re.compile(r"^(\d{1,2}/\d{1,2}/\d{2,4}), (\d{1,2}:\d{2})\s*([ap])\.\s*m\.\s+-\s+(.*)$")
CONSULTORES = ("optimiza", "marin sistemas")
SISTEMA = ("cifrados de extremo a extremo", "creó el grupo", "añadió a", "te añadió", "salió", "eliminó a",
           "cambió", "se unió", "Multimedia omitido", "Se eliminó este mensaje", "Eliminaste este mensaje",
           "archivo adjunto", "(archivo adjunto)", "<Se editó este mensaje.>", "Mensaje de voz omitido")
PIDE = re.compile(r"\b(c[oó]mo|d[oó]nde|cu[aá]l|qu[eé] (hago|pasa|significa|es)|por ?qu[eé]|no (me )?(deja|sale|aparece|"
                  r"permite|carga|jala|trae|cuadra|figura|guarda|calcula|reconoce)|error|ayuda|ay[uú]denme|apoyo|consulta|"
                  r"se puede|puedo|hay forma|me sale|sale un|no encuentro|necesito)\b", re.I)


def anonimizar(t: str, nombres: list[str]) -> str:
    t = re.sub("@\u2068[^\u2069]*\u2069", "@persona", t)   # mención de WhatsApp: @⁨nombre con espacios⁩
    t = re.sub(r"@[^\s@]+", "@persona", t)
    t = t.replace("\u2068", "").replace("\u2069", "")
    t = re.sub(r"[\w.+-]+@[\w-]+\.[\w.]+", "<correo>", t)
    t = re.sub(r"https?://\S+", "<enlace>", t)
    t = re.sub(r"\+?\d[\d\s-]{7,}\d", "<numero>", t)       # teléfonos, RUC, DNI, cuentas
    for n in sorted(nombres, key=len, reverse=True):
        for parte in [n] + [p for p in n.split() if len(p) > 3]:
            t = re.sub(r"\b" + re.escape(parte) + r"\b", "<nombre>", t, flags=re.I)
    return re.sub(r"\s+", " ", t).strip()


def leer(ruta: Path, grupo: str) -> list[dict]:
    msgs, actual = [], None
    for linea in ruta.read_text(encoding="utf-8", errors="replace").splitlines():
        linea = linea.replace("‎", "").replace(" ", " ")
        m = LINEA.match(linea)
        if m:
            fecha, hora, ap, resto = m.groups()
            if ": " not in resto:
                actual = None
                continue
            autor, texto = resto.split(": ", 1)
            h, mi = map(int, hora.split(":"))
            h = h % 12 + (12 if ap == "p" else 0)
            d, mes, a = map(int, fecha.split("/"))
            a = a + 2000 if a < 100 else a
            actual = {"grupo": grupo, "t": datetime(a, mes, d, h, mi).isoformat(), "autor": autor.strip(" ~"), "texto": texto}
            msgs.append(actual)
        elif actual is not None:
            actual["texto"] += "\n" + linea
    return msgs


def main():
    todos, candidatas = [], []
    for ruta in sorted(REALES.glob("crudo/*/*.txt")):
        grupo = ruta.parent.name
        msgs = leer(ruta, grupo)
        nombres = sorted({m["autor"] for m in msgs if not m["autor"].startswith("+")})
        alias, nc = {}, 0
        for m in msgs:
            a = m["autor"]
            if a not in alias:
                if any(c in a.lower() for c in CONSULTORES):
                    alias[a] = "consultor"
                else:
                    nc += 1
                    alias[a] = f"cliente{nc}"
        limpios = []
        for m in msgs:
            t = m["texto"].strip()
            if not t or any(s in t for s in SISTEMA):
                continue
            limpios.append({"grupo": grupo, "t": m["t"], "rol": alias[m["autor"]], "texto": anonimizar(t, nombres)})
        # unir mensajes seguidos del mismo hablante (< 2 min)
        unidos = []
        for m in limpios:
            if unidos and unidos[-1]["rol"] == m["rol"] and (
                    datetime.fromisoformat(m["t"]) - datetime.fromisoformat(unidos[-1]["t"])).total_seconds() < 120:
                unidos[-1]["texto"] += " " + m["texto"]
                continue
            unidos.append(dict(m))
        for i, m in enumerate(unidos):
            m["id"] = f"{grupo}-{i:05d}"
            todos.append(m)
            n = len(m["texto"].split())
            if m["rol"].startswith("cliente") and 3 <= n <= 60 and ("?" in m["texto"] or PIDE.search(m["texto"])):
                ctx = [f'{x["rol"]}: {x["texto"][:200]}' for x in unidos[max(0, i - 2):i]]
                candidatas.append({"id": m["id"], "grupo": grupo, "t": m["t"], "texto": m["texto"], "contexto": ctx})
    REALES.mkdir(exist_ok=True)
    for nombre, filas in (("mensajes", todos), ("candidatas", candidatas)):
        with open(REALES / f"{nombre}.jsonl", "w", encoding="utf-8") as f:
            for x in filas:
                f.write(json.dumps(x, ensure_ascii=False) + "\n")
    por_grupo = {}
    for c in candidatas:
        por_grupo[c["grupo"]] = por_grupo.get(c["grupo"], 0) + 1
    print(json.dumps({"mensajes": len(todos), "de_clientes": sum(m["rol"].startswith("cliente") for m in todos),
                      "candidatas": len(candidatas), "por_grupo": por_grupo}, ensure_ascii=False, indent=1))


if __name__ == "__main__":
    main()
