#!/usr/bin/env python3
"""
admin.py — panel de administración de la base de conocimiento S10.

  .venv/bin/python admin.py            # http://127.0.0.1:4750

Secciones: resumen · Cortex (entrar con sesión) · fuentes (PDF, portal, YouTube, internas)
· agregar fuentes · tutoriales · herramientas · almacenamiento.

Acceso con usuario y clave de .env (ADMIN_USUARIO / ADMIN_CLAVE). Si falta la clave,
se genera una al primer arranque y se guarda en .env (se imprime una sola vez).
Solo escucha en 127.0.0.1.
"""
from __future__ import annotations

import html
import json
import os
import re
import secrets
import shutil
import subprocess
import sys
import threading
import time
from functools import wraps
from pathlib import Path
from urllib.parse import quote

import markdown as md
import requests
from flask import Flask, Response, abort, redirect, request, send_file, session, url_for

RAIZ = Path(__file__).resolve().parent
DATA, KB, TUT = RAIZ / "data", RAIZ / "kb", RAIZ / "tutoriales"
PY = str(RAIZ / ".venv" / "bin" / "python")
CORTEX_API = os.environ.get("S10_CORTEX_API", "http://localhost:4748/api")
# Cortex solo acepta Host: localhost (protección anti DNS-rebinding); en Docker se fuerza.
CORTEX_H = {"Host": os.environ["S10_CORTEX_HOST"]} if os.environ.get("S10_CORTEX_HOST") else {}
NIVELES = {
    "oficial": "Oficial S10",
    "tercero-sin-verificar": "Copia de tercero",
    "academico": "Académico",
    "estado-2013": "Estado 2013",
    "propio": "Optimiza 360",
}


# ─────────────────────────────── configuración ─────────────────────────────
def leer_env() -> dict:
    f = RAIZ / ".env"
    d = {}
    if f.exists():
        for l in f.read_text(encoding="utf-8").splitlines():
            if "=" in l and not l.strip().startswith("#"):
                k, v = l.split("=", 1)
                d[k.strip()] = v.strip().strip('"')
    return d


def asegurar_credenciales() -> dict:
    env = leer_env()
    if not env.get("ADMIN_CLAVE"):
        clave = secrets.token_urlsafe(12)
        with (RAIZ / ".env").open("a", encoding="utf-8") as f:
            f.write(f"\n# Panel admin.py\nADMIN_USUARIO=admin\nADMIN_CLAVE={clave}\nADMIN_SECRETO={secrets.token_hex(32)}\n")
        print(f"\n  Panel: usuario 'admin', clave generada: {clave}\n  (guardada en .env; no se vuelve a mostrar)\n")
        env = leer_env()
    return env


ENV = asegurar_credenciales()
app = Flask(__name__)
app.secret_key = ENV.get("ADMIN_SECRETO") or secrets.token_hex(32)
app.config.update(SESSION_COOKIE_HTTPONLY=True, SESSION_COOKIE_SAMESITE="Lax", MAX_CONTENT_LENGTH=300 * 1024 * 1024)


def requiere_login(f):
    @wraps(f)
    def env(*a, **kw):
        if not session.get("ok"):
            return redirect(url_for("login", siguiente=request.path))
        return f(*a, **kw)
    return env


def csrf() -> str:
    if "csrf" not in session:
        session["csrf"] = secrets.token_hex(16)
    return session["csrf"]


def validar_csrf() -> None:
    if request.form.get("csrf") != session.get("csrf"):
        abort(400, "token de formulario inválido; recarga la página")


# ─────────────────────────────── datos ─────────────────────────────────────
def jsonl(p: Path) -> list[dict]:
    return [json.loads(l) for l in p.read_text(encoding="utf-8").splitlines() if l.strip()] if p.exists() else []


def confianza_manual() -> dict:
    f = DATA / "confianza_manual.json"
    return json.loads(f.read_text(encoding="utf-8")) if f.exists() else {}


def documentos() -> list[dict]:
    docs = {d["archivo"]: d for d in (json.loads((KB / "documentos.json").read_text()) if (KB / "documentos.json").exists() else [])}
    manual = confianza_manual()
    out = []
    for m in jsonl(DATA / "manifiesto.jsonl"):
        if m.get("estado") != "ok":
            continue
        d = docs.get(m["archivo"], {})
        stem = Path(m["archivo"]).stem
        out.append({
            "titulo": m.get("etiqueta") or stem, "archivo": m["archivo"], "url": m["url"],
            "origen": m.get("origen") or "portal", "pagina": m.get("pagina"),
            "confianza": manual.get(m["url"]) or d.get("confianza") or "oficial",
            "paginas": d.get("paginas"), "fragmentos": d.get("fragmentos"),
            "ocr": (DATA / "ocr" / f"{stem}.txt").exists(), "kb": m.get("bytes", 0) // 1024,
        })
    return out


def paginas_web() -> list[dict]:
    u = {}
    for p in jsonl(DATA / "paginas.jsonl"):
        u[p["pagina"]] = p
    return sorted(u.values(), key=lambda p: (p.get("muro_de_miembros", False), p.get("titulo") or ""))


def videos() -> list[dict]:
    f = DATA / "yt" / "videos.txt"
    out = []
    if not f.exists():
        return out
    for l in f.read_text(encoding="utf-8").splitlines():
        partes = l.split("|", 2)
        if len(partes) < 3:
            continue
        vid, seg, titulo = partes
        t = DATA / "yt" / "transcripciones" / f"{vid}.txt"
        tam = t.stat().st_size if t.exists() else -1
        estado = "pendiente" if tam < 0 else ("sin narración" if tam < 20 else "transcrito")
        out.append({"id": vid, "titulo": titulo, "min": (int(seg) // 60 if seg.isdigit() else None),
                    "url": f"https://youtu.be/{vid}", "estado": estado,
                    "palabras": len(t.read_text(encoding="utf-8").split()) if tam > 0 else 0})
    return out


def tutoriales() -> list[dict]:
    idx = json.loads((TUT / "indice.json").read_text()) if (TUT / "indice.json").exists() else {}
    por_archivo = {v["archivo"]: (k, v) for k, v in idx.items()}
    out = []
    for f in sorted(TUT.glob("*.md")):
        tarea, v = por_archivo.get(f.name, (None, {}))
        titulo = (f.read_text(encoding="utf-8").splitlines() or ["# ?"])[0].lstrip("# ").strip()
        out.append({"archivo": f.name, "titulo": titulo, "tarea": tarea, "citas": v.get("citas"),
                    "sin_doc": v.get("sin_documentar"), "fecha": v.get("fecha", "")[:16].replace("T", " ")})
    return out


def cortex_vivo() -> bool:
    try:
        return requests.get(f"{CORTEX_API}/health", headers=CORTEX_H, timeout=2).ok
    except requests.RequestException:
        return False


def contar_nodos() -> int:
    return len(list((RAIZ / ".cortex" / "tree").rglob("*.md")))


def tam_carpeta(p: Path) -> int:
    return sum(f.stat().st_size for f in p.rglob("*") if f.is_file()) if p.exists() else 0


# ─────────────────────────────── trabajos en segundo plano ─────────────────
TRABAJOS: list[dict] = []
_cerrojo = threading.Lock()


def lanzar(nombre: str, pasos: list[list[str]]) -> None:
    """Corre comandos en serie en un hilo; guarda el log. Uno a la vez para no pisarse."""
    t = {"id": len(TRABAJOS) + 1, "nombre": nombre, "estado": "en cola", "log": "", "inicio": time.strftime("%H:%M:%S")}
    TRABAJOS.insert(0, t)

    def correr():
        with _cerrojo:
            t["estado"] = "corriendo"
            for cmd in pasos:
                t["log"] += f"$ {' '.join(cmd[1:] if cmd[0] == PY else cmd)}\n"
                r = subprocess.run(cmd, cwd=RAIZ, capture_output=True, text=True)
                t["log"] += (r.stdout[-3000:] + r.stderr[-1500:])
                if r.returncode != 0:
                    t["estado"] = f"error ({r.returncode})"
                    return
            t["estado"] = "listo"
            t["fin"] = time.strftime("%H:%M:%S")

    threading.Thread(target=correr, daemon=True).start()


REINDEXAR = [[PY, "s10kb.py", "indexar"]]


# ─────────────────────────────── plantilla ─────────────────────────────────
CSS = """
:root{color-scheme:light;--bg:#f5f7f6;--surface:#fff;--soft:#f1f5f3;--ink:#18231f;--ink2:#4c5b54;--muted:#7b8982;--line:#e1e8e3;--brand:#087b63;--brand-soft:#e6f4ee;--nav:#17231f;--nav-muted:#aab8b0;--ok:#087b55;--warn:#9b6100;--bad:#b33b36;--r:10px;--sidebar:244px}
*{box-sizing:border-box}body{margin:0;font:14px/1.55 Inter,system-ui,-apple-system,Segoe UI,sans-serif;background:var(--bg);color:var(--ink)}
a{color:var(--brand);text-decoration:none}a:hover{text-decoration:underline}button{font:inherit}
.shell{min-height:100vh}.lado{position:fixed;z-index:20;inset:0 auto 0 0;width:var(--sidebar);background:var(--nav);color:#fff;padding:20px 12px 14px;display:flex;flex-direction:column;transition:width .18s ease;overflow:hidden}
.brand{height:42px;display:flex;align-items:center;gap:11px;padding:0 9px;margin:0 0 26px;color:#fff;text-decoration:none!important;white-space:nowrap}.brand-mark{width:30px;height:30px;display:grid;place-items:center;background:var(--brand);border-radius:8px;color:white}.brand strong{font-size:14px;font-weight:680;letter-spacing:0}.brand small{display:block;color:var(--nav-muted);font-size:11px;font-weight:450}
.nav-group{margin:0 0 19px}.nav-label{padding:0 11px;margin:0 0 6px;color:#82938a;font-size:10px;font-weight:700;text-transform:uppercase;letter-spacing:.08em;white-space:nowrap}.nav-link{min-height:38px;display:flex;align-items:center;gap:11px;padding:8px 11px;margin:2px 0;color:var(--nav-muted);border-radius:7px;white-space:nowrap;transition:background .15s,color .15s}.nav-link svg{width:17px;height:17px;flex:none}.nav-link:hover,.nav-link.on{color:#fff;background:#293a33;text-decoration:none}.nav-link.on{box-shadow:inset 2px 0 #57c49c}.nav-spacer{flex:1}.nav-foot{border-top:1px solid #34443c;padding:14px 10px 0;color:#91a198;font-size:11px;white-space:nowrap}.nav-running{display:block;color:#d8e7df;margin-bottom:8px}
.shell.collapsed .lado{width:68px}.shell.collapsed main{margin-left:68px}.shell.collapsed .brand{padding-left:7px}.shell.collapsed .brand-copy,.shell.collapsed .nav-label,.shell.collapsed .nav-text,.shell.collapsed .nav-foot span{display:none}.shell.collapsed .nav-link{justify-content:center;padding-inline:0}.shell.collapsed .nav-link.on{box-shadow:inset 0 -2px #57c49c}.shell.collapsed .nav-foot{padding-inline:0;text-align:center}
main{margin-left:var(--sidebar);min-height:100vh;padding:0 36px 48px;max-width:none;transition:margin-left .18s ease;min-width:0}.topbar{min-height:80px;display:flex;align-items:center;justify-content:space-between;gap:20px;border-bottom:1px solid var(--line);margin-bottom:30px}.top-left,.top-actions{display:flex;align-items:center;gap:13px;min-width:0}.page-heading h1{margin:0;font-size:21px;line-height:1.25;font-weight:680;letter-spacing:0}.page-heading p{margin:4px 0 0;color:var(--muted);font-size:12px}.top-actions{flex:none}.icon-button,.mobile-menu{width:36px;height:36px;display:grid;place-items:center;border:1px solid var(--line);border-radius:7px;color:var(--ink2);background:var(--surface);cursor:pointer}.icon-button:hover,.mobile-menu:hover{background:var(--soft);color:var(--ink)}.icon-button svg,.mobile-menu svg{width:17px;height:17px}.logout{display:flex;align-items:center;gap:8px;height:36px;padding:0 11px;border:1px solid var(--line);border-radius:7px;background:var(--surface);color:var(--ink2);font-size:12px;font-weight:600}.logout:hover{background:var(--soft);text-decoration:none}.logout svg{width:16px;height:16px}.info{position:relative}.info summary{list-style:none}.info summary::-webkit-details-marker{display:none}.info-pop{position:absolute;z-index:30;right:0;top:43px;width:min(300px,calc(100vw - 32px));padding:14px;background:var(--surface);border:1px solid var(--line);border-radius:8px;box-shadow:0 10px 28px #14241b1a;color:var(--ink2);font-size:12px}.info-pop strong{color:var(--ink);display:block;margin-bottom:5px}.mobile-menu{display:none}
.sub{display:none}h2{font-size:15px;line-height:1.35;margin:26px 0 10px;font-weight:650}.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(165px,1fr));gap:10px}.card{background:var(--surface);border:1px solid var(--line);border-radius:var(--r);padding:16px}.num{font-size:25px;line-height:1.2;font-weight:700;letter-spacing:0}.et{color:var(--ink2);font-size:12px;margin-top:4px}
table{width:100%;border-collapse:collapse;background:var(--surface);border:1px solid var(--line);border-radius:var(--r);overflow:hidden;font-size:13px}th,td{padding:10px 12px;border-bottom:1px solid var(--line);text-align:left;vertical-align:top}th{background:var(--soft);font-weight:650;color:var(--ink2);font-size:10px;text-transform:uppercase;letter-spacing:.06em}tr:last-child td{border-bottom:0}td b{font-weight:620}
.chip{display:inline-flex;align-items:center;min-height:22px;padding:2px 8px;border-radius:5px;font-size:11px;background:var(--brand-soft);color:#17624f;white-space:nowrap}.chip.ok{background:#e7f5ed;color:var(--ok)}.chip.warn{background:#fff4dc;color:var(--warn)}.chip.mal{background:#fae9e7;color:var(--bad)}
.btn{display:inline-flex;align-items:center;justify-content:center;gap:7px;min-height:36px;background:var(--brand);color:#fff;border:1px solid var(--brand);border-radius:7px;padding:7px 12px;font:inherit;font-size:12px;font-weight:650;cursor:pointer}.btn:hover{background:#06664f;color:#fff;text-decoration:none}.btn.sec{background:var(--surface);color:var(--ink2);border-color:var(--line)}.btn.sec:hover{background:var(--soft)}
input[type=text],input[type=url],input[type=password],select,textarea{font:inherit;font-size:13px;padding:8px 10px;border:1px solid var(--line);border-radius:7px;background:var(--surface);color:var(--ink);width:100%;min-height:36px}input:focus,select:focus,textarea:focus{outline:2px solid #087b6330;border-color:var(--brand)}form.fila{display:flex;gap:8px;align-items:center}form.fila input{flex:1}
.tabs{display:flex;gap:4px;margin-bottom:14px;padding-bottom:8px;border-bottom:1px solid var(--line);flex-wrap:wrap}.tabs a{padding:7px 10px;border-radius:6px;color:var(--ink2);font-size:12px}.tabs a:hover,.tabs a.on{background:var(--brand-soft);color:#145f4b;text-decoration:none}.tabs a.on{font-weight:650}
pre{background:#1a2721;color:#e2eae5;padding:14px;border-radius:8px;overflow:auto;font-size:12px;max-height:420px;white-space:pre-wrap}.doc{background:var(--surface);border:1px solid var(--line);border-radius:var(--r);padding:24px 28px}.doc h1{font-size:21px}.aviso{background:#edf5f1;border:1px solid #d8e9e0;border-radius:8px;padding:12px 14px;color:#315949}.muted{color:var(--muted)}.dos{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px}.card h2:first-child{margin-top:0}main>table{display:table}main{overflow-x:auto}
.shade{display:none}
@media(max-width:780px){.lado{width:260px;transform:translateX(-102%);transition:transform .2s ease}.shell.nav-open .lado{transform:translateX(0)}.shell.collapsed .lado{width:260px}.shell.collapsed main{margin-left:0}.shell.collapsed .brand-copy,.shell.collapsed .nav-label,.shell.collapsed .nav-text,.shell.collapsed .nav-foot span{display:initial}.shell.collapsed .nav-link{justify-content:flex-start;padding:8px 11px}.shell.collapsed .brand{padding-left:9px}.shell.collapsed .nav-label{display:block}.shell.collapsed .nav-foot{text-align:left;padding:14px 10px 0}.shell.nav-open .shade{display:block;position:fixed;inset:0;z-index:15;background:#101a16a8}.mobile-menu{display:grid}main,.shell.collapsed main{margin:0;padding:0 18px 36px}.topbar{min-height:70px;margin-bottom:23px;gap:8px}.top-left,.top-actions{gap:8px}.page-heading h1{font-size:18px}.page-heading p{max-width:52vw;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.collapse-toggle{display:none}.logout{width:36px;padding:0;justify-content:center}.logout span{display:none}.grid{grid-template-columns:repeat(2,minmax(0,1fr));gap:8px}.card{padding:13px}.num{font-size:22px}.dos{grid-template-columns:1fr}.doc{padding:18px}table{font-size:12px}th,td{padding:8px}.tabs{overflow-x:auto;flex-wrap:nowrap}.tabs a{white-space:nowrap}}
@media(prefers-reduced-motion:reduce){*,*::before,*::after{scroll-behavior:auto!important;transition:none!important}}
"""

MENU = [
    ("Conocimiento", [("resumen", "Resumen", "home"), ("fuentes", "Fuentes", "book"), ("agregar", "Agregar fuentes", "plus")]),
    ("Automatización", [("programacion", "Programación", "clock"), ("tutoriales_vista", "Tutoriales", "guide"), ("trabajos", "Trabajos", "activity")]),
    ("Sistema", [("cortex", "Cortex", "nodes"), ("herramientas", "Herramientas", "tools"), ("almacenamiento", "Almacenamiento", "storage")]),
]


def icono(nombre: str) -> str:
    paths = {
        "home": '<path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1h-6v-7h-4v7H4a1 1 0 0 1-1-1z"/>',
        "book": '<path d="M4 5.5A2.5 2.5 0 0 1 6.5 3H20v17H6.5A2.5 2.5 0 0 0 4 22z"/><path d="M4 5.5v14A2.5 2.5 0 0 1 6.5 17H20"/>',
        "plus": '<path d="M12 5v14M5 12h14"/>',
        "clock": '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
        "guide": '<path d="M4 4.5A2.5 2.5 0 0 1 6.5 2H20v18H6.5A2.5 2.5 0 0 0 4 22z"/><path d="M8 7h8M8 11h8M8 15h5"/>',
        "activity": '<path d="M3 12h4l3-8 4 16 3-8h4"/>',
        "nodes": '<circle cx="12" cy="5" r="2"/><circle cx="5" cy="19" r="2"/><circle cx="19" cy="19" r="2"/><path d="m11 7-5 10m7-10 5 10M7 19h10"/>',
        "tools": '<path d="M14.7 6.3a5 5 0 0 0-6.4 6.4L3 18l3 3 5.3-5.3a5 5 0 0 0 6.4-6.4L14 12l-3-3z"/>',
        "storage": '<ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3"/>',
        "menu": '<path d="M4 7h16M4 12h16M4 17h16"/>',
        "collapse": '<path d="m15 18-6-6 6-6"/>',
        "info": '<circle cx="12" cy="12" r="9"/><path d="M12 11v5m0-8h.01"/>',
        "logout": '<path d="M10 17l5-5-5-5m5 5H3"/><path d="M12 3h6a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-6"/>',
    }
    return f'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">{paths[nombre]}</svg>'


def pagina(activa: str, titulo: str, cuerpo: str, sub: str = "") -> str:
    menu = "".join(
        f'<section class="nav-group"><p class="nav-label">{grupo}</p>'
        + "".join(
            f'<a href="{url_for(k)}" class="nav-link {"on" if k == activa else ""}" title="{label}" '
            f'aria-current="{"page" if k == activa else "false"}">{icono(ico)}<span class="nav-text">{label}</span></a>'
            for k, label, ico in entradas
        )
        + "</section>"
        for grupo, entradas in MENU
    )
    corriendo = sum(1 for t in TRABAJOS if t["estado"] in ("corriendo", "en cola"))
    aviso = f'<a class="nav-running" href="{url_for("trabajos")}">{icono("activity")} {corriendo} en curso</a>' if corriendo else ""
    return f"""<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{html.escape(titulo)} · S10 Conocimiento</title><style>{CSS}</style></head><body>
<div class="shell" id="shell"><div class="shade" data-menu-close></div>
<nav class="lado" aria-label="Navegación principal">
<a class="brand" href="{url_for('resumen')}" title="S10 Conocimiento"><span class="brand-mark">{icono("book")}</span><span class="brand-copy"><strong>S10 Conocimiento</strong><small>Base para Metrín</small></span></a>
{menu}<div class="nav-spacer"></div><div class="nav-foot"><span>{aviso}</span><span>Optimiza 360 · Conocimiento</span></div></nav>
<main><header class="topbar"><div class="top-left"><button class="mobile-menu" type="button" aria-label="Abrir menú" aria-expanded="false" data-menu-toggle>{icono("menu")}</button>
<button class="icon-button collapse-toggle" type="button" aria-label="Contraer navegación" title="Contraer navegación" data-collapse-toggle>{icono("collapse")}</button>
<div class="page-heading"><h1>{html.escape(titulo)}</h1><p>{html.escape(sub)}</p></div></div>
<div class="top-actions"><details class="info"><summary class="icon-button" aria-label="Información de esta sección" title="Información">{icono("info")}</summary><div class="info-pop"><strong>Información</strong>{html.escape(sub or "Panel de conocimiento S10")}</div></details>
<a class="logout" href="{url_for('salir')}" title="Cerrar sesión">{icono("logout")}<span>Cerrar sesión</span></a></div></header>
{cuerpo}</main></div><script>
(()=>{{const shell=document.getElementById('shell'),mobile=document.querySelector('[data-menu-toggle]');
const collapsed=localStorage.getItem('s10-nav-collapsed')==='1';if(collapsed)shell.classList.add('collapsed');
document.querySelector('[data-collapse-toggle]')?.addEventListener('click',()=>{{shell.classList.toggle('collapsed');localStorage.setItem('s10-nav-collapsed',shell.classList.contains('collapsed')?'1':'0')}});
mobile?.addEventListener('click',()=>{{const open=shell.classList.toggle('nav-open');mobile.setAttribute('aria-expanded',String(open))}});
document.querySelector('[data-menu-close]')?.addEventListener('click',()=>{{shell.classList.remove('nav-open');mobile?.setAttribute('aria-expanded','false')}});
document.addEventListener('keydown',e=>{{if(e.key==='Escape'){{shell.classList.remove('nav-open');mobile?.setAttribute('aria-expanded','false')}}}});
}})();</script></body></html>"""


def e(x) -> str:
    return html.escape("" if x is None else str(x))


def chip_conf(nivel: str) -> str:
    clase = {"oficial": "ok", "tercero-sin-verificar": "warn", "academico": "", "estado-2013": ""}.get(nivel, "")
    return f'<span class="chip {clase}">{e(NIVELES.get(nivel, nivel))}</span>'


# ─────────────────────────────── rutas ─────────────────────────────────────
@app.route("/login", methods=["GET", "POST"])
def login():
    error = ""
    if request.method == "POST":
        env = leer_env()
        if secrets.compare_digest(request.form.get("usuario", ""), env.get("ADMIN_USUARIO", "admin")) and \
                secrets.compare_digest(request.form.get("clave", ""), env.get("ADMIN_CLAVE", "")):
            session.clear()
            session["ok"] = True
            destino = request.args.get("siguiente", "/")
            return redirect(destino if destino.startswith("/") and not destino.startswith("//") else "/")
        error = '<p class="chip mal">Usuario o clave incorrectos</p>'
        time.sleep(1)
    return f"""<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Acceso · S10 Conocimiento</title><style>{CSS} body{{display:grid;place-items:center;min-height:100vh}}</style></head><body>
<form method="post" class="card" style="width:340px;display:grid;gap:12px"><h1 style="font-size:20px">S10 Conocimiento</h1>
<span class="muted">Panel de administración</span>{error}
<input type="text" name="usuario" placeholder="Usuario" autofocus required><input type="password" name="clave" placeholder="Clave" required>
<button class="btn">Entrar</button></form></body></html>"""


@app.route("/salir")
def salir():
    session.clear()
    return redirect(url_for("login"))


@app.route("/")
@requiere_login
def resumen():
    docs, web, vids, tuts = documentos(), paginas_web(), videos(), tutoriales()
    frag = sum(1 for f in KB.glob("fragmentos*.jsonl") for l in f.read_text(encoding="utf-8").splitlines() if l.strip())
    muro = sum(1 for p in web if p.get("muro_de_miembros"))
    tarjetas = [
        (len(docs), "documentos PDF"), (sum(d["paginas"] or 0 for d in docs), "páginas de PDF"),
        (len(web) - muro, f"páginas web abiertas ({muro} tras login)"),
        (sum(1 for v in vids if v["estado"] == "transcrito"), f"videos transcritos de {len(vids)}"),
        (frag, "fragmentos en la base"), (contar_nodos(), "nodos en Cortex"), (len(tuts), "tutoriales"),
    ]
    grid = "".join(f'<div class="card"><div class="num">{n}</div><div class="et">{e(t)}</div></div>' for n, t in tarjetas)
    ult = "".join(f'<tr><td><a href="{url_for("ver_tutorial", nombre=t["archivo"])}">{e(t["titulo"])}</a></td><td>{e(t["citas"])}</td><td class="muted">{e(t["fecha"])}</td></tr>' for t in tuts[:6])
    cuerpo = f"""<div class="grid">{grid}</div>
<h2>Últimos tutoriales</h2><table><tr><th>Tutorial</th><th>Citas</th><th>Fecha</th></tr>{ult or '<tr><td colspan=3 class="muted">Aún no hay tutoriales.</td></tr>'}</table>
<h2>Accesos</h2><p><a class="btn" href="{url_for('entrar_cortex')}" target="_blank">Abrir Cortex con sesión</a>
<a class="btn sec" href="{url_for('agregar')}">Agregar fuentes</a></p>"""
    return pagina("resumen", "Resumen", cuerpo, "Base de conocimiento del ERP S10 para Metrín.")


@app.route("/cortex")
@requiere_login
def cortex():
    vivo = cortex_vivo()
    estado = '<span class="chip ok">en línea</span>' if vivo else '<span class="chip mal">apagado</span>'
    ramas = []
    for f in sorted((RAIZ / ".cortex" / "tree").glob("*.md")):
        if f.name == "_node.md":
            continue
        m = re.search(r"^title:\s*(.+)$", f.read_text(encoding="utf-8"), re.M)
        hijos = len(list((RAIZ / ".cortex" / "tree" / f.stem).glob("*.md"))) if (RAIZ / ".cortex" / "tree" / f.stem).is_dir() else 0
        ramas.append(f"<tr><td>{e(m.group(1) if m else f.stem)}</td><td class='muted'>{f.stem}</td><td>{hijos}</td></tr>")
    cuerpo = f"""<div class="card"><p>Tablero de Cortex: {estado} en <code>localhost:4748</code>.</p>
<p><a class="btn" href="{url_for('entrar_cortex')}" target="_blank">Entrar a Cortex con sesión iniciada</a>
{'' if vivo else f'<a class="btn sec" href="{url_for("encender_cortex")}">Encender tablero</a>'}</p>
<p class="muted">El enlace de entrada lo genera <code>cortexboard login</code> para el usuario dueño y vale 10 minutos.
Cada clic genera uno nuevo.</p></div>
<h2>Ramas del conocimiento</h2><table><tr><th>Rama</th><th>Ruta</th><th>Nodos hijos</th></tr>{''.join(ramas)}</table>"""
    return pagina("cortex", "Cortex", cuerpo, "El cerebro compartido: nodos por módulo, fuentes y tutoriales.")


@app.route("/cortex/entrar")
@requiere_login
def entrar_cortex():
    if not cortex_vivo():
        return pagina("cortex", "Cortex apagado", f'<p class="aviso">El tablero no está encendido. <a href="{url_for("encender_cortex")}">Encenderlo</a>.</p>')
    cli = [shutil.which("cortexboard")] if shutil.which("cortexboard") else ["npx", "-y", "cortexboard@0.4.0"]
    r = subprocess.run([*cli, "login", "--actor", "owner"], cwd=RAIZ,
                       capture_output=True, text=True, timeout=90)
    enlaces = re.findall(r"https?://\S+", r.stdout + r.stderr)
    if not enlaces:
        return pagina("cortex", "No se pudo generar el enlace", f"<pre>{e(r.stdout + r.stderr)}</pre>")
    # dentro de Docker el CLI ve su propio localhost; el navegador entra por el puerto publicado
    publico = os.environ.get("CORTEX_URL_PUBLICA")
    enlace = enlaces[-1]
    if publico:
        enlace = re.sub(r"^https?://[^/]+", publico.rstrip("/"), enlace)
    return redirect(enlace)


@app.route("/cortex/encender")
@requiere_login
def encender_cortex():
    if not cortex_vivo():
        if os.environ.get("EN_DOCKER"):
            return pagina("cortex", "Cortex apagado", '<p class="aviso">En Docker, Cortex es su propio servicio: <code>docker compose up -d cortex</code>.</p>')
        subprocess.Popen(["npx", "-y", "cortexboard@0.4.0", "start", "--port", "4748"], cwd=RAIZ,
                         stdout=open(RAIZ / "data" / "cortex.log", "a"), stderr=subprocess.STDOUT, start_new_session=True)
        for _ in range(20):
            time.sleep(1)
            if cortex_vivo():
                break
    return redirect(url_for("cortex"))


@app.route("/fuentes")
@requiere_login
def fuentes():
    tab = request.args.get("t", "pdf")
    tabs = [("pdf", "Documentos PDF"), ("web", "Portal de ayuda"), ("youtube", "YouTube"), ("internas", "Internas")]
    nav = '<div class="tabs">' + "".join(f'<a class="{"on" if k == tab else ""}" href="?t={k}">{v}</a>' for k, v in tabs) + "</div>"
    if tab == "pdf":
        filas = ""
        for d in documentos():
            opciones = "".join(f'<option value="{k}" {"selected" if k == d["confianza"] else ""}>{v}</option>' for k, v in NIVELES.items())
            filas += f"""<tr><td><b>{e(d['titulo'])}</b><div class="muted">{e(d['origen'])} · {d['kb']} KB · <a href="{e(d['url']) if d['url'].startswith('http') else '#'}" target="_blank">origen</a></div></td>
<td>{e(d['paginas'])}</td><td>{e(d['fragmentos'])}</td><td>{'<span class="chip">OCR</span>' if d['ocr'] else '<span class="muted">nativo</span>'}</td>
<td><form method="post" action="{url_for('cambiar_confianza')}" class="fila"><input type="hidden" name="csrf" value="{csrf()}">
<input type="hidden" name="url" value="{e(d['url'])}"><select name="nivel" onchange="this.form.submit()">{opciones}</select></form></td>
<td><a href="{url_for('ver_pdf', nombre=Path(d['archivo']).name)}" target="_blank">PDF</a> · <a href="{url_for('ver_texto', tipo='pdf', nombre=Path(d['archivo']).stem)}">texto</a></td></tr>"""
        cuerpo = nav + f"""<p class="muted">El nivel de confianza lo deciden ustedes; se aplica en la próxima reindexación.
<a href="{url_for('reindexar')}">Reindexar ahora</a>.</p>
<table><tr><th>Documento</th><th>Págs.</th><th>Fragm.</th><th>Texto</th><th>Confianza</th><th></th></tr>{filas}</table>"""
    elif tab == "web":
        filas = "".join(f"""<tr><td>{e(p.get('titulo'))}<div class="muted"><a href="{e(p['pagina'])}" target="_blank">{e(p['pagina'])}</a></div></td>
<td>{'<span class="chip warn">requiere login</span>' if p.get('muro_de_miembros') else '<span class="chip ok">abierta</span>'}</td>
<td>{f'<a href="{url_for("ver_texto", tipo="web", nombre=Path(p["archivo"]).stem)}">texto</a>' if p.get('archivo') and not p.get('muro_de_miembros') else ''}</td></tr>"""
                        for p in paginas_web())
        cuerpo = nav + f"<table><tr><th>Página</th><th>Acceso</th><th></th></tr>{filas}</table>"
    elif tab == "youtube":
        vids = videos()
        clase = {"transcrito": "ok", "sin narración": "warn", "pendiente": ""}
        filas = "".join(f"""<tr><td><b>{e(v['titulo'])}</b><div class="muted"><a href="{v['url']}" target="_blank">{v['url']}</a></div></td>
<td>{e(v['min']) if v['min'] is not None else '—'} min</td><td><span class="chip {clase[v['estado']]}">{v['estado']}</span></td><td>{v['palabras'] or ''}</td>
<td>{f'<a href="{url_for("ver_texto", tipo="yt", nombre=v["id"])}">transcripción</a>' if v['estado'] == 'transcrito' else ''}</td></tr>""" for v in vids)
        cuerpo = nav + f"""<p class="muted">Canal oficial «Marketing S10». Audio bajado con yt-dlp y transcrito en local con faster-whisper.
Los videos «sin narración» solo tienen música o pantalla: no aportan texto.</p>
<table><tr><th>Video</th><th>Duración</th><th>Estado</th><th>Palabras</th><th></th></tr>{filas}</table>"""
    else:
        internas = [("METRIN.md", "Ficha de Metrín, el asistente que usa esta base"), ("FUENTES.md", "Catálogo de fuentes"), ("README.md", "Cómo funciona el proyecto")]
        filas = "".join(f'<tr><td><a href="{url_for("ver_interna", nombre=n)}">{n}</a></td><td class="muted">{d}</td></tr>' for n, d in internas if (RAIZ / n).exists())
        cuerpo = nav + f"<table><tr><th>Archivo</th><th>Qué es</th></tr>{filas}</table>"
    return pagina("fuentes", "Fuentes", cuerpo, "Todo lo que alimenta la base, con su origen y su texto extraído.")


@app.route("/fuentes/confianza", methods=["POST"])
@requiere_login
def cambiar_confianza():
    validar_csrf()
    nivel = request.form.get("nivel")
    if nivel not in NIVELES:
        abort(400)
    f = DATA / "confianza_manual.json"
    d = confianza_manual()
    d[request.form.get("url", "")] = nivel
    f.write_text(json.dumps(d, indent=1, ensure_ascii=False), encoding="utf-8")
    return redirect(url_for("fuentes", t="pdf"))


def _nombre_seguro(nombre: str) -> str:
    if not re.fullmatch(r"[A-Za-z0-9._\-]+", nombre or ""):
        abort(404)
    return nombre


@app.route("/ver/pdf/<nombre>")
@requiere_login
def ver_pdf(nombre):
    p = DATA / "pdf" / _nombre_seguro(nombre)
    if not p.exists():
        abort(404)
    return send_file(p, mimetype="application/pdf")


@app.route("/ver/texto/<tipo>/<nombre>")
@requiere_login
def ver_texto(tipo, nombre):
    nombre = _nombre_seguro(nombre)
    if tipo == "pdf":
        p = DATA / "ocr" / f"{nombre}.txt"
        if not p.exists():
            p = DATA / "texto" / f"{nombre}.txt"
        if not p.exists():
            pdf = DATA / "pdf" / f"{nombre}.pdf"
            txt = subprocess.run(["pdftotext", "-layout", str(pdf), "-"], capture_output=True, text=True).stdout if pdf.exists() else ""
        else:
            txt = p.read_text(encoding="utf-8")
        titulo = f"Texto extraído · {nombre}"
    elif tipo == "web":
        p = DATA / "paginas" / f"{nombre}.md"
        txt, titulo = (p.read_text(encoding="utf-8") if p.exists() else ""), f"Página · {nombre}"
    elif tipo == "yt":
        p = DATA / "yt" / "transcripciones" / f"{nombre}.txt"
        txt, titulo = (p.read_text(encoding="utf-8") if p.exists() else ""), f"Transcripción · {nombre}"
    else:
        abort(404)
    if not txt:
        abort(404)
    paginas = txt.split("\f")
    cuerpo = "".join(f'<h2>{"Página " + str(i) if len(paginas) > 1 else ""}</h2><pre>{e(t)}</pre>' for i, t in enumerate(paginas, 1) if t.strip())
    return pagina("fuentes", titulo, f'<p><a href="javascript:history.back()">← volver</a></p>{cuerpo}')


@app.route("/ver/interna/<nombre>")
@requiere_login
def ver_interna(nombre):
    if nombre not in ("METRIN.md", "FUENTES.md", "README.md"):
        abort(404)
    h = md.markdown((RAIZ / nombre).read_text(encoding="utf-8"), extensions=["tables", "fenced_code"])
    return pagina("fuentes", nombre, f'<div class="doc">{h}</div>')


@app.route("/agregar")
@requiere_login
def agregar():
    c = csrf()
    cuerpo = f"""<div class="dos">
<div class="card"><h2 style="margin-top:0">Documento PDF</h2>
<p class="muted">Se guarda en <code>entrada/</code>, se registra, se le pasa OCR si viene escaneado o con capturas, y se reindexa.</p>
<form method="post" action="{url_for('agregar_pdf')}" enctype="multipart/form-data"><input type="hidden" name="csrf" value="{c}">
<p><input type="file" name="pdf" accept="application/pdf" multiple required></p>
<p><select name="nivel">{''.join(f'<option value="{k}">{v}</option>' for k, v in NIVELES.items())}</select></p>
<button class="btn">Subir y procesar</button></form></div>

<div class="card"><h2 style="margin-top:0">Video de YouTube</h2>
<p class="muted">Se baja solo el audio, se transcribe en local y se indexa por minutos.</p>
<form method="post" action="{url_for('agregar_youtube')}"><input type="hidden" name="csrf" value="{c}">
<p><input type="url" name="url" placeholder="https://www.youtube.com/watch?v=…" required></p>
<button class="btn">Agregar video</button></form></div>

<div class="card"><h2 style="margin-top:0">Página web pública</h2>
<p class="muted">Se guarda el texto de la página. Solo páginas abiertas: si pide login, queda registrada sin contenido.</p>
<form method="post" action="{url_for('agregar_web')}"><input type="hidden" name="csrf" value="{c}">
<p><input type="url" name="url" placeholder="https://documentacion.s10peru.com/…" required></p>
<button class="btn">Agregar página</button></form></div>

<div class="card"><h2 style="margin-top:0">Portal de miembros de S10</h2>
<p class="muted">Cuando tengan la cuenta de miembro: complétenla en <code>.env</code> (S10_USUARIO, S10_CLAVE) y corran el rastreo completo.</p>
<form method="post" action="{url_for('rastrear_portal')}"><input type="hidden" name="csrf" value="{c}">
<button class="btn sec">Rastrear portal con la cuenta de .env</button></form></div>
</div>"""
    return pagina("agregar", "Agregar fuentes", cuerpo, "Cada alta corre en segundo plano; el avance se ve en Trabajos.")


@app.route("/agregar/pdf", methods=["POST"])
@requiere_login
def agregar_pdf():
    validar_csrf()
    (RAIZ / "entrada").mkdir(exist_ok=True)
    nombres = []
    for f in request.files.getlist("pdf"):
        if not f.filename.lower().endswith(".pdf"):
            continue
        base = re.sub(r"[^A-Za-z0-9 ._\-áéíóúñÁÉÍÓÚÑ]", "", Path(f.filename).name)[:120] or "documento.pdf"
        destino = RAIZ / "entrada" / base
        f.save(destino)
        if destino.read_bytes()[:4] != b"%PDF":
            destino.unlink()
            continue
        nombres.append(base)
    if not nombres:
        abort(400, "no llegó ningún PDF válido")
    nivel = request.form.get("nivel", "oficial")
    lanzar(f"PDF: {', '.join(nombres)}", [[PY, "s10kb.py", "importar"], [PY, "s10kb.py", "ocr", "--completo"],
                                          [PY, "-c", f"import json,pathlib;"
                                                     f"m=[json.loads(l) for l in open('data/manifiesto.jsonl') if l.strip()];"
                                                     f"f=pathlib.Path('data/confianza_manual.json');d=json.loads(f.read_text()) if f.exists() else {{}};"
                                                     f"[d.__setitem__(x['url'],{nivel!r}) for x in m if x['url'].replace('manual://','') in {nombres!r}];"
                                                     f"f.write_text(json.dumps(d,indent=1,ensure_ascii=False))"]] + REINDEXAR)
    return redirect(url_for("trabajos"))


@app.route("/agregar/youtube", methods=["POST"])
@requiere_login
def agregar_youtube():
    validar_csrf()
    url = request.form.get("url", "")
    m = re.search(r"(?:v=|youtu\.be/|shorts/)([A-Za-z0-9_-]{11})", url)
    if not m:
        abort(400, "no reconozco el enlace de YouTube")
    vid = m.group(1)
    ytdlp = str(RAIZ / ".venv" / "bin" / "yt-dlp")
    alta = (f"import subprocess,pathlib;f=pathlib.Path('data/yt/videos.txt');f.parent.mkdir(parents=True,exist_ok=True);"
            f"txt=f.read_text() if f.exists() else '';"
            f"r=subprocess.run([{ytdlp!r},'--skip-download','--print','%(id)s|%(duration)s|%(title)s','https://youtu.be/{vid}'],capture_output=True,text=True);"
            f"l=r.stdout.strip().splitlines()[-1] if r.stdout.strip() else '{vid}|NA|Video {vid}';"
            f"(f.write_text(txt.rstrip()+'\\n'+l+'\\n') if '{vid}|' not in txt else None);print('alta:',l)")
    lanzar(f"YouTube: {vid}", [[PY, "-c", alta], [PY, "youtube.py", "audio"], [PY, "youtube.py", "transcribir"],
                               [PY, "youtube.py", "indexar"]])
    return redirect(url_for("trabajos"))


@app.route("/agregar/web", methods=["POST"])
@requiere_login
def agregar_web():
    validar_csrf()
    url = request.form.get("url", "")
    if not url.startswith(("http://", "https://")):
        abort(400)
    script = (f"import s10kb,requests;from bs4 import BeautifulSoup;s10kb.cargar_env();c=s10kb.Cliente();"
              f"s10kb.rastrear(c,[s10kb.normalizar({url!r})])")
    lanzar(f"Web: {url}", [[PY, "-c", script]] + REINDEXAR)
    return redirect(url_for("trabajos"))


@app.route("/agregar/portal", methods=["POST"])
@requiere_login
def rastrear_portal():
    validar_csrf()
    env = leer_env()
    if not env.get("S10_USUARIO") or not env.get("S10_CLAVE"):
        return pagina("agregar", "Falta la cuenta de miembro", '<p class="aviso">Completa S10_USUARIO y S10_CLAVE en <code>.env</code> y vuelve a intentarlo.</p>')
    lanzar("Portal de miembros S10", [[PY, "s10kb.py", "todo"], [PY, "s10kb.py", "ocr", "--completo"]] + REINDEXAR)
    return redirect(url_for("trabajos"))


@app.route("/reindexar")
@requiere_login
def reindexar():
    lanzar("Reindexar base", REINDEXAR)
    return redirect(url_for("trabajos"))


@app.route("/tutoriales", methods=["GET", "POST"])
@requiere_login
def tutoriales_vista():
    if request.method == "POST":
        validar_csrf()
        if request.form.get("modo") == "auto":
            lanzar("Tutoriales automáticos", [[PY, "tutor.py", "auto", "--max", str(int(request.form.get("max", 5)))]])
        else:
            tarea = request.form.get("tarea", "").strip()
            if tarea:
                lanzar(f"Tutorial: {tarea[:60]}", [[PY, "tutor.py", "preguntar", tarea]])
        return redirect(url_for("trabajos"))
    c = csrf()
    filas = "".join(f"""<tr><td><a href="{url_for('ver_tutorial', nombre=t['archivo'])}">{e(t['titulo'])}</a><div class="muted">{e(t['tarea'])}</div></td>
<td>{e(t['citas'])}</td><td>{e(t['sin_doc'])}</td><td class="muted">{e(t['fecha'])}</td></tr>""" for t in tutoriales())
    cuerpo = f"""<div class="dos"><div class="card"><h2 style="margin-top:0">Pedir un tutorial</h2>
<form method="post"><input type="hidden" name="csrf" value="{c}"><p><input type="text" name="tarea" placeholder="¿Cómo registro los feriados del año en Nóminas?" required></p>
<button class="btn">Generar</button></form></div>
<div class="card"><h2 style="margin-top:0">Generación automática</h2><p class="muted">Propone tareas desde el árbol de Cortex y genera las que falten.
Rehace las que se hicieron con una versión anterior de la base.</p>
<form method="post" class="fila"><input type="hidden" name="csrf" value="{c}"><input type="hidden" name="modo" value="auto">
<select name="max" style="width:auto"><option>3</option><option selected>5</option><option>10</option></select><button class="btn">Correr</button></form></div></div>
<h2>Tutoriales</h2><table><tr><th>Tutorial</th><th>Citas</th><th>Sin documentar</th><th>Fecha</th></tr>{filas or '<tr><td colspan=4 class="muted">Aún no hay.</td></tr>'}</table>"""
    return pagina("tutoriales_vista", "Tutoriales", cuerpo, "Guías paso a paso generadas con citas a las fuentes. También quedan en Cortex, rama tutoriales.")


@app.route("/tutoriales/<nombre>")
@requiere_login
def ver_tutorial(nombre):
    p = TUT / _nombre_seguro(nombre)
    if not p.exists() or p.suffix != ".md":
        abort(404)
    h = md.markdown(p.read_text(encoding="utf-8"), extensions=["tables", "fenced_code"])
    return pagina("tutoriales_vista", "Tutorial", f'<p><a href="{url_for("tutoriales_vista")}">← tutoriales</a> · <a href="{url_for("descargar_tutorial", nombre=nombre)}">descargar .md</a></p><div class="doc">{h}</div>')


@app.route("/tutoriales/<nombre>/md")
@requiere_login
def descargar_tutorial(nombre):
    p = TUT / _nombre_seguro(nombre)
    if not p.exists():
        abort(404)
    return send_file(p, as_attachment=True)


HERRAMIENTAS = [
    ("s10kb.py", "Rastreo del portal y de s10peru.com, descarga en colas, OCR e indexación en fragmentos.", "propio", None),
    ("tutor.py", "Genera tutoriales: planificar → recuperar (BM25 + Cortex) → evaluar → generar, con citas.", "propio", None),
    ("youtube.py", "Inventario del canal oficial, audio y transcripción local de los videos.", "propio", None),
    ("Cortex (cortexboard)", "Árbol de conocimiento compartido por módulos, con API, tablero y MCP.", "npm cortexboard@0.4.0", None),
    ("pdftotext / pdftoppm", "Extraen el texto nativo de los PDF y rasterizan páginas para el OCR.", "poppler", ["pdftotext", "-v"]),
    ("OCR", "Reconoce texto en español de PDF escaneados y capturas: Vision de macOS en la Mac, tesseract en Docker/Linux.", "herramientas/ocr.swift · tesseract-ocr-spa", ["tesseract", "--version"]),
    ("yt-dlp", "Baja solo el audio de cada video y lee su título y duración.", "pip", [str(RAIZ / ".venv/bin/yt-dlp"), "--version"]),
    ("faster-whisper", "Transcribe el audio de los videos en local, sin enviar nada fuera.", "pip", None),
    ("ffmpeg", "Convierte el audio antes de transcribir.", "homebrew", ["ffmpeg", "-version"]),
    ("Claude", "Modelo que planifica, evalúa y redacta los tutoriales (SDK de Anthropic o Claude Code).", "Anthropic", None),
    ("Metrín (metrin/)", "Chat de Metrín: RAG en Go con embeddings locales, índice chromem y respuesta con Ollama.", "Go · Ollama", None),
]


@app.route("/herramientas")
@requiere_login
def herramientas():
    filas = ""
    for nombre, que, origen, cmd in HERRAMIENTAS:
        version = ""
        if cmd:
            try:
                r = subprocess.run(cmd, capture_output=True, text=True, timeout=10)
                version = (r.stdout or r.stderr).strip().splitlines()[0][:60]
            except Exception:
                version = "no instalado"
        filas += f"<tr><td><b>{e(nombre)}</b></td><td>{e(que)}</td><td class='muted'>{e(origen)}</td><td class='muted'>{e(version)}</td></tr>"
    cuerpo = f"<table><tr><th>Herramienta</th><th>Para qué</th><th>Origen</th><th>Versión</th></tr>{filas}</table>"
    return pagina("herramientas", "Herramientas", cuerpo, "Todo corre en local salvo el modelo de Claude.")


@app.route("/almacenamiento")
@requiere_login
def almacenamiento():
    carpetas = [("data/pdf", "PDF originales"), ("data/yt/audio", "Audio de videos"), ("data/ocr", "Texto de OCR"),
                ("data/paginas", "Páginas web"), ("kb", "Base de fragmentos"), ("tutoriales", "Tutoriales"), (".cortex", "Cortex")]
    filas = "".join(f"<tr><td>{d}</td><td><code>{c}</code></td><td>{tam_carpeta(RAIZ / c) / 1e6:.1f} MB</td></tr>" for c, d in carpetas)
    total = sum(tam_carpeta(RAIZ / c) for c, _ in carpetas) / 1e6
    cuerpo = f"""<table><tr><th>Qué</th><th>Carpeta</th><th>Tamaño</th></tr>{filas}<tr><td><b>Total</b></td><td></td><td><b>{total:.1f} MB</b></td></tr></table>
<h2>Mover a un almacén S3 (Garage)</h2>
<div class="aviso">Hoy todo vive en esta Mac. Para servir la base a Metrín desde un servidor conviene subir los binarios
(PDF, audio, OCR) a un cubo S3 y dejar en disco solo la base de fragmentos y Cortex.
Pendiente de decidir: en qué Garage y con qué credenciales. Se recomienda un cubo y una clave propios de S10,
separados de los de PjgFactSalud.</div>"""
    return pagina("almacenamiento", "Almacenamiento", cuerpo, "Dónde está cada cosa y cuánto pesa.")


FRECUENCIAS = [(6, "cada 6 h"), (12, "cada 12 h"), (24, "cada día"), (72, "cada 3 días"), (168, "cada semana"), (720, "cada mes")]


def _programador():
    sys.path.insert(0, str(RAIZ))
    import programador  # noqa: E402  (mismo catálogo de fuentes que el servicio)
    return programador


@app.route("/programacion", methods=["GET", "POST"])
@requiere_login
def programacion():
    prog = _programador()
    conf = prog.config()
    if request.method == "POST":
        validar_csrf()
        fid = request.form.get("fuente")
        if fid not in prog.FUENTES:
            abort(400)
        if request.form.get("accion") == "ahora":
            prog.PEDIDOS.mkdir(parents=True, exist_ok=True)
            (prog.PEDIDOS / fid).write_text(time.strftime("%Y-%m-%dT%H:%M:%S"))
        else:
            horas = int(request.form.get("cada_horas", 24))
            conf[fid] = {"cada_horas": horas, "activo": request.form.get("activo") == "1"}
            prog.guardar(prog.CONFIG, conf)
        return redirect(url_for("programacion"))
    estado = prog.leer(prog.ESTADO)
    c = csrf()
    vivo = (DATA / "programacion_estado.json").exists() and time.time() - (DATA / "programacion_estado.json").stat().st_mtime < 7 * 86400
    filas = ""
    for fid, f in prog.FUENTES.items():
        k, e = conf.get(fid, {}), estado.get(fid, {})
        pedido = (prog.PEDIDOS / fid).exists()
        opciones = "".join(f'<option value="{h}" {"selected" if h == k.get("cada_horas") else ""}>{t}</option>' for h, t in FRECUENCIAS)
        est = e.get("estado", "—")
        clase = {"ok": "ok", "error": "mal", "corriendo": "warn"}.get(est, "")
        prox = time.strftime("%Y-%m-%d %H:%M", time.localtime(e["proxima"])) if e.get("proxima") and k.get("activo") else "—"
        registro = (prog.LOGS / f"{fid}.log").exists()
        filas += f"""<tr><td><b>{e_(f['nombre'])}</b><div class="muted">{' → '.join(' '.join(p) for p in f['pasos'])}</div></td>
<td><form method="post" class="fila"><input type="hidden" name="csrf" value="{c}"><input type="hidden" name="fuente" value="{fid}">
<select name="cada_horas" style="width:auto">{opciones}</select>
<label class="muted" style="display:flex;gap:4px;align-items:center"><input type="checkbox" name="activo" value="1" {"checked" if k.get("activo") else ""}>activa</label>
<button class="btn sec">Guardar</button></form></td>
<td>{e_(e.get('ultima', '—'))}<div class="muted">próxima: {prox}</div></td>
<td><span class="chip {clase}">{e_('pedido en cola' if pedido else est)}</span><div class="muted">{e_(e.get('resumen', ''))}</div></td>
<td><form method="post"><input type="hidden" name="csrf" value="{c}"><input type="hidden" name="fuente" value="{fid}">
<input type="hidden" name="accion" value="ahora"><button class="btn">Ejecutar ahora</button></form>
{f'<a href="{url_for("registro_programacion", fid=fid)}">registro</a>' if registro else ''}</td></tr>"""
    aviso = "" if vivo else '<p class="aviso">El programador no está corriendo. En Docker: <code>docker compose up -d programador</code>; fuera de Docker: <code>.venv/bin/python programador.py</code>.</p>'
    refresco = '<meta http-equiv="refresh" content="10">' if any(v.get("estado") == "corriendo" for v in estado.values()) or any(prog.PEDIDOS.glob("*")) else ""
    cuerpo = f"""{refresco}{aviso}<table><tr><th>Fuente</th><th>Frecuencia</th><th>Última corrida</th><th>Resultado</th><th></th></tr>{filas}</table>
<p class="muted" style="margin-top:14px">Cada corrida con cambios reindexa la base y Metrín recarga su índice. Todo queda registrado en Cortex:
nodo <code>fuentes/programacion</code> y el historial de actividad. <a href="{url_for('entrar_cortex')}" target="_blank">Abrir Cortex</a>.</p>"""
    return pagina("programacion", "Programación", cuerpo, "Cada cuánto se actualiza cada fuente, cuándo corrió y qué trajo.")


def e_(x) -> str:
    return html.escape("" if x is None else str(x))


@app.route("/programacion/<fid>/registro")
@requiere_login
def registro_programacion(fid):
    prog = _programador()
    if fid not in prog.FUENTES:
        abort(404)
    p = prog.LOGS / f"{fid}.log"
    txt = p.read_text(encoding="utf-8") if p.exists() else "Sin registro todavía."
    return pagina("programacion", f"Registro · {prog.FUENTES[fid]['nombre']}", f'<p><a href="{url_for("programacion")}">← programación</a></p><pre>{e_(txt[-20000:])}</pre>')


@app.route("/trabajos")
@requiere_login
def trabajos():
    auto = RAIZ / "tutoriales" / "auto.log"
    filas = "".join(f"""<div class="card" style="margin-bottom:10px"><b>#{t['id']} {e(t['nombre'])}</b>
<span class="chip {'ok' if t['estado'] == 'listo' else 'mal' if t['estado'].startswith('error') else ''}">{e(t['estado'])}</span>
<span class="muted">{t['inicio']}{' → ' + t['fin'] if t.get('fin') else ''}</span><pre>{e(t['log'][-4000:]) or '…'}</pre></div>""" for t in TRABAJOS)
    extra = f"<h2>Última corrida automática por consola</h2><pre>{e(auto.read_text(encoding='utf-8')[-3000:])}</pre>" if auto.exists() else ""
    refresco = '<meta http-equiv="refresh" content="5">' if any(t["estado"] in ("corriendo", "en cola") for t in TRABAJOS) else ""
    return pagina("trabajos", "Trabajos", refresco + (filas or '<p class="muted">Sin trabajos en esta sesión del panel.</p>') + extra)


if __name__ == "__main__":
    puerto = int(os.environ.get("ADMIN_PUERTO", "4750"))
    print(f"Panel en http://127.0.0.1:{puerto}")
    app.run(host=os.environ.get("ADMIN_HOST", "127.0.0.1"), port=puerto, debug=False, threaded=True)
