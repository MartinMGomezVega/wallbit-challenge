# 006 — Plan

- `domain/signal/catalog.go`: `type Code string`, constantes, `type Severity int` (`Green=0, Yellow=1, Red=2`) con `String()` → `"GREEN"|"YELLOW"|"RED"`, tabla `catalog map[Code]Meta{Severity, Priority, Label}`.
- `domain/signal/evaluate.go`: una función por regla `func checkMissingInvoices(in Input) (Signal, bool)`; `Evaluate` las corre, aplica exclusiones, ordena (`sort.SliceStable`), arma `Primary` y `Description`.
- `domain/signal/format.go`: `formatARS`, `formatUSD`, `formatDate`, `formatPct(bp)`, `formatRatio(bp)` es-AR, sin `fmt` con locale (implementación manual de separadores). Tests exactos.
- `domain/signal/atypical.go`: mediana (copiar y ordenar; par → promedio de los dos centrales con half-even).
- Handler `GET /api/v1/accountant/signals/catalog` → `[{code, severity, priority, label}]`.
