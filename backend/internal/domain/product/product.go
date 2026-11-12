package product

import (
	"strings"
	"time"

	domerrors "github.com/fhardow/foodo/pkg/errors"
	"github.com/google/uuid"
)

type ID = uuid.UUID

type Product struct {
	id          ID
	name        string
	description string
	priceCents  int64
	unit        string
	available   bool
	imageURL    string
	categories  []string
	createdAt   time.Time
	updatedAt   time.Time
}

func New(name, description string, priceCents int64, unit string, available bool) (*Product, error) {
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
		available:   available,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func Reconstitute(id ID, name, description string, priceCents int64, unit string, available bool, imageURL string, categories []string, createdAt, updatedAt time.Time) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		priceCents:  priceCents,
		unit:        unit,
		available:   available,
		imageURL:    imageURL,
		categories:  categories,
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
func (p *Product) ImageURL() string    { return p.imageURL }

// Categories returns a copy of the product's category names.
func (p *Product) Categories() []string { return append([]string(nil), p.categories...) }
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

func (p *Product) SetImageURL(url string) {
	p.imageURL = url
	p.updatedAt = time.Now().UTC()
}

const (
	MaxCategories         = 10
	MaxCategoryNameLength = 40
)

// SetCategories replaces the product's categories. Names are trimmed, internal
// whitespace is collapsed, empty names are dropped and duplicates (compared
// case-insensitively) are removed, keeping the first occurrence.
func (p *Product) SetCategories(names []string) error {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.Join(strings.Fields(n), " ")
		if n == "" {
			continue
		}
		if len([]rune(n)) > MaxCategoryNameLength {
			return domerrors.BadRequest("category %q is longer than %d characters", n, MaxCategoryNameLength)
		}
		key := strings.ToLower(n)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, n)
	}
	if len(out) > MaxCategories {
		return domerrors.BadRequest("a product can have at most %d categories", MaxCategories)
	}
	p.categories = out
	p.updatedAt = time.Now().UTC()
	return nil
}
