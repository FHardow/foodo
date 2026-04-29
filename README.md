# Foodo

A food ordering app with a React frontend, Go backend, Keycloak authentication, and PostgreSQL.

- [Backend](backend/README.md) — Go API, migrations, local dev, env vars
- [Frontend](frontend/README.md) — React SPA, testing, local dev, env vars
- [Keycloak Theme](keycloak-theme/README.md) — Keycloakify login theme, build, deployment

## Deployment

```bash
git clone <repo-url>
cd foodo
cp .env.example .env   # fill in DOMAIN, DB_PASSWORD, KEYCLOAK_URL, etc.
docker compose up -d --build
```

Migrations run automatically on first start.

## Updating

```bash
git pull
docker compose up -d --build
```
