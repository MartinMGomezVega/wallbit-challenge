# Challenge — Herramienta del contador de Wallbit

> Resumen del enunciado (confidencial: no publicar) y decisiones de producto ya tomadas.
> Este archivo es la fuente de verdad del "qué" y el "por qué". Si algo de una feature contradice este archivo, gana este archivo.

## 1. El problema

- Wallbit tiene ~2.000 usuarios freelancers/contractors argentinos que cobran en USD de clientes del exterior.
- Wallbit quiere lanzar un servicio contable interno. Lo que el usuario valora es **dejar de preocuparse por los impuestos**, no "tener un contador".
- Hoy un contador trabaja cliente por cliente (pide resúmenes y capturas, separa a mano ingresos de no ingresos, controla lo facturado, calcula lo que hay que pagar y avisa) y llega a **~30 clientes por mes**.
- **Objetivo:** construir la herramienta con la que trabaja el contador de Wallbit para que pueda atender **varios cientos de clientes por mes**.
- **Momento clave: cierre de mes.** Cuando arranca un mes nuevo, cada cliente tiene que quedar con (a) lo que facturó el mes anterior, (b) lo que tiene que pagar, y (c) un aviso.

## 2. Contexto y restricciones

| Tema | Detalle |
|---|---|
| Lo que Wallbit TIENE | Todos los movimientos del usuario dentro de la plataforma: cobros, transferencias, conversiones a stablecoins, gastos con tarjeta, con el tipo de cambio de cada operación. |
| Lo que Wallbit NO TIENE | Las facturas que emite el usuario (las hace en ARCA) ni lo que cobra fuera de Wallbit. |
| Regulación | Todo lo que se le diga al usuario sobre sus impuestos tiene que estar **validado por el contador matriculado**. |
| Datos | Los datos de los usuarios **no pueden salir de los sistemas de Wallbit** (no se envían a servicios externos, tampoco a LLMs externos). |
| Equipo | Un solo engineer + el contador. Primera versión en manos de usuarios reales en 6 semanas. |
| Integraciones | No hace falta integrarse con nada externo. ARCA y tipo de cambio se **mockean** detrás de puertos (ver §4). |

## 3. Entregables del challenge

1. Prototipo funcional con datos inventados + instrucciones para correrlo (README).
2. Nota corta de decisiones (≤ 10 líneas, viñetas): qué hace la v1, qué quedó afuera, riesgo más grande, cómo sabríamos si funcionó, una línea sobre uso de IA.

## 4. Decisiones de producto ya tomadas (NO re-discutir en las features)

### 4.1 Principio rector: gestión por excepción
- La herramienta **arma** el cierre; el contador **revisa y firma**.
- El contador dedica tiempo solo a lo que se sale de lo normal. Los casos "verdes" se aprueban en lote.
- **Ningún número impositivo llega al usuario sin aprobación del contador.** Cada aprobación queda auditada (quién, cuándo, versión del cálculo, hash del resultado).

### 4.2 Qué hace la herramienta sola
- Cierra el período el día 1 (job automático) y toma los movimientos del mes anterior.
- Clasifica movimientos con **reglas determinísticas** (no IA externa): ingreso / no ingreso / sin clasificar, con nivel de confianza.
- Pesifica (ver reglas de tipo de cambio en `domain-context.md`).
- Concilia cobros ↔ facturas (cobros sin factura, facturas sin cobro).
- Calcula acumulado móvil de 12 meses, categoría de Monotributo requerida, cuota, y detecta exclusión (pase a Responsable Inscripto).
- Calcula obligaciones de Responsable Inscripto (IVA, autónomos, anticipo de Ganancias parametrizado).
- Genera señales de riesgo y un **semáforo** por cliente (🔴/🟡/🟢) con una descripción breve.
- Genera el borrador del aviso al usuario.

### 4.3 Qué sigue haciendo el contador
- Define **parámetros** (umbrales, criterio de tipo de cambio, si la factura E computa para el tope, etc.).
- Revisa rojos y amarillos en un **modal de revisión con Anterior / Siguiente** que respeta el filtro activo.
- Corrige clasificaciones (cada corrección recalcula el cierre).
- Aprueba: individual (cualquier color) o **en lote (solo verdes)**.
- Decide temas de criterio (recategorización, pase a RI).

### 4.4 Cómo se entera el usuario
- **Acción requerida** (no es asesoramiento, puede salir sin aprobación): subir facturas (CSV de "Mis Comprobantes" o carga manual), declarar cobros fuera de Wallbit, confirmar que terminó.
- **Cierre listo** (solo después de aprobación): cuánto facturó, cuánto tiene que pagar y hasta cuándo, y si tiene que hacer algo.
- **Recordatorio** antes del vencimiento si no marcó "Ya pagué".
- Email/push **sin montos** ("Tenés novedades"); el detalle solo dentro de Wallbit.
- Botones **"Emitir factura en ARCA"** y **"Pagar en ARCA"** visibles pero deshabilitados (`aria-disabled`) con tooltip **"Próximamente"**. El README explica la integración futura.

### 4.5 Alcance de regímenes
- **Monotributo** (actividad: prestación de servicios) — incluido.
- **Responsable Inscripto** — incluido, incluyendo la **transición** Monotributo → RI por exceder el tope de la categoría K.
- **Fuera de alcance:** Ingresos Brutos, DDJJ anual de Ganancias/Bienes Personales, venta de bienes, emisión real de facturas, pagos reales, integración real con ARCA/BCRA.

### 4.6 Decisiones técnicas
- Backend en **Go** (Clean Architecture). Frontend en **React + TypeScript**. Ver `architecture.md`.
- Hay backend real (no mocks en el front) porque: el cierre corre solo (job), la aprobación necesita auditoría del lado servidor, los datos no deben bajarse completos al navegador y el aviso sale del servidor.
- Lo único mockeado son los **sistemas externos** (ARCA: escalas/vencimientos; tipo de cambio), detrás de puertos, con **datos reales** embebidos.
- Sin TDD. Tests unitarios con **coverage ≥ 90%** (backend y frontend).

## 5. Supuestos explícitos (van al README)

1. El período a cerrar es el **mes calendario anterior** en zona `America/Argentina/Buenos_Aires`.
2. Fecha "hoy" del prototipo configurable (`APP_NOW`, default `2026-11-05T09:00:00-03:00`) → se cierra **octubre 2026**.
3. Todos los clientes Monotributo son de **prestación de servicios**.
4. El cálculo de categoría usa **lo facturado** (facturas), no lo cobrado. Lo cobrado se usa para conciliación y controles de depósitos.
5. Si la factura E computa o no para el tope es un **parámetro del contador** (default: computa = criterio conservador). Las fuentes no coinciden.
6. Tipo de cambio para pesificar facturas: el de la propia factura (columna "Tipo Cambio" del CSV); si no viene, el del proveedor de tipo de cambio según el parámetro `fx_rule`.
7. Anticipos de Ganancias y aportes de autónomos de RI los carga el contador por cliente (dependen de la DDJJ anterior, que Wallbit no tiene).
8. No hay autenticación real: un selector de rol de demo (contador / cliente X).
9. Persistencia en memoria, datos semilla determinísticos al arrancar (~300 clientes).
