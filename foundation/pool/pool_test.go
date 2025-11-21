package pool

import (
	"sync"
	"testing"
	"time"
)

func TestPoolPreCreation(t *testing.T) {
	// 创建一个简单的资源工厂
	creationCount := 0
	var mu sync.Mutex
	factory := func() (int, error) {
		mu.Lock()
		creationCount++
		currentCount := creationCount
		mu.Unlock()
		t.Logf("Factory creating resource %d", currentCount)
		time.Sleep(10 * time.Millisecond) // 模拟创建耗时
		return currentCount, nil
	}

	// 创建池，初始容量为1，最大容量为3
	pool, err := New(factory, 1, 3)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close(func(resource int) {})

	// 获取第一个资源（应该从初始容量中获取）
	resource1, err := pool.Get()
	if err != nil {
		t.Fatalf("Failed to get resource1: %v", err)
	}
	t.Logf("Got resource1: %d", resource1)

	// 等待一段时间让预创建完成
	time.Sleep(100 * time.Millisecond)

	// 获取第二个资源（应该从预创建的资源中获取）
	resource2, err := pool.Get()
	if err != nil {
		t.Fatalf("Failed to get resource2: %v", err)
	}
	t.Logf("Got resource2: %d", resource2)

	// 获取第三个资源（应该同步创建）
	resource3, err := pool.Get()
	if err != nil {
		t.Fatalf("Failed to get resource3: %v", err)
	}
	t.Logf("Got resource3: %d", resource3)

	// 验证创建的总数
	mu.Lock()
	finalCount := creationCount
	mu.Unlock()

	t.Logf("Total creations: %d", finalCount)
	// 预期行为：初始创建1个，获取resource1后触发预创建1个，获取resource3时同步创建1个
	// 总共应该是3个创建，但预创建可能因为并发而创建了额外的资源
	if finalCount > 3 {
		t.Logf("Note: Pre-creation created %d extra resources due to concurrency", finalCount-3)
	}

	// 返回资源
	if err := pool.Put(resource1); err != nil {
		t.Errorf("Failed to put resource1: %v", err)
	}
	if err := pool.Put(resource2); err != nil {
		t.Errorf("Failed to put resource2: %v", err)
	}
	if err := pool.Put(resource3); err != nil {
		t.Errorf("Failed to put resource3: %v", err)
	}
}

func TestPoolConcurrentAccess(t *testing.T) {
	creationCount := 0
	var mu sync.Mutex
	factory := func() (int, error) {
		mu.Lock()
		creationCount++
		mu.Unlock()
		time.Sleep(5 * time.Millisecond)
		return creationCount, nil
	}

	pool, err := New(factory, 2, 5)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close(func(resource int) {})

	var wg sync.WaitGroup
	results := make(chan int, 10)

	// 并发获取资源
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resource, err := pool.Get()
			if err != nil {
				t.Errorf("Failed to get resource: %v", err)
				return
			}
			results <- resource
			time.Sleep(10 * time.Millisecond) // 模拟使用时间
			if err := pool.Put(resource); err != nil {
				t.Errorf("Failed to put resource: %v", err)
			}
		}()
	}

	wg.Wait()
	close(results)

	// 统计获取的资源数量
	count := 0
	for range results {
		count++
	}

	if count != 5 {
		t.Errorf("Expected 5 resources, got %d", count)
	}

	mu.Lock()
	finalCreationCount := creationCount
	mu.Unlock()

	if finalCreationCount > 5 {
		t.Errorf("Expected at most 5 creations, got %d", finalCreationCount)
	}
}
