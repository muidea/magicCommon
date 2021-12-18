package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	cache := NewCache()
	if nil == cache {
		t.Error("create new memorycache failed")
		return
	}

	time.Sleep(100)
	data := "memorycache"
	id := cache.Put(data, 0.000)
	if len(id) == 0 {
		t.Error("Put data to memorycache failed")
	}

	timeOutTimer := time.NewTicker(6 * time.Second)
	select {
	case <-timeOutTimer.C:
	}
	_, found := cache.Fetch(id)
	if found {
		t.Error("memorycache maxAge unexpect.")
	}

	id = cache.Put(data, 2)
	if len(id) == 0 {
		t.Error("Put data to memorycache failed")
	}
	time.Sleep(100)

	val, found := cache.Fetch(id)
	if !found {
		t.Error("memorycache Fetch unexpect.")
	}

	if data != val.(string) {
		t.Error("Fetchout unexpect data")
	}

	cache.Remove(id)
	_, found = cache.Fetch(id)
	if found {
		t.Error("memorycache maxAge unexpect.")
	}
	time.Sleep(10000)

	cache.Release()
}
