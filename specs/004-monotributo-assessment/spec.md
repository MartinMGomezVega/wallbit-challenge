# 004 — Liquidación de Monotributo

## Objetivo
Función de dominio pura que, para un cliente Monotributo y un período, calcula facturado, acumulado 12 meses, categoría requerida, cuota a pagar, vencimiento, porcentajes de tope y exclusión (`domain-context.md §7.1, §7.4`).

## Criterios de aceptación
1. `tax.AssessMonotributo(in MonotributoInput) (MonotributoAssessment, error)` es **pura** (sin I/O, sin `time.Now`).
2. `MonotributoInput{Period, CurrentCategory, Invoices []PricedInvoice (con AmountARS ya calculado), ScaleAt func(time.Time) tax.Scale, FeeDueDate time.Time, Params, InboundMovementsARS []DatedAmount, CardSpendARS []DatedAmount}`.
3. `MonotributoAssessment{BilledInPeriodUSD, BilledInPeriodARS, Billed12mARS, CurrentCategory, RequiredCategory?, PctOfCurrentCapBp, PctOfKCapBp, Excluded bool, ExclusionDate?, ExcessARS?, CanReturnFrom?, Fee, FeeDueDate, Inbound12mARS, CardSpend12mARS, ScaleValidFrom}`.
4. Facturas `E` se excluyen del acumulado si `InvoiceECountsForCap == false` (pero siempre cuentan en `BilledInPeriod*`).
5. Casos de test obligatorios (escala agosto 2026):
   - acumulado exactamente igual al cap de D → requerida D;
   - cap de D + 1 centavo → requerida E;
   - acumulado > cap K → `Excluded`, `ExclusionDate` = día exacto del cruce (armar facturas para que cruce el 14/10/2026), `ExcessARS` correcto, `CanReturnFrom = 2030-01-01`;
   - cruce de escala: ventana con facturas antes y después del 01/08/2026 y exclusión evaluada con el cap de la escala vigente en cada día;
   - `InvoiceECountsForCap=false` con solo facturas E → acumulado 0, categoría A;
   - `Fee` = cuota de la **categoría actual**, aunque la requerida sea mayor;
   - porcentajes en bp con redondeo hacia abajo.
6. Performance: 300 clientes × 13 meses de facturas se liquidan en < 200 ms en total en un benchmark (`BenchmarkAssessMonotributo`), usando sumas de prefijos por día.
