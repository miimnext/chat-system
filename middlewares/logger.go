package middlewares

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 中间件记录每个请求的日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求开始时间
		start := time.Now()

		// 获取客户端的 IP 地址
		clientIP := c.ClientIP()

		// 读取请求体内容
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
		}
		// 将读取的请求体内容恢复到原始的 c.Request.Body，以便后续处理
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

		// 处理请求前的日志
		log.Printf("Started %s %s from %s with body: %s", c.Request.Method, c.Request.URL.Path, clientIP, string(requestBody))

		// 处理请求
		c.Next()

		// 计算请求的响应时间
		duration := time.Since(start)

		// 处理请求后的日志
		log.Printf("Completed %s %s with status %d from %s in %v", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), clientIP, duration)
	}
}
