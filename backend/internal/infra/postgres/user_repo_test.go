package postgres_test

import (
	"context"
	"testing"

	repopostgres "github.com/fhardow/bread-order/internal/infra/postgres"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fhardow/bread-order/internal/domain/user"
)

func TestUserRepo_SaveAndFindByID(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewUserRepo(db)
	ctx := context.Background()

	u, err := user.New(uuid.New(), "Alice", "alice@example.com", "+1234")
	require.NoError(t, err)

	require.NoError(t, repo.Save(ctx, u))

	found, err := repo.FindByID(ctx, u.ID())
	require.NoError(t, err)

	assert.Equal(t, u.ID(), found.ID())
	assert.Equal(t, "Alice", found.Name())
	assert.Equal(t, "alice@example.com", found.Email())
	assert.Equal(t, "+1234", found.Phone())
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewUserRepo(db)

	_, err := repo.FindByID(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestUserRepo_FindByEmail(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewUserRepo(db)
	ctx := context.Background()

	u, _ := user.New(uuid.New(), "Bob", "bob@example.com", "")
	require.NoError(t, repo.Save(ctx, u))

	found, err := repo.FindByEmail(ctx, "bob@example.com")
	require.NoError(t, err)
	assert.Equal(t, u.ID(), found.ID())
}

func TestUserRepo_FindByEmail_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewUserRepo(db)

	_, err := repo.FindByEmail(context.Background(), "nobody@example.com")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestUserRepo_List(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewUserRepo(db)
	ctx := context.Background()

	u1, _ := user.New(uuid.New(), "Alice", "alice2@example.com", "")
	u2, _ := user.New(uuid.New(), "Bob", "bob2@example.com", "")
	require.NoError(t, repo.Save(ctx, u1))
	require.NoError(t, repo.Save(ctx, u2))

	users, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepo_List_Empty(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewUserRepo(db)

	users, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestUserRepo_Save_UpdatesExistingRecord(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewUserRepo(db)
	ctx := context.Background()

	u, _ := user.New(uuid.New(), "Alice", "alice3@example.com", "")
	require.NoError(t, repo.Save(ctx, u))

	require.NoError(t, u.UpdateContact("Alicia", "alicia@example.com", "999"))
	require.NoError(t, repo.Save(ctx, u))

	found, err := repo.FindByID(ctx, u.ID())
	require.NoError(t, err)
	assert.Equal(t, "Alicia", found.Name())
	assert.Equal(t, "alicia@example.com", found.Email())
}

func TestUserRepo_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewUserRepo(db)
	ctx := context.Background()

	u, _ := user.New(uuid.New(), "DeleteMe", "deleteme@example.com", "")
	require.NoError(t, repo.Save(ctx, u))

	require.NoError(t, repo.Delete(ctx, u.ID()))

	_, err := repo.FindByID(ctx, u.ID())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestUserRepo_Delete_NonExistentIDIsNoOp(t *testing.T) {
	db := newTestDB(t)
	repo := repopostgres.NewUserRepo(db)

	// GORM Delete does not error on missing rows — it's a soft "0 rows affected".
	err := repo.Delete(context.Background(), uuid.New())
	assert.NoError(t, err)
}
