# Order Status Push Notifications — Design Spec

**Date:** 2026-08-06
**Status:** Draft

---

## Problem

Customers have no way to know their order status changed (accepted / being prepared / ready) unless they keep the Order Status page (`frontend/src/pages/OrderStatus.tsx`) open and polling (`refetchInterval: 10_000`). The admin already gets a Telegram DM when a new order comes in (`backend/internal/infra/telegram/notifier.go`), but customers get nothing outside the app.

Goal: push a real OS-level notification to the customer's device when the admin moves their order on the kanban board (`frontend/src/pages/admin/Orders.tsx`) — even if the customer's browser tab or the site itself is closed.

---

## Scope

- Customer opts in to push notifications from the Order Status page.
- Backend pushes on forward status transitions only: **accept → accepted**, **start → ongoing**, **finish → finished**. Admin "undo" actions (unaccept/stop/unfinish) stay silent — those are corrections, not customer-facing progress.
- Multiple devices per customer supported (e.g. subscribed on phone and desktop).

Out of scope:
- Notifying the admin/owner via push (Telegram already covers this).
- Notifying on order creation/confirmation (customer is already looking at the page when they place the order).
- A native mobile app / iOS Safari without home-screen install (Web Push in a plain Safari tab isn't supported by iOS; PWA install is a possible future step, not this pass).
- Retry queues / delivery guarantees beyond a single best-effort send.

---

## Approach

Standard **Web Push**: browser Push API + Service Worker on the frontend, VAPID-signed payloads from the backend. This is the only way to reach a closed tab/site — there's no simpler substitute.

For the backend send path, two options:

| Option | Description | Verdict |
|---|---|---|
| **In-process, fire-and-forget** (chosen) | Same pattern already used for Telegram (`go s.notifier.OrderConfirmed(o)`): on successful status transition, spawn a goroutine that looks up the customer's subscriptions and sends. Failures are logged, not retried. | Matches existing codebase convention, no new infra, right-sized for current traffic (single Postgres, single backend instance). |
| Outbox + background worker (queue table, retry with backoff) | More resilient to transient push-service failures. | Overkill for current scale; nothing else in this codebase has a job queue. Rejected — YAGNI. |

---

## Backend

### New domain package: `internal/domain/push`

```go
type Subscription struct {
    id        ID
    userID    uuid.UUID
    endpoint  string
    p256dh    string
    auth      string
    createdAt time.Time
}

type Repository interface {
    Save(ctx context.Context, s *Subscription) error
    ListByUser(ctx context.Context, userID uuid.UUID) ([]*Subscription, error)
    DeleteByEndpoint(ctx context.Context, endpoint string) error
}
```

A subscription is one browser+device pairing. `Save` upserts on `endpoint` (re-subscribing the same browser replaces the old row, no dupes).

### Order domain: second notifier hook

`order.Service` currently has one `Notifier` (Telegram, admin-facing, fired only from `Confirm`). Add a second, customer-facing interface so Telegram and push stay independent and neither needs no-op methods for the other's events:

```go
// internal/domain/order/customer_notifier.go
type CustomerNotifier interface {
    OrderAccepted(o *Order)
    OrderStarted(o *Order)
    OrderFinished(o *Order)
}

type noopCustomerNotifier struct{}
func (noopCustomerNotifier) OrderAccepted(*Order)  {}
func (noopCustomerNotifier) OrderStarted(*Order)   {}
func (noopCustomerNotifier) OrderFinished(*Order)  {}
```

`Service` gets a `customerNotifier CustomerNotifier` field (defaults to noop) and a `WithCustomerNotifier(n)` builder, mirroring `WithNotifier`. `Accept`, `StartProgress`, and `Finish` each fire their event via `go s.customerNotifier.X(o)` after a successful save — same fire-and-forget shape as `Confirm`'s existing `go s.notifier.OrderConfirmed(o)`. `Unaccept`/`StopProgress`/`Unfinish` fire nothing.

### New infra: `internal/infra/webpush`

Implements `order.CustomerNotifier` using [`github.com/SherClockHolmes/webpush-go`](https://github.com/SherClockHolmes/webpush-go) (VAPID + RFC 8291 payload encryption — the standard Go library for this, no reason to hand-roll crypto).

```go
type Notifier struct {
    subs     push.Repository
    vapidPub string
    vapidPriv string
    subject  string // "mailto:owner@example.com", required by VAPID
}

func (n *Notifier) OrderAccepted(o *order.Order) { n.notify(o, "Order accepted", "Your order has been accepted.") }
func (n *Notifier) OrderStarted(o *order.Order)  { n.notify(o, "Being prepared", "Your order is being prepared.") }
func (n *Notifier) OrderFinished(o *order.Order) { n.notify(o, "Order ready", "Your order is ready!") }
```

`notify` looks up `subs.ListByUser(o.UserID())`, sends the JSON payload `{title, body, url: "/orders/<id>"}` to each subscription. On a `404`/`410` response (push service says the subscription is gone — browser uninstalled, permission revoked, etc.) it calls `DeleteByEndpoint` to clean up; other errors are logged via `slog` and otherwise ignored, matching the Telegram notifier's error handling.

### Migration

`backend/migrations/007_add_push_subscriptions.up.sql`:

```sql
CREATE TABLE push_subscriptions (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint   TEXT NOT NULL UNIQUE,
    p256dh     TEXT NOT NULL,
    auth       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_push_subscriptions_user_id ON push_subscriptions(user_id);
```

Plus matching `.down.sql` (`DROP TABLE push_subscriptions;`), and a GORM model in `internal/infra/postgres/models/push_subscription.go` following the existing `models.User`/`models.Order` pattern.

### Config & routes

- `config.go`: `VAPIDPublicKey`, `VAPIDPrivateKey`, `VAPIDSubject` read from env (`VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY`, `VAPID_SUBJECT`). Keys generated once with `webpush-go`'s keygen helper and stored as secrets like `TELEGRAM_BOT_TOKEN` today. Private key never leaves the backend.
- `.env.example`: add the three vars, plus `VITE_VAPID_PUBLIC_KEY` (public key only — safe to bake into the frontend build, same treatment as `VITE_KEYCLOAK_CLIENT_ID`).
- New routes under the existing authenticated `/api/v1` group in `router.go`:
  ```go
  push := v1.Group("/push")
  push.POST("/subscribe", pushHandler.Subscribe)
  push.DELETE("/subscribe", pushHandler.Unsubscribe)
  ```
  `Subscribe`/`Unsubscribe` read the user ID from `middleware.UserIDKey` (JWT-derived, same as `handler/order.go:68`) — **never** from the request body — so a client can't register a subscription against another user's account.

---

## Frontend

### Service worker — `frontend/public/sw.js`

Plain static file (Vite serves `public/` as-is, no build plugin needed):

```js
self.addEventListener('push', (event) => {
  const data = event.data.json()
  event.waitUntil(self.registration.showNotification(data.title, {
    body: data.body,
    data: { url: data.url },
  }))
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  event.waitUntil(clients.openWindow(event.notification.data.url))
})
```

### Subscribe flow — `frontend/src/push/subscribe.ts`

- `isPushSupported()` — feature-detects `'serviceWorker' in navigator && 'PushManager' in window` (false on iOS Safari outside a PWA — UI hides the opt-in entirely in that case rather than showing a button that will fail).
- `subscribeToPush()` — registers `/sw.js`, calls `Notification.requestPermission()`, then `pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: <VITE_VAPID_PUBLIC_KEY, base64→Uint8Array> })`, then POSTs the resulting `{endpoint, keys: {p256dh, auth}}` to `/api/v1/push/subscribe` via the existing `apiFetch` helper.
- `getExistingSubscription()` — `registration.pushManager.getSubscription()`, used to render the opt-in as already-granted on repeat visits.

### `frontend/src/api/push.ts`

```ts
export const subscribePush = (sub: PushSubscriptionJSON) =>
  apiFetch('/api/v1/push/subscribe', { method: 'POST', body: JSON.stringify(sub) })

export const unsubscribePush = (endpoint: string) =>
  apiFetch('/api/v1/push/subscribe', { method: 'DELETE', body: JSON.stringify({ endpoint }) })
```

### UI — `OrderStatus.tsx`

Small inline control near the status badge:
- Not supported → nothing rendered.
- Supported, not subscribed → button "Get notified about this order" → calls `subscribeToPush()`.
- Already subscribed → static "Notifications on" indicator, no action needed (subscription is account-wide, so it silently covers this and future orders once granted — no per-order re-prompt).
- Permission denied by the browser → button replaced with a muted "Notifications blocked — enable in browser settings" hint, not repeated nagging.

---

## Data flow

```
Admin drags card (Accept) on kanban
  → POST /api/v1/orders/:id/accept
  → order.Service.Accept() saves status=accepted
  → go customerNotifier.OrderAccepted(o)
      → push.Repository.ListByUser(o.UserID())
      → webpush.SendNotification(payload, sub, vapidOpts)  [per device]
      → 410/404 response → push.Repository.DeleteByEndpoint(sub.Endpoint)
  → customer's browser (any open or closed tab, service worker alive)
      → 'push' event → showNotification("Order accepted", ...)
      → tap → 'notificationclick' → opens /orders/:id
```

---

## Security

- VAPID private key stays server-side only; frontend only ever sees the public key.
- Subscribe/unsubscribe endpoints trust the JWT-derived user ID from `middleware.UserIDKey`, never a client-supplied user id — prevents subscribing/unsubscribing on someone else's behalf.
- Endpoint URLs (which double as bearer-like tokens for the push service) are stored server-side only, never returned in any list/get response beyond delete-by-caller's-own-endpoint.

---

## Testing

- Backend: unit tests for `push.Repository` (Postgres testcontainer, same pattern as `order_repo_test.go`), and for `order.Service` verifying `Accept`/`StartProgress`/`Finish` call the right `CustomerNotifier` method (fake notifier, same pattern likely already used for the Telegram `Notifier` interface in `service_test.go`) and that `Unaccept`/`StopProgress`/`Unfinish` call none.
- `webpush.Notifier` itself: unit test the payload/message text per method; skip testing actual delivery against a real push service (no reasonable way to do that in CI) — trust the library.
- Frontend: no realistic way to assert actual OS notifications in Playwright/CI (requires a real browser push subscription + HTTPS or localhost secure context + user gesture). Manual QA: subscribe in a real browser, drag a card on the kanban board, confirm the OS notification appears, including with the tab closed.
- Note for local dev: Push API requires a secure context — `localhost` is exempted by browsers, so dev works over plain HTTP on `localhost` without extra TLS setup.

---

## Open items resolved during design

- Trigger scope: customer's own drag-triggered action only was rejected in favor of the customer being notified (see conversation) — the person dragging the card (admin) is not the person who needs the notification.
- Only forward transitions notify; undo actions are silent.
- Opt-in lives on the Order Status page, not a separate global settings page — account-wide once granted.
