# Backend

Go REST API using Gin, GORM, and PostgreSQL.

## Architecture

### Deployment

```
Internet → Traefik (TLS termination)
              ├── foodo.example.de        → Frontend (nginx + React SPA)
              ├── foodo.example.de/api/*  → Backend (Go API)
              └── foodo.example.de/uploads/* → static product images (served by Go)
                       ↓
                  PostgreSQL (internal)

Keycloak runs separately (auth.example.de) — JWT validation via JWKS endpoint.
Telegram bot is optional — receives a notification on each confirmed order.
```

### Layers

```
cmd/api/main.go
    │
    ├── infra/http          Gin router + handlers
    │     ├── middleware     JWTAuth (Keycloak JWKS) · SyncUser · RequireOwner
    │     └── handler        UserHandler · ProductHandler · OrderHandler
    │
    ├── domain              Pure business logic, no framework deps
    │     ├── user           User (customer/owner roles)
    │     ├── product        Product (name, price, unit, availability, image)
    │     └── order          Order + Items, status machine, Notifier interface
    │
    ├── infra/postgres       GORM adapters implementing domain Repository interfaces
    └── infra/telegram       Notifier implementation — fires on order confirmed
```

### API Routes (`/api/v1`)

All routes require a valid Keycloak JWT. `owner` routes additionally require `role=owner`.

| Method | Path | Role | Description |
|--------|------|------|-------------|
| `GET` | `/users` | any | List users |
| `POST` | `/users` | any | Register profile (synced from JWT on first request) |
| `GET` | `/users/me` | any | Current user |
| `GET` | `/users/:id` | any | Get user by ID |
| `PUT` | `/users/:id` | any | Update contact info |
| `GET` | `/products` | any | List products |
| `GET` | `/products/:id` | any | Get product |
| `POST` | `/products` | owner | Create product |
| `PUT` | `/products/:id` | owner | Update product |
| `PATCH` | `/products/:id/availability` | owner | Toggle availability |
| `DELETE` | `/products/:id` | owner | Delete product |
| `POST` | `/products/:id/image` | owner | Upload product image |
| `GET` | `/orders` | any | List orders |
| `POST` | `/orders` | any | Create order |
| `GET` | `/orders/:id` | any | Get order |
| `POST` | `/orders/:id/items` | any | Add item |
| `DELETE` | `/orders/:id/items/:productID` | any | Remove item |
| `POST` | `/orders/:id/confirm` | any | Confirm order (→ notifies Telegram) |
| `POST` | `/orders/:id/accept` | owner | Accept order |
| `POST` | `/orders/:id/start` | owner | Start progress |
| `POST` | `/orders/:id/finish` | owner | Finish order |
| `POST` | `/orders/:id/unaccept` | owner | Revert to created |
| `POST` | `/orders/:id/stop` | owner | Revert to accepted |
| `POST` | `/orders/:id/unfinish` | owner | Revert to ongoing |

### Order State Machine

```
pending → [confirm] → created → [accept] → accepted → [start] → ongoing → [finish] → finished
                                    ↑                      ↑
                               [unaccept]             [stop]
                                                      [unfinish] ←─ finished
```

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
