# Bread Order

A bread ordering app with a React frontend, Go backend, Keycloak authentication, and PostgreSQL.

## Architecture

```
Internet → Traefik (TLS termination)
              ├── bread.fhardow.de        → Frontend (nginx + React SPA)
              └── bread.fhardow.de/api/*  → Backend (Go API)
                       ↓
                  PostgreSQL (internal)

Keycloak runs separately and is referenced via KEYCLOAK_URL.
```

## Prerequisites

- Docker and Docker Compose
- A Linux server with ports 80 and 443 open
- DNS A records pointing to your server:
  - `bread.fhardow.de`
- A running Keycloak instance (e.g. `auth.fhardow.de`) with the `bread-order` realm configured

## Deployment

### 1. Clone the repository

```bash
git clone <repo-url>
cd bread-order
```

### 2. Start Traefik

Traefik runs in its own compose file and creates the shared `traefik-public` Docker network that other services attach to.

```bash
docker compose -f docker-compose.traefik.yml up -d
```

This only needs to be done once per host. If you already have Traefik running elsewhere, make sure the `traefik-public` network exists:

```bash
docker network create traefik-public
```

### 3. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your values:

| Variable | Description |
|----------|-------------|
| `DOMAIN` | Your domain, e.g. `fhardow.de` |
| `ACME_EMAIL` | Email for Let's Encrypt certificate notifications |
| `DB_PASSWORD` | PostgreSQL password |
| `KEYCLOAK_URL` | URL of your Keycloak instance |
| `VITE_KEYCLOAK_URL` | Same URL, baked into the frontend at build time |

### 4. Build and start

```bash
docker compose up -d --build
```

On first start the Go backend runs database migrations automatically.

### 5. Verify

```bash
docker compose ps        # all services should be healthy/running
docker compose logs -f   # follow logs
```

## Updating

```bash
git pull
docker compose up -d --build
```

## Useful commands

```bash
# View logs for a specific service
docker compose logs -f backend

# Open a psql shell
docker compose exec postgres psql -U bread -d bread_order

# Rebuild a single service without downtime
docker compose up -d --build backend
```
