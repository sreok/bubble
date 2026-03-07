package i18n

import (
	"fmt"
	"sync"
)

// Language 语言类型
type Language string

const (
	// 语言常量
	English Language = "en"
	Chinese Language = "zh"
)

// 翻译映射
var translations = map[Language]map[string]string{
	English: {
		"config_load_failed":                     "Failed to load config",
		"config_create_dir_failed":               "Failed to create config directory",
		"config_stat_failed":                     "Failed to get config file status",
		"config_modified_reload":                 "Reload config file",
		"config_unmarshal_failed":                "Failed to unmarshal config file",
		"config_write_failed":                    "Failed to write config file",
		"invalid_request_payload":                "Invalid request payload",
		"no_providers_found":                     "No providers found in config",
		"failed_to_get_message":                  "Failed to get message",
		"config_get_home_dir_failed":             "Failed to get user home directory",
		"config_read_failed":                     "Failed to read config file",
		"config_validation_failed":               "Failed to validate config file",
		"failed_to_select_provider":              "Failed to select provider",
		"failed_to_select_model":                 "Failed to select model",
		"failed_to_load_config":                  "Failed to load config",
		"default_provider_not_available":         "Default provider not available",
		"default_model_not_available":            "Default model not available",
		"using_provider":                         "Using provider",
		"using_model":                            "Using model",
		"no_available_providers":                 "No available providers",
		"no_models_found_for_provider":           "Provider has no available models",
		"request_failed_with_provider":           "Request failed with provider %s",
		"attempting_failover_to_other_providers": "Attempting failover to other providers",
		"switching_to_provider":                  "Switching to provider",
		"failed_to_create_client":                "Failed to create AI client",
		"all_providers_failed":                   "All providers failed",
		"Response":                               "Response",
		"Error":                                  "Error",
		"Provider":                               "Provider",
		"Model":                                  "Model",
		"provider_not_available":                 "Provider not available",
		"no_default_provider_specified":          "No default provider specified",
		"no_default_model_specified":             "No default model specified",
		"checking_provider":                      "Checking provider",
		"checking_model":                         "Checking model",
		"selected_provider":                      "Selected provider",
		"selected_model":                         "Selected model",
		"failed_to_select_provider_or_model":     "Failed to select provider or model",
	},
	Chinese: {
		"config_load_failed":                     "加载配置失败",
		"config_create_dir_failed":               "创建配置目录失败",
		"config_stat_failed":                     "获取配置文件状态失败",
		"config_modified_reload":                 "重载配置文件",
		"config_unmarshal_failed":                "解析配置文件失败",
		"config_write_failed":                    "写入配置文件失败",
		"invalid_request_payload":                "无效的请求负载",
		"no_providers_found":                     "配置中未找到提供商",
		"failed_to_get_message":                  "获取消息失败",
		"config_get_home_dir_failed":             "获取用户主目录失败",
		"config_read_failed":                     "读取配置文件失败",
		"config_validation_failed":               "配置验证失败",
		"failed_to_select_provider":              "选择提供商失败",
		"failed_to_select_model":                 "选择模型失败",
		"failed_to_load_config":                  "加载配置失败",
		"default_provider_not_available":         "默认提供商不可用",
		"default_model_not_available":            "默认模型不可用",
		"using_provider":                         "正在使用提供商",
		"using_model":                            "正在使用模型",
		"no_available_providers":                 "没有可用的提供商",
		"no_models_found_for_provider":           "提供商未找到可用模型",
		"request_failed_with_provider":           "提供商请求失败",
		"attempting_failover_to_other_providers": "尝试故障转移到其他提供商",
		"switching_to_provider":                  "切换到提供商",
		"failed_to_create_client":                "创建客户端失败",
		"all_providers_failed":                   "所有提供商都失败了",
		"Response":                               "响应",
		"Error":                                  "错误",
		"Provider":                               "提供商",
		"Model":                                  "模型",
		"provider_not_available":                 "提供商不可用",
		"no_default_provider_specified":          "未指定默认提供商",
		"no_default_model_specified":             "未指定默认模型",
		"checking_provider":                      "检查提供商",
		"checking_model":                         "检查模型",
		"selected_provider":                      "选择提供商",
		"selected_model":                         "选择模型",
		"failed_to_select_provider_or_model":     "选择提供商或模型失败",
	},
}

var (
	currentLang Language = English
	mu          sync.RWMutex
)

// SetLanguage 设置当前语言
func SetLanguage(lang Language) {
	mu.Lock()
	defer mu.Unlock()

	// 验证语言是否有效
	if _, ok := translations[lang]; !ok {
		// 如果语言无效，使用默认语言英语
		// log.Printf("Invalid language: %s, using default language: %s\n", lang, English)
		lang = English
	}

	currentLang = lang
	// log.Printf("Language set to: %s\n", lang)
}

// GetLanguage 获取当前语言
func GetLanguage() Language {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// T 翻译函数
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if trans, ok := translations[currentLang][key]; ok {
		return trans
	}
	// 如果找不到翻译，返回原文
	return key
}

// TF 带格式化的翻译函数
func TF(key string, args ...interface{}) string {
	mu.RLock()
	defer mu.RUnlock()
	if trans, ok := translations[currentLang][key]; ok {
		return sprintf(trans, args...)
	}
	// 如果找不到翻译，返回原文
	return sprintf(key, args...)
}

// sprintf 格式化字符串
func sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
