package aicli

import "context"

// ContextMessage 上下文消息
type ContextMessage struct {
	Role    string `json:"role"` // system, user, assistant, etc.
	Content string `json:"content"`
}

// StreamReply 结构体用于表示流式响应
type StreamReply struct {
	Content chan string
	Err     error // 如果为 nil 表示成功响应
}
type Assistanter interface {
	Send(ctx context.Context, prompt string, files ...string) (string, error)
	RefreshContext()
	ListModelNames(ctx context.Context) ([]string, error)
	SendStream(ctx context.Context, prompt string, files ...string) *StreamReply
}
