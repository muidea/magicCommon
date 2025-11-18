package configuration

import (
	"fmt"
	"os"
	"strings"
)

// EnvConfigLoader 环境变量配置加载器
type EnvConfigLoader struct {
	prefix string
}

// NewEnvConfigLoader 创建环境变量配置加载器
func NewEnvConfigLoader(prefix string) *EnvConfigLoader {
	return &EnvConfigLoader{
		prefix: prefix,
	}
}

// Load 加载环境变量配置
func (l *EnvConfigLoader) Load() (map[string]interface{}, error) {
	config := make(map[string]interface{})

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key := pair[0]
		value := pair[1]

		// 如果设置了前缀，只处理带前缀的环境变量
		if l.prefix != "" && !strings.HasPrefix(key, l.prefix) {
			continue
		}

		// 移除前缀
		if l.prefix != "" {
			key = strings.TrimPrefix(key, l.prefix)
		}

		// 将环境变量名转换为配置键名（下划线转点号，大写转小写）
		configKey := l.normalizeKey(key)
		config[configKey] = l.parseValue(value)
	}

	return config, nil
}

// normalizeKey 规范化键名
func (l *EnvConfigLoader) normalizeKey(key string) string {
	// 将下划线转换为点号
	key = strings.ReplaceAll(key, "_", ".")
	// 转换为小写
	return strings.ToLower(key)
}

// parseValue 解析环境变量值
func (l *EnvConfigLoader) parseValue(value string) interface{} {
	// 尝试解析为布尔值
	if value == "true" || value == "TRUE" {
		return true
	}
	if value == "false" || value == "FALSE" {
		return false
	}

	// 尝试解析为数字
	if intVal, err := parseInt(value); err == nil {
		return intVal
	}
	if floatVal, err := parseFloat(value); err == nil {
		return floatVal
	}

	// 默认为字符串
	return value
}

// parseInt 尝试解析为整数
func parseInt(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// parseFloat 尝试解析为浮点数
func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// EnvConfigMerger 环境变量配置合并器
type EnvConfigMerger struct {
	envLoader *EnvConfigLoader
}

// NewEnvConfigMerger 创建环境变量配置合并器
func NewEnvConfigMerger(prefix string) *EnvConfigMerger {
	return &EnvConfigMerger{
		envLoader: NewEnvConfigLoader(prefix),
	}
}

// Merge 合并环境变量配置到现有配置中
func (m *EnvConfigMerger) Merge(existingConfig map[string]interface{}) (map[string]interface{}, error) {
	envConfig, err := m.envLoader.Load()
	if err != nil {
		return nil, err
	}

	// 创建新的配置映射
	mergedConfig := make(map[string]interface{})

	// 先复制现有配置
	for k, v := range existingConfig {
		mergedConfig[k] = v
	}

	// 然后用环境变量配置覆盖（环境变量优先级更高）
	for k, v := range envConfig {
		mergedConfig[k] = v
	}

	return mergedConfig, nil
}
