---
title: Carpeta de trabajo, árbol, acceso y usuarios
summary: "S10 funciona como un escritorio: barra de vistas con escenarios y un
  árbol con carpetas (Escritorio, Obras Ganadas, Bandeja, Archivo Central,
  Papelera). Solo se edita en Escritorio. Se entra como sa, con servidor y base
  de datos; sa crea usuarios."
tags:
  - presupuestos
  - carpeta-de-trabajo
  - usuarios
  - acceso
  - tercero
id: 01M3RJP5V71A0VDHD1ARYXHQJZ
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.231Z
---

## Partes de la ventana
- **Barra de títulos**: módulo y presupuesto activo.
- **Barra de menús**: *Archivo*, *Ver*, *Hoja del Presupuesto*, *Catálogos*, *Herramientas*. Los menús cambian según el escenario.
- **Barra de herramientas**: Grabar (solo en Datos Generales), Imprimir, Vista preliminar, Retroceder/Avanzar, Lista y **Notas** (admite textos y archivos de Word, Excel, AutoCAD o MS Project).
- **Barra de vistas**: botones hacia los escenarios (Datos Generales, Hoja del Presupuesto, Diseño Pie de Presupuesto, Gastos Generales, Fórmula Polinómica, Planeamiento, Precios, Transportabilidad, Utilitarios).
- **Barra de estado**: servidor y usuario. Un doble clic en la «carita» inicia sesión como otro usuario o en otra base.

## Carpetas del árbol
- **Escritorio**: presupuestos en uso. Es la única carpeta con todas las opciones de edición.
- **Obras Ganadas**: presupuestos cerrados; no se pueden modificar.
- **Bandeja**: presupuestos que no están en uso; se pueden devolver al Escritorio.
- **Archivo Central**: presupuestos que ya no se usarán.
- **Papelera de Reciclaje**: presupuestos eliminados. Desde aquí se restauran o se borran de forma definitiva.

Se mueven con clic derecho > **Enviar a**.

## Herramientas útiles
*Definir estructura ítem* (niveles, longitud, color), *Configuración* (propia de cada usuario), *Correo interno* y *Limpiar registro de control*, que desbloquea un escenario tras una falla.

## Acceso
Doble clic en el ícono del módulo. La primera vez se entra como **sa** (administrador). En *Detalles* se indica el **Servidor** (en monousuario, el nombre de la PC) y la **Base de datos** (en red, la del servidor).

## Usuarios y grupo de trabajo
En *Utilitarios > Registro de usuarios*, sa crea los usuarios, que antes deben existir en el catálogo de identificadores. Las categorías son Administrador de sistemas, Administrador de módulo, Jefe de grupo, Especialista, Usuario e Invitado (este último solo ve). Los accesos se dan por módulos, escenarios, catálogos, documentos y estaciones.

Un presupuesto solo lo ve quien lo creó y sus superiores. Para compartirlo, clic derecho > **Grupo de trabajo** y se agregan usuarios.

**Confiabilidad:** copia de tercero sin verificar. Puede no corresponder a la versión actual; la captura muestra un servidor «2000-SP3».

## Fuentes
- guia-de-usuario-de-s10-presupuestos.pdf, págs. 6-10, 69-70 y 125-127
- manual-de-s10-costos-y-presupuestos.pdf, págs. 5-12 y 92-93
