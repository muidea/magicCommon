package event

import "testing"

func TestDefaultHubOptionsCapsPerLaneChannelSize(t *testing.T) {
	opts := defaultHubOptions(500000)
	if opts.hubActionChanSize != 500000 {
		t.Fatalf("hubActionChanSize = %d, want 500000", opts.hubActionChanSize)
	}
	if opts.perLaneChanSize != defaultMaxPerLaneChanSize {
		t.Fatalf("perLaneChanSize = %d, want %d", opts.perLaneChanSize, defaultMaxPerLaneChanSize)
	}
	if opts.workerPoolSize != 500000 {
		t.Fatalf("workerPoolSize = %d, want 500000", opts.workerPoolSize)
	}
}

func TestDefaultHubOptionsKeepsSmallCapacityUntouched(t *testing.T) {
	opts := defaultHubOptions(8)
	if opts.perLaneChanSize != 8 {
		t.Fatalf("perLaneChanSize = %d, want 8", opts.perLaneChanSize)
	}
}
