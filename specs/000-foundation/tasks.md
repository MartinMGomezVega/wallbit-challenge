# 000 — Tasks

> Marcá `[x]` al terminar. Cada tarea tiene un "Listo cuando" verificable.

- [x] **T000.1** Crear estructura de carpetas del monorepo (`architecture.md §1, §2.2, §3.2`) con `.gitkeep` donde haga falta, `.gitignore`, `.editorconfig`.
  - Listo cuando: `tree -L 3` coincide con la arquitectura.
- [x] **T000.2** `go mod init`, `internal/config` con `Load()` y tests (defaults, override por env, `APP_NOW` inválido → error).
  - Listo cuando: `go test ./internal/config/...` pasa.
- [x] **T000.3** `ports.Clock`, `ports.IDGenerator` + implementaciones en `infrastructure/clock` y `infrastructure/idgen`, con tests.
- [x] **T000.4** Paquete `interfaces/http`: `writeJSON`, `writeError` (mapeo de errores), `Identity`, middlewares (`Recover`, `RequestLog`, `CORS`, `DemoIdentity`, `RequireAccountant`, `RequireClient`). Tests con `httptest` para cada criterio 2 y 3 de la spec.
- [x] **T000.5** `NewRouter` con `GET /api/v1/health` y rutas dummy `/api/v1/accountant/ping` y `/api/v1/me/ping` para testear permisos.
  - Listo cuando: tests de criterios 1 y 2 pasan.
- [x] **T000.6** `cmd/api/main.go` con shutdown graceful.
  - Listo cuando: `go run ./cmd/api` + `curl localhost:8080/api/v1/health` responde ok.
- [x] **T000.7** `internal/archtest` con las tres reglas (criterio 5). Agregar un caso que demuestre que la regla detecta una violación (usar un directorio `testdata` con un paquete falso).
- [x] **T000.8** `scripts/coverage.sh` (excluye `cmd/`, falla < 90).
- [x] **T000.9** Frontend: scaffold Vite + TS strict + Tailwind + ESLint + Vitest + coverage thresholds.
  - Listo cuando: `npm run build`, `npm run test -- --coverage` corren.
- [x] **T000.10** `shared/api/http.ts` + `ApiError` + tests con MSW (éxito, error estándar, error de red, schema inválido).
- [x] **T000.11** `router`, `Layout`, `PeriodSelect` global, `WallbitLogo` (SVG oficiales), favicons en `index.html` y `useDocumentTitle`. Tests: `/` redirige a `/contador`, pestaña activa, cambio de período actualiza URL y `<title>`.
- [x] **T000.12** `Makefile` (`dev`, `test`, `coverage`, `lint`).
  - Listo cuando: `make test` y `make coverage` pasan con ≥ 90% en ambos lados.
