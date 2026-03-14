package configuration

import (
	"os"
	"testing"
	"time"
)

// TestEnvConfigLoader 测试环境变量配置加载器
func TestEnvConfigLoader(t *testing.T) {
	// 设置测试环境变量
	_ = os.Setenv("TEST_APP_NAME", "EnvApp")
	_ = os.Setenv("TEST_SERVER_PORT", "9090")
	_ = os.Setenv("TEST_DEBUG_ENABLED", "true")
	defer func() {
		_ = os.Unsetenv("TEST_APP_NAME")
		_ = os.Unsetenv("TEST_SERVER_PORT")
		_ = os.Unsetenv("TEST_DEBUG_ENABLED")
	}()

	loader := NewEnvConfigLoader("TEST_")
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load env config: %v", err)
	}

	// 验证环境变量配置（现在支持嵌套结构）
	if appConfig, ok := config["app"].(map[string]any); ok {
		if appConfig["name"] != "EnvApp" {
			t.Errorf("Expected app.name 'EnvApp', got '%v'", appConfig["name"])
		}
	} else {
		t.Errorf("Expected app to be a map, got %T", config["app"])
	}

	// 环境变量解析器应该将数字字符串转换为数字类型
	if serverConfig, ok := config["server"].(map[string]any); ok {
		serverPort := serverConfig["port"]
		switch port := serverPort.(type) {
		case int64:
			if port != 9090 {
				t.Errorf("Expected server.port 9090, got '%v'", serverPort)
			}
		case string:
			if port != "9090" {
				t.Errorf("Expected server.port '9090', got '%v'", serverPort)
			}
		default:
			t.Errorf("Unexpected type for server.port: %T, value: %v", serverPort, serverPort)
		}
	} else {
		t.Errorf("Expected server to be a map, got %T", config["server"])
	}

	if debugConfig, ok := config["debug"].(map[string]any); ok {
		if debugConfig["enabled"] != true {
			t.Errorf("Expected debug.enabled true, got '%v'", debugConfig["enabled"])
		}
	} else {
		t.Errorf("Expected debug to be a map, got %T", config["debug"])
	}
}

// TestEventManager 测试事件管理器
func TestEventManager(t *testing.T) {
	eventManager := NewEventManager()

	eventReceived := make(chan bool, 1)
	var receivedEvent ConfigChangeEvent

	// 定义监听器函数
	var handler ConfigChangeHandler = func(event ConfigChangeEvent) {
		receivedEvent = event
		eventReceived <- true
	}

	// 注册监听器
	eventManager.RegisterGlobalWatcher("test.key", handler)

	// 触发事件
	eventManager.NotifyGlobalChange("test.key", "old_value", "new_value")

	// 等待事件
	select {
	case <-eventReceived:
		if receivedEvent.Key != "test.key" {
			t.Errorf("Expected event key 'test.key', got '%s'", receivedEvent.Key)
		}
		if receivedEvent.OldValue != "old_value" {
			t.Errorf("Expected old value 'old_value', got '%v'", receivedEvent.OldValue)
		}
		if receivedEvent.NewValue != "new_value" {
			t.Errorf("Expected new value 'new_value', got '%v'", receivedEvent.NewValue)
		}
	case <-time.After(time.Second * 1):
		t.Error("Timeout waiting for event")
	}

	eventManager.UnregisterGlobalWatcher("test.key", handler)

	if count := eventManager.GetGlobalWatcherCount("test.key"); count != 0 {
		t.Fatalf("Expected watcher count 0 after unregister, got %d", count)
	}

	receivedEvent = ConfigChangeEvent{}
	eventManager.NotifyGlobalChange("test.key", "value1", "value2")

	select {
	case <-eventReceived:
		t.Fatal("unexpected event after unregister")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestEventManager_NotifyUsesSnapshot(t *testing.T) {
	eventManager := NewEventManager()
	done := make(chan struct{}, 1)

	var handler ConfigChangeHandler
	handler = func(event ConfigChangeEvent) {
		eventManager.UnregisterGlobalWatcher("test.key", handler)
		done <- struct{}{}
	}

	eventManager.RegisterGlobalWatcher("test.key", handler)
	eventManager.NotifyGlobalChange("test.key", "old", "new")

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for watcher")
	}

	if count := eventManager.GetGlobalWatcherCount("test.key"); count != 0 {
		t.Fatalf("Expected watcher count 0 after self-unregister, got %d", count)
	}
}

// TestHelperFunctions 测试辅助函数
func TestHelperFunctions(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "config_test_helper")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// 创建测试配置文件
	if err := createTestConfigs(tempDir); err != nil {
		t.Fatalf("Failed to create test configs: %v", err)
	}

	// 初始化默认配置管理器
	if err := InitDefaultConfigManager(tempDir); err != nil {
		t.Fatalf("Failed to init default config manager: %v", err)
	}

	// 测试字符串辅助函数
	strVal, err := GetString("app_name")
	if err != nil {
		t.Errorf("GetString failed: %v", err)
	}
	if strVal != "Test Application" {
		t.Errorf("Expected 'Test Application', got '%s'", strVal)
	}

	strValWithDefault := GetStringWithDefault("nonexistent", "default")
	if strValWithDefault != "default" {
		t.Errorf("Expected 'default', got '%s'", strValWithDefault)
	}

	// 测试整数辅助函数
	intVal, err := GetInt("server.port")
	if err != nil {
		t.Errorf("GetInt failed: %v", err)
	}
	if intVal != 8080 {
		t.Errorf("Expected 8080, got %d", intVal)
	}

	intValWithDefault := GetIntWithDefault("nonexistent", 9999)
	if intValWithDefault != 9999 {
		t.Errorf("Expected 9999, got %d", intValWithDefault)
	}

	// 测试布尔辅助函数
	boolVal, err := GetBool("debug.enabled")
	if err != nil {
		t.Errorf("GetBool failed: %v", err)
	}
	if !boolVal {
		t.Errorf("Expected true, got %t", boolVal)
	}

	boolValWithDefault := GetBoolWithDefault("nonexistent", false)
	if boolValWithDefault {
		t.Errorf("Expected false, got %t", boolValWithDefault)
	}

	// 测试模块辅助函数
	moduleStr, err := GetModuleString("payment", "api_key")
	if err != nil {
		t.Errorf("GetModuleString failed: %v", err)
	}
	if moduleStr != "test_api_key_123" {
		t.Errorf("Expected 'test_api_key_123', got '%s'", moduleStr)
	}

	moduleStrWithDefault := GetModuleStringWithDefault("payment", "nonexistent", "default")
	if moduleStrWithDefault != "default" {
		t.Errorf("Expected 'default', got '%s'", moduleStrWithDefault)
	}
}

// BenchmarkConfigManager 性能测试
func BenchmarkConfigManager(b *testing.B) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "config_benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// 创建测试配置文件
	if err := createTestConfigs(tempDir); err != nil {
		b.Fatalf("Failed to create test configs: %v", err)
	}

	// 创建配置管理器
	options := &ConfigOptions{
		ConfigDir:       tempDir,
		WatchInterval:   time.Second * 1,
		EnableHotReload: false,
		Validator:       &defaultValidator{},
	}

	manager, err := NewConfigManager(options)
	if err != nil {
		b.Fatalf("Failed to create config manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	b.ResetTimer()

	// 测试全局配置获取性能
	b.Run("GetGlobalConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.Get("app_name")
			if err != nil {
				b.Errorf("Failed to get config: %v", err)
			}
		}
	})

	// 测试模块配置获取性能
	b.Run("GetModuleConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := manager.GetModuleConfig("payment", "api_key")
			if err != nil {
				b.Errorf("Failed to get module config: %v", err)
			}
		}
	})

	// 测试配置重新加载性能
	b.Run("ReloadConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := manager.Reload()
			if err != nil {
				b.Errorf("Failed to reload config: %v", err)
			}
		}
	})
}
