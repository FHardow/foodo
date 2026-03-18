package product

import "context"

type Repository interface {
	FindByID(ctx context.Context, id ID) (*Product, error)
	List(ctx context.Context, availableOnly bool) ([]*Product, error)
	Save(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id ID) error
}
