package controllers

import (
	"chat-system/services"

	"github.com/gin-gonic/gin"
)

func WSController(ctx *gin.Context) {
	services.HandleWebSocket(ctx)
}
