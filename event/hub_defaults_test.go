package event

import "testing"

func TestDefaultHubOptionsCapsPerDestinationChannelSize(t *testing.T) {
	opts := defaultHubOptions(500000)
	if opts.hubActionChanSize != 500000 {
		t.Fatalf("hubActionChanSize = %d, want 500000", opts.hubActionChanSize)
	}
	if opts.perDestinationChanSize != defaultMaxPerDestinationChanSize {
		t.Fatalf("perDestinationChanSize = %d, want %d", opts.perDestinationChanSize, defaultMaxPerDestinationChanSize)
	}
	if opts.workerPoolSize != 500000 {
		t.Fatalf("workerPoolSize = %d, want 500000", opts.workerPoolSize)
	}
}

func TestDefaultHubOptionsKeepsSmallCapacityUntouched(t *testing.T) {
	opts := defaultHubOptions(8)
	if opts.perDestinationChanSize != 8 {
		t.Fatalf("perDestinationChanSize = %d, want 8", opts.perDestinationChanSize)
	}
}
