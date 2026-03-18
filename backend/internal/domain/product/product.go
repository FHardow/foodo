package product

import (
	"time"

	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
)

type ID = uuid.UUID

type Product struct {
	id          ID
	name        string
	description string
	// price in cents to avoid floating-point issues
	priceCents  int64
	unit        string
	available   bool
	createdAt   time.Time
	updatedAt   time.Time
}

func New(name, description string, priceCents int64, unit string) (*Product, error) {
	if name == "" {
		return nil, domerrors.BadRequest("name is required")
	}
	if priceCents < 0 {
		return nil, domerrors.BadRequest("price cannot be negative")
	}
	if unit == "" {
		return nil, domerrors.BadRequest("unit is required")
	}
	now := time.Now().UTC()
	return &Product{
		id:          uuid.New(),
		name:        name,
		description: description,
		priceCents:  priceCents,
		unit:        unit,
		available:   true,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func Reconstitute(id ID, name, description string, priceCents int64, unit string, available bool, createdAt, updatedAt time.Time) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		priceCents:  priceCents,
		unit:        unit,
		available:   available,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (p *Product) ID() ID              { return p.id }
func (p *Product) Name() string        { return p.name }
func (p *Product) Description() string { return p.description }
func (p *Product) PriceCents() int64   { return p.priceCents }
func (p *Product) Unit() string        { return p.unit }
func (p *Product) Available() bool     { return p.available }
func (p *Product) CreatedAt() time.Time { return p.createdAt }
func (p *Product) UpdatedAt() time.Time { return p.updatedAt }

func (p *Product) Update(name, description string, priceCents int64, unit string) error {
	if name == "" {
		return domerrors.BadRequest("name is required")
	}
	if priceCents < 0 {
		return domerrors.BadRequest("price cannot be negative")
	}
	if unit == "" {
		return domerrors.BadRequest("unit is required")
	}
	p.name = name
	p.description = description
	p.priceCents = priceCents
	p.unit = unit
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Product) SetAvailable(available bool) {
	p.available = available
	p.updatedAt = time.Now().UTC()
}
