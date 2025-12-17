package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/muidea/magicCommon/framework/configuration"
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

func main() {
	// 初始化默认配置管理器
	err := configuration.InitDefaultConfigManager("./config")
	if err != nil {
		log.Fatalf("Failed to initialize config manager: %v", err)
	}
	defer configuration.CloseConfigManager()

	fmt.Println("=== 配置管理框架使用示例 ===")

	// 1. 基本配置获取
	fmt.Println("\n1. 基本配置获取:")
	appName, err := configuration.GetString("app_name")
	if err != nil {
		log.Printf("Failed to get app_name: %v", err)
	} else {
		fmt.Printf("应用名称: %s\n", appName)
	}

	serverPort, err := configuration.GetInt("server.port")
	if err != nil {
		log.Printf("Failed to get server.port: %v", err)
	} else {
		fmt.Printf("服务器端口: %d\n", serverPort)
	}

	debugEnabled, err := configuration.GetBool("debug.enabled")
	if err != nil {
		log.Printf("Failed to get debug.enabled: %v", err)
	} else {
		fmt.Printf("调试模式: %t\n", debugEnabled)
	}

	// 2. 带默认值的配置获取
	fmt.Println("\n2. 带默认值的配置获取:")
	defaultValue := configuration.GetStringWithDefault("nonexistent.key", "default_value")
	fmt.Printf("不存在的配置项: %s\n", defaultValue)

	// 3. 模块配置获取
	fmt.Println("\n3. 模块配置获取:")
	apiKey, err := configuration.GetModuleString("payment", "api_key")
	if err != nil {
		log.Printf("Failed to get payment.api_key: %v", err)
	} else {
		fmt.Printf("支付API密钥: %s\n", apiKey)
	}

	// 4. Section配置获取
	fmt.Println("\n4. Section配置获取:")

	// 获取applicationInfo section
	var appInfo ApplicationInfo
	err = configuration.GetSection("applicationInfo", &appInfo)
	if err != nil {
		log.Printf("Failed to get applicationInfo section: %v", err)
	} else {
		fmt.Printf("应用信息:\n")
		fmt.Printf("  UUID: %s\n", appInfo.UUID)
		fmt.Printf("  名称: %s\n", appInfo.Name)
		fmt.Printf("  简称: %s\n", appInfo.ShortName)
		fmt.Printf("  版本: %s\n", appInfo.Version)
		fmt.Printf("  域名: %s\n", appInfo.Domain)
		fmt.Printf("  邮箱: %s\n", appInfo.Email)
		fmt.Printf("  作者: %s\n", appInfo.Author)
		fmt.Printf("  描述: %s\n", appInfo.Description)
	}

	// 获取嵌套的database section
	var database DatabaseDeclare
	err = configuration.GetSection("applicationInfo.database", &database)
	if err != nil {
		log.Printf("Failed to get applicationInfo.database section: %v", err)
	} else {
		fmt.Printf("数据库配置:\n")
		fmt.Printf("  服务器: %s\n", database.DBServer)
		fmt.Printf("  数据库: %s\n", database.DBName)
		fmt.Printf("  用户名: %s\n", database.Username)
		fmt.Printf("  密码: %s\n", database.Password)
	}

	// 5. 配置监听
	fmt.Println("\n5. 配置监听示例:")

	// 监听全局配置变更
	err = configuration.WatchConfig("app_name", func(event configuration.ConfigChangeEvent) {
		fmt.Printf("配置变更事件 - 键: %s, 旧值: %v, 新值: %v\n",
			event.Key, event.OldValue, event.NewValue)
	})
	if err != nil {
		log.Printf("Failed to watch config: %v", err)
	}

	// 监听section配置变更
	err = configuration.WatchSection("applicationInfo", func(event configuration.ConfigChangeEvent) {
		fmt.Printf("Section变更事件 - 键: %s, 旧值: %v, 新值: %v\n",
			event.Key, event.OldValue, event.NewValue)
	})
	if err != nil {
		log.Printf("Failed to watch section: %v", err)
	}

	// 6. 其他辅助函数
	fmt.Println("\n6. 其他辅助函数:")

	// 检查配置管理器是否已初始化
	if configuration.IsConfigManagerInitialized() {
		fmt.Println("配置管理器已初始化")
	}

	// 获取浮点数配置
	floatValue, err := configuration.GetFloat64("server.port")
	if err != nil {
		log.Printf("Failed to get float value: %v", err)
	} else {
		fmt.Printf("服务器端口(浮点数): %.1f\n", floatValue)
	}

	// 获取模块布尔配置
	creditCardEnabled, err := configuration.GetModuleBool("payment", "methods.credit_card")
	if err != nil {
		log.Printf("Failed to get payment.methods.credit_card: %v", err)
	} else {
		fmt.Printf("信用卡支付启用: %t\n", creditCardEnabled)
	}

	// 7. 重新加载配置
	fmt.Println("\n7. 重新加载配置:")
	err = configuration.ReloadConfig()
	if err != nil {
		log.Printf("Failed to reload config: %v", err)
	} else {
		fmt.Println("配置重新加载成功")
	}

	// 8. 导出所有配置
	fmt.Println("\n8. 导出所有配置:")
	allConfigs, err := configuration.ExportAllConfigs()
	if err != nil {
		log.Printf("Failed to export all configs: %v", err)
	} else {
		// 将配置转换为JSON格式输出
		configJSON, err := json.MarshalIndent(allConfigs, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal configs to JSON: %v", err)
		} else {
			fmt.Println("所有配置项(JSON格式):")
			fmt.Println(string(configJSON))
		}

		// 展示配置结构
		fmt.Println("\n配置结构概览:")
		if appConfig, ok := allConfigs["application"].(map[string]any); ok {
			fmt.Printf("应用程序配置项数量: %d\n", len(appConfig))
			for key := range appConfig {
				fmt.Printf("  - %s\n", key)
			}
		}

		if modulesConfig, ok := allConfigs["modules"].(map[string]any); ok {
			fmt.Printf("模块配置数量: %d\n", len(modulesConfig))
			for moduleName := range modulesConfig {
				fmt.Printf("  - %s\n", moduleName)
			}
		}
	}

	// 9. 环境变量配置加载器演示
	fmt.Println("\n9. 环境变量配置加载器演示:")
	DemoEnvLoader()

	fmt.Println("\n=== 示例完成 ===")
}
