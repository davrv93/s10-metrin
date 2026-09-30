#!/usr/bin/env python3
"""Limpieza de ui-pantallas: borra los nodos duplicados/basura de las primeras cargas.

La API de Cortex reserva el borrado a humanos ("Only humans can delete
knowledge"): corre esto como dueño. El token owner se lee de
.cortex/.secrets.yaml y nunca se imprime.

    .venv/bin/python herramientas/limpiar_ui_pantallas.py --solo-listar  # ver qué borraría
    .venv/bin/python herramientas/limpiar_ui_pantallas.py                # lista y borra
"""
import argparse
import re
import sys
from pathlib import Path

import requests

RAIZ = Path(__file__).resolve().parent.parent
API = "http://localhost:4748/api"

sec = (RAIZ / ".cortex/.secrets.yaml").read_text()
token = re.search(r"owner:\s*(\S+)", sec).group(1)
H = {"Authorization": f"Bearer {token}"}

# Se conservan: _node.md (raíz) y el consolidado por video. Todo lo demás
# generado en las primeras cargas (por captura, con ?t=N, títulos mal leídos)
# se borra.
CONSERVAR = {"_node.md", "17w0yki8iew"}

carpeta = RAIZ / ".cortex/tree/ui-pantallas"
a_borrar = sorted(p.stem for p in carpeta.glob("*.md") if p.stem not in CONSERVAR)
print(f"{len(a_borrar)} nodos a borrar de ui-pantallas:")
for nombre in a_borrar:
    print("  -", nombre)

if "--solo-listar" in sys.argv:
    sys.exit(0)

ok = 0
for nombre in a_borrar:
    r = requests.delete(f"{API}/node/ui-pantallas/{nombre}", headers=H, timeout=15)
    ok += r.status_code in (200, 202, 204)
    if r.status_code not in (200, 202, 204):
        print(f"ERROR {r.status_code} en {nombre}: {r.text[:120]}")
print(f"\n{ok}/{len(a_borrar)} borrados. Quedan en ui-pantallas:")
for p in sorted(carpeta.glob("*.md")):
    print("  ", p.stem)
