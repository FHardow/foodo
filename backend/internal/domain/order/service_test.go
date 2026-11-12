package order_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/product"
	"github.com/fhardow/foodo/internal/domain/user"
	"github.com/fhardow/foodo/internal/testutil/mock"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newOrderService(orderRepo *mock.OrderRepo, productRepo *mock.ProductRepo, userRepo *mock.UserRepo) *order.Service {
	return order.NewService(orderRepo, productRepo, userRepo)
}

// seededUser inserts a user into the mock repo and returns their ID.
func seededUser(t *testing.T, repo *mock.UserRepo) uuid.UUID {
	t.Helper()
	id := uuid.New()
	u, err := user.New(id, "Test User", "test@example.com", "")
	require.NoError(t, err)
	require.NoError(t, repo.Save(context.Background(), u))
	return id
}

// seededProduct inserts a product directly into the mock repo and returns its ID.
func seededProduct(t *testing.T, repo *mock.ProductRepo, available bool) product.ID {
	t.Helper()
	p, err := product.New("Sourdough", "test", 450, "loaf", true)
	require.NoError(t, err)
	if !available {
		p.SetAvailable(false)
	}
	require.NoError(t, repo.Save(context.Background(), p))
	return p.ID()
}

func createOrder(t *testing.T, svc *order.Service, uRepo *mock.UserRepo) *order.Order {
	t.Helper()
	userID := seededUser(t, uRepo)
	o, err := svc.Create(context.Background(), userID)
	require.NoError(t, err)
	return o
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestOrderService_Create_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	userID := seededUser(t, uRepo)

	o, err := svc.Create(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, o)
	assert.Equal(t, order.StatusPending, o.Status())
	assert.Equal(t, userID, o.UserID())

	// verify persisted
	found, err := oRepo.FindByID(context.Background(), o.ID())
	require.NoError(t, err)
	assert.Equal(t, o.ID(), found.ID())
}

func TestOrderService_Create_ValidationError(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	_, err := svc.Create(context.Background(), uuid.Nil)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrderService_Create_UserNotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	_, err := svc.Create(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrderService_Create_SaveError(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	oRepo.ErrSave = errors.New("db write failed")
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	userID := seededUser(t, uRepo)
	_, err := svc.Create(context.Background(), userID)
	require.Error(t, err)
	assert.Equal(t, "db write failed", err.Error())
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestOrderService_GetByID_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)

	found, err := svc.GetByID(context.Background(), o.ID())
	require.NoError(t, err)
	assert.Equal(t, o.ID(), found.ID())
}

func TestOrderService_GetByID_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	_, err := svc.GetByID(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

// ---------------------------------------------------------------------------
// ListByUser / List
// ---------------------------------------------------------------------------

func TestOrderService_ListByUser_FiltersCorrectly(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	userA := seededUser(t, uRepo)
	userB := seededUser(t, uRepo)

	_, err := svc.Create(context.Background(), userA)
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), userA)
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), userB)
	require.NoError(t, err)

	orders, err := svc.ListByUser(context.Background(), userA)
	require.NoError(t, err)
	assert.Len(t, orders, 2)
	for _, o := range orders {
		assert.Equal(t, userA, o.UserID())
	}
}

func TestOrderService_List_ReturnsAll(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	createOrder(t, svc, uRepo)
	createOrder(t, svc, uRepo)

	all, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

// ---------------------------------------------------------------------------
// AddItem
// ---------------------------------------------------------------------------

func TestOrderService_AddItem_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)
	pid := seededProduct(t, pRepo, true)

	updated, err := svc.AddItem(context.Background(), o.ID(), pid, 2)
	require.NoError(t, err)
	require.Len(t, updated.Items(), 1)
	assert.Equal(t, 2, updated.Items()[0].Quantity)
}

func TestOrderService_AddItem_UnavailableProduct(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)
	pid := seededProduct(t, pRepo, false)

	_, err := svc.AddItem(context.Background(), o.ID(), pid, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, product.ErrUnavailable)
}

func TestOrderService_AddItem_OrderNotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	pid := seededProduct(t, pRepo, true)
	_, err := svc.AddItem(context.Background(), uuid.New(), pid, 1)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestOrderService_AddItem_ProductNotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)
	_, err := svc.AddItem(context.Background(), o.ID(), uuid.New(), 1)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

// ---------------------------------------------------------------------------
// RemoveItem
// ---------------------------------------------------------------------------

func TestOrderService_RemoveItem_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)
	pid := seededProduct(t, pRepo, true)
	_, err := svc.AddItem(context.Background(), o.ID(), pid, 1)
	require.NoError(t, err)

	updated, err := svc.RemoveItem(context.Background(), o.ID(), pid)
	require.NoError(t, err)
	assert.Empty(t, updated.Items())
}

func TestOrderService_RemoveItem_ItemNotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)
	_, err := svc.RemoveItem(context.Background(), o.ID(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestOrderService_RemoveItem_OrderNotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	_, err := svc.RemoveItem(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Confirm
// ---------------------------------------------------------------------------

func TestOrderService_Confirm_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)
	pid := seededProduct(t, pRepo, true)
	_, err := svc.AddItem(context.Background(), o.ID(), pid, 1)
	require.NoError(t, err)

	confirmed, err := svc.Confirm(context.Background(), o.ID())
	require.NoError(t, err)
	assert.Equal(t, order.StatusCreated, confirmed.Status())
}

func TestOrderService_Confirm_EmptyOrder(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)
	_, err := svc.Confirm(context.Background(), o.ID())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrderService_Confirm_NotFound(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	_, err := svc.Confirm(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Finish (Accept → StartProgress → Finish)
// ---------------------------------------------------------------------------

func TestOrderService_Finish_Success(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)
	pid := seededProduct(t, pRepo, true)
	svc.AddItem(context.Background(), o.ID(), pid, 1)
	svc.Confirm(context.Background(), o.ID())
	svc.Accept(context.Background(), o.ID())
	svc.StartProgress(context.Background(), o.ID())

	finished, err := svc.Finish(context.Background(), o.ID())
	require.NoError(t, err)
	assert.Equal(t, order.StatusFinished, finished.Status())
}

func TestOrderService_Finish_NotOngoing(t *testing.T) {
	oRepo := mock.NewOrderRepo()
	pRepo := mock.NewProductRepo()
	uRepo := mock.NewUserRepo()
	svc := newOrderService(oRepo, pRepo, uRepo)

	o := createOrder(t, svc, uRepo)
	_, err := svc.Finish(context.Background(), o.ID())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

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
