# 002 — Clientes, movimientos, clasificación y datos semilla

## Objetivo
Modelar clientes y movimientos de Wallbit, clasificar automáticamente cada movimiento con reglas determinísticas (`domain-context.md §4`), permitir la corrección del contador y generar datos semilla realistas para ~300 clientes, incluyendo **clientes de demo** con escenarios fijos.

## Historias
- Como sistema, clasifico cada movimiento como ingreso / no ingreso / sin clasificar, con regla y confianza.
- Como contador, puedo corregir una clasificación y queda registrado quién y cuándo.
- Como evaluador de la demo, encuentro al menos un cliente por cada señal del catálogo.

## Criterios de aceptación
### Modelo
1. `Client{ID, Name, Email, CUIT, Regime (MONOTRIBUTO|RI), MonotributoCategory (si MONO), RegimeHistory[]{Regime, From, To?, Reason}, RIProfile (si RI): {SelfEmployedMonthlyContribution, IncomeTaxAdvances[], PreviousVatCredit}}`.
2. `Movement{ID, ClientID, Type, Amount (USD), FXRate (Wallbit, de la operación), OccurredAt, Counterparty, Description}`.
3. `Classification{MovementID, Kind, Confidence, Rule, OverriddenBy?, OverriddenAt?}`.

### Clasificación
4. `movement.Classify(m, knownPayers set[string])` aplica R1..R9 en orden; tests: **un caso por regla** + caso donde R1 gana sobre R6 (transferencia propia desde el exterior).
5. `knownPayers` = nombres normalizados de receptores de facturas emitidas por el cliente antes del inicio del período (lo arma 003; en esta feature se recibe como parámetro y la clasificación inicial al arrancar usa un set vacío; 007 reclasifica con el set real).
6. Override: `PATCH /api/v1/accountant/movements/{id}/classification` body `{ "kind": "INCOME"|"NOT_INCOME", "note": "..." }` → `200` con la clasificación (`rule: "MANUAL_OVERRIDE"`, `overriddenBy: "accountant"`). `kind: "UNCLASSIFIED"` → 422. Movimiento inexistente → 404. (El recálculo del cierre se conecta en 007.)

### Seed
7. Determinístico: misma `SEED` y `SEED_CLIENTS` ⇒ mismos datos (test compara hash del dataset).
8. Historial de **13 meses**: 2025-10-01 a 2026-10-31, más movimientos 2026-11-01 a `APP_NOW`.
9. Distribución aproximada (tolerancia ±5 pp, test): 85% Monotributo, 15% RI; ingresos mensuales USD 800–9.000 con la mayoría entre 1.500 y 4.500.
10. Los **15 clientes de demo** existen con IDs fijos `cli_demo_01..15` y los datos que producen exactamente las señales de la tabla del plan (la verificación end-to-end está en 007).
11. CUITs válidos (dígito verificador), terminaciones distribuidas 0–9.
12. `FXRate` de cada movimiento = compra BNA del día (mock fx) — supuesto: Wallbit liquida a compra BNA.

## Fuera de alcance
Facturas (003). Cálculo impositivo (004/005).
