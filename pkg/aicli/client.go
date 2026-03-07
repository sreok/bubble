package aicli

import (
	"bubble/internal/config"
	"bubble/internal/i18n"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)

// 定义 Client 结构体
type Client struct {
	apiKey          string
	maxTokens       int
	ModelName       string
	Cli             *openai.Client
	temperature     float32
	roleDesc        string
	enableContext   bool
	contextMessages []openai.ChatCompletionMessage
	baseURL         string
}

// 加载配置文件
func loadConfig() error {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf(i18n.T("failed_to_load_config")+" %v\n", err)
	}
	return nil
}

// NewClient creates a new chat client.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {

	if apiKey == "" {
		return nil, errors.New("API key cannot be empty")
	}

	c := defaultClientOptions()
	c.apply(opts...)

	if c.ModelName == "" {
		c.ModelName = DefaultModel
	}

	if c.roleDesc != "" {
		c.contextMessages = append(c.contextMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: c.roleDesc,
		})
	}

	c.apiKey = apiKey

	// 根据是否设置了自定义 base URL 来创建客户端
	if c.baseURL != "" {
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = c.baseURL
		c.Cli = openai.NewClientWithConfig(config)
	} else {
		c.Cli = openai.NewClient(apiKey)
	}

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

// ModifyInitialRole 修改初始角色描述
func (c *Client) ModifyInitialRole(roleDesc string) {
	if roleDesc == "" {
		return
	}
	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: roleDesc,
	}

	if len(c.contextMessages) == 0 {
		c.contextMessages = []openai.ChatCompletionMessage{message}
	} else {
		if c.roleDesc == c.contextMessages[0].Content {
			c.contextMessages[0].Content = roleDesc
		} else {
			c.contextMessages = append([]openai.ChatCompletionMessage{message}, c.contextMessages...)
		}
	}
}

// appendAssistantContext 将助手上下文添加到上下文消息中
func (c *Client) appendAssistantContext(prompt string, replyContent string) {
	if c.enableContext && replyContent != "" {
		c.contextMessages = append(c.contextMessages, []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: replyContent,
			},
		}...)
	}
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

// 创建 AI 客户端
// func createClient(provider *config.ProviderConfig, model *config.ModelConfig) (*Client, error) {
// 	var client *Client
// 	var err error

// 	if provider.BaseURL != "" {
// 		// 使用自定义 base URL
// 		client, err = NewClient(
// 			provider.APIKey,
// 			WithBaseURL(provider.BaseURL),
// 			WithModel(model.ID),
// 			WithEnableContext(),
// 			WithInitialRole(GenericRoleDescCN),
// 		)
// 	} else {
// 		// 使用默认 base URL
// 		client, err = NewClient(
// 			provider.APIKey,
// 			WithModel(model.ID),
// 			WithEnableContext(),
// 			WithInitialRole(GenericRoleDescCN),
// 		)
// 	}

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create AI client: %w", err)
// 	}

// 	return client, nil
// }

// 发送请求
func sendRequest(client *Client, prompt string) (string, error) {
	ctx := context.Background()
	response, err := client.Send(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	log.Printf("Response: %s\n", response)
	return response, nil
}

// 发送流式请求
func sendStreamRequest(client *Client, prompt string) error {
	ctx := context.Background()
	stream := client.SendStream(ctx, prompt)

	for chunk := range stream.Content {
		log.Print(chunk)
	}

	if stream.Err != nil {
		return fmt.Errorf("error in streaming: %w", stream.Err)
	}

	return nil
}

// GetAIResponse 获取 AI 响应
func GetAIResponse(prompt string) (string, string, string, error) {
	// 检查 Providers 是否为空
	if len(config.AppConfig.Models.Providers) == 0 {
		log.Fatalf("%s\n", i18n.T("no_providers_found"))
	}

	// 选择供应商和模型
	provider, model, err := selectProviderAndModel()
	if err != nil {
		return "", "", "", err
	}

	// 创建 AI 客户端
	client, err := NewClient(
		provider.APIKey,
		WithModel(model.ID),
		WithBaseURL(provider.BaseURL),
		WithEnableContext(),
		WithInitialRole(GenericRoleDescCN),
	)
	if err != nil {
		return "", "", "", err
	}

	// 发送请求
	message, err := client.Send(context.Background(), prompt)
	if err != nil {
		log.Printf(i18n.T("request_failed_with_provider")+" %s \n"+i18n.T("Error")+": %v", provider.Name, err)
		// 如果请求失败且启用了故障转移，尝试其他提供商
		if config.AppConfig.Models.Failover {
			otherProvider, otherModel, content, err := providersFailover(prompt, provider)
			if err != nil || content == "" {
				return "", "", "", err
			} else {
				provider = otherProvider
				model = otherModel
				message = content
			}
		} else {
			return "", "", "", err
		}
	}

	log.Printf(i18n.T("response_from_provider")+" %s \n"+i18n.T("Model")+": %s \n"+i18n.T("Response")+": %s", provider.Name, model.ID, message)
	return message, provider.Name, model.ID, nil
}
