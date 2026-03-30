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

func newValidOrder(t *testing.T) *order.Order {
	t.Helper()
	o, err := order.New(uuid.New())
	require.NoError(t, err)
	return o
}

func newProduct(t *testing.T) (product.ID, string, int64) {
	t.Helper()
	return uuid.New(), "Sourdough", int64(450)
}

func TestOrder_New_Success(t *testing.T) {
	userID := uuid.New()
	o, err := order.New(userID)
	require.NoError(t, err)
	require.NotNil(t, o)
	assert.NotEqual(t, uuid.Nil, o.ID())
	assert.Equal(t, userID, o.UserID())
	assert.Equal(t, order.StatusPending, o.Status())
	assert.Empty(t, o.Items())
	assert.False(t, o.CreatedAt().IsZero())
	assert.False(t, o.UpdatedAt().IsZero())
}

func TestOrder_New_MissingUserID(t *testing.T) {
	o, err := order.New(uuid.Nil)
	assert.Nil(t, o)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_AddItem_Success(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	err := o.AddItem(pid, name, "loaf", 2, price)
	require.NoError(t, err)
	items := o.Items()
	require.Len(t, items, 1)
	assert.Equal(t, pid, items[0].ProductID)
	assert.Equal(t, 2, items[0].Quantity)
}

func TestOrder_AddItem_MergesSameProduct(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, "loaf", 2, price))
	require.NoError(t, o.AddItem(pid, name, "loaf", 3, price))
	items := o.Items()
	require.Len(t, items, 1)
	assert.Equal(t, 5, items[0].Quantity)
}

func TestOrder_AddItem_MultipleDistinctProducts(t *testing.T) {
	o := newValidOrder(t)
	pid1, name1, price1 := newProduct(t)
	pid2, name2, price2 := newProduct(t)
	require.NoError(t, o.AddItem(pid1, name1, "loaf", 1, price1))
	require.NoError(t, o.AddItem(pid2, name2, "loaf", 1, price2))
	assert.Len(t, o.Items(), 2)
}

func TestOrder_AddItem_ZeroQuantity(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	err := o.AddItem(pid, name, "loaf", 0, price)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_AddItem_NegativeQuantity(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	err := o.AddItem(pid, name, "loaf", -1, price)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_AddItem_NonPendingOrder(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, "loaf", 1, price))
	require.NoError(t, o.Confirm())
	err := o.AddItem(uuid.New(), "Another", "loaf", 1, 100)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrForbidden))
}

func TestOrder_RemoveItem_Success(t *testing.T) {
	o := newValidOrder(t)
	pid1, name1, price1 := newProduct(t)
	pid2, name2, price2 := newProduct(t)
	require.NoError(t, o.AddItem(pid1, name1, "loaf", 2, price1))
	require.NoError(t, o.AddItem(pid2, name2, "loaf", 1, price2))
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
	require.NoError(t, o.AddItem(pid, name, "loaf", 1, price))
	require.NoError(t, o.Confirm())
	err := o.RemoveItem(pid)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrForbidden))
}

func TestOrder_Confirm_Success(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, "loaf", 1, price))
	before := o.UpdatedAt()
	time.Sleep(time.Millisecond)
	err := o.Confirm()
	require.NoError(t, err)
	assert.Equal(t, order.StatusCreated, o.Status())
	assert.True(t, o.UpdatedAt().After(before))
}

func TestOrder_Confirm_EmptyOrder(t *testing.T) {
	o := newValidOrder(t)
	err := o.Confirm()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_Confirm_AlreadyConfirmed(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, "loaf", 1, price))
	require.NoError(t, o.Confirm())
	err := o.Confirm()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_Accept_Success(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, "loaf", 1, price))
	require.NoError(t, o.Confirm())
	before := o.UpdatedAt()
	time.Sleep(time.Millisecond)
	err := o.Accept()
	require.NoError(t, err)
	assert.Equal(t, order.StatusAccepted, o.Status())
	assert.True(t, o.UpdatedAt().After(before))
}

func TestOrder_Accept_NotCreated(t *testing.T) {
	o := newValidOrder(t)
	err := o.Accept()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_StartProgress_Success(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, "loaf", 1, price))
	require.NoError(t, o.Confirm())
	require.NoError(t, o.Accept())
	err := o.StartProgress()
	require.NoError(t, err)
	assert.Equal(t, order.StatusOngoing, o.Status())
}

func TestOrder_StartProgress_NotAccepted(t *testing.T) {
	o := newValidOrder(t)
	err := o.StartProgress()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_Finish_Success(t *testing.T) {
	o := newValidOrder(t)
	pid, name, price := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, "loaf", 1, price))
	require.NoError(t, o.Confirm())
	require.NoError(t, o.Accept())
	require.NoError(t, o.StartProgress())
	before := o.UpdatedAt()
	time.Sleep(time.Millisecond)
	err := o.Finish()
	require.NoError(t, err)
	assert.Equal(t, order.StatusFinished, o.Status())
	assert.True(t, o.UpdatedAt().After(before))
}

func TestOrder_Finish_NotOngoing(t *testing.T) {
	o := newValidOrder(t)
	err := o.Finish()
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestOrder_TotalCents_Empty(t *testing.T) {
	o := newValidOrder(t)
	assert.Equal(t, int64(0), o.TotalCents())
}

func TestOrder_TotalCents_SingleItem(t *testing.T) {
	o := newValidOrder(t)
	pid, name, _ := newProduct(t)
	require.NoError(t, o.AddItem(pid, name, "loaf", 3, 200))
	assert.Equal(t, int64(600), o.TotalCents())
}

func TestOrder_TotalCents_MultipleItems(t *testing.T) {
	o := newValidOrder(t)
	pid1, name1, _ := newProduct(t)
	pid2, name2, _ := newProduct(t)
	require.NoError(t, o.AddItem(pid1, name1, "loaf", 2, 300))
	require.NoError(t, o.AddItem(pid2, name2, "loaf", 1, 500))
	assert.Equal(t, int64(1100), o.TotalCents())
}

func TestOrder_TotalCents_AfterRemoveItem(t *testing.T) {
	o := newValidOrder(t)
	pid1, name1, _ := newProduct(t)
	pid2, name2, _ := newProduct(t)
	require.NoError(t, o.AddItem(pid1, name1, "loaf", 2, 300))
	require.NoError(t, o.AddItem(pid2, name2, "loaf", 1, 500))
	require.NoError(t, o.RemoveItem(pid2))
	assert.Equal(t, int64(600), o.TotalCents())
}

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

func TestOrder_Reconstitute_PreservesAllFields(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	pid := uuid.New()
	items := []order.Item{{ProductID: pid, ProductName: "Rye", Unit: "loaf", Quantity: 2, UnitPriceCents: 300}}
	created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	o := order.Reconstitute(id, userID, "", order.StatusCreated, items, created, updated)

	assert.Equal(t, id, o.ID())
	assert.Equal(t, userID, o.UserID())
	assert.Equal(t, order.StatusCreated, o.Status())
	assert.Equal(t, created, o.CreatedAt())
	assert.Equal(t, updated, o.UpdatedAt())
	require.Len(t, o.Items(), 1)
	assert.Equal(t, 2, o.Items()[0].Quantity)
}
