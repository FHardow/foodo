package models

import "time"

type User struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	Name      string    `gorm:"not null"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Phone     string
	Role      string    `gorm:"not null;default:customer"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (User) TableName() string { return "users" }
