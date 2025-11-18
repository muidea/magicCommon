package configuration

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultConfigManager 默认配置管理器实例
var DefaultConfigManager ConfigManager

// InitDefaultConfigManager 初始化默认配置管理器
func InitDefaultConfigManager(configDir string) error {
	// 如果未指定配置目录，使用默认值
	if configDir == "" {
		// 尝试从环境变量获取配置目录
		if envConfigDir := os.Getenv("CONFIG_DIR"); envConfigDir != "" {
			configDir = envConfigDir
		} else {
			// 使用当前工作目录下的config目录
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			configDir = filepath.Join(wd, "config")
		}
	}

	// 创建配置选项
	options := &ConfigOptions{
		ConfigDir:       configDir,
		WatchInterval:   time.Second * 5,
		EnableHotReload: true,
		Validator:       &defaultValidator{},
	}

	// 创建配置管理器
	manager, err := NewConfigManager(options)
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	DefaultConfigManager = manager
	return nil
}

// GetDefaultConfigManager 获取默认配置管理器
func GetDefaultConfigManager() ConfigManager {
	return DefaultConfigManager
}

// MustInitDefaultConfigManager 初始化默认配置管理器，如果失败则panic
func MustInitDefaultConfigManager(configDir string) {
	if err := InitDefaultConfigManager(configDir); err != nil {
		panic(fmt.Sprintf("Failed to initialize default config manager: %v", err))
	}
}

// CreateConfigManager 创建配置管理器（工厂函数）
func CreateConfigManager(options *ConfigOptions) (ConfigManager, error) {
	return NewConfigManager(options)
}

// CreateConfigManagerWithDir 使用配置目录创建配置管理器
func CreateConfigManagerWithDir(configDir string, enableHotReload bool) (ConfigManager, error) {
	options := &ConfigOptions{
		ConfigDir:       configDir,
		WatchInterval:   time.Second * 5,
		EnableHotReload: enableHotReload,
		Validator:       &defaultValidator{},
	}
	return NewConfigManager(options)
}

// CreateSimpleConfigManager 创建简单的配置管理器（无热加载）
func CreateSimpleConfigManager(configDir string) (ConfigManager, error) {
	return CreateConfigManagerWithDir(configDir, false)
}

// CreateConfigManagerWithValidator 创建带验证器的配置管理器
func CreateConfigManagerWithValidator(configDir string, validator ConfigValidator) (ConfigManager, error) {
	options := &ConfigOptions{
		ConfigDir:       configDir,
		WatchInterval:   time.Second * 5,
		EnableHotReload: true,
		Validator:       validator,
	}
	return NewConfigManager(options)
}

// Helper functions for common configuration operations

// GetString 获取字符串配置值
func GetString(key string) (string, error) {
	if DefaultConfigManager == nil {
		return "", fmt.Errorf("default config manager not initialized")
	}

	value, err := DefaultConfigManager.Get(key)
	if err != nil {
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("config value is not a string: %s", key)
	}

	return str, nil
}

// GetStringWithDefault 获取字符串配置值，如果不存在则返回默认值
func GetStringWithDefault(key, defaultValue string) string {
	if DefaultConfigManager == nil {
		return defaultValue
	}

	value := DefaultConfigManager.GetWithDefault(key, defaultValue)
	str, ok := value.(string)
	if !ok {
		return defaultValue
	}

	return str
}

// GetInt 获取整数配置值
func GetInt(key string) (int, error) {
	if DefaultConfigManager == nil {
		return 0, fmt.Errorf("default config manager not initialized")
	}

	value, err := DefaultConfigManager.Get(key)
	if err != nil {
		return 0, err
	}

	// 尝试转换为整数
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("config value is not an integer: %s", key)
	}
}

// GetIntWithDefault 获取整数配置值，如果不存在则返回默认值
func GetIntWithDefault(key string, defaultValue int) int {
	if DefaultConfigManager == nil {
		return defaultValue
	}

	value := DefaultConfigManager.GetWithDefault(key, defaultValue)

	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return defaultValue
	}
}

// GetBool 获取布尔配置值
func GetBool(key string) (bool, error) {
	if DefaultConfigManager == nil {
		return false, fmt.Errorf("default config manager not initialized")
	}

	value, err := DefaultConfigManager.Get(key)
	if err != nil {
		return false, err
	}

	boolVal, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("config value is not a boolean: %s", key)
	}

	return boolVal, nil
}

// GetBoolWithDefault 获取布尔配置值，如果不存在则返回默认值
func GetBoolWithDefault(key string, defaultValue bool) bool {
	if DefaultConfigManager == nil {
		return defaultValue
	}

	value := DefaultConfigManager.GetWithDefault(key, defaultValue)
	boolVal, ok := value.(bool)
	if !ok {
		return defaultValue
	}

	return boolVal
}

// GetModuleString 获取模块字符串配置值
func GetModuleString(moduleName, key string) (string, error) {
	if DefaultConfigManager == nil {
		return "", fmt.Errorf("default config manager not initialized")
	}

	value, err := DefaultConfigManager.GetModuleConfig(moduleName, key)
	if err != nil {
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("module config value is not a string: %s.%s", moduleName, key)
	}

	return str, nil
}

// GetModuleStringWithDefault 获取模块字符串配置值，如果不存在则返回默认值
func GetModuleStringWithDefault(moduleName, key, defaultValue string) string {
	if DefaultConfigManager == nil {
		return defaultValue
	}

	value := DefaultConfigManager.GetModuleConfigWithDefault(moduleName, key, defaultValue)
	str, ok := value.(string)
	if !ok {
		return defaultValue
	}

	return str
}

// WatchConfig 监听配置变更（使用默认配置管理器）
func WatchConfig(key string, handler ConfigChangeHandler) error {
	if DefaultConfigManager == nil {
		return fmt.Errorf("default config manager not initialized")
	}

	return DefaultConfigManager.Watch(key, handler)
}

// WatchModuleConfig 监听模块配置变更（使用默认配置管理器）
func WatchModuleConfig(moduleName, key string, handler ConfigChangeHandler) error {
	if DefaultConfigManager == nil {
		return fmt.Errorf("default config manager not initialized")
	}

	return DefaultConfigManager.WatchModule(moduleName, key, handler)
}
