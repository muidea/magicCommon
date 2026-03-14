package pool

import (
	"errors"
	"fmt"
	"sync"

	"log/slog"
)

/*
pool, err := New(func() (int, error) {
    return 42, nil // 简单示例：创建一个整数资源
}, WithInitialCapacity(5), WithMaxSize(20))
if err != nil {
    slog.Error("Failed to create pool", "error", err); os.Exit(1)
}


resource, err := pool.Get()
if err != nil {
    slog.Error("Failed to get resource", "error", err); os.Exit(1)
}

err = pool.Put(resource)
if err != nil {
    slog.Info("Failed to return resource", "error", err)
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

// PoolConfig holds configuration options for a Pool.
type PoolConfig struct {
	initialCapacity int
	maxSize         int
}

// PoolOption defines a function type for configuring a Pool.
type PoolOption func(*PoolConfig)

// WithInitialCapacity sets the initial capacity of the pool.
func WithInitialCapacity(capacity int) PoolOption {
	return func(c *PoolConfig) {
		c.initialCapacity = capacity
	}
}

// WithMaxSize sets the maximum size of the pool.
func WithMaxSize(maxSize int) PoolOption {
	return func(c *PoolConfig) {
		c.maxSize = maxSize
	}
}

// New creates a new Pool with the given factory function and options.
func New[T any](factory func() (T, error), opts ...PoolOption) (*Pool[T], error) {
	if factory == nil {
		return nil, errors.New("factory function cannot be nil")
	}

	// Default configuration
	config := &PoolConfig{
		initialCapacity: 0,
		maxSize:         10,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Validate configuration
	if config.initialCapacity < 0 {
		return nil, errors.New("initial capacity cannot be negative")
	}
	if config.maxSize <= 0 {
		return nil, errors.New("max size must be positive")
	}
	if config.initialCapacity > config.maxSize {
		return nil, errors.New("initial capacity cannot exceed max size")
	}

	pool := &Pool[T]{
		mu:        &sync.Mutex{},
		factory:   factory,
		idleQueue: make([]T, config.initialCapacity),
		maxSize:   config.maxSize,
		totalSize: 0,
	}
	for idx := range config.initialCapacity {
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
			//slog.Info("get resource from pool, maxSize:s.maxSize, totalSize:s.totalSize, idleQueueSize:idleSize busySize:s.busyCount", "field", s.maxSize, "error", s.totalSize, "id", idleSize, "key", s.busyCount)
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

	//slog.Info(fmt.Sprintf("put resource to pool, maxSize:%d, totalSize:%d, idleQueueSize:%d busySize:%d"s.maxSize, s.totalSize, len(s.idleQueue)), s.busyCount)

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
		slog.Error("Failed to pre-create resource: tErr", "field", tErr)
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
			if releaseFunc != nil {
				releaseFunc(tVal)
			}
		}
		// Log warning if not all resources were returned
		if len(s.idleQueue) < s.totalSize {
			slog.Warn("Warning: s.totalSize-len(s.idleQueue resources were not returned to the pool before close", "field", s.totalSize-len(s.idleQueue))
		}
		s.idleQueue = nil
	}()

	s.cond.Broadcast()
}
