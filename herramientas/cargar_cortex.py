#!/usr/bin/env python3
"""Carga los borradores data/cortex-borradores/*.json en el Cortex local (API en :4748).

Escribe como el actor ai-agent (token en .cortex/.secrets.yaml, nunca se imprime).
Padres antes que hijos. Uso: .venv/bin/python herramientas/cargar_cortex.py [archivo.json ...]
"""
import json, os, re, sys
from pathlib import Path
import requests

RAIZ = Path(__file__).resolve().parent.parent
API = os.environ.get("S10_CORTEX_API", "http://localhost:4748/api")
sec = (RAIZ / ".cortex/.secrets.yaml").read_text()
token = re.search(r"ai-agent:\s*(\S+)", sec).group(1)
H = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
if os.environ.get("S10_CORTEX_HOST"):  # en Docker: Cortex solo acepta Host: localhost
    H["Host"] = os.environ["S10_CORTEX_HOST"]

archivos = [Path(a) for a in sys.argv[1:]] or sorted((RAIZ / "data/cortex-borradores").glob("*.json"))
nodos = [n for a in archivos for n in json.loads(a.read_text())]
nodos.sort(key=lambda n: (n["path"].count("/") + (1 if n["path"] else 0)))
ok = 0
for n in nodos:
    cuerpo = {k: n[k] for k in ("title", "summary", "body", "tags") if k in n}
    cuerpo["reason"] = "Conocimiento redactado desde las fuentes del proyecto (PDF, portal, YouTube, Metrín)"
    url = f"{API}/node/{n['path']}" if n["path"] else f"{API}/node"
    r = requests.put(url, json=cuerpo, headers=H, timeout=30)
    estado = "ok" if r.status_code in (200, 202) else f"ERROR {r.status_code} {r.text[:200]}"
    ok += r.status_code in (200, 202)
    print(f"{r.status_code} {n['path'] or '(raíz)'} {'' if r.status_code in (200,202) else estado}")
print(f"{ok}/{len(nodos)} nodos cargados")
