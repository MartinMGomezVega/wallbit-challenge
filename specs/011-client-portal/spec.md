# 011 — Vista del cliente

## Objetivo
Lo que ve el usuario de Wallbit: qué tiene que hacer (acción requerida), y una vez aprobado por el contador, cuánto facturó, cuánto paga y hasta cuándo. Con los botones "Emitir factura en ARCA" y "Pagar en ARCA" visibles pero deshabilitados ("Próximamente").

## Criterios de aceptación
### Backend
1. `GET /api/v1/demo/clients` (cualquier rol, solo demo) → `[{id, name, regime}]` con los 15 demos primero; lo usa el `RoleSwitcher`.
2. `GET /api/v1/me` → `{id, name, cuitMasked ("20-****5678-9"), regime, category?}`.
3. Al notificar (010), el cierre guarda `Published{Version, Title, Body, Figures{billedUSD, billedARS, items[{concept, amount, dueDate}], total}, PublishedAt}`. **El cliente solo ve `Published`**, nunca el cálculo vigente sin aprobar (test: reabrir un cierre no cambia lo que ve el cliente hasta la re-aprobación).
4. `GET /api/v1/me/closings` → por período: `{closingId, period, state: "ACTION_REQUIRED"|"IN_REVIEW"|"READY", published?, paidAt?, input{answered, confirmed, invoicesCount}}`.
   - `ACTION_REQUIRED` si el input no está confirmado; `IN_REVIEW` si está confirmado y no hay `Published` (o hay uno viejo y se reabrió: se muestra el viejo con la marca "Actualizando"); `READY` si hay `Published`.
5. `POST /api/v1/me/closings/{id}/paid` → setea `paidAt` (solo si `READY`; si no → 409). Idempotente.
6. `GET /api/v1/me/notifications` (orden desc) y `POST /api/v1/me/notifications/{id}/read`. Un cliente no puede leer ni marcar notificaciones de otro (403).

### Frontend (`/cliente`)
7. Header con campana y contador de no leídas; panel con la lista; al abrirla las marca como leídas.
8. Tarjeta del período actual según `state`:
   - **ACTION_REQUIRED** — checklist de 3 pasos:
     1. "Subí tus facturas de octubre": zona de arrastre para el CSV de Mis Comprobantes, link "Descargar CSV de ejemplo" (`/samples/mis_comprobantes_ok.csv`, armado para Paula Rojas — ver 003), "Cargar una factura a mano" (formulario). Resultado de importación visible (importadas, duplicadas, advertencias, errores por línea).
     2. "¿Cobraste algo fuera de Wallbit en octubre?": Sí (lista de montos + descripción) / No.
     3. "Listo, ya cargué todo" (deshabilitado hasta responder el paso 2).
     - Junto al paso 1: **"Emitir factura en ARCA"** como `ComingSoonButton`.
   - **IN_REVIEW** — "Tu contador está revisando tu cierre de octubre. Te avisamos cuando esté listo." + resumen de lo cargado.
   - **READY** — título y cuerpo del mensaje publicado; tabla de conceptos a pagar con vencimiento; total; botones **"Pagar en ARCA"** (`ComingSoonButton`) y **"Ya pagué"** (habilitado). Con `paidAt` → badge "Pagado el dd/MM".
9. RI: además, "Cargar factura de compra" (formulario) en el paso 1.
10. Historial: lista de períodos anteriores con su resumen publicado.
11. `ComingSoonButton` (en `shared/ui`): `aria-disabled="true"`, estilo atenuado, **enfocable**, tooltip "Próximamente" en hover **y** foco, click/Enter no hace nada. Test que verifica que el tooltip aparece en hover y en foco (motivo: un `<button disabled>` no dispara hover en todos los navegadores).
12. Tests: los tres estados, import con errores parciales, flujo de confirmación, "Ya pagué", notificaciones, ComingSoonButton.
