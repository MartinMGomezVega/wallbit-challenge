# 007 — Plan

## Domain
- `domain/closing/closing.go`: `Status` (`PENDING_CLIENT_INPUT`, `READY_FOR_REVIEW`, `APPROVED`, `NOTIFIED`), `Closing`, `Approval{ID, Version, ResultHash, ApprovedBy, ApprovedAt, Note, Batch bool}`, métodos `ApplyResult(result, hash, now)`, `MarkReady()`, `Approve(...)`, `MarkNotified()`, `MarkPaid(now)` con validación de transiciones.
- `domain/closing/hash.go`: JSON canónico (marshal de structs con campos ordenados; maps → slices ordenadas) + SHA-256 hex.
- `domain/closing/query.go`: `Filter`, `Sort`, `func Apply(rows []Row, f Filter, s Sort) []Row` **pura** (la usan el listado, el summary y la cola de IDs de 009).

## Ports
```go
type ClosingRepository interface {
  Get(ctx, id string) (closing.Closing, error)
  ByClientPeriod(ctx, clientID string, p period.Period) (closing.Closing, bool, error)
  ListByPeriod(ctx, p period.Period) ([]closing.Closing, error)
  Save(ctx, c closing.Closing) error // upsert por (clientID, period)
  LatestExclusionDate(ctx, clientID string) (*time.Time, error) // implementa el puerto de 005
}
```

## Use cases
- `ComputeClosing` (criterio 1). Inyecta todos los repos + providers + Clock + IDs.
- `RunMonthlyClosing` con worker pool: canal de `clientID`, `N` workers, `sync.WaitGroup`, resultados por canal, `ctx` cancelable.
- `RecalcOnChange` implementa `ports.ChangeNotifier`.
- `RecalcPeriod` (para PUT parameters).
- `ListClosings`, `ClosingsSummary`, `GetClosingDetail`. `ListClosings` construye `[]closing.Row` (vista plana con datos del cliente) y aplica `closing.Apply`.
- `Scheduler` en `infrastructure/scheduler` (ticker + clock).

## HTTP
- Parseo de filtros en un helper compartido `parseClosingFilter(r *http.Request) (closing.Filter, closing.Sort, error)` → lo reusa 009.
- Rutas de la spec. `period` inválido → 422.

## Notas
- Para evitar N+1 en el listado, `ListClosings` carga clientes una vez en un map.
- Recalcular de forma sincrónica es suficiente para el prototipo (documentar que en producción sería asincrónico con cola).
