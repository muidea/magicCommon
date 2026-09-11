package event

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	cd "github.com/muidea/magicCommon/def"
)

type controlGateObserver struct {
	entered chan struct{}
	release chan struct{}
}

func (s *controlGateObserver) ID() string {
	close(s.entered)
	<-s.release
	return "" // Reject after the gate; never publish this observer.
}
func (*controlGateObserver) Notify(Event, Result) {}

// Occupy the worker and fill its queue without relying on scheduler timing.
func congestSubscriptions(t *testing.T, hub *hubImpl) func() {
	t.Helper()
	gate := &controlGateObserver{entered: make(chan struct{}), release: make(chan struct{})}
	done := make(chan *cd.Error, 1)
	go func() { done <- hub.Subscribe("gate", gate) }()
	<-gate.entered
	filler := &subscribeData{eventID: "filler", observer: NewSimpleObserver("filler", hub), result: make(chan *cd.Error, 1)}
	hub.hubActionChannel <- filler
	var once sync.Once
	release := func() {
		once.Do(func() {
			close(gate.release)
			<-done
			if err := <-filler.result; err != nil {
				t.Error(err)
			}
		})
	}
	t.Cleanup(release)
	return release
}

func TestSubscriptionRejectionPreservesLocalStateAndAllowsRetry(t *testing.T) {
	for _, remove := range []bool{false, true} {
		t.Run(map[bool]string{false: "subscribe", true: "unsubscribe"}[remove], func(t *testing.T) {
			hub := NewHubWithOptions(1, WithHubActionChanSize(1)).(*hubImpl)
			t.Cleanup(func() { hub.Terminate(context.Background()) })
			observer := NewSimpleObserver("owner", hub).(*simpleObserver)
			handler := func(_ Event, result Result) { result.Set("ready", nil) }
			if remove {
				if err := observer.Subscribe("command", handler); err != nil {
					t.Fatal(err)
				}
			}
			release := congestSubscriptions(t, hub)
			var err *cd.Error
			if remove {
				err = observer.Unsubscribe("command")
			} else {
				err = observer.Subscribe("command", handler)
			}
			if err == nil || err.Code != cd.ResourceExhausted {
				t.Fatalf("expected admission failure, got %v", err)
			}
			observer.eventIDLock.RLock()
			_, local := observer.eventID2ObserverFunc["command"]
			observer.eventIDLock.RUnlock()
			if local != remove {
				t.Fatal("rejected change mutated local callbacks")
			}
			release()
			result := hub.Send(NewEvent("command", "caller", "owner", nil, nil))
			if (result.Error() == nil) != remove {
				t.Fatal("rejected change mutated Hub subscriptions")
			}
			if remove {
				if err := observer.Unsubscribe("command"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := observer.Subscribe("command", handler); err != nil {
					t.Fatal(err)
				}
			}
			result = hub.Send(NewEvent("command", "caller", "owner", nil, nil))
			if (result.Error() == nil) == remove {
				t.Fatal("retry did not update subscription/cache")
			}
		})
	}
}

func TestSubscriptionAcknowledgementIsNotDropped(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		hub := NewHub(1).(*hubImpl)
		defer hub.Terminate(context.Background())
		// Force a delayed acknowledgement consumer; the control worker must not
		// drop the result after 10ms and leave the real waiter blocked forever.
		data := &subscribeData{eventID: "command", observer: NewSimpleObserver("owner", hub), result: make(chan *cd.Error)}
		done := make(chan struct{})
		go func() { hub.handleAction(data); close(done) }()
		time.Sleep(100 * time.Millisecond)
		select {
		case <-done:
			t.Fatal("acknowledgement was abandoned")
		default:
		}
		if err := <-data.result; err != nil {
			t.Fatal(err)
		}
		<-done
	})
}

func TestAdmittedSubscriptionWaitsForCompletion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		hub := NewHubWithOptions(1, WithHubActionChanSize(1)).(*hubImpl)
		defer hub.Terminate(context.Background())
		gate := &controlGateObserver{entered: make(chan struct{}), release: make(chan struct{})}
		gateDone := make(chan *cd.Error, 1)
		go func() { gateDone <- hub.Subscribe("gate", gate) }()
		<-gate.entered
		done := make(chan *cd.Error, 1)
		go func() { done <- hub.Subscribe("command", NewSimpleObserver("owner", hub)) }()
		synctest.Wait()
		time.Sleep(100 * time.Millisecond)
		select {
		case err := <-done:
			t.Fatalf("admitted operation returned early: %v", err)
		default:
		}
		close(gate.release)
		<-gateDone
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	})
}

type panicSubscriptionObserver struct{}

func (*panicSubscriptionObserver) ID() string           { panic("broken ID") }
func (*panicSubscriptionObserver) Notify(Event, Result) {}

func TestSubscriptionValidationAndControlWorkerRecovery(t *testing.T) {
	hub := NewHub(1)
	defer hub.Terminate(context.Background())
	for _, observer := range []Observer{nil, &panicSubscriptionObserver{}, NewSimpleObserver("", hub)} {
		if err := hub.Subscribe("command", observer); err == nil {
			t.Fatal("invalid observer accepted")
		}
	}
	observer := NewSimpleObserver("owner", hub)
	if err := observer.Subscribe("command", nil); err == nil {
		t.Fatal("nil handler accepted")
	}
	if err := observer.Subscribe("command", func(_ Event, result Result) { result.Set(true, nil) }); err != nil {
		t.Fatal(err)
	}
	if err := hub.Send(NewEvent("command", "caller", "owner", nil, nil)).Error(); err != nil {
		t.Fatal(err)
	}
}

func TestClosedHubRejectsSubscriptions(t *testing.T) {
	hub := NewHub(1)
	observer := NewSimpleObserver("owner", hub)
	if err := observer.Subscribe("existing", func(Event, Result) {}); err != nil {
		t.Fatal(err)
	}
	hub.Terminate(context.Background())
	for _, err := range []*cd.Error{
		hub.Subscribe("new", observer), hub.Unsubscribe("existing", observer),
		observer.Subscribe("new", func(Event, Result) {}), observer.Unsubscribe("existing"),
	} {
		if err == nil || err.Code != cd.InvalidOperation {
			t.Fatalf("closed Hub returned %v", err)
		}
	}
}

func TestSubscriptionDoesNotWaitForExecutionPoolCapacity(t *testing.T) {
	hub := NewHubWithOptions(1, WithWorkerPoolSize(1)).(*hubImpl)
	t.Cleanup(func() { hub.Terminate(context.Background()) })
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	entered := make(chan struct{})
	hub.Run(func() { close(entered); <-release })
	<-entered
	done := make(chan *cd.Error, 1)
	go func() { done <- hub.Subscribe("command", NewSimpleObserver("owner", hub)) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("subscription admission was blocked by execution capacity")
	}
}

func TestConcurrentSimpleObserverChangesStayConsistent(t *testing.T) {
	hub := NewHub(2)
	defer hub.Terminate(context.Background())
	observer := NewSimpleObserver("owner", hub)
	handler := func(_ Event, result Result) { result.Set("ready", nil) }
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for range 20 {
				if err := observer.Subscribe("command", handler); err != nil && err.Code != cd.Duplicated {
					t.Error(err)
				}
				if err := observer.Unsubscribe("command"); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	if err := observer.Unsubscribe("command"); err != nil {
		t.Fatal(err)
	}
	if hub.Send(NewEvent("command", "caller", "owner", nil, nil)).Error() == nil {
		t.Fatal("stale subscription")
	}
	if err := observer.Subscribe("command", handler); err != nil {
		t.Fatal(err)
	}
	if err := hub.Send(NewEvent("command", "caller", "owner", nil, nil)).Error(); err != nil {
		t.Fatal(err)
	}
}

type cacheGateObserver struct {
	entered       chan struct{}
	release       chan struct{}
	removing      atomic.Bool
	removeEntered chan struct{}
	removeOnce    sync.Once
}

func (s *cacheGateObserver) ID() string {
	if s.removing.Load() {
		s.removeOnce.Do(func() { close(s.removeEntered) })
	}
	return "owner"
}
func (s *cacheGateObserver) MatchID() string {
	close(s.entered)
	<-s.release
	return "owner"
}
func (*cacheGateObserver) Notify(Event, Result) {}

func TestUnsubscribeCannotBeUndoneByConcurrentCachePublication(t *testing.T) {
	hub := NewHub(1).(*hubImpl)
	defer hub.Terminate(context.Background())
	observer := &cacheGateObserver{entered: make(chan struct{}), release: make(chan struct{}), removeEntered: make(chan struct{})}
	if err := hub.Subscribe("command", observer); err != nil {
		t.Fatal(err)
	}
	lookupDone := make(chan struct{})
	go func() { hub.matchingObservers(NewEvent("command", "caller", "owner", nil, nil)); close(lookupDone) }()
	<-observer.entered
	observer.removing.Store(true)
	removeDone := make(chan *cd.Error, 1)
	go func() { removeDone <- hub.Unsubscribe("command", observer) }()
	<-observer.removeEntered
	close(observer.release)
	<-lookupDone
	if err := <-removeDone; err != nil {
		t.Fatal(err)
	}
	if hub.Send(NewEvent("command", "caller", "owner", nil, nil)).Error() == nil {
		t.Fatal("old lookup republished an unsubscribed observer")
	}
}
