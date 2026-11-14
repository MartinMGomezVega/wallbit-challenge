# 013 — Persistencia en PostgreSQL y entorno con Docker

## Objetivo
Reemplazar la persistencia en memoria por **PostgreSQL** como almacenamiento por defecto, sin tocar el dominio ni los casos de uso (solo adaptadores nuevos detrás de los mismos puertos), y levantar todo el prototipo con **un solo comando**: `docker compose up --build`.

## Por qué (decisión de producto/técnica)
- Una base propia no es un "servicio externo": es infraestructura del backend. Los mocks siguen siendo solo ARCA y tipo de cambio.
- Persistencia real: la auditoría de aprobaciones, los overrides y los avisos sobreviven a un reinicio.
- Transacciones: aprobar + publicar + notificar + encolar salientes es **atómico** (outbox real).
- Concurrencia optimista con columna `version`.
- El tablero y el dashboard (014) se resuelven con consultas SQL (filtros, orden, paginación y agregaciones en la base).
- Postgres y no DynamoDB: dominio relacional, transacciones multi-entidad y agregaciones.

## Criterios de aceptación
### Entorno
1. `docker compose up --build` levanta `db` (Postgres 17), `api` (Go) y `web` (front compilado servido por nginx con proxy `/api` → `api:8080`). La app queda en `http://localhost:5173` y la API en `http://localhost:8080`.
2. El primer arranque crea el esquema (migraciones embebidas), genera el seed (300 clientes, mismos datos que en memoria) y corre el cierre del período. Los arranques siguientes **no** regeneran nada (idempotente). `SEED_RESET=true` vacía las tablas y vuelve a sembrar.
3. `STORAGE=postgres|memory` (default `postgres` en Docker; `make dev-memory` sigue funcionando sin Docker).
4. `api` espera a que `db` esté healthy (`pg_isready`) y reintenta la conexión hasta 30 s.

### Adaptador
5. Implementaciones Postgres de **todos** los repositorios de `application/ports` (Client, Movement, Classification, Invoice, PurchaseInvoice, ClientInput, Closing, Notification, Outbox, Parameters).
6. Dinero en columnas `BIGINT` (centavos) + `currency`. Nunca `NUMERIC` → `float`.
7. **Consultas del tablero en SQL:** nuevo puerto `ClosingQuery` con `Page(ctx, filter, sort, page)`, `IDs(ctx, filter, sort)`, `Summary(ctx, filter)`. La implementación en memoria usa `closing.Apply`; la de Postgres usa SQL con índices. Ambas pasan **la misma suite de tests de contrato** (mismos resultados y mismo orden para 10 combinaciones de filtro/orden).
8. **Unidad de trabajo:** puerto `TxManager.WithinTx(ctx, fn)`. `ApproveClosing`, `ApproveBatch` (por ítem), `ComputeClosing`, `ChangeRegime` e `ImportInvoicesCSV` corren dentro de una transacción. En memoria, `WithinTx` toma un lock global.
9. **Concurrencia optimista:** `UPDATE closings ... WHERE id = $1 AND version = $2`; 0 filas → `ErrVersionConflict`. Test de integración con dos aprobaciones concurrentes: una gana, la otra recibe 409.
10. Upsert idempotente de cierres por `(client_id, period)`.
11. El seed inserta con `COPY` (pgx `CopyFrom`) y tarda < 5 s para 300 clientes.

### Tests
12. Los tests unitarios siguen usando memoria; el **coverage unitario ≥ 90%** excluye `internal/infrastructure/postgres/**` (documentado en `scripts/coverage.sh` y README).
13. Tests de integración con build tag `integration` contra el Postgres del compose (`make test-integration`): suites de contrato de repositorios, `ClosingQuery`, `TxManager` (rollback ante error) y concurrencia optimista. Coverage del paquete `postgres` ≥ 80% en esa corrida.
14. `make test` no requiere Docker.

## Fuera de alcance
Base gestionada en la nube, réplicas de lectura, backups, particionado por período (van al README como "lo que dejé afuera").
