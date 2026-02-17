package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	Name               string         `gorm:"not null" json:"name"`
	Email              string         `gorm:"uniqueIndex;not null" json:"email"`
	Password           string         `gorm:"not null" json:"-"`
	Phone              string         `json:"phone"`
	Role               string         `gorm:"default:'seeker'" json:"role"`                  // 'seeker' or 'owner' or 'admin' or 'developer'
	CompanyName        string         `json:"company_name"`                                  // Optional for developers
	PublicPreference   string         `gorm:"default:'Anonymized'" json:"public_preference"` // 'Anonymized' or 'Full'
	ContactPreference  string         `gorm:"default:'In-app'" json:"contact_preference"`    // 'In-app' or 'Email' or 'Phone'
	EmailNotifications bool           `gorm:"default:true" json:"email_notifications"`
	InAppNotifications bool           `gorm:"default:true" json:"in_app_notifications"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}
