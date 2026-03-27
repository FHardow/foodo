# Bread Order

A bread ordering app with a React frontend, Go backend, Keycloak authentication, and PostgreSQL.

## Architecture

```
Internet → Traefik (TLS termination)
              ├── bread.example.com        → Frontend (nginx + React SPA)
              ├── bread.example.com/api/*  → Backend (Go API)
              └── auth.bread.example.com   → Keycloak
                       ↓
                  PostgreSQL (internal)
```

## Prerequisites

- Docker and Docker Compose
- A Linux server with ports 80 and 443 open
- DNS A records pointing to your server:
  - `bread.example.com`
  - `auth.bread.example.com`

## Deployment

### 1. Clone the repository

```bash
git clone <repo-url>
cd bread-order
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your values:

| Variable | Description |
|----------|-------------|
| `DOMAIN` | Your public domain, e.g. `bread.example.com` |
| `ACME_EMAIL` | Email for Let's Encrypt certificate notifications |
| `DB_PASSWORD` | PostgreSQL password |
| `KEYCLOAK_ADMIN_PASSWORD` | Keycloak admin console password |

### 3. Build and start

```bash
docker compose up -d --build
```

On first start:
- PostgreSQL initializes both the app database and the Keycloak database
- Keycloak imports `realm-export.json` and sets up the `bread-order` realm
- The Go backend runs database migrations automatically

### 4. Verify

```bash
docker compose ps        # all services should be healthy/running
docker compose logs -f   # follow logs
```

The app is available at `https://bread.example.com` once Keycloak has finished starting (it can take ~30–60 seconds on first boot).

## Moving Traefik to a shared instance

If you already have a Traefik instance running on the host:

1. Remove the `traefik` service from `docker-compose.yml`
2. Remove the `letsencrypt` volume
3. Create an external Docker network (e.g. `traefik-public`) and add it to each service:
   ```yaml
   networks:
     default:
     traefik-public:
       external: true
   ```
4. Add `--providers.docker.network=traefik-public` to your existing Traefik config

The Traefik labels on each service stay the same.

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
