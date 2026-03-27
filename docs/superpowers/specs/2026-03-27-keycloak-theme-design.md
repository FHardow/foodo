# Keycloak Theme Design

**Date:** 2026-03-27
**Status:** Approved

## Overview

A custom Keycloak login theme built with Keycloakify (React + Tailwind), styled to match the bread-order frontend. Lives in a `keycloak-theme/` subfolder within the existing repo.

## Approach

**Keycloakify** — builds Keycloak themes as a React + Vite project. Running `npm run build` produces a `.jar` file that is deployed to Keycloak's `providers/` directory. The realm's login theme is then set in the Keycloak admin console.

Chosen over a plain FreeMarker theme because it allows using the same stack (React, Tailwind) as the main app, making it straightforward to match the visual style precisely.

## Repository Structure

```
bread-order/
  frontend/               ← existing app (untouched)
  keycloak-theme/         ← new Vite + Keycloakify project
    src/
      login/
        KcPage.tsx        ← entry point, routes to the correct page component
        Login.tsx         ← login form page
        Register.tsx      ← register form page
    tailwind.config.ts    ← duplicates color tokens from frontend
    vite.config.ts
    package.json
```

## Visual Style

Matches the existing app's warm bakery aesthetic.

| Token | Value | Usage |
|-------|-------|-------|
| Page background | `#faf7f2` | Full-page background |
| Card background | `#ffffff` | Form card |
| Card border | `#e8ddd0` | Card border |
| Primary brown | `#5c3d1e` | Buttons, labels, links, brand text |
| Primary dark | `#3d2b1a` | Headings, input text |
| Muted | `#8a6a50` | Tagline, secondary text |
| Input background | `#faf7f2` | Form inputs |

**Card:** white, `rounded-lg`, `border border-[#e8ddd0]`, subtle shadow, centered on page.

**Brand header** (top of card): `🍞 Bread Order` in bold `#5c3d1e` with "Your local bakery" tagline in muted text.

**Labels:** small, `#5c3d1e`, semibold.

**Inputs:** cream background, tan border, rounded, focus ring in `#5c3d1e`.

**Primary button:** `bg-[#5c3d1e] text-white rounded-lg hover:bg-[#3d2b1a] transition-colors`, full width.

## Pages

### Login

Fields:
- Email (type `email`)
- Password (type `password`)

Actions:
- **Sign in** button (primary, full width)
- **Forgot password?** link (bottom left)
- **Create account** link (bottom right)

### Register

Fields:
- First name + Last name (side by side, equal width)
- Email (type `email`)
- Password (type `password`)
- Confirm password (type `password`)

Actions:
- **Create account** button (primary, full width)
- **Already have an account? Sign in** link (centered, below divider)

## Error Handling

Keycloakify passes field-level and global error messages as props. Each input renders an error message directly below it as small red text when present. Global errors (e.g. invalid credentials) appear above the form fields inside the card.

## Deployment

1. `cd keycloak-theme && npm run build` → produces `keycloak-theme-*.jar`
2. Copy `.jar` to Keycloak's `providers/` directory
3. Restart Keycloak
4. Keycloak admin console → Realm → Themes → set **Login theme** to `bread-order`
