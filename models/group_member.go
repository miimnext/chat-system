package models

import (
	"gorm.io/gorm"
)

// GroupMember 群组成员模型
type GroupMember struct {
	gorm.Model
	GroupID uint `json:"group_id"`
	UserID  uint `json:"user_id"`
}
