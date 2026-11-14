# 010 — Tasks

- [x] **T010.1** `domain/notification` (tipos, `BuildClosingMessage` MONO/RI/agregados/nota, `OutboundFor`, `ContainsAmounts`) + tests con strings exactos.
- [x] **T010.2** Repos en memoria (Notification, Outbox) + `notify/log` sender + tests.
- [x] **T010.3** `ApproveClosing` + handler (criterios 1–2). Tests de cada error y del warning.
- [x] **T010.4** `NotifyClosing` + `DispatchOutbox` con reintentos. Tests (sender fake que falla 2 veces y después anda; falla 4 veces → queda con `LastError`).
- [x] **T010.5** `ApproveBatch` + handler (criterio 3). Tests de lote mixto con `-race`.
- [x] **T010.6** `draftMessage` en el detalle (criterio 5).
- [x] **T010.7** `CreateActionRequiredNotices` conectado al job (criterio 9) y `CLOSING_UPDATED` (criterio 10). Tests.
- [x] **T010.8** `SendPaymentReminders` + scheduler + `POST /reminders/run`. Tests de bordes (justo en el umbral, vencido, ya pagado, ya recordado).
- [x] **T010.9** Test de privacidad: ningún `OutboundMessage` generado en el e2e contiene montos (criterio 7).
- [x] **T010.10** Front: botones de aprobar en el modal, nota para el cliente, sección "Aviso al cliente", atajo `A`, manejo de 409. Tests.
- [x] **T010.11** Front: `BatchApproveDialog` desde la barra de selección. Tests.
- [x] **T010.12** Coverage ≥ 90% en ambos lados.
