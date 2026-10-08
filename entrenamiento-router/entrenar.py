#!/usr/bin/env python3
"""Entrena el router de Metrín: pregunta → procedimiento (60 casos + «_ninguno»). docs/TOC-RUTEO-METRIN.md.

Uso (desde entrenamiento-router/, tras construir_datos.py):
    .venv/bin/python entrenar.py                 # recorta, entrena (congelado y afinado), calibra, exporta
    .venv/bin/python entrenar.py --sin-exportar  # solo mide

El modelo es exactamente lo que Go puede ejecutar sin dependencias:
    v = normalizar_L2( suma de las filas de los tokens de la pregunta, sin [UNK], máx. 512 )   # = embed.Modelo.Embeber
    logits = W·v + b ;  p = softmax(logits)
Se ajustan a la vez las filas del modelo estático (fine-tuning de potion-multilingual-128M recortado al vocabulario de
S10) y la cabeza lineal W, b. «congelado» = la misma cabeza con las filas sin tocar, para medir qué aporta el ajuste.

Salidas en salida/:
- router-s10.pjge (+ .json, .tokenizer.json): embeddings afinados, int8, formato PJGE (metrin/internal/embed).
- router-s10.cabeza.json: clases, módulo de cada clase, W, b, umbrales calibrados y huella del .pjge.
- router-s10.paridad.jsonl: frases con ids, embedding y probabilidades para la prueba de paridad Go/Python.
- informe.json: métricas en calibración y en prueba (metrin/eval/v2_oro.jsonl, nunca usado para entrenar).
"""
from __future__ import annotations

import argparse
import hashlib
import json
import random
import struct
import sys
import unicodedata
from collections import Counter, defaultdict
from pathlib import Path

import numpy as np
import torch
import torch.nn.functional as F
from tokenizers import Tokenizer

AQUI = Path(__file__).resolve().parent
RAIZ = AQUI.parent
sys.path.insert(0, str(AQUI))
import destilar as D  # noqa: E402

NINGUNO = "_ninguno"
MAX_TOKENS = 512
SEMILLA = 20261007


# ---------------------------------------------------------------- datos

def leer_jsonl(ruta: Path) -> list[dict]:
    return [json.loads(l) for l in open(ruta, encoding="utf-8") if l.strip()]


def sin_tildes(t: str) -> str:
    t = unicodedata.normalize("NFKD", t.lower())
    return unicodedata.normalize("NFC", "".join(c for c in t if not unicodedata.combining(c) or c == "̃"))


def piezas_pjge(ruta: Path) -> list[str]:
    """Piezas de un .pjge existente (para conservar su vocabulario general del español)."""
    b = ruta.read_bytes()
    assert b[:4] == b"PJGE"
    _, dim, n, unk = struct.unpack_from("<IIII", b, 4)
    off = 4 + 16 + 4
    (nnorm,) = struct.unpack_from("<I", b, off)
    off += 4
    for _ in range(nnorm):
        _, largo = struct.unpack_from("<IB", b, off)
        off += 5 + largo
    piezas = []
    for _ in range(n):
        largo = b[off]
        piezas.append(b[off + 1:off + 1 + largo].decode("utf-8"))
        off += 1 + largo + 4
    return piezas


def corpus_recorte(entren: list[dict], calib: list[dict]) -> list[str]:
    # Los textos REALES de clientes no deciden el vocabulario del .pjge, que va en git (repo público).
    textos = [x["texto"] for x in entren + calib if x.get("origen") != "real"]
    for ruta in sorted((RAIZ / "kb").glob("fragmentos*.jsonl")):
        for x in leer_jsonl(ruta):
            textos.append((x.get("titulo") or "") + " " + (x.get("texto") or ""))
    for ruta in sorted((RAIZ / "kb/procedimientos").glob("*/*.yml")):
        textos.extend(ruta.read_text(encoding="utf-8").splitlines())
    ruta = RAIZ / "data/preguntas/preguntas.jsonl"
    if ruta.exists():
        for x in leer_jsonl(ruta):
            textos.append(x.get("pregunta") or "")
            textos.extend(v.get("texto", "") for v in x.get("variantes") or [])
    textos = [t.strip() for t in textos if t and t.strip()]
    return textos + [t.lower() for t in textos]


def recortar(base: Path, textos: list[str], pjge_previo: Path | None):
    """Mismo recorte que destilar.py (modo tomar) + las piezas del .pjge previo. Devuelve lo necesario para exportar."""
    modelo = D.cargar_estatico(str(base), "tomar", 0)
    tj = json.loads(modelo.tokenizer.to_str())
    if tj["model"]["type"] != "Unigram":
        sys.exit("solo Unigram")
    pre = tj["pre_tokenizer"]
    pres = pre["pretokenizers"] if pre.get("type") == "Sequence" else [pre]
    meta = {"type": "Metaspace", "replacement": "▁", "prepend_scheme": "always"}
    if [p["type"] for p in pres] not in (["Metaspace"], ["WhitespaceSplit", "Metaspace"]) or any(
            pres[-1].get(k) != v for k, v in meta.items()):
        sys.exit(f"pre-tokenizador no soportado: {pre}")
    if any("▁" in p[1:] for p, _ in tj["model"]["vocab"]):
        sys.exit("piezas con ▁ interno")
    charsmap, banderas = D.analizar_normalizador(tj)
    filas = D.filas_efectivas(modelo)
    completo = Tokenizer.from_str(modelo.tokenizer.to_str())
    conservar = D.piezas_del_corpus(completo, textos) | {tj["model"]["unk_id"]}
    if pjge_previo and pjge_previo.exists():
        indice = {p: i for i, (p, _) in enumerate(tj["model"]["vocab"])}
        conservar |= {indice[p] for p in piezas_pjge(pjge_previo) if p in indice}
    conservar = sorted(conservar)
    tj2, piezas, unk = D.tokenizador_recortado(tj, conservar)
    tok = Tokenizer.from_str(json.dumps(tj2))
    return {"tok": tok, "tj": tj2, "piezas": piezas, "unk": unk, "banderas": banderas, "charsmap": charsmap,
            "filas": filas[conservar].astype(np.float32), "originales": len(tj["model"]["vocab"])}


def tokenizar(tok: Tokenizer, unk: int, textos: list[str]) -> list[list[int]]:
    out = []
    for enc in tok.encode_batch(textos, add_special_tokens=False):
        out.append([i for i in enc.ids if i != unk][:MAX_TOKENS])
    return out


# ---------------------------------------------------------------- modelo

class Router(torch.nn.Module):
    def __init__(self, filas: np.ndarray, n_clases: int, ajustar: bool):
        super().__init__()
        self.emb = torch.nn.EmbeddingBag.from_pretrained(torch.from_numpy(filas), freeze=not ajustar, mode="sum")
        self.base = torch.from_numpy(filas.copy())
        self.cabeza = torch.nn.Linear(filas.shape[1], n_clases)

    def vector(self, ids: torch.Tensor, offsets: torch.Tensor) -> torch.Tensor:
        # suma y norma L2 = media y norma L2 (Go divide por n antes de normalizar: misma dirección)
        return F.normalize(self.emb(ids, offsets), dim=1, eps=1e-12)

    def forward(self, ids, offsets):
        return self.cabeza(self.vector(ids, offsets))


def lote(seqs: list[list[int]]):
    ids, offsets, pos = [], [], 0
    for s in seqs:
        offsets.append(pos)
        ids.extend(s)
        pos += len(s)
    return torch.tensor(ids, dtype=torch.long), torch.tensor(offsets, dtype=torch.long)


def probabilidades(modelo: Router, seqs: list[list[int]]) -> np.ndarray:
    modelo.eval()
    with torch.no_grad():
        out = []
        for i in range(0, len(seqs), 512):
            parte = [s if s else [0] for s in seqs[i:i + 512]]  # sin piezas conocidas: vector de la pieza 0 (raro)
            out.append(F.softmax(modelo(*lote(parte)), dim=1).numpy())
    return np.concatenate(out) if out else np.zeros((0, modelo.cabeza.out_features))


def entrenar(filas, n_clases, X, y, Xc, yc, ajustar: bool, epocas=40, lr_cabeza=3e-3, lr_emb=1e-3, deriva=1e-2,
             pesos_clase=None, log=print, semilla=SEMILLA):
    torch.manual_seed(semilla)
    rnd = random.Random(semilla)
    m = Router(filas, n_clases, ajustar)
    grupos = [{"params": m.cabeza.parameters(), "lr": lr_cabeza, "weight_decay": 1e-4}]
    if ajustar:
        grupos.append({"params": m.emb.parameters(), "lr": lr_emb, "weight_decay": 0.0})
    opt = torch.optim.AdamW(grupos)
    peso = torch.tensor(pesos_clase, dtype=torch.float32) if pesos_clase is not None else None
    mejor, mejor_estado, sin_mejora = -1.0, None, 0
    orden = list(range(len(X)))
    for ep in range(epocas):
        m.train()
        rnd.shuffle(orden)
        perdida = 0.0
        for i in range(0, len(orden), 64):
            idx = [j for j in orden[i:i + 64] if X[j]]
            if not idx:
                continue
            logits = m(*lote([X[j] for j in idx]))
            loss = F.cross_entropy(logits, torch.tensor([y[j] for j in idx]), weight=peso, label_smoothing=0.05)
            if ajustar and deriva:
                # que las filas no se alejen sin necesidad del modelo original (solo pesa en las que se usan)
                usados = torch.unique(torch.cat([torch.tensor(X[j]) for j in idx]))
                loss = loss + deriva * (m.emb.weight[usados] - m.base[usados]).pow(2).sum(dim=1).mean()
            opt.zero_grad()
            loss.backward()
            opt.step()
            perdida += float(loss.detach()) * len(idx)
        acc = float((probabilidades(m, Xc).argmax(1) == np.array(yc)).mean())
        log(f"  época {ep + 1:2d}  pérdida {perdida / len(X):.4f}  acierto calibración {acc:.4f}")
        if acc > mejor + 1e-4:
            mejor, sin_mejora = acc, 0
            mejor_estado = {k: v.clone() for k, v in m.state_dict().items()}
        else:
            sin_mejora += 1
            if sin_mejora >= 6:
                break
    m.load_state_dict(mejor_estado)
    return m, mejor


# ---------------------------------------------------------------- decisión (el árbol de docs/TOC-RUTEO-METRIN.md §5)

def decidir(p: np.ndarray, clases: list[str], modulo_de: list[str], u: dict, uso: float | None = None
            ) -> tuple[str, list[str]]:
    """Devuelve (acción, candidatos). acción ∈ elegir | aclarar_caso | aclarar_modulo | delegar.

    uso: probabilidad del filtro «¿es una pregunta de uso de S10?»; por debajo de u["uso"], delegar.
    Se elige con p1 ≥ acepta y p1 − p2 ≥ margen, o también (si ratio > 0) con p1 ≥ acepta_min y p1 ≥ ratio·p2:
    con 125 clases la probabilidad se reparte y el primero puede sacar mucha ventaja sin llegar a «acepta»."""
    if uso is not None and uso < u.get("uso", 0.0):
        return "delegar", []
    ini = clases.index(NINGUNO)
    orden = [i for i in np.argsort(-p) if i != ini]
    if p[ini] >= p[orden[0]] or p[orden[0]] < u["minimo"]:
        return "delegar", []
    pm = defaultdict(float)
    for i, c in enumerate(clases):
        if i != ini:
            pm[modulo_de[i]] += p[i]
    mods = sorted(pm, key=pm.get, reverse=True)
    if pm[mods[0]] < u["modulo"]:
        return "aclarar_modulo", mods[:3]
    dentro = [i for i in orden if modulo_de[i] == mods[0]]
    p1 = p[dentro[0]]
    p2 = p[dentro[1]] if len(dentro) > 1 else 0.0
    ratio = u.get("ratio", 0.0)
    if (p1 >= u["acepta"] and p1 - p2 >= u["margen"]) or (ratio > 0 and p1 >= u.get("acepta_min", 1.0)
                                                          and p1 >= ratio * p2):
        return "elegir", [clases[dentro[0]]]
    return "aclarar_caso", [clases[i] for i in dentro[:3]]


def medir(P: np.ndarray, y: list[int], clases, modulo_de, u: dict, usos=None) -> dict:
    ini = clases.index(NINGUNO)
    c = Counter()
    for k, (p, yi) in enumerate(zip(P, y)):
        accion, cand = decidir(p, clases, modulo_de, u, None if usos is None else float(usos[k]))
        real = clases[yi]
        pos = yi != ini
        c["con_caso" if pos else "sin_caso"] += 1
        if accion == "elegir":
            c[("elige_bien" if cand[0] == real else "elige_mal") + ("" if pos else "_sin_caso")] += 1
        elif accion == "aclarar_caso":
            c["aclara_caso" + ("_con_correcto" if real in cand else "_sin_correcto") + ("" if pos else "_sin_caso")] += 1
        elif accion == "aclarar_modulo":
            c["aclara_modulo" + ("" if pos else "_sin_caso")] += 1
            if pos and modulo_de[yi] in cand:
                c["aclara_modulo_con_correcto"] += 1
        else:
            c["delega" + ("" if pos else "_sin_caso")] += 1
        if pos:
            top = [i for i in np.argsort(-p) if i != ini]
            c["top1"] += top[0] == yi
            c["top3"] += yi in top[:3]
            c["modulo_top1"] += modulo_de[top[0]] == modulo_de[yi]
    n, s = max(c["con_caso"], 1), max(c["sin_caso"], 1)
    return {
        "con_caso": c["con_caso"], "sin_caso": c["sin_caso"],
        "top1_caso": round(c["top1"] / n, 4), "top3_caso": round(c["top3"] / n, 4),
        "top1_modulo": round(c["modulo_top1"] / n, 4),
        "elige_bien": round(c["elige_bien"] / n, 4), "elige_mal": round(c["elige_mal"] / n, 4),
        "aclara_con_correcto": round((c["aclara_caso_con_correcto"] + c["aclara_modulo_con_correcto"]) / n, 4),
        "delega_con_caso": round(c["delega"] / n, 4),
        "sin_caso_delega": round(c["delega_sin_caso"] / s, 4),
        "sin_caso_elige": round(c["elige_mal_sin_caso"] / s, 4),
        "conteos": {k: int(v) for k, v in c.items()},
    }


PENALIZACION = 3  # elegir mal cuesta 3 elecciones buenas; 10 es más prudente pero pregunta de más (README, v5)


def calibrar(P, y, clases, modulo_de, penalizacion: float = PENALIZACION, usos=None, uso: float = 0.0) -> dict:
    """Elige umbrales en CALIBRACIÓN: máximo de (bien − penalizacion·mal) entre las elecciones; una aclaración vale 0,3
    si trae el correcto. Elegir en una pregunta sin caso cuenta como mal. El umbral del filtro de uso viene dado."""
    mejor, mu = None, None
    for minimo in (0.10, 0.20, 0.30):
        for modulo in (0.40, 0.60, 0.80):
            for acepta in (0.50, 0.60, 0.70, 0.80):
                for margen in (0.05, 0.20):
                    for ratio, acepta_min in ((0, 1.0), (2, 0.3), (2, 0.4), (3, 0.3), (3, 0.4), (5, 0.3), (5, 0.4)):
                        u = {"minimo": minimo, "modulo": modulo, "acepta": acepta, "margen": margen, "ratio": ratio,
                             "acepta_min": acepta_min, "uso": uso}
                        r = medir(P, y, clases, modulo_de, u, usos)["conteos"]
                        v = (r.get("elige_bien", 0) - penalizacion * (r.get("elige_mal", 0) + r.get("elige_mal_sin_caso", 0))
                             + 0.3 * (r.get("aclara_caso_con_correcto", 0) + r.get("aclara_modulo_con_correcto", 0)))
                        if mejor is None or v > mejor:
                            mejor, mu = v, u
    return mu


# ---------------------------------------------------------------- filtro «¿es una pregunta de uso?»

def es_uso(x: dict) -> int:
    """1 si la fila es una pregunta de uso de S10 (con o sin procedimiento); 0 si es coordinación, charla u otro tema."""
    if x.get("origen") == "real":
        return 0 if x.get("etiqueta_original") == "no_pregunta" else 1
    return 0 if x["clase"] == NINGUNO and (x.get("intencion") or "").upper() in ("SOCIAL", "OTRO") else 1


def vectores(m: "Router", seqs: list[list[int]]) -> np.ndarray:
    m.eval()
    with torch.no_grad():
        out = []
        for i in range(0, len(seqs), 512):
            parte = [s if s else [0] for s in seqs[i:i + 512]]
            out.append(m.vector(*lote(parte)).numpy())
    return np.concatenate(out)


def entrenar_filtro(Ve, ye, Vc, yc, recall_min: float = 0.97):
    """Regresión logística sobre el embedding (norma 1). Umbral: el mayor que conserva recall_min de las preguntas de
    uso en calibración (no hay que callar preguntas de verdad), para rechazar el máximo de coordinación."""
    from sklearn.linear_model import LogisticRegression
    lr = LogisticRegression(C=4.0, class_weight="balanced", max_iter=2000).fit(Ve, ye)
    pc = lr.predict_proba(Vc)[:, 1]
    pos = np.sort(pc[np.array(yc) == 1])
    umbral = float(pos[int((1 - recall_min) * len(pos))]) if len(pos) else 0.5
    rech = float((pc[np.array(yc) == 0] < umbral).mean()) if (np.array(yc) == 0).any() else 0.0
    return lr.coef_[0].astype(np.float32), float(lr.intercept_[0]), umbral, rech




# ---------------------------------------------------------------- principal

def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--base", type=Path, default=AQUI / "base/potion-multilingual-128M")
    ap.add_argument("--pjge-previo", type=Path, default=RAIZ / "metrin/modelos/potion-es-int8.pjge")
    ap.add_argument("--salida", type=Path, default=AQUI / "salida")
    ap.add_argument("--epocas", type=int, default=40)
    ap.add_argument("--sin-exportar", action="store_true")
    ap.add_argument("--semilla", type=int, default=SEMILLA, help="para medir la varianza entre entrenamientos")
    a = ap.parse_args()
    a.salida.mkdir(exist_ok=True)
    log_f = open(a.salida / "entrenamiento.log", "w", encoding="utf-8")

    def log(*x):
        s = " ".join(str(v) for v in x)
        print(s, flush=True)
        log_f.write(s + "\n")

    cl = json.loads((AQUI / "datos/clases.json").read_text(encoding="utf-8"))
    clases = cl["clases"]
    modulo_de = [cl["modulo"].get(c, NINGUNO) for c in clases]
    idx = {c: i for i, c in enumerate(clases)}
    entren = leer_jsonl(AQUI / "datos/entrenamiento.jsonl")
    calib = leer_jsonl(AQUI / "datos/calibracion.jsonl")
    prueba = leer_jsonl(AQUI / "datos/prueba.jsonl")
    ruta_real = AQUI / "reales/prueba_real.jsonl"
    prueba_real = leer_jsonl(ruta_real) if ruta_real.exists() else []
    ruta_nuevos = AQUI / "datos/prueba_nuevos.jsonl"  # prueba de los procedimientos añadidos (escrita sin ver el entrenamiento)
    prueba_nuevos = leer_jsonl(ruta_nuevos) if ruta_nuevos.exists() else []

    log("recortando vocabulario…")
    r = recortar(a.base, corpus_recorte(entren, calib), a.pjge_previo)
    log(f"  piezas {len(r['piezas'])} de {r['originales']}, dim {r['filas'].shape[1]}")
    tok, unk = r["tok"], r["unk"]

    # aumento: cada pregunta también en minúsculas y sin tildes (así escribe la gente por chat)
    # las preguntas reales cuentan doble (son pocas y son el estilo que importa)
    entren = entren + [x for x in entren if x.get("origen") == "real"]
    textos_e = [x["texto"] for x in entren] + [sin_tildes(x["texto"]) for x in entren]
    y_e = [idx[x["clase"]] for x in entren] * 2
    X_e = tokenizar(tok, unk, textos_e)
    X_c, y_c = tokenizar(tok, unk, [x["texto"] for x in calib]), [idx[x["clase"]] for x in calib]
    X_p, y_p = tokenizar(tok, unk, [x["texto"] for x in prueba]), [idx[x["clase"]] for x in prueba]
    X_r, y_r = tokenizar(tok, unk, [x["texto"] for x in prueba_real]), [idx[x["clase"]] for x in prueba_real]
    X_n, y_n = tokenizar(tok, unk, [x["texto"] for x in prueba_nuevos]), [idx[x["clase"]] for x in prueba_nuevos]
    uso_e = [es_uso(x) for x in entren] * 2
    uso_c = [es_uso(x) for x in calib]
    cuenta = Counter(y_e)
    pesos = [len(y_e) / (len(clases) * max(cuenta[i], 1)) for i in range(len(clases))]
    log(f"  entrenamiento {len(X_e)} (con aumento), calibración {len(X_c)}, prueba {len(X_p)}")

    informe = {"piezas": len(r["piezas"]), "entrenamiento": len(X_e), "calibracion": len(X_c), "prueba": len(X_p)}
    modelos = {}
    for nombre, ajustar in (("congelado", False), ("afinado", True)):
        log(f"\n== {nombre}")
        m, acc_c = entrenar(r["filas"], len(clases), X_e, y_e, X_c, y_c, ajustar, epocas=a.epocas,
                            pesos_clase=pesos, log=log, semilla=a.semilla)
        Pc, Pp = probabilidades(m, X_c), probabilidades(m, X_p)
        wf, bf, uf, rech = entrenar_filtro(vectores(m, X_e), uso_e, vectores(m, X_c), uso_c)
        log(f"  filtro de uso: umbral {uf:.3f}, rechaza el {rech:.1%} de la coordinación de calibración (recall 97 %)")
        def usos(seqs):
            return 1 / (1 + np.exp(-(vectores(m, seqs) @ wf + bf))) if seqs else None
        u = calibrar(Pc, y_c, clases, modulo_de, usos=usos(X_c), uso=uf)
        informe[nombre] = {"umbrales": u, "acierto_calibracion": round(acc_c, 4), "filtro_rechazo_calibracion": round(rech, 4),
                           "calibracion": medir(Pc, y_c, clases, modulo_de, u, usos(X_c)),
                           "prueba": medir(Pp, y_p, clases, modulo_de, u, usos(X_p))}
        if X_r:
            informe[nombre]["prueba_real"] = medir(probabilidades(m, X_r), y_r, clases, modulo_de, u, usos(X_r))
            log("  real:", json.dumps({k: v for k, v in informe[nombre]["prueba_real"].items() if k != "conteos"},
                                      ensure_ascii=False))
        if X_n:
            informe[nombre]["prueba_nuevos"] = medir(probabilidades(m, X_n), y_n, clases, modulo_de, u, usos(X_n))
            log("  nuevos:", json.dumps({k: v for k, v in informe[nombre]["prueba_nuevos"].items() if k != "conteos"},
                                        ensure_ascii=False))
        log(json.dumps({k: v for k, v in informe[nombre]["prueba"].items() if k != "conteos"}, ensure_ascii=False))
        modelos[nombre] = (m, u, (wf, bf))

    # detalle de la prueba con el afinado (para leer los fallos)
    m, u, (wf, bf) = modelos["afinado"]
    Pp = probabilidades(m, X_p)
    with open(a.salida / "prueba_detalle.jsonl", "w", encoding="utf-8") as f:
        for x, p in zip(prueba, Pp):
            accion, cand = decidir(p, clases, modulo_de, u)
            top = [(clases[i], round(float(p[i]), 3)) for i in np.argsort(-p)[:3]]
            f.write(json.dumps({"id": x["id"], "categoria": x["categoria"], "texto": x["texto"], "esperado": x["clase"],
                                "accion": accion, "candidatos": cand, "top3": top}, ensure_ascii=False) + "\n")

    if prueba_real:  # detalle de la prueba real: dentro de reales/ (fuera de git), nunca en salida/
        with open(AQUI / "reales/prueba_real_detalle.jsonl", "w", encoding="utf-8") as f:
            for x, p in zip(prueba_real, probabilidades(m, X_r)):
                accion, cand = decidir(p, clases, modulo_de, u)
                top = [(clases[i], round(float(p[i]), 3)) for i in np.argsort(-p)[:3]]
                f.write(json.dumps({"id": x["id"], "texto": x["texto"], "esperado": x["clase"], "accion": accion,
                                    "candidatos": cand, "top3": top}, ensure_ascii=False) + "\n")

    if not a.sin_exportar:
        filas = m.emb.weight.detach().numpy().astype(np.float32)
        matriz, deq = D.cuantizar(filas, "int8")
        tabla = D.tabla_normalizacion(r["charsmap"])
        ruta = a.salida / "router-s10.pjge"
        sha = D.escribir_pjge(ruta, filas.shape[1], r["piezas"], unk, r["banderas"], "int8", tabla, matriz)
        ruta.with_suffix(".tokenizer.json").write_text(json.dumps(r["tj"], ensure_ascii=False), encoding="utf-8")
        W = m.cabeza.weight.detach().numpy().astype(np.float32)
        b = m.cabeza.bias.detach().numpy().astype(np.float32)

        # métricas reales tras cuantizar (lo que ejecutará Go)
        def vq(s):
            v = deq[s].mean(axis=0) if s else np.zeros(deq.shape[1], np.float32)
            return v / (np.linalg.norm(v) + 1e-32)

        def prob_q(seqs):
            out = []
            for s in seqs:
                z = W @ vq(s) + b
                z = np.exp(z - z.max())
                out.append(z / z.sum())
            return np.array(out)

        def uso_q(seqs):
            return np.array([1 / (1 + np.exp(-(float(vq(s) @ wf) + bf))) for s in seqs]) if seqs else None
        informe["afinado_int8"] = {
            "prueba": medir(prob_q(X_p), y_p, clases, modulo_de, u, uso_q(X_p)),
            "prueba_real": medir(prob_q(X_r), y_r, clases, modulo_de, u, uso_q(X_r)) if X_r else None,
            "prueba_nuevos": medir(prob_q(X_n), y_n, clases, modulo_de, u, uso_q(X_n)) if X_n else None,
            "calibracion": medir(prob_q(X_c), y_c, clases, modulo_de, u, uso_q(X_c))}
        cabeza = {
            "version": 1,
            "pjge": ruta.name, "sha256_pjge": sha,
            "clases": clases, "modulo": modulo_de, "ninguno": NINGUNO,
            "umbrales": u,
            "filtro": {"w": [round(float(v), 7) for v in wf], "b": round(float(bf), 7)},
            "W": [[round(float(v), 7) for v in fila] for fila in W],
            "b": [round(float(v), 7) for v in b],
        }
        (a.salida / "router-s10.cabeza.json").write_text(json.dumps(cabeza, ensure_ascii=False), encoding="utf-8")
        (a.salida / "router-s10.json").write_text(json.dumps({
            "base": "minishlab/potion-multilingual-128M@73908c3438cf03b6a01bcb9611d62b23d0726f08",
            "modo": "afinado (router S10)", "piezas": len(r["piezas"]), "piezas_originales": r["originales"],
            "dim": int(filas.shape[1]), "cuant": "int8", "bytes": ruta.stat().st_size, "sha256": sha,
            "tabla_normalizacion": len(tabla)}, indent=2, ensure_ascii=False), encoding="utf-8")
        # paridad: 30 frases de calibración y 10 de prueba
        # va en git (repo público): solo frases sintéticas, nunca reales
        frases = [x["texto"] for x in calib if x.get("origen") != "real"][:30] + [x["texto"] for x in prueba[:10]]
        seqs = tokenizar(tok, unk, frases)
        P, U = prob_q(seqs), uso_q(seqs)
        with open(a.salida / "router-s10.paridad.jsonl", "w", encoding="utf-8") as f:
            for frase, s, p, us in zip(frases, seqs, P, U):
                f.write(json.dumps({"texto": frase, "emb": [round(float(x), 6) for x in vq(s)],
                                    "p": [round(float(x), 6) for x in p], "uso": round(float(us), 6)},
                                   ensure_ascii=False) + "\n")
        log(f"\nexportado {ruta} ({ruta.stat().st_size} bytes, sha256 {sha[:12]})")
        log(json.dumps({k: v for k, v in informe["afinado_int8"]["prueba"].items() if k != "conteos"},
                       ensure_ascii=False))

    (a.salida / "informe.json").write_text(json.dumps(informe, ensure_ascii=False, indent=1), encoding="utf-8")


if __name__ == "__main__":
    main()
