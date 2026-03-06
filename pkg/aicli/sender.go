// Package aicli 提供发送请求到助手的接口
package aicli

import "context"

// Assistanter 助手接口，用于发送请求到助手
type Assistanter interface {
	Send(ctx context.Context, prompt string, files ...string) (string, error)
	SendStream(ctx context.Context, prompt string, files ...string) *StreamReply
	RefreshContext()
	ListModelNames(ctx context.Context) ([]string, error)
}

// 定义 StreamReply 结构体
// StreamReply 结构体用于表示流式响应
type StreamReply struct {
	Content chan string
	Err     error // 如果为 nil 表示成功响应
}

// 定义默认的角色描述
const (
	GenericRoleDescEN = `You are a helpful assistant, able to answer user questions in a clear and friendly manner.`
	GenericRoleDescCN = `你是一个有帮助的助手，能够以清晰、友好的方式回答用户的问题。`

	GopherRoleDescEN = `You are an experienced Go development engineer, specializing in designing and implementing efficient and scalable business logic using the Go programming language.  
You are well-versed in Go’s concurrency model, performance optimization, code structure design, and system integration.  
You can help write high-quality Go code, solve complex business problems, and provide best practice recommendations.  
You excel at using Go’s standard library or third-party libraries to implement business logic code.`
	GopherRoleDescCN = `你是一位经验丰富的 Go 开发工程师，专注于使用 Go 语言设计和实现高效、可扩展的业务逻辑。
你熟悉 Go 的并发模型、性能优化、代码结构设计以及与其他系统的集成。
你可以帮助编写高质量的 Go 代码，解决复杂的业务问题，并提供最佳实践建议。
你擅长使用 Go 的标准库或第三方库来实现业务逻辑代码。`
)
