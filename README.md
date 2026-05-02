# upvote-service

A Go HTTP microservice for managing upvotes on articles. Built with a layered dependency-injection architecture using [chi](https://github.com/go-chi/chi) as the HTTP router and PostgreSQL as the database.

## Features

- Create and delete upvotes
- List upvotes by article or by user
- One upvote per user per article

## Requirements

- Go 1.26+
- Docker (to run the service and for integration tests)
- [golangci-lint](https://golangci-lint.run) (optional, for linting)

## Getting started

### With Docker (recommended)

Builds and starts both PostgreSQL and the service in one command:

```bash
./scripts/run.sh
```

The service will be available at `http://localhost:8080`. Migrations are applied automatically on startup.

### Without Docker

#### 1. Start PostgreSQL

```bash
docker run -d \
  --name upvote-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=upvote_service \
  -p 5432:5432 \
  postgres:17
```

#### 2. Configure environment

```bash
cp .env.example .env
```

Default `.env`:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/upvote_service
ADDR=:8080
```

#### 3. Run

```bash
go run ./cmd/upvote-service/main.go
```

Migrations are applied automatically on startup.

## Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | yes | — | PostgreSQL DSN (`postgres://user:pass@host:port/db`) |
| `ADDR` | no | `:8080` | Address the HTTP server listens on |

## API

### Create an upvote

```
POST /upvotes
```

```json
{
  "article_id": "article-123",
  "user_id": "user-456"
}
```

**Responses**

| Status | Description |
|---|---|
| `201 Created` | Upvote created successfully |
| `400 Bad Request` | Invalid request body |
| `409 Conflict` | User has already upvoted this article |

---

### Delete an upvote

```
DELETE /upvotes/{upvoteID}
```

**Responses**

| Status | Description |
|---|---|
| `204 No Content` | Upvote deleted |
| `500 Internal Server Error` | Upvote not found or server error |

---

### Get an upvote by ID

```
GET /upvotes/{upvoteID}
```

**Responses**

| Status | Description |
|---|---|
| `200 OK` | Upvote found |
| `500 Internal Server Error` | Upvote not found or server error |

---

### List upvotes

```
GET /upvotes?article_id={articleID}
GET /upvotes?user_id={userID}
```

Exactly one query parameter is required.

**Responses**

| Status | Description |
|---|---|
| `200 OK` | List of upvotes |
| `400 Bad Request` | Missing `article_id` or `user_id` query param |

## Architecture

Layered dependency-injection pattern. Each layer exposes an interface consumed by the layer above — concrete implementations are injected at startup in `main.go`.

```
cmd/upvote-service/   → wires all layers and starts the HTTP server
internal/
  handlers/           → decodes HTTP requests, encodes responses
  service/            → business logic and domain rules
  repository/         → SQL queries against PostgreSQL
  domain/             → Upvote struct and domain errors
  server/             → chi router and middleware
  db/
    db.go             → connection pool and migration runner
    migrations/       → SQL migration files (embedded in the binary)
```

## Development

```bash
# Build
go build -o upvote-service ./cmd/upvote-service/

# Run all tests (unit + integration, requires Docker)
./scripts/test.sh

# Run only unit tests (no Docker needed)
go test ./internal/service/... ./internal/handlers/... -race

# Run only integration tests
go test ./internal/repository/... -race -timeout 120s

# Lint
golangci-lint run
```

## Migrations

Migration files live in `internal/db/migrations/` and are embedded in the binary at compile time via `//go:embed`. They run automatically at startup and are idempotent (`CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`).

To add a new migration, create the next numbered pair of files:

```
internal/db/migrations/000002_your_description.up.sql
internal/db/migrations/000002_your_description.down.sql
```

## License

[MIT](LICENSE)
