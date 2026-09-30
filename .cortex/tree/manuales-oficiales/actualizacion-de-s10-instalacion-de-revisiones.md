---
title: Actualización de S10 (Instalación de revisiones)
summary: A fin de que el sistema funcione, se debe procurar mantener el sistema
  actualizado con las revisiones completas. Si el software esta siendo usado por
  múltiples usuarios se recomienda realizarlo en un horario que no estén
  trabajando como bien temprano, antes del inicio de…
tags:
  - oficial-s10
  - manual
  - portal-miembros
  - soporte
id: 01M3SP285CSGW5C4RSPN25A199
status: active
updated_by: ai-agent
updated_at: 2026-09-30T17:47:59.467Z
---

Manual oficial de S10, bajado del portal de miembros con la cuenta de la empresa.

- **Página:** https://documentacion.s10peru.com/actualizacion-s10-instalacion-revisiones/
- **Módulo:** `soporte` (ver esa rama para el resumen redactado)

## Texto de la página

# Actualización de S10 (Instalación de revisiones)

Fuente: https://documentacion.s10peru.com/actualizacion-s10-instalacion-revisiones/

A fin de que el sistema funcione, se debe procurar mantener el sistema actualizado con las revisiones completas. Si el software esta siendo usado por múltiples usuarios se recomienda realizarlo en un horario que no estén trabajando como bien temprano, antes del inicio de labores, en la hora de almuerzo o en la noche.
En caso sea un equipo que ya cuenta con la instalación y se desea actualizar, verificar desde el Panel de Control cual es la última revisión que tienen instalada en el equipo, y según ello instalar siempre la última versión de la revisión más reciente antes de ejecutar un nuevo número.
Insistimos en que las revisiones deben ejecutarse una por una en orden numérico, hasta llegar a la última.
Importante:
Siempre considerar que si se tiene ya algunas revisiones instaladas se instale la última versión de la última en la lista. Esto debido a que existen varias versiones de una misma revisión y se debe tener instalada siempre la que tenga fechas más recientes.
En caso de mostrársele algún error o mensaje indicándole que no se pueden instalar.
Verificar que tenga los permisos  de Administrador verificar también que el antivirus este desactivado y también que el Control de cuentas de usuario esté inactivo. De persistir los problemas, registrar un ticket para un diagnóstico más completo.
Si aparece el siguiente mensaje durante el proceso de instalación, es porque el modulo está en ejecución en el equipo o desde otra sesión del mismo, para continuar deben cerrar el S10 o finalizar el proceso desde el administrador de tareas, para luego dar a la opción de refrescar para continuar.
Considerar también que en caso de tener problemas con una revisión, no puede continuar instalando las siguientes revisiones, ya que darán como resultado errores que conllevarán a una desinstalación.
Actualización de Base de Datos
Luego de instalar las revisiones, es importante permitir que se ejecute correctamente y sin interrupciones la actualización de base de datos, para ello se debe asegurar de ser el primer usuario en conectarse a la base de datos luego de la actualización.
De estar seguro de ser el primer usuario en conectarse y no se mostró esta ventana de actualización de base de datos, es posible que algún otro usuario se conectase a la base de datos desde otro equipo. Es importante este proceso por lo que podrían presentarse errores en el uso si no se ejecutó correctamente, de tener problemas consultar con SoporteS10.
Unos de los mensajes que puede aparecer antes de que actualice la BD es el siguiente, al ver el mensaje la ventana brinda opciones para visualizar, si se elige la opción SI se mostrara una segunda ventana con la lista para identificar que usuario está conectado, para continuar debe coordinar que cierren el S10 y luego elegir el botón continuar para que se realice el proceso de actualización de base de datos.

El texto completo también está troceado en `kb/fragmentos.jsonl` y se cita con el nombre del manual y la página.

**Confiabilidad:** oficial S10
