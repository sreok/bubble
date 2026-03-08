package main

import (
	"bubble/api"
	"bubble/internal/config"
	"bubble/internal/i18n"
	"bubble/pkg/aicli"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Printf(i18n.T("config_load_failed")+" %v\n", err)
		return
	}

	// 语言已经在LoadConfig中设置，不需要重复设置

	// 启动配置文件监视器，每5秒检查一次
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			err, modified := config.CheckAndReload()
			if err != nil {
				log.Printf(i18n.T("config_reload_failed"), err)
			} else if modified {
				// 配置文件重载成功，重置不可用列表
				aicli.RefreshAvailableLists()
			}
		}
	}()

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
		apiGroup.POST("/messages", api.SendMessageHandler)
		// 设置路由
		sessionGroup := apiGroup.Group("/sessions")
		{
			sessionGroup.POST("", api.CreateSessionHandler)
			sessionGroup.POST("/list", api.ListSessionsHandler)
			sessionGroup.GET("/context", api.GetContextHandler)
			sessionGroup.DELETE("", api.DeleteSessionHandler)
		}
	}

	r.Run(":8080")
}
