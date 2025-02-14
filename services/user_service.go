package services

import (
	"chat-system/config"
	"chat-system/models"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET")) // 使用环境变量获取密钥

// Claims 是自定义的 JWT Claims 结构体
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// CreateUser 用于创建新用户
func CreateUser(user models.User) (models.User, error) {
	// Check if a user with the same username already exists
	var existingUser models.User
	if err := config.DB.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return models.User{}, errors.New("username already exists")
	}
	// Insert the new user into the database
	if err := config.DB.Create(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// CheckPassword 验证密码（实际项目中应该加密存储并验证）
func CheckPassword(storedPassword, providedPassword string) bool {
	// 简单示例，实际应该使用密码哈希进行比较（如 bcrypt）
	return storedPassword == providedPassword
}

// GenerateToken 生成 JWT Token
func GenerateToken(user models.User) (string, error) {
	// 设置 Token 过期时间为 24 小时
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(), // 使用 Unix 时间戳表示过期时间
			Issuer:    "my-gin-project",      // 可以设置为应用名称
		},
	}

	// 创建 JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名 Token
	return token.SignedString(jwtKey)
}

// ValidateToken 验证 JWT Token
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证 Token 的签名方法是否为 HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}

func GetAllUser() ([]models.User, error) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		return nil, errors.New("failed to fetch users")
	}
	if len(users) == 0 {
		return nil, errors.New("no users found")
	}
	return users, nil
}

// GetCurrentUser 根据提供的 token 获取当前用户信息
func GetCurrentUser(tokenString string) (*models.User, error) {
	// 假设你的 JWT 密钥存在 config 中
	secretKey := jwtKey

	// 解析 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 确保 token 签名方法正确
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	// 从 token 中获取 Claims（即用户信息）
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// 假设你的 token 中包含用户 ID 字段
	username := claims["username"].(string)

	// 根据用户 ID 从数据库中查找用户
	var user models.User
	if err := config.DB.First(&user, "username = ?", username).Error; err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	return &user, nil
}
