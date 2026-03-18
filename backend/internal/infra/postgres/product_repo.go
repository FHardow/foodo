package postgres

import (
	"context"
	"errors"

	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/fhardow/bread-order/internal/infra/postgres/models"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type productRepo struct{ db *gorm.DB }

func NewProductRepo(db *gorm.DB) product.Repository {
	return &productRepo{db: db}
}

func (r *productRepo) FindByID(ctx context.Context, id product.ID) (*product.Product, error) {
	var m models.Product
	err := dbFromCtx(ctx, r.db).First(&m, "id = ?", id.String()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domerrors.NotFound("product %s not found", id)
		}
		return nil, err
	}
	return productToDomain(&m)
}

func (r *productRepo) List(ctx context.Context, availableOnly bool) ([]*product.Product, error) {
	var ms []models.Product
	q := dbFromCtx(ctx, r.db).Order("name asc")
	if availableOnly {
		q = q.Where("available = true")
	}
	if err := q.Find(&ms).Error; err != nil {
		return nil, err
	}
	products := make([]*product.Product, 0, len(ms))
	for i := range ms {
		p, err := productToDomain(&ms[i])
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *productRepo) Save(ctx context.Context, p *product.Product) error {
	m := productToModel(p)
	return dbFromCtx(ctx, r.db).Save(&m).Error
}

func (r *productRepo) Delete(ctx context.Context, id product.ID) error {
	return dbFromCtx(ctx, r.db).Delete(&models.Product{}, "id = ?", id.String()).Error
}

func productToDomain(m *models.Product) (*product.Product, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	return product.Reconstitute(id, m.Name, m.Description, m.PriceCents, m.Unit, m.Available, m.CreatedAt, m.UpdatedAt), nil
}

func productToModel(p *product.Product) models.Product {
	return models.Product{
		ID:          p.ID().String(),
		Name:        p.Name(),
		Description: p.Description(),
		PriceCents:  p.PriceCents(),
		Unit:        p.Unit(),
		Available:   p.Available(),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
	}
}
