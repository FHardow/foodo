package postgres_test

import (
	"context"
	"testing"

	repopostgres "github.com/fhardow/foodo/internal/infra/postgres"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductRepo_Categories_RoundTrip(t *testing.T) {
	db := newTestDB(t)
	products := repopostgres.NewProductRepo(db)
	categories := repopostgres.NewCategoryRepo(db)
	ctx := context.Background()

	p1 := makeProduct(t, "Rye")
	require.NoError(t, p1.SetCategories([]string{"Bakery", "Grains"}))
	require.NoError(t, products.Save(ctx, p1))

	// Different casing must reuse the existing category rather than create one.
	p2 := makeProduct(t, "Sourdough")
	require.NoError(t, p2.SetCategories([]string{"bakery"}))
	require.NoError(t, products.Save(ctx, p2))

	found, err := products.FindByID(ctx, p1.ID())
	require.NoError(t, err)
	assert.Equal(t, []string{"Bakery", "Grains"}, found.Categories())

	found2, err := products.FindByID(ctx, p2.ID())
	require.NoError(t, err)
	assert.Equal(t, []string{"Bakery"}, found2.Categories(), "existing name casing is kept")

	cats, err := categories.List(ctx)
	require.NoError(t, err)
	require.Len(t, cats, 2)
	assert.Equal(t, "Bakery", cats[0].Name())
	assert.Equal(t, "Grains", cats[1].Name())

	// Replacing categories drops old links.
	require.NoError(t, p1.SetCategories([]string{"Grains"}))
	require.NoError(t, products.Save(ctx, p1))
	list, err := products.List(ctx, false)
	require.NoError(t, err)
	for _, p := range list {
		if p.ID() == p1.ID() {
			assert.Equal(t, []string{"Grains"}, p.Categories())
		}
	}

	// Deleting a category detaches it from products.
	require.NoError(t, categories.Delete(ctx, cats[0].ID()))
	found2, err = products.FindByID(ctx, p2.ID())
	require.NoError(t, err)
	assert.Empty(t, found2.Categories())

	// Deleting a product with categories works.
	require.NoError(t, products.Delete(ctx, p1.ID()))
}

func TestCategoryRepo_Delete_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewCategoryRepo(db)
	err := repo.Delete(context.Background(), uuid.New())
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}
