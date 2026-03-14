package session

import (
	"net/http/httptest"
	"testing"
	"time"
)

type testSessionObserver struct {
	id       string
	statusCh chan Status
}

func (t *testSessionObserver) ID() string {
	return t.id
}

func (t *testSessionObserver) OnStatusChange(session Session, status Status) {
	select {
	case t.statusCh <- status:
	default:
	}
}

func TestSessionResetClearsOptionsAndObservers(t *testing.T) {
	registry := NewRegistry(nil).(*sessionRegistryImpl)
	defer registry.Release()

	sessionPtr := &sessionImpl{
		id: "session-reset",
		context: map[string]any{
			InnerStartTime:        int64(100),
			InnerRemoteAccessAddr: "127.0.0.1",
			InnerUseAgent:         "agent",
			"custom":              "value",
		},
		observer: map[string]Observer{},
		registry: registry,
		status:   sessionActive,
	}

	observer := &testSessionObserver{id: "observer-1", statusCh: make(chan Status, 1)}
	sessionPtr.BindObserver(observer)
	sessionPtr.Reset()

	if _, ok := sessionPtr.GetOption("custom"); ok {
		t.Fatal("custom option should be cleared after reset")
	}
	if len(sessionPtr.observer) != 0 {
		t.Fatal("observers should be cleared after reset")
	}
	if _, ok := sessionPtr.GetOption(InnerRemoteAccessAddr); !ok {
		t.Fatal("remote access addr should be preserved after reset")
	}
	if sessionPtr.status != sessionUpdate {
		t.Fatalf("expected status update after reset, got %d", sessionPtr.status)
	}
}

func TestSessionSubmitOptionsAndTerminateNotifyObservers(t *testing.T) {
	registry := NewRegistry(nil).(*sessionRegistryImpl)
	defer registry.Release()

	observer := &testSessionObserver{id: "observer-1", statusCh: make(chan Status, 2)}
	sessionPtr := &sessionImpl{
		id: "session-submit",
		context: map[string]any{
			InnerStartTime:  int64(100),
			innerExpireTime: time.Now().Add(time.Minute).UTC().UnixMilli(),
		},
		observer: map[string]Observer{},
		registry: registry,
		status:   sessionUpdate,
	}

	sessionPtr.BindObserver(observer)
	sessionPtr.SubmitOptions()

	select {
	case status := <-observer.statusCh:
		if status != StatusUpdate {
			t.Fatalf("expected update status, got %v", status)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for update notification")
	}

	sessionPtr.terminate()
	select {
	case status := <-observer.statusCh:
		if status != StatusTerminate {
			t.Fatalf("expected terminate status, got %v", status)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for terminate notification")
	}
}

func TestRegistryCountDoesNotTerminateWorker(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	firstSession := registry.GetSession(nil, req)
	if firstSession == nil {
		t.Fatal("expected first session")
	}

	if got := registry.CountSession(nil); got != 1 {
		t.Fatalf("expected count 1, got %d", got)
	}

	nextReq := httptest.NewRequest("GET", "http://example.com/next", nil)
	secondSession := registry.GetSession(nil, nextReq)
	if secondSession == nil {
		t.Fatal("expected second session after count")
	}

	if got := registry.CountSession(nil); got != 2 {
		t.Fatalf("expected count 2 after second session, got %d", got)
	}
}

func TestRegistryReleaseIsIdempotent(t *testing.T) {
	registry := NewRegistry(nil)

	registry.Release()
	registry.Release()
}
