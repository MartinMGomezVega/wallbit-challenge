#!/usr/bin/env bash
# Coverage del backend excluyendo cmd/. Falla si el total es < 90%.
set -euo pipefail
cd "$(dirname "$0")/.."
THRESHOLD="${COVERAGE_THRESHOLD:-90}"
go test -race -count=1 -covermode=atomic -coverpkg=./internal/... -coverprofile=coverage.out ./...
# Filtra cualquier línea de /cmd/ que haya quedado en el perfil.
grep -v '/cmd/' coverage.out > coverage.filtered.out
TOTAL=$(go tool cover -func=coverage.filtered.out | awk '/^total:/ {gsub("%","",$3); print $3}')
echo "Backend coverage total: ${TOTAL}% (mínimo ${THRESHOLD}%)"
awk -v t="$TOTAL" -v min="$THRESHOLD" 'BEGIN { exit (t+0 >= min+0) ? 0 : 1 }' || {
  echo "ERROR: coverage ${TOTAL}% < ${THRESHOLD}%" >&2
  exit 1
}

