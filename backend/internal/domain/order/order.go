package order

import (
	"time"

	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/fhardow/bread-order/internal/domain/product"
	"github.com/google/uuid"
)

type ID = uuid.UUID

type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusFulfilled Status = "fulfilled"
	StatusCancelled Status = "cancelled"
)

type Item struct {
	ProductID      product.ID
	ProductName    string
	Unit           string
	Quantity       int
	UnitPriceCents int64
}

func (i Item) TotalCents() int64 {
	return int64(i.Quantity) * i.UnitPriceCents
}

type Order struct {
	id        ID
	userID    uuid.UUID
	status    Status
	items     []Item
	createdAt time.Time
	updatedAt time.Time
}

func New(userID uuid.UUID) (*Order, error) {
	if userID == uuid.Nil {
		return nil, domerrors.BadRequest("user ID is required")
	}
	now := time.Now().UTC()
	return &Order{
		id:        uuid.New(),
		userID:    userID,
		status:    StatusPending,
		items:     nil,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func Reconstitute(id, userID uuid.UUID, status Status, items []Item, createdAt, updatedAt time.Time) *Order {
	return &Order{
		id:        id,
		userID:    userID,
		status:    status,
		items:     items,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (o *Order) ID() ID               { return o.id }
func (o *Order) UserID() uuid.UUID    { return o.userID }
func (o *Order) Status() Status       { return o.status }
func (o *Order) Items() []Item        { return o.items }
func (o *Order) CreatedAt() time.Time { return o.createdAt }
func (o *Order) UpdatedAt() time.Time { return o.updatedAt }

func (o *Order) TotalCents() int64 {
	var total int64
	for _, item := range o.items {
		total += item.TotalCents()
	}
	return total
}

func (o *Order) AddItem(productID product.ID, productName string, unit string, quantity int, unitPriceCents int64) error {
	if o.status != StatusPending {
		return domerrors.Forbidden("can only modify pending orders")
	}
	if quantity <= 0 {
		return domerrors.BadRequest("quantity must be positive")
	}
	for i, item := range o.items {
		if item.ProductID == productID {
			o.items[i].Quantity += quantity
			o.updatedAt = time.Now().UTC()
			return nil
		}
	}
	o.items = append(o.items, Item{
		ProductID:      productID,
		ProductName:    productName,
		Unit:           unit,
		Quantity:       quantity,
		UnitPriceCents: unitPriceCents,
	})
	o.updatedAt = time.Now().UTC()
	return nil
}

func (o *Order) RemoveItem(productID product.ID) error {
	if o.status != StatusPending {
		return domerrors.Forbidden("can only modify pending orders")
	}
	for i, item := range o.items {
		if item.ProductID == productID {
			o.items = append(o.items[:i], o.items[i+1:]...)
			o.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return domerrors.NotFound("item with product ID %s not found in order", productID)
}

func (o *Order) Confirm() error {
	if o.status != StatusPending {
		return domerrors.BadRequest("only pending orders can be confirmed")
	}
	if len(o.items) == 0 {
		return domerrors.BadRequest("cannot confirm an empty order")
	}
	o.status = StatusConfirmed
	o.updatedAt = time.Now().UTC()
	return nil
}

func (o *Order) Fulfill() error {
	if o.status != StatusConfirmed {
		return domerrors.BadRequest("only confirmed orders can be fulfilled")
	}
	o.status = StatusFulfilled
	o.updatedAt = time.Now().UTC()
	return nil
}

func (o *Order) Cancel() error {
	if o.status == StatusFulfilled {
		return domerrors.BadRequest("fulfilled orders cannot be cancelled")
	}
	if o.status == StatusCancelled {
		return domerrors.BadRequest("order is already cancelled")
	}
	o.status = StatusCancelled
	o.updatedAt = time.Now().UTC()
	return nil
}
