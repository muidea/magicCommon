package cache

import (
	"testing"
	"time"
)

type Demo struct {
	strValue string
	intValue int
}

func TestKVCache(t *testing.T) {
	cache := NewKVCache(nil)
	if nil == cache {
		t.Error("create new kvcache failed")
		return
	}

	time.Sleep(100)

	kVal := "dKey"
	dPtr := &Demo{strValue: "abc", intValue: 100}
	key := cache.Put(kVal, dPtr, OneMinuteAgeValue)
	if key != kVal {
		t.Error("putIn kv value failed")
		return
	}

	timeOutTimer := time.NewTicker(6 * time.Second)
	select {
	case <-timeOutTimer.C:
	}

	vVal := cache.Fetch(kVal)
	if vVal == nil {
		t.Error("fetch kv value failed")
		return
	}
	if vVal.(*Demo).strValue != dPtr.strValue {
		t.Error("fetch kv value failed")
		return
	}
	cache.Remove(kVal)
	vVal = cache.Fetch(kVal)
	if vVal != nil {
		t.Error("fetch kv value failed")
		return
	}
}
