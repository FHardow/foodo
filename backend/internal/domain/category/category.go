package category

import (
	"time"

	"github.com/google/uuid"
)

type ID = uuid.UUID

// Category is a label that can be attached to products. Categories are created
// implicitly when an admin assigns a new name to a product, and can be deleted
// explicitly by an admin.
type Category struct {
	id        ID
	name      string
	createdAt time.Time
}

func Reconstitute(id ID, name string, createdAt time.Time) *Category {
	return &Category{id: id, name: name, createdAt: createdAt}
}

func (c *Category) ID() ID               { return c.id }
func (c *Category) Name() string         { return c.name }
func (c *Category) CreatedAt() time.Time { return c.createdAt }
