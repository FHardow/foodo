package user

import (
	"context"

	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, id uuid.UUID, name, email, phone string) (*User, error) {
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !domerrors.Is(err, domerrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domerrors.Conflict("user with email %q already exists", email)
	}

	u, err := New(id, name, email, phone)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Upsert creates the user if they don't exist yet, or updates their name/email
// from the identity provider token claims if they do. Phone is preserved on update.
func (s *Service) Upsert(ctx context.Context, id uuid.UUID, name, email string) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil && !domerrors.Is(err, domerrors.ErrNotFound) {
		return err
	}
	if existing == nil {
		u, err := New(id, name, email, "")
		if err != nil {
			return err
		}
		return s.repo.Save(ctx, u)
	}
	if err := existing.UpdateContact(name, email, existing.Phone()); err != nil {
		return err
	}
	return s.repo.Save(ctx, existing)
}

func (s *Service) GetByID(ctx context.Context, id ID) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*User, error) {
	return s.repo.List(ctx)
}

func (s *Service) UpdateContact(ctx context.Context, id ID, name, email, phone string) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.UpdateContact(name, email, phone); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
