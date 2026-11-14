# 014 — Tasks

- [x] **T014.1** `domain/closing/stats.go` (`ComputeStats`) con tests de cada definición y bordes.
- [x] **T014.2** Puerto `StatsQuery` + implementación en memoria + suite de contrato en `ports/repotest`.
- [x] **T014.3** `first_published_at` en el cierre (memoria y migración `0003` en Postgres) al publicar por primera vez.
- [x] **T014.4** `StatsQuery` Postgres + suite de contrato en integración (mismos números que memoria con el seed).
- [x] **T014.5** Use case + handler `GET /accountant/periods/{period}/stats` + `openapi.yaml` + test de contrato de rutas.
- [x] **T014.6** Front: `api.ts`, `schemas.ts`, `useStats`, ruta `/contador/resumen`, navegación del header.
- [x] **T014.7** `SeverityOverview` arriba (3 tarjetas con link a la tabla) + 3 KPIs (avance con barra apilada clickeable, facturado, a pagar; sin "Todo ok"). Tests.
- [x] **T014.8** Gráficos (situaciones, categorías, vencimientos) con drill-down y tablas `sr-only`. Tests.
- [x] **T014.9** Sección `ServiceHealth` ("¿Está funcionando el servicio?") con numeradores/denominadores `"X de Y"` + estados de carga/vacío/error + `vitest-axe`. Coverage ≥ 90%.
- [x] **T014.10** Backend: numeradores/denominadores en `stats.success` (`greenCount`, `total`, `notifiedBeforeDay10Count`, `notifiedCount`, `paidOnTimeCount`) + tests + `openapi.yaml`.
