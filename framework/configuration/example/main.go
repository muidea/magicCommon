package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/muidea/magicCommon/framework/configuration"
)

func main() {
	// 示例1: 基本使用
	basicExample()

	// 示例2: 模块配置使用
	moduleExample()

	// 示例3: 配置监听使用
	watchExample()
}

// basicExample 基本使用示例
func basicExample() {
	fmt.Println("=== 基本使用示例 ===")

	// 初始化配置管理器
	configDir := getExampleConfigDir()
	if err := configuration.InitDefaultConfigManager(configDir); err != nil {
		log.Fatalf("Failed to initialize config manager: %v", err)
	}

	// 获取全局配置
	serverPort, err := configuration.GetInt("server.port")
	if err != nil {
		fmt.Printf("Error getting server.port: %v\n", err)
	} else {
		fmt.Printf("Server port: %d\n", serverPort)
	}

	// 获取带默认值的配置
	databaseHost := configuration.GetStringWithDefault("database.host", "localhost")
	fmt.Printf("Database host: %s\n", databaseHost)

	// 获取布尔配置
	debugMode := configuration.GetBoolWithDefault("debug.enabled", false)
	fmt.Printf("Debug mode: %t\n", debugMode)

	fmt.Println()
}

// moduleExample 模块配置使用示例
func moduleExample() {
	fmt.Println("=== 模块配置使用示例 ===")

	// 获取模块配置
	moduleName := "payment"
	apiKey, err := configuration.GetModuleString(moduleName, "api_key")
	if err != nil {
		fmt.Printf("Error getting payment.api_key: %v\n", err)
	} else {
		fmt.Printf("Payment API key: %s\n", apiKey)
	}

	// 获取带默认值的模块配置
	timeout := configuration.GetModuleStringWithDefault(moduleName, "timeout", "30s")
	fmt.Printf("Payment timeout: %s\n", timeout)

	fmt.Println()
}

// watchExample 配置监听使用示例
func watchExample() {
	fmt.Println("=== 配置监听使用示例 ===")

	// 监听全局配置变更
	err := configuration.WatchConfig("server.port", func(event configuration.ConfigChangeEvent) {
		fmt.Printf("Config changed: %s, old: %v, new: %v\n",
			event.Key, event.OldValue, event.NewValue)
	})
	if err != nil {
		fmt.Printf("Error watching config: %v\n", err)
	}

	// 监听模块配置变更
	err = configuration.WatchModuleConfig("payment", "api_key", func(event configuration.ConfigChangeEvent) {
		fmt.Printf("Module config changed: %s, old: %v, new: %v\n",
			event.Key, event.OldValue, event.NewValue)
	})
	if err != nil {
		fmt.Printf("Error watching module config: %v\n", err)
	}

	fmt.Println("Configuration watchers registered. Try modifying config files to see changes.")
	fmt.Println()
}

// getExampleConfigDir 获取示例配置目录
func getExampleConfigDir() string {
	// 获取当前文件所在目录
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	// 返回示例配置目录
	return filepath.Join(dir, "..", "..", "..", "..", "config")
}

// advancedExample 高级使用示例
func advancedExample() {
	fmt.Println("=== 高级使用示例 ===")

	// 创建自定义配置管理器
	configDir := getExampleConfigDir()
	manager, err := configuration.CreateConfigManagerWithDir(configDir, true)
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}
	defer manager.Close()

	// 直接使用管理器接口
	value, err := manager.Get("server.host")
	if err != nil {
		fmt.Printf("Error getting server.host: %v\n", err)
	} else {
		fmt.Printf("Server host: %v\n", value)
	}

	// 获取模块配置
	moduleValue, err := manager.GetModuleConfig("auth", "secret")
	if err != nil {
		fmt.Printf("Error getting auth.secret: %v\n", err)
	} else {
		fmt.Printf("Auth secret: %v\n", moduleValue)
	}

	fmt.Println()
}
