package api

import (
	"bubble/internal/i18n"
	"bubble/pkg/aicli"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChatResponse 聊天响应结构体
type ChatResponse struct {
	Message  string `json:"message"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Status   string `json:"status"`
}

// Chat 聊天接口
func Chat(c *gin.Context) {
	// 从请求体中获取用户输入
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ChatResponse{
			Message: i18n.T("invalid_request_payload"),
			Status:  "error",
		})
		return
	}

	// 获取 AI 响应
	message, provider, model, err := aicli.GetAIResponse(req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ChatResponse{
			Message: i18n.T("failed_to_get_message") + ": " + err.Error(),
			Status:  "error",
		})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, ChatResponse{
		Message:  message,
		Provider: provider,
		Model:    model,
		Status:   "success",
	})
}
