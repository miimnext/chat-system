package services

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleWebSocket(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	client := &Client{
		Conn:     conn,
		Send:     make(chan []byte),
		ID:       ctx.Query("user_id"),
		LastPing: time.Now(), // 初始化心跳时间
	}

	Manager.register <- client

	go client.ReadMessages()
	go client.WriteMessages()
}
