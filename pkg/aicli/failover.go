package aicli

import (
	"bubble/internal/config"
	"bubble/internal/i18n"
	"context"
	"fmt"
	"log"
)

var (
	unavailableProviders []string
	unavailableModels    []string
)

// ResetUnavailableLists 重置不可用列表
func ResetUnavailableLists() {
	unavailableProviders = []string{}
	unavailableModels = []string{}
	log.Println("Unavailable lists reset")
}

// 选择提供商和模型
func selectProviderAndModel() (*config.ProviderConfig, *config.ModelConfig, error) {
	log.Println(i18n.T("checking_provider"))

	// 选择提供商
	selectedProvider, err := selectProvider()
	if err != nil {
		log.Printf(i18n.T("failed_to_select_provider")+" %s", err)
		return nil, nil, fmt.Errorf(i18n.T("failed_to_select_provider")+" %w", err)
	}

	log.Printf(i18n.T("selected_provider")+" %s", selectedProvider.Name)
	log.Println(i18n.T("checking_model"))

	// 选择模型
	modelConfig, err := selectModel(selectedProvider)
	if err != nil {
		log.Printf(i18n.T("failed_to_select_model")+" %s", err)
		// 将当前提供商添加到不可用列表，因为它没有可用的模型
		unavailableProviders = append(unavailableProviders, selectedProvider.Name)
		return nil, nil, fmt.Errorf(i18n.T("failed_to_select_model")+" %w", err)
	}

	log.Printf(i18n.T("selected_model")+" %s", modelConfig.ID)

	return selectedProvider, modelConfig, nil
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

// 选择提供商
func selectProvider() (*config.ProviderConfig, error) {
	// 优先使用默认提供商
	defaultProviderName := config.AppConfig.Models.Default
	// 设置了默认提供商，并且不在不可用列表中
	if defaultProviderName != "" && !contains(unavailableProviders, defaultProviderName) {
		// 查找默认提供商是否存在且启用
		for i := range config.AppConfig.Models.Providers {
			provider := &config.AppConfig.Models.Providers[i]
			if provider.Name == defaultProviderName && provider.Enable {
				log.Printf("%s %s", i18n.T("using_provider"), provider.Name)
				return provider, nil
			}
		}
		log.Printf("%s %s", i18n.T("default_provider_not_available"), defaultProviderName)
	} else if defaultProviderName == "" {
		// 没有指定默认提供商
		log.Printf("%s", i18n.T("no_default_provider_specified"))
	} else {
		// 默认提供商在不可用列表中
		log.Printf("%s %s", i18n.T("provider_not_available"), defaultProviderName)
	}

	// 故障转移启用时，选择其他提供商
	if config.AppConfig.Models.Failover {
		for i := range config.AppConfig.Models.Providers {
			provider := &config.AppConfig.Models.Providers[i]
			if !contains(unavailableProviders, provider.Name) && provider.Enable {
				log.Printf("%s %s", i18n.T("using_provider"), provider.Name)
				return provider, nil
			}
		}
	}

	return nil, fmt.Errorf("%s", i18n.T("no_available_providers"))
}

// 选择模型
func selectModel(provider *config.ProviderConfig) (*config.ModelConfig, error) {
	// 优先使用默认模型
	defaultModelName := provider.Default
	if defaultModelName != "" && !contains(unavailableModels, defaultModelName) {
		// 查找默认模型是否存在
		for i := range provider.Models {
			model := &provider.Models[i]
			if model.ID == defaultModelName {
				return model, nil
			}
		}
		log.Printf("%s %s", i18n.T("default_model_not_available"), defaultModelName)
	} else if defaultModelName == "" {
		// 没有指定默认模型
		log.Printf("%s", i18n.T("no_default_model_specified"))
	} else {
		// 默认模型在不可用列表中
		log.Printf("%s %s", i18n.T("default_model_not_available"), defaultModelName)
	}

	// 启用故障转移，选择其他模型
	if provider.Failover {
		for i := range provider.Models {
			model := &provider.Models[i]
			if model.ID != "" && !contains(unavailableModels, model.ID) {
				log.Printf("%s %s", i18n.T("using_model"), model.ID)
				return model, nil
			}
		}
	}

	return nil, fmt.Errorf("%s", i18n.T("no_available_models"))
}

// 故障转移
func providersFailover(prompt string, failedProvider *config.ProviderConfig) (*config.ProviderConfig, *config.ModelConfig, string, error) {
	log.Println(i18n.T("attempting_failover_to_other_providers"))

	// 将失败的提供商添加到不可用列表
	unavailableProviders = append(unavailableProviders, failedProvider.Name)

	// 外层循环：遍历所有提供商
	for _, provider := range config.AppConfig.Models.Providers {
		// 跳过不可用的提供商或未启用的提供商
		if contains(unavailableProviders, provider.Name) || !provider.Enable {
			continue
		}

		log.Printf(i18n.T("using_provider")+" %s", provider.Name)

		// 检查该提供商是否启用了模型级故障转移
		if provider.Failover {
			// 启用了故障转移，尝试所有模型
			for _, model := range provider.Models {
				// 跳过不可用的模型
				if contains(unavailableModels, model.ID) {
					continue
				}

				log.Printf(i18n.T("using_model")+" %s", model.ID)

				// 尝试创建客户端
				client, err := NewClient(
					provider.APIKey,
					WithModel(model.ID),
					WithBaseURL(provider.BaseURL),
					WithEnableContext(),
					WithInitialRole(GenericRoleDescCN),
				)
				if err != nil {
					log.Printf(i18n.T("failed_to_create_client")+" %s %s", provider.Name, err)
					// 将模型添加到不可用列表
					unavailableModels = append(unavailableModels, model.ID)
					continue
				}

				// 尝试发送请求
				response, err := client.Send(context.Background(), prompt)
				if err != nil {
					log.Printf(i18n.T("request_failed_with_provider")+" %s %s", provider.Name, err)
					// 将模型添加到不可用列表
					unavailableModels = append(unavailableModels, model.ID)
					continue
				}

				// 请求成功
				log.Printf(i18n.T("Response")+": %s", response)
				return &provider, &model, response, nil
			}
			// 该提供商的所有模型都失败，将其添加到不可用列表
			unavailableProviders = append(unavailableProviders, provider.Name)
		} else {
			// 未启用故障转移，只尝试默认模型
			if provider.Default != "" && !contains(unavailableModels, provider.Default) {
				log.Printf(i18n.T("using_model")+" %s", provider.Default)

				// 尝试创建客户端
				client, err := NewClient(
					provider.APIKey,
					WithModel(provider.Default),
					WithBaseURL(provider.BaseURL),
					WithEnableContext(),
					WithInitialRole(GenericRoleDescCN),
				)
				if err != nil {
					log.Printf(i18n.T("failed_to_create_client")+" %s %s", provider.Name, err)
					// 将模型添加到不可用列表
					unavailableModels = append(unavailableModels, provider.Default)
					// 未启用故障转移，将提供商添加到不可用列表
					unavailableProviders = append(unavailableProviders, provider.Name)
				} else {
					// 尝试发送请求
					response, err := client.Send(context.Background(), prompt)
					if err != nil {
						log.Printf(i18n.T("request_failed_with_provider")+" %s %s", provider.Name, err)
						// 将模型添加到不可用列表
						unavailableModels = append(unavailableModels, provider.Default)
						// 未启用故障转移，将提供商添加到不可用列表
						unavailableProviders = append(unavailableProviders, provider.Name)
					} else {
						// 请求成功
						log.Printf(i18n.T("Response")+": %s", response)
						// 查找默认模型
						var selectedModel *config.ModelConfig
						for i := range provider.Models {
							model := &provider.Models[i]
							if model.ID == provider.Default {
								selectedModel = model
								break
							}
						}
						return &provider, selectedModel, response, nil
					}
				}
			} else {
				// 没有可用的默认模型，将提供商添加到不可用列表
				unavailableProviders = append(unavailableProviders, provider.Name)
			}
		}
	}

	// 所有提供商和模型都失败
	log.Println(i18n.T("all_providers_failed"))
	return nil, nil, "", fmt.Errorf("%s", i18n.T("all_providers_failed"))
}
