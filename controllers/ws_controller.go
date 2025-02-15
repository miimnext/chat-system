package controllers

import (
	"chat-system/config"
	"chat-system/models"
	"chat-system/services"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket 连接处理
// WebSocket 连接处理
func Connect(c *gin.Context) {
	// 升级 HTTP 请求为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
		return
	}
	defer conn.Close()

	// 获取用户ID
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// 生成 WebSocket 连接ID
	wsConnectionID := uuid.New().String()

	// 将 WebSocket 连接添加到 WebSocket 管理器
	GetWSManager().AddConnection(wsConnectionID, userIDStr, conn)

	// 启动心跳 Goroutine
	go sendHeartbeat(conn)

	// 获取会话列表并发送给客户端
	go sendConversations(conn, userIDStr)

	// 监听并处理消息
	receiveMessages(conn, userIDStr, wsConnectionID)
}

// 获取并发送会话列表
func sendConversations(conn *websocket.Conn, userID string) {
	// 查询用户的会话列表（包括私聊和群聊），并预加载参与者信息
	var conversations []models.Conversation
	err := config.DB.
		Preload("ParticipantAUser").
		Preload("ParticipantBUser").
		Where("participant_a = ? OR participant_b = ? OR group_id IS NOT NULL", userID, userID).
		Find(&conversations).
		Error
	if err != nil {
		log.Println("Error fetching conversations:", err)
		return
	}

	// 格式化会话列表并发送
	conversationsData := make([]map[string]interface{}, len(conversations))
	for i, conv := range conversations {
		conversationsData[i] = map[string]interface{}{
			"conversation_id": conv.ConversationID,
			"type":            conv.Type,
			"participant_a":   conv.ParticipantAUser.Username, // 使用预加载的参与者A信息
			"participant_b":   conv.ParticipantBUser.Username, // 使用预加载的参与者B信息
			"group_id":        conv.GroupID,
		}
	}

	// 发送会话列表到客户端
	responseMessage := map[string]interface{}{
		"Type":          "conversations",
		"Conversations": conversationsData,
	}

	responseBytes, err := json.Marshal(responseMessage)
	if err != nil {
		log.Println("Failed to marshal conversations response:", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
		log.Println("Failed to send conversations:", err)
	}
}

// 发送心跳消息
func sendHeartbeat(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送一次心跳
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
			log.Println("Failed to send ping message:", err)
			return
		}
	}
}

// 接收并处理消息
func receiveMessages(conn *websocket.Conn, userID string, wsConnectionID string) {
	log.Println("Listening for messages...")
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Println("Unexpected WebSocket close:", err)
			} else {
				log.Println("Error while reading message:", err)
			}
			break
		}

		// 解析消息
		var msgData models.Message
		if err := json.Unmarshal(message, &msgData); err != nil {
			log.Println("Failed to unmarshal message:", err)
			continue
		}

		// 设置消息的发送者、连接ID
		msgData.SenderID = userID
		msgData.MessageID = generateMessageID() // 生成唯一Message ID
		msgData.Status = "sent"
		msgData.CreatedAt = time.Now()

		// 如果会话 ID 为空，则创建一个新的会话
		if msgData.ConversationID == "" {
			if msgData.Type == "private" && msgData.ReceiverID != "" {
				// 创建私聊会话
				conversationID, err := services.GetOrCreateConversation(userID, msgData.ReceiverID)
				if err != nil {
					log.Println("Failed to get or create conversation:", err)
					continue
				}
				msgData.ConversationID = conversationID
			} else if msgData.Type == "group" && msgData.GroupID != "" {
				// 群聊使用 groupID 作为会话 ID
				msgData.ConversationID = msgData.GroupID
			} else {
				log.Println("Invalid message: missing conversation ID, receiver ID, or group ID")
				continue
			}
		}

		// 判断消息类型并发送消息
		switch msgData.Type {
		case "private":
			// 发送私聊消息
			if err := GetWSManager().SendMessage(msgData.ReceiverID, msgData); err != nil {
				log.Println("Failed to send private message:", err)
			}
		case "group":
			// 发送群聊消息
			if err := GetWSManager().BroadcastGroupMessage(msgData.GroupID, []byte(msgData.Content)); err != nil {
				log.Println("Failed to send group message:", err)
			}
		default:
			log.Println("Message is neither private nor group message:", msgData.Type)
		}

		// 保存消息到数据库
		if err := config.DB.Create(&msgData).Error; err != nil {
			log.Println("Failed to save message:", err)
			continue
		}

		// 发送确认消息
		sendConfirmation(conn, msgData.MessageID, wsConnectionID)
	}
}

// 发送确认消息
func sendConfirmation(conn *websocket.Conn, messageID string, wsConnectionID string) {
	responseMessage := map[string]string{
		"Type":           "confirmation",
		"MessageID":      messageID,      // 返回Message ID
		"WSConnectionID": wsConnectionID, // 返回 WebSocket 连接ID
	}

	responseBytes, err := json.Marshal(responseMessage)
	if err != nil {
		log.Println("Failed to marshal response message:", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
		log.Println("Failed to send response:", err)
	}
}

// 生成唯一Message ID
func generateMessageID() string {
	return uuid.New().String()
}
