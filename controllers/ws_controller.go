package controllers

import (
	"log"
	"net/http"

	"chat-system/services" // 替换为你的项目路径

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有跨域请求
	},
}

// HandleWebSocket 处理 WebSocket 连接
func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
		return
	}
	defer conn.Close()

	userID := c.Query("user_id")
	if userID == "" {
		log.Println("User ID is required")
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "user_id is required"}`))
		conn.Close() // 确保连接正确关闭
		return
	}

	connectionID := uuid.New().String()
	manager := services.GetWSManager()
	manager.AddConnection(userID, connectionID, conn)
}
