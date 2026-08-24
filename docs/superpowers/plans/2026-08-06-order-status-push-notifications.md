# Order Status Push Notifications Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Push a real OS-level notification to a customer's device when the admin moves their order on the kanban board (accept/start/finish), even if the customer's browser/tab is closed.

**Architecture:** Standard Web Push — browser Push API + a static service worker on the frontend, VAPID-signed sends from the backend via `github.com/SherClockHolmes/webpush-go`. `order.Service` gains a second, customer-facing `CustomerNotifier` hook (parallel to the existing Telegram `Notifier`), fired fire-and-forget from `Accept`/`StartProgress`/`Finish`. Subscriptions are stored per-user in a new `push_subscriptions` Postgres table, with `endpoint`/`p256dh`/`auth` encrypted at rest (AES-256-GCM).

**Tech Stack:** Go 1.26 / Gin / GORM / Postgres (backend), React 19 / TypeScript / Vite / TanStack Query / vitest (frontend). New dependency: `github.com/SherClockHolmes/webpush-go`.

**Spec:** `docs/superpowers/specs/2026-08-06-order-status-push-notifications-design.md`

## Global Constraints

- Push fires only on forward transitions: `Accept` → "Order accepted", `StartProgress` → "Being prepared", `Finish` → "Order ready!". `Unaccept`/`StopProgress`/`Unfinish` never notify.
- No push on order creation/confirmation. No push to the admin/owner (Telegram already covers that channel).
- Backend send path is in-process, fire-and-forget (`go` + log-on-error), matching the existing Telegram notifier. No retry queue.
- `endpoint`, `p256dh`, `auth` are encrypted at rest with AES-256-GCM before hitting Postgres. Key is `PUSH_ENCRYPTION_KEY` (32 random bytes, base64), **mandatory** at backend startup (`config.Load()` fails fast if missing/invalid, same pattern as the existing DSN check).
- A separate `endpoint_hash` (hex SHA-256 of the plaintext endpoint) column exists purely for equality lookups (upsert-on-resubscribe, delete-on-410), because AES-GCM ciphertext is non-deterministic and can't carry a `UNIQUE` constraint.
- `/api/v1/push/subscribe` derives the user ID from the JWT (`middleware.UserIDKey`), never from the request body.
- Opt-in UI lives only on the Order Status page (`frontend/src/pages/OrderStatus.tsx`); the subscription itself is account-wide once granted.
- Multiple devices per customer are supported (one row per browser/device).

---

## Task 1: Fix pre-existing compile error blocking `internal/infra/postgres` tests

**Context:** `go vet ./...` currently fails on this repo with two pre-existing, unrelated errors:
1. `internal/infra/http/handler/order_test.go` — references an old domain API (`Fulfill`/`Cancel`) that no longer exists. This is a wholesale-obsolete test file; rewriting it is out of scope for this feature (see Task 8's note).
2. `internal/infra/postgres/order_repo_test.go` — three calls to `order.AddItem(...)` are missing the `unit` argument that was added to that method's signature after these tests were written. This is a trivial 3-line signature-drift fix.

Because all `_test.go` files in a directory compile into one test binary regardless of `package X` vs `package X_test`, error #2 blocks *every* test we're about to add under `internal/infra/postgres/` (Task 5's `crypto_test.go`, Task 6's `push_subscription_repo_test.go`). It must be fixed first. Error #1 is left alone — different package, much larger, unrelated to this feature.

**Files:**
- Modify: `backend/internal/infra/postgres/order_repo_test.go:52`, `:111`, `:145`

**Interfaces:**
- Consumes: `order.Order.AddItem(productID product.ID, productName, unit string, quantity int, unitPriceCents int64) error` (existing, unchanged)
- Produces: a compiling `internal/infra/postgres` test binary for later tasks to build on

- [ ] **Step 1: Confirm the current failure**

Run: `cd backend && go vet ./internal/infra/postgres/...`
Expected: FAIL — `order_repo_test.go:52:73: not enough arguments in call to o.AddItem`

- [ ] **Step 2: Fix the three call sites**

In `backend/internal/infra/postgres/order_repo_test.go`, add the missing `unit` argument (use `"loaf"`, matching the convention already used in `internal/domain/order/service_test.go`'s `seededProduct` helper):

```go
// line 52 — was: o.AddItem(product.ID(productID), "Sourdough", 3, 450)
require.NoError(t, o.AddItem(product.ID(productID), "Sourdough", "loaf", 3, 450))
```

```go
// line 111 — was: o.AddItem(product.ID(productID), "Rye", 2, 300)
require.NoError(t, o.AddItem(product.ID(productID), "Rye", "loaf", 2, 300))
```

```go
// line 145 — was: o.AddItem(product.ID(productID), "Sourdough", 1, 450)
require.NoError(t, o.AddItem(product.ID(productID), "Sourdough", "loaf", 1, 450))
```

- [ ] **Step 3: Verify the fix**

Run: `go vet ./internal/infra/postgres/...`
Expected: PASS, no output.

Run: `go test ./internal/infra/postgres/... -short`
Expected: PASS (short mode skips the testcontainer-based tests via `newTestDB`'s `testing.Short()` check, but confirms the package still compiles and any short tests pass).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/infra/postgres/order_repo_test.go
git commit -m "fix: update stale AddItem calls in order_repo_test.go to current signature"
```

---

## Task 2: `push` domain — Subscription entity + Repository interface

**Files:**
- Create: `backend/internal/domain/push/subscription.go`
- Create: `backend/internal/domain/push/repository.go`
- Test: `backend/internal/domain/push/subscription_test.go`

**Interfaces:**
- Produces: `push.Subscription` (getters: `ID() push.ID`, `UserID() uuid.UUID`, `Endpoint() string`, `P256dh() string`, `Auth() string`, `CreatedAt() time.Time`), `push.New(userID uuid.UUID, endpoint, p256dh, auth string) (*Subscription, error)`, `push.Reconstitute(id, userID uuid.UUID, endpoint, p256dh, auth string, createdAt time.Time) *Subscription`, `push.Repository` interface (`Save`, `ListByUser`, `DeleteByEndpoint`)

- [ ] **Step 1: Write the failing tests**

Create `backend/internal/domain/push/subscription_test.go`:

```go
package push_test

import (
	"testing"

	"github.com/fhardow/foodo/internal/domain/push"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Success(t *testing.T) {
	userID := uuid.New()
	sub, err := push.New(userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)
	assert.Equal(t, userID, sub.UserID())
	assert.Equal(t, "https://push.example.com/abc", sub.Endpoint())
	assert.Equal(t, "p256dh-key", sub.P256dh())
	assert.Equal(t, "auth-key", sub.Auth())
	assert.NotEqual(t, uuid.Nil, sub.ID())
}

func TestNew_RequiresUserID(t *testing.T) {
	_, err := push.New(uuid.Nil, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_RequiresEndpoint(t *testing.T) {
	_, err := push.New(uuid.New(), "", "p256dh-key", "auth-key")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_RequiresP256dh(t *testing.T) {
	_, err := push.New(uuid.New(), "https://push.example.com/abc", "", "auth-key")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_RequiresAuth(t *testing.T) {
	_, err := push.New(uuid.New(), "https://push.example.com/abc", "p256dh-key", "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/domain/push/... -v`
Expected: FAIL — `package push is not in GOROOT` / `undefined: push.New` (package doesn't exist yet).

- [ ] **Step 3: Implement the domain type**

Create `backend/internal/domain/push/subscription.go`:

```go
package push

import (
	"time"

	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
)

type ID = uuid.UUID

// Subscription is one browser+device push registration for a customer.
type Subscription struct {
	id        ID
	userID    uuid.UUID
	endpoint  string
	p256dh    string
	auth      string
	createdAt time.Time
}

func New(userID uuid.UUID, endpoint, p256dh, auth string) (*Subscription, error) {
	if userID == uuid.Nil {
		return nil, domerrors.BadRequest("user ID is required")
	}
	if endpoint == "" {
		return nil, domerrors.BadRequest("endpoint is required")
	}
	if p256dh == "" {
		return nil, domerrors.BadRequest("p256dh is required")
	}
	if auth == "" {
		return nil, domerrors.BadRequest("auth is required")
	}
	return &Subscription{
		id:        uuid.New(),
		userID:    userID,
		endpoint:  endpoint,
		p256dh:    p256dh,
		auth:      auth,
		createdAt: time.Now().UTC(),
	}, nil
}

// Reconstitute rebuilds a Subscription from persistence without re-running validation.
func Reconstitute(id, userID uuid.UUID, endpoint, p256dh, auth string, createdAt time.Time) *Subscription {
	return &Subscription{
		id:        id,
		userID:    userID,
		endpoint:  endpoint,
		p256dh:    p256dh,
		auth:      auth,
		createdAt: createdAt,
	}
}

func (s *Subscription) ID() ID               { return s.id }
func (s *Subscription) UserID() uuid.UUID    { return s.userID }
func (s *Subscription) Endpoint() string     { return s.endpoint }
func (s *Subscription) P256dh() string       { return s.p256dh }
func (s *Subscription) Auth() string         { return s.auth }
func (s *Subscription) CreatedAt() time.Time { return s.createdAt }
```

Create `backend/internal/domain/push/repository.go`:

```go
package push

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, s *Subscription) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Subscription, error)
	DeleteByEndpoint(ctx context.Context, endpoint string) error
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/domain/push/... -v`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/push/subscription.go backend/internal/domain/push/repository.go backend/internal/domain/push/subscription_test.go
git commit -m "feat: add push.Subscription domain entity and Repository interface"
```

---

## Task 3: `push` domain — Service + mock repository

**Files:**
- Create: `backend/internal/domain/push/service.go`
- Test: `backend/internal/domain/push/service_test.go`
- Modify: `backend/internal/testutil/mock/repos.go`

**Interfaces:**
- Consumes: `push.Subscription`, `push.Repository`, `push.New` (Task 2)
- Produces: `push.Service` (`push.NewService(repo Repository) *Service`, `(*Service).Subscribe(ctx, userID uuid.UUID, endpoint, p256dh, auth string) (*Subscription, error)`, `(*Service).Unsubscribe(ctx, endpoint string) error`), `mock.PushRepo` (`mock.NewPushRepo() *PushRepo`, fields `ErrSave`, `ErrListByUser`, `ErrDeleteByEndpoint`)

- [ ] **Step 1: Add the mock repository**

In `backend/internal/testutil/mock/repos.go`, add `"github.com/fhardow/foodo/internal/domain/push"` to the import block, then append at the end of the file:

```go
type PushRepo struct {
	mu sync.RWMutex

	ErrSave             error
	ErrListByUser       error
	ErrDeleteByEndpoint error

	subs map[string]*push.Subscription // keyed by endpoint
}

func NewPushRepo() *PushRepo {
	return &PushRepo{subs: make(map[string]*push.Subscription)}
}

func (r *PushRepo) Save(_ context.Context, s *push.Subscription) error {
	if r.ErrSave != nil {
		return r.ErrSave
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subs[s.Endpoint()] = s
	return nil
}

func (r *PushRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]*push.Subscription, error) {
	if r.ErrListByUser != nil {
		return nil, r.ErrListByUser
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*push.Subscription, 0)
	for _, s := range r.subs {
		if s.UserID() == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *PushRepo) DeleteByEndpoint(_ context.Context, endpoint string) error {
	if r.ErrDeleteByEndpoint != nil {
		return r.ErrDeleteByEndpoint
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.subs, endpoint)
	return nil
}
```

- [ ] **Step 2: Write the failing service tests**

Create `backend/internal/domain/push/service_test.go`:

```go
package push_test

import (
	"context"
	"testing"

	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/testutil/mock"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Subscribe_Success(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)

	userID := uuid.New()
	sub, err := svc.Subscribe(context.Background(), userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)
	assert.Equal(t, userID, sub.UserID())

	found, err := repo.ListByUser(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "https://push.example.com/abc", found[0].Endpoint())
}

func TestService_Subscribe_ValidationError(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)

	_, err := svc.Subscribe(context.Background(), uuid.Nil, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestService_Subscribe_Resubscribe_ReplacesOldEntry(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)
	userID := uuid.New()

	_, err := svc.Subscribe(context.Background(), userID, "https://push.example.com/abc", "old-p256dh", "old-auth")
	require.NoError(t, err)
	_, err = svc.Subscribe(context.Background(), userID, "https://push.example.com/abc", "new-p256dh", "new-auth")
	require.NoError(t, err)

	found, err := repo.ListByUser(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, found, 1, "resubscribing the same endpoint must not create a duplicate")
	assert.Equal(t, "new-p256dh", found[0].P256dh())
}

func TestService_Unsubscribe_Success(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)
	userID := uuid.New()

	_, err := svc.Subscribe(context.Background(), userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)

	require.NoError(t, svc.Unsubscribe(context.Background(), "https://push.example.com/abc"))

	found, err := repo.ListByUser(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, found)
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/domain/push/... -v`
Expected: FAIL — `undefined: push.NewService`.

- [ ] **Step 4: Implement the service**

Create `backend/internal/domain/push/service.go`:

```go
package push

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Subscribe(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth string) (*Subscription, error) {
	sub, err := New(userID, endpoint, p256dh, auth)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *Service) Unsubscribe(ctx context.Context, endpoint string) error {
	return s.repo.DeleteByEndpoint(ctx, endpoint)
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/domain/push/... -v`
Expected: PASS (9 tests total).

Run: `go build ./...` from `backend/` to confirm `testutil/mock` still compiles cleanly.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/push/service.go backend/internal/domain/push/service_test.go backend/internal/testutil/mock/repos.go
git commit -m "feat: add push.Service and mock.PushRepo"
```

---

## Task 4: `order.CustomerNotifier` — wire the accept/start/finish hook

**Files:**
- Create: `backend/internal/domain/order/customer_notifier.go`
- Modify: `backend/internal/domain/order/service.go`
- Modify: `backend/internal/domain/order/service_test.go`

**Interfaces:**
- Produces: `order.CustomerNotifier` interface (`OrderAccepted(*Order)`, `OrderStarted(*Order)`, `OrderFinished(*Order)`), `(*order.Service).WithCustomerNotifier(n CustomerNotifier) *Service`
- This is what Task 7 (`webpush.Notifier`) will implement, and what Task 8 wires up in `main.go`.

- [ ] **Step 1: Write the failing tests**

In `backend/internal/domain/order/service_test.go`, add `"sync"` and `"time"` to the import block, then append at the end of the file:

```go
// ---------------------------------------------------------------------------
// Customer push notifications
// ---------------------------------------------------------------------------

type fakeCustomerNotifier struct {
	mu       sync.Mutex
	accepted []order.ID
	started  []order.ID
	finished []order.ID
}

func (f *fakeCustomerNotifier) OrderAccepted(o *order.Order) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.accepted = append(f.accepted, o.ID())
}

func (f *fakeCustomerNotifier) OrderStarted(o *order.Order) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started = append(f.started, o.ID())
}

func (f *fakeCustomerNotifier) OrderFinished(o *order.Order) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.finished = append(f.finished, o.ID())
}

func (f *fakeCustomerNotifier) acceptedCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.accepted)
}

func (f *fakeCustomerNotifier) startedCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.started)
}

func (f *fakeCustomerNotifier) finishedCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.finished)
}

// progressToOngoing drives a fresh order through Confirm/Accept/StartProgress via the service.
func progressToOngoing(t *testing.T, svc *order.Service, uRepo *mock.UserRepo, pRepo *mock.ProductRepo) *order.Order {
	t.Helper()
	o := createOrder(t, svc, uRepo)
	pid := seededProduct(t, pRepo, true)
	_, err := svc.AddItem(context.Background(), o.ID(), pid, 1)
	require.NoError(t, err)
	_, err = svc.Confirm(context.Background(), o.ID())
	require.NoError(t, err)
	_, err = svc.Accept(context.Background(), o.ID())
	require.NoError(t, err)
	_, err = svc.StartProgress(context.Background(), o.ID())
	require.NoError(t, err)
	return o
}

func TestOrderService_Accept_NotifiesCustomer(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)
	fake := &fakeCustomerNotifier{}
	svc.WithCustomerNotifier(fake)

	o := createOrder(t, svc, uRepo)
	pid := seededProduct(t, pRepo, true)
	_, err := svc.AddItem(context.Background(), o.ID(), pid, 1)
	require.NoError(t, err)
	_, err = svc.Confirm(context.Background(), o.ID())
	require.NoError(t, err)

	_, err = svc.Accept(context.Background(), o.ID())
	require.NoError(t, err)

	require.Eventually(t, func() bool { return fake.acceptedCount() == 1 }, time.Second, 10*time.Millisecond)
}

func TestOrderService_StartProgress_NotifiesCustomer(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)
	fake := &fakeCustomerNotifier{}
	svc.WithCustomerNotifier(fake)

	o := createOrder(t, svc, uRepo)
	pid := seededProduct(t, pRepo, true)
	_, err := svc.AddItem(context.Background(), o.ID(), pid, 1)
	require.NoError(t, err)
	_, err = svc.Confirm(context.Background(), o.ID())
	require.NoError(t, err)
	_, err = svc.Accept(context.Background(), o.ID())
	require.NoError(t, err)

	_, err = svc.StartProgress(context.Background(), o.ID())
	require.NoError(t, err)

	require.Eventually(t, func() bool { return fake.startedCount() == 1 }, time.Second, 10*time.Millisecond)
}

func TestOrderService_Finish_NotifiesCustomer(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)
	fake := &fakeCustomerNotifier{}
	svc.WithCustomerNotifier(fake)

	o := progressToOngoing(t, svc, uRepo, pRepo)

	_, err := svc.Finish(context.Background(), o.ID())
	require.NoError(t, err)

	require.Eventually(t, func() bool { return fake.finishedCount() == 1 }, time.Second, 10*time.Millisecond)
}

func TestOrderService_UndoTransitions_DoNotNotifyCustomer(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)
	fake := &fakeCustomerNotifier{}
	svc.WithCustomerNotifier(fake)

	o := progressToOngoing(t, svc, uRepo, pRepo)
	_, err := svc.Finish(context.Background(), o.ID())
	require.NoError(t, err)
	require.Eventually(t, func() bool { return fake.finishedCount() == 1 }, time.Second, 10*time.Millisecond)

	_, err = svc.Unfinish(context.Background(), o.ID())
	require.NoError(t, err)
	_, err = svc.StopProgress(context.Background(), o.ID())
	require.NoError(t, err)
	_, err = svc.Unaccept(context.Background(), o.ID())
	require.NoError(t, err)

	// Give any stray goroutine a chance to fire before asserting silence.
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, fake.acceptedCount(), "unaccept must not fire a second accepted notification")
	assert.Equal(t, 1, fake.startedCount(), "stop must not fire a second started notification")
	assert.Equal(t, 1, fake.finishedCount(), "unfinish must not fire a second finished notification")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/domain/order/... -run NotifiesCustomer -v`
Expected: FAIL — `svc.WithCustomerNotifier undefined`.

- [ ] **Step 3: Implement `CustomerNotifier` and wire it into the service**

Create `backend/internal/domain/order/customer_notifier.go`:

```go
package order

// CustomerNotifier is called after order transitions the customer should be
// told about. Unlike Notifier (admin-facing, Telegram), this fires only on
// forward status progress: accepted, started, finished.
type CustomerNotifier interface {
	OrderAccepted(o *Order)
	OrderStarted(o *Order)
	OrderFinished(o *Order)
}

type noopCustomerNotifier struct{}

func (noopCustomerNotifier) OrderAccepted(*Order) {}
func (noopCustomerNotifier) OrderStarted(*Order)  {}
func (noopCustomerNotifier) OrderFinished(*Order) {}
```

In `backend/internal/domain/order/service.go`, update the struct and constructor:

```go
type Service struct {
	repo             Repository
	productRepo      product.Repository
	userRepo         user.Repository
	notifier         Notifier
	customerNotifier CustomerNotifier
}

func NewService(repo Repository, productRepo product.Repository, userRepo user.Repository) *Service {
	return &Service{
		repo:             repo,
		productRepo:      productRepo,
		userRepo:         userRepo,
		notifier:         noopNotifier{},
		customerNotifier: noopCustomerNotifier{},
	}
}

func (s *Service) WithNotifier(n Notifier) *Service {
	s.notifier = n
	return s
}

func (s *Service) WithCustomerNotifier(n CustomerNotifier) *Service {
	s.customerNotifier = n
	return s
}
```

Add a `go s.customerNotifier.X(o)` call at the end of each of these three methods, right before `return o, nil` (after the existing `s.repo.Save` check):

```go
func (s *Service) Accept(ctx context.Context, id ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := o.Accept(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	go s.customerNotifier.OrderAccepted(o)
	return o, nil
}
```

```go
func (s *Service) StartProgress(ctx context.Context, id ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := o.StartProgress(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	go s.customerNotifier.OrderStarted(o)
	return o, nil
}
```

```go
func (s *Service) Finish(ctx context.Context, id ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := o.Finish(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	go s.customerNotifier.OrderFinished(o)
	return o, nil
}
```

Leave `Unaccept`, `StopProgress`, and `Unfinish` untouched — no notifier call.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/domain/order/... -v`
Expected: PASS, including all pre-existing tests in this file (no regressions) plus the 4 new ones.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/order/customer_notifier.go backend/internal/domain/order/service.go backend/internal/domain/order/service_test.go
git commit -m "feat: add order.CustomerNotifier hook on accept/start/finish"
```

---

## Task 5: Postgres — AES-256-GCM field encryption helpers

**Files:**
- Create: `backend/internal/infra/postgres/crypto.go`
- Test: `backend/internal/infra/postgres/crypto_test.go`

**Interfaces:**
- Produces (unexported, package-internal): `encryptField(key []byte, plaintext string) (string, error)`, `decryptField(key []byte, encoded string) (string, error)`, `hashEndpoint(endpoint string) string`
- Consumed by Task 6's `push_subscription_repo.go`.

- [ ] **Step 1: Write the failing tests**

Create `backend/internal/infra/postgres/crypto_test.go`:

```go
package postgres

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKey() []byte {
	return bytes.Repeat([]byte("k"), 32)
}

func TestEncryptDecryptField_RoundTrip(t *testing.T) {
	key := testKey()
	ciphertext, err := encryptField(key, "https://push.example.com/abc123")
	require.NoError(t, err)
	assert.NotEqual(t, "https://push.example.com/abc123", ciphertext, "ciphertext must not equal plaintext")

	plaintext, err := decryptField(key, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, "https://push.example.com/abc123", plaintext)
}

func TestEncryptField_NonDeterministic(t *testing.T) {
	key := testKey()
	a, err := encryptField(key, "same input")
	require.NoError(t, err)
	b, err := encryptField(key, "same input")
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "AES-GCM uses a random nonce, so repeated encryption of the same plaintext must differ")
}

func TestDecryptField_WrongKeyFails(t *testing.T) {
	ciphertext, err := encryptField(testKey(), "secret value")
	require.NoError(t, err)

	wrongKey := bytes.Repeat([]byte("x"), 32)
	_, err = decryptField(wrongKey, ciphertext)
	assert.Error(t, err)
}

func TestHashEndpoint_Deterministic(t *testing.T) {
	a := hashEndpoint("https://push.example.com/abc123")
	b := hashEndpoint("https://push.example.com/abc123")
	assert.Equal(t, a, b)
	assert.NotEqual(t, a, hashEndpoint("https://push.example.com/different"))
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/infra/postgres/... -run "EncryptDecrypt|NonDeterministic|WrongKey|HashEndpoint" -v`
Expected: FAIL — `undefined: encryptField`.

- [ ] **Step 3: Implement the helpers**

Create `backend/internal/infra/postgres/crypto.go`:

```go
package postgres

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// encryptField encrypts plaintext with AES-256-GCM using a random nonce,
// returning base64(nonce || ciphertext).
func encryptField(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptField reverses encryptField.
func decryptField(key []byte, encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// hashEndpoint is a deterministic SHA-256 hex digest, used only for equality
// lookups against the non-deterministic encrypted endpoint column.
func hashEndpoint(endpoint string) string {
	sum := sha256.Sum256([]byte(endpoint))
	return hex.EncodeToString(sum[:])
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/infra/postgres/... -run "EncryptDecrypt|NonDeterministic|WrongKey|HashEndpoint" -v`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add backend/internal/infra/postgres/crypto.go backend/internal/infra/postgres/crypto_test.go
git commit -m "feat: add AES-256-GCM field encryption helpers for at-rest secrets"
```

---

## Task 6: Postgres — `push_subscriptions` migration, model, and repository

**Files:**
- Create: `backend/migrations/007_add_push_subscriptions.up.sql`
- Create: `backend/migrations/007_add_push_subscriptions.down.sql`
- Create: `backend/internal/infra/postgres/models/push_subscription.go`
- Create: `backend/internal/infra/postgres/push_subscription_repo.go`
- Modify: `backend/internal/infra/postgres/testhelper_test.go`
- Test: `backend/internal/infra/postgres/push_subscription_repo_test.go`

**Interfaces:**
- Consumes: `push.Subscription`, `push.Repository`, `push.Reconstitute` (Task 2); `encryptField`/`decryptField`/`hashEndpoint` (Task 5)
- Produces: `postgres.NewPushSubscriptionRepo(db *gorm.DB, encryptionKey string) (push.Repository, error)` — consumed by Task 8's `main.go` wiring.

- [ ] **Step 1: Add the migration**

Create `backend/migrations/007_add_push_subscriptions.up.sql`:

```sql
CREATE TABLE push_subscriptions (
    id            UUID PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint      TEXT NOT NULL,        -- AES-256-GCM ciphertext, base64
    endpoint_hash TEXT NOT NULL UNIQUE, -- SHA-256(endpoint), hex — lookup/upsert key
    p256dh        TEXT NOT NULL,        -- ciphertext, base64
    auth          TEXT NOT NULL,        -- ciphertext, base64
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_push_subscriptions_user_id ON push_subscriptions(user_id);
```

Create `backend/migrations/007_add_push_subscriptions.down.sql`:

```sql
DROP TABLE push_subscriptions;
```

- [ ] **Step 2: Add the GORM model**

Create `backend/internal/infra/postgres/models/push_subscription.go`:

```go
package models

import "time"

type PushSubscription struct {
	ID           string    `gorm:"primaryKey;type:uuid"`
	UserID       string    `gorm:"not null;index;type:uuid"`
	Endpoint     string    `gorm:"not null"`
	EndpointHash string    `gorm:"not null;uniqueIndex;column:endpoint_hash"`
	P256dh       string    `gorm:"not null;column:p256dh"`
	Auth         string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (PushSubscription) TableName() string { return "push_subscriptions" }
```

- [ ] **Step 3: Register the model in the test AutoMigrate list**

In `backend/internal/infra/postgres/testhelper_test.go`, add `&models.PushSubscription{}` to the `db.AutoMigrate(...)` call:

```go
	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.PushSubscription{},
	)
```

- [ ] **Step 4: Write the failing repository tests**

Create `backend/internal/infra/postgres/push_subscription_repo_test.go`:

```go
package postgres_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/fhardow/foodo/internal/domain/push"
	repopostgres "github.com/fhardow/foodo/internal/infra/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func testEncryptionKey() string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("k"), 32))
}

func newTestPushRepo(t *testing.T, db *gorm.DB) push.Repository {
	t.Helper()
	repo, err := repopostgres.NewPushSubscriptionRepo(db, testEncryptionKey())
	require.NoError(t, err)
	return repo
}

func TestPushSubscriptionRepo_SaveAndListByUser(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)
	ctx := context.Background()

	userID := uuid.New()
	sub, err := push.New(userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub))

	found, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "https://push.example.com/abc", found[0].Endpoint())
	assert.Equal(t, "p256dh-key", found[0].P256dh())
	assert.Equal(t, "auth-key", found[0].Auth())
}

func TestPushSubscriptionRepo_ListByUser_Empty(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)

	found, err := repo.ListByUser(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestPushSubscriptionRepo_Save_UpsertsOnEndpoint(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)
	ctx := context.Background()

	userID := uuid.New()
	sub1, err := push.New(userID, "https://push.example.com/abc", "old-p256dh", "old-auth")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub1))

	sub2, err := push.New(userID, "https://push.example.com/abc", "new-p256dh", "new-auth")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub2))

	found, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, found, 1, "re-subscribing the same endpoint must upsert, not duplicate")
	assert.Equal(t, "new-p256dh", found[0].P256dh())
}

func TestPushSubscriptionRepo_DeleteByEndpoint(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)
	ctx := context.Background()

	userID := uuid.New()
	sub, err := push.New(userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub))

	require.NoError(t, repo.DeleteByEndpoint(ctx, "https://push.example.com/abc"))

	found, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestPushSubscriptionRepo_DeleteByEndpoint_NonExistentIsNoOp(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)

	err := repo.DeleteByEndpoint(context.Background(), "https://push.example.com/does-not-exist")
	assert.NoError(t, err)
}

func TestPushSubscriptionRepo_Save_EncryptsAtRest(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)
	ctx := context.Background()

	userID := uuid.New()
	sub, err := push.New(userID, "https://push.example.com/plaintext-check", "p256dh-key", "auth-key")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub))

	var raw struct {
		Endpoint string
		P256dh   string
		Auth     string
	}
	require.NoError(t, db.Raw(
		`SELECT endpoint AS endpoint, p256dh AS p256dh, auth AS auth FROM push_subscriptions WHERE user_id = ?`,
		userID.String(),
	).Scan(&raw).Error)

	assert.NotEqual(t, "https://push.example.com/plaintext-check", raw.Endpoint, "endpoint must be encrypted at rest")
	assert.NotEqual(t, "p256dh-key", raw.P256dh, "p256dh must be encrypted at rest")
	assert.NotEqual(t, "auth-key", raw.Auth, "auth must be encrypted at rest")
}
```

- [ ] **Step 5: Run tests to verify they fail**

Run: `go test ./internal/infra/postgres/... -run TestPushSubscriptionRepo -v`
Expected: FAIL — `undefined: repopostgres.NewPushSubscriptionRepo`.

- [ ] **Step 6: Implement the repository**

Create `backend/internal/infra/postgres/push_subscription_repo.go`:

```go
package postgres

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/infra/postgres/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pushSubscriptionRepo struct {
	db  *gorm.DB
	key []byte
}

// NewPushSubscriptionRepo decodes encryptionKey (base64, must be exactly 32
// bytes for AES-256) once at construction time and fails fast if it's invalid.
func NewPushSubscriptionRepo(db *gorm.DB, encryptionKey string) (push.Repository, error) {
	key, err := base64.StdEncoding.DecodeString(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decode push encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("push encryption key must decode to 32 bytes, got %d", len(key))
	}
	return &pushSubscriptionRepo{db: db, key: key}, nil
}

func (r *pushSubscriptionRepo) Save(ctx context.Context, s *push.Subscription) error {
	encEndpoint, err := encryptField(r.key, s.Endpoint())
	if err != nil {
		return err
	}
	encP256dh, err := encryptField(r.key, s.P256dh())
	if err != nil {
		return err
	}
	encAuth, err := encryptField(r.key, s.Auth())
	if err != nil {
		return err
	}
	m := models.PushSubscription{
		ID:           s.ID().String(),
		UserID:       s.UserID().String(),
		Endpoint:     encEndpoint,
		EndpointHash: hashEndpoint(s.Endpoint()),
		P256dh:       encP256dh,
		Auth:         encAuth,
		CreatedAt:    s.CreatedAt(),
	}
	return dbFromCtx(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "endpoint_hash"}},
		DoUpdates: clause.AssignmentColumns([]string{"user_id", "endpoint", "p256dh", "auth"}),
	}).Create(&m).Error
}

func (r *pushSubscriptionRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*push.Subscription, error) {
	var ms []models.PushSubscription
	if err := dbFromCtx(ctx, r.db).Where("user_id = ?", userID.String()).Find(&ms).Error; err != nil {
		return nil, err
	}
	subs := make([]*push.Subscription, 0, len(ms))
	for i := range ms {
		s, err := r.toDomain(&ms[i])
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *pushSubscriptionRepo) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	return dbFromCtx(ctx, r.db).Where("endpoint_hash = ?", hashEndpoint(endpoint)).Delete(&models.PushSubscription{}).Error
}

func (r *pushSubscriptionRepo) toDomain(m *models.PushSubscription) (*push.Subscription, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(m.UserID)
	if err != nil {
		return nil, err
	}
	endpoint, err := decryptField(r.key, m.Endpoint)
	if err != nil {
		return nil, err
	}
	p256dh, err := decryptField(r.key, m.P256dh)
	if err != nil {
		return nil, err
	}
	auth, err := decryptField(r.key, m.Auth)
	if err != nil {
		return nil, err
	}
	return push.Reconstitute(id, userID, endpoint, p256dh, auth, m.CreatedAt), nil
}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test ./internal/infra/postgres/... -run TestPushSubscriptionRepo -v`
Expected: PASS (6 tests). This spins up a real Postgres testcontainer (Docker required) — if Docker isn't available, this will fail to start the container rather than fail the assertions; that's an environment issue, not a code issue.

- [ ] **Step 8: Commit**

```bash
git add backend/migrations/007_add_push_subscriptions.up.sql backend/migrations/007_add_push_subscriptions.down.sql backend/internal/infra/postgres/models/push_subscription.go backend/internal/infra/postgres/push_subscription_repo.go backend/internal/infra/postgres/push_subscription_repo_test.go backend/internal/infra/postgres/testhelper_test.go
git commit -m "feat: add push_subscriptions table and encrypted Postgres repository"
```

---

## Task 7: `internal/infra/webpush` — VAPID push sender

**Files:**
- Create: `backend/internal/infra/webpush/notifier.go`
- Test: `backend/internal/infra/webpush/notifier_test.go`

**Interfaces:**
- Consumes: `order.CustomerNotifier` (Task 4), `push.Repository`, `push.New` (Task 2/3)
- Produces: `webpush.NewNotifier(subs push.Repository, vapidPub, vapidPriv, subject string) *Notifier`, implementing `order.CustomerNotifier` — consumed by Task 8's `main.go` wiring.

- [ ] **Step 1: Add the dependency**

Run (from `backend/`): `go get github.com/SherClockHolmes/webpush-go`
Expected: `go.mod`/`go.sum` updated with the new dependency.

- [ ] **Step 2: Write the failing tests**

Create `backend/internal/infra/webpush/notifier_test.go` (white-box — same `package webpush` as the implementation, so tests can construct `&Notifier{...}` directly with a fake `send` function, avoiding any real network calls):

```go
package webpush

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	sherwebpush "github.com/SherClockHolmes/webpush-go"
	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/testutil/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestOrder(t *testing.T, userID uuid.UUID) *order.Order {
	t.Helper()
	o, err := order.New(userID)
	require.NoError(t, err)
	return o
}

func seedSubscription(t *testing.T, repo *mock.PushRepo, userID uuid.UUID, endpoint string) {
	t.Helper()
	sub, err := push.New(userID, endpoint, "p256dh-key", "auth-key")
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), sub))
}

func TestOrderAccepted_SendsExpectedPayloadToEachSubscription(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/1")
	seedSubscription(t, repo, userID, "https://push.example.com/2")

	var sentMessages [][]byte
	var sentEndpoints []string
	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			sentMessages = append(sentMessages, message)
			sentEndpoints = append(sentEndpoints, s.Endpoint)
			return &http.Response{StatusCode: http.StatusCreated, Body: http.NoBody}, nil
		},
		vapidPub:  "pub",
		vapidPriv: "priv",
		subject:   "mailto:owner@example.com",
	}

	o := newTestOrder(t, userID)
	n.OrderAccepted(o)

	require.Len(t, sentMessages, 2)
	assert.ElementsMatch(t, []string{"https://push.example.com/1", "https://push.example.com/2"}, sentEndpoints)

	var payload notificationPayload
	require.NoError(t, json.Unmarshal(sentMessages[0], &payload))
	assert.Equal(t, "Order accepted", payload.Title)
	assert.Equal(t, "Your order has been accepted.", payload.Body)
	assert.Equal(t, "/orders/"+o.ID().String(), payload.URL)
}

func TestOrderStarted_SendsExpectedPayload(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/1")

	var sent notificationPayload
	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			require.NoError(t, json.Unmarshal(message, &sent))
			return &http.Response{StatusCode: http.StatusCreated, Body: http.NoBody}, nil
		},
	}

	n.OrderStarted(newTestOrder(t, userID))

	assert.Equal(t, "Being prepared", sent.Title)
	assert.Equal(t, "Your order is being prepared.", sent.Body)
}

func TestOrderFinished_SendsExpectedPayload(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/1")

	var sent notificationPayload
	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			require.NoError(t, json.Unmarshal(message, &sent))
			return &http.Response{StatusCode: http.StatusCreated, Body: http.NoBody}, nil
		},
	}

	n.OrderFinished(newTestOrder(t, userID))

	assert.Equal(t, "Order ready", sent.Title)
	assert.Equal(t, "Your order is ready!", sent.Body)
}

func TestNotify_DeletesSubscriptionOnGone(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/gone")

	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusGone, Body: http.NoBody}, nil
		},
	}

	n.OrderAccepted(newTestOrder(t, userID))

	remaining, err := repo.ListByUser(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, remaining, "subscription must be deleted after a 410 Gone response")
}

func TestNotify_SendErrorDoesNotPanic(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New()
	seedSubscription(t, repo, userID, "https://push.example.com/broken")

	n := &Notifier{
		subs: repo,
		send: func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error) {
			return nil, errors.New("network unreachable")
		},
	}

	assert.NotPanics(t, func() { n.OrderAccepted(newTestOrder(t, userID)) })
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/infra/webpush/... -v`
Expected: FAIL — `undefined: Notifier` (package doesn't exist yet).

- [ ] **Step 4: Implement the notifier**

Create `backend/internal/infra/webpush/notifier.go`:

```go
package webpush

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	sherwebpush "github.com/SherClockHolmes/webpush-go"
	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/push"
)

type sendFunc func(message []byte, s *sherwebpush.Subscription, options *sherwebpush.Options) (*http.Response, error)

type Notifier struct {
	subs      push.Repository
	send      sendFunc
	vapidPub  string
	vapidPriv string
	subject   string
}

func NewNotifier(subs push.Repository, vapidPub, vapidPriv, subject string) *Notifier {
	return &Notifier{
		subs:      subs,
		send:      sherwebpush.SendNotification,
		vapidPub:  vapidPub,
		vapidPriv: vapidPriv,
		subject:   subject,
	}
}

type notificationPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

func (n *Notifier) OrderAccepted(o *order.Order) {
	n.notify(o, "Order accepted", "Your order has been accepted.")
}

func (n *Notifier) OrderStarted(o *order.Order) {
	n.notify(o, "Being prepared", "Your order is being prepared.")
}

func (n *Notifier) OrderFinished(o *order.Order) {
	n.notify(o, "Order ready", "Your order is ready!")
}

func (n *Notifier) notify(o *order.Order, title, body string) {
	ctx := context.Background()
	subs, err := n.subs.ListByUser(ctx, o.UserID())
	if err != nil {
		slog.Error("push: failed to list subscriptions", "err", err, "order_id", o.ID())
		return
	}
	msg, err := json.Marshal(notificationPayload{Title: title, Body: body, URL: "/orders/" + o.ID().String()})
	if err != nil {
		slog.Error("push: failed to marshal payload", "err", err)
		return
	}
	for _, s := range subs {
		resp, err := n.send(msg, &sherwebpush.Subscription{
			Endpoint: s.Endpoint(),
			Keys:     sherwebpush.Keys{P256dh: s.P256dh(), Auth: s.Auth()},
		}, &sherwebpush.Options{
			Subscriber:      n.subject,
			VAPIDPublicKey:  n.vapidPub,
			VAPIDPrivateKey: n.vapidPriv,
			TTL:             30,
		})
		if err != nil {
			slog.Error("push: send failed", "err", err, "order_id", o.ID())
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
			if delErr := n.subs.DeleteByEndpoint(ctx, s.Endpoint()); delErr != nil {
				slog.Error("push: failed to delete stale subscription", "err", delErr)
			}
		}
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/infra/webpush/... -v`
Expected: PASS (5 tests).

Run: `go build ./...` to confirm `order.CustomerNotifier` is satisfied by `*webpush.Notifier` (compile-time interface check — there's no explicit `var _ order.CustomerNotifier = (*Notifier)(nil)` assertion in the code above; add one to catch drift early):

In `backend/internal/infra/webpush/notifier.go`, add near the top (after the `Notifier` struct definition):

```go
var _ order.CustomerNotifier = (*Notifier)(nil)
```

Run `go build ./...` again to confirm this compiles.

- [ ] **Step 6: Commit**

```bash
git add backend/go.mod backend/go.sum backend/internal/infra/webpush/notifier.go backend/internal/infra/webpush/notifier_test.go
git commit -m "feat: add webpush.Notifier implementing order.CustomerNotifier via VAPID"
```

---

## Task 8: Config, HTTP handler, routes, and `main.go` wiring

**Files:**
- Modify: `backend/internal/config/config.go`
- Create: `backend/internal/infra/http/handler/push.go`
- Test: `backend/internal/infra/http/handler/push_test.go`
- Modify: `backend/internal/infra/http/router.go`
- Modify: `backend/cmd/api/main.go`

**Interfaces:**
- Consumes: `push.Service` (Task 3), `webpush.NewNotifier`/`postgres.NewPushSubscriptionRepo` (Tasks 6/7), `order.Service.WithCustomerNotifier` (Task 4), the `seedUser` test helper and shared `postJSON`/`assertErrorBody` helpers already in `package handler_test` (added when `order_test.go` was rewritten to fix a pre-existing, unrelated compile break — see the plan's ledger/history; that fix landed before this task, so `internal/infra/http/handler` compiles and its tests run cleanly now)
- Produces: `handler.PushHandler` (`handler.NewPushHandler(svc *push.Service) *PushHandler`, methods `Subscribe`/`Unsubscribe`), routes `POST /api/v1/push/subscribe` and `DELETE /api/v1/push/subscribe` — consumed by Task 12's frontend `api/push.ts`.

- [ ] **Step 1: Add config fields with fail-fast validation**

In `backend/internal/config/config.go`, add `"encoding/base64"` to the imports, add fields to `Config`, populate them in `Load()`, and validate `PushEncryptionKey`:

```go
type Config struct {
	Env               string
	Port              string
	DSN               string
	KeycloakURL       string
	KeycloakRealm     string
	CORSOrigin        string
	TelegramBotToken  string
	TelegramChatID    string
	VAPIDPublicKey    string
	VAPIDPrivateKey   string
	VAPIDSubject      string
	PushEncryptionKey string
}
```

```go
func Load() (*Config, error) {
	cfg := &Config{
		Env:               getEnv("ENV", "development"),
		Port:              getEnv("PORT", "8080"),
		KeycloakURL:       getEnv("KEYCLOAK_URL", "http://localhost:8180"),
		KeycloakRealm:     getEnv("KEYCLOAK_REALM", "foodo"),
		CORSOrigin:        os.Getenv("CORS_ORIGIN"),
		TelegramBotToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:    os.Getenv("TELEGRAM_CHAT_ID"),
		VAPIDPublicKey:    os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey:   os.Getenv("VAPID_PRIVATE_KEY"),
		VAPIDSubject:      os.Getenv("VAPID_SUBJECT"),
		PushEncryptionKey: os.Getenv("PUSH_ENCRYPTION_KEY"),
	}

	cfg.DSN = buildDSN()
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database configuration is required")
	}

	if cfg.PushEncryptionKey == "" {
		return nil, fmt.Errorf("PUSH_ENCRYPTION_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(cfg.PushEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("PUSH_ENCRYPTION_KEY must be valid base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("PUSH_ENCRYPTION_KEY must decode to 32 bytes, got %d", len(key))
	}

	return cfg, nil
}
```

- [ ] **Step 2: Verify config compiles**

Run: `go build ./...`
Expected: Success. (No `config_test.go` exists in this codebase, matching the existing convention of not unit-testing this thin env-var loader — Task 9 will make sure the new env vars are actually documented/plumbed so this fail-fast check doesn't break local dev.)

- [ ] **Step 3: Add the HTTP handler**

Create `backend/internal/infra/http/handler/push.go`:

```go
package handler

import (
	"net/http"

	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/infra/http/middleware"
	"github.com/fhardow/foodo/internal/infra/http/respond"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PushHandler struct {
	svc *push.Service
}

func NewPushHandler(svc *push.Service) *PushHandler {
	return &PushHandler{svc: svc}
}

type subscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
	Keys     struct {
		P256dh string `json:"p256dh" binding:"required"`
		Auth   string `json:"auth" binding:"required"`
	} `json:"keys" binding:"required"`
}

type unsubscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
}

func (h *PushHandler) Subscribe(c *gin.Context) {
	subRaw, _ := c.Get(middleware.UserIDKey)
	sub, _ := subRaw.(string)
	userID, err := uuid.Parse(sub)
	if err != nil {
		respond.Error(c, domerrors.BadRequest("invalid user ID in token"))
		return
	}
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	if _, err := h.svc.Subscribe(c.Request.Context(), userID, req.Endpoint, req.Keys.P256dh, req.Keys.Auth); err != nil {
		respond.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PushHandler) Unsubscribe(c *gin.Context) {
	var req unsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Error(c, domerrors.BadRequest("%s", err.Error()))
		return
	}
	if err := h.svc.Unsubscribe(c.Request.Context(), req.Endpoint); err != nil {
		respond.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
```

- [ ] **Step 4: Write and run `PushHandler` tests**

Create `backend/internal/infra/http/handler/push_test.go` (same `package handler_test` as `order_test.go`/`product_test.go`/`user_test.go` — reuses their shared `postJSON` and `assertErrorBody` helpers without redefining them):

```go
package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/infra/http/handler"
	"github.com/fhardow/foodo/internal/infra/http/middleware"
	"github.com/fhardow/foodo/internal/testutil/mock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPushRouter(repo *mock.PushRepo, authUserID string) *gin.Engine {
	svc := push.NewService(repo)
	h := handler.NewPushHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, authUserID)
		c.Next()
	})
	r.POST("/push/subscribe", h.Subscribe)
	r.DELETE("/push/subscribe", h.Unsubscribe)

	return r
}

func deleteJSON(router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodDelete, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func validSubscribeBody() map[string]any {
	return map[string]any{
		"endpoint": "https://push.example.com/abc",
		"keys": map[string]any{
			"p256dh": "p256dh-key",
			"auth":   "auth-key",
		},
	}
}

func TestPushHandler_Subscribe_Success(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New().String()
	router := setupPushRouter(repo, userID)

	w := postJSON(router, "/push/subscribe", validSubscribeBody())
	assert.Equal(t, http.StatusNoContent, w.Code)

	parsedUserID, err := uuid.Parse(userID)
	require.NoError(t, err)
	found, err := repo.ListByUser(context.Background(), parsedUserID)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "https://push.example.com/abc", found[0].Endpoint())
}

func TestPushHandler_Subscribe_InvalidUserID(t *testing.T) {
	router := setupPushRouter(mock.NewPushRepo(), "not-a-uuid")

	w := postJSON(router, "/push/subscribe", validSubscribeBody())
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushHandler_Subscribe_MissingFields(t *testing.T) {
	router := setupPushRouter(mock.NewPushRepo(), uuid.New().String())

	w := postJSON(router, "/push/subscribe", map[string]any{"endpoint": "https://push.example.com/abc"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorBody(t, w)
}

func TestPushHandler_Unsubscribe_Success(t *testing.T) {
	repo := mock.NewPushRepo()
	userID := uuid.New().String()
	router := setupPushRouter(repo, userID)

	postJSON(router, "/push/subscribe", validSubscribeBody())

	w := deleteJSON(router, "/push/subscribe", map[string]any{"endpoint": "https://push.example.com/abc"})
	assert.Equal(t, http.StatusNoContent, w.Code)

	parsedUserID, err := uuid.Parse(userID)
	require.NoError(t, err)
	found, err := repo.ListByUser(context.Background(), parsedUserID)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestPushHandler_Unsubscribe_MissingEndpoint(t *testing.T) {
	router := setupPushRouter(mock.NewPushRepo(), uuid.New().String())

	w := deleteJSON(router, "/push/subscribe", map[string]any{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

Note: `deleteJSON` needs `"bytes"` added to the import block (used for `bytes.NewReader(b)`, same as `postJSON`'s pattern in `user_test.go:66-73`).

Run: `go test ./internal/infra/http/handler/... -run TestPushHandler -v`
Expected: PASS (6 tests).

Run: `go test ./internal/infra/http/handler/...`
Expected: all tests in the package pass (this package compiles cleanly now — the `order_test.go` rewrite that unblocked it landed as a prerequisite fix before this task).

- [ ] **Step 5: Register the routes**

In `backend/internal/infra/http/router.go`, add `pushHandler *handler.PushHandler` as a parameter to `NewRouter` (right after `orders`):

```go
func NewRouter(
	users *handler.UserHandler,
	products *handler.ProductHandler,
	orders *handler.OrderHandler,
	pushHandler *handler.PushHandler,
	userSvc *user.Service,
	keycloakURL string,
	keycloakRealm string,
	uploadsDir string,
	corsOrigin string,
) http.Handler {
```

Inside the authenticated `v1` group, after the `o := v1.Group("/orders")` block, add:

```go
		push := v1.Group("/push")
		push.POST("/subscribe", pushHandler.Subscribe)
		push.DELETE("/subscribe", pushHandler.Unsubscribe)
```

- [ ] **Step 6: Wire it up in `main.go`**

In `backend/cmd/api/main.go`, add imports for `"github.com/fhardow/foodo/internal/domain/push"` and `"github.com/fhardow/foodo/internal/infra/webpush"`.

Update the repositories/services/handlers section:

```go
	// Repositories
	userRepo    := postgres.NewUserRepo(db)
	productRepo := postgres.NewProductRepo(db)
	orderRepo   := postgres.NewOrderRepo(db)
	pushRepo, err := postgres.NewPushSubscriptionRepo(db, cfg.PushEncryptionKey)
	if err != nil {
		log.Error("failed to init push subscription repo", "err", err)
		os.Exit(1)
	}

	// Domain services
	userSvc    := user.NewService(userRepo)
	productSvc := product.NewService(productRepo)
	orderSvc   := order.NewService(orderRepo, productRepo, userRepo)
	pushSvc    := push.NewService(pushRepo)
	if cfg.TelegramBotToken != "" && cfg.TelegramChatID != "" {
		orderSvc.WithNotifier(telegram.NewNotifier(cfg.TelegramBotToken, cfg.TelegramChatID))
		log.Info("telegram order notifications enabled")
	}
	if cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != "" {
		orderSvc.WithCustomerNotifier(webpush.NewNotifier(pushRepo, cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject))
		log.Info("push order notifications enabled")
	}

	// HTTP handlers
	userHandler    := handler.NewUserHandler(userSvc)
	productHandler := handler.NewProductHandler(productSvc, uploadsDir)
	orderHandler   := handler.NewOrderHandler(orderSvc)
	pushHandler    := handler.NewPushHandler(pushSvc)

	router := apphttp.NewRouter(userHandler, productHandler, orderHandler, pushHandler, userSvc, cfg.KeycloakURL, cfg.KeycloakRealm, uploadsDir, cfg.CORSOrigin)
```

- [ ] **Step 7: Verify it builds**

Run: `go build ./...`
Expected: Success.

Run: `go vet ./internal/domain/... ./internal/infra/postgres/... ./internal/infra/webpush/... ./internal/infra/http/...`
Expected: clean.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/config/config.go backend/internal/infra/http/handler/push.go backend/internal/infra/http/handler/push_test.go backend/internal/infra/http/router.go backend/cmd/api/main.go
git commit -m "feat: wire push subscribe/unsubscribe endpoints and customer notifier"
```

---

## Task 9: Ops — env vars and Docker plumbing

**Files:**
- Modify: `.env.example`
- Modify: `docker-compose.prod.yml`
- Modify: `frontend/Dockerfile`

**Interfaces:**
- Consumes: env var names defined in Task 8 (`VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY`, `VAPID_SUBJECT`, `PUSH_ENCRYPTION_KEY`) and Task 12 (`VITE_VAPID_PUBLIC_KEY`).
- No code interfaces — this task is pure configuration.

**Note:** After Task 8, the backend refuses to start without `PUSH_ENCRYPTION_KEY` set. If you have a local `backend/.env` (gitignored, not checked into this repo) used for `make run`, add the var there too — generate one with `openssl rand -base64 32`.

- [ ] **Step 1: Update `.env.example`**

Append to `.env.example` (after the existing `VITE_KEYCLOAK_CLIENT_ID=foodo-app` line):

```
# Web Push — customer notifications on order status changes.
# Generate keys with: npx web-push generate-vapid-keys
VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=
VAPID_SUBJECT=mailto:you@example.com
# 32 random bytes, base64-encoded. Generate with: openssl rand -base64 32
PUSH_ENCRYPTION_KEY=

# Baked into the frontend at build time — same value as VAPID_PUBLIC_KEY above
VITE_VAPID_PUBLIC_KEY=
```

- [ ] **Step 2: Pass the vars through in `docker-compose.prod.yml`**

In the `backend` service's `environment` block, after `TELEGRAM_CHAT_ID: ${TELEGRAM_CHAT_ID:-}`:

```yaml
      VAPID_PUBLIC_KEY: ${VAPID_PUBLIC_KEY:-}
      VAPID_PRIVATE_KEY: ${VAPID_PRIVATE_KEY:-}
      VAPID_SUBJECT: ${VAPID_SUBJECT:-}
      PUSH_ENCRYPTION_KEY: ${PUSH_ENCRYPTION_KEY}
```

(No `:-` fallback on `PUSH_ENCRYPTION_KEY` — it's mandatory, same style as `DB_PASSWORD` above it in this file, so a missing value fails loudly rather than silently deploying a broken backend.)

In the `frontend` service's `build.args` block, after `VITE_KEYCLOAK_CLIENT_ID: ${KEYCLOAK_CLIENT_ID:-foodo-frontend}`:

```yaml
        VITE_VAPID_PUBLIC_KEY: ${VAPID_PUBLIC_KEY:-}
```

(Reuses the same root `VAPID_PUBLIC_KEY` variable that the backend consumes — one source of truth for the public key, same pattern as `KEYCLOAK_CLIENT_ID` feeding both `KEYCLOAK_CLIENT_ID` and `VITE_KEYCLOAK_CLIENT_ID`.)

- [ ] **Step 3: Add the build arg to `frontend/Dockerfile`**

```dockerfile
ARG VITE_API_URL=""
ARG VITE_KEYCLOAK_URL
ARG VITE_KEYCLOAK_REALM=foodo
ARG VITE_KEYCLOAK_CLIENT_ID=bread-frontend
ARG VITE_VAPID_PUBLIC_KEY=""

ENV VITE_API_URL=$VITE_API_URL \
    VITE_KEYCLOAK_URL=$VITE_KEYCLOAK_URL \
    VITE_KEYCLOAK_REALM=$VITE_KEYCLOAK_REALM \
    VITE_KEYCLOAK_CLIENT_ID=$VITE_KEYCLOAK_CLIENT_ID \
    VITE_VAPID_PUBLIC_KEY=$VITE_VAPID_PUBLIC_KEY
```

- [ ] **Step 4: Verify**

There's no automated check for compose/Dockerfile syntax in this repo. Manually confirm: `docker compose -f docker-compose.prod.yml config` parses without error (this only validates YAML/interpolation syntax, not that the app runs — it does not require the referenced images to exist).

- [ ] **Step 5: Commit**

```bash
git add .env.example docker-compose.prod.yml frontend/Dockerfile
git commit -m "chore: plumb VAPID/push encryption env vars through docker-compose and .env.example"
```

---

## Task 10: Frontend — service worker

**Files:**
- Create: `frontend/public/sw.js`

**Interfaces:**
- Produces: a static file served at `/sw.js` by Vite (dev) and nginx (prod, via `frontend/nginx.conf`'s `try_files`, no config change needed since it's a real file in the build output root).
- No automated test — per the design spec, there's no realistic way to assert real OS push delivery in Playwright/vitest (requires a real browser push subscription, a secure context, and a push service round-trip). Verified manually in Task 13.

- [ ] **Step 1: Create the service worker**

Create `frontend/public/sw.js`:

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

- [ ] **Step 2: Verify it's served**

Run: `cd frontend && npm run dev` (or your usual dev command), then in a browser, navigate to `http://localhost:5173/sw.js` (adjust port if different) and confirm the raw JS content loads (not a 404 or the SPA's `index.html`).

Stop the dev server when done.

- [ ] **Step 3: Commit**

```bash
git add frontend/public/sw.js
git commit -m "feat: add service worker for Web Push notifications"
```

---

## Task 11: Frontend — `api/push.ts`

**Files:**
- Create: `frontend/src/api/push.ts`

**Interfaces:**
- Consumes: `apiFetch` from `frontend/src/api/client.ts` (existing)
- Produces: `subscribePush(sub: { endpoint: string; keys: { p256dh: string; auth: string } }): Promise<unknown>`, `unsubscribePush(endpoint: string): Promise<unknown>` — consumed by Task 12's `push/subscribe.ts`.
- No dedicated test file — matches the existing convention in this codebase: `frontend/src/api/orders.ts` and `frontend/src/api/products.ts` are thin `apiFetch` wrappers with no unit tests of their own (only `client.ts` itself is tested).

- [ ] **Step 1: Implement the wrapper functions**

Create `frontend/src/api/push.ts`:

```ts
import { apiFetch } from './client'

export const subscribePush = (sub: { endpoint: string; keys: { p256dh: string; auth: string } }) =>
  apiFetch('/api/v1/push/subscribe', { method: 'POST', body: JSON.stringify(sub) })

export const unsubscribePush = (endpoint: string) =>
  apiFetch('/api/v1/push/subscribe', { method: 'DELETE', body: JSON.stringify({ endpoint }) })
```

- [ ] **Step 2: Verify it compiles**

Run: `cd frontend && npx tsc -b --noEmit`
Expected: no new errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api/push.ts
git commit -m "feat: add push subscribe/unsubscribe API client functions"
```

---

## Task 12: Frontend — `push/subscribe.ts`

**Files:**
- Create: `frontend/src/push/subscribe.ts`
- Test: `frontend/src/push/subscribe.test.ts`

**Interfaces:**
- Consumes: `subscribePush` from `frontend/src/api/push.ts` (Task 11)
- Produces: `isPushSupported(): boolean`, `subscribeToPush(): Promise<PushSubscription>`, `getExistingSubscription(): Promise<PushSubscription | null>` — consumed by Task 13's `NotifyMeButton.tsx`.

- [ ] **Step 1: Write the failing tests**

Create `frontend/src/push/subscribe.test.ts`:

```ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { isPushSupported, subscribeToPush } from './subscribe'
import * as pushApi from '../api/push'

vi.mock('../api/push')

beforeEach(() => {
  // subscribeToPush() reads this at call time (not at module load), so
  // stubbing it fresh per test is sufficient — no dynamic re-import needed.
  vi.stubEnv('VITE_VAPID_PUBLIC_KEY', 'BEl62iUYgUivxIkv69yViEuiBIa40HI80NM9-_uk_Vd0mrs4X6qKZfDgO2c5NEnXwuh3AteEBzY-Ov6uJU')
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.unstubAllEnvs()
  vi.clearAllMocks()
})

describe('isPushSupported', () => {
  it('returns false when serviceWorker/PushManager are absent', () => {
    expect(isPushSupported()).toBe(false)
  })

  it('returns true when both are present', () => {
    vi.stubGlobal('PushManager', function () {})
    Object.defineProperty(navigator, 'serviceWorker', { value: {}, configurable: true })
    expect(isPushSupported()).toBe(true)
  })
})

describe('subscribeToPush', () => {
  const fakeSubscription = {
    endpoint: 'https://push.example.com/abc',
    toJSON: () => ({
      endpoint: 'https://push.example.com/abc',
      keys: { p256dh: 'p256dh-key', auth: 'auth-key' },
    }),
  }

  beforeEach(() => {
    vi.stubGlobal('Notification', { requestPermission: vi.fn().mockResolvedValue('granted') })
    Object.defineProperty(navigator, 'serviceWorker', {
      value: {
        register: vi.fn().mockResolvedValue({
          pushManager: { subscribe: vi.fn().mockResolvedValue(fakeSubscription) },
        }),
      },
      configurable: true,
    })
  })

  it('registers the service worker, subscribes, and posts to the backend', async () => {
    await subscribeToPush()
    expect(pushApi.subscribePush).toHaveBeenCalledWith({
      endpoint: 'https://push.example.com/abc',
      keys: { p256dh: 'p256dh-key', auth: 'auth-key' },
    })
  })

  it('throws when permission is denied', async () => {
    vi.stubGlobal('Notification', { requestPermission: vi.fn().mockResolvedValue('denied') })
    await expect(subscribeToPush()).rejects.toThrow('denied')
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npx vitest run src/push/subscribe.test.ts`
Expected: FAIL — `Failed to resolve import "./subscribe"`.

- [ ] **Step 3: Implement `subscribe.ts`**

Create `frontend/src/push/subscribe.ts`:

```ts
import { subscribePush } from '../api/push'

function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = window.atob(base64)
  const outputArray = new Uint8Array(rawData.length)
  for (let i = 0; i < rawData.length; i++) {
    outputArray[i] = rawData.charCodeAt(i)
  }
  return outputArray
}

export function isPushSupported(): boolean {
  return 'serviceWorker' in navigator && 'PushManager' in window
}

export async function getExistingSubscription(): Promise<PushSubscription | null> {
  if (!isPushSupported()) return null
  const registration = await navigator.serviceWorker.getRegistration('/sw.js')
  if (!registration) return null
  return registration.pushManager.getSubscription()
}

export async function subscribeToPush(): Promise<PushSubscription> {
  const permission = await Notification.requestPermission()
  if (permission !== 'granted') {
    throw new Error(`notification permission ${permission}`)
  }
  const registration = await navigator.serviceWorker.register('/sw.js')
  const subscription = await registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(import.meta.env.VITE_VAPID_PUBLIC_KEY as string),
  })
  const json = subscription.toJSON()
  await subscribePush({
    endpoint: json.endpoint!,
    keys: { p256dh: json.keys!.p256dh, auth: json.keys!.auth },
  })
  return subscription
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `npx vitest run src/push/subscribe.test.ts`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/push/subscribe.ts frontend/src/push/subscribe.test.ts
git commit -m "feat: add push subscription orchestration (permission, service worker, backend sync)"
```

---

## Task 13: Frontend — `NotifyMeButton` and `OrderStatus.tsx` integration

**Files:**
- Create: `frontend/src/components/NotifyMeButton.tsx`
- Test: `frontend/src/components/NotifyMeButton.test.tsx`
- Modify: `frontend/src/pages/OrderStatus.tsx`

**Interfaces:**
- Consumes: `isPushSupported`, `subscribeToPush`, `getExistingSubscription` from `frontend/src/push/subscribe.ts` (Task 12)
- Produces: `<NotifyMeButton />` (default export, no props), rendered inside `OrderStatus.tsx`.

- [ ] **Step 1: Write the failing component tests**

Create `frontend/src/components/NotifyMeButton.test.tsx`:

```tsx
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import NotifyMeButton from './NotifyMeButton'
import * as subscribeModule from '../push/subscribe'

vi.mock('../push/subscribe')
vi.mock('sonner', () => ({ toast: { error: vi.fn() } }))

beforeEach(() => {
  vi.clearAllMocks()
  // jsdom does not implement the Notification API at all, so it must be
  // stubbed wholesale rather than patched via Object.defineProperty.
  vi.stubGlobal('Notification', { permission: 'default', requestPermission: vi.fn() })
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('NotifyMeButton', () => {
  it('renders nothing when push is unsupported', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(false)

    const { container } = render(<NotifyMeButton />)

    await waitFor(() => expect(container).not.toHaveTextContent('checking'))
    expect(container).toBeEmptyDOMElement()
  })

  it('shows the opt-in button when supported and not yet subscribed', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(true)
    vi.mocked(subscribeModule.getExistingSubscription).mockResolvedValue(null)

    render(<NotifyMeButton />)

    expect(await screen.findByRole('button', { name: /get notified/i })).toBeInTheDocument()
  })

  it('shows "Notifications on" when already subscribed', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(true)
    vi.mocked(subscribeModule.getExistingSubscription).mockResolvedValue({} as unknown as PushSubscription)

    render(<NotifyMeButton />)

    expect(await screen.findByText(/notifications on/i)).toBeInTheDocument()
  })

  it('subscribes on click and switches to the subscribed state', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(true)
    vi.mocked(subscribeModule.getExistingSubscription).mockResolvedValue(null)
    vi.mocked(subscribeModule.subscribeToPush).mockResolvedValue({} as unknown as PushSubscription)

    render(<NotifyMeButton />)
    const button = await screen.findByRole('button', { name: /get notified/i })
    await userEvent.click(button)

    expect(await screen.findByText(/notifications on/i)).toBeInTheDocument()
  })

  it('shows the blocked hint when permission was previously denied', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(true)
    vi.stubGlobal('Notification', { permission: 'denied', requestPermission: vi.fn() })

    render(<NotifyMeButton />)

    expect(await screen.findByText(/blocked/i)).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npx vitest run src/components/NotifyMeButton.test.tsx`
Expected: FAIL — `Failed to resolve import "./NotifyMeButton"`.

- [ ] **Step 3: Implement the component**

Create `frontend/src/components/NotifyMeButton.tsx`:

```tsx
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { isPushSupported, subscribeToPush, getExistingSubscription } from '../push/subscribe'
import { Button } from './ui/button'

type Status = 'checking' | 'unsupported' | 'blocked' | 'subscribed' | 'available'

export default function NotifyMeButton() {
  const [status, setStatus] = useState<Status>('checking')

  useEffect(() => {
    let cancelled = false
    async function check() {
      if (!isPushSupported()) {
        if (!cancelled) setStatus('unsupported')
        return
      }
      if (Notification.permission === 'denied') {
        if (!cancelled) setStatus('blocked')
        return
      }
      const existing = await getExistingSubscription()
      if (!cancelled) setStatus(existing ? 'subscribed' : 'available')
    }
    check()
    return () => {
      cancelled = true
    }
  }, [])

  async function handleClick() {
    try {
      await subscribeToPush()
      setStatus('subscribed')
    } catch {
      if (Notification.permission === 'denied') {
        setStatus('blocked')
      } else {
        toast.error('Could not enable notifications')
      }
    }
  }

  if (status === 'checking' || status === 'unsupported') return null

  if (status === 'blocked') {
    return <p className="text-xs text-muted-foreground">Notifications blocked — enable in browser settings</p>
  }

  if (status === 'subscribed') {
    return <p className="text-xs text-muted-foreground">Notifications on</p>
  }

  return (
    <Button variant="outline" size="sm" onClick={handleClick}>
      Get notified about this order
    </Button>
  )
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `npx vitest run src/components/NotifyMeButton.test.tsx`
Expected: PASS (5 tests).

- [ ] **Step 5: Integrate into `OrderStatus.tsx`**

In `frontend/src/pages/OrderStatus.tsx`, add the import:

```tsx
import NotifyMeButton from '../components/NotifyMeButton'
```

Render it right after the status badge row:

```tsx
      <div className="flex items-center gap-3 mb-2">
        <h1 className="text-2xl font-bold text-foreground">Order</h1>
        <StatusBadge status={order.status} />
      </div>
      <div className="mb-4">
        <NotifyMeButton />
      </div>
      <p className="text-sm text-muted-foreground mb-6">
        Placed {new Date(order.created_at).toLocaleDateString()}
      </p>
```

- [ ] **Step 6: Run the full frontend test suite**

Run: `npm run test`
Expected: PASS, no regressions (in particular, no existing `OrderStatus`-related test breaks — check `frontend/e2e/*.spec.ts` don't assert exact DOM structure around the status badge that this new element would disrupt; the e2e mocks in `frontend/e2e/fixtures/index.ts` don't intercept `/api/v1/push/*`, so in e2e runs `isPushSupported()` will genuinely reflect the test browser's capabilities — Chromium in Playwright supports Push API, so the button may attempt to call `subscribeToPush()` if a test happens to interact near it; none of the existing e2e specs click near this area, so no changes should be needed there, but re-run `npm run test:e2e` to confirm).

Run: `npm run test:e2e`
Expected: PASS, no regressions.

- [ ] **Step 7: Manual QA (real browser, per the design spec's Testing section)**

This is the only way to verify actual OS-level push delivery — no automated test can do this:

1. Set real `VAPID_PUBLIC_KEY`/`VAPID_PRIVATE_KEY`/`VAPID_SUBJECT`/`PUSH_ENCRYPTION_KEY` in your local backend env (generate per Task 9's instructions).
2. Set `VITE_VAPID_PUBLIC_KEY` to the same public key for the frontend dev server.
3. Run the backend and frontend locally (`localhost` is exempt from the Push API's secure-context requirement, so plain HTTP works for this).
4. Log in as a customer, place an order, open its Order Status page, click "Get notified about this order", accept the browser permission prompt.
5. In another session/tab, log in as the owner, open the kanban board (`/admin/orders`), drag the order's card from "New Orders" to "Accepted".
6. Confirm an OS notification titled "Order accepted" appears — including with the customer's tab in the background or the browser minimized (not fully quit, since this is a dev-only plain-HTTP setup without a real deployed origin for the browser vendor's push service to always reach a fully closed browser — full closed-browser delivery is best verified against the deployed HTTPS site).
7. Repeat for Start (→ "Being prepared") and Finish (→ "Order ready").
8. Click the notification and confirm it navigates to `/orders/<id>`.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/components/NotifyMeButton.tsx frontend/src/components/NotifyMeButton.test.tsx frontend/src/pages/OrderStatus.tsx
git commit -m "feat: add NotifyMeButton opt-in control to Order Status page"
```

---

## Final verification checklist

- [ ] `cd backend && go build ./...` succeeds
- [ ] `cd backend && go vet ./...` is clean (the `internal/infra/http/handler` package's pre-existing `order_test.go` breakage was fixed as a prerequisite before Task 8)
- [ ] `cd backend && go test ./... -v` all pass (excluding testcontainer-based tests if Docker is unavailable in this environment)
- [ ] `cd frontend && npx tsc -b --noEmit` succeeds
- [ ] `cd frontend && npm run test` passes
- [ ] `cd frontend && npm run test:e2e` passes
- [ ] Manual QA from Task 13 Step 7 completed
