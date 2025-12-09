package cache

import (
	"context"
	"sync"
	"time"
)

// KVCacheGeneric 泛型缓存接口（键和值均为泛型）
type KVCacheGeneric[K comparable, V any] interface {
	// Put maxAge单位second
	Put(key K, data V, maxAge int64) K
	Fetch(key K) V
	Search(opr func(V) bool) V
	Remove(key K)
	GetAll() []V
	ClearAll()
	Release()
}

// ExpiredCleanCallBackFuncGeneric 过期清理回调函数（泛型键）
type ExpiredCleanCallBackFuncGeneric[K any] func(K)

// genericPutInKVData 存放数据
type genericPutInKVData[K comparable, V any] struct {
	key    K
	data   V
	maxAge int64
}

// genericPutInKVResult 存放结果
type genericPutInKVResult[K comparable] struct {
	value K
}

// genericSearchKVData 搜索数据
type genericSearchKVData[V any] struct {
	opr func(V) bool
}

// genericSearchKVResult 搜索结果
type genericSearchKVResult[V any] struct {
	value V
}

// genericGetAllKVResult 获取所有结果
type genericGetAllKVResult[V any] struct {
	value []V
}

// genericRemoveKVData 删除数据
type genericRemoveKVData[K comparable] struct {
	key K
}

// genericCacheKVData 缓存数据
type genericCacheKVData[K comparable, V any] struct {
	cacheData *genericPutInKVData[K, V]
	cacheTime time.Time
}

// GenericKVCache 泛型内存缓存
type GenericKVCache[K comparable, V any] struct {
	commandChannel       chan commandData
	cancelFunc           context.CancelFunc
	cacheWg              sync.WaitGroup
	localCacheData       sync.Map
	pool                 sync.Pool
	expiredCleanCallBack ExpiredCleanCallBackFuncGeneric[K]
	rwLock               *sync.RWMutex
}

// NewGenericKVCache 创建泛型Cache对象
func NewGenericKVCache[K comparable, V any](cleanCallBack ExpiredCleanCallBackFuncGeneric[K]) KVCacheGeneric[K, V] {
	cacheCtx, cacheCancel := context.WithCancel(context.Background())

	cache := &GenericKVCache[K, V]{
		commandChannel:       make(chan commandData, 100),
		cancelFunc:           cacheCancel,
		expiredCleanCallBack: cleanCallBack,
		rwLock:               new(sync.RWMutex),
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

// Put 投放数据，返回数据的唯一标示
func (s *GenericKVCache[K, V]) Put(key K, data V, maxAge int64) K {
	dataPtr := &genericPutInKVData[K, V]{
		key:    key,
		data:   data,
		maxAge: maxAge,
	}

	result := s.sendCommand(commandData{action: putIn, value: dataPtr}).(*genericPutInKVResult[K])
	return result.value
}

// Fetch 获取数据
func (s *GenericKVCache[K, V]) Fetch(key K) V {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	v, found := s.localCacheData.Load(key)
	if !found {
		var zero V
		return zero
	}

	dataPtr := v.(*genericCacheKVData[K, V])
	if s.isExpired(dataPtr) {
		s.localCacheData.Delete(key)
		var zero V
		return zero
	}

	return dataPtr.cacheData.data
}

// Search 搜索数据
func (s *GenericKVCache[K, V]) Search(opr func(V) bool) V {
	if opr == nil {
		var zero V
		return zero
	}

	dataPtr := &genericSearchKVData[V]{opr: opr}
	result := s.sendCommand(commandData{action: search, value: dataPtr}).(*genericSearchKVResult[V])
	return result.value
}

// Remove 清除数据
func (s *GenericKVCache[K, V]) Remove(key K) {
	dataPtr := &genericRemoveKVData[K]{key: key}
	s.sendCommand(commandData{action: remove, value: dataPtr})
}

// GetAll 获取所有的数据
func (s *GenericKVCache[K, V]) GetAll() []V {
	result := s.sendCommand(commandData{action: getAll}).(*genericGetAllKVResult[V])
	return result.value
}

// ClearAll 清除所有数据
func (s *GenericKVCache[K, V]) ClearAll() {
	s.sendCommand(commandData{action: clearAll})
}

// Release 释放Cache
func (s *GenericKVCache[K, V]) Release() {
	s.cancelFunc()

	// 为每个worker发送end命令
	for i := 0; i < ConcurrentGoroutines; i++ {
		s.sendCommand(commandData{action: end})
	}

	s.cacheWg.Wait()
	close(s.commandChannel)
}

func (s *GenericKVCache[K, V]) sendCommand(command commandData) any {
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

func (s *GenericKVCache[K, V]) run() {
	defer s.cacheWg.Done()

	for command := range s.commandChannel {
		switch command.action {
		case putIn:
			s.rwLock.Lock()
			dataPtr := &genericCacheKVData[K, V]{
				cacheData: command.value.(*genericPutInKVData[K, V]),
				cacheTime: time.Now(),
			}
			s.localCacheData.Store(dataPtr.cacheData.key, dataPtr)
			s.rwLock.Unlock()

			result := &genericPutInKVResult[K]{value: dataPtr.cacheData.key}
			command.result <- result

		case search:
			opr := command.value.(*genericSearchKVData[V]).opr

			result := &genericSearchKVResult[V]{}
			s.localCacheData.Range(func(k, v any) bool {
				dataPtr := v.(*genericCacheKVData[K, V])
				if opr(dataPtr.cacheData.data) {
					dataPtr.cacheTime = time.Now()
					s.localCacheData.Store(k, dataPtr)
					result.value = dataPtr.cacheData.data
					return false
				}
				return true
			})

			command.result <- result

		case remove:
			key := command.value.(*genericRemoveKVData[K]).key
			s.localCacheData.Delete(key)
			command.result <- true

		case getAll:
			result := &genericGetAllKVResult[V]{value: []V{}}
			s.localCacheData.Range(func(k, v any) bool {
				dataPtr := v.(*genericCacheKVData[K, V])
				dataPtr.cacheTime = time.Now()
				result.value = append(result.value, dataPtr.cacheData.data)
				return true
			})
			command.result <- result

		case clearAll:
			s.localCacheData = sync.Map{}
			command.result <- true

		case checkTimeOut:
			keys := s.getExpiredKeys()
			go func() {
				for _, v := range keys {
					if s.expiredCleanCallBack != nil {
						s.expiredCleanCallBack(v)
					}
					s.Remove(v)
				}
			}()

		case end:
			command.result <- true
			return
		}
	}
}

func (s *GenericKVCache[K, V]) getExpiredKeys() []K {
	keys := []K{}
	s.localCacheData.Range(func(k, v any) bool {
		dataPtr := v.(*genericCacheKVData[K, V])
		if dataPtr.cacheData.maxAge != ForeverAgeValue {
			elapse := int64(time.Since(dataPtr.cacheTime).Seconds())
			if elapse > dataPtr.cacheData.maxAge {
				keys = append(keys, k.(K))
			}
		}
		return true
	})
	return keys
}

func (s *GenericKVCache[K, V]) checkTimeOut(ctx context.Context) {
	defer s.cacheWg.Done()

	timeOutTimer := time.NewTicker(5 * time.Second)
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

func (s *GenericKVCache[K, V]) isExpired(data *genericCacheKVData[K, V]) bool {
	if data.cacheData.maxAge != ForeverAgeValue {
		return int64(time.Since(data.cacheTime).Seconds()) > data.cacheData.maxAge
	}
	return false
}
