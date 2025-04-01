package routes

import (
	"chat-system/controllers"
	"chat-system/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes() *gin.Engine {

	r := gin.Default()
	// 配置跨域中间件
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},                                       // 允许的域名，可以是前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // 允许的 HTTP 方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // 允许的请求头
		AllowCredentials: true,                                                // 是否允许发送 cookies
	}

	// 使用 CORS 中间件
	r.Use(cors.New(corsConfig))
	r.GET("/ws", controllers.WSController)
	protected := r.Group("/api")

	// 注册路由

	protected.POST("/register", controllers.Register) // 绑定注册接口
	protected.POST("/login", controllers.Login)       // 绑定登录接口

	{
		protected.Use(middlewares.TokenAuthMiddleware())
		protected.GET("/userinfo", controllers.GetUserInfo)
		protected.GET("/conversation", controllers.GetConversation)
		protected.POST("/createConversation", controllers.CreateConversationHandler)
		protected.GET("/conversation/:conversation_id", controllers.GetMessagesByConversationID)
	}

	return r
}
