package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 发送消息响应结构体
type SendMessageResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	Status    string `json:"status"`
}

// SendMessageHandler 发送消息处理函数
func SendMessageHandler(c *gin.Context) {
	var req struct {
		SessionID string   `json:"session_id"`
		Prompt    string   `json:"prompt"`
		Files     []string `json:"files"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 如果没有提供 session_id，自动创建一个新会话
	if req.SessionID == "" {
		// 创建会话
		sessionID, err := manager.CreateSession()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session: " + err.Error()})
			return
		}

		req.SessionID = sessionID
	}

	// 发送消息
	response, err := manager.SendMessage(req.SessionID, req.Prompt, req.Files...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message: " + err.Error()})
		return
	}

	// 返回响应
	resp := SendMessageResponse{
		Message:   response,
		SessionID: req.SessionID,
		Status:    "success",
	}

	c.JSON(http.StatusOK, resp)
}
