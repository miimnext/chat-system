package services

import (
	"chat-system/config"
	"chat-system/models"
	"encoding/json"
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
	clients    map[string][]*Client // 存储多个客户端连接，按 user_id 分组
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.Mutex
}

var Manager = &WSManager{
	clients:    make(map[string][]*Client),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	broadcast:  make(chan []byte),
}

type Message struct {
	Type           string `json:"type"` // "broadcast" 或 "private"
	To             string `json:"to,omitempty"`
	Content        string `json:"content"`
	ConversationID string `json:"conversation_id"`
}

func (m *WSManager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.ID] = append(m.clients[client.ID], client)
			m.mu.Unlock()
			fmt.Println("New client registered:", client.ID)
			go client.StartHeartbeat()

		case client := <-m.unregister:
			m.mu.Lock()
			if clients, ok := m.clients[client.ID]; ok {
				for i, c := range clients {
					if c == client {
						m.clients[client.ID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(m.clients[client.ID]) == 0 {
					client.closeOnce.Do(func() {
						defer func() {
							if r := recover(); r != nil {
								fmt.Println("Attempted to close an already closed channel")
							}
						}()
						close(client.Send)
					})
				}
				fmt.Println("Client unregistered:", client.ID)
			}
			m.mu.Unlock()

		case msg := <-m.broadcast:
			m.mu.Lock()
			for _, clients := range m.clients {
				for _, client := range clients {
					select {
					case client.Send <- msg:
					default:
						fmt.Println("Skipping client:", client.ID)
					}
				}
			}
			m.mu.Unlock()
		}
	}
}

func (c *Client) ReadMessages() {
	defer func() {
		Manager.unregister <- c
		c.CloseSendChannel()

		c.mu.Lock()
		if c.Conn != nil {
			fmt.Println("Closing connection for client:", c.ID)
			c.Conn.Close() // ✅ 确保不会误关
			c.Conn = nil
		}
		c.mu.Unlock()
	}()
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		if string(msg) == "pong" {
			c.mu.Lock()
			c.LastPing = time.Now()
			c.mu.Unlock()
			continue
		}
		var data Message

		if err := json.Unmarshal(msg, &data); err != nil {
			fmt.Println("Invalid message format:", string(msg))
			continue
		}

		if data.Type == "private" {
			ReceiverID := getReceiverID(c.ID, data.ConversationID)
			message := models.Message{
				ConversationID: data.ConversationID,
				SenderID:       c.ID,
				ReceiverID:     ReceiverID,
				Content:        data.Content,
				MessageType:    data.Type,
				Status:         "sent",
			}
			err := Manager.SendMessage(data.ConversationID, ReceiverID, message)
			if err != nil {
				fmt.Println("Failed to send private message:", err)
			}

			if err := config.DB.Create(&message).Error; err != nil {
				fmt.Println("Failed to send message")
				return
			}
		} else {
			Manager.broadcast <- msg
		}
	}
}

func (c *Client) WriteMessages() {
	defer func() {
		c.closeOnce.Do(func() {
			c.mu.Lock()
			if c.Send != nil {
				close(c.Send)
				c.Send = nil
			}
			c.mu.Unlock()
		})
	}()
	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}

func (m *WSManager) SendMessage(ConversationId, clientID string, message models.Message) error {
	m.mu.Lock()
	clients, exists := m.clients[clientID]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	for _, client := range clients {
		client.mu.Lock()
		msg, err := json.Marshal(message)
		if err != nil {
			fmt.Println("Error marshaling message:", err)
			client.mu.Unlock()
			return err
		}
		err = client.Conn.WriteMessage(websocket.TextMessage, msg)
		client.mu.Unlock()

		if err != nil {
			fmt.Println("Error sending message to", clientID, ":", err)
			return err
		}

		fmt.Println("Message sent to", clientID, ":", string(msg))
	}

	return nil
}

func (c *Client) StartHeartbeat() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		// 如果连接已关闭，直接退出
		if c.Conn == nil {
			fmt.Println("Connection already closed, exiting heartbeat:", c.ID)
			c.mu.Unlock()
			return
		}

		// 尝试发送 Ping
		err := c.Conn.WriteMessage(websocket.TextMessage, []byte("ping"))
		if err != nil {
			c.mu.Unlock()
			fmt.Println("Ping failed, closing connection:", c.ID, err)
			c.closeClient()
			return
		}

		// 打印日志，每次发送心跳
		fmt.Printf("Sent ping to client %s at %v\n", c.ID, time.Now())

		// 检测最近的 Pong 是否超时
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
		fmt.Println("Closing client connection:", c.ID)
		c.Conn.Close() // ✅ 确保连接被关闭
		Manager.unregister <- c
	})
}

func (c *Client) CloseSendChannel() {
	c.closeOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Attempted to close an already closed channel")
			}
		}()
		close(c.Send)
	})
}
