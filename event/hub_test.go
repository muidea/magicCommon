package event

import (
	"context"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
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
	t.notifyCount++
	t.lastEvent = event
	t.lastResult = result
	select {
	case t.notifySignal <- struct{}{}:
	default:
	}
}

func TestHubImpl(t *testing.T) {
	hub := NewHub(10)

	// Test Subscribe/Post/Unsubscribe
	observer1 := newTestObserver("observer1")
	observer2 := newTestObserver("observer2")

	eventID := "test/event"
	hub.Subscribe(eventID, observer1)
	hub.Subscribe(eventID, observer2)

	// Sleep a short time to ensure the subscription is processed
	time.Sleep(100 * time.Millisecond)

	// Test Post
	event := NewEvent(eventID, "source", "#", NewValues(), "data")
	hub.Post(event)

	// Allow time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	if observer1.notifyCount == 0 {
		t.Errorf("Observer1 notify count = %d, want > 0", observer1.notifyCount)
	}
	if observer2.notifyCount == 0 {
		t.Errorf("Observer2 notify count = %d, want > 0", observer2.notifyCount)
	}

	// Test Unsubscribe
	hub.Unsubscribe(eventID, observer1)

	// Sleep a short time to ensure the unsubscription is processed
	time.Sleep(100 * time.Millisecond)

	// Reset notification count
	observer1.notifyCount = 0
	observer2.notifyCount = 0

	// Post again
	hub.Post(event)

	// Allow time for the event to be processed
	time.Sleep(100 * time.Millisecond)

	// Check that only observer2 was notified
	if observer1.notifyCount != 0 {
		t.Errorf("Observer1 notify count after unsubscribe = %d, want 0", observer1.notifyCount)
	}
	if observer2.notifyCount == 0 {
		t.Errorf("Observer2 notify count = %d, want > 0", observer2.notifyCount)
	}

	// Clean up
	hub.Terminate()
	time.Sleep(100 * time.Millisecond)
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
	log.Infof("notify event:%s, source:%s, destination:%s", ev.ID(), ev.Source(), ev.Destination())
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
