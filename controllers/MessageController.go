package controllers

import (
	"chat-system/config"
	"chat-system/models"
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
func GetMessages(c *gin.Context) {
	var input struct {
		ConversationID string `json:"conversation_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var messages []models.Message
	if err := config.DB.Where("conversation_id = ?", input.ConversationID).Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}
