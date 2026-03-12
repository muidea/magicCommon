package event

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// benchNoLogObserver：不打印日志的轻量 Observer
type benchNoLogObserver struct {
	id string
}

func (o *benchNoLogObserver) ID() string { return o.id }

func (o *benchNoLogObserver) Notify(ev Event, re Result) {
	if re != nil {
		re.Set(ev.Data(), nil)
	}
}

// benchHeavyObserver 用于 Send 重 handler 基准测试
type benchHeavyObserver struct {
	id string
}

func (h *benchHeavyObserver) ID() string { return h.id }

func (h *benchHeavyObserver) Notify(ev Event, re Result) {
	x := 0
	for i := 0; i < 200; i++ {
		x += i * i
	}
	if re != nil {
		re.Set(x, nil)
	}
	_ = x
}

// BenchmarkHubPostHighThroughput 基准：单 Observer，多 goroutine 并发 Post 事件
func BenchmarkHubPostHighThroughput(b *testing.B) {
	hub := NewHub(1024)
	defer hub.Terminate()

	observer := newSequenceObserver("bench-post-observer")
	eventID := "/bench/post"
	hub.Subscribe(eventID, observer)

	// 等待订阅完成
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		seq := 0
		for pb.Next() {
			ev := NewEvent(eventID, "bench-source", observer.id, NewValues(), nil)
			ev.SetData("sequence", seq)
			seq++
			hub.Post(ev)
		}
	})
}

// BenchmarkHubSendHighThroughput 基准：单 Observer，多 goroutine 并发 Send 事件（同步）
func BenchmarkHubSendHighThroughput(b *testing.B) {
	hub := NewHub(1024)
	defer hub.Terminate()

	handler := &eventHandler{handlerID: "/bench-send-observer"}
	eventID := "/bench/send"
	hub.Subscribe(eventID, handler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		seq := 0
		for pb.Next() {
			ev := NewEvent(eventID, "bench-source", handler.ID(), NewValues(), nil)
			ev.SetData("sequence", seq)
			seq++
			_ = hub.Send(ev)
		}
	})
}

// BenchmarkHubSendNoLog 基准：不打印日志的轻量级 Send 场景
func BenchmarkHubSendNoLog(b *testing.B) {
	hub := NewHubWithOptions(
		1024,
		WithPerDestinationChanSize(256),
		WithHubActionChanSize(512),
		WithWorkerPoolSize(256),
	)
	defer hub.Terminate()

	// 自定义 Observer，不做日志，只设置结果
	observer := &benchNoLogObserver{id: "/bench-send-nolog"}
	eventID := "/bench/send/nolog"
	hub.Subscribe(eventID, observer)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ev := NewEvent(eventID, "bench-source", observer.ID(), NewValues(), nil)
			_ = hub.Send(ev)
		}
	})
}

// BenchmarkHubSendLightHandler 基准：轻量级业务处理（少量 map 操作），贴近一般读写型 handler
func BenchmarkHubSendLightHandler(b *testing.B) {
	hub := NewHubWithOptions(
		1024,
		WithPerDestinationChanSize(256),
		WithHubActionChanSize(512),
		WithWorkerPoolSize(256),
	)
	defer hub.Terminate()

	type lightHandler struct {
		eventHandler
	}

	handler := &lightHandler{eventHandler{handlerID: "/bench-send-light"}}
	eventID := "/bench/send/light"
	hub.Subscribe(eventID, handler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		values := NewValues()
		for pb.Next() {
			ev := NewEvent(eventID, "bench-source", handler.ID(), values, nil)
			// 模拟轻量业务：少量 header/data 读写
			h := ev.Header()
			h.Set("k1", time.Now().UnixNano())
			_ = h.Get("k1")
			_ = ev.ID()
			_ = hub.Send(ev)
		}
	})
}

// BenchmarkHubSendHeavyHandler 基准：CPU 密集型 handler，模拟更重的业务逻辑
func BenchmarkHubSendHeavyHandler(b *testing.B) {
	hub := NewHubWithOptions(
		1024,
		WithPerDestinationChanSize(256),
		WithHubActionChanSize(512),
		WithWorkerPoolSize(256),
	)
	defer hub.Terminate()

	// 使用自定义 Observer，在 Notify 中模拟较重的 CPU 逻辑
	handler := &benchHeavyObserver{id: "/bench-send-heavy"}
	eventID := "/bench/send/heavy"
	hub.Subscribe(eventID, handler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ev := NewEvent(eventID, "bench-source", handler.ID(), NewValues(), nil)
			_ = hub.Send(ev)
		}
	})
}

// BenchmarkHubPostManySubscribers 基准：大量订阅者场景下的 Post 吞吐
func BenchmarkHubPostManySubscribers(b *testing.B) {
	const subscriberCount = 64

	hub := NewHub(2048)
	defer hub.Terminate()

	eventID := "/bench/many"

	subs := make([]*sequenceObserver, subscriberCount)
	for i := 0; i < subscriberCount; i++ {
		subs[i] = newSequenceObserver(fmt.Sprintf("bench-sub-%d", i))
		hub.Subscribe(eventID, subs[i])
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		seq := 0
		for pb.Next() {
			ev := NewEvent(eventID, "bench-source", "#", NewValues(), nil)
			ev.SetData("sequence", seq)
			seq++
			hub.Post(ev)
		}
	})
}

// BenchmarkHubPostHighConcurrencyPublishers 基准：多个发布者 + 单订阅者场景
func BenchmarkHubPostHighConcurrencyPublishers(b *testing.B) {
	hub := NewHub(1024)
	defer hub.Terminate()

	observer := newSequenceObserver("bench-concurrent-observer")
	eventID := "/bench/concurrent"
	hub.Subscribe(eventID, observer)

	b.ResetTimer()

	var wg sync.WaitGroup
	publishers := b.N

	wg.Add(publishers)
	for p := 0; p < publishers; p++ {
		go func(publisherID int) {
			defer wg.Done()
			ev := NewEvent(eventID, fmt.Sprintf("publisher-%d", publisherID), observer.id, NewValues(), nil)
			ev.SetData("sequence", publisherID)
			hub.Post(ev)
		}(p)
	}

	wg.Wait()
}

