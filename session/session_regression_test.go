package session

import (
	"net/http"
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
			t.Fatalf("expected StatusUpdate, got %v", status)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for update notification")
	}

	sessionPtr.terminate()
	select {
	case status := <-observer.statusCh:
		if status != StatusTerminate {
			t.Fatalf("expected StatusTerminate, got %v", status)
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

func TestOpaqueCookieRestoresLocalSessionOnly(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	created := registry.GetSession(httptest.NewRecorder(), req)
	if created == nil {
		t.Fatal("expected created session")
	}
	created.SetOption("custom", "value")
	id := created.ID()

	// Opaque cookie (session id) restores local container — not JWT claims.
	req2 := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req2.AddCookie(&http.Cookie{Name: SessionToken, Value: id})
	loaded := LookupSession(registry, req2)
	if loaded == nil {
		t.Fatal("expected local session restore via opaque cookie")
	}
	if loaded.ID() != id {
		t.Fatalf("session ID = %s, want %s", loaded.ID(), id)
	}
	if val, ok := loaded.GetString("custom"); !ok || val != "value" {
		t.Fatalf("custom value = %q, %v, want value, true", val, ok)
	}
}

func TestJWTCookieDoesNotRestoreAuthIdentity(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()

	// A compact JWT-shaped value must never be decoded by Registry.
	token := "legacy.header.signature"

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: SessionToken, Value: token})
	if loaded := LookupSession(registry, req); loaded != nil {
		t.Fatal("JWT cookie must not restore session identity into Registry")
	}
	if got := registry.CountSession(nil); got != 0 {
		t.Fatalf("registry count = %d, want 0", got)
	}
}

func TestBearerJWTDoesNotRestoreAuthIdentity(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()

	// Bearer access tokens belong to resource-service JWKS validation, not Registry.
	token := "access.header.signature"

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set(Authorization, "Bearer "+token)
	if loaded := LookupSession(registry, req); loaded != nil {
		t.Fatal("Bearer JWT must not restore session identity into Registry")
	}
}

func TestResolveSessionReturnsAnonymousWithoutPersisting(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()

	req := httptest.NewRequest(http.MethodGet, "http://example.com/public", nil)
	loaded := ResolveSession(registry, req)
	if loaded == nil {
		t.Fatal("expected anonymous session")
	}
	if got := registry.CountSession(nil); got != 0 {
		t.Fatalf("registry count = %d, want 0", got)
	}
}

func TestRegistryReleaseIsIdempotent(t *testing.T) {
	registry := NewRegistry(nil)

	registry.Release()
	registry.Release()
}

func TestOpaqueCookieRefreshUpdatesLocalExpiry(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	created := registry.GetSession(httptest.NewRecorder(), req)
	if created == nil {
		t.Fatal("expected created session")
	}
	id := created.ID()
	created.SetOption(innerExpireTime, time.Now().Add(time.Second).UTC().UnixMilli())
	oldExpire, ok := created.GetInt(innerExpireTime)
	if !ok {
		t.Fatal("expected innerExpireTime")
	}

	req2 := httptest.NewRequest(http.MethodGet, "http://example.com/next", nil)
	req2.AddCookie(&http.Cookie{Name: SessionToken, Value: id})
	loaded := LookupSession(registry, req2)
	if loaded == nil {
		t.Fatal("expected opaque cookie lookup")
	}
	newExpire, ok := loaded.GetInt(innerExpireTime)
	if !ok {
		t.Fatal("expected refreshed innerExpireTime")
	}
	if newExpire <= oldExpire {
		t.Fatalf("innerExpireTime=%d want > %d", newExpire, oldExpire)
	}
}

func TestFinalLocalSessionNotRestoredByOpaqueCookie(t *testing.T) {
	registry := NewRegistry(nil).(*sessionRegistryImpl)
	defer registry.Release()

	staleLocal := &sessionImpl{
		id: "shared-session-id",
		context: map[string]any{
			InnerStartTime:  time.Now().Add(-11 * time.Minute).UTC().UnixMilli(),
			innerExpireTime: time.Now().Add(-time.Minute).UTC().UnixMilli(),
		},
		observer: map[string]Observer{},
		status:   sessionTerminate,
		registry: registry,
	}
	registry.sessionMap[staleLocal.id] = staleLocal

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: SessionToken, Value: "shared-session-id"})
	if loaded := LookupSession(registry, req); loaded != nil {
		t.Fatal("final local session must not be restored")
	}
}
