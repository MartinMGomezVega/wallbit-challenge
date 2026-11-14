# 004 — Plan

- Archivo `domain/tax/monotributo.go`.
- `type PricedInvoice struct { Invoice invoice.Invoice; AmountARS money.Money; AmountUSD *money.Money }`.
- `type DatedAmount struct { Day time.Time; Amount money.Money }`.
- Algoritmo:
  1. Filtrar facturas computables (C, y E según parámetro).
  2. Construir `dailyTotals map[civilDay]int64` → slice ordenado de los 365+365 días necesarios → prefix sums.
  3. `sum(a,b]` en O(1). `Billed12m = sum(periodEnd−12m, periodEnd]`.
  4. Exclusión: iterar días `d` de `periodEnd−364` a `periodEnd`; si `sum(d−12m, d] > ScaleAt(d).MaxCap()` → primer `d` es `ExclusionDate`; `ExcessARS` = acumulado a `periodEnd` − cap K vigente a `periodEnd`.
  5. Categoría requerida con `ScaleAt(periodEnd)`.
  6. `Fee` con `ScaleAt(FeeDueDate).Row(CurrentCategory).ServicesFeeARS`.
  7. Inbound/card 12m con la misma técnica.
- Usar un tipo `civilDay` (`struct{Y int; M time.Month; D int}`) para evitar problemas de zona horaria.
- Sin dependencias de `application`. El caso de uso que arma el input vive en 007.
