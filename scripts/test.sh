#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Integration tests require Docker (testcontainers)."
    exit 1
fi

echo "==> Unit tests (service, handlers)"
go test ./internal/service/... ./internal/handlers/... -v -race

echo ""
echo "==> Integration tests (repository — spins up a real Postgres via Docker)"
go test ./internal/repository/... -v -race -timeout 120s
