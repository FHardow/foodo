# Backend

Go REST API using Gin, GORM, and PostgreSQL.

## Stack

- **Go 1.26** / Gin HTTP framework
- **GORM** + PostgreSQL
- **Keycloak** JWT validation
- **Testcontainers** for integration tests

## Project Structure

```
cmd/api/        entry point
internal/
  config/       env-based config
  domain/       business logic (order, product, user)
  infra/        adapters (http, postgres, telegram)
migrations/     SQL migration files
pkg/            shared utilities (errors, logger)
```

## Local Development

```bash
# Copy and fill in env
cp .env.example .env   # or edit .env directly

# Run (requires a local Postgres)
make run

# Build binary
make build
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `DB_HOST` | Postgres host |
| `DB_USER` | Postgres user |
| `DB_PASSWORD` | Postgres password |
| `DB_NAME` | Database name |
| `DB_PORT` | Postgres port (default `5432`) |
| `DB_SSLMODE` | `disable` for local, `require` for prod |
| `CORS_ORIGIN` | Allowed frontend origin |
| `KEYCLOAK_URL` | Keycloak base URL |

## Migrations

Migrations run automatically on startup. To run or roll back manually (requires [golang-migrate](https://github.com/golang-migrate/migrate)):

```bash
make migrate-up      # apply all pending
make migrate-down    # roll back one step
```

Migration files live in `migrations/` as `NNN_description.{up,down}.sql`.

## Testing

```bash
make test   # runs all tests including integration tests via Testcontainers
```

Integration tests spin up a real Postgres container — no mocks needed.

## Lint

```bash
make lint   # requires golangci-lint
```

## Useful Commands

```bash
# Follow backend logs (Docker)
docker compose logs -f backend

# Open a psql shell (Docker)
docker compose exec postgres psql -U foodo -d foodo

# Rebuild backend only (Docker, no downtime)
docker compose up -d --build backend
```
