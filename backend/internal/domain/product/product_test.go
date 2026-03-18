package product_test

import (
	"testing"
	"time"

	"github.com/fhardow/bread-order/internal/domain/product"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_Success(t *testing.T) {
	p, err := product.New("Sourdough", "classic loaf", 450, "loaf")
	require.NoError(t, err)
	require.NotNil(t, p)

	assert.NotEqual(t, uuid.Nil, p.ID())
	assert.Equal(t, "Sourdough", p.Name())
	assert.Equal(t, "classic loaf", p.Description())
	assert.Equal(t, int64(450), p.PriceCents())
	assert.Equal(t, "loaf", p.Unit())
	assert.True(t, p.Available(), "new products should be available by default")
	assert.False(t, p.CreatedAt().IsZero())
	assert.False(t, p.UpdatedAt().IsZero())
}

func TestNew_ZeroPriceIsValid(t *testing.T) {
	p, err := product.New("Free Sample", "", 0, "piece")
	require.NoError(t, err)
	assert.Equal(t, int64(0), p.PriceCents())
}

func TestNew_MissingName(t *testing.T) {
	p, err := product.New("", "desc", 100, "unit")
	assert.Nil(t, p)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_NegativePrice(t *testing.T) {
	p, err := product.New("Bread", "", -1, "loaf")
	assert.Nil(t, p)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_MissingUnit(t *testing.T) {
	p, err := product.New("Bread", "", 100, "")
	assert.Nil(t, p)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdate_Success(t *testing.T) {
	p, _ := product.New("Sourdough", "old desc", 450, "loaf")

	before := p.UpdatedAt()
	time.Sleep(time.Millisecond)

	err := p.Update("Whole Wheat", "new desc", 500, "kg")
	require.NoError(t, err)

	assert.Equal(t, "Whole Wheat", p.Name())
	assert.Equal(t, "new desc", p.Description())
	assert.Equal(t, int64(500), p.PriceCents())
	assert.Equal(t, "kg", p.Unit())
	assert.True(t, p.UpdatedAt().After(before))
}

func TestUpdate_MissingName(t *testing.T) {
	p, _ := product.New("Sourdough", "", 450, "loaf")
	err := p.Update("", "", 450, "loaf")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
	assert.Equal(t, "Sourdough", p.Name(), "name must remain unchanged on error")
}

func TestUpdate_NegativePrice(t *testing.T) {
	p, _ := product.New("Sourdough", "", 450, "loaf")
	err := p.Update("Sourdough", "", -50, "loaf")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
	assert.Equal(t, int64(450), p.PriceCents())
}

func TestUpdate_MissingUnit(t *testing.T) {
	p, _ := product.New("Sourdough", "", 450, "loaf")
	err := p.Update("Sourdough", "", 450, "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
	assert.Equal(t, "loaf", p.Unit())
}

// ---------------------------------------------------------------------------
// SetAvailable
// ---------------------------------------------------------------------------

func TestSetAvailable_ToFalse(t *testing.T) {
	p, _ := product.New("Sourdough", "", 450, "loaf")
	assert.True(t, p.Available())

	before := p.UpdatedAt()
	time.Sleep(time.Millisecond)

	p.SetAvailable(false)
	assert.False(t, p.Available())
	assert.True(t, p.UpdatedAt().After(before))
}

func TestSetAvailable_ToTrue(t *testing.T) {
	p, _ := product.New("Sourdough", "", 450, "loaf")
	p.SetAvailable(false)
	p.SetAvailable(true)
	assert.True(t, p.Available())
}

func TestSetAvailable_Idempotent(t *testing.T) {
	p, _ := product.New("Sourdough", "", 450, "loaf")
	p.SetAvailable(true) // already true
	assert.True(t, p.Available())
}

// ---------------------------------------------------------------------------
// Reconstitute
// ---------------------------------------------------------------------------

func TestReconstitute_PreservesAllFields(t *testing.T) {
	id := uuid.New()
	created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	p := product.Reconstitute(id, "Rye", "dark", 300, "loaf", false, created, updated)

	assert.Equal(t, id, p.ID())
	assert.Equal(t, "Rye", p.Name())
	assert.Equal(t, "dark", p.Description())
	assert.Equal(t, int64(300), p.PriceCents())
	assert.Equal(t, "loaf", p.Unit())
	assert.False(t, p.Available())
	assert.Equal(t, created, p.CreatedAt())
	assert.Equal(t, updated, p.UpdatedAt())
}
