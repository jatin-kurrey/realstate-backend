package models

import (
	"time"

	"gorm.io/gorm"
)

type Property struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Title        string         `gorm:"not null" json:"title"`
	Status       string         `gorm:"not null" json:"status"` // Sale, Rent
	Type         string         `gorm:"not null" json:"type"`   // Residential, Commercial, Land
	Area         float64        `gorm:"not null" json:"area"`
	Dimensions   string         `json:"dimensions"`
	Description  string         `gorm:"type:text" json:"description"`
	Price        float64        `gorm:"not null" json:"price"`
	Location     string         `gorm:"not null" json:"location"`
	Landmark     string         `json:"landmark"`
	ImageUrl     string         `json:"imageUrl"`
	Images       string         `gorm:"type:text" json:"images"` // Comma-separated URLs
	IsVerified   bool           `gorm:"default:false" json:"is_verified"`
	IsFeatured   bool           `gorm:"default:false" json:"is_featured"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	IsNegotiable bool           `gorm:"default:false" json:"is_negotiable"`
	OwnerID      uint           `json:"owner_id"`
	Owner        User           `gorm:"foreignKey:OwnerID" json:"owner"`
	ExpiryDate   time.Time      `json:"expiry_date"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
