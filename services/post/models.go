package main

import (
	"time"
)

// Post maps to the "posts" table in our post_db
type Post struct {
	ID        uint      `gorm:"primaryKey"`
	AuthorID  string    `gorm:"type:varchar(50);not null"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type DBPost struct {
	ID        string    `gorm:"primaryKey"`
	AuthorID  string    `gorm:"column:author_id;not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

type DBComment struct {
	ID        string    `gorm:"primaryKey"`
	PostID    string    `gorm:"column:post_id;not null"`
	AuthorID  string    `gorm:"column:author_id;not null"`
	Content   string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}