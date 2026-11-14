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
       "greenCount": 223,
       "total": 300,
       "notifiedBeforeDay10Pct": 2666,
       "notifiedBeforeDay10Count": 80,
       "notifiedCount": 80,
       "paidOnTimePct": 6250,
       "paidOnTimeCount": 50
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
   - `greenPct` = verdes / total (`greenCount`, `total`).
   - `notifiedBeforeDay10Pct` = cierres cuya **primera** publicación fue ≤ día 10 del mes siguiente al período / total (`notifiedBeforeDay10Count`, `total`).
   - `paidOnTimePct` = cierres notificados con `paidAt` ≤ su vencimiento más cercano / cierres notificados (`paidOnTimeCount`, `notifiedCount`).
3. Puerto `StatsQuery.PeriodStats(ctx, period)`: implementación en memoria (Go) y en Postgres (SQL con agregaciones). **Misma suite de contrato** para ambas (013).
4. Tests: cada definición del criterio 2 con casos borde (período vacío, todos verdes, sin notificados).
5. `GET /api/v1/accountant/periods` devuelve `{ "periods": ["2026-10", "2026-09"] }` con los períodos disponibles ordenados descendente. El selector `PeriodSelect` del header consume este endpoint dinámicamente.

### Frontend
6. Ruta `/contador/resumen`; navegación del header: **Cierres · Resumen · Parámetros**. Título `"Resumen de octubre 2026"` (mes en minúscula, sin `capitalize`), `<title>`: `"Resumen · Wallbit Contabilidad"`. Lee `period` de la URL (controlado por el `PeriodSelect` global del header; no renderiza selector propio). En datos semilla, seleccionar septiembre muestra el período cerrado con las métricas de éxito históricas (`notifiedBeforeDay10Pct > 0` y `paidOnTimePct > 0`).
7. **Orden de la página:**
   1. **Estado de los clientes (`SeverityOverview`):** 3 tarjetas (**Rojos · Amarillos · Verdes**), cada una con ícono y borde izquierdo de 4 px de su severidad, cantidad grande, `"% del total"` y link `"Ver en la tabla →"` (`/contador?period=…&severity=RED|YELLOW|GREEN`).
   2. **Indicadores del cierre (3 KPIs):**
      - **Progreso del cierre:** "220 de 300 sin aprobar" + barra apilada por estado + "≈ 17 h 36 min de revisión".
      - **Facturado del mes:** US$ grande, ARS chico.
      - **A pagar estimado:** ARS grande + "Próximo vence 20/11".
   3. **Tres gráficos:**
      - **Situaciones detectadas:** barras horizontales coloreadas por severidad (top 8).
      - **Distribución por categoría:** barras verticales A–K + RI, en `--color-primary`; las barras I, J, K se marcan con borde amarillo si hay clientes cerca del tope K.
      - **Vencimientos próximos:** barras por fecha con el monto a pagar.
   4. **Sección separada "¿Está funcionando el servicio?" (`ServiceHealth`):** separador + título `text-lg` + subtítulo `"Las métricas con las que mido si el servicio cumple su objetivo."` y 3 tarjetas: Clientes en verde, Avisados antes del día 10, Pagos en término. Cada una muestra el porcentaje grande y debajo `"X de Y"` con su definición en una línea.
7. **Drill-down:** click en `"Ver en la tabla →"` → `/contador?period=…&severity=…`; click en una barra de situación → `/contador?period=…&signal=CODE`; click en un segmento de la barra de avance → `/contador?status=…`.
8. Estados de carga (skeletons con la forma de cada tarjeta), vacío y error (§5 de `ui-guidelines.md`).
9. Accesibilidad: cada gráfico tiene título, `aria-label` con resumen textual ("Faltan facturas: 14 clientes; …") y una tabla equivalente oculta visualmente (`sr-only`).
10. Tests: orden de secciones, links de las tarjetas de estado, ausencia de "Todo ok", `"X de Y"` correcto, drill-down, estados, `vitest-axe` sin violaciones serias.
