# 011 — API del cliente

## Objetivo
Exponer la API `/api/v1/me/**` que consumiría la app de Wallbit para que el usuario vea qué tiene que hacer (acción requerida) y, una vez aprobado por el contador, cuánto facturó, cuánto paga y hasta cuándo.

## Criterios de aceptación
### Backend
1. `GET /api/v1/me` → `{id, name, cuitMasked ("20-****5678-9"), regime, category?}`.
2. Al notificar (010), el cierre guarda `Published{Version, Title, Body, Figures{billedUSD, billedARS, items[{concept, amount, dueDate}], total}, PublishedAt}`. **El cliente solo ve `Published`**, nunca el cálculo vigente sin aprobar (test: reabrir un cierre no cambia lo que ve el cliente hasta la re-aprobación).
3. `GET /api/v1/me/closings` → por período: `{closingId, period, state: "ACTION_REQUIRED"|"IN_REVIEW"|"READY", published?, paidAt?, input{answered, confirmed, invoicesCount}}`.
   - `ACTION_REQUIRED` si el input no está confirmado; `IN_REVIEW` si está confirmado y no hay `Published` (o hay uno viejo y se reabrió: se muestra el viejo con la marca "Actualizando"); `READY` si hay `Published`.
4. `POST /api/v1/me/closings/{id}/paid` → setea `paidAt` (solo si `READY`; si no → 409). Idempotente.
5. `GET /api/v1/me/notifications` (orden desc) y `POST /api/v1/me/notifications/{id}/read`. Un cliente no puede leer ni marcar notificaciones de otro (403).
6. Tests: los tres estados, `Published` inmutable al reabrir, "Ya pagué" (idempotente y 409), notificaciones (orden y 403 entre clientes).
