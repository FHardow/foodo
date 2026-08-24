package models

import "time"

type PushSubscription struct {
	ID           string    `gorm:"primaryKey;type:uuid"`
	UserID       string    `gorm:"not null;index;type:uuid"`
	Endpoint     string    `gorm:"not null"`
	EndpointHash string    `gorm:"not null;uniqueIndex;column:endpoint_hash"`
	P256dh       string    `gorm:"not null;column:p256dh"`
	Auth         string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (PushSubscription) TableName() string { return "push_subscriptions" }
