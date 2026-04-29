package product

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, name, description string, priceCents int64, unit string, available bool) (*Product, error) {
	p, err := New(name, description, priceCents, unit, available)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) GetByID(ctx context.Context, id ID) (*Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context, availableOnly bool) ([]*Product, error) {
	return s.repo.List(ctx, availableOnly)
}

func (s *Service) Update(ctx context.Context, id ID, name, description string, priceCents int64, unit string) (*Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := p.Update(name, description, priceCents, unit); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) SetAvailable(ctx context.Context, id ID, available bool) (*Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.SetAvailable(available)
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) SetImageURL(ctx context.Context, id ID, imageURL string) (*Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.SetImageURL(imageURL)
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id ID) error {
	return s.repo.Delete(ctx, id)
}
