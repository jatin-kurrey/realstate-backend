package models

import (
	"time"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // 'inquiry', 'payment', 'system'
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
