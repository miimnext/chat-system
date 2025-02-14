package models

import "time"

// Conversation 会话模型
type Conversation struct {
	ConversationID string    `gorm:"primaryKey;type:varchar(36)" json:"conversation_id"`
	Type           string    `gorm:"type:varchar(10);index" json:"type"`                   // "private" or "group"
	ParticipantA   string    `gorm:"type:varchar(36);index;nullable" json:"participant_a"` // 仅私聊用
	ParticipantB   string    `gorm:"type:varchar(36);index;nullable" json:"participant_b"` // 仅私聊用
	GroupID        string    `gorm:"type:varchar(36);index;nullable" json:"group_id"`      // 仅群聊用
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}
