package event

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// This benchmark uses baseline-compatible APIs so the identical file can run
// before and after the change. Cold includes one cache publication per lookup;
// hot reuses one event/destination pair. It measures routing, not handler work.
func BenchmarkTopicRoutingScale(b *testing.B) {
	for _, n := range []int{1, 100, 1000} {
		for _, size := range []int{64, 65535} {
			for _, mode := range []string{"exact", "wildcard", "overlap", "miss"} {
				for _, cache := range []string{"cold", "hot"} {
					b.Run(fmt.Sprintf("n=%d/bytes=%d/%s/%s", n, size, mode, cache), func(b *testing.B) {
						hub := NewHub(8).(*hubImpl)
						defer hub.Terminate(context.Background())
						topic := "a/" + strings.Repeat("x", size-2)
						observer := &benchNoLogObserver{id: "owner"}
						for i := 0; i < n; i++ {
							filter := fmt.Sprintf("other/%d", i)
							if mode == "wildcard" {
								filter = fmt.Sprintf("other/%d/+", i)
							}
							if i == 0 {
								switch mode {
								case "exact", "overlap":
									filter = topic
								case "wildcard":
									filter = "a/+"
								}
							}
							if err := hub.Subscribe(filter, observer); err != nil {
								b.Fatal(err)
							}
						}
						if mode == "overlap" {
							for _, filter := range []string{"a/+", "a/#"} {
								if err := hub.Subscribe(filter, observer); err != nil {
									b.Fatal(err)
								}
							}
						}
						ev := NewEvent(topic, "caller", "owner", nil, nil)
						hub.matchingObservers(ev)
						b.ReportAllocs()
						b.ResetTimer()
						for i := 0; i < b.N; i++ {
							if cache == "cold" {
								hub.eventMatchCacheLock.Lock()
								hub.eventMatchCache = map[string]ObserverList{}
								hub.eventMatchCacheLock.Unlock()
							}
							hub.matchingObservers(ev)
						}
						b.StopTimer()
					})
				}
			}
		}
	}
}
