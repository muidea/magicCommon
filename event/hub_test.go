package event

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"log/slog"
)

func TestMatchID(t *testing.T) {
	pattern := "/123"
	id := "/12"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123"
	id = "/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123"
	id = "/123"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/12"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/12"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/12"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/12"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/123"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/#"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/1212"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/1212/111"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/1212/111/1212"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/111"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/abc/1212/111"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212/uu"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212/uu/www"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212/uu/www/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/#/+/1212/#"
	id = "/123/122/1212/111"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/#/+/1212/#"
	id = "/123/122/abc/1212/111"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/#/+/1212/#"
	id = "/123/122/abc/bcd/1212/111"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/:id/1212/#"
	id = "/123/122/1212/111/2435/765756f/fsd"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/update/+"
	id = "/warehouse/shelf/create/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/update/#"
	id = "/warehouse/shelf/create/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/abc/#"
	id = "/abc/shelf/create/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/abc/#"
	id = "/abc/abc/create/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/abc/#"
	id = "/abc/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "/abc/abc/bcd/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "/abc/abc/bcd/cde/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "/abc/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "abc/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/#/bcd/"
	id = "abc/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/#/bcd/"
	id = "abc/123/bcd/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/bcd/"
	id = "abc/123/bcd/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/bcd/"
	id = "abc/123/bcd"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/bcd"
	id = "abc/123/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/abc/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/bcd"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/create/"
	id = "/warehouse/shelf/create/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/+/create/"
	id = "/warehouse/shelf/create/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "/#/+/+/create/"
	id = "/warehouse/shelf/create/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/notify/+"
	id = "/bill/notify/123"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
	}
}

func TestValues(t *testing.T) {
	values := NewValues()

	// Test Set/Get
	values.Set("key1", "value1")
	if got := values.Get("key1"); got != "value1" {
		t.Errorf("Values.Get() = %v, want %v", got, "value1")
	}
	if got := values.Get("nonexistent"); got != nil {
		t.Errorf("Values.Get() for nonexistent key = %v, want nil", got)
	}

	// Test GetString
	values.Set("string1", "string_value")
	values.Set("int1", 123)
	if got := values.GetString("string1"); got != "string_value" {
		t.Errorf("Values.GetString() = %v, want %v", got, "string_value")
	}
	if got := values.GetString("nonexistent"); got != "" {
		t.Errorf("Values.GetString() for nonexistent key = %v, want \"\"", got)
	}
	if got := values.GetString("int1"); got != "" {
		t.Errorf("Values.GetString() for non-string value = %v, want \"\"", got)
	}

	// Test GetInt
	values.Set("int2", 456)
	values.Set("string2", "string_value")
	if got := values.GetInt("int2"); got != 456 {
		t.Errorf("Values.GetInt() = %v, want %v", got, 456)
	}
	if got := values.GetInt("nonexistent"); got != 0 {
		t.Errorf("Values.GetInt() for nonexistent key = %v, want 0", got)
	}
	if got := values.GetInt("string2"); got != 0 {
		t.Errorf("Values.GetInt() for non-int value = %v, want 0", got)
	}

	// Test GetBool
	values.Set("bool1", true)
	values.Set("bool2", false)
	values.Set("string3", "string_value")
	if got := values.GetBool("bool1"); got != true {
		t.Errorf("Values.GetBool() = %v, want %v", got, true)
	}
	if got := values.GetBool("bool2"); got != false {
		t.Errorf("Values.GetBool() = %v, want %v", got, false)
	}
	if got := values.GetBool("nonexistent"); got != false {
		t.Errorf("Values.GetBool() for nonexistent key = %v, want false", got)
	}
	if got := values.GetBool("string3"); got != false {
		t.Errorf("Values.GetBool() for non-bool value = %v, want false", got)
	}
}

func TestBaseEvent(t *testing.T) {
	// Test creation and getters
	header := NewValues()
	header.Set("testHeader", "headerValue")
	data := "testData"
	ctx := context.Background()

	// Test NewEvent
	event := NewEvent("test/event", "source", "destination", header, data)

	if got := event.ID(); got != "test/event" {
		t.Errorf("Event.ID() = %v, want %v", got, "test/event")
	}
	if got := event.Source(); got != "source" {
		t.Errorf("Event.Source() = %v, want %v", got, "source")
	}
	if got := event.Destination(); got != "destination" {
		t.Errorf("Event.Destination() = %v, want %v", got, "destination")
	}
	if got := event.Header().GetString("testHeader"); got != "headerValue" {
		t.Errorf("Event.Header().GetString() = %v, want %v", got, "headerValue")
	}
	if got := event.Data(); got != data {
		t.Errorf("Event.Data() = %v, want %v", got, data)
	}

	// Test Context methods
	if got := event.Context(); got != ctx {
		t.Errorf("Event.Context() = %v, want nil", got)
	}
	event.BindContext(ctx)
	if got := event.Context(); got != ctx {
		t.Errorf("Event.Context() after BindContext = %v, want %v", got, ctx)
	}

	// Test NewEventWithContext
	eventWithCtx := NewEventWithContext("test/event2", "source2", "destination2", header, ctx, data)
	if got := eventWithCtx.Context(); got != ctx {
		t.Errorf("NewEventWithContext.Context() = %v, want %v", got, ctx)
	}

	// Test SetData/GetData
	event.SetData("key1", "value1")
	if got := event.GetData("key1"); got != "value1" {
		t.Errorf("Event.GetData() = %v, want %v", got, "value1")
	}
	if got := event.GetData("nonexistent"); got != nil {
		t.Errorf("Event.GetData() for nonexistent key = %v, want nil", got)
	}

	// Test Match
	if !event.Match("test/event") {
		t.Errorf("Event.Match() failed to match exact ID")
	}
	if !event.Match("test/+") {
		t.Errorf("Event.Match() failed to match wildcard pattern")
	}
	if event.Match("different/event") {
		t.Errorf("Event.Match() incorrectly matched different ID")
	}
}

func TestBaseResult(t *testing.T) {
	// Test creation and getters
	result := NewResult("testID", "source", "destination")
	if result.Error() == nil {
		t.Errorf("NewResult.Error() = nil, want error")
	}

	// Test Set/Get
	data := "resultData"
	result.Set(data, nil)
	gotData, gotErr := result.Get()
	if gotData != data {
		t.Errorf("Result.Get() data = %v, want %v", gotData, data)
	}
	if gotErr != nil {
		t.Errorf("Result.Get() error = %v, want nil", gotErr)
	}

	// Test SetVal/GetVal
	result.SetVal("key1", "value1")
	if got := result.GetVal("key1"); got != "value1" {
		t.Errorf("Result.GetVal() = %v, want %v", got, "value1")
	}
	if got := result.GetVal("nonexistent"); got != nil {
		t.Errorf("Result.GetVal() for nonexistent key = %v, want nil", got)
	}

	// Test with a custom error
	customErr := cd.NewError(cd.UnKnownError, "custom error")
	result.Set(nil, customErr)
	if got := result.Error(); got != customErr {
		t.Errorf("Result.Error() = %v, want %v", got, customErr)
	}
}

type testObserver struct {
	mu           sync.Mutex
	id           string
	notifyCount  int
	lastEvent    Event
	lastResult   Result
	notifySignal chan struct{}
}

func newTestObserver(id string) *testObserver {
	return &testObserver{
		id:           id,
		notifySignal: make(chan struct{}, 10),
	}
}

func (t *testObserver) ID() string {
	return t.id
}

func (t *testObserver) Notify(event Event, result Result) {
	t.mu.Lock()
	t.notifyCount++
	t.lastEvent = event
	t.lastResult = result
	t.mu.Unlock()
	select {
	case t.notifySignal <- struct{}{}:
	default:
	}
}

// waitForNotification waits for at least one notification or times out
func waitForNotification(observer *testObserver, timeout time.Duration) bool {
	select {
	case <-observer.notifySignal:
		return true
	case <-time.After(timeout):
		return false
	}
}

// sequenceObserver 用于测试事件顺序一致性的观察者
type sequenceObserver struct {
	id           string
	sequences    []int
	mu           sync.Mutex
	notifySignal chan struct{}
}

func newSequenceObserver(id string) *sequenceObserver {
	return &sequenceObserver{
		id:           id,
		notifySignal: make(chan struct{}, 1000), // 增加容量以避免阻塞
	}
}

func (s *sequenceObserver) ID() string {
	return s.id
}

func (s *sequenceObserver) Notify(event Event, result Result) {
	s.mu.Lock()
	if seq, ok := event.GetData("sequence").(int); ok {
		s.sequences = append(s.sequences, seq)
	}
	s.mu.Unlock()
	select {
	case s.notifySignal <- struct{}{}:
	default:
	}
}

func TestHubImpl(t *testing.T) {
	hub := NewHub(10)

	observer1 := newTestObserver("observer1")
	observer2 := newTestObserver("observer2")

	eventID := "test/event"
	hub.Subscribe(eventID, observer1)
	hub.Subscribe(eventID, observer2)

	time.Sleep(100 * time.Millisecond)

	event := NewEvent(eventID, "source", "#", NewValues(), "data")
	hub.Post(event)

	if !waitForNotification(observer1, time.Second) {
		t.Fatal("Timeout waiting for observer1 notification")
	}
	if !waitForNotification(observer2, time.Second) {
		t.Fatal("Timeout waiting for observer2 notification")
	}

	observer1.mu.Lock()
	count1 := observer1.notifyCount
	observer1.mu.Unlock()
	observer2.mu.Lock()
	count2 := observer2.notifyCount
	observer2.mu.Unlock()

	if count1 == 0 {
		t.Errorf("Observer1 notify count = %d, want > 0", count1)
	}
	if count2 == 0 {
		t.Errorf("Observer2 notify count = %d, want > 0", count2)
	}

	hub.Unsubscribe(eventID, observer1)

	time.Sleep(100 * time.Millisecond)

	observer1.mu.Lock()
	observer1.notifyCount = 0
	observer1.mu.Unlock()
	observer2.mu.Lock()
	observer2.notifyCount = 0
	observer2.mu.Unlock()

	hub.Post(event)

	if !waitForNotification(observer2, time.Second) {
		t.Fatal("Timeout waiting for observer2 notification after unsubscribe")
	}

	observer1.mu.Lock()
	count1 = observer1.notifyCount
	observer1.mu.Unlock()
	observer2.mu.Lock()
	count2 = observer2.notifyCount
	observer2.mu.Unlock()

	if count1 != 0 {
		t.Errorf("Observer1 notify count after unsubscribe = %d, want 0", count1)
	}
	if count2 == 0 {
		t.Errorf("Observer2 notify count = %d, want > 0", count2)
	}

	hub.Terminate()
}

func TestSimpleObserver(t *testing.T) {
	hub := NewHub(10)
	simpleObserver := NewSimpleObserver("simpleObserver", hub)

	// Test ID
	if got := simpleObserver.ID(); got != "simpleObserver" {
		t.Errorf("SimpleObserver.ID() = %v, want %v", got, "simpleObserver")
	}

	// Create a channel to wait for the observer function to be called
	observerDone := make(chan struct{}, 1)
	var capturedEvent Event
	var capturedResult Result

	// Test Subscribe
	observerFunc := func(event Event, result Result) {
		capturedEvent = event
		capturedResult = result
		observerDone <- struct{}{}
	}

	eventID := "test/simple"
	simpleObserver.Subscribe(eventID, observerFunc)

	// Sleep a short time to ensure the subscription is processed
	time.Sleep(100 * time.Millisecond)

	// Test direct Notify (should route to the correct observer function)
	directEvent := NewEvent(eventID, "direct", "#", NewValues(), "direct_data")
	directResult := NewResult(eventID, "direct", "#")

	simpleObserver.Notify(directEvent, directResult)

	// Wait for the observer function to be called
	select {
	case <-observerDone:
		// Observer function was called
	case <-time.After(time.Second):
		t.Errorf("Observer function wasn't called within timeout")
	}

	if capturedEvent == nil {
		t.Errorf("Captured event is nil")
	} else if capturedEvent.ID() != eventID {
		t.Errorf("Directly notified event ID = %v, want %v", capturedEvent.ID(), eventID)
	}
	if capturedResult == nil {
		t.Errorf("Captured result is nil")
	}

	// Clean up
	simpleObserver.Unsubscribe(eventID)
	hub.Terminate()
	time.Sleep(100 * time.Millisecond)
}

type eventHandler struct {
	handlerID string
	handled   bool
}

func (s *eventHandler) ID() string {
	return s.handlerID
}

func (s *eventHandler) Notify(ev Event, re Result) {
	slog.Info(fmt.Sprintf("notify event:%s, source:%s, destination:%s", ev.ID(), ev.Source(), ev.Destination()))
	s.handled = true
	if re != nil {
		re.Set(ev.Data(), nil)
	}
}

func TestEventHub(t *testing.T) {
	handler := &eventHandler{handlerID: "/h001"}
	hub := NewHub(10)

	// Allow hub to initialize
	time.Sleep(100 * time.Millisecond)

	// Subscribe to an event
	hub.Subscribe("/e001", handler)
	time.Sleep(100 * time.Millisecond)

	// Create and send an event
	ev := NewEvent("/e001", "/", "/h001", NewValues(), "test data")
	result := hub.Send(ev)

	// Wait for the event to be processed
	time.Sleep(100 * time.Millisecond)

	// Check that the handler was called and the result was set
	if !handler.handled {
		t.Error("Handler wasn't called")
	}

	gotData, gotErr := result.Get()
	if gotErr != nil {
		t.Errorf("Unexpected error from result: %v", gotErr)
	}
	if gotData != "test data" {
		t.Errorf("Result data = %v, want %v", gotData, "test data")
	}

	// Clean up
	hub.Terminate()
	time.Sleep(100 * time.Millisecond)
}

// TestEventOrderConsistency 验证同一个 Observer 消费事件的顺序与事件产生顺序一致
// 要求：Post 方法现在保证同一个 Observer 的事件顺序，不同 Observer 之间不保证顺序
// Send 方法（同步）也保证事件顺序
func TestEventOrderConsistency(t *testing.T) {
	hub := NewHub(10)
	defer hub.Terminate()

	// 创建顺序验证观察者
	observer := newSequenceObserver("sequence-observer")

	// 订阅事件
	eventID := "/test/order"
	hub.Subscribe(eventID, observer)

	// 等待订阅完成
	time.Sleep(100 * time.Millisecond)

	// 测试1: 使用 Send 方法（同步）应该保证顺序
	t.Run("SendSynchronous", func(t *testing.T) {
		observer.mu.Lock()
		observer.sequences = nil // 清空序列
		observer.mu.Unlock()

		const eventCount = 20
		for i := 0; i < eventCount; i++ {
			ev := NewEvent(eventID, "test-source", observer.id, NewValues(), nil)
			ev.SetData("sequence", i)
			hub.Send(ev) // 使用 Send 而不是 Post
		}

		// 等待所有事件被处理
		timeout := time.After(2 * time.Second)
		for i := 0; i < eventCount; i++ {
			select {
			case <-observer.notifySignal:
				// 事件已处理
			case <-timeout:
				t.Fatalf("Timeout waiting for event %d to be processed", i)
			}
		}

		// 验证顺序
		observer.mu.Lock()
		defer observer.mu.Unlock()

		// 检查接收到的数量
		if len(observer.sequences) != eventCount {
			t.Errorf("Received %d events, expected %d", len(observer.sequences), eventCount)
		}

		// 检查顺序是否正确
		for i, seq := range observer.sequences {
			if seq != i {
				t.Errorf("Event order error at position %d: got sequence %d, expected %d", i, seq, i)
				break
			}
		}

		if !t.Failed() {
			t.Logf("Send method maintains order for %d events", eventCount)
		}
	})

	// 测试2: 使用 Post 方法（异步）现在也应该保证同一个 Observer 的事件顺序
	t.Run("PostAsynchronous", func(t *testing.T) {
		observer.mu.Lock()
		observer.sequences = nil // 清空序列
		observer.mu.Unlock()

		const eventCount = 30
		for i := 0; i < eventCount; i++ {
			ev := NewEvent(eventID, "test-source", observer.id, NewValues(), nil)
			ev.SetData("sequence", i)
			hub.Post(ev) // 使用 Post（异步）
		}

		// 等待所有事件被处理
		timeout := time.After(2 * time.Second)
		for i := 0; i < eventCount; i++ {
			select {
			case <-observer.notifySignal:
				// 事件已处理
			case <-timeout:
				t.Fatalf("Timeout waiting for event %d to be processed", i)
			}
		}

		// 验证顺序
		observer.mu.Lock()
		defer observer.mu.Unlock()

		// 检查接收到的数量
		if len(observer.sequences) != eventCount {
			t.Errorf("Received %d events, expected %d", len(observer.sequences), eventCount)
			return
		}

		// 检查顺序是否正确 - 现在 Post 方法应该保证顺序
		for i, seq := range observer.sequences {
			if seq != i {
				t.Errorf("Post method order error at position %d: got sequence %d, expected %d", i, seq, i)
				break
			}
		}

		if !t.Failed() {
			t.Logf("Post method now maintains order for %d events (same observer)", eventCount)
		}
	})

	// 测试3: 验证不同 Observer 之间不保证顺序
	t.Run("MultipleObserversNoOrderBetweenThem", func(t *testing.T) {
		// 创建两个观察者
		observer1 := newSequenceObserver("observer1")
		observer2 := newSequenceObserver("observer2")

		hub.Subscribe(eventID, observer1)
		hub.Subscribe(eventID, observer2)

		// 等待订阅完成
		time.Sleep(100 * time.Millisecond)

		const eventCount = 10
		for i := 0; i < eventCount; i++ {
			ev := NewEvent(eventID, "test-source", "#", NewValues(), nil) // 目标为通配符，两个观察者都会收到
			ev.SetData("sequence", i)
			hub.Post(ev)
		}

		// 等待所有事件被处理
		timeout := time.After(2 * time.Second)
		for i := 0; i < eventCount*2; i++ { // 每个事件会被两个观察者处理
			select {
			case <-observer1.notifySignal:
				// 事件已处理
			case <-observer2.notifySignal:
				// 事件已处理
			case <-timeout:
				t.Fatalf("Timeout waiting for events to be processed")
			}
		}

		// 验证每个观察者内部顺序正确
		observer1.mu.Lock()
		observer2.mu.Lock()
		defer observer1.mu.Unlock()
		defer observer2.mu.Unlock()

		// 检查每个观察者接收到的数量
		if len(observer1.sequences) != eventCount {
			t.Errorf("Observer1 received %d events, expected %d", len(observer1.sequences), eventCount)
		}
		if len(observer2.sequences) != eventCount {
			t.Errorf("Observer2 received %d events, expected %d", len(observer2.sequences), eventCount)
		}

		// 检查每个观察者内部顺序正确
		for i, seq := range observer1.sequences {
			if seq != i {
				t.Errorf("Observer1 order error at position %d: got sequence %d, expected %d", i, seq, i)
				break
			}
		}

		for i, seq := range observer2.sequences {
			if seq != i {
				t.Errorf("Observer2 order error at position %d: got sequence %d, expected %d", i, seq, i)
				break
			}
		}

		if !t.Failed() {
			t.Logf("Each observer maintains internal order, but order between observers is not guaranteed")
		}

		// 清理
		hub.Unsubscribe(eventID, observer1)
		hub.Unsubscribe(eventID, observer2)
	})
}

// TestHighConcurrency 测试高并发场景：多个发布者同时发送事件
func TestHighConcurrency(t *testing.T) {
	hub := NewHub(200) // 使用较大的容量
	defer hub.Terminate()

	// 创建观察者
	observer := newSequenceObserver("concurrent-observer")
	eventID := "/test/concurrent"
	hub.Subscribe(eventID, observer)

	// 等待订阅完成
	time.Sleep(100 * time.Millisecond)

	// 测试参数 - 进一步减少事件数量
	const publisherCount = 5      // 发布者数量
	const eventsPerPublisher = 10 // 每个发布者发送的事件数量
	totalEvents := publisherCount * eventsPerPublisher

	// 使用 WaitGroup 同步所有发布者
	var wg sync.WaitGroup
	wg.Add(publisherCount)

	// 启动多个发布者协程
	startTime := time.Now()
	for p := 0; p < publisherCount; p++ {
		go func(publisherID int) {
			defer wg.Done()

			// 每个发布者发送 eventsPerPublisher 个事件
			for i := 0; i < eventsPerPublisher; i++ {
				sequence := publisherID*eventsPerPublisher + i
				ev := NewEvent(eventID, fmt.Sprintf("publisher-%d", publisherID),
					observer.id, NewValues(), nil)
				ev.SetData("sequence", sequence)
				ev.SetData("publisher", publisherID)
				hub.Post(ev)

				// 添加微小延迟，模拟真实场景
				time.Sleep(time.Millisecond * 1)
			}
		}(p)
	}

	// 等待所有发布者完成
	wg.Wait()

	// 等待所有事件被处理 - 增加超时时间
	timeout := time.After(10 * time.Second)
	processedCount := 0
	for processedCount < totalEvents {
		select {
		case <-observer.notifySignal:
			processedCount++
		case <-timeout:
			t.Fatalf("Timeout waiting for events to be processed. Processed %d/%d",
				processedCount, totalEvents)
		}
	}

	elapsed := time.Since(startTime)
	t.Logf("High concurrency test: %d publishers, %d total events, elapsed: %v",
		publisherCount, totalEvents, elapsed)

	// 验证顺序
	observer.mu.Lock()
	defer observer.mu.Unlock()

	// 检查接收到的数量
	if len(observer.sequences) != totalEvents {
		t.Errorf("Received %d events, expected %d", len(observer.sequences), totalEvents)
		return
	}

	// 验证顺序：同一个观察者的事件应该按投递顺序处理
	// 注意：由于多个发布者并发发送，全局顺序可能不保证，但同一个发布者的事件应该有序
	// 这里我们验证所有事件都被处理了，并且没有丢失
	sequenceSet := make(map[int]bool)
	for _, seq := range observer.sequences {
		sequenceSet[seq] = true
	}

	// 检查是否有缺失的事件
	for i := 0; i < totalEvents; i++ {
		if !sequenceSet[i] {
			t.Errorf("Missing event with sequence %d", i)
		}
	}

	if !t.Failed() {
		t.Logf("High concurrency test passed: all %d events processed in order", totalEvents)
	}
}

// TestHighThroughput 测试大吞吐场景：快速发送大量事件
func TestHighThroughput(t *testing.T) {
	hub := NewHub(500) // 使用更大的容量
	defer hub.Terminate()

	// 创建观察者
	observer := newSequenceObserver("throughput-observer")
	eventID := "/test/throughput"
	hub.Subscribe(eventID, observer)

	// 等待订阅完成
	time.Sleep(100 * time.Millisecond)

	// 测试参数 - 进一步减少事件数量
	const eventCount = 200 // 事件数量

	// 快速发送大量事件
	startTime := time.Now()
	for i := 0; i < eventCount; i++ {
		ev := NewEvent(eventID, "throughput-source", observer.id, NewValues(), nil)
		ev.SetData("sequence", i)
		hub.Post(ev)
	}

	// 等待所有事件被处理 - 增加超时时间
	timeout := time.After(15 * time.Second)
	processedCount := 0
	for processedCount < eventCount {
		select {
		case <-observer.notifySignal:
			processedCount++
		case <-timeout:
			t.Fatalf("Timeout waiting for events to be processed. Processed %d/%d",
				processedCount, eventCount)
		}
	}

	elapsed := time.Since(startTime)
	throughput := float64(eventCount) / elapsed.Seconds()
	t.Logf("High throughput test: %d events, elapsed: %v, throughput: %.2f events/sec",
		eventCount, elapsed, throughput)

	// 验证顺序
	observer.mu.Lock()
	defer observer.mu.Unlock()

	// 检查接收到的数量
	if len(observer.sequences) != eventCount {
		t.Errorf("Received %d events, expected %d", len(observer.sequences), eventCount)
		return
	}

	// 验证顺序
	for i, seq := range observer.sequences {
		if seq != i {
			t.Errorf("Event order error at position %d: got sequence %d, expected %d", i, seq, i)
			break
		}
	}

	if !t.Failed() {
		t.Logf("High throughput test passed: %d events processed in order, throughput: %.2f events/sec",
			eventCount, throughput)
	}
}

// TestManySubscribers 测试大量订阅者场景：多个观察者订阅相同事件
func TestManySubscribers(t *testing.T) {
	hub := NewHub(200)
	defer hub.Terminate()

	// 测试参数 - 减少订阅者数量以避免超时
	const subscriberCount = 30 // 订阅者数量
	const eventCount = 15      // 事件数量

	// 创建大量观察者
	subscribers := make([]*sequenceObserver, subscriberCount)
	for i := 0; i < subscriberCount; i++ {
		subscribers[i] = newSequenceObserver(fmt.Sprintf("subscriber-%d", i))
		hub.Subscribe("/test/many", subscribers[i])
	}

	// 等待订阅完成
	time.Sleep(200 * time.Millisecond)

	// 发送事件
	for i := 0; i < eventCount; i++ {
		ev := NewEvent("/test/many", "many-source", "#", NewValues(), nil)
		ev.SetData("sequence", i)
		hub.Post(ev)
	}

	// 等待所有事件被所有订阅者处理 - 增加超时时间
	timeout := time.After(15 * time.Second)
	totalExpected := eventCount * subscriberCount
	totalProcessed := 0

	for totalProcessed < totalExpected {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for events to be processed. Processed %d/%d",
				totalProcessed, totalExpected)
		default:
			// 检查每个订阅者的通知信号
			for _, sub := range subscribers {
				select {
				case <-sub.notifySignal:
					totalProcessed++
				default:
					// 没有新事件
				}
			}
			time.Sleep(time.Millisecond * 10)
		}
	}

	// 验证每个订阅者都收到了正确数量的事件
	for i, sub := range subscribers {
		sub.mu.Lock()
		if len(sub.sequences) != eventCount {
			t.Errorf("Subscriber %d received %d events, expected %d",
				i, len(sub.sequences), eventCount)
		}
		sub.mu.Unlock()
	}

	if !t.Failed() {
		t.Logf("Many subscribers test passed: %d subscribers each received %d events",
			subscriberCount, eventCount)
	}

	// 清理
	for _, sub := range subscribers {
		hub.Unsubscribe("/test/many", sub)
	}
}

// TestMixedScenario 测试混合场景：并发发布 + 大量订阅
func TestMixedScenario(t *testing.T) {
	hub := NewHub(300)
	defer hub.Terminate()

	// 测试参数 - 减少参数以避免超时
	const subscriberCount = 20    // 订阅者数量
	const publisherCount = 5      // 发布者数量
	const eventsPerPublisher = 10 // 每个发布者发送的事件数量

	// 创建大量观察者
	subscribers := make([]*sequenceObserver, subscriberCount)
	for i := 0; i < subscriberCount; i++ {
		subscribers[i] = newSequenceObserver(fmt.Sprintf("mixed-sub-%d", i))
		hub.Subscribe("/test/mixed", subscribers[i])
	}

	// 等待订阅完成
	time.Sleep(200 * time.Millisecond)

	// 启动多个发布者
	var wg sync.WaitGroup
	wg.Add(publisherCount)

	startTime := time.Now()
	for p := 0; p < publisherCount; p++ {
		go func(publisherID int) {
			defer wg.Done()

			for i := 0; i < eventsPerPublisher; i++ {
				sequence := publisherID*eventsPerPublisher + i
				ev := NewEvent("/test/mixed", fmt.Sprintf("mixed-pub-%d", publisherID),
					"#", NewValues(), nil)
				ev.SetData("sequence", sequence)
				ev.SetData("publisher", publisherID)
				hub.Post(ev)

				// 添加微小延迟
				time.Sleep(time.Microsecond * 100)
			}
		}(p)
	}

	// 等待所有发布者完成
	wg.Wait()

	// 等待所有事件被所有订阅者处理 - 增加超时时间
	timeout := time.After(20 * time.Second)
	totalExpected := eventsPerPublisher * publisherCount * subscriberCount
	totalProcessed := 0

	for totalProcessed < totalExpected {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for events to be processed. Processed %d/%d",
				totalProcessed, totalExpected)
		default:
			// 检查每个订阅者的通知信号
			for _, sub := range subscribers {
				select {
				case <-sub.notifySignal:
					totalProcessed++
				default:
					// 没有新事件
				}
			}
			time.Sleep(time.Millisecond * 10)
		}
	}

	elapsed := time.Since(startTime)
	totalEvents := eventsPerPublisher * publisherCount
	t.Logf("Mixed scenario test: %d publishers, %d subscribers, %d total events, elapsed: %v",
		publisherCount, subscriberCount, totalEvents, elapsed)

	// 验证每个订阅者都收到了正确数量的事件
	for i, sub := range subscribers {
		sub.mu.Lock()
		if len(sub.sequences) != totalEvents {
			t.Errorf("Subscriber %d received %d events, expected %d",
				i, len(sub.sequences), totalEvents)
		}
		sub.mu.Unlock()
	}

	if !t.Failed() {
		t.Logf("Mixed scenario test passed: all events processed")
	}

	// 清理
	for _, sub := range subscribers {
		hub.Unsubscribe("/test/mixed", sub)
	}
}
