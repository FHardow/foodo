package user_test

import (
	"testing"
	"time"

	"github.com/fhardow/bread-order/internal/domain/user"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_Success(t *testing.T) {
	u, err := user.New("Alice", "alice@example.com", "+1234")
	require.NoError(t, err)
	require.NotNil(t, u)

	assert.NotEqual(t, uuid.Nil, u.ID())
	assert.Equal(t, "Alice", u.Name())
	assert.Equal(t, "alice@example.com", u.Email())
	assert.Equal(t, "+1234", u.Phone())
	assert.False(t, u.CreatedAt().IsZero())
	assert.False(t, u.UpdatedAt().IsZero())
}

func TestNew_PhoneOptional(t *testing.T) {
	u, err := user.New("Bob", "bob@example.com", "")
	require.NoError(t, err)
	assert.Equal(t, "", u.Phone())
}

func TestNew_MissingName(t *testing.T) {
	u, err := user.New("", "alice@example.com", "")
	assert.Nil(t, u)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_MissingEmail(t *testing.T) {
	u, err := user.New("Alice", "", "")
	assert.Nil(t, u)
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_UniqueIDs(t *testing.T) {
	u1, _ := user.New("A", "a@example.com", "")
	u2, _ := user.New("B", "b@example.com", "")
	assert.NotEqual(t, u1.ID(), u2.ID())
}

// ---------------------------------------------------------------------------
// UpdateContact
// ---------------------------------------------------------------------------

func TestUpdateContact_Success(t *testing.T) {
	u, _ := user.New("Alice", "alice@example.com", "111")

	before := u.UpdatedAt()
	time.Sleep(time.Millisecond) // ensure clock advances

	err := u.UpdateContact("Alicia", "alicia@example.com", "999")
	require.NoError(t, err)

	assert.Equal(t, "Alicia", u.Name())
	assert.Equal(t, "alicia@example.com", u.Email())
	assert.Equal(t, "999", u.Phone())
	assert.True(t, u.UpdatedAt().After(before), "UpdatedAt should be bumped")
}

func TestUpdateContact_ClearPhone(t *testing.T) {
	u, _ := user.New("Alice", "alice@example.com", "111")
	err := u.UpdateContact("Alice", "alice@example.com", "")
	require.NoError(t, err)
	assert.Equal(t, "", u.Phone())
}

func TestUpdateContact_MissingName(t *testing.T) {
	u, _ := user.New("Alice", "alice@example.com", "")
	err := u.UpdateContact("", "alice@example.com", "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
	// original name must be unchanged
	assert.Equal(t, "Alice", u.Name())
}

func TestUpdateContact_MissingEmail(t *testing.T) {
	u, _ := user.New("Alice", "alice@example.com", "")
	err := u.UpdateContact("Alice", "", "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
	assert.Equal(t, "alice@example.com", u.Email())
}

// ---------------------------------------------------------------------------
// Reconstitute
// ---------------------------------------------------------------------------

func TestReconstitute_PreservesAllFields(t *testing.T) {
	id := uuid.New()
	created := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updated := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	u := user.Reconstitute(id, "Bob", "bob@example.com", "+9999", user.RoleCustomer, created, updated)

	assert.Equal(t, id, u.ID())
	assert.Equal(t, "Bob", u.Name())
	assert.Equal(t, "bob@example.com", u.Email())
	assert.Equal(t, "+9999", u.Phone())
	assert.Equal(t, created, u.CreatedAt())
	assert.Equal(t, updated, u.UpdatedAt())
}

func TestReconstitute_DoesNotValidate(t *testing.T) {
	// Reconstitute must not run validation — it trusts persisted data.
	id := uuid.New()
	u := user.Reconstitute(id, "", "", "", user.RoleCustomer, time.Now(), time.Now())
	assert.NotNil(t, u, "Reconstitute should succeed even with empty fields")
}
