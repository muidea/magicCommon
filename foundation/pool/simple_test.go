package pool

import (
	"sync"
	"testing"
	"time"
)

func TestSimplePreCreation(t *testing.T) {
	creationCount := 0
	var mu sync.Mutex
	factory := func() (int, error) {
		mu.Lock()
		creationCount++
		currentCount := creationCount
		mu.Unlock()
		t.Logf("Creating resource %d", currentCount)
		return currentCount, nil
	}

	// 创建池，初始容量为0，最大容量为2
	pool, err := New(factory, 0, 2)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close(func(resource int) {})

	// 获取第一个资源（同步创建）
	resource1, err := pool.Get()
	if err != nil {
		t.Fatalf("Failed to get resource1: %v", err)
	}
	t.Logf("Got resource1: %d", resource1)

	// 此时空闲队列为空，应该触发预创建
	time.Sleep(50 * time.Millisecond)

	// 获取第二个资源（应该从预创建中获取）
	resource2, err := pool.Get()
	if err != nil {
		t.Fatalf("Failed to get resource2: %v", err)
	}
	t.Logf("Got resource2: %d", resource2)

	// 验证创建的总数
	mu.Lock()
	finalCount := creationCount
	mu.Unlock()

	t.Logf("Total creations: %d", finalCount)

	// 预期：resource1创建1个，预创建1个，总共2个
	if finalCount != 2 {
		t.Errorf("Expected 2 creations, got %d", finalCount)
	}

	// 返回资源
	if err := pool.Put(resource1); err != nil {
		t.Errorf("Failed to put resource1: %v", err)
	}
	if err := pool.Put(resource2); err != nil {
		t.Errorf("Failed to put resource2: %v", err)
	}
}
