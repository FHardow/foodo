package models

import "time"

type Order struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	UserID    string    `gorm:"not null;index;type:uuid"`
	Status    string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Items    []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	UserName string      `gorm:"-"` // populated via JOIN with users table, not persisted
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	OrderID        string    `gorm:"not null;index;type:uuid"`
	ProductID      string    `gorm:"not null;type:uuid"`
	ProductName    string    `gorm:"not null"`
	Unit           string    `gorm:"not null;default:''"`
	Quantity       int       `gorm:"not null"`
	UnitPriceCents int64     `gorm:"not null"`
	CreatedAt      time.Time `gorm:"not null"`
}

func (OrderItem) TableName() string { return "order_items" }
