package health

import (
	"context"
	"sync"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/task"
)

// ProbePolicy bounds one complete probe round and the lifetime of its result.
// Dependency callbacks must honor cancellation; actual completion is tracked
// even if a callback violates its budget, and stale results cannot keep ready.
type ProbePolicy struct {
	Interval time.Duration
	Timeout  time.Duration
	MaxAge   time.Duration
}

func DefaultProbePolicy() ProbePolicy {
	return ProbePolicy{Interval: 5 * time.Second, Timeout: 3 * time.Second, MaxAge: 15 * time.Second}
}
func (p ProbePolicy) Validate() *cd.Error {
	if p.Interval <= 0 || p.Timeout <= 0 || p.MaxAge < p.Interval || p.Timeout > p.MaxAge-p.Interval {
		return cd.NewError(cd.IllegalParam, "dependency probe requires positive interval/timeout and max age at least their sum")
	}
	return nil
}

// DependencyMonitor is owned by one process Service. Timer and probe execution
// use the injected BackgroundRoutine, and Stop precedes dependency teardown.
type DependencyMonitor struct {
	mu          sync.Mutex
	manager     *Manager
	ctx         context.Context
	cancel      context.CancelFunc
	deps        []Dependency
	policy      ProbePolicy
	closed      bool
	running     bool
	done        chan struct{}
	completedAt time.Time
}

func (s *Manager) StartDependencyMonitoring(ctx context.Context, bg task.BackgroundRoutine, deps []Dependency, policy ProbePolicy) (*DependencyMonitor, *cd.Error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	if ctx == nil || ctx.Err() != nil {
		return nil, cd.NewError(cd.InvalidOperation, "dependency monitor context is unavailable")
	}
	if len(deps) == 0 {
		return nil, nil
	}
	if bg == nil {
		return nil, cd.NewError(cd.InvalidOperation, "dependency monitor requires the process background routine")
	}
	probeCtx, cancel := context.WithTimeout(ctx, policy.Timeout)
	err := s.CheckDependencies(probeCtx, deps)
	cancel()
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, cd.NewError(cd.InvalidOperation, "dependency monitor context is unavailable")
	}
	monitorCtx, monitorCancel := context.WithCancel(ctx)
	monitor := &DependencyMonitor{manager: s, ctx: monitorCtx, cancel: monitorCancel, deps: append([]Dependency(nil), deps...), policy: policy, completedAt: time.Now()}
	s.mu.Lock()
	s.maxAge = policy.MaxAge
	s.mu.Unlock()
	if err := bg.Timer(monitorCtx, monitor, policy.Interval, 0); err != nil {
		monitorCancel()
		return nil, cd.NewError(cd.ServiceUnavailable, "dependency monitor scheduling failed")
	}
	return monitor, nil
}

// Run coalesces overlapping/queued timer ticks; at most one real probe round
// runs. A canceled queued task returns without touching released dependencies.
func (m *DependencyMonitor) Run() {
	m.RunScheduled(time.Now())
}

// RunScheduled discards ticks covered by an already completed round. Using
// the timer timestamp preserves fresh ticks despite variable queue delays.
func (m *DependencyMonitor) RunScheduled(scheduledAt time.Time) {
	m.mu.Lock()
	if m.closed || m.running || m.ctx.Err() != nil || !scheduledAt.After(m.completedAt) {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.done = make(chan struct{})
	done := m.done
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		m.completedAt = time.Now()
		m.running = false
		close(done)
		m.mu.Unlock()
	}()
	ctx, cancel := context.WithTimeout(m.ctx, m.policy.Timeout)
	defer cancel()
	_ = m.manager.CheckDependencies(ctx, m.deps) // the atomic snapshot retains failure
}

// BeginShutdown cancels scheduling and closes admission without waiting.
func (m *DependencyMonitor) BeginShutdown() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.closed = true
	m.cancel()
	m.mu.Unlock()
}

// Stop cancels scheduling and waits for the actual in-flight callback. A
// timeout is incomplete shutdown, so the Service must retain its dependencies.
func (m *DependencyMonitor) Stop(ctx context.Context) *cd.Error {
	if m == nil {
		return nil
	}
	if ctx == nil {
		return cd.NewError(cd.IllegalParam, "dependency monitor stop context is required")
	}
	m.BeginShutdown()
	m.mu.Lock()
	running, done := m.running, m.done
	m.mu.Unlock()
	if running {
		select {
		case <-done:
		case <-ctx.Done():
			return cd.NewError(cd.Timeout, "dependency checks have not drained")
		}
	}
	return nil
}
