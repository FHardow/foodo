# Foodo

A food ordering app with a React frontend, Go backend, Keycloak authentication, and PostgreSQL.

- [Backend](backend/README.md) — Go API, migrations, local dev, env vars
- [Frontend](frontend/README.md) — React SPA, testing, local dev, env vars
- [Keycloak Theme](keycloak-theme/README.md) — Keycloakify login theme, build, deployment

## Architecture

```
Internet → Traefik (TLS termination)
              ├── foodo.example.de        → Frontend (nginx + React SPA)
              └── foodo.example.de/api/*  → Backend (Go API)
                       ↓
                  PostgreSQL (internal)

Keycloak runs separately and is referenced via KEYCLOAK_URL.
```

## Prerequisites

- Docker and Docker Compose
- A Linux server with ports 80 and 443 open
- DNS A record pointing to your server (`foodo.example.de`)
- A running Keycloak instance (e.g. `auth.example.de`) with the `foodo` realm configured

## Deployment

### 1. Clone the repository

```bash
git clone <repo-url>
cd foodo
```

### 2. Start Traefik

Traefik runs in its own compose file and creates the shared `traefik-public` Docker network.

```bash
docker compose -f docker-compose.traefik.yml up -d
```

Only needed once per host. If Traefik is already running elsewhere:

```bash
docker network create traefik-public
```

### 3. Configure environment

```bash
cp .env.example .env
```

| Variable | Description |
|----------|-------------|
| `DOMAIN` | Your domain, e.g. `example.de` |
| `ACME_EMAIL` | Email for Let's Encrypt notifications |
| `DB_PASSWORD` | PostgreSQL password |
| `KEYCLOAK_URL` | URL of your Keycloak instance |
| `VITE_KEYCLOAK_URL` | Same URL, baked into the frontend at build time |

### 4. Build and start

```bash
docker compose up -d --build
```

On first start the backend runs database migrations automatically.

### 5. Verify

```bash
docker compose ps        # all services healthy/running
docker compose logs -f   # follow logs
```

## Updating

```bash
git pull
docker compose up -d --build
```
