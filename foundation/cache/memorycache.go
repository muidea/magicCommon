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
func NewCache(cleanCallBack CleanCallBackFunc) Cache {
	cache := make(MemoryCache)

	go cache.run(cleanCallBack)
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
	cacheData *putInData
	cacheTime time.Time
}

// MemoryCache 内存缓存
type MemoryCache chan commandData

// Put 投放数据，返回数据的唯一标示
func (s *MemoryCache) Put(data interface{}, maxAge float64) string {

	reply := make(chan interface{})

	dataPtr := &putInData{}
	dataPtr.data = data
	dataPtr.maxAge = maxAge

	*s <- commandData{action: putIn, value: dataPtr, result: reply}

	result := (<-reply).(*putInResult).value
	return result
}

// Fetch 获取数据
func (s *MemoryCache) Fetch(id string) interface{} {

	reply := make(chan interface{})

	dataPtr := &fetchOutData{}
	dataPtr.id = id

	*s <- commandData{action: fetchOut, value: dataPtr, result: reply}

	result := (<-reply).(*fetchOutResult)
	return result.value
}

func (s *MemoryCache) Search(opr SearchOpr) interface{} {
	if opr == nil {
		return nil
	}

	reply := make(chan interface{})

	dataPtr := &searchData{}
	dataPtr.opr = opr

	*s <- commandData{action: search, value: dataPtr, result: reply}

	result := (<-reply).(*searchResult)
	return result.value
}

// Remove 清除数据
func (s *MemoryCache) Remove(id string) {
	dataPtr := &removeData{}
	dataPtr.id = id

	*s <- commandData{action: remove, value: dataPtr}
}

// ClearAll 清除所有数据
func (s *MemoryCache) ClearAll() {

	*s <- commandData{action: clearAll}
}

// Release 释放Cache
func (s *MemoryCache) Release() {
	*s <- commandData{action: end}

	close(*s)
}

func (s *MemoryCache) run(cleanCallBack CleanCallBackFunc) {
	localCacheData := make(map[string]cacheData)

	for command := range *s {
		switch command.action {
		case putIn:
			id := strings.ToLower(util.RandomAlphanumeric(32))

			dataPtr := cacheData{}
			dataPtr.cacheData = command.value.(*putInData)
			dataPtr.cacheTime = time.Now()

			localCacheData[id] = dataPtr

			result := &putInResult{}
			result.value = id

			command.result <- result
		case fetchOut:
			id := command.value.(*fetchOutData).id
			dataPtr, found := localCacheData[id]

			result := &fetchOutResult{}
			if found {
				dataPtr.cacheTime = time.Now()
				localCacheData[id] = dataPtr

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
			id := command.value.(*removeData).id

			delete(localCacheData, id)
		case clearAll:
			localCacheData = make(map[string]cacheData)
		case checkTimeOut:
			keys := []string{}
			// 检查每项数据是否超时，超时数据需要主动清除掉
			for k, v := range localCacheData {
				if math.Abs(v.cacheData.maxAge-ForeverAgeValue) > 0.001 {
					current := time.Now()
					elapse := current.Sub(v.cacheTime).Minutes()
					if elapse > v.cacheData.maxAge {
						keys = append(keys, k)
						//delete(localCacheData, k)
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

func (s *MemoryCache) checkTimeOut() {
	timeOutTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeOutTimer.C:
			*s <- commandData{action: checkTimeOut}
		}
	}
}
