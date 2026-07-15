[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/golang-migrate/migrate/ci.yaml?branch=master)](https://github.com/golang-migrate/migrate/actions/workflows/ci.yaml?query=branch%3Amaster)
[![GoDoc](https://pkg.go.dev/badge/github.com/golang-migrate/migrate)](https://pkg.go.dev/github.com/golang-migrate/migrate/v4)
[![Coverage Status](https://img.shields.io/coveralls/github/golang-migrate/migrate/master.svg)](https://coveralls.io/github/golang-migrate/migrate?branch=master)
[![packagecloud.io](https://img.shields.io/badge/deb-packagecloud.io-844fec.svg)](https://packagecloud.io/golang-migrate/migrate?filter=debs)
[![Docker Pulls](https://img.shields.io/docker/pulls/migrate/migrate.svg)](https://hub.docker.com/r/migrate/migrate/)
![Supported Go Versions](https://img.shields.io/badge/Go-1.24%2C%201.25-lightgrey.svg)
[![GitHub Release](https://img.shields.io/github/release/golang-migrate/migrate.svg)](https://github.com/golang-migrate/migrate/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/golang-migrate/migrate/v4)](https://goreportcard.com/report/github.com/golang-migrate/migrate/v4)

# migrate

__Database migrations written in Go. Use as [CLI](#cli-usage) or import as [library](#use-in-your-go-project).__

* Migrate reads migrations from [sources](#migration-sources)
   and applies them in correct order to a [database](#databases).
* Drivers are "dumb", migrate glues everything together and makes sure the logic is bulletproof.
   (Keeps the drivers lightweight, too.)
* Database drivers don't assume things or try to correct user input. When in doubt, fail.

Forked from [mattes/migrate](https://github.com/mattes/migrate)

## Databases

Database drivers run migrations. [Add a new database?](database/driver.go)

* [PostgreSQL](database/postgres)
* [PGX v4](database/pgx)
* [PGX v5](database/pgx/v5)
* [Redshift](database/redshift)
* [Ql](database/ql)
* [Cassandra / ScyllaDB](database/cassandra)
* [SQLite](database/sqlite)
* [SQLite3](database/sqlite3) ([todo #165](https://github.com/mattes/migrate/issues/165))
* [SQLCipher](database/sqlcipher)
* [MySQL / MariaDB](database/mysql)
* [Neo4j](database/neo4j)
* [MongoDB](database/mongodb)
* [CrateDB](database/crate) ([todo #170](https://github.com/mattes/migrate/issues/170))
* [Shell](database/shell) ([todo #171](https://github.com/mattes/migrate/issues/171))
* [Google Cloud Spanner](database/spanner)
* [CockroachDB](database/cockroachdb)
* [YugabyteDB](database/yugabytedb)
* [ClickHouse](database/clickhouse)
* [Firebird](database/firebird)
* [MS SQL Server](database/sqlserver)
* [rqlite](database/rqlite)

### Database URLs

Database connection strings are specified via URLs. The URL format is driver dependent but generally has the form: `dbdriver://username:password@host:port/dbname?param1=true&param2=false`

Any [reserved URL characters](https://en.wikipedia.org/wiki/Percent-encoding#Percent-encoding_reserved_characters) need to be escaped. Note, the `%` character also [needs to be escaped](https://en.wikipedia.org/wiki/Percent-encoding#Percent-encoding_the_percent_character)

Explicitly, the following characters need to be escaped:
`!`, `#`, `$`, `%`, `&`, `'`, `(`, `)`, `*`, `+`, `,`, `/`, `:`, `;`, `=`, `?`, `@`, `[`, `]`

It's easiest to always run the URL parts of your DB connection URL (e.g. username, password, etc) through an URL encoder. See the example Python snippets below:

```bash
$ python3 -c 'import urllib.parse; print(urllib.parse.quote(input("String to encode: "), ""))'
String to encode: FAKEpassword!#$%&'()*+,/:;=?@[]
FAKEpassword%21%23%24%25%26%27%28%29%2A%2B%2C%2F%3A%3B%3D%3F%40%5B%5D
$ python2 -c 'import urllib; print urllib.quote(raw_input("String to encode: "), "")'
String to encode: FAKEpassword!#$%&'()*+,/:;=?@[]
FAKEpassword%21%23%24%25%26%27%28%29%2A%2B%2C%2F%3A%3B%3D%3F%40%5B%5D
$
```

## Migration Sources

Source drivers read migrations from local or remote sources. [Add a new source?](source/driver.go)

* [Filesystem](source/file) - read from filesystem
* [io/fs](source/iofs) - read from a Go [io/fs](https://pkg.go.dev/io/fs#FS)
* [Go-Bindata](source/go_bindata) - read from embedded binary data ([jteeuwen/go-bindata](https://github.com/jteeuwen/go-bindata))
* [pkger](source/pkger) - read from embedded binary data ([markbates/pkger](https://github.com/markbates/pkger))
* [GitHub](source/github) - read from remote GitHub repositories
* [GitHub Enterprise](source/github_ee) - read from remote GitHub Enterprise repositories
* [Bitbucket](source/bitbucket) - read from remote Bitbucket repositories
* [Gitlab](source/gitlab) - read from remote Gitlab repositories
* [AWS S3](source/aws_s3) - read from Amazon Web Services S3
* [Google Cloud Storage](source/google_cloud_storage) - read from Google Cloud Platform Storage

## CLI usage

* Simple wrapper around this library.
* Handles ctrl+c (SIGINT) gracefully.
* No config search paths, no config files, no magic ENV var injections.

[CLI Documentation](cmd/migrate) (includes CLI install instructions)

### Basic usage

```bash
$ migrate -source file://path/to/migrations -database postgres://localhost:5432/database up 2
```

### Docker usage

```bash
$ docker run -v {{ migration dir }}:/migrations --network host migrate/migrate
    -path=/migrations/ -database postgres://localhost:5432/database up 2
```

## Use in your Go project

* API is stable and frozen for this release (v3 & v4).
* Uses [Go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) to manage dependencies.
* To help prevent database corruptions, it supports graceful stops via `GracefulStop chan bool`.
* Bring your own logger.
* Uses `io.Reader` streams internally for low memory overhead.
* Thread-safe and no goroutine leaks.

__[Go Documentation](https://pkg.go.dev/github.com/golang-migrate/migrate/v4)__

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/github"
)

func main() {
    m, err := migrate.New(
        "github://mattes:personal-access-token@mattes/migrate_test",
        "postgres://localhost:5432/database?sslmode=enable")
    m.Steps(2)
}
```

Want to use an existing database client?

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    db, err := sql.Open("postgres", "postgres://localhost:5432/database?sslmode=enable")
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    m, err := migrate.NewWithDatabaseInstance(
        "file:///migrations",
        "postgres", driver)
    m.Up() // or m.Steps(2) if you want to explicitly set the number of migrations to run
}
```

## Getting started

Go to [getting started](GETTING_STARTED.md)

## Tutorials

* [CockroachDB](database/cockroachdb/TUTORIAL.md)
* [PostgreSQL](database/postgres/TUTORIAL.md)

(more tutorials to come)

## Migration files

Each migration has an up and down migration. [Why?](FAQ.md#why-two-separate-files-up-and-down-for-a-migration)

```bash
1481574547_create_users_table.up.sql
1481574547_create_users_table.down.sql
```

[Best practices: How to write migrations.](MIGRATIONS.md)

## Coming from another db migration tool?

Check out [migradaptor](https://github.com/musinit/migradaptor/).
*Note: migradaptor is not affiliated or supported by this project*

## Versions

Version | Supported? | Import | Notes
--------|------------|--------|------
**master** | :white_check_mark: | `import "github.com/golang-migrate/migrate/v4"` | New features and bug fixes arrive here first |
**v4** | :white_check_mark: | `import "github.com/golang-migrate/migrate/v4"` | Used for stable releases |
**v3** | :x: | `import "github.com/golang-migrate/migrate"` (with package manager) or `import "gopkg.in/golang-migrate/migrate.v3"` (not recommended) | **DO NOT USE** - No longer supported |

## Development and Contributing

Yes, please! [`Makefile`](Makefile) is your friend,
read the [development guide](CONTRIBUTING.md).

Also have a look at the [FAQ](FAQ.md).

---

Looking for alternatives? [https://awesome-go.com/#database](https://awesome-go.com/#database).
=======
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
