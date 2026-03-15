package health

import (
	"context"
	"testing"

	cd "github.com/muidea/magicCommon/def"
)

func TestManagerCheckDependencies(t *testing.T) {
	manager := NewManager()
	manager.SetService("test")
	manager.MarkStarting()
	manager.RegisterDependencyChecker("magicCas", func(context.Context) *cd.Error {
		return nil
	})

	err := manager.CheckDependencies(context.Background(), []Dependency{
		Required("magicCas"),
	})
	if err != nil {
		t.Fatalf("expected dependency check success, got %v", err)
	}

	snapshot := manager.Snapshot()
	if snapshot.Checks["magicCas"].Status != "ready" {
		t.Fatalf("expected ready check status, got %#v", snapshot.Checks["magicCas"])
	}
}

func TestManagerCheckDependenciesRequiredFailure(t *testing.T) {
	manager := NewManager()
	manager.SetService("test")
	manager.MarkStarting()
	manager.RegisterDependencyChecker("magicCas", func(context.Context) *cd.Error {
		return cd.NewError(cd.Unexpected, "dependency not ready")
	})

	err := manager.CheckDependencies(context.Background(), []Dependency{
		Required("magicCas"),
	})
	if err == nil {
		t.Fatalf("expected dependency check failure")
	}

	snapshot := manager.Snapshot()
	if snapshot.Checks["magicCas"].Status != "failed" {
		t.Fatalf("expected failed check status, got %#v", snapshot.Checks["magicCas"])
	}
}

func TestManagerCheckDependenciesWithoutChecker(t *testing.T) {
	manager := NewManager()
	manager.SetService("test")
	manager.MarkStarting()

	err := manager.CheckDependencies(context.Background(), []Dependency{
		Required("magicCas"),
	})
	if err != nil {
		t.Fatalf("expected declaration-only dependency to pass before checker is registered, got %v", err)
	}

	snapshot := manager.Snapshot()
	if snapshot.Checks["magicCas"].Status != "declared" {
		t.Fatalf("expected declared check status, got %#v", snapshot.Checks["magicCas"])
	}
}
