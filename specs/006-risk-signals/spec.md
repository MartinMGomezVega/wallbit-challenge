# 006 — Señales de riesgo y semáforo

## Objetivo
Función de dominio pura que, a partir de la liquidación, la conciliación, las clasificaciones y los inputs del cliente, produce la lista de señales del catálogo (`domain-context.md §8`), el semáforo y la descripción breve que ve el contador.

## Criterios de aceptación
1. `signal.Evaluate(in Input) Evaluation` pura.
   - `Input{Regime, Mono *tax.MonotributoAssessment, RI *tax.RIAssessment, Reconciliation reconciliation.Result, PeriodMovements []ClassifiedMovement, HistoricIncomeUSD []money.Money, ClientInput clientinput.ClientPeriodInput, CInvoicesAfterRI []invoice.Invoice (con ARS), Params}`.
   - `Evaluation{Severity, Signals []Signal (ordenadas por severidad desc y prioridad asc), Primary *Signal, Description string}`.
2. `Signal{Code, Severity, Priority, Message string, Data map[string]string}`. `Data` lleva los valores crudos (ej. `"totalUSD": "4200.00"`) para que el front pueda mostrar detalle.
3. Un test por cada código del catálogo que verifica disparo **y** no disparo en el borde (ej. `MONO_NEAR_K_CAP` con `pct = umbral − 1bp` no dispara, `= umbral` dispara).
4. Exclusiones mutuas:
   - `MONO_EXCLUSION_K_CAP_EXCEEDED` presente ⇒ no se emiten `MONO_NEAR_K_CAP`, `MONO_CATEGORY_CHANGE_REQUIRED`, `MONO_NEAR_CATEGORY_CAP`, `MONO_CATEGORY_DOWNGRADE_POSSIBLE`.
   - `MONO_CATEGORY_CHANGE_REQUIRED` ⇒ no `MONO_NEAR_CATEGORY_CAP`.
   - Señales `MONO_*`, `DEPOSITS_EXCEED_K_CAP`, `HIGH_CARD_SPENDING` solo si `Regime == MONOTRIBUTO`; `POST_EXCLUSION_C_INVOICES` solo si RI.
5. Sin señales ⇒ `GREEN` con descripción del template verde según régimen.
6. Descripción = `Primary.Message`. Formatos es-AR exactos de `domain-context.md §8` (test con strings exactos, ej. `"2 cobros por US$ 4.200,00 sin factura asociada."`).
7. `signal.Catalog()` expone códigos, severidad, prioridad y una etiqueta corta para filtros (ej. `MISSING_INVOICES → "Faltan facturas"`). Endpoint `GET /api/v1/accountant/signals/catalog` la devuelve (lo usa el front para filtros).

## Etiquetas cortas (filtros)
| Code | Label |
|---|---|
| MONO_EXCLUSION_K_CAP_EXCEEDED | Excede tope K (pasa a RI) |
| POST_EXCLUSION_C_INVOICES | Facturas C post-exclusión |
| MISSING_INVOICES | Faltan facturas |
| DEPOSITS_EXCEED_K_CAP | Depósitos sobre tope K |
| MONO_NEAR_K_CAP | Cerca del tope K |
| MONO_CATEGORY_CHANGE_REQUIRED | Debe recategorizar |
| MONO_NEAR_CATEGORY_CAP | Cerca del tope de su categoría |
| ATYPICAL_INCOME | Cobro fuera de lo habitual |
| UNCLASSIFIED_MOVEMENTS | Movimientos sin clasificar |
| EXTERNAL_INCOME_DECLARED | Cobros fuera de Wallbit |
| INVOICE_WITHOUT_PAYMENT | Facturas sin cobro |
| HIGH_CARD_SPENDING | Gastos con tarjeta altos |
| MONO_CATEGORY_DOWNGRADE_POSSIBLE | Puede bajar de categoría |
