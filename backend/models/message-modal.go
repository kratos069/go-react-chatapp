package models

import (
	"time"
)

type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   uint      `gorm:"not null" json:"senderId"`
	ReceiverID uint      `gorm:"not null" json:"receiverId"`
	Text       string    `json:"text"`
	Image      string    `json:"image"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
