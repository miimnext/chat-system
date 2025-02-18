package controllers

import (
	"chat-system/config"
	"chat-system/models"
	"chat-system/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetConversation(c *gin.Context) {
	// 从上下文中获取用户信息
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 类型断言
	userInfo, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	var conversations []models.Conversation
	err := config.DB.
		Preload("ParticipantAUser").
		Preload("ParticipantBUser").
		Where("participant_a = ? OR participant_b = ? OR group_id IS NOT NULL", userInfo.ID, userInfo.ID).
		Find(&conversations).Error
	if err != nil {
		log.Println("Error fetching conversations:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	// 处理返回的数据，仅返回对方用户信息
	formattedConversations := make([]map[string]interface{}, 0)

	for _, conv := range conversations {
		// 处理私聊
		if conv.GroupID == "" {
			var otherUser *models.User

			if fmt.Sprintf("%d", userInfo.ID) == conv.ParticipantA {
				otherUser = &conv.ParticipantBUser
			} else {
				otherUser = &conv.ParticipantAUser
			}
			formattedConversations = append(formattedConversations, map[string]interface{}{
				"conversation_id": conv.ConversationID,
				"type":            "private",
				"participant": map[string]interface{}{
					"user_id":    otherUser.ID,
					"username":   otherUser.Username,
					"email":      otherUser.Email,
					"avatar":     otherUser.AvatarURL,
					"last_login": otherUser.LastLogin,
				},
			})
		} else {
			// 处理群聊
			formattedConversations = append(formattedConversations, map[string]interface{}{
				"conversation_id": conv.ConversationID,
				"type":            "group",
				"group_id":        conv.GroupID,
			})
		}
	}

	// 返回处理后的数据
	utils.RespondSuccess(c, formattedConversations, nil)
}

// CreateConversationHandler 创建会话（使用POST请求）
func CreateConversationHandler(c *gin.Context) {
	// 获取请求的 JSON 数据
	var requestData struct {
		UserID     string `json:"user_id"`     // 当前用户ID
		ReceiverID string `json:"receiver_id"` // 目标用户ID
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := requestData.UserID
	receiverID := requestData.ReceiverID

	// 校验 receiverID 是否存在
	var receiverUser models.User
	err := config.DB.Where("id = ?", receiverID).First(&receiverUser).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Receiver does not exist", "code": "404"})
		return
	}

	// 校验 receiverID 是否与当前用户的 ID 匹配
	if userID == receiverID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create a conversation with yourself"})
		return
	}

	// 查找是否已有私聊会话
	var existingConversation models.Conversation
	err = config.DB.Where("(participant_a = ? AND participant_b = ?) OR (participant_a = ? AND participant_b = ?)", userID, receiverID, receiverID, userID).First(&existingConversation).Error
	if err == nil {
		// 如果找到了已有的会话，返回已存在的会话 ID
		data := map[string]interface{}{
			"conversation_id": existingConversation.ConversationID,
		}
		// 返回新创建的会话 ID
		utils.RespondSuccess(c, data, nil)
		return
	}

	// 如果没有找到已有会话，创建一个新的会话
	conversationID := uuid.New().String()
	newConversation := models.Conversation{
		ConversationID: conversationID,
		Type:           "private",
		ParticipantA:   userID,
		ParticipantB:   receiverID,
	}

	// 保存新会话到数据库
	if err := config.DB.Create(&newConversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		log.Println("Error creating conversation:", err)
		return
	}
	data := map[string]interface{}{
		"conversation_id": conversationID,
		"code":            "200",
	}
	// 返回新创建的会话 ID
	utils.RespondSuccess(c, data, nil)

}

// GetConversationByID 根据会话 ID 获取会话信息
func GetConversationByID(c *gin.Context) {
	conversationID := c.Param("conversation_id") // 从 URL 参数获取 conversation_id
	userID := c.GetString("user_id")             // 获取当前用户ID

	var conversation models.Conversation
	err := config.DB.
		Preload("ParticipantAUser").
		Preload("ParticipantBUser").
		Where("conversation_id = ?", conversationID).
		First(&conversation).Error
	if err != nil {
		log.Println("Error fetching conversation:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// 处理返回的数据
	responseData := map[string]interface{}{
		"conversation_id": conversation.ConversationID,
		"type":            conversation.Type,
	}

	if conversation.GroupID == "" {
		// 私聊: 获取对方的信息
		var otherUser *models.User
		if conversation.ParticipantA == userID {
			otherUser = &conversation.ParticipantBUser
		} else {
			otherUser = &conversation.ParticipantAUser
		}

		responseData["participant"] = map[string]interface{}{
			"user_id":    otherUser.ID,
			"username":   otherUser.Username,
			"email":      otherUser.Email,
			"avatar":     otherUser.AvatarURL,
			"last_login": otherUser.LastLogin,
		}
	} else {
		// 群聊
		responseData["group_id"] = conversation.GroupID
	}

	utils.RespondSuccess(c, responseData, nil)
}
