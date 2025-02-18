package controllers

import (
	"chat-system/config"
	"chat-system/models"
	"chat-system/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 发送消息
func SendMessage(c *gin.Context) {
	var input struct {
		ConversationID string `json:"conversation_id" binding:"required"`
		SenderID       string `json:"sender_id" binding:"required"`
		ReceiverID     string `json:"receiver_id" binding:"required"`
		Content        string `json:"content" binding:"required"`
		MessageType    string `json:"message_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := models.Message{
		ConversationID: input.ConversationID,
		SenderID:       input.SenderID,
		ReceiverID:     input.ReceiverID,
		Content:        input.Content,
		MessageType:    input.MessageType,
		Status:         "sent", // 初始状态为已发送
	}

	if err := config.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// 更新会话中的最后一条消息ID
	if err := config.DB.Model(&models.Conversation{}).Where("conversation_id = ?", input.ConversationID).
		Update("last_message_id", message.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Message sent successfully",
		"message_id": message.MessageID,
	})
}

// 获取会话的消息列表
func GetMessagesByConversationID(c *gin.Context) {
	conversationID := c.Param("conversation_id") // 从 URL 获取 conversation_id
	user, exists := c.Get("user")
	if !exists {
		// 如果用户信息不存在，返回 404 错误
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 假设 user 是 *models.User 类型，进行类型断言
	userInfo, ok := user.(*models.User)
	if !ok {
		// 如果类型断言失败，返回 400 错误
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}
	// 校验会话是否存在
	var conversation models.Conversation
	err := config.DB.Where("conversation_id = ?", conversationID).First(&conversation).Error
	if err != nil {
		log.Println("Conversation not found:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// 确保用户是该会话的成员
	if conversation.ParticipantA != fmt.Sprint(userInfo.ID) && conversation.ParticipantB != fmt.Sprint(userInfo.ID) && conversation.GroupID == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not part of this conversation"})
		return
	}

	// 获取会话的所有消息
	var messages []models.Message
	err = config.DB.
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC"). // 按时间排序（最早的在前）
		Find(&messages).Error
	if err != nil {
		log.Println("Error fetching messages:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	// 返回消息列表
	utils.RespondSuccess(c, messages, nil)
}
