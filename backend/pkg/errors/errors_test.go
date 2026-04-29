package errors_test

import (
	"testing"

	"github.com/fhardow/foodo/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinels_AreDistinct(t *testing.T) {
	sentinels := []error{
		errors.ErrNotFound,
		errors.ErrConflict,
		errors.ErrForbidden,
		errors.ErrBadRequest,
	}
	for i, a := range sentinels {
		for j, b := range sentinels {
			if i == j {
				continue
			}
			assert.False(t, errors.Is(a, b), "sentinel %v should not match %v", a, b)
		}
	}
}

func TestNotFound_ConstructionAndUnwrap(t *testing.T) {
	err := errors.NotFound("item %d not found", 42)
	require.Error(t, err)
	assert.Equal(t, "item 42 not found", err.Error())
	assert.True(t, errors.Is(err, errors.ErrNotFound))
	assert.False(t, errors.Is(err, errors.ErrConflict))
	assert.False(t, errors.Is(err, errors.ErrBadRequest))
	assert.False(t, errors.Is(err, errors.ErrForbidden))
}

func TestConflict_ConstructionAndUnwrap(t *testing.T) {
	err := errors.Conflict("email %q already taken", "a@b.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "a@b.com")
	assert.True(t, errors.Is(err, errors.ErrConflict))
	assert.False(t, errors.Is(err, errors.ErrNotFound))
}

func TestForbidden_ConstructionAndUnwrap(t *testing.T) {
	err := errors.Forbidden("access denied")
	require.Error(t, err)
	assert.Equal(t, "access denied", err.Error())
	assert.True(t, errors.Is(err, errors.ErrForbidden))
	assert.False(t, errors.Is(err, errors.ErrBadRequest))
}

func TestBadRequest_ConstructionAndUnwrap(t *testing.T) {
	err := errors.BadRequest("field %s is required", "name")
	require.Error(t, err)
	assert.Equal(t, "field name is required", err.Error())
	assert.True(t, errors.Is(err, errors.ErrBadRequest))
	assert.False(t, errors.Is(err, errors.ErrForbidden))
}

func TestDomainError_ImplementsErrorInterface(t *testing.T) {
	err := errors.NotFound("gone")
	// Must satisfy the error interface — this is a compile-time guarantee,
	// but an explicit runtime check documents the contract clearly.
	var _ error = err
	assert.NotEmpty(t, err.Error())
}

func TestIs_StdlibDelegation(t *testing.T) {
	// errors.Is should behave identically to stdlib errors.Is for sentinels.
	wrapped := errors.NotFound("not here")
	assert.True(t, errors.Is(wrapped, errors.ErrNotFound))
}

func TestFormatting_NoArgs(t *testing.T) {
	err := errors.BadRequest("plain message")
	assert.Equal(t, "plain message", err.Error())
}

func TestFormatting_MultipleArgs(t *testing.T) {
	err := errors.NotFound("user %s with role %s not found", "alice", "admin")
	assert.Equal(t, "user alice with role admin not found", err.Error())
}
