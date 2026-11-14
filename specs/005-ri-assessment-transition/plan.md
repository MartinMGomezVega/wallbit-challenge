# 005 — Plan

- `domain/tax/ri.go`: `AssessRI`. IVA de facturas A/B: usar `VatAmount` de la factura (no recalcular 21% si viene informado; si no viene, `Net × 21%`). Facturas E: `VatAmount = 0`.
- `domain/client/regime.go`: `func (c *Client) ChangeRegime(to Regime, effective time.Time, reason, note string, now time.Time) error` con las validaciones del criterio 5 (excepto la de ExclusionDate, que hace el caso de uso porque necesita el cierre).
- Use cases: `ChangeRegime{Clients, Closings (lectura del último cierre), Clock, Changes}`, `UpdateRIProfile`.
- Handlers con rutas de la spec. Errores: `ErrInvalidTransition` → 409.
- El repositorio de cierres (`ClosingRepository`) se define en 007; para no bloquear, `ChangeRegime` recibe un puerto mínimo `LatestExclusionDateReader{ LatestExclusionDate(ctx, clientID) (*time.Time, error) }` que 007 implementa. En esta feature usar un fake en tests y un no-op en wiring.
