package category

import "context"

type Repository interface {
	List(ctx context.Context) ([]*Category, error)
	// Delete removes the category and detaches it from all products.
	// Returns a NotFound error if the category does not exist.
	Delete(ctx context.Context, id ID) error
}
