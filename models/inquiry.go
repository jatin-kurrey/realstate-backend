package models

import (
	"time"

	"gorm.io/gorm"
)

type Inquiry struct {
	ID             uint             `gorm:"primaryKey" json:"id"`
	PropertyID     uint             `json:"property_id"`
	Property       Property         `gorm:"foreignKey:PropertyID" json:"property"`
	SeekerID       uint             `json:"seeker_id"`
	Seeker         User             `gorm:"foreignKey:SeekerID" json:"seeker"`
	OwnerID        uint             `json:"owner_id"`
	Owner          User             `gorm:"foreignKey:OwnerID" json:"owner"`
	InitialMessage string           `gorm:"type:text" json:"initial_message"`
	ExpectedDate   string           `json:"expected_date"`
	Budget         float64          `json:"budget"`
	Status         string           `gorm:"default:'Open'" json:"status"` // 'Open', 'Closed', 'Accepted'
	Messages       []InquiryMessage `gorm:"foreignKey:InquiryID" json:"messages"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	DeletedAt      gorm.DeletedAt   `gorm:"index" json:"-"`
}

type InquiryMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	InquiryID uint      `json:"inquiry_id"`
	SenderID  uint      `json:"sender_id"`
	Sender    User      `gorm:"foreignKey:SenderID" json:"sender"`
	Message   string    `gorm:"type:text" json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
