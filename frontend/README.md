# Frontend

React SPA using Vite, TypeScript, Tailwind CSS, and Keycloak auth.

## Stack

- **React 19** + TypeScript
- **Vite 8** dev server and bundler
- **Tailwind CSS 4**
- **TanStack Query** for server state
- **Zustand** for client state (basket)
- **Keycloak-js** for authentication
- **dnd-kit** for drag-and-drop
- **Vitest** + Testing Library for unit tests
- **Playwright** for end-to-end tests

## Project Structure

```
src/
  api/          API client and resource modules (orders, products)
  auth/         Keycloak init and auth utilities
  components/   Shared UI components (Nav, ProductCard, StatusBadge)
  hooks/        Custom hooks (useBasket)
  pages/        Route-level page components
    admin/      Admin views
  store/        Zustand stores (basket)
  types/        Shared TypeScript types
e2e/            Playwright end-to-end tests
```

## Local Development

```bash
npm install
npm run dev    # starts at http://localhost:5173
```

### Environment Variables

Create `.env.local`:

| Variable | Description |
|----------|-------------|
| `VITE_KEYCLOAK_URL` | Keycloak base URL (baked in at build time) |
| `VITE_API_URL` | Backend API base URL |

## Testing

```bash
npm test              # unit tests (Vitest), single run
npm run test:watch    # unit tests in watch mode

npm run test:e2e      # Playwright e2e (headless)
npm run test:e2e:ui   # Playwright e2e with UI
```

## Build

```bash
npm run build    # tsc + vite build → dist/
npm run preview  # preview the production build locally
```

## Lint

```bash
npm run lint
```

## Useful Commands

```bash
# Follow frontend logs (Docker)
docker compose logs -f frontend

# Rebuild frontend only (Docker, no downtime)
docker compose up -d --build frontend
```
