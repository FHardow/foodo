package category_test

import (
	"context"
	"testing"

	"github.com/fhardow/foodo/internal/domain/category"
	"github.com/fhardow/foodo/internal/testutil/mock"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_ListAndDelete(t *testing.T) {
	repo := mock.NewCategoryRepo()
	fruit := repo.Seed("Fruit")
	repo.Seed("Bakery")
	svc := category.NewService(repo)
	ctx := context.Background()

	cats, err := svc.List(ctx)
	require.NoError(t, err)
	require.Len(t, cats, 2)
	assert.Equal(t, "Bakery", cats[0].Name())

	require.NoError(t, svc.Delete(ctx, fruit.ID()))
	cats, _ = svc.List(ctx)
	assert.Len(t, cats, 1)
}

func TestService_Delete_NotFound(t *testing.T) {
	svc := category.NewService(mock.NewCategoryRepo())
	err := svc.Delete(context.Background(), uuid.New())
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}
