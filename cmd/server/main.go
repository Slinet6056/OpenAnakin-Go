package main

import (
	"github.com/Slinet6056/OpenAnakin-Go/internal/client"
	"github.com/Slinet6056/OpenAnakin-Go/internal/config"
	"github.com/Slinet6056/OpenAnakin-Go/internal/handler"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建Anakin客户端
	anakinClient := client.NewAnakinClient(config.AppConfig.Models)

	// 创建OpenAI兼容处理器
	openAIHandler := handler.NewOpenAIHandler(anakinClient)

	// 设置路由
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		v1.POST("/chat/completions", openAIHandler.ChatCompletions)
	}

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
