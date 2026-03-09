package aicli

import (
	openai "github.com/sashabaranov/go-openai"
)

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

// WithMaxContextSize 设置最大上下文消息数
func WithMaxContextSize(size int) ClientOption {
	return func(c *Client) {
		if size > 0 {
			c.maxContextSize = size
		}
	}
}

// WithMaxContextTokens 设置最大上下文 token 数
func WithMaxContextTokens(tokens int) ClientOption {
	return func(c *Client) {
		if tokens > 0 {
			c.maxContextTokens = tokens
		}
	}
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
