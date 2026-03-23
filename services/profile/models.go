package main

import (
	"time"
)

type Profile struct {
	ID        		uint      `gorm:"primaryKey"`
	UserId        string    `gorm:"type:varchar(50);not null"`
	DisplayName   string    `gorm:"type:varchar(50);not null"`
	AvatarUrl     string    `gorm:"type:varchar(50);not null"`
	Bio           string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}