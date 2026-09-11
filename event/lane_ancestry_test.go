package event

import (
	"context"
	"testing"
	"time"
)

func TestSynchronousSendCanRevisitActiveAncestorLane(t *testing.T) {
	hub := NewHub(8)
	t.Cleanup(func() { hub.Terminate(context.Background()) })
	a, b := NewSimpleObserver("a", hub), NewSimpleObserver("b", hub)
	a.Subscribe("read", func(ev Event, result Result) { result.Set("ok", nil) })
	b.Subscribe("middle", func(ev Event, result Result) {
		r := hub.Send(NewEventWithContext("read", "b", "a", nil, ev.Context(), nil))
		v, err := r.Get()
		result.Set(v, err)
	})
	a.Subscribe("outer", func(ev Event, result Result) {
		r := hub.Send(NewEventWithContext("middle", "a", "b", nil, ev.Context(), nil))
		v, err := r.Get()
		result.Set(v, err)
	})
	done := make(chan Result, 1)
	go func() { done <- hub.Send(NewEvent("outer", "caller", "a", nil, nil)) }()
	select {
	case result := <-done:
		v, err := result.Get()
		if err != nil || v != "ok" {
			t.Fatal(v, err)
		}
	case <-time.After(time.Second):
		t.Fatal("ancestor lane cycle deadlocked")
	}
}

func TestLaneContextExpiresAndIsBoundToItsHub(t *testing.T) {
	hub := NewHub(8).(*hubImpl)
	other := NewHub(8).(*hubImpl)
	t.Cleanup(func() { hub.Terminate(context.Background()); other.Terminate(context.Background()) })
	ev, finish := hub.eventWithLaneContext(NewEvent("outer", "caller", "a", nil, nil), true)
	nested := NewEventWithContext("inner", "a", "a", nil, ev.Context(), nil)
	if !hub.isReentrantLaneExecution(nested, "a") || other.isReentrantLaneExecution(nested, "a") {
		t.Fatal("incorrect hub scope")
	}
	finish()
	if hub.isReentrantLaneExecution(nested, "a") {
		t.Fatal("escaped context retained lane authority")
	}
	posted, done := hub.eventWithLaneContext(nested, false)
	defer done()
	if frame := posted.Context().Value(laneExecutionContextKey{}).(*laneExecutionFrame); frame.parent != nil {
		t.Fatal("Post inherited ancestry")
	}
}
