package models

import (
	"time"
)

// SiteConfig stores global key-value settings for the application
type SiteConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"unique;not null" json:"key"`     // e.g., "site_name", "contact_email"
	Value     string    `gorm:"type:text" json:"value"`         // e.g., "RJG Property Connect"
	Group     string    `gorm:"default:'general'" json:"group"` // e.g., "general", "social", "seo"
	Type      string    `gorm:"default:'text'" json:"type"`     // e.g., "text", "image", "boolean"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PageContent stores dynamic content for specific pages/sections
type PageContent struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PageKey    string    `gorm:"index" json:"page_key"`        // e.g., "home", "about", "contact"
	SectionKey string    `gorm:"index" json:"section_key"`     // e.g., "hero", "testimonials"
	Content    string    `gorm:"type:text" json:"content"`     // JSON string or simple text content
	Schema     string    `gorm:"default:'text'" json:"schema"` // Used to hint frontend editor type (text, rich-text, json-list)
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
