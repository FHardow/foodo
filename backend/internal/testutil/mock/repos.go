// Package mock provides hand-rolled in-memory implementations of the domain
// repository interfaces for use in unit tests. Each mock supports error
// injection via exported Err* fields that, when non-nil, are returned instead
// of performing the normal in-memory operation.
package mock

import (
	"context"
	"sync"

	"github.com/fhardow/bread-order/internal/domain/order"
	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/fhardow/bread-order/internal/domain/user"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// UserRepo
// ---------------------------------------------------------------------------

// UserRepo is an in-memory implementation of user.Repository.
type UserRepo struct {
	mu sync.RWMutex

	// Error injection — set before the call you want to fail.
	ErrFindByID    error
	ErrFindByEmail error
	ErrList        error
	ErrSave        error
	ErrDelete      error

	users map[string]*user.User
}

// NewUserRepo returns a ready-to-use UserRepo.
func NewUserRepo() *UserRepo {
	return &UserRepo{users: make(map[string]*user.User)}
}

func (r *UserRepo) FindByID(_ context.Context, id user.ID) (*user.User, error) {
	if r.ErrFindByID != nil {
		return nil, r.ErrFindByID
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id.String()]
	if !ok {
		return nil, domerrors.NotFound("user %s not found", id)
	}
	return u, nil
}

func (r *UserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if r.ErrFindByEmail != nil {
		return nil, r.ErrFindByEmail
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email() == email {
			return u, nil
		}
	}
	return nil, domerrors.NotFound("user with email %s not found", email)
}

func (r *UserRepo) List(_ context.Context) ([]*user.User, error) {
	if r.ErrList != nil {
		return nil, r.ErrList
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*user.User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, u)
	}
	return out, nil
}

func (r *UserRepo) Save(_ context.Context, u *user.User) error {
	if r.ErrSave != nil {
		return r.ErrSave
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID().String()] = u
	return nil
}

func (r *UserRepo) Delete(_ context.Context, id user.ID) error {
	if r.ErrDelete != nil {
		return r.ErrDelete
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id.String())
	return nil
}

// ---------------------------------------------------------------------------
// ProductRepo
// ---------------------------------------------------------------------------

// ProductRepo is an in-memory implementation of product.Repository.
type ProductRepo struct {
	mu sync.RWMutex

	ErrFindByID error
	ErrList     error
	ErrSave     error
	ErrDelete   error

	products map[string]*product.Product
}

// NewProductRepo returns a ready-to-use ProductRepo.
func NewProductRepo() *ProductRepo {
	return &ProductRepo{products: make(map[string]*product.Product)}
}

func (r *ProductRepo) FindByID(_ context.Context, id product.ID) (*product.Product, error) {
	if r.ErrFindByID != nil {
		return nil, r.ErrFindByID
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[id.String()]
	if !ok {
		return nil, domerrors.NotFound("product %s not found", id)
	}
	return p, nil
}

func (r *ProductRepo) List(_ context.Context, availableOnly bool) ([]*product.Product, error) {
	if r.ErrList != nil {
		return nil, r.ErrList
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*product.Product, 0, len(r.products))
	for _, p := range r.products {
		if availableOnly && !p.Available() {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *ProductRepo) Save(_ context.Context, p *product.Product) error {
	if r.ErrSave != nil {
		return r.ErrSave
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.products[p.ID().String()] = p
	return nil
}

func (r *ProductRepo) Delete(_ context.Context, id product.ID) error {
	if r.ErrDelete != nil {
		return r.ErrDelete
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.products, id.String())
	return nil
}

// ---------------------------------------------------------------------------
// OrderRepo
// ---------------------------------------------------------------------------

// OrderRepo is an in-memory implementation of order.Repository.
type OrderRepo struct {
	mu sync.RWMutex

	ErrFindByID   error
	ErrListByUser error
	ErrList       error
	ErrSave       error
	ErrDelete     error

	orders map[string]*order.Order
}

// NewOrderRepo returns a ready-to-use OrderRepo.
func NewOrderRepo() *OrderRepo {
	return &OrderRepo{orders: make(map[string]*order.Order)}
}

func (r *OrderRepo) FindByID(_ context.Context, id order.ID) (*order.Order, error) {
	if r.ErrFindByID != nil {
		return nil, r.ErrFindByID
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.orders[id.String()]
	if !ok {
		return nil, domerrors.NotFound("order %s not found", id)
	}
	return o, nil
}

func (r *OrderRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]*order.Order, error) {
	if r.ErrListByUser != nil {
		return nil, r.ErrListByUser
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*order.Order, 0)
	for _, o := range r.orders {
		if o.UserID() == userID {
			out = append(out, o)
		}
	}
	return out, nil
}

func (r *OrderRepo) List(_ context.Context) ([]*order.Order, error) {
	if r.ErrList != nil {
		return nil, r.ErrList
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*order.Order, 0, len(r.orders))
	for _, o := range r.orders {
		out = append(out, o)
	}
	return out, nil
}

func (r *OrderRepo) Save(_ context.Context, o *order.Order) error {
	if r.ErrSave != nil {
		return r.ErrSave
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[o.ID().String()] = o
	return nil
}

func (r *OrderRepo) Delete(_ context.Context, id order.ID) error {
	if r.ErrDelete != nil {
		return r.ErrDelete
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.orders, id.String())
	return nil
}
