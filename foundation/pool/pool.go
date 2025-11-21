package pool

import (
	"errors"
	"fmt"
	"sync"

	"github.com/muidea/magicCommon/foundation/log"
)

/*

pool, err := New(func() (int, error) {
    return 42, nil // 简单示例：创建一个整数资源
}, 10)
if err != nil {
    log.Fatalf("Failed to create pool: %v", err)
}


resource, err := pool.Get()
if err != nil {
    log.Fatalf("Failed to get resource: %v", err)
}

err = pool.Put(resource)
if err != nil {
    log.Printf("Failed to return resource: %v", err)
}

pool.Close(func(resource int) {
    fmt.Printf("Releasing resource: %d\n", resource)
})

*/

// Pool represents a generic pool of resources.
type Pool[T any] struct {
	mu          *sync.Mutex
	cond        *sync.Cond
	factory     func() (T, error) // Factory function to create new resources
	idleQueue   []T               // Resource queue
	totalSize   int
	busyCount   int
	maxSize     int
	closed      bool // Flag to indicate if the pool is closed
	preCreating bool // Flag to indicate if pre-creation is in progress
}

// New creates a new Pool.
func New[T any](factory func() (T, error), initialCapacity, maxSize int) (*Pool[T], error) {
	if factory == nil {
		return nil, errors.New("factory function cannot be nil")
	}

	pool := &Pool[T]{
		mu:        &sync.Mutex{},
		factory:   factory,
		idleQueue: make([]T, initialCapacity),
		maxSize:   maxSize,
		totalSize: 0,
	}
	for idx := range initialCapacity {
		tVal, tErr := factory()
		if tErr != nil {
			return nil, tErr
		}
		pool.idleQueue[idx] = tVal
		pool.totalSize++
	}
	pool.cond = sync.NewCond(pool.mu)

	return pool, nil
}

// Get retrieves a resource from the pool, creating a new one if necessary.
func (s *Pool[T]) Get() (ret T, err error) {
	getOK := false
	//idleSize := 0
	getFunc := func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		// 如果池已关闭，直接返回错误
		if s.closed {
			err = fmt.Errorf("pool is closed")
			return
		}

		if len(s.idleQueue) > 0 {
			getOK = true
			// 从队列中获取资源
			ret = s.idleQueue[len(s.idleQueue)-1]
			s.idleQueue = s.idleQueue[:len(s.idleQueue)-1]
			//idleSize = len(s.idleQueue)
			s.busyCount++

			// 检查是否需要预创建：空闲队列为空且总数未达最大值且没有正在进行的预创建
			if len(s.idleQueue) == 0 && s.totalSize < s.maxSize && !s.preCreating {
				s.preCreating = true
				go s.preCreateResource()
			}
			return
		}

		if s.totalSize < s.maxSize {
			tVal, tErr := s.factory()
			if tErr != nil {
				err = tErr
				return
			}

			getOK = true
			s.totalSize++
			ret = tVal
			//idleSize = len(s.idleQueue)
			s.busyCount++
			return
		}
		s.cond.Wait()
	}

	for {
		getFunc()
		if getOK || err != nil {
			//log.Infof("get resource from pool, maxSize:%d, totalSize:%d, idleQueueSize:%d busySize:%d", s.maxSize, s.totalSize, idleSize, s.busyCount)
			return
		}
	}
}

// Put returns a resource to the pool.
func (s *Pool[T]) Put(tVal T) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		err = errors.New("pool is closed")
		return
	}

	s.busyCount--
	s.idleQueue = append(s.idleQueue, tVal)
	if err != nil {
		return
	}

	//log.Infof("put resource to pool, maxSize:%d, totalSize:%d, idleQueueSize:%d busySize:%d", s.maxSize, s.totalSize, len(s.idleQueue), s.busyCount)

	s.cond.Signal()
	return
}

// preCreateResource asynchronously creates a resource and adds it to the idle queue.
func (s *Pool[T]) preCreateResource() {
	s.mu.Lock()
	// 在预创建开始前再次检查条件，避免竞态条件
	if s.closed || s.totalSize >= s.maxSize {
		s.preCreating = false
		s.mu.Unlock()
		return
	}
	// 在锁的保护下递增totalSize，确保不会超过maxSize
	s.totalSize++
	s.mu.Unlock()

	tVal, tErr := s.factory()
	if tErr != nil {
		log.Errorf("Failed to pre-create resource: %v", tErr)
		s.mu.Lock()
		s.totalSize-- // 回滚totalSize
		s.preCreating = false
		s.mu.Unlock()
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果池已关闭，不添加新资源
	if s.closed {
		s.totalSize-- // 回滚totalSize
		s.preCreating = false
		return
	}

	// 添加预创建的资源到空闲队列
	s.idleQueue = append(s.idleQueue, tVal)
	s.cond.Signal() // 通知等待的goroutine有新资源可用

	// 重置预创建状态
	s.preCreating = false
}

// Close shuts down the pool and releases all resources.
func (s *Pool[T]) Close(releaseFunc func(T)) {
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.closed {
			return
		}

		s.closed = true
		// Release remaining resources
		for _, tVal := range s.idleQueue {
			releaseFunc(tVal)
		}
		// Log warning if not all resources were returned
		if len(s.idleQueue) < s.totalSize {
			log.Warnf("Warning: %d resources were not returned to the pool before close", s.totalSize-len(s.idleQueue))
		}
		s.idleQueue = nil
	}()

	s.cond.Broadcast()
}
