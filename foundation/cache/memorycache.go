package cache

import (
	"math"
	"strings"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
)

// Cache 缓存对象
type Cache interface {
	// maxAge单位minute
	Put(data interface{}, maxAge float64) string
	Fetch(id string) interface{}
	Search(opr SearchOpr) interface{}
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
}

type searchData struct {
	opr SearchOpr
}

type searchResult fetchOutKVResult

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

	*right <- commandData{action: putIn, value: putInData, result: reply}

	result := (<-reply).(*putInResult).value
	return result
}

// Fetch 获取数据
func (right *MemoryCache) Fetch(id string) interface{} {

	reply := make(chan interface{})

	fetchOutData := &fetchOutData{}
	fetchOutData.id = id

	*right <- commandData{action: fetchOut, value: fetchOutData, result: reply}

	result := (<-reply).(*fetchOutResult)
	return result.value
}

func (right *MemoryCache) Search(opr SearchOpr) interface{} {
	if opr == nil {
		return nil
	}

	reply := make(chan interface{})

	searchData := &searchData{}
	searchData.opr = opr

	*right <- commandData{action: search, value: searchData, result: reply}

	result := (<-reply).(*searchResult)
	return result.value
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
	localCacheData := make(map[string]cacheData)

	for command := range *right {
		switch command.action {
		case putIn:
			id := strings.ToLower(util.RandomAlphanumeric(32))

			cacheData := cacheData{}
			cacheData.putInData = *(command.value.(*putInData))
			cacheData.cacheTime = time.Now()

			localCacheData[id] = cacheData

			result := &putInResult{}
			result.value = id

			command.result <- result
		case fetchOut:
			id := command.value.(*fetchOutData).id

			cacheData, found := localCacheData[id]

			result := &fetchOutResult{}
			if found {
				cacheData.cacheTime = time.Now()
				localCacheData[id] = cacheData

				result.value = cacheData.data
			}

			command.result <- result
		case search:
			opr := command.value.(*searchKVData).opr

			result := &searchKVResult{}
			for _, v := range localCacheData {
				if opr(v.data) {
					result.value = v.data
					break
				}
			}

			command.result <- result
		case remove:
			id := command.value.(*removeData).id

			delete(localCacheData, id)
		case clearAll:
			localCacheData = make(map[string]cacheData)
		case checkTimeOut:
			// 检查每项数据是否超时，超时数据需要主动清除掉
			for k, v := range localCacheData {
				if math.Abs(v.maxAge-MaxAgeValue) > 0.001 {
					current := time.Now()
					elapse := current.Sub(v.cacheTime).Minutes()
					if elapse > v.maxAge {
						delete(localCacheData, k)
					}
				}
			}
		case end:
			localCacheData = nil
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
