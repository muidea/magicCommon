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

func TestAnonymousSessionSignatureCanBeLoadedIntoRegistry(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()

	sessionPtr := NewAnonymousSession("127.0.0.1", "ua")
	sessionPtr.SetOption("custom", "value")

	token, err := sessionPtr.Signature()
	if err != nil {
		t.Fatalf("Signature() failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: SessionToken, Value: token})

	loaded := LookupSession(registry, req)
	if loaded == nil {
		t.Fatal("expected session to be loaded from signed anonymous session")
	}
	if loaded.ID() != sessionPtr.ID() {
		t.Fatalf("session ID = %s, want %s", loaded.ID(), sessionPtr.ID())
	}
	if val, ok := loaded.GetString("custom"); !ok || val != "value" {
		t.Fatalf("custom value = %q, %v, want value, true", val, ok)
	}
	if got := registry.CountSession(nil); got != 1 {
		t.Fatalf("registry count = %d, want 1", got)
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

func TestLookupSessionRefreshesExistingSessionClaimsFromNewJWT(t *testing.T) {
	const (
		authScopeKey  = "X-Mp-Auth-Scope"
		authEntityKey = "X-Mp-Auth-Entity"
	)

	registry := NewRegistry(nil)
	defer registry.Release()

	initialSession := &sessionImpl{
		id: "shared-session-id",
		context: map[string]any{
			InnerStartTime:        time.Now().Add(-time.Minute).UTC().UnixMilli(),
			innerExpireTime:       time.Now().Add(time.Minute).UTC().UnixMilli(),
			authScopeKey:          "autotest:read",
			authEntityKey:         map[string]any{"id": float64(1), "eID": float64(7), "eType": "account", "eName": "demo", "status": float64(1)},
		},
		observer: map[string]Observer{},
		status:   sessionActive,
	}
	initialToken, err := initialSession.Signature()
	if err != nil {
		t.Fatalf("initial Signature() failed: %v", err)
	}

	initialReq := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	initialReq.Header.Set(Authorization, "Bearer "+initialToken)
	loadedInitial := LookupSession(registry, initialReq)
	if loadedInitial == nil {
		t.Fatal("expected initial session to load")
	}

	refreshedSession := &sessionImpl{
		id: "shared-session-id",
		context: map[string]any{
			InnerStartTime:        time.Now().UTC().UnixMilli(),
			innerExpireTime:       time.Now().Add(9 * time.Minute).UTC().UnixMilli(),
			authScopeKey:          "autotest:*;panel:*",
			authEntityKey:         map[string]any{"id": float64(2), "eID": float64(9), "eType": "account", "eName": "refreshed", "status": float64(1)},
		},
		observer: map[string]Observer{},
		status:   sessionActive,
	}
	refreshedToken, err := refreshedSession.Signature()
	if err != nil {
		t.Fatalf("refreshed Signature() failed: %v", err)
	}

	refreshedReq := httptest.NewRequest(http.MethodGet, "http://example.com/refreshed", nil)
	refreshedReq.Header.Set(Authorization, "Bearer "+refreshedToken)
	loadedRefreshed := LookupSession(registry, refreshedReq)
	if loadedRefreshed == nil {
		t.Fatal("expected refreshed session to load")
	}
	if loadedRefreshed.ID() != "shared-session-id" {
		t.Fatalf("session ID = %s, want shared-session-id", loadedRefreshed.ID())
	}

	scopeVal, ok := loadedRefreshed.GetString(authScopeKey)
	if !ok || scopeVal != "autotest:*;panel:*" {
		t.Fatalf("scope=%q, ok=%v, want autotest:*;panel:*", scopeVal, ok)
	}

	entityVal, ok := loadedRefreshed.GetOption(authEntityKey)
	if !ok {
		t.Fatal("expected refreshed auth entity")
	}
	entityMap, ok := entityVal.(map[string]any)
	if !ok {
		t.Fatalf("entity type = %T, want map[string]any", entityVal)
	}
	if got := entityMap["eName"]; got != "refreshed" {
		t.Fatalf("entity name = %v, want refreshed", got)
	}

	expireVal, ok := loadedRefreshed.GetInt(innerExpireTime)
	if !ok || expireVal < time.Now().Add(8*time.Minute).UTC().UnixMilli() {
		t.Fatalf("expire=%d, ok=%v, expected refreshed expiry", expireVal, ok)
	}
}

func TestLookupSessionRefreshKeepsLocalUnsignedContext(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()
	registryImpl := registry.(*sessionRegistryImpl)

	initialSession := &sessionImpl{
		id: "shared-session-id",
		context: map[string]any{
			InnerStartTime:        time.Now().Add(-time.Minute).UTC().UnixMilli(),
			innerExpireTime:       time.Now().Add(time.Minute).UTC().UnixMilli(),
			"X-Mp-Auth-Entity":    map[string]any{"id": float64(1), "eID": float64(7), "eType": "account", "eName": "demo", "status": float64(1)},
			"_AuthRole":           "cached-role",
			"_authType":           AuthJWTSession,
			"_verifiedNamespace":  "example",
			"_verifiedAt":         int64(1234567890),
		},
		observer: map[string]Observer{},
		status:   sessionActive,
	}
	loadedInitial := registryImpl.insertSession(initialSession)
	if loadedInitial == nil {
		t.Fatal("expected initial session to load")
	}

	refreshedSession := &sessionImpl{
		id: "shared-session-id",
		context: map[string]any{
			InnerStartTime:      time.Now().UTC().UnixMilli(),
			innerExpireTime:     time.Now().Add(9 * time.Minute).UTC().UnixMilli(),
			"X-Mp-Auth-Entity":  map[string]any{"id": float64(2), "eID": float64(9), "eType": "account", "eName": "refreshed", "status": float64(1)},
		},
		observer: map[string]Observer{},
		status:   sessionActive,
	}
	refreshedToken, err := refreshedSession.Signature()
	if err != nil {
		t.Fatalf("refreshed Signature() failed: %v", err)
	}

	refreshedReq := httptest.NewRequest(http.MethodGet, "http://example.com/refreshed", nil)
	refreshedReq.Header.Set(Authorization, "Bearer "+refreshedToken)
	loadedRefreshed := LookupSession(registry, refreshedReq)
	if loadedRefreshed == nil {
		t.Fatal("expected refreshed session to load")
	}

	if roleVal, ok := loadedRefreshed.GetOption("_AuthRole"); !ok || roleVal != "cached-role" {
		t.Fatalf("unsigned role cache = %#v, %v, want cached-role, true", roleVal, ok)
	}
	if authType, ok := loadedRefreshed.GetString("_authType"); !ok || authType != AuthJWTSession {
		t.Fatalf("unsigned authType = %q, %v, want %q, true", authType, ok, AuthJWTSession)
	}
	if verifiedNamespace, ok := loadedRefreshed.GetString("_verifiedNamespace"); !ok || verifiedNamespace != "example" {
		t.Fatalf("unsigned verified namespace = %q, %v, want example, true", verifiedNamespace, ok)
	}
	if verifiedAt, ok := loadedRefreshed.GetInt("_verifiedAt"); !ok || verifiedAt != 1234567890 {
		t.Fatalf("unsigned verifiedAt = %d, %v, want 1234567890, true", verifiedAt, ok)
	}
}

func TestLookupSessionValidJWTAccessRefreshesExistingSessionExpiry(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()

	initialSession := &sessionImpl{
		id: "shared-session-id",
		context: map[string]any{
			InnerStartTime:  time.Now().Add(-time.Minute).UTC().UnixMilli(),
			innerExpireTime: time.Now().Add(time.Minute).UTC().UnixMilli(),
			"scope":         "demo:*",
		},
		observer: map[string]Observer{},
		status:   sessionActive,
	}
	initialToken, err := initialSession.Signature()
	if err != nil {
		t.Fatalf("initial Signature() failed: %v", err)
	}

	initialReq := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	initialReq.Header.Set(Authorization, "Bearer "+initialToken)
	loadedInitial := LookupSession(registry, initialReq)
	if loadedInitial == nil {
		t.Fatal("expected initial session to load")
	}

	loadedInitial.SetOption(innerExpireTime, time.Now().Add(time.Second).UTC().UnixMilli())
	oldExpire, ok := loadedInitial.GetInt(innerExpireTime)
	if !ok {
		t.Fatal("expected innerExpireTime on initial session")
	}

	// 再次使用同一合法 JWT 访问时，应刷新本地 session 有效期，而不是继续沿用旧的 innerExpireTime。
	secondReq := httptest.NewRequest(http.MethodGet, "http://example.com/next", nil)
	secondReq.Header.Set(Authorization, "Bearer "+initialToken)
	loadedAgain := LookupSession(registry, secondReq)
	if loadedAgain == nil {
		t.Fatal("expected refreshed session to load")
	}

	newExpire, ok := loadedAgain.GetInt(innerExpireTime)
	if !ok {
		t.Fatal("expected refreshed innerExpireTime")
	}
	if newExpire <= oldExpire {
		t.Fatalf("innerExpireTime=%d want > %d", newExpire, oldExpire)
	}
	if newExpire < time.Now().Add(9*time.Minute).UTC().UnixMilli() {
		t.Fatalf("innerExpireTime=%d expected local session to be refreshed", newExpire)
	}
}

func TestLookupSessionValidJWTRecreatesSessionAfterLocalFinal(t *testing.T) {
	registry := NewRegistry(nil)
	defer registry.Release()
	registryImpl := registry.(*sessionRegistryImpl)

	validSession := &sessionImpl{
		id: "shared-session-id",
		context: map[string]any{
			InnerStartTime:  time.Now().Add(-time.Minute).UTC().UnixMilli(),
			innerExpireTime: time.Now().Add(time.Minute).UTC().UnixMilli(),
			AuthExpireTime:  time.Now().Add(9 * time.Minute).UTC().UnixMilli(),
			"scope":         "demo:*",
		},
		observer: map[string]Observer{},
		status:   sessionActive,
	}
	validToken, err := validSession.Signature()
	if err != nil {
		t.Fatalf("Signature() failed: %v", err)
	}

	staleLocal := &sessionImpl{
		id: "shared-session-id",
		context: map[string]any{
			InnerStartTime:  time.Now().Add(-11 * time.Minute).UTC().UnixMilli(),
			innerExpireTime: time.Now().Add(-time.Minute).UTC().UnixMilli(),
		},
		observer: map[string]Observer{},
		status:   sessionTerminate,
		registry: registryImpl,
	}
	registryImpl.sessionMap[staleLocal.id] = staleLocal

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set(Authorization, "Bearer "+validToken)
	loaded := LookupSession(registry, req)
	if loaded == nil {
		t.Fatal("expected valid JWT to recreate session from local final state")
	}
	if loaded.ID() != "shared-session-id" {
		t.Fatalf("session ID = %s, want shared-session-id", loaded.ID())
	}
	expireVal, ok := loaded.GetInt(AuthExpireTime)
	if !ok || expireVal < time.Now().Add(8*time.Minute).UTC().UnixMilli() {
		t.Fatalf("authExpireTime=%d ok=%v expected valid recreated auth session", expireVal, ok)
	}
}
