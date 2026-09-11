package event

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
)

func TestQueuedSendCancellationDoesNotInvokeHandler(t *testing.T) {
	hub := NewHub(8).(*hubImpl)
	observer := NewSimpleObserver("lane", hub)
	entered, release := make(chan struct{}), make(chan struct{})
	observer.Subscribe("hold", func(_ Event, result Result) { close(entered); <-release; result.Set(nil, nil) })
	var called atomic.Bool
	observer.Subscribe("write", func(_ Event, result Result) { called.Store(true); result.Set(nil, nil) })
	held := make(chan Result, 1)
	go func() { held <- hub.Send(NewEvent("hold", "caller", "lane", nil, nil)) }()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	done := make(chan Result, 1)
	go func() { done <- hub.Send(NewEventWithContext("write", "caller", "lane", nil, ctx, nil)) }()
	select {
	case result := <-done:
		if result == nil || result.Error() == nil {
			t.Error("canceled send reported success")
		}
	case <-time.After(time.Second):
		t.Error("queued send ignored its deadline")
	}
	close(release)
	<-held
	if err := hub.TerminateChecked(context.Background()); err != nil {
		t.Fatal(err)
	}
	if called.Load() {
		t.Fatal("canceled queued write executed later")
	}
}

func TestExecutingSendRetainsCallerUntilCallbackReturns(t *testing.T) {
	hub := NewHub(8)
	defer hub.Terminate(context.Background())
	observer := NewSimpleObserver("lane", hub)
	entered, release := make(chan struct{}), make(chan struct{})
	observer.Subscribe("hold", func(_ Event, result Result) { close(entered); <-release; result.Set("completed", nil) })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan Result, 1)
	go func() { done <- hub.Send(NewEventWithContext("hold", "caller", "lane", nil, ctx, nil)) }()
	<-entered
	cancel()
	select {
	case <-done:
		t.Error("running callback detached from caller on cancellation")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	result := <-done
	if value, err := result.Get(); err != nil || value != "completed" {
		t.Fatal(value, err)
	}
}

func TestIndependentRootLaneCyclesFailWithoutDeadline(t *testing.T) {
	for _, lanes := range [][]string{{"a", "b"}, {"a", "b", "c"}} {
		t.Run(lanes[len(lanes)-1], func(t *testing.T) {
			hub := NewHub(8)
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				hub.Terminate(ctx)
			})
			entered, barrier := make(chan struct{}, len(lanes)), make(chan struct{})
			for idx, lane := range lanes {
				target := lanes[(idx+1)%len(lanes)]
				observer := NewSimpleObserver(lane, hub)
				observer.Subscribe("leaf", func(_ Event, result Result) { result.Set(true, nil) })
				observer.Subscribe("outer", func(ev Event, result Result) {
					entered <- struct{}{}
					<-barrier
					value, err := hub.Send(NewEventWithContext("leaf", lane, target, nil, ev.Context(), nil)).Get()
					result.Set(value, err)
				})
			}
			done := make(chan Result, len(lanes))
			for _, lane := range lanes {
				go func() { done <- hub.Send(NewEvent("outer", "caller", lane, nil, nil)) }()
			}
			for range lanes {
				<-entered
			}
			close(barrier)
			rejected := 0
			for range lanes {
				select {
				case result := <-done:
					if err := result.Error(); err != nil {
						if err.Code != cd.InvalidOperation {
							t.Fatal(err)
						}
						rejected++
					}
				case <-time.After(time.Second):
					t.Fatal("independent root cycle deadlocked")
				}
			}
			if rejected == 0 {
				t.Fatal("cyclic lane dependency was not rejected")
			}
			impl := hub.(*hubImpl)
			impl.waitsMu.Lock()
			defer impl.waitsMu.Unlock()
			if len(impl.waits) != 0 {
				t.Fatal("completed calls retained wait graph entries")
			}
		})
	}
}

func TestCheckedHubShutdownRetainsNestedDependenciesAndRetries(t *testing.T) {
	hub := NewHub(8).(*hubImpl)
	a, b := NewSimpleObserver("a", hub), NewSimpleObserver("b", hub)
	entered, release := make(chan struct{}), make(chan struct{})
	b.Subscribe("leaf", func(_ Event, result Result) { result.Set(true, nil) })
	a.Subscribe("outer", func(ev Event, result Result) {
		close(entered)
		<-release
		r := hub.Send(NewEventWithContext("leaf", "a", "b", nil, ev.Context(), nil))
		if r == nil {
			result.Set(nil, cd.NewError(cd.Unexpected, "live handler lost its dependency"))
			return
		}
		value, err := r.Get()
		result.Set(value, err)
	})
	done := make(chan Result, 1)
	go func() { done <- hub.Send(NewEvent("outer", "caller", "a", nil, nil)) }()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := hub.TerminateChecked(ctx); err == nil {
		t.Error("live dispatch was reported as terminated")
	}
	if hub.Send(NewEvent("leaf", "new-root", "b", nil, nil)) != nil {
		t.Error("terminating hub accepted new root")
	}
	close(release)
	if result := <-done; result.Error() != nil {
		t.Fatal(result.Error())
	}
	if err := hub.TerminateChecked(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := hub.TerminateChecked(context.Background()); err != nil {
		t.Fatal("repeated shutdown", err)
	}
}

func TestPostQueuesBehindActiveAncestorWithoutInheritingItsAuthority(t *testing.T) {
	hub := NewHub(8)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		hub.Terminate(ctx)
	})
	a, b := NewSimpleObserver("a", hub), NewSimpleObserver("b", hub)
	var returned atomic.Bool
	done := make(chan *cd.Error, 1)
	b.Subscribe("leaf", func(_ Event, result Result) { result.Set(true, nil) })
	a.Subscribe("posted", func(ev Event, _ Result) {
		if !returned.Load() {
			done <- cd.NewError(cd.Unexpected, "Post ran inline before its ancestor returned")
			return
		}
		result := hub.Send(NewEventWithContext("leaf", "a", "b", nil, ev.Context(), nil))
		done <- result.Error()
	})
	b.Subscribe("middle", func(ev Event, result Result) {
		hub.Post(NewEventWithContext("posted", "b", "a", nil, ev.Context(), nil))
		result.Set(nil, nil)
	})
	a.Subscribe("outer", func(ev Event, result Result) {
		value, err := hub.Send(NewEventWithContext("middle", "a", "b", nil, ev.Context(), nil)).Get()
		returned.Store(true)
		result.Set(value, err)
	})
	root := make(chan Result, 1)
	go func() { root <- hub.Send(NewEvent("outer", "caller", "a", nil, nil)) }()
	select {
	case result := <-root:
		if result.Error() != nil {
			t.Fatal(result.Error())
		}
	case <-time.After(time.Second):
		t.Fatal("Post caused ancestor deadlock")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("posted callback failed to drain")
	}
}
