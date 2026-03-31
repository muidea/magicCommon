package event

import (
	"testing"
	"time"
)

type laneAwareObserver struct {
	id        string
	started   chan string
	releaseCh map[string]chan struct{}
}

func (s *laneAwareObserver) ID() string {
	return s.id
}

func (s *laneAwareObserver) Notify(ev Event, re Result) {
	laneKey := ev.LaneKey()
	if s.started != nil {
		select {
		case s.started <- laneKey:
		default:
		}
	}

	if ch, ok := s.releaseCh[laneKey]; ok {
		<-ch
	}

	if re != nil {
		re.Set(laneKey, nil)
	}
}

func TestHubSendUsesLaneKeyForScheduling(t *testing.T) {
	hub := NewHubWithOptions(1, WithPerLaneChanSize(1))
	defer hub.Terminate()

	observer := &laneAwareObserver{
		id:      "/lane/test",
		started: make(chan string, 2),
		releaseCh: map[string]chan struct{}{
			"lane/a": make(chan struct{}),
		},
	}
	hub.Subscribe("/lane/event", observer)
	time.Sleep(20 * time.Millisecond)

	blockedEvent := NewEvent("/lane/event", "source-a", observer.ID(), NewValues(), nil)
	blockedEvent.BindLaneKey("lane/a")
	hub.Post(blockedEvent)

	select {
	case lane := <-observer.started:
		if lane != "lane/a" {
			t.Fatalf("observer started on lane %s, want lane/a", lane)
		}
	case <-time.After(time.Second):
		t.Fatal("observer did not start blocked lane")
	}

	done := make(chan Result, 1)
	go func() {
		fastEvent := NewEvent("/lane/event", "source-b", observer.ID(), NewValues(), nil)
		fastEvent.BindLaneKey("lane/b")
		done <- hub.Send(fastEvent)
	}()

	select {
	case result := <-done:
		if result == nil || result.Error() != nil {
			t.Fatalf("expected fast lane result, got error %v", result.Error())
		}
		val, err := result.Get()
		if err != nil || val != "lane/b" {
			t.Fatalf("unexpected result, val=%v err=%v", val, err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Send on different lane blocked behind same destination")
	}

	close(observer.releaseCh["lane/a"])
}

func TestSimpleObserverWithMatchID(t *testing.T) {
	hub := NewHubWithOptions(2)
	defer hub.Terminate()

	done := make(chan Event, 1)
	observer := NewSimpleObserverWithMatchID("base-observer", "/internal/modules/kernel/base/#", hub)
	observer.Subscribe("/value/query", func(ev Event, re Result) {
		done <- ev
		if re != nil {
			re.Set("ok", nil)
		}
	})

	ev := NewEvent("/value/query", "source", "/internal/modules/kernel/base/read/app/entity", NewValues(), nil)
	result := hub.Send(ev)
	if result == nil || result.Error() != nil {
		t.Fatalf("expected matched result, got error %v", result.Error())
	}

	select {
	case got := <-done:
		if got.Destination() != ev.Destination() {
			t.Fatalf("destination mismatch, got %s want %s", got.Destination(), ev.Destination())
		}
	case <-time.After(time.Second):
		t.Fatal("observer with custom match ID did not receive event")
	}
}
