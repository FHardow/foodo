# Frontend Design Spec — Bread Order

**Date:** 2026-03-23
**Status:** Draft

---

## Overview

A mobile-first React frontend for a friend-facing bread ordering app. Friends browse available products, add them to a basket, place an order, then track its status and view their history. No prices are shown anywhere in the UI.

---

## Stack

| Concern | Library |
|---------|---------|
| Framework | React + TypeScript (Vite) |
| Routing | React Router v6 |
| Server state | TanStack Query |
| Client state | Zustand |
| Styling | Tailwind CSS |
| Toasts | sonner |

---

## Auth

Hardcoded test user UUID for the initial build, stored in `src/constants.ts`:

```ts
export const CURRENT_USER_ID = '<test-uuid>'
```

Used anywhere a `user_id` is required (order creation, order list filtering). When Keycloak is integrated, the backend must stop accepting `user_id` from the request body and derive it from the token instead — `CURRENT_USER_ID` will be replaced by the token subject.

---

## Visual Style

Warm & bakery: cream backgrounds (`#faf7f2`), brown text and accents (`#5c3d1e`, `#3d2b1a`), light warm borders (`#e8ddd0`). Buttons use dark brown as the primary colour.

---

## Navigation

Single nav pattern across all screen sizes:

- **Desktop:** top bar with logo, "Store" link, "History" link, and a basket icon showing item count
- **Mobile:** top bar shows logo + basket icon only; "View order history" text link appears at the bottom of the Basket page

Implemented in a single `Nav` component with responsive Tailwind classes.

---

## Pages

### Store (`/`)

- Fetches `GET /api/v1/products`, renders only `available: true` items
- Responsive product grid: 2 columns on mobile, 3–4 on desktop
- Each `ProductCard` shows: name, description, unit, and an "Add" button
- Tapping "Add" calls `useBasket.addItem(product)` — creates draft order if none exists, then adds the item; if the item is already in the basket, its quantity increments by 1
- Loading state: skeleton cards (`animate-pulse`)
- Error state: "Could not load products. Try again." with a retry button

### Basket (`/basket`)

- Reads the current basket order from the server (`GET /api/v1/orders/:id`) using the stored UUID
- Also fetches `GET /api/v1/products` (shared TanStack Query cache with Store page) to resolve `unit` per item by `product_id`; if the product cache is cold (direct navigation to `/basket`), this fetch runs automatically and `unit` shows a loading placeholder until resolved
- Shows items with name, unit, and quantity controls (+ / − / remove); the − button calls `removeItem` when quantity reaches 1 (never calls `updateQuantity` with 0)
- "Place Order" button → `useBasket.confirm()` → navigate to `/orders/:id`
- Empty state: "Your basket is empty" + "Browse products" link
- Mobile: "View order history" text link at the bottom of the page

### Order Status (`/orders/:id`)

- Fetches `GET /api/v1/orders/:id`, polls every 10 s via TanStack Query `refetchInterval`
- Polling stops automatically when status is `fulfilled` or `cancelled` (pass a function to `refetchInterval` that returns `false` for terminal statuses)
- Read-only: status badge, items list, `created_at` date
- Status badge colours: pending → amber, confirmed → blue, fulfilled → green, cancelled → red

### Order History (`/orders`)

- Fetches `GET /api/v1/orders?user_id=<CURRENT_USER_ID>`
- Sorted by `created_at` descending — client-side sort (backend does not guarantee order)
- Each row: status badge + created date, links to `/orders/:id`

---

## Types

```ts
type UUID = string

interface Product {
  id: UUID
  name: string
  description: string
  unit: string
  available: boolean
}

interface OrderItem {
  product_id: UUID
  product_name: string
  quantity: number
  // unit is not returned by the order API; cross-reference Product by product_id for display
  // API also returns unit_price_cents, total_cents — intentionally omitted
}

interface Order {
  id: UUID
  status: 'pending' | 'confirmed' | 'fulfilled' | 'cancelled'
  items: OrderItem[]
  created_at: string
}
```

All price fields returned by the backend (`unit_price_cents`, `total_cents`) are not included in frontend types.

---

## Basket Lifecycle

The basket is backed by a real server-side order in `pending` status. Its UUID is persisted in `localStorage` under the key `basketOrderId`.

### `useBasket` hook (wraps Zustand store + API calls)

**On app init:**
1. Read `basketOrderId` from localStorage
2. If present, validate with `GET /api/v1/orders/:id`:
   - 404 or status !== `pending` → clear `basketOrderId` from localStorage
3. Expose an `isValidating` boolean — the Store page "Add" button is disabled while validation is in progress to prevent race conditions

**`addItem(product: Product)`:**
1. Wait for `isValidating` to resolve
2. If no `basketOrderId` → `POST /api/v1/orders` body: `{ user_id: CURRENT_USER_ID }` → persist returned `id` to localStorage
   ⚠️ Requires backend prerequisite: `delivery_date` and `notes` must be removed from the `POST /orders` handler first
3. `POST /api/v1/orders/:id/items` body: `{ product_id: product.id, quantity: 1 }`
4. Backend auto-merges quantity if item already exists

**`removeItem(productId: UUID)`:**
- `DELETE /api/v1/orders/:id/items/:productId`

**`updateQuantity(productId: UUID, newQuantity: number)`:**
- `newQuantity` must be ≥ 1; the UI ensures the − button calls `removeItem` at quantity 1, so this function is never called with 0
1. `DELETE /api/v1/orders/:id/items/:productId` (remove existing)
2. `POST /api/v1/orders/:id/items` body: `{ product_id: productId, quantity: newQuantity }`
- (Remove-then-add because the backend merges quantities on `addItem` — there is no set-exact-quantity endpoint)

**`confirm()`:**
1. `POST /api/v1/orders/:id/confirm`
2. On success: clear `basketOrderId` from localStorage, navigate to `/orders/:id`

**Cache invalidation:**
- After any mutation (`addItem`, `removeItem`, `updateQuantity`), invalidate the `['order', id]` TanStack Query key so the basket page re-fetches the updated order

---

## Error Handling

| Scenario | Behaviour |
|----------|-----------|
| API error (generic) | Toast via `sonner`: "Something went wrong, try again" |
| Stale basket ID (404 or non-pending) | Silently clear localStorage; user starts fresh |
| Confirm failure | Show inline error in basket, do not navigate |
| Products fetch failure | Error message + retry button on Store page |
| Empty basket | "Your basket is empty" + link to Store |

---

## Backend Changes Required

Before starting frontend implementation, the following backend changes must be made:

1. **Remove `delivery_date` from `POST /api/v1/orders`** — strip from request body, domain `New()`, `Reconstitute()`, DB schema (`migrations/`), and order handler/response
2. **Remove `notes` from `POST /api/v1/orders`** — same scope

These fields can be reintroduced later when the UX for them is designed.

---

## Project Structure

```
frontend/
├── src/
│   ├── api/
│   │   ├── client.ts         # base fetch wrapper
│   │   ├── products.ts       # getProducts, getProduct
│   │   └── orders.ts         # getOrder, getOrders, createOrder, addItem,
│   │                         # removeItem, confirmOrder
│   ├── pages/
│   │   ├── Store.tsx
│   │   ├── Basket.tsx
│   │   ├── OrderStatus.tsx
│   │   └── OrderHistory.tsx
│   ├── components/
│   │   ├── Nav.tsx
│   │   ├── ProductCard.tsx
│   │   └── StatusBadge.tsx
│   ├── hooks/
│   │   └── useBasket.ts
│   ├── store/
│   │   └── basket.ts         # Zustand: basketOrderId (UUID | null), isValidating
│   ├── types/
│   │   └── index.ts
│   └── constants.ts          # CURRENT_USER_ID placeholder
```

---

## Out of Scope (for this iteration)

- Authentication (Keycloak) — deferred
- Prices — never shown in UI
- Delivery date and notes — deferred
- Admin views (product management, order fulfilment) — separate concern
- Real-time updates (WebSocket/SSE) — polling is sufficient
