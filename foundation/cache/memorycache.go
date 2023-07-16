package cache

import (
	"strings"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
)

// Cache 缓存对象
type Cache interface {
	// maxAge单位minute
	Put(data interface{}, maxAge float64) string
	Fetch(id string) (interface{}, bool)
	Remove(id string)
	ClearAll()
	Release()
}

// NewCache 创建Cache对象
func NewCache() Cache {
	cache := make(MemoryCache)

	go cache.run()
	go cache.checkTimeOut()

	return &cache
}

type putInData struct {
	data   interface{}
	maxAge float64
}

type putInResult struct {
	value string
}

type fetchOutData struct {
	id string
}

type fetchOutResult struct {
	value interface{}
	found bool
}

type removeData struct {
	id string
}

type cacheData struct {
	putInData
	cacheTime time.Time
}

// MemoryCache 内存缓存
type MemoryCache chan commandData

// Put 投放数据，返回数据的唯一标示
func (right *MemoryCache) Put(data interface{}, maxAge float64) string {

	reply := make(chan interface{})

	putInData := &putInData{}
	putInData.data = data
	putInData.maxAge = maxAge

	*right <- commandData{action: putData, value: putInData, result: reply}

	result := (<-reply).(*putInResult).value
	return result
}

// Fetch 获取数据
func (right *MemoryCache) Fetch(id string) (interface{}, bool) {

	reply := make(chan interface{})

	fetchOutData := &fetchOutData{}
	fetchOutData.id = id

	*right <- commandData{action: fetchData, value: fetchOutData, result: reply}

	result := (<-reply).(*fetchOutResult)
	return result.value, result.found
}

// Remove 清除数据
func (right *MemoryCache) Remove(id string) {
	removeData := &removeData{}
	removeData.id = id

	*right <- commandData{action: remove, value: removeData}
}

// ClearAll 清除所有数据
func (right *MemoryCache) ClearAll() {

	*right <- commandData{action: clearAll}
}

// Release 释放Cache
func (right *MemoryCache) Release() {
	*right <- commandData{action: end}

	close(*right)
}

func (right *MemoryCache) run() {
	_cacheData := make(map[string]cacheData)

	for command := range *right {
		switch command.action {
		case putData:
			id := strings.ToLower(util.RandomAlphanumeric(32))

			cacheData := cacheData{}
			cacheData.putInData = *(command.value.(*putInData))
			cacheData.cacheTime = time.Now()

			_cacheData[id] = cacheData

			result := &putInResult{}
			result.value = id

			command.result <- result
		case fetchData:
			id := command.value.(*fetchOutData).id

			cacheData, found := _cacheData[id]

			result := &fetchOutResult{}
			result.found = found
			if found {
				cacheData.cacheTime = time.Now()
				_cacheData[id] = cacheData

				result.value = cacheData.data
			}

			command.result <- result
		case remove:
			id := command.value.(*removeData).id

			delete(_cacheData, id)
		case clearAll:
			_cacheData = make(map[string]cacheData)

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

func (right *MemoryCache) checkTimeOut() {
	timeOutTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeOutTimer.C:
			*right <- commandData{action: checkTimeOut}
		}
	}
}
