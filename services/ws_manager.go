package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pingInterval = 10 * time.Second // 发送 Ping 的间隔
	pongTimeout  = 15 * time.Second // 超过 15 秒未收到 Pong 断开连接
)

type Client struct {
	Conn      *websocket.Conn
	Send      chan []byte
	ID        string
	LastPing  time.Time
	mu        sync.Mutex
	closeOnce sync.Once
}

type WSManager struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.Mutex
}

var Manager = &WSManager{
	clients:    make(map[string]*Client),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	broadcast:  make(chan []byte),
}

func (m *WSManager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.ID] = client
			m.mu.Unlock()
			fmt.Println("New client registered:", client.ID)
			go client.StartHeartbeat() // 启动心跳检测

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client.ID]; ok {
				client.closeOnce.Do(func() {
					close(client.Send) // **确保 Send 只关闭一次**
				})
				delete(m.clients, client.ID)
				fmt.Println("Client unregistered:", client.ID)
			}
			m.mu.Unlock()

		case msg := <-m.broadcast:
			m.mu.Lock()
			for _, client := range m.clients {
				select {
				case client.Send <- msg:
				default:
					// 如果客户端无法接收，跳过，避免阻塞
					fmt.Println("Skipping client:", client.ID)
				}
			}
			m.mu.Unlock()
		}
	}
}

func (c *Client) ReadMessages() {
	defer func() {
		Manager.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetPongHandler(func(appData string) error {
		c.mu.Lock()
		c.LastPing = time.Now() // ✅ 更新心跳时间
		c.mu.Unlock()
		return nil
	})

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		Manager.broadcast <- msg
	}
}

func (c *Client) WriteMessages() {
	defer func() {
		c.closeOnce.Do(func() {
			close(c.Send) // 确保只关闭一次
		})
	}()

	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.PingMessage, msg)
		if err != nil {
			break
		}
	}
}

func (m *WSManager) SendMessage(clientID string, messageType int, message []byte) error {
	m.mu.Lock()
	client, exists := m.clients[clientID]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	err := client.Conn.WriteMessage(messageType, message)
	if err != nil {
		fmt.Println("Error sending message to", clientID, ":", err)
		return err
	}

	fmt.Println("Message sent to", clientID, ":", string(message))
	return nil
}

func (c *Client) StartHeartbeat() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		// 1️⃣ **如果连接已关闭，直接退出**
		if c.Conn == nil {
			fmt.Println("Connection already closed, exiting heartbeat:", c.ID)
			c.mu.Unlock()
			return
		}

		// 2️⃣ **尝试发送 Ping**
		err := c.Conn.WriteMessage(websocket.TextMessage, []byte("ping"))
		if err != nil {
			c.mu.Unlock()
			fmt.Println("Ping failed, closing connection:", c.ID)
			c.closeClient()
			return
		}

		// 3️⃣ **检测最近的 Pong 是否超时**
		if time.Since(c.LastPing) > pongTimeout {
			c.mu.Unlock()
			fmt.Println("Client timeout, closing connection:", c.ID)
			c.closeClient()
			return
		}

		c.mu.Unlock()
	}
}

func (c *Client) closeClient() {
	c.closeOnce.Do(func() {
		Manager.unregister <- c
	})
}
