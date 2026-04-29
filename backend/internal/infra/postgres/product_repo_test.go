package postgres_test

import (
	"context"
	"testing"

	repopostgres "github.com/fhardow/foodo/internal/infra/postgres"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fhardow/foodo/internal/domain/product"
)

func makeProduct(t *testing.T, name string) *product.Product {
	t.Helper()
	p, err := product.New(name, "test desc", 450, "loaf", true)
	require.NoError(t, err)
	return p
}

func TestProductRepo_SaveAndFindByID(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewProductRepo(db)
	ctx := context.Background()

	p := makeProduct(t, "Sourdough")
	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByID(ctx, p.ID())
	require.NoError(t, err)

	assert.Equal(t, p.ID(), found.ID())
	assert.Equal(t, "Sourdough", found.Name())
	assert.Equal(t, int64(450), found.PriceCents())
	assert.Equal(t, "loaf", found.Unit())
	assert.True(t, found.Available())
}

func TestProductRepo_FindByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewProductRepo(db)

	_, err := repo.FindByID(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestProductRepo_Save_UpdatesExistingRecord(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewProductRepo(db)
	ctx := context.Background()

	p := makeProduct(t, "Sourdough")
	require.NoError(t, repo.Save(ctx, p))

	require.NoError(t, p.Update("Whole Wheat", "updated", 600, "kg"))
	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByID(ctx, p.ID())
	require.NoError(t, err)
	assert.Equal(t, "Whole Wheat", found.Name())
	assert.Equal(t, int64(600), found.PriceCents())
}

func TestProductRepo_List_All(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewProductRepo(db)
	ctx := context.Background()

	p1 := makeProduct(t, "ProductA")
	p2 := makeProduct(t, "ProductB")
	p2.SetAvailable(false)

	require.NoError(t, repo.Save(ctx, p1))
	require.NoError(t, repo.Save(ctx, p2))

	all, err := repo.List(ctx, false)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestProductRepo_List_AvailableOnly(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewProductRepo(db)
	ctx := context.Background()

	p1 := makeProduct(t, "Available")
	p2 := makeProduct(t, "Unavailable")
	p2.SetAvailable(false)

	require.NoError(t, repo.Save(ctx, p1))
	require.NoError(t, repo.Save(ctx, p2))

	available, err := repo.List(ctx, true)
	require.NoError(t, err)
	require.Len(t, available, 1)
	assert.Equal(t, p1.ID(), available[0].ID())
}

func TestProductRepo_List_Empty(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewProductRepo(db)

	products, err := repo.List(context.Background(), false)
	require.NoError(t, err)
	assert.Empty(t, products)
}

func TestProductRepo_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewProductRepo(db)
	ctx := context.Background()

	p := makeProduct(t, "DeleteMe")
	require.NoError(t, repo.Save(ctx, p))

	require.NoError(t, repo.Delete(ctx, p.ID()))

	_, err := repo.FindByID(ctx, p.ID())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestProductRepo_Delete_NonExistentIDIsNoOp(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewProductRepo(db)

	err := repo.Delete(context.Background(), uuid.New())
	assert.NoError(t, err)
}
