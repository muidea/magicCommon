package pool

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_BasicOperations(t *testing.T) {
	factory := func() (int, error) {
		return 42, nil
	}

	pool, err := New(factory, 2)
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
}

func TestPool_ConcurrentAccess(t *testing.T) {
	factory := func() (int, error) {
		return 42, nil
	}

	pool, err := New(factory, 2)
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

	pool, err := New(factory, 1)
	require.NoError(t, err)

	resource, err := pool.Get()
	require.NoError(t, err)
	assert.Equal(t, 42, resource)

	pool.Close(releaseFunc)

	// Ensure that pool is closed
	_, err = pool.Get()
	assert.Error(t, err)
	assert.True(t, releaseCalled)
}
