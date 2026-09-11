package task

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackgroundRoutineSmallAndDefaultCapacitiesDrain(t *testing.T) {
	for _, capacity := range []int{-1, 0, 1, 2} {
		t.Run(fmt.Sprint(capacity), func(t *testing.T) {
			q := NewBackgroundRoutine(capacity)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			defer q.Shutdown(ctx)
			var calls atomic.Int32
			for range 8 {
				if err := q.AsyncTaskContext(ctx, &routineTask{funcPtr: func() { calls.Add(1) }}); err != nil {
					t.Fatal(err)
				}
			}
			if !q.Shutdown(ctx) || calls.Load() != 8 {
				t.Fatal("accepted work failed to execute/drain", calls.Load())
			}
			if err := q.AsyncFunction(func() { t.Error("closed queue ran new work") }); err == nil {
				t.Fatal("closed queue accepted work")
			}
		})
	}
}
