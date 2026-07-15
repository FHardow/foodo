# Keycloak Theme

Custom Keycloakify login theme for foodo. Styled to match the app's warm bakery aesthetic (Tailwind v4, React).

## Stack

- **Keycloakify 11** — React-based Keycloak theme framework
- **React 18** + TypeScript
- **Tailwind CSS 4**
- **Storybook** for isolated page development

## Project Structure

```
src/
  login/
    pages/       Page components (Login, Register)
    Template.tsx  Shared page layout/wrapper
    KcPage.tsx    Page router (maps Keycloak page IDs to components)
    KcContext.ts  Keycloak context type extensions
    i18n.ts       i18n customizations
dist_keycloak/   Built .jar files ready for Keycloak deployment
```

## Pages

- **Login** — email/username + password, forgot password and create account links
- **Register** — first name, last name, email, password, confirm password

## Development

```bash
npm install
npm run dev        # Vite dev server with mock Keycloak context
npm run storybook  # Storybook at http://localhost:6006 for previewing individual pages
```

## Build

Requires [Maven](https://maven.apache.org/) (>= 3.1.1) and Java (>= 7) for JAR packaging.

```bash
npm run build-keycloak-theme
```

Produces two JARs in `dist_keycloak/`:

| File | Keycloak version |
|------|-----------------|
| `keycloak-theme-for-kc-22-to-25.jar` | Keycloak 22–25 |
| `keycloak-theme-for-kc-all-other-versions.jar` | All other versions |

## Deployment

1. **Build the theme:**
   ```bash
   npm run build-keycloak-theme
   ```

2. **Copy the JAR to Keycloak** (pick the right one for your Keycloak version):
   ```bash
   cp dist_keycloak/keycloak-theme-for-kc-22-to-25.jar /opt/keycloak/providers/
   ```
   For Docker, add a volume mount or `COPY` step in your Dockerfile:
   ```dockerfile
   COPY keycloak-theme/dist_keycloak/keycloak-theme-for-kc-22-to-25.jar /opt/keycloak/providers/
   ```

3. **Restart Keycloak** so it picks up the new provider:
   ```bash
   # Standalone
   /opt/keycloak/bin/kc.sh start

   # Docker
   docker restart <keycloak-container>
   ```

4. **Activate the theme** in the Keycloak admin console:
   - Go to your realm → **Realm settings** → **Themes**
   - Set **Login theme** to `keycloak-theme`
   - Save
