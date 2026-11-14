# 016 — Tasks

- [x] **T016.1** Dominio: `invoice.UploadedBy` y `ClientPeriodInput.ConfirmedBy/ConfirmationNote/ConfirmBy`. Tests.
- [x] **T016.2** `ImportInvoicesCSV` recibe `UploadedBy`; el handler de `/me` pasa `CLIENT`. Seed con `SEED`. Tests existentes verdes.
- [x] **T016.3** Caso de uso `ConfirmInputOnBehalf` (transaccional) + tests (nota obligatoria, ALREADY_CONFIRMED, transición a READY_FOR_REVIEW, reapertura si estaba aprobado).
- [x] **T016.4** Handlers `POST /accountant/clients/{id}/invoices/import` y `POST /accountant/clients/{id}/periods/{period}/confirm-input` + tests httptest (200, 403 con rol cliente, 404, 409, 422).
- [x] **T016.5** Postgres: migración `0004`, repos y seed loader. Suites de contrato en memoria y Postgres.
- [x] **T016.6** Detalle del cierre expone `uploadedBy`, `confirmedBy` y `confirmationNote`. OpenAPI y test de contrato de rutas.
- [x] **T016.7** Fixtures reales de las respuestas nuevas + validación Zod en el front.
- [x] **T016.8** Front: `ImportInvoicesDialog` (archivo, ejemplo, resultado, aviso si ya está aprobado) + refresco manteniendo el índice de la cola. Tests.
- [x] **T016.9** Front: badge "Cargada por el contador" y `ConfirmOnBehalfForm` con su línea en Historial. Tests y `vitest-axe`.
- [x] **T016.10** E2E backend con el flujo de Paula Rojas (criterio 13). Prueba manual completa en Docker. Coverage ≥ 90%.
