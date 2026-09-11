package application

import (
	"context"
	"testing"
	"time"

	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
)

func TestStoppedApplicationOnlyRecreatesRuntimeOnStartup(t *testing.T) {
	app := NewApplication(Options{ConfigDir: t.TempDir(), BackgroundQueueSize: 1, EventHubQueueSize: 2})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	t.Cleanup(func() {
		if err := app.ShutdownChecked(ctx); err != nil {
			t.Error(err)
		}
	})
	for range 2 {
		if err := app.Startup(ctx, &MockService{}); err != nil {
			t.Fatal(err)
		}
		queue, hub := app.BackgroundRoutine(), app.EventHub()
		if err := queue.SyncFunction(func() {}); err != nil {
			t.Fatal(err)
		}
		if err := app.ShutdownChecked(ctx); err != nil {
			t.Fatal(err)
		}
		if app.BackgroundRoutine() != queue || app.EventHub() != hub {
			t.Fatal("shutdown created live runtime")
		}
		if err := queue.AsyncFunction(func() { t.Error("stopped runtime executed work") }); err == nil {
			t.Fatal("stopped queue accepted work")
		}
		result := hub.Send(event.NewEventWithContext("after-stop", "caller", "target", nil, ctx, nil))
		if result != nil && result.Error() == nil {
			t.Fatal("stopped hub accepted work")
		}
		if err := app.ShutdownChecked(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRestartDoesNotReuseOwnedInjectedRuntime(t *testing.T) {
	app := NewApplication(Options{ConfigDir: t.TempDir(), BackgroundRoutine: task.NewBackgroundRoutine(1),
		Ownership: RuntimeOwnership{BackgroundRoutine: true}})
	if err := app.Startup(context.Background(), &MockService{}); err != nil {
		t.Fatal(err)
	}
	if err := app.ShutdownChecked(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := app.Startup(context.Background(), &MockService{}); err == nil {
		t.Fatal("restart reused closed owned runtime")
	}
}
