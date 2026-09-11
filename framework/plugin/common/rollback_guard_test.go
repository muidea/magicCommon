package common

import (
	"context"
	"reflect"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
)

type guardedSetupOwner struct {
	shutdownOwner
	failSetup, panicTeardown bool
}

func (s *guardedSetupOwner) Setup(context.Context, event.Hub, task.BackgroundRoutine) *cd.Error {
	*s.calls = append(*s.calls, s.name+":setup")
	if s.failSetup {
		return cd.NewError(cd.Unexpected, "partial setup failed")
	}
	return nil
}
func (s *guardedSetupOwner) Teardown(context.Context) {
	*s.calls = append(*s.calls, s.name+":release")
	if s.panicTeardown {
		panic("still owns dependencies")
	}
}

func TestFailedSetupKeepsAllEnteredOwnersUntilGuardSucceeds(t *testing.T) {
	manager := NewPluginMgr("test")
	var calls []string
	a := &guardedSetupOwner{shutdownOwner: shutdownOwner{name: "a", calls: &calls, reject: true}}
	b := &guardedSetupOwner{shutdownOwner: shutdownOwner{name: "b", calls: &calls}, failSetup: true}
	c := &guardedSetupOwner{shutdownOwner: shutdownOwner{name: "c", calls: &calls}}
	for _, owner := range []*guardedSetupOwner{a, b, c} {
		if err := manager.Register(owner); err != nil {
			t.Fatal(err)
		}
	}
	if err := manager.Setup(context.Background(), nil, nil); err == nil {
		t.Fatal("setup failure discarded")
	}
	if !reflect.DeepEqual(calls, []string{"a:setup", "b:setup"}) {
		t.Fatal("eager cleanup bypassed shutdown barrier", calls)
	}
	if err := manager.TeardownChecked(context.Background()); err == nil {
		t.Fatal("guard failure discarded")
	}
	if !reflect.DeepEqual(calls, []string{"a:setup", "b:setup", "a:close", "b:close", "a:drain"}) {
		t.Fatal("released busy or uninitialized owner", calls)
	}
	a.reject = false
	if err := manager.TeardownChecked(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{"a:setup", "b:setup", "a:close", "b:close", "a:drain", "a:close", "b:close", "a:drain", "b:drain", "b:release", "a:release"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatal(calls)
	}
}

func TestFinalTeardownPanicStopsDependenciesAndRetriesOnlyIncompleteOwners(t *testing.T) {
	manager := NewPluginMgr("test")
	var calls []string
	a := &guardedSetupOwner{shutdownOwner: shutdownOwner{name: "a", calls: &calls}}
	b := &guardedSetupOwner{shutdownOwner: shutdownOwner{name: "b", calls: &calls}, panicTeardown: true}
	c := &guardedSetupOwner{shutdownOwner: shutdownOwner{name: "c", calls: &calls}}
	for _, owner := range []*guardedSetupOwner{a, b, c} {
		if err := manager.Register(owner); err != nil {
			t.Fatal(err)
		}
	}
	if err := manager.Setup(context.Background(), nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := manager.TeardownChecked(context.Background()); err == nil {
		t.Fatal("teardown panic discarded")
	}
	calls = nil
	b.panicTeardown = false
	if err := manager.TeardownChecked(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{"a:close", "b:close", "a:drain", "b:drain", "b:release", "a:release"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatal("released completed stages again or skipped dependencies", calls)
	}
}
