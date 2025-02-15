package models

import "time"

type ConversationParticipant struct {
	ConversationID string    `gorm:"primaryKey;type:varchar(36)" json:"conversation_id"`
	UserID         string    `gorm:"primaryKey;type:varchar(36)" json:"user_id"` // 用户 ID
	LastRead       time.Time `gorm:"nullable" json:"last_read"`                  // 用户最后一次阅读时间
	JoinedAt       time.Time `gorm:"autoCreateTime" json:"joined_at"`            // 用户加入会话的时间
}
