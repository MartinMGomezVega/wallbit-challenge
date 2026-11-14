# 013 — Plan

## Punto de partida (estado real del repo al terminar 000–012)
- Módulo: `github.com/MartinMGomezVega/wallbit-challenge/backend`. Hoy no tiene dependencias externas.
- Puertos en `internal/application/ports/repositories.go`. Los repos Postgres implementan **esas interfaces tal como están** (incluyendo `SaveAll`, `ByClient`, `ListAll`, `ByMovementIDs`, `LatestExclusionDate`, etc.). No renombrar métodos existentes.
- Repos en memoria: `internal/infrastructure/memory/*`. Se mantienen.
- `cmd/api/main.go` construye los repos en memoria y carga el seed directamente en `run()`. Hay que extraer esa construcción a una función `buildRepos(cfg)` que elija `memory` o `postgres`, y mover la carga del seed al `seed_loader` (Postgres) o a un loader equivalente en memoria.
- `internal/application/usecase/list_closings.go`: `buildClosingRows` + `closing.Apply` + `closing.ComputeSummary` arman listado, IDs y summary en memoria. Se reemplaza por el puerto `ClosingQuery` (la versión en memoria **reutiliza** exactamente esas funciones).
- `internal/application/usecase/approval.go`: hoy compara la versión leyendo el cierre y después guarda (no es atómico). Se agrega `ClosingRepository.SaveIfVersion(ctx, c, expectedVersion) error` (devuelve `domain.ErrVersionConflict`); memoria lo resuelve bajo lock, Postgres con `UPDATE ... WHERE id = $1 AND version = $2`.
- `internal/config/config.go`: agregar `Storage`, `DatabaseURL`, `SeedReset` (con tests como los existentes).

## Dependencias (justificación, ver `architecture.md §2.1`)
- `github.com/jackc/pgx/v5` (driver + `pgxpool` + `CopyFrom`): driver nativo, sin `database/sql`, soporte de `COPY`.
- Sin ORM ni librería de migraciones: migrador propio de ~60 líneas.

## Estructura
```
backend/internal/infrastructure/postgres/
├── db.go                 # Connect(ctx, url) con reintentos; pgxpool
├── migrate.go            # aplica migrations/*.sql en orden, tabla schema_migrations, dentro de tx
├── migrations/
│   ├── 0001_init.sql
│   └── 0002_indexes.sql
├── tx.go                 # TxManager: guarda pgx.Tx en el context; helper q(ctx) devuelve Tx o Pool
├── client_repo.go, movement_repo.go, classification_repo.go, invoice_repo.go,
│   purchase_invoice_repo.go, client_input_repo.go, closing_repo.go,
│   closing_query.go, notification_repo.go, outbox_repo.go, parameters_repo.go
├── seed_loader.go        # si clients está vacío (o SEED_RESET): TRUNCATE ... CASCADE + CopyFrom del Dataset
└── *_integration_test.go # //go:build integration
backend/internal/application/ports/repotest/   # suites de contrato reutilizables
├── closing_query_suite.go   # func RunClosingQuerySuite(t *testing.T, newRepos func(t *testing.T) Repos)
└── ...
```

## Esquema (0001_init.sql, resumido)
```sql
CREATE TABLE clients (
  id TEXT PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL, cuit TEXT NOT NULL UNIQUE,
  regime TEXT NOT NULL CHECK (regime IN ('MONOTRIBUTO','RI')), category TEXT,
  ri_self_employed_cents BIGINT NOT NULL DEFAULT 0, ri_prev_vat_credit_cents BIGINT NOT NULL DEFAULT 0,
  name_search TEXT NOT NULL                       -- nombre normalizado para q=
);
CREATE TABLE regime_history (client_id TEXT REFERENCES clients(id), regime TEXT NOT NULL,
  valid_from DATE NOT NULL, valid_to DATE, reason TEXT, note TEXT, PRIMARY KEY (client_id, valid_from));
CREATE TABLE income_tax_advances (client_id TEXT REFERENCES clients(id), period TEXT, amount_cents BIGINT, due_date DATE,
  PRIMARY KEY (client_id, period));
CREATE TABLE movements (
  id TEXT PRIMARY KEY, client_id TEXT NOT NULL REFERENCES clients(id), type TEXT NOT NULL,
  amount_cents BIGINT NOT NULL, currency TEXT NOT NULL, fx_rate_scaled BIGINT NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL, cp_name TEXT, cp_country TEXT, cp_kind TEXT, cp_own BOOLEAN NOT NULL, description TEXT);
CREATE TABLE classifications (movement_id TEXT PRIMARY KEY REFERENCES movements(id), kind TEXT NOT NULL,
  confidence TEXT NOT NULL, rule TEXT NOT NULL, overridden_by TEXT, overridden_at TIMESTAMPTZ, note TEXT);
CREATE TABLE invoices (
  id TEXT PRIMARY KEY, client_id TEXT NOT NULL REFERENCES clients(id), type TEXT NOT NULL,
  point_of_sale INT NOT NULL, number BIGINT NOT NULL, issue_date DATE NOT NULL, currency TEXT NOT NULL,
  net_cents BIGINT NOT NULL, vat_cents BIGINT NOT NULL, total_cents BIGINT NOT NULL, fx_rate_scaled BIGINT,
  receiver_name TEXT NOT NULL, receiver_doc TEXT, source TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (client_id, type, point_of_sale, number));
CREATE TABLE purchase_invoices (...);
CREATE TABLE client_period_inputs (client_id TEXT, period TEXT, declared_none BOOLEAN NOT NULL DEFAULT false,
  confirmed_at TIMESTAMPTZ, PRIMARY KEY (client_id, period));
CREATE TABLE external_income_items (client_id TEXT, period TEXT, seq INT, amount_cents BIGINT, description TEXT,
  PRIMARY KEY (client_id, period, seq));
CREATE TABLE closings (
  id TEXT PRIMARY KEY, client_id TEXT NOT NULL REFERENCES clients(id), period TEXT NOT NULL,
  status TEXT NOT NULL, version INT NOT NULL, result_hash TEXT NOT NULL,
  severity SMALLINT NOT NULL,                     -- 0 green, 1 yellow, 2 red (para ORDER BY)
  primary_signal TEXT, primary_priority SMALLINT NOT NULL DEFAULT 99,
  description TEXT NOT NULL, regime TEXT NOT NULL, category TEXT,
  billed_usd_cents BIGINT NOT NULL, billed_ars_cents BIGINT NOT NULL,
  acc12m_ars_cents BIGINT NOT NULL, pct_k_cap_bp INT NOT NULL,
  to_pay_cents BIGINT NOT NULL, to_pay_due DATE,
  payload JSONB NOT NULL,                         -- assessment, señales completas, conciliación
  published JSONB, published_at TIMESTAMPTZ, paid_at TIMESTAMPTZ, computed_at TIMESTAMPTZ NOT NULL,
  UNIQUE (client_id, period));
CREATE TABLE closing_signals (closing_id TEXT REFERENCES closings(id) ON DELETE CASCADE, code TEXT NOT NULL,
  severity SMALLINT NOT NULL, PRIMARY KEY (closing_id, code));
CREATE TABLE approvals (id TEXT PRIMARY KEY, closing_id TEXT NOT NULL REFERENCES closings(id), version INT NOT NULL,
  result_hash TEXT NOT NULL, approved_by TEXT NOT NULL, approved_at TIMESTAMPTZ NOT NULL, client_note TEXT, batch BOOLEAN NOT NULL);
CREATE TABLE notifications (id TEXT PRIMARY KEY, client_id TEXT NOT NULL, closing_id TEXT, kind TEXT NOT NULL,
  dedup_key TEXT, title TEXT NOT NULL, body TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL, read_at TIMESTAMPTZ,
  UNIQUE (client_id, closing_id, kind, dedup_key));
CREATE TABLE outbox (id TEXT PRIMARY KEY, channel TEXT NOT NULL, client_id TEXT NOT NULL, subject TEXT NOT NULL,
  body TEXT NOT NULL, attempts INT NOT NULL DEFAULT 0, sent_at TIMESTAMPTZ, last_error TEXT, created_at TIMESTAMPTZ NOT NULL);
CREATE TABLE parameters (id SMALLINT PRIMARY KEY CHECK (id = 1), data JSONB NOT NULL, updated_at TIMESTAMPTZ, updated_by TEXT);
```
`0002_indexes.sql`: `closings (period, severity DESC, primary_priority, ...)`, `closing_signals (code)`, `movements (client_id, occurred_at)`, `invoices (client_id, issue_date)`, `outbox (sent_at) WHERE sent_at IS NULL`, `clients (name_search text_pattern_ops)`.

## `ClosingQuery` en SQL
- `WHERE c.period = $1` + filtros: `severity = ANY($n)`, `status = ANY($n)`, `regime = $n`, `EXISTS (SELECT 1 FROM closing_signals s WHERE s.closing_id = c.id AND s.code = ANY($n))`, `(cl.name_search LIKE '%'||$n||'%' OR cl.cuit LIKE ...)`.
- Orden `severity`: `ORDER BY c.severity DESC, c.primary_priority ASC, cl.name_search ASC, c.client_id ASC`. Para los otros sorts, siempre terminar con `c.client_id ASC` (orden total).
- `name_search` y el orden por nombre deben coincidir con `NormalizeName` + comparación de strings de Go (por eso se guarda normalizado y se ordena con `COLLATE "C"`).
- `Summary`: `COUNT(*) FILTER (WHERE ...)` y `GROUP BY` sobre `closing_signals`.
- Construcción de SQL con un builder mínimo propio (slice de condiciones + args), **sin concatenar input del usuario**.

## TxManager
```go
type TxManager interface { WithinTx(ctx context.Context, fn func(ctx context.Context) error) error }
```
Postgres: `pool.BeginTx` con `ReadCommitted`, guarda la tx en el ctx, commit/rollback. Los repos obtienen el ejecutor con `q(ctx)`.

## Docker
`docker-compose.yml`:
```yaml
services:
  db:
    image: postgres:17-alpine
    environment: { POSTGRES_USER: wallbit, POSTGRES_PASSWORD: wallbit, POSTGRES_DB: accountant }
    ports: ["5432:5432"]
    volumes: [pgdata:/var/lib/postgresql/data]
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U wallbit -d accountant"], interval: 2s, timeout: 3s, retries: 30 }
  api:
    build: ./backend
    environment:
      STORAGE: postgres
      DATABASE_URL: postgres://wallbit:wallbit@db:5432/accountant?sslmode=disable
      APP_NOW: "2026-11-05T09:00:00-03:00"
      CORS_ORIGIN: http://localhost:5173
    ports: ["8080:8080"]
    depends_on: { db: { condition: service_healthy } }
  web:
    build: ./frontend
    ports: ["5173:80"]
    depends_on: [api]
volumes: { pgdata: {} }
```
- `backend/Dockerfile`: multi-stage `golang:1.23-alpine` → `CGO_ENABLED=0 go build -o /api ./cmd/api` → `gcr.io/distroless/static` (usuario no root).
- `frontend/Dockerfile`: `node:20-alpine` (`npm ci && npm run build`) → `nginx:alpine` con `nginx.conf` (`location /api/ { proxy_pass http://api:8080; }`, `try_files $uri /index.html` para el router).
- Credenciales de demo en el compose (documentado: solo local).

## Makefile
| Target | Hace |
|---|---|
| `make up` | `docker compose up --build` (todo en Docker) |
| `make dev` | `docker compose up -d db` + backend local (`STORAGE=postgres`) + front con Vite |
| `make dev-memory` | backend en memoria + front, sin Docker |
| `make test-integration` | `docker compose up -d db` + `go test -tags=integration ./internal/infrastructure/postgres/...` con `TEST_DATABASE_URL` (base `accountant_test`, se crea/limpia sola) |
| `make reset-db` | `docker compose down -v` |
