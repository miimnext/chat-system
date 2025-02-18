package models

import "time"

type Conversation struct {
	ConversationID string    `gorm:"primaryKey;type:varchar(36)" json:"conversation_id"`
	Type           string    `gorm:"type:varchar(10);index" json:"type"`                   // "private" or "group"
	ParticipantA   string    `gorm:"type:varchar(36);index;nullable" json:"participant_a"` // 仅私聊用
	ParticipantB   string    `gorm:"type:varchar(36);index;nullable" json:"participant_b"` // 仅私聊用
	GroupID        string    `gorm:"type:varchar(36);index;nullable" json:"group_id"`      // 仅群聊用
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	// 关联用户A和用户B
	ParticipantAUser User `gorm:"foreignKey:ParticipantA;references:ID" json:"participant_a_user"`
	ParticipantBUser User `gorm:"foreignKey:ParticipantB;references:ID" json:"participant_b_user"`
}
