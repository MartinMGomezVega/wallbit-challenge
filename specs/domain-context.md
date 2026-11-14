# Domain context — Impuestos de freelancers que exportan servicios (Argentina, 2026)

> Reglas de negocio, datos de referencia reales, parámetros del contador y catálogo de señales.
> Todas las features usan estos nombres (códigos, campos, enums). **No inventes nombres nuevos**: si falta algo, agregalo acá primero.

---

## 1. Glosario

| Término | Significado |
|---|---|
| **Cliente** | Usuario de Wallbit que contrató el servicio contable. |
| **Contador** | Contador matriculado de Wallbit. Único rol que aprueba. |
| **Período** | Mes calendario, formato `YYYY-MM` (ej. `2026-10`). Zona horaria `America/Argentina/Buenos_Aires`. |
| **Cierre (Closing)** | Resultado del período para un cliente: movimientos clasificados, facturas, conciliación, liquidación (assessment), señales, semáforo, estado. |
| **Movimiento** | Operación dentro de Wallbit (cobro, transferencia, gasto con tarjeta, conversión a stablecoin, comisión). Siempre trae el tipo de cambio de la operación. |
| **Factura** | Comprobante emitido por el cliente en ARCA. Wallbit no las tiene: las carga el cliente (CSV de "Mis Comprobantes" o manual). |
| **Factura E** | Comprobante de exportación (cliente del exterior). Código ARCA `19`. IVA 0%. |
| **Factura C** | Comprobante de monotributista a cliente local. Código ARCA `11`. |
| **Factura A / B** | Comprobantes de Responsable Inscripto. Códigos ARCA `1` (A) y `6` (B). IVA discriminado. |
| **Monotributo** | Régimen simplificado: cuota fija mensual por categoría (A a K) según ingresos brutos de los últimos 12 meses. |
| **Responsable Inscripto (RI)** | Régimen general: IVA mensual, aportes de autónomos, Ganancias (anticipos + DDJJ anual). |
| **Exclusión** | Pérdida del Monotributo por superar el tope de la categoría máxima (K). Pasa a RI desde la fecha en que se superó. No puede volver por 3 años calendario. |
| **Semáforo / Severity** | `RED` (riesgo), `YELLOW` (revisar), `GREEN` (todo ok). |
| **Señal (Signal)** | Situación detectada por reglas. Un cierre puede tener varias; el semáforo es la peor. |

---

## 2. Datos de referencia reales (van en los mocks)

### 2.1 Escala de Monotributo vigente desde 01/08/2026 (servicios)

Fuente: ARCA vía hacecuentas.com, iProfesional, cotizacion-dolar.com.ar (ver §9). Montos en ARS. Las categorías C a G se publican redondeadas a pesos (sin centavos): usar `,00`.

| Cat | Tope anual ingresos brutos | Cuota mensual servicios |
|---|---|---|
| A | 12.009.410,45 | 49.527,18 |
| B | 17.595.183,00 | 56.379,08 |
| C | 24.670.494,00 | 66.020,00 |
| D | 30.628.651,00 | 84.613,00 |
| E | 36.028.231,00 | 119.811,00 |
| F | 45.151.659,00 | 150.784,00 |
| G | 53.995.799,00 | 230.313,00 |
| H | 81.924.660,37 | 522.706,68 |
| I | 91.699.761,90 | 963.747,86 |
| J | 105.012.519,20 | 1.167.299,76 |
| K | 126.610.838,75 | 1.614.446,04 |

### 2.2 Escala de Monotributo vigente 01/02/2026 – 31/07/2026 (servicios)

| Cat | Tope anual | Cuota mensual servicios |
|---|---|---|
| A | 10.277.988,13 | 42.386,74 |
| B | 15.058.447,71 | 48.250,78 |
| C | 21.113.696,52 | 56.501,85 |
| D | 26.212.853,42 | 72.414,10 |
| E | 30.833.964,37 | 102.537,97 |
| F | 38.642.048,36 | 129.045,32 |
| G | 46.211.109,37 | 197.108,23 |
| H | 70.113.407,33 | 447.346,93 |
| I | 78.479.211,62 | 824.802,26 |
| J | 89.872.640,30 | 999.007,65 |
| K | 108.357.084,00 | 1.381.687,90 |

> Para fechas anteriores al 01/02/2026 el mock devuelve la escala de febrero (supuesto documentado; el prototipo cierra octubre 2026 y solo necesita la escala de agosto).

### 2.3 Vencimientos de noviembre 2026 (obligaciones del período octubre 2026)

| Obligación | Terminación CUIT | Vencimiento |
|---|---|---|
| Monotributo (cuota de noviembre) | todas | **20/11/2026** |
| IVA (DDJJ + pago, período 10/2026) | 0 y 1 | 18/11/2026 |
| | 2 y 3 | 19/11/2026 |
| | 4 y 5 | 20/11/2026 |
| | 6 y 7 | 24/11/2026 |
| | 8 y 9 | 25/11/2026 |
| Autónomos (aporte mensual) | 0, 1, 2, 3 | 05/11/2026 |
| | 4 a 9 | 11/11/2026 (estimado por la fuente) |

Para otros meses el mock usa el **patrón referencial**: Monotributo día 20; IVA 18/19/20/21/22 según terminación (0-1, 2-3, 4-5, 6-7, 8-9); autónomos día 5 (0-3) y día 9 (4-9). Si cae sábado/domingo, pasa al lunes. Sin feriados (supuesto).

"Terminación de CUIT" = último dígito **antes** del dígito verificador. Ej. `20-12345678-9` → terminación `8`.

### 2.4 Tipo de cambio — Dólar oficial BNA (billete), octubre 2026

Fuente: dolarhistorico.com. Solo días hábiles; para un día sin cotización se usa el **día hábil anterior** con cotización.

| Fecha | Compra | Venta |
|---|---|---|
| 2026-10-01 | 1486.11 | 1539.87 |
| 2026-10-02 | 1489.80 | 1543.18 |
| 2026-10-05 | 1487.15 | 1540.24 |
| 2026-10-06 | 1488.05 | 1541.41 |
| 2026-10-07 | 1489.54 | 1543.08 |
| 2026-10-08 | 1486.60 | 1539.30 |
| 2026-10-09 | 1483.80 | 1537.18 |
| 2026-10-13 | 1482.15 | 1535.57 |
| 2026-10-14 | 1481.80 | 1535.17 |
| 2026-10-15 | 1479.86 | 1533.28 |
| 2026-10-16 | 1481.43 | 1534.80 |
| 2026-10-19 | 1487.64 | 1541.09 |
| 2026-10-20 | 1491.64 | 1545.01 |
| 2026-10-21 | 1492.78 | 1546.15 |
| 2026-10-22 | 1496.30 | 1549.75 |
| 2026-10-23 | 1496.32 | 1549.77 |
| 2026-10-26 | 1507.26 | 1560.40 |
| 2026-10-27 | 1506.87 | 1559.94 |
| 2026-10-28 | 1507.05 | 1560.12 |
| 2026-10-29 | 1507.68 | 1560.70 |
| 2026-10-30 | 1509.25 | 1561.96 |

Los meses **octubre 2025 a septiembre 2026** (octubre 2025 agregado porque el historial de movimientos del seed arranca el 2025-10-01) y **noviembre 2026 (1 al 13)** se completan en la feature `001` desde
`https://dolarhistorico.com/cotizacion-dolar-oficial/mes/{mes}-{anio}` (ej. `septiembre-2026`), mismo formato. Hacen falta para pesificar el historial de 12 meses.

---

## 3. Modelo de dinero

- **Nunca `float64` para dinero.** `Money{Amount int64 /* unidades menores (centavos) */, Currency "ARS"|"USD"}`.
- Tipo de cambio: `Rate` con 4 decimales como entero escalado (`int64`, ×10.000). Ej. `1539.87` → `15398700`.
- Conversión USD→ARS: `ars_cents = round_half_even(usd_cents × rate_scaled / 10_000)`. Implementar con `math/big` para evitar overflow.
- Porcentajes: enteros en "basis points" (`8000` = 80,00%).
- Formato en UI: `Intl.NumberFormat('es-AR', { style: 'currency', currency })`.

---

## 4. Reglas de clasificación de movimientos

Tipos de movimiento (`MovementType`): `INBOUND_TRANSFER`, `OUTBOUND_TRANSFER`, `CARD_SPEND`, `CARD_REFUND`, `STABLECOIN_BUY`, `STABLECOIN_SELL`, `FEE`.

Contraparte (`Counterparty`): `name`, `country` (ISO-2), `kind` (`COMPANY`|`PERSON`|`PLATFORM`), `isOwnAccount` (bool).

Resultado (`ClassificationKind`): `INCOME`, `NOT_INCOME`, `UNCLASSIFIED`. Confianza: `HIGH`, `LOW`.

Reglas en orden; **la primera que matchea gana**. Cada regla tiene un código (`RuleCode`) que se muestra al contador.

| # | RuleCode | Condición | Resultado |
|---|---|---|---|
| R1 | `OWN_ACCOUNT_TRANSFER` | `counterparty.isOwnAccount == true` | `NOT_INCOME` / HIGH |
| R2 | `STABLECOIN_CONVERSION` | tipo `STABLECOIN_BUY` o `STABLECOIN_SELL` | `NOT_INCOME` / HIGH |
| R3 | `CARD_EXPENSE` | tipo `CARD_SPEND` o `CARD_REFUND` | `NOT_INCOME` / HIGH |
| R4 | `OUTBOUND` | tipo `OUTBOUND_TRANSFER` o `FEE` | `NOT_INCOME` / HIGH |
| R5 | `KNOWN_PAYER` | `INBOUND_TRANSFER` y el nombre normalizado de la contraparte coincide con el receptor de alguna factura emitida por el cliente **antes del inicio del período** | `INCOME` / HIGH |
| R6 | `FOREIGN_COMPANY` | `INBOUND_TRANSFER`, `country != "AR"`, `kind ∈ {COMPANY, PLATFORM}` | `INCOME` / HIGH |
| R7 | `FOREIGN_PERSON` | `INBOUND_TRANSFER`, `country != "AR"`, `kind == PERSON` | `INCOME` / LOW |
| R8 | `LOCAL_INBOUND` | `INBOUND_TRANSFER`, `country == "AR"` | `UNCLASSIFIED` / LOW |
| R9 | `FALLBACK` | cualquier otro | `UNCLASSIFIED` / LOW |

- El contador puede **sobrescribir** una clasificación (`MANUAL_OVERRIDE`, guarda quién y cuándo). La sobrescritura gana sobre las reglas y dispara recálculo del cierre.
- Pesificación de movimientos: usar el **tipo de cambio de la operación** que trae Wallbit.

---

## 5. Pesificación de facturas

1. Si la factura trae `fxRate` (columna "Tipo Cambio" del CSV), se usa ese.
2. Si no, se pide al `ExchangeRateProvider` según el parámetro `fx_rule`:
   - `BNA_BUY_PREV_BUSINESS_DAY` (default): compra del día hábil anterior a la emisión.
   - `BNA_SELL_PREV_BUSINESS_DAY`: venta del día hábil anterior.
   - `BNA_BUY_SAME_DAY`: compra del mismo día (o día hábil anterior si no hay).
3. Facturas en ARS: rate = 1.

> El criterio exacto lo valida el contador; por eso es parámetro.

---

## 6. Conciliación cobros ↔ facturas

- Universo: movimientos `INCOME` del período y facturas emitidas del período **más** facturas de hasta `reconciliation_window_before_days` antes que sigan sin conciliar.
- Un movimiento matchea con una factura si:
  - misma moneda;
  - `|invoice.amount − movement.amount| ≤ invoice.amount × reconciliation_tolerance_bp / 10.000` (cubre comisiones de plataformas);
  - `movement.date ∈ [invoice.date − reconciliation_window_after_days, invoice.date + reconciliation_window_before_days]`. El cobro puede ser antes (anticipo) o después de la factura;
  - el nombre normalizado de la contraparte (minúsculas, sin acentos, sin sufijos `inc|llc|ltd|sa|srl|gmbh|corp|co`, sin puntuación) es igual al del receptor de la factura, o uno contiene al otro.
- Matching **1 a 1, greedy**: por cada movimiento (orden por fecha asc) elegir la factura candidata con menor diferencia de fecha; empate → menor diferencia de monto.
- Resultado: `matched[]`, `unmatchedIncome[]` (cobros sin factura), `unmatchedInvoices[]` (facturas sin cobro).

---

## 7. Liquidación (Assessment)

### 7.1 Monotributo

Entradas: facturas (tipos `C` y `E`; `E` solo si `invoice_e_counts_for_cap == true`), escala vigente, categoría actual del cliente, parámetros.

- `billed12m` = suma ARS de facturas con fecha en `(periodEnd − 12 meses, periodEnd]`.
- `scale` = escala vigente a `periodEnd`.
- `requiredCategory` = menor categoría con `cap ≥ billed12m`. Si `billed12m > cap(K)` → `requiredCategory = null` y `excluded = true`.
- `pctOfCurrentCapBp` = `billed12m / cap(currentCategory)` en bp. `pctOfKCapBp` = `billed12m / cap(K)` en bp.
- `fee` = cuota servicios de la **categoría actual** según la escala vigente a la fecha de vencimiento. `feeDueDate` = vencimiento Monotributo del mes siguiente al período.
- `exclusionDate` (solo si `excluded`): primer día `d` en los 365 días que terminan en `periodEnd` tal que la suma de facturas en `(d − 12 meses, d]` supera el `cap(K)` de la escala vigente en `d`. Usar sumas de prefijos por día.
- `canReturnToMonotributoFrom` = 1 de enero del año `exclusionDate.year + 4` (3 años calendario completos posteriores; interpretación a validar por el contador).
- `billedInPeriod` = facturas del período (USD y ARS).

### 7.2 Responsable Inscripto

- `exportSales` = facturas `E` del período (IVA 0).
- `localSalesNet`, `outputVat` = neto e IVA de facturas `A`/`B` del período.
- `inputVat` = IVA de **facturas de compra** (recibidas) del período, cargadas por el cliente.
- `vatBalance = outputVat − inputVat − previousVatCredit`. Si `< 0` → `vatToPay = 0`, `newVatCredit = −vatBalance`; si no, `vatToPay = vatBalance`, `newVatCredit = 0`.
- `selfEmployedContribution` (autónomos) = parámetro mensual del cliente cargado por el contador.
- `incomeTaxAdvance` = anticipo del período si el contador lo cargó (`{period, amount, dueDate}`), sino 0.
- Vencimientos: IVA y autónomos según terminación de CUIT (§2.3).
- Fuera de alcance: reintegro de IVA de exportadores, retenciones/percepciones, DDJJ anual.

### 7.3 Transición Monotributo → RI

- Detectada la exclusión, el contador confirma "Pase a RI" con fecha efectiva = `exclusionDate`. Se registra en `regimeHistory`.
- Desde la fecha efectiva, el cierre usa §7.2.
- Facturas `C` emitidas con fecha ≥ fecha efectiva → señal `POST_EXCLUSION_C_INVOICES`.

### 7.4 Controles adicionales (solo Monotributo)

- `inbound12m` = suma ARS de `INBOUND_TRANSFER` con `isOwnAccount == false` en 12 meses. Si `> cap(K)` → `DEPOSITS_EXCEED_K_CAP`.
- `cardSpend12m` = suma ARS de `CARD_SPEND` − `CARD_REFUND` en 12 meses. Si `> cap(K) × card_spend_k_cap_bp / 10.000` → `HIGH_CARD_SPENDING`.

### 7.5 Cobro atípico

- Historial: movimientos `INCOME` individuales de los `atypical_window_months` meses anteriores al período.
- Si hay ≥ 3 movimientos históricos: `median` = mediana de montos USD. Cada `INCOME` del período con `amount > median × atypical_multiplier_bp / 10.000` → `ATYPICAL_INCOME`.

---

## 8. Catálogo de señales

Severidad y prioridad fijas. El **semáforo** del cierre = máxima severidad. La **señal principal** = la de mayor severidad y, a igual severidad, menor número de prioridad. La descripción de la tabla es el template de la señal principal; las demás se muestran como chips (`+N`).

| Prio | Code | Severity | Régimen | Disparador | Template (es-AR) |
|---|---|---|---|---|---|
| 1 | `MONO_EXCLUSION_K_CAP_EXCEEDED` | RED | MONO | §7.1 `excluded` | "Superó el tope de la categoría K el {exclusionDate} por {excessARS}. Corresponde pase a Responsable Inscripto." |
| 2 | `POST_EXCLUSION_C_INVOICES` | RED | RI | §7.3 | "{count} facturas C emitidas después de la exclusión ({totalARS})." |
| 3 | `MISSING_INVOICES` | RED | ambos | cobros sin factura con total ≥ `missing_invoice_min_usd` | "{count} cobros por {totalUSD} sin factura asociada." |
| 4 | `DEPOSITS_EXCEED_K_CAP` | RED | MONO | §7.4 | "Depósitos de los últimos 12 meses ({inboundARS}) superan el tope de la categoría K." |
| 5 | `MONO_NEAR_K_CAP` | YELLOW | MONO | `pctOfKCapBp ≥ k_cap_warning_bp` y no excluido | "Acumulado 12 meses al {pct} del tope de la categoría K." |
| 6 | `MONO_CATEGORY_CHANGE_REQUIRED` | YELLOW | MONO | `requiredCategory > currentCategory` | "Facturó {billed12m} en 12 meses: supera el tope de la {current}. Recategorizar a {required}." |
| 7 | `MONO_NEAR_CATEGORY_CAP` | YELLOW | MONO | `pctOfCurrentCapBp ≥ category_cap_warning_bp` y sin cambio requerido | "Acumulado al {pct} del tope de la categoría {current}." |
| 8 | `ATYPICAL_INCOME` | YELLOW | ambos | §7.5 | "Cobro de {amountUSD}, {ratio}× su mediana de {window} meses." |
| 9 | `UNCLASSIFIED_MOVEMENTS` | YELLOW | ambos | movimientos `UNCLASSIFIED` en el período | "{count} movimientos sin clasificar ({totalUSD})." |
| 10 | `EXTERNAL_INCOME_DECLARED` | YELLOW | ambos | el cliente declaró cobros fuera de Wallbit | "Declaró cobros fuera de Wallbit por {totalUSD}." |
| 11 | `INVOICE_WITHOUT_PAYMENT` | YELLOW | ambos | facturas sin cobro conciliado | "{count} facturas por {totalUSD} sin cobro en Wallbit." |
| 12 | `HIGH_CARD_SPENDING` | YELLOW | MONO | §7.4 | "Gastos con tarjeta en 12 meses al {pct} del tope de la categoría K." |
| 13 | `MONO_CATEGORY_DOWNGRADE_POSSIBLE` | YELLOW | MONO | `requiredCategory < currentCategory` | "Podría bajar a la categoría {required} en la próxima recategorización." |
| — | (ninguna) | GREEN | — | — | Monotributo: "Facturó {billedUSD} · cuota {fee} · todo conciliado." · RI: "Facturó {billedUSD} · IVA {vatToPay} · todo conciliado." |

Formato en templates: fechas `dd/MM/yyyy`; ARS `$ 1.234.567,89`; USD `US$ 1.234,56`; porcentajes `87%`; ratio con 1 decimal (`3,1×`).

---

## 9. Parámetros del contador (defaults)

| Clave | Tipo | Default | Descripción |
|---|---|---|---|
| `invoice_e_counts_for_cap` | bool | `true` | Si la factura E suma al acumulado para categoría/tope. Las fuentes no coinciden: decisión del contador. |
| `fx_rule` | enum | `BNA_BUY_PREV_BUSINESS_DAY` | §5. |
| `category_cap_warning_bp` | int | `8000` | Umbral "cerca del tope de su categoría". |
| `k_cap_warning_bp` | int | `8000` | Umbral "cerca del tope de la K". |
| `atypical_multiplier_bp` | int | `25000` | 2,5× la mediana. |
| `atypical_window_months` | int | `6` | Ventana del histórico. |
| `missing_invoice_min_usd` | int (centavos) | `5000` | US$ 50 mínimo para disparar `MISSING_INVOICES`. |
| `reconciliation_tolerance_bp` | int | `300` | 3% de tolerancia de monto. |
| `reconciliation_window_before_days` | int | `45` | Cobro hasta 45 días después de la factura. |
| `reconciliation_window_after_days` | int | `10` | Cobro hasta 10 días antes de la factura. |
| `card_spend_k_cap_bp` | int | `4000` | 40% del tope K. |
| `reminder_days_before_due` | int | `3` | Días antes del vencimiento para el recordatorio. |

Por cliente (solo RI), cargados por el contador: `selfEmployedMonthlyContribution` (ARS), `incomeTaxAdvances[] {period, amount, dueDate}`, `previousVatCredit` (ARS).

Cambiar un parámetro global **recalcula todos los cierres no aprobados** del período abierto y genera nueva versión de cálculo (`calcVersion`).

---

## 10. Estados del cierre

```
             (job día 1)
                 │
                 ▼
     ┌──── PENDING_CLIENT_INPUT ────┐  cliente confirma carga
     │                              ▼
     │                      READY_FOR_REVIEW ◄────────────┐
     │  (contador puede              │ aprobar             │ cambio en datos
     │   revisar igual)              ▼                     │ (factura nueva,
     └────────────────────────►   APPROVED                 │  override, params)
                                     │ outbox despacha     │
                                     ▼                     │
                                  NOTIFIED ────────────────┘
```

- `PENDING_CLIENT_INPUT`: el cliente aún no confirmó "Ya cargué todo". El contador puede verlo y aprobarlo (decisión del contador), pero la aprobación en lote lo excluye.
- Cualquier cambio de datos sobre un cierre `APPROVED`/`NOTIFIED` genera **nueva versión**, vuelve a `READY_FOR_REVIEW` y conserva el historial de aprobaciones.
- `paidAt` (el cliente marca "Ya pagué") es independiente del estado.

---

## 11. Fuentes

- Escala desde agosto 2026: https://hacecuentas.com/datos-monotributo-2026 · https://www.iprofesional.com/impuestos/462480-monotributo-como-quedan-los-topes-de-facturacion-y-los-montos-de-agosto · https://www.cotizacion-dolar.com.ar/economia/categorias-monotributo-agosto-2026.html
- Escala febrero 2026: https://www.iprofesional.com/impuestos/452037-arca-cual-es-la-fecha-limite-para-pagar-el-monotributo-en-abril-2026
- Exclusión y 3 años: https://www.afip.gob.ar/monotributo/ayuda/exclusion.asp · https://yo-facturo.com/blog/exclusion-del-monotributo/
- Factura E: https://tributosimple.com/como-emitir-una-factura-e/ · https://contablix.ar/blog/monotributo-freelancers-digitales-instructores-2026 (dice que E no computa) vs https://www.iprofesional.com/impuestos/248968 (pesificado computa) → parámetro.
- Vencimientos noviembre 2026: https://www.econoblog.com.ar/?p=160745 · https://calcularsueldo.com.ar/?p=16367 · https://servidos.ar/calendario-vencimientos-arca
- Tipo de cambio: https://dolarhistorico.com/cotizacion-dolar-oficial/mes/octubre-2026
