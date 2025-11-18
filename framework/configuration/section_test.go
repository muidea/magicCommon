package configuration

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// DatabaseDeclare 数据库配置结构体
type DatabaseDeclare struct {
	ID       int64  `json:"id"`
	DBServer string `json:"dbServer"`
	DBName   string `json:"dbName"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ApplicationInfo 应用信息结构体
type ApplicationInfo struct {
	UUID        string          `json:"uuid"`
	Name        string          `json:"name"`
	ShortName   string          `json:"shortName"`
	Icon        string          `json:"icon"`
	Version     string          `json:"version"`
	Domain      string          `json:"domain"`
	Email       string          `json:"email"`
	Author      string          `json:"author"`
	Description string          `json:"description"`
	Database    DatabaseDeclare `json:"database"`
}

// TestSectionConfig 测试section配置功能
func TestSectionConfig(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "section_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试配置文件
	configContent := `platformService = "http://127.0.0.1:8080"
casService = "http://127.0.0.1:8081"
fileService = "http://127.0.0.1:8083"
superNamespace = "super"

[applicationInfo]
uuid = "128ec90fc5e54340934954220de3d1e7"
name = "magicVMI"
shortName = "vmi"
icon = "magicVMI.png"
version = "1.3.0"
domain = "mulife.vip"
email = "rangh@mulife.vip"
author = "rangh"
description = "magic vmi"

[applicationInfo.database]
dbServer = "127.0.0.1:3306"
dbName = "magicvmi_db"
username = "magicvmi"
password = "magicvmi"

[server]
host = "localhost"
port = 8080
`

	configPath := filepath.Join(tempDir, "application.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
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

	// 测试获取applicationInfo section
	t.Run("GetApplicationInfoSection", func(t *testing.T) {
		var appInfo ApplicationInfo
		err := manager.GetSection("applicationInfo", &appInfo)
		if err != nil {
			t.Errorf("Failed to get applicationInfo section: %v", err)
		}

		// 验证应用信息
		if appInfo.UUID != "128ec90fc5e54340934954220de3d1e7" {
			t.Errorf("Expected UUID '128ec90fc5e54340934954220de3d1e7', got '%s'", appInfo.UUID)
		}
		if appInfo.Name != "magicVMI" {
			t.Errorf("Expected Name 'magicVMI', got '%s'", appInfo.Name)
		}
		if appInfo.ShortName != "vmi" {
			t.Errorf("Expected ShortName 'vmi', got '%s'", appInfo.ShortName)
		}
		if appInfo.Version != "1.3.0" {
			t.Errorf("Expected Version '1.3.0', got '%s'", appInfo.Version)
		}
		if appInfo.Domain != "mulife.vip" {
			t.Errorf("Expected Domain 'mulife.vip', got '%s'", appInfo.Domain)
		}
		if appInfo.Email != "rangh@mulife.vip" {
			t.Errorf("Expected Email 'rangh@mulife.vip', got '%s'", appInfo.Email)
		}
		if appInfo.Author != "rangh" {
			t.Errorf("Expected Author 'rangh', got '%s'", appInfo.Author)
		}
		if appInfo.Description != "magic vmi" {
			t.Errorf("Expected Description 'magic vmi', got '%s'", appInfo.Description)
		}

		// 验证嵌套的database配置
		if appInfo.Database.DBServer != "127.0.0.1:3306" {
			t.Errorf("Expected Database.DBServer '127.0.0.1:3306', got '%s'", appInfo.Database.DBServer)
		}
		if appInfo.Database.DBName != "magicvmi_db" {
			t.Errorf("Expected Database.DBName 'magicvmi_db', got '%s'", appInfo.Database.DBName)
		}
		if appInfo.Database.Username != "magicvmi" {
			t.Errorf("Expected Database.Username 'magicvmi', got '%s'", appInfo.Database.Username)
		}
		if appInfo.Database.Password != "magicvmi" {
			t.Errorf("Expected Database.Password 'magicvmi', got '%s'", appInfo.Database.Password)
		}
	})

	// 测试获取嵌套的database section
	t.Run("GetDatabaseSection", func(t *testing.T) {
		var database DatabaseDeclare
		err := manager.GetSection("applicationInfo.database", &database)
		if err != nil {
			t.Errorf("Failed to get applicationInfo.database section: %v", err)
		}

		// 验证数据库配置
		if database.DBServer != "127.0.0.1:3306" {
			t.Errorf("Expected DBServer '127.0.0.1:3306', got '%s'", database.DBServer)
		}
		if database.DBName != "magicvmi_db" {
			t.Errorf("Expected DBName 'magicvmi_db', got '%s'", database.DBName)
		}
		if database.Username != "magicvmi" {
			t.Errorf("Expected Username 'magicvmi', got '%s'", database.Username)
		}
		if database.Password != "magicvmi" {
			t.Errorf("Expected Password 'magicvmi', got '%s'", database.Password)
		}
	})

	// 测试获取server section
	t.Run("GetServerSection", func(t *testing.T) {
		type ServerConfig struct {
			Host string `json:"host"`
			Port int64  `json:"port"`
		}

		var server ServerConfig
		err := manager.GetSection("server", &server)
		if err != nil {
			t.Errorf("Failed to get server section: %v", err)
		}

		if server.Host != "localhost" {
			t.Errorf("Expected Host 'localhost', got '%s'", server.Host)
		}
		if server.Port != 8080 {
			t.Errorf("Expected Port 8080, got %d", server.Port)
		}
	})

	// 测试获取不存在的section
	t.Run("GetNonExistentSection", func(t *testing.T) {
		var config struct{}
		err := manager.GetSection("nonexistent.section", &config)
		if err == nil {
			t.Error("Expected error for nonexistent section, but got none")
		}
	})

	// 测试section监听功能
	t.Run("WatchSection", func(t *testing.T) {
		eventReceived := make(chan bool, 1)
		var receivedEvent ConfigChangeEvent

		// 注册section监听器
		err := manager.WatchSection("applicationInfo", func(event ConfigChangeEvent) {
			receivedEvent = event
			eventReceived <- true
		})
		if err != nil {
			t.Errorf("Failed to watch section: %v", err)
		}

		// 修改配置文件
		modifiedConfig := `platformService = "http://127.0.0.1:8080"
casService = "http://127.0.0.1:8081"
fileService = "http://127.0.0.1:8083"
superNamespace = "super"

[applicationInfo]
uuid = "modified_uuid_123456789"
name = "ModifiedApp"
shortName = "vmi"
icon = "magicVMI.png"
version = "1.3.0"
domain = "mulife.vip"
email = "rangh@mulife.vip"
author = "rangh"
description = "magic vmi"

[applicationInfo.database]
dbServer = "127.0.0.1:3306"
dbName = "magicvmi_db"
username = "magicvmi"
password = "magicvmi"

[server]
host = "localhost"
port = 8080
`

		if err := os.WriteFile(configPath, []byte(modifiedConfig), 0644); err != nil {
			t.Fatalf("Failed to write modified config: %v", err)
		}

		// 重新加载配置以触发事件
		err = manager.Reload()
		if err != nil {
			t.Errorf("Failed to reload config: %v", err)
		}

		// 等待事件
		select {
		case <-eventReceived:
			// 事件接收成功
			if receivedEvent.Key != "applicationInfo" {
				t.Errorf("Expected event key 'applicationInfo', got '%s'", receivedEvent.Key)
			}
		case <-time.After(time.Second * 2):
			t.Error("Timeout waiting for section change event")
		}
	})
}
