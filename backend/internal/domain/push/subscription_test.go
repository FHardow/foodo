package push_test

import (
	"testing"

	"github.com/fhardow/foodo/internal/domain/push"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Success(t *testing.T) {
	userID := uuid.New()
	sub, err := push.New(userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)
	assert.Equal(t, userID, sub.UserID())
	assert.Equal(t, "https://push.example.com/abc", sub.Endpoint())
	assert.Equal(t, "p256dh-key", sub.P256dh())
	assert.Equal(t, "auth-key", sub.Auth())
	assert.NotEqual(t, uuid.Nil, sub.ID())
}

func TestNew_RequiresUserID(t *testing.T) {
	_, err := push.New(uuid.Nil, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_RequiresEndpoint(t *testing.T) {
	_, err := push.New(uuid.New(), "", "p256dh-key", "auth-key")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_RequiresP256dh(t *testing.T) {
	_, err := push.New(uuid.New(), "https://push.example.com/abc", "", "auth-key")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestNew_RequiresAuth(t *testing.T) {
	_, err := push.New(uuid.New(), "https://push.example.com/abc", "p256dh-key", "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}
