# 009 — Plan

## Backend
- Use case `ListClosingIDs` que reutiliza el armado de `[]closing.Row` + `closing.Apply` de `ListClosings` (extraer a una función común `buildRows` si no existe).
- Handler con `parseClosingFilter` (007).

## Frontend
```
features/closing-review/
├── index.ts
├── api.ts                    # getClosingIds(filter), getClosingDetail(id)
├── reviewQueue.ts            # lógica pura de la cola (sin React)
├── useReviewQueue.ts         # estado: ids, index, reviewed:Set, origin filter
├── ReviewModal.tsx
├── ReviewHeader.tsx          # Anterior / "3 de 47" / Siguiente / cerrar
├── sections/ SummarySection.tsx, SignalsSection.tsx, MovementsSection.tsx,
│            InvoicesSection.tsx, ClientMessageSection.tsx, HistorySection.tsx
├── useKeyboardNav.ts
└── __tests__/
```

### `reviewQueue.ts` (puro, 100% testeado)
```ts
type Queue = { ids: string[]; index: number; reviewed: Set<string>; label: string };
createQueue(ids, startId?): Queue        // startId ausente → 0; startId no en ids → lo antepone
next(q): Queue; prev(q): Queue            // no se pasan de los bordes
canNext(q), canPrev(q), current(q)
markReviewed(q, id): Queue
pageOf(q, pageSize): number               // para criterio 8
```

### Integración con el tablero
- `DashboardPage` expone `openReview({mode, startId?, severity?})`:
  - `row`: `getClosingIds(currentFilter)` → `createQueue(ids, startId)`.
  - `chip`: `getClosingIds({...currentFilter, severity:[color]})`.
  - `selection`: IDs seleccionados en el orden de la tabla (no pide al backend).
- Cola guardada en estado de React (contexto `ReviewQueueProvider`) + `client` en la URL. Si se recarga con `client=` y no hay cola en memoria, se reconstruye en modo `row`.
- Mientras se arma la cola, el modal abre con el detalle del cliente clickeado (no esperar a los IDs).
