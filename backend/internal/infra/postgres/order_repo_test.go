package postgres_test

import (
	"context"
	"testing"

	repopostgres "github.com/fhardow/foodo/internal/infra/postgres"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fhardow/foodo/internal/domain/order"
	"github.com/fhardow/foodo/internal/domain/product"
)

func saveOrder(t *testing.T, repo order.Repository, userID uuid.UUID) *order.Order {
	t.Helper()
	ctx := context.Background()
	o, err := order.New(userID)
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, o))
	return o
}

func TestOrderRepo_SaveAndFindByID(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewOrderRepo(db)
	ctx := context.Background()

	userID := uuid.New()
	o := saveOrder(t, repo, userID)

	found, err := repo.FindByID(ctx, o.ID())
	require.NoError(t, err)

	assert.Equal(t, o.ID(), found.ID())
	assert.Equal(t, userID, found.UserID())
	assert.Equal(t, order.StatusPending, found.Status())
	assert.Empty(t, found.Items())
}

func TestOrderRepo_FindByID_ItemsPreloaded(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewOrderRepo(db)
	ctx := context.Background()

	productID := uuid.New()

	o, err := order.New(uuid.New())
	require.NoError(t, err)
	require.NoError(t, o.AddItem(product.ID(productID), "Sourdough", 3, 450))
	require.NoError(t, repo.Save(ctx, o))

	found, err := repo.FindByID(ctx, o.ID())
	require.NoError(t, err)

	items := found.Items()
	require.Len(t, items, 1, "items must be preloaded when fetching an order")
	assert.Equal(t, product.ID(productID), items[0].ProductID)
	assert.Equal(t, "Sourdough", items[0].ProductName)
	assert.Equal(t, 3, items[0].Quantity)
	assert.Equal(t, int64(450), items[0].UnitPriceCents)
}

func TestOrderRepo_FindByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewOrderRepo(db)

	_, err := repo.FindByID(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestOrderRepo_ListByUser(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewOrderRepo(db)

	userA := uuid.New()
	userB := uuid.New()

	saveOrder(t, repo, userA)
	saveOrder(t, repo, userA)
	saveOrder(t, repo, userB)

	orders, err := repo.ListByUser(context.Background(), userA)
	require.NoError(t, err)
	require.Len(t, orders, 2)
	for _, o := range orders {
		assert.Equal(t, userA, o.UserID())
	}
}

func TestOrderRepo_ListByUser_Empty(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewOrderRepo(db)

	orders, err := repo.ListByUser(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, orders)
}

func TestOrderRepo_Save_UpdatesStatusAndItems(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewOrderRepo(db)
	ctx := context.Background()

	productID := uuid.New()
	o, err := order.New(uuid.New())
	require.NoError(t, err)
	require.NoError(t, o.AddItem(product.ID(productID), "Rye", 2, 300))
	require.NoError(t, repo.Save(ctx, o))
	require.NoError(t, o.Confirm())
	require.NoError(t, repo.Save(ctx, o))

	found, err := repo.FindByID(ctx, o.ID())
	require.NoError(t, err)
	assert.Equal(t, order.StatusConfirmed, found.Status())
	require.Len(t, found.Items(), 1)
	assert.Equal(t, 2, found.Items()[0].Quantity)
}

func TestOrderRepo_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewOrderRepo(db)
	ctx := context.Background()

	o := saveOrder(t, repo, uuid.New())

	require.NoError(t, repo.Delete(ctx, o.ID()))

	_, err := repo.FindByID(ctx, o.ID())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestOrderRepo_Delete_CascadesItems(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewOrderRepo(db)
	ctx := context.Background()

	productID := uuid.New()
	o, err := order.New(uuid.New())
	require.NoError(t, err)
	require.NoError(t, o.AddItem(product.ID(productID), "Sourdough", 1, 450))
	require.NoError(t, repo.Save(ctx, o))

	require.NoError(t, repo.Delete(ctx, o.ID()))

	// Verify the order is gone.
	_, err = repo.FindByID(ctx, o.ID())
	require.Error(t, err)

	// Verify orphan items are gone via the cascade constraint. We query the
	// underlying table directly using GORM's raw query.
	var count int64
	db.Raw("SELECT COUNT(*) FROM order_items WHERE order_id = ?", o.ID().String()).Scan(&count)
	assert.Equal(t, int64(0), count, "order_items rows must be deleted via CASCADE")
}

func TestOrderRepo_Delete_NonExistentIDIsNoOp(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewOrderRepo(db)

	err := repo.Delete(context.Background(), uuid.New())
	assert.NoError(t, err)
}
