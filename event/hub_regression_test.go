package event

import (
	"testing"
	"time"
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

func TestHubCacheRespectsDestination(t *testing.T) {
	hub := NewHubWithOptions(4)
	defer hub.Terminate()

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
	defer hub.Terminate()

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
			hub.Terminate()
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
		hub.Terminate()
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
