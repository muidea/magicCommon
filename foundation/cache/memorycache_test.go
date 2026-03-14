package cache

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"log/slog"
)

func TestMemoryCache_PutAndFetch(t *testing.T) {
	cache := NewCache(nil).(*MemoryCache)
	defer cache.Release()

	// 测试正常插入和获取
	data := "test data"
	id := cache.Put(data, 10)
	assert.NotEmpty(t, id)

	fetched := cache.Fetch(id)
	assert.Equal(t, data, fetched)

	// 测试获取不存在的key
	notExist := cache.Fetch("not-exist")
	assert.Nil(t, notExist)
}

func TestMemoryCache_Search(t *testing.T) {
	cache := NewCache(nil).(*MemoryCache)
	defer cache.Release()

	// 插入测试数据
	data1 := "data1"
	data2 := "data2"
	cache.Put(data1, 10)
	cache.Put(data2, 10)

	// 测试搜索
	result := cache.Search(func(v interface{}) bool {
		return v == data1
	})
	assert.Equal(t, data1, result)

	// 测试搜索不存在的条件
	result = cache.Search(func(v interface{}) bool {
		return false
	})
	assert.Nil(t, result)
}

func TestMemoryCache_Remove(t *testing.T) {
	cache := NewCache(nil).(*MemoryCache)
	defer cache.Release()

	// 插入并删除
	data := "test data"
	id := cache.Put(data, 10)
	cache.Remove(id)

	// 验证删除
	fetched := cache.Fetch(id)
	assert.Nil(t, fetched)
}

func TestMemoryCache_ClearAll(t *testing.T) {
	cache := NewCache(nil).(*MemoryCache)
	defer cache.Release()

	// 插入多个数据
	cache.Put("data1", 10)
	cache.Put("data2", 10)

	// 清空缓存
	cache.ClearAll()

	// 验证清空
	result := cache.Search(func(v interface{}) bool {
		return true
	})
	assert.Nil(t, result)
}

func TestMemoryCache_Timeout(t *testing.T) {
	cleanCalled := false
	cleanCallback := func(id string) {
		slog.Warn("Timeout cleanup callback called", "key", id)
		cleanCalled = true
	}

	cache := NewCache(cleanCallback).(*MemoryCache)
	defer cache.Release()

	// 插入短期数据
	cache.Put("test data", 1) // 1秒

	// 等待超时
	time.Sleep(10 * time.Second)

	// 验证清理回调被调用
	assert.True(t, cleanCalled)
}

func TestMemoryCache_ReleaseIsIdempotent(t *testing.T) {
	cache := NewCache(nil).(*MemoryCache)

	cache.Put("test data", 10)
	cache.Release()
	cache.Release()
}

func TestMemoryCache_TimeoutCleanupDoesNotBreakRelease(t *testing.T) {
	var cleanCalled atomic.Bool
	cache := NewCache(func(id string) {
		cleanCalled.Store(true)
	}).(*MemoryCache)

	cache.Put("test data", 1)
	time.Sleep(6 * time.Second)
	cache.Release()

	assert.True(t, cleanCalled.Load())
}

func TestMemoryCacheOptionsCapacityAndStats(t *testing.T) {
	cache := NewCacheWithOptions(nil, &CacheOptions{
		Capacity:        1,
		CleanupInterval: 50 * time.Millisecond,
	}).(*MemoryCache)
	defer cache.Release()

	firstID := cache.Put("first", 10)
	_ = cache.Fetch(firstID)
	cache.Put("second", 10)

	if got := cache.Fetch(firstID); got != nil {
		t.Fatalf("expected oldest entry to be evicted, got %v", got)
	}

	stats := cache.Stats()
	assert.Equal(t, 1, stats.Capacity)
	assert.Equal(t, 1, stats.Entries)
	assert.Equal(t, int64(2), stats.Puts)
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Evictions)
	assert.Equal(t, int64(1), stats.Misses)
}
