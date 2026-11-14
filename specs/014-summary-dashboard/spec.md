# 014 — Resumen del cierre (dashboard)

## Objetivo
Una pestaña **"Resumen"** al lado de "Cierres" que responde en 10 segundos: ¿cuánto falta para terminar el cierre?, ¿cuánto facturaron y cuánto tienen que pagar mis clientes?, ¿dónde está el riesgo?, y ¿está funcionando el servicio? (las métricas de éxito de la nota de decisiones).

No es un dashboard decorativo: **cada número lleva a la tabla filtrada** que lo explica.

## Criterios de aceptación
### Backend
1. `GET /api/v1/accountant/periods/{period}/stats` →
   ```json
   {
     "period": "2026-10",
     "progress": {
       "total": 300,
       "byStatus": { "PENDING_CLIENT_INPUT": 30, "READY_FOR_REVIEW": 190, "APPROVED": 0, "NOTIFIED": 80 },
       "approvedPct": 2666,
       "pendingReview": { "RED": 21, "YELLOW": 56, "GREEN": 143 },
       "estimatedReviewMinutes": 1056
     },
     "totals": {
       "billedUSD": { "amount": "...", "currency": "USD" },
       "billedARS": { "amount": "...", "currency": "ARS" },
       "toPayARS": { "amount": "...", "currency": "ARS" },
       "toPayByDueDate": [ { "date": "2026-11-20", "amount": {...}, "clients": 255 } ]
     },
     "risk": {
       "bySeverity": { "RED": 21, "YELLOW": 56, "GREEN": 223 },
       "bySignal": [ { "code": "MISSING_INVOICES", "label": "Faltan facturas", "severity": "RED", "count": 14 } ],
       "byCategory": [ { "category": "A", "clients": 12 }, ..., { "category": "RI", "clients": 45 } ],
       "nearKCap": 9
     },
     "success": {
       "greenPct": 7433,
       "notifiedBeforeDay10Pct": 2666,
       "paidOnTimePct": 6250
     }
   }
   ```
2. Definiciones (porcentajes en bp, redondeo hacia abajo; denominador 0 → 0):
   - `approvedPct` = (`APPROVED` + `NOTIFIED`) / total.
   - `pendingReview` = cierres **no aprobados** por severidad.
   - `estimatedReviewMinutes` = verdes pendientes × 2 + (amarillos + rojos) pendientes × 10 (constantes documentadas en README; mismas de la cuenta de dimensionamiento).
   - `toPayARS` = suma de `toPay` de todos los cierres del período. `toPayByDueDate` agrupa por vencimiento.
   - `bySignal` ordenado por severidad desc y `count` desc; cuenta cierres (no señales repetidas).
   - `byCategory`: categoría actual para Monotributo (A–K, siempre las 11 aunque estén en 0) + una fila `RI`.
   - `nearKCap` = cierres con `MONO_NEAR_K_CAP` o `MONO_EXCLUSION_K_CAP_EXCEEDED`.
   - `greenPct` = verdes / total.
   - `notifiedBeforeDay10Pct` = cierres cuya **primera** publicación fue ≤ día 10 del mes siguiente al período / total.
   - `paidOnTimePct` = cierres notificados con `paidAt` ≤ su vencimiento más cercano / cierres notificados.
3. Puerto `StatsQuery.PeriodStats(ctx, period)`: implementación en memoria (Go) y en Postgres (SQL con agregaciones). **Misma suite de contrato** para ambas (013).
4. Tests: cada definición del criterio 2 con casos borde (período vacío, todos verdes, sin notificados).

### Frontend
5. Ruta `/contador/resumen`; navegación del header: **Cierres · Resumen · Parámetros**. Respeta `period` en la URL.
6. **Fila de 4 KPIs**:
   - **Avance del cierre:** "220 de 300 sin aprobar" + barra apilada por estado + "≈ 17 h 36 min de revisión".
   - **Facturado en octubre:** US$ grande, ARS chico.
   - **A pagar por tus clientes:** ARS grande + "próximo vencimiento 20/11".
   - **Todo ok:** % verdes con el ícono verde.
7. **Tres gráficos**:
   - **Situaciones más frecuentes:** barras horizontales coloreadas por severidad (top 8).
   - **Clientes por categoría:** barras verticales A–K + RI, en `--color-primary`; las barras I, J, K se marcan con borde amarillo si hay clientes cerca del tope K.
   - **Vencimientos:** barras por fecha con el monto a pagar.
8. **Métricas de éxito** (fila de 3 tarjetas chicas con texto explicativo): % verdes, % avisados antes del día 10, % de cuotas pagadas en término.
9. **Drill-down:** click en una barra de situación → `/contador?period=…&signal=CODE`; click en un segmento de la barra de avance → `/contador?status=…`; click en "Todo ok" → `/contador?severity=GREEN`.
10. Estados de carga (skeletons con la forma de cada tarjeta), vacío y error (§5 de `ui-guidelines.md`).
11. Accesibilidad: cada gráfico tiene título, `aria-label` con resumen textual ("Faltan facturas: 14 clientes; …") y una tabla equivalente oculta visualmente (`sr-only`).
12. Tests: render de KPIs y formatos, drill-down navega con los filtros correctos, estados, `vitest-axe` sin violaciones serias.
