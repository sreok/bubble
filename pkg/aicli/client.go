package aicli

import (
	"bubble/internal/config"
	"bubble/internal/tools/web_search"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)

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

// Assistant 接口定义了与助手交互的方法
type Assistanter interface {
	Send(ctx context.Context, prompt string, files ...string) (string, error)
	RefreshContext()
	ListModelNames(ctx context.Context) ([]string, error)
	SendStream(ctx context.Context, prompt string, files ...string) *StreamReply
}

// 定义 Client 结构体
type Client struct {
	apiKey           string
	maxTokens        int
	ModelName        string
	Cli              *openai.Client
	temperature      float32
	roleDesc         string
	enableContext    bool
	contextMessages  []openai.ChatCompletionMessage
	baseURL          string
	maxContextSize   int
	maxContextTokens int
}

// ClientOption 客户端选项
type ClientOption func(*Client)

const (
	defaultModel         = openai.GPT4o // default
	defaultMaxTokens     = 8192
	defaultTemperature   = 0.7
	defaultContextSize   = 10
	defaultContextTokens = 8192
)

func defaultClientOptions() *Client {
	return &Client{
		enableContext:    false, // default is false
		maxTokens:        defaultMaxTokens,
		temperature:      defaultTemperature,
		maxContextSize:   defaultContextSize, // 默认最大上下文消息数
		maxContextTokens: defaultMaxTokens,   // 默认最大上下文 token 数
		ModelName:        defaultModel,
	}
}

func (c *Client) apply(opts ...ClientOption) {
	for _, opt := range opts {
		opt(c)
	}
}

// NewClient creates a new chat client.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {

	c := defaultClientOptions()
	c.apply(opts...)

	if c.roleDesc != "" {
		c.contextMessages = append(c.contextMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: c.roleDesc,
		})
	}

	c.apiKey = apiKey

	config := openai.DefaultConfig(apiKey)

	c.Cli = openai.NewClientWithConfig(config)

	return c, nil
}

// Send 发送提示词到聊天模型并返回响应
func (c *Client) Send(ctx context.Context, prompt string, files ...string) (string, error) {
	if prompt == "" {
		return "", errors.New("prompt cannot be empty")
	}

	messages, err := c.setMessages(ctx, prompt, files...)
	if err != nil {
		return "", err
	}

	reply, err := c.Cli.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:               c.ModelName,
			Messages:            messages,
			Temperature:         c.temperature,
			MaxCompletionTokens: c.maxTokens,
			MaxTokens:           c.maxTokens, // Deprecated
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	replyContent := ""
	for _, choice := range reply.Choices {
		replyContent += choice.Message.Content
	}
	if replyContent == "" {
		return "", errors.New("reply content is empty")
	}
	c.appendAssistantContext(prompt, replyContent)

	return replyContent, nil
}

// EnhancedSend 增强 AI 响应，添加搜索能力
func (c *Client) EnhancedSend(ctx context.Context, prompt string, files ...string) (string, error) {
	// 检查是否需要搜索
	if web_search.ContainsSearchKeywords(prompt) && web_search.IsTavilyEnabled() {
		// 从配置中获取 Tavily API key
		tavilyAPIKey := config.AppConfig.Tools.WebSearch.Tavily.APIKey
		if tavilyAPIKey == "" {
			log.Println("Tavily API key not configured")
			// 没有配置 API key，直接使用普通发送
			return c.Send(ctx, prompt, files...)
		}

		// 创建 Tavily 客户端
		tavilyClient := web_search.NewTavilyClient(tavilyAPIKey)

		// 搜索相关信息
		searchResult, err := tavilyClient.Search(ctx, prompt)
		if err != nil {
			log.Printf("Failed to search with Tavily: %v", err)
			// 搜索失败，直接使用普通发送
			return c.Send(ctx, prompt, files...)
		}

		// 增强 prompt
		enhancedPrompt := fmt.Sprintf(`
你是一个有帮助的助手，能够以清晰、友好的方式回答用户的问题。

基于以下搜索结果，请回答用户的问题：
%s

用户的问题：%s
`, searchResult, prompt)

		// 使用增强后的 prompt 发送
		return c.Send(ctx, enhancedPrompt, files...)
	}

	// 不需要搜索，直接使用普通发送
	return c.Send(ctx, prompt, files...)
}

// SendStream 发送提示词到聊天模型并返回响应流
func (c *Client) SendStream(ctx context.Context, prompt string, files ...string) *StreamReply {
	response := &StreamReply{Content: make(chan string), Err: error(nil)}

	go func() {
		defer func() { close(response.Content) }()

		messages, err := c.setMessages(ctx, prompt, files...)
		if err != nil {
			response.Err = err
			return
		}

		req := openai.ChatCompletionRequest{
			Model:               c.ModelName,
			Messages:            messages,
			Stream:              true,
			Temperature:         c.temperature,
			MaxCompletionTokens: c.maxTokens,
			MaxTokens:           c.maxTokens, // Deprecated
		}
		stream, err := c.Cli.CreateChatCompletionStream(ctx, req)
		if err != nil {
			response.Err = err
			return
		}
		defer func() {
			_ = stream.Close() //nolint
		}()

		var replyContent string
		defer func() {
			if response.Err == nil && replyContent != "" {
				c.appendAssistantContext(prompt, replyContent)
			}
		}()

		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				response.Err = err
				return
			}

			for _, choice := range resp.Choices {
				select {
				case <-ctx.Done():
					response.Err = ctx.Err()
					return
				case response.Content <- choice.Delta.Content:
					replyContent += choice.Delta.Content
				}
			}
		}
	}()

	return response
}

// ListModelNames 列出所有可用的模型名称
func (c *Client) ListModelNames(ctx context.Context) ([]string, error) {
	list, err := c.Cli.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	var modelNames []string
	for _, model := range list.Models {
		modelNames = append(modelNames, model.ID)
	}

	return modelNames, nil
}

// ListContextMessages 列出助手上下文消息
func (c *Client) ListContextMessages() []*ContextMessage {
	contextMessages := make([]*ContextMessage, 0, len(c.contextMessages))
	for _, message := range c.contextMessages {
		contextMessages = append(contextMessages, &ContextMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	return contextMessages
}

// RefreshContext 刷新助手上下文
func (c *Client) RefreshContext() {
	if len(c.contextMessages) > 0 {
		c.contextMessages = []openai.ChatCompletionMessage{}
	}
}

// appendAssistantContext 将助手上下文添加到上下文消息中
func (c *Client) appendAssistantContext(prompt string, replyContent string) {
	if c.enableContext && replyContent != "" {
		// 添加新的上下文消息
		newMessages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: replyContent,
			},
		}
		c.contextMessages = append(c.contextMessages, newMessages...)

		// 限制上下文消息数量
		c.trimContext()
	}
}

// trimContext 修剪上下文消息，确保不超过限制
func (c *Client) trimContext() {
	// 1. 限制消息数量
	if c.maxContextSize > 0 && len(c.contextMessages) > c.maxContextSize {
		// 保留系统消息，只修剪用户和助手消息
		var systemMessages []openai.ChatCompletionMessage
		var userAssistantMessages []openai.ChatCompletionMessage

		for _, msg := range c.contextMessages {
			if msg.Role == openai.ChatMessageRoleSystem {
				systemMessages = append(systemMessages, msg)
			} else {
				userAssistantMessages = append(userAssistantMessages, msg)
			}
		}

		// 计算需要保留的用户助手消息数量
		retainCount := c.maxContextSize - len(systemMessages)
		if retainCount > 0 && len(userAssistantMessages) > retainCount {
			// 保留最新的消息
			userAssistantMessages = userAssistantMessages[len(userAssistantMessages)-retainCount:]
		}

		// 重新组合消息
		c.contextMessages = append(systemMessages, userAssistantMessages...)
	}

	// 2. 限制 token 数量（简单实现，实际应该计算 token 数）
	// 这里可以使用 tiktoken 库来计算 token 数
}

// setMessages 设置聊天模型的消息
func (c *Client) setMessages(ctx context.Context, prompt string, files ...string) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage

	// history context
	if len(c.contextMessages) > 0 {
		messages = append(messages, c.contextMessages...)
	}

	// file message
	if len(files) > 0 {
		fileIDs, err := c.uploadFiles(ctx, files)
		if err != nil {
			return nil, err
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("Please refer to the content of the following document ID: %v", fileIDs),
		})
	}

	// user message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	// initial role
	if c.roleDesc != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: c.roleDesc,
		})
	}

	return messages, nil
}

// UploadFiles 上传文件到聊天模型
func (c *Client) uploadFiles(ctx context.Context, files []string) ([]string, error) {
	fileIDs := make([]string, 0, len(files))
	for _, filePath := range files {
		_, name := filepath.Split(filePath)
		fileResp, err := c.Cli.CreateFile(ctx, openai.FileRequest{
			FileName: name,
			FilePath: filePath,
			Purpose:  "assistants", // for assistants
		})
		if err != nil {
			return nil, err
		}
		fileIDs = append(fileIDs, fileResp.ID)
	}
	return fileIDs, nil
}
