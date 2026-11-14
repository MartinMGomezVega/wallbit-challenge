# 002 — Plan

## Domain
- `domain/client/client.go`: tipos del criterio 1. `func (c Client) RegimeAt(t time.Time) Regime` usando `RegimeHistory`.
- `domain/movement/movement.go`: `MovementType`, `Counterparty`, `Movement`, `func (m Movement) AmountARS() (money.Money, error)` usando `FXRate`.
- `domain/movement/classify.go`: `type Rule string` (`OWN_ACCOUNT_TRANSFER`, …, `FALLBACK`, `MANUAL_OVERRIDE`); `Classify(m Movement, knownPayers map[string]bool) Classification`.
- `domain/movement/normalize.go`: `NormalizeName(s string) string` (minúsculas, sin acentos con `unicode`/tabla manual, sin puntuación, quita sufijos `inc|llc|ltd|sa|srl|gmbh|corp|co`). Lo reusa 003.

## Ports
```go
type ClientRepository interface {
  List(ctx) ([]client.Client, error); Get(ctx, id string) (client.Client, error); Save(ctx, c client.Client) error
}
type MovementRepository interface {
  ByClientBetween(ctx, clientID string, from, to time.Time) ([]movement.Movement, error)
  Get(ctx, id string) (movement.Movement, error)
  SaveAll(ctx, ms []movement.Movement) error
}
type ClassificationRepository interface {
  Get(ctx, movementID string) (movement.Classification, bool, error)
  Save(ctx, c movement.Classification) error
}
```
Repos en `infrastructure/memory` con índices por `clientID` y slices ordenados por fecha.

## Use case
- `usecase.OverrideClassification{Movements, Classifications, Clock}` → valida kind, guarda, devuelve. Emite un evento interno `ClosingInputsChanged{ClientID, Period}` vía puerto `ports.ChangeNotifier` (no-op en esta feature; 007 lo implementa).

## Seed (`infrastructure/seed`)
- `Generate(cfg SeedConfig, fx ports.ExchangeRateProvider, ids ports.IDGenerator) (Dataset, error)`; `Dataset{Clients, Movements}` (003 agrega `Invoices`, `PurchaseInvoices`, `ClientInputs`).
- RNG: `math/rand/v2` con `rand.NewPCG(seed, seed)`.
- Perfil por cliente: `baseMonthlyUSD`, 1–3 pagadores extranjeros (pool de ~40 empresas inventadas: "Northwind Labs LLC" (US), "Brightloop GmbH" (DE), "Kiwi Studio Ltd" (GB), "Acme Digital Inc" (US), "Orbital SaaS Corp" (US), …, más plataformas "Deel", "Upwork" kind `PLATFORM`), día de cobro habitual, variación ±15%/mes.
- Otros movimientos por mes: 5–25 `CARD_SPEND` (USD 5–300), 0–2 `STABLECOIN_BUY`, 0–2 `OUTBOUND_TRANSFER`, 0–1 `FEE`, ocasional `CARD_REFUND`, ocasional transferencia propia (`isOwnAccount`).
- **Clientes de demo** (IDs fijos, datos escritos a mano en `seed/demo.go`, nombres inventados):

| ID | Nombre | Régimen | Escenario (señal principal esperada) |
|---|---|---|---|
| cli_demo_01 | Lucía Fernández | MONO D | GREEN |
| cli_demo_02 | Ignacio Sosa | MONO K | `MONO_EXCLUSION_K_CAP_EXCEEDED` (cruza el tope K el 14/10/2026) |
| cli_demo_03 | Carla Benítez | MONO E | `MISSING_INVOICES` (2 cobros de octubre sin factura) |
| cli_demo_04 | Diego Romero | MONO C | `MONO_CATEGORY_CHANGE_REQUIRED` (necesita E) |
| cli_demo_05 | Valentina Ruiz | MONO F | `MONO_NEAR_CATEGORY_CAP` (~90% de F) |
| cli_demo_06 | Tomás Álvarez | MONO D | `ATYPICAL_INCOME` (cobro 3× mediana, facturado) |
| cli_demo_07 | Sofía Medina | MONO C | `UNCLASSIFIED_MOVEMENTS` (transferencia local de persona AR) |
| cli_demo_08 | Joaquín Herrera | MONO K | `MONO_NEAR_K_CAP` (~88% de K) |
| cli_demo_09 | Camila Torres | RI | GREEN (solo factura E, saldo a favor de IVA) |
| cli_demo_10 | Federico Díaz | RI | `INVOICE_WITHOUT_PAYMENT` + IVA a pagar por factura A local |
| cli_demo_11 | Agustina Castro | RI desde 2026-08-12 | `POST_EXCLUSION_C_INVOICES` (2 facturas C en octubre) |
| cli_demo_12 | Nicolás Peralta | MONO G | `DEPOSITS_EXCEED_K_CAP` (+ `HIGH_CARD_SPENDING`) |
| cli_demo_13 | Julieta Molina | MONO D | `EXTERNAL_INCOME_DECLARED` |
| cli_demo_14 | Bruno Acosta | MONO H | `MONO_CATEGORY_DOWNGRADE_POSSIBLE` |
| cli_demo_15 | Paula Rojas | MONO B | Estado `PENDING_CLIENT_INPUT`, sin facturas de octubre → `MISSING_INVOICES`. Después de subir el CSV de ejemplo y confirmar → GREEN y `READY_FOR_REVIEW` (flujo de la demo). |

> Los montos exactos de cada demo se calculan contra la escala de agosto 2026 y el tipo de cambio del fixture. Escribir en `demo.go` un comentario por cliente con la cuenta que justifica el escenario.

- Categoría actual de clientes generados: la requerida por lo facturado `2025-07-01..2026-06-30` con la escala de agosto (simula la recategorización de agosto).
