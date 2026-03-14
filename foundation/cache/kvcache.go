package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// KVCache 缓存对象
// Deprecated: Use KVCacheGeneric[string, any] instead for type safety.
type KVCache interface {
	// Put maxAge单位second
	Put(key string, data any, maxAge int64) string
	Fetch(key string) any
	Search(opr SearchOpr) any
	Remove(key string)
	GetAll() []any
	ClearAll()
	Release()
}

// 定义缺失的类型
type putInKVData struct {
	key    string
	data   any
	maxAge int64
}

type putInKVResult struct {
	value string
}

type searchKVData struct {
	opr SearchOpr
}

type searchKVResult struct {
	value any
}

type getAllKVResult struct {
	value []any
}

type removeKVData struct {
	key string
}

type cacheKVData struct {
	cacheData *putInKVData
	cacheTime time.Time
}

// NewKVCache 创建Cache对象
func NewKVCache(cleanCallBack ExpiredCleanCallBackFunc) KVCache {
	return NewKVCacheWithOptions(cleanCallBack, nil)
}

func NewKVCacheWithOptions(cleanCallBack ExpiredCleanCallBackFunc, options *CacheOptions) KVCache {
	cacheCtx, cacheCancel := context.WithCancel(context.Background())
	cacheOptions := normalizeCacheOptions(options)

	cache := &MemoryKVCache{
		commandChannel:       make(chan commandData, 100),
		cancelFunc:           cacheCancel,
		expiredCleanCallBack: cleanCallBack,
		rwLock:               new(sync.RWMutex),
		capacity:             cacheOptions.Capacity,
		cleanupInterval:      cacheOptions.CleanupInterval,
	}

	// 启动多个worker处理命令
	for range ConcurrentGoroutines {
		cache.cacheWg.Add(1)
		go cache.run()
	}

	cache.cacheWg.Add(1)
	go cache.checkTimeOut(cacheCtx)

	return cache
}

// MemoryKVCache 内存缓存
type MemoryKVCache struct {
	commandChannel       chan commandData
	cancelFunc           context.CancelFunc
	cacheWg              sync.WaitGroup
	localCacheData       sync.Map
	pool                 sync.Pool
	expiredCleanCallBack ExpiredCleanCallBackFunc
	rwLock               *sync.RWMutex
	releasing            atomic.Bool
	released             atomic.Bool
	capacity             int
	cleanupInterval      time.Duration
	metrics              cacheMetrics
}

// Put 投放数据，返回数据的唯一标示
func (s *MemoryKVCache) Put(key string, data any, maxAge int64) string {
	dataPtr := &putInKVData{
		key:    key,
		data:   data,
		maxAge: maxAge,
	}

	result := s.sendCommand(commandData{action: putIn, value: dataPtr}).(*putInKVResult)
	return result.value
}

// Fetch 获取数据
func (s *MemoryKVCache) Fetch(key string) any {
	s.rwLock.RLock()
	v, found := s.localCacheData.Load(key)
	if !found {
		s.rwLock.RUnlock()
		s.metrics.misses.Add(1)
		return nil
	}

	dataPtr := v.(*cacheKVData)
	if s.isExpired(dataPtr) {
		s.rwLock.RUnlock()
		s.rwLock.Lock()
		s.localCacheData.Delete(key)
		s.rwLock.Unlock()
		s.metrics.expirations.Add(1)
		s.metrics.misses.Add(1)
		return nil
	}

	s.rwLock.RUnlock()
	s.metrics.hits.Add(1)
	return dataPtr.cacheData.data
}

// Search 搜索数据
func (s *MemoryKVCache) Search(opr SearchOpr) any {
	if opr == nil {
		return nil
	}

	dataPtr := &searchKVData{}
	dataPtr.opr = opr

	result := s.sendCommand(commandData{action: search, value: dataPtr}).(*searchKVResult)
	if result.value == nil {
		s.metrics.misses.Add(1)
	} else {
		s.metrics.hits.Add(1)
	}
	return result.value
}

// Remove 清除数据
func (s *MemoryKVCache) Remove(key string) {
	dataPtr := &removeKVData{}
	dataPtr.key = key

	s.sendCommand(commandData{action: remove, value: dataPtr})
}

// GetAll 获取所有的数据
func (s *MemoryKVCache) GetAll() (ret []any) {
	result := s.sendCommand(commandData{action: getAll}).(*getAllKVResult)
	ret = result.value
	return
}

// ClearAll 清除所有数据
func (s *MemoryKVCache) ClearAll() {
	s.sendCommand(commandData{action: clearAll})
}

// Release 释放Cache
func (s *MemoryKVCache) Release() {
	if !s.releasing.CompareAndSwap(false, true) {
		return
	}

	s.cancelFunc()

	// 为每个worker发送end命令
	for i := 0; i < ConcurrentGoroutines; i++ {
		s.sendCommand(commandData{action: end})
	}

	s.cacheWg.Wait()
	close(s.commandChannel)
	s.released.Store(true)
}

func (s *MemoryKVCache) sendCommand(command commandData) any {
	if s.released.Load() {
		return nil
	}

	var reply chan any
	if v := s.pool.Get(); v != nil {
		reply = v.(chan any)
	} else {
		reply = make(chan any)
	}
	defer s.pool.Put(reply)

	command.result = reply
	s.commandChannel <- command
	return <-reply
}

func (s *MemoryKVCache) run() {
	defer s.cacheWg.Done()

	for command := range s.commandChannel {
		switch command.action {
		case putIn:
			s.rwLock.Lock()
			key := command.value.(*putInKVData).key
			if _, exists := s.localCacheData.Load(key); !exists {
				s.enforceCapacityLocked()
			}
			dataPtr := &cacheKVData{
				cacheData: command.value.(*putInKVData),
				cacheTime: time.Now(),
			}
			s.localCacheData.Store(dataPtr.cacheData.key, dataPtr)
			s.rwLock.Unlock()
			s.metrics.puts.Add(1)

			result := &putInKVResult{value: dataPtr.cacheData.key}
			command.result <- result

		case search:
			opr := command.value.(*searchKVData).opr

			result := &searchKVResult{}
			s.rwLock.Lock()
			s.localCacheData.Range(func(k, v any) bool {
				dataPtr := v.(*cacheKVData)
				if opr(dataPtr.cacheData.data) {
					dataPtr.cacheTime = time.Now()
					s.localCacheData.Store(k, dataPtr)
					result.value = dataPtr.cacheData.data
					return false
				}
				return true
			})
			s.rwLock.Unlock()

			command.result <- result

		case remove:
			key := command.value.(*removeKVData).key
			s.localCacheData.Delete(key)
			command.result <- true

		case getAll:
			result := &getAllKVResult{value: []any{}}
			s.rwLock.Lock()
			s.localCacheData.Range(func(k, v any) bool {
				dataPtr := v.(*cacheKVData)
				dataPtr.cacheTime = time.Now()
				result.value = append(result.value, dataPtr.cacheData.data)
				return true
			})
			s.rwLock.Unlock()
			command.result <- result

		case clearAll:
			s.rwLock.Lock()
			s.localCacheData = sync.Map{}
			s.rwLock.Unlock()
			command.result <- true

		case checkTimeOut:
			s.rwLock.Lock()
			keys := s.getExpiredKeys()
			for _, v := range keys {
				if s.expiredCleanCallBack != nil {
					s.expiredCleanCallBack(v)
				}
				s.localCacheData.Delete(v)
				s.metrics.expirations.Add(1)
			}
			s.rwLock.Unlock()

		case end:
			command.result <- true
			return
		}
	}
}

func (s *MemoryKVCache) Stats() CacheStats {
	s.rwLock.RLock()
	entries := countSyncMap(&s.localCacheData)
	s.rwLock.RUnlock()
	return s.metrics.snapshot(entries, s.capacity)
}

func (s *MemoryKVCache) getExpiredKeys() []string {
	keys := []string{}
	s.localCacheData.Range(func(k, v any) bool {
		dataPtr := v.(*cacheKVData)
		if dataPtr.cacheData.maxAge != ForeverAgeValue {
			elapse := int64(time.Since(dataPtr.cacheTime).Seconds())
			if elapse > dataPtr.cacheData.maxAge {
				keys = append(keys, k.(string))
			}
		}
		return true
	})
	return keys
}

func (s *MemoryKVCache) checkTimeOut(ctx context.Context) {
	defer s.cacheWg.Done()

	timeOutTimer := time.NewTicker(s.cleanupInterval)
	defer timeOutTimer.Stop()

	for {
		select {
		case <-timeOutTimer.C:
			s.commandChannel <- commandData{action: checkTimeOut}
		case <-ctx.Done():
			return
		}
	}
}

func (s *MemoryKVCache) enforceCapacityLocked() {
	if s.capacity <= 0 || countSyncMap(&s.localCacheData) < s.capacity {
		return
	}

	var oldestKey string
	var oldestTime time.Time
	s.localCacheData.Range(func(k, v any) bool {
		dataPtr := v.(*cacheKVData)
		if oldestKey == "" || dataPtr.cacheTime.Before(oldestTime) {
			oldestKey = k.(string)
			oldestTime = dataPtr.cacheTime
		}
		return true
	})

	if oldestKey != "" {
		s.localCacheData.Delete(oldestKey)
		s.metrics.evictions.Add(1)
	}
}

func (s *MemoryKVCache) isExpired(data *cacheKVData) bool {
	if data.cacheData.maxAge != ForeverAgeValue {
		return int64(time.Since(data.cacheTime).Seconds()) > data.cacheData.maxAge
	}
	return false
}
