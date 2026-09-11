package application

import (
	"context"
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
	"testing"
)

type guardedService struct {
	MockService
	rejected bool
	panics   bool
}

func (s *guardedService) Quiesce(context.Context) *cd.Error {
	if s.panics {
		panic("guard failure")
	}
	if s.rejected {
		return cd.NewError(cd.Timeout, "still executing")
	}
	return nil
}

type retainedHub struct {
	event.Hub
	closes int
}

func (h *retainedHub) Terminate(context.Context) { h.closes++ }

type retainedQueue struct {
	task.BackgroundRoutine
	closes int
}

func (q *retainedQueue) Shutdown(context.Context) bool { q.closes++; return true }

func TestCheckedShutdownRetainsDependenciesAndCanRetry(t *testing.T) {
	for _, mode := range []string{"timeout", "panic", "startup-failure"} {
		t.Run(mode, func(t *testing.T) {
			h, q := &retainedHub{}, &retainedQueue{}
			svc := &guardedService{rejected: true, panics: mode == "panic"}
			if mode == "startup-failure" {
				svc.startupError = cd.NewError(cd.Unexpected, "startup failed")
			}
			app := NewApplication(Options{ConfigDir: t.TempDir(), EventHub: h, BackgroundRoutine: q, Ownership: RuntimeOwnership{EventHub: true, BackgroundRoutine: true}})
			startErr := app.Startup(context.Background(), svc)
			if (startErr != nil) != (mode == "startup-failure") {
				t.Fatal("unexpected startup result", startErr)
			}
			if err := app.ShutdownChecked(context.Background()); err == nil {
				t.Fatal("guard failure was ignored")
			}
			if h.closes != 0 || q.closes != 0 || svc.shutdownCalled || app.EventHub() != h || app.BackgroundRoutine() != q {
				t.Fatal("failed guard released live dependencies")
			}
			if app.Run(context.Background()) == nil || app.Startup(context.Background(), svc) == nil {
				t.Fatal("stopping application admitted work or restart")
			}
			svc.rejected, svc.panics = false, false
			if err := app.ShutdownChecked(context.Background()); err != nil {
				t.Fatal(err)
			}
			if h.closes != 1 || q.closes != 1 || !svc.shutdownCalled {
				t.Fatal("retry did not release dependencies exactly once")
			}
			app.Shutdown(context.Background())
			if h.closes != 1 || q.closes != 1 {
				t.Fatal("repeated shutdown released twice")
			}
		})
	}
}
