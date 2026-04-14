package models

import (
	"time"
)

type Profile struct {
	ID          string       `gorm:"primaryKey"`
	UserId      string       `gorm:"type:varchar(50);not null"`
	DisplayName string       `gorm:"type:varchar(50);not null"`
	AvatarUrl   string       `gorm:"type:text;not null"`
	Bio         string       `gorm:"type:text;not null"`
	Headline    string       `gorm:"type:text"`
	About       string       `gorm:"type:text"`
	Experiences []Experience `gorm:"foreignKey:ProfileID"`
	CreatedAt   time.Time    `gorm:"default:CURRENT_TIMESTAMP"`
}

type Experience struct {
	ID        uint   `gorm:"primaryKey"`
	ProfileID string `gorm:"not null"`
	Title     string `gorm:"type:varchar(100);not null"`
	Company   string `gorm:"type:varchar(100);not null"`
	StartDate time.Time
	EndDate   *time.Time
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
