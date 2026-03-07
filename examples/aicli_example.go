package main

import (
	"bubble/internal/config"
	"bubble/pkg/aicli"
	"context"
	"fmt"
	"log"
)

// 加载配置文件
func loadConfig() error {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v\n", err)
	}

	// 检查 Providers 是否为空
	if len(config.AppConfig.Models.Providers) == 0 {
		log.Fatalf("No providers found in config\n")
	}

	return nil
}

// 选择供应商和模型
func selectProviderAndModel() (*config.ProviderConfig, *config.ModelConfig, error) {
	// 选择供应商
	selectedProvider, err := selectProvider()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select provider: %w", err)
	}

	// 选择模型
	modelConfig, err := selectModel(selectedProvider)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select model: %w", err)
	}

	return selectedProvider, modelConfig, nil
}

// 选择供应商
func selectProvider() (*config.ProviderConfig, error) {
	// 优先使用主供应商
	defaultProviderName := config.AppConfig.Models.Default
	if defaultProviderName != "" {
		// 查找主供应商
		for i := range config.AppConfig.Models.Providers {
			provider := &config.AppConfig.Models.Providers[i]
			if provider.Name == defaultProviderName && provider.Enable && provider.APIKey != "" {
				return provider, nil
			}
		}

		// 默认供应商不存在或未启用，如果启用了故障转移，尝试其他供应商
		if config.AppConfig.Models.Failover {
			for i := range config.AppConfig.Models.Providers {
				provider := &config.AppConfig.Models.Providers[i]
				if provider.Enable && provider.APIKey != "" && provider.Name != defaultProviderName {
					log.Printf("Default provider %s not available, switching to %s\n", defaultProviderName, provider.Name)
					return provider, nil
				}
			}
		}
	} else {
		// 没有指定默认供应商，使用第一个启用的供应商
		for i := range config.AppConfig.Models.Providers {
			provider := &config.AppConfig.Models.Providers[i]
			if provider.Enable && provider.APIKey != "" {
				log.Printf("Using provider %s\n", provider.Name)
				return provider, nil
			}
		}
	}

	return nil, fmt.Errorf("no available providers found in config")
}

// 选择模型
func selectModel(provider *config.ProviderConfig) (*config.ModelConfig, error) {
	// 优先使用主模型
	defaultModelName := provider.Default
	if defaultModelName != "" {
		// 查找主模型
		for i := range provider.Models {
			model := &provider.Models[i]
			if model.ID == defaultModelName {
				return model, nil
			}
		}

		// 默认模型不存在，如果启用了故障转移，尝试其他模型
		if provider.Failover && len(provider.Models) > 0 {
			firstModel := &provider.Models[0]
			log.Printf("Default model %s not available, switching to %s\n", defaultModelName, firstModel.ID)
			return firstModel, nil
		}
	} else if len(provider.Models) > 0 {
		// 没有指定主模型，使用第一个模型
		return &provider.Models[0], nil
	}

	return nil, fmt.Errorf("no models found for provider %s", provider.Name)
}

// 创建 AI 客户端
func createClient(provider *config.ProviderConfig, model *config.ModelConfig) (*aicli.Client, error) {
	var client *aicli.Client
	var err error

	if provider.BaseURL != "" {
		// 使用自定义 base URL
		client, err = aicli.NewClient(
			provider.APIKey,
			aicli.WithBaseURL(provider.BaseURL),
			aicli.WithModel(model.ID),
			aicli.WithEnableContext(),
			aicli.WithInitialRole(aicli.GenericRoleDescCN),
		)
	} else {
		// 使用默认 base URL
		client, err = aicli.NewClient(
			provider.APIKey,
			aicli.WithModel(model.ID),
			aicli.WithEnableContext(),
			aicli.WithInitialRole(aicli.GenericRoleDescCN),
		)
	}

	if err != nil {
		log.Fatalf("Failed to create AI client: %v\n", err)
	}

	return client, nil
}

// 发送请求
func sendRequest(client *aicli.Client, prompt string) (string, error) {
	ctx := context.Background()
	response, err := client.Send(ctx, prompt)
	if err != nil {
		log.Fatalf("Failed to send request: %v\n", err)
	}

	log.Printf("Response: %s\n", response)
	return response, nil
}

// 发送流式请求
func sendStreamRequest(client *aicli.Client, prompt string) error {
	ctx := context.Background()
	stream := client.SendStream(ctx, prompt)

	for chunk := range stream.Content {
		log.Print(chunk)
	}

	if stream.Err != nil {
		log.Fatalf("Error in streaming: %v\n", stream.Err)
		return fmt.Errorf("error in streaming: %w", stream.Err)
	}

	return nil
}

// 故障转移尝试其他供应商
func failoverToOtherProviders(prompt string, failedProvider *config.ProviderConfig) error {
	log.Printf("Request failed with provider %s\n", failedProvider.Name)
	log.Println("Attempting failover to other providers...")

	// 尝试其他供应商
	for i := range config.AppConfig.Models.Providers {
		otherProvider := &config.AppConfig.Models.Providers[i]
		if otherProvider.Name != failedProvider.Name && otherProvider.Enable && otherProvider.APIKey != "" {
			log.Printf("Switching to provider: %s\n", otherProvider.Name)

			// 选择模型
			otherModel, err := selectModel(otherProvider)
			if err != nil {
				log.Printf("Failed to select model for provider %s: %v\n", otherProvider.Name, err)
				continue
			}

			// 创建客户端
			otherClient, err := createClient(otherProvider, otherModel)
			if err != nil {
				log.Printf("Failed to create client for provider %s: %v\n", otherProvider.Name, err)
				continue
			}

			// 尝试发送请求
			response, err := sendRequest(otherClient, prompt)
			if err != nil {
				log.Printf("Request failed with provider %s: %v\n", otherProvider.Name, err)
				continue
			}

			// 请求成功
			log.Printf("Provider: %s\nModel: %s\nResponse: %s\n", otherProvider.Name, otherModel.ID, response)

			// 发送流式请求
			log.Printf("\nProvider: %s\nModel: %s\nStreaming response: ", otherProvider.Name, otherModel.ID)
			if err := sendStreamRequest(otherClient, prompt); err != nil {
				log.Printf("Error in streaming: %v\n", err)
			}

			log.Println()
			return nil
		}
	}

	// 所有供应商都失败
	log.Println("All providers failed")
	return fmt.Errorf("all providers failed")
}

// 尝试发送请求，如果失败且启用了故障转移，则尝试其他供应商
func tryRequestWithFailover(prompt string) error {
	// 选择供应商和模型
	provider, model, err := selectProviderAndModel()
	if err != nil {
		return err
	}

	// 创建 AI 客户端
	client, err := createClient(provider, model)
	if err != nil {
		return err
	}

	// 发送请求
	response, err := sendRequest(client, prompt)
	if err != nil {
		// 如果请求失败且启用了故障转移，尝试其他供应商
		if config.AppConfig.Models.Failover {
			if err := failoverToOtherProviders(prompt, provider); err != nil {
				log.Fatalf("Request failed: %v\n", err)
			}
			return nil
		}

		return err
	}

	log.Printf("Provider: %s\nModel: %s\nResponse: %s\n", provider.Name, model.ID, response)

	// 发送流式请求
	log.Printf("\nProvider: %s\nModel: %s\nStreaming response: ", provider.Name, model.ID)
	if err := sendStreamRequest(client, prompt); err != nil {
		log.Printf("Error in streaming: %v\n", err)
		return err
	}

	log.Println()
	return nil
}

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	// 发送请求，支持故障转移
	prompt := "你是什么模型？"
	if err := tryRequestWithFailover(prompt); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
}
