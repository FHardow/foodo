package order

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, id ID) (*Order, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Order, error)
	List(ctx context.Context) ([]*Order, error)
	Save(ctx context.Context, o *Order) error
	Delete(ctx context.Context, id ID) error
}
