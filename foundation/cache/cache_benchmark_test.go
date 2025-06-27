package cache

import (
	"testing"
)

func BenchmarkKVCache_Put(b *testing.B) {
	cache := NewKVCache(nil)
	defer cache.Release()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put("key", "value", 10)
	}
}

func BenchmarkKVCache_Fetch(b *testing.B) {
	cache := NewKVCache(nil)
	defer cache.Release()
	cache.Put("key", "value", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Fetch("key")
	}
}

func BenchmarkKVCache_ConcurrentPut(b *testing.B) {
	cache := NewKVCache(nil)
	defer cache.Release()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Put("key", "value", 10)
		}
	})
}

func BenchmarkKVCache_ConcurrentFetch(b *testing.B) {
	cache := NewKVCache(nil)
	defer cache.Release()
	cache.Put("key", "value", 10)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Fetch("key")
		}
	})
}

func BenchmarkMemoryCache_Put(b *testing.B) {
	cache := NewCache(nil)
	defer cache.Release()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put("value", 10)
	}
}

func BenchmarkMemoryCache_Fetch(b *testing.B) {
	cache := NewCache(nil)
	defer cache.Release()
	id := cache.Put("value", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Fetch(id)
	}
}

func BenchmarkMemoryCache_ConcurrentPut(b *testing.B) {
	cache := NewCache(nil)
	defer cache.Release()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Put("value", 10)
		}
	})
}

func BenchmarkMemoryCache_ConcurrentFetch(b *testing.B) {
	cache := NewCache(nil)
	defer cache.Release()
	id := cache.Put("value", 10)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Fetch(id)
		}
	})
}
