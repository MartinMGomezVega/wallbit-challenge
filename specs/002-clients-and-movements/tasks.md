# 002 — Tasks

- [x] **T002.1** `domain/client`: Client, Regime, RegimeHistory, RIProfile, `RegimeAt`. Tests.
- [x] **T002.2** `domain/movement`: tipos + `AmountARS`. Tests.
- [x] **T002.3** `NormalizeName` con tests (acentos, sufijos, puntuación, mayúsculas, espacios dobles).
- [x] **T002.4** `Classify` R1..R9. Tests: un caso por regla + precedencia R1>R6 + knownPayer local (R5 gana sobre R8).
- [x] **T002.5** Ports + repos en memoria (Client, Movement, Classification) con `sync.RWMutex`. Tests incluyendo acceso concurrente con `-race` (100 goroutines leyendo/escribiendo).
- [x] **T002.6** `ports.ChangeNotifier` (interfaz) + implementación no-op.
- [x] **T002.7** Use case `OverrideClassification` + handler `PATCH /api/v1/accountant/movements/{id}/classification`. Tests: 200, 404, 422, 403.
- [x] **T002.8** Seed: generador de clientes y movimientos (perfiles, pagadores, ruido). Tests de determinismo (criterio 7) y distribución (criterio 9).
- [x] **T002.9** Seed: `demo.go` con los 15 clientes de demo (movimientos; las facturas se agregan en T003.x). Test: existen los 15 IDs y sus regímenes.
- [x] **T002.10** Wiring: al arrancar, `seed.Generate` carga repos y corre la clasificación inicial de todos los movimientos (worker pool). `make coverage` ≥ 90%.
