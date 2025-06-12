package pool

import (
	"errors"
	"sync"
	"time"

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
	mu      sync.Mutex
	cond    *sync.Cond
	factory func() (T, error) // Factory function to create new resources
	queue   []T               // Resource queue
	closed  bool              // Flag to indicate if the pool is closed
}

// New creates a new Pool.
func New[T any](factory func() (T, error), initialCapacity int) (*Pool[T], error) {
	if factory == nil {
		return nil, errors.New("factory function cannot be nil")
	}

	pool := &Pool[T]{
		factory: factory,
		queue:   make([]T, initialCapacity),
	}
	for idx := range pool.queue {
		pool.queue[idx], _ = factory()
	}
	pool.cond = sync.NewCond(&pool.mu)

	return pool, nil
}

// Get retrieves a resource from the pool, creating a new one if necessary.
func (p *Pool[T]) Get() (T, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.queue) == 0 && !p.closed {
		p.cond.Wait()
	}

	if p.closed {
		var zero T
		return zero, errors.New("pool is closed")
	}

	resource := p.queue[len(p.queue)-1]
	p.queue = p.queue[:len(p.queue)-1]
	return resource, nil
}

// Put returns a resource to the pool.
func (p *Pool[T]) Put(resource T) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return errors.New("pool is closed")
	}

	p.queue = append(p.queue, resource)
	p.cond.Signal()
	return nil
}

// Close shuts down the pool and releases all resources.
func (p *Pool[T]) Close(releaseFunc func(T)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	// Wait for a grace period to allow resources to be returned
	gracePeriod := 5 // seconds
	timeout := make(chan struct{})
	go func() {
		time.Sleep(time.Duration(gracePeriod) * time.Second)
		close(timeout)
	}()

	for len(p.queue) < cap(p.queue) {
		select {
		case <-timeout:
			p.cond.Signal()
			break
		default:
			p.cond.Wait()
		}
	}

	// Release remaining resources
	for _, resource := range p.queue {
		releaseFunc(resource)
	}
	p.queue = nil

	// Log warning if not all resources were returned
	if len(p.queue) < cap(p.queue) {
		log.Warnf("Warning: %d resources were not returned to the pool before close", cap(p.queue)-len(p.queue))
	}

	p.cond.Broadcast()
}
