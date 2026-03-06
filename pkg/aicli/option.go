package aicli

import (
	openai "github.com/sashabaranov/go-openai"
)

// openai 模型名称
const (
	ModelGPT3Dot5Turbo = openai.GPT3Dot5Turbo
	ModelGPT4          = openai.GPT4
	ModelGPT4Turbo     = openai.GPT4Turbo
	ModelGPT4o         = openai.GPT4o // default
	ModelGPT4oMini     = openai.GPT4oMini
	ModelO1Mini        = openai.O1Mini
	ModelO1Preview     = openai.O1Preview

	DefaultModel     = ModelGPT4o
	defaultMaxTokens = 8192
)

// ClientOption 客户端选项
type ClientOption func(*Client)

func defaultClientOptions() *Client {
	return &Client{
		enableContext: false, // default is false
		maxTokens:     defaultMaxTokens,
		temperature:   0.0,
	}
}

func (c *Client) apply(opts ...ClientOption) {
	for _, opt := range opts {
		opt(c)
	}
}

// WithMaxTokens 设置最大令牌数
func WithMaxTokens(maxTokens int) ClientOption {
	return func(c *Client) {
		if maxTokens < 1000 {
			c.maxTokens = defaultMaxTokens
		}
		c.maxTokens = maxTokens
	}
}

// WithModel 设置模型名称
func WithModel(name string) ClientOption {
	return func(c *Client) {
		c.ModelName = name
	}
}

// WithTemperature 设置温度
func WithTemperature(temperature float32) ClientOption {
	return func(c *Client) {
		c.temperature = temperature
	}
}

// WithInitialRole 设置初始角色类型
func WithInitialRole(roleDesc string) ClientOption {
	return func(c *Client) {
		c.roleDesc = roleDesc
	}
}

// WithEnableContext 设置助手上下文
func WithEnableContext() ClientOption {
	return func(c *Client) {
		c.enableContext = true
	}
}

// WithBaseURL 设置自定义 base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// ContextMessage 上下文消息
type ContextMessage struct {
	Role    string `json:"role"` // system, user, assistant, etc.
	Content string `json:"content"`
}

// WithInitialContextMessages 设置初始上下文消息
func WithInitialContextMessages(messages ...*ContextMessage) ClientOption {
	return func(c *Client) {
		if len(messages) > 0 {
			c.enableContext = true
			for _, message := range messages {
				c.contextMessages = append(c.contextMessages, openai.ChatCompletionMessage{
					Role:    message.Role,
					Content: message.Content,
				})
			}
		}
	}
}
