package main

import (
	"bubble/api"
	"bubble/internal/config"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Printf("Failed to load config: %v\n", err)
		return
	}

	// 启动配置文件监视器，每5秒检查一次
	go config.StartConfigWatcher(5 * time.Second)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, world!",
		})
	})

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/models", api.GetModels)
		apiGroup.POST("/chat", api.Chat)
	}
	r.Run(":8080")
}
