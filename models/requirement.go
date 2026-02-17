package models

import (
	"time"

	"gorm.io/gorm"
)

type Requirement struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Purpose       string         `gorm:"not null" json:"purpose"` // Buy, Rent
	Type          string         `gorm:"not null" json:"type"`
	MinArea       float64        `json:"minArea"`
	MaxArea       float64        `json:"maxArea"`
	MinBudget     float64        `json:"minBudget"`
	MaxBudget     float64        `json:"maxBudget"`
	Location      string         `json:"location"`
	Description   string         `json:"description"`
	ContactMethod string         `json:"contactMethod"`
	UserID        uint           `json:"user_id"`
	User          User           `gorm:"foreignKey:UserID" json:"user"`
	IsVerified    bool           `gorm:"default:false" json:"is_verified"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
