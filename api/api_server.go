package api

import (
	"bubble/internal/config"
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

// 发送消息请求结构体
type SendMessageRequest struct {
	SessionID string   `json:"session_id"`
	Prompt    string   `json:"prompt"`
	Files     []string `json:"files"`
}

// 发送消息响应结构体
type SendMessageResponse struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
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
var manager *aicli.SessionManager

func init() {
	manager = aicli.GetSessionManager()
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

// 创建会话处理函数
func CreateSessionHandler(c *gin.Context) {
	// 选择供应商和模型
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
		aicli.WithInitialRole(aicli.GenericRoleDescCN),
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

// 发送消息处理函数
func SendMessageHandler(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 如果没有提供 session_id，自动创建一个新会话
	if req.SessionID == "" {
		// 选择供应商和模型
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
			aicli.WithInitialRole(aicli.GenericRoleDescCN),
		)
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

	// 获取会话信息
	provider, model, err := manager.GetSessionInfo(req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session info: " + err.Error()})
		return
	}

	// 返回响应
	resp := SendMessageResponse{
		Message:   response,
		SessionID: req.SessionID,
		Provider:  provider,
		Model:     model,
		Status:    "success",
	}

	c.JSON(http.StatusOK, resp)
}

// 获取上下文处理函数
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

// 删除会话处理函数
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

// listSessionsHandler 列出所有会话
func ListSessionsHandler(c *gin.Context) {
	sessions := manager.ListSessions()

	resp := ListSessionsResponse{
		Sessions: sessions,
		Status:   "success",
	}

	c.JSON(http.StatusOK, resp)
}
