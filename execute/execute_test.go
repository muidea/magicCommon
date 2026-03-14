package execute

import (
	"context"
	"testing"
	"time"
)

func TestExecuteWaitTimeoutReturnsTrueWhenTasksDrain(t *testing.T) {
	exec := NewExecute(2)
	exec.Run(func() {
		time.Sleep(40 * time.Millisecond)
	})

	start := time.Now()
	if !exec.WaitTimeout(500 * time.Millisecond) {
		t.Fatalf("expected WaitTimeout to report drained tasks")
	}
	if time.Since(start) < 30*time.Millisecond {
		t.Fatalf("expected WaitTimeout to wait for task completion")
	}
}

func TestExecuteWaitTimeoutReturnsFalseOnTimeout(t *testing.T) {
	exec := NewExecute(2)
	exec.Run(func() {
		time.Sleep(150 * time.Millisecond)
	})

	start := time.Now()
	if exec.WaitTimeout(20 * time.Millisecond) {
		t.Fatalf("expected WaitTimeout to time out")
	}
	if time.Since(start) > 120*time.Millisecond {
		t.Fatalf("expected WaitTimeout to return before task completion")
	}
}

func TestExecuteWaitContextRespectsContextCancel(t *testing.T) {
	exec := NewExecute(1)
	exec.Run(func() {
		time.Sleep(100 * time.Millisecond)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if exec.WaitContext(ctx) {
		t.Fatalf("expected WaitContext to stop on context cancellation")
	}
}

func TestExecuteIdleReportsDrainState(t *testing.T) {
	exec := NewExecute(1)
	if !exec.Idle() {
		t.Fatalf("new executor should be idle")
	}

	release := make(chan struct{})
	exec.Run(func() {
		<-release
	})

	time.Sleep(20 * time.Millisecond)
	if exec.Idle() {
		t.Fatalf("executor should report busy while task is running")
	}

	close(release)
	if !exec.WaitTimeout(200 * time.Millisecond) {
		t.Fatalf("expected task to drain")
	}
	if !exec.Idle() {
		t.Fatalf("executor should be idle after task drains")
	}
}
