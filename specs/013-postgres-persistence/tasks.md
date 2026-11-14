# 013 — Tasks

> Antes de empezar: verificar que 000–012 están completas (la entrega pasó a 015) y `make coverage` pasa en memoria.

- [x] **T013.1** Puertos nuevos: `TxManager` y `ClosingQuery` (`Page`, `IDs`, `Summary`). Implementación en memoria (`WithinTx` con lock global; `ClosingQuery` con `closing.Apply`). Refactor de `ListClosings`, `ListClosingIDs`, `ClosingsSummary` para usar `ClosingQuery`. Tests unitarios siguen verdes.
- [x] **T013.1b** `ClosingRepository.SaveIfVersion` (memoria bajo lock) y usarlo en `ApproveClosing`/`ApproveBatch`. Test unitario con dos aprobaciones concurrentes sobre la misma versión: una OK, otra `ErrVersionConflict`.
- [x] **T013.2** Envolver en `WithinTx`: `ComputeClosing`, `ApproveClosing`, cada ítem de `ApproveBatch`, `ChangeRegime`, `ImportInvoicesCSV`. Test unitario: si un paso falla dentro de la tx, no queda nada a medio guardar (con fakes que fallan).
- [x] **T013.3** Suites de contrato en `ports/repotest` (repos + `ClosingQuery` con 10 combinaciones de filtro/orden + `TxManager` rollback). Correrlas contra memoria en tests unitarios.
- [x] **T013.4** `postgres/db.go` (pool + reintentos) y `migrate.go` + migraciones `0001`, `0002`. Test de integración: migrar dos veces es idempotente.
- [x] **T013.5** Repos Postgres (Client, Movement, Classification, Invoice, PurchaseInvoice, ClientInput) + suites de contrato en integración.
- [x] **T013.6** `ClosingRepository` Postgres (upsert, `closing_signals`, concurrencia optimista con `version`) + test de dos aprobaciones concurrentes.
- [x] **T013.7** `ClosingQuery` Postgres (builder SQL parametrizado) + suite de contrato (mismo orden que memoria).
- [x] **T013.8** Notification, Outbox, Parameters repos + `TxManager` Postgres + suites.
- [x] **T013.9** `seed_loader.go` con `CopyFrom`, idempotencia y `SEED_RESET`. Test: dos arranques seguidos no duplican; reset regenera.
- [x] **T013.10** Config `STORAGE`, `DATABASE_URL`, `SEED_RESET`; wiring en `main.go` según `STORAGE`.
- [x] **T013.11** Dockerfiles (backend distroless, frontend nginx con proxy y fallback SPA), `docker-compose.yml`, `.dockerignore`.
  - Listo cuando: `docker compose up --build` en una carpeta limpia deja la app andando en `:5173` y el e2e manual de la demo funciona.
- [x] **T013.12** Makefile (`up`, `dev`, `dev-memory`, `test-integration`, `reset-db`) + `scripts/coverage.sh` excluye `infrastructure/postgres`. `make coverage` ≥ 90% y `make test-integration` verde.
- [x] **T013.13** Suite de contrato memoria vs Postgres (`CompareClosingQueries` con `reflect.DeepEqual` sobre `Page` y `Summary`, 10 combinaciones sobre el seed completo de 300 clientes).
