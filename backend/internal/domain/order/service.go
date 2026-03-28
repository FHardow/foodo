package order

import (
	"context"

	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/google/uuid"
)

type Service struct {
	repo        Repository
	productRepo product.Repository
}

func NewService(repo Repository, productRepo product.Repository) *Service {
	return &Service{repo: repo, productRepo: productRepo}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID) (*Order, error) {
	o, err := New(userID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
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

func (s *Service) Accept(ctx context.Context, id ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := o.Accept(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) StartProgress(ctx context.Context, id ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := o.StartProgress(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) Finish(ctx context.Context, id ID) (*Order, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := o.Finish(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}
