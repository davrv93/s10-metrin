# Términos candidatos sin definición en las fuentes

Búsqueda hecha el 2026-10-06 sobre `kb/fragmentos.jsonl` (oficial, tercero-sin-verificar,
académico y, como último recurso, OCR de imágenes) y `kb/fragmentos_faq.jsonl`. Patrones
usados: «<término> es/son», «se denomina», «se entiende por», «consiste en», «permite» y
encabezados «## x.y <Término>». Ninguno de estos términos tiene `.yml` propio.

| Término | Qué se buscó / qué se encontró |
|---|---|
| Insumo | Sin definición propia. El Manual de Nóminas dice «Los recursos en S10 son insumos…»; queda como sinónimo en `recurso.yml`. |
| Nota de ingreso | El término no aparece. S10 habla de «ingreso» y «guía de ingreso»; ver `ingreso_de_almacen.yml`. |
| Planilla / nómina (genérico) | Sin definición general. Solo aparecen «planilla de pago» (lista de ejemplos), «planillón» (sin definir), el tipo de nómina y la planilla electrónica, que sí tienen `.yml`. |
| Centro de costo | Solo pasos de configuración («Definir estructura», importar desde Excel) y su asignación contable; no hay definición. |
| Asiento contable | Solo el escenario «Asientos manuales» (reclasificaciones, depreciación, correcciones que van al Libro Diario) y texto OCR; no se define el concepto. |
| Comprobante (de pago) | Solo aparece dentro de la definición de factura electrónica y como nombre (p. ej., «Comprobantes de retención»), sin definirse. |
| Detracción | Solo el catálogo de porcentajes de SUNAT, el escenario Detracciones (informes) y el procedimiento de pago con constancias. No hay definición. |
| Factor de rendimiento | Solo «Aplica factor de rendimiento a la partida elegida» (opción de la hoja del presupuesto); se menciona en `rendimiento.yml`. |
| Cronograma (genérico) | Solo como nombre de escenarios: «Cronograma por periodos» (tiene `.yml`), «Cronograma de pagos» y «Cronograma de cobranzas». |
| Parte diario | Solo «partes diarios» del equipo; está cubierto en `parte_de_equipo.yml`. |
| Requerimiento | Aparece como equivalente de pedido («el pedido o requerimiento»); queda como sinónimo en `pedido.yml`. |
| Proveedor | Sin definición propia; forma parte del catálogo de socios de negocio. El «Portal de Proveedores» sí se define, pero es otro concepto. |
| Conformidad | Solo «no conformidad» en Calidad (sin definir) y «dar conformidad» al cerrar un ticket de soporte. |
| Nota de crédito | Solo pasos para crearla en Facturación y para usarla en una orden de pago; no hay definición. |
| Gratificación | Solo el sílabo («Cálculo de gratificación») y nombres de conceptos en capturas OCR. |
| Lookahead | Solo «publicar un proyecto desde el LOOKAHEAD, para calidad tipo TAREA PROTOCOLO», sin explicar qué es. |
| Last planner / último planificador | Solo «el usuario que maneja Gerencia de proyectos equivale al Último planificador del PMI»; no se define el método. |
| Restricción (Lean) | La palabra aparece en otros sentidos (pagos, temario de CTS); no hay definición Lean. |

## Con definición en las fuentes, pero sin `.yml` por el tope de 50 términos

Línea base, costo directo, subpartida, stock, recosteo, destino específico, frente, centro de
compra, carta fianza, fondo de garantía, pago a cuenta, curva S, documentos por pagar,
compensación, esquema contable, cuentas genéricas, periodo contable, categoría (de mano de
obra), PLAME, reparto de utilidades, clasificación ABC, unidad operativa, incremento de obra,
favorito, ítem alterno, Calidad móvil, Tareo móvil y Portal de Proveedores. Varios ya se
mencionan dentro de la definición de otro término.

## Incluidos con descripción de uso, sin definición formal en las fuentes

Estos cinco términos SÍ tienen `.yml`, pero las fuentes no traen una frase del tipo «X es…»: la `definicion` describe
solo lo que dicen los manuales oficiales sobre su lugar y uso en S10, y cada archivo lo advierte en `notas`. Son los
primeros que debe revisar una persona (o retirar, si se exige definición formal): `metrado`, `partida`,
`subpresupuesto`, `titulo` y `escenario`.

## Decisión del usuario: mantener con aviso

Fecha: **06-10-2026**. Estado: **decidido: opción 1, mantener con aviso** para los seis términos. Los YAML siguen
cargados y V2 los usa para responder «¿qué es X?». Para que el aviso cumpla su función, V2 lo muestra en la respuesta
(ver «Dónde se ve el aviso»). `revisado_por_humano` sigue en `false` hasta que una persona los lea contra el ERP.

Los cinco términos de la sección anterior tienen `.yml`, pero las fuentes no traen una definición formal («X es…»):
su `definicion` está **armada por uso** a partir de lo que dicen los manuales oficiales (dónde se registra, qué
contiene, para qué sirve). Se suma **CTS**: su definición formal venía de una página de marketing (el temario de una
capacitación publicada en s10peru.com, no un manual) y se reescribió «por uso» solo con fuentes oficiales, así que
queda en el mismo caso. Todos llevan `revision: {generado: 2026-10-06, por: claude, revisado_por_humano: false}`.

| Término (`.yml`) | Qué dice hoy el YAML (`definicion`, definida por uso) | Aviso que lleva (`notas`) | Casos del benchmark que dependen de él |
|---|---|---|---|
| Metrado (`metrado.yml`) | «Dato que se ingresa para cada partida en la celda de metrados de la hoja del presupuesto. El módulo de Presupuestos elabora presupuestos de obra a partir de los metrados, y el presupuesto entrega al proyecto las partidas, metrados, recursos, cantidades y precios.» | «Las fuentes no traen una definición formal de metrado; se describe solo por su uso en S10.» | `conc-001` (`eval/v2_oro.jsonl`) |
| Partida (`partida.yml`) | «Registro de la hoja del presupuesto que se ubica bajo un título y lleva su metrado y su análisis de precios unitarios, que es su estructura de costo. Las partidas se registran en el catálogo de partidas, en forma directa o desde la hoja del presupuesto, antes de elegirlas para un presupuesto.» | «Las fuentes no traen una definición formal de partida; se describe por su lugar en la hoja del presupuesto.» | ninguno como `concepto_esperado` |
| Subpresupuesto (`subpresupuesto.yml`) | «Componente de un presupuesto, como Estructuras, Gastos generales o Instalaciones eléctricas y sanitarias. Al registrar un presupuesto, el sistema crea automáticamente un subpresupuesto con el mismo nombre, que el usuario debe editar y cambiar.» | «Las fuentes no traen una definición formal de subpresupuesto; se describe por su uso en el módulo de Presupuestos.» | `conc-027` |
| Título (`titulo.yml`) | «Registro de la hoja del presupuesto bajo el cual se ubican las partidas que le corresponden. Los títulos se toman del catálogo de títulos, cuyos registros se usan en la hoja del presupuesto.» | «Las fuentes no traen una definición formal de título; se describe por su uso en la hoja del presupuesto.» | ninguno |
| Escenario (`escenario.yml`) | «Cada ventana de trabajo de un módulo de S10. Las ventanas se asemejan a un escritorio y el entorno de trabajo es casi el mismo en todos los escenarios: botones comunes (imprimir, vista preliminar, retroceder, avanzar, ver árbol), botones exclusivos de cada escenario y menús cuyas acciones cambian según el escenario.» | «Los manuales no dan una definición formal de escenario; se describe a partir de la carpeta de trabajo y de la barra de vistas.» | ninguno |
| CTS (`cts.yml`) | Reescrita el 06-10-2026: «Compensación por tiempo de servicios. En los manuales oficiales de S10 aparece como uno de los conceptos que se eligen al preparar una boleta de liquidación en el módulo de Nóminas, junto a los beneficios sociales, las gratificaciones y los reintegros de vacaciones.» Fuentes: Manual de Nóminas › 7.1.6.1 Preparación de Boleta, Nóminas › 1. Generalidades y Manual Administrativo › 7.4 Cuentas Bancarias del Socio Negocio. La definición anterior («Beneficio social derivado de una obligación legal…») salía del temario de capacitación de s10peru.com (`38071a17d2a2-0000`), fuente de marketing que el quality gate rechaza | «Definido por uso: las fuentes oficiales no traen una definición formal de CTS («X es…»); se describe solo por lo que dicen los manuales de Nóminas y Administrativo. La definición anterior […] se retiró el 06-10-2026.» | `conc-009` |

**Dónde se ve el aviso.** Solo en el YAML: V2 lo carga (`conocimiento.Concepto.Notas`), pero el plan
(`tipos.ConceptoDef`) y las plantillas de `metrin/plantillas/respuestas.yml` no lo llevan, así que la persona que
pregunta **no lo ve** en la respuesta (comprobado en el código del 06-10-2026).

**Opciones para cada término** (se puede decidir término por término):

1. **Mantener con aviso.** El YAML se queda como está y V2 sigue contestando «¿qué es X?» con la descripción por uso
   y sus citas. Si se elige esta opción para que el aviso cumpla su función, hay que **mostrarlo** en la respuesta
   (añadir `notas` al plan y una línea en la plantilla de CONCEPT, p. ej. «Los manuales no definen este término; lo
   describo por su uso en S10»), y marcar `revisado_por_humano: true` cuando una persona lo haya leído.
2. **Retirar hasta tener definición oficial.** Se mueve el `.yml` fuera de `kb/conceptos/` (p. ej. a
   `kb/conceptos/_retirados/`: `CargarConceptos` solo lee `kb/conceptos/*.yml`, sin subcarpetas) y el término vuelve
   a la tabla de arriba. Consecuencia medible: «¿qué es un metrado?» (`conc-001`) y «¿qué es un subpresupuesto?» (`conc-027`) dejarían de tener concepto
   y V2 respondería con fragmentos de la búsqueda o con SIN_EVIDENCIA; habría que regenerar `eval/v2_oro.jsonl`
   (`eval/construir_v2_oro.py`) para que esas expectativas no cuenten como fallo. Volvería cuando haya una definición
   oficial citada (manual de S10 o glosario del fabricante).

| Término | Decisión | Fecha | Quién |
|---|---|---|---|
| Metrado | mantener con aviso | 06-10-2026 | usuario |
| Partida | mantener con aviso | 06-10-2026 | usuario |
| Subpresupuesto | mantener con aviso | 06-10-2026 | usuario |
| Título | mantener con aviso | 06-10-2026 | usuario |
| Escenario | mantener con aviso | 06-10-2026 | usuario |
| CTS | mantener con aviso | 06-10-2026 | usuario |

En CTS, la opción 2 dejaría «¿qué es la CTS?» (`conc-009`) sin concepto; volver a la definición del temario no es
una opción, porque es una fuente de marketing.

Los términos de la primera tabla (insumo, nota de ingreso, centro de costo, detracción…) no tienen `.yml`: no hay nada
que retirar, y siguen fuera del glosario hasta que aparezca una definición en las fuentes.
