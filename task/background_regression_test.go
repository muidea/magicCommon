package task

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type slowTask struct {
	started atomic.Bool
	done    atomic.Bool
	delay   time.Duration
}

func (s *slowTask) Run() {
	s.started.Store(true)
	time.Sleep(s.delay)
	s.done.Store(true)
}

func TestSyncTaskWithTimeOutDoesNotPanicAfterTimeout(t *testing.T) {
	taskRoutine := NewBackgroundRoutine(8)
	defer taskRoutine.Shutdown(time.Second)
	taskPtr := &slowTask{delay: 20 * time.Millisecond}

	_ = taskRoutine.SyncTaskWithTimeOut(taskPtr, 1*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	if !taskPtr.started.Load() {
		t.Fatalf("expected timed task to start")
	}
	if !taskPtr.done.Load() {
		t.Fatalf("expected timed task to complete after timeout")
	}
}

func TestBackgroundRoutineTimerWithContextStopsAfterCancel(t *testing.T) {
	taskRoutine := NewBackgroundRoutine(8)
	defer taskRoutine.Shutdown(time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count atomic.Int32
	err := taskRoutine.TimerWithContext(ctx, Task(&routineTask{funcPtr: func() {
		count.Add(1)
	}}), 50*time.Millisecond, 0)
	if err != nil {
		t.Fatalf("expected timer setup to succeed: %v", err)
	}

	time.Sleep(130 * time.Millisecond)
	cancel()
	before := count.Load()
	time.Sleep(120 * time.Millisecond)
	after := count.Load()

	if before == 0 {
		t.Fatalf("expected timer to trigger before cancellation")
	}
	if after != before {
		t.Fatalf("expected timer to stop after cancellation, before=%d after=%d", before, after)
	}
}

func TestBackgroundRoutineShutdownRejectsNewTasks(t *testing.T) {
	taskRoutine := NewBackgroundRoutine(8)

	if !taskRoutine.Shutdown(time.Second) {
		t.Fatalf("expected shutdown to drain")
	}
	if taskRoutine.Shutdown(time.Second) != true {
		t.Fatalf("expected repeated shutdown to stay successful")
	}
	if err := taskRoutine.AsyncFunction(func() {}); err == nil {
		t.Fatalf("expected submitting task after shutdown to fail")
	}
}
