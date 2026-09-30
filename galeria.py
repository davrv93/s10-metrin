#!/usr/bin/env python3
"""Galería estática de data/imagenes + vista con: python3 -m http.server 4792 -d data/imagenes."""
import json
from pathlib import Path

RAIZ = Path(__file__).resolve().parent
imgs = [json.loads(l) for l in (RAIZ / "data" / "imagenes.jsonl").read_text(encoding="utf-8").splitlines() if l.strip()]
partes = [
    "<meta charset=utf-8><title>Imagenes S10</title>",
    "<style>img{max-width:320px}figure{display:inline-block;margin:8px;vertical-align:top}body{font-family:sans-serif}</style>",
    f"<h1>{len(imgs)} imagenes del portal</h1>",
]
for im in imgs:
    rel = im["archivo"].split("/", 1)[1]
    alt = (im.get("alt") or "")[:80]
    partes.append(
        f"<figure><img src={rel!r} loading=lazy><figcaption>{alt}<br><small>{im['pagina'][:70]}</small></figcaption></figure>"
    )
(RAIZ / "data" / "imagenes" / "index.html").write_text("\n".join(partes), encoding="utf-8")
print("galeria:", len(imgs))
