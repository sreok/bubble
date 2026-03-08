package config

import (
	"bubble/internal/i18n"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"log"

	"gopkg.in/yaml.v3"
)

var (
	AppConfig    Config
	configPath   string
	lastModified time.Time
	mu           sync.RWMutex
)

// init 初始化配置文件路径
func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf(i18n.T("config_get_home_dir_failed"), err)
		return
	}
	configPath = filepath.Join(homeDir, ".bubble", "config.yaml")
	// 设置默认工作空间
	if AppConfig.Agent.Workspace == "" {
		AppConfig.Agent.Workspace = filepath.Join(homeDir, ".bubble", "workspace")
	}
}

// LoadConfig 加载配置文件
func LoadConfig() error {
	mu.Lock()
	defer mu.Unlock()

	// 检查配置文件是否存在
	fileInfo, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		// 如果配置文件不存在，创建最小配置文件
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf(i18n.T("config_create_dir_failed"), err)
		}
		// 创建最小配置文件
		defaultConfig := Config{
			Language: "en", // 默认语言为英语
			Models: ModelsConfig{
				Failover: true,
				Default:  "chatgpt",
				Providers: []ProviderConfig{
					{
						Name:     "chatgpt",
						Enable:   true,
						APIKey:   "sk-1234567890abcdef1234567890abcdef",
						BaseURL:  "https://api.openai.com/v1",
						Failover: true,
						Default:  "gpt-3.5-turbo",
						Models: []ModelConfig{
							{ID: "gpt-3.5-turbo", Input: []string{"text", "image"}, Output: []string{"text"}},
							{ID: "gpt-4", Input: []string{"text", "image"}, Output: []string{"text"}},
						},
					},
					{
						Name:     "deepseek",
						Enable:   true,
						APIKey:   "sk-1234567890abcdef1234567890abcdef",
						BaseURL:  "https://api.deepseek.cn/v1",
						Failover: true,
						Default:  "deepseek-r1",
						Models: []ModelConfig{
							{ID: "deepseek-r1", Input: []string{"text"}, Output: []string{"text"}},
							{ID: "deepseek-coder-plus", Input: []string{"text"}, Output: []string{"text"}},
						},
					},
				},
			},
		}
		// 写入配置文件
		data, err := yaml.Marshal(defaultConfig)
		if err != nil {
			return fmt.Errorf(i18n.T("config_unmarshal_failed"), err)
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf(i18n.T("config_write_failed"), err)
		}
		// 设置 AppConfig 为默认配置
		AppConfig = defaultConfig
		lastModified = time.Now()
		return nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf(i18n.T("config_read_failed"), configPath, err)
		return fmt.Errorf(i18n.T("config_read_failed"), err)
	}

	// 解析配置文件
	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		log.Printf(i18n.T("config_unmarshal_failed"), configPath, err)
		return fmt.Errorf(i18n.T("config_unmarshal_failed"), err)
	}

	// 更新最后修改时间
	lastModified = fileInfo.ModTime()

	// 设置语言
	if AppConfig.Language != "" {
		i18n.SetLanguage(i18n.Language(AppConfig.Language))
	} else {
		i18n.SetLanguage(i18n.English)
	}

	return nil
}

// ReloadConfig 重新加载配置文件
func ReloadConfig() error {
	mu.Lock()
	defer mu.Unlock()

	return LoadConfig()
}

// CheckAndReload 检查配置文件是否变化并重新加载
// 返回值：
// - error: 错误信息
// - bool: 配置文件是否被修改
func CheckAndReload() (error, bool) {
	// 先获取文件信息，不持有锁
	fileInfo, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		// 配置文件不存在，不需要重新加载
		return nil, false
	} else if err != nil {
		return fmt.Errorf(i18n.T("config_stat_failed"), err), false
	}

	// 检查文件是否被修改
	mu.RLock()
	modified := fileInfo.ModTime().After(lastModified)
	mu.RUnlock()

	if modified {
		log.Printf(i18n.T("config_modified_reload")+": %s", configPath)
		err := LoadConfig()
		return err, true
	}

	return nil, false
}

// ValidateConfig 验证配置的有效性
// func ValidateConfig() error {
// 	// 检查是否有启用的模型提供商
// 	enabledProviders := 0
// 	for _, provider := range AppConfig.Models.Providers {
// 		if provider.Enable {
// 			enabledProviders++
// 		}
// 	}
// 	if enabledProviders == 0 {
// 		return fmt.Errorf(i18n.T("no_enabled_providers_found"))
// 	}

// 	// 检查默认模型提供商是否存在且启用
// 	defaultProviderFound := false
// 	for _, provider := range AppConfig.Models.Providers {
// 		if provider.Name == AppConfig.Models.Default && provider.Enable {
// 			defaultProviderFound = true
// 			break
// 		}
// 	}
// 	if !defaultProviderFound {
// 		return fmt.Errorf(i18n.T("default_provider_not_found"), AppConfig.Models.Default)
// 	}

// 	return nil
// }

// StartConfigWatcher 启动配置文件监视
func StartConfigWatcher(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		err, _ := CheckAndReload()
		if err != nil {
			log.Printf(i18n.T("config_reload_failed"), err)
		}
		// else {
		// 	// 验证配置有效性
		// 	if err := ValidateConfig(); err != nil {
		// 		log.Printf(i18n.T("config_validation_failed"), err)
		// 	}
		// }
	}
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	return configPath
}
