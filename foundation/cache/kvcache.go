package cache

import (
	"math"
	"time"
)

type CleanCallBackFunc func(string)

// KVCache 缓存对象
type KVCache interface {
	// Put maxAge单位minute
	Put(key string, data interface{}, maxAge float64) string
	Fetch(key string) interface{}
	Search(opr SearchOpr) interface{}
	Remove(key string)
	GetAll() []interface{}
	ClearAll()
	Release()
}

// NewKVCache 创建Cache对象
func NewKVCache(cleanCallBack CleanCallBackFunc) KVCache {
	cache := make(MemoryKVCache)

	go cache.run(cleanCallBack)
	go cache.checkTimeOut()

	return &cache
}

type putInKVData struct {
	key    string
	data   interface{}
	maxAge float64
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

// MemoryKVCache 内存缓存
type MemoryKVCache chan commandData

// Put 投放数据，返回数据的唯一标示
func (s *MemoryKVCache) Put(key string, data interface{}, maxAge float64) string {
	reply := make(chan interface{})

	dataPtr := &putInKVData{}
	dataPtr.key = key
	dataPtr.data = data
	dataPtr.maxAge = maxAge

	*s <- commandData{action: putIn, value: dataPtr, result: reply}

	result := (<-reply).(*putInKVResult).value
	return result
}

// Fetch 获取数据
func (s *MemoryKVCache) Fetch(key string) interface{} {
	reply := make(chan interface{})

	dataPtr := &fetchOutKVData{}
	dataPtr.key = key

	*s <- commandData{action: fetchOut, value: dataPtr, result: reply}

	result := (<-reply).(*fetchOutKVResult)
	return result.value
}

func (s *MemoryKVCache) Search(opr SearchOpr) interface{} {
	if opr == nil {
		return nil
	}

	reply := make(chan interface{})

	dataPtr := &searchKVData{}
	dataPtr.opr = opr

	*s <- commandData{action: search, value: dataPtr, result: reply}

	result := (<-reply).(*searchKVResult)
	return result.value
}

// Remove 清除数据
func (s *MemoryKVCache) Remove(key string) {
	dataPtr := &removeKVData{}
	dataPtr.key = key

	*s <- commandData{action: remove, value: dataPtr}
}

// GetAll 获取所有的数据
func (s *MemoryKVCache) GetAll() (ret []interface{}) {
	reply := make(chan interface{})

	*s <- commandData{action: getAll, value: nil, result: reply}

	result := (<-reply).(*getAllKVResult)

	ret = result.value

	return
}

// ClearAll 清除所有数据
func (s *MemoryKVCache) ClearAll() {

	*s <- commandData{action: clearAll}
}

// Release 释放Cache
func (s *MemoryKVCache) Release() {
	*s <- commandData{action: end}

	close(*s)
}

func (s *MemoryKVCache) run(cleanCallBack CleanCallBackFunc) {
	localCacheData := make(map[string]*cacheKVData)

	for command := range *s {
		switch command.action {
		case putIn:
			dataPtr := &cacheKVData{}
			dataPtr.cacheData = command.value.(*putInKVData)
			dataPtr.cacheTime = time.Now()

			localCacheData[dataPtr.cacheData.key] = dataPtr

			result := &putInKVResult{}
			result.value = dataPtr.cacheData.key

			command.result <- result
		case fetchOut:
			key := command.value.(*fetchOutKVData).key
			dataPtr, found := localCacheData[key]

			result := &fetchOutKVResult{}
			if found {
				dataPtr.cacheTime = time.Now()
				localCacheData[key] = dataPtr

				result.value = dataPtr.cacheData.data
			}

			command.result <- result
		case search:
			opr := command.value.(*searchKVData).opr

			result := &searchKVResult{}
			for _, v := range localCacheData {
				if opr(v.cacheData.data) {
					v.cacheTime = time.Now()
					result.value = v.cacheData.data
					break
				}
			}

			command.result <- result
		case remove:
			key := command.value.(*removeKVData).key

			delete(localCacheData, key)
		case getAll:
			result := &getAllKVResult{value: []interface{}{}}
			for _, v := range localCacheData {
				v.cacheTime = time.Now()
				result.value = append(result.value, v.cacheData.data)
			}

			command.result <- result
		case clearAll:
			localCacheData = make(map[string]*cacheKVData)
		case checkTimeOut:
			keys := []string{}
			// 检查每项数据是否超时，超时数据需要主动清除掉
			for k, v := range localCacheData {
				if math.Abs(v.cacheData.maxAge-ForeverAgeValue) > 0.001 {
					current := time.Now()
					elapse := current.Sub(v.cacheTime).Minutes()
					if elapse > v.cacheData.maxAge {
						keys = append(keys, k)
					}
				}
			}

			go func() {
				for _, v := range keys {
					if cleanCallBack != nil {
						cleanCallBack(v)
					}
					s.Remove(v)
				}
			}()
		case end:
			localCacheData = nil
		}
	}
}

func (s *MemoryKVCache) checkTimeOut() {
	timeOutTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeOutTimer.C:
			*s <- commandData{action: checkTimeOut}
		}
	}
}
