"""Tests del visor de trazas (trazas.py). unittest — no hay pytest en .venv.

  cd s10-conocimiento && .venv/bin/python -m unittest trazas.test_trazas -v
"""
from __future__ import annotations

import json
import os
import sys
import tempfile
import unittest
from pathlib import Path

from trazas.trazas import (
    RepositorioArchivo,
    RepositorioMock,
    RepositorioPostgres,
    RepositorioTrazas,
    crear_app,
    estadisticas,
    filtrar,
    repo_desde_env,
    resumen,
)


class RepositorioFake(RepositorioTrazas):
    """Trazas chicas y controladas: choice, score y noul."""

    TRAZAS = [
        {
            "id": "T1",
            "ts": "2026-09-30T10:00:00-05:00",
            "modelo": "nemotron-mini",
            "origen": "metrin-chat",
            "plantilla": "jev-default",
            "ms": 100,
            "tokens_entrada": 1000,
            "estado": {"mensaje": "¿Cómo registro una guía de remisión?"},
            "preguntas": [
                {
                    "id": "modulo",
                    "tipo": "choice",
                    "instrucciones": "¿Qué módulo responde?",
                    "criterios": {"almacenes": "Guías y stock.", "compras": "Pedidos."},
                    "respuesta": {
                        "tipo": "choice",
                        "choice": "almacenes",
                        "probabilidades": {"almacenes": 0.8, "compras": 0.2},
                        "confianza": 0.52,
                    },
                    "ms": 50,
                }
            ],
            "correcta": True,
        },
        {
            "id": "T2",
            "ts": "2026-09-30T11:00:00-05:00",
            "modelo": "glm-4-flash",
            "origen": "reportes",
            "plantilla": "jev-default",
            "ms": 200,
            "tokens_entrada": 2000,
            "estado": {"mensaje": "UIT 2024"},
            "preguntas": [
                {
                    "id": "urgencia",
                    "tipo": "score",
                    "instrucciones": {"texto": "¿Qué tan urgente?"},
                    "criterios": ["Baja.", "Media.", "Alta."],
                    "respuesta": {
                        "tipo": "score",
                        "score": 2.4,
                        "probabilidades": {"0": 0.1, "1": 0.2, "2": 0.7},
                        "legend": {"0": "Baja.", "1": "Media.", "2": "Alta."},
                    },
                    "ms": 90,
                },
                {
                    "id": "segura",
                    "tipo": "noul",
                    "instrucciones": "¿Es seguro?",
                    "criterios": {"true": "Sí.", "false": "No."},
                    "respuesta": {"tipo": "noul", "noul": 0.9},
                    "ms": 10,
                },
            ],
            "correcta": False,
        },
        {
            "id": "T3",
            "ts": "2026-10-01T08:00:00-05:00",
            "modelo": "glm-4-flash",
            "origen": "metrin-chat",
            "plantilla": "jev-default",
            "ms": 50,
            "tokens_entrada": 500,
            "estado": "texto plano",
            "preguntas": [],
            "correcta": None,
        },
    ]

    def listar(self) -> list[dict]:
        return self.TRAZAS

    def obtener(self, traza_id: str) -> dict | None:
        for t in self.TRAZAS:
            if t["id"] == traza_id:
                return t
        return None

    def origen(self) -> str:
        return "repositorio de prueba"


class TestVista(unittest.TestCase):
    def setUp(self):
        self.app = crear_app(RepositorioFake()).test_client()

    def test_salud(self):
        r = self.app.get("/api/salud")
        self.assertEqual(r.status_code, 200)
        self.assertTrue(r.json["ok"])
        self.assertEqual(r.json["origen"], "repositorio de prueba")

    def test_visor_sirve_html(self):
        r = self.app.get("/")
        self.assertEqual(r.status_code, 200)
        self.assertIn(b"Trazas", r.data)

    def test_listar_orden_y_resumen(self):
        # Más nuevas primero: T3, T2, T1.
        r = self.app.get("/api/trazas")
        self.assertEqual(r.status_code, 200)
        datos = r.json
        self.assertEqual(datos["total"], 3)
        self.assertEqual([t["id"] for t in datos["trazas"]], ["T3", "T2", "T1"])
        t1 = datos["trazas"][2]
        self.assertEqual(t1["modelo"], "nemotron-mini")
        self.assertEqual(t1["tipos"], ["choice"])
        self.assertEqual(t1["n_preguntas"], 1)
        self.assertEqual(t1["confianza"], 0.52)
        self.assertEqual(t1["titulo"], "¿Qué módulo responde?")

    def test_resumen_titulo_dict_y_sin_preguntas(self):
        t2 = resumen(RepositorioFake.TRAZAS[1])
        self.assertEqual(t2["titulo"], "¿Qué tan urgente?")
        self.assertIsNone(t2["confianza"])  # score y noul no traen confianza
        t3 = resumen(RepositorioFake.TRAZAS[2])
        self.assertEqual(t3["titulo"], "traza sin instrucciones")
        self.assertEqual(t3["tipos"], [])

    def test_filtrar_texto(self):
        self.assertEqual(len(filtrar(RepositorioFake.TRAZAS, {"q": "guía"})), 1)
        self.assertEqual(len(filtrar(RepositorioFake.TRAZAS, {"q": "uit"})), 1)
        self.assertEqual(len(filtrar(RepositorioFake.TRAZAS, {"q": "inexistente"})), 0)

    def test_filtrar_tipo(self):
        self.assertEqual({t["id"] for t in filtrar(RepositorioFake.TRAZAS, {"tipo": "noul"})}, {"T2"})
        self.assertEqual({t["id"] for t in filtrar(RepositorioFake.TRAZAS, {"tipo": "choice"})}, {"T1"})

    def test_filtrar_modelo(self):
        self.assertEqual({t["id"] for t in filtrar(RepositorioFake.TRAZAS, {"modelo": "glm-4-flash"})}, {"T2", "T3"})

    def test_filtrar_falladas(self):
        self.assertEqual({t["id"] for t in filtrar(RepositorioFake.TRAZAS, {"fallidas": "1"})}, {"T2"})
        self.assertEqual({t["id"] for t in filtrar(RepositorioFake.TRAZAS, {"fallidas": "true"})}, {"T2"})
        self.assertEqual(len(filtrar(RepositorioFake.TRAZAS, {"fallidas": "0"})), 3)

    def test_filtrar_compuesto(self):
        solo = filtrar(RepositorioFake.TRAZAS, {"modelo": "glm-4-flash", "fallidas": "1"})
        self.assertEqual([t["id"] for t in solo], ["T2"])

    def test_detalle_y_404(self):
        r = self.app.get("/api/trazas/T1")
        self.assertEqual(r.status_code, 200)
        self.assertEqual(r.json["id"], "T1")
        self.assertEqual(len(r.json["preguntas"]), 1)
        r = self.app.get("/api/trazas/NO-EXISTE")
        self.assertEqual(r.status_code, 404)

    def test_paginacion(self):
        r = self.app.get("/api/trazas?limite=1&offset=1")
        self.assertEqual([t["id"] for t in r.json["trazas"]], ["T2"])

    def test_estadisticas(self):
        e = estadisticas(RepositorioFake.TRAZAS)
        self.assertEqual(e["total"], 3)
        self.assertEqual(e["por_modelo"], {"nemotron-mini": 1, "glm-4-flash": 2})
        self.assertEqual(e["por_tipo"], {"choice": 1, "score": 1, "noul": 1})
        # Confianza: solo T1 (0.52); score y noul no aportan.
        self.assertEqual(e["confianza_media"], 0.52)
        self.assertEqual(e["confianza_min"], 0.52)
        self.assertEqual(e["confianza_max"], 0.52)
        # Latencias: 50, 100, 200.
        self.assertEqual(e["latencia"], {"media": 117, "p50": 100, "p95": 200})
        self.assertEqual(e["aciertos"], {"n": 1, "total": 2, "pct": 50.0})
        self.assertEqual(e["latencia_cubetas"]["etiquetas"], ["<50", "50–100", "100–200", "200–500", "≥500"])
        # 50 cae en 50–100, 100 en 100–200, 200 en 200–500 (bordes inclusivos por abajo).
        self.assertEqual(e["latencia_cubetas"]["valores"], [0, 1, 1, 1, 0])

    def test_estadisticas_cubetas_confianza(self):
        # 0.52 cae en la cubeta 0.4–0.6 (índice 1); 3 bordes = 4 cubetas.
        e = estadisticas(RepositorioFake.TRAZAS)
        self.assertEqual(e["confianza_cubetas"]["etiquetas"], ["<0.4", "0.4–0.6", "0.6–0.8", "≥0.8"])
        self.assertEqual(e["confianza_cubetas"]["valores"], [0, 1, 0, 0])

    def test_estadisticas_vacio(self):
        e = estadisticas([])
        self.assertEqual(e["total"], 0)
        self.assertIsNone(e["confianza_media"])
        self.assertEqual(e["latencia"], {"media": 0, "p50": 0, "p95": 0})
        self.assertEqual(e["aciertos"], {"n": 0, "total": 0, "pct": None})
        self.assertEqual(e["latencia_cubetas"]["valores"], [0] * 5)

    def test_modelos(self):
        r = self.app.get("/api/modelos")
        self.assertEqual(r.json["modelos"], ["glm-4-flash", "nemotron-mini"])

    def test_estadisticas_respeta_filtros(self):
        r = self.app.get("/api/estadisticas?modelo=glm-4-flash")
        self.assertEqual(r.json["total"], 2)


class TestMockReal(unittest.TestCase):
    """El mock del repo (21 trazas): humo de integración con los datos de verdad."""

    def setUp(self):
        from pathlib import Path
        from trazas.trazas import RAIZ, RepositorioMock

        self.app = crear_app(RepositorioMock(RAIZ / "mock" / "trazas.json")).test_client()

    def test_salud_y_lista(self):
        self.assertTrue(self.app.get("/api/salud").json["ok"])
        datos = self.app.get("/api/trazas").json
        self.assertGreater(datos["total"], 10)
        self.assertEqual(len(datos["trazas"]), datos["total"])
        for t in datos["trazas"]:
            self.assertIn("id", t)
            self.assertIn("confianza", t)
            self.assertIn("correcta", t)

    def test_estadisticas_mock(self):
        e = self.app.get("/api/estadisticas").json
        self.assertEqual(e["total"], datos_total())
        self.assertGreater(e["latencia"]["p95"], 0)
        self.assertIsNotNone(e["confianza_media"])

    def test_busqueda_dominio(self):
        r = self.app.get("/api/trazas?q=guía%20de%20remisión")
        self.assertGreaterEqual(r.json["total"], 1)

    def test_detalle_ultima(self):
        ultima = self.app.get("/api/trazas").json["trazas"][0]
        r = self.app.get("/api/trazas/" + ultima["id"])
        self.assertEqual(r.status_code, 200)
        self.assertTrue(r.json["preguntas"])


def datos_total() -> int:
    import json
    from pathlib import Path
    from trazas.trazas import RAIZ

    with open(RAIZ / "mock" / "trazas.json", encoding="utf-8") as f:
        return len(json.load(f)["trazas"])


class TestRepositorioArchivo(unittest.TestCase):
    """El JSONL que emite el cliente Go de Metrín."""

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.ruta = Path(self.tmp.name) / "trazas.jsonl"

    def tearDown(self):
        self.tmp.cleanup()

    def _escribir(self, *trazas):
        with open(self.ruta, "a", encoding="utf-8") as f:
            for t in trazas:
                f.write(json.dumps(t, ensure_ascii=False) + "\n")

    def test_inexistente_vacio(self):
        self.assertEqual(RepositorioArchivo(self.ruta).listar(), [])

    def test_listar_y_obtener(self):
        t1 = {"id": "a1", "ts": "2026-10-01T10:00:00+00:00", "preguntas": []}
        t2 = {"id": "b2", "ts": "2026-10-01T11:00:00+00:00", "preguntas": []}
        self._escribir(t1, t2)
        repo = RepositorioArchivo(self.ruta)
        self.assertEqual(repo.listar(), [t1, t2])
        self.assertEqual(repo.obtener("b2"), t2)
        self.assertIsNone(repo.obtener("zzz"))

    def test_revalida_al_cambiar(self):
        self._escribir({"id": "a1", "ts": "t", "preguntas": []})
        repo = RepositorioArchivo(self.ruta)
        self.assertEqual(len(repo.listar()), 1)
        self._escribir({"id": "a2", "ts": "t", "preguntas": []})
        self.assertEqual(len(repo.listar()), 2)

    def test_origen(self):
        self.assertIn("archivo", RepositorioArchivo(self.ruta).origen())


class TestRepoDesdeEnv(unittest.TestCase):
    def test_mock_por_defecto(self):
        with tempfile.TemporaryDirectory() as tmp:
            os.environ["TRAZAS_REPO"] = "mock"
            try:
                self.assertIsInstance(repo_desde_env(), RepositorioMock)
            finally:
                os.environ.pop("TRAZAS_REPO", None)

    def test_archivo(self):
        with tempfile.TemporaryDirectory() as tmp:
            ruta = str(Path(tmp) / "t.jsonl")
            os.environ["TRAZAS_REPO"] = "archivo:" + ruta
            try:
                repo = repo_desde_env()
                self.assertIsInstance(repo, RepositorioArchivo)
                self.assertEqual(repo.ruta, Path(ruta))
            finally:
                os.environ.pop("TRAZAS_REPO", None)

    def test_postgres_sin_dsn_falla(self):
        os.environ["TRAZAS_REPO"] = "postgres"
        os.environ.pop("TRAZAS_PG_DSN", None)
        try:
            with self.assertRaises(RuntimeError):
                repo_desde_env()
        finally:
            os.environ.pop("TRAZAS_REPO", None)

    def test_postgres_no_importa_psycopg_al_construir(self):
        # El visor arranca sin psycopg: se importa al conectar.
        sys_modules = __import__("sys").modules
        self.assertNotIn("psycopg", sys_modules)
        repo = RepositorioPostgres("postgresql://x/y")
        self.assertNotIn("psycopg", sys_modules)
        self.assertEqual(repo.origen(), "postgres")


if __name__ == "__main__":
    unittest.main()
