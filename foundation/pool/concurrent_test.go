package pool

import (
	"sync"
	"testing"
	"time"
)

// TestHighConcurrencyPreCreation 测试高并发场景下的预创建行为
func TestHighConcurrencyPreCreation(t *testing.T) {
	creationCount := 0
	var mu sync.Mutex
	factory := func() (int, error) {
		mu.Lock()
		creationCount++
		currentCount := creationCount
		mu.Unlock()
		time.Sleep(1 * time.Millisecond) // 模拟较短的创建时间
		return currentCount, nil
	}

	// 创建池，初始容量为5，最大容量为15
	pool, err := New(factory, 5, 15)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close(func(resource int) {})

	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan int, 50)

	// 启动10个并发客户端
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			<-start // 等待所有goroutine就绪

			// 每个客户端获取并立即返回资源，模拟高频使用
			for j := 0; j < 3; j++ {
				resource, err := pool.Get()
				if err != nil {
					t.Logf("Client %d failed to get resource: %v", clientID, err)
					continue
				}
				results <- resource
				time.Sleep(1 * time.Millisecond) // 模拟短暂使用
				if err := pool.Put(resource); err != nil {
					t.Errorf("Client %d failed to put resource: %v", clientID, err)
					return
				}
			}
		}(i)
	}

	// 同时启动所有客户端
	close(start)
	wg.Wait()
	close(results)

	// 统计获取的资源数量
	count := 0
	for range results {
		count++
	}

	mu.Lock()
	finalCreationCount := creationCount
	mu.Unlock()

	t.Logf("Total operations: %d, Total creations: %d", count, finalCreationCount)

	// 验证创建数量不超过maxSize
	if finalCreationCount > 15 {
		t.Errorf("Expected at most 15 creations, got %d", finalCreationCount)
	}

	// 验证大部分操作都成功完成
	if count < 25 { // 10 clients * 3 operations each = 30, 允许少量失败
		t.Errorf("Expected at least 25 operations, got %d", count)
	}
}

// TestConcurrentPreCreationEdgeCase 测试边界情况下的并发预创建
func TestConcurrentPreCreationEdgeCase(t *testing.T) {
	creationCount := 0
	var mu sync.Mutex
	factory := func() (int, error) {
		mu.Lock()
		creationCount++
		currentCount := creationCount
		mu.Unlock()
		time.Sleep(5 * time.Millisecond) // 模拟创建耗时
		return currentCount, nil
	}

	// 创建池，初始容量为1，最大容量为3
	pool, err := New(factory, 1, 3)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close(func(resource int) {})

	var wg sync.WaitGroup

	// 场景：多个客户端同时获取资源，触发预创建
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			resource, err := pool.Get()
			if err != nil {
				t.Logf("Client %d got expected error: %v", clientID, err)
				return
			}
			t.Logf("Client %d got resource: %d", clientID, resource)
			time.Sleep(10 * time.Millisecond)
			if err := pool.Put(resource); err != nil {
				t.Errorf("Client %d failed to put resource: %v", clientID, err)
			}
		}(i)
	}

	wg.Wait()

	mu.Lock()
	finalCreationCount := creationCount
	mu.Unlock()

	t.Logf("Total creations in edge case: %d", finalCreationCount)

	// 验证创建数量不超过maxSize
	if finalCreationCount > 3 {
		t.Errorf("Expected at most 3 creations in edge case, got %d", finalCreationCount)
	}
}

// TestConcurrentGetPutStress 测试高压力下的并发获取和归还
func TestConcurrentGetPutStress(t *testing.T) {
	creationCount := 0
	var mu sync.Mutex
	factory := func() (int, error) {
		mu.Lock()
		creationCount++
		currentCount := creationCount
		mu.Unlock()
		return currentCount, nil
	}

	// 创建池，初始容量为5，最大容量为20
	pool, err := New(factory, 5, 20)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close(func(resource int) {})

	var wg sync.WaitGroup
	operations := 1000
	successCount := 0
	var successMu sync.Mutex

	// 启动多个goroutine进行高频率的获取和归还
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < operations/10; j++ {
				resource, err := pool.Get()
				if err != nil {
					t.Logf("Worker %d failed to get resource: %v", workerID, err)
					continue
				}
				successMu.Lock()
				successCount++
				successMu.Unlock()
				// 立即归还，模拟高频使用场景
				if err := pool.Put(resource); err != nil {
					t.Errorf("Worker %d failed to put resource: %v", workerID, err)
				}
			}
		}(i)
	}

	wg.Wait()

	mu.Lock()
	finalCreationCount := creationCount
	mu.Unlock()

	t.Logf("Stress test: %d successful operations, %d total creations", successCount, finalCreationCount)

	// 验证创建数量不超过maxSize
	if finalCreationCount > 20 {
		t.Errorf("Expected at most 20 creations in stress test, got %d", finalCreationCount)
	}

	// 验证大部分操作都成功完成
	if successCount < operations*9/10 { // 允许10%的失败率
		t.Errorf("Expected at least %d successful operations, got %d", operations*9/10, successCount)
	}
}
