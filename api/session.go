package api

import (
	"bubble/internal/config"
	"bubble/internal/tools/session"
	"bubble/pkg/aicli"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 创建会话响应结构体
type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

// 获取会话上下文响应结构体
type GetContextResponse struct {
	ContextMessages []*aicli.ContextMessage `json:"context_messages"`
	Count           int                     `json:"count"`
}

// 删除会话响应结构体
type DeleteSessionResponse struct {
	Status    string `json:"status"`
	SessionID string `json:"session_id,omitempty"`
}

// 列出会话响应结构体
type ListSessionsResponse struct {
	Sessions []string `json:"sessions"`
	Status   string   `json:"status"`
}

// 全局会话管理器
var manager *session.SessionManager

func init() {
	manager = session.GetSessionManager()
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

// CreateSessionHandler 创建会话处理函数
func CreateSessionHandler(c *gin.Context) {
	// 选择提供商和模型
	provider, model, err := aicli.SelectAvailableProviderAndModel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to select provider and model: " + err.Error()})
		return
	}

	// 创建会话
	sessionID, err := manager.CreateSession(
		provider.APIKey,
		provider.Name,
		model.ID,
		aicli.WithModel(model.ID),
		aicli.WithBaseURL(provider.BaseURL),
		aicli.WithEnableContext(),
		aicli.WithInitialRole(config.GenericRoleDescCN),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session: " + err.Error()})
		return
	}

	// 返回响应
	resp := CreateSessionResponse{
		SessionID: sessionID,
		Status:    "success",
	}

	c.JSON(http.StatusOK, resp)
}

// GetContextHandler 获取上下文处理函数
func GetContextHandler(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	// 获取上下文
	contextMessages, err := manager.ListContextMessages(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get context: " + err.Error()})
		return
	}

	// 返回响应
	resp := GetContextResponse{
		ContextMessages: contextMessages,
		Count:           len(contextMessages),
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteSessionHandler 删除会话处理函数
func DeleteSessionHandler(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	// 删除会话
	if err := manager.DeleteSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session: " + err.Error()})
		return
	}

	// 返回响应
	resp := DeleteSessionResponse{
		Status:    "success",
		SessionID: sessionID,
	}

	c.JSON(http.StatusOK, resp)
}

// ListSessionsHandler 列出所有会话
func ListSessionsHandler(c *gin.Context) {
	sessions := manager.ListSessions()

	resp := ListSessionsResponse{
		Sessions: sessions,
		Status:   "success",
	}

	c.JSON(http.StatusOK, resp)
}
