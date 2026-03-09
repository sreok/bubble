package session

import (
	"bubble/pkg/aicli"
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// Session 表示一个会话，包含一个客户端实例
type Session struct {
	ID     string
	Client *aicli.Client
}

// SessionManager 管理多个会话
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var (
	// 全局会话管理器实例
	manager *SessionManager
	once    sync.Once
)

// GetSessionManager 获取会话管理器实例
func GetSessionManager() *SessionManager {
	once.Do(func() {
		manager = &SessionManager{
			sessions: make(map[string]*Session),
		}
	})
	return manager
}

// CreateSession 创建一个新的会话
func (sm *SessionManager) CreateSession() (string, error) {

	// 从配置中获取默认提供商
	provider, model, err := aicli.SelectAvailableProviderAndModel()
	if err != nil {
		return "", err
	}

	apiKey := provider.APIKey
	baseURL := provider.BaseURL
	modelName := model.ID

	// 创建客户端选项
	var opts []aicli.ClientOption
	if baseURL != "" {
		opts = append(opts, aicli.WithBaseURL(baseURL))
	}
	if modelName != "" {
		opts = append(opts, aicli.WithModel(modelName))
	}

	// 创建客户端
	client, err := aicli.NewClient(apiKey, opts...)
	if err != nil {
		return "", err
	}

	// 生成会话 ID
	sessionID := uuid.New().String()

	// 添加到会话管理器
	sm.mu.Lock()
	sm.sessions[sessionID] = &Session{
		ID:     sessionID,
		Client: client,
	}
	sm.mu.Unlock()

	return sessionID, nil
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	sm.mu.RLock()
	session, ok := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !ok {
		return nil, errors.New("session not found")
	}

	return session, nil
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID cannot be empty")
	}

	sm.mu.Lock()
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()

	return nil
}

// ListSessions 列出所有会话
func (sm *SessionManager) ListSessions() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessionIDs []string
	for sessionID := range sm.sessions {
		sessionIDs = append(sessionIDs, sessionID)
	}

	return sessionIDs
}

// SendMessage 发送消息到指定会话
func (sm *SessionManager) SendMessage(sessionID, prompt string, files ...string) (string, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	// 增强发送消息，可以判断是否需要搜索相关信息
	return session.Client.EnhancedSend(context.Background(), prompt, files...)
}

// SendMessageStream 发送流式消息到指定会话
func (sm *SessionManager) SendMessageStream(sessionID, prompt string, files ...string) (*aicli.StreamReply, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session.Client.SendStream(context.Background(), prompt, files...), nil
}

// ListContextMessages 列出会话的上下文消息
func (sm *SessionManager) ListContextMessages(sessionID string) ([]*aicli.ContextMessage, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return session.Client.ListContextMessages(), nil
}

// RefreshContext 刷新会话的上下文
func (sm *SessionManager) RefreshContext(sessionID string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.Client.RefreshContext()
	return nil
}
