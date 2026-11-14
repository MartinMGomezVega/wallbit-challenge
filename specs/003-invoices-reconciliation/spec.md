# 003 — Facturas, inputs del cliente y conciliación

## Objetivo
Recibir lo que Wallbit no tiene (facturas emitidas, facturas de compra para RI, cobros fuera de Wallbit), pesificar facturas y conciliar cobros ↔ facturas.

## Historias
- Como cliente, subo el CSV que exporto de "Mis Comprobantes" de ARCA y se cargan todas mis facturas del mes.
- Como cliente, si no tengo el CSV, cargo una factura a mano.
- Como cliente RI, cargo facturas de compra para el crédito fiscal.
- Como cliente, declaro si cobré algo fuera de Wallbit (o "no cobré nada afuera") y confirmo que terminé.
- Como sistema, concilio cobros con facturas y detecto faltantes.

## Criterios de aceptación
### Modelo
1. `Invoice{ID, ClientID, Type (A|B|C|E), PointOfSale, Number, IssueDate, Currency, NetAmount, VatAmount, TotalAmount, FXRate?, ReceiverName, ReceiverDoc, Source (CSV|MANUAL|SEED), CreatedAt}`. Clave única: `(ClientID, Type, PointOfSale, Number)`.
2. `PurchaseInvoice{ID, ClientID, IssueDate, SupplierName, SupplierCUIT, NetAmount, VatAmount (ARS)}`.
3. `ClientPeriodInput{ClientID, Period, ExternalIncome[]{Amount USD, Description}, DeclaredNoExternalIncome bool, ConfirmedAt?}`.

### Pesificación
4. `invoice.AmountARS(inv, params, fx)` según `domain-context.md §5`. Tests para cada `fx_rule`, factura con TC propio y factura ARS.

### CSV "Mis Comprobantes" (formato supuesto, documentado en README)
5. Encabezados (orden indistinto, match por nombre normalizado): `Fecha`, `Tipo`, `Punto de Venta`, `Número Desde`, `Número Hasta`, `Cód. Autorización`, `Tipo Doc. Receptor`, `Nro. Doc. Receptor`, `Denominación Receptor`, `Tipo Cambio`, `Moneda`, `Imp. Neto Gravado`, `Imp. Neto No Gravado`, `Imp. Op. Exentas`, `IVA`, `Imp. Total`.
6. Separador `;` o `,` (autodetectado por el encabezado). Fecha `dd/mm/yyyy`. Números con coma decimal y punto de miles (`1.234,56`). `Tipo` como `"19 - Factura E"` → se toma el código numérico: `1→A`, `6→B`, `11→C`, `19→E`. Notas de crédito/débito (`2,3,7,8,12,13,20,21`) se **saltean con advertencia** "Notas de crédito/débito fuera de alcance".
7. `Moneda`: `DOL|USD|U$S` → USD; `$|PES|ARS` → ARS. Montos en la moneda de la factura.
8. `POST /api/v1/me/invoices/import` (multipart, campo `file`, máx 2 MB) → `200 { imported, skippedDuplicates, warnings[{line, message}], errors[{line, message}] }`. Las filas con error no se importan; las válidas sí (importación parcial). Archivo sin encabezados válidos → 422.
9. Idempotente: importar dos veces el mismo CSV → la segunda vez `imported: 0`, `skippedDuplicates: N`.

### Carga manual y lectura
10. `POST /api/v1/me/invoices` (JSON) con validación: tipo permitido según régimen a la fecha (MONO: C/E; RI: A/B/E), fecha no futura respecto del Clock, montos > 0, `fxRate` opcional. Duplicado → 409.
11. `GET /api/v1/me/invoices?period=2026-10` → lista del cliente. `GET /api/v1/accountant/clients/{id}/invoices?period=` → lo mismo para el contador.
12. `POST /api/v1/me/purchase-invoices` (solo RI; MONO → 403 con `code: "NOT_ALLOWED_FOR_REGIME"`).

### Inputs del período
13. `PUT /api/v1/me/periods/{period}/external-income` body `{ "items": [{ "amount": {...USD}, "description": "..." }] }` o `{ "none": true }`.
14. `POST /api/v1/me/periods/{period}/confirm` → setea `ConfirmedAt`. Requiere haber respondido external-income (si no → 422 `EXTERNAL_INCOME_NOT_ANSWERED`).
15. Toda escritura de esta feature emite `ClosingInputsChanged{ClientID, Period}`.

### Conciliación
16. `reconciliation.Reconcile(incomeMovements, invoices, params) Result` según `domain-context.md §6`. Tests: match exacto, con comisión dentro de tolerancia, fuera de tolerancia, cobro anticipado (antes de la factura), nombre con sufijo distinto ("Acme Inc." vs "ACME"), dos facturas iguales para dos cobros (1 a 1), sin candidatos.
17. `reconciliation.KnownPayers(invoicesBefore []Invoice) map[string]bool`.

### Seed
18. El seed genera facturas para los cobros de clientes generados (factura E fechada 0–7 días antes del cobro, `fxRate` = compra BNA del día hábil anterior), facturas de compra para RI, e inputs del período (90% confirmados, 10% sin confirmar). Los demos (002) reciben sus facturas exactas.
