package order

import (
	"context"

	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/google/uuid"
)

type Service struct {
	repo        Repository
	productRepo product.Repository
	notifier    Notifier
}

func NewService(repo Repository, productRepo product.Repository, notifier Notifier) *Service {
	return &Service{repo: repo, productRepo: productRepo, notifier: notifier}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID) (*Order, error) {
	o, err := New(userID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	if s.notifier != nil {
		// Fire-and-forget: notification failure must not break order creation.
		go func() {
			if err := s.notifier.OrderCreated(context.Background(), o); err != nil {
				// Errors are intentionally ignored here; add logging if desired.
				_ = err
			}
		}()
	}
	return o, nil
}

func (s *Service) GetByID(ctx context.Context, id ID) (*Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID) ([]*Order, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) List(ctx context.Context) ([]*Order, error) {
	return s.repo.List(ctx)
}

func (s *Service) AddItem(ctx context.Context, orderID ID, productID product.ID, quantity int) (*Order, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	p, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if !p.Available() {
		return nil, product.ErrUnavailable
	}
	if err := o.AddItem(p.ID(), p.Name(), p.Unit(), quantity, p.PriceCents()); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) RemoveItem(ctx context.Context, orderID ID, productID product.ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if err := o.RemoveItem(productID); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) Confirm(ctx context.Context, id ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := o.Confirm(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) Fulfill(ctx context.Context, id ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := o.Fulfill(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) Cancel(ctx context.Context, id ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := o.Cancel(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}
