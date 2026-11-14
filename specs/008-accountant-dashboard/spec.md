# 008 — Tablero del contador (frontend)

## Objetivo
La pantalla principal del contador: una tabla formal y moderna donde a simple vista se distinguen los clientes OK, los que hay que revisar y los que tienen problemas, con filtros por color y por situación. Más la pantalla de parámetros.

## Historias
- Como contador, abro el tablero y veo primero los rojos, después los amarillos, después los verdes.
- Como contador, filtro por color, por situación (tope, faltan facturas, recategorización, cobro atípico, …), por régimen y por estado del cierre.
- Como contador, comparto o recargo la URL y vuelvo a ver exactamente el mismo filtro.
- Como contador, edito los parámetros y veo cuántos cierres se recalcularon.

## Criterios de aceptación
### Layout
1. Ruta `/contador`. Encabezado: "Cierre de octubre 2026" (mes en minúscula, sin `capitalize`; período controlado por el `PeriodSelect` global del header en la query `period`, default = último cerrado según `/health.now`). `<title>`: `"Cierres · Wallbit Contabilidad"`.
2. **Chips de resumen** (de `/summary`): `● Rojos 12` `▲ Amarillos 47` `✓ Verdes 341`. Click = alterna ese color en el filtro `severity` (multi). Cada chip tiene un botón secundario **"Revisar"** (lo conecta 009).
3. **Barra de filtros**: Situación (multi-select con las etiquetas del catálogo agrupadas por severidad, con contador de `bySignal`), Régimen (Todos/Monotributo/RI), Estado del cierre (multi), búsqueda por nombre/CUIT (debounce 300 ms), "Limpiar filtros".
4. **Tabla** (TanStack Table, datos server-side). Columnas, en este orden:
   1. checkbox de selección;
   2. **Semáforo**: círculo de color + ícono (● rojo, ▲ amarillo, ✓ verde) con `aria-label` ("Rojo: requiere acción", "Amarillo: revisar", "Verde: todo ok");
   3. **Situación**: descripción (máx. 2 líneas, `title` con el texto completo) + chip de la señal principal + chip `+N` con tooltip que lista las otras señales con su color;
   4. **Cliente**: nombre + CUIT en gris;
   5. **Régimen**: badge `Mono D` / `RI`;
   6. **Facturado**: USD grande, ARS chico;
   7. **Acum. 12 meses**: barra de progreso con `% del tope K` (verde < 80%, amarilla 80–99%, roja ≥ 100%) y el valor ARS en tooltip;
   8. **A pagar**: monto + "vence dd/MM";
   9. **Estado**: badge (`Esperando al cliente`, `Para revisar`, `Aprobado`, `Notificado`) + "Pagado" si `paidAt`;
   10. **Acción**: botón "Revisar".
5. Orden por click en encabezados ordenables (Situación→`severity`, Cliente→`name`, Facturado→`billed`, Acum.→`pctKCap`, A pagar→`toPay`). Dirección por defecto: `severity desc`, `name asc`, `billed/pctKCap/toPay desc`. Indicador `aria-sort` en el encabezado y en el botón refleja la dirección real.
6. Paginación server-side (50/100/200 por página) con "Mostrando 1–50 de 400".
7. Estados: skeleton de carga, vacío ("No hay clientes con estos filtros" + limpiar), error con "Reintentar".
8. **Todo el estado de la vista vive en la URL**: `period, severity, signal, regime, status, q, sort, dir, page, pageSize` (y `client` en 009). Una única función `filtersToSearchParams` construye los parámetros de consulta compartidos de forma coherente entre el listado de cierres, el resumen y la cola de IDs de revisión. Cambiar un filtro resetea `page=1`.
9. Selección: checkbox por fila y en el encabezado (página actual). Barra flotante: "N seleccionados" + botón "Aprobar verdes (k)" (deshabilitado con tooltip si `k = 0`; la acción la conecta 010) + "Revisar seleccionados" (009).
10. Click en una fila (fuera del checkbox) = "Revisar" de esa fila (009).

### Parámetros
11. Ruta `/contador/parametros`: formulario con todos los parámetros de `domain-context.md §9`, con ayuda breve por campo, porcentajes mostrados como % (convertir bp), US$ para el mínimo de facturas faltantes. Guardar → `PUT`; errores 422 bajo cada campo; éxito → toast "Parámetros guardados. Se recalcularon N cierres." (el `PUT` devuelve `{ parameters, recalculated }`; ajustar el handler de 001/007 si hace falta).

### Calidad
12. Tests (Vitest + Testing Library + MSW): filtros ↔ URL ↔ query enviada; chips alternan severidad; orden; paginación; formatos es-AR; aria-labels del semáforo; estados de carga/vacío/error; formulario de parámetros (422 y éxito).
