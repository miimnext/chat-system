package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client 代表一个 WebSocket 客户端
type Client struct {
	Conn         *websocket.Conn // WebSocket 连接
	Send         chan []byte     // 用于发送消息的通道
	LastActive   time.Time       // 上次活动时间
	UserID       string          // 用户 ID
	ConnectionID string          // 连接 ID
}

// Message 定义统一的消息格式
type Message struct {
	Type    string      `json:"type"`    // 消息类型（例如 "text", "notification"）
	Content interface{} `json:"content"` // 消息内容
	Sender  string      `json:"sender"`  // 发送者 ID
	Target  string      `json:"target"`  // 接收者 ID（用于私聊）
}

// WebSocketManager 管理所有 WebSocket 连接
type WebSocketManager struct {
	Connections map[string]*Client // 存储连接及客户端映射
	Broadcast   chan []byte        // 广播消息通道
	Register    chan *Client       // 注册新连接
	Unregister  chan *Client       // 注销连接
	Mutex       sync.RWMutex       // 读写锁
}

var (
	wsManager *WebSocketManager
	once      sync.Once
)

// GetWSManager 获取 WebSocket 管理器实例（单例模式）
func GetWSManager() *WebSocketManager {
	once.Do(func() {
		wsManager = &WebSocketManager{
			Connections: make(map[string]*Client),
			Broadcast:   make(chan []byte),
			Register:    make(chan *Client),
			Unregister:  make(chan *Client),
		}
		go wsManager.Start()
	})
	return wsManager
}

// Start 运行 WebSocket 管理器
func (manager *WebSocketManager) Start() {
	ticker := time.NewTicker(30 * time.Second) // 每 30 秒检查一次连接活跃状态
	defer ticker.Stop()

	for {
		select {
		case client := <-manager.Register:
			manager.Mutex.Lock()
			manager.Connections[client.UserID] = client
			manager.Mutex.Unlock()
			log.Printf("🔵 New client connected: %s", client.UserID)

		case client := <-manager.Unregister:
			manager.Mutex.Lock()
			if _, ok := manager.Connections[client.UserID]; ok {
				delete(manager.Connections, client.UserID)
				close(client.Send)
				log.Printf("🔴 Client disconnected: %s", client.UserID)
			}
			manager.Mutex.Unlock()

		case message := <-manager.Broadcast:
			manager.Mutex.Lock()
			for _, client := range manager.Connections {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(manager.Connections, client.UserID)
					log.Printf("⚠️ Failed to send message to client: %s", client.UserID)
				}
			}
			manager.Mutex.Unlock()

		case <-ticker.C:
			manager.Mutex.Lock()
			for userID, client := range manager.Connections {
				if time.Since(client.LastActive) > 60*time.Second { // 超过 60 秒未活跃
					client.Conn.Close()
					close(client.Send)
					delete(manager.Connections, userID)
					log.Printf("🔴 Inactive client removed: %s", userID)
				}
			}
			manager.Mutex.Unlock()
		}
	}
}

// AddConnection 添加新连接
func (manager *WebSocketManager) AddConnection(userID, connectionID string, conn *websocket.Conn) {
	manager.Mutex.Lock()
	defer manager.Mutex.Unlock()

	client := &Client{
		Conn:         conn,
		Send:         make(chan []byte, 256), // 缓冲通道，避免阻塞
		LastActive:   time.Now(),
		UserID:       userID,
		ConnectionID: connectionID,
	}
	manager.Connections[userID] = client
	log.Printf("🟢 User %s connected with connection ID %s", userID, connectionID)

	go manager.handleMessages(client)
}

// RemoveConnection 移除连接
func (manager *WebSocketManager) RemoveConnection(userID string) {
	manager.Mutex.Lock()
	defer manager.Mutex.Unlock()

	if client, exists := manager.Connections[userID]; exists {
		client.Conn.Close()
		close(client.Send)
		delete(manager.Connections, userID)
		log.Printf("🔴 Connection removed for user: %s", userID)
	}
}

// SendMessage 发送私聊消息
func (manager *WebSocketManager) SendMessage(receiverID string, msg Message) error {
	manager.Mutex.RLock()
	defer manager.Mutex.RUnlock()

	client, exists := manager.Connections[receiverID]
	if !exists {
		log.Printf("⚠️ No active connection found for user %s", receiverID)
		return fmt.Errorf("client not found")
	}

	message, err := json.Marshal(msg)
	if err != nil {
		log.Printf("⚠️ Failed to marshal message: %v", err)
		return err
	}

	client.Send <- message
	log.Printf("📩 Private message sent to user %s: %s", receiverID, string(message))
	return nil
}

// handleMessages 处理消息发送
func (manager *WebSocketManager) handleMessages(client *Client) {
	defer manager.RemoveConnection(client.UserID)

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				return // 通道关闭，退出协程
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("⚠️ Error sending message to %s: %v", client.UserID, err)
				return
			}
		}
	}
}
