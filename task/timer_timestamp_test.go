package task

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

type timestampTask struct {
	ticks []time.Time
	runs  int
}

func (s *timestampTask) Run()                      { s.runs++ }
func (s *timestampTask) RunScheduled(at time.Time) { s.ticks = append(s.ticks, at) }

func TestTimerPreservesScheduledTimeAcrossQueueDelay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bg := NewBackgroundRoutine(1)
		release := make(chan struct{})
		if err := bg.AsyncFunction(func() { <-release }); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		job := &timestampTask{}
		start := time.Now()
		if err := bg.Timer(ctx, job, time.Second, 0); err != nil {
			t.Fatal(err)
		}
		time.Sleep(3500 * time.Millisecond)
		close(release)
		synctest.Wait()
		if len(job.ticks) != 3 || job.runs != 0 {
			t.Fatalf("unexpected timer dispatch: %v, ordinary runs=%d", job.ticks, job.runs)
		}
		for i, at := range job.ticks {
			if want := start.Add(time.Duration(i+1) * time.Second); !at.Equal(want) {
				t.Fatalf("tick %d lost scheduled time: %v, want %v", i, at, want)
			}
		}
		if err := bg.AsyncTask(job); err != nil {
			t.Fatal(err)
		}
		cancel()
		if !bg.Shutdown(context.Background()) || job.runs != 1 {
			t.Fatal("ordinary submission semantics changed", job.runs)
		}
	})
}
