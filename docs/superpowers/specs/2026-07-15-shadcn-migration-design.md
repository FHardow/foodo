# shadcn Migration Design Spec — Foodo Frontend

**Date:** 2026-07-15
**Status:** Draft

---

## Problem

Reported on mobile phone use:

1. **Drag-and-drop broken.** The admin Orders kanban board (`frontend/src/pages/admin/Orders.tsx`) only registers `PointerSensor` from `@dnd-kit/core`. On touch devices the browser's native touch-scroll gesture wins over drag activation, so cards can't be dragged between columns.
2. **No mobile nav.** `Nav.tsx` hides all links except Store/basket below the `sm` breakpoint (`hidden sm:block`), with no replacement — there is no way to reach History, Orders, or Products on a phone.

Decision: rather than patch just these two, migrate the whole frontend UI to [shadcn/ui](https://ui.shadcn.com) (Radix primitives + Tailwind) for a consistent, accessible, touch-friendly component system going forward, and fix both bugs as part of that migration.

---

## Scope

Full migration, single pass (big-bang), covering:

- shadcn foundation (init, theme, icon library)
- `Nav.tsx` — mobile burger menu
- `admin/Orders.tsx` — kanban touch drag-and-drop fix + re-skin
- All remaining pages/components: `Store.tsx`, `Product.tsx`, `ProductCard.tsx`, `Basket.tsx`, `OrderHistory.tsx`, `OrderStatus.tsx`, `StatusBadge.tsx`, `admin/Products.tsx`

Out of scope: backend changes, new features/routes, dark mode (not requested), keycloak-theme package (separate app, untouched).

---

## Foundation

| Concern | Decision |
|---|---|
| Install | `npx shadcn@latest init` — Tailwind v4 mode, CSS variables, "new-york" style |
| Theme | CSS variables remapped to existing bakery palette: primary `#5c3d1e`, muted/border `#e8ddd0`, white surfaces. No default shadcn slate/zinc theme — brand colors carry over exactly |
| Icons | `lucide-react` added. All emoji (currently just 🛒 in `Nav.tsx`) replaced with icon components |
| Toasts | Already on `sonner` (`frontend/package.json`) — shadcn's toast wraps `sonner`, so no swap needed, just adopt shadcn's `<Toaster>` wrapper/styling if it differs from current usage |
| Components pulled in | `button`, `card`, `badge`, `input`, `select`, `dialog`, `sheet`, `label`, `form`, `separator` |

Components are copied into `frontend/src/components/ui/` by the shadcn CLI (standard shadcn convention — not an npm dependency, source lives in-repo).

---

## Nav: mobile burger menu

- **Desktop (`≥sm`):** unchanged — horizontal link row (Store, History, Orders/Products if owner)
- **Mobile (`<sm`):** burger icon (lucide `Menu`) opens a shadcn `Sheet` sliding in from the side, listing the same links stacked vertically
- Cart link + item count stays outside the sheet, always visible/tappable — basket access shouldn't require opening the menu
- Owner-only links (`Orders`, `Products`) keep their `keycloak.hasRealmRole('owner')` gate, in both desktop row and mobile sheet

---

## Kanban drag-and-drop fix

Root cause: `useSensors` in `admin/Orders.tsx:147` only includes `PointerSensor`. Fix:

- Add `TouchSensor` alongside `PointerSensor`, with `activationConstraint: { delay: 150, tolerance: 5 }` so a quick tap/scroll isn't mistaken for a drag
- Set `touch-action: none` on the draggable card wrapper so the browser doesn't intercept the gesture for native scrolling mid-drag
- Columns re-skinned with shadcn `Card`; status pills use `Badge`
- **Test-id contract preserved exactly:** `data-testid="kanban-column"` and `data-testid="column-count"` keep their current values/placement so `frontend/e2e/kanban.spec.ts` continues to pass unmodified

---

## Remaining pages

Rebuilt on shadcn primitives, same visual language (bakery palette), no behavior changes beyond the components used to render them:

| File | shadcn components used |
|---|---|
| `ProductCard.tsx` | `Card` |
| `Store.tsx`, `Product.tsx` | `Card`, `Button` |
| `Basket.tsx` | `Card`, `Button` (quantity +/−, Remove, Place Order) |
| `OrderHistory.tsx`, `OrderStatus.tsx` | `Card`, `Badge` (via `StatusBadge`) |
| `StatusBadge.tsx` | `Badge` |
| `admin/Products.tsx` | `Card`, `Button`, `Input`, `Select`, `Label`, `Form`, and its hand-rolled create/edit modal (`modalOpen` state, lines ~213+) replaced with shadcn `Dialog` |

**e2e safety constraint:** every `getByRole`/`getByText` target currently asserted in `frontend/e2e/basket.spec.ts` and `kanban.spec.ts` — button names ("Place Order", "−", "+", "Remove"), heading text ("Your Basket", "Order Board"), link text ("Browse products"), item text — must resolve identically after migration. shadcn's `Button`/`Card`/etc. render real semantic elements (`<button>`, etc.), so accessible names carry over as long as visible text/aria-labels are kept unchanged.

---

## Testing

- `frontend/e2e/*.spec.ts` (Playwright) must pass unmodified except where explicitly noted (kanban touch drag may need a new test, additive only)
- `frontend/src/test` (vitest/RTL) — any component-level tests updated only if they assert on DOM structure that shadcn changes (e.g. class names); behavior-level assertions should be unaffected
- Manual check: kanban drag-and-drop on an actual touch device or Chrome DevTools touch emulation, since Playwright's default driver doesn't exercise real touch events the same way
