.PHONY: all dev test coverage lint

all: test coverage lint

dev:
	@trap 'kill 0' EXIT; \
	(cd backend && go run ./cmd/api) & \
	(cd frontend && npm run dev) & \
	wait

test:
	cd backend && go test -race ./...
	cd frontend && npm run test

coverage:
	cd backend && ./scripts/coverage.sh
	cd frontend && npm run coverage

lint:
	cd backend && [ -z "$$(gofmt -l .)" ] || (echo "Archivos sin formatear:"; gofmt -l .; exit 1)
	cd backend && go vet ./...
	cd frontend && npm run lint
	cd frontend && npm run typecheck
