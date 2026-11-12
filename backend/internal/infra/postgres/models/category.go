package models

import "time"

type Category struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

func (Category) TableName() string { return "categories" }
