package application

import (
	"context"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/framework/service"
)

type executionApplication struct {
	Application
	startupErr, runErr *cd.Error
	panicRun           bool
	runs               int
	shutdownContexts   []context.Context
	t                  *testing.T
}

func (a *executionApplication) Startup(context.Context, service.Service) *cd.Error {
	return a.startupErr
}
func (a *executionApplication) Run(context.Context) *cd.Error {
	a.runs++
	if a.panicRun {
		panic("run panic")
	}
	return a.runErr
}
func (a *executionApplication) ShutdownChecked(ctx context.Context) *cd.Error {
	if ctx.Err() != nil {
		a.t.Error("shutdown inherited canceled run context")
	}
	if _, ok := ctx.Deadline(); !ok {
		a.t.Error("shutdown has no budget")
	}
	for _, previous := range a.shutdownContexts {
		if previous == ctx || previous.Err() == nil {
			a.t.Error("retry did not release previous budget")
		}
	}
	a.shutdownContexts = append(a.shutdownContexts, ctx)
	if len(a.shutdownContexts) < 3 {
		return cd.NewError(cd.Timeout, "still draining")
	}
	return nil
}

func TestExecuteAlwaysWaitsForCheckedShutdown(t *testing.T) {
	for _, phase := range []string{"success", "startup", "run", "panic"} {
		t.Run(phase, func(t *testing.T) {
			app := &executionApplication{t: t}
			failure := cd.NewError(cd.Unexpected, phase)
			if phase == "startup" {
				app.startupErr = failure
			}
			if phase == "run" {
				app.runErr = failure
			}
			app.panicRun = phase == "panic"
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			var result *cd.Error
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				result = execute(ctx, app, nil, time.Second, 0)
			}()
			if len(app.shutdownContexts) != 3 {
				t.Fatal("returned without complete shutdown")
			}
			if app.shutdownContexts[2].Err() == nil {
				t.Fatal("final budget was not released")
			}
			if phase == "startup" && app.runs != 0 {
				t.Fatal("Run followed failed Startup")
			}
			if (phase == "startup" || phase == "run") && result != failure {
				t.Fatal("lost execution failure", result)
			}
			if phase == "success" && result != nil {
				t.Fatal(result)
			}
			if (recovered != nil) != app.panicRun {
				t.Fatal("panic propagation changed", recovered)
			}
		})
	}
}
