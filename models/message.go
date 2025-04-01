package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	MessageID      string    `json:"message_id" gorm:"primaryKey"` // Message ID (as string)
	ConversationID string    `json:"conversation_id"`              // Conversation ID (as string)
	SenderID       string    `json:"sender_id"`                    // Sender User ID (as uint)
	ReceiverID     string    `json:"receiver_id"`                  // Receiver User ID (as uint)
	GroupID        string    `json:"group_id,omitempty"`
	Type           string    `json:"type"`
	IsRead         bool      `json:"is_read" gorm:"default:false"` // 是否已读
	Content        string    `json:"content"`                      // Message content
	MessageType    string    `json:"message_type"`                 // Message type (text, image, etc.)
	Status         string    `json:"status"`                       // Message status (sent, read, etc.)
	CreatedAt      time.Time `json:"created_at"`
}
