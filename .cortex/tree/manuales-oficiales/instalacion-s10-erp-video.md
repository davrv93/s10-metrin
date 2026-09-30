---
title: Manual de Instalación S10 ERP [VIDEO]
summary: Aprende a cómo instalar el S10 ERP en tu dispositivo windows. 1.-
  Consideraciones previas para la instalación. Para poder realizar la
  instalación de S10 en el equipo y poder configurar correctamente el equipo,
  necesitará poseer un usuario con permisos para instalar o desinstalar…
tags:
  - oficial-s10
  - manual
  - portal-miembros
  - soporte
id: 01M3SP28NY40P8H8JDQC5NQ9KX
status: active
updated_by: ai-agent
updated_at: 2026-09-30T17:48:00.086Z
---

Manual oficial de S10, bajado del portal de miembros con la cuenta de la empresa.

- **Página:** https://documentacion.s10peru.com/manual-instalacion-s10-erp/
- **Módulo:** `soporte` (ver esa rama para el resumen redactado)

## Texto de la página

# Manual de Instalación S10 ERP [VIDEO]

Fuente: https://documentacion.s10peru.com/manual-instalacion-s10-erp/

Mira el Videotutorial
Aprende a cómo instalar el S10 ERP en tu dispositivo windows.
1.- Consideraciones previas para la instalación.
Table of Contents
Toggle
1.1.- Permisos de administrador
Para poder realizar la instalación de S10 en el equipo y poder configurar correctamente el equipo, necesitará poseer un usuario con permisos para instalar o desinstalar programas, o en su defecto un usuario Administrador de equipo (Preferible usar usuario Administrador).
Esto puede verificarse desde el “Administrador de equipos” – “Usuarios y grupos locales” – “Administradores” y ver que el usuario de su equipo esté incluido en ésta lista.
Podría tambien habilitar el usuario administrador del equipo para la instalación.
De no contar con los permisos para hacer esto o de pertenecer a un dominio, debe consultar con su administrador de dominio o personal de sistemas para que le otorguen los permisos necesarios.
1.2.- Componente Net. Framework 2.0
Si el equipo a instalar tiene instalada una versión de “Windows” 8 o superior. Se debe verificar que el componente “.NET Framework 3.5 (Incluye .NET 2.0 y 3.0)”
Esto puede verificarse desde el “Panel de Control”, “Desinstalar un programa”
En la opción “Activar y desactivar las características de Windows”
De no estar activada, debe marcarse las casillas como se muestra en la imagen y “Aceptar”, tener en consideración que para poder realizar este cambio, debe usar un usuario con permisos de administrador. (Ver “Permisos de administrador”)
Este componente debe ser activado y descargado directamente con uso de “Windows Update”
Si luego de la descarga e instalación del componente, el equipo le pide reiniciar, debe hacerlo para que los cambios se apliquen, tener en consideración que para poder realizar este cambio, debe usar un usuario con permisos de administrador. (Ver “Permisos de administrador”), de no poder activar este componente, consultar con un técnico.
1.3.- Desactivar Antivirus
Otro punto que debe tener en cuenta, es que debe desactivar su antivirus antes de comenzar con la instalación, de preferencia desactivarlo permanente hasta que se concluya la instalación.
(Consulte con su personal de sistema o proveedor si no puede desactivarlo).
1.4.- Verificación de versiones S10 instaladas
La instalación de S10-ERP, requiere que no exista ninguna instalación previa del software
De existir instalaciones previas debe desinstalarlo.
1.5.- Desactivar control de cuentas de usuario
La instalación de S10, le solicitará que desactive esta característica de Windows (principalmente presente en versiones actuales de Windows) para poder realizarlo descargue el siguiente archivo
clic aquí
, descomprimir el archivo, y dar doble clic para ejecutar, aceptar los mensaje hasta
completar y finalmente reiniciar la PC.
2.- Instalación de S10 – ERP
Para poder comenzar con la instalación, es necesario contar con el DVD de instalación, desde el siguiente link podrá descargar el DVD instalador,
clic aquí
:
Si al ejecutar el instalador del DVD, el sistema le muestra el mensaje siguiente, se deben seguir las indicaciones para instalar el componente “DirectPlay” necesario para la ejecución del instalador S10 y otros aplicativos útiles que se mostrarán más adelante.
Instalado el componente, ya podrá ejecutar el instalador sin problemas,
El siguiente mensaje informativo del instalador, nos indica que debemos desactivar el antivirus en lo que dure el proceso de instalación. (Ver “Desactivar antivirus”).
Si el antivirus ya fue desactivado, cerrar el mensaje informativo para continuar.
En la primera ventana del instalador se muestran 2 casillas a seleccionar para la instalación.
CLIENTE: la casilla cliente instalará S10 en el equipo
SERVIDOR: instará un motor de base de datos (por defecto SQL Server 2012).
Si cuenta con un servidor, solo debe activar la opción Cliente. Caso contrario active Cliente y Servidor
Para este manual, se dejarán activas ambas casillas, a fin de mostrar los pasos completos de instalación.
Si al ejecutar el instalador el sistema le muestra el mensaje siguiente, debe desactivar el control de cuentas de usuario, de otra forma la instalación no podrá completarse.
(Ver “Desactivar control de cuentas de usuario”)
Leemos y aceptamos los términos del contrato.
Determinamos en que parte del equipo se instalará el software, podemos cambiar la ruta de instalación con el botón “Examinar”
El instalador nos indica el nombre de la carpeta que se generará en Inicio.
Luego de esas indicaciones, se procederán a instalar los componentes necesarios para que trabaje S10.
Completada la instalación de los componentes, se comenzará a instalar S10-ERP en la ruta que le especificamos en pasos anteriores.
Al terminar la instalación de S10-ERP, si se desmarcó la casilla “Servidor” la instalación base de S10 culminará en esta ventana.
Si hemos dejado marcada la casilla “Servidor” al cerrar la ventana anterior se mostrará el instalador del motor de base de datos que viene por defecto con el instalador de S10, dar clic en instalar para iniciar la instalación.
3.- Actualización de S10 (Instalación de revisiones)
En el contenido del instalador del S10ERP se encuentra la carpeta Revisiones, debe copiar y pegar dicha carpeta a la ruta C:\S102000 en su PC, ingresar a la carpeta Revisiones y dar doble clic en el archivo Instalador de revisiones S10 para iniciar la instalación.
4.- Instalar utilitarios S10
Para instalar los utilitarios por favor seguir el siguiente procedimiento; en la carpeta del instalador del S10ERP ubique la carpeta “Utilitarios”, ingrese a dicha carpeta y ubique el archivo Instalar utilitarios S10, de clic derecho sobre el archivo y seleccione la opción Ejecutar como administrador, al completar la instalación el sistema le indicará el mensaje “Presione una tecla para continuar…”
5.- Activación Sentinel de licencia S10
Procederemos a activar usando el ejecutable “S10Certificado.exe”, al ejecutar se mostrará la siguiente ventana, en la cual debemos ingresar el N° Keylok, RUC(debe anteponer la letra R) y la clave del producto. Estos datos son proporcionados vía correo electrónico por el área de ventas S10 luego de la compra de la licencia (ventas@s10peru.com)
5.1.- Si cuenta con un servidor y PCs cliente debe realizar el siguiente procedimiento de activación en las PCs cliente.
Para poder activar la licencia de S10 en la PC cliente, se requiere que en la PC cliente este instalado el S10(realizar los puntos 1,2,3 y 4 del presente manual).
– Inicie un explorador de internet (ejemplo; Chrome, Firefox, IE, etc), copiar y pegar la siguiente dirección en la barra de direcciones del explorador
http://localhost:1947
– En la página resultante, en la barra Options, ubique y de clic en la opción Configuraction.
– Seleccionar la pestaña Access to Remote License Managers
En el espacio Remote Licence Search Parameters debemos ingresar el IP del servidor de licencias y pulsar el botón “Submit”
Puede obtener el IP en el equipo que trabajará como servidor de licencias desde: Panel de Control / Redes e Internet / Centro de redes y recursos compartidos.
Con esto, ya estará listo para poder ingresar al sistema usando las licencias que tenga en su servidor de licencias, ingresando sus datos de acceso.

El texto completo también está troceado en `kb/fragmentos.jsonl` y se cita con el nombre del manual y la página.

**Confiabilidad:** oficial S10
