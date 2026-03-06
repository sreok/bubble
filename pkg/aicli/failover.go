package aicli

import (
	"bubble/internal/config"
	"context"
	"fmt"
	"log"
)

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
	primaryProviderName := config.AppConfig.Models.Primary
	if primaryProviderName != "" {
		// 查找主供应商
		for i := range config.AppConfig.Models.Providers {
			provider := &config.AppConfig.Models.Providers[i]
			if provider.Name == primaryProviderName && provider.Enable && provider.APIKey != "" {
				return provider, nil
			}
		}

		// 主供应商不存在或未启用，如果启用了故障转移，尝试其他供应商
		if config.AppConfig.Models.Failover {
			for i := range config.AppConfig.Models.Providers {
				provider := &config.AppConfig.Models.Providers[i]
				if provider.Enable && provider.APIKey != "" && provider.Name != primaryProviderName {
					log.Printf("Primary provider %s not available, switching to %s\n", primaryProviderName, provider.Name)
					return provider, nil
				}
			}
		}
	} else {
		// 没有指定主供应商，使用第一个启用的供应商
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
	primaryModelName := provider.Primary
	if primaryModelName != "" {
		// 查找主模型
		for i := range provider.Models {
			model := &provider.Models[i]
			if model.ID == primaryModelName {
				return model, nil
			}
		}

		// 主模型不存在，如果启用了故障转移，尝试其他模型
		if provider.Failover && len(provider.Models) > 0 {
			firstModel := &provider.Models[0]
			log.Printf("Primary model %s not available, switching to %s\n", primaryModelName, firstModel.ID)
			return firstModel, nil
		}
	} else if len(provider.Models) > 0 {
		// 没有指定主模型，使用第一个模型
		return &provider.Models[0], nil
	}

	return nil, fmt.Errorf("no models found for provider %s", provider.Name)
}

// 故障转移尝试其他供应商
func failoverToOtherProviders(prompt string, failedProvider *config.ProviderConfig) (string, error) {
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
			response, err := otherClient.Send(context.Background(), prompt)
			if err != nil {
				log.Printf("Request failed with provider %s: %v\n", otherProvider.Name, err)
				continue
			}

			// 请求成功
			log.Printf("Provider: %s\nModel: %s\nResponse: %s\n", otherProvider.Name, otherModel.ID, response)
			return response, nil
		}
	}

	// 所有供应商都失败
	log.Println("All providers failed")
	return "", fmt.Errorf("all providers failed")
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
			_, err := failoverToOtherProviders(prompt, provider)
			if err != nil {
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
