# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o upvote-service ./cmd/upvote-service/

# Run (DATABASE_URL required)
DATABASE_URL=postgres://user:pass@localhost:5432/upvotes go run ./cmd/upvote-service/main.go

# Test
go test ./...

# Run a single test
go test ./internal/service/ -run TestFunctionName

# Lint (requires golangci-lint)
golangci-lint run
```

## Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | yes | — | PostgreSQL DSN (`postgres://user:pass@host:port/db`) |
| `ADDR` | no | `:8080` | Address the HTTP server listens on |

## Architecture

This is a Go HTTP microservice for managing upvotes, using a **layered dependency-injection pattern** with chi as the HTTP router.

**Layers and dependency flow** (top-down):

| Layer | Package | Role |
|---|---|---|
| Entry point | `cmd/upvote-service/` | Wires everything together and starts the server |
| Transport | `internal/handlers/` | HTTP request/response handling |
| Business logic | `internal/service/` | Domain operations, UUID v7 generation |
| Persistence | `internal/repository/` | PostgreSQL data access via `pgx/v5` |
| Domain model | `internal/domain/` | `Upvote` entity (ID, UserID, ArticleID, CreatedAt) |
| HTTP server | `internal/server/` | Chi router setup, middleware (Logger, Recoverer, RequestID), route registration |
| Database | `internal/db/` | Connection pool setup and migration runner |

**Interface-based design**: each layer defines the interface it requires from the layer below. The handler package defines `upvoteService`, and the service package defines `upvoteRepository`. Concrete implementations are injected at startup.

**Routes** (defined in `internal/server/server.go`):
- `POST /upvotes` — Create
- `DELETE /upvotes/{upvoteID}` — Delete
- `GET /upvotes` — List
- `GET /upvotes/{upvoteID}` — Get

**Migrations** live in `internal/db/migrations/` and are embedded in the binary via `//go:embed`. They run automatically at startup via `db.Migrate()` and are idempotent (`CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`).

**UUID generation**: IDs are generated in the service layer (UUID v7, time-ordered) and passed to the repository — the database never generates IDs.
