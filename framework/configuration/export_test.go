package configuration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestExportAllConfigs 测试导出所有配置功能
func TestExportAllConfigs(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "export_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试配置文件
	configContent := `app_name = "Test Application"
version = "1.0.0"

[server]
host = "localhost"
port = 8080

[database]
host = "localhost"
port = 5432

[debug]
enabled = true

[applicationInfo]
uuid = "128ec90fc5e54340934954220de3d1e7"
name = "magicVMI"
shortName = "vmi"

[applicationInfo.database]
dbServer = "127.0.0.1:3306"
dbName = "magicvmi_db"
username = "magicvmi"
password = "magicvmi"
`

	configPath := filepath.Join(tempDir, "application.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// 创建模块配置文件
	moduleDir := filepath.Join(tempDir, "config.d")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("Failed to create module directory: %v", err)
	}

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
		t.Fatalf("Failed to write payment config: %v", err)
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
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer manager.Close()

	// 测试导出所有配置
	t.Run("ExportAllConfigs", func(t *testing.T) {
		exportedConfigs, err := manager.ExportAllConfigs()
		if err != nil {
			t.Fatalf("Failed to export all configs: %v", err)
		}

		// 验证导出结果结构
		if exportedConfigs == nil {
			t.Error("Exported configs should not be nil")
		}

		// 验证应用程序配置
		appConfig, exists := exportedConfigs["application"]
		if !exists {
			t.Error("Exported configs should contain 'application' section")
		}

		appMap, ok := appConfig.(map[string]any)
		if !ok {
			t.Error("Application config should be a map")
		}

		// 验证应用程序配置项
		if appMap["app_name"] != "Test Application" {
			t.Errorf("Expected app_name 'Test Application', got '%v'", appMap["app_name"])
		}

		if appMap["version"] != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%v'", appMap["version"])
		}

		// 验证嵌套配置
		serverConfig, exists := appMap["server"]
		if !exists {
			t.Error("Application config should contain 'server' section")
		}

		serverMap, ok := serverConfig.(map[string]any)
		if !ok {
			t.Error("Server config should be a map")
		}

		if serverMap["host"] != "localhost" {
			t.Errorf("Expected server.host 'localhost', got '%v'", serverMap["host"])
		}

		if serverMap["port"] != int64(8080) {
			t.Errorf("Expected server.port 8080, got '%v'", serverMap["port"])
		}

		// 验证模块配置
		modulesConfig, exists := exportedConfigs["modules"]
		if !exists {
			t.Error("Exported configs should contain 'modules' section")
		}

		modulesMap, ok := modulesConfig.(map[string]any)
		if !ok {
			t.Error("Modules config should be a map")
		}

		paymentConfig, exists := modulesMap["payment"]
		if !exists {
			t.Error("Modules should contain 'payment' module")
		}

		paymentMap, ok := paymentConfig.(map[string]any)
		if !ok {
			t.Error("Payment config should be a map")
		}

		if paymentMap["api_key"] != "test_api_key_123" {
			t.Errorf("Expected payment.api_key 'test_api_key_123', got '%v'", paymentMap["api_key"])
		}

		// 验证嵌套的模块配置
		gatewayConfig, exists := paymentMap["gateway"]
		if !exists {
			t.Error("Payment config should contain 'gateway' section")
		}

		gatewayMap, ok := gatewayConfig.(map[string]any)
		if !ok {
			t.Error("Gateway config should be a map")
		}

		if gatewayMap["url"] != "https://api.payment.test/v1" {
			t.Errorf("Expected payment.gateway.url 'https://api.payment.test/v1', got '%v'", gatewayMap["url"])
		}

		// 验证配置可以被序列化为JSON
		jsonBytes, err := json.Marshal(exportedConfigs)
		if err != nil {
			t.Fatalf("Failed to marshal exported configs to JSON: %v", err)
		}

		// 验证JSON内容
		var jsonConfig map[string]any
		if err := json.Unmarshal(jsonBytes, &jsonConfig); err != nil {
			t.Fatalf("Failed to unmarshal JSON config: %v", err)
		}

		// 验证JSON结构
		if _, exists := jsonConfig["application"]; !exists {
			t.Error("JSON config should contain 'application' section")
		}

		if _, exists := jsonConfig["modules"]; !exists {
			t.Error("JSON config should contain 'modules' section")
		}
	})

	// 测试使用默认配置管理器导出
	t.Run("ExportWithDefaultManager", func(t *testing.T) {
		// 初始化默认配置管理器
		err := InitDefaultConfigManager(tempDir)
		if err != nil {
			t.Fatalf("Failed to init default config manager: %v", err)
		}
		defer CloseConfigManager()

		exportedConfigs, err := ExportAllConfigs()
		if err != nil {
			t.Fatalf("Failed to export all configs with default manager: %v", err)
		}

		if exportedConfigs == nil {
			t.Error("Exported configs should not be nil")
		}

		// 验证基本结构
		if _, exists := exportedConfigs["application"]; !exists {
			t.Error("Exported configs should contain 'application' section")
		}

		if _, exists := exportedConfigs["modules"]; !exists {
			t.Error("Exported configs should contain 'modules' section")
		}
	})

	// 测试未初始化时的错误处理
	t.Run("ExportWithoutInitialization", func(t *testing.T) {
		// 重置默认配置管理器
		DefaultConfigManager = nil

		_, err := ExportAllConfigs()
		if err == nil {
			t.Error("Expected error when exporting without initialization")
		}
	})
}
