#!/usr/bin/env python3
"""Evalúa un router exportado (.pjge + cabeza.json, lo mismo que carga Go) sobre un JSONL {texto, clase}.

    .venv/bin/python evaluar_modelo.py <carpeta con router-s10.*> <jsonl> [<jsonl>…]
"""
import json
import struct
import sys
from pathlib import Path

import numpy as np
from tokenizers import Tokenizer

sys.path.insert(0, str(Path(__file__).resolve().parent))
from entrenar import decidir, medir, MAX_TOKENS  # noqa: E402


def cargar_pjge(ruta: Path):
    b = ruta.read_bytes()
    _, dim, n, unk = struct.unpack_from("<IIII", b, 4)
    cuant = b[20]
    off = 24
    (nnorm,) = struct.unpack_from("<I", b, off)
    off += 4
    for _ in range(nnorm):
        _, largo = struct.unpack_from("<IB", b, off)
        off += 5 + largo
    for _ in range(n):
        off += 1 + b[off] + 4
    assert cuant == 1, "solo int8"
    q = np.frombuffer(b, dtype=np.int8, count=n * dim, offset=off).reshape(n, dim).astype(np.float32)
    esc = np.frombuffer(b, dtype="<f4", count=n, offset=off + n * dim)
    return q * esc[:, None], unk


def main():
    d = Path(sys.argv[1])
    cab = json.loads((d / "router-s10.cabeza.json").read_text())
    deq, unk = cargar_pjge(d / "router-s10.pjge")
    tok = Tokenizer.from_file(str(d / "router-s10.tokenizer.json"))
    tok.no_padding()      # como Go: sin relleno ni recorte
    tok.no_truncation()
    W, b = np.array(cab["W"], np.float32), np.array(cab["b"], np.float32)
    clases, modulo_de, u = cab["clases"], cab["modulo"], cab["umbrales"]
    idx = {c: i for i, c in enumerate(clases)}
    for ruta in sys.argv[2:]:
        filas = [json.loads(l) for l in open(ruta, encoding="utf-8") if l.strip()]
        P = []
        for enc in tok.encode_batch([x["texto"] for x in filas], add_special_tokens=False):
            ids = [i for i in enc.ids if i != unk][:MAX_TOKENS]
            v = deq[ids].mean(axis=0) if ids else np.zeros(deq.shape[1], np.float32)
            v = v / (np.linalg.norm(v) + 1e-32)
            z = W @ v + b
            z = np.exp(z - z.max())
            P.append(z / z.sum())
        r = medir(np.array(P), [idx[x["clase"]] for x in filas], clases, modulo_de, u)
        print(Path(ruta).name, json.dumps({k: v for k, v in r.items() if k != "conteos"}, ensure_ascii=False))
        print("   conteos", json.dumps(r["conteos"], ensure_ascii=False))


if __name__ == "__main__":
    main()
