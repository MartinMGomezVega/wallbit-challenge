.PHONY: all install up dev dev-memory test test-integration coverage lint reset-db fixtures

TEST_DATABASE_URL ?= postgres://wallbit:wallbit@localhost:5432/accountant_test?sslmode=disable
DATABASE_URL ?= postgres://wallbit:wallbit@localhost:5432/accountant?sslmode=disable

all: test coverage lint

install:
	cd backend && go mod download
	cd frontend && npm ci

up:
	docker compose up --build

dev:
	docker compose up -d db
	@trap 'kill 0' EXIT; \
	(cd backend && STORAGE=postgres DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api) & \
	(cd frontend && npm run dev) & \
	wait

dev-memory:
	@trap 'kill 0' EXIT; \
	(cd backend && STORAGE=memory go run ./cmd/api) & \
	(cd frontend && npm run dev) & \
	wait

fixtures:
	cd backend && WRITE_FIXTURES=1 go test -run TestWriteFrontendFixtures ./internal/interfaces/http

test: fixtures
	cd backend && go test -race ./...
	cd frontend && npm run test

test-integration:
	docker compose up -d db
	cd backend && TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -race -tags=integration ./internal/infrastructure/postgres/...

coverage:
	cd backend && ./scripts/coverage.sh
	cd frontend && npm run coverage

lint:
	cd backend && [ -z "$$(gofmt -l .)" ] || (echo "Archivos sin formatear:"; gofmt -l .; exit 1)
	cd backend && go vet ./...
	cd frontend && npm run lint
	cd frontend && npm run typecheck

reset-db:
	docker compose down -v
