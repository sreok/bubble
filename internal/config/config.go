package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"log"

	"gopkg.in/yaml.v3"
)

type ModelConfig struct {
	ID     string   `yaml:"id"`
	Input  []string `yaml:"input"`
	Output []string `yaml:"output"`
}

type ProviderConfig struct {
	Name     string        `yaml:"name"`
	Enable   bool          `yaml:"enable"`
	Failover bool          `yaml:"failover"`
	Primary  string        `yaml:"primary"`
	Models   []ModelConfig `yaml:"models"`
	APIKey   string        `yaml:"api_key"`
	BaseURL  string        `yaml:"base_url"`
}

type Config struct {
	Models struct {
		Failover  bool             `yaml:"failover"`
		Primary   string           `yaml:"primary"`
		Providers []ProviderConfig `yaml:"providers"`
	} `yaml:"models"`
}

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
		log.Printf("Failed to get user home directory: %v\n", err)
		return
	}
	configPath = filepath.Join(homeDir, ".bubble", "config.yaml")
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
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		// 创建最小配置文件
		defaultConfig := Config{
			Models: struct {
				Failover  bool             `yaml:"failover"`
				Primary   string           `yaml:"primary"`
				Providers []ProviderConfig `yaml:"providers"`
			}{
				Failover: true,
				Primary:  "iflow",
				Providers: []ProviderConfig{
					{Name: "custom", Enable: true, Failover: true, Primary: "qwen3.5:9b", Models: []ModelConfig{
						{ID: "qwen3.5:9b", Input: []string{"text", "image"}, Output: []string{"text"}},
						{ID: "qwen3:8b", Input: []string{"text", "image"}, Output: []string{"text"}},
					}},
					{Name: "iflow", Enable: true, Failover: true, Primary: "qwen3-max", Models: []ModelConfig{
						{ID: "qwen3-max", Input: []string{"text"}, Output: []string{"text"}},
						{ID: "qwen3-coder-plus", Input: []string{"text"}, Output: []string{"text"}},
						{ID: "kimi-k2-0905", Input: []string{"text"}, Output: []string{"text"}},
					}},
				},
			},
		}
		// 写入配置文件
		data, err := yaml.Marshal(defaultConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal default config: %w", err)
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		// 设置 AppConfig 为默认配置
		AppConfig = defaultConfig
		lastModified = time.Now()
		return nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Failed to read config file: %v\n", err)
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置文件
	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		log.Printf("Failed to unmarshal config file: %v\n", err)
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	// 更新最后修改时间
	lastModified = fileInfo.ModTime()
	return nil
}

// ReloadConfig 重新加载配置文件
func ReloadConfig() error {
	mu.Lock()
	defer mu.Unlock()

	return LoadConfig()
}

// CheckAndReload 检查配置文件是否变化并重新加载
func CheckAndReload() error {
	// 先获取文件信息，不持有锁
	fileInfo, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		// 配置文件不存在，不需要重新加载
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	// 检查文件是否被修改
	mu.RLock()
	modified := fileInfo.ModTime().After(lastModified)
	mu.RUnlock()

	if modified {
		log.Printf("Config file %s has been modified, reloading...\n", configPath)
		return LoadConfig()
	}

	return nil
}

// StartConfigWatcher 启动配置文件监视
func StartConfigWatcher(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := CheckAndReload(); err != nil {
			log.Printf("Failed to reload config: %v\n", err)
		}
	}
}
