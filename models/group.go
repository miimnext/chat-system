package models

import (
	"gorm.io/gorm"
)

// Group 群组模型
type Group struct {
	gorm.Model
	GroupID     uint   `json:"group_id" gorm:"primaryKey"` // 群组ID
	GroupName   string `json:"group_name"`                 // 群组名称
	OwnerID     uint   `json:"owner_id"`                   // 群主ID
	Description string `json:"description"`                // 群组描述
}
