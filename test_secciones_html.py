import unittest

from secciones_html import secciones_texto


class SeccionesTextoTest(unittest.TestCase):
    def test_fallback_separa_jerarquia_y_conserva_orden(self):
        texto = """# Manual de ejemplo
Fuente: https://documentacion.s10peru.com/manual-ejemplo/
Table of Contents
Toggle
1 GENERALIDADES
El módulo permite administrar el proyecto.
1.1 Registro del proyecto
Ingrese al escenario de Datos Generales.
Seleccione Adicionar para crear el registro.
1.2 Revisión
Verifique los datos del proyecto.
2 PLANIFICACIÓN
Defina las actividades del proyecto.
"""
        secciones = secciones_texto(texto)

        self.assertEqual(
            [s["titulo"] for s in secciones],
            [
                "1 GENERALIDADES",
                "1 GENERALIDADES › 1.1 Registro del proyecto",
                "1 GENERALIDADES › 1.2 Revisión",
                "2 PLANIFICACIÓN",
            ],
        )
        self.assertEqual(secciones[1]["pasos"][0]["texto"], "Ingrese al escenario de Datos Generales.")
        self.assertEqual(secciones[1]["pasos"][1]["texto"], "Seleccione Adicionar para crear el registro.")
        self.assertTrue(all(not paso["fotos"] for s in secciones for paso in s["pasos"]))

    def test_no_crea_secciones_desde_listas_de_pasos_sin_encabezado(self):
        texto = """# Artículo
1.- Ingrese al módulo.
2.- Seleccione una opción.
"""
        self.assertEqual(secciones_texto(texto), [])


if __name__ == "__main__":
    unittest.main()
