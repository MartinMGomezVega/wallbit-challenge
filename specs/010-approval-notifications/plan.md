# 010 — Plan

## Domain
- `domain/notification/`: `Kind`, `Notification`, `OutboundMessage{ID, Channel, ClientID, Subject, Body, Attempts, SentAt?, LastError?}`, `BuildClosingMessage`, `OutboundFor(kind, monthLabel) []OutboundMessage` (EMAIL + PUSH con textos fijos), `ContainsAmounts(s string) bool` (regex `\$|US\$|\d{1,3}(\.\d{3})+(,\d{2})?`) para el test de privacidad.
- `domain/closing`: `Approve(version, hash, by, note, batch, now)` ya existe (007); agregar `ErrVersionConflict`, `ErrAlreadyApproved`.

## Ports
```go
type NotificationRepository interface { Save(ctx, n notification.Notification) error; ByClient(ctx, clientID string) ([]notification.Notification, error); MarkRead(ctx, id string, at time.Time) error; ExistsFor(ctx, clientID, closingID string, kind notification.Kind, key string) (bool, error) }
type OutboxRepository interface { Enqueue(ctx, m notification.OutboundMessage) error; Pending(ctx, limit int) ([]notification.OutboundMessage, error); Update(ctx, m notification.OutboundMessage) error }
type NotificationSender interface { Send(ctx, m notification.OutboundMessage) error }
```
`key` en `ExistsFor` permite deduplicar recordatorios por vencimiento (ej. `"VAT:2026-11-25"`).

## Use cases
- `ApproveClosing`, `ApproveBatch` (worker pool), `NotifyClosing` (in-app + outbox + `MarkNotified`), `DispatchOutbox` (sincrónico después de encolar en el prototipo + también en el scheduler), `CreateActionRequiredNotices` (lo llama `RunMonthlyClosing` al terminar), `SendPaymentReminders`.

## Frontend
- `features/closing-review/useApprove.ts` (mutation, invalida `['closings']`, `['summary']`, `['closing', id]`; en "Aprobar y siguiente" marca revisado y avanza).
- `features/closing-review/sections/ClientMessageSection.tsx`.
- `features/closings-dashboard/BatchApproveDialog.tsx`.
- Atajo `A` en `useKeyboardNav`.
