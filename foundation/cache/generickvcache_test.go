package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/log"
)

func TestGenericKVCache_StringKeyIntValue(t *testing.T) {
	cache := NewGenericKVCache[string, int](nil)
	defer cache.Release()

	// Put
	key := cache.Put("testKey", 42, 1)
	if key != "testKey" {
		t.Errorf("Put() = %v, want %v", key, "testKey")
	}

	// Fetch existing
	value := cache.Fetch("testKey")
	if value != 42 {
		t.Errorf("Fetch() = %v, want %v", value, 42)
	}

	// Fetch non-existing
	value = cache.Fetch("nonExistKey")
	if value != 0 {
		t.Errorf("Fetch() non exist key = %v, want 0", value)
	}

	// Search
	result := cache.Search(func(v int) bool {
		return v == 42
	})
	if result != 42 {
		t.Errorf("Search() = %v, want %v", result, 42)
	}

	// Search no match
	result = cache.Search(func(v int) bool {
		return v == 999
	})
	if result != 0 {
		t.Errorf("Search() no match = %v, want 0", result)
	}

	// Remove
	cache.Remove("testKey")
	value = cache.Fetch("testKey")
	if value != 0 {
		t.Errorf("Fetch() after Remove() = %v, want 0", value)
	}

	// GetAll
	cache.Put("key1", 1, 1)
	cache.Put("key2", 2, 1)
	all := cache.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() length = %v, want 2", len(all))
	}

	// ClearAll
	cache.ClearAll()
	all = cache.GetAll()
	if len(all) != 0 {
		t.Errorf("GetAll() after ClearAll() = %v, want empty", all)
	}
}

func TestGenericKVCache_IntKeyStringValue(t *testing.T) {
	cache := NewGenericKVCache[int, string](nil)
	defer cache.Release()

	cache.Put(100, "value100", 1)
	val := cache.Fetch(100)
	if val != "value100" {
		t.Errorf("Fetch() = %v, want %v", val, "value100")
	}

	cache.Put(200, "value200", 1)
	val = cache.Fetch(200)
	if val != "value200" {
		t.Errorf("Fetch() = %v, want %v", val, "value200")
	}

	// Search
	result := cache.Search(func(v string) bool {
		return v == "value100"
	})
	if result != "value100" {
		t.Errorf("Search() = %v, want %v", result, "value100")
	}

	// Remove
	cache.Remove(100)
	val = cache.Fetch(100)
	if val != "" {
		t.Errorf("Fetch() after Remove() = %v, want empty", val)
	}
}

func TestGenericKVCache_TimeoutCleanup(t *testing.T) {
	callbackCalled := false
	cache := NewGenericKVCache[string, string](func(key string) {
		log.Warnf("Timeout cleanup callback called for key %s", key)
		callbackCalled = true
	})
	defer cache.Release()

	cache.Put("testKey", "testValue", 1) // 1 second timeout
	// Wait for timeout plus extra to ensure cleanup runs (check interval is 5 seconds)
	time.Sleep(10 * time.Second)

	if !callbackCalled {
		t.Error("Timeout cleanup callback not called")
	}

	val := cache.Fetch("testKey")
	if val != "" {
		t.Errorf("Fetch() after timeout = %v, want empty", val)
	}
}

func TestGenericKVCache_ConcurrentAccess(t *testing.T) {
	cache := NewGenericKVCache[string, int](nil)
	defer cache.Release()

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
}

func TestGenericKVCache_DifferentDataTypes(t *testing.T) {
	// Test with struct value
	type MyStruct struct {
		Name string
		Age  int
	}
	cache := NewGenericKVCache[string, MyStruct](nil)
	defer cache.Release()

	expected := MyStruct{Name: "Alice", Age: 30}
	cache.Put("struct", expected, 1)
	val := cache.Fetch("struct")
	if val != expected {
		t.Errorf("Failed to store struct value")
	}

	// Test with slice value
	cache2 := NewGenericKVCache[string, []int](nil)
	defer cache2.Release()
	slice := []int{1, 2, 3}
	cache2.Put("slice", slice, 1)
	val2 := cache2.Fetch("slice")
	if len(val2) != 3 || val2[0] != 1 {
		t.Errorf("Failed to store slice value")
	}
}

func TestGenericKVCache_NegativeTimeout(t *testing.T) {
	cache := NewGenericKVCache[string, string](nil)
	defer cache.Release()

	// negative maxAge means forever
	key := cache.Put("foreverKey", "foreverValue", -1)
	if key != "foreverKey" {
		t.Error("Should handle negative timeout")
	}

	val := cache.Fetch("foreverKey")
	if val != "foreverValue" {
		t.Errorf("Fetch() = %v, want %v", val, "foreverValue")
	}
}

func TestGenericKVCache_EmptyKey(t *testing.T) {
	cache := NewGenericKVCache[string, string](nil)
	defer cache.Release()

	// empty key is allowed?
	key := cache.Put("", "value", 1)
	if key != "" {
		t.Error("Should allow empty key")
	}

	val := cache.Fetch("")
	if val != "value" {
		t.Errorf("Fetch() empty key = %v, want value", val)
	}
}
