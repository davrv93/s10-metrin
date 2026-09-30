# Crear un presupuesto de obra desde cero en S10 Presupuestos

**Para qué sirve.** S10 Presupuestos permite elaborar presupuestos de todo tipo de obras a partir de los metrados, con hasta tres presupuestos por obra —Venta, Meta y Línea Base— que luego se asignan a proyectos del módulo Gerencia de Proyectos [N1]. La secuencia oficial de trabajo es: Datos generales, Hoja del presupuesto (partidas, metrados, procesamiento, precios), diseño del pie, gastos generales, fórmula polinómica, planeamiento y reportes [N3]. Este tutorial cubre desde la carpeta de trabajo hasta los reportes.

## Antes de empezar

- Acceso al módulo con usuario y clave. La primera vez se entra como **sa** (administrador); en *Detalles* se indica el **Servidor** (en monousuario, el nombre de la PC) y la **Base de datos** (en red, la del servidor) [N2] [F11].
- Los usuarios se crean en *Utilitarios* [F1] [N2]. Para que otros usuarios vean el presupuesto hay que registrarlos en el **Grupo de trabajo** de ese presupuesto; ese derecho lo tiene el usuario que registra, los jefes de grupo y los administradores del módulo [F7].
- Conviene tener definidos los datos del presupuesto: nombre, propietario, fecha, lugar y plazo [F10] [F11].
- Verificar la configuración (botón de configuración, **Shift+F1**): marcar **Permitir arrastrar y pegar**, **Hacer propio en forma automática partidas principales** y **… subpartidas**, para que los cambios de APU afecten solo a este presupuesto [N7] [F40].

## Pasos

### 1. Ubicar la carpeta de trabajo

1.1. Entre al módulo Presupuestos. El árbol muestra las carpetas **Escritorio**, **Obras Ganadas**, **Bandeja**, **Archivo Central** y **Papelera de Reciclaje** [F1] [N2].
1.2. Trabaje en **Escritorio**: es la única carpeta donde están activas todas las opciones de edición [F1].
1.3. Use clic derecho para ver las opciones del menú [F1].
1.4. Para mover presupuestos entre carpetas, use clic derecho > **Enviar a** [N2] [F13].

### 2. Registrar el presupuesto nuevo

2.1. Vaya al escenario **Datos Generales** [F9] [F11] [N3].
2.2. Clic derecho en el árbol > **Nuevo** [F9] [F13].
2.3. En el **Catálogo de Presupuestos**, elija el grupo o cree uno con **Nuevo SubItem** (por ejemplo OBRAS VIALES y, dentro, PUENTES) [F9] [N3].
2.4. En la ventana **Presupuesto**, registre la **Descripción** (nombre de obra) [F9] [N3].
2.5. Elija el **Cliente** desde el catálogo; si no existe, adiciónelo con clic derecho > *Adicionar* [F16] [N3].
2.6. Registre la **Ubicación geográfica** (distrito) y la **Fecha** del presupuesto. Los precios se guardan por fecha y lugar [N3].
2.7. Registre el **Plazo** en días calendario. En el ejemplo del manual son 180 días, dato del propietario que se usa para el cálculo analítico de gastos generales [F10] [F11].
2.8. Registre la **Jornada diaria** en horas; el sistema asume 8 por defecto e influye en el rendimiento [F36] [N3].
2.9. En **Datos adicionales** defina decimales de precios, incidencias y metrados, la moneda principal y secundaria, el factor de cambio y la casilla **Análisis de precios unitarios tipo 2** (con check calcula para carreteras, sin check para edificaciones) [F53] [F16] [N3].
2.10. **Adicionar** y doble clic para llevarlo al árbol. Luego sobrescriba el nombre del subpresupuesto (por ejemplo ESTRUCTURAS); se añaden más con clic derecho > *Adicionar Subpresupuesto* [N3].

### 3. Definir títulos y subtítulos

3.1. Vaya al escenario **Hoja del Presupuesto** y active el modo **Solo partidas**: es el más rápido para armar la hoja [F32] [N7].
3.2. Clic derecho > **Adicionar Título**. Permite insertar uno o más títulos seleccionados desde el catálogo [F18].
3.3. En el **Catálogo de Títulos**, en la ventana superior ubique los títulos y con doble clic preselecciónelos hacia la ventana inferior. Para quitar uno de la ventana inferior, doble clic sobre él [F21].
3.4. Use el botón que traslada los datos preseleccionados a la hoja del presupuesto [F21].
3.5. Jerarquice con **Desplazar a la derecha** (aumenta la sangría, un nivel más; hasta 10 niveles, sin doble espacio) y **Desplazar a la izquierda** (disminuye la sangría) [F31] [F21].
3.6. Reordene con **Desplazar hacia arriba** y **Desplazar hacia abajo**; no se aplican a selección múltiple [F22] [F23].
3.7. Use **Generar ítems** para numerar títulos y partidas. La estructura depende de la configuración (niveles, dígitos por nivel, tipo de letra, color): clic derecho sobre el nombre del presupuesto en el árbol > *Opciones – Definir ítems*. Si hay error de indentado, el sistema avisa y no genera ítems hasta corregirlo [F22] [F23].

> La columna ITEM no es editable. Para numeración propia use la columna **Alterno** [F22] [F23].

### 4. Insertar las partidas

4.1. Ubique el cursor en el título que recibirá las partidas [F19] [F20].
4.2. Clic derecho > **Adicionar Partida**. La partida adicionada toma el mismo nivel (indentado) del ítem de la fila actual [F18].
4.3. Antes de usar esta opción, verifique la configuración, en particular las opciones para hacer propias las partidas [F18] [F40].
4.4. Elija las partidas en el catálogo que muestra el sistema [F19].
4.5. Registre el **metrado** de cada partida. No documentado en las fuentes disponibles el nombre exacto de la celda de captura; consulte el manual de Presupuestos en el Portal de Ayuda [N5]. Los metrados se escriben en la columna Metrado y sus decimales se definen en *Datos adicionales* [N8] [F53].
4.6. Si necesita una unidad distinta a la del catálogo, use clic derecho > **PROCESOS ESPECIALES > UNIDAD PROPIA**, o digite cualquier carácter en la columna de unidad (U). Esta unidad no se modifica en el catálogo [F27] [F28].
4.7. Si necesita un texto distinto, use **PROCESOS ESPECIALES > DESCRIPCION PROPIA**; si excede 250 caracteres, guárdelo en **NOTAS** [F29].
4.8. Si los metrados están en Excel, use **PROCESOS ESPECIALES > VINCULAR METRADO A CELDA EXCEL** y luego *Tomar metrado de celda Excel* [F28] [F27].
4.9. Vuelva a **Generar ítems** después de reubicar títulos o partidas [F22].

### 5. Asignar o crear el análisis de precios unitarios

5.1. Para ver el APU de una partida, haga doble clic sobre ella, o active el modo **Análisis de precios unitarios**, que lo muestra en la parte inferior de la pantalla [F31] [F32].
5.2. Con el cursor en el APU, clic derecho > **Adicionar recurso** y elija uno o varios del catálogo de recursos [F39].
5.3. Cargue los datos según el tipo: a mano de obra se ingresa la **cuadrilla**, siempre que tenga rendimiento; a materiales se ingresa la **cantidad** para una unidad; en equipos también se registra cuadrilla; y en herramientas manuales, el **porcentaje** en cantidad [F39].
5.4. Use **Adicionar recurso del presupuesto** para tomar recursos ya usados en el mismo presupuesto [F39].
5.5. Use **Adicionar subpartida** para incorporar partidas del catálogo dentro del APU; verifique antes la configuración de hacer propias partidas y subpartidas [F39] [F40].
5.6. Use **Adicionar plantilla** para insertar una agrupación de recursos y subpartidas definida por el usuario [F39].
5.7. Registre el **rendimiento** diario de mano de obra y de equipo. El reporte de APU muestra Rendimiento, MO., EQ. y el costo unitario directo por unidad [F37] [F56] [N4].
5.8. Si la partida vino sin APU (partida formato, como las importadas de Excel), clic derecho > **Reasignar partida del catálogo** y elija con doble clic una partida cuyo APU necesita. El nombre que se imprime es el de la partida de la hoja, y el APU asignado se convierte en propio y se puede personalizar [F30].
5.9. Para localizar pendientes, use el **Filtro** de la hoja: *SIN ANALISIS* muestra las partidas sin APU; *ANALISIS COPIA*, las copiadas de otra partida; *PROTEGIDO*, las que no se pueden modificar [F34] [F24].

### 6. Precios de los recursos

6.1. Procese el presupuesto. En la configuración (pestaña de procesamiento) puede activar **Verifica Metrados**, **Verifica Análisis de Precios Unitarios**, **Verifica Precios Recursos** y **Verifica Cantidades en Análisis de precios unitarios** [F33].
6.2. Cuando falten precios, el sistema pregunta si los ingresará; responda **Sí** e ingrese los precios en la ventana **Recursos y precios**, **sin IGV** [N8] [F33].
6.3. Filtre los recursos por **Tipo** (Mano de obra, Materiales, Equipo, Subcontrato, Comodines), por **Filtro** (Con precio, Sin precios, Vinculados a Excel, Con factor precio, Con factor cantidad, Todos) y por **Moneda** (Principal o Secundaria) [F14].
6.4. Recuerde que el precio de un recurso se guarda por **lugar, fecha y presupuesto** [N8]. Si cambia el lugar o la fecha del presupuesto, los precios pasan a corresponder a la nueva fecha y lugar [F7].
6.5. Vuelva a procesar [N8].

> Cifras de cotización: los precios de los ejemplos del manual corresponden a un presupuesto fechado el **3 de enero de 2005** [F10] [F11]. Verifique con sus cotizaciones vigentes y con los valores legales actuales (UIT, RMV, jornales de construcción civil), que cambian cada año.

### 7. Pie del presupuesto, gastos generales y fórmula polinómica

7.1. Vaya al escenario **Diseño Pie de Presupuesto** y elija el presupuesto en el árbol [F45].
7.2. Registre las líneas con **N° Línea, Descripción, Variable, Macro** y **Omitir Polinómica**. Las variables obligatorias son `nDirecto` (costo directo) y `P_T` (total) [F45] [F44] [N12].
7.3. Pie de ejemplo para venta u oferta: GASTOS GENERALES `GG` = `nDirecto*0.15`; UTILIDAD `UTI` = `nDirecto*0.10`; SUBTOTAL `ST` = `nDirecto+GG+UTI`; IMPUESTO `IGV` = `ST*0.19`; TOTAL `P_T` = `ST+IGV` [F45] [F44]. **El IGV de 19 % es la tasa de las fuentes de 2004-2005; verifique la tasa vigente** [F44] [F46] [N12].
7.4. Para el presupuesto **meta**, el pie de ejemplo lleva solo COSTO DIRECTO y TOTAL PRESUPUESTO [F44].
7.5. Si hará gastos generales analíticos, use la variable **GGP** o **ggp** en la columna Macro; es la condición para acceder a ese escenario [F47] [F44] [F46].
7.6. En **Gastos Generales**, adicione rubros (personal profesional y auxiliar, personal técnico, alquiler de equipo menor, tributos) y, por rubro, los conceptos con unidad, personas, % de participación, tiempo y sueldo o jornal [F48] [N12].
7.7. Marque **Omitir** en la línea del IMPUESTO: para elaborar la fórmula polinómica hay que omitir en el pie el ítem relacionado al impuesto [F49] [F50] [F51] [F46].
7.8. Entre a **Fórmula Polinómica**, elija el presupuesto y luego el subpresupuesto, y confirme «¿Desea elaborar la Fórmula Polinómica?» [F49].
7.9. Haga el **agrupamiento preliminar**: de preferencia no agrupe el índice unificado 47 (mano de obra incluido leyes) y agrupe semejantes hasta que el porcentaje llegue al 5 % para que tenga representatividad [F49].
7.10. Conforme los monomios respetando los requisitos: hasta 8 monomios; coeficiente de incidencia mínimo 0.05; coeficientes al milésimo; suma de coeficientes igual a 1; factor K al milésimo; monomios compuestos de hasta 3 elementos componentes [F50] [F51]. **Estos límites siguen normas peruanas según las fuentes (2004-2005); verifique la norma vigente** [F52] [F57].
7.11. Vuelva a procesar el presupuesto [N12] [N3].

### 8. Emitir los reportes

8.1. Use el botón que envía el reporte a la impresora predeterminada de Windows, o el que lo muestra en pantalla [F52] [F57].
8.2. Desde **Datos Generales** imprima: *Datos generales*; *Resumen del presupuesto estándar* (presupuesto y subpresupuestos con montos globales); *Resumen desagregado* (además mano de obra, materiales, equipo y subcontratos del total) [F52] [F57].
8.3. Desde la **Hoja del Presupuesto**, con el cursor sobre el presupuesto o sobre un subpresupuesto: *Estándar interno* (muestra el código de catálogo de la partida), *Estándar cliente* (el que se entrega al dueño de la obra) y *Desagregado precios unitarios* (formato horizontal con los componentes por tipo de recurso) [F52] [F54] [F57].
8.4. Con el cursor sobre el subpresupuesto, imprima **Análisis de precios unitarios** de todo el subpresupuesto [F54].
8.5. Para **Recursos y precios**, use el botón de la barra del escenario Hoja del Presupuesto; desde un subpresupuesto solo imprime los recursos de ese subpresupuesto [F54].
8.6. Para revisar antes de emitir, use **Procesos especiales > Resumen de partidas**, que lista las partidas y permite filtrarlas: *Partidas utilizadas*, *Partidas propias no utilizadas*, *Partidas propias* y *Partidas con descripciones propias* [F55] [F56].
8.7. Si quiere cabeceras propias, use *Diseño de cabeceras para reportes* y active **Imprimir con diseño de cabecera** en la pestaña *Impresión* de la configuración [F14] [F15].

## Advertencias

- **Fecha y lugar:** si cambia la fecha o el lugar del presupuesto, los precios de los recursos pasan a corresponder a la nueva fecha y lugar [F7]. Decida ambos datos antes de cargar precios.
- **Partidas propias:** sin el check de *Hacer propio en forma automática*, modificar el APU desde el presupuesto también modifica el APU origen en el catálogo [F40].
- **Ítems:** no se pueden generar ítems mientras exista un error de indentado; el sistema envía mensaje, por ejemplo «El Item CONTENEDOR OFICINA no está inde…» [F23] [F22].
- **Partidas en uso:** si una partida está siendo utilizada en un proyecto del módulo Gerencia de Proyectos, no es posible cortarla: «Error -2147217900: La partida 01.01.01 está siendo utilizada en Gerencia de Proyectos» [F24] [F25].
- **Precios sin IGV:** ingrese los precios de los recursos sin IGV para no duplicar el impuesto que el pie ya calcula [N8].
- **Fórmula polinómica:** se elabora para el presupuesto venta u oferta y exige omitir el ítem del impuesto en el pie [F49] [F50] [F51].
- **Cifras legales y tarifas:** el IGV de 19 %, los sueldos y jornales de los ejemplos y los precios de recursos provienen de fuentes de **2004-2005** [F44] [F48] [F10]. Verifique la tasa de IGV, la UIT, la RMV y los jornales de construcción civil vigentes antes de cerrar el presupuesto.
- **Permisos:** la pestaña de procesamiento solo la ve **sa** o quien tenga atributos de supervisor, y rige para toda la empresa [N8] [F33].
- **Respaldo:** antes de cambios masivos, saque copia de seguridad en *Utilitarios > Mantenimiento de base de datos* [F2] [N9].
- Los pasos concretos de instalación y de copia de seguridad están en el Portal de Ayuda, tras inicio de sesión [N5].
## Fuentes

- **[F1]** Manual de S10 Costos y Presupuestos, pág. 9 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F2]** Guia de Usuario de S10 Presupuestos, pág. 122 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F7]** Manual de S10 Costos y Presupuestos, pág. 81 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F9]** Guia de Usuario de S10 Presupuestos, pág. 11 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F10]** Manual de S10 Costos y Presupuestos, pág. 12 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F11]** Guia de Usuario de S10 Presupuestos, pág. 10 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F13]** Manual de S10 Costos y Presupuestos, pág. 78 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F14]** Manual de S10 Costos y Presupuestos, pág. 119 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F15]** Guia de Usuario de S10 Presupuestos, pág. 111 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F16]** Manual de S10 Costos y Presupuestos, pág. 16 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F18]** Manual de S10 Costos y Presupuestos, pág. 97 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F19]** Manual de S10 Costos y Presupuestos, pág. 33 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F20]** Guia de Usuario de S10 Presupuestos, pág. 28 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F21]** Manual de S10 Costos y Presupuestos, pág. 32 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F22]** Manual de S10 Costos y Presupuestos, pág. 43 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F23]** Guia de Usuario de S10 Presupuestos, pág. 35 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F24]** Manual de S10 Costos y Presupuestos, pág. 44 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F25]** Guia de Usuario de S10 Presupuestos, pág. 90 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F27]** Manual de S10 Costos y Presupuestos, pág. 105 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F28]** Guia de Usuario de S10 Presupuestos, pág. 95 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F29]** Manual de S10 Costos y Presupuestos, pág. 104 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F30]** Manual de S10 Costos y Presupuestos, pág. 30 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F31]** Manual de S10 Costos y Presupuestos, pág. 42 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F32]** Guia de Usuario de S10 Presupuestos, pág. 35 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F33]** Manual de S10 Costos y Presupuestos, pág. 86 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F34]** Manual de S10 Costos y Presupuestos, pág. 44 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F36]** Manual de S10 Costos y Presupuestos, pág. 68 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F37]** Guia de Usuario de S10 Presupuestos, pág. 189 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F39]** Manual de S10 Costos y Presupuestos, pág. 108 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F40]** Manual de S10 Costos y Presupuestos, pág. 23 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F44]** Guia de Usuario de S10 Presupuestos, pág. 38 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F45]** Guia de Usuario de S10 Presupuestos, pág. 38 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F46]** Guia de Usuario de S10 Presupuestos, pág. 174 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F47]** Manual de S10 Costos y Presupuestos, pág. 47 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F48]** Guia de Usuario de S10 Presupuestos, pág. 175 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F49]** Guia de Usuario de S10 Presupuestos, pág. 45 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F50]** Guia de Usuario de S10 Presupuestos, pág. 45 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F51]** Manual de S10 Costos y Presupuestos, pág. 53 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F52]** Guia de Usuario de S10 Presupuestos, pág. 51 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F53]** Guia de Usuario de S10 Presupuestos, pág. 13 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F54]** Manual de S10 Costos y Presupuestos, pág. 61 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F55]** Manual de S10 Costos y Presupuestos, pág. 89 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[F56]** Guia de Usuario de S10 Presupuestos, pág. 77 · manual://Guia de Usuario de S10 Presupuestos.pdf
- **[F57]** Manual de S10 Costos y Presupuestos, pág. 60 · manual://Manual de S10 Costos y Presupuestos.pdf
- **[N1]** nodo de Cortex `presupuestos` · S10 Presupuestos (Costos y Presupuestos)
- **[N2]** nodo de Cortex `presupuestos/carpeta-de-trabajo-y-acceso` · Carpeta de trabajo, árbol, acceso y usuarios
- **[N3]** nodo de Cortex `presupuestos/registro-y-secuencia-de-uso` · Secuencia de uso y registro del presupuesto
- **[N4]** nodo de Cortex `presupuestos/analisis-de-precios-unitarios-y-catalogos` · Análisis de precios unitarios y catálogos
- **[N5]** nodo de Cortex `soporte/portal-de-ayuda` · Portal de Ayuda: manuales, prerrequisitos y acceso
- **[N7]** nodo de Cortex `presupuestos/armar-la-hoja-del-presupuesto` · Formas de armar la hoja del presupuesto
- **[N8]** nodo de Cortex `presupuestos/metrados-precios-y-procesamiento` · Metrados, precios y procesamiento
- **[N9]** nodo de Cortex `presupuestos/reportes-y-transportabilidad` · Reportes, exportación y traslado de la base de datos
- **[N12]** nodo de Cortex `presupuestos/pie-gastos-generales-y-formula-polinomica` · Pie de presupuesto, gastos generales y fórmula polinómica

---
_Generado por tutor.py el 2026-09-30 02:33._