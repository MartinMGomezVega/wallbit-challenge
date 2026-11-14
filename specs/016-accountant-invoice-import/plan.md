# 016 — Plan

## Punto de partida (código actual)
- `usecase.ImportInvoicesCSV.Execute(ctx, clientID, r)` ya recibe el `clientID` como parámetro: el endpoint del contador lo reutiliza sin duplicar lógica. Se agrega un campo `UploadedBy invoice.UploadedBy` en el request (o un parámetro), y el handler de `/me` pasa `CLIENT`.
- `invoice.Invoice` tiene `Source` (`CSV|MANUAL|SEED`): **no se toca**. `UploadedBy` es un campo nuevo e independiente (origen = cómo se cargó; uploadedBy = quién la cargó).
- `clientinput.ClientPeriodInput` tiene `ConfirmedAt`. Se agregan `ConfirmedBy` (`CLIENT|ACCOUNTANT`) y `ConfirmationNote`.
- `usecase.SetExternalIncome` y `usecase.ConfirmPeriodInput` existen. El nuevo caso de uso `ConfirmInputOnBehalf` los compone dentro de una sola transacción (`TxManager`).
- Migraciones Postgres hasta `0003`. Esta feature agrega `0004_uploaded_by_confirmed_by.sql`.
- Front: `features/closing-review/sections/InvoicesSection.tsx` existe. El sample CSV ya está en `frontend/public/samples/mis_comprobantes_ok.csv`.

## Backend
- `domain/invoice`: `type UploadedBy string` (`CLIENT`, `ACCOUNTANT`, `SEED`) + campo `UploadedBy` en `Invoice`. El seed usa `SEED`.
- `domain/clientinput`: `ConfirmBy(by, note, now)` con validación de la nota cuando `by == ACCOUNTANT`.
- `usecase.ConfirmInputOnBehalf{Inputs, Closings, Clock, Notifier, Tx}`.
- `interfaces/http`:
  - `POST /api/v1/accountant/clients/{id}/invoices/import` (reusar el helper multipart de `invoices_handler.go`);
  - `POST /api/v1/accountant/clients/{id}/periods/{period}/confirm-input`.
- Postgres: columnas `invoices.uploaded_by TEXT NOT NULL DEFAULT 'CLIENT'`, `client_period_inputs.confirmed_by TEXT`, `client_period_inputs.confirmation_note TEXT`. Actualizar repos y seed loader.
- Fixtures para el front: respuestas reales de los dos endpoints nuevos.

## Frontend
```
features/closing-review/
├── import/ImportInvoicesDialog.tsx
├── import/ImportResultView.tsx
├── import/useImportInvoices.ts        # mutation multipart (FormData, sin setear Content-Type)
├── confirm/ConfirmOnBehalfForm.tsx
├── api.ts                              # importInvoicesAsAccountant, confirmInputOnBehalf
└── schemas.ts                          # ImportResultSchema, uploadedBy, confirmedBy
```
- Invalidaciones al terminar: `['closing', id]`, `['closings']`, `['summary']`, `['stats']`.
