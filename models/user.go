package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string         `json:"username" gorm:"unique;not null"`
	Password  string         `json:"password" gorm:"not null"`
	Email     string         `json:"email"`
	Phone     string         `json:"phone"`
	AvatarURL string         `json:"avatar_url"`
	Status    string         `json:"status" gorm:"default:'offline'"`
	LastLogin *time.Time     `json:"last_login" gorm:"default:NULL"` // 允许 NULL
	Bio       string         `json:"bio"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
