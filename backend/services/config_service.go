package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigData YAML配置文件的数据结构
type ConfigData struct {
	OpenAIAPIURL    string `yaml:"openai_api_url" json:"openai_api_url"`
	OpenAIModelType string `yaml:"openai_model_type" json:"openai_model_type"`
	OpenAIAPIKey    string `yaml:"openai_api_key" json:"openai_api_key"`
}

// ConfigItem 配置项结构
type ConfigItem struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ConfigService 配置服务
type ConfigService struct {
	configPath string
	configs    map[string]*ConfigItem
}

// NewConfigService 创建配置服务实例
func NewConfigService() *ConfigService {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get user home directory: %v", err)
		homeDir = "."
	}

	// 设置配置文件路径
	configPath := filepath.Join(homeDir, ".loji-app", "config.yaml")

	return &ConfigService{
		configPath: configPath,
		configs: map[string]*ConfigItem{
			"OPENAI_API_URL": {
				Key:         "OPENAI_API_URL",
				Description: "OpenAI API地址",
				Required:    false,
			},
			"OPENAI_MODEL_TYPE": {
				Key:         "OPENAI_MODEL_TYPE",
				Description: "OpenAI模型类型",
				Required:    false,
			},
			"OPENAI_API_KEY": {
				Key:         "OPENAI_API_KEY",
				Description: "OpenAI API密钥",
				Required:    true,
			},
		},
	}
}

// loadConfig 从YAML文件加载配置
func (s *ConfigService) loadConfig() (*ConfigData, error) {
	// 如果配置文件不存在，返回空配置
	if _, err := os.Stat(s.configPath); os.IsNotExist(err) {
		return &ConfigData{}, nil
	}

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config ConfigData
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// saveConfig 保存配置到YAML文件
func (s *ConfigService) saveConfig(config *ConfigData) error {
	// 确保目录存在
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	log.Printf("配置已保存到: %s", s.configPath)
	return nil
}

// GetConfig 获取单个配置项
func (s *ConfigService) GetConfig(key string) (*ConfigItem, error) {
	config, exists := s.configs[key]
	if !exists {
		return nil, fmt.Errorf("配置项 %s 不存在", key)
	}

	// 从YAML文件获取当前值
	configData, err := s.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 根据key设置对应的值
	switch key {
	case "OPENAI_API_URL":
		config.Value = configData.OpenAIAPIURL
	case "OPENAI_MODEL_TYPE":
		config.Value = configData.OpenAIModelType
	case "OPENAI_API_KEY":
		config.Value = configData.OpenAIAPIKey
	default:
		config.Value = ""
	}

	return config, nil
}

// GetAllConfigs 获取所有配置项
func (s *ConfigService) GetAllConfigs() (map[string]*ConfigItem, error) {
	result := make(map[string]*ConfigItem)

	// 从YAML文件获取当前值
	configData, err := s.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	b, _ := json.Marshal(configData)
	log.Printf("configData: %s", string(b))

	for key, config := range s.configs {
		// 深拷贝避免引用问题
		configCopy := *config

		// 根据key设置对应的值
		switch key {
		case "OPENAI_API_URL":
			configCopy.Value = configData.OpenAIAPIURL
		case "OPENAI_MODEL_TYPE":
			configCopy.Value = configData.OpenAIModelType
		case "OPENAI_API_KEY":
			configCopy.Value = configData.OpenAIAPIKey
		default:
			configCopy.Value = ""
		}

		result[key] = &configCopy
	}

	return result, nil
}

// SetConfig 设置配置项并保存到YAML文件
func (s *ConfigService) SetConfig(key, value string) error {
	config, exists := s.configs[key]
	if !exists {
		return fmt.Errorf("配置项 %s 不存在", key)
	}

	// 验证必填项
	if config.Required && value == "" {
		return fmt.Errorf("配置项 %s 是必填项，不能为空", key)
	}

	// 加载现有配置
	configData, err := s.loadConfig()
	if err != nil {
		return fmt.Errorf("加载现有配置失败: %w", err)
	}

	// 根据key设置对应的值
	switch key {
	case "OPENAI_API_URL":
		configData.OpenAIAPIURL = value
	case "OPENAI_MODEL_TYPE":
		configData.OpenAIModelType = value
	case "OPENAI_API_KEY":
		configData.OpenAIAPIKey = value
	}

	// 保存配置到YAML文件
	if err := s.saveConfig(configData); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	log.Printf("配置项 %s 已设置为: %s", key, value)
	return nil
}

// SetConfigs 批量设置配置项
func (s *ConfigService) SetConfigs(configs map[string]string) error {
	for key, value := range configs {
		if err := s.SetConfig(key, value); err != nil {
			return fmt.Errorf("设置配置项 %s 失败: %w", key, err)
		}
	}
	return nil
}

// ValidateConfigs 验证所有配置项
func (s *ConfigService) ValidateConfigs() error {
	configs, err := s.GetAllConfigs()
	if err != nil {
		return err
	}

	var errors []string
	for _, config := range configs {
		if config.Required && config.Value == "" {
			errors = append(errors, fmt.Sprintf("配置项 %s 是必填项，不能为空", config.Key))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("配置验证失败: %v", errors)
	}

	return nil
}

// ResetConfig 重置配置项为默认值或空值
func (s *ConfigService) ResetConfig(key string) error {
	_, exists := s.configs[key]
	if !exists {
		return fmt.Errorf("配置项 %s 不存在", key)
	}

	// 加载现有配置
	configData, err := s.loadConfig()
	if err != nil {
		return fmt.Errorf("加载现有配置失败: %w", err)
	}

	// 根据key清空对应的值
	switch key {
	case "OPENAI_API_URL":
		configData.OpenAIAPIURL = ""
	case "OPENAI_MODEL_TYPE":
		configData.OpenAIModelType = ""
	case "OPENAI_API_KEY":
		configData.OpenAIAPIKey = ""
	}

	// 保存配置到YAML文件
	if err := s.saveConfig(configData); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	log.Printf("配置项 %s 已重置", key)
	return nil
}

// GetConfigurableKeys 获取所有可配置的键名
func (s *ConfigService) GetConfigurableKeys() []string {
	keys := make([]string, 0, len(s.configs))
	for key := range s.configs {
		keys = append(keys, key)
	}
	return keys
}
