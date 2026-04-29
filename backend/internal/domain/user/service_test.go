package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fhardow/foodo/internal/domain/user"
	"github.com/fhardow/foodo/internal/testutil/mock"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUserService(repo *mock.UserRepo) *user.Service {
	return user.NewService(repo)
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestService_Register_Success(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	id := uuid.New()
	u, err := svc.Register(context.Background(), id, "Alice", "alice@example.com", "+1")
	require.NoError(t, err)
	require.NotNil(t, u)

	assert.Equal(t, id, u.ID())
	assert.Equal(t, "Alice", u.Name())
	assert.Equal(t, "alice@example.com", u.Email())

	// verify persisted
	found, err := repo.FindByID(context.Background(), u.ID())
	require.NoError(t, err)
	assert.Equal(t, u.ID(), found.ID())
}

func TestService_Register_ConflictOnDuplicateEmail(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	_, err := svc.Register(context.Background(), uuid.New(), "Alice", "alice@example.com", "")
	require.NoError(t, err)

	_, err = svc.Register(context.Background(), uuid.New(), "Alicia", "alice@example.com", "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrConflict))
}

func TestService_Register_ValidationError(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	_, err := svc.Register(context.Background(), uuid.New(), "", "alice@example.com", "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestService_Register_RepoSaveError(t *testing.T) {
	repo := mock.NewUserRepo()
	repo.ErrSave = errors.New("db unavailable")
	svc := newUserService(repo)

	_, err := svc.Register(context.Background(), uuid.New(), "Alice", "alice@example.com", "")
	require.Error(t, err)
	assert.Equal(t, "db unavailable", err.Error())
}

func TestService_Register_RepoFindByEmailError(t *testing.T) {
	repo := mock.NewUserRepo()
	repo.ErrFindByEmail = errors.New("db error")
	svc := newUserService(repo)

	_, err := svc.Register(context.Background(), uuid.New(), "Alice", "alice@example.com", "")
	require.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestService_GetByID_Success(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	u, _ := svc.Register(context.Background(), uuid.New(), "Bob", "bob@example.com", "")

	found, err := svc.GetByID(context.Background(), u.ID())
	require.NoError(t, err)
	assert.Equal(t, u.ID(), found.ID())
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	_, err := svc.GetByID(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestService_List_Empty(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	users, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestService_List_MultipleUsers(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	svc.Register(context.Background(), uuid.New(), "Alice", "alice@example.com", "")
	svc.Register(context.Background(), uuid.New(), "Bob", "bob@example.com", "")

	users, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestService_List_RepoError(t *testing.T) {
	repo := mock.NewUserRepo()
	repo.ErrList = errors.New("connection reset")
	svc := newUserService(repo)

	_, err := svc.List(context.Background())
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// UpdateContact
// ---------------------------------------------------------------------------

func TestService_UpdateContact_Success(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	u, _ := svc.Register(context.Background(), uuid.New(), "Alice", "alice@example.com", "111")

	updated, err := svc.UpdateContact(context.Background(), u.ID(), "Alicia", "alicia@example.com", "999")
	require.NoError(t, err)
	assert.Equal(t, "Alicia", updated.Name())
	assert.Equal(t, "alicia@example.com", updated.Email())

	// Check persisted state reflects update.
	fetched, _ := repo.FindByID(context.Background(), u.ID())
	assert.Equal(t, "Alicia", fetched.Name())
}

func TestService_UpdateContact_NotFound(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	_, err := svc.UpdateContact(context.Background(), uuid.New(), "X", "x@example.com", "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrNotFound))
}

func TestService_UpdateContact_ValidationError(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	u, _ := svc.Register(context.Background(), uuid.New(), "Alice", "alice@example.com", "")

	_, err := svc.UpdateContact(context.Background(), u.ID(), "", "alice@example.com", "")
	require.Error(t, err)
	assert.True(t, domerrors.Is(err, domerrors.ErrBadRequest))
}

func TestService_UpdateContact_SaveError(t *testing.T) {
	repo := mock.NewUserRepo()
	svc := newUserService(repo)

	u, _ := svc.Register(context.Background(), uuid.New(), "Alice", "alice@example.com", "")

	repo.ErrSave = errors.New("write failed")
	_, err := svc.UpdateContact(context.Background(), u.ID(), "Alicia", "alicia@example.com", "")
	require.Error(t, err)
	assert.Equal(t, "write failed", err.Error())
}
