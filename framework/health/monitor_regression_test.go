package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/task"
)

func TestOptionalTimeoutCannotStarveRequiredChecks(t *testing.T) {
	for _, mode := range []string{"checker", "http"} {
		t.Run(mode, func(t *testing.T) {
			m := NewManager()
			requiredCalls, optionalCalls := 0, 0
			requiredFailed := false
			m.RegisterDependencyChecker("z_required", func(context.Context) *cd.Error {
				requiredCalls++
				if requiredFailed {
					return cd.NewError(cd.ServiceUnavailable, "offline")
				}
				return nil
			})
			optional := Optional("a_optional")
			if mode == "checker" {
				m.RegisterDependencyChecker(optional.Name, func(ctx context.Context) *cd.Error {
					optionalCalls++
					<-ctx.Done()
					return nil
				})
			} else {
				server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
				defer server.Close()
				optional.Target = server.URL
			}
			if err := m.RegisterDependency(optional); err != nil {
				t.Fatal(err)
			}
			deps, err := m.Dependencies(nil) // lexical order puts optional first
			if err != nil || len(deps) != 2 || deps[0].Name != optional.Name {
				t.Fatalf("unexpected declarations: %v, %v", deps, err)
			}
			policy := ProbePolicy{Interval: time.Second, Timeout: 20 * time.Millisecond, MaxAge: 2 * time.Second}
			monitor, err := m.StartDependencyMonitoring(t.Context(), &capturedTimer{}, deps, policy)
			if err != nil {
				t.Fatal("optional timeout rejected startup:", err)
			}
			defer monitor.Stop(context.Background())
			m.MarkReady()
			for range 2 {
				assertReadyHTTP(t, m, http.StatusOK)
				checks := m.Snapshot().Checks
				if checks["z_required"].Status != "ready" || checks[optional.Name].Status != "failed" {
					t.Fatal("optional timeout corrupted required health", checks)
				}
				monitor.Run()
			}
			if requiredCalls != 3 || (mode == "checker" && optionalCalls != 3) {
				t.Fatalf("checks did not execute: required=%d optional=%d", requiredCalls, optionalCalls)
			}
			requiredFailed = true
			monitor.Run()
			assertReadyHTTP(t, m, http.StatusServiceUnavailable)
			if _, err := m.StartDependencyMonitoring(t.Context(), &capturedTimer{}, deps, policy); err == nil {
				t.Fatal("required failure accepted at startup")
			}
		})
	}
}

func TestDependencyDeclarationsValidatedBeforeIO(t *testing.T) {
	for _, invalid := range [][]Dependency{
		{Required("db"), Optional("db")},
		{Required("db"), {Name: "bad", Kind: "invalid"}},
	} {
		m := NewManager()
		m.RegisterDependencyChecker("db", func(context.Context) *cd.Error { t.Error("invalid round performed IO"); return nil })
		if err := m.CheckDependencies(t.Context(), invalid); err == nil {
			t.Fatal("invalid declaration accepted")
		}
		if m.Snapshot().CheckedAt != "" {
			t.Fatal("invalid round published partial checks")
		}
	}
}

func TestOptionalCheckCannotHideCanceledMonitorParent(t *testing.T) {
	m := NewManager()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	m.RegisterDependencyChecker("required", func(context.Context) *cd.Error { return nil })
	m.RegisterDependencyChecker("optional", func(context.Context) *cd.Error { cancel(); return nil })
	if _, err := m.StartDependencyMonitoring(ctx, &capturedTimer{}, []Dependency{Optional("optional"), Required("required")}, probePolicy()); err == nil {
		t.Fatal("canceled parent accepted as optional timeout")
	}
}

func TestDependencyMonitorFixedTicksSurviveDispatchDelay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := NewManager()
		var calls atomic.Int32
		m.RegisterDependencyChecker("required", func(context.Context) *cd.Error { calls.Add(1); return nil })
		bg := task.NewBackgroundRoutine(1)
		release := make(chan struct{})
		if err := bg.AsyncFunction(func() { <-release }); err != nil {
			t.Fatal(err)
		}
		policy := ProbePolicy{Interval: 100 * time.Millisecond, Timeout: 10 * time.Millisecond, MaxAge: 110 * time.Millisecond}
		monitor, err := m.StartDependencyMonitoring(t.Context(), bg, []Dependency{Required("required")}, policy)
		if err != nil {
			t.Fatal(err)
		}
		m.MarkReady()
		time.Sleep(130 * time.Millisecond) // 100ms tick executes at 130ms
		close(release)
		synctest.Wait()
		if calls.Load() != 2 {
			t.Fatal("delayed tick did not run", calls.Load())
		}
		time.Sleep(115 * time.Millisecond) // 200ms tick must run, despite <100ms execution gap
		synctest.Wait()
		assertReadyHTTP(t, m, http.StatusOK)
		if calls.Load() != 3 {
			t.Fatal("fresh timer tick was skipped", calls.Load())
		}
		if err := monitor.Stop(t.Context()); err != nil || !bg.Shutdown(t.Context()) {
			t.Fatal("monitor did not drain", err)
		}
	})
}

func TestDependencyMonitorCoalescesQueuedTickTimestamps(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m := NewManager()
		var calls atomic.Int32
		m.RegisterDependencyChecker("required", func(context.Context) *cd.Error { calls.Add(1); return nil })
		bg := task.NewBackgroundRoutine(1)
		release := make(chan struct{})
		if err := bg.AsyncFunction(func() { <-release }); err != nil {
			t.Fatal(err)
		}
		monitor, err := m.StartDependencyMonitoring(t.Context(), bg, []Dependency{Required("required")}, ProbePolicy{Interval: 100 * time.Millisecond, Timeout: 10 * time.Millisecond, MaxAge: time.Second})
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(350 * time.Millisecond) // ticks queue at 100, 200, 300ms
		close(release)
		synctest.Wait()
		if calls.Load() != 2 {
			t.Fatal("backlog produced extra real rounds", calls.Load())
		}
		time.Sleep(50 * time.Millisecond)
		synctest.Wait()
		if calls.Load() != 3 {
			t.Fatal("fresh tick after coalescing was lost", calls.Load())
		}
		if err := monitor.Stop(t.Context()); err != nil {
			t.Fatal(err)
		}
		monitor.RunScheduled(time.Now().Add(time.Hour))
		if !bg.Shutdown(t.Context()) || calls.Load() != 3 {
			t.Fatal("stopped queued tick accessed dependency", calls.Load())
		}
	})
}
