package event

import "testing"

func TestDefaultHubOptionsCapsPerLaneChannelSize(t *testing.T) {
	opts := defaultHubOptions(500000)
	if opts.hubActionChanSize != defaultMaxHubActionChanSize {
		t.Fatalf("hubActionChanSize = %d, want %d", opts.hubActionChanSize, defaultMaxHubActionChanSize)
	}
	if opts.perLaneChanSize != defaultMaxPerLaneChanSize {
		t.Fatalf("perLaneChanSize = %d, want %d", opts.perLaneChanSize, defaultMaxPerLaneChanSize)
	}
	if opts.workerPoolSize != defaultMaxWorkerPoolSize {
		t.Fatalf("workerPoolSize = %d, want %d", opts.workerPoolSize, defaultMaxWorkerPoolSize)
	}
}

func TestDefaultHubOptionsKeepsSmallCapacityUntouched(t *testing.T) {
	opts := defaultHubOptions(8)
	if opts.perLaneChanSize != 8 {
		t.Fatalf("perLaneChanSize = %d, want 8", opts.perLaneChanSize)
	}
	if opts.hubActionChanSize != 8 {
		t.Fatalf("hubActionChanSize = %d, want 8", opts.hubActionChanSize)
	}
	if opts.workerPoolSize != 8 {
		t.Fatalf("workerPoolSize = %d, want 8", opts.workerPoolSize)
	}
}
