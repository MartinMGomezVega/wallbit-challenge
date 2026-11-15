# 011 — Plan

## Backend
- `domain/closing`: agregar `Published` y `Publish(snapshot, now)`; `NotifyClosing` (010) lo llama antes de crear la notificación.
- Use cases: `GetMe`, `ListMyClosings`, `MarkClosingPaid`, `ListMyNotifications`, `MarkNotificationRead`.
- Ownership: los endpoints de cliente validan que el `clientID` de la URL coincida con el caller (o sea `accountant`).
- `cuitMasked`: oculta los 4 primeros dígitos del cuerpo.
