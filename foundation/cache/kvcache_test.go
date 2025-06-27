package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/log"
)

func TestKVCache(t *testing.T) {
	// 创建KVCache实例
	cache := NewKVCache(nil)
	defer cache.Release()

	// 测试Put方法
	t.Run("Test Put", func(t *testing.T) {
		key := cache.Put("testKey", "testValue", 1)
		if key != "testKey" {
			t.Errorf("Put() = %v, want %v", key, "testKey")
		}

		// 测试重复Put
		key = cache.Put("testKey", "newValue", 1)
		if key != "testKey" {
			t.Errorf("Put() existing key = %v, want %v", key, "testKey")
		}
	})

	// 测试Fetch方法
	t.Run("Test Fetch", func(t *testing.T) {
		// 测试存在key
		value := cache.Fetch("testKey")
		if value != "newValue" {
			t.Errorf("Fetch() = %v, want %v", value, "newValue")
		}

		// 测试不存在key
		value = cache.Fetch("nonExistKey")
		if value != nil {
			t.Errorf("Fetch() non exist key = %v, want nil", value)
		}
	})

	// 测试Search方法
	t.Run("Test Search", func(t *testing.T) {
		// 测试匹配条件
		result := cache.Search(func(data interface{}) bool {
			log.Infof("%+v", data)
			return data == "newValue"
		})
		if result != "newValue" {
			t.Errorf("Search() = %v, want %v", result, "newValue")
		}

		// 测试不匹配条件
		result = cache.Search(func(data interface{}) bool {
			return data == "nonExistValue"
		})
		if result != nil {
			t.Errorf("Search() non exist value = %v, want nil", result)
		}

		// 测试nil条件
		result = cache.Search(nil)
		if result != nil {
			t.Errorf("Search() nil condition = %v, want nil", result)
		}
	})

	// 测试Remove方法
	t.Run("Test Remove", func(t *testing.T) {
		// 测试移除存在key
		cache.Remove("testKey")
		value := cache.Fetch("testKey")
		if value != nil {
			t.Errorf("Fetch() after Remove() = %v, want nil", value)
		}

		// 测试移除不存在key
		cache.Remove("nonExistKey")
	})

	// 测试GetAll方法
	t.Run("Test GetAll", func(t *testing.T) {
		// 清空缓存
		cache.ClearAll()

		// 测试空缓存
		allValues := cache.GetAll()
		if len(allValues) != 0 {
			t.Errorf("GetAll() empty cache = %v, want empty slice", allValues)
		}

		// 添加多个值
		cache.Put("key1", "value1", 1)
		cache.Put("key2", "value2", 1)
		cache.Put("key3", "value3", 1)

		// 测试获取所有值
		allValues = cache.GetAll()
		if len(allValues) != 3 {
			t.Errorf("GetAll() = %v, want 3 items", allValues)
		}
	})

	// 测试ClearAll方法
	t.Run("Test ClearAll", func(t *testing.T) {
		cache.ClearAll()
		allValues := cache.GetAll()
		if len(allValues) != 0 {
			t.Errorf("GetAll() after ClearAll() = %v, want empty slice", allValues)
		}
	})

	// 测试超时清理功能
	t.Run("Test Timeout Cleanup", func(t *testing.T) {
		callbackCalled := false
		cacheWithCallback := NewKVCache(func(key string) {
			log.Warnf("Timeout cleanup callback called for key %s", key)
			callbackCalled = true
		})
		defer cacheWithCallback.Release()

		cacheWithCallback.Put("testKey", "testValue", 1) // 设置1秒超时
		time.Sleep(10 * time.Second)                      // 等待超时
		if !callbackCalled {
			t.Error("Timeout cleanup callback not called")
		}
	})

	// 测试并发访问
	t.Run("Test Concurrent Access", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", i)
				cache.Put(key, i, 1)
				value := cache.Fetch(key)
				if value != i {
					t.Errorf("Concurrent access failed for key %s", key)
				}
			}(i)
		}
		wg.Wait()
	})

	// 测试不同类型数据存储
	t.Run("Test Different Data Types", func(t *testing.T) {
		testCases := []struct {
			key   string
			value interface{}
		}{
			{"string", "test string"},
			{"int", 123},
			{"float", 3.14},
			{"bool", true},
			{"struct", struct{ name string }{name: "test"}},
		}

		for _, tc := range testCases {
			cache.Put(tc.key, tc.value, 1)
			value := cache.Fetch(tc.key)
			if value != tc.value {
				t.Errorf("Failed to store %T value", tc.value)
			}
		}
	})

	// 测试错误处理
	t.Run("Test Error Handling", func(t *testing.T) {
		// 测试nil key
		key := cache.Put("", "value", 1)
		if key != "" {
			t.Error("Should not allow empty key")
		}

		// 测试负超时时间
		key = cache.Put("negativeTimeoutKey", "value", -1)
		if key != "negativeTimeoutKey" {
			t.Error("Should handle negative timeout")
		}
	})
}
