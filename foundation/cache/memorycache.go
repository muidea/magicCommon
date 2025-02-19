package cache

import (
	"context"
	"math"
	"strings"
	"sync"
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
func NewCache(cleanCallBack ExpiredCleanCallBackFunc) Cache {
	cacheCtx, cacheCancel := context.WithCancel(context.Background())
	cache := &MemoryCache{
		commandChannel:       make(chan commandData, 100),
		cancelFunc:           cacheCancel,
		expiredCleanCallBack: cleanCallBack,
		rwLock:               new(sync.RWMutex),
	}

	// 启动多个worker处理命令
	for i := 0; i < 4; i++ {
		cache.cacheWg.Add(1)
		go cache.run()
	}

	cache.cacheWg.Add(1)
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
	commandChannel       chan commandData
	cancelFunc           context.CancelFunc
	cacheWg              sync.WaitGroup
	localCacheData       sync.Map
	pool                 sync.Pool
	expiredCleanCallBack ExpiredCleanCallBackFunc
	rwLock               *sync.RWMutex
}

// Put 投放数据，返回数据的唯一标示
func (s *MemoryCache) Put(data interface{}, maxAge float64) string {
	dataPtr := &putInData{
		data:   data,
		maxAge: maxAge,
	}

	result := s.sendCommand(commandData{action: putIn, value: dataPtr}).(*putInResult)
	return result.value
}

// Fetch 获取数据
func (s *MemoryCache) Fetch(id string) interface{} {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	v, found := s.localCacheData.Load(id)
	if !found {
		return nil
	}

	dataPtr := v.(cacheData)
	if s.isExpired(dataPtr) {
		s.localCacheData.Delete(id)
		return nil
	}

	return dataPtr.cacheData.data
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

	// 为每个worker发送end命令
	for i := 0; i < 4; i++ {
		s.sendCommand(commandData{action: end})
	}

	s.cacheWg.Wait()
	close(s.commandChannel)
}

func (s *MemoryCache) sendCommand(command commandData) interface{} {
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

func (s *MemoryCache) run() {
	defer s.cacheWg.Done()

	for command := range s.commandChannel {
		switch command.action {
		case putIn:
			s.rwLock.Lock()
			id := strings.ToLower(util.RandomAlphanumeric(32))
			dataPtr := cacheData{
				cacheData: command.value.(*putInData),
				cacheTime: time.Now(),
			}
			s.localCacheData.Store(id, dataPtr)
			s.rwLock.Unlock()

			result := &putInResult{value: id}
			command.result <- result

		case fetchOut:
			id := command.value.(*fetchOutData).id
			v, found := s.localCacheData.Load(id)

			result := &fetchOutResult{}
			if found {
				dataPtr := v.(cacheData)
				dataPtr.cacheTime = time.Now()
				s.localCacheData.Store(id, dataPtr)
				result.value = dataPtr.cacheData.data
			}

			command.result <- result

		case search:
			opr := command.value.(*searchData).opr

			result := &searchResult{}
			s.localCacheData.Range(func(k, v interface{}) bool {
				dataPtr := v.(cacheData)
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
			id := command.value.(*removeData).id
			s.localCacheData.Delete(id)
			command.result <- true

		case clearAll:
			s.localCacheData.Range(func(k, v interface{}) bool {
				s.localCacheData.Delete(k)
				return true
			})
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

func (s *MemoryCache) getExpiredKeys() []string {
	keys := []string{}
	s.localCacheData.Range(func(k, v interface{}) bool {
		dataPtr := v.(cacheData)
		if math.Abs(dataPtr.cacheData.maxAge-ForeverAgeValue) > 0.001 {
			current := time.Now()
			elapse := current.Sub(dataPtr.cacheTime).Minutes()
			if elapse > dataPtr.cacheData.maxAge {
				keys = append(keys, k.(string))
			}
		}
		return true
	})
	return keys
}

func (s *MemoryCache) checkTimeOut(ctx context.Context) {
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

func (s *MemoryCache) isExpired(data cacheData) bool {
	if math.Abs(data.cacheData.maxAge-ForeverAgeValue) > 0.001 {
		return time.Since(data.cacheTime).Minutes() > data.cacheData.maxAge
	}
	return false
}
