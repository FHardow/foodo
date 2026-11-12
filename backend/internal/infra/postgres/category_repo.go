package postgres

import (
	"context"

	"github.com/fhardow/foodo/internal/domain/category"
	"github.com/fhardow/foodo/internal/infra/postgres/models"
	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type categoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) category.Repository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) List(ctx context.Context) ([]*category.Category, error) {
	var ms []models.Category
	if err := dbFromCtx(ctx, r.db).Order("lower(name) asc").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]*category.Category, 0, len(ms))
	for _, m := range ms {
		id, err := uuid.Parse(m.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, category.Reconstitute(id, m.Name, m.CreatedAt))
	}
	return out, nil
}

func (r *categoryRepo) Delete(ctx context.Context, id category.ID) error {
	return dbFromCtx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM product_categories WHERE category_id = ?", id.String()).Error; err != nil {
			return err
		}
		res := tx.Delete(&models.Category{}, "id = ?", id.String())
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domerrors.NotFound("category %s not found", id)
		}
		return nil
	})
}
