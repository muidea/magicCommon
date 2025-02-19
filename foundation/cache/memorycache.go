package cache

import (
	"context"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicCommon/foundation/log"
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
	cacheCtx, cacheCancel := context.WithCancel(context.Background())
	cache := &MemoryCache{
		commandChannel: make(chan commandData, 10),
		cancelFunc:     cacheCancel,
	}

	cache.cacheWg.Add(2)
	go cache.run(cleanCallBack)
	go cache.checkTimeOut(cacheCtx)

	return cache
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

type searchResult fetchOutResult

type removeData struct {
	id string
}

type cacheData struct {
	cacheData *putInData
	cacheTime time.Time
}

// MemoryCache 内存缓存
type MemoryCache struct {
	commandChannel chan commandData
	cancelFunc     context.CancelFunc
	cacheWg        sync.WaitGroup
}

// Put 投放数据，返回数据的唯一标示
func (s *MemoryCache) Put(data interface{}, maxAge float64) string {
	dataPtr := &putInData{}
	dataPtr.data = data
	dataPtr.maxAge = maxAge

	result := s.sendCommand(commandData{action: putIn, value: dataPtr}).(*putInResult)
	return result.value
}

// Fetch 获取数据
func (s *MemoryCache) Fetch(id string) interface{} {
	dataPtr := &fetchOutData{}
	dataPtr.id = id

	result := s.sendCommand(commandData{action: fetchOut, value: dataPtr}).(*fetchOutResult)
	return result.value
}

func (s *MemoryCache) Search(opr SearchOpr) interface{} {
	if opr == nil {
		return nil
	}

	dataPtr := &searchData{}
	dataPtr.opr = opr

	result := s.sendCommand(commandData{action: search, value: dataPtr}).(*searchResult)
	return result.value
}

// Remove 清除数据
func (s *MemoryCache) Remove(id string) {
	dataPtr := &removeData{}
	dataPtr.id = id

	s.sendCommand(commandData{action: remove, value: dataPtr})
}

// ClearAll 清除所有数据
func (s *MemoryCache) ClearAll() {
	s.sendCommand(commandData{action: clearAll})
}

// Release 释放Cache
func (s *MemoryCache) Release() {
	s.cancelFunc()

	s.sendCommand(commandData{action: end})

	s.cacheWg.Wait()
	close(s.commandChannel)
}

func (s *MemoryCache) sendCommand(command commandData) interface{} {
	reply := make(chan interface{})
	command.result = reply
	s.commandChannel <- command
	return <-reply
}

func (s *MemoryCache) run(cleanCallBack CleanCallBackFunc) {
	localCacheData := make(map[string]cacheData)
	defer func() {
		log.Warnf("run, release cache")
		s.cacheWg.Done()
	}()

	for command := range s.commandChannel {
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
			opr := command.value.(*searchData).opr

			result := &searchResult{}
			for k, v := range localCacheData {
				if opr(v.cacheData.data) {
					v.cacheTime = time.Now()
					localCacheData[k] = v
					result.value = v.cacheData.data
					break
				}
			}

			command.result <- result
		case remove:
			id := command.value.(*removeData).id
			delete(localCacheData, id)
			command.result <- true
		case clearAll:
			localCacheData = make(map[string]cacheData)
			command.result <- true
		case checkTimeOut:
			keys := s.getExpiredKeys(localCacheData)
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
			command.result <- true
			return
		}
	}
}

func (s *MemoryCache) getExpiredKeys(localCacheData map[string]cacheData) []string {
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
	return keys
}

func (s *MemoryCache) checkTimeOut(ctx context.Context) {
	defer func() {
		log.Warnf("checkTimeOut, release cache")
		s.cacheWg.Done()
	}()

	timeOutTimer := time.NewTicker(5 * time.Second)
	defer timeOutTimer.Stop()

	for {
		select {
		case <-timeOutTimer.C:
			s.commandChannel <- commandData{action: checkTimeOut}
		case <-ctx.Done():
			log.Infof("checkTimeOut exit")
			return
		}
	}
}

