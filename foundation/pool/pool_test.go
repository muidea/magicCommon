package pool

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_BasicOperations(t *testing.T) {
	factory := func() (int, error) {
		return 42, nil
	}

	pool, err := New(factory, 1, 2)
	require.NoError(t, err)
	defer pool.Close(func(res int) {})

	// Get a resource from the pool
	resource, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, 42, resource)

	// Return the resource to the pool
	err = pool.Put(resource)
	assert.NoError(t, err)

	// Retrieve the resource again
	resource, err = pool.Get()
	require.NoError(t, err)
	assert.Equal(t, 42, resource)

	err = pool.Put(resource)
	assert.NoError(t, err)
}

func TestPool_ConcurrentAccess(t *testing.T) {
	factory := func() (int, error) {
		return 42, nil
	}

	pool, err := New(factory, 1, 20)
	require.NoError(t, err)
	defer pool.Close(func(res int) {})

	var wg sync.WaitGroup

	const numGoroutines = 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			resource, err := pool.Get()
			require.NoError(t, err)
			assert.Equal(t, 42, resource)
			err = pool.Put(resource)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

func TestPool_Close(t *testing.T) {
	factory := func() (int, error) {
		return 42, nil
	}

	releaseCalled := false
	releaseFunc := func(res int) {
		releaseCalled = true
		assert.Equal(t, 42, res)
	}

	pool, err := New(factory, 1, 1)
	require.NoError(t, err)

	resource, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, 42, resource)

	pool.Put(resource)
	pool.Close(releaseFunc)

	// Ensure that pool is closed
	_, err = pool.Get()
	assert.Error(t, err)
	assert.True(t, releaseCalled)
}

func TestPool_MaxSize(t *testing.T) {
	factory := func() (int, error) {
		return 42, nil
	}

	// 设置maxSize为1
	pool, err := New(factory, 1, 1)
	require.NoError(t, err)
	defer pool.Close(func(res int) {})

	// 获取第一个资源
	resource1, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, 42, resource1)

	// 尝试获取第二个资源(应该阻塞)
	var resource2 int
	var err2 error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		resource2, err2 = pool.Get()
	}()

	// 确保goroutine已启动
	time.Sleep(100 * time.Millisecond)

	// 归还第一个资源
	err = pool.Put(resource1)
	assert.NoError(t, err)

	// 等待goroutine完成
	wg.Wait()

	// 验证第二个资源获取成功
	require.NoError(t, err2)
	assert.Equal(t, 42, resource2)

	err = pool.Put(resource2)
	assert.NoError(t, err)
}

func TestPool_NewResourceCreation(t *testing.T) {
	factory := func() (int, error) {
		return 42, nil
	}

	// 设置maxSize为2
	pool, err := New(factory, 1, 2)
	require.NoError(t, err)
	defer pool.Close(func(res int) {})

	// 获取第一个资源
	resource1, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, 42, resource1)

	// 获取第二个资源(应该创建新资源)
	resource2, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, 42, resource2)

	// 尝试获取第三个资源(应该阻塞)
	var resource3 int
	var err3 error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		resource3, err3 = pool.Get()
	}()

	// 确保goroutine已启动
	time.Sleep(100 * time.Millisecond)

	// 归还第一个资源
	err = pool.Put(resource1)
	assert.NoError(t, err)

	// 等待goroutine完成
	wg.Wait()

	// 验证第三个资源获取成功
	require.NoError(t, err3)
	assert.Equal(t, 42, resource3)

	err = pool.Put(resource2)
	assert.NoError(t, err)
	err = pool.Put(resource3)
	assert.NoError(t, err)
}
