---
title: Instalación y activación de S10 Presupuestos
summary: "Activar .NET Framework 3.5 y 4.8, bajar el Control de cuentas de
  usuario, desactivar el antivirus, ejecutar autorun > Instalación ERP (cliente
  y servidor) y el motor SQL. Después: revisiones .bat, driver de licencia, S10
  Certificado y actualizar la base al primer ingreso."
tags:
  - presupuestos
  - instalacion
  - licencia
  - oficial
id: 01M3RJP5W4HWCPEQBJW4SXMN2W
status: active
updated_by: ai-agent
updated_at: 2026-09-30T07:15:28.260Z
---

## Requisitos previos (Windows)
1. *Panel de control > Programas > Activar o desactivar las características de Windows*: marcar **.NET Framework 3.5** y **4.8**.
2. *Panel de control > Sistema y seguridad > Cambiar configuración de Control de cuentas de usuario*: bajar el deslizador al mínimo y aceptar.
3. Desactivar el **antivirus**; el instalador lo advierte.

## Instalar
1. Abrir la carpeta de instalación, ya descargada y descomprimida, y ejecutar el **autorun**.
2. Elegir **Instalación ERP** con las casillas **Cliente** y **Servidor** marcadas > *Instalar ERP* > Siguiente. La ruta por defecto es el disco C:.
3. Al terminar, *Finalizar* abre el instalador del **motor de SQL** (la base de datos) > *Instalar* > Aceptar.

## Revisiones
Descomprimir las revisiones recibidas y ejecutar el **.bat**. En el video, «184, 186» instala de la 184 a la 186 en orden. Pulsar una tecla al final.

## Licencia
1. En la carpeta del instalador, subcarpeta **certificado**, instalar el **driver de la licencia**.
2. Ir a `C:\S10-2000`, ordenar por tipo y abrir **S10 Certificado**.
3. Con los códigos que envía el área de ventas, llenar:
   - **Número de key lock**: el certificado, p. ej. «C2764».
   - El código con **R** y el número de su empresa.
   - La **Clave de producto**.
4. **Activar**. Debe aparecer el mensaje de certificado activado.

## Primer ingreso
*Inicio > S10ERP > S10 Presupuestos* abre el login. Al aceptar, arranca la **actualización de la base de datos**, que tarda unos minutos. Al terminar, *Cerrar*, y ya se puede usar la hoja del presupuesto.

## Notas
- La transcripción es automática: «autorran», «instalación RP», «loin» y «secar» se interpretaron como autorun, Instalación ERP, login y aceptar. La carpeta del instalador se oye como «Masterlight», un nombre dudoso. Los números de revisión (184-186) son los del video, no una regla.
- La guía de tercero asume que, cuando se instala S10, viene una base de consulta con partidas de ejemplo.

**Confiabilidad:** oficial S10 (video del canal de YouTube de S10, transcripción automática). La nota sobre la base de ejemplo viene de la copia de tercero.

## Fuentes
- Video hvqEKVrCXpA, «Aprende a cómo instalar S10 Presupuestos» (8 min 39 s)
- guia-de-usuario-de-s10-presupuestos.pdf, pág. 29
