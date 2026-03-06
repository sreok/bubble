package api

import (
	"bubble/internal/config"
	"bubble/pkg/aicli"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Chat 聊天接口
func Chat(c *gin.Context) {
	// 从请求体中获取用户输入
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// 检查配置文件是否变化并重新加载
	if err := config.CheckAndReload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reload config",
		})
		return
	}

	// 检查 Providers 是否为空
	if len(config.AppConfig.Models.Providers) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No providers found in config",
		})
		return
	}

	// 获取 AI 响应
	content, provider, model, err := aicli.GetAIResponse(req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get AI response",
		})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"message":  content,
		"provider": provider,
		"model":    model,
	})
}
