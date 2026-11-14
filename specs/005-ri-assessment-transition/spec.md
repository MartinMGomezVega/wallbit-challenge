# 005 — Liquidación de Responsable Inscripto y transición Monotributo → RI

## Objetivo
Calcular las obligaciones mensuales de un RI (`domain-context.md §7.2`) y modelar el pase de Monotributo a RI confirmado por el contador (§7.3).

## Criterios de aceptación
### Liquidación RI
1. `tax.AssessRI(in RIInput) (RIAssessment, error)` pura.
2. `RIInput{Period, Invoices []PricedInvoice, PurchaseInvoices, Profile client.RIProfile, DueDates ports.DueDates (pasadas como valores), Params}`.
3. `RIAssessment{BilledInPeriodUSD, BilledInPeriodARS, ExportSalesARS, LocalSalesNetARS, OutputVat, InputVat, PreviousVatCredit, VatToPay, NewVatCredit, VatDueDate, SelfEmployedContribution, SelfEmployedDueDate, IncomeTaxAdvance?, IncomeTaxAdvanceDueDate?, TotalToPay}`.
4. Tests: solo exportación (IVA 0, saldo a favor crece con compras), factura A local (débito 21% del neto), saldo a favor previo que absorbe débito parcial y total, anticipo del período presente/ausente, `TotalToPay` = IVA + autónomos + anticipo.

### Transición
5. `POST /api/v1/accountant/clients/{id}/regime-change` body `{ "to": "RI", "effectiveDate": "2026-10-14", "note": "..." }`:
   - solo si el cliente es MONO a esa fecha;
   - `effectiveDate` no futura; si el último cierre tiene `ExclusionDate`, debe coincidir (si no → 422 `EFFECTIVE_DATE_MISMATCH`, el contador puede forzar con `"force": true` y queda en la nota);
   - cierra el tramo MONO en `effectiveDate − 1 día`, abre RI desde `effectiveDate`, `Reason: "EXCLUSION_K_CAP"`;
   - crea `RIProfile` vacío (autónomos 0) y emite `ClosingInputsChanged`.
6. `PUT /api/v1/accountant/clients/{id}/ri-profile` body `{ selfEmployedMonthlyContribution, previousVatCredit, incomeTaxAdvances: [{period, amount, dueDate}] }` → valida montos ≥ 0, períodos válidos; emite `ClosingInputsChanged`.
7. Un cierre de un período que contiene la fecha efectiva se liquida como **RI** (supuesto: el contador decide; el prototipo usa el régimen vigente al **último día** del período). Documentar en README.
8. `POST_EXCLUSION_C_INVOICES` se calcula en 006 con `RegimeHistory`; esta feature expone `client.RIFrom() (time.Time, bool)`.
