package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

const defaultCleanupInterval = 5 * time.Second

type CacheOptions struct {
	Capacity        int
	CleanupInterval time.Duration
}

func DefaultCacheOptions() CacheOptions {
	return CacheOptions{
		Capacity:        0,
		CleanupInterval: defaultCleanupInterval,
	}
}

func normalizeCacheOptions(options *CacheOptions) CacheOptions {
	if options == nil {
		return DefaultCacheOptions()
	}

	normalized := *options
	if normalized.Capacity < 0 {
		normalized.Capacity = 0
	}
	if normalized.CleanupInterval <= 0 {
		normalized.CleanupInterval = defaultCleanupInterval
	}

	return normalized
}

type CacheStats struct {
	Entries     int
	Capacity    int
	Puts        int64
	Hits        int64
	Misses      int64
	Evictions   int64
	Expirations int64
}

type cacheMetrics struct {
	puts        atomic.Int64
	hits        atomic.Int64
	misses      atomic.Int64
	evictions   atomic.Int64
	expirations atomic.Int64
}

func (s *cacheMetrics) snapshot(entries int, capacity int) CacheStats {
	return CacheStats{
		Entries:     entries,
		Capacity:    capacity,
		Puts:        s.puts.Load(),
		Hits:        s.hits.Load(),
		Misses:      s.misses.Load(),
		Evictions:   s.evictions.Load(),
		Expirations: s.expirations.Load(),
	}
}

func countSyncMap(data *sync.Map) int {
	count := 0
	data.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}
