package order_test

import (
	"testing"
	"time"

	"github.com/fhardow/bread-order/internal/domain/order"
	"github.com/fhardow/bread-order/internal/domain/product"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// futureDate returns a delivery date safely in the future.
func futureDate() time.Time {
	return time.Now().UTC().Add(48 * time.Hour)
}

func newValidOrder(t *testing.T) *order.Order {
	t.Helper()
	userID := uuid.New()
	o, err := order.New(userID, futureDate(), "please ring bell")
	require.NoError(t, err)
	return o
}

func newProduct(t *testing.T) (product.ID, string, int64) {
	t.Helper()
	return uuid.New(), "Sourdough", int64(450)
}

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestOrder_New_Success(t *testing.T) {
	userID := uuid.New()
	delivery := futureDate()
	o, err := order.New(userID, delivery, "ring bell")
	require.NoError(t, err)
	require.NotNil(t, o)

	assert.NotEqual(t, uuid.Nil, o.ID())
	assert.Equal(t, userID, o.UserID())
	assert.Equal(t, order.StatusPending, o.Status())
	assert.Equal(t, "ring bell", o.Notes())
	assert.Equal(t, delivery.Unix(), o.DeliveryDate().Unix())
	assert.Empty(t, o.Items())
	assert.False(t, o.CreatedAt().IsZero())
	assert.False(t, o.UpdatedAt().IsZero())
}

func TestOrder_New_MissingUserID(t *testing.T) {
	o, err := order.New(uuid.Nil, futureDate(), "")
	assert.Nil(t, o)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_New_ZeroDeliveryDate(t *testing.T) {
	o, err := order.New(uuid.New(), time.Time{}, "")
	assert.Nil(t, o)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_New_PastDeliveryDate(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour)
	o, err := order.New(uuid.New(), past, "")
	assert.Nil(t, o)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

// ---------------------------------------------------------------------------
// AddItem
// ---------------------------------------------------------------------------

func TestOrder_AddItem_Success(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)

	err := o.AddItem(pid, name, 2, price)
	require.NoError(t, err)

	items := o.Items()
	require.Len(t, items, 1)
	assert.Equal(t, pid, items[0].ProductID)
	assert.Equal(t, name, items[0].ProductName)
	assert.Equal(t, 2, items[0].Quantity)
	assert.Equal(t, price, items[0].UnitPriceCents)
}

func TestOrder_AddItem_MergesSameProduct(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)

	require.NoError(t, o.AddItem(pid, name, 2, price))
	require.NoError(t, o.AddItem(pid, name, 3, price))

	items := o.Items()
	require.Len(t, items, 1, "duplicate product should be merged into one line")
	assert.Equal(t, 5, items[0].Quantity)
}

func TestOrder_AddItem_MultipleDistinctProducts(t *testing.T) {
	o := newValidOrder(t)
	pid1, name1, price1 := newProduct(t)
	pid2, name2, price2 := newProduct(t)

	require.NoError(t, o.AddItem(pid1, name1, 1, price1))
	require.NoError(t, o.AddItem(pid2, name2, 1, price2))
	assert.Len(t, o.Items(), 2)
}

func TestOrder_AddItem_ZeroQuantity(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	err := o.AddItem(pid, name, 0, price)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_AddItem_NegativeQuantity(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	err := o.AddItem(pid, name, -1, price)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_AddItem_NonPendingOrder(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, 1, price))
	require.NoError(t, o.Confirm())

	err := o.AddItem(uuid.New(), "Another", 1, 100)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrForbidden))
}

// ---------------------------------------------------------------------------
// RemoveItem
// ---------------------------------------------------------------------------

func TestOrder_RemoveItem_Success(t *testing.T) {
	o := newValidOrder(t)
	pid1, name1, price1 := newProduct(t)
	pid2, name2, price2 := newProduct(t)
	require.NoError(t, o.AddItem(pid1, name1, 2, price1))
	require.NoError(t, o.AddItem(pid2, name2, 1, price2))

	err := o.RemoveItem(pid1)
	require.NoError(t, err)

	items := o.Items()
	require.Len(t, items, 1)
	assert.Equal(t, pid2, items[0].ProductID)
}

func TestOrder_RemoveItem_NotFound(t *testing.T) {
	o := newValidOrder(t)
	err := o.RemoveItem(uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestOrder_RemoveItem_NonPendingOrder(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, 1, price))
	require.NoError(t, o.Confirm())

	err := o.RemoveItem(pid)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrForbidden))
}

// ---------------------------------------------------------------------------
// Confirm
// ---------------------------------------------------------------------------

func TestOrder_Confirm_Success(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, 1, price))

	before := o.UpdatedAt()
	time.Sleep(time.Millisecond)

	err := o.Confirm()
	require.NoError(t, err)
	assert.Equal(t, order.StatusConfirmed, o.Status())
	assert.True(t, o.UpdatedAt().After(before))
}

func TestOrder_Confirm_EmptyOrder(t *testing.T) {
	o := newValidOrder(t)
	err := o.Confirm()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
	assert.Equal(t, order.StatusPending, o.Status())
}

func TestOrder_Confirm_AlreadyConfirmed(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, 1, price))
	require.NoError(t, o.Confirm())

	err := o.Confirm()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

// ---------------------------------------------------------------------------
// Fulfill
// ---------------------------------------------------------------------------

func TestOrder_Fulfill_Success(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, 1, price))
	require.NoError(t, o.Confirm())

	before := o.UpdatedAt()
	time.Sleep(time.Millisecond)

	err := o.Fulfill()
	require.NoError(t, err)
	assert.Equal(t, order.StatusFulfilled, o.Status())
	assert.True(t, o.UpdatedAt().After(before))
}

func TestOrder_Fulfill_NotConfirmed(t *testing.T) {
	o := newValidOrder(t)
	err := o.Fulfill()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

// ---------------------------------------------------------------------------
// Cancel
// ---------------------------------------------------------------------------

func TestOrder_Cancel_FromPending(t *testing.T) {
	o := newValidOrder(t)
	err := o.Cancel()
	require.NoError(t, err)
	assert.Equal(t, order.StatusCancelled, o.Status())
}

func TestOrder_Cancel_FromConfirmed(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, 1, price))
	require.NoError(t, o.Confirm())

	err := o.Cancel()
	require.NoError(t, err)
	assert.Equal(t, order.StatusCancelled, o.Status())
}

func TestOrder_Cancel_FromFulfilled(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, 1, price))
	require.NoError(t, o.Confirm())
	require.NoError(t, o.Fulfill())

	err := o.Cancel()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
	assert.Equal(t, order.StatusFulfilled, o.Status())
}

func TestOrder_Cancel_AlreadyCancelled(t *testing.T) {
	o := newValidOrder(t)
	require.NoError(t, o.Cancel())

	err := o.Cancel()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

// ---------------------------------------------------------------------------
// TotalCents
// ---------------------------------------------------------------------------

func TestOrder_TotalCents_Empty(t *testing.T) {
	o := newValidOrder(t)
	assert.Equal(t, int64(0), o.TotalCents())
}

func TestOrder_TotalCents_SingleItem(t *testing.T) {
	o := newValidOrder(t)
	pid, name, _ := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, 3, 200))
	// 3 * 200 = 600
	assert.Equal(t, int64(600), o.TotalCents())
}

func TestOrder_TotalCents_MultipleItems(t *testing.T) {
	o := newValidOrder(t)
	pid1, name1, _ := newProduct(t)
	pid2, name2, _ := newProduct(t)
	require.NoError(t, o.AddItem(pid1, name1, 2, 300)) // 600
	require.NoError(t, o.AddItem(pid2, name2, 1, 500)) // 500
	// total = 1100
	assert.Equal(t, int64(1100), o.TotalCents())
}

func TestOrder_TotalCents_AfterRemoveItem(t *testing.T) {
	o := newValidOrder(t)
	pid1, name1, _ := newProduct(t)
	pid2, name2, _ := newProduct(t)
	require.NoError(t, o.AddItem(pid1, name1, 2, 300)) // 600
	require.NoError(t, o.AddItem(pid2, name2, 1, 500)) // 500
	require.NoError(t, o.RemoveItem(pid2))
	assert.Equal(t, int64(600), o.TotalCents())
}

// ---------------------------------------------------------------------------
// Item.TotalCents
// ---------------------------------------------------------------------------

func TestItem_TotalCents(t *testing.T) {
	tests := []struct {
		quantity int
		price    int64
		want     int64
	}{
		{1, 100, 100},
		{3, 250, 750},
		{0, 500, 0},
	}
	for _, tc := range tests {
		item := order.Item{Quantity: tc.quantity, UnitPriceCents: tc.price}
		assert.Equal(t, tc.want, item.TotalCents())
	}
}

// ---------------------------------------------------------------------------
// Reconstitute
// ---------------------------------------------------------------------------

func TestOrder_Reconstitute_PreservesAllFields(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	pid := uuid.New()
	items := []order.Item{{ProductID: pid, ProductName: "Rye", Quantity: 2, UnitPriceCents: 300}}
	delivery := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	o := order.Reconstitute(id, userID, order.StatusConfirmed, items, "note", delivery, created, updated)

	assert.Equal(t, id, o.ID())
	assert.Equal(t, userID, o.UserID())
	assert.Equal(t, order.StatusConfirmed, o.Status())
	assert.Equal(t, "note", o.Notes())
	assert.Equal(t, delivery, o.DeliveryDate())
	assert.Equal(t, created, o.CreatedAt())
	assert.Equal(t, updated, o.UpdatedAt())
	require.Len(t, o.Items(), 1)
	assert.Equal(t, 2, o.Items()[0].Quantity)
}
