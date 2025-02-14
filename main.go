package main

import (
	"chat-system/config"
	"chat-system/controllers"
	"chat-system/models" // 导入 controllers 包
	"chat-system/routes"
	"log"
)

func main() {
	// 初始化数据库
	config.InitDB()
	// 自动迁移
	models.Migrate()
	go controllers.GetWSManager().Start()
	// 注册路由
	r := routes.RegisterRoutes()

	// 启动服务
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
