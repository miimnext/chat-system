package controllers

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

// WebSocket 连接管理器
type WebSocketManager struct {
	Connections  map[string]*Client // 存储连接及客户端映射
	Broadcast    chan []byte        // 广播消息通道
	Register     chan *Client       // 注册新连接
	Unregister   chan *Client       // 注销连接
	PingInterval time.Duration      // 心跳检测间隔
	PongTimeout  time.Duration      // Pong 超时
	Mutex        sync.RWMutex       // 读写锁
}

// 创建 WebSocket 管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		Connections: make(map[string]*Client),
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
	}
}

// 运行 WebSocket 管理器（管理连接的注册/注销）
func (manager *WebSocketManager) Start() {
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
		}
	}
}

// 发送私聊消息给指定用户
func (manager *WebSocketManager) SendMessage(receiverID string, msgData interface{}) error {
	manager.Mutex.RLock()
	defer manager.Mutex.RUnlock()

	client, exists := manager.Connections[receiverID]
	if !exists {
		log.Printf("⚠️ No active connection found for user %s", receiverID)
		return fmt.Errorf("client not found")
	}

	message, err := json.Marshal(msgData)
	if err != nil {
		log.Printf("⚠️ Failed to marshal message: %v", err)
		return err
	}

	// 异步发送消息
	go func() {
		if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("⚠️ Failed to send message to user %s: %v", receiverID, err)
			manager.RemoveConnection(receiverID) // Remove invalid connection
		} else {
			log.Printf("📩 Private message sent to user %s: %s", receiverID, string(message))
		}
	}()

	return nil
}

// 移除连接
func (manager *WebSocketManager) RemoveConnection(userID string) {
	manager.Mutex.Lock()
	defer manager.Mutex.Unlock()

	if client, exists := manager.Connections[userID]; exists {
		client.Conn.Close()
		delete(manager.Connections, userID)
		log.Printf("🔴 Connection removed for user: %s", userID)
	}
}

// 添加连接，关联用户ID和连接ID
func (manager *WebSocketManager) AddConnection(userID, connectionID string, conn *websocket.Conn) {
	manager.Mutex.Lock()
	defer manager.Mutex.Unlock()

	// 添加连接到管理器
	client := &Client{
		Conn:         conn,
		Send:         make(chan []byte),
		LastActive:   time.Now(),
		UserID:       userID,
		ConnectionID: connectionID,
	}
	manager.Connections[userID] = client
	log.Printf("🟢 User %s connected with connection ID %s", userID, connectionID)

	// 开始处理发送通道
	go manager.handleMessages(client)
}

// 处理消息发送
func (manager *WebSocketManager) handleMessages(client *Client) {
	for msg := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("⚠️ Error sending message to %s: %v", client.UserID, err)
			manager.RemoveConnection(client.UserID)
			break
		}
	}
}

// 全局 WebSocket 管理器实例
var (
	wsManager *WebSocketManager
	once      sync.Once
)

// 获取 WebSocket 管理器实例
func GetWSManager(pingInterval, pongTimeout time.Duration) *WebSocketManager {
	once.Do(func() {
		wsManager = NewWebSocketManager()
		go wsManager.Start()
	})
	return wsManager
}
