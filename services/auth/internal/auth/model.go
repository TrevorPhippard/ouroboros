package auth

import "time"

type UserRecord struct {
	ID           string    `gorm:"primaryKey"`
	Email        string    `gorm:"unique;not null"`
	DisplayName  string    `gorm:"not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (UserRecord) TableName() string {
	return "users"
}
