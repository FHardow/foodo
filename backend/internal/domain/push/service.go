package push

import (
	"context"

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

func (s *Service) Unsubscribe(ctx context.Context, endpoint string) error {
	return s.repo.DeleteByEndpoint(ctx, endpoint)
}
