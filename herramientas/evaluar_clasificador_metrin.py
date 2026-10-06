#!/usr/bin/env python3
"""Entrena los clasificadores CANDIDATOS de Metrín con los catálogos YAML y los evalúa contra su prueba aparte.

    .venv/bin/python3 herramientas/evaluar_clasificador_metrin.py            # estático (mecanismo de Metrín) + MiniLM
    .venv/bin/python3 herramientas/evaluar_clasificador_metrin.py --sin-minilm

Qué hace:
  1. Convierte los YAML al formato del clasificador ({intención: [textos]}): intenciones.yml + ruido.yml (clase
     «ninguno») y, aparte, tipo_respuesta.yml.
  2. Entrena con el MISMO código que usa el servicio (rag-go/internal/clasificar.Entrenar sobre el embebedor estático
     metrin/modelos/potion-es-int8.pjge) mediante el arnés herramientas/clasificador_metrin/ (Go, módulo aparte que no
     toca metrin/). `rag entrenar-clasificador` no sirve para esto: su tabla `intencionDe` solo admite 8 categorías y
     las pliega en 4 clases.
  3. Puntúa la prueba aparte (que NUNCA entra al entrenamiento) con el modelo ACTUAL (defecto.json, embarcado) y con el
     CANDIDATO, con Clasificar y Puntajes de Go (kNN k=1: mejor coseno por intención).
  4. Elige el umbral de abstención CON LOS DATOS: máxima cobertura con precisión ≥ OBJETIVO_PRECISION en lo que se
     responde; si ningún umbral llega, el que maximiza aciertos − COSTE_ERROR × errores. Para no engañarse, la cifra
     honesta es la calibración cruzada (se elige el umbral con una mitad de la prueba y se mide en la otra, 10 veces).
  5. Guarda los candidatos (con el umbral elegido) en metrin/modelos/ y escribe kb/catalogos/EVALUACION.md
     (conserva el bloque de análisis escrito a mano entre <!-- analisis:inicio --> y <!-- analisis:fin -->)
     y kb/catalogos/evaluacion.json.
  6. Opcional: repite el mismo kNN con paraphrase-multilingual-MiniLM-L12-v2 (caché local de Hugging Face, inferencia
     en numpy: no hace falta torch ni red) para medir cuánto mejora un embebedor contextual.
"""
from __future__ import annotations

import argparse
import datetime as dt
import glob
import json
import math
import os
import re
import struct
import subprocess
import sys
import tempfile
import time

import numpy as np

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from validar_catalogos_metrin import CAT, RAIZ, clases_de, ejemplos_de, leer_yaml, ruido_de  # noqa: E402

ARNES = os.path.join(RAIZ, "herramientas", "clasificador_metrin")
MODELOS = os.path.join(RAIZ, "metrin", "modelos")
SALIDA_INT = os.path.join(MODELOS, "clasificador-candidato.json")
SALIDA_TIPO = os.path.join(MODELOS, "clasificador-tipo-respuesta-candidato.json")
INFORME = os.path.join(CAT, "EVALUACION.md")
INFORME_JSON = os.path.join(CAT, "evaluacion.json")

OBJETIVO_PRECISION = 0.90   # de lo que el clasificador se atreve a etiquetar, al menos el 90 % bien
COSTE_ERROR = 2.0           # respaldo si nadie llega al objetivo: un error cuesta el doble que una aclaración
NINGUNO = "ninguno"
ABST = "∅"                  # abstención (intenciones): el código pregunta con ACLARAR_TAREA
UNKNOWN = "UNKNOWN"         # en tipo_respuesta la abstención ES la clase UNKNOWN (pedir aclaración)

# Mapeo a las 4 clases del modelo ACTUAL (cmd/rag/cmd_clasificar.go: intencionDe). El actual no tiene clases de
# derivación ni de concepto: todo lo técnico, comercial o humano cae en «trabajo» (ruta segura: RAG).
GRUESA = {"saludo": "social", "despedida": "social", "agradecimiento": "social",
          "avance_paso": "ayuda", "pedir_aclaracion": "ayuda", "ambiguo": "ayuda", "feedback_respuesta": "ayuda",
          NINGUNO: "limite"}
# Ruta efectiva hoy (internal/rag/orquestador.go): solo social y limite desvían la búsqueda; ayuda y trabajo van al RAG.
RUTA_ACTUAL = {"social": "conversacion", "limite": "fuera_de_alcance", "ayuda": "rag", "trabajo": "rag"}
# Reglas actuales de tipo de consulta (internal/rag/orquestador.go: tipoConsulta) → clases de tipo_respuesta.
REGLA_A_TIPO = {"problema": "TROUBLESHOOTING", "comparacion": "COMPARISON", "procedimiento": "PROCEDURE",
                "concepto": "CONCEPT", "aclaracion": UNKNOWN, "informacion_directa": UNKNOWN}


def gruesa(etiqueta: str) -> str:
    if etiqueta in ("social", "trabajo", "ayuda", "limite"):
        return etiqueta
    return GRUESA.get(etiqueta, "trabajo")


# --- arnés Go --------------------------------------------------------------------------------------------------------
class ArnesGo:
    """Compila herramientas/clasificador_metrin y llama a entrenar/puntuar/embeber."""

    def __init__(self, tmp: str):
        self.tmp = tmp
        self.bin = os.path.join(tmp, "clasificadormetrin")
        r = subprocess.run(["go", "build", "-o", self.bin, "."], cwd=ARNES, capture_output=True, text=True)
        if r.returncode != 0:
            raise SystemExit(f"no compila el arnés Go:\n{r.stderr}")
        self.emb = os.path.join(RAIZ, "metrin", "modelos", "potion-es-int8.pjge")

    def _json(self, nombre, obj) -> str:
        ruta = os.path.join(self.tmp, nombre)
        with open(ruta, "w", encoding="utf-8") as fh:
            json.dump(obj, fh, ensure_ascii=False)
        return ruta

    def entrenar(self, ejemplos: dict, salida: str, umbral: float) -> str:
        r = subprocess.run([self.bin, "entrenar", "--ejemplos", self._json("ejemplos.json", ejemplos), "--salida", salida,
                            "--umbral", f"{umbral:.4f}", "--emb", self.emb], capture_output=True, text=True, cwd=ARNES)
        if r.returncode != 0:
            raise SystemExit(f"falló el entrenamiento:\n{r.stderr}")
        return r.stderr.strip()

    def puntuar(self, frases: list, clasificador: str | None = None) -> list:
        args = [self.bin, "puntuar", "--frases", self._json("frases.json", frases), "--emb", self.emb]
        if clasificador:
            args += ["--clasificador", clasificador]
        r = subprocess.run(args, capture_output=True, text=True, cwd=ARNES)
        if r.returncode != 0:
            raise SystemExit(f"falló puntuar:\n{r.stderr}")
        return [json.loads(l) for l in r.stdout.splitlines() if l.strip()]

    def embeber(self, frases: list) -> np.ndarray:
        r = subprocess.run([self.bin, "embeber", "--frases", self._json("frases.json", frases), "--emb", self.emb],
                           capture_output=True, text=True, cwd=ARNES)
        if r.returncode != 0:
            raise SystemExit(f"falló embeber:\n{r.stderr}")
        filas = [json.loads(l) for l in r.stdout.splitlines() if l.strip()]
        dim = next(len(f["vec"]) for f in filas if "vec" in f)
        return np.array([f.get("vec") or [0.0] * dim for f in filas], dtype=np.float64)


# --- MiniLM en numpy (opcional) ---------------------------------------------------------------------------------------
class MiniLM:
    """paraphrase-multilingual-MiniLM-L12-v2 (BertModel, 12 capas, 384) leído de la caché de Hugging Face, inferencia en
    numpy y media de tokens normalizada (el pooling de sentence-transformers para este modelo)."""

    nombre = "paraphrase-multilingual-MiniLM-L12-v2"

    def __init__(self):
        snaps = glob.glob(os.path.expanduser(
            "~/.cache/huggingface/hub/models--sentence-transformers--paraphrase-multilingual-MiniLM-L12-v2/snapshots/*/"))
        if not snaps:
            raise FileNotFoundError("MiniLM no está en la caché local de Hugging Face")
        from tokenizers import Tokenizer
        d = snaps[0]
        self.tok = Tokenizer.from_file(os.path.join(d, "tokenizer.json"))
        self.tok.enable_truncation(128)
        self.cfg = json.load(open(os.path.join(d, "config.json")))
        self.w = self._safetensors(os.path.join(d, "model.safetensors"))

    @staticmethod
    def _safetensors(ruta):
        with open(ruta, "rb") as fh:
            n = struct.unpack("<Q", fh.read(8))[0]
            cab = json.loads(fh.read(n))
        datos = np.memmap(ruta, dtype=np.uint8, mode="r", offset=8 + n)
        tipos = {"F32": np.float32, "F16": np.float16, "I64": np.int64}
        out = {}
        for k, v in cab.items():
            if k == "__metadata__":
                continue
            a, b = v["data_offsets"]
            out[k] = np.frombuffer(datos[a:b], dtype=tipos[v["dtype"]]).reshape(v["shape"])
        return out

    @staticmethod
    def _ln(x, g, b, eps=1e-12):
        mu = x.mean(-1, keepdims=True)
        var = ((x - mu) ** 2).mean(-1, keepdims=True)
        return (x - mu) / np.sqrt(var + eps) * g + b

    @staticmethod
    def _gelu(x):  # gelu exacta con erf (Abramowitz-Stegun 7.1.26, error < 1,5e-7)
        z = x / math.sqrt(2.0)
        s = np.sign(z)
        a = np.abs(z)
        t = 1.0 / (1.0 + 0.3275911 * a)
        y = 1.0 - (((((1.061405429 * t - 1.453152027) * t) + 1.421413741) * t - 0.284496736) * t + 0.254829592) * t * np.exp(-a * a)
        return 0.5 * x * (1.0 + s * y)

    def __call__(self, textos: list, lote: int = 64) -> np.ndarray:
        w, H = self.w, self.cfg["num_attention_heads"]
        out = []
        for i in range(0, len(textos), lote):
            # tokenizer.json trae el relleno activado (<pad>): la máscara sale de attention_mask, no de len(ids)
            encs = self.tok.encode_batch([t.strip() for t in textos[i:i + lote]])
            L = max(len(e.ids) for e in encs)
            ids = np.full((len(encs), L), self.tok.padding["pad_id"] if self.tok.padding else 1, dtype=np.int64)
            m = np.zeros((len(encs), L), dtype=np.float32)
            for j, e in enumerate(encs):
                ids[j, :len(e.ids)] = e.ids
                m[j, :len(e.attention_mask)] = e.attention_mask
            x = (w["embeddings.word_embeddings.weight"][ids] + w["embeddings.position_embeddings.weight"][:L][None]
                 + w["embeddings.token_type_embeddings.weight"][0][None, None]).astype(np.float32)
            x = self._ln(x, w["embeddings.LayerNorm.weight"], w["embeddings.LayerNorm.bias"])
            B, D = x.shape[0], x.shape[2]
            hd = D // H
            sesgo = ((1.0 - m) * -1e9)[:, None, None, :]
            for c in range(self.cfg["num_hidden_layers"]):
                p = f"encoder.layer.{c}."
                lin = lambda t, n: t @ w[p + n + ".weight"].T + w[p + n + ".bias"]  # noqa: E731
                q = lin(x, "attention.self.query").reshape(B, L, H, hd).transpose(0, 2, 1, 3)
                k = lin(x, "attention.self.key").reshape(B, L, H, hd).transpose(0, 2, 1, 3)
                v = lin(x, "attention.self.value").reshape(B, L, H, hd).transpose(0, 2, 1, 3)
                s = q @ k.transpose(0, 1, 3, 2) / math.sqrt(hd) + sesgo
                s = np.exp(s - s.max(-1, keepdims=True))
                s /= s.sum(-1, keepdims=True)
                ctx = (s @ v).transpose(0, 2, 1, 3).reshape(B, L, D)
                x = self._ln(lin(ctx, "attention.output.dense") + x, w[p + "attention.output.LayerNorm.weight"],
                             w[p + "attention.output.LayerNorm.bias"])
                h = self._gelu(lin(x, "intermediate.dense"))
                x = self._ln(lin(h, "output.dense") + x, w[p + "output.LayerNorm.weight"], w[p + "output.LayerNorm.bias"])
            e = (x * m[..., None]).sum(1) / m.sum(1, keepdims=True)
            out.append(e / np.linalg.norm(e, axis=1, keepdims=True))
        return np.vstack(out).astype(np.float64)


# --- kNN igual al de Go (Puntajes con ejemplos: mejor coseno por intención) -------------------------------------------
def puntajes_knn(E: np.ndarray, y: list, nombres: list, Q: np.ndarray) -> np.ndarray:
    S = Q @ E.T
    y = np.array(y)
    return np.stack([S[:, y == n].max(axis=1) for n in nombres], axis=1)


def puntajes_centroide(E: np.ndarray, y: list, nombres: list, Q: np.ndarray) -> np.ndarray:
    """Lo que hace Puntajes cuando el modelo NO trae ejemplos: coseno con el centroide normalizado de cada intención."""
    y = np.array(y)
    C = np.stack([E[y == n].mean(axis=0) for n in nombres])
    C /= np.linalg.norm(C, axis=1, keepdims=True)
    return Q @ C.T


def loo(E: np.ndarray, y: list, nombres: list, metodo: str) -> np.ndarray:
    """Leave-one-out sobre el ENTRENAMIENTO (sin mirar la prueba): cada ejemplo se puntúa sin contarse a sí mismo."""
    y = np.array(y)
    if metodo == "knn1":
        S = E @ E.T
        np.fill_diagonal(S, -np.inf)
        return np.stack([S[:, y == n].max(axis=1) for n in nombres], axis=1)
    sumas = np.stack([E[y == n].sum(axis=0) for n in nombres])
    cuentas = np.array([(y == n).sum() for n in nombres], dtype=float)
    C = sumas / np.linalg.norm(sumas, axis=1, keepdims=True)
    P = E @ C.T
    idx = {n: j for j, n in enumerate(nombres)}
    for i, n in enumerate(y):
        j = idx[n]
        if cuentas[j] > 1:
            c = sumas[j] - E[i]
            P[i, j] = float(E[i] @ (c / np.linalg.norm(c)))
    return P


METODOS_GO = {"knn1": "kNN k=1 (modelo con `ejemplos`)", "centroide": "centroides (modelo sin `ejemplos`)"}


def elegir_metodo(E: np.ndarray, y: list, nombres: list) -> tuple:
    """Elige entre los dos modos que Clasificar ya soporta por macro-F1 leave-one-out del entrenamiento."""
    res = {}
    for m in METODOS_GO:
        P = loo(E, y, nombres, m)
        pred = [nombres[i] for i in P.argmax(axis=1)]
        res[m] = {"loo_exactitud": float(np.mean([a == b for a, b in zip(y, pred)])),
                  "loo_macro_f1": macro_f1(prf(list(y), pred, nombres))}
    return max(res, key=lambda m: res[m]["loo_macro_f1"]), res


def matriz_desde_go(filas: list, nombres: list) -> np.ndarray:
    P = np.full((len(filas), len(nombres)), -1.0)
    idx = {n: i for i, n in enumerate(nombres)}
    for i, f in enumerate(filas):
        for c in f.get("candidatas") or []:
            if c["intencion"] in idx:
                P[i, idx[c["intencion"]]] = c["similitud"]
    return P


# --- métricas -------------------------------------------------------------------------------------------------------
def prf(y_true: list, y_pred: list, clases: list) -> dict:
    out = {}
    for c in clases:
        tp = sum(1 for t, p in zip(y_true, y_pred) if t == c and p == c)
        fp = sum(1 for t, p in zip(y_true, y_pred) if t != c and p == c)
        fn = sum(1 for t, p in zip(y_true, y_pred) if t == c and p != c)
        pr = tp / (tp + fp) if tp + fp else 0.0
        rc = tp / (tp + fn) if tp + fn else 0.0
        out[c] = {"precision": pr, "recall": rc, "f1": 2 * pr * rc / (pr + rc) if pr + rc else 0.0, "n": tp + fn}
    return out


def macro_f1(m: dict) -> float:
    return float(np.mean([v["f1"] for v in m.values() if v["n"]])) if m else 0.0


def aplicar_umbral(pred: list, s1: np.ndarray, tau: float, abst: str) -> list:
    return [p if s >= tau else abst for p, s in zip(pred, s1)]


def cobertura_precision(y: list, pred_t: list, abst: str) -> tuple:
    resp = [p != abst for p in pred_t]
    n = sum(resp)
    ok = sum(1 for t, p, r in zip(y, pred_t, resp) if r and t == p)
    return (n / len(y) if y else 0.0), (ok / n if n else float("nan")), ok, n - ok


def elegir_umbral(y: list, pred: list, s1: np.ndarray, abst: str) -> dict:
    """Criterio de coste: el umbral que maximiza aciertos − COSTE_ERROR × errores entre lo que se etiqueta (abstenerse
    vale 0: una pregunta de aclaración). Equivale a etiquetar mientras la precisión marginal supere
    COSTE_ERROR/(1+COSTE_ERROR). Se informa además el punto de «máxima cobertura con precisión ≥ OBJETIVO_PRECISION»."""
    cands = np.unique(np.round(np.concatenate([[0.0], s1]), 4))
    mejor, punto_p = None, None
    for tau in cands:
        cob, prec, ok, mal = cobertura_precision(y, aplicar_umbral(pred, s1, tau, abst), abst)
        u = ok - COSTE_ERROR * mal
        if mejor is None or u > mejor["utilidad"] + 1e-12:
            mejor = {"umbral": float(tau), "cobertura": cob, "precision": prec, "utilidad": u}
        if not math.isnan(prec) and prec >= OBJETIVO_PRECISION and (punto_p is None or cob > punto_p["cobertura"]):
            punto_p = {"umbral": float(tau), "cobertura": cob, "precision": prec}
    mejor["criterio"] = f"máx. aciertos − {COSTE_ERROR:g}×errores"
    mejor["punto_precision_objetivo"] = punto_p
    return mejor


def calibracion_cruzada(y: list, pred: list, s1: np.ndarray, abst: str, repeticiones: int = 10) -> dict:
    y_arr = np.array(y)
    res = []
    for semilla in range(repeticiones):
        rng = np.random.default_rng(semilla)
        mitad = np.zeros(len(y), dtype=bool)
        for c in np.unique(y_arr):
            idx = np.where(y_arr == c)[0]
            rng.shuffle(idx)
            mitad[idx[: len(idx) // 2]] = True
        for a in (True, False):
            A, B = np.where(mitad == a)[0], np.where(mitad != a)[0]
            u = elegir_umbral([y[i] for i in A], [pred[i] for i in A], s1[A], abst)
            cob, prec, _, _ = cobertura_precision([y[i] for i in B], aplicar_umbral([pred[i] for i in B], s1[B], u["umbral"], abst), abst)
            res.append((u["umbral"], cob, prec))
    r = np.array(res, dtype=float)
    return {"umbral_medio": float(np.mean(r[:, 0])), "umbral_sd": float(np.std(r[:, 0])),
            "cobertura": float(np.mean(r[:, 1])), "precision": float(np.nanmean(r[:, 2])),
            "precision_sd": float(np.nanstd(r[:, 2])), "n": len(res)}


def curva(y: list, pred: list, s1: np.ndarray, abst: str, taus) -> list:
    filas = []
    for tau in taus:
        pt = aplicar_umbral(pred, s1, tau, abst)
        cob, prec, _, _ = cobertura_precision(y, pt, abst)
        filas.append((float(tau), cob, prec))
    return filas


def evaluar_matriz(P: np.ndarray, nombres: list, y: list, abst: str) -> dict:
    """Todo lo que se reporta de un clasificador sobre una prueba: P = puntajes (n × clases)."""
    orden = np.argsort(-P, axis=1)
    pred = [nombres[i] for i in orden[:, 0]]
    s1 = P[np.arange(len(P)), orden[:, 0]]
    s2 = P[np.arange(len(P)), orden[:, 1]]
    top3 = float(np.mean([y[i] in [nombres[j] for j in orden[i, :3]] for i in range(len(y))]))
    clases = sorted(set(y) | set(nombres))
    bruto = prf(y, pred, [c for c in nombres])
    eleg = elegir_umbral(y, pred, s1, abst)
    pt = aplicar_umbral(pred, s1, eleg["umbral"], abst)
    con = prf(y, pt, [c for c in nombres])
    return {"pred": pred, "s1": s1, "margen": s1 - s2, "top3": top3, "orden": orden,
            "exactitud": float(np.mean([a == b for a, b in zip(y, pred)])), "prf": bruto, "macro_f1": macro_f1(bruto),
            "umbral": eleg, "pred_umbral": pt, "prf_umbral": con, "macro_f1_umbral": macro_f1(con),
            "exactitud_umbral": float(np.mean([a == b for a, b in zip(y, pt)])),
            "cruzada": calibracion_cruzada(y, pred, s1, abst),
            "curva": curva(y, pred, s1, abst, [round(x, 2) for x in np.arange(0.30, 0.96, 0.05)] + [eleg["umbral"]]),
            "clases": clases}


# --- reglas actuales de tipo de consulta (port fiel de internal/rag/orquestador.go, sin hilo) --------------------------
def _norm_consulta(s: str) -> str:
    s = " ".join(s.split()).lower()
    for a, b in (("á", "a"), ("é", "e"), ("í", "i"), ("ó", "o"), ("ú", "u"), ("ü", "u")):
        s = s.replace(a, b)
    for c in "¿?¡!,.:;":
        s = s.replace(c, " ")
    return s


def _es_how_to(p: str) -> bool:
    p = " " + p.lower() + " "
    return any(w in p for w in ["cómo", "como ", "pasos", "paso a paso", "crear", "calcular", "guía", "guia", "tutorial",
                                "cómo se", "como se", "ayúdame a", "ayudame a", "enséñame", "ensename a", "que debo hacer",
                                "qué debo hacer", "dónde", "donde", "muéstrame", "muestrame", "ubicar", "encontrar",
                                "me guías", "me guias"])


def tipo_consulta_reglas(pregunta: str) -> str:
    t = " " + _norm_consulta(pregunta).strip() + " "
    alguna = lambda *fs: any(f in t for f in fs)  # noqa: E731
    if alguna("error", "falla", "fallo", "problema", "no funciona", "no aparece", "no permite", "se bloquea", "no carga"):
        return "problema"
    if alguna("diferencia", "comparar", "compara", "comparacion", "versus", " frente a ", "cual conviene", "cual es mejor"):
        return "comparacion"
    if _es_how_to(t):
        return "procedimiento"
    if alguna("que es", "que significa", "para que sirve", "en que consiste", "define", "definicion", "concepto de"):
        return "concepto"
    if alguna("no entiendo", "no me queda claro", "explicame de otra forma", "reformula", "mas sencillo", "mas simple"):
        return "aclaracion"
    return "informacion_directa"


# --- informe ----------------------------------------------------------------------------------------------------------
def pct(x) -> str:
    return "—" if x is None or (isinstance(x, float) and math.isnan(x)) else f"{100 * x:.1f} %"


def tabla_prf(m: dict, extra: dict | None = None) -> list:
    filas = ["| Clase | n | Precisión | Recall | F1 |" + (" " + " | ".join(extra["cab"]) + " |" if extra else ""),
             "|---|---:|---:|---:|---:|" + ("---:|" * len(extra["cab"]) if extra else "")]
    for c, v in sorted(m.items(), key=lambda kv: kv[1]["f1"]):
        if not v["n"]:
            continue
        fila = f"| {c} | {v['n']} | {pct(v['precision'])} | {pct(v['recall'])} | {pct(v['f1'])} |"
        if extra:
            fila += " " + " | ".join(extra["filas"].get(c, ["—"] * len(extra["cab"]))) + " |"
        filas.append(fila)
    return filas


def matriz_confusion(y: list, pred: list, filas_c: list, cols_c: list) -> list:
    corto = {c: c[:5] for c in cols_c}
    out = ["| real \\ predicha | " + " | ".join(corto[c] for c in cols_c) + " |", "|---|" + "---:|" * len(cols_c)]
    for r in filas_c:
        cuenta = [sum(1 for t, p in zip(y, pred) if t == r and p == c) for c in cols_c]
        out.append(f"| {r} | " + " | ".join(f"**{v}**" if c == r else (str(v) if v else "·") for v, c in zip(cuenta, cols_c)) + " |")
    return out


def confusiones_resumen(y: list, pred: list, n: int = 12) -> list:
    from collections import Counter
    c = Counter((t, p) for t, p in zip(y, pred) if t != p)
    return [f"{a} → {b} ×{v}" for (a, b), v in c.most_common(n)]


def punto_txt(u: dict) -> str:
    p = u.get("punto_precision_objetivo")
    if not p:
        return f"Ningún umbral llega a precisión ≥ {OBJETIVO_PRECISION:.0%}."
    return (f"Referencia: la máxima cobertura con precisión ≥ {OBJETIVO_PRECISION:.0%} se logra con umbral {p['umbral']:.3f} "
            f"y cubre solo {pct(p['cobertura'])}.")


def tabla_metodos(metodos: dict) -> list:
    nombres = {"int": "Intenciones · estático", "tipo": "Tipo de respuesta · estático",
               "int_minilm": "Intenciones · MiniLM", "tipo_minilm": "Tipo de respuesta · MiniLM"}
    out = ["| Catálogo · embebedor | Modo | Macro-F1 LOO (entren.) | Exactitud prueba | Macro-F1 prueba | Umbral | Cruzada cob. / prec. |",
           "|---|---|---:|---:|---:|---:|---:|"]
    for k, d in metodos.items():
        mejor = max(d, key=lambda m: d[m]["loo_macro_f1"])
        for m, v in d.items():
            e = v["ev"]
            out.append(f"| {nombres.get(k, k)} | {m}{' ✓' if m == mejor else ''} | {pct(v['loo_macro_f1'])} | {pct(e['exactitud'])} | "
                       f"{pct(e['macro_f1'])} | {e['umbral']['umbral']:.3f} | {pct(e['cruzada']['cobertura'])} / {pct(e['cruzada']['precision'])} |")
    return out


def preservar_analisis(texto_nuevo: str) -> str:
    ini, fin = "<!-- analisis:inicio -->", "<!-- analisis:fin -->"
    if os.path.exists(INFORME):
        viejo = open(INFORME, encoding="utf-8").read()
        m = re.search(re.escape(ini) + r".*?" + re.escape(fin), viejo, re.S)
        if m:
            return re.sub(re.escape(ini) + r".*?" + re.escape(fin), lambda _: m.group(0), texto_nuevo, flags=re.S)
    return texto_nuevo


# --- principal --------------------------------------------------------------------------------------------------------
def cargar_datos():
    di = leer_yaml(os.path.join(CAT, "intenciones.yml"))
    dr = leer_yaml(os.path.join(CAT, "ruido.yml"))
    dp = leer_yaml(os.path.join(CAT, "intenciones_prueba.yml"))
    dt_ = leer_yaml(os.path.join(CAT, "tipo_respuesta.yml"))
    dtp = leer_yaml(os.path.join(CAT, "tipo_respuesta_prueba.yml"))
    ent_int = ejemplos_de(di)
    ent_int[NINGUNO] = ruido_de(dr)
    prueba_int = [(f, c) for c, fs in ejemplos_de(dp).items() for f in fs]
    ent_tipo = ejemplos_de(dt_)
    prueba_tipo = [(f, c) for c, fs in ejemplos_de(dtp).items() for f in fs]
    return ent_int, prueba_int, ent_tipo, prueba_tipo, di


def main(argv=None) -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--sin-minilm", action="store_true", help="no medir el embebedor MiniLM")
    a = ap.parse_args(argv)
    t0 = time.time()
    ent_int, prueba_int, ent_tipo, prueba_tipo, di = cargar_datos()
    R = {"fecha": dt.date.today().isoformat(), "objetivo_precision": OBJETIVO_PRECISION, "coste_error": COSTE_ERROR}
    nombres_int, nombres_tipo = sorted(ent_int), sorted(ent_tipo)
    frases_int, y_int = [f for f, _ in prueba_int], [c for _, c in prueba_int]
    frases_tipo, y_tipo = [f for f, _ in prueba_tipo], [c for _, c in prueba_tipo]
    tx_int = [t for n in nombres_int for t in ent_int[n]]
    yy_int = [n for n in nombres_int for _ in ent_int[n]]
    tx_tipo = [t for n in nombres_tipo for t in ent_tipo[n]]
    yy_tipo = [n for n in nombres_tipo for _ in ent_tipo[n]]
    metodos = {}
    with tempfile.TemporaryDirectory() as tmp:
        go = ArnesGo(tmp)
        # Vectores estáticos de Go para elegir el modo por leave-one-out (sin mirar la prueba) y para medir los dos modos.
        E_int, Q_int = go.embeber(tx_int), go.embeber(frases_int)
        E_tipo, Q_tipo = go.embeber(tx_tipo), go.embeber(frases_tipo)
        modo_int, loo_int = elegir_metodo(E_int, yy_int, nombres_int)
        modo_tipo, loo_tipo = elegir_metodo(E_tipo, yy_tipo, nombres_tipo)
        for clave, (E, yy, nombres, Q, y, abst, loo_r) in {
                "int": (E_int, yy_int, nombres_int, Q_int, y_int, ABST, loo_int),
                "tipo": (E_tipo, yy_tipo, nombres_tipo, Q_tipo, y_tipo, UNKNOWN, loo_tipo)}.items():
            metodos[clave] = {m: {**loo_r[m], "ev": evaluar_matriz(
                (puntajes_knn if m == "knn1" else puntajes_centroide)(E, yy, nombres, Q), nombres, y, abst)}
                for m in METODOS_GO}

        def guardar(ent, salida, modo, tau):
            log = go.entrenar(ent, salida, tau)
            if modo == "centroide":            # sin ejemplos, Puntajes usa los centroides (clasificar.go)
                with open(salida, encoding="utf-8") as fh:
                    mod = json.load(fh)
                mod["ejemplos"] = []
                with open(salida, "w", encoding="utf-8") as fh:
                    json.dump(mod, fh, ensure_ascii=False, indent=2)
                    fh.write("\n")
                log += " (guardado en modo centroides: sin `ejemplos`)"
            return log

        # ---------- intenciones: candidato con el mecanismo de Metrín, en el modo elegido ----------
        tau_int = metodos["int"][modo_int]["ev"]["umbral"]["umbral"]
        log_ent = guardar(ent_int, SALIDA_INT, modo_int, tau_int)
        filas_cand = go.puntuar(frases_int, SALIDA_INT)            # Clasificar/Puntajes de Go con el umbral elegido
        filas_act = go.puntuar(frases_int)                          # defecto.json (ACTUAL)
        P_go = matriz_desde_go(filas_cand, nombres_int)
        ev_int = evaluar_matriz(P_go, nombres_int, y_int, ABST)
        paridad = float(np.max(np.abs(P_go - (puntajes_knn if modo_int == "knn1" else puntajes_centroide)(
            E_int, yy_int, nombres_int, Q_int))))
        # ---------- tipo de respuesta ----------
        tau_tipo = metodos["tipo"][modo_tipo]["ev"]["umbral"]["umbral"]
        guardar(ent_tipo, SALIDA_TIPO, modo_tipo, tau_tipo)
        filas_tipo = go.puntuar(frases_tipo, SALIDA_TIPO)
        ev_tipo = evaluar_matriz(matriz_desde_go(filas_tipo, nombres_tipo), nombres_tipo, y_tipo, UNKNOWN)
        reglas = [REGLA_A_TIPO[tipo_consulta_reglas(f)] for f in frases_tipo]
        prf_reglas = prf(y_tipo, reglas, nombres_tipo)
    R["paridad_knn_python_vs_go"] = paridad
    R["modo_int"], R["modo_tipo"] = modo_int, modo_tipo

    # ---------- comparación ACTUAL vs CANDIDATO (4 clases del actual y ruta efectiva) ----------
    act = [f["intencion"] for f in filas_act]
    cand = [f["intencion"] for f in filas_cand]
    esperada_g = [gruesa(c) for c in y_int]
    act_g = [gruesa(c) for c in act]
    cand_g = [gruesa(c) for c in cand]
    clases_g = ["social", "ayuda", "limite", "trabajo"]

    def exact(a_, b_):
        return float(np.mean([x == z for x, z in zip(a_, b_)]))

    comp = {"gruesa_actual": exact(esperada_g, act_g), "gruesa_candidato": exact(esperada_g, cand_g),
            "ruta_actual": exact([RUTA_ACTUAL[g] for g in esperada_g], [RUTA_ACTUAL[g] for g in act_g]),
            "ruta_candidato": exact([RUTA_ACTUAL[g] for g in esperada_g], [RUTA_ACTUAL[g] for g in cand_g]),
            "prf_actual": prf(esperada_g, act_g, clases_g), "prf_candidato": prf(esperada_g, cand_g, clases_g),
            "trabajo_por_abstencion_actual": sum(1 for f, e in zip(filas_act, esperada_g)
                                                 if f["intencion"] == "trabajo" and e == "trabajo" and f["similitud"] < 0.40),
            "fina_candidato_clasificar": exact(y_int, cand)}
    casos = {}
    for texto in ("¿qué es un metrado?", "¿Cómo registro un nuevo presupuesto en S10?"):
        i = frases_int.index(texto)
        casos[texto] = {"esperado": y_int[i],
                        "actual": {"intencion": filas_act[i]["intencion"], "similitud": filas_act[i]["similitud"],
                                   "top3": [(c["intencion"], round(c["similitud"], 3)) for c in filas_act[i]["candidatas"][:3]]},
                        "candidato": {"intencion": filas_cand[i]["intencion"], "similitud": filas_cand[i]["similitud"],
                                      "top3": [(c["intencion"], round(c["similitud"], 3)) for c in filas_cand[i]["candidatas"][:3]]}}
    literales = {}
    for texto in ("como hago un metrado", "como registro metrado", "cómo modifico una partida", "que es metrado",
                  "no puedo guardar metrado", "donde veo metrados", "metrado s10"):
        i = frases_tipo.index(texto)
        literales[texto] = {"esperado": y_tipo[i], "estatico": (ev_tipo["pred"][i], round(float(ev_tipo["s1"][i]), 3),
                                                                 ev_tipo["pred_umbral"][i]), "reglas": reglas[i]}

    # ---------- MiniLM ----------
    mini = None
    if not a.sin_minilm:
        try:
            t1 = time.time()
            m = MiniLM()
            Ei, Qi, Et, Qt = m(tx_int), m(frases_int), m(tx_tipo), m(frases_tipo)
            mi_modo_int, mi_loo_int = elegir_metodo(Ei, yy_int, nombres_int)
            mi_modo_tipo, mi_loo_tipo = elegir_metodo(Et, yy_tipo, nombres_tipo)
            fn = {"knn1": puntajes_knn, "centroide": puntajes_centroide}
            ev_mi = evaluar_matriz(fn[mi_modo_int](Ei, yy_int, nombres_int, Qi), nombres_int, y_int, ABST)
            ev_mt = evaluar_matriz(fn[mi_modo_tipo](Et, yy_tipo, nombres_tipo, Qt), nombres_tipo, y_tipo, UNKNOWN)
            for clave, (E, yy, nombres, Q, y, abst, loo_r) in {
                    "int_minilm": (Ei, yy_int, nombres_int, Qi, y_int, ABST, mi_loo_int),
                    "tipo_minilm": (Et, yy_tipo, nombres_tipo, Qt, y_tipo, UNKNOWN, mi_loo_tipo)}.items():
                metodos[clave] = {mm: {**loo_r[mm], "ev": evaluar_matriz(fn[mm](E, yy, nombres, Q), nombres, y, abst)}
                                  for mm in METODOS_GO}
            for texto in literales:
                i = frases_tipo.index(texto)
                literales[texto]["minilm"] = (ev_mt["pred"][i], round(float(ev_mt["s1"][i]), 3), ev_mt["pred_umbral"][i])
            for texto in casos:
                i = frases_int.index(texto)
                casos[texto]["minilm"] = {"intencion": ev_mi["pred"][i], "similitud": float(ev_mi["s1"][i]),
                                          "con_umbral": ev_mi["pred_umbral"][i]}
            mini = {"int": ev_mi, "tipo": ev_mt, "segundos": time.time() - t1, "modo_int": mi_modo_int, "modo_tipo": mi_modo_tipo}
        except Exception as e:  # no es obligatorio: se informa y se sigue
            mini = {"error": f"{type(e).__name__}: {e}"}

    # ---------- informe ----------
    md = escribir_informe(R, di, ent_int, prueba_int, ent_tipo, prueba_tipo, ev_int, ev_tipo, comp, casos, literales,
                          reglas, prf_reglas, y_int, y_tipo, act_g, cand_g, esperada_g, mini, log_ent, nombres_int, nombres_tipo,
                          metodos)
    md = preservar_analisis(md)          # leer el análisis ANTES de abrir el informe para escribir (lo truncaría)
    with open(INFORME, "w", encoding="utf-8") as fh:
        fh.write(md)
    resumen = {
        **R,
        "intenciones": {"exactitud": ev_int["exactitud"], "macro_f1": ev_int["macro_f1"], "top3": ev_int["top3"],
                        "umbral": ev_int["umbral"], "cruzada": ev_int["cruzada"], "macro_f1_umbral": ev_int["macro_f1_umbral"],
                        "por_clase": ev_int["prf"]},
        "tipo_respuesta": {"exactitud": ev_tipo["exactitud"], "macro_f1": ev_tipo["macro_f1"], "umbral": ev_tipo["umbral"],
                           "cruzada": ev_tipo["cruzada"], "macro_f1_umbral": ev_tipo["macro_f1_umbral"],
                           "por_clase": ev_tipo["prf_umbral"], "reglas_actuales": prf_reglas},
        "actual_vs_candidato": {k: v for k, v in comp.items() if not k.startswith("prf")},
        "metodos": {k: {m: {"loo_macro_f1": v["loo_macro_f1"], "exactitud": v["ev"]["exactitud"], "macro_f1": v["ev"]["macro_f1"],
                            "umbral": v["ev"]["umbral"], "cruzada": v["ev"]["cruzada"]} for m, v in d.items()}
                    for k, d in metodos.items()},
        "casos_medidos": casos, "literales_tipo_respuesta": literales,
        "minilm": None if not mini else ({"error": mini["error"]} if "error" in mini else {
            "intenciones": {"exactitud": mini["int"]["exactitud"], "macro_f1": mini["int"]["macro_f1"],
                            "umbral": mini["int"]["umbral"], "cruzada": mini["int"]["cruzada"]},
            "tipo_respuesta": {"exactitud": mini["tipo"]["exactitud"], "macro_f1": mini["tipo"]["macro_f1"],
                               "umbral": mini["tipo"]["umbral"], "cruzada": mini["tipo"]["cruzada"],
                               "macro_f1_umbral": mini["tipo"]["macro_f1_umbral"]}}),
    }
    with open(INFORME_JSON, "w", encoding="utf-8") as fh:
        json.dump(resumen, fh, ensure_ascii=False, indent=1, default=lambda o: o.item() if hasattr(o, "item") else str(o))
    print(f"intenciones: exactitud {pct(ev_int['exactitud'])}, macro-F1 {pct(ev_int['macro_f1'])}, umbral {ev_int['umbral']['umbral']:.3f} "
          f"→ cobertura {pct(ev_int['umbral']['cobertura'])} / precisión {pct(ev_int['umbral']['precision'])}; cruzada "
          f"{pct(ev_int['cruzada']['cobertura'])} / {pct(ev_int['cruzada']['precision'])}")
    print(f"tipo_respuesta: exactitud {pct(ev_tipo['exactitud'])}, macro-F1 {pct(ev_tipo['macro_f1'])}, umbral "
          f"{ev_tipo['umbral']['umbral']:.3f}; reglas actuales macro-F1 {pct(macro_f1(prf_reglas))}")
    print(f"ACTUAL vs CANDIDATO (4 clases): {pct(comp['gruesa_actual'])} vs {pct(comp['gruesa_candidato'])}; ruta "
          f"{pct(comp['ruta_actual'])} vs {pct(comp['ruta_candidato'])}")
    for t, c in casos.items():
        print(f"  «{t}»: esperado {c['esperado']}; actual {c['actual']['intencion']} ({c['actual']['similitud']:.3f}); "
              f"candidato {c['candidato']['intencion']} ({c['candidato']['similitud']:.3f})")
    if mini and "error" not in mini:
        print(f"MiniLM: intenciones {pct(mini['int']['exactitud'])} / macro-F1 {pct(mini['int']['macro_f1'])}; tipo "
              f"{pct(mini['tipo']['exactitud'])} / macro-F1 {pct(mini['tipo']['macro_f1'])}")
    elif mini:
        print("MiniLM no disponible:", mini["error"])
    print(f"paridad kNN Python vs Go: {paridad:.2e}; escrito {INFORME}, {INFORME_JSON}, {SALIDA_INT}, {SALIDA_TIPO} "
          f"({time.time() - t0:.0f} s)")
    return 0


def escribir_informe(R, di, ent_int, prueba_int, ent_tipo, prueba_tipo, ev_int, ev_tipo, comp, casos, literales, reglas,
                     prf_reglas, y_int, y_tipo, act_g, cand_g, esperada_g, mini, log_ent, nombres_int, nombres_tipo,
                     metodos) -> str:
    L = []
    w = L.append
    ok_mini = mini and "error" not in mini
    w("# Evaluación de los clasificadores de Metrín (catálogos YAML)")
    w("")
    w(f"Generado por `herramientas/evaluar_clasificador_metrin.py` el {R['fecha']}. Las tablas se regeneran al volver a "
      "correrlo; el bloque «Análisis» está escrito a mano y se conserva.")
    w("")
    w("## Resumen")
    w("")
    ui, ut = ev_int["umbral"], ev_tipo["umbral"]
    w("| | Intenciones (candidato) | Tipo de respuesta (candidato) |")
    w("|---|---:|---:|")
    w(f"| Clases / ejemplos de entrenamiento | {len(ent_int)} / {sum(map(len, ent_int.values()))} | {len(ent_tipo)} / {sum(map(len, ent_tipo.values()))} |")
    w(f"| Frases de prueba aparte | {len(prueba_int)} | {len(prueba_tipo)} |")
    w(f"| Modo de puntuación (elegido por leave-one-out del entrenamiento) | {R['modo_int']} | {R['modo_tipo']} |")
    w(f"| Exactitud top-1 sin umbral | {pct(ev_int['exactitud'])} | {pct(ev_tipo['exactitud'])} |")
    w(f"| Top-3 | {pct(ev_int['top3'])} | {pct(ev_tipo['top3'])} |")
    w(f"| Macro-F1 sin umbral | {pct(ev_int['macro_f1'])} | {pct(ev_tipo['macro_f1'])} |")
    w(f"| Umbral elegido con los datos | {ui['umbral']:.3f} | {ut['umbral']:.3f} |")
    w(f"| Cobertura / precisión con ese umbral (en muestra) | {pct(ui['cobertura'])} / {pct(ui['precision'])} | {pct(ut['cobertura'])} / {pct(ut['precision'])} |")
    w(f"| **Calibración cruzada** cobertura / precisión | **{pct(ev_int['cruzada']['cobertura'])} / {pct(ev_int['cruzada']['precision'])}** | "
      f"**{pct(ev_tipo['cruzada']['cobertura'])} / {pct(ev_tipo['cruzada']['precision'])}** |")
    w(f"| Macro-F1 con umbral | {pct(ev_int['macro_f1_umbral'])} | {pct(ev_tipo['macro_f1_umbral'])} |")
    if ok_mini:
        w(f"| Exactitud / macro-F1 con MiniLM (mismo kNN) | {pct(mini['int']['exactitud'])} / {pct(mini['int']['macro_f1'])} | "
          f"{pct(mini['tipo']['exactitud'])} / {pct(mini['tipo']['macro_f1'])} |")
    w("")
    w(f"ACTUAL (`defecto.json`, 4 clases) frente a CANDIDATO, en las {len(prueba_int)} frases de prueba de intenciones "
      f"llevadas a las 4 clases del actual: **{pct(comp['gruesa_actual'])} → {pct(comp['gruesa_candidato'])}**; en la ruta "
      f"efectiva de hoy (conversación / fuera de alcance / RAG): {pct(comp['ruta_actual'])} → {pct(comp['ruta_candidato'])}.")
    w("")
    w("Casos medidos:")
    w("")
    w("| Frase | Esperado | ACTUAL | CANDIDATO (estático)" + (" | MiniLM" if ok_mini else "") + " |")
    w("|---|---|---|---|" + ("---|" if ok_mini else ""))
    for t, c in casos.items():
        fila = (f"| «{t}» | {c['esperado']} | {c['actual']['intencion']} ({c['actual']['similitud']:.3f}) | "
                f"{c['candidato']['intencion']} ({c['candidato']['similitud']:.3f}) |")
        if ok_mini:
            fila += f" {c['minilm']['intencion']} ({c['minilm']['similitud']:.3f}) |"
        w(fila)
    w("")
    w("<!-- analisis:inicio -->")
    w("## Análisis")
    w("")
    w("_(Pendiente de redactar a mano tras la primera corrida.)_")
    w("<!-- analisis:fin -->")
    w("")
    w("## Método")
    w("")
    w("- **Mecanismo**: el de Metrín, sin cambios. `rag-go/internal/clasificar.Entrenar` y `Modelo.Clasificar`/`Puntajes` "
      "(kNN k=1: para cada intención, el mejor coseno entre la frase y sus ejemplos; texto en minúsculas, sin tildes ni "
      "puntuación externa) sobre el embebedor estático `metrin/modelos/potion-es-int8.pjge` (model2vec, 256 dim, int8). "
      "Se llama desde `herramientas/clasificador_metrin/` (Go, módulo aparte enlazado por `go.work`; no toca `metrin/`). "
      "`rag entrenar-clasificador` no sirve: su tabla `intencionDe` solo acepta 8 categorías y las pliega en 4 clases.")
    w("- **Modo de puntuación**: `Puntajes` tiene dos modos sin tocar Go: kNN k=1 si el modelo trae `ejemplos`, centroides "
      "(coseno con la media normalizada de cada intención) si no los trae. Se elige el de mayor macro-F1 en leave-one-out "
      "del ENTRENAMIENTO (sin mirar la prueba) y el candidato se guarda en ese modo (sin `ejemplos` si gana centroides).")
    w(f"- **Paridad**: los puntajes del modo elegido reimplementados en Python con los vectores de Go reproducen los de Go "
      f"con una diferencia máxima de {R['paridad_knn_python_vs_go']:.1e}; por eso la comparación con MiniLM usa el mismo algoritmo.")
    w("- **Datos**: entrenamiento = `intenciones.yml` + `ruido.yml` (clase `ninguno`) y, aparte, `tipo_respuesta.yml`. "
      "Prueba = `intenciones_prueba.yml` y `tipo_respuesta_prueba.yml`, que nunca entran al entrenamiento (el validador "
      "rechaza frases iguales tras normalizar y casi iguales ≥ 0,90). Todo es sintético: ver «Límites».")
    w(f"- **Umbral de abstención, elegido con los datos**: se recorre cada puntaje observado de la prueba como umbral y se "
      f"toma el que **maximiza aciertos − {R['coste_error']:g} × errores** entre lo que se etiqueta; abstenerse vale 0 "
      f"(cuesta una pregunta de aclaración, ACLARAR_TAREA) y un error vale el doble (respuesta en el modo equivocado que el "
      f"usuario tiene que corregir). Equivale a etiquetar mientras la precisión marginal supere {R['coste_error'] / (1 + R['coste_error']):.0%}. "
      f"El criterio inicial, «máxima cobertura con precisión ≥ {R['objetivo_precision']:.0%}», se descartó porque con este "
      "embebedor degenera (solo lo cumplen umbrales que etiquetan un puñado de frases); su punto se sigue informando. "
      "Bajo el umbral, intenciones se abstiene y tipo de respuesta cae en UNKNOWN. El umbral se elige sobre la misma "
      "prueba (cifra «en muestra»); la estimación honesta es la **calibración cruzada**: 10 particiones estratificadas en "
      "dos mitades, se elige el umbral con una y se mide en la otra.")
    w("- **Cobertura** = fracción de frases etiquetadas (no abstenidas; en tipo de respuesta, no UNKNOWN). **Precisión** = "
      "aciertos entre las etiquetadas.")
    w("- **ACTUAL vs CANDIDATO**: el actual solo tiene `social`, `ayuda`, `limite` y `trabajo`. Para comparar sobre la "
      "misma prueba se lleva cada etiqueta a esas 4: saludo/despedida/agradecimiento → social; avance_paso, "
      "pedir_aclaracion, ambiguo, feedback_respuesta → ayuda (las categorías «seguimiento», «explicaciones», «ambiguas» y "
      "«correctivo» del entrenamiento actual); ninguno → limite; el resto (concepto, cómo hacer, error, reporte, "
      "funcionalidad, comparación, comercial, humano, aprendizaje) → trabajo. Las salidas se comparan tal como las devuelve "
      "`Clasificar` (umbral y capa léxica social incluidos): bajo el umbral, ambos dicen `trabajo`.")
    w("- **Reglas actuales de tipo de consulta**: port literal a Python de `tipoConsulta`/`esHowTo` "
      "(`internal/rag/orquestador.go` y `rag.go`), sin hilo (no hay «seguimiento»). problema → TROUBLESHOOTING, "
      "comparacion → COMPARISON, procedimiento → PROCEDURE, concepto → CONCEPT, aclaracion e informacion_directa → "
      "UNKNOWN. Las reglas no tienen NAVIGATION ni CONFIGURATION (\"dónde\" y \"encontrar\" las mandan a procedimiento).")
    if ok_mini:
        w(f"- **Embebedor alternativo**: `{MiniLM.nombre}` (BERT de 12 capas, 384 dim, multilingüe; uno de los candidatos "
          "de §3 de la especificación), leído de la caché local de Hugging Face e inferido en numpy (sin torch ni red), "
          "media de tokens normalizada; texto crudo. Mismo kNN y mismo criterio de umbral. "
          f"Tardó {mini['segundos']:.0f} s para {sum(map(len, ent_int.values())) + len(prueba_int) + sum(map(len, ent_tipo.values())) + len(prueba_tipo)} frases.")
    elif mini:
        w(f"- Embebedor alternativo: no se pudo medir ({mini['error']}).")
    w("")
    w("Reproducir: `.venv/bin/python3 herramientas/validar_catalogos_metrin.py && .venv/bin/python3 "
      "herramientas/evaluar_clasificador_metrin.py`. Modelos: `metrin/modelos/clasificador-candidato.json` y "
      "`metrin/modelos/clasificador-tipo-respuesta-candidato.json` (formato de `clasificar.Modelo`, se cargan con "
      "`rag clasificar --modelo <ruta>`), con el umbral elegido en `umbral`. `defecto.json` NO se ha tocado.")
    w("")

    w("### Modos de puntuación (✓ = elegido por leave-one-out del entrenamiento)")
    w("")
    L.extend(tabla_metodos(metodos))
    w("")
    w("El modo híbrido (½ kNN top-3 + ½ centroide, el que ganó en kddesign) no se puede usar sin cambiar `clasificar.go`; "
      "no se mide aquí para no elegir nada que el servicio no ejecute.")
    w("")
    # ---------- intenciones ----------
    w("## 1. Intenciones")
    w("")
    w(f"Entrenamiento: {log_ent}")
    w("")
    w("### 1.1 Por intención (candidato, estático)")
    w("")
    ab = {c: sum(1 for t, p in zip(y_int, ev_int["pred_umbral"]) if t == c and p == ABST) for c in nombres_int}
    extra = {"cab": ["Recall con umbral", "Abstenciones"],
             "filas": {c: [pct(ev_int["prf_umbral"][c]["recall"]), str(ab[c])] for c in nombres_int}}
    L.extend(tabla_prf(ev_int["prf"], extra))
    w("")
    w("Precisión, recall y F1 sin umbral (top-1); las dos últimas columnas, con el umbral elegido. Ordenado de peor a mejor F1.")
    w("")
    w("### 1.2 Confusiones (matriz resumida)")
    w("")
    w("Pares real → predicha más frecuentes (top-1, sin umbral): " + "; ".join(confusiones_resumen(y_int, ev_int["pred"], 15)) + ".")
    w("")
    fp_ning = [(f, p) for (f, t), p, s in zip(prueba_int, ev_int["pred"], ev_int["s1"]) if t == NINGUNO and p != NINGUNO and s >= ui["umbral"]]
    fn_ning = [(f, t) for (f, t), p, s in zip(prueba_int, ev_int["pred"], ev_int["s1"]) if t != NINGUNO and p == NINGUNO and s >= ui["umbral"]]
    w(f"Clase `ninguno` (fuera de dominio y basura): {sum(1 for t in y_int if t == NINGUNO)} frases de prueba; con el umbral, "
      f"{len(fp_ning)} se etiquetan como una intención real (falsos positivos: {', '.join(f'«{f}» → {p}' for f, p in fp_ning) or 'ninguno'}). "
      f"Frases del dominio mandadas a `ninguno`: {len(fn_ning)} ({', '.join(f'«{f}»' for f, _ in fn_ning) or '—'}).")
    w("")
    w("Fallos de la prueba (top-1, sin umbral):")
    w("")
    for (f, t), p, s in zip(prueba_int, ev_int["pred"], ev_int["s1"]):
        if t != p:
            w(f"- «{f}»: esperado {t}, salió {p} ({s:.3f}{', bajo el umbral → abstención' if s < ui['umbral'] else ''})")
    w("")
    w("### 1.3 Cobertura y precisión según el umbral")
    w("")
    w("| Umbral | Cobertura | Precisión |")
    w("|---:|---:|---:|")
    for tau, cob, prec in sorted(set(ev_int["curva"])):
        marca = " ← elegido" if abs(tau - ui["umbral"]) < 1e-9 else ""
        w(f"| {tau:.3f}{marca} | {pct(cob)} | {pct(prec)} |")
    w("")
    w(f"Criterio: {ui['criterio']}. {punto_txt(ui)} Calibración cruzada (n={ev_int['cruzada']['n']}): umbral {ev_int['cruzada']['umbral_medio']:.3f} "
      f"± {ev_int['cruzada']['umbral_sd']:.3f}, cobertura {pct(ev_int['cruzada']['cobertura'])}, precisión "
      f"{pct(ev_int['cruzada']['precision'])} ± {pct(ev_int['cruzada']['precision_sd'])}.")
    w("")
    w("### 1.4 ACTUAL (`defecto.json`) frente a CANDIDATO, en las 4 clases del actual")
    w("")
    w("| Clase del actual | n | Recall ACTUAL | Precisión ACTUAL | Recall CANDIDATO | Precisión CANDIDATO |")
    w("|---|---:|---:|---:|---:|---:|")
    for c in ["social", "ayuda", "limite", "trabajo"]:
        a_, b_ = comp["prf_actual"][c], comp["prf_candidato"][c]
        w(f"| {c} | {a_['n']} | {pct(a_['recall'])} | {pct(a_['precision'])} | {pct(b_['recall'])} | {pct(b_['precision'])} |")
    w(f"| **global (exactitud)** | {len(y_int)} | {pct(comp['gruesa_actual'])} | | {pct(comp['gruesa_candidato'])} | |")
    w("")
    w(f"- De los aciertos del ACTUAL en `trabajo`, {comp['trabajo_por_abstencion_actual']} llegan por abstención (similitud < 0,40 "
      "→ `trabajo` por defecto), no porque el modelo reconozca la frase.")
    w(f"- Ruta efectiva de hoy (social y limite desvían; ayuda y trabajo van al RAG): ACTUAL {pct(comp['ruta_actual'])}, "
      f"CANDIDATO {pct(comp['ruta_candidato'])}.")
    w(f"- Exactitud FINA del candidato tal como responde `Clasificar` (con su umbral; abstención = fallo): "
      f"{pct(comp['fina_candidato_clasificar'])}. El actual no tiene etiquetas finas: no se puede medir ahí.")
    w("")
    w("Confusiones del ACTUAL en 4 clases: " + "; ".join(confusiones_resumen(esperada_g, act_g, 8)) + ".")
    w("")
    w("Confusiones del CANDIDATO en 4 clases: " + "; ".join(confusiones_resumen(esperada_g, cand_g, 8)) + ".")
    w("")
    w("### 1.5 Casos medidos")
    w("")
    for t, c in casos.items():
        w(f"- «{t}» (esperado **{c['esperado']}**): ACTUAL {c['actual']['intencion']} ({c['actual']['similitud']:.3f}; top-3 "
          f"{c['actual']['top3']}); CANDIDATO {c['candidato']['intencion']} ({c['candidato']['similitud']:.3f}; top-3 "
          f"{c['candidato']['top3']})" + (f"; MiniLM {c['minilm']['intencion']} ({c['minilm']['similitud']:.3f}, con su umbral: "
                                         f"{c['minilm']['con_umbral']})" if ok_mini else "") + ".")
    w("")

    # ---------- tipo de respuesta ----------
    w("## 2. Tipo de respuesta (modo de la V2)")
    w("")
    w("### 2.1 Precisión, recall y F1 por clase")
    w("")
    w(f"Con el umbral elegido ({ut['umbral']:.3f}; bajo él, UNKNOWN). Entre paréntesis, sin umbral.")
    w("")
    w("| Clase | n | Precisión | Recall | F1 | F1 sin umbral | F1 reglas actuales |" + (" F1 MiniLM (con su umbral) |" if ok_mini else ""))
    w("|---|---:|---:|---:|---:|---:|---:|" + ("---:|" if ok_mini else ""))
    for c in nombres_tipo:
        v = ev_tipo["prf_umbral"][c]
        fila = (f"| {c} | {v['n']} | {pct(v['precision'])} | {pct(v['recall'])} | {pct(v['f1'])} | "
                f"{pct(ev_tipo['prf'][c]['f1'])} | {pct(prf_reglas[c]['f1'])} |")
        if ok_mini:
            fila += f" {pct(mini['tipo']['prf_umbral'][c]['f1'])} |"
        w(fila)
    fila = (f"| **macro** | {len(y_tipo)} | | | **{pct(ev_tipo['macro_f1_umbral'])}** | {pct(ev_tipo['macro_f1'])} | "
            f"{pct(macro_f1(prf_reglas))} |")
    if ok_mini:
        fila += f" **{pct(mini['tipo']['macro_f1_umbral'])}** |"
    w(fila)
    w(f"| exactitud | | | | {pct(ev_tipo['exactitud_umbral'])} | {pct(ev_tipo['exactitud'])} | "
      f"{pct(float(np.mean([a_ == b_ for a_, b_ in zip(y_tipo, reglas)])))} |" + (f" {pct(mini['tipo']['exactitud_umbral'])} |" if ok_mini else ""))
    w("")
    w("### 2.2 Matriz de confusión (estático, con umbral)")
    w("")
    L.extend(matriz_confusion(y_tipo, ev_tipo["pred_umbral"], nombres_tipo, nombres_tipo))
    w("")
    w("Reglas actuales (`tipoConsulta`):")
    w("")
    L.extend(matriz_confusion(y_tipo, reglas, nombres_tipo, nombres_tipo))
    w("")
    if ok_mini:
        w("MiniLM (mismo kNN, con su umbral):")
        w("")
        L.extend(matriz_confusion(y_tipo, mini["tipo"]["pred_umbral"], nombres_tipo, nombres_tipo))
        w("")
    w("### 2.3 Cobertura y precisión según el umbral (estático)")
    w("")
    w("| Umbral | Cobertura (no UNKNOWN) | Precisión |")
    w("|---:|---:|---:|")
    for tau, cob, prec in sorted(set(ev_tipo["curva"])):
        marca = " ← elegido" if abs(tau - ut["umbral"]) < 1e-9 else ""
        w(f"| {tau:.3f}{marca} | {pct(cob)} | {pct(prec)} |")
    w("")
    w(f"Criterio: {ut['criterio']}. {punto_txt(ut)} Calibración cruzada: umbral {ev_tipo['cruzada']['umbral_medio']:.3f} ± "
      f"{ev_tipo['cruzada']['umbral_sd']:.3f}, cobertura {pct(ev_tipo['cruzada']['cobertura'])}, precisión "
      f"{pct(ev_tipo['cruzada']['precision'])} ± {pct(ev_tipo['cruzada']['precision_sd'])}.")
    if ok_mini:
        mu = mini["tipo"]
        w(f"MiniLM: umbral {mu['umbral']['umbral']:.3f} ({mu['umbral']['criterio']}), cobertura/precisión en muestra "
          f"{pct(mu['umbral']['cobertura'])} / {pct(mu['umbral']['precision'])}; cruzada {pct(mu['cruzada']['cobertura'])} / "
          f"{pct(mu['cruzada']['precision'])}.")
    w("")
    w("### 2.4 Casos literales pedidos")
    w("")
    w("| Frase | Esperado | Estático (top-1, sim → con umbral) | Reglas actuales |" + (" MiniLM |" if ok_mini else ""))
    w("|---|---|---|---|" + ("---|" if ok_mini else ""))
    for t, c in literales.items():
        e = c["estatico"]
        fila = f"| «{t}» | {c['esperado']} | {e[0]} ({e[1]:.3f}) → {e[2]} | {c['reglas']} |"
        if ok_mini:
            m_ = c["minilm"]
            fila += f" {m_[0]} ({m_[1]:.3f}) → {m_[2]} |"
        w(fila)
    w("")
    w("Fallos del estático (con umbral): " + "; ".join(confusiones_resumen(y_tipo, ev_tipo["pred_umbral"], 12)) + ".")
    w("")
    if ok_mini:
        w("## 3. Estático frente a MiniLM (mismo kNN, mismo criterio de umbral)")
        w("")
        w("| | Intenciones estático | Intenciones MiniLM | Tipo estático | Tipo MiniLM |")
        w("|---|---:|---:|---:|---:|")
        mi, mt = mini["int"], mini["tipo"]
        w(f"| Exactitud top-1 | {pct(ev_int['exactitud'])} | {pct(mi['exactitud'])} | {pct(ev_tipo['exactitud'])} | {pct(mt['exactitud'])} |")
        w(f"| Top-3 | {pct(ev_int['top3'])} | {pct(mi['top3'])} | {pct(ev_tipo['top3'])} | {pct(mt['top3'])} |")
        w(f"| Macro-F1 sin umbral | {pct(ev_int['macro_f1'])} | {pct(mi['macro_f1'])} | {pct(ev_tipo['macro_f1'])} | {pct(mt['macro_f1'])} |")
        w(f"| Modo (LOO) | {R['modo_int']} | {mini['modo_int']} | {R['modo_tipo']} | {mini['modo_tipo']} |")
        w(f"| Umbral elegido | {ui['umbral']:.3f} | {mi['umbral']['umbral']:.3f} | {ut['umbral']:.3f} | {mt['umbral']['umbral']:.3f} |")
        w(f"| Cruzada cobertura / precisión | {pct(ev_int['cruzada']['cobertura'])} / {pct(ev_int['cruzada']['precision'])} | "
          f"{pct(mi['cruzada']['cobertura'])} / {pct(mi['cruzada']['precision'])} | {pct(ev_tipo['cruzada']['cobertura'])} / "
          f"{pct(ev_tipo['cruzada']['precision'])} | {pct(mt['cruzada']['cobertura'])} / {pct(mt['cruzada']['precision'])} |")
        w(f"| Macro-F1 con umbral | {pct(ev_int['macro_f1_umbral'])} | {pct(mi['macro_f1_umbral'])} | {pct(ev_tipo['macro_f1_umbral'])} | {pct(mt['macro_f1_umbral'])} |")
        w("")
        w("Peores intenciones con MiniLM (F1 sin umbral): " + ", ".join(
            f"{c} {pct(v['f1'])}" for c, v in sorted(mi["prf"].items(), key=lambda kv: kv[1]["f1"])[:5] if v["n"]) + ".")
        w("")
    w("## Límites")
    w("")
    w("- Entrenamiento y prueba son **sintéticos** y los escribió el mismo equipo con el mismo estilo: las cifras son "
      "optimistas respecto de mensajes reales. Cuando haya consultas reales anonimizadas, deben ser la prueba y se recalibra.")
    w("- El umbral elegido «en muestra» usa la misma prueba que mide; la cifra para decidir es la calibración cruzada.")
    w("- La prueba de intenciones tiene 8–12 frases por clase: un acierto mueve el recall de una clase 8–12 puntos.")
    w("- El candidato no está integrado: el orquestador actual (`internal/rag/orquestador.go`) solo entiende `social`, "
      "`limite`, `ayuda` y `trabajo`, y `Clasificar` devuelve `trabajo` bajo el umbral. Usarlo exige mapear las nuevas "
      "etiquetas a rutas (campo `ruta` de `intenciones.yml`) y tratar `trabajo` como abstención; no se ha cambiado Go.")
    w("")
    return "\n".join(L) + "\n"


if __name__ == "__main__":
    sys.exit(main())
