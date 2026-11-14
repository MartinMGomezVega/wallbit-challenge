# 011 — Plan

## Backend
- `domain/closing`: agregar `Published` y `Publish(snapshot, now)`; `NotifyClosing` (010) lo llama antes de crear la notificación.
- Use cases: `GetMe`, `ListMyClosings`, `MarkPaid`, `ListMyNotifications`, `MarkNotificationRead`, `ListDemoClients`.
- Ownership: todos los `/me/**` toman `clientID` del `Identity`; nunca de la URL o el body.
- `cuitMasked`: oculta los 4 primeros dígitos del cuerpo.

## Frontend
```
features/client-portal/
├── index.ts, api.ts, schemas.ts
├── ClientPage.tsx
├── NotificationsBell.tsx
├── PeriodCard.tsx               # switch por state
├── steps/UploadInvoicesStep.tsx, ExternalIncomeStep.tsx, ConfirmStep.tsx
├── forms/ManualInvoiceForm.tsx, PurchaseInvoiceForm.tsx
├── ReadySummary.tsx
├── History.tsx
└── __tests__/
shared/ui/ComingSoonButton.tsx
```
- Upload con `FormData` + `fetch` (no setear `Content-Type` a mano).
- El `RoleSwitcher` (000) pasa a usar `GET /api/v1/demo/clients`.
