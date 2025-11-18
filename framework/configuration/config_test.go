package configuration

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestConfigManager 测试配置管理器
func TestConfigManager(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试配置文件
	if err := createTestConfigs(tempDir); err != nil {
		t.Fatalf("Failed to create test configs: %v", err)
	}

	// 测试配置管理器创建
	options := &ConfigOptions{
		ConfigDir:       tempDir,
		WatchInterval:   time.Second * 1,
		EnableHotReload: false, // 测试中禁用热加载
		Validator:       &defaultValidator{},
	}

	manager, err := NewConfigManager(options)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer manager.Close()

	// 测试全局配置获取
	t.Run("GlobalConfig", func(t *testing.T) {
		testGlobalConfig(t, manager)
	})

	// 测试模块配置获取
	t.Run("ModuleConfig", func(t *testing.T) {
		testModuleConfig(t, manager)
	})

	// 测试配置监听
	t.Run("ConfigWatching", func(t *testing.T) {
		testConfigWatching(t, manager, tempDir)
	})

	// 测试配置重新加载
	t.Run("ConfigReload", func(t *testing.T) {
		testConfigReload(t, manager, tempDir)
	})
}

// testGlobalConfig 测试全局配置
func testGlobalConfig(t *testing.T, manager ConfigManager) {
	// 测试获取字符串配置
	appName, err := manager.Get("app_name")
	if err != nil {
		t.Errorf("Failed to get app_name: %v", err)
	}
	if appName != "Test Application" {
		t.Errorf("Expected app_name 'Test Application', got '%v'", appName)
	}

	// 测试获取嵌套配置
	serverPort, err := manager.Get("server.port")
	if err != nil {
		t.Errorf("Failed to get server.port: %v", err)
	}
	// TOML解析后数字可能是int64或float64
	switch port := serverPort.(type) {
	case int64:
		if port != 8080 {
			t.Errorf("Expected server.port 8080, got '%v'", serverPort)
		}
	case float64:
		if port != 8080 {
			t.Errorf("Expected server.port 8080, got '%v'", serverPort)
		}
	default:
		t.Errorf("Unexpected type for server.port: %T, value: %v", serverPort, serverPort)
	}

	// 测试获取带默认值的配置
	defaultValue := manager.GetWithDefault("nonexistent.key", "default_value")
	if defaultValue != "default_value" {
		t.Errorf("Expected default value 'default_value', got '%v'", defaultValue)
	}

	// 测试获取布尔配置
	debugEnabled, err := manager.Get("debug.enabled")
	if err != nil {
		t.Errorf("Failed to get debug.enabled: %v", err)
	}
	if debugEnabled != true {
		t.Errorf("Expected debug.enabled true, got '%v'", debugEnabled)
	}
}

// testModuleConfig 测试模块配置
func testModuleConfig(t *testing.T, manager ConfigManager) {
	// 测试获取模块配置
	apiKey, err := manager.GetModuleConfig("payment", "api_key")
	if err != nil {
		t.Errorf("Failed to get payment.api_key: %v", err)
	}
	if apiKey != "test_api_key_123" {
		t.Errorf("Expected payment.api_key 'test_api_key_123', got '%v'", apiKey)
	}

	// 测试获取模块嵌套配置
	gatewayURL, err := manager.GetModuleConfig("payment", "gateway.url")
	if err != nil {
		t.Errorf("Failed to get payment.gateway.url: %v", err)
	}
	if gatewayURL != "https://api.payment.test/v1" {
		t.Errorf("Expected payment.gateway.url 'https://api.payment.test/v1', got '%v'", gatewayURL)
	}

	// 测试获取不存在的模块配置
	_, err = manager.GetModuleConfig("nonexistent", "key")
	if err == nil {
		t.Error("Expected error for nonexistent module")
	}

	// 测试获取模块配置带默认值
	defaultValue := manager.GetModuleConfigWithDefault("payment", "nonexistent.key", "default")
	if defaultValue != "default" {
		t.Errorf("Expected default value 'default', got '%v'", defaultValue)
	}
}

// testConfigWatching 测试配置监听
func testConfigWatching(t *testing.T, manager ConfigManager, tempDir string) {
	eventReceived := make(chan bool, 1)
	var receivedEvent ConfigChangeEvent

	// 注册全局配置监听器
	err := manager.Watch("app_name", func(event ConfigChangeEvent) {
		receivedEvent = event
		eventReceived <- true
	})
	if err != nil {
		t.Errorf("Failed to watch config: %v", err)
	}

	// 注册模块配置监听器
	moduleEventReceived := make(chan bool, 1)
	var receivedModuleEvent ConfigChangeEvent

	err = manager.WatchModule("payment", "api_key", func(event ConfigChangeEvent) {
		receivedModuleEvent = event
		moduleEventReceived <- true
	})
	if err != nil {
		t.Errorf("Failed to watch module config: %v", err)
	}

	// 修改配置文件以触发配置变更事件
	globalConfigPath := filepath.Join(tempDir, "application.toml")
	modifiedConfig := `app_name = "Modified Application"
version = "1.0.0"

[server]
host = "localhost"
port = 8080

[database]
host = "localhost"
port = 5432

[debug]
enabled = true
`

	if err := os.WriteFile(globalConfigPath, []byte(modifiedConfig), 0644); err != nil {
		t.Fatalf("Failed to write modified config: %v", err)
	}

	// 重新加载配置以触发事件
	err = manager.Reload()
	if err != nil {
		t.Errorf("Failed to reload config: %v", err)
	}

	// 等待事件（使用超时）
	select {
	case <-eventReceived:
		// 事件接收成功
		if receivedEvent.Key != "app_name" {
			t.Errorf("Expected event key 'app_name', got '%s'", receivedEvent.Key)
		}
		if receivedEvent.NewValue != "Modified Application" {
			t.Errorf("Expected new value 'Modified Application', got '%v'", receivedEvent.NewValue)
		}
	case <-time.After(time.Second * 2):
		t.Error("Timeout waiting for config change event")
	}

	// 修改模块配置文件
	moduleConfigPath := filepath.Join(tempDir, "config.d", "payment.toml")
	modifiedModuleConfig := `api_key = "modified_api_key_789"
secret_key = "test_secret_key_456"

[gateway]
url = "https://api.payment.test/v1"
timeout = "30s"

[methods]
credit_card = true
paypal = false
`

	if err := os.WriteFile(moduleConfigPath, []byte(modifiedModuleConfig), 0644); err != nil {
		t.Fatalf("Failed to write modified module config: %v", err)
	}

	// 重新加载配置以触发模块事件
	err = manager.Reload()
	if err != nil {
		t.Errorf("Failed to reload config: %v", err)
	}

	select {
	case <-moduleEventReceived:
		// 事件接收成功
		if receivedModuleEvent.Key != "payment.api_key" {
			t.Errorf("Expected event key 'payment.api_key', got '%s'", receivedModuleEvent.Key)
		}
		if receivedModuleEvent.NewValue != "modified_api_key_789" {
			t.Errorf("Expected new value 'modified_api_key_789', got '%v'", receivedModuleEvent.NewValue)
		}
	case <-time.After(time.Second * 2):
		t.Error("Timeout waiting for module config change event")
	}
}

// testConfigReload 测试配置重新加载
func testConfigReload(t *testing.T, manager ConfigManager, tempDir string) {
	// 修改配置文件
	globalConfigPath := filepath.Join(tempDir, "application.toml")
	newConfig := `app_name = "Reloaded Application"
version = "2.0.0"

[server]
host = "127.0.0.1"
port = 9090

[debug]
enabled = false
`

	if err := os.WriteFile(globalConfigPath, []byte(newConfig), 0644); err != nil {
		t.Fatalf("Failed to write updated config: %v", err)
	}

	// 重新加载配置
	if err := manager.Reload(); err != nil {
		t.Errorf("Failed to reload config: %v", err)
	}

	// 验证配置已更新
	appName, err := manager.Get("app_name")
	if err != nil {
		t.Errorf("Failed to get app_name after reload: %v", err)
	}
	if appName != "Reloaded Application" {
		t.Errorf("Expected app_name 'Reloaded Application' after reload, got '%v'", appName)
	}

	debugEnabled, err := manager.Get("debug.enabled")
	if err != nil {
		t.Errorf("Failed to get debug.enabled after reload: %v", err)
	}
	if debugEnabled != false {
		t.Errorf("Expected debug.enabled false after reload, got '%v'", debugEnabled)
	}
}

// createTestConfigs 创建测试配置文件
func createTestConfigs(tempDir string) error {
	// 创建全局配置文件
	globalConfig := `app_name = "Test Application"
version = "1.0.0"

[server]
host = "localhost"
port = 8080

[database]
host = "localhost"
port = 5432

[debug]
enabled = true
`

	globalConfigPath := filepath.Join(tempDir, "application.toml")
	if err := os.WriteFile(globalConfigPath, []byte(globalConfig), 0644); err != nil {
		return err
	}

	// 创建模块配置目录
	moduleDir := filepath.Join(tempDir, "config.d")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return err
	}

	// 创建支付模块配置文件
	paymentConfig := `api_key = "test_api_key_123"
secret_key = "test_secret_key_456"

[gateway]
url = "https://api.payment.test/v1"
timeout = "30s"

[methods]
credit_card = true
paypal = false
`

	paymentConfigPath := filepath.Join(moduleDir, "payment.toml")
	if err := os.WriteFile(paymentConfigPath, []byte(paymentConfig), 0644); err != nil {
		return err
	}

	// 创建认证模块配置文件
	authConfig := `secret = "test_jwt_secret"
issuer = "test.app"

[jwt]
algorithm = "HS256"
expiration = "24h"
`

	authConfigPath := filepath.Join(moduleDir, "auth.toml")
	if err := os.WriteFile(authConfigPath, []byte(authConfig), 0644); err != nil {
		return err
	}

	return nil
}
