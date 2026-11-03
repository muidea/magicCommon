package session

import (
	"sync"
	"testing"
)

func TestBaseClientReleaseAndCloneHaveIndependentLifecycles(t *testing.T) {
	base := NewBaseClient("http://example.test")
	clone := base.Clone()
	transport := base.GetHTTPClient()
	if transport == nil || clone.GetHTTPClient() != transport {
		t.Fatal("clone lost configured HTTP client")
	}
	clone.Release()
	clone.Release()
	if clone.GetHTTPClient() != nil || base.GetHTTPClient() != transport {
		t.Fatal("releasing clone changed source lifecycle")
	}
	before := base.Clone()
	defer before.Release()
	base.Release()
	after := base.Clone()
	if after.GetHTTPClient() != nil || before.GetHTTPClient() != transport {
		t.Fatal("release either revived a client or invalidated a previous clone")
	}
	var zero BaseClient
	zero.Release()
	if zero.GetHTTPClient() != nil {
		t.Fatal("zero client unexpectedly acquired a transport")
	}
}

func TestBaseClientConcurrentReleaseReadAndClone(t *testing.T) {
	for range 20 {
		base := NewBaseClient("http://example.test")
		start := make(chan struct{})
		var workers sync.WaitGroup
		for i := range 8 {
			workers.Go(func() {
				<-start
				for range 50 {
					if i%2 == 0 {
						base.Release()
					} else {
						_ = base.GetHTTPClient()
						clone := base.Clone()
						clone.Release()
					}
				}
			})
		}
		close(start)
		workers.Wait()
		if base.GetHTTPClient() != nil {
			t.Fatal("released transport revived")
		}
	}
}
