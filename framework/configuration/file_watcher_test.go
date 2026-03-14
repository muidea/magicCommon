package configuration

import (
	"testing"
	"time"
)

func TestSimpleFileWatcherCloseIsIdempotent(t *testing.T) {
	watcher := NewSimpleFileWatcher(10 * time.Millisecond)
	watcher.Start()

	if err := watcher.Close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}

	if err := watcher.Close(); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
}
