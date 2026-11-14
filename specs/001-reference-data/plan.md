# 001 — Plan

## Domain
- `domain/money`
  - `type Currency string` (`ARS`, `USD`).
  - `type Money struct { Amount int64; Currency Currency }` (centavos).
  - `type Rate int64` (×10.000). `RateFromDecimal("1539.87")`.
  - `type BasisPoints int64`. `ApplyBp(m Money, bp BasisPoints) Money`, `RatioBp(a, b Money) BasisPoints`.
  - Conversión con `math/big.Int` y half-even.
- `domain/period`: `type Period struct{ Year int; Month time.Month }`; helpers de §criterios 3–4. Location AR inyectada vía `period.SetLocation` o constante del paquete.
- `domain/tax/scale.go`
  ```go
  type Category string // "A".."K"
  type CategoryRow struct { Category Category; CapARS money.Money; ServicesFeeARS money.Money }
  type Scale struct { ValidFrom time.Time; Rows []CategoryRow } // Rows ordenadas A..K
  func (s Scale) CategoryFor(amount money.Money) (Category, error)
  func (s Scale) Row(c Category) (CategoryRow, bool)
  func (s Scale) MaxCap() money.Money
  func (c Category) Less(o Category) bool
  ```
- `domain/client/cuit.go`: `type CUIT string`; `ParseCUIT` valida formato `NN-NNNNNNNN-N` y dígito verificador (módulo 11); `Terminacion() int` (penúltimo dígito).
- `domain/parameters`: struct `AccountantParameters` con los campos de §9 (`InvoiceECountsForCap bool`, `FXRule FXRule`, `CategoryCapWarningBp`, …), `Defaults()`, `Validate() error`, `UpdatedAt`, `UpdatedBy`.

## Ports (`application/ports`)
```go
type TaxScaleProvider interface { MonotributoScale(ctx context.Context, at time.Time) (tax.Scale, error) }
type DueDates struct { Monotributo, VAT, SelfEmployed time.Time }
type DueDateProvider interface { DueDates(ctx context.Context, p period.Period, cuit client.CUIT) (DueDates, error) }
type Quote struct { Day time.Time; Buy, Sell money.Rate }
type ExchangeRateProvider interface { USDQuote(ctx context.Context, day time.Time) (Quote, error) }
type ParametersRepository interface { Get(ctx) (parameters.AccountantParameters, error); Save(ctx, p) error }
```

## Infrastructure
- `infrastructure/arca/mock/`
  - `data/monotributo_scales.json`:
    ```json
    [{"validFrom":"2026-02-01","rows":[{"category":"A","cap":"10277988.13","servicesFee":"42386.74"}, ...]},
     {"validFrom":"2026-08-01","rows":[...]}]
    ```
  - `data/due_dates.json`: `{"2026-11":{"monotributo":"2026-11-20","vat":{"0":"2026-11-18","1":"2026-11-18",...},"selfEmployed":{"0":"2026-11-05",...}}}`.
  - `client.go`: `type Client struct{ latency time.Duration; failEvery int; calls atomic.Int64 }`, `New(opts ...Option)`. Comentario en el tipo: *"Simula el cliente HTTP de ARCA. La implementación real haría GET a <servicio> y mapearía la respuesta a tax.Scale."* Carga con `//go:embed data/*.json` al construir.
- `infrastructure/fx/mock/`
  - `data/usd_bna.json`: `[{"day":"2026-10-01","buy":"1486.11","sell":"1539.87"}, ...]` ordenado.
  - `client.go`: búsqueda binaria por día; fallback hacia atrás.
- `infrastructure/memory/parameters_repo.go`: inicializa con `Defaults()`.

## HTTP
- `GET /api/v1/accountant/parameters`, `PUT /api/v1/accountant/parameters`.
- DTO con bp como número y `missingInvoiceMinUsd` como Money DTO.
- Use cases: `usecase.GetParameters`, `usecase.UpdateParameters` (valida, setea `UpdatedAt` con Clock y `UpdatedBy` = "accountant").

## Datos de tipo de cambio faltantes
Para cada mes de oct-2025 a sep-2026 y nov-2026 (1 al 13), tomar la tabla de `https://dolarhistorico.com/cotizacion-dolar-oficial/mes/{mes}-{anio}` (mes en español minúscula). Si el agente no tiene acceso a la web, **parar y pedirle al usuario** los datos; no inventar.
