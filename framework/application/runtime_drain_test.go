package application

import (
	"context"
	"errors"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/framework/service"
	"github.com/muidea/magicCommon/task"
)

func TestRuntimeDrainPrecedesFinalServiceTeardown(t *testing.T) {
	for _, mode := range []string{"queue", "hub"} {
		t.Run(mode, func(t *testing.T) {
			hub, queue := event.NewHub(8), task.NewBackgroundRoutine(2)
			svc := &guardedService{}
			app := NewApplication(Options{ConfigDir: t.TempDir(), EventHub: hub, BackgroundRoutine: queue, Ownership: RuntimeOwnership{EventHub: true, BackgroundRoutine: true}})
			if err := app.Startup(context.Background(), svc); err != nil {
				t.Fatal(err)
			}
			observer := event.NewSimpleObserver("owner", hub)
			if err := observer.Subscribe("probe", func(_ event.Event, result event.Result) { result.Set(true, nil) }); err != nil {
				t.Fatal(err)
			}
			entered, release, done := make(chan struct{}), make(chan struct{}), make(chan event.Result, 1)
			work := func() { close(entered); <-release; done <- hub.Send(event.NewEvent("probe", "job", "owner", nil, nil)) }
			if mode == "queue" {
				if err := queue.AsyncFunction(work); err != nil {
					t.Fatal(err)
				}
			} else {
				other := event.NewSimpleObserver("other", hub)
				if err := other.Subscribe("work", func(_ event.Event, result event.Result) { work(); result.Set(nil, nil) }); err != nil {
					t.Fatal(err)
				}
				go hub.Send(event.NewEvent("work", "caller", "other", nil, nil))
			}
			<-entered
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			if err := app.ShutdownChecked(ctx); err == nil {
				t.Error("live runtime was reported shut down")
			}
			if svc.shutdownCalled || app.EventHub() != hub || app.BackgroundRoutine() != queue {
				t.Error("incomplete runtime drain released dependencies")
			}
			close(release)
			if result := <-done; result == nil || result.Error() != nil {
				t.Error("live task lost its hub dependency", result)
			}
			if err := app.ShutdownChecked(context.Background()); err != nil {
				t.Fatal(err)
			}
			if !svc.shutdownCalled {
				t.Fatal("drained service was not torn down")
			}
		})
	}
}

type finalStageService struct {
	guardedService
	fail, panics bool
	attempts     int
}

func (s *finalStageService) ShutdownChecked(context.Context) *cd.Error {
	s.attempts++
	if s.panics {
		panic("teardown panic")
	}
	if s.fail {
		return cd.NewError(cd.Unexpected, "teardown failed")
	}
	return nil
}

func TestFinalTeardownErrorRetainsHubAndDoesNotRedrainQueue(t *testing.T) {
	for _, panics := range []bool{false, true} {
		hub, queue := &retainedHub{}, &retainedQueue{}
		svc := &finalStageService{fail: true, panics: panics}
		app := NewApplication(Options{ConfigDir: t.TempDir(), EventHub: hub, BackgroundRoutine: queue, Ownership: RuntimeOwnership{EventHub: true, BackgroundRoutine: true}})
		if err := app.Startup(context.Background(), svc); err != nil {
			t.Fatal(err)
		}
		if err := app.ShutdownChecked(context.Background()); err == nil {
			t.Fatal("final teardown failure was discarded")
		}
		if hub.closes != 0 || queue.closes != 1 {
			t.Fatal("failed final teardown advanced dependent phases")
		}
		svc.fail, svc.panics = false, false
		if err := app.ShutdownChecked(context.Background()); err != nil {
			t.Fatal(err)
		}
		if hub.closes != 1 || queue.closes != 1 || svc.attempts != 2 {
			t.Fatal("incorrect phase retry")
		}
	}
}

func TestLifecycleAdapterPropagatesShutdownFailure(t *testing.T) {
	lifecycle := &mockLifecycle{shutdownErr: errors.New("still running")}
	svc := service.AdaptLifecycle("audit", lifecycle)
	if err := svc.(service.CheckedShutdown).ShutdownChecked(context.Background()); err == nil {
		t.Fatal("lifecycle error discarded")
	}
}

type checkedRetainedHub struct {
	retainedHub
	fail bool
}

func (h *checkedRetainedHub) Drain(context.Context) *cd.Error { return nil }
func (h *checkedRetainedHub) TerminateChecked(context.Context) *cd.Error {
	if h.fail {
		return cd.NewError(cd.Timeout, "hub still stopping")
	}
	h.closes++
	return nil
}

func TestHubFinalizationFailureDoesNotRepeatCompletedServiceStage(t *testing.T) {
	hub, queue := &checkedRetainedHub{fail: true}, &retainedQueue{}
	svc := &finalStageService{}
	app := NewApplication(Options{ConfigDir: t.TempDir(), EventHub: hub, BackgroundRoutine: queue, Ownership: RuntimeOwnership{EventHub: true, BackgroundRoutine: true}})
	if err := app.Startup(context.Background(), svc); err != nil {
		t.Fatal(err)
	}
	if err := app.ShutdownChecked(context.Background()); err == nil {
		t.Fatal("hub finalization failure discarded")
	}
	if app.EventHub() != hub || app.Run(context.Background()) == nil {
		t.Fatal("incomplete finalization reset the runtime")
	}
	hub.fail = false
	if err := app.ShutdownChecked(context.Background()); err != nil {
		t.Fatal(err)
	}
	if svc.attempts != 1 || queue.closes != 1 || hub.closes != 1 {
		t.Fatal("completed phases were executed again")
	}
}
