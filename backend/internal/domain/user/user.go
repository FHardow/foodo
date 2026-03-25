package user

import (
	"time"

	domerrors "github.com/fhardow/bread-order/pkg/errors"
	"github.com/google/uuid"
)

type ID = uuid.UUID

type Role string

const (
	RoleCustomer Role = "customer"
	RoleOwner    Role = "owner"
)

type User struct {
	id        ID
	name      string
	email     string
	phone     string
	role      Role
	createdAt time.Time
	updatedAt time.Time
}

func New(name, email, phone string) (*User, error) {
	if name == "" {
		return nil, domerrors.BadRequest("name is required")
	}
	if email == "" {
		return nil, domerrors.BadRequest("email is required")
	}
	now := time.Now().UTC()
	return &User{
		id:        uuid.New(),
		name:      name,
		email:     email,
		phone:     phone,
		role:      RoleCustomer,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Reconstitute rebuilds a User from persistence without re-running validation.
func Reconstitute(id ID, name, email, phone string, role Role, createdAt, updatedAt time.Time) *User {
	return &User{
		id:        id,
		name:      name,
		email:     email,
		phone:     phone,
		role:      role,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (u *User) ID() ID               { return u.id }
func (u *User) Name() string         { return u.name }
func (u *User) Email() string        { return u.email }
func (u *User) Phone() string        { return u.phone }
func (u *User) Role() Role           { return u.role }
func (u *User) IsOwner() bool        { return u.role == RoleOwner }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

func (u *User) UpdateContact(name, email, phone string) error {
	if name == "" {
		return domerrors.BadRequest("name is required")
	}
	if email == "" {
		return domerrors.BadRequest("email is required")
	}
	u.name = name
	u.email = email
	u.phone = phone
	u.updatedAt = time.Now().UTC()
	return nil
}
