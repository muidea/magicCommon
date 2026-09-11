package task

import (
	"context"
	"testing"
	"time"
)

func TestCanceledAdmissionAndShutdownWakeSaturatedSubmitter(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		t.Run(map[bool]string{false: "cancel", true: "shutdown"}[shutdown], func(t *testing.T) {
			q := NewBackgroundRoutine(1)
			entered, release := make(chan struct{}), make(chan struct{})
			t.Cleanup(func() { close(release); q.Shutdown(context.Background()) })
			q.AsyncFunction(func() { close(entered); <-release })
			<-entered
			for range 2 {
				q.AsyncFunction(func() {})
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				done <- q.AsyncTaskContext(ctx, &routineTask{funcPtr: func() { t.Error("rejected task ran") }})
			}()
			if shutdown {
				budget, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
				defer stop()
				if q.Shutdown(budget) {
					t.Fatal("shutdown ignored active task")
				}
			} else {
				cancel()
			}
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("saturated admission succeeded")
				}
			case <-time.After(time.Second):
				t.Fatal("admission did not respond to cancel/close")
			}
		})
	}
}
