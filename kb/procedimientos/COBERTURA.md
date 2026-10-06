# Cobertura de las plantillas de procedimiento

Estado al 06-10-2026. Generado por Claude a partir de los manuales; **ninguna plantilla está revisada todavía por una
persona contra el ERP** (`revisado_por_humano: false` en todas).

## Resumen

- **60 plantillas activas** (las del índice) con **645 pasos** contando subpasos. **380 pasos (58,9 %) llevan foto**, 400 fotos
  en total.
- **10 plantillas más en `_reserva/`**, también validadas. Quedaron fuera para respetar el tope de 60 (ver más abajo).
- Validador (`python3 herramientas/validar_procedimientos.py`): **0 errores, 0 dudosos léxicos y 13 fotos dudosas**,
  revisadas a mano una por una. Con `--reserva`: 70 plantillas, 0 errores.
- Fuentes: 455 citas de manuales oficiales (`documentacion.s10peru.com` y el OCR de sus capturas) y 19 de los dos PDF
  «Importado a mano» (`tercero-sin-verificar`). Estas últimas solo están en Presupuestos y en todos los casos acompañan
  a la sección oficial equivalente: aportan el nombre del botón o menú que la sección oficial no menciona.
- Ninguna plantilla usa páginas de marketing ni fragmentos secundarios (FAQ de Cortex, pantallas, YouTube).

| Módulo | Activas | En reserva | Mínimo pedido |
|---|---|---|---|
| Presupuestos | 12 | 4 | 10 |
| Gerencia de Proyectos | 9 | — | 8 |
| Almacenes | 7 | — | |
| Compras | 6 | 3 | |
| Nóminas | 6 | — | |
| Contabilidad | 6 | — | |
| Facturación | 4 | 1 | |
| Administrativo | 4 | — | |
| Facturación Electrónica | 4 | — | |
| Instalación | 2 | — | |
| Tareo Móvil | — | 2 | |
| Portal del Proveedor, Calidad Móvil | — | — | |

## Tareas cubiertas por módulo

### Presupuestos (12)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `presupuestos.calcular-gastos-generales` | Calcular en forma analítica los gastos generales | intermedio | 8 | 5 |
| `presupuestos.copiar-partidas-de-otro-presupuesto` | Copiar partidas de un presupuesto elaborado a otro | basico | 5 | 3 |
| `presupuestos.disenar-pie-de-presupuesto` | Diseñar el pie del presupuesto | intermedio | 11 | 7 |
| `presupuestos.elaborar-formula-polinomica` | Elaborar la fórmula polinómica del presupuesto | avanzado | 13 | 11 |
| `presupuestos.importar-presupuesto-desde-excel` | Importar al S10 un presupuesto elaborado en Excel | intermedio | 9 | 5 |
| `presupuestos.imprimir-reportes` | Imprimir y exportar los reportes del presupuesto | basico | 8 | 8 |
| `presupuestos.ingresar-metrados` | Ingresar los metrados en la hoja del presupuesto | basico | 7 | 1 |
| `presupuestos.procesar-presupuesto` | Procesar el presupuesto e ingresar los precios faltantes | intermedio | 7 | 4 |
| `presupuestos.registrar-partida-en-el-catalogo` | Registrar una partida nueva en el catálogo de partidas | intermedio | 14 | 7 |
| `presupuestos.registrar-presupuesto-nuevo` | Registrar un presupuesto nuevo | basico | 19 | 10 |
| `presupuestos.registrar-titulos-y-partidas` | Registrar títulos y partidas en la hoja del presupuesto | basico | 10 | 8 |
| `presupuestos.registrar-usuarios` | Registrar usuarios del módulo de Presupuestos | intermedio | 9 | 7 |

### Gerencia de Proyectos (9)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `gerencia-proyectos.asignar-presupuestos-al-proyecto` | Asignar presupuestos al proyecto | basico | 6 | 3 |
| `gerencia-proyectos.configurar-calendario-del-proyecto` | Configurar el calendario del proyecto | intermedio | 6 | 5 |
| `gerencia-proyectos.definir-periodos` | Definir la escala y los periodos visibles del proyecto | basico | 6 | 5 |
| `gerencia-proyectos.planificar-en-cronograma-por-periodos` | Planificar el proyecto en el Cronograma por Periodos | intermedio | 10 | 10 |
| `gerencia-proyectos.registrar-fases-wbs-y-asignar-partidas` | Registrar las fases del WBS y asignarles partidas | intermedio | 16 | 10 |
| `gerencia-proyectos.registrar-pedido-manual-de-compra` | Registrar un pedido manual para comprar | basico | 7 | 6 |
| `gerencia-proyectos.registrar-proyecto` | Registrar un proyecto nuevo | basico | 16 | 3 |
| `gerencia-proyectos.registrar-subcontrato` | Registrar un subcontrato | intermedio | 19 | 8 |
| `gerencia-proyectos.registrar-usuarios-del-proyecto` | Registrar usuarios del proyecto y asignarles atributos | basico | 6 | 4 |

### Almacenes (7)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `almacenes.asignar-stock-minimo` | Asignar stock mínimo a los recursos | basico | 5 | 4 |
| `almacenes.consultar-kardex` | Consultar el kardex de un recurso e imprimir el reporte para la SUNAT | basico | 9 | 6 |
| `almacenes.registrar-egreso-a-partida-de-control` | Registrar un egreso a partida de control | intermedio | 17 | 8 |
| `almacenes.registrar-ingreso-por-compra` | Registrar un ingreso por compra con documento de pago | intermedio | 11 | 8 |
| `almacenes.registrar-ingreso-por-inventario` | Registrar un ingreso por inventario | basico | 18 | 10 |
| `almacenes.registrar-ingreso-por-orden-de-compra` | Registrar un ingreso por orden de compra | basico | 8 | 4 |
| `almacenes.regularizar-ingreso-por-compra` | Regularizar un ingreso por compra con la factura | intermedio | 16 | 4 |

### Compras (6)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `compras.aprobar-orden-de-compra` | Enviar a aprobación y aprobar una orden de compra | basico | 6 | 6 |
| `compras.aprobar-pedido` | Aprobar un pedido de compra | basico | 6 | 5 |
| `compras.generar-orden-de-compra-desde-pedido` | Generar una orden de compra desde un pedido de compra | intermedio | 16 | 5 |
| `compras.registrar-aprobadores-de-orden-de-compra` | Registrar los aprobadores de las órdenes de compra | avanzado | 10 | 6 |
| `compras.registrar-centro-de-compra` | Registrar un centro de compra | intermedio | 10 | 8 |
| `compras.registrar-pedido-de-compra` | Registrar un pedido manual para comprar | basico | 12 | 6 |

### Nóminas (6)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `nominas.calcular-la-nomina` | Calcular la nómina de un periodo | avanzado | 17 | 10 |
| `nominas.importar-socios-de-negocio-desde-excel` | Importar socios de negocio desde una hoja de Excel | intermedio | 7 | 3 |
| `nominas.imprimir-boletas-de-pago` | Preparar e imprimir las boletas de pago | intermedio | 20 | 10 |
| `nominas.registrar-tareo-estandar` | Registrar el tareo estándar con ayuda de Excel | intermedio | 15 | 8 |
| `nominas.registrar-trabajador-como-socio-de-negocio` | Registrar un trabajador como socio de negocio | basico | 36 | 7 |
| `nominas.seleccionar-personal-disponible` | Seleccionar el personal disponible y copiarlo a los proyectos destino | intermedio | 9 | 7 |

### Contabilidad (6)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `contabilidad.cierre-permanente-del-periodo` | Hacer el cierre permanente de un periodo contable | intermedio | 5 | 3 |
| `contabilidad.configurar-informe-contable` | Configurar un informe contable (estado financiero) | avanzado | 18 | 10 |
| `contabilidad.crear-plan-de-cuentas` | Crear e importar el plan de cuentas | avanzado | 20 | 9 |
| `contabilidad.emitir-estados-financieros` | Visualizar e imprimir los estados financieros | basico | 5 | 3 |
| `contabilidad.generar-libros-electronicos` | Generar un libro electrónico para el PLE | intermedio | 6 | 4 |
| `contabilidad.registrar-asiento-manual` | Registrar un asiento manual | basico | 6 | 1 |

### Facturación (4)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `facturacion.definir-usuarios-y-accesos` | Definir usuarios y sus accesos en Facturación | intermedio | 8 | 8 |
| `facturacion.registrar-cliente` | Registrar un cliente | basico | 9 | 8 |
| `facturacion.registrar-entrega-a-rendir` | Registrar un comprobante de entrega a rendir | intermedio | 9 | 7 |
| `facturacion.registrar-nota-de-credito` | Registrar una nota de crédito (o de débito) de una factura | intermedio | 9 | 5 |

### Administrativo (4)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `administrativo.definir-aprobadores-de-pagos` | Definir aprobadores, moneda y monto para la aprobación de pagos | intermedio | 10 | 8 |
| `administrativo.registrar-cobranza-en-efectivo` | Registrar una cobranza en efectivo en el diario de caja | basico | 8 | 6 |
| `administrativo.registrar-estado-bancario-manual` | Registrar el estado bancario manual y conciliarlo | avanzado | 11 | 10 |
| `administrativo.registrar-pago-con-cheque` | Registrar una orden de pago y pagarla con cheque | intermedio | 14 | 10 |

### Facturación Electrónica (4)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `facturacion-electronica.comunicar-baja-de-comprobante` | Dar de baja un comprobante electrónico con la comunicación de baja | intermedio | 5 | 2 |
| `facturacion-electronica.configurar-emision-electronica` | Configurar la emisión electrónica (empresa, tipos de documento, talonarios y clientes) | avanzado | 10 | 6 |
| `facturacion-electronica.crear-usuario-del-portal` | Crear un usuario en el portal de facturación electrónica | basico | 8 | 3 |
| `facturacion-electronica.enviar-comprobantes-a-sunat` | Enviar comprobantes electrónicos a SUNAT | basico | 8 | 6 |

### Instalación (2)

| id | tarea | nivel | pasos | con foto |
|---|---|---|---|---|
| `instalacion.instalar-s10-erp` | Instalar el S10 ERP en un equipo | intermedio | 10 | 10 |
| `instalacion.sacar-copia-de-seguridad-de-la-base-de-datos` | Sacar una copia de seguridad de la base de datos del S10 | basico | 6 | 6 |

## En reserva (`kb/procedimientos/_reserva/`)

Las 10 están redactadas y validadas; no entran a `INDICE.json` ni a la validación por defecto. Para activar una,
muévala a `kb/procedimientos/<modulo>/` (y, si quiere, añádala a los `relacionados` de las otras) y vuelva a validar.
Se apartaron por ser menos frecuentes, por ser delgadas o por estar fuera de los módulos pedidos:

| id | Motivo de apartarla |
|---|---|
| `presupuestos.usar-favoritos` | 4 pasos; depende de favoritos ya registrados. |
| `presupuestos.reasignar-recurso` | Tarea correctiva, poco frecuente. |
| `presupuestos.registrar-recurso-en-el-catalogo` | Casi todo sale del OCR de las capturas; la sección oficial tiene 3 frases. |
| `presupuestos.exportar-presupuesto` | Avanzada, poco frecuente. |
| `compras.registrar-cotizacion`, `compras.registrar-recepcion-de-cotizaciones`, `compras.elaborar-cuadro-comparativo` | El circuito de cotizaciones completo (9, 9 y 6 pasos). Conviene activarlo junto. |
| `facturacion.importar-saldos-por-cobrar` | Tarea de puesta en marcha, se hace una sola vez. |
| `tareo-movil.publicar-parte-de-tareo`, `tareo-movil.sincronizar-y-cerrar-el-tareo` | Módulo fuera de la lista prioritaria. |

## Tareas NO cubiertas

«Tiene fuente» significa que hay material suficiente para una segunda tanda. Si no, el motivo es la falta de fuente.

### Presupuestos
| Tarea | Motivo |
|---|---|
| Configurar el módulo (2.2 Configuración) | Las 9 partes describen opciones de la ventana de configuración (llamadas), no un procedimiento. Lo esencial (opciones de procesamiento, partidas propias) va como prerrequisito o paso en `procesar-presupuesto` y en `registrar-titulos-y-partidas`. |
| Buscar en los catálogos (3.1 búsqueda global/local) | Tiene fuente (s029–s030). No es una tarea de registro; queda para una segunda tanda. |
| Precios por presupuesto, por grupos, generales e históricos (5.1.5, 5.2) | Tienen fuente (s109–s122). Pendientes para una segunda tanda. |
| Diseño de cabeceras para reportes (5.1.6) | Tiene fuente (s113). Pendiente. |
| Importar presupuesto, importar datos Windows 1.x/2.x y DOS 7.x/8.x (5.3.3–5.3.5) | Solo texto descriptivo, sin pasos ni capturas asociadas. |
| Actualización de base de datos (5.5.2) | 180 caracteres sin pasos. |
| Planeamiento desde Presupuestos (cap. 6: proyecto, calendario, WBS, MS Project, avances) | Es el mismo contenido que el manual de Gerencia de Proyectos; se cubre allí para no duplicar. |
| Registrar el subpresupuesto (3.4) | Una sola frase; queda como último paso de `registrar-presupuesto-nuevo`. |

### Gerencia de Proyectos
| Tarea | Motivo |
|---|---|
| Cerrar el proyecto (10.15, s310) | Dos frases; habla de concluir la calificación de calidad en el móvil, no del cierre del proyecto. Sin fuente suficiente. |
| Generar pedido automático (3.7.2, s113–s117) | Tiene fuente (~7 pasos con capturas). Pendiente para una segunda tanda. |
| Aprobar pedidos (3.7.3–3.7.4, s119–s120) | Tiene fuente. Pendiente (Compras ya tiene `compras.aprobar-pedido`). |
| Registrar avances en la rama meta (8.1, s254) | Tiene fuente (~6 pasos). Pendiente. |
| Ver resultados operativos (9.2–9.3, s256–s258) | Tiene fuente (~7 pasos). Pendiente. |
| Valorizar un subcontrato (4.6.10, s159–s162) | Tiene fuente. Pendiente. |
| Exportar al MS Project y actualizar desde él (2.6.2, s085–s088) | Tiene fuente. Se priorizó la planificación manual en el cronograma. |
| Datos adicionales del proyecto (2.5.2, s047–s048) | Tiene fuente; no estaba entre las prioridades. |

### Almacenes
| Tarea | Motivo |
|---|---|
| Ingreso con guía de remisión y descarga directa (s044–s045, EPCDD) | Casi todo está en capturas y el formulario solo se describe por llamadas. Lo esencial quedó en `regularizar-ingreso-por-compra` y en el aviso de **Descarga directa** de `registrar-ingreso-por-inventario`. |
| Saldos y stock reservado, stock valorizado, ubicación física (3.2, 3.3, 3.5) | De 2 a 4 pasos cada una, casi sin texto accionable. El recosteo (3.1) va dentro de `consultar-kardex`. |
| Ingresos IP, IDP, IDS, IF y egresos EP, ETE, ETI, EDOC, EV, EAC, EBM (s048–s051, s065–s079) | Texto corto, sin pasos ni capturas. |

### Compras
| Tarea | Motivo |
|---|---|
| Imprimir y enviar la orden de compra (12.3–12.4, s051–s052) | Unos 4 pasos, y la fuente no aclara qué llamada es el envío por correo y cuál la exportación a Excel. |
| Orden de servicio (15, s057–s062), pedido A Almacén (14, s055–s056), orden de compra desde su propio escenario (12.1, s049) | Tienen fuente. Pendientes. |
| Cotizaciones, recepción y cuadro comparativo (9–11) | Redactadas: están en `_reserva/`. |

### Nóminas
| Tarea | Motivo |
|---|---|
| Ficha del trabajador (5.2.3, s047–s051) | Tiene fuente abundante. Pendiente para una segunda tanda. |
| PLAME, T-Registro, AFP NET (8.x), feriados (4.8.1.1), registro de usuarios (2.1.1) | No evaluadas en esta tanda (objetivo del módulo ya cubierto). |

### Contabilidad
| Tarea | Motivo |
|---|---|
| Crear empresas y sucursales (3.1, s016) | Tres frases cortas más capturas: ~4 pasos poco claros. |
| Registrar el tipo de cambio (2.3, s006) | Una frase: se actualiza solo desde SUNAT. No hay pasos. |
| Periodos contables (5.2, s053) y datos generales (5.1, s052) | No dan para una plantilla propia: van dentro de `cierre-permanente-del-periodo` y `crear-plan-de-cuentas`. |

### Facturación
| Tarea | Motivo |
|---|---|
| Crear empresas y sucursales (3.1, s014) | Tres frases; el resto solo en capturas. Sin fuente suficiente. |
| Letras únicas de cambio (5.6.3, s035), fondo rotatorio (6.6, s054–s058), detracciones (7.3, s063) | Tienen fuente. Pendientes. |

### Administrativo
| Tarea | Motivo |
|---|---|
| Cobranza por depósito bancario (3.5, s014) | Tres pasos y el tipo de pago solo aparece deformado en el OCR («Depasita»). |
| Orden de pago en efectivo (5.1, s026) | Cinco pasos muy cortos; el registro, envío a aprobación y aprobación de la orden ya están en `registrar-pago-con-cheque`. |
| Talonarios de cheques (2.7, s009) y libros caja (3.1, s010) | Cuatro y dos pasos. Libro Caja y Diario de Caja quedan como prerrequisitos de la cobranza. |

### Facturación Electrónica
| Tarea | Motivo |
|---|---|
| Recuperar la contraseña del portal (3.1.3, s015) | Tiene fuente. Pendiente. |
| Configurar talonarios (2.1.3) | Integrado en `configurar-emision-electronica`. |

### Instalación
| Tarea | Motivo |
|---|---|
| Activación Sentinel de la licencia (5, 5.1) | Tiene fuente (s011–s012). Pendiente. |
| Instalar utilitarios S10 (4) | Un solo paso. |
| Instalación de Microsoft SQL Server (documento aparte) | Fuera de la prioridad de esta tanda. |

### Tareo Móvil (las dos plantillas están en `_reserva/`)
| Tarea | Motivo |
|---|---|
| Instalar y activar la app (1.2, «Valida tus datos», «Genera tu código…») | Pasos de una línea cada uno, casi todo en capturas. |
| Resumen semanal (2.4) | Es una consulta; va como verificación de `sincronizar-y-cerrar-el-tareo`. |

### Portal del Proveedor y Calidad Móvil
| Módulo | Motivo |
|---|---|
| Portal del Proveedor | Secciones de 1–3 pasos (acceso, recuperar contraseña, consultas). Fuente escasa; candidato: «recuperar contraseña» (s004). |
| Calidad Móvil | Tiene fuente (asignar protocolos s002, publicar actividades s003, sincronizar s012–s013), pero no estaba en la prioridad pedida. Candidatos para una segunda tanda. |

### Tutoriales
Los tutoriales de `tutoriales/` (el índice lista 5) no se convirtieron a YAML: se usaron como referencia de tono.
`presupuestos.registrar-presupuesto-nuevo`, `registrar-titulos-y-partidas`, `ingresar-metrados`, `procesar-presupuesto`,
`disenar-pie-de-presupuesto`, `calcular-gastos-generales`, `elaborar-formula-polinomica` e `imprimir-reportes` cubren
por partes el tutorial «Crear un presupuesto de obra desde cero», pero citando la sección oficial en vez del PDF.

## Elementos dudosos

### Soporte léxico

**Ninguno.** Durante la redacción, el validador marcó unos 15 elementos que parafraseaban demasiado su fuente. Se
reescribieron más cerca del texto del manual; ninguno quedó dudoso.

### Fotos dudosas (13, todas revisadas: son correctas)

Se marcan porque el paso y el texto de la foto (su leyenda más el OCR) comparten menos de 2 palabras de contenido.
En las 13 la causa es la misma: un OCR deformado, vacío o reducido a un icono o botón. Ninguna es una foto equivocada.

| plantilla | lugar | foto | raíces en común | qué muestra |
|---|---|---|---|---|
| `almacenes.consultar-kardex` | paso 5 | `6e30b4a547c3.png` | 0/2 | Icono del botón de reporte; OCR vacío. |
| `compras.registrar-aprobadores-de-orden-de-compra` | paso 9 | `8ccb3fec2a05.png` | 1/8 | Catálogo de usuarios («Catélogo» en el OCR). |
| `compras.registrar-centro-de-compra` | paso 7 | `413ba14d09df.png` | 1/2 | Catálogo de usuarios («Catdlogo» en el OCR). |
| `compras.registrar-pedido-de-compra` | paso 2 | `bf9f3b4bcadc.png` | 0/3 | Icono del botón de configuración; OCR vacío. |
| `gerencia-proyectos.planificar-en-cronograma-por-periodos` | paso 6 | `607948ac1a75.png` | 1/7 | Ventana con las opciones «Repetir» y «Repartir». |
| `gerencia-proyectos.planificar-en-cronograma-por-periodos` | paso 10 | `989083841cdc.png` | 1/8 | Ventana «Recalcular»; el OCR solo trae esa palabra. |
| `gerencia-proyectos.registrar-pedido-manual-de-compra` | paso 5 | `273637016f0b.png` | 0/8 | Ventana de cantidades del recurso («Sako», «Steck» en el OCR). |
| `gerencia-proyectos.registrar-proyecto` | paso 4 | `c7de5d040d6e.png` | 1/7 | Menú del catálogo con «Nuevo Subltem». |
| `gerencia-proyectos.registrar-subcontrato` | paso 2 | `97245ff02063.png` | 1/6 | Menú de clic derecho «Adicionar subcontrato», deformado en el OCR. |
| `nominas.imprimir-boletas-de-pago` | paso 3 | `679d83292a1c.png` | 1/4 | Menú de clic derecho con «Adicionar». |
| `presupuestos.copiar-partidas-de-otro-presupuesto` | paso 5 | `328411277798.png` | 1/2 | «Use este botón para generar los ítems», deformado en el OCR. |
| `presupuestos.registrar-partida-en-el-catalogo` | paso 14 | `01e503c33740.png` | 1/12 | Lista de recursos con sus cantidades en el análisis. |
| `presupuestos.registrar-titulos-y-partidas` | paso 4 | `7eeb5fd28530.png` | 1/2 | «Catdlogo de Titulos… Elija los títulos uno a uno con doble clic». |
| `presupuestos.registrar-usuarios` | paso 2 | `3b937f97f14d.png` | 1/5 | «Haga clic derecho… Mantenimiento de Usuarios… Adicionar». |

### Pasos que conviene comprobar en el ERP antes de marcarlos como revisados

Todos tienen fuente, pero la fuente es solo una captura o un dato que puede haber cambiado:

- `contabilidad.registrar-asiento-manual`: la sección oficial tiene una sola frase. Los pasos salen del OCR de una
  captura que lleva la nota del propio manual «Clic Derecho, Opcion Adicionar» y muestra los campos Fecha, Moneda y
  Cabecera de información.
- `contabilidad.cierre-permanente-del-periodo`: el paso «cambie el Estado a Cerrado» sale del OCR de la ventana
  Periodos Contables.
- `compras.registrar-centro-de-compra`: la pestaña **Proyectos** solo se nombra en una captura de la misma ventana
  que trae la sección de aprobadores.
- `presupuestos.registrar-presupuesto-nuevo`, paso 4: abrir la ventana Presupuesto con **Nuevo SubItem** desde el
  último nivel del grupo. Sale de la Guía de Usuario en PDF (pág. 12); la sección oficial solo dice «el último nivel
  de registro en el árbol es la hojita».
- `facturacion-electronica.enviar-comprobantes-a-sunat`: el plazo de **7 días** para enviar la factura es el que da
  el manual. Es un dato normativo que puede haber cambiado; compruébelo antes de mostrarlo como vigente.
- `presupuestos.elaborar-formula-polinomica`: los requisitos (8 monomios, mínimo 0,05) son los del manual para el
  Perú. Son normas que también conviene comprobar.

## Problemas de los datos encontrados al redactar

- **Ids repetidos en `kb/fragmentos.jsonl`.** El id de las secciones web trunca el slug del manual a 8 caracteres
  (`web-https-documentacion-s10peru-com-manual-d-s032`): **200 ids se repiten y afectan a 1.061 fragmentos** de casi
  todos los manuales. El par `(id, manual)` sí es único, y por eso cada plantilla trae la tabla `fuentes` (ver
  ESQUEMA.md). Si Metrín indexa los fragmentos solo por `id`, hoy pierde o mezcla secciones de manuales distintos.
- **Fotos corridas un paso** en varias secciones del manual de Nóminas (p. ej. s041: la leyenda «pestaña cuentas de
  Banco» acompaña a una captura de datos de ejemplo). En esos pasos la foto se asoció por su OCR, no por la leyenda.
- **Facturación Electrónica** tiene dos secciones numeradas 2.2.3 (s010 y s011). s010 repite casi entera s009, y en s012
  hay pegado un párrafo de la comunicación de baja.
- **Erratas de la fuente**:
  - Nóminas s042: «Llamada h: xxx», «Llamada i: xxx».
  - Nóminas s041, llamada d: dice «dirección fiscal» donde la captura muestra Dirección Postal.
  - Contabilidad s002: dice «módulo de facturación».
  - Administrativo: la misma opción se llama «Documento Actual» en s027 y «Pago Actual» en s028.
- **OCR deformado** en nombres clave («Andlisis», «Catdlogo», «Subltem», «Némina», «Depasita»). Esos nombres no se pusieron
  en negrita salvo que otra fuente del mismo paso los trajera bien escritos.
