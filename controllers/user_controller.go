package controllers

import (
	"chat-system/config"
	"chat-system/models"
	"chat-system/services"
	"chat-system/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
		utils.RespondFailed(c, "Username already exists")
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondFailed(c, "Failed to hash password")
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
		utils.RespondFailed(c, "Failed to create user")
		return
	}
	// 生成 JWT Token
	token, err := services.GenerateToken(newUser)
	if err != nil {
		utils.RespondFailed(c, "Failed to generate token")
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
		utils.RespondFailed(c, "Invalid username or password")
		return
	}
	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginInput.Password)); err != nil {
		utils.RespondFailed(c, "Invalid username or password")
		return
	}
	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now // 这里用指针
	config.DB.Save(&user)

	// 生成 JWT Token
	token, err := services.GenerateToken(user)
	if err != nil {
		utils.RespondFailed(c, "Failed to generate token")
		return
	}

	utils.RespondSuccess(c, gin.H{"token": token}, nil)
}

func GetUserInfo(c *gin.Context) {
	// 从上下文中获取用户信息
	user, exists := c.Get("user")
	if !exists {
		// 如果用户信息不存在，返回 404 错误
		utils.RespondFailed(c, "User not found")
		return
	}

	// 假设 user 是 *models.User 类型，进行类型断言
	userInfo, ok := user.(*models.User)
	if !ok {
		// 如果类型断言失败，返回 400 错误
		utils.RespondFailed(c, "Invalid user data")
		return
	}

	data := UserInfoResponse{
		ID:       userInfo.ID,
		Username: userInfo.Username,
	}
	utils.RespondSuccess(c, data, nil)
}
