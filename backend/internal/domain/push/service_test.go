package push_test

import (
	"context"
	"testing"

	"github.com/fhardow/foodo/internal/domain/push"
	"github.com/fhardow/foodo/internal/testutil/mock"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Subscribe_Success(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)

	userID := uuid.New()
	sub, err := svc.Subscribe(context.Background(), userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)
	assert.Equal(t, userID, sub.UserID())

	found, err := repo.ListByUser(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "https://push.example.com/abc", found[0].Endpoint())
}

func TestService_Subscribe_ValidationError(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)

	_, err := svc.Subscribe(context.Background(), uuid.Nil, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestService_Subscribe_Resubscribe_ReplacesOldEntry(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)
	userID := uuid.New()

	_, err := svc.Subscribe(context.Background(), userID, "https://push.example.com/abc", "old-p256dh", "old-auth")
	require.NoError(t, err)
	_, err = svc.Subscribe(context.Background(), userID, "https://push.example.com/abc", "new-p256dh", "new-auth")
	require.NoError(t, err)

	found, err := repo.ListByUser(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, found, 1, "resubscribing the same endpoint must not create a duplicate")
	assert.Equal(t, "new-p256dh", found[0].P256dh())
}

func TestService_Unsubscribe_Success(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)
	userID := uuid.New()

	_, err := svc.Subscribe(context.Background(), userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)

	require.NoError(t, svc.Unsubscribe(context.Background(), userID, false, "https://push.example.com/abc"))

	found, err := repo.ListByUser(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestService_Unsubscribe_NotFound(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)

	err := svc.Unsubscribe(context.Background(), uuid.New(), false, "https://push.example.com/never-subscribed")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestService_Unsubscribe_WrongUser_Forbidden(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)
	ownerID := uuid.New()
	attackerID := uuid.New()

	_, err := svc.Subscribe(context.Background(), ownerID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)

	err = svc.Unsubscribe(context.Background(), attackerID, false, "https://push.example.com/abc")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrForbidden))

	// subscription must still exist
	found, err := repo.ListByUser(context.Background(), ownerID)
	require.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestService_Unsubscribe_AdminCanDeleteAnyUsersSubscription(t *testing.T) {
	repo := mock.NewPushRepo()
	svc := push.NewService(repo)
	ownerID := uuid.New()
	adminID := uuid.New()

	_, err := svc.Subscribe(context.Background(), ownerID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)

	require.NoError(t, svc.Unsubscribe(context.Background(), adminID, true, "https://push.example.com/abc"))

	found, err := repo.ListByUser(context.Background(), ownerID)
	require.NoError(t, err)
	assert.Empty(t, found)
}
