package cache

import (
	"context"
	"sync"
	"time"
)

// KVCache 缓存对象
type KVCache interface {
	// Put maxAge单位minute
	Put(key string, data interface{}, maxAge int64) string
	Fetch(key string) interface{}
	Search(opr SearchOpr) interface{}
	Remove(key string)
	GetAll() []interface{}
	ClearAll()
	Release()
}

// 定义缺失的类型
type putInKVData struct {
	key    string
	data   interface{}
	maxAge int64
}

type putInKVResult struct {
	value string
}

type fetchOutKVData struct {
	key string
}

type fetchOutKVResult struct {
	value interface{}
}

type searchKVData struct {
	opr SearchOpr
}

type searchKVResult fetchOutKVResult

type getAllKVResult struct {
	value []interface{}
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
	cacheCtx, cacheCancel := context.WithCancel(context.Background())

	cache := &MemoryKVCache{
		commandChannel:       make(chan commandData, 100),
		cancelFunc:           cacheCancel,
		expiredCleanCallBack: cleanCallBack,
		rwLock:               new(sync.RWMutex),
	}

	// 启动多个worker处理命令
	for i := 0; i < ConcurrentGoroutines; i++ {
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
}

// Put 投放数据，返回数据的唯一标示
func (s *MemoryKVCache) Put(key string, data interface{}, maxAge int64) string {
	dataPtr := &putInKVData{
		key:    key,
		data:   data,
		maxAge: maxAge,
	}

	result := s.sendCommand(commandData{action: putIn, value: dataPtr}).(*putInKVResult)
	return result.value
}

// Fetch 获取数据
func (s *MemoryKVCache) Fetch(key string) interface{} {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	v, found := s.localCacheData.Load(key)
	if !found {
		return nil
	}

	dataPtr := v.(*cacheKVData)
	if s.isExpired(dataPtr) {
		s.localCacheData.Delete(key)
		return nil
	}

	return dataPtr.cacheData.data
}

// Search 搜索数据
func (s *MemoryKVCache) Search(opr SearchOpr) interface{} {
	if opr == nil {
		return nil
	}

	dataPtr := &searchKVData{}
	dataPtr.opr = opr

	result := s.sendCommand(commandData{action: search, value: dataPtr}).(*searchKVResult)
	return result.value
}

// Remove 清除数据
func (s *MemoryKVCache) Remove(key string) {
	dataPtr := &removeKVData{}
	dataPtr.key = key

	s.sendCommand(commandData{action: remove, value: dataPtr})
}

// GetAll 获取所有的数据
func (s *MemoryKVCache) GetAll() (ret []interface{}) {
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
	s.cancelFunc()

	// 为每个worker发送end命令
	for i := 0; i < ConcurrentGoroutines; i++ {
		s.sendCommand(commandData{action: end})
	}

	s.cacheWg.Wait()
	close(s.commandChannel)
}

func (s *MemoryKVCache) sendCommand(command commandData) interface{} {
	var reply chan interface{}
	if v := s.pool.Get(); v != nil {
		reply = v.(chan interface{})
	} else {
		reply = make(chan interface{})
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
			dataPtr := &cacheKVData{
				cacheData: command.value.(*putInKVData),
				cacheTime: time.Now(),
			}
			s.localCacheData.Store(dataPtr.cacheData.key, dataPtr)
			s.rwLock.Unlock()

			result := &putInKVResult{value: dataPtr.cacheData.key}
			command.result <- result

		case fetchOut:
			key := command.value.(*fetchOutKVData).key
			v, found := s.localCacheData.Load(key)

			result := &fetchOutKVResult{}
			if found {
				dataPtr := v.(*cacheKVData)
				dataPtr.cacheTime = time.Now()
				s.localCacheData.Store(key, dataPtr)
				result.value = dataPtr.cacheData.data
			}

			command.result <- result

		case search:
			opr := command.value.(*searchKVData).opr

			result := &searchKVResult{}
			s.localCacheData.Range(func(k, v interface{}) bool {
				dataPtr := v.(*cacheKVData)
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
			key := command.value.(*removeKVData).key
			s.localCacheData.Delete(key)
			command.result <- true

		case getAll:
			result := &getAllKVResult{value: []interface{}{}}
			s.localCacheData.Range(func(k, v interface{}) bool {
				dataPtr := v.(*cacheKVData)
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

func (s *MemoryKVCache) getExpiredKeys() []string {
	keys := []string{}
	s.localCacheData.Range(func(k, v interface{}) bool {
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

func (s *MemoryKVCache) isExpired(data *cacheKVData) bool {
	if data.cacheData.maxAge != ForeverAgeValue {
		return int64(time.Since(data.cacheTime).Seconds()) > data.cacheData.maxAge
	}
	return false
}
