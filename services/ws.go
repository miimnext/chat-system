package services

import (
	"chat-system/config"
	"chat-system/models"
	"fmt"
	"sort"
)

// 获取或创建会话
func GetOrCreateConversation(userID1, userID2 string) (string, error) {
	// 确保两个用户ID有序，避免重复会话
	userIDs := []string{userID1, userID2}
	sort.Strings(userIDs)
	// 生成私聊会话ID
	conversationID := fmt.Sprintf("%s_%s", userIDs[0], userIDs[1])
	// 查询是否已经存在会话
	var conversation models.Conversation
	if err := config.DB.Where("conversation_id = ?", conversationID).First(&conversation).Error; err != nil {
		// 如果会话不存在，创建新的
		newConversation := models.Conversation{
			ConversationID: conversationID,
			Type:           "private", // 这里直接设置为私聊
			ParticipantA:   userID1,
			ParticipantB:   userID2,
		}
		if err := config.DB.Create(&newConversation).Error; err != nil {
			return "", fmt.Errorf("failed to create conversation: %v", err)
		}
		return newConversation.ConversationID, nil
	}

	// 如果已存在，返回现有会话ID
	return conversation.ConversationID, nil
}

// 生成会话ID
func GenerateConversationID(userID1, userID2 string) string {
	userIDs := []string{userID1, userID2}
	sort.Strings(userIDs) // 确保顺序一致
	return fmt.Sprintf("%s_%s", userIDs[0], userIDs[1])
}

// GenerateGroupConversationID 生成群聊会话ID
func GenerateGroupConversationID(groupID string) string {
	return fmt.Sprintf("group_%s", groupID)
}
