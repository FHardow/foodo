package models

import "time"

type Product struct {
	ID          string    `gorm:"primaryKey;type:uuid"`
	Name        string    `gorm:"not null"`
	Description string
	PriceCents  int64     `gorm:"not null"`
	Unit        string    `gorm:"not null"`
	Available   bool      `gorm:"not null;default:true"`
	ImageURL    string
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Product) TableName() string { return "products" }
