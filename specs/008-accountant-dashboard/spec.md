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
   1. checkbox de selección (columna fija sticky a la izquierda);
   2. **Semáforo**: círculo de color + ícono (● rojo, ▲ amarillo, ✓ verde) con `aria-label` ("Rojo: requiere acción", "Amarillo: revisar", "Verde: todo ok") (columna fija sticky);
   3. **Cliente**: nombre (botón/link accesible que abre el modal de revisión) + CUIT en gris (columna fija sticky con borde separador);
   4. **Situación**: descripción flexible (máx. 2 líneas, tooltip con el texto completo) + chip de la señal principal + chip `+N` con tooltip que lista las otras señales con su color;
   5. **Régimen**: badge `Monotributo [Categoría]` (ej. `Monotributo D`) o `Resp. Inscripto` con tooltip "Responsable Inscripto";
   6. **Facturado**: USD grande, ARS chico, alineado a la derecha, en una sola línea (`white-space: nowrap`, `tabular-nums`) y con ancho suficiente para que montos grandes entren completos;
   7. **Acum. 12 meses**: barra de progreso con `% del tope K` (verde < 80%, naranja `--color-progress-warning` #EA580C 80–99%, roja ≥ 100%) y el valor ARS en tooltip;
   8. **A pagar**: monto + "vence dd/MM", alineado a la derecha, en una sola línea (`white-space: nowrap`, `tabular-nums`);
   9. **Estado**: badge (`Esperando al cliente`, `Para revisar`, `Aprobado`, `Notificado`) + "Pagado" si `paidAt`;
   10. **Acción**: botón "Revisar" que abre el modal de revisión.
   Layout responsive: en pantallas de menor ancho (ej. 1024 px) se utiliza scroll horizontal dentro de la tarjeta de la tabla, manteniendo fijas (sticky) a la izquierda las columnas de Checkbox, Semáforo y Cliente para no perder de vista a qué cliente corresponden los datos. En anchos cómodos (1280, 1440, 1920 px), las columnas numéricas no se cortan y la columna Situación ocupa el espacio flexible restante.
5. **Orden server-side por cualquier columna de datos**: click en encabezados ordenables (Semáforo y Situación→`severity`, Cliente→`name`, Régimen→`regime` [régimen y categoría], Facturado→`billed`, Acum. 12 meses→`pctKCap`, A pagar→`toPay`, Estado→`status` [orden de flujo: esperando → para revisar → aprobado → notificado]). Dirección por defecto: `severity desc`, `name asc`, `regime asc`, `status asc`, `billed/pctKCap/toPay desc`, con desempate estable por `clientId asc`. Indicador `aria-sort` en el encabezado refleja la dirección real. No son ordenables el checkbox ni la columna Acción.
6. Paginación server-side (50/100/200 por página) con "Mostrando 1–50 de 400".
7. Estados: skeleton de carga, vacío ("No hay clientes con estos filtros" + limpiar), error con "Reintentar".
8. **Todo el estado de la vista vive en la URL**: `period, severity, signal, regime, status, q, sort, dir, page, pageSize` (y `client` en 009). Una única función `filtersToSearchParams` construye los parámetros de consulta compartidos de forma coherente entre el listado de cierres, el resumen y la cola de IDs de revisión. Cambiar un filtro resetea `page=1`.
9. Selección: checkbox por fila y en el encabezado (página actual). Barra flotante: "N seleccionados" + botón "Aprobar verdes (k)" (deshabilitado con tooltip si `k = 0`; la acción la conecta 010) + "Revisar seleccionados" (009).
10. **Apertura del modal de revisión**: abren el modal exclusivamente el botón "Revisar" y el nombre del cliente (como link). El resto de la fila no tiene acción de click ni cursor pointer, permitiendo seleccionar y copiar texto libremente; el hover de la fila se mantiene solo como resaltado visual.

### Parámetros
11. Ruta `/contador/parametros`: formulario con todos los parámetros de `domain-context.md §9`, con ayuda breve por campo, porcentajes mostrados como % (convertir bp), US$ para el mínimo de facturas faltantes.
    - Botón "Guardar" sin ícono, deshabilitado si no hay cambios respecto al último valor guardado (si el usuario modifica y vuelve al valor original, el botón se deshabilita).
    - Al guardar con éxito: toast flotante "Parámetros guardados. Se recalcularon N cierres." usando el `recalculated` del `PUT`. Errores 422 bajo cada campo y toast de error si la petición falla.
    - Navegación protegida ante cambios sin guardar: si hay cambios sin guardar y el usuario navega dentro de la aplicación (`useBlocker`), se abre un modal "¿Salir sin guardar?" con el mensaje "Tenés cambios sin guardar en los parámetros." y tres acciones: "Seguir editando" (cierra modal y permanece en la página), "Salir sin guardar" (descarta y navega) y "Guardar y salir" (guarda y si es exitoso navega, o se queda mostrando errores si falla). El cierre o recarga de la pestaña activa el evento `beforeunload`.

### Calidad
12. Tests (Vitest + Testing Library + MSW): filtros ↔ URL ↔ query enviada; chips alternan severidad; orden por cada columna; paginación; formatos es-AR; aria-labels del semáforo; estados de carga/vacío/error; apertura de modal solo por botón y nombre; números largos sin truncar; formulario de parámetros (Guardar deshabilitado sin cambios, toast, modal de salida con sus 3 caminos y 422).
