package postgres_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/fhardow/foodo/internal/domain/push"
	repopostgres "github.com/fhardow/foodo/internal/infra/postgres"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func testEncryptionKey() string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("k"), 32))
}

func newTestPushRepo(t *testing.T, db *gorm.DB) push.Repository {
	t.Helper()
	repo, err := repopostgres.NewPushSubscriptionRepo(db, testEncryptionKey())
	require.NoError(t, err)
	return repo
}

func TestPushSubscriptionRepo_SaveAndListByUser(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)
	ctx := context.Background()

	userID := uuid.New()
	sub, err := push.New(userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub))

	found, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "https://push.example.com/abc", found[0].Endpoint())
	assert.Equal(t, "p256dh-key", found[0].P256dh())
	assert.Equal(t, "auth-key", found[0].Auth())
}

func TestPushSubscriptionRepo_ListByUser_Empty(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)

	found, err := repo.ListByUser(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestPushSubscriptionRepo_Save_UpsertsOnEndpoint(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)
	ctx := context.Background()

	userID := uuid.New()
	sub1, err := push.New(userID, "https://push.example.com/abc", "old-p256dh", "old-auth")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub1))

	sub2, err := push.New(userID, "https://push.example.com/abc", "new-p256dh", "new-auth")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub2))

	found, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, found, 1, "re-subscribing the same endpoint must upsert, not duplicate")
	assert.Equal(t, "new-p256dh", found[0].P256dh())
}

func TestPushSubscriptionRepo_FindByEndpoint_Success(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)
	ctx := context.Background()

	userID := uuid.New()
	sub, err := push.New(userID, "https://push.example.com/find-me", "p256dh-key", "auth-key")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub))

	found, err := repo.FindByEndpoint(ctx, "https://push.example.com/find-me")
	require.NoError(t, err)
	assert.Equal(t, userID, found.UserID())
	assert.Equal(t, "p256dh-key", found.P256dh())
}

func TestPushSubscriptionRepo_FindByEndpoint_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)

	_, err := repo.FindByEndpoint(context.Background(), "https://push.example.com/does-not-exist")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestPushSubscriptionRepo_DeleteByEndpoint(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)
	ctx := context.Background()

	userID := uuid.New()
	sub, err := push.New(userID, "https://push.example.com/abc", "p256dh-key", "auth-key")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub))

	require.NoError(t, repo.DeleteByEndpoint(ctx, "https://push.example.com/abc"))

	found, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestPushSubscriptionRepo_DeleteByEndpoint_NonExistentIsNoOp(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)

	err := repo.DeleteByEndpoint(context.Background(), "https://push.example.com/does-not-exist")
	assert.NoError(t, err)
}

func TestPushSubscriptionRepo_Save_EncryptsAtRest(t *testing.T) {
	db := newTestDB(t)
	repo := newTestPushRepo(t, db)
	ctx := context.Background()

	userID := uuid.New()
	sub, err := push.New(userID, "https://push.example.com/plaintext-check", "p256dh-key", "auth-key")
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, sub))

	var raw struct {
		Endpoint string
		P256dh   string
		Auth     string
	}
	require.NoError(t, db.Raw(
		`SELECT endpoint AS endpoint, p256dh AS p256dh, auth AS auth FROM push_subscriptions WHERE user_id = ?`,
		userID.String(),
	).Scan(&raw).Error)

	assert.NotEqual(t, "https://push.example.com/plaintext-check", raw.Endpoint, "endpoint must be encrypted at rest")
	assert.NotEqual(t, "p256dh-key", raw.P256dh, "p256dh must be encrypted at rest")
	assert.NotEqual(t, "auth-key", raw.Auth, "auth must be encrypted at rest")
}
