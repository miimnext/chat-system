package controllers

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket 管理器
type WSManager struct {
	connections     map[string]*websocket.Conn // 存储连接的映射
	connectionIDMap map[string]string          // 存储wsConnectionID与userID的映射
	groupMembers    map[string][]string        // 存储GroupID与成员wsConnectionID的映射
	mu              sync.RWMutex               // 读写锁，确保并发安全
}

// 创建一个新的 WebSocket 管理器
func NewWSManager() *WSManager {
	return &WSManager{
		connections:     make(map[string]*websocket.Conn),
		connectionIDMap: make(map[string]string),   // 初始化映射
		groupMembers:    make(map[string][]string), // 初始化群组成员映射
	}
}

// 启动 WebSocket 管理器
func (m *WSManager) Start() {
	log.Println("WebSocket Manager started")
}

// 添加连接，关联用户ID和wsConnectionID
func (m *WSManager) AddConnection(wsConnectionID, userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[wsConnectionID] = conn
	m.connectionIDMap[wsConnectionID] = userID
	log.Printf("User %s connected with connection ID %s", userID, wsConnectionID)
}

// 移除连接
func (m *WSManager) RemoveConnection(wsConnectionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if conn, ok := m.connections[wsConnectionID]; ok {
		conn.Close() // 关闭连接
		delete(m.connections, wsConnectionID)
		delete(m.connectionIDMap, wsConnectionID) // 移除wsConnectionID的映射
		log.Printf("Connection ID %s disconnected", wsConnectionID)
	}
}

// 获取连接
func (m *WSManager) GetConnection(wsConnectionID string) *websocket.Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connections[wsConnectionID]
}

// 获取用户ID通过 wsConnectionID
func (m *WSManager) GetUserID(wsConnectionID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connectionIDMap[wsConnectionID]
}

// 添加群组成员
func (m *WSManager) AddGroupMember(groupID, wsConnectionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.groupMembers[groupID] = append(m.groupMembers[groupID], wsConnectionID)
	log.Printf("Connection ID %s added to group %s", wsConnectionID, groupID)
}

// 发送私聊消息给指定用户
func (m *WSManager) SendMessage(receiverID string, message []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 获取接收者的 WebSocket 连接
	conn := m.GetConnection(receiverID)
	if conn == nil {
		log.Printf("No active connection found for user %s", receiverID)
		return nil
	}

	// 检查连接是否有效
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Printf("Failed to send message to user %s: %v", receiverID, err)
		m.RemoveConnection(receiverID) // 移除无效连接
		return err
	}

	log.Printf("Private message sent to user %s: %s", receiverID, string(message))
	return nil
}

// 广播消息给指定群组
func (m *WSManager) BroadcastGroupMessage(groupID string, message []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 获取群组中的所有成员
	groupConnections, exists := m.groupMembers[groupID]
	if !exists {
		log.Printf("Group %s not found", groupID)
		return nil
	}

	// 向所有成员发送消息
	for _, wsConnectionID := range groupConnections {
		conn := m.GetConnection(wsConnectionID)
		if conn == nil {
			log.Printf("No active connection found for connection ID %s", wsConnectionID)
			continue
		}

		// 检查连接是否有效
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send message to connection ID %s: %v", wsConnectionID, err)
			m.RemoveConnection(wsConnectionID) // 移除无效连接
			continue
		}

		log.Printf("Message sent to group %s, connection ID %s", groupID, wsConnectionID)
	}

	return nil
}

// 广播消息给所有连接
func (m *WSManager) Broadcast(message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for wsConnectionID, conn := range m.connections {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Failed to send broadcast message to connection ID %s: %v", wsConnectionID, err)
			m.RemoveConnection(wsConnectionID) // 移除无效连接
		} else {
			log.Printf("Broadcast message sent to connection ID %s: %s", wsConnectionID, string(message))
		}
	}
}

// 全局 WebSocket 管理器实例
var (
	wsManager *WSManager
	once      sync.Once
)

// 获取 WebSocket 管理器实例
func GetWSManager() *WSManager {
	once.Do(func() {
		wsManager = NewWSManager()
	})
	return wsManager
}
