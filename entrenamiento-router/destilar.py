# Copia sin cambios de pjgfarma-analista/analista/herramientas/destilar/destilar.py (07-10-2026):
# entrenar.py importa sus funciones de recorte, cuantización y escritura PJGE.
"""Destila (o toma) un modelo de embeddings estáticos tipo model2vec, recorta
su vocabulario al español y al dominio de boticas y lo exporta al formato
PJGE que lee el paquete Go `analista/interno/preguntas`.

Uso (ver README.md):

    python destilar.py --base minishlab/potion-multilingual-128M --modo tomar \
        --corpus corpus/*.txt --salida potion-es.pjge --cuant int8 \
        --paridad frases_paridad.txt --paridad-salida paridad.jsonl

Qué hace, en orden:
1. Carga el modelo estático (`tomar`) o lo destila de un sentence-transformer
   (`destilar`, con model2vec y PCA).
2. Tokeniza el corpus (palabras del español, productos, principios activos,
   categorías y ejemplos de entrenamiento) con el tokenizador COMPLETO y se
   queda con las piezas que salen, más las de un solo carácter del alfabeto.
3. Arma un tokenizador Unigram recortado con esas piezas (puntajes
   redondeados a float32, igual que en el fichero) y lo guarda en JSON: es la
   referencia exacta contra la que se compara el tokenizador de Go.
4. Escribe el fichero PJGE: tabla de normalización, piezas con puntaje y
   matriz cuantizada (int8 con escala por fila, o float16).
5. Opcional: escribe embeddings de referencia para la prueba de paridad Go.

No abre conexiones a ninguna base: los ficheros del corpus se generan aparte
(ver `corpus_desde_bd.py`).
"""

from __future__ import annotations

import argparse
import base64
import copy
import hashlib
import json
import struct
import sys
import zlib
from pathlib import Path

import numpy as np
from tokenizers import Tokenizer, normalizers

MAGIA = b"PJGE"
VERSION = 1
CUANT_INT8, CUANT_F16 = 1, 2
BANDERA_PUNTUACION = 1  # el normalizador aísla la puntuación ASCII con espacios

PUNTUACION_ASCII = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
# caracteres que siempre deben tener pieza propia, para no caer en [UNK]
ALFABETO = (
    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    "áéíóúüñÁÉÍÓÚÜÑ¿¡°ºª" + PUNTUACION_ASCII
)


def cargar_estatico(base: str, modo: str, pca: int):
    from model2vec import StaticModel

    if modo == "tomar":
        return StaticModel.from_pretrained(base)
    from model2vec.distill import distill

    return distill(model_name=base, pca_dims=pca)


def filas_efectivas(modelo) -> np.ndarray:
    """Una fila por id del tokenizador, con pesos y mapeo ya aplicados."""
    emb = np.asarray(modelo.embedding, dtype=np.float32)
    if modelo.token_mapping is not None:
        emb = emb[np.asarray(modelo.token_mapping)]
    if modelo.weights is not None:
        emb = emb * np.asarray(modelo.weights, dtype=np.float32)[:, None]
    return emb


def analizar_normalizador(tj: dict) -> tuple[bytes, int]:
    """Comprueba que el normalizador es uno de los que Go sabe replicar y
    devuelve (charsmap, banderas). Si no lo es, aborta: mejor un error aquí
    que un tokenizador que difiere en silencio."""
    n = tj["normalizer"]
    pasos = []

    def aplanar(x):
        if x is None:
            return
        if x["type"] == "Sequence":
            for y in x["normalizers"]:
                aplanar(y)
        else:
            pasos.append(x)

    aplanar(n)
    if not pasos or pasos[0]["type"] != "Precompiled":
        sys.exit(f"normalizador no soportado: {[p['type'] for p in pasos]}")
    charsmap = base64.b64decode(pasos[0]["precompiled_charsmap"])
    resto = pasos[1:]
    # colapsar espacios no cambia nada: Go parte por espacios de todos modos
    if resto[:1] == [{"type": "Replace", "pattern": {"Regex": " {2,}"}, "content": " "}]:
        resto = resto[1:]
    if not resto:
        return charsmap, 0
    punt = [{"type": "Replace", "pattern": {"String": c}, "content": f" {c} "} for c in PUNTUACION_ASCII]
    cola = [
        {"type": "Replace", "pattern": {"Regex": "\\s+"}, "content": " "},
        {"type": "Strip", "strip_left": True, "strip_right": True},
    ]
    if resto != punt + cola:
        sys.exit("normalizador no soportado: la lista de reemplazos no es la conocida")
    return charsmap, BANDERA_PUNTUACION


def tabla_normalizacion(charsmap: bytes) -> list[tuple[int, str]]:
    """Recorre todos los puntos de código y guarda los que el Precompiled
    cambia. Las reglas de secuencias (p. ej. «a» + tilde combinante) no se
    capturan: los teclados de celular mandan la letra ya compuesta."""
    pre = normalizers.Precompiled(charsmap)
    tabla = []
    for cp in range(0x110000):
        if 0xD800 <= cp <= 0xDFFF:
            continue
        ch = chr(cp)
        out = pre.normalize_str(ch)
        if out != ch:
            tabla.append((cp, out))
    return tabla


def piezas_del_corpus(tok: Tokenizer, textos: list[str]) -> set[int]:
    ids: set[int] = set()
    for enc in tok.encode_batch(textos, add_special_tokens=False):
        ids.update(enc.ids)
    vocab = tok.get_vocab()
    for ch in ALFABETO:
        for pieza in (ch, "▁" + ch):
            if pieza in vocab:
                ids.add(vocab[pieza])
    if "▁" in vocab:
        ids.add(vocab["▁"])
    return ids


def tokenizador_recortado(tj: dict, conservar: list[int]) -> tuple[dict, list[tuple[str, float]], int]:
    vocab = tj["model"]["vocab"]
    unk_orig = tj["model"]["unk_id"]
    nuevos = []
    unk_nuevo = None
    for i in conservar:
        pieza, puntaje = vocab[i]
        if i == unk_orig:
            unk_nuevo = len(nuevos)
        nuevos.append((pieza, float(np.float32(puntaje))))
    if unk_nuevo is None:
        sys.exit("la pieza [UNK] no quedó en el vocabulario recortado")
    tj2 = copy.deepcopy(tj)
    tj2["model"]["vocab"] = [[p, s] for p, s in nuevos]
    tj2["model"]["unk_id"] = unk_nuevo
    # solo se conserva [UNK] como token añadido, con su id nuevo
    tj2["added_tokens"] = [
        dict(a, id=unk_nuevo) for a in tj["added_tokens"] if a["id"] == unk_orig
    ]
    return tj2, nuevos, unk_nuevo


def cuantizar(filas: np.ndarray, cuant: str) -> tuple[bytes, np.ndarray]:
    """Devuelve (bytes de la matriz, matriz decuantizada para la referencia)."""
    if cuant == "int8":
        escala = np.abs(filas).max(axis=1) / 127.0
        escala[escala == 0] = 1.0
        escala = escala.astype(np.float32)
        q = np.clip(np.rint(filas / escala[:, None]), -127, 127).astype(np.int8)
        deq = q.astype(np.float32) * escala[:, None]
        return q.tobytes() + escala.astype("<f4").tobytes(), deq
    h = filas.astype("<f2")
    return h.tobytes(), h.astype(np.float32)


def escribir_pjge(ruta: Path, dim: int, piezas, unk: int, banderas: int, cuant: str, tabla, matriz: bytes):
    b = bytearray()
    b += MAGIA
    b += struct.pack("<IIII", VERSION, dim, len(piezas), unk)
    b += struct.pack("<BBH", CUANT_INT8 if cuant == "int8" else CUANT_F16, banderas, 0)
    b += struct.pack("<I", len(tabla))
    for cp, out in tabla:
        ob = out.encode("utf-8")
        b += struct.pack("<IB", cp, len(ob)) + ob
    for pieza, puntaje in piezas:
        pb = pieza.encode("utf-8")
        if len(pb) > 255:
            sys.exit(f"pieza demasiado larga: {pieza!r}")
        b += struct.pack("<B", len(pb)) + pb + struct.pack("<f", puntaje)
    b += matriz
    b += struct.pack("<I", zlib.crc32(bytes(b)) & 0xFFFFFFFF)
    ruta.write_bytes(bytes(b))
    return hashlib.sha256(bytes(b)).hexdigest()


def embeber_referencia(tok: Tokenizer, deq: np.ndarray, unk: int, textos: list[str]) -> np.ndarray:
    """Lo mismo que StaticModel.encode: tokens sin [UNK], media y norma L2."""
    out = np.zeros((len(textos), deq.shape[1]), dtype=np.float32)
    for k, enc in enumerate(tok.encode_batch(textos, add_special_tokens=False)):
        ids = [i for i in enc.ids if i != unk][:512]
        if ids:
            v = deq[ids].mean(axis=0)
            out[k] = v / (np.linalg.norm(v) + 1e-32)
    return out


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--base", required=True, help="id de Hugging Face o carpeta local del modelo")
    ap.add_argument("--procedencia", help="texto para el .json, p. ej. «minishlab/potion-multilingual-128M@73908c3»")
    ap.add_argument("--modo", choices=["tomar", "destilar"], default="tomar")
    ap.add_argument("--pca", type=int, default=256)
    ap.add_argument("--corpus", nargs="+", required=True, help="ficheros de texto, una entrada por línea")
    ap.add_argument("--salida", required=True, type=Path)
    ap.add_argument("--cuant", choices=["int8", "f16"], default="int8")
    ap.add_argument("--sin-recorte", action="store_true", help="conserva todo el vocabulario (solo para medir)")
    ap.add_argument("--paridad", type=Path, help="frases para la prueba de paridad Go/Python")
    ap.add_argument("--paridad-salida", type=Path)
    a = ap.parse_args()

    modelo = cargar_estatico(a.base, a.modo, a.pca)
    tj = json.loads(modelo.tokenizer.to_str())
    if tj["model"]["type"] != "Unigram":
        sys.exit(f"tokenizador {tj['model']['type']} no soportado (solo Unigram)")
    # Go parte por espacios y antepone ▁ a cada palabra. Es equivalente a
    # Metaspace (con split o sin él) siempre que ninguna pieza lleve ▁ en
    # medio, que se comprueba justo debajo.
    pre = tj["pre_tokenizer"]
    meta = {"type": "Metaspace", "replacement": "▁", "prepend_scheme": "always"}
    pres = pre["pretokenizers"] if pre.get("type") == "Sequence" else [pre]
    if [p["type"] for p in pres] not in (["Metaspace"], ["WhitespaceSplit", "Metaspace"]) or any(
        pres[-1].get(k) != v for k, v in meta.items()
    ):
        sys.exit(f"pre-tokenizador no soportado: {pre}")
    if any("▁" in p[1:] for p, _ in tj["model"]["vocab"]):
        sys.exit("hay piezas con ▁ interno: el tokenizador de Go no las soporta")
    charsmap, banderas = analizar_normalizador(tj)
    filas = filas_efectivas(modelo)
    completo = Tokenizer.from_str(modelo.tokenizer.to_str())

    textos = []
    for ruta in a.corpus:
        for linea in Path(ruta).read_text(encoding="utf-8").splitlines():
            linea = linea.strip()
            if linea:
                textos.append(linea)
                textos.append(linea.lower())  # así se pregunta por WhatsApp
    if a.sin_recorte:
        conservar = list(range(len(tj["model"]["vocab"])))
    else:
        conservar = sorted(piezas_del_corpus(completo, textos) | {tj["model"]["unk_id"]})
    tj2, piezas, unk = tokenizador_recortado(tj, conservar)
    tok = Tokenizer.from_str(json.dumps(tj2))
    matriz, deq = cuantizar(filas[conservar], a.cuant)
    tabla = tabla_normalizacion(charsmap)
    sha = escribir_pjge(a.salida, filas.shape[1], piezas, unk, banderas, a.cuant, tabla, matriz)
    a.salida.with_suffix(".tokenizer.json").write_text(json.dumps(tj2, ensure_ascii=False))
    np.save(a.salida.with_suffix(".deq.npy"), deq)
    info = {
        "base": a.procedencia or a.base,
        "modo": a.modo,
        "piezas": len(piezas),
        "piezas_originales": len(tj["model"]["vocab"]),
        "dim": int(filas.shape[1]),
        "cuant": a.cuant,
        "bytes": a.salida.stat().st_size,
        "sha256": sha,
        "textos_corpus": len(textos),
        "tabla_normalizacion": len(tabla),
    }
    a.salida.with_suffix(".json").write_text(json.dumps(info, indent=2, ensure_ascii=False))
    print(json.dumps(info, ensure_ascii=False))

    if a.paridad:
        frases = [l for l in a.paridad.read_text(encoding="utf-8").splitlines() if l.strip()]
        emb = embeber_referencia(tok, deq, unk, frases)
        with open(a.paridad_salida, "w", encoding="utf-8") as f:
            for frase, e, enc in zip(frases, emb, tok.encode_batch(frases, add_special_tokens=False)):
                f.write(json.dumps({"texto": frase, "ids": enc.ids, "emb": [round(float(x), 6) for x in e]}, ensure_ascii=False) + "\n")


if __name__ == "__main__":
    main()
