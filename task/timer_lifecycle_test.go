package task

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

func TestRecurringTimerCancellationUnblocksSaturatedAdmission(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := NewBackgroundRoutine(1).(*backgroundRoutine)
		ctx, cancel := context.WithCancel(context.Background())
		entered, release := make(chan struct{}), make(chan struct{})
		calls := 0
		if err := queue.Timer(ctx, &routineTask{funcPtr: func() { calls++; close(entered); <-release }}, time.Second, 0); err != nil {
			t.Fatal(err)
		}
		<-entered
		for range 2 {
			if err := queue.AsyncFunction(func() {}); err != nil {
				t.Fatal(err)
			}
		}
		time.Sleep(2 * time.Second) // recurring tick is now blocked by the full queue
		cancel()
		timerDone := make(chan struct{})
		go func() { queue.timers.Wait(); close(timerDone) }()
		synctest.Wait()
		select {
		case <-timerDone:
		default:
			t.Error("recurring timer ignored cancellation during admission")
		}
		close(release)
		if !queue.Shutdown(context.Background()) {
			t.Fatal("queue failed to drain")
		}
		if calls != 1 {
			t.Fatal("canceled timer queued another callback", calls)
		}
	})
}

func TestTimerRegistrationAndShutdownHaveCheckedReceipts(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := NewBackgroundRoutine(2)
		job := &routineTask{funcPtr: func() { t.Error("long timer executed during shutdown") }}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := queue.Timer(ctx, job, time.Hour, 0); err == nil {
			t.Fatal("canceled timer registered")
		}
		if err := queue.Timer(context.Background(), job, time.Hour, 0); err != nil {
			t.Fatal(err)
		}
		budget, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		if !queue.Shutdown(budget) {
			t.Fatal("shutdown did not stop sleeping timer")
		}
		if err := queue.Timer(context.Background(), job, time.Hour, 0); err == nil {
			t.Fatal("closed scheduler registered timer")
		}
	})
}
