package profiling

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWindowCountsConcurrentFailuresAndDrains(t *testing.T) {
	start := time.Now()
	w := NewWindow(start)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				w.Record("sql", "select", 5*time.Millisecond, j%10 == 0)
			}
		}()
	}
	wg.Wait()
	result := w.Drain(start.Add(time.Minute))
	e := result.Entries[0]
	if len(result.Entries) != 1 || e.Calls != 1000 || e.Errors != 100 || e.TotalMS != 5000 || e.MaxMS != 5 || e.Buckets[1] != 1000 {
		t.Fatalf("incorrect aggregate: %+v", result)
	}
	if len(w.Drain(start.Add(2*time.Minute)).Entries) != 0 {
		t.Fatal("drain repeated counts")
	}
}
func TestWindowBoundsCardinality(t *testing.T) {
	w := NewWindow(time.Now())
	for i := 0; i < 1000; i++ {
		w.Record("test", fmt.Sprint(i), 0, false)
	}
	s := w.Drain(time.Now())
	var count uint64
	for _, e := range s.Entries {
		count += e.Calls
	}
	if len(s.Entries) > maxKeys+1 || count != 1000 {
		t.Fatalf("unbounded or lost counts: %d %d", len(s.Entries), count)
	}
}
func TestHTTPOperationRemovesSecretsAndIdentities(t *testing.T) {
	a := HTTPOperation("GET", "https://user:password@private/namespaces/t001/application/query/123?token=secret")
	b := HTTPOperation("GET", "http://other/namespaces/t002/application/query/456")
	if a != b || a != "GET /namespaces/_/application/query/_" {
		t.Fatalf("unsafe operation: %s %s", a, b)
	}
}
