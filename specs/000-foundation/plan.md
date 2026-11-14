# 000 — Plan

## Backend
- `go mod init github.com/<user>/wallbit-accountant/backend`.
- `internal/infrastructure/clock`: `type Fixed struct{ t time.Time }`, `type Real struct{}`; ambos implementan `ports.Clock { Now() time.Time }`. `Location()` siempre `America/Argentina/Buenos_Aires` (cargar con `time.LoadLocation`; fallback `time.FixedZone("ART", -3*3600)` si no hay tzdata; importar `time/tzdata` en `main`).
- `internal/application/ports/clock.go`, `idgen.go` (`IDGenerator{ New(prefix string) string }`; impl `infrastructure/idgen` secuencial `prefix_000001` y otra con `crypto/rand`).
- `internal/interfaces/http/`
  - `server.go`: `NewRouter(deps Deps) http.Handler`.
  - `middleware.go`: `Recover`, `RequestLog`, `CORS`, `DemoIdentity`, `RequireAccountant`, `RequireClient`.
  - `errors.go`: `writeError(w, err)` con el mapeo de `architecture.md §2.3`; `writeJSON(w, status, v)`.
  - `identity.go`: `type Identity struct{ Role string; ClientID string }`, `FromContext(ctx)`.
- `internal/config`: lee env con defaults (`architecture.md §2.4`); `Load() (Config, error)`; valida `APP_NOW` (RFC3339).
- `internal/archtest/arch_test.go`: recorre `internal/` con `go/parser` (ImportsOnly) y aplica las reglas; busca el token `float64` con `go/scanner` en `internal/domain`.
- `cmd/api/main.go`: config → clock → router → `http.Server` con `ReadHeaderTimeout`, shutdown graceful con `signal.NotifyContext`.
- `scripts/coverage.sh`: `go test ./... -race -coverprofile=coverage.out`, filtra líneas de `/cmd/` del perfil, `go tool cover -func` y compara total con 90.

## Frontend
- `npm create vite@latest frontend -- --template react-ts`.
- Deps: `@tanstack/react-query @tanstack/react-table react-router-dom zod @radix-ui/react-dialog @radix-ui/react-tooltip @radix-ui/react-checkbox @radix-ui/react-dropdown-menu clsx`.
- Dev deps: `tailwindcss postcss autoprefixer vitest @vitest/coverage-v8 @testing-library/react @testing-library/user-event @testing-library/jest-dom jsdom msw eslint typescript-eslint`.
- `vite.config.ts`: proxy `/api` → `http://localhost:8080`; bloque `test` con `environment: 'jsdom'`, `setupFiles: ['src/test/setup.ts']`, coverage thresholds 90 en las cuatro métricas y excludes de `architecture.md §3.1`.
- `src/shared/api/http.ts`: `apiFetch<T>(path, schema: ZodType<T>, init?)` → agrega `X-Demo-User` desde `roleStore`, parsea error estándar a `ApiError {code, message, fields}`.
- `src/app/roleStore.ts`: `getRole()/setRole()` con `localStorage` en try/catch y fallback en memoria; evento para re-render (`useSyncExternalStore`).
- `src/app/RoleSwitcher.tsx`: dropdown "Ver como: Contador | Cliente…". La lista de clientes se completa en la feature 011; por ahora solo Contador y un input libre de `clientId`.
- `src/app/router.tsx`: `/` → redirect según rol; `/contador`; `/contador/parametros`; `/cliente`.

## Raíz
- `Makefile` con targets de `architecture.md §4`. `make dev` usa `trap 'kill 0' EXIT` y corre ambos con `&`.
- `.editorconfig`, `.gitignore` (node_modules, coverage, dist, coverage.out).
