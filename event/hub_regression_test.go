package event

import (
	"context"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
)

type blockingObserver struct {
	id        string
	started   chan struct{}
	releaseCh chan struct{}
}

func (b *blockingObserver) ID() string {
	return b.id
}

func (b *blockingObserver) Notify(event Event, result Result) {
	select {
	case b.started <- struct{}{}:
	default:
	}
	<-b.releaseCh
}

type blockingIDObserver struct {
	releaseCh chan struct{}
}

func (b *blockingIDObserver) ID() string {
	<-b.releaseCh
	return "/dest/block-id"
}

func (b *blockingIDObserver) Notify(event Event, result Result) {}

type reentrantSendObserver struct {
	id      string
	eventID string
	hub     Hub
	errCh   chan error
}

type laneContextProbeObserver struct {
	id        string
	started   chan context.Context
	releaseCh chan struct{}
}

func (s *reentrantSendObserver) ID() string {
	return s.id
}

func (s *laneContextProbeObserver) ID() string {
	return s.id
}

func (s *laneContextProbeObserver) Notify(event Event, result Result) {
	select {
	case s.started <- event.Context():
	default:
	}
	<-s.releaseCh
	if result != nil {
		result.Set(nil, nil)
	}
}

func (s *reentrantSendObserver) Notify(event Event, result Result) {
	depth, _ := event.Data().(int)
	if depth > 0 {
		nestedEvent := NewEvent(s.eventID, s.id, s.id, NewValues(), depth-1)
		nestedEvent.BindContext(event.Context())
		nestedResult := s.hub.Send(nestedEvent)
		if nestedResult == nil {
			s.errCh <- cd.NewError(cd.Unexpected, "nested result is nil")
			return
		}

		nestedValue, nestedErr := nestedResult.Get()
		if nestedErr != nil {
			s.errCh <- nestedErr
			return
		}
		if nestedValue != depth-1 {
			s.errCh <- cd.NewError(cd.Unexpected, "nested value mismatch")
			return
		}
	}

	if result != nil {
		result.Set(depth, nil)
	}
	select {
	case s.errCh <- nil:
	default:
	}
}

func TestHubLaneContextDoesNotMutateOriginalEvent(t *testing.T) {
	hub := NewHubWithOptions(1, WithPerLaneChanSize(1))
	defer hub.Terminate(context.Background())

	eventID := "/lane/context"
	observer := &laneContextProbeObserver{
		id:        "/dest/lane-context",
		started:   make(chan context.Context, 1),
		releaseCh: make(chan struct{}),
	}
	hub.Subscribe(eventID, observer)
	time.Sleep(20 * time.Millisecond)

	ev := NewEvent(eventID, "source", observer.id, NewValues(), nil)
	hub.Post(ev)

	var observerCtx context.Context
	select {
	case observerCtx = <-observer.started:
	case <-time.After(time.Second):
		t.Fatal("observer did not start")
	}

	if got, ok := observerCtx.Value(laneExecutionContextKey{}).(string); !ok || got != observer.id {
		t.Fatalf("observer context missing lane key, got=%v ok=%v", got, ok)
	}
	if got := ev.Context().Value(laneExecutionContextKey{}); got != nil {
		t.Fatalf("hub mutated original event context, got lane key %v", got)
	}

	close(observer.releaseCh)
}

func TestHubSendNilEventReturnsError(t *testing.T) {
	hub := NewHubWithOptions(1)
	defer hub.Terminate(context.Background())

	result := hub.Send(nil)
	if result == nil || result.Error() == nil {
		t.Fatal("expected nil event error")
	}
	if result.Error().Code != cd.IllegalParam {
		t.Fatalf("error code=%d want=%d", result.Error().Code, cd.IllegalParam)
	}
}

func TestHubCacheRespectsDestination(t *testing.T) {
	hub := NewHubWithOptions(4)
	defer hub.Terminate(context.Background())

	eventID := "/cache/destination"
	observerA := newTestObserver("/dest/a")
	observerB := newTestObserver("/dest/b")

	hub.Subscribe(eventID, observerA)
	hub.Subscribe(eventID, observerB)

	time.Sleep(50 * time.Millisecond)

	hub.Post(NewEvent(eventID, "source", "/dest/a", NewValues(), nil))
	if !waitForNotification(observerA, time.Second) {
		t.Fatal("timeout waiting for observerA")
	}

	hub.Post(NewEvent(eventID, "source", "/dest/b", NewValues(), nil))
	if !waitForNotification(observerB, time.Second) {
		t.Fatal("timeout waiting for observerB after cache reuse")
	}
}

func TestHubSendTimeoutDoesNotBlock(t *testing.T) {
	hub := NewHubWithOptions(1, WithPerDestinationChanSize(1))
	defer hub.Terminate(context.Background())

	eventID := "/send/timeout"
	observer := &blockingObserver{
		id:        "/dest/blocking",
		started:   make(chan struct{}, 1),
		releaseCh: make(chan struct{}),
	}
	hub.Subscribe(eventID, observer)
	time.Sleep(20 * time.Millisecond)

	hub.Post(NewEvent(eventID, "source", observer.id, NewValues(), nil))
	select {
	case <-observer.started:
	case <-time.After(time.Second):
		t.Fatal("observer did not start")
	}

	// Fill the per-destination channel while the worker goroutine is blocked in Notify.
	hub.Post(NewEvent(eventID, "source-2", observer.id, NewValues(), nil))

	done := make(chan Result, 1)
	go func() {
		done <- hub.Send(NewEvent(eventID, "source-3", observer.id, NewValues(), nil))
	}()

	select {
	case result := <-done:
		if result == nil || result.Error() == nil {
			t.Fatal("expected timeout result")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Send blocked after channel timeout")
	}

	close(observer.releaseCh)
}

func TestHubTerminateIsConcurrentSafe(t *testing.T) {
	hub := NewHubWithOptions(4)

	done := make(chan struct{}, 2)
	for i := 0; i < 2; i++ {
		go func() {
			hub.Terminate(context.Background())
			done <- struct{}{}
		}()
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("first terminate call did not complete")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("second terminate call did not complete")
	}
}

func TestHubTerminateDoesNotBlockWhenHubActionChannelIsBusy(t *testing.T) {
	hub := NewHubWithOptions(1, WithHubActionChanSize(1))
	hubPtr := hub.(*hubImpl)

	blocker := &blockingIDObserver{releaseCh: make(chan struct{})}
	firstResult := make(chan bool, 1)
	hubPtr.hubActionChannel <- &subscribeData{eventID: "/busy", observer: blocker, result: firstResult}
	hubPtr.hubActionChannel <- &postData{event: NewEvent("/queued", "src", "dst", NewValues(), nil)}

	done := make(chan struct{})
	go func() {
		hub.Terminate(context.Background())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Terminate blocked while hubActionChannel send timed out")
	}

	close(blocker.releaseCh)
	select {
	case <-firstResult:
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHubSendDoesNotDeadlockOnReentrantSameLane(t *testing.T) {
	hub := NewHubWithOptions(1, WithPerLaneChanSize(1))
	defer hub.Terminate(context.Background())

	eventID := "/send/reentrant"
	observer := &reentrantSendObserver{
		id:      "/dest/reentrant",
		eventID: eventID,
		hub:     hub,
		errCh:   make(chan error, 4),
	}
	hub.Subscribe(eventID, observer)
	time.Sleep(20 * time.Millisecond)

	done := make(chan Result, 1)
	go func() {
		done <- hub.Send(NewEvent(eventID, "external", observer.id, NewValues(), 1))
	}()

	select {
	case result := <-done:
		if result == nil || result.Error() != nil {
			t.Fatalf("expected successful result, got error %v", result.Error())
		}
		value, err := result.Get()
		if err != nil || value != 1 {
			t.Fatalf("unexpected result, value=%v err=%v", value, err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("reentrant send blocked on same lane")
	}

	select {
	case err := <-observer.errCh:
		if err != nil {
			t.Fatalf("observer reported error: %v", err)
		}
	default:
	}
}
