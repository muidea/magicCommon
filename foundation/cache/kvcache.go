package cache

import (
	"time"
)

// KVCache 缓存对象
type KVCache interface {
	// maxAge单位minute
	Put(key string, data interface{}, maxAge float64) string
	Fetch(key string) (interface{}, bool)
	Remove(key string)
	GetAll() []string
	ClearAll()
	Release()
}

// NewKVCache 创建Cache对象
func NewKVCache() KVCache {
	cache := make(MemoryKVCache)

	go cache.run()
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
	found bool
}

type getAllKVResult struct {
	value []string
}

type removeKVData struct {
	key string
}

type cacheKVData struct {
	putInKVData
	cacheTime time.Time
}

// MemoryKVCache 内存缓存
type MemoryKVCache chan commandData

// Put 投放数据，返回数据的唯一标示
func (right *MemoryKVCache) Put(key string, data interface{}, maxAge float64) string {

	reply := make(chan interface{})

	putInKVData := &putInKVData{}
	putInKVData.key = key
	putInKVData.data = data
	putInKVData.maxAge = maxAge

	*right <- commandData{action: putData, value: putInKVData, result: reply}

	result := (<-reply).(*putInKVResult).value
	return result
}

// Fetch 获取数据
func (right *MemoryKVCache) Fetch(key string) (interface{}, bool) {

	reply := make(chan interface{})

	fetchOutKVData := &fetchOutKVData{}
	fetchOutKVData.key = key

	*right <- commandData{action: fetchData, value: fetchOutKVData, result: reply}

	result := (<-reply).(*fetchOutKVResult)
	return result.value, result.found
}

// Remove 清除数据
func (right *MemoryKVCache) Remove(key string) {
	removeKVData := &removeKVData{}
	removeKVData.key = key

	*right <- commandData{action: remove, value: removeKVData}
}

// GetAll 获取所有的数据
func (right *MemoryKVCache) GetAll() (ret []string) {
	reply := make(chan interface{})

	*right <- commandData{action: getAll, value: nil, result: reply}

	result := (<-reply).(*getAllKVResult)

	ret = result.value

	return
}

// ClearAll 清除所有数据
func (right *MemoryKVCache) ClearAll() {

	*right <- commandData{action: clearAll}
}

// Release 释放Cache
func (right *MemoryKVCache) Release() {
	*right <- commandData{action: end}

	close(*right)
}

func (right *MemoryKVCache) run() {
	_cacheData := make(map[string]cacheKVData)

	for command := range *right {
		switch command.action {
		case putData:
			cacheKVData := cacheKVData{}
			cacheKVData.putInKVData = *(command.value.(*putInKVData))
			cacheKVData.cacheTime = time.Now()

			_cacheData[cacheKVData.key] = cacheKVData

			result := &putInKVResult{}
			result.value = cacheKVData.key

			command.result <- result
		case fetchData:
			key := command.value.(*fetchOutKVData).key

			cacheKVData, found := _cacheData[key]

			result := &fetchOutKVResult{}
			result.found = found
			if found {
				cacheKVData.cacheTime = time.Now()
				_cacheData[key] = cacheKVData

				result.value = cacheKVData.data
			}

			command.result <- result
		case remove:
			key := command.value.(*removeKVData).key

			delete(_cacheData, key)
		case getAll:
			result := &getAllKVResult{value: []string{}}
			for k := range _cacheData {
				result.value = append(result.value, k)
			}

			command.result <- result
		case clearAll:
			_cacheData = make(map[string]cacheKVData)

		case checkTimeOut:
			// 检查每项数据是否超时，超时数据需要主动清除掉
			for k, v := range _cacheData {
				if v.maxAge != MaxAgeValue {
					current := time.Now()
					elapse := current.Sub(v.cacheTime).Minutes()
					if elapse > v.maxAge {
						delete(_cacheData, k)
					}
				}
			}
		case end:
			_cacheData = nil
		}
	}
}

func (right *MemoryKVCache) checkTimeOut() {
	timeOutTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeOutTimer.C:
			*right <- commandData{action: checkTimeOut}
		}
	}
}
