package controllers

import (
	"chat-system/config"
	"chat-system/models"
	"chat-system/services"
	"chat-system/utils"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserInfoResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// 用户注册
func Register(c *gin.Context) {
	var userInput struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := config.DB.Where("username = ?", userInput.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 创建新用户
	newUser := models.User{
		Username:  userInput.Username,
		Password:  string(hashedPassword),
		LastLogin: nil, // 让它默认 NULL
	}

	// 插入数据库
	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	// 生成 JWT Token
	token, err := services.GenerateToken(newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	utils.RespondSuccess(c, gin.H{"token": token}, nil)
}

// 用户登录
func Login(c *gin.Context) {
	var loginInput struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&loginInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 查找用户
	var user models.User
	if err := config.DB.Where("username = ?", loginInput.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginInput.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now // 这里用指针
	config.DB.Save(&user)

	// 生成 JWT Token
	token, err := services.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	utils.RespondSuccess(c, gin.H{"token": token}, nil)
}

func GetUserInfo(c *gin.Context) {
	// 从上下文中获取用户信息
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
	data := UserInfoResponse{
		ID:       userInfo.ID,
		Username: userInfo.Username,
	}
	utils.RespondSuccess(c, data, nil)
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
		c.JSON(http.StatusOK, gin.H{"conversation_id": existingConversation.ConversationID})
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
		"code":            200,
	}
	log.Println(data, 123123)
	// 返回新创建的会话 ID
	utils.RespondSuccess(c, data, nil)

}
