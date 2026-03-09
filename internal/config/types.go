package config

// ProviderConfig 提供商配置结构体
type ProviderConfig struct {
	Name     string        `yaml:"name"`     // 提供商名称
	Enable   bool          `yaml:"enable"`   // 是否启用
	Failover bool          `yaml:"failover"` // 是否参与故障转移
	Default  string        `yaml:"default"`  // 默认模型
	Models   []ModelConfig `yaml:"models"`   // 支持的模型列表
	APIKey   string        `yaml:"api_key"`  // API密钥
	BaseURL  string        `yaml:"base_url"` // API基础URL
}

// ModelsConfig 模型相关配置结构体
type ModelsConfig struct {
	Failover  bool             `yaml:"failover"`  // 是否启用故障转移
	Default   string           `yaml:"default"`   // 默认提供商
	Providers []ProviderConfig `yaml:"providers"` // 提供商列表
}

// ModelConfig 模型配置结构体
type ModelConfig struct {
	ID     string   `yaml:"id"`     // 模型ID
	Input  []string `yaml:"input"`  // 支持的输入类型，如text、image等
	Output []string `yaml:"output"` // 支持的输出类型，如text等
}

// ToolsConfig 工具配置结构体
type ToolsConfig struct {
	WebSearch WebSearchConfig `yaml:"web_search"` // 网络搜索配置
}

// WebSearchConfig 网络搜索配置结构体
type WebSearchConfig struct {
	Tavily TavilyConfig `yaml:"tavily"` // Tavily 配置
}

// TavilyConfig Tavily 配置结构体
type TavilyConfig struct {
	Enable bool   `yaml:"enable"`  // 是否启用
	APIKey string `yaml:"api_key"` // API 密钥
}

type AgentConfig struct {
	Workspace string `yaml:"workspace"` // 工作空间,默认值为"～/.bubble/workspace"
}

// Config 主配置结构体
type Config struct {
	Language string       `yaml:"language"` // 语言设置，支持 en 和 zh
	Models   ModelsConfig `yaml:"models"`   // 模型配置
	Tools    ToolsConfig  `yaml:"tools"`    // 工具配置
	Agent    AgentConfig  `yaml:"agent"`    // 智能体配置
}

// 定义默认的角色描述
const (
	GenericRoleDescEN = `Your name is Bubble, and you are a helpful assistant, able to answer user questions in a clear and friendly manner.`
	GenericRoleDescCN = `你叫泡泡，你是一个有帮助的助手，能够以清晰、友好的方式回答用户的问题。`

	EnhancedRoleDescCN = `你是一个有帮助的助手，能够以清晰、友好的方式回答用户的问题。
你擅长使用 Go 的标准库或第三方库来实现业务逻辑代码。`
	EnhancedRoleDescEN = `You are a helpful assistant able to answer user questions in a clear and friendly manner.
You excel at using Go's standard library or third-party libraries to implement business logic code.`
)
