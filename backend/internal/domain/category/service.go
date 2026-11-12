package category

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]*Category, error) {
	return s.repo.List(ctx)
}

func (s *Service) Delete(ctx context.Context, id ID) error {
	return s.repo.Delete(ctx, id)
}
