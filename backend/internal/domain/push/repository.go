package push

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, s *Subscription) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Subscription, error)
	FindByEndpoint(ctx context.Context, endpoint string) (*Subscription, error)
	DeleteByEndpoint(ctx context.Context, endpoint string) error
}
