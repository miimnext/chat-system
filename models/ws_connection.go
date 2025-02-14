package models

import (
	"time"

	"gorm.io/gorm"
)

// WSConnection WebSocket 连接模型
type WSConnection struct {
	gorm.Model
	ConnectionID string    `json:"connection_id" gorm:"primaryKey"` // 连接ID
	UserID       uint      `json:"user_id"`                         // 用户ID
	ConnectedAt  time.Time `json:"connected_at"`                    // 连接时间
}
