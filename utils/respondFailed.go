package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func RespondFailed(c *gin.Context, message string) {
	// 构建基础的响应数据
	response := ResponseData{
		Code:    404,
		Message: message,
	}

	// 返回响应
	c.JSON(http.StatusOK, response)
}
