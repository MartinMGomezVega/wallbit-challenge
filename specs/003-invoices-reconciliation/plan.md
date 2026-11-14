# 003 — Plan

## Domain
- `domain/invoice/invoice.go`: `Type`, `Source`, `Invoice`, `PurchaseInvoice`, `Key()`.
- `domain/invoice/pesify.go`: `func AmountARS(inv Invoice, rule parameters.FXRule, q QuoteLookup) (money.Money, error)` donde `type QuoteLookup func(day time.Time) (Quote, error)` (función, para no depender de ports en domain).
- `domain/invoice/csv.go`: `ParseMisComprobantes(r io.Reader) (rows []ParsedRow, warnings, errors []LineIssue, err error)`. Solo `encoding/csv` + parsing propio de números/fechas. `ParsedRow` → `Invoice` (sin ID).
- `domain/invoice/validate.go`: `ValidateManual(inv, regimeAtDate, now) error` (ValidationError por campo).
- `domain/clientinput/input.go`: `ClientPeriodInput`, `Answered() bool`, `Confirm(now) error`.
- `domain/reconciliation/reconcile.go`: `Result{Matched []Pair; UnmatchedIncome []movement.Movement; UnmatchedInvoices []invoice.Invoice}`; `Reconcile(...)`; `KnownPayers(...)`.

## Ports
```go
type InvoiceRepository interface {
  ByClientBetween(ctx, clientID string, from, to time.Time) ([]invoice.Invoice, error)
  Exists(ctx, clientID string, key invoice.Key) (bool, error)
  Save(ctx, inv invoice.Invoice) error
}
type PurchaseInvoiceRepository interface { ByClientBetween(...); Save(...) }
type ClientInputRepository interface { Get(ctx, clientID string, p period.Period) (clientinput.ClientPeriodInput, error); Save(...) error }
```

## Use cases
- `ImportInvoicesCSV{Clients, Invoices, IDs, Clock, Changes}` → parsea, valida cada fila (régimen a la fecha), deduplica por `Key`, guarda, emite cambio por cada período afectado.
- `AddInvoice`, `AddPurchaseInvoice`, `ListInvoices`, `SetExternalIncome`, `ConfirmPeriodInput`.

## HTTP
- Multipart con `r.ParseMultipartForm(2 << 20)`; `http.MaxBytesReader`.
- Rutas de la spec. DTO de factura con montos como Money DTO.

## Testdata
- `internal/domain/invoice/testdata/mis_comprobantes_ok.csv` (5 filas: 3 E en USD, 1 C en ARS, 1 nota de crédito).
- `..._semicolon.csv`, `..._bad_rows.csv` (fecha inválida, tipo desconocido, monto vacío).
- `mis_comprobantes_demo_paula.csv`: facturas E de octubre 2026 que concilian **exactamente** con los cobros de octubre de `cli_demo_15` (mismos receptores, montos y fechas compatibles). Copiarlo a `frontend/public/samples/mis_comprobantes_ok.csv` para que la cliente de demo lo descargue y lo suba (lo usa 011 y el guion de 012). Test e2e: importarlo como `cli_demo_15` + confirmar → cierre GREEN y `READY_FOR_REVIEW`.
