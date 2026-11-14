# 011 — Plan

## Backend
- `domain/closing`: agregar `Published` y `Publish(snapshot, now)`; `NotifyClosing` (010) lo llama antes de crear la notificación.
- Use cases: `GetMe`, `ListMyClosings`, `MarkClosingPaid`, `ListMyNotifications`, `MarkNotificationRead`.
- Ownership: todos los `/me/**` toman `clientID` del `Identity`; nunca de la URL o el body.
- `cuitMasked`: oculta los 4 primeros dígitos del cuerpo.
