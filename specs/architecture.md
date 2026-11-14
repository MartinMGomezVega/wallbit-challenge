# Architecture

> Cómo está organizado el código. Todas las features respetan esta estructura y estas convenciones.

## 1. Monorepo

```
wallbit-challenge/
├── backend/                 # Go (+ Dockerfile)
├── frontend/                # React + TS (Vite) (+ Dockerfile con nginx)
├── docs/                    # Swagger UI para GitHub Pages
├── specs/                   # este directorio (SDD)
├── docker-compose.yml       # db (Postgres) + api + web
├── Makefile                 # make up | dev | dev-memory | test | coverage | test-integration | lint
└── README.md
```

## 2. Backend (Go)

### 2.1 Stack
- Go ≥ 1.23. **Solo stdlib** salvo excepciones listadas: `net/http` con routing por patrones (`mux.HandleFunc("GET /api/v1/accountant/closings/{id}", ...)`), `log/slog`, `encoding/json`, `embed`, `math/big`, `testing`.
- Sin framework web, sin ORM. Sin dependencias externas en `domain` ni `application`.
- Dependencias permitidas: `github.com/jackc/pgx/v5` (solo en `infrastructure/postgres`, desde 013). Cualquier otra se justifica en el `plan.md` de la feature.
- Persistencia: **PostgreSQL 17** por defecto (013); repositorios en memoria para tests unitarios y `STORAGE=memory`.

### 2.2 Capas (Clean Architecture)

```
backend/
├── cmd/api/main.go                  # wiring: config → infra → use cases → http
├── internal/
│   ├── domain/                      # entidades, value objects, reglas puras. SIN imports de otras capas.
│   │   ├── money/                   # Money, Rate, conversión, bp
│   │   ├── period/                  # Period (YYYY-MM), rangos, días hábiles
│   │   ├── client/                  # Client, Regime, CUIT, RegimeHistory, RIProfile
│   │   ├── movement/                # Movement, Counterparty, Classification, reglas R1..R9
│   │   ├── invoice/                 # Invoice, InvoiceType, pesificación, parser CSV Mis Comprobantes
│   │   ├── reconciliation/          # algoritmo de conciliación
│   │   ├── tax/                     # Scale, Category, monotributo assessment, ri assessment
│   │   ├── signal/                  # Signal, Severity, catálogo, templates
│   │   ├── closing/                 # Closing, Status, máquina de estados, Approval, Version
│   │   └── parameters/              # AccountantParameters + defaults + validación
│   ├── application/                 # casos de uso + puertos (interfaces)
│   │   ├── ports/                   # repositorios, proveedores externos, clock, notifier, idgen
│   │   └── usecase/                 # un archivo por caso de uso
│   ├── infrastructure/
│   │   ├── arca/mock/               # TaxScaleProvider + DueDateProvider (go:embed JSON reales)
│   │   ├── fx/mock/                 # ExchangeRateProvider (go:embed JSON reales)
│   │   ├── memory/                  # repositorios en memoria (sync.RWMutex)
│   │   ├── postgres/                # repositorios Postgres, ClosingQuery/StatsQuery SQL, TxManager, migraciones (013)
│   │   ├── scheduler/               # job del día 1 y recordatorios
│   │   ├── notify/log/              # NotificationSender que loguea (simula email/push)
│   │   ├── clock/                   # Clock real y fijo (APP_NOW)
│   │   └── seed/                    # generador determinístico de clientes/movimientos/facturas
│   └── interfaces/http/             # handlers, DTOs, middlewares, mapeo de errores
└── go.mod                           # module github.com/<user>/wallbit-accountant/backend
```

**Regla de dependencias:** `domain` ← `application` ← (`infrastructure`, `interfaces`) ← `cmd`. Nunca al revés. Se verifica con un test (`internal/archtest`) que parsea imports con `go/parser`.

### 2.3 Convenciones
- **Errores:** errores de dominio tipados (`var ErrNotFound = errors.New(...)`, `type ValidationError struct{Field, Msg string}`). El handler mapea: NotFound→404, Validation→422, Conflict/InvalidTransition→409, Forbidden→403, resto→500.
- **Dinero:** ver `domain-context.md §3`. Prohibido `float64` en domain/application. Un test de arquitectura busca `float64` en `internal/domain` y falla si aparece.
- **Tiempo:** nunca `time.Now()` fuera de `infrastructure/clock`. Siempre `ports.Clock`.
- **IDs:** strings con prefijo: `cli_`, `mov_`, `inv_`, `clo_`, `apr_`, `ntf_`. Generados por `ports.IDGenerator` (determinístico en seed y tests).
- **Concurrencia:** repos en memoria con `sync.RWMutex`. El job de cierre procesa clientes con un **worker pool** de tamaño `CLOSING_WORKERS` (default `runtime.NumCPU()`), con `context` y cancelación. Cada cierre es **idempotente** por `(clientID, period)`: re-ejecutar no duplica.
- **Versionado de cálculo:** cada recálculo incrementa `Closing.Version`. La aprobación guarda `version` + `assessmentHash` (SHA-256 del JSON canónico del assessment + señales).
- **Logging:** `slog` JSON. Nunca loguear montos ni CUIT completos en el notificador simulado (privacidad).

### 2.4 Configuración (env)

| Var | Default | Uso |
|---|---|---|
| `PORT` | `8080` | HTTP |
| `APP_NOW` | `2026-11-05T09:00:00-03:00` | Clock fijo de la demo. Vacío = reloj real. |
| `SEED_CLIENTS` | `300` | Clientes generados. |
| `SEED` | `42` | Semilla del generador. |
| `CLOSING_WORKERS` | `NumCPU` | Workers del job. |
| `ARCA_CLIENT` | `mock` | Único valor soportado en el prototipo (`http` futuro). |
| `FX_CLIENT` | `mock` | Ídem. |
| `MOCK_LATENCY_MS` | `0` | Latencia simulada de los mocks externos. |
| `CORS_ORIGIN` | `http://localhost:5173` | Front en dev. |
| `STORAGE` | `postgres` | `postgres` o `memory`. |
| `DATABASE_URL` | `postgres://wallbit:wallbit@localhost:5432/accountant?sslmode=disable` | Conexión (credenciales solo de demo local). |
| `SEED_RESET` | `false` | `true` vacía las tablas y vuelve a sembrar al arrancar. |

### 2.5 Identidad de demo
- Header `X-Demo-User`: `accountant` o `client:<clientId>`. Middleware lo valida y lo pone en el `context`.
- Endpoints `/api/v1/accountant/**` requieren `accountant`. Endpoints `/api/v1/me/**` requieren `client:*`.

### 2.6 API
- Prefijo `/api/v1`. JSON camelCase. Fechas ISO-8601. Dinero como `{ "amount": "1234.56", "currency": "ARS" }` (string decimal para no perder precisión en JS).
- Error: `{ "error": { "code": "VALIDATION_ERROR", "message": "...", "fields": {"x": "..."} } }`.
- Listados: `?page=1&pageSize=50` → `{ "items": [...], "total": 300, "page": 1, "pageSize": 50 }`.
- El endpoint de IDs para la cola de revisión y el listado comparten **el mismo caso de uso de filtrado/orden** (garantía de mismo orden).

### 2.7 Tests backend
- `testing` stdlib, table-driven, sin mocks generados: fakes escritos a mano en `internal/application/ports/fakes`.
- Coverage: `go test ./... -coverprofile=coverage.out -covermode=atomic` → **≥ 90% total** excluyendo `cmd/` e `internal/infrastructure/postgres/`. Script `scripts/coverage.sh` que filtra esos paths y falla si < 90.
- Tests de integración (`//go:build integration`) contra el Postgres del compose: `make test-integration`. Las suites de contrato de `ports/repotest` corren contra memoria (unitarios) y contra Postgres (integración) para garantizar el mismo comportamiento.
- Tests de handlers con `httptest`.
- Test de concurrencia del job con `-race`.

## 3. Frontend (React + TS)

### 3.1 Stack
- Vite + React 18 + TypeScript `strict`.
- **TanStack Query** (data fetching, cache, prefetch), **TanStack Table** (tabla), **React Router** (estado en URL).
- **Tailwind CSS** para estilos, con los tokens de **`ui-guidelines.md`** (fuente de verdad de colores, tipografía, componentes y redacción). **Radix UI** (`@radix-ui/react-dialog`, `@radix-ui/react-tooltip`, `@radix-ui/react-checkbox`, `@radix-ui/react-dropdown-menu`, `@radix-ui/react-tabs`, `@radix-ui/react-toast`) para primitivas accesibles. **lucide-react** para íconos. **@fontsource-variable/inter** para la tipografía (self-hosted).
- **Zod** para validar respuestas de la API.
- Tests: **Vitest + Testing Library + MSW** (MSW solo en tests). Coverage `v8` **≥ 90%** líneas/branches/functions/statements (excluye `main.tsx`, `*.d.ts`, `src/test/**`, `src/mocks/**`).

### 3.2 Estructura (por feature)

```
frontend/src/
├── app/                 # router, providers (QueryClient), layout, RoleSwitcher
├── shared/
│   ├── api/             # http client (fetch + X-Demo-User + zod), schemas comunes (Money)
│   ├── ui/              # Button, ComingSoonButton, Tooltip, Badge, SeverityIndicator, Dialog, Table primitives, EmptyState
│   ├── format/          # money, date, percent (es-AR)
│   └── hooks/
├── features/
│   ├── closings-dashboard/   # tabla, filtros, chips de resumen, aprobación en lote
│   ├── closing-review/       # modal de revisión con cola Anterior/Siguiente
│   ├── summary-dashboard/    # Resumen: avance, totales, riesgo, métricas de éxito (014)
│   ├── parameters/           # pantalla de parámetros del contador
│   └── client-portal/        # vista del cliente
├── mocks/               # handlers MSW (solo tests)
└── test/                # setup de Vitest
```

### 3.3 Convenciones
- Cada feature expone un `index.ts` público; nada importa archivos internos de otra feature.
- Componentes presentacionales puros + hooks de datos (`useClosings`, `useClosingDetail`, ...).
- Estado de filtros, orden, página y cliente abierto **en la URL** (`useSearchParams`).
- Accesibilidad: el color nunca es la única señal (ícono + texto); navegación por teclado; foco atrapado en el modal.
- Idioma de la UI: español rioplatense.

## 4. Makefile

| Target | Hace |
|---|---|
| `make up` | `docker compose up --build`: Postgres + API + web. **Forma principal de correrlo.** |
| `make install` | `cd frontend && npm ci`. |
| `make dev` | Postgres en Docker + backend local (`go run ./cmd/api`) + frontend (`npm run dev`). |
| `make dev-memory` | Backend en memoria + frontend, sin Docker. |
| `make test` | Tests unitarios backend (`-race`) y frontend. No requiere Docker. |
| `make coverage` | Coverage de ambos; falla si alguno < 90%. |
| `make test-integration` | Tests de integración contra Postgres. |
| `make reset-db` | Borra el volumen de Postgres. |
| `make lint` | `go vet ./...`, `gofmt -l`, `npm run lint`, `npm run typecheck`. |
> Hasta la feature 013, `make dev` corre el backend en memoria.

## 5. Orden de features

```
000-foundation
001-reference-data           (mocks ARCA + FX + parámetros)
002-clients-and-movements    (modelo, seed, clasificación)
003-invoices-reconciliation
004-monotributo-assessment
005-ri-assessment-transition
006-risk-signals
007-monthly-closing          (job, estados, API de cierres)
008-accountant-dashboard     (front)
009-review-modal             (front + endpoint de IDs)
010-approval-notifications
011-client-portal
013-postgres-persistence     (Postgres + Docker Compose, tests de integración)
014-summary-dashboard        (Resumen: avance, totales, riesgo, métricas de éxito)
015-delivery                 (README, nota de decisiones, Swagger)
```
Cada feature depende de las anteriores. No empezar una feature con tareas pendientes de la anterior.
(La entrega original, `012-delivery`, se reemplazó por `015-delivery` cuando se sumaron Postgres y el Resumen; la carpeta 012 se borra.)
