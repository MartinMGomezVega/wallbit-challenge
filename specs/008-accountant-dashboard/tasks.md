# 008 — Tasks

- [x] **T008.1** `shared/format` (money, date, percent) + tests.
- [x] **T008.2** `shared/ui`: SeverityIndicator, SignalChip, ProgressBar, StatusBadge, Tooltip, Pagination + tests (aria).
- [x] **T008.3** `api.ts` + `schemas.ts` (zod) para list, summary, catalog + handlers MSW en `src/mocks` con un dataset chico (10 filas, los 3 colores, RI y MONO).
- [x] **T008.4** `useDashboardParams` (parse/serialize URL, defaults, reset de página) + tests.
- [x] **T008.5** `SummaryChips` + `FiltersBar` + `SignalMultiSelect` + tests.
- [x] **T008.6** `ClosingsTable` + `columns` (las 10 columnas) con orden y paginación server-side + tests.
- [x] **T008.7** Selección + `SelectionBar` (botones aún sin acción de aprobar) + tests.
- [x] **T008.8** `DashboardPage` (loading/empty/error) + ruta `/contador` + tests.
- [x] **T008.9** Backend: `PUT /parameters` devuelve `{parameters, recalculated}` (si no lo hace aún) + test.
- [x] **T008.10** `ParametersPage` + ruta + tests.
- [x] **T008.11** Probar contra el backend real (`make dev`) y ajustar. `npm run test -- --coverage` ≥ 90%.
