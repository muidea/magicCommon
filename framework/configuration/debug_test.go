package configuration

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// TestDebugConfig 调试配置解析
func TestDebugConfig(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "config_debug")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试配置文件
	if err := createTestConfigs(tempDir); err != nil {
		t.Fatalf("Failed to create test configs: %v", err)
	}

	// 创建配置管理器
	options := &ConfigOptions{
		ConfigDir:       tempDir,
		WatchInterval:   time.Second * 1,
		EnableHotReload: false,
		Validator:       &defaultValidator{},
	}

	configManager, err := NewConfigManager(options)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer configManager.Close()

	// 调试全局配置
	fmt.Println("=== Global Config Keys ===")
	globalConfig := configManager.GetGlobalConfig()
	for key, value := range globalConfig {
		fmt.Printf("Key: %s, Type: %T, Value: %v\n", key, value, value)
	}

	// 调试模块配置
	fmt.Println("\n=== Module Config Keys ===")
	modules := configManager.GetModuleNames()
	for _, module := range modules {
		fmt.Printf("Module: %s\n", module)
		// 这里需要访问内部字段来获取模块配置
	}
}
