#!/usr/bin/env python3
"""
reportes.py — reportes SQL dinámicos SIN SQL libre: la IA elige y rellena una
plantilla del catálogo aprobado (reportes/plantillas.json).

Fases (el plan del agente de reportes, sin LangGraph):

  1 recuperar   pregunta -> plantilla del catálogo (solapamiento léxico)
  2 resolver    parámetros (fechas, periodo, categoría) con reglas explícitas
  3 validar     sqlglot: parsea y exige SELECT único
  4 ejecutar    SQLite o PostgreSQL (read-only + LIMIT de cortesía)
  5 responder   tabla Markdown + SQL ejecutado

Uso:
  .venv/bin/python reportes.py sembrar
  .venv/bin/python reportes.py catalogo
  .venv/bin/python reportes.py preguntar "¿Cuánto costó la planilla de junio 2024?"
  .venv/bin/python reportes.py preguntar "horas tareadas de marzo 2024" --sql
  .venv/bin/python reportes.py preguntar "..." --bd postgresql://user@host/db
"""
from __future__ import annotations

import argparse
import json
import re
import sqlite3
import sys
import unicodedata
from pathlib import Path

import sqlglot
from sqlglot import exp

sys.path.insert(0, str(Path(__file__).resolve().parent))
from s10kb import DATA, RAIZ, agregar_jsonl, log, slug

PLANTILLAS = RAIZ / "reportes" / "plantillas.json"
_cache: dict | None = None

MESES = {"enero": 1, "febrero": 2, "marzo": 3, "abril": 4, "mayo": 5, "junio": 6,
         "julio": 7, "agosto": 8, "setiembre": 9, "septiembre": 9, "octubre": 10,
         "noviembre": 11, "diciembre": 12}


def cargar_catalogo() -> list[dict]:
    global _cache
    if _cache is None:
        _cache = json.loads(PLANTILLAS.read_text(encoding="utf-8"))["plantillas"]
    return _cache


# ─────────────────────────── 1. recuperar plantilla ────────────────────────
def _tokens(t: str) -> set[str]:
    t = re.sub(r"[^a-z0-9áéíóúñü ]", " ", t.lower())
    parar = {"de", "la", "el", "los", "las", "del", "y", "o", "a", "en", "que", "cuanto",
             "cuanta", "cuantos", "cuantas", "cual", "cuales", "muestrame", "muestra",
             "dame", "por", "para", "con", "un", "una", "es", "son", "al", "sobre",
             "reporte", "reportes", "quiero", "necesito", "hay", "se", "su", "sus", "the"}
    return {p for p in t.split() if p and p not in parar}


def recuperar(pregunta: str) -> tuple[dict, float]:
    q = _tokens(pregunta)
    mejor, puntaje = None, 0.0
    for p in cargar_catalogo():
        corpus = _tokens(
            p["descripcion"] + " " + p["nombre"].replace("-", " ") + " " + p["titulo"]
            + " " + " ".join(p.get("ejemplos", [])) + " " + p["sql"].lower())
        corpus.discard("select")
        inter = q & corpus
        s = len(inter) / max(1, len(q)) if q else 0.0
        if s > puntaje:
            mejor, puntaje = p, s
    if mejor is None or puntaje == 0:
        raise SystemExit("Ninguna plantilla corresponde a la pregunta. Catálogo:\n"
                         + "\n".join(f"  - {p['id']} {p['titulo']}" for p in cargar_catalogo()))
    return mejor, round(puntaje, 2)


# ─────────────────────────── 2. resolver parámetros ────────────────────────
def resolver(pregunta: str) -> dict:
    t = unicodedata.normalize("NFKD", pregunta.lower()).encode("ascii", "ignore").decode()
    params: dict = {}

    # rango "entre 2024-01 y 2024-06"
    m = re.search(r"entre (\d{4})-(\d{2}) y (\d{4})-(\d{2})", t)
    if m:
        params["desde"] = f"{m.group(1)}-{m.group(2)}-01"
        params["hasta"] = f"{m.group(3)}-{m.group(4)}-28"

    # periodo "2024-06"
    if "periodo" not in params:
        m = re.search(r"(\d{4})-(\d{2})", t)
        if m:
            params["periodo"] = f"{m.group(1)}-{m.group(2)}"

    # "junio 2024" / "marzo de 2024"
    m = re.search(r"(enero|febrero|marzo|abril|mayo|junio|julio|agosto|setiembre|"
                  r"septiembre|octubre|noviembre|diciembre)\s+(?:de\s+)?(\d{4})", t)
    if m:
        anio, mm = int(m.group(2)), MESES[m.group(1)]
        params["periodo"] = f"{anio}-{mm:02d}"
        params.setdefault("desde", f"{anio}-{mm:02d}-01")
        params.setdefault("hasta", f"{anio}-{mm:02d}-31")

    # año solo -> rango anual
    m = re.search(r"\b(20\d{2})\b", t)
    if m and "desde" not in params:
        y = m.group(1)
        params.setdefault("desde", f"{y}-01-01")
        params.setdefault("hasta", f"{y}-12-31")

    # categoría
    for patron, valor in ((r"\boperari[oa]s?\b", "operario"), (r"\boficial(es)?\b", "oficial"),
                          (r"\bpeon(es)?\b", "peón")):
        if re.search(patron, t):
            params["categoria"] = valor
            break
    return params


# ─────────────────────────── 3. validar con sqlglot ────────────────────────
def validar(sql: str, dialecto: str = "sqlite") -> str:
    arbol = sqlglot.parse_one(sql, read=dialecto)
    if not isinstance(arbol, exp.Select):
        raise ValueError("solo se permiten consultas SELECT")
    for nodo in arbol.find_all(exp.Insert, exp.Update, exp.Delete, exp.Drop, exp.Alter,
                               exp.Create, exp.Attach, exp.Detach, exp.Command):
        raise ValueError(f"sentencia no permitida: {type(nodo).__name__}")
    return arbol.sql(dialect=dialecto, pretty=True)


# ─────────────────────────── 4. ejecutar ───────────────────────────────────
def ejecutar(sql: str, params: dict, bd: str) -> tuple[list[dict], list[str]]:
    if bd.startswith(("postgresql://", "postgres://")):
        try:
            import psycopg2
        except ImportError as e:
            raise SystemExit("para PostgreSQL: .venv/bin/pip install psycopg2-binary") from e
        conn = psycopg2.connect(bd, options="-c statement_timeout=15000")
        conn.set_session(readonly=True, autocommit=True)
        cur = conn.cursor()
        comp = sql
        for k, v in params.items():
            comp = comp.replace(f":{k}", "NULL" if v is None else f"%({k})s")
        cur.execute(comp, {k: v for k, v in params.items() if v is not None})
        columnas = [c.name for c in cur.description] if cur.description else []
        filas = [dict(zip(columnas, r)) for r in cur.fetchmany(500)]
        conn.close()
    else:
        conn = sqlite3.connect(bd, timeout=5)
        conn.row_factory = sqlite3.Row
        cur = conn.execute(sql, params)
        filas = [dict(r) for r in cur.fetchmany(500)]
        conn.close()
    return filas, list(filas[0].keys()) if filas else []


# ─────────────────────────── 5. responder ──────────────────────────────────
def responder(plantilla: dict, params: dict, sql: str, filas: list[dict],
              columnas: list[str], puntaje: float) -> str:
    lin = "─" * 66
    out = [lin, f" REPORTE {plantilla['id']} — {plantilla['titulo']}",
           f" plantilla del catálogo (puntaje léxico {puntaje})", lin]
    if params:
        out.append("parámetros: " + json.dumps(params, ensure_ascii=False))
    out.append("")
    out.append(sql.rstrip(";") + ";")
    out.append("")
    if not filas:
        out.append("(sin filas)")
    else:
        anchos = [max(len(str(c)), *(len(str(f[c])) for f in filas)) for c in columnas]

        def fila(vals: list) -> str:
            return "| " + " | ".join(str(v).rjust(anchos[i]) for i, v in enumerate(vals)) + " |"

        out.append(fila(columnas))
        out.append("|" + "|".join("-" * (a + 2) for a in anchos) + "|")
        for f in filas[:50]:
            out.append(fila([f[c] for c in columnas]))
        if len(filas) > 50:
            out.append(f"… {len(filas) - 50} filas más (mostradas 50)")
    out.append(lin)
    return "\n".join(out)


# ─────────────────────────── BD demo S10-like ──────────────────────────────
def sembrar() -> Path:
    """SQLite demo con esquema S10-like y datos citados (video 17W0yKI8iew)."""
    bd = DATA / "demo_s10.db"
    bd.parent.mkdir(exist_ok=True)
    if bd.exists():
        bd.unlink()
    conn = sqlite3.connect(bd)
    conn.executescript("""
CREATE TABLE constantes_fecha (
  id INTEGER PRIMARY KEY, concepto TEXT, abreviatura TEXT,
  desde TEXT, hasta TEXT, valor REAL, observaciones TEXT);
CREATE TABLE trabajadores (
  id INTEGER PRIMARY KEY, nombre TEXT, categoria TEXT);
CREATE TABLE tareas (
  id INTEGER PRIMARY KEY, trabajador_id INT, fecha TEXT, horas REAL,
  FOREIGN KEY (trabajador_id) REFERENCES trabajadores(id));
CREATE TABLE planilla (
  id INTEGER PRIMARY KEY, trabajador_id INT, periodo TEXT, concepto TEXT, monto REAL,
  FOREIGN KEY (trabajador_id) REFERENCES trabajadores(id));

-- Constantes reales leídas en el video del piloto (pantalla Constantes por
-- Fecha General): UIT 2020=4300, UIT 2021=4400, UIT 2024=5150, RMV 2020=930.
-- Fechas en ISO (la BD compara; el formato visual DD/MM/YYYY es del presentador).
INSERT INTO constantes_fecha VALUES
 (1,'REMUNERACION MINIMA VITAL','RMV','2020-01-01','2020-12-31',930,'RMV 2020'),
 (2,'UIT','UIT','2020-01-01','2020-12-31',4300,'UIT 2020'),
 (3,'UIT','UIT','2021-01-01','2021-12-31',4400,'UIT 2021'),
 (4,'UIT','UIT','2024-01-01','2024-12-31',5150,'UIT 2024');

INSERT INTO trabajadores VALUES
 (1,'Carlos Ramos','operario'), (2,'Ana Torres','oficial'), (3,'Luis Quispe','peón'),
 (4,'María Huamán','operario'), (5,'Jorge Silva','oficial');

-- Tareo de enero 2024 (fechas ISO)
INSERT INTO tareas VALUES
 (1,1,'2024-01-15',8),(2,2,'2024-01-15',8),(3,3,'2024-01-15',6),
 (4,1,'2024-01-16',9),(5,2,'2024-01-16',8),(6,4,'2024-01-16',8),
 (7,3,'2024-01-17',7),(8,4,'2024-01-17',8),(9,5,'2024-01-17',8);

-- Planilla: básicos de junio 2024, gratificación de julio 2024, CTS nov 2024
INSERT INTO planilla VALUES
 (1,1,'2024-06','BASICO',3200),(2,2,'2024-06','BASICO',2500),
 (3,3,'2024-06','BASICO',1800),(4,4,'2024-06','BASICO',3100),
 (5,5,'2024-06','BASICO',2600),
 (6,1,'2024-07','GRATIFICACION FIESTAS PATRIAS',3200),
 (7,2,'2024-07','GRATIFICACION FIESTAS PATRIAS',2500),
 (8,3,'2024-07','GRATIFICACION FIESTAS PATRIAS',1800),
 (9,4,'2024-07','GRATIFICACION FIESTAS PATRIAS',3100),
 (10,5,'2024-07','GRATIFICACION FIESTAS PATRIAS',2600),
 (11,1,'2024-11','CTS',3200),(12,2,'2024-11','CTS',2500),
 (13,3,'2024-11','CTS',1800),(14,4,'2024-11','CTS',3100),(15,5,'2024-11','CTS',2600);
""")
    conn.commit()
    conn.close()
    log(f"BD demo -> {bd.relative_to(RAIZ)} (UIT 4300/4400/5150 y RMV 930 citadas del video)")
    return bd


# ─────────────────── 6. publicar: KB y Cortex ──────────────────────────────
def fragmentos_reporte(p: dict) -> list[str]:
    """Texto natural del reporte, para la KB del RAG."""
    params = "; ".join(
        f"{k} ({s['tipo']}, p. ej. {s.get('defecto') or 'vacío = sin filtro'})"
        for k, s in p["parametros"].items())
    return [
        f"Reporte SQL aprobado del ERP S10: {p['id']} «{p['titulo']}».",
        f"Responde preguntas como: {'; '.join(p['ejemplos'])}.",
        f"Parámetros: {params}.",
        f"SQL pre-aprobado (la IA solo lo rellena, nunca lo inventa): {p['sql']}",
        "Se ejecuta con el catálogo de reportes.py: validación sqlglot (solo SELECT), "
        "sesión de base de datos de solo lectura y máximo 500 filas.",
    ]


def paso_kb() -> None:
    """Escribe kb/fragmentos_reportes.jsonl (lo levantan Metrín y tutor.py por el glob)."""
    kb_dir = RAIZ / "kb"
    kb_dir.mkdir(exist_ok=True)
    salida = kb_dir / "fragmentos_reportes.jsonl"
    salida.write_text("", encoding="utf-8")
    n = 0
    for p in cargar_catalogo():
        texto = "\n".join(fragmentos_reporte(p))
        agregar_jsonl(salida, {
            "id": f"reporte-{p['nombre']}", "documento": f"reporte-{p['nombre']}",
            "tipo": "reporte", "titulo": f"{p['id']} — {p['titulo']}",
            "fuente": "reportes/plantillas.json (catálogo aprobado a mano)",
            "campos": list(p["parametros"]), "texto": texto,
        })
        n += 1
    log(f"{n} reportes -> {salida.relative_to(RAIZ)}")


def paso_cortex() -> None:
    """Genera data/cortex-borradores/reportes-sql.json: nodo-método + un hijo por reporte."""
    nodos = [{
        "path": "reportes-sql",
        "title": "Reportes SQL del ERP (catálogo aprobado)",
        "summary": "Preguntas en lenguaje natural a reportes SIN SQL generado por IA: elige y rellena plantillas pre-aprobadas, valida con sqlglot (solo SELECT) y ejecuta en sesión read-only.",
        "body": (
            "Método: nada de SQL libre. La IA recupera la plantilla del catálogo "
            "aprobado (reportes/plantillas.json), resuelve parámetros con reglas "
            "explícitas (fechas, periodo, categoría) y sqlglot valida por AST que la "
            "consulta sea un SELECT único (rechaza INSERT/UPDATE/DROP/ATTACH incluso "
            "anidados). La ejecución va en sesión read-only de la BD con tope de 500 "
            "filas: la garantía real es el rol de la base, no la cortesía del código.\n\n"
            "Catálogo: 6 plantillas (constantes UIT/RMV, tareo, jornal, planilla, "
            "gratificaciones, CTS). Probado contra BD demo sembrada con los datos "
            "citados del video del piloto (UIT 4300/4400/5150, RMV 930).\n\n"
            "Pendiente para la BD real de S10 (debe dar Arturo): cadena de conexión y "
            "usuario de SOLO LECTURA. Con eso basta "
            "`reportes.py preguntar \"…\" --bd postgresql://…`."
        ),
        "tags": ["reportes", "sql", "metodo", "seguridad"],
    }]
    for p in cargar_catalogo():
        nodos.append({
            "path": f"reportes-sql/{p['nombre']}",
            "title": f"{p['id']} — {p['titulo']}",
            "summary": f"Reporte aprobado. Ejemplos: {'; '.join(p['ejemplos'][:2])}.",
            "body": "\n".join(fragmentos_reporte(p))
                   + "\n\n**Uso:** `reportes.py preguntar \"…\"` — la plantilla solo se "
                     "rellena, no se edita. Cambios de SQL = nueva versión del catálogo "
                     "revisada a mano.",
            "tags": ["reporte", "sql"],
        })
    destino = DATA / "cortex-borradores" / "reportes-sql.json"
    destino.parent.mkdir(exist_ok=True)
    destino.write_text(json.dumps(nodos, ensure_ascii=False, indent=1), encoding="utf-8")
    log(f"{len(nodos)} nodos borrador -> {destino.relative_to(RAIZ)} (cargar con herramientas/cargar_cortex.py)")


# ─────────────────────────── CLI ───────────────────────────────────────────
def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="paso", required=True)
    sub.add_parser("sembrar", help="crea la BD demo SQLite con datos del video")
    sub.add_parser("catalogo", help="lista las plantillas aprobadas")
    sub.add_parser("kb", help="publica el catálogo como fragmentos de la KB (Metrín/tutor)")
    sub.add_parser("cortex", help="genera el borrador de Cortex del catálogo")
    p = sub.add_parser("preguntar", help="pregunta en lenguaje natural")
    p.add_argument("pregunta")
    p.add_argument("--bd", default=str(DATA / "demo_s10.db"),
                   help="ruta SQLite o URL postgresql:// (read-only)")
    p.add_argument("--sql", action="store_true", help="muestra solo el SQL resuelto")
    a = ap.parse_args()

    if a.paso == "sembrar":
        sembrar()
        return
    if a.paso == "kb":
        paso_kb()
        return
    if a.paso == "cortex":
        paso_cortex()
        return
    if a.paso == "catalogo":
        for p_ in cargar_catalogo():
            print(f"{p_['id']}  {p_['titulo']}\n      ej: {' · '.join(p_['ejemplos'][:3])}")
        return

    plantilla, puntaje = recuperar(a.pregunta)
    params = resolver(a.pregunta)
    # completar con defectos declarados en la plantilla
    for nombre, spec in plantilla["parametros"].items():
        params.setdefault(nombre, spec.get("defecto"))
    params = {k: v for k, v in params.items() if k in plantilla["parametros"]}
    sql = validar(plantilla["sql"])
    if a.sql:
        print(sql)
        print("parámetros:", json.dumps(params, ensure_ascii=False))
        return
    filas, columnas = ejecutar(sql, params, a.bd)
    print(responder(plantilla, params, sql, filas, columnas, puntaje))


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit("\ninterrumpido")
