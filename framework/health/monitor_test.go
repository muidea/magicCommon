package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/task"
)

// The captured timer permits deterministic fault and drain tests without
// running dependency callbacks outside the monitor's lifecycle.
type capturedTimer struct {
	task.BackgroundRoutine
	job task.Task
	err error
}

func (b *capturedTimer) Timer(_ context.Context, job task.Task, _, _ time.Duration) error {
	b.job = job
	return b.err
}

func probePolicy() ProbePolicy {
	return ProbePolicy{Interval: time.Millisecond, Timeout: 20 * time.Millisecond, MaxAge: time.Second}
}

func assertReadyHTTP(t *testing.T, manager *Manager, want int) {
	t.Helper()
	r := httptest.NewRecorder()
	manager.ReadyHandler(t.Context(), r, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if r.Code != want {
		t.Fatalf("ready status=%d, want=%d: %s", r.Code, want, r.Body.String())
	}
}

func TestDependencyMonitorFailureRecoveryAndStopping(t *testing.T) {
	m := NewManager()
	var requiredFailed, optionalFailed atomic.Bool
	m.RegisterDependencyChecker("database", func(context.Context) *cd.Error {
		if requiredFailed.Load() {
			return cd.NewError(cd.Unexpected, "password=must-never-appear")
		}
		return nil
	})
	m.RegisterDependencyChecker("telemetry", func(context.Context) *cd.Error {
		if optionalFailed.Load() {
			panic("secret-driver-error")
		}
		return nil
	})
	b := &capturedTimer{}
	monitor, err := m.StartDependencyMonitoring(t.Context(), b, []Dependency{Required("database"), Optional("telemetry")}, probePolicy())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = monitor.Stop(context.Background()) })
	m.MarkReady()
	optionalFailed.Store(true)
	monitor.Run()
	assertReadyHTTP(t, m, http.StatusOK)
	if m.Snapshot().Checks["telemetry"].Status != "failed" {
		t.Fatal("optional failure lost")
	}
	requiredFailed.Store(true)
	monitor.Run()
	assertReadyHTTP(t, m, http.StatusServiceUnavailable)
	encoded, _ := json.Marshal(m.Snapshot())
	if strings.Contains(string(encoded), "must-never") || strings.Contains(string(encoded), "secret-driver") {
		t.Fatal("checker details leaked")
	}
	live := httptest.NewRecorder()
	m.LiveHandler(t.Context(), live, nil)
	if live.Code != http.StatusOK {
		t.Fatal("dependency failure removed liveness")
	}
	requiredFailed.Store(false)
	monitor.Run()
	assertReadyHTTP(t, m, http.StatusOK)
	m.MarkStopping()
	monitor.Run()
	m.MarkReady()
	assertReadyHTTP(t, m, http.StatusServiceUnavailable)
}

func TestDependencyMonitorAtomicSnapshotStaleAndRealDrain(t *testing.T) {
	m := NewManager()
	entered, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	m.RegisterDependencyChecker("database", func(context.Context) *cd.Error {
		if calls.Add(1) > 1 {
			close(entered)
			<-release
		} // deliberately violates cancellation
		return nil
	})
	monitor, err := m.StartDependencyMonitoring(t.Context(), &capturedTimer{}, []Dependency{Required("database")}, probePolicy())
	if err != nil {
		t.Fatal(err)
	}
	m.MarkReady()
	before := m.Snapshot()
	go func() { monitor.Run(); close(finished) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("probe did not enter")
	}
	t.Cleanup(func() {
		close(release)
		<-finished
		if err := monitor.Stop(context.Background()); err != nil {
			t.Error(err)
		}
		monitor.Run()
		if calls.Load() != 2 || m.Snapshot().Checks["database"].Status != "failed" {
			t.Error("canceled probe published ready or queued probe ran")
		}
	})
	for i := 0; i < 20; i++ {
		monitor.Run()
	}
	if calls.Load() != 2 {
		t.Fatal("overlap was not coalesced")
	}
	if got := m.Snapshot(); got.CheckedAt != before.CheckedAt || got.Checks["database"].Status != "ready" {
		t.Fatal("in-flight round overwrote completed snapshot")
	}
	assertReadyHTTP(t, m, http.StatusOK)
	m.mu.Lock()
	m.checkedAt = time.Now().Add(-2 * time.Second)
	m.mu.Unlock()
	assertReadyHTTP(t, m, http.StatusServiceUnavailable)
	stopCtx, cancel := context.WithTimeout(t.Context(), 5*time.Millisecond)
	defer cancel()
	if err := monitor.Stop(stopCtx); err == nil {
		t.Fatal("uncompleted checker reported drained")
	}
	select {
	case <-finished:
		t.Fatal("checker was detached by cancellation")
	default:
	}
	// Actual callback completion and Stop retry are exercised by cleanup. Even
	// after it returns nil, the canceled callback cannot publish a ready check.
}

func TestDependencyMonitorCancellationAndSchedulingFailure(t *testing.T) {
	for _, mode := range []string{"unbound", "schedule", "canceled", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			m := NewManager()
			m.RegisterDependencyChecker("database", func(ctx context.Context) *cd.Error {
				if mode == "timeout" {
					<-ctx.Done()
				}
				return nil
			})
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var bg task.BackgroundRoutine = &capturedTimer{}
			if mode == "unbound" {
				bg = nil
			}
			if mode == "schedule" {
				bg = &capturedTimer{err: errors.New("queue closed")}
			}
			if mode == "canceled" {
				cancel()
			}
			if _, err := m.StartDependencyMonitoring(ctx, bg, []Dependency{Required("database")}, probePolicy()); err == nil {
				t.Fatal("startup accepted unavailable dependency monitor")
			}
		})
	}
	m := NewManager()
	var calls atomic.Int32
	m.RegisterDependencyChecker("database", func(context.Context) *cd.Error { calls.Add(1); return nil })
	b := &capturedTimer{}
	monitor, err := m.StartDependencyMonitoring(t.Context(), b, []Dependency{Required("database")}, probePolicy())
	if err != nil {
		t.Fatal(err)
	}
	if err := monitor.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
	b.job.Run()
	if calls.Load() != 1 {
		t.Fatal("queued task touched stopped dependency")
	}
}

func TestDependencyMonitorUsesBackgroundTimer(t *testing.T) {
	m := NewManager()
	var fail atomic.Bool
	m.RegisterDependencyChecker("database", func(context.Context) *cd.Error {
		if fail.Load() {
			return cd.NewError(cd.Unexpected, "offline")
		}
		return nil
	})
	bg := task.NewBackgroundRoutine(2)
	monitor, err := m.StartDependencyMonitoring(t.Context(), bg, []Dependency{Required("database")}, probePolicy())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = monitor.Stop(context.Background())
		if !bg.Shutdown(context.Background()) {
			t.Error("background did not drain")
		}
	})
	m.MarkReady()
	fail.Store(true)
	deadline := time.Now().Add(time.Second)
	for m.Snapshot().Status == StatusReady && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	assertReadyHTTP(t, m, http.StatusServiceUnavailable)
	fail.Store(false)
	deadline = time.Now().Add(time.Second)
	for m.Snapshot().Status != StatusReady && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	assertReadyHTTP(t, m, http.StatusOK)
}

func TestDependencyDeclarationAndHTTPIsolation(t *testing.T) {
	m := NewManager()
	for _, dep := range []Dependency{
		{Name: "db", Kind: "unsupported"},
		{Name: "db", Kind: RequiredDependency, Target: "http://user:secret@example.test"},
		{Name: "db", Kind: RequiredDependency, Target: "http://example.test?token=secret"},
	} {
		if m.RegisterDependency(dep) == nil {
			t.Fatal("invalid dependency accepted")
		}
	}
	if err := m.RegisterDependency(Required("db")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Dependencies([]Dependency{Optional("db")}); err == nil {
		t.Fatal("required dependency downgraded")
	}
	m.RegisterDependencyChecker("db", func(context.Context) *cd.Error { return nil })
	deps, err := m.Dependencies(nil)
	if err != nil || len(deps) != 1 || deps[0] != Required("db") {
		t.Fatal(deps, err)
	}
	var downstream atomic.Int32
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { downstream.Add(1); w.WriteHeader(200) }))
	defer other.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, other.URL, http.StatusFound) }))
	defer server.Close()
	if err := m.CheckDependencies(t.Context(), []Dependency{{Name: "remote", Kind: RequiredDependency, Target: server.URL}}); err == nil {
		t.Fatal("redirect accepted")
	}
	if downstream.Load() != 0 {
		t.Fatal("probe followed redirect")
	}
	assertReadyHTTP(t, m, http.StatusServiceUnavailable)
}

func TestOptionalOnlyStalenessDoesNotRemoveReadiness(t *testing.T) {
	m := NewManager()
	m.RegisterDependencyChecker("telemetry", func(context.Context) *cd.Error { return nil })
	monitor, err := m.StartDependencyMonitoring(t.Context(), &capturedTimer{}, []Dependency{Optional("telemetry")}, probePolicy())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := monitor.Stop(context.Background()); err != nil {
			t.Errorf("stop dependency monitor: %v", err)
		}
	}()
	m.MarkReady()
	m.mu.Lock()
	m.checkedAt = time.Now().Add(-time.Hour)
	m.mu.Unlock()
	assertReadyHTTP(t, m, http.StatusOK)
}

func TestHTTPDependencyMonitorFailureRecoveryAndSnapshotOnlyHandler(t *testing.T) {
	var offline atomic.Bool
	var requests atomic.Int32
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/health/ready" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("unexpected health path or business credential")
		}
		if offline.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer remote.Close()
	m := NewManager()
	monitor, err := m.StartDependencyMonitoring(t.Context(), &capturedTimer{}, []Dependency{{Name: "remote", Kind: RequiredDependency, Target: remote.URL}}, probePolicy())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := monitor.Stop(context.Background()); err != nil {
			t.Errorf("stop dependency monitor: %v", err)
		}
	}()
	m.MarkReady()
	assertReadyHTTP(t, m, http.StatusOK)
	if requests.Load() != 1 {
		t.Fatal("ready handler performed downstream IO")
	}
	offline.Store(true)
	monitor.Run()
	assertReadyHTTP(t, m, http.StatusServiceUnavailable)
	offline.Store(false)
	monitor.Run()
	assertReadyHTTP(t, m, http.StatusOK)
	if requests.Load() != 3 {
		t.Fatal("unexpected probe count")
	}
}
