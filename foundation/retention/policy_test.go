package retention

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func seed(t *testing.T, path string, policy Policy) {
	t.Helper()
	data, err := json.Marshal(Document{Policies: map[string]Policy{"logs": policy}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
func TestPolicySaveVersionsAndPreservesRestart(t *testing.T) {
	file := filepath.Join(t.TempDir(), "policy.json")
	seed(t, file, Policy{90, 300, 500, 20})
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := Save(file, map[string]Policy{"logs": {30, 10, 5, 2}}, 0, "panel:account:7")
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
	doc, err := Read(file)
	if err != nil || doc.ResourceVersion != 1 || doc.Policies["logs"].RetentionDays != 30 {
		t.Fatalf("doc=%#v err=%v", doc, err)
	}
}
func TestPolicyReadRejectsCorruptionAndSymlink(t *testing.T) {
	file := filepath.Join(t.TempDir(), "policy.json")
	if err := os.WriteFile(file, []byte(`{"resourceVersion":1,"policies":{"logs":{"cleanupIntervalSeconds":1,"cleanupBatchSize":1,"cleanupMaxBatches":1}}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(file); err == nil {
		t.Fatal("missing retentionDays became forever")
	}
	seed(t, file, Policy{90, 300, 500, 20})
	link := file + "-link"
	if err := os.Symlink(file, link); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(link); err == nil {
		t.Fatal("symlink accepted")
	}
	if _, err := Save(file, map[string]Policy{"logs": {-1, 300, 500, 20}}, 0, "admin"); err == nil {
		t.Fatal("invalid policy saved")
	}
}
func TestControllerHotReloadDisablesResumesAndBoundsConcurrency(t *testing.T) {
	now := time.Now()
	file := filepath.Join(t.TempDir(), "policy.json")
	seed(t, file, Policy{0, 1, 500, 20})
	controller, err := NewController(file, "logs", Policy{90, 300, 500, 20}, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, due, err := controller.Begin(now.Add(time.Second)); err != nil || due {
		t.Fatalf("forever due=%v err=%v", due, err)
	}
	_, err = Save(file, map[string]Policy{"logs": {30, 1, 5, 2}}, 0, "admin")
	if err != nil {
		t.Fatal(err)
	}
	p, due, err := controller.Begin(now.Add(11 * time.Second))
	if err != nil || !due || p.CleanupBatchSize != 5 {
		t.Fatalf("hot policy=%v due=%v err=%v", p, due, err)
	}
	if _, due, _ := controller.Begin(now.Add(12 * time.Second)); due {
		t.Fatal("cleanup overlapped")
	}
	controller.Done()
	if _, due, _ := controller.Begin(now.Add(13 * time.Second)); !due {
		t.Fatal("cleanup did not resume")
	}
	controller.Done()
	if err := os.WriteFile(file, []byte("invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, due, err := controller.Begin(now.Add(25 * time.Second)); due || err == nil {
		t.Fatal("corrupt configuration authorized deletion")
	}
}

func TestControllerMissingProjectionDoesNotAuthorizeCleanup(t *testing.T) {
	now := time.Now()
	root := t.TempDir()
	for _, tc := range []struct {
		name    string
		file    string
		wantDue bool
	}{
		{"standalone", filepath.Join(root, "unmounted", "policy.json"), true},
		{"mounted-missing-file", filepath.Join(root, "policy.json"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			controller, err := NewController(tc.file, "logs", Policy{30, 1, 5, 2}, now)
			if err != nil {
				t.Fatal(err)
			}
			_, due, err := controller.Begin(now.Add(time.Second))
			if due != tc.wantDue || (err == nil) != tc.wantDue {
				t.Fatalf("due=%v err=%v, want due=%v", due, err, tc.wantDue)
			}
		})
	}
}
