# 009 — Modal de revisión con cola Anterior / Siguiente

## Objetivo
Que el contador revise clientes uno tras otro sin volver a la tabla, recorriendo **exactamente** la lista que tiene en pantalla (mismo filtro, mismo orden), con el detalle completo del cierre.

## Formas de abrir (todas generan una cola)
| Acción | Cola | Arranca en |
|---|---|---|
| Click en una fila / botón "Revisar" de la fila | todos los IDs del filtro actual (todas las páginas) | ese cliente |
| "Revisar" en un chip de color | IDs del filtro actual **+** ese color | el primero |
| "Revisar seleccionados" | los seleccionados, en el orden de la tabla | el primero |

## Criterios de aceptación
### Backend
1. `GET /api/v1/accountant/periods/{period}/closings/ids` con **los mismos** query params que el listado (sin `page/pageSize`) → `{ "ids": ["clo_..."], "total": N }`. Usa `closing.Apply` (misma función que el listado). Test: para 5 combinaciones de filtro/orden, `ids` == concatenación de los `closingId` de todas las páginas del listado.

### Cola
2. La cola es una **foto** tomada al abrir el modal (IDs congelados). Si un cierre cambia de color o deja de cumplir el filtro, **sigue en la cola**; se muestra con un aviso "Ahora 🟢 Verde" (o el color nuevo).
3. Encabezado del modal: `← Anterior` · **"Cliente 3 de 47"** · descripción del filtro ("Rojos · RI · 'Faltan facturas'") · `Siguiente →` · `✕`. La cola preserva el orden exacto del listado de la tabla (abrir la primera fila muestra "Cliente 1 de N").
4. En el primero, "Anterior" queda `aria-disabled` con tooltip "Es el primer cliente del filtro". En el último, "Siguiente" igual con "Es el último cliente del filtro".
5. Atajos: `←` / `→` navegan, `A` aprueba (lo conecta 010), `Esc` cierra. No se disparan mientras el foco está en un input/textarea.
6. La URL refleja el cliente abierto (`&client=clo_123`) además del filtro. Recargar la página con `client=` reabre el modal en ese cliente reconstruyendo la cola con el filtro de la URL; si el cierre ya no está en la cola, se agrega al principio y se muestra el aviso "Este cliente ya no cumple el filtro".
7. Al navegar, se precarga (`queryClient.prefetchQuery`) el detalle del siguiente y del anterior.
8. Al cerrar, la tabla hace scroll hasta la fila del último cliente visto y la resalta 2 s. Si está en otra página, navega a esa página (cálculo con el índice en la cola y `pageSize`).
9. Clientes ya revisados en esta sesión de cola muestran ✓ en un mini-indicador de progreso ("12 de 47 revisados").

### Contenido del modal (detalle del cierre)
10. Cabecera: semáforo, nombre, CUIT, régimen/categoría, estado, versión.
11. Secciones (tabs o acordeón):
    - **Resumen**: facturado del mes, acumulado 12 m (barra vs tope de la categoría actual y vs K), a pagar y vencimientos; para RI desglose IVA/autónomos/anticipo.
    - **Señales**: todas, con su color, mensaje y datos.
    - **Movimientos**: tabla con tipo, contraparte, monto USD/ARS, TC, clasificación (regla + confianza). Acción "Cambiar a Ingreso / No ingreso" (PATCH de 002); al guardar se refresca el detalle y la cola sigue en el mismo índice.
    - **Facturas y conciliación**: facturas con TC usado y ARS, pares conciliados, cobros sin factura, facturas sin cobro.
    - **Aviso al cliente**: borrador (lo agrega 010) y, si fue notificado, lo publicado (`Published`) con fecha.
    - **Historial**: versiones y aprobaciones.
12. Pie: "Aprobar y siguiente" (primario), "Aprobar" , "Confirmar pase a RI" (solo si hay `MONO_EXCLUSION_K_CAP_EXCEEDED` y fecha en `mono.exclusionDate`; abre confirmación con la fecha de exclusión en formato ISO formateada y envía `YYYY-MM-DD`; usa 005). Los botones de aprobar los conecta 010.
13. Accesibilidad: Radix Dialog (foco atrapado, `aria-labelledby`), navegación completa por teclado.
14. Tests: armado de cola en los 3 modos, límites (primero/último), atajos, URL ↔ modal, reapertura desde URL, cola congelada ante cambio de color, prefetch llamado, scroll/resaltado al cerrar.
