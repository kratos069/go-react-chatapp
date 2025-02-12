package models

import (
	"time"
)

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Email      string    `gorm:"unique;not null" json:"email"`
	FullName   string    `gorm:"not null" json:"fullname"`
	Password   string    `gorm:"not null;size:255" json:"password"`
	ProfilePic string    `gorm:"default:''" json:"profilePic"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
