package push

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

func (s *Service) Subscribe(ctx context.Context, userID uuid.UUID, endpoint, p256dh, auth string) (*Subscription, error) {
	sub, err := New(userID, endpoint, p256dh, auth)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *Service) Unsubscribe(ctx context.Context, callerUserID uuid.UUID, isOwner bool, endpoint string) error {
	sub, err := s.repo.FindByEndpoint(ctx, endpoint)
	if err != nil {
		return err
	}
	if sub.UserID() != callerUserID && !isOwner {
		return domerrors.Forbidden("cannot unsubscribe another user's push subscription")
	}
	return s.repo.DeleteByEndpoint(ctx, endpoint)
}
