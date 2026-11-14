# 003 — Tasks

- [x] **T003.1** `domain/invoice` tipos + `Key` + tests.
- [x] **T003.2** `AmountARS` con las 3 `fx_rule`, TC propio y ARS. Tests.
- [x] **T003.3** Parser `ParseMisComprobantes` + testdata (criterios 5–7). Tests: ok coma, ok punto y coma, filas malas, notas de crédito como warning, encabezados faltantes → error.
- [x] **T003.4** `ValidateManual` + tests (tipo vs régimen, fecha futura, montos).
- [x] **T003.5** `domain/clientinput` + tests.
- [x] **T003.6** `domain/reconciliation` `Reconcile` + `KnownPayers`. Tests del criterio 16.
- [x] **T003.7** Ports + repos en memoria (Invoice, PurchaseInvoice, ClientInput). Tests de concurrencia.
- [x] **T003.8** Use cases + handlers (`/me/invoices`, `/me/invoices/import`, `/me/purchase-invoices`, `/me/periods/{period}/external-income`, `/me/periods/{period}/confirm`, `/accountant/clients/{id}/invoices`). Tests httptest: éxito, 403, 409, 422, idempotencia del import.
- [x] **T003.9** Seed: facturas, facturas de compra e inputs (criterio 18) + facturas de los 15 demos.
- [x] **T003.10** Copiar CSV de ejemplo a `frontend/public/samples/`. `make coverage` ≥ 90%.
