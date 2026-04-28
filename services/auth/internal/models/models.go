package models

import (
	"time"
)

type DBUser struct {
	ID           string    `gorm:"primaryKey"`
	Email        string    `gorm:"not null"`
	PasswordHash string    `gorm:"not null"`
	DisplayName  string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (DBUser) TableName() string {
	return "users"
}
