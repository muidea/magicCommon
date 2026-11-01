package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/framework/configuration"
	"github.com/muidea/magicCommon/framework/health"
	"github.com/muidea/magicCommon/framework/plugin/initiator"
	"github.com/muidea/magicCommon/framework/plugin/module"
	"github.com/muidea/magicCommon/task"
)

type healthTimer struct {
	task.BackgroundRoutine
	job task.Task
}

func (b *healthTimer) Timer(_ context.Context, job task.Task, _, _ time.Duration) error {
	b.job = job
	return nil
}

type healthAdmission struct {
	id           string
	closed       atomic.Bool
	released     atomic.Bool
	panicOnClose bool
}

func (s *healthAdmission) ID() string { return s.id }
func (s *healthAdmission) Setup(context.Context, event.Hub, task.BackgroundRoutine) *cd.Error {
	s.closed.Store(false)
	s.released.Store(false)
	return nil
}
func (s *healthAdmission) Run(context.Context) *cd.Error { return nil }
func (s *healthAdmission) BeginShutdown(context.Context) {
	s.closed.Store(true)
	if s.panicOnClose {
		panic("close failed")
	}
}
func (s *healthAdmission) Teardown(context.Context) { s.released.Store(true) }

func TestServiceRetainsDependenciesUntilProbeActuallyDrains(t *testing.T) {
	health.ResetDefaultManager()
	t.Cleanup(health.ResetDefaultManager)
	m := health.DefaultManager()
	input, producer := &healthAdmission{id: "health-input"}, &healthAdmission{id: "health-producer"}
	initiator.MustRegister(input)
	module.MustRegister(producer)
	if err := initiator.Setup(t.Context(), nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := module.Setup(t.Context(), nil, nil); err != nil {
		t.Fatal(err)
	}
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	calls := 0
	m.RegisterDependencyChecker("database", func(context.Context) *cd.Error {
		calls++
		if calls > 1 {
			close(entered)
			<-release
		}
		return nil
	})
	bg := &healthTimer{}
	monitor, err := m.StartDependencyMonitoring(t.Context(), bg, []health.Dependency{health.Required("database")}, health.DefaultProbePolicy())
	if err != nil {
		t.Fatal(err)
	}
	m.MarkReady()
	s := &defaultService{monitor: monitor, initiatorsStarted: true, modulesStarted: true}
	go func() { bg.job.Run(); close(done) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("probe did not start")
	}
	t.Cleanup(func() {
		close(release)
		<-done
		if err := s.ShutdownChecked(context.Background()); err != nil {
			t.Error(err)
		}
		if !s.tornDown || !input.released.Load() || !producer.released.Load() {
			t.Error("shutdown retry did not release drained owners")
		}
		bg.job.Run()
		if calls != 2 {
			t.Error("queued task accessed released dependency")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Millisecond)
	defer cancel()
	if err := s.ShutdownChecked(ctx); err == nil {
		t.Fatal("blocked probe reported shutdown success")
	}
	if s.tornDown || s.quiesced || m.Snapshot().Status != health.StatusStopping {
		t.Fatal("shutdown released dependencies or readiness too soon")
	}
	if !input.closed.Load() || !producer.closed.Load() {
		t.Fatal("probe timeout left an input owner open")
	}
	if input.released.Load() || producer.released.Load() {
		t.Fatal("dependencies released while probe still running")
	}
}

func TestServiceNotifiesAllOwnersWhenShutdownStarterPanics(t *testing.T) {
	health.ResetDefaultManager()
	t.Cleanup(health.ResetDefaultManager)
	input := &healthAdmission{id: "health-panic-input", panicOnClose: true}
	producer := &healthAdmission{id: "health-panic-producer"}
	initiator.MustRegister(input)
	module.MustRegister(producer)
	if err := initiator.Setup(t.Context(), nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := module.Setup(t.Context(), nil, nil); err != nil {
		t.Fatal(err)
	}
	m := health.DefaultManager()
	calls := 0
	m.RegisterDependencyChecker("db", func(context.Context) *cd.Error { calls++; return nil })
	bg := &healthTimer{}
	monitor, err := m.StartDependencyMonitoring(t.Context(), bg, []health.Dependency{health.Required("db")}, health.DefaultProbePolicy())
	if err != nil {
		t.Fatal(err)
	}
	s := &defaultService{monitor: monitor, initiatorsStarted: true, modulesStarted: true}
	t.Cleanup(func() {
		input.panicOnClose = false
		if err := s.ShutdownChecked(context.Background()); err != nil {
			t.Error(err)
		}
	})
	if err := s.Quiesce(t.Context()); err == nil {
		t.Fatal("shutdown panic was hidden")
	}
	if !input.closed.Load() || !producer.closed.Load() {
		t.Fatal("shutdown failed to notify all owners")
	}
	bg.job.Run()
	if calls != 1 || input.released.Load() || producer.released.Load() || s.quiesced {
		t.Fatal("failed shutdown kept probe admission open or released resources")
	}
}

func TestServiceStartupRejectsCanceledParentAfterOptionalCheck(t *testing.T) {
	health.ResetDefaultManager()
	t.Cleanup(health.ResetDefaultManager)
	if err := configuration.InitDefaultConfigManager(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configuration.CloseConfigManager() })
	m := health.DefaultManager()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if err := m.RegisterDependency(health.Optional("optional")); err != nil {
		t.Fatal(err)
	}
	m.RegisterDependencyChecker("optional", func(context.Context) *cd.Error { cancel(); return nil })
	s := &defaultService{}
	t.Cleanup(func() { _ = s.ShutdownChecked(context.Background()) })
	if err := s.Startup(ctx, "canceled-startup", nil, nil); err == nil {
		t.Fatal("optional check hid canceled startup context")
	}
	if s.modulesStarted || m.Snapshot().Status != health.StatusFailed {
		t.Fatal("canceled startup entered modules or did not fail readiness")
	}
}
