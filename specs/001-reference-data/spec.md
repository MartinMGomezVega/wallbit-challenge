# 001 — Reference data: dinero, períodos, mocks de ARCA y tipo de cambio, parámetros del contador

## Objetivo
Tener los value objects base (`Money`, `Rate`, `Period`) y los **proveedores externos mockeados con datos reales**, escritos como si fueran clientes de servicios externos, detrás de puertos. Más los parámetros globales del contador.

## Historias
- Como sistema, necesito la escala de Monotributo vigente a una fecha para calcular categorías.
- Como sistema, necesito los vencimientos de un período según la terminación de CUIT.
- Como sistema, necesito el tipo de cambio BNA de un día (compra/venta) para pesificar.
- Como contador, quiero ver y editar los parámetros globales.

## Criterios de aceptación
### Money / Rate / Period
1. `money.New(amountMinor int64, currency)`; `Add/Sub` fallan (error) con monedas distintas; `ConvertUSDToARS(usd Money, rate Rate) (Money, error)` con redondeo half-even (casos: `.5` hacia par).
2. `money.ParseDecimal("1234.56", "ARS")` y `Money.DecimalString()` ida y vuelta sin pérdida. Acepta coma decimal es-AR (`"1.234,56"`) vía `ParseLocalized`.
3. `period.Parse("2026-10")`; `Start()`/`End()` en zona AR; `Prev()`, `Next()`; `Last12MonthsWindow()` → `(periodEnd − 12 meses, periodEnd]`.
4. `period.NextBusinessDay(date)` (sáb/dom → lunes) y `PrevBusinessDay`.

### TaxScaleProvider (mock ARCA)
5. `MonotributoScale(ctx, at time.Time) (tax.Scale, error)`:
   - `2026-10-15` → escala vigente desde 2026-08-01 (`domain-context.md §2.1`), K cap = `126.610.838,75`.
   - `2026-03-01` → escala de febrero (§2.2).
   - `2025-12-01` → escala de febrero (supuesto §2.2 nota).
6. `Scale.CategoryFor(amountARS)` devuelve la menor categoría con `cap ≥ amount`, o `ErrExceedsMaxCategory`.

### DueDateProvider (mock ARCA)
7. `DueDates(ctx, obligationPeriod period.Period, cuit client.CUIT) (DueDates, error)` con `Monotributo`, `VAT`, `SelfEmployed`. Para `2026-11` usa la tabla real (§2.3); para otros meses, el patrón referencial con corrimiento de fin de semana.
   - CUIT `20-12345678-9` (terminación 8) en `2026-11` → IVA `2026-11-25`, autónomos `2026-11-11`, monotributo `2026-11-20`.

### ExchangeRateProvider (mock tipo de cambio)
8. `USDQuote(ctx, day time.Time) (Quote{Day, Buy, Sell Rate}, error)`. Si el día no tiene cotización, devuelve la del día hábil anterior más cercano con dato (hasta 10 días hacia atrás; si no → `ErrRateNotFound`).
   - `2026-10-12` (feriado) → cotización del `2026-10-09`.
9. Fixture con días hábiles **2025-10-01 a 2026-11-13** con datos reales (octubre ya está en `domain-context.md §2.4`).
10. Ambos mocks: latencia configurable (`MOCK_LATENCY_MS`), respetan cancelación del `context`, y pueden inyectarse para fallar (`FailEvery n`) en tests.

### Parámetros del contador
11. `parameters.Defaults()` con los valores de `domain-context.md §9`. `Validate()` rechaza: bp fuera de `[0, 100000]`, ventanas ≤ 0, `fx_rule` desconocido.
12. `GET /api/v1/accountant/parameters` → parámetros actuales. `PUT` con el objeto completo → valida, guarda, recalcula cierres afectados y devuelve `{ "parameters": ..., "recalculated": N }`. En la UI, el botón "Guardar" permanece deshabilitado sin cambios, notifica con toast al guardar y protege con modal de confirmación si el usuario intenta navegar con cambios pendientes.

## Fuera de alcance
Cliente HTTP real contra ARCA/BCRA (solo se documenta la interfaz).
