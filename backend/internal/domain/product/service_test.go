package product_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fhardow/foodo/internal/domain/product"
	"github.com/fhardow/foodo/internal/testutil/mock"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProductService(repo *mock.ProductRepo) *product.Service {
	return product.NewService(repo)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestService_Create_Success(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	p, err := svc.Create(context.Background(), "Sourdough", "classic", 450, "loaf", true)
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, "Sourdough", p.Name())
	assert.True(t, p.Available())

	// verify persisted
	found, err := repo.FindByID(context.Background(), p.ID())
	require.NoError(t, err)
	assert.Equal(t, p.ID(), found.ID())
}

func TestService_Create_ValidationError(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	_, err := svc.Create(context.Background(), "", "desc", 100, "loaf", true)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestService_Create_SaveError(t *testing.T) {
	repo := mock.NewProductRepo()
	repo.ErrSave = errors.New("disk full")
	svc := newProductService(repo)

	_, err := svc.Create(context.Background(), "Bread", "", 100, "loaf", true)
	require.Error(t, err)
	assert.Equal(t, "disk full", err.Error())
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestService_GetByID_Success(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	p, _ := svc.Create(context.Background(), "Rye", "", 300, "loaf", true)

	found, err := svc.GetByID(context.Background(), p.ID())
	require.NoError(t, err)
	assert.Equal(t, p.ID(), found.ID())
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	_, err := svc.GetByID(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestService_List_AllProducts(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	svc.Create(context.Background(), "A", "", 100, "loaf", true)
	svc.Create(context.Background(), "B", "", 200, "kg", true)

	products, err := svc.List(context.Background(), false)
	require.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestService_List_AvailableOnly(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	p1, _ := svc.Create(context.Background(), "Available", "", 100, "loaf", true)
	p2, _ := svc.Create(context.Background(), "Unavailable", "", 200, "loaf", true)
	svc.SetAvailable(context.Background(), p2.ID(), false)

	products, err := svc.List(context.Background(), true)
	require.NoError(t, err)
	require.Len(t, products, 1)
	assert.Equal(t, p1.ID(), products[0].ID())
}

func TestService_List_RepoError(t *testing.T) {
	repo := mock.NewProductRepo()
	repo.ErrList = errors.New("timeout")
	svc := newProductService(repo)

	_, err := svc.List(context.Background(), false)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestService_Update_Success(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	p, _ := svc.Create(context.Background(), "Sourdough", "old", 450, "loaf", true)

	updated, err := svc.Update(context.Background(), p.ID(), "Whole Wheat", "new", 500, "kg")
	require.NoError(t, err)
	assert.Equal(t, "Whole Wheat", updated.Name())
	assert.Equal(t, int64(500), updated.PriceCents())

	// verify persistence
	fetched, _ := repo.FindByID(context.Background(), p.ID())
	assert.Equal(t, "Whole Wheat", fetched.Name())
}

func TestService_Update_NotFound(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	_, err := svc.Update(context.Background(), uuid.New(), "X", "", 100, "loaf")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestService_Update_ValidationError(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	p, _ := svc.Create(context.Background(), "Sourdough", "", 450, "loaf", true)

	_, err := svc.Update(context.Background(), p.ID(), "", "", -1, "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

// ---------------------------------------------------------------------------
// SetAvailable
// ---------------------------------------------------------------------------

func TestService_SetAvailable_ToFalse(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	p, _ := svc.Create(context.Background(), "Sourdough", "", 450, "loaf", true)

	updated, err := svc.SetAvailable(context.Background(), p.ID(), false)
	require.NoError(t, err)
	assert.False(t, updated.Available())
}

func TestService_SetAvailable_ToTrue(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	p, _ := svc.Create(context.Background(), "Sourdough", "", 450, "loaf", true)
	svc.SetAvailable(context.Background(), p.ID(), false)

	updated, err := svc.SetAvailable(context.Background(), p.ID(), true)
	require.NoError(t, err)
	assert.True(t, updated.Available())
}

func TestService_SetAvailable_NotFound(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	_, err := svc.SetAvailable(context.Background(), uuid.New(), false)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestService_Delete_Success(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)

	p, _ := svc.Create(context.Background(), "Sourdough", "", 450, "loaf", true)

	err := svc.Delete(context.Background(), p.ID())
	require.NoError(t, err)

	_, err = svc.GetByID(context.Background(), p.ID())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestService_Delete_RepoError(t *testing.T) {
	repo := mock.NewProductRepo()
	repo.ErrDelete = errors.New("constraint violation")
	svc := newProductService(repo)

	err := svc.Delete(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, "constraint violation", err.Error())
}

// ---------------------------------------------------------------------------
// SetImageURL
// ---------------------------------------------------------------------------

func TestService_SetImageURL(t *testing.T) {
	repo := mock.NewProductRepo()
	svc := newProductService(repo)
	ctx := context.Background()

	p, err := svc.Create(ctx, "Rye", "", 300, "loaf", true)
	require.NoError(t, err)

	updated, err := svc.SetImageURL(ctx, p.ID(), "/uploads/"+p.ID().String()+".jpg")
	require.NoError(t, err)
	assert.Equal(t, "/uploads/"+p.ID().String()+".jpg", updated.ImageURL())
}
