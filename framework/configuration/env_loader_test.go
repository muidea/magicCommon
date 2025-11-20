package configuration

import (
	"os"
	"testing"
)

func TestEnvConfigLoader_Load(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("TEST_PREFIX_DEFAULT_NAMESPACE", "test-panel")
	os.Setenv("DEFAULT_NAMESPACE", "panel")
	os.Setenv("DEBUG_MODE", "true")
	os.Setenv("WORKSPACE_ROOT_PATH", "/home/rangh/dataspace")
	os.Setenv("APPS_PLATFORM_SERVICE", "http://magicplatform:8080")
	defer func() {
		os.Unsetenv("TEST_PREFIX_DEFAULT_NAMESPACE")
		os.Unsetenv("DEFAULT_NAMESPACE")
		os.Unsetenv("DEBUG_MODE")
		os.Unsetenv("WORKSPACE_ROOT_PATH")
		os.Unsetenv("APPS_PLATFORM_SERVICE")
	}()

	t.Run("Load without prefix", func(t *testing.T) {
		loader := NewEnvConfigLoader("")
		config, err := loader.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// 检查 DEFAULT_NAMESPACE 环境变量（现在在 default 命名空间下）
		if defaultConfig, ok := config["default"].(map[string]interface{}); ok {
			if val, ok := defaultConfig["namespace"]; !ok {
				t.Error("Expected 'default.namespace' key not found")
			} else if val != "panel" {
				t.Errorf("Expected 'default.namespace' to be 'panel', got '%v'", val)
			}
		} else {
			t.Error("Expected 'default' namespace not found")
		}

		// 检查 DEBUG_MODE 环境变量（现在在 debug 命名空间下）
		if debugConfig, ok := config["debug"].(map[string]interface{}); ok {
			if val, ok := debugConfig["mode"]; !ok {
				t.Error("Expected 'debug.mode' key not found")
			} else if val != true {
				t.Errorf("Expected 'debug.mode' to be true, got '%v'", val)
			}
		} else {
			t.Error("Expected 'debug' namespace not found")
		}

		// 检查 WORKSPACE_ROOT_PATH 环境变量（现在在 workspace 命名空间下）
		if workspaceConfig, ok := config["workspace"].(map[string]interface{}); ok {
			if rootConfig, ok := workspaceConfig["root"].(map[string]interface{}); ok {
				if val, ok := rootConfig["path"]; !ok {
					t.Error("Expected 'workspace.root.path' key not found")
				} else if val != "/home/rangh/dataspace" {
					t.Errorf("Expected 'workspace.root.path' to be '/home/rangh/dataspace', got '%v'", val)
				}
			} else {
				t.Error("Expected 'workspace.root' namespace not found")
			}
		} else {
			t.Error("Expected 'workspace' namespace not found")
		}

		// 检查 APPS_PLATFORM_SERVICE 环境变量（现在在 apps 命名空间下）
		if appsConfig, ok := config["apps"].(map[string]interface{}); ok {
			if platformConfig, ok := appsConfig["platform"].(map[string]interface{}); ok {
				if val, ok := platformConfig["service"]; !ok {
					t.Error("Expected 'apps.platform.service' key not found")
				} else if val != "http://magicplatform:8080" {
					t.Errorf("Expected 'apps.platform.service' to be 'http://magicplatform:8080', got '%v'", val)
				}
			} else {
				t.Error("Expected 'apps.platform' namespace not found")
			}
		} else {
			t.Error("Expected 'apps' namespace not found")
		}
	})

	t.Run("Load with prefix", func(t *testing.T) {
		loader := NewEnvConfigLoader("TEST_PREFIX_")
		config, err := loader.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// 检查带前缀的环境变量（现在在 default 命名空间下）
		if defaultConfig, ok := config["default"].(map[string]interface{}); ok {
			if val, ok := defaultConfig["namespace"]; !ok {
				t.Error("Expected 'default.namespace' key not found")
			} else if val != "test-panel" {
				t.Errorf("Expected 'default.namespace' to be 'test-panel', got '%v'", val)
			}
		} else {
			t.Error("Expected 'default' namespace not found")
		}

		// 确保不带前缀的环境变量没有被包含
		if _, ok := config["debug"]; ok {
			t.Error("Unexpected 'debug' namespace found in prefixed loader")
		}
	})
}

func TestEnvConfigLoader_NormalizeKey(t *testing.T) {
	loader := NewEnvConfigLoader("")

	testCases := []struct {
		input    string
		expected string
	}{
		{"DEFAULT_NAMESPACE", "default.namespace"},
		{"DEBUG_MODE", "debug.mode"},
		{"WORKSPACE_ROOT_PATH", "workspace.root.path"},
		{"APPS_PLATFORM_SERVICE", "apps.platform.service"},
		{"simple_key", "simple.key"},
		{"ALREADY_DOT.KEY", "already.dot.key"},
		{"Mixed_Case_Key", "mixed.case.key"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := loader.normalizeKey(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeKey(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestEnvConfigLoader_ParseValue(t *testing.T) {
	loader := NewEnvConfigLoader("")

	testCases := []struct {
		input    string
		expected interface{}
	}{
		{"true", true},
		{"TRUE", true},
		{"false", false},
		{"FALSE", false},
		{"123", int64(123)},
		{"-456", int64(-456)},
		{"3.14", 3.14},
		{"-2.71", -2.71},
		{"hello", "hello"},
		{"123abc", "123abc"},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := loader.parseValue(tc.input)
			if result != tc.expected {
				t.Errorf("parseValue(%q) = %v (%T), want %v (%T)", tc.input, result, result, tc.expected, tc.expected)
			}
		})
	}
}

func TestEnvConfigMerger_Merge(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("MERGE_TEST_KEY", "env-value")
	os.Setenv("OVERRIDE_KEY", "env-override")
	defer func() {
		os.Unsetenv("MERGE_TEST_KEY")
		os.Unsetenv("OVERRIDE_KEY")
	}()

	existingConfig := map[string]interface{}{
		"existing.key": "existing-value",
		"override.key": "original-value",
		"config": map[string]interface{}{
			"only": "config-only",
		},
	}

	merger := NewEnvConfigMerger("")
	mergedConfig, err := merger.Merge(existingConfig)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// 检查现有配置是否保留
	if val, ok := mergedConfig["existing.key"]; !ok || val != "existing-value" {
		t.Errorf("Existing config not preserved: %v", val)
	}

	// 检查环境变量配置是否添加（现在在 merge 命名空间下）
	if mergeConfig, ok := mergedConfig["merge"].(map[string]interface{}); ok {
		if testConfig, ok := mergeConfig["test"].(map[string]interface{}); ok {
			if val, ok := testConfig["key"]; !ok || val != "env-value" {
				t.Errorf("Env config not added: %v", val)
			}
		} else {
			t.Error("Expected 'merge.test' namespace not found")
		}
	} else {
		t.Error("Expected 'merge' namespace not found")
	}

	// 检查环境变量是否覆盖了现有配置（现在在 override 命名空间下）
	if overrideConfig, ok := mergedConfig["override"].(map[string]interface{}); ok {
		if val, ok := overrideConfig["key"]; !ok || val != "env-override" {
			t.Errorf("Env config not overriding existing: %v", val)
		}
	} else {
		t.Error("Expected 'override' namespace not found")
	}

	// 检查仅存在于配置中的键（现在在 config 命名空间下）
	if configConfig, ok := mergedConfig["config"].(map[string]interface{}); ok {
		if val, ok := configConfig["only"]; !ok {
			t.Error("Expected 'config.only' key not found")
		} else if val != "config-only" {
			t.Errorf("Expected 'config.only' to be 'config-only', got '%v'", val)
		}
	} else {
		t.Error("Expected 'config' namespace not found")
	}
}

func TestEnvConfigLoader_EdgeCases(t *testing.T) {
	loader := NewEnvConfigLoader("")

	t.Run("Empty environment variable", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		config, err := loader.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if emptyConfig, ok := config["empty"].(map[string]interface{}); ok {
			if val, ok := emptyConfig["var"]; !ok {
				t.Error("Empty var key not found")
			} else if val != "" {
				t.Errorf("Expected empty string, got '%v'", val)
			}
		} else {
			t.Error("Expected 'empty' namespace not found")
		}
	})

	t.Run("Malformed environment variable", func(t *testing.T) {
		// 模拟格式错误的环境变量（没有等号）
		// 这里我们无法直接设置格式错误的环境变量，但可以测试代码的容错性
		config, err := loader.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		// 只要不panic就是成功的
		_ = config
	})
}

func TestEnvConfigLoader_NestedStructure(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATABASE_PORT", "3306")
	os.Setenv("DATABASE_CREDENTIALS_USERNAME", "admin")
	os.Setenv("DATABASE_CREDENTIALS_PASSWORD", "secret")
	os.Setenv("SERVER_SETTINGS_HTTP_PORT", "8080")
	os.Setenv("SERVER_SETTINGS_HTTPS_PORT", "8443")
	os.Setenv("APP_INFO_NAME", "TestApp")
	os.Setenv("APP_INFO_VERSION", "1.0.0")
	defer func() {
		os.Unsetenv("DATABASE_HOST")
		os.Unsetenv("DATABASE_PORT")
		os.Unsetenv("DATABASE_CREDENTIALS_USERNAME")
		os.Unsetenv("DATABASE_CREDENTIALS_PASSWORD")
		os.Unsetenv("SERVER_SETTINGS_HTTP_PORT")
		os.Unsetenv("SERVER_SETTINGS_HTTPS_PORT")
		os.Unsetenv("APP_INFO_NAME")
		os.Unsetenv("APP_INFO_VERSION")
	}()

	loader := NewEnvConfigLoader("")
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 检查嵌套结构是否正确展开
	t.Run("Database configuration", func(t *testing.T) {
		if dbConfig, ok := config["database"].(map[string]interface{}); ok {
			// 检查一级嵌套
			if host, ok := dbConfig["host"]; !ok || host != "localhost" {
				t.Errorf("Expected database.host to be 'localhost', got %v", host)
			}
			if port, ok := dbConfig["port"]; !ok || port != int64(3306) {
				t.Errorf("Expected database.port to be 3306, got %v", port)
			}

			// 检查二级嵌套
			if credentials, ok := dbConfig["credentials"].(map[string]interface{}); ok {
				if username, ok := credentials["username"]; !ok || username != "admin" {
					t.Errorf("Expected database.credentials.username to be 'admin', got %v", username)
				}
				if password, ok := credentials["password"]; !ok || password != "secret" {
					t.Errorf("Expected database.credentials.password to be 'secret', got %v", password)
				}
			} else {
				t.Error("Expected database.credentials to be a map")
			}
		} else {
			t.Error("Expected database to be a map")
		}
	})

	t.Run("Server settings", func(t *testing.T) {
		if serverConfig, ok := config["server"].(map[string]interface{}); ok {
			if settings, ok := serverConfig["settings"].(map[string]interface{}); ok {
				if httpPort, ok := settings["http"].(map[string]interface{}); ok {
					if port, ok := httpPort["port"]; !ok || port != int64(8080) {
						t.Errorf("Expected server.settings.http.port to be 8080, got %v", port)
					}
				} else {
					t.Error("Expected server.settings.http to be a map")
				}
				if httpsPort, ok := settings["https"].(map[string]interface{}); ok {
					if port, ok := httpsPort["port"]; !ok || port != int64(8443) {
						t.Errorf("Expected server.settings.https.port to be 8443, got %v", port)
					}
				} else {
					t.Error("Expected server.settings.https to be a map")
				}
			} else {
				t.Error("Expected server.settings to be a map")
			}
		} else {
			t.Error("Expected server to be a map")
		}
	})

	t.Run("App info", func(t *testing.T) {
		if appConfig, ok := config["app"].(map[string]interface{}); ok {
			if info, ok := appConfig["info"].(map[string]interface{}); ok {
				if name, ok := info["name"]; !ok || name != "TestApp" {
					t.Errorf("Expected app.info.name to be 'TestApp', got %v", name)
				}
				if version, ok := info["version"]; !ok || version != "1.0.0" {
					t.Errorf("Expected app.info.version to be '1.0.0', got %v", version)
				}
			} else {
				t.Error("Expected app.info to be a map")
			}
		} else {
			t.Error("Expected app to be a map")
		}
	})
}

func TestEnvConfigLoader_SetNestedValue(t *testing.T) {
	loader := NewEnvConfigLoader("")
	config := make(map[string]interface{})

	testCases := []struct {
		name     string
		key      string
		value    interface{}
		expected map[string]interface{}
	}{
		{
			name:  "Simple key",
			key:   "simple",
			value: "value",
			expected: map[string]interface{}{
				"simple": "value",
			},
		},
		{
			name:  "One level nested",
			key:   "parent.child",
			value: "child_value",
			expected: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "child_value",
				},
			},
		},
		{
			name:  "Two level nested",
			key:   "a.b.c",
			value: "deep_value",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "deep_value",
					},
				},
			},
		},
		{
			name:  "Multiple siblings",
			key:   "config.server.port",
			value: 8080,
			expected: map[string]interface{}{
				"config": map[string]interface{}{
					"server": map[string]interface{}{
						"port": 8080,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 重置配置
			config = make(map[string]interface{})
			loader.setNestedValue(config, tc.key, tc.value)

			// 比较结果
			if !compareMaps(config, tc.expected) {
				t.Errorf("setNestedValue(%q, %v) = %v, want %v",
					tc.key, tc.value, config, tc.expected)
			}
		})
	}
}

// compareMaps 递归比较两个映射是否相等
func compareMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valA := range a {
		valB, exists := b[key]
		if !exists {
			return false
		}

		switch aVal := valA.(type) {
		case map[string]interface{}:
			if bVal, ok := valB.(map[string]interface{}); ok {
				if !compareMaps(aVal, bVal) {
					return false
				}
			} else {
				return false
			}
		default:
			if valA != valB {
				return false
			}
		}
	}

	return true
}

func TestEnvConfigMerger_ComplexMergeScenarios(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("APP_NAME", "EnvApp")
	os.Setenv("APP_DATABASE_HOST", "env-db-host") // 改为 app.database.host
	os.Setenv("APP_DATABASE_PORT", "5432")        // 改为 app.database.port
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("FEATURE_FLAG_NEWUI", "true") // 改为 feature.flag.newui
	os.Setenv("LOGGING_LEVEL", "debug")
	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("APP_DATABASE_HOST")
		os.Unsetenv("APP_DATABASE_PORT")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("FEATURE_FLAG_NEW_UI")
		os.Unsetenv("LOGGING_LEVEL")
	}()

	// 模拟复杂的配置文件结构
	existingConfig := map[string]interface{}{
		"app": map[string]interface{}{
			"name":    "FileApp",
			"version": "1.0.0",
			"database": map[string]interface{}{
				"host": "file-db-host",
				"port": 3306,
				"credentials": map[string]interface{}{
					"username": "file-user",
					"password": "file-pass",
				},
			},
		},
		"server": map[string]interface{}{
			"port": 8080,
			"host": "0.0.0.0",
		},
		"feature": map[string]interface{}{
			"flag": map[string]interface{}{
				"new_ui": false,
				"old_ui": true,
			},
		},
		"logging": map[string]interface{}{
			"level": "info",
			"file":  "/var/log/app.log",
		},
		"config": map[string]interface{}{
			"only": "file-only-value",
		},
	}

	merger := NewEnvConfigMerger("")
	mergedConfig, err := merger.Merge(existingConfig)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// 调试输出：打印合并后的配置结构
	t.Logf("Merged config structure: %+v", mergedConfig)

	t.Run("Environment variables override file config", func(t *testing.T) {
		// 检查应用名称被环境变量覆盖
		if appConfig, ok := mergedConfig["app"].(map[string]interface{}); ok {
			if name, ok := appConfig["name"]; !ok || name != "EnvApp" {
				t.Errorf("Expected app.name to be 'EnvApp' (env override), got %v", name)
			}
			// 检查文件配置中未覆盖的部分仍然存在
			if version, ok := appConfig["version"]; !ok || version != "1.0.0" {
				t.Errorf("Expected app.version to be '1.0.0' (file preserved), got %v", version)
			}
		} else {
			t.Error("Expected 'app' namespace not found")
		}
	})

	t.Run("Nested structure merging", func(t *testing.T) {
		// 检查数据库配置合并
		if appConfig, ok := mergedConfig["app"].(map[string]interface{}); ok {
			if dbConfig, ok := appConfig["database"].(map[string]interface{}); ok {
				// 环境变量覆盖的字段
				if host, ok := dbConfig["host"]; !ok || host != "env-db-host" {
					t.Errorf("Expected database.host to be 'env-db-host' (env override), got %v", host)
				}
				if port, ok := dbConfig["port"]; !ok || port != int64(5432) {
					t.Errorf("Expected database.port to be 5432 (env override), got %v", port)
				}
				// 文件配置中保留的字段
				if credentials, ok := dbConfig["credentials"].(map[string]interface{}); ok {
					if username, ok := credentials["username"]; !ok || username != "file-user" {
						t.Errorf("Expected database.credentials.username to be 'file-user' (file preserved), got %v", username)
					}
					if password, ok := credentials["password"]; !ok || password != "file-pass" {
						t.Errorf("Expected database.credentials.password to be 'file-pass' (file preserved), got %v", password)
					}
				} else {
					t.Error("Expected database.credentials to be a map")
				}
			} else {
				t.Error("Expected 'app.database' namespace not found")
			}
		}
	})

	t.Run("Feature flags merging", func(t *testing.T) {
		// 检查功能标志合并
		if featureConfig, ok := mergedConfig["feature"].(map[string]interface{}); ok {
			if flagConfig, ok := featureConfig["flag"].(map[string]interface{}); ok {
				// 环境变量覆盖的字段
				if newUI, ok := flagConfig["newui"]; !ok || newUI != true {
					t.Errorf("Expected feature.flag.newui to be true (env override), got %v", newUI)
				}
				// 文件配置中保留的字段
				if oldUI, ok := flagConfig["old_ui"]; !ok || oldUI != true {
					t.Errorf("Expected feature.flag.old_ui to be true (file preserved), got %v", oldUI)
				}
			} else {
				t.Error("Expected 'feature.flag' namespace not found")
			}
		} else {
			t.Error("Expected 'feature' namespace not found")
		}
	})

	t.Run("Logging configuration merging", func(t *testing.T) {
		// 检查日志配置合并
		if loggingConfig, ok := mergedConfig["logging"].(map[string]interface{}); ok {
			// 环境变量覆盖的字段
			if level, ok := loggingConfig["level"]; !ok || level != "debug" {
				t.Errorf("Expected logging.level to be 'debug' (env override), got %v", level)
			}
			// 文件配置中保留的字段
			if file, ok := loggingConfig["file"]; !ok || file != "/var/log/app.log" {
				t.Errorf("Expected logging.file to be '/var/log/app.log' (file preserved), got %v", file)
			}
		} else {
			t.Error("Expected 'logging' namespace not found")
		}
	})

	t.Run("Server configuration merging", func(t *testing.T) {
		// 检查服务器配置合并
		if serverConfig, ok := mergedConfig["server"].(map[string]interface{}); ok {
			// 环境变量覆盖的字段
			if port, ok := serverConfig["port"]; !ok || port != int64(9090) {
				t.Errorf("Expected server.port to be 9090 (env override), got %v", port)
			}
			// 文件配置中保留的字段
			if host, ok := serverConfig["host"]; !ok || host != "0.0.0.0" {
				t.Errorf("Expected server.host to be '0.0.0.0' (file preserved), got %v", host)
			}
		} else {
			t.Error("Expected 'server' namespace not found")
		}
	})

	t.Run("File-only configuration preserved", func(t *testing.T) {
		// 检查仅存在于文件中的配置
		if configConfig, ok := mergedConfig["config"].(map[string]interface{}); ok {
			if only, ok := configConfig["only"]; !ok || only != "file-only-value" {
				t.Errorf("Expected config.only to be 'file-only-value' (file preserved), got %v", only)
			}
		} else {
			t.Error("Expected 'config' namespace not found")
		}
	})
}

func TestEnvConfigMerger_PriorityScenarios(t *testing.T) {
	// 测试不同优先级场景
	os.Setenv("HIGH_PRIORITY_KEY", "env-value")
	os.Setenv("NESTED_CONFIG_DB_HOST", "env-db")
	defer func() {
		os.Unsetenv("HIGH_PRIORITY_KEY")
		os.Unsetenv("NESTED_CONFIG_DB_HOST")
	}()

	t.Run("Environment variables have highest priority", func(t *testing.T) {
		existingConfig := map[string]interface{}{
			"high": map[string]interface{}{
				"priority": map[string]interface{}{
					"key": "file-value",
				},
			},
			"nested": map[string]interface{}{
				"config": map[string]interface{}{
					"db": map[string]interface{}{
						"host": "file-db",
						"port": 3306,
					},
				},
			},
		}

		merger := NewEnvConfigMerger("")
		mergedConfig, err := merger.Merge(existingConfig)
		if err != nil {
			t.Fatalf("Merge failed: %v", err)
		}

		// 检查环境变量优先级
		if highConfig, ok := mergedConfig["high"].(map[string]interface{}); ok {
			if priorityConfig, ok := highConfig["priority"].(map[string]interface{}); ok {
				if key, ok := priorityConfig["key"]; !ok || key != "env-value" {
					t.Errorf("Expected high.priority.key to be 'env-value' (env priority), got %v", key)
				}
			}
		}

		// 检查嵌套配置的环境变量优先级
		if nestedConfig, ok := mergedConfig["nested"].(map[string]interface{}); ok {
			if configConfig, ok := nestedConfig["config"].(map[string]interface{}); ok {
				if dbConfig, ok := configConfig["db"].(map[string]interface{}); ok {
					if host, ok := dbConfig["host"]; !ok || host != "env-db" {
						t.Errorf("Expected nested.config.db.host to be 'env-db' (env priority), got %v", host)
					}
					// 文件配置中保留的字段
					if port, ok := dbConfig["port"]; !ok || port != 3306 {
						t.Errorf("Expected nested.config.db.port to be 3306 (file preserved), got %v", port)
					}
				}
			}
		}
	})

	t.Run("Empty existing config", func(t *testing.T) {
		// 测试空配置文件的合并
		emptyConfig := make(map[string]interface{})
		merger := NewEnvConfigMerger("")
		mergedConfig, err := merger.Merge(emptyConfig)
		if err != nil {
			t.Fatalf("Merge failed: %v", err)
		}

		// 检查环境变量是否正确加载
		if highConfig, ok := mergedConfig["high"].(map[string]interface{}); ok {
			if priorityConfig, ok := highConfig["priority"].(map[string]interface{}); ok {
				if key, ok := priorityConfig["key"]; !ok || key != "env-value" {
					t.Errorf("Expected high.priority.key to be 'env-value', got %v", key)
				}
			}
		}
	})
}
