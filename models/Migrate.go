package models

import (
	"chat-system/config"
	"log"
)

// Migrate 执行数据库迁移操作
// 该函数使用 GORM 自动迁移功能创建数据库表
func Migrate() {
	// 自动迁移数据库模型到数据库
	err := config.DB.AutoMigrate(
		&User{},                    // 用户表
		&Message{},                 // 消息表
		&Conversation{},            // 会话表
		&Group{},                   // 群组表
		&WSConnection{},            // WebSocket 连接表
		&GroupMember{},             // WebSocket 连接表
		&ConversationParticipant{}, // WebSocket 连接表
	)
	config.DB.Exec("ALTER TABLE users AUTO_INCREMENT = 10000;")
	if err != nil {
		log.Fatalf("Error migrating database: %v", err) // 错误处理
	} else {
		log.Println("Database migration completed successfully.") // 迁移成功提示
	}
}
