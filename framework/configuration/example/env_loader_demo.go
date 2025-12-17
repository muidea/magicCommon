package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/muidea/magicCommon/framework/configuration"
)

// getNestedValue 从嵌套配置中获取值
func getNestedValue(config map[string]any, key string) any {
	parts := strings.Split(key, ".")
	current := config

	for i, part := range parts {
		if val, ok := current[part]; ok {
			if i == len(parts)-1 {
				return val
			}
			if nested, ok := val.(map[string]any); ok {
				current = nested
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
	return nil
}

// DemoEnvLoader 环境变量加载器演示函数
func DemoEnvLoader() {
	fmt.Println("=== 环境变量加载器演示程序 ===")

	// 设置测试环境变量
	fmt.Println("\n1. 设置测试环境变量:")
	envVars := map[string]string{
		"DEFAULT_NAMESPACE":      "panel",
		"DEBUG_MODE":             "true",
		"WORKSPACE_ROOT_PATH":    "/home/rangh/dataspace",
		"APPS_PLATFORM_SERVICE":  "http://magicplatform:8080",
		"APPS_CAS_SERVICE":       "http://magiccas:8080",
		"APPS_FILE_SERVICE":      "http://magicfile:8080",
		"TEST_PREFIX_CUSTOM_KEY": "prefixed-value",
		"NUMERIC_INT_VALUE":      "123",
		"NUMERIC_FLOAT_VALUE":    "3.14",
		"BOOLEAN_FALSE_VALUE":    "false",
		// 添加嵌套结构测试环境变量
		"DATABASE_HOST":                 "localhost",
		"DATABASE_PORT":                 "3306",
		"DATABASE_CREDENTIALS_USERNAME": "admin",
		"DATABASE_CREDENTIALS_PASSWORD": "secret",
		"SERVER_SETTINGS_HTTP_PORT":     "8080",
		"SERVER_SETTINGS_HTTPS_PORT":    "8443",
		"APP_INFO_NAME":                 "TestApp",
		"APP_INFO_VERSION":              "1.0.0",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		fmt.Printf("  设置: %s=%s\n", key, value)
	}

	defer func() {
		// 清理环境变量
		for key := range envVars {
			os.Unsetenv(key)
		}
	}()

	// 2. 无前缀加载器测试
	fmt.Println("\n2. 无前缀环境变量加载器测试:")
	noPrefixLoader := configuration.NewEnvConfigLoader("")
	noPrefixConfig, err := noPrefixLoader.Load()
	if err != nil {
		fmt.Printf("   加载失败: %v\n", err)
		return
	}

	fmt.Println("   加载的配置项:")
	for key, value := range noPrefixConfig {
		fmt.Printf("     %s = %v (%T)\n", key, value, value)
	}

	// 3. 检查特定键的转换
	fmt.Println("\n3. 环境变量键名转换验证:")
	expectedMappings := map[string]string{
		"DEFAULT_NAMESPACE":     "default.namespace",
		"DEBUG_MODE":            "debug.mode",
		"WORKSPACE_ROOT_PATH":   "workspace.root.path",
		"APPS_PLATFORM_SERVICE": "apps.platform.service",
		"APPS_CAS_SERVICE":      "apps.cas.service",
		"APPS_FILE_SERVICE":     "apps.file.service",
		"NUMERIC_INT_VALUE":     "numeric.int.value",
		"NUMERIC_FLOAT_VALUE":   "numeric.float.value",
		"BOOLEAN_FALSE_VALUE":   "boolean.false.value",
		// 嵌套结构键名映射
		"DATABASE_HOST":                 "database.host",
		"DATABASE_PORT":                 "database.port",
		"DATABASE_CREDENTIALS_USERNAME": "database.credentials.username",
		"DATABASE_CREDENTIALS_PASSWORD": "database.credentials.password",
		"SERVER_SETTINGS_HTTP_PORT":     "server.settings.http.port",
		"SERVER_SETTINGS_HTTPS_PORT":    "server.settings.https.port",
		"APP_INFO_NAME":                 "app.info.name",
		"APP_INFO_VERSION":              "app.info.version",
	}

	for envKey, configKey := range expectedMappings {
		val := getNestedValue(noPrefixConfig, configKey)
		if val != nil {
			fmt.Printf("   ✅ %s -> %s = %v\n", envKey, configKey, val)
		} else {
			fmt.Printf("   ❌ %s -> %s 未找到\n", envKey, configKey)
		}
	}

	// 4. 带前缀加载器测试
	fmt.Println("\n4. 带前缀环境变量加载器测试 (前缀: TEST_PREFIX_):")
	prefixLoader := configuration.NewEnvConfigLoader("TEST_PREFIX_")
	prefixConfig, err := prefixLoader.Load()
	if err != nil {
		fmt.Printf("   加载失败: %v\n", err)
		return
	}

	fmt.Println("   加载的配置项:")
	for key, value := range prefixConfig {
		fmt.Printf("     %s = %v\n", key, value)
	}

	// 5. 配置合并器测试
	fmt.Println("\n5. 配置合并器测试:")
	existingFileConfig := map[string]any{
		"default.namespace": "file-config-value",
		"debug.mode":        false,
		"file.only.key":     "only-in-file",
		"numeric.int.value": 999, // 文件中的值
	}

	fmt.Println("   现有配置文件:")
	for key, value := range existingFileConfig {
		fmt.Printf("     %s = %v\n", key, value)
	}

	merger := configuration.NewEnvConfigMerger("")
	mergedConfig, err := merger.Merge(existingFileConfig)
	if err != nil {
		fmt.Printf("   合并失败: %v\n", err)
		return
	}

	fmt.Println("\n   合并后的配置 (环境变量优先级更高):")
	for key, value := range mergedConfig {
		fmt.Printf("     %s = %v\n", key, value)
	}

	// 6. 数据类型解析验证
	fmt.Println("\n6. 数据类型解析验证:")
	typeTestCases := []struct {
		key          string
		expectedType string
	}{
		{"debug.mode", "bool"},
		{"boolean.false.value", "bool"},
		{"numeric.int.value", "int64"},
		{"numeric.float.value", "float64"},
		{"default.namespace", "string"},
		{"apps.platform.service", "string"},
	}

	for _, tc := range typeTestCases {
		val := getNestedValue(mergedConfig, tc.key)
		if val != nil {
			actualType := fmt.Sprintf("%T", val)
			status := "❌"
			if actualType == tc.expectedType {
				status = "✅"
			}
			fmt.Printf("   %s %s: %v (%s), 期望类型: %s\n",
				status, tc.key, val, actualType, tc.expectedType)
		} else {
			fmt.Printf("   ❌ %s: 未找到\n", tc.key)
		}
	}

	// 7. 嵌套配置项访问演示
	fmt.Println("\n7. 嵌套配置项访问演示:")
	fmt.Println("   环境变量中的下划线会被转换为点号，形成嵌套结构:")

	// 演示如何访问嵌套配置项
	fmt.Println("\n   访问数据库配置:")
	if databaseConfig, ok := mergedConfig["database"].(map[string]any); ok {
		if host, ok := databaseConfig["host"]; ok {
			fmt.Printf("     database.host = %v\n", host)
		}
		if port, ok := databaseConfig["port"]; ok {
			fmt.Printf("     database.port = %v\n", port)
		}
		if credentials, ok := databaseConfig["credentials"].(map[string]any); ok {
			if username, ok := credentials["username"]; ok {
				fmt.Printf("     database.credentials.username = %v\n", username)
			}
			if password, ok := credentials["password"]; ok {
				fmt.Printf("     database.credentials.password = %v\n", password)
			}
		}
	}

	fmt.Println("\n   访问服务器设置:")
	if serverConfig, ok := mergedConfig["server"].(map[string]any); ok {
		if settings, ok := serverConfig["settings"].(map[string]any); ok {
			if httpConfig, ok := settings["http"].(map[string]any); ok {
				if port, ok := httpConfig["port"]; ok {
					fmt.Printf("     server.settings.http.port = %v\n", port)
				}
			}
			if httpsConfig, ok := settings["https"].(map[string]any); ok {
				if port, ok := httpsConfig["port"]; ok {
					fmt.Printf("     server.settings.https.port = %v\n", port)
				}
			}
		}
	}

	fmt.Println("\n   访问应用信息:")
	if appConfig, ok := mergedConfig["app"].(map[string]any); ok {
		if info, ok := appConfig["info"].(map[string]any); ok {
			if name, ok := info["name"]; ok {
				fmt.Printf("     app.info.name = %v\n", name)
			}
			if version, ok := info["version"]; ok {
				fmt.Printf("     app.info.version = %v\n", version)
			}
		}
	}

	// 8. 问题诊断
	fmt.Println("\n8. 问题诊断:")
	fmt.Println("   如果无法通过 'default.namespace' 读取 'DEFAULT_NAMESPACE' 环境变量，请检查:")
	fmt.Println("   - 环境变量是否已正确设置")
	fmt.Println("   - 环境变量名称是否包含特殊字符")
	fmt.Println("   - 是否使用了正确的前缀")
	fmt.Println("   - 环境变量值是否包含非法字符")

	fmt.Println("\n=== 演示程序结束 ===")
}
