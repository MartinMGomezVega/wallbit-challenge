# 010 — Aprobación, auditoría y avisos

## Objetivo
Implementar la "firma" del contador (individual y en lote, auditada) y todo lo que el usuario recibe: acción requerida, cierre listo, recordatorio de pago. Respetando que **ningún número impositivo sale sin aprobación** y que **email/push no llevan montos**.

## Criterios de aceptación
### Aprobación individual
1. `POST /api/v1/accountant/closings/{id}/approve` body `{ "version": 3, "clientNote": "opcional, máx 500" }`:
   - `version` ≠ versión actual → `409 VERSION_CONFLICT` (concurrencia optimista: el cierre cambió mientras el contador lo miraba);
   - estados permitidos: `READY_FOR_REVIEW`, `PENDING_CLIENT_INPUT` (en este último, la respuesta incluye `warning: "CLIENT_INPUT_NOT_CONFIRMED"`);
   - registra `Approval{version, resultHash, approvedBy:"accountant", approvedAt, clientNote, batch:false}`;
   - genera la notificación `CLOSING_READY` + mensajes salientes (email/push) y deja el cierre `NOTIFIED`;
   - responde el cierre actualizado.
2. Aprobar un cierre ya `APPROVED`/`NOTIFIED` en la misma versión → `409 ALREADY_APPROVED`.

### Aprobación en lote
3. `POST /api/v1/accountant/periods/{period}/closings/approve-batch` body `{ "items": [{ "id": "clo_1", "version": 2 }] }` (máx. 500):
   - solo aprueba cierres **GREEN** y **READY_FOR_REVIEW** con versión coincidente;
   - resultado por ítem: `{ id, ok, error?: "NOT_GREEN"|"NOT_READY"|"VERSION_CONFLICT"|"NOT_FOUND" }`;
   - `Approval.batch = true`. Las notificaciones se generan igual que en la individual.
   - Respuesta `{ approved: n, failed: m, results: [...] }`. Procesa en paralelo con el worker pool; test con `-race`.

### Borrador del aviso
4. `notification.BuildClosingMessage(closing, client) Message{Title, Body}` es-AR, puro:
   - Título: "Tu cierre de octubre 2026 está listo".
   - Monotributo: "Facturaste US$ {usd} ({ars}). Cuota de Monotributo categoría {cat}: {fee}, vence el {dd/MM/yyyy}."
   - RI: "Facturaste US$ {usd} ({ars}). IVA a pagar: {vat} (vence {date}). Aportes de autónomos: {aut} (vence {date}). {anticipo si hay}. Total: {total}." Si `vatToPay = 0` y hay saldo a favor: "No tenés IVA a pagar; te queda un saldo a favor de {credit}."
   - Agregados por señal (solo estos códigos, con texto para el usuario, no para el contador):
     - `MONO_EXCLUSION_K_CAP_EXCEEDED`: "Superaste el tope del Monotributo. Tu contador te va a contactar para el pase a Responsable Inscripto."
     - `MONO_CATEGORY_CHANGE_REQUIRED`: "En la próxima recategorización tenés que pasar a la categoría {required}."
     - `MONO_CATEGORY_DOWNGRADE_POSSIBLE`: "Buenas noticias: en la próxima recategorización podrías bajar a la categoría {required}."
     - `MISSING_INVOICES`: "Tenés {count} cobros sin factura. Emitilas y subilas para que quede todo en orden."
   - Si hay `clientNote`, se agrega al final: "Nota de tu contador: …".
5. `GET /api/v1/accountant/closings/{id}` incluye `draftMessage` (lo que se enviaría ahora).

### Notificaciones
6. `Notification{ID, ClientID, ClosingID?, Kind, Title, Body, CreatedAt, ReadAt?}`, `Kind ∈ ACTION_REQUIRED | CLOSING_READY | CLOSING_UPDATED | PAYMENT_REMINDER`.
7. **In-app** lleva el detalle (montos). **Salientes** (`OutboundMessage{Channel: EMAIL|PUSH, ClientID, Subject, Body}`) **nunca** llevan montos ni CUIT: textos fijos ("Tenés novedades sobre tus impuestos en Wallbit", "Tu cierre de {mes} está listo. Entrá a Wallbit para verlo."). Test que falla si un body saliente contiene `$`, `US$` o dígitos de montos.
8. Puerto `NotificationSender.Send(ctx, OutboundMessage) error`; adaptador `notify/log` que loguea `channel`, `clientId` y `subject` (no el body). Envío con **outbox**: se guarda el mensaje, un dispatcher lo envía y marca `SentAt`; reintenta hasta 3 veces con backoff simple si `Send` falla.
9. `ACTION_REQUIRED`: al correr el job, por cada cierre `PENDING_CLIENT_INPUT` sin aviso previo del período → "Subí tus facturas de {mes} y contanos si cobraste algo fuera de Wallbit." (no es asesoramiento → no requiere aprobación).
10. `CLOSING_UPDATED`: si un cierre `NOTIFIED` se reabre por cambio de datos, **no** se avisa nada hasta la nueva aprobación; al re-aprobar, el aviso es `CLOSING_UPDATED` ("Actualizamos tu cierre de {mes}").
11. `PAYMENT_REMINDER`: proceso `SendPaymentReminders(now)` — para cierres `NOTIFIED` sin `paidAt` con algún vencimiento en `[now, now + reminder_days_before_due]` y sin recordatorio previo para ese vencimiento → notificación + saliente. Lo corre el scheduler diario y `POST /api/v1/accountant/reminders/run` (demo).

### Frontend (contador)
12. Modal: "Aprobar y siguiente" (primario), "Aprobar", atajo `A`; campo opcional "Nota para el cliente"; sección "Aviso al cliente" con el `draftMessage` en vivo y, si el cierre fue notificado, lo publicado (`Published`) con versión, fecha y total. `409 VERSION_CONFLICT` → banner "El cierre cambió mientras lo revisabas" + botón "Ver cambios" (refresca).
13. Tablero: "Aprobar verdes (k)" en la barra de selección → diálogo de confirmación "Vas a aprobar y notificar k clientes verdes" → resultado con fallidos listados.
14. Tests: backend (todos los códigos de error, lote parcial, outbox con reintentos, sin montos en salientes, recordatorios en bordes de fecha) y frontend (botones, atajo, conflicto, lote).
