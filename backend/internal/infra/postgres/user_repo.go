package postgres

import (
	"context"
	"errors"

	"github.com/fhardow/bread-order/internal/domain/user"
	"github.com/fhardow/bread-order/internal/infra/postgres/models"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) user.Repository {
	return &userRepo{db: db}
}

func (r *userRepo) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	var m models.User
	err := dbFromCtx(ctx, r.db).First(&m, "id = ?", id.String()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domerrors.NotFound("user %s not found", id)
		}
		return nil, err
	}
	return userToDomain(&m)
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var m models.User
	err := dbFromCtx(ctx, r.db).First(&m, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domerrors.NotFound("user with email %s not found", email)
		}
		return nil, err
	}
	return userToDomain(&m)
}

func (r *userRepo) List(ctx context.Context) ([]*user.User, error) {
	var ms []models.User
	if err := dbFromCtx(ctx, r.db).Order("created_at asc").Find(&ms).Error; err != nil {
		return nil, err
	}
	users := make([]*user.User, 0, len(ms))
	for i := range ms {
		u, err := userToDomain(&ms[i])
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *userRepo) Save(ctx context.Context, u *user.User) error {
	m := userToModel(u)
	return dbFromCtx(ctx, r.db).Save(&m).Error
}

func (r *userRepo) Delete(ctx context.Context, id user.ID) error {
	return dbFromCtx(ctx, r.db).Delete(&models.User{}, "id = ?", id.String()).Error
}

func userToDomain(m *models.User) (*user.User, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	return user.Reconstitute(id, m.Name, m.Email, m.Phone, m.CreatedAt, m.UpdatedAt), nil
}

func userToModel(u *user.User) models.User {
	return models.User{
		ID:        u.ID().String(),
		Name:      u.Name(),
		Email:     u.Email(),
		Phone:     u.Phone(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}
