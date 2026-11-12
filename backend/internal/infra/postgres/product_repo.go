package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/fhardow/foodo/internal/domain/product"
	"github.com/fhardow/foodo/internal/infra/postgres/models"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type productRepo struct{ db *gorm.DB }

func NewProductRepo(db *gorm.DB) product.Repository {
	return &productRepo{db: db}
}

func (r *productRepo) FindByID(ctx context.Context, id product.ID) (*product.Product, error) {
	var m models.Product
	err := preloadCategories(dbFromCtx(ctx, r.db)).First(&m, "id = ?", id.String()).Error
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
	q := preloadCategories(dbFromCtx(ctx, r.db)).Order("name asc")
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

// Save upserts the product and replaces its category links. Category names
// that don't exist yet (compared case-insensitively) are created.
func (r *productRepo) Save(ctx context.Context, p *product.Product) error {
	return dbFromCtx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		m := productToModel(p)
		if err := tx.Omit(clause.Associations).Save(&m).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM product_categories WHERE product_id = ?", m.ID).Error; err != nil {
			return err
		}
		for _, name := range p.Categories() {
			catID, err := findOrCreateCategory(tx, name)
			if err != nil {
				return err
			}
			if err := tx.Exec(
				"INSERT INTO product_categories (product_id, category_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
				m.ID, catID,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *productRepo) Delete(ctx context.Context, id product.ID) error {
	return dbFromCtx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM product_categories WHERE product_id = ?", id.String()).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Product{}, "id = ?", id.String()).Error
	})
}

func preloadCategories(db *gorm.DB) *gorm.DB {
	return db.Preload("Categories", func(db *gorm.DB) *gorm.DB {
		return db.Order("lower(categories.name) asc")
	})
}

// findOrCreateCategory returns the ID of the category matching name
// case-insensitively, creating it if missing.
func findOrCreateCategory(tx *gorm.DB, name string) (string, error) {
	var existing models.Category
	err := tx.Where("lower(name) = lower(?)", name).Take(&existing).Error
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	c := models.Category{ID: uuid.NewString(), Name: name, CreatedAt: time.Now().UTC()}
	if err := tx.Create(&c).Error; err != nil {
		return "", err
	}
	return c.ID, nil
}

func productToDomain(m *models.Product) (*product.Product, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	categories := make([]string, 0, len(m.Categories))
	for _, c := range m.Categories {
		categories = append(categories, c.Name)
	}
	return product.Reconstitute(id, m.Name, m.Description, m.PriceCents, m.Unit, m.Available, m.ImageURL, categories, m.CreatedAt, m.UpdatedAt), nil
}

func productToModel(p *product.Product) models.Product {
	return models.Product{
		ID:          p.ID().String(),
		Name:        p.Name(),
		Description: p.Description(),
		PriceCents:  p.PriceCents(),
		Unit:        p.Unit(),
		Available:   p.Available(),
		ImageURL:    p.ImageURL(),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
	}
}
