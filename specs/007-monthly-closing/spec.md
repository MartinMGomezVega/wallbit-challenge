# 007 — Cierre mensual: job, estados, recálculo y API de cierres

## Objetivo
Orquestar todo lo anterior: el día 1 (o al arrancar el prototipo) se genera el cierre de cada cliente para el mes anterior, en paralelo y de forma idempotente. Cualquier cambio de datos recalcula. El contador consulta cierres filtrados y ordenados.

## Historias
- Como contador, cuando arranca el mes ya tengo los cierres de todos mis clientes armados, sin hacer nada.
- Como contador, si corrijo una clasificación o el cliente sube una factura, el cierre se actualiza solo.
- Como contador, filtro y ordeno cientos de cierres en una sola tabla.

## Criterios de aceptación
### Cálculo de un cierre
1. `usecase.ComputeClosing.Execute(ctx, clientID, period)`:
   1. carga cliente, movimientos (13 meses + período), clasificaciones (override gana), facturas (13 meses), facturas de compra, input del período, parámetros;
   2. `knownPayers` = receptores de facturas anteriores al período; clasifica movimientos sin override;
   3. pesifica facturas (`AmountARS`) usando el `ExchangeRateProvider`;
   4. concilia;
   5. liquida según el régimen vigente al último día del período (004 o 005), calculando los vencimientos del mes siguiente al período (`p.Next()`);
   6. evalúa señales (006);
   7. persiste `Closing` con `Version+1` si el resultado cambió (comparar `ResultHash`); si no cambió, no crea versión.
2. `Closing{ID, ClientID, Period, Status, Version, ResultHash, Severity, Signals, Description, Mono?, RI?, Reconciliation (IDs), ComputedAt, Approvals[] , PaidAt?}`. `ResultHash` = SHA-256 de JSON canónico de assessment + señales (orden de claves estable).
3. Estado inicial: `READY_FOR_REVIEW` si el input del período está confirmado; si no, `PENDING_CLIENT_INPUT`. Transiciones según `domain-context.md §10`; transición inválida → `ErrInvalidTransition`.
4. Si un cierre `APPROVED`/`NOTIFIED` cambia de `ResultHash` → nueva versión, `Status = READY_FOR_REVIEW`, las aprobaciones anteriores se conservan (con su versión).

### Job
5. `usecase.RunMonthlyClosing.Execute(ctx, period)` procesa todos los clientes con un worker pool (`CLOSING_WORKERS`), devuelve `{processed, created, updated, unchanged, failed[]{clientId, error}}`. Un error en un cliente no frena al resto.
6. Idempotente: correrlo dos veces seguidas → la segunda `created=0, updated=0`.
7. Se ejecuta al arrancar: para los datos semilla se procesan 2026-09 (cerrado, aprobado y notificado para habilitar métricas históricas) y 2026-10 (mes en curso para revisión). También `POST /api/v1/accountant/periods/{period}/run` (devuelve el resumen).
8. Scheduler: goroutine que, cada minuto, si `clock.Now()` es día 1 y no corrió ese período, lo corre (con clock fijo nunca dispara de nuevo; test con fake clock).
9. Test de carrera: `go test -race` corriendo el job mientras se hacen overrides concurrentes.
10. Performance: 300 clientes en < 2 s en la máquina de dev (test con `testing.Short()` skip para CI lento).

### Recálculo por cambios
11. `ChangeNotifier` real: al recibir `ClosingInputsChanged{ClientID, Period}` recalcula **sincrónicamente** ese cierre (y el del período siguiente si existe). Los handlers que escriben devuelven ya con el cierre actualizado.
12. `PUT /parameters` recalcula todos los cierres **no aprobados** del último período cerrado y responde `{ "parameters": {...}, "recalculated": N }`.
13. Confirmar input del cliente (`/me/periods/{p}/confirm`) mueve `PENDING_CLIENT_INPUT → READY_FOR_REVIEW`.

### API de consulta (contador)
14. `GET /api/v1/accountant/periods/{period}/closings` con query:
    - `severity` (CSV de `RED,YELLOW,GREEN`), `signal` (CSV de códigos; match si el cierre tiene **alguno**), `regime` (`MONOTRIBUTO|RI`), `status` (CSV), `q` (nombre o CUIT, normalizado, substring);
    - `sort` ∈ `severity` (default), `name`, `billed`, `toPay`, `pctKCap`; `dir` ∈ `asc|desc`;
    - `page`, `pageSize` (default 50, máx 500).
    - Orden `severity` = severidad desc, prioridad de la señal principal asc, nombre asc. **Siempre** desempatar por `clientId` para orden total estable.
    - Fila: `{closingId, clientId, clientName, cuit, regime, category, severity, primarySignal{code,label}?, description, otherSignals[{code,label,severity}], billedInPeriod{usd, ars}, accumulated12m{ars, pctOfKCapBp}, toPay{amount, dueDate}, status, version, paidAt}`. `toPay` MONO = fee; RI = `TotalToPay` con el vencimiento más cercano.
15. `GET /api/v1/accountant/periods/{period}/closings/summary` con los mismos filtros **excepto** `severity` → `{total, bySeverity{RED,YELLOW,GREEN}, bySignal{code: n}, byStatus{...}}` (para los chips con contador).
16. `GET /api/v1/accountant/closings/{id}` → detalle: cliente, assessment, señales con `data`, movimientos del período con clasificación (regla, confianza, override), facturas del período con ARS y TC usado, pares conciliados, cobros/facturas sin conciliar, input del cliente, historial de versiones y aprobaciones. (El borrador del aviso lo agrega 010.)
17. Test end-to-end con el seed: para cada `cli_demo_XX` la señal principal y el estado coinciden con la tabla de `002/plan.md`. Y la distribución global de semáforos queda aprox. en 70–80% verde, 15–22% amarillo, 5–10% rojo (si no, ajustar el seed, no las reglas).
