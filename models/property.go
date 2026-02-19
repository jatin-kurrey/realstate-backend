package models

import (
	"time"

	"gorm.io/gorm"
)

type Property struct {
	ID                       uint           `gorm:"primaryKey" json:"id"`
	Title                    string         `gorm:"not null" json:"title"`
	Status                   string         `gorm:"not null" json:"status"` // Sale, Rent
	Type                     string         `gorm:"not null" json:"type"`   // Residential, Commercial, Land
	Area                     float64        `gorm:"not null" json:"area"`
	Dimensions               string         `json:"dimensions"`
	AreaUnit                 string         `json:"area_unit"` // sqft, acre
	Frontage                 string         `json:"frontage"`
	LandUse                  string         `json:"land_use"`
	StreetName               string         `json:"street_name"`
	Village                  string         `json:"village"`
	RevenueInspectorCircle   string         `json:"revenue_inspector_circle"`
	Tehsil                   string         `json:"tehsil"`
	District                 string         `json:"district"`
	GoogleMapUrl             string         `json:"google_map_url"`
	DistanceFromMainLocation string         `json:"distance_from_main_location"`
	Description              string         `gorm:"type:text" json:"description"`
	Price                    float64        `gorm:"not null" json:"price"`
	Location                 string         `gorm:"not null" json:"location"`
	Landmark                 string         `json:"landmark"`
	ImageUrl                 string         `json:"imageUrl"`
	Images                   string         `gorm:"type:text" json:"images"` // Comma-separated URLs
	IsVerified               bool           `gorm:"default:false" json:"is_verified"`
	IsFeatured               bool           `gorm:"default:false" json:"is_featured"`
	IsActive                 bool           `gorm:"default:true" json:"is_active"`
	IsNegotiable             bool           `gorm:"default:false" json:"is_negotiable"`
	PostedAs                 string         `gorm:"default:'Owner'" json:"posted_as"` // Owner, Broker, Builder
	OwnerID                  uint           `json:"owner_id"`
	Owner                    User           `gorm:"foreignKey:OwnerID" json:"owner"`
	ExpiryDate               time.Time      `json:"expiry_date"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
}
