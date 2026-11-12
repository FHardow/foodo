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

Migrations are **not** applied automatically — run them manually after the containers are up (see [backend/README.md](backend/README.md#migrations)):

```bash
cd backend && make migrate-up
```

`PUSH_ENCRYPTION_KEY` in `.env` is required — the backend refuses to start without it, even if you don't set the `VAPID_*` push notification keys. Generate it with `openssl rand -base64 32`.

> **Keycloak** must be running separately before starting the app. Create a `foodo` realm, set `KEYCLOAK_URL` in `.env`, and deploy the login theme from [`keycloak-theme/`](keycloak-theme/README.md).

## Updating

```bash
git pull
docker compose up -d --build
```
