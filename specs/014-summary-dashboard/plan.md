# 014 — Plan

## Punto de partida (estado real del repo)
- `closing.Closing` ya tiene `Published *Published` y `Publish(...)`. Agregar `FirstPublishedAt *time.Time` que se setea **solo** la primera vez que se publica.
- `closing.ComputeSummary` ya calcula conteos: reutilizarlo dentro de `ComputeStats` donde aplique.
- Frontend: hoy hay `src/features/{closings-dashboard, closing-review, parameters}` y rutas en `src/app/router.tsx`; la navegación del header está en `src/app/Layout.tsx`. Hay que agregar la dependencia `recharts` (no está instalada).

## Backend
- `application/ports/stats.go`: `type PeriodStats struct{...}` (tipos de dominio, Money/bp) y `StatsQuery` interface.
- `domain/closing/stats.go`: función pura `ComputeStats(rows []StatsRow, firstPublishedAt map[id]time.Time, catalog, period) PeriodStats` que usa la implementación en memoria y que sirve de **oráculo** en la suite de contrato.
- `infrastructure/postgres/stats_query.go`: una consulta por bloque (progress, totals, risk, success) con `COUNT(*) FILTER`, `SUM`, `GROUP BY`. `firstPublishedAt` = `MIN(approved_at)` de aprobaciones que llevaron a `NOTIFIED` (o `published_at` de la primera versión publicada: guardar `first_published_at` en `closings` al publicar por primera vez).
- Use case `GetPeriodStats` + handler. Se agrega al `openapi.yaml`.

## Frontend
```
features/summary-dashboard/
├── index.ts, api.ts, schemas.ts, useStats.ts
├── SummaryPage.tsx
├── SeverityOverview.tsx
├── KpiCard.tsx, ProgressKpi.tsx
├── charts/SignalsBarChart.tsx, CategoryBarChart.tsx, DueDatesBarChart.tsx
├── ServiceHealth.tsx
└── __tests__/
```
- Gráficos con **Recharts** (`recharts`): sin animaciones si `prefers-reduced-motion`; `isAnimationActive={false}` en tests.
- Estilo de gráficos: `ui-guidelines.md §4.7`.
- Drill-down con `useNavigate` + `toQueryString` de `useDashboardParams` (008) para no duplicar el formato de la URL.
