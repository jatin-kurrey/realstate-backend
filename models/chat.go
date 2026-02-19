package models

import (
	"time"
)

type ChatThread struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	PropertyID     *uint         `json:"property_id,omitempty"`
	Property       *Property     `gorm:"foreignKey:PropertyID" json:"property,omitempty"`
	Participant1ID uint          `json:"participant1_id"`
	Participant1   User          `gorm:"foreignKey:Participant1ID" json:"participant1"`
	Participant2ID uint          `json:"participant2_id"`
	Participant2   User          `gorm:"foreignKey:Participant2ID" json:"participant2"`
	LastMessage    string        `json:"last_message"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Messages       []ChatMessage `gorm:"foreignKey:ThreadID" json:"messages"`
	UnreadCount    int           `gorm:"-" json:"unread_count"` // Computed field
	IsTyping1      bool          `gorm:"default:false" json:"is_typing1"` // Participant1 typing status
	IsTyping2      bool          `gorm:"default:false" json:"is_typing2"` // Participant2 typing status
	LastActivity   time.Time     `json:"last_activity"` // Last activity timestamp
}

type ChatMessage struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ThreadID    uint           `json:"thread_id"`
	SenderID    uint           `json:"sender_id"`
	Sender      User           `gorm:"foreignKey:SenderID" json:"sender"`
	Content     string         `gorm:"type:text" json:"content"`
	Status      MessageStatus  `gorm:"default:'sent'" json:"status"` // Message status: sent, delivered, read
	IsEdited    bool           `gorm:"default:false" json:"is_edited"`
	EditedAt    *time.Time     `json:"edited_at,omitempty"`
	IsDeleted   bool           `gorm:"default:false" json:"is_deleted"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	DeliveredAt *time.Time     `json:"delivered_at,omitempty"`
	ReadAt      *time.Time     `json:"read_at,omitempty"`
	ReplyToID   *uint          `json:"reply_to_id,omitempty"`
	ReplyTo     *ChatMessage   `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
}

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

type TypingStatus struct {
	UserID   uint  `json:"user_id"`
	ThreadID uint  `json:"thread_id"`
	IsTyping bool  `json:"is_typing"`
}

type MessageSearchResult struct {
	Message   ChatMessage `json:"message"`
	Thread    ChatThread  `json:"thread"`
	Highlight string      `json:"highlight"`
}
