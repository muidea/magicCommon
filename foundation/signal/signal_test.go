package signal

import (
	"testing"
	"time"
)

func TestGardWaitAndTrigger(t *testing.T) {
	gard := &Gard{}
	if err := gard.PutSignal(1); err != nil {
		t.Fatalf("PutSignal failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(10 * time.Millisecond)
		if err := gard.TriggerSignal(1, "ok"); err != nil {
			t.Errorf("TriggerSignal failed: %v", err)
		}
	}()

	val, err := gard.WaitSignal(1, 1)
	if err != nil {
		t.Fatalf("WaitSignal failed: %v", err)
	}
	if val != "ok" {
		t.Fatalf("unexpected signal value: %v", val)
	}
	<-done
}

func TestGardTriggerAfterCleanReturnsError(t *testing.T) {
	gard := &Gard{}
	if err := gard.PutSignal(2); err != nil {
		t.Fatalf("PutSignal failed: %v", err)
	}

	gard.CleanSignal(2)
	if err := gard.TriggerSignal(2, "late"); err == nil {
		t.Fatalf("expected TriggerSignal to fail after CleanSignal")
	}
}

func TestGardDuplicateSignalRejected(t *testing.T) {
	gard := &Gard{}
	if err := gard.PutSignal(3); err != nil {
		t.Fatalf("PutSignal failed: %v", err)
	}
	if err := gard.PutSignal(3); err == nil {
		t.Fatalf("expected duplicate PutSignal to fail")
	}
}

func TestGardResetClosesSignals(t *testing.T) {
	gard := &Gard{}
	if err := gard.PutSignal(4); err != nil {
		t.Fatalf("PutSignal failed: %v", err)
	}

	gard.Reset()
	if err := gard.TriggerSignal(4, "late"); err == nil {
		t.Fatalf("expected TriggerSignal to fail after Reset")
	}
}
