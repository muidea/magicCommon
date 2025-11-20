package configuration

import (
	"os"
	"testing"
)

// TestDebugMode 测试调试模式功能
func TestDebugMode(t *testing.T) {
	// 保存原始环境变量
	originalDebug := os.Getenv("DEBUG_MODE")
	defer os.Setenv("DEBUG_MODE", originalDebug)

	// 测试场景1: 启用调试模式
	t.Run("EnableDebugMode", func(t *testing.T) {
		os.Setenv("DEBUG_MODE", "true")

		// 创建临时配置目录
		tempDir := t.TempDir()
		options := &ConfigOptions{
			ConfigDir:       tempDir,
			EnableHotReload: false,
		}

		manager, err := NewConfigManager(options)
		if err != nil {
			t.Fatalf("Failed to create config manager: %v", err)
		}
		defer manager.Close()

		// 检查调试模式是否启用
		if !manager.IsDebugMode() {
			t.Error("Debug mode should be enabled when DEBUG_MODE=true")
		}

		// 测试获取不存在的配置项，应该输出调试信息
		_, err = manager.Get("nonexistent.key")
		if err == nil {
			t.Error("Should return error for nonexistent key")
		}
		// 注意：这里我们无法直接捕获fmt.Printf的输出，但可以确认方法执行了
	})

	// 测试场景2: 禁用调试模式
	t.Run("DisableDebugMode", func(t *testing.T) {
		os.Setenv("DEBUG_MODE", "false")

		// 创建临时配置目录
		tempDir := t.TempDir()
		options := &ConfigOptions{
			ConfigDir:       tempDir,
			EnableHotReload: false,
		}

		manager, err := NewConfigManager(options)
		if err != nil {
			t.Fatalf("Failed to create config manager: %v", err)
		}
		defer manager.Close()

		// 检查调试模式是否禁用
		if manager.IsDebugMode() {
			t.Error("Debug mode should be disabled when DEBUG_MODE=false")
		}
	})

	// 测试场景3: 使用数字1启用调试模式
	t.Run("EnableDebugModeWith1", func(t *testing.T) {
		os.Setenv("DEBUG_MODE", "1")

		// 创建临时配置目录
		tempDir := t.TempDir()
		options := &ConfigOptions{
			ConfigDir:       tempDir,
			EnableHotReload: false,
		}

		manager, err := NewConfigManager(options)
		if err != nil {
			t.Fatalf("Failed to create config manager: %v", err)
		}
		defer manager.Close()

		// 检查调试模式是否启用
		if !manager.IsDebugMode() {
			t.Error("Debug mode should be enabled when DEBUG_MODE=1")
		}
	})

	// 测试场景4: 动态设置调试模式
	t.Run("DynamicDebugMode", func(t *testing.T) {
		os.Setenv("DEBUG_MODE", "false")

		// 创建临时配置目录
		tempDir := t.TempDir()
		options := &ConfigOptions{
			ConfigDir:       tempDir,
			EnableHotReload: false,
		}

		manager, err := NewConfigManager(options)
		if err != nil {
			t.Fatalf("Failed to create config manager: %v", err)
		}
		defer manager.Close()

		// 初始状态应该是禁用
		if manager.IsDebugMode() {
			t.Error("Debug mode should be disabled initially")
		}

		// 动态启用调试模式
		manager.SetDebugMode(true)
		if !manager.IsDebugMode() {
			t.Error("Debug mode should be enabled after SetDebugMode(true)")
		}

		// 动态禁用调试模式
		manager.SetDebugMode(false)
		if manager.IsDebugMode() {
			t.Error("Debug mode should be disabled after SetDebugMode(false)")
		}
	})
}

// TestDebugModeModuleConfig 测试模块配置的调试模式
func TestDebugModeModuleConfig(t *testing.T) {
	// 保存原始环境变量
	originalDebug := os.Getenv("DEBUG_MODE")
	defer os.Setenv("DEBUG_MODE", originalDebug)

	os.Setenv("DEBUG_MODE", "true")

	// 创建临时配置目录
	tempDir := t.TempDir()
	options := &ConfigOptions{
		ConfigDir:       tempDir,
		EnableHotReload: false,
	}

	manager, err := NewConfigManager(options)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	defer manager.Close()

	// 测试获取不存在的模块配置项，应该输出调试信息
	_, err = manager.GetModuleConfig("nonexistent.module", "nonexistent.key")
	if err == nil {
		t.Error("Should return error for nonexistent module and key")
	}
}
