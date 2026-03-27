# Keycloak Theme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Keycloakify (React + Tailwind v4) login theme for the bread-order app with Login and Register pages styled to match the existing frontend.

**Architecture:** A standalone Vite project in `keycloak-theme/` scaffolded by the Keycloakify CLI, then modified with our custom pages. A shared `Template` component provides the centered white card + brand header; `Login` and `Register` pages compose it. `npm run build` emits a `.jar` for deployment to Keycloak.

**Tech Stack:** Keycloakify v11, React 19, Tailwind v4 (`@tailwindcss/vite`), TypeScript, Vite

---

## File Map

| File | Status | Responsibility |
|------|--------|----------------|
| `keycloak-theme/` | Create (scaffold) | Standalone Vite+Keycloakify project root |
| `keycloak-theme/vite.config.ts` | Modify | Add `@tailwindcss/vite` plugin |
| `keycloak-theme/src/index.css` | Modify | `@import "tailwindcss"` entry |
| `keycloak-theme/src/main.tsx` | Keep as-is | Dev entry — renders with mock KC context |
| `keycloak-theme/src/login/KcContext.ts` | Keep as-is | Auto-generated KC context union type |
| `keycloak-theme/src/login/i18n.ts` | Keep as-is | Auto-generated i18n hook |
| `keycloak-theme/src/login/KcPage.tsx` | Modify | Route `login.ftl` → Login, `register.ftl` → Register |
| `keycloak-theme/src/login/Template.tsx` | Create | Centered card wrapper + brand header |
| `keycloak-theme/src/login/pages/Login.tsx` | Replace | Login form (email, password, links) |
| `keycloak-theme/src/login/pages/Register.tsx` | Replace | Register form (name, email, password x2) |

---

## Task 1: Scaffold the Keycloakify project

**Files:**
- Create: `keycloak-theme/` (entire directory via CLI)

- [ ] **Step 1: Run the Keycloakify scaffolding CLI**

```bash
cd /path/to/bread-order
npx create-keycloakify@latest
```

When prompted:
- **Project name:** `keycloak-theme`
- **Theme type:** `login` (select only login, not account)
- **Bundler:** `vite`
- **UI framework:** `react` (no component library)

- [ ] **Step 2: Verify it builds**

```bash
cd keycloak-theme
npm install
npm run build
```

Expected: build succeeds, `dist_keycloak/` contains a `.jar` file.

- [ ] **Step 3: Verify dev server starts**

```bash
npm run dev
```

Expected: Vite dev server starts on `http://localhost:5173` (or similar). You should see the default Keycloakify login page with mock data.

Stop the dev server (`Ctrl+C`).

- [ ] **Step 4: Commit the scaffold**

```bash
cd ..
git add keycloak-theme/
git commit -m "chore: scaffold keycloakify login theme"
```

---

## Task 2: Add Tailwind v4

**Files:**
- Modify: `keycloak-theme/vite.config.ts`
- Modify: `keycloak-theme/src/index.css` (or create if not present)

- [ ] **Step 1: Install Tailwind v4**

```bash
cd keycloak-theme
npm install tailwindcss @tailwindcss/vite
```

- [ ] **Step 2: Add Tailwind plugin to vite.config.ts**

Open `keycloak-theme/vite.config.ts`. It will look something like:

```ts
import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import { keycloakify } from "keycloakify/vite-plugin"

export default defineConfig({
  plugins: [
    react(),
    keycloakify({ accountThemeImplementation: "none" })
  ]
})
```

Add the Tailwind import and plugin — result:

```ts
import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"
import { keycloakify } from "keycloakify/vite-plugin"

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    keycloakify({ accountThemeImplementation: "none" })
  ]
})
```

- [ ] **Step 3: Add Tailwind import to CSS**

Find the main CSS file imported in `src/main.tsx` (likely `src/index.css` or `src/main.css`). Replace its entire contents with:

```css
@import "tailwindcss";
```

If no CSS file is imported in `src/main.tsx`, create `src/index.css` with that content and add this import at the top of `src/main.tsx`:

```ts
import "./index.css"
```

- [ ] **Step 4: Verify Tailwind works**

```bash
npm run dev
```

Open the browser. In the dev tools console, verify no Tailwind-related errors. Stop the dev server.

- [ ] **Step 5: Commit**

```bash
cd ..
git add keycloak-theme/
git commit -m "chore: add tailwind v4 to keycloak theme"
```

---

## Task 3: Create the shared Template component

**Files:**
- Create: `keycloak-theme/src/login/Template.tsx`

- [ ] **Step 1: Create Template.tsx**

```tsx
// keycloak-theme/src/login/Template.tsx
import type { ReactNode } from "react"

type Props = {
  children: ReactNode
}

export default function Template({ children }: Props) {
  return (
    <div className="min-h-screen bg-[#faf7f2] flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md bg-white border border-[#e8ddd0] rounded-lg shadow-sm p-8">
        <div className="text-center mb-6">
          <p className="text-xl font-bold text-[#5c3d1e]">🍞 Bread Order</p>
          <p className="text-sm text-[#8a6a50] mt-1">Your local bakery</p>
        </div>
        {children}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd keycloak-theme
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd ..
git add keycloak-theme/src/login/Template.tsx
git commit -m "feat(theme): add shared Template card component"
```

---

## Task 4: Implement the Login page

**Files:**
- Replace: `keycloak-theme/src/login/pages/Login.tsx`

The login form posts to `kcContext.url.loginAction`. Keycloak expects the field names `username` (used for email when `loginWithEmailAllowed` is true) and `password`. Field errors come from `kcContext.messagesPerField`; the global message (e.g. "Invalid credentials") comes from `kcContext.message`.

- [ ] **Step 1: Replace Login.tsx**

Open `keycloak-theme/src/login/pages/Login.tsx`. The scaffold generated a version — replace its entire contents:

```tsx
// keycloak-theme/src/login/pages/Login.tsx
import type { KcContext } from "../KcContext"
import type { I18n } from "../i18n"
import Template from "../Template"

type LoginKcContext = Extract<KcContext, { pageId: "login.ftl" }>

type Props = {
  kcContext: LoginKcContext
  i18n: I18n
}

export default function Login({ kcContext }: Props) {
  const { url, realm, messagesPerField, message } = kcContext

  return (
    <Template>
      <h1 className="text-lg font-bold text-[#3d2b1a] mb-6">Sign in</h1>

      {message && message.type === "error" && (
        <p className="text-sm text-red-600 mb-4">{message.summary}</p>
      )}

      <form action={url.loginAction} method="post" className="space-y-4">
        <div>
          <label
            htmlFor="username"
            className="block text-sm font-semibold text-[#5c3d1e] mb-1"
          >
            Email
          </label>
          <input
            id="username"
            name="username"
            type="email"
            autoComplete="email"
            className="w-full bg-[#faf7f2] border border-[#e8ddd0] rounded-lg px-3 py-2 text-[#3d2b1a] text-sm focus:outline-none focus:ring-2 focus:ring-[#5c3d1e] focus:border-transparent"
          />
          {messagesPerField.existsError("username") && (
            <p className="text-xs text-red-600 mt-1">
              {messagesPerField.getFirstError("username")}
            </p>
          )}
        </div>

        <div>
          <label
            htmlFor="password"
            className="block text-sm font-semibold text-[#5c3d1e] mb-1"
          >
            Password
          </label>
          <input
            id="password"
            name="password"
            type="password"
            autoComplete="current-password"
            className="w-full bg-[#faf7f2] border border-[#e8ddd0] rounded-lg px-3 py-2 text-[#3d2b1a] text-sm focus:outline-none focus:ring-2 focus:ring-[#5c3d1e] focus:border-transparent"
          />
          {messagesPerField.existsError("password") && (
            <p className="text-xs text-red-600 mt-1">
              {messagesPerField.getFirstError("password")}
            </p>
          )}
        </div>

        <button
          type="submit"
          className="w-full bg-[#5c3d1e] text-white rounded-lg py-2.5 text-sm font-semibold hover:bg-[#3d2b1a] transition-colors mt-2"
        >
          Sign in
        </button>
      </form>

      <div className="flex justify-between mt-4">
        {realm.resetPasswordAllowed && (
          <a
            href={url.loginResetCredentialsUrl}
            className="text-sm text-[#5c3d1e] hover:underline"
          >
            Forgot password?
          </a>
        )}
        {realm.registrationAllowed && (
          <a
            href={url.registrationUrl}
            className="text-sm text-[#5c3d1e] hover:underline"
          >
            Create account
          </a>
        )}
      </div>
    </Template>
  )
}
```

- [ ] **Step 2: Fix any TypeScript errors**

```bash
cd keycloak-theme
npx tsc --noEmit
```

If `messagesPerField.getFirstError` does not exist on the type (Keycloakify versions vary), check the actual type definition:

```bash
grep -r "getFirstError\|messagesPerField" node_modules/keycloakify/login/KcContext.d.ts | head -20
```

Use whichever method name the type declares (may be `get` instead of `getFirstError`). Fix accordingly and re-run `tsc --noEmit` until clean.

- [ ] **Step 3: Preview in dev mode**

```bash
npm run dev
```

The dev server renders the Login page with mock data. Verify:
- Page background is `#faf7f2` cream
- White card centered on page
- `🍞 Bread Order` brand + tagline visible
- Email and password inputs styled correctly
- "Sign in" button is dark brown
- "Forgot password?" and "Create account" links visible at bottom

Stop the dev server.

- [ ] **Step 4: Commit**

```bash
cd ..
git add keycloak-theme/src/login/pages/Login.tsx
git commit -m "feat(theme): implement login page"
```

---

## Task 5: Implement the Register page

**Files:**
- Replace: `keycloak-theme/src/login/pages/Register.tsx`

The register form posts to `kcContext.url.registrationAction`. Keycloak expects field names: `firstName`, `lastName`, `email`, `password`, `password-confirm`. The link back to login is `kcContext.url.loginUrl`.

- [ ] **Step 1: Replace Register.tsx**

Open `keycloak-theme/src/login/pages/Register.tsx`. Replace its entire contents:

```tsx
// keycloak-theme/src/login/pages/Register.tsx
import type { KcContext } from "../KcContext"
import type { I18n } from "../i18n"
import Template from "../Template"

type RegisterKcContext = Extract<KcContext, { pageId: "register.ftl" }>

type Props = {
  kcContext: RegisterKcContext
  i18n: I18n
}

const inputClass =
  "w-full bg-[#faf7f2] border border-[#e8ddd0] rounded-lg px-3 py-2 text-[#3d2b1a] text-sm focus:outline-none focus:ring-2 focus:ring-[#5c3d1e] focus:border-transparent"

const labelClass = "block text-sm font-semibold text-[#5c3d1e] mb-1"

export default function Register({ kcContext }: Props) {
  const { url, messagesPerField, message } = kcContext

  return (
    <Template>
      <h1 className="text-lg font-bold text-[#3d2b1a] mb-6">Create account</h1>

      {message && message.type === "error" && (
        <p className="text-sm text-red-600 mb-4">{message.summary}</p>
      )}

      <form action={url.registrationAction} method="post" className="space-y-4">
        <div className="flex gap-3">
          <div className="flex-1">
            <label htmlFor="firstName" className={labelClass}>
              First name
            </label>
            <input
              id="firstName"
              name="firstName"
              type="text"
              autoComplete="given-name"
              className={inputClass}
            />
            {messagesPerField.existsError("firstName") && (
              <p className="text-xs text-red-600 mt-1">
                {messagesPerField.getFirstError("firstName")}
              </p>
            )}
          </div>
          <div className="flex-1">
            <label htmlFor="lastName" className={labelClass}>
              Last name
            </label>
            <input
              id="lastName"
              name="lastName"
              type="text"
              autoComplete="family-name"
              className={inputClass}
            />
            {messagesPerField.existsError("lastName") && (
              <p className="text-xs text-red-600 mt-1">
                {messagesPerField.getFirstError("lastName")}
              </p>
            )}
          </div>
        </div>

        <div>
          <label htmlFor="email" className={labelClass}>
            Email
          </label>
          <input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            className={inputClass}
          />
          {messagesPerField.existsError("email") && (
            <p className="text-xs text-red-600 mt-1">
              {messagesPerField.getFirstError("email")}
            </p>
          )}
        </div>

        <div>
          <label htmlFor="password" className={labelClass}>
            Password
          </label>
          <input
            id="password"
            name="password"
            type="password"
            autoComplete="new-password"
            className={inputClass}
          />
          {messagesPerField.existsError("password") && (
            <p className="text-xs text-red-600 mt-1">
              {messagesPerField.getFirstError("password")}
            </p>
          )}
        </div>

        <div>
          <label htmlFor="password-confirm" className={labelClass}>
            Confirm password
          </label>
          <input
            id="password-confirm"
            name="password-confirm"
            type="password"
            autoComplete="new-password"
            className={inputClass}
          />
          {messagesPerField.existsError("password-confirm") && (
            <p className="text-xs text-red-600 mt-1">
              {messagesPerField.getFirstError("password-confirm")}
            </p>
          )}
        </div>

        <button
          type="submit"
          className="w-full bg-[#5c3d1e] text-white rounded-lg py-2.5 text-sm font-semibold hover:bg-[#3d2b1a] transition-colors mt-2"
        >
          Create account
        </button>
      </form>

      <hr className="border-[#e8ddd0] my-5" />

      <p className="text-center text-sm">
        <a href={url.loginUrl} className="text-[#5c3d1e] hover:underline">
          Already have an account? Sign in
        </a>
      </p>
    </Template>
  )
}
```

- [ ] **Step 2: Fix any TypeScript errors**

```bash
cd keycloak-theme
npx tsc --noEmit
```

If `messagesPerField.getFirstError` doesn't exist on the type, use the same method name you confirmed in Task 4 Step 2. Fix and re-run until clean.

- [ ] **Step 3: Preview the Register page in dev mode**

The dev server defaults to previewing `login.ftl`. To preview the register page, find `src/main.tsx` and change the `pageId` in the mock context:

```tsx
// Change this line in src/main.tsx (find the getKcContextMock call):
// from:
const kcContext = getKcContextMock({ pageId: "login.ftl" })
// to:
const kcContext = getKcContextMock({ pageId: "register.ftl" })
```

```bash
npm run dev
```

Verify:
- First name + Last name inputs sit side by side
- Email, password, confirm password fields below
- "Create account" button is dark brown
- "Already have an account? Sign in" link below the divider

**Revert** `src/main.tsx` back to `login.ftl` after verifying.

Stop the dev server.

- [ ] **Step 4: Commit**

```bash
cd ..
git add keycloak-theme/src/login/pages/Register.tsx keycloak-theme/src/main.tsx
git commit -m "feat(theme): implement register page"
```

---

## Task 6: Wire KcPage.tsx and do a final build

**Files:**
- Modify: `keycloak-theme/src/login/KcPage.tsx`

The scaffold generates a `KcPage.tsx` that already has routing logic. We need to make sure our `Login` and `Register` components are used for the relevant pages instead of the default ones.

- [ ] **Step 1: Check the current KcPage.tsx**

Open `keycloak-theme/src/login/KcPage.tsx`. The scaffold generates a switch on `kcContext.pageId`. Look for the `"login.ftl"` and `"register.ftl"` cases.

If they already import from `./pages/Login` and `./pages/Register`, the scaffold wired them up automatically — no changes needed, skip to Step 3.

If they fall through to a `<DefaultPage>` fallback, update those two cases to use our components. Find the switch statement and update:

```tsx
// Add these imports at the top if not already present:
import { lazy } from "react"
const Login = lazy(() => import("./pages/Login"))
const Register = lazy(() => import("./pages/Register"))

// In the switch, ensure these cases exist:
case "login.ftl":
  return (
    <Suspense>
      <Login kcContext={kcContext} i18n={i18n} />
    </Suspense>
  )
case "register.ftl":
  return (
    <Suspense>
      <Register kcContext={kcContext} i18n={i18n} />
    </Suspense>
  )
```

Note: `i18n` comes from `useI18n({ kcContext })` already in the scaffold — do not add a second call.

- [ ] **Step 2: Verify TypeScript is clean**

```bash
cd keycloak-theme
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Final production build**

```bash
npm run build
```

Expected: build succeeds. Check that a `.jar` file exists:

```bash
ls dist_keycloak/*.jar
```

Expected output: something like `dist_keycloak/keycloak-theme-1.0.0.jar`

- [ ] **Step 4: Commit**

```bash
cd ..
git add keycloak-theme/src/login/KcPage.tsx
git commit -m "feat(theme): wire KcPage routing for login and register"
```

---

## Deployment Reference

After the above tasks are complete, to deploy the theme to a running Keycloak instance:

```bash
# Build
cd keycloak-theme && npm run build

# Copy the jar (adjust Keycloak path as needed)
cp dist_keycloak/*.jar /opt/keycloak/providers/

# Restart Keycloak
# (Docker: docker restart <keycloak-container>)

# Then in Keycloak admin console:
# Realm Settings → Themes → Login theme → select "keycloak-theme" (or your theme name)
# Save
```

The theme name is set in `keycloak-theme/package.json` under the `keycloakify.themeName` field — defaults to the package name.
