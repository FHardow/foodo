package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/fhardow/bread-order/internal/domain/order"
	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/fhardow/bread-order/internal/infra/postgres/models"
	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type orderRepo struct{ db *gorm.DB }

func NewOrderRepo(db *gorm.DB) order.Repository {
	return &orderRepo{db: db}
}

func (r *orderRepo) FindByID(ctx context.Context, id order.ID) (*order.Order, error) {
	var m models.Order
	err := dbFromCtx(ctx, r.db).
		Joins("LEFT JOIN users ON users.id = orders.user_id").
		Select("orders.*, users.name AS user_name").
		Preload("Items").
		First(&m, "orders.id = ?", id.String()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domerrors.NotFound("order %s not found", id)
		}
		return nil, err
	}
	return orderToDomain(&m)
}

func (r *orderRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*order.Order, error) {
	var ms []models.Order
	err := dbFromCtx(ctx, r.db).
		Joins("LEFT JOIN users ON users.id = orders.user_id").
		Select("orders.*, users.name AS user_name").
		Preload("Items").
		Where("orders.user_id = ?", userID.String()).
		Order("orders.created_at desc").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}
	return ordersSliceToDomain(ms)
}

func (r *orderRepo) List(ctx context.Context) ([]*order.Order, error) {
	var ms []models.Order
	err := dbFromCtx(ctx, r.db).
		Joins("LEFT JOIN users ON users.id = orders.user_id").
		Select("orders.*, users.name AS user_name").
		Preload("Items").
		Order("orders.created_at desc").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}
	return ordersSliceToDomain(ms)
}

func (r *orderRepo) Save(ctx context.Context, o *order.Order) error {
	m := orderToModel(o)
	db := dbFromCtx(ctx, r.db)

	// Use a transaction to replace items atomically.
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&m).Error; err != nil {
			return err
		}
		// Delete existing items and re-insert to handle removals cleanly.
		if err := tx.Where("order_id = ?", m.ID).Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}
		if len(m.Items) > 0 {
			return tx.Create(&m.Items).Error
		}
		return nil
	})
}

func (r *orderRepo) Delete(ctx context.Context, id order.ID) error {
	return dbFromCtx(ctx, r.db).Delete(&models.Order{}, "id = ?", id.String()).Error
}

func orderToDomain(m *models.Order) (*order.Order, error) {
	id, err := uuid.Parse(m.ID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(m.UserID)
	if err != nil {
		return nil, err
	}
	items := make([]order.Item, 0, len(m.Items))
	for _, mi := range m.Items {
		pid, err := uuid.Parse(mi.ProductID)
		if err != nil {
			return nil, err
		}
		items = append(items, order.Item{
			ProductID:      product.ID(pid),
			ProductName:    mi.ProductName,
			Unit:           mi.Unit,
			Quantity:       mi.Quantity,
			UnitPriceCents: mi.UnitPriceCents,
		})
	}
	return order.Reconstitute(id, userID, m.UserName, order.Status(m.Status), items, m.CreatedAt, m.UpdatedAt), nil
}

func ordersSliceToDomain(ms []models.Order) ([]*order.Order, error) {
	orders := make([]*order.Order, 0, len(ms))
	for i := range ms {
		o, err := orderToDomain(&ms[i])
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func orderToModel(o *order.Order) models.Order {
	items := make([]models.OrderItem, 0, len(o.Items()))
	for _, item := range o.Items() {
		items = append(items, models.OrderItem{
			OrderID:        o.ID().String(),
			ProductID:      item.ProductID.String(),
			ProductName:    item.ProductName,
			Unit:           item.Unit,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			CreatedAt:      time.Now().UTC(),
		})
	}
	return models.Order{
		ID:        o.ID().String(),
		UserID:    o.UserID().String(),
		Status:    string(o.Status()),
		CreatedAt: o.CreatedAt(),
		UpdatedAt: o.UpdatedAt(),
		Items:     items,
	}
}
