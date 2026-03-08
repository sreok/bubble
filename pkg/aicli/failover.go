package aicli

import (
	"bubble/internal/config"
	"bubble/internal/i18n"
	"context"
	"fmt"
	"log"
	"time"
)

var (
	unavailableProviders   []string
	unavailableModels      []string
	lastSuccessfulProvider *config.ProviderConfig
	lastSuccessfulModel    *config.ModelConfig
)

func RefreshAvailableLists() {
	unavailableProviders = []string{}
	unavailableModels = []string{}
	// 重置最后成功的提供商和模型
	lastSuccessfulProvider = nil
	lastSuccessfulModel = nil
}

// 检查字符串是否在切片中
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// 选择提供商和模型
func SelectAvailableProviderAndModel() (*config.ProviderConfig, *config.ModelConfig, error) {

	// 1. 优先使用最后一次成功的提供商和模型
	if lastSuccessfulProvider != nil && lastSuccessfulModel != nil {
		// 检查提供商是否仍然可用
		if !contains(unavailableProviders, lastSuccessfulProvider.Name) && lastSuccessfulProvider.Enable {
			// 检查模型是否仍然可用
			if !contains(unavailableModels, lastSuccessfulModel.ID) {
				// 验证模型是否仍然有效
				_, err := tryModel(lastSuccessfulModel, lastSuccessfulProvider)
				if err == nil {
					log.Printf("%s: %s", i18n.T("using_provider"), lastSuccessfulProvider.Name)
					log.Printf("%s: %s", i18n.T("using_model"), lastSuccessfulModel.ID)
					return lastSuccessfulProvider, lastSuccessfulModel, nil
				}
			}
		}
		return lastSuccessfulProvider, lastSuccessfulModel, nil
	}

	// 2. 检查默认提供商
	defaultProviderName := config.AppConfig.Models.Default
	if defaultProviderName != "" && !contains(unavailableProviders, defaultProviderName) {
		// 查找默认提供商是否存在且启用
		for i := range config.AppConfig.Models.Providers {
			provider := &config.AppConfig.Models.Providers[i]
			if provider.Name == defaultProviderName && provider.Enable {
				log.Printf("%s: %s", i18n.T("checking_provider"), provider.Name)
				// 尝试使用默认模型
				model, err := tryDefaultModel(provider)
				if err == nil {
					// 记录最后成功的提供商和模型
					lastSuccessfulProvider = provider
					lastSuccessfulModel = model
					return provider, model, nil
				}
				// 默认模型不可用，检查是否启用了模型级故障转移
				if provider.Failover {
					// 尝试其他模型
					model, err = tryOtherModels(provider)
					if err == nil {
						// 记录最后成功的提供商和模型
						lastSuccessfulProvider = provider
						lastSuccessfulModel = model
						return provider, model, nil
					}
				}
				// 所有模型都不可用，将提供商添加到不可用列表
				unavailableProviders = append(unavailableProviders, provider.Name)
				log.Printf("%s: %s", i18n.T("provider_added_to_unavailable"), provider.Name)
				break
			}
		}
		log.Printf("%s: %s", i18n.T("default_provider_not_available"), defaultProviderName)
	}

	// 3. 检查其他提供商（如果启用了全局故障转移）
	if config.AppConfig.Models.Failover {
		for i := range config.AppConfig.Models.Providers {
			provider := &config.AppConfig.Models.Providers[i]
			// 跳过不可用或未启用的提供商
			if contains(unavailableProviders, provider.Name) || !provider.Enable {
				continue
			}
			// 跳过默认提供商（已经检查过）
			if provider.Name == defaultProviderName {
				continue
			}

			log.Printf("%s: %s", i18n.T("checking_provider"), provider.Name)
			// 尝试使用默认模型
			model, err := tryDefaultModel(provider)
			if err == nil {
				// 记录最后成功的提供商和模型
				lastSuccessfulProvider = provider
				lastSuccessfulModel = model
				return provider, model, nil
			}
			// 默认模型不可用，检查是否启用了模型级故障转移
			if provider.Failover {
				// 尝试其他模型
				model, err = tryOtherModels(provider)
				if err == nil {
					// 记录最后成功的提供商和模型
					lastSuccessfulProvider = provider
					lastSuccessfulModel = model
					return provider, model, nil
				}
			}
			// 所有模型都不可用，将提供商添加到不可用列表
			unavailableProviders = append(unavailableProviders, provider.Name)
			log.Printf("%s: %s", i18n.T("provider_added_to_unavailable"), provider.Name)
		}
	}

	return nil, nil, fmt.Errorf("%s", i18n.T("no_available_providers"))
}

// tryDefaultModel 尝试使用提供商的默认模型
func tryDefaultModel(provider *config.ProviderConfig) (*config.ModelConfig, error) {
	defaultModelName := provider.Default
	if defaultModelName == "" {
		return nil, fmt.Errorf("%s", i18n.T("no_default_model_specified"))
	}

	// 查找默认模型
	for i := range provider.Models {
		model := &provider.Models[i]
		if model.ID == defaultModelName {
			// 检查模型是否在不可用列表中
			if contains(unavailableModels, model.ID) {
				log.Printf("%s %s", i18n.T("default_model_not_available"), model.ID)
				return nil, fmt.Errorf("%s %s", i18n.T("default_model_not_available"), model.ID)
			}
			// 尝试使用模型
			return tryModel(model, provider)
		}
	}

	log.Printf("%s %s", i18n.T("default_model_not_available"), defaultModelName)
	return nil, fmt.Errorf("%s %s", i18n.T("default_model_not_available"), defaultModelName)
}

// tryOtherModels 尝试使用提供商的其他模型
func tryOtherModels(provider *config.ProviderConfig) (*config.ModelConfig, error) {
	for i := range provider.Models {
		model := &provider.Models[i]
		// 跳过默认模型（已经尝试过）
		if model.ID == provider.Default {
			continue
		}
		// 跳过不可用的模型
		if contains(unavailableModels, model.ID) {
			continue
		}
		// 尝试使用模型
		model, err := tryModel(model, provider)
		if err == nil {
			return model, nil
		}
	}

	return nil, fmt.Errorf("%s", i18n.T("no_available_models"))
}

// tryModel 尝试使用模型发送测试请求
func tryModel(model *config.ModelConfig, provider *config.ProviderConfig) (*config.ModelConfig, error) {
	log.Printf("%s %s", i18n.T("checking_model"), model.ID)

	// 创建带超时的上下文，设置较短的超时时间
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 尝试创建客户端
	client, err := NewClient(
		provider.APIKey,
		WithModel(model.ID),
		WithBaseURL(provider.BaseURL),
		WithMaxTokens(10), // 减少响应长度，加快测试速度
	)
	if err != nil {
		log.Printf("%s %s %s", i18n.T("failed_to_create_client"), provider.Name, err)
		// 将模型添加到不可用列表
		unavailableModels = append(unavailableModels, model.ID)
		return nil, err
	}

	// 尝试发送请求，使用更短的测试消息
	_, err = client.Send(timeoutCtx, "hi")
	if err != nil {
		log.Printf("%s %s %s", i18n.T("request_failed_with_provider"), provider.Name, err)
		// 将模型添加到不可用列表
		unavailableModels = append(unavailableModels, model.ID)
		return nil, err
	}

	// 请求成功
	return model, nil
}
